package liquidity

import (
	"context"
	"log/slog"
	"math"
	"os"
	"testing"
	"time"
)

// Benchmark tests for critical liquidity calculation functions
// These tests measure performance to ensure the system can handle ISX data volumes

// BenchmarkILLIQCalculation benchmarks the ILLIQ calculation performance
func BenchmarkILLIQCalculation(b *testing.B) {
	// Create realistic data sizes for different scenarios
	benchmarks := []struct {
		name string
		size int
	}{
		{"small_window_20_days", 20},
		{"medium_window_60_days", 60},
		{"large_window_120_days", 120},
		{"full_year_250_days", 250},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Generate test data once
			data := generateBenchmarkTradingData(bm.size, "TASC")
			
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ComputeILLIQ(data, 0.05, 0.95)
			}
		})
	}
}

// BenchmarkPenaltyFunctions benchmarks penalty function calculations
func BenchmarkPenaltyFunctions(b *testing.B) {
	params := DefaultPenaltyParams()
	prices := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 5.0, 10.0}

	b.Run("piecewise_penalty", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, price := range prices {
				PiecewisePenalty(price, params.PiecewiseBeta, params.PiecewiseGamma,
					params.PiecewisePStar, params.PiecewiseMaxMult)
			}
		}
	})

	b.Run("exponential_penalty", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, price := range prices {
				ExponentialPenalty(price, params.ExponentialAlpha, params.ExponentialMaxMult)
			}
		}
	})
}

// BenchmarkContinuityCalculation benchmarks continuity-related calculations
func BenchmarkContinuityCalculation(b *testing.B) {
	// Test different data sizes and continuity patterns
	benchmarks := []struct {
		name           string
		size           int
		continuityRate float64
	}{
		{"high_continuity_60_days", 60, 0.9},
		{"medium_continuity_60_days", 60, 0.7},
		{"low_continuity_60_days", 60, 0.5},
		{"high_continuity_250_days", 250, 0.9},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := generateBenchmarkTradingData(bm.size, "TEST")
			// Modify data to match continuity rate
			nonTradingDays := int(float64(bm.size) * (1.0 - bm.continuityRate))
			for i := 0; i < nonTradingDays && i < len(data); i++ {
				data[i*3%len(data)].Volume = 0
				data[i*3%len(data)].NumTrades = 0
				data[i*3%len(data)].TradingStatus = "SUSPENDED"
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				continuity := CalculateContinuity(data)
				ContinuityNL(continuity, DefaultContinuityDelta)
			}
		})
	}
}

// BenchmarkCorwinSchultzCalculation benchmarks spread calculation
func BenchmarkCorwinSchultzCalculation(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"single_calculation", 1},
		{"small_series", 20},
		{"medium_series", 60},
		{"large_series", 250},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := generateBenchmarkTradingData(bm.size+1, "TEST") // +1 for spread calculation
			
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				if bm.name == "single_calculation" {
					CorwinSchultz(data[0].High, data[0].Low, data[1].High, data[1].Low)
				} else {
					CalculateSpreadSeries(data)
				}
			}
		})
	}
}

// BenchmarkRobustScaling benchmarks cross-sectional scaling operations
func BenchmarkRobustScaling(b *testing.B) {
	benchmarks := []struct {
		name         string
		size         int
		invert       bool
		logTransform bool
	}{
		{"small_basic", 10, false, false},
		{"small_log", 10, false, true},
		{"small_invert_log", 10, true, true},
		{"medium_basic", 50, false, false},
		{"medium_log", 50, false, true},
		{"medium_invert_log", 50, true, true},
		{"large_basic", 200, false, false},
		{"large_log", 200, false, true},
		{"large_invert_log", 200, true, true},
		{"isx_daily_scale", 45, true, true}, // Typical daily ISX stock count
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Generate values with realistic ISX ILLIQ distribution (log-normal)
			values := make([]float64, bm.size)
			for i := 0; i < bm.size; i++ {
				// Simulate log-normal distribution typical of ILLIQ values
				logValue := float64(i%10) + float64(i)*0.01
				values[i] = math.Exp(logValue) * 1e-8 // Scale to realistic ILLIQ range
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				RobustScale(values, bm.invert, bm.logTransform)
			}
		})
	}
}

