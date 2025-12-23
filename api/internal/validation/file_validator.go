package validation

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// FileValidator provides common file validation functions for all executables
type FileValidator struct {
	logger *slog.Logger
}

// NewFileValidator creates a new file validator
func NewFileValidator(logger *slog.Logger) *FileValidator {
	if logger == nil {
		logger = slog.Default()
	}
	return &FileValidator{
		logger: logger,
	}
}

// ValidateInputDirectory validates that input directory exists and contains expected files
func (v *FileValidator) ValidateInputDirectory(dir string, requiredPattern string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		v.logger.Error("Input directory does not exist",
			slog.String("directory", dir))
		return fmt.Errorf("input directory %s does not exist", dir)
	}
	if err != nil {
		v.logger.Error("Failed to stat input directory",
			slog.String("directory", dir),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		v.logger.Error("Input path is not a directory",
			slog.String("path", dir))
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Check for files matching pattern if provided
	if requiredPattern != "" {
		pattern := filepath.Join(dir, requiredPattern)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			v.logger.Error("Failed to check for files",
				slog.String("pattern", pattern),
				slog.String("error", err.Error()))
			return fmt.Errorf("failed to check for files: %w", err)
		}
		
		if len(matches) == 0 {
			v.logger.Warn("No files matching pattern found",
				slog.String("directory", dir),
				slog.String("pattern", requiredPattern))
			// This is not an error - just no files to process
			return nil
		}
		
		v.logger.Info("Input directory validated",
			slog.String("directory", dir),
			slog.Int("files_found", len(matches)),
			slog.String("pattern", requiredPattern))
	}

	return nil
}

// ValidateOutputDirectory ensures output directory exists or can be created
func (v *FileValidator) ValidateOutputDirectory(dir string) error {
	// Try to create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		v.logger.Error("Failed to create output directory",
			slog.String("directory", dir),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to create output directory %s: %w", dir, err)
	}

	// Verify it's writable by creating a test file
	testFile := filepath.Join(dir, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		v.logger.Error("Output directory is not writable",
			slog.String("directory", dir),
			slog.String("error", err.Error()))
		return fmt.Errorf("output directory %s is not writable: %w", dir, err)
	}
	file.Close()
	os.Remove(testFile)

	v.logger.Info("Output directory validated",
		slog.String("directory", dir))
	return nil
}

// ValidateFile checks if a specific file exists and is readable
func (v *FileValidator) ValidateFile(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		v.logger.Error("File does not exist",
			slog.String("file", path))
		return fmt.Errorf("file %s does not exist", path)
	}
	if err != nil {
		v.logger.Error("Failed to stat file",
			slog.String("file", path),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to stat file %s: %w", path, err)
	}
	if info.IsDir() {
		v.logger.Error("Path is a directory, not a file",
			slog.String("path", path))
		return fmt.Errorf("%s is a directory, not a file", path)
	}

	// Check if file is readable by opening it
	file, err := os.Open(path)
	if err != nil {
		v.logger.Error("File is not readable",
			slog.String("file", path),
			slog.String("error", err.Error()))
		return fmt.Errorf("file %s is not readable: %w", path, err)
	}
	file.Close()

	v.logger.Debug("File validated",
		slog.String("file", path),
		slog.Int64("size", info.Size()))
	return nil
}

// CountFiles counts files matching a pattern in a directory
func (v *FileValidator) CountFiles(dir string, pattern string) (int, error) {
	fullPattern := filepath.Join(dir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		v.logger.Error("Failed to count files",
			slog.String("pattern", fullPattern),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Filter out directories from matches
	fileCount := 0
	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && !info.IsDir() {
			fileCount++
		}
	}

	v.logger.Debug("Files counted",
		slog.String("directory", dir),
		slog.String("pattern", pattern),
		slog.Int("count", fileCount))
	return fileCount, nil
}

// ValidateExcelFile checks if a file is a valid Excel file
func (v *FileValidator) ValidateExcelFile(path string) error {
	// First validate it exists
	if err := v.ValidateFile(path); err != nil {
		return err
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".xlsx" && ext != ".xls" {
		v.logger.Error("File is not an Excel file",
			slog.String("file", path),
			slog.String("extension", ext))
		return fmt.Errorf("file %s is not an Excel file (extension: %s)", path, ext)
	}

	// Check it's not a temp file
	base := filepath.Base(path)
	if strings.HasPrefix(base, "~$") {
		v.logger.Warn("Skipping temporary Excel file",
			slog.String("file", path))
		return fmt.Errorf("file %s is a temporary Excel file", path)
	}

	return nil
}

// ValidateCSVFile checks if a file is a valid CSV file
func (v *FileValidator) ValidateCSVFile(path string) error {
	// First validate it exists
	if err := v.ValidateFile(path); err != nil {
		return err
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".csv" {
		v.logger.Error("File is not a CSV file",
			slog.String("file", path),
			slog.String("extension", ext))
		return fmt.Errorf("file %s is not a CSV file (extension: %s)", path, ext)
	}

	return nil
}

// CreateEmptyCSV creates an empty CSV file with headers
func (v *FileValidator) CreateEmptyCSV(path string, headers []string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := v.ValidateOutputDirectory(dir); err != nil {
		return err
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		v.logger.Error("Failed to create empty CSV",
			slog.String("file", path),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to create CSV %s: %w", path, err)
	}
	defer file.Close()

	// Write headers if provided
	if len(headers) > 0 {
		headerLine := strings.Join(headers, ",") + "\n"
		if _, err := file.WriteString(headerLine); err != nil {
			v.logger.Error("Failed to write CSV headers",
				slog.String("file", path),
				slog.String("error", err.Error()))
			return fmt.Errorf("failed to write headers: %w", err)
		}
	}

	v.logger.Info("Created empty CSV file",
		slog.String("file", path),
		slog.Int("headers", len(headers)))
	return nil
}