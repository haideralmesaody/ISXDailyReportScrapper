package strategies

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

// RSI14MeanReversionEODStrategy generates EOD signals from RSI(14) crossings:
// - BUY when RSI crosses below 30
// - SELL when RSI crosses above 50
// - HOLD otherwise
//
// Confirmed data rules:
// - Use Close prices
// - Rows with Close == 0 should already be skipped by the data loader
type RSI14MeanReversionEODStrategy struct {
	*BaseStrategy
}

func NewRSI14MeanReversionEODStrategy(logger *slog.Logger) *RSI14MeanReversionEODStrategy {
	base := NewBaseStrategy(
		"rsi14_mr_eod_v1",
		"RSI(14) Mean Reversion (EOD)",
		"BUY on RSI(14) cross below 30; SELL on RSI(14) cross above 50 (EOD, long-only)",
		logger,
	)

	return &RSI14MeanReversionEODStrategy{BaseStrategy: base}
}

func (s *RSI14MeanReversionEODStrategy) Parameters() strategy.StrategyParams {
	return strategy.StrategyParams{
		"rsi_period":       14,
		"entry_threshold":  30.0,
		"exit_threshold":   50.0,
		"allow_short":      false,
		"signal_frequency": "EOD",
	}
}

func (s *RSI14MeanReversionEODStrategy) ChartPreset() strategy.ChartPreset {
	return strategy.ChartPreset{
		StrategyID:        s.ID(),
		ChartType:         "candlestick",
		Timeframe:         "MAX",
		EnabledIndicators: []string{"showVolume", "showRSI"},
		Momentum: &strategy.ChartPresetMomentum{
			RSIPeriod:     14,
			RSIOversold:   30.0,
			RSIOverbought: 50.0,
		},
	}
}

func (s *RSI14MeanReversionEODStrategy) Validate(ctx context.Context, params strategy.StrategyParams) error {
	// v1 is fixed parameters; nothing to validate yet.
	return nil
}

func (s *RSI14MeanReversionEODStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (strategy.Signal, error) {
	const (
		period         = 14
		entryThreshold = 30.0
		exitThreshold  = 50.0
	)

	// Need RSI for today and yesterday => need at least period+2 closes.
	minPoints := period + 2
	if len(data) < minPoints {
		return strategy.Signal{}, strategy.NewInsufficientDataError(minPoints, len(data))
	}

	prices := make([]float64, 0, len(data))
	for _, day := range data {
		if day.Close <= 0 {
			// The loader should skip Close==0, but keep a defensive guard.
			continue
		}
		prices = append(prices, day.Close)
	}
	if len(prices) < minPoints {
		return strategy.Signal{}, strategy.NewInsufficientDataError(minPoints, len(prices))
	}

	rsiSeries := calculateRSI(prices, period)
	if len(rsiSeries) < 2 {
		return strategy.Signal{}, fmt.Errorf("insufficient RSI series length: %d", len(rsiSeries))
	}

	prevRSI := rsiSeries[len(rsiSeries)-2]
	currRSI := rsiSeries[len(rsiSeries)-1]

	latest := data[len(data)-1]

	action := strategy.SignalHold
	reasoning := fmt.Sprintf("No cross: RSI(14) prev=%.2f curr=%.2f", prevRSI, currRSI)
	strength := 50.0

	if prevRSI >= entryThreshold && currRSI < entryThreshold {
		action = strategy.SignalBuy
		strength = math.Min(100.0, 50.0+(entryThreshold-currRSI)*2.0)
		reasoning = fmt.Sprintf("BUY: RSI(14) crossed below %.0f (prev=%.2f, curr=%.2f)", entryThreshold, prevRSI, currRSI)
	} else if prevRSI <= exitThreshold && currRSI > exitThreshold {
		action = strategy.SignalSell
		strength = math.Min(100.0, 50.0+(currRSI-exitThreshold)*2.0)
		reasoning = fmt.Sprintf("SELL: RSI(14) crossed above %.0f (prev=%.2f, curr=%.2f)", exitThreshold, prevRSI, currRSI)
	}

	signal := s.CreateSignal(
		latest.Symbol,
		action,
		strength,
		latest.Close,
		reasoning,
		map[string]interface{}{
			"rsi_period":       period,
			"rsi":              currRSI,
			"prev_rsi":         prevRSI,
			"entry_threshold":  entryThreshold,
			"exit_threshold":   exitThreshold,
			"signal_frequency": "EOD",
		},
	)

	// For EOD signals, use the last data point date as the signal time.
	signal.Timestamp = latest.Date
	signal.ValidUntil = latest.Date.Add(24 * time.Hour)

	s.LogExecution(ctx, latest.Symbol, signal)
	return signal, nil
}

func (s *RSI14MeanReversionEODStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config strategy.BacktestConfig) (strategy.BacktestResult, error) {
	// v1: keep backtest simple by reusing Execute over expanding window (same pattern as other strategies).
	const period = 14
	minPoints := period + 2
	if len(data) < minPoints*2 {
		return strategy.BacktestResult{}, strategy.NewInsufficientDataError(minPoints*2, len(data))
	}

	cash := config.InitialCash
	position := 0.0
	signals := []strategy.Signal{}
	trades := 0
	wins := 0

	for i := minPoints; i < len(data)-1; i++ {
		window := data[:i+1]
		signal, err := s.Execute(ctx, window)
		if err != nil {
			continue
		}
		if signal.Action == strategy.SignalHold {
			continue
		}

		signals = append(signals, signal)
		nextPrice := data[i+1].Close
		if nextPrice <= 0 {
			continue
		}

		if signal.Action == strategy.SignalBuy && position == 0 {
			position = cash / nextPrice
			cash = 0
			trades++
		} else if signal.Action == strategy.SignalSell && position > 0 {
			cash = position * nextPrice
			if cash > config.InitialCash {
				wins++
			}
			position = 0
			trades++
		}
	}

	finalValue := cash
	if position > 0 {
		finalValue += position * data[len(data)-1].Close
	}

	totalReturn := 0.0
	if config.InitialCash > 0 {
		totalReturn = (finalValue - config.InitialCash) / config.InitialCash
	}

	winRate := 0.0
	if trades > 0 {
		winRate = float64(wins) / float64(trades)
	}

	return strategy.BacktestResult{
		StrategyID:  s.ID(),
		Symbol:      data[0].Symbol,
		StartDate:   config.StartDate,
		EndDate:     config.EndDate,
		TotalReturn: totalReturn,
		WinRate:     winRate,
		TotalTrades: trades,
		Signals:     signals,
	}, nil
}

// calculateRSI computes RSI using Wilder smoothing.
// Returns a slice aligned to the last (len(prices)-1) changes, starting at index `period`.
func calculateRSI(prices []float64, period int) []float64 {
	if period <= 0 || len(prices) < period+2 {
		return nil
	}

	changes := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		changes[i-1] = prices[i] - prices[i-1]
	}

	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))
	for i, change := range changes {
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	rsi := make([]float64, len(changes))
	for i := period; i < len(changes); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
			continue
		}
		rs := avgGain / avgLoss
		rsi[i] = 100 - (100 / (1 + rs))
	}

	return rsi[period:]
}
