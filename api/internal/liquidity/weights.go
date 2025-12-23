package liquidity

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"
)

// FitWeights estimates optimal component weights using cross-validation
// This finds the best combination of impact, value, and continuity weights
// to maximize correlation with the spread proxy
//
// Parameters:
//   - impactScores: scaled price impact scores
//   - valueScores: scaled value scores  
//   - continuityScores: scaled continuity scores
//   - spreads: spread proxy values for correlation target
//   - config: calibration configuration
//
// Returns: optimized component weights
func FitWeights(ctx context.Context, impactScores, valueScores, continuityScores, spreads []float64, config CalibrationConfig) (ComponentWeights, error) {
	logger := slog.Default()
	
	logger.InfoContext(ctx, "fitting component weights",
		"num_observations", len(impactScores),
		"k_folds", config.KFolds,
		"target_metric", config.TargetMetric,
	)
	
	// Validate inputs
	if err := validateWeightFittingInputs(impactScores, valueScores, continuityScores, spreads); err != nil {
		return ComponentWeights{}, fmt.Errorf("validate inputs: %w", err)
	}
	
	if len(impactScores) < config.MinTickers {
		return DefaultWeights(), fmt.Errorf("insufficient observations for weight fitting: %d < %d", len(impactScores), config.MinTickers)
	}
	
	// Set up random seed for reproducible results
	if config.RandomSeed != 0 {
		rand.Seed(config.RandomSeed)
	} else {
		rand.Seed(time.Now().UnixNano())
	}
	
	// Generate weight combinations to test
	weightCombinations := generateWeightCombinations(config.ParamGridSize)
	logger.InfoContext(ctx, "generated weight combinations", "count", len(weightCombinations))
	
	// Perform k-fold cross-validation
	bestWeights := ComponentWeights{}
	bestScore := -math.Inf(1)
	
	for i, weights := range weightCombinations {
		select {
		case <-ctx.Done():
			return ComponentWeights{}, fmt.Errorf("context cancelled during weight fitting: %w", ctx.Err())
		default:
		}
		
		score, err := evaluateWeightsCV(impactScores, valueScores, continuityScores, spreads, weights, config)
		if err != nil {
			logger.WarnContext(ctx, "failed to evaluate weights",
				"combination", i,
				"weights", weights,
				"error", err,
			)
			continue
		}
		
		if score > bestScore {
			bestScore = score
			bestWeights = weights
			
			logger.DebugContext(ctx, "found better weights",
				"score", score,
				"weights", weights,
			)
		}
	}
	
	if bestScore == -math.Inf(1) {
		logger.WarnContext(ctx, "no valid weight combinations found, using defaults")
		return DefaultWeights(), nil
	}
	
	logger.InfoContext(ctx, "weight fitting completed",
		"best_score", bestScore,
		"best_weights", bestWeights,
	)
	
	return bestWeights, nil
}

// evaluateWeightsCV evaluates a weight combination using k-fold cross-validation
func evaluateWeightsCV(impactScores, valueScores, continuityScores, spreads []float64, weights ComponentWeights, config CalibrationConfig) (float64, error) {
	n := len(impactScores)
	foldSize := n / config.KFolds
	
	var cvScores []float64
	
	// Perform k-fold cross-validation
	for fold := 0; fold < config.KFolds; fold++ {
		// Create training and test sets
		testStart := fold * foldSize
		testEnd := testStart + foldSize
		if fold == config.KFolds-1 {
			testEnd = n // Include remainder in last fold
		}
		
		// Extract test set
		testImpact := impactScores[testStart:testEnd]
		testValue := valueScores[testStart:testEnd]
		testContinuity := continuityScores[testStart:testEnd]
		testSpreads := spreads[testStart:testEnd]
		
		if len(testImpact) == 0 {
			continue
		}
		
		// Calculate hybrid scores for test set
		testHybridScores := make([]float64, len(testImpact))
		for i := range testImpact {
			testHybridScores[i] = weights.Impact*testImpact[i] + 
				weights.Value*testValue[i] + 
				weights.Continuity*testContinuity[i] +
				weights.Spread*0  // Spread not available in this context, use 0
		}
		
		// Calculate correlation with spreads
		correlation := calculateCorrelation(testHybridScores, testSpreads)
		
		// Calculate R² if needed
		var r2 float64
		if config.TargetMetric == "r2" || config.TargetMetric == "combined" {
			r2 = calculateRSquared(testHybridScores, testSpreads)
		}
		
		// Calculate combined score based on target metric
		var score float64
		switch config.TargetMetric {
		case "correlation":
			score = math.Abs(correlation) // Use absolute correlation
		case "r2":
			score = r2
		case "combined":
			score = config.CorrelationWeight*math.Abs(correlation) + config.R2Weight*r2
		default:
			score = math.Abs(correlation)
		}
		
		cvScores = append(cvScores, score)
	}
	
	if len(cvScores) == 0 {
		return 0, fmt.Errorf("no valid cross-validation scores")
	}
	
	// Return average CV score
	return calculateMean(cvScores), nil
}

