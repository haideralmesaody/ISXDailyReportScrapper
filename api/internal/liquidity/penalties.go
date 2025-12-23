package liquidity

import (
	"fmt"
	"math"
)

// PiecewisePenalty calculates the piecewise linear penalty function
// as described in the ISX Hybrid Liquidity Metric paper
//
// Parameters:
//   - p0: inactivity ratio (proportion of non-trading days, 0.0 to 1.0)
//   - beta: slope parameter for p0 below pStar (mild penalty)
//   - gamma: slope parameter for p0 above pStar (steep penalty)
//   - pStar: transition point between mild and steep penalty regimes (typically 0.5)
//   - maxMult: maximum penalty multiplier to prevent extreme values
//
// Returns: penalty multiplier (>= 1.0, bounded by maxMult)
func PiecewisePenalty(p0, beta, gamma, pStar, maxMult float64) float64 {
	// Validate inputs - p0 should be an inactivity ratio between 0 and 1
	if p0 < 0 || p0 > 1 {
		return 1.0 // Invalid inactivity ratio, no penalty
	}
	
	if pStar <= 0 || pStar >= 1 || maxMult <= 1 {
		return 1.0 // Invalid parameters
	}
	
	var penalty float64
	
	if p0 <= pStar {
		// Low inactivity regime: penalty INCREASES with inactivity
		// At p0=0 (perfect continuity): penalty = 1.0 (no penalty)
		// At p0=pStar: penalty = 1.0 + beta
		penalty = 1.0 + beta * (p0 / pStar)
	} else {
		// High inactivity regime: steeper penalty for poor continuity
		// Continuous at pStar: starts at 1.0 + beta
		// Increases to maxMult as p0 approaches 1.0
		penalty = 1.0 + beta + gamma * ((p0 - pStar) / (1.0 - pStar))
	}
	
	// Ensure penalty is at least 1.0 (no negative penalties)
	if penalty < 1.0 {
		penalty = 1.0
	}
	
	// Cap penalty at maximum multiplier
	if penalty > maxMult {
		penalty = maxMult
	}
	
	return penalty
}

// ExponentialPenalty calculates the exponential penalty function
// for volume-based adjustments in the ISX Hybrid Liquidity Metric
//
// Parameters:
//   - p0: inactivity ratio (proportion of non-trading days, 0.0 to 1.0)
//   - alpha: exponential growth rate for penalty
//   - maxMult: maximum penalty multiplier to prevent extreme values
//
// Returns: penalty multiplier (>= 1.0, bounded by maxMult)
func ExponentialPenalty(p0, alpha, maxMult float64) float64 {
	if p0 <= 0 || alpha <= 0 || maxMult <= 1 {
		return 1.0 // No penalty for invalid inputs
	}
	
	// Exponential penalty based on inactivity ratio
	// Higher inactivity gets exponentially higher penalties
	penalty := math.Exp(alpha * p0)
	
	// Normalize to ensure penalty >= 1.0
	if penalty < 1.0 {
		penalty = 1.0 / penalty // Invert if less than 1
	}
	
	// Cap penalty at maximum multiplier
	if penalty > maxMult {
		penalty = maxMult
	}
	
	// Handle NaN/Inf cases
	if math.IsNaN(penalty) || math.IsInf(penalty, 0) {
		return 1.0
	}
	
	return penalty
}

// DefaultPenaltyParams returns the default penalty parameters
// calibrated for the Iraqi Stock Exchange based on empirical analysis
func DefaultPenaltyParams() PenaltyParams {
	return PenaltyParams{
		// Piecewise penalty parameters
		// These values are calibrated based on ISX price distributions
		PiecewiseP0:       1.0,   // Reference price level (normalized)
		PiecewiseBeta:     0.3,   // Low-price penalty slope (30% penalty per unit)
		PiecewiseGamma:    0.1,   // High-price penalty slope (10% penalty per unit)
		PiecewisePStar:    2.0,   // Transition price (2 IQD as reference)
		PiecewiseMaxMult:  3.0,   // Maximum 300% penalty
		
		// Exponential penalty parameters
		// These control volume adjustments based on price levels
		ExponentialP0:     1.0,   // Reference price level
		ExponentialAlpha:  0.2,   // Moderate exponential decay rate
		ExponentialMaxMult: 2.5,  // Maximum 250% penalty for volume
	}
}

