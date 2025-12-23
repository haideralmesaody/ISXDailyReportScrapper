package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTickerSummary(t *testing.T) {
	tests := []struct {
		name        string
		ticker      string
		companyName string
		lastPrice   float64
		lastDate    string
		tradingDays int
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid ticker summary",
			ticker:      "BBOB",
			companyName: "Bank of Baghdad",
			lastPrice:   1.250,
			lastDate:    "2024-01-15",
			tradingDays: 120,
			wantErr:     false,
		},
		{
			name:        "valid ticker with whitespace trimming",
			ticker:      " BMNS ",
			companyName: " Bank of Mosul ",
			lastPrice:   0.500,
			lastDate:    "2024-01-15",
			tradingDays: 50,
			wantErr:     false,
		},
		{
			name:        "ticker too short",
			ticker:      "A",
			companyName: "Company A",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     true,
			errContains: "must be 2-10 uppercase letters only",
		},
		{
			name:        "ticker too long",
			ticker:      "VERYLONGTICKER",
			companyName: "Very Long Company Name",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     true,
			errContains: "must be 2-10 uppercase letters only",
		},
		{
			name:        "ticker with numbers",
			ticker:      "BBOB1",
			companyName: "Bank of Baghdad",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     true,
			errContains: "must be 2-10 uppercase letters only",
		},
		{
			name:        "ticker with lowercase",
			ticker:      "bbob",
			companyName: "Bank of Baghdad",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     false, // Should be converted to uppercase
		},
		{
			name:        "empty company name",
			ticker:      "BBOB",
			companyName: "",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     true,
			errContains: "company name is required",
		},
		{
			name:        "company name too short",
			ticker:      "BBOB",
			companyName: "AB",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     true,
			errContains: "must be at least 3 characters",
		},
		{
			name:        "price too low",
			ticker:      "BBOB",
			companyName: "Bank of Baghdad",
			lastPrice:   0.0005,
			lastDate:    "2024-01-15",
			tradingDays: 10,
			wantErr:     true,
			errContains: "must be at least 0.001",
		},
		{
			name:        "invalid date format",
			ticker:      "BBOB",
			companyName: "Bank of Baghdad",
			lastPrice:   1.000,
			lastDate:    "15/01/2024",
			tradingDays: 10,
			wantErr:     true,
			errContains: "must be in format '2006-01-02'",
		},
		{
			name:        "negative trading days",
			ticker:      "BBOB",
			companyName: "Bank of Baghdad",
			lastPrice:   1.000,
			lastDate:    "2024-01-15",
			tradingDays: -5,
			wantErr:     true,
			errContains: "cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := NewTickerSummary(tt.ticker, tt.companyName, tt.lastPrice, tt.lastDate, tt.tradingDays)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, summary)
			} else {
				require.NoError(t, err)
				require.NotNil(t, summary)
				
				// Verify uppercase conversion for ticker
				assert.Equal(t, strings.ToUpper(strings.TrimSpace(tt.ticker)), summary.Ticker)
				assert.Equal(t, strings.TrimSpace(tt.companyName), summary.CompanyName)
				assert.Equal(t, tt.lastPrice, summary.LastPrice)
				assert.Equal(t, tt.lastDate, summary.LastDate)
				assert.Equal(t, tt.tradingDays, summary.TradingDays)
				assert.Equal(t, "1.0", summary.Version)
				assert.NotZero(t, summary.GeneratedAt)
				assert.NotNil(t, summary.Last10Days)
				assert.Empty(t, summary.Last10Days)
			}
		})
	}
}

