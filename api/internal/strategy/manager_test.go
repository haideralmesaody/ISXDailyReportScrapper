package strategy

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	apperrors "github.com/isxcli/isxcli/internal/errors"
	"github.com/isxcli/isxcli/internal/liquidity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_RegisterStrategy(t *testing.T) {
	cases := []struct {
		name        string
		strategy    Strategy
		prepare     func(*Manager)
		wantErr     bool
		wantErrType apperrors.ErrorType
	}{
		{
			name:     "valid strategy registration",
			strategy: &mockStrategy{id: "test", name: "Test Strategy"},
		},
		{
			name:        "nil strategy",
			strategy:    nil,
			wantErr:     true,
			wantErrType: apperrors.ErrTypeValidation,
		},
		{
			name:     "duplicate strategy ID",
			strategy: &mockStrategy{id: "test", name: "Duplicate"},
			prepare: func(m *Manager) {
				_ = m.RegisterStrategy(&mockStrategy{id: "test", name: "Original"})
			},
			wantErr:     true,
			wantErrType: apperrors.ErrTypeValidation,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewManager()
			if tc.prepare != nil {
				tc.prepare(manager)
			}

			err := manager.RegisterStrategy(tc.strategy)

			if tc.wantErr {
				require.Error(t, err)
				var appErr *apperrors.AppError
				require.True(t, stdErrors.As(err, &appErr))
				assert.Equal(t, tc.wantErrType, appErr.Type)
				return
			}

			require.NoError(t, err)

			// For successful registration ensure strategy retrievable
			if tc.strategy != nil {
				got, getErr := manager.GetStrategy(context.Background(), tc.strategy.ID())
				require.NoError(t, getErr)
				assert.Equal(t, tc.strategy, got)
			}
		})
	}
}

func TestManager_GetStrategy(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()
	strategy := &mockStrategy{id: "test", name: "Test Strategy"}

	// Non-existent
	_, err := manager.GetStrategy(ctx, "nonexistent")
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, stdErrors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrTypeValidation, appErr.Type)

	// Register and retrieve
	require.NoError(t, manager.RegisterStrategy(strategy))

	retrieved, err := manager.GetStrategy(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, strategy, retrieved)
}

func TestManager_ListStrategies(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	infos := manager.ListStrategies(ctx)
	assert.Empty(t, infos)

	s1 := &mockStrategy{id: "strategy1", name: "Strategy 1", desc: "first"}
	s2 := &mockStrategy{id: "strategy2", name: "Strategy 2", desc: "second"}

	require.NoError(t, manager.RegisterStrategy(s1))
	require.NoError(t, manager.RegisterStrategy(s2))

	infos = manager.ListStrategies(ctx)
	require.Len(t, infos, 2)

	ids := map[string]struct{}{}
	for _, info := range infos {
		ids[info.ID] = struct{}{}
		if info.ID == s1.ID() {
			assert.Equal(t, s1.Name(), info.Name)
			assert.Equal(t, s1.Description(), info.Description)
		}
		if info.ID == s2.ID() {
			assert.Equal(t, s2.Name(), info.Name)
			assert.Equal(t, s2.Description(), info.Description)
		}
	}
	assert.Contains(t, ids, "strategy1")
	assert.Contains(t, ids, "strategy2")
}

func TestManager_Execute(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	_, err := manager.Execute(ctx, "nonexistent", nil)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, stdErrors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrTypeValidation, appErr.Type)

	strategy := &mockStrategy{
		id:   "test",
		name: "Test Strategy",
		signal: Signal{
			ID:         "signal-1",
			StrategyID: "test",
			Symbol:     "BBNI",
			Action:     SignalBuy,
			Strength:   75.5,
			Price:      1500.0,
			Timestamp:  time.Now(),
			Reasoning:  "Test signal",
		},
	}
	require.NoError(t, manager.RegisterStrategy(strategy))

	data := []liquidity.TradingDay{
		{Symbol: "BBNI", Date: time.Now(), Close: 1500.0, Volume: 1_000_000},
	}

	signal, err := manager.Execute(ctx, "test", data)
	require.NoError(t, err)
	assert.Equal(t, strategy.signal, signal)
	assert.Equal(t, data, strategy.lastData)
}

