> THE PROJECT IS NOT YET PRODUCTION READY

# Veda

Veda Anchor is a Windows-based desktop application for monitoring and controlling processes and web activity on your system. Built with the Wails framework, it combines a Go backend with a Svelte frontend into a native executable, featuring dual-mode operation for both GUI interaction and browser extension communication.

## Architecture

Veda Anchor operates in two distinct modes:

### GUI Application Mode
The primary mode where users interact with the application through a native desktop window. The application includes:

- **Wails Runtime**: Packages Go backend and Svelte frontend into a single native Windows executable
- **API Server**: Exposes methods to the frontend via Wails bindings for data access and system control
- **SQLite Database**: Stores all activity logs, metadata, and configuration
- **Background Services**: Process monitoring, screen time tracking, and blocklist enforcement run independently of the GUI

### Native Messaging Host Mode
A headless mode launched by Chrome to communicate with the browser extension. This mode:

- Uses stdio-based Native Messaging protocol (4-byte length prefix + JSON) 
- Runs as a separate process with no GUI
- Maintains a heartbeat file updated every 2 seconds for connection status detection
- Automatically registers with Chrome/Edge/Firefox on startup

### Monitoring System
Uses a publisher-subscriber architecture coordinated by `MonitoringManager`:

- **2-second polling interval**: Captures process snapshots using platform-specific APIs
- **ProcessEventSubscriber**: Detects process start/stop events and logs to database
- **BlocklistSubscriber**: Terminates blocked processes in real-time
- **Self-healing**: Automatic panic recovery and restart capabilities

## Features

### Process Monitoring
- **Real-time tracking**: Monitors all running processes every 2 seconds
- **Lifecycle logging**: Records process start/stop events with timestamps, PIDs, and executable paths
- **Smart filtering**: Excludes system processes and applies deduplication to reduce noise
- **Historical data**: Query process activity logs with time-range filtering

### Application Blocking
- **Instant termination**: Blocked applications are killed within 2 seconds of detection
- **Persistent enforcement**: Blocklist is checked continuously, preventing blocked apps from running
- **Case-insensitive matching**: Process names are normalized for reliable blocking 

### Web Activity Monitoring
- **Browser extension integration**: Chrome extension captures all page visits via content scripts
- **Domain extraction**: URLs are automatically parsed to extract and store domains
- **Metadata collection**: Stores page titles and favicon URLs for rich display
- **Real-time sync**: Web events are logged immediately via native messaging protocol  

### Website Blocking
- **Extension-enforced blocking**: Blocklist is pushed to the browser extension for client-side enforcement
- **Proactive updates**: Changes to the blocklist are automatically synced to the extension every 500ms
- **Multi-browser support**: Works with Chrome, Edge, and Firefox

### Screen Time Tracking
- **Active window monitoring**: Tracks which application has focus using platform-specific APIs  
- **Buffered writes**: Accumulates time in memory and flushes to database periodically to reduce I/O
- **Per-application aggregation**: Groups screen time by executable path for accurate reporting

### Desktop GUI
- **Native window**: Frameless Wails application with custom title bar
- **Svelte frontend**: Modern, reactive UI with Bootstrap styling
- **Multiple views**: Leaderboards, activity logs, blocklist management, and settings
- **Extension status**: Real-time indicator showing browser extension connection state
- **Background operation**: Continues monitoring when window is hidden

### Reliability Features
- **Panic recovery**: All critical goroutines include panic handlers with stack trace logging 
- **Comprehensive logging**: Separate log files for GUI (`Veda_debug.log`) and native host (`native_host.log`)
- **Graceful degradation**: Database failures don't crash the application
- **Single instance lock**: Prevents multiple GUI instances from running simultaneously  
- **Autostart support**: Configures Windows registry for automatic startup with `--background` flag

## How to Build

### Prerequisites
1. All packages listed in `flake.nix` (recommended: use direnv)
2. Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

### Build Steps

```bash
make build
```

The compiled executable will be created in the `build/bin` directory.

## System Requirements

- **OS**: Windows 10 or later
- **Browser**: Chrome, Edge, or Firefox (for web monitoring)
- **Permissions**: Admin-level access
