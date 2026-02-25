//go:generate goversioninfo -64

package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/mgr"
)

//go:embed all:bin
var embeddedBinaries embed.FS

const (
	pipeName    = `\\.\pipe\veda-anchor`
	serviceName = "VedaAnchorEngine"
)

func main() {
	// Determine install directory
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}
	installDir := filepath.Join(programFiles, "VedaAnchor")

	// Setup logging (Shared data root)
	progData := os.Getenv("ProgramData")
	if progData == "" {
		progData = `C:\ProgramData`
	}
	logDir := filepath.Join(progData, "VedaAnchor", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "veda-anchor_launcher.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	log.Printf("=== VEDA ANCHOR LAUNCHER STARTED === Args: %v", os.Args)

	enginePath := filepath.Join(installDir, "veda-anchor-engine.exe")
	uiPath := filepath.Join(installDir, "veda-anchor-ui.exe")

	// --- Install if needed ---
	serviceOK := isServiceInstalled()
	binariesOK := fileExists(enginePath) && fileExists(uiPath)

	if serviceOK && binariesOK {
		log.Println("[INSTALL] Already installed, skipping")
	} else {
		// If service exists but binaries are missing, clean up stale service first
		if serviceOK && !binariesOK {
			log.Println("[INSTALL] Stale service found (binaries missing), cleaning up...")
			deleteService()
		}
		log.Println("[INSTALL] Running install...")
		if err := install(installDir, enginePath, uiPath); err != nil {
			log.Fatalf("[INSTALL] Failed: %v", err)
		}
	}

	// --- Launch logic ---
	if isEngineRunning() {
		// Engine is already running (service is active)
		// Just launch the UI
		log.Println("[LAUNCH] Engine already running, launching UI only")
	} else {
		// Engine not running, start the service
		log.Println("[LAUNCH] Engine not running, starting service...")
		if err := startService(); err != nil {
			log.Printf("[LAUNCH] Warning: failed to start service: %v", err)
		} else {
			log.Println("[LAUNCH] Service started, waiting for pipe...")
			// Wait for the IPC pipe to become available
			waitForEngine(5 * time.Second)
		}
	}

	log.Println("[LAUNCH] Starting veda-anchor-ui...")
	uiCmd := exec.Command(uiPath)
	if err := uiCmd.Start(); err != nil {
		log.Printf("[LAUNCH] Failed to start UI: %v", err)
	}

	log.Println("[LAUNCH] UI launched, launcher exiting")
}

// install performs first-time setup: deploy binaries, register service, set up UI autostart.
func install(installDir, enginePath, uiPath string) error {
	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	// Deploy binaries
	if err := extractFile("bin/veda-anchor-engine.exe", enginePath); err != nil {
		return fmt.Errorf("extract engine: %w", err)
	}
	if err := extractFile("bin/veda-anchor-ui.exe", uiPath); err != nil {
		return fmt.Errorf("extract UI: %w", err)
	}
	log.Printf("[INSTALL] Binaries deployed to %s", installDir)

	// Register Windows Service
	if err := registerService(enginePath); err != nil {
		return fmt.Errorf("register service: %w", err)
	}
	log.Println("[INSTALL] Service registered")

	// Register UI autostart in HKLM
	if err := registerUIAutostart(uiPath); err != nil {
		log.Printf("[INSTALL] Warning: failed to register UI autostart: %v", err)
	}

	return nil
}

// isServiceInstalled checks if the VedaEngine service exists in SCM.
func isServiceInstalled() bool {
	m, err := mgr.Connect()
	if err != nil {
		return false
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return false
	}
	s.Close()
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// deleteService removes the existing service from SCM.
func deleteService() {
	m, err := mgr.Connect()
	if err != nil {
		log.Printf("[INSTALL] Warning: could not connect to SCM for cleanup: %v", err)
		return
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return // service doesn't exist, nothing to clean
	}
	_ = s.Delete()
	s.Close()
	time.Sleep(500 * time.Millisecond)
}

// registerService creates the AnchorEngine Windows Service with recovery actions.
func registerService(exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.CreateService(serviceName, exePath, mgr.Config{
		DisplayName:      "Veda Anchor Engine",
		Description:      "Core monitoring and blocking engine for Veda Anchor",
		StartType:        mgr.StartAutomatic,
		ServiceStartName: "LocalSystem",
	})
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	defer s.Close()

	// Recovery actions: restart on failure
	recoveryActions := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 1 * time.Minute},
		{Type: mgr.ServiceRestart, Delay: 2 * time.Minute},
		{Type: mgr.ServiceRestart, Delay: 5 * time.Minute},
	}
	if err := s.SetRecoveryActions(recoveryActions, uint32(24*60*60)); err != nil {
		log.Printf("Warning: failed to set recovery actions: %v", err)
	}

	// Recover even on non-crash failures (non-zero exit code)
	if err := s.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		log.Printf("Warning: failed to set non-crash recovery: %v", err)
	}

	return nil
}

// registerUIAutostart adds veda-anchor-ui.exe to HKLM Run for all users.
func registerUIAutostart(uiPath string) error {
	key, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("open HKLM Run key: %w", err)
	}
	defer key.Close()

	return key.SetStringValue("VedaAnchorUI", fmt.Sprintf(`"%s"`, uiPath))
}

// startService starts the VedaEngine service via SCM.
func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Start()
}

// waitForEngine polls the named pipe until the engine is ready or timeout.
func waitForEngine(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if isEngineRunning() {
			log.Println("[LAUNCH] Engine pipe is ready")
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	log.Println("[LAUNCH] Warning: engine pipe not ready after timeout")
}

// isEngineRunning checks if veda-anchor-engine is already running via named pipe.
func isEngineRunning() bool {
	conn, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func extractFile(srcPath, dstPath string) error {
	data, err := embeddedBinaries.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, data, 0755)
}
