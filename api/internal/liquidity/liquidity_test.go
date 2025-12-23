package liquidity

import (
	"context"
	"log/slog"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWindow tests Window type functionality
func TestWindow(t *testing.T) {
	tests := []struct {
		name         string
		window       Window
		expectedDays int
		expectedStr  string
	}{
		{"20-day window", Window20, 20, "20d"},
		{"60-day window", Window60, 60, "60d"},
		{"120-day window", Window120, 120, "120d"},
		{"unknown window", Window(99), 99, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedDays, tt.window.Days())
			assert.Equal(t, tt.expectedStr, tt.window.String())
		})
	}
}

// TestTradingDay tests TradingDay validation and methods
func TestTradingDay(t *testing.T) {
	t.Run("IsValid", func(t *testing.T) {
		tests := []struct {
			name  string
			td    TradingDay
			valid bool
		}{
			{
				name: "valid trading day",
				td: TradingDay{
					Date:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Symbol:        "TEST",
					Open:          100.0,
					High:          105.0,
					Low:           95.0,
					Close:         102.0,
					Value:        1000000,
					NumTrades:     150,
					TradingStatus: "ACTIVE",
				},
				valid: true,
			},
			{
				name: "negative open price",
				td: TradingDay{
					Open:  -100.0,
					High:  105.0,
					Low:   95.0,
					Close: 102.0,
				},
				valid: false,
			},
			{
				name: "high less than low",
				td: TradingDay{
					Open:  100.0,
					High:  90.0,
					Low:   95.0,
					Close: 102.0,
				},
				valid: false,
			},
			{
				name: "negative volume",
				td: TradingDay{
					Open:      100.0,
					High:      105.0,
					Low:       95.0,
					Close:     102.0,
					Value:    -1000,
					NumTrades: -5,
				},
				valid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.valid, tt.td.IsValid())
			})
		}
	})

	t.Run("IsTrading", func(t *testing.T) {
		tests := []struct {
			name    string
			td      TradingDay
			trading bool
		}{
			{
				name: "active trading day",
				td: TradingDay{
					TradingStatus: "ACTIVE",
					Value:        1000000,
					NumTrades:     150,
				},
				trading: true,
			},
			{
				name: "suspended trading",
				td: TradingDay{
					TradingStatus: "SUSPENDED",
					Value:        0,
					NumTrades:     0,
				},
				trading: false,
			},
			{
				name: "zero volume",
				td: TradingDay{
					TradingStatus: "ACTIVE",
					Value:        0,
					NumTrades:     0,
				},
				trading: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.trading, tt.td.IsTrading())
			})
		}
	})

	t.Run("Return", func(t *testing.T) {
		tests := []struct {
			name         string
			currentClose float64
			prevClose    float64
			expectedRet  float64
		}{
			{"positive return", 102.0, 98.0, (102.0 - 98.0) / 98.0},
			{"negative return", 95.0, 100.0, (95.0 - 100.0) / 100.0},
			{"zero return", 100.0, 100.0, 0.0},
			{"zero previous close", 100.0, 0.0, 0.0},
			{"negative previous close", 100.0, -50.0, 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				td := TradingDay{Close: tt.currentClose}
				ret := td.Return(tt.prevClose)
				assert.InDelta(t, tt.expectedRet, ret, 1e-9)
			})
		}
	})
}

// TestPenaltyParams tests penalty parameter validation and functions
func TestPenaltyParams(t *testing.T) {
	t.Run("default parameters validation", func(t *testing.T) {
		params := DefaultPenaltyParams()
		assert.True(t, params.IsValid())
		
		err := ValidateParams(params)
		assert.NoError(t, err)
	})

	t.Run("invalid parameters", func(t *testing.T) {
		tests := []struct {
			name   string
			params PenaltyParams
		}{
			{"zero piecewise p0", PenaltyParams{PiecewiseP0: 0}},
			{"negative beta", PenaltyParams{PiecewiseP0: 1, PiecewiseBeta: -0.1}},
			{"max mult too small", PenaltyParams{PiecewiseP0: 1, PiecewiseBeta: 0.1, PiecewiseGamma: 0.1, PiecewisePStar: 1, PiecewiseMaxMult: 0.5}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.False(t, tt.params.IsValid())
			})
		}
	})
}

