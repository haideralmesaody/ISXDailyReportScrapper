package dataprocessing

import (
	"context"
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

func TestIntegrationExample_GenerateTickerSummaryFromCombinedCSV(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	
	// Create test CSV file that simulates the BASH scenario
	// (shows Aug 13 as last date but Aug 11 as last trading date)
	csvContent := `Symbol,CompanyName,Date,OpenPrice,HighPrice,LowPrice,ClosePrice,Volume,NumTrades,TradingStatus
BASH,Bank of Baghdad,2024-08-11,1.480,1.520,1.480,1.500,1000,10,true
BASH,Bank of Baghdad,2024-08-12,1.500,1.500,1.500,1.500,0,0,false
BASH,Bank of Baghdad,2024-08-13,1.500,1.500,1.500,1.500,0,0,false
TAQA,National Company for Tourism Investments,2024-08-11,11.900,12.100,11.800,12.000,500,5,true
TAQA,National Company for Tourism Investments,2024-08-12,12.000,12.000,12.000,12.000,200,2,true
TAQA,National Company for Tourism Investments,2024-08-13,12.000,12.000,12.000,12.000,0,0,false`

	combinedFile := filepath.Join(tempDir, "isx_combined_data.csv")
	err := os.WriteFile(combinedFile, []byte(csvContent), 0644)
	require.NoError(t, err)

	// Test the integration
	integrationExample := NewIntegrationExample(slog.Default())
	err = integrationExample.GenerateTickerSummaryFromCombinedCSV(ctx, combinedFile, tempDir)
	require.NoError(t, err)

	// Verify CSV output exists and has correct content
	csvPath := filepath.Join(tempDir, "summary", "ticker", "ticker_summary.csv")
	require.FileExists(t, csvPath)

	csvContent_output, err := os.ReadFile(csvPath)
	require.NoError(t, err)

	lines := strings.Split(string(csvContent_output), "\n")
	require.GreaterOrEqual(t, len(lines), 3) // Header + 2 tickers

	// Verify header
	header := strings.Split(lines[0], ",")
	expectedHeaders := []string{"Ticker", "CompanyName", "LastPrice", "LastDate", "TradingDays", "Last10Days", "TotalVolume", "TotalValue", "AveragePrice", "HighestPrice", "LowestPrice", "Change", "ChangePercent", "LastTradingStatus"}
	assert.Equal(t, expectedHeaders, header)

	// Verify BASH data (should show Aug 11 as LastDate, not Aug 13)
	bashFound := false
	taqaFound := false
	
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		
		fields := strings.Split(lines[i], ",")
		if len(fields) < 6 {
			continue
		}
		
		ticker := fields[0]
		lastDate := fields[3]
		tradingDays := fields[4]
		
		switch ticker {
		case "BASH":
			bashFound = true
			assert.Equal(t, "2024-08-11", lastDate, "BASH LastDate should be Aug 11 (last trading day), not Aug 13")
			assert.Equal(t, "1", tradingDays, "BASH should have 1 trading day")
		case "TAQA":
			taqaFound = true
			assert.Equal(t, "2024-08-12", lastDate, "TAQA LastDate should be Aug 12 (last trading day)")
			assert.Equal(t, "2", tradingDays, "TAQA should have 2 trading days")
		}
	}
	
	assert.True(t, bashFound, "BASH ticker should be found in output")
	assert.True(t, taqaFound, "TAQA ticker should be found in output")

	// Verify JSON output exists
	jsonPath := filepath.Join(tempDir, "summary", "ticker", "ticker_summary.json")
	assert.FileExists(t, jsonPath)

	t.Logf("✅ SSOT correctly identified BASH last trading date as 2024-08-11 (not 2024-08-13)")
	t.Logf("✅ CSV output: %s", csvPath)
	t.Logf("✅ JSON output: %s", jsonPath)
}

