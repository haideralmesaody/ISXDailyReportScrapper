package liquidity

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"
)

// Calibrate performs comprehensive parameter calibration for the ISX Hybrid Liquidity Metric
// This includes grid search optimization of penalty parameters and component weights
//
// Parameters:
//   - data: map of ticker symbol to trading data
//   - config: calibration configuration settings
//
// Returns: calibration results with optimal parameters and performance metrics
func Calibrate(ctx context.Context, data map[string][]TradingDay, config CalibrationConfig) (*CalibrationResult, error) {
	start := time.Now()
	logger := slog.Default()
	
	logger.InfoContext(ctx, "starting parameter calibration",
		"num_tickers", len(data),
		"grid_size", config.ParamGridSize,
		"k_folds", config.KFolds,
		"target_metric", config.TargetMetric,
	)
	
	// Validate configuration
	if err := validateCalibrationConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Validate input data
	if err := validateCalibrationData(data, config); err != nil {
		return nil, fmt.Errorf("invalid input data: %w", err)
	}
	
	// Prepare calibration data
	calibrationData, err := prepareCalibrationData(ctx, data, config)
	if err != nil {
		return nil, fmt.Errorf("prepare calibration data: %w", err)
	}
	
	logger.InfoContext(ctx, "prepared calibration data",
		"observations", len(calibrationData.impactScores),
		"valid_tickers", len(calibrationData.tickerIndex),
	)
	
	// Generate parameter combinations
	paramCombinations := generateParameterCombinations(config.ParamGridSize)
	logger.InfoContext(ctx, "generated parameter combinations", "count", len(paramCombinations))
	
	// Perform grid search optimization
	bestResult, err := performGridSearch(ctx, calibrationData, paramCombinations, config)
	if err != nil {
		return nil, fmt.Errorf("grid search optimization: %w", err)
	}
	
	// Finalize calibration results
	result := &CalibrationResult{
		OptimalParams:     bestResult.params,
		OptimalWeights:    bestResult.weights,
		CrossValidationR2: bestResult.cvR2,
		SpreadCorrelation: bestResult.spreadCorr,
		CalibrationDate:   time.Now(),
		WindowUsed:        Window(len(data)),
		NumTickers:        len(data),
		NumObservations:   len(calibrationData.impactScores),
	}
	
	duration := time.Since(start)
	logger.InfoContext(ctx, "parameter calibration completed",
		"duration", duration,
		"optimal_r2", result.CrossValidationR2,
		"spread_correlation", result.SpreadCorrelation,
		"optimal_params", result.OptimalParams,
		"optimal_weights", result.OptimalWeights,
	)
	
	return result, nil
}

// calibrationData holds prepared data for parameter optimization
type calibrationData struct {
	impactScores     []float64
	volumeScores     []float64
	continuityScores []float64
	spreadProxies    []float64
	tickerIndex      []string
	dateIndex        []time.Time
}

// optimizationResult holds results from parameter evaluation
type optimizationResult struct {
	params     PenaltyParams
	weights    ComponentWeights
	cvR2       float64
	spreadCorr float64
	score      float64
}

// prepareCalibrationData extracts and preprocesses data for parameter calibration
func prepareCalibrationData(ctx context.Context, data map[string][]TradingDay, config CalibrationConfig) (*calibrationData, error) {
	var allImpact, allVolume, allContinuity, allSpreads []float64
	var tickerIndex []string
	var dateIndex []time.Time
	
	// Use default parameters for initial metric calculation
	defaultParams := DefaultPenaltyParams()
	defaultWeights := DefaultWeights()
	
	// Create a calculator with default settings
	calculator := NewCalculator(Window60, defaultParams, defaultWeights, slog.Default())
	
	// Calculate raw metrics for each ticker
	for symbol, tickerData := range data {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during data preparation: %w", ctx.Err())
		default:
		}
		
		if len(tickerData) < config.MinTradingDays {
			continue
		}
		
		windowData := tickerData[len(tickerData)-config.MinTradingDays:]
		metric, err := calculator.calculateWindowMetrics(ctx, symbol, tickerData[len(tickerData)-1].Date, windowData)
		if err != nil {
			continue
		}
		
		// Extract raw component scores (before cross-sectional scaling)
		allImpact = append(allImpact, metric.ILLIQ)
		allVolume = append(allVolume, metric.Value)
		allContinuity = append(allContinuity, metric.ContinuityNL)
		allSpreads = append(allSpreads, metric.SpreadProxy)
		tickerIndex = append(tickerIndex, symbol)
		dateIndex = append(dateIndex, metric.Date)
	}
	
	if len(allImpact) < config.MinTickers {
		return nil, fmt.Errorf("insufficient valid tickers for calibration: %d < %d", len(allImpact), config.MinTickers)
	}
	
	return &calibrationData{
		impactScores:     allImpact,
		volumeScores:     allVolume,
		continuityScores: allContinuity,
		spreadProxies:    allSpreads,
		tickerIndex:      tickerIndex,
		dateIndex:        dateIndex,
	}, nil
}

