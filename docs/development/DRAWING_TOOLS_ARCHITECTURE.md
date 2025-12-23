# Drawing Tools Architecture (LWC)

This document defines the drawing system used on the Analysis chart.
It targets TradingView Lightweight Charts (LWC) primitives and avoids
React render loops by keeping draw state in a dedicated store.

## Goals

- TradingView-style tools (phased rollout).
- Stable chart behavior (no Y-axis shrink, no React render loops).
- Clean separation: data model, tool registry, rendering, interaction.
- Backend persistence by ticker (data/indicators/<ticker>.json).

## Core Concepts

### Drawing Model

Each drawing is a shape with:
- id (uuid)
- type (trendLine, horizontalLine, verticalLine, rectangle, text, measure, ...)
- points (time/price anchors)
- pane (pane index, default 0)
- style (stroke, fill, width, dash)
- meta (tool-specific payload for future extensions)
- locked (boolean)
- createdAt / updatedAt (unix ms)

Shapes are stored in a single collection keyed by id.

### Tool Registry

Tools are defined as metadata + behavior:
- id, label, icon
- cursor
- requiredPoints count
- default style
- renderer + hitTest mapping

Tools are grouped in the toolbar by phase.

### Store (State Machine)

A lightweight store tracks:
- activeTool
- draft shape (while drawing)
- selected shape ids
- hover id
- mode: idle | drawing | selected | editing | moving | resizing

Rendering observes the store, not React props, to avoid chart re-init loops.

## Rendering

Use LWC series primitives attached to the main candlestick series:
- One primitive renders all shapes.
- Primitive re-renders when store changes via requestUpdate().
- Selection overlays and handles are drawn in the same primitive.

### Redraw Throttling

During drag/move operations, the store can update many times per second. The drawing primitive throttles repaint work to
one `requestAnimationFrame` tick to avoid excessive canvas redraws.

Coordinate helpers live in `web/components/analysis/drawings/coordinates.ts`:
- timeScale().timeToCoordinate()
- timeScale().coordinateToTime()
- series.priceToCoordinate()
- series.coordinateToPrice()

## Interaction

Pointer events are bound to the chart container:
- On draw: create draft, update on move, commit on click/up.
- On select: hit-test, set selection.
- On edit: move/resize using handles.

Chart pan/zoom is disabled while drawing or dragging.

## Persistence

API:
- GET /api/v1/drawings/{ticker}
- PUT /api/v1/drawings/{ticker}

Storage:
- data/indicators/<ticker>.json
- Versioned JSON payload with schema validation.

Frontend saves on ticker change and debounced edits.

### Envelope Compatibility

The backend and frontend accept a versioned envelope:
- `{ version, ticker, updatedAt, shapes }`

For backward compatibility, a raw `[]shapes` array is also accepted and normalized.

## Phase 1 Tool Set

- Trend line
- Ray
- Horizontal line
- Vertical line
- Rectangle
- Text
- Arrow
- Measure

## Phase 2 Tool Set (Advanced)

- Parallel Channel (3-point)
- Brush (freehand)
- Fibonacci Extension
- Pitchfork
- Ellipse
