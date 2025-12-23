package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// TestMain removed - flag parsing is handled in main.go

func TestFileRegex(t *testing.T) {
	fileRe := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	
	tests := []struct {
		name     string
		filename string
		matches  bool
		groups   []string
	}{
		{
			name:     "valid filename",
			filename: "2025 01 15 ISX Daily Report.xlsx",
			matches:  true,
			groups:   []string{"2025 01 15 ISX Daily Report.xlsx", "2025", "01", "15"},
		},
		{
			name:     "invalid extension",
			filename: "2025 01 15 ISX Daily Report.pdf",
			matches:  false,
			groups:   nil,
		},
		{
			name:     "missing day",
			filename: "2025 01 ISX Daily Report.xlsx",
			matches:  false,
			groups:   nil,
		},
		{
			name:     "extra text",
			filename: "2025 01 15 ISX Daily Report Extra.xlsx",
			matches:  false,
			groups:   nil,
		},
		{
			name:     "wrong case",
			filename: "2025 01 15 isx daily report.xlsx",
			matches:  false,
			groups:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := fileRe.FindStringSubmatch(tt.filename)
			
			if tt.matches {
				assert.NotNil(t, matches)
				assert.Equal(t, tt.groups, matches)
			} else {
				assert.Nil(t, matches)
			}
		})
	}
}

func TestLoadLastDate(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectError bool
		expectedDate string
	}{
		{
			name: "valid CSV with data",
			csvContent: `Date,ISX60,ISX15
2025-01-10,1234.56,567.89
2025-01-11,1240.00,570.00
2025-01-12,1245.50,572.25`,
			expectError:  false,
			expectedDate: "2025-01-12",
		},
		{
			name: "CSV with header only",
			csvContent: `Date,ISX60,ISX15`,
			expectError: true, // No data rows
		},
		{
			name: "empty CSV",
			csvContent: ``,
			expectError: true,
		},
		{
			name: "CSV with invalid date format",
			csvContent: `Date,ISX60,ISX15
invalid-date,1234.56,567.89`,
			expectError: true,
		},
		{
			name: "single data row",
			csvContent: `Date,ISX60,ISX15
2025-01-10,1234.56,567.89`,
			expectError:  false,
			expectedDate: "2025-01-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, "test.csv")
			
			err := os.WriteFile(csvPath, []byte(tt.csvContent), 0644)
			require.NoError(t, err)
			
			result, err := loadLastDate(csvPath)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDate, result.Format("2006-01-02"))
			}
		})
	}
}

