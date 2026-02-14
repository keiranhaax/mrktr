package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := loadDotEnvFile(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load .env: %v\n", err)
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
