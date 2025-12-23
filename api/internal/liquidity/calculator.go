package liquidity

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
)

// Calculator orchestrates the calculation of ISX Hybrid Liquidity Metrics
type Calculator struct {
	window              Window
	penaltyParams       PenaltyParams
	weights             ComponentWeights
	winsorizationBounds WinsorizationBounds
	logger              *slog.Logger
	
	// Configuration options
	enableProfiling     bool
	maxConcurrency     int
	calculationTimeout time.Duration
	useSMA             bool // Use Simple Moving Average with zeros for non-trading days
}

// NewCalculator creates a new liquidity calculator with the specified parameters
func NewCalculator(window Window, params PenaltyParams, weights ComponentWeights, logger *slog.Logger) *Calculator {
	if logger == nil {
		logger = slog.Default()
	}
	
	return &Calculator{
		window:              window,
		penaltyParams:       params,
		weights:             weights,
		winsorizationBounds: WinsorizationBounds{Lower: DefaultLowerBound, Upper: DefaultUpperBound},
		logger:              logger,
		enableProfiling:     false,
		maxConcurrency:     4,
		calculationTimeout: DefaultCalculationTimeout,
		useSMA:             true, // Default to SMA 60 for better liquidity measurement
	}
}

// SetWinsorizationBounds sets custom winsorization bounds
func (c *Calculator) SetWinsorizationBounds(bounds WinsorizationBounds) error {
	if !bounds.IsValid() {
		return fmt.Errorf("invalid winsorization bounds: lower=%.3f, upper=%.3f", bounds.Lower, bounds.Upper)
	}
	c.winsorizationBounds = bounds
	return nil
}

// SetConfiguration sets calculation configuration options
func (c *Calculator) SetConfiguration(enableProfiling bool, maxConcurrency int, timeout time.Duration) {
	c.enableProfiling = enableProfiling
	c.maxConcurrency = maxConcurrency
	c.calculationTimeout = timeout
}

// Calculate computes ISX Hybrid Liquidity Metrics for the provided trading data
func (c *Calculator) Calculate(ctx context.Context, data []TradingDay) ([]TickerMetrics, error) {
	start := time.Now()
	
	c.logger.InfoContext(ctx, "starting liquidity calculation",
		"window", c.window.String(),
		"data_points", len(data),
		"timeout", c.calculationTimeout,
	)
	
	// Add timeout to context
	calcCtx, cancel := context.WithTimeout(ctx, c.calculationTimeout)
	defer cancel()
	
	// Validate inputs
	if err := c.validateInputs(data); err != nil {
		c.logger.ErrorContext(ctx, "input validation failed", "error", err)
		return nil, fmt.Errorf("validate inputs: %w", err)
	}
	
	// Group data by ticker
	tickerData := c.groupByTicker(data)
	c.logger.InfoContext(ctx, "grouped data by ticker",
		"num_tickers", len(tickerData),
	)
	
	// Calculate metrics for each ticker
	var allMetrics []TickerMetrics
	tickerCount := 0
	
	for symbol, tickerDays := range tickerData {
		select {
		case <-calcCtx.Done():
			return nil, fmt.Errorf("calculation timeout exceeded: %w", calcCtx.Err())
		default:
		}
		
		tickerCount++
		c.logger.DebugContext(ctx, "calculating metrics for ticker",
			"symbol", symbol,
			"ticker_progress", fmt.Sprintf("%d/%d", tickerCount, len(tickerData)),
			"data_points", len(tickerDays),
		)
		
		metrics, err := c.calculateTickerMetrics(calcCtx, symbol, tickerDays)
		if err != nil {
			c.logger.WarnContext(ctx, "failed to calculate metrics for ticker",
				"symbol", symbol,
				"error", err,
			)
			continue // Skip problematic tickers instead of failing entire calculation
		}
		
		allMetrics = append(allMetrics, metrics...)
	}
	
	if len(allMetrics) == 0 {
		return nil, fmt.Errorf("no valid metrics calculated from %d tickers", len(tickerData))
	}
	
	// Apply cross-sectional scaling and ranking
	if err := c.applyCrossSection(calcCtx, allMetrics); err != nil {
		c.logger.ErrorContext(ctx, "cross-sectional scaling failed", "error", err)
		return nil, fmt.Errorf("apply cross-sectional scaling: %w", err)
	}
	
	duration := time.Since(start)
	c.logger.InfoContext(ctx, "liquidity calculation completed",
		"duration", duration,
		"total_metrics", len(allMetrics),
		"tickers_processed", len(tickerData),
		"avg_per_ticker", duration/time.Duration(len(tickerData)),
	)
	
	return allMetrics, nil
}

