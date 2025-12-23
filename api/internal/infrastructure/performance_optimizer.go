package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

// PerformanceOptimizer provides performance monitoring and optimization recommendations
type PerformanceOptimizer struct {
	logger           *slog.Logger
	systemCollector  *SystemMetricsCollector
	observations     []PerformanceObservation
	mu              sync.RWMutex
	alertThresholds PerformanceThresholds
}

// PerformanceThresholds defines performance alerting thresholds
type PerformanceThresholds struct {
	MemoryUsageMB        int64   // MB
	GoroutineCount       int64   // Number of goroutines
	GCPauseMS           int64   // Milliseconds
	CPUUsagePercent     float64 // Percentage
	ResponseTimeMS      int64   // Milliseconds
	ErrorRatePercent    float64 // Percentage
}

// PerformanceObservation represents a performance observation
type PerformanceObservation struct {
	Timestamp   time.Time
	Type        string
	Severity    string
	Message     string
	Metrics     map[string]interface{}
	Suggestion  string
}

// PerformanceReport contains performance analysis results
type PerformanceReport struct {
	Timestamp       time.Time                  `json:"timestamp"`
	OverallScore    int                       `json:"overall_score"`
	SystemHealth    SystemHealthStatus        `json:"system_health"`
	Bottlenecks     []PerformanceBottleneck   `json:"bottlenecks"`
	Recommendations []PerformanceRecommendation `json:"recommendations"`
	Trends          PerformanceTrends         `json:"trends"`
	Alerts          []PerformanceAlert        `json:"alerts"`
}

// SystemHealthStatus represents the overall system health
type SystemHealthStatus struct {
	Status        string  `json:"status"`
	Score         int     `json:"score"`
	MemoryUsage   int64   `json:"memory_usage_mb"`
	CPUUsage      float64 `json:"cpu_usage_percent"`
	GoroutineCount int64  `json:"goroutine_count"`
	Uptime        float64 `json:"uptime_seconds"`
}

// PerformanceBottleneck represents a performance bottleneck
type PerformanceBottleneck struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// PerformanceRecommendation represents a performance optimization recommendation
type PerformanceRecommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
	Action      string `json:"action"`
}

// PerformanceTrends represents performance trends over time
type PerformanceTrends struct {
	MemoryTrend     string `json:"memory_trend"`
	CPUTrend        string `json:"cpu_trend"`
	ResponseTrend   string `json:"response_trend"`
	ErrorTrend      string `json:"error_trend"`
	ThroughputTrend string `json:"throughput_trend"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Threshold   string    `json:"threshold"`
	CurrentValue string   `json:"current_value"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer(logger *slog.Logger, systemCollector *SystemMetricsCollector) *PerformanceOptimizer {
	return &PerformanceOptimizer{
		logger:          logger.With(slog.String("component", "performance_optimizer")),
		systemCollector: systemCollector,
		observations:    make([]PerformanceObservation, 0),
		alertThresholds: DefaultPerformanceThresholds(),
	}
}

// DefaultPerformanceThresholds returns default performance thresholds
func DefaultPerformanceThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MemoryUsageMB:       1000,  // 1GB
		GoroutineCount:      1000,  // 1000 goroutines
		GCPauseMS:          100,   // 100ms
		CPUUsagePercent:    85.0,  // 85%
		ResponseTimeMS:     2000,  // 2 seconds
		ErrorRatePercent:   5.0,   // 5%
	}
}

// AnalyzePerformance performs comprehensive performance analysis
func (po *PerformanceOptimizer) AnalyzePerformance(ctx context.Context) *PerformanceReport {
	po.logger.InfoContext(ctx, "Starting performance analysis")

	stats := po.systemCollector.GetCurrentStats(ctx)
	
	report := &PerformanceReport{
		Timestamp:    time.Now(),
		SystemHealth: po.analyzeSystemHealth(stats),
		Bottlenecks:  po.identifyBottlenecks(stats),
		Trends:       po.analyzeTrends(),
		Alerts:       po.checkAlerts(stats),
	}

	report.OverallScore = po.calculateOverallScore(report)
	report.Recommendations = po.generateRecommendations(report)

	po.recordObservation("performance_analysis", "info", 
		fmt.Sprintf("Performance analysis completed with score %d", report.OverallScore),
		map[string]interface{}{
			"score": report.OverallScore,
			"bottlenecks_count": len(report.Bottlenecks),
			"alerts_count": len(report.Alerts),
		}, "")

	po.logger.InfoContext(ctx, "Performance analysis completed",
		slog.Int("overall_score", report.OverallScore),
		slog.Int("bottlenecks", len(report.Bottlenecks)),
		slog.Int("alerts", len(report.Alerts)))

	return report
}