// BenchmarkCalculatorIntegration benchmarks the complete calculator workflow
func BenchmarkCalculatorIntegration(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	benchmarks := []struct {
		name       string
		window     Window
		numSymbols int
		numDays    int
	}{
		{"single_stock_60_days", Window60, 1, 90},
		{"small_portfolio_60_days", Window60, 5, 90},
		{"medium_portfolio_60_days", Window60, 20, 90},
		{"large_portfolio_60_days", Window60, 45, 90}, // Full ISX active stocks
		{"single_stock_120_days", Window120, 1, 150},
		{"medium_portfolio_120_days", Window120, 20, 150},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Create calculator
			calc := NewCalculator(bm.window, DefaultPenaltyParams(), DefaultWeights(), logger)
			
			// Generate test data
			symbols := make([]string, bm.numSymbols)
			for i := 0; i < bm.numSymbols; i++ {
				symbols[i] = generateISXSymbol(i)
			}
			
			baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			data := generateMultiSymbolBenchmarkData(symbols, bm.numDays, baseDate)
			
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := calc.Calculate(ctx, data)
				if err != nil {
					b.Fatalf("Calculator error: %v", err)
				}
			}
		})
	}
}

// BenchmarkDataValidation benchmarks input data validation
func BenchmarkDataValidation(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"small_dataset", 100},
		{"medium_dataset", 1000},
		{"large_dataset", 10000},
		{"full_isx_year", 11250}, // 45 stocks * 250 trading days
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := generateBenchmarkTradingData(bm.size, "TEST")
			
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ValidateTradingData(data)
			}
		})
	}
}

// BenchmarkMemoryUsage tests memory efficiency for large datasets
func BenchmarkMemoryUsage(b *testing.B) {
	// Test with realistic ISX data volume: 45 stocks * 250 trading days
	symbols := make([]string, 45)
	for i := 0; i < 45; i++ {
		symbols[i] = generateISXSymbol(i)
	}
	
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := generateMultiSymbolBenchmarkData(symbols, 250, baseDate)

	b.Run("full_isx_calculation", func(b *testing.B) {
		ctx := context.Background()
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
		calc := NewCalculator(Window60, DefaultPenaltyParams(), DefaultWeights(), logger)
		
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			metrics, err := calc.Calculate(ctx, data)
			if err != nil {
				b.Fatalf("Calculator error: %v", err)
			}
			_ = metrics // Prevent optimization
		}
	})
}

// Benchmark helper functions

// generateBenchmarkTradingData creates trading data for benchmarking
func generateBenchmarkTradingData(size int, symbol string) []TradingDay {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	basePrice := 2.0
	baseVolume := 1000000.0
	
	data := make([]TradingDay, 0, size)
	
	for i := 0; i < size; i++ {
		currentDate := baseDate.AddDate(0, 0, i)
		
		// Skip weekends
		if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
			continue
		}
		
		// Simple price and volume variation
		priceVariation := basePrice * (1.0 + float64(i%20-10)*0.01)
		volumeVariation := baseVolume * (0.5 + float64(i%10)*0.1)
		
		// Create OHLC
		open := priceVariation
		high := open * (1.0 + float64(i%5)*0.005)
		low := open * (1.0 - float64(i%3)*0.004)
		close := low + (high-low)*0.6
		
		td := TradingDay{
			Date:          currentDate,
			Symbol:        symbol,
			Open:          open,
			High:          high,
			Low:           low,
			Close:         close,
			Volume:        volumeVariation,
			NumTrades:     int(volumeVariation / 10000),
			TradingStatus: "ACTIVE",
		}
		
		if td.IsValid() && len(data) < size {
			data = append(data, td)
		}
	}
	
	return data
}

