package api

import (
	"time"
	"wails-app/internal/data"
)

// AppLeaderboardItem represents a single item in the application leaderboard.
type AppLeaderboardItem struct {
	Rank        int    `json:"rank"`
	Name        string `json:"name"`        // Display name (commercial name if available)
	ProcessName string `json:"processName"` // Actual process name for blocking
	Icon        string `json:"icon"`
	Count       int    `json:"count"`
}

// WebLeaderboardItem represents a single item in the web leaderboard.
type WebLeaderboardItem struct {
	Rank   int    `json:"rank"`
	Domain string `json:"domain"`
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Count  int    `json:"count"`
}

// GetAppLeaderboard retrieves the top 10 most used applications and returns them as a leaderboard.
func (s *Server) GetAppLeaderboard(since, until string) ([]AppLeaderboardItem, error) {
	return s.getAppLeaderboard(since, until)
}

// getAppLeaderboard retrieves the top 10 most used applications from the database.
func (s *Server) getAppLeaderboard(since, until string) ([]AppLeaderboardItem, error) {
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = data.ParseTime(since)
		if err != nil {
			return nil, err
		}
	}

	if until != "" {
		untilTime, err = data.ParseTime(until)
		if err != nil {
			return nil, err
		}
	}

	q := "SELECT process_name, COUNT(*) as count FROM app_events WHERE 1=1"
	args := make([]interface{}, 0)

	if !sinceTime.IsZero() {
		q += " AND start_time >= ?"
		args = append(args, sinceTime.Unix())
	}

	if !untilTime.IsZero() {
		q += " AND start_time <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " GROUP BY process_name ORDER BY count DESC LIMIT 10"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.Logger.Printf("Failed to close rows: %v", err)
		}
	}()

	var leaderboard []AppLeaderboardItem
	rank := 1
	for rows.Next() {
		var item AppLeaderboardItem
		item.Rank = rank
		var processName string
		if err := rows.Scan(&processName, &item.Count); err != nil {
			continue
		}

		// Store the actual process name for blocking
		item.ProcessName = processName
		item.Name = processName // Default display name

		// Enrich with icon and commercial name
		var exePath string
		row := s.db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1", processName)
		if err := row.Scan(&exePath); err == nil {
			commercialName, icon := s.getAppDetails(exePath)
			if commercialName != "" {
				item.Name = commercialName // Use commercial name for display
			}
			item.Icon = icon
		}

		leaderboard = append(leaderboard, item)
		rank++
	}

	return leaderboard, nil
}

// GetWebLeaderboard retrieves the top 10 most visited websites and returns them as a leaderboard.
func (s *Server) GetWebLeaderboard(since, until string) ([]WebLeaderboardItem, error) {
	return s.getWebLeaderboard(since, until)
}

// getWebLeaderboard retrieves the top 10 most visited websites from the database.
func (s *Server) getWebLeaderboard(since, until string) ([]WebLeaderboardItem, error) {
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = data.ParseTime(since)
		if err != nil {
			return nil, err
		}
	}

	if until != "" {
		untilTime, err = data.ParseTime(until)
		if err != nil {
			return nil, err
		}
	}

	q := `
		SELECT
			CASE
				WHEN INSTR(SUBSTR(url, INSTR(url, '//') + 2), '/') > 0
				THEN SUBSTR(url, INSTR(url, '//') + 2, INSTR(SUBSTR(url, INSTR(url, '//') + 2), '/') - 1)
				ELSE SUBSTR(url, INSTR(url, '//') + 2)
			END as domain,
			COUNT(*) as count
		FROM web_events
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if !sinceTime.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, sinceTime.Unix())
	}

	if !untilTime.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " GROUP BY domain ORDER BY count DESC LIMIT 10"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.Logger.Printf("Failed to close rows: %v", err)
		}
	}()

	var leaderboard []WebLeaderboardItem
	rank := 1
	for rows.Next() {
		var item WebLeaderboardItem
		item.Rank = rank
		if err := rows.Scan(&item.Domain, &item.Count); err != nil {
			continue
		}

		// Enrich with metadata
		if meta, err := data.GetWebMetadata(s.db, item.Domain); err == nil && meta != nil {
			item.Title = meta.Title
			item.Icon = meta.IconURL
		}

		leaderboard = append(leaderboard, item)
		rank++
	}

	return leaderboard, nil
}
