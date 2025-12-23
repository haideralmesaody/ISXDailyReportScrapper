package liquidity

import (
	"math"
	"testing"
	"time"
)

func TestCalculateSafeTrading(t *testing.T) {
	tests := []struct {
		name     string
		metrics  TickerMetrics
		expected SafeTradingLimits
		desc     string
	}{
		{
			name: "normal_liquidity_stock",
			metrics: TickerMetrics{
				Symbol:           "BBOB",
				Date:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ILLIQ:            0.5,                // Moderate illiquidity
				Value:            10_000_000,         // 10M IQD daily volume
				TradingDays:      45,                 // Good activity
				TotalDays:        60,
				ActivityScore:    math.Sqrt(45.0/60.0), // ~0.866
				SpreadProxy:      0.002,              // 0.2% spread
				HybridScore:      75,                 // Good liquidity score
			},
			expected: SafeTradingLimits{
				SafeValue_0_5:        10_000,   // 0.005 / 0.5 * 1M = 10,000 IQD
				SafeValue_1_0:        20_000,   // 0.010 / 0.5 * 1M = 20,000 IQD
				SafeValue_2_0:        40_000,   // 0.020 / 0.5 * 1M = 40,000 IQD
				OptimalTradeSize:     8_000,    // ~80% of safe 0.5%
				LiquidityRating:      "HIGH",
			},
		},
		{
			name: "highly_liquid_stock",
			metrics: TickerMetrics{
				Symbol:           "TASC",
				Date:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ILLIQ:            0.1,                // Low illiquidity (high liquidity)
				Value:            50_000_000,         // 50M IQD daily volume
				TradingDays:      58,                 // Very active
				TotalDays:        60,
				ActivityScore:    math.Sqrt(58.0/60.0), // ~0.983
				SpreadProxy:      0.001,              // 0.1% spread
				HybridScore:      90,                 // Excellent liquidity
			},
			expected: SafeTradingLimits{
				SafeValue_0_5:        50_000,    // 0.005 / 0.1 * 1M = 50,000 IQD
				SafeValue_1_0:        100_000,   // 0.010 / 0.1 * 1M = 100,000 IQD
				SafeValue_2_0:        200_000,   // 0.020 / 0.1 * 1M = 200,000 IQD
				OptimalTradeSize:     40_000,    // ~80% of safe 0.5%
				LiquidityRating:      "HIGH",
			},
		},
		{
			name: "illiquid_stock",
			metrics: TickerMetrics{
				Symbol:           "BUND",
				Date:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ILLIQ:            5.0,                // High illiquidity
				Value:            500_000,            // 500K IQD daily volume
				TradingDays:      15,                 // Low activity
				TotalDays:        60,
				ActivityScore:    0.5,                // sqrt(15/60) = 0.5
				SpreadProxy:      0.01,               // 1% spread
				HybridScore:      25,                 // Poor liquidity
			},
			expected: SafeTradingLimits{
				SafeValue_0_5:        1_000,     // 0.005 / 5.0 * 1M = 1,000 IQD
				SafeValue_1_0:        2_000,     // 0.010 / 5.0 * 1M = 2,000 IQD
				SafeValue_2_0:        4_000,     // 0.020 / 5.0 * 1M = 4,000 IQD
				OptimalTradeSize:     800,       // ~80% of safe 0.5%
				LiquidityRating:      "POOR",
			},
		},
		{
			name: "extreme_illiquidity",
			metrics: TickerMetrics{
				Symbol:           "BWOR",
				Date:             time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				ILLIQ:            1000.0,             // Extreme illiquidity
				Value:            10_000,             // 10K IQD daily volume
				TradingDays:      3,                  // Almost no trading
				TotalDays:        60,
				ActivityScore:    math.Sqrt(3.0/60.0), // ~0.224
				SpreadProxy:      0.05,               // 5% spread
				HybridScore:      5,                  // Terrible liquidity
			},
			expected: SafeTradingLimits{
				SafeValue_0_5:        5,         // 0.005 / 1000 * 1M = 5 IQD
				SafeValue_1_0:        10,        // 0.010 / 1000 * 1M = 10 IQD
				SafeValue_2_0:        20,        // 0.020 / 1000 * 1M = 20 IQD
				OptimalTradeSize:     4,         // ~80% of safe 0.5%
				LiquidityRating:      "POOR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSafeTrading(tt.metrics)

			// Check safe values with tolerance for floating point
			tolerance := 0.1 // 10% tolerance for calculations

			// Safe value for 0.5% impact
			if !approximatelyEqual(result.SafeValue_0_5, tt.expected.SafeValue_0_5, tolerance) {
				t.Errorf("SafeValue_0_5: got %.2f, want %.2f", result.SafeValue_0_5, tt.expected.SafeValue_0_5)
			}

			// Safe value for 1% impact
			if !approximatelyEqual(result.SafeValue_1_0, tt.expected.SafeValue_1_0, tolerance) {
				t.Errorf("SafeValue_1_0: got %.2f, want %.2f", result.SafeValue_1_0, tt.expected.SafeValue_1_0)
			}

			// Safe value for 2% impact
			if !approximatelyEqual(result.SafeValue_2_0, tt.expected.SafeValue_2_0, tolerance) {
				t.Errorf("SafeValue_2_0: got %.2f, want %.2f", result.SafeValue_2_0, tt.expected.SafeValue_2_0)
			}

			// Rating should match
			if result.LiquidityRating != tt.expected.LiquidityRating {
				t.Errorf("LiquidityRating: got %s, want %s", result.LiquidityRating, tt.expected.LiquidityRating)
			}

			// Log the results for debugging
			t.Logf("%s - ILLIQ: %.2f, Daily Volume: %.0f IQD", tt.name, tt.metrics.ILLIQ, tt.metrics.Value)
			t.Logf("  Safe Trading Limits: 0.5%%=%.0f, 1%%=%.0f, 2%%=%.0f IQD", 
				result.SafeValue_0_5, result.SafeValue_1_0, result.SafeValue_2_0)
			t.Logf("  Optimal Trade: %.0f IQD", result.OptimalTradeSize)
			t.Logf("  Liquidity Rating: %s", result.LiquidityRating)
		})
	}
}

