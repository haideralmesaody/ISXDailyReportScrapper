package domain

import (
	"time"
)

// Report represents a generated report
type Report struct {
	ID           string                 `json:"id" db:"id" validate:"required,uuid"`
	Type         ReportType             `json:"type" db:"type" validate:"required"`
	Title        string                 `json:"title" db:"title" validate:"required,min=3,max=200"`
	Description  string                 `json:"description,omitempty" db:"description"`
	Status       ReportStatus           `json:"status" db:"status"`
	Format       ReportFormat           `json:"format" db:"format" validate:"required"`
	FilePath     string                 `json:"file_path,omitempty" db:"file_path"`
	FileSize     int64                  `json:"file_size,omitempty" db:"file_size"`
	GeneratedAt  time.Time              `json:"generated_at" db:"generated_at"`
	GeneratedBy  string                 `json:"generated_by" db:"generated_by"`
	DateFrom     time.Time              `json:"date_from" db:"date_from"`
	DateTo       time.Time              `json:"date_to" db:"date_to"`
	Parameters   map[string]interface{} `json:"parameters,omitempty" db:"parameters"`
	Metadata     ReportMetadata         `json:"metadata" db:"metadata"`
	ExpiresAt    time.Time              `json:"expires_at,omitempty" db:"expires_at"`
	DownloadURL  string                 `json:"download_url,omitempty" db:"download_url"`
	DownloadCount int                   `json:"download_count" db:"download_count"`
	Tags         []string               `json:"tags,omitempty" db:"tags"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportType defines the type of report
type ReportType string

const (
	ReportTypeDaily     ReportType = "daily"
	ReportTypeWeekly    ReportType = "weekly"
	ReportTypeMonthly   ReportType = "monthly"
	ReportTypeSummary   ReportType = "summary"
	ReportTypeAnalysis  ReportType = "analysis"
	ReportTypeCustom    ReportType = "custom"
	ReportTypeCompliance ReportType = "compliance"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusProcessing ReportStatus = "processing"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusExpired    ReportStatus = "expired"
)

// ReportFormat defines the format of a report
type ReportFormat string

const (
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatExcel ReportFormat = "excel"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatHTML  ReportFormat = "html"
	ReportFormatXML   ReportFormat = "xml"
)

// ReportMetadata contains metadata about a report
type ReportMetadata struct {
	RecordCount      int64                  `json:"record_count"`
	PageCount        int                    `json:"page_count,omitempty"`
	ProcessingTime   time.Duration          `json:"processing_time"`
	DataSources      []string               `json:"data_sources"`
	IncludedSections []string               `json:"included_sections"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	Version          string                 `json:"version"`
	Checksum         string                 `json:"checksum,omitempty"`
}

// ReportTemplate represents a report template
type ReportTemplate struct {
	ID          string                 `json:"id" db:"id" validate:"required,uuid"`
	Name        string                 `json:"name" db:"name" validate:"required,min=3,max=100"`
	Type        ReportType             `json:"type" db:"type" validate:"required"`
	Description string                 `json:"description,omitempty" db:"description"`
	Template    string                 `json:"template" db:"template" validate:"required"`
	Sections    []ReportSection        `json:"sections" db:"sections"`
	Parameters  []ReportParameter      `json:"parameters" db:"parameters"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Active      bool                   `json:"active" db:"active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportSection represents a section in a report
type ReportSection struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name" validate:"required"`
	Title       string                 `json:"title" validate:"required"`
	Order       int                    `json:"order" validate:"min=0"`
	Type        string                 `json:"type" validate:"required,oneof=text table chart summary"`
	DataSource  string                 `json:"data_source,omitempty"`
	Query       string                 `json:"query,omitempty"`
	Template    string                 `json:"template,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Conditional bool                   `json:"conditional"`
	Condition   string                 `json:"condition,omitempty"`
}

// ReportParameter represents a parameter for report generation
type ReportParameter struct {
	Name         string      `json:"name" validate:"required"`
	Type         string      `json:"type" validate:"required,oneof=string number date boolean select"`
	Label        string      `json:"label" validate:"required"`
	Description  string      `json:"description,omitempty"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Options      []string    `json:"options,omitempty"` // For select type
	Validation   string      `json:"validation,omitempty"`
}

// ReportSchedule represents a scheduled report
type ReportSchedule struct {
	ID           string                 `json:"id" db:"id" validate:"required,uuid"`
	TemplateID   string                 `json:"template_id" db:"template_id" validate:"required,uuid"`
	Name         string                 `json:"name" db:"name" validate:"required"`
	Schedule     string                 `json:"schedule" db:"schedule" validate:"required"` // Cron expression
	Parameters   map[string]interface{} `json:"parameters,omitempty" db:"parameters"`
	Recipients   []string               `json:"recipients" db:"recipients" validate:"min=1"`
	Format       ReportFormat           `json:"format" db:"format" validate:"required"`
	Active       bool                   `json:"active" db:"active"`
	LastRun      *time.Time             `json:"last_run,omitempty" db:"last_run"`
	NextRun      *time.Time             `json:"next_run,omitempty" db:"next_run"`
	CreatedBy    string                 `json:"created_by" db:"created_by"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ReportFilter represents filters for report queries
type ReportFilter struct {
	Types        []ReportType           `json:"types,omitempty"`
	Statuses     []ReportStatus         `json:"statuses,omitempty"`
	DateFrom     *time.Time             `json:"date_from,omitempty"`
	DateTo       *time.Time             `json:"date_to,omitempty"`
	GeneratedBy  string                 `json:"generated_by,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	SearchTerm   string                 `json:"search_term,omitempty"`
}