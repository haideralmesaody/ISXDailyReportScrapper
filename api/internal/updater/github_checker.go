package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Version      string    `json:"version"`
	DownloadURL  string    `json:"download_url"`
	ReleaseNotes string    `json:"release_notes"`
	PublishedAt  time.Time `json:"published_at"`
	IsAlpha      bool      `json:"is_alpha"`
	FileSize     int64     `json:"file_size"`
}

// GitHubRelease represents a GitHub release from the API
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
	Prerelease bool `json:"prerelease"`
}

// AlphaUpdateChecker checks for alpha updates from GitHub
type AlphaUpdateChecker struct {
	CurrentVersion string
	GitHubRepo     string
}

// NewAlphaChecker creates a new alpha update checker
func NewAlphaChecker(currentVersion, githubRepo string) *AlphaUpdateChecker {
	return &AlphaUpdateChecker{
		CurrentVersion: currentVersion,
		GitHubRepo:     githubRepo,
	}
}

// CheckForUpdates checks GitHub for newer alpha versions
func (uc *AlphaUpdateChecker) CheckForUpdates() (*UpdateInfo, error) {
	// GitHub API endpoint for releases
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", uc.GitHubRepo)

	// Create HTTP request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set user agent for GitHub API
	req.Header.Set("User-Agent", "ISXPulse-Updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse releases
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode releases: %w", err)
	}

	// Filter alpha pre-releases and find newer versions
	for _, release := range releases {
		// Only consider alpha pre-releases
		if !release.Prerelease {
			continue
		}

		// Check if it's an alpha version
		if !strings.Contains(release.TagName, "alpha") {
			continue
		}

		// Compare versions (newer alpha version)
		if isNewerAlphaVersion(release.TagName, uc.CurrentVersion) {
			// Find installer asset
			installerURL, fileSize := findInstallerAsset(release.Assets)
			if installerURL == "" {
				continue // No installer found, skip this release
			}

			return &UpdateInfo{
				Version:      release.TagName,
				DownloadURL:  installerURL,
				ReleaseNotes: release.Body,
				PublishedAt:  release.PublishedAt,
				IsAlpha:      true,
				FileSize:     fileSize,
			}, nil
		}
	}

	return nil, nil // No updates available
}

// isNewerAlphaVersion compares two alpha version strings
// Returns true if latest is newer than current
func isNewerAlphaVersion(latest, current string) bool {
	// Extract alpha number from version strings
	// Example: "v0.1.0-alpha.5" -> 5
	latestNum := extractAlphaNumber(latest)
	currentNum := extractAlphaNumber(current)

	return latestNum > currentNum
}

// extractAlphaNumber extracts the alpha number from a version string
func extractAlphaNumber(version string) int {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Split by "alpha."
	parts := strings.Split(version, "alpha.")
	if len(parts) != 2 {
		return 0
	}

	// Parse the number
	var num int
	fmt.Sscanf(parts[1], "%d", &num)
	return num
}

// findInstallerAsset finds the installer file in release assets
func findInstallerAsset(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}) (string, int64) {
	for _, asset := range assets {
		// Look for ISXPulseSetup-Alpha-*.exe or ISXPulseSetup-Alpha.exe
		if strings.HasPrefix(asset.Name, "ISXPulseSetup") && strings.HasSuffix(asset.Name, ".exe") {
			return asset.BrowserDownloadURL, asset.Size
		}
	}
	return "", 0
}
