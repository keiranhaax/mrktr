# Statistics Panel Redesign - Implementation Plan (Validated)

## Validation Summary

The original redesign concept is strong, but it is not directly implementable as written in this repository.

### What works
- The data model goals are compatible with existing listing fields (`Platform`, `Condition`, `Status`, `Price`).
- Histogram, market breakdown, percentiles, and skeleton states fit Bubble Tea + Lip Gloss.
- Existing reveal animation patterns can be reused.

### What must change before implementation
1. `mrktr/idea` is a separate Go package path. It cannot directly access unexported symbols from root `package main` (`Model`, styles, helpers).
2. Current stats UI is tightly wired to a fixed 6-line reveal and fixed panel height.
3. Stats tab switching (`1/2/3`) is not currently handled in key bindings or update routing.
4. A standalone `mrktr/idea` prototype with zero root changes will not appear in the running TUI.

## Ground Truth in Current Code
- Stats panel is currently fixed to 6 content lines in `mrktr/view_panels.go:140`.
- Stats reveal tick is hardcoded to 6 lines in `mrktr/update.go:111`.
- Stats panel height is fixed to 6 in `mrktr/view.go:45`.
- Keyboard bindings do not include stats tabs in `mrktr/keys.go:5`.
- `Model` currently stores only `types.Statistics` in `mrktr/model.go:80`.

## Implementation Strategy

Use a **two-layer approach**:
- `mrktr/idea` holds reusable stats-domain logic and rendering helpers (pure and testable).
- Root `mrktr` files integrate the new panel into the existing Bubble Tea model/update/view loop.

This preserves the request for `mrktr/idea` implementation files while still making the live app work.

## Phase Plan

### Phase 0 - Scaffolding and Contracts

Create package and contracts first.

Files to create:
- `mrktr/idea/stats_model.go`
- `mrktr/idea/stats_helpers.go`
- `mrktr/idea/stats_model_test.go`

Deliverables:
- `ExtendedStatistics`
- `PlatformStat`, `ConditionStat`, `HistogramBin`
- `StatsViewMode` enum (`Summary`, `Distribution`, `Market`)
- Pure calculations:
  - percentiles
  - standard deviation / CoV / spread classification
  - status splits (sold vs active)
  - platform and condition aggregates

Gate:
- `cd mrktr && go test ./idea/...`

### Phase 1 - Integrate Extended Stats into Root Model

Files to modify:
- `mrktr/model.go`
- `mrktr/update.go`

Deliverables:
- Add fields to `Model`:
  - `statsViewMode`
  - `extendedStats` (from `mrktr/idea`)
  - stats animation state container
- In `handleSearchResults`, compute extended stats once per result set.
- Keep existing `types.CalculateStats` for calculator compatibility during migration.

Gate:
- `cd mrktr && go test ./...`

### Phase 2 - Tab System and Key Handling

Files to modify:
- `mrktr/keys.go`
- `mrktr/update.go`
- `mrktr/update_test.go`

Deliverables:
- Add stats-tab key bindings (`1`, `2`, `3`).
- Route keys only when `panelStats` is focused.
- On tab change:
  - restart stats reveal animation
  - keep global panel focus unchanged
- Update help text to advertise stats tab shortcuts.

Gate:
- New tests:
  - tab keys ignored outside stats panel
  - tab keys switch mode inside stats panel
  - mode switch resets stats reveal state

### Phase 3 - Summary View (Tab 1) First

Files to create:
- `mrktr/idea/stats_summary.go`
- `mrktr/idea/stats_styles.go`

Files to modify:
- `mrktr/view_panels.go`

Deliverables:
- Render enhanced summary with:
  - result count
  - spread indicator
  - sparkline
  - min/max/avg/median
  - P25/P75
- Keep current appearance fallback for narrow widths.
- Replace hardcoded 6-line assumptions with per-view line count.

Gate:
- `go test ./...`
- Manual run with widths 64, 80, 120

### Phase 4 - Distribution View (Tab 2)

Files to create:
- `mrktr/idea/stats_distribution.go`

Files to modify:
- `mrktr/idea/stats_helpers.go`
- `mrktr/view_panels.go`

Deliverables:
- Auto-bin histogram (Sturges rule with sane min/max bin clamps).
- Unicode bars with gradient fill.
- Deal marker on bins that include sub-P25 prices.
- Percentile band line below histogram.

Gate:
- Unit tests for histogram edge cases:
  - 0 listings
  - 1 listing
  - identical prices
  - high outliers

### Phase 5 - Market View (Tab 3)

Files to create:
- `mrktr/idea/stats_market.go`

Files to modify:
- `mrktr/view_panels.go`

Deliverables:
- Platform average bars with counts.
- Sold vs active row with averages.
- Sell-through percentage bar.
- Stable ordering (count desc, then name) for deterministic UI/tests.

Gate:
- Unit tests for platform/status aggregation and ordering.

### Phase 6 - Animation Pass

Files to create:
- `mrktr/idea/stats_animation.go`

Files to modify:
- `mrktr/model.go`
- `mrktr/update.go`
- `mrktr/view_panels.go`

Deliverables:
- Skeleton shimmer when `loading && len(results)==0`.
- Tab-transition reveal (line-by-line) on mode switch.
- Rolling number interpolation for summary metrics.
- Delta indicators with short fade.

Notes:
- Start at 30 FPS tick (`~33ms`) for lower terminal churn; tune to 16ms only if smooth and stable.

Gate:
- Animation message-generation checks in `update_test.go`.

### Phase 7 - Layout and Final Hardening

Files to modify:
- `mrktr/view.go`
- `mrktr/view_test.go`
- `mrktr/styles.go` (if shared style tokens need extension)

Deliverables:
- Increase stats panel height from fixed `6` to dynamic target suitable for tab content.
- Preserve minimum terminal constraints and avoid overlap with calculator/history.
- Ensure narrow-width fallback remains readable.

Gate:
- `go test ./...`
- Manual resize and navigation verification.

## Test Plan (Required)

Automated:
- `cd mrktr && go test ./...`
- New unit tests in `mrktr/idea` for percentile/bin/spread/aggregation math.
- Update tests for changed stats reveal behavior.

Manual:
- Search flows with real providers and no providers.
- Keyboard flow: focus stats panel, switch `1/2/3`, switch back.
- Terminal sizes: 64x14, 80x24, 120x40.
- Edge datasets: empty, single listing, uniform prices, mixed sold/active, single platform.

## Risks and Mitigations

- Risk: Package boundary friction (`main` vs `mrktr/idea`).
  - Mitigation: keep `mrktr/idea` pure and importable; integration remains in root.

- Risk: Over-animating causes flicker/high CPU.
  - Mitigation: animation toggles and conservative tick rates first.

- Risk: Layout regressions in small terminals.
  - Mitigation: compact fallback rendering and explicit width guards.

## First Implementation Slice (Start Here)

1. Implement `ExtendedStatistics` + pure calculators in `mrktr/idea`.
2. Wire `extendedStats` into `Model` and `handleSearchResults`.
3. Add stats tab mode state and `1/2/3` key handling.
4. Ship Tab 1 (Summary) only behind full integration.

After that, Tab 2/3 and advanced animations can land incrementally without breaking existing behavior.
