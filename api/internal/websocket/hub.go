package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/isxcli/isxcli/internal/infrastructure"
)

const (
	defaultBroadcastBuffer = 256
)

// Legacy message type constants for backward compatibility
const (
	TypeConnection       = "connection"
	TypeProgress         = "progress"
	TypeOutput           = "output"
	TypeError            = "error"
	TypeDataUpdate       = "data_update"
	TypeOperationStatus  = "operation:status"
	TypePipelineProgress = "operation:progress"
	TypePipelineComplete = "operation:complete"
	TypeLog              = "log"
	TypeStatus           = "status"

	// Subtypes
	SubtypeAll           = "all"
	SubtypeTickerSummary = "ticker_summary"

	// Actions
	ActionRefresh = "refresh"
	ActionCreated = "created"

	// Message levels
	LevelInfo    = "info"
	LevelSuccess = "success"
	LevelWarning = "warning"
	LevelError   = "error"
)

// Error codes for websocket operations
const (
	ErrScrapingTimeout       = "SCRAPING_TIMEOUT"
	ErrScrapingNoData        = "SCRAPING_NO_DATA"
	ErrProcessingInvalidFile = "PROCESSING_INVALID_FILE"
	ErrSystemDiskFull        = "SYSTEM_DISK_FULL"
)

// ErrorRecoveryHints provides user-friendly recovery suggestions
var ErrorRecoveryHints = map[string]string{
	"default": "Please try again or contact support",
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger instance
	logger *slog.Logger

	// Metrics
	totalConnections  int64
	activeConnections int64
	messagesSent      int64

	// Control
	quit    chan struct{}
	running bool
}

// NewHub creates a new Hub instance with dependency injection
func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = infrastructure.GetLogger()
	}
	logger = logger.With(slog.String("component", "websocket.hub"))

	return &Hub{
		broadcast:  make(chan []byte, defaultBroadcastBuffer),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		logger:     logger,
		quit:       make(chan struct{}),
	}
}

// safeSend safely sends a message to a client without causing panic
func (h *Hub) safeSend(client *Client, data []byte) bool {
	// Check if client is still registered
	h.mu.RLock()
	_, exists := h.clients[client]
	h.mu.RUnlock()

	if !exists {
		h.logger.Warn("Attempting to send to unregistered client",
			slog.String("client_id", client.id))
		return false
	}

	// Use recover to catch closed channel panics
	defer func() {
		if r := recover(); r != nil {
			h.logger.Warn("Recovered from closed channel panic during send",
				slog.String("client_id", client.id),
				slog.Any("panic", r))
			// Schedule unregistration to avoid deadlock
			go func() {
				h.unregister <- client
			}()
		}
	}()

	// Try non-blocking send first
	select {
	case client.send <- data:
		return true
	default:
		// Channel is full, try with short timeout
		select {
		case client.send <- data:
			return true
		case <-time.After(25 * time.Millisecond):
			h.logger.Warn("Client send channel timeout, disconnecting slow client",
				slog.String("client_id", client.id),
				slog.Int("message_size", len(data)))
			// Schedule unregistration to avoid deadlock
			go func() {
				h.unregister <- client
			}()
			return false
		}
	}
}

// Start starts the hub's goroutines
func (h *Hub) Start() {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.mu.Unlock()

	// Start the main hub loop
	go h.Run()
}

