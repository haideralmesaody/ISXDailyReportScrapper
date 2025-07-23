package common

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// LogCategory represents different logging categories
const (
	CategoryPipeline  = "PIPELINE"
	CategoryWebSocket = "WEBSOCKET"
	CategoryLicense   = "LICENSE"
	CategoryFile      = "FILE"
	CategoryData      = "DATA"
	CategoryHTTP      = "HTTP"
	CategorySystem    = "SYSTEM"
)

// Logger provides conditional logging based on build mode
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	component  string
	categories map[string]bool
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return NewLoggerWithComponent("")
}

// NewLoggerWithComponent creates a logger for a specific component
func NewLoggerWithComponent(component string) *Logger {
	level := INFO
	
	// Enable debug logging only if ISX_DEBUG environment variable is set
	// This allows debugging in development without exposing sensitive data in production
	if os.Getenv("ISX_DEBUG") == "true" {
		level = DEBUG
	}
	
	// Parse debug categories if specified
	categories := make(map[string]bool)
	if debugCategories := os.Getenv("ISX_DEBUG_CATEGORIES"); debugCategories != "" {
		for _, cat := range strings.Split(debugCategories, ",") {
			categories[strings.TrimSpace(cat)] = true
		}
	}
	
	// Use microsecond precision for detailed timing
	logger := log.New(os.Stdout, "", 0) // We'll format our own timestamps
	
	return &Logger{
		level:      level,
		logger:     logger,
		component:  component,
		categories: categories,
	}
}

// NewLoggerWithLevel creates a logger with a specific level
func NewLoggerWithLevel(level LogLevel) *Logger {
	categories := make(map[string]bool)
	if debugCategories := os.Getenv("ISX_DEBUG_CATEGORIES"); debugCategories != "" {
		for _, cat := range strings.Split(debugCategories, ",") {
			categories[strings.TrimSpace(cat)] = true
		}
	}
	
	return &Logger{
		level:      level,
		logger:     log.New(os.Stdout, "", 0),
		component:  "",
		categories: categories,
	}
}

// formatMessage creates a formatted log message with timestamp and context
func (l *Logger) formatMessage(level, category, format string, v ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, v...)
	
	// Get caller information for debug mode
	callerInfo := ""
	if l.level <= DEBUG {
		_, file, line, ok := runtime.Caller(3)
		if ok {
			// Extract just the filename
			parts := strings.Split(file, "/")
			if len(parts) > 0 {
				file = parts[len(parts)-1]
			}
			callerInfo = fmt.Sprintf(" [%s:%d]", file, line)
		}
	}
	
	// Build the log message
	component := ""
	if l.component != "" {
		component = fmt.Sprintf(" [%s]", l.component)
	}
	
	categoryStr := ""
	if category != "" {
		categoryStr = fmt.Sprintf(" [%s]", category)
	}
	
	return fmt.Sprintf("[%s] [%s]%s%s%s %s", timestamp, level, categoryStr, component, callerInfo, message)
}

// shouldLog checks if a message should be logged based on level and category
func (l *Logger) shouldLog(level LogLevel, category string) bool {
	if l.level > level {
		return false
	}
	
	// If no categories specified, log everything at the appropriate level
	if len(l.categories) == 0 {
		return true
	}
	
	// If categories are specified, only log matching categories
	return category == "" || l.categories[category]
}

// Debug logs a debug message (only in debug mode)
func (l *Logger) Debug(format string, v ...interface{}) {
	l.DebugCategory("", format, v...)
}

// DebugCategory logs a debug message with a category
func (l *Logger) DebugCategory(category, format string, v ...interface{}) {
	if l.shouldLog(DEBUG, category) {
		l.logger.Print(l.formatMessage("DEBUG", category, format, v...))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.InfoCategory("", format, v...)
}

// InfoCategory logs an info message with a category
func (l *Logger) InfoCategory(category, format string, v ...interface{}) {
	if l.shouldLog(INFO, category) {
		l.logger.Print(l.formatMessage("INFO", category, format, v...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.WarnCategory("", format, v...)
}

// WarnCategory logs a warning message with a category
func (l *Logger) WarnCategory(category, format string, v ...interface{}) {
	if l.shouldLog(WARN, category) {
		l.logger.Print(l.formatMessage("WARN", category, format, v...))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.ErrorCategory("", format, v...)
}

// ErrorCategory logs an error message with a category
func (l *Logger) ErrorCategory(category, format string, v ...interface{}) {
	if l.shouldLog(ERROR, category) {
		l.logger.Print(l.formatMessage("ERROR", category, format, v...))
	}
}

// Fatal logs an error message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.logger.Fatalf("[FATAL] "+format, v...)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// IsDebugEnabled returns true if debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.level <= DEBUG
}

// Printf provides compatibility with standard log.Printf
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Print provides compatibility with standard log.Print
func (l *Logger) Print(v ...interface{}) {
	l.Info(fmt.Sprint(v...))
}

// Println provides compatibility with standard log.Println
func (l *Logger) Println(v ...interface{}) {
	l.Info(fmt.Sprintln(v...))
}

// LogVersion logs the application version information
func (l *Logger) LogVersion() {
	l.InfoCategory(CategorySystem, "=====================================")
	l.InfoCategory(CategorySystem, GetFullVersionString())
	l.InfoCategory(CategorySystem, "Data Format: %s", DataFormatVersion)
	l.InfoCategory(CategorySystem, "API Version: %s", APIVersion)
	if IsAlpha() {
		l.WarnCategory(CategorySystem, "This is an ALPHA version - not for production use")
	} else if IsBeta() {
		l.WarnCategory(CategorySystem, "This is a BETA version - use with caution")
	}
	l.InfoCategory(CategorySystem, "=====================================")
}

// LogTiming logs the duration of an operation
func (l *Logger) LogTiming(category, operation string, start time.Time) {
	duration := time.Since(start)
	l.DebugCategory(category, "%s completed in %v", operation, duration)
}

// LogFileOperation logs file operations with details
func (l *Logger) LogFileOperation(operation, path string, size int64) {
	l.DebugCategory(CategoryFile, "%s: %s (size: %d bytes)", operation, path, size)
}

// LogDecision logs a decision point in the code
func (l *Logger) LogDecision(category, decision, reason string) {
	l.DebugCategory(category, "Decision: %s - Reason: %s", decision, reason)
}

// LogStart logs the start of a major operation
func (l *Logger) LogStart(category, operation string) {
	l.InfoCategory(category, "Starting %s", operation)
}

// LogComplete logs the completion of a major operation
func (l *Logger) LogComplete(category, operation string) {
	l.InfoCategory(category, "Completed %s", operation)
}

// LogProgress logs progress of a long-running operation
func (l *Logger) LogProgress(category, operation string, current, total int) {
	percentage := float64(current) / float64(total) * 100
	l.DebugCategory(category, "%s progress: %d/%d (%.1f%%)", operation, current, total, percentage)
}