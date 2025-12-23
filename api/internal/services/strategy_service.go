package services

import (
	"context"
	"fmt"
	"log/slog"
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
	// This would integrate with the existing data service
	// For now, create mock data based on the existing pattern
	if s.dataService != nil {
		// Try to use existing data service method if available
		return s.getMockData(symbol, dataPoints), nil
	}

	return s.getMockData(symbol, dataPoints), nil
}

// getHistoricalData retrieves historical trading data
func (s *StrategyService) getHistoricalData(ctx context.Context, symbol string, start, end time.Time) ([]liquidity.TradingDay, error) {
	days := int(end.Sub(start).Hours() / 24)
	return s.getMockData(symbol, days), nil
}

// getMockData creates mock trading data for testing
func (s *StrategyService) getMockData(symbol string, dataPoints int) []liquidity.TradingDay {
	data := make([]liquidity.TradingDay, dataPoints)
	basePrice := 100.0

	for i := 0; i < dataPoints; i++ {
		// Create trending data with some noise
		trend := float64(i) * 0.002 // Small upward trend
		noise := float64((i%7)-3) * 0.01 // Random noise
		price := basePrice * (1 + trend + noise)

		data[i] = liquidity.TradingDay{
			Date:          time.Now().AddDate(0, 0, -dataPoints+i),
			Symbol:        symbol,
			Open:          price * 0.99,
			High:          price * 1.02,
			Low:           price * 0.98,
			Close:         price,
			Volume:        10000 + float64(i*100),
			Value:         price * (10000 + float64(i*100)),
			NumTrades:     50 + i,
			TradingStatus: "ACTIVE",
		}
	}

	return data
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
	StrategyID string           `json:"strategy_id"`
	Symbol     string           `json:"symbol"`
	Success    bool             `json:"success"`
	Signal     strategy.Signal  `json:"signal,omitempty"`
	Error      string           `json:"error,omitempty"`
}