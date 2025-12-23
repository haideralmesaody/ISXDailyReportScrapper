package liquidity

import (
	"math"
)

// SafeTradingLimits contains the calculated safe trading values for different impact thresholds
type SafeTradingLimits struct {
	Symbol           string  `json:"symbol"`
	
	// Safe trading values for different price impact thresholds (in IQD)
	SafeValue_0_5    float64 `json:"safe_value_0_5"`    // Max value for 0.5% price impact
	SafeValue_1_0    float64 `json:"safe_value_1_0"`    // Max value for 1.0% price impact
	SafeValue_2_0    float64 `json:"safe_value_2_0"`    // Max value for 2.0% price impact
	
	// Optimal trade size considering all factors
	OptimalTradeSize float64 `json:"optimal_trade_size"` // Recommended size (IQD)
	
	// Supporting information
	MaxDailyPercent  float64 `json:"max_daily_percent"`  // Max % of daily volume
	SpreadCost       float64 `json:"spread_cost"`        // Estimated spread cost %
	LiquidityRating  string  `json:"liquidity_rating"`   // HIGH/MEDIUM/LOW/POOR
	
	// Constraints applied
	VolumeCap        float64 `json:"volume_cap"`         // Volume-based limit (IQD)
	ActivityAdjust   float64 `json:"activity_adjust"`    // Activity adjustment factor
	SpreadAdjust     float64 `json:"spread_adjust"`      // Spread adjustment factor
}

// CalculateSafeTrading calculates safe trading limits based on liquidity metrics
// This determines the maximum trading value that won't significantly impact the price
//
// The calculation considers:
//   - ILLIQ (price impact per million IQD traded)
//   - Average daily trading value
//   - Trading activity/continuity
//   - Bid-ask spread costs
//
// Returns SafeTradingLimits with values for different impact thresholds
func CalculateSafeTrading(metrics TickerMetrics) SafeTradingLimits {
	limits := SafeTradingLimits{
		Symbol: metrics.Symbol,
	}
	
	// Handle edge cases
	if metrics.ILLIQ <= 0 || math.IsNaN(metrics.ILLIQ) || math.IsInf(metrics.ILLIQ, 0) {
		// Invalid ILLIQ - return zero limits
		limits.LiquidityRating = "INVALID"
		return limits
	}
	
	// 1. Calculate base safe values from ILLIQ
	// ILLIQ = |Return %| / Value(millions), so Value = Impact / ILLIQ
	// Convert back to IQD (multiply by 1,000,000)
	limits.SafeValue_0_5 = (0.005 / metrics.ILLIQ) * 1_000_000  // 0.5% impact
	limits.SafeValue_1_0 = (0.010 / metrics.ILLIQ) * 1_000_000  // 1.0% impact
	limits.SafeValue_2_0 = (0.020 / metrics.ILLIQ) * 1_000_000  // 2.0% impact
	
	// 2. Apply volume constraint (typically 10-20% of average daily volume)
	// More conservative for less liquid stocks
	var maxDailyPercent float64
	if metrics.HybridScore >= 70 {
		maxDailyPercent = 0.20  // 20% for highly liquid stocks
		limits.LiquidityRating = "HIGH"
	} else if metrics.HybridScore >= 50 {
		maxDailyPercent = 0.15  // 15% for medium liquidity
		limits.LiquidityRating = "MEDIUM"
	} else if metrics.HybridScore >= 30 {
		maxDailyPercent = 0.10  // 10% for low liquidity
		limits.LiquidityRating = "LOW"
	} else {
		maxDailyPercent = 0.05  // 5% for poor liquidity
		limits.LiquidityRating = "POOR"
	}
	
	limits.MaxDailyPercent = maxDailyPercent * 100  // Store as percentage
	limits.VolumeCap = metrics.Value * maxDailyPercent
	
	// Apply volume cap to all thresholds
	if limits.SafeValue_0_5 > limits.VolumeCap {
		limits.SafeValue_0_5 = limits.VolumeCap
	}
	if limits.SafeValue_1_0 > limits.VolumeCap {
		limits.SafeValue_1_0 = limits.VolumeCap
	}
	if limits.SafeValue_2_0 > limits.VolumeCap {
		limits.SafeValue_2_0 = limits.VolumeCap
	}
	
	// 3. Adjust for trading activity (continuity)
	// Less active stocks need smaller trade sizes
	activityAdjustment := calculateActivityAdjustment(metrics.ActivityScore)
	limits.ActivityAdjust = activityAdjustment
	
	// 4. Adjust for spread costs
	// Higher spread means higher transaction costs, reduce trade size
	spreadAdjustment := calculateSpreadAdjustment(metrics.SpreadProxy)
	limits.SpreadAdjust = spreadAdjustment
	limits.SpreadCost = metrics.SpreadProxy * 100  // Store as percentage
	
	// 5. Calculate optimal trade size (target 1% impact with adjustments)
	baseOptimal := limits.SafeValue_1_0
	limits.OptimalTradeSize = baseOptimal * activityAdjustment * spreadAdjustment
	
	// 6. Apply minimum trade size (avoid micro trades)
	minTradeSize := 100_000.0  // 100K IQD minimum
	if limits.OptimalTradeSize < minTradeSize && metrics.Value > minTradeSize*10 {
		// Only apply minimum if daily volume is sufficient
		limits.OptimalTradeSize = minTradeSize
	}
	
	// 7. Final safety checks
	ensureSafeLimits(&limits)
	
	return limits
}

