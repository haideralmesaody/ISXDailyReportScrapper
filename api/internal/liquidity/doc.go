// Package liquidity implements the ISX Hybrid Liquidity Metric for the Iraqi Stock Exchange.
//
// This package provides a comprehensive implementation of a hybrid liquidity measurement system
// that combines multiple liquidity dimensions into a single, robust metric suitable for
// emerging markets like the Iraqi Stock Exchange (ISX).
//
// # Core Components
//
// The ISX Hybrid Liquidity Metric incorporates three main components:
//
//  1. Price Impact (ILLIQ): Based on the Amihud (2002) illiquidity measure with log-winsorization
//  2. Volume: Trading volume with exponential penalty adjustments for price levels
//  3. Trading Continuity: Frequency of trading with non-linear transformations
//
// # Architecture
//
// The package follows clean architecture principles with clear separation of concerns:
//
//   - types.go: Core data structures and interfaces
//   - calculator.go: Main orchestrator for metric calculation
//   - penalties.go: Penalty functions for price-level adjustments
//   - impact.go: ILLIQ (Amihud illiquidity) calculations with winsorization
//   - continuity.go: Trading continuity calculations and transformations
//   - scaling.go: Cross-sectional scaling using robust statistics
//   - corwin_schultz.go: Bid-ask spread proxy estimation
//   - window.go: Data loading and time window management
//   - weights.go: Component weight estimation and optimization
//   - calibration.go: Parameter calibration using grid search
//   - persist.go: Output formatting and persistence
//   - validate.go: Comprehensive input and output validation
//
// # Usage Example
//
//	// Load trading data
//	data, err := AssembleWindow(ctx, "data/csv", Window60, []string{"TASC", "BMFI"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Create calculator with default parameters
//	calculator := NewCalculator(
//	    Window60,
//	    DefaultPenaltyParams(),
//	    DefaultWeights(),
//	    slog.Default(),
//	)
//	
//	// Calculate metrics
//	var allMetrics []TickerMetrics
//	for symbol, tickerData := range data {
//	    metrics, err := calculator.Calculate(ctx, tickerData)
//	    if err != nil {
//	        log.Printf("Error calculating metrics for %s: %v", symbol, err)
//	        continue
//	    }
//	    allMetrics = append(allMetrics, metrics...)
//	}
//	
//	// Save results
//	err = SaveToCSV(allMetrics, "output/liquidity_metrics.csv")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Parameter Calibration
//
// The package supports automated parameter calibration using cross-validation:
//
//	// Set up calibration configuration
//	config := DefaultCalibrationConfig()
//	config.TargetMetric = "combined"
//	config.KFolds = 5
//	
//	// Perform calibration
//	result, err := Calibrate(ctx, data, config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Use calibrated parameters
//	calculator := NewCalculator(
//	    Window60,
//	    result.OptimalParams,
//	    result.OptimalWeights,
//	    slog.Default(),
//	)
//
// # Key Features
//
//   - Robust to outliers through winsorization and robust scaling
//   - Handles missing data and non-trading periods gracefully
//   - Cross-sectional scaling ensures comparability across tickers
//   - Penalty functions adjust for price-level effects common in emerging markets
//   - Comprehensive validation ensures data quality and result reliability
//   - Structured logging with context propagation
//   - Concurrent processing for improved performance
//   - Extensive testing and documentation
//
// # Data Requirements
//
// Input data should be provided as CSV files with the following columns:
//   - Date: Trading date (YYYY-MM-DD format)
//   - Symbol: Ticker symbol
//   - Open: Opening price
//   - High: Highest price
//   - Low: Lowest price  
//   - Close: Closing price
//   - Volume: Trading volume
//   - NumTrades: Number of trades (optional)
//   - Status: Trading status ("ACTIVE", "SUSPENDED", etc.)
//
// # Output Format
//
// The package generates comprehensive output including:
//   - Raw component scores (ILLIQ, volume, continuity)
//   - Scaled component scores (cross-sectionally normalized)
//   - Penalty adjustments for price levels
//   - Final hybrid liquidity scores (0-100 scale)
//   - Relative rankings within each time period
//   - Data quality indicators
//   - Supporting statistics (returns, volatility, trading frequency)
//
// # Mathematical Foundation
//
// The ISX Hybrid Liquidity Metric is calculated as:
//
//   Hybrid Score = w₁ × ILLIQ_scaled × P₁ + w₂ × Volume_scaled × P₂ + w₃ × Continuity_scaled
//
// Where:
//   - w₁, w₂, w₃ are component weights (sum to 1)
//   - ILLIQ_scaled is cross-sectionally scaled Amihud illiquidity (inverted)
//   - Volume_scaled is cross-sectionally scaled average volume
//   - Continuity_scaled is cross-sectionally scaled trading continuity
//   - P₁, P₂ are penalty multipliers based on price levels
//
// # Performance Considerations
//
// The package is optimized for production use:
//   - Memory-efficient processing of large datasets
//   - Concurrent calculation across tickers when possible
//   - Robust error handling with graceful degradation
//   - Configurable timeouts and resource limits
//   - Comprehensive logging for monitoring and debugging
//
// # Validation and Testing
//
// All functions include comprehensive validation:
//   - Input data validation with detailed error messages
//   - Parameter range checking with sensible defaults
//   - Output validation to ensure metric reasonableness
//   - Statistical tests for distribution properties
//   - Integration tests with real ISX data
//
// # Extensions and Customization
//
// The package is designed for extensibility:
//   - Custom penalty functions can be implemented
//   - Alternative scaling methods are supported
//   - Component weights can be dynamically adjusted
//   - New data sources can be easily integrated
//   - Output formats can be extended
//
// # References
//
// This implementation is based on:
//   - Amihud, Y. (2002). Illiquidity and stock returns
//   - Corwin, S.A. and Schultz, P. (2012). High-low spread estimation
//   - Various emerging market liquidity studies
//   - ISX-specific empirical analysis and calibration
//
// For detailed methodology and validation results, see the accompanying
// research documentation and technical specifications.
package liquidity