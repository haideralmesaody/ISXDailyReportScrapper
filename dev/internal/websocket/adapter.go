package websocket

import (
	"log/slog"
)

// MessageAdapter provides compatibility between different message formats
type MessageAdapter struct {
	hub    *Hub
	logger *slog.Logger
}

// NewMessageAdapter creates a new message adapter
func NewMessageAdapter(hub *Hub, logger *slog.Logger) *MessageAdapter {
	return &MessageAdapter{
		hub:    hub,
		logger: logger.With(slog.String("component", "websocket_adapter")),
	}
}

// BroadcastUpdate adapts various update types to hub broadcast methods
func (a *MessageAdapter) BroadcastUpdate(updateType, subtype, action string, data interface{}) {
	switch updateType {
	case "stage_progress":
		// Convert to output message
		if msg, ok := data.(map[string]interface{}); ok {
			if progress, ok := msg["progress"].(float64); ok {
				stage := ""
				if s, ok := msg["stage"].(string); ok {
					stage = s
				}
				message := ""
				if m, ok := msg["message"].(string); ok {
					message = m
				}
				a.hub.BroadcastOutput(formatProgressMessage(stage, int(progress), message), "info")
			}
		}
		
	case "data_update":
		// Use refresh for data updates
		components := []string{"all"}
		if subtype != "" {
			components = []string{subtype}
		}
		a.hub.BroadcastRefresh("adapter", components)
		
	case "output":
		// Direct output message
		if msg, ok := data.(map[string]interface{}); ok {
			message := msg["message"].(string)
			level := "info"
			if lvl, ok := msg["level"].(string); ok {
				level = lvl
			}
			a.hub.BroadcastOutput(message, level)
		}
		
	case "error":
		// Error message
		if msg, ok := data.(map[string]interface{}); ok {
			code := "ERR_UNKNOWN"
			if c, ok := msg["code"].(string); ok {
				code = c
			}
			stage := "system"
			if s, ok := msg["stage"].(string); ok {
				stage = s
			}
			message := msg["message"].(string)
			details := ""
			if d, ok := msg["details"].(string); ok {
				details = d
			}
			isRecoverable := false
			if recover, ok := msg["recoverable"].(bool); ok {
				isRecoverable = recover
			}
			a.hub.BroadcastError(code, message, details, stage, isRecoverable)
		}
		
	default:
		// For unknown types, broadcast as output
		a.hub.BroadcastOutput(formatGenericMessage(updateType, data), "info")
	}
}

// Helper functions
func formatProgressMessage(stage string, progress int, message string) string {
	if stage != "" {
		return stage + ": " + message
	}
	return message
}

func formatGenericMessage(msgType string, data interface{}) string {
	return "Update received"
}

// Register adds a client to the hub
func (a *MessageAdapter) Register(client *Client) {
	a.hub.Register(client)
}