package strategy

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/isxcli/isxcli/internal/liquidity"
)

// Manager handles strategy registration, execution, and lifecycle
type Manager struct {
	strategies map[string]Strategy
	mu         sync.RWMutex
	logger     *slog.Logger
}

// NewManager creates a new strategy manager
func NewManager() *Manager {
	return &Manager{
		strategies: make(map[string]Strategy),
		logger:     slog.Default(),
	}
}

// RegisterStrategy adds a strategy to the manager
func (m *Manager) RegisterStrategy(strategy Strategy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strategy == nil {
		return NewInvalidStrategyError("strategy cannot be nil")
	}

	id := strategy.ID()
	if id == "" {
		return NewInvalidStrategyError("strategy ID cannot be empty")
	}

	if _, exists := m.strategies[id]; exists {
		return NewStrategyExistsError(id)
	}

	m.strategies[id] = strategy
	m.logger.InfoContext(context.Background(), "strategy registered",
		"strategy_id", id,
		"strategy_name", strategy.Name(),
	)

	return nil
}

// GetStrategy retrieves a strategy by ID
func (m *Manager) GetStrategy(ctx context.Context, id string) (Strategy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	strategy, exists := m.strategies[id]
	if !exists {
		return nil, NewStrategyNotFoundError(id)
	}

	return strategy, nil
}

// ListStrategies returns all registered strategies
func (m *Manager) ListStrategies(ctx context.Context) []StrategyInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]StrategyInfo, 0, len(m.strategies))
	for _, strategy := range m.strategies {
		info := StrategyInfo{
			ID:          strategy.ID(),
			Name:        strategy.Name(),
			Description: strategy.Description(),
		}
		infos = append(infos, info)
	}

	return infos
}

// Execute runs a strategy against the provided data
func (m *Manager) Execute(ctx context.Context, strategyID string, data []liquidity.TradingDay) (Signal, error) {
	strategy, err := m.GetStrategy(ctx, strategyID)
	if err != nil {
		return Signal{}, fmt.Errorf("get strategy %s: %w", strategyID, err)
	}

	m.logger.InfoContext(ctx, "executing strategy",
		"strategy_id", strategyID,
		"data_points", len(data),
	)

	signal, err := strategy.Execute(ctx, data)
	if err != nil {
		m.logger.ErrorContext(ctx, "strategy execution failed",
			"strategy_id", strategyID,
			"error", err,
		)
		return Signal{}, fmt.Errorf("execute strategy %s: %w", strategyID, err)
	}

	m.logger.InfoContext(ctx, "strategy execution completed",
		"strategy_id", strategyID,
		"signal_action", signal.Action,
		"signal_strength", signal.Strength,
	)

	return signal, nil
}

// Backtest runs a backtest for a strategy
func (m *Manager) Backtest(ctx context.Context, strategyID string, data []liquidity.TradingDay, config BacktestConfig) (BacktestResult, error) {
	strategy, err := m.GetStrategy(ctx, strategyID)
	if err != nil {
		return BacktestResult{}, fmt.Errorf("get strategy %s: %w", strategyID, err)
	}

	m.logger.InfoContext(ctx, "starting backtest",
		"strategy_id", strategyID,
		"start_date", config.StartDate,
		"end_date", config.EndDate,
		"initial_cash", config.InitialCash,
	)

	result, err := strategy.Backtest(ctx, data, config)
	if err != nil {
		return BacktestResult{}, fmt.Errorf("backtest strategy %s: %w", strategyID, err)
	}

	m.logger.InfoContext(ctx, "backtest completed",
		"strategy_id", strategyID,
		"total_return", result.TotalReturn,
		"sharpe_ratio", result.SharpeRatio,
		"total_trades", result.TotalTrades,
	)

	return result, nil
}

// Unregister removes a strategy from the manager
func (m *Manager) Unregister(ctx context.Context, strategyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.strategies[strategyID]; !exists {
		return NewStrategyNotFoundError(strategyID)
	}

	delete(m.strategies, strategyID)
	m.logger.InfoContext(ctx, "strategy unregistered", "strategy_id", strategyID)

	return nil
}