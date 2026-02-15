package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const historyMaxEntries = 20

// HistoryEntry stores one past search.
type HistoryEntry struct {
	Query       string    `json:"query"`
	Timestamp   time.Time `json:"timestamp"`
	ResultCount int       `json:"result_count"`
}

// HistoryStore persists search history between runs.
type HistoryStore interface {
	Load() ([]HistoryEntry, error)
	Save(entries []HistoryEntry) error
}

type FileHistoryStore struct {
	path string
}

func NewFileHistoryStore() *FileHistoryStore {
	path, _ := defaultHistoryPath()
	return &FileHistoryStore{path: path}
}

func NewFileHistoryStoreAt(path string) *FileHistoryStore {
	return &FileHistoryStore{path: path}
}

func (s *FileHistoryStore) Load() ([]HistoryEntry, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return nil, fmt.Errorf("history store path is empty")
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode history: %w", err)
	}

	return normalizeHistoryEntries(entries), nil
}

func (s *FileHistoryStore) Save(entries []HistoryEntry) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return fmt.Errorf("history store path is empty")
	}

	normalized := normalizeHistoryEntries(entries)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}

	body, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return fmt.Errorf("encode history: %w", err)
	}
	body = append(body, '\n')

	if err := os.WriteFile(s.path, body, 0o644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

func defaultHistoryPath() (string, error) {
	if configDir, err := os.UserConfigDir(); err == nil && strings.TrimSpace(configDir) != "" {
		return filepath.Join(configDir, "mrktr", "history.json"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(homeDir) == "" {
		return "", fmt.Errorf("resolve history path: %w", err)
	}

	return filepath.Join(homeDir, ".config", "mrktr", "history.json"), nil
}

func normalizeHistoryEntries(entries []HistoryEntry) []HistoryEntry {
	if len(entries) == 0 {
		return []HistoryEntry{}
	}

	out := make([]HistoryEntry, 0, min(len(entries), historyMaxEntries))
	seen := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		query := strings.TrimSpace(entry.Query)
		if query == "" {
			continue
		}
		key := strings.ToLower(query)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		if entry.Timestamp.IsZero() {
			entry.Timestamp = time.Now().UTC()
		}
		entry.Query = query
		out = append(out, entry)
		if len(out) == historyMaxEntries {
			break
		}
	}

	return out
}

func formatRelativeTime(ts, now time.Time) string {
	if ts.IsZero() {
		return ""
	}

	if now.Before(ts) {
		return "just now"
	}

	d := now.Sub(ts)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return ts.Format("2006-01-02")
	}
}
