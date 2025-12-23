package strategies

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

// MeanReversionStrategy implements a mean reversion trading strategy
type MeanReversionStrategy struct {
	*BaseStrategy
	config MeanReversionConfig
}

// MeanReversionConfig holds configuration for mean reversion strategy
type MeanReversionConfig struct {
	LookbackPeriod    int     `json:"lookback_period"`     // Period for moving average
	StdDevThreshold   float64 `json:"stddev_threshold"`    // Standard deviation threshold
	MinLiquidityScore float64 `json:"min_liquidity_score"` // Minimum liquidity requirement
	OverboughtLevel   float64 `json:"overbought_level"`    // RSI overbought level
	OversoldLevel     float64 `json:"oversold_level"`      // RSI oversold level
}

// NewMeanReversionStrategy creates a new mean reversion strategy
func NewMeanReversionStrategy(config MeanReversionConfig, logger *slog.Logger) *MeanReversionStrategy {
	base := NewBaseStrategy(
		"mean_reversion_v1",
		"Mean Reversion Strategy",
		"Identifies mean reversion opportunities using Bollinger Bands and RSI",
		logger,
	)

	return &MeanReversionStrategy{
		BaseStrategy: base,
		config:       config,
	}
}

// Parameters returns the strategy parameters
func (mr *MeanReversionStrategy) Parameters() strategy.StrategyParams {
	return strategy.StrategyParams{
		"lookback_period":     mr.config.LookbackPeriod,
		"stddev_threshold":    mr.config.StdDevThreshold,
		"min_liquidity_score": mr.config.MinLiquidityScore,
		"overbought_level":    mr.config.OverboughtLevel,
		"oversold_level":      mr.config.OversoldLevel,
	}
}

// Execute runs the mean reversion strategy
func (mr *MeanReversionStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (strategy.Signal, error) {
	minPoints := mr.config.LookbackPeriod + 20 // Need extra points for RSI
	if err := mr.ValidateData(ctx, data, minPoints); err != nil {
		return strategy.Signal{}, err
	}

	latest := data[len(data)-1]

	// Calculate Bollinger Bands
	prices := mr.extractClosePrices(data)
	sma := mr.CalculateSimpleMovingAverage(prices, mr.config.LookbackPeriod)
	stdDev := mr.calculateRollingStdDev(prices, mr.config.LookbackPeriod)

	if len(sma) == 0 || len(stdDev) == 0 {
		return strategy.Signal{}, fmt.Errorf("insufficient data for Bollinger Bands calculation")
	}

	currentSMA := sma[len(sma)-1]
	currentStdDev := stdDev[len(stdDev)-1]
	upperBand := currentSMA + (mr.config.StdDevThreshold * currentStdDev)
	lowerBand := currentSMA - (mr.config.StdDevThreshold * currentStdDev)

	// Calculate RSI
	rsi := mr.calculateRSI(prices, 14)
	currentRSI := 50.0 // Default neutral RSI
	if len(rsi) > 0 {
		currentRSI = rsi[len(rsi)-1]
	}

	// Calculate Bollinger Band position
	bandPosition := (latest.Close - lowerBand) / (upperBand - lowerBand)

	// Generate signal
	action, strength, reasoning := mr.evaluateMeanReversion(
		latest.Close,
		currentSMA,
		upperBand,
		lowerBand,
		currentRSI,
		bandPosition,
	)

	signal := mr.CreateSignal(
		latest.Symbol,
		action,
		strength,
		latest.Close,
		reasoning,
		map[string]interface{}{
			"sma":           currentSMA,
			"upper_band":    upperBand,
			"lower_band":    lowerBand,
			"rsi":           currentRSI,
			"band_position": bandPosition,
		},
	)

	mr.LogExecution(ctx, latest.Symbol, signal)
	return signal, nil
}

// extractClosePrices extracts closing prices from trading data
func (mr *MeanReversionStrategy) extractClosePrices(data []liquidity.TradingDay) []float64 {
	prices := make([]float64, len(data))
	for i, day := range data {
		prices[i] = day.Close
	}
	return prices
}

// calculateRollingStdDev calculates rolling standard deviation
func (mr *MeanReversionStrategy) calculateRollingStdDev(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	stdDevs := make([]float64, len(prices))

	for i := period - 1; i < len(prices); i++ {
		window := prices[i-period+1 : i+1]
		stdDevs[i] = mr.CalculateStandardDeviation(window)
	}

	return stdDevs
}

// calculateRSI calculates Relative Strength Index
func (mr *MeanReversionStrategy) calculateRSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 {
		return nil
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
		} else {
			losses[i] = -change
		}
	}

	// Calculate RSI
	rsi := make([]float64, len(changes))

	// Initial average gain and loss
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI values
	for i := period; i < len(changes); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi[period:]
}

