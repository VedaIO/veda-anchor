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
	pipeName    = `\\.\pipe\veda-anchor-anchor-anchor`
	serviceName = "VedaEngine"
)

func main() {
	// Setup logging
	cacheDir, _ := os.UserCacheDir()
	logDir := filepath.Join(cacheDir, "veda-anchor-anchor-anchor", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "veda-anchor-anchor-anchor_launcher.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	log.Printf("=== VEDA LAUNCHER STARTED === Args: %v", os.Args)

	// Determine install directory
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}
	installDir := filepath.Join(programFiles, "VedaIO")
	enginePath := filepath.Join(installDir, "veda-anchor-anchor-anchor-engine.exe")
	uiPath := filepath.Join(installDir, "veda-anchor-anchor-anchor-ui.exe")

	// --- Install if needed (first run or upgrade) ---
	if !isServiceInstalled() {
		log.Println("[INSTALL] Service not found, running first-time install...")
		if err := install(installDir, enginePath, uiPath); err != nil {
			log.Fatalf("[INSTALL] Failed: %v", err)
		}
	} else {
		log.Println("[INSTALL] Service already installed, skipping install")
		// Still update the binaries in case this is an upgrade
		log.Println("[INSTALL] Updating binaries...")
		_ = extractFile("bin/veda-anchor-anchor-anchor-engine.exe", enginePath)
		_ = extractFile("bin/veda-anchor-anchor-anchor-ui.exe", uiPath)
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

	log.Println("[LAUNCH] Starting veda-anchor-anchor-anchor-ui...")
	uiCmd := exec.Command(uiPath)
	if err := uiCmd.Run(); err != nil {
		log.Printf("[LAUNCH] UI exited with error: %v", err)
	}

	log.Println("[LAUNCH] UI exited, launcher exiting")
}

// install performs first-time setup: deploy binaries, register service, set up UI autostart.
func install(installDir, enginePath, uiPath string) error {
	// Create install directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	// Deploy binaries
	if err := extractFile("bin/veda-anchor-anchor-anchor-engine.exe", enginePath); err != nil {
		return fmt.Errorf("extract engine: %w", err)
	}
	if err := extractFile("bin/veda-anchor-anchor-anchor-ui.exe", uiPath); err != nil {
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

// registerService creates the VedaEngine Windows Service with recovery actions.
func registerService(exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to SCM: %w", err)
	}
	defer m.Disconnect()

	binPath := fmt.Sprintf(`"%s"`, exePath)

	s, err := m.CreateService(serviceName, binPath, mgr.Config{
		DisplayName:      "VedaIO Engine",
		Description:      "Core monitoring and blocking engine for VedaIO",
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

// registerUIAutostart adds veda-anchor-anchor-anchor-ui.exe to HKLM Run for all users.
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

	return key.SetStringValue("VedaUI", fmt.Sprintf(`"%s"`, uiPath))
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

// isEngineRunning checks if veda-anchor-anchor-anchor-engine is already running via named pipe.
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
