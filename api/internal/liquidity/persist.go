package liquidity

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// SaveToCSV saves liquidity metrics to a CSV file with comprehensive formatting
// This creates a standardized output format for further analysis
// In accumulative mode, it merges with existing data
func SaveToCSV(metrics []TickerMetrics, outputPath string) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics to save")
	}
	
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	
	// Load existing metrics if file exists and merge with new ones
	var allMetrics []TickerMetrics
	if _, err := os.Stat(outputPath); err == nil {
		// File exists, load existing metrics
		existingMetrics, err := LoadFromCSV(outputPath)
		if err == nil {
			// Create a map to track which (Date, Symbol, Window) tuples we've updated
			type metricKey struct {
				Symbol string
				Date   time.Time
				Window Window
			}
			updatedKeys := make(map[metricKey]bool)
			for _, m := range metrics {
				key := metricKey{
					Symbol: m.Symbol,
					Date:   m.Date,
					Window: m.Window,
				}
				updatedKeys[key] = true
			}

			// Keep existing metrics that weren't updated (preserves historical data)
			for _, existing := range existingMetrics {
				key := metricKey{
					Symbol: existing.Symbol,
					Date:   existing.Date,
					Window: existing.Window,
				}
				if !updatedKeys[key] {
					allMetrics = append(allMetrics, existing)
				}
			}
		}
	}
	
	// Add all new/updated metrics
	allMetrics = append(allMetrics, metrics...)
	
	// Create CSV file (overwrite with merged data)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create CSV file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header - Including safe trading values
	header := []string{
		"Date",
		"Symbol", 
		"Window",
		"ILLIQ_Raw",
		"ILLIQ_Scaled",
		"Value_Raw",
		"Value_Scaled", 
		"Continuity_Raw",
		"Continuity_Scaled",
		"Activity_Score",     // Unified activity score (0-1)
		"Spread_Proxy",
		"Spread_Scaled",      // Scaled spread for completeness
		"Hybrid_Score",
		"Hybrid_Rank",
		"Trading_Days",
		"Data_Quality",
		"Safe_Trade_0.5%",    // Safe trading value for 0.5% impact
		"Safe_Trade_1%",      // Safe trading value for 1% impact
		"Safe_Trade_2%",      // Safe trading value for 2% impact
		"Optimal_Trade",      // Optimal trade size recommendation
	}
	
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}
	
	// Sort metrics by date then by symbol for consistent output
	sort.Slice(allMetrics, func(i, j int) bool {
		if allMetrics[i].Date.Equal(allMetrics[j].Date) {
			return allMetrics[i].Symbol < allMetrics[j].Symbol
		}
		return allMetrics[i].Date.Before(allMetrics[j].Date)
	})
	
	// Write data rows
	for _, metric := range allMetrics {
		record, err := formatMetricRecord(metric)
		if err != nil {
			return fmt.Errorf("format metric record for %s: %w", metric.Symbol, err)
		}
		
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write CSV record for %s: %w", metric.Symbol, err)
		}
	}
	
	return nil
}

// LoadFromCSV loads existing liquidity metrics from a CSV file
func LoadFromCSV(filepath string) ([]TickerMetrics, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open CSV file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	
	var metrics []TickerMetrics
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}
		
		if len(record) < 19 { // Ensure we have all required fields
			continue
		}
		
		// Parse date
		date, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			continue // Skip malformed records
		}
		
		// Parse window
		window := ParseWindow(record[2])
		
		// Parse numeric fields
		metric := TickerMetrics{
			Date:              date,
			Symbol:            record[1],
			Window:            window,
			ILLIQ:             parseFloatSafe(record[3]),
			ILLIQScaled:       parseFloatSafe(record[4]),
			Value:             parseFloatSafe(record[5]),
			ValueScaled:       parseFloatSafe(record[6]),
			Continuity:        parseFloatSafe(record[7]),
			ContinuityScaled:  parseFloatSafe(record[8]),
			ActivityScore:     parseFloatSafe(record[9]),
			SpreadProxy:       parseFloatSafe(record[10]),
			SpreadScaled:      parseFloatSafe(record[11]),
			HybridScore:       parseFloatSafe(record[12]),
			HybridRank:        parseIntSafe(record[13]),
			TradingDays:       parseIntSafe(record[14]),
			// Data quality is calculated, not stored
			SafeValue_0_5:     parseFloatSafe(record[16]),
			SafeValue_1_0:     parseFloatSafe(record[17]),
			SafeValue_2_0:     parseFloatSafe(record[18]),
		}
		
		// Set OptimalTradeSize if available (might be in column 19)
		if len(record) > 19 {
			metric.OptimalTradeSize = parseFloatSafe(record[19])
		}
		
		metrics = append(metrics, metric)
	}
	
	return metrics, nil
}