// evaluateMeanReversion determines signal based on mean reversion indicators
func (mr *MeanReversionStrategy) evaluateMeanReversion(
	currentPrice, sma, upperBand, lowerBand, rsi, bandPosition float64,
) (strategy.SignalAction, float64, string) {

	// Oversold conditions (buy signal)
	if currentPrice <= lowerBand && rsi <= mr.config.OversoldLevel {
		strength := math.Min(90.0, (mr.config.OversoldLevel-rsi)*2+50)
		return strategy.SignalBuy, strength,
			fmt.Sprintf("Oversold: Price %.2f below lower band %.2f, RSI %.1f",
				currentPrice, lowerBand, rsi)
	}

	// Moderately oversold
	if bandPosition < 0.2 && rsi < 40 {
		return strategy.SignalBuy, 65.0,
			fmt.Sprintf("Moderately oversold: Band position %.2f, RSI %.1f",
				bandPosition, rsi)
	}

	// Overbought conditions (sell signal)
	if currentPrice >= upperBand && rsi >= mr.config.OverboughtLevel {
		strength := math.Min(90.0, (rsi-mr.config.OverboughtLevel)*2+50)
		return strategy.SignalSell, strength,
			fmt.Sprintf("Overbought: Price %.2f above upper band %.2f, RSI %.1f",
				currentPrice, upperBand, rsi)
	}

	// Moderately overbought
	if bandPosition > 0.8 && rsi > 60 {
		return strategy.SignalSell, 65.0,
			fmt.Sprintf("Moderately overbought: Band position %.2f, RSI %.1f",
				bandPosition, rsi)
	}

	// Neutral conditions
	return strategy.SignalHold, 50.0,
		fmt.Sprintf("Neutral: Price %.2f, RSI %.1f, Band position %.2f",
			currentPrice, rsi, bandPosition)
}

// Validate validates the strategy parameters
func (mr *MeanReversionStrategy) Validate(ctx context.Context, params strategy.StrategyParams) error {
	if lookback, ok := params["lookback_period"].(int); ok {
		if lookback < 10 || lookback > 50 {
			return strategy.NewInvalidParametersError("lookback_period", lookback)
		}
	}

	if threshold, ok := params["stddev_threshold"].(float64); ok {
		if threshold < 1.0 || threshold > 3.0 {
			return strategy.NewInvalidParametersError("stddev_threshold", threshold)
		}
	}

	return nil
}

// Backtest runs a backtest for the mean reversion strategy
func (mr *MeanReversionStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config strategy.BacktestConfig) (strategy.BacktestResult, error) {
	minPoints := mr.config.LookbackPeriod + 20
	if len(data) < minPoints*2 {
		return strategy.BacktestResult{}, strategy.NewInsufficientDataError(minPoints*2, len(data))
	}

	// Initialize backtest state
	cash := config.InitialCash
	position := 0.0
	signals := []strategy.Signal{}
	trades := 0
	wins := 0

	// Run strategy on historical data
	for i := minPoints; i < len(data)-1; i++ {
		// Get data window
		window := data[:i+1]

		// Execute strategy
		signal, err := mr.Execute(ctx, window)
		if err != nil {
			continue
		}

		// Process signal
		if signal.Action != strategy.SignalHold {
			signals = append(signals, signal)

			// Simple position management
			currentPrice := data[i+1].Close // Next day's price

			if signal.Action == strategy.SignalBuy && position == 0 {
				position = cash / currentPrice
				cash = 0
				trades++
			} else if signal.Action == strategy.SignalSell && position > 0 {
				cash = position * currentPrice
				if cash > config.InitialCash {
					wins++
				}
				position = 0
				trades++
			}
		}
	}

	// Calculate final value
	finalValue := cash
	if position > 0 {
		finalValue += position * data[len(data)-1].Close
	}

	// Calculate metrics
	totalReturn := (finalValue - config.InitialCash) / config.InitialCash
	winRate := 0.0
	if trades > 0 {
		winRate = float64(wins) / float64(trades)
	}

	return strategy.BacktestResult{
		StrategyID:  mr.ID(),
		Symbol:      data[0].Symbol,
		StartDate:   config.StartDate,
		EndDate:     config.EndDate,
		TotalReturn: totalReturn,
		WinRate:     winRate,
		TotalTrades: trades,
		Signals:     signals,
	}, nil
}