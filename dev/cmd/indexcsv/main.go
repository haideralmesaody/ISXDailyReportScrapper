package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// regex for filenames like "2025 06 24 ISX Daily Report.xlsx"
var fileRe = regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)

func main() {
	mode := flag.String("mode", "initial", "initial | accumulative")
	dir := flag.String("dir", "data/downloads", "directory containing xlsx reports")
	out := flag.String("out", "data/reports/indexes.csv", "output csv file path")
	flag.Parse()

	fmt.Printf("Starting index extraction in %s mode...\n", *mode)

	var lastDate time.Time
	if *mode == "accumulative" {
		if d, err := loadLastDate(*out); err == nil {
			lastDate = d
			fmt.Printf("[accumulative] Existing CSV last date: %s\n", lastDate.Format("2006-01-02"))
		} else {
			fmt.Printf("[accumulative] No existing CSV found, switching to initial mode\n")
			*mode = "initial"
		}
	}

	if *mode == "initial" {
		// initial mode: create/truncate csv with header
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot create %s: %v\n", *out, err)
			os.Exit(1)
		}
		w := csv.NewWriter(f)
		w.Write([]string{"Date", "ISX60", "ISX15"})
		w.Flush()
		_ = f.Close()
		fmt.Printf("[initial] Created new CSV file: %s\n", *out)
	}

	entries, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dir failed: %v\n", err)
		os.Exit(1)
	}

	type fileInfo struct {
		path string
		date time.Time
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := fileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		t, _ := time.Parse("2006 01 02", strings.Join(m[1:4], " "))
		if !lastDate.IsZero() && !t.After(lastDate) {
			continue // already processed
		}
		files = append(files, fileInfo{path: filepath.Join(*dir, e.Name()), date: t})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].date.Before(files[j].date) })

	fmt.Printf("Found %d Excel files to process\n", len(files))
	if len(files) == 0 {
		fmt.Println("No new files to process.")
		return
	}

	outF, err := os.OpenFile(*out, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open output failed: %v\n", err)
		os.Exit(1)
	}
	defer outF.Close()
	writer := csv.NewWriter(outF)

	processedCount := 0
	for i, fi := range files {
		fmt.Printf("Processing file %d/%d: %s\n", i+1, len(files), filepath.Base(fi.path))

		isx60, isx15, err := extractIndices(fi.path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filepath.Base(fi.path), err)
			continue
		}

		rec := []string{fi.date.Format("2006-01-02"), formatFloat(isx60)}
		if isx15 > 0 {
			rec = append(rec, formatFloat(isx15))
		} else {
			rec = append(rec, "")
		}
		writer.Write(rec)
		processedCount++

		if isx15 > 0 {
			fmt.Printf("✓ Added %s (ISX60=%.2f, ISX15=%.2f)\n", fi.date.Format("2006-01-02"), isx60, isx15)
		} else {
			fmt.Printf("✓ Added %s (ISX60=%.2f, ISX15=N/A)\n", fi.date.Format("2006-01-02"), isx60)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "write csv error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Index extraction completed successfully!\n")
	fmt.Printf("Processed %d files\n", processedCount)
	fmt.Printf("Output written to: %s\n", *out)
}

func loadLastDate(csvPath string) (time.Time, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	var last string
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return time.Time{}, err
		}
		if rec[0] == "Date" {
			continue
		}
		last = rec[0]
	}
	if last == "" {
		return time.Time{}, fmt.Errorf("no data rows")
	}
	t, err := time.Parse("2006-01-02", last)
	return t, err
}

func extractIndices(path string) (isx60, isx15 float64, err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	// Build list of sheets to inspect: prefer "Indices" if exists, otherwise all
	var sheets []string
	hasIndices := false
	for _, sh := range f.GetSheetList() {
		if strings.EqualFold(sh, "indices") {
			hasIndices = true
			break
		}
	}
	if hasIndices {
		sheets = []string{"Indices"}
	} else {
		sheets = f.GetSheetList()
	}

	joinRe := regexp.MustCompile(`\s+`)
	for _, sheet := range sheets {
		rows, _ := f.GetRows(sheet)
		for _, row := range rows {
			line := strings.TrimSpace(joinRe.ReplaceAllString(strings.Join(row, " "), " "))
			if line == "" {
				continue
			}
			// Case 1: Both 60 and 15 on the same line
			if strings.Contains(line, "ISX Index 60") && strings.Contains(line, "ISX Index 15") {
				numRe := regexp.MustCompile(`ISX Index 60\s+([0-9.,]+).*?ISX Index 15\s+([0-9.,]+)`) // non-greedy
				if m := numRe.FindStringSubmatch(line); m != nil {
					isx60, _ = parseFloat(m[1])
					isx15, _ = parseFloat(m[2])
					return isx60, isx15, nil
				}
			}

			// Case 2: Only 60 present (older reports)
			if strings.Contains(line, "ISX Index 60") {
				numRe := regexp.MustCompile(`ISX Index 60\s+([0-9.,]+)`)
				if m := numRe.FindStringSubmatch(line); m != nil {
					isx60, _ = parseFloat(m[1])
					return isx60, 0, nil
				}
			}

			// Case 3: Very old format – "ISX Price Index"
			if strings.Contains(line, "ISX Price Index") {
				numRe := regexp.MustCompile(`ISX Price Index\s+([0-9.,]+)`)
				if m := numRe.FindStringSubmatch(line); m != nil {
					isx60, _ = parseFloat(m[1]) // treat as 60 index
					return isx60, 0, nil
				}
			}
		}
	}
	return 0, 0, fmt.Errorf("indices not found in %s", filepath.Base(path))
}

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
