package api

import (
	"fmt"
	"wails-app/internal/auth"
	"wails-app/internal/data"
)

// GetIsAuthenticated checks if the user is authenticated.
func (s *Server) GetIsAuthenticated() bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return s.IsAuthenticated
}

// Logout handles the user logout.
func (s *Server) Logout() {
	s.Mu.Lock()
	s.IsAuthenticated = false
	s.Mu.Unlock()
}

// HasPassword checks if a password has been set for the application.
func (s *Server) HasPassword() (bool, error) {
	cfg, err := data.LoadConfig()
	if err != nil {
		return false, err
	}
	return cfg.PasswordHash != "", nil
}

// Login handles the user login.
func (s *Server) Login(password string) (bool, error) {
	cfg, err := data.LoadConfig()
	if err != nil {
		return false, err
	}

	if auth.CheckPasswordHash(password, cfg.PasswordHash) {
		s.Mu.Lock()
		s.IsAuthenticated = true
		s.Mu.Unlock()
		return true, nil
	}
	return false, nil
}

// SetPassword handles the initial password setup.
func (s *Server) SetPassword(password string) error {
	cfg, err := data.LoadConfig()
	if err != nil {
		return err
	}

	if cfg.PasswordHash != "" {
		return fmt.Errorf("password already set")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	cfg.PasswordHash = hash
	if err := cfg.Save(); err != nil {
		return err
	}

	s.Mu.Lock()
	s.IsAuthenticated = true
	s.Mu.Unlock()
	return nil
}
