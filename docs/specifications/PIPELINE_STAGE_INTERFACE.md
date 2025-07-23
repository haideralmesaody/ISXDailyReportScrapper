# Pipeline Stage Interface Specification

## Overview

This document specifies the interface and contracts for pipeline stages in the ISX Daily Reports Scrapper. All pipeline stages must implement this interface to be compatible with the Pipeline Manager.

## Stage Interface Definition

```go
package pipeline

import (
    "context"
)

// Stage represents a single stage in the pipeline
type Stage interface {
    // ID returns the unique identifier for this stage
    ID() string
    
    // Name returns the human-readable name for this stage
    Name() string
    
    // Dependencies returns the IDs of stages that must complete before this stage
    Dependencies() []string
    
    // Execute runs the stage logic
    Execute(ctx context.Context, state *PipelineState) error
    
    // Validate checks if the stage can be executed
    Validate(state *PipelineState) error
}
```

## Interface Methods

### ID() string

**Purpose**: Returns a unique identifier for the stage.

**Requirements**:
- Must be unique across all stages in the pipeline
- Should use lowercase with hyphens (e.g., "data-processing")
- Must not change between executions
- Used for dependency references

**Example**:
```go
func (s *ScrapingStage) ID() string {
    return "scraping"
}
```

### Name() string

**Purpose**: Returns a human-readable name for display purposes.

**Requirements**:
- Should be descriptive and user-friendly
- Can contain spaces and proper capitalization
- Used in UI displays and log messages

**Example**:
```go
func (s *ScrapingStage) Name() string {
    return "Data Collection"
}
```

### Dependencies() []string

**Purpose**: Declares which stages must complete successfully before this stage can run.

**Requirements**:
- Return empty slice if no dependencies
- All referenced stage IDs must exist
- No circular dependencies allowed
- Pipeline Manager enforces execution order

**Example**:
```go
func (s *ProcessingStage) Dependencies() []string {
    return []string{"scraping"}  // Must run after scraping
}
```

### Execute(ctx context.Context, state *PipelineState) error

**Purpose**: Contains the main logic for the stage.

**Requirements**:
- Must respect context cancellation
- Should update progress regularly via state
- Return nil on success
- Return appropriate error type on failure
- Must be idempotent when possible

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `state`: Pipeline state for configuration and progress

**Return Values**:
- `nil`: Stage completed successfully
- `error`: Stage failed (see Error Types section)

**Example**:
```go
func (s *ScrapingStage) Execute(ctx context.Context, state *PipelineState) error {
    stageState := state.GetStage(s.ID())
    
    // Check for cancellation
    select {
    case <-ctx.Done():
        return NewCancellationError(s.ID())
    default:
    }
    
    // Update progress
    stageState.UpdateProgress(10, "Starting data collection...")
    
    // Execute stage logic
    err := s.collectData(ctx, state)
    if err != nil {
        return NewExecutionError(s.ID(), err, false)
    }
    
    // Success
    stageState.UpdateProgress(100, "Data collection completed")
    return nil
}
```

### Validate(state *PipelineState) error

**Purpose**: Checks if the stage can be executed with the current state.

**Requirements**:
- Called before Execute()
- Should be fast (no heavy operations)
- Check all prerequisites
- Return nil if stage can proceed
- Return error if stage cannot run

**Common Validations**:
- Required configuration present
- Input files exist
- Output directories writable
- External dependencies available
- License valid (if applicable)

**Example**:
```go
func (s *ProcessingStage) Validate(state *PipelineState) error {
    // Check configuration
    downloadDir, ok := state.GetConfig(ContextKeyDownloadDir).(string)
    if !ok || downloadDir == "" {
        return NewValidationError(s.ID(), "download directory not configured")
    }
    
    // Check input files exist
    files, err := filepath.Glob(filepath.Join(downloadDir, "*.xlsx"))
    if err != nil || len(files) == 0 {
        return NewValidationError(s.ID(), "no Excel files found to process")
    }
    
    return nil
}
```

## BaseStage Implementation

A `BaseStage` struct is provided for common functionality:

