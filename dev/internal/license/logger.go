package license

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel represents different logging levels
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Level      LogLevel  `json:"level"`
	Action     string    `json:"action"`
	LicenseKey string    `json:"license_key,omitempty"`
	MachineID  string    `json:"machine_id,omitempty"`
	Result     string    `json:"result"`
	Duration   int64     `json:"duration_ms,omitempty"`
	Error      string    `json:"error,omitempty"`
	Details    any       `json:"details,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
}

// Logger handles structured logging for the license system
type Logger struct {
	logFile   *os.File
	auditFile *os.File
	mutex     sync.Mutex
	level     LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(logLevel LogLevel) (*Logger, error) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join("logs", "license.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	auditFile, err := os.OpenFile(
		filepath.Join("logs", "audit.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to open audit file: %v", err)
	}

	return &Logger{
		logFile:   logFile,
		auditFile: auditFile,
		level:     logLevel,
	}, nil
}

// Log writes a structured log entry
func (l *Logger) Log(entry LogEntry) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.shouldLog(entry.Level) {
		return
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	l.logFile.WriteString(string(data) + "\n")
	l.logFile.Sync()

	if l.isAuditableAction(entry.Action) {
		l.auditFile.WriteString(string(data) + "\n")
		l.auditFile.Sync()
	}

	fmt.Printf("[%s] %s: %s\n", entry.Level, entry.Action, entry.Result)
}

func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	currentLevel, exists := levels[l.level]
	if !exists {
		currentLevel = 1
	}

	entryLevel, exists := levels[level]
	if !exists {
		entryLevel = 1
	}

	return entryLevel >= currentLevel
}

func (l *Logger) isAuditableAction(action string) bool {
	auditableActions := map[string]bool{
		"license_activation": true,
		"license_transfer":   true,
		"license_revocation": true,
		"license_extension":  true,
		"validation_failure": true,
		"security_violation": true,
		"admin_access":       true,
	}

	return auditableActions[action]
}

func (l *Logger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var errors []string
	if err := l.logFile.Close(); err != nil {
		errors = append(errors, fmt.Sprintf("log file: %v", err))
	}
	if err := l.auditFile.Close(); err != nil {
		errors = append(errors, fmt.Sprintf("audit file: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to close logger: %s", strings.Join(errors, ", "))
	}
	return nil
}