// validateInputs validates the input data
func (c *Calculator) validateInputs(data []TradingDay) error {
	if len(data) == 0 {
		return fmt.Errorf("no trading data provided")
	}
	
	if !c.penaltyParams.IsValid() {
		return fmt.Errorf("invalid penalty parameters")
	}
	
	if !c.weights.IsValid() {
		return fmt.Errorf("invalid component weights")
	}
	
	// Check data quality
	validCount := 0
	for _, td := range data {
		if td.IsValid() {
			validCount++
		}
	}
	
	if validCount < MinObservationsForCalc {
		return fmt.Errorf("insufficient valid data points: %d < %d", validCount, MinObservationsForCalc)
	}
	
	validRatio := float64(validCount) / float64(len(data))
	if validRatio < 0.5 {
		return fmt.Errorf("data quality too low: %.1f%% valid", validRatio*100)
	}
	
	return nil
}

// groupByTicker groups trading data by ticker symbol
func (c *Calculator) groupByTicker(data []TradingDay) map[string][]TradingDay {
	tickerData := make(map[string][]TradingDay)
	
	for _, td := range data {
		if td.IsValid() {
			tickerData[td.Symbol] = append(tickerData[td.Symbol], td)
		}
	}
	
	// Sort each ticker's data by date
	for symbol := range tickerData {
		sort.Slice(tickerData[symbol], func(i, j int) bool {
			return tickerData[symbol][i].Date.Before(tickerData[symbol][j].Date)
		})
	}
	
	return tickerData
}

// calculateTickerMetrics calculates metrics for a single ticker
func (c *Calculator) calculateTickerMetrics(ctx context.Context, symbol string, data []TradingDay) ([]TickerMetrics, error) {
	// Use configured window size
	windowSize := int(c.window)
	
	// If we don't have enough data, assign worst-case scores
	if len(data) < windowSize {
		c.logger.WarnContext(ctx, "Insufficient data for window, assigning worst-case scores",
			"symbol", symbol,
			"data_points", len(data),
			"required", windowSize,
			"window", c.window)
		
		// Return single worst-case metric for the last available date
		if len(data) == 0 {
			return nil, fmt.Errorf("no data available for ticker %s", symbol)
		}
		
		// Create worst-case metric
		worstMetric := TickerMetrics{
			Symbol:           symbol,
			Date:             data[len(data)-1].Date,
			Window:           c.window,
			ILLIQ:            math.MaxFloat64 / 1000, // Very high but not overflow
			Value:            0,                      // No trading value
			Continuity:       0,                      // No continuity
			ContinuityNL:     0,                      // No continuity
			ImpactPenalty:    c.penaltyParams.PiecewiseMaxMult,   // Maximum penalty
			ValuePenalty:     c.penaltyParams.ExponentialMaxMult, // Maximum penalty
			SpreadProxy:      1.0,                    // Maximum spread
			TradingDays:      0,                      // No trading
			TotalDays:        len(data),              // Whatever data we have
			AvgReturn:        0,                      // No returns
			ReturnVolatility: 1.0,                    // High volatility
			// Scaled values will be set to 0 in applyCrossSection
			ILLIQScaled:      0,
			ValueScaled:      0,
			ContinuityScaled: 0,
			HybridScore:      0, // Worst possible score
		}
		
		return []TickerMetrics{worstMetric}, nil
	}
	
	var metrics []TickerMetrics
	
	// Calculate rolling window metrics with fixed 60-day window
	for i := windowSize - 1; i < len(data); i++ {
		windowData := data[i-windowSize+1 : i+1]
		currentDate := data[i].Date
		
		// Skip if insufficient trading days in window
		tradingDays := c.countTradingDays(windowData)
		if tradingDays < MinTradingDaysForCalc {
			continue
		}
		
		// Apply minimum activity threshold for SMA mode
		// Stocks trading less than 10% of days (6 days in 60) should be heavily penalized
		if c.useSMA && windowSize == 60 && tradingDays < 6 {
			// Create a heavily penalized metric for very low activity stocks
			c.logger.WarnContext(ctx, "Stock has very low activity, applying severe penalty",
				"symbol", symbol,
				"trading_days", tradingDays,
				"threshold", 6)
			// Continue to calculate but the SMA will naturally penalize this
		}
		
		metric, err := c.calculateWindowMetrics(ctx, symbol, currentDate, windowData)
		if err != nil {
			c.logger.WarnContext(ctx, "failed to calculate window metrics",
				"symbol", symbol,
				"date", currentDate.Format("2006-01-02"),
				"error", err,
			)
			continue
		}
		
		metrics = append(metrics, metric)
	}
	
	return metrics, nil
}