```go
type BaseStage struct {
    id           string
    name         string
    dependencies []string
}

func NewBaseStage(id, name string, dependencies []string) BaseStage {
    if dependencies == nil {
        dependencies = []string{}
    }
    return BaseStage{
        id:           id,
        name:         name,
        dependencies: dependencies,
    }
}

func (b BaseStage) ID() string           { return b.id }
func (b BaseStage) Name() string         { return b.name }
func (b BaseStage) Dependencies() []string { return b.dependencies }

// Default validate (always passes)
func (b BaseStage) Validate(state *PipelineState) error {
    return nil
}
```

## Stage State Management

### Accessing Stage State

```go
func (s *MyStage) Execute(ctx context.Context, state *PipelineState) error {
    // Get this stage's state
    stageState := state.GetStage(s.ID())
    
    // Update progress (0-100)
    stageState.UpdateProgress(50, "Halfway complete...")
    
    // Set metadata
    stageState.Metadata["files_processed"] = 10
    stageState.Metadata["current_file"] = "data.xlsx"
    
    return nil
}
```

### StageState Structure

```go
type StageState struct {
    ID        string                 // Stage ID
    Name      string                 // Stage name
    Status    StageStatus           // Current status
    Progress  float64               // 0-100
    Message   string                // Current status message
    Error     error                 // Error if failed
    StartTime time.Time             // When stage started
    EndTime   time.Time             // When stage ended
    Metadata  map[string]interface{} // Custom metadata
}
```

### Stage Status Values

```go
type StageStatus string

const (
    StageStatusPending   StageStatus = "pending"   // Not yet started
    StageStatusRunning   StageStatus = "running"   // Currently executing
    StageStatusCompleted StageStatus = "completed" // Finished successfully
    StageStatusFailed    StageStatus = "failed"    // Failed with error
    StageStatusSkipped   StageStatus = "skipped"   // Skipped (conditional)
    StageStatusCancelled StageStatus = "cancelled" // Cancelled by user/timeout
)
```

## Configuration Access

### Reading Configuration

```go
func (s *MyStage) Execute(ctx context.Context, state *PipelineState) error {
    // Type assertion with default
    downloadDir, ok := state.GetConfig(ContextKeyDownloadDir).(string)
    if !ok || downloadDir == "" {
        downloadDir = "data/downloads"
    }
    
    // Direct access (be careful with nil)
    mode, _ := state.GetConfig(ContextKeyMode).(string)
    
    // Check specific mode
    if mode == ModeFull {
        // Full processing mode
    }
    
    return nil
}
```

### Standard Configuration Keys

```go
const (
    ContextKeyFromDate    = "from_date"
    ContextKeyToDate      = "to_date"
    ContextKeyMode        = "mode"
    ContextKeyDownloadDir = "download_dir"
    ContextKeyReportDir   = "report_dir"
)
```

## Error Types

### ValidationError

Used when stage cannot run due to missing prerequisites:

```go
type ValidationError struct {
    StageID string
    Message string
}

func NewValidationError(stageID, message string) error {
    return &ValidationError{
        StageID: stageID,
        Message: message,
    }
}
```

### ExecutionError

Used when stage fails during execution:

```go
type ExecutionError struct {
    StageID    string
    Err        error
    Retryable  bool
}

func NewExecutionError(stageID string, err error, retryable bool) error {
    return &ExecutionError{
        StageID:   stageID,
        Err:       err,
        Retryable: retryable,
    }
}
```

### CancellationError

Used when stage is cancelled:

```go
type CancellationError struct {
    StageID string
}

func NewCancellationError(stageID string) error {
    return &CancellationError{
        StageID: stageID,
    }
}
```

## Progress Reporting

### Progress Updates

Progress should be reported regularly during execution:

```go
func (s *MyStage) processFiles(ctx context.Context, stageState *StageState, files []string) error {
    total := len(files)
    
    for i, file := range files {
        // Check cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // Update progress
        progress := float64(i) / float64(total) * 100
        stageState.UpdateProgress(progress, fmt.Sprintf("Processing %s (%d/%d)", 
            filepath.Base(file), i+1, total))
        
        // Process file
        if err := s.processFile(file); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Progress Guidelines

1. **Frequency**: Update at least every 5 seconds for long operations
2. **Granularity**: Provide meaningful progress increments
3. **Messages**: Include current operation details
4. **Accuracy**: Avoid jumping backwards or over 100%

## Context Sharing

### Setting Context Data

Data can be shared between stages via context:

```go
// In scraping stage
func (s *ScrapingStage) Execute(ctx context.Context, state *PipelineState) error {
    filesDownloaded := s.downloadFiles(ctx)
    
    // Share with later stages
    state.SetContext("files_downloaded", filesDownloaded)
    state.SetContext("download_timestamp", time.Now())
    
    return nil
}
```

### Reading Context Data

```go
// In processing stage
func (s *ProcessingStage) Execute(ctx context.Context, state *PipelineState) error {
    // Read data from previous stage
    filesDownloaded, _ := state.GetContext("files_downloaded").(int)
    timestamp, _ := state.GetContext("download_timestamp").(time.Time)
    
    log.Printf("Processing %d files downloaded at %v", filesDownloaded, timestamp)
    
    return nil
}
```

## Best Practices

### 1. Idempotency

Stages should be idempotent when possible:

```go
func (s *MyStage) Execute(ctx context.Context, state *PipelineState) error {
    outputFile := s.getOutputPath()
    
    // Check if already processed
    if s.isAlreadyProcessed(outputFile) {
        stageState := state.GetStage(s.ID())
        stageState.UpdateProgress(100, "Already processed, skipping...")
        return nil
    }
    
    // Process normally
    return s.process(ctx, state)
}
```

### 2. Resource Cleanup

Always clean up resources:

```go
func (s *MyStage) Execute(ctx context.Context, state *PipelineState) error {
    // Acquire resource
    conn, err := s.connect()
    if err != nil {
        return err
    }
    defer conn.Close() // Always cleanup
    
    // Use resource
    return s.useConnection(conn)
}
```

### 3. Cancellation Handling

Check for cancellation in loops:

```go
func (s *MyStage) Execute(ctx context.Context, state *PipelineState) error {
    for _, item := range s.getItems() {
        select {
        case <-ctx.Done():
            return NewCancellationError(s.ID())
        default:
            if err := s.processItem(item); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### 4. Error Context

Provide helpful error messages:

```go
func (s *MyStage) Execute(ctx context.Context, state *PipelineState) error {
    file := s.getInputFile()
    
    data, err := os.ReadFile(file)
    if err != nil {
        return NewExecutionError(s.ID(), 
            fmt.Errorf("failed to read input file %s: %w", file, err), 
            false)
    }
    
    return nil
}
```

## Testing Stages

### Unit Testing

```go
func TestMyStage_Execute(t *testing.T) {
    // Create test state
    state := &PipelineState{
        ID:      "test-pipeline",
        Stages:  make(map[string]*StageState),
        Context: make(map[string]interface{}),
        Config: map[string]interface{}{
            ContextKeyReportDir: "testdata",
        },
    }
    
    // Add stage state
    state.Stages["mystage"] = &StageState{
        ID:       "mystage",
        Name:     "My Stage",
        Status:   StageStatusPending,
        Metadata: make(map[string]interface{}),
    }
    
    // Create stage
    stage := NewMyStage(".", testLogger)
    
    // Execute
    err := stage.Execute(context.Background(), state)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, StageStatusCompleted, state.Stages["mystage"].Status)
    assert.Equal(t, 100.0, state.Stages["mystage"].Progress)
}
```

### Mocking External Commands

```go
type MockCommandExecutor struct {
    ExitCode int
    Output   string
}

func (m *MockCommandExecutor) Execute(ctx context.Context, name string, args ...string) error {
    if m.ExitCode != 0 {
        return fmt.Errorf("command failed with exit code %d", m.ExitCode)
    }
    return nil
}

func TestStageWithCommand(t *testing.T) {
    stage := NewCommandStage()
    stage.executor = &MockCommandExecutor{
        ExitCode: 0,
        Output:   "Success",
    }
    
    err := stage.Execute(context.Background(), testState)
    assert.NoError(t, err)
}
```

## Summary

The Stage Interface provides a clean contract for pipeline components:

1. **Clear Responsibilities**: Each stage has a single purpose
2. **Dependency Management**: Automatic ordering based on dependencies
3. **Progress Tracking**: Built-in progress reporting
4. **Error Handling**: Consistent error types and recovery
5. **Context Sharing**: Easy data passing between stages
6. **Testability**: Clean interfaces enable easy testing

By following this specification, stages will integrate seamlessly with the Pipeline Manager and provide a consistent experience for users and developers.