package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiscovery(t *testing.T) {
	basePath := "/test/base"
	discovery := NewDiscovery(basePath)
	
	assert.NotNil(t, discovery)
	assert.Equal(t, basePath, discovery.basePath)
}

func TestFindExcelFiles(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedCount int
		description   string
	}{
		{
			name:          "only Excel files",
			files:         []string{"report1.xlsx", "report2.xls", "report3.XLSX"},
			expectedCount: 3,
			description:   "Should find all Excel files regardless of case",
		},
		{
			name:          "mixed file types",
			files:         []string{"report.xlsx", "data.csv", "doc.pdf", "sheet.xls"},
			expectedCount: 2,
			description:   "Should find only Excel files",
		},
		{
			name:          "no Excel files",
			files:         []string{"data.csv", "doc.pdf", "readme.txt"},
			expectedCount: 0,
			description:   "Should find no Excel files",
		},
		{
			name:          "empty directory",
			files:         []string{},
			expectedCount: 0,
			description:   "Should handle empty directory",
		},
		{
			name:          "Excel files with various names",
			files:         []string{"2025_01_15_report.xlsx", "daily-report.xls", "index.XLSX"},
			expectedCount: 3,
			description:   "Should find Excel files with various naming patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			discovery := NewDiscovery(tmpDir)

			testDir := "excel_test"
			fullTestDir := filepath.Join(tmpDir, testDir)
			err := os.MkdirAll(fullTestDir, 0755)
			require.NoError(t, err)

			// Create test files with different modification times
			for i, filename := range tt.files {
				filePath := filepath.Join(fullTestDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)

				// Set different modification times to test sorting
				modTime := time.Now().Add(time.Duration(i) * time.Minute)
				err = os.Chtimes(filePath, modTime, modTime)
				require.NoError(t, err)
			}

			files, err := discovery.FindExcelFiles(testDir)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedCount, len(files), tt.description)

			// Verify files are sorted by modification time (oldest first)
			if len(files) > 1 {
				for i := 1; i < len(files); i++ {
					assert.True(t, files[i-1].ModTime.Before(files[i].ModTime) ||
						files[i-1].ModTime.Equal(files[i].ModTime),
						"Files should be sorted by modification time")
				}
			}

			// Verify file properties
			for _, file := range files {
				assert.NotEmpty(t, file.Name)
				assert.NotEmpty(t, file.Path)
				assert.False(t, file.IsDir)
				assert.Greater(t, file.Size, int64(0))
				assert.False(t, file.ModTime.IsZero())
			}
		})
	}
}

func TestFindCSVFiles(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedCount int
		description   string
	}{
		{
			name:          "only CSV files",
			files:         []string{"data1.csv", "data2.CSV", "report.csv"},
			expectedCount: 3,
			description:   "Should find all CSV files regardless of case",
		},
		{
			name:          "mixed file types",
			files:         []string{"data.csv", "report.xlsx", "doc.pdf"},
			expectedCount: 1,
			description:   "Should find only CSV files",
		},
		{
			name:          "no CSV files",
			files:         []string{"report.xlsx", "doc.pdf", "readme.txt"},
			expectedCount: 0,
			description:   "Should find no CSV files",
		},
		{
			name:          "empty directory",
			files:         []string{},
			expectedCount: 0,
			description:   "Should handle empty directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			discovery := NewDiscovery(tmpDir)

			testDir := "csv_test"
			fullTestDir := filepath.Join(tmpDir, testDir)
			err := os.MkdirAll(fullTestDir, 0755)
			require.NoError(t, err)

			// Create test files
			for _, filename := range tt.files {
				filePath := filepath.Join(fullTestDir, filename)
				err := os.WriteFile(filePath, []byte("test,csv,content"), 0644)
				require.NoError(t, err)
			}

			files, err := discovery.FindCSVFiles(testDir)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedCount, len(files), tt.description)

			// Verify file properties
			for _, file := range files {
				assert.NotEmpty(t, file.Name)
				assert.True(t, filepath.Ext(file.Name) == ".csv" || filepath.Ext(file.Name) == ".CSV")
				assert.False(t, file.IsDir)
			}
		})
	}
}

