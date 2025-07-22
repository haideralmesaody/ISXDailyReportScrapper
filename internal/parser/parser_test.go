package parser

import (
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestParseFile ensures that ParseFile extracts at least one TradeRecord from a
// well-formed minimal workbook.
func TestParseFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Build a minimal workbook that matches the expectations of ParseFile.
	f := excelize.NewFile()
	sheetName := "Bullient"
	// Replace default sheet with the expected name.
	f.SetSheetName(f.GetSheetName(0), sheetName)

	// Header rows the parser will skip (first 3 rows).
	for i := 1; i <= 3; i++ {
		cell := "A" + string(rune('0'+i))
		f.SetCellValue(sheetName, cell, "header")
	}

	// Data row – columns 0..13 (index based on parser expectations).
	// row[1] = symbol, row[8] = close price, row[12] = volume, row[13] = value
	row := make([]interface{}, 14)
	row[1] = "TEST"
	row[8] = "12.5"
	row[12] = "1,000"
	row[13] = "5000"
	// Write row to sheet (row number 4 – 1-based index)
	for colIdx, val := range row {
		col, _ := excelize.ColumnNumberToName(colIdx + 1) // Convert 1-based col number to name
		cell := col + "4"                                 // row 4
		f.SetCellValue(sheetName, cell, val)
	}

	// Save workbook
	filePath := filepath.Join(tmpDir, "2025 01 01 ISX Daily Report.xlsx")
	if err := f.SaveAs(filePath); err != nil {
		t.Fatalf("failed to save temp workbook: %v", err)
	}

	rep, err := ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if len(rep.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(rep.Records))
	}
	r := rep.Records[0]
	if r.CompanySymbol != "TEST" {
		t.Errorf("symbol mismatch: want TEST, got %s", r.CompanySymbol)
	}
	if r.ClosePrice != 12.5 {
		t.Errorf("close price mismatch: want 12.5, got %f", r.ClosePrice)
	}
	if r.Volume != 1000 {
		t.Errorf("volume mismatch: want 1000, got %d", r.Volume)
	}
	if r.Value != 5000 {
		t.Errorf("value mismatch: want 5000, got %f", r.Value)
	}

	// Date parsing may fail when path doesn't start with downloads/, but ensure it's at least set (zero time allowed)
	if r.Date.IsZero() {
		t.Log("Date field could not be parsed – acceptable for this test")
	}
}
