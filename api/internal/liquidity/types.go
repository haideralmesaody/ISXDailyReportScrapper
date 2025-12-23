package liquidity

import (
	"strings"
	"time"
)

// Window represents different time windows for liquidity calculations
type Window int

const (
	// Window20 represents 20-day rolling window
	Window20 Window = 20
	// Window60 represents 60-day rolling window
	Window60 Window = 60
	// Window120 represents 120-day rolling window
	Window120 Window = 120
)

// String returns the string representation of the window
func (w Window) String() string {
	switch w {
	case Window20:
		return "20d"
	case Window60:
		return "60d"
	case Window120:
		return "120d"
	default:
		return "unknown"
	}
}

// Days returns the number of days in the window
func (w Window) Days() int {
	return int(w)
}

// TradingDay represents a single day's trading data for a ticker
type TradingDay struct {
	Date          time.Time `json:"date"`
	Symbol        string    `json:"symbol"`
	Open          float64   `json:"open"`
	High          float64   `json:"high"`
	Low           float64   `json:"low"`
	Close         float64   `json:"close"`
	Volume        float64   `json:"volume"`        // Number of shares traded (kept for compatibility)
	ShareVolume   float64   `json:"share_volume"`   // Explicit: Number of shares traded
	Value         float64   `json:"value"`          // Trading value in IQD
	NumTrades     int       `json:"num_trades"`
	TradingStatus string    `json:"trading_status"` // "ACTIVE", "SUSPENDED", etc.
}

// IsValid checks if the trading day data is valid
func (td TradingDay) IsValid() bool {
	return td.Open > 0 && td.High > 0 && td.Low > 0 && td.Close > 0 &&
		td.Volume >= 0 && td.Value >= 0 && td.NumTrades >= 0 &&
		td.High >= td.Low && td.High >= td.Open && td.High >= td.Close &&
		td.Low <= td.Open && td.Low <= td.Close
}

// IsTrading checks if the ticker was actively trading on this day
func (td TradingDay) IsTrading() bool {
	// Normalize status string (handle case variations and whitespace)
	status := strings.ToLower(strings.TrimSpace(td.TradingStatus))

	// Check explicit status values (from CSV/Excel files)
	validStatuses := map[string]bool{
		"true":   true, // Primary status from CSV files
		"active": true, // Alternative status format
		"traded": true, // Alternative status format
		"1":      true, // Numeric boolean
		"yes":    true, // Alternative boolean
	}

	// Trading day if EITHER:
	// 1. Status explicitly indicates trading, OR
	// 2. Financial metrics indicate actual trading activity
	// This dual-check prevents missing trading days due to status field issues
	return (validStatuses[status] || td.Value > 0) &&
	       (td.NumTrades > 0 || td.Value > 0)
}

// Return calculates the daily return
func (td TradingDay) Return(prevClose float64) float64 {
	if prevClose <= 0 {
		return 0
	}
	return (td.Close - prevClose) / prevClose
}

// TickerMetrics contains all calculated liquidity metrics for a ticker
type TickerMetrics struct {
	Symbol           string    `json:"symbol"`
	Date             time.Time `json:"date"`
	Window           Window    `json:"window"`
	
	// Core liquidity components
	ILLIQ            float64   `json:"illiq"`             // Amihud illiquidity
	ILLIQScaled      float64   `json:"illiq_scaled"`      // Cross-sectionally scaled ILLIQ
	Value            float64   `json:"value"`             // Average trading value (IQD)
	ValueScaled      float64   `json:"value_scaled"`      // Cross-sectionally scaled value
	Continuity       float64   `json:"continuity"`        // Trading continuity
	ContinuityNL     float64   `json:"continuity_nl"`     // Non-linear continuity
	ContinuityScaled float64   `json:"continuity_scaled"` // Scaled continuity
	SpreadScaled     float64   `json:"spread_scaled"`     // Cross-sectionally scaled spread proxy
	
	// Penalty adjustments
	ImpactPenalty    float64   `json:"impact_penalty"`    // Price impact penalty (unified)
	ValuePenalty     float64   `json:"value_penalty"`     // Value penalty (unified)
	ActivityScore    float64   `json:"activity_score"`    // Unified activity score (0-1)
	
	// Final hybrid score
	HybridScore      float64   `json:"hybrid_score"`      // ISX Hybrid Liquidity Score
	HybridRank       int       `json:"hybrid_rank"`       // Relative ranking
	
	// Supporting metrics
	SpreadProxy      float64   `json:"spread_proxy"`      // Corwin-Schultz spread estimate
	TradingDays      int       `json:"trading_days"`      // Number of active trading days
	TotalDays        int       `json:"total_days"`        // Total days in window
	AvgReturn        float64   `json:"avg_return"`        // Average daily return
	ReturnVolatility float64   `json:"return_volatility"` // Return volatility
	
	// Safe trading values (in IQD)
	SafeValue_0_5    float64   `json:"safe_value_0_5"`    // Max value for 0.5% price impact
	SafeValue_1_0    float64   `json:"safe_value_1_0"`    // Max value for 1.0% price impact
	SafeValue_2_0    float64   `json:"safe_value_2_0"`    // Max value for 2.0% price impact
	OptimalTradeSize float64   `json:"optimal_trade_size"` // Recommended trade size
}

