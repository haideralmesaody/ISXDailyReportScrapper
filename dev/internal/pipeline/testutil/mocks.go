package testutil

import (
	"context"
	"sync"
	"time"

	"isxcli/internal/pipeline"
)

// MockStage is a configurable mock implementation of the Stage interface
type MockStage struct {
	IDValue          string
	NameValue        string
	DependenciesValue []string
	
	// Configurable functions
	ExecuteFunc  func(ctx context.Context, state *pipeline.PipelineState) error
	ValidateFunc func(state *pipeline.PipelineState) error
	
	// Call tracking
	mu               sync.Mutex
	ExecuteCalls     int
	ExecuteArgs      []ExecuteCall
	ValidateCalls    int
	ValidateArgs     []ValidateCall
}

// ExecuteCall tracks arguments passed to Execute
type ExecuteCall struct {
	Ctx   context.Context
	State *pipeline.PipelineState
	Time  time.Time
}

// ValidateCall tracks arguments passed to Validate
type ValidateCall struct {
	State *pipeline.PipelineState
	Time  time.Time
}

// ID returns the stage ID
func (m *MockStage) ID() string {
	return m.IDValue
}

// Name returns the stage name
func (m *MockStage) Name() string {
	return m.NameValue
}

// GetDependencies returns the stage dependencies
func (m *MockStage) GetDependencies() []string {
	if m.DependenciesValue == nil {
		return []string{}
	}
	return m.DependenciesValue
}

// Execute runs the mock execute function
func (m *MockStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
	m.mu.Lock()
	m.ExecuteCalls++
	m.ExecuteArgs = append(m.ExecuteArgs, ExecuteCall{
		Ctx:   ctx,
		State: state,
		Time:  time.Now(),
	})
	m.mu.Unlock()
	
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, state)
	}
	return nil
}

// Validate runs the mock validate function
func (m *MockStage) Validate(state *pipeline.PipelineState) error {
	m.mu.Lock()
	m.ValidateCalls++
	m.ValidateArgs = append(m.ValidateArgs, ValidateCall{
		State: state,
		Time:  time.Now(),
	})
	m.mu.Unlock()
	
	if m.ValidateFunc != nil {
		return m.ValidateFunc(state)
	}
	return nil
}

// GetExecuteCalls returns the number of Execute calls
func (m *MockStage) GetExecuteCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ExecuteCalls
}

// GetValidateCalls returns the number of Validate calls
func (m *MockStage) GetValidateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ValidateCalls
}

// MockWebSocketHub captures WebSocket messages for testing
type MockWebSocketHub struct {
	mu       sync.Mutex
	Messages []WebSocketMessage
}

// WebSocketMessage represents a captured WebSocket message
type WebSocketMessage struct {
	EventType string
	Stage     string
	Status    string
	Metadata  interface{}
	Time      time.Time
}

// BroadcastUpdate captures WebSocket messages
func (m *MockWebSocketHub) BroadcastUpdate(eventType, stage, status string, metadata interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Messages = append(m.Messages, WebSocketMessage{
		EventType: eventType,
		Stage:     stage,
		Status:    status,
		Metadata:  metadata,
		Time:      time.Now(),
	})
}

// BroadcastRefresh captures refresh messages
func (m *MockWebSocketHub) BroadcastRefresh(source string, components []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Messages = append(m.Messages, WebSocketMessage{
		EventType: "refresh",
		Stage:     source,
		Metadata: map[string]interface{}{
			"components": components,
		},
		Time: time.Now(),
	})
}

// GetMessages returns all captured messages
func (m *MockWebSocketHub) GetMessages() []WebSocketMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	messages := make([]WebSocketMessage, len(m.Messages))
	copy(messages, m.Messages)
	return messages
}

// GetMessagesByType returns messages of a specific type
func (m *MockWebSocketHub) GetMessagesByType(eventType string) []WebSocketMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var filtered []WebSocketMessage
	for _, msg := range m.Messages {
		if msg.EventType == eventType {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// Clear removes all captured messages
func (m *MockWebSocketHub) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = nil
}

// MockLogger captures log messages for testing
type MockLogger struct {
	mu          sync.Mutex
	InfoLogs    []LogEntry
	ErrorLogs   []LogEntry
	WarningLogs []LogEntry
	DebugLogs   []LogEntry
}

// LogEntry represents a captured log entry
type LogEntry struct {
	Format string
	Args   []interface{}
	Time   time.Time
}

// Info captures info log messages
func (m *MockLogger) Info(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.InfoLogs = append(m.InfoLogs, LogEntry{
		Format: format,
		Args:   args,
		Time:   time.Now(),
	})
}

// Error captures error log messages
func (m *MockLogger) Error(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.ErrorLogs = append(m.ErrorLogs, LogEntry{
		Format: format,
		Args:   args,
		Time:   time.Now(),
	})
}

// Warn captures warning log messages
func (m *MockLogger) Warn(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.WarningLogs = append(m.WarningLogs, LogEntry{
		Format: format,
		Args:   args,
		Time:   time.Now(),
	})
}

// Debug captures debug log messages
func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.DebugLogs = append(m.DebugLogs, LogEntry{
		Format: format,
		Args:   args,
		Time:   time.Now(),
	})
}

// GetInfoLogs returns all info logs
func (m *MockLogger) GetInfoLogs() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	logs := make([]LogEntry, len(m.InfoLogs))
	copy(logs, m.InfoLogs)
	return logs
}

// GetErrorLogs returns all error logs
func (m *MockLogger) GetErrorLogs() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	logs := make([]LogEntry, len(m.ErrorLogs))
	copy(logs, m.ErrorLogs)
	return logs
}

// GetWarningLogs returns all warning logs
func (m *MockLogger) GetWarningLogs() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	logs := make([]LogEntry, len(m.WarningLogs))
	copy(logs, m.WarningLogs)
	return logs
}

// Clear removes all captured logs
func (m *MockLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.InfoLogs = nil
	m.ErrorLogs = nil
	m.WarningLogs = nil
	m.DebugLogs = nil
}