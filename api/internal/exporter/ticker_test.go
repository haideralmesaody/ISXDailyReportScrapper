package exporter

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

// Helper function to create test trade records for ticker testing
func createTickerTestRecords() []domain.TradeRecord {
	date1, _ := time.Parse("2006-01-02", "2024-01-15")
	date2, _ := time.Parse("2006-01-02", "2024-01-16")
	date3, _ := time.Parse("2006-01-02", "2024-01-17")

	return []domain.TradeRecord{
		{
			CompanyName:   "Apple Inc",
			CompanySymbol: "AAPL",
			Date:          date1,
			OpenPrice:     150.0,
			HighPrice:     155.0,
			LowPrice:      148.0,
			ClosePrice:    153.0,
			NumTrades:     1000,
			Volume:        500000,
			Value:         76250000.0,
			TradingStatus: true,
		},
		{
			CompanyName:   "Apple Inc",
			CompanySymbol: "AAPL",
			Date:          date2,
			OpenPrice:     153.0,
			HighPrice:     157.0,
			LowPrice:      151.0,
			ClosePrice:    156.0,
			NumTrades:     1200,
			Volume:        600000,
			Value:         92700000.0,
			TradingStatus: true,
		},
		{
			CompanyName:   "Apple Inc",
			CompanySymbol: "AAPL",
			Date:          date3,
			OpenPrice:     156.0,
			HighPrice:     159.0,
			LowPrice:      154.0,
			ClosePrice:    0.0, // No trading this day
			NumTrades:     0,
			Volume:        0,
			Value:         0.0,
			TradingStatus: false,
		},
		{
			CompanyName:   "Microsoft Corp",
			CompanySymbol: "MSFT",
			Date:          date1,
			OpenPrice:     280.0,
			HighPrice:     285.0,
			LowPrice:      278.0,
			ClosePrice:    284.0,
			NumTrades:     800,
			Volume:        300000,
			Value:         84600000.0,
			TradingStatus: true,
		},
		{
			CompanyName:   "Microsoft Corp",
			CompanySymbol: "MSFT",
			Date:          date2,
			OpenPrice:     284.0,
			HighPrice:     288.0,
			LowPrice:      282.0,
			ClosePrice:    287.0,
			NumTrades:     900,
			Volume:        350000,
			Value:         100450000.0,
			TradingStatus: true,
		},
		{
			CompanyName:   "Google Inc",
			CompanySymbol: "GOOGL",
			Date:          date2,
			OpenPrice:     2800.0,
			HighPrice:     2850.0,
			LowPrice:      2780.0,
			ClosePrice:    2840.0,
			NumTrades:     500,
			Volume:        100000,
			Value:         281500000.0,
			TradingStatus: true,
		},
	}
}

func TestNewTickerExporter(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)
	
	assert.NotNil(t, exporter)
	assert.NotNil(t, exporter.csvWriter)
	assert.Equal(t, paths, exporter.csvWriter.paths)
}

func TestTickerExporter_GetHeaders(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)
	
	headers := exporter.getHeaders()
	expectedHeaders := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
	
	assert.Equal(t, expectedHeaders, headers)
}

func TestTickerExporter_RecordToCSVRow(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)
	
	date, _ := time.Parse("2006-01-02", "2024-01-15")
	record := domain.TradeRecord{
		CompanyName:      "Test Company",
		CompanySymbol:    "TEST",
		Date:             date,
		OpenPrice:        100.50,
		HighPrice:        105.75,
		LowPrice:         99.25,
		AveragePrice:     102.50,
		PrevAveragePrice: 101.00,
		ClosePrice:       104.25,
		PrevClosePrice:   101.00,
		Change:           3.25,
		ChangePercent:    3.22,
		NumTrades:        150,
		Volume:           75000,
		Value:            7687500.0,
		TradingStatus:    true,
	}
	
	csvRow := exporter.recordToCSVRow(record)
	expectedRow := []string{
		"2024-01-15",
		"Test Company",
		"TEST",
		"100.50",      // Float formatting: 2 decimal places
		"105.75",
		"99.25",
		"102.50",      // Float formatting: 2 decimal places
		"101.00",      // Float formatting: 2 decimal places
		"104.25",
		"101.00",      // Float formatting: 2 decimal places
		"3.25",
		"3.22",
		"150",
		"75000",
		"7687500.00",  // Float formatting: 2 decimal places
		"true",
	}
	
	assert.Equal(t, expectedRow, csvRow)
}

