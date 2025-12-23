# Drawing Tools Phased Rollout Plan

Last updated: 2025-12-23

Note: A git checkpoint (status/log/diff) is run before each step. This plan is updated after each step to track progress or adjustments.

## Phase 1 (Core Toolset)

### Step 0 — Repo sanity + baseline inventory
Status: completed
Git checkpoint: `git status -sb`, `git log -1`, `git diff --stat`
Scope:
- Confirm chart container and LWC integration points.
- Identify existing drawing/indicator codepaths.
- Note backend storage conventions.
Notes:
- Repo has no commits yet; git checkpoints use explicit git path.

### Step 1 — Drawing architecture + data model
Status: completed
Git checkpoint before changes.
Scope:
- Define `DrawingShape` + `ToolDefinition` + tool registry.
- Add drawing state machine (idle → drawing → selected → editing → moving → resizing).
- Add LWC coordinate helpers (time/price conversions).
- Update `docs/development/DRAWING_TOOLS_ARCHITECTURE.md` (schema, lifecycle, rendering/hit-test contracts).

### Step 2 — Rendering layer (core tools)
Status: completed
Git checkpoint before changes.
Core tools:
- Trend Line, Ray, Horizontal Line, Vertical Line, Rectangle, Text, Arrow, Measure
Scope:
- Implement drawing layer on pane 0 using LWC primitives.
- Per-shape renderers + hit tests.
- Selection overlay + resize handles.

### Step 3 — Interaction engine (core tools)
Status: completed
Git checkpoint before changes.
Scope:
- Pointer handling: create/preview/finalize, select/move/resize.
- Delete/undo/redo.
- Disable pan/zoom while drawing.
- Throttle updates with `requestAnimationFrame`.
Notes:
- Optional debug seed: set `localStorage.isx_drawings_seed = "1"` then reload to pre-populate sample shapes for testing selection/move/resize/undo.

### Step 3.1 — Chart parity fixes (volume + tooltip)
Status: completed
Git checkpoint: `git status -sb`, `git log -1`, `git diff --stat`
Scope:
- Fix volume bars not rendering (intensity coloring now supports `rgba(...)` theme colors; volume MA uses `chartData.volumeData`).
- Restore analysis tooltip overlay (crosshair subscription now waits for `chartReady`).
- Add a small Jest smoke test for chart data → volume + tooltip data generation.
Notes:
- Console errors from `chrome-extension://...` are browser extensions; disable extensions while debugging app issues.

### Step 3.2 — Chart UX defaults (timeframe + volume height)
Status: completed
Git checkpoint: `git status -sb`, `git log -1`, `git diff --stat`
Scope:
- Default analysis timeframe is now `3M` (instead of `1D`).
- Volume pane initial height is now 20% of the chart container.

### Step 4 — Toolbar UI (Phase 1)
Status: completed
Git checkpoint before changes.
Scope:
- Vertical toolbar with icons, tooltips, groupings, separators.
- Actions: select tool, delete, lock/unlock, visibility, undo/redo.
- Persist last tool in `localStorage`.
Notes:
- Visibility toggle hides/shows the entire drawing layer (to allow un-hiding without an objects list UI).
- Dist verified via `build-dist.cmd` + `/api/health` + `/analysis` (200).

### Step 4.1 — Toolbar parity + usability fixes
Status: completed
Git checkpoint: `git status -sb`, `git log -1`, `git diff --stat`
Scope:
- Toolbar restyled to more closely match TradingView’s compact left rail.
- Text tool: visible on dark charts and prompts for content on placement; double-click text to edit.
- Delete/Lock: operate on selected drawing (or hovered drawing if nothing selected).
- Magnet: optional snap-to-candle (nearest OHLC) when placing/moving points.
Notes:
- The console errors referencing `chrome-extension://.../contentScript.js` are from a browser extension, not ISX Pulse.
- Fix: toolbar clicks no longer trigger the Text prompt (SVG targets no longer bypass the toolbar guard).
- Chart now follows the app theme (fixed theme override bug); drawings sync text colors accordingly.
- Theme toggle uses app-level selector (top-right); chart now reads `next-themes` `theme` and updates via `applyOptions`.
- Theme fallback: chart theme now also follows `html.dark` class to avoid mismatches where the UI is light but `next-themes` still reports dark.

