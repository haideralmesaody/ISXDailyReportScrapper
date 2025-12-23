package strategy

import (
	"encoding/json"
	"fmt"
	"time"
)

// Config holds strategy system configuration
type Config struct {
	// Execution settings
	MaxConcurrentStrategies int           `json:"max_concurrent_strategies"`
	ExecutionTimeout        time.Duration `json:"execution_timeout"`
	BacktestTimeout         time.Duration `json:"backtest_timeout"`

	// Data requirements
	MinDataPoints int `json:"min_data_points"`
	MaxDataPoints int `json:"max_data_points"`

	// Signal settings
	SignalValidityDuration time.Duration `json:"signal_validity_duration"`
	MaxSignalsPerDay       int           `json:"max_signals_per_day"`

	// Risk management
	MaxPositionSize float64 `json:"max_position_size"`
	MaxDailyLoss    float64 `json:"max_daily_loss"`

	// Performance
	CacheEnabled bool          `json:"cache_enabled"`
	CacheTTL     time.Duration `json:"cache_ttl"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrentStrategies: 10,
		ExecutionTimeout:        30 * time.Second,
		BacktestTimeout:         5 * time.Minute,
		MinDataPoints:           20,
		MaxDataPoints:           1000,
		SignalValidityDuration:  24 * time.Hour,
		MaxSignalsPerDay:        50,
		MaxPositionSize:         0.1, // 10% of portfolio
		MaxDailyLoss:            0.05, // 5% daily loss limit
		CacheEnabled:            true,
		CacheTTL:                1 * time.Hour,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.MaxConcurrentStrategies <= 0 {
		return fmt.Errorf("max_concurrent_strategies must be positive")
	}

	if c.ExecutionTimeout <= 0 {
		return fmt.Errorf("execution_timeout must be positive")
	}

	if c.MinDataPoints <= 0 {
		return fmt.Errorf("min_data_points must be positive")
	}

	if c.MaxDataPoints < c.MinDataPoints {
		return fmt.Errorf("max_data_points must be >= min_data_points")
	}

	if c.MaxPositionSize <= 0 || c.MaxPositionSize > 1 {
		return fmt.Errorf("max_position_size must be between 0 and 1")
	}

	if c.MaxDailyLoss <= 0 || c.MaxDailyLoss > 1 {
		return fmt.Errorf("max_daily_loss must be between 0 and 1")
	}

	return nil
}

// ToJSON serializes config to JSON
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// FromJSON deserializes config from JSON
func (c *Config) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}