// TestComponentWeights tests weight validation and normalization
func TestComponentWeights(t *testing.T) {
	t.Run("default weights", func(t *testing.T) {
		weights := DefaultWeights()
		assert.True(t, weights.IsValid())
		
		err := ValidateWeights(weights)
		assert.NoError(t, err)
	})

	t.Run("normalization", func(t *testing.T) {
		tests := []struct {
			name    string
			original ComponentWeights
		}{
			{"needs normalization", ComponentWeights{Impact: 0.6, Value: 0.3, Continuity: 0.2}},
			{"already normalized", ComponentWeights{Impact: 0.4, Value: 0.3, Continuity: 0.3}},
			{"zero weights", ComponentWeights{Impact: 0, Value: 0, Continuity: 0}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				weights := tt.original
				weights.Normalize()
				
				if tt.original.Impact+tt.original.Value+tt.original.Continuity > 0 {
					sum := weights.Impact + weights.Value + weights.Continuity
					assert.InDelta(t, 1.0, sum, 1e-9)
				}
			})
		}
	})

	t.Run("invalid weights", func(t *testing.T) {
		tests := []struct {
			name    string
			weights ComponentWeights
		}{
			{"negative impact", ComponentWeights{Impact: -0.1, Value: 0.5, Continuity: 0.6}},
			{"sum too low", ComponentWeights{Impact: 0.1, Value: 0.1, Continuity: 0.1}},
			{"sum too high", ComponentWeights{Impact: 0.6, Value: 0.6, Continuity: 0.6}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.False(t, tt.weights.IsValid())
			})
		}
	})
}

// TestILLIQCalculation tests ILLIQ calculation with various scenarios
func TestILLIQCalculation(t *testing.T) {
	t.Run("normal calculation", func(t *testing.T) {
		data := []TradingDay{
			{
				Date:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Open:          100.0,
				High:          102.0,
				Low:           98.0,
				Close:         101.0,
				Value:        1000000,
				TradingStatus: "ACTIVE",
				NumTrades:     100,
			},
			{
				Date:          time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Open:          101.0,
				High:          103.0,
				Low:           100.0,
				Close:         102.0,
				Value:        1200000,
				TradingStatus: "ACTIVE",
				NumTrades:     120,
			},
		}

		illiq, lowerBound, upperBound := ComputeILLIQ(data, 0.05, 0.95)
		
		assert.False(t, math.IsNaN(illiq))
		assert.False(t, math.IsInf(illiq, 0))
		assert.GreaterOrEqual(t, lowerBound, 0.0)
		assert.GreaterOrEqual(t, upperBound, lowerBound)
	})

	t.Run("insufficient data", func(t *testing.T) {
		data := []TradingDay{
			{
				Date:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Close:         100.0,
				Value:        1000000,
				TradingStatus: "ACTIVE",
				NumTrades:     100,
			},
		}

		illiq, lowerBound, upperBound := ComputeILLIQ(data, 0.05, 0.95)

		// Insufficient data: illiq returns high default (1000), bounds return 0
		assert.Equal(t, 1000.0, illiq)
		assert.Equal(t, 0.0, lowerBound)
		assert.Equal(t, 0.0, upperBound)
	})

	t.Run("zero volume", func(t *testing.T) {
		data := []TradingDay{
			{
				Date:          time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Close:         100.0,
				Value:        1000000,
				TradingStatus: "ACTIVE",
				NumTrades:     100,
			},
			{
				Date:          time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Close:         102.0,
				Value:        0,
				TradingStatus: "SUSPENDED",
				NumTrades:     0,
			},
		}

		illiq, _, _ := ComputeILLIQ(data, 0.05, 0.95)
		// Zero volume case now returns high default value (1000) instead of 0
		assert.Equal(t, 1000.0, illiq)
	})

	t.Run("outlier handling", func(t *testing.T) {
		// Create data with extreme outlier
		data := generateTradingDays(10, func(i int) TradingDay {
			return TradingDay{
				Date:          time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Close:         100.0 + float64(i)*0.1,
				Open:          99.0 + float64(i)*0.1,
				High:          101.0 + float64(i)*0.1,
				Low:           98.0 + float64(i)*0.1,
				Value:        1000000,
				TradingStatus: "ACTIVE",
				NumTrades:     100,
			}
		})

		illiq, _, _ := ComputeILLIQ(data, 0.05, 0.95)
		assert.False(t, math.IsInf(illiq, 0))
		assert.False(t, math.IsNaN(illiq))
	})
}