// analyzeSystemHealth analyzes overall system health
func (po *PerformanceOptimizer) analyzeSystemHealth(stats *SystemStats) SystemHealthStatus {
	memUsageMB := stats.MemoryUsage / 1024 / 1024
	
	// Calculate health score
	score := 100
	status := "healthy"
	
	if memUsageMB > po.alertThresholds.MemoryUsageMB {
		score -= 30
		status = "degraded"
	} else if memUsageMB > po.alertThresholds.MemoryUsageMB/2 {
		score -= 15
	}
	
	if stats.GoRoutines > po.alertThresholds.GoroutineCount {
		score -= 25
		status = "degraded"
	} else if stats.GoRoutines > po.alertThresholds.GoroutineCount/2 {
		score -= 10
	}
	
	if stats.LastGCPause > time.Duration(po.alertThresholds.GCPauseMS)*time.Millisecond {
		score -= 20
		if status == "healthy" {
			status = "warning"
		}
	}
	
	if score < 70 {
		status = "critical"
	} else if score < 85 && status == "healthy" {
		status = "warning"
	}
	
	return SystemHealthStatus{
		Status:         status,
		Score:          max(0, score),
		MemoryUsage:    memUsageMB,
		CPUUsage:       stats.CPUUsage,
		GoroutineCount: stats.GoRoutines,
		Uptime:         stats.ProcessUptime.Seconds(),
	}
}

// identifyBottlenecks identifies performance bottlenecks
func (po *PerformanceOptimizer) identifyBottlenecks(stats *SystemStats) []PerformanceBottleneck {
	var bottlenecks []PerformanceBottleneck
	
	memUsageMB := stats.MemoryUsage / 1024 / 1024
	
	// Memory bottlenecks
	if memUsageMB > po.alertThresholds.MemoryUsageMB {
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Type:        "memory",
			Severity:    "high",
			Description: "High memory usage detected",
			Impact:      "May cause out-of-memory errors and degraded performance",
			Metrics: map[string]interface{}{
				"current_mb": memUsageMB,
				"threshold_mb": po.alertThresholds.MemoryUsageMB,
				"usage_percent": float64(memUsageMB) / float64(po.alertThresholds.MemoryUsageMB) * 100,
			},
		})
	}
	
	// Goroutine bottlenecks
	if stats.GoRoutines > po.alertThresholds.GoroutineCount {
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Type:        "goroutines",
			Severity:    "medium",
			Description: "High goroutine count detected",
			Impact:      "May indicate goroutine leaks or inefficient concurrency patterns",
			Metrics: map[string]interface{}{
				"current_count": stats.GoRoutines,
				"threshold": po.alertThresholds.GoroutineCount,
				"ratio": float64(stats.GoRoutines) / float64(po.alertThresholds.GoroutineCount),
			},
		})
	}
	
	// GC bottlenecks
	if stats.LastGCPause > time.Duration(po.alertThresholds.GCPauseMS)*time.Millisecond {
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Type:        "garbage_collection",
			Severity:    "medium",
			Description: "Long garbage collection pauses detected",
			Impact:      "May cause request latency spikes and reduced throughput",
			Metrics: map[string]interface{}{
				"current_pause_ms": stats.LastGCPause.Milliseconds(),
				"threshold_ms": po.alertThresholds.GCPauseMS,
				"ratio": float64(stats.LastGCPause.Milliseconds()) / float64(po.alertThresholds.GCPauseMS),
			},
		})
	}
	
	return bottlenecks
}

