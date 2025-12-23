package liquidity

import (
	"fmt"
	"testing"
	"time"
)

func TestGapForgivenessDebug(t *testing.T) {
	// Test case: Two 5-day gaps, one should be forgiven
	data := generateTradingDataWithSpecificGaps([]int{5, 5})
	
	config := GapPenaltyConfig{
		ShortGapThreshold:    2,
		MediumGapThreshold:   7,
		ShortGapPenaltyRate:  0.05,
		MediumGapPenaltyRate: 0.10,
		LongGapPenaltyRate:   0.20,
		AllowedGapLength:     5,
		AllowedGapCount:      1, // Only one gap forgiven
		MaxPenalty:           10.0,
	}
	
	// Identify gaps
	gaps := identifyDetailedGaps(data)
	fmt.Printf("Identified gaps before forgiveness:\n")
	for i, gap := range gaps {
		fmt.Printf("  Gap %d: Length=%d, StartIndex=%d, EndIndex=%d\n", 
			i+1, gap.Length, gap.StartIndex, gap.EndIndex)
	}
	
	// Apply forgiveness
	remainingGaps := applyGapForgiveness(gaps, config)
	fmt.Printf("\nGaps after forgiveness:\n")
	for i, gap := range remainingGaps {
		fmt.Printf("  Gap %d: Length=%d, StartIndex=%d, EndIndex=%d\n", 
			i+1, gap.Length, gap.StartIndex, gap.EndIndex)
	}
	
	// Calculate penalty
	penalty := CalculateGapPenalty(data, config)
	fmt.Printf("\nCalculated penalty: %.4f\n", penalty)
	
	// Calculate expected penalty for one 5-day gap
	// 5-day gap: 2 days at 0.05 + 3 days at 0.10 = 0.10 + 0.30 = 0.40
	expectedPenalty := 1.0 + 0.10 + 0.30
	fmt.Printf("Expected penalty: %.4f\n", expectedPenalty)
}

func TestSingleGapCalculation(t *testing.T) {
	// Create data with a single 7-day gap
	baseDate := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	var data []TradingDay
	
	// Add 10 trading days
	for i := 0; i < 10; i++ {
		data = append(data, TradingDay{
			Date:          baseDate.AddDate(0, 0, i),
			Symbol:        "TEST",
			Close:         100.0 + float64(i)*0.1,
			Volume:        1000000,
			Value:         10000000,
			NumTrades:     100,
			TradingStatus: "ACTIVE",
		})
	}
	
	// Add 7-day gap
	for i := 10; i < 17; i++ {
		data = append(data, TradingDay{
			Date:          baseDate.AddDate(0, 0, i),
			Symbol:        "TEST",
			Close:         100.0,
			Volume:        0,
			Value:         0,
			NumTrades:     0,
			TradingStatus: "false",
		})
	}
	
	// Add more trading days
	for i := 17; i < 27; i++ {
		data = append(data, TradingDay{
			Date:          baseDate.AddDate(0, 0, i),
			Symbol:        "TEST",
			Close:         100.0 + float64(i)*0.1,
			Volume:        1000000,
			Value:         10000000,
			NumTrades:     100,
			TradingStatus: "ACTIVE",
		})
	}
	
	config := GapPenaltyConfig{
		ShortGapThreshold:    2,
		MediumGapThreshold:   7,
		ShortGapPenaltyRate:  0.05,
		MediumGapPenaltyRate: 0.10,
		LongGapPenaltyRate:   0.20,
		AllowedGapLength:     5, // 7-day gap won't be forgiven
		AllowedGapCount:      1,
		MaxPenalty:           10.0,
	}
	
	gaps := identifyDetailedGaps(data)
	fmt.Printf("\n7-day gap test:\n")
	fmt.Printf("Identified gaps: %d\n", len(gaps))
	for _, gap := range gaps {
		fmt.Printf("  Gap: Length=%d\n", gap.Length)
	}
	
	penalty := CalculateGapPenalty(data, config)
	// 7-day gap: 2 days at 0.05 + 5 days at 0.10 = 0.10 + 0.50 = 0.60
	expectedPenalty := 1.0 + 0.10 + 0.50
	fmt.Printf("Calculated penalty: %.4f\n", penalty)
	fmt.Printf("Expected penalty: %.4f\n", expectedPenalty)
}