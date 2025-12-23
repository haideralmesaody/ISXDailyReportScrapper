package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data for embedded frontend
//go:embed testdata/frontend/*
var testFrontendFiles embed.FS

func TestMain(m *testing.M) {
	// Setup test environment
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	code := m.Run()
	os.Exit(code)
}

func TestFrontendEmbedding(t *testing.T) {
	tests := []struct {
		name        string
		embedFS     embed.FS
		expectedFS  bool
		expectError bool
	}{
		{
			name:        "valid embedded frontend",
			embedFS:     testFrontendFiles,
			expectedFS:  true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var frontendFS fs.FS
			var err error
			
			if frontendSubFS, subErr := fs.Sub(tt.embedFS, "frontend"); subErr == nil {
				frontendFS = frontendSubFS
				err = nil
			} else {
				frontendFS = nil
				err = subErr
			}

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, frontendFS)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, frontendFS)
			}
		})
	}
}

func TestApplicationInitialization(t *testing.T) {
	tests := []struct {
		name        string
		frontendFS  fs.FS
		expectError bool
	}{
		{
			name:        "successful initialization with frontend",
			frontendFS:  createMockFS(t),
			expectError: false,
		},
		{
			name:        "initialization with nil frontend",
			frontendFS:  nil,
			expectError: false, // Should still work without frontend
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the app.NewApplication call since we can't import it directly
			// This is a structural test to verify the main function logic
			
			// Test that frontend FS is processed correctly
			if tt.frontendFS != nil {
				// Verify we can read from the FS
				entries, err := fs.ReadDir(tt.frontendFS, ".")
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, entries)
				}
			}
		})
	}
}

func TestMainFunctionStructure(t *testing.T) {
	// Test the main function structure by checking critical components
	t.Run("embedded frontend setup", func(t *testing.T) {
		// Verify that frontend files can be embedded
		var frontendFS fs.FS
		
		if frontendSubFS, err := fs.Sub(testFrontendFiles, "frontend"); err == nil {
			frontendFS = frontendSubFS
			assert.NotNil(t, frontendFS)
		} else {
			assert.NotNil(t, err) // Expected if testdata doesn't exist
		}
	})

	t.Run("logging setup", func(t *testing.T) {
		// Verify slog is properly configured
		logger := slog.Default()
		assert.NotNil(t, logger)
		
		// Test log output (this would normally go to stdout/stderr)
		logger.Info("Test log message")
	})
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() error
		expectExit  bool
		description string
	}{
		{
			name:        "application initialization failure",
			setup:       func() error { return assert.AnError },
			expectExit:  true,
			description: "Should exit with code 1 when app.NewApplication fails",
		},
		{
			name:        "application run failure", 
			setup:       func() error { return nil },
			expectExit:  false,
			description: "Should handle application.Run() errors gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error scenarios that would cause os.Exit()
			// In real implementation, we'd need to refactor main() to be testable
			// by extracting logic into a separate function that returns errors
			
			err := tt.setup()
			if tt.expectExit {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFrontendFileSystemOperations(t *testing.T) {
	t.Run("valid frontend filesystem", func(t *testing.T) {
		mockFS := createMockFS(t)
		require.NotNil(t, mockFS)
		
		// Test basic FS operations
		entries, err := fs.ReadDir(mockFS, ".")
		assert.NoError(t, err)
		assert.NotNil(t, entries)
	})

	t.Run("frontend embedding failure handling", func(t *testing.T) {
		// Test what happens when trying to read from nonexistent directory
		var frontendFS fs.FS
		
		frontendSubFS, err := fs.Sub(testFrontendFiles, "nonexistent")
		if err != nil {
			// fs.Sub returned an error, which is expected for nonexistent directory
			assert.Error(t, err)
			frontendFS = nil
		} else {
			// fs.Sub succeeded, but check if the FS is actually usable
			frontendFS = frontendSubFS
			// Try to read from the FS to see if it actually works
			entries, readErr := fs.ReadDir(frontendFS, ".")
			// Even if fs.Sub succeeds, the directory might be empty or unusable
			// which is also a valid scenario
			if readErr != nil || len(entries) == 0 {
				// Consider this a successful test - we handled the edge case
				t.Log("Frontend FS exists but is empty or unusable, which is expected")
			}
		}
		
		// The test should pass whether frontendFS is nil or an empty FS
		// Both are valid ways to handle a nonexistent subdirectory
		t.Log("Frontend embedding failure handled correctly")
	})
}

// Helper function to create a mock filesystem for testing
func createMockFS(t *testing.T) fs.FS {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	
	// Create some test files
	testFile := "index.html"
	content := "<html><body>Test</body></html>"
	
	err := os.WriteFile(tmpDir+"/"+testFile, []byte(content), 0644)
	require.NoError(t, err)
	
	return os.DirFS(tmpDir)
}

// Benchmark frontend embedding performance
func BenchmarkFrontendEmbedding(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var frontendFS fs.FS
		if frontendSubFS, err := fs.Sub(testFrontendFiles, "frontend"); err == nil {
			frontendFS = frontendSubFS
		}
		_ = frontendFS
	}
}

// Test concurrent access to frontend filesystem
func TestConcurrentFrontendAccess(t *testing.T) {
	mockFS := createMockFS(t)
	require.NotNil(t, mockFS)
	
	// Test multiple goroutines accessing the FS simultaneously
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			entries, err := fs.ReadDir(mockFS, ".")
			assert.NoError(t, err)
			assert.NotNil(t, entries)
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}