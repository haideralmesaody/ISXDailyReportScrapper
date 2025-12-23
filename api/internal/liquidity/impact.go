package liquidity

import (
	"math"
	"sort"
)

// ComputeILLIQ calculates the Amihud (2002) illiquidity measure with adjustments for ISX
// This is the core price impact component of the ISX Hybrid Liquidity Metric
//
// The ILLIQ measure is calculated as the average ratio of absolute daily return
// to daily trading value, representing the price impact per unit of trading value
//
// Parameters:
//   - data: slice of TradingDay data for the calculation window
//   - kLower: lower percentile for winsorization (e.g., 0.05 for 5th percentile)
//   - kUpper: upper percentile for winsorization (e.g., 0.95 for 95th percentile)
//
// Returns:
//   - illiq: the calculated ILLIQ measure (always positive, higher = more illiquid)
//   - lowerBound: the lower winsorization bound applied
//   - upperBound: the upper winsorization bound applied
func ComputeILLIQ(data []TradingDay, kLower, kUpper float64) (illiq, lowerBound, upperBound float64) {
	return ComputeILLIQWithGapPenalty(data, kLower, kUpper, true, nil)
}

// ComputeILLIQWithGapPenalty calculates ILLIQ with optional gap-based penalties
// This allows for more nuanced liquidity measurement considering trading continuity
//
// Parameters:
//   - data: slice of TradingDay data for the calculation window
//   - kLower: lower percentile for winsorization
//   - kUpper: upper percentile for winsorization
//   - applyGapPenalty: whether to apply gap-based penalties
//   - gapConfig: custom gap penalty configuration (uses defaults if nil)
//
// Returns:
//   - illiq: the calculated ILLIQ measure with gap penalties applied
//   - lowerBound: the lower winsorization bound applied
//   - upperBound: the upper winsorization bound applied
func ComputeILLIQWithGapPenalty(data []TradingDay, kLower, kUpper float64,
	applyGapPenalty bool, gapConfig *GapPenaltyConfig) (illiq, lowerBound, upperBound float64) {
	if len(data) < 2 {
		// Insufficient data - return worst case illiquidity
		return 1000.0, 0, 0
	}
	
	// STEP 1: Calculate average price for normalization
	totalPrice := 0.0
	priceCount := 0
	for _, d := range data {
		if d.IsTrading() && d.Close > 0 {
			totalPrice += d.Close
			priceCount++
		}
	}
	
	if priceCount == 0 {
		// No valid prices - worst case
		return 1000.0, 0, 0
	}
	
	avgPrice := totalPrice / float64(priceCount)
	isPennyStock := avgPrice < 0.5 // Below 0.5 IQD is considered penny stock
	
	// STEP 2: Calculate daily ILLIQ with adjustments
	var dailyILLIQ []float64
	var lowValueDays int
	const minMeaningfulValue = 1_000_000.0 // 1M IQD minimum for meaningful liquidity
	
	for i := 1; i < len(data); i++ {
		prev := data[i-1]
		curr := data[i]
		
		// Skip if current day is non-trading
		if !curr.IsTrading() {
			continue
		}
		
		// Count low-value days
		if curr.Value < minMeaningfulValue {
			lowValueDays++
		}
		
		// Skip if previous day wasn't trading
		if !prev.IsTrading() {
			continue
		}
		
		// Calculate return
		absReturn := math.Abs(curr.Return(prev.Close))
		
		// CRITICAL FIX: Handle zero returns properly
		if absReturn == 0 {
			// Zero return doesn't mean perfect liquidity
			if isPennyStock {
				// For penny stocks, assume minimum spread impact
				absReturn = 0.01 // 1% minimum for penny stocks
			} else {
				// For normal stocks, smaller minimum
				absReturn = 0.001 // 0.1% minimum
			}
		}
		
		// Calculate value in millions with tiny epsilon for division safety
		const epsilon = 1e-6 // Tiny guard against division by zero
		valueMillions := math.Max(curr.Value / 1_000_000, epsilon)

		// Calculate ILLIQ ratio
		illiqRatio := absReturn / valueMillions

		// Cap the output instead of flooring the input (preserves accuracy)
		if illiqRatio > 1000 {
			illiqRatio = 1000 // Cap extreme values at 1000 (max penalty)
		}
		
		// PRICE NORMALIZATION: Adjust for penny stocks
		if isPennyStock {
			// Scale up penny stock ILLIQ to reflect true impact
			priceAdjustment := math.Sqrt(0.5 / avgPrice)
			illiqRatio *= priceAdjustment
		}
		
		if !math.IsNaN(illiqRatio) && !math.IsInf(illiqRatio, 0) {
			dailyILLIQ = append(dailyILLIQ, illiqRatio)
		}
	}
	
	if len(dailyILLIQ) == 0 {
		// No valid data - worst case
		return 1000.0, 0, 0
	}
	
	// STEP 3: Apply quality penalty for low-value trading
	lowValueRatio := float64(lowValueDays) / float64(len(data))
	qualityMultiplier := 1.0 + lowValueRatio * 2.0 // Up to 3x penalty
	
	// STEP 4: Calculate mean ILLIQ (simple average, no log transform)
	sum := 0.0
	for _, val := range dailyILLIQ {
		sum += val
	}
	illiq = (sum / float64(len(dailyILLIQ))) * qualityMultiplier
	
	// STEP 5: Apply gap-based penalty if requested
	if applyGapPenalty {
		if gapConfig == nil {
			defaultConfig := DefaultGapPenaltyConfig()
			gapConfig = &defaultConfig
		}
		gapPenalty := CalculateGapPenalty(data, *gapConfig)
		illiq = illiq * gapPenalty
	}
	
	// STEP 6: Ensure positive ILLIQ (no negative values)
	illiq = math.Max(illiq, 0.0001)
	
	return illiq, 0, 0
}

