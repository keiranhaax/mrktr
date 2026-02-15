package idea

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type platformRow struct {
	Name string
	Stat PlatformStat
}

// RenderMarketBody renders platform averages plus sold/active summary.
func RenderMarketBody(stats ExtendedStatistics, width int, maxRows int) []string {
	if width < 24 {
		width = 24
	}
	if maxRows < 3 {
		maxRows = 3
	}
	if len(stats.PlatformStats) == 0 {
		return []string{
			"Market Breakdown",
			"~ insufficient data ~",
		}
	}

	rows := make([]platformRow, 0, len(stats.PlatformStats))
	for name, stat := range stats.PlatformStats {
		rows = append(rows, platformRow{Name: name, Stat: stat})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Stat.Count != rows[j].Stat.Count {
			return rows[i].Stat.Count > rows[j].Stat.Count
		}
		return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name)
	})

	maxPlatformRows := maxRows - 2 // sold/active and sell-through lines
	if maxPlatformRows < 1 {
		maxPlatformRows = 1
	}
	if len(rows) > maxPlatformRows {
		rows = rows[:maxPlatformRows]
	}

	maxAvg := 0.0
	for _, row := range rows {
		if row.Stat.Average > maxAvg {
			maxAvg = row.Stat.Average
		}
	}
	if maxAvg <= 0 {
		maxAvg = 1
	}

	lines := make([]string, 0, len(rows)+2)

	for _, row := range rows {
		labelWidth := minInt(8, maxInt(4, width/5))
		label := truncate(row.Name, labelWidth)

		suffix := fmt.Sprintf("%s (%d)", formatPrice(row.Stat.Average), row.Stat.Count)
		if width < 34 {
			suffix = fmt.Sprintf("%s %d", formatPrice(row.Stat.Average), row.Stat.Count)
		}
		if width < 28 {
			suffix = fmt.Sprintf("%d", row.Stat.Count)
		}

		barWidth := width - labelWidth - len([]rune(suffix)) - 2
		barWidth = minInt(16, barWidth)
		if barWidth < 2 {
			labelWidth = maxInt(3, width-len([]rune(suffix))-4)
			label = truncate(row.Name, labelWidth)
			barWidth = width - labelWidth - len([]rune(suffix)) - 2
		}
		barWidth = maxInt(2, barWidth)

		ratio := row.Stat.Average / maxAvg
		bar := renderPlatformBar(row.Name, ratio, barWidth)
		line := fmt.Sprintf("%-*s %s %s", labelWidth, label, bar, suffix)
		lines = append(lines, clipANSIWidth(line, width))
	}

	var statusLine string
	switch {
	case width >= 46:
		statusLine = fmt.Sprintf("Sold:%d avg %s  Active:%d avg %s",
			stats.SoldCount,
			formatPrice(stats.SoldAvg),
			stats.ActiveCount,
			formatPrice(stats.ActiveAvg),
		)
	case width >= 30:
		statusLine = fmt.Sprintf("S:%d %s  A:%d %s",
			stats.SoldCount,
			formatPrice(stats.SoldAvg),
			stats.ActiveCount,
			formatPrice(stats.ActiveAvg),
		)
	default:
		statusLine = fmt.Sprintf("S:%d  A:%d", stats.SoldCount, stats.ActiveCount)
	}
	lines = append(lines, clipANSIWidth(statusLine, width))

	total := stats.SoldCount + stats.ActiveCount
	sellThrough := 0.0
	if total > 0 {
		sellThrough = float64(stats.SoldCount) / float64(total)
	}
	sellBar := renderRateBar(sellThrough, maxInt(3, minInt(18, width-20)))
	sellPct := int(math.Round(sellThrough * 100))
	sellPrefix := "Sell-through"
	if width < 36 {
		sellPrefix = "ST"
	}
	lines = append(lines, clipANSIWidth(fmt.Sprintf("%s: %d%%  %s", sellPrefix, sellPct, sellBar), width))

	return lines
}

func renderPlatformBar(platform string, ratio float64, width int) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(math.Round(ratio * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	empty := width - filled
	style := platformBarStyle(platform)
	return style.Render(strings.Repeat("█", filled)) + strings.Repeat("░", empty)
}

func platformBarStyle(name string) lipgloss.Style {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "ebay":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#E53238"))
	case "mercari":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#4DC9F6"))
	case "amazon":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9900"))
	case "facebook", "facebook marketplace":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#1877F2"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#98A2B3"))
	}
}

func renderRateBar(ratio float64, width int) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(math.Round(ratio * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
