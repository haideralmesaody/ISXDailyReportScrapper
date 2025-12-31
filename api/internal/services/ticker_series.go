package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/isxcli/isxcli/internal/liquidity"
)

const tickerTradingHistorySuffix = "_trading_history.csv"

// ListTickerSymbols returns all ticker symbols that have a trading history CSV.
// Data source: <reports_dir>/ticker/<SYMBOL>_trading_history.csv
func (ds *DataService) ListTickerSymbols(ctx context.Context) ([]string, error) {
	if ds == nil || ds.paths == nil {
		return nil, fmt.Errorf("data service not initialized")
	}

	tickerDir := ds.paths.TickerReportsDir
	entries, err := os.ReadDir(tickerDir)
	if err != nil {
		return nil, fmt.Errorf("read ticker reports dir (%s): %w", tickerDir, err)
	}

	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), tickerTradingHistorySuffix) {
			continue
		}

		symbol := strings.TrimSuffix(name, tickerTradingHistorySuffix)
		symbol = strings.TrimSpace(strings.ToUpper(symbol))
		if symbol == "" {
			continue
		}
		seen[symbol] = struct{}{}
	}

	symbols := make([]string, 0, len(seen))
	for symbol := range seen {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	return symbols, nil
}

// LoadTickerTradingHistory loads a ticker's trading history from the processed reports output.
// Implementation detail: uses ClosePrice and skips rows where ClosePrice == 0.
func (ds *DataService) LoadTickerTradingHistory(ctx context.Context, symbol string) ([]liquidity.TradingDay, error) {
	if ds == nil || ds.paths == nil {
		return nil, fmt.Errorf("data service not initialized")
	}

	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	filePath := filepath.Join(ds.paths.TickerReportsDir, symbol+tickerTradingHistorySuffix)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open ticker history (%s): %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read ticker history header: %w", err)
	}

	index := make(map[string]int, len(header))
	for i, col := range header {
		key := strings.TrimSpace(col)
		if key == "" {
			continue
		}
		index[key] = i
	}

	required := []string{"Date", "Symbol", "OpenPrice", "HighPrice", "LowPrice", "ClosePrice"}
	for _, col := range required {
		if _, ok := index[col]; !ok {
			return nil, fmt.Errorf("ticker history missing required column %q", col)
		}
	}

	tradingDays := make([]liquidity.TradingDay, 0, 512)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read ticker history row: %w", readErr)
		}

		// Guard against malformed rows.
		if index["ClosePrice"] >= len(record) || index["Date"] >= len(record) || index["Symbol"] >= len(record) {
			continue
		}

		date, err := time.Parse("2006-01-02", strings.TrimSpace(record[index["Date"]]))
		if err != nil {
			continue
		}

		rowSymbol := strings.TrimSpace(strings.ToUpper(record[index["Symbol"]]))
		if rowSymbol == "" {
			continue
		}

		closePrice, err := strconv.ParseFloat(strings.TrimSpace(record[index["ClosePrice"]]), 64)
		if err != nil {
			continue
		}
		// Confirmed rule: skip rows where ClosePrice == 0 (no effective trading).
		if closePrice == 0 {
			continue
		}

		openPrice := parseFloatOrZero(record, index, "OpenPrice")
		highPrice := parseFloatOrZero(record, index, "HighPrice")
		lowPrice := parseFloatOrZero(record, index, "LowPrice")

		numTrades := parseIntOrZero(record, index, "NumTrades")
		volume := parseFloatOrZero(record, index, "Volume")
		value := parseFloatOrZero(record, index, "Value")
		if value == 0 && volume != 0 && closePrice != 0 {
			value = closePrice * volume
		}

		tradingStatus := "ACTIVE"
		if idx, ok := index["TradingStatus"]; ok && idx < len(record) {
			if strings.EqualFold(strings.TrimSpace(record[idx]), "false") {
				tradingStatus = "SUSPENDED"
			}
		}

		tradingDays = append(tradingDays, liquidity.TradingDay{
			Date:          date,
			Symbol:        rowSymbol,
			Open:          openPrice,
			High:          highPrice,
			Low:           lowPrice,
			Close:         closePrice,
			Volume:        volume,
			ShareVolume:   volume,
			Value:         value,
			NumTrades:     numTrades,
			TradingStatus: tradingStatus,
		})
	}

	if len(tradingDays) == 0 {
		return nil, fmt.Errorf("ticker history has no usable rows for %s", symbol)
	}

	sort.Slice(tradingDays, func(i, j int) bool {
		return tradingDays[i].Date.Before(tradingDays[j].Date)
	})
	return tradingDays, nil
}

// LoadTickerTradingHistoryTail returns the last N trading days for a symbol.
func (ds *DataService) LoadTickerTradingHistoryTail(ctx context.Context, symbol string, n int) ([]liquidity.TradingDay, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be > 0")
	}
	data, err := ds.LoadTickerTradingHistory(ctx, symbol)
	if err != nil {
		return nil, err
	}
	if len(data) <= n {
		return data, nil
	}
	return data[len(data)-n:], nil
}

// LoadTickerTradingHistoryRange returns trading days between start and end (inclusive).
func (ds *DataService) LoadTickerTradingHistoryRange(ctx context.Context, symbol string, start, end time.Time) ([]liquidity.TradingDay, error) {
	data, err := ds.LoadTickerTradingHistory(ctx, symbol)
	if err != nil {
		return nil, err
	}

	startKey := start.UTC().Format("2006-01-02")
	endKey := end.UTC().Format("2006-01-02")

	filtered := make([]liquidity.TradingDay, 0, len(data))
	for _, day := range data {
		dayKey := day.Date.UTC().Format("2006-01-02")
		if dayKey < startKey || dayKey > endKey {
			continue
		}
		filtered = append(filtered, day)
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no ticker history rows for %s in range %s..%s", symbol, startKey, endKey)
	}

	return filtered, nil
}

func parseFloatOrZero(record []string, index map[string]int, col string) float64 {
	idx, ok := index[col]
	if !ok || idx >= len(record) {
		return 0
	}
	value := strings.TrimSpace(record[idx])
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseIntOrZero(record []string, index map[string]int, col string) int {
	idx, ok := index[col]
	if !ok || idx >= len(record) {
		return 0
	}
	value := strings.TrimSpace(record[idx])
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
