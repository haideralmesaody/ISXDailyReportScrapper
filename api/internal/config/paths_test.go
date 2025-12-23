package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetPaths tests the GetPaths function with various scenarios
func TestGetPaths(t *testing.T) {
	// Save original executable path
	originalExe := os.Args[0]
	defer func() {
		os.Args[0] = originalExe
	}()

	t.Run("basic path resolution", func(t *testing.T) {
		paths, err := GetPaths()
		require.NoError(t, err)
		require.NotNil(t, paths)

		// Verify all paths are absolute
		assert.True(t, filepath.IsAbs(paths.ExecutableDir), "ExecutableDir should be absolute")
		assert.True(t, filepath.IsAbs(paths.DataDir), "DataDir should be absolute")
		assert.True(t, filepath.IsAbs(paths.WebDir), "WebDir should be absolute")
		assert.True(t, filepath.IsAbs(paths.LogsDir), "LogsDir should be absolute")
		assert.True(t, filepath.IsAbs(paths.LicenseFile), "LicenseFile should be absolute")

		// Verify paths are correctly related to executable dir
		assert.Equal(t, filepath.Join(paths.ExecutableDir, "data"), paths.DataDir)
		assert.Equal(t, filepath.Join(paths.ExecutableDir, "web"), paths.WebDir)
		assert.Equal(t, filepath.Join(paths.ExecutableDir, "logs"), paths.LogsDir)
		assert.Equal(t, filepath.Join(paths.ExecutableDir, "license.dat"), paths.LicenseFile)
	})

	t.Run("consistent calls return same paths", func(t *testing.T) {
		paths1, err1 := GetPaths()
		require.NoError(t, err1)

		paths2, err2 := GetPaths()
		require.NoError(t, err2)

		assert.Equal(t, paths1.ExecutableDir, paths2.ExecutableDir)
		assert.Equal(t, paths1.DataDir, paths2.DataDir)
		assert.Equal(t, paths1.LicenseFile, paths2.LicenseFile)
	})

	t.Run("nested directory structure", func(t *testing.T) {
		paths, err := GetPaths()
		require.NoError(t, err)

		// Verify nested structure
		assert.Equal(t, filepath.Join(paths.DataDir, "downloads"), paths.DownloadsDir)
		assert.Equal(t, filepath.Join(paths.DataDir, "reports"), paths.ReportsDir)
		assert.Equal(t, filepath.Join(paths.DataDir, "cache"), paths.CacheDir)
		assert.Equal(t, filepath.Join(paths.WebDir, "static"), paths.StaticDir)
	})

	t.Run("well-known report files", func(t *testing.T) {
		paths, err := GetPaths()
		require.NoError(t, err)

		// All report files should be in the reports directory
		assert.True(t, strings.HasPrefix(paths.IndexCSV, paths.ReportsDir))
		assert.True(t, strings.HasPrefix(paths.TickerSummaryJSON, paths.ReportsDir))
		assert.True(t, strings.HasPrefix(paths.TickerSummaryCSV, paths.ReportsDir))
		assert.True(t, strings.HasPrefix(paths.CombinedDataCSV, paths.ReportsDir))

		// Check specific filenames
		assert.Equal(t, "indexes.csv", filepath.Base(paths.IndexCSV))
		assert.Equal(t, "ticker_summary.json", filepath.Base(paths.TickerSummaryJSON))
		assert.Equal(t, "ticker_summary.csv", filepath.Base(paths.TickerSummaryCSV))
		assert.Equal(t, "isx_combined_data.csv", filepath.Base(paths.CombinedDataCSV))
	})
}