func TestFindFilesByPattern(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		pattern       string
		expectedCount int
		description   string
	}{
		{
			name:          "wildcard pattern",
			files:         []string{"test1.txt", "test2.txt", "other.csv"},
			pattern:       "test*.txt",
			expectedCount: 2,
			description:   "Should find files matching wildcard pattern",
		},
		{
			name:          "specific extension pattern",
			files:         []string{"file1.log", "file2.log", "file3.txt"},
			pattern:       "*.log",
			expectedCount: 2,
			description:   "Should find files with specific extension",
		},
		{
			name:          "no matches",
			files:         []string{"file1.txt", "file2.csv"},
			pattern:       "*.log",
			expectedCount: 0,
			description:   "Should return empty when no matches",
		},
		{
			name:          "exact filename pattern",
			files:         []string{"exact.txt", "other.txt"},
			pattern:       "exact.txt",
			expectedCount: 1,
			description:   "Should find exact filename match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			discovery := NewDiscovery(tmpDir)

			testDir := "pattern_test"
			fullTestDir := filepath.Join(tmpDir, testDir)
			err := os.MkdirAll(fullTestDir, 0755)
			require.NoError(t, err)

			// Create test files
			for _, filename := range tt.files {
				filePath := filepath.Join(fullTestDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			files, err := discovery.FindFilesByPattern(testDir, tt.pattern)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedCount, len(files), tt.description)

			// Verify file properties
			for _, file := range files {
				assert.NotEmpty(t, file.Name)
				assert.NotEmpty(t, file.Path)
				assert.False(t, file.IsDir)
			}
		})
	}
}

func TestFindDailyCSVFiles(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedDates []string
		description   string
	}{
		{
			name: "valid daily CSV files",
			files: []string{
				"isx_daily_2025_01_10.csv",
				"isx_daily_2025_01_11.csv",
				"isx_daily_2025_01_12.csv",
			},
			expectedDates: []string{"2025_01_10", "2025_01_11", "2025_01_12"},
			description:   "Should find and map daily CSV files by date",
		},
		{
			name: "mixed CSV files",
			files: []string{
				"isx_daily_2025_01_10.csv",
				"other_data.csv",
				"isx_daily_2025_01_11.csv",
				"summary.csv",
			},
			expectedDates: []string{"2025_01_10", "2025_01_11"},
			description:   "Should find only daily CSV files",
		},
		{
			name:          "no daily CSV files",
			files:         []string{"data.csv", "report.csv", "summary.csv"},
			expectedDates: []string{},
			description:   "Should return empty map when no daily files found",
		},
		{
			name:          "empty directory",
			files:         []string{},
			expectedDates: []string{},
			description:   "Should handle empty directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			discovery := NewDiscovery(tmpDir)

			testDir := "daily_csv_test"
			fullTestDir := filepath.Join(tmpDir, testDir)
			err := os.MkdirAll(fullTestDir, 0755)
			require.NoError(t, err)

			// Create test files
			for _, filename := range tt.files {
				filePath := filepath.Join(fullTestDir, filename)
				err := os.WriteFile(filePath, []byte("test,csv,content"), 0644)
				require.NoError(t, err)
			}

			dailyFiles, err := discovery.FindDailyCSVFiles(testDir)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, len(tt.expectedDates), len(dailyFiles), tt.description)

			// Verify expected dates are found
			for _, expectedDate := range tt.expectedDates {
				file, exists := dailyFiles[expectedDate]
				assert.True(t, exists, "Expected date %s should be found", expectedDate)
				assert.NotEmpty(t, file.Name)
				assert.NotEmpty(t, file.Path)
				assert.False(t, file.IsDir)
			}
		})
	}
}

