package license

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// determinePathSource determines the source of the license path for debugging.
func determinePathSource(licenseFile string) string {
	if strings.Contains(licenseFile, "license.dat") {
		return "config_system"
	}
	return "custom_path"
}

// checkFileReadable checks if a file is readable.
func checkFileReadable(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// getFileSizeString returns a human-readable file size string.
func getFileSizeString(filePath string) string {
	info, err := os.Stat(filePath)
	if err != nil {
		return "unknown"
	}

	if info.Size() == 0 {
		return "0 bytes"
	}

	const KB = 1024
	const MB = KB * 1024
	const GB = MB * 1024

	size := info.Size()
	switch {
	case size < KB:
		return fmt.Sprintf("%d bytes", size)
	case size < MB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	case size < GB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	default:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	}
}

// validateLicensePath performs input validation for license file paths.
func validateLicensePath(licenseFile string) error {
	if licenseFile == "" {
		return fmt.Errorf("license file path cannot be empty")
	}

	invalidChars := []string{"<", ">", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(licenseFile, char) {
			return fmt.Errorf("license file path contains invalid character: %s", char)
		}
	}

	if len(licenseFile) > 260 {
		return fmt.Errorf("license file path too long (max 260 characters)")
	}

	return nil
}

// getWorkingDir returns the current working directory for logging.
func (m *Manager) getWorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "unknown"
}

// isWritable checks if a directory is writable.
func isWritable(path string) bool {
	testFile := filepath.Join(path, ".write_test_"+fmt.Sprintf("%d", time.Now().UnixNano()))
	if f, err := os.Create(testFile); err == nil {
		f.Close()
		os.Remove(testFile)
		return true
	}
	return false
}