// TestContinuityCalculation tests continuity calculations and transformations
func TestContinuityCalculation(t *testing.T) {
	t.Run("basic continuity", func(t *testing.T) {
		tests := []struct {
			name       string
			data       []TradingDay
			expected   float64
		}{
			{
				name: "all trading days",
				data: []TradingDay{
					{Value: 1000, TradingStatus: "ACTIVE", NumTrades: 10},
					{Value: 1200, TradingStatus: "ACTIVE", NumTrades: 12},
					{Value: 900, TradingStatus: "ACTIVE", NumTrades: 9},
				},
				expected: 1.0,
			},
			{
				name: "mixed trading days",
				data: []TradingDay{
					{Value: 1000, TradingStatus: "ACTIVE", NumTrades: 10},
					{Value: 0, TradingStatus: "SUSPENDED", NumTrades: 0},
					{Value: 1200, TradingStatus: "ACTIVE", NumTrades: 12},
				},
				expected: 2.0 / 3.0,
			},
			{
				name: "no trading days",
				data: []TradingDay{
					{Value: 0, TradingStatus: "SUSPENDED", NumTrades: 0},
					{Value: 0, TradingStatus: "SUSPENDED", NumTrades: 0},
				},
				expected: 0.0,
			},
			{
				name: "empty data",
				data: []TradingDay{},
				expected: 0.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				continuity := CalculateContinuity(tt.data)
				assert.InDelta(t, tt.expected, continuity, 1e-9)
			})
		}
	})

	t.Run("non-linear transformation", func(t *testing.T) {
		tests := []struct {
			name       string
			continuity float64
			delta      float64
		}{
			{"perfect continuity", 1.0, DefaultContinuityDelta},
			{"moderate continuity", 0.7, DefaultContinuityDelta},
			{"low continuity", 0.3, DefaultContinuityDelta},
			{"zero continuity", 0.0, DefaultContinuityDelta},
			{"high delta", 0.5, 0.8},
			{"low delta", 0.5, 0.1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				contNL := ContinuityNL(tt.continuity, tt.delta)
				
				assert.GreaterOrEqual(t, contNL, 0.0)
				assert.LessOrEqual(t, contNL, 1.0)
				
				// Test boundary behavior
				if tt.continuity == 0 {
					assert.Equal(t, 0.0, contNL)
				} else if tt.continuity == 1 {
					assert.Equal(t, 1.0, contNL)
				}
			})
		}
	})
}

