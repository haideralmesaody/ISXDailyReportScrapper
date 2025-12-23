package operations

import "time"

// FileProcessingStatus represents the status of an individual file being processed.
type FileProcessingStatus struct {
	FileName     string     `json:"file_name,omitempty"`
	Status       string     `json:"status"`   // "pending", "processing", "completed", "failed"
	Progress     float64    `json:"progress"` // 0-100 for this file
	SizeMB       float64    `json:"size_mb,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	ProcessingMs int64      `json:"processing_ms,omitempty"`
}
