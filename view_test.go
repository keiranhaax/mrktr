package main

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestLayoutOverheadConsistency(t *testing.T) {
	if layoutOverhead != 14 {
		t.Fatalf("expected layoutOverhead to be 14, got %d", layoutOverhead)
	}

	m := NewModel()
	m.width = 120
	m.height = 32
	m.results = makeListings(25)
	m.selectedIndex = 0
	m.revealedRows = len(m.results)
	m.revealing = false
	m.statsRevealed = 6

	resultsHeight := max(4, m.height-layoutOverhead)
	if got, want := m.visibleResultRows(), max(1, resultsHeight-2); got != want {
		t.Fatalf("expected visible rows %d, got %d", want, got)
	}
}

func TestSparklineWidth(t *testing.T) {
	prices := []float64{10, 12, 14, 13, 15, 17, 16, 18, 20, 19}
	width := 24

	sparkline := renderSparkline(prices, width)
	if got := lipgloss.Width(sparkline); got != width {
		t.Fatalf("expected sparkline visual width %d, got %d", width, got)
	}
}