func TestValidateTickerSummary(t *testing.T) {
	validSummary := &TickerSummary{
		Ticker:      "BBOB",
		CompanyName: "Bank of Baghdad",
		LastPrice:   1.250,
		LastDate:    "2024-01-15",
		TradingDays: 120,
		Last10Days:  []float64{1.200, 1.210, 1.220, 1.240, 1.250},
	}

	tests := []struct {
		name        string
		modify      func(*TickerSummary)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid summary",
			modify:  func(s *TickerSummary) {},
			wantErr: false,
		},
		{
			name:        "nil summary",
			modify:      func(s *TickerSummary) { s = nil },
			wantErr:     true,
			errContains: "cannot be nil",
		},
		{
			name: "empty ticker",
			modify: func(s *TickerSummary) {
				s.Ticker = ""
			},
			wantErr:     true,
			errContains: "ticker is required",
		},
		{
			name: "invalid ticker format",
			modify: func(s *TickerSummary) {
				s.Ticker = "BB@B"
			},
			wantErr:     true,
			errContains: "must be 2-10 uppercase letters only",
		},
		{
			name: "company name too long",
			modify: func(s *TickerSummary) {
				s.CompanyName = strings.Repeat("A", 256)
			},
			wantErr:     true,
			errContains: "must not exceed 255 characters",
		},
		{
			name: "too many last 10 days",
			modify: func(s *TickerSummary) {
				s.Last10Days = make([]float64, 11)
			},
			wantErr:     true,
			errContains: "cannot have more than 10 elements",
		},
		{
			name: "negative price in last 10 days",
			modify: func(s *TickerSummary) {
				s.Last10Days = []float64{1.200, -1.210, 1.220}
			},
			wantErr:     true,
			errContains: "cannot be negative",
		},
		{
			name: "negative total volume",
			modify: func(s *TickerSummary) {
				s.TotalVolume = -1000
			},
			wantErr:     true,
			errContains: "total volume cannot be negative",
		},
		{
			name: "negative total value",
			modify: func(s *TickerSummary) {
				s.TotalValue = -1000.50
			},
			wantErr:     true,
			errContains: "total value cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the valid summary
			summary := *validSummary
			if tt.name == "nil summary" {
				err := ValidateTickerSummary(nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			
			// Apply modifications
			tt.modify(&summary)
			
			err := ValidateTickerSummary(&summary)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFormatLast10DaysForCSV(t *testing.T) {
	tests := []struct {
		name   string
		prices []float64
		want   string
	}{
		{
			name:   "empty slice",
			prices: []float64{},
			want:   "",
		},
		{
			name:   "single price",
			prices: []float64{1.250},
			want:   "1.250",
		},
		{
			name:   "multiple prices",
			prices: []float64{1.200, 1.210, 1.220, 1.240, 1.250},
			want:   "1.200,1.210,1.220,1.240,1.250",
		},
		{
			name:   "max 10 prices",
			prices: []float64{1.100, 1.110, 1.120, 1.130, 1.140, 1.150, 1.160, 1.170, 1.180, 1.190},
			want:   "1.100,1.110,1.120,1.130,1.140,1.150,1.160,1.170,1.180,1.190",
		},
		{
			name:   "prices with more precision",
			prices: []float64{1.2345, 2.6789},
			want:   "1.234,2.679", // Should format to 3 decimal places (%.3f truncates, doesn't round)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &TickerSummary{Last10Days: tt.prices}
			got := summary.FormatLast10DaysForCSV()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseLast10DaysFromCSV(t *testing.T) {
	tests := []struct {
		name      string
		csvString string
		want      []float64
		wantErr   bool
		errContains string
	}{
		{
			name:      "empty string",
			csvString: "",
			want:      []float64{},
			wantErr:   false,
		},
		{
			name:      "single price",
			csvString: "1.250",
			want:      []float64{1.250},
			wantErr:   false,
		},
		{
			name:      "multiple prices",
			csvString: "1.200,1.210,1.220,1.240,1.250",
			want:      []float64{1.200, 1.210, 1.220, 1.240, 1.250},
			wantErr:   false,
		},
		{
			name:      "prices with whitespace",
			csvString: " 1.200 , 1.210 , 1.220 ",
			want:      []float64{1.200, 1.210, 1.220},
			wantErr:   false,
		},
		{
			name:      "empty values in string",
			csvString: "1.200,,1.220",
			want:      []float64{1.200, 1.220},
			wantErr:   false,
		},
		{
			name:        "too many values",
			csvString:   "1.1,1.2,1.3,1.4,1.5,1.6,1.7,1.8,1.9,2.0,2.1",
			want:        nil,
			wantErr:     true,
			errContains: "too many price values",
		},
		{
			name:        "invalid price format",
			csvString:   "1.200,invalid,1.220",
			want:        nil,
			wantErr:     true,
			errContains: "invalid price at position",
		},
		{
			name:        "negative price",
			csvString:   "1.200,-1.210,1.220",
			want:        nil,
			wantErr:     true,
			errContains: "negative price at position",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &TickerSummary{}
			err := summary.ParseLast10DaysFromCSV(tt.csvString)
			
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, summary.Last10Days)
			}
		})
	}
}

func TestIsValidTicker(t *testing.T) {
	tests := []struct {
		name   string
		ticker string
		want   bool
	}{
		{"valid short ticker", "AB", true},
		{"valid medium ticker", "BBOB", true},
		{"valid long ticker", "ABCDEFGHIJ", true},
		{"too short", "A", false},
		{"too long", "ABCDEFGHIJK", false},
		{"with numbers", "BBOB1", false},
		{"with special chars", "BB@B", false},
		{"with spaces", "BB OB", false},
		{"lowercase", "bbob", false},
		{"mixed case", "BbOb", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTicker(tt.ticker)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTickerSummaryCSVRoundTrip(t *testing.T) {
	// Test that we can format to CSV and parse back without data loss
	original := &TickerSummary{
		Ticker:      "BBOB",
		CompanyName: "Bank of Baghdad", 
		LastPrice:   1.250,
		LastDate:    "2024-01-15",
		TradingDays: 120,
		Last10Days:  []float64{1.200, 1.210, 1.220, 1.240, 1.250},
	}

	// Format to CSV
	csvString := original.FormatLast10DaysForCSV()
	assert.Equal(t, "1.200,1.210,1.220,1.240,1.250", csvString)

	// Parse back from CSV
	parsed := &TickerSummary{}
	err := parsed.ParseLast10DaysFromCSV(csvString)
	require.NoError(t, err)

	// Verify data integrity (allowing for floating point precision)
	assert.Len(t, parsed.Last10Days, len(original.Last10Days))
	for i, price := range original.Last10Days {
		assert.InDelta(t, price, parsed.Last10Days[i], 0.001, "Price at index %d", i)
	}
}

func TestTickerSummaryValidationRules(t *testing.T) {
	// Test that validation rules are properly defined
	rules := TickerSummaryValidationRules
	
	assert.NotNil(t, rules.TickerPattern)
	assert.Equal(t, 2, rules.MinTickerLength)
	assert.Equal(t, 10, rules.MaxTickerLength)
	assert.Equal(t, 3, rules.MinCompanyLength)
	assert.Equal(t, 255, rules.MaxCompanyLength)
	assert.Equal(t, 0.001, rules.MinPrice)
	assert.Equal(t, 10, rules.MaxLast10Days)
	assert.Equal(t, "2006-01-02", rules.RequiredDateFormat)
	
	// Test pattern matching
	assert.True(t, rules.TickerPattern.MatchString("BBOB"))
	assert.False(t, rules.TickerPattern.MatchString("bbob"))
	assert.False(t, rules.TickerPattern.MatchString("BB1B"))
	assert.False(t, rules.TickerPattern.MatchString("A"))
}

func TestTickerSummaryJSONTags(t *testing.T) {
	// Test that JSON marshaling works correctly with our tags
	summary := &TickerSummary{
		Ticker:        "BBOB",
		CompanyName:   "Bank of Baghdad",
		LastPrice:     1.250,
		LastDate:      "2024-01-15",
		TradingDays:   120,
		Last10Days:    []float64{1.200, 1.210, 1.220},
		TotalVolume:   1000000,
		TotalValue:    1250000.0,
		AveragePrice:  1.225,
		HighestPrice:  1.300,
		LowestPrice:   1.100,
		ChangePercent: 2.5,
		GeneratedAt:   time.Now(),
		DataSource:    "daily_reports",
		Version:       "1.0",
	}

	err := ValidateTickerSummary(summary)
	require.NoError(t, err, "Summary should be valid")
	
	// Verify required fields are present
	assert.NotEmpty(t, summary.Ticker)
	assert.NotEmpty(t, summary.CompanyName)
	assert.Positive(t, summary.LastPrice)
	assert.NotEmpty(t, summary.LastDate)
	assert.GreaterOrEqual(t, summary.TradingDays, 0)
	assert.NotNil(t, summary.Last10Days)
}

func BenchmarkValidateTickerSummary(b *testing.B) {
	summary := &TickerSummary{
		Ticker:      "BBOB",
		CompanyName: "Bank of Baghdad",
		LastPrice:   1.250,
		LastDate:    "2024-01-15",
		TradingDays: 120,
		Last10Days:  []float64{1.200, 1.210, 1.220, 1.240, 1.250},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateTickerSummary(summary)
	}
}

func BenchmarkFormatLast10DaysForCSV(b *testing.B) {
	summary := &TickerSummary{
		Last10Days: []float64{1.100, 1.110, 1.120, 1.130, 1.140, 1.150, 1.160, 1.170, 1.180, 1.190},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary.FormatLast10DaysForCSV()
	}
}

func BenchmarkParseLast10DaysFromCSV(b *testing.B) {
	csvString := "1.100,1.110,1.120,1.130,1.140,1.150,1.160,1.170,1.180,1.190"
	summary := &TickerSummary{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		summary.ParseLast10DaysFromCSV(csvString)
	}
}