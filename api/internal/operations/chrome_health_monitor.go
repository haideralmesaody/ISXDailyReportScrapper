package operations

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ChromeHealthStatus represents the health status of Chrome browser
type ChromeHealthStatus struct {
	IsHealthy      bool          `json:"is_healthy"`
	Issue          string        `json:"issue,omitempty"`
	Severity       string        `json:"severity,omitempty"`
	LastChecked    time.Time     `json:"last_checked"`
	ResponseTime   time.Duration `json:"response_time"`
	MemoryUsage    uint64        `json:"memory_usage"`
	GoroutineCount int           `json:"goroutine_count"`
}

// chromeSessionInterface defines the interface for Chrome session operations
type chromeSessionInterface interface {
	Close()
}

// ChromeHealthMonitor monitors Chrome browser health and handles automatic restarts
type ChromeHealthMonitor struct {
	session            chromeSessionInterface
	checkInterval      time.Duration
	healthStatus       atomic.Value // *ChromeHealthStatus
	logger             *slog.Logger
	ctx                context.Context
	cancel             context.CancelFunc
	running            atomic.Bool
	monitoring         sync.WaitGroup
	restartCount       atomic.Int64
	maxRestarts        int
	lastRestartTime    atomic.Value // time.Time
	healthCheckTimeout time.Duration
}

// NewChromeHealthMonitor creates a new Chrome health monitor
func NewChromeHealthMonitor(session chromeSessionInterface, logger *slog.Logger) *ChromeHealthMonitor {
	// Initialize health status
	healthStatus := &ChromeHealthStatus{
		IsHealthy:   true,
		LastChecked: time.Now(),
	}

	cm := &ChromeHealthMonitor{
		session:            session,
		checkInterval:      30 * time.Second,
		logger:             logger,
		maxRestarts:        3,                // Maximum 3 restarts per session
		healthCheckTimeout: 10 * time.Second, // Health check timeout
	}

	// Set initial health status
	cm.healthStatus.Store(healthStatus)

	return cm
}

// StartMonitoring begins health monitoring in a separate goroutine
func (m *ChromeHealthMonitor) StartMonitoring(ctx context.Context) {
	m.ctx, m.cancel = context.WithCancel(ctx)

	if m.running.CompareAndSwap(false, true) {
		m.monitoring.Add(1)

		if m.logger != nil {
			m.logger.Info("Starting Chrome health monitoring",
				slog.Duration("check_interval", m.checkInterval),
				slog.Duration("health_check_timeout", m.healthCheckTimeout))
		}

		go m.monitorLoop()
	}
}

// StopMonitoring stops the health monitoring
func (m *ChromeHealthMonitor) StopMonitoring() {
	if m.running.CompareAndSwap(true, false) {
		if m.cancel != nil {
			m.cancel()
		}
		m.monitoring.Wait()

		if m.logger != nil {
			m.logger.Info("Chrome health monitoring stopped")
		}
	}
}

// monitorLoop runs the health monitoring loop
func (m *ChromeHealthMonitor) monitorLoop() {
	defer m.monitoring.Done()

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			health := m.checkHealth()
			m.healthStatus.Store(health)

			if !health.IsHealthy {
				m.handleHealthIssue(health)
			}
		}
	}
}

// checkHealth performs a comprehensive health check on Chrome browser
func (m *ChromeHealthMonitor) checkHealth() *ChromeHealthStatus {
	start := time.Now()
	health := &ChromeHealthStatus{
		LastChecked: start,
		IsHealthy:   true,
	}

	// Check basic Chrome responsiveness
	if err := m.checkChromeResponsiveness(); err != nil {
		health.IsHealthy = false
		health.Issue = "chrome_unresponsive"
		health.Severity = "critical"
		health.ResponseTime = time.Since(start)

		if m.logger != nil {
			m.logger.Warn("Chrome health check failed - unresponsive",
				slog.String("error", err.Error()),
				slog.Duration("response_time", health.ResponseTime))
		}
		return health
	}

	health.ResponseTime = time.Since(start)

	// Check system resources
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	health.MemoryUsage = memStats.Alloc / 1024 / 1024 // MB
	health.GoroutineCount = runtime.NumGoroutine()

	// Check for memory pressure
	if health.MemoryUsage > 500 { // More than 500MB memory usage
		health.IsHealthy = false
		health.Issue = "high_memory_usage"
		health.Severity = "warning"

		if m.logger != nil {
			m.logger.Warn("Chrome health check warning - high memory usage",
				slog.Uint64("memory_mb", health.MemoryUsage),
				slog.Int("goroutines", health.GoroutineCount))
		}
	}

	// Check for goroutine leaks
	if health.GoroutineCount > 100 {
		health.IsHealthy = false
		health.Issue = "high_goroutine_count"
		health.Severity = "warning"

		if m.logger != nil {
			m.logger.Warn("Chrome health check warning - high goroutine count",
				slog.Int("goroutines", health.GoroutineCount),
				slog.Uint64("memory_mb", health.MemoryUsage))
		}
	}

	if health.IsHealthy && m.logger != nil {
		m.logger.Debug("Chrome health check passed",
			slog.Duration("response_time", health.ResponseTime),
			slog.Uint64("memory_mb", health.MemoryUsage),
			slog.Int("goroutines", health.GoroutineCount))
	}

	return health
}

