package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: debugindices <xlsx file>")
		os.Exit(1)
	}
	file := os.Args[1]
	f, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatal(err)
	}
	for _, sh := range f.GetSheetList() {
		fmt.Printf("\n=== %s ===\n", sh)
		rows, _ := f.GetRows(sh)
		for i, row := range rows {
			line := strings.Join(row, " | ")
			fmt.Printf("%3d: %s\n", i+1, line)
			if i >= 30 {
				break
			}
		}
	}
}
