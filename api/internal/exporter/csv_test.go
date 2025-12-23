package exporter

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isxcli/isxcli/internal/config"
)

// MockPaths implements config.Paths interface for testing
type MockPaths struct {
	basePath string
}

func NewMockPaths(basePath string) *MockPaths {
	return &MockPaths{basePath: basePath}
}

func (m *MockPaths) GetReportPath(filename string) string {
	return filepath.Join(m.basePath, "reports", filename)
}

func (m *MockPaths) GetDownloadPath(filename string) string {
	return filepath.Join(m.basePath, "downloads", filename)
}

func (m *MockPaths) GetCachePath(filename string) string {
	return filepath.Join(m.basePath, "cache", filename)
}

// Setup test environment
func setupTestEnv(t *testing.T) (*CSVWriter, string, func()) {
	t.Helper()
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "exporter_test_*")
	require.NoError(t, err)
	
	// Create subdirectories
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "downloads"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "cache"), 0755))
	
	// Create CSV writer
	writer := NewCSVWriter(&config.Paths{
		ReportsDir:   filepath.Join(tempDir, "reports"),
		DownloadsDir: filepath.Join(tempDir, "downloads"),
		CacheDir:     filepath.Join(tempDir, "cache"),
	})
	
	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return writer, tempDir, cleanup
}

func TestNewCSVWriter(t *testing.T) {
	paths := &config.Paths{}
	writer := NewCSVWriter(paths)
	
	assert.NotNil(t, writer)
	assert.Equal(t, paths, writer.paths)
}

func TestCSVWriter_WriteCSV(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name        string
		filePath    string
		options     WriteOptions
		expectError bool
		validate    func(t *testing.T, filePath string)
	}{
		{
			name:     "basic write with headers",
			filePath: "test_basic.csv",
			options: WriteOptions{
				Headers: []string{"Name", "Age", "City"},
				Records: [][]string{
					{"John", "25", "New York"},
					{"Jane", "30", "London"},
				},
				Append:    false,
				BOMPrefix: false,
			},
			expectError: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				assert.Len(t, lines, 3) // header + 2 records
				assert.Equal(t, "Name,Age,City", lines[0])
				assert.Equal(t, "John,25,New York", lines[1])
				assert.Equal(t, "Jane,30,London", lines[2])
			},
		},
		{
			name:     "write with BOM prefix",
			filePath: "test_bom.csv",
			options: WriteOptions{
				Headers: []string{"Symbol", "Price"},
				Records: [][]string{
					{"AAPL", "150.25"},
				},
				Append:    false,
				BOMPrefix: true,
			},
			expectError: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				// Check for UTF-8 BOM
				assert.True(t, bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}))
				
				// Remove BOM and check content
				contentWithoutBOM := content[3:]
				lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
				assert.Equal(t, "Symbol,Price", lines[0])
				assert.Equal(t, "AAPL,150.25", lines[1])
			},
		},
		{
			name:     "write without headers",
			filePath: "test_no_headers.csv",
			options: WriteOptions{
				Headers: nil,
				Records: [][]string{
					{"Data1", "Data2"},
					{"Data3", "Data4"},
				},
				Append:    false,
				BOMPrefix: false,
			},
			expectError: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				assert.Len(t, lines, 2) // only records, no headers
				assert.Equal(t, "Data1,Data2", lines[0])
				assert.Equal(t, "Data3,Data4", lines[1])
			},
		},
		{
			name:     "append to existing file",
			filePath: "test_append.csv",
			options: WriteOptions{
				Records: [][]string{
					{"AppendedData1", "AppendedData2"},
				},
				Append:    true,
				BOMPrefix: false,
			},
			expectError: false,
			validate: func(t *testing.T, filePath string) {
				// This test will be run after creating the initial file
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				// Should contain both original and appended data
				assert.Contains(t, string(content), "AppendedData1,AppendedData2")
			},
		},
		{
			name:     "empty records",
			filePath: "test_empty.csv",
			options: WriteOptions{
				Headers: []string{"Col1", "Col2"},
				Records: [][]string{},
				Append:    false,
				BOMPrefix: false,
			},
			expectError: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				lines := strings.Split(strings.TrimSpace(string(content)), "\n")
				assert.Len(t, lines, 1) // only headers
				assert.Equal(t, "Col1,Col2", lines[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(tempDir, "reports", tt.filePath)
			
			// For append test, create initial file first
			if tt.name == "append to existing file" {
				initialOptions := WriteOptions{
					Headers: []string{"Initial1", "Initial2"},
					Records: [][]string{{"InitData1", "InitData2"}},
					Append:    false,
					BOMPrefix: false,
				}
				err := writer.WriteCSV(tt.filePath, initialOptions)
				require.NoError(t, err)
			}
			
			err := writer.WriteCSV(tt.filePath, tt.options)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validate(t, fullPath)
			}
		})
	}
}

