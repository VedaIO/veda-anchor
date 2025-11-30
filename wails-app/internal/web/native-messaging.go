package web

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"time"
	"wails-app/internal/data"
)

const (
	// pollInterval is the interval at which the web blocklist is polled for changes.
	pollInterval = 500 * time.Millisecond
)

// WebLogPayload is the payload for the log_url message from the extension.
type WebLogPayload struct {
	Url       string `json:"url"`
	Title     string `json:"title"`
	VisitTime int64  `json:"visitTime"`
}

// WebMetadataPayload is the payload for the log_web_metadata message from the extension.
type WebMetadataPayload struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"iconUrl"`
}

// Request is a message received from the browser extension.
type Request struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Response is a message sent to the browser extension.
type Response struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Run starts the native messaging host loop
func Run() {
	// Setup logging to file (CRITICAL for debugging native messaging)
	cacheDir, _ := os.UserCacheDir()
	logDir := filepath.Join(cacheDir, "procguard", "logs")
	_ = os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "native_host.log")
	logFile, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer func() { _ = logFile.Close() }()
		log.SetOutput(logFile)
	}

	// Initialize Database (CRITICAL: Required for logging)
	if _, err := data.InitDB(); err != nil {
		log.Printf("CRITICAL: Failed to initialize database: %v", err)
		// We continue anyway, but DB writes will fail
	} else {
		log.Println("Database initialized successfully")
	}

	// Catch panics to see why it crashes
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CRITICAL PANIC in Native Host: %v\nStack: %s", r, string(debug.Stack()))
		}
	}()

	log.Println("=== NATIVE MESSAGING HOST STARTED ===")

	// Start blocklist poller (panic safe)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC in Blocklist Poller: %v", r)
			}
		}()
		pollWebBlocklist()
	}()

	// Start continuous heartbeat updater
	// This ensures the GUI knows we are alive even if no messages are flowing
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			updateHeartbeat()
		}
	}()

	// Main Message Loop
	for {
		log.Println("Waiting for message...")

		// 1. Read Message Length (4 bytes)
		var length uint32
		if err := binary.Read(os.Stdin, binary.LittleEndian, &length); err != nil {
			if err == io.EOF {
				log.Println("Chrome disconnected (EOF)")
				return
			}
			log.Printf("Error reading length: %v", err)
			return
		}

		// 2. Read Message Body
		msg := make([]byte, length)
		if _, err := io.ReadFull(os.Stdin, msg); err != nil {
			log.Printf("Error reading body: %v", err)
			return
		}

		log.Printf("Received message (%d bytes): %s", length, string(msg))

		// 3. Update Heartbeat File (for GUI detection)
		updateHeartbeat()

		// 4. Process Message
		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("JSON Error: %v", err)
			continue
		}

		log.Printf("Processing message type: %s", req.Type)

		switch req.Type {
		case "ping":
			sendResponse(map[string]string{"type": "pong"})

		case "log_url":
			// Handle URL logging
			var payload WebLogPayload
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling log_url: %v", err)
				continue
			}

			log.Printf("Logging URL: %s", payload.Url)
			// Write to DB
			if err := data.LogWebActivity(payload.Url, payload.Title, payload.VisitTime); err != nil {
				log.Printf("DB Error: %v", err)
			}

		case "log_web_metadata":
			var payload WebMetadataPayload
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling log_web_metadata payload: %v", err)
				continue
			}

			// Log metadata directly to the database
			data.EnqueueWrite("INSERT OR REPLACE INTO web_metadata (domain, title, icon_url, timestamp) VALUES (?, ?, ?, ?)",
				payload.Domain, payload.Title, payload.IconURL, time.Now().Unix())

		case "get_web_blocklist":
			// Send blocklist
			blocklist, err := data.LoadWebBlocklist()
			if err != nil {
				log.Printf("Error loading blocklist: %v", err)
				blocklist = []string{} // Send empty list on error
			}
			sendResponse(map[string]interface{}{
				"type":    "web_blocklist",
				"payload": blocklist,
			})
		case "add_to_web_blocklist":
			var domain string
			if err := json.Unmarshal(req.Payload, &domain); err != nil {
				log.Printf("Error unmarshalling add_to_web_blocklist payload: %v", err)
				continue
			}
			if _, err := data.AddWebsiteToBlocklist(domain); err != nil {
				log.Printf("Error adding to web blocklist: %v", err)
			}
		default:
			log.Printf("Unknown message type: %s", req.Type)
		}

		log.Println("Message processed successfully")
	}
}

// pollWebBlocklist periodically checks for changes in the web blocklist and sends updates to the extension.
func pollWebBlocklist() {
	var lastBlocklist []string
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Load blocklist directly from data package
		list, err := data.LoadWebBlocklist()
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

// Stop sends a stopping message to the extension to prevent it from reconnecting.
func Stop() {
	sendResponse(map[string]interface{}{
		"type":    "stopping",
		"payload": nil,
	})
}

func updateHeartbeat() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	heartbeatPath := filepath.Join(cacheDir, "procguard", "extension_heartbeat")
	// Ensure directory exists
	_ = os.MkdirAll(filepath.Dir(heartbeatPath), 0755)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(heartbeatPath, []byte(timestamp), 0644)
}

func sendResponse(msg interface{}) {
	bytes, _ := json.Marshal(msg)
	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(bytes))); err != nil {
		log.Printf("Error writing length to stdout: %v", err)
		return
	}
	if _, err := os.Stdout.Write(bytes); err != nil {
		log.Printf("Error writing message to stdout: %v", err)
	}
}