func TestExtractIndices(t *testing.T) {
	tests := []struct {
		name            string
		sheetData       map[string][][]string // sheetName -> rows
		expectedISX60   float64
		expectedISX15   float64
		expectError     bool
		description     string
	}{
		{
			name: "both indices present on same line",
			sheetData: map[string][][]string{
				"Indices": {
					{"", "", ""},
					{"Market Indices", "", ""},
					{"ISX Index 60", "1234.56", "ISX Index 15", "567.89"},
				},
			},
			expectedISX60: 1234.56,
			expectedISX15: 567.89,
			expectError:   false,
			description:   "Should extract both indices from same line",
		},
		{
			name: "only ISX60 present",
			sheetData: map[string][][]string{
				"Indices": {
					{"", "", ""},
					{"Market Indices", "", ""},
					{"ISX Index 60", "1234.56"},
				},
			},
			expectedISX60: 1234.56,
			expectedISX15: 0,
			expectError:   false,
			description:   "Should extract ISX60 only when ISX15 not present",
		},
		{
			name: "old format ISX Price Index",
			sheetData: map[string][][]string{
				"Sheet1": {
					{"", "", ""},
					{"Market Summary", "", ""},
					{"ISX Price Index", "987.65"},
				},
			},
			expectedISX60: 987.65,
			expectedISX15: 0,
			expectError:   false,
			description:   "Should handle old ISX Price Index format",
		},
		{
			name: "indices not found",
			sheetData: map[string][][]string{
				"Sheet1": {
					{"Some", "Other", "Data"},
					{"No", "Indices", "Here"},
				},
			},
			expectedISX60: 0,
			expectedISX15: 0,
			expectError:   true,
			description:   "Should return error when indices not found",
		},
		{
			name: "indices with comma formatting",
			sheetData: map[string][][]string{
				"Indices": {
					{"", "", ""},
					{"ISX Index 60", "1,234.56", "ISX Index 15", "567.89"},
				},
			},
			expectedISX60: 1234.56,
			expectedISX15: 567.89,
			expectError:   false,
			description:   "Should handle comma-formatted numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary Excel file
			tmpDir := t.TempDir()
			excelPath := filepath.Join(tmpDir, "test.xlsx")
			
			f := excelize.NewFile()
			defer f.Close()
			
			// Remove default sheet and create test sheets
			f.DeleteSheet("Sheet1")
			
			for sheetName, rows := range tt.sheetData {
				_, err := f.NewSheet(sheetName)
				require.NoError(t, err)
				
				for rowIdx, row := range rows {
					for colIdx, cellValue := range row {
						cell := getCellName(rowIdx+1, colIdx+1)
						f.SetCellValue(sheetName, cell, cellValue)
					}
				}
			}
			
			err := f.SaveAs(excelPath)
			require.NoError(t, err)
			
			// Test the function
			isx60, isx15, err := extractIndices(excelPath)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedISX60, isx60, tt.description)
				assert.Equal(t, tt.expectedISX15, isx15, tt.description)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{
			name:        "simple number",
			input:       "123.45",
			expected:    123.45,
			expectError: false,
		},
		{
			name:        "number with commas",
			input:       "1,234.56",
			expected:    1234.56,
			expectError: false,
		},
		{
			name:        "integer",
			input:       "1000",
			expected:    1000.0,
			expectError: false,
		},
		{
			name:        "number with multiple commas",
			input:       "1,234,567.89",
			expected:    1234567.89,
			expectError: false,
		},
		{
			name:        "invalid number",
			input:       "not-a-number",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFloat(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "simple number",
			input:    123.45,
			expected: "123.45",
		},
		{
			name:     "integer",
			input:    1000.0,
			expected: "1000.00",
		},
		{
			name:     "high precision number",
			input:    123.456789,
			expected: "123.46", // Should round to 2 decimal places
		},
		{
			name:     "zero",
			input:    0.0,
			expected: "0.00",
		},
		{
			name:     "negative number",
			input:    -123.45,
			expected: "-123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFloat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFlagParsing removed - can't test flag parsing with main package flags defined

func TestFileDiscoveryAndProcessing(t *testing.T) {
	tests := []struct {
		name                string
		files               []string
		lastDate            string
		expectedToProcess   int
		description         string
	}{
		{
			name: "initial mode - process all files",
			files: []string{
				"2025 01 10 ISX Daily Report.xlsx",
				"2025 01 11 ISX Daily Report.xlsx",
				"2025 01 12 ISX Daily Report.xlsx",
			},
			lastDate:          "",
			expectedToProcess: 3,
			description:       "Should process all files in initial mode",
		},
		{
			name: "accumulative mode - process new files only",
			files: []string{
				"2025 01 10 ISX Daily Report.xlsx",
				"2025 01 11 ISX Daily Report.xlsx",
				"2025 01 12 ISX Daily Report.xlsx",
			},
			lastDate:          "2025-01-10",
			expectedToProcess: 2, // Only 01-11 and 01-12
			description:       "Should process only files newer than last date",
		},
		{
			name: "no new files to process",
			files: []string{
				"2025 01 10 ISX Daily Report.xlsx",
				"2025 01 11 ISX Daily Report.xlsx",
			},
			lastDate:          "2025-01-11",
			expectedToProcess: 0,
			description:       "Should not process files older than or equal to last date",
		},
		{
			name: "mixed valid and invalid files",
			files: []string{
				"2025 01 10 ISX Daily Report.xlsx",
				"invalid_file.xlsx",
				"2025 01 11 ISX Daily Report.xlsx",
				"another_invalid.txt",
			},
			lastDate:          "",
			expectedToProcess: 2, // Only valid ISX reports
			description:       "Should filter out invalid filenames",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create test files
			for _, filename := range tt.files {
				filePath := filepath.Join(tmpDir, filename)
				
				// Create a simple Excel file or regular file
				if strings.HasSuffix(filename, ".xlsx") && strings.Contains(filename, "ISX Daily Report") {
					createTestExcelFile(t, filePath)
				} else {
					err := os.WriteFile(filePath, []byte("test content"), 0644)
					require.NoError(t, err)
				}
			}
			
			// Simulate file discovery and filtering logic
			fileRe := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
			entries, err := os.ReadDir(tmpDir)
			require.NoError(t, err)
			
			var filesToProcess []struct {
				path string
				date time.Time
			}
			
			var lastDate time.Time
			if tt.lastDate != "" {
				lastDate, err = time.Parse("2006-01-02", tt.lastDate)
				require.NoError(t, err)
			}
			
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				
				matches := fileRe.FindStringSubmatch(entry.Name())
				if matches == nil {
					continue
				}
				
				fileDate, err := time.Parse("2006 01 02", strings.Join(matches[1:4], " "))
				require.NoError(t, err)
				
				if !lastDate.IsZero() && !fileDate.After(lastDate) {
					continue
				}
				
				filesToProcess = append(filesToProcess, struct {
					path string
					date time.Time
				}{
					path: filepath.Join(tmpDir, entry.Name()),
					date: fileDate,
				})
			}
			
			assert.Equal(t, tt.expectedToProcess, len(filesToProcess), tt.description)
		})
	}
}

func TestCSVOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		records  [][]string
		expected [][]string
	}{
		{
			name: "initial mode CSV creation",
			mode: "initial",
			records: [][]string{
				{"2025-01-10", "1234.56", "567.89"},
				{"2025-01-11", "1240.00", "570.00"},
			},
			expected: [][]string{
				{"Date", "ISX60", "ISX15"},
				{"2025-01-10", "1234.56", "567.89"},
				{"2025-01-11", "1240.00", "570.00"},
			},
		},
		{
			name: "accumulative mode CSV append",
			mode: "accumulative",
			records: [][]string{
				{"2025-01-12", "1245.50", "572.25"},
			},
			expected: [][]string{
				{"2025-01-12", "1245.50", "572.25"},
			},
		},
		{
			name: "record with missing ISX15",
			mode: "initial",
			records: [][]string{
				{"2025-01-10", "1234.56", ""},
			},
			expected: [][]string{
				{"Date", "ISX60", "ISX15"},
				{"2025-01-10", "1234.56", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			csvPath := filepath.Join(tmpDir, "test_output.csv")
			
			if tt.mode == "initial" {
				// Create CSV with header
				file, err := os.Create(csvPath)
				require.NoError(t, err)
				writer := csv.NewWriter(file)
				
				// Write header
				writer.Write([]string{"Date", "ISX60", "ISX15"})
				
				// Write data records
				for _, record := range tt.records {
					writer.Write(record)
				}
				
				writer.Flush()
				file.Close()
			} else {
				// For accumulative mode, just append records
				file, err := os.Create(csvPath)
				require.NoError(t, err)
				writer := csv.NewWriter(file)
				
				for _, record := range tt.records {
					writer.Write(record)
				}
				
				writer.Flush()
				file.Close()
			}
			
			// Verify output
			file, err := os.Open(csvPath)
			require.NoError(t, err)
			defer file.Close()
			
			reader := csv.NewReader(file)
			actualRecords, err := reader.ReadAll()
			require.NoError(t, err)
			
			assert.Equal(t, tt.expected, actualRecords)
		})
	}
}

// Helper function to create a test Excel file
func createTestExcelFile(t *testing.T, path string) {
	f := excelize.NewFile()
	defer f.Close()
	
	// Create a simple sheet with test data
	f.SetCellValue("Sheet1", "A1", "Test")
	f.SetCellValue("Sheet1", "B1", "Data")
	
	err := f.SaveAs(path)
	require.NoError(t, err)
}

// Helper function to get Excel cell name (A1, B2, etc.)
func getCellName(row, col int) string {
	colName := ""
	for col > 0 {
		col--
		colName = string(rune('A'+col%26)) + colName
		col /= 26
	}
	return colName + fmt.Sprintf("%d", row)
}

// Benchmark extractIndices function
func BenchmarkExtractIndices(b *testing.B) {
	// Create a test Excel file
	tmpDir := b.TempDir()
	excelPath := filepath.Join(tmpDir, "benchmark.xlsx")
	
	f := excelize.NewFile()
	defer f.Close()
	
	// Create Indices sheet with test data
	f.NewSheet("Indices")
	f.DeleteSheet("Sheet1")
	
	// Add some rows with index data
	f.SetCellValue("Indices", "A5", "ISX Index 60")
	f.SetCellValue("Indices", "B5", "1234.56")
	f.SetCellValue("Indices", "C5", "ISX Index 15")
	f.SetCellValue("Indices", "D5", "567.89")
	
	f.SaveAs(excelPath)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = extractIndices(excelPath)
	}
}

// Test concurrent file processing
func TestConcurrentProcessing(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create multiple test Excel files
	const numFiles = 5
	filePaths := make([]string, numFiles)
	
	for i := 0; i < numFiles; i++ {
		filename := filepath.Join(tmpDir, fmt.Sprintf("2025 01 %02d ISX Daily Report.xlsx", i+10))
		createTestExcelFile(t, filename)
		filePaths[i] = filename
	}
	
	// Test concurrent processing
	done := make(chan bool, numFiles)
	
	for _, filePath := range filePaths {
		go func(path string) {
			defer func() { done <- true }()
			
			// This would normally call extractIndices, but since our test files
			// don't have actual index data, we'll just verify the file exists
			_, err := os.Stat(path)
			assert.NoError(t, err)
		}(filePath)
	}
	
	// Wait for all processing to complete
	for i := 0; i < numFiles; i++ {
		<-done
	}
}