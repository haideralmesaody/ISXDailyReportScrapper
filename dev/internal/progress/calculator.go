package progress

import (
	"encoding/json"
	"fmt"
	"time"
)

// Calculator manages progress tracking and ETA estimation
type Calculator struct {
	StartTime    time.Time
	TotalItems   int
	ProcessedItems int
	Stage        string
	itemTimes    []time.Duration
}

// NewCalculator creates a new progress calculator
func NewCalculator(stage string, totalItems int) *Calculator {
	return &Calculator{
		StartTime:    time.Now(),
		TotalItems:   totalItems,
		ProcessedItems: 0,
		Stage:        stage,
		itemTimes:    make([]time.Duration, 0),
	}
}

// Update records progress for an item
func (c *Calculator) Update(itemsProcessed int) {
	if itemsProcessed > c.ProcessedItems {
		elapsed := time.Since(c.StartTime)
		itemTime := elapsed / time.Duration(itemsProcessed)
		c.itemTimes = append(c.itemTimes, itemTime)
		c.ProcessedItems = itemsProcessed
	}
}

// GetProgress returns current progress percentage
func (c *Calculator) GetProgress() float64 {
	if c.TotalItems == 0 {
		return 0
	}
	return float64(c.ProcessedItems) / float64(c.TotalItems) * 100
}

// GetETA estimates time remaining
func (c *Calculator) GetETA() string {
	if c.ProcessedItems == 0 || c.ProcessedItems >= c.TotalItems {
		return ""
	}

	// Calculate average time per item
	elapsed := time.Since(c.StartTime)
	avgTimePerItem := elapsed / time.Duration(c.ProcessedItems)
	
	// Estimate remaining time
	remainingItems := c.TotalItems - c.ProcessedItems
	estimatedRemaining := avgTimePerItem * time.Duration(remainingItems)
	
	// Format ETA
	if estimatedRemaining < time.Minute {
		return fmt.Sprintf("%d seconds remaining", int(estimatedRemaining.Seconds()))
	} else if estimatedRemaining < time.Hour {
		minutes := int(estimatedRemaining.Minutes())
		return fmt.Sprintf("%d minutes remaining", minutes)
	} else {
		hours := int(estimatedRemaining.Hours())
		minutes := int(estimatedRemaining.Minutes()) % 60
		return fmt.Sprintf("%dh %dm remaining", hours, minutes)
	}
}

// ProgressMessage represents a WebSocket progress message
type ProgressMessage struct {
	Stage      string                 `json:"stage"`
	Current    int                    `json:"current"`
	Total      int                    `json:"total"`
	Percentage float64                `json:"percentage"`
	Message    string                 `json:"message"`
	ETA        string                 `json:"eta,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// GetProgressMessage returns a formatted progress message
func (c *Calculator) GetProgressMessage(message string, details map[string]interface{}) *ProgressMessage {
	return &ProgressMessage{
		Stage:      c.Stage,
		Current:    c.ProcessedItems,
		Total:      c.TotalItems,
		Percentage: c.GetProgress(),
		Message:    message,
		ETA:        c.GetETA(),
		Details:    details,
	}
}

// ToJSON converts progress message to JSON string with proper WebSocket format
func (c *Calculator) ToJSON(message string, details map[string]interface{}) (string, error) {
	msg := c.GetProgressMessage(message, details)
	
	// Wrap in WebSocket message format
	wsMsg := map[string]interface{}{
		"type": "progress",
		"data": msg,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	data, err := json.Marshal(wsMsg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// StatusMessage represents a WebSocket status message
type StatusMessage struct {
	Stage   string `json:"stage"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CreateStatusMessage creates a status message
func CreateStatusMessage(stage, status, message string) (*StatusMessage, error) {
	return &StatusMessage{
		Stage:   stage,
		Status:  status,
		Message: message,
	}, nil
}

// StatusToJSON converts status message to JSON string with proper WebSocket format
func StatusToJSON(stage, status, message string) (string, error) {
	msg := &StatusMessage{
		Stage:   stage,
		Status:  status,
		Message: message,
	}
	
	// Wrap in WebSocket message format
	wsMsg := map[string]interface{}{
		"type": "pipeline_status",
		"data": msg,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	data, err := json.Marshal(wsMsg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ErrorMessage represents a WebSocket error message
type ErrorMessage struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Details     string `json:"details,omitempty"`
	Stage       string `json:"stage,omitempty"`
	Recoverable bool   `json:"recoverable"`
	Hint        string `json:"hint,omitempty"`
}

// ErrorToJSON converts error message to JSON string with proper WebSocket format
func ErrorToJSON(code, message, details, stage string, recoverable bool, hint string) (string, error) {
	msg := &ErrorMessage{
		Code:        code,
		Message:     message,
		Details:     details,
		Stage:       stage,
		Recoverable: recoverable,
		Hint:        hint,
	}
	
	// Wrap in WebSocket message format
	wsMsg := map[string]interface{}{
		"type": "error",
		"data": msg,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	data, err := json.Marshal(wsMsg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}