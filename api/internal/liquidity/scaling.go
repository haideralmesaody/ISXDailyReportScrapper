package liquidity

import (
	"math"
	"sort"
)

// RobustScale applies percentile-based cross-sectional scaling
// This method preserves relative differences better than MAD-based scaling
//
// Parameters:
//   - values: slice of values to scale
//   - invert: if true, inverts the scaling (for ILLIQ where higher values mean lower liquidity)
//   - useLog: if true, applies log transformation before scaling
//
// Returns: slice of scaled values (0-100 percentile scale)
func RobustScale(values []float64, invert, useLog bool) []float64 {
	n := len(values)
	if n == 0 {
		return values
	}
	
	// Handle single value case
	if n == 1 {
		return []float64{50.0} // Middle percentile for single value
	}
	
	// Create working copy and handle log transformation
	workingValues := make([]float64, 0, n)
	validIndices := make([]int, 0, n)
	
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		
		transformedVal := v
		if useLog && v > 0 {
			transformedVal = math.Log(v + 1) // +1 to handle zero
		}
		
		if !math.IsNaN(transformedVal) && !math.IsInf(transformedVal, 0) {
			workingValues = append(workingValues, transformedVal)
			validIndices = append(validIndices, i)
		}
	}
	
	if len(workingValues) == 0 {
		// All values were invalid
		result := make([]float64, n)
		for i := range result {
			result[i] = 0
		}
		return result
	}
	
	// Sort for percentile calculation
	sorted := make([]float64, len(workingValues))
	copy(sorted, workingValues)
	sort.Float64s(sorted)
	
	// Calculate percentiles - use wider range (5th and 95th) for better differentiation
	p5 := getPercentileValue(sorted, 0.05)
	p95 := getPercentileValue(sorted, 0.95)
	
	// Handle edge case where all values are similar
	if math.Abs(p95 - p5) < 0.0001 {
		result := make([]float64, n)
		for i := range result {
			result[i] = 50.0 // All get middle score
		}
		return result
	}
	
	// Scale values with improved preservation of ratios
	scaledValues := make([]float64, len(workingValues))
	iqr := p95 - p5
	
	for i, val := range workingValues {
		var scaled float64
		
		if val <= p5 {
			// Below 5th percentile: scale 0-5
			if val <= sorted[0] {
				scaled = 0
			} else {
				scaled = 5 * (val - sorted[0]) / (p5 - sorted[0])
			}
		} else if val >= p95 {
			// Above 95th percentile: scale 95-100
			if val >= sorted[len(sorted)-1] {
				scaled = 100
			} else {
				scaled = 95 + 5 * (val - p95) / (sorted[len(sorted)-1] - p95)
			}
		} else {
			// Within IQR: scale 5-95 with logarithmic adjustment to preserve ratios
			// This helps maintain relative differences better
			normalizedPos := (val - p5) / iqr
			
			// Apply mild logarithmic transformation to preserve ratios
			// This reduces compression of differences in the middle range
			if useLog && normalizedPos > 0 {
				// Soften the linear scaling with log component
				linearComponent := normalizedPos * 0.7
				logComponent := (math.Log(1 + normalizedPos*9) / math.Log(10)) * 0.3
				normalizedPos = linearComponent + logComponent
			}
			
			scaled = 5 + 90 * normalizedPos
		}
		
		// Apply inversion if needed
		if invert {
			scaled = 100 - scaled
		}
		
		scaledValues[i] = scaled
	}
	
	// Map back to original indices
	result := make([]float64, n)
	for i := range result {
		result[i] = 0 // Default for invalid values
	}
	
	for i, validIdx := range validIndices {
		result[validIdx] = scaledValues[i]
	}
	
	return result
}

// getPercentileValue calculates the value at a given percentile
func getPercentileValue(sorted []float64, percentile float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	
	if percentile <= 0 {
		return sorted[0]
	}
	if percentile >= 1 {
		return sorted[n-1]
	}
	
	index := percentile * float64(n-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sorted[lower]
	}
	
	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// calculateMedian computes the median of a slice of float64 values
func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Create sorted copy
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)
	
	n := len(sortedValues)
	if n%2 == 0 {
		// Even number of values
		return (sortedValues[n/2-1] + sortedValues[n/2]) / 2
	} else {
		// Odd number of values
		return sortedValues[n/2]
	}
}

