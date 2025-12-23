package strategies

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
	"github.com/isxcli/isxcli/internal/strategy"
)

// BaseStrategy provides common functionality for all strategies
type BaseStrategy struct {
	id          string
	name        string
	description string
	version     string
	logger      *slog.Logger
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(id, name, description string, logger *slog.Logger) *BaseStrategy {
	return &BaseStrategy{
		id:          id,
		name:        name,
		description: description,
		version:     "1.0.0",
		logger:      logger,
	}
}

// ID returns the strategy ID
func (b *BaseStrategy) ID() string {
	return b.id
}

// Name returns the strategy name
func (b *BaseStrategy) Name() string {
	return b.name
}

// Description returns the strategy description
func (b *BaseStrategy) Description() string {
	return b.description
}

// Version returns the strategy version
func (b *BaseStrategy) Version() string {
	return b.version
}

// GenerateSignalID creates a unique signal ID
func (b *BaseStrategy) GenerateSignalID() string {
	timestamp := time.Now().UnixNano()
	random, _ := rand.Int(rand.Reader, big.NewInt(1000))
	return fmt.Sprintf("%s_%d_%d", b.id, timestamp, random.Int64())
}

// ValidateData checks if the provided data is sufficient
func (b *BaseStrategy) ValidateData(ctx context.Context, data []liquidity.TradingDay, minPoints int) error {
	if len(data) < minPoints {
		return strategy.NewInsufficientDataError(minPoints, len(data))
	}

	// Check for valid trading days
	validDays := 0
	for _, day := range data {
		if day.IsValid() && day.IsTrading() {
			validDays++
		}
	}

	if validDays < minPoints {
		return strategy.NewInsufficientDataError(minPoints, validDays)
	}

	return nil
}

// CalculateSimpleMovingAverage calculates SMA for the given period
func (b *BaseStrategy) CalculateSimpleMovingAverage(prices []float64, period int) []float64 {
	if len(prices) < period {
		return nil
	}

	sma := make([]float64, len(prices))

	for i := period - 1; i < len(prices); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += prices[j]
		}
		sma[i] = sum / float64(period)
	}

	return sma
}

// CalculateStandardDeviation calculates standard deviation
func (b *BaseStrategy) CalculateStandardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))

	return math.Sqrt(variance)
}

// CalculateReturns calculates daily returns from trading data
func (b *BaseStrategy) CalculateReturns(data []liquidity.TradingDay) []float64 {
	if len(data) < 2 {
		return nil
	}

	returns := make([]float64, len(data)-1)
	for i := 1; i < len(data); i++ {
		if data[i-1].Close > 0 {
			returns[i-1] = (data[i].Close - data[i-1].Close) / data[i-1].Close
		}
	}

	return returns
}

// CreateSignal creates a standardized signal
func (b *BaseStrategy) CreateSignal(
	symbol string,
	action strategy.SignalAction,
	strength float64,
	price float64,
	reasoning string,
	metadata map[string]interface{},
) strategy.Signal {
	return strategy.Signal{
		ID:         b.GenerateSignalID(),
		StrategyID: b.id,
		Symbol:     symbol,
		Action:     action,
		Strength:   strength,
		Price:      price,
		Quantity:   0, // To be calculated by position sizing
		Timestamp:  time.Now(),
		Reasoning:  reasoning,
		Metadata:   metadata,
		ValidUntil: time.Now().Add(24 * time.Hour),
	}
}

// LogExecution logs strategy execution details
func (b *BaseStrategy) LogExecution(ctx context.Context, symbol string, signal strategy.Signal) {
	b.logger.InfoContext(ctx, "strategy executed",
		"strategy_id", b.id,
		"symbol", symbol,
		"action", signal.Action,
		"strength", signal.Strength,
		"price", signal.Price,
		"reasoning", signal.Reasoning,
	)
}