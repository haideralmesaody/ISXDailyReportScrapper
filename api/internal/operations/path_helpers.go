package operations

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvironmentType represents the type of environment the application is running in
type EnvironmentType int

const (
	EnvironmentProduction EnvironmentType = iota
	EnvironmentDevelopment
	EnvironmentTest
)

// getEnvironmentType detects the current environment
func getEnvironmentType() EnvironmentType {
	// Check if we're in a temporary Go build directory
	if exe, err := os.Executable(); err == nil {
		exePath := strings.ToLower(exe)
		if strings.Contains(exePath, "go-build") || strings.Contains(exePath, "temp") {
			return EnvironmentDevelopment
		}

		// Check if running from test binary
		if strings.Contains(exePath, ".test") || strings.HasSuffix(exePath, ".exe.test") {
			return EnvironmentTest
		}
	}

	// Check development indicators
	if isDevelopmentEnvironment() {
		return EnvironmentDevelopment
	}

	return EnvironmentProduction
}

// isDevelopmentEnvironment checks if we're running in a development environment
func isDevelopmentEnvironment() bool {
	// Check for typical development indicators
	if _, err := os.Stat("go.mod"); err == nil {
		return true
	}
	if _, err := os.Stat("api"); err == nil {
		return true
	}
	if _, err := os.Stat("web"); err == nil {
		return true
	}
	return false
}

// getCorrectBaseDirectory returns the correct base directory for the current environment
func getCorrectBaseDirectory(providedExecDir string) string {
	env := getEnvironmentType()

	switch env {
	case EnvironmentDevelopment:
		// In development, prefer a project dist directory that actually contains the data tree.
		// Repos can include other "dist" folders (e.g. under api/) that only contain frontend assets.
		if hasDataDir("dist") {
			return "dist"
		}
		// If no dist directory, try to find it
		if distDir := findDistDirectory(); distDir != "" {
			return distDir
		}
		// Fallback to current directory
		return "."

	case EnvironmentTest:
		// In tests, use the provided executable directory or create a temp directory
		if providedExecDir != "" && !strings.Contains(providedExecDir, "go-build") {
			return providedExecDir
		}
		// Let tests create their own temp directories
		return providedExecDir

	case EnvironmentProduction:
		fallthrough
	default:
		// In production, use the actual executable directory
		return providedExecDir
	}
}

// findDistDirectory searches for the dist directory in common locations
func findDistDirectory() string {
	// Common locations to check
	locations := []string{
		"dist",
		"../dist",
		"../../dist",
		"../../../dist",
	}

	// Also check if we're in a subdirectory of the project
	if wd, err := os.Getwd(); err == nil {
		// Navigate up looking for dist directory
		currentDir := wd
		for i := 0; i < 5; i++ { // Look up at most 5 levels
			testPath := filepath.Join(currentDir, "dist")
			if hasDataDir(testPath) {
				return testPath
			}

			parent := filepath.Dir(currentDir)
			if parent == currentDir { // Reached root
				break
			}
			currentDir = parent
		}
	}

	// Check the explicit locations
	for _, location := range locations {
		if absPath, err := filepath.Abs(location); err == nil {
			if hasDataDir(absPath) {
				return absPath
			}
		}
	}

	return ""
}

func hasDataDir(baseDir string) bool {
	info, err := os.Stat(filepath.Join(baseDir, "data"))
	return err == nil && info.IsDir()
}

// GetAbsoluteDataPath returns the absolute path to a data-relative directory
func GetAbsoluteDataPath(relativePath string, providedExecDir string) string {
	baseDir := getCorrectBaseDirectory(providedExecDir)
	return filepath.Join(baseDir, relativePath)
}

// IsRunningFromDist checks if the application is running from the dist directory
func IsRunningFromDist() bool {
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		return strings.HasSuffix(strings.ToLower(exeDir), "dist")
	}
	return false
}

// GetProjectRoot attempts to find the project root directory
func GetProjectRoot() string {
	// Start from current directory and work upwards
	if wd, err := os.Getwd(); err == nil {
		currentDir := wd
		for i := 0; i < 10; i++ { // Look up at most 10 levels
			// Check for project indicators
			if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
				return currentDir
			}
			if _, err := os.Stat(filepath.Join(currentDir, "api")); err == nil {
				if _, err := os.Stat(filepath.Join(currentDir, "web")); err == nil {
					return currentDir
				}
			}

			parent := filepath.Dir(currentDir)
			if parent == currentDir { // Reached root
				break
			}
			currentDir = parent
		}
	}
	return ""
}

func relativeDownloadsRoot() string {
	return filepath.Join("data", "downloads")
}

func relativeReportsRoot() string {
	return filepath.Join("data", "reports")
}

func relativeDownloadFilePath(fileName string) string {
	return filepath.Join(relativeDownloadsRoot(), fileName)
}

// AbsolutePath resolves a relative path to an absolute path using the correct executable directory
// This function should be used by all stages instead of filepath.Join(stage.executableDir, ...)
func AbsolutePath(executableDir, relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}

	// Use the corrected base directory
	correctExecDir := getCorrectBaseDirectory(executableDir)
	joined := filepath.Join(correctExecDir, relPath)
	abs, err := filepath.Abs(joined)
	if err != nil {
		return joined
	}
	return abs
}

// AbsoluteDataPath resolves a path relative to the data directory
func AbsoluteDataPath(executableDir, relPath string) string {
	return AbsolutePath(executableDir, filepath.Join("data", relPath))
}

// AbsoluteDownloadsPath resolves a path relative to the downloads directory
func AbsoluteDownloadsPath(executableDir, relPath string) string {
	return AbsolutePath(executableDir, relativeDownloadsRoot())
}

// AbsoluteReportsPath resolves a path relative to the reports directory
func AbsoluteReportsPath(executableDir, relPath string) string {
	return AbsolutePath(executableDir, relativeReportsRoot())
}

// GetEnvironmentInfo returns information about the current environment for debugging
func GetEnvironmentInfo() map[string]interface{} {
	exe, _ := os.Executable()
	wd, _ := os.Getwd()
	env := getEnvironmentType()

	info := map[string]interface{}{
		"executable_path":      exe,
		"working_dir":          wd,
		"environment":          env,
		"os":                   runtime.GOOS,
		"arch":                 runtime.GOARCH,
		"is_running_from_dist": IsRunningFromDist(),
		"project_root":         GetProjectRoot(),
	}

	// Add environment-specific details
	switch env {
	case EnvironmentDevelopment:
		info["detected_reason"] = "Go build directory or development indicators detected"
		if distDir := findDistDirectory(); distDir != "" {
			info["found_dist_dir"] = distDir
		}
	case EnvironmentProduction:
		info["detected_reason"] = "Running from production binary location"
	case EnvironmentTest:
		info["detected_reason"] = "Test binary detected"
	}

	return info
}
