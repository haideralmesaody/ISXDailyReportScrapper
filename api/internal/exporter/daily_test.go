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

// Helper function to create test trade records
func createTestTradeRecords() []domain.TradeRecord {
	date1, _ := time.Parse("2006-01-02", "2024-01-15")
	date2, _ := time.Parse("2006-01-02", "2024-01-16")
	date3, _ := time.Parse("2006-01-02", "2024-01-17")

	return []domain.TradeRecord{
		{
			CompanyName:      "Apple Inc",
			CompanySymbol:    "AAPL",
			Date:             date1,
			OpenPrice:        150.0,
			HighPrice:        155.0,
			LowPrice:         148.0,
			AveragePrice:     152.5,
			PrevAveragePrice: 150.0,
			ClosePrice:       153.0,
			PrevClosePrice:   149.0,
			Change:           4.0,
			ChangePercent:    2.68,
			NumTrades:        1000,
			Volume:           500000,
			Value:            76250000.0,
			TradingStatus:    true,
		},
		{
			CompanyName:      "Microsoft Corp",
			CompanySymbol:    "MSFT",
			Date:             date1,
			OpenPrice:        280.0,
			HighPrice:        285.0,
			LowPrice:         278.0,
			AveragePrice:     282.0,
			PrevAveragePrice: 279.0,
			ClosePrice:       284.0,
			PrevClosePrice:   279.0,
			Change:           5.0,
			ChangePercent:    1.79,
			NumTrades:        800,
			Volume:           300000,
			Value:            84600000.0,
			TradingStatus:    true,
		},
		{
			CompanyName:      "Apple Inc",
			CompanySymbol:    "AAPL",
			Date:             date2,
			OpenPrice:        153.0,
			HighPrice:        157.0,
			LowPrice:         151.0,
			AveragePrice:     154.5,
			PrevAveragePrice: 152.5,
			ClosePrice:       156.0,
			PrevClosePrice:   153.0,
			Change:           3.0,
			ChangePercent:    1.96,
			NumTrades:        1200,
			Volume:           600000,
			Value:            92700000.0,
			TradingStatus:    true,
		},
		{
			CompanyName:      "Google Inc",
			CompanySymbol:    "GOOGL",
			Date:             date2,
			OpenPrice:        2800.0,
			HighPrice:        2850.0,
			LowPrice:         2780.0,
			AveragePrice:     2815.0,
			PrevAveragePrice: 2800.0,
			ClosePrice:       2840.0,
			PrevClosePrice:   2800.0,
			Change:           40.0,
			ChangePercent:    1.43,
			NumTrades:        500,
			Volume:           100000,
			Value:            281500000.0,
			TradingStatus:    true,
		},
		{
			CompanyName:      "Tesla Inc",
			CompanySymbol:    "TSLA",
			Date:             date3,
			OpenPrice:        200.0,
			HighPrice:        205.0,
			LowPrice:         195.0,
			AveragePrice:     201.0,
			PrevAveragePrice: 198.0,
			ClosePrice:       203.0,
			PrevClosePrice:   198.0,
			Change:           5.0,
			ChangePercent:    2.53,
			NumTrades:        900,
			Volume:           400000,
			Value:            80400000.0,
			TradingStatus:    true,
		},
	}
}

func TestNewDailyExporter(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewDailyExporter(paths)
	
	assert.NotNil(t, exporter)
	assert.NotNil(t, exporter.csvWriter)
	assert.Equal(t, paths, exporter.csvWriter.paths)
}

func TestDailyExporter_GetHeaders(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewDailyExporter(paths)
	
	headers := exporter.getHeaders()
	expectedHeaders := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
	
	assert.Equal(t, expectedHeaders, headers)
}

