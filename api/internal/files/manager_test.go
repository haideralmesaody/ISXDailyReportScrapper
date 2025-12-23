package files

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/isxcli/isxcli/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	paths := &config.Paths{
		ExecutableDir: "/test/executable",
		DataDir:      "/test/data",
	}
	
	manager := NewManager(paths)
	assert.NotNil(t, manager)
	assert.Equal(t, paths, manager.paths)
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name           string
		setupFile      bool
		relativePath   string
		expectedExists bool
	}{
		{
			name:           "existing file",
			setupFile:      true,
			relativePath:   "test_file.txt",
			expectedExists: true,
		},
		{
			name:           "non-existing file",
			setupFile:      false,
			relativePath:   "non_existing.txt",
			expectedExists: false,
		},
		{
			name:           "absolute path existing",
			setupFile:      true,
			relativePath:   "", // Will be set to absolute path
			expectedExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			var testPath string
			if tt.relativePath == "" {
				// Test absolute path
				testPath = filepath.Join(tmpDir, "absolute_test.txt")
			} else {
				testPath = tt.relativePath
			}

			if tt.setupFile {
				fullPath := testPath
				if !filepath.IsAbs(testPath) {
					fullPath = filepath.Join(tmpDir, testPath)
				}
				err := os.WriteFile(fullPath, []byte("test content"), 0644)
				require.NoError(t, err)
				
				if tt.relativePath == "" {
					testPath = fullPath // Use absolute path for test
				}
			}

			exists := manager.FileExists(testPath)
			assert.Equal(t, tt.expectedExists, exists)
		})
	}
}

func TestCreateDirectory(t *testing.T) {
	tests := []struct {
		name        string
		dirPath     string
		expectError bool
	}{
		{
			name:        "simple directory",
			dirPath:     "test_dir",
			expectError: false,
		},
		{
			name:        "nested directory",
			dirPath:     "parent/child/grandchild",
			expectError: false,
		},
		{
			name:        "absolute path directory",
			dirPath:     "", // Will be set to absolute path
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			testPath := tt.dirPath
			if testPath == "" {
				testPath = filepath.Join(tmpDir, "absolute_test_dir")
			}

			err := manager.CreateDirectory(testPath)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify directory was created
				fullPath := testPath
				if !filepath.IsAbs(testPath) {
					fullPath = filepath.Join(tmpDir, testPath)
				}
				
				info, statErr := os.Stat(fullPath)
				assert.NoError(t, statErr)
				assert.True(t, info.IsDir())
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		srcContent  string
		expectError bool
		description string
	}{
		{
			name:        "simple text file",
			srcContent:  "Hello, World!",
			expectError: false,
			description: "Should copy simple text file successfully",
		},
		{
			name:        "binary content",
			srcContent:  "\x00\x01\x02\x03\xFF",
			expectError: false,
			description: "Should copy binary content successfully",
		},
		{
			name:        "empty file",
			srcContent:  "",
			expectError: false,
			description: "Should copy empty file successfully",
		},
		{
			name:        "large content",
			srcContent:  strings.Repeat("Large content test. ", 1000),
			expectError: false,
			description: "Should copy large content successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			// Create source file
			srcPath := filepath.Join(tmpDir, "source.txt")
			err := os.WriteFile(srcPath, []byte(tt.srcContent), 0644)
			require.NoError(t, err)

			// Test copy operation
			dstPath := "copied_file.txt"
			err = manager.CopyFile(srcPath, dstPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, tt.description)
				
				// Verify file was copied correctly
				fullDstPath := filepath.Join(tmpDir, dstPath)
				copiedContent, err := os.ReadFile(fullDstPath)
				assert.NoError(t, err)
				assert.Equal(t, tt.srcContent, string(copiedContent), tt.description)
				
				// Verify file permissions and metadata
				srcInfo, err := os.Stat(srcPath)
				assert.NoError(t, err)
				dstInfo, err := os.Stat(fullDstPath)
				assert.NoError(t, err)
				assert.Equal(t, srcInfo.Size(), dstInfo.Size())
			}
		})
	}
}

