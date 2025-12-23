package operations

// WebSocketHub interface for sending WebSocket messages
type WebSocketHub interface {
	BroadcastUpdate(eventType, step, status string, metadata interface{})
}

// ProgressReporter interface for steps that can report progress
type ProgressReporter interface {
	ReportProgress(progress int, message string) error
}

// LicenseChecker interface for steps that need license validation
type LicenseChecker interface {
	CheckLicense() error
	RequiresLicense() bool
}

// StageOptions contains optional dependencies for steps
type StageOptions struct {
	WebSocketManager WebSocketHub
	LicenseChecker   LicenseChecker
	EnableProgress   bool
}