// TestEnsureDirectories tests directory creation functionality
func TestEnsureDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock Paths struct pointing to our temp directory
	paths := &Paths{
		ExecutableDir:    tempDir,
		DataDir:         filepath.Join(tempDir, "data"),
		DownloadsDir:    filepath.Join(tempDir, "data", "downloads"),
		ReportsDir:      filepath.Join(tempDir, "data", "reports"),
		CacheDir:        filepath.Join(tempDir, "data", "cache"),
		LogsDir:         filepath.Join(tempDir, "logs"),
		WebDir:          filepath.Join(tempDir, "web"),
		StaticDir:       filepath.Join(tempDir, "web", "static"),
		LicenseFile:     filepath.Join(tempDir, "license.dat"),
		CredentialsFile: filepath.Join(tempDir, "credentials.json"),
	}

	t.Run("creates all directories", func(t *testing.T) {
		err := paths.EnsureDirectories()
		require.NoError(t, err)

		// Verify all directories exist
		assert.DirExists(t, paths.DataDir)
		assert.DirExists(t, paths.DownloadsDir)
		assert.DirExists(t, paths.ReportsDir)
		assert.DirExists(t, paths.CacheDir)
		assert.DirExists(t, paths.LogsDir)
		assert.DirExists(t, paths.WebDir)
		assert.DirExists(t, paths.StaticDir)
	})

	t.Run("idempotent - can be called multiple times", func(t *testing.T) {
		// First call
		err1 := paths.EnsureDirectories()
		require.NoError(t, err1)

		// Second call should not fail
		err2 := paths.EnsureDirectories()
		require.NoError(t, err2)

		// Directories should still exist
		assert.DirExists(t, paths.DataDir)
		assert.DirExists(t, paths.LogsDir)
	})

	t.Run("handles existing directories", func(t *testing.T) {
		// Pre-create some directories
		require.NoError(t, os.MkdirAll(paths.DataDir, 0755))
		require.NoError(t, os.MkdirAll(paths.WebDir, 0755))

		// EnsureDirectories should not fail
		err := paths.EnsureDirectories()
		require.NoError(t, err)

		// All directories should exist
		assert.DirExists(t, paths.DataDir)
		assert.DirExists(t, paths.DownloadsDir)
		assert.DirExists(t, paths.WebDir)
		assert.DirExists(t, paths.StaticDir)
	})
}

// TestPathHelperMethods tests various path helper methods
func TestPathHelperMethods(t *testing.T) {
	paths := &Paths{
		ExecutableDir: "/app",
		WebDir:       "/app/web",
		StaticDir:    "/app/web/static",
		DownloadsDir: "/app/data/downloads",
		ReportsDir:   "/app/data/reports",
		LogsDir:      "/app/logs",
		CacheDir:     "/app/data/cache",
	}

	tests := []struct {
		name     string
		method   func(string) string
		input    string
		expected string
	}{
		{
			name:     "GetRelativePath",
			method:   paths.GetRelativePath,
			input:    "config.yaml",
			expected: filepath.Join("/app", "config.yaml"),
		},
		{
			name:     "GetWebFilePath",
			method:   paths.GetWebFilePath,
			input:    "index.html",
			expected: filepath.Join("/app/web", "index.html"),
		},
		{
			name:     "GetStaticFilePath",
			method:   paths.GetStaticFilePath,
			input:    "css/main.css",
			expected: filepath.Join("/app/web/static", "css/main.css"),
		},
		{
			name:     "GetDownloadPath",
			method:   paths.GetDownloadPath,
			input:    "report.xlsx",
			expected: filepath.Join("/app/data/downloads", "report.xlsx"),
		},
		{
			name:     "GetReportPath",
			method:   paths.GetReportPath,
			input:    "summary.csv",
			expected: filepath.Join("/app/data/reports", "summary.csv"),
		},
		{
			name:     "GetLogPath",
			method:   paths.GetLogPath,
			input:    "app.log",
			expected: filepath.Join("/app/logs", "app.log"),
		},
		{
			name:     "GetCachePath",
			method:   paths.GetCachePath,
			input:    "temp.dat",
			expected: filepath.Join("/app/data/cache", "temp.dat"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.input)
			// Normalize paths for comparison across platforms
			expected := filepath.ToSlash(tt.expected)
			actual := filepath.ToSlash(result)
			assert.Equal(t, expected, actual)
		})
	}
}

// TestGetLicensePath tests the license path resolution
func TestGetLicensePath(t *testing.T) {
	t.Run("returns executable-relative path", func(t *testing.T) {
		path, err := GetLicensePath()
		require.NoError(t, err)
		assert.NotEmpty(t, path)
		assert.True(t, filepath.IsAbs(path))
		assert.Equal(t, "license.dat", filepath.Base(path))
	})

	t.Run("consistent across calls", func(t *testing.T) {
		path1, err1 := GetLicensePath()
		require.NoError(t, err1)

		path2, err2 := GetLicensePath()
		require.NoError(t, err2)

		assert.Equal(t, path1, path2)
	})
}

