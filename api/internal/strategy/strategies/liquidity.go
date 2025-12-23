package strategies

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

// LiquidityStrategy implements a liquidity-based trading strategy
type LiquidityStrategy struct {
	*BaseStrategy
	config LiquidityConfig
}

// LiquidityConfig holds configuration for liquidity strategy
type LiquidityConfig struct {
	MinLiquidityScore float64 `json:"min_liquidity_score"`  // Minimum hybrid score
	MinVolume         float64 `json:"min_volume"`           // Minimum daily volume (IQD)
	MinContinuity     float64 `json:"min_continuity"`       // Minimum trading continuity
	LiquidityTrend    bool    `json:"liquidity_trend"`      // Consider liquidity trend
	ScoreThreshold    float64 `json:"score_threshold"`      // Score threshold for signals
	VolumeGrowthRate  float64 `json:"volume_growth_rate"`   // Minimum volume growth rate
}

// NewLiquidityStrategy creates a new liquidity-based strategy
func NewLiquidityStrategy(config LiquidityConfig, logger *slog.Logger) *LiquidityStrategy {
	base := NewBaseStrategy(
		"liquidity_v1",
		"Liquidity-Based Strategy",
		"Uses ISX liquidity metrics to identify trading opportunities",
		logger,
	)

	return &LiquidityStrategy{
		BaseStrategy: base,
		config:       config,
	}
}

// Parameters returns the strategy parameters
func (l *LiquidityStrategy) Parameters() strategy.StrategyParams {
	return strategy.StrategyParams{
		"min_liquidity_score": l.config.MinLiquidityScore,
		"min_volume":          l.config.MinVolume,
		"min_continuity":      l.config.MinContinuity,
		"liquidity_trend":     l.config.LiquidityTrend,
		"score_threshold":     l.config.ScoreThreshold,
		"volume_growth_rate":  l.config.VolumeGrowthRate,
	}
}

