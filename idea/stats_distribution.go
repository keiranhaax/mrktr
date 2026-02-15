package idea

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	xansi "github.com/charmbracelet/x/ansi"
)

// RenderDistributionBody renders histogram rows and percentile band.
func RenderDistributionBody(stats ExtendedStatistics, width int, maxRows int) []string {
	if width < 24 {
		width = 24
	}
	if maxRows < 2 {
		maxRows = 2
	}

	if len(stats.Histogram) == 0 {
		return []string{
			"Price Distribution",
			"~ insufficient data ~",
		}
	}

	bins := stats.Histogram
	maxBins := maxRows - 1 // reserve one line for percentile labels
	if maxBins < 1 {
		maxBins = 1
	}
	if len(bins) > maxBins {
		bins = bins[:maxBins]
	}

	maxCount := 0
	for _, b := range bins {
		if b.Count > maxCount {
			maxCount = b.Count
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}

	barWidth := maxInt(6, minInt(22, width-18))
	lines := make([]string, 0, len(bins)+1)

	for i, bin := range bins {
		ratio := float64(bin.Count) / float64(maxCount)
		bar := renderHistogramBar(ratio, barWidth, i, len(bins))
		label := truncate(bin.Label, 11)
		marker := " "
		if bin.MinPrice <= stats.P25 {
			marker = "★"
		}
		line := fmt.Sprintf("%-11s %s %2d %s", label, bar, bin.Count, marker)
		lines = append(lines, clipANSIWidth(line, width))
	}

	var percentileLine string
	switch {
	case width >= 56:
		percentileLine = fmt.Sprintf("P10:%s P25:%s Med:%s P75:%s P90:%s",
			formatPrice(stats.P10),
			formatPrice(stats.P25),
			formatPrice(stats.Median),
			formatPrice(stats.P75),
			formatPrice(stats.P90),
		)
	case width >= 40:
		percentileLine = fmt.Sprintf("P25:%s Med:%s P75:%s",
			formatPrice(stats.P25),
			formatPrice(stats.Median),
			formatPrice(stats.P75),
		)
	default:
		percentileLine = fmt.Sprintf("Med:%s P25:%s P75:%s",
			formatPrice(stats.Median),
			formatPrice(stats.P25),
			formatPrice(stats.P75),
		)
	}
	lines = append(lines, clipANSIWidth(percentileLine, width))

	return lines
}

func renderHistogramBar(ratio float64, width, idx, total int) string {
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
	gradient := 0.0
	if total > 1 {
		gradient = float64(idx) / float64(total-1)
	}
	color := blendHex("#12B76A", "#D92D20", gradient)
	filledBar := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(strings.Repeat("█", filled))
	return filledBar + strings.Repeat("░", empty)
}

func blendHex(a, b string, t float64) string {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	parse := func(hex string) (int, int, int) {
		if len(hex) != 7 || hex[0] != '#' {
			return 255, 255, 255
		}
		var r, g, bl int
		_, _ = fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &bl)
		return r, g, bl
	}

	ar, ag, ab := parse(a)
	br, bg, bb := parse(b)
	lerp := func(x, y int) int {
		return int(math.Round(float64(x) + (float64(y)-float64(x))*t))
	}

	return fmt.Sprintf("#%02X%02X%02X", lerp(ar, br), lerp(ag, bg), lerp(ab, bb))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

func clipANSIWidth(line string, width int) string {
	if width <= 0 {
		return ""
	}
	if xansi.StringWidth(line) <= width {
		return line
	}
	if width <= 1 {
		return "…"
	}
	return xansi.Truncate(line, width, "…")
}
