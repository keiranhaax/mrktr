package main

import (
	"strings"
	"testing"

	"mrktr/types"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchInputHandlesSingleKeyOnce(t *testing.T) {
	m := NewModel()
	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if got := m.searchInput.Value(); got != "a" {
		t.Fatalf("expected search input to be %q, got %q", "a", got)
	}
}

func TestCalculatorInputHandlesSingleKeyOnce(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelCalculator
	m = m.updateFocus()

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if got := m.costInput.Value(); got != "1" {
		t.Fatalf("expected cost input to be %q, got %q", "1", got)
	}
}

func TestCKeyTypesInCalculatorInsteadOfRefocus(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelCalculator
	m = m.updateFocus()

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if m.focusedPanel != panelCalculator {
		t.Fatalf("expected focus to remain on calculator, got panel %d", m.focusedPanel)
	}
	if got := m.costInput.Value(); got != "c" {
		t.Fatalf("expected calculator input to contain typed c, got %q", got)
	}
}

func TestCKeyFocusesCalculatorFromResults(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if m.focusedPanel != panelCalculator {
		t.Fatalf("expected calculator panel to be focused, got %d", m.focusedPanel)
	}
}

func TestHistoryEnterReplaysSelectedQuery(t *testing.T) {
	m := NewModel()
	m.history = []string{"ps5", "switch"}
	m.historyIndex = 1
	m.focusedPanel = panelHistory
	m = m.updateFocus()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected search command to be returned on history enter")
	}
	if !um.loading {
		t.Fatal("expected loading state to be true after history replay")
	}
	if got := um.searchInput.Value(); got != "switch" {
		t.Fatalf("expected replayed query %q, got %q", "switch", got)
	}
}

func TestResultsEnterReturnsOpenURLCommand(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelResults
	m.results = []types.Listing{{URL: "https://example.com"}}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected open URL command when enter is pressed on a result")
	}
}

func TestRenderHistoryPanelShowsSelectionMarker(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelHistory
	m.history = []string{"ps5", "switch", "airpods"}
	m.historyIndex = 1

	got := m.renderHistoryPanel(120, 2)
	if !strings.Contains(got, "> switch") {
		t.Fatalf("expected selected history marker in view, got: %s", got)
	}
}

func TestViewShowsSmallTerminalMessage(t *testing.T) {
	m := NewModel()
	m.width = 50
	m.height = 10

	out := m.View()
	if !strings.Contains(out, "Terminal too small") {
		t.Fatalf("expected small-terminal warning, got: %s", out)
	}
}

func TestSpinnerActiveWhileLoading(t *testing.T) {
	m := NewModel()
	m.loading = true

	if got := strings.TrimSpace(m.spinner.View()); got == "" {
		t.Fatal("expected spinner view to be non-empty while loading")
	}
}

func TestScrollOffsetOnNavigation(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()
	m.height = 16
	m.results = makeListings(20)

	for i := 0; i < 3; i++ {
		m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}
	if m.selectedIndex != 3 {
		t.Fatalf("expected selected index 3 after moving down, got %d", m.selectedIndex)
	}
	if m.resultsOffset != 2 {
		t.Fatalf("expected results offset 2 after moving down, got %d", m.resultsOffset)
	}

	for i := 0; i < 2; i++ {
		m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	}
	if m.selectedIndex != 1 {
		t.Fatalf("expected selected index 1 after moving up, got %d", m.selectedIndex)
	}
	if m.resultsOffset != 1 {
		t.Fatalf("expected results offset 1 after moving up, got %d", m.resultsOffset)
	}
}

func TestScrollOffsetResetsOnNewResults(t *testing.T) {
	m := NewModel()
	m.results = makeListings(20)
	m.selectedIndex = 7
	m.resultsOffset = 5

	updated, _ := m.Update(SearchResultsMsg{
		Results: makeListings(3),
		Mode:    searchModeDemo,
	})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.selectedIndex != 0 {
		t.Fatalf("expected selected index reset to 0, got %d", um.selectedIndex)
	}
	if um.resultsOffset != 0 {
		t.Fatalf("expected results offset reset to 0, got %d", um.resultsOffset)
	}
}

func TestScrollOffsetClampsOnResize(t *testing.T) {
	m := NewModel()
	m.results = makeListings(20)
	m.resultsOffset = 15

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 25})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.resultsOffset != 11 {
		t.Fatalf("expected clamped results offset 11, got %d", um.resultsOffset)
	}
}

