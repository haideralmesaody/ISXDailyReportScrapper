package liquidity

import (
	"fmt"
	"math"
	"time"
)

// ValidateParams performs comprehensive validation of penalty parameters
func ValidateParams(params PenaltyParams) error {
	// Basic parameter validation
	if err := ValidatePenaltyParams(params); err != nil {
		return fmt.Errorf("penalty parameters validation failed: %w", err)
	}
	
	// Additional range checks for ISX-specific constraints
	if params.PiecewisePStar < 0.5 || params.PiecewisePStar > 10.0 {
		return &ValidationError{
			Field:   "PiecewisePStar",
			Message: "piecewise pStar should be between 0.5 and 10.0 for ISX data",
			Value:   params.PiecewisePStar,
		}
	}
	
	// Check parameter relationships
	if params.PiecewiseBeta <= params.PiecewiseGamma {
		return &ValidationError{
			Field:   "PiecewiseBetaGamma",
			Message: "piecewise beta should be greater than gamma (low-price penalty > high-price penalty)",
			Value:   map[string]float64{"beta": params.PiecewiseBeta, "gamma": params.PiecewiseGamma},
		}
	}
	
	return nil
}

// ValidateWeights performs comprehensive validation of component weights
func ValidateWeights(weights ComponentWeights) error {
	// Check individual weight bounds
	if weights.Impact < 0 || weights.Impact > 1 {
		return &ValidationError{
			Field:   "Impact",
			Message: "impact weight must be between 0 and 1",
			Value:   weights.Impact,
		}
	}
	
	if weights.Value < 0 || weights.Value > 1 {
		return &ValidationError{
			Field:   "Value",
			Message: "value weight must be between 0 and 1",
			Value:   weights.Value,
		}
	}
	
	if weights.Continuity < 0 || weights.Continuity > 1 {
		return &ValidationError{
			Field:   "Continuity",
			Message: "continuity weight must be between 0 and 1",
			Value:   weights.Continuity,
		}
	}
	
	// Check sum constraint
	sum := weights.Impact + weights.Value + weights.Continuity
	if sum < 0.99 || sum > 1.01 {
		return &ValidationError{
			Field:   "WeightSum",
			Message: "component weights must sum to 1.0",
			Value:   sum,
		}
	}
	
	// Check minimum weight constraints to avoid extreme distributions
	minWeight := 0.05 // 5% minimum
	if weights.Impact < minWeight || weights.Value < minWeight || weights.Continuity < minWeight {
		return &ValidationError{
			Field:   "MinimumWeights",
			Message: "each component weight should be at least 5% to ensure balanced measurement",
			Value:   weights,
		}
	}
	
	return nil
}

// ValidateTradingData performs comprehensive validation of trading data
func ValidateTradingData(data []TradingDay) error {
	if len(data) == 0 {
		return &ValidationError{
			Field:   "data",
			Message: "no trading data provided",
		}
	}
	
	// Check minimum data requirements
	if len(data) < MinObservationsForCalc {
		return &ValidationError{
			Field:   "dataLength",
			Message: "insufficient trading data for calculation",
			Value:   map[string]int{"provided": len(data), "required": MinObservationsForCalc},
		}
	}
	
	var validCount, tradingCount int
	var symbols = make(map[string]bool)
	var dates []time.Time
	var priceErrors, volumeErrors, dateErrors int
	
	for i, td := range data {
		// Track symbols
		if td.Symbol != "" {
			symbols[td.Symbol] = true
		}
		
		// Collect dates for sorting check
		dates = append(dates, td.Date)
		
		// Validate individual record
		if err := validateTradingDayRecord(td, i); err != nil {
			switch err.(*ValidationError).Field {
			case "prices":
				priceErrors++
			case "volume":
				volumeErrors++
			case "date":
				dateErrors++
			}
			continue
		}
		
		validCount++
		if td.IsTrading() {
			tradingCount++
		}
	}
	
	// Check data quality thresholds
	validRatio := float64(validCount) / float64(len(data))
	if validRatio < 0.5 {
		return &ValidationError{
			Field:   "dataQuality",
			Message: "data quality too low for reliable calculation",
			Value:   map[string]interface{}{
				"valid_ratio":   validRatio,
				"price_errors":  priceErrors,
				"volume_errors": volumeErrors,
				"date_errors":   dateErrors,
			},
		}
	}
	
	if tradingCount < MinTradingDaysForCalc {
		return &ValidationError{
			Field:   "tradingDays",
			Message: "insufficient trading days for reliable calculation",
			Value:   map[string]int{"trading_days": tradingCount, "required": MinTradingDaysForCalc},
		}
	}
	
	// Check for multiple symbols (data consistency)
	if len(symbols) > 1 {
		return &ValidationError{
			Field:   "symbols",
			Message: "trading data contains multiple symbols, expected single ticker",
			Value:   len(symbols),
		}
	}
	
	// Check date sorting
	if !isDatesSorted(dates) {
		return &ValidationError{
			Field:   "dateOrder",
			Message: "trading data must be sorted by date in ascending order",
		}
	}
	
	return nil
}