// generateWeightCombinations creates a grid of weight combinations to test
func generateWeightCombinations(gridSize int) []ComponentWeights {
	var combinations []ComponentWeights
	
	// Generate combinations that sum to 1.0
	// Use smaller grid for 4 dimensions to keep reasonable computation time
	effectiveGridSize := gridSize
	if gridSize > 5 {
		effectiveGridSize = 5  // Limit to prevent combinatorial explosion
	}
	step := 1.0 / float64(effectiveGridSize-1)
	
	for i := 0; i < effectiveGridSize; i++ {
		for j := 0; j < effectiveGridSize; j++ {
			for k := 0; k < effectiveGridSize; k++ {
				for l := 0; l < effectiveGridSize; l++ {
					impact := float64(i) * step
					value := float64(j) * step
					continuity := float64(k) * step
					spread := float64(l) * step
					
					sum := impact + value + continuity + spread
					if sum > 0 {
						// Normalize to sum to 1
						weights := ComponentWeights{
							Impact:     impact / sum,
							Value:      value / sum,
							Continuity: continuity / sum,
							Spread:     spread / sum,
						}
						
						// Skip if any weight is too small (less than 5%)
						if weights.Impact >= 0.05 && weights.Value >= 0.05 && 
						   weights.Continuity >= 0.05 && weights.Spread >= 0.05 {
							combinations = append(combinations, weights)
						}
					}
				}
			}
		}
	}
	
	return combinations
}

// calculateRSquared computes the coefficient of determination (R²)
func calculateRSquared(predicted, actual []float64) float64 {
	if len(predicted) != len(actual) || len(predicted) < 2 {
		return 0
	}
	
	// Calculate mean of actual values
	actualMean := calculateMean(actual)
	
	// Calculate total sum of squares and residual sum of squares
	var totalSumSquares, residualSumSquares float64
	validCount := 0
	
	for i := 0; i < len(actual); i++ {
		if !math.IsNaN(actual[i]) && !math.IsInf(actual[i], 0) &&
			!math.IsNaN(predicted[i]) && !math.IsInf(predicted[i], 0) {
			
			totalSumSquares += (actual[i] - actualMean) * (actual[i] - actualMean)
			residualSumSquares += (actual[i] - predicted[i]) * (actual[i] - predicted[i])
			validCount++
		}
	}
	
	if validCount < 2 || totalSumSquares == 0 {
		return 0
	}
	
	r2 := 1 - (residualSumSquares / totalSumSquares)
	
	// Ensure R² is between 0 and 1
	if r2 < 0 {
		r2 = 0
	} else if r2 > 1 {
		r2 = 1
	}
	
	return r2
}