// Run starts the hub's main loop with panic recovery
func (h *Hub) Run() {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Hub panic recovered, preventing server crash",
				slog.Any("panic", r))
		}
	}()

	for {
		select {
		case <-h.quit:
			h.logger.Info("Hub shutting down")
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient handles client registration - simplified without retry logic
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()

	// Check if hub is still running
	if !h.running {
		h.mu.Unlock()
		h.logger.Error("Hub is not running, cannot register client",
			slog.String("client_id", client.id))
		return
	}

	// Check if client is already registered
	if _, exists := h.clients[client]; exists {
		h.mu.Unlock()
		h.logger.Warn("Client already registered",
			slog.String("client_id", client.id))
		return
	}

	// Add client to registry
	h.clients[client] = true
	count := len(h.clients)
	h.totalConnections++
	h.activeConnections = int64(count)

	// Store connection time
	client.connectedAt = time.Now()

	h.mu.Unlock()

	ctx := context.Background()
	if client.traceID != "" {
		ctx = infrastructure.WithTraceID(ctx, client.traceID)
	}

	h.logger.InfoContext(ctx, "WebSocket client registered successfully",
		slog.Int("total_clients", count),
		slog.Int64("total_connections", h.totalConnections),
		slog.String("client_id", client.id),
		slog.String("remote_addr", client.remoteAddr))

	// Send connection success message
	h.sendConnectionMessage(client, ctx)
}

// sendConnectionMessage sends the connection message safely
func (h *Hub) sendConnectionMessage(client *Client, ctx context.Context) {
	connMsg := map[string]interface{}{
		"type": TypeConnection,
		"data": map[string]interface{}{
			"status":    "connected",
			"message":   "Connected to ISX WebSocket",
			"client_id": client.id,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if client.traceID != "" {
		connMsg["trace_id"] = client.traceID
	}

	jsonData, err := json.Marshal(connMsg)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to marshal connection message",
			slog.String("error", err.Error()),
			slog.String("client_id", client.id))
		return
	}

	if !h.safeSend(client, jsonData) {
		h.logger.WarnContext(ctx, "Failed to send connection message",
			slog.String("client_id", client.id))
	}
}

// unregisterClient handles client unregistration
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		count := len(h.clients)
		h.activeConnections = int64(count)
		h.mu.Unlock()

		ctx := context.Background()
		if client.traceID != "" {
			ctx = infrastructure.WithTraceID(ctx, client.traceID)
		}

		h.logger.InfoContext(ctx, "Client unregistered",
			slog.Int("total_clients", count),
			slog.String("client_id", client.id),
			slog.Duration("connection_duration", time.Since(client.connectedAt)))

	} else {
		h.mu.Unlock()
	}
}

// broadcastMessage handles message broadcasting
func (h *Hub) broadcastMessage(message []byte) {
	h.mu.RLock()
	// Create a copy of clients to avoid holding lock during send
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// Try to parse message for enhanced logging
	var messageData map[string]interface{}
	messageType := "unknown"
	operationID := "unknown"
	status := "unknown"

	if err := json.Unmarshal(message, &messageData); err == nil {
		if msgType, ok := messageData["type"].(string); ok {
			messageType = msgType
		}
		if data, ok := messageData["data"].(map[string]interface{}); ok {
			if opID, ok := data["operation_id"].(string); ok {
				operationID = opID
			}
			if opStatus, ok := data["status"].(string); ok {
				status = opStatus
			}
		}
	}

	h.logger.Info("BROADCASTING MESSAGE TO CLIENTS",
		slog.Int("client_count", len(clients)),
		slog.Int("message_size", len(message)),
		slog.String("message_type", messageType),
		slog.String("operation_id", operationID),
		slog.String("status", status),
		slog.String("timestamp", time.Now().Format(time.RFC3339Nano)),
		slog.String("component", "Hub.broadcastMessage"))

	successCount := 0
	failCount := 0
	clientDetails := []map[string]interface{}{}

	// Send to all clients
	for _, client := range clients {
		clientSuccess := h.safeSend(client, message)
		if clientSuccess {
			successCount++
			h.messagesSent++
		} else {
			failCount++
		}

		clientDetails = append(clientDetails, map[string]interface{}{
			"client_id":    client.id,
			"remote_addr":  client.remoteAddr,
			"success":      clientSuccess,
			"connected_at": client.connectedAt.Format(time.RFC3339),
		})
	}

	h.logger.Info("BROADCAST COMPLETED",
		slog.Int("success_count", successCount),
		slog.Int("fail_count", failCount),
		slog.String("message_type", messageType),
		slog.String("operation_id", operationID),
		slog.String("status", status),
		slog.Any("client_details", clientDetails),
		slog.String("timestamp", time.Now().Format(time.RFC3339Nano)),
		slog.String("component", "Hub.broadcastMessage"))

	if failCount > 0 {
		h.logger.Warn("SOME CLIENTS FAILED TO RECEIVE BROADCAST",
			slog.Int("success_count", successCount),
			slog.Int("fail_count", failCount),
			slog.String("message_type", messageType),
			slog.String("operation_id", operationID),
			slog.String("timestamp", time.Now().Format(time.RFC3339Nano)),
			slog.String("component", "Hub.broadcastMessage"))
	}
}