// CalibratedPenaltyParams returns penalty parameters calibrated for specific market conditions
// These can be adjusted based on empirical analysis of ISX data
func CalibratedPenaltyParams(marketRegime string) PenaltyParams {
	switch marketRegime {
	case "low_volatility":
		// Lower penalties for stable market conditions
		return PenaltyParams{
			PiecewiseP0:        1.0,
			PiecewiseBeta:      0.2,  // Reduced low-price penalty
			PiecewiseGamma:     0.05, // Reduced high-price penalty
			PiecewisePStar:     2.0,
			PiecewiseMaxMult:   2.0,  // Lower maximum penalty
			ExponentialP0:      1.0,
			ExponentialAlpha:   0.15, // Gentler exponential decay
			ExponentialMaxMult: 2.0,
		}
		
	case "high_volatility":
		// Higher penalties for volatile market conditions
		return PenaltyParams{
			PiecewiseP0:        1.0,
			PiecewiseBeta:      0.5,  // Increased low-price penalty
			PiecewiseGamma:     0.2,  // Increased high-price penalty
			PiecewisePStar:     2.0,
			PiecewiseMaxMult:   4.0,  // Higher maximum penalty
			ExponentialP0:      1.0,
			ExponentialAlpha:   0.3,  // Stronger exponential decay
			ExponentialMaxMult: 3.0,
		}
		
	case "small_cap_focused":
		// Adjusted penalties for small-cap dominated analysis
		return PenaltyParams{
			PiecewiseP0:        1.0,
			PiecewiseBeta:      0.4,  // Higher penalty for very low prices
			PiecewiseGamma:     0.15, // Moderate penalty for higher prices
			PiecewisePStar:     1.5,  // Lower transition point
			PiecewiseMaxMult:   3.5,
			ExponentialP0:      1.0,
			ExponentialAlpha:   0.25,
			ExponentialMaxMult: 2.8,
		}
		
	default:
		// Return default parameters for unknown regimes
		return DefaultPenaltyParams()
	}
}

// ValidatePenaltyParams performs comprehensive validation of penalty parameters
func ValidatePenaltyParams(params PenaltyParams) error {
	// Check piecewise parameters
	if params.PiecewiseP0 <= 0 {
		return &ValidationError{
			Field:   "PiecewiseP0",
			Message: "piecewise P0 must be positive",
			Value:   params.PiecewiseP0,
		}
	}
	
	if params.PiecewiseBeta <= 0 {
		return &ValidationError{
			Field:   "PiecewiseBeta",
			Message: "piecewise beta must be positive",
			Value:   params.PiecewiseBeta,
		}
	}
	
	if params.PiecewiseGamma <= 0 {
		return &ValidationError{
			Field:   "PiecewiseGamma",
			Message: "piecewise gamma must be positive",
			Value:   params.PiecewiseGamma,
		}
	}
	
	if params.PiecewisePStar <= 0 {
		return &ValidationError{
			Field:   "PiecewisePStar",
			Message: "piecewise pStar must be positive",
			Value:   params.PiecewisePStar,
		}
	}
	
	if params.PiecewiseMaxMult <= 1 {
		return &ValidationError{
			Field:   "PiecewiseMaxMult",
			Message: "piecewise maximum multiplier must be greater than 1",
			Value:   params.PiecewiseMaxMult,
		}
	}
	
	// Check exponential parameters
	if params.ExponentialP0 <= 0 {
		return &ValidationError{
			Field:   "ExponentialP0",
			Message: "exponential P0 must be positive",
			Value:   params.ExponentialP0,
		}
	}
	
	if params.ExponentialAlpha <= 0 {
		return &ValidationError{
			Field:   "ExponentialAlpha",
			Message: "exponential alpha must be positive",
			Value:   params.ExponentialAlpha,
		}
	}
	
	if params.ExponentialMaxMult <= 1 {
		return &ValidationError{
			Field:   "ExponentialMaxMult",
			Message: "exponential maximum multiplier must be greater than 1",
			Value:   params.ExponentialMaxMult,
		}
	}
	
	// Check for reasonable ranges
	if params.PiecewiseBeta > 2.0 {
		return &ValidationError{
			Field:   "PiecewiseBeta",
			Message: "piecewise beta seems too large (> 2.0), may cause excessive penalties",
			Value:   params.PiecewiseBeta,
		}
	}
	
	if params.PiecewiseGamma > 1.0 {
		return &ValidationError{
			Field:   "PiecewiseGamma", 
			Message: "piecewise gamma seems too large (> 1.0), may cause excessive penalties",
			Value:   params.PiecewiseGamma,
		}
	}
	
	if params.ExponentialAlpha > 1.0 {
		return &ValidationError{
			Field:   "ExponentialAlpha",
			Message: "exponential alpha seems too large (> 1.0), may cause excessive penalties",
			Value:   params.ExponentialAlpha,
		}
	}
	
	if params.PiecewiseMaxMult > 10.0 || params.ExponentialMaxMult > 10.0 {
		return &ValidationError{
			Field:   "MaxMultipliers",
			Message: "maximum multipliers seem too large (> 10.0), may cause unstable results",
		}
	}
	
	return nil
}

