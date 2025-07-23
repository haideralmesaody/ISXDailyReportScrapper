package analytics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	
	"isxcli/internal/exporter"
)

// SummaryGenerator handles ticker summary generation
type SummaryGenerator struct {
	csvWriter *exporter.CSVWriter
}

// NewSummaryGenerator creates a new summary generator
func NewSummaryGenerator(basePath string) *SummaryGenerator {
	return &SummaryGenerator{
		csvWriter: exporter.NewCSVWriter(basePath),
	}
}

// TickerInfo represents summary information for a ticker
type TickerInfo struct {
	Ticker      string
	CompanyName string
	LastPrice   float64
	LastDate    string
	TradingDays int
	Last10Days  []float64
	
	// New fields for gainers/losers dashboard
	DailyChangePercent   float64 `json:"daily_change_percent"`
	WeeklyChangePercent  float64 `json:"weekly_change_percent"`
	MonthlyChangePercent float64 `json:"monthly_change_percent"`
	DailyVolume         int64   `json:"daily_volume"`
	DailyValue          float64 `json:"daily_value"`
	PreviousClose       float64 `json:"previous_close"`
	High52Week          float64 `json:"high_52_week"`
	Low52Week           float64 `json:"low_52_week"`
}

// GenerateFromCombinedCSV generates ticker summary from the combined CSV file
// It creates both CSV and JSON formats for compatibility with the web interface
func (s *SummaryGenerator) GenerateFromCombinedCSV(combinedFile, summaryFile string) error {
	// Check if combined file exists
	if _, err := os.Stat(combinedFile); os.IsNotExist(err) {
		return fmt.Errorf("combined CSV file not found: %s", combinedFile)
	}

	// Read combined CSV
	file, err := os.Open(combinedFile)
	if err != nil {
		return fmt.Errorf("failed to open combined file: %v", err)
	}
	defer file.Close()

	// Read file content to handle BOM
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file content: %v", err)
	}
	
	// Remove BOM if present
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		content = content[3:]
	}
	
	// Create CSV reader from cleaned content
	reader := csv.NewReader(strings.NewReader(string(content)))
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read combined CSV: %v", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("combined CSV has no data rows")
	}

	// Parse data and generate summaries
	summaries, err := s.parseAndGenerateSummaries(records)
	if err != nil {
		return err
	}

	// Write ticker summary CSV
	if err := s.writeSummaryCSV(summaryFile, summaries); err != nil {
		return err
	}
	
	// Also write JSON format for web interface
	jsonFile := strings.TrimSuffix(summaryFile, ".csv") + ".json"
	return s.writeSummaryJSON(jsonFile, summaries)
}

// parseAndGenerateSummaries parses CSV records and generates ticker summaries
func (s *SummaryGenerator) parseAndGenerateSummaries(records [][]string) ([]TickerInfo, error) {
	// Parse header to find column indices
	header := records[0]
	columns := s.findColumnIndices(header)
	
	// Log found columns for debugging
	fmt.Printf("[INFO] Found columns - Ticker: %d, Company: %d, Date: %d, Close: %d, Volume: %d, NumTrades: %d\n",
		columns.tickerCol, columns.companyCol, columns.dateCol, columns.closeCol, columns.volumeCol, columns.numTradesCol)
	
	if columns.tickerCol == -1 || columns.companyCol == -1 || 
	   columns.dateCol == -1 || columns.closeCol == -1 {
		// Debug: show which columns were not found
		missing := []string{}
		if columns.tickerCol == -1 {
			missing = append(missing, "Symbol/Ticker")
		}
		if columns.companyCol == -1 {
			missing = append(missing, "CompanyName")
		}
		if columns.dateCol == -1 {
			missing = append(missing, "Date")
		}
		if columns.closeCol == -1 {
			missing = append(missing, "ClosePrice")
		}
		return nil, fmt.Errorf("required columns not found: %v. Header: %v", missing, header)
	}

	// Group data by ticker
	tickerData := s.groupByTicker(records[1:], columns)

	// Create ticker summaries
	return s.createSummaries(tickerData), nil
}

// columnIndices holds the indices of required columns
type columnIndices struct {
	tickerCol    int
	companyCol   int
	dateCol      int
	closeCol     int
	volumeCol    int
	numTradesCol int
	highCol      int
	lowCol       int
	valueCol     int
}

