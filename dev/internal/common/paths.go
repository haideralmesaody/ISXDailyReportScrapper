package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths contains all the application paths
type Paths struct {
	ExecutableDir string
	WebDir        string
	StaticDir     string
	DownloadsDir  string
	ReportsDir    string
	LogsDir       string
	LicenseFile   string
}

// GetPaths returns the application paths relative to the executable location
func GetPaths() (*Paths, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	
	// Get the directory containing the executable
	exeDir := filepath.Dir(exe)
	
	// All paths are relative to the executable directory
	// This ensures the application works correctly whether run from dev/ or release/
	paths := &Paths{
		ExecutableDir: exeDir,
		WebDir:        filepath.Join(exeDir, "web"),
		StaticDir:     filepath.Join(exeDir, "web", "static"),
		DownloadsDir:  filepath.Join(exeDir, "downloads"),
		ReportsDir:    filepath.Join(exeDir, "reports"),
		LogsDir:       filepath.Join(exeDir, "logs"),
		LicenseFile:   filepath.Join(exeDir, "license.dat"),
	}
	
	return paths, nil
}

// EnsureDirectories creates all required directories if they don't exist
func (p *Paths) EnsureDirectories() error {
	directories := []string{
		p.DownloadsDir,
		p.ReportsDir,
		p.LogsDir,
		p.WebDir,
		p.StaticDir,
	}
	
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	
	return nil
}

// GetRelativePath returns a path relative to the executable directory
func (p *Paths) GetRelativePath(subpath string) string {
	return filepath.Join(p.ExecutableDir, subpath)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetLicensePath returns the license file path, checking multiple locations
func GetLicensePath() (string, error) {
	// First, try to get the standard paths
	paths, err := GetPaths()
	if err != nil {
		return "", err
	}
	
	// Check if license.dat exists in the executable directory
	if FileExists(paths.LicenseFile) {
		return paths.LicenseFile, nil
	}
	
	// For backward compatibility, also check current working directory
	cwd, err := os.Getwd()
	if err == nil {
		cwdLicense := filepath.Join(cwd, "license.dat")
		if FileExists(cwdLicense) {
			return cwdLicense, nil
		}
	}
	
	// Return the default path (even if file doesn't exist yet)
	return paths.LicenseFile, nil
}

// GetWebFilePath returns the path to a web file
func (p *Paths) GetWebFilePath(filename string) string {
	return filepath.Join(p.WebDir, filename)
}

// GetStaticFilePath returns the path to a static file
func (p *Paths) GetStaticFilePath(filename string) string {
	return filepath.Join(p.StaticDir, filename)
}

// GetDownloadPath returns the path for a downloaded file
func (p *Paths) GetDownloadPath(filename string) string {
	return filepath.Join(p.DownloadsDir, filename)
}

// GetReportPath returns the path for a report file
func (p *Paths) GetReportPath(filename string) string {
	return filepath.Join(p.ReportsDir, filename)
}

// GetLogPath returns the path for a log file
func (p *Paths) GetLogPath(filename string) string {
	return filepath.Join(p.LogsDir, filename)
}