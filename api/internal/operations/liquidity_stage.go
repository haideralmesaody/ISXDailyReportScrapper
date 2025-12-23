package operations

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
)

// LiquidityStage handles liquidity calculation
type LiquidityStage struct {
	BaseStage
	executableDir string
	logger        *slog.Logger
	options       *StageOptions
	broadcaster   StageBroadcaster
}

// NewLiquidityStage creates a new liquidity calculation step
func NewLiquidityStage(executableDir string, logger *slog.Logger, options *StageOptions) *LiquidityStage {
	if options == nil {
		options = &StageOptions{}
	}

	// Create logger with Step context
	if logger != nil {
		logger = logger.With(slog.String("Step", StageIDLiquidity))
		logger.Info("Liquidity calculation step initialized",
			slog.String("executable_dir", executableDir))
	}

	// Create stage broadcaster
	var broadcaster StageBroadcaster
	if options.WebSocketManager != nil {
		broadcaster = NewBaseStageBroadcaster("temp", StageIDLiquidity, options.WebSocketManager, logger, options.EnableProgress)
	}

	return &LiquidityStage{
		BaseStage:     NewBaseStage(StageIDLiquidity, StageNameLiquidity, []string{StageIDIndices}), // Depends on indices (for CSV files)
		executableDir: executableDir,
		logger:        logger,
		options:       options,
		broadcaster:   broadcaster,
	}
}

// setOperationID updates the broadcaster with the actual operation ID
func (l *LiquidityStage) setOperationID(operationID string) {
	if l.broadcaster != nil {
		if baseBroadcaster, ok := l.broadcaster.(*BaseStageBroadcaster); ok {
			// Create a new broadcaster with the correct operation ID
			l.broadcaster = NewBaseStageBroadcaster(
				operationID,
				StageIDLiquidity,
				baseBroadcaster.hub,
				l.logger,
				baseBroadcaster.IsEnabled(),
			)
		}
	}
}

