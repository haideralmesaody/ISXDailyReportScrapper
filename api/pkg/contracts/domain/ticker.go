package domain

import (
	"time"
)

// Ticker represents a stock ticker/symbol
type Ticker struct {
	ID              string                 `json:"id" db:"id" validate:"required,uuid"`
	Symbol          string                 `json:"symbol" db:"symbol" validate:"required,min=1,max=10"`
	CompanyID       string                 `json:"company_id" db:"company_id" validate:"required,uuid"`
	CompanyName     string                 `json:"company_name" db:"company_name" validate:"required"`
	CompanyNameAr   string                 `json:"company_name_ar,omitempty" db:"company_name_ar"`
	ISINCode        string                 `json:"isin_code" db:"isin_code" validate:"required,len=12"`
	Sector          string                 `json:"sector" db:"sector" validate:"required"`
	SubSector       string                 `json:"sub_sector,omitempty" db:"sub_sector"`
	MarketCap       float64                `json:"market_cap" db:"market_cap"`
	Currency        string                 `json:"currency" db:"currency" validate:"required,len=3"`
	Status          TickerStatus           `json:"status" db:"status"`
	ListingDate     time.Time              `json:"listing_date" db:"listing_date"`
	DelistingDate   *time.Time             `json:"delisting_date,omitempty" db:"delisting_date"`
	LastTradeDate   time.Time              `json:"last_trade_date" db:"last_trade_date"`
	LastPrice       float64                `json:"last_price" db:"last_price"`
	PreviousClose   float64                `json:"previous_close" db:"previous_close"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// TickerStatus represents the trading status of a ticker
type TickerStatus string

const (
	TickerStatusActive    TickerStatus = "active"
	TickerStatusSuspended TickerStatus = "suspended"
	TickerStatusDelisted  TickerStatus = "delisted"
	TickerStatusHalted    TickerStatus = "halted"
)

// TickerData represents historical ticker data
type TickerData struct {
	ID              string    `json:"id" db:"id" validate:"required,uuid"`
	TickerID        string    `json:"ticker_id" db:"ticker_id" validate:"required,uuid"`
	Symbol          string    `json:"symbol" db:"symbol" validate:"required"`
	Date            time.Time `json:"date" db:"date" validate:"required"`
	Open            float64   `json:"open" db:"open" validate:"min=0"`
	High            float64   `json:"high" db:"high" validate:"min=0"`
	Low             float64   `json:"low" db:"low" validate:"min=0"`
	Close           float64   `json:"close" db:"close" validate:"min=0"`
	AdjustedClose   float64   `json:"adjusted_close" db:"adjusted_close" validate:"min=0"`
	Volume          int64     `json:"volume" db:"volume" validate:"min=0"`
	Value           float64   `json:"value" db:"value" validate:"min=0"`
	Trades          int       `json:"trades" db:"trades" validate:"min=0"`
	Change          float64   `json:"change" db:"change"`
	ChangePercent   float64   `json:"change_percent" db:"change_percent"`
	VWAP            float64   `json:"vwap" db:"vwap"` // Volume Weighted Average Price
	MarketCap       float64   `json:"market_cap,omitempty" db:"market_cap"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// TickerAnalysis represents analysis data for a ticker
type TickerAnalysis struct {
	ID              string                 `json:"id" db:"id" validate:"required,uuid"`
	TickerID        string                 `json:"ticker_id" db:"ticker_id" validate:"required,uuid"`
	Symbol          string                 `json:"symbol" db:"symbol" validate:"required"`
	AnalysisDate    time.Time              `json:"analysis_date" db:"analysis_date"`
	Period          string                 `json:"period" db:"period" validate:"required,oneof=daily weekly monthly quarterly yearly"`
	Metrics         TickerMetrics          `json:"metrics" db:"metrics"`
	Signals         []TradingSignal        `json:"signals,omitempty" db:"signals"`
	Score           float64                `json:"score" db:"score"` // 0-100
	Rating          string                 `json:"rating" db:"rating" validate:"omitempty,oneof=buy hold sell strong_buy strong_sell"`
	Confidence      float64                `json:"confidence" db:"confidence"` // 0-1
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
}

// TickerMetrics represents various metrics for a ticker
type TickerMetrics struct {
	Beta            float64 `json:"beta"`
	PE              float64 `json:"pe_ratio"`
	EPS             float64 `json:"eps"`
	DividendYield   float64 `json:"dividend_yield"`
	ROE             float64 `json:"roe"` // Return on Equity
	ROA             float64 `json:"roa"` // Return on Assets
	DebtToEquity    float64 `json:"debt_to_equity"`
	CurrentRatio    float64 `json:"current_ratio"`
	QuickRatio      float64 `json:"quick_ratio"`
	GrossMargin     float64 `json:"gross_margin"`
	OperatingMargin float64 `json:"operating_margin"`
	NetMargin       float64 `json:"net_margin"`
	BookValue       float64 `json:"book_value"`
	PriceToBook     float64 `json:"price_to_book"`
	RevenueGrowth   float64 `json:"revenue_growth"`
	EarningsGrowth  float64 `json:"earnings_growth"`
}


// TradingSignal represents a trading signal
type TradingSignal struct {
	Type        string    `json:"type" validate:"required,oneof=buy sell hold"`
	Source      string    `json:"source" validate:"required"`
	Strength    float64   `json:"strength" validate:"min=0,max=1"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// TickerFilter represents filters for ticker queries
type TickerFilter struct {
	Symbols       []string      `json:"symbols,omitempty"`
	Sectors       []string      `json:"sectors,omitempty"`
	SubSectors    []string      `json:"sub_sectors,omitempty"`
	Statuses      []TickerStatus `json:"statuses,omitempty"`
	MinMarketCap  float64       `json:"min_market_cap,omitempty"`
	MaxMarketCap  float64       `json:"max_market_cap,omitempty"`
	MinPrice      float64       `json:"min_price,omitempty"`
	MaxPrice      float64       `json:"max_price,omitempty"`
	MinVolume     int64         `json:"min_volume,omitempty"`
	MaxVolume     int64         `json:"max_volume,omitempty"`
	SearchTerm    string        `json:"search_term,omitempty"`
}