func TestFocusFlashGeneration(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelSearch
	m = m.updateFocus()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected focus flash tick command on focus change")
	}
	if !um.focusFlashActive {
		t.Fatal("expected focus flash to be active after tab")
	}
	if um.focusFlashTicks != 3 {
		t.Fatalf("expected 3 flash ticks after focus change, got %d", um.focusFlashTicks)
	}
	if um.focusFlashGen != 1 {
		t.Fatalf("expected flash generation 1, got %d", um.focusFlashGen)
	}

	staleUpdated, _ := um.Update(focusFlashTickMsg{gen: 0})
	staleModel, ok := staleUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", um, staleUpdated)
	}
	if staleModel.focusFlashTicks != 3 {
		t.Fatalf("expected stale flash tick to be ignored, got ticks=%d", staleModel.focusFlashTicks)
	}

	currentUpdated, _ := staleModel.Update(focusFlashTickMsg{gen: 1})
	currentModel, ok := currentUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", staleModel, currentUpdated)
	}
	if currentModel.focusFlashTicks != 2 {
		t.Fatalf("expected current flash tick to decrement to 2, got %d", currentModel.focusFlashTicks)
	}
}

func TestRevealInterruptOnNavigate(t *testing.T) {
	m := NewModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()
	m.height = 24
	m.results = makeListings(10)
	m.revealing = true
	m.revealedRows = 2

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.revealing {
		t.Fatal("expected reveal animation to stop on navigation")
	}
	if m.revealedRows != len(m.results) {
		t.Fatalf("expected revealed rows to jump to full length %d, got %d", len(m.results), m.revealedRows)
	}
}

func TestRevealGeneration(t *testing.T) {
	m := NewModel()
	m.height = 24
	m.results = makeListings(12)
	m.revealing = true
	m.revealedRows = 1
	m.revealGen = 2

	staleUpdated, _ := m.Update(revealRowTickMsg{gen: 1})
	staleModel, ok := staleUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, staleUpdated)
	}
	if staleModel.revealedRows != 1 {
		t.Fatalf("expected stale reveal tick to be ignored, got %d", staleModel.revealedRows)
	}

	currentUpdated, _ := staleModel.Update(revealRowTickMsg{gen: 2})
	currentModel, ok := currentUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", staleModel, currentUpdated)
	}
	if currentModel.revealedRows != 2 {
		t.Fatalf("expected current reveal tick to increment to 2, got %d", currentModel.revealedRows)
	}
}

func TestLoadingDotsReset(t *testing.T) {
	m := NewModel()
	m.searchInput.SetValue("ps5")
	m.loadingDots = 3
	m.focusedPanel = panelSearch
	m = m.updateFocus()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected search command on enter")
	}
	if !um.loading {
		t.Fatal("expected loading=true after search enter")
	}
	if um.loadingDots != 0 {
		t.Fatalf("expected loading dots reset to 0 on search enter, got %d", um.loadingDots)
	}

	updatedTick, _ := um.Update(spinner.TickMsg{})
	tickModel, ok := updatedTick.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", um, updatedTick)
	}
	if tickModel.loadingDots != 1 {
		t.Fatalf("expected loading dots to increment while loading, got %d", tickModel.loadingDots)
	}

	updatedDone, _ := tickModel.Update(SearchResultsMsg{
		Results: makeListings(2),
		Mode:    searchModeDemo,
	})
	doneModel, ok := updatedDone.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", tickModel, updatedDone)
	}
	if doneModel.loading {
		t.Fatal("expected loading=false after receiving search results")
	}
	before := doneModel.loadingDots
	afterUpdated, _ := doneModel.Update(spinner.TickMsg{})
	afterModel, ok := afterUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", doneModel, afterUpdated)
	}
	if afterModel.loadingDots != before {
		t.Fatalf("expected loading dots to stop incrementing when loading=false (before=%d after=%d)", before, afterModel.loadingDots)
	}

	m2 := NewModel()
	m2.focusedPanel = panelHistory
	m2 = m2.updateFocus()
	m2.history = []string{"switch"}
	m2.historyIndex = 0
	m2.loadingDots = 2

	updatedHistory, historyCmd := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	historyModel, ok := updatedHistory.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m2, updatedHistory)
	}
	if historyCmd == nil {
		t.Fatal("expected search command to be returned from history replay")
	}
	if historyModel.loadingDots != 0 {
		t.Fatalf("expected loading dots reset to 0 on history replay, got %d", historyModel.loadingDots)
	}
}

func makeListings(n int) []types.Listing {
	out := make([]types.Listing, n)
	for i := 0; i < n; i++ {
		out[i] = types.Listing{
			Platform:  "eBay",
			Price:     float64(100 + i),
			Condition: "Used",
			Status:    "Active",
			URL:       "https://example.com/item",
			Title:     "Item",
		}
	}
	return out
}

func sendKey(t *testing.T, m Model, msg tea.KeyMsg) Model {
	t.Helper()

	updated, _ := m.Update(msg)
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	return um
}