// calculateActivityAdjustment returns an adjustment factor based on trading activity
// More active stocks (higher ActivityScore) get less penalty
func calculateActivityAdjustment(activityScore float64) float64 {
	if activityScore >= 0.8 {
		return 1.0  // No penalty for very active stocks
	} else if activityScore >= 0.5 {
		return 0.85  // Small penalty for moderately active
	} else if activityScore >= 0.3 {
		return 0.70  // Moderate penalty
	} else if activityScore >= 0.1 {
		return 0.50  // Significant penalty
	}
	return 0.30  // Severe penalty for very inactive stocks
}

// calculateSpreadAdjustment returns an adjustment factor based on bid-ask spread
// Higher spreads get more penalty
func calculateSpreadAdjustment(spreadProxy float64) float64 {
	if spreadProxy <= 0.001 {
		return 1.0  // No penalty for tight spreads (< 0.1%)
	} else if spreadProxy <= 0.005 {
		return 0.95  // Small penalty for reasonable spreads (< 0.5%)
	} else if spreadProxy <= 0.01 {
		return 0.85  // Moderate penalty for 0.5-1% spreads
	} else if spreadProxy <= 0.02 {
		return 0.70  // Significant penalty for 1-2% spreads
	} else if spreadProxy <= 0.05 {
		return 0.50  // High penalty for 2-5% spreads
	}
	return 0.30  // Severe penalty for very wide spreads (> 5%)
}

// ensureSafeLimits applies final safety checks and constraints
func ensureSafeLimits(limits *SafeTradingLimits) {
	// Ensure no negative values
	if limits.SafeValue_0_5 < 0 {
		limits.SafeValue_0_5 = 0
	}
	if limits.SafeValue_1_0 < 0 {
		limits.SafeValue_1_0 = 0
	}
	if limits.SafeValue_2_0 < 0 {
		limits.SafeValue_2_0 = 0
	}
	if limits.OptimalTradeSize < 0 {
		limits.OptimalTradeSize = 0
	}
	
	// Ensure logical ordering (0.5% < 1% < 2%)
	if limits.SafeValue_1_0 < limits.SafeValue_0_5 {
		limits.SafeValue_1_0 = limits.SafeValue_0_5
	}
	if limits.SafeValue_2_0 < limits.SafeValue_1_0 {
		limits.SafeValue_2_0 = limits.SafeValue_1_0
	}
	
	// Cap at reasonable maximum (100M IQD)
	maxTradeSize := 100_000_000.0
	if limits.SafeValue_2_0 > maxTradeSize {
		limits.SafeValue_2_0 = maxTradeSize
		limits.SafeValue_1_0 = math.Min(limits.SafeValue_1_0, maxTradeSize)
		limits.SafeValue_0_5 = math.Min(limits.SafeValue_0_5, maxTradeSize)
		limits.OptimalTradeSize = math.Min(limits.OptimalTradeSize, maxTradeSize/2)
	}
}

