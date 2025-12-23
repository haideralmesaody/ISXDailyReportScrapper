package operations

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isxcli/isxcli/pkg/contracts/domain"
)

func TestNewSummarizer(t *testing.T) {
	tests := []struct {
		name     string
		logger   *slog.Logger
		config   SummarizerConfig
		wantDays int
		wantFmt  string
	}{
		{
			name:     "default config",
			logger:   slog.Default(),
			config:   DefaultSummarizerConfig(),
			wantDays: 10,
			wantFmt:  "2006-01-02",
		},
		{
			name:     "extended config",
			logger:   slog.Default(),
			config:   ExtendedSummarizerConfig(),
			wantDays: 10,
			wantFmt:  "2006-01-02",
		},
		{
			name:   "custom config",
			logger: slog.Default(),
			config: SummarizerConfig{
				IncludeExtendedMetrics: true,
				MaxLast10Days:          5,
				DateFormat:             "01/02/2006",
			},
			wantDays: 5,
			wantFmt:  "01/02/2006",
		},
		{
			name:     "nil logger uses default",
			logger:   nil,
			config:   DefaultSummarizerConfig(),
			wantDays: 10,
			wantFmt:  "2006-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summarizer := NewSummarizer(tt.logger, tt.config)

			assert.NotNil(t, summarizer)
			assert.Equal(t, tt.wantDays, summarizer.maxLast10Days)
			assert.Equal(t, tt.wantFmt, summarizer.dateFormat)
			assert.NotNil(t, summarizer.logger)
		})
	}
}

