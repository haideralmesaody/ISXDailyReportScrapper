package pipeline

import (
	"time"
)

// Pipeline stage identifiers
const (
	StageIDScraping  = "scraping"
	StageIDProcessing = "processing"
	StageIDIndices   = "indices"
	StageIDAnalysis  = "analysis"
)

// Pipeline stage names
const (
	StageNameScraping  = "Data Collection"
	StageNameProcessing = "Data Processing"
	StageNameIndices   = "Index Extraction"
	StageNameAnalysis  = "Ticker Analysis"
)

// Context keys for pipeline state
const (
	ContextKeyFromDate      = "from_date"
	ContextKeyToDate        = "to_date"
	ContextKeyMode          = "mode"
	ContextKeyDownloadDir   = "download_dir"
	ContextKeyReportDir     = "report_dir"
	ContextKeyFilesFound    = "files_found"
	ContextKeyFilesProcessed = "files_processed"
	ContextKeyScraperSuccess = "scraper_success"
)

// Pipeline modes
const (
	ModeInitial     = "initial"
	ModeAccumulative = "accumulative"
	ModeFull        = "full"
)

// WebSocket event types - using frontend format
const (
	EventTypePipelineStatus   = "pipeline:status"
	EventTypePipelineProgress = "pipeline:progress"
	EventTypePipelineComplete = "pipeline:complete"
	EventTypePipelineError    = "pipeline:error"
	EventTypePipelineReset    = "pipeline:reset"
)

// Default timeouts
const (
	DefaultStageTimeout     = 30 * time.Minute
	DefaultScrapingTimeout  = 60 * time.Minute
	DefaultProcessingTimeout = 30 * time.Minute
	DefaultIndicesTimeout   = 10 * time.Minute
	DefaultAnalysisTimeout  = 5 * time.Minute
)

// ExecutionMode defines how stages are executed
type ExecutionMode string

const (
	ExecutionModeSequential ExecutionMode = "sequential"
	ExecutionModeParallel   ExecutionMode = "parallel"
)

// RetryConfig defines retry behavior for stages
type RetryConfig struct {
	MaxAttempts int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
	Multiplier  float64       `json:"multiplier"`
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// StageExecutionResult represents the result of a stage execution
type StageExecutionResult struct {
	StageID   string                 `json:"stage_id"`
	Success   bool                   `json:"success"`
	Output    string                 `json:"output,omitempty"`
	Error     error                  `json:"error,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PipelineRequest represents a request to execute a pipeline
type PipelineRequest struct {
	ID         string                 `json:"id"`
	Mode       string                 `json:"mode"`
	FromDate   string                 `json:"from_date,omitempty"`
	ToDate     string                 `json:"to_date,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// PipelineResponse represents the response from a pipeline execution
type PipelineResponse struct {
	ID       string                   `json:"id"`
	Status   PipelineStatus           `json:"status"`
	Duration time.Duration            `json:"duration"`
	Stages   map[string]*StageState   `json:"stages"`
	Error    string                   `json:"error,omitempty"`
}

// ProgressUpdate represents a progress update from a stage
type ProgressUpdate struct {
	StageID    string                 `json:"stage_id"`
	Progress   float64                `json:"progress"`
	Message    string                 `json:"message"`
	ETA        string                 `json:"eta,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// StageMetrics represents performance metrics for a stage
type StageMetrics struct {
	StageID        string        `json:"stage_id"`
	ExecutionCount int           `json:"execution_count"`
	SuccessCount   int           `json:"success_count"`
	FailureCount   int           `json:"failure_count"`
	AverageDuration time.Duration `json:"average_duration"`
	LastExecution  *time.Time    `json:"last_execution,omitempty"`
}