// findColumnIndices finds the indices of required columns in the header
func (s *SummaryGenerator) findColumnIndices(header []string) columnIndices {
	indices := columnIndices{
		tickerCol:    -1,
		companyCol:   -1,
		dateCol:      -1,
		closeCol:     -1,
		volumeCol:    -1,
		numTradesCol: -1,
		highCol:      -1,
		lowCol:       -1,
		valueCol:     -1,
	}

	for i, col := range header {
		// Clean BOM and normalize column name
		// Handle UTF-8 BOM and other invisible characters
		cleanCol := strings.TrimSpace(col)
		
		// Remove BOM if present (UTF-8 BOM is \ufeff)
		cleanCol = strings.TrimPrefix(cleanCol, "\ufeff")
		
		// Also handle case where BOM appears as bytes in string
		if strings.HasPrefix(cleanCol, string([]byte{0xEF, 0xBB, 0xBF})) {
			cleanCol = cleanCol[3:]
		}
		
		// Remove any other zero-width characters
		cleanCol = strings.TrimLeft(cleanCol, "\u200B\u200C\u200D\u2060\uFEFF")
		cleanCol = strings.TrimSpace(cleanCol)
		lowerCol := strings.ToLower(cleanCol)
		
		// Debug logging - log all columns for debugging
		if os.Getenv("ISX_DEBUG") == "true" {
			fmt.Printf("[DEBUG] Column %d: original='%s' (len=%d, bytes=%x), clean='%s', lower='%s'\n", 
				i, col, len(col), []byte(col), cleanCol, lowerCol)
		}
		
		// Match against the actual column name (case-sensitive first, then lowercase)
		switch cleanCol {
		case "Symbol":
			indices.tickerCol = i
		case "CompanyName":
			indices.companyCol = i
		case "Date":
			indices.dateCol = i
		case "ClosePrice":
			indices.closeCol = i
		case "Volume":
			indices.volumeCol = i
		case "NumTrades":
			indices.numTradesCol = i
		case "HighPrice":
			indices.highCol = i
		case "LowPrice":
			indices.lowCol = i
		case "Value":
			indices.valueCol = i
		default:
			// Fallback to lowercase matching
			switch lowerCol {
			case "symbol", "ticker", "company_symbol":
				indices.tickerCol = i
			case "companyname", "company_name", "company", "name":
				indices.companyCol = i
			case "date":
				indices.dateCol = i
			case "closeprice", "close_price", "close", "closingprice":
				indices.closeCol = i
			case "volume":
				indices.volumeCol = i
			case "numtrades", "num_trades":
				indices.numTradesCol = i
			case "highprice", "high_price", "high":
				indices.highCol = i
			case "lowprice", "low_price", "low":
				indices.lowCol = i
			case "value":
				indices.valueCol = i
			}
		}
	}

	return indices
}

// groupByTicker groups records by ticker symbol
func (s *SummaryGenerator) groupByTicker(records [][]string, cols columnIndices) map[string][]map[string]string {
	tickerData := make(map[string][]map[string]string)

	for _, record := range records {
		if len(record) <= cols.tickerCol || len(record) <= cols.companyCol || 
		   len(record) <= cols.dateCol || len(record) <= cols.closeCol {
			continue
		}

		ticker := strings.TrimSpace(record[cols.tickerCol])
		if ticker == "" {
			continue
		}

		rowData := map[string]string{
			"ticker":       ticker,
			"company_name": strings.TrimSpace(record[cols.companyCol]),
			"date":         strings.TrimSpace(record[cols.dateCol]),
			"close_price":  strings.TrimSpace(record[cols.closeCol]),
		}
		
		// Add volume and num_trades if available
		if cols.volumeCol != -1 && len(record) > cols.volumeCol {
			rowData["volume"] = strings.TrimSpace(record[cols.volumeCol])
		} else {
			rowData["volume"] = "0"
		}
		
		if cols.numTradesCol != -1 && len(record) > cols.numTradesCol {
			rowData["num_trades"] = strings.TrimSpace(record[cols.numTradesCol])
		} else {
			rowData["num_trades"] = "0"
		}
		
		// Add high price, low price, and value if available
		if cols.highCol != -1 && len(record) > cols.highCol {
			rowData["high_price"] = strings.TrimSpace(record[cols.highCol])
		} else {
			rowData["high_price"] = "0"
		}
		
		if cols.lowCol != -1 && len(record) > cols.lowCol {
			rowData["low_price"] = strings.TrimSpace(record[cols.lowCol])
		} else {
			rowData["low_price"] = "0"
		}
		
		if cols.valueCol != -1 && len(record) > cols.valueCol {
			rowData["value"] = strings.TrimSpace(record[cols.valueCol])
		} else {
			rowData["value"] = "0"
		}

		tickerData[ticker] = append(tickerData[ticker], rowData)
	}

	return tickerData
}

