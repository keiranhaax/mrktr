package main

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"mrktr/api"
	"mrktr/idea"
	"mrktr/types"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type captureQueryProvider struct {
	query string
}

type cancelAwareProvider struct {
	mu             sync.Mutex
	calls          int
	startedOnce    sync.Once
	canceledOnce   sync.Once
	firstStartedCh chan struct{}
	firstCancelCh  chan struct{}
}

func (p *captureQueryProvider) Name() string {
	return "Capture"
}

func (p *captureQueryProvider) Configured() bool {
	return true
}

func (p *captureQueryProvider) Search(_ context.Context, query string) ([]types.Listing, error) {
	p.query = query
	return []types.Listing{}, nil
}

func (p *cancelAwareProvider) Name() string {
	return "CancelAware"
}

func (p *cancelAwareProvider) Configured() bool {
	return true
}

func (p *cancelAwareProvider) Search(ctx context.Context, query string) ([]types.Listing, error) {
	p.mu.Lock()
	p.calls++
	callNum := p.calls
	p.mu.Unlock()

	if callNum == 1 {
		p.startedOnce.Do(func() {
			close(p.firstStartedCh)
		})
		<-ctx.Done()
		p.canceledOnce.Do(func() {
			close(p.firstCancelCh)
		})
		return nil, ctx.Err()
	}

	return []types.Listing{
		{
			Platform:  "eBay",
			Price:     199.0,
			Condition: "Used",
			Status:    "Active",
			URL:       "https://example.com/item",
			Title:     query,
		},
	}, nil
}

func TestSearchInputHandlesSingleKeyOnce(t *testing.T) {
	m := newTestModel()
	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if got := m.searchInput.Value(); got != "a" {
		t.Fatalf("expected search input to be %q, got %q", "a", got)
	}
}

func TestCalculatorInputHandlesSingleKeyOnce(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelCalculator
	m = m.updateFocus()

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if got := m.costInput.Value(); got != "1" {
		t.Fatalf("expected cost input to be %q, got %q", "1", got)
	}
}

func TestCKeyTypesInCalculatorInsteadOfRefocus(t *testing.T) {
	m := newTestModel()
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
	m := newTestModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if m.focusedPanel != panelCalculator {
		t.Fatalf("expected calculator panel to be focused, got %d", m.focusedPanel)
	}
}

func TestHistoryEnterReplaysSelectedQuery(t *testing.T) {
	m := newTestModel()
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
	m := newTestModel()
	m.focusedPanel = panelResults
	m.results = []types.Listing{{URL: "https://example.com"}}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected open URL command when enter is pressed on a result")
	}
}

func TestRenderHistoryPanelShowsSelectionMarker(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelHistory
	m.history = []string{"ps5", "switch", "airpods"}
	m.historyIndex = 1

	got := m.renderHistoryPanel(120, 2)
	if !strings.Contains(got, "> switch") {
		t.Fatalf("expected selected history marker in view, got: %s", got)
	}
}

func TestViewShowsSmallTerminalMessage(t *testing.T) {
	m := newTestModel()
	m.width = 50
	m.height = 10

	out := m.View()
	if !strings.Contains(out, "Terminal too small") {
		t.Fatalf("expected small-terminal warning, got: %s", out)
	}
}

func TestSpinnerActiveWhileLoading(t *testing.T) {
	m := newTestModel()
	m.loading = true

	if got := strings.TrimSpace(m.spinner.View()); got == "" {
		t.Fatal("expected spinner view to be non-empty while loading")
	}
}

