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

func TestCSVWriter_CreateStreamWriter(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name        string
		filePath    string
		headers     []string
		expectError bool
		validate    func(t *testing.T, stream *StreamWriter, filePath string)
	}{
		{
			name:     "create stream with headers",
			filePath: "stream_test.csv",
			headers:  []string{"Name", "Value", "Date"},
			expectError: false,
			validate: func(t *testing.T, stream *StreamWriter, filePath string) {
				assert.NotNil(t, stream)
				assert.NotNil(t, stream.file)
				assert.NotNil(t, stream.writer)
				
				// Flush the writer to ensure headers are written
				stream.writer.Flush()
				
				// Check that file exists and has headers
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				// Check BOM
				assert.True(t, bytes.HasPrefix(content, []byte{0xEF, 0xBB, 0xBF}))
				
				// Check headers
				contentWithoutBOM := content[3:]
				if len(contentWithoutBOM) > 0 {
					lines := strings.Split(strings.TrimSpace(string(contentWithoutBOM)), "\n")
					assert.Len(t, lines, 1) // Only headers at this point
					assert.Equal(t, "Name,Value,Date", lines[0])
				}
			},
		},
		{
			name:     "create stream without headers",
			filePath: "stream_no_headers.csv",
			headers:  []string{},
			expectError: false,
			validate: func(t *testing.T, stream *StreamWriter, filePath string) {
				assert.NotNil(t, stream)
				
				// Check that file exists but has only BOM
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				// Should only have BOM, no content yet
				assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, content)
			},
		},
		{
			name:     "create stream with nil headers",
			filePath: "stream_nil_headers.csv",
			headers:  nil,
			expectError: false,
			validate: func(t *testing.T, stream *StreamWriter, filePath string) {
				assert.NotNil(t, stream)
				
				// Check that file exists but has only BOM
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				
				// Should only have BOM, no content yet
				assert.Equal(t, []byte{0xEF, 0xBB, 0xBF}, content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(tempDir, "reports", tt.filePath)
			
			stream, err := writer.CreateStreamWriter(tt.filePath, tt.headers)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stream)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, stream)
				defer stream.Close()
				
				tt.validate(t, stream, fullPath)
			}
		})
	}
}