// Execute runs the liquidity calculation
func (l *LiquidityStage) Execute(ctx context.Context, state *OperationState) error {
	// Update broadcaster with actual operation ID
	l.setOperationID(state.ID)

	StepState := state.GetStage(l.ID())
	if StepState == nil {
		return fmt.Errorf("stage state not found for stage %s", l.ID())
	}
	if StepState.StartTime == nil {
		now := time.Now()
		StepState.StartTime = &now
	}

	// Log step execution start
	if l.logger != nil {
		l.logger.InfoContext(ctx, "Liquidity calculation step started",
			slog.String("operation_id", state.ID))
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return l.failStep(StepState, "Liquidity calculation cancelled", ctx.Err())
	default:
	}

	// 1. Load CSV files from reports/daily directory (where processor.exe outputs them)
	// Use centralized path resolution to ensure dist/dev/prod all read the same location.
	reportsDir := AbsoluteReportsPath(l.executableDir, "")
	dailyDir := filepath.Join(reportsDir, "daily")
	csvFiles, err := filepath.Glob(filepath.Join(dailyDir, "*.csv"))
	if err != nil {
		if l.logger != nil {
			l.logger.ErrorContext(ctx, "Failed to scan reports directory",
				slog.String("directory", reportsDir),
				slog.String("error", err.Error()))
		}
		return l.failStep(StepState, "Failed to scan reports directory", err)
	}

	if len(csvFiles) == 0 {
		if l.logger != nil {
			l.logger.WarnContext(ctx, "No CSV files found for liquidity calculation",
				slog.String("directory", reportsDir))
		}
		return l.failStep(StepState, "No CSV files found to calculate liquidity", fmt.Errorf("reports_dir=%s", reportsDir))
	}

	// Estimate tickers/companies from ticker history outputs (used for early KPI display).
	tickerDir := filepath.Join(reportsDir, "ticker")
	tickerFiles, _ := filepath.Glob(filepath.Join(tickerDir, "*_trading_history.csv"))
	companiesEstimated := len(tickerFiles)

	StepState.mu.Lock()
	if StepState.Metadata == nil {
		StepState.Metadata = make(map[string]interface{})
	}
	StepState.Metadata["analysis_files"] = 0
	StepState.Metadata["companies_analyzed"] = companiesEstimated
	StepState.mu.Unlock()

	updateStageProgress(l.ID(), StepState, l.broadcaster, 10, "Starting liquidity calculation...", "running", 0, len(csvFiles), 0, "")

	if l.logger != nil {
		l.logger.InfoContext(ctx, "Found CSV files for liquidity calculation",
			slog.String("directory", reportsDir),
			slog.Int("file_count", len(csvFiles)))
	}

	// Initialize liquidity calculator with default parameters
	calculator := liquidity.NewCalculator(
		liquidity.Window60, // 60-day window for better liquidity measurement
		liquidity.DefaultPenaltyParams(),
		liquidity.DefaultWeights(),
		l.logger,
	)

	// 2. Process each CSV file and collect TradingDay data
	var allTradingDays []liquidity.TradingDay
	var filesSkipped int

	StepState.mu.Lock()
	StepState.Metadata["analysis_files"] = 0
	StepState.Metadata["companies_analyzed"] = companiesEstimated
	StepState.mu.Unlock()
	updateStageProgress(l.ID(), StepState, l.broadcaster, 20, fmt.Sprintf("Processing %d CSV files...", len(csvFiles)), "running", 0, len(csvFiles), 0, "")

	for i, csvFile := range csvFiles {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return l.failStep(StepState, "Liquidity calculation cancelled", ctx.Err())
		default:
		}

		progress := 20 + int((float64(i)/float64(len(csvFiles)))*40) // 20% to 60%

		StepState.mu.Lock()
		StepState.Metadata["analysis_files"] = i + 1
		StepState.Metadata["companies_analyzed"] = companiesEstimated
		StepState.mu.Unlock()

		updateStageProgress(
			l.ID(),
			StepState,
			l.broadcaster,
			progress,
			fmt.Sprintf("Reading %s...", filepath.Base(csvFile)),
			"running",
			i+1,
			len(csvFiles),
			0,
			filepath.Base(csvFile),
		)

		tradingDays, err := l.processCSVToTradingDays(csvFile)
		if err != nil {
			if l.logger != nil {
				l.logger.WarnContext(ctx, "Failed to process CSV file",
					slog.String("file", csvFile),
					slog.String("error", err.Error()))
			}
			filesSkipped++
			continue // Skip this file and continue with others
		}

		allTradingDays = append(allTradingDays, tradingDays...)
	}

	if len(allTradingDays) == 0 {
		return l.failStep(StepState, "No valid trading data found in CSV files", fmt.Errorf("files_skipped=%d", filesSkipped))
	}

	if l.logger != nil {
		l.logger.InfoContext(ctx, "Collected trading data for liquidity calculation",
			slog.Int("total_trading_days", len(allTradingDays)),
			slog.Int("files_skipped", filesSkipped))
	}

	// 3. Calculate liquidity metrics using the Calculator
	StepState.mu.Lock()
	StepState.Metadata["analysis_files"] = len(csvFiles)
	StepState.Metadata["companies_analyzed"] = companiesEstimated
	StepState.mu.Unlock()
	updateStageProgress(l.ID(), StepState, l.broadcaster, 65, "Calculating liquidity metrics...", "running", len(csvFiles), len(csvFiles), 0, "")

	metrics, err := calculator.Calculate(ctx, allTradingDays)
	if err != nil {
		if l.logger != nil {
			l.logger.ErrorContext(ctx, "Liquidity calculation failed",
				slog.String("error", err.Error()))
		}
		return l.failStep(StepState, "Liquidity calculation failed", err)
	}

	if l.logger != nil {
		l.logger.InfoContext(ctx, "Liquidity calculation completed",
			slog.Int("metrics_calculated", len(metrics)))
	}

	// 4. Categorize results into liquidity buckets
	StepState.mu.Lock()
	StepState.Metadata["analysis_files"] = len(csvFiles)
	StepState.Metadata["companies_analyzed"] = companiesEstimated
	StepState.Metadata["metrics_calculated"] = len(metrics)
	StepState.mu.Unlock()
	updateStageProgress(l.ID(), StepState, l.broadcaster, 85, "Categorizing liquidity results...", "running", len(csvFiles), len(csvFiles), 0, "")

	snapshot := &LiquidityTelemetrySnapshot{
		TickersAnalyzed: 0,
		AverageValue:    0,
		MarketCapTotal:  0,
		LiquidityBuckets: map[string]int{
			"high":   0,
			"medium": 0,
			"low":    0,
		},
	}

	// Get unique tickers and their latest metrics
	latestMetrics := make(map[string]liquidity.TickerMetrics)
	var totalValue float64
	for _, m := range metrics {
		existing, exists := latestMetrics[m.Symbol]
		if !exists || m.Date.After(existing.Date) {
			latestMetrics[m.Symbol] = m
		}
		totalValue += m.Value
	}

	snapshot.TickersAnalyzed = len(latestMetrics)
	snapshot.LiquidityScores = len(metrics)
	if len(latestMetrics) > 0 {
		snapshot.AverageValue = totalValue / float64(len(metrics))
	}
	snapshot.MarketCapTotal = totalValue

	// Categorize by HybridScore (0-100 scale)
	for _, m := range latestMetrics {
		switch {
		case m.HybridScore >= 70:
			snapshot.LiquidityBuckets["high"]++
		case m.HybridScore >= 40:
			snapshot.LiquidityBuckets["medium"]++
		default:
			snapshot.LiquidityBuckets["low"]++
		}
	}

	snapshot.IlliquidCount = snapshot.LiquidityBuckets["low"]
	snapshot.LiquidCount = snapshot.LiquidityBuckets["high"] + snapshot.LiquidityBuckets["medium"]

	// 5. Save liquidity results to CSV
	StepState.mu.Lock()
	StepState.Metadata["analysis_files"] = len(csvFiles)
	StepState.Metadata["companies_analyzed"] = snapshot.TickersAnalyzed
	StepState.Metadata["tickers_analyzed"] = snapshot.TickersAnalyzed
	StepState.Metadata["liquidity_scores"] = snapshot.LiquidityScores
	StepState.Metadata["liquidity_buckets"] = snapshot.LiquidityBuckets
	StepState.mu.Unlock()

	updateStageProgress(l.ID(), StepState, l.broadcaster, 90, "Saving liquidity results...", "running", len(csvFiles), len(csvFiles), 0, "")

	// Legacy output (used by analysis stage and for quick inspection).
	legacyOutputFile := filepath.Join(reportsDir, "liquidity_metrics.csv")
	if err := l.saveTickerMetrics(metrics, legacyOutputFile); err != nil {
		if l.logger != nil {
			l.logger.ErrorContext(ctx, "Failed to save liquidity results",
				slog.String("file", legacyOutputFile),
				slog.String("error", err.Error()))
		}
		return l.failStep(StepState, "Failed to save liquidity results", err)
	}

	// Primary output (used by /api/liquidity/insights and the Liquidity page).
	// LiquidityService expects versioned files under data/reports/liquidity_reports:
	//   liquidity_scores_*.csv with the canonical header from internal/liquidity.SaveToCSV.
	liquidityReportsDir := filepath.Join(reportsDir, "liquidity_reports")
	outputStamp := time.Now().Format("2006-01-02_150405")
	scoresOutputFile := filepath.Join(liquidityReportsDir, fmt.Sprintf("liquidity_scores_%s.csv", outputStamp))
	if err := liquidity.SaveToCSV(metrics, scoresOutputFile); err != nil {
		if l.logger != nil {
			l.logger.ErrorContext(ctx, "Failed to save liquidity scores",
				slog.String("file", scoresOutputFile),
				slog.String("error", err.Error()))
		}
		return l.failStep(StepState, "Failed to save liquidity scores", err)
	}

	// 6. Final progress update with rich telemetry
	totalUnits := len(latestMetrics)
	fileStatuses := make([]*FileProcessingStatus, 0, totalUnits)
	for ticker := range latestMetrics {
		fileStatuses = append(fileStatuses, &FileProcessingStatus{
			FileName: ticker,
			Status:   "completed",
			Progress: 100,
		})
	}
	StepState.mu.Lock()
	if StepState.Metadata == nil {
		StepState.Metadata = make(map[string]interface{})
	}
	StepState.Metadata["file_statuses"] = fileStatuses
	StepState.Metadata["files_processed"] = totalUnits
	StepState.Metadata["total_files"] = totalUnits
	StepState.Metadata["failed_files"] = 0
	StepState.Metadata["tickers_analyzed"] = snapshot.TickersAnalyzed
	StepState.Metadata["liquidity_buckets"] = snapshot.LiquidityBuckets
	StepState.Metadata["average_value"] = snapshot.AverageValue
	StepState.Metadata["analysis_files"] = len(csvFiles)
	StepState.Metadata["companies_analyzed"] = snapshot.TickersAnalyzed
	StepState.Metadata["files_generated"] = 2
	StepState.mu.Unlock()

	completionMsg := fmt.Sprintf("Liquidity calculated: %d tickers (High: %d, Medium: %d, Low: %d)",
		snapshot.TickersAnalyzed,
		snapshot.LiquidityBuckets["high"],
		snapshot.LiquidityBuckets["medium"],
		snapshot.LiquidityBuckets["low"])

	updateStageProgress(l.ID(), StepState, l.broadcaster, 100, completionMsg, "completed", totalUnits, maxInt(totalUnits, 1), 0, "")
	finalizeStageSuccess(l.ID(), StepState, l.broadcaster, completionMsg)

	if l.logger != nil {
		l.logger.InfoContext(ctx, "Liquidity calculation completed successfully",
			slog.Int("tickers_analyzed", snapshot.TickersAnalyzed),
			slog.Int("total_metrics", len(metrics)),
			slog.Int("high_liquidity", snapshot.LiquidityBuckets["high"]),
			slog.Int("medium_liquidity", snapshot.LiquidityBuckets["medium"]),
			slog.Int("low_liquidity", snapshot.LiquidityBuckets["low"]),
			slog.Int("files_skipped", filesSkipped))
	}

	return nil
}

