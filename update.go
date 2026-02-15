package main

import (
	"context"
	"errors"
	"fmt"
	"mrktr/api"
	"mrktr/idea"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Total ticks for each intro phase.
const (
	introRevealTicks = 30 // character-by-character reveal
	introGlowTicks   = 15 // glow sweep across text
	introFadeTicks   = 8  // fade out
	introTotalTicks  = introRevealTicks + introGlowTicks + introFadeTicks
)

// Update handles messages and updates the model (required by tea.Model interface).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case introTickMsg:
		if !m.intro.Show {
			return m, nil
		}
		m.intro.Tick++

		if m.intro.Tick < introRevealTicks {
			m.intro.Phase = 0
		} else if m.intro.Tick < introRevealTicks+introGlowTicks {
			m.intro.Phase = 1
		} else if m.intro.Tick < introTotalTicks {
			m.intro.Phase = 2
		} else {
			m.intro.Show = false
			m.intro.Completed = true
			return m, nil
		}

		return m, tea.Tick(40*time.Millisecond, func(time.Time) tea.Msg {
			return introTickMsg{}
		})

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		visible := m.visibleResultRows()
		if m.resultsOffset > 0 && m.resultsOffset+visible > len(m.results) {
			m.resultsOffset = max(0, len(m.results)-visible)
		}
		return m, nil

	case SearchResultsMsg:
		return m.handleSearchResults(msg)

	case openURLResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
		}
		return m, nil

	case focusFlashTickMsg:
		if msg.gen != m.focusFlash.Gen || !m.focusFlash.Active {
			return m, nil
		}
		if m.focusFlash.Ticks <= 1 {
			m.focusFlash.Ticks = 0
			m.focusFlash.Active = false
			return m, nil
		}
		m.focusFlash.Ticks--
		gen := m.focusFlash.Gen
		return m, tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
			return focusFlashTickMsg{gen: gen}
		})

	case revealRowTickMsg:
		if msg.gen != m.reveal.Gen || !m.reveal.Revealing {
			return m, nil
		}
		targetRows := min(len(m.results), m.visibleResultRows())
		if m.reveal.Rows < targetRows {
			m.reveal.Rows++
		}
		if m.reveal.Rows >= targetRows {
			m.reveal.Revealing = false
			return m, nil
		}
		gen := m.reveal.Gen
		return m, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
			return revealRowTickMsg{gen: gen}
		})

	case statsRevealTickMsg:
		if msg.gen != m.statsReveal.Gen {
			return m, nil
		}
		if len(m.results) == 0 {
			return m, nil
		}
		targetLines := m.statsRevealTargetLines()
		if m.statsReveal.Revealed >= targetLines {
			return m, nil
		}
		m.statsReveal.Revealed++
		if m.statsReveal.Revealed >= targetLines {
			return m, nil
		}
		gen := m.statsReveal.Gen
		return m, tea.Tick(m.statsRevealTickDuration(), func(time.Time) tea.Msg {
			return statsRevealTickMsg{gen: gen}
		})

	case statsValueTickMsg:
		if msg.gen != m.statsAnim.ValueTweenGen {
			return m, nil
		}

		if m.statsAnim.ValueTweenOn {
			m.statsAnim.ValueStep++
			if m.statsAnim.ValueStep >= m.statsAnim.ValueSteps {
				m.statsAnim.ValueStep = m.statsAnim.ValueSteps
				m.statsAnim.ValueTweenOn = false
			}
		}

		if m.statsAnim.DeltaTicks > 0 {
			m.statsAnim.DeltaTicks--
		}

		if m.statsAnim.ValueTweenOn || m.statsAnim.DeltaTicks > 0 {
			gen := m.statsAnim.ValueTweenGen
			return m, tea.Tick(33*time.Millisecond, func(time.Time) tea.Msg {
				return statsValueTickMsg{gen: gen}
			})
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			m.loadingDots = (m.loadingDots + 1) % 4
			if !m.reduceMotion {
				m.statsAnim.SkeletonFrame++
			}
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		updated, cmd := m.handleKeyMsg(msg)
		return updated, cmd
	}

	return m, nil
}