func TestSummarizer_GenerateFromRecords(t *testing.T) {
	ctx := context.Background()
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	tests := []struct {
		name    string
		records []domain.TradeRecord
		want    int
		wantErr bool
	}{
		{
			name:    "empty records",
			records: []domain.TradeRecord{},
			want:    0,
			wantErr: false,
		},
		{
			name: "single ticker with trading activity",
			records: []domain.TradeRecord{
				{
					CompanySymbol: "BASH",
					CompanyName:   "Bank of Baghdad",
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					Volume:        1000,
					NumTrades:     10,
					TradingStatus: true,
				},
				{
					CompanySymbol: "BASH",
					CompanyName:   "Bank of Baghdad",
					Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.520,
					Volume:        0,
					NumTrades:     0,
					TradingStatus: false, // Forward-filled
				},
				{
					CompanySymbol: "BASH",
					CompanyName:   "Bank of Baghdad",
					Date:          time.Date(2024, 8, 13, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.520,
					Volume:        0,
					NumTrades:     0,
					TradingStatus: false, // Forward-filled
				},
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "multiple tickers",
			records: []domain.TradeRecord{
				{
					CompanySymbol: "BASH",
					CompanyName:   "Bank of Baghdad",
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					TradingStatus: true,
				},
				{
					CompanySymbol: "TAQA",
					CompanyName:   "National Company for Tourism Investments",
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    12.000,
					TradingStatus: true,
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "ticker with empty symbol",
			records: []domain.TradeRecord{
				{
					CompanySymbol: "",
					CompanyName:   "Invalid Company",
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.000,
					TradingStatus: true,
				},
				{
					CompanySymbol: "BASH",
					CompanyName:   "Bank of Baghdad",
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					TradingStatus: true,
				},
			},
			want:    1, // Only BASH should be included
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summaries, err := summarizer.GenerateFromRecords(ctx, tt.records)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, summaries, tt.want)

			// Verify summaries are sorted by ticker
			for i := 1; i < len(summaries); i++ {
				assert.True(t, summaries[i-1].Ticker <= summaries[i].Ticker,
					"summaries should be sorted by ticker")
			}
		})
	}
}

func TestSummarizer_FindLastTradingRecord(t *testing.T) {
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	tests := []struct {
		name      string
		records   []domain.TradeRecord
		wantIndex int
		wantPrice float64
		wantDate  string
	}{
		{
			name: "last record has trading status true",
			records: []domain.TradeRecord{
				{
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					TradingStatus: true,
				},
				{
					Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.520,
					TradingStatus: false,
				},
				{
					Date:          time.Date(2024, 8, 13, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.530,
					TradingStatus: true,
				},
			},
			wantIndex: 2,
			wantPrice: 1.530,
			wantDate:  "2024-08-13",
		},
		{
			name: "fallback to volume check",
			records: []domain.TradeRecord{
				{
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					Volume:        1000,
					TradingStatus: false,
				},
				{
					Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.520,
					Volume:        0,
					TradingStatus: false,
				},
			},
			wantIndex: 0,
			wantPrice: 1.500,
			wantDate:  "2024-08-11",
		},
		{
			name: "fallback to numTrades check",
			records: []domain.TradeRecord{
				{
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					NumTrades:     5,
					TradingStatus: false,
				},
				{
					Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.520,
					NumTrades:     0,
					TradingStatus: false,
				},
			},
			wantIndex: 0,
			wantPrice: 1.500,
			wantDate:  "2024-08-11",
		},
		{
			name: "no trading activity found",
			records: []domain.TradeRecord{
				{
					Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
					ClosePrice:    1.500,
					Volume:        0,
					NumTrades:     0,
					TradingStatus: false,
				},
			},
			wantIndex: -1, // No trading activity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, index := summarizer.findLastTradingRecord(tt.records)

			assert.Equal(t, tt.wantIndex, index)

			if tt.wantIndex >= 0 {
				require.NotNil(t, record)
				assert.Equal(t, tt.wantPrice, record.ClosePrice)
				assert.Equal(t, tt.wantDate, record.Date.Format("2006-01-02"))
			} else {
				assert.Nil(t, record)
			}
		})
	}
}

func TestSummarizer_CountTradingDays(t *testing.T) {
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	tests := []struct {
		name    string
		records []domain.TradeRecord
		want    int
	}{
		{
			name:    "empty records",
			records: []domain.TradeRecord{},
			want:    0,
		},
		{
			name: "all trading days",
			records: []domain.TradeRecord{
				{TradingStatus: true},
				{TradingStatus: true},
				{TradingStatus: true},
			},
			want: 3,
		},
		{
			name: "mixed trading and non-trading days",
			records: []domain.TradeRecord{
				{TradingStatus: true},
				{TradingStatus: false, Volume: 0, NumTrades: 0},
				{TradingStatus: false, Volume: 100, NumTrades: 0}, // Volume fallback
				{TradingStatus: false, Volume: 0, NumTrades: 5},   // NumTrades fallback
			},
			want: 3, // First, third, and fourth records
		},
		{
			name: "no trading days",
			records: []domain.TradeRecord{
				{TradingStatus: false, Volume: 0, NumTrades: 0},
				{TradingStatus: false, Volume: 0, NumTrades: 0},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizer.countTradingDays(tt.records)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSummarizer_GetLastTradingPrices(t *testing.T) {
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	tests := []struct {
		name    string
		records []domain.TradeRecord
		maxDays int
		want    []float64
	}{
		{
			name:    "empty records",
			records: []domain.TradeRecord{},
			maxDays: 10,
			want:    []float64{},
		},
		{
			name: "get last 3 trading days",
			records: []domain.TradeRecord{
				{ClosePrice: 1.000, TradingStatus: true},
				{ClosePrice: 1.100, TradingStatus: false}, // Skip
				{ClosePrice: 1.200, TradingStatus: true},
				{ClosePrice: 1.300, TradingStatus: true},
				{ClosePrice: 1.400, TradingStatus: false}, // Skip
				{ClosePrice: 1.500, TradingStatus: true},
			},
			maxDays: 3,
			want:    []float64{1.200, 1.300, 1.500}, // Last 3 trading days in chronological order
		},
		{
			name: "more maxDays than trading days",
			records: []domain.TradeRecord{
				{ClosePrice: 1.000, TradingStatus: true},
				{ClosePrice: 1.200, TradingStatus: true},
			},
			maxDays: 5,
			want:    []float64{1.000, 1.200}, // All available trading days
		},
		{
			name: "volume fallback",
			records: []domain.TradeRecord{
				{ClosePrice: 1.000, TradingStatus: false, Volume: 100},
				{ClosePrice: 1.100, TradingStatus: false, Volume: 0},
				{ClosePrice: 1.200, TradingStatus: false, NumTrades: 5},
			},
			maxDays: 10,
			want:    []float64{1.000, 1.200}, // First and third records have activity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizer.getLastTradingPrices(tt.records, tt.maxDays)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSummarizer_CalculatePercentageChanges(t *testing.T) {
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	tests := []struct {
		name   string
		prices []float64
		want   float64
	}{
		{
			name:   "empty prices",
			prices: []float64{},
			want:   0.0,
		},
		{
			name:   "single price",
			prices: []float64{1.000},
			want:   0.0,
		},
		{
			name:   "price increase",
			prices: []float64{1.000, 1.100},
			want:   10.0, // 10% increase
		},
		{
			name:   "price decrease",
			prices: []float64{1.100, 1.000},
			want:   -9.090909090909092, // ~9.09% decrease
		},
		{
			name:   "no change",
			prices: []float64{1.000, 1.000},
			want:   0.0,
		},
		{
			name:   "zero previous price",
			prices: []float64{0.0, 1.000},
			want:   0.0, // Handle division by zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizer.calculateDailyChangePercent(tt.prices)
			assert.InDelta(t, tt.want, got, 0.0001, "daily change calculation")
		})
	}
}

func TestSummarizer_WriteCSV(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		includeExtended bool
		summaries       []TickerSummary
		wantHeaderCount int
	}{
		{
			name:            "basic CSV output",
			includeExtended: false,
			summaries: []TickerSummary{
				{
					Ticker:      "BASH",
					CompanyName: "Bank of Baghdad",
					LastPrice:   1.500,
					LastDate:    "2024-08-11",
					TradingDays: 5,
					Last10Days:  []float64{1.400, 1.450, 1.500},
				},
			},
			wantHeaderCount: 9, // Basic fields + Change, ChangePercent, LastTradingStatus
		},
		{
			name:            "extended CSV output",
			includeExtended: true,
			summaries: []TickerSummary{
				{
					Ticker:       "BASH",
					CompanyName:  "Bank of Baghdad",
					LastPrice:    1.500,
					LastDate:     "2024-08-11",
					TradingDays:  5,
					Last10Days:   []float64{1.400, 1.450, 1.500},
					TotalVolume:  1000,
					TotalValue:   1500.0,
					AveragePrice: 1.450,
					HighestPrice: 1.500,
					LowestPrice:  1.400,
				},
			},
			wantHeaderCount: 14, // Basic + extended fields + Change, ChangePercent, LastTradingStatus
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSummarizerConfig()
			config.IncludeExtendedMetrics = tt.includeExtended
			summarizer := NewSummarizer(slog.Default(), config)

			csvPath := filepath.Join(tempDir, tt.name+".csv")
			err := summarizer.WriteCSV(ctx, csvPath, tt.summaries)
			require.NoError(t, err)

			// Verify file exists and has content
			content, err := os.ReadFile(csvPath)
			require.NoError(t, err)
			assert.NotEmpty(t, content)

			// Verify header count
			lines := strings.Split(string(content), "\n")
			require.GreaterOrEqual(t, len(lines), 2) // Header + at least one data row

			headers := strings.Split(lines[0], ",")
			assert.Len(t, headers, tt.wantHeaderCount)
		})
	}
}

func TestSummarizer_WriteJSON(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	summaries := []TickerSummary{
		{
			Ticker:      "BASH",
			CompanyName: "Bank of Baghdad",
			LastPrice:   1.500,
			LastDate:    "2024-08-11",
			TradingDays: 5,
			Last10Days:  []float64{1.400, 1.450, 1.500},
		},
		{
			Ticker:      "TAQA",
			CompanyName: "National Company for Tourism Investments",
			LastPrice:   12.000,
			LastDate:    "2024-08-11",
			TradingDays: 3,
			Last10Days:  []float64{11.800, 11.900, 12.000},
		},
	}

	jsonPath := filepath.Join(tempDir, "test_summaries.json")
	err := summarizer.WriteJSON(ctx, jsonPath, summaries)
	require.NoError(t, err)

	// Verify file exists and has valid JSON
	content, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	// Parse and verify JSON structure
	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	require.NoError(t, err)

	// Verify required fields
	assert.Contains(t, jsonData, "tickers")
	assert.Contains(t, jsonData, "count")
	assert.Contains(t, jsonData, "generated_at")
	assert.Contains(t, jsonData, "format")

	// Verify count matches
	assert.Equal(t, float64(len(summaries)), jsonData["count"])
	assert.Equal(t, "ticker_summary_v1", jsonData["format"])

	// Verify tickers array
	tickers, ok := jsonData["tickers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tickers, len(summaries))
}

func TestSummarizer_FormatLast10Days(t *testing.T) {
	summarizer := NewSummarizer(slog.Default(), DefaultSummarizerConfig())

	tests := []struct {
		name   string
		prices []float64
		want   string
	}{
		{
			name:   "empty prices",
			prices: []float64{},
			want:   "",
		},
		{
			name:   "single price",
			prices: []float64{1.500},
			want:   "1.500",
		},
		{
			name:   "multiple prices",
			prices: []float64{1.400, 1.450, 1.500},
			want:   "1.400,1.450,1.500",
		},
		{
			name:   "prices with many decimals",
			prices: []float64{1.123456, 2.987654},
			want:   "1.123,2.988", // Rounded to 3 decimal places
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizer.formatLast10Days(tt.prices)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSummarizer_RealWorldScenario(t *testing.T) {
	// Test with realistic BASH data showing the correct last trading date logic
	ctx := context.Background()
	summarizer := NewSummarizer(slog.Default(), ExtendedSummarizerConfig())

	records := []domain.TradeRecord{
		// August 11 - Last actual trading day
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500,
			Volume:        1000,
			NumTrades:     10,
			TradingStatus: true,
		},
		// August 12 - Forward-filled (no trading)
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500, // Same price as previous day
			Volume:        0,
			NumTrades:     0,
			TradingStatus: false,
		},
		// August 13 - Forward-filled (no trading)
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 13, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500, // Same price as previous day
			Volume:        0,
			NumTrades:     0,
			TradingStatus: false,
		},
	}

	summaries, err := summarizer.GenerateFromRecords(ctx, records)
	require.NoError(t, err)
	require.Len(t, summaries, 1)

	summary := summaries[0]
	assert.Equal(t, "BASH", summary.Ticker)
	assert.Equal(t, "Bank of Baghdad", summary.CompanyName)
	assert.Equal(t, 1.500, summary.LastPrice)
	assert.Equal(t, "2024-08-11", summary.LastDate)       // Should be Aug 11, not Aug 13
	assert.Equal(t, 1, summary.TradingDays)               // Only one actual trading day
	assert.Equal(t, []float64{1.500}, summary.Last10Days) // Only one trading price

	t.Logf("Summary: Ticker=%s, LastDate=%s, TradingDays=%d, LastPrice=%.3f",
		summary.Ticker, summary.LastDate, summary.TradingDays, summary.LastPrice)
}