// validateTradingDayRecord validates a single trading day record
func validateTradingDayRecord(td TradingDay, index int) error {
	// Check date
	if td.Date.IsZero() {
		return &ValidationError{
			Field:   "date",
			Message: fmt.Sprintf("invalid date at record %d", index),
			Value:   td.Date,
		}
	}
	
	// Check symbol
	if td.Symbol == "" {
		return &ValidationError{
			Field:   "symbol",
			Message: fmt.Sprintf("empty symbol at record %d", index),
			Value:   td.Symbol,
		}
	}
	
	// Check OHLC prices
	if td.Open <= 0 || td.High <= 0 || td.Low <= 0 || td.Close <= 0 {
		return &ValidationError{
			Field:   "prices",
			Message: fmt.Sprintf("invalid OHLC prices at record %d", index),
			Value:   map[string]float64{"open": td.Open, "high": td.High, "low": td.Low, "close": td.Close},
		}
	}
	
	// Check price relationships
	if td.High < td.Low || td.High < td.Open || td.High < td.Close || 
	   td.Low > td.Open || td.Low > td.Close {
		return &ValidationError{
			Field:   "prices",
			Message: fmt.Sprintf("inconsistent OHLC relationships at record %d", index),
			Value:   map[string]float64{"open": td.Open, "high": td.High, "low": td.Low, "close": td.Close},
		}
	}
	
	// Check for NaN or Inf values
	prices := []float64{td.Open, td.High, td.Low, td.Close, td.Volume}
	for _, price := range prices {
		if math.IsNaN(price) || math.IsInf(price, 0) {
			return &ValidationError{
				Field:   "prices",
				Message: fmt.Sprintf("NaN or Inf price values at record %d", index),
			}
		}
	}
	
	// Check volume
	if td.Volume < 0 {
		return &ValidationError{
			Field:   "volume",
			Message: fmt.Sprintf("negative volume at record %d", index),
			Value:   td.Volume,
		}
	}
	
	// Check number of trades
	if td.NumTrades < 0 {
		return &ValidationError{
			Field:   "numTrades",
			Message: fmt.Sprintf("negative number of trades at record %d", index),
			Value:   td.NumTrades,
		}
	}
	
	// Check consistency: if volume > 0, should have trades
	if td.Volume > 0 && td.NumTrades == 0 {
		return &ValidationError{
			Field:   "consistency",
			Message: fmt.Sprintf("volume without trades at record %d", index),
			Value:   map[string]interface{}{"volume": td.Volume, "trades": td.NumTrades},
		}
	}
	
	return nil
}

// ValidateCalculationInputs validates inputs for metric calculation
func ValidateCalculationInputs(data []TradingDay, window Window, params PenaltyParams, weights ComponentWeights) error {
	// Validate trading data
	if err := ValidateTradingData(data); err != nil {
		return fmt.Errorf("trading data validation: %w", err)
	}
	
	// Validate window
	if err := ValidateWindow(window, data); err != nil {
		return fmt.Errorf("window validation: %w", err)
	}
	
	// Validate parameters
	if err := ValidateParams(params); err != nil {
		return fmt.Errorf("parameter validation: %w", err)
	}
	
	// Validate weights
	if err := ValidateWeights(weights); err != nil {
		return fmt.Errorf("weights validation: %w", err)
	}
	
	return nil
}

