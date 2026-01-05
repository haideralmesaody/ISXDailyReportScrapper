package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

const backtestCacheVersion = 2

type backtestCacheMeta struct {
	Version        int                 `json:"version"`
	StrategyID     string              `json:"strategy_id"`
	StartDate      string              `json:"start_date"`
	EndDate        string              `json:"end_date"`
	TransactionFee float64             `json:"transaction_fee"`
	CombinedCSV    CombinedCSVSnapshot `json:"combined_csv"`
	UpdatedAt      string              `json:"updated_at"`
}

type backtestState struct {
	InPosition bool          `json:"in_position"`
	EntryPrice float64       `json:"entry_price"`
	Entry      BacktestTrade `json:"entry"`

	CompletedTrades int     `json:"completed_trades"`
	WinningTrades   int     `json:"winning_trades"`
	LosingTrades    int     `json:"losing_trades"`
	GrossEquity     float64 `json:"gross_equity"`
	NetEquity       float64 `json:"net_equity"`
}

type backtestCacheTicker struct {
	Symbol string `json:"symbol"`

	// CachedEndDate is the inclusive end date used when computing this cache entry.
	CachedEndDate string `json:"cached_end_date"`

	State        backtestState   `json:"state"`
	ClosedTrades []BacktestTrade `json:"closed_trades"`

	// LastPrice is the last observed close within CachedEndDate (may be earlier if ticker data ends earlier).
	LastPriceDate string  `json:"last_price_date,omitempty"`
	LastPrice     float64 `json:"last_price,omitempty"`

	Error string `json:"error,omitempty"`
}

func defaultBacktestState() backtestState {
	return backtestState{
		GrossEquity: 1,
		NetEquity:   1,
	}
}

type resolvedBatchBacktestOptions struct {
	Start time.Time
	End   time.Time
	Fee   float64
	Snap  CombinedCSVSnapshot
}

func (s *StrategyService) resolveBatchBacktestOptions(ctx context.Context, req ExecuteBatchRequest) (resolvedBatchBacktestOptions, error) {
	startProvided := strings.TrimSpace(req.BacktestStartDate) != ""
	endProvided := strings.TrimSpace(req.BacktestEndDate) != ""

	var start time.Time
	var end time.Time

	if startProvided {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.BacktestStartDate))
		if err != nil {
			return resolvedBatchBacktestOptions{}, fmt.Errorf("invalid backtest_start_date: %w", err)
		}
		start = parsed.UTC()
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	}
	if endProvided {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.BacktestEndDate))
		if err != nil {
			return resolvedBatchBacktestOptions{}, fmt.Errorf("invalid backtest_end_date: %w", err)
		}
		end = parsed.UTC()
		end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC)
	}

	fee := req.TransactionFee
	if fee == 0 {
		fee = 0.006
	}
	if fee < 0 || fee > 0.05 {
		return resolvedBatchBacktestOptions{}, fmt.Errorf("transaction_fee must be between 0 and 0.05")
	}

	min, max, snap, err := s.dataService.GetCombinedDateRange(ctx)
	if err != nil {
		return resolvedBatchBacktestOptions{}, err
	}

	if !startProvided {
		start = min
	}
	if !endProvided {
		end = max
	}
	if end.Before(start) {
		return resolvedBatchBacktestOptions{}, fmt.Errorf("backtest_end_date must be after backtest_start_date")
	}

	return resolvedBatchBacktestOptions{
		Start: start,
		End:   end,
		Fee:   fee,
		Snap:  snap,
	}, nil
}

func backtestCacheDir(dataDir, strategyID string) string {
	return filepath.Join(dataDir, "strategies", strategyID, "backtest_cache")
}

func loadBacktestCacheMeta(cacheDir string) (backtestCacheMeta, bool, error) {
	path := filepath.Join(cacheDir, "meta.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return backtestCacheMeta{}, false, nil
		}
		return backtestCacheMeta{}, false, fmt.Errorf("read backtest cache meta: %w", err)
	}
	var meta backtestCacheMeta
	if err := json.Unmarshal(b, &meta); err != nil {
		return backtestCacheMeta{}, false, fmt.Errorf("parse backtest cache meta: %w", err)
	}
	if meta.Version != backtestCacheVersion {
		return backtestCacheMeta{}, false, nil
	}
	return meta, true, nil
}

