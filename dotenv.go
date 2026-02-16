package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var allowedDotEnvKeys = map[string]struct{}{
	"BRAVE_API_KEY":       {},
	"TAVILY_API_KEY":      {},
	"FIRECRAWL_API_KEY":   {},
	"MRKTR_LOW_POWER":     {},
	"MRKTR_REDUCE_MOTION": {},
}

// loadDotEnvFile loads KEY=VALUE pairs from a dotenv-style file.
// Existing process env values are preserved and take precedence.
func loadDotEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open dotenv file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	ignoredKeys := map[string]struct{}{}
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := parseDotEnvLine(line)
		if !ok {
			continue
		}
		if !isAllowedDotEnvKey(key) {
			ignoredKeys[key] = struct{}{}
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %q from .env line %d: %w", key, lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan dotenv file: %w", err)
	}
	if len(ignoredKeys) > 0 {
		keys := make([]string, 0, len(ignoredKeys))
		for key := range ignoredKeys {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		return fmt.Errorf("ignored unknown .env keys: %s", strings.Join(keys, ", "))
	}

	return nil
}

func isAllowedDotEnvKey(key string) bool {
	_, ok := allowedDotEnvKeys[strings.TrimSpace(key)]
	return ok
}

func parseDotEnvLine(line string) (key, value string, ok bool) {
	idx := strings.IndexRune(line, '=')
	if idx <= 0 {
		return "", "", false
	}

	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", false
	}

	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
			return key, value, true
		}
	}

	// Support trailing comments for unquoted values: KEY=value # comment
	if i := strings.Index(value, " #"); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}

	return key, value, true
}