func TestManager_Backtest(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	_, err := manager.Backtest(ctx, "missing", nil, BacktestConfig{})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, stdErrors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrTypeValidation, appErr.Type)

	strategy := &mockStrategy{
		id:   "test",
		name: "Test Strategy",
		result: BacktestResult{
			StrategyID:       "test",
			Symbol:           "BBNI",
			StartDate:        time.Now().AddDate(0, 0, -30),
			EndDate:          time.Now(),
			TotalReturn:      0.15,
			AnnualizedReturn: 0.18,
			Volatility:       0.25,
			SharpeRatio:      0.72,
			MaxDrawdown:      -0.08,
			WinRate:          0.65,
			TotalTrades:      25,
		},
	}
	require.NoError(t, manager.RegisterStrategy(strategy))

	data := []liquidity.TradingDay{
		{Symbol: "BBNI", Date: time.Now().AddDate(0, 0, -30), Close: 1400.0, Volume: 1_000_000},
		{Symbol: "BBNI", Date: time.Now(), Close: 1500.0, Volume: 1_200_000},
	}
	config := BacktestConfig{
		StartDate:   time.Now().AddDate(0, 0, -30),
		EndDate:     time.Now(),
		InitialCash: 100_000.0,
		Commission:  0.001,
		Slippage:    0.001,
	}

	result, err := manager.Backtest(ctx, "test", data, config)
	require.NoError(t, err)
	assert.Equal(t, strategy.result, result)
	assert.Equal(t, data, strategy.lastData)
	assert.Equal(t, config, strategy.lastConfig)
}

func TestManager_Unregister(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	require.Error(t, manager.Unregister(ctx, "missing"))

	strategy := &mockStrategy{id: "to-remove", name: "Remove"}
	require.NoError(t, manager.RegisterStrategy(strategy))

	require.NoError(t, manager.Unregister(ctx, "to-remove"))
	_, err := manager.GetStrategy(ctx, "to-remove")
	require.Error(t, err)
}

func TestManager_ConcurrentExecute(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	strategy := &mockStrategy{
		id:     "concurrent",
		name:   "Concurrent Strategy",
		signal: Signal{ID: "signal", StrategyID: "concurrent", Action: SignalHold},
	}
	require.NoError(t, manager.RegisterStrategy(strategy))

	data := []liquidity.TradingDay{
		{Symbol: "BBNI", Date: time.Now(), Close: 1500.0, Volume: 1_000_000},
	}

	const operations = 50
	errs := make(chan error, operations)

	for i := 0; i < operations; i++ {
		go func() {
			_, err := manager.Execute(ctx, "concurrent", data)
			errs <- err
		}()
	}

	for i := 0; i < operations; i++ {
		err := <-errs
		assert.NoError(t, err)
	}
}

// Mock strategy implementation for testing
// (rest of file remains unchanged)
type mockStrategy struct {
	id          string
	name        string
	desc        string
	params      StrategyParams
	signal      Signal
	result      BacktestResult
	lastData    []liquidity.TradingDay
	lastParams  StrategyParams
	lastConfig  BacktestConfig
	validateErr error
	executeErr  error
	backtestErr error
}

func (m *mockStrategy) ID() string                 { return m.id }
func (m *mockStrategy) Name() string               { return m.name }
func (m *mockStrategy) Description() string        { return m.desc }
func (m *mockStrategy) Parameters() StrategyParams { return m.params }

func (m *mockStrategy) Execute(ctx context.Context, data []liquidity.TradingDay) (Signal, error) {
	m.lastData = data
	if m.executeErr != nil {
		return Signal{}, m.executeErr
	}
	return m.signal, nil
}

func (m *mockStrategy) Validate(ctx context.Context, params StrategyParams) error {
	m.lastParams = params
	return m.validateErr
}

func (m *mockStrategy) Backtest(ctx context.Context, data []liquidity.TradingDay, config BacktestConfig) (BacktestResult, error) {
	m.lastData = data
	m.lastConfig = config
	if m.backtestErr != nil {
		return BacktestResult{}, m.backtestErr
	}
	return m.result, nil
}