func TestCSVWriter_WriteSimpleCSV(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	headers := []string{"Company", "Symbol", "Price"}
	records := [][]string{
		{"Apple Inc", "AAPL", "150.25"},
		{"Microsoft Corp", "MSFT", "280.75"},
	}

	err := writer.WriteSimpleCSV("simple_test.csv", headers, records)
	assert.NoError(t, err)

	// Validate file content
	filePath := filepath.Join(tempDir, "reports", "simple_test.csv")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Check for BOM (WriteSimpleCSV uses BOMPrefix: true)
	assert.True(t, bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}))

	// Remove BOM and check content
	contentWithoutBOM := content[3:]
	lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
	assert.Len(t, lines, 3) // header + 2 records
	assert.Equal(t, "Company,Symbol,Price", lines[0])
	assert.Equal(t, "Apple Inc,AAPL,150.25", lines[1])
	assert.Equal(t, "Microsoft Corp,MSFT,280.75", lines[2])
}

func TestCSVWriter_AppendToCSV(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	filePath := "append_test.csv"
	fullPath := filepath.Join(tempDir, "reports", filePath)

	// Create initial file
	initialRecords := [][]string{
		{"Initial1", "Initial2"},
		{"Data1", "Data2"},
	}
	err := writer.WriteSimpleCSV(filePath, []string{"Col1", "Col2"}, initialRecords)
	require.NoError(t, err)

	// Append new records
	appendRecords := [][]string{
		{"Appended1", "Appended2"},
		{"NewData1", "NewData2"},
	}
	err = writer.AppendToCSV(filePath, appendRecords)
	assert.NoError(t, err)

	// Validate content
	content, err := os.ReadFile(fullPath)
	require.NoError(t, err)

	// Remove BOM for easier parsing
	contentWithoutBOM := content[3:]
	lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
	
	assert.Len(t, lines, 5) // header + 2 initial + 2 appended
	assert.Equal(t, "Col1,Col2", lines[0])
	assert.Equal(t, "Initial1,Initial2", lines[1])
	assert.Equal(t, "Data1,Data2", lines[2])
	assert.Equal(t, "Appended1,Appended2", lines[3])
	assert.Equal(t, "NewData1,NewData2", lines[4])
}

func TestCSVWriter_ResolvePath(t *testing.T) {
	writer, _, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name          string
		inputPath     string
		expectedSuffix string
		isAbsolute    bool
	}{
		{
			name:          "absolute path",
			inputPath:     `C:\absolute\path\file.csv`,
			expectedSuffix: `C:\absolute\path\file.csv`,
			isAbsolute:    true,
		},
		{
			name:          "downloads path",
			inputPath:     "downloads/report.csv",
			expectedSuffix: filepath.Join("downloads", "report.csv"),
			isAbsolute:    false,
		},
		{
			name:          "cache path",
			inputPath:     "cache/temp.csv",
			expectedSuffix: filepath.Join("cache", "temp.csv"),
			isAbsolute:    false,
		},
		{
			name:          "default to reports",
			inputPath:     "regular_report.csv",
			expectedSuffix: filepath.Join("reports", "regular_report.csv"),
			isAbsolute:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writer.resolvePath(tt.inputPath)
			
			if tt.isAbsolute {
				assert.Equal(t, tt.inputPath, result)
			} else {
				assert.True(t, strings.HasSuffix(result, tt.expectedSuffix))
			}
		})
	}
}