func TestTickerExporter_ExportTickerFiles(t *testing.T) {
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "ticker_export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewTickerExporter(paths)

	tests := []struct {
		name        string
		records     []domain.TradeRecord
		outputDir   string
		expectError bool
		validate    func(t *testing.T, outputDir string)
	}{
		{
			name:        "export multiple tickers",
			records:     createTickerTestRecords(),
			outputDir:   "ticker_files",
			expectError: false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Check that files were created for each ticker
				expectedFiles := []string{
					"AAPL_trading_history.csv",
					"MSFT_trading_history.csv",
					"GOOGL_trading_history.csv",
				}
				
				for _, filename := range expectedFiles {
					filePath := filepath.Join(fullOutputDir, filename)
					_, err := os.Stat(filePath)
					assert.NoError(t, err, "File %s should exist", filename)
				}
				
				// Validate AAPL file content (should have 3 records sorted by date)
				aaplPath := filepath.Join(fullOutputDir, "AAPL_trading_history.csv")
				content, err := os.ReadFile(aaplPath)
				require.NoError(t, err)
				
				// Remove BOM and parse
				contentWithoutBOM := content[3:]
				lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
				
				// Should have header + 3 records
				assert.Len(t, lines, 4)
				
				// Check that records are sorted by date (oldest first)
				assert.Contains(t, lines[1], "2024-01-15") // First record
				assert.Contains(t, lines[2], "2024-01-16") // Second record
				assert.Contains(t, lines[3], "2024-01-17") // Third record
			},
		},
		{
			name:        "export single ticker",
			records:     createTickerTestRecords()[:2], // Only AAPL records from first two dates
			outputDir:   "single_ticker",
			expectError: false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Should only have AAPL file
				files, err := os.ReadDir(fullOutputDir)
				require.NoError(t, err)
				assert.Len(t, files, 1)
				assert.Equal(t, "AAPL_trading_history.csv", files[0].Name())
			},
		},
		{
			name:        "export empty records",
			records:     []domain.TradeRecord{},
			outputDir:   "empty_ticker_records",
			expectError: false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Should have no files
				_, err := os.Stat(fullOutputDir)
				assert.True(t, os.IsNotExist(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exporter.ExportTickerFiles(tt.records, tt.outputDir)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, tt.outputDir)
			}
		})
	}
}

