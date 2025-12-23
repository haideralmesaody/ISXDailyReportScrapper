package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/operations"
	"github.com/isxcli/isxcli/internal/services"
)

// HealthHandler handles health-related HTTP requests
type HealthHandler struct {
	service      *services.HealthService
	manager      *operations.Manager
	jobQueue     *operations.JobQueue
	logger       *slog.Logger
	errorHandler *errors.ErrorHandler
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	service *services.HealthService,
	manager *operations.Manager,
	jobQueue *operations.JobQueue,
	logger *slog.Logger,
	errorHandler *errors.ErrorHandler,
) *HealthHandler {
	return &HealthHandler{
		service:      service,
		manager:      manager,
		jobQueue:     jobQueue,
		logger:       logger.With(slog.String("handler", "health")),
		errorHandler: errorHandler,
	}
}

// HealthCheck handles GET /api/health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.HealthCheck(r.Context()))
}

// ReadinessCheck handles GET /api/health/ready
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.ReadinessCheck(r.Context()))
}

// LivenessCheck handles GET /api/health/live
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.LivenessCheck(r.Context()))
}

// Version handles GET /api/version
func (h *HealthHandler) Version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.Version())
}

// LicenseStatus handles GET /api/license/status
func (h *HealthHandler) LicenseStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.LicenseStatus(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get license status",
			slog.String("error", err.Error()))
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, status)
}

// Routes returns chi.Router for health and diagnostics endpoints
func (h *HealthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Basic health endpoints
	r.Get("/health", h.HealthCheck)
	r.Get("/ready", h.ReadinessCheck)
	r.Get("/live", h.LivenessCheck)
	r.Get("/version", h.Version)
	r.Get("/license/status", h.LicenseStatus)

	// System diagnostics endpoint (moved from operations handler)
	r.Get("/operations/diagnostics", h.GetDiagnostics)
	r.Post("/operations/cleanup-stuck-jobs", h.CleanupStuckJobs)

	return r
}

// GetDiagnostics handles GET /api/operations/diagnostics with system-wide health information
// This endpoint aggregates system health from multiple components
func (h *HealthHandler) GetDiagnostics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	h.logger.InfoContext(ctx, "operations diagnostics requested",
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	// Check job queue health
	queueInfo := make(map[string]interface{})
	var jobs []*operations.Job
	var err error

	if h.jobQueue != nil {
		jobs, err = h.jobQueue.ListJobs(operations.JobFilter{})
		if err != nil {
			h.logger.ErrorContext(ctx, "failed to list jobs for diagnostics",
				slog.String("error", err.Error()),
				slog.String("request_id", reqID),
			)
			h.errorHandler.HandleError(w, r, errors.New(
				http.StatusInternalServerError,
				"DIAGNOSTICS_ERROR",
				"Failed to gather system diagnostics",
			))
			return
		}

		// Job statistics
		statusCount := make(map[string]int)
		for _, job := range jobs {
			statusCount[string(job.Status)]++
		}

		queueInfo = map[string]interface{}{
			"total_jobs":     len(jobs),
			"status_breakdown": statusCount,
			"queue_healthy":  h.jobQueue != nil,
		}
	}

	// Check operation manager health
	managerInfo := make(map[string]interface{})
	var operationsList []*operations.OperationState
	if h.manager != nil {
		operationsList = h.manager.ListOperations()

		statusCount := make(map[string]int)
		for _, op := range operationsList {
			statusCount[string(op.Status)]++
		}

		managerInfo = map[string]interface{}{
			"total_operations":    len(operationsList),
			"status_breakdown":    statusCount,
			"registry_healthy":   h.manager.GetRegistry() != nil,
		}
	}

	// Memory and system statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemInfo := map[string]interface{}{
		"alloc_mb":         bToMb(m.Alloc),
		"total_alloc_mb":   bToMb(m.TotalAlloc),
		"sys_mb":           bToMb(m.Sys),
		"num_gc":          m.NumGC,
		"num_goroutines":  runtime.NumGoroutine(),
	}

	// Response with all diagnostics
	response := map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
		"trace_id":  reqID,
		"components": map[string]interface{}{
			"job_queue":  queueInfo,
			"operations": managerInfo,
			"system":     systemInfo,
		},
	}

	h.logger.InfoContext(ctx, "diagnostics completed successfully",
		slog.String("request_id", reqID),
		slog.Int("total_jobs", len(jobs)),
		slog.Int("total_operations", len(operationsList)),
		slog.Int("num_goroutines", runtime.NumGoroutine()),
	)

	render.JSON(w, r, response)
}

// CleanupStuckJobs handles POST /api/operations/cleanup-stuck-jobs
// This is an admin endpoint for recovering from job queue issues
func (h *HealthHandler) CleanupStuckJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)

	logger := h.logger.With(
		slog.String("handler", "CleanupStuckJobs"),
		slog.String("request_id", reqID),
	)

	logger.InfoContext(ctx, "job cleanup requested",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	// Validate job queue availability
	if h.jobQueue == nil {
		logger.ErrorContext(ctx, "job queue not available")
		h.errorHandler.HandleError(w, r, errors.New(
			http.StatusServiceUnavailable,
			"JOB_QUEUE_UNAVAILABLE",
			"Job queue service is not available",
		))
		return
	}

	// Get stuck jobs (pending for more than 5 minutes)
	cutoffTime := time.Now().Add(-5 * time.Minute)
	filter := operations.JobFilter{
		Status: operations.JobStatusPending,
		Since:  cutoffTime,
		Limit:  100, // Limit to prevent runaway cleanup
	}

	stuckJobs, err := h.jobQueue.ListJobs(filter)
	if err != nil {
		logger.ErrorContext(ctx, "failed to list stuck jobs",
			slog.String("error", err.Error()),
			slog.String("filter_status", string(filter.Status)),
			slog.String("cutoff", cutoffTime.String()),
		)

		h.errorHandler.HandleError(w, r, errors.New(
			http.StatusInternalServerError,
			"CLEANUP_ERROR",
			"Failed to list stuck jobs",
		))
		return
	}

	cleanedCount := 0
	for _, job := range stuckJobs {
		// Mark stuck job as failed
		job.Status = operations.JobStatusFailed
		job.Error = "Job cleaned up due to stale timeout"
		completedAt := time.Now()
		job.CompletedAt = &completedAt

		if err := h.jobQueue.UpdateJob(job); err != nil {
			logger.WarnContext(ctx, "failed to update stuck job",
				slog.String("job_id", job.ID),
				slog.String("error", err.Error()))
			continue
		}

		cleanedCount++
		logger.InfoContext(ctx, "cleaned up stuck job",
			slog.String("job_id", job.ID),
			slog.String("stage_id", job.StageID),
			slog.Duration("stuck_duration", time.Since(job.CreatedAt)),
		)
	}

	logger.InfoContext(ctx, "cleanup completed",
		slog.Int("cleaned_count", cleanedCount),
		slog.Int("total_stuck", len(stuckJobs)),
		slog.Time("cutoff_time", cutoffTime),
		slog.String("request_id", reqID),
	)

	response := map[string]interface{}{
		"status":       "success",
		"message":      fmt.Sprintf("Cleaned up %d stuck jobs", cleanedCount),
		"cleaned_count": cleanedCount,
		"total_stuck":  len(stuckJobs),
		"cutoff_time":   cutoffTime.Format(time.RFC3339),
		"timestamp":     time.Now().Format(time.RFC3339),
		"trace_id":      reqID,
	}

	render.JSON(w, r, response)
}

// Helper function to convert bytes to megabytes
func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}