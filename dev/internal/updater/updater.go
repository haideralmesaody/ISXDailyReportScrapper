package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo contains update information
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	UpdateURL      string
	ReleaseNotes   string
	Size           int64
}

// Updater handles application updates
type Updater struct {
	currentVersion string
	repoURL        string
	executablePath string
}

// NewUpdater creates a new updater instance
func NewUpdater(currentVersion, repoURL string) (*Updater, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}

	return &Updater{
		currentVersion: currentVersion,
		repoURL:        repoURL,
		executablePath: execPath,
	}, nil
}

// CheckForUpdates checks if a new version is available
func (u *Updater) CheckForUpdates() (*UpdateInfo, error) {
	// Get latest release from GitHub API
	apiURL := strings.Replace(u.repoURL, "github.com", "api.github.com/repos", 1)
	apiURL = strings.TrimSuffix(apiURL, ".git") + "/releases/latest"

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %v", err)
	}

	// Check if update is needed
	if release.TagName == u.currentVersion {
		return nil, nil // No update needed
	}

	// Find appropriate asset for current platform
	assetName := u.getAssetName()
	var downloadURL string
	var size int64

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) {
			downloadURL = asset.BrowserDownloadURL
			size = asset.Size
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no suitable release asset found for %s", runtime.GOOS)
	}

	return &UpdateInfo{
		CurrentVersion: u.currentVersion,
		LatestVersion:  release.TagName,
		UpdateURL:      downloadURL,
		ReleaseNotes:   release.Name,
		Size:           size,
	}, nil
}

// PerformUpdate downloads and installs the update
func (u *Updater) PerformUpdate(updateInfo *UpdateInfo) error {
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), "isx-update")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Download update
	downloadPath := filepath.Join(tempDir, "update.zip")
	if err := u.downloadFile(updateInfo.UpdateURL, downloadPath); err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}

	// Extract update
	extractDir := filepath.Join(tempDir, "extracted")
	if err := u.extractZip(downloadPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract update: %v", err)
	}

	// Find executable in extracted files
	newExePath, err := u.findExecutable(extractDir)
	if err != nil {
		return fmt.Errorf("failed to find executable in update: %v", err)
	}

	// Backup current executable
	backupPath := u.executablePath + ".backup"
	if err := u.copyFile(u.executablePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %v", err)
	}

	// Replace executable
	if err := u.replaceExecutable(newExePath, u.executablePath); err != nil {
		// Restore backup on failure
		u.copyFile(backupPath, u.executablePath)
		return fmt.Errorf("failed to replace executable: %v", err)
	}

	// Clean up backup
	os.Remove(backupPath)

	return nil
}

// getAssetName returns the appropriate asset name for current platform
func (u *Updater) getAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
}

// downloadFile downloads a file from URL to local path
func (u *Updater) downloadFile(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractZip extracts a zip file to destination directory
func (u *Updater) extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			rc.Close()
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			rc.Close()
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// findExecutable finds the main executable in extracted directory
func (u *Updater) findExecutable(dir string) (string, error) {
	var exeName string
	if runtime.GOOS == "windows" {
		exeName = "web.exe"
	} else {
		exeName = "web"
	}

	var foundPath string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(info.Name(), exeName) {
			foundPath = path
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("executable not found in update package")
	}

	return foundPath, nil
}

// copyFile copies a file from src to dst
func (u *Updater) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// replaceExecutable replaces the current executable with new one
func (u *Updater) replaceExecutable(newPath, currentPath string) error {
	// On Windows, we might need to handle file locking differently
	if runtime.GOOS == "windows" {
		// Move current executable to temp name
		tempPath := currentPath + ".old"
		if err := os.Rename(currentPath, tempPath); err != nil {
			return err
		}

		// Copy new executable
		if err := u.copyFile(newPath, currentPath); err != nil {
			// Restore original on failure
			os.Rename(tempPath, currentPath)
			return err
		}

		// Mark old file for deletion on reboot if immediate deletion fails
		os.Remove(tempPath)
	} else {
		// On Unix systems, we can usually replace the file directly
		return u.copyFile(newPath, currentPath)
	}

	return nil
}

// AutoUpdateChecker runs periodic update checks
type AutoUpdateChecker struct {
	updater  *Updater
	interval time.Duration
	callback func(*UpdateInfo) bool // Returns true if update should be installed
}

// NewAutoUpdateChecker creates a new auto-update checker
func NewAutoUpdateChecker(updater *Updater, interval time.Duration, callback func(*UpdateInfo) bool) *AutoUpdateChecker {
	return &AutoUpdateChecker{
		updater:  updater,
		interval: interval,
		callback: callback,
	}
}

// Start begins the auto-update checking process
func (auc *AutoUpdateChecker) Start() {
	ticker := time.NewTicker(auc.interval)
	go func() {
		for range ticker.C {
			updateInfo, err := auc.updater.CheckForUpdates()
			if err != nil {
				continue // Log error in production
			}

			if updateInfo != nil && auc.callback(updateInfo) {
				if err := auc.updater.PerformUpdate(updateInfo); err != nil {
					// Log error in production
					continue
				}
				// Application should restart after update
				break
			}
		}
	}()
}

// Stop stops the auto-update checker
func (auc *AutoUpdateChecker) Stop() {
	// Implementation to stop the checker
}
