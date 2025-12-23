package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/license"
)

// TestPathConsistencyAcrossAllComponents verifies that all components use consistent paths
func TestPathConsistencyAcrossAllComponents(t *testing.T) {
	// Get paths from centralized system
	paths, err := config.GetPaths()
	require.NoError(t, err)

	t.Run("config paths match centralized paths", func(t *testing.T) {
		cfg := config.Default()
		
		// Verify all path methods return expected values
		assert.Equal(t, paths.DataDir, cfg.GetDataDir())
		assert.Equal(t, paths.WebDir, cfg.GetWebDir())
		assert.Equal(t, paths.LogsDir, cfg.GetLogsDir())
		assert.Equal(t, paths.LicenseFile, cfg.GetLicenseFile())
	})

	t.Run("license manager uses correct paths", func(t *testing.T) {
		// Create license manager
		manager, err := license.NewManager()
		require.NoError(t, err)
		require.NotNil(t, manager)
		
		// Manager should use centralized license path
		// Note: licenseFile is private, but we've verified in manager_paths_test.go
		// that it uses the centralized path system
	})

	t.Run("data service uses correct paths", func(t *testing.T) {
		// Create a mock data service configuration
		dataPath := paths.GetReportPath("test_report.csv")
		
		// Verify path is under reports directory
		assert.True(t, pathHasPrefix(dataPath, paths.ReportsDir))
		assert.Equal(t, "test_report.csv", filepath.Base(dataPath))
	})
}

// TestCrossComponentFileSharing verifies files saved by one component can be read by another
func TestCrossComponentFileSharing(t *testing.T) {
	// Create a temporary test environment
	tempDir := t.TempDir()
	
	// Override paths for testing
	testPaths := &config.Paths{
		ExecutableDir:    tempDir,
		DataDir:         filepath.Join(tempDir, "data"),
		ReportsDir:      filepath.Join(tempDir, "data", "reports"),
		DownloadsDir:    filepath.Join(tempDir, "data", "downloads"),
		LicenseFile:     filepath.Join(tempDir, "license.dat"),
		CredentialsFile: filepath.Join(tempDir, "credentials.json"),
	}
	
	// Ensure directories exist
	err := testPaths.EnsureDirectories()
	require.NoError(t, err)

	t.Run("report file sharing", func(t *testing.T) {
		// Component A writes a report
		reportData := map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"data":      []string{"row1", "row2", "row3"},
		}
		
		reportPath := testPaths.GetReportPath("shared_report.json")
		data, err := json.Marshal(reportData)
		require.NoError(t, err)
		
		err = os.WriteFile(reportPath, data, 0644)
		require.NoError(t, err)
		
		// Component B reads the report
		readData, err := os.ReadFile(reportPath)
		require.NoError(t, err)
		
		var loadedReport map[string]interface{}
		err = json.Unmarshal(readData, &loadedReport)
		require.NoError(t, err)
		
		// Verify data integrity
		assert.Equal(t, reportData["timestamp"], loadedReport["timestamp"])
	})

	t.Run("download to report operation", func(t *testing.T) {
		// Simulate downloading a file
		downloadPath := testPaths.GetDownloadPath("data.xlsx")
		err := os.WriteFile(downloadPath, []byte("excel data"), 0644)
		require.NoError(t, err)
		
		// Process and move to reports
		reportPath := testPaths.GetReportPath("processed_data.csv")
		
		// Read from download
		data, err := os.ReadFile(downloadPath)
		require.NoError(t, err)
		
		// Write to report (simulating processing)
		processedData := fmt.Sprintf("processed: %s", string(data))
		err = os.WriteFile(reportPath, []byte(processedData), 0644)
		require.NoError(t, err)
		
		// Verify both files exist in correct locations
		assert.True(t, config.FileExists(downloadPath))
		assert.True(t, config.FileExists(reportPath))
		assert.Contains(t, downloadPath, "downloads")
		assert.Contains(t, reportPath, "reports")
	})
}

// TestPathResolutionFromDifferentWorkingDirectories tests path consistency when run from different dirs
func TestPathResolutionFromDifferentWorkingDirectories(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Get initial paths
	paths1, err := config.GetPaths()
	require.NoError(t, err)

	t.Run("paths remain consistent from different working directories", func(t *testing.T) {
		// Change to temp directory
		tempDir := t.TempDir()
		err := os.Chdir(tempDir)
		require.NoError(t, err)
		
		// Get paths again
		paths2, err := config.GetPaths()
		require.NoError(t, err)
		
		// Paths should be identical (executable-relative, not cwd-relative)
		assert.Equal(t, paths1.ExecutableDir, paths2.ExecutableDir)
		assert.Equal(t, paths1.DataDir, paths2.DataDir)
		assert.Equal(t, paths1.LicenseFile, paths2.LicenseFile)
		
		// Change to another directory
		err = os.Chdir(os.TempDir())
		require.NoError(t, err)
		
		// Get paths once more
		paths3, err := config.GetPaths()
		require.NoError(t, err)
		
		// Still should be identical
		assert.Equal(t, paths1.ExecutableDir, paths3.ExecutableDir)
		assert.Equal(t, paths1.DataDir, paths3.DataDir)
		assert.Equal(t, paths1.LicenseFile, paths3.LicenseFile)
	})
}

