//go:build windows

package api

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"wails-app/internal/auth"
	"wails-app/internal/daemon"
	"wails-app/internal/data"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows/registry"
)

const appName = "ProcGuard"
const hostName = "com.nixuris.procguard"

// Uninstall handles the uninstallation of the application.
// It performs a series of cleanup tasks in a separate goroutine and then initiates a self-deletion process.
func (s *Server) Uninstall(password string) error {
	cfg, err := data.LoadConfig()
	if err != nil {
		return err
	}

	if !auth.CheckPasswordHash(password, cfg.PasswordHash) {
		return fmt.Errorf("invalid password")
	}

	go func() {
		// Close the logger and database to release file handles before deletion.
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			// We can't use the logger here, so just print to stderr.
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", err)
		}

		// Terminate any other running ProcGuard processes.
		killOtherProcGuardProcesses(s.Logger)

		// Unblock any files that were blocked by the application.
		if err := unblockAll(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unblock all files: %v\n", err)
		}

		// Perform other cleanup tasks.
		if err := daemon.RemoveAutostart(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove autostart: %v\n", err)
		}
		if err := removeNativeHost(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove native host: %v\n", err)
		}

		// Initiate the self-deletion process.
		if err := selfDelete(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initiate self-deletion: %v\n", err)
		}

		// Exit the application to allow the self-deletion to complete.
		os.Exit(0)
	}()

	return nil
}

// selfDelete creates and executes a batch script that deletes the application files after the main process has exited.
// This is a common technique for applications on Windows to perform self-uninstallation.
func selfDelete() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("self-deletion is currently implemented only for Windows")
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	appDataDir := filepath.Join(localAppData, appName)

	// Create a temporary batch file in the system's temp directory.
	tempDir := os.TempDir()
	batchFileName := fmt.Sprintf("delete_procguard_%d.bat", time.Now().UnixNano())
	batchFilePath := filepath.Join(tempDir, batchFileName)

	// The batch script waits for a moment to ensure the main process has exited,
	// then deletes the application's data directory and finally deletes itself.
	batchContent := fmt.Sprintf(`
@echo off
timeout /t 2 /nobreak > nul
rmdir /s /q "%s"
del "%s"
`, appDataDir, batchFilePath)

	err := os.WriteFile(batchFilePath, []byte(batchContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write batch file: %w", err)
	}

	// Execute the batch file in a new, detached process so it can run independently of the main application.
	cmd := exec.Command("cmd.exe", "/C", batchFilePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start batch process: %w", err)
	}

	return nil
}

// killOtherProcGuardProcesses finds and terminates any other running ProcGuard processes.
func killOtherProcGuardProcesses(logger data.Logger) {
	currentPid := os.Getpid()
	procs, err := process.Processes()
	if err != nil {
		return
	}

	for _, p := range procs {
		if p.Pid == int32(currentPid) {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.HasPrefix(strings.ToLower(name), "procguard") {
			if err := p.Kill(); err != nil {
				logger.Printf("Failed to kill process %s: %v", name, err)
			}
		}
	}
}

// unblockAll restores the original names of any files that were blocked by the application.
func unblockAll() error {
	list, err := data.LoadAppBlocklist()
	if err != nil {
		return fmt.Errorf("could not load blocklist: %w", err)
	}

	for _, name := range list {
		if strings.HasSuffix(name, ".blocked") {
			newName := strings.TrimSuffix(name, ".blocked")
			if err := os.Rename(name, newName); err != nil {
				// Log the error but continue trying to unblock other files.
				data.GetLogger().Printf("Failed to unblock file %s: %v", name, err)
			}
		}
	}

	return nil
}

// removeNativeHost removes the native messaging host configuration from the system.
func removeNativeHost() error {
	// Delete the registry key for the native messaging host.
	keyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + hostName
	if err := registry.DeleteKey(registry.CURRENT_USER, keyPath); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Delete the manifest file.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(cacheDir, "procguard", "procguard.json")
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