func (m Model) handleSearchResults(msg SearchResultsMsg) (tea.Model, tea.Cmd) {
	if msg.gen != 0 && msg.gen != m.searchGen {
		return m, nil
	}

	m.cancelActiveSearch()
	m.loading = false
	m.dataMode = msg.Mode
	m.warning = msg.Warning
	if msg.Err != nil {
		if errors.Is(msg.Err, context.Canceled) {
			return m, nil
		}
		m.err = msg.Err
		m.reveal.Revealing = false
		m.reveal.Rows = 0
		m.statsReveal.Revealed = 0
		m.statsAnim.ValueTweenOn = false
		m.statsAnim.ValueStep = 0
		m.statsAnim.ValueSteps = 0
		m.statsAnim.DeltaTicks = 0
		return m, nil
	}

	prevStats := m.extendedStats
	m.results = msg.Results
	m.extendedStats = idea.CalculateExtendedStats(m.results)
	m.stats = m.extendedStats.Statistics
	m.selectedIndex = 0
	m.resultsOffset = 0
	m.err = nil

	var cmds []tea.Cmd
	if len(m.results) > 0 {
		m.reveal.Gen++
		m.statsReveal.Gen++
		if m.reduceMotion {
			m.reveal.Rows = min(len(m.results), m.visibleResultRows())
			m.reveal.Revealing = false
			m.statsReveal.Revealed = m.statsRevealTargetLines()
		} else {
			m.reveal.Rows = 0
			m.reveal.Revealing = true
			revealGen := m.reveal.Gen
			cmds = append(cmds, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
				return revealRowTickMsg{gen: revealGen}
			}))

			m.statsReveal.Revealed = 0
			statsGen := m.statsReveal.Gen
			cmds = append(cmds, tea.Tick(m.statsRevealTickDuration(), func(time.Time) tea.Msg {
				return statsRevealTickMsg{gen: statsGen}
			}))
		}

		m.statsAnim.ValueTweenOn = false
		m.statsAnim.ValueStep = 0
		m.statsAnim.ValueSteps = 0
		m.statsAnim.DeltaTicks = 0

		if prevStats.Count > 0 && !m.reduceMotion {
			m.statsAnim.FromStats = prevStats
			m.statsAnim.ToStats = m.extendedStats
			m.statsAnim.ValueTweenOn = true
			m.statsAnim.ValueStep = 0
			m.statsAnim.ValueSteps = 15
			m.statsAnim.DeltaTotal = 24
			m.statsAnim.DeltaTicks = m.statsAnim.DeltaTotal
			m.statsAnim.DeltaMin = m.extendedStats.Min - prevStats.Min
			m.statsAnim.DeltaMax = m.extendedStats.Max - prevStats.Max
			m.statsAnim.DeltaAvg = m.extendedStats.Average - prevStats.Average
			m.statsAnim.DeltaMedian = m.extendedStats.Median - prevStats.Median
			m.statsAnim.DeltaP25 = m.extendedStats.P25 - prevStats.P25
			m.statsAnim.DeltaP75 = m.extendedStats.P75 - prevStats.P75
			m.statsAnim.ValueTweenGen++
			valueGen := m.statsAnim.ValueTweenGen
			cmds = append(cmds, tea.Tick(33*time.Millisecond, func(time.Time) tea.Msg {
				return statsValueTickMsg{gen: valueGen}
			}))
		}
	} else {
		m.reveal.Revealing = false
		m.reveal.Rows = 0
		m.statsReveal.Revealed = 0
		m.statsAnim.ValueTweenOn = false
		m.statsAnim.ValueStep = 0
		m.statsAnim.ValueSteps = 0
		m.statsAnim.DeltaTicks = 0
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.intro.Show {
		m.intro.Show = false
		m.intro.Completed = true
		return m, nil
	}

	// Let text inputs accept literal "m". Use motion toggle from non-input panels.
	if key.Matches(msg, m.keys.ToggleAnim) &&
		m.focusedPanel != panelSearch &&
		m.focusedPanel != panelCalculator {
		m = m.toggleReduceMotion()
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.ForceQuit):
		m.cancelActiveSearch()
		return m, tea.Quit

	case key.Matches(msg, m.keys.Quit):
		if m.focusedPanel != panelSearch && m.focusedPanel != panelCalculator {
			m.cancelActiveSearch()
			return m, tea.Quit
		}

	case key.Matches(msg, m.keys.Tab):
		// Prioritize textinput autocomplete in search panel before global focus cycling.
		if m.focusedPanel == panelSearch &&
			m.searchInput.ShowSuggestions &&
			m.hasSearchSuggestionMatch() {
			return m.handleSearchKeys(msg)
		}

		nextPanel := (m.focusedPanel + 1) % 5
		return m.changeFocus(nextPanel)

	case key.Matches(msg, m.keys.ShiftTab):
		prevPanel := m.focusedPanel - 1
		if prevPanel < 0 {
			prevPanel = panelHistory
		}
		return m.changeFocus(prevPanel)

	case key.Matches(msg, m.keys.Search):
		return m.changeFocus(panelSearch)

	case key.Matches(msg, m.keys.Calculator):
		if m.focusedPanel != panelSearch && m.focusedPanel != panelCalculator {
			return m.changeFocus(panelCalculator)
		}

	case key.Matches(msg, m.keys.Escape):
		return m.changeFocus(panelResults)
	}

	switch m.focusedPanel {
	case panelSearch:
		return m.handleSearchKeys(msg)
	case panelResults:
		return m.handleResultsKeys(msg)
	case panelStats:
		return m.handleStatsKeys(msg)
	case panelCalculator:
		return m.handleCalculatorKeys(msg)
	case panelHistory:
		return m.handleHistoryKeys(msg)
	}

	return m, nil
}