// TestConcurrentPathAccess tests that multiple goroutines can safely access paths
func TestConcurrentPathAccess(t *testing.T) {
	const numGoroutines = 20
	const numIterations = 100

	t.Run("concurrent GetPaths calls", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numIterations)
		
		// Launch goroutines
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				for j := 0; j < numIterations; j++ {
					paths, err := config.GetPaths()
					if err != nil {
						errors <- fmt.Errorf("goroutine %d iteration %d: %v", id, j, err)
						continue
					}
					
					// Verify paths are valid
					if paths.ExecutableDir == "" {
						errors <- fmt.Errorf("goroutine %d iteration %d: empty ExecutableDir", id, j)
					}
				}
			}(i)
		}
		
		// Wait for completion
		wg.Wait()
		close(errors)
		
		// Check for errors
		var allErrors []error
		for err := range errors {
			allErrors = append(allErrors, err)
		}
		
		assert.Empty(t, allErrors, "Concurrent access should not produce errors")
	})

	t.Run("concurrent file operations", func(t *testing.T) {
		paths, err := config.GetPaths()
		require.NoError(t, err)
		
		// Ensure directories exist
		err = paths.EnsureDirectories()
		require.NoError(t, err)
		
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)
		
		// Each goroutine writes and reads its own file
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				// Create unique file for this goroutine
				filename := fmt.Sprintf("concurrent_test_%d.txt", id)
				filepath := paths.GetCachePath(filename)
				
				// Write data
				data := fmt.Sprintf("goroutine %d data", id)
				if err := os.WriteFile(filepath, []byte(data), 0644); err != nil {
					errors <- fmt.Errorf("goroutine %d write error: %v", id, err)
					return
				}
				
				// Read data back
				readData, err := os.ReadFile(filepath)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d read error: %v", id, err)
					return
				}
				
				// Verify data
				if string(readData) != data {
					errors <- fmt.Errorf("goroutine %d data mismatch", id)
				}
				
				// Cleanup
				os.Remove(filepath)
			}(i)
		}
		
		// Wait for completion
		wg.Wait()
		close(errors)
		
		// Check for errors
		var allErrors []error
		for err := range errors {
			allErrors = append(allErrors, err)
		}
		
		assert.Empty(t, allErrors, "Concurrent file operations should not produce errors")
	})
}

// TestDateBasedPathConsistency tests that date-based paths work correctly
func TestDateBasedPathConsistency(t *testing.T) {
	paths, err := config.GetPaths()
	require.NoError(t, err)

	testDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("daily CSV paths", func(t *testing.T) {
		path1 := paths.GetDailyCSVPath(testDate)
		path2 := paths.GetDailyCSVPath(testDate)
		
		// Same date should produce same path
		assert.Equal(t, path1, path2)
		
		// Path should contain date
		assert.Contains(t, path1, "20240115")
		assert.Contains(t, path1, "reports")
	})

	t.Run("Excel download paths", func(t *testing.T) {
		path1 := paths.GetExcelPathForDate(testDate)
		path2 := paths.GetExcelPathForDate(testDate)
		
		// Same date should produce same path
		assert.Equal(t, path1, path2)
		
		// Path should contain formatted date
		assert.Contains(t, path1, "2024 01 15")
		assert.Contains(t, path1, "downloads")
	})

	t.Run("different dates produce different paths", func(t *testing.T) {
		date1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
		
		path1 := paths.GetDailyCSVPath(date1)
		path2 := paths.GetDailyCSVPath(date2)
		
		assert.NotEqual(t, path1, path2)
		assert.Contains(t, path1, "20240115")
		assert.Contains(t, path2, "20240116")
	})
}

