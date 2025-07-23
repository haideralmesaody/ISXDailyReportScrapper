package testutil

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestDataGenerator helps create test data for pipeline tests
type TestDataGenerator struct {
	t       *testing.T
	baseDir string
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(t *testing.T, baseDir string) *TestDataGenerator {
	return &TestDataGenerator{
		t:       t,
		baseDir: baseDir,
	}
}

// GenerateISXReport generates a mock ISX daily report Excel file
func (g *TestDataGenerator) GenerateISXReport(date string) string {
	// Parse date to format filename
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		g.t.Fatalf("invalid date format: %s", date)
	}
	
	filename := fmt.Sprintf("%s %s %s ISX Daily Report.xlsx", parts[0], parts[1], parts[2])
	content := fmt.Sprintf("Mock Excel data for %s", date)
	
	return CreateTestFile(g.t, g.baseDir, filename, content)
}

// GenerateDateRange generates mock ISX reports for a date range
func (g *TestDataGenerator) GenerateDateRange(fromDate, toDate string) []string {
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		g.t.Fatalf("invalid from date: %v", err)
	}
	
	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		g.t.Fatalf("invalid to date: %v", err)
	}
	
	var files []string
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		// Skip weekends (Friday and Saturday in Iraq)
		if d.Weekday() == time.Friday || d.Weekday() == time.Saturday {
			continue
		}
		
		dateStr := d.Format("2006-01-02")
		file := g.GenerateISXReport(dateStr)
		files = append(files, file)
	}
	
	return files
}

// GenerateCombinedCSV generates a mock combined data CSV
func (g *TestDataGenerator) GenerateCombinedCSV() string {
	headers := []string{
		"Date", "Symbol", "Company", "Open", "High", "Low", "Close",
		"Volume", "Value", "Trades", "Change", "ChangePercent", "TradingStatus",
	}
	
	rows := [][]string{
		{"2024-01-01", "BANK", "Bank of Baghdad", "1.50", "1.55", "1.48", "1.52", "1000000", "1520000", "50", "0.02", "1.33", "true"},
		{"2024-01-02", "BANK", "Bank of Baghdad", "1.52", "1.58", "1.51", "1.56", "1200000", "1872000", "60", "0.04", "2.63", "true"},
		{"2024-01-03", "BANK", "Bank of Baghdad", "1.56", "1.56", "1.56", "1.56", "0", "0", "0", "0.00", "0.00", "false"},
		{"2024-01-01", "TASC", "Asia Cell", "7.20", "7.30", "7.15", "7.25", "500000", "3625000", "30", "0.05", "0.69", "true"},
		{"2024-01-02", "TASC", "Asia Cell", "7.25", "7.35", "7.20", "7.30", "600000", "4380000", "35", "0.05", "0.69", "true"},
	}
	
	return CreateCSVFile(g.t, g.baseDir, "isx_combined_data.csv", headers, rows)
}

// GenerateIndexesCSV generates a mock indexes CSV
func (g *TestDataGenerator) GenerateIndexesCSV() string {
	headers := []string{"Date", "ISX60", "ISX15"}
	
	rows := [][]string{
		{"2024-01-01", "850.25", "1250.50"},
		{"2024-01-02", "852.30", "1253.75"},
		{"2024-01-03", "851.90", "1252.80"},
		{"2024-01-04", "853.45", "1255.20"},
		{"2024-01-05", "854.10", "1256.90"},
	}
	
	return CreateCSVFile(g.t, g.baseDir, "indexes.csv", headers, rows)
}

