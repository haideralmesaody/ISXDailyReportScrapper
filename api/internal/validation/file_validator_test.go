package validation

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileValidator_ValidateInputDirectory(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(t *testing.T) string
		requiredPattern string
		wantErr         bool
		errorContains   string
	}{
		{
			name: "valid directory with files",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "test.xlsx")
				require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				return dir
			},
			requiredPattern: "*.xlsx",
			wantErr:         false,
		},
		{
			name: "valid directory without files",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			requiredPattern: "*.xlsx",
			wantErr:         false, // No files is not an error
		},
		{
			name: "non-existent directory",
			setupFunc: func(t *testing.T) string {
				return "/non/existent/path"
			},
			requiredPattern: "",
			wantErr:         true,
			errorContains:   "does not exist",
		},
		{
			name: "path is file not directory",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "test.txt")
				require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				return file
			},
			requiredPattern: "",
			wantErr:         true,
			errorContains:   "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFileValidator(slog.Default())
			dir := tt.setupFunc(t)
			
			err := validator.ValidateInputDirectory(dir, tt.requiredPattern)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileValidator_ValidateOutputDirectory(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		wantErr       bool
		errorContains string
	}{
		{
			name: "existing directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: false,
		},
		{
			name: "non-existent directory (should be created)",
			setupFunc: func(t *testing.T) string {
				base := t.TempDir()
				return filepath.Join(base, "new", "nested", "dir")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFileValidator(slog.Default())
			dir := tt.setupFunc(t)
			
			err := validator.ValidateOutputDirectory(dir)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				// Verify directory exists
				info, err := os.Stat(dir)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			}
		})
	}
}

func TestFileValidator_ValidateExcelFile(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		wantErr       bool
		errorContains string
	}{
		{
			name: "valid Excel file (.xlsx)",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "test.xlsx")
				require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				return file
			},
			wantErr: false,
		},
		{
			name: "valid Excel file (.xls)",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "test.xls")
				require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				return file
			},
			wantErr: false,
		},
		{
			name: "temp Excel file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "~$test.xlsx")
				require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				return file
			},
			wantErr:       true,
			errorContains: "temporary",
		},
		{
			name: "non-Excel file",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				file := filepath.Join(dir, "test.txt")
				require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				return file
			},
			wantErr:       true,
			errorContains: "not an Excel file",
		},
		{
			name: "non-existent file",
			setupFunc: func(t *testing.T) string {
				return "/non/existent/file.xlsx"
			},
			wantErr:       true,
			errorContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFileValidator(slog.Default())
			file := tt.setupFunc(t)
			
			err := validator.ValidateExcelFile(file)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileValidator_CountFiles(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		pattern   string
		wantCount int
		wantErr   bool
	}{
		{
			name: "count Excel files",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				for i := 0; i < 3; i++ {
					file := filepath.Join(dir, fmt.Sprintf("file%d.xlsx", i))
					require.NoError(t, os.WriteFile(file, []byte("test"), 0644))
				}
				// Add non-Excel file
				require.NoError(t, os.WriteFile(filepath.Join(dir, "other.txt"), []byte("test"), 0644))
				return dir
			},
			pattern:   "*.xlsx",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "no matching files",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			pattern:   "*.xlsx",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "exclude directories",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				// Create file
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644))
				// Create directory
				require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0755))
				return dir
			},
			pattern:   "*",
			wantCount: 1, // Only the file, not the directory
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFileValidator(slog.Default())
			dir := tt.setupFunc(t)
			
			count, err := validator.CountFiles(dir, tt.pattern)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
}

func TestFileValidator_CreateEmptyCSV(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		wantErr bool
	}{
		{
			name:    "with headers",
			headers: []string{"Date", "Symbol", "Price"},
			wantErr: false,
		},
		{
			name:    "without headers",
			headers: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewFileValidator(slog.Default())
			dir := t.TempDir()
			csvPath := filepath.Join(dir, "test.csv")
			
			err := validator.CreateEmptyCSV(csvPath, tt.headers)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify file exists
				assert.FileExists(t, csvPath)
				
				// Verify content
				content, err := os.ReadFile(csvPath)
				assert.NoError(t, err)
				
				if len(tt.headers) > 0 {
					expectedHeader := strings.Join(tt.headers, ",") + "\n"
					assert.Equal(t, expectedHeader, string(content))
				} else {
					assert.Empty(t, content)
				}
			}
		})
	}
}