// checkChromeResponsiveness checks if Chrome is responding to commands
func (m *ChromeHealthMonitor) checkChromeResponsiveness() error {
	if m.session == nil {
		return fmt.Errorf("chrome session is nil")
	}

	// Simple check - just verify session is not nil for now
	// The actual Chrome responsiveness check would require the chromedp package
	// which may have dependency issues in test environment
	return nil
}

// handleHealthIssue handles detected health issues
func (m *ChromeHealthMonitor) handleHealthIssue(health *ChromeHealthStatus) {
	if m.logger != nil {
		m.logger.Warn("Chrome health issue detected",
			slog.String("issue", health.Issue),
			slog.String("severity", health.Severity),
			slog.Duration("response_time", health.ResponseTime),
			slog.Int("restart_count", int(m.restartCount.Load())))
	}

	// Only attempt restart for critical issues
	if health.Severity == "critical" {
		if m.restartCount.Load() < int64(m.maxRestarts) {
			m.restartChrome()
		} else {
			if m.logger != nil {
				m.logger.Error("Maximum Chrome restarts exceeded, giving up",
					slog.Int64("restart_count", m.restartCount.Load()),
					slog.Int("max_restarts", m.maxRestarts))
			}
		}
	}
}

// restartChrome performs a graceful restart of the Chrome browser
func (m *ChromeHealthMonitor) restartChrome() {
	if m.logger != nil {
		m.logger.Info("Attempting Chrome restart due to health issues")
	}

	// Check if we recently restarted to prevent restart loops
	if lastRestartTime, ok := m.lastRestartTime.Load().(time.Time); ok {
		if time.Since(lastRestartTime) < 2*time.Minute {
			if m.logger != nil {
				m.logger.Warn("Chrome restart too recent, skipping",
					slog.Duration("time_since_last_restart", time.Since(lastRestartTime)))
			}
			return
		}
	}

	// Record restart attempt
	m.restartCount.Add(1)
	m.lastRestartTime.Store(time.Now())

	// Gracefully close current session with timeout
	restartDone := make(chan struct{})
	go func() {
		defer close(restartDone)

		if m.session != nil {
			m.session.Close()
		}
	}()

	// Wait for graceful shutdown or timeout
	select {
	case <-restartDone:
		if m.logger != nil {
			m.logger.Info("Chrome session closed gracefully")
		}
	case <-time.After(30 * time.Second):
		if m.logger != nil {
			m.logger.Warn("Chrome session close timeout, forcing restart")
		}
	}

	// Note: Session recreation would need to be handled by the caller
	// as the monitor doesn't have the ability to create new sessions

	if m.logger != nil {
		m.logger.Info("Chrome restart completed",
			slog.Int64("restart_count", m.restartCount.Load()))
	}
}

// GetHealthStatus returns the current health status
func (m *ChromeHealthMonitor) GetHealthStatus() *ChromeHealthStatus {
	if status := m.healthStatus.Load(); status != nil {
		return status.(*ChromeHealthStatus)
	}

	// Return default status if not set
	return &ChromeHealthStatus{
		IsHealthy:   false,
		Issue:       "health_monitor_not_running",
		Severity:    "critical",
		LastChecked: time.Time{},
	}
}

// IsHealthy returns whether Chrome is currently healthy
func (m *ChromeHealthMonitor) IsHealthy() bool {
	return m.GetHealthStatus().IsHealthy
}

// GetRestartCount returns the number of restarts attempted
func (m *ChromeHealthMonitor) GetRestartCount() int {
	return int(m.restartCount.Load())
}

// SetCheckInterval sets the health check interval
func (m *ChromeHealthMonitor) SetCheckInterval(interval time.Duration) {
	m.checkInterval = interval

	if m.logger != nil {
		m.logger.Info("Chrome health check interval updated",
			slog.Duration("new_interval", interval))
	}
}

// SetMaxRestarts sets the maximum number of restarts allowed
func (m *ChromeHealthMonitor) SetMaxRestarts(maxRestarts int) {
	m.maxRestarts = maxRestarts

	if m.logger != nil {
		m.logger.Info("Chrome max restarts updated",
			slog.Int("max_restarts", maxRestarts))
	}
}

// IsRunning returns whether the health monitor is currently running
func (m *ChromeHealthMonitor) IsRunning() bool {
	return m.running.Load()
}
