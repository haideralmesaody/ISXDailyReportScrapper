package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

// StrategyService handles strategy-related business logic
type StrategyService struct {
	manager     *strategy.Manager
	dataService *DataService
	logger      *slog.Logger
	config      *strategy.Config
}

// NewStrategyService creates a new strategy service
func NewStrategyService(
	manager *strategy.Manager,
	dataService *DataService,
	logger *slog.Logger,
) *StrategyService {
	return &StrategyService{
		manager:     manager,
		dataService: dataService,
		logger:      logger,
		config:      strategy.DefaultConfig(),
	}
}

// ListStrategies returns all available strategies
func (s *StrategyService) ListStrategies(ctx context.Context) ([]strategy.StrategyInfo, error) {
	s.logger.InfoContext(ctx, "listing available strategies")

	strategies := s.manager.ListStrategies(ctx)

	s.logger.InfoContext(ctx, "strategies listed",
		"count", len(strategies),
	)

	return strategies, nil
}

// GetStrategy returns a specific strategy by ID
func (s *StrategyService) GetStrategy(ctx context.Context, strategyID string) (strategy.StrategyInfo, error) {
	s.logger.InfoContext(ctx, "getting strategy",
		"strategy_id", strategyID,
	)

	strat, err := s.manager.GetStrategy(ctx, strategyID)
	if err != nil {
		return strategy.StrategyInfo{}, fmt.Errorf("get strategy: %w", err)
	}

	info := strategy.StrategyInfo{
		ID:          strat.ID(),
		Name:        strat.Name(),
		Description: strat.Description(),
	}

	return info, nil
}

// ExecuteStrategy runs a strategy against symbol data
func (s *StrategyService) ExecuteStrategy(ctx context.Context, req ExecuteStrategyRequest) (strategy.Signal, error) {
	s.logger.InfoContext(ctx, "executing strategy",
		"strategy_id", req.StrategyID,
		"symbol", req.Symbol,
		"data_points", req.DataPoints,
	)

	// Get trading data
	data, err := s.getSymbolData(ctx, req.Symbol, req.DataPoints)
	if err != nil {
		return strategy.Signal{}, fmt.Errorf("get symbol data: %w", err)
	}

	// Execute strategy
	signal, err := s.manager.Execute(ctx, req.StrategyID, data)
	if err != nil {
		return strategy.Signal{}, fmt.Errorf("execute strategy: %w", err)
	}

	s.logger.InfoContext(ctx, "strategy executed successfully",
		"strategy_id", req.StrategyID,
		"symbol", req.Symbol,
		"action", signal.Action,
		"strength", signal.Strength,
	)

	return signal, nil
}

// ExecuteMultipleStrategies runs multiple strategies against the same data
func (s *StrategyService) ExecuteMultipleStrategies(ctx context.Context, req ExecuteMultipleRequest) ([]StrategyResult, error) {
	s.logger.InfoContext(ctx, "executing multiple strategies",
		"strategies", len(req.StrategyIDs),
		"symbol", req.Symbol,
	)

	// Get trading data once
	data, err := s.getSymbolData(ctx, req.Symbol, req.DataPoints)
	if err != nil {
		return nil, fmt.Errorf("get symbol data: %w", err)
	}

	results := make([]StrategyResult, 0, len(req.StrategyIDs))

	// Execute each strategy
	for _, strategyID := range req.StrategyIDs {
		signal, err := s.manager.Execute(ctx, strategyID, data)

		result := StrategyResult{
			StrategyID: strategyID,
			Symbol:     req.Symbol,
			Success:    err == nil,
		}

		if err != nil {
			result.Error = err.Error()
		} else {
			result.Signal = signal
		}

		results = append(results, result)
	}

	s.logger.InfoContext(ctx, "multiple strategies executed",
		"symbol", req.Symbol,
		"total", len(results),
		"successful", s.countSuccessful(results),
	)

	return results, nil
}