// logWinsorize applies log-transformation followed by winsorization
// This handles the extreme right-skewness typical in ILLIQ distributions
func logWinsorize(values []float64, kLower, kUpper float64) ([]float64, float64, float64) {
	if len(values) == 0 {
		return values, 0, 0
	}
	
	// Transform to log space (add small constant to handle zeros)
	const epsilon = 1e-12
	logValues := make([]float64, len(values))
	for i, v := range values {
		if v > 0 {
			logValues[i] = math.Log(v + epsilon)
		} else {
			logValues[i] = math.Log(epsilon)
		}
	}
	
	// Sort for percentile calculation
	sortedLogValues := make([]float64, len(logValues))
	copy(sortedLogValues, logValues)
	sort.Float64s(sortedLogValues)
	
	// Calculate percentile bounds
	lowerIdx := int(math.Floor(kLower * float64(len(sortedLogValues)-1)))
	upperIdx := int(math.Ceil(kUpper * float64(len(sortedLogValues)-1)))
	
	if lowerIdx < 0 {
		lowerIdx = 0
	}
	if upperIdx >= len(sortedLogValues) {
		upperIdx = len(sortedLogValues) - 1
	}
	
	lowerBound := sortedLogValues[lowerIdx]
	upperBound := sortedLogValues[upperIdx]
	
	// Apply winsorization
	winsorizedValues := make([]float64, len(logValues))
	for i, logVal := range logValues {
		if logVal < lowerBound {
			winsorizedValues[i] = lowerBound
		} else if logVal > upperBound {
			winsorizedValues[i] = upperBound
		} else {
			winsorizedValues[i] = logVal
		}
	}
	
	return winsorizedValues, math.Exp(lowerBound), math.Exp(upperBound)
}

// calculateMean computes the arithmetic mean of a slice of float64 values
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	validCount := 0
	
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			sum += v
			validCount++
		}
	}
	
	if validCount == 0 {
		return 0
	}
	
	return sum / float64(validCount)
}