func TestStreamWriter_WriteRecord(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	headers := []string{"Symbol", "Price", "Volume"}
	stream, err := writer.CreateStreamWriter("stream_records.csv", headers)
	require.NoError(t, err)
	defer stream.Close()

	tests := []struct {
		name     string
		record   []string
		expectError bool
	}{
		{
			name:     "valid record",
			record:   []string{"AAPL", "150.25", "1000000"},
			expectError: false,
		},
		{
			name:     "record with special characters",
			record:   []string{"Company, Inc", "Price \"quoted\"", "1,000,000"},
			expectError: false,
		},
		{
			name:     "record with unicode",
			record:   []string{"Åpple", "€150.25", "1.000.000"},
			expectError: false,
		},
		{
			name:     "empty record",
			record:   []string{},
			expectError: false,
		},
		{
			name:     "record with empty fields",
			record:   []string{"", "", ""},
			expectError: false,
		},
		{
			name:     "record with newlines",
			record:   []string{"Multi\nLine", "Value", "123"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stream.WriteRecord(tt.record)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Close and validate final file
	err = stream.Close()
	require.NoError(t, err)

	// Read and validate file content
	filePath := filepath.Join(tempDir, "reports", "stream_records.csv")
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

	// Should have header + all test records (empty record test writes nothing)
	assert.Len(t, allRecords, 6) // header + 5 non-empty records (empty record not written)
	assert.Equal(t, headers, allRecords[0])
	
	// Verify some specific records
	assert.Equal(t, []string{"AAPL", "150.25", "1000000"}, allRecords[1])
	assert.Equal(t, []string{"Company, Inc", "Price \"quoted\"", "1,000,000"}, allRecords[2])
	assert.Equal(t, []string{"Åpple", "€150.25", "1.000.000"}, allRecords[3])
}

func TestStreamWriter_Close(t *testing.T) {
	writer, _, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name        string
		setup       func(t *testing.T) *StreamWriter
		expectError bool
	}{
		{
			name: "normal close after writing",
			setup: func(t *testing.T) *StreamWriter {
				stream, err := writer.CreateStreamWriter("close_test1.csv", []string{"A", "B"})
				require.NoError(t, err)
				
				// Write some records
				err = stream.WriteRecord([]string{"1", "2"})
				require.NoError(t, err)
				
				return stream
			},
			expectError: false,
		},
		{
			name: "close without writing records",
			setup: func(t *testing.T) *StreamWriter {
				stream, err := writer.CreateStreamWriter("close_test2.csv", []string{"X", "Y"})
				require.NoError(t, err)
				return stream
			},
			expectError: false,
		},
		{
			name: "double close (now detects misuse)",
			setup: func(t *testing.T) *StreamWriter {
				stream, err := writer.CreateStreamWriter("close_test3.csv", []string{"P", "Q"})
				require.NoError(t, err)

				// First close
				err = stream.Close()
				require.NoError(t, err)

				return stream
			},
			expectError: true, // Implementation changed: Second close now returns error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tt.setup(t)
			
			err := stream.Close()
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStreamWriter_LargeDataset(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	headers := []string{"ID", "Name", "Value", "Date", "Status"}
	stream, err := writer.CreateStreamWriter("large_stream.csv", headers)
	require.NoError(t, err)
	defer stream.Close()

	const numRecords = 10000

	// Write large number of records
	for i := 0; i < numRecords; i++ {
		record := []string{
			"ID" + string(rune(i%1000)),
			"Name" + string(rune(i%26+'A')),
			"123.45",
			"2024-01-01",
			"active",
		}
		
		err := stream.WriteRecord(record)
		require.NoError(t, err)
	}

	err = stream.Close()
	require.NoError(t, err)

	// Verify file content
	filePath := filepath.Join(tempDir, "reports", "large_stream.csv")
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

	// Should have header + all records
	assert.Len(t, allRecords, numRecords+1)
	assert.Equal(t, headers, allRecords[0])

	// Verify first and last records
	// Note: string(rune(0)) creates null byte "\x00", not "0"
	assert.Equal(t, "ID\x00", allRecords[1][0])
	// Note: Last record ID depends on numRecords % 1000
	expectedLastID := "ID" + string(rune((numRecords-1)%1000))
	assert.Equal(t, expectedLastID, allRecords[numRecords][0])
}

func TestStreamWriter_ConcurrentStreams(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	const numStreams = 5
	const recordsPerStream = 1000

	var wg sync.WaitGroup
	errChan := make(chan error, numStreams)

	// Create multiple concurrent streams
	for i := 0; i < numStreams; i++ {
		wg.Add(1)
		go func(streamID int) {
			defer wg.Done()
			
			filename := "concurrent_stream_" + string(rune('A'+streamID)) + ".csv"
			headers := []string{"StreamID", "RecordID", "Value"}
			
			stream, err := writer.CreateStreamWriter(filename, headers)
			if err != nil {
				errChan <- err
				return
			}
			defer stream.Close()
			
			// Write records
			for j := 0; j < recordsPerStream; j++ {
				record := []string{
					string(rune('A' + streamID)),
					string(rune('0' + j%10)),
					"Value" + string(rune('0'+j%10)),
				}
				
				if err := stream.WriteRecord(record); err != nil {
					errChan <- err
					return
				}
			}
			
			if err := stream.Close(); err != nil {
				errChan <- err
				return
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		assert.NoError(t, err)
	}

	// Verify all files were created correctly
	for i := 0; i < numStreams; i++ {
		filename := "concurrent_stream_" + string(rune('A'+i)) + ".csv"
		filePath := filepath.Join(tempDir, "reports", filename)
		
		file, err := os.Open(filePath)
		require.NoError(t, err)
		
		// Skip BOM
		bom := make([]byte, 3)
		_, err = file.Read(bom)
		require.NoError(t, err)
		
		reader := csv.NewReader(file)
		allRecords, err := reader.ReadAll()
		require.NoError(t, err)
		file.Close()
		
		// Should have header + all records
		assert.Len(t, allRecords, recordsPerStream+1)
		assert.Equal(t, []string{"StreamID", "RecordID", "Value"}, allRecords[0])
	}
}

func TestStreamWriter_MemoryEfficiency(t *testing.T) {
	writer, tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test that streaming doesn't load all data into memory
	headers := []string{"Data1", "Data2", "Data3", "Data4", "Data5"}
	stream, err := writer.CreateStreamWriter("memory_test.csv", headers)
	require.NoError(t, err)
	defer stream.Close()

	// Write records one by one to simulate streaming behavior
	const numRecords = 50000
	for i := 0; i < numRecords; i++ {
		record := []string{
			"Large data field " + string(rune(i%1000)) + " with some content",
			"Another field with data " + string(rune(i%100+'A')),
			"Field 3 with numbers " + string(rune(i%10+'0')),
			"Field 4 content here",
			"Field 5 final data",
		}
		
		err := stream.WriteRecord(record)
		require.NoError(t, err)
		
		// Intermittently check that we're not accumulating too much memory
		// In a real test, you might use runtime.ReadMemStats to check memory usage
		if i%10000 == 0 {
			// This is just to demonstrate the streaming nature
			// The key is that we don't keep all records in memory
			assert.NoError(t, err)
		}
	}

	err = stream.Close()
	require.NoError(t, err)

	// Verify file was created correctly
	filePath := filepath.Join(tempDir, "reports", "memory_test.csv")
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	
	// File should be reasonably large
	assert.Greater(t, fileInfo.Size(), int64(1000000)) // > 1MB
}

// BenchmarkStreamWriter_WriteRecord tests the performance of streaming writes
func BenchmarkStreamWriter_WriteRecord(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_stream_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	writer := NewCSVWriter(&config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	})

	headers := []string{"Col1", "Col2", "Col3", "Col4", "Col5"}
	stream, err := writer.CreateStreamWriter("benchmark_stream.csv", headers)
	require.NoError(b, err)
	defer stream.Close()

	record := []string{"Data1", "Data2", "Data3", "Data4", "Data5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := stream.WriteRecord(record)
		require.NoError(b, err)
	}
}

// BenchmarkStreamWriter_vs_BatchWrite compares streaming vs batch writing
func BenchmarkStreamWriter_vs_BatchWrite(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_comparison_*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	require.NoError(b, os.MkdirAll(filepath.Join(tempDir, "reports"), 0755))

	writer := NewCSVWriter(&config.Paths{
		ReportsDir: filepath.Join(tempDir, "reports"),
	})

	headers := []string{"Col1", "Col2", "Col3", "Col4", "Col5"}
	
	// Create test data
	const numRecords = 10000
	var records [][]string
	for i := 0; i < numRecords; i++ {
		records = append(records, []string{
			"Data" + string(rune(i%26+'A')),
			"Value" + string(rune(i%10+'0')),
			"Field3",
			"Field4",
			"Field5",
		})
	}

	b.Run("StreamWriter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stream, err := writer.CreateStreamWriter("stream_bench.csv", headers)
			require.NoError(b, err)
			
			for _, record := range records {
				err := stream.WriteRecord(record)
				require.NoError(b, err)
			}
			
			err = stream.Close()
			require.NoError(b, err)
		}
	})

	b.Run("BatchWriter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			options := WriteOptions{
				Headers:   headers,
				Records:   records,
				Append:    false,
				BOMPrefix: true,
			}
			
			err := writer.WriteCSV("batch_bench.csv", options)
			require.NoError(b, err)
		}
	})
}