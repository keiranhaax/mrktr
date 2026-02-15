package main

import (
	"fmt"
	"math"
	"mrktr/idea"
	"mrktr/types"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

		platformCell := platformStyleFor(r.Platform).Render(fmt.Sprintf("%-*s", colPlatform, platformRaw))
		priceCell := priceStyle.Render(fmt.Sprintf("%*s", colPrice, price))
		conditionCell := fmt.Sprintf("%-*s", colCondition, cond)

		row := fmt.Sprintf("%-*s %*d %s %s  %s ",
			colCursor, cursor,
			colNum, i+1,
			platformCell,
			priceCell,
			conditionCell,
		)

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
	tabs := idea.RenderStatsTabs(m.statsViewMode)

	if len(m.results) == 0 {
		if m.loading {
			skeleton := strings.Join(idea.RenderStatsSkeleton(m.statsAnim.SkeletonFrame, max(12, width-8)), "\n")
			content := tabs + "\n" + skeleton
			return renderPanel("~", "Statistics", content, width, height, active, flashActive)
		}
		content := tabs + "\n" + emptyStyle.Render("── ── ──") + "\n" + mutedStyle.Render("awaiting data")
		return renderPanel("~", "Statistics", content, width, height, active, flashActive)
	}

	s := m.extendedStats
	animated := m.currentAnimatedStats()
	prices := make([]float64, len(m.results))
	for i, result := range m.results {
		prices[i] = result.Price
	}

	sparklineWidth := max(8, width-12)
	sparkline := renderSparkline(prices, sparklineWidth)
	bodyMaxRows := max(1, height-1)

	var lines []string
	switch m.statsViewMode {
	case idea.StatsViewDistribution:
		lines = idea.RenderDistributionBody(s, max(12, width-8), bodyMaxRows)
	case idea.StatsViewMarket:
		lines = idea.RenderMarketBody(animated, max(12, width-8), bodyMaxRows)
	default:
		lines = m.renderStatsSummaryLines(animated, sparkline, max(12, width-8), bodyMaxRows)
	}

	if len(lines) > bodyMaxRows {
		lines = lines[:bodyMaxRows]
	}

	revealCount := len(lines)
	if m.statsReveal.Revealed < revealCount {
		revealCount = max(0, m.statsReveal.Revealed)
	}
	lines = lines[:revealCount]
	if len(lines) == 0 {
		lines = []string{mutedStyle.Render(" ")}
	}

	content := tabs + "\n" + strings.Join(lines, "\n")
	return renderPanel("~", "Statistics", content, width, height, active, flashActive)
}

func (m Model) currentAnimatedStats() idea.ExtendedStatistics {
	target := m.extendedStats
	if !m.statsAnim.ValueTweenOn && m.statsAnim.ValueStep == 0 {
		return target
	}
	if m.statsAnim.ValueSteps <= 0 || m.statsAnim.ToStats.Count == 0 {
		return target
	}

	from := m.statsAnim.FromStats
	to := m.statsAnim.ToStats
	if from.Count == 0 {
		return target
	}

	t := float64(m.statsAnim.ValueStep) / float64(max(1, m.statsAnim.ValueSteps))
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	eased := 1 - math.Pow(1-t, 3)

	out := to
	out.Min = lerp(from.Min, to.Min, eased)
	out.Max = lerp(from.Max, to.Max, eased)
	out.Average = lerp(from.Average, to.Average, eased)
	out.Median = lerp(from.Median, to.Median, eased)
	out.P25 = lerp(from.P25, to.P25, eased)
	out.P75 = lerp(from.P75, to.P75, eased)
	out.SoldAvg = lerp(from.SoldAvg, to.SoldAvg, eased)
	out.ActiveAvg = lerp(from.ActiveAvg, to.ActiveAvg, eased)
	out.SoldCount = lerpInt(from.SoldCount, to.SoldCount, eased)
	out.ActiveCount = lerpInt(from.ActiveCount, to.ActiveCount, eased)
	out.PlatformStats = interpolatePlatformStats(from.PlatformStats, to.PlatformStats, eased)
	out.Statistics = types.Statistics{
		Count:   to.Count,
		Min:     out.Min,
		Max:     out.Max,
		Average: out.Average,
		Median:  out.Median,
	}

	return out
}

