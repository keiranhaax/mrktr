package main

import (
	"context"
	"mrktr/api"
	"mrktr/idea"
	"mrktr/types"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Panel focus states
const (
	panelSearch = iota
	panelResults
	panelStats
	panelCalculator
	panelHistory
)

const layoutOverhead = 14

// IntroAnimation groups intro animation state.
type IntroAnimation struct {
	Show      bool
	Completed bool
	Tick      int
	Phase     int // 0=reveal letters, 1=glow sweep, 2=fade out
}

// FocusFlash groups focus highlight animation state.
type FocusFlash struct {
	Ticks  int
	Gen    int
	Active bool
}

// RevealAnim groups results reveal animation state.
type RevealAnim struct {
	Rows      int
	Gen       int
	Revealing bool
}

// StatsReveal groups statistics reveal animation state.
type StatsReveal struct {
	Revealed int
	Gen      int
}

// StatsPanelAnimation groups future stats-panel animation state.
type StatsPanelAnimation struct {
	SkeletonFrame int
	ValueTweenGen int
	ValueTweenOn  bool
	ValueStep     int
	ValueSteps    int
	DeltaTicks    int
	DeltaTotal    int
	FromStats     idea.ExtendedStatistics
	ToStats       idea.ExtendedStatistics
	DeltaMin      float64
	DeltaMax      float64
	DeltaAvg      float64
	DeltaMedian   float64
	DeltaP25      float64
	DeltaP75      float64
}

// Model represents the application state.
type Model struct {
	// Terminal dimensions
	width  int
	height int

	// Shared components
	keys keyMap
	help help.Model

	// Intro animation
	intro IntroAnimation

	// Focus management
	focusedPanel int

	// Search
	searchInput textinput.Model

	// Query enhancement
	productIndex *api.ProductIndex

	// Results
	rawResults      []types.Listing
	results         []types.Listing
	selectedIndex   int
	resultsOffset   int
	sortField       types.SortField
	sortDirection   types.SortDirection
	resultFilter    types.ResultFilter
	filterBarActive bool
	detailOpen      bool
	stats           types.Statistics
	extendedStats   idea.ExtendedStatistics
	statsViewMode   idea.StatsViewMode

	// Profit calculator
	costInput    textinput.Model
	cost         float64
	calcPlatform string

	// History
	history      []string
	historyIndex int
	historyMeta  map[string]HistoryEntry
	historyStore HistoryStore

	// State
	loading        bool
	loadingDots    int
	reduceMotion   bool
	spinner        spinner.Model
	err            error
	dataMode       api.SearchMode
	warning        string
	statusFlash    string
	statusFlashGen int
	lastQuery      string

	// API
	apiClient *api.Client

	// Search cancellation and stale-response protection
	searchCancel context.CancelFunc
	searchGen    int

	// Animations
	focusFlash  FocusFlash
	reveal      RevealAnim
	statsReveal StatsReveal
	statsAnim   StatsPanelAnimation
}

// NewModel creates a new application model with initial state.
func NewModel() Model {
	// Initialize search input
	si := textinput.New()
	si.Focus()
	si.CharLimit = 100
	si.Width = 30
	si.ShowSuggestions = true

	// Initialize cost input
	ci := textinput.New()
	ci.CharLimit = 10
	ci.Width = 10

	// Initialize loading spinner
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = spinnerStyle

	hp := help.New()
	hp.ShortSeparator = "  "
	hp.FullSeparator = "   "
	hp.Styles.ShortKey = keyStyle
	hp.Styles.ShortDesc = keyDescStyle
	hp.Styles.ShortSeparator = separatorStyle
	hp.Styles.Ellipsis = separatorStyle
	hp.Styles.FullKey = keyStyle
	hp.Styles.FullDesc = keyDescStyle
	hp.Styles.FullSeparator = separatorStyle

	var historyStore HistoryStore
	startupWarning := ""
	if store, err := NewFileHistoryStore(); err != nil {
		startupWarning = "History persistence unavailable."
	} else {
		historyStore = store
	}

	return Model{
		keys:          defaultKeyMap(),
		help:          hp,
		intro:         IntroAnimation{Show: true},
		focusedPanel:  panelSearch,
		searchInput:   si,
		productIndex:  api.NewProductIndex(),
		costInput:     ci,
		spinner:       sp,
		rawResults:    []types.Listing{},
		results:       []types.Listing{},
		sortField:     types.SortFieldPrice,
		sortDirection: types.SortDirectionAsc,
		resultFilter:  types.ResultFilter{},
		calcPlatform:  "eBay",
		statsViewMode: idea.StatsViewSummary,
		extendedStats: idea.CalculateExtendedStats(nil),
		reduceMotion:  shouldReduceMotionFromEnv(),
		history:       []string{},
		historyMeta:   map[string]HistoryEntry{},
		historyStore:  historyStore,
		apiClient:     api.NewEnvClient(),
		warning:       startupWarning,
	}
}

// Init initializes the model (required by tea.Model interface).
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		loadHistoryCmd(m.historyStore),
		tea.Tick(40*time.Millisecond, func(time.Time) tea.Msg {
			return introTickMsg{}
		}),
	)
}

// visibleResultRows returns how many result rows fit in the results panel.
func (m Model) visibleResultRows() int {
	overhead := layoutOverhead
	if m.width > 0 && m.width < 80 {
		overhead += 8
	}
	resultsHeight := max(4, m.height-overhead)
	return max(1, resultsHeight-2)
}

// SearchResultsMsg contains search results from the API.
type SearchResultsMsg struct {
	Results        []types.Listing
	Mode           api.SearchMode
	Warning        string
	Err            error
	ProviderErrors []api.ProviderError
	gen            int
}

type openURLResultMsg struct {
	Err error
}

type clipboardResultMsg struct {
	Err error
}

type exportResultMsg struct {
	Path string
	Err  error
}

type historyLoadedMsg struct {
	Entries []HistoryEntry
	Err     error
}

type historySavedMsg struct {
	Err error
}

type statusFlashClearMsg struct {
	gen int
}

type focusFlashTickMsg struct {
	gen int
}

type revealRowTickMsg struct {
	gen int
}

type statsRevealTickMsg struct {
	gen int
}

type statsValueTickMsg struct {
	gen int
}

type introTickMsg struct{}

func shouldReduceMotionFromEnv() bool {
	return parseBoolishEnv(os.Getenv("MRKTR_LOW_POWER")) ||
		parseBoolishEnv(os.Getenv("MRKTR_REDUCE_MOTION"))
}

func parseBoolishEnv(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
