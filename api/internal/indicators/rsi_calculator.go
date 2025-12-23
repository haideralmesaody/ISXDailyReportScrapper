package indicators

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/isxcli/isxcli/internal/models"
)

// RSICalculator calculates RSI (Relative Strength Index) values
type RSICalculator struct {
	period       int
	smoothing    float64
	minDataPoints int
}

// NewRSICalculator creates a new RSI calculator instance
func NewRSICalculator(period int, smoothing float64, minDataPoints int) *RSICalculator {
	// Validate parameters
	if period < 1 || period > 100 {
		period = 14 // Default to 14
	}
	if smoothing < 0 || smoothing > 1 {
		smoothing = 0.0 // Standard RSI (no smoothing)
	}
	if minDataPoints < period+1 {
		minDataPoints = period + 1
	}

	return &RSICalculator{
		period:        period,
		smoothing:     smoothing,
		minDataPoints: minDataPoints,
	}
}

// PriceDataPoint represents a single price data point with timestamp
type PriceDataPoint struct {
	Date   time.Time `json:"date"`
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Volume int64     `json:"volume"`
}

// RSIResult represents the result of RSI calculation
type RSIResult struct {
	Symbol        string    `json:"symbol"`
	RSIValue      float64   `json:"rsi_value"`
	RSIPeriod     int       `json:"rsi_period"`
	SignalStrength string  `json:"signal_strength"`
	TrendDirection string  `json:"trend_direction"`
	LastPrice     float64   `json:"last_price"`
	PriceChange   float64   `json:"price_change"`
	PercentChange float64   `json:"percent_change"`
	DataPoints    int       `json:"data_points"`
	CalculatedAt  time.Time `json:"calculated_at"`
}

// CalculateRSI calculates RSI for a single symbol using price history
func (r *RSICalculator) CalculateRSI(symbol string, priceData []PriceDataPoint) (*RSIResult, error) {
	if len(priceData) < r.minDataPoints {
		return nil, fmt.Errorf("insufficient data points: need %d, got %d", r.minDataPoints, len(priceData))
	}

	// Sort data by date
	sort.Slice(priceData, func(i, j int) bool {
		return priceData[i].Date.Before(priceData[j].Date)
	})

	// Extract prices
	prices := make([]float64, len(priceData))
	for i, point := range priceData {
		prices[i] = point.Price
	}

	// Calculate RSI
	rsiValue, err := r.calculateRSIValue(prices)
	if err != nil {
		return nil, fmt.Errorf("calculate RSI value: %w", err)
	}

	// Determine signal strength and trend
	signalStrength := models.GetSignalStrength(rsiValue, 70, "OVERBOUGHT")
	if rsiValue <= 30 {
		signalStrength = models.GetSignalStrength(rsiValue, 30, "OVERSOLD")
	}

	trendDirection := r.calculateTrendDirection(prices)

	// Calculate price change
	var priceChange, percentChange float64
	if len(prices) >= 2 {
		priceChange = prices[len(prices)-1] - prices[len(prices)-2]
		percentChange = (priceChange / prices[len(prices)-2]) * 100
	}

	result := &RSIResult{
		Symbol:         symbol,
		RSIValue:       rsiValue,
		RSIPeriod:      r.period,
		SignalStrength: signalStrength,
		TrendDirection: trendDirection,
		LastPrice:      prices[len(prices)-1],
		PriceChange:    priceChange,
		PercentChange:  percentChange,
		DataPoints:     len(priceData),
		CalculatedAt:   time.Now(),
	}

	return result, nil
}

