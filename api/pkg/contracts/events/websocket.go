// Package events contains simplified event contract definitions for WebSocket communication
// in the ISX Daily Reports Scrapper system.
package events

import (
	"time"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Core operational message - the primary event type
	MessageTypeOperationSnapshot MessageType = "operation:snapshot"
	
	// System messages
	MessageTypeSystemStatus  MessageType = "system:status"
	MessageTypeSystemMetrics MessageType = "system:metrics"
	
	// Market messages  
	MessageTypeMarketUpdate MessageType = "market:update"
	
	// Connection messages
	MessageTypeConnect    MessageType = "connect"
	MessageTypeDisconnect MessageType = "disconnect"
	MessageTypeError      MessageType = "error"
)

// BaseMessage represents the base structure for all WebSocket messages
type BaseMessage struct {
	ID        string      `json:"id,omitempty"`        // Unique message ID
	Type      MessageType `json:"type"`                // Message type
	Timestamp time.Time   `json:"timestamp"`           // Message timestamp
	TraceID   string      `json:"trace_id,omitempty"`  // Request trace ID
}

// WebSocketMessage represents a complete WebSocket message
type WebSocketMessage struct {
	BaseMessage
	Data     interface{} `json:"data,omitempty"`     // Message payload
	Subtype  string      `json:"subtype,omitempty"`  // Legacy support
	Action   string      `json:"action,omitempty"`   // Legacy support
}

// OperationSnapshot is the primary message type for all operation updates
// This is the ONLY message type used for scraping/processing progress
type OperationSnapshot struct {
	OperationID string         `json:"operation_id"`
	Status      string         `json:"status"`       // pending|running|completed|failed|cancelled
	Progress    int            `json:"progress"`     // 0-100
	CurrentStep string         `json:"current_step"` // Current active step name
	Steps       []StepSnapshot `json:"steps"`        // All steps with their status
	StartedAt   time.Time      `json:"started_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Error       string         `json:"error,omitempty"`
	Message     string         `json:"message,omitempty"`
}

// StepSnapshot represents the state of a single step
type StepSnapshot struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Status   string                 `json:"status"`   // pending|running|completed|failed|skipped
	Progress int                    `json:"progress"` // 0-100
	Message  string                 `json:"message,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Contains scraping details
}

// SubscriptionOptions represents subscription options (for protocol.go compatibility)
type SubscriptionOptions struct {
	BufferSize     int    `json:"buffer_size,omitempty"`
	MaxFrequency   int    `json:"max_frequency,omitempty"` // Max messages per second
	IncludeHistory bool   `json:"include_history,omitempty"`
	HistoryLimit   int    `json:"history_limit,omitempty"`
	Quality        string `json:"quality,omitempty"` // realtime, delayed, snapshot
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	BaseMessage
	Data struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
		Retry   bool        `json:"retry"`
		Fatal   bool        `json:"fatal"`
	} `json:"data"`
}

// SystemStatusEvent represents a system status event
type SystemStatusEvent struct {
	BaseMessage
	Data struct {
		Status     string              `json:"status"` // healthy|degraded|unhealthy
		Components map[string]string   `json:"components"`
		Uptime     string              `json:"uptime"`
		Version    string              `json:"version"`
	} `json:"data"`
}

// SystemMetricsEvent represents system metrics event
type SystemMetricsEvent struct {
	BaseMessage
	Data struct {
		CPU         float64   `json:"cpu_percent"`
		Memory      float64   `json:"memory_percent"`
		Disk        float64   `json:"disk_percent"`
		Connections int       `json:"active_connections"`
		RequestRate float64   `json:"request_rate"`
		ErrorRate   float64   `json:"error_rate"`
		Timestamp   time.Time `json:"timestamp"`
	} `json:"data"`
}