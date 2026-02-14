package main

import (
	"fmt"
	"math"
	"mrktr/api"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderModeBadge(mode api.SearchMode) string {
	dotStyle := warningStyle
	switch mode {
	case api.SearchModeLive:
		dotStyle = successStyle
	case api.SearchModeUnavailable:
		dotStyle = dangerStyle
	}
	return dotStyle.Render("●") + " " + mutedStyle.Render(string(mode))
}

func renderSparkline(prices []float64, width int) string {
	if width <= 0 {
		return ""
	}
	if len(prices) == 0 {
		return strings.Repeat(" ", width)
	}

	blocks := []rune("▁▂▃▄▅▆▇█")
	bins := make([]float64, width)
	for i := 0; i < width; i++ {
		start := i * len(prices) / width
		end := (i + 1) * len(prices) / width
		if end <= start {
			end = min(len(prices), start+1)
		}
		if start >= len(prices) {
			start = len(prices) - 1
		}
		var sum float64
		for j := start; j < end; j++ {
			sum += prices[j]
		}
		bins[i] = sum / float64(max(1, end-start))
	}

	minPrice := bins[0]
	maxPrice := bins[0]
	for _, value := range bins[1:] {
		if value < minPrice {
			minPrice = value
		}
		if value > maxPrice {
			maxPrice = value
		}
	}

	var b strings.Builder
	for _, value := range bins {
		ratio := 0.5
		if maxPrice > minPrice {
			ratio = (value - minPrice) / (maxPrice - minPrice)
		}
		level := int(math.Round(ratio * float64(len(blocks)-1)))
		if level < 0 {
			level = 0
		}
		if level >= len(blocks) {
			level = len(blocks) - 1
		}
		color := interpolateHexColor("#12B76A", "#D92D20", ratio)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(string(blocks[level])))
	}

	return b.String()
}

func renderProfitBar(profit, maxProfit float64, width int) string {
	if width <= 0 {
		return ""
	}
	if maxProfit <= 0 {
		return strings.Repeat("░", width)
	}

	ratio := math.Abs(profit) / maxProfit
	if ratio > 1 {
		ratio = 1
	}
	filled := int(math.Round(ratio * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	if profit >= 0 {
		return successStyle.Render(bar)
	}
	return dangerStyle.Render(bar)
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
		return "\u2026"
	}
	return string(runes[:max-1]) + "\u2026"
}

func formatProfit(profit float64) string {
	if profit >= 0 {
		return successStyle.Render(fmt.Sprintf("+$%.2f", profit))
	}
	return dangerStyle.Render(fmt.Sprintf("-$%.2f", -profit))
}

func formatPercent(pct float64) string {
	if pct >= 0 {
		return successStyle.Render(fmt.Sprintf("+%.0f%%", pct))
	}
	return dangerStyle.Render(fmt.Sprintf("%.0f%%", pct))
}

func maxAbs(values ...float64) float64 {
	maxValue := 0.0
	for _, value := range values {
		if abs := math.Abs(value); abs > maxValue {
			maxValue = abs
		}
	}
	return maxValue
}
