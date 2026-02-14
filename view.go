package main

import (
	"fmt"
	"math"
	"mrktr/types"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI (required by tea.Model interface)
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	if m.width < 64 || m.height < 14 {
		return helpStyle.Render(
			fmt.Sprintf(
				"Terminal too small (%dx%d). Resize to at least 64x14.",
				m.width,
				m.height,
			),
		)
	}

	// Calculate panel dimensions.
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
	statsHeight := 6
	calcHeight := 6
	historyHeight := 2

	// Render panels.
	appHeader := m.renderAppHeader(contentWidth)
	searchPanel := m.renderSearchPanel(leftWidth, searchHeight)
	resultsPanel := m.renderResultsPanel(leftWidth, resultsHeight)
	statsPanel := m.renderStatsPanel(rightWidth, statsHeight)
	calcPanel := m.renderCalculatorPanel(rightWidth, calcHeight)
	historyPanel := m.renderHistoryPanel(m.width-4, historyHeight)
	helpBar := m.renderHelpBar()

	// Compose left column.
	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		searchPanel,
		resultsPanel,
	)

	// Compose right column.
	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		statsPanel,
		calcPanel,
	)

	// Join columns horizontally.
	mainArea := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		rightColumn,
	)

	// Full layout.
	return lipgloss.JoinVertical(
		lipgloss.Left,
		appHeader,
		mainArea,
		historyPanel,
		helpBar,
	)
}

func (m Model) renderAppHeader(contentWidth int) string {
	title := renderGradientText("m r k t r", "#7D56F4", "#EA80FC")
	subtitle := mutedStyle.Render("Reseller Price Research")
	separator := renderGradientText(strings.Repeat("━", max(8, contentWidth)), "#7D56F4", "#EA80FC")
	return lipgloss.JoinVertical(lipgloss.Left, title+" "+subtitle, separator)
}

func (m Model) renderSearchPanel(width, height int) string {
	active := m.focusedPanel == panelSearch
	flashActive := active && m.focusFlashActive

	content := m.searchInput.View()
	if m.loading {
		content += " " + m.spinner.View() + " Searching" + strings.Repeat(".", m.loadingDots)
	}
	return renderPanel("/", "Search", content, width, height, active, flashActive)
}

func (m Model) renderResultsPanel(width, height int) string {
	active := m.focusedPanel == panelResults
	flashActive := active && m.focusFlashActive

	if len(m.results) == 0 {
		content := emptyStyle.Render("~ No results yet ~") + "\n" + keyStyle.Render("/") + keyDescStyle.Render(" search")
		return renderPanel("#", "Results", content, width, height, active, flashActive)
	}

	visibleRows := m.visibleResultRows()
	if m.revealing {
		visibleRows = min(visibleRows, m.revealedRows)
		if visibleRows < 0 {
			visibleRows = 0
		}
	}

	start := m.resultsOffset
	if start < 0 {
		start = 0
	}
	maxStart := max(0, len(m.results)-max(1, visibleRows))
	if start > maxStart {
		start = maxStart
	}
	end := min(len(m.results), start+visibleRows)

	// Column widths for results table.
	const (
		colCursor    = 2
		colNum       = 3
		colPlatform  = 11
		colPrice     = 10
		colCondition = 10
		colStatus    = 8
	)

	var lines []string
	header := fmt.Sprintf("%-*s %-*s %-*s %*s  %-*s %-*s",
		colCursor, "",
		colNum, "#",
		colPlatform, "Platform",
		colPrice, "Price",
		colCondition, "Condition",
		colStatus, "Status",
	)
	lines = append(lines, headerStyle.Render(header))

	for i := start; i < end; i++ {
		r := m.results[i]
		price := fmt.Sprintf("$%.2f", r.Price)
		cond := truncate(r.Condition, colCondition)
		status := r.Status
		platformRaw := truncate(r.Platform, colPlatform)

		cursor := ""
		if i == m.selectedIndex {
			cursor = "▸"
		}

		// Build row with plain-text padding first, then colorize individual fields.
		row := fmt.Sprintf("%-*s %*d %-*s %*s  %-*s ",
			colCursor, cursor,
			colNum, i+1,
			colPlatform, platformRaw,
			colPrice, price,
			colCondition, cond,
		)

		// Replace plain fields with styled versions in-place.
		row = strings.Replace(row, platformRaw, platformStyleFor(r.Platform).Render(platformRaw), 1)
		row = strings.Replace(row, price, priceStyle.Render(price), 1)

		var statusStyled string
		if status == "Sold" {
			statusStyled = soldStyle.Render(fmt.Sprintf("%-*s", colStatus, status))
		} else {
			statusStyled = activeStyle.Render(fmt.Sprintf("%-*s", colStatus, status))
		}
		row += statusStyled

		if i == m.selectedIndex && active {
			row = selectedStyle.Render(row)
		} else if i%2 == 1 {
			row = rowAltStyle.Render(row)
		} else {
			row = rowStyle.Render(row)
		}

		lines = append(lines, row)
	}

	if visibleRows == 0 {
		lines = append(lines, scrollInfoStyle.Render("revealing..."))
	} else if len(m.results) > m.visibleResultRows() {
		lines = append(lines, scrollInfoStyle.Render(fmt.Sprintf("showing %d-%d of %d", start+1, end, len(m.results))))
	} else {
		lines = append(lines, scrollInfoStyle.Render(fmt.Sprintf("showing 1-%d of %d", end, len(m.results))))
	}

	content := strings.Join(lines, "\n")
	return renderPanel("#", "Results", content, width, height, active, flashActive)
}

