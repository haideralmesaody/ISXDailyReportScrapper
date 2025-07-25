package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	licenseErrors "isxcli/internal/errors"
	"isxcli/internal/services"
)

// LicenseHandler handles license-related HTTP requests with clean architecture
type LicenseHandler struct {
	service services.LicenseService
	logger  *slog.Logger
}

// NewLicenseHandler creates a new license handler
func NewLicenseHandler(service services.LicenseService, logger *slog.Logger) *LicenseHandler {
	return &LicenseHandler{
		service: service,
		logger:  logger.With(slog.String("handler", "license")),
	}
}

// LicenseActivationRequest represents the license activation request payload
type LicenseActivationRequest struct {
	LicenseKey string `json:"license_key" validate:"required"`
}

// Bind implements the render.Binder interface for request validation
func (l *LicenseActivationRequest) Bind(r *http.Request) error {
	if l.LicenseKey == "" {
		return errors.New("license_key is required")
	}
	return nil
}

// LicenseActivationResponse represents the license activation response
type LicenseActivationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

// Routes returns a chi router for license endpoints
func (h *LicenseHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	// Apply timeout middleware to all license routes
	r.Use(middleware.Timeout(30 * time.Second))
	
	r.Get("/status", h.GetStatus)
	r.Post("/activate", h.Activate)
	
	return r
}

// GetStatus handles GET /api/license/status
func (h *LicenseHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get status with timeout
	statusCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	response, err := h.service.GetStatus(statusCtx)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	// Return the standardized response
	render.JSON(w, r, response)
}

// Activate handles POST /api/license/activate
func (h *LicenseHandler) Activate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	// Decode and validate request
	data := &LicenseActivationRequest{}
	if err := render.Bind(r, data); err != nil {
		h.logger.ErrorContext(ctx, "failed to bind license activation request",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID))
		
		problem := licenseErrors.NewProblemDetails(
			http.StatusBadRequest,
			"/errors/invalid-request",
			"Invalid Request",
			err.Error(),
			"/api/license/activate#"+reqID,
		).WithExtension("trace_id", reqID)
		
		render.Render(w, r, problem)
		return
	}
	
	// Activate license with timeout
	activateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	err := h.service.Activate(activateCtx, data.LicenseKey)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	
	// Success response
	render.JSON(w, r, LicenseActivationResponse{
		Success: true,
		Message: "License activated successfully. You can now use all features.",
		TraceID: reqID,
	})
}

// handleError centralizes error handling for the handler
func (h *LicenseHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()
	reqID := middleware.GetReqID(ctx)
	
	h.logger.ErrorContext(ctx, "request failed",
		slog.String("error", err.Error()),
		slog.String("request_id", reqID),
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method))
	
	// Handle timeout specifically
	if errors.Is(err, context.DeadlineExceeded) {
		problem := licenseErrors.NewProblemDetails(
			http.StatusGatewayTimeout,
			"/errors/timeout",
			"Request Timeout",
			"The request timed out while processing",
			r.URL.Path+"#"+reqID,
		).WithExtension("trace_id", reqID)
		
		render.Render(w, r, problem)
		return
	}
	
	// Use the centralized error mapper
	problem := licenseErrors.MapLicenseError(err, reqID)
	render.Render(w, r, problem)
}