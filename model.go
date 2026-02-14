package main

import (
	"mrktr/types"

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

// Model represents the application state
type Model struct {
	// Terminal dimensions
	width  int
	height int

	// Focus management
	focusedPanel int

	// Search
	searchInput textinput.Model

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
	dataMode    searchMode
	warning     string

	// Animation: focus flash
	focusFlashTicks  int
	focusFlashActive bool
	focusFlashGen    int

	// Animation: results reveal
	revealedRows int
	revealing    bool
	revealGen    int

	// Animation: stats reveal
	statsRevealed  int
	statsRevealGen int
}

// NewModel creates a new application model with initial state
func NewModel() Model {
	// Initialize search input
	si := textinput.New()
	si.Placeholder = "Enter item to search..."
	si.Focus()
	si.CharLimit = 100
	si.Width = 30

	// Initialize cost input
	ci := textinput.New()
	ci.Placeholder = "0.00"
	ci.CharLimit = 10
	ci.Width = 10

	// Initialize loading spinner
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = spinnerStyle

	return Model{
		focusedPanel: panelSearch,
		searchInput:  si,
		costInput:    ci,
		spinner:      sp,
		results:      []types.Listing{},
		history:      []string{},
		dataMode:     searchModeDemo,
	}
}

// Init initializes the model (required by tea.Model interface)
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// visibleResultRows returns how many result rows fit in the results panel.
func (m Model) visibleResultRows() int {
	resultsHeight := max(4, m.height-layoutOverhead)
	return max(1, resultsHeight-2)
}

// SearchResultsMsg contains search results from the API
type SearchResultsMsg struct {
	Results []types.Listing
	Mode    searchMode
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
