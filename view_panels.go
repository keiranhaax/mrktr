package main

import (
	"fmt"
	"mrktr/types"
	"strings"
)

func (m Model) renderSearchPanel(width, height int) string {
	active := m.focusedPanel == panelSearch
	flashActive := active && m.focusFlash.Active

	content := m.searchInput.View()
	if m.loading {
		content += " " + m.spinner.View() + " Searching" + strings.Repeat(".", m.loadingDots)
	}
	return renderPanel("/", "Search", content, width, height, active, flashActive)
}

func (m Model) renderResultsPanel(width, height int) string {
	active := m.focusedPanel == panelResults
	flashActive := active && m.focusFlash.Active

	if len(m.results) == 0 {
		content := emptyStyle.Render("~ No results yet ~") + "\n" + keyStyle.Render("/") + keyDescStyle.Render(" search")
		return renderPanel("#", "Results", content, width, height, active, flashActive)
	}

	visibleRows := m.visibleResultRows()
	if m.reveal.Revealing {
		visibleRows = min(visibleRows, m.reveal.Rows)
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

		row := fmt.Sprintf("%-*s %*d %-*s %*s  %-*s ",
			colCursor, cursor,
			colNum, i+1,
			colPlatform, platformRaw,
			colPrice, price,
			colCondition, cond,
		)

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
	flashActive := active && m.focusFlash.Active

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
	if m.statsReveal.Revealed < revealCount {
		revealCount = max(0, m.statsReveal.Revealed)
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
	flashActive := active && m.focusFlash.Active

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
	flashActive := active && m.focusFlash.Active

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
	helpModel := m.help
	helpModel.Width = max(0, m.width-2)
	help := helpModel.View(m.keys)

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