// CalculateRealizedVolatility computes realized volatility from high-frequency returns
// This can be used as an alternative price impact measure
func CalculateRealizedVolatility(data []TradingDay) float64 {
	if len(data) < 2 {
		return 0
	}
	
	var squaredReturns []float64
	
	for i := 1; i < len(data); i++ {
		prev := data[i-1]
		curr := data[i]
		
		if !prev.IsTrading() || !curr.IsTrading() {
			continue
		}
		
		ret := curr.Return(prev.Close)
		if !math.IsNaN(ret) && !math.IsInf(ret, 0) {
			squaredReturns = append(squaredReturns, ret*ret)
		}
	}
	
	if len(squaredReturns) == 0 {
		return 0
	}
	
	// Sum of squared returns approximates realized variance
	realizedVariance := calculateMean(squaredReturns)
	return math.Sqrt(realizedVariance)
}

// CalculateIntraday ILLIQ calculates ILLIQ using intraday price ranges
// This provides a more granular measure when intraday data is limited
func CalculateIntradayILLIQ(data []TradingDay, kLower, kUpper float64) (float64, float64, float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}
	
	var intradayILLIQ []float64
	
	for _, td := range data {
		if !td.IsTrading() || td.Value == 0 {
			continue
		}
		
		// Use high-low range as proxy for intraday price movement
		priceRange := (td.High - td.Low) / td.Close // Normalize by closing price
		if math.IsNaN(priceRange) || math.IsInf(priceRange, 0) || priceRange < 0 {
			continue
		}
		
		// Use Value (IQD) instead of Volume (shares)
		valueMillions := td.Value / 1_000_000 // Convert IQD to millions
		if valueMillions > 0 {
			illiqRatio := priceRange / valueMillions
			if !math.IsNaN(illiqRatio) && !math.IsInf(illiqRatio, 0) && illiqRatio >= 0 {
				intradayILLIQ = append(intradayILLIQ, illiqRatio)
			}
		}
	}
	
	if len(intradayILLIQ) == 0 {
		return 0, 0, 0
	}
	
	// Apply log-winsorization
	logWinsorizedILLIQ, lowerBound, upperBound := logWinsorize(intradayILLIQ, kLower, kUpper)
	
	// Calculate average and transform back
	illiq := calculateMean(logWinsorizedILLIQ)
	if illiq > 0 {
		illiq = math.Exp(illiq)
	}
	
	return illiq, lowerBound, upperBound
}

// CalculateRollMeasure calculates the Roll (1984) effective spread measure
// This provides an alternative liquidity measure based on bid-ask bounce
func CalculateRollMeasure(data []TradingDay) float64 {
	if len(data) < 3 {
		return 0
	}
	
	// Calculate first-order return autocovariance
	var returns []float64
	for i := 1; i < len(data); i++ {
		prev := data[i-1]
		curr := data[i]
		
		if prev.IsTrading() && curr.IsTrading() {
			ret := curr.Return(prev.Close)
			if !math.IsNaN(ret) && !math.IsInf(ret, 0) {
				returns = append(returns, ret)
			}
		}
	}
	
	if len(returns) < 3 {
		return 0
	}
	
	// Calculate autocovariance at lag 1
	autoCovariance := calculateAutoCovariance(returns, 1)
	
	// Roll measure: 2 * sqrt(-autocovariance) if autocovariance is negative
	if autoCovariance < 0 {
		rollMeasure := 2 * math.Sqrt(-autoCovariance)
		if !math.IsNaN(rollMeasure) && !math.IsInf(rollMeasure, 0) {
			return rollMeasure
		}
	}
	
	return 0
}

