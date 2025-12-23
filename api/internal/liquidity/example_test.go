package liquidity

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Example_basicUsage demonstrates the complete ISX Hybrid Liquidity Metric calculation
func Example_basicUsage() {
	ctx := context.Background()
	
	// Create sample ISX trading data for demonstration
	sampleData := generateSampleISXData()
	
	// Create calculator with default parameters
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	calculator := NewCalculator(
		Window60,                // 60-day rolling window
		DefaultPenaltyParams(),  // Default penalty parameters for ISX
		DefaultWeights(),        // Default component weights
		logger,
	)
	
	// Calculate liquidity metrics
	metrics, err := calculator.Calculate(ctx, sampleData)
	if err != nil {
		fmt.Printf("Error calculating metrics: %v\n", err)
		return
	}
	
	// Display results for the first few metrics
	fmt.Printf("ISX Hybrid Liquidity Metrics Results:\n")
	fmt.Printf("====================================\n")
	for i, metric := range metrics {
		if i >= 3 { // Show only first 3 for brevity
			break
		}
		
		fmt.Printf("Symbol: %s | Date: %s | Hybrid Score: %.2f | Rank: %d\n",
			metric.Symbol, 
			metric.Date.Format("2006-01-02"), 
			metric.HybridScore, 
			metric.HybridRank,
		)
		fmt.Printf("  ILLIQ: %.6f | Value: %.0f | Continuity: %.3f\n",
			metric.ILLIQ,
			metric.Value,
			metric.Continuity,
		)
		fmt.Printf("  Scaled - ILLIQ: %.1f | Value: %.1f | Continuity: %.1f\n\n",
			metric.ILLIQScaled,
			metric.ValueScaled,
			metric.ContinuityScaled,
		)
	}
	
	fmt.Printf("Total metrics calculated: %d\n", len(metrics))
}

// generateSampleISXData creates sample trading data representative of ISX characteristics
func generateSampleISXData() []TradingDay {
	// Simulate typical ISX trading patterns
	symbols := []string{"TASC", "BMFI", "BAGH", "ISFF", "IMAP"}
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	var allData []TradingDay
	
	for _, symbol := range symbols {
		// Generate 90 days of data for each symbol (enough for 60-day window)
		for day := 0; day < 90; day++ {
			currentDate := baseDate.AddDate(0, 0, day)
			
			// Skip weekends (simplified)
			if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
				continue
			}
			
			// Simulate different liquidity characteristics per symbol
			var basePrice, baseVolume float64
			var tradingFrequency float64
			
			switch symbol {
			case "TASC": // High liquidity stock
				basePrice, baseVolume, tradingFrequency = 2.50, 5000000, 0.9
			case "BMFI": // Medium liquidity stock
				basePrice, baseVolume, tradingFrequency = 1.20, 2000000, 0.7
			case "BAGH": // Lower liquidity stock
				basePrice, baseVolume, tradingFrequency = 0.85, 800000, 0.5
			case "ISFF": // Intermittent trading
				basePrice, baseVolume, tradingFrequency = 1.80, 1200000, 0.4
			case "IMAP": // Very low liquidity
				basePrice, baseVolume, tradingFrequency = 0.65, 300000, 0.3
			}
			
			// Add some random variation
			dayFactor := 1.0 + 0.1*float64(day%10-5)/5 // ±10% variation
			priceVariation := basePrice * dayFactor
			volumeVariation := baseVolume * (0.5 + 0.8*float64(day%7)/6) // Volume varies by day
			
			// Determine if trading occurs (based on trading frequency)
			isTrading := float64(day%10) < tradingFrequency*10
			
			var volume float64
			var numTrades int
			var status string
			
			if isTrading {
				volume = volumeVariation
				numTrades = int(volume / 10000) // Approximate trades based on volume
				status = "ACTIVE"
			} else {
				volume = 0
				numTrades = 0
				status = "SUSPENDED"
			}
			
			// Create realistic OHLC data
			open := priceVariation
			high := open * (1.0 + 0.05*float64(day%5)/4) // Up to 5% daily high
			low := open * (1.0 - 0.04*float64(day%3)/2)  // Up to 4% daily low
			close := low + (high-low)*0.6                 // Close somewhere in range
			
			tradingDay := TradingDay{
				Date:          currentDate,
				Symbol:        symbol,
				Open:          open,
				High:          high,
				Low:           low,
				Close:         close,
				Volume:        volume,
				NumTrades:     numTrades,
				TradingStatus: status,
			}
			
			// Only add if it passes basic validation
			if tradingDay.IsValid() {
				allData = append(allData, tradingDay)
			}
		}
	}
	
	return allData
}