// calculateMAD computes the Median Absolute Deviation
func calculateMAD(values []float64, median float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Calculate absolute deviations from median
	absDeviations := make([]float64, len(values))
	for i, v := range values {
		absDeviations[i] = math.Abs(v - median)
	}
	
	// Return median of absolute deviations
	// Scale by 1.4826 for consistency with normal distribution standard deviation
	return 1.4826 * calculateMedian(absDeviations)
}

// convertToPercentiles converts z-scores to percentile rankings (0-100)
func convertToPercentiles(zScores []float64, invert bool) []float64 {
	if len(zScores) == 0 {
		return zScores
	}
	
	// Create index-value pairs for sorting
	type indexValue struct {
		index int
		value float64
	}
	
	indexValues := make([]indexValue, len(zScores))
	for i, v := range zScores {
		indexValues[i] = indexValue{index: i, value: v}
	}
	
	// Sort by value
	sort.Slice(indexValues, func(i, j int) bool {
		if invert {
			return indexValues[i].value > indexValues[j].value // Descending for inversion
		}
		return indexValues[i].value < indexValues[j].value // Ascending for normal
	})
	
	// Assign percentile ranks
	percentiles := make([]float64, len(zScores))
	for rank, iv := range indexValues {
		// Convert rank to percentile (0-100)
		percentile := float64(rank) / float64(len(zScores)-1) * 100
		percentiles[iv.index] = percentile
	}
	
	return percentiles
}

// WinsorizeValues applies winsorization to extreme values
func WinsorizeValues(values []float64, lowerPercentile, upperPercentile float64) []float64 {
	if len(values) == 0 {
		return values
	}
	
	// Create sorted copy to find percentile bounds
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)
	
	// Calculate percentile indices
	lowerIdx := int(math.Floor(lowerPercentile * float64(len(sortedValues)-1)))
	upperIdx := int(math.Ceil(upperPercentile * float64(len(sortedValues)-1)))
	
	if lowerIdx < 0 {
		lowerIdx = 0
	}
	if upperIdx >= len(sortedValues) {
		upperIdx = len(sortedValues) - 1
	}
	
	lowerBound := sortedValues[lowerIdx]
	upperBound := sortedValues[upperIdx]
	
	// Apply winsorization
	winsorizedValues := make([]float64, len(values))
	for i, v := range values {
		if v < lowerBound {
			winsorizedValues[i] = lowerBound
		} else if v > upperBound {
			winsorizedValues[i] = upperBound
		} else {
			winsorizedValues[i] = v
		}
	}
	
	return winsorizedValues
}

// StandardScale applies standard z-score normalization
// This is provided as an alternative to robust scaling
func StandardScale(values []float64, invert, useLog bool) []float64 {
	if len(values) == 0 {
		return values
	}
	
	// Handle log transformation
	workingValues := make([]float64, 0, len(values))
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		
		if useLog {
			if v > 0 {
				workingValues = append(workingValues, math.Log(v))
			} else {
				workingValues = append(workingValues, math.Log(1e-12))
			}
		} else {
			workingValues = append(workingValues, v)
		}
	}
	
	if len(workingValues) == 0 {
		result := make([]float64, len(values))
		return result
	}
	
	// Calculate mean and standard deviation
	mean := calculateMean(workingValues)
	stdDev := calculateStandardDeviation(workingValues, mean)
	
	// Scale values
	scaledValues := make([]float64, len(workingValues))
	for i, v := range workingValues {
		if stdDev > 0 {
			scaledValues[i] = (v - mean) / stdDev
		} else {
			scaledValues[i] = 0
		}
	}
	
	// Convert to percentiles
	percentileValues := convertToPercentiles(scaledValues, invert)
	
	// Map back to original length
	result := make([]float64, len(values))
	scaledIdx := 0
	
	for i, originalValue := range values {
		if math.IsNaN(originalValue) || math.IsInf(originalValue, 0) {
			result[i] = 0
		} else {
			result[i] = percentileValues[scaledIdx]
			scaledIdx++
		}
	}
	
	return result
}

// calculateStandardDeviation computes sample standard deviation
func calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	
	sumSquaredDeviations := 0.0
	validCount := 0
	
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			deviation := v - mean
			sumSquaredDeviations += deviation * deviation
			validCount++
		}
	}
	
	if validCount <= 1 {
		return 0
	}
	
	return math.Sqrt(sumSquaredDeviations / float64(validCount-1))
}