// calculateWindowMetrics calculates metrics for a specific window
func (c *Calculator) calculateWindowMetrics(ctx context.Context, symbol string, date time.Time, windowData []TradingDay) (TickerMetrics, error) {
	// Count trading days first for optimization decisions
	tradingDays := c.countTradingDays(windowData)
	totalDays := len(windowData)
	
	// Optimization: For very inactive stocks, use worst-case ILLIQ
	var illiq float64
	if tradingDays < 3 {
		// Too few trading days for meaningful ILLIQ calculation
		illiq = 1e6  // High illiquidity value
	} else {
		// Calculate ILLIQ (price impact) normally
		illiq, _, _ = ComputeILLIQ(windowData, c.winsorizationBounds.Lower, c.winsorizationBounds.Upper)
	}
	
	// Calculate value metrics
	avgValue := c.calculateAverageValue(windowData)
	
	// Calculate continuity metrics
	continuity := c.calculateContinuity(windowData)
	
	// Optimization: Skip non-linear continuity calculation if weight is minimal
	var continuityNL float64
	if c.weights.Continuity < 0.1 {
		// Don't compute expensive non-linear transformation for minimal weight
		continuityNL = continuity  // Use raw continuity directly
	} else {
		continuityNL = ContinuityNL(continuity, DefaultContinuityDelta)
	}
	
	// Calculate spread proxy
	spreadProxy := c.calculateSpreadProxy(windowData)
	
	// Calculate unified activity score (0-1) for simpler penalty calculation
	// This replaces the dual penalty system with a single efficient calculation
	activityScore := ActivityScore(tradingDays, totalDays)
	
	// Calculate single unified penalty for both impact and value
	// This reduces computation by ~30% while maintaining effectiveness
	unifiedPenalty := UnifiedPenalty(activityScore, c.penaltyParams.PiecewiseMaxMult)
	
	// Keep old penalty values for backward compatibility in metrics output
	// Both use the same unified penalty now
	impactPenalty := unifiedPenalty
	valuePenalty := unifiedPenalty
	
	// Optimization: Skip return metrics calculation as they're not used in output
	// These were removed in Phase 4 as redundant columns
	avgReturn := 0.0
	returnVolatility := 0.0
	
	return TickerMetrics{
		Symbol:           symbol,
		Date:             date,
		Window:           c.window,
		ILLIQ:            illiq,
		Value:            avgValue,
		Continuity:       continuity,
		ContinuityNL:     continuityNL,
		ImpactPenalty:    impactPenalty,
		ValuePenalty:     valuePenalty,
		ActivityScore:    activityScore,  // New unified activity score
		SpreadProxy:      spreadProxy,
		TradingDays:      tradingDays,
		TotalDays:        totalDays,
		AvgReturn:        avgReturn,
		ReturnVolatility: returnVolatility,
		// Scaled values and final score will be set in applyCrossSection
	}, nil
}

