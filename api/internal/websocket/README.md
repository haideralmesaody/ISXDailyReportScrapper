# WebSocket Layer (Simplified)

The WebSocket implementation has been reduced to the smallest viable shape for same-machine communication.  
Only three primitives remain:

- `hub.go` – central broadcaster that owns registration, safe send, and lifecycle management.
- `client.go` – thin pump wrapping a `*websocket.Conn`; no wrappers, metrics, or retry loops.
- `hub_simplified_test.go` – lifecycle tests that exercise real HTTP upgrades and broadcast delivery.

## Testing

Run the focused package tests to validate registration and broadcast behaviour:

```bash
cd api
go test ./internal/websocket
```

## Design Notes

- No dead-letter queues, exponential backoff, or OTEL metrics.
- JSON payloads map directly to frontend expectations (`operation:snapshot`, `status`, `operation:progress`).
- Structured logging relies on `slog`; trace IDs propagate when present.
- Frontend reconnection logic lives in `web/lib/websocket.ts`.