### Step 5 — Persistence (backend + frontend)
Status: completed
Git checkpoint before changes.
Storage:
- Backend: `data/indicators/{ticker}.json`
Scope:
- Go API: `GET/PUT /api/v1/drawings/{ticker}/` (validation + per-ticker locking; atomic writes).
- Frontend auto-load on ticker change; debounce save (~1s) when drawings become dirty.
Notes:
- Dist writes to `dist/data/indicators/{TICKER}.json` (data dir is resolved from executable directory).
- Smoke test: `scripts/windows/check-drawings-persistence.bat` (expects server on `http://localhost:8080`).

### Step 6 — Indicators audit + UI wiring
Status: completed
Git checkpoint before changes.
Scope:
- Audit indicator components + their setting keys (fixed mismatches so settings actually affect calculations).
- Align Indicators panel categories to TradingView-style groups (Trend, Momentum, Volatility, Volume, Levels).
- Add settings editing (gear) for common indicators; persist both toggles and parameter settings to `localStorage`.
Notes:
- Indicator parameter settings are now shared across the app (Zustand + persistence) instead of per-component `useState`.
- Analysis chart preferences now persist (indicator toggles + timeframe/chart type + activation order).
- Fix: Indicators panel now controls the same state instance as the chart (previously it used a separate `useChartState()` instance, so toggles had no effect).
- UX: Indicators panel moved to a left-side collapsible sidebar on `/analysis` (toggle via the `Indicators` button or the chevron handle).
- UX: Added TradingView-style `Tickers` + `Indicators` toggle buttons into the chart controls row (left of `Candlestick`), matching the same button style.

### Step 6.1 — Sidebar polish (ticker animation + preserve zoom)
Status: completed
Git checkpoint: `git status -sb`, `git log -1`, `git diff --stat`
Scope:
- Ticker panel now collapses/expands with the same smooth width animation as the Indicators panel (single `ResizablePanel` stays mounted; width is controlled).
- Adding/removing indicators no longer resets the chart's time zoom (removed `fitContent()` calls tied to indicator toggles; VWAP overlay no longer forces `fitContent()` on updates).
Notes:
- Pane transitions still preserve the visible range via `usePaneTransitionManager` (fallback `fitContent()` only on restoration failure).

### Step 7 — Tests + performance
Status: completed
Git checkpoint before changes.
Scope:
- Unit tests for serialization + hit testing.
- API tests for save/load.
- Performance safeguards (minimal invalidation, RAF throttling).
Notes:
- Added unit tests for drawings envelope normalization + hit testing.
- Added backend tests for drawings save/load (including backward-compatible array payloads).
- Throttled drawing layer primitive redraws to one frame during drag operations.

### Step 8 — Documentation + changelog
Status: completed
Git checkpoint before changes.
Scope:
- Update `DRAWING_TOOLS_ARCHITECTURE.md`, `ChartsBasicsSection.tsx`, `CHANGELOG.md`.
- Add rollout notes + feature flag if needed.
Notes:
- Added a Guide section for charts at `/guide?section=charts` (linked from the Analysis page help button).
- Updated architecture doc with persistence and performance notes.

## Phase 2 (Advanced Tools)
Status: in_progress
Git checkpoint before changes.
Scope:
- Add Parallel Channel, Brush, Fibonacci Retracement/Extension, Pitchfork, Ellipse.
- Extend hit testing/editing; update docs/tests/UI accordingly.

### Step 9 — Parallel Channel
Status: completed
Git checkpoint: `git status -sb`, `git log -1`, `git diff --stat`
Scope:
- Add 3-point Parallel Channel tool (base line + offset line + fill).
- Multi-click draft placement for N-point tools.
- Hit testing: lines + interior; resize via 3 handles.
- Backend validation updated to allow `parallelChannel`.
Tests:
- Jest: drawings hit testing + envelope normalization
- Go: drawings handler round-trip + validation
Notes:
- Placement: click 1st point, click 2nd point (base), click 3rd point (channel width).
