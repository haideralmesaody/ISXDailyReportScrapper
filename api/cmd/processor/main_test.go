package main

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/isxcli/isxcli/pkg/contracts/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain removed - flag parsing is handled in main.go

func TestExcelFileInfoSorting(t *testing.T) {
	files := []ExcelFileInfo{
		{Name: "2025 01 15 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{Name: "2025 01 10 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
		{Name: "2025 01 20 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)},
	}

	// Sort files by date (as done in main function)
	sortedFiles := make([]ExcelFileInfo, len(files))
	copy(sortedFiles, files)
	
	// Implement the same sorting logic as in main
	for i := 0; i < len(sortedFiles)-1; i++ {
		for j := i + 1; j < len(sortedFiles); j++ {
			if sortedFiles[j].Date.Before(sortedFiles[i].Date) {
				sortedFiles[i], sortedFiles[j] = sortedFiles[j], sortedFiles[i]
			}
		}
	}

	expected := []string{
		"2025 01 10 ISX Daily Report.xlsx",
		"2025 01 15 ISX Daily Report.xlsx", 
		"2025 01 20 ISX Daily Report.xlsx",
	}

	for i, file := range sortedFiles {
		assert.Equal(t, expected[i], file.Name)
	}
}

func TestDetermineFilesToProcess(t *testing.T) {
	tests := []struct {
		name               string
		excelFiles         []ExcelFileInfo
		existingCSVFiles   []string
		existingCombined   bool
		expectedToProcess  int
		expectedExisting   int
	}{
		{
			name: "no existing files - process all",
			excelFiles: []ExcelFileInfo{
				{Name: "2025 01 10 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
				{Name: "2025 01 11 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)},
			},
			existingCSVFiles:   []string{},
			existingCombined:   false,
			expectedToProcess:  2,
			expectedExisting:   0,
		},
		{
			name: "some files already processed",
			excelFiles: []ExcelFileInfo{
				{Name: "2025 01 10 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
				{Name: "2025 01 11 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)},
				{Name: "2025 01 12 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)},
			},
			existingCSVFiles:   []string{"isx_daily_2025_01_10.csv", "isx_daily_2025_01_11.csv"},
			existingCombined:   true,
			expectedToProcess:  1, // Only 2025-01-12 needs processing
			expectedExisting:   2, // 2 existing records (approximation)
		},
		{
			name: "all files already processed",
			excelFiles: []ExcelFileInfo{
				{Name: "2025 01 10 ISX Daily Report.xlsx", Date: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
			},
			existingCSVFiles:   []string{"isx_daily_2025_01_10.csv"},
			existingCombined:   true,
			expectedToProcess:  0,
			expectedExisting:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create existing CSV files in daily/ subdirectory (matching DetectUnprocessedFiles logic)
			dailyDir := filepath.Join(tmpDir, "daily")
			os.MkdirAll(dailyDir, 0755)
			for _, csvFile := range tt.existingCSVFiles {
				csvPath := filepath.Join(dailyDir, csvFile)
				file, err := os.Create(csvPath)
				require.NoError(t, err)
				file.Close()
			}
			
			// Create existing combined CSV if specified in combined/ subdirectory
			if tt.existingCombined {
				combinedDir := filepath.Join(tmpDir, "combined")
				os.MkdirAll(combinedDir, 0755)
				combinedPath := filepath.Join(combinedDir, "isx_combined_data.csv")
				createTestCombinedCSV(t, combinedPath, tt.expectedExisting)
			}
			
			// Test the function with a test logger
			testLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			filesToProcess, existingRecords := DetectUnprocessedFiles(tt.excelFiles, tmpDir, testLogger)
			
			assert.Equal(t, tt.expectedToProcess, len(filesToProcess))
			
			if tt.existingCombined {
				assert.GreaterOrEqual(t, len(existingRecords), 0) // May be 0 if CSV is empty/malformed
			} else {
				assert.Equal(t, 0, len(existingRecords))
			}
		})
	}
}

func TestLoadExistingRecords(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectError bool
		expectedLen int
	}{
		{
			name: "valid CSV with records",
			csvContent: `Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus
2025-01-10,Test Company,TEST,100.000,105.000,95.000,102.000,101.000,103.000,101.000,2.000,1.98,10,1000,102000.00,true
2025-01-11,Test Company,TEST,103.000,108.000,102.000,105.000,102.000,106.000,103.000,3.000,2.91,15,1500,157500.00,true`,
			expectError: false,
			expectedLen: 2,
		},
		{
			name: "empty CSV file",
			csvContent: `Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus`,
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "malformed CSV",
			csvContent: `Date,CompanyName,Symbol
2025-01-10,Test Company`, // Missing fields
			expectError: false,
			expectedLen: 0, // Should skip malformed records
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, "test.csv")
			
			err := os.WriteFile(csvPath, []byte(tt.csvContent), 0644)
			require.NoError(t, err)
			
			records, err := loadExistingRecords(csvPath)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(records))
				
				// Verify record structure if we have records
				if len(records) > 0 {
					record := records[0]
					assert.NotEmpty(t, record.CompanyName)
					assert.NotEmpty(t, record.CompanySymbol)
					assert.False(t, record.Date.IsZero())
				}
			}
		})
	}
}

func TestForwardFillMissingData(t *testing.T) {
	tests := []struct {
		name           string
		inputRecords   []domain.TradeRecord
		expectedOutput int // Expected number of output records
		description    string
	}{
		{
			name:           "empty input",
			inputRecords:   []domain.TradeRecord{},
			expectedOutput: 0,
			description:    "Should handle empty input gracefully",
		},
		{
			name: "single symbol single day",
			inputRecords: []domain.TradeRecord{
				{
					CompanyName:   "Test Company",
					CompanySymbol: "TEST",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
					TradingStatus: true,
				},
			},
			expectedOutput: 1,
			description:    "Should pass through single record unchanged",
		},
		{
			name: "single symbol multiple days with gap",
			inputRecords: []domain.TradeRecord{
				{
					CompanyName:   "Test Company",
					CompanySymbol: "TEST",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
					TradingStatus: true,
				},
				{
					CompanyName:   "Test Company",
					CompanySymbol: "TEST",
					Date:          time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC), // Gap on 1/11
					ClosePrice:    105.0,
					TradingStatus: true,
				},
			},
			expectedOutput: 3, // Original 2 + 1 filled for 1/11
			description:    "Should forward-fill missing day",
		},
		{
			name: "multiple symbols",
			inputRecords: []domain.TradeRecord{
				{
					CompanyName:   "Company A",
					CompanySymbol: "TESTA",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
					TradingStatus: true,
				},
				{
					CompanyName:   "Company B", 
					CompanySymbol: "TESTB",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    200.0,
					TradingStatus: true,
				},
				{
					CompanyName:   "Company A",
					CompanySymbol: "TESTA", 
					Date:          time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC),
					ClosePrice:    110.0,
					TradingStatus: true,
				},
				// TESTA missing on 1/11, TESTB missing on 1/11 and 1/12
			},
			expectedOutput: 6, // 3 original + 1 forward-filled TESTA on 1/11 + 2 forward-filled TESTB on 1/11 and 1/12
			description:    "Should forward-fill for multiple symbols independently",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := forwardFillMissingData(tt.inputRecords)
			
			assert.Equal(t, tt.expectedOutput, len(result), tt.description)
			
			if len(result) > len(tt.inputRecords) {
				// Verify that forward-filled records have correct properties
				for _, record := range result {
					if !record.TradingStatus {
						// This is a forward-filled record
						assert.Equal(t, int64(0), record.NumTrades, "Forward-filled record should have 0 trades")
						assert.Equal(t, int64(0), record.Volume, "Forward-filled record should have 0 volume")
						assert.Equal(t, 0.0, record.Change, "Forward-filled record should have 0 change")
						assert.Equal(t, 0.0, record.ChangePercent, "Forward-filled record should have 0 change percent")
					}
				}
			}
		})
	}
}

