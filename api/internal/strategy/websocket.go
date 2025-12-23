package strategy

import (
	"context"
	"log/slog"
	"time"

	"github.com/isxcli/isxcli/internal/websocket"
)

// WebSocketManager handles real-time strategy updates
type WebSocketManager struct {
	hub    *websocket.Hub
	logger *slog.Logger
}

// NewWebSocketManager creates a new WebSocket manager for strategies
func NewWebSocketManager(hub *websocket.Hub, logger *slog.Logger) *WebSocketManager {
	return &WebSocketManager{
		hub:    hub,
		logger: logger,
	}
}

// BroadcastSignal sends a new signal to all connected clients
func (w *WebSocketManager) BroadcastSignal(ctx context.Context, signal Signal) error {
	message := StrategyMessage{
		Type:      "signal",
		Signal:    &signal,
		Timestamp: time.Now(),
	}

	w.hub.Broadcast("strategy_signal", message)

	w.logger.InfoContext(ctx, "signal broadcast",
		"signal_id", signal.ID,
		"strategy_id", signal.StrategyID,
		"symbol", signal.Symbol,
		"action", signal.Action,
	)

	return nil
}

// BroadcastBacktestUpdate sends backtest progress updates
func (w *WebSocketManager) BroadcastBacktestUpdate(ctx context.Context, update BacktestUpdate) error {
	message := StrategyMessage{
		Type:           "backtest_update",
		BacktestUpdate: &update,
		Timestamp:      time.Now(),
	}

	w.hub.Broadcast("strategy_backtest", message)
	return nil
}

// BroadcastStrategyStatus sends strategy status updates
func (w *WebSocketManager) BroadcastStrategyStatus(ctx context.Context, status StrategyStatus) error {
	message := StrategyMessage{
		Type:           "strategy_status",
		StrategyStatus: &status,
		Timestamp:      time.Now(),
	}

	w.hub.Broadcast("strategy_status", message)
	return nil
}

// StrategyMessage represents a WebSocket message for strategy updates
type StrategyMessage struct {
	Type           string           `json:"type"`
	Signal         *Signal          `json:"signal,omitempty"`
	BacktestUpdate *BacktestUpdate  `json:"backtest_update,omitempty"`
	StrategyStatus *StrategyStatus  `json:"strategy_status,omitempty"`
	Timestamp      time.Time        `json:"timestamp"`
}

// BacktestUpdate represents progress of a running backtest
type BacktestUpdate struct {
	BacktestID string  `json:"backtest_id"`
	StrategyID string  `json:"strategy_id"`
	Symbol     string  `json:"symbol"`
	Progress   float64 `json:"progress"` // 0-100
	Status     string  `json:"status"`   // "running", "completed", "failed"
	Message    string  `json:"message,omitempty"`
}

// StrategyStatus represents the current status of a strategy
type StrategyStatus struct {
	StrategyID   string    `json:"strategy_id"`
	Status       string    `json:"status"` // "active", "inactive", "error"
	LastExecuted time.Time `json:"last_executed,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	SignalCount  int       `json:"signal_count"`
}