// processCSVToTradingDays reads a CSV file and returns TradingDay structures for liquidity calculation
func (l *LiquidityStage) processCSVToTradingDays(csvFile string) ([]liquidity.TradingDay, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var tradingDays []liquidity.TradingDay

	// Stream rows to keep memory bounded
	headerRead := false
	for {
		record, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("failed to read CSV records: %w", readErr)
		}
		if !headerRead {
			headerRead = true
			continue
		}

		if len(record) < 15 {
			continue // Skip incomplete records (CSV has 16 columns)
		}

		// CSV columns: Date(0), CompanyName(1), Symbol(2), OpenPrice(3), HighPrice(4),
		// LowPrice(5), AveragePrice(6), PrevAveragePrice(7), ClosePrice(8), PrevClosePrice(9),
		// Change(10), ChangePercent(11), NumTrades(12), Volume(13), Value(14), TradingStatus(15)

		// Parse date (format: YYYY-MM-DD)
		date, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			continue // Skip records with invalid date
		}

		symbol := record[2]

		openPrice, _ := strconv.ParseFloat(record[3], 64)
		highPrice, _ := strconv.ParseFloat(record[4], 64)
		lowPrice, _ := strconv.ParseFloat(record[5], 64)
		closePrice, err := strconv.ParseFloat(record[8], 64)
		if err != nil {
			continue // Skip records with invalid close price
		}

		numTrades, _ := strconv.Atoi(record[12])
		volume, _ := strconv.ParseFloat(record[13], 64) // TradingDay uses float64 for volume
		value, err := strconv.ParseFloat(record[14], 64)
		if err != nil {
			value = closePrice * volume // Fallback calculation
		}

		tradingStatus := "ACTIVE"
		if len(record) > 15 && record[15] != "" {
			if record[15] == "false" {
				tradingStatus = "SUSPENDED"
			}
		}

		tradingDays = append(tradingDays, liquidity.TradingDay{
			Date:          date,
			Symbol:        symbol,
			Open:          openPrice,
			High:          highPrice,
			Low:           lowPrice,
			Close:         closePrice,
			Volume:        volume,
			ShareVolume:   volume,
			Value:         value,
			NumTrades:     numTrades,
			TradingStatus: tradingStatus,
		})
	}

	if len(tradingDays) == 0 {
		return nil, fmt.Errorf("CSV file has no valid data rows")
	}

	return tradingDays, nil
}

