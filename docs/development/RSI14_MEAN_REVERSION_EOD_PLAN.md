# RSI(14) Mean Reversion (EOD) — Development Plan + User Stories

## Summary
We will implement a **real-data** RSI(14) mean-reversion strategy that can be **manually triggered** and runs **across all tickers**. Results will be displayed in the **Strategies** page, including an **Alerts panel inside the strategy page** (the standalone Alerts page remains removed).

**Core rules (RSI(14), long-only baseline):**
- **Entry (BUY):** RSI(14) crosses **below 30**
- **Exit (SELL):** RSI(14) crosses **above 50**

**Optional variant (configurable later):**
- Exit at RSI(14) crosses **above 70** (stronger/rarer exit), or “exit=50 and take-profit=70” depending on chosen execution model.

No mock data is allowed anywhere in the strategy path.

---

## Current Implementation Status (as of now)

### Frontend (`web/`)
- Large parts of the old **Analysis/Indicators UI and charting system** are now staged for deletion (`web/app/analysis/*`, `web/components/analysis/*`, `web/lib/indicators/*`, chart hooks/utils, and `web/types/analysis.ts`).
- TypeScript still fails because **many references remain**:
  - Navigation still includes `/analysis` in `web/components/app-content-client.tsx`.
  - Market overview still links to `/analysis?ticker=...` in `web/app/market-overview/market-overview-client.tsx`.
  - Guide content still describes analysis/indicators and references `/analysis` and indicator demos.
  - Operations configs/types still list `analysis`/`indicators` stages (frontend must remove those stage IDs entirely).
- Strategies page exists as a placeholder: `web/app/strategy/page.tsx`, `web/app/strategy/strategy-client.tsx`.
- Alerts route/page is already **not present** (`web/app/alerts` does not exist).

### Backend (`api/`)
- A strategy subsystem already exists (`api/internal/strategy/*`) including:
  - Strategy types (`Signal`, `BacktestResult`, etc.)
  - A WebSocket broadcaster for `strategy_status`, `strategy_signal`, `strategy_backtest`
  - Existing strategies including a `mean_reversion_v1` that uses RSI internally.
- **But** `api/internal/services/strategy_service.go` currently uses **mock data** (`getMockData`) for `ExecuteStrategy`/`Backtest`. This must be removed and replaced with real pipeline data access.

---

## Goals / Non-Goals

### Goals
1. Remove all TA/indicators pre-calculation **stages and UI** (including `/analysis` route) so the app compiles and reflects the new direction.
2. Implement **RSI(14) Mean Reversion (EOD)** using **real data** and run it across **all tickers** on-demand.
3. Persist strategy outputs (signals + summary metrics) to disk (and later optionally DB).
4. Display results in the Strategies page:
   - Strategy list
   - Last run status
   - Signals table
   - “Alerts” panel (inside Strategies)
5. Add backtesting per ticker + aggregated results, also using real data.

### Non-Goals (for this milestone)
- A standalone Alerts page (explicitly removed for now).
- Intraday execution (this is **EOD**).
- Auto-scheduling strategies (manual trigger first).
- Advanced portfolio sizing/risk engine (start with simple fixed sizing and/or “signal only” mode).

---

## Architecture Decision (Recommended)

### A) Use the existing Strategy subsystem (recommended)
- Keep strategy execution under `api/internal/strategy` + `api/internal/services/strategy_service.go`.
- Add a **batch-run endpoint** that runs a strategy across all tickers, emits progress via existing **strategy WebSocket events**, and persists output to `data/strategies/...`.

This matches the repo’s current backend direction (strategy endpoints already exist) and avoids coupling with Operations pipeline UI prematurely.

### B) Alternative: treat Strategy as an Operations stage
- Add a new operation type/stage (e.g. `strategy_rsi14_mr_eod`) under operations.
- Reuse the Operations UI and telemetry.

This is feasible later, but higher complexity for the first clean strategy milestone.

---

## Data Contract (Proposed)

### Strategy Run identity
- `strategy_id`: `rsi14_mr_eod_v1`
- `run_id`: UUID (generated per run)
- `as_of_date`: last trading day included in computation

