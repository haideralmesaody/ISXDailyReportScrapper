package handlers

import (
	"net/http"
	"github.com/go-chi/render"
	"isxcli/internal/services"
)

// HealthHandler handles health-related HTTP requests
type HealthHandler struct {
	service *services.HealthService
	logger  services.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(service *services.HealthService, logger services.Logger) *HealthHandler {
	return &HealthHandler{
		service: service,
		logger:  logger,
	}
}

// HealthCheck handles GET /api/health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.HealthCheck())
}

// ReadinessCheck handles GET /api/health/ready
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.ReadinessCheck())
}

// LivenessCheck handles GET /api/health/live
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.LivenessCheck())
}

// Version handles GET /api/version
func (h *HealthHandler) Version(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.service.Version())
}

// LicenseStatus handles GET /api/license/status
func (h *HealthHandler) LicenseStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.LicenseStatus()
	if err != nil {
		h.logger.Error("Failed to get license status", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, status)
}