func TestMoveFile(t *testing.T) {
	tests := []struct {
		name            string
		srcContent      string
		crossFileSystem bool
		expectError     bool
		description     string
	}{
		{
			name:        "simple move within same directory",
			srcContent:  "Move test content",
			expectError: false,
			description: "Should move file successfully within same directory",
		},
		{
			name:        "move to subdirectory",
			srcContent:  "Move to subdirectory content",
			expectError: false,
			description: "Should move file to subdirectory successfully",
		},
		{
			name:        "move empty file",
			srcContent:  "",
			expectError: false,
			description: "Should move empty file successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			// Create source file
			srcPath := filepath.Join(tmpDir, "source_move.txt")
			err := os.WriteFile(srcPath, []byte(tt.srcContent), 0644)
			require.NoError(t, err)

			// Test move operation
			dstPath := "subdir/moved_file.txt"
			err = manager.MoveFile(srcPath, dstPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, tt.description)
				
				// Verify source file no longer exists
				_, err = os.Stat(srcPath)
				assert.True(t, os.IsNotExist(err), "Source file should not exist after move")
				
				// Verify destination file exists with correct content
				fullDstPath := filepath.Join(tmpDir, dstPath)
				movedContent, err := os.ReadFile(fullDstPath)
				assert.NoError(t, err)
				assert.Equal(t, tt.srcContent, string(movedContent), tt.description)
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   bool
		expectError bool
	}{
		{
			name:        "delete existing file",
			setupFile:   true,
			expectError: false,
		},
		{
			name:        "delete non-existing file",
			setupFile:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			testPath := "test_delete.txt"
			fullPath := filepath.Join(tmpDir, testPath)

			if tt.setupFile {
				err := os.WriteFile(fullPath, []byte("delete me"), 0644)
				require.NoError(t, err)
			}

			err := manager.DeleteFile(testPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify file no longer exists
				_, statErr := os.Stat(fullPath)
				assert.True(t, os.IsNotExist(statErr))
			}
		})
	}
}

func TestGetFileSize(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedSize int64
		expectError  bool
	}{
		{
			name:         "small file",
			content:      "Hello",
			expectedSize: 5,
			expectError:  false,
		},
		{
			name:         "empty file",
			content:      "",
			expectedSize: 0,
			expectError:  false,
		},
		{
			name:         "larger file",
			content:      strings.Repeat("A", 1024),
			expectedSize: 1024,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			testPath := "size_test.txt"
			fullPath := filepath.Join(tmpDir, testPath)
			err := os.WriteFile(fullPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			size, err := manager.GetFileSize(testPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSize, size)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "read text file",
			content:     "Test content for reading",
			expectError: false,
		},
		{
			name:        "read binary file",
			content:     "\x00\x01\x02\x03\xFF",
			expectError: false,
		},
		{
			name:        "read empty file",
			content:     "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			testPath := "read_test.txt"
			fullPath := filepath.Join(tmpDir, testPath)
			err := os.WriteFile(fullPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			data, err := manager.ReadFile(testPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.content, string(data))
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		expectError bool
	}{
		{
			name:        "write text content",
			content:     []byte("Hello, World!"),
			expectError: false,
		},
		{
			name:        "write binary content",
			content:     []byte{0x00, 0x01, 0x02, 0xFF},
			expectError: false,
		},
		{
			name:        "write empty content",
			content:     []byte{},
			expectError: false,
		},
		{
			name:        "write large content",
			content:     []byte(strings.Repeat("Large content test. ", 1000)),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			testPath := "write_test.txt"
			err := manager.WriteFile(testPath, tt.content)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify file was written correctly
				fullPath := filepath.Join(tmpDir, testPath)
				writtenContent, err := os.ReadFile(fullPath)
				assert.NoError(t, err)
				assert.Equal(t, tt.content, writtenContent)
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		directories   []string
		expectedFiles []string
	}{
		{
			name:          "empty directory",
			files:         []string{},
			directories:   []string{},
			expectedFiles: []string{},
		},
		{
			name:          "files only",
			files:         []string{"file1.txt", "file2.csv", "file3.json"},
			directories:   []string{},
			expectedFiles: []string{"file1.txt", "file2.csv", "file3.json"},
		},
		{
			name:          "mixed files and directories",
			files:         []string{"file1.txt", "file2.csv"},
			directories:   []string{"subdir1", "subdir2"},
			expectedFiles: []string{"file1.txt", "file2.csv"},
		},
		{
			name:          "directories only",
			files:         []string{},
			directories:   []string{"dir1", "dir2", "dir3"},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      tmpDir,
			}
			manager := NewManager(paths)

			testDir := "test_list_dir"
			fullTestDir := filepath.Join(tmpDir, testDir)
			err := os.MkdirAll(fullTestDir, 0755)
			require.NoError(t, err)

			// Create test files
			for _, file := range tt.files {
				filePath := filepath.Join(fullTestDir, file)
				err := os.WriteFile(filePath, []byte("test"), 0644)
				require.NoError(t, err)
			}

			// Create test directories
			for _, dir := range tt.directories {
				dirPath := filepath.Join(fullTestDir, dir)
				err := os.MkdirAll(dirPath, 0755)
				require.NoError(t, err)
			}

			files, err := manager.ListFiles(testDir)
			assert.NoError(t, err)
			
			// Sort both slices for comparison
			assert.ElementsMatch(t, tt.expectedFiles, files)
		})
	}
}

func TestPathResolution(t *testing.T) {
	tests := []struct {
		name         string
		inputPath    string
		pathsConfig  *config.Paths
		expectedFunc func(*config.Paths, string) string
		description  string
	}{
		{
			name:      "downloads prefix",
			inputPath: "downloads/test.xlsx",
			expectedFunc: func(p *config.Paths, subPath string) string {
				return p.GetDownloadPath("test.xlsx")
			},
			description: "Should resolve downloads/ prefix correctly",
		},
		{
			name:      "reports prefix",
			inputPath: "reports/output.csv",
			expectedFunc: func(p *config.Paths, subPath string) string {
				return p.GetReportPath("output.csv")
			},
			description: "Should resolve reports/ prefix correctly",
		},
		{
			name:      "cache prefix",
			inputPath: "cache/temp.dat",
			expectedFunc: func(p *config.Paths, subPath string) string {
				return p.GetCachePath("temp.dat")
			},
			description: "Should resolve cache/ prefix correctly",
		},
		{
			name:      "logs prefix",
			inputPath: "logs/app.log",
			expectedFunc: func(p *config.Paths, subPath string) string {
				return p.GetLogPath("app.log")
			},
			description: "Should resolve logs/ prefix correctly",
		},
		{
			name:      "absolute path",
			inputPath: "__ABS__",
			expectedFunc: func(p *config.Paths, subPath string) string {
				return subPath
			},
			description: "Should return absolute path unchanged",
		},
		{
			name:      "default data directory",
			inputPath: "somefile.txt",
			expectedFunc: func(p *config.Paths, subPath string) string {
				return filepath.Join(p.DataDir, "somefile.txt")
			},
			description: "Should resolve to data directory for unknown prefixes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Create mock paths
			paths := &config.Paths{
				ExecutableDir: tmpDir,
				DataDir:      filepath.Join(tmpDir, "data"),
				DownloadsDir: filepath.Join(tmpDir, "data", "downloads"),
				ReportsDir:   filepath.Join(tmpDir, "data", "reports"),
				CacheDir:     filepath.Join(tmpDir, "data", "cache"),
				LogsDir:      filepath.Join(tmpDir, "logs"),
				WebDir:       filepath.Join(tmpDir, "web"),
				StaticDir:    filepath.Join(tmpDir, "static"),
			}

			manager := NewManager(paths)

			inputPath := tt.inputPath
			if tt.name == "absolute path" {
				inputPath = filepath.Join(tmpDir, "absolute", "path", "file.txt")
			}

			// Test the path resolution
			resolved := manager.resolvePath(inputPath)
			expected := tt.expectedFunc(paths, inputPath)
			
			assert.Equal(t, expected, resolved, tt.description)
		})
	}
}