// calculateAutoCovariance computes the autocovariance of returns at a given lag
func calculateAutoCovariance(returns []float64, lag int) float64 {
	if len(returns) <= lag {
		return 0
	}
	
	mean := calculateMean(returns)
	
	var covariance float64
	validCount := 0
	
	for i := lag; i < len(returns); i++ {
		if !math.IsNaN(returns[i]) && !math.IsNaN(returns[i-lag]) {
			covariance += (returns[i] - mean) * (returns[i-lag] - mean)
			validCount++
		}
	}
	
	if validCount == 0 {
		return 0
	}
	
	return covariance / float64(validCount)
}

// CalculateGapPenalty computes a penalty multiplier based on trading gaps
// Longer consecutive non-trading periods result in higher penalties
//
// Parameters:
//   - data: slice of TradingDay data
//   - config: GapPenaltyConfig for customization
//
// Returns: penalty multiplier (always >= 1.0)
func CalculateGapPenalty(data []TradingDay, config GapPenaltyConfig) float64 {
	gaps := identifyDetailedGaps(data)
	if len(gaps) == 0 {
		return 1.0 // No penalty if no gaps
	}

	// Apply gap forgiveness if configured
	gaps = applyGapForgiveness(gaps, config)
	if len(gaps) == 0 {
		return 1.0 // All gaps were forgiven
	}

	// Calculate weighted penalty based on gap characteristics
	totalPenalty := 1.0

	for _, gap := range gaps {
		// Apply different penalty curves based on gap length
		var gapMultiplier float64

		if gap.Length <= config.ShortGapThreshold {
			// Mild penalty for short gaps (1-2 days)
			gapMultiplier = 1.0 + (float64(gap.Length) * config.ShortGapPenaltyRate)
		} else if gap.Length <= config.MediumGapThreshold {
			// Moderate penalty for medium gaps (3-7 days)
			gapMultiplier = 1.0 + float64(config.ShortGapThreshold)*config.ShortGapPenaltyRate +
				(float64(gap.Length-config.ShortGapThreshold) * config.MediumGapPenaltyRate)
		} else {
			// Severe penalty for long gaps (>7 days)
			gapMultiplier = 1.0 + float64(config.ShortGapThreshold)*config.ShortGapPenaltyRate +
				float64(config.MediumGapThreshold-config.ShortGapThreshold)*config.MediumGapPenaltyRate +
				(float64(gap.Length-config.MediumGapThreshold) * config.LongGapPenaltyRate)
		}

		// Compound the penalties (multiplicative effect)
		totalPenalty *= gapMultiplier
	}

	// Apply additional penalties based on gap patterns
	if config.EnableFrequencyPenalty {
		totalPenalty *= calculateGapFrequencyPenalty(gaps, len(data))
	}
	if config.EnableClusteringPenalty {
		totalPenalty *= calculateGapClusteringPenalty(data)
	}

	// Cap the maximum penalty
	return math.Min(totalPenalty, config.MaxPenalty)
}

// applyGapForgiveness removes gaps that should be forgiven based on config
// This allows legitimate market closures (e.g., meetings, holidays) without penalty
func applyGapForgiveness(gaps []GapInfo, config GapPenaltyConfig) []GapInfo {
	// If no forgiveness configured, return all gaps
	if config.AllowedGapLength <= 0 || config.AllowedGapCount <= 0 {
		return gaps
	}

	// Sort gaps by length (shortest first) to forgive the smallest eligible gaps
	// This is more lenient than forgiving the largest gaps
	sortedGaps := make([]GapInfo, len(gaps))
	copy(sortedGaps, gaps)
	
	// Create a map to track which gaps to forgive
	forgiveMap := make(map[int]bool)
	forgivenCount := 0
	
	// Find gaps that qualify for forgiveness (length <= AllowedGapLength)
	for i, gap := range sortedGaps {
		if gap.Length <= config.AllowedGapLength && forgivenCount < config.AllowedGapCount {
			forgiveMap[i] = true
			forgivenCount++
		}
	}
	
	// Build the result list excluding forgiven gaps
	var result []GapInfo
	for _, gap := range gaps {
		shouldForgive := false
		// Check if this gap matches any forgiven gap
		for j, sortedGap := range sortedGaps {
			if forgiveMap[j] && gap.StartIndex == sortedGap.StartIndex && gap.Length == sortedGap.Length {
				shouldForgive = true
				break
			}
		}
		if !shouldForgive {
			result = append(result, gap)
		}
	}
	
	return result
}