// EstimateImpact estimates the price impact for a given trade size
// Returns the estimated price impact as a percentage
func EstimateImpact(metrics TickerMetrics, tradeValue float64) float64 {
	if metrics.ILLIQ <= 0 || tradeValue <= 0 {
		return 0
	}
	
	// ILLIQ = Impact / Value(millions)
	// So Impact = ILLIQ * Value(millions)
	valueMillions := tradeValue / 1_000_000
	estimatedImpact := metrics.ILLIQ * valueMillions
	
	// Apply non-linear adjustment for large trades
	// Large trades have disproportionate impact
	if valueMillions > metrics.Value/1_000_000 * 0.1 {
		// Trade is > 10% of daily volume, apply exponential penalty
		volumeRatio := valueMillions / (metrics.Value / 1_000_000)
		nonLinearMultiplier := math.Exp(volumeRatio - 0.1)
		estimatedImpact *= nonLinearMultiplier
	}
	
	return estimatedImpact * 100  // Return as percentage
}

// RecommendTradeSchedule suggests how to split a large trade to minimize impact
type TradeSchedule struct {
	TotalValue      float64   `json:"total_value"`
	NumTranches     int       `json:"num_tranches"`
	TrancheSize     float64   `json:"tranche_size"`
	IntervalMinutes int       `json:"interval_minutes"`
	EstimatedImpact float64   `json:"estimated_impact"`
	Recommendation  string    `json:"recommendation"`
}

// CreateTradeSchedule creates an execution schedule for large trades
func CreateTradeSchedule(metrics TickerMetrics, totalTradeValue float64) TradeSchedule {
	schedule := TradeSchedule{
		TotalValue: totalTradeValue,
	}
	
	// Get safe trading limits
	limits := CalculateSafeTrading(metrics)
	
	// Determine appropriate tranche size
	var trancheSize float64
	
	if totalTradeValue <= limits.OptimalTradeSize {
		// Single trade is fine
		schedule.NumTranches = 1
		schedule.TrancheSize = totalTradeValue
		schedule.IntervalMinutes = 0
		schedule.EstimatedImpact = EstimateImpact(metrics, totalTradeValue)
		schedule.Recommendation = "Execute as single trade - within safe limits"
	} else if totalTradeValue <= limits.SafeValue_1_0 * 3 {
		// Split into 3-5 tranches
		schedule.NumTranches = 3
		trancheSize = totalTradeValue / 3
		schedule.IntervalMinutes = 30  // 30 minutes between tranches
		schedule.Recommendation = "Split into 3 trades over 1.5 hours"
	} else if totalTradeValue <= limits.SafeValue_1_0 * 10 {
		// Split into 5-10 tranches
		schedule.NumTranches = int(math.Ceil(totalTradeValue / limits.SafeValue_1_0))
		trancheSize = totalTradeValue / float64(schedule.NumTranches)
		schedule.IntervalMinutes = 20
		schedule.Recommendation = "Split into multiple trades throughout the day"
	} else {
		// Very large trade - needs multiple days
		daysNeeded := int(math.Ceil(totalTradeValue / (limits.SafeValue_1_0 * 5)))
		schedule.NumTranches = daysNeeded * 5
		trancheSize = totalTradeValue / float64(schedule.NumTranches)
		schedule.IntervalMinutes = 60
		schedule.Recommendation = "Execute over multiple trading days"
	}
	
	schedule.TrancheSize = trancheSize
	schedule.EstimatedImpact = EstimateImpact(metrics, trancheSize)
	
	return schedule
}