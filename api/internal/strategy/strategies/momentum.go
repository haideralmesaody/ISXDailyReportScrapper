package strategies

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

// MomentumStrategy implements a momentum-based trading strategy
type MomentumStrategy struct {
	*BaseStrategy
	config MomentumConfig
}

// MomentumConfig holds configuration for momentum strategy
type MomentumConfig struct {
	LookbackPeriod   int     `json:"lookback_period"`    // Period to calculate momentum
	Threshold        float64 `json:"threshold"`          // Momentum threshold for signals
	VolatilityFilter bool    `json:"volatility_filter"`  // Whether to filter by volatility
	VolumeFilter     bool    `json:"volume_filter"`      // Whether to filter by volume
	MinVolume        float64 `json:"min_volume"`         // Minimum daily volume
	MaxVolatility    float64 `json:"max_volatility"`     // Maximum volatility threshold
}

// NewMomentumStrategy creates a new momentum strategy
func NewMomentumStrategy(config MomentumConfig, logger *slog.Logger) *MomentumStrategy {
	base := NewBaseStrategy(
		"momentum_v1",
		"Momentum Strategy",
		"Identifies strong price momentum trends for entry signals",
		logger,
	)

	return &MomentumStrategy{
		BaseStrategy: base,
		config:       config,
	}
}

// Parameters returns the strategy parameters
func (m *MomentumStrategy) Parameters() strategy.StrategyParams {
	return strategy.StrategyParams{
		"lookback_period":   m.config.LookbackPeriod,
		"threshold":         m.config.Threshold,
		"volatility_filter": m.config.VolatilityFilter,
		"volume_filter":     m.config.VolumeFilter,
		"min_volume":        m.config.MinVolume,
		"max_volatility":    m.config.MaxVolatility,
	}
}

// Execute runs the momentum strategy
func (m *MomentumStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (strategy.Signal, error) {
	// Validate input data
	minPoints := m.config.LookbackPeriod + 10 // Need extra points for calculations
	if err := m.ValidateData(ctx, data, minPoints); err != nil {
		return strategy.Signal{}, err
	}

	// Get the latest data point
	latest := data[len(data)-1]

	// Calculate momentum
	momentum := m.calculateMomentum(data)

	// Apply filters
	if m.config.VolumeFilter && latest.Value < m.config.MinVolume {
		return m.CreateSignal(
			latest.Symbol,
			strategy.SignalHold,
			0.0,
			latest.Close,
			"Insufficient volume for momentum strategy",
			map[string]interface{}{
				"momentum": momentum,
				"volume":   latest.Value,
			},
		), nil
	}

	if m.config.VolatilityFilter {
		volatility := m.calculateVolatility(data)
		if volatility > m.config.MaxVolatility {
			return m.CreateSignal(
				latest.Symbol,
				strategy.SignalHold,
				0.0,
				latest.Close,
				"High volatility - avoiding momentum signal",
				map[string]interface{}{
					"momentum":   momentum,
					"volatility": volatility,
				},
			), nil
		}
	}

	// Generate signal based on momentum
	action, strength, reasoning := m.evaluateMomentum(momentum)

	signal := m.CreateSignal(
		latest.Symbol,
		action,
		strength,
		latest.Close,
		reasoning,
		map[string]interface{}{
			"momentum":        momentum,
			"threshold":       m.config.Threshold,
			"lookback_period": m.config.LookbackPeriod,
		},
	)

	m.LogExecution(ctx, latest.Symbol, signal)
	return signal, nil
}

// calculateMomentum calculates price momentum over the lookback period
func (m *MomentumStrategy) calculateMomentum(data []liquidity.TradingDay) float64 {
	if len(data) < m.config.LookbackPeriod+1 {
		return 0
	}

	current := data[len(data)-1].Close
	past := data[len(data)-1-m.config.LookbackPeriod].Close

	if past == 0 {
		return 0
	}

	return (current - past) / past
}

// calculateVolatility calculates price volatility over the lookback period
func (m *MomentumStrategy) calculateVolatility(data []liquidity.TradingDay) float64 {
	if len(data) < m.config.LookbackPeriod {
		return 0
	}

	// Get recent data
	recentData := data[len(data)-m.config.LookbackPeriod:]
	returns := m.CalculateReturns(recentData)

	return m.CalculateStandardDeviation(returns)
}

// evaluateMomentum determines the signal action and strength
func (m *MomentumStrategy) evaluateMomentum(momentum float64) (strategy.SignalAction, float64, string) {
	threshold := m.config.Threshold

	// Strong positive momentum
	if momentum >= threshold*2 {
		return strategy.SignalBuy, 90.0, fmt.Sprintf("Strong positive momentum: %.2f%%", momentum*100)
	}

	// Moderate positive momentum
	if momentum >= threshold {
		return strategy.SignalBuy, 70.0, fmt.Sprintf("Positive momentum: %.2f%%", momentum*100)
	}

	// Strong negative momentum
	if momentum <= -threshold*2 {
		return strategy.SignalSell, 90.0, fmt.Sprintf("Strong negative momentum: %.2f%%", momentum*100)
	}

	// Moderate negative momentum
	if momentum <= -threshold {
		return strategy.SignalSell, 70.0, fmt.Sprintf("Negative momentum: %.2f%%", momentum*100)
	}

	// Neutral momentum
	return strategy.SignalHold, 50.0, fmt.Sprintf("Neutral momentum: %.2f%%", momentum*100)
}

// Validate validates the strategy parameters
func (m *MomentumStrategy) Validate(ctx context.Context, params strategy.StrategyParams) error {
	// Validate lookback period
	if lookback, ok := params["lookback_period"].(int); ok {
		if lookback < 5 || lookback > 100 {
			return strategy.NewInvalidParametersError("lookback_period", lookback)
		}
	}

	// Validate threshold
	if threshold, ok := params["threshold"].(float64); ok {
		if threshold < 0.01 || threshold > 0.5 {
			return strategy.NewInvalidParametersError("threshold", threshold)
		}
	}

	return nil
}

// Backtest runs a backtest for the momentum strategy
func (m *MomentumStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config strategy.BacktestConfig) (strategy.BacktestResult, error) {
	if len(data) < m.config.LookbackPeriod*2 {
		return strategy.BacktestResult{}, strategy.NewInsufficientDataError(m.config.LookbackPeriod*2, len(data))
	}

	// Initialize backtest state
	cash := config.InitialCash
	position := 0.0
	signals := []strategy.Signal{}
	trades := 0
	wins := 0

	// Run strategy on historical data
	for i := m.config.LookbackPeriod; i < len(data)-1; i++ {
		// Get data window
		window := data[:i+1]

		// Execute strategy
		signal, err := m.Execute(ctx, window)
		if err != nil {
			continue
		}

		// Process signal
		if signal.Action != strategy.SignalHold {
			signals = append(signals, signal)

			// Simple position management
			currentPrice := data[i+1].Close // Next day's price (realistic execution)

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
		StrategyID:  m.ID(),
		Symbol:      data[0].Symbol,
		StartDate:   config.StartDate,
		EndDate:     config.EndDate,
		TotalReturn: totalReturn,
		WinRate:     winRate,
		TotalTrades: trades,
		Signals:     signals,
	}, nil
}