func TestIntegrationExample_GenerateTickerSummaryFromTradeRecords(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	
	// Create test records that demonstrate the SSOT logic
	records := []domain.TradeRecord{
		// BASH - Last trading on Aug 11, then forward-filled
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
			ClosePrice:    1.500, // Same price (forward-filled)
			Volume:        0,
			NumTrades:     0,
			TradingStatus: false,
		},
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 13, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500, // Same price (forward-filled)
			Volume:        0,
			NumTrades:     0,
			TradingStatus: false,
		},
		// ZAHRA - Volume fallback scenario
		{
			CompanySymbol: "ZAHRA",
			CompanyName:   "Al-Zahra Company for Financial Investment",
			Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
			ClosePrice:    3.200,
			Volume:        500,    // Has volume but TradingStatus false
			NumTrades:     3,
			TradingStatus: false,  // TradingStatus field not available/reliable
		},
		{
			CompanySymbol: "ZAHRA",
			CompanyName:   "Al-Zahra Company for Financial Investment",
			Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
			ClosePrice:    3.200,
			Volume:        0,
			NumTrades:     0,
			TradingStatus: false,
		},
	}

	integrationExample := NewIntegrationExample(slog.Default())
	outputPath := filepath.Join(tempDir, "trade_records_summary.csv")
	
	err := integrationExample.GenerateTickerSummaryFromTradeRecords(ctx, records, outputPath)
	require.NoError(t, err)

	// Verify output
	require.FileExists(t, outputPath)

	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")
	require.GreaterOrEqual(t, len(lines), 3) // Header + 2 tickers

	// Parse and verify results
	bashFound := false
	zahraFound := false
	
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		
		fields := strings.Split(lines[i], ",")
		if len(fields) < 6 {
			continue
		}
		
		ticker := fields[0]
		lastDate := fields[3]
		tradingDays := fields[4]
		
		switch ticker {
		case "BASH":
			bashFound = true
			assert.Equal(t, "2024-08-11", lastDate, "BASH LastDate should be Aug 11 (TradingStatus=true)")
			assert.Equal(t, "1", tradingDays, "BASH should have 1 trading day")
		case "ZAHRA":
			zahraFound = true
			assert.Equal(t, "2024-08-11", lastDate, "ZAHRA LastDate should be Aug 11 (volume fallback)")
			assert.Equal(t, "1", tradingDays, "ZAHRA should have 1 trading day")
		}
	}
	
	assert.True(t, bashFound, "BASH ticker should be found")
	assert.True(t, zahraFound, "ZAHRA ticker should be found")

	t.Logf("✅ SSOT correctly handled TradingStatus=true for BASH")
	t.Logf("✅ SSOT correctly used volume fallback for ZAHRA")
}

func TestIntegrationExample_ReadCombinedCSV(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test CSV with BOM and various formats
	csvWithBOM := "\uFEFF" + `Symbol,CompanyName,Date,ClosePrice,Volume,NumTrades,TradingStatus
BASH,Bank of Baghdad,2024-08-11,1.500,1000,10,true
TAQA,National Company,2024-08-12,12.000,500,5,false`

	csvFile := filepath.Join(tempDir, "test_combined.csv")
	err := os.WriteFile(csvFile, []byte(csvWithBOM), 0644)
	require.NoError(t, err)

	integrationExample := NewIntegrationExample(slog.Default())
	records, err := integrationExample.readCombinedCSV(csvFile)
	require.NoError(t, err)
	
	assert.Len(t, records, 2)
	
	// Verify BASH record
	bash := records[0]
	assert.Equal(t, "BASH", bash.CompanySymbol)
	assert.Equal(t, "Bank of Baghdad", bash.CompanyName)
	assert.Equal(t, 1.500, bash.ClosePrice)
	assert.Equal(t, int64(1000), bash.Volume)
	assert.Equal(t, int64(10), bash.NumTrades)
	assert.True(t, bash.TradingStatus)
	
	// Verify TAQA record
	taqa := records[1]
	assert.Equal(t, "TAQA", taqa.CompanySymbol)
	assert.Equal(t, "National Company", taqa.CompanyName)
	assert.Equal(t, 12.000, taqa.ClosePrice)
	assert.Equal(t, int64(500), taqa.Volume)
	assert.Equal(t, int64(5), taqa.NumTrades)
	assert.False(t, taqa.TradingStatus)

	t.Logf("✅ Successfully parsed CSV with BOM and converted to TradeRecord structs")
}

