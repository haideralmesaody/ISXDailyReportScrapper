package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
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

	return resp, nil
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

// GetStrategySignals returns recent signals for a strategy
func (s *StrategyService) GetStrategySignals(ctx context.Context, strategyID string, limit int) ([]strategy.Signal, error) {
	// In a real implementation, this would fetch from a database
	// For now, return empty slice
	return []strategy.Signal{}, nil
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
}