// ExecuteStrategyBatch runs a strategy across all (or selected) tickers using real data.
// It also persists results to: <data_dir>/strategies/<strategy_id>/<run_id>/...
func (s *StrategyService) ExecuteStrategyBatch(ctx context.Context, strategyID string, req ExecuteBatchRequest) (ExecuteBatchResponse, error) {
	startedAt := time.Now()

	if s.dataService == nil {
		return ExecuteBatchResponse{}, fmt.Errorf("data service not configured")
	}
	if strings.TrimSpace(strategyID) == "" {
		return ExecuteBatchResponse{}, fmt.Errorf("strategy_id is required")
	}

	dataPoints := req.DataPoints
	if dataPoints == 0 {
		dataPoints = 120
	}
	if dataPoints < 16 {
		return ExecuteBatchResponse{}, fmt.Errorf("data_points must be >= 16 for RSI(14) crossings")
	}

	var symbols []string
	if len(req.Symbols) > 0 {
		symbols = make([]string, 0, len(req.Symbols))
		for _, sym := range req.Symbols {
			sym = strings.TrimSpace(strings.ToUpper(sym))
			if sym == "" {
				continue
			}
			symbols = append(symbols, sym)
		}
		sort.Strings(symbols)
	} else {
		all, err := s.dataService.ListTickerSymbols(ctx)
		if err != nil {
			return ExecuteBatchResponse{}, fmt.Errorf("list ticker symbols: %w", err)
		}
		symbols = all
	}

	runID := fmt.Sprintf("%s_%d", time.Now().UTC().Format("20060102_150405"), time.Now().UTC().UnixNano())

	signals := make([]strategy.Signal, 0, len(symbols))
	errorsList := make([]ExecuteBatchError, 0)
	buyCount := 0
	sellCount := 0
	holdCount := 0

	for _, symbol := range symbols {
		select {
		case <-ctx.Done():
			return ExecuteBatchResponse{}, ctx.Err()
		default:
		}

		data, err := s.getSymbolData(ctx, symbol, dataPoints)
		if err != nil {
			errorsList = append(errorsList, ExecuteBatchError{Symbol: symbol, Error: err.Error()})
			continue
		}

		signal, err := s.manager.Execute(ctx, strategyID, data)
		if err != nil {
			errorsList = append(errorsList, ExecuteBatchError{Symbol: symbol, Error: err.Error()})
			continue
		}

		switch signal.Action {
		case strategy.SignalBuy:
			buyCount++
		case strategy.SignalSell:
			sellCount++
		default:
			holdCount++
		}

		signals = append(signals, signal)
	}

	completedAt := time.Now()

	resp := ExecuteBatchResponse{
		RunID:       runID,
		StrategyID:  strategyID,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Total:       len(symbols),
		BuyCount:    buyCount,
		SellCount:   sellCount,
		HoldCount:   holdCount,
		Signals:     signals,
		Errors:      errorsList,
	}

	if err := s.persistBatchRun(ctx, resp); err != nil {
		return resp, fmt.Errorf("persist batch run: %w", err)
	}

	if req.IncludeBacktest {
		backtestSummary, detailsBySymbol, err := s.runBatchBacktest(ctx, strategyID, symbols, req)
		if err != nil {
			return resp, fmt.Errorf("run batch backtest: %w", err)
		}
		resp.Backtest = backtestSummary

		// Rewrite summary.json to include backtest and persist the per-ticker details.
		if err := s.persistBatchRun(ctx, resp); err != nil {
			return resp, fmt.Errorf("persist batch run (with backtest): %w", err)
		}
		if err := s.persistBatchBacktest(ctx, resp, detailsBySymbol); err != nil {
			return resp, fmt.Errorf("persist batch backtest: %w", err)
		}
	}

	return resp, nil
}