// RankTransform applies rank-based transformation
// This is most robust to outliers and preserves only ordinal relationships
func RankTransform(values []float64, invert bool) []float64 {
	if len(values) == 0 {
		return values
	}
	
	// Create index-value pairs, filtering out invalid values
	type indexValue struct {
		originalIndex int
		value         float64
		isValid       bool
	}
	
	indexValues := make([]indexValue, len(values))
	validCount := 0
	
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			indexValues[i] = indexValue{originalIndex: i, isValid: false}
		} else {
			indexValues[i] = indexValue{originalIndex: i, value: v, isValid: true}
			validCount++
		}
	}
	
	if validCount == 0 {
		// All values invalid
		result := make([]float64, len(values))
		return result
	}
	
	// Sort valid values only
	validValues := make([]indexValue, 0, validCount)
	for _, iv := range indexValues {
		if iv.isValid {
			validValues = append(validValues, iv)
		}
	}
	
	sort.Slice(validValues, func(i, j int) bool {
		if invert {
			return validValues[i].value > validValues[j].value
		}
		return validValues[i].value < validValues[j].value
	})
	
	// Assign ranks (handle ties with average ranking)
	result := make([]float64, len(values))
	
	// First pass: assign ranks
	for rank, iv := range validValues {
		percentile := float64(rank) / float64(validCount-1) * 100
		result[iv.originalIndex] = percentile
	}
	
	// Handle ties by averaging ranks
	for i := 0; i < len(validValues); {
		currentValue := validValues[i].value
		tieStart := i
		
		// Find end of tie group
		for i < len(validValues) && validValues[i].value == currentValue {
			i++
		}
		tieEnd := i
		
		// If there's a tie, average the ranks
		if tieEnd-tieStart > 1 {
			sumRanks := 0.0
			for j := tieStart; j < tieEnd; j++ {
				sumRanks += float64(j) / float64(validCount-1) * 100
			}
			avgRank := sumRanks / float64(tieEnd-tieStart)
			
			for j := tieStart; j < tieEnd; j++ {
				result[validValues[j].originalIndex] = avgRank
			}
		}
	}
	
	return result
}

// CombineScaledComponents combines multiple scaled components with weights
func CombineScaledComponents(components [][]float64, weights []float64) ([]float64, error) {
	if len(components) == 0 || len(weights) == 0 {
		return nil, &ValidationError{
			Field:   "components",
			Message: "no components or weights provided",
		}
	}
	
	if len(components) != len(weights) {
		return nil, &ValidationError{
			Field:   "weights",
			Message: "number of weights must match number of components",
		}
	}
	
	// Check that all components have the same length
	n := len(components[0])
	for i, comp := range components {
		if len(comp) != n {
			return nil, &ValidationError{
				Field:   "components",
				Message: "all components must have the same length",
				Value:   map[string]int{"component": i, "length": len(comp), "expected": n},
			}
		}
	}
	
	// Normalize weights
	weightSum := 0.0
	for _, w := range weights {
		weightSum += w
	}
	if weightSum == 0 {
		return nil, &ValidationError{
			Field:   "weights",
			Message: "sum of weights must be positive",
		}
	}
	
	normalizedWeights := make([]float64, len(weights))
	for i, w := range weights {
		normalizedWeights[i] = w / weightSum
	}
	
	// Combine components
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		for j, comp := range components {
			result[i] += comp[i] * normalizedWeights[j]
		}
	}
	
	return result, nil
}

// ValidateScalingInputs validates inputs for scaling functions
func ValidateScalingInputs(values []float64, lowerBound, upperBound float64) error {
	if len(values) == 0 {
		return &ValidationError{
			Field:   "values",
			Message: "no values provided for scaling",
		}
	}
	
	if lowerBound < 0 || upperBound > 1 || lowerBound >= upperBound {
		return &ValidationError{
			Field:   "bounds",
			Message: "invalid bounds for scaling",
			Value:   map[string]float64{"lower": lowerBound, "upper": upperBound},
		}
	}
	
	// Check for minimum number of valid values
	validCount := 0
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			validCount++
		}
	}
	
	if validCount < 2 {
		return &ValidationError{
			Field:   "values",
			Message: "need at least 2 valid values for meaningful scaling",
			Value:   validCount,
		}
	}
	
	return nil
}