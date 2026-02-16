package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFileHistoryStoreSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	store := NewFileHistoryStoreAt(path)

	in := []HistoryEntry{
		{Query: "ps5", Timestamp: time.Date(2026, 2, 10, 9, 0, 0, 0, time.UTC), ResultCount: 14},
		{Query: "switch", Timestamp: time.Date(2026, 2, 9, 8, 0, 0, 0, time.UTC), ResultCount: 22},
	}
	if err := store.Save(in); err != nil {
		t.Fatalf("save history: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(got))
	}
	if got[0].Query != "ps5" || got[0].ResultCount != 14 {
		t.Fatalf("unexpected first history entry: %+v", got[0])
	}
}

func TestNewFileHistoryStoreBuildsDefaultPath(t *testing.T) {
	store, err := NewFileHistoryStore()
	if err != nil {
		t.Fatalf("expected default history store to initialize, got %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil history store")
	}
}

func TestNormalizeHistoryEntriesDedupesAndTrims(t *testing.T) {
	entries := []HistoryEntry{
		{Query: "ps5", Timestamp: time.Now().UTC()},
		{Query: "PS5", Timestamp: time.Now().UTC()},
		{Query: "switch", Timestamp: time.Now().UTC()},
		{Query: "  ", Timestamp: time.Now().UTC()},
	}

	got := normalizeHistoryEntries(entries)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique entries, got %d", len(got))
	}
	if got[0].Query != "ps5" || got[1].Query != "switch" {
		t.Fatalf("unexpected normalized order: %+v", got)
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		ts   time.Time
		want string
	}{
		{name: "just now", ts: now.Add(-30 * time.Second), want: "just now"},
		{name: "minutes", ts: now.Add(-5 * time.Minute), want: "5m ago"},
		{name: "hours", ts: now.Add(-2 * time.Hour), want: "2h ago"},
		{name: "days", ts: now.Add(-3 * 24 * time.Hour), want: "3d ago"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatRelativeTime(tc.ts, now)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