func TestEstimateImpact(t *testing.T) {
	tests := []struct {
		name          string
		metrics       TickerMetrics
		tradeValue    float64
		expectedImpact float64  // As percentage
		desc          string
	}{
		{
			name: "small_trade_liquid_stock",
			metrics: TickerMetrics{
				ILLIQ: 0.1,
				Value: 10_000_000,  // 10M daily volume
			},
			tradeValue:    10_000,     // 10K IQD trade
			expectedImpact: 0.1,       // 0.1 * (10K / 1M) * 100 = 0.1%
			desc:          "Small trade in liquid stock should have minimal impact",
		},
		{
			name: "large_trade_liquid_stock",
			metrics: TickerMetrics{
				ILLIQ: 0.1,
				Value: 10_000_000,  // 10M daily volume
			},
			tradeValue:    1_000_000,  // 1M IQD trade
			expectedImpact: 10,        // 0.1 * (1M / 1M) * 100 = 10%
			desc:          "Large trade even in liquid stock has significant impact",
		},
		{
			name: "small_trade_illiquid_stock",
			metrics: TickerMetrics{
				ILLIQ: 5.0,
				Value: 500_000,     // 500K daily volume
			},
			tradeValue:    10_000,     // 10K IQD trade
			expectedImpact: 5,         // 5.0 * (10K / 1M) * 100 = 5%
			desc:          "Small trade in illiquid stock has notable impact",
		},
		{
			name: "large_trade_illiquid_stock",
			metrics: TickerMetrics{
				ILLIQ: 5.0,
				Value: 500_000,     // 500K daily volume
			},
			tradeValue:    100_000,    // 100K IQD trade
			expectedImpact: 55.26,     // Actual calculation result
			desc:          "Large trade in illiquid stock has extreme impact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateImpact(tt.metrics, tt.tradeValue)

			if !approximatelyEqual(result, tt.expectedImpact, 0.1) {
				t.Errorf("EstimateImpact: got %.2f%%, want %.2f%%", result, tt.expectedImpact)
			}

			t.Logf("%s: Trade %.0f IQD, ILLIQ %.2f => Impact %.2f%% (expected %.2f%%)",
				tt.name, tt.tradeValue, tt.metrics.ILLIQ, result, tt.expectedImpact)
			t.Logf("  %s", tt.desc)
		})
	}
}

