package operations

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMetric tracks historical performance data
type PerformanceMetric struct {
	Duration      time.Duration `json:"duration"`
	FileCount     int           `json:"file_count"`
	Timestamp     time.Time     `json:"timestamp"`
	OperationType string        `json:"operation_type"`
	Success       bool          `json:"success"`
}

// AdaptiveWatchdog provides intelligent timeout management based on performance history
type AdaptiveWatchdog struct {
	baseTimeout        time.Duration
	perFileTimeout     time.Duration
	multiplier         float64
	lastActivity       time.Time
	learningMode       bool
	performanceHistory []PerformanceMetric
	maxHistorySize     int
	fileCount          int
	operationType      string
	triggered          atomic.Bool
	cancelFunc         context.CancelFunc
	logger             *slog.Logger
	mutex              sync.RWMutex
}

// NewAdaptiveWatchdog creates a new adaptive watchdog with 10s per file as requested
func NewAdaptiveWatchdog(fileCount int, operationType string, logger *slog.Logger) *AdaptiveWatchdog {
	return &AdaptiveWatchdog{
		baseTimeout:        2 * time.Minute,  // Reduced base timeout
		perFileTimeout:     10 * time.Second, // As requested by user
		multiplier:         1.3,              // 30% buffer for safety
		learningMode:       true,
		performanceHistory: make([]PerformanceMetric, 0),
		maxHistorySize:     50, // Keep last 50 operations
		fileCount:          fileCount,
		operationType:      operationType,
		logger:             logger,
		lastActivity:       time.Now(),
	}
}

// GetTimeout calculates the dynamic timeout based on performance history and file count
func (w *AdaptiveWatchdog) GetTimeout() time.Duration {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// If we have enough historical data, use it
	if w.learningMode && len(w.performanceHistory) >= 3 {
		avgDuration := w.calculateAverageDuration()
		calculatedTimeout := time.Duration(float64(avgDuration) * w.multiplier)

		if w.logger != nil {
			w.logger.Debug("Using calculated timeout based on performance history",
				slog.Duration("calculated_timeout", calculatedTimeout),
				slog.Duration("average_duration", avgDuration),
				slog.Float64("multiplier", w.multiplier),
				slog.Int("historical_operations", len(w.performanceHistory)))
		}

		return calculatedTimeout
	}

	// Fall back to base + per-file calculation
	baseTimeout := w.baseTimeout + time.Duration(w.fileCount)*w.perFileTimeout

	// Add safety margin for network conditions
	safetyMargin := 2 * time.Minute
	finalTimeout := baseTimeout + safetyMargin

	// Cap at reasonable maximum (30 minutes for large operations)
	maxTimeout := 30 * time.Minute
	if finalTimeout > maxTimeout {
		finalTimeout = maxTimeout
	}

	if w.logger != nil {
		w.logger.Info("Using dynamic timeout (10s per file)",
			slog.Int("file_count", w.fileCount),
			slog.Duration("base_timeout", w.baseTimeout),
			slog.Duration("per_file_timeout", w.perFileTimeout),
			slog.Duration("safety_margin", safetyMargin),
			slog.Duration("calculated_timeout", finalTimeout),
			slog.Duration("max_timeout", maxTimeout),
			slog.Bool("learning_mode", w.learningMode && len(w.performanceHistory) >= 3))
	}

	return finalTimeout
}

// RecordPerformance records a performance metric for future learning
func (w *AdaptiveWatchdog) RecordPerformance(duration time.Duration, success bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	metric := PerformanceMetric{
		Duration:      duration,
		FileCount:     w.fileCount,
		Timestamp:     time.Now(),
		OperationType: w.operationType,
		Success:       success,
	}

	// Add to history
	w.performanceHistory = append(w.performanceHistory, metric)

	// Trim history if too large
	if len(w.performanceHistory) > w.maxHistorySize {
		w.performanceHistory = w.performanceHistory[1:]
	}

	if w.logger != nil {
		w.logger.Info("Recorded performance metric",
			slog.Duration("duration", duration),
			slog.Int("file_count", w.fileCount),
			slog.Bool("success", success),
			slog.String("operation_type", w.operationType),
			slog.Int("history_size", len(w.performanceHistory)))
	}
}

// UpdateActivity updates the last activity timestamp
func (w *AdaptiveWatchdog) UpdateActivity() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.lastActivity = time.Now()
}

