package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Manager provides file management operations
type Manager struct {
	basePath string
}

// NewManager creates a new file manager instance
func NewManager(basePath string) *Manager {
	return &Manager{basePath: basePath}
}

// FileExists checks if a file exists at the given path
func (m *Manager) FileExists(path string) bool {
	fullPath := filepath.Join(m.basePath, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// CreateDirectory creates a directory with all parent directories
func (m *Manager) CreateDirectory(path string) error {
	fullPath := filepath.Join(m.basePath, path)
	return os.MkdirAll(fullPath, 0755)
}

// CopyFile copies a file from source to destination
func (m *Manager) CopyFile(src, dst string) error {
	srcPath := filepath.Join(m.basePath, src)
	dstPath := filepath.Join(m.basePath, dst)
	
	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()
	
	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()
	
	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	
	// Sync to ensure write is complete
	return dstFile.Sync()
}

// MoveFile moves a file from source to destination
func (m *Manager) MoveFile(src, dst string) error {
	srcPath := filepath.Join(m.basePath, src)
	dstPath := filepath.Join(m.basePath, dst)
	
	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Try rename first (atomic if on same filesystem)
	if err := os.Rename(srcPath, dstPath); err == nil {
		return nil
	}
	
	// Fall back to copy and delete
	if err := m.CopyFile(src, dst); err != nil {
		return err
	}
	
	return os.Remove(srcPath)
}

// DeleteFile deletes a file
func (m *Manager) DeleteFile(path string) error {
	fullPath := filepath.Join(m.basePath, path)
	return os.Remove(fullPath)
}

// GetFileSize returns the size of a file in bytes
func (m *Manager) GetFileSize(path string) (int64, error) {
	fullPath := filepath.Join(m.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ReadFile reads the entire content of a file
func (m *Manager) ReadFile(path string) ([]byte, error) {
	fullPath := filepath.Join(m.basePath, path)
	return os.ReadFile(fullPath)
}

// WriteFile writes data to a file
func (m *Manager) WriteFile(path string, data []byte) error {
	fullPath := filepath.Join(m.basePath, path)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return os.WriteFile(fullPath, data, 0644)
}

// CleanPath returns a clean, absolute path
func (m *Manager) CleanPath(path string) string {
	return filepath.Clean(filepath.Join(m.basePath, path))
}

// GetRelativePath returns the path relative to the base path
func (m *Manager) GetRelativePath(fullPath string) (string, error) {
	return filepath.Rel(m.basePath, fullPath)
}

// ListFiles returns all files in a directory (non-recursive)
func (m *Manager) ListFiles(dir string) ([]string, error) {
	fullPath := filepath.Join(m.basePath, dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// EnsureDirectory creates a directory if it doesn't exist
func (m *Manager) EnsureDirectory(path string) error {
	fullPath := filepath.Join(m.basePath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return os.MkdirAll(fullPath, 0755)
	}
	return nil
}