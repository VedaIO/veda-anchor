package main

import (
	"context"
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"log"
	"os"
	"path/filepath"
	"strings"
	"wails-app/api"
	"wails-app/internal/daemon"
	"wails-app/internal/data"
	"wails-app/internal/data/logger"
	"wails-app/internal/platform/nativehost"
	"wails-app/internal/web/native_messaging"
)

// Embed the entire frontend/dist directory into the Go binary
// This allows the app to be distributed as a single executable
//
//go:embed all:frontend/dist
var assets embed.FS

// startup is called when the Wails app starts
// The context is saved so we can call runtime methods (WindowShow, etc.) later
//
// Responsibilities:
//  1. Save the Wails runtime context for later use
//  2. Initialize the database
//  3. Initialize the logger
//  4. Create the API server
//  5. Start the background daemon for process/web monitoring
//  6. Start the native messaging host for Chrome extension communication

func (a *App) startup(ctx context.Context) {
	// Save context - CRITICAL for calling ShowWindow() and other runtime methods
	a.ctx = ctx

	// Initialize database connection
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize logger with database
	logger.NewLogger(db)

	// Create API server with database connection
	a.Server = api.NewServer(db)

	// Start the background daemon that monitors processes and web activity
	// This runs independently of the GUI - continues even when window is hidden
	daemon.Start(a.Logger, db)

	// Ensure Native Messaging Host is registered
	// This creates the registry key and manifest file so Chrome can find us
	// We do this on every startup to ensure the config is correct
	if err := nativehost.RegisterExtension("hkanepohpflociaodcicmmfbdaohpceo"); err != nil {
		log.Printf("Failed to register Store extension: %v", err)
	}
	if err := nativehost.RegisterExtension("gpaafgcbiejjpfdgmjglehboafdicdjb"); err != nil {
		log.Printf("Failed to register Dev extension: %v", err)
	}
}

func main() {
	// CRITICAL: Log startup for debugging
	// Use absolute path in CacheDir because CWD varies when launched by Chrome
	cacheDir, _ := os.UserCacheDir()
	logDir := filepath.Join(cacheDir, "ProcGuard", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "procguard_debug.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	log.Printf("=== PROCGUARD LAUNCHED === Args: %v", os.Args)
	log.Printf("CWD: %v", func() string { wd, _ := os.Getwd(); return wd }())

	// MODE 1: NATIVE MESSAGING HOST
	// Chrome launches us with the extension ID as an argument: chrome-extension://...
	// In this mode, we MUST NOT show a GUI. We only run the message loop.
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		log.Println("[MODE] Native Messaging Host detected")
		native_messaging.Run()
		log.Println("[MODE] Native Messaging Host exited")
		return
	}

	// MODE 2: GUI APPLICATION
	// User launched us (double-click, start menu, etc.)
	log.Println("[MODE] GUI Application detected")

	app := NewApp()

	// Check for --background flag to start hidden (e.g. from autostart)
	startHidden := false
	for _, arg := range os.Args {
		if arg == "--background" {
			startHidden = true
			break
		}
	}
	if startHidden {
		log.Println("[MODE] Background/Autostart detected - starting hidden")
	}

	// Create and run the Wails application
	err := wails.Run(&options.App{
		Title:       "ProcGuard",
		Width:       1024,
		Height:      768,
		Frameless:   true, // Enable frameless mode
		StartHidden: startHidden,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,

		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               filepath.Join(os.Getenv("LOCALAPPDATA"), "ProcGuard", "webview"),
		},

		// HideWindowOnClose: Keep app running in background
		HideWindowOnClose: true,

		// SingleInstanceLock: Ensure only one GUI instance runs
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "com.procguard.wails-app",
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				log.Println("Second GUI instance detected - showing existing window")
				app.ShowWindow()
			},
		},

		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error running Wails app:", err)
	}
}
