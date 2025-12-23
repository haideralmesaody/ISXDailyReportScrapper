package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LiquidityService handles liquidity-related operations
type LiquidityService struct {
	dataDir string
	logger  *slog.Logger
}

// NewLiquidityService creates a new liquidity service
func NewLiquidityService(dataDir string, logger *slog.Logger) *LiquidityService {
	return &LiquidityService{
		dataDir: dataDir,
		logger:  logger,
	}
}

// TradingThreshold represents safe trading sizes
type TradingThreshold struct {
	Conservative float64 `json:"conservative"`
	Moderate     float64 `json:"moderate"`
	Aggressive   float64 `json:"aggressive"`
	Optimal      float64 `json:"optimal"`
}

// StockMetrics holds calculated values for a single scoring mode
type StockMetrics struct {
	Score        float64          `json:"score"`
	Thresholds   TradingThreshold `json:"thresholds"`
	Continuity   float64          `json:"continuity"`
	DailyVolume  float64          `json:"dailyVolume"`
}

// StockRecommendation represents a stock with liquidity analysis
type StockRecommendation struct {
	Symbol       string           `json:"symbol"`
	CompanyName  string           `json:"companyName,omitempty"` // Company name from ISX data
	Score        float64          `json:"score"`        // Best score
	EMA20Score   float64          `json:"ema20Score"`   // 20-period EMA
	LatestScore  float64          `json:"latestScore"`  // Most recent score
	Thresholds   TradingThreshold `json:"thresholds"`
	Action       string           `json:"action"`
	Rationale    string           `json:"rationale"`
	Continuity   float64          `json:"continuity"`
	DailyVolume  float64          `json:"dailyVolume"`
	DataQuality  string           `json:"dataQuality"`
	Categories   []string         `json:"categories,omitempty"` // Categories this stock belongs to
	
	// Component scores for transparency (3-metric system)
	ILLIQScore      float64 `json:"illiqScore,omitempty"`      // Price impact score (0-100)
	VolumeScore     float64 `json:"volumeScore,omitempty"`     // Trading volume score (0-100)
	ContinuityScore float64 `json:"continuityScore,omitempty"` // Continuity score (0-100)
	ILLIQRaw        float64 `json:"illiqRaw,omitempty"`        // Raw ILLIQ value
	VolumeRaw       float64 `json:"volumeRaw,omitempty"`       // Raw volume in IQD
	
	// New scoring mode fields (backward compatible - omitempty)
	EMAMetrics     *StockMetrics `json:"emaMetrics,omitempty"`     // EMA-based metrics
	LatestMetrics  *StockMetrics `json:"latestMetrics,omitempty"`  // Most recent metrics
	AverageMetrics *StockMetrics `json:"averageMetrics,omitempty"` // Average metrics
	ActiveMode     string        `json:"activeMode,omitempty"`      // Current display mode
}

// LiquidityInsights represents the complete liquidity analysis
type LiquidityInsights struct {
	GeneratedAt        time.Time              `json:"generatedAt"`
	MarketHealthScore  float64                `json:"marketHealthScore"`
	TotalStocks        int                    `json:"totalStocks"`
	HighQualityStocks  int                    `json:"highQualityStocks"`
	AverageContinuity  float64                `json:"averageContinuity"`
	MedianDailyVolume  float64                `json:"medianDailyVolume"`
	
	// SSOT: Single master list of all stocks
	AllStocks          []StockRecommendation  `json:"allStocks,omitempty"`
	
	// Category references (ticker symbols only for SSOT, full objects for legacy)
	TopOpportunities   interface{}            `json:"topOpportunities"`
	BestForLargeTrades interface{}            `json:"bestForLargeTrades"`
	BestForDayTrading  interface{}            `json:"bestForDayTrading"`
	HighRisk          interface{}            `json:"highRisk"`
}

// Helper functions for SSOT implementation

