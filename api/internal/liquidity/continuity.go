package liquidity

import (
	"math"
)

// ContinuityNL applies a non-linear transformation to trading continuity
// This transformation addresses the non-linear relationship between raw continuity
// and liquidity, giving higher weight to improvements at low continuity levels
//
// Parameters:
//   - cont: raw continuity ratio (0 to 1, where 1 = traded every day)
//   - delta: transformation parameter controlling non-linearity (typically 0.5)
//
// Returns: transformed continuity score
func ContinuityNL(cont, delta float64) float64 {
	if cont < 0 {
		return 0
	}
	if cont > 1 {
		cont = 1
	}
	if delta <= 0 {
		return cont // Linear transformation if delta invalid
	}
	
	// Non-linear transformation: cont^(1-delta)
	// When delta = 0.5, this becomes sqrt(cont)
	// Higher delta values make the transformation more aggressive
	exponent := 1.0 - delta
	transformed := math.Pow(cont, exponent)
	
	// Handle edge cases
	if math.IsNaN(transformed) || math.IsInf(transformed, 0) {
		return cont
	}
	
	return transformed
}

// CalculateContinuity computes the basic trading continuity ratio
// This is the foundation metric before non-linear transformation
//
// Parameters:
//   - data: slice of TradingDay data
//
// Returns: continuity ratio (trading days / total days)
func CalculateContinuity(data []TradingDay) float64 {
	if len(data) == 0 {
		return 0
	}
	
	tradingDays := 0
	totalDays := len(data)
	
	for _, td := range data {
		if td.IsTrading() {
			tradingDays++
		}
	}
	
	return float64(tradingDays) / float64(totalDays)
}

// CalculateWeightedContinuity computes volume-weighted trading continuity
// This gives more weight to high-volume trading days
//
// Parameters:
//   - data: slice of TradingDay data
//
// Returns: volume-weighted continuity ratio
func CalculateWeightedContinuity(data []TradingDay) float64 {
	if len(data) == 0 {
		return 0
	}
	
	totalVolumeWeight := 0.0
	tradingVolumeWeight := 0.0
	maxVolume := findMaxVolume(data)
	
	if maxVolume == 0 {
		return CalculateContinuity(data) // Fallback to unweighted
	}
	
	for _, td := range data {
		// Normalize volume to [0,1] for weighting
		volumeWeight := td.Volume / maxVolume
		totalVolumeWeight += volumeWeight
		
		if td.IsTrading() {
			tradingVolumeWeight += volumeWeight
		}
	}
	
	if totalVolumeWeight == 0 {
		return 0
	}
	
	return tradingVolumeWeight / totalVolumeWeight
}

// CalculateConsistentContinuity measures consistency of trading activity
// This penalizes sporadic trading patterns and rewards consistent activity
//
// Parameters:
//   - data: slice of TradingDay data
//   - windowSize: rolling window size for consistency measurement
//
// Returns: consistency-adjusted continuity score
func CalculateConsistentContinuity(data []TradingDay, windowSize int) float64 {
	if len(data) < windowSize || windowSize <= 0 {
		return CalculateContinuity(data)
	}
	
	var rollingContinuities []float64
	
	// Calculate rolling continuity for each window
	for i := windowSize - 1; i < len(data); i++ {
		windowData := data[i-windowSize+1 : i+1]
		rollCont := CalculateContinuity(windowData)
		rollingContinuities = append(rollingContinuities, rollCont)
	}
	
	if len(rollingContinuities) == 0 {
		return 0
	}
	
	// Calculate mean and standard deviation of rolling continuities
	meanCont := calculateMean(rollingContinuities)
	stdCont := calculateStdDev(rollingContinuities, meanCont)
	
	// Consistency penalty: higher standard deviation = lower consistency
	// Use coefficient of variation as consistency measure
	if meanCont > 0 {
		consistencyPenalty := stdCont / meanCont
		// Apply penalty (max penalty of 50%)
		adjustedContinuity := meanCont * (1.0 - math.Min(0.5, consistencyPenalty))
		return math.Max(0, adjustedContinuity)
	}
	
	return meanCont
}

// CalculateTemporalContinuity measures continuity considering temporal patterns
// This accounts for weekends, holidays, and other market closures
//
// Parameters:
//   - data: slice of TradingDay data (must be sorted by date)
//   - expectedTradingDays: expected number of trading days in the period
//
// Returns: temporal-adjusted continuity score
func CalculateTemporalContinuity(data []TradingDay, expectedTradingDays int) float64 {
	if len(data) == 0 || expectedTradingDays <= 0 {
		return 0
	}
	
	tradingDays := 0
	for _, td := range data {
		if td.IsTrading() {
			tradingDays++
		}
	}
	
	return float64(tradingDays) / float64(expectedTradingDays)
}

