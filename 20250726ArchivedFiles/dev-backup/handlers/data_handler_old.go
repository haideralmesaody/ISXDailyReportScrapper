package handlers

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"isxcli/internal/services"
)

// DataHandler handles data-related HTTP requests
type DataHandler struct {
	service *services.DataService
	logger  services.Logger
}

// NewDataHandler creates a new data handler
func NewDataHandler(service *services.DataService, logger services.Logger) *DataHandler {
	return &DataHandler{
		service: service,
		logger:  logger,
	}
}

// Routes returns the data routes
func (h *DataHandler) Routes() chi.Router {
	r := chi.NewRouter()
	
	r.Get("/reports", h.GetReports)
	r.Get("/tickers", h.GetTickers)
	r.Get("/indices", h.GetIndices)
	r.Get("/files", h.GetFiles)
	r.Get("/market-movers", h.GetMarketMovers)
	r.Get("/ticker/{ticker}/chart", h.GetTickerChart)
	r.Get("/download/{type}/{filename}", h.DownloadFile)
	
	return r
}

// GetReports handles GET /api/data/reports
func (h *DataHandler) GetReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.service.GetReports()
	if err != nil {
		h.logger.Error("Failed to get reports", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, reports)
}

// GetTickers handles GET /api/data/tickers
func (h *DataHandler) GetTickers(w http.ResponseWriter, r *http.Request) {
	tickers, err := h.service.GetTickers()
	if err != nil {
		h.logger.Error("Failed to get tickers", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, tickers)
}

// GetIndices handles GET /api/data/indices
func (h *DataHandler) GetIndices(w http.ResponseWriter, r *http.Request) {
	indices, err := h.service.GetIndices()
	if err != nil {
		h.logger.Error("Failed to get indices", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, indices)
}

// GetFiles handles GET /api/data/files
func (h *DataHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	files, err := h.service.GetFiles()
	if err != nil {
		h.logger.Error("Failed to get files", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, files)
}

// GetMarketMovers handles GET /api/data/market-movers
func (h *DataHandler) GetMarketMovers(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}
	minVolume := r.URL.Query().Get("minVolume")
	if minVolume == "" {
		minVolume = "0"
	}

	movers, err := h.service.GetMarketMovers(period, limit, minVolume)
	if err != nil {
		h.logger.Error("Failed to get market movers", "error", err)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, movers)
}

// GetTickerChart handles GET /api/data/ticker/{ticker}/chart
func (h *DataHandler) GetTickerChart(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	chart, err := h.service.GetTickerChart(ticker)
	if err != nil {
		h.logger.Error("Failed to get ticker chart", "error", err, "ticker", ticker)
		render.JSON(w, r, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	render.JSON(w, r, chart)
}

// DownloadFile handles GET /api/data/download/{type}/{filename}
func (h *DataHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileType := chi.URLParam(r, "type")
	filename := chi.URLParam(r, "filename")
	
	if err := h.service.DownloadFile(w, r, fileType, filename); err != nil {
		h.logger.Error("Failed to download file", "error", err, "type", fileType, "filename", filename)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}