// IsValid checks if the metrics are valid
func (tm TickerMetrics) IsValid() bool {
	return tm.Symbol != "" && !tm.Date.IsZero() &&
		tm.TradingDays > 0 && tm.TotalDays > 0 &&
		tm.TradingDays <= tm.TotalDays
}

// PenaltyParams contains parameters for penalty functions
type PenaltyParams struct {
	// Piecewise penalty parameters
	PiecewiseP0       float64 `json:"piecewise_p0"`        // Base penalty level
	PiecewiseBeta     float64 `json:"piecewise_beta"`      // Low-price slope
	PiecewiseGamma    float64 `json:"piecewise_gamma"`     // High-price slope
	PiecewisePStar    float64 `json:"piecewise_p_star"`    // Transition price
	PiecewiseMaxMult  float64 `json:"piecewise_max_mult"`  // Maximum multiplier
	
	// Exponential penalty parameters
	ExponentialP0     float64 `json:"exponential_p0"`      // Base penalty level
	ExponentialAlpha  float64 `json:"exponential_alpha"`   // Exponential decay rate
	ExponentialMaxMult float64 `json:"exponential_max_mult"` // Maximum multiplier
}

// IsValid checks if penalty parameters are valid
func (pp PenaltyParams) IsValid() bool {
	return pp.PiecewiseP0 > 0 && pp.PiecewiseBeta > 0 && pp.PiecewiseGamma > 0 &&
		pp.PiecewisePStar > 0 && pp.PiecewiseMaxMult > 1 &&
		pp.ExponentialP0 > 0 && pp.ExponentialAlpha > 0 && pp.ExponentialMaxMult > 1
}

// ComponentWeights contains weights for different liquidity components
// Updated for 3-metric system: ILLIQ (40%), Volume (35%), Continuity (25%)
type ComponentWeights struct {
	Impact     float64 `json:"impact"`     // Price impact weight (ILLIQ) - 40%
	Value      float64 `json:"value"`      // Trading value weight - 35%
	Continuity float64 `json:"continuity"` // Continuity weight - 25%
	Spread     float64 `json:"spread"`     // DEPRECATED - No longer used in scoring
}

// IsValid checks if weights are valid (sum to 1)
// Updated for 3-metric system - ignores Spread
func (cw ComponentWeights) IsValid() bool {
	// 3-metric system: only Impact, Value, Continuity
	sum := cw.Impact + cw.Value + cw.Continuity
	return cw.Impact >= 0 && cw.Value >= 0 && cw.Continuity >= 0 &&
		sum > 0.99 && sum < 1.01 // Allow small floating point errors
}

// Normalize ensures weights sum to 1
// Updated for 3-metric system - ignores Spread
func (cw *ComponentWeights) Normalize() {
	// 3-metric system: only Impact, Value, Continuity
	sum := cw.Impact + cw.Value + cw.Continuity
	if sum > 0 {
		cw.Impact /= sum
		cw.Value /= sum
		cw.Continuity /= sum
		cw.Spread = 0 // Explicitly set to 0 - no longer used
	}
}

// CalibrationResult contains the results of parameter calibration
type CalibrationResult struct {
	OptimalParams     PenaltyParams     `json:"optimal_params"`
	OptimalWeights    ComponentWeights  `json:"optimal_weights"`
	CrossValidationR2 float64          `json:"cv_r2"`           // Cross-validation R²
	SpreadCorrelation float64          `json:"spread_corr"`     // Correlation with spread proxy
	OptimizationError error            `json:"-"`               // Error during optimization
	CalibrationDate   time.Time        `json:"calibration_date"`
	WindowUsed        Window           `json:"window_used"`
	NumTickers        int              `json:"num_tickers"`
	NumObservations   int              `json:"num_observations"`
}

// IsValid checks if calibration results are valid
func (cr CalibrationResult) IsValid() bool {
	return cr.OptimalParams.IsValid() && cr.OptimalWeights.IsValid() &&
		cr.CrossValidationR2 >= 0 && cr.CrossValidationR2 <= 1 &&
		cr.SpreadCorrelation >= -1 && cr.SpreadCorrelation <= 1 &&
		cr.NumTickers > 0 && cr.NumObservations > 0
}

// CalibrationConfig contains configuration for parameter calibration
type CalibrationConfig struct {
	// Grid search parameters
	ParamGridSize     int     `json:"param_grid_size"`     // Grid points per parameter
	MinIterations     int     `json:"min_iterations"`      // Minimum optimization iterations
	MaxIterations     int     `json:"max_iterations"`      // Maximum optimization iterations
	Tolerance         float64 `json:"tolerance"`           // Convergence tolerance
	
	// Cross-validation settings
	KFolds            int     `json:"k_folds"`             // Number of CV folds
	RandomSeed        int64   `json:"random_seed"`         // Random seed for reproducibility
	
	// Optimization targets
	TargetMetric      string  `json:"target_metric"`       // "r2", "correlation", "combined"
	R2Weight          float64 `json:"r2_weight"`           // Weight for R² in combined metric
	CorrelationWeight float64 `json:"correlation_weight"`  // Weight for correlation in combined metric
	
	// Constraints
	MinTradingDays    int     `json:"min_trading_days"`    // Minimum trading days required
	MinTickers        int     `json:"min_tickers"`         // Minimum tickers required
	
	// Performance settings
	MaxConcurrency    int     `json:"max_concurrency"`     // Maximum concurrent calibrations
	EnableProfiling   bool    `json:"enable_profiling"`    // Enable performance profiling
}