func TestDailyExporter_RecordToCSVRow(t *testing.T) {
	paths := &config.Paths{}
	exporter := NewDailyExporter(paths)
	
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

func TestDailyExporter_ExportDailyReports(t *testing.T) {
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "daily_exporter_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	tests := []struct {
		name        string
		records     []domain.TradeRecord
		outputDir   string
		expectError bool
		validate    func(t *testing.T, outputDir string)
	}{
		{
			name:        "export multiple days",
			records:     createTestTradeRecords(),
			outputDir:   "daily_reports",
			expectError: false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Check that files were created for each date
				expectedFiles := []string{
					"isx_daily_2024_01_15.csv",
					"isx_daily_2024_01_16.csv",
					"isx_daily_2024_01_17.csv",
				}
				
				for _, filename := range expectedFiles {
					filePath := filepath.Join(fullOutputDir, filename)
					_, err := os.Stat(filePath)
					assert.NoError(t, err, "File %s should exist", filename)
				}
				
				// Validate content of first file (2024-01-15)
				file1Path := filepath.Join(fullOutputDir, "isx_daily_2024_01_15.csv")
				content, err := os.ReadFile(file1Path)
				require.NoError(t, err)
				
				// Remove BOM and parse
				contentWithoutBOM := content[3:]
				lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
				
				// Should have header + 2 records (AAPL and MSFT for 2024-01-15)
				assert.Len(t, lines, 3)
				
				// Check header
				expectedHeader := "Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus"
				assert.Equal(t, expectedHeader, lines[0])
				
				// Records should be sorted by symbol (AAPL before MSFT)
				assert.Contains(t, lines[1], "AAPL")
				assert.Contains(t, lines[2], "MSFT")
			},
		},
		{
			name:        "export single day",
			records:     createTestTradeRecords()[:2], // Only first two records (same date)
			outputDir:   "single_day",
			expectError: false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Should only have one file
				files, err := os.ReadDir(fullOutputDir)
				require.NoError(t, err)
				assert.Len(t, files, 1)
				assert.Equal(t, "isx_daily_2024_01_15.csv", files[0].Name())
			},
		},
		{
			name:        "export empty records",
			records:     []domain.TradeRecord{},
			outputDir:   "empty_records",
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
			err := exporter.ExportDailyReports(tt.records, tt.outputDir)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, tt.outputDir)
			}
		})
	}
}

func TestDailyExporter_ExportCombinedData(t *testing.T) {
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "combined_export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	records := createTestTradeRecords()
	outputPath := "combined_data.csv"

	err = exporter.ExportCombinedData(records, outputPath)
	assert.NoError(t, err)

	// Validate file content
	filePath := filepath.Join(tempDir, "reports", outputPath)
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	// Note: Combined export uses BOMPrefix: false
	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	require.NoError(t, err)

	// Should have header + all records
	assert.Len(t, allRecords, 6) // header + 5 records

	// Check header
	expectedHeaders := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
	assert.Equal(t, expectedHeaders, allRecords[0])

	// Records should be sorted by date first, then by symbol
	// First record should be AAPL from 2024-01-15
	assert.Equal(t, "2024-01-15", allRecords[1][0])
	assert.Equal(t, "AAPL", allRecords[1][2])

	// Second record should be MSFT from 2024-01-15
	assert.Equal(t, "2024-01-15", allRecords[2][0])
	assert.Equal(t, "MSFT", allRecords[2][2])

	// Third record should be AAPL from 2024-01-16
	assert.Equal(t, "2024-01-16", allRecords[3][0])
	assert.Equal(t, "AAPL", allRecords[3][2])
}

