package services

import (
	"context"
	"testing"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

type mockSignalStrategy struct{}

func (m mockSignalStrategy) ID() string          { return "mock_v1" }
func (m mockSignalStrategy) Name() string        { return "mock" }
func (m mockSignalStrategy) Description() string { return "mock strategy" }
func (m mockSignalStrategy) Parameters() strategy.StrategyParams {
	return strategy.StrategyParams{}
}
func (m mockSignalStrategy) Validate(ctx context.Context, params strategy.StrategyParams) error {
	return nil
}
func (m mockSignalStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config strategy.BacktestConfig) (strategy.BacktestResult, error) {
	return strategy.BacktestResult{}, nil
}
func (m mockSignalStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (strategy.Signal, error) {
	if len(data) == 0 {
		return strategy.Signal{Action: strategy.SignalHold}, nil
	}
	last := data[len(data)-1]
	switch last.Close {
	case 1:
		return strategy.Signal{Action: strategy.SignalBuy}, nil
	case 2:
		return strategy.Signal{Action: strategy.SignalSell}, nil
	default:
		return strategy.Signal{Action: strategy.SignalHold}, nil
	}
}

func TestRunBacktestWindow_ExtendMatchesFull(t *testing.T) {
	t.Parallel()

	manager := strategy.NewManager()
	if err := manager.RegisterStrategy(mockSignalStrategy{}); err != nil {
		t.Fatalf("register: %v", err)
	}
	svc := &StrategyService{manager: manager}

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	all := []liquidity.TradingDay{
		{Date: base.AddDate(0, 0, 0), Symbol: "AAA", Close: 10},
		{Date: base.AddDate(0, 0, 1), Symbol: "AAA", Close: 1},   // BUY signal, exec day2
		{Date: base.AddDate(0, 0, 2), Symbol: "AAA", Close: 100}, // BUY exec price
		{Date: base.AddDate(0, 0, 3), Symbol: "AAA", Close: 2},   // SELL signal, exec day4
		{Date: base.AddDate(0, 0, 4), Symbol: "AAA", Close: 110}, // SELL exec price
		{Date: base.AddDate(0, 0, 5), Symbol: "AAA", Close: 1},   // BUY signal, exec day6
		{Date: base.AddDate(0, 0, 6), Symbol: "AAA", Close: 50},  // BUY exec price
		{Date: base.AddDate(0, 0, 7), Symbol: "AAA", Close: 2},   // SELL signal, exec day8
		{Date: base.AddDate(0, 0, 8), Symbol: "AAA", Close: 55},  // SELL exec price
	}

	start := base.AddDate(0, 0, 2) // first exec day
	end1 := base.AddDate(0, 0, 6)  // after second BUY exec, before its SELL exec
	end2 := base.AddDate(0, 0, 8)  // includes second SELL exec

	series1 := all[:7] // up to end1
	series2 := all     // up to end2

	fee := 0.006

	cached, err := svc.runBacktestWindow(context.Background(), "mock_v1", "AAA", series1, start, end1, fee, defaultBacktestState(), nil)
	if err != nil {
		t.Fatalf("cached: %v", err)
	}
	if !cached.State.InPosition {
		t.Fatalf("expected open position at end1")
	}

	part, err := svc.runBacktestWindow(context.Background(), "mock_v1", "AAA", series2, start, end2, fee, cached.State, &end1)
	if err != nil {
		t.Fatalf("extend: %v", err)
	}
	merged := part
	merged.ClosedTrades = append(append([]BacktestTrade{}, cached.ClosedTrades...), part.ClosedTrades...)

	full, err := svc.runBacktestWindow(context.Background(), "mock_v1", "AAA", series2, start, end2, fee, defaultBacktestState(), nil)
	if err != nil {
		t.Fatalf("full: %v", err)
	}

	if merged.State.CompletedTrades != full.State.CompletedTrades {
		t.Fatalf("completed trades mismatch: got %d want %d", merged.State.CompletedTrades, full.State.CompletedTrades)
	}
	if len(merged.ClosedTrades) != len(full.ClosedTrades) {
		t.Fatalf("closed trade count mismatch: got %d want %d", len(merged.ClosedTrades), len(full.ClosedTrades))
	}

	mergedDetails, err := backtestCacheTickerToDetails(merged, start, end2, fee)
	if err != nil {
		t.Fatalf("merged details: %v", err)
	}
	fullDetails, err := backtestCacheTickerToDetails(full, start, end2, fee)
	if err != nil {
		t.Fatalf("full details: %v", err)
	}

	if mergedDetails.Summary.NetProfitPct != fullDetails.Summary.NetProfitPct {
		t.Fatalf("net profit mismatch: got %v want %v", mergedDetails.Summary.NetProfitPct, fullDetails.Summary.NetProfitPct)
	}
	if mergedDetails.Summary.GrossProfitPct != fullDetails.Summary.GrossProfitPct {
		t.Fatalf("gross profit mismatch: got %v want %v", mergedDetails.Summary.GrossProfitPct, fullDetails.Summary.GrossProfitPct)
	}
	if mergedDetails.Summary.OpenPosition != fullDetails.Summary.OpenPosition {
		t.Fatalf("open position mismatch: got %v want %v", mergedDetails.Summary.OpenPosition, fullDetails.Summary.OpenPosition)
	}
}
