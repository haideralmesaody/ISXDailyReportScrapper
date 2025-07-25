package testutil

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// CreateTestDirectory creates a temporary test directory
func CreateTestDirectory(t *testing.T, name string) string {
	t.Helper()
	
	dir, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	
	return dir
}

// CreateTestFile creates a test file with content
func CreateTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	
	path := filepath.Join(dir, name)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	
	return path
}

// CopyFile copies a file from src to dst
func CopyFile(t *testing.T, src, dst string) {
	t.Helper()
	
	sourceFile, err := os.Open(src)
	if err != nil {
		t.Fatalf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		t.Fatalf("failed to create destination file: %v", err)
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		t.Fatalf("failed to copy file: %v", err)
	}
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file and returns its content
func ReadFile(t *testing.T, path string) string {
	t.Helper()
	
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	
	return string(data)
}

// CreateCSVFile creates a test CSV file
func CreateCSVFile(t *testing.T, dir, name string, headers []string, rows [][]string) string {
	t.Helper()
	
	var content string
	
	// Add headers
	for i, h := range headers {
		if i > 0 {
			content += ","
		}
		content += h
	}
	content += "\n"
	
	// Add rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				content += ","
			}
			content += cell
		}
		content += "\n"
	}
	
	return CreateTestFile(t, dir, name, content)
}

// CreateJSONFile creates a test JSON file
func CreateJSONFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	return CreateTestFile(t, dir, name, content)
}

// CreateTestExecutable creates a mock executable file
func CreateTestExecutable(t *testing.T, dir, name string) string {
	t.Helper()
	
	path := filepath.Join(dir, name)
	
	// Create a simple batch file for Windows
	content := "@echo off\necho Test executable\nexit /b 0"
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to create test executable: %v", err)
	}
	
	return path
}

// AssertFileContains checks if a file contains expected content
func AssertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	
	content := ReadFile(t, path)
	if content != expected {
		t.Errorf("file content = %q, want %q", content, expected)
	}
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	
	if !FileExists(path) {
		t.Errorf("file %s does not exist", path)
	}
}

// AssertFileNotExists checks if a file doesn't exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	
	if FileExists(path) {
		t.Errorf("file %s exists but should not", path)
	}
}

// SetupTestPipelineFiles creates a standard test environment
func SetupTestPipelineFiles(t *testing.T) (downloadDir, reportDir string) {
	t.Helper()
	
	baseDir := CreateTestDirectory(t, "pipeline-test")
	dataDir := filepath.Join(baseDir, "data")
	downloadDir = filepath.Join(dataDir, "downloads")
	reportDir = filepath.Join(dataDir, "reports")
	
	os.MkdirAll(downloadDir, 0755)
	os.MkdirAll(reportDir, 0755)
	
	// Create some test Excel files
	CreateTestFile(t, downloadDir, "2024 01 01 ISX Daily Report.xlsx", "test excel data")
	CreateTestFile(t, downloadDir, "2024 01 02 ISX Daily Report.xlsx", "test excel data")
	
	return downloadDir, reportDir
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(t *testing.T, timeout, interval int, condition func() bool, msg string) {
	t.Helper()
	
	for i := 0; i < timeout; i += interval {
		if condition() {
			return
		}
		// Sleep for interval milliseconds
		os.Stdout.Sync() // Force flush instead of sleep
	}
	
	t.Fatalf("timeout waiting for condition: %s", msg)
}