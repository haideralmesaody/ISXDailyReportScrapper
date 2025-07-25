package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"isxcli/internal/errors"
	"isxcli/internal/services"
)

// ClientLogHandler handles client-side logging requests
type ClientLogHandler struct {
	logger services.Logger
}

// NewClientLogHandler creates a new client log handler
func NewClientLogHandler(logger services.Logger) *ClientLogHandler {
	return &ClientLogHandler{
		logger: logger,
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
	logFields := []interface{}{
		"client_source", req.Source,
		"timestamp", time.Now().Format(time.RFC3339),
	}

	if req.Data != nil {
		logFields = append(logFields, "data", req.Data)
	}

	switch req.Level {
	case "debug":
		h.logger.Debug(req.Message, logFields...)
	case "info":
		h.logger.Info(req.Message, logFields...)
	case "warn":
		h.logger.Warn(req.Message, logFields...)
	case "error":
		h.logger.Error(req.Message, logFields...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}