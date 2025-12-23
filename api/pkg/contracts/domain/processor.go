package domain

import (
	"time"
)

// ProcessorConfig represents configuration for data processing operations
type ProcessorConfig struct {
	Type              string                 `json:"type" validate:"required,oneof=csv excel json xml pdf"`
	InputPath         string                 `json:"input_path" validate:"required"`
	OutputPath        string                 `json:"output_path" validate:"required"`
	Mode              ProcessingMode         `json:"mode" validate:"required"`
	BatchSize         int                    `json:"batch_size" validate:"min=1,max=10000"`
	MaxWorkers        int                    `json:"max_workers" validate:"min=1,max=100"`
	Timeout           time.Duration          `json:"timeout" validate:"min=1s"`
	RetryAttempts     int                    `json:"retry_attempts" validate:"min=0,max=5"`
	ValidationLevel   ValidationLevel        `json:"validation_level"`
	ErrorHandling     ErrorHandlingStrategy  `json:"error_handling"`
	Encoding          string                 `json:"encoding" validate:"omitempty,oneof=utf8 utf16 windows1256"`
	Delimiter         string                 `json:"delimiter,omitempty"`
	DateFormat        string                 `json:"date_format,omitempty"`
	NumberFormat      string                 `json:"number_format,omitempty"`
	SkipRows          int                    `json:"skip_rows" validate:"min=0"`
	MaxRows           int                    `json:"max_rows" validate:"min=0"`
	ColumnMappings    map[string]string      `json:"column_mappings,omitempty"`
	Transformations   []TransformationRule   `json:"transformations,omitempty"`
	Filters           []FilterRule           `json:"filters,omitempty"`
	OutputOptions     OutputOptions          `json:"output_options"`
	CustomParameters  map[string]interface{} `json:"custom_parameters,omitempty"`
}

// ProcessingMode defines how data is processed
type ProcessingMode string

const (
	ProcessingModeFull        ProcessingMode = "full"        // Process entire file
	ProcessingModeIncremental ProcessingMode = "incremental" // Process only new/changed data
	ProcessingModeDelta       ProcessingMode = "delta"       // Process differences
	ProcessingModeStreaming   ProcessingMode = "streaming"   // Stream processing
)

// ValidationLevel defines the level of data validation
type ValidationLevel string

const (
	ValidationLevelNone   ValidationLevel = "none"   // No validation
	ValidationLevelBasic  ValidationLevel = "basic"  // Basic type checking
	ValidationLevelStrict ValidationLevel = "strict" // Full validation with business rules
)

// ErrorHandlingStrategy defines how errors are handled during processing
type ErrorHandlingStrategy string

const (
	ErrorHandlingSkip     ErrorHandlingStrategy = "skip"     // Skip error records
	ErrorHandlingLog      ErrorHandlingStrategy = "log"      // Log errors and continue
	ErrorHandlingFail     ErrorHandlingStrategy = "fail"     // Fail on first error
	ErrorHandlingIsolate  ErrorHandlingStrategy = "isolate"  // Move error records to separate output
)

