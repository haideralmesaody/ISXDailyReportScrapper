# Strategy UI Notes

## Goals
- No mock data: the Strategy page renders only data returned by the backend strategy endpoints.
- SSOT for derived UI annotations (e.g. Better Buy/Sell): compute once and reuse across the page.
- Keep strategy `action` semantics stable (`BUY|SELL|HOLD`) to avoid breaking counts/backtests.

## Better Buy / Better Sell
The Strategy page annotates signals with an additional “opportunity” badge derived from the
last backtest action vs the current price:

- **Better Sell:** last backtest action was `SELL` and current price is higher than the last sell price.
- **Better Buy:** last backtest action was `BUY` (open position) and current price is lower than the last buy price.

Implementation:
- Backend exposes `last_action`, `last_action_date`, `last_action_price` on `backtest.by_ticker[]`.
- Frontend computes the derived badge via `web/lib/utils/better-opportunity.ts` and uses it in:
  - Latest Run summary (alerts + opportunities)
  - Signals table rows

## Rendering model
The Strategy page builds a view-model for the selected run:
- merge `signals[]` with `backtest.by_ticker[]` summary for the same symbol
- compute derived fields (better opportunity, deltas, formatted values)
This avoids divergent rendering paths (e.g. showing a badge in one list but not another).

