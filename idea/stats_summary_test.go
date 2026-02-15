package idea

import (
	"mrktr/types"
	"strings"
	"testing"
)

func TestRenderStatsTabs(t *testing.T) {
	got := RenderStatsTabs(StatsViewDistribution)
	if !strings.Contains(got, "[1:Sum]") || !strings.Contains(got, "[2:Dist]") || !strings.Contains(got, "[3:Mkt]") {
		t.Fatalf("expected all tab labels in rendered tabs, got %q", got)
	}
}

func TestRenderSummaryBody(t *testing.T) {
	stats := ExtendedStatistics{
		Statistics: types.Statistics{
			Count:   4,
			Min:     100,
			Max:     400,
			Average: 250,
			Median:  250,
		},
		P25:    175,
		P75:    325,
		Spread: "Moderate",
	}

	lines := RenderSummaryBody(stats, "▁▂▃▄▅", 60)
	if len(lines) != 5 {
		t.Fatalf("expected 5 summary lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "Results: 4") {
		t.Fatalf("expected first summary line to include count, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "Trend:") {
		t.Fatalf("expected second summary line to include trend label, got %q", lines[1])
	}
}