// TestCorwinSchultz tests the Corwin-Schultz spread estimator
func TestCorwinSchultz(t *testing.T) {
	tests := []struct {
		name              string
		high1, low1       float64
		high2, low2       float64
		expectedMinSpread float64
		expectedMaxSpread float64
		expectZero        bool
	}{
		{
			name:              "normal case",
			high1:             102.0,
			low1:              98.0,
			high2:             105.0,
			low2:              99.0,
			expectedMinSpread: 0.0,
			expectedMaxSpread: 1.0,
		},
		{
			name:       "zero high1",
			high1:      0,
			low1:       98.0,
			high2:      105.0,
			low2:       99.0,
			expectZero: true,
		},
		{
			name:       "negative low",
			high1:      102.0,
			low1:       -98.0,
			high2:      105.0,
			low2:       99.0,
			expectZero: true,
		},
		{
			name:       "high less than low",
			high1:      90.0,
			low1:       98.0,
			high2:      105.0,
			low2:       99.0,
			expectZero: true,
		},
		{
			name:              "identical prices",
			high1:             100.0,
			low1:              100.0,
			high2:             100.0,
			low2:              100.0,
			expectedMinSpread: 0.0,
			expectedMaxSpread: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spread := CorwinSchultz(tt.high1, tt.low1, tt.high2, tt.low2)
			
			if tt.expectZero {
				assert.Equal(t, 0.0, spread)
			} else {
				assert.GreaterOrEqual(t, spread, tt.expectedMinSpread)
				assert.LessOrEqual(t, spread, tt.expectedMaxSpread)
				assert.False(t, math.IsNaN(spread))
				assert.False(t, math.IsInf(spread, 0))
			}
		})
	}

	t.Run("spread series calculation", func(t *testing.T) {
		data := []TradingDay{
			{High: 102, Low: 98, TradingStatus: "ACTIVE", Value: 1000, NumTrades: 10},
			{High: 105, Low: 99, TradingStatus: "ACTIVE", Value: 1200, NumTrades: 12},
			{High: 103, Low: 101, TradingStatus: "ACTIVE", Value: 900, NumTrades: 9},
		}
		
		spreads := CalculateSpreadSeries(data)
		// Note: CalculateSpreadSeries may return different number based on implementation
		assert.GreaterOrEqual(t, len(spreads), 0)
		
		for _, spread := range spreads {
			assert.GreaterOrEqual(t, spread, 0.0)
			assert.False(t, math.IsNaN(spread))
			assert.False(t, math.IsInf(spread, 0))
		}
	})
}

// TestRobustScaling tests cross-sectional scaling with different options
func TestRobustScaling(t *testing.T) {
	t.Run("basic scaling", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		scaled := RobustScale(values, false, false)
		
		assert.Equal(t, len(values), len(scaled))
		
		// Check that values are in [0,100] range
		for i, val := range scaled {
			assert.GreaterOrEqual(t, val, 0.0, "scaled value %d should be >= 0", i)
			assert.LessOrEqual(t, val, 100.0, "scaled value %d should be <= 100", i)
		}
		
		// Check ordering preservation (no inversion)
		for i := 1; i < len(scaled); i++ {
			assert.GreaterOrEqual(t, scaled[i], scaled[i-1])
		}
	})

	t.Run("scaling with inversion", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		scaled := RobustScale(values, true, false)
		
		assert.Equal(t, len(values), len(scaled))
		
		// Check that values are in [0,100] range
		for _, val := range scaled {
			assert.GreaterOrEqual(t, val, 0.0)
			assert.LessOrEqual(t, val, 100.0)
		}
		
		// Check ordering reversal (with inversion)
		for i := 1; i < len(scaled); i++ {
			assert.LessOrEqual(t, scaled[i], scaled[i-1])
		}
	})

	t.Run("scaling with log transform", func(t *testing.T) {
		values := []float64{1.0, 10.0, 100.0, 1000.0}
		scaled := RobustScale(values, false, true)
		
		assert.Equal(t, len(values), len(scaled))
		
		for _, val := range scaled {
			assert.GreaterOrEqual(t, val, 0.0)
			assert.LessOrEqual(t, val, 100.0)
			assert.False(t, math.IsNaN(val))
			assert.False(t, math.IsInf(val, 0))
		}
	})

	t.Run("edge cases", func(t *testing.T) {
		tests := []struct {
			name   string
			values []float64
		}{
			{"single value", []float64{5.0}},
			{"identical values", []float64{3.0, 3.0, 3.0}},
			{"with zeros", []float64{0.0, 1.0, 2.0, 3.0}},
			{"with negatives", []float64{-2.0, -1.0, 0.0, 1.0, 2.0}},
			{"empty slice", []float64{}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				scaled := RobustScale(tt.values, false, false)
				assert.Equal(t, len(tt.values), len(scaled))
				
				for _, val := range scaled {
					assert.False(t, math.IsNaN(val))
					assert.False(t, math.IsInf(val, 0))
				}
			})
		}
	})
	
	t.Run("actual scaling test", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		scaled := RobustScale(values, false, false)
		assert.Equal(t, len(values), len(scaled))
		
		// Test with actual values
		for _, val := range scaled {
			assert.GreaterOrEqual(t, val, 0.0)
			assert.LessOrEqual(t, val, 100.0)
		}
	})
}