// IsValid checks if calibration config is valid
func (cc CalibrationConfig) IsValid() bool {
	return cc.ParamGridSize > 0 && cc.MinIterations > 0 && cc.MaxIterations >= cc.MinIterations &&
		cc.Tolerance > 0 && cc.KFolds > 1 && cc.MinTradingDays > 0 && cc.MinTickers > 0 &&
		cc.MaxConcurrency > 0 && cc.R2Weight >= 0 && cc.CorrelationWeight >= 0 &&
		(cc.R2Weight + cc.CorrelationWeight) > 0
}

// WinsorizationBounds contains parameters for outlier handling
type WinsorizationBounds struct {
	Lower float64 `json:"lower"` // Lower percentile for winsorization
	Upper float64 `json:"upper"` // Upper percentile for winsorization
}

// IsValid checks if winsorization bounds are valid
func (wb WinsorizationBounds) IsValid() bool {
	return wb.Lower >= 0 && wb.Upper <= 1 && wb.Lower < wb.Upper
}

// Constants for default values
const (
	// Default winsorization bounds (5th and 95th percentiles)
	DefaultLowerBound = 0.05
	DefaultUpperBound = 0.95
	
	// Default continuity non-linear parameter
	DefaultContinuityDelta = 0.5
	
	// Minimum number of observations for reliable calculations
	MinObservationsForCalc = 10
	MinTradingDaysForCalc  = 5
	
	// Default timeout for calculations
	DefaultCalculationTimeout = 30 * time.Second
)

// GapPenaltyConfig configures the gap-based penalty calculation
type GapPenaltyConfig struct {
	// Gap length thresholds
	ShortGapThreshold  int `json:"short_gap_threshold"`  // Days (e.g., 2)
	MediumGapThreshold int `json:"medium_gap_threshold"` // Days (e.g., 7)

	// Penalty rates per day for each category
	ShortGapPenaltyRate  float64 `json:"short_gap_penalty_rate"`  // e.g., 0.05 (5% per day)
	MediumGapPenaltyRate float64 `json:"medium_gap_penalty_rate"` // e.g., 0.10 (10% per day)
	LongGapPenaltyRate   float64 `json:"long_gap_penalty_rate"`   // e.g., 0.20 (20% per day)

	// Gap forgiveness parameters
	AllowedGapLength int `json:"allowed_gap_length"` // Maximum gap length to forgive (e.g., 5 days)
	AllowedGapCount  int `json:"allowed_gap_count"`  // Number of gaps to forgive per window (e.g., 1)
	
	// Additional penalty factors
	EnableFrequencyPenalty  bool `json:"enable_frequency_penalty"`
	EnableClusteringPenalty bool `json:"enable_clustering_penalty"`

	// Maximum penalty cap
	MaxPenalty float64 `json:"max_penalty"` // e.g., 10.0 (max 10x worse)
}

// DefaultGapPenaltyConfig returns recommended default configuration
func DefaultGapPenaltyConfig() GapPenaltyConfig {
	return GapPenaltyConfig{
		ShortGapThreshold:       2,
		MediumGapThreshold:      7,
		ShortGapPenaltyRate:     0.05, // 5% per day for 1-2 day gaps
		MediumGapPenaltyRate:    0.10, // 10% per day for 3-7 day gaps
		LongGapPenaltyRate:      0.20, // 20% per day for >7 day gaps
		AllowedGapLength:        5,    // Forgive gaps up to 5 days (for meetings/holidays)
		AllowedGapCount:         1,    // Forgive 1 gap per calculation window
		EnableFrequencyPenalty:  true,
		EnableClusteringPenalty: true,
		MaxPenalty:              10.0, // Maximum 10x penalty
	}
}

// IsValid checks if gap penalty config is valid
func (gpc GapPenaltyConfig) IsValid() bool {
	return gpc.ShortGapThreshold > 0 && gpc.MediumGapThreshold > gpc.ShortGapThreshold &&
		gpc.ShortGapPenaltyRate >= 0 && gpc.MediumGapPenaltyRate >= 0 &&
		gpc.LongGapPenaltyRate >= 0 && gpc.MaxPenalty > 1
}

// GapInfo contains detailed information about a trading gap
type GapInfo struct {
	StartIndex int       `json:"start_index"`
	EndIndex   int       `json:"end_index"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Length     int       `json:"length"`
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return ve.Message
}

// TickerData represents data for a single ticker used in liquidity calculations
type TickerData struct {
	Ticker string  `json:"ticker"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Value  float64 `json:"value"`
}