// createSummaries creates ticker summaries from grouped data
func (s *SummaryGenerator) createSummaries(tickerData map[string][]map[string]string) []TickerInfo {
	var summaries []TickerInfo

	for ticker, data := range tickerData {
		if len(data) == 0 {
			continue
		}

		// Sort by date
		sort.Slice(data, func(i, j int) bool {
			return data[i]["date"] < data[j]["date"]
		})

		lastRecord := data[len(data)-1]
		lastPrice, _ := strconv.ParseFloat(lastRecord["close_price"], 64)

		// Count actual trading days and get last 10 actual trading days
		var last10Days []float64
		actualTradingDays := 0
		
		// Go backwards from the most recent date to find actual trading days
		for i := len(data) - 1; i >= 0; i-- {
			// Check if this is an actual trading day by looking at volume or number of trades
			volume, _ := strconv.ParseInt(data[i]["volume"], 10, 64)
			numTrades, _ := strconv.ParseInt(data[i]["num_trades"], 10, 64)
			
			// Only include days with actual trading activity
			if volume > 0 || numTrades > 0 {
				actualTradingDays++
				if len(last10Days) < 10 {
					price, _ := strconv.ParseFloat(data[i]["close_price"], 64)
					last10Days = append([]float64{price}, last10Days...) // Prepend to maintain chronological order
				}
			}
		}

		// Calculate percentage changes and volume metrics
		dailyChangePercent := s.calculateDailyChangePercent(last10Days)
		weeklyChangePercent := s.calculatePeriodChangePercent(last10Days, 7)
		monthlyChangePercent := s.calculatePeriodChangePercent(last10Days, 30)
		previousClose := s.getPreviousClose(last10Days)
		high52Week, low52Week := s.calculate52WeekHighLow(data)
		dailyVolume, dailyValue := s.getDailyVolumeValue(lastRecord)

		summary := TickerInfo{
			Ticker:      ticker,
			CompanyName: lastRecord["company_name"],
			LastPrice:   lastPrice,
			LastDate:    lastRecord["date"],
			TradingDays: actualTradingDays,
			Last10Days:  last10Days,
			
			// New fields for gainers/losers dashboard
			DailyChangePercent:   dailyChangePercent,
			WeeklyChangePercent:  weeklyChangePercent,
			MonthlyChangePercent: monthlyChangePercent,
			DailyVolume:         dailyVolume,
			DailyValue:          dailyValue,
			PreviousClose:       previousClose,
			High52Week:          high52Week,
			Low52Week:           low52Week,
		}

		summaries = append(summaries, summary)
	}

	// Sort summaries by ticker
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})

	return summaries
}

// writeSummaryCSV writes ticker summaries to a CSV file
func (s *SummaryGenerator) writeSummaryCSV(filename string, summaries []TickerInfo) error {
	// Create stream writer
	stream, err := s.csvWriter.CreateStreamWriter(filename, []string{
		"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days",
	})
	if err != nil {
		return fmt.Errorf("failed to create summary file: %v", err)
	}
	defer stream.Close()

	// Write data
	for _, summary := range summaries {
		last10DaysStr := s.formatLast10Days(summary.Last10Days)
		
		if err := stream.WriteRecord([]string{
			summary.Ticker,
			summary.CompanyName,
			fmt.Sprintf("%.3f", summary.LastPrice),
			summary.LastDate,
			fmt.Sprintf("%d", summary.TradingDays),
			last10DaysStr,
		}); err != nil {
			return err
		}
	}

	return nil
}

// formatLast10Days formats the last 10 days prices as a comma-separated string
func (s *SummaryGenerator) formatLast10Days(prices []float64) string {
	parts := make([]string, len(prices))
	for i, price := range prices {
		parts[i] = fmt.Sprintf("%.3f", price)
	}
	return strings.Join(parts, ",")
}

