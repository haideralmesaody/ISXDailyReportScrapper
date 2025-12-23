package liquidity

import (
	"math"
)

// LinearScaleILLIQ applies piecewise linear scaling to ILLIQ values
// ILLIQ measures price impact - lower is better (more liquid)
// 
// Scaling ranges:
//   - 0 to 0.001: Score 100-90 (excellent liquidity)
//   - 0.001 to 0.01: Score 90-70 (good liquidity)  
//   - 0.01 to 0.1: Score 70-40 (moderate liquidity)
//   - 0.1 to 1: Score 40-10 (poor liquidity)
//   - Above 1: Score 10-0 (very poor liquidity)
func LinearScaleILLIQ(values []float64) []float64 {
	n := len(values)
	if n == 0 {
		return values
	}

	result := make([]float64, n)
	
	for i, value := range values {
		// Handle invalid values
		if math.IsNaN(value) || math.IsInf(value, 0) {
			result[i] = 0
			continue
		}
		
		// Clamp negative values to 0 (shouldn't happen but be safe)
		if value < 0 {
			value = 0
		}
		
		// Piecewise linear scaling
		var score float64
		switch {
		case value <= 0.001:
			// Excellent liquidity: 0 to 0.001 maps to 100 to 90
			score = 100 - (value/0.001)*10
		case value <= 0.01:
			// Good liquidity: 0.001 to 0.01 maps to 90 to 70
			score = 90 - ((value-0.001)/(0.01-0.001))*20
		case value <= 0.1:
			// Moderate liquidity: 0.01 to 0.1 maps to 70 to 40
			score = 70 - ((value-0.01)/(0.1-0.01))*30
		case value <= 1.0:
			// Poor liquidity: 0.1 to 1.0 maps to 40 to 10
			score = 40 - ((value-0.1)/(1.0-0.1))*30
		default:
			// Very poor liquidity: Above 1.0 maps to 10 to 0
			// Use logarithmic decay for extreme values
			if value >= 10 {
				score = 0
			} else {
				score = 10 * (1 - math.Log10(value))
			}
		}
		
		// Ensure score is within bounds
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		
		result[i] = score
	}
	
	return result
}

// LinearScaleVolume applies linear scaling to trading volume values
// Volume is in IQD (Iraqi Dinars) - higher is better
//
// Scaling ranges:
//   - Above 500M: Score 100
//   - 100M to 500M: Score 70-100
//   - 10M to 100M: Score 40-70
//   - 1M to 10M: Score 20-40
//   - Below 1M: Score 0-20
func LinearScaleVolume(values []float64) []float64 {
	n := len(values)
	if n == 0 {
		return values
	}

	result := make([]float64, n)
	
	for i, value := range values {
		// Handle invalid values
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			result[i] = 0
			continue
		}
		
		// Piecewise linear scaling based on IQD amounts
		var score float64
		switch {
		case value >= 500_000_000:
			// Excellent volume: 500M+ IQD
			score = 100
		case value >= 100_000_000:
			// Very good volume: 100M to 500M IQD
			score = 70 + ((value-100_000_000)/(500_000_000-100_000_000))*30
		case value >= 10_000_000:
			// Good volume: 10M to 100M IQD
			score = 40 + ((value-10_000_000)/(100_000_000-10_000_000))*30
		case value >= 1_000_000:
			// Moderate volume: 1M to 10M IQD
			score = 20 + ((value-1_000_000)/(10_000_000-1_000_000))*20
		default:
			// Low volume: Below 1M IQD
			score = (value / 1_000_000) * 20
		}
		
		// Ensure score is within bounds
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		
		result[i] = score
	}
	
	return result
}

// LinearScaleContinuity applies direct linear scaling to continuity values
// Continuity is already a percentage (0-1) - directly maps to score
//
// Scaling:
//   - 100% continuity = 100 score
//   - 50% continuity = 50 score
//   - 0% continuity = 0 score
func LinearScaleContinuity(values []float64) []float64 {
	n := len(values)
	if n == 0 {
		return values
	}

	result := make([]float64, n)
	
	for i, value := range values {
		// Handle invalid values
		if math.IsNaN(value) || math.IsInf(value, 0) {
			result[i] = 0
			continue
		}
		
		// Continuity is already normalized 0-1, just scale to 0-100
		score := value * 100
		
		// Ensure score is within bounds
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		
		result[i] = score
	}
	
	return result
}

// LinearScaleValues is a generic linear scaling function for cross-sectional normalization
// It scales values linearly between min and max to 0-100 range
func LinearScaleValues(values []float64, invert bool) []float64 {
	n := len(values)
	if n == 0 {
		return values
	}
	
	// Find valid min and max
	var validValues []float64
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			validValues = append(validValues, v)
		}
	}
	
	if len(validValues) == 0 {
		// All values invalid
		result := make([]float64, n)
		return result
	}
	
	// Find min and max
	min, max := validValues[0], validValues[0]
	for _, v := range validValues {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	// Handle case where all values are the same
	if max-min < 0.0001 {
		result := make([]float64, n)
		for i := range result {
			if !math.IsNaN(values[i]) && !math.IsInf(values[i], 0) {
				result[i] = 50 // Middle score for all equal values
			}
		}
		return result
	}
	
	// Scale values
	result := make([]float64, n)
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			result[i] = 0
			continue
		}
		
		// Linear scaling
		score := ((v - min) / (max - min)) * 100
		
		// Apply inversion if needed
		if invert {
			score = 100 - score
		}
		
		// Ensure bounds
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		
		result[i] = score
	}
	
	return result
}

// GetILLIQBounds returns the boundary values used for ILLIQ scaling
// Useful for debugging and documentation
func GetILLIQBounds() map[string]float64 {
	return map[string]float64{
		"excellent_max": 0.001,
		"good_max":      0.01,
		"moderate_max":  0.1,
		"poor_max":      1.0,
	}
}

// GetVolumeBounds returns the boundary values used for volume scaling
// Useful for debugging and documentation
func GetVolumeBounds() map[string]float64 {
	return map[string]float64{
		"excellent_min": 500_000_000,
		"very_good_min": 100_000_000,
		"good_min":      10_000_000,
		"moderate_min":  1_000_000,
	}
}

// CalculateLinearScore calculates a simple linear score between two points
func CalculateLinearScore(value, minVal, maxVal, minScore, maxScore float64) float64 {
	if value <= minVal {
		return minScore
	}
	if value >= maxVal {
		return maxScore
	}
	
	// Linear interpolation
	ratio := (value - minVal) / (maxVal - minVal)
	score := minScore + ratio*(maxScore-minScore)
	
	return score
}