// CalculateGapAdjustedContinuity penalizes trading gaps
// This considers the impact of consecutive non-trading days
//
// Parameters:
//   - data: slice of TradingDay data (must be sorted by date)
//   - maxGapPenalty: maximum penalty for long gaps (0 to 1)
//
// Returns: gap-adjusted continuity score
func CalculateGapAdjustedContinuity(data []TradingDay, maxGapPenalty float64) float64 {
	if len(data) == 0 {
		return 0
	}
	
	baseContinuity := CalculateContinuity(data)
	if baseContinuity == 0 {
		return 0
	}
	
	// Identify gaps in trading
	gaps := identifyTradingGaps(data)
	if len(gaps) == 0 {
		return baseContinuity
	}
	
	// Calculate gap penalty
	totalGapPenalty := 0.0
	for _, gapLength := range gaps {
		// Penalty increases with gap length (logarithmically)
		gapPenalty := math.Log(1 + float64(gapLength)) * 0.1
		totalGapPenalty += math.Min(maxGapPenalty, gapPenalty)
	}
	
	// Apply penalty (normalize by number of gaps)
	averageGapPenalty := totalGapPenalty / float64(len(gaps))
	adjustedContinuity := baseContinuity * (1.0 - averageGapPenalty)
	
	return math.Max(0, adjustedContinuity)
}

// identifyTradingGaps finds consecutive non-trading day sequences
func identifyTradingGaps(data []TradingDay) []int {
	var gaps []int
	currentGap := 0
	
	for _, td := range data {
		if !td.IsTrading() {
			currentGap++
		} else {
			if currentGap > 0 {
				gaps = append(gaps, currentGap)
				currentGap = 0
			}
		}
	}
	
	// Don't forget the last gap if data ends with non-trading days
	if currentGap > 0 {
		gaps = append(gaps, currentGap)
	}
	
	return gaps
}

// findMaxVolume finds the maximum volume in the dataset
func findMaxVolume(data []TradingDay) float64 {
	maxVol := 0.0
	for _, td := range data {
		if td.Volume > maxVol {
			maxVol = td.Volume
		}
	}
	return maxVol
}

// calculateStdDev calculates standard deviation given mean
func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	
	sumSquaredDiff := 0.0
	validCount := 0
	
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			diff := v - mean
			sumSquaredDiff += diff * diff
			validCount++
		}
	}
	
	if validCount <= 1 {
		return 0
	}
	
	return math.Sqrt(sumSquaredDiff / float64(validCount-1))
}

// OptimizeNonLinearParameter finds the optimal delta parameter for non-linear transformation
// This uses cross-validation to find the delta that maximizes correlation with liquidity proxy
//
// Parameters:
//   - continuityValues: slice of raw continuity ratios
//   - liquidityProxy: corresponding liquidity proxy values (e.g., spread estimates)
//   - deltaRange: range of delta values to test
//
// Returns: optimal delta parameter
func OptimizeNonLinearParameter(continuityValues, liquidityProxy []float64, deltaRange []float64) float64 {
	if len(continuityValues) != len(liquidityProxy) || len(continuityValues) == 0 {
		return DefaultContinuityDelta
	}
	
	bestDelta := DefaultContinuityDelta
	bestCorrelation := 0.0
	
	for _, delta := range deltaRange {
		// Transform continuity values with current delta
		var transformedCont []float64
		for _, cont := range continuityValues {
			transformedCont = append(transformedCont, ContinuityNL(cont, delta))
		}
		
		// Calculate correlation with liquidity proxy
		correlation := calculateCorrelation(transformedCont, liquidityProxy)
		
		// Prefer higher correlation (negative correlation expected with spread proxy)
		if math.Abs(correlation) > math.Abs(bestCorrelation) {
			bestCorrelation = correlation
			bestDelta = delta
		}
	}
	
	return bestDelta
}

// calculateCorrelation computes Pearson correlation coefficient
func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}
	
	meanX := calculateMean(x)
	meanY := calculateMean(y)
	
	var sumXY, sumXX, sumYY float64
	validCount := 0
	
	for i := 0; i < len(x); i++ {
		if !math.IsNaN(x[i]) && !math.IsInf(x[i], 0) && !math.IsNaN(y[i]) && !math.IsInf(y[i], 0) {
			dx := x[i] - meanX
			dy := y[i] - meanY
			sumXY += dx * dy
			sumXX += dx * dx
			sumYY += dy * dy
			validCount++
		}
	}
	
	if validCount < 2 || sumXX == 0 || sumYY == 0 {
		return 0
	}
	
	correlation := sumXY / math.Sqrt(sumXX*sumYY)
	
	if math.IsNaN(correlation) || math.IsInf(correlation, 0) {
		return 0
	}
	
	return correlation
}

// ValidateContinuityInputs validates inputs for continuity calculations
func ValidateContinuityInputs(data []TradingDay, delta float64) error {
	if len(data) == 0 {
		return &ValidationError{
			Field:   "data",
			Message: "no trading data provided",
		}
	}
	
	if delta < 0 || delta > 1 {
		return &ValidationError{
			Field:   "delta",
			Message: "delta parameter must be between 0 and 1",
			Value:   delta,
		}
	}
	
	// Check if dates are sorted
	for i := 1; i < len(data); i++ {
		if data[i].Date.Before(data[i-1].Date) {
			return &ValidationError{
				Field:   "data",
				Message: "trading data must be sorted by date",
			}
		}
	}
	
	return nil
}