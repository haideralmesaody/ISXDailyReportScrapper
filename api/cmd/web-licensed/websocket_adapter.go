package main

import (
	"log/slog"
	"os"
)

// stageAliases maintain backward compatibility with historic stage identifiers
// that the frontend expects. These mappings are intentionally minimal.
var stageAliases = map[string]string{
	"scraping":      "scrape",
	"processing":    "process",
	"indices":       "index",
	"liquidity":     "liquidity",
	"run_stage":     "run_stage",
	"full_pipeline": "full_pipeline",
}

// eventAliases normalize historical event names to the simplified variants.
var eventAliases = map[string]string{
	"operation_snapshot": "operation:snapshot",
}

// debugMode enables verbose adapter logging when ISX_DEBUG=true.
var debugMode = os.Getenv("ISX_DEBUG") == "true"

// hubBroadcaster captures the subset of methods required from the WebSocket hub.
type hubBroadcaster interface {
	BroadcastUpdate(eventType, step, status string, metadata interface{})
}

// WebSocketAdapter normalizes operation events before forwarding them to the hub.
type WebSocketAdapter struct {
	hub    hubBroadcaster
	logger *slog.Logger
}

// NewWebSocketAdapter creates an adapter that forwards messages to the provided hub.
func NewWebSocketAdapter(hub hubBroadcaster, logger *slog.Logger) *WebSocketAdapter {
	if logger == nil {
		logger = slog.Default()
	}

	return &WebSocketAdapter{
		hub:    hub,
		logger: logger.With(slog.String("component", "websocket_adapter")),
	}
}

// BroadcastUpdate implements operations.WebSocketHub by normalizing legacy payloads.
func (a *WebSocketAdapter) BroadcastUpdate(eventType, step, status string, metadata interface{}) {
	if a.hub == nil {
		a.logger.Warn("dropping WebSocket update: hub unavailable",
			slog.String("event_type", eventType),
			slog.String("step", step))
		return
	}

	normalizedType := normalizeEventType(eventType)
	normalizedStep := normalizeStage(step)
	payload := normalizeMetadata(metadata, normalizedStep, status)

	if debugMode {
		a.logger.Debug("forwarding WebSocket update",
			slog.String("event_type", normalizedType),
			slog.String("step", normalizedStep),
			slog.String("status", status))
	}

	a.hub.BroadcastUpdate(normalizedType, normalizedStep, status, payload)
}

func normalizeEventType(eventType string) string {
	if eventType == "" {
		return eventType
	}
	if alias, ok := eventAliases[eventType]; ok {
		return alias
	}
	return eventType
}

func normalizeStage(step string) string {
	if step == "" {
		return step
	}
	if alias, ok := stageAliases[step]; ok {
		return alias
	}
	return step
}

func normalizeMetadata(metadata interface{}, step, status string) interface{} {
	if metadata == nil {
		if step == "" && status == "" {
			return nil
		}
		return map[string]interface{}{
			"step":   step,
			"status": status,
		}
	}

	if data, ok := metadata.(map[string]interface{}); ok {
		if step != "" {
			if _, exists := data["step"]; !exists {
				data["step"] = step
			}
		}
		if status != "" {
			if _, exists := data["status"]; !exists {
				data["status"] = status
			}
		}
		return data
	}

	// Wrap non-map metadata so consumers always receive a structured payload.
	return map[string]interface{}{
		"value":  metadata,
		"step":   step,
		"status": status,
	}
}
