package files

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/isxcli/isxcli/internal/config"
)

// Manager provides file management operations
type Manager struct {
	paths *config.Paths
}

// NewManager creates a new file manager instance
func NewManager(paths *config.Paths) *Manager {
	return &Manager{paths: paths}
}

// FileExists checks if a file exists at the given path
func (m *Manager) FileExists(path string) bool {
	fullPath := m.resolvePath(path)
	_, err := os.Stat(fullPath)
	exists := err == nil
	
	slog.Debug("FileExists check",
		slog.String("path", path),
		slog.String("full_path", fullPath),
		slog.Bool("exists", exists))
	
	return exists
}

// CreateDirectory creates a directory with all parent directories
func (m *Manager) CreateDirectory(path string) error {
	fullPath := m.resolvePath(path)
	
	slog.Info("Creating directory",
		slog.String("path", path),
		slog.String("full_path", fullPath))
	
	return os.MkdirAll(fullPath, 0755)
}

// CopyFile copies a file from source to destination
func (m *Manager) CopyFile(src, dst string) error {
	srcPath := m.resolvePath(src)
	dstPath := m.resolvePath(dst)
	
	slog.Info("Copying file",
		slog.String("src", src),
		slog.String("src_path", srcPath),
		slog.String("dst", dst),
		slog.String("dst_path", dstPath))
	
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
	srcPath := m.resolvePath(src)
	dstPath := m.resolvePath(dst)
	
	slog.Info("Moving file",
		slog.String("src", src),
		slog.String("src_path", srcPath),
		slog.String("dst", dst),
		slog.String("dst_path", dstPath))
	
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
	fullPath := m.resolvePath(path)
	
	slog.Info("Deleting file",
		slog.String("path", path),
		slog.String("full_path", fullPath))
	
	return os.Remove(fullPath)
}

// GetFileSize returns the size of a file in bytes
func (m *Manager) GetFileSize(path string) (int64, error) {
	fullPath := m.resolvePath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ReadFile reads the entire content of a file
func (m *Manager) ReadFile(path string) ([]byte, error) {
	fullPath := m.resolvePath(path)
	
	slog.Debug("Reading file",
		slog.String("path", path),
		slog.String("full_path", fullPath))
	
	return os.ReadFile(fullPath)
}

// WriteFile writes data to a file
func (m *Manager) WriteFile(path string, data []byte) error {
	fullPath := m.resolvePath(path)
	
	slog.Info("Writing file",
		slog.String("path", path),
		slog.String("full_path", fullPath),
		slog.Int("size_bytes", len(data)))
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return os.WriteFile(fullPath, data, 0644)
}

// CleanPath returns a clean, absolute path
func (m *Manager) CleanPath(path string) string {
	return filepath.Clean(m.resolvePath(path))
}

// GetRelativePath returns the path relative to the base path
func (m *Manager) GetRelativePath(fullPath string) (string, error) {
	return filepath.Rel(m.paths.ExecutableDir, fullPath)
}

// ListFiles returns all files in a directory (non-recursive)
func (m *Manager) ListFiles(dir string) ([]string, error) {
	fullPath := m.resolvePath(dir)
	
	slog.Debug("Listing files",
		slog.String("dir", dir),
		slog.String("full_path", fullPath))
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
	fullPath := m.resolvePath(path)
	
	slog.Debug("Ensuring directory exists",
		slog.String("path", path),
		slog.String("full_path", fullPath))
	
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return os.MkdirAll(fullPath, 0755)
	}
	return nil
}

// resolvePath resolves a path relative to the appropriate base directory
func (m *Manager) resolvePath(path string) string {
	// If the path is already absolute, return it as-is
	if filepath.IsAbs(path) {
		return path
	}
	
	// Determine which directory to use based on the path
	switch {
	case strings.HasPrefix(path, "downloads/"):
		return m.paths.GetDownloadPath(strings.TrimPrefix(path, "downloads/"))
	case strings.HasPrefix(path, "reports/"):
		return m.paths.GetReportPath(strings.TrimPrefix(path, "reports/"))
	case strings.HasPrefix(path, "cache/"):
		return m.paths.GetCachePath(strings.TrimPrefix(path, "cache/"))
	case strings.HasPrefix(path, "logs/"):
		return m.paths.GetLogPath(strings.TrimPrefix(path, "logs/"))
	case strings.HasPrefix(path, "web/"):
		return m.paths.GetWebFilePath(strings.TrimPrefix(path, "web/"))
	case strings.HasPrefix(path, "static/"):
		return m.paths.GetStaticFilePath(strings.TrimPrefix(path, "static/"))
	default:
		// For files in the data directory
		return filepath.Join(m.paths.DataDir, path)
	}
}