### Persisted outputs (disk)
Create a run folder:
- `data/strategies/rsi14_mr_eod_v1/<run_id>/`
  - `signals.json` (all tickers, latest EOD)
  - `signals.csv` (same data for Excel)
  - `summary.json` (counts, win-rate placeholders, etc.)
  - `backtest/` (optional later; per-ticker results)

### Signal (minimum fields)
- `strategy_id`, `run_id`, `symbol`, `timestamp`, `action` (`BUY|SELL|HOLD`), `strength`
- `price` (EOD close), `reasoning`
- `metadata`: at least `{ rsi14, entry_threshold=30, exit_threshold=50 }`

---

## Milestones (Revised End-to-End Plan)

### Milestone 0 — Cleanup & Compile (remove TA/Indicators)
**Outcome:** TS type-check passes; no Analysis/Indicators pages; strategy direction is the only “analysis-like” UI.

Work items:
1. Remove `/analysis` navigation and all `/analysis` references in UI.
2. Remove guide sections/demos that refer to indicators/charting/analysis and delete the indicator demo files.
3. Remove `analysis` and `indicators` from frontend operations config (`web/lib/api/operation-request-builder.ts`) and from stage model (`web/lib/operations/*`, `web/types/index.ts`).
4. Ensure `npm run -s type-check` passes for the web app.

Deliverables:
- No `/analysis` route and no references.
- No indicator/TA pre-calc code paths.

---

### Milestone 1 — Real Data Strategy Data Access (NO mock data)
**Outcome:** Strategy service reads real ticker history from the pipeline outputs (processed CSVs / stored datasets) and never generates fake prices.

Work items:
1. Replace `getMockData()` usage in `api/internal/services/strategy_service.go` with real data retrieval.
2. Decide the canonical source for OHLCV:
   - Preferred: use the same dataset produced by the pipeline (post-processing).
3. Implement a `StrategyDataProvider`:
   - `ListSymbols() []string`
   - `GetSymbolSeries(symbol, limit)` for execute
   - `GetSymbolSeriesRange(symbol, start, end)` for backtest
4. Ensure errors are explicit:
   - missing file vs insufficient data vs parse failure.

Deliverables:
- Strategy execution for a single symbol returns real EOD-based signal.

---

### Milestone 2 — Implement RSI(14) Mean Reversion Strategy (EOD)
**Outcome:** New strategy id exists with the exact RSI rules agreed above.

Work items:
1. Create a dedicated strategy implementation, e.g.:
   - `api/internal/strategy/strategies/rsi14_mr_eod.go`
   - Strategy ID: `rsi14_mr_eod_v1`
2. Parameters (v1):
   - `rsi_period = 14` (fixed)
   - `entry_threshold = 30` (fixed for v1)
   - `exit_threshold = 50` (fixed for v1)
   - optionally: `allow_short = false` (fixed for v1)
3. Signal rules:
   - BUY when RSI crosses below 30 (today RSI < 30 and yesterday RSI >= 30)
   - SELL when RSI crosses above 50 (today RSI > 50 and yesterday RSI <= 50)
   - HOLD otherwise
4. Metadata include RSI value and thresholds.

Deliverables:
- Unit tests for RSI calculation and crossing logic using small deterministic series (not mock market data; just math inputs).

---

### Milestone 3 — Batch Run Across All Tickers (manual trigger)
**Outcome:** The strategy runs “as a stage” over all tickers, with progress + persisted results.

Work items:
1. Add endpoint:
   - `POST /api/v1/strategies/{strategyID}/execute-batch`
   - Body: `{ "as_of_date"?: "...", "data_points"?: number, "params"?: {...} }`
2. Implementation steps (batch “stages”):
   - Stage A: load symbol list
   - Stage B: iterate symbols, compute RSI + signal (EOD)
   - Stage C: persist `signals.json/csv` + `summary.json`
3. WebSocket updates (existing mechanism):
   - `strategy_status`: stage transitions + counts
   - `strategy_signal`: optionally stream signals as computed
   - final `strategy_status`: completed
4. Add simple concurrency guard (avoid two runs overlapping for the same strategy unless explicitly allowed).

