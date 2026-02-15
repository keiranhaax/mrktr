package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"mrktr/idea"
	"mrktr/types"

	tea "github.com/charmbracelet/bubbletea"
	xansi "github.com/charmbracelet/x/ansi"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
var updateGoldens = os.Getenv("UPDATE_GOLDEN") == "1"

func TestStatsPanelSnapshotsByWidth(t *testing.T) {
	widths := []int{40, 60, 80, 120}
	modes := []struct {
		name string
		mode idea.StatsViewMode
	}{
		{name: "summary", mode: idea.StatsViewSummary},
		{name: "distribution", mode: idea.StatsViewDistribution},
		{name: "market", mode: idea.StatsViewMarket},
	}

	for _, mode := range modes {
		for _, width := range widths {
			t.Run(fmt.Sprintf("%s-w%d", mode.name, width), func(t *testing.T) {
				m := statsFixtureModel()
				m.statsViewMode = mode.mode
				m.statsReveal.Revealed = 20

				out := m.renderStatsPanel(width, 9)
				assertNoVisualOverflow(t, out)
				plain := normalizeForGolden(stripANSI(out))
				assertGoldenFile(t, fmt.Sprintf("stats_%s_w%d.golden", mode.name, width), plain)
			})
		}
	}
}

func TestStatsPanelSnapshotTabSwitching(t *testing.T) {
	m := statsFixtureModel()
	m.focusedPanel = panelStats
	m = m.updateFocus()
	m.statsReveal.Revealed = 20

	summary := normalizeForGolden(stripANSI(m.renderStatsPanel(80, 9)))
	assertGoldenFile(t, "stats_tabswitch_summary_w80.golden", summary)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	um.statsReveal.Revealed = 20
	distribution := normalizeForGolden(stripANSI(um.renderStatsPanel(80, 9)))
	assertGoldenFile(t, "stats_tabswitch_distribution_w80.golden", distribution)

	if summary == distribution {
		t.Fatal("expected summary and distribution snapshots to differ after tab switch")
	}
}

func TestLayoutNoOverflowAtCommonTerminalSizes(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{
		{width: 64, height: 14},
		{width: 70, height: 16},
		{width: 80, height: 24},
		{width: 120, height: 40},
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			m := statsFixtureModel()
			m.width = size.width
			m.height = size.height
			m.focusedPanel = panelResults
			m = m.updateFocus()
			m.reveal.Rows = len(m.results)
			m.reveal.Revealing = false
			m.statsReveal.Revealed = 20

			contentWidth := m.width - 4
			leftWidth := contentWidth * 2 / 3
			rightWidth := contentWidth - leftWidth
			if leftWidth < 24 {
				leftWidth = 24
				rightWidth = contentWidth - leftWidth
			}
			if rightWidth < 20 {
				rightWidth = 20
				leftWidth = contentWidth - rightWidth
			}

			searchHeight := 2
			resultsHeight := max(4, m.height-layoutOverhead)
			historyHeight := 2

			leftTotal := (searchHeight + 2) + (resultsHeight + 2)
			const (
				calcMinHeight  = 4
				statsMinHeight = 6
				statsMaxHeight = 9
			)
			statsHeight := min(statsMaxHeight, max(statsMinHeight, leftTotal-(calcMinHeight+4)))
			calcHeight := max(calcMinHeight, leftTotal-(statsHeight+2)-2)

			panels := []struct {
				name string
				out  string
			}{
				{name: "search", out: m.renderSearchPanel(leftWidth, searchHeight)},
				{name: "results", out: m.renderResultsPanel(leftWidth, resultsHeight)},
				{name: "stats", out: m.renderStatsPanel(rightWidth, statsHeight)},
				{name: "calculator", out: m.renderCalculatorPanel(rightWidth, calcHeight)},
				{name: "history", out: m.renderHistoryPanel(m.width-2, historyHeight)},
			}
			for _, panel := range panels {
				t.Run(panel.name, func(t *testing.T) {
					assertNoVisualOverflow(t, panel.out)
				})
			}
		})
	}
}

func statsFixtureModel() Model {
	m := newTestModel()
	m.width = 120
	m.height = 32
	m.results = []types.Listing{
		{Platform: "eBay", Price: 105, Condition: "Used", Status: "Sold"},
		{Platform: "eBay", Price: 250, Condition: "Used", Status: "Active"},
		{Platform: "Mercari", Price: 499, Condition: "Good", Status: "Sold"},
		{Platform: "Amazon", Price: 750, Condition: "New", Status: "Active"},
		{Platform: "Facebook", Price: 199, Condition: "Fair", Status: "Active"},
		{Platform: "Mercari", Price: 420, Condition: "Good", Status: "Sold"},
		{Platform: "eBay", Price: 1050, Condition: "New", Status: "Active"},
	}
	m.extendedStats = idea.CalculateExtendedStats(m.results)
	m.stats = m.extendedStats.Statistics
	return m
}

func assertNoVisualOverflow(t *testing.T, rendered string) {
	t.Helper()
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		t.Fatal("expected rendered output to have at least one line")
	}

	frameWidth := xansi.StringWidth(lines[0])
	for i, line := range lines {
		if got := xansi.StringWidth(line); got > frameWidth {
			t.Fatalf("line %d exceeds frame width (%d > %d): %q", i+1, got, frameWidth, stripANSI(line))
		}
	}
}

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func normalizeForGolden(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

func assertGoldenFile(t *testing.T, fileName, got string) {
	t.Helper()

	path := filepath.Join("testdata", fileName)
	if updateGoldens {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for golden file %q: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden file %q: %v", path, err)
		}
	}

	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file %q: %v (run with UPDATE_GOLDEN=1)", path, err)
	}
	want := normalizeForGolden(string(wantBytes))
	if want != got {
		t.Fatalf("snapshot mismatch for %s\n--- want ---\n%s--- got ---\n%s", fileName, want, got)
	}
}