func TestTickerExporter_GenerateTickerSummaries(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)

	tests := []struct {
		name     string
		records  []domain.TradeRecord
		validate func(t *testing.T, summaries []TickerSummary)
	}{
		{
			name:    "generate summaries for multiple tickers",
			records: createTickerTestRecords(),
			validate: func(t *testing.T, summaries []TickerSummary) {
				// Should have summaries for AAPL, MSFT, and GOOGL
				assert.Len(t, summaries, 3)
				
				// Find AAPL summary
				var aaplSummary *TickerSummary
				for i := range summaries {
					if summaries[i].Ticker == "AAPL" {
						aaplSummary = &summaries[i]
						break
					}
				}
				require.NotNil(t, aaplSummary, "AAPL summary should exist")
				
				// AAPL has 2 trading days (third day has ClosePrice = 0)
				assert.Equal(t, "AAPL", aaplSummary.Ticker)
				assert.Equal(t, "Apple Inc", aaplSummary.CompanyName)
				assert.Equal(t, 2, aaplSummary.TradingDays)
				assert.Equal(t, int64(1100000), aaplSummary.TotalVolume) // 500000 + 600000
				assert.Equal(t, 168950000.0, aaplSummary.TotalValue)     // 76250000 + 92700000
				assert.Equal(t, 154.5, aaplSummary.AveragePrice)         // (153 + 156) / 2
				assert.Equal(t, 157.0, aaplSummary.HighestPrice)         // max of HighPrice fields
				assert.Equal(t, 148.0, aaplSummary.LowestPrice)          // min of LowPrice fields
				assert.Equal(t, 156.0, aaplSummary.LastPrice)            // Last trading close price
				assert.Equal(t, "2024-01-16", aaplSummary.LastDate)      // Last trading date
				assert.Equal(t, "2/10", aaplSummary.Last10Days)          // 2 out of last 10 days
				
				// Find MSFT summary
				var msftSummary *TickerSummary
				for i := range summaries {
					if summaries[i].Ticker == "MSFT" {
						msftSummary = &summaries[i]
						break
					}
				}
				require.NotNil(t, msftSummary, "MSFT summary should exist")
				
				// MSFT has 2 trading days
				assert.Equal(t, "MSFT", msftSummary.Ticker)
				assert.Equal(t, "Microsoft Corp", msftSummary.CompanyName)
				assert.Equal(t, 2, msftSummary.TradingDays)
				assert.Equal(t, 285.5, msftSummary.AveragePrice) // (284 + 287) / 2
				assert.Equal(t, 287.0, msftSummary.LastPrice)
				assert.Equal(t, "2024-01-16", msftSummary.LastDate)
				
				// Find GOOGL summary
				var googlSummary *TickerSummary
				for i := range summaries {
					if summaries[i].Ticker == "GOOGL" {
						googlSummary = &summaries[i]
						break
					}
				}
				require.NotNil(t, googlSummary, "GOOGL summary should exist")
				
				// GOOGL has 1 trading day
				assert.Equal(t, "GOOGL", googlSummary.Ticker)
				assert.Equal(t, "Google Inc", googlSummary.CompanyName)
				assert.Equal(t, 1, googlSummary.TradingDays)
				assert.Equal(t, 2840.0, googlSummary.AveragePrice) // Only one price
				assert.Equal(t, 2840.0, googlSummary.LastPrice)
				assert.Equal(t, "2024-01-16", googlSummary.LastDate)
				assert.Equal(t, "1/10", googlSummary.Last10Days)
			},
		},
		{
			name:    "generate summaries with no trading days",
			records: []domain.TradeRecord{
				{
					CompanyName:   "Non-Trading Co",
					CompanySymbol: "NOTRADE",
					Date:          time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
					ClosePrice:    0.0, // No trading
					TradingStatus: false,
				},
			},
			validate: func(t *testing.T, summaries []TickerSummary) {
				assert.Len(t, summaries, 1)
				
				summary := summaries[0]
				assert.Equal(t, "NOTRADE", summary.Ticker)
				assert.Equal(t, "Non-Trading Co", summary.CompanyName)
				assert.Equal(t, 0, summary.TradingDays)
				assert.Equal(t, int64(0), summary.TotalVolume)
				assert.Equal(t, 0.0, summary.TotalValue)
				assert.Equal(t, 0.0, summary.AveragePrice)
				assert.Equal(t, 0.0, summary.LastPrice)
				assert.Equal(t, "", summary.LastDate)
				assert.Equal(t, "", summary.Last10Days) // Empty when no trading days
			},
		},
		{
			name:    "empty records",
			records: []domain.TradeRecord{},
			validate: func(t *testing.T, summaries []TickerSummary) {
				assert.Len(t, summaries, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summaries := exporter.GenerateTickerSummaries(tt.records)
			tt.validate(t, summaries)
		})
	}
}

func TestTickerExporter_SummaryToCSVRow(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)
	
	summary := TickerSummary{
		Ticker:         "TEST",
		CompanyName:    "Test Company",
		LastPrice:      123.45,
		LastDate:       "2024-01-15",
		TradingDays:    10,
		Last10Days:     "8/10",
		TotalVolume:    1000000,
		TotalValue:     123450000.0,
		AveragePrice:   123.45,
		HighestPrice:   130.00,
		LowestPrice:    115.50,
	}
	
	csvRow := exporter.summaryToCSVRow(summary)
	expectedRow := []string{
		"TEST",
		"Test Company",
		"123.45",
		"2024-01-15",
		"10",
		"8/10",
		"1000000",
		"123450000.00",  // Float formatting: 2 decimal places
		"123.45",
		"130.00",        // Float formatting: 2 decimal places
		"115.50",        // Float formatting: 2 decimal places
	}
	
	assert.Equal(t, expectedRow, csvRow)
}

