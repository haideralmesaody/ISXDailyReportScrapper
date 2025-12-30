# RSI(14) Mean Reversion (EOD) - Development Plan + User Stories

## Summary
Implement a **real-data** RSI(14) mean-reversion strategy that can be **manually triggered** and runs **across all tickers**. Results will be displayed in the **Strategies** page (`/strategy`). The standalone Alerts page remains removed.

**Core rules (RSI(14), long-only baseline):**
- **Entry (BUY):** RSI(14) crosses **below 30**
- **Exit (SELL):** RSI(14) crosses **above 50**

No mock data is allowed anywhere in the strategy or backtesting path.

---

## Progress / Current State

### Done
- Removed the old technical analysis / indicators UI and related pipeline stages (the direction is now "strategies only").
- Frontend TypeScript is green again (`web`: `npm.cmd run -s type-check`).
- `data/` is ignored by git (real data stays local, uncommitted).
- Alerts page is not present and remains removed for now.

### Not Done Yet (next)
- Backend still contains mock strategy data paths; strategy execution/backtesting must be wired to **real pipeline output** (CSV files produced by the scraper + processor).
- Implement batch execution for RSI(14) across all tickers and persist results under `data/strategies/...`.
- Implement batch backtesting per ticker (real data only), plus aggregated results and UI rendering under `/strategy`.

---

## Current Implementation Status (high level)

### Frontend (`web/`)
- `/analysis` is removed; `/strategy` is the home for new strategies going forward.
- Reports and Operations UIs compile after removing now-unused analysis/alerts-related pieces.

### Backend (`api/`)
- A strategy subsystem exists (`api/internal/strategy/*`) with WebSocket events:
  - `strategy_status`, `strategy_signal`, `strategy_backtest`
- The service layer still needs to remove mock paths and implement a real-data provider (see Milestone 1).

---

## Architecture Decision (Recommended)

### Option A (recommended): Use existing strategy subsystem
- Keep strategy execution in `api/internal/strategy` + `api/internal/services/strategy_service.go`.
- Add batch endpoints to run a strategy across all tickers, emit progress via existing WebSocket events, and persist outputs under `data/strategies/...`.

This matches the existing backend direction and avoids prematurely coupling strategy runs to the Operations pipeline UI.

---

## Milestones

### Milestone 1 - Remove mock data, add real-data provider
**Outcome:** Strategy execution reads real EOD data and never generates fake prices.

Work items:
1. Replace `getMockData()` usage in `api/internal/services/strategy_service.go` with real data retrieval.
2. Decide the canonical source for OHLCV (preferred: pipeline output after processing).
3. Implement a `StrategyDataProvider` interface:
   - `ListSymbols() []string`
   - `GetSymbolSeries(symbol, limit)` (execute)
   - `GetSymbolSeriesRange(symbol, start, end)` (backtest)
4. Make errors explicit:
   - Missing file vs insufficient history vs parse failure.

Deliverables:
- Strategy execution for a single symbol returns a real EOD-based signal.

---

### Milestone 2 - Implement RSI(14) Mean Reversion Strategy (EOD)
**Outcome:** A new strategy ID exists with the exact RSI rules above.

Work items:
1. Add a dedicated strategy implementation, e.g. `api/internal/strategy/strategies/rsi14_mr_eod.go`
   - Strategy ID: `rsi14_mr_eod_v1`
2. Parameters (v1):
   - `rsi_period = 14` (fixed)
   - `entry_threshold = 30` (fixed)
   - `exit_threshold = 50` (fixed)
   - `allow_short = false` (fixed)
3. Signal rules:
   - BUY when RSI crosses below 30 (today < 30 and yesterday >= 30)
   - SELL when RSI crosses above 50 (today > 50 and yesterday <= 50)
   - HOLD otherwise
4. Include RSI and threshold metadata in emitted signals.

Deliverables:
- Unit tests for RSI calculation + crossing logic using deterministic numeric sequences (math-only tests, not mock market data).

---

### Milestone 3 - Batch Run Across All Tickers (manual trigger)
**Outcome:** RSI strategy runs across all tickers with progress + persisted results.

Work items:
1. Add endpoint:
   - `POST /api/v1/strategies/{strategyID}/execute-batch`
