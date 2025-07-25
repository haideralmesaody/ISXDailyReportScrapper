package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// TickerSummary represents summary data for a ticker
type TickerSummary struct {
	Ticker       string  `json:"ticker"`
	LastPrice    float64 `json:"last_price"`
	LastDate     string  `json:"last_date"`
	TotalVolume  float64 `json:"total_volume"`
	TotalTrades  int64   `json:"total_trades"`
	ActualDays   int     `json:"actual_days"`
	AvgPrice     float64 `json:"avg_price"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	StdDev       float64 `json:"std_dev"`
}

// generateTickerSummaryJSON creates a JSON ticker summary from the combined CSV file
func generateTickerSummaryJSON(outDir string) error {
	combinedFile := filepath.Join(outDir, "isx_combined_data.csv")
	summaryFile := filepath.Join(outDir, "ticker_summary.json")

	// Open combined CSV file
	file, err := os.Open(combinedFile)
	if err != nil {
		return fmt.Errorf("failed to open combined CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Create column index map
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[col] = i
	}

	// Verify required columns exist
	requiredColumns := []string{"Ticker", "Date", "ClosingPrice", "Volume", "Trades", "TradingStatus"}
	for _, col := range requiredColumns {
		if _, exists := columnIndex[col]; !exists {
			return fmt.Errorf("required column %s not found", col)
		}
	}

	// Process records
	tickerData := make(map[string]*TickerSummary)
	tickerPrices := make(map[string][]float64)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %v", err)
		}

		ticker := record[columnIndex["Ticker"]]
		
		// Skip if TradingStatus is false (forward-filled data)
		if record[columnIndex["TradingStatus"]] != "true" {
			continue
		}

		// Parse data
		price, _ := strconv.ParseFloat(record[columnIndex["ClosingPrice"]], 64)
		volume, _ := strconv.ParseFloat(record[columnIndex["Volume"]], 64)
		trades, _ := strconv.ParseInt(record[columnIndex["Trades"]], 10, 64)
		date := record[columnIndex["Date"]]

		// Initialize ticker if not exists
		if _, exists := tickerData[ticker]; !exists {
			tickerData[ticker] = &TickerSummary{
				Ticker:   ticker,
				MinPrice: price,
				MaxPrice: price,
			}
			tickerPrices[ticker] = []float64{}
		}

		// Update summary data
		summary := tickerData[ticker]
		summary.LastPrice = price
		summary.LastDate = date
		summary.TotalVolume += volume
		summary.TotalTrades += trades
		summary.ActualDays++

		// Track prices for statistics
		tickerPrices[ticker] = append(tickerPrices[ticker], price)

		// Update min/max
		if price < summary.MinPrice {
			summary.MinPrice = price
		}
		if price > summary.MaxPrice {
			summary.MaxPrice = price
		}
	}

	// Calculate statistics for each ticker
	for ticker, prices := range tickerPrices {
		if len(prices) == 0 {
			continue
		}

		summary := tickerData[ticker]

		// Calculate average
		sum := 0.0
		for _, p := range prices {
			sum += p
		}
		summary.AvgPrice = sum / float64(len(prices))

		// Calculate standard deviation
		if len(prices) > 1 {
			sumSquaredDiff := 0.0
			for _, p := range prices {
				diff := p - summary.AvgPrice
				sumSquaredDiff += diff * diff
			}
			summary.StdDev = math.Sqrt(sumSquaredDiff / float64(len(prices)-1))
		}
	}

	// Convert map to slice and sort by ticker
	var summaries []TickerSummary
	for _, summary := range tickerData {
		summaries = append(summaries, *summary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Ticker < summaries[j].Ticker
	})

	// Write JSON file
	jsonData, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(summaryFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	fmt.Printf("Generated ticker summary with %d tickers\n", len(summaries))
	return nil
}