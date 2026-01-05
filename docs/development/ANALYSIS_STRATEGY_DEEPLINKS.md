# Analysis Deep Links (Market/Strategy → Analysis)

## URLs (public contract)

### Market Overview → Analysis (ticker only)
`/analysis?ticker=<SYMBOL>`

Behavior:
- Auto-selects the ticker in the analysis page.
- No strategy-specific indicators or overlays are applied.

### Strategy → Analysis (ticker + strategy overlay)
`/analysis?ticker=<SYMBOL>&strategy_id=<STRATEGY_ID>&run_id=<RUN_ID>`

Behavior:
- Auto-selects the ticker in the analysis page.
- Fetches the strategy chart preset and applies it (chart type, timeframe, indicators, RSI parameters).
- Fetches backtest trades for the selected ticker from that run and overlays BUY/SELL markers on the chart.
- Markers are filtered to **tradable dates only** (no non-trading day markers).

## Backend

### Strategy chart preset endpoint
`GET /api/v1/strategies/{strategyID}/chart-preset`

Returned shape (example):
```json
{
  "strategy_id": "rsi14_mr_eod_v1",
  "chart_type": "candlestick",
  "timeframe": "MAX",
  "enabled_indicators": ["showVolume", "showRSI"],
  "momentum": { "rsi_period": 14, "rsi_overbought": 50, "rsi_oversold": 30 }
}
```

How it works:
- Strategies can optionally implement `strategy.ChartPresetProvider`.
- The strategy service returns 404 if no preset is provided.

### Backtest trade details endpoint (for markers)
`GET /api/v1/strategies/{strategyID}/runs/{runID}/backtest/{symbol}`

The analysis page uses `trades[].buy_date/sell_date` (or `signal_*_date`) as marker dates.

## Frontend

### Strategy symbol click
- Implemented in `web/app/strategy/strategy-client.tsx` and navigates to `/analysis` with `ticker`, `strategy_id`, and `run_id` (when available).

### Analysis page behavior
- Implemented in `web/app/analysis/analysis-client.tsx`.
- Applies preset once per deep-link (then user can change chart/indicators normally).
- Overlays BUY/SELL markers from the run backtest on the chart.

### Marker mapping
- Implemented in `web/lib/utils/trade-markers.ts`.
- Dates not present in the current ticker’s tradable chart data are ignored.