func TestCSVWriter_SpecialCharacters(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test with special characters that need CSV escaping
	headers := []string{"Name", "Description", "Notes"}
	records := [][]string{
		{"Company, Inc", "Description with \"quotes\"", "Notes with\nnewlines"},
		{"Åpple", "Émojis: 😀🚀", "Special chars: ñáéíóú"},
		{"Company;With;Semicolons", "Text,with,commas", "Text\twith\ttabs"},
	}

	err := writer.WriteSimpleCSV("special_chars.csv", headers, records)
	assert.NoError(t, err)

	// Read back and parse to verify CSV escaping worked correctly
	filePath := filepath.Join(tempDir, "reports", "special_chars.csv")
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

	assert.Len(t, allRecords, 4) // header + 3 records

	// Verify headers
	assert.Equal(t, headers, allRecords[0])

	// Verify first record with special characters
	assert.Equal(t, "Company, Inc", allRecords[1][0])
	assert.Equal(t, "Description with \"quotes\"", allRecords[1][1])
	assert.Equal(t, "Notes with\nnewlines", allRecords[1][2])

	// Verify Unicode characters
	assert.Equal(t, "Åpple", allRecords[2][0])
	assert.Equal(t, "Émojis: 😀🚀", allRecords[2][1])
	assert.Equal(t, "Special chars: ñáéíóú", allRecords[2][2])
}

func TestCSVWriter_ConcurrentWrites(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	const numGoroutines = 10
	const recordsPerGoroutine = 100

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	// Test concurrent writes to different files
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			filePath := filepath.Join("concurrent", "file_"+string(rune('A'+id))+".csv")
			
			var records [][]string
			for j := 0; j < recordsPerGoroutine; j++ {
				records = append(records, []string{
					"Record" + string(rune('A'+id)),
					string(rune('0' + j%10)),
				})
			}
			
			err := writer.WriteSimpleCSV(filePath, []string{"Name", "Number"}, records)
			if err != nil {
				errChan <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		assert.NoError(t, err)
	}

	// Verify all files were created correctly
	for i := 0; i < numGoroutines; i++ {
		filePath := filepath.Join(tempDir, "reports", "concurrent", "file_"+string(rune('A'+i))+".csv")
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "File %s should exist", filePath)
		
		// Verify content
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		
		// Remove BOM and count lines
		contentWithoutBOM := content[3:]
		lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
		assert.Len(t, lines, recordsPerGoroutine+1) // header + records
	}
}

func TestCSVWriter_ErrorScenarios(t *testing.T) {
	// Test with invalid paths configuration
	invalidPaths := &config.Paths{
		ReportsDir: `C:\invalid\path\that\cannot\be\created\due\to\very\long\path\and\permissions`,
	}
	writer := NewCSVWriter(invalidPaths)

	options := WriteOptions{
		Headers: []string{"Test"},
		Records: [][]string{{"Data"}},
	}

	// This should fail due to path issues
	err := writer.WriteCSV("test.csv", options)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to")
	} else {
		t.Skip("Could not create error scenario on this system")
	}
}

// BenchmarkCSVWriter_WriteCSV tests CSV writing performance
func BenchmarkCSVWriter_WriteCSV(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "benchmark_csv_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	writer := NewCSVWriter(&config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	})

	// Create test data
	headers := []string{"Col1", "Col2", "Col3", "Col4", "Col5"}
	var records [][]string
	for i := 0; i < 1000; i++ {
		records = append(records, []string{
			"Data" + string(rune(i%26+'A')),
			"Value" + string(rune(i%10+'0')),
			"Text" + string(rune(i%26+'A')),
			"Number" + string(rune(i%10+'0')),
			"Field" + string(rune(i%26+'A')),
		})
	}

	options := WriteOptions{
		Headers:   headers,
		Records:   records,
		Append:    false,
		BOMPrefix: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := "benchmark_" + string(rune(i%26+'A')) + ".csv"
		err := writer.WriteCSV(filePath, options)
		require.NoError(b, err)
	}
}

// BenchmarkCSVWriter_LargeDataset tests performance with very large datasets
func BenchmarkCSVWriter_LargeDataset(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_large_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	writer := NewCSVWriter(&config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	})

	// Create larger test dataset
	headers := []string{"ID", "Name", "Price", "Volume", "Date", "Status", "Change", "Percent"}
	var records [][]string
	for i := 0; i < 10000; i++ {
		records = append(records, []string{
			"ID" + string(rune(i%1000)),
			"Company" + string(rune(i%100+'A')),
			"123.45",
			"1000000",
			"2024-01-01",
			"active",
			"5.25",
			"2.5%",
		})
	}

	options := WriteOptions{
		Headers:   headers,
		Records:   records,
		Append:    false,
		BOMPrefix: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := writer.WriteCSV("large_dataset.csv", options)
		require.NoError(b, err)
	}
}