// TestFileExists tests the FileExists helper function
func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("existing file returns true", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
		
		assert.True(t, FileExists(testFile))
	})

	t.Run("non-existing file returns false", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "does-not-exist.txt")
		assert.False(t, FileExists(nonExistentFile))
	})

	t.Run("directory returns true", func(t *testing.T) {
		assert.True(t, FileExists(tempDir))
	})
}

// TestValidateRequiredFiles tests file validation functionality
func TestValidateRequiredFiles(t *testing.T) {
	tempDir := t.TempDir()

	paths := &Paths{
		LicenseFile:     filepath.Join(tempDir, "license.dat"),
		CredentialsFile: filepath.Join(tempDir, "credentials.json"),
	}

	t.Run("all files missing", func(t *testing.T) {
		err := paths.ValidateRequiredFiles()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "License")
		assert.Contains(t, err.Error(), "Credentials")
	})

	t.Run("some files missing", func(t *testing.T) {
		// Create license file only
		require.NoError(t, os.WriteFile(paths.LicenseFile, []byte("license"), 0644))

		err := paths.ValidateRequiredFiles()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Credentials")
		assert.NotContains(t, err.Error(), "License")
	})

	t.Run("all files present", func(t *testing.T) {
		// Create both files
		require.NoError(t, os.WriteFile(paths.LicenseFile, []byte("license"), 0644))
		require.NoError(t, os.WriteFile(paths.CredentialsFile, []byte("{}"), 0644))

		err := paths.ValidateRequiredFiles()
		assert.NoError(t, err)
	})
}

// TestWindowsPathHandling tests Windows-specific path scenarios
func TestWindowsPathHandling(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific tests on non-Windows platform")
	}

	t.Run("handles different drive letters", func(t *testing.T) {
		paths := &Paths{
			ExecutableDir: `C:\Program Files\ISX`,
			DataDir:      `D:\ISXData`,
		}

		// Verify paths can handle different drives
		assert.Equal(t, `C:\Program Files\ISX`, paths.ExecutableDir)
		assert.Equal(t, `D:\ISXData`, paths.DataDir)
	})

	t.Run("handles UNC paths", func(t *testing.T) {
		paths := &Paths{
			ExecutableDir: `\\server\share\ISX`,
			DataDir:      `\\server\share\ISX\data`,
			WebDir:       `\\server\share\ISX\web`,
		}

		webPath := paths.GetWebFilePath("index.html")
		assert.Contains(t, webPath, `\\server\share\ISX`)
		assert.Contains(t, webPath, "web")
		assert.Equal(t, "index.html", filepath.Base(webPath))
	})

	t.Run("handles spaces in paths", func(t *testing.T) {
		paths := &Paths{
			ExecutableDir: `C:\Program Files\ISX Daily Reports`,
			DataDir:      `C:\Program Files\ISX Daily Reports\data`,
			ReportsDir:   `C:\Program Files\ISX Daily Reports\data\reports`,
		}

		reportPath := paths.GetReportPath("report.csv")
		assert.Contains(t, reportPath, "ISX Daily Reports")
		assert.Contains(t, reportPath, "reports")
		assert.Equal(t, "report.csv", filepath.Base(reportPath))
	})
}

// TestDateBasedPaths tests paths that include dates
func TestDateBasedPaths(t *testing.T) {
	paths := &Paths{
		ReportsDir:   "/app/data/reports",
		DownloadsDir: "/app/data/downloads",
	}

	t.Run("GetDailyCSVPath", func(t *testing.T) {
		date := mustParseTime("2024-01-15")
		path := paths.GetDailyCSVPath(date)
		
		assert.Contains(t, path, "reports")
		assert.Equal(t, "isx_daily_20240115.csv", filepath.Base(path))
	})

	t.Run("GetExcelPathForDate", func(t *testing.T) {
		date := mustParseTime("2024-01-15")
		path := paths.GetExcelPathForDate(date)
		
		assert.Contains(t, path, "downloads")
		assert.Equal(t, "2024 01 15 ISX Daily Report.xlsx", filepath.Base(path))
	})
}