// parseFloatSafe is a helper to parse float values from strings
func parseFloatSafe(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// parseIntSafe is a helper to parse int values from strings
func parseIntSafe(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// ParseWindow parses a window string (e.g., "20d", "60d", "120d") into a Window type
func ParseWindow(s string) Window {
	switch s {
	case "20d":
		return Window20
	case "60d":
		return Window60
	case "120d":
		return Window120
	default:
		// Default to 20-day window if parsing fails
		return Window20
	}
}

// formatMetricRecord converts a TickerMetrics struct to CSV record
func formatMetricRecord(metric TickerMetrics) ([]string, error) {
	// Calculate derived fields
	dataQuality := calculateDataQuality(metric)
	
	// Record including safe trading values
	record := []string{
		metric.Date.Format("2006-01-02"),
		metric.Symbol,
		metric.Window.String(),
		formatFloat(metric.ILLIQ, 8),
		formatFloat(metric.ILLIQScaled, 2),
		formatFloat(metric.Value, 0),
		formatFloat(metric.ValueScaled, 2),
		formatFloat(metric.Continuity, 4),
		formatFloat(metric.ContinuityScaled, 2),
		formatFloat(metric.ActivityScore, 4),      // Unified activity score
		formatFloat(metric.SpreadProxy, 6),
		formatFloat(metric.SpreadScaled, 2),        // Scaled spread
		formatFloat(metric.HybridScore, 4),
		strconv.Itoa(metric.HybridRank),
		strconv.Itoa(metric.TradingDays),
		dataQuality,
		formatFloat(metric.SafeValue_0_5, 0),       // Safe trade for 0.5% impact
		formatFloat(metric.SafeValue_1_0, 0),       // Safe trade for 1% impact
		formatFloat(metric.SafeValue_2_0, 0),       // Safe trade for 2% impact
		formatFloat(metric.OptimalTradeSize, 0),    // Optimal trade size
	}
	
	return record, nil
}

// calculateDataQuality assigns a quality score to the metric
func calculateDataQuality(metric TickerMetrics) string {
	tradingRatio := float64(metric.TradingDays) / float64(metric.TotalDays)
	
	switch {
	case tradingRatio >= 0.8 && metric.TradingDays >= metric.Window.Days()*3/4:
		return "HIGH"
	case tradingRatio >= 0.5 && metric.TradingDays >= metric.Window.Days()/2:
		return "MEDIUM"
	case tradingRatio >= 0.2 && metric.TradingDays >= MinTradingDaysForCalc:
		return "LOW"
	default:
		return "POOR"
	}
}

// formatFloat formats a float64 value for CSV output with specified precision
func formatFloat(value float64, precision int) string {
	if precision == 0 {
		return strconv.FormatFloat(value, 'f', 0, 64)
	}
	return strconv.FormatFloat(value, 'f', precision, 64)
}

// SaveToJSON saves liquidity metrics to a JSON file with structured format
func SaveToJSON(metrics []TickerMetrics, outputPath string) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics to save")
	}
	
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	
	// Create structured output format
	output := map[string]interface{}{
		"metadata": map[string]interface{}{
			"generated_at":    time.Now().Format(time.RFC3339),
			"total_records":   len(metrics),
			"unique_symbols":  countUniqueSymbols(metrics),
			"date_range":     getDateRange(metrics),
			"windows":        getWindows(metrics),
		},
		"metrics": metrics,
	}
	
	// Create JSON file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create JSON file: %w", err)
	}
	defer file.Close()
	
	// Write JSON with pretty printing
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	
	return nil
}

