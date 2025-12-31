package services

import "testing"

func TestComputeBacktestAggregate_ExcludesErrorsAndTotals(t *testing.T) {
	input := []BacktestTickerSummary{
		{
			Symbol:          "AAA",
			CompletedTrades: 2,
			WinningTrades:   1,
			LosingTrades:    1,
			GrossProfitPct:  12,
			NetProfitPct:    10,
			OpenPosition:    false,
		},
		{
			Symbol: "BBB",
			Error:  "read error",
		},
		{
			Symbol:          "CCC",
			CompletedTrades: 1,
			WinningTrades:   1,
			LosingTrades:    0,
			GrossProfitPct:  -1,
			NetProfitPct:    -5,
			OpenPosition:    true,
		},
	}

	agg := computeBacktestAggregate(input)

	if agg.TotalTickers != 3 {
		t.Fatalf("TotalTickers=%d, want 3", agg.TotalTickers)
	}
	if agg.SuccessfulTickers != 2 {
		t.Fatalf("SuccessfulTickers=%d, want 2", agg.SuccessfulTickers)
	}
	if agg.ErrorTickers != 1 {
		t.Fatalf("ErrorTickers=%d, want 1", agg.ErrorTickers)
	}
	if agg.OpenPositions != 1 {
		t.Fatalf("OpenPositions=%d, want 1", agg.OpenPositions)
	}

	if agg.TotalCompletedTrades != 3 {
		t.Fatalf("TotalCompletedTrades=%d, want 3", agg.TotalCompletedTrades)
	}
	if agg.TotalWinningTrades != 2 {
		t.Fatalf("TotalWinningTrades=%d, want 2", agg.TotalWinningTrades)
	}
	if agg.TotalLosingTrades != 1 {
		t.Fatalf("TotalLosingTrades=%d, want 1", agg.TotalLosingTrades)
	}

	if agg.AvgNetProfitPct != 2.5 {
		t.Fatalf("AvgNetProfitPct=%v, want 2.5", agg.AvgNetProfitPct)
	}
	if agg.MedianNetProfitPct != 2.5 {
		t.Fatalf("MedianNetProfitPct=%v, want 2.5", agg.MedianNetProfitPct)
	}
}

func TestComputeBacktestAggregate_TopBottomOrdering(t *testing.T) {
	input := []BacktestTickerSummary{
		{Symbol: "BBB", NetProfitPct: 10, GrossProfitPct: 11, CompletedTrades: 1, WinningTrades: 1},
		{Symbol: "AAA", NetProfitPct: 10, GrossProfitPct: 12, CompletedTrades: 2, WinningTrades: 2},
		{Symbol: "CCC", NetProfitPct: -1, GrossProfitPct: -0.5, CompletedTrades: 1, LosingTrades: 1},
		{Symbol: "DDD", NetProfitPct: -1, GrossProfitPct: -2, CompletedTrades: 2, LosingTrades: 2},
	}

	agg := computeBacktestAggregate(input)
	if len(agg.TopTickers) != 4 {
		t.Fatalf("TopTickers len=%d, want 4", len(agg.TopTickers))
	}
	if len(agg.BottomTickers) != 4 {
		t.Fatalf("BottomTickers len=%d, want 4", len(agg.BottomTickers))
	}

	// Top: highest net profit, symbol tie-break ascending.
	if agg.TopTickers[0].Symbol != "AAA" || agg.TopTickers[1].Symbol != "BBB" {
		t.Fatalf("TopTickers[0..1]=%s,%s, want AAA,BBB", agg.TopTickers[0].Symbol, agg.TopTickers[1].Symbol)
	}

	// Bottom: lowest net profit, symbol tie-break ascending.
	if agg.BottomTickers[0].Symbol != "CCC" || agg.BottomTickers[1].Symbol != "DDD" {
		t.Fatalf("BottomTickers[0..1]=%s,%s, want CCC,DDD", agg.BottomTickers[0].Symbol, agg.BottomTickers[1].Symbol)
	}
}