// mergeStockData merges duplicate stock entries, combining their categories
func mergeStockData(existing, new StockRecommendation) StockRecommendation {
	// Keep the better scores
	if new.Score > existing.Score {
		existing.Score = new.Score
	}
	if new.EMA20Score > existing.EMA20Score {
		existing.EMA20Score = new.EMA20Score
	}
	if new.LatestScore > existing.LatestScore {
		existing.LatestScore = new.LatestScore
	}
	
	// Merge categories
	categoryMap := make(map[string]bool)
	for _, cat := range existing.Categories {
		categoryMap[cat] = true
	}
	for _, cat := range new.Categories {
		categoryMap[cat] = true
	}
	
	existing.Categories = make([]string, 0, len(categoryMap))
	for cat := range categoryMap {
		existing.Categories = append(existing.Categories, cat)
	}
	
	// Keep better thresholds
	if new.Thresholds.Optimal > existing.Thresholds.Optimal {
		existing.Thresholds = new.Thresholds
	}
	
	// Handle continuity - keep non-zero value or average if both are non-zero
	if existing.Continuity > 0 && new.Continuity > 0 {
		// Both have continuity values, average them
		existing.Continuity = (existing.Continuity + new.Continuity) / 2
	} else if new.Continuity > 0 {
		// Use the new non-zero continuity value
		existing.Continuity = new.Continuity
	}
	// If only existing has continuity > 0, keep it as is
	
	// Handle daily volume similarly
	if existing.DailyVolume > 0 && new.DailyVolume > 0 {
		// Both have volume values, average them
		existing.DailyVolume = (existing.DailyVolume + new.DailyVolume) / 2
	} else if new.DailyVolume > 0 {
		// Use the new non-zero volume value
		existing.DailyVolume = new.DailyVolume
	}
	// If only existing has volume > 0, keep it as is
	
	return existing
}

// hasCategory checks if a stock has a specific category
func hasCategory(stock StockRecommendation, category string) bool {
	for _, cat := range stock.Categories {
		if cat == category {
			return true
		}
	}
	return false
}

// findStock finds a stock by symbol in the AllStocks list
func findStock(stocks []StockRecommendation, symbol string) (*StockRecommendation, int) {
	for i, stock := range stocks {
		if stock.Symbol == symbol {
			return &stocks[i], i
		}
	}
	return nil, -1
}

// GetLatestInsights returns the latest liquidity insights using the specified scoring mode
func (s *LiquidityService) GetLatestInsights(ctx context.Context, mode string) (*LiquidityInsights, error) {
	// Always use liquidity scores which has complete component data
	// The insights file is just a summary without component scores
	s.logger.Info("Using liquidity scores for complete component data",
		slog.String("dataDir", s.dataDir),
		slog.String("mode", mode))
	return s.parseFromLiquidityScores(ctx, mode)
}