func TestListDirectories(t *testing.T) {
	tests := []struct {
		name        string
		directories []string
		files       []string
		expectedDirs int
		description string
	}{
		{
			name:         "only directories",
			directories:  []string{"dir1", "dir2", "dir3"},
			files:        []string{},
			expectedDirs: 3,
			description:  "Should find all directories",
		},
		{
			name:         "mixed directories and files",
			directories:  []string{"subdir1", "subdir2"},
			files:        []string{"file1.txt", "file2.csv"},
			expectedDirs: 2,
			description:  "Should find only directories",
		},
		{
			name:         "no directories",
			directories:  []string{},
			files:        []string{"file1.txt", "file2.csv"},
			expectedDirs: 0,
			description:  "Should find no directories",
		},
		{
			name:         "empty directory",
			directories:  []string{},
			files:        []string{},
			expectedDirs: 0,
			description:  "Should handle empty directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			discovery := NewDiscovery(tmpDir)

			testDir := "list_dirs_test"
			fullTestDir := filepath.Join(tmpDir, testDir)
			err := os.MkdirAll(fullTestDir, 0755)
			require.NoError(t, err)

			// Create test directories
			for _, dirName := range tt.directories {
				dirPath := filepath.Join(fullTestDir, dirName)
				err := os.MkdirAll(dirPath, 0755)
				require.NoError(t, err)
			}

			// Create test files
			for _, fileName := range tt.files {
				filePath := filepath.Join(fullTestDir, fileName)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			dirs, err := discovery.ListDirectories(testDir)
			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedDirs, len(dirs), tt.description)

			// Verify directory properties
			for _, dir := range dirs {
				assert.NotEmpty(t, dir.Name)
				assert.NotEmpty(t, dir.Path)
				assert.True(t, dir.IsDir)
			}
		})
	}
}

func TestGetLatestFile(t *testing.T) {
	tests := []struct {
		name        string
		files       []FileInfo
		expectFound bool
		expectedIdx int
		description string
	}{
		{
			name: "multiple files with different times",
			files: []FileInfo{
				{Name: "old.txt", ModTime: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
				{Name: "latest.txt", ModTime: time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)},
				{Name: "middle.txt", ModTime: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)},
			},
			expectFound: true,
			expectedIdx: 1, // latest.txt
			description: "Should return file with latest modification time",
		},
		{
			name: "single file",
			files: []FileInfo{
				{Name: "only.txt", ModTime: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
			},
			expectFound: true,
			expectedIdx: 0,
			description: "Should return single file",
		},
		{
			name:        "empty slice",
			files:       []FileInfo{},
			expectFound: false,
			expectedIdx: -1,
			description: "Should return false for empty slice",
		},
		{
			name: "files with same time",
			files: []FileInfo{
				{Name: "file1.txt", ModTime: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
				{Name: "file2.txt", ModTime: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)},
			},
			expectFound: true,
			expectedIdx: 0, // Should return first one
			description: "Should return first file when times are equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latest, found := GetLatestFile(tt.files)
			
			assert.Equal(t, tt.expectFound, found, tt.description)
			
			if tt.expectFound {
				expectedFile := tt.files[tt.expectedIdx]
				assert.Equal(t, expectedFile.Name, latest.Name)
				assert.Equal(t, expectedFile.ModTime, latest.ModTime)
			}
		})
	}
}