// generateRecommendations generates performance optimization recommendations
func (po *PerformanceOptimizer) generateRecommendations(report *PerformanceReport) []PerformanceRecommendation {
	var recommendations []PerformanceRecommendation
	
	// Memory optimization recommendations
	if report.SystemHealth.MemoryUsage > po.alertThresholds.MemoryUsageMB/2 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Category:    "memory",
			Priority:    "high",
			Title:       "Optimize Memory Usage",
			Description: "Memory usage is high. Consider implementing memory optimization strategies.",
			Impact:      "Reduce memory pressure and improve application stability",
			Effort:      "medium",
			Action:      "Review memory allocation patterns, implement object pooling, and optimize data structures",
		})
	}
	
	// Goroutine optimization recommendations
	if report.SystemHealth.GoroutineCount > po.alertThresholds.GoroutineCount/2 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Category:    "concurrency",
			Priority:    "medium",
			Title:       "Optimize Goroutine Usage",
			Description: "High goroutine count detected. Review goroutine lifecycle management.",
			Impact:      "Reduce resource overhead and prevent goroutine leaks",
			Effort:      "medium",
			Action:      "Implement goroutine pools, review goroutine termination logic, and add goroutine monitoring",
		})
	}
	
	// GC optimization recommendations
	for _, bottleneck := range report.Bottlenecks {
		if bottleneck.Type == "garbage_collection" {
			recommendations = append(recommendations, PerformanceRecommendation{
				Category:    "gc",
				Priority:    "medium",
				Title:       "Optimize Garbage Collection",
				Description: "Long GC pauses are affecting performance. Consider GC tuning strategies.",
				Impact:      "Reduce latency spikes and improve throughput",
				Effort:      "low",
				Action:      "Tune GOGC environment variable, reduce allocation rate, or implement custom memory management",
			})
			break
		}
	}
	
	// General performance recommendations
	recommendations = append(recommendations, PerformanceRecommendation{
		Category:    "monitoring",
		Priority:    "low",
		Title:       "Enhance Performance Monitoring",
		Description: "Consider implementing additional performance monitoring and alerting.",
		Impact:      "Proactive identification of performance issues",
		Effort:      "low",
		Action:      "Add custom metrics, implement alerting rules, and create performance dashboards",
	})
	
	// Database optimization (if applicable)
	recommendations = append(recommendations, PerformanceRecommendation{
		Category:    "database",
		Priority:    "medium",
		Title:       "Optimize Database Operations",
		Description: "Review database query patterns and connection pooling.",
		Impact:      "Reduce database load and improve response times",
		Effort:      "medium",
		Action:      "Implement connection pooling, add query caching, and optimize slow queries",
	})
	
	return recommendations
}

// analyzeTrends analyzes performance trends
func (po *PerformanceOptimizer) analyzeTrends() PerformanceTrends {
	// In a real implementation, this would analyze historical data
	// For now, return placeholder trends
	return PerformanceTrends{
		MemoryTrend:     "stable",
		CPUTrend:        "stable",
		ResponseTrend:   "improving",
		ErrorTrend:      "stable",
		ThroughputTrend: "increasing",
	}
}

// checkAlerts checks for performance alerts
func (po *PerformanceOptimizer) checkAlerts(stats *SystemStats) []PerformanceAlert {
	var alerts []PerformanceAlert
	now := time.Now()
	
	memUsageMB := stats.MemoryUsage / 1024 / 1024
	
	if memUsageMB > po.alertThresholds.MemoryUsageMB {
		alerts = append(alerts, PerformanceAlert{
			Type:        "memory_usage",
			Severity:    "warning",
			Message:     "Memory usage exceeds threshold",
			Threshold:   fmt.Sprintf("%d MB", po.alertThresholds.MemoryUsageMB),
			CurrentValue: fmt.Sprintf("%d MB", memUsageMB),
			Timestamp:   now,
		})
	}
	
	if stats.GoRoutines > po.alertThresholds.GoroutineCount {
		alerts = append(alerts, PerformanceAlert{
			Type:        "goroutine_count",
			Severity:    "warning",
			Message:     "Goroutine count exceeds threshold",
			Threshold:   fmt.Sprintf("%d", po.alertThresholds.GoroutineCount),
			CurrentValue: fmt.Sprintf("%d", stats.GoRoutines),
			Timestamp:   now,
		})
	}
	
	if stats.LastGCPause > time.Duration(po.alertThresholds.GCPauseMS)*time.Millisecond {
		alerts = append(alerts, PerformanceAlert{
			Type:        "gc_pause",
			Severity:    "warning",
			Message:     "GC pause duration exceeds threshold",
			Threshold:   fmt.Sprintf("%d ms", po.alertThresholds.GCPauseMS),
			CurrentValue: fmt.Sprintf("%d ms", stats.LastGCPause.Milliseconds()),
			Timestamp:   now,
		})
	}
	
	return alerts
}

