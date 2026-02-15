package main

import (
	"strings"
	"testing"

	"mrktr/idea"
	"mrktr/types"

	"github.com/charmbracelet/lipgloss"
)

func TestLayoutOverheadConsistency(t *testing.T) {
	if layoutOverhead != 14 {
		t.Fatalf("expected layoutOverhead to be 14, got %d", layoutOverhead)
	}

	m := newTestModel()
	m.width = 120
	m.height = 32
	m.results = makeListings(25)
	m.selectedIndex = 0
	m.reveal.Rows = len(m.results)
	m.reveal.Revealing = false
	m.statsReveal.Revealed = 6

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

func TestStatsPanelShowsSkeletonWhileLoadingWithNoResults(t *testing.T) {
	m := newTestModel()
	m.loading = true
	m.statsAnim.SkeletonFrame = 1

	out := m.renderStatsPanel(44, 6)
	if !strings.Contains(out, "Results:") || !strings.Contains(out, "Average:") {
		t.Fatalf("expected skeleton placeholders while loading, got: %s", out)
	}
}

func TestStatsPanelShowsDeltaIndicators(t *testing.T) {
	m := newTestModel()
	m.results = makeListings(5)
	m.extendedStats.Count = 5
	m.extendedStats.Min = 100
	m.extendedStats.Max = 200
	m.extendedStats.Average = 150
	m.extendedStats.Median = 155
	m.extendedStats.P25 = 120
	m.extendedStats.P75 = 180
	m.extendedStats.Spread = "Moderate"
	m.statsAnim.DeltaTotal = 10
	m.statsAnim.DeltaTicks = 10
	m.statsAnim.DeltaAvg = 12.5
	m.statsAnim.DeltaMin = -5
	m.statsAnim.DeltaMax = 8
	m.statsAnim.DeltaMedian = 4
	m.statsAnim.DeltaP25 = 2
	m.statsAnim.DeltaP75 = -3
	m.statsReveal.Revealed = 5

	out := m.renderStatsPanel(60, 6)
	if !strings.Contains(out, "↑$12.50") || !strings.Contains(out, "↓$5.00") {
		t.Fatalf("expected rendered delta indicators, got: %s", out)
	}
}

func TestStatsPanelDeltaFadeEndStateHidesIndicators(t *testing.T) {
	m := newTestModel()
	m.results = makeListings(5)
	m.extendedStats.Count = 5
	m.extendedStats.Min = 100
	m.extendedStats.Max = 200
	m.extendedStats.Average = 150
	m.extendedStats.Median = 155
	m.extendedStats.P25 = 120
	m.extendedStats.P75 = 180
	m.extendedStats.Spread = "Moderate"
	m.statsAnim.DeltaTotal = 10
	m.statsAnim.DeltaTicks = 0
	m.statsAnim.DeltaAvg = 12.5
	m.statsAnim.DeltaMin = -5
	m.statsReveal.Revealed = 6

	out := m.renderStatsPanel(60, 7)
	if strings.Contains(out, "↑$") || strings.Contains(out, "↓$") {
		t.Fatalf("expected no delta indicators after fade completes, got: %s", out)
	}
}

func TestCurrentAnimatedStatsInterpolatesMarketCounts(t *testing.T) {
	m := newTestModel()

	from := idea.ExtendedStatistics{
		Statistics: types.Statistics{
			Count:   6,
			Min:     80,
			Max:     300,
			Average: 140,
			Median:  130,
		},
		SoldCount:   1,
		ActiveCount: 5,
		SoldAvg:     100,
		ActiveAvg:   150,
		PlatformStats: map[string]idea.PlatformStat{
			"eBay": {Count: 4, Average: 120, Min: 80, Max: 180},
		},
	}
	to := idea.ExtendedStatistics{
		Statistics: types.Statistics{
			Count:   10,
			Min:     90,
			Max:     420,
			Average: 210,
			Median:  205,
		},
		SoldCount:   4,
		ActiveCount: 6,
		SoldAvg:     220,
		ActiveAvg:   200,
		PlatformStats: map[string]idea.PlatformStat{
			"eBay":   {Count: 5, Average: 180, Min: 90, Max: 260},
			"Amazon": {Count: 5, Average: 240, Min: 150, Max: 420},
		},
	}

	m.extendedStats = to
	m.statsAnim.FromStats = from
	m.statsAnim.ToStats = to
	m.statsAnim.ValueTweenOn = true
	m.statsAnim.ValueSteps = 10
	m.statsAnim.ValueStep = 4

	got := m.currentAnimatedStats()
	if got.SoldCount <= from.SoldCount || got.SoldCount > to.SoldCount {
		t.Fatalf("expected sold count to interpolate from %d toward %d, got %d", from.SoldCount, to.SoldCount, got.SoldCount)
	}
	if got.ActiveCount < from.ActiveCount || got.ActiveCount > to.ActiveCount {
		t.Fatalf("expected active count to interpolate from %d toward %d, got %d", from.ActiveCount, to.ActiveCount, got.ActiveCount)
	}

	amazon, ok := got.PlatformStats["Amazon"]
	if !ok {
		t.Fatal("expected interpolated market stats to include incoming platform")
	}
	if amazon.Count <= 0 || amazon.Count > to.PlatformStats["Amazon"].Count {
		t.Fatalf("expected Amazon count to interpolate toward %d, got %d", to.PlatformStats["Amazon"].Count, amazon.Count)
	}
}