func TestDailyExporter_ExportDailyReportsStreaming(t *testing.T) {
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "streaming_export_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	tests := []struct {
		name          string
		records       []domain.TradeRecord
		outputDir     string
		existingDates map[string]bool
		expectError   bool
		validate      func(t *testing.T, outputDir string)
	}{
		{
			name:          "streaming export all dates",
			records:       createTestTradeRecords(),
			outputDir:     "streaming_all",
			existingDates: nil,
			expectError:   false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Check that files were created for each date
				expectedFiles := []string{
					"isx_daily_2024_01_15.csv",
					"isx_daily_2024_01_16.csv",
					"isx_daily_2024_01_17.csv",
				}
				
				for _, filename := range expectedFiles {
					filePath := filepath.Join(fullOutputDir, filename)
					_, err := os.Stat(filePath)
					assert.NoError(t, err, "File %s should exist", filename)
					
					// Verify each file has BOM (streaming uses BOM)
					content, err := os.ReadFile(filePath)
					require.NoError(t, err)
					assert.True(t, len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF)
				}
			},
		},
		{
			name:          "streaming export with existing dates",
			records:       createTestTradeRecords(),
			outputDir:     "streaming_selective",
			existingDates: map[string]bool{"2024_01_15": true, "2024_01_17": true},
			expectError:   false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Only 2024-01-16 should be created
				expectedFile := "isx_daily_2024_01_16.csv"
				filePath := filepath.Join(fullOutputDir, expectedFile)
				_, err := os.Stat(filePath)
				assert.NoError(t, err, "File %s should exist", expectedFile)
				
				// Files for existing dates should not exist
				notExpectedFiles := []string{
					"isx_daily_2024_01_15.csv",
					"isx_daily_2024_01_17.csv",
				}
				
				for _, filename := range notExpectedFiles {
					filePath := filepath.Join(fullOutputDir, filename)
					_, err := os.Stat(filePath)
					assert.True(t, os.IsNotExist(err), "File %s should not exist", filename)
				}
			},
		},
		{
			name:          "streaming export empty existing dates",
			records:       createTestTradeRecords()[:2], // Only one date
			outputDir:     "streaming_empty_existing",
			existingDates: map[string]bool{},
			expectError:   false,
			validate: func(t *testing.T, outputDir string) {
				fullOutputDir := filepath.Join(tempDir, "reports", outputDir)
				
				// Should create file for the one date
				expectedFile := "isx_daily_2024_01_15.csv"
				filePath := filepath.Join(fullOutputDir, expectedFile)
				_, err := os.Stat(filePath)
				assert.NoError(t, err, "File %s should exist", expectedFile)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := exporter.ExportDailyReportsStreaming(tt.records, tt.outputDir, tt.existingDates)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, tt.outputDir)
			}
		})
	}
}

func TestDailyExporter_RecordSorting(t *testing.T) {
	// Test that records are properly sorted by date and symbol
	tempDir, err := os.MkdirTemp("", "sorting_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	// Create records in mixed order
	date1, _ := time.Parse("2006-01-02", "2024-01-15")
	date2, _ := time.Parse("2006-01-02", "2024-01-16")
	
	records := []domain.TradeRecord{
		{CompanySymbol: "ZULU", CompanyName: "Zulu Corp", Date: date1, ClosePrice: 100, TradingStatus: true},
		{CompanySymbol: "ALPHA", CompanyName: "Alpha Inc", Date: date2, ClosePrice: 200, TradingStatus: true},
		{CompanySymbol: "BETA", CompanyName: "Beta Ltd", Date: date1, ClosePrice: 150, TradingStatus: true},
		{CompanySymbol: "CHARLIE", CompanyName: "Charlie Co", Date: date2, ClosePrice: 175, TradingStatus: true},
	}

	// Test combined export (sorts by date first, then symbol)
	err = exporter.ExportCombinedData(records, "sorted_combined.csv")
	require.NoError(t, err)

	filePath := filepath.Join(tempDir, "reports", "sorted_combined.csv")
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	require.NoError(t, err)

	// Check sorting: date1 records first (BETA, ZULU), then date2 records (ALPHA, CHARLIE)
	assert.Equal(t, "BETA", allRecords[1][2])  // First date1 record
	assert.Equal(t, "ZULU", allRecords[2][2])  // Second date1 record
	assert.Equal(t, "ALPHA", allRecords[3][2]) // First date2 record
	assert.Equal(t, "CHARLIE", allRecords[4][2]) // Second date2 record
}

func TestDailyExporter_SpecialCharactersInData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "special_chars_daily_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	date, _ := time.Parse("2006-01-02", "2024-01-15")
	
	// Create record with special characters
	record := domain.TradeRecord{
		CompanyName:   "Company, Inc \"Special\"",
		CompanySymbol: "SPEC",
		Date:          date,
		OpenPrice:     123.45,
		HighPrice:     125.67,
		LowPrice:      120.23,
		AveragePrice:  123.45,
		ClosePrice:    124.56,
		NumTrades:     100,
		Volume:        50000,
		Value:         6172800,
		TradingStatus: true,
	}

	err = exporter.ExportCombinedData([]domain.TradeRecord{record}, "special_chars.csv")
	require.NoError(t, err)

	// Read back and verify CSV escaping worked
	filePath := filepath.Join(tempDir, "reports", "special_chars.csv")
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	allRecords, err := reader.ReadAll()
	require.NoError(t, err)

	assert.Len(t, allRecords, 2) // header + 1 record
	assert.Equal(t, "Company, Inc \"Special\"", allRecords[1][1]) // Company name should be preserved
	assert.Equal(t, "SPEC", allRecords[1][2])
}

