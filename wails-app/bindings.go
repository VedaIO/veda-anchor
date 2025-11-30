package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"wails-app/api"
	"wails-app/internal/data"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds the application context and server instance
// ctx: Wails runtime context - used to call runtime methods like WindowShow, WindowUnminimise
// *api.Server: Embedded server instance that handles all business logic
type App struct {
	ctx context.Context
	*api.Server

	// Legacy field (not used anymore)
	IsNativeMessagingActive bool
}

// NewApp creates a new App application struct
// This is called from main() to initialize the application
func NewApp() *App {
	return &App{}
}

// CheckChromeExtension checks if the Chrome extension is connected
func (a *App) CheckChromeExtension() bool {
	log := data.GetLogger() // Use the logger we set up in main.go

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("[CheckExt] Error getting cache dir: %v", err)
		return false
	}

	heartbeatPath := filepath.Join(cacheDir, "procguard", "extension_heartbeat")

	// Read timestamp from file
	content, err := os.ReadFile(heartbeatPath)
	if err != nil {
		// Don't log "not exist" errors too noisily as it's expected when not installed
		if !os.IsNotExist(err) {
			log.Printf("[CheckExt] Error reading file: %v", err)
		}
		return false
	}

	// Parse timestamp
	var lastPing int64
	if _, err := fmt.Sscanf(string(content), "%d", &lastPing); err != nil {
		log.Printf("[CheckExt] Error parsing timestamp '%s': %v", string(content), err)
		return false
	}

	// Check if ping is recent (within last 10 seconds)
	pingTime := time.Unix(lastPing, 0)
	timeSince := time.Since(pingTime)
	isValid := timeSince < 10*time.Second

	return isValid
}

// OpenBrowser opens a URL in the user's default system browser
//
// Why this exists: window.open() opens URLs inside the Wails WebView, not external browser
// Problem this fixes: Clicking "Install Extension" was trying to open Chrome Web Store in WebView
// Solution: Use OS-specific commands to open external browser
//
// Platform support:
//   - Windows: uses 'cmd /c start'
//   - macOS: uses 'open'
//   - Linux: uses 'xdg-open'
//
// Returns: error if command fails to start, nil on success
func (a *App) OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// ShowWindow brings the application window to the foreground
// Unminimizes the window if it's minimized, then makes it visible
//
// IMPORTANT: Only call this when user explicitly wants to see the window!
// DO NOT call this from polling/background operations or it will interrupt the user
//
// Used by: OnSecondInstanceLaunch callback - when user double-clicks exe while app is running
// Context: With HideWindowOnClose=true, closing the window hides it but keeps daemon running
//
//	When user runs exe again, SingleInstanceLock prevents new process and calls this instead
func (a *App) ShowWindow() {
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.Show(a.ctx)
}