// saveTickerMetrics saves calculated liquidity metrics to a CSV file
func (l *LiquidityStage) saveTickerMetrics(metrics []liquidity.TickerMetrics, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header with comprehensive liquidity metrics
	header := []string{
		"symbol", "date", "hybrid_score", "hybrid_rank", "liquidity_category",
		"illiq", "illiq_scaled", "value", "value_scaled",
		"continuity", "continuity_scaled", "trading_days", "total_days",
		"safe_value_0_5", "safe_value_1_0", "safe_value_2_0", "optimal_trade_size",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Sort metrics by symbol and date for consistent output
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].Symbol != metrics[j].Symbol {
			return metrics[i].Symbol < metrics[j].Symbol
		}
		return metrics[i].Date.Before(metrics[j].Date)
	})

	for _, m := range metrics {
		category := "low"
		switch {
		case m.HybridScore >= 70:
			category = "high"
		case m.HybridScore >= 40:
			category = "medium"
		}

		record := []string{
			m.Symbol,
			m.Date.Format("2006-01-02"),
			strconv.FormatFloat(m.HybridScore, 'f', 4, 64),
			strconv.Itoa(m.HybridRank),
			category,
			strconv.FormatFloat(m.ILLIQ, 'f', 8, 64),
			strconv.FormatFloat(m.ILLIQScaled, 'f', 4, 64),
			strconv.FormatFloat(m.Value, 'f', 2, 64),
			strconv.FormatFloat(m.ValueScaled, 'f', 4, 64),
			strconv.FormatFloat(m.Continuity, 'f', 4, 64),
			strconv.FormatFloat(m.ContinuityScaled, 'f', 4, 64),
			strconv.Itoa(m.TradingDays),
			strconv.Itoa(m.TotalDays),
			strconv.FormatFloat(m.SafeValue_0_5, 'f', 2, 64),
			strconv.FormatFloat(m.SafeValue_1_0, 'f', 2, 64),
			strconv.FormatFloat(m.SafeValue_2_0, 'f', 2, 64),
			strconv.FormatFloat(m.OptimalTradeSize, 'f', 2, 64),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record for ticker %s: %w", m.Symbol, err)
		}
	}

	return nil
}