// ValidateWindow validates the calculation window against available data
func ValidateWindow(window Window, data []TradingDay) error {
	if window <= 0 {
		return &ValidationError{
			Field:   "window",
			Message: "window size must be positive",
			Value:   window,
		}
	}
	
	// Check supported windows
	supportedWindows := []Window{Window20, Window60, Window120}
	supported := false
	for _, w := range supportedWindows {
		if window == w {
			supported = true
			break
		}
	}
	
	if !supported {
		return &ValidationError{
			Field:   "window",
			Message: "unsupported window size",
			Value:   window,
		}
	}
	
	// Check data sufficiency
	if len(data) < window.Days() {
		return &ValidationError{
			Field:   "dataWindow",
			Message: "insufficient data for requested window",
			Value:   map[string]int{"data_points": len(data), "window_required": window.Days()},
		}
	}
	
	// Count trading days in the data
	tradingDays := 0
	for _, td := range data {
		if td.IsTrading() {
			tradingDays++
		}
	}
	
	requiredTradingDays := int(float64(window.Days()) * 0.3) // At least 30% should be trading days
	if tradingDays < requiredTradingDays {
		return &ValidationError{
			Field:   "tradingDaysWindow",
			Message: "insufficient trading days in window for reliable calculation",
			Value:   map[string]int{"trading_days": tradingDays, "required": requiredTradingDays},
		}
	}
	
	return nil
}

// ValidateMetricsOutput validates calculated metrics for reasonableness
func ValidateMetricsOutput(metrics []TickerMetrics) error {
	if len(metrics) == 0 {
		return &ValidationError{
			Field:   "metrics",
			Message: "no metrics calculated",
		}
	}
	
	for i, metric := range metrics {
		if err := validateSingleMetric(metric, i); err != nil {
			return fmt.Errorf("metric validation failed at index %d: %w", i, err)
		}
	}
	
	// Check for reasonable distribution
	if err := validateMetricsDistribution(metrics); err != nil {
		return fmt.Errorf("metrics distribution validation: %w", err)
	}
	
	return nil
}

// validateSingleMetric validates a single ticker metric
func validateSingleMetric(metric TickerMetrics, index int) error {
	// Check required fields
	if metric.Symbol == "" {
		return &ValidationError{
			Field:   "symbol",
			Message: "empty symbol",
		}
	}
	
	if metric.Date.IsZero() {
		return &ValidationError{
			Field:   "date",
			Message: "invalid date",
		}
	}
	
	// Check metric ranges
	if metric.HybridScore < 0 || metric.HybridScore > 100 {
		return &ValidationError{
			Field:   "hybridScore",
			Message: "hybrid score outside expected range [0, 100]",
			Value:   metric.HybridScore,
		}
	}
	
	// Check scaled values
	scaledValues := []float64{metric.ILLIQScaled, metric.ValueScaled, metric.ContinuityScaled}
	for j, val := range scaledValues {
		if val < 0 || val > 100 {
			fieldNames := []string{"ILLIQScaled", "ValueScaled", "ContinuityScaled"}
			return &ValidationError{
				Field:   fieldNames[j],
				Message: "scaled value outside expected range [0, 100]",
				Value:   val,
			}
		}
	}
	
	// Check penalty multipliers
	if metric.ImpactPenalty < 1 || metric.ValuePenalty < 1 {
		return &ValidationError{
			Field:   "penalties",
			Message: "penalty multipliers should be >= 1.0",
			Value:   map[string]float64{"impact": metric.ImpactPenalty, "value": metric.ValuePenalty},
		}
	}
	
	// Check continuity
	if metric.Continuity < 0 || metric.Continuity > 1 {
		return &ValidationError{
			Field:   "continuity",
			Message: "continuity ratio outside range [0, 1]",
			Value:   metric.Continuity,
		}
	}
	
	// Check trading days consistency
	if metric.TradingDays > metric.TotalDays {
		return &ValidationError{
			Field:   "tradingDays",
			Message: "trading days cannot exceed total days",
			Value:   map[string]int{"trading": metric.TradingDays, "total": metric.TotalDays},
		}
	}
	
	// Check for NaN or Inf values
	values := []float64{
		metric.ILLIQ, metric.ILLIQScaled, metric.Value, metric.ValueScaled,
		metric.Continuity, metric.ContinuityNL, metric.ContinuityScaled,
		metric.ImpactPenalty, metric.ValuePenalty, metric.HybridScore,
		metric.SpreadProxy, metric.AvgReturn, metric.ReturnVolatility,
	}
	
	for j, val := range values {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return &ValidationError{
				Field:   "numericValues",
				Message: "metric contains NaN or Inf values",
				Value:   map[string]interface{}{"index": j, "value": val},
			}
		}
	}
	
	return nil
}

