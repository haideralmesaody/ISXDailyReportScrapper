package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// MetricsHandler handles system metrics and health endpoints
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// Routes sets up the metrics routes
func (h *MetricsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/health", h.GetHealth)
	r.Get("/metrics", h.GetMetrics)
	return r
}

// GetHealth returns basic health status
func (h *MetricsHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	render.JSON(w, r, response)
}

// GetMetrics returns basic metrics
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "ok",
		"metrics": map[string]interface{}{
			"requests_total":     0,
			"active_connections": 0,
		},
	}
	render.JSON(w, r, response)
}
