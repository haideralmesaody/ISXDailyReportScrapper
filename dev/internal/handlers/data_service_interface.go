package handlers

import (
	"net/http"
)

// DataServiceInterface defines the interface for data operations
type DataServiceInterface interface {
	GetReports() ([]map[string]interface{}, error)
	GetTickers() (interface{}, error)
	GetIndices() (map[string]interface{}, error)
	GetFiles() (map[string]interface{}, error)
	GetMarketMovers(period, limit, minVolume string) (map[string]interface{}, error)
	GetTickerChart(ticker string) (map[string]interface{}, error)
	DownloadFile(w http.ResponseWriter, r *http.Request, fileType, filename string) error
}