func TestScrollOffsetOnNavigation(t *testing.T) {
	m := newTestModel()
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
	m := newTestModel()
	m.results = makeListings(20)
	m.selectedIndex = 7
	m.resultsOffset = 5

	updated, _ := m.Update(SearchResultsMsg{
		Results: makeListings(3),
		Mode:    api.SearchModeLive,
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

func TestSearchResultsPopulateExtendedStats(t *testing.T) {
	m := newTestModel()
	m.statsViewMode = idea.StatsViewMarket

	listings := []types.Listing{
		{Platform: "eBay", Price: 100, Condition: "Used", Status: "Sold"},
		{Platform: "Mercari", Price: 200, Condition: "New", Status: "Active"},
		{Platform: "eBay", Price: 300, Condition: "Used", Status: "Active"},
	}

	updated, _ := m.Update(SearchResultsMsg{
		Results: listings,
		Mode:    api.SearchModeLive,
	})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.statsViewMode != idea.StatsViewMarket {
		t.Fatalf("expected stats view mode to be preserved, got %v", um.statsViewMode)
	}

	if um.extendedStats.Count != len(listings) {
		t.Fatalf("expected extended stats count %d, got %d", len(listings), um.extendedStats.Count)
	}
	if um.extendedStats.Statistics != um.stats {
		t.Fatalf("expected base stats and embedded stats to match, base=%+v extended=%+v", um.stats, um.extendedStats.Statistics)
	}

	if um.extendedStats.SoldCount != 1 || um.extendedStats.ActiveCount != 2 {
		t.Fatalf("unexpected sold/active counts: sold=%d active=%d", um.extendedStats.SoldCount, um.extendedStats.ActiveCount)
	}

	eBayStats := um.extendedStats.PlatformStats["eBay"]
	if eBayStats.Count != 2 {
		t.Fatalf("expected eBay platform count 2, got %d", eBayStats.Count)
	}
}

func TestScrollOffsetClampsOnResize(t *testing.T) {
	m := newTestModel()
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
	m := newTestModel()
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
	if !um.focusFlash.Active {
		t.Fatal("expected focus flash to be active after tab")
	}
	if um.focusFlash.Ticks != 3 {
		t.Fatalf("expected 3 flash ticks after focus change, got %d", um.focusFlash.Ticks)
	}
	if um.focusFlash.Gen != 1 {
		t.Fatalf("expected flash generation 1, got %d", um.focusFlash.Gen)
	}

	staleUpdated, _ := um.Update(focusFlashTickMsg{gen: 0})
	staleModel, ok := staleUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", um, staleUpdated)
	}
	if staleModel.focusFlash.Ticks != 3 {
		t.Fatalf("expected stale flash tick to be ignored, got ticks=%d", staleModel.focusFlash.Ticks)
	}

	currentUpdated, _ := staleModel.Update(focusFlashTickMsg{gen: 1})
	currentModel, ok := currentUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", staleModel, currentUpdated)
	}
	if currentModel.focusFlash.Ticks != 2 {
		t.Fatalf("expected current flash tick to decrement to 2, got %d", currentModel.focusFlash.Ticks)
	}
}

func TestTabAcceptsSearchSuggestionBeforeChangingPanel(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelSearch
	m = m.updateFocus()
	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.focusedPanel != panelSearch {
		t.Fatalf("expected focus to stay on search panel, got %d", um.focusedPanel)
	}
	if got := um.searchInput.Value(); got == "swi" {
		t.Fatalf("expected tab to accept suggestion value, input remained %q", got)
	}
}

func TestTabChangesPanelWhenSearchSuggestionDoesNotMatch(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelSearch
	m = m.updateFocus()
	m.searchInput.SetValue("zzz")
	m.searchInput.SetSuggestions([]string{"Nintendo Switch OLED"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.focusedPanel != panelResults {
		t.Fatalf("expected tab to change focus to results panel, got %d", um.focusedPanel)
	}
}

func TestStatsViewKeysIgnoredOutsideStatsPanel(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()
	m.statsViewMode = idea.StatsViewSummary

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.statsViewMode != idea.StatsViewSummary {
		t.Fatalf("expected stats mode unchanged outside stats panel, got %v", um.statsViewMode)
	}
}

func TestStatsViewKeysSwitchWhenStatsPanelFocused(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelStats
	m = m.updateFocus()
	m.results = makeListings(5)
	m.statsViewMode = idea.StatsViewSummary
	m.statsReveal.Gen = 4
	m.statsReveal.Revealed = 3

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.focusedPanel != panelStats {
		t.Fatalf("expected stats panel focus to remain, got %d", um.focusedPanel)
	}
	if um.statsViewMode != idea.StatsViewDistribution {
		t.Fatalf("expected stats mode distribution, got %v", um.statsViewMode)
	}
	if um.statsReveal.Revealed != 0 {
		t.Fatalf("expected stats reveal reset to 0, got %d", um.statsReveal.Revealed)
	}
	if um.statsReveal.Gen != 5 {
		t.Fatalf("expected stats reveal generation increment to 5, got %d", um.statsReveal.Gen)
	}
	if cmd == nil {
		t.Fatal("expected stats reveal tick command after stats mode change")
	}
}

func TestSearchResultsStartStatsValueTweenWhenPreviousStatsExist(t *testing.T) {
	m := newTestModel()
	m.extendedStats = idea.CalculateExtendedStats([]types.Listing{
		{Platform: "eBay", Price: 100, Condition: "Used", Status: "Active"},
		{Platform: "eBay", Price: 110, Condition: "Used", Status: "Active"},
	})
	m.stats = m.extendedStats.Statistics

	updated, cmd := m.Update(SearchResultsMsg{
		Results: []types.Listing{
			{Platform: "eBay", Price: 200, Condition: "Used", Status: "Active"},
			{Platform: "Mercari", Price: 260, Condition: "New", Status: "Sold"},
		},
		Mode: api.SearchModeLive,
	})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected command batch on search results")
	}

	if !um.statsAnim.ValueTweenOn {
		t.Fatal("expected stats value tween to start")
	}
	if um.statsAnim.ValueSteps <= 0 {
		t.Fatalf("expected positive value tween steps, got %d", um.statsAnim.ValueSteps)
	}
	if um.statsAnim.DeltaTicks <= 0 {
		t.Fatalf("expected positive delta ticks, got %d", um.statsAnim.DeltaTicks)
	}
	if um.statsAnim.ValueTweenGen == 0 {
		t.Fatal("expected non-zero tween generation")
	}
}

func TestStatsValueTickAdvancesTweenAndDelta(t *testing.T) {
	m := newTestModel()
	m.statsAnim.ValueTweenGen = 7
	m.statsAnim.ValueTweenOn = true
	m.statsAnim.ValueStep = 0
	m.statsAnim.ValueSteps = 2
	m.statsAnim.DeltaTotal = 3
	m.statsAnim.DeltaTicks = 3

	updated, cmd := m.Update(statsValueTickMsg{gen: 7})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected follow-up tick command while tween/delta active")
	}
	if um.statsAnim.ValueStep != 1 {
		t.Fatalf("expected tween step to increment to 1, got %d", um.statsAnim.ValueStep)
	}
	if um.statsAnim.DeltaTicks != 2 {
		t.Fatalf("expected delta ticks to decrement to 2, got %d", um.statsAnim.DeltaTicks)
	}

	updated2, cmd2 := um.Update(statsValueTickMsg{gen: 7})
	um2, ok := updated2.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", um, updated2)
	}
	if cmd2 == nil {
		t.Fatal("expected follow-up tick command while delta is still active")
	}
	if um2.statsAnim.ValueTweenOn {
		t.Fatal("expected tween to end at configured step count")
	}
	if um2.statsAnim.ValueStep != 2 {
		t.Fatalf("expected tween step clamped at 2, got %d", um2.statsAnim.ValueStep)
	}
	if um2.statsAnim.DeltaTicks != 1 {
		t.Fatalf("expected delta ticks to decrement to 1, got %d", um2.statsAnim.DeltaTicks)
	}

	updated3, cmd3 := um2.Update(statsValueTickMsg{gen: 7})
	um3, ok := updated3.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", um2, updated3)
	}
	if cmd3 != nil {
		t.Fatal("expected no follow-up tick command after tween and delta complete")
	}
	if um3.statsAnim.DeltaTicks != 0 {
		t.Fatalf("expected delta ticks to reach 0, got %d", um3.statsAnim.DeltaTicks)
	}
}

