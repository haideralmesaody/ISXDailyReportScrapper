package liquidity

import (
	"math"
)

// CorwinSchultz calculates the Corwin-Schultz (2012) high-low spread estimator
// This provides a proxy for bid-ask spreads using only daily OHLC data
//
// The method uses high-low price ratios across two consecutive trading days
// to estimate the effective bid-ask spread
//
// Parameters:
//   - high1, low1: High and low prices for day t-1
//   - high2, low2: High and low prices for day t
//
// Returns: estimated bid-ask spread as a fraction of price
//
// Reference: Corwin, S.A. and Schultz, P., 2012. A simple way to estimate bid‐ask
// spreads from daily high and low prices. The Journal of Finance, 67(2), pp.719-760.
func CorwinSchultz(high1, low1, high2, low2 float64) float64 {
	// Validate inputs
	if high1 <= 0 || low1 <= 0 || high2 <= 0 || low2 <= 0 ||
		high1 < low1 || high2 < low2 {
		return 0
	}
	
	// Calculate beta (two-day high-low ratio)
	maxHigh := math.Max(high1, high2)
	minLow := math.Min(low1, low2)
	
	if minLow <= 0 {
		return 0
	}
	
	beta := math.Log(maxHigh / minLow)
	if beta <= 0 {
		return 0
	}
	
	// Calculate gamma (sum of individual day high-low ratios)
	gamma1 := math.Log(high1 / low1)
	gamma2 := math.Log(high2 / low2)
	gamma := gamma1 + gamma2
	
	if gamma <= 0 {
		return 0
	}
	
	// Calculate alpha (measure of spread component)
	// alpha = (sqrt(2*beta) - sqrt(beta)) / (3 - 2*sqrt(2)) - sqrt(gamma/(3-2*sqrt(2)))
	
	sqrt2 := math.Sqrt(2)
	sqrtBeta := math.Sqrt(beta)
	denominator := 3 - 2*sqrt2
	
	if denominator <= 0 {
		return 0
	}
	
	// First term: (sqrt(2*beta) - sqrt(beta)) / (3 - 2*sqrt(2))
	term1 := (sqrt2*sqrtBeta - sqrtBeta) / denominator
	
	// Second term: sqrt(gamma / (3 - 2*sqrt(2)))
	term2 := math.Sqrt(gamma / denominator)
	
	alpha := term1 - term2
	
	// The spread estimate is 2 * (e^alpha - 1) / (1 + e^alpha)
	if alpha > 10 { // Prevent overflow
		return 1.0 // Maximum theoretical spread
	}
	if alpha < -10 { // Prevent underflow
		return 0.0
	}
	
	expAlpha := math.Exp(alpha)
	spread := 2 * (expAlpha - 1) / (1 + expAlpha)
	
	// Ensure spread is non-negative and reasonable
	if spread < 0 {
		spread = 0
	}
	if spread > 1 {
		spread = 1 // Cap at 100% spread (theoretical maximum)
	}
	
	// Handle NaN or Inf results
	if math.IsNaN(spread) || math.IsInf(spread, 0) {
		return 0
	}
	
	return spread
}

// CalculateSpreadSeries computes Corwin-Schultz spreads for a time series
// This handles the rolling calculation across consecutive trading days
func CalculateSpreadSeries(data []TradingDay) []float64 {
	if len(data) < 2 {
		return nil
	}
	
	// Find valid trading day pairs
	var spreads []float64
	var validDates []int // Track which dates have valid spreads
	
	for i := 1; i < len(data); i++ {
		day1 := data[i-1]
		day2 := data[i]
		
		// Both days must be valid trading days
		if !day1.IsValid() || !day2.IsValid() || 
		   !day1.IsTrading() || !day2.IsTrading() {
			continue
		}
		
		spread := CorwinSchultz(day1.High, day1.Low, day2.High, day2.Low)
		spreads = append(spreads, spread)
		validDates = append(validDates, i)
	}
	
	return spreads
}

// CalculateRollingSpread computes rolling average of Corwin-Schultz spreads
// This provides a smoother spread estimate over a specified window
func CalculateRollingSpread(data []TradingDay, windowSize int) []float64 {
	if len(data) < windowSize+1 || windowSize < 2 {
		return nil
	}
	
	var rollingSpreads []float64
	
	for i := windowSize; i < len(data); i++ {
		windowData := data[i-windowSize : i+1]
		windowSpreads := CalculateSpreadSeries(windowData)
		
		if len(windowSpreads) > 0 {
			avgSpread := calculateMean(windowSpreads)
			rollingSpreads = append(rollingSpreads, avgSpread)
		} else {
			rollingSpreads = append(rollingSpreads, 0)
		}
	}
	
	return rollingSpreads
}

// CalculateWeightedSpread computes volume-weighted Corwin-Schultz spreads
// This gives more weight to high-volume days in the spread calculation
func CalculateWeightedSpread(data []TradingDay, windowSize int) []float64 {
	if len(data) < windowSize+1 || windowSize < 2 {
		return nil
	}
	
	var weightedSpreads []float64
	
	for i := windowSize; i < len(data); i++ {
		windowData := data[i-windowSize : i+1]
		weightedSpread := calculateVolumeWeightedSpread(windowData)
		weightedSpreads = append(weightedSpreads, weightedSpread)
	}
	
	return weightedSpreads
}

