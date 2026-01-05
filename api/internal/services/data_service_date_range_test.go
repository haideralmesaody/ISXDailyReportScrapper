package services

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/isxcli/isxcli/internal/config"
)

func TestDataService_GetCombinedDateRange(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	combinedDir := filepath.Join(tmp, "combined")
	if err := os.MkdirAll(combinedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	path := filepath.Join(combinedDir, "isx_combined_data.csv")
	csv := "Date,Symbol,ClosePrice\n" +
		"2025-01-05,AAA,1\n" +
		"2024-12-31,BBB,2\n" +
		"2025-02-01,CCC,3\n"
	if err := os.WriteFile(path, []byte(csv), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	ds := &DataService{
		paths:  &config.Paths{ReportsDir: tmp},
		logger: slog.Default(),
	}

	start, end, snap, err := ds.GetCombinedDateRange(context.Background())
	if err != nil {
		t.Fatalf("GetCombinedDateRange: %v", err)
	}
	if snap.Size == 0 || snap.ModTime.IsZero() {
		t.Fatalf("expected non-empty snapshot: %+v", snap)
	}

	wantStart := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) {
		t.Fatalf("start mismatch: got %v want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Fatalf("end mismatch: got %v want %v", end, wantEnd)
	}
}
