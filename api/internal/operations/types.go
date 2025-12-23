package operations

import (
	"time"
)

// operation Step identifiers
const (
	StageIDScraping   = "scraping"
	StageIDProcessing = "processing"
	StageIDIndices    = "indices"
	StageIDLiquidity  = "liquidity"
)

// Stage order mapping for dependency validation
var stageOrderMap = map[string]int{
	StageIDScraping:   1,
	StageIDProcessing: 2,
	StageIDIndices:    3,
	StageIDLiquidity:  4,
}

var orderStageMap = map[int]string{
	1: StageIDScraping,
	2: StageIDProcessing,
	3: StageIDIndices,
	4: StageIDLiquidity,
}

// resolveStageOrder returns the execution order for a stage
func resolveStageOrder(stageID string) int {
	if order, exists := stageOrderMap[stageID]; exists {
		return order
	}
	return 0 // Unknown stage
}

// stageIDForOrder returns the stage ID for a given order
func stageIDForOrder(order int) string {
	if stageID, exists := orderStageMap[order]; exists {
		return stageID
	}
	return "" // Unknown order
}

// operation Step names
const (
	StageNameScraping   = "Data Collection"
	StageNameProcessing = "Data Processing"
	StageNameIndices    = "Index Extraction"
	StageNameLiquidity  = "Liquidity Calculation"
)

// Stage timeout constants (in seconds)
const (
	DefaultScrapingTimeout   = 300 * time.Second // 5 minutes
	DefaultProcessingTimeout = 180 * time.Second // 3 minutes
	DefaultIndicesTimeout    = 120 * time.Second // 2 minutes
	DefaultLiquidityTimeout  = 240 * time.Second // 4 minutes
	DefaultStageTimeout      = 600 * time.Second // 10 minutes
)

// Context keys for operation state
const (
	ContextKeyFromDate       = "from_date"
	ContextKeyToDate         = "to_date"
	ContextKeyMode           = "mode"
	ContextKeyDownloadDir    = "download_dir"
	ContextKeyReportDir      = "report_dir"
	ContextKeyFilesFound     = "files_found"
	ContextKeyFilesProcessed = "files_processed"
	ContextKeyScraperSuccess = "scraper_success"
)

// operation modes
const (
	ModeInitial      = "initial"
	ModeAccumulative = "accumulative"
	ModeFull         = "full"
	ModeSkip         = "skip"
)

// WebSocket event types - using frontend format
//
// IMPORTANT: The application now uses EventTypeOperationSnapshot exclusively.
// The other event types are DEPRECATED and kept only for backward compatibility.

const (
	// DEPRECATED: Use EventTypeOperationSnapshot instead
	EventTypeOperationStart    = "operation_start"
	EventTypeOperationProgress = "operation_progress"
	EventTypeOperationComplete = "operation_complete"
	EventTypeOperationError    = "operation_error"
	EventTypeOperationFailed   = "operation_failed"

	// DEPRECATED: Use EventTypeOperationSnapshot instead
	EventTypeStageStart    = "stage_start"
	EventTypeStageProgress = "stage_progress"
	EventTypeStageComplete = "stage_complete"
	EventTypeStageError    = "stage_error"
	EventTypeStageFailed   = "stage_failed"

	// DEPRECATED: Use EventTypeOperationSnapshot instead
	EventTypePhaseStart    = "phase_start"
	EventTypePhaseProgress = "phase_progress"
	EventTypePhaseComplete = "phase_complete"
	EventTypePhaseError    = "phase_error"
	EventTypePhaseFailed   = "phase_failed"

	// CURRENT: Use this event type for all operation communications
	EventTypeOperationSnapshot = "operation:snapshot"
	EventTypeOperationDelta    = "operation:delta"
	EventTypeOperationStatus   = "operation_status"
)

// Status values
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// ============================================================================
// Core Types
// ============================================================================

// Operation represents a complete data processing operation
type Operation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Stages      []StageStatus          `json:"stages,omitempty"`
}

// StageStatus represents the status of an operation stage
type StageStatus struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	SubStatus   *SubStatus             `json:"sub_status,omitempty"`
	Processed   int                    `json:"processed"`
	Total       int                    `json:"total"`
	CurrentItem string                 `json:"current_item,omitempty"`
}