// calculateOverallScore calculates the overall performance score
func (po *PerformanceOptimizer) calculateOverallScore(report *PerformanceReport) int {
	score := report.SystemHealth.Score
	
	// Deduct points for bottlenecks
	for _, bottleneck := range report.Bottlenecks {
		switch bottleneck.Severity {
		case "high":
			score -= 15
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}
	
	// Deduct points for alerts
	for _, alert := range report.Alerts {
		switch alert.Severity {
		case "critical":
			score -= 20
		case "warning":
			score -= 10
		case "info":
			score -= 2
		}
	}
	
	return max(0, min(100, score))
}

// recordObservation records a performance observation
func (po *PerformanceOptimizer) recordObservation(obsType, severity, message string, metrics map[string]interface{}, suggestion string) {
	po.mu.Lock()
	defer po.mu.Unlock()
	
	observation := PerformanceObservation{
		Timestamp:  time.Now(),
		Type:       obsType,
		Severity:   severity,
		Message:    message,
		Metrics:    metrics,
		Suggestion: suggestion,
	}
	
	po.observations = append(po.observations, observation)
	
	// Keep only last 1000 observations
	if len(po.observations) > 1000 {
		po.observations = po.observations[len(po.observations)-1000:]
	}
}

// GetObservations returns recent performance observations
func (po *PerformanceOptimizer) GetObservations(limit int) []PerformanceObservation {
	po.mu.RLock()
	defer po.mu.RUnlock()
	
	if limit <= 0 || limit > len(po.observations) {
		limit = len(po.observations)
	}
	
	start := len(po.observations) - limit
	if start < 0 {
		start = 0
	}
	
	observations := make([]PerformanceObservation, limit)
	copy(observations, po.observations[start:])
	
	return observations
}

// SetThresholds updates performance thresholds
func (po *PerformanceOptimizer) SetThresholds(thresholds PerformanceThresholds) {
	po.mu.Lock()
	defer po.mu.Unlock()
	
	po.alertThresholds = thresholds
	
	po.logger.Info("Performance thresholds updated",
		slog.Int64("memory_mb", thresholds.MemoryUsageMB),
		slog.Int64("goroutines", thresholds.GoroutineCount),
		slog.Int64("gc_pause_ms", thresholds.GCPauseMS))
}

// OptimizeRuntime applies runtime optimizations
func (po *PerformanceOptimizer) OptimizeRuntime(ctx context.Context) {
	po.logger.InfoContext(ctx, "Applying runtime optimizations")
	
	// Force garbage collection if memory usage is high
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	memUsageMB := memStats.Alloc / 1024 / 1024
	if memUsageMB > uint64(po.alertThresholds.MemoryUsageMB)*3/4 {
		po.logger.InfoContext(ctx, "Forcing garbage collection due to high memory usage",
			slog.Uint64("memory_mb", memUsageMB))
		runtime.GC()
		
		po.recordObservation("runtime_optimization", "info",
			"Forced garbage collection due to high memory usage",
			map[string]interface{}{
				"memory_before_mb": memUsageMB,
			}, "Consider optimizing memory allocation patterns")
	}
	
	// Log goroutine count if high
	goroutines := runtime.NumGoroutine()
	if int64(goroutines) > po.alertThresholds.GoroutineCount*3/4 {
		po.logger.WarnContext(ctx, "High goroutine count detected",
			slog.Int("goroutines", goroutines))
		
		po.recordObservation("runtime_optimization", "warning",
			"High goroutine count detected",
			map[string]interface{}{
				"goroutine_count": goroutines,
			}, "Review goroutine lifecycle management")
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}