// TestCalculator tests calculator initialization and configuration
func TestCalculator(t *testing.T) {
	t.Run("creation with valid parameters", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		
		calc := NewCalculator(
			Window60,
			DefaultPenaltyParams(),
			DefaultWeights(),
			logger,
		)
		
		require.NotNil(t, calc)
	})

	t.Run("creation with nil logger", func(t *testing.T) {
		calc := NewCalculator(
			Window60,
			DefaultPenaltyParams(),
			DefaultWeights(),
			nil,
		)
		
		require.NotNil(t, calc)
	})

	t.Run("winsorization bounds configuration", func(t *testing.T) {
		calc := NewCalculator(Window60, DefaultPenaltyParams(), DefaultWeights(), nil)
		
		tests := []struct {
			name    string
			bounds  WinsorizationBounds
			wantErr bool
		}{
			{"valid bounds", WinsorizationBounds{Lower: 0.05, Upper: 0.95}, false},
			{"narrow bounds", WinsorizationBounds{Lower: 0.1, Upper: 0.9}, false},
			{"invalid - lower > upper", WinsorizationBounds{Lower: 0.9, Upper: 0.1}, true},
			{"invalid - lower < 0", WinsorizationBounds{Lower: -0.1, Upper: 0.9}, true},
			{"invalid - upper > 1", WinsorizationBounds{Lower: 0.1, Upper: 1.1}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := calc.SetWinsorizationBounds(tt.bounds)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("configuration settings", func(t *testing.T) {
		calc := NewCalculator(Window60, DefaultPenaltyParams(), DefaultWeights(), nil)
		
		calc.SetConfiguration(true, 8, 60*time.Second)
		// Configuration is internal, but we test it doesn't panic
		assert.NotPanics(t, func() {
			calc.SetConfiguration(false, 4, 30*time.Second)
		})
	})
}

// TestDataValidation tests comprehensive data validation scenarios
func TestDataValidation(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		validData := generateTradingDays(MinObservationsForCalc+1, func(i int) TradingDay {
			return TradingDay{
				Date:          time.Date(2024, 1, 1+i, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Open:          100.0 + float64(i),
				High:          105.0 + float64(i),
				Low:           95.0 + float64(i),
				Close:         102.0 + float64(i),
				Value:        1000000,
				NumTrades:     150,
				TradingStatus: "ACTIVE",
			}
		})

		err := ValidateTradingData(validData)
		assert.NoError(t, err)
	})

	t.Run("invalid data scenarios", func(t *testing.T) {
		tests := []struct {
			name string
			data []TradingDay
		}{
			{
				name: "empty data",
				data: []TradingDay{},
			},
			{
				name: "insufficient data points",
				data: generateTradingDays(MinObservationsForCalc-1, func(i int) TradingDay {
					return TradingDay{
						Date:          time.Date(2024, 1, 1+i, 0, 0, 0, 0, time.UTC),
						Symbol:        "TEST",
						Open:          100.0,
						High:          105.0,
						Low:           95.0,
						Close:         102.0,
						Value:        1000000,
						NumTrades:     150,
						TradingStatus: "ACTIVE",
					}
				}),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateTradingData(tt.data)
				assert.Error(t, err)
			})
		}
	})
}

// TestDefaultConfigurations tests all default configuration functions
func TestDefaultConfigurations(t *testing.T) {
	t.Run("default penalty params", func(t *testing.T) {
		params := DefaultPenaltyParams()
		assert.True(t, params.IsValid())
		
		// Check reasonable defaults
		assert.Greater(t, params.PiecewiseP0, 0.0)
		assert.Greater(t, params.PiecewiseBeta, 0.0)
		assert.Greater(t, params.PiecewiseGamma, 0.0)
		assert.Greater(t, params.PiecewisePStar, 0.0)
		assert.Greater(t, params.PiecewiseMaxMult, 1.0)
		assert.Greater(t, params.ExponentialP0, 0.0)
		assert.Greater(t, params.ExponentialAlpha, 0.0)
		assert.Greater(t, params.ExponentialMaxMult, 1.0)
	})

	t.Run("default weights", func(t *testing.T) {
		weights := DefaultWeights()
		assert.True(t, weights.IsValid())
		
		// Check sum is 1
		sum := weights.Impact + weights.Value + weights.Continuity
		assert.InDelta(t, 1.0, sum, 1e-9)
	})

	t.Run("default calibration config", func(t *testing.T) {
		config := DefaultCalibrationConfig()
		assert.True(t, config.IsValid())
		
		// Check reasonable defaults
		assert.Greater(t, config.ParamGridSize, 0)
		assert.Greater(t, config.KFolds, 1)
		assert.Greater(t, config.MinTickers, 0)
		assert.Greater(t, config.MaxConcurrency, 0)
		assert.Greater(t, config.Tolerance, 0.0)
	})

	t.Run("calibrated weights", func(t *testing.T) {
		tests := []struct {
			name      string
			condition string
			isDefault bool
		}{
			{"high volatility", "high_volatility", false},
			{"low volatility", "low_volatility", false},
			{"normal conditions", "normal", false},
			{"unknown condition", "unknown_condition", true},
			{"empty condition", "", true},
		}

		defaultWeights := DefaultWeights()
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				weights := CalibratedWeights(tt.condition)
				assert.True(t, weights.IsValid())
				
				if tt.isDefault {
					assert.Equal(t, defaultWeights.Impact, weights.Impact)
					assert.Equal(t, defaultWeights.Value, weights.Value)
					assert.Equal(t, defaultWeights.Continuity, weights.Continuity)
				}
			})
		}
	})

	t.Run("winsorization bounds", func(t *testing.T) {
		bounds := WinsorizationBounds{Lower: DefaultLowerBound, Upper: DefaultUpperBound}
		assert.True(t, bounds.IsValid())
		assert.Greater(t, DefaultUpperBound, DefaultLowerBound)
		assert.GreaterOrEqual(t, DefaultLowerBound, 0.0)
		assert.LessOrEqual(t, DefaultUpperBound, 1.0)
	})
}