func (m Model) renderStatsPanel(width, height int) string {
	active := m.focusedPanel == panelStats
	flashActive := active && m.focusFlashActive

	if len(m.results) == 0 {
		content := emptyStyle.Render("── ── ──") + "\n" + mutedStyle.Render("awaiting data")
		return renderPanel("~", "Statistics", content, width, height, active, flashActive)
	}

	s := m.stats
	prices := make([]float64, len(m.results))
	for i, result := range m.results {
		prices[i] = result.Price
	}

	sparklineWidth := max(8, width-12)
	sparkline := renderSparkline(prices, sparklineWidth)

	lines := []string{
		labelStyle.Render("Results:") + " " + valueStyle.Render(fmt.Sprintf("%d listings", s.Count)),
		labelStyle.Render("Trend:") + " " + sparkline,
		separatorStyle.Render(strings.Repeat("╌", max(12, width-8))),
		labelStyle.Render("Min:") + " " + priceStyle.Render(fmt.Sprintf("$%.2f", s.Min)),
		labelStyle.Render("Max:") + " " + priceStyle.Render(fmt.Sprintf("$%.2f", s.Max)),
		labelStyle.Render("Avg:") + " " + valueStyle.Render(fmt.Sprintf("$%.2f", s.Average)) +
			"  " + labelStyle.Render("Median:") + " " + valueStyle.Render(fmt.Sprintf("$%.2f", s.Median)),
	}

	revealCount := len(lines)
	if m.statsRevealed < revealCount {
		revealCount = max(0, m.statsRevealed)
	}
	lines = lines[:revealCount]
	if len(lines) == 0 {
		lines = []string{mutedStyle.Render(" ")}
	}

	content := strings.Join(lines, "\n")
	return renderPanel("~", "Statistics", content, width, height, active, flashActive)
}

