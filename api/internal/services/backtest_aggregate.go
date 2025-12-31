package services

import (
	"sort"
)

type BatchBacktestAggregate struct {
	TotalTickers      int `json:"total_tickers"`
	SuccessfulTickers int `json:"successful_tickers"`
	ErrorTickers      int `json:"error_tickers"`
	OpenPositions     int `json:"open_positions"`

	TotalCompletedTrades int `json:"total_completed_trades"`
	TotalWinningTrades   int `json:"total_winning_trades"`
	TotalLosingTrades    int `json:"total_losing_trades"`

	AvgNetProfitPct    float64 `json:"avg_net_profit_pct"`
	MedianNetProfitPct float64 `json:"median_net_profit_pct"`

	TopTickers    []BacktestAggregateTicker `json:"top_tickers,omitempty"`
	BottomTickers []BacktestAggregateTicker `json:"bottom_tickers,omitempty"`
}

type BacktestAggregateTicker struct {
	Symbol          string  `json:"symbol"`
	NetProfitPct    float64 `json:"net_profit_pct"`
	GrossProfitPct  float64 `json:"gross_profit_pct"`
	CompletedTrades int     `json:"completed_trades"`
	WinningTrades   int     `json:"winning_trades"`
	LosingTrades    int     `json:"losing_trades"`
	OpenPosition    bool    `json:"open_position"`
}

func computeBacktestAggregate(byTicker []BacktestTickerSummary) BatchBacktestAggregate {
	agg := BatchBacktestAggregate{
		TotalTickers: len(byTicker),
	}

	valid := make([]BacktestTickerSummary, 0, len(byTicker))
	for _, item := range byTicker {
		if item.Error != "" {
			agg.ErrorTickers++
			continue
		}
		agg.SuccessfulTickers++
		if item.OpenPosition {
			agg.OpenPositions++
		}
		agg.TotalCompletedTrades += item.CompletedTrades
		agg.TotalWinningTrades += item.WinningTrades
		agg.TotalLosingTrades += item.LosingTrades
		valid = append(valid, item)
	}

	if len(valid) == 0 {
		return agg
	}

	// Avg/median of per-ticker net profit (not trade-weighted) to keep comparability.
	netValues := make([]float64, 0, len(valid))
	sumNet := 0.0
	for _, v := range valid {
		netValues = append(netValues, v.NetProfitPct)
		sumNet += v.NetProfitPct
	}
	agg.AvgNetProfitPct = sumNet / float64(len(netValues))

	sort.Float64s(netValues)
	mid := len(netValues) / 2
	if len(netValues)%2 == 1 {
		agg.MedianNetProfitPct = netValues[mid]
	} else {
		agg.MedianNetProfitPct = (netValues[mid-1] + netValues[mid]) / 2
	}

	// Top/bottom tickers by net profit.
	sort.Slice(valid, func(i, j int) bool {
		if valid[i].NetProfitPct == valid[j].NetProfitPct {
			return valid[i].Symbol < valid[j].Symbol
		}
		return valid[i].NetProfitPct > valid[j].NetProfitPct
	})

	limit := 10
	if len(valid) < limit {
		limit = len(valid)
	}

	toAggTicker := func(v BacktestTickerSummary) BacktestAggregateTicker {
		return BacktestAggregateTicker{
			Symbol:          v.Symbol,
			NetProfitPct:    v.NetProfitPct,
			GrossProfitPct:  v.GrossProfitPct,
			CompletedTrades: v.CompletedTrades,
			WinningTrades:   v.WinningTrades,
			LosingTrades:    v.LosingTrades,
			OpenPosition:    v.OpenPosition,
		}
	}

	agg.TopTickers = make([]BacktestAggregateTicker, 0, limit)
	for i := 0; i < limit; i++ {
		agg.TopTickers = append(agg.TopTickers, toAggTicker(valid[i]))
	}

	sort.Slice(valid, func(i, j int) bool {
		if valid[i].NetProfitPct == valid[j].NetProfitPct {
			return valid[i].Symbol < valid[j].Symbol
		}
		return valid[i].NetProfitPct < valid[j].NetProfitPct
	})

	agg.BottomTickers = make([]BacktestAggregateTicker, 0, limit)
	for i := 0; i < limit; i++ {
		agg.BottomTickers = append(agg.BottomTickers, toAggTicker(valid[i]))
	}

	return agg
}