// Execute runs the liquidity-based strategy
func (l *LiquidityStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (strategy.Signal, error) {
	if err := l.ValidateData(ctx, data, 20); err != nil {
		return strategy.Signal{}, err
	}

	latest := data[len(data)-1]

	// Check basic liquidity filters
	if latest.Value < l.config.MinVolume {
		return l.CreateSignal(
			latest.Symbol,
			strategy.SignalHold,
			0.0,
			latest.Close,
			fmt.Sprintf("Volume %.0f below minimum %.0f", latest.Value, l.config.MinVolume),
			map[string]interface{}{
				"volume":     latest.Value,
				"min_volume": l.config.MinVolume,
			},
		), nil
	}

	// Calculate liquidity metrics (simplified - in real implementation,
	// this would integrate with the liquidity module)
	liquidityScore := l.calculateLiquidityScore(data)
	continuityScore := l.calculateContinuityScore(data)
	volumeTrend := l.calculateVolumeTrend(data)

	// Evaluate liquidity conditions
	action, strength, reasoning := l.evaluateLiquidityConditions(
		latest,
		liquidityScore,
		continuityScore,
		volumeTrend,
	)

	signal := l.CreateSignal(
		latest.Symbol,
		action,
		strength,
		latest.Close,
		reasoning,
		map[string]interface{}{
			"liquidity_score":  liquidityScore,
			"continuity_score": continuityScore,
			"volume_trend":     volumeTrend,
			"daily_volume":     latest.Value,
		},
	)

	l.LogExecution(ctx, latest.Symbol, signal)
	return signal, nil
}

// calculateLiquidityScore calculates a simplified liquidity score
func (l *LiquidityStrategy) calculateLiquidityScore(data []liquidity.TradingDay) float64 {
	if len(data) < 20 {
		return 0
	}

	recent := data[len(data)-20:]

	// Calculate average volume and value
	totalVolume := 0.0
	totalValue := 0.0
	activeDays := 0

	for _, day := range recent {
		if day.IsTrading() {
			totalVolume += day.Volume
			totalValue += day.Value
			activeDays++
		}
	}

	if activeDays == 0 {
		return 0
	}

	avgVolume := totalVolume / float64(activeDays)
	avgValue := totalValue / float64(activeDays)

	// Simple scoring based on volume and value
	volumeScore := avgVolume / 1000000   // Scale by million shares
	valueScore := avgValue / 100000000   // Scale by 100M IQD

	// Combine scores (0-1 range)
	score := (volumeScore + valueScore) / 2
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateContinuityScore calculates trading continuity
func (l *LiquidityStrategy) calculateContinuityScore(data []liquidity.TradingDay) float64 {
	if len(data) < 20 {
		return 0
	}

	recent := data[len(data)-20:]
	tradingDays := 0

	for _, day := range recent {
		if day.IsTrading() {
			tradingDays++
		}
	}

	return float64(tradingDays) / float64(len(recent))
}

// calculateVolumeTrend calculates volume trend over recent periods
func (l *LiquidityStrategy) calculateVolumeTrend(data []liquidity.TradingDay) float64 {
	if len(data) < 40 {
		return 0
	}

	// Compare recent 20 days with previous 20 days
	recent := data[len(data)-20:]
	previous := data[len(data)-40 : len(data)-20]

	recentAvg := l.averageVolume(recent)
	previousAvg := l.averageVolume(previous)

	if previousAvg == 0 {
		return 0
	}

	return (recentAvg - previousAvg) / previousAvg
}

// averageVolume calculates average volume for trading days
func (l *LiquidityStrategy) averageVolume(data []liquidity.TradingDay) float64 {
	total := 0.0
	count := 0

	for _, day := range data {
		if day.IsTrading() {
			total += day.Value
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

// evaluateLiquidityConditions determines signal based on liquidity metrics
func (l *LiquidityStrategy) evaluateLiquidityConditions(
	latest liquidity.TradingDay,
	liquidityScore, continuityScore, volumeTrend float64,
) (strategy.SignalAction, float64, string) {

	// High liquidity with positive volume trend
	if liquidityScore >= l.config.ScoreThreshold &&
		continuityScore >= l.config.MinContinuity &&
		volumeTrend >= l.config.VolumeGrowthRate {

		strength := 70.0 + (liquidityScore * 20) // 70-90 range
		return strategy.SignalBuy, strength,
			fmt.Sprintf("High liquidity: Score %.2f, Continuity %.2f, Volume trend %.2f%%",
				liquidityScore, continuityScore, volumeTrend*100)
	}

	// Good liquidity but declining trend
	if liquidityScore >= l.config.MinLiquidityScore &&
		continuityScore >= l.config.MinContinuity &&
		volumeTrend < -l.config.VolumeGrowthRate {

		return strategy.SignalSell, 60.0,
			fmt.Sprintf("Declining liquidity: Score %.2f, Volume trend %.2f%%",
				liquidityScore, volumeTrend*100)
	}

	// Moderate liquidity conditions
	if liquidityScore >= l.config.MinLiquidityScore &&
		continuityScore >= l.config.MinContinuity {

		return strategy.SignalHold, 55.0,
			fmt.Sprintf("Stable liquidity: Score %.2f, Continuity %.2f",
				liquidityScore, continuityScore)
	}

	// Poor liquidity conditions
	return strategy.SignalHold, 30.0,
		fmt.Sprintf("Poor liquidity: Score %.2f, Continuity %.2f",
			liquidityScore, continuityScore)
}

// Validate validates the strategy parameters
func (l *LiquidityStrategy) Validate(ctx context.Context, params strategy.StrategyParams) error {
	if score, ok := params["min_liquidity_score"].(float64); ok {
		if score < 0 || score > 1 {
			return strategy.NewInvalidParametersError("min_liquidity_score", score)
		}
	}

	if volume, ok := params["min_volume"].(float64); ok {
		if volume < 0 {
			return strategy.NewInvalidParametersError("min_volume", volume)
		}
	}

	return nil
}

// Backtest runs a backtest for the liquidity strategy
func (l *LiquidityStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config strategy.BacktestConfig) (strategy.BacktestResult, error) {
	if len(data) < 60 {
		return strategy.BacktestResult{}, strategy.NewInsufficientDataError(60, len(data))
	}

	// Initialize backtest state
	cash := config.InitialCash
	position := 0.0
	signals := []strategy.Signal{}
	trades := 0
	wins := 0

	// Run strategy on historical data
	for i := 40; i < len(data)-1; i++ {
		// Get data window
		window := data[:i+1]

		// Execute strategy
		signal, err := l.Execute(ctx, window)
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
		StrategyID:  l.ID(),
		Symbol:      data[0].Symbol,
		StartDate:   config.StartDate,
		EndDate:     config.EndDate,
		TotalReturn: totalReturn,
		WinRate:     winRate,
		TotalTrades: trades,
		Signals:     signals,
	}, nil
}