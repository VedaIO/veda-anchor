## Frontend Architecture

The frontend is a Single Page Application (SPA) built with **Svelte 5** and **TypeScript**, running inside Wails' WebView2 container on Windows.

### Technology Stack

- **Framework**: Svelte 5 with TypeScript
- **UI Library**: Bootstrap 5.3.8
- **Date Picker**: flatpickr 4.6.13
- **Build Tool**: Vite 5.0
- **Linting**: ESLint + Prettier

### Project Structure

```
frontend/
├── src/
│   ├── App.svelte           # Main entry point, hash-based routing
│   └── lib/
│       ├── extensionStore.ts # Extension status polling
│       ├── WebManagement.svelte # Web features container
│       ├── WebLog.svelte     # Web history view
│       ├── Welcome.svelte    # Dashboard with screen time
│       ├── Settings.svelte   # App configuration
│       ├── Login.svelte      # Authentication
│       ├── Toast.svelte      # Global notifications
│       ├── GlobalTitleBar.svelte # Custom window controls
│       └── modalStore.ts     # Password confirmation modal
├── dist/                    # Build output (embedded in Go binary)
└── wailsjs/                 # Auto-generated Wails bindings (DO NOT TOUCH)
```

### Backend Communication

The frontend communicates with Go backend via the `window.go` object:

```typescript
// Example API calls
await window.go.main.App.BlockApps(['appname'])
await window.go.main.App.CheckChromeExtension()
await window.go.main.App.GetWebLogs(query, since, until)
```

### Key Components

1. **App.svelte** - Root component managing:
   - Hash-based routing with route map
   - Authentication guard on mount
   - Global navbar with navigation
   - Password confirmation modal

2. **Extension Detection** (`extensionStore.ts`):
   - Polls every 3 seconds to check browser extension status
   - Reads heartbeat file timestamp to determine connectivity

3. **Web Management** (`WebManagement.svelte`):
   - Three tabs: Leaderboard, Web Logs, Blocklist
   - Shows install prompt when extension disconnected 

4. **Web Logs** (`WebLog.svelte`):
   - Displays URL history with metadata enrichment
   - Supports filtering by time range and search query
   - Allows bulk blocking of domains 

5. **Authentication** (`Login.svelte`):
   - Handles both login and password creation
   - Updates global auth store on success

### State Management

Uses Svelte stores for global state:
- `authStore` - Authentication status
- `router` - Current navigation path
- `extensionStore` - Browser extension connectivity
- `toastStore` - Global notifications
- `modalStore` - Password confirmation modal

### Important Patterns

1. **Hash-based Routing**: Client-side routing using Svelte stores, suitable for WebView2 context
2. **Extension Polling**: Continuous 3-second polling to detect browser extension connectivity
3. **Global Modal**: Single password confirmation modal reused across the application
4. **Toast Notifications**: Bootstrap-based toast system for user feedback