// DefaultWeights returns the default component weights based on empirical analysis
// Updated for 3-metric system after removing unreliable spread proxy
func DefaultWeights() ComponentWeights {
	return ComponentWeights{
		Impact:     0.40, // 40% weight to price impact (ILLIQ)
		Value:      0.35, // 35% weight to trading value
		Continuity: 0.25, // 25% weight to continuity
		Spread:     0.00, // REMOVED - spread proxy unreliable (57% zeros)
	}
}

// CalibratedWeights returns weights calibrated for different market conditions
// Updated for 3-metric system without spread proxy
func CalibratedWeights(marketCondition string) ComponentWeights {
	switch marketCondition {
	case "high_volatility":
		// Emphasize price impact during volatile periods
		return ComponentWeights{
			Impact:     0.50, // Increased from 45%
			Value:      0.30, // Slightly increased
			Continuity: 0.20, // Increased to compensate for spread removal
			Spread:     0.00, // Removed
		}
		
	case "low_volatility":
		// More balanced approach during calm periods
		return ComponentWeights{
			Impact:     0.35, // Slightly increased
			Value:      0.45, // Increased
			Continuity: 0.20, // Doubled to compensate
			Spread:     0.00, // Removed
		}
		
	case "value_focused":
		// Emphasize value when value patterns are key
		return ComponentWeights{
			Impact:     0.30, // Slightly increased
			Value:      0.55, // Increased
			Continuity: 0.15, // Tripled to compensate
			Spread:     0.00, // Removed
		}
		
	case "continuity_focused":
		// Emphasize continuity for long-term analysis
		return ComponentWeights{
			Impact:     0.35, // Slightly increased
			Value:      0.35, // Slightly increased
			Continuity: 0.30, // Increased to compensate
			Spread:     0.00, // Removed
		}
		
	default:
		return DefaultWeights()
	}
}

// OptimizeWeightsWithConstraints finds optimal weights subject to constraints
func OptimizeWeightsWithConstraints(ctx context.Context, impactScores, valueScores, continuityScores, spreads []float64, 
	minWeights, maxWeights ComponentWeights, config CalibrationConfig) (ComponentWeights, error) {
	
	logger := slog.Default()
	
	// Validate constraints
	if !validateWeightConstraints(minWeights, maxWeights) {
		return ComponentWeights{}, fmt.Errorf("invalid weight constraints")
	}
	
	logger.InfoContext(ctx, "optimizing weights with constraints",
		"min_weights", minWeights,
		"max_weights", maxWeights,
	)
	
	// Generate constrained weight combinations
	combinations := generateConstrainedWeights(minWeights, maxWeights, config.ParamGridSize)
	
	bestWeights := ComponentWeights{}
	bestScore := -math.Inf(1)
	
	for _, weights := range combinations {
		select {
		case <-ctx.Done():
			return ComponentWeights{}, fmt.Errorf("context cancelled during constrained optimization: %w", ctx.Err())
		default:
		}
		
		score, err := evaluateWeightsCV(impactScores, valueScores, continuityScores, spreads, weights, config)
		if err != nil {
			continue
		}
		
		if score > bestScore {
			bestScore = score
			bestWeights = weights
		}
	}
	
	if bestScore == -math.Inf(1) {
		return DefaultWeights(), fmt.Errorf("no valid constrained weight combinations found")
	}
	
	return bestWeights, nil
}