func (m Model) handleStatsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.StatsSum):
		return m.changeStatsViewMode(idea.StatsViewSummary)
	case key.Matches(msg, m.keys.StatsDist):
		return m.changeStatsViewMode(idea.StatsViewDistribution)
	case key.Matches(msg, m.keys.StatsMkt):
		return m.changeStatsViewMode(idea.StatsViewMarket)
	default:
		return m, nil
	}
}

func (m Model) toggleReduceMotion() Model {
	m.reduceMotion = !m.reduceMotion
	if !m.reduceMotion {
		return m
	}

	// Snap every animation channel to a stable resting state immediately.
	m.focusFlash.Gen++
	m.focusFlash.Active = false
	m.focusFlash.Ticks = 0

	m.reveal.Gen++
	m.reveal.Revealing = false
	if len(m.results) > 0 {
		m.reveal.Rows = min(len(m.results), m.visibleResultRows())
	} else {
		m.reveal.Rows = 0
	}

	m.statsReveal.Gen++
	if len(m.results) > 0 {
		m.statsReveal.Revealed = m.statsRevealTargetLines()
	} else {
		m.statsReveal.Revealed = 0
	}

	m.statsAnim.ValueTweenGen++
	m.statsAnim.ValueTweenOn = false
	m.statsAnim.ValueStep = 0
	m.statsAnim.ValueSteps = 0
	m.statsAnim.DeltaTicks = 0
	m.statsAnim.DeltaTotal = 0
	m.statsAnim.DeltaMin = 0
	m.statsAnim.DeltaMax = 0
	m.statsAnim.DeltaAvg = 0
	m.statsAnim.DeltaMedian = 0
	m.statsAnim.DeltaP25 = 0
	m.statsAnim.DeltaP75 = 0

	return m
}

func (m Model) changeStatsViewMode(mode idea.StatsViewMode) (tea.Model, tea.Cmd) {
	if m.statsViewMode == mode {
		return m, nil
	}

	m.statsViewMode = mode
	if len(m.results) == 0 {
		m.statsReveal.Revealed = 0
		return m, nil
	}

	m.statsReveal.Gen++
	if m.reduceMotion {
		m.statsReveal.Revealed = m.statsRevealTargetLines()
		return m, nil
	}
	m.statsReveal.Revealed = 0
	statsGen := m.statsReveal.Gen

	return m, tea.Tick(m.statsRevealTickDuration(), func(time.Time) tea.Msg {
		return statsRevealTickMsg{gen: statsGen}
	})
}