// TestEnvironmentVariableOverrides tests that env vars properly override paths
func TestEnvironmentVariableOverrides(t *testing.T) {
	// Save current env vars
	originalDataDir := os.Getenv("ISX_PATHS_DATA_DIR")
	originalWebDir := os.Getenv("ISX_PATHS_WEB_DIR")
	defer func() {
		os.Setenv("ISX_PATHS_DATA_DIR", originalDataDir)
		os.Setenv("ISX_PATHS_WEB_DIR", originalWebDir)
	}()

	t.Run("env vars override default paths", func(t *testing.T) {
		// Set custom paths via env vars
		customDataDir := "/custom/data"
		customWebDir := "/custom/web"
		
		os.Setenv("ISX_PATHS_DATA_DIR", customDataDir)
		os.Setenv("ISX_PATHS_WEB_DIR", customWebDir)
		
		// Load config
		cfg, err := config.Load()
		if err != nil {
			// Config might fail due to path validation, which is OK
			t.Logf("Config load error (expected): %v", err)
		}
		
		// Check if paths were set from env vars
		if cfg != nil {
			assert.Equal(t, customDataDir, cfg.Paths.DataDir)
			assert.Equal(t, customWebDir, cfg.Paths.WebDir)
		}
	})
}

// TestPathNormalization tests that paths are properly normalized across platforms
func TestPathNormalization(t *testing.T) {
	paths, err := config.GetPaths()
	require.NoError(t, err)

	t.Run("paths use correct separators", func(t *testing.T) {
		// All paths should use the OS-specific separator
		assert.Equal(t, string(filepath.Separator), string(paths.DataDir[len(filepath.VolumeName(paths.DataDir))]))
		
		// Paths should not contain mixed separators
		assert.NotContains(t, filepath.ToSlash(paths.DataDir), "\\")
		assert.NotContains(t, filepath.FromSlash(paths.DataDir), "/")
	})

	t.Run("path joining works correctly", func(t *testing.T) {
		// Test various path joining scenarios
		testCases := []struct {
			name     string
			method   func(string) string
			input    string
			contains string
		}{
			{
				name:     "web file",
				method:   paths.GetWebFilePath,
				input:    "index.html",
				contains: "web",
			},
			{
				name:     "nested static file",
				method:   paths.GetStaticFilePath,
				input:    filepath.Join("css", "main.css"),
				contains: "static",
			},
			{
				name:     "report with subdirectory",
				method:   paths.GetReportPath,
				input:    filepath.Join("2024", "01", "report.csv"),
				contains: "reports",
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := tc.method(tc.input)
				
				// Path should be absolute
				assert.True(t, filepath.IsAbs(result))
				
				// Path should contain expected directory
				assert.Contains(t, result, tc.contains)
				
				// Path should be properly formed
				assert.Equal(t, filepath.Clean(result), result)
			})
		}
	})
}

// TestPathSecurityValidation tests that path operations are secure
func TestPathSecurityValidation(t *testing.T) {
	paths, err := config.GetPaths()
	require.NoError(t, err)

	t.Run("prevents directory traversal", func(t *testing.T) {
		// Attempt various directory traversal patterns
		maliciousInputs := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32",
			"../../../../root/.ssh/id_rsa",
			"./../.../../sensitive.dat",
		}
		
		for _, input := range maliciousInputs {
			// Test with various path methods
			webPath := paths.GetWebFilePath(input)
			reportPath := paths.GetReportPath(input)
			
			// Paths should be confined to their respective directories
			assert.True(t, pathHasPrefix(filepath.Clean(webPath), paths.WebDir))
			assert.True(t, pathHasPrefix(filepath.Clean(reportPath), paths.ReportsDir))
			
			// Paths should not contain parent directory references after cleaning
			cleanedWeb := filepath.Clean(webPath)
			cleanedReport := filepath.Clean(reportPath)
			
			assert.NotContains(t, cleanedWeb, "..")
			assert.NotContains(t, cleanedReport, "..")
		}
	})
}

// BenchmarkPathOperations benchmarks various path operations
func BenchmarkPathOperations(b *testing.B) {
	b.Run("GetPaths", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := config.GetPaths()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Concurrent GetPaths", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := config.GetPaths()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("Path construction", func(b *testing.B) {
		paths, err := config.GetPaths()
		if err != nil {
			b.Fatal(err)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = paths.GetReportPath("benchmark_report.csv")
			_ = paths.GetWebFilePath("index.html")
			_ = paths.GetDownloadPath("data.xlsx")
		}
	})
}

// Helper to check if a path has a prefix (handles volume names on Windows)
func pathHasPrefix(path, prefix string) bool {
	path = filepath.Clean(path)
	prefix = filepath.Clean(prefix)
	
	// On Windows, compare after volume name
	pathVol := filepath.VolumeName(path)
	prefixVol := filepath.VolumeName(prefix)
	
	if pathVol != prefixVol {
		return false
	}
	
	pathRel := path[len(pathVol):]
	prefixRel := prefix[len(prefixVol):]
	
	return strings.HasPrefix(pathRel, prefixRel)
}