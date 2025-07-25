package handlers

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"isxcli/internal/services"
)

// PipelineHandler handles pipeline-related HTTP requests
type PipelineHandler struct {
	service *services.PipelineService
	logger  services.Logger
}

// NewPipelineHandler creates a new pipeline handler
func NewPipelineHandler(service *services.PipelineService, logger services.Logger) *PipelineHandler {
	return &PipelineHandler{
		service: service,
		logger:  logger,
	}
}

// Routes returns the pipeline routes
func (h *PipelineHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	r.Post("/start", h.StartPipeline)
	r.Post("/stop", h.StopPipeline)
	r.Get("/status", h.GetStatus)
	r.Get("/pipelines", h.ListPipelines)
	
	return r
}

// StartPipeline handles POST /api/pipeline/start
func (h *PipelineHandler) StartPipeline(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if err := render.DecodeJSON(r.Body, &params); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	pipelineID, err := h.service.StartPipeline(params)
	if err != nil {
		h.logger.Error("Failed to start pipeline", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	render.JSON(w, r, map[string]interface{}{
		"pipeline_id": pipelineID,
		"status": "started",
	})
}

// StopPipeline handles POST /api/pipeline/stop
func (h *PipelineHandler) StopPipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PipelineID string `json:"pipeline_id"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": "Invalid request body",
		})
		return
	}

	if err := h.service.StopPipeline(req.PipelineID); err != nil {
		h.logger.Error("Failed to stop pipeline", "error", err, "pipeline_id", req.PipelineID)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "stopped",
	})
}

// GetStatus handles GET /api/pipeline/status
func (h *PipelineHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	pipelineID := r.URL.Query().Get("pipeline_id")
	status, err := h.service.GetStatus(pipelineID)
	if err != nil {
		h.logger.Error("Failed to get pipeline status", "error", err, "pipeline_id", pipelineID)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, status)
}

// ListPipelines handles GET /api/pipeline/pipelines
func (h *PipelineHandler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	pipelines := h.service.ListPipelines()
	render.JSON(w, r, pipelines)
}