func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Enter) {
		return m.startSearch(m.searchInput.Value(), true)
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.refreshSearchSuggestions()
	return m, cmd
}

func (m Model) handleResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Down) {
		if m.reveal.Revealing {
			m.reveal.Revealing = false
			m.reveal.Rows = len(m.results)
		}
		if m.selectedIndex < len(m.results)-1 {
			m.selectedIndex++
			visible := m.visibleResultRows()
			if m.selectedIndex >= m.resultsOffset+visible {
				m.resultsOffset = m.selectedIndex - visible + 1
			}
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Up) {
		if m.reveal.Revealing {
			m.reveal.Revealing = false
			m.reveal.Rows = len(m.results)
		}
		if m.selectedIndex > 0 {
			m.selectedIndex--
			if m.selectedIndex < m.resultsOffset {
				m.resultsOffset = m.selectedIndex
			}
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Enter) {
		if len(m.results) > 0 && m.selectedIndex < len(m.results) {
			url := m.results[m.selectedIndex].URL
			if url != "" {
				return m, openURLCmd(url)
			}
		}
	}

	return m, nil
}

func (m Model) handleCalculatorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Enter) {
		if val, err := strconv.ParseFloat(m.costInput.Value(), 64); err == nil {
			m.cost = val
		} else {
			m.cost = 0
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.costInput, cmd = m.costInput.Update(msg)
	if val, err := strconv.ParseFloat(m.costInput.Value(), 64); err == nil {
		m.cost = val
	} else {
		m.cost = 0
	}
	return m, cmd
}

func (m Model) handleHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.HistNext) {
		if m.historyIndex < len(m.history)-1 {
			m.historyIndex++
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.HistPrev) {
		if m.historyIndex > 0 {
			m.historyIndex--
		}
		return m, nil
	}

	if key.Matches(msg, m.keys.Enter) {
		if len(m.history) > 0 && m.historyIndex < len(m.history) {
			query := m.history[m.historyIndex]
			m.searchInput.SetValue(query)
			return m.startSearch(query, false)
		}
	}

	return m, nil
}

// updateFocus manages focus state for text inputs.
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
	m.focusFlash.Gen++
	if m.reduceMotion {
		m.focusFlash.Ticks = 0
		m.focusFlash.Active = false
		return m, nil
	}
	m.focusFlash.Ticks = 3
	m.focusFlash.Active = true
	gen := m.focusFlash.Gen

	return m, tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg {
		return focusFlashTickMsg{gen: gen}
	})
}

// addToHistory adds a search query to history (avoiding duplicates).
func (m *Model) addToHistory(query string) {
	for i, h := range m.history {
		if h == query {
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}
	m.history = append([]string{query}, m.history...)
	if len(m.history) > 20 {
		m.history = m.history[:20]
	}
}

func (m *Model) refreshSearchSuggestions() {
	prefix := strings.TrimSpace(m.searchInput.Value())
	if len(prefix) < 2 {
		m.searchInput.SetSuggestions(nil)
		return
	}

	history := filterHistory(m.history, prefix, 4)
	var products []string
	if m.productIndex != nil {
		products = m.productIndex.Suggest(prefix)
	}

	m.searchInput.SetSuggestions(mergeSuggestions(history, products, 8))
}

func filterHistory(history []string, prefix string, limit int) []string {
	if len(history) == 0 || limit <= 0 {
		return nil
	}

	p := strings.ToLower(strings.TrimSpace(prefix))
	if p == "" {
		return nil
	}

	out := make([]string, 0, limit)
	for _, item := range history {
		if strings.HasPrefix(strings.ToLower(item), p) {
			out = append(out, item)
			if len(out) == limit {
				break
			}
		}
	}
	return out
}

func mergeSuggestions(first, second []string, limit int) []string {
	if limit <= 0 {
		return nil
	}

	out := make([]string, 0, limit)
	seen := map[string]struct{}{}

	appendUnique := func(items []string) {
		for _, item := range items {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			key := strings.ToLower(trimmed)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, trimmed)
			if len(out) == limit {
				return
			}
		}
	}

	appendUnique(first)
	if len(out) < limit {
		appendUnique(second)
	}
	return out
}

