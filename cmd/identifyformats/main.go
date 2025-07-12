package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type Sample struct {
	Quarter string `json:"quarter"`
	File    string `json:"file"`
	Sheet   string `json:"sheet"`
	Row     int    `json:"row"`
	Text    string `json:"text"`
}

var fileRe = regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)

func main() {
	dir := flag.String("dir", "downloads", "directory containing xlsx reports")
	out := flag.String("out", "index_formats.json", "output JSON file")
	flag.Parse()

	entries, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dir failed: %v\n", err)
		os.Exit(1)
	}

	type info struct {
		path string
		date time.Time
	}
	var infos []info
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := fileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		t, _ := time.Parse("2006 01 02", strings.Join(m[1:4], " "))
		infos = append(infos, info{path: filepath.Join(*dir, e.Name()), date: t})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].date.Before(infos[j].date) })

	// keep first file every 3 months (quarter)
	seenQuarter := make(map[string]bool)
	var samples []Sample

	for _, fi := range infos {
		q := quarterKey(fi.date)
		if seenQuarter[q] {
			continue
		}
		seenQuarter[q] = true

		sheet, rowIdx, text := findIndexLine(fi.path)
		samples = append(samples, Sample{
			Quarter: q,
			File:    filepath.Base(fi.path),
			Sheet:   sheet,
			Row:     rowIdx + 1, // 1-based
			Text:    text,
		})
		fmt.Printf("%s -> %s (sheet %s row %d)\n", q, filepath.Base(fi.path), sheet, rowIdx+1)
	}

	// write json
	fp, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create json failed: %v\n", err)
		os.Exit(1)
	}
	defer fp.Close()
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(samples); err != nil {
		fmt.Fprintf(os.Stderr, "encode json: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved %d format samples to %s\n", len(samples), *out)
}

func quarterKey(t time.Time) string {
	q := (int(t.Month())-1)/3 + 1
	return fmt.Sprintf("%04d-Q%d", t.Year(), q)
}

func findIndexLine(path string) (sheet string, rowIdx int, text string) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return "", -1, "open error"
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	// regex to detect line with both indices
	re60 := regexp.MustCompile(`(?i)ISX[^\n]{0,40}60`)
	re15 := regexp.MustCompile(`(?i)ISX[^\n]{0,40}15`)

	for _, sh := range sheetList {
		rows, _ := f.GetRows(sh)
		for i, row := range rows {
			joined := strings.Join(row, " ")
			joined = strings.TrimSpace(joined)
			if joined == "" {
				continue
			}
			if re60.MatchString(joined) && re15.MatchString(joined) {
				return sh, i, joined
			}
		}
	}
	return "", -1, "not found"
}
