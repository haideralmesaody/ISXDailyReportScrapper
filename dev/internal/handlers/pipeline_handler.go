package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	
	apierrors "isxcli/internal/errors"
	"isxcli/internal/services"
)

// PipelineServiceInterface defines the interface for pipeline operations
type PipelineServiceInterface interface {
	StartPipeline(params map[string]interface{}) (string, error)
	StopPipeline(pipelineID string) error
	GetStatus(pipelineID string) (map[string]interface{}, error)
	ListPipelines() []map[string]interface{}
}

// PipelineHandler handles pipeline-related HTTP requests with RFC 7807 compliance
type PipelineHandler struct {
	service      PipelineServiceInterface
	logger       *slog.Logger
	errorHandler *apierrors.ErrorHandler
}

// NewPipelineHandler creates a new pipeline handler with RFC 7807 error handling
func NewPipelineHandler(service PipelineServiceInterface, logger *slog.Logger, errorHandler *apierrors.ErrorHandler) *PipelineHandler {
	return &PipelineHandler{
		service:      service,
		logger:       logger.With(slog.String("component", "pipeline_handler")),
		errorHandler: errorHandler,
	}
}

// Routes returns the pipeline routes with proper Chi patterns
func (h *PipelineHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	// Use render for consistent JSON responses
	r.Use(render.SetContentType(render.ContentTypeJSON))
	
	// Pipeline operations
	r.Post("/start", h.StartPipeline)
	r.Post("/stop", h.StopPipeline)
	r.Get("/status", h.GetStatus)
	r.Get("/pipelines", h.ListPipelines)
	
	// Individual pipeline routes
	r.Route("/{pipelineID}", func(r chi.Router) {
		r.Use(h.PipelineCtx) // Load pipeline into context
		r.Get("/", h.GetPipeline)
		r.Delete("/", h.DeletePipeline)
		r.Get("/logs", h.GetPipelineLogs)
	})
	
	return r
}

// PipelineCtx middleware validates and loads pipeline context
func (h *PipelineHandler) PipelineCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pipelineID := chi.URLParam(r, "pipelineID")
		if pipelineID == "" {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("pipelineID", "Pipeline ID is required"))
			return
		}
		
		// Validate pipeline ID format (e.g., UUID)
		if len(pipelineID) < 8 {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("pipelineID", "Invalid pipeline ID format"))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// StartPipelineRequest represents the request body for starting a pipeline
