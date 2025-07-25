package websocket

// Message types
const (
	// Core message types
	TypeOutput           = "output"
	TypeDataUpdate       = "data_update"
	TypePipelineStatus   = "pipeline:status"      // Changed to frontend format
	TypePipelineProgress = "pipeline:progress"    // Changed to frontend format
	TypePipelineReset    = "pipeline:reset"       // Changed to frontend format
	TypePipelineComplete = "pipeline:complete"    // Changed to frontend format
	TypeProgress         = "progress"
	TypeStatus           = "status"
	TypeError            = "error"
	TypeConnection       = "connection"
	TypeLog              = "log"
	TypeRefresh          = "refresh" // Deprecated - use data_update instead
)

// Message levels
const (
	LevelInfo    = "info"
	LevelSuccess = "success"
	LevelWarning = "warning"
	LevelError   = "error"
)

// Pipeline stages
const (
	StageScraping   = "scraping"
	StageProcessing = "processing"
	StageIndices    = "indices"
	StageAnalysis   = "analysis"
)

// Pipeline status values
const (
	StatusInactive   = "inactive"
	StatusActive     = "active"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusError      = "error"
)

// Data update subtypes
const (
	SubtypeTickerSummary  = "ticker_summary"
	SubtypeCombinedData   = "combined_data"
	SubtypeDailyReport    = "daily_report"
	SubtypeIndexes        = "indexes"
	SubtypeTickerHistory  = "ticker_history"
	SubtypeAll            = "all"
)

// Data update actions
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
	ActionRefresh = "refresh"
)

// Error codes - Scraping errors (1xxx)
const (
	ErrScrapingTimeout     = "ERR_1001"
	ErrScrapingNoData      = "ERR_1002"
	ErrScrapingAuthFailed  = "ERR_1003"
	ErrScrapingInvalidDate = "ERR_1004"
)

// Error codes - Processing errors (2xxx)
const (
	ErrProcessingInvalidFile = "ERR_2001"
	ErrProcessingNoColumns   = "ERR_2002"
	ErrProcessingParseFailed = "ERR_2003"
)

// Error codes - System errors (9xxx)
const (
	ErrSystemOutOfMemory = "ERR_9001"
	ErrSystemDiskFull    = "ERR_9002"
	ErrSystemFileAccess  = "ERR_9003"
)

// ErrorRecoveryHints provides user-friendly recovery suggestions
var ErrorRecoveryHints = map[string]string{
	ErrScrapingTimeout:       "Check your internet connection and try again",
	ErrScrapingNoData:        "No data available for the specified date range",
	ErrScrapingAuthFailed:    "Website authentication may have changed, please report this issue",
	ErrScrapingInvalidDate:   "Please select a valid date range",
	ErrProcessingInvalidFile: "File may be corrupted, try re-downloading",
	ErrProcessingNoColumns:   "Excel file format may have changed, please report this issue",
	ErrProcessingParseFailed: "Unable to parse the Excel file, ensure it's a valid ISX report",
	ErrSystemOutOfMemory:     "System is low on memory, try closing other applications",
	ErrSystemDiskFull:        "Not enough disk space, please free up some space",
	ErrSystemFileAccess:      "Cannot access the file, check permissions",
}