func TestReduceMotionSkipsStatsRevealAnimationOnTabChange(t *testing.T) {
	m := newTestModel()
	m.reduceMotion = true
	m.focusedPanel = panelStats
	m = m.updateFocus()
	m.results = makeListings(5)
	m.extendedStats = idea.CalculateExtendedStats(m.results)
	m.stats = m.extendedStats.Statistics
	m.statsViewMode = idea.StatsViewSummary
	m.statsReveal.Revealed = 0

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if cmd != nil {
		t.Fatal("expected no reveal animation command when reduceMotion=true")
	}
	if um.statsViewMode != idea.StatsViewDistribution {
		t.Fatalf("expected distribution mode, got %v", um.statsViewMode)
	}
	if um.statsReveal.Revealed != um.statsRevealTargetLines() {
		t.Fatalf("expected stats reveal to jump to target (%d), got %d", um.statsRevealTargetLines(), um.statsReveal.Revealed)
	}
}

func TestReduceMotionSkipsStatsValueTweenOnSearchResults(t *testing.T) {
	m := newTestModel()
	m.reduceMotion = true
	m.extendedStats = idea.CalculateExtendedStats([]types.Listing{
		{Platform: "eBay", Price: 100, Status: "Active"},
		{Platform: "eBay", Price: 120, Status: "Sold"},
	})
	m.stats = m.extendedStats.Statistics

	updated, _ := m.Update(SearchResultsMsg{
		Results: []types.Listing{
			{Platform: "Mercari", Price: 240, Status: "Active"},
			{Platform: "Amazon", Price: 260, Status: "Sold"},
		},
		Mode: api.SearchModeLive,
	})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if um.statsAnim.ValueTweenOn {
		t.Fatal("expected value tween to remain disabled in reduceMotion mode")
	}
	if um.statsAnim.DeltaTicks != 0 {
		t.Fatalf("expected no delta ticks in reduceMotion mode, got %d", um.statsAnim.DeltaTicks)
	}
}

