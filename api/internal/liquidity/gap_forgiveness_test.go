package liquidity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGapForgiveness(t *testing.T) {
	tests := []struct {
		name           string
		data           []TradingDay
		config         GapPenaltyConfig
		expectedPenalty float64
		description    string
	}{
		{
			name: "single_5_day_gap_forgiven",
			data: generateTradingDataWithSpecificGaps([]int{5}), // One 5-day gap
			config: GapPenaltyConfig{
				ShortGapThreshold:    2,
				MediumGapThreshold:   7,
				ShortGapPenaltyRate:  0.05,
				MediumGapPenaltyRate: 0.10,
				LongGapPenaltyRate:   0.20,
				AllowedGapLength:     5,
				AllowedGapCount:      1,
				MaxPenalty:           10.0,
			},
			expectedPenalty: 1.0, // No penalty - gap is forgiven
			description:     "5-day gap should be forgiven when allowed",
		},
		{
			name: "two_5_day_gaps_one_forgiven",
			data: generateTradingDataWithSpecificGaps([]int{5, 5}), // Two 5-day gaps
			config: GapPenaltyConfig{
				ShortGapThreshold:    2,
				MediumGapThreshold:   7,
				ShortGapPenaltyRate:  0.05,
				MediumGapPenaltyRate: 0.10,
				LongGapPenaltyRate:   0.20,
				AllowedGapLength:     5,
				AllowedGapCount:      1, // Only one gap forgiven
				MaxPenalty:           10.0,
			},
			expectedPenalty: 1.4, // One gap forgiven, one penalized (2 days @ 0.05 + 3 days @ 0.10 = 0.4)
			description:     "Only first 5-day gap forgiven, second is penalized",
		},
		{
			name: "gap_too_long_not_forgiven",
			data: generateTradingDataWithSpecificGaps([]int{7}), // 7-day gap
			config: GapPenaltyConfig{
				ShortGapThreshold:    2,
				MediumGapThreshold:   7,
				ShortGapPenaltyRate:  0.05,
				MediumGapPenaltyRate: 0.10,
				LongGapPenaltyRate:   0.20,
				AllowedGapLength:     5, // Only gaps up to 5 days forgiven
				AllowedGapCount:      1,
				MaxPenalty:           10.0,
			},
			expectedPenalty: 1.6, // 7-day gap is penalized (2 @ 0.05 + 5 @ 0.10 = 0.6)
			description:     "7-day gap exceeds forgiveness threshold",
		},
		{
			name: "mixed_gaps_shortest_forgiven",
			data: generateTradingDataWithSpecificGaps([]int{3, 5, 2}), // Mix of gaps
			config: GapPenaltyConfig{
				ShortGapThreshold:    2,
				MediumGapThreshold:   7,
				ShortGapPenaltyRate:  0.05,
				MediumGapPenaltyRate: 0.10,
				LongGapPenaltyRate:   0.20,
				AllowedGapLength:     5,
				AllowedGapCount:      1, // Only one gap forgiven
				MaxPenalty:           10.0,
			},
			expectedPenalty: 1.54, // 2-day gap forgiven, 3-day (0.1 + 0.1) and 5-day (0.1 + 0.3) penalized
			description:     "Shortest eligible gap is forgiven first",
		},
		{
			name: "no_forgiveness_configured",
			data: generateTradingDataWithSpecificGaps([]int{5}),
			config: GapPenaltyConfig{
				ShortGapThreshold:    2,
				MediumGapThreshold:   7,
				ShortGapPenaltyRate:  0.05,
				MediumGapPenaltyRate: 0.10,
				LongGapPenaltyRate:   0.20,
				AllowedGapLength:     0, // No forgiveness
				AllowedGapCount:      0,
				MaxPenalty:           10.0,
			},
			expectedPenalty: 1.4, // 5-day gap is penalized (2 @ 0.05 + 3 @ 0.10 = 0.4)
			description:     "No gaps forgiven when forgiveness disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			penalty := CalculateGapPenalty(tt.data, tt.config)
			
			// Allow small tolerance for floating point comparison
			assert.InDelta(t, tt.expectedPenalty, penalty, 0.01, 
				"Gap penalty mismatch: %s", tt.description)
		})
	}
}

// generateTradingDataWithSpecificGaps creates test data with specific gap patterns
func generateTradingDataWithSpecificGaps(gapLengths []int) []TradingDay {
	var data []TradingDay
	baseDate := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	currentDate := baseDate
	basePrice := 10.0
	
	// For each gap specified
	for i := 0; i < len(gapLengths); i++ {
		// Add 10 trading days
		for j := 0; j < 10; j++ {
			data = append(data, TradingDay{
				Date:          currentDate,
				Symbol:        "TEST",
				Close:         basePrice + float64(j)*0.01,
				Volume:        1000000,
				Value:         10000000,
				NumTrades:     100,
				TradingStatus: "ACTIVE",
			})
			currentDate = currentDate.AddDate(0, 0, 1)
		}
		
		// Add gap (non-trading days)
		for j := 0; j < gapLengths[i]; j++ {
			data = append(data, TradingDay{
				Date:          currentDate,
				Symbol:        "TEST",
				Close:         basePrice,
				Volume:        0,
				Value:         0,
				NumTrades:     0,
				TradingStatus: "false",
			})
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}
	
	// Add final trading days
	for j := 0; j < 10; j++ {
		data = append(data, TradingDay{
			Date:          currentDate,
			Symbol:        "TEST",
			Close:         basePrice + float64(j)*0.01,
			Volume:        1000000,
			Value:         10000000,
			NumTrades:     100,
			TradingStatus: "ACTIVE",
		})
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	
	return data
}

func TestApplyGapForgiveness(t *testing.T) {
	tests := []struct {
		name             string
		gaps             []GapInfo
		config           GapPenaltyConfig
		expectedRemaining int
		description      string
	}{
		{
			name: "forgive_single_eligible_gap",
			gaps: []GapInfo{
				{Length: 5, StartIndex: 10, EndIndex: 14},
			},
			config: GapPenaltyConfig{
				AllowedGapLength: 5,
				AllowedGapCount:  1,
			},
			expectedRemaining: 0,
			description:      "Single 5-day gap should be forgiven",
		},
		{
			name: "forgive_one_of_two_gaps",
			gaps: []GapInfo{
				{Length: 5, StartIndex: 10, EndIndex: 14},
				{Length: 4, StartIndex: 25, EndIndex: 28},
			},
			config: GapPenaltyConfig{
				AllowedGapLength: 5,
				AllowedGapCount:  1,
			},
			expectedRemaining: 1,
			description:      "One of two eligible gaps should be forgiven",
		},
		{
			name: "no_forgiveness_for_long_gaps",
			gaps: []GapInfo{
				{Length: 7, StartIndex: 10, EndIndex: 16},
				{Length: 10, StartIndex: 25, EndIndex: 34},
			},
			config: GapPenaltyConfig{
				AllowedGapLength: 5,
				AllowedGapCount:  1,
			},
			expectedRemaining: 2,
			description:      "Gaps exceeding allowed length are not forgiven",
		},
		{
			name: "forgive_multiple_gaps",
			gaps: []GapInfo{
				{Length: 3, StartIndex: 10, EndIndex: 12},
				{Length: 4, StartIndex: 20, EndIndex: 23},
				{Length: 5, StartIndex: 30, EndIndex: 34},
			},
			config: GapPenaltyConfig{
				AllowedGapLength: 5,
				AllowedGapCount:  2,
			},
			expectedRemaining: 1,
			description:      "Two gaps should be forgiven",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyGapForgiveness(tt.gaps, tt.config)
			assert.Equal(t, tt.expectedRemaining, len(result), tt.description)
		})
	}
}

func TestGapForgivenessIntegration(t *testing.T) {
	// Test with real-world scenario: 60-day window with one 5-day meeting gap
	baseDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	var data []TradingDay
	
	// Create 60 days of data with one 5-day gap in the middle
	for i := 0; i < 60; i++ {
		currentDate := baseDate.AddDate(0, 0, i)
		
		// Insert 5-day gap from day 25-29 (simulating a meeting period)
		if i >= 25 && i <= 29 {
			data = append(data, TradingDay{
				Date:          currentDate,
				Symbol:        "TEST",
				Open:          100.0,
				High:          100.0,
				Low:           100.0,
				Close:         100.0,
				Volume:        0,
				ShareVolume:   0,
				Value:         0,
				NumTrades:     0,
				TradingStatus: "false",
			})
		} else {
			// Normal trading day with price variation
			price := 100.0 + float64(i)*0.5
			data = append(data, TradingDay{
				Date:          currentDate,
				Symbol:        "TEST",
				Open:          price - 0.5,
				High:          price + 1.0,
				Low:           price - 1.0,
				Close:         price,
				Volume:        1000000,
				ShareVolume:   1000000,
				Value:         100000000,
				NumTrades:     100,
				TradingStatus: "ACTIVE",
			})
		}
	}
	
	// Calculate ILLIQ with gap penalty (with forgiveness)
	config := DefaultGapPenaltyConfig()
	illiqWithForgiveness, _, _ := ComputeILLIQWithGapPenalty(data, 0.05, 0.95, true, &config)
	
	// Calculate ILLIQ without gap penalty
	illiqWithoutPenalty, _, _ := ComputeILLIQWithGapPenalty(data, 0.05, 0.95, false, nil)
	
	// With forgiveness, the 5-day gap should not affect ILLIQ
	assert.InDelta(t, illiqWithoutPenalty, illiqWithForgiveness, 0.01,
		"ILLIQ should be same with forgiven gap")
	
	// Now test with a gap that won't be forgiven (8 days)
	data = nil
	for i := 0; i < 60; i++ {
		currentDate := baseDate.AddDate(0, 0, i)
		
		// Insert 8-day gap from day 25-32 (too long to forgive)
		if i >= 25 && i <= 32 {
			data = append(data, TradingDay{
				Date:          currentDate,
				Symbol:        "TEST",
				Open:          100.0,
				High:          100.0,
				Low:           100.0,
				Close:         100.0,
				Volume:        0,
				ShareVolume:   0,
				Value:         0,
				NumTrades:     0,
				TradingStatus: "false",
			})
		} else {
			// Normal trading day with price variation
			price := 100.0 + float64(i)*0.5
			data = append(data, TradingDay{
				Date:          currentDate,
				Symbol:        "TEST",
				Open:          price - 0.5,
				High:          price + 1.0,
				Low:           price - 1.0,
				Close:         price,
				Volume:        1000000,
				ShareVolume:   1000000,
				Value:         100000000,
				NumTrades:     100,
				TradingStatus: "ACTIVE",
			})
		}
	}
	
	illiqWithLongGap, _, _ := ComputeILLIQWithGapPenalty(data, 0.05, 0.95, true, &config)
	
	// The 8-day gap should be penalized
	assert.Greater(t, illiqWithLongGap, illiqWithoutPenalty,
		"ILLIQ should be worse with unforgivable 8-day gap")
}