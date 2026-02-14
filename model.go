package main

import (
	"mrktr/api"
	"mrktr/types"
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
	results       []types.Listing
	selectedIndex int
	resultsOffset int
	stats         types.Statistics

	// Profit calculator
	costInput textinput.Model
	cost      float64

	// History
	history      []string
	historyIndex int

	// State
	loading     bool
	loadingDots int
	spinner     spinner.Model
	err         error
	dataMode    api.SearchMode
	warning     string

	// API
	apiClient *api.Client

	// Animations
	focusFlash  FocusFlash
	reveal      RevealAnim
	statsReveal StatsReveal
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

	return Model{
		keys:         defaultKeyMap(),
		help:         hp,
		intro:        IntroAnimation{Show: true},
		focusedPanel: panelSearch,
		searchInput:  si,
		productIndex: api.NewProductIndex(),
		costInput:    ci,
		spinner:      sp,
		results:      []types.Listing{},
		history:      []string{},
		apiClient:    api.NewEnvClient(),
	}
}

// Init initializes the model (required by tea.Model interface).
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		tea.Tick(40*time.Millisecond, func(time.Time) tea.Msg {
			return introTickMsg{}
		}),
	)
}

// visibleResultRows returns how many result rows fit in the results panel.
func (m Model) visibleResultRows() int {
	resultsHeight := max(4, m.height-layoutOverhead)
	return max(1, resultsHeight-2)
}

// SearchResultsMsg contains search results from the API.
type SearchResultsMsg struct {
	Results []types.Listing
	Mode    api.SearchMode
	Warning string
	Err     error
}

type openURLResultMsg struct {
	Err error
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

type introTickMsg struct{}