// calculateVolumeWeightedSpread computes volume-weighted spread for a window
func calculateVolumeWeightedSpread(data []TradingDay) float64 {
	if len(data) < 2 {
		return 0
	}
	
	var weightedSpreadSum float64
	var totalVolumeWeight float64
	
	for i := 1; i < len(data); i++ {
		day1 := data[i-1]
		day2 := data[i]
		
		if !day1.IsValid() || !day2.IsValid() || 
		   !day1.IsTrading() || !day2.IsTrading() {
			continue
		}
		
		spread := CorwinSchultz(day1.High, day1.Low, day2.High, day2.Low)
		
		// Weight by average volume of the two days
		avgVolume := (day1.Volume + day2.Volume) / 2
		if avgVolume > 0 && !math.IsNaN(spread) && !math.IsInf(spread, 0) {
			weightedSpreadSum += spread * avgVolume
			totalVolumeWeight += avgVolume
		}
	}
	
	if totalVolumeWeight > 0 {
		return weightedSpreadSum / totalVolumeWeight
	}
	
	return 0
}

// AdjustForMicrostructure applies microstructure adjustments to spread estimates
// This accounts for known biases in the Corwin-Schultz estimator
func AdjustForMicrostructure(spreads []float64, volumes []float64) []float64 {
	if len(spreads) != len(volumes) {
		return spreads
	}
	
	adjusted := make([]float64, len(spreads))
	
	for i, spread := range spreads {
		if i >= len(volumes) || volumes[i] <= 0 {
			adjusted[i] = spread
			continue
		}
		
		// Apply volume-based adjustment
		// Higher volume typically indicates lower spreads
		volumeAdjustment := math.Log(1 + volumes[i]/1000000) * 0.1
		adjustedSpread := spread * (1 - volumeAdjustment)
		
		// Ensure non-negative
		if adjustedSpread < 0 {
			adjustedSpread = 0
		}
		
		adjusted[i] = adjustedSpread
	}
	
	return adjusted
}

// ValidateCorwinSchultzInputs validates inputs for Corwin-Schultz calculation
func ValidateCorwinSchultzInputs(high1, low1, high2, low2 float64) error {
	prices := []float64{high1, low1, high2, low2}
	priceNames := []string{"high1", "low1", "high2", "low2"}
	
	// Check for positive prices
	for i, price := range prices {
		if price <= 0 {
			return &ValidationError{
				Field:   priceNames[i],
				Message: "price must be positive",
				Value:   price,
			}
		}
		
		if math.IsNaN(price) || math.IsInf(price, 0) {
			return &ValidationError{
				Field:   priceNames[i],
				Message: "price must be a valid number",
				Value:   price,
			}
		}
	}
	
	// Check high >= low constraints
	if high1 < low1 {
		return &ValidationError{
			Field:   "high1_low1",
			Message: "high1 must be >= low1",
			Value:   map[string]float64{"high1": high1, "low1": low1},
		}
	}
	
	if high2 < low2 {
		return &ValidationError{
			Field:   "high2_low2",
			Message: "high2 must be >= low2",
			Value:   map[string]float64{"high2": high2, "low2": low2},
		}
	}
	
	return nil
}

// CalculateIntraday Spread estimates spread using intraday price ranges
// This is an alternative when only daily OHLC data is available
func CalculateIntradaySpread(data []TradingDay) []float64 {
	if len(data) == 0 {
		return nil
	}
	
	var spreads []float64
	
	for _, td := range data {
		if !td.IsValid() || !td.IsTrading() {
			continue
		}
		
		// Simple high-low spread proxy
		if td.Close > 0 {
			hlSpread := (td.High - td.Low) / td.Close
			
			// Apply empirical adjustment factor (typically 0.3-0.5 for daily data)
			adjustedSpread := hlSpread * 0.4
			
			if !math.IsNaN(adjustedSpread) && !math.IsInf(adjustedSpread, 0) && adjustedSpread >= 0 {
				spreads = append(spreads, adjustedSpread)
			}
		}
	}
	
	return spreads
}

// BenchmarkSpreadEstimator compares different spread estimation methods
// This is useful for validation and method selection
func BenchmarkSpreadEstimator(data []TradingDay) map[string]float64 {
	results := make(map[string]float64)
	
	// Corwin-Schultz method
	csSpreads := CalculateSpreadSeries(data)
	if len(csSpreads) > 0 {
		results["corwin_schultz_mean"] = calculateMean(csSpreads)
		results["corwin_schultz_median"] = calculateMedian(csSpreads)
	}
	
	// Intraday method
	intradaySpreads := CalculateIntradaySpread(data)
	if len(intradaySpreads) > 0 {
		results["intraday_mean"] = calculateMean(intradaySpreads)
		results["intraday_median"] = calculateMedian(intradaySpreads)
	}
	
	// Rolling averages
	if len(data) > 5 {
		rolling5 := CalculateRollingSpread(data, 5)
		if len(rolling5) > 0 {
			results["rolling_5day_mean"] = calculateMean(rolling5)
		}
	}
	
	if len(data) > 10 {
		rolling10 := CalculateRollingSpread(data, 10)
		if len(rolling10) > 0 {
			results["rolling_10day_mean"] = calculateMean(rolling10)
		}
	}
	
	return results
}