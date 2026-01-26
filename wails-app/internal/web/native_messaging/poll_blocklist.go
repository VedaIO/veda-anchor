package native_messaging

import (
	"log"
	"reflect"
	"time"
	blocklist "wails-app/internal/blocklist/web"
)

const (
	// pollInterval is the interval at which the web blocklist is polled for changes.
	pollInterval = 500 * time.Millisecond
)

// pollWebBlocklist periodically checks for changes in the web blocklist and sends updates to the extension.
func pollWebBlocklist() {
	var lastBlocklist []string
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Load blocklist directly from data package
		list, err := blocklist.LoadWebBlocklist()
		if err != nil {
			log.Printf("Failed to get web blocklist: %v", err)
			continue
		}

		// Only send an update if the blocklist has changed.
		if !reflect.DeepEqual(list, lastBlocklist) {
			lastBlocklist = list
			sendResponse(map[string]interface{}{
				"type":    "web_blocklist",
				"payload": list,
			})
		}
	}
}