// LiquidityTelemetrySnapshot contains liquidity calculation telemetry data
type LiquidityTelemetrySnapshot struct {
	TickersAnalyzed  int            `json:"tickers_analyzed"`
	LiquidityScores  int            `json:"liquidity_scores"`
	LiquidityBuckets map[string]int `json:"liquidity_buckets"`
	AverageValue     float64        `json:"average_value"`
	MarketCapTotal   float64        `json:"market_cap_total"`
	IlliquidCount    int            `json:"illiquid_count"`
	LiquidCount      int            `json:"liquid_count"`
}

// updateProgress updates progress through the stage broadcaster
func (l *LiquidityStage) updateProgress(operationID string, stepState *StepState, progress int, message string, snapshot *LiquidityTelemetrySnapshot) {
	stepState.UpdateProgress(float64(progress), message)

	// Store snapshot in metadata if provided
	if snapshot != nil {
		l.storeLiquiditySnapshot(snapshot, stepState.Metadata)
	}

	// Use stage broadcaster for all updates
	if l.broadcaster != nil {
		l.broadcaster.UpdateProgressWithMetadata(progress, message, stepState.Metadata)
	}
}

// storeLiquiditySnapshot stores liquidity telemetry snapshot in metadata
func (l *LiquidityStage) storeLiquiditySnapshot(s *LiquidityTelemetrySnapshot, metadata map[string]interface{}) {
	if metadata == nil || s == nil {
		return
	}
	metadata["tickers_analyzed"] = s.TickersAnalyzed
	metadata["liquidity_scores"] = s.LiquidityScores
	buckets := make(map[string]int, len(s.LiquidityBuckets))
	for k, v := range s.LiquidityBuckets {
		buckets[k] = v
	}
	metadata["liquidity_buckets"] = buckets
	metadata["average_value"] = s.AverageValue
	metadata["market_cap_total"] = s.MarketCapTotal
	metadata["illiquid_count"] = s.IlliquidCount
	metadata["liquid_count"] = s.LiquidCount
	metadata["metrics_calculated"] = s.LiquidityScores
}