func (s *StrategyService) runBatchBacktest(
	ctx context.Context,
	strategyID string,
	symbols []string,
	req ExecuteBatchRequest,
) (*BatchBacktestSummary, map[string]BacktestTickerDetails, error) {
	start, end, fee, err := parseBacktestOptions(req)
	if err != nil {
		return nil, nil, err
	}

	byTicker := make([]BacktestTickerSummary, 0, len(symbols))
	detailsBySymbol := make(map[string]BacktestTickerDetails, len(symbols))

	// Load extra warmup data before start so RSI is initialized more realistically.
	startLoad := start.AddDate(0, 0, -60)

	for _, symbol := range symbols {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}

		series, err := s.dataService.LoadTickerTradingHistoryRange(ctx, symbol, startLoad, end)
		if err != nil {
			summary := BacktestTickerSummary{Symbol: symbol, Error: err.Error()}
			byTicker = append(byTicker, summary)
			detailsBySymbol[symbol] = BacktestTickerDetails{Symbol: symbol, Summary: summary, Trades: []BacktestTrade{}}
			continue
		}

		details, err := s.backtestTradesForSeries(ctx, strategyID, symbol, series, start, end, fee)
		if err != nil {
			summary := BacktestTickerSummary{Symbol: symbol, Error: err.Error()}
			byTicker = append(byTicker, summary)
			detailsBySymbol[symbol] = BacktestTickerDetails{Symbol: symbol, Summary: summary, Trades: []BacktestTrade{}}
			continue
		}

		byTicker = append(byTicker, details.Summary)
		detailsBySymbol[symbol] = details
	}

	summary := &BatchBacktestSummary{
		StartDate:      start.UTC().Format("2006-01-02"),
		EndDate:        end.UTC().Format("2006-01-02"),
		TransactionFee: fee,
		ByTicker:       byTicker,
		Aggregate:      computeBacktestAggregate(byTicker),
	}

	return summary, detailsBySymbol, nil
}

func parseBacktestOptions(req ExecuteBatchRequest) (time.Time, time.Time, float64, error) {
	now := time.Now().UTC()
	endDefault := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDefault := endDefault.AddDate(0, 0, -90)

	start := startDefault
	end := endDefault

	if strings.TrimSpace(req.BacktestStartDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.BacktestStartDate))
		if err != nil {
			return time.Time{}, time.Time{}, 0, fmt.Errorf("invalid backtest_start_date: %w", err)
		}
		start = parsed.UTC()
	}
	if strings.TrimSpace(req.BacktestEndDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.BacktestEndDate))
		if err != nil {
			return time.Time{}, time.Time{}, 0, fmt.Errorf("invalid backtest_end_date: %w", err)
		}
		end = parsed.UTC()
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("backtest_end_date must be after backtest_start_date")
	}

	fee := req.TransactionFee
	if fee == 0 {
		fee = 0.006
	}
	if fee < 0 || fee > 0.05 {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("transaction_fee must be between 0 and 0.05")
	}

	return start, end, fee, nil
}

