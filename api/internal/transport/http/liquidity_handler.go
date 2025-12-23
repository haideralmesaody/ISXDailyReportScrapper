package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	apierrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/services"
)

// LiquidityHandler handles liquidity-related HTTP requests
type LiquidityHandler struct {
	service      *services.LiquidityService
	logger       *slog.Logger
	errorHandler *apierrors.ErrorHandler
}

// NewLiquidityHandler creates a new liquidity handler
func NewLiquidityHandler(service *services.LiquidityService, logger *slog.Logger) *LiquidityHandler {
	return &LiquidityHandler{
		service:      service,
		logger:       logger,
		errorHandler: apierrors.NewErrorHandler(logger, false),
	}
}

// RegisterRoutes registers the liquidity routes
func (h *LiquidityHandler) RegisterRoutes(r chi.Router) {
	r.Route("/liquidity", func(r chi.Router) {
		r.Get("/insights", h.GetInsights)
	})
}

// GetInsights returns the latest liquidity insights
func (h *LiquidityHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse optional mode parameter
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "ema" // Default to EMA mode
	}
	
	// Validate mode
	validModes := map[string]bool{
		"ema":     true,
		"latest":  true,
		"average": true,
	}
	
	if !validModes[mode] {
		h.logger.WarnContext(ctx, "Invalid scoring mode requested",
			slog.String("mode", mode))
		
		h.errorHandler.HandleError(w, r, apierrors.New(
			http.StatusBadRequest,
			"INVALID_MODE",
			"Invalid scoring mode. Use: ema, latest, or average",
		))
		return
	}
	
	h.logger.InfoContext(ctx, "Getting liquidity insights",
		slog.String("mode", mode))

	insights, err := h.service.GetLatestInsights(ctx, mode)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get liquidity insights",
			slog.String("error", err.Error()))
		
		h.errorHandler.HandleError(w, r, apierrors.New(
			http.StatusInternalServerError,
			"LIQUIDITY_ERROR",
			"Failed to retrieve liquidity insights",
		))
		return
	}
	
	// Add mode to response header for client reference
	w.Header().Set("X-Liquidity-Mode", mode)
	
	// Set active mode in all stocks
	if insights != nil && insights.AllStocks != nil {
		for i := range insights.AllStocks {
			insights.AllStocks[i].ActiveMode = mode
		}
	}
	
	// Success response
	render.JSON(w, r, insights)
}