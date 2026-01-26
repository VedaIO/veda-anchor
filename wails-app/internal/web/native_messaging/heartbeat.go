package native_messaging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// updateHeartbeat updates a local file with the current timestamp.
// This is used by the GUI to verify that the native messaging host (and thus the extension) is active.
func updateHeartbeat() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	heartbeatPath := filepath.Join(cacheDir, "ProcGuard", "extension_heartbeat")
	// Ensure directory exists
	_ = os.MkdirAll(filepath.Dir(heartbeatPath), 0755)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	_ = os.WriteFile(heartbeatPath, []byte(timestamp), 0644)
}