// generateConstrainedWeights creates weight combinations within specified bounds
func generateConstrainedWeights(minWeights, maxWeights ComponentWeights, gridSize int) []ComponentWeights {
	var combinations []ComponentWeights
	
	// Use smaller grid for 4 dimensions
	effectiveGridSize := gridSize
	if gridSize > 4 {
		effectiveGridSize = 4
	}
	step := 1.0 / float64(effectiveGridSize-1)
	
	for i := 0; i < effectiveGridSize; i++ {
		for j := 0; j < effectiveGridSize; j++ {
			for k := 0; k < effectiveGridSize; k++ {
				for l := 0; l < effectiveGridSize; l++ {
					impact := minWeights.Impact + float64(i)*step*(maxWeights.Impact-minWeights.Impact)
					value := minWeights.Value + float64(j)*step*(maxWeights.Value-minWeights.Value)
					continuity := minWeights.Continuity + float64(k)*step*(maxWeights.Continuity-minWeights.Continuity)
					spread := minWeights.Spread + float64(l)*step*(maxWeights.Spread-minWeights.Spread)
					
					sum := impact + value + continuity + spread
					if sum > 0 {
						weights := ComponentWeights{
							Impact:     impact / sum,
							Value:      value / sum,
							Continuity: continuity / sum,
							Spread:     spread / sum,
						}
						
						// Check constraints after normalization
						if weights.Impact >= minWeights.Impact && weights.Impact <= maxWeights.Impact &&
							weights.Value >= minWeights.Value && weights.Value <= maxWeights.Value &&
							weights.Continuity >= minWeights.Continuity && weights.Continuity <= maxWeights.Continuity &&
							weights.Spread >= minWeights.Spread && weights.Spread <= maxWeights.Spread {
							combinations = append(combinations, weights)
						}
					}
				}
			}
		}
	}
	
	return combinations
}

// AnalyzeWeightSensitivity performs sensitivity analysis on weight parameters
func AnalyzeWeightSensitivity(ctx context.Context, impactScores, valueScores, continuityScores, spreads []float64, 
	baseWeights ComponentWeights, perturbationSize float64) (map[string]float64, error) {
	
	results := make(map[string]float64)
	
	// Calculate baseline score
	baselineScore := calculateWeightScore(impactScores, valueScores, continuityScores, spreads, baseWeights)
	results["baseline"] = baselineScore
	
	// Test perturbations
	perturbations := []struct {
		name   string
		weights ComponentWeights
	}{
		{"impact_up", ComponentWeights{Impact: baseWeights.Impact + perturbationSize, Value: baseWeights.Value - perturbationSize/3, Continuity: baseWeights.Continuity - perturbationSize/3, Spread: baseWeights.Spread - perturbationSize/3}},
		{"impact_down", ComponentWeights{Impact: baseWeights.Impact - perturbationSize, Value: baseWeights.Value + perturbationSize/3, Continuity: baseWeights.Continuity + perturbationSize/3, Spread: baseWeights.Spread + perturbationSize/3}},
		{"value_up", ComponentWeights{Impact: baseWeights.Impact - perturbationSize/3, Value: baseWeights.Value + perturbationSize, Continuity: baseWeights.Continuity - perturbationSize/3, Spread: baseWeights.Spread - perturbationSize/3}},
		{"value_down", ComponentWeights{Impact: baseWeights.Impact + perturbationSize/3, Value: baseWeights.Value - perturbationSize, Continuity: baseWeights.Continuity + perturbationSize/3, Spread: baseWeights.Spread + perturbationSize/3}},
		{"continuity_up", ComponentWeights{Impact: baseWeights.Impact - perturbationSize/3, Value: baseWeights.Value - perturbationSize/3, Continuity: baseWeights.Continuity + perturbationSize, Spread: baseWeights.Spread - perturbationSize/3}},
		{"continuity_down", ComponentWeights{Impact: baseWeights.Impact + perturbationSize/3, Value: baseWeights.Value + perturbationSize/3, Continuity: baseWeights.Continuity - perturbationSize, Spread: baseWeights.Spread + perturbationSize/3}},
		{"spread_up", ComponentWeights{Impact: baseWeights.Impact - perturbationSize/3, Value: baseWeights.Value - perturbationSize/3, Continuity: baseWeights.Continuity - perturbationSize/3, Spread: baseWeights.Spread + perturbationSize}},
		{"spread_down", ComponentWeights{Impact: baseWeights.Impact + perturbationSize/3, Value: baseWeights.Value + perturbationSize/3, Continuity: baseWeights.Continuity + perturbationSize/3, Spread: baseWeights.Spread - perturbationSize}},
	}
	
	for _, p := range perturbations {
		// Normalize weights
		sum := p.weights.Impact + p.weights.Value + p.weights.Continuity + p.weights.Spread
		if sum > 0 && p.weights.Impact >= 0 && p.weights.Value >= 0 && p.weights.Continuity >= 0 && p.weights.Spread >= 0 {
			normalizedWeights := ComponentWeights{
				Impact:     p.weights.Impact / sum,
				Value:      p.weights.Value / sum,
				Continuity: p.weights.Continuity / sum,
				Spread:     p.weights.Spread / sum,
			}
			
			score := calculateWeightScore(impactScores, valueScores, continuityScores, spreads, normalizedWeights)
			results[p.name] = score
		}
	}
	
	return results, nil
}

// calculateWeightScore computes a single score for weight evaluation
func calculateWeightScore(impactScores, valueScores, continuityScores, spreads []float64, weights ComponentWeights) float64 {
	if len(impactScores) != len(spreads) {
		return 0
	}
	
	// Calculate hybrid scores
	hybridScores := make([]float64, len(impactScores))
	for i := range impactScores {
		hybridScores[i] = weights.Impact*impactScores[i] + 
			weights.Value*valueScores[i] + 
			weights.Continuity*continuityScores[i] +
			weights.Spread*0  // Spread not available in this context, use 0
	}
	
	// Return absolute correlation as score
	return math.Abs(calculateCorrelation(hybridScores, spreads))
}

// validateWeightFittingInputs validates inputs for weight fitting
func validateWeightFittingInputs(impactScores, valueScores, continuityScores, spreads []float64) error {
	n := len(impactScores)
	
	if n == 0 {
		return &ValidationError{
			Field:   "impactScores",
			Message: "no impact scores provided",
		}
	}
	
	if len(valueScores) != n {
		return &ValidationError{
			Field:   "valueScores",
			Message: "value scores length mismatch",
			Value:   map[string]int{"expected": n, "actual": len(valueScores)},
		}
	}
	
	if len(continuityScores) != n {
		return &ValidationError{
			Field:   "continuityScores",
			Message: "continuity scores length mismatch",
			Value:   map[string]int{"expected": n, "actual": len(continuityScores)},
		}
	}
	
	if len(spreads) != n {
		return &ValidationError{
			Field:   "spreads",
			Message: "spreads length mismatch",
			Value:   map[string]int{"expected": n, "actual": len(spreads)},
		}
	}
	
	// Check for sufficient valid data
	validCount := 0
	for i := 0; i < n; i++ {
		if !math.IsNaN(impactScores[i]) && !math.IsInf(impactScores[i], 0) &&
			!math.IsNaN(valueScores[i]) && !math.IsInf(valueScores[i], 0) &&
			!math.IsNaN(continuityScores[i]) && !math.IsInf(continuityScores[i], 0) &&
			!math.IsNaN(spreads[i]) && !math.IsInf(spreads[i], 0) {
			validCount++
		}
	}
	
	if validCount < MinObservationsForCalc {
		return &ValidationError{
			Field:   "validCount",
			Message: "insufficient valid observations for weight fitting",
			Value:   validCount,
		}
	}
	
	return nil
}

// validateWeightConstraints validates weight constraint bounds
func validateWeightConstraints(minWeights, maxWeights ComponentWeights) bool {
	return minWeights.Impact >= 0 && minWeights.Value >= 0 && minWeights.Continuity >= 0 && minWeights.Spread >= 0 &&
		maxWeights.Impact <= 1 && maxWeights.Value <= 1 && maxWeights.Continuity <= 1 && maxWeights.Spread <= 1 &&
		minWeights.Impact <= maxWeights.Impact &&
		minWeights.Value <= maxWeights.Value &&
		minWeights.Continuity <= maxWeights.Continuity &&
		minWeights.Spread <= maxWeights.Spread &&
		minWeights.Impact+minWeights.Value+minWeights.Continuity+minWeights.Spread <= 1.1 && // Allow small tolerance
		maxWeights.Impact+maxWeights.Value+maxWeights.Continuity+maxWeights.Spread >= 0.9
}