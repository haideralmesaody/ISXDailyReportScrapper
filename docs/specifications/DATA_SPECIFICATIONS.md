# ISX Daily Reports Scrapper - Data Specifications

This document defines the data structures and file formats used throughout the ISX Daily Reports Scrapper application. All components generating or consuming data must adhere to these specifications to ensure compatibility.

## Table of Contents

1. [CSV File Formats](#csv-file-formats)
2. [JSON API Responses](#json-api-responses)
3. [WebSocket Messages](#websocket-messages)
4. [Internal Data Structures](#internal-data-structures)

---

## CSV File Formats

### 1. Daily Report CSV (`isx_daily_YYYY_MM_DD.csv`)

**Location**: `data/reports/isx_daily_YYYY_MM_DD.csv`

**Headers** (exact order):
```csv
Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus
```

**Important**: These headers are derived from Excel columns:
- `Symbol` comes from Excel's "Company Code" column
- `CompanyName` comes from Excel's "Company Name" column
- `NumTrades` comes from Excel's "No. of Trades" column
- All other fields follow PascalCase naming without spaces

**Field Specifications**:
| Field | Type | Format | Description |
|-------|------|--------|-------------|
| Date | String | YYYY-MM-DD | Trading date |
| CompanyName | String | Text | Full company name |
| Symbol | String | Text | Ticker symbol (uppercase) |
| OpenPrice | Float | %.3f | Opening price (3 decimal places) |
| HighPrice | Float | %.3f | Daily high price |
| LowPrice | Float | %.3f | Daily low price |
| AveragePrice | Float | %.3f | Average trading price |
| PrevAveragePrice | Float | %.3f | Previous day's average price |
| ClosePrice | Float | %.3f | Closing price |
| PrevClosePrice | Float | %.3f | Previous day's closing price |
| Change | Float | %.3f | Price change amount |
| ChangePercent | Float | %.2f | Percentage change (2 decimal places) |
| NumTrades | Integer | %d | Number of trades |
| Volume | Integer | %d | Trading volume |
| Value | Float | %.2f | Total traded value |
| TradingStatus | Boolean | true/false | true = actual trading, false = forward-filled |

### 2. Combined Data CSV (`isx_combined_data.csv`)

**Location**: `data/reports/isx_combined_data.csv`

**Headers**: Same as Daily Report CSV

**Description**: Contains all historical data with forward-filled records for non-trading days.

### 3. Ticker Trading History CSV (`{SYMBOL}_trading_history.csv`)

**Location**: `data/reports/{SYMBOL}_trading_history.csv`

**Headers**: Same as Daily Report CSV

**Description**: Individual ticker's complete trading history including forward-filled data.

### 4. Ticker Summary CSV (`ticker_summary.csv`)

**Location**: `data/reports/ticker_summary.csv`

**Headers**:
```csv
Ticker,CompanyName,LastPrice,LastDate,TradingDays,Last10Days
```

**Field Specifications**:
| Field | Type | Format | Description |
|-------|------|--------|-------------|
| Ticker | String | Text | Ticker symbol |
| CompanyName | String | Text | Full company name |
| LastPrice | Float | %.3f | Most recent closing price |
| LastDate | String | YYYY-MM-DD | Date of last price |
| TradingDays | Integer | %d | Count of actual trading days (not forward-filled) |
| Last10Days | String | %.3f,%.3f,... | Comma-separated last 10 actual trading prices |

### 5. Market Indices CSV (`indexes.csv`)

**Location**: `data/reports/indexes.csv`

**Headers**:
```csv
Date,ISX60,ISX15
```

**Field Specifications**:
| Field | Type | Format | Description |
|-------|------|--------|-------------|
| Date | String | YYYY-MM-DD | Index date |
| ISX60 | Float | %.2f | ISX60 index value |
| ISX15 | Float | %.2f | ISX15 index value |

---

## JSON API Responses

### 1. Ticker Summary JSON (`ticker_summary.json`)

**Location**: `data/reports/ticker_summary.json`

**Structure**:
```json
{
  "tickers": [
    {
      "ticker": "SYMBOL",
      "company_name": "Company Name",
      "last_price": 123.456,
      "last_date": "YYYY-MM-DD",
      "trading_days": 100,
      "last_10_days": [120.5, 121.0, 122.3, ...],
      "daily_change_percent": 2.5,
      "weekly_change_percent": 5.2,
      "monthly_change_percent": -1.8,
      "daily_volume": 1500000,
      "daily_value": 185000000.75,
      "previous_close": 120.5,
      "high_52_week": 145.0,
      "low_52_week": 98.5
    }
  ],
  "count": 82,
  "generated_at": "2025-07-18T15:30:00Z"
}
```

**Field Specifications**:
| Field | Type | Description |
|-------|------|-------------|
| tickers | Array | Array of ticker summary objects |
| ticker | String | Ticker symbol (uppercase) |
| company_name | String | Full company name (snake_case) |
| last_price | Float | Most recent closing price |
| last_date | String | Date of last price (YYYY-MM-DD) |
| trading_days | Integer | Count of actual trading days only |
| last_10_days | Array[Float] | Last 10 actual trading prices (never null, empty array if no data) |
| daily_change_percent | Float | Daily percentage change (last vs previous trading day) |
| weekly_change_percent | Float | Weekly percentage change (last vs 7 trading days ago) |
| monthly_change_percent | Float | Monthly percentage change (last vs 30 trading days ago) |
| daily_volume | Integer | Most recent day's trading volume |
| daily_value | Float | Most recent day's trading value |
| previous_close | Float | Previous trading day's closing price |
| high_52_week | Float | Highest price in last 252 trading days |
| low_52_week | Float | Lowest price in last 252 trading days |
| count | Integer | Total number of tickers |
| generated_at | String | ISO 8601 timestamp of generation |

### 2. API Endpoints Response Format

#### `/api/tickers` (GET)
Returns the ticker_summary.json content as-is.

#### `/api/ticker/{symbol}` (GET)
```json
{
  "ticker": "SYMBOL",
  "company_name": "Company Name",
  "last_price": 123.456,
  "last_date": "YYYY-MM-DD",
  "trading_days": 100,
  "last_10_days": [120.5, 121.0, ...],
  "history": [
    {
      "date": "YYYY-MM-DD",
      "open": 120.0,
      "high": 125.0,
      "low": 119.0,
      "close": 123.456,
      "volume": 1000000
    }
  ]
}
```

#### `/api/files` (GET)
```json
{
  "files": [
    {
      "name": "isx_daily_2025_07_18.csv",
      "size": 12345,
      "date": "2025-07-18T10:00:00Z"
    }
  ],
  "count": 150
}
```

#### `/api/status` (GET)
```json
{
  "processing": {
    "active": false,
    "progress": 0,
    "message": "Idle"
  },
  "last_update": "2025-07-18T10:00:00Z",
  "data_status": {
    "total_files": 150,
    "total_tickers": 82,
    "latest_date": "2025-07-18"
  }
}
```

#### `/api/gainers-losers` (GET)
**Query Parameters**:
- `period`: `1d` (default), `1w`, `1m` - Time period for percentage change
- `limit`: `5` (default), max `50` - Number of results per category
- `min_volume`: `0` (default) - Minimum daily volume filter

**Response**:
```json
{
  "top_gainers": [
    {
      "ticker": "SYMBOL",
      "company_name": "Company Name", 
      "current_price": 123.45,
      "change_percent": 5.67,
      "volume": 1500000,
      "value": 185000000.75
    }
  ],
  "top_losers": [
    {
      "ticker": "SYMBOL",
      "company_name": "Company Name",
      "current_price": 98.75,
      "change_percent": -3.21,
      "volume": 750000,
      "value": 74000000.50
    }
  ],
  "most_active": [
    {
      "ticker": "SYMBOL", 
      "company_name": "Company Name",
      "current_price": 156.78,
      "change_percent": 1.23,
      "volume": 5000000,
      "value": 783000000.25
    }
  ],
  "metadata": {
    "period": "1d",
    "limit": 5,
    "min_volume": 0,
    "generated_at": "2025-07-20T10:15:30Z"
  }
}
```

---

## WebSocket Messages

### 1. Progress Update
```json
{
  "type": "progress",
  "data": {
    "stage": "scraping|processing|extracting",
    "progress": 75,
    "message": "Processing file 15 of 20..."
  }
}
```

### 2. Status Update
```json
{
  "type": "status",
  "data": {
    "status": "idle|running|completed|error",
    "message": "Process completed successfully"
  }
}
```

### 3. Output Message
```json
{
  "type": "output",
  "data": {
    "message": "Found 82 tickers",
    "level": "info|warning|error|success"
  }
}
```

---

## Internal Data Structures

### 1. Trade Record (Go struct)
```go
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
    TradingStatus    bool  // true = actual trading, false = forward-filled
}
```

### 2. Ticker Summary (Go struct)
```go
type TickerInfo struct {
    Ticker      string
    CompanyName string
    LastPrice   float64
    LastDate    string
    TradingDays int      // Count of actual trading days only
    Last10Days  []float64 // Last 10 actual trading prices
    
    // Enhanced fields for Market Movers feature
    DailyChangePercent   float64 `json:"daily_change_percent"`
    WeeklyChangePercent  float64 `json:"weekly_change_percent"`
    MonthlyChangePercent float64 `json:"monthly_change_percent"`
    DailyVolume         int64   `json:"daily_volume"`
    DailyValue          float64 `json:"daily_value"`
    PreviousClose       float64 `json:"previous_close"`
    High52Week          float64 `json:"high_52_week"`
    Low52Week           float64 `json:"low_52_week"`
}
```

---

## Important Notes

1. **Date Formats**: All dates use `YYYY-MM-DD` format in CSVs and JSON
2. **Float Precision**: 
   - Prices: 3 decimal places (%.3f)
   - Percentages: 2 decimal places (%.2f)
   - Index values: 2 decimal places (%.2f)
3. **Boolean Values**: Use lowercase `true`/`false` in CSV
4. **Empty Values**: 
   - Numeric fields: Use `0` or `0.0`
   - Arrays: Use empty array `[]` never `null`
   - Strings: Use empty string `""`
5. **File Encoding**: All CSV files must include UTF-8 BOM for Excel compatibility
6. **Field Naming**:
   - CSV headers: PascalCase (e.g., `CompanyName`)
   - JSON fields: snake_case (e.g., `company_name`)
7. **Trading Status**: 
   - `true` = Actual trading day with volume > 0 or numTrades > 0
   - `false` = Forward-filled data for non-trading days

## Version History

- **v1.0** (2025-07-18): Initial specification document