# Copilot Instructions

## Build & Run

```bash
go build -o mrktr .        # Build binary (not committed)
go run .                    # Run in development
go test ./...               # Run all tests
go test ./types/...         # Run tests for a specific package
go test -run TestCalc ./... # Run a single test by name
go vet ./...                # Static analysis
air                         # Live reload (requires: go install github.com/air-verse/air@latest)
```

## Architecture

This is a **Bubble Tea** TUI app following the **Elm Architecture (Model-View-Update)** pattern for comparing reseller prices across marketplaces (eBay, Mercari, Amazon, Facebook).

### Core MVU files (all in `main` package)

- `model.go` — Single source of truth for app state, focus flags, and custom message types (`SearchResultsMsg`)
- `update.go` — Processes all messages: keyboard input, window resize, API responses. Side effects (HTTP calls, opening URLs) are initiated here via `tea.Cmd`
- `view.go` — Pure rendering function. Each panel has its own `render*Panel()` method that returns a string
- `styles.go` — All Lip Gloss styles and the `renderPanel()` helper that handles active/inactive border states

### Panel focus system

Five panels cycle via Tab/Shift+Tab using an `iota` enum (`panelSearch` through `panelHistory`). The `updateFocus()` method on Model manages which `textinput` has focus. Only `panelSearch` and `panelCalculator` have text inputs; the others respond to j/k navigation.

### API fallback chain (`api.go`)

1. Firecrawl API (`FIRECRAWL_API_KEY`) — primary
2. Tavily API (`TAVILY_API_KEY`) — fallback
3. Mock data — demo mode when no keys are set

Search results are parsed with regex price extraction (`$X,XXX.XX` pattern), URL-based platform detection, and keyword-based condition/status inference. New API providers should follow this same pattern: isolated function, gated on env var, returning `[]types.Listing`.

### Types package (`types/listing.go`)

Shared types (`Listing`, `Statistics`, `ProfitCalculation`) and pure computation functions (`CalculateStats`, `CalculateProfit`). Extend types here before wiring into the model.

## Conventions

- Format with `gofmt`. Tabs, standard Go ordering.
- Message types use `Msg` suffix per Bubble Tea convention (e.g., `SearchResultsMsg`).
- Keep views pure — no side effects in render functions.
- Parsing and formatting logic should be pure helper functions; side effects stay near `update.go` and `api.go`.
- Panel rendering follows the pattern: `render*Panel(width, height int) string` calling `renderPanel(title, content, width, height, active)`.
- Prefer table-driven tests for parsers, stats, and platform detection.
- Commits: Conventional Commits format (e.g., `feat: add mercari parser`, `fix: handle missing prices`).

## Environment Variables

```bash
FIRECRAWL_API_KEY  # Primary search API
TAVILY_API_KEY     # Fallback search API
```

Without these, the app runs in demo mode with mock data — useful for deterministic UI testing.