func saveBacktestCacheMeta(cacheDir string, meta backtestCacheMeta) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("create backtest cache dir: %w", err)
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backtest cache meta: %w", err)
	}
	path := filepath.Join(cacheDir, "meta.json")
	if err := os.WriteFile(path, metaBytes, 0644); err != nil {
		return fmt.Errorf("write backtest cache meta: %w", err)
	}
	return nil
}

func clearBacktestCache(cacheDir string) error {
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("clear backtest cache: %w", err)
	}
	return nil
}

func loadBacktestCacheTicker(cacheDir, symbol string) (backtestCacheTicker, bool, error) {
	path := filepath.Join(cacheDir, "by_ticker", strings.ToUpper(symbol)+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return backtestCacheTicker{}, false, nil
		}
		return backtestCacheTicker{}, false, fmt.Errorf("read backtest cache ticker (%s): %w", symbol, err)
	}
	var t backtestCacheTicker
	if err := json.Unmarshal(b, &t); err != nil {
		return backtestCacheTicker{}, false, fmt.Errorf("parse backtest cache ticker (%s): %w", symbol, err)
	}
	return t, true, nil
}

func saveBacktestCacheTicker(cacheDir string, ticker backtestCacheTicker) error {
	dir := filepath.Join(cacheDir, "by_ticker")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create backtest cache ticker dir: %w", err)
	}
	b, err := json.MarshalIndent(ticker, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backtest cache ticker (%s): %w", ticker.Symbol, err)
	}
	path := filepath.Join(dir, strings.ToUpper(ticker.Symbol)+".json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("write backtest cache ticker (%s): %w", ticker.Symbol, err)
	}
	return nil
}

func parseCacheDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, err
	}
	parsed = parsed.UTC()
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}

func (m backtestCacheMeta) isCompatible(strategyID string, start time.Time, fee float64, snap CombinedCSVSnapshot) bool {
	if m.Version != backtestCacheVersion {
		return false
	}
	if m.StrategyID != strategyID {
		return false
	}
	if m.TransactionFee != fee {
		return false
	}
	wantStart := start.UTC().Format("2006-01-02")
	if m.StartDate != wantStart {
		return false
	}
	if !m.CombinedCSV.ModTime.Equal(snap.ModTime) || m.CombinedCSV.Size != snap.Size {
		return false
	}
	return true
}

func backtestCacheTickerToDetails(t backtestCacheTicker, start, end time.Time, fee float64) (BacktestTickerDetails, error) {
	if t.Error != "" {
		summary := BacktestTickerSummary{Symbol: t.Symbol, Error: t.Error}
		return BacktestTickerDetails{Symbol: t.Symbol, Summary: summary, Trades: []BacktestTrade{}}, nil
	}
	if t.State.GrossEquity == 0 {
		t.State.GrossEquity = 1
	}
	if t.State.NetEquity == 0 {
		t.State.NetEquity = 1
	}

	trades := make([]BacktestTrade, 0, len(t.ClosedTrades)+1)
	trades = append(trades, t.ClosedTrades...)

	openPosition := false
	grossEquity := t.State.GrossEquity
	netEquity := t.State.NetEquity

	if t.State.InPosition && t.State.EntryPrice > 0 && t.LastPrice > 0 {
		lastDate, err := parseCacheDate(t.LastPriceDate)
		if err == nil && !lastDate.Before(start) && !lastDate.After(end) {
			openPosition = true
			grossFactor := t.LastPrice / t.State.EntryPrice
			netFactor := (1 - fee) * grossFactor * (1 - fee)

			entry := t.State.Entry
			entry.Symbol = t.Symbol
			entry.Status = "OPEN"
			entry.TransactionFee = fee
			entry.SellDate = lastDate.Format("2006-01-02")
			entry.SellPrice = t.LastPrice
			entry.GrossReturnPct = (grossFactor - 1) * 100
			entry.NetReturnPct = (netFactor - 1) * 100

			trades = append(trades, entry)
			grossEquity *= grossFactor
			netEquity *= netFactor
		}
	}

	summary := BacktestTickerSummary{
		Symbol:          t.Symbol,
		CompletedTrades: t.State.CompletedTrades,
		WinningTrades:   t.State.WinningTrades,
		LosingTrades:    t.State.LosingTrades,
		GrossProfitPct:  round2((grossEquity - 1) * 100),
		NetProfitPct:    round2((netEquity - 1) * 100),
		OpenPosition:    openPosition,
	}

	return BacktestTickerDetails{
		Symbol:  t.Symbol,
		Summary: summary,
		Trades:  trades,
	}, nil
}