// generateMultiSymbolBenchmarkData creates data for multiple symbols
func generateMultiSymbolBenchmarkData(symbols []string, days int, baseDate time.Time) []TradingDay {
	var allData []TradingDay
	
	for _, symbol := range symbols {
		// Different characteristics per symbol for realistic benchmarking
		var basePrice, baseVolume float64
		var volatility float64
		
		// Simple symbol-based variation
		symbolHash := int(symbol[0]) % 10
		basePrice = 1.0 + float64(symbolHash)*0.3
		baseVolume = 500000 + float64(symbolHash)*500000
		volatility = 0.01 + float64(symbolHash)*0.002
		
		for day := 0; day < days; day++ {
			currentDate := baseDate.AddDate(0, 0, day)
			
			// Skip weekends
			if currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
				continue
			}
			
			// Price movement
			priceChange := math.Sin(float64(day)*2*math.Pi/20) * volatility
			currentPrice := basePrice * (1.0 + priceChange)
			
			// Volume variation
			volumeChange := math.Cos(float64(day)*2*math.Pi/15) * 0.3
			currentVolume := baseVolume * (1.0 + volumeChange)
			
			// Create OHLC
			spread := currentPrice * 0.01
			open := currentPrice
			high := currentPrice + spread/2
			low := currentPrice - spread/2
			close := currentPrice + priceChange*spread/4
			
			td := TradingDay{
				Date:          currentDate,
				Symbol:        symbol,
				Open:          open,
				High:          high,
				Low:           low,
				Close:         close,
				Volume:        currentVolume,
				NumTrades:     int(currentVolume / 10000),
				TradingStatus: "ACTIVE",
			}
			
			if td.IsValid() {
				allData = append(allData, td)
			}
		}
	}
	
	return allData
}

// generateISXSymbol generates realistic ISX stock symbols for benchmarking
func generateISXSymbol(index int) string {
	// Common ISX stock prefixes and patterns
	prefixes := []string{
		"TASC", "BMFI", "BAGH", "ISFF", "IMAP", "IHII", "IDIB", "IMAR", "IHIB", "ISHI",
		"IMFA", "IBSD", "ILNG", "IEOB", "IHFI", "ICBK", "IKOM", "IZFM", "IECM", "ITFI",
		"ISRI", "IVMS", "IHTL", "IMSC", "IBKR", "IPCO", "ILDC", "ISPC", "IMPR", "IRCM",
		"IFER", "ISEM", "IWTR", "IAGR", "IMNF", "ITED", "IAIR", "IPHR", "ICMT", "IINS",
		"IRET", "ILOG", "IFIN", "IMTL", "IOIL",
	}
	
	if index < len(prefixes) {
		return prefixes[index]
	}
	
	// Generate additional symbols if needed
	return prefixes[index%len(prefixes)]
}

// BenchmarkCalibration benchmarks the parameter calibration process
func BenchmarkCalibration(b *testing.B) {
	// Note: This would benchmark the actual calibration process when implemented
	// For now, it benchmarks the setup and validation of calibration parameters
	
	b.Run("calibration_setup", func(b *testing.B) {
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			config := DefaultCalibrationConfig()
			params := DefaultPenaltyParams()
			weights := DefaultWeights()
			
			// Validate configuration
			_ = config.IsValid()
			_ = params.IsValid()
			_ = weights.IsValid()
			
			// Test different market condition weights
			_ = CalibratedWeights("high_volatility")
			_ = CalibratedWeights("low_volatility")
			_ = CalibratedWeights("normal")
		}
	})
}

// BenchmarkConcurrentCalculation benchmarks concurrent processing
func BenchmarkConcurrentCalculation(b *testing.B) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	
	// Create calculators with different concurrency settings
	benchmarks := []struct {
		name           string
		maxConcurrency int
	}{
		{"sequential", 1},
		{"concurrent_2", 2},
		{"concurrent_4", 4},
		{"concurrent_8", 8},
	}
	
	// Generate substantial dataset
	symbols := make([]string, 20)
	for i := 0; i < 20; i++ {
		symbols[i] = generateISXSymbol(i)
	}
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := generateMultiSymbolBenchmarkData(symbols, 90, baseDate)
	
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			calc := NewCalculator(Window60, DefaultPenaltyParams(), DefaultWeights(), logger)
			calc.SetConfiguration(false, bm.maxConcurrency, 60*time.Second)
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				_, err := calc.Calculate(ctx, data)
				if err != nil {
					b.Fatalf("Calculator error: %v", err)
				}
			}
		})
	}
}