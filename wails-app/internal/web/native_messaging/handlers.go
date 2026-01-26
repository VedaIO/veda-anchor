package native_messaging

import (
	"encoding/json"
	"log"
	"time"
	blocklist "wails-app/internal/blocklist/web"
	"wails-app/internal/data/query"
	"wails-app/internal/data/write"
)

// handleRequest dispatches the incoming request to the appropriate handler logic.
func handleRequest(req Request) {
	log.Printf("Processing message type: %s", req.Type)

	switch req.Type {
	case "ping":
		sendResponse(map[string]string{"type": "pong"})

	case "log_url":
		// Handle URL logging
		var payload WebLogPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			log.Printf("Error unmarshalling log_url: %v", err)
			return
		}

		log.Printf("Logging URL: %s", payload.Url)
		// Write to DB
		if err := query.LogWebActivity(payload.Url, payload.Title, payload.VisitTime); err != nil {
			log.Printf("DB Error: %v", err)
		}

	case "log_web_metadata":
		var payload WebMetadataPayload
		if err := json.Unmarshal(req.Payload, &payload); err != nil {
			log.Printf("Error unmarshalling log_web_metadata payload: %v", err)
			return
		}

		// Log metadata directly to the database
		write.EnqueueWrite("INSERT OR REPLACE INTO web_metadata (domain, title, icon_url, timestamp) VALUES (?, ?, ?, ?)",
			payload.Domain, payload.Title, payload.IconURL, time.Now().Unix())

	case "get_web_blocklist":
		// Send blocklist
		bl, err := blocklist.LoadWebBlocklist()
		if err != nil {
			log.Printf("Error loading blocklist: %v", err)
			bl = []string{} // Send empty list on error
		}
		sendResponse(map[string]interface{}{
			"type":    "web_blocklist",
			"payload": bl,
		})
	case "add_to_web_blocklist":
		var domain string
		if err := json.Unmarshal(req.Payload, &domain); err != nil {
			log.Printf("Error unmarshalling add_to_web_blocklist payload: %v", err)
			return
		}
		if _, err := blocklist.AddWebsiteToBlocklist(domain); err != nil {
			log.Printf("Error adding to web blocklist: %v", err)
		}
	default:
		log.Printf("Unknown message type: %s", req.Type)
	}
}