2. Batch stages:
   - A: Load symbol list
   - B: Iterate symbols, compute RSI + signal (EOD)
   - C: Persist `signals.json`/`signals.csv` + `summary.json`
3. WebSocket updates:
   - Stage transitions + counts via `strategy_status`
   - Optionally stream signals via `strategy_signal`
4. Concurrency guard:
   - Prevent overlapping runs for the same strategy unless explicitly allowed.

Deliverables:
- A run produces `data/strategies/rsi14_mr_eod_v1/<run_id>/...` with signals for all symbols.

---

### Milestone 4 - Backtesting (per ticker + aggregated)
**Outcome:** Backtest RSI rules over a date range using real EOD data.

Work items:
1. Add endpoint:
   - `POST /api/v1/strategies/{strategyID}/backtest-batch`
2. Backtest model (v1):
   - Long-only, 1 position max per ticker
   - Entry/exit cross rules only
3. Persist:
   - `data/strategies/<id>/<run_id>/backtest/summary.json`
   - `data/strategies/<id>/<run_id>/backtest/by_ticker/<symbol>.json`
4. WebSocket:
   - Stream progress and publish final summary.

Deliverables:
- UI can display backtest summary + per-ticker results.

---

### Milestone 5 - Frontend Strategies Page (new strategies only)
**Outcome:** `/strategy` becomes the canonical place to run and view strategy outputs.

Work items:
1. Add strategy list UI (start with RSI14 MR EOD).
2. Add "Run now" and "Backtest" actions.
3. Add results panels:
   - Last run metadata (run_id, as_of_date, BUY/SELL/HOLD counts)
   - Signals table with filters (ticker/action)
4. Subscribe to strategy WebSocket events and update UI live.
5. Market Overview links to `/strategy?ticker=...` (analysis is removed).

Deliverables:
- A trader can run RSI14 MR EOD and see signals in `/strategy`.

---

## User Stories (with Acceptance Criteria)

### Epic: Cleanup (Remove TA/Indicators)
**US-0.1 - Analysis is gone**
- As a user, I do not see an Analysis page in navigation.
- Acceptance:
  - No nav item links to `/analysis`.

**US-0.2 - Strategies is the primary signals UI**
- As a user, when I click a ticker from Market Overview, I land on `/strategy?ticker=<symbol>`.
- Acceptance:
  - Market Overview click behavior uses `/strategy`.

---

### Epic: RSI Strategy (Real Data)
**US-1.1 - Run RSI strategy across all tickers**
- As a trader, I can click "Run RSI(14) Mean Reversion (EOD)" and generate signals for all tickers using real data.
- Acceptance:
  - API returns a `run_id`.
  - WebSocket shows progress and completion.
  - Output files are created under `data/strategies/rsi14_mr_eod_v1/<run_id>/`.

**US-1.2 - View latest signals**
- As a trader, I can view today's BUY/SELL signals in one place in `/strategy`.
- Acceptance:
  - Signals view can filter by ticker and action.

**US-1.3 - Verify crossing logic**
- As a trader, I only get a BUY when RSI crosses below 30 (not every day RSI stays below 30).
- Acceptance:
  - For a known RSI sequence, backend produces BUY only on the crossing day.

---

### Epic: Backtesting
**US-2.1 - Backtest RSI per ticker**
- As a trader, I can backtest RSI14 MR over a date range and see per-ticker results.
- Acceptance:
  - Backtest endpoint returns per-ticker results plus totals.
  - UI displays an aggregated summary + sortable per-ticker table.

**US-2.2 - Reproducible results**
- As a trader, running the same backtest date range yields the same results.
- Acceptance:
  - No randomness in backtest.
  - Same input => same output files.

---

## Open Decisions (can default)
1. Long-only vs allow short:
   - Default: long-only for v1.
2. Exit rule:
   - Default: exit when RSI crosses above 50.
   - Optional variant later: exit at 70 or dual-exit logic.
3. Minimum history required:
   - Default: at least 16 trading days (14 + 2 for crossing), recommended 60+ for stability.
4. Liquidity filter:
   - Optional v2: only run signals for tickers above a liquidity threshold.