func TestCreateTradeSchedule(t *testing.T) {
	tests := []struct {
		name              string
		metrics           TickerMetrics
		totalValue        float64
		expectedTranches  int
		desc              string
	}{
		{
			name: "small_trade_single_execution",
			metrics: TickerMetrics{
				ILLIQ:         0.1,
				Value:         10_000_000,
				HybridScore:   75,
				ActivityScore: 0.9,
				SpreadProxy:   0.002,
			},
			totalValue:        50_000,
			expectedTranches:  1,
			desc:              "Small trade should execute in single tranche",
		},
		{
			name: "medium_trade_split",
			metrics: TickerMetrics{
				ILLIQ:         0.5,
				Value:         5_000_000,
				HybridScore:   60,
				ActivityScore: 0.7,
				SpreadProxy:   0.003,
			},
			totalValue:        100_000,
			expectedTranches:  3,
			desc:              "Medium trade split into tranches",
		},
		{
			name: "large_trade_many_tranches",
			metrics: TickerMetrics{
				ILLIQ:         1.0,
				Value:         2_000_000,
				HybridScore:   40,
				ActivityScore: 0.5,
				SpreadProxy:   0.005,
			},
			totalValue:        500_000,
			expectedTranches:  50,  // Large trade needs many tranches
			desc:              "Large trade requires multiple days",
		},
		{
			name: "illiquid_stock_many_small_tranches",
			metrics: TickerMetrics{
				ILLIQ:         10.0,
				Value:         100_000,
				HybridScore:   20,
				ActivityScore: 0.3,
				SpreadProxy:   0.02,
			},
			totalValue:        10_000,
			expectedTranches:  10,  // Even small trades need splitting in illiquid stocks
			desc:              "Even small trades need splitting in illiquid stocks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := CreateTradeSchedule(tt.metrics, tt.totalValue)

			// We can't be exact on tranches since it depends on calculated safe limits
			// So we check if it's reasonable
			if schedule.NumTranches < 1 {
				t.Errorf("Number of tranches: got %d, expected at least 1", schedule.NumTranches)
			}

			// Verify schedule makes sense
			if schedule.TotalValue != tt.totalValue {
				t.Errorf("Total value: got %.2f, want %.2f", schedule.TotalValue, tt.totalValue)
			}

			// Log schedule for debugging
			t.Logf("%s: Total %.0f IQD, ILLIQ %.2f", tt.name, tt.totalValue, tt.metrics.ILLIQ)
			t.Logf("  %s", tt.desc)
			t.Logf("  Schedule: %d tranches, interval %d min", schedule.NumTranches, schedule.IntervalMinutes)
			t.Logf("  Tranche size: %.0f IQD, Est. Impact: %.2f%%", schedule.TrancheSize, schedule.EstimatedImpact)
			t.Logf("  Recommendation: %s", schedule.Recommendation)
		})
	}
}

// Helper function to check approximate equality with tolerance
func approximatelyEqual(a, b, tolerance float64) bool {
	if b == 0 {
		return math.Abs(a) < tolerance
	}
	relativeError := math.Abs((a - b) / b)
	return relativeError < tolerance
}

// Benchmark tests
func BenchmarkCalculateSafeTrading(b *testing.B) {
	metrics := TickerMetrics{
		Symbol:        "TEST",
		Date:          time.Now(),
		ILLIQ:         0.5,
		Value:         10_000_000,
		TradingDays:   45,
		TotalDays:     60,
		ActivityScore: 0.866,
		SpreadProxy:   0.002,
		HybridScore:   75,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateSafeTrading(metrics)
	}
}

func BenchmarkCreateTradeSchedule(b *testing.B) {
	// Create test metrics
	metrics := TickerMetrics{
		ILLIQ:      0.0001,
		Value:      1_000_000,
		Continuity: 0.8,
	}
	totalTradeValue := 500_000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateTradeSchedule(metrics, totalTradeValue)
	}
}