Deliverables:
- One run produces a folder under `data/strategies/...` with signals for all symbols.

---

### Milestone 4 — Backtesting (per ticker + aggregated)
**Outcome:** backtest the RSI rules over a date range using real data.

Work items:
1. Add endpoint:
   - `POST /api/v1/strategies/{strategyID}/backtest-batch`
   - Body: `{ "start_date": "...", "end_date": "...", "initial_cash": ..., "commission": ..., "slippage": ... }`
2. Backtest logic:
   - Start with a simple position model (long-only, 1 position max per ticker).
   - Use entry/exit cross rules.
3. Persist:
   - `data/strategies/<id>/<run_id>/backtest/summary.json`
   - `.../backtest/by_ticker/<symbol>.json`
4. WebSocket:
   - stream progress and publish the final aggregated summary.

Deliverables:
- Backtest results visible in UI (summary + per ticker table).

---

### Milestone 5 — Frontend Strategies Page (new strategies only)
**Outcome:** `/strategy` becomes the canonical place to run and view strategies and “alerts”.

Work items:
1. Add strategies list UI (starting with RSI14 MR EOD).
2. Add “Run now” and “Backtest” actions.
3. Add results panels:
   - Last run metadata (run_id, as_of_date, count of BUY/SELL/HOLD)
   - Signals table with filters (ticker/action)
   - Alerts panel (just signals that are actionable, e.g. BUY/SELL)
4. Subscribe to strategy WebSocket messages and update the UI live.
5. Link Market Overview to `/strategy?ticker=...` (since `/analysis` is gone).

Deliverables:
- A trader can run RSI14 MR EOD and see results without leaving `/strategy`.

---

## User Stories (with Acceptance Criteria)

### Epic: Cleanup (Remove TA/Indicators)
**US-0.1** — Remove Analysis navigation
- As a user, I do not see an “Analysis” page in the navigation.
- Acceptance:
  - No nav item links to `/analysis`.
  - `rg "/analysis" web/` only matches historical docs (or returns none).

**US-0.2** — Strategy page is the primary “signals” UI
- As a user, when I click a ticker from Market Overview, I land on `/strategy?ticker=<symbol>`.
- Acceptance:
  - Market Overview click behavior uses `/strategy`.

---

### Epic: RSI Strategy (Real Data)
**US-1.1** — Run RSI strategy across all tickers
- As a trader, I can click “Run RSI(14) Mean Reversion (EOD)” and generate signals for all tickers using real data.
- Acceptance:
  - API returns a `run_id`.
  - WebSocket shows progress and completion.
  - Output files are created under `data/strategies/rsi14_mr_eod_v1/<run_id>/`.

**US-1.2** — View latest signals and “alerts”
- As a trader, I can view today’s BUY/SELL signals in one place (Alerts panel inside Strategies).
- Acceptance:
  - Alerts panel shows actionable signals only (BUY/SELL).
  - Table can filter by ticker and action.

**US-1.3** — Verify signal logic (crossing)
- As a trader, I only get a BUY when RSI crosses below 30 (not every day RSI stays below 30).
- Acceptance:
  - For a known RSI sequence, the backend produces BUY only on the cross day.

---

### Epic: Backtesting
**US-2.1** — Backtest RSI strategy per ticker
- As a trader, I can backtest RSI14 MR over a date range and see per-ticker results.
- Acceptance:
  - Backtest endpoint returns results for all tickers (or top N) with totals.
  - UI displays aggregated summary + sortable per-ticker table.

**US-2.2** — Reproducible results
- As a trader, running the same backtest date range yields the same results.
- Acceptance:
  - No randomness in backtest.
  - Same input => same output files.

---

## Open Decisions (need confirmation, but can default)
1. Long-only vs allow short:
   - Default: **long-only** for v1.
2. Exit rule:
   - Default: exit when RSI crosses above **50**.
   - Optional variant later: exit at **70** or dual exit logic.
3. Minimum history required:
   - Default: at least `14 + 2` trading days (for RSI + crossing), but recommended `60+` for stability.
4. Liquidity filter:
   - Optional v2: only run signals for tickers above a liquidity score threshold.

