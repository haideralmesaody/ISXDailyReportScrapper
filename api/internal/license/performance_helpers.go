package license

import (
	"context"
	"log/slog"
	"time"
)

// TrackOperation wraps an operation with performance tracking and logging.
func (m *Manager) TrackOperation(operation string, fn func() error) error {
	start := time.Now()

	m.logDebug(context.Background(), operation+"_start", "Operation initiated")

	err := fn()
	duration := time.Since(start)

	m.recordPerformanceMetric(operation, duration, err == nil)

	if err != nil {
		m.logError(context.Background(), operation+"_complete", "Operation failed",
			slog.Duration("duration", duration),
			slog.String("error", err.Error()),
		)
	} else {
		m.logInfo(context.Background(), operation+"_complete", "Operation completed successfully",
			slog.Duration("duration", duration),
		)
	}

	return err
}

// recordPerformanceMetric updates performance statistics.
func (m *Manager) recordPerformanceMetric(operation string, duration time.Duration, success bool) {
	if m.performanceTracker == nil {
		m.performanceTracker = NewPerformanceTracker()
	}
	m.performanceTracker.Record(operation, duration, success)
}

// GetPerformanceMetrics returns performance statistics.
func (m *Manager) GetPerformanceMetrics() map[string]*PerformanceMetrics {
	if m.performanceTracker == nil {
		return map[string]*PerformanceMetrics{}
	}
	return m.performanceTracker.Snapshot()
}

// GetSystemStats returns comprehensive system statistics.
func (m *Manager) GetSystemStats() map[string]interface{} {
	stats := map[string]interface{}{
		"performance": m.GetPerformanceMetrics(),
		"timestamp":   time.Now(),
		"version":     "enhanced-v2.0.0",
	}

	if m.cache != nil {
		stats["cache"] = m.cache.GetStats()
	}

	if m.security != nil {
		stats["security"] = m.security.GetStats()
	}

	return stats
}