func (s *StrategyService) runBacktestWindow(
	ctx context.Context,
	strategyID string,
	symbol string,
	series []liquidity.TradingDay,
	start time.Time,
	end time.Time,
	fee float64,
	initial backtestState,
	resumeAfter *time.Time,
) (backtestCacheTicker, error) {
	if len(series) < 3 {
		return backtestCacheTicker{Symbol: symbol, Error: "insufficient data for backtest"}, nil
	}

	start = start.UTC()
	end = end.UTC()

	state := initial
	if state.GrossEquity == 0 {
		state.GrossEquity = 1
	}
	if state.NetEquity == 0 {
		state.NetEquity = 1
	}

	closedTrades := make([]BacktestTrade, 0, 16)

	for i := 0; i < len(series)-1; i++ {
		select {
		case <-ctx.Done():
			return backtestCacheTicker{}, ctx.Err()
		default:
		}

		signalDate := series[i].Date.UTC()
		execDate := series[i+1].Date.UTC()
		if execDate.Before(start) {
			continue
		}
		if execDate.After(end) {
			break
		}
		if resumeAfter != nil && !execDate.After(*resumeAfter) {
			continue
		}

		window := series[:i+1]
		signal, err := s.manager.Execute(ctx, strategyID, window)
		if err != nil {
			continue
		}

		if signal.Action == strategy.SignalBuy && !state.InPosition {
			buyPrice := series[i+1].Close
			if buyPrice <= 0 {
				continue
			}
			state.InPosition = true
			state.EntryPrice = buyPrice
			state.Entry = BacktestTrade{
				Symbol:         symbol,
				Status:         "OPEN",
				SignalBuyDate:  signalDate.Format("2006-01-02"),
				BuyDate:        execDate.Format("2006-01-02"),
				BuyPrice:       buyPrice,
				TransactionFee: fee,
			}
			continue
		}

		if signal.Action == strategy.SignalSell && state.InPosition {
			sellPrice := series[i+1].Close
			if sellPrice <= 0 || state.EntryPrice <= 0 {
				continue
			}

			grossFactor := sellPrice / state.EntryPrice
			netFactor := (1 - fee) * grossFactor * (1 - fee)

			trade := state.Entry
			trade.Status = "CLOSED"
			trade.SignalSellDate = signalDate.Format("2006-01-02")
			trade.SellDate = execDate.Format("2006-01-02")
			trade.SellPrice = sellPrice
			trade.GrossReturnPct = (grossFactor - 1) * 100
			trade.NetReturnPct = (netFactor - 1) * 100
			trade.TransactionFee = fee

			closedTrades = append(closedTrades, trade)
			state.CompletedTrades++
			if trade.NetReturnPct > 0 {
				state.WinningTrades++
			} else if trade.NetReturnPct < 0 {
				state.LosingTrades++
			}

			state.GrossEquity *= grossFactor
			state.NetEquity *= netFactor

			state.InPosition = false
			state.EntryPrice = 0
			state.Entry = BacktestTrade{}
		}
	}

	last := series[len(series)-1]
	lastDate := last.Date.UTC()
	lastPrice := last.Close
	if lastPrice <= 0 {
		lastPrice = 0
	}

	return backtestCacheTicker{
		Symbol:        symbol,
		CachedEndDate: end.Format("2006-01-02"),
		State:         state,
		ClosedTrades:  closedTrades,
		LastPriceDate: lastDate.Format("2006-01-02"),
		LastPrice:     lastPrice,
	}, nil
}