// SaveSummaryReport creates a summary report of the liquidity analysis
func SaveSummaryReport(metrics []TickerMetrics, outputPath string) error {
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics to save")
	}
	
	// Calculate summary statistics
	summary := calculateSummaryStatistics(metrics)
	
	// Create summary report file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create summary file: %w", err)
	}
	defer file.Close()
	
	// Write summary report
	fmt.Fprintf(file, "ISX Hybrid Liquidity Metric - Summary Report\n")
	fmt.Fprintf(file, "============================================\n\n")
	fmt.Fprintf(file, "Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	fmt.Fprintf(file, "DATASET OVERVIEW\n")
	fmt.Fprintf(file, "----------------\n")
	fmt.Fprintf(file, "Total Records: %d\n", summary.TotalRecords)
	fmt.Fprintf(file, "Unique Symbols: %d\n", summary.UniqueSymbols)
	fmt.Fprintf(file, "Date Range: %s to %s\n", summary.StartDate, summary.EndDate)
	fmt.Fprintf(file, "Windows: %s\n\n", summary.Windows)
	
	fmt.Fprintf(file, "HYBRID SCORE STATISTICS\n")
	fmt.Fprintf(file, "-----------------------\n")
	fmt.Fprintf(file, "Mean: %.4f\n", summary.HybridScore.Mean)
	fmt.Fprintf(file, "Median: %.4f\n", summary.HybridScore.Median)
	fmt.Fprintf(file, "Std Dev: %.4f\n", summary.HybridScore.StdDev)
	fmt.Fprintf(file, "Min: %.4f (%s)\n", summary.HybridScore.Min, summary.HybridScore.MinSymbol)
	fmt.Fprintf(file, "Max: %.4f (%s)\n\n", summary.HybridScore.Max, summary.HybridScore.MaxSymbol)
	
	fmt.Fprintf(file, "COMPONENT STATISTICS\n")
	fmt.Fprintf(file, "--------------------\n")
	fmt.Fprintf(file, "ILLIQ (Scaled) - Mean: %.4f, Median: %.4f\n", 
		summary.ILLIQScaled.Mean, summary.ILLIQScaled.Median)
	fmt.Fprintf(file, "Value (Scaled) - Mean: %.4f, Median: %.4f\n", 
		summary.ValueScaled.Mean, summary.ValueScaled.Median)
	fmt.Fprintf(file, "Continuity (Scaled) - Mean: %.4f, Median: %.4f\n\n", 
		summary.ContinuityScaled.Mean, summary.ContinuityScaled.Median)
	
	fmt.Fprintf(file, "DATA QUALITY DISTRIBUTION\n")
	fmt.Fprintf(file, "-------------------------\n")
	for quality, count := range summary.DataQuality {
		percentage := float64(count) / float64(summary.TotalRecords) * 100
		fmt.Fprintf(file, "%s: %d (%.1f%%)\n", quality, count, percentage)
	}
	fmt.Fprintf(file, "\n")
	
	fmt.Fprintf(file, "TOP 10 MOST LIQUID (Highest Hybrid Score)\n")
	fmt.Fprintf(file, "------------------------------------------\n")
	for i, ticker := range summary.TopLiquid {
		fmt.Fprintf(file, "%2d. %s: %.4f\n", i+1, ticker.Symbol, ticker.HybridScore)
	}
	fmt.Fprintf(file, "\n")
	
	fmt.Fprintf(file, "TOP 10 LEAST LIQUID (Lowest Hybrid Score)\n")
	fmt.Fprintf(file, "------------------------------------------\n")
	for i, ticker := range summary.LeastLiquid {
		fmt.Fprintf(file, "%2d. %s: %.4f\n", i+1, ticker.Symbol, ticker.HybridScore)
	}
	
	return nil
}

// SummaryStatistics holds summary statistics for the report
type SummaryStatistics struct {
	TotalRecords   int
	UniqueSymbols  int
	StartDate      string
	EndDate        string
	Windows        string
	HybridScore    StatsSummary
	ILLIQScaled    StatsSummary
	ValueScaled    StatsSummary
	ContinuityScaled StatsSummary
	DataQuality    map[string]int
	TopLiquid      []TickerMetrics
	LeastLiquid    []TickerMetrics
}

// StatsSummary holds statistical summary for a metric
type StatsSummary struct {
	Mean      float64
	Median    float64
	StdDev    float64
	Min       float64
	Max       float64
	MinSymbol string
	MaxSymbol string
}