// GetLastActivity returns the last activity timestamp
func (w *AdaptiveWatchdog) GetLastActivity() time.Time {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.lastActivity
}

// StartMonitoring starts the watchdog monitoring in a separate goroutine
func (w *AdaptiveWatchdog) StartMonitoring(ctx context.Context, cancelFunc context.CancelFunc) {
	w.cancelFunc = cancelFunc

	timeout := w.GetTimeout()

	if w.logger != nil {
		w.logger.Info("Starting adaptive watchdog monitoring",
			slog.Duration("timeout", timeout),
			slog.Int("file_count", w.fileCount),
			slog.String("operation_type", w.operationType))
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if w.logger != nil {
					w.logger.Debug("Adaptive watchdog stopped", "reason", "context_done")
				}
				return
			case <-ticker.C:
				if w.isTimeoutExceeded() {
					if w.triggered.CompareAndSwap(false, true) {
						if w.logger != nil {
							w.logger.Warn("Adaptive watchdog timeout triggered",
								slog.Duration("timeout", timeout),
								slog.Time("last_activity", w.GetLastActivity()),
								slog.Duration("inactivity_duration", time.Since(w.GetLastActivity())))
						}

						// Cancel the operation
						if w.cancelFunc != nil {
							w.cancelFunc()
						}
					}
					return
				}
			}
		}
	}()
}

// isTimeoutExceeded checks if the timeout has been exceeded
func (w *AdaptiveWatchdog) isTimeoutExceeded() bool {
	lastActivity := w.GetLastActivity()
	if lastActivity.IsZero() {
		return false
	}

	timeout := w.GetTimeout()
	return time.Since(lastActivity) > timeout
}

// Triggered returns whether the watchdog has been triggered
func (w *AdaptiveWatchdog) Triggered() bool {
	return w.triggered.Load()
}

// calculateAverageDuration calculates the average duration from similar historical operations
func (w *AdaptiveWatchdog) calculateAverageDuration() time.Duration {
	if len(w.performanceHistory) == 0 {
		return 0
	}

	// Filter for similar operations (same type and similar file count)
	var similarOps []PerformanceMetric
	fileCountVariance := w.fileCount / 4 // Allow 25% variance

	for _, metric := range w.performanceHistory {
		if metric.OperationType == w.operationType && metric.Success {
			countDiff := metric.FileCount - w.fileCount
			if countDiff < 0 {
				countDiff = -countDiff
			}

			if countDiff <= fileCountVariance {
				similarOps = append(similarOps, metric)
			}
		}
	}

	if len(similarOps) == 0 {
		// Fall back to all successful operations of same type
		for _, metric := range w.performanceHistory {
			if metric.OperationType == w.operationType && metric.Success {
				similarOps = append(similarOps, metric)
			}
		}
	}

	if len(similarOps) == 0 {
		return 0
	}

	// Calculate weighted average (more recent operations have higher weight)
	var totalWeightedDuration time.Duration
	var totalWeight int

	for i, metric := range similarOps {
		// Newer operations get higher weight
		weight := (i + 1)
		totalWeightedDuration += metric.Duration * time.Duration(weight)
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	avgDuration := totalWeightedDuration / time.Duration(totalWeight)

	if w.logger != nil {
		w.logger.Debug("Calculated average duration from similar operations",
			slog.Duration("average_duration", avgDuration),
			slog.Int("similar_operations", len(similarOps)),
			slog.Int("file_count", w.fileCount))
	}

	return avgDuration
}

// GetPerformanceHistory returns the current performance history for debugging
func (w *AdaptiveWatchdog) GetPerformanceHistory() []PerformanceMetric {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// Return a copy to prevent external modification
	history := make([]PerformanceMetric, len(w.performanceHistory))
	copy(history, w.performanceHistory)

	return history
}

// SetLearningMode enables or disables learning mode
func (w *AdaptiveWatchdog) SetLearningMode(enabled bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.learningMode = enabled

	if w.logger != nil {
		w.logger.Info("Adaptive watchdog learning mode changed",
			slog.Bool("enabled", enabled),
			slog.Int("history_size", len(w.performanceHistory)))
	}
}

// StopMonitoring stops the adaptive watchdog monitoring
func (w *AdaptiveWatchdog) StopMonitoring() {
	if w.cancelFunc != nil {
		w.cancelFunc()
	}

	if w.logger != nil {
		w.logger.Debug("Adaptive watchdog monitoring stopped")
	}
}