func TestTickerExporter_ExportTickerSummary(t *testing.T) {
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "ticker_summary_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewTickerExporter(paths)

	// Create test summaries
	summaries := []TickerSummary{
		{
			Ticker:         "ZULU",
			CompanyName:    "Zulu Corp",
			LastPrice:      100.0,
			LastDate:       "2024-01-15",
			TradingDays:    5,
			Last10Days:     "5/10",
			TotalVolume:    500000,
			TotalValue:     50000000.0,
			AveragePrice:   100.0,
			HighestPrice:   105.0,
			LowestPrice:    95.0,
		},
		{
			Ticker:         "ALPHA",
			CompanyName:    "Alpha Inc",
			LastPrice:      200.0,
			LastDate:       "2024-01-16",
			TradingDays:    8,
			Last10Days:     "8/10",
			TotalVolume:    800000,
			TotalValue:     160000000.0,
			AveragePrice:   200.0,
			HighestPrice:   210.0,
			LowestPrice:    190.0,
		},
	}

	outputPath := "ticker_summary.csv"
	err = exporter.ExportTickerSummary(summaries, outputPath)
	assert.NoError(t, err)

	// Validate file content
	filePath := filepath.Join(tempDir, "reports", outputPath)
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	// Skip BOM
	bom := make([]byte, 3)
	_, err = file.Read(bom)
	require.NoError(t, err)

	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + 2 records
	assert.Len(t, allRecords, 3)

	// Check header
	expectedHeaders := []string{
		"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days",
		"TotalVolume", "TotalValue", "AveragePrice", "HighestPrice", "LowestPrice",
	}
	assert.Equal(t, expectedHeaders, allRecords[0])

	// Records should be sorted by ticker (ALPHA before ZULU)
	assert.Equal(t, "ALPHA", allRecords[1][0])
	assert.Equal(t, "ZULU", allRecords[2][0])

	// Verify ALPHA record values
	assert.Equal(t, "Alpha Inc", allRecords[1][1])
	assert.Equal(t, "200.00", allRecords[1][2])  // Float formatting: 2 decimal places
	assert.Equal(t, "2024-01-16", allRecords[1][3])
	assert.Equal(t, "8", allRecords[1][4])
	assert.Equal(t, "8/10", allRecords[1][5])
}

func TestTickerExporter_IntegratedWorkflow(t *testing.T) {
	// Test the complete workflow: records -> summaries -> export
	tempDir, err := os.MkdirTemp("", "ticker_workflow_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewTickerExporter(paths)

	records := createTickerTestRecords()

	// Step 1: Generate summaries
	summaries := exporter.GenerateTickerSummaries(records)
	assert.Len(t, summaries, 3) // AAPL, MSFT, GOOGL

	// Step 2: Export ticker files
	err = exporter.ExportTickerFiles(records, "ticker_history")
	assert.NoError(t, err)

	// Step 3: Export summary
	err = exporter.ExportTickerSummary(summaries, "summary.csv")
	assert.NoError(t, err)

	// Validate all files exist
	historyDir := filepath.Join(tempDir, "reports", "ticker_history")
	files, err := os.ReadDir(historyDir)
	require.NoError(t, err)
	assert.Len(t, files, 3) // One file per ticker

	summaryFile := filepath.Join(tempDir, "reports", "summary.csv")
	_, err = os.Stat(summaryFile)
	assert.NoError(t, err)
}

func TestTickerExporter_Last10DaysCalculation(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)

	// Create records with exactly 15 days, where 12 have trading activity
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	var records []domain.TradeRecord

	for i := 0; i < 15; i++ {
		date := baseDate.AddDate(0, 0, i)
		closePrice := 0.0
		tradingStatus := false
		
		// Make only 12 days have trading activity (skip days 2, 5, 8)
		if i != 2 && i != 5 && i != 8 {
			closePrice = 100.0 + float64(i)
			tradingStatus = true
		}
		
		records = append(records, domain.TradeRecord{
			CompanyName:   "Test Company",
			CompanySymbol: "TEST",
			Date:          date,
			ClosePrice:    closePrice,
			TradingStatus: tradingStatus,
		})
	}

	summaries := exporter.GenerateTickerSummaries(records)
	require.Len(t, summaries, 1)

	summary := summaries[0]
	
	// Should have 12 total trading days
	assert.Equal(t, 12, summary.TradingDays)
	
	// For last 10 days calculation, we look backwards from the end and count trading days up to 10
	// The algorithm in the code stops when last10Count reaches 10 OR runs out of records
	// Since we have 12 trading days out of 15 total, and it counts backwards, it will find 10 trading days
	assert.Equal(t, "10/10", summary.Last10Days)
}

