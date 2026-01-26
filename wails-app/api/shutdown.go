package api

import (
	"os"
	"time"
	"wails-app/internal/web/native_messaging"
)

// Stop handles the graceful shutdown of the application.
func (s *Server) Shutdown() {
	s.Logger.Println("Received stop request. Shutting down...")

	// Send stopping message to extension to prevent reconnection
	native_messaging.Stop()

	go func() {
		time.Sleep(1 * time.Second)
		s.Logger.Close()
		if err := s.db.Close(); err != nil {
			s.Logger.Printf("Failed to close database: %v", err)
		}
		os.Exit(0)
	}()
}
