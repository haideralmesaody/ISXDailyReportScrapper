# RSI(14) Mean Reversion (EOD) - Development Plan + User Stories

## Summary
Implement a **real-data** RSI(14) mean-reversion strategy that can be **manually triggered** and runs **across all tickers**. Results will be displayed in the **Strategies** page (`/strategy`). The standalone Alerts page remains removed.

**Core rules (RSI(14), long-only baseline):**
- **Entry (BUY):** RSI(14) crosses **below 30**
- **Exit (SELL):** RSI(14) crosses **above 50**

No mock data is allowed anywhere in the strategy or backtesting path.

**New requirement (Backtest integrated into Run Batch):**
- The user runs the strategy via **Run Batch** and can optionally enable **Include Backtest** (checkbox).
- Backtest uses a date range (default **last 90 days**) and can be adjusted.
- Results table shows per-ticker:
  - `completed_trades` (BUY→SELL pairs)
  - `winning_trades`, `losing_trades`
  - `gross_profit_pct` and `net_profit_pct` (fee-adjusted)
- The user can expand a ticker to view per-trade details (buy/sell dates, prices, gross/net %).
- Execution model: signals at day D, execute at **next day close (D+1)**.
- End-of-range: **mark-to-market** open position (affects profit %, does not count as completed trade).
- Fees: `transaction_fee = 0.006` per transaction (apply on BUY and SELL).

**UI annotation (Better Buy / Better Sell):**
- The `/strategy` signals table keeps the strategy `action` as `BUY|SELL|HOLD`.
- When backtest summary is available, the UI derives an additional “opportunity” badge from the **last backtest action**:
  - **Better Sell:** last action was `SELL` and the current price is **higher** than the last sell price.
  - **Better Buy:** last action was `BUY` (open position) and the current price is **lower** than the last buy price.
- The Latest Run panel displays these tags both in the Alerts list (when applicable) and in a dedicated “Better Opportunities” list.

---

## Progress / Current State

### Done
- Removed the old technical analysis / indicators UI and related pipeline stages (the direction is now "strategies only").
- Frontend TypeScript is green again (`web`: `npm.cmd run -s type-check`) and `/strategy` hosts the new strategies UI.
- `data/` is ignored by git (real data stays local, uncommitted).
- Alerts page is not present and remains removed for now.
- Strategy execution reads **real** EOD data from `data/reports/ticker/<SYMBOL>_trading_history.csv` and skips rows where `ClosePrice == 0` (no mock data).
- Implemented and registered the RSI strategy: `rsi14_mr_eod_v1` (BUY on cross below 30, SELL on cross above 50).
- Added batch execution endpoint: `POST /api/v1/strategies/{strategyID}/execute-batch` which persists results under `data/strategies/<strategy_id>/<run_id>/...`.
- Backtest integrated into Run Batch (`include_backtest` + 90d default + fees) and rendered in `/strategy` (metrics + expandable per-ticker trade details).
- Run history browsing implemented:
  - `GET /api/v1/strategies/{strategyID}/runs?limit=25`
  - `GET /api/v1/strategies/{strategyID}/runs/{runID}`

### Not Done Yet (next)
- Add WebSocket progress streaming for batch runs (optional; current batch runs are synchronous HTTP).
- Add concurrency guard to prevent overlapping batch runs per strategy.

---

## Current Implementation Status (high level)

### Frontend (`web/`)
- `/analysis` is available for charting and supports strategy deep-linking via query params:
  - `/analysis?ticker=<SYMBOL>`
  - `/analysis?ticker=<SYMBOL>&strategy_id=<STRATEGY_ID>&run_id=<RUN_ID>`
- Reports and Operations UIs compile after removing now-unused analysis/alerts-related pieces.

### Backend (`api/`)
- A strategy subsystem exists (`api/internal/strategy/*`) with WebSocket events:
  - `strategy_status`, `strategy_signal`, `strategy_backtest`
- The service layer uses a real-data provider (`api/internal/services/ticker_series.go`) and no longer uses mock series generation.

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

**Status:** Completed ✅

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

**Status:** Completed ✅

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

**Status:** Implemented (HTTP + persisted files). WebSocket progress: not yet. ✅/⬜

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
**Outcome:** Backtest RSI rules over a date range using real EOD data, integrated into Run Batch.

**Status:** Completed.

Work items:
1. Extend `POST /api/v1/strategies/{strategyID}/execute-batch` request:
   - `include_backtest` + `backtest_start_date` + `backtest_end_date` (default last 90 days)
   - `transaction_fee` (default `0.006`)
2. Backtest model (v1):
   - Long-only, 1 position max per ticker
   - Signal at day D, execute at next day close (D+1)
   - Mark-to-market open position at end date
3. Per-ticker outputs:
   - Summary metrics: completed/win/lose trades, gross/net profit %
   - Trades list: buy/sell dates/prices + gross/net return %
4. Persist:
   - `data/strategies/<id>/<run_id>/backtest/summary.json`
   - `data/strategies/<id>/<run_id>/backtest/aggregate.json`
   - `data/strategies/<id>/<run_id>/backtest/by_ticker/<symbol>.json`
5. Add API to fetch trade details:
   - `GET /api/v1/strategies/{strategyID}/runs/{runID}/backtest/{symbol}`

Deliverables:
- UI can display per-ticker metrics and expand tickers to view trade details.
- UI can display an aggregated backtest summary (avg/median net %, totals, top/bottom tickers).

---

### Milestone 5 - Frontend Strategies Page (new strategies only)
**Outcome:** `/strategy` becomes the canonical place to run and view strategy outputs.

**Status:** Completed (run UI + backtest + run browser). WebSocket streaming: not yet.

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
