package domain

import (
	"time"
)

// AnalyticsOptions represents options for analytics operations
type AnalyticsOptions struct {
	Type          AnalyticsType          `json:"type" validate:"required"`
	Period        string                 `json:"period" validate:"required,oneof=daily weekly monthly quarterly yearly custom"`
	DateFrom      time.Time              `json:"date_from" validate:"required"`
	DateTo        time.Time              `json:"date_to" validate:"required,gtefield=DateFrom"`
	Symbols       []string               `json:"symbols,omitempty"`
	Sectors       []string               `json:"sectors,omitempty"`
	Metrics       []string               `json:"metrics,omitempty"`
	Benchmarks    []string               `json:"benchmarks,omitempty"`
	Confidence    float64                `json:"confidence" validate:"min=0,max=1"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

// AnalyticsType defines the type of analytics
type AnalyticsType string

const (
	AnalyticsTypeMarket      AnalyticsType = "market"
	AnalyticsTypeSector      AnalyticsType = "sector"
	AnalyticsTypeTicker      AnalyticsType = "ticker"
	AnalyticsTypePortfolio   AnalyticsType = "portfolio"
	AnalyticsTypeCorrelation AnalyticsType = "correlation"
	AnalyticsTypeRisk        AnalyticsType = "risk"
	AnalyticsTypeCustom      AnalyticsType = "custom"
)

// MarketStatistics represents overall market statistics
type MarketStatistics struct {
	Date              time.Time              `json:"date"`
	TotalMarketCap    float64                `json:"total_market_cap"`
	TotalVolume       int64                  `json:"total_volume"`
	TotalValue        float64                `json:"total_value"`
	TotalTrades       int64                  `json:"total_trades"`
	ActiveSymbols     int                    `json:"active_symbols"`
	AdvanceDecline    AdvanceDeclineRatio    `json:"advance_decline"`
	MarketBreadth     MarketBreadth          `json:"market_breadth"`
	SectorPerformance []SectorPerformance    `json:"sector_performance"`
	TopGainers        []SymbolPerformance    `json:"top_gainers"`
	TopLosers         []SymbolPerformance    `json:"top_losers"`
	MostActive        []SymbolActivity       `json:"most_active"`
	}

// AdvanceDeclineRatio represents advance/decline statistics
type AdvanceDeclineRatio struct {
	Advancing int     `json:"advancing"`
	Declining int     `json:"declining"`
	Unchanged int     `json:"unchanged"`
	Ratio     float64 `json:"ratio"`
	Line      float64 `json:"line"`
}

// SectorPerformance represents performance metrics for a sector
type SectorPerformance struct {
	Sector           string               `json:"sector"`
	Performance      float64              `json:"performance"` // Percentage
	Volume           int64                `json:"volume"`
	Value            float64              `json:"value"`
	MarketCap        float64              `json:"market_cap"`
	ActiveSymbols    int                  `json:"active_symbols"`
	TopPerformers    []SymbolPerformance  `json:"top_performers,omitempty"`
	WorstPerformers  []SymbolPerformance  `json:"worst_performers,omitempty"`
}

// SymbolPerformance represents performance metrics for a symbol
type SymbolPerformance struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Volume        int64   `json:"volume"`
	Value         float64 `json:"value"`
}

// SymbolActivity represents activity metrics for a symbol
type SymbolActivity struct {
	Symbol     string  `json:"symbol"`
	Name       string  `json:"name"`
	Volume     int64   `json:"volume"`
	Value      float64 `json:"value"`
	Trades     int     `json:"trades"`
	Turnover   float64 `json:"turnover"`
}

// MarketTrend represents a market trend analysis
type MarketTrend struct {
	Period        string                 `json:"period"`
	Trend         string                 `json:"trend"` // bullish, bearish, neutral
	Strength      float64                `json:"strength"` // 0-1
	Momentum      float64                `json:"momentum"`
	Volatility    float64                `json:"volatility"`
	Support       []float64              `json:"support_levels"`
	Resistance    []float64              `json:"resistance_levels"`
	Indicators    TechnicalIndicators    `json:"indicators"`
	Forecast      MarketForecast         `json:"forecast,omitempty"`
}

// TechnicalIndicators represents key indicator readings for a market or symbol.
// This is intentionally minimal to keep the build passing while the richer indicator
// model is implemented.
type TechnicalIndicators struct {
	RSI             float64 `json:"rsi,omitempty"`
	MACD            float64 `json:"macd,omitempty"`
	SMA50           float64 `json:"sma_50,omitempty"`
	SMA200          float64 `json:"sma_200,omitempty"`
	BollingerUpper  float64 `json:"bollinger_upper,omitempty"`
	BollingerLower  float64 `json:"bollinger_lower,omitempty"`
	ATR             float64 `json:"atr,omitempty"`
	Volume          float64 `json:"volume,omitempty"`
}

// MarketForecast represents market forecast
type MarketForecast struct {
	Direction    string                 `json:"direction"` // up, down, sideways
	Target       float64                `json:"target"`
	Probability  float64                `json:"probability"`
	TimeFrame    string                 `json:"time_frame"`
	Factors      []string               `json:"factors"`
	Risks        []string               `json:"risks"`
	Confidence   float64                `json:"confidence"`
}

// CorrelationMatrix represents correlation analysis between symbols
type CorrelationMatrix struct {
	Symbols       []string              `json:"symbols"`
	Period        string                `json:"period"`
	Matrix        [][]float64           `json:"matrix"`
	Significance  [][]float64           `json:"significance"`
	Relationships []CorrelationPair     `json:"significant_relationships"`
}

// CorrelationPair represents a correlation between two symbols
type CorrelationPair struct {
	Symbol1      string  `json:"symbol1"`
	Symbol2      string  `json:"symbol2"`
	Correlation  float64 `json:"correlation"`
	Significance float64 `json:"significance"`
	Type         string  `json:"type"` // positive, negative, neutral
}

// RiskMetrics represents risk analysis metrics
type RiskMetrics struct {
	VaR95         float64               `json:"var_95"` // Value at Risk 95%
	VaR99         float64               `json:"var_99"` // Value at Risk 99%
	CVaR          float64               `json:"cvar"`   // Conditional VaR
	Beta          float64               `json:"beta"`
	StandardDev   float64               `json:"standard_deviation"`
	Sharpe        float64               `json:"sharpe_ratio"`
	Sortino       float64               `json:"sortino_ratio"`
	MaxDrawdown   float64               `json:"max_drawdown"`
	DownsideRisk  float64               `json:"downside_risk"`
	TrackingError float64               `json:"tracking_error"`
	InformationRatio float64            `json:"information_ratio"`
	StressTests   []StressTestResult    `json:"stress_tests,omitempty"`
}

// StressTestResult represents a stress test result
type StressTestResult struct {
	Scenario    string  `json:"scenario"`
	Impact      float64 `json:"impact"` // Percentage loss
	Probability float64 `json:"probability"`
	Description string  `json:"description"`
}

// PortfolioAnalysis represents portfolio analysis results
type PortfolioAnalysis struct {
	TotalValue       float64              `json:"total_value"`
	TotalCost        float64              `json:"total_cost"`
	UnrealizedPnL    float64              `json:"unrealized_pnl"`
	UnrealizedPnLPct float64              `json:"unrealized_pnl_pct"`
	RealizedPnL      float64              `json:"realized_pnl"`
	TotalReturn      float64              `json:"total_return"`
	AnnualizedReturn float64              `json:"annualized_return"`
	Holdings         []PortfolioHolding   `json:"holdings"`
	Allocation       []AllocationItem     `json:"allocation"`
	Performance      TradingPerformanceMetrics   `json:"performance"`
	Risk             RiskMetrics          `json:"risk"`
	Diversification  float64              `json:"diversification_ratio"`
}

// PortfolioHolding represents a holding in a portfolio
type PortfolioHolding struct {
	Symbol         string    `json:"symbol"`
	Quantity       int64     `json:"quantity"`
	AverageCost    float64   `json:"average_cost"`
	CurrentPrice   float64   `json:"current_price"`
	MarketValue    float64   `json:"market_value"`
	UnrealizedPnL  float64   `json:"unrealized_pnl"`
	Weight         float64   `json:"weight"` // Portfolio percentage
	FirstPurchase  time.Time `json:"first_purchase"`
	LastPurchase   time.Time `json:"last_purchase"`
}

// AllocationItem represents an allocation item
type AllocationItem struct {
	Category   string  `json:"category"` // sector, asset_class, etc.
	Value      float64 `json:"value"`
	Weight     float64 `json:"weight"`
	Target     float64 `json:"target,omitempty"`
	Deviation  float64 `json:"deviation,omitempty"`
}

// TradingPerformanceMetrics represents trading performance metrics
type TradingPerformanceMetrics struct {
	DailyReturn   float64           `json:"daily_return"`
	WeeklyReturn  float64           `json:"weekly_return"`
	MonthlyReturn float64           `json:"monthly_return"`
	YearlyReturn  float64           `json:"yearly_return"`
	YTDReturn     float64           `json:"ytd_return"`
	Volatility    float64           `json:"volatility"`
	WinRate       float64           `json:"win_rate"`
	ProfitFactor  float64           `json:"profit_factor"`
	ExpectedReturn float64          `json:"expected_return"`
	TrackingError float64           `json:"tracking_error,omitempty"`
	Alpha         float64           `json:"alpha,omitempty"`
	Beta          float64           `json:"beta,omitempty"`
}

// AnalyticsResult represents the result of an analytics operation
type AnalyticsResult struct {
	ID           string                 `json:"id"`
	Type         AnalyticsType          `json:"type"`
	Status       string                 `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	Duration     time.Duration          `json:"duration"`
	Options      AnalyticsOptions       `json:"options"`
	Data         interface{}            `json:"data"`
	Summary      map[string]interface{} `json:"summary"`
	Errors       []string               `json:"errors,omitempty"`
	Warnings     []string               `json:"warnings,omitempty"`
}

// Point represents a data point in time series
type Point struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

// VolumeLevel represents volume analysis level
type VolumeLevel string

const (
	VolumeLevelLow      VolumeLevel = "low"
	VolumeLevelNormal   VolumeLevel = "normal"
	VolumeLevelHigh     VolumeLevel = "high"
	VolumeLevelExtreme  VolumeLevel = "extreme"
)
