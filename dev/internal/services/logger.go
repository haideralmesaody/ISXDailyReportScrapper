package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"isxcli/internal/config"
)

// Logger defines the logging interface
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
}

// structuredLogger implements the Logger interface
type structuredLogger struct {
	output io.Writer
	level  LogLevel
	prefix string
}

// LogLevel represents logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// NewLogger creates a new logger instance
func NewLogger(cfg config.LoggingConfig) (Logger, error) {
	var output io.Writer = os.Stdout
	
	if cfg.Output == "file" && cfg.FilePath != "" {
		filePath := cfg.FilePath
		if !filepath.IsAbs(filePath) {
			execPath, err := os.Executable()
			if err != nil {
				return nil, fmt.Errorf("failed to get executable path: %w", err)
			}
			filePath = filepath.Join(filepath.Dir(execPath), filePath)
		}

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	return &structuredLogger{
		output: output,
		prefix: cfg.Format,
	}, nil
}

// Debug logs debug messages
func (l *structuredLogger) Debug(msg string, args ...interface{}) {
	l.log(DebugLevel, msg, args...)
}

// Info logs info messages
func (l *structuredLogger) Info(msg string, args ...interface{}) {
	l.log(InfoLevel, msg, args...)
}

// Warn logs warning messages
func (l *structuredLogger) Warn(msg string, args ...interface{}) {
	l.log(WarnLevel, msg, args...)
}

// Error logs error messages
func (l *structuredLogger) Error(msg string, args ...interface{}) {
	l.log(ErrorLevel, msg, args...)
}

// Fatal logs fatal messages and exits
func (l *structuredLogger) Fatal(msg string, args ...interface{}) {
	l.log(FatalLevel, msg, args...)
	os.Exit(1)
}

// log writes the log message
func (l *structuredLogger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := l.levelString(level)
	
	// Check if we have key-value pairs
	if len(args) > 0 && len(args)%2 == 0 {
		// Try to format as key-value pairs
		var hasKeyValuePairs = true
		for i := 0; i < len(args); i += 2 {
			if _, ok := args[i].(string); !ok {
				hasKeyValuePairs = false
				break
			}
		}
		
		if hasKeyValuePairs {
			// Format as structured log with key-value pairs
			fmt.Fprintf(l.output, "[%s] %s: %s", timestamp, levelStr, msg)
			for i := 0; i < len(args); i += 2 {
				fmt.Fprintf(l.output, " %s=%v", args[i], args[i+1])
			}
			fmt.Fprintln(l.output)
			return
		}
	}
	
	// Fall back to sprintf formatting
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}
	fmt.Fprintf(l.output, "[%s] %s: %s\n", timestamp, levelStr, formattedMsg)
}

// levelString converts LogLevel to string
func (l *structuredLogger) levelString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}