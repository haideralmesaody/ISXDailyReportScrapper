package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/isxcli/isxcli/internal/errors"
)

// ClientLogHandler handles client-side logging requests
type ClientLogHandler struct {
	logger *slog.Logger
}

// NewClientLogHandler creates a new client log handler
func NewClientLogHandler(logger *slog.Logger) *ClientLogHandler {
	return &ClientLogHandler{
		logger: logger.With(slog.String("handler", "client_log")),
	}
}

// LogRequest represents a client log entry
type LogRequest struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Source  string                 `json:"source,omitempty"`
}

// Handle processes client logging requests
func (h *ClientLogHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewValidationError("Invalid request format"))
		return
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[req.Level] {
		req.Level = "info"
	}

	// Log with client context
	attrs := []slog.Attr{
		slog.String("client_source", req.Source),
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
	}

	if req.Data != nil {
		attrs = append(attrs, slog.Any("data", req.Data))
	}

	var level slog.Level
	switch req.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	h.logger.LogAttrs(r.Context(), level, req.Message, attrs...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}