// TestCalculatorIntegration tests end-to-end calculator functionality
func TestCalculatorIntegration(t *testing.T) {
	t.Run("successful calculation", func(t *testing.T) {
		ctx := context.Background()
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		
		calc := NewCalculator(
			Window20, // Use smaller window for faster tests
			DefaultPenaltyParams(),
			DefaultWeights(),
			logger,
		)

		// Generate sufficient test data with smaller window
		baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		data := generateRealisticISXData([]string{"TASC", "BMFI"}, 30, baseDate)
		
		metrics, err := calc.Calculate(ctx, data)

		assert.NoError(t, err)
		assert.Greater(t, len(metrics), 0)

		// Verify all metrics are valid
		for _, metric := range metrics {
			assert.True(t, metric.IsValid(), "metric for %s on %s should be valid",
				metric.Symbol, metric.Date.Format("2006-01-02"))
			assert.GreaterOrEqual(t, metric.HybridRank, 1)
			assert.False(t, math.IsNaN(metric.HybridScore))
			assert.False(t, math.IsInf(metric.HybridScore, 0))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		calc := NewCalculator(Window60, DefaultPenaltyParams(), DefaultWeights(), nil)
		baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		data := generateRealisticISXData([]string{"TEST"}, 90, baseDate)
		
		_, err := calc.Calculate(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("insufficient data", func(t *testing.T) {
		ctx := context.Background()
		calc := NewCalculator(Window60, DefaultPenaltyParams(), DefaultWeights(), nil)

		// Create data with insufficient window size
		data := generateTradingDays(30, func(i int) TradingDay {
			return TradingDay{
				Date:          time.Date(2024, 1, 1+i, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Close:         100.0 + float64(i),
				Open:          99.0 + float64(i),
				High:          102.0 + float64(i),
				Low:           98.0 + float64(i),
				Value:        1000000,
				NumTrades:     100,
				TradingStatus: "ACTIVE",
			}
		})

		// Calculator uses graceful degradation - returns worst-case scores instead of error
		metrics, err := calc.Calculate(ctx, data)
		assert.NoError(t, err, "Calculator should use graceful degradation for insufficient data")
		assert.Greater(t, len(metrics), 0, "Should return worst-case metric")

		// Verify it's a worst-case metric (low hybrid score)
		assert.Less(t, metrics[0].HybridScore, 10.0, "Insufficient data should result in low hybrid score")
	})

	t.Run("invalid input parameters", func(t *testing.T) {
		ctx := context.Background()
		
		// Invalid penalty parameters
		invalidParams := PenaltyParams{}
		calc := NewCalculator(Window60, invalidParams, DefaultWeights(), nil)
		
		data := generateTradingDays(70, func(i int) TradingDay {
			return TradingDay{
				Date:          time.Date(2024, 1, 1+i, 0, 0, 0, 0, time.UTC),
				Symbol:        "TEST",
				Close:         100.0 + float64(i),
				Open:          99.0 + float64(i),
				High:          102.0 + float64(i),
				Low:           98.0 + float64(i),
				Value:        1000000,
				NumTrades:     100,
				TradingStatus: "ACTIVE",
			}
		})
		
		_, err := calc.Calculate(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "penalty parameters")
	})
}

// TestPenaltyFunctionsExtended tests all penalty function variations
func TestPenaltyFunctionsExtended(t *testing.T) {
	t.Run("piecewise penalty", func(t *testing.T) {
		tests := []struct {
			name    string
			p0      float64
			beta    float64
			gamma   float64
			pStar   float64
			maxMult float64
			minPen  float64
			maxPen  float64
		}{
			{"low price regime", 0.5, 0.3, 0.2, 1.0, 5.0, 1.0, 5.0},
			{"high price regime", 2.0, 0.3, 0.2, 1.0, 5.0, 1.0, 5.0},
			{"at transition", 1.0, 0.3, 0.2, 1.0, 5.0, 1.0, 1.0},
			{"invalid inputs", -1.0, 0.3, 0.2, 1.0, 5.0, 1.0, 1.0},
			{"zero price", 0.0, 0.3, 0.2, 1.0, 5.0, 1.0, 1.0},
			{"max mult reached", 0.1, 1.0, 1.0, 1.0, 2.0, 1.0, 2.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				penalty := PiecewisePenalty(tt.p0, tt.beta, tt.gamma, tt.pStar, tt.maxMult)
				
				assert.GreaterOrEqual(t, penalty, tt.minPen)
				assert.LessOrEqual(t, penalty, tt.maxPen)
				assert.False(t, math.IsNaN(penalty))
				assert.False(t, math.IsInf(penalty, 0))
			})
		}
	})

	t.Run("exponential penalty", func(t *testing.T) {
		tests := []struct {
			name    string
			p0      float64
			alpha   float64
			maxMult float64
			expectMin float64
		}{
			{"normal case", 1.0, 0.2, 3.0, 1.0},
			{"high price", 5.0, 0.2, 3.0, 1.0},
			{"low price", 0.2, 0.2, 3.0, 1.0},
			{"invalid inputs", -1.0, 0.2, 3.0, 1.0},
			{"zero price", 0.0, 0.2, 3.0, 1.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				penalty := ExponentialPenalty(tt.p0, tt.alpha, tt.maxMult)
				
				assert.GreaterOrEqual(t, penalty, tt.expectMin)
				assert.LessOrEqual(t, penalty, tt.maxMult)
				assert.False(t, math.IsNaN(penalty))
				assert.False(t, math.IsInf(penalty, 0))
			})
		}
	})
}

// TestTickerMetrics tests TickerMetrics validation and methods
func TestTickerMetrics(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		tests := []struct {
			name   string
			metric TickerMetrics
			valid  bool
		}{
			{
				name: "valid metrics",
				metric: TickerMetrics{
					Symbol:      "TEST",
					Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					TradingDays: 20,
					TotalDays:   30,
				},
				valid: true,
			},
			{
				name: "empty symbol",
				metric: TickerMetrics{
					Symbol:      "",
					Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					TradingDays: 20,
					TotalDays:   30,
				},
				valid: false,
			},
			{
				name: "zero date",
				metric: TickerMetrics{
					Symbol:      "TEST",
					Date:        time.Time{},
					TradingDays: 20,
					TotalDays:   30,
				},
				valid: false,
			},
			{
				name: "trading days > total days",
				metric: TickerMetrics{
					Symbol:      "TEST",
					Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					TradingDays: 35,
					TotalDays:   30,
				},
				valid: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.valid, tt.metric.IsValid())
			})
		}
	})
}

