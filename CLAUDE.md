# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

All commands run from the `mrktr/` directory:

```bash
go run .                         # Run in development
go build -o mrktr .              # Build binary (gitignored)
go test ./...                    # Run all tests
go test ./types/...              # Run tests for a specific package
go test -run TestCalc ./...      # Run a single test by name
go test -coverprofile=coverage.out ./...  # Coverage report
go vet ./...                     # Static analysis
go fmt ./...                     # Format code
air                              # Live reload (requires: go install github.com/air-verse/air@latest)
```

## Architecture

Bubble Tea TUI app following the **Elm Architecture (Model-View-Update)** for comparing reseller prices across eBay, Mercari, Amazon, and Facebook Marketplace.

### Core MVU files (all in `main` package)

- **`model.go`** — Application state, focus flags, custom message types (`SearchResultsMsg`). `NewModel()` initializes text inputs and defaults.
- **`update.go`** — Processes all messages: keyboard input, window resize, API responses. Side effects (HTTP calls, `openURL`) are initiated here via `tea.Cmd`.
- **`view.go`** — Pure rendering. Each panel has a `render*Panel(width, height int) string` method calling the shared `renderPanel()` helper. Layout is two-column (2/3 left, 1/3 right) joined with `lipgloss.JoinHorizontal/Vertical`.
- **`styles.go`** — All Lip Gloss styles and the `renderPanel(title, content, width, height, active)` helper that handles active/inactive border states.
- **`api/`** — Search providers, fallback chain, parser, and local query suggestion index.
- **`main.go`** — Entry point. Creates `tea.NewProgram` with alt screen and mouse support.

### Panel focus system

Five panels cycle via Tab/Shift+Tab using an `iota` enum (`panelSearch` through `panelHistory`). `updateFocus()` manages which `textinput.Model` has focus. Only `panelSearch` and `panelCalculator` accept text input; others respond to j/k navigation. In the search panel, Tab first accepts inline suggestions when present.

### API fallback chain (`api/search.go`)

1. Brave Search API (`BRAVE_API_KEY`) — primary
2. Tavily API (`TAVILY_API_KEY`) — secondary provider
3. Firecrawl (`FIRECRAWL_API_KEY`) — tertiary provider

`parseSearchResults()` uses regex price extraction (`$X,XXX.XX`), URL-based platform detection, and keyword-based condition/status inference. New API providers should follow the same pattern: isolated function, gated on env var, returning `[]types.Listing`.

### Types package (`types/listing.go`)

Shared domain types and pure computation functions:
- `Listing` — platform, price, condition, status, URL, title
- `Statistics` — count, min, max, average, median (via `CalculateStats`)
- `ProfitCalculation` — cost, sell price, profit, percent (via `CalculateProfit`)

Extend types here before wiring into the model.

## Project Status

API providers now live in `api/` (`brave.go`, `tavily.go`, `firecrawl.go`) with shared client logic in `api/search.go`. Query suggestions/expansion live in `api/suggest.go` with embedded product data.

## Conventions

- Format with `gofmt`. Tabs, standard Go ordering.
- Message types use `Msg` suffix per Bubble Tea convention (e.g., `SearchResultsMsg`).
- Keep views pure — no side effects in render functions.
- Panel rendering follows the pattern: `render*Panel(width, height int) string`.
- Prefer table-driven tests for parsers, stats, and platform detection.
- Commits: Conventional Commits format (e.g., `feat: add mercari parser`, `fix: handle missing prices`).

## Environment Variables

```bash
BRAVE_API_KEY      # Primary search API
TAVILY_API_KEY     # Secondary search API
FIRECRAWL_API_KEY  # Tertiary search API
```

The app auto-loads `.env` from the `mrktr/` directory before creating the model.
