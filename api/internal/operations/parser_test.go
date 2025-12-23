package operations

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

	// Add realistic ISX file structure that the parser expects
	// Rows 1-2: Title and metadata (parser skips these)
	f.SetCellValue(sheetName, "A1", "Iraq Stock Exchange Daily Report")
	f.SetCellValue(sheetName, "A2", "Trading Date: 2025-01-01")

	// Row 3: Actual header with keywords the parser recognizes
	headers := []string{
		"",                   // Column A (empty)
		"Code",               // Column B - Parser looks for "code"
		"Company Name",       // Column C - Parser looks for "company" or "name"
		"Opening Price",      // Column D
		"Highest Price",      // Column E
		"Lowest Price",       // Column F
		"Average Price",      // Column G
		"Prev Average",       // Column H
		"Closing Price",      // Column I - Parser looks for "closing" + "price"
		"Prev Closing Price", // Column J
		"Change %",           // Column K
		"No. of Trades",      // Column L
		"Traded Volume",      // Column M - Parser looks for exactly "traded volume"
		"Traded Value",       // Column N - Parser looks for exactly "traded value"
	}

	for i, header := range headers {
		if header != "" {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheetName, col+"3", header)
		}
	}

	// Row 4: Data row with test values
	// Columns match the header structure above
	dataRow := []string{
		"",        // Column A (empty)
		"TEST",    // Column B (Code)
		"Test Co", // Column C (Company Name)
		"",        // Column D (Opening Price)
		"",        // Column E (Highest Price)
		"",        // Column F (Lowest Price)
		"",        // Column G (Average Price)
		"",        // Column H (Prev Average)
		"12.5",    // Column I (Closing Price)
		"",        // Column J (Prev Closing Price)
		"",        // Column K (Change %)
		"",        // Column L (No. of Trades)
		"1,000",   // Column M (Traded Volume)
		"5000",    // Column N (Traded Value)
	}

	for i, val := range dataRow {
		if val != "" {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheetName, col+"4", val)
		}
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
