package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/config"
	"github.com/isxcli/isxcli/internal/infrastructure"

	"github.com/xuri/excelize/v2"
)

// regex for filenames like "2025 06 24 ISX Daily Report.xlsx"
var fileRe = regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)

func main() {
	mode := flag.String("mode", "initial", "initial | accumulative")
	dir := flag.String("dir", "", "directory containing xlsx reports (defaults to data/downloads relative to executable)")
	out := flag.String("out", "", "output csv file path (defaults to data/reports/indexes.csv)")
	flag.Parse()

	// Initialize paths first to get default directories
	paths, err := config.GetPaths()
	if err != nil {
		slog.Error("Failed to initialize paths", "error", err)
		os.Exit(1)
	}

	// Use centralized directories as defaults if not specified
	if *dir == "" {
		*dir = paths.DownloadsDir
	}
	if *out == "" {
		*out = paths.IndexCSV
	}
	
	// Ensure all required directories exist
	if err := paths.EnsureDirectories(); err != nil {
		slog.Error("Failed to create required directories", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger per CLAUDE.md
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("Failed to load config, using defaults", "error", err)
		cfg = &config.Config{
			Logging: config.LoggingConfig{
				Level:       "info",
				Format:      "json",
				Output:      "both",
				FilePath:    paths.GetLogPath("indexcsv.log"),
				Development: false,
			},
		}
	}

	logger, err := infrastructure.InitializeLogger(cfg.Logging)
	if err != nil {
		slog.Warn("Failed to initialize logger, using default", "error", err)
		logger = slog.Default()
	}

	logger.Info("Starting index extraction",
		slog.String("mode", *mode),
		slog.String("input_dir", *dir),
		slog.String("output_file", *out),
		slog.String("executable_dir", paths.ExecutableDir))

	// Ensure output directory exists for both initial and accumulative modes
	// Each process creates its own directories as needed
	outDir := filepath.Dir(*out)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		logger.Error("Cannot create output directory", 
			slog.String("path", outDir),
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Ensured output directory exists", slog.String("path", outDir))

	var lastDate time.Time
	if *mode == "accumulative" {
		if d, err := loadLastDate(*out); err == nil {
			lastDate = d
			logger.Info("Existing CSV last date", slog.String("last_date", lastDate.Format("2006-01-02")))
		} else {
					logger.Warn("No existing CSV found, switching to initial mode", slog.String("error", err.Error()))
			*mode = "initial"
		}
	}

	if *mode == "initial" {
		// initial mode: create/truncate csv with header
		f, err := os.Create(*out)
		if err != nil {
				logger.Error("Cannot create output file", 
				slog.String("path", *out),
				slog.String("error", err.Error()))
			slog.Error("Cannot create output file", "path", *out, "error", err)
			os.Exit(1)
		}
		w := csv.NewWriter(f)
		w.Write([]string{"Date", "ISX60", "ISX15"})
		w.Flush()
		_ = f.Close()
		logger.Info("Created new CSV file", slog.String("path", *out))
	}

	entries, err := os.ReadDir(*dir)
	if err != nil {
			logger.Error("Failed to read directory",
			slog.String("dir", *dir),
			slog.String("error", err.Error()))
		slog.Error("Failed to read directory", "dir", *dir, "error", err)
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
			logger.Debug("Skipping already processed file",
				slog.String("filename", e.Name()),
				slog.String("file_date", t.Format("2006-01-02")),
				slog.String("last_processed_date", lastDate.Format("2006-01-02")))
			continue // already processed
		}
		files = append(files, fileInfo{path: filepath.Join(*dir, e.Name()), date: t})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].date.Before(files[j].date) })

	logger.Info("Excel files found", slog.Int("count", len(files)))
	
	// Output progress message for stages.go to parse
	fmt.Printf("Found %d Excel files\n", len(files))
	if len(files) == 0 {
		logger.Info("No new files to process")
		
		// Create empty indices CSV if it doesn't exist (for consistency)
		if *mode == "initial" {
			// CSV was already created with headers above
			logger.Info("Created empty indices CSV with headers", slog.String("path", *out))
		}
		
		// Signal completion to stages.go
		fmt.Println("Index extraction complete: 0 files")
		return
	}
	
	// Output file list for stages.go to parse (for segmented progress)
	if len(files) > 0 {
		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, filepath.Base(f.path))
		}
		fmt.Printf("Files to process: %s\n", strings.Join(fileNames, "|"))
	}

	outF, err := os.OpenFile(*out, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
			logger.Error("Failed to open output file",
			slog.String("path", *out),
			slog.String("error", err.Error()))
		slog.Error("Failed to open output file", "path", *out, "error", err)
		os.Exit(1)
	}
	defer outF.Close()
	writer := csv.NewWriter(outF)

	processedCount := 0
	for i, fi := range files {
		logger.Info("Processing file",
			slog.Int("current", i+1),
			slog.Int("total", len(files)),
			slog.String("filename", filepath.Base(fi.path)))
		
		// Output progress message for stages.go to parse
		fmt.Printf("Processing file %d of %d: %s\n", i+1, len(files), filepath.Base(fi.path))

		isx60, isx15, err := extractIndices(fi.path)
		if err != nil {
				logger.Error("Error processing file",
				slog.String("filename", filepath.Base(fi.path)),
				slog.String("error", err.Error()))
			slog.Warn("Error processing file", "filename", filepath.Base(fi.path), "error", err)
			continue
		}

		rec := []string{fi.date.Format("2006-01-02"), formatFloat(isx60)}
		if isx15 > 0 {
			rec = append(rec, formatFloat(isx15))
		} else {
			rec = append(rec, "")
		}
		
		// Write and immediately check for errors
		if err := writer.Write(rec); err != nil {
			logger.Error("Failed to write CSV record",
				slog.String("date", fi.date.Format("2006-01-02")),
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		
		// Flush after each write to catch disk errors immediately
		writer.Flush()
		if err := writer.Error(); err != nil {
			logger.Error("CSV flush error",
				slog.String("date", fi.date.Format("2006-01-02")),
				slog.String("error", err.Error()))
			os.Exit(1)
		}
		
		processedCount++

		// Output extraction success for tracking
		if isx15 > 0 {
			fmt.Printf("Extracted indices from %s: ISX60=%.2f, ISX15=%.2f\n",
				filepath.Base(fi.path), isx60, isx15)
			logger.Info("Added index data",
				slog.String("date", fi.date.Format("2006-01-02")),
				slog.Float64("ISX60", isx60),
				slog.Float64("ISX15", isx15))
		} else {
			fmt.Printf("Extracted index from %s: ISX60=%.2f (ISX15 N/A)\n",
				filepath.Base(fi.path), isx60)
			logger.Info("Added index data (ISX15 N/A)",
				slog.String("date", fi.date.Format("2006-01-02")),
				slog.Float64("ISX60", isx60))
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
			logger.Error("CSV write error", slog.String("error", err.Error()))
		slog.Error("Failed to write CSV", "error", err)
		os.Exit(1)
	}

	slog.Info("Index extraction completed successfully!")
	logger.Info("Index extraction completed",
		slog.Int("processed_files", processedCount),
		slog.String("output_path", *out))
	
	// Output completion message for stages.go to parse
	fmt.Printf("Index extraction complete: %d files\n", processedCount)
}

func loadLastDate(csvPath string) (time.Time, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	// Get file size
	stat, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	size := stat.Size()
	if size == 0 {
		return time.Time{}, fmt.Errorf("empty CSV file")
	}

	// Read last 1KB of file (or entire file if smaller)
	bufSize := int64(1024)
	if bufSize > size {
		bufSize = size
	}

	// Seek to position for reading last chunk
	offset := size - bufSize
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return time.Time{}, err
	}

	// Read the last chunk
	buf := make([]byte, bufSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return time.Time{}, err
	}

	// Split into lines and find last valid date
	lines := strings.Split(string(buf[:n]), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "Date") {
			continue
		}
		
		// Parse CSV line to get date field
		fields := strings.Split(line, ",")
		if len(fields) > 0 && fields[0] != "" {
			// Try to parse the date
			if t, err := time.Parse("2006-01-02", fields[0]); err == nil {
				return t, nil
			}
		}
	}
	
	return time.Time{}, fmt.Errorf("no valid data rows found")
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