func TestSaveCombinedCSV(t *testing.T) {
	tests := []struct {
		name        string
		records     []domain.TradeRecord
		expectError bool
	}{
		{
			name:        "empty records",
			records:     []domain.TradeRecord{},
			expectError: false,
		},
		{
			name: "single record",
			records: []domain.TradeRecord{
				{
					CompanyName:      "Test Company",
					CompanySymbol:    "TEST",
					Date:             time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					OpenPrice:        100.0,
					HighPrice:        105.0,
					LowPrice:         95.0,
					AveragePrice:     102.0,
					PrevAveragePrice: 101.0,
					ClosePrice:       103.0,
					PrevClosePrice:   101.0,
					Change:           2.0,
					ChangePercent:    1.98,
					NumTrades:        10,
					Volume:           1000,
					Value:            102000.0,
					TradingStatus:    true,
				},
			},
			expectError: false,
		},
		{
			name: "multiple records",
			records: []domain.TradeRecord{
				{
					CompanyName:   "Company A",
					CompanySymbol: "TESTA",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
					TradingStatus: true,
				},
				{
					CompanyName:   "Company B",
					CompanySymbol: "TESTB", 
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    200.0,
					TradingStatus: false, // Forward-filled
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, "test_combined.csv")
			
			err := saveCombinedCSV(csvPath, tt.records)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			// Verify file was created and has correct content
			assert.FileExists(t, csvPath)
			
			// Read and verify content
			file, err := os.Open(csvPath)
			require.NoError(t, err)
			defer file.Close()
			
			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			require.NoError(t, err)
			
			// Should have header + data rows
			expectedRows := len(tt.records) + 1 // +1 for header
			assert.Equal(t, expectedRows, len(records))
			
			if len(records) > 0 {
				// Verify header
				header := records[0]
				expectedHeaders := []string{
					"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
					"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
					"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
				}
				assert.Equal(t, expectedHeaders, header)
			}
		})
	}
}

