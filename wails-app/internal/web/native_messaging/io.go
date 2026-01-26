package native_messaging

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
)

// sendResponse sends a JSON response to the browser extension via standard output.
// It follows the native messaging protocol: a 4-byte length prefix followed by the JSON body.
func sendResponse(msg interface{}) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(bytes))); err != nil {
		log.Printf("Error writing length to stdout: %v", err)
		return
	}
	if _, err := os.Stdout.Write(bytes); err != nil {
		log.Printf("Error writing message to stdout: %v", err)
	}
}