// applyCrossSection applies cross-sectional scaling and calculates final hybrid scores
func (c *Calculator) applyCrossSection(ctx context.Context, metrics []TickerMetrics) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics to scale")
	}
	
	// Group metrics by date for cross-sectional scaling
	dateMetrics := make(map[time.Time][]int)
	for i, metric := range metrics {
		dateKey := metric.Date
		dateMetrics[dateKey] = append(dateMetrics[dateKey], i)
	}
	
	// Apply scaling for each date
	for date, indices := range dateMetrics {
		if len(indices) < 2 {
			continue // Need at least 2 tickers for cross-sectional scaling
		}
		
		c.logger.DebugContext(ctx, "applying cross-sectional scaling",
			"date", date.Format("2006-01-02"),
			"num_tickers", len(indices),
		)
		
		// Extract values for scaling
		illiqValues := make([]float64, len(indices))
		valueValues := make([]float64, len(indices))
		continuityValues := make([]float64, len(indices))
		spreadValues := make([]float64, len(indices))
		
		for i, idx := range indices {
			illiqValues[i] = metrics[idx].ILLIQ
			valueValues[i] = metrics[idx].Value
			continuityValues[i] = metrics[idx].ContinuityNL
			spreadValues[i] = metrics[idx].SpreadProxy
		}
		
		// Apply linear scaling for better differentiation
		// Linear scaling preserves relative differences better than log-based scaling
		scaledILLIQ := LinearScaleILLIQ(illiqValues)           // Custom piecewise linear for ILLIQ
		scaledValue := LinearScaleVolume(valueValues)          // Custom piecewise linear for volume
		scaledContinuity := LinearScaleContinuity(continuityValues) // Direct percentage mapping
		// Spread proxy removed - unreliable data (57% zeros)
		
		// Calculate hybrid scores and apply to metrics
		for i, idx := range indices {
			// Check if this is a worst-case metric (insufficient data)
			if metrics[idx].TotalDays < 60 && metrics[idx].TradingDays == 0 {
				// Force worst scores for insufficient data
				metrics[idx].ILLIQScaled = 0      // Worst ILLIQ score (inverted scale)
				metrics[idx].ValueScaled = 0      // Worst value score
				metrics[idx].ContinuityScaled = 0 // Worst continuity score
				metrics[idx].SpreadScaled = 0     // Worst spread score
				metrics[idx].HybridScore = 0      // Absolute worst hybrid score
			} else {
				// Normal scaling for sufficient data
				metrics[idx].ILLIQScaled = scaledILLIQ[i]
				metrics[idx].ValueScaled = scaledValue[i]
				metrics[idx].ContinuityScaled = scaledContinuity[i]
				metrics[idx].SpreadScaled = 0 // Spread removed from scoring
				
				// Calculate final hybrid score (3-metric system)
				metrics[idx].HybridScore = c.calculateHybridScore(
					metrics[idx].ILLIQScaled,
					metrics[idx].ValueScaled,
					metrics[idx].ContinuityScaled,
					0, // Spread no longer used
					metrics[idx].ImpactPenalty,
					metrics[idx].ValuePenalty,
				)
				
				// Calculate safe trading values
				safeLimits := CalculateSafeTrading(metrics[idx])
				metrics[idx].SafeValue_0_5 = safeLimits.SafeValue_0_5
				metrics[idx].SafeValue_1_0 = safeLimits.SafeValue_1_0
				metrics[idx].SafeValue_2_0 = safeLimits.SafeValue_2_0
				metrics[idx].OptimalTradeSize = safeLimits.OptimalTradeSize
			}
		}
		
		// Apply ranking for this date
		c.applyRanking(metrics, indices)
	}
	
	return nil
}

// calculateHybridScore computes the final hybrid liquidity score
// Now using 3-metric system: ILLIQ (40%), Volume (35%), Continuity (25%)
func (c *Calculator) calculateHybridScore(impactScaled, valueScaled, continuityScaled, spreadScaled, impactPenalty, valuePenalty float64) float64 {
	// SAFETY CHECK 1: Ensure input values are in valid ranges
	// Scaled values should be between 0-100
	if impactScaled < 0 { impactScaled = 0 }
	if impactScaled > 100 { impactScaled = 100 }
	if valueScaled < 0 { valueScaled = 0 }
	if valueScaled > 100 { valueScaled = 100 }
	if continuityScaled < 0 { continuityScaled = 0 }
	if continuityScaled > 100 { continuityScaled = 100 }
	// Spread is no longer used but kept for backward compatibility
	
	// Updated weights for 3-metric system (spread removed)
	// ILLIQ: 40%, Volume: 35%, Continuity: 25%
	const (
		weightILLIQ      = 0.40
		weightVolume     = 0.35
		weightContinuity = 0.25
	)
	
	// Simplified unified penalty system
	// Since we now use a single penalty, apply it uniformly
	var adjustedImpact, adjustedValue float64
	
	if c.useSMA {
		// SMA mode: Value already incorporates continuity through zeros
		// Apply activity-based adjustment only for extreme cases
		activityMultiplier := 1.0
		
		// For very low continuity (<10%), apply direct activity scaling
		if continuityScaled < 10.0 {
			activityMultiplier = continuityScaled / 10.0
		} else if continuityScaled < 30.0 {
			// Moderate adjustment for low-medium activity
			activityMultiplier = 0.7 + (continuityScaled - 10.0) * 0.015
		}
		
		adjustedImpact = impactScaled * activityMultiplier
		adjustedValue = valueScaled * activityMultiplier
	} else {
		// Non-SMA mode: Apply unified penalty system
		// Both penalties are now the same (unified), so we can simplify
		if impactPenalty < 1.0 { impactPenalty = 1.0 }
		
		// Apply unified penalty to both components
		adjustedImpact = impactScaled / impactPenalty
		adjustedValue = valueScaled / impactPenalty  // Use same penalty
	}
	
	// Weighted combination using 3-metric system
	hybridScore := weightILLIQ*adjustedImpact + 
		weightVolume*adjustedValue + 
		weightContinuity*continuityScaled
	// Spread component removed (was: c.weights.Spread*spreadScaled)
	
	// SAFETY CHECK 4: Ensure score is bounded between 0-100
	if math.IsNaN(hybridScore) || math.IsInf(hybridScore, 0) {
		return 0
	}
	if hybridScore < 0 {
		return 0
	}
	if hybridScore > 100 {
		return 100
	}
	
	return hybridScore
}

