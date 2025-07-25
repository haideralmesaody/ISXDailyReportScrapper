package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// TradeRecord represents a single company's trading data for one day.
type TradeRecord struct {
	CompanyName      string
	CompanySymbol    string
	Date             time.Time
	OpenPrice        float64
	HighPrice        float64
	LowPrice         float64
	AveragePrice     float64
	PrevAveragePrice float64
	ClosePrice       float64
	PrevClosePrice   float64
	Change           float64
	ChangePercent    float64
	NumTrades        int64
	Volume           int64
	Value            float64
	TradingStatus    bool // true if actively traded, false if forward-filled
}

// DailyReport represents all trades in a single day's file.
type DailyReport struct {
	Records []TradeRecord
}

// ParseFile reads an ISX daily report Excel file and extracts the trading data.
func ParseFile(filePath string) (*DailyReport, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Find the correct sheet name by looking for one that contains trading data
	var rows [][]string
	var sheetFound bool
	var sheetName string

	// Try different possible sheet names
	possibleNames := []string{"Bullient  ", "Bullient", "Bulletin", "Bulletin  ", "trading", "Trading"}

	for _, name := range possibleNames {
		if testRows, testErr := f.GetRows(name); testErr == nil {
			rows = testRows
			sheetFound = true
			sheetName = name
			break
		}
	}

	// If none of the common names work, try to find a sheet with trading data
	if !sheetFound {
		for _, name := range f.GetSheetList() {
			if testRows, testErr := f.GetRows(name); testErr == nil && len(testRows) > 3 {
				// Check if this sheet contains trading data by looking for typical headers
				for _, row := range testRows[:4] {
					rowText := strings.ToLower(strings.Join(row, " "))
					if strings.Contains(rowText, "company name") && strings.Contains(rowText, "code") &&
						(strings.Contains(rowText, "price") || strings.Contains(rowText, "volume")) {
						rows = testRows
						sheetFound = true
						sheetName = name
						break
					}
				}
				if sheetFound {
					break
				}
			}
		}
	}

	if !sheetFound {
		return nil, fmt.Errorf("could not find trading data sheet in file")
	}

	fmt.Printf("Found trading data in sheet: %s\n", sheetName)
	fmt.Printf("Total rows in sheet: %d\n", len(rows))

	// Print first 20 rows to understand the structure
	fmt.Println("=== First 20 rows ===")
	for i := 0; i < len(rows) && i < 20; i++ {
		fmt.Printf("Row %d: %v\n", i, rows[i])
	}

	// Find the last row with actual data
	lastDataRow := -1
	for i := len(rows) - 1; i >= 0; i-- {
		if len(rows[i]) > 5 {
			// Check if this row has meaningful data (not just empty cells)
			hasData := false
			for _, cell := range rows[i] {
				if strings.TrimSpace(cell) != "" {
					hasData = true
					break
				}
			}
			if hasData {
				lastDataRow = i
				break
			}
		}
	}

	fmt.Printf("Last row with data: %d\n", lastDataRow)
	if lastDataRow > 0 {
		fmt.Printf("Last data row content: %v\n", rows[lastDataRow])
	}

	report := &DailyReport{}
	date, _ := time.Parse("2006 01 02", strings.TrimSuffix(strings.TrimPrefix(filePath, "downloads/"), " ISX Daily Report.xlsx"))

	// Find the header row and map column positions dynamically
	headerRow := -1
	columnMap := make(map[string]int)

	for i, row := range rows {
		if len(row) < 5 {
			continue
		}

		// Look for header row containing key column names
		rowText := strings.ToLower(strings.Join(row, " "))

		// Debug: Show what we're looking for in each row
		fmt.Printf("Row %d text: %s\n", i, rowText)

		// More flexible header detection - look for key trading columns
		if (strings.Contains(rowText, "company") || strings.Contains(rowText, "name")) &&
			strings.Contains(rowText, "code") &&
			(strings.Contains(rowText, "closing") || strings.Contains(rowText, "price")) &&
			strings.Contains(rowText, "volume") {
			headerRow = i
			fmt.Printf("*** FOUND HEADER ROW AT %d ***\n", i)

			// Map column positions based on header names
			for j, header := range row {
				headerLower := strings.ToLower(strings.TrimSpace(header))
				fmt.Printf("  Column %d: '%s'\n", j, headerLower)

				// Map different variations of column names
				switch {
				case strings.Contains(headerLower, "company") || (strings.Contains(headerLower, "name") && !strings.Contains(headerLower, "code")):
					columnMap["company"] = j
					fmt.Printf("    -> Mapped to COMPANY\n")
				case headerLower == "code":
					columnMap["code"] = j
					fmt.Printf("    -> Mapped to CODE\n")
				case strings.Contains(headerLower, "opening") && strings.Contains(headerLower, "price"):
					columnMap["open"] = j
					fmt.Printf("    -> Mapped to OPEN\n")
				case strings.Contains(headerLower, "highest") && strings.Contains(headerLower, "price"):
					columnMap["high"] = j
					fmt.Printf("    -> Mapped to HIGH\n")
				case strings.Contains(headerLower, "lowest") && strings.Contains(headerLower, "price"):
					columnMap["low"] = j
					fmt.Printf("    -> Mapped to LOW\n")
				case strings.Contains(headerLower, "average") && strings.Contains(headerLower, "price") && !strings.Contains(headerLower, "prev"):
					columnMap["avg"] = j
					fmt.Printf("    -> Mapped to AVERAGE\n")
				case strings.Contains(headerLower, "prev") && strings.Contains(headerLower, "average"):
					columnMap["prev_avg"] = j
					fmt.Printf("    -> Mapped to PREV_AVERAGE\n")
				case strings.Contains(headerLower, "closing") && strings.Contains(headerLower, "price"):
					columnMap["close"] = j
					fmt.Printf("    -> Mapped to CLOSE\n")
				case strings.Contains(headerLower, "prev") && strings.Contains(headerLower, "closing"):
					columnMap["prev_close"] = j
					fmt.Printf("    -> Mapped to PREV_CLOSE\n")
				case strings.Contains(headerLower, "change") && strings.Contains(headerLower, "%"):
					columnMap["change_pct"] = j
					fmt.Printf("    -> Mapped to CHANGE_PCT\n")
				case strings.Contains(headerLower, "no") && strings.Contains(headerLower, "trades"):
					columnMap["num_trades"] = j
					fmt.Printf("    -> Mapped to NUM_TRADES\n")
				case headerLower == "traded volume":
					columnMap["volume"] = j
					fmt.Printf("    -> Mapped to VOLUME\n")
				case headerLower == "traded value":
					columnMap["value"] = j
					fmt.Printf("    -> Mapped to VALUE\n")
				}
			}
			fmt.Printf("Final column mapping: %+v\n", columnMap)
			break
		}
	}

	if headerRow == -1 {
		return nil, fmt.Errorf("could not find header row in trading data")
	}

	// Verify we found all required columns
	requiredCols := []string{"code", "close", "volume", "value"}
	for _, col := range requiredCols {
		if _, exists := columnMap[col]; !exists {
			return nil, fmt.Errorf("could not find required column: %s", col)
		}
	}

	// Process data rows starting after the header, up to the last data row
	dataEndRow := len(rows)
	if lastDataRow > 0 {
		dataEndRow = lastDataRow + 1
	}

	fmt.Printf("Processing data rows from %d to %d\n", headerRow+1, dataEndRow-1)

	for i := headerRow + 1; i < dataEndRow; i++ {
		row := rows[i]

		fmt.Printf("Processing row %d: %v\n", i, row)

		// Skip if not enough columns
		if len(row) <= columnMap["value"] {
			fmt.Printf("  -> Skipped: Not enough columns (need %d, got %d)\n", columnMap["value"]+1, len(row))
			continue
		}

		// Skip empty rows - check if all relevant columns are empty
		isEmpty := true
		for _, colIndex := range columnMap {
			if colIndex < len(row) && strings.TrimSpace(row[colIndex]) != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			fmt.Printf("  -> Skipped: Empty row\n")
			continue
		}

		// Skip sector headers (merged cells or rows containing "Sector")
		if strings.Contains(row[0], "Sector") || strings.Contains(row[0], "Total") {
			fmt.Printf("  -> Skipped: Sector/Total row\n")
			continue
		}

		// Skip if code column is empty (likely a merged/header row)
		if columnMap["code"] < len(row) && strings.TrimSpace(row[columnMap["code"]]) == "" {
			fmt.Printf("  -> Skipped: Empty code column\n")
			continue
		}

		// Extract data using dynamic column mapping
		companyCode := strings.TrimSpace(row[columnMap["code"]])
		if companyCode == "" {
			fmt.Printf("  -> Skipped: Empty company code after trim\n")
			continue
		}

		fmt.Printf("  -> Processing: Code=%s\n", companyCode)

		// Helper function to safely parse float
		parseFloat := func(colName string) float64 {
			if idx, exists := columnMap[colName]; exists && idx < len(row) {
				val, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(row[idx]), ",", ""), 64)
				return val
			}
			return 0.0
		}

		// Helper function to safely parse int
		parseInt := func(colName string) int64 {
			if idx, exists := columnMap[colName]; exists && idx < len(row) {
				val, _ := strconv.ParseInt(strings.ReplaceAll(strings.TrimSpace(row[idx]), ",", ""), 10, 64)
				return val
			}
			return 0
		}

		// Helper function to safely get string
		getString := func(colName string) string {
			if idx, exists := columnMap[colName]; exists && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		// Extract all available fields
		companyName := getString("company")
		openPrice := parseFloat("open")
		highPrice := parseFloat("high")
		lowPrice := parseFloat("low")
		avgPrice := parseFloat("avg")
		prevAvgPrice := parseFloat("prev_avg")
		closePrice := parseFloat("close")
		prevClosePrice := parseFloat("prev_close")
		changePercent := parseFloat("change_pct")
		numTrades := parseInt("num_trades")
		volume := parseInt("volume")
		value := parseFloat("value")

		// Calculate change if not available
		change := closePrice - prevClosePrice

		record := TradeRecord{
			CompanyName:      companyName,
			CompanySymbol:    companyCode,
			Date:             date,
			OpenPrice:        openPrice,
			HighPrice:        highPrice,
			LowPrice:         lowPrice,
			AveragePrice:     avgPrice,
			PrevAveragePrice: prevAvgPrice,
			ClosePrice:       closePrice,
			PrevClosePrice:   prevClosePrice,
			Change:           change,
			ChangePercent:    changePercent,
			NumTrades:        numTrades,
			Volume:           volume,
			Value:            value,
			TradingStatus:    true, // Actual trading data
		}
		report.Records = append(report.Records, record)

		// Debug: Show first few records
		if len(report.Records) <= 5 {
			fmt.Printf("Record %d: %s (%s) - Open: %.3f, High: %.3f, Low: %.3f, Close: %.3f, Volume: %d, Value: %.2f\n",
				len(report.Records), companyCode, companyName, openPrice, highPrice, lowPrice, closePrice, volume, value)
		}
	}

	fmt.Printf("Total records processed: %d\n", len(report.Records))

	return report, nil
}
