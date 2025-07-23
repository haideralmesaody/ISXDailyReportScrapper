# Data Validation Checklist

This checklist helps ensure data consistency across the ISX Daily Reports Scrapper application. Use this when implementing new features or modifying existing data structures.

## Before Making Changes

- [ ] Review `DATA_SPECIFICATIONS.md` for the relevant data structure
- [ ] Identify all producers of the data (which components generate it)
- [ ] Identify all consumers of the data (which components use it)
- [ ] Check if the change affects CSV headers, JSON field names, or data types

## When Generating Data

### CSV Files
- [ ] Headers match exactly as specified (case-sensitive)
- [ ] Column order matches specification
- [ ] Float values use correct decimal places (%.3f for prices, %.2f for percentages)
- [ ] Date format is YYYY-MM-DD
- [ ] Boolean values are lowercase "true" or "false"
- [ ] File includes UTF-8 BOM for Excel compatibility
- [ ] No trailing commas or extra columns

### JSON Files
- [ ] Field names use snake_case (e.g., `last_price` not `lastPrice`)
- [ ] Arrays are never null (use empty array `[]` instead)
- [ ] Dates use YYYY-MM-DD format
- [ ] Timestamps use ISO 8601 format
- [ ] Numbers don't have unnecessary decimal places
- [ ] Root structure includes metadata fields (`count`, `generated_at`, etc.)

### Ticker Summary Specific
- [ ] `trading_days` counts only actual trading days (volume > 0 or numTrades > 0)
- [ ] `last_10_days` contains only actual trading prices, not forward-filled
- [ ] `last_10_days` is in chronological order (oldest to newest)
- [ ] Empty tickers have `last_10_days` as empty array, not null

## When Consuming Data

### JavaScript/Frontend
- [ ] Use snake_case field names when accessing JSON data
- [ ] Check for null/undefined before calling methods like `.toFixed()`
- [ ] Handle empty arrays gracefully
- [ ] Parse dates consistently

### Go/Backend
- [ ] CSV reader handles BOM correctly
- [ ] Column mapping uses normalized names (lowercase, trimmed)
- [ ] Type conversions have error handling
- [ ] Forward-fill logic checks TradingStatus field

## Testing Checklist

- [ ] Generate sample data and verify format matches specification
- [ ] Test with edge cases (empty data, single record, missing fields)
- [ ] Verify Excel can open CSV files correctly
- [ ] Check JSON is valid and parseable
- [ ] Test frontend displays data without errors
- [ ] Verify calculations (trading days, percentages) are correct

## Common Issues to Avoid

1. **Mismatched field names**: JavaScript expects snake_case, Go uses PascalCase internally
2. **Null vs empty array**: Always use empty arrays in JSON
3. **Decimal precision**: Prices need 3 decimals, percentages need 2
4. **Date formats**: Consistent YYYY-MM-DD format everywhere
5. **Trading vs calendar days**: Only count actual trading activity
6. **Column order**: CSV readers may depend on exact column positions

## Update Process

When changing data structures:

1. Update `DATA_SPECIFICATIONS.md` first
2. Update all producers to generate new format
3. Update all consumers to handle new format
4. Test end-to-end data flow
5. Document the change in version history

## Quick Reference

### CSV Headers
```
Date,CompanyName,Symbol,OpenPrice,HighPrice,LowPrice,AveragePrice,PrevAveragePrice,ClosePrice,PrevClosePrice,Change,ChangePercent,NumTrades,Volume,Value,TradingStatus
```

### JSON Ticker Structure
```json
{
  "ticker": "SYMBOL",
  "company_name": "Name",
  "last_price": 123.456,
  "last_date": "YYYY-MM-DD",
  "trading_days": 100,
  "last_10_days": [120.5, 121.0, ...]
}
```

### WebSocket Message
```json
{
  "type": "progress|status|output",
  "data": {
    "message": "...",
    "progress": 50
  }
}
```