func TestReduceMotionSkipsFocusFlashAnimation(t *testing.T) {
	m := newTestModel()
	m.reduceMotion = true
	m.focusedPanel = panelSearch
	m = m.updateFocus()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}

	if cmd != nil {
		t.Fatal("expected no focus flash command in reduceMotion mode")
	}
	if um.focusFlash.Active || um.focusFlash.Ticks != 0 {
		t.Fatalf("expected inactive focus flash in reduceMotion mode, got active=%v ticks=%d", um.focusFlash.Active, um.focusFlash.Ticks)
	}
}

func TestGlobalMotionToggleFlipsSetting(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()

	updatedOn, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	onModel, ok := updatedOn.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updatedOn)
	}
	if !onModel.reduceMotion {
		t.Fatal("expected global m key to enable reduceMotion")
	}

	updatedOff, _ := onModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	offModel, ok := updatedOff.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", onModel, updatedOff)
	}
	if offModel.reduceMotion {
		t.Fatal("expected second global m key to disable reduceMotion")
	}
}

func TestGlobalMotionToggleSettlesInFlightAnimations(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()
	m.results = makeListings(10)
	m.extendedStats = idea.CalculateExtendedStats(m.results)
	m.stats = m.extendedStats.Statistics
	m.focusFlash.Gen = 2
	m.focusFlash.Active = true
	m.focusFlash.Ticks = 2
	m.reveal.Gen = 3
	m.reveal.Revealing = true
	m.reveal.Rows = 1
	m.statsReveal.Gen = 4
	m.statsReveal.Revealed = 1
	m.statsAnim.ValueTweenGen = 5
	m.statsAnim.ValueTweenOn = true
	m.statsAnim.ValueStep = 6
	m.statsAnim.ValueSteps = 15
	m.statsAnim.DeltaTotal = 24
	m.statsAnim.DeltaTicks = 8
	m.statsAnim.DeltaAvg = 10
	m.statsAnim.DeltaMin = -3

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if !um.reduceMotion {
		t.Fatal("expected reduceMotion to be enabled")
	}
	if um.focusFlash.Active || um.focusFlash.Ticks != 0 {
		t.Fatalf("expected focus flash to be disabled, got active=%v ticks=%d", um.focusFlash.Active, um.focusFlash.Ticks)
	}
	if um.reveal.Revealing {
		t.Fatal("expected results reveal animation to stop")
	}
	expectedRows := min(len(um.results), um.visibleResultRows())
	if um.reveal.Rows != expectedRows {
		t.Fatalf("expected reveal rows to snap to %d, got %d", expectedRows, um.reveal.Rows)
	}
	if um.statsReveal.Revealed != um.statsRevealTargetLines() {
		t.Fatalf("expected stats reveal to snap to target %d, got %d", um.statsRevealTargetLines(), um.statsReveal.Revealed)
	}
	if um.statsAnim.ValueTweenOn {
		t.Fatal("expected stats value tween to stop")
	}
	if um.statsAnim.ValueStep != 0 || um.statsAnim.ValueSteps != 0 {
		t.Fatalf("expected stats tween counters reset, got step=%d steps=%d", um.statsAnim.ValueStep, um.statsAnim.ValueSteps)
	}
	if um.statsAnim.DeltaTicks != 0 || um.statsAnim.DeltaTotal != 0 {
		t.Fatalf("expected delta animation to reset, got ticks=%d total=%d", um.statsAnim.DeltaTicks, um.statsAnim.DeltaTotal)
	}
	if um.statsAnim.DeltaAvg != 0 || um.statsAnim.DeltaMin != 0 {
		t.Fatalf("expected stored delta values reset, got avg=%f min=%f", um.statsAnim.DeltaAvg, um.statsAnim.DeltaMin)
	}
	if um.statsAnim.ValueTweenGen != m.statsAnim.ValueTweenGen+1 {
		t.Fatalf("expected tween generation to increment, got %d", um.statsAnim.ValueTweenGen)
	}
}

