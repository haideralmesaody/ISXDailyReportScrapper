package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestLatestDownloadedDate verifies that the most recent date is correctly detected.
func TestLatestDownloadedDate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample files.
	names := []string{
		"2024 12 30 ISX Daily Report.xlsx",
		"2025 01 01 ISX Daily Report.xlsx",
		"2023 11 15 ISX Daily Report.xlsx",
	}
	for _, n := range names {
		path := filepath.Join(tmpDir, n)
		if err := os.WriteFile(path, []byte("dummy"), 0o644); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
	}

	pattern := regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	t.Logf("Found %d entries in %s", len(entries), tmpDir)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		t.Logf("File: %s", e.Name())
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			t.Logf("No match for %s", e.Name())
			continue
		}
		t.Logf("Match: %v", m)
		dateStr := strings.Join(m[1:4], " ")
		t.Logf("Date str: %s", dateStr)
		parsedDate, parseErr := time.Parse("2006 01 02", dateStr)
		if parseErr != nil {
			t.Logf("Parse err: %v", parseErr)
			continue
		}
		t.Logf("Parsed date: %s", parsedDate.Format("2006-01-02"))
	}

	d, ok := latestDownloadedDate(tmpDir)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	want := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !d.Equal(want) {
		t.Fatalf("wrong date: want %s, got %s", want.Format("2006-01-02"), d.Format("2006-01-02"))
	}
}
