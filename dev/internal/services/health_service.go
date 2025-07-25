package services

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"isxcli/internal/config"
	"isxcli/internal/license"
	"isxcli/internal/pipeline"
	ws "isxcli/internal/websocket"
)

// HealthService provides health check functionality
type HealthService struct {
	version        string
	repoURL        string
	paths          config.PathsConfig
	licenseManager *license.Manager
	pipeline       *pipeline.Manager
	webSocketHub   *ws.Hub
	logger         Logger
	startTime      time.Time
}

// HealthStatus represents the health status response
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Runtime   map[string]interface{} `json:"runtime,omitempty"`
	Services  map[string]interface{} `json:"services,omitempty"`
}

// ServiceHealth represents individual service health
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// LicenseStatus represents license information
type LicenseStatus struct {
	IsValid   bool   `json:"is_valid"`
	DaysLeft  int    `json:"days_left"`
	ExpiryDate string `json:"expiry_date"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

// SystemStats represents system statistics
type SystemStats struct {
	UptimeSeconds    float64 `json:"uptime_seconds"`
	TotalFiles       int     `json:"total_files"`
	TotalSizeBytes   int64   `json:"total_size_bytes"`
	WebSocketClients int     `json:"websocket_clients"`
	ActivePipelines  int     `json:"active_pipelines"`
	GoVersion        string  `json:"go_version"`
	OS               string  `json:"os"`
	Arch             string  `json:"arch"`
}

// NewHealthService creates a new health service
func NewHealthService(version, repoURL string, paths config.PathsConfig, licenseManager *license.Manager, pipeline *pipeline.Manager, webSocketHub *ws.Hub, logger Logger) *HealthService {
	return &HealthService{
		version:        version,
		repoURL:        repoURL,
		paths:          paths,
		licenseManager: licenseManager,
		pipeline:       pipeline,
		webSocketHub:   webSocketHub,
		logger:         logger,
		startTime:      time.Now(),
	}
}

// HealthCheck returns overall health status
func (hs *HealthService) HealthCheck() HealthStatus {
	return HealthStatus{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   hs.version,
	}
}

// ReadinessCheck returns readiness status
func (hs *HealthService) ReadinessCheck() HealthStatus {
	status := HealthStatus{
		Status:    "ready",
		Timestamp: time.Now(),
		Version:   hs.version,
		Services:  make(map[string]interface{}),
	}

	// Check individual services
	status.Services["license"] = hs.checkLicenseHealth()
	status.Services["websocket"] = hs.checkWebSocketHealth()
	status.Services["pipeline"] = hs.checkPipelineHealth()
	status.Services["data"] = hs.checkDataHealth()

	// Determine overall readiness
	allReady := true
	for _, service := range status.Services {
		if sh, ok := service.(ServiceHealth); ok && sh.Status != "ready" {
			allReady = false
			break
		}
	}

	if !allReady {
		status.Status = "not_ready"
	}

	return status
}

// LivenessCheck returns liveness status
func (hs *HealthService) LivenessCheck() HealthStatus {
	return HealthStatus{
		Status:    "alive",
		Timestamp: time.Now(),
		Version:   hs.version,
		Runtime: map[string]interface{}{
			"uptime": time.Since(hs.startTime).Seconds(),
			"go_version": runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
		},
	}
}

// Version returns version information
func (hs *HealthService) Version() map[string]interface{} {
	return map[string]interface{}{
		"version":     hs.version,
		"build_time":  time.Now().Format(time.RFC3339),
		"go_version":  runtime.Version(),
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"repo_url":    hs.repoURL,
		"uptime":      time.Since(hs.startTime).Seconds(),
		"start_time":  hs.startTime.Format(time.RFC3339),
	}
}

// LicenseStatus returns license information
func (hs *HealthService) LicenseStatus() (LicenseStatus, error) {
	info, err := hs.licenseManager.GetLicenseInfo()
	if err != nil {
		return LicenseStatus{
			IsValid: false,
			Status:  "error",
			Message: err.Error(),
		}, nil
	}

	now := time.Now().UTC()
	expiryUTC := info.ExpiryDate.UTC()
	daysLeft := int(expiryUTC.Sub(now).Hours() / 24)

	isValid := daysLeft > 0 && (info.Status == "Activated" || info.Status == "Active" || info.Status == "Valid")

	return LicenseStatus{
		IsValid:   isValid,
		DaysLeft:  daysLeft,
		ExpiryDate: info.ExpiryDate.Format("2006-01-02"),
		Status:    info.Status,
	}, nil
}

// SystemStats returns system statistics
func (hs *HealthService) SystemStats() (SystemStats, error) {
	dataDir := hs.paths.DataDir

	var totalFiles int
	var totalSize int64

	// Count files and calculate size
	filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			totalFiles++
			totalSize += info.Size()
		}
		return nil
	})

	return SystemStats{
		UptimeSeconds:    time.Since(hs.startTime).Seconds(),
		TotalFiles:       totalFiles,
		TotalSizeBytes:   totalSize,
		WebSocketClients: hs.webSocketHub.ClientCount(),
		ActivePipelines:  len(hs.pipeline.ListPipelines()),
		GoVersion:        runtime.Version(),
		OS:               runtime.GOOS,
		Arch:             runtime.GOARCH,
	}, nil
}

// checkLicenseHealth checks license service health
func (hs *HealthService) checkLicenseHealth() ServiceHealth {
	_, err := hs.licenseManager.GetLicenseInfo()
	if err != nil {
		return ServiceHealth{
			Status:  "not_ready",
			Message: fmt.Sprintf("License error: %v", err),
		}
	}

	return ServiceHealth{
		Status: "ready",
		Message: "License service is healthy",
	}
}

// checkWebSocketHealth checks WebSocket service health
func (hs *HealthService) checkWebSocketHealth() ServiceHealth {
	// WebSocket hub is always considered healthy if it's running
	return ServiceHealth{
		Status: "ready",
		Message: "WebSocket service is healthy",
		Uptime: time.Since(hs.startTime).String(),
	}
}

// checkPipelineHealth checks pipeline service health
func (hs *HealthService) checkPipelineHealth() ServiceHealth {
	// Check if pipeline manager is initialized
	if hs.pipeline == nil {
		return ServiceHealth{
			Status:  "not_ready",
			Message: "Pipeline manager not initialized",
		}
	}

	return ServiceHealth{
		Status: "ready",
		Message: "Pipeline service is healthy",
	}
}

// checkDataHealth checks data service health
func (hs *HealthService) checkDataHealth() ServiceHealth {
	// Check if data directories exist and are accessible
	dataDir := hs.paths.DataDir
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return ServiceHealth{
			Status:  "not_ready",
			Message: fmt.Sprintf("Data directory not found: %s", dataDir),
		}
	}

	// Check if we can write to data directory
	if err := os.MkdirAll(filepath.Join(dataDir, "test"), 0755); err != nil {
		return ServiceHealth{
			Status:  "not_ready",
			Message: fmt.Sprintf("Cannot write to data directory: %v", err),
		}
	}

	return ServiceHealth{
		Status: "ready",
		Message: "Data service is healthy",
	}
}

// GetDetailedHealth returns comprehensive health information
func (hs *HealthService) GetDetailedHealth() map[string]interface{} {
	licenseStatus, _ := hs.LicenseStatus()
	stats, _ := hs.SystemStats()

	return map[string]interface{}{
		"health":   hs.HealthCheck(),
		"readiness": hs.ReadinessCheck(),
		"liveness":  hs.LivenessCheck(),
		"license":   licenseStatus,
		"stats":     stats,
	}
}