// parseInsightsFile parses an insights CSV file
func (s *LiquidityService) parseInsightsFile(ctx context.Context, filePath string) (*LiquidityInsights, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open insights file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	// Set FieldsPerRecord to -1 to allow variable number of fields
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	
	// Initialize with SSOT structure
	insights := &LiquidityInsights{
		GeneratedAt:        time.Now(),
		AllStocks:          []StockRecommendation{},
		TopOpportunities:   []string{},
		BestForLargeTrades: []string{},
		BestForDayTrading:  []string{},
		HighRisk:          []string{},
	}
	
	// Map to track stocks by symbol for deduplication
	stockMap := make(map[string]*StockRecommendation)
	
	// Parse the CSV structure
	section := ""
	for i, row := range records {
		if len(row) == 0 {
			continue
		}
		
		// Check for metadata rows
		if row[0] == "Generated:" && len(row) > 1 {
			if t, err := time.Parse("2006-01-02 15:04:05", row[1]); err == nil {
				insights.GeneratedAt = t
			}
		} else if row[0] == "Market Health Score:" && len(row) > 1 {
			insights.MarketHealthScore, _ = strconv.ParseFloat(row[1], 64)
		} else if row[0] == "Total Stocks Analyzed:" && len(row) > 1 {
			insights.TotalStocks, _ = strconv.Atoi(row[1])
		} else if row[0] == "High Quality Stocks:" && len(row) > 1 {
			insights.HighQualityStocks, _ = strconv.Atoi(row[1])
		} else if row[0] == "Average Continuity:" && len(row) > 1 {
			continuityStr := strings.TrimSuffix(row[1], "%")
			insights.AverageContinuity, _ = strconv.ParseFloat(continuityStr, 64)
		} else if row[0] == "Median Daily Volume:" && len(row) > 1 {
			volumeStr := strings.TrimSuffix(row[1], " IQD")
			insights.MedianDailyVolume, _ = strconv.ParseFloat(volumeStr, 64)
		}
		
		// Check for section headers
		if row[0] == "TOP TRADING OPPORTUNITIES" {
			section = "opportunities"
			continue
		} else if row[0] == "BEST FOR LARGE TRADES" {
			section = "large"
			continue
		} else if row[0] == "BEST FOR DAY TRADING" {
			section = "daytrading"
			continue
		} else if row[0] == "HIGH RISK - AVOID" {
			section = "risk"
			continue
		} else if row[0] == "TRADING GUIDELINES" {
			section = ""
			continue
		}
		
		// Skip header rows
		if i > 0 && (row[0] == "Symbol" || strings.HasPrefix(row[0], "Conservative")) {
			continue
		}
		
		// Parse data rows based on section
		if section != "" && len(row) >= 4 {
			stock := s.parseStockRow(row, section)
			if stock != nil {
				// Add category to stock
				stock.Categories = []string{section}
				
				// Check if stock already exists in map
				if existing, exists := stockMap[stock.Symbol]; exists {
					// Merge with existing stock
					*existing = mergeStockData(*existing, *stock)
				} else {
					// Add new stock to map
					stockMap[stock.Symbol] = stock
				}
				
				// Add ticker reference to appropriate category
				switch section {
				case "opportunities":
					insights.TopOpportunities = append(insights.TopOpportunities.([]string), stock.Symbol)
				case "large":
					insights.BestForLargeTrades = append(insights.BestForLargeTrades.([]string), stock.Symbol)
				case "daytrading":
					insights.BestForDayTrading = append(insights.BestForDayTrading.([]string), stock.Symbol)
				case "risk":
					insights.HighRisk = append(insights.HighRisk.([]string), stock.Symbol)
				}
			}
		}
	}
	
	// Convert map to AllStocks slice
	for _, stock := range stockMap {
		insights.AllStocks = append(insights.AllStocks, *stock)
	}
	
	// Sort AllStocks by score for consistency
	sort.Slice(insights.AllStocks, func(i, j int) bool {
		// Primary sort by EMA20
		if insights.AllStocks[i].EMA20Score != insights.AllStocks[j].EMA20Score {
			return insights.AllStocks[i].EMA20Score > insights.AllStocks[j].EMA20Score
		}
		// Secondary sort by best score
		return insights.AllStocks[i].Score > insights.AllStocks[j].Score
	})
	
	return insights, nil
}

// parseStockRow parses a single stock row from the CSV
func (s *LiquidityService) parseStockRow(row []string, section string) *StockRecommendation {
	if len(row) < 4 {
		return nil
	}
	
	stock := &StockRecommendation{
		Symbol: row[0],
	}
	
	// Parse based on section
	switch section {
	case "opportunities":
		if len(row) >= 12 { // Updated for new columns
			stock.EMA20Score, _ = strconv.ParseFloat(row[1], 64)
			stock.LatestScore, _ = strconv.ParseFloat(row[2], 64)
			stock.Score, _ = strconv.ParseFloat(row[3], 64) // Best score
			stock.Thresholds.Conservative, _ = strconv.ParseFloat(row[4], 64)
			stock.Thresholds.Moderate, _ = strconv.ParseFloat(row[5], 64)
			stock.Thresholds.Aggressive, _ = strconv.ParseFloat(row[6], 64)
			stock.Thresholds.Optimal, _ = strconv.ParseFloat(row[7], 64)
			// Parse continuity (remove % sign)
			continuityStr := strings.TrimSuffix(row[8], "%")
			stock.Continuity, _ = strconv.ParseFloat(continuityStr, 64)
			stock.Continuity /= 100 // Convert to decimal
			stock.DailyVolume, _ = strconv.ParseFloat(row[9], 64)
			stock.Action = row[10]
			stock.Rationale = row[11]
			stock.DataQuality = "GOOD" // Default for top opportunities
		}
	case "large":
		if len(row) >= 5 {
			stock.Score, _ = strconv.ParseFloat(row[1], 64)
			stock.Thresholds.Optimal, _ = strconv.ParseFloat(row[2], 64)
			stock.DailyVolume, _ = strconv.ParseFloat(row[3], 64)
			stock.Rationale = row[4]
			stock.Action = "BUY_LARGE"
			stock.DataQuality = "GOOD"
		}
	case "daytrading":
		if len(row) >= 5 {
			stock.Score, _ = strconv.ParseFloat(row[1], 64)
			continuityStr := strings.TrimSuffix(row[2], "%")
			stock.Continuity, _ = strconv.ParseFloat(continuityStr, 64)
			stock.Continuity /= 100 // Convert to decimal
			stock.Rationale = row[4]
			stock.Action = "DAY_TRADE"
			stock.DataQuality = "GOOD"
		}
	case "risk":
		if len(row) >= 5 {
			stock.Score, _ = strconv.ParseFloat(row[1], 64)
			stock.DataQuality = row[2]
			stock.Rationale = row[4]
			stock.Action = "AVOID"
		}
	}
	
	return stock
}

