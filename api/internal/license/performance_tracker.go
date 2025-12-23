package license

import (
	"sync"
	"time"
)

// PerformanceTracker aggregates per-operation metrics with concurrency safety.
type PerformanceTracker struct {
	mu      sync.RWMutex
	metrics map[string]*PerformanceMetrics
}

// NewPerformanceTracker creates a ready-to-use tracker.
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics: make(map[string]*PerformanceMetrics),
	}
}

// Record updates the statistics for an operation.
func (pt *PerformanceTracker) Record(operation string, duration time.Duration, success bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.metrics == nil {
		pt.metrics = make(map[string]*PerformanceMetrics)
	}

	metric, exists := pt.metrics[operation]
	if !exists {
		metric = &PerformanceMetrics{
			MinTime: duration,
			MaxTime: duration,
		}
		pt.metrics[operation] = metric
	}

	metric.Count++
	metric.TotalTime += duration
	if metric.Count > 0 {
		metric.AverageTime = time.Duration(int64(metric.TotalTime) / metric.Count)
	}
	metric.LastUpdated = time.Now()

	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}
	if duration < metric.MinTime {
		metric.MinTime = duration
	}

	if success {
		metric.SuccessCount++
	} else {
		metric.ErrorCount++
	}
}

// Snapshot returns a copy of the current performance metrics map.
func (pt *PerformanceTracker) Snapshot() map[string]*PerformanceMetrics {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if len(pt.metrics) == 0 {
		return map[string]*PerformanceMetrics{}
	}

	result := make(map[string]*PerformanceMetrics, len(pt.metrics))
	for k, v := range pt.metrics {
		copy := *v
		result[k] = &copy
	}
	return result
}
