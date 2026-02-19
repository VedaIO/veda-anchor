package main

import (
	"embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// To build the launcher, you must first build veda-engine and veda-ui
// and place the binaries in the bin/ directory relative to this file.
//
//go:embed all:bin
var embeddedBinaries embed.FS

func main() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("failed to get cache dir: %v", err)
	}

	installDir := filepath.Join(cacheDir, "vedaio", "bin")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		log.Fatalf("failed to create install dir: %v", err)
	}

	// Extract binaries
	enginePath := filepath.Join(installDir, "veda-engine.exe")
	uiPath := filepath.Join(installDir, "veda-ui.exe")

	if err := extractFile("bin/veda-engine.exe", enginePath); err != nil {
		log.Printf("warning: failed to extract engine: %v", err)
	}
	if err := extractFile("bin/veda-ui.exe", uiPath); err != nil {
		log.Printf("warning: failed to extract UI: %v", err)
	}

	// Start engine in background
	log.Println("Starting Veda Engine...")
	engineCmd := exec.Command(enginePath)
	engineCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := engineCmd.Start(); err != nil {
		log.Printf("error starting engine: %v", err)
	}

	// Start UI
	log.Println("Starting Veda UI...")
	uiCmd := exec.Command(uiPath)
	if err := uiCmd.Run(); err != nil {
		log.Fatalf("error starting UI: %v", err)
	}
}

func extractFile(srcPath, dstPath string) error {
	data, err := embeddedBinaries.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Only write if different or not exists (simplified here)
	return os.WriteFile(dstPath, data, 0755)
}