// loadCompanyNames loads company names from the combined CSV file
func (s *LiquidityService) loadCompanyNames() map[string]string {
	companyNames := make(map[string]string)

	// Try to find the combined data file
	combinedDir := filepath.Join(s.dataDir, "combined")
	combinedFile := filepath.Join(combinedDir, "isx_combined_data.csv")

	file, err := os.Open(combinedFile)
	if err != nil {
		s.logger.Warn("Could not load company names from combined data",
			slog.String("file", combinedFile),
			slog.String("error", err.Error()))
		return companyNames
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		s.logger.Warn("Could not parse company names from combined data")
		return companyNames
	}

	// Find column indices
	header := records[0]
	var symbolIdx, companyNameIdx int = -1, -1
	for i, col := range header {
		if col == "Symbol" {
			symbolIdx = i
		} else if col == "CompanyName" {
			companyNameIdx = i
		}
	}

	if symbolIdx == -1 || companyNameIdx == -1 {
		s.logger.Warn("Could not find Symbol or CompanyName columns in combined data")
		return companyNames
	}

	// Build map
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) > companyNameIdx && len(row) > symbolIdx {
			symbol := row[symbolIdx]
			companyName := row[companyNameIdx]
			if symbol != "" && companyName != "" {
				companyNames[symbol] = companyName
			}
		}
	}

	s.logger.Info("Loaded company names",
		slog.Int("count", len(companyNames)))

	return companyNames
}

