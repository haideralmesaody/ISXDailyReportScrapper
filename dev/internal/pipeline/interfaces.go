package pipeline

// WebSocketHub interface for sending WebSocket messages
type WebSocketHub interface {
	BroadcastUpdate(eventType, stage, status string, metadata interface{})
}

// Logger interface for logging
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}