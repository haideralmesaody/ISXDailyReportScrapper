package services

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CombinedCSVSnapshot struct {
	ModTime time.Time
	Size    int64
}

type combinedDateRangeCache struct {
	mu      sync.Mutex
	start   time.Time
	end     time.Time
	snap    CombinedCSVSnapshot
	hasData bool
}

func (ds *DataService) GetCombinedDateRange(ctx context.Context) (time.Time, time.Time, CombinedCSVSnapshot, error) {
	if ds == nil || ds.paths == nil {
		return time.Time{}, time.Time{}, CombinedCSVSnapshot{}, fmt.Errorf("data service paths not configured")
	}

	combinedCSV := filepath.Join(ds.paths.ReportsDir, "combined", "isx_combined_data.csv")

	info, err := os.Stat(combinedCSV)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, time.Time{}, CombinedCSVSnapshot{}, fmt.Errorf("combined data CSV not found: %s", combinedCSV)
		}
		return time.Time{}, time.Time{}, CombinedCSVSnapshot{}, fmt.Errorf("stat combined data CSV: %w", err)
	}

	snap := CombinedCSVSnapshot{
		ModTime: info.ModTime().UTC(),
		Size:    info.Size(),
	}

	ds.combinedDateRange.mu.Lock()
	if ds.combinedDateRange.hasData &&
		ds.combinedDateRange.snap.ModTime.Equal(snap.ModTime) &&
		ds.combinedDateRange.snap.Size == snap.Size {
		start := ds.combinedDateRange.start
		end := ds.combinedDateRange.end
		ds.combinedDateRange.mu.Unlock()
		return start, end, snap, nil
	}
	ds.combinedDateRange.mu.Unlock()

	start, end, err := scanCombinedDateRange(ctx, combinedCSV)
	if err != nil {
		return time.Time{}, time.Time{}, CombinedCSVSnapshot{}, err
	}

	ds.combinedDateRange.mu.Lock()
	ds.combinedDateRange.start = start
	ds.combinedDateRange.end = end
	ds.combinedDateRange.snap = snap
	ds.combinedDateRange.hasData = true
	ds.combinedDateRange.mu.Unlock()

	return start, end, snap, nil
}

func scanCombinedDateRange(ctx context.Context, path string) (time.Time, time.Time, error) {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("open combined data CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("read combined CSV header: %w", err)
	}

	dateIdx := -1
	for i, col := range header {
		colName := strings.TrimSpace(strings.TrimPrefix(col, "\uFEFF"))
		if strings.EqualFold(colName, "Date") {
			dateIdx = i
			break
		}
	}
	if dateIdx == -1 {
		// Fall back to legacy: Date is the first column.
		dateIdx = 0
	}

	var min time.Time
	var max time.Time
	seen := 0

	for {
		select {
		case <-ctx.Done():
			return time.Time{}, time.Time{}, ctx.Err()
		default:
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if dateIdx >= len(record) {
			continue
		}

		raw := strings.TrimSpace(record[dateIdx])
		if raw == "" {
			continue
		}
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			continue
		}
		parsed = parsed.UTC()
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)

		if seen == 0 || parsed.Before(min) {
			min = parsed
		}
		if seen == 0 || parsed.After(max) {
			max = parsed
		}
		seen++
	}

	if seen == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("combined CSV has no parsable dates: %s", path)
	}

	return min, max, nil
}
