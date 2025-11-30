//go:build windows

package app

import (
	"database/sql"
	"debug/pe"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"syscall"
	"time"
	"unsafe"
	"wails-app/internal/data"

	"github.com/shirou/gopsutil/v3/process"
	"go.mozilla.org/pkcs7"
)

const (
	processCheckInterval     = 2 * time.Second
	blocklistEnforceInterval = 2 * time.Second
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")

	enumWindowsCallback = syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		//nolint:govet
		params := (*enumWindowsParams)(unsafe.Pointer(lParam))
		var windowPid uint32
		_, _, err := procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&windowPid)))
		if err != syscall.Errno(0) {
			return 1 // Continue on error
		}

		if windowPid == params.pid {
			if isVisible, _, _ := procIsWindowVisible.Call(uintptr(hwnd)); isVisible != 0 {
				params.found = true
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue
	})
)

type enumWindowsParams struct {
	pid   uint32
	found bool
}

// hasVisibleWindow checks if a process with the given PID has a visible window.
func hasVisibleWindow(pid uint32) bool {
	params := &enumWindowsParams{pid: pid, found: false}
	_, _, err := procEnumWindows.Call(enumWindowsCallback, uintptr(unsafe.Pointer(params)))
	if err != syscall.Errno(0) {
		data.GetLogger().Printf("Error enumerating windows: %v", err)
	}
	return params.found
}

// StartProcessEventLogger starts a long-running goroutine that monitors process creation and termination events.
func StartProcessEventLogger(appLogger data.Logger, db *sql.DB) {
	go func() {
		runningProcs := make(map[int32]bool)
		initializeRunningProcs(runningProcs, db)

		ticker := time.NewTicker(processCheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			procs, err := process.Processes()
			if err != nil {
				appLogger.Printf("Failed to get processes: %v", err)
				continue
			}

			currentProcs := make(map[int32]bool)
			for _, p := range procs {
				currentProcs[p.Pid] = true
			}

			logEndedProcesses(appLogger, db, runningProcs, currentProcs)
			logNewProcesses(appLogger, db, runningProcs, procs)
		}
	}()
}

func logEndedProcesses(appLogger data.Logger, db *sql.DB, runningProcs, currentProcs map[int32]bool) {
	for pid := range runningProcs {
		if !currentProcs[pid] {
			data.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
			delete(runningProcs, pid)
		}
	}
}

func logNewProcesses(appLogger data.Logger, db *sql.DB, runningProcs map[int32]bool, procs []*process.Process) {
	for _, p := range procs {
		if !runningProcs[p.Pid] {
			if shouldLogProcess(p) {
				name, _ := p.Name()

				// Skip logging ProcGuard itself
				if strings.ToLower(name) == "procguard.exe" {
					runningProcs[p.Pid] = true
					continue
				}

				parent, _ := p.Parent()
				parentName := ""
				if parent != nil {
					parentName, _ = parent.Name()
				}

				exePath, err := p.Exe()
				if err != nil {
					appLogger.Printf("Failed to get exe path for %s (pid %d): %v", name, p.Pid, err)
				}
				data.EnqueueWrite("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time) VALUES (?, ?, ?, ?, ?)",
					name, p.Pid, parentName, exePath, time.Now().Unix())
				runningProcs[p.Pid] = true
			}
		}
	}
}

func getPublisherName(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			data.GetLogger().Printf("Failed to close file %s: %v", filePath, err)
		}
	}()

	peFile, err := pe.NewFile(file)
	if err != nil {
		return "", fmt.Errorf("error parsing PE file %s: %w", filePath, err)
	}

	var securityDir pe.DataDirectory
	switch oh := peFile.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		securityDir = oh.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY]
	case *pe.OptionalHeader64:
		securityDir = oh.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY]
	default:
		return "", fmt.Errorf("unsupported PE optional header type: %T", peFile.OptionalHeader)
	}

	if securityDir.Size == 0 {
		return "", fmt.Errorf("no security directory found")
	}

	pkcs7Offset := int64(securityDir.VirtualAddress + 8)
	pkcs7Size := int64(securityDir.Size - 8)

	if pkcs7Size <= 0 {
		return "", fmt.Errorf("invalid signature size")
	}

	signatureBytes := make([]byte, pkcs7Size)
	_, err = file.ReadAt(signatureBytes, pkcs7Offset)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading signature from file: %w", err)
	}

	p7, err := pkcs7.Parse(signatureBytes)
	if err != nil {
		return "", fmt.Errorf("error parsing PKCS#7 signature: %w", err)
	}

	if len(p7.Certificates) == 0 {
		return "", fmt.Errorf("no certificates found in signature")
	}

	for _, cert := range p7.Certificates {
		if len(cert.Subject.Organization) > 0 {
			return cert.Subject.Organization[0], nil
		}
	}

	return "", fmt.Errorf("no organization name found in any certificate")
}

func isMicrosoftProcess(p *process.Process) bool {
	exePath, err := p.Exe()
	if err != nil {
		return false
	}

	publisher, err := getPublisherName(exePath)
	if err != nil {
		return false
	}

	return strings.Contains(publisher, "Microsoft")
}

func initializeRunningProcs(runningProcs map[int32]bool, db *sql.DB) {
	rows, err := db.Query("SELECT pid FROM app_events WHERE end_time IS NULL")
	if err != nil {
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			data.GetLogger().Printf("Failed to close rows: %v", err)
		}
	}()

	for rows.Next() {
		var pid int32
		if err := rows.Scan(&pid); err == nil {
			if exists, _ := process.PidExists(pid); exists {
				runningProcs[pid] = true
			} else {
				data.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
			}
		}
	}
}

func StartBlocklistEnforcer(appLogger data.Logger) {
	go func() {
		killTick := time.NewTicker(blocklistEnforceInterval)
		defer killTick.Stop()
		for range killTick.C {
			list, err := data.LoadAppBlocklist()
			if err != nil {
				appLogger.Printf("failed to fetch blocklist: %v", err)
				continue
			}
			if len(list) == 0 {
				continue
			}
			procs, err := process.Processes()
			if err != nil {
				appLogger.Printf("Failed to get processes: %v", err)
				continue
			}
			for _, p := range procs {
				name, _ := p.Name()
				if name == "" {
					continue
				}

				if slices.Contains(list, strings.ToLower(name)) {
					err := p.Kill()
					if err != nil {
						appLogger.Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
					} else {
						appLogger.Printf("killed blocked process %s (pid %d)", name, p.Pid)
					}
				}
			}
		}
	}()
}

var ignoredProcesses = []string{
	"textinputhost.exe",
}

func shouldLogProcess(p *process.Process) bool {
	name, err := p.Name()
	if err != nil || name == "" {
		return false
	}

	for _, ignoredName := range ignoredProcesses {
		if strings.EqualFold(name, ignoredName) {
			return false
		}
	}

	if name == "ProcGuardSvc.exe" {
		return false
	}

	if isMicrosoftProcess(p) {
		return false
	}

	if hasVisibleWindow(uint32(p.Pid)) {
		return true
	}

	il, err := GetProcessIntegrityLevel(uint32(p.Pid))
	if err == nil && il >= SECURITY_MANDATORY_HIGH_RID {
		return false
	}

	parent, err := p.Parent()
	if err != nil {
		return true
	}

	if isMicrosoftProcess(parent) {
		return false
	}

	return false
}
