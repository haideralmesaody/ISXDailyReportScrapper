# Drawing Tools + Indicators Plan (Option 2)

This plan implements TradingView-style drawing tools and standard indicators using
TradingView Lightweight Charts (LWC). It avoids proprietary Charting Library usage.

## Scope

### Drawing Tools

Phase 1 (core):
- Trend line
- Ray
- Horizontal line
- Vertical line
- Rectangle
- Text
- Measure (price + percent + bars)
- Arrow

Phase 2:
- Parallel channel
- Brush

Phase 3:
- Fibonacci retracement
- Fibonacci extension
- Pitchfork
- Ellipse

### Indicators (Standard TradingView)

- Moving averages: SMA 20/50/200, EMA 20, custom SMA/EMA
- Momentum: RSI, MACD, MFI, Stochastic, CCI, Williams %R, Momentum, ROC, ADX
- Volatility: Bollinger Bands, ATR, Keltner Channels
- Volume: Volume, OBV, VWAP
- Trend/overlay: Donchian, Ichimoku, Parabolic SAR, Fibonacci levels

## Storage Requirement

- Persist drawings per ticker (not per user).
- Storage path: data/indicators/<ticker>.json
- API:
  - GET /api/v1/drawings/{ticker}
  - PUT /api/v1/drawings/{ticker}

## Implementation Phases

### Step 1 - Architecture (Design)
- Define drawing models and tool registry.
- Implement drawing state machine.
- Add coordinate conversion helpers for time/price <-> pixel.
- Document architecture and conventions.

### Step 2 - Rendering Layer
- Implement DrawingLayer with LWC primitives.
- Add renderers and hit-testing for Phase 1 tools.
- Add selection overlay + handles.

### Step 3 - Interaction Engine
- Pointer event handling for draw/select/move/resize.
- Disable chart pan/zoom while drawing.
- Throttle updates with requestAnimationFrame.

### Step 4 - Toolbar UI
- Add vertical toolbar with grouped tools.
- Add delete, lock, undo/redo, visibility actions.
- Persist last tool in localStorage.

### Step 5 - Backend Persistence
- Add HTTP handler for drawing GET/PUT.
- Validate payloads and ticker id.
- Save to data/indicators/<ticker>.json.
- Wire frontend load/save on ticker change.

### Step 6 - Indicator Parity
- Audit existing indicators against standard list.
- Add missing indicators or settings.
- Wire indicator toggles + tooltips.

### Step 7 - Tests + Performance
- Unit tests for drawing serialization + hit-testing.
- API tests for save/load endpoints.
- Performance safeguards: RAF throttling, minimal re-renders.

### Step 8 - Documentation
- Update guide section for drawings and indicators.
- Update CHANGELOG.