// calculateSummaryStatistics computes comprehensive summary statistics
func calculateSummaryStatistics(metrics []TickerMetrics) SummaryStatistics {
	if len(metrics) == 0 {
		return SummaryStatistics{}
	}
	
	// Sort by hybrid score for ranking
	sortedMetrics := make([]TickerMetrics, len(metrics))
	copy(sortedMetrics, metrics)
	sort.Slice(sortedMetrics, func(i, j int) bool {
		return sortedMetrics[i].HybridScore > sortedMetrics[j].HybridScore
	})
	
	// Extract values for statistics
	var hybridScores, illiqScaled, valueScaled, continuityScaled []float64
	dataQuality := make(map[string]int)
	
	for _, metric := range metrics {
		hybridScores = append(hybridScores, metric.HybridScore)
		illiqScaled = append(illiqScaled, metric.ILLIQScaled)
		valueScaled = append(valueScaled, metric.ValueScaled)
		continuityScaled = append(continuityScaled, metric.ContinuityScaled)
		
		quality := calculateDataQuality(metric)
		dataQuality[quality]++
	}
	
	return SummaryStatistics{
		TotalRecords:     len(metrics),
		UniqueSymbols:    countUniqueSymbols(metrics),
		StartDate:        getEarliestDate(metrics),
		EndDate:          getLatestDate(metrics),
		Windows:          getWindowsSummary(metrics),
		HybridScore:      calculateStats(hybridScores, metrics, func(m TickerMetrics) float64 { return m.HybridScore }),
		ILLIQScaled:      calculateStats(illiqScaled, metrics, func(m TickerMetrics) float64 { return m.ILLIQScaled }),
		ValueScaled:      calculateStats(valueScaled, metrics, func(m TickerMetrics) float64 { return m.ValueScaled }),
		ContinuityScaled: calculateStats(continuityScaled, metrics, func(m TickerMetrics) float64 { return m.ContinuityScaled }),
		DataQuality:      dataQuality,
		TopLiquid:        getTopN(sortedMetrics, 10, true),
		LeastLiquid:      getTopN(sortedMetrics, 10, false),
	}
}

// calculateStats computes statistical summary for a slice of values
func calculateStats(values []float64, metrics []TickerMetrics, extractor func(TickerMetrics) float64) StatsSummary {
	if len(values) == 0 {
		return StatsSummary{}
	}
	
	// Calculate basic statistics
	mean := calculateMean(values)
	median := calculateMedian(values)
	stdDev := calculateStandardDeviation(values, mean)
	
	// Find min and max with symbols
	minVal, maxVal := values[0], values[0]
	minSymbol, maxSymbol := metrics[0].Symbol, metrics[0].Symbol
	
	for i, val := range values {
		if val < minVal {
			minVal = val
			minSymbol = metrics[i].Symbol
		}
		if val > maxVal {
			maxVal = val
			maxSymbol = metrics[i].Symbol
		}
	}
	
	return StatsSummary{
		Mean:      mean,
		Median:    median,
		StdDev:    stdDev,
		Min:       minVal,
		Max:       maxVal,
		MinSymbol: minSymbol,
		MaxSymbol: maxSymbol,
	}
}

// Helper functions for summary calculations
func countUniqueSymbols(metrics []TickerMetrics) int {
	symbols := make(map[string]bool)
	for _, metric := range metrics {
		symbols[metric.Symbol] = true
	}
	return len(symbols)
}

func getDateRange(metrics []TickerMetrics) string {
	if len(metrics) == 0 {
		return "N/A"
	}
	
	earliest := getEarliestDate(metrics)
	latest := getLatestDate(metrics)
	return fmt.Sprintf("%s to %s", earliest, latest)
}

func getEarliestDate(metrics []TickerMetrics) string {
	if len(metrics) == 0 {
		return "N/A"
	}
	
	earliest := metrics[0].Date
	for _, metric := range metrics[1:] {
		if metric.Date.Before(earliest) {
			earliest = metric.Date
		}
	}
	return earliest.Format("2006-01-02")
}

func getLatestDate(metrics []TickerMetrics) string {
	if len(metrics) == 0 {
		return "N/A"
	}
	
	latest := metrics[0].Date
	for _, metric := range metrics[1:] {
		if metric.Date.After(latest) {
			latest = metric.Date
		}
	}
	return latest.Format("2006-01-02")
}

func getWindows(metrics []TickerMetrics) []string {
	windows := make(map[Window]bool)
	for _, metric := range metrics {
		windows[metric.Window] = true
	}
	
	var result []string
	for window := range windows {
		result = append(result, window.String())
	}
	sort.Strings(result)
	return result
}

func getWindowsSummary(metrics []TickerMetrics) string {
	windows := getWindows(metrics)
	if len(windows) == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%v", windows)
}

func getTopN(sortedMetrics []TickerMetrics, n int, fromTop bool) []TickerMetrics {
	if len(sortedMetrics) == 0 {
		return nil
	}
	
	if n > len(sortedMetrics) {
		n = len(sortedMetrics)
	}
	
	if fromTop {
		return sortedMetrics[:n]
	} else {
		// Return bottom n (least liquid)
		start := len(sortedMetrics) - n
		if start < 0 {
			start = 0
		}
		return sortedMetrics[start:]
	}
}

// ExportCalibrationResults saves calibration results to JSON file
func ExportCalibrationResults(result *CalibrationResult, outputPath string) error {
	if result == nil {
		return fmt.Errorf("no calibration results to export")
	}
	
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create calibration results file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode calibration results: %w", err)
	}
	
	return nil
}