// generateParameterCombinations creates a grid of penalty parameter combinations
func generateParameterCombinations(gridSize int) []PenaltyParams {
	var combinations []PenaltyParams
	
	// Define parameter ranges based on empirical analysis
	betaRange := linspace(0.1, 0.8, gridSize)
	gammaRange := linspace(0.05, 0.4, gridSize)
	pStarRange := linspace(1.0, 5.0, gridSize)
	alphaRange := linspace(0.1, 0.5, gridSize)
	
	// Generate all combinations (reduced for performance)
	stepSize := max(1, gridSize/3) // Reduce combinations for efficiency
	
	for i := 0; i < len(betaRange); i += stepSize {
		for j := 0; j < len(gammaRange); j += stepSize {
			for k := 0; k < len(pStarRange); k += stepSize {
				for l := 0; l < len(alphaRange); l += stepSize {
					params := PenaltyParams{
						PiecewiseP0:        1.0, // Fixed reference level
						PiecewiseBeta:      betaRange[i],
						PiecewiseGamma:     gammaRange[j],
						PiecewisePStar:     pStarRange[k],
						PiecewiseMaxMult:   3.0, // Fixed maximum
						ExponentialP0:      1.0, // Fixed reference level
						ExponentialAlpha:   alphaRange[l],
						ExponentialMaxMult: 2.5, // Fixed maximum
					}
					
					if params.IsValid() {
						combinations = append(combinations, params)
					}
				}
			}
		}
	}
	
	return combinations
}

// performGridSearch executes the grid search optimization
func performGridSearch(ctx context.Context, data *calibrationData, paramCombinations []PenaltyParams, config CalibrationConfig) (*optimizationResult, error) {
	logger := slog.Default()
	
	bestResult := &optimizationResult{
		score: -math.Inf(1),
	}
	
	// Set up concurrent processing
	resultsChan := make(chan *optimizationResult, config.MaxConcurrency)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.MaxConcurrency)
	
	// Process parameter combinations
	for i, params := range paramCombinations {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during grid search: %w", ctx.Err())
		default:
		}
		
		wg.Add(1)
		go func(idx int, p PenaltyParams) {
			defer wg.Done()
			
			// Acquire semaphore for concurrency control
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result, err := evaluateParameterCombination(ctx, data, p, config)
			if err != nil {
				logger.DebugContext(ctx, "failed to evaluate parameter combination",
					"index", idx,
					"params", p,
					"error", err,
				)
				return
			}
			
			resultsChan <- result
		}(i, params)
	}
	
	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	// Collect results
	evaluatedCount := 0
	for result := range resultsChan {
		evaluatedCount++
		
		if result.score > bestResult.score {
			bestResult = result
			
			logger.DebugContext(ctx, "found better parameter combination",
				"evaluated", evaluatedCount,
				"score", result.score,
				"cv_r2", result.cvR2,
				"spread_corr", result.spreadCorr,
			)
		}
		
		if evaluatedCount%100 == 0 {
			logger.InfoContext(ctx, "grid search progress",
				"evaluated", evaluatedCount,
				"total", len(paramCombinations),
				"best_score", bestResult.score,
			)
		}
	}
	
	if bestResult.score == -math.Inf(1) {
		return nil, fmt.Errorf("no valid parameter combinations found")
	}
	
	logger.InfoContext(ctx, "grid search completed",
		"evaluated_combinations", evaluatedCount,
		"best_score", bestResult.score,
	)
	
	return bestResult, nil
}

// evaluateParameterCombination evaluates a single parameter combination
func evaluateParameterCombination(ctx context.Context, data *calibrationData, params PenaltyParams, config CalibrationConfig) (*optimizationResult, error) {
	// Apply penalties to get adjusted scores
	adjustedImpact := applyImpactPenalties(data.impactScores, params)
	adjustedVolume := applyVolumePenalties(data.volumeScores, params)
	
	// Scale the adjusted scores
	scaledImpact := RobustScale(adjustedImpact, true, true)   // Invert ILLIQ
	scaledVolume := RobustScale(adjustedVolume, false, true) // Don't invert volume
	scaledContinuity := RobustScale(data.continuityScores, false, false)
	
	// Fit optimal weights for this parameter combination
	weights, err := FitWeights(ctx, scaledImpact, scaledVolume, scaledContinuity, data.spreadProxies, config)
	if err != nil {
		return nil, fmt.Errorf("fit weights: %w", err)
	}
	
	// Calculate final hybrid scores
	hybridScores := make([]float64, len(scaledImpact))
	for i := range scaledImpact {
		hybridScores[i] = weights.Impact*scaledImpact[i] + 
			weights.Value*scaledVolume[i] + 
			weights.Continuity*scaledContinuity[i]
	}
	
	// Calculate performance metrics
	spreadCorr := calculateCorrelation(hybridScores, data.spreadProxies)
	r2 := calculateRSquared(hybridScores, data.spreadProxies)
	
	// Calculate combined score based on target metric
	var score float64
	switch config.TargetMetric {
	case "correlation":
		score = math.Abs(spreadCorr)
	case "r2":
		score = r2
	case "combined":
		score = config.CorrelationWeight*math.Abs(spreadCorr) + config.R2Weight*r2
	default:
		score = math.Abs(spreadCorr)
	}
	
	return &optimizationResult{
		params:     params,
		weights:    weights,
		cvR2:       r2,
		spreadCorr: spreadCorr,
		score:      score,
	}, nil
}