// BenchmarkDailyExporter_ExportDailyReports tests performance of daily export
func BenchmarkDailyExporter_ExportDailyReports(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_daily_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	// Create larger dataset for benchmarking
	var records []domain.TradeRecord
	baseDate, _ := time.Parse("2006-01-02", "2024-01-01")
	
	for day := 0; day < 10; day++ {
		currentDate := baseDate.AddDate(0, 0, day)
		for i := 0; i < 100; i++ {
			records = append(records, domain.TradeRecord{
				CompanyName:   "Company" + string(rune('A'+i%26)),
				CompanySymbol: "SYM" + string(rune('A'+i%26)),
				Date:          currentDate,
				OpenPrice:     100.0 + float64(i),
				HighPrice:     105.0 + float64(i),
				LowPrice:      95.0 + float64(i),
				AveragePrice:  100.0 + float64(i),
				ClosePrice:    102.0 + float64(i),
				NumTrades:     100 + int64(i),
				Volume:        10000 + int64(i*100),
				Value:         1000000.0 + float64(i*1000),
				TradingStatus: true,
			})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputDir := "bench_daily_" + string(rune('A'+i%26))
		err := exporter.ExportDailyReports(records, outputDir)
		require.NoError(b, err)
	}
}

// BenchmarkDailyExporter_ExportCombinedData tests performance of combined export
func BenchmarkDailyExporter_ExportCombinedData(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_combined_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	paths := &config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	}
	exporter := NewDailyExporter(paths)

	// Create test dataset
	var records []domain.TradeRecord
	baseDate, _ := time.Parse("2006-01-02", "2024-01-01")
	
	for i := 0; i < 5000; i++ {
		records = append(records, domain.TradeRecord{
			CompanyName:   "Company" + string(rune(i%100+'A')),
			CompanySymbol: "SYM" + string(rune(i%100+'A')),
			Date:          baseDate.AddDate(0, 0, i%30),
			OpenPrice:     100.0 + float64(i%1000),
			HighPrice:     105.0 + float64(i%1000),
			LowPrice:      95.0 + float64(i%1000),
			AveragePrice:  100.0 + float64(i%1000),
			ClosePrice:    102.0 + float64(i%1000),
			NumTrades:     100 + int64(i),
			Volume:        10000 + int64(i*100),
			Value:         1000000.0 + float64(i*1000),
			TradingStatus: true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputFile := "bench_combined_" + string(rune('A'+i%26)) + ".csv"
		err := exporter.ExportCombinedData(records, outputFile)
		require.NoError(b, err)
	}
}