func TestFilterFilesByDateRange(t *testing.T) {
	baseTime := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	files := []FileInfo{
		{Name: "file1.txt", ModTime: baseTime.Add(-2 * 24 * time.Hour)}, // 2025-01-08
		{Name: "file2.txt", ModTime: baseTime.Add(-1 * 24 * time.Hour)}, // 2025-01-09
		{Name: "file3.txt", ModTime: baseTime},                          // 2025-01-10
		{Name: "file4.txt", ModTime: baseTime.Add(1 * 24 * time.Hour)},  // 2025-01-11
		{Name: "file5.txt", ModTime: baseTime.Add(2 * 24 * time.Hour)},  // 2025-01-12
	}

	tests := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		expectedFiles []string
		description   string
	}{
		{
			name:          "middle range",
			startDate:     baseTime.Add(-1*24*time.Hour - time.Hour), // Just before 2025-01-09
			endDate:       baseTime.Add(1*24*time.Hour + time.Hour),  // Just after 2025-01-11
			expectedFiles: []string{"file2.txt", "file3.txt", "file4.txt"},
			description:   "Should filter files within date range",
		},
		{
			name:          "no files in range",
			startDate:     baseTime.Add(10 * 24 * time.Hour), // Far future
			endDate:       baseTime.Add(20 * 24 * time.Hour), // Far future
			expectedFiles: []string{},
			description:   "Should return empty when no files in range",
		},
		{
			name:          "all files in range",
			startDate:     baseTime.Add(-10 * 24 * time.Hour), // Far past
			endDate:       baseTime.Add(10 * 24 * time.Hour),  // Far future
			expectedFiles: []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"},
			description:   "Should return all files when range covers all",
		},
		{
			name:          "single day range",
			startDate:     baseTime.Add(-time.Hour),  // Just before base time
			endDate:       baseTime.Add(time.Hour),   // Just after base time
			expectedFiles: []string{"file3.txt"},
			description:   "Should filter files within single day range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterFilesByDateRange(files, tt.startDate, tt.endDate)
			
			assert.Equal(t, len(tt.expectedFiles), len(filtered), tt.description)
			
			// Verify expected files are in the result
			for i, expectedFile := range tt.expectedFiles {
				if i < len(filtered) {
					assert.Equal(t, expectedFile, filtered[i].Name)
				}
			}
		})
	}
}

func TestAbsolutePaths(t *testing.T) {
	tmpDir := t.TempDir()
	discovery := NewDiscovery("/base/path") // Different from tmpDir

	// Create test directory with absolute path
	testDir := filepath.Join(tmpDir, "absolute_test")
	err := os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Create test files
	testFiles := []string{"test1.xlsx", "test2.csv"}
	for _, filename := range testFiles {
		filePath := filepath.Join(testDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	t.Run("FindExcelFiles with absolute path", func(t *testing.T) {
		files, err := discovery.FindExcelFiles(testDir) // Using absolute path
		assert.NoError(t, err)
		assert.Equal(t, 1, len(files)) // Only .xlsx files
	})

	t.Run("FindCSVFiles with absolute path", func(t *testing.T) {
		files, err := discovery.FindCSVFiles(testDir) // Using absolute path
		assert.NoError(t, err)
		assert.Equal(t, 1, len(files)) // Only .csv files
	})

	t.Run("ListDirectories with absolute path", func(t *testing.T) {
		dirs, err := discovery.ListDirectories(tmpDir) // Parent directory
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(dirs), 1) // Should find at least the test directory
	})
}

func TestErrorHandling(t *testing.T) {
	discovery := NewDiscovery("/base/path")

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := discovery.FindExcelFiles("/non/existent/directory")
		assert.Error(t, err)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := discovery.FindFilesByPattern(tmpDir, "[invalid")
		assert.Error(t, err)
	})
}

// Benchmark file discovery operations
func BenchmarkFindExcelFiles(b *testing.B) {
	tmpDir := b.TempDir()
	discovery := NewDiscovery(tmpDir)

	// Create many test files
	testDir := filepath.Join(tmpDir, "benchmark_test")
	os.MkdirAll(testDir, 0755)

	for i := 0; i < 100; i++ {
		filename := filepath.Join(testDir, fmt.Sprintf("file_%03d.xlsx", i))
		os.WriteFile(filename, []byte("test"), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = discovery.FindExcelFiles("benchmark_test")
	}
}