// identifyDetailedGaps returns gap information with positions
func identifyDetailedGaps(data []TradingDay) []GapInfo {
	var gaps []GapInfo
	currentGap := GapInfo{StartIndex: -1}

	for i, td := range data {
		if !td.IsTrading() {
			if currentGap.StartIndex == -1 {
				currentGap.StartIndex = i
				currentGap.StartDate = td.Date
			}
			currentGap.Length++
		} else {
			if currentGap.Length > 0 {
				currentGap.EndIndex = i - 1
				currentGap.EndDate = data[i-1].Date
				gaps = append(gaps, currentGap)
				currentGap = GapInfo{StartIndex: -1}
			}
		}
	}

	// Handle gap at end
	if currentGap.Length > 0 {
		currentGap.EndIndex = len(data) - 1
		currentGap.EndDate = data[len(data)-1].Date
		gaps = append(gaps, currentGap)
	}

	return gaps
}

// calculateGapFrequencyPenalty penalizes frequent gaps
func calculateGapFrequencyPenalty(gaps []GapInfo, totalDays int) float64 {
	if len(gaps) == 0 {
		return 1.0
	}

	// Penalty based on number of gaps relative to total period
	gapFrequency := float64(len(gaps)) / float64(totalDays)

	// More gaps = worse liquidity
	// Use sqrt to avoid over-penalizing
	return 1.0 + math.Sqrt(gapFrequency)*0.5
}

// calculateGapClusteringPenalty penalizes clustered gaps
func calculateGapClusteringPenalty(data []TradingDay) float64 {
	// Identify if gaps are clustered together
	// Clustered gaps are worse than evenly distributed gaps

	var gapPositions []int
	for i, td := range data {
		if !td.IsTrading() {
			gapPositions = append(gapPositions, i)
		}
	}

	if len(gapPositions) < 2 {
		return 1.0
	}

	// Calculate variance in gap positions
	mean := float64(len(data)) / 2.0
	variance := 0.0
	for _, pos := range gapPositions {
		diff := float64(pos) - mean
		variance += diff * diff
	}
	variance /= float64(len(gapPositions))

	// Low variance = clustered gaps = higher penalty
	maxVariance := float64(len(data)*len(data)) / 12.0 // Theoretical max
	clusteringScore := 1.0 - (variance / maxVariance)

	return 1.0 + clusteringScore*0.3 // Up to 30% additional penalty
}

// ValidateILLIQInputs validates inputs for ILLIQ calculation
func ValidateILLIQInputs(data []TradingDay, kLower, kUpper float64) error {
	if len(data) < 2 {
		return &ValidationError{
			Field:   "data",
			Message: "insufficient data for ILLIQ calculation",
			Value:   len(data),
		}
	}
	
	if kLower < 0 || kLower >= kUpper || kUpper > 1 {
		return &ValidationError{
			Field:   "winsorization_bounds",
			Message: "invalid winsorization bounds",
			Value:   map[string]float64{"lower": kLower, "upper": kUpper},
		}
	}
	
	// Check data quality
	tradingDays := 0
	for _, td := range data {
		if td.IsTrading() {
			tradingDays++
		}
	}
	
	if tradingDays < MinTradingDaysForCalc {
		return &ValidationError{
			Field:   "trading_days",
			Message: "insufficient trading days for reliable ILLIQ calculation",
			Value:   tradingDays,
		}
	}
	
	return nil
}