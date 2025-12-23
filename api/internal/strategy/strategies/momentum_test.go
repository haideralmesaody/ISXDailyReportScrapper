package strategies

import (
	"context"
	stdErrors "errors"
	"io"
	"log/slog"
	"math"
	"testing"
	"time"

	apperrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMomentumStrategy_ExecuteSignals(t *testing.T) {
	ctx := context.Background()
	strat := newTestMomentumStrategy(MomentumConfig{
		LookbackPeriod:   20,
		Threshold:        0.05,
		VolatilityFilter: false,
		VolumeFilter:     false,
		MinVolume:        0,
		MaxVolatility:    1.0,
	})

	upward := generateTrendingData(60, 1000, 0.02)
	signal, err := strat.Execute(ctx, upward)
	require.NoError(t, err)
	assert.Equal(t, strategy.SignalBuy, signal.Action)
	assert.Greater(t, signal.Strength, 0.0)

	downward := generateTrendingData(60, 1000, -0.02)
	signal, err = strat.Execute(ctx, downward)
	require.NoError(t, err)
	assert.Equal(t, strategy.SignalSell, signal.Action)

	flat := generateTrendingData(60, 1000, 0.0005)
	signal, err = strat.Execute(ctx, flat)
	require.NoError(t, err)
	assert.Equal(t, strategy.SignalHold, signal.Action)
}

func TestMomentumStrategy_Execute_InsufficientData(t *testing.T) {
	ctx := context.Background()
	strat := newTestMomentumStrategy(MomentumConfig{LookbackPeriod: 20, Threshold: 0.05})

	data := generateTrendingData(10, 1000, 0.02)
	_, err := strat.Execute(ctx, data)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.True(t, stdErrors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrTypeValidation, appErr.Type)
}

func TestMomentumStrategy_Backtest(t *testing.T) {
	ctx := context.Background()
	strat := newTestMomentumStrategy(MomentumConfig{
		LookbackPeriod:   15,
		Threshold:        0.04,
		VolatilityFilter: false,
		VolumeFilter:     false,
		MinVolume:        0,
		MaxVolatility:    1.0,
	})

	data := generateTrendingData(120, 1000, 0.01)
	config := strategy.BacktestConfig{
		StartDate:   data[0].Date,
		EndDate:     data[len(data)-1].Date,
		InitialCash: 100_000,
		Commission:  0.001,
		Slippage:    0.001,
	}

	result, err := strat.Backtest(ctx, data, config)
	require.NoError(t, err)

	assert.Equal(t, strat.ID(), result.StrategyID)
	assert.Equal(t, data[len(data)-1].Symbol, result.Symbol)
	assert.Greater(t, result.TotalReturn, -1.0) // sanity check
	assert.GreaterOrEqual(t, result.TotalTrades, 0)
	assert.NotNil(t, result.Signals)
}

func TestMomentumStrategy_Validate(t *testing.T) {
	strat := newTestMomentumStrategy(MomentumConfig{LookbackPeriod: 20, Threshold: 0.05})
	ctx := context.Background()

	t.Run("valid parameters", func(t *testing.T) {
		params := strategy.StrategyParams{
			"lookback_period": 30,
			"threshold":       0.1,
		}
		err := strat.Validate(ctx, params)
		assert.NoError(t, err)
	})

	t.Run("invalid lookback", func(t *testing.T) {
		params := strategy.StrategyParams{"lookback_period": 2}
		err := strat.Validate(ctx, params)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, stdErrors.As(err, &appErr))
		assert.Equal(t, apperrors.ErrTypeValidation, appErr.Type)
	})

	t.Run("invalid threshold", func(t *testing.T) {
		params := strategy.StrategyParams{"threshold": 0.8}
		err := strat.Validate(ctx, params)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, stdErrors.As(err, &appErr))
		assert.Equal(t, apperrors.ErrTypeValidation, appErr.Type)
	})
}

// Helpers

func newTestMomentumStrategy(cfg MomentumConfig) *MomentumStrategy {
	if cfg.LookbackPeriod == 0 {
		cfg.LookbackPeriod = 20
	}
	if cfg.Threshold == 0 {
		cfg.Threshold = 0.05
	}
	if cfg.MaxVolatility == 0 {
		cfg.MaxVolatility = 1.0
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	return NewMomentumStrategy(cfg, logger)
}

func generateTrendingData(days int, startPrice float64, dailyChange float64) []liquidity.TradingDay {
	data := make([]liquidity.TradingDay, days)
	price := startPrice

	for i := 0; i < days; i++ {
		nextPrice := price * (1 + dailyChange)
		high := math.Max(price, nextPrice) * 1.01
		low := math.Min(price, nextPrice) * 0.99
		volume := 1_000_000.0 + float64(i)*1_000
		value := nextPrice * volume

		data[i] = liquidity.TradingDay{
			Date:          time.Now().AddDate(0, 0, -days+i),
			Symbol:        "BBNI",
			Open:          price,
			High:          high,
			Low:           low,
			Close:         nextPrice,
			Volume:        volume,
			ShareVolume:   volume,
			Value:         value,
			NumTrades:     100 + i,
			TradingStatus: "ACTIVE",
		}

		price = nextPrice
	}

	return data
}