// GenerateTickerSummary generates a mock ticker summary JSON
func (g *TestDataGenerator) GenerateTickerSummary() string {
	content := `{
  "generated_at": "2024-01-05T15:30:00Z",
  "tickers": [
    {
      "symbol": "BANK",
      "company": "Bank of Baghdad",
      "last_price": 1.56,
      "change": 0.06,
      "change_percent": 4.00,
      "volume": 2200000,
      "value": 3392000,
      "trades": 110,
      "actual_trading_days": 2,
      "total_days": 3,
      "average_volume": 1100000,
      "average_value": 1696000
    },
    {
      "symbol": "TASC",
      "company": "Asia Cell",
      "last_price": 7.30,
      "change": 0.10,
      "change_percent": 1.39,
      "volume": 1100000,
      "value": 8005000,
      "trades": 65,
      "actual_trading_days": 2,
      "total_days": 2,
      "average_volume": 550000,
      "average_value": 4002500
    }
  ],
  "summary": {
    "total_tickers": 2,
    "active_tickers": 2,
    "total_volume": 3300000,
    "total_value": 11397000,
    "total_trades": 175
  }
}`
	
	return CreateJSONFile(g.t, g.baseDir, "ticker_summary.json", content)
}

// GenerateCorruptedExcel generates a corrupted Excel file for error testing
func (g *TestDataGenerator) GenerateCorruptedExcel(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		g.t.Fatalf("invalid date format: %s", date)
	}
	
	filename := fmt.Sprintf("%s %s %s ISX Daily Report.xlsx", parts[0], parts[1], parts[2])
	content := "This is not a valid Excel file - corrupted data for testing"
	
	return CreateTestFile(g.t, g.baseDir, filename, content)
}

// GenerateEmptyExcel generates an empty Excel file
func (g *TestDataGenerator) GenerateEmptyExcel(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		g.t.Fatalf("invalid date format: %s", date)
	}
	
	filename := fmt.Sprintf("%s %s %s ISX Daily Report.xlsx", parts[0], parts[1], parts[2])
	content := "" // Empty file
	
	return CreateTestFile(g.t, g.baseDir, filename, content)
}

// GenerateLargeDataset generates a large dataset for performance testing
func (g *TestDataGenerator) GenerateLargeDataset(numDays, tickersPerDay int) {
	// Generate date range
	startDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for i := 0; i < numDays; i++ {
		date := startDate.AddDate(0, 0, i)
		
		// Skip weekends
		if date.Weekday() == time.Friday || date.Weekday() == time.Saturday {
			continue
		}
		
		// Generate report for this day
		g.GenerateISXReport(date.Format("2006-01-02"))
	}
	
	// Generate large combined CSV
	var headers []string = []string{
		"Date", "Symbol", "Company", "Open", "High", "Low", "Close",
		"Volume", "Value", "Trades", "Change", "ChangePercent", "TradingStatus",
	}
	
	var rows [][]string
	symbols := []string{"BANK", "TASC", "IBSD", "AISP", "AMAP", "BNOI", "BCOI", "BGUC", "BIBI", "BIIB"}
	
	for i := 0; i < numDays; i++ {
		date := startDate.AddDate(0, 0, i)
		if date.Weekday() == time.Friday || date.Weekday() == time.Saturday {
			continue
		}
		
		dateStr := date.Format("2006-01-02")
		
		for j, symbol := range symbols {
			if j >= tickersPerDay {
				break
			}
			
			row := []string{
				dateStr,
				symbol,
				fmt.Sprintf("%s Company", symbol),
				fmt.Sprintf("%.2f", 1.0+float64(j)*0.1),
				fmt.Sprintf("%.2f", 1.05+float64(j)*0.1),
				fmt.Sprintf("%.2f", 0.95+float64(j)*0.1),
				fmt.Sprintf("%.2f", 1.02+float64(j)*0.1),
				fmt.Sprintf("%d", 100000*(j+1)),
				fmt.Sprintf("%d", 102000*(j+1)),
				fmt.Sprintf("%d", 10*(j+1)),
				"0.02",
				"2.00",
				"true",
			}
			rows = append(rows, row)
		}
	}
	
	CreateCSVFile(g.t, g.baseDir, "isx_combined_data.csv", headers, rows)
}

// CleanupTestData removes all test data
func (g *TestDataGenerator) CleanupTestData() {
	// Cleanup is handled by test cleanup in CreateTestDirectory
}