// parseFromLiquidityScores parses directly from liquidity scores with mode-specific recommendations
func (s *LiquidityService) parseFromLiquidityScores(ctx context.Context, mode string) (*LiquidityInsights, error) {
	// Load company names from combined data
	companyNames := s.loadCompanyNames()

	// Find the most recent liquidity scores file in the new liquidity_reports subdirectory
	liquidityReportsDir := filepath.Join(s.dataDir, "liquidity_reports")
	pattern := filepath.Join(liquidityReportsDir, "liquidity_scores_*.csv")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob liquidity files: %w", err)
	}

	// Fallback to old location if no files found in new location
	if len(files) == 0 {
		// Try old location for backward compatibility
		oldPattern := filepath.Join(s.dataDir, "liquidity_scores_*.csv")
		files, err = filepath.Glob(oldPattern)
		if err != nil {
			return nil, fmt.Errorf("glob liquidity files: %w", err)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no liquidity data available")
	}

	// Sort to get the most recent
	sort.Strings(files)
	latestFile := files[len(files)-1]
	
	// Parse the liquidity scores CSV
	file, err := os.Open(latestFile)
	if err != nil {
		return nil, fmt.Errorf("open liquidity file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	// Allow flexible CSV format
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	
	if len(records) < 2 {
		return nil, fmt.Errorf("insufficient data in liquidity file")
	}
	
	// Find column indices
	header := records[0]
	indices := make(map[string]int)
	for i, col := range header {
		indices[col] = i
	}
	
	// Group entries by symbol for aggregation
	symbolData := make(map[string][]StockRecommendation)
	
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < len(header) {
			continue
		}
		
		symbol := row[indices["Symbol"]]
		stock := StockRecommendation{
			Symbol:      symbol,
			CompanyName: companyNames[symbol], // Lookup company name
			DataQuality: "GOOD",
		}

		// Parse numeric values
		if idx, ok := indices["Hybrid_Score"]; ok {
			stock.Score, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Safe_Trade_0.5%"]; ok {
			stock.Thresholds.Conservative, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Safe_Trade_1%"]; ok {
			stock.Thresholds.Moderate, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Safe_Trade_2%"]; ok {
			stock.Thresholds.Aggressive, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Optimal_Trade"]; ok {
			stock.Thresholds.Optimal, _ = strconv.ParseFloat(row[idx], 64)
		}
		// Use Continuity_Raw which is already in decimal format (0.9 = 90%)
		if idx, ok := indices["Continuity_Raw"]; ok {
			stock.Continuity, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Value_Raw"]; ok {
			stock.DailyVolume, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Data_Quality"]; ok {
			stock.DataQuality = row[idx]
		}
		
		// Parse component scores for transparency
		if idx, ok := indices["ILLIQ_Scaled"]; ok {
			stock.ILLIQScore, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Value_Scaled"]; ok {
			stock.VolumeScore, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Continuity_Scaled"]; ok {
			stock.ContinuityScore, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["ILLIQ_Raw"]; ok {
			stock.ILLIQRaw, _ = strconv.ParseFloat(row[idx], 64)
		}
		if idx, ok := indices["Value_Raw"]; ok {
			stock.VolumeRaw, _ = strconv.ParseFloat(row[idx], 64)
		}
		
		// Collect all entries for this symbol
		symbolData[stock.Symbol] = append(symbolData[stock.Symbol], stock)
	}
	
	// Aggregate data for each symbol
	var stocks []StockRecommendation
	for _, entries := range symbolData {
		// Calculate EMA20 and aggregated values
		aggregated := s.aggregateStockData(entries)

		// Generate action and rationale based on mode-specific score
		aggregated.Action, aggregated.Rationale = s.generateRecommendation(aggregated, mode)

		stocks = append(stocks, aggregated)
	}
	
	// Sort by EMA20 score as primary, best score as secondary
	sort.Slice(stocks, func(i, j int) bool {
		// Primary sort by EMA20
		if stocks[i].EMA20Score != stocks[j].EMA20Score {
			return stocks[i].EMA20Score > stocks[j].EMA20Score
		}
		// Secondary sort by best score
		return stocks[i].Score > stocks[j].Score
	})
	
	// Build insights with SSOT structure
	insights := &LiquidityInsights{
		GeneratedAt:       time.Now(),
		TotalStocks:       len(stocks),
		AllStocks:         []StockRecommendation{},
		TopOpportunities:  []string{},
		BestForLargeTrades: []string{},
		BestForDayTrading: []string{},
		HighRisk:         []string{},
	}
	
	// Calculate market health (average of top 20)
	topN := 20
	if len(stocks) < topN {
		topN = len(stocks)
	}
	
	var totalScore, totalContinuity float64
	var volumes []float64
	highQualityCount := 0
	
	for i, stock := range stocks {
		if i < topN {
			totalScore += stock.Score
		}
		
		if stock.DataQuality != "POOR" {
			highQualityCount++
			totalContinuity += stock.Continuity
			volumes = append(volumes, stock.DailyVolume)
		}
		
		// Build categories for each stock
		stockCopy := stock
		stockCopy.Categories = []string{}
		
		// Categorize stocks
		if i < 10 && stock.Score >= 50 {
			stockCopy.Categories = append(stockCopy.Categories, "opportunities")
			insights.TopOpportunities = append(insights.TopOpportunities.([]string), stock.Symbol)
		}
		
		if stock.DataQuality == "POOR" || stock.Score < 30 {
			if len(insights.HighRisk.([]string)) < 10 {
				stockCopy.Categories = append(stockCopy.Categories, "risk")
				insights.HighRisk = append(insights.HighRisk.([]string), stock.Symbol)
			}
		}
		
		if stock.Thresholds.Optimal >= 5_000_000 && stock.DataQuality != "POOR" {
			if len(insights.BestForLargeTrades.([]string)) < 5 {
				stockCopy.Categories = append(stockCopy.Categories, "large")
				insights.BestForLargeTrades = append(insights.BestForLargeTrades.([]string), stock.Symbol)
			}
		}
		
		// Use EMA metrics for day trading categorization to fix the empty category issue
		// Check if EMA metrics exist, otherwise fall back to regular metrics
		dayTradingScore := stock.Score
		dayTradingContinuity := stock.Continuity
		
		if stock.EMAMetrics != nil {
			dayTradingScore = stock.EMAMetrics.Score
			dayTradingContinuity = stock.EMAMetrics.Continuity
		}
		
		if dayTradingContinuity >= 0.7 && dayTradingScore >= 50 {
			if len(insights.BestForDayTrading.([]string)) < 5 {
				stockCopy.Categories = append(stockCopy.Categories, "daytrading")
				insights.BestForDayTrading = append(insights.BestForDayTrading.([]string), stock.Symbol)
			}
		}
		
		// Add to AllStocks
		insights.AllStocks = append(insights.AllStocks, stockCopy)
	}
	
	insights.MarketHealthScore = totalScore / float64(topN)
	insights.HighQualityStocks = highQualityCount
	
	if highQualityCount > 0 {
		insights.AverageContinuity = (totalContinuity / float64(highQualityCount)) * 100
	}
	
	// Calculate median volume
	if len(volumes) > 0 {
		sort.Float64s(volumes)
		mid := len(volumes) / 2
		if len(volumes)%2 == 0 {
			insights.MedianDailyVolume = (volumes[mid-1] + volumes[mid]) / 2
		} else {
			insights.MedianDailyVolume = volumes[mid]
		}
	}
	
	return insights, nil
}

// calculateEMA20 calculates 20-period exponential moving average
func (s *LiquidityService) calculateEMA20(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// EMA smoothing factor for 20 periods
	alpha := 2.0 / 21.0
	ema := values[0]
	
	for i := 1; i < len(values); i++ {
		ema = alpha*values[i] + (1-alpha)*ema
	}
	
	return ema
}

// calculateEMA20WithOutlierRemoval calculates EMA after removing outliers
func (s *LiquidityService) calculateEMA20WithOutlierRemoval(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Remove outliers first
	cleanedValues := s.removeOutliers(values)
	if len(cleanedValues) == 0 {
		return 0
	}
	
	// Calculate EMA on cleaned data
	alpha := 2.0 / 21.0
	ema := cleanedValues[0]
	
	for i := 1; i < len(cleanedValues); i++ {
		ema = alpha*cleanedValues[i] + (1-alpha)*ema
	}
	
	return ema
}

// removeOutliers removes statistical outliers using IQR method
func (s *LiquidityService) removeOutliers(values []float64) []float64 {
	if len(values) < 4 {
		return values // Not enough data for IQR
	}
	
	// Create a sorted copy
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	
	// Calculate Q1, Q3, and IQR
	q1Index := len(sorted) / 4
	q3Index := 3 * len(sorted) / 4
	q1 := sorted[q1Index]
	q3 := sorted[q3Index]
	iqr := q3 - q1
	
	// Calculate bounds
	lowerBound := q1 - 1.5*iqr
	upperBound := q3 + 1.5*iqr
	
	// Filter out outliers
	var cleaned []float64
	for _, v := range values {
		if v >= lowerBound && v <= upperBound {
			cleaned = append(cleaned, v)
		}
	}
	
	// If we removed everything, return original (safety check)
	if len(cleaned) == 0 {
		return values
	}
	
	return cleaned
}

// zeroOutPoorQualityMetrics sets thresholds to zero for POOR quality data
func (s *LiquidityService) zeroOutPoorQualityMetrics(metrics *StockMetrics) {
	if metrics == nil {
		return
	}
	
	// Zero out all trading thresholds for POOR quality
	metrics.Thresholds.Conservative = 0
	metrics.Thresholds.Moderate = 0
	metrics.Thresholds.Aggressive = 0
	metrics.Thresholds.Optimal = 0
}

// calculateEMAThresholds calculates EMA-based thresholds
func (s *LiquidityService) calculateEMAThresholds(entries []StockRecommendation) TradingThreshold {
	if len(entries) == 0 {
		return TradingThreshold{}
	}
	
	conservative := make([]float64, 0, len(entries))
	moderate := make([]float64, 0, len(entries))
	aggressive := make([]float64, 0, len(entries))
	optimal := make([]float64, 0, len(entries))
	
	for _, e := range entries {
		conservative = append(conservative, e.Thresholds.Conservative)
		moderate = append(moderate, e.Thresholds.Moderate)
		aggressive = append(aggressive, e.Thresholds.Aggressive)
		optimal = append(optimal, e.Thresholds.Optimal)
	}
	
	return TradingThreshold{
		Conservative: s.calculateEMA20WithOutlierRemoval(conservative),
		Moderate:     s.calculateEMA20WithOutlierRemoval(moderate),
		Aggressive:   s.calculateEMA20WithOutlierRemoval(aggressive),
		Optimal:      s.calculateEMA20WithOutlierRemoval(optimal),
	}
}

// calculateAverageThresholds calculates average thresholds
func (s *LiquidityService) calculateAverageThresholds(entries []StockRecommendation) TradingThreshold {
	if len(entries) == 0 {
		return TradingThreshold{}
	}
	
	var result TradingThreshold
	for _, e := range entries {
		result.Conservative += e.Thresholds.Conservative
		result.Moderate += e.Thresholds.Moderate
		result.Aggressive += e.Thresholds.Aggressive
		result.Optimal += e.Thresholds.Optimal
	}
	
	n := float64(len(entries))
	result.Conservative /= n
	result.Moderate /= n
	result.Aggressive /= n
	result.Optimal /= n
	
	return result
}

// aggregateStockData aggregates multiple entries for a single symbol
func (s *LiquidityService) aggregateStockData(entries []StockRecommendation) StockRecommendation {
	// Use the new enhanced method internally
	return s.aggregateStockDataWithModes(entries)
}

// aggregateStockDataWithModes calculates metrics for all scoring modes
func (s *LiquidityService) aggregateStockDataWithModes(entries []StockRecommendation) StockRecommendation {
	if len(entries) == 0 {
		return StockRecommendation{}
	}
	
	// Collect time series data
	scores := make([]float64, 0, len(entries))
	volumes := make([]float64, 0, len(entries))
	continuities := make([]float64, 0, len(entries))
	totalVolume := 0.0
	totalContinuity := 0.0
	bestIdx := 0
	
	for i, e := range entries {
		scores = append(scores, e.Score)
		volumes = append(volumes, e.DailyVolume)
		continuities = append(continuities, e.Continuity)
		totalVolume += e.DailyVolume
		totalContinuity += e.Continuity
		
		// Track best score for legacy compatibility only
		if e.Score > entries[bestIdx].Score {
			bestIdx = i
		}
	}
	
	// Start with the best scoring entry (for backward compatibility)
	result := entries[bestIdx]
	
	// Keep legacy fields for backward compatibility but use mode-appropriate values
	result.EMA20Score = s.calculateEMA20WithOutlierRemoval(scores)
	result.LatestScore = entries[len(entries)-1].Score
	result.Score = s.calculateAverage(scores) // Use average instead of cherry-picked best
	
	// Average the volume and continuity across all dates
	if len(entries) > 0 {
		result.DailyVolume = totalVolume / float64(len(entries))
		result.Continuity = totalContinuity / float64(len(entries))
	}
	
	// Calculate EMA metrics (with outlier removal)
	result.EMAMetrics = &StockMetrics{
		Score:       s.calculateEMA20WithOutlierRemoval(scores),
		Thresholds:  s.calculateEMAThresholds(entries),
		Continuity:  s.calculateEMA20WithOutlierRemoval(continuities),
		DailyVolume: s.calculateEMA20WithOutlierRemoval(volumes),
	}
	
	// Latest metrics (most recent entry)
	lastEntry := entries[len(entries)-1]
	result.LatestMetrics = &StockMetrics{
		Score:       lastEntry.Score,
		Thresholds:  lastEntry.Thresholds,
		Continuity:  lastEntry.Continuity,
		DailyVolume: lastEntry.DailyVolume,
	}
	
	// Average metrics
	result.AverageMetrics = &StockMetrics{
		Score:       s.calculateAverage(scores),
		Thresholds:  s.calculateAverageThresholds(entries),
		Continuity:  totalContinuity / float64(len(entries)),
		DailyVolume: totalVolume / float64(len(entries)),
	}
	
	// Handle POOR quality data - zero out thresholds
	if result.DataQuality == "POOR" {
		s.zeroOutPoorQualityMetrics(result.EMAMetrics)
		s.zeroOutPoorQualityMetrics(result.LatestMetrics)
		s.zeroOutPoorQualityMetrics(result.AverageMetrics)
		// Also zero out the legacy thresholds
		result.Thresholds = TradingThreshold{}
	}
	
	return result
}

// calculateAverage calculates simple average of values
func (s *LiquidityService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// generateRecommendation creates action and rationale based on mode-specific metrics
func (s *LiquidityService) generateRecommendation(stock StockRecommendation, mode string) (string, string) {
	if stock.DataQuality == "POOR" {
		return "AVOID", fmt.Sprintf("Insufficient trading data, unreliable metrics")
	}

	// Select score and metrics based on mode
	var effectiveScore float64
	var effectiveThresholds TradingThreshold
	var effectiveContinuity float64
	var modeLabel string

	switch mode {
	case "ema":
		if stock.EMAMetrics != nil {
			effectiveScore = stock.EMAMetrics.Score
			effectiveThresholds = stock.EMAMetrics.Thresholds
			effectiveContinuity = stock.EMAMetrics.Continuity
		} else {
			// Fallback to legacy fields
			effectiveScore = stock.EMA20Score
			effectiveThresholds = stock.Thresholds
			effectiveContinuity = stock.Continuity
		}
		modeLabel = "EMA20"
	case "latest":
		if stock.LatestMetrics != nil {
			effectiveScore = stock.LatestMetrics.Score
			effectiveThresholds = stock.LatestMetrics.Thresholds
			effectiveContinuity = stock.LatestMetrics.Continuity
		} else {
			// Fallback to legacy fields
			effectiveScore = stock.LatestScore
			effectiveThresholds = stock.Thresholds
			effectiveContinuity = stock.Continuity
		}
		modeLabel = "Latest"
	case "average":
		if stock.AverageMetrics != nil {
			effectiveScore = stock.AverageMetrics.Score
			effectiveThresholds = stock.AverageMetrics.Thresholds
			effectiveContinuity = stock.AverageMetrics.Continuity
		} else {
			// Fallback to legacy fields
			effectiveScore = stock.Score
			effectiveThresholds = stock.Thresholds
			effectiveContinuity = stock.Continuity
		}
		modeLabel = "Average"
	default:
		// Default to average
		effectiveScore = stock.Score
		effectiveThresholds = stock.Thresholds
		effectiveContinuity = stock.Continuity
		modeLabel = "Average"
	}

	// Generate recommendations using mode-specific values
	if effectiveScore >= 85 {
		if effectiveThresholds.Optimal >= 10_000_000 {
			return "BUY_LARGE", fmt.Sprintf("Excellent liquidity (%s: %.1f), supports large trades up to %.0fM IQD",
				modeLabel, effectiveScore, effectiveThresholds.Optimal/1_000_000)
		}
		return "BUY", fmt.Sprintf("High liquidity (%s: %.1f), good for active trading", modeLabel, effectiveScore)
	}

	if effectiveContinuity >= 0.7 && effectiveScore >= 60 {
		return "DAY_TRADE", fmt.Sprintf("High continuity (%.0f%%), suitable for frequent trading (%s: %.1f)",
			effectiveContinuity*100, modeLabel, effectiveScore)
	}

	if effectiveScore >= 50 {
		return "HOLD", fmt.Sprintf("Moderate liquidity (%s: %.1f), trade carefully within %.0fK IQD",
			modeLabel, effectiveScore, effectiveThresholds.Optimal/1000)
	}

	if effectiveScore >= 30 {
		return "CAUTION", fmt.Sprintf("Low liquidity (%s: %.1f), limit trades to %.0fK IQD",
			modeLabel, effectiveScore, effectiveThresholds.Conservative/1000)
	}

	return "AVOID", fmt.Sprintf("Very low liquidity (%s: %.1f), high price impact risk", modeLabel, effectiveScore)
}