// BroadcastUpdate sends a data update message to all connected clients
func (h *Hub) BroadcastUpdate(updateType, subtype, action string, data interface{}) {
	h.BroadcastUpdateWithTrace(updateType, subtype, action, data, "")
}

// BroadcastUpdateWithTrace sends a data update message with trace ID
func (h *Hub) BroadcastUpdateWithTrace(updateType, subtype, action string, data interface{}, traceID string) {
	message := map[string]interface{}{
		"type":      updateType,
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Add trace ID if provided
	if traceID != "" {
		message["trace_id"] = traceID
	}

	// For operation:snapshot, data already contains necessary information
	// For other events, preserve backward compatibility
	if updateType != "operation:snapshot" && updateType != "" {
		message["subtype"] = subtype
		message["action"] = action
	}

	h.broadcastJSON(message)
}

// broadcastJSON is a helper method to send JSON messages
func (h *Hub) broadcastJSON(message map[string]interface{}) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		ctx := context.Background()
		if traceID, ok := message["trace_id"].(string); ok && traceID != "" {
			ctx = infrastructure.WithTraceID(ctx, traceID)
		}
		h.logger.ErrorContext(ctx, "Error marshaling message",
			slog.String("error", err.Error()),
			slog.String("message_type", message["type"].(string)))
		return
	}

	select {
	case h.broadcast <- jsonData:
		h.mu.Lock()
		h.messagesSent++
		h.mu.Unlock()
	default:
		h.logger.Warn("Broadcast channel full, dropping message")
	}
}

// BroadcastProgress sends a progress update message
func (h *Hub) BroadcastProgress(step string, progress int, message string) {
	update := map[string]interface{}{
		"type": TypeProgress,
		"data": map[string]interface{}{
			"step":     step,
			"progress": progress,
			"message":  message,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.broadcastJSON(update)
}

// BroadcastStatus sends a status update message
func (h *Hub) BroadcastStatus(status, message string) {
	h.BroadcastStatusWithTrace(status, message, "")
}

// BroadcastStatusWithTrace sends a status update message with trace ID
func (h *Hub) BroadcastStatusWithTrace(status, message, traceID string) {
	update := map[string]interface{}{
		"type": "status",
		"data": map[string]interface{}{
			"status":  status,
			"message": message,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if traceID != "" {
		update["trace_id"] = traceID
	}

	h.broadcastJSON(update)
}

// BroadcastError sends a structured error message
func (h *Hub) BroadcastError(code, message, details, step string, recoverable bool) {
	hint := ErrorRecoveryHints[code]
	if hint == "" {
		hint = "Please try again or contact support"
	}

	errorMsg := map[string]interface{}{
		"type": TypeError,
		"data": map[string]interface{}{
			"code":        code,
			"message":     message,
			"details":     details,
			"step":        step,
			"recoverable": recoverable,
			"hint":        hint,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.broadcastJSON(errorMsg)
}

// BroadcastOutput sends an output message
func (h *Hub) BroadcastOutput(message, level string) {
	update := map[string]interface{}{
		"type": TypeOutput,
		"data": map[string]interface{}{
			"message": message,
			"level":   level,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.broadcastJSON(update)
}

// BroadcastRefresh sends a data refresh notification
func (h *Hub) BroadcastRefresh(source string, components []string) {
	h.BroadcastUpdate(TypeDataUpdate, SubtypeAll, ActionRefresh, map[string]interface{}{
		"source":     source,
		"components": components,
	})
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Broadcast implements the services.WebSocketHub interface
func (h *Hub) Broadcast(messageType string, data interface{}) {
	h.BroadcastUpdate(messageType, "", "", data)
}

// IsRunning returns whether the hub is currently running
func (h *Hub) IsRunning() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// Stop gracefully stops the hub
func (h *Hub) Stop() {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.mu.Unlock()

	// Signal goroutines to stop
	close(h.quit)

	// Close all client connections
	h.mu.Lock()
	defer h.mu.Unlock()
	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	select {
	case h.register <- client:
	case <-time.After(5 * time.Second):
		h.logger.Error("Client registration timed out",
			slog.String("client_id", client.id))
	}
}

// GetHubMetrics returns current hub metrics
func (h *Hub) GetHubMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"active_clients":    len(h.clients),
		"total_connections": h.totalConnections,
		"messages_sent":     h.messagesSent,
	}
}
