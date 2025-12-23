package domain

import (
	"time"
)

// Operation represents a complete data processing workflow consisting of multiple steps.
// This package has been renamed from pipeline to operations for business-friendly terminology.

// Operation represents a data processing operation
type Operation struct {
	ID          string                 `json:"id" db:"id" validate:"required,uuid"`
	Name        string                 `json:"name" db:"name" validate:"required,min=3,max=100"`
	Type        OperationType          `json:"type" db:"type" validate:"required,oneof=scraping processing indexing liquidity"`
	Status      OperationStatus         `json:"status" db:"status"`
	Config      OperationConfig         `json:"config" db:"config"`
	Steps       []Step                 `json:"steps" db:"steps"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Metrics     OperationMetrics       `json:"metrics,omitempty" db:"metrics"`
}

// OperationType defines the type of operation
type OperationType string

const (
	OperationTypeScraping   OperationType = "scraping"
	OperationTypeProcessing OperationType = "processing"
	OperationTypeIndexing   OperationType = "indexing"
	OperationTypeLiquidity   OperationType = "liquidity"
)

// OperationStatus represents the status of an operation
type OperationStatus string

const (
	OperationStatusPending   OperationStatus = "pending"
	OperationStatusRunning   OperationStatus = "running"
	OperationStatusCompleted OperationStatus = "completed"
	OperationStatusFailed    OperationStatus = "failed"
	OperationStatusCancelled OperationStatus = "cancelled"
	OperationStatusPaused    OperationStatus = "paused"
	OperationStatusRetrying  OperationStatus = "retrying"
)

// OperationConfig represents operation configuration
type OperationConfig struct {
	Mode           string            `json:"mode" validate:"required,oneof=initial accumulative full"`
	StartDate      string            `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate        string            `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	MaxRetries     int               `json:"max_retries" validate:"min=0,max=10"`
	RetryDelay     int               `json:"retry_delay" validate:"min=1"` // seconds
	Timeout        int               `json:"timeout" validate:"min=60"`    // seconds
	Parallel       bool              `json:"parallel"`
	MaxWorkers     int               `json:"max_workers" validate:"min=1,max=100"`
	StepConfigs   map[string]interface{} `json:"step_configs,omitempty"`
	ErrorHandling  string            `json:"error_handling" validate:"omitempty,oneof=stop continue skip"`
	NotifyOnComplete bool            `json:"notify_on_complete"`
	NotifyOnError    bool            `json:"notify_on_error"`
	SaveIntermediateResults bool   `json:"save_intermediate_results"`
}

// Step represents an operation step
type Step struct {
	ID          string                 `json:"id" db:"id" validate:"required"`
	Name        string                 `json:"name" db:"name" validate:"required"`
	Type        StepType               `json:"type" db:"type" validate:"required"`
	Status      StepStatus            `json:"status" db:"status"`
	Order       int                    `json:"order" db:"order" validate:"min=0"`
	Config      map[string]interface{} `json:"config,omitempty" db:"config"`
	Dependencies []string              `json:"dependencies,omitempty" db:"dependencies"`
	StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	Duration    *time.Duration         `json:"duration,omitempty" db:"duration"`
	State       StepState             `json:"state,omitempty" db:"state"`
	Metrics     StepMetrics            `json:"metrics,omitempty" db:"metrics"`
	Error       string                 `json:"error,omitempty" db:"error"`
}

// StepType defines the type of operation step
type StepType string

const (
	StepTypeScraping     StepType = "scraping"
	StepTypeProcessing   StepType = "processing"
	StepTypeValidation   StepType = "validation"
	StepTypeTransform    StepType = "transform"
	StepTypeLiquidity     StepType = "liquidity"
	StepTypeExport       StepType = "export"
	StepTypeNotification StepType = "notification"
)

// StepStatus represents the status of a step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusRetrying  StepStatus = "retrying"
)

// StepState represents the internal state of a step
type StepState struct {
	Progress       float64                `json:"progress"` // 0-100
	CurrentItem    string                 `json:"current_item,omitempty"`
	ItemsProcessed int64                  `json:"items_processed"`
	ItemsTotal     int64                  `json:"items_total,omitempty"`
	LastError      string                 `json:"last_error,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	Checkpoints    map[string]interface{} `json:"checkpoints,omitempty"`
}