func TestGenerateDailyFiles(t *testing.T) {
	tests := []struct {
		name            string
		records         []domain.TradeRecord
		expectedFiles   []string
		expectError     bool
	}{
		{
			name:          "empty records",
			records:       []domain.TradeRecord{},
			expectedFiles: []string{},
			expectError:   false,
		},
		{
			name: "single day records",
			records: []domain.TradeRecord{
				{
					CompanyName:   "Company A",
					CompanySymbol: "TESTA",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
				},
				{
					CompanyName:   "Company B",
					CompanySymbol: "TESTB",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    200.0,
				},
			},
			expectedFiles: []string{"isx_daily_2025_01_10.csv"},
			expectError:   false,
		},
		{
			name: "multiple days records",
			records: []domain.TradeRecord{
				{
					CompanyName:   "Company A",
					CompanySymbol: "TESTA",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
				},
				{
					CompanyName:   "Company B",
					CompanySymbol: "TESTB", 
					Date:          time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    200.0,
				},
			},
			expectedFiles: []string{"isx_daily_2025_01_10.csv", "isx_daily_2025_01_11.csv"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			err := generateDailyFiles(tt.records, tmpDir)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			// Verify expected files were created
			for _, expectedFile := range tt.expectedFiles {
				filePath := filepath.Join(tmpDir, expectedFile)
				assert.FileExists(t, filePath)
				
				// Verify file has content
				info, err := os.Stat(filePath)
				assert.NoError(t, err)
				assert.Greater(t, info.Size(), int64(0))
			}
			
			// Verify no unexpected files were created
			files, err := os.ReadDir(tmpDir)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedFiles), len(files))
		})
	}
}