func TestConcurrentFileOperations(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &config.Paths{
		ExecutableDir: tmpDir,
		DataDir:      tmpDir,
	}
	manager := NewManager(paths)

	const numGoroutines = 10
	var wg sync.WaitGroup
	
	// Test concurrent file creation
	t.Run("concurrent file creation", func(t *testing.T) {
		wg.Add(numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				
				filename := fmt.Sprintf("concurrent_%d.txt", id)
				content := fmt.Sprintf("Content for file %d", id)
				
				err := manager.WriteFile(filename, []byte(content))
				assert.NoError(t, err)
				
				// Verify file was written
				exists := manager.FileExists(filename)
				assert.True(t, exists)
			}(i)
		}
		
		wg.Wait()
		
		// Verify all files exist
		for i := 0; i < numGoroutines; i++ {
			filename := fmt.Sprintf("concurrent_%d.txt", i)
			exists := manager.FileExists(filename)
			assert.True(t, exists)
		}
	})

	// Test concurrent read operations
	t.Run("concurrent file reading", func(t *testing.T) {
		// Create a shared file first
		sharedFile := "shared_file.txt"
		sharedContent := "Shared content for concurrent reading"
		err := manager.WriteFile(sharedFile, []byte(sharedContent))
		require.NoError(t, err)
		
		wg.Add(numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				
				data, err := manager.ReadFile(sharedFile)
				assert.NoError(t, err)
				assert.Equal(t, sharedContent, string(data))
			}(i)
		}
		
		wg.Wait()
	})
}

func TestManagerErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	paths := &config.Paths{
		ExecutableDir: tmpDir,
		DataDir:      tmpDir,
	}
	manager := NewManager(paths)

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := manager.ReadFile("non_existent.txt")
		assert.Error(t, err)
	})

	t.Run("copy non-existent source", func(t *testing.T) {
		err := manager.CopyFile("non_existent.txt", "destination.txt")
		assert.Error(t, err)
	})

	t.Run("move non-existent source", func(t *testing.T) {
		err := manager.MoveFile("non_existent.txt", "destination.txt")
		assert.Error(t, err)
	})

	t.Run("get size of non-existent file", func(t *testing.T) {
		_, err := manager.GetFileSize("non_existent.txt")
		assert.Error(t, err)
	})

	t.Run("list files in non-existent directory", func(t *testing.T) {
		_, err := manager.ListFiles("non_existent_dir")
		assert.Error(t, err)
	})
}

// Disable slog output during tests to reduce noise
func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
}
