package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"log/slog"

	"github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/services"
	"github.com/isxcli/isxcli/internal/strategy"
)

// StrategyHandler handles strategy-related HTTP requests
type StrategyHandler struct {
	strategyService *services.StrategyService
	logger          *slog.Logger
	errorHandler    *errors.ErrorHandler
}

// NewStrategyHandler creates a new strategy handler
func NewStrategyHandler(strategyService *services.StrategyService, logger *slog.Logger) *StrategyHandler {
	return &StrategyHandler{
		strategyService: strategyService,
		logger:          logger,
		errorHandler:    errors.NewErrorHandler(logger, false),
	}
}

// RegisterRoutes registers strategy routes with the router
func (h *StrategyHandler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/strategies", func(r chi.Router) {
		r.Get("/", h.ListStrategies)
		r.Get("/{strategyID}", h.GetStrategy)
		r.Get("/{strategyID}/chart-preset", h.GetStrategyChartPreset)
		r.Post("/{strategyID}/execute", h.ExecuteStrategy)
		r.Post("/{strategyID}/execute-batch", h.ExecuteStrategyBatch)
		r.Get("/{strategyID}/runs", h.ListStrategyRuns)
		r.Get("/{strategyID}/runs/{runID}", h.GetStrategyRun)
		r.Get("/{strategyID}/runs/{runID}/backtest/{symbol}", h.GetBacktestTickerDetails)
		r.Post("/{strategyID}/backtest", h.RunBacktest)
		r.Post("/{strategyID}/validate", h.ValidateParameters)
		r.Get("/{strategyID}/signals", h.GetSignals)
		r.Post("/execute-multiple", h.ExecuteMultipleStrategies)
	})
}

// GetStrategyChartPreset handles GET /api/v1/strategies/{strategyID}/chart-preset
func (h *StrategyHandler) GetStrategyChartPreset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	h.logger.InfoContext(ctx, "getting strategy chart preset", "strategy_id", strategyID)

	preset, err := h.strategyService.GetStrategyChartPreset(ctx, strategyID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, preset)
}

// ListStrategies handles GET /api/v1/strategies
func (h *StrategyHandler) ListStrategies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.InfoContext(ctx, "listing strategies")

	strategies, err := h.strategyService.ListStrategies(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"strategies": strategies,
		"count":      len(strategies),
	})
}

// GetStrategy handles GET /api/v1/strategies/{strategyID}
func (h *StrategyHandler) GetStrategy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	h.logger.InfoContext(ctx, "getting strategy", "strategy_id", strategyID)

	strategy, err := h.strategyService.GetStrategy(ctx, strategyID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, strategy)
}

// ExecuteStrategy handles POST /api/v1/strategies/{strategyID}/execute
func (h *StrategyHandler) ExecuteStrategy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	var req services.ExecuteStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, errors.NewValidationError("invalid request body"))
		return
	}

	// Set strategy ID from URL parameter
	req.StrategyID = strategyID

	// Validate request
	if err := h.validateExecuteRequest(req); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.InfoContext(ctx, "executing strategy",
		"strategy_id", strategyID,
		"symbol", req.Symbol,
	)

	signal, err := h.strategyService.ExecuteStrategy(ctx, req)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"signal":      signal,
		"strategy_id": strategyID,
		"symbol":      req.Symbol,
		"timestamp":   time.Now(),
	})
}

// ExecuteStrategyBatch handles POST /api/v1/strategies/{strategyID}/execute-batch
func (h *StrategyHandler) ExecuteStrategyBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	var req services.ExecuteBatchRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			h.errorHandler.HandleError(w, r, errors.NewValidationError("invalid request body"))
			return
		}
	}

	// Optional validation (service does the final validation).
	if req.DataPoints != 0 && (req.DataPoints < 16 || req.DataPoints > 2000) {
		h.errorHandler.HandleError(w, r, errors.NewValidationError("data_points must be between 16 and 2000"))
		return
	}

	h.logger.InfoContext(ctx, "executing strategy batch",
		"strategy_id", strategyID,
		"symbols", len(req.Symbols),
		"data_points", req.DataPoints,
	)

	result, err := h.strategyService.ExecuteStrategyBatch(ctx, strategyID, req)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, result)
}

// ListStrategyRuns handles GET /api/v1/strategies/{strategyID}/runs
func (h *StrategyHandler) ListStrategyRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	limit := 25
	if strings.TrimSpace(r.URL.Query().Get("limit")) != "" {
		parsed, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || parsed < 1 || parsed > 200 {
			h.errorHandler.HandleError(w, r, errors.NewValidationError("limit must be between 1 and 200"))
			return
		}
		limit = parsed
	}

	h.logger.InfoContext(ctx, "listing strategy runs",
		"strategy_id", strategyID,
		"limit", limit,
	)

	runs, err := h.strategyService.ListStrategyRuns(ctx, strategyID, limit)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"runs":  runs,
		"count": len(runs),
	})
}

// GetStrategyRun handles GET /api/v1/strategies/{strategyID}/runs/{runID}
func (h *StrategyHandler) GetStrategyRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")
	runID := chi.URLParam(r, "runID")

	h.logger.InfoContext(ctx, "getting strategy run",
		"strategy_id", strategyID,
		"run_id", runID,
	)

	run, err := h.strategyService.GetStrategyRun(ctx, strategyID, runID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, run)
}