// TransformationRule represents a data transformation rule
type TransformationRule struct {
	Name        string                 `json:"name" validate:"required"`
	Type        TransformationType     `json:"type" validate:"required"`
	Field       string                 `json:"field" validate:"required"`
	TargetField string                 `json:"target_field,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Condition   string                 `json:"condition,omitempty"`
}

// TransformationType defines types of transformations
type TransformationType string

const (
	TransformTypeMap         TransformationType = "map"          // Map values
	TransformTypeCalculate   TransformationType = "calculate"    // Calculate new values
	TransformTypeAggregate   TransformationType = "aggregate"    // Aggregate values
	TransformTypeNormalize   TransformationType = "normalize"    // Normalize data
	TransformTypeSplit       TransformationType = "split"        // Split fields
	TransformTypeMerge       TransformationType = "merge"        // Merge fields
	TransformTypeFormat      TransformationType = "format"       // Format values
	TransformTypeValidate    TransformationType = "validate"     // Validate and clean
)

// FilterRule represents a data filtering rule
type FilterRule struct {
	Name       string         `json:"name" validate:"required"`
	Type       FilterType     `json:"type" validate:"required"`
	Field      string         `json:"field" validate:"required"`
	Operator   FilterOperator `json:"operator" validate:"required"`
	Value      interface{}    `json:"value"`
	Values     []interface{}  `json:"values,omitempty"`
	CaseSensitive bool        `json:"case_sensitive"`
}

// FilterType defines types of filters
type FilterType string

const (
	FilterTypeInclude FilterType = "include" // Include matching records
	FilterTypeExclude FilterType = "exclude" // Exclude matching records
)

// FilterOperator defines filter comparison operators
type FilterOperator string

const (
	FilterOpEquals         FilterOperator = "equals"
	FilterOpNotEquals      FilterOperator = "not_equals"
	FilterOpGreaterThan    FilterOperator = "greater_than"
	FilterOpLessThan       FilterOperator = "less_than"
	FilterOpGreaterOrEqual FilterOperator = "greater_or_equal"
	FilterOpLessOrEqual    FilterOperator = "less_or_equal"
	FilterOpContains       FilterOperator = "contains"
	FilterOpNotContains    FilterOperator = "not_contains"
	FilterOpStartsWith     FilterOperator = "starts_with"
	FilterOpEndsWith       FilterOperator = "ends_with"
	FilterOpIn             FilterOperator = "in"
	FilterOpNotIn          FilterOperator = "not_in"
	FilterOpBetween        FilterOperator = "between"
	FilterOpRegex          FilterOperator = "regex"
)

// OutputOptions defines options for output generation
type OutputOptions struct {
	Format           string            `json:"format" validate:"required,oneof=csv excel json xml pdf html"`
	Compression      CompressionType   `json:"compression"`
	Encryption       EncryptionConfig  `json:"encryption,omitempty"`
	IncludeHeaders   bool              `json:"include_headers"`
	IncludeMetadata  bool              `json:"include_metadata"`
	IncludeTimestamp bool              `json:"include_timestamp"`
	FileNaming       FileNamingPattern `json:"file_naming"`
	SplitSize        int64             `json:"split_size"` // Split files larger than this size
	SplitRows        int               `json:"split_rows"` // Split files with more rows than this
}

// CompressionType defines compression options
type CompressionType string

const (
	CompressionNone CompressionType = "none"
	CompressionGzip CompressionType = "gzip"
	CompressionZip  CompressionType = "zip"
	CompressionBz2  CompressionType = "bz2"
)

// EncryptionConfig defines encryption configuration
type EncryptionConfig struct {
	Algorithm   string `json:"algorithm" validate:"required,oneof=aes256 rsa"`
	KeyPath     string `json:"key_path,omitempty"`
	KeyData     string `json:"key_data,omitempty"`
	Passphrase  string `json:"passphrase,omitempty"`
}

// FileNamingPattern defines how output files are named
type FileNamingPattern struct {
	Template    string `json:"template" validate:"required"` // e.g., "{date}_{type}_{sequence}"
	DateFormat  string `json:"date_format"`
	SequenceFormat string `json:"sequence_format"`
	CaseSensitive bool `json:"case_sensitive"`
}

// ProcessingResult represents the result of a processing operation
type ProcessingResult struct {
	ID              string                 `json:"id"`
	ProcessorType   string                 `json:"processor_type"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Status          ProcessingStatus       `json:"status"`
	InputFile       string                 `json:"input_file"`
	OutputFiles     []string               `json:"output_files"`
	RecordsRead     int64                  `json:"records_read"`
	RecordsWritten  int64                  `json:"records_written"`
	RecordsSkipped  int64                  `json:"records_skipped"`
	RecordsErrored  int64                  `json:"records_errored"`
	BytesRead       int64                  `json:"bytes_read"`
	BytesWritten    int64                  `json:"bytes_written"`
	Errors          []ProcessingError      `json:"errors,omitempty"`
	Warnings        []ProcessingWarning    `json:"warnings,omitempty"`
	Statistics      map[string]interface{} `json:"statistics,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessingStatus represents the status of a processing operation
type ProcessingStatus string

const (
	ProcessingStatusPending   ProcessingStatus = "pending"
	ProcessingStatusRunning   ProcessingStatus = "running"
	ProcessingStatusCompleted ProcessingStatus = "completed"
	ProcessingStatusFailed    ProcessingStatus = "failed"
	ProcessingStatusCancelled ProcessingStatus = "cancelled"
	ProcessingStatusPartial   ProcessingStatus = "partial" // Completed with errors
)

// ProcessingError represents an error during processing
type ProcessingError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	RecordIndex int64                  `json:"record_index,omitempty"`
	RecordData  map[string]interface{} `json:"record_data,omitempty"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    string                 `json:"severity"` // critical, error, warning
}

// ProcessingWarning represents a warning during processing
type ProcessingWarning struct {
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	RecordIndex int64     `json:"record_index,omitempty"`
	Field       string    `json:"field,omitempty"`
	Suggestion  string    `json:"suggestion,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// DataQualityReport represents a data quality analysis report
type DataQualityReport struct {
	ID              string                   `json:"id"`
	FileName        string                   `json:"file_name"`
	AnalyzedAt      time.Time                `json:"analyzed_at"`
	TotalRecords    int64                    `json:"total_records"`
	ValidRecords    int64                    `json:"valid_records"`
	InvalidRecords  int64                    `json:"invalid_records"`
	QualityScore    float64                  `json:"quality_score"` // 0-100
	FieldAnalysis   []FieldQualityAnalysis   `json:"field_analysis"`
	DataIssues      []DataIssue              `json:"data_issues"`
	Recommendations []string                 `json:"recommendations"`
}

// FieldQualityAnalysis represents quality analysis for a single field
type FieldQualityAnalysis struct {
	FieldName        string                 `json:"field_name"`
	DataType         string                 `json:"data_type"`
	Completeness     float64                `json:"completeness"` // Percentage of non-null values
	Uniqueness       float64                `json:"uniqueness"`   // Percentage of unique values
	Validity         float64                `json:"validity"`     // Percentage of valid values
	NullCount        int64                  `json:"null_count"`
	UniqueCount      int64                  `json:"unique_count"`
	MinValue         interface{}            `json:"min_value,omitempty"`
	MaxValue         interface{}            `json:"max_value,omitempty"`
	AverageLength    float64                `json:"average_length,omitempty"`
	Patterns         []string               `json:"patterns,omitempty"`
	FrequentValues   map[string]int         `json:"frequent_values,omitempty"`
}

// DataIssue represents a data quality issue
type DataIssue struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Field       string   `json:"field,omitempty"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
	Impact      string   `json:"impact"`
	Resolution  string   `json:"resolution"`
}