// writeSummaryJSON writes ticker summaries to a JSON file for web interface compatibility
func (s *SummaryGenerator) writeSummaryJSON(filename string, summaries []TickerInfo) error {
	// Convert to the format expected by the web interface
	type WebTickerSummary struct {
		Ticker      string   `json:"ticker"`
		CompanyName string   `json:"company_name"`
		LastPrice   float64  `json:"last_price"`
		LastDate    string   `json:"last_date"`
		TradingDays int      `json:"trading_days"`
		Last10Days  []float64 `json:"last_10_days"`
		
		// New fields for gainers/losers dashboard
		DailyChangePercent   float64 `json:"daily_change_percent"`
		WeeklyChangePercent  float64 `json:"weekly_change_percent"`
		MonthlyChangePercent float64 `json:"monthly_change_percent"`
		DailyVolume         int64   `json:"daily_volume"`
		DailyValue          float64 `json:"daily_value"`
		PreviousClose       float64 `json:"previous_close"`
		High52Week          float64 `json:"high_52_week"`
		Low52Week           float64 `json:"low_52_week"`
	}
	
	webSummaries := make([]WebTickerSummary, len(summaries))
	for i, summary := range summaries {
		// Ensure last10Days is never nil
		last10Days := summary.Last10Days
		if last10Days == nil {
			last10Days = []float64{}
		}
		
		webSummaries[i] = WebTickerSummary{
			Ticker:      summary.Ticker,
			CompanyName: summary.CompanyName,
			LastPrice:   summary.LastPrice,
			LastDate:    summary.LastDate,
			TradingDays: summary.TradingDays,
			Last10Days:  last10Days,
			
			// New fields for gainers/losers dashboard
			DailyChangePercent:   summary.DailyChangePercent,
			WeeklyChangePercent:  summary.WeeklyChangePercent,
			MonthlyChangePercent: summary.MonthlyChangePercent,
			DailyVolume:         summary.DailyVolume,
			DailyValue:          summary.DailyValue,
			PreviousClose:       summary.PreviousClose,
			High52Week:          summary.High52Week,
			Low52Week:           summary.Low52Week,
		}
	}
	
	// Create the expected JSON structure with metadata
	jsonData := map[string]interface{}{
		"tickers":      webSummaries,
		"count":        len(webSummaries),
		"generated_at": time.Now().Format(time.RFC3339),
	}
	
	// Create the JSON file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %v", err)
	}
	defer file.Close()
	
	// Write JSON with indentation for readability
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonData); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}
	
	return nil
}

// calculateDailyChangePercent calculates the daily percentage change
func (s *SummaryGenerator) calculateDailyChangePercent(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}
	current := prices[len(prices)-1]
	previous := prices[len(prices)-2]
	if previous == 0 {
		return 0.0
	}
	return ((current - previous) / previous) * 100
}

// calculatePeriodChangePercent calculates percentage change over a period
func (s *SummaryGenerator) calculatePeriodChangePercent(prices []float64, days int) float64 {
	if len(prices) < 2 {
		return 0.0
	}
	
	current := prices[len(prices)-1]
	
	// Get price from 'days' ago, or earliest available
	pastIndex := len(prices) - 1 - days
	if pastIndex < 0 {
		pastIndex = 0
	}
	
	past := prices[pastIndex]
	if past == 0 {
		return 0.0
	}
	
	return ((current - past) / past) * 100
}

// getPreviousClose gets the previous trading day's close price
func (s *SummaryGenerator) getPreviousClose(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.0
	}
	return prices[len(prices)-2]
}

// calculate52WeekHighLow calculates 52-week high and low from all available data
func (s *SummaryGenerator) calculate52WeekHighLow(data []map[string]string) (float64, float64) {
	if len(data) == 0 {
		return 0.0, 0.0
	}
	
	var high, low float64 = 0.0, 999999.0
	
	// Look at up to 252 trading days (roughly 52 weeks)
	startIndex := 0
	if len(data) > 252 {
		startIndex = len(data) - 252
	}
	
	for i := startIndex; i < len(data); i++ {
		// Use the available column names (snake_case keys)
		if highPrice, err := strconv.ParseFloat(data[i]["high_price"], 64); err == nil && highPrice > 0 {
			if highPrice > high {
				high = highPrice
			}
		}
		
		if lowPrice, err := strconv.ParseFloat(data[i]["low_price"], 64); err == nil && lowPrice > 0 {
			if lowPrice < low {
				low = lowPrice
			}
		}
	}
	
	if low == 999999.0 {
		low = 0.0
	}
	
	return high, low
}

// getDailyVolumeValue extracts daily volume and value from the latest record
func (s *SummaryGenerator) getDailyVolumeValue(record map[string]string) (int64, float64) {
	volume, _ := strconv.ParseInt(record["volume"], 10, 64)
	value, _ := strconv.ParseFloat(record["value"], 64)
	return volume, value
}