package common

import (
	"fmt"
	"runtime"
)

const (
	// Version is the current version of the application
	Version = "0.1.0-alpha.1"
	
	// VersionMajor is the major version number
	VersionMajor = 0
	
	// VersionMinor is the minor version number
	VersionMinor = 1
	
	// VersionPatch is the patch version number
	VersionPatch = 0
	
	// VersionPrerelease is the pre-release identifier
	VersionPrerelease = "alpha.1"
	
	// VersionStage represents the current development stage
	VersionStage = "alpha"
	
	// DataFormatVersion is the version of the data format
	DataFormatVersion = "v1"
	
	// APIVersion is the version of the API (WebSocket messages)
	APIVersion = "v1-alpha"
)

var (
	// BuildTime is set during build using ldflags
	BuildTime = "unknown"
	
	// GitCommit is set during build using ldflags
	GitCommit = "unknown"
	
	// GitBranch is set during build using ldflags
	GitBranch = "unknown"
)

// VersionInfo contains detailed version information
type VersionInfo struct {
	Version      string `json:"version"`
	Stage        string `json:"stage"`
	BuildTime    string `json:"build_time"`
	GitCommit    string `json:"git_commit"`
	GitBranch    string `json:"git_branch"`
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	DataFormat   string `json:"data_format"`
	APIVersion   string `json:"api_version"`
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:      Version,
		Stage:        VersionStage,
		BuildTime:    BuildTime,
		GitCommit:    GitCommit,
		GitBranch:    GitBranch,
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		DataFormat:   DataFormatVersion,
		APIVersion:   APIVersion,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return fmt.Sprintf("ISX Daily Reports Scrapper v%s", Version)
}

// GetFullVersionString returns a detailed version string
func GetFullVersionString() string {
	info := GetVersionInfo()
	return fmt.Sprintf(
		"%s (built: %s, commit: %s, go: %s, os: %s/%s)",
		GetVersionString(),
		info.BuildTime,
		info.GitCommit,
		info.GoVersion,
		info.OS,
		info.Architecture,
	)
}

// IsAlpha returns true if this is an alpha version
func IsAlpha() bool {
	return VersionStage == "alpha"
}

// IsBeta returns true if this is a beta version
func IsBeta() bool {
	return VersionStage == "beta"
}

// IsStable returns true if this is a stable version
func IsStable() bool {
	return VersionMajor >= 1 && VersionPrerelease == ""
}

// IsPrerelease returns true if this is a pre-release version
func IsPrerelease() bool {
	return VersionPrerelease != ""
}