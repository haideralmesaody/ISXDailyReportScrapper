package dataprocessing

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"log/slog"
	
	"github.com/isxcli/isxcli/pkg/contracts/domain"
)


// ParseFile reads an ISX daily report Excel file and extracts the trading data.
func ParseFile(filePath string) (*domain.DailyReport, error) {
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

	slog.Info("Found trading data in sheet", slog.String("sheet_name", sheetName))
	slog.Info("Sheet information", slog.Int("total_rows", len(rows)))

	// Print first 20 rows to understand the structure
	slog.Info("=== First 20 rows ===")
	for i := 0; i < len(rows) && i < 20; i++ {
		slog.Info("Row data", slog.Int("row_number", i), slog.Any("content", rows[i]))
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

	slog.Info("Data analysis", slog.Int("last_data_row", lastDataRow))
	if lastDataRow > 0 {
		slog.Debug("Last data row found", 
			slog.Int("row_index", lastDataRow),
			slog.Any("row_content", rows[lastDataRow]))
	}

	report := &domain.DailyReport{}
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
		slog.Info("Row analysis", slog.Int("row_number", i), slog.String("text", rowText))

		// More flexible header detection - look for key trading columns
		if (strings.Contains(rowText, "company") || strings.Contains(rowText, "name")) &&
			strings.Contains(rowText, "code") &&
			(strings.Contains(rowText, "closing") || strings.Contains(rowText, "price")) &&
			strings.Contains(rowText, "volume") {
			headerRow = i
			slog.Info("*** FOUND HEADER ROW ***", slog.Int("row_number", i))

			// Map column positions based on header names
			for j, header := range row {
				headerLower := strings.ToLower(strings.TrimSpace(header))
				slog.Info("Column mapping", slog.Int("column_index", j), slog.String("header", headerLower))

				// Map different variations of column names
				switch {
				case strings.Contains(headerLower, "company") || (strings.Contains(headerLower, "name") && !strings.Contains(headerLower, "code")):
					columnMap["company"] = j
					slog.Info("Column mapped", slog.String("type", "COMPANY"), slog.Int("index", j), slog.String("header", headerLower))
				case headerLower == "code":
					columnMap["code"] = j
					slog.Info("Column mapped", slog.String("type", "CODE"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "opening") && strings.Contains(headerLower, "price"):
					columnMap["open"] = j
					slog.Info("Column mapped", slog.String("type", "OPEN"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "highest") && strings.Contains(headerLower, "price"):
					columnMap["high"] = j
					slog.Info("Column mapped", slog.String("type", "HIGH"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "lowest") && strings.Contains(headerLower, "price"):
					columnMap["low"] = j
					slog.Info("Column mapped", slog.String("type", "LOW"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "average") && strings.Contains(headerLower, "price") && !strings.Contains(headerLower, "prev"):
					columnMap["avg"] = j
					slog.Info("Column mapped", slog.String("type", "AVERAGE"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "prev") && strings.Contains(headerLower, "average"):
					columnMap["prev_avg"] = j
					slog.Info("Column mapped", slog.String("type", "PREV_AVERAGE"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "closing") && strings.Contains(headerLower, "price") && !strings.Contains(headerLower, "prev"):
					columnMap["close"] = j
					slog.Info("Column mapped", slog.String("type", "CLOSE"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "prev") && strings.Contains(headerLower, "closing"):
					columnMap["prev_close"] = j
					slog.Info("Column mapped", slog.String("type", "PREV_CLOSE"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "change") && strings.Contains(headerLower, "%"):
					columnMap["change_pct"] = j
					slog.Info("Column mapped", slog.String("type", "CHANGE_PCT"), slog.Int("index", j), slog.String("header", headerLower))
				case strings.Contains(headerLower, "no") && strings.Contains(headerLower, "trades"):
					columnMap["num_trades"] = j
					slog.Info("Column mapped", slog.String("type", "NUM_TRADES"), slog.Int("index", j), slog.String("header", headerLower))
				case headerLower == "traded volume":
					columnMap["volume"] = j
					slog.Info("Column mapped", slog.String("type", "VOLUME"), slog.Int("index", j), slog.String("header", headerLower))
				case headerLower == "traded value":
					columnMap["value"] = j
					slog.Info("Column mapped", slog.String("type", "VALUE"), slog.Int("index", j), slog.String("header", headerLower))
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

	slog.Info("Processing data rows", 
		slog.Int("start_row", headerRow+1),
		slog.Int("end_row", dataEndRow-1))

	for i := headerRow + 1; i < dataEndRow; i++ {
		row := rows[i]

		slog.Info("Processing row", slog.Int("row_number", i), slog.Any("content", row))

		// Skip if not enough columns
		if len(row) <= columnMap["value"] {
			slog.Info("Skipped row - insufficient columns",
				slog.Int("needed", columnMap["value"]+1),
				slog.Int("got", len(row)))
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

		slog.Info("Processing company", slog.String("code", companyCode))
		
		// Debug logging for BBOB specifically
		if companyCode == "BBOB" {
			slog.Info("BBOB Row Data Debug")
			for colName, colIdx := range columnMap {
				if colIdx < len(row) {
					slog.Info("BBOB column value", 
						slog.String("column", colName), 
						slog.Int("index", colIdx), 
						slog.String("value", row[colIdx]))
				}
			}
		}

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

		record := domain.TradeRecord{
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
			slog.Debug("Record parsed", 
				slog.Int("record_number", len(report.Records)),
				slog.String("company_code", companyCode),
				slog.String("company_name", companyName),
				slog.Float64("open_price", openPrice),
				slog.Float64("high_price", highPrice),
				slog.Float64("low_price", lowPrice),
				slog.Float64("close_price", closePrice),
				slog.Int64("volume", volume),
				slog.Float64("value", value))
		}
	}

	slog.Info("Processing complete", slog.Int("total_records", len(report.Records)))

	return report, nil
}
