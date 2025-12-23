package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain removed - flag parsing is handled in main.go

func TestLatestDownloadedDate(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		expected  string
		expectOk  bool
	}{
		{
			name: "valid files with dates",
			files: []string{
				"2025 01 15 ISX Daily Report.xlsx",
				"2025 01 20 ISX Daily Report.xlsx", 
				"2025 01 10 ISX Daily Report.xlsx",
			},
			expected: "2025-01-20",
			expectOk: true,
		},
		{
			name: "no matching files",
			files: []string{
				"other_file.txt",
				"report.pdf",
			},
			expected: "",
			expectOk: false,
		},
		{
			name:     "empty directory",
			files:    []string{},
			expected: "",
			expectOk: false,
		},
		{
			name: "single valid file",
			files: []string{
				"2025 01 01 ISX Daily Report.xlsx",
			},
			expected: "2025-01-01",
			expectOk: true,
		},
		{
			name: "mixed valid and invalid files",
			files: []string{
				"2025 01 15 ISX Daily Report.xlsx",
				"invalid_file.xlsx",
				"2025 01 25 ISX Daily Report.xlsx",
				"another_file.txt",
			},
			expected: "2025-01-25",
			expectOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			
			// Create test files
			for _, filename := range tt.files {
				filePath := filepath.Join(tmpDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}
			
			// Test the function
			result, ok := latestDownloadedDate(tmpDir)
			
			if tt.expectOk {
				assert.True(t, ok)
				assert.Equal(t, tt.expected, result.Format("2006-01-02"))
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestIsValidLicenseFormat(t *testing.T) {
	tests := []struct {
		name        string
		licenseKey  string
		expected    bool
	}{
		{
			name:       "valid ISX1M format",
			licenseKey: "ISX1M-ABC123DEF456GHI789JKL",
			expected:   true,
		},
		{
			name:       "valid ISX3M format",
			licenseKey: "ISX3M-ABC123DEF456GHI789JKL",
			expected:   true,
		},
		{
			name:       "valid ISX6M format",
			licenseKey: "ISX6M-ABC123DEF456GHI789JKL",
			expected:   true,
		},
		{
			name:       "valid ISX1Y format",
			licenseKey: "ISX1Y-ABC123DEF456GHI789JKL",
			expected:   true,
		},
		{
			name:       "invalid prefix",
			licenseKey: "INVALID-ABC123DEF456GHI789JKL",
			expected:   false,
		},
		{
			name:       "empty string",
			licenseKey: "",
			expected:   false,
		},
		{
			name:       "partial prefix",
			licenseKey: "ISX-ABC123DEF456GHI789JKL",
			expected:   false,
		},
		{
			name:       "lowercase prefix",
			licenseKey: "isx1m-ABC123DEF456GHI789JKL", 
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidLicenseFormat(tt.licenseKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDownloadFile(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		expectError    bool
		expectedSize   int
	}{
		{
			name:           "successful download",
			serverResponse: "test file content",
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedSize:   17, // len("test file content")
		},
		{
			name:           "server error",
			serverResponse: "",
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			expectedSize:   0,
		},
		{
			name:           "not found error",
			serverResponse: "",
			statusCode:     http.StatusNotFound,
			expectError:    true,
			expectedSize:   0,
		},
		{
			name:           "empty response",
			serverResponse: "",
			statusCode:     http.StatusOK,
			expectError:    false,
			expectedSize:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Create temporary file for download
			tmpDir := t.TempDir()
			destPath := filepath.Join(tmpDir, "downloaded_file.xlsx")

			// Test download
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := downloadFile(server.URL, destPath, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify file was created and has correct content
				if _, statErr := os.Stat(destPath); statErr == nil {
					content, readErr := os.ReadFile(destPath)
					assert.NoError(t, readErr)
					assert.Equal(t, tt.expectedSize, len(content))
					if tt.expectedSize > 0 {
						assert.Equal(t, tt.serverResponse, string(content))
					}
				}
			}
		})
	}
}

func TestTimedAction(t *testing.T) {
	tests := []struct {
		name        string
		actionName  string
		actionFunc  func() error
		expectError bool
	}{
		{
			name:       "successful action",
			actionName: "test_action",
			actionFunc: func() error { return nil },
			expectError: false,
		},
		{
			name:       "failing action",
			actionName: "failing_action", 
			actionFunc: func() error { return assert.AnError },
			expectError: true,
		},
		{
			name:       "slow action",
			actionName: "slow_action",
			actionFunc: func() error { 
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock chromedp action
			mockAction := &mockChromedpAction{
				doFunc: tt.actionFunc,
			}

			timedActionWrapper := timedAction(tt.actionName, mockAction)

			ctx := context.Background()
			err := timedActionWrapper.Do(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScrapePage(t *testing.T) {
	tests := []struct {
		name                string
		mockRows            []map[string]string
		outDir              string
		expectContinue      bool
		expectError         bool
		expectedDownloads   int
		expectedExisting    int
	}{
		{
			name: "new files to download",
			mockRows: []map[string]string{
				{
					"href": "/reports/2025_01_01_report.xlsx",
					"date": "01/01/2025",
					"typ":  "Daily",
				},
				{
					"href": "/reports/2025_01_02_report.xlsx", 
					"date": "02/01/2025",
					"typ":  "Daily",
				},
			},
			expectContinue:    true,
			expectError:       false,
			expectedDownloads: 2,
			expectedExisting:  0,
		},
		{
			name: "existing files found",
			mockRows: []map[string]string{
				{
					"href": "/reports/existing_report.xlsx",
					"date": "01/01/2025", 
					"typ":  "Daily",
				},
			},
			expectContinue:    false,
			expectError:       false,
			expectedDownloads: 0,
			expectedExisting:  1,
		},
		{
			name: "non-daily reports filtered out",
			mockRows: []map[string]string{
				{
					"href": "/reports/weekly_report.xlsx",
					"date": "01/01/2025",
					"typ":  "Weekly",
				},
			},
			expectContinue:    true,
			expectError:       false,
			expectedDownloads: 0,
			expectedExisting:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Pre-create existing files for "existing files found" test
			if tt.name == "existing files found" {
				existingPath := filepath.Join(tmpDir, "2025 01 01 ISX Daily Report.xlsx")
				err := os.WriteFile(existingPath, []byte("existing"), 0644)
				require.NoError(t, err)
			}

			// This test requires significant refactoring of scrapePage to be testable
			// For now, we test the supporting functions and logic
			t.Skip("scrapePage requires ChromeDP context - testing individual components")
		})
	}
}

// TestFlagParsing removed - can't test flag parsing with main package flags defined

func TestRegexPatterns(t *testing.T) {
	// Test filename pattern used in latestDownloadedDate
	filePattern := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	
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
			name:     "invalid format - wrong extension",
			filename: "2025 01 15 ISX Daily Report.pdf",
			matches:  false,
			groups:   nil,
		},
		{
			name:     "invalid format - missing parts",
			filename: "2025 01 ISX Daily Report.xlsx",
			matches:  false,
			groups:   nil,
		},
		{
			name:     "invalid format - extra characters",
			filename: "2025 01 15 ISX Daily Report Extra.xlsx",
			matches:  false,
			groups:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := filePattern.FindStringSubmatch(tt.filename)
			
			if tt.matches {
				assert.NotNil(t, matches)
				assert.Equal(t, tt.groups, matches)
			} else {
				assert.Nil(t, matches)
			}
		})
	}
}

// Mock ChromeDP action for testing
type mockChromedpAction struct {
	doFunc func() error
}

func (m *mockChromedpAction) Do(ctx context.Context) error {
	if m.doFunc != nil {
		return m.doFunc()
	}
	return nil
}

// Benchmark latestDownloadedDate function
func BenchmarkLatestDownloadedDate(b *testing.B) {
	// Create temp directory with test files
	tmpDir := b.TempDir()
	
	// Create many test files
	for i := 1; i <= 100; i++ {
		filename := fmt.Sprintf("2025 01 %02d ISX Daily Report.xlsx", i)
		filePath := filepath.Join(tmpDir, filename)
		os.WriteFile(filePath, []byte("test"), 0644)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		latestDownloadedDate(tmpDir)
	}
}

// Test concurrent file operations
func TestConcurrentOperations(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test concurrent file downloads (simulated)
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()
	
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				destPath := filepath.Join(tmpDir, fmt.Sprintf("file_%d.xlsx", id))
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				err := downloadFile(server.URL, destPath, logger)
				assert.NoError(t, err)
			}(i)
		}
	
	// Wait for all downloads to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify all files were created
	files, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, numGoroutines, len(files))
}

// Test error handling in various scenarios
func TestErrorHandling(t *testing.T) {
	t.Run("download file with invalid URL", func(t *testing.T) {
		tmpDir := t.TempDir()
		destPath := filepath.Join(tmpDir, "test.xlsx")
		
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := downloadFile("invalid-url", destPath, logger)
		assert.Error(t, err)
	})

	t.Run("download file to invalid destination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}))
		defer server.Close()
		
		// Try to write to a directory that doesn't exist without creating parent dirs
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := downloadFile(server.URL, "/nonexistent/path/file.xlsx", logger)
		assert.Error(t, err)
	})

	t.Run("latest downloaded date with unreadable directory", func(t *testing.T) {
		// Test with directory that doesn't exist
		_, ok := latestDownloadedDate("/nonexistent/directory")
		assert.False(t, ok)
	})
}
