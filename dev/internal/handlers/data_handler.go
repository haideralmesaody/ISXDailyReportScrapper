package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	
	apierrors "isxcli/internal/errors"
	"isxcli/internal/services"
)

// DataHandler handles data-related HTTP requests with RFC 7807 compliance
type DataHandler struct {
	service      DataServiceInterface
	logger       *slog.Logger
	errorHandler *apierrors.ErrorHandler
}

// NewDataHandler creates a new data handler with RFC 7807 error handling
func NewDataHandler(service DataServiceInterface, logger *slog.Logger, errorHandler *apierrors.ErrorHandler) *DataHandler {
	return &DataHandler{
		service:      service,
		logger:       logger.With(slog.String("component", "data_handler")),
		errorHandler: errorHandler,
	}
}

// Routes returns the data routes with proper Chi patterns
func (h *DataHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	// Use render for consistent JSON responses
	r.Use(render.SetContentType(render.ContentTypeJSON))
	
	// Resource routes following REST patterns
	r.Get("/reports", h.GetReports)
	r.Get("/tickers", h.GetTickers)
	r.Get("/indices", h.GetIndices)
	r.Get("/files", h.GetFiles)
	r.Get("/market-movers", h.GetMarketMovers)
	
	// Sub-resource routes
	r.Route("/ticker/{ticker}", func(r chi.Router) {
		r.Use(h.TickerCtx) // Load ticker into context
		r.Get("/chart", h.GetTickerChart)
	})
	
	// Download routes
	r.Route("/download/{type}/{filename}", func(r chi.Router) {
		r.Use(h.DownloadCtx) // Validate download parameters
		r.Get("/", h.DownloadFile)
	})
	
	return r
}

// TickerCtx middleware validates ticker parameter
func (h *DataHandler) TickerCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ticker := chi.URLParam(r, "ticker")
		if ticker == "" {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("ticker", "Ticker symbol is required"))
			return
		}
		
		// Validate ticker format (basic validation)
		if len(ticker) > 10 || len(ticker) < 2 {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("ticker", "Invalid ticker symbol format"))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// DownloadCtx middleware validates download parameters