func TestGlobalMotionToggleDoesNotOverrideSearchInputRune(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelSearch
	m = m.updateFocus()
	m.searchInput.SetValue("")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if um.reduceMotion {
		t.Fatal("expected search input to keep literal m behavior without toggling motion")
	}
	if got := um.searchInput.Value(); got != "m" {
		t.Fatalf("expected m key to type into search input, got %q", got)
	}
}

func TestSearchSuggestionsMergeHistoryAndCatalogMatches(t *testing.T) {
	m := newTestModel()
	m.history = []string{"switch carrying case", "ps5"}
	m.searchInput.SetValue("sw")

	m.refreshSearchSuggestions()
	got := m.searchInput.AvailableSuggestions()
	if len(got) == 0 {
		t.Fatal("expected suggestions for prefix")
	}
	if got[0] != "switch carrying case" {
		t.Fatalf("expected history suggestion first, got %q", got[0])
	}

	foundCatalogSuggestion := false
	for _, suggestion := range got {
		if strings.Contains(strings.ToLower(suggestion), "switch") &&
			suggestion != "switch carrying case" {
			foundCatalogSuggestion = true
			break
		}
	}
	if !foundCatalogSuggestion {
		t.Fatalf("expected catalog-backed suggestion in %v", got)
	}
}

func TestRevealInterruptOnNavigate(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelResults
	m = m.updateFocus()
	m.height = 24
	m.results = makeListings(10)
	m.reveal.Revealing = true
	m.reveal.Rows = 2

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.reveal.Revealing {
		t.Fatal("expected reveal animation to stop on navigation")
	}
	if m.reveal.Rows != len(m.results) {
		t.Fatalf("expected revealed rows to jump to full length %d, got %d", len(m.results), m.reveal.Rows)
	}
}

func TestRevealGeneration(t *testing.T) {
	m := newTestModel()
	m.height = 24
	m.results = makeListings(12)
	m.reveal.Revealing = true
	m.reveal.Rows = 1
	m.reveal.Gen = 2

	staleUpdated, _ := m.Update(revealRowTickMsg{gen: 1})
	staleModel, ok := staleUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, staleUpdated)
	}
	if staleModel.reveal.Rows != 1 {
		t.Fatalf("expected stale reveal tick to be ignored, got %d", staleModel.reveal.Rows)
	}

	currentUpdated, _ := staleModel.Update(revealRowTickMsg{gen: 2})
	currentModel, ok := currentUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", staleModel, currentUpdated)
	}
	if currentModel.reveal.Rows != 2 {
		t.Fatalf("expected current reveal tick to increment to 2, got %d", currentModel.reveal.Rows)
	}
}