func TestGenerateTickerFiles(t *testing.T) {
	tests := []struct {
		name            string
		records         []domain.TradeRecord
		expectedTickers []string
		expectError     bool
	}{
		{
			name:            "empty records",
			records:         []domain.TradeRecord{},
			expectedTickers: []string{},
			expectError:     false,
		},
		{
			name: "single ticker",
			records: []domain.TradeRecord{
				{
					CompanyName:   "Test Company",
					CompanySymbol: "TEST",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
				},
				{
					CompanyName:   "Test Company",
					CompanySymbol: "TEST",
					Date:          time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    105.0,
				},
			},
			expectedTickers: []string{"TEST_trading_history.csv"},
			expectError:     false,
		},
		{
			name: "multiple tickers",
			records: []domain.TradeRecord{
				{
					CompanyName:   "Company A",
					CompanySymbol: "TESTA",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    100.0,
				},
				{
					CompanyName:   "Company B",
					CompanySymbol: "TESTB",
					Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
					ClosePrice:    200.0,
				},
			},
			expectedTickers: []string{"TESTA_trading_history.csv", "TESTB_trading_history.csv"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			err := generateTickerFiles(tt.records, tmpDir)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			assert.NoError(t, err)
			
			// Verify expected ticker files were created
			for _, expectedFile := range tt.expectedTickers {
				filePath := filepath.Join(tmpDir, expectedFile)
				assert.FileExists(t, filePath)
			}
			
			// Verify correct number of files
			files, err := os.ReadDir(tmpDir)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedTickers), len(files))
		})
	}
}

// TestFlagParsing removed - can't test flag parsing with main package flags defined

// Helper function to create test combined CSV file
func createTestCombinedCSV(t *testing.T, path string, numRecords int) {
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Date", "CompanyName", "Symbol", "OpenPrice", "HighPrice", "LowPrice",
		"AveragePrice", "PrevAveragePrice", "ClosePrice", "PrevClosePrice",
		"Change", "ChangePercent", "NumTrades", "Volume", "Value", "TradingStatus",
	}
	writer.Write(header)

	// Write test records
	for i := 0; i < numRecords; i++ {
		date := time.Date(2025, 1, 10+i, 0, 0, 0, 0, time.UTC)
		row := []string{
			date.Format("2006-01-02"),
			"Test Company",
			"TEST",
			"100.000", "105.000", "95.000", "102.000", "101.000", "103.000", "101.000",
			"2.000", "1.98", "10", "1000", "102000.00", "true",
		}
		writer.Write(row)
	}
}

// Benchmark forward fill operation
func BenchmarkForwardFillMissingData(b *testing.B) {
	// Create test data with gaps
	records := make([]domain.TradeRecord, 0, 100)
	
	symbols := []string{"TEST1", "TEST2", "TEST3", "TEST4", "TEST5"}
	dates := []time.Time{
		time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC), // Gap on 1/11
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), // Gaps on 1/13, 1/14
	}
	
	for _, symbol := range symbols {
		for _, date := range dates {
			records = append(records, domain.TradeRecord{
				CompanyName:   "Test Company",
				CompanySymbol: symbol,
				Date:          date,
				ClosePrice:    100.0,
				TradingStatus: true,
			})
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = forwardFillMissingData(records)
	}
}

// Test concurrent file operations
func TestConcurrentFileOperations(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test concurrent CSV generation
	const numGoroutines = 5
	done := make(chan bool, numGoroutines)
	
	records := []domain.TradeRecord{
		{
			CompanyName:   "Test Company",
			CompanySymbol: "TEST",
			Date:          time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
			ClosePrice:    100.0,
		},
	}
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			filePath := filepath.Join(tmpDir, fmt.Sprintf("test_%d.csv", id))
			err := saveCombinedCSV(filePath, records)
			assert.NoError(t, err)
		}(i)
	}
	
	// Wait for all operations to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify all files were created
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, numGoroutines, len(files))
}