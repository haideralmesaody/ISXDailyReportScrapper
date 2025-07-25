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

// ProgressReporter interface for stages that can report progress
type ProgressReporter interface {
	ReportProgress(progress int, message string) error
}

// LicenseChecker interface for stages that need license validation
type LicenseChecker interface {
	CheckLicense() error
	RequiresLicense() bool
}

// StageOptions contains optional dependencies for stages
type StageOptions struct {
	WebSocketManager WebSocketHub
	LicenseChecker   LicenseChecker
	EnableProgress   bool
}