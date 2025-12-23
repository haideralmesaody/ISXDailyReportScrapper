// Package api contains API contract definitions for the ISX Daily Reports Scrapper.
// Version v1 represents the current stable API version.
package api

import (
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// Common request parameters

// PaginationRequest represents common pagination parameters
type PaginationRequest struct {
	Page     int    `json:"page" query:"page" validate:"min=1"`
	PageSize int    `json:"page_size" query:"page_size" validate:"min=1,max=100"`
	Sort     string `json:"sort" query:"sort" validate:"omitempty,oneof=asc desc"`
	SortBy   string `json:"sort_by" query:"sort_by"`
}

// DateRangeRequest represents a date range in requests
type DateRangeRequest struct {
	From string `json:"from" query:"from" validate:"omitempty,datetime=2006-01-02"`
	To   string `json:"to" query:"to" validate:"omitempty,datetime=2006-01-02"`
}

// License API Requests

// LicenseActivateRequest represents a license activation request
type LicenseActivateRequest struct {
	LicenseKey string `json:"license_key" validate:"required,min=10"`
	Email      string `json:"email" validate:"required,email"`
}

// LicenseValidateRequest represents a license validation request
type LicenseValidateRequest struct {
	LicenseKey string `json:"license_key" validate:"required,min=10"`
	MachineID  string `json:"machine_id" validate:"required"`
}

// LicenseRenewRequest represents a license renewal request
type LicenseRenewRequest struct {
	LicenseKey string `json:"license_key" validate:"required,min=10"`
	Duration   string `json:"duration" validate:"required,oneof=monthly quarterly yearly"`
}

// operation API Requests

// PipelineStartRequest represents a request to start a operation
type PipelineStartRequest struct {
	Mode       string                 `json:"mode" validate:"required,oneof=initial accumulative full"`
	FromDate   string                 `json:"from_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	ToDate     string                 `json:"to_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// PipelineStopRequest represents a request to stop a operation
type PipelineStopRequest struct {
	PipelineID string `json:"pipeline_id" param:"id" validate:"required,uuid"`
	Force      bool   `json:"force" query:"force"`
}

// PipelineListRequest represents a request to list operations
type PipelineListRequest struct {
	PaginationRequest
	Status   string `json:"status" query:"status" validate:"omitempty,oneof=pending running completed failed cancelled"`
	Type     string `json:"type" query:"type" validate:"omitempty,oneof=scraping processing indexing liquidity"`
	DateFrom string `json:"date_from" query:"date_from" validate:"omitempty,datetime=2006-01-02"`
	DateTo   string `json:"date_to" query:"date_to" validate:"omitempty,datetime=2006-01-02"`
}

// Data API Requests

// DataScrapingRequest represents a request to start data scraping
type DataScrapingRequest struct {
	StartDate    string   `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate      string   `json:"end_date" validate:"required,datetime=2006-01-02,gtefield=StartDate"`
	Mode         string   `json:"mode" validate:"required,oneof=initial accumulative full"`
	FileTypes    []string `json:"file_types,omitempty" validate:"omitempty,dive,oneof=csv excel pdf"`
	OverwriteExisting bool `json:"overwrite_existing"`
}

// DataProcessingRequest represents a request to process data
type DataProcessingRequest struct {
	InputPath    string                    `json:"input_path" validate:"required"`
	OutputPath   string                    `json:"output_path" validate:"required"`
	ProcessorType string                   `json:"processor_type" validate:"required,oneof=csv excel json xml"`
	Options      domain.ProcessorConfig    `json:"options,omitempty"`
}

// DataExportRequest represents a request to export data
type DataExportRequest struct {
	DateRange    DateRangeRequest `json:"date_range" validate:"required"`
	Format       string           `json:"format" validate:"required,oneof=csv excel json pdf"`
	IncludeTickers []string       `json:"include_tickers,omitempty"`
	ExcludeTickers []string       `json:"exclude_tickers,omitempty"`
	Fields       []string         `json:"fields,omitempty"`
	Compression  string           `json:"compression,omitempty" validate:"omitempty,oneof=none gzip zip"`
}

// Report API Requests

// ReportListRequest represents a request to list reports
type ReportListRequest struct {
	PaginationRequest
	DateRange DateRangeRequest      `json:"date_range,omitempty"`
	Types     []string              `json:"types,omitempty" validate:"omitempty,dive,oneof=daily weekly monthly summary"`
	Statuses  []string              `json:"statuses,omitempty" validate:"omitempty,dive,oneof=pending processing completed failed"`
	Filter    domain.ReportFilter   `json:"filter,omitempty"`
}

// ReportDownloadRequest represents a request to download a report
type ReportDownloadRequest struct {
	ReportID string `json:"report_id" param:"id" validate:"required,uuid"`
	Format   string `json:"format" query:"format" validate:"omitempty,oneof=original pdf csv excel"`
}

// ReportGenerateRequest represents a request to generate a report
type ReportGenerateRequest struct {
	Type         string                 `json:"type" validate:"required,oneof=daily weekly monthly summary custom"`
	DateRange    DateRangeRequest       `json:"date_range" validate:"required"`
	IncludeCharts bool                  `json:"include_charts"`
	IncludeLiquidity bool                `json:"include_liquidity"`
	Template     string                 `json:"template,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// Ticker API Requests

// TickerListRequest represents a request to list tickers
type TickerListRequest struct {
	PaginationRequest
	Filter     domain.TickerFilter `json:"filter,omitempty"`
	IncludeInactive bool           `json:"include_inactive" query:"include_inactive"`
}

// TickerDataRequest represents a request for ticker data
type TickerDataRequest struct {
	Symbol    string           `json:"symbol" param:"symbol" validate:"required"`
	DateRange DateRangeRequest `json:"date_range,omitempty"`
	Interval  string           `json:"interval" query:"interval" validate:"omitempty,oneof=1m 5m 15m 30m 1h 1d 1w 1M"`
}

// TickerLiquidityRequest represents a request for ticker liquidity analysis
type TickerLiquidityRequest struct {
	Symbol     string   `json:"symbol" param:"symbol" validate:"required"`
	DateRange  DateRangeRequest `json:"date_range,omitempty"`
	Benchmarks []string `json:"benchmarks,omitempty"`
}

// TickerRankingRequest represents a request for ticker rankings
type TickerRankingRequest struct {
	Date      string `json:"date" query:"date" validate:"omitempty,datetime=2006-01-02"`
	Period    string `json:"period" query:"period" validate:"omitempty,oneof=daily weekly monthly"`
	Category  string `json:"category" query:"category" validate:"omitempty,oneof=gainers losers active volume value"`
	Limit     int    `json:"limit" query:"limit" validate:"omitempty,min=1,max=100"`
	Sector    string `json:"sector" query:"sector"`
}

// Company API Requests

// CompanyListRequest represents a request to list companies
type CompanyListRequest struct {
	PaginationRequest
	Filter domain.CompanyFilter `json:"filter,omitempty"`
}

// CompanyDetailsRequest represents a request for company details
type CompanyDetailsRequest struct {
	CompanyID string `json:"company_id" param:"id" validate:"required,uuid"`
	IncludeFinancials bool `json:"include_financials" query:"include_financials"`
	IncludeOfficers bool `json:"include_officers" query:"include_officers"`
	IncludeEvents bool `json:"include_events" query:"include_events"`
}

// Analytics API Requests

// AnalyticsRequest represents a general analytics request
type AnalyticsRequest struct {
	Type       string                   `json:"type" validate:"required,oneof=market sector ticker portfolio correlation risk"`
	DateRange  DateRangeRequest         `json:"date_range" validate:"required"`
	Options    domain.AnalyticsOptions  `json:"options"`
}

// MarketAnalyticsRequest represents a market analytics request
type MarketAnalyticsRequest struct {
	DateRange  DateRangeRequest `json:"date_range" validate:"required"`
	Metrics    []string         `json:"metrics,omitempty"`
	Sectors    []string         `json:"sectors,omitempty"`
	TimeFrame  string           `json:"time_frame" validate:"omitempty,oneof=daily weekly monthly quarterly yearly"`
}

// PortfolioAnalysisRequest represents a portfolio analysis request
type PortfolioAnalysisRequest struct {
	Holdings   []PortfolioHoldingInput `json:"holdings" validate:"required,min=1,dive"`
	DateRange  DateRangeRequest        `json:"date_range,omitempty"`
	Benchmarks []string                `json:"benchmarks,omitempty"`
	Analytics  []string                `json:"analytics,omitempty"`
}

// PortfolioHoldingInput represents a portfolio holding in a request
type PortfolioHoldingInput struct {
	Symbol       string    `json:"symbol" validate:"required"`
	Quantity     int64     `json:"quantity" validate:"required,min=1"`
	AveragePrice float64   `json:"average_price" validate:"required,min=0"`
	PurchaseDate string    `json:"purchase_date" validate:"omitempty,datetime=2006-01-02"`
}

// WebSocket API Requests

// WebSocketSubscribeRequest represents a WebSocket subscription request
type WebSocketSubscribeRequest struct {
	Type     string   `json:"type" validate:"required,oneof=ticker operation market all"`
	Channels []string `json:"channels" validate:"required,min=1"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

// WebSocketUnsubscribeRequest represents a WebSocket unsubscription request
type WebSocketUnsubscribeRequest struct {
	Type     string   `json:"type" validate:"required,oneof=ticker operation market all"`
	Channels []string `json:"channels,omitempty"`
}

// Health API Requests

// HealthCheckRequest represents a health check request
type HealthCheckRequest struct {
	Verbose bool `json:"verbose" query:"verbose"`
	Include []string `json:"include" query:"include" validate:"omitempty,dive,oneof=license database websocket services"`
}

// System API Requests

// SystemConfigRequest represents a system configuration request
type SystemConfigRequest struct {
	Section string `json:"section" query:"section" validate:"omitempty,oneof=general security logging paths"`
}

// SystemLogsRequest represents a system logs request
type SystemLogsRequest struct {
	PaginationRequest
	Level     string           `json:"level" query:"level" validate:"omitempty,oneof=debug info warn error"`
	DateRange DateRangeRequest `json:"date_range,omitempty"`
	Component string           `json:"component" query:"component"`
	TraceID   string           `json:"trace_id" query:"trace_id"`
}