// validateMetricsDistribution validates the overall distribution of calculated metrics
func validateMetricsDistribution(metrics []TickerMetrics) error {
	if len(metrics) < 2 {
		return nil // Cannot validate distribution with less than 2 metrics
	}
	
	// Extract hybrid scores for distribution analysis
	var hybridScores []float64
	for _, metric := range metrics {
		hybridScores = append(hybridScores, metric.HybridScore)
	}
	
	// Check for reasonable variance
	mean := calculateMean(hybridScores)
	stdDev := calculateStandardDeviation(hybridScores, mean)
	
	if stdDev < 0.001 {
		return &ValidationError{
			Field:   "distribution",
			Message: "hybrid scores show insufficient variance, possible calculation error",
			Value:   stdDev,
		}
	}
	
	// Check for extreme outliers (more than 3 standard deviations)
	outlierCount := 0
	for _, score := range hybridScores {
		zScore := math.Abs((score - mean) / stdDev)
		if zScore > 3 {
			outlierCount++
		}
	}
	
	outlierRatio := float64(outlierCount) / float64(len(hybridScores))
	if outlierRatio > 0.1 { // More than 10% outliers
		return &ValidationError{
			Field:   "outliers",
			Message: "excessive number of outliers in hybrid scores",
			Value:   map[string]interface{}{
				"outlier_ratio": outlierRatio,
				"outlier_count": outlierCount,
				"total":         len(hybridScores),
			},
		}
	}
	
	return nil
}

// isDatesSorted checks if dates are in ascending order
func isDatesSorted(dates []time.Time) bool {
	for i := 1; i < len(dates); i++ {
		if dates[i].Before(dates[i-1]) {
			return false
		}
	}
	return true
}

// ValidateCalibrationResult validates calibration results
func ValidateCalibrationResult(result *CalibrationResult) error {
	if result == nil {
		return &ValidationError{
			Field:   "result",
			Message: "calibration result is nil",
		}
	}
	
	// Validate optimal parameters
	if err := ValidateParams(result.OptimalParams); err != nil {
		return fmt.Errorf("optimal parameters validation: %w", err)
	}
	
	// Validate optimal weights
	if err := ValidateWeights(result.OptimalWeights); err != nil {
		return fmt.Errorf("optimal weights validation: %w", err)
	}
	
	// Validate performance metrics
	if result.CrossValidationR2 < 0 || result.CrossValidationR2 > 1 {
		return &ValidationError{
			Field:   "crossValidationR2",
			Message: "R² must be between 0 and 1",
			Value:   result.CrossValidationR2,
		}
	}
	
	if result.SpreadCorrelation < -1 || result.SpreadCorrelation > 1 {
		return &ValidationError{
			Field:   "spreadCorrelation",
			Message: "correlation must be between -1 and 1",
			Value:   result.SpreadCorrelation,
		}
	}
	
	// Check minimum data requirements
	if result.NumTickers < 2 {
		return &ValidationError{
			Field:   "numTickers",
			Message: "calibration requires at least 2 tickers",
			Value:   result.NumTickers,
		}
	}
	
	if result.NumObservations < MinObservationsForCalc {
		return &ValidationError{
			Field:   "numObservations",
			Message: "insufficient observations for reliable calibration",
			Value:   result.NumObservations,
		}
	}
	
	return nil
}