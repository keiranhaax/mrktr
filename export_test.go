package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mrktr/types"
)

func TestExportCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	data := []types.Listing{
		{Platform: "eBay", Price: 99.99, Condition: "Used", Status: "Active", Title: "PS5", URL: "https://example.com/1"},
	}

	if err := ExportCSV(path, data); err != nil {
		t.Fatalf("export csv: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported csv: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "platform,price,condition,status,title,url") {
		t.Fatalf("expected csv header, got %q", text)
	}
	if !strings.Contains(text, "eBay,99.99,Used,Active,PS5,https://example.com/1") {
		t.Fatalf("expected listing row in csv, got %q", text)
	}
}

func TestExportJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	data := []types.Listing{
		{Platform: "Mercari", Price: 120.0, Condition: "New", Status: "Sold", Title: "Switch", URL: "https://example.com/2"},
	}

	if err := ExportJSON(path, data); err != nil {
		t.Fatalf("export json: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read exported json: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "\"Platform\": \"Mercari\"") {
		t.Fatalf("expected platform in json, got %q", text)
	}
}

func TestBuildExportPathSanitizesQuery(t *testing.T) {
	now := time.Date(2026, 2, 15, 12, 34, 56, 0, time.UTC)
	path := BuildExportPath("/tmp", " Nintendo Switch / OLED ", "csv", now)
	if !strings.Contains(path, "mrktr-export-nintendo-switch-oled-20260215-123456.csv") {
		t.Fatalf("unexpected export path: %s", path)
	}
}
