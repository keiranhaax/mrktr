package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mrktr/types"
)

func ExportCSV(path string, listings []types.Listing) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv export: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.Write([]string{"platform", "price", "condition", "status", "title", "url"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, listing := range listings {
		row := []string{
			listing.Platform,
			fmt.Sprintf("%.2f", listing.Price),
			listing.Condition,
			listing.Status,
			listing.Title,
			listing.URL,
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flush csv export: %w", err)
	}
	return nil
}

func ExportJSON(path string, listings []types.Listing) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create json export: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(listings); err != nil {
		return fmt.Errorf("encode json export: %w", err)
	}
	return nil
}

func BuildExportPath(homeDir, query, ext string, now time.Time) string {
	sanitized := sanitizeFilename(query)
	if sanitized == "" {
		sanitized = "results"
	}
	if ext == "" {
		ext = "csv"
	}
	name := fmt.Sprintf("mrktr-export-%s-%s.%s", sanitized, now.Format("20060102-150405"), ext)
	return filepath.Join(homeDir, name)
}

func sanitizeFilename(query string) string {
	trimmed := strings.TrimSpace(strings.ToLower(query))
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	prevDash := false
	for _, r := range trimmed {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	return out
}