func TestLoadingDotsReset(t *testing.T) {
	m := newTestModel()
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
	if um.statsAnim.SkeletonFrame != 0 {
		t.Fatalf("expected stats skeleton frame reset to 0 on search enter, got %d", um.statsAnim.SkeletonFrame)
	}

	updatedTick, _ := um.Update(spinner.TickMsg{})
	tickModel, ok := updatedTick.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", um, updatedTick)
	}
	if tickModel.loadingDots != 1 {
		t.Fatalf("expected loading dots to increment while loading, got %d", tickModel.loadingDots)
	}
	if tickModel.statsAnim.SkeletonFrame != 1 {
		t.Fatalf("expected stats skeleton frame to increment while loading, got %d", tickModel.statsAnim.SkeletonFrame)
	}

	updatedDone, _ := tickModel.Update(SearchResultsMsg{
		Results: makeListings(2),
		Mode:    api.SearchModeLive,
	})
	doneModel, ok := updatedDone.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", tickModel, updatedDone)
	}
	if doneModel.loading {
		t.Fatal("expected loading=false after receiving search results")
	}
	before := doneModel.loadingDots
	beforeSkeleton := doneModel.statsAnim.SkeletonFrame
	afterUpdated, _ := doneModel.Update(spinner.TickMsg{})
	afterModel, ok := afterUpdated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", doneModel, afterUpdated)
	}
	if afterModel.loadingDots != before {
		t.Fatalf("expected loading dots to stop incrementing when loading=false (before=%d after=%d)", before, afterModel.loadingDots)
	}
	if afterModel.statsAnim.SkeletonFrame != beforeSkeleton {
		t.Fatalf("expected skeleton frame to stop incrementing when loading=false (before=%d after=%d)", beforeSkeleton, afterModel.statsAnim.SkeletonFrame)
	}

	m2 := newTestModel()
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
	if historyModel.statsAnim.SkeletonFrame != 0 {
		t.Fatalf("expected stats skeleton frame reset to 0 on history replay, got %d", historyModel.statsAnim.SkeletonFrame)
	}
}

func TestSearchEnterWhileLoadingStartsReplacementSearch(t *testing.T) {
	m := newTestModel()
	m.loading = true
	m.searchGen = 3
	canceled := false
	m.searchCancel = func() {
		canceled = true
	}
	m.searchInput.SetValue("ps5")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected replacement search command while loading")
	}
	if !canceled {
		t.Fatal("expected prior in-flight search context to be canceled")
	}
	if um.searchGen != 4 {
		t.Fatalf("expected search generation increment to 4, got %d", um.searchGen)
	}
	if got := len(um.history); got != 1 {
		t.Fatalf("expected replacement search to update history, got %d items", got)
	}
}

func TestHistoryEnterWhileLoadingStartsReplacementSearch(t *testing.T) {
	m := newTestModel()
	m.loading = true
	m.searchGen = 8
	canceled := false
	m.searchCancel = func() {
		canceled = true
	}
	m.history = []string{"ps5"}
	m.historyIndex = 0
	m.focusedPanel = panelHistory
	m = m.updateFocus()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected history replay replacement command while loading")
	}
	if !canceled {
		t.Fatal("expected prior in-flight search context to be canceled")
	}
	if got := um.searchInput.Value(); got != "ps5" {
		t.Fatalf("expected search input to switch to replay query, got %q", got)
	}
	if um.searchGen != 9 {
		t.Fatalf("expected search generation increment to 9, got %d", um.searchGen)
	}
}

func TestHistoryEnterUsesExpandedQuery(t *testing.T) {
	provider := &captureQueryProvider{}
	m := newTestModel()
	m.apiClient = api.NewClient(provider)
	m.history = []string{"ps5"}
	m.historyIndex = 0
	m.focusedPanel = panelHistory
	m = m.updateFocus()

	wantQuery := "ps5"
	if m.productIndex != nil {
		wantQuery = m.productIndex.Expand("ps5")
	}
	if wantQuery == "ps5" {
		t.Fatal("expected test precondition to expand ps5 query")
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated)
	}
	if cmd == nil {
		t.Fatal("expected history replay command")
	}

	cmdMsg := cmd()
	batchMsg, ok := cmdMsg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected batch command message, got %T", cmdMsg)
	}
	if len(batchMsg) == 0 {
		t.Fatal("expected batch message to contain commands")
	}

	lastCmd := batchMsg[len(batchMsg)-1]
	if lastCmd == nil {
		t.Fatal("expected search command in batch")
	}
	if _, ok := lastCmd().(SearchResultsMsg); !ok {
		t.Fatalf("expected search command to return SearchResultsMsg")
	}
	if provider.query != wantQuery {
		t.Fatalf("expected expanded query %q, got %q", wantQuery, provider.query)
	}
}

func TestCalculatorInvalidInputResetsCost(t *testing.T) {
	m := newTestModel()
	m.focusedPanel = panelCalculator
	m = m.updateFocus()
	m.cost = 10

	m = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.cost != 0 {
		t.Fatalf("expected invalid cost input to reset cost to 0, got %v", m.cost)
	}
}

func TestValidateOpenURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{name: "https", rawURL: "https://example.com", wantErr: false},
		{name: "http", rawURL: "http://example.com/path", wantErr: false},
		{name: "file scheme", rawURL: "file:///etc/passwd", wantErr: true},
		{name: "javascript scheme", rawURL: "javascript:alert(1)", wantErr: true},
		{name: "missing host", rawURL: "https:///missing-host", wantErr: true},
		{name: "relative url", rawURL: "/relative/path", wantErr: true},
		{name: "empty", rawURL: "   ", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateOpenURL(tc.rawURL)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for %q", tc.rawURL)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error for %q, got %v", tc.rawURL, err)
			}
		})
	}
}

func TestStartingNewSearchCancelsPreviousAndIgnoresStaleResults(t *testing.T) {
	provider := &cancelAwareProvider{
		firstStartedCh: make(chan struct{}),
		firstCancelCh:  make(chan struct{}),
	}

	m := newTestModel()
	m.apiClient = api.NewClient(provider)
	m.searchInput.SetValue("first")

	updated1, cmd1 := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m1, ok := updated1.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m, updated1)
	}
	if cmd1 == nil {
		t.Fatal("expected first search command")
	}

	cmdMsg1 := cmd1()
	batch1, ok := cmdMsg1.(tea.BatchMsg)
	if !ok || len(batch1) == 0 {
		t.Fatalf("expected first batch command message, got %T", cmdMsg1)
	}
	searchCmd1 := batch1[len(batch1)-1]
	if searchCmd1 == nil {
		t.Fatal("expected first search command in batch")
	}

	firstSearchResult := make(chan SearchResultsMsg, 1)
	go func() {
		msg, _ := searchCmd1().(SearchResultsMsg)
		firstSearchResult <- msg
	}()

	select {
	case <-provider.firstStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first search request to start")
	}

	m1.searchInput.SetValue("second")
	updated2, cmd2 := m1.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2, ok := updated2.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m1, updated2)
	}
	if cmd2 == nil {
		t.Fatal("expected second search command")
	}
	if m2.searchGen <= m1.searchGen {
		t.Fatalf("expected search generation to advance (%d -> %d)", m1.searchGen, m2.searchGen)
	}

	select {
	case <-provider.firstCancelCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first search request cancellation")
	}

	cmdMsg2 := cmd2()
	batch2, ok := cmdMsg2.(tea.BatchMsg)
	if !ok || len(batch2) == 0 {
		t.Fatalf("expected second batch command message, got %T", cmdMsg2)
	}
	searchCmd2 := batch2[len(batch2)-1]
	if searchCmd2 == nil {
		t.Fatal("expected second search command in batch")
	}

	secondMsg, ok := searchCmd2().(SearchResultsMsg)
	if !ok {
		t.Fatalf("expected second search command to produce SearchResultsMsg")
	}
	if secondMsg.gen != m2.searchGen {
		t.Fatalf("expected second message generation %d, got %d", m2.searchGen, secondMsg.gen)
	}

	updatedDone, _ := m2.Update(secondMsg)
	doneModel, ok := updatedDone.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", m2, updatedDone)
	}
	if doneModel.loading {
		t.Fatal("expected loading=false after second search completes")
	}
	if len(doneModel.results) != 1 {
		t.Fatalf("expected one result from second search, got %d", len(doneModel.results))
	}
	if doneModel.results[0].Title != "second" {
		t.Fatalf("expected second query result title, got %q", doneModel.results[0].Title)
	}

	var staleMsg SearchResultsMsg
	select {
	case staleMsg = <-firstSearchResult:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first search response")
	}
	if staleMsg.gen == 0 {
		t.Fatal("expected stale message to include non-zero generation")
	}

	updatedAfterStale, _ := doneModel.Update(staleMsg)
	afterStale, ok := updatedAfterStale.(Model)
	if !ok {
		t.Fatalf("expected model type %T, got %T", doneModel, updatedAfterStale)
	}
	if afterStale.loading != doneModel.loading || len(afterStale.results) != len(doneModel.results) || afterStale.err != doneModel.err {
		t.Fatal("expected stale search result message to be ignored")
	}
}

func newTestModel() Model {
	m := NewModel()
	m.intro.Show = false
	m.intro.Completed = true
	m.reduceMotion = false
	return m
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
