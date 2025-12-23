package infrastructure

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// SystemMetrics provides system resource monitoring
type SystemMetrics struct {
	meter metric.Meter

	// Runtime metrics
	goRoutines           metric.Int64Gauge
	memoryUsage          metric.Int64Gauge
	memoryAllocated      metric.Int64Gauge
	memorySystem         metric.Int64Gauge
	gcCount              metric.Int64Counter
	gcPause              metric.Float64Histogram

	// CPU metrics
	cpuUsage             metric.Float64Gauge
	cpuCount             metric.Int64Gauge

	// Process metrics
	processUptime        metric.Float64Gauge
	fileDescriptors      metric.Int64Gauge
	openConnections      metric.Int64Gauge
}

// NewSystemMetrics creates a new system metrics collector
func NewSystemMetrics(meter metric.Meter) (*SystemMetrics, error) {
	// Go runtime metrics
	goRoutines, err := meter.Int64Gauge(
		"system_goroutines",
		metric.WithDescription("Number of active goroutines"),
	)
	if err != nil {
		return nil, err
	}

	memoryUsage, err := meter.Int64Gauge(
		"system_memory_usage_bytes",
		metric.WithDescription("Memory usage in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	memoryAllocated, err := meter.Int64Gauge(
		"system_memory_allocated_bytes",
		metric.WithDescription("Memory allocated by Go runtime in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	memorySystem, err := meter.Int64Gauge(
		"system_memory_system_bytes",
		metric.WithDescription("Memory obtained from the OS in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	gcCount, err := meter.Int64Counter(
		"system_gc_count_total",
		metric.WithDescription("Total number of garbage collections"),
	)
	if err != nil {
		return nil, err
	}

	gcPause, err := meter.Float64Histogram(
		"system_gc_pause_seconds",
		metric.WithDescription("Garbage collection pause duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// CPU metrics
	cpuUsage, err := meter.Float64Gauge(
		"system_cpu_usage_percent",
		metric.WithDescription("CPU usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return nil, err
	}

	cpuCount, err := meter.Int64Gauge(
		"system_cpu_count",
		metric.WithDescription("Number of logical CPUs"),
	)
	if err != nil {
		return nil, err
	}

	// Process metrics
	processUptime, err := meter.Float64Gauge(
		"system_process_uptime_seconds",
		metric.WithDescription("Process uptime in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	fileDescriptors, err := meter.Int64Gauge(
		"system_file_descriptors",
		metric.WithDescription("Number of open file descriptors"),
	)
	if err != nil {
		return nil, err
	}

	openConnections, err := meter.Int64Gauge(
		"system_open_connections",
		metric.WithDescription("Number of open network connections"),
	)
	if err != nil {
		return nil, err
	}

	return &SystemMetrics{
		meter:               meter,
		goRoutines:          goRoutines,
		memoryUsage:         memoryUsage,
		memoryAllocated:     memoryAllocated,
		memorySystem:        memorySystem,
		gcCount:             gcCount,
		gcPause:             gcPause,
		cpuUsage:            cpuUsage,
		cpuCount:            cpuCount,
		processUptime:       processUptime,
		fileDescriptors:     fileDescriptors,
		openConnections:     openConnections,
	}, nil
}

// SystemStats holds current system statistics
type SystemStats struct {
	GoRoutines      int64
	MemoryUsage     int64
	MemoryAllocated int64
	MemorySystem    int64
	GCCount         uint32
	LastGCPause     time.Duration
	CPUUsage        float64
	CPUCount        int
	ProcessUptime   time.Duration
	FileDescriptors int64
	OpenConnections int64
	Timestamp       time.Time
}

// Collect collects and records system metrics
func (sm *SystemMetrics) Collect(ctx context.Context, startTime time.Time) *SystemStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := &SystemStats{
		GoRoutines:      int64(runtime.NumGoroutine()),
		MemoryUsage:     int64(memStats.Alloc),
		MemoryAllocated: int64(memStats.TotalAlloc),
		MemorySystem:    int64(memStats.Sys),
		GCCount:         memStats.NumGC,
		LastGCPause:     time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
		CPUCount:        runtime.NumCPU(),
		ProcessUptime:   time.Since(startTime),
		Timestamp:       time.Now(),
	}

	// Record runtime metrics
	sm.goRoutines.Record(ctx, stats.GoRoutines)
	sm.memoryUsage.Record(ctx, stats.MemoryUsage)
	sm.memoryAllocated.Record(ctx, stats.MemoryAllocated)
	sm.memorySystem.Record(ctx, stats.MemorySystem)
	sm.cpuCount.Record(ctx, int64(stats.CPUCount))
	sm.processUptime.Record(ctx, stats.ProcessUptime.Seconds())

	// Record GC metrics if there was a collection
	if stats.LastGCPause > 0 {
		sm.gcPause.Record(ctx, stats.LastGCPause.Seconds())
	}

	return stats
}

// RecordGCEvent records a garbage collection event
func (sm *SystemMetrics) RecordGCEvent(ctx context.Context, pauseDuration time.Duration, generation int) {
	attrs := []attribute.KeyValue{
		attribute.Int("gc_generation", generation),
	}

	sm.gcCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	sm.gcPause.Record(ctx, pauseDuration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordCPUUsage records CPU usage percentage
func (sm *SystemMetrics) RecordCPUUsage(ctx context.Context, usage float64) {
	sm.cpuUsage.Record(ctx, usage)
}

// RecordFileDescriptors records the number of open file descriptors
func (sm *SystemMetrics) RecordFileDescriptors(ctx context.Context, count int64) {
	sm.fileDescriptors.Record(ctx, count)
}

// RecordOpenConnections records the number of open network connections
func (sm *SystemMetrics) RecordOpenConnections(ctx context.Context, count int64) {
	sm.openConnections.Record(ctx, count)
}

// FormatStats returns a human-readable representation of system stats
func (stats *SystemStats) FormatStats() map[string]interface{} {
	return map[string]interface{}{
		"runtime": map[string]interface{}{
			"goroutines":       stats.GoRoutines,
			"memory_usage_mb":  stats.MemoryUsage / 1024 / 1024,
			"memory_alloc_mb":  stats.MemoryAllocated / 1024 / 1024,
			"memory_system_mb": stats.MemorySystem / 1024 / 1024,
			"gc_count":         stats.GCCount,
			"last_gc_pause_ms": stats.LastGCPause.Milliseconds(),
		},
		"system": map[string]interface{}{
			"cpu_count":           stats.CPUCount,
			"cpu_usage_percent":   stats.CPUUsage,
			"uptime_seconds":      stats.ProcessUptime.Seconds(),
			"file_descriptors":    stats.FileDescriptors,
			"open_connections":    stats.OpenConnections,
		},
		"timestamp": stats.Timestamp.Format(time.RFC3339),
	}
}

// SystemMetricsCollector manages periodic system metrics collection
type SystemMetricsCollector struct {
	metrics   *SystemMetrics
	startTime time.Time
	interval  time.Duration
	stopCh    chan struct{}
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector(meter metric.Meter, interval time.Duration) (*SystemMetricsCollector, error) {
	metrics, err := NewSystemMetrics(meter)
	if err != nil {
		return nil, fmt.Errorf("failed to create system metrics: %w", err)
	}

	return &SystemMetricsCollector{
		metrics:   metrics,
		startTime: time.Now(),
		interval:  interval,
		stopCh:    make(chan struct{}),
	}, nil
}

// Start begins periodic metrics collection
func (smc *SystemMetricsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(smc.interval)
	defer ticker.Stop()

	// Collect initial metrics
	smc.metrics.Collect(ctx, smc.startTime)

	for {
		select {
		case <-ticker.C:
			smc.metrics.Collect(ctx, smc.startTime)
		case <-smc.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops the metrics collection
func (smc *SystemMetricsCollector) Stop() {
	close(smc.stopCh)
}

// GetCurrentStats returns the current system statistics
func (smc *SystemMetricsCollector) GetCurrentStats(ctx context.Context) *SystemStats {
	return smc.metrics.Collect(ctx, smc.startTime)
}

// GetMetrics returns the underlying metrics instance
func (smc *SystemMetricsCollector) GetMetrics() *SystemMetrics {
	return smc.metrics
}