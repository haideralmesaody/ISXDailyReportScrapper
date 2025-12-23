# Domain Models

Core business domain models for the ISX Daily Reports Scrapper.

## Models

- **Report** - Daily report data structure
- **Ticker** - Ticker information and metadata
- **TickerSummary** - **SSOT** for ticker summary data across the system
- **Company** - Company details and relationships
- **Operations** - Data processing operations and steps
- **operation** - operation execution state and progress (legacy, use Operations)
- **Analytics** - Analytics and summary data structures
- **License** - License and user authentication models
- **Processor** - Data processing contracts
- **Trade** - Trading activity and statistics

## Guidelines

1. All models must include appropriate tags:
   - `json` - JSON serialization
   - `db` - Database column mapping
   - `validate` - Validation rules

2. Example:

```go
type Report struct {
    ID        int       `json:"id" db:"id" validate:"required,min=1"`
    Date      time.Time `json:"date" db:"date" validate:"required"`
    TickerID  int       `json:"ticker_id" db:"ticker_id" validate:"required"`
    Content   string    `json:"content" db:"content" validate:"required"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

3. Include TableName() methods for GORM compatibility when needed
4. Document all fields with meaningful comments
5. Include validation methods for complex business rules
6. Use value objects for domain concepts
7. Maintain backward compatibility when evolving models

## Single Source of Truth (SSOT)

### TickerSummary
The `TickerSummary` model is the **Single Source of Truth** for all ticker summary data across the ISX Daily Reports Scrapper system. This includes:

- **Authoritative Definition**: All components must use `domain.TickerSummary`
- **Validation**: Built-in validation with `ValidateTickerSummary()`
- **CSV/JSON Support**: Native CSV and JSON serialization support
- **Constructor**: Safe initialization with `NewTickerSummary()`
- **Migration Guide**: See `TICKER_SUMMARY_MIGRATION.md` for migration instructions

**Key Features:**
- Comprehensive field validation with business rules
- CSV round-trip support for data export/import
- Extended metrics for enhanced analysis
- Filtering and querying support
- Version tracking for backward compatibility

**Usage:**
```go
// Create a new summary
summary, err := domain.NewTickerSummary("BBOB", "Bank of Baghdad", 1.250, "2024-01-15", 120)
if err != nil {
    return fmt.Errorf("create summary: %w", err)
}

// Validate existing summary
if err := domain.ValidateTickerSummary(summary); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

// CSV operations
csvString := summary.FormatLast10DaysForCSV()
err = summary.ParseLast10DaysFromCSV("1.200,1.210,1.220")
```

## Change Log
- 2025-08-15: **Added TickerSummary SSOT contract** - Authoritative ticker summary definition
- 2025-07-30: Added Operations model for data processing operations (Phase 4)
- 2025-07-30: Added Processor and Trade models for enhanced domain modeling
- 2025-07-30: Updated documentation with expanded guidelines and best practices