package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// MockCommandExecutor mocks external command execution
type MockCommandExecutor struct {
	mu              sync.Mutex
	Commands        []CommandCall
	DefaultExitCode int
	DefaultOutput   string
	CommandOutputs  map[string]CommandOutput // Maps command to output
}

// CommandCall records a command execution
type CommandCall struct {
	Name string
	Args []string
}

// CommandOutput defines the output for a specific command
type CommandOutput struct {
	Output   string
	ExitCode int
	Error    error
}

// Execute mocks command execution
func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Record the command
	m.Commands = append(m.Commands, CommandCall{
		Name: name,
		Args: args,
	})
	
	// Check for specific command output
	cmdKey := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if output, ok := m.CommandOutputs[cmdKey]; ok {
		if output.Error != nil {
			return output.Error
		}
		if output.ExitCode != 0 {
			return fmt.Errorf("command failed with exit code %d", output.ExitCode)
		}
		return nil
	}
	
	// Use default behavior
	if m.DefaultExitCode != 0 {
		return fmt.Errorf("command failed with exit code %d", m.DefaultExitCode)
	}
	
	return nil
}

// GetCommands returns all executed commands
func (m *MockCommandExecutor) GetCommands() []CommandCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	commands := make([]CommandCall, len(m.Commands))
	copy(commands, m.Commands)
	return commands
}

// SetCommandOutput sets the output for a specific command
func (m *MockCommandExecutor) SetCommandOutput(command string, output CommandOutput) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.CommandOutputs == nil {
		m.CommandOutputs = make(map[string]CommandOutput)
	}
	m.CommandOutputs[command] = output
}

// Clear resets the executor
func (m *MockCommandExecutor) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Commands = nil
	m.CommandOutputs = nil
}

// MockProgressReporter mocks progress reporting
type MockProgressReporter struct {
	mu       sync.Mutex
	Updates  []ProgressUpdate
}

// ProgressUpdate records a progress update
type ProgressUpdate struct {
	StageID  string
	Progress float64
	Message  string
}

// UpdateProgress records a progress update
func (m *MockProgressReporter) UpdateProgress(stageID string, progress float64, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Updates = append(m.Updates, ProgressUpdate{
		StageID:  stageID,
		Progress: progress,
		Message:  message,
	})
}

// GetUpdates returns all progress updates
func (m *MockProgressReporter) GetUpdates() []ProgressUpdate {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	updates := make([]ProgressUpdate, len(m.Updates))
	copy(updates, m.Updates)
	return updates
}

// GetLastUpdate returns the last progress update for a stage
func (m *MockProgressReporter) GetLastUpdate(stageID string) *ProgressUpdate {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for i := len(m.Updates) - 1; i >= 0; i-- {
		if m.Updates[i].StageID == stageID {
			update := m.Updates[i]
			return &update
		}
	}
	return nil
}

// Clear resets the reporter
func (m *MockProgressReporter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Updates = nil
}

// MockMetricsCollector mocks metrics collection
type MockMetricsCollector struct {
	mu      sync.Mutex
	Metrics map[string]interface{}
}

// RecordMetric records a metric
func (m *MockMetricsCollector) RecordMetric(name string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.Metrics == nil {
		m.Metrics = make(map[string]interface{})
	}
	m.Metrics[name] = value
}

// GetMetric retrieves a metric
func (m *MockMetricsCollector) GetMetric(name string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	val, ok := m.Metrics[name]
	return val, ok
}

// Clear resets the collector
func (m *MockMetricsCollector) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.Metrics = nil
}