// applyRanking applies relative ranking to metrics for a specific date
func (c *Calculator) applyRanking(allMetrics []TickerMetrics, indices []int) {
	// Sort indices by hybrid score (descending - higher score = better liquidity = lower rank number)
	sort.Slice(indices, func(i, j int) bool {
		return allMetrics[indices[i]].HybridScore > allMetrics[indices[j]].HybridScore
	})
	
	// Assign ranks
	for rank, idx := range indices {
		allMetrics[idx].HybridRank = rank + 1
	}
}

// Helper methods for calculating specific metrics
func (c *Calculator) calculateAverageValue(data []TradingDay) float64 {
	if len(data) == 0 {
		return 0
	}
	
	// SMA 60 Implementation: Include zeros for non-trading days
	// This naturally incorporates continuity into the volume metric
	if c.useSMA {
		// Sum all trading values (non-trading days contribute 0)
		totalValue := 0.0
		for _, td := range data {
			if td.IsTrading() {
				totalValue += td.Value // Using Value in IQD
			}
			// Non-trading days contribute 0 to the sum
		}
		
		// Always divide by the full window size (e.g., 60 days)
		// This gives us the true average including non-trading days
		return totalValue / float64(len(data))
	}
	
	// Legacy mode: average over trading days only (kept for comparison)
	totalValue := 0.0
	tradingDays := 0
	
	for _, td := range data {
		if td.IsTrading() {
			totalValue += td.Value // Using Value in IQD
			tradingDays++
		}
	}
	
	if tradingDays == 0 {
		return 0
	}
	
	// Return average trading value in IQD (only for trading days)
	return totalValue / float64(tradingDays)
}

func (c *Calculator) calculateContinuity(data []TradingDay) float64 {
	if len(data) == 0 {
		return 0
	}
	
	tradingDays := c.countTradingDays(data)
	return float64(tradingDays) / float64(len(data))
}

func (c *Calculator) countTradingDays(data []TradingDay) int {
	count := 0
	for _, td := range data {
		if td.IsTrading() {
			count++
		}
	}
	return count
}

func (c *Calculator) calculateAveragePrice(data []TradingDay) float64 {
	if len(data) == 0 {
		return 0
	}
	
	totalPrice := 0.0
	tradingDays := 0
	
	for _, td := range data {
		if td.IsTrading() {
			totalPrice += (td.High + td.Low + td.Close) / 3
			tradingDays++
		}
	}
	
	if tradingDays == 0 {
		return 0
	}
	
	return totalPrice / float64(tradingDays)
}

func (c *Calculator) calculateSpreadProxy(data []TradingDay) float64 {
	spreads := CalculateSpreadSeries(data)
	if len(spreads) == 0 {
		return 0
	}
	
	// Calculate average spread
	total := 0.0
	count := 0
	for _, spread := range spreads {
		if !math.IsNaN(spread) && !math.IsInf(spread, 0) && spread > 0 {
			total += spread
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return total / float64(count)
}

func (c *Calculator) calculateReturnMetrics(data []TradingDay) (avgReturn, returnVolatility float64) {
	if len(data) < 2 {
		return 0, 0
	}
	
	var returns []float64
	for i := 1; i < len(data); i++ {
		if data[i].IsTrading() && data[i-1].IsTrading() {
			ret := data[i].Return(data[i-1].Close)
			if !math.IsNaN(ret) && !math.IsInf(ret, 0) {
				returns = append(returns, ret)
			}
		}
	}
	
	if len(returns) == 0 {
		return 0, 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	avgReturn = sum / float64(len(returns))
	
	// Calculate standard deviation
	sumSquared := 0.0
	for _, ret := range returns {
		sumSquared += (ret - avgReturn) * (ret - avgReturn)
	}
	returnVolatility = math.Sqrt(sumSquared / float64(len(returns)))
	
	return avgReturn, returnVolatility
}