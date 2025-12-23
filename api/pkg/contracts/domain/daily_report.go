package domain

import (
	"time"
)

// TradeRecord represents a single company's trading data for one day.
// This is the primary data structure for ISX daily report entries.
type TradeRecord struct {
	CompanyName      string    `json:"company_name" db:"company_name" validate:"required"`
	CompanySymbol    string    `json:"company_symbol" db:"company_symbol" validate:"required"`
	Date             time.Time `json:"date" db:"date" validate:"required"`
	OpenPrice        float64   `json:"open_price" db:"open_price" validate:"min=0"`
	HighPrice        float64   `json:"high_price" db:"high_price" validate:"min=0"`
	LowPrice         float64   `json:"low_price" db:"low_price" validate:"min=0"`
	AveragePrice     float64   `json:"average_price" db:"average_price" validate:"min=0"`
	PrevAveragePrice float64   `json:"prev_average_price" db:"prev_average_price" validate:"min=0"`
	ClosePrice       float64   `json:"close_price" db:"close_price" validate:"min=0"`
	PrevClosePrice   float64   `json:"prev_close_price" db:"prev_close_price" validate:"min=0"`
	Change           float64   `json:"change" db:"change"`
	ChangePercent    float64   `json:"change_percent" db:"change_percent"`
	NumTrades        int64     `json:"num_trades" db:"num_trades" validate:"min=0"`
	Volume           int64     `json:"volume" db:"volume" validate:"min=0"`
	Value            float64   `json:"value" db:"value" validate:"min=0"`
	TradingStatus    bool      `json:"trading_status" db:"trading_status"` // true if actively traded, false if forward-filled
}

// DailyReport represents all trades in a single day's ISX report file.
// It contains the complete set of trading records for all companies
// that were included in the daily bulletin.
type DailyReport struct {
	Records []TradeRecord `json:"records" validate:"required,dive"`
}

// DailyReportSummary represents aggregated statistics for a daily report.
// This is used for providing overview information about the trading day.
type DailyReportSummary struct {
	Date              time.Time `json:"date" validate:"required"`
	TotalCompanies    int       `json:"total_companies" validate:"min=0"`
	ActivelyTraded    int       `json:"actively_traded" validate:"min=0"`
	TotalVolume       int64     `json:"total_volume" validate:"min=0"`
	TotalValue        float64   `json:"total_value" validate:"min=0"`
	TotalTrades       int64     `json:"total_trades" validate:"min=0"`
	AdvancingStocks   int       `json:"advancing_stocks" validate:"min=0"`
	DecliningStocks   int       `json:"declining_stocks" validate:"min=0"`
	UnchangedStocks   int       `json:"unchanged_stocks" validate:"min=0"`
	GeneratedAt       time.Time `json:"generated_at"`
}

// DailyReportFilter represents filters for querying trade records.
// This is used for filtering and searching through historical report data.
type DailyReportFilter struct {
	Symbols      []string   `json:"symbols,omitempty"`
	DateFrom     *time.Time `json:"date_from,omitempty"`
	DateTo       *time.Time `json:"date_to,omitempty"`
	MinVolume    int64      `json:"min_volume,omitempty"`
	MaxVolume    int64      `json:"max_volume,omitempty"`
	MinValue     float64    `json:"min_value,omitempty"`
	MaxValue     float64    `json:"max_value,omitempty"`
	MinChange    float64    `json:"min_change,omitempty"`
	MaxChange    float64    `json:"max_change,omitempty"`
	TradingOnly  bool       `json:"trading_only,omitempty"` // Filter out forward-filled records
}

// DailyReportMetadata represents metadata about a processed report file.
// This tracks the source and processing information for audit purposes.
type DailyReportMetadata struct {
	ID             string    `json:"id" db:"id" validate:"required,uuid"`
	FileName       string    `json:"file_name" db:"file_name" validate:"required"`
	FileSize       int64     `json:"file_size" db:"file_size" validate:"min=0"`
	ReportDate     time.Time `json:"report_date" db:"report_date" validate:"required"`
	ProcessedAt    time.Time `json:"processed_at" db:"processed_at"`
	RecordCount    int       `json:"record_count" db:"record_count" validate:"min=0"`
	ProcessingTime int64     `json:"processing_time_ms" db:"processing_time_ms"` // milliseconds
	Status         string    `json:"status" db:"status" validate:"required,oneof=pending processing completed failed"`
	ErrorMessage   string    `json:"error_message,omitempty" db:"error_message"`
}