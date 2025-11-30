package api

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
	"wails-app/internal/data"
)

// BlockApps adds one or more applications to the blocklist.
func (s *Server) BlockApps(names []string) error {
	list, err := data.LoadAppBlocklist()
	if err != nil {
		return err
	}

	for _, name := range names {
		lowerName := strings.ToLower(name)
		if !slices.Contains(list, lowerName) {
			list = append(list, lowerName)
		}
	}

	return data.SaveAppBlocklist(list)
}

// UnblockApps removes one or more applications from the blocklist.
func (s *Server) UnblockApps(names []string) error {
	list, err := data.LoadAppBlocklist()
	if err != nil {
		return err
	}

	for _, name := range names {
		lowerName := strings.ToLower(name)
		list = slices.DeleteFunc(list, func(item string) bool {
			return item == lowerName
		})
	}

	return data.SaveAppBlocklist(list)
}

// GetAppBlocklist returns the list of blocked applications with their details.
func (s *Server) GetAppBlocklist() ([]data.BlockedAppDetail, error) {
	return data.GetBlockedAppsWithDetails(s.db)
}

// ClearAppBlocklist removes all applications from the blocklist.
func (s *Server) ClearAppBlocklist() error {
	return data.ClearAppBlocklist()
}

// SaveAppBlocklist saves the current application blocklist to a file for export.
func (s *Server) SaveAppBlocklist() ([]byte, error) {
	list, err := data.LoadAppBlocklist()
	if err != nil {
		return nil, err
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	return json.MarshalIndent(header, "", "  ")
}

// LoadAppBlocklist loads an application blocklist from an uploaded file and merges it with the existing blocklist.
func (s *Server) LoadAppBlocklist(content []byte) error {
	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}

	err := json.Unmarshal(content, &newEntries)
	if err != nil {
		err2 := json.Unmarshal(content, &savedList)
		if err2 != nil {
			return fmt.Errorf("invalid JSON format in uploaded file")
		}
		newEntries = savedList.Blocked
	}

	existingList, err := data.LoadAppBlocklist()
	if err != nil {
		return err
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	return data.SaveAppBlocklist(existingList)
}
