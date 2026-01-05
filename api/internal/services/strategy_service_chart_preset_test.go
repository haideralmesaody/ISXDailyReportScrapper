package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	apierrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
	"github.com/isxcli/isxcli/internal/strategy/strategies"
	"github.com/stretchr/testify/require"
)

type noopStrategy struct{}

func (n *noopStrategy) ID() string          { return "noop_v1" }
func (n *noopStrategy) Name() string        { return "noop" }
func (n *noopStrategy) Description() string { return "noop" }
func (n *noopStrategy) Parameters() strategy.StrategyParams {
	return strategy.StrategyParams{}
}
func (n *noopStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (strategy.Signal, error) {
	return strategy.Signal{}, nil
}
func (n *noopStrategy) Validate(ctx context.Context, params strategy.StrategyParams) error { return nil }
func (n *noopStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config strategy.BacktestConfig) (strategy.BacktestResult, error) {
	return strategy.BacktestResult{
		StrategyID: n.ID(),
		Symbol:     "TEST",
		StartDate:  time.Time{},
		EndDate:    time.Time{},
	}, nil
}

func TestStrategyService_GetStrategyChartPreset(t *testing.T) {
	t.Run("returns chart preset when available", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		mgr := strategy.NewManager()
		require.NoError(t, mgr.RegisterStrategy(strategies.NewRSI14MeanReversionEODStrategy(logger)))

		svc := NewStrategyService(mgr, nil, logger)

		got, err := svc.GetStrategyChartPreset(context.Background(), "rsi14_mr_eod_v1")
		require.NoError(t, err)
		require.Equal(t, "rsi14_mr_eod_v1", got.StrategyID)
		require.Equal(t, "candlestick", got.ChartType)
		require.Equal(t, "MAX", got.Timeframe)
		require.Contains(t, got.EnabledIndicators, "showRSI")
		require.NotNil(t, got.Momentum)
		require.Equal(t, 14, got.Momentum.RSIPeriod)
		require.Equal(t, 30.0, got.Momentum.RSIOversold)
		require.Equal(t, 50.0, got.Momentum.RSIOverbought)
	})

	t.Run("returns not found when preset not supported", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		mgr := strategy.NewManager()
		require.NoError(t, mgr.RegisterStrategy(&noopStrategy{}))

		svc := NewStrategyService(mgr, nil, logger)

		_, err := svc.GetStrategyChartPreset(context.Background(), "noop_v1")
		require.Error(t, err)

		var apiErr *apierrors.APIError
		require.True(t, errors.As(err, &apiErr))
		require.Equal(t, 404, apiErr.StatusCode)
	})
}