func (h *DataHandler) DownloadCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileType := chi.URLParam(r, "type")
		filename := chi.URLParam(r, "filename")
		
		// Validate file type
		validTypes := map[string]bool{
			"report": true,
			"excel":  true,
			"csv":    true,
			"json":   true,
		}
		
		if !validTypes[fileType] {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("type", fmt.Sprintf("Invalid file type: %s", fileType)))
			return
		}
		
		if filename == "" {
			h.errorHandler.HandleError(w, r, apierrors.ErrValidation("filename", "Filename is required"))
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// GetReports handles GET /api/data/reports with RFC 7807 errors
func (h *DataHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	h.logger.InfoContext(r.Context(), "fetching reports",
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	
	reports, err := h.service.GetReports()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get reports",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		// Map service errors to API errors
		if errors.Is(err, services.ErrNoReportsFound) {
			h.errorHandler.HandleError(w, r, apierrors.New(
				http.StatusNotFound,
				"NO_REPORTS_FOUND",
				"No reports available",
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Success response
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   reports,
		"count":  len(reports),
	})
}

// GetTickers handles GET /api/data/tickers with RFC 7807 errors
func (h *DataHandler) GetTickers(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	h.logger.InfoContext(r.Context(), "fetching tickers",
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	
	tickers, err := h.service.GetTickers()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get tickers",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		if errors.Is(err, services.ErrNoTickersFound) {
			h.errorHandler.HandleError(w, r, apierrors.New(
				http.StatusNotFound,
				"NO_TICKERS_FOUND",
				"No tickers available",
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Count depends on return type from service
	count := 0
	if arr, ok := tickers.([]interface{}); ok {
		count = len(arr)
	} else if _, ok := tickers.(map[string]interface{}); ok {
		// If it's a map, we can't easily count items
		count = 1
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   tickers,
		"count":  count,
	})
}

// GetIndices handles GET /api/data/indices with RFC 7807 errors
func (h *DataHandler) GetIndices(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	h.logger.InfoContext(r.Context(), "fetching indices",
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	
	indices, err := h.service.GetIndices()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get indices",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		if errors.Is(err, services.ErrNoIndicesFound) {
			h.errorHandler.HandleError(w, r, apierrors.New(
				http.StatusNotFound,
				"NO_INDICES_FOUND",
				"No indices available",
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Indices is a map with isx60 and isx15 arrays
	count := 0
	if isx60, ok := indices["isx60"].([]interface{}); ok {
		count += len(isx60)
	}
	if isx15, ok := indices["isx15"].([]interface{}); ok {
		count += len(isx15)
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   indices,
		"count":  count,
	})
}

// GetFiles handles GET /api/data/files with RFC 7807 errors
func (h *DataHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	h.logger.InfoContext(r.Context(), "fetching files",
		slog.String("request_id", reqID),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	
	files, err := h.service.GetFiles()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get files",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		if errors.Is(err, services.ErrNoFilesFound) {
			h.errorHandler.HandleError(w, r, apierrors.New(
				http.StatusNotFound,
				"NO_FILES_FOUND",
				"No files available",
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Files is a map with multiple arrays
	count := 0
	if downloads, ok := files["downloads"].([]interface{}); ok {
		count += len(downloads)
	}
	if reports, ok := files["reports"].([]interface{}); ok {
		count += len(reports)
	}
	if csvFiles, ok := files["csvFiles"].([]interface{}); ok {
		count += len(csvFiles)
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   files,
		"count":  count,
	})
}

// GetMarketMovers handles GET /api/data/market-movers with RFC 7807 errors
func (h *DataHandler) GetMarketMovers(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	
	// Parse and validate query parameters
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}
	
	// Validate period
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	
	if !validPeriods[period] {
		h.errorHandler.HandleError(w, r, apierrors.ErrValidation("period", "Invalid period. Must be one of: daily, weekly, monthly"))
		return
	}
	
	// Parse limit
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		h.errorHandler.HandleError(w, r, apierrors.ErrValidation("limit", "Limit must be a number between 1 and 100"))
		return
	}
	
	// Parse minVolume
	minVolumeStr := r.URL.Query().Get("minVolume")
	if minVolumeStr == "" {
		minVolumeStr = "0"
	}
	
	minVolume, err := strconv.ParseFloat(minVolumeStr, 64)
	if err != nil || minVolume < 0 {
		h.errorHandler.HandleError(w, r, apierrors.ErrValidation("minVolume", "Min volume must be a positive number"))
		return
	}
	
	h.logger.InfoContext(r.Context(), "fetching market movers",
		slog.String("request_id", reqID),
		slog.String("period", period),
		slog.Int("limit", limit),
		slog.Float64("min_volume", minVolume),
	)
	
	movers, err := h.service.GetMarketMovers(period, limitStr, minVolumeStr)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get market movers",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
		)
		
		if errors.Is(err, services.ErrNoMarketMovers) {
			h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
				http.StatusNotFound,
				"NO_MARKET_MOVERS",
				"No market movers found for the specified criteria",
				map[string]interface{}{
					"period":     period,
					"limit":      limit,
					"min_volume": minVolume,
				},
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	// Movers is a map with gainers, losers, mostActive arrays
	count := 0
	if gainers, ok := movers["gainers"].([]interface{}); ok {
		count += len(gainers)
	}
	if losers, ok := movers["losers"].([]interface{}); ok {
		count += len(losers)
	}
	if mostActive, ok := movers["mostActive"].([]interface{}); ok {
		count += len(mostActive)
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   movers,
		"count":  count,
		"params": map[string]interface{}{
			"period":     period,
			"limit":      limit,
			"min_volume": minVolume,
		},
	})
}

// GetTickerChart handles GET /api/data/ticker/{ticker}/chart with RFC 7807 errors
func (h *DataHandler) GetTickerChart(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	ticker := chi.URLParam(r, "ticker")
	
	h.logger.InfoContext(r.Context(), "fetching ticker chart",
		slog.String("request_id", reqID),
		slog.String("ticker", ticker),
	)
	
	chart, err := h.service.GetTickerChart(ticker)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to get ticker chart",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("ticker", ticker),
		)
		
		if errors.Is(err, services.ErrTickerNotFound) {
			h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
				http.StatusNotFound,
				"TICKER_NOT_FOUND",
				fmt.Sprintf("Ticker '%s' not found", ticker),
				map[string]interface{}{
					"ticker": ticker,
				},
			))
			return
		}
		
		if errors.Is(err, services.ErrNoChartData) {
			h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
				http.StatusNotFound,
				"NO_CHART_DATA",
				fmt.Sprintf("No chart data available for ticker '%s'", ticker),
				map[string]interface{}{
					"ticker": ticker,
				},
			))
			return
		}
		
		h.errorHandler.HandleError(w, r, err)
		return
	}
	
	render.JSON(w, r, map[string]interface{}{
		"status": "success",
		"data":   chart,
		"ticker": ticker,
	})
}

// DownloadFile handles GET /api/data/download/{type}/{filename} with RFC 7807 errors
func (h *DataHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	fileType := chi.URLParam(r, "type")
	filename := chi.URLParam(r, "filename")
	
	h.logger.InfoContext(r.Context(), "downloading file",
		slog.String("request_id", reqID),
		slog.String("file_type", fileType),
		slog.String("filename", filename),
	)
	
	// Let service handle the download (it writes directly to response)
	if err := h.service.DownloadFile(w, r, fileType, filename); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to download file",
			slog.String("error", err.Error()),
			slog.String("request_id", reqID),
			slog.String("file_type", fileType),
			slog.String("filename", filename),
		)
		
		// Only handle error if response not yet written
		if !isResponseWritten(w) {
			if errors.Is(err, services.ErrFileNotFound) {
				h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
					http.StatusNotFound,
					"FILE_NOT_FOUND",
					fmt.Sprintf("File '%s' not found", filename),
					map[string]interface{}{
						"type":     fileType,
						"filename": filename,
					},
				))
				return
			}
			
			if errors.Is(err, services.ErrInvalidFileType) {
				h.errorHandler.HandleError(w, r, apierrors.NewWithDetails(
					http.StatusBadRequest,
					"INVALID_FILE_TYPE",
					fmt.Sprintf("Invalid file type: %s", fileType),
					map[string]interface{}{
						"type":     fileType,
						"filename": filename,
					},
				))
				return
			}
			
			h.errorHandler.HandleError(w, r, err)
		}
	}
}

// isResponseWritten checks if response has already been written
func isResponseWritten(w http.ResponseWriter) bool {
	// Check if writer is a wrapped response writer with status
	if ww, ok := w.(interface{ Status() int }); ok {
		return ww.Status() != 0
	}
	return false
}