func (m Model) renderCalculatorPanel(width, height int) string {
	active := m.focusedPanel == panelCalculator
	flashActive := active && m.focusFlashActive

	lines := []string{
		labelStyle.Render("Your Cost:") + " $" + m.costInput.View(),
	}

	if m.cost > 0 && len(m.results) > 0 {
		lines = append(lines, separatorStyle.Render(strings.Repeat("╌", max(12, width-8))))

		avgProfit := types.CalculateProfit(m.cost, m.stats.Average)
		minProfit := types.CalculateProfit(m.cost, m.stats.Min)
		maxProfit := types.CalculateProfit(m.cost, m.stats.Max)
		maxProfitMagnitude := maxAbs(avgProfit.Profit, minProfit.Profit, maxProfit.Profit)
		barWidth := max(8, min(18, width/3))

		lines = append(lines, fmt.Sprintf("%s %s (%s) %s",
			labelStyle.Render("At Avg:"),
			formatProfit(avgProfit.Profit),
			formatPercent(avgProfit.ProfitPercent),
			renderProfitBar(avgProfit.Profit, maxProfitMagnitude, barWidth),
		))
		lines = append(lines, fmt.Sprintf("%s %s (%s) %s",
			labelStyle.Render("At Min:"),
			formatProfit(minProfit.Profit),
			formatPercent(minProfit.ProfitPercent),
			renderProfitBar(minProfit.Profit, maxProfitMagnitude, barWidth),
		))
		lines = append(lines, fmt.Sprintf("%s %s (%s) %s",
			labelStyle.Render("At Max:"),
			formatProfit(maxProfit.Profit),
			formatPercent(maxProfit.ProfitPercent),
			renderProfitBar(maxProfit.Profit, maxProfitMagnitude, barWidth),
		))
	} else {
		lines = append(lines, emptyStyle.Render("~ Enter cost to see profits ~"))
	}

	content := strings.Join(lines, "\n")
	return renderPanel("$", "Profit Calculator", content, width, height, active, flashActive)
}

func (m Model) renderHistoryPanel(width, height int) string {
	active := m.focusedPanel == panelHistory
	flashActive := active && m.focusFlashActive

	if len(m.history) == 0 {
		content := mutedStyle.Render("No recent searches")
		return renderPanel(">", "History", content, width, height, active, flashActive)
	}

	const maxItems = 5
	start := 0
	if len(m.history) > maxItems && m.historyIndex >= maxItems {
		start = m.historyIndex - maxItems + 1
	}
	if start+maxItems > len(m.history) {
		start = len(m.history) - maxItems
	}
	if start < 0 {
		start = 0
	}

	end := min(len(m.history), start+maxItems)
	items := m.history[start:end]
	rendered := make([]string, len(items))
	for i, item := range items {
		label := truncate(item, 18)
		selected := start+i == m.historyIndex
		if selected {
			marker := "> " + label
			if active {
				rendered[i] = historySelectedStyle.Render(marker)
			} else {
				rendered[i] = activeTitleStyle.Render(marker)
			}
			continue
		}
		rendered[i] = historyItemStyle.Render(label)
	}

	content := labelStyle.Render("Recent:") + " " + strings.Join(rendered, separatorStyle.Render(" › "))
	return renderPanel(">", "History", content, width, height, active, flashActive)
}

func (m Model) renderHelpBar() string {
	searchGroup := strings.Join([]string{
		keyStyle.Render("/") + keyDescStyle.Render(" search"),
		keyStyle.Render("Enter") + keyDescStyle.Render(" run/open"),
	}, "  ")
	navGroup := strings.Join([]string{
		keyStyle.Render("j/k") + keyDescStyle.Render(" navigate"),
		keyStyle.Render("Tab") + keyDescStyle.Render(" panels"),
		keyStyle.Render("Esc") + keyDescStyle.Render(" results"),
	}, "  ")
	systemGroup := strings.Join([]string{
		keyStyle.Render("c") + keyDescStyle.Render(" cost"),
		keyStyle.Render("q") + keyDescStyle.Render(" quit"),
	}, "  ")

	help := strings.Join(
		[]string{searchGroup, navGroup, systemGroup},
		separatorStyle.Render(" │ "),
	)

	if m.dataMode != "" {
		help = renderModeBadge(m.dataMode) + "  " + help
	}
	if m.warning != "" {
		help = warningStyle.Render(m.warning) + "  " + help
	}
	if m.err != nil {
		errLine := dangerStyle.Render(fmt.Sprintf("Error: %v", m.err))
		return helpStyle.Render(errLine + "\n" + help)
	}
	return helpStyle.Render(help)
}

func renderModeBadge(mode searchMode) string {
	dotStyle := warningStyle
	switch mode {
	case searchModeLive:
		dotStyle = successStyle
	case searchModeFallback:
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

// Helper functions

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
