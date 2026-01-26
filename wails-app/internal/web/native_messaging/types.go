package native_messaging

import "encoding/json"

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
