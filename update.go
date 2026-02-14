package main

import (
	"fmt"
	"mrktr/types"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model (required by tea.Model interface)
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		visible := m.visibleResultRows()
		if m.resultsOffset > 0 && m.resultsOffset+visible > len(m.results) {
			m.resultsOffset = max(0, len(m.results)-visible)
		}
		return m, nil

	case SearchResultsMsg:
		m.loading = false
		m.dataMode = msg.Mode
		m.warning = msg.Warning
		if msg.Err != nil {
			m.err = msg.Err
			m.revealing = false
			m.revealedRows = 0
			m.statsRevealed = 0
			return m, nil
		}
		m.results = msg.Results
		m.stats = types.CalculateStats(m.results)
		m.selectedIndex = 0
		m.resultsOffset = 0
		m.err = nil

		var cmds []tea.Cmd
		if len(m.results) > 0 {
			m.revealGen++
			m.revealedRows = 0
			m.revealing = true
			revealGen := m.revealGen
			cmds = append(cmds, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
				return revealRowTickMsg{gen: revealGen}
			}))

			m.statsRevealGen++
			m.statsRevealed = 0
			statsGen := m.statsRevealGen
			cmds = append(cmds, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
				return statsRevealTickMsg{gen: statsGen}
			}))
		} else {
			m.revealing = false
			m.revealedRows = 0
			m.statsRevealed = 0
		}

		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case openURLResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, nil

	case focusFlashTickMsg:
		if msg.gen != m.focusFlashGen || !m.focusFlashActive {
			return m, nil
		}
		if m.focusFlashTicks <= 1 {
			m.focusFlashTicks = 0
			m.focusFlashActive = false
			return m, nil
		}
		m.focusFlashTicks--
		gen := m.focusFlashGen
		return m, tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
			return focusFlashTickMsg{gen: gen}
		})

	case revealRowTickMsg:
		if msg.gen != m.revealGen || !m.revealing {
			return m, nil
		}
		targetRows := min(len(m.results), m.visibleResultRows())
		if m.revealedRows < targetRows {
			m.revealedRows++
		}
		if m.revealedRows >= targetRows {
			m.revealing = false
			return m, nil
		}
		gen := m.revealGen
		return m, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
			return revealRowTickMsg{gen: gen}
		})

	case statsRevealTickMsg:
		if msg.gen != m.statsRevealGen {
			return m, nil
		}
		if len(m.results) == 0 {
			return m, nil
		}
		const totalStatsLines = 6
		if m.statsRevealed >= totalStatsLines {
			return m, nil
		}
		m.statsRevealed++
		if m.statsRevealed >= totalStatsLines {
			return m, nil
		}
		gen := m.statsRevealGen
		return m, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
			return statsRevealTickMsg{gen: gen}
		})

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			m.loadingDots = (m.loadingDots + 1) % 4
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.focusedPanel != panelSearch && m.focusedPanel != panelCalculator {
				return m, tea.Quit
			}

		case "tab":
			nextPanel := (m.focusedPanel + 1) % 5
			return m.changeFocus(nextPanel)

		case "shift+tab":
			prevPanel := m.focusedPanel - 1
			if prevPanel < 0 {
				prevPanel = panelHistory
			}
			return m.changeFocus(prevPanel)

		case "/":
			return m.changeFocus(panelSearch)

		case "c":
			// Focus calculator only when not in text inputs.
			if m.focusedPanel != panelSearch && m.focusedPanel != panelCalculator {
				return m.changeFocus(panelCalculator)
			}

		case "esc":
			return m.changeFocus(panelResults)
		}

		// Panel-specific keys
		switch m.focusedPanel {
		case panelSearch:
			switch msg.String() {
			case "enter":
				// Execute search
				query := strings.TrimSpace(m.searchInput.Value())
				if query != "" {
					m.loading = true
					m.loadingDots = 0
					m.warning = ""
					m.err = nil
					m.addToHistory(query)
					return m, m.doSearch(query)
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd

		case panelResults:
			switch msg.String() {
			case "j", "down":
				if m.revealing {
					m.revealing = false
					m.revealedRows = len(m.results)
				}
				if m.selectedIndex < len(m.results)-1 {
					m.selectedIndex++
					visible := m.visibleResultRows()
					if m.selectedIndex >= m.resultsOffset+visible {
						m.resultsOffset = m.selectedIndex - visible + 1
					}
				}
			case "k", "up":
				if m.revealing {
					m.revealing = false
					m.revealedRows = len(m.results)
				}
				if m.selectedIndex > 0 {
					m.selectedIndex--
					if m.selectedIndex < m.resultsOffset {
						m.resultsOffset = m.selectedIndex
					}
				}
			case "enter":
				// Open URL in browser
				if len(m.results) > 0 && m.selectedIndex < len(m.results) {
					url := m.results[m.selectedIndex].URL
					if url != "" {
						return m, openURLCmd(url)
					}
				}
			}
			return m, nil

		case panelCalculator:
			switch msg.String() {
			case "enter":
				// Parse cost and unfocus
				if val, err := strconv.ParseFloat(m.costInput.Value(), 64); err == nil {
					m.cost = val
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.costInput, cmd = m.costInput.Update(msg)
			// Auto-update cost as typing
			if val, err := strconv.ParseFloat(m.costInput.Value(), 64); err == nil {
				m.cost = val
			}
			return m, cmd

		case panelHistory:
			switch msg.String() {
			case "j", "right":
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
				}
			case "k", "left":
				if m.historyIndex > 0 {
					m.historyIndex--
				}
			case "enter":
				// Re-run selected history search
				if len(m.history) > 0 && m.historyIndex < len(m.history) {
					query := m.history[m.historyIndex]
					m.searchInput.SetValue(query)
					m.loading = true
					m.loadingDots = 0
					m.warning = ""
					m.err = nil
					return m, m.doSearch(query)
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// updateFocus manages focus state for text inputs
func (m Model) updateFocus() Model {
	if m.focusedPanel == panelSearch {
		m.searchInput.Focus()
		m.costInput.Blur()
	} else if m.focusedPanel == panelCalculator {
		m.costInput.Focus()
		m.searchInput.Blur()
	} else {
		m.searchInput.Blur()
		m.costInput.Blur()
	}
	return m
}

func (m Model) changeFocus(newPanel int) (tea.Model, tea.Cmd) {
	if m.focusedPanel == newPanel {
		m = m.updateFocus()
		return m, nil
	}

	m.focusedPanel = newPanel
	m = m.updateFocus()
	m.focusFlashGen++
	m.focusFlashTicks = 3
	m.focusFlashActive = true
	gen := m.focusFlashGen

	return m, tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
		return focusFlashTickMsg{gen: gen}
	})
}

// addToHistory adds a search query to history (avoiding duplicates)
func (m *Model) addToHistory(query string) {
	// Remove if already exists
	for i, h := range m.history {
		if h == query {
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}
	// Add to front
	m.history = append([]string{query}, m.history...)
	// Limit to 20 items
	if len(m.history) > 20 {
		m.history = m.history[:20]
	}
}

// doSearch creates a command to fetch search results
func (m Model) doSearch(query string) tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			response := SearchPrices(strings.TrimSpace(query))
			return SearchResultsMsg{
				Results: response.Results,
				Mode:    response.Mode,
				Warning: response.Warning,
				Err:     response.Err,
			}
		},
	)
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return openURLResultMsg{Err: openURL(url)}
	}
}

// openURL opens a URL in the default browser
func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("open URL: %w", err)
		}
		return nil
	}
	return fmt.Errorf("open URL: unsupported platform %q", runtime.GOOS)
}
