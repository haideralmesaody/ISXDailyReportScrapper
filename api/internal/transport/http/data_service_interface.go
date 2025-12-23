package http

import (
	"context"
	"net/http"
)

// DataServiceInterface defines the interface for data operations
type DataServiceInterface interface {
	GetReports(ctx context.Context) ([]map[string]interface{}, error)
	GetTickers(ctx context.Context) (interface{}, error)
	GetIndices(ctx context.Context) (map[string]interface{}, error)
	GetIndicesHistory(ctx context.Context, startDate, endDate string) (map[string]interface{}, error)
	GetFiles(ctx context.Context) (map[string]interface{}, error)
	GetMarketMovers(ctx context.Context, period, limit, minVolume string) (map[string]interface{}, error)
	GetTradingDates(ctx context.Context) ([]string, error)
	GetDailyMarketData(ctx context.Context, date string) ([]map[string]interface{}, error)
	GetDailyMarketDataWithHistory(ctx context.Context, date string, historyDays int) ([]map[string]interface{}, error)
	GetTickerChart(ctx context.Context, ticker string) (map[string]interface{}, error)
	DownloadFile(ctx context.Context, w http.ResponseWriter, r *http.Request, fileType, filename string) error

	// Safe trading methods
	GetSafeTradingLimits(ctx context.Context, ticker string) (interface{}, error)
	EstimateTradeImpact(ctx context.Context, ticker string, tradeValue float64) (float64, error)
	CreateTradeSchedule(ctx context.Context, ticker string, totalTradeValue float64) (interface{}, error)
}