// StepExecutionResult represents the result of step execution
type StepExecutionResult struct {
	StepID       string                 `json:"step_id"`
	Success      bool                   `json:"success"`
	Data         interface{}            `json:"data,omitempty"`
	Error        error                  `json:"error,omitempty"`
	Duration     time.Duration          `json:"duration"`
	NextStepID   string                 `json:"next_step_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// OperationMetrics represents operation execution metrics
type OperationMetrics struct {
	TotalDuration    time.Duration          `json:"total_duration"`
	StepsCompleted   int                    `json:"steps_completed"`
	StepsFailed      int                    `json:"steps_failed"`
	StepsSkipped     int                    `json:"steps_skipped"`
	ItemsProcessed   int64                  `json:"items_processed"`
	BytesProcessed   int64                  `json:"bytes_processed"`
	ErrorRate        float64                `json:"error_rate"`
	AvgStepTime      time.Duration          `json:"avg_step_time"`
	ResourceUsage    ResourceMetrics        `json:"resource_usage"`
}

// StepMetrics represents step execution metrics
type StepMetrics struct {
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	Duration       time.Duration          `json:"duration"`
	ItemsProcessed int64                  `json:"items_processed"`
	ItemsFailed    int64                  `json:"items_failed"`
	BytesRead      int64                  `json:"bytes_read"`
	BytesWritten   int64                  `json:"bytes_written"`
	CPUUsage       float64                `json:"cpu_usage"`
	MemoryUsage    int64                  `json:"memory_usage"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// ResourceMetrics represents resource usage metrics
type ResourceMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsageBytes   int64   `json:"memory_usage_bytes"`
	DiskReadBytes      int64   `json:"disk_read_bytes"`
	DiskWriteBytes     int64   `json:"disk_write_bytes"`
	NetworkInBytes     int64   `json:"network_in_bytes"`
	NetworkOutBytes    int64   `json:"network_out_bytes"`
}

// RetryConfig represents retry configuration for steps
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts" validate:"min=1,max=10"`
	InitialDelay    time.Duration `json:"initial_delay" validate:"min=1s"`
	MaxDelay        time.Duration `json:"max_delay" validate:"min=1s"`
	BackoffFactor   float64       `json:"backoff_factor" validate:"min=1.0,max=10.0"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"`
}

// ProgressUpdate represents a progress update for an operation or step
type ProgressUpdate struct {
	OperationID    string                 `json:"operation_id"`
	StepID         string                 `json:"step_id,omitempty"`
	Progress       float64                `json:"progress"` // 0-100
	Message        string                 `json:"message,omitempty"`
	ItemsProcessed int64                  `json:"items_processed,omitempty"`
	ItemsTotal     int64                  `json:"items_total,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// OperationProgressUpdate represents a progress update for an operation or step
// This is the new terminology - operation/step are being renamed to Operation/Step
type OperationProgressUpdate struct {
	OperationID    string                 `json:"operation_id"`
	StepID         string                 `json:"step_id,omitempty"`
	Progress       float64                `json:"progress"` // 0-100
	Message        string                 `json:"message,omitempty"`
	ItemsProcessed int64                  `json:"items_processed,omitempty"`
	ItemsTotal     int64                  `json:"items_total,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// OperationRequest represents a request to start a operation
type OperationRequest struct {
	Type       OperationType          `json:"type" validate:"required"`
	Mode       string                 `json:"mode" validate:"required,oneof=initial accumulative full"`
	StartDate  string                 `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate    string                 `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

// OperationResponse represents an operation execution response
type OperationResponse struct {
	OperationID string         `json:"operation_id"`
	Status     OperationStatus `json:"status"`
	Message    string         `json:"message"`
	StartedAt  time.Time      `json:"started_at"`
	WebSocketURL string       `json:"websocket_url,omitempty"`
}

// Operation modes
const (
	OperationModeInitial     = "initial"
	OperationModeAccumulative = "accumulative"
	OperationModeFull        = "full"
)

// Context keys for operation execution
const (
	ContextKeyFromDate       = "from_date"
	ContextKeyToDate         = "to_date"
	ContextKeyMode           = "mode"
	ContextKeyDownloadDir    = "download_dir"
	ContextKeyReportDir      = "report_dir"
	ContextKeyFilesFound     = "files_found"
	ContextKeyFilesProcessed = "files_processed"
	ContextKeyScraperSuccess = "scraper_success"
	ContextKeyTraceID        = "trace_id"
	ContextKeyUserID         = "user_id"
)

// Step identifiers
const (
	StepIDScraping   = "scraping"
	StepIDProcessing = "processing"
	StepIDIndices    = "indices"
	StepIDLiquidity   = "liquidity"
)

// Step names
const (
	StepNameScraping   = "Data Scraping"
	StepNameProcessing = "Data Processing"
	StepNameIndices    = "Index Calculation"
	StepNameLiquidity   = "Liquidity Calculation"
)