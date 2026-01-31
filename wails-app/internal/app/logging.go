package app

import (
	"database/sql"
	"strings"
	"sync"
	"time"
	"wails-app/internal/data/logger"
	"wails-app/internal/data/write"
	"wails-app/internal/platform/app_filter"

	"github.com/shirou/gopsutil/v3/process"
)

const processCheckInterval = 2 * time.Second

// loggedApps tracks which applications have already been logged (deduplication)
// Key is lowercase process name (e.g., "chrome.exe")
var loggedApps = make(map[string]bool)
var loggedAppsMu sync.Mutex

var resetLoggerCh = make(chan struct{}, 1)

// ResetLoggedApps clears the in-memory cache of logged applications.
// This allows applications that were previously logged to be logged again
// after a history clear.
func ResetLoggedApps() {
	resetLoggerCh <- struct{}{}
}

// StartProcessEventLogger starts a long-running goroutine that monitors process creation and termination events.
func StartProcessEventLogger(appLogger logger.Logger, db *sql.DB) {
	go func() {
		runningProcs := make(map[int32]string)
		initializeRunningProcs(runningProcs, db)

		ticker := time.NewTicker(processCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
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
			case <-resetLoggerCh:
				appLogger.Printf("[Logger] Reset signal received. Clearing in-memory state.")
				loggedAppsMu.Lock()
				loggedApps = make(map[string]bool)
				loggedAppsMu.Unlock()

				runningProcs = make(map[int32]string)
			}
		}
	}()
}

func logEndedProcesses(appLogger logger.Logger, db *sql.DB, runningProcs map[int32]string, currentProcs map[int32]bool) {
	for pid, nameLower := range runningProcs {
		if !currentProcs[pid] {
			write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
			delete(runningProcs, pid)

			// Check if any other running process has the same name
			isOtherSameNameRunning := false
			for _, otherName := range runningProcs {
				if otherName == nameLower {
					isOtherSameNameRunning = true
					break
				}
			}

			// If no other instances are running, allow the app to be logged again if restarted
			if !isOtherSameNameRunning {
				loggedAppsMu.Lock()
				delete(loggedApps, nameLower)
				loggedAppsMu.Unlock()
			}
		}
	}
}

type logStatus int

const (
	logStatusLog logStatus = iota
	logStatusExclude
	logStatusRetry
)

func logNewProcesses(appLogger logger.Logger, db *sql.DB, runningProcs map[int32]string, procs []*process.Process) {
	for _, p := range procs {
		if _, exists := runningProcs[p.Pid]; !exists {
			status := evaluateProcessForLogging(p)

			if status == logStatusLog {
				name, _ := p.Name()
				parent, _ := p.Parent()
				parentName := ""
				if parent != nil {
					parentName, _ = parent.Name()
				}

				exePath, _ := p.Exe()
				write.EnqueueWrite("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time) VALUES (?, ?, ?, ?, ?)",
					name, p.Pid, parentName, exePath, time.Now().Unix())
				
				// Mark as logged in the session deduplication map
				nameLower := strings.ToLower(name)
				loggedAppsMu.Lock()
				loggedApps[nameLower] = true
				loggedAppsMu.Unlock()
			}

			// If the status is Log or Exclude, we add it to runningProcs so we don't re-evaluate it.
			// If it's Retry, we leave it out of runningProcs so it's checked again in the next tick.
			if status != logStatusRetry {
				if name, err := p.Name(); err == nil && name != "" {
					runningProcs[p.Pid] = strings.ToLower(name)
				}
			}
		}
	}
}

func initializeRunningProcs(runningProcs map[int32]string, db *sql.DB) {
	rows, err := db.Query("SELECT pid, process_name FROM app_events WHERE end_time IS NULL")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pid int32
		var name string
		if err := rows.Scan(&pid, &name); err == nil {
			if exists, _ := process.PidExists(pid); exists {
				nameLower := strings.ToLower(name)
				runningProcs[pid] = nameLower
				
				// Ensure already running apps are marked as logged
				loggedAppsMu.Lock()
				loggedApps[nameLower] = true
				loggedAppsMu.Unlock()
			} else {
				write.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
			}
		}
	}
}

func evaluateProcessForLogging(p *process.Process) logStatus {
	name, err := p.Name()
	if err != nil || name == "" {
		return logStatusRetry // Try again later when metadata might be available
	}

	nameLower := strings.ToLower(name)
	exePath, err := p.Exe()
	if err != nil {
		return logStatusRetry // Try again later
	}

	// Rule 1: Platform-specific system exclusion (expensive checks inside)
	if app_filter.ShouldExclude(exePath, p) {
		return logStatusExclude
	}

	// Rule 2: Deduplication - Only log first instance of each application in the current session
	loggedAppsMu.Lock()
	alreadyLogged := loggedApps[nameLower]
	loggedAppsMu.Unlock()

	if alreadyLogged {
		// Even if already logged, we mark it as Exclude so we don't re-evaluate this specific PID
		return logStatusExclude
	}

	// Rule 3: Must be a trackable user application (e.g. has a window or is a specifically allowed process)
	if !app_filter.ShouldTrack(exePath, p) {
		// If it's not excluded and not logged, but not yet trackable, it's a candidate for retry.
		// Eventually it will either get a window (Log) or exit (Cleanup).
		return logStatusRetry
	}

	return logStatusLog
}
