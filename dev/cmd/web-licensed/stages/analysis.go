package stages

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"isxcli/internal/common"
	"isxcli/internal/pipeline"
)

// TickerSummary holds summary statistics for a ticker
type TickerSummary struct {
	Ticker        string  `json:"ticker"`
	LastPrice     float64 `json:"last_price"`
	LastDate      string  `json:"last_date"`
	TotalVolume   float64 `json:"total_volume"`
	TotalTrades   int     `json:"total_trades"`
	TradingDays   int     `json:"trading_days"`
	AvgPrice      float64 `json:"avg_price"`
	MinPrice      float64 `json:"min_price"`
	MaxPrice      float64 `json:"max_price"`
	StdDev        float64 `json:"std_dev"`
	Prices        []float64 `json:"-"` // Used for calculations, not exported
}

// CalculateStatistics calculates the summary statistics
func (t *TickerSummary) CalculateStatistics() {
	if len(t.Prices) == 0 {
		return
	}
	
	// Calculate average
	sum := 0.0
	for _, p := range t.Prices {
		sum += p
	}
	t.AvgPrice = sum / float64(len(t.Prices))
	
	// Find min and max
	t.MinPrice = t.Prices[0]
	t.MaxPrice = t.Prices[0]
	for _, p := range t.Prices {
		if p < t.MinPrice {
			t.MinPrice = p
		}
		if p > t.MaxPrice {
			t.MaxPrice = p
		}
	}
	
	// Calculate standard deviation
	sumSquaredDiff := 0.0
	for _, p := range t.Prices {
		diff := p - t.AvgPrice
		sumSquaredDiff += diff * diff
	}
	if len(t.Prices) > 1 {
		t.StdDev = math.Sqrt(sumSquaredDiff / float64(len(t.Prices)-1))
	}
}

// AnalysisStage implements the ticker analysis stage
type AnalysisStage struct {
	pipeline.BaseStage
	executableDir string
	logger        *common.Logger
}

// NewAnalysisStage creates a new analysis stage
func NewAnalysisStage(executableDir string, logger *common.Logger) *AnalysisStage {
	return &AnalysisStage{
		BaseStage:     pipeline.NewBaseStage(pipeline.StageIDAnalysis, pipeline.StageNameAnalysis, []string{pipeline.StageIDIndices}),
		executableDir: executableDir,
		logger:        logger,
	}
}

// Execute runs the ticker analysis to generate summary statistics
func (a *AnalysisStage) Execute(ctx context.Context, state *pipeline.PipelineState) error {
	stageState := state.GetStage(a.ID())
	
	// Get report directory
	reportDirI, _ := state.GetConfig(pipeline.ContextKeyReportDir)
	reportDir, ok := reportDirI.(string)
	if !ok || reportDir == "" {
		reportDir = filepath.Join(a.executableDir, "data", "reports")
	}
	
	// File paths
	combinedFile := filepath.Join(reportDir, "isx_combined_data.csv")
	summaryJSONFile := filepath.Join(reportDir, "ticker_summary.json")
	
	// Debug logging
	a.logger.Info("[ANALYSIS] Report directory: %s", reportDir)
	a.logger.Info("[ANALYSIS] Combined file: %s", combinedFile)
	a.logger.Info("[ANALYSIS] Summary file: %s", summaryJSONFile)
	
	// Update progress
	stageState.UpdateProgress(10, "Starting ticker summary generation...")
	stageState.Metadata["input_file"] = combinedFile
	stageState.Metadata["output_file"] = summaryJSONFile
	
	// Check for cancellation
	select {
	case <-ctx.Done():
		return pipeline.NewCancellationError(a.ID())
	default:
	}
	
	// Generate the ticker summary
	err := a.generateTickerSummary(ctx, stageState, combinedFile, summaryJSONFile)
	if err != nil {
		a.logger.Error("[ANALYSIS] Failed to generate ticker summary: %v", err)
		return pipeline.NewExecutionError(a.ID(), fmt.Errorf("failed to generate ticker summary: %v", err), false)
	}
	
	stageState.UpdateProgress(100, "Ticker analysis completed successfully")
	
	return nil
}

// Validate checks if the stage can be executed
func (a *AnalysisStage) Validate(state *pipeline.PipelineState) error {
	// Check if combined data file exists
	reportDirI, _ := state.GetConfig(pipeline.ContextKeyReportDir)
	reportDir, ok := reportDirI.(string)
	if !ok || reportDir == "" {
		reportDir = filepath.Join(a.executableDir, "data", "reports")
	}
	
	combinedFile := filepath.Join(reportDir, "isx_combined_data.csv")
	if _, err := os.Stat(combinedFile); err != nil {
		return pipeline.NewValidationError(a.ID(), "combined data file not found - processing stage may have failed")
	}
	
	return nil
}

// generateTickerSummary creates the ticker summary JSON file
func (a *AnalysisStage) generateTickerSummary(ctx context.Context, stageState *pipeline.StageState, combinedFile, summaryFile string) error {
	// Update progress
	stageState.UpdateProgress(20, "Reading combined data file...")
	
	// Read combined CSV
	file, err := os.Open(combinedFile)
	if err != nil {
		return fmt.Errorf("failed to open combined file: %v", err)
	}
	defer file.Close()
	
	// Skip BOM if present
	buf := make([]byte, 3)
	n, _ := file.Read(buf)
	if n < 3 || (buf[0] != 0xEF || buf[1] != 0xBB || buf[2] != 0xBF) {
		// Not a BOM, seek back to start
		file.Seek(0, 0)
	}
	
	reader := csv.NewReader(file)
	
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read combined CSV: %v", err)
	}
	
	if len(records) < 2 {
		return fmt.Errorf("no data found in combined CSV")
	}
	
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	stageState.UpdateProgress(40, "Processing ticker data...")
	
	// Get header indices
	header := records[0]
	indices := make(map[string]int)
	for i, col := range header {
		indices[strings.ToLower(col)] = i
	}
	
	// Map expected column names to actual CSV column names
	columnMapping := map[string][]string{
		"ticker": {"symbol", "ticker"},
		"date": {"date"},
		"closing_price": {"closeprice", "closing_price", "close_price"},
		"volume": {"volume"},
		"trades": {"numtrades", "trades", "num_trades"},
	}
	
	// Find actual column names
	actualColumns := make(map[string]int)
	for expected, alternatives := range columnMapping {
		found := false
		for _, alt := range alternatives {
			if idx, ok := indices[strings.ToLower(alt)]; ok {
				actualColumns[expected] = idx
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required column '%s' not found in CSV (tried: %v)", expected, alternatives)
		}
	}
	
	// Process data
	tickerMap := make(map[string]*TickerSummary)
	totalRecords := len(records) - 1
	
	for i, record := range records[1:] {
		// Update progress periodically
		if i%1000 == 0 {
			progress := 40 + (float64(i)/float64(totalRecords))*40
			stageState.UpdateProgress(progress, fmt.Sprintf("Processing ticker %d of %d", i, totalRecords))
			
			// Check for cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		
		ticker := record[actualColumns["ticker"]]
		
		// Skip empty tickers
		if ticker == "" {
			continue
		}
		
		// Get or create ticker summary
		summary, exists := tickerMap[ticker]
		if !exists {
			summary = &TickerSummary{
				Ticker: ticker,
				Prices: []float64{},
			}
			tickerMap[ticker] = summary
		}
		
		// Parse data
		price, _ := strconv.ParseFloat(record[actualColumns["closing_price"]], 64)
		volume, _ := strconv.ParseFloat(record[actualColumns["volume"]], 64)
		trades, _ := strconv.Atoi(record[actualColumns["trades"]])
		
		// Add to summary
		if price > 0 {
			summary.Prices = append(summary.Prices, price)
			summary.TotalVolume += volume
			summary.TotalTrades += trades
			summary.TradingDays++
		}
		
		// Track latest date
		dateStr := record[actualColumns["date"]]
		if dateStr > summary.LastDate {
			summary.LastDate = dateStr
			summary.LastPrice = price
		}
	}
	
	stageState.UpdateProgress(80, "Calculating statistics...")
	
	// Calculate statistics for each ticker
	summaries := make([]*TickerSummary, 0, len(tickerMap))
	for _, summary := range tickerMap {
		if len(summary.Prices) > 0 {
			summary.CalculateStatistics()
			summaries = append(summaries, summary)
		}
	}
	
	// Sort by ticker name
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})
	
	stageState.UpdateProgress(90, "Writing summary file...")
	stageState.Metadata["tickers_analyzed"] = len(summaries)
	
	// Write JSON file
	outputData := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"total_tickers": len(summaries),
		"tickers": summaries,
	}
	
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	
	err = os.WriteFile(summaryFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write summary file: %v", err)
	}
	
	stageState.UpdateProgress(95, fmt.Sprintf("Generated summary for %d tickers", len(summaries)))
	
	return nil
}