type StartPipelineRequest struct {
	Type       string                 `json:"type" validate:"required,oneof=scraping processing full"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Priority   string                 `json:"priority,omitempty" validate:"omitempty,oneof=low normal high"`
}

// StopPipelineRequest represents the request body for stopping a pipeline
type StopPipelineRequest struct {
	PipelineID string `json:"pipeline_id" validate:"required"`
	Force      bool   `json:"force,omitempty"`
}

// StartPipeline handles POST /api/pipeline/start with RFC 7807 errors
func (h *PipelineHandler) StartPipeline(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	// Decode and validate request
	var req StartPipelineRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.logger.WarnContext(r.Context(), "invalid request body",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		h.errorHandler.HandleError(w, r, apierrors.InvalidRequestWithError(err))
		return
	}
	
	// Set defaults
	if req.Priority == "" {
		req.Priority = "normal"
	}
	
	h.logger.InfoContext(r.Context(), "starting pipeline",
		slog.String("request_id", reqID),
		slog.String("type", req.Type),
		slog.String("priority", req.Priority),
	)
	
	// Prepare parameters
	params := make(map[string]interface{})
	params["type"] = req.Type
	params["priority"] = req.Priority
	for k, v := range req.Parameters {
		params[k] = v
	}
	
	// Start the pipeline
	pipelineID, err := h.service.StartPipeline(params)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to start pipeline",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("type", req.Type),
		)
		
		// Map service errors to API errors
		if errors.Is(err, services.ErrPipelineRunning) {
			h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
				http.StatusConflict,
				"PIPELINE_ALREADY_RUNNING",
				"A pipeline is already running",
				map[string]interface{}{
					"type": req.Type,
				},
			))
			return
		}
		
		if errors.Is(err, services.ErrInvalidInput) {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("parameters", err.Error()))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Success response
	h.logger.InfoContext(r.Context(), "pipeline started successfully",
		slog.String("pipeline_id", pipelineID),
		slog.String("request_id", reqID),
	)
	
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"status":      "success",
		"pipeline_id": pipelineID,
		"message":     fmt.Sprintf("Pipeline started successfully"),
		"type":        req.Type,
		"priority":    req.Priority,
	})
}

// StopPipeline handles POST /api/pipeline/stop with RFC 7807 errors
func (h *PipelineHandler) StopPipeline(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	// Decode and validate request
	var req StopPipelineRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.logger.WarnContext(r.Context(), "invalid request body",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		h.errorHandler.HandleError(w, r, apierrors.InvalidRequestWithError(err))
		return
	}
	
	if req.PipelineID == "" {
		h.errorHandler.HandleError(w, r, apierrors.ErrValidation("pipeline_id", "Pipeline ID is required"))
		return
	}
	
	h.logger.InfoContext(r.Context(), "stopping pipeline",
		slog.String("request_id", reqID),
		slog.String("pipeline_id", req.PipelineID),
		slog.Bool("force", req.Force),
	)
	
	// Stop the pipeline
	if err := h.service.StopPipeline(req.PipelineID); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to stop pipeline",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("pipeline_id", req.PipelineID),
		)
		
		if errors.Is(err, services.ErrPipelineNotFound) {
			h.errorHandler.HandleError(w, r, apierrors.NotFoundError("pipeline"))
			return
		}
		
		if errors.Is(err, services.ErrPipelineNotRunning) {
			h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
				http.StatusConflict,
				"PIPELINE_NOT_RUNNING",
				"Pipeline is not running",
				map[string]interface{}{
					"pipeline_id": req.PipelineID,
				},
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Success response
	h.logger.InfoContext(r.Context(), "pipeline stopped successfully",
		slog.String("pipeline_id", req.PipelineID),
		slog.String("request_id", reqID),
	)
	
	render.JSON(w, r, map[string]interface{}{
		"status":      "success",
		"pipeline_id": req.PipelineID,
		"message":     "Pipeline stopped successfully",
	})
}

// GetStatus handles GET /api/pipeline/status with RFC 7807 errors
func (h *PipelineHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	pipelineID := r.URL.Query().Get("pipeline_id")
	
	// If no specific pipeline ID, return overall status
	if pipelineID == "" {
		h.logger.InfoContext(r.Context(), "fetching overall pipeline status",
			slog.String("request_id", reqID),
		)
		
		// Get all pipelines and return summary
		pipelines := h.service.ListPipelines()
		
		// Calculate summary
		activePipelines := 0
		for _, p := range pipelines {
			if status, ok := p["status"].(string); ok && (status == "running" || status == "processing") {
				activePipelines++
			}
		}
		
		render.JSON(w, r, map[string]interface{}{
			"status":            "success",
			"active_pipelines":  activePipelines,
			"total_pipelines":   len(pipelines),
			"service_status":    "operational",
		})
		return
	}
	
	h.logger.InfoContext(r.Context(), "fetching pipeline status",
		slog.String("request_id", reqID),
		slog.String("pipeline_id", pipelineID),
	)
	
	status, err := h.service.GetStatus(pipelineID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get pipeline status",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("pipeline_id", pipelineID),
		)
		
		if errors.Is(err, services.ErrPipelineNotFound) {
			h.errorHandler.HandleError(w, r, apierrors.NotFoundError("pipeline"))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

// ListPipelines handles GET /api/pipeline/pipelines with RFC 7807 errors
func (h *PipelineHandler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	h.logger.InfoContext(r.Context(), "listing pipelines",
		slog.String("request_id", reqID),
	)
	
	// Get query parameters for filtering
	status := r.URL.Query().Get("status")
	pipelineType := r.URL.Query().Get("type")
	
	pipelines := h.service.ListPipelines()
	
	// Apply filters if provided
	var filtered []map[string]interface{}
	for _, p := range pipelines {
		// Filter by status
		if status != "" {
			if pStatus, ok := p["status"].(string); !ok || pStatus != status {
				continue
			}
		}
		
		// Filter by type
		if pipelineType != "" {
			if pType, ok := p["type"].(string); !ok || pType != pipelineType {
				continue
			}
		}
		
		filtered = append(filtered, p)
	}
	
	// If no filters applied, use original list
	if status == "" && pipelineType == "" {
		filtered = pipelines
	}
	
	h.logger.InfoContext(r.Context(), "pipelines retrieved",
		slog.String("request_id", reqID),
		slog.Int("total", len(pipelines)),
		slog.Int("filtered", len(filtered)),
	)
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   filtered,
		"count":  len(filtered),
		"filters": map[string]string{
			"status": status,
			"type":   pipelineType,
		},
	})
}

// GetPipeline handles GET /api/pipeline/{pipelineID} with RFC 7807 errors
func (h *PipelineHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	pipelineID := chi.URLParam(r, "pipelineID")
	
	h.logger.InfoContext(r.Context(), "fetching pipeline details",
		slog.String("request_id", reqID),
		slog.String("pipeline_id", pipelineID),
	)
	
	status, err := h.service.GetStatus(pipelineID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get pipeline details",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("pipeline_id", pipelineID),
		)
		
		if errors.Is(err, services.ErrPipelineNotFound) {
			h.errorHandler.HandleError(w, r, apierrors.NotFoundError("pipeline"))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   status,
	})
}

// DeletePipeline handles DELETE /api/pipeline/{pipelineID} with RFC 7807 errors
func (h *PipelineHandler) DeletePipeline(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	pipelineID := chi.URLParam(r, "pipelineID")
	
	h.logger.InfoContext(r.Context(), "deleting pipeline",
		slog.String("request_id", reqID),
		slog.String("pipeline_id", pipelineID),
	)
	
	// For now, just stop the pipeline
	// In a real implementation, this might clean up resources, logs, etc.
	if err := h.service.StopPipeline(pipelineID); err != nil {
		if !errors.Is(err, services.ErrPipelineNotFound) && !errors.Is(err, services.ErrPipelineNotRunning) {
			h.logger.ErrorContext(r.Context(), "failed to delete pipeline",
				slog.String("error", err.Error()),
				slog.String("request_id", reqID),
				slog.String("pipeline_id", pipelineID),
			)
			h.errorHandler.HandleError(w, r, err)
			return
		}
	}
	
	render.Status(r, http.StatusNoContent)
}

// GetPipelineLogs handles GET /api/pipeline/{pipelineID}/logs with RFC 7807 errors
func (h *PipelineHandler) GetPipelineLogs(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	pipelineID := chi.URLParam(r, "pipelineID")
	
	h.logger.InfoContext(r.Context(), "fetching pipeline logs",
		slog.String("request_id", reqID),
		slog.String("pipeline_id", pipelineID),
	)
	
	// This is a placeholder - in real implementation, fetch logs from service
	// For now, return empty logs
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"pipeline_id": pipelineID,
			"logs":        []string{},
			"has_more":    false,
		},
	})
}