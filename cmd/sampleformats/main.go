package main

import (
	"encoding/csv"
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

var fileRe = regexp.MustCompile(`^(\d{4}) (\d{2}) (\d{2}) ISX Daily Report\.xlsx$`)

func main() {
	dir := flag.String("dir", "downloads", "directory containing reports")
	out := flag.String("out", "index_format_samples.csv", "output file")
	gap := flag.Int("days", 90, "minimum gap between samples in days")
	flag.Parse()

	entries, err := os.ReadDir(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read dir error:", err)
		os.Exit(1)
	}
	type fileInfo struct {
		path string
		date time.Time
		name string
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
		files = append(files, fileInfo{path: filepath.Join(*dir, e.Name()), date: t, name: e.Name()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].date.Before(files[j].date) })

	outF, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer outF.Close()
	w := csv.NewWriter(outF)
	w.Write([]string{"Date", "File", "Sheet", "Row", "Line"})

	var last time.Time
	for _, fi := range files {
		if !last.IsZero() && fi.date.Sub(last).Hours() < float64(*gap*24) {
			continue
		}
		// sample this file
		sheet, row, line, ok := findIndexLine(fi.path)
		if ok {
			w.Write([]string{fi.date.Format("2006-01-02"), fi.name, sheet, fmt.Sprintf("%d", row), line})
		} else {
			w.Write([]string{fi.date.Format("2006-01-02"), fi.name, "", "", "not found"})
		}
		last = fi.date
		fmt.Printf("Sampled %s\n", fi.name)
	}
	w.Flush()
	fmt.Println("Done. Output", *out)
}

func findIndexLine(path string) (sheet string, rowIdx int, line string, ok bool) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return
	}
	keyRe := regexp.MustCompile(`(?i)index\s*(60|price|15)`) // search for keyword index

	for _, sh := range f.GetSheetList() {
		rows, _ := f.GetRows(sh)
		for i, r := range rows {
			joined := strings.TrimSpace(strings.Join(r, " "))
			if keyRe.MatchString(joined) {
				return sh, i + 1, joined, true
			}
			if i > 50 { // only scan first 50 rows for speed
				break
			}
		}
	}
	return
}