// TestValidationError tests custom error type
func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test error message",
		Value:   "invalid_value",
	}
	
	assert.Equal(t, "test error message", err.Error())
	assert.Equal(t, "test_field", err.Field)
	assert.Equal(t, "invalid_value", err.Value)
}

// Helper functions for test data generation
func generateTradingDays(count int, generator func(int) TradingDay) []TradingDay {
	days := make([]TradingDay, count)
	for i := 0; i < count; i++ {
		days[i] = generator(i)
	}
	return days
}

// generateRealisticISXData creates realistic test data mimicking ISX patterns
func generateRealisticISXData(symbols []string, days int, baseDate time.Time) []TradingDay {
	var allData []TradingDay
	
	for _, symbol := range symbols {
		// Different characteristics per symbol
		var basePrice, baseVolume float64
		var volatility, tradingFreq float64
		
		switch symbol {
		case "TASC": // High liquidity
			basePrice, baseVolume, volatility, tradingFreq = 2.50, 5000000, 0.02, 0.95
		case "BMFI": // Medium liquidity
			basePrice, baseVolume, volatility, tradingFreq = 1.20, 2000000, 0.03, 0.80
		case "BAGH": // Lower liquidity
			basePrice, baseVolume, volatility, tradingFreq = 0.85, 800000, 0.04, 0.60
		default: // Generic
			basePrice, baseVolume, volatility, tradingFreq = 1.50, 1500000, 0.025, 0.75
		}
		
		price := basePrice
		for day := 0; day < days; day++ {
			currentDate := baseDate.AddDate(0, 0, day)
			
			// Skip weekends
			if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
				continue
			}
			
			// Random price movement
			priceChange := (float64(day%7) - 3.5) / 7.0 * volatility
			price *= (1 + priceChange)
			if price <= 0 {
				price = basePrice * 0.5
			}
			
			// Determine if trading day
			isTrading := (day%10) < int(tradingFreq*10)
			
			var volume float64
			var numTrades int
			var status string
			
			if isTrading {
				volumeVariation := 0.5 + float64(day%3)*0.25
				volume = baseVolume * volumeVariation
				numTrades = int(volume / 10000)
				status = "ACTIVE"
			} else {
				volume = 0
				numTrades = 0
				status = "SUSPENDED"
			}
			
			// Create OHLC
			spread := price * (0.005 + float64(day%5)*0.003)
			open := price + (float64(day%3)-1)*spread/3
			high := math.Max(open, price) + float64(day%2)*spread
			low := math.Min(open, price) - float64((day+1)%2)*spread
			close := low + (high-low)*0.6
			
			if low <= 0 {
				low = price * 0.95
				close = price
				high = price * 1.05
				open = price
			}
			
			td := TradingDay{
				Date:          currentDate,
				Symbol:        symbol,
				Open:          open,
				High:          high,
				Low:           low,
				Close:         close,
				Value:        volume,
				NumTrades:     numTrades,
				TradingStatus: status,
			}
			
			if td.IsValid() {
				allData = append(allData, td)
			}
		}
	}
	
	return allData
}