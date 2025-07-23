package common

import (
	"fmt"
	"sync"
	"time"
)

// ProgressTracker tracks progress for long-running operations
type ProgressTracker struct {
	Stage     string
	Total     int
	Current   int
	StartTime time.Time
	Message   string
	mu        sync.Mutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(stage string, total int) *ProgressTracker {
	return &ProgressTracker{
		Stage:     stage,
		Total:     total,
		Current:   0,
		StartTime: time.Now(),
	}
}

// Update updates the current progress
func (p *ProgressTracker) Update(current int, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.Current = current
	p.Message = message
}

// Increment increments the current progress by 1
func (p *ProgressTracker) Increment(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.Current++
	p.Message = message
}

// GetProgress returns the current progress state
func (p *ProgressTracker) GetProgress() (current, total int, percentage float64, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	percentage = 0
	if p.Total > 0 {
		percentage = float64(p.Current) / float64(p.Total) * 100
	}
	
	return p.Current, p.Total, percentage, p.Message
}

// GetETA calculates the estimated time remaining
func (p *ProgressTracker) GetETA() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.Current == 0 || p.Total == 0 {
		return "calculating..."
	}
	
	elapsed := time.Since(p.StartTime)
	rate := float64(p.Current) / elapsed.Seconds()
	
	if rate == 0 {
		return "calculating..."
	}
	
	remaining := float64(p.Total-p.Current) / rate
	
	if remaining < 60 {
		return fmt.Sprintf("%.0f seconds", remaining)
	} else if remaining < 3600 {
		return fmt.Sprintf("%.1f minutes", remaining/60)
	} else {
		return fmt.Sprintf("%.1f hours", remaining/3600)
	}
}

// IsComplete returns true if the operation is complete
func (p *ProgressTracker) IsComplete() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	return p.Current >= p.Total
}

// GetElapsedTime returns the elapsed time since start
func (p *ProgressTracker) GetElapsedTime() time.Duration {
	return time.Since(p.StartTime)
}

// GetElapsedTimeString returns a formatted elapsed time string
func (p *ProgressTracker) GetElapsedTimeString() string {
	elapsed := p.GetElapsedTime()
	
	if elapsed < time.Minute {
		return fmt.Sprintf("%.0f seconds", elapsed.Seconds())
	} else if elapsed < time.Hour {
		return fmt.Sprintf("%.1f minutes", elapsed.Minutes())
	} else {
		return fmt.Sprintf("%.1f hours", elapsed.Hours())
	}
}