// calculateRSIValue calculates the RSI value using the standard formula
func (r *RSICalculator) calculateRSIValue(prices []float64) (float64, error) {
	if len(prices) < r.period+1 {
		return 0, fmt.Errorf("need at least %d data points for %d-period RSI", r.period+1, r.period)
	}

	// Calculate price changes
	changes := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		changes[i-1] = prices[i] - prices[i-1]
	}

	// Separate gains and losses
	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))
	for i, change := range changes {
		if change > 0 {
			gains[i] = change
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = -change
		}
	}

	// Calculate average gains and losses
	var avgGain, avgLoss float64

	if r.smoothing == 0.0 {
		// Simple moving average (standard RSI)
		avgGain = r.calculateSMA(gains[:r.period])
		avgLoss = r.calculateSMA(losses[:r.period])
	} else {
		// Exponential moving average (smoothed RSI)
		avgGain = r.calculateEMA(gains[:r.period], r.smoothing)
		avgLoss = r.calculateEMA(losses[:r.period], r.smoothing)
	}

	// Calculate RSI for remaining periods using smoothed averages
	for i := r.period; i < len(changes); i++ {
		if r.smoothing == 0.0 {
			// Simple moving average sliding window
			start := i - r.period + 1
			end := i + 1
			avgGain = r.calculateSMA(gains[start:end])
			avgLoss = r.calculateSMA(losses[start:end])
		} else {
			// Exponential moving average
			avgGain = (avgGain*float64(r.period-1) + gains[i]) / float64(r.period)
			avgLoss = (avgLoss*float64(r.period-1) + losses[i]) / float64(r.period)
		}
	}

	// Calculate RSI
	if avgLoss == 0 {
		return 100.0, nil // Perfectly trending up
	}

	if avgGain == 0 {
		return 0.0, nil // Perfectly trending down
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi, nil
}

