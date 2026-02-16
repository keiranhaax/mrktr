package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := loadDotEnvFile(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load .env: %v\n", err)
	}
	if !hasAnyProviderKeyConfigured() {
		fmt.Fprintln(
			os.Stderr,
			"Warning: no live search providers configured. Set BRAVE_API_KEY, TAVILY_API_KEY, or FIRECRAWL_API_KEY.",
		)
	}

	// Create new program with our model
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

func hasAnyProviderKeyConfigured() bool {
	keys := []string{"BRAVE_API_KEY", "TAVILY_API_KEY", "FIRECRAWL_API_KEY"}
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return true
		}
	}
	return false
}
