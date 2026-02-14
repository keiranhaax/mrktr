package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotEnvLine(t *testing.T) {
	tests := []struct {
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{line: "BRAVE_API_KEY=abc123", wantKey: "BRAVE_API_KEY", wantValue: "abc123", wantOK: true},
		{line: "TAVILY_API_KEY=\"quoted value\"", wantKey: "TAVILY_API_KEY", wantValue: "quoted value", wantOK: true},
		{line: "FIRECRAWL_API_KEY='single quoted'", wantKey: "FIRECRAWL_API_KEY", wantValue: "single quoted", wantOK: true},
		{line: "BRAVE_API_KEY=abc # trailing", wantKey: "BRAVE_API_KEY", wantValue: "abc", wantOK: true},
		{line: "not_a_pair", wantOK: false},
	}

	for _, tc := range tests {
		key, value, ok := parseDotEnvLine(tc.line)
		if ok != tc.wantOK {
			t.Fatalf("line %q: expected ok=%v got %v", tc.line, tc.wantOK, ok)
		}
		if !ok {
			continue
		}
		if key != tc.wantKey || value != tc.wantValue {
			t.Fatalf("line %q: expected (%q,%q), got (%q,%q)", tc.line, tc.wantKey, tc.wantValue, key, value)
		}
	}
}

func TestLoadDotEnvFileSetsValuesAndPreservesExistingEnv(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".env")
	content := "BRAVE_API_KEY=from_file\nTAVILY_API_KEY=from_file\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	const (
		braveKey  = "BRAVE_API_KEY"
		tavilyKey = "TAVILY_API_KEY"
	)
	t.Setenv(braveKey, "already_set")
	t.Setenv(tavilyKey, "")

	if err := os.Unsetenv(tavilyKey); err != nil {
		t.Fatalf("unset %s: %v", tavilyKey, err)
	}

	if err := loadDotEnvFile(path); err != nil {
		t.Fatalf("load dotenv: %v", err)
	}

	if got := os.Getenv(braveKey); got != "already_set" {
		t.Fatalf("expected existing env to be preserved, got %q", got)
	}
	if got := os.Getenv(tavilyKey); got != "from_file" {
		t.Fatalf("expected tavily key from file, got %q", got)
	}
}

func TestLoadDotEnvFileMissingFileIsNotError(t *testing.T) {
	if err := loadDotEnvFile(filepath.Join(t.TempDir(), "does-not-exist.env")); err != nil {
		t.Fatalf("expected missing dotenv file to be ignored, got %v", err)
	}
}