func (m Model) hasSearchSuggestionMatch() bool {
	if len(m.searchInput.MatchedSuggestions()) > 0 {
		return true
	}

	prefix := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if prefix == "" {
		return false
	}

	for _, suggestion := range m.searchInput.AvailableSuggestions() {
		if strings.HasPrefix(strings.ToLower(suggestion), prefix) {
			return true
		}
	}
	return false
}

func (m *Model) cancelActiveSearch() {
	if m.searchCancel == nil {
		return
	}
	m.searchCancel()
	m.searchCancel = nil
}

func (m Model) startSearch(rawQuery string, addToHistory bool) (tea.Model, tea.Cmd) {
	query := strings.TrimSpace(rawQuery)
	if query == "" {
		return m, nil
	}

	expandedQuery := query
	if m.productIndex != nil {
		expandedQuery = m.productIndex.Expand(query)
	}

	m.cancelActiveSearch()
	ctx, cancel := context.WithCancel(context.Background())
	m.searchCancel = cancel
	m.searchGen++

	m.loading = true
	m.loadingDots = 0
	m.statsAnim.SkeletonFrame = 0
	m.statsAnim.ValueTweenOn = false
	m.statsAnim.ValueStep = 0
	m.statsAnim.ValueSteps = 0
	m.statsAnim.DeltaTicks = 0
	m.warning = ""
	m.err = nil
	if addToHistory {
		m.addToHistory(query)
	}

	return m, m.doSearch(ctx, expandedQuery, m.searchGen)
}

// doSearch creates a command to fetch search results.
func (m Model) doSearch(ctx context.Context, query string, gen int) tea.Cmd {
	client := m.apiClient
	if client == nil {
		client = api.NewEnvClient()
	}

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			response := client.SearchPricesContext(ctx, strings.TrimSpace(query))
			return SearchResultsMsg{
				Results: response.Results,
				Mode:    response.Mode,
				Warning: response.Warning,
				Err:     response.Err,
				gen:     gen,
			}
		},
	)
}

func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		return openURLResultMsg{Err: openURL(url)}
	}
}

// openURL opens a URL in the default browser.
func validateOpenURL(rawURL string) (*url.URL, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return nil, fmt.Errorf("open URL: empty URL")
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("open URL: invalid URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("open URL: invalid URL %q", trimmed)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("open URL: unsupported URL scheme %q", parsedURL.Scheme)
	}

	return parsedURL, nil
}

func openURL(rawURL string) error {
	parsedURL, err := validateOpenURL(rawURL)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", parsedURL.String())
	case "linux":
		cmd = exec.Command("xdg-open", parsedURL.String())
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", parsedURL.String())
	}
	if cmd != nil {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("open URL: %w", err)
		}
		return nil
	}
	return fmt.Errorf("open URL: unsupported platform %q", runtime.GOOS)
}

func (m Model) statsRevealTargetLines() int {
	switch m.statsViewMode {
	case idea.StatsViewDistribution:
		stats := m.extendedStats
		if len(stats.Histogram) == 0 {
			return 2
		}
		bins := len(stats.Histogram)
		if bins > 6 {
			bins = 6
		}
		return bins + 1
	case idea.StatsViewMarket:
		stats := m.extendedStats
		if len(stats.PlatformStats) == 0 {
			return 2
		}
		platformLines := len(stats.PlatformStats)
		if platformLines > 4 {
			platformLines = 4
		}
		return platformLines + 2
	default:
		return 6
	}
}

func (m Model) statsRevealTickDuration() time.Duration {
	switch m.statsViewMode {
	case idea.StatsViewDistribution, idea.StatsViewMarket:
		return 20 * time.Millisecond
	default:
		return 40 * time.Millisecond
	}
}
