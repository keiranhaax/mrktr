package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI (required by tea.Model interface).
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.intro.Show {
		return m.renderIntro()
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

	contentWidth := m.width - 4
	stacked := m.width < 80
	leftWidth := contentWidth * 2 / 3
	rightWidth := contentWidth - leftWidth

	if stacked {
		leftWidth = contentWidth
		rightWidth = contentWidth
	} else {
		if leftWidth < 24 {
			leftWidth = 24
			rightWidth = contentWidth - leftWidth
		}
		if rightWidth < 20 {
			rightWidth = 20
			leftWidth = contentWidth - rightWidth
		}
	}

	searchHeight := 2
	historyHeight := 2
	resultsHeight := max(4, m.height-layoutOverhead)

	const (
		calcMinHeight  = 4
		statsMinHeight = 6
		statsMaxHeight = 9
	)
	statsHeight := statsMinHeight
	calcHeight := calcMinHeight

	if !stacked {
		leftTotal := (searchHeight + 2) + (resultsHeight + 2)
		statsHeight = min(statsMaxHeight, max(statsMinHeight, leftTotal-(calcMinHeight+4)))
		calcHeight = max(calcMinHeight, leftTotal-(statsHeight+2)-2)
	} else {
		statsHeight = min(statsMaxHeight, max(statsMinHeight, m.height/4))
		calcHeight = calcMinHeight
		resultsHeight = max(
			4,
			m.height-((searchHeight+2)+(statsHeight+2)+(calcHeight+2)+(historyHeight+2)+5),
		)
	}

	appHeader := m.renderAppHeader(m.width - 2)
	searchPanel := m.renderSearchPanel(leftWidth, searchHeight)
	resultsPanel := m.renderResultsPanel(leftWidth, resultsHeight)
	statsPanel := m.renderStatsPanel(rightWidth, statsHeight)
	calcPanel := m.renderCalculatorPanel(rightWidth, calcHeight)
	historyPanel := m.renderHistoryPanel(m.width-2, historyHeight)
	helpBar := m.renderHelpBar()

	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		searchPanel,
		resultsPanel,
	)

	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		statsPanel,
		calcPanel,
	)

	mainArea := ""
	if stacked {
		mainArea = lipgloss.JoinVertical(
			lipgloss.Left,
			leftColumn,
			rightColumn,
		)
	} else {
		mainArea = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftColumn,
			rightColumn,
		)
	}

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