// calculateSMA calculates Simple Moving Average
func (r *RSICalculator) calculateSMA(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

// calculateEMA calculates Exponential Moving Average
func (r *RSICalculator) calculateEMA(values []float64, smoothing float64) float64 {
	if len(values) == 0 {
		return 0
	}

	if len(values) == 1 {
		return values[0]
	}

	// Calculate multiplier
	multiplier := smoothing / (1.0 + float64(len(values)-1))

	// Start with SMA
	ema := r.calculateSMA(values)

	// Apply EMA formula
	for i := 1; i < len(values); i++ {
		ema = (values[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// calculateTrendDirection determines the trend direction based on recent RSI values
func (r *RSICalculator) calculateTrendDirection(prices []float64) string {
	if len(prices) < r.period*2+1 {
		return "STABLE"
	}

	// Calculate RSI for recent period and previous period
	recentPrices := prices[len(prices)-r.period:]
	previousPrices := prices[len(prices)-r.period*2 : len(prices)-r.period]

	recentRSI, _ := r.calculateRSIValue(recentPrices)
	previousRSI, _ := r.calculateRSIValue(previousPrices)

	return models.GetTrendDirection(recentRSI, previousRSI)
}

// CalculateMultipleRSI calculates RSI for multiple symbols
func (r *RSICalculator) CalculateMultipleRSI(priceDataBySymbol map[string][]PriceDataPoint) (map[string]*RSIResult, error) {
	results := make(map[string]*RSIResult)

	for symbol, priceData := range priceDataBySymbol {
		result, err := r.CalculateRSI(symbol, priceData)
		if err != nil {
			// Log error but continue with other symbols
			continue
		}
		results[symbol] = result
	}

	return results, nil
}

// GetOverboughtOversoldSignals identifies overbought and oversold signals
func (r *RSICalculator) GetOverboughtOversoldSignals(results map[string]*RSIResult, overboughtThreshold, oversoldThreshold int) ([]*RSIResult, []*RSIResult) {
	var overbought, oversold []*RSIResult

	for _, result := range results {
		if result.RSIValue >= float64(overboughtThreshold) {
			overbought = append(overbought, result)
		} else if result.RSIValue <= float64(oversoldThreshold) {
			oversold = append(oversold, result)
		}
	}

	return overbought, oversold
}

// ValidateRSIConfiguration validates RSI calculation parameters
func ValidateRSIConfiguration(period int, threshold int, minDataPoints int) error {
	if period < 1 || period > 100 {
		return fmt.Errorf("RSI period must be between 1 and 100")
	}

	if threshold < 0 || threshold > 100 {
		return fmt.Errorf("RSI threshold must be between 0 and 100")
	}

	if minDataPoints < period+1 {
		return fmt.Errorf("minimum data points must be at least period + 1")
	}

	return nil
}

// GetOptimalRSIPeriod suggests the optimal RSI period based on data characteristics
func GetOptimalRSIPeriod(dataPoints int, volatility float64) int {
	if dataPoints < 30 {
		return 7 // Shorter period for limited data
	}

	if dataPoints < 100 {
		return 14 // Standard period for moderate data
	}

	// For longer datasets, adjust based on volatility
	if volatility > 0.05 { // High volatility (>5%)
		return 21 // Longer period to smooth out noise
	} else if volatility > 0.02 { // Medium volatility (2-5%)
		return 14 // Standard period
	} else { // Low volatility (<2%)
		return 9 // Shorter period for sensitivity
	}
}

// CalculateVolatility measures price volatility as standard deviation of returns
func CalculateVolatility(priceData []PriceDataPoint) float64 {
	if len(priceData) < 2 {
		return 0
	}

	// Sort data by date
	sort.Slice(priceData, func(i, j int) bool {
		return priceData[i].Date.Before(priceData[j].Date)
	})

	// Calculate returns
	returns := make([]float64, len(priceData)-1)
	for i := 1; i < len(priceData); i++ {
		if priceData[i-1].Price != 0 {
			returns[i-1] = (priceData[i].Price - priceData[i-1].Price) / priceData[i-1].Price
		}
	}

	// Calculate mean return
	mean := 0.0
	for _, ret := range returns {
		mean += ret
	}
	mean /= float64(len(returns))

	// Calculate variance
	variance := 0.0
	for _, ret := range returns {
		diff := ret - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	// Return standard deviation
	return math.Sqrt(variance)
}

// FilterByTradingStatus filters price data to include only actively traded stocks
func (r *RSICalculator) FilterByTradingStatus(priceData []PriceDataPoint, tradingStatus map[string]bool) []PriceDataPoint {
	var filtered []PriceDataPoint

	for _, point := range priceData {
		if isActive, exists := tradingStatus[point.Symbol]; exists && isActive {
			filtered = append(filtered, point)
		}
	}

	return filtered
}

// GenerateRSIStatistics generates statistics about RSI calculations
func (r *RSICalculator) GenerateRSIStatistics(results map[string]*RSIResult) map[string]interface{} {
	if len(results) == 0 {
		return map[string]interface{}{
			"total_symbols": 0,
		}
	}

	var rsiValues []float64
	var priceChanges []float64
	overboughtCount := 0
	oversoldCount := 0

	for _, result := range results {
		rsiValues = append(rsiValues, result.RSIValue)
		priceChanges = append(priceChanges, result.PercentChange)

		if result.RSIValue >= 70 {
			overboughtCount++
		} else if result.RSIValue <= 30 {
			oversoldCount++
		}
	}

	// Calculate statistics
	meanRSI := r.calculateSMA(rsiValues)
	meanPriceChange := r.calculateSMA(priceChanges)

	// Calculate RSI distribution
	buckets := make(map[string]int)
	for _, rsi := range rsiValues {
		bucket := r.getRSIBucket(rsi)
		buckets[bucket]++
	}

	return map[string]interface{}{
		"total_symbols":        len(results),
		"mean_rsi":             meanRSI,
		"mean_price_change":    meanPriceChange,
		"overbought_count":     overboughtCount,
		"oversold_count":       oversoldCount,
		"overbought_percent":   float64(overboughtCount) / float64(len(results)) * 100,
		"oversold_percent":     float64(oversoldCount) / float64(len(results)) * 100,
		"rsi_distribution":     buckets,
		"rsi_period":           r.period,
		"calculated_at":        time.Now(),
	}
}

// getRSIBucket categorizes RSI values into buckets
func (r *RSICalculator) getRSIBucket(rsi float64) string {
	switch {
	case rsi >= 80:
		return "80-100"
	case rsi >= 70:
		return "70-79"
	case rsi >= 60:
		return "60-69"
	case rsi >= 50:
		return "50-59"
	case rsi >= 40:
		return "40-49"
	case rsi >= 30:
		return "30-39"
	case rsi >= 20:
		return "20-29"
	default:
		return "0-19"
	}
}