// TestPenaltyFunctions performs unit tests on penalty functions with various inputs
// This is useful for validating penalty function behavior during calibration
func TestPenaltyFunctions(params PenaltyParams) map[string]float64 {
	testResults := make(map[string]float64)
	
	// Test piecewise penalty with various price levels
	testPrices := []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0}
	
	for _, price := range testPrices {
		piecewise := PiecewisePenalty(price, params.PiecewiseBeta, params.PiecewiseGamma, 
			params.PiecewisePStar, params.PiecewiseMaxMult)
		exponential := ExponentialPenalty(price, params.ExponentialAlpha, params.ExponentialMaxMult)
		
		testResults[fmt.Sprintf("piecewise_%.1f", price)] = piecewise
		testResults[fmt.Sprintf("exponential_%.1f", price)] = exponential
	}
	
	// Test edge cases
	testResults["piecewise_zero"] = PiecewisePenalty(0, params.PiecewiseBeta, params.PiecewiseGamma, 
		params.PiecewisePStar, params.PiecewiseMaxMult)
	testResults["exponential_zero"] = ExponentialPenalty(0, params.ExponentialAlpha, params.ExponentialMaxMult)
	
	testResults["piecewise_negative"] = PiecewisePenalty(-1, params.PiecewiseBeta, params.PiecewiseGamma, 
		params.PiecewisePStar, params.PiecewiseMaxMult)
	testResults["exponential_negative"] = ExponentialPenalty(-1, params.ExponentialAlpha, params.ExponentialMaxMult)
	
	return testResults
}

// ActivityScore calculates a unified activity-based score (0-1) for liquidity adjustments
// This replaces the dual penalty system with a single, more efficient calculation
//
// The score represents how active a stock is:
//   - 1.0 = perfectly active (trades every day)
//   - 0.0 = completely inactive (never trades)
//
// Parameters:
//   - tradingDays: number of days the stock traded
//   - totalDays: total days in the window (e.g., 60)
//   - config: optional configuration for fine-tuning
//
// Returns: activity score between 0 and 1
func ActivityScore(tradingDays, totalDays int) float64 {
	if totalDays <= 0 {
		return 0 // Invalid window
	}
	
	if tradingDays <= 0 {
		return 0 // No trading activity
	}
	
	if tradingDays >= totalDays {
		return 1.0 // Perfect continuity
	}
	
	// Calculate raw continuity ratio
	continuity := float64(tradingDays) / float64(totalDays)
	
	// Apply non-linear transformation for better sensitivity
	// Square root gives more differentiation in the lower range
	// where most ISX stocks operate (10-50% continuity)
	activityScore := math.Sqrt(continuity)
	
	// Additional boost for moderate activity (>30% trading days)
	// This helps differentiate between truly inactive and moderately active stocks
	if continuity > 0.3 {
		// Smooth transition: add up to 10% bonus for stocks trading 30-70% of days
		bonus := 0.1 * math.Min(1.0, (continuity - 0.3) / 0.4)
		activityScore = math.Min(1.0, activityScore + bonus)
	}
	
	// Severe penalty for very low activity (<10% trading days)
	if continuity < 0.1 {
		// Apply exponential decay for very inactive stocks
		activityScore *= math.Exp(-2.0 * (0.1 - continuity))
	}
	
	// Ensure bounds
	if activityScore < 0 {
		activityScore = 0
	}
	if activityScore > 1 {
		activityScore = 1
	}
	
	return activityScore
}

// UnifiedPenalty calculates a single penalty multiplier based on activity score
// This replaces both PiecewisePenalty and ExponentialPenalty for simpler computation
//
// Parameters:
//   - activityScore: output from ActivityScore function (0-1)
//   - maxPenalty: maximum penalty multiplier (e.g., 3.0 for 300% penalty)
//
// Returns: penalty multiplier (>= 1.0)
func UnifiedPenalty(activityScore, maxPenalty float64) float64 {
	if activityScore >= 1.0 {
		return 1.0 // No penalty for perfect activity
	}
	
	if activityScore <= 0 || maxPenalty <= 1.0 {
		return maxPenalty // Maximum penalty for no activity
	}
	
	// Inverse relationship: lower activity = higher penalty
	// Using exponential curve for smooth transition
	penaltyRange := maxPenalty - 1.0
	penalty := 1.0 + penaltyRange * math.Exp(-3.0 * activityScore)
	
	// Ensure bounds
	if penalty < 1.0 {
		penalty = 1.0
	}
	if penalty > maxPenalty {
		penalty = maxPenalty
	}
	
	return penalty
}