func TestIntegrationExample_ParseDate(t *testing.T) {
	integrationExample := NewIntegrationExample(slog.Default())
	
	tests := []struct {
		name     string
		input    string
		wantYear int
		wantOK   bool
	}{
		{"ISO format", "2024-08-11", 2024, true},
		{"US format", "08/11/2024", 2024, true},
		{"EU format", "11/08/2024", 2024, true},
		{"Slash with year first", "2024/08/11", 2024, true},
		{"Dash EU format", "11-08-2024", 2024, true},
		{"Dash US format", "08-11-2024", 2024, true},
		{"Invalid format", "Aug 11, 2024", 0, false},
		{"Empty string", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := integrationExample.parseDate(tt.input)
			
			if tt.wantOK {
				require.NoError(t, err)
				assert.Equal(t, tt.wantYear, date.Year())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestIntegrationExample_SafeParsing(t *testing.T) {
	integrationExample := NewIntegrationExample(slog.Default())
	
	// Test safe float parsing
	testRow := []string{"BASH", "Bank", "1.500", "-", "N/A", "", "100"}
	
	assert.Equal(t, 1.500, integrationExample.parseFloat(testRow, 2))  // Valid float
	assert.Equal(t, 0.0, integrationExample.parseFloat(testRow, 3))    // Dash
	assert.Equal(t, 0.0, integrationExample.parseFloat(testRow, 4))    // N/A
	assert.Equal(t, 0.0, integrationExample.parseFloat(testRow, 5))    // Empty
	assert.Equal(t, 0.0, integrationExample.parseFloat(testRow, 10))   // Out of bounds
	
	// Test safe int parsing
	assert.Equal(t, int64(100), integrationExample.parseInt(testRow, 6)) // Valid int
	assert.Equal(t, int64(0), integrationExample.parseInt(testRow, 3))   // Dash
	assert.Equal(t, int64(0), integrationExample.parseInt(testRow, 10))  // Out of bounds
	
	// Test safe bool parsing
	boolRow := []string{"true", "1", "yes", "false", "0", "no", "active", "inactive"}
	assert.True(t, integrationExample.parseBool(boolRow, 0))   // "true"
	assert.True(t, integrationExample.parseBool(boolRow, 1))   // "1"
	assert.True(t, integrationExample.parseBool(boolRow, 2))   // "yes"
	assert.False(t, integrationExample.parseBool(boolRow, 3))  // "false"
	assert.False(t, integrationExample.parseBool(boolRow, 4))  // "0"
	assert.False(t, integrationExample.parseBool(boolRow, 5))  // "no"
	assert.True(t, integrationExample.parseBool(boolRow, 6))   // "active"
	assert.False(t, integrationExample.parseBool(boolRow, 7))  // "inactive"
	assert.False(t, integrationExample.parseBool(boolRow, 10)) // Out of bounds
}

func TestMigrationGuide(t *testing.T) {
	guide := MigrationGuide()
	
	// Verify the guide contains key information
	assert.Contains(t, guide, "MIGRATION GUIDE")
	assert.Contains(t, guide, "cmd/processor/main.go")
	assert.Contains(t, guide, "internal/dataprocessing/analytics.go")
	assert.Contains(t, guide, "internal/exporter/ticker.go")
	assert.Contains(t, guide, "Summarizer.GenerateFromRecords()")
	assert.Contains(t, guide, "TradingStatus field")
	assert.Contains(t, guide, "EXAMPLE USAGE")
	
	t.Logf("Migration guide length: %d characters", len(guide))
	t.Logf("✅ Migration guide provides comprehensive replacement instructions")
}

// TestRealWorldComparison demonstrates that the SSOT implementation
// produces correct results for the BASH scenario mentioned in the task.
func TestRealWorldComparison(t *testing.T) {
	ctx := context.Background()
	
	// Simulate the real BASH scenario:
	// - Last actual trading was Aug 11
	// - Aug 12 and 13 are forward-filled (same price, no volume/trades)
	// - Old implementation would show LastDate as Aug 13 (wrong)
	// - SSOT implementation should show LastDate as Aug 11 (correct)
	
	records := []domain.TradeRecord{
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500,
			HighPrice:     1.520,
			LowPrice:      1.480,
			Volume:        1000,
			NumTrades:     10,
			Value:         1500.0,
			TradingStatus: true,
		},
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 12, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500, // Same as previous (forward-filled)
			HighPrice:     1.500,
			LowPrice:      1.500,
			Volume:        0,     // No trading
			NumTrades:     0,     // No trading
			Value:         0.0,
			TradingStatus: false, // Not a trading day
		},
		{
			CompanySymbol: "BASH",
			CompanyName:   "Bank of Baghdad",
			Date:          time.Date(2024, 8, 13, 0, 0, 0, 0, time.UTC),
			ClosePrice:    1.500, // Same as previous (forward-filled)
			HighPrice:     1.500,
			LowPrice:      1.500,
			Volume:        0,     // No trading
			NumTrades:     0,     // No trading
			Value:         0.0,
			TradingStatus: false, // Not a trading day
		},
	}
	
	// Test with SSOT implementation
	config := DefaultSummarizerConfig()
	summarizer := NewSummarizer(slog.Default(), config)
	
	summaries, err := summarizer.GenerateFromRecords(ctx, records)
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	
	bashSummary := summaries[0]
	
	// Verify correct results
	assert.Equal(t, "BASH", bashSummary.Ticker)
	assert.Equal(t, "Bank of Baghdad", bashSummary.CompanyName)
	assert.Equal(t, 1.500, bashSummary.LastPrice)
	assert.Equal(t, "2024-08-11", bashSummary.LastDate) // ✅ Correct: Aug 11, not Aug 13
	assert.Equal(t, 1, bashSummary.TradingDays)         // ✅ Correct: Only 1 trading day
	assert.Equal(t, []float64{1.500}, bashSummary.Last10Days) // ✅ Correct: Only trading prices
	
	t.Logf("🎯 VERIFICATION COMPLETE:")
	t.Logf("   ✅ LastDate: %s (correctly Aug 11, not Aug 13)", bashSummary.LastDate)
	t.Logf("   ✅ TradingDays: %d (correctly 1, not 3)", bashSummary.TradingDays)
	t.Logf("   ✅ Last10Days: %v (correctly contains only trading prices)", bashSummary.Last10Days)
	t.Logf("")
	t.Logf("🔧 PROBLEM SOLVED:")
	t.Logf("   - Old implementations used last chronological date (Aug 13)")
	t.Logf("   - SSOT implementation uses last TRADING date (Aug 11)")
	t.Logf("   - TradingStatus field properly checked with volume/trades fallback")
}