// backtestTradesForSeries runs a per-ticker backtest:
// - Signals are evaluated on day D (EOD)
// - Executions happen at next-day close (D+1)
// - Open positions are marked-to-market at end of series (does not count as completed trade)
func (s *StrategyService) backtestTradesForSeries(
	ctx context.Context,
	strategyID string,
	symbol string,
	series []liquidity.TradingDay,
	start time.Time,
	end time.Time,
	fee float64,
) (BacktestTickerDetails, error) {
	if len(series) < 3 {
		return BacktestTickerDetails{}, fmt.Errorf("insufficient data for backtest")
	}

	start = start.UTC()
	end = end.UTC()

	inPosition := false
	entryPrice := 0.0
	entry := BacktestTrade{}

	trades := make([]BacktestTrade, 0, 16)
	completedTrades := 0
	winningTrades := 0
	losingTrades := 0

	grossEquity := 1.0
	netEquity := 1.0

	for i := 0; i < len(series)-1; i++ {
		select {
		case <-ctx.Done():
			return BacktestTickerDetails{}, ctx.Err()
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

		window := series[:i+1]
		signal, err := s.manager.Execute(ctx, strategyID, window)
		if err != nil {
			continue
		}

		if signal.Action == strategy.SignalBuy && !inPosition {
			buyPrice := series[i+1].Close
			if buyPrice <= 0 {
				continue
			}

			inPosition = true
			entryPrice = buyPrice
			entry = BacktestTrade{
				Symbol:         symbol,
				Status:         "OPEN",
				SignalBuyDate:  signalDate.Format("2006-01-02"),
				BuyDate:        execDate.Format("2006-01-02"),
				BuyPrice:       buyPrice,
				TransactionFee: fee,
			}
			continue
		}

		if signal.Action == strategy.SignalSell && inPosition {
			sellPrice := series[i+1].Close
			if sellPrice <= 0 || entryPrice <= 0 {
				continue
			}

			grossFactor := sellPrice / entryPrice
			netFactor := (1 - fee) * grossFactor * (1 - fee)

			trade := entry
			trade.Status = "CLOSED"
			trade.SignalSellDate = signalDate.Format("2006-01-02")
			trade.SellDate = execDate.Format("2006-01-02")
			trade.SellPrice = sellPrice
			trade.GrossReturnPct = (grossFactor - 1) * 100
			trade.NetReturnPct = (netFactor - 1) * 100

			trades = append(trades, trade)
			completedTrades++
			if trade.NetReturnPct > 0 {
				winningTrades++
			} else if trade.NetReturnPct < 0 {
				losingTrades++
			}

			grossEquity *= grossFactor
			netEquity *= netFactor

			inPosition = false
			entryPrice = 0
			entry = BacktestTrade{}
		}
	}

	openPosition := false
	if inPosition && entryPrice > 0 {
		last := series[len(series)-1]
		lastDate := last.Date.UTC()
		lastPrice := last.Close
		if lastPrice > 0 && !lastDate.Before(start) && !lastDate.After(end) {
			openPosition = true
			grossFactor := lastPrice / entryPrice
			netFactor := (1 - fee) * grossFactor * (1 - fee)

			entry.Status = "OPEN"
			entry.SellDate = lastDate.Format("2006-01-02") // mark-to-market date
			entry.SellPrice = lastPrice                    // mark-to-market price
			entry.GrossReturnPct = (grossFactor - 1) * 100
			entry.NetReturnPct = (netFactor - 1) * 100

			trades = append(trades, entry)
			grossEquity *= grossFactor
			netEquity *= netFactor
		}
	}

	grossProfitPct := (grossEquity - 1) * 100
	netProfitPct := (netEquity - 1) * 100

	summary := BacktestTickerSummary{
		Symbol:          symbol,
		CompletedTrades: completedTrades,
		WinningTrades:   winningTrades,
		LosingTrades:    losingTrades,
		GrossProfitPct:  round2(grossProfitPct),
		NetProfitPct:    round2(netProfitPct),
		OpenPosition:    openPosition,
	}

	return BacktestTickerDetails{
		Symbol:  symbol,
		Summary: summary,
		Trades:  trades,
	}, nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// RunBacktest performs backtesting for a strategy
func (s *StrategyService) RunBacktest(ctx context.Context, req BacktestRequest) (strategy.BacktestResult, error) {
	s.logger.InfoContext(ctx, "starting backtest",
		"strategy_id", req.StrategyID,
		"symbol", req.Symbol,
		"start_date", req.StartDate,
		"end_date", req.EndDate,
	)

	// Get historical data
	data, err := s.getHistoricalData(ctx, req.Symbol, req.StartDate, req.EndDate)
	if err != nil {
		return strategy.BacktestResult{}, fmt.Errorf("get historical data: %w", err)
	}

	// Prepare backtest config
	config := strategy.BacktestConfig{
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		InitialCash: req.InitialCash,
		Commission:  req.Commission,
		Slippage:    req.Slippage,
	}

	// Run backtest
	result, err := s.manager.Backtest(ctx, req.StrategyID, data, config)
	if err != nil {
		return strategy.BacktestResult{}, fmt.Errorf("run backtest: %w", err)
	}

	s.logger.InfoContext(ctx, "backtest completed",
		"strategy_id", req.StrategyID,
		"symbol", req.Symbol,
		"total_return", result.TotalReturn,
		"total_trades", result.TotalTrades,
	)

	return result, nil
}

func (s *StrategyService) persistBatchRun(ctx context.Context, result ExecuteBatchResponse) error {
	if s.dataService == nil || s.dataService.paths == nil {
		return fmt.Errorf("data service paths not configured")
	}

	baseDir := filepath.Join(s.dataService.paths.DataDir, "strategies", result.StrategyID, result.RunID)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("create run dir: %w", err)
	}

	// summary.json (includes signals)
	summaryPath := filepath.Join(baseDir, "summary.json")
	summaryBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal summary: %w", err)
	}
	if err := os.WriteFile(summaryPath, summaryBytes, 0644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}

	// signals.json (signals only)
	signalsPath := filepath.Join(baseDir, "signals.json")
	signalsBytes, err := json.MarshalIndent(result.Signals, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal signals: %w", err)
	}
	if err := os.WriteFile(signalsPath, signalsBytes, 0644); err != nil {
		return fmt.Errorf("write signals: %w", err)
	}

	// signals.csv
	csvPath := filepath.Join(baseDir, "signals.csv")
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("create signals.csv: %w", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	if err := writer.Write([]string{"date", "symbol", "action", "strength", "price", "rsi", "prev_rsi", "reason"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	for _, sig := range result.Signals {
		rsi := ""
		prevRSI := ""
		if sig.Metadata != nil {
			if v, ok := sig.Metadata["rsi"]; ok {
				rsi = fmt.Sprintf("%v", v)
			}
			if v, ok := sig.Metadata["prev_rsi"]; ok {
				prevRSI = fmt.Sprintf("%v", v)
			}
		}
		dateStr := sig.Timestamp.UTC().Format("2006-01-02")
		row := []string{
			dateStr,
			sig.Symbol,
			string(sig.Action),
			fmt.Sprintf("%.2f", sig.Strength),
			fmt.Sprintf("%.6f", sig.Price),
			rsi,
			prevRSI,
			sig.Reasoning,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}

	s.logger.InfoContext(ctx, "strategy batch results persisted",
		"strategy_id", result.StrategyID,
		"run_id", result.RunID,
		"dir", baseDir,
	)
	return nil
}

func (s *StrategyService) persistBatchBacktest(ctx context.Context, result ExecuteBatchResponse, detailsBySymbol map[string]BacktestTickerDetails) error {
	if result.Backtest == nil {
		return nil
	}
	if s.dataService == nil || s.dataService.paths == nil {
		return fmt.Errorf("data service paths not configured")
	}

	baseDir := filepath.Join(s.dataService.paths.DataDir, "strategies", result.StrategyID, result.RunID, "backtest")
	byTickerDir := filepath.Join(baseDir, "by_ticker")
	if err := os.MkdirAll(byTickerDir, 0755); err != nil {
		return fmt.Errorf("create backtest dir: %w", err)
	}

	// summary.json
	summaryPath := filepath.Join(baseDir, "summary.json")
	summaryBytes, err := json.MarshalIndent(result.Backtest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backtest summary: %w", err)
	}
	if err := os.WriteFile(summaryPath, summaryBytes, 0644); err != nil {
		return fmt.Errorf("write backtest summary: %w", err)
	}

	// aggregate.json
	aggregatePath := filepath.Join(baseDir, "aggregate.json")
	aggregateBytes, err := json.MarshalIndent(result.Backtest.Aggregate, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backtest aggregate: %w", err)
	}
	if err := os.WriteFile(aggregatePath, aggregateBytes, 0644); err != nil {
		return fmt.Errorf("write backtest aggregate: %w", err)
	}

	// summary.csv
	csvPath := filepath.Join(baseDir, "summary.csv")
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("create backtest summary.csv: %w", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	if err := writer.Write([]string{"symbol", "completed_trades", "winning_trades", "losing_trades", "gross_profit_pct", "net_profit_pct", "open_position", "error"}); err != nil {
		return fmt.Errorf("write backtest csv header: %w", err)
	}
	for _, item := range result.Backtest.ByTicker {
		row := []string{
			item.Symbol,
			fmt.Sprintf("%d", item.CompletedTrades),
			fmt.Sprintf("%d", item.WinningTrades),
			fmt.Sprintf("%d", item.LosingTrades),
			fmt.Sprintf("%.2f", item.GrossProfitPct),
			fmt.Sprintf("%.2f", item.NetProfitPct),
			fmt.Sprintf("%t", item.OpenPosition),
			item.Error,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("write backtest csv row: %w", err)
		}
	}

	for symbol, details := range detailsBySymbol {
		path := filepath.Join(byTickerDir, strings.ToUpper(symbol)+".json")
		bytes, err := json.MarshalIndent(details, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal backtest details %s: %w", symbol, err)
		}
		if err := os.WriteFile(path, bytes, 0644); err != nil {
			return fmt.Errorf("write backtest details %s: %w", symbol, err)
		}
	}

	s.logger.InfoContext(ctx, "strategy batch backtest persisted",
		"strategy_id", result.StrategyID,
		"run_id", result.RunID,
		"dir", baseDir,
	)
	return nil
}

// GetStrategySignals returns recent signals for a strategy
func (s *StrategyService) GetStrategySignals(ctx context.Context, strategyID string, limit int) ([]strategy.Signal, error) {
	// In a real implementation, this would fetch from a database
	// For now, return empty slice
	return []strategy.Signal{}, nil
}

// GetBacktestTickerDetails loads persisted backtest details for a specific strategy run and symbol.
// Data source: <data_dir>/strategies/<strategy_id>/<run_id>/backtest/by_ticker/<SYMBOL>.json
func (s *StrategyService) GetBacktestTickerDetails(ctx context.Context, strategyID, runID, symbol string) (BacktestTickerDetails, error) {
	if s.dataService == nil || s.dataService.paths == nil {
		return BacktestTickerDetails{}, fmt.Errorf("data service paths not configured")
	}

	strategyID = strings.TrimSpace(strategyID)
	runID = strings.TrimSpace(runID)
	symbol = strings.TrimSpace(strings.ToUpper(symbol))

	if strategyID == "" {
		return BacktestTickerDetails{}, fmt.Errorf("strategy_id is required")
	}
	if runID == "" {
		return BacktestTickerDetails{}, fmt.Errorf("run_id is required")
	}
	if symbol == "" {
		return BacktestTickerDetails{}, fmt.Errorf("symbol is required")
	}

	select {
	case <-ctx.Done():
		return BacktestTickerDetails{}, ctx.Err()
	default:
	}

	path := filepath.Join(s.dataService.paths.DataDir, "strategies", strategyID, runID, "backtest", "by_ticker", symbol+".json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return BacktestTickerDetails{}, fmt.Errorf("read backtest details (%s): %w", path, err)
	}

	var details BacktestTickerDetails
	if err := json.Unmarshal(bytes, &details); err != nil {
		return BacktestTickerDetails{}, fmt.Errorf("parse backtest details: %w", err)
	}
	return details, nil
}

// ValidateStrategyParameters validates strategy parameters
func (s *StrategyService) ValidateStrategyParameters(ctx context.Context, strategyID string, params strategy.StrategyParams) error {
	strat, err := s.manager.GetStrategy(ctx, strategyID)
	if err != nil {
		return fmt.Errorf("get strategy: %w", err)
	}

	return strat.Validate(ctx, params)
}

// getSymbolData retrieves trading data for a symbol
func (s *StrategyService) getSymbolData(ctx context.Context, symbol string, dataPoints int) ([]liquidity.TradingDay, error) {
	if s.dataService == nil {
		return nil, fmt.Errorf("data service not configured")
	}
	return s.dataService.LoadTickerTradingHistoryTail(ctx, symbol, dataPoints)
}

// getHistoricalData retrieves historical trading data
func (s *StrategyService) getHistoricalData(ctx context.Context, symbol string, start, end time.Time) ([]liquidity.TradingDay, error) {
	if s.dataService == nil {
		return nil, fmt.Errorf("data service not configured")
	}
	return s.dataService.LoadTickerTradingHistoryRange(ctx, symbol, start, end)
}

// countSuccessful counts successful strategy executions
func (s *StrategyService) countSuccessful(results []StrategyResult) int {
	count := 0
	for _, result := range results {
		if result.Success {
			count++
		}
	}
	return count
}

// Request/Response types
type ExecuteStrategyRequest struct {
	StrategyID string `json:"strategy_id" validate:"required"`
	Symbol     string `json:"symbol" validate:"required"`
	DataPoints int    `json:"data_points" validate:"min=20,max=1000"`
}

type ExecuteMultipleRequest struct {
	StrategyIDs []string `json:"strategy_ids" validate:"required,min=1"`
	Symbol      string   `json:"symbol" validate:"required"`
	DataPoints  int      `json:"data_points" validate:"min=20,max=1000"`
}

type BacktestRequest struct {
	StrategyID  string    `json:"strategy_id" validate:"required"`
	Symbol      string    `json:"symbol" validate:"required"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	InitialCash float64   `json:"initial_cash" validate:"min=1000"`
	Commission  float64   `json:"commission" validate:"min=0,max=0.1"`
	Slippage    float64   `json:"slippage" validate:"min=0,max=0.1"`
}

type StrategyResult struct {
	StrategyID string          `json:"strategy_id"`
	Symbol     string          `json:"symbol"`
	Success    bool            `json:"success"`
	Signal     strategy.Signal `json:"signal,omitempty"`
	Error      string          `json:"error,omitempty"`
}

type ExecuteBatchRequest struct {
	// DataPoints is the number of most-recent rows to load per ticker.
	// Must be >= 16 for RSI(14) crossings (14 + 2 points).
	DataPoints int `json:"data_points"`

	// Symbols optionally limits execution to a subset of tickers.
	// When omitted, runs across all tickers found in reports/ticker.
	Symbols []string `json:"symbols,omitempty"`

	// IncludeBacktest runs a backtest per ticker for the provided date range.
	IncludeBacktest bool `json:"include_backtest,omitempty"`

	// BacktestStartDate and BacktestEndDate are inclusive dates in YYYY-MM-DD.
	BacktestStartDate string `json:"backtest_start_date,omitempty"`
	BacktestEndDate   string `json:"backtest_end_date,omitempty"`

	// TransactionFee is applied per transaction (BUY and SELL), as a fraction (e.g. 0.006 = 0.6%).
	TransactionFee float64 `json:"transaction_fee,omitempty"`
}

type ExecuteBatchError struct {
	Symbol string `json:"symbol"`
	Error  string `json:"error"`
}

type ExecuteBatchResponse struct {
	RunID       string              `json:"run_id"`
	StrategyID  string              `json:"strategy_id"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt time.Time           `json:"completed_at"`
	Total       int                 `json:"total"`
	BuyCount    int                 `json:"buy_count"`
	SellCount   int                 `json:"sell_count"`
	HoldCount   int                 `json:"hold_count"`
	Signals     []strategy.Signal   `json:"signals"`
	Errors      []ExecuteBatchError `json:"errors,omitempty"`

	Backtest *BatchBacktestSummary `json:"backtest,omitempty"`
}

type BatchBacktestSummary struct {
	StartDate      string                  `json:"start_date"`
	EndDate        string                  `json:"end_date"`
	TransactionFee float64                 `json:"transaction_fee"`
	ByTicker       []BacktestTickerSummary `json:"by_ticker"`
	Aggregate      BatchBacktestAggregate  `json:"aggregate"`
}

type BacktestTickerSummary struct {
	Symbol          string  `json:"symbol"`
	CompletedTrades int     `json:"completed_trades"`
	WinningTrades   int     `json:"winning_trades"`
	LosingTrades    int     `json:"losing_trades"`
	GrossProfitPct  float64 `json:"gross_profit_pct"`
	NetProfitPct    float64 `json:"net_profit_pct"`
	OpenPosition    bool    `json:"open_position"`
	Error           string  `json:"error,omitempty"`
}

type BacktestTrade struct {
	Symbol string `json:"symbol"`
	Status string `json:"status"`

	SignalBuyDate  string  `json:"signal_buy_date,omitempty"`
	BuyDate        string  `json:"buy_date,omitempty"`
	BuyPrice       float64 `json:"buy_price,omitempty"`
	SignalSellDate string  `json:"signal_sell_date,omitempty"`
	SellDate       string  `json:"sell_date,omitempty"`
	SellPrice      float64 `json:"sell_price,omitempty"`

	GrossReturnPct float64 `json:"gross_return_pct"`
	NetReturnPct   float64 `json:"net_return_pct"`

	TransactionFee float64 `json:"transaction_fee"`
}

type BacktestTickerDetails struct {
	Symbol  string                `json:"symbol"`
	Summary BacktestTickerSummary `json:"summary"`
	Trades  []BacktestTrade       `json:"trades"`
}