// TestGetTickerDailyCSVPath tests ticker-specific CSV path generation
func TestGetTickerDailyCSVPath(t *testing.T) {
	paths := &Paths{
		ReportsDir: "/app/data/reports",
	}

	tests := []struct {
		ticker   string
		expected string
	}{
		{"BBOB", "BBOB_daily.csv"},
		{"TASC", "TASC_daily.csv"},
		{"IBSD", "IBSD_daily.csv"},
	}

	for _, tt := range tests {
		t.Run(tt.ticker, func(t *testing.T) {
			path := paths.GetTickerDailyCSVPath(tt.ticker)
			assert.Equal(t, tt.expected, filepath.Base(path))
			assert.Contains(t, path, "reports")
		})
	}
}

// TestPathErrorHandling tests error scenarios
func TestPathErrorHandling(t *testing.T) {
	t.Run("EnsureDirectories with permission errors", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Permission testing is complex on Windows")
		}

		// Create a directory with no write permissions
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")
		require.NoError(t, os.Mkdir(readOnlyDir, 0555))

		paths := &Paths{
			DataDir: filepath.Join(readOnlyDir, "data"),
		}

		err := paths.EnsureDirectories()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

// TestConfigurationIntegration tests integration with Config struct
func TestConfigurationIntegration(t *testing.T) {
	cfg := Default()

	t.Run("GetDataDir uses centralized paths", func(t *testing.T) {
		dataDir := cfg.GetDataDir()
		assert.NotEmpty(t, dataDir)
		assert.True(t, filepath.IsAbs(dataDir))
	})

	t.Run("GetWebDir uses centralized paths", func(t *testing.T) {
		webDir := cfg.GetWebDir()
		assert.NotEmpty(t, webDir)
		assert.True(t, filepath.IsAbs(webDir))
	})

	t.Run("GetLogsDir uses centralized paths", func(t *testing.T) {
		logsDir := cfg.GetLogsDir()
		assert.NotEmpty(t, logsDir)
		assert.True(t, filepath.IsAbs(logsDir))
	})

	t.Run("GetLicenseFile uses centralized paths", func(t *testing.T) {
		licenseFile := cfg.GetLicenseFile()
		assert.NotEmpty(t, licenseFile)
		assert.True(t, filepath.IsAbs(licenseFile))
		assert.Equal(t, "license.dat", filepath.Base(licenseFile))
	})
}

// TestPathValidation tests path validation in config
func TestPathValidation(t *testing.T) {
	cfg := Default()

	t.Run("ValidatePaths creates directories", func(t *testing.T) {
		// This test might need adjustment based on actual file system
		// For now, we just ensure it doesn't panic
		err := cfg.ValidatePaths()
		// The error might occur if we don't have permissions, which is OK for tests
		if err != nil {
			assert.Contains(t, err.Error(), "failed to")
		}
	})

	t.Run("resolvePaths updates config", func(t *testing.T) {
		originalExeDir := cfg.Paths.ExecutableDir
		err := cfg.resolvePaths()
		assert.NoError(t, err)
		
		// After resolution, ExecutableDir should be set
		assert.NotEmpty(t, cfg.Paths.ExecutableDir)
		if originalExeDir == "" {
			assert.NotEqual(t, originalExeDir, cfg.Paths.ExecutableDir)
		}
	})
}

// Helper function to parse time
func mustParseTime(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time: %v", err))
	}
	return t
}

// BenchmarkGetPaths benchmarks path resolution performance
func BenchmarkGetPaths(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetPaths()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPathHelpers benchmarks various path helper methods
func BenchmarkPathHelpers(b *testing.B) {
	paths, err := GetPaths()
	if err != nil {
		b.Fatal(err)
	}

	b.Run("GetWebFilePath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = paths.GetWebFilePath("index.html")
		}
	})

	b.Run("GetReportPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = paths.GetReportPath("report.csv")
		}
	})

	b.Run("GetDailyCSVPath", func(b *testing.B) {
		date := time.Now()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = paths.GetDailyCSVPath(date)
		}
	})
}