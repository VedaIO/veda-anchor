package web

import (
	"os"
	"path/filepath"
	"runtime"
)

// CheckChromeExtension checks if the ProcGuard Chrome extension is installed
// by looking for it in Chrome's extensions directory on the filesystem
//
// We check for both:
//   - Store version: hkanepohpflociaodcicmmfbdaohpceo
//   - Dev version: gpaafgcbiejjpfdgmjglehboafdicdjb (fixed via manifest.json key)
//
// Returns: true if either extension is found, false otherwise
func CheckChromeExtension() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check for Store Extension
	if checkExtensionPath("hkanepohpflociaodcicmmfbdaohpceo", homeDir) {
		return true
	}

	// Check for Dev Extension (has fixed ID via manifest.json key)
	if checkExtensionPath("gpaafgcbiejjpfdgmjglehboafdicdjb", homeDir) {
		return true
	}

	return false
}

func checkExtensionPath(id, homeDir string) bool {
	var extensionPath string

	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		extensionPath = filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Extensions", id)
	case "darwin":
		extensionPath = filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "Extensions", id)
	case "linux":
		extensionPath = filepath.Join(homeDir, ".config", "google-chrome", "Default", "Extensions", id)
	default:
		return false
	}

	if _, err := os.Stat(extensionPath); err == nil {
		return true
	}
	return false
}
