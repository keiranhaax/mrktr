package idea

import (
	"mrktr/types"
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

func TestRenderDistributionBody(t *testing.T) {
	stats := CalculateExtendedStats([]types.Listing{
		{Price: 100}, {Price: 120}, {Price: 140}, {Price: 220}, {Price: 260}, {Price: 400}, {Price: 1000},
	})

	lines := RenderDistributionBody(stats, 70, 6)
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[len(lines)-1], "P10:") || !strings.Contains(lines[len(lines)-1], "P90:") {
		t.Fatalf("expected percentile line at bottom, got %q", lines[len(lines)-1])
	}
}

func TestRenderDistributionBodyNoHistogram(t *testing.T) {
	lines := RenderDistributionBody(ExtendedStatistics{}, 50, 4)
	if len(lines) != 2 {
		t.Fatalf("expected 2 fallback lines, got %d", len(lines))
	}
	if !strings.Contains(strings.ToLower(lines[1]), "insufficient") {
		t.Fatalf("expected insufficient data fallback, got %q", lines[1])
	}
}

func TestRenderDistributionBodyClipsPercentileLineOnNarrowWidths(t *testing.T) {
	stats := CalculateExtendedStats([]types.Listing{
		{Price: 100}, {Price: 120}, {Price: 140}, {Price: 220}, {Price: 260}, {Price: 400}, {Price: 1000},
	})

	width := 32
	lines := RenderDistributionBody(stats, width, 4)
	if len(lines) == 0 {
		t.Fatal("expected rendered lines")
	}
	for i, line := range lines {
		if got := xansi.StringWidth(line); got > width {
			t.Fatalf("line %d exceeds width (%d > %d): %q", i+1, got, width, line)
		}
	}
	last := lines[len(lines)-1]
	if !strings.Contains(last, "Med:") {
		t.Fatalf("expected compact percentile line to include Med:, got %q", last)
	}
}