// SubStatus represents nested status information (e.g., files in a stage)
type SubStatus struct {
	Type      string        `json:"type"` // "files", "records", "items"
	Processed int           `json:"processed"`
	Total     int           `json:"total"`
	Progress  float64       `json:"progress"`
	Current   string        `json:"current"`
	Error     string        `json:"error"`
	Items     []StatusItem  `json:"items,omitempty"`
	StartTime *time.Time    `json:"start_time,omitempty"`
	EndTime   *time.Time    `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// StatusItem represents individual item status within a sub-status
type StatusItem struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Progress  float64                `json:"progress"`
	Size      int64                  `json:"size"`
	Processed int64                  `json:"processed"`
	Error     string                 `json:"error,omitempty"`
	StartTime *time.Time             `json:"start_time,omitempty"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Phase represents a phase within a stage (e.g., download, process, upload)
type Phase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Processed   int                    `json:"processed"`
	Total       int                    `json:"total"`
	CurrentItem string                 `json:"current_item,omitempty"`
}

// OperationProgress represents operation progress data
type OperationProgress struct {
	OperationID string      `json:"operation_id"`
	StageID     string      `json:"stage_id,omitempty"`
	PhaseID     string      `json:"phase_id,omitempty"`
	Progress    float64     `json:"progress"`
	Status      string      `json:"status"`
	Message     string      `json:"message,omitempty"`
	Error       string      `json:"error,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// OperationLog represents a log entry for an operation
type OperationLog struct {
	ID          string                 `json:"id"`
	OperationID string                 `json:"operation_id"`
	StageID     string                 `json:"stage_id,omitempty"`
	PhaseID     string                 `json:"phase_id,omitempty"`
	Level       string                 `json:"level"` // "info", "warn", "error", "debug"
	Message     string                 `json:"message"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// OperationConfig represents configuration for an operation
type OperationConfig struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// OperationFilter represents filters for querying operations
type OperationFilter struct {
	Type     string     `json:"type,omitempty"`
	Status   string     `json:"status,omitempty"`
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	Offset   int        `json:"offset,omitempty"`
}

// OperationList represents a paginated list of operations
type OperationList struct {
	Operations []Operation `json:"operations"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

// OperationStats represents statistics for operations
type OperationStats struct {
	Total       int                    `json:"total"`
	ByType      map[string]int         `json:"by_type"`
	ByStatus    map[string]int         `json:"by_status"`
	RecentStats map[string]interface{} `json:"recent_stats"`
	LastUpdated time.Time              `json:"last_updated"`
}

// StageStats represents statistics for stages
type StageStats struct {
	StageID         string        `json:"stage_id"`
	ExecutionCount  int           `json:"execution_count"`
	SuccessCount    int           `json:"success_count"`
	FailureCount    int           `json:"failure_count"`
	AverageDuration time.Duration `json:"average_duration"`
	LastExecution   *time.Time    `json:"last_execution,omitempty"`
}

// OperationType represents an available operation type
type OperationType struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Dependencies []string              `json:"dependencies"`
	CanRunAlone  bool                  `json:"can_run_alone"`
	Parameters   []ParameterDefinition `json:"parameters"`
}

// ============================================================================
// Missing Type Definitions
// ============================================================================

// ExecutionMode represents how operations are executed
type ExecutionMode int

const (
	ExecutionModeSequential ExecutionMode = iota
	ExecutionModeParallel
)

// DataType constants for stage input/output types
const (
	DataTypeExcelFiles       string = "excel_files"
	DataTypeCSVFiles         string = "csv_files"
	DataTypeProcessedData    string = "processed_data"
	DataTypeIndexFiles       string = "index_files"
	DataTypeLiquidityMetrics string = "liquidity_metrics"
	DataTypeLiquidityScores  string = "liquidity_scores"
	DataTypeMarketAnalysis   string = "market_analysis"
)

// DataRequirement represents input data requirements for a stage
type DataRequirement struct {
	Type     string `json:"type"`      // e.g., "excel_files", "csv_files", "processed_data"
	Source   string `json:"source"`    // e.g., "scraping_stage", "processing_stage"
	Optional bool   `json:"optional"`  // Whether this input is optional
	Required bool   `json:"required"`  // Whether this input is required
	MinCount int    `json:"min_count"` // Minimum number of items required
	Location string `json:"location"`  // File location or directory path
}

// DataOutput represents what data a stage produces
type DataOutput struct {
	Type        string `json:"type"`         // e.g., "excel_files", "csv_files", "indicators_data"
	FilePattern string `json:"file_pattern"` // File pattern or wildcard
	Location    string `json:"location"`     // File location or directory path
	Required    bool   `json:"required"`     // Whether this output is required
	MinCount    int    `json:"min_count"`    // Minimum number of items required
}

// ParameterDefinition represents a configurable parameter for operations
type ParameterDefinition struct {
	Name         string      `json:"name"`          // Parameter name (e.g., "from_date", "to_date")
	Type         string      `json:"type"`          // Data type (e.g., "string", "date", "boolean")
	Description  string      `json:"description"`   // Human-readable description
	Required     bool        `json:"required"`      // Whether this parameter is required
	DefaultValue interface{} `json:"default_value"` // Default value if any
	Options      []string    `json:"options"`       // Valid options for enum parameters
}

// ExecutionMetrics represents runtime execution metrics
type ExecutionMetrics struct {
	// Execution statistics
	TotalOperations     int `json:"total_operations"`
	RunningOperations   int `json:"running_operations"`
	CompletedOperations int `json:"completed_operations"`
	FailedOperations    int `json:"failed_operations"`

	// Additional fields expected by code
	TotalExecutions  int       `json:"total_executions"`
	ActiveExecutions int       `json:"active_executions"`
	LastUpdated      time.Time `json:"last_updated"`

	// Execution mode usage
	ExecutionModes map[ExecutionMode]int `json:"execution_modes"`

	// Performance metrics
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastExecutionTime    *time.Time    `json:"last_execution_time,omitempty"`

	// Stage-specific metrics
	StageMetrics map[string]StageMetric `json:"stage_metrics"`
}

// StageMetric represents metrics for a specific stage
type StageMetric struct {
	Executions      int           `json:"executions"`
	Successes       int           `json:"successes"`
	Failures        int           `json:"failures"`
	AverageDuration time.Duration `json:"average_duration"`
	LastExecution   *time.Time    `json:"last_execution,omitempty"`
}

// ============================================================================
// NOTE: Complex orchestrator system removed for simplicity and reliability
// The factory-based registration in stages.go is used instead.
// ============================================================================

// ============================================================================
// Additional Missing Types
// ============================================================================

// RetryConfig represents retry configuration for operations
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	MaxAttempts     int           `json:"max_attempts,omitempty"` // Legacy alias for compatibility
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	Multiplier      float64       `json:"multiplier"`
	Jitter          bool          `json:"jitter"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// NewRetryConfig returns the default retry configuration
func NewRetryConfig() *RetryConfig {
	cfg := &RetryConfig{
		MaxRetries:      3,
		MaxAttempts:     3,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
		RetryableErrors: []string{"connection_error", "timeout", "temporary_failure"},
	}
	return cfg
}

// SyncLegacyFields keeps MaxRetries/MaxAttempts mirrored for backwards compatibility.
func (r *RetryConfig) SyncLegacyFields() {
	if r == nil {
		return
	}
	if r.MaxRetries == 0 && r.MaxAttempts != 0 {
		r.MaxRetries = r.MaxAttempts
	} else if r.MaxAttempts == 0 && r.MaxRetries != 0 {
		r.MaxAttempts = r.MaxRetries
	}
}

// AttemptLimit returns the configured retry limit, honoring both legacy aliases.
func (r RetryConfig) AttemptLimit() int {
	if r.MaxRetries != 0 {
		return r.MaxRetries
	}
	return r.MaxAttempts
}

// OperationRequest represents a request to execute an operation
type OperationRequest struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Options    map[string]interface{} `json:"options,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	Mode       string                 `json:"mode,omitempty"`
	FromDate   string                 `json:"from_date,omitempty"`
	ToDate     string                 `json:"to_date,omitempty"`
}

// OperationResponse represents the response from an operation
type OperationResponse struct {
	ID          string                 `json:"id"`
	OperationID string                 `json:"operation_id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Steps       []interface{}          `json:"steps,omitempty"`
}

// StageDescriptor describes a stage
type StageDescriptor struct {
	StageID      string                 `json:"stage_id"`
	StageName    string                 `json:"stage_name"`
	Description  string                 `json:"description"`
	Dependencies []string               `json:"dependencies"`
	Parameters   map[string]interface{} `json:"parameters"`
	Required     bool                   `json:"required"`
	Timeout      time.Duration          `json:"timeout"`
}
