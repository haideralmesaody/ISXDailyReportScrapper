package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	errors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/operations"
	"github.com/isxcli/isxcli/internal/services"
)

// OperationsHandler exposes the minimal operations endpoints needed by the frontend.
// It intentionally stays thin: delegates to OperationService and returns JSON.
type OperationsHandler struct {
	service      *services.OperationService
	logger       *slog.Logger
	errorHandler *errors.ErrorHandler
}

// NewOperationsHandler creates a new operations handler.
func NewOperationsHandler(service *services.OperationService, logger *slog.Logger, errHandler *errors.ErrorHandler) *OperationsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if errHandler == nil {
		errHandler = errors.NewErrorHandler(logger, false)
	}
	return &OperationsHandler{
		service:      service,
		logger:       logger.With(slog.String("handler", "operations")),
		errorHandler: errHandler,
	}
}

// Routes returns the router for /api/operations endpoints.
func (h *OperationsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	// Keep a generous timeout for long-running stage starts.
	r.Use(middleware.Timeout(5 * time.Minute))

	r.Get("/", h.ListOperations)
	r.Get("/types", h.GetOperationTypes)
	// Start operations is retained for backward compatibility; keep as thin pass-through.
	r.Post("/start", h.StartOperation)
	r.Get("/{id}", h.GetOperation)
	r.Get("/{id}/status", h.GetOperationStatus)

	return r
}

// StageRoutes returns the router for /api/stages endpoints.
func (h *OperationsHandler) StageRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(5 * time.Minute))

	r.Get("/", h.GetStages)
	// Dedicated stage execution endpoints for the first four stages
	r.Post("/scraping/execute", h.ExecuteScrapingStage)
	r.Post("/processing/execute", h.ExecuteProcessingStage)
	r.Post("/indices/execute", h.ExecuteIndicesStage)
	r.Post("/liquidity/execute", h.ExecuteLiquidityStage)

	r.Post("/{id}/execute", h.ExecuteStage)

	return r
}

// GetOperationTypes handles GET /api/operations/types
func (h *OperationsHandler) GetOperationTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	types, err := h.service.GetOperationTypes(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	render.JSON(w, r, types)
}

// ListOperations handles GET /api/operations
func (h *OperationsHandler) ListOperations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ops, err := h.service.ListOperations(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	render.JSON(w, r, ops)
}

// GetOperation handles GET /api/operations/{id}
func (h *OperationsHandler) GetOperation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("operation id is required"))
		return
	}

	state, err := h.service.GetOperationStatus(r.Context(), id)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	render.JSON(w, r, state)
}

// GetOperationStatus handles GET /api/operations/{id}/status
func (h *OperationsHandler) GetOperationStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("operation id is required"))
		return
	}
	state, err := h.service.GetOperationStatus(r.Context(), id)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	render.JSON(w, r, state)
}

// StartOperation handles POST /api/operations/start
// Accepts a generic JSON payload and forwards to the OperationService.
func (h *OperationsHandler) StartOperation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.errorHandler.HandleError(w, r, errors.NewValidationError(fmt.Sprintf("invalid JSON payload: %v", err)))
		return
	}

	id, err := h.service.StartOperation(ctx, payload)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	render.JSON(w, r, map[string]interface{}{
		"id":     id,
		"status": "started",
	})
}

// GetStages handles GET /api/stages
func (h *OperationsHandler) GetStages(w http.ResponseWriter, r *http.Request) {
	summary := h.service.GetStageSummary()
	render.JSON(w, r, summary)
}

// ExecuteStage handles POST /api/stages/{id}/execute
// It converts the request into a run_stage operation for simplified single-stage execution.
func (h *OperationsHandler) ExecuteStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stageID := chi.URLParam(r, "id")
	if stageID == "" {
		h.errorHandler.HandleError(w, r, errors.NewValidationError("stage id is required"))
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
		h.errorHandler.HandleError(w, r, errors.NewValidationError(fmt.Sprintf("invalid JSON payload: %v", err)))
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	// Prepare an operation request compatible with ProcessSimplifiedOperationRequest
	req := operations.OperationRequest{
		ID:         fmt.Sprintf("stage-%s-%d", stageID, time.Now().Unix()),
		Mode:       "run_stage",
		Parameters: map[string]interface{}{"stage_id": stageID},
	}
	for k, v := range payload {
		req.Parameters[k] = v
	}

	resp, err := h.service.ExecuteOperation(ctx, &req)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	render.JSON(w, r, resp)
}

// executeSpecificStage is a helper to run a named stage with a typed endpoint.
func (h *OperationsHandler) executeSpecificStage(w http.ResponseWriter, r *http.Request, stageID string) {
	ctx := r.Context()
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
		h.errorHandler.HandleError(w, r, errors.NewValidationError(fmt.Sprintf("invalid JSON payload: %v", err)))
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	// Ensure the target stage is explicitly set for the service layer.
	payload["stage_id"] = stageID
	payload["step"] = stageID

	// Build a minimal request targeting the specific stage.
	req := operations.OperationRequest{
		ID:         fmt.Sprintf("stage-%s-%d", stageID, time.Now().Unix()),
		Mode:       stageID,
		Parameters: payload,
	}
	// Ensure stage_id/step are set for the service layer.
	req.Parameters["stage_id"] = stageID
	req.Parameters["step"] = stageID

	resp, err := h.service.ExecuteOperation(ctx, &req)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	render.JSON(w, r, resp)
}

// ExecuteScrapingStage handles POST /api/stages/scraping/execute
func (h *OperationsHandler) ExecuteScrapingStage(w http.ResponseWriter, r *http.Request) {
	h.executeSpecificStage(w, r, operations.StageIDScraping)
}

// ExecuteProcessingStage handles POST /api/stages/processing/execute
func (h *OperationsHandler) ExecuteProcessingStage(w http.ResponseWriter, r *http.Request) {
	h.executeSpecificStage(w, r, operations.StageIDProcessing)
}

// ExecuteIndicesStage handles POST /api/stages/indices/execute
func (h *OperationsHandler) ExecuteIndicesStage(w http.ResponseWriter, r *http.Request) {
	h.executeSpecificStage(w, r, operations.StageIDIndices)
}

// ExecuteLiquidityStage handles POST /api/stages/liquidity/execute
func (h *OperationsHandler) ExecuteLiquidityStage(w http.ResponseWriter, r *http.Request) {
	h.executeSpecificStage(w, r, operations.StageIDLiquidity)
}
