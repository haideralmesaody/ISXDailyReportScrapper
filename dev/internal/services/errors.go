package services

import "errors"

// Data service errors
var (
	// Report errors
	ErrNoReportsFound = errors.New("no reports found")
	
	// Ticker errors
	ErrNoTickersFound = errors.New("no tickers found")
	ErrTickerNotFound = errors.New("ticker not found")
	ErrNoChartData    = errors.New("no chart data available")
	
	// Index errors
	ErrNoIndicesFound = errors.New("no indices found")
	
	// File errors
	ErrNoFilesFound    = errors.New("no files found")
	ErrFileNotFound    = errors.New("file not found")
	ErrInvalidFileType = errors.New("invalid file type")
	
	// Market movers errors
	ErrNoMarketMovers = errors.New("no market movers found")
	
	// Pipeline errors
	ErrPipelineNotFound    = errors.New("pipeline not found")
	ErrPipelineRunning     = errors.New("pipeline already running")
	ErrPipelineNotRunning  = errors.New("pipeline not running")
	ErrInvalidStage        = errors.New("invalid pipeline stage")
	
	// WebSocket errors
	ErrWebSocketUpgrade    = errors.New("websocket upgrade failed")
	ErrWebSocketClosed     = errors.New("websocket connection closed")
	
	// General errors
	ErrInvalidInput      = errors.New("invalid input")
	ErrOperationTimeout  = errors.New("operation timed out")
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
)