func TestTickerExporter_SpecialCharactersInTicker(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ticker_special_chars_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewTickerExporter(paths)

	date := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	
	// Create record with special characters in company name
	record := domain.TradeRecord{
		CompanyName:   "Company, Inc \"Special\"",
		CompanySymbol: "SPEC&CO",
		Date:          date,
		ClosePrice:    100.0,
		TradingStatus: true,
	}

	// Test ticker files export
	err = exporter.ExportTickerFiles([]domain.TradeRecord{record}, "special_chars")
	assert.NoError(t, err)

	// Verify file was created with safe filename (& might be problematic in some filesystems)
	outputDir := filepath.Join(tempDir, "reports", "special_chars")
	files, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	assert.Len(t, files, 1)

	// Read and verify content preserves special characters
	filePath := filepath.Join(outputDir, files[0].Name())
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	// Skip BOM
	bom := make([]byte, 3)
	_, err = file.Read(bom)
	require.NoError(t, err)

	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	require.NoError(t, err)

	assert.Len(t, allRecords, 2) // header + 1 record
	assert.Equal(t, "Company, Inc \"Special\"", allRecords[1][1]) // Company name preserved
	assert.Equal(t, "SPEC&CO", allRecords[1][2])                 // Symbol preserved
}

// BenchmarkTickerExporter_GenerateTickerSummaries tests performance of summary generation
func BenchmarkTickerExporter_GenerateTickerSummaries(b *testing.B) {
	paths := &config.Paths{}
	exporter := NewTickerExporter(paths)

	// Create larger dataset for benchmarking
	var records []domain.TradeRecord
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Create 100 tickers with 30 days each
	for ticker := 0; ticker < 100; ticker++ {
		for day := 0; day < 30; day++ {
			records = append(records, domain.TradeRecord{
				CompanyName:   "Company" + string(rune('A'+ticker%26)),
				CompanySymbol: "TKR" + string(rune('A'+ticker%26)) + string(rune('0'+ticker/26)),
				Date:          baseDate.AddDate(0, 0, day),
				OpenPrice:     100.0 + float64(ticker),
				HighPrice:     105.0 + float64(ticker),
				LowPrice:      95.0 + float64(ticker),
				ClosePrice:    102.0 + float64(ticker),
				NumTrades:     100 + int64(ticker),
				Volume:        10000 + int64(ticker*100),
				Value:         1000000.0 + float64(ticker*10000),
				TradingStatus: true,
			})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = exporter.GenerateTickerSummaries(records)
	}
}

// BenchmarkTickerExporter_ExportTickerFiles tests performance of ticker file export
func BenchmarkTickerExporter_ExportTickerFiles(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_ticker_export_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewTickerExporter(paths)

	// Create test dataset
	var records []domain.TradeRecord
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Create 50 tickers with 100 records each
	for ticker := 0; ticker < 50; ticker++ {
		for day := 0; day < 100; day++ {
			records = append(records, domain.TradeRecord{
				CompanyName:   "Company" + string(rune('A'+ticker%26)),
				CompanySymbol: "TKR" + string(rune('A'+ticker%26)),
				Date:          baseDate.AddDate(0, 0, day),
				ClosePrice:    100.0 + float64(ticker),
				TradingStatus: true,
			})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputDir := "bench_ticker_" + string(rune('A'+i%26))
		err := exporter.ExportTickerFiles(records, outputDir)
		require.NoError(b, err)
	}
}