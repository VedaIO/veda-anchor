//go:generate goversioninfo -64

package main

import (
	"embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Microsoft/go-winio"
)

// To build the launcher, you must first build veda-engine and veda-ui
// and place the binaries in the bin/ directory relative to this file.
//
//go:embed all:bin
var embeddedBinaries embed.FS

const pipeName = `\\.\pipe\veda`

func main() {
	// Setup logging
	cacheDir, _ := os.UserCacheDir()
	logDir := filepath.Join(cacheDir, "veda", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "veda_launcher.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	log.Printf("=== VEDA LAUNCHER STARTED === Args: %v", os.Args)

	// --- Step 1: Extract binaries ---
	installDir := filepath.Join(cacheDir, "veda", "bin")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		log.Fatalf("failed to create install dir: %v", err)
	}

	enginePath := filepath.Join(installDir, "veda-engine.exe")
	uiPath := filepath.Join(installDir, "veda-ui.exe")

	if err := extractFile("bin/veda-engine.exe", enginePath); err != nil {
		log.Printf("warning: failed to extract engine: %v", err)
	}
	if err := extractFile("bin/veda-ui.exe", uiPath); err != nil {
		log.Printf("warning: failed to extract UI: %v", err)
	}

	// --- Step 2: Detect if engine is already running ---
	engineRunning := isEngineRunning()

	if engineRunning {
		log.Println("[LAUNCH] Engine already running (pipe detected), skipping engine start")
	} else {
		log.Println("[LAUNCH] Engine not running, starting veda-engine...")
		engineCmd := exec.Command(enginePath)
		engineCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if err := engineCmd.Start(); err != nil {
			log.Printf("error starting engine: %v", err)
		} else {
			log.Printf("[LAUNCH] Engine started (PID: %d)", engineCmd.Process.Pid)
			// Give engine a moment to initialize the IPC pipe
			time.Sleep(500 * time.Millisecond)
		}
	}

	// --- Step 3: Always start the UI ---
	log.Println("[LAUNCH] Starting veda-ui...")
	uiCmd := exec.Command(uiPath)
	if err := uiCmd.Run(); err != nil {
		log.Fatalf("[LAUNCH] error running UI: %v", err)
	}

	log.Println("[LAUNCH] UI exited, launcher exiting")
}

// isEngineRunning checks if veda-engine is already running by attempting
// to connect to the named pipe via go-winio. If the dial succeeds, the engine is up.
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

	// Only write if different or not exists (simplified here)
	return os.WriteFile(dstPath, data, 0755)
}
