package idea

import (
	"mrktr/types"
	"strings"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
)

func TestRenderMarketBody(t *testing.T) {
	stats := CalculateExtendedStats([]types.Listing{
		{Platform: "eBay", Status: "Sold", Price: 100},
		{Platform: "eBay", Status: "Active", Price: 120},
		{Platform: "Mercari", Status: "Active", Price: 90},
		{Platform: "Amazon", Status: "Sold", Price: 150},
	})

	lines := RenderMarketBody(stats, 70, 6)
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 market lines, got %d", len(lines))
	}
	if !strings.Contains(lines[len(lines)-1], "Sell-through:") {
		t.Fatalf("expected sell-through line at bottom, got %q", lines[len(lines)-1])
	}
}

func TestRenderMarketBodyNoPlatforms(t *testing.T) {
	lines := RenderMarketBody(ExtendedStatistics{}, 50, 4)
	if len(lines) != 2 {
		t.Fatalf("expected 2 fallback lines, got %d", len(lines))
	}
	if !strings.Contains(strings.ToLower(lines[1]), "insufficient") {
		t.Fatalf("expected insufficient data fallback, got %q", lines[1])
	}
}

func TestRenderMarketBodyClipsStatusLinesOnNarrowWidths(t *testing.T) {
	stats := CalculateExtendedStats([]types.Listing{
		{Platform: "eBay", Status: "Sold", Price: 100},
		{Platform: "eBay", Status: "Active", Price: 120},
		{Platform: "Mercari", Status: "Active", Price: 90},
		{Platform: "Amazon", Status: "Sold", Price: 150},
	})

	width := 30
	lines := RenderMarketBody(stats, width, 4)
	if len(lines) == 0 {
		t.Fatal("expected rendered lines")
	}
	for i, line := range lines {
		if got := xansi.StringWidth(line); got > width {
			t.Fatalf("line %d exceeds width (%d > %d): %q", i+1, got, width, line)
		}
	}
	if !strings.Contains(lines[len(lines)-2], "S:") {
		t.Fatalf("expected compact sold/active status line, got %q", lines[len(lines)-2])
	}
}