// GetBacktestTickerDetails handles GET /api/v1/strategies/{strategyID}/runs/{runID}/backtest/{symbol}
func (h *StrategyHandler) GetBacktestTickerDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")
	runID := chi.URLParam(r, "runID")
	symbol := chi.URLParam(r, "symbol")

	h.logger.InfoContext(ctx, "loading backtest details",
		"strategy_id", strategyID,
		"run_id", runID,
		"symbol", symbol,
	)

	details, err := h.strategyService.GetBacktestTickerDetails(ctx, strategyID, runID, symbol)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, details)
}

// ExecuteMultipleStrategies handles POST /api/v1/strategies/execute-multiple
func (h *StrategyHandler) ExecuteMultipleStrategies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req services.ExecuteMultipleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, errors.NewValidationError("invalid request body"))
		return
	}

	if err := h.validateMultipleExecuteRequest(req); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.InfoContext(ctx, "executing multiple strategies",
		"strategies", len(req.StrategyIDs),
		"symbol", req.Symbol,
	)

	results, err := h.strategyService.ExecuteMultipleStrategies(ctx, req)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Count successful executions
	successful := 0
	for _, result := range results {
		if result.Success {
			successful++
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"results":    results,
		"total":      len(results),
		"successful": successful,
		"symbol":     req.Symbol,
		"timestamp":  time.Now(),
	})
}

// RunBacktest handles POST /api/v1/strategies/{strategyID}/backtest
func (h *StrategyHandler) RunBacktest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	var req services.BacktestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, errors.NewValidationError("invalid request body"))
		return
	}

	req.StrategyID = strategyID

	if err := h.validateBacktestRequest(req); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.logger.InfoContext(ctx, "running backtest",
		"strategy_id", strategyID,
		"symbol", req.Symbol,
		"start_date", req.StartDate,
		"end_date", req.EndDate,
	)

	result, err := h.strategyService.RunBacktest(ctx, req)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"backtest_result": result,
		"strategy_id":     strategyID,
		"timestamp":       time.Now(),
	})
}

// ValidateParameters handles POST /api/v1/strategies/{strategyID}/validate
func (h *StrategyHandler) ValidateParameters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	var params strategy.StrategyParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		h.errorHandler.HandleError(w, r, errors.NewValidationError("invalid request body"))
		return
	}

	err := h.strategyService.ValidateStrategyParameters(ctx, strategyID, params)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":       true,
		"strategy_id": strategyID,
		"message":     "Parameters are valid",
	})
}

// GetSignals handles GET /api/v1/strategies/{strategyID}/signals
func (h *StrategyHandler) GetSignals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	strategyID := chi.URLParam(r, "strategyID")

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	signals, err := h.strategyService.GetStrategySignals(ctx, strategyID, limit)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"signals":     signals,
		"count":       len(signals),
		"strategy_id": strategyID,
		"limit":       limit,
	})
}

// respondJSON sends a JSON response
func (h *StrategyHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// validateExecuteRequest validates execute strategy request
func (h *StrategyHandler) validateExecuteRequest(req services.ExecuteStrategyRequest) error {
	if req.StrategyID == "" {
		return errors.NewValidationError("strategy_id is required")
	}
	if req.Symbol == "" {
		return errors.NewValidationError("symbol is required")
	}
	if req.DataPoints < 20 || req.DataPoints > 1000 {
		return errors.NewValidationError("data_points must be between 20 and 1000")
	}
	return nil
}

// validateMultipleExecuteRequest validates multiple execute request
func (h *StrategyHandler) validateMultipleExecuteRequest(req services.ExecuteMultipleRequest) error {
	if len(req.StrategyIDs) == 0 {
		return errors.NewValidationError("strategy_ids is required")
	}
	if len(req.StrategyIDs) > 10 {
		return errors.NewValidationError("maximum 10 strategies allowed")
	}
	if req.Symbol == "" {
		return errors.NewValidationError("symbol is required")
	}
	if req.DataPoints < 20 || req.DataPoints > 1000 {
		return errors.NewValidationError("data_points must be between 20 and 1000")
	}
	return nil
}

// validateBacktestRequest validates backtest request
func (h *StrategyHandler) validateBacktestRequest(req services.BacktestRequest) error {
	if req.StrategyID == "" {
		return errors.NewValidationError("strategy_id is required")
	}
	if req.Symbol == "" {
		return errors.NewValidationError("symbol is required")
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		return errors.NewValidationError("start_date and end_date are required")
	}
	if req.EndDate.Before(req.StartDate) {
		return errors.NewValidationError("end_date must be after start_date")
	}
	if req.InitialCash < 1000 {
		return errors.NewValidationError("initial_cash must be at least 1000")
	}
	if req.Commission < 0 || req.Commission > 0.1 {
		return errors.NewValidationError("commission must be between 0 and 0.1")
	}
	if req.Slippage < 0 || req.Slippage > 0.1 {
		return errors.NewValidationError("slippage must be between 0 and 0.1")
	}
	return nil
}
