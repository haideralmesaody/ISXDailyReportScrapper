package operations

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileDetector provides centralized file detection logic following SSOT principle
// This ensures consistent file detection across all stages
type FileDetector struct {
	logger *slog.Logger
}

// NewFileDetector creates a new FileDetector with optional logger
func NewFileDetector(logger *slog.Logger) *FileDetector {
	return &FileDetector{
		logger: logger,
	}
}

// DetectExcelFiles detects Excel files (.xls and .xlsx) in the specified directory
// Returns the count of Excel files found and any error encountered
// Implements defensive programming with multiple detection methods
func (fd *FileDetector) DetectExcelFiles(dir string) (int, error) {
	// Defensive: Check if detector is nil
	if fd == nil {
		return 0, fmt.Errorf("FileDetector is nil")
	}

	// Log entry per CLAUDE.md requirements
	if fd.logger != nil {
		fd.logger.Debug("Starting Excel file detection",
			slog.String("directory", dir),
			slog.String("method", "FileDetector.DetectExcelFiles"))
	}

	// Defensive: Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if fd.logger != nil {
			fd.logger.Warn("Directory does not exist for Excel detection",
				slog.String("directory", dir),
				slog.String("error", err.Error()))
		}
		return 0, fmt.Errorf("directory does not exist: %s", dir)
	}

	// Primary method: Use filepath.Glob for .xlsx files
	xlsxFiles, xlsxErr := filepath.Glob(filepath.Join(dir, "*.xlsx"))
	if xlsxErr != nil && fd.logger != nil {
		fd.logger.Error("Failed to glob .xlsx files",
			slog.String("directory", dir),
			slog.String("pattern", filepath.Join(dir, "*.xlsx")),
			slog.String("error", xlsxErr.Error()))
	}

	// Also check for .xls files (older Excel format)
	xlsFiles, xlsErr := filepath.Glob(filepath.Join(dir, "*.xls"))
	if xlsErr != nil && fd.logger != nil {
		fd.logger.Error("Failed to glob .xls files",
			slog.String("directory", dir),
			slog.String("pattern", filepath.Join(dir, "*.xls")),
			slog.String("error", xlsErr.Error()))
	}

	// Calculate total from glob results
	totalFromGlob := 0
	if xlsxErr == nil {
		totalFromGlob += len(xlsxFiles)
	}
	if xlsErr == nil {
		totalFromGlob += len(xlsFiles)
	}

	// Log glob results
	if fd.logger != nil {
		fd.logger.Info("Glob detection results",
			slog.String("directory", dir),
			slog.String("method", "filepath.Glob"),
			slog.Int("xlsx_count", len(xlsxFiles)),
			slog.Int("xls_count", len(xlsFiles)),
			slog.Int("total_count", totalFromGlob),
			slog.Bool("xlsx_error", xlsxErr != nil),
			slog.Bool("xls_error", xlsErr != nil))
	}

	// If glob succeeded and found files, return the count
	if totalFromGlob > 0 && xlsxErr == nil && xlsErr == nil {
		return totalFromGlob, nil
	}

	// Fallback method: Use os.ReadDir if glob failed or found nothing
	if fd.logger != nil {
		fd.logger.Info("Using fallback detection method",
			slog.String("directory", dir),
			slog.String("method", "os.ReadDir"),
			slog.String("reason", "glob failed or returned zero files"))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if fd.logger != nil {
			fd.logger.Error("Failed to read directory",
				slog.String("directory", dir),
				slog.String("error", err.Error()))
		}
		// If both methods failed, return the glob count (might be partial)
		if totalFromGlob > 0 {
			return totalFromGlob, nil
		}
		return 0, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	// Count Excel files from directory entries
	fallbackCount := 0
	var excelFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".xlsx") ||
			strings.HasSuffix(strings.ToLower(name), ".xls") {
			fallbackCount++
			excelFiles = append(excelFiles, name)
		}
	}

	// Log fallback results with file names for debugging
	if fd.logger != nil {
		fd.logger.Info("Fallback detection complete",
			slog.String("directory", dir),
			slog.String("method", "os.ReadDir"),
			slog.Int("excel_files_found", fallbackCount),
			slog.Int("total_entries", len(entries)),
			slog.Any("excel_files", excelFiles))
	}

	// Return the maximum count from both methods (defensive)
	if fallbackCount > totalFromGlob {
		return fallbackCount, nil
	}
	return totalFromGlob, nil
}

// DetectCSVFiles detects CSV files in the specified directory
// Follows the same defensive pattern as DetectExcelFiles
func (fd *FileDetector) DetectCSVFiles(dir string) (int, error) {
	// Defensive: Check if detector is nil
	if fd == nil {
		return 0, fmt.Errorf("FileDetector is nil")
	}

	if fd.logger != nil {
		fd.logger.Debug("Starting CSV file detection",
			slog.String("directory", dir),
			slog.String("method", "FileDetector.DetectCSVFiles"))
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if fd.logger != nil {
			fd.logger.Warn("Directory does not exist for CSV detection",
				slog.String("directory", dir))
		}
		return 0, fmt.Errorf("directory does not exist: %s", dir)
	}

	// Use glob to find CSV files
	csvFiles, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		if fd.logger != nil {
			fd.logger.Error("Failed to glob CSV files",
				slog.String("directory", dir),
				slog.String("error", err.Error()))
		}
		return 0, fmt.Errorf("failed to glob CSV files: %w", err)
	}

	if fd.logger != nil {
		fd.logger.Info("CSV detection results",
			slog.String("directory", dir),
			slog.Int("csv_count", len(csvFiles)))
	}

	return len(csvFiles), nil
}

// FileInfo provides detailed information about detected files
type FileInfo struct {
	Name      string
	Size      int64
	ModTime   string
	Extension string
}

// GetExcelFileDetails returns detailed information about Excel files
// Useful for debugging and logging purposes
func (fd *FileDetector) GetExcelFileDetails(dir string) ([]FileInfo, error) {
	if fd == nil {
		return nil, fmt.Errorf("FileDetector is nil")
	}

	var fileInfos []FileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".xlsx") ||
			strings.HasSuffix(strings.ToLower(name), ".xls") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			fileInfos = append(fileInfos, FileInfo{
				Name:      name,
				Size:      info.Size(),
				ModTime:   info.ModTime().Format(time.RFC3339),
				Extension: filepath.Ext(name),
			})
		}
	}

	return fileInfos, nil
}

// GetCSVFileDetails returns detailed information about CSV files in dir
func (fd *FileDetector) GetCSVFileDetails(dir string) ([]FileInfo, error) {
	if fd == nil {
		return nil, fmt.Errorf("FileDetector is nil")
	}

	var fileInfos []FileInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".csv") {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			fileInfos = append(fileInfos, FileInfo{
				Name:      name,
				Size:      info.Size(),
				ModTime:   info.ModTime().Format(time.RFC3339),
				Extension: filepath.Ext(name),
			})
		}
	}

	return fileInfos, nil
}