// applyImpactPenalties applies penalty adjustments to impact scores
func applyImpactPenalties(impactScores []float64, params PenaltyParams) []float64 {
	adjusted := make([]float64, len(impactScores))
	for i, score := range impactScores {
		// Use a representative price for penalty calculation
		// This is simplified - in practice, you'd use actual price data
		representativePrice := 2.0 // Median ISX price level
		penalty := PiecewisePenalty(representativePrice, params.PiecewiseBeta, 
			params.PiecewiseGamma, params.PiecewisePStar, params.PiecewiseMaxMult)
		adjusted[i] = score * penalty
	}
	return adjusted
}

// applyVolumePenalties applies penalty adjustments to volume scores
func applyVolumePenalties(volumeScores []float64, params PenaltyParams) []float64 {
	adjusted := make([]float64, len(volumeScores))
	for i, score := range volumeScores {
		representativePrice := 2.0
		penalty := ExponentialPenalty(representativePrice, params.ExponentialAlpha, params.ExponentialMaxMult)
		adjusted[i] = score * penalty
	}
	return adjusted
}

// validateCalibrationConfig validates the calibration configuration
func validateCalibrationConfig(config CalibrationConfig) error {
	if config.ParamGridSize < 2 {
		return &ValidationError{
			Field:   "ParamGridSize",
			Message: "parameter grid size must be at least 2",
			Value:   config.ParamGridSize,
		}
	}
	
	if config.KFolds < 2 {
		return &ValidationError{
			Field:   "KFolds",
			Message: "k-folds must be at least 2",
			Value:   config.KFolds,
		}
	}
	
	if config.MinTickers < 2 {
		return &ValidationError{
			Field:   "MinTickers",
			Message: "minimum tickers must be at least 2",
			Value:   config.MinTickers,
		}
	}
	
	if config.MaxConcurrency < 1 {
		return &ValidationError{
			Field:   "MaxConcurrency",
			Message: "max concurrency must be at least 1",
			Value:   config.MaxConcurrency,
		}
	}
	
	validMetrics := []string{"correlation", "r2", "combined"}
	valid := false
	for _, metric := range validMetrics {
		if config.TargetMetric == metric {
			valid = true
			break
		}
	}
	if !valid {
		return &ValidationError{
			Field:   "TargetMetric",
			Message: "target metric must be one of: correlation, r2, combined",
			Value:   config.TargetMetric,
		}
	}
	
	return nil
}

// validateCalibrationData validates the input data for calibration
func validateCalibrationData(data map[string][]TradingDay, config CalibrationConfig) error {
	if len(data) < config.MinTickers {
		return &ValidationError{
			Field:   "data",
			Message: "insufficient tickers for calibration",
			Value:   len(data),
		}
	}
	
	validTickers := 0
	for _, tickerData := range data {
		if len(tickerData) >= config.MinTradingDays {
			validTickers++
		}
	}
	
	if validTickers < config.MinTickers {
		return &ValidationError{
			Field:   "validTickers",
			Message: "insufficient valid tickers for calibration",
			Value:   validTickers,
		}
	}
	
	return nil
}

// linspace creates linearly spaced values between start and stop
func linspace(start, stop float64, num int) []float64 {
	if num <= 0 {
		return nil
	}
	if num == 1 {
		return []float64{start}
	}
	
	result := make([]float64, num)
	step := (stop - start) / float64(num-1)
	
	for i := 0; i < num; i++ {
		result[i] = start + float64(i)*step
	}
	
	return result
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// DefaultCalibrationConfig returns default configuration for parameter calibration
func DefaultCalibrationConfig() CalibrationConfig {
	return CalibrationConfig{
		ParamGridSize:     5,            // 5x5x5x5 = 625 combinations (manageable)
		MinIterations:     10,
		MaxIterations:     1000,
		Tolerance:         1e-6,
		KFolds:            5,
		RandomSeed:        42,
		TargetMetric:      "combined",
		R2Weight:          0.6,
		CorrelationWeight: 0.4,
		MinTradingDays:    MinTradingDaysForCalc,
		MinTickers:        10,
		MaxConcurrency:    4,
		EnableProfiling:   false,
	}
}