// Example_parameterCalibration demonstrates parameter calibration workflow
func Example_parameterCalibration() {
	fmt.Printf("Parameter Calibration Example:\n")
	fmt.Printf("==============================\n")
	
	// This would typically use real historical data
	fmt.Printf("1. Load historical ISX data from CSV files\n")
	fmt.Printf("2. Configure calibration parameters\n")
	fmt.Printf("3. Run grid search optimization\n")
	fmt.Printf("4. Validate results with cross-validation\n")
	fmt.Printf("5. Apply calibrated parameters to new data\n\n")
	
	// Show default configuration
	config := DefaultCalibrationConfig()
	fmt.Printf("Default Calibration Configuration:\n")
	fmt.Printf("- Grid Size: %d\n", config.ParamGridSize)
	fmt.Printf("- K-Folds: %d\n", config.KFolds)
	fmt.Printf("- Target Metric: %s\n", config.TargetMetric)
	fmt.Printf("- Min Tickers: %d\n", config.MinTickers)
	
	// Show parameter ranges that would be tested
	fmt.Printf("\nParameter Ranges Tested:\n")
	fmt.Printf("- Piecewise Beta: 0.1 - 0.8\n")
	fmt.Printf("- Piecewise Gamma: 0.05 - 0.4\n")
	fmt.Printf("- Piecewise P*: 1.0 - 5.0 IQD\n")
	fmt.Printf("- Exponential Alpha: 0.1 - 0.5\n")
	fmt.Printf("\nNote: Use real calibration with: result, err := Calibrate(ctx, data, config)\n")
}

// Example_dataLoading demonstrates data loading and validation
func Example_dataLoading() {
	fmt.Printf("Data Loading and Validation Example:\n")
	fmt.Printf("====================================\n")
	
	// Show expected CSV format
	fmt.Printf("Expected CSV Format:\n")
	fmt.Printf("Date,Symbol,Open,High,Low,Close,Volume,NumTrades,Status\n")
	fmt.Printf("2024-01-01,TASC,2.50,2.55,2.45,2.52,5000000,150,ACTIVE\n")
	fmt.Printf("2024-01-02,TASC,2.52,2.58,2.48,2.56,4800000,145,ACTIVE\n")
	fmt.Printf("...\n\n")
	
	// Show validation checks performed
	fmt.Printf("Validation Checks Performed:\n")
	fmt.Printf("- Price consistency (High >= Open, Close, Low)\n")
	fmt.Printf("- Positive prices and volumes\n")
	fmt.Printf("- Date formatting and ordering\n")
	fmt.Printf("- Minimum data requirements\n")
	fmt.Printf("- Trading status consistency\n")
	fmt.Printf("- Data quality thresholds\n\n")
	
	// Show data quality metrics
	fmt.Printf("Data Quality Assessment:\n")
	fmt.Printf("- Valid ratio: percentage of valid records\n")
	fmt.Printf("- Trading ratio: percentage of active trading days\n")
	fmt.Printf("- Continuity: consistency of trading activity\n")
	fmt.Printf("- Coverage: date range and completeness\n")
}

// Example_outputFormats demonstrates different output options
func Example_outputFormats() {
	fmt.Printf("Output Formats Example:\n")
	fmt.Printf("=======================\n")
	
	fmt.Printf("1. CSV Output (for analysis):\n")
	fmt.Printf("   - Comprehensive metrics with all components\n")
	fmt.Printf("   - Data quality indicators\n")
	fmt.Printf("   - Usage: SaveToCSV(metrics, \"output.csv\")\n\n")
	
	fmt.Printf("2. JSON Output (for systems integration):\n")
	fmt.Printf("   - Structured data with metadata\n")
	fmt.Printf("   - Machine-readable format\n")
	fmt.Printf("   - Usage: SaveToJSON(metrics, \"output.json\")\n\n")
	
	fmt.Printf("3. Summary Report (for stakeholders):\n")
	fmt.Printf("   - Statistical summaries\n")
	fmt.Printf("   - Top/bottom performers\n")
	fmt.Printf("   - Data quality distribution\n")
	fmt.Printf("   - Usage: SaveSummaryReport(metrics, \"summary.txt\")\n\n")
	
	fmt.Printf("Key Output Metrics:\n")
	fmt.Printf("- Hybrid Score: 0-100 liquidity ranking\n")
	fmt.Printf("- Component Scores: ILLIQ, Volume, Continuity\n")
	fmt.Printf("- Penalty Adjustments: Price-level corrections\n")
	fmt.Printf("- Rankings: Relative position vs peers\n")
	fmt.Printf("- Quality Indicators: Data reliability measures\n")
}