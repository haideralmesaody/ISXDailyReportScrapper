# Column Name Mapping Guide

This document maps column names from their source (Excel files) through all transformations in the application.

## Source: Excel Daily Reports

Based on the parser code, the Excel files from ISX contain these headers:
- **Company Code** → mapped to "code" internally
- **Company Name** → mapped to "name" internally  
- **Opening Price** → mapped to "open"
- **High Price** → mapped to "high"
- **Low Price** → mapped to "low"
- **Average Price** → mapped to "average"
- **Prev. Average Price** → mapped to "prev_avg"
- **Closing Price** → mapped to "close"
- **Prev. Closing Price** → mapped to "prev_close"
- **Change (%)** → mapped to "change_pct"
- **No. of Trades** → mapped to "num_trades"
- **Traded Volume** → mapped to "volume"
- **Traded Value** → mapped to "value"

## Internal Go Struct (parser.TradeRecord)

```go
type TradeRecord struct {
    CompanyName      string  // From Excel "Company Name"
    CompanySymbol    string  // From Excel "Company Code"
    Date             time.Time
    OpenPrice        float64 // From Excel "Opening Price"
    HighPrice        float64 // From Excel "High Price"
    LowPrice         float64 // From Excel "Low Price"
    AveragePrice     float64 // From Excel "Average Price"
    PrevAveragePrice float64 // From Excel "Prev. Average Price"
    ClosePrice       float64 // From Excel "Closing Price"
    PrevClosePrice   float64 // From Excel "Prev. Closing Price"
    Change           float64 // Calculated field
    ChangePercent    float64 // From Excel "Change (%)"
    NumTrades        int64   // From Excel "No. of Trades"
    Volume           int64   // From Excel "Traded Volume"
    Value            float64 // From Excel "Traded Value"
    TradingStatus    bool    // Calculated field (true if actual trading)
}
```

## CSV Output Headers

All CSV files should use these exact headers (PascalCase, no spaces):

```csv
Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus
```

### Mapping:
- Company Code → Symbol
- Company Name → CompanyName
- No. of Trades → NumTrades
- All price fields remove spaces and use PascalCase

## Ticker Summary CSV

```csv
Ticker,CompanyName,LastPrice,LastDate,TradingDays,Last10Days
```

### Mapping:
- Symbol → Ticker (for summary)
- ClosePrice → LastPrice (most recent)

## JSON API Fields

All JSON responses use snake_case:

```json
{
  "ticker": "SYMBOL",           // From Symbol/CompanySymbol
  "company_name": "Name",       // From CompanyName
  "last_price": 123.45,        // From ClosePrice (most recent)
  "last_date": "2025-07-18",   // From Date (most recent)
  "trading_days": 100,         // Calculated (TradingStatus = true)
  "last_10_days": [...]        // From ClosePrice array
}
```

## Consistency Rules

1. **Excel Import**: Use flexible matching (case-insensitive, partial matches)
2. **Internal Processing**: Use Go struct field names (CompanySymbol, CompanyName)
3. **CSV Export**: Use PascalCase headers exactly as specified
4. **JSON Export**: Convert to snake_case
5. **Web Display**: Use snake_case when accessing JSON fields

## Column Name Transformations

| Excel Header | Parser Map | Go Struct Field | CSV Header | JSON Field |
|--------------|------------|-----------------|------------|------------|
| Company Code | code | CompanySymbol | Symbol | ticker |
| Company Name | name | CompanyName | CompanyName | company_name |
| Opening Price | open | OpenPrice | OpenPrice | open_price |
| High Price | high | HighPrice | HighPrice | high_price |
| Low Price | low | LowPrice | LowPrice | low_price |
| Average Price | average | AveragePrice | AveragePrice | average_price |
| Prev. Average Price | prev_avg | PrevAveragePrice | PrevAveragePrice | prev_average_price |
| Closing Price | close | ClosePrice | ClosePrice | close_price |
| Prev. Closing Price | prev_close | PrevClosePrice | PrevClosePrice | prev_close_price |
| Change (%) | change_pct | ChangePercent | ChangePercent | change_percent |
| No. of Trades | num_trades | NumTrades | NumTrades | num_trades |
| Traded Volume | volume | Volume | Volume | volume |
| Traded Value | value | Value | Value | value |
| (calculated) | - | TradingStatus | TradingStatus | trading_status |

## Implementation Checklist

When reading/writing data:

- [ ] Excel parser: Use flexible, case-insensitive matching
- [ ] CSV headers: Use exact PascalCase format
- [ ] JSON fields: Convert to snake_case
- [ ] Column matching: Handle both PascalCase and snake_case variations
- [ ] BOM handling: Strip UTF-8 BOM when reading CSVs