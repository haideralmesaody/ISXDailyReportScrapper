package main

import (
	"flag"
	"fmt"
	"io"
	"github.com/isxcli/isxcli/internal/updater"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	GitHubRepo = "haideralmesaody/ISXDailyReportScrapper"
)

func main() {
	// Command line flags
	checkOnly := flag.Bool("check", false, "Check for updates and display information")
	checkSilent := flag.Bool("check-silent", false, "Check for updates silently (no output if no update)")
	flag.Parse()

	// Read current version
	currentVersion, err := readVersion()
	if err != nil {
		log.Fatalf("Failed to read version: %v", err)
	}

	// Create update checker
	checker := updater.NewAlphaChecker(currentVersion, GitHubRepo)

	// Check for updates
	update, err := checker.CheckForUpdates()
	if err != nil {
		if !*checkSilent {
			log.Fatalf("Failed to check for updates: %v", err)
		}
		os.Exit(1)
	}

	// No update available
	if update == nil {
		if *checkOnly {
			fmt.Printf("No updates available. Current version: %s\n", currentVersion)
		}
		os.Exit(0)
	}

	// Update available
	if *checkOnly || *checkSilent {
		if *checkOnly {
			fmt.Printf("\n🚀 New Alpha Version Available!\n")
			fmt.Printf("Current: %s\n", currentVersion)
			fmt.Printf("Latest:  %s\n", update.Version)
			fmt.Printf("Size:    %.2f MB\n", float64(update.FileSize)/1024/1024)
			fmt.Printf("\nDownload: %s\n", update.DownloadURL)
			if update.ReleaseNotes != "" {
				fmt.Printf("\nRelease Notes:\n%s\n", update.ReleaseNotes)
			}
		}
		os.Exit(0)
	}

	// Download and install update
	fmt.Printf("Downloading ISX Pulse %s...\n", update.Version)

	installerPath, err := downloadWithProgress(update.DownloadURL, update.FileSize)
	if err != nil {
		log.Fatalf("Failed to download update: %v", err)
	}

	fmt.Println("\nDownload complete!")
	fmt.Println("Preparing to install update...")

	// Close main application
	if err := killProcess("ISXPulse.exe"); err != nil {
		fmt.Printf("Warning: Failed to close ISXPulse: %v\n", err)
	}

	// Run installer in silent update mode
	fmt.Println("Installing update...")
	cmd := exec.Command(installerPath, "/SILENT", "/UPDATE", "/PRESERVEDATA")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start installer: %v", err)
	}

	fmt.Println("Update installer started successfully!")
	os.Exit(0)
}

// readVersion reads the current version from version.txt
func readVersion() (string, error) {
	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}

	exeDir := filepath.Dir(exePath)
	versionFile := filepath.Join(exeDir, "version.txt")

	// Read version file
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", fmt.Errorf("read version file: %w", err)
	}

	version := strings.TrimSpace(string(data))
	if version == "" {
		return "", fmt.Errorf("version file is empty")
	}

	return version, nil
}

// downloadWithProgress downloads a file and shows progress
func downloadWithProgress(url string, totalSize int64) (string, error) {
	// Create temp file
	tempFile, err := os.CreateTemp("", "ISXPulseSetup-*.exe")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tempFile.Close()

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	// Copy with progress
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := tempFile.Write(buf[:n]); err != nil {
				return "", fmt.Errorf("write file: %w", err)
			}
			downloaded += int64(n)

			// Show progress
			if totalSize > 0 {
				percent := float64(downloaded) / float64(totalSize) * 100
				fmt.Printf("\rProgress: %.1f%% (%.2f MB / %.2f MB)",
					percent,
					float64(downloaded)/1024/1024,
					float64(totalSize)/1024/1024)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read response: %w", err)
		}
	}

	fmt.Println() // New line after progress
	return tempFile.Name(), nil
}

// killProcess attempts to kill a process by name
func killProcess(name string) error {
	// Use taskkill on Windows
	cmd := exec.Command("taskkill", "/F", "/IM", name)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	// Ignore errors - process might not be running
	_ = cmd.Run()
	return nil
}