func (l *LiquidityStage) buildLiquidityTelemetry(stepState *StepState, progress int, message string, snapshot *LiquidityTelemetrySnapshot) *StageTelemetry {
	if stepState == nil {
		return nil
	}

	if snapshot == nil {
		snapshot = &LiquidityTelemetrySnapshot{
			LiquidityBuckets: map[string]int{
				"high":   0,
				"medium": 0,
				"low":    0,
			},
		}
	}

	phase := deriveLiquidityPhase(progress)

	return NewStageTelemetry(
		StageIDLiquidity,
		stepState,
		phase,
		message,
		float64(progress),
		snapshot.LiquidityScores,
		maxInt(snapshot.LiquidityScores, 1),
		"",
		nil,
	)
}

func extractLiquiditySnapshot(metadata map[string]interface{}) *LiquidityTelemetrySnapshot {
	snapshot := &LiquidityTelemetrySnapshot{
		LiquidityBuckets: map[string]int{
			"high":   0,
			"medium": 0,
			"low":    0,
		},
	}
	if metadata == nil {
		return snapshot
	}

	if v, ok := metadata["tickers_analyzed"]; ok {
		snapshot.TickersAnalyzed = pickInt(v)
	}
	if v, ok := metadata["liquidity_scores"]; ok {
		snapshot.LiquidityScores = pickInt(v)
	}
	if v, ok := metadata["average_value"]; ok {
		snapshot.AverageValue = pickFloat(v)
	}
	if v, ok := metadata["market_cap_total"]; ok {
		snapshot.MarketCapTotal = pickFloat(v)
	}
	if v, ok := metadata["illiquid_count"]; ok {
		snapshot.IlliquidCount = pickInt(v)
	}
	if v, ok := metadata["liquid_count"]; ok {
		snapshot.LiquidCount = pickInt(v)
	}
	if v, ok := metadata["liquidity_buckets"]; ok {
		snapshot.LiquidityBuckets = normalizeBucketMap(v)
	}

	return snapshot
}

func normalizeBucketMap(v interface{}) map[string]int {
	result := map[string]int{
		"high":   0,
		"medium": 0,
		"low":    0,
	}

	if m, ok := v.(map[string]int); ok {
		for key, val := range m {
			result[key] = val
		}
		return result
	}

	if m, ok := v.(map[string]interface{}); ok {
		for key, val := range m {
			result[key] = pickInt(val)
		}
	}

	return result
}

func (l *LiquidityStage) failStep(stepState *StepState, message string, err error) error {
	fullMessage := message
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", message, err)
	}
	progress := int(stepState.Progress)
	updateStageProgress(l.ID(), stepState, l.broadcaster, progress, fullMessage, "failed", 0, 0, 0, "")
	return fmt.Errorf(fullMessage)
}

// deriveLiquidityPhase determines the current phase based on progress percentage
func deriveLiquidityPhase(progress int) StagePhase {
	switch {
	case progress < 20:
		return StagePhaseStarting
	case progress < 85:
		return StagePhaseRunning
	case progress < 100:
		return StagePhaseCompleting
	default:
		return StagePhaseCompleted
	}
}

// GetPhase returns the current phase of the liquidity calculation
func (l *LiquidityStage) GetPhase(stepState *StepState) (StagePhase, string) {
	if stepState == nil {
		return StagePhasePending, "Ready to start"
	}

	progress := int(stepState.Progress)
	if progress < 20 {
		return StagePhaseStarting, "Initializing liquidity calculation"
	} else if progress < 85 {
		return StagePhaseRunning, fmt.Sprintf("Calculating liquidity metrics (%d%%)", progress)
	} else if progress < 100 {
		return StagePhaseCompleting, "Finalizing liquidity analysis"
	} else {
		return StagePhaseCompleted, "Liquidity calculation completed successfully"
	}
}

// CanRun checks if the liquidity stage can run (standalone)
func (l *LiquidityStage) CanRun() bool {
	// Check if CSV files exist in reports/daily directory
	reportsDir := AbsoluteReportsPath(l.executableDir, "")
	dailyDir := filepath.Join(reportsDir, "daily")
	csvFiles, err := filepath.Glob(filepath.Join(dailyDir, "*.csv"))
	if err != nil {
		if l.logger != nil {
			l.logger.Warn("Failed to check for CSV files in standalone mode",
				slog.String("directory", reportsDir),
				slog.String("error", err.Error()))
		}
		return false
	}

	hasCSVFiles := len(csvFiles) > 0
	if l.logger != nil {
		l.logger.Info("Liquidity stage standalone validation",
			slog.String("directory", reportsDir),
			slog.Int("csv_files_found", len(csvFiles)),
			slog.Bool("can_run", hasCSVFiles))
	}

	return hasCSVFiles
}

// RequiredInputs returns the data requirements for this stage
func (l *LiquidityStage) RequiredInputs() []DataRequirement {
	return []DataRequirement{
		{
			Type:     DataTypeCSVFiles,
			Required: true,
			MinCount: 1, // At least one CSV file is required
		},
	}
}

// ProducedOutputs returns the data outputs this stage produces
func (l *LiquidityStage) ProducedOutputs() []DataOutput {
	reportsDir := AbsoluteReportsPath(l.executableDir, "")
	return []DataOutput{
		{
			Type:        DataTypeLiquidityMetrics,
			Location:    reportsDir,
			FilePattern: "liquidity_metrics.csv",
			Required:    true,
			MinCount:    1, // Should produce exactly one liquidity metrics file
		},
		{
			Type:        DataTypeLiquidityScores,
			Location:    filepath.Join(reportsDir, "liquidity_reports"),
			FilePattern: "liquidity_scores_*.csv",
			Required:    true,
			MinCount:    1,
		},
	}
}

// Validate checks if the stage can be executed with the current state
func (l *LiquidityStage) Validate(state *OperationState) error {
	// Check if reports directory exists
	reportsDir := AbsoluteReportsPath(l.executableDir, "")
	if _, err := os.Stat(reportsDir); err != nil {
		return fmt.Errorf("reports directory not found: %w", err)
	}

	// Check if CSV files exist from processing stage (in daily subdirectory)
	dailyDir := filepath.Join(reportsDir, "daily")
	csvFiles, err := filepath.Glob(filepath.Join(dailyDir, "*.csv"))
	if err != nil {
		return fmt.Errorf("failed to check for CSV files: %w", err)
	}

	if len(csvFiles) == 0 {
		return fmt.Errorf("no CSV files found in reports directory - indices stage may not have completed successfully")
	}

	return nil
}

// Helper functions are now consolidated in stage_broadcaster.go to avoid duplication
