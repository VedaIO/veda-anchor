//go:generate goversioninfo -64

package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/mgr"
)

//go:embed all:bin
var embeddedBinaries embed.FS

const (
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

	log.Printf("=== VEDA ANCHOR LAUNCHER STARTED === Args: %v, Admin: %v", os.Args, isAdmin())

	enginePath := filepath.Join(installDir, "veda-anchor-engine.exe")
	uiPath := filepath.Join(installDir, "veda-anchor-ui.exe")

	// --- Fast path: engine already running, no admin needed ---
	if isEngineRunning() {
		log.Println("[LAUNCH] Engine already running, launching UI directly")
		launchUI(uiPath)
		return
	}

	// --- Engine not running: need admin privileges ---
	if !isAdmin() {
		log.Println("[LAUNCH] Engine not running and not admin, showing error prompt")
		showErrorAndExit("Veda Anchor", "The engine is not running.\nPlease right-click the launcher and select \"Run as administrator\" to install or restart the service.")
		return
	}

	// --- Admin path: install if needed, then start service ---
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

	// Start the service
	log.Println("[LAUNCH] Starting service...")
	if err := startService(); err != nil {
		log.Printf("[LAUNCH] Warning: failed to start service: %v", err)
	} else {
		log.Println("[LAUNCH] Service started, waiting for pipe...")
		// Wait for the IPC pipe to become available
		waitForEngine(5 * time.Second)
	}

	launchUI(uiPath)
}

// launchUI starts the UI executable and exits the launcher.
func launchUI(uiPath string) {
	log.Println("[LAUNCH] Starting veda-anchor-ui...")
	uiCmd := exec.Command(uiPath)
	if err := uiCmd.Start(); err != nil {
		log.Printf("[LAUNCH] Failed to start UI: %v", err)
	}
	log.Println("[LAUNCH] UI launched, launcher exiting")
}

// isAdmin checks if the current process is running with elevated privileges.
func isAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

// showErrorAndExit displays a Windows message box and exits.
func showErrorAndExit(title, message string) {
	var (
		user32         = syscall.NewLazyDLL("user32.dll")
		procMessageBox = user32.NewProc("MessageBoxW")
	)

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(message)

	// MB_OK | MB_ICONERROR = 0x00000010
	procMessageBox.Call(0,
		uintptr(unsafe.Pointer(msgPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(0x10),
	)

	os.Exit(1)
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

// isServiceInstalled checks if the VedaAnchorEngine service exists in SCM.
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

// registerService creates the VedaAnchorEngine Windows Service with recovery actions.
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

// startService starts the VedaAnchorEngine service via SCM.
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

// isEngineRunning checks if veda-anchor-engine.exe is running (non-admin safe).
func isEngineRunning() bool {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(snapshot)

	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	for err = windows.Process32First(snapshot, &procEntry); err == nil; err = windows.Process32Next(snapshot, &procEntry) {
		exeName := windows.UTF16ToString(procEntry.ExeFile[:])
		if exeName == "veda-anchor-engine.exe" {
			return true
		}
	}
	return false
}

func extractFile(srcPath, dstPath string) error {
	data, err := embeddedBinaries.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, data, 0755)
}