func (m Model) renderStatsSummaryLines(stats idea.ExtendedStatistics, sparkline string, width int, maxRows int) []string {
	if width < 20 {
		width = 20
	}
	if maxRows < 1 {
		maxRows = 1
	}

	resultSpread := fmt.Sprintf("Results: %d  Spread: %s", stats.Count, idea.RenderSpreadValue(stats.Spread))
	minValue := fmt.Sprintf("$%.2f", stats.Min) + m.renderStatsDelta(m.statsAnim.DeltaMin)
	maxValue := fmt.Sprintf("$%.2f", stats.Max) + m.renderStatsDelta(m.statsAnim.DeltaMax)
	avgValue := fmt.Sprintf("$%.2f", stats.Average) + m.renderStatsDelta(m.statsAnim.DeltaAvg)
	medianValue := fmt.Sprintf("$%.2f", stats.Median) + m.renderStatsDelta(m.statsAnim.DeltaMedian)
	p25Value := fmt.Sprintf("$%.2f", stats.P25) + m.renderStatsDelta(m.statsAnim.DeltaP25)
	p75Value := fmt.Sprintf("$%.2f", stats.P75) + m.renderStatsDelta(m.statsAnim.DeltaP75)

	if width < 44 {
		lines := []string{
			resultSpread,
			"Trend: " + sparkline,
			fmt.Sprintf("Min: %s  Max: %s", minValue, maxValue),
			fmt.Sprintf("Avg: %s  Med: %s", avgValue, medianValue),
			fmt.Sprintf("P25: %s  P75: %s", p25Value, p75Value),
		}
		if maxRows >= 6 {
			lines = append(lines, fmt.Sprintf("StdDev: $%.2f  CoV: %.2f", stats.StdDev, stats.CoV))
		}
		return lines
	}

	lines := []string{
		resultSpread,
		"Trend: " + sparkline,
		fmt.Sprintf("Min: %s   P25: %s", minValue, p25Value),
		fmt.Sprintf("Max: %s   P75: %s", maxValue, p75Value),
		fmt.Sprintf("Avg: %s  Med: %s", avgValue, medianValue),
	}
	if maxRows >= 6 {
		lines = append(lines, fmt.Sprintf("StdDev: $%.2f  CoV: %.2f", stats.StdDev, stats.CoV))
	}
	return lines
}

func (m Model) renderStatsDelta(delta float64) string {
	if m.statsAnim.DeltaTicks <= 0 || math.Abs(delta) < 0.005 {
		return ""
	}

	ratio := float64(m.statsAnim.DeltaTicks) / float64(max(1, m.statsAnim.DeltaTotal))
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	target := "#12B76A"
	arrow := "↑"
	value := delta
	if delta < 0 {
		target = "#D92D20"
		arrow = "↓"
		value = -delta
	}

	color := interpolateHexColor("#667085", target, ratio)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Render(fmt.Sprintf(" %s$%.2f", arrow, value))
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func lerpInt(a, b int, t float64) int {
	return int(math.Round(lerp(float64(a), float64(b), t)))
}

func interpolatePlatformStats(from, to map[string]idea.PlatformStat, t float64) map[string]idea.PlatformStat {
	if len(from) == 0 && len(to) == 0 {
		return map[string]idea.PlatformStat{}
	}

	out := make(map[string]idea.PlatformStat, len(from)+len(to))
	for name, toStat := range to {
		fromStat := from[name]
		out[name] = idea.PlatformStat{
			Count:   lerpInt(fromStat.Count, toStat.Count, t),
			Average: lerp(fromStat.Average, toStat.Average, t),
			Min:     lerp(fromStat.Min, toStat.Min, t),
			Max:     lerp(fromStat.Max, toStat.Max, t),
		}
	}
	for name, fromStat := range from {
		if _, exists := out[name]; exists {
			continue
		}
		out[name] = idea.PlatformStat{
			Count:   lerpInt(fromStat.Count, 0, t),
			Average: lerp(fromStat.Average, 0, t),
			Min:     lerp(fromStat.Min, 0, t),
			Max:     lerp(fromStat.Max, 0, t),
		}
	}

	for name, stat := range out {
		if stat.Count <= 0 {
			delete(out, name)
		}
	}

	return out
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
