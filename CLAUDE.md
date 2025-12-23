# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Core Principles](#core-principles)
3. [Architecture](#architecture)
4. [EOD-Only Data Processing Patterns](#eod-only-data-processing-patterns)
5. [CSV-Based Data Storage with Caching and Indexing](#csv-based-data-storage-with-caching-and-indexing)
6. [Thread-Safe Service Patterns](#thread-safe-service-patterns)
7. [Build & Development Commands](#build--development-commands)
8. [Key Standards](#key-standards)
9. [Critical Patterns](#critical-patterns)
10. [Error Handling Best Practices](#error-handling-best-practices)
11. [Interface & Dependency Management](#interface--dependency-management)
12. [Testing Requirements](#testing-requirements)
13. [Configuration Management](#configuration-management)
14. [Integration Patterns](#integration-patterns)
15. [Performance Guidelines](#performance-guidelines)
16. [Security Considerations](#security-considerations)
17. [Common Workflows](#common-workflows)
18. [Observability](#observability)
19. [Monitoring & Alerting](#monitoring--alerting)
20. [Documentation Standards](#documentation-standards)
21. [Code Review Checklist](#code-review-checklist)
22. [Version Management](#version-management)
23. [Debugging & Troubleshooting](#debugging--troubleshooting)
24. [Database/Storage Patterns](#databasestorage-patterns)
25. [Deployment Notes](#deployment-notes)
26. [React Hydration Best Practices](#react-hydration-best-practices)
27. [Important Notes](#important-notes)
28. [Project Structure](#project-structure)
29. [Interactive Guide Feature](#interactive-guide-feature)

## Project Overview

ISX Daily Reports Scrapper is a financial data processing system for the Iraqi Stock Exchange (ISX) that provides automated data collection, processing, analysis, and reporting capabilities with enterprise-level security and reliability.

**🚀 Key Features:**
- **Automated Data Pipeline**: 5-stage processing pipeline for reliable data transformation
- **EOD-Only Processing**: Optimized for end-of-day data (no real-time complexity)
- **High Performance**: Efficient data processing and instant access to market data
- **Real-time Updates**: WebSocket integration for live progress monitoring
- **Enterprise Security**: Hardware-locked licensing and encrypted data storage

**Key Technologies:**
- **Backend**: Go 1.21+ with Chi v5 router, slog logging
- **Frontend**: Next.js 14 with TypeScript, Tailwind CSS, Shadcn/ui
- **Data**: CSV/Excel processing, Google Sheets integration
- **Security**: AES-256 encryption, hardware-locked licensing
- **Deployment**: Single binary with embedded frontend

## Core Principles

1. **Security First**: All sensitive data encrypted, license-protected operations
2. **Clean Architecture**: Clear separation of concerns, dependency injection
3. **Type Safety**: Strong typing in both Go and TypeScript
4. **Testability**: TDD approach, minimum 80% coverage
5. **Performance**: Efficient data processing with sub-millisecond response times
6. **Single Source of Truth**: Elimination of data duplication and consistency issues
7. **EOD-First Design**: Optimized for end-of-day data processing patterns
8. **Observability**: Structured logging, metrics, distributed tracing

## Architecture

### Backend Structure (Go)
```
api/
├── cmd/                    # Application entry points
│   ├── web-licensed/       # Main web server
│   ├── scraper/           # ISX data scraper
│   ├── processor/         # Data processor
│   ├── indexcsv/          # Index extractor
│   └── activate-license/  # License activation tool
├── internal/              # Private application code
│   ├── app/              # Application initialization
│   ├── config/           # Configuration management
│   ├── errors/           # Error handling (RFC 7807)
│   ├── license/          # License management
│   ├── middleware/       # HTTP middleware (Chi)
│   ├── operations/       # Pipeline operations
│   ├── security/         # Encryption, auth
│   ├── services/         # Business logic
│   ├── transport/http/   # HTTP handlers
│   └── websocket/        # WebSocket handlers
└── pkg/                  # Public packages
    └── contracts/        # Shared types/interfaces
```

### Frontend Structure (Next.js)
```
web/
├── app/                  # Next.js app router pages
├── components/           # React components
│   ├── ui/              # Shadcn/ui components
│   ├── operations/      # Operation-specific
│   └── layout/          # Layout components
├── lib/                 # Utilities and hooks
├── types/               # TypeScript definitions
└── public/              # Static assets
```

## EOD-Only Data Processing Patterns

**Design Philosophy**: ISX Pulse is optimized for End-of-Day (EOD) data processing, eliminating the complexity and overhead of real-time data handling while providing superior performance and reliability for Iraqi Stock Exchange analysis.

### 🌅 **EOD-First Architecture Benefits**

#### Why EOD-Only Processing?
1. **ISX Market Characteristics**: Iraqi Stock Exchange operates with daily settlements
2. **Performance Focus**: No real-time latency constraints
3. **Data Quality**: Clean, verified daily data vs noisy real-time feeds
4. **Cost Efficiency**: Reduced infrastructure and data feed costs
5. **User Experience**: Predictable, fast load times with cached data

#### EOD Processing Timeline
```
Daily EOD Processing Schedule
┌─────────────────────────────────────────────────────────┐
│  4:00 PM  - Market Close (Iraq time)                    │
│  4:30 PM  - ISX publishes daily reports                 │
│  5:00 PM  - Automated scraping starts                   │
│  5:05 PM  - Data processing and validation              │
│  5:15 PM  - Data transformation and analysis           │
│  5:20 PM  - Data publishing and cache updates          │
│  5:25 PM  - System ready for user access                │
└─────────────────────────────────────────────────────────┘
```

### 🔄 **EOD Pipeline Architecture**

#### Pipeline Stage Dependencies
```go
type EODPipeline struct {
    stages []PipelineStage
    // Execution order ensures data dependencies
}

type PipelineStage interface {
    Name() string
    Dependencies() []string
    Execute(ctx context.Context, input StageInput) (StageOutput, error)
    Rollback(ctx context.Context) error
}

// EOD Pipeline Configuration
var eodPipeline = EODPipeline{
    stages: []PipelineStage{
        &ScrapingStage{},      // 1. Download Excel files
        &ProcessingStage{},    // 2. Convert to CSV
        &IndexStage{},         // 3. Extract ISX60 data
        &LiquidityStage{},     // 4. Calculate liquidity
        &AnalysisStage{},      // 5. Complete analysis
    },
}
```

#### Stage Dependency Management
```go
func (p *EODPipeline) Execute(ctx context.Context) error {
    completed := make(map[string]StageOutput)

    for _, stage := range p.stages {
        // Check dependencies
        for _, dep := range stage.Dependencies() {
            if _, exists := completed[dep]; !exists {
                return fmt.Errorf("dependency %s not completed for stage %s", dep, stage.Name())
            }
        }

        // Execute stage
        input := p.prepareStageInput(stage, completed)
        output, err := stage.Execute(ctx, input)
        if err != nil {
            return fmt.Errorf("stage %s failed: %w", stage.Name(), err)
        }

        completed[stage.Name()] = output
        slog.InfoContext(ctx, "stage completed",
            "stage", stage.Name(),
            "records_processed", output.RecordCount,
            "duration_ms", output.Duration.Milliseconds(),
        )
    }

    return nil
}
```

### 📊 **EOD Data Management Patterns**

#### Daily File Organization
```
data/eod/
├── 2025-10-17/                 # Daily directory
│   ├── raw/                    # Original ISX Excel files
│   │   ├── market_summary_2025-10-17.xlsx
│   │   └── price_list_2025-10-17.xlsx
│   ├── processed/              # Cleaned CSV files
│   │   ├── price_data_2025-10-17.csv
│   │   └── market_data_2025-10-17.csv
│   ├── indices/                # Index data
│   │   └── isx60_index_2025-10-17.csv
│   ├── liquidity/              # Liquidity metrics
│   │   └── liquidity_metrics_2025-10-17.csv
│   └── metadata.json           # Processing metadata
├── 2025-10-16/                 # Previous day
└── latest/                     # Symbolic links to latest
    ├── price_data.csv → ../2025-10-17/processed/price_data_2025-10-17.csv
    └── isx60_index.csv → ../2025-10-17/indices/isx60_index_2025-10-17.csv
```

#### Metadata Tracking
```json
{
  "date": "2025-10-17",
  "processing_started_at": "2025-10-17T17:00:00Z",
  "processing_completed_at": "2025-10-17T17:25:00Z",
  "total_duration_minutes": 25,
  "stages": [
    {
      "name": "scraping",
      "started_at": "2025-10-17T17:00:00Z",
      "completed_at": "2025-10-17T17:03:00Z",
      "status": "success",
      "records_processed": 95,
      "files_downloaded": 2
    },
    {
      "name": "processing",
      "started_at": "2025-10-17T17:05:00Z",
      "completed_at": "2025-10-17T17:15:00Z",
      "status": "success",
      "records_processed": 95,
      "data_quality": "100%"
    }
  ],
  "data_quality": {
    "completeness": "100%",
    "validation_errors": 0,
    "anomalies_detected": 0
  }
}
```

### ⚡ **EOD Performance Optimizations**

#### Batch Processing Patterns
```go
func (s *ProcessingStage) processBatchData(tickers []string) error {
    const batchSize = 20
    var wg sync.WaitGroup

    for i := 0; i < len(tickers); i += batchSize {
        end := i + batchSize
        if end > len(tickers) {
            end = len(tickers)
        }

        batch := tickers[i:end]
        wg.Add(1)

        go func(batch []string) {
            defer wg.Done()
            s.processTickerBatch(batch)
        }(batch)
    }

    wg.Wait()
    return nil
}

func (s *ProcessingStage) processTickerBatch(tickers []string) {
    for _, ticker := range tickers {
        // Process all data for ticker in one pass
        data := s.processTickerData(ticker)
        s.saveProcessedData(ticker, data)
    }
}
```

#### Memory-Efficient Processing
```go
func (s *ProcessingStage) processWithMemoryLimit() error {
    const maxMemoryMB = 100
    const recordsPerBatch = 1000

    // Process in batches to control memory usage
    batch := make([]PriceRecord, 0, recordsPerBatch)

    for {
        record, err := s.nextPriceRecord()
        if err == io.EOF {
            break
        }

        batch = append(batch, record)

        if len(batch) >= recordsPerBatch {
            if err := s.processBatch(batch); err != nil {
                return err
            }
            batch = batch[:0] // Reset slice
        }
    }

    // Process remaining records
    if len(batch) > 0 {
        return s.processBatch(batch)
    }

    return nil
}
```

### 🎯 **EOD Data Quality Assurance**

#### Validation Framework
```go
type DataValidator struct {
    rules []ValidationRule
}

func (v *DataValidator) ValidateEODData(data []PriceRecord) ValidationReport {
    report := ValidationReport{
        TotalRecords: len(data),
        ValidRecords: 0,
        Errors:       []ValidationError{},
    }

    for _, record := range data {
        recordValid := true

        for _, rule := range v.rules {
            if err := rule.Validate(record); err != nil {
                report.Errors = append(report.Errors, ValidationError{
                    Record: record.Ticker,
                    Field:  rule.Field,
                    Error:  err.Error(),
                })
                recordValid = false
            }
        }

        if recordValid {
            report.ValidRecords++
        }
    }

    return report
}
```

#### Anomaly Detection
```go
func (s *EODProcessor) detectAnomalies(data []PriceRecord) []Anomaly {
    var anomalies []Anomaly

    for _, record := range data {
        // Price anomaly detection
        if s.isPriceAnomalous(record) {
            anomalies = append(anomalies, Anomaly{
                Type:        "price_spike",
                Ticker:      record.Ticker,
                Date:        record.Date,
                Value:       record.Close,
                Expected:    s.getExpectedPrice(record),
                Confidence:  0.95,
            })
        }

        // Volume anomaly detection
        if s.isVolumeAnomalous(record) {
            anomalies = append(anomalies, Anomaly{
                Type:        "volume_spike",
                Ticker:      record.Ticker,
                Date:        record.Date,
                Value:       record.Volume,
                Expected:    s.getExpectedVolume(record),
                Confidence:  0.90,
            })
        }
    }

    return anomalies
}
```

### 🔄 **EOD Cache Management**

#### Multi-Level Caching Strategy
```go
type EODCacheManager struct {
    l1Cache *sync.Map          // In-memory cache (current day)
    l2Cache *FileCache         // File-based cache (recent days)
    l3Cache *CompressedCache   // Compressed cache (historical)
}

func (c *EODCacheManager) Get(date string, key string) (interface{}, error) {
    // L1: Current day in-memory
    if date == c.currentDate {
        if value, ok := c.l1Cache.Load(key); ok {
            return value, nil
        }
    }

    // L2: Recent days file cache
    if c.isRecentDate(date) {
        return c.l2Cache.Get(date, key)
    }

    // L3: Historical compressed cache
    return c.l3Cache.Get(date, key)
}
```

#### Cache Warming Strategy
```go
func (c *EODCacheManager) WarmCache() error {
    // Pre-load common queries
    commonQueries := []CacheQuery{
        {Date: c.currentDate, Type: "price_data", Tickers: c.getActiveTickers()},
        {Date: c.currentDate, Type: "market_data"},
        {Date: c.currentDate, Type: "index_data"},
    }

    for _, query := range commonQueries {
        data, err := c.loadData(query)
        if err != nil {
            slog.Error("cache warming failed", "query", query, "error", err)
            continue
        }

        c.store(query, data)
    }

    return nil
}
```

### 📈 **EOD Monitoring and Observability**

#### EOD Processing Metrics
```go
type EODMetrics struct {
    Date                    time.Time `json:"date"`
    ProcessingStartTime     time.Time `json:"processing_start_time"`
    ProcessingEndTime       time.Time `json:"processing_end_time"`
    TotalDuration           string    `json:"total_duration"`
    RecordsProcessed        int       `json:"records_processed"`
    ValidationErrors        int       `json:"validation_errors"`
    CacheHitRate           float64   `json:"cache_hit_rate"`
    FileSizes              map[string]int64 `json:"file_sizes"`
}

func (m *EODMetrics) RecordStageMetrics(stage string, duration time.Duration, records int) {
    slog.Info("EOD stage completed",
        "date", m.Date,
        "stage", stage,
        "duration_ms", duration.Milliseconds(),
        "records", records,
    )
}
```

#### Health Check Implementation
```go
func (s *EODProcessor) HealthCheck() HealthStatus {
    latestData := s.getLatestDataDate()
    now := time.Now()

    // Check if latest data is within expected window
    expectedWindow := time.Hour * 6 // 6 hours after market close
    if now.Sub(latestData) > expectedWindow {
        return HealthStatus{
            Status:  "unhealthy",
            Message: fmt.Sprintf("Latest data from %s is too old", latestData),
        }
    }

    // Check data quality
    if s.hasValidationErrors() {
        return HealthStatus{
            Status:  "degraded",
            Message: "Data validation errors detected",
        }
    }

    return HealthStatus{
        Status:  "healthy",
        Message: fmt.Sprintf("EOD processing complete for %s", latestData),
    }
}
```

### 🎯 **EOD Integration Patterns**

#### Frontend EOD Data Access
```typescript
// EOD-specific API calls
class EODDataService {
    async getLatestPriceData(ticker: string): Promise<PriceData> {
        const response = await fetch(`/api/eod/price/${ticker}`);
        return response.json(); // Returns cached EOD data instantly
    }

    async getMarketData(date?: string): Promise<MarketData> {
        const targetDate = date || this.getLatestTradingDate();
        const response = await fetch(`/api/eod/market/${targetDate}`);
        return response.json();
    }
}
```

---

**EOD-Processing Advantage**: By focusing on end-of-day data processing, ISX Pulse achieves superior performance, reliability, and data quality while eliminating the complexity and costs associated with real-time data systems.

## CSV-Based Data Storage with Caching and Indexing

**Design Philosophy**: ISX Pulse uses CSV as the primary storage format for all financial data, combined with intelligent caching and indexing strategies to achieve optimal performance for both data processing and real-time access.

### 📁 **CSV Storage Architecture**

#### Why CSV for Financial Data?
1. **Human Readable**: Easy to inspect and debug data issues
2. **Universal Compatibility**: Works with any data analysis tool
3. **Performance**: Excellent read performance with proper indexing
4. **Version Control**: Git-friendly for tracking data changes
5. **Backup Simplicity**: Simple file-based backup and restore
6. **Compressibility**: Excellent compression ratios for archival

#### CSV File Organization
```
data/
├── csv/
│   ├── prices/
│   │   ├── price_data_2025-10-17.csv
│   │   ├── price_data_2025-10-16.csv
│   │   └── price_data_latest.csv → price_data_2025-10-17.csv
│   ├── indices/
│   │   ├── isx60_index_2025-10-17.csv
│   │   └── market_summary_2025-10-17.csv
│   ├── liquidity/
│   │   └── liquidity_metrics_2025-10-17.csv
│   └── metadata/
│       ├── schema_definitions.json    # CSV column schemas
│       ├── file_manifest.json        # File metadata and checksums
│       └── processing_log.json        # Data processing history
├── indexes/
│   ├── ticker_index.json          # Ticker → file mappings
│   ├── date_index.json            # Date → file mappings
│   └── composite_index.json       # Multi-dimensional indexes
├── cache/
│   ├── memory_cache.db            # In-memory cache dump
│   ├── compressed_cache/          # Compressed historical data
│   └── index_cache.json          # Cached index mappings
```

### 🗂️ **CSV Schema Standards**

#### Price Data Schema
```csv
ticker,date,open,high,low,close,volume,change_pct,market_cap,sector
ISX-BKCP,2025-10-17,1.25,1.30,1.22,1.28,150000,2.40,1250000000,Banking
ISX-BMHR,2025-10-17,0.85,0.88,0.82,0.86,95000,-1.15,860000000,Manufacturing
```

#### Schema Definition Format
```json
{
  "schema_name": "price_data",
  "version": "1.2",
  "created_at": "2025-10-17T16:30:00Z",
  "columns": [
    {
      "name": "ticker",
      "type": "string",
      "required": true,
      "index": true,
      "description": "ISX ticker symbol (format: XXX-YYYY)"
    },
    {
      "name": "date",
      "type": "date",
      "required": true,
      "index": true,
      "description": "Trading date (YYYY-MM-DD format)"
    },
    {
      "name": "close",
      "type": "float",
      "required": true,
      "index": false,
      "constraints": {"min": 0, "precision": 4},
      "description": "Closing price in USD"
    }
  ],
  "primary_key": ["ticker", "date"],
  "indexes": [
    {"name": "ticker_index", "columns": ["ticker"]},
    {"name": "date_index", "columns": ["date"]},
    {"name": "ticker_date_index", "columns": ["ticker", "date"]}
  ]
}
```

### ⚡ **High-Performance CSV Indexing**

#### Multi-Level Index Structure
```go
type CSVIndex struct {
    PrimaryIndex   map[string]map[string]int    // ticker -> date -> row_offset
    SecondaryIndex map[string][]string          // date -> []tickers
    ReverseIndex  map[int]IndexEntry           // row_offset -> {ticker, date}
    Metadata      IndexMetadata
}

type IndexEntry struct {
    Ticker string
    Date   string
    Offset int64    // Byte offset in CSV file
    Length int      // Row length in bytes
}

type IndexManager struct {
    indexes map[string]*CSVIndex    // filename -> index
    mutex   sync.RWMutex            // Thread-safe access
    cache   *IndexCache             // In-memory index cache
}

// Build index for CSV file
func (im *IndexManager) BuildIndex(csvPath string) (*CSVIndex, error) {
    file, err := os.Open(csvPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    index := &CSVIndex{
        PrimaryIndex:   make(map[string]map[string]int),
        SecondaryIndex: make(map[string][]string),
        ReverseIndex:  make(map[int]IndexEntry),
    }

    reader := csv.NewReader(file)
    headers, err := reader.Read() // Skip header
    if err != nil {
        return nil, err
    }

    rowNumber := 0
    offset := int64(0)

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        // Extract key fields
        ticker := record[0]  // Assuming ticker is first column
        date := record[1]    // Assuming date is second column

        // Build primary index
        if index.PrimaryIndex[ticker] == nil {
            index.PrimaryIndex[ticker] = make(map[string]int)
        }
        index.PrimaryIndex[ticker][date] = rowNumber

        // Build secondary index
        index.SecondaryIndex[date] = append(index.SecondaryIndex[date], ticker)

        // Build reverse index
        index.ReverseIndex[rowNumber] = IndexEntry{
            Ticker: ticker,
            Date:   date,
            Offset: offset,
            Length: len(strings.Join(record, ",")) + 1, // +1 for newline
        }

        rowNumber++
        offset += int64(len(strings.Join(record, ",")) + 1)
    }

    return index, nil
}
```

#### Optimized CSV Access Patterns
```go
type CSVReader struct {
    file   *os.File
    index  *CSVIndex
    cache  map[string][]string  // Cache of recent reads
    mutex  sync.RWMutex
}

// Fast lookup using index
func (r *CSVReader) GetRow(ticker, date string) ([]string, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("%s:%s", ticker, date)
    r.mutex.RLock()
    if row, exists := r.cache[cacheKey]; exists {
        r.mutex.RUnlock()
        return row, nil
    }
    r.mutex.RUnlock()

    // Use index to find row location
    if r.index.PrimaryIndex[ticker] == nil {
        return nil, fmt.Errorf("ticker not found: %s", ticker)
    }

    rowNumber, exists := r.index.PrimaryIndex[ticker][date]
    if !exists {
        return nil, fmt.Errorf("date not found for ticker %s: %s", ticker, date)
    }

    // Seek directly to row position
    entry := r.index.ReverseIndex[rowNumber]
    _, err := r.file.Seek(entry.Offset, io.SeekStart)
    if err != nil {
        return nil, err
    }

    // Read only the specific row
    reader := bufio.NewReader(r.file)
    line, err := reader.ReadString('\n')
    if err != nil {
        return nil, err
    }

    // Parse CSV row
    record := strings.Split(strings.TrimSpace(line), ",")

    // Update cache
    r.mutex.Lock()
    r.cache[cacheKey] = record
    if len(r.cache) > 1000 { // Limit cache size
        r.evictOldestCacheEntry()
    }
    r.mutex.Unlock()

    return record, nil
}

// Batch read for multiple tickers
func (r *CSVReader) GetBatch(tickers []string, date string) (map[string][]string, error) {
    result := make(map[string][]string)

    // Sort by file position for optimal I/O
    type lookup struct {
        ticker string
        row    int
    }

    var lookups []lookup
    for _, ticker := range tickers {
        if rowNumber, exists := r.index.PrimaryIndex[ticker][date]; exists {
            lookups = append(lookups, lookup{ticker: ticker, row: rowNumber})
        }
    }

    // Sort by file offset
    sort.Slice(lookups, func(i, j int) bool {
        return r.index.ReverseIndex[lookups[i].row].Offset <
               r.index.ReverseIndex[lookups[j].row].Offset
    })

    // Sequential read for optimal performance
    currentOffset := int64(-1)
    for _, lookup := range lookups {
        entry := r.index.ReverseIndex[lookup.row]

        if entry.Offset != currentOffset {
            // Seek to new position
            _, err := r.file.Seek(entry.Offset, io.SeekStart)
            if err != nil {
                return nil, err
            }
            currentOffset = entry.Offset
        }

        // Read row
        reader := bufio.NewReader(r.file)
        line, err := reader.ReadString('\n')
        if err != nil {
            continue
        }

        record := strings.Split(strings.TrimSpace(line), ",")
        result[lookup.ticker] = record
    }

    return result, nil
}
```

### 🗄️ **Intelligent Caching System**

#### Multi-Tier Cache Architecture
```go
type CacheTier int

const (
    MemoryCache CacheTier = iota
    FileCache
    CompressedCache
)

type CacheManager struct {
    l1Cache *MemoryCache    // Hot data (current day)
    l2Cache *FileCache      // Recent data (last 7 days)
    l3Cache *CompressedCache // Historical data (older than 7 days)
    metrics *CacheMetrics
}

type MemoryCache struct {
    data sync.Map           // ticker:date -> data
    maxSize int             // Maximum number of entries
    evictions int64         // Eviction counter
    mutex sync.RWMutex
}

type FileCache struct {
    basePath string
    index    map[string]string  // key -> file path
    mutex    sync.RWMutex
}

type CompressedCache struct {
    basePath string
    compressionLevel int
    index    map[string]string  // key -> compressed file path
    mutex    sync.RWMutex
}

// Get data with automatic tier selection
func (cm *CacheManager) Get(ticker, dataType, date string) (interface{}, error) {
    key := fmt.Sprintf("%s:%s:%s", ticker, dataType, date)

    // L1: Memory cache (fastest)
    if value, err := cm.l1Cache.Get(key); err == nil {
        cm.metrics.RecordHit(MemoryCache)
        return value, nil
    }

    // L2: File cache (fast)
    if cm.isRecentDate(date) {
        if value, err := cm.l2Cache.Get(key); err == nil {
            cm.metrics.RecordHit(FileCache)
            // Promote to L1
            cm.l1Cache.Set(key, value)
            return value, nil
        }
    }

    // L3: Compressed cache (slower)
    if value, err := cm.l3Cache.Get(key); err == nil {
        cm.metrics.RecordHit(CompressedCache)
        return value, nil
    }

    cm.metrics.RecordMiss()
    return nil, fmt.Errorf("data not found in cache")
}
```

#### Cache Warming and Preloading
```go
func (cm *CacheManager) WarmCacheForDate(date string) error {
    // Pre-load commonly accessed data
    commonQueries := []struct {
        tickers   []string
        dataTypes []string
    }{
        {cm.getMostActiveTickers(10), []string{"price_data", "volume_data"}},
        {cm.getAllTickers(), []string{"price_data"}},
        {cm.getIndexConstituents(), cm.getAllDataTypes()},
    }

    for _, query := range commonQueries {
        for _, ticker := range query.tickers {
            for _, dataType := range query.dataTypes {
                value, err := cm.loadFromSource(ticker, dataType, date)
                if err != nil {
                    slog.Warn("failed to load data for cache warming",
                        "ticker", ticker,
                        "data_type", dataType,
                        "date", date,
                        "error", err)
                    continue
                }

                key := fmt.Sprintf("%s:%s:%s", ticker, dataType, date)
                cm.l1Cache.Set(key, value)
            }
        }
    }

    slog.Info("cache warming completed",
        "date", date,
        "entries_loaded", cm.l1Cache.Size(),
        "memory_usage_mb", cm.l1Cache.MemoryUsageMB())

    return nil
}
```

### 📊 **CSV Performance Optimization**

#### Compression and Archival
```go
type CSVArchiver struct {
    compressionLevel int
    retentionDays    int
    basePath        string
}

func (a *CSVArchiver) ArchiveOldFiles() error {
    cutoff := time.Now().AddDate(0, 0, -a.retentionDays)

    err := filepath.Walk(a.basePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() || !strings.HasSuffix(path, ".csv") {
            return nil
        }

        if info.ModTime().Before(cutoff) {
            return a.compressFile(path)
        }

        return nil
    })

    return err
}

func (a *CSVArchiver) compressFile(filePath string) error {
    inputFile, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer inputFile.Close()

    compressedPath := filePath + ".gz"
    outputFile, err := os.Create(compressedPath)
    if err != nil {
        return err
    }
    defer outputFile.Close()

    gzipWriter := gzip.NewWriter(outputFile)
    gzipWriter.CompressionLevel = a.compressionLevel

    _, err = io.Copy(gzipWriter, inputFile)
    if err != nil {
        return err
    }

    if err := gzipWriter.Close(); err != nil {
        return err
    }

    // Remove original file
    return os.Remove(filePath)
}
```

#### Parallel CSV Processing
```go
type CSVProcessor struct {
    workers   int
    batchSize int
}

func (p *CSVProcessor) ProcessLargeCSV(filePath string, processor func([]string) error) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    reader := csv.NewReader(file)

    // Read header
    header, err := reader.Read()
    if err != nil {
        return err
    }

    // Create worker pool
    jobs := make(chan []string, p.batchSize*2)
    results := make(chan error, p.workers)

    var wg sync.WaitGroup

    // Start workers
    for i := 0; i < p.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for batch := range jobs {
                for _, record := range batch {
                    if err := processor(record); err != nil {
                        results <- err
                        return
                    }
                }
            }
        }()
    }

    // Error collection goroutine
    var processErrors []error
    go func() {
        for err := range results {
            processErrors = append(processErrors, err)
        }
    }()

    // Read and dispatch batches
    batch := make([]string, 0, p.batchSize)
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        batch = append(batch, record)

        if len(batch) >= p.batchSize {
            jobs <- batch
            batch = make([]string, 0, p.batchSize)
        }
    }

    // Process remaining records
    if len(batch) > 0 {
        jobs <- batch
    }

    close(jobs)
    wg.Wait()
    close(results)

    if len(processErrors) > 0 {
        return fmt.Errorf("processing errors: %v", processErrors)
    }

    return nil
}
```

### 🔍 **CSV Data Validation and Quality**

#### Schema Validation
```go
type CSVValidator struct {
    schema SchemaDefinition
}

func (v *CSVValidator) ValidateCSV(filePath string) ValidationReport {
    file, err := os.Open(filePath)
    if err != nil {
        return ValidationReport{Error: err.Error()}
    }
    defer file.Close()

    reader := csv.NewReader(file)
    headers, err := reader.Read()
    if err != nil {
        return ValidationReport{Error: err.Error()}
    }

    // Validate headers against schema
    if err := v.validateHeaders(headers); err != nil {
        return ValidationReport{Error: fmt.Sprintf("header validation failed: %v", err)}
    }

    report := ValidationReport{
        TotalRows: 0,
        ValidRows: 0,
        Errors:    []ValidationError{},
    }

    rowNumber := 1 // Start at 1 (after header)
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            report.Errors = append(report.Errors, ValidationError{
                Row:    rowNumber,
                Field:  "parse",
                Error:  err.Error(),
            })
            break
        }

        rowNumber++
        report.TotalRows++

        // Validate row against schema
        if rowErrors := v.validateRow(record, rowNumber); len(rowErrors) > 0 {
            report.Errors = append(report.Errors, rowErrors...)
        } else {
            report.ValidRows++
        }
    }

    return report
}

func (v *CSVValidator) validateRow(record []string, rowNumber int) []ValidationError {
    var errors []ValidationError

    for i, value := range record {
        if i >= len(v.schema.Columns) {
            errors = append(errors, ValidationError{
                Row:    rowNumber,
                Field:  fmt.Sprintf("column_%d", i+1),
                Error:  "too many columns in row",
            })
            continue
        }

        column := v.schema.Columns[i]

        // Required field validation
        if column.Required && (value == "" || strings.TrimSpace(value) == "") {
            errors = append(errors, ValidationError{
                Row:    rowNumber,
                Field:  column.Name,
                Error:  "required field is empty",
            })
            continue
        }

        // Type validation
        if value != "" {
            if err := v.validateType(value, column.Type, column.Constraints); err != nil {
                errors = append(errors, ValidationError{
                    Row:    rowNumber,
                    Field:  column.Name,
                    Error:  fmt.Sprintf("type validation failed: %v", err),
                })
            }
        }
    }

    return errors
}
```

### 📈 **CSV Analytics and Monitoring**

#### File Usage Analytics
```go
type CSVAnalytics struct {
    accessCounts map[string]int64    // file_path -> access_count
    accessTimes  map[string]time.Time // file_path -> last_access
    queryStats  map[string]int      // query_type -> count
    mutex       sync.RWMutex
}

func (a *CSVAnalytics) RecordAccess(filePath, queryType string) {
    a.mutex.Lock()
    defer a.mutex.Unlock()

    a.accessCounts[filePath]++
    a.accessTimes[filePath] = time.Now()
    a.queryStats[queryType]++
}

func (a *CSVAnalytics) GetUsageReport() UsageReport {
    a.mutex.RLock()
    defer a.mutex.RUnlock()

    report := UsageReport{
        TotalFilesAccessed: len(a.accessCounts),
        TotalQueries:       0,
        MostAccessedFiles:  []FileUsage{},
        QueryBreakdown:     a.queryStats,
    }

    // Calculate total queries
    for _, count := range a.accessCounts {
        report.TotalQueries += count
    }

    // Find most accessed files
    type fileUsage struct {
        Path  string
        Count int64
    }

    var usages []fileUsage
    for path, count := range a.accessCounts {
        usages = append(usages, fileUsage{Path: path, Count: count})
    }

    sort.Slice(usages, func(i, j int) bool {
        return usages[i].Count > usages[j].Count
    })

    for i, usage := range usages {
        if i >= 10 { // Top 10
            break
        }
        report.MostAccessedFiles = append(report.MostAccessedFiles, FileUsage{
            Path:         usage.Path,
            AccessCount:  usage.Count,
            LastAccessed: a.accessTimes[usage.Path],
        })
    }

    return report
}
```

---

**CSV Storage Advantage**: The CSV-based storage system provides the perfect balance of performance, reliability, and maintainability for financial data, while the intelligent caching and indexing ensure sub-millisecond access times for all queries.

## Thread-Safe Service Patterns

**Design Philosophy**: ISX Pulse implements comprehensive thread-safety patterns throughout all services to ensure reliable concurrent access to shared resources while maintaining optimal performance in multi-user environments.

### 🔒 **Thread Safety Fundamentals**

#### Why Thread Safety Matters
1. **Concurrent Access**: Multiple users accessing data simultaneously
2. **Pipeline Processing**: Multiple goroutines processing data concurrently
3. **Cache Consistency**: Shared cache accessed by read/write operations
4. **Data Integrity**: Preventing race conditions in financial data
5. **WebSocket Updates**: Real-time updates while users are reading data

#### Core Thread Safety Principles
- **Immutable Data**: Where possible, use immutable data structures
- **Copy-on-Write**: Minimize shared mutable state
- **Explicit Locking**: Clear locking hierarchy to avoid deadlocks
- **Lock-Free Patterns**: Use atomic operations where appropriate
- **Resource Pooling**: Manage shared resources efficiently

### 🔧 **Thread-Safe Patterns Implementation**

#### 1. Read-Write Mutex Pattern
```go
type DataService struct {
    dataCache map[string]*DataRecord
    csvIndex  map[string]int
    mutex     sync.RWMutex  // Multiple readers, single writer
    dataPath  string
}

// Safe read operation - allows concurrent readers
func (s *DataService) GetData(ticker, dataType, date string) (interface{}, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    // Multiple goroutines can read simultaneously
    if tickerData, exists := s.dataCache[ticker]; exists {
        if value, exists := tickerData.Values[dataType][date]; exists {
            return value, nil
        }
    }

    return nil, fmt.Errorf("data not found")
}

// Safe write operation - exclusive access
func (s *DataService) UpdateData(newData map[string]*DataRecord) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Only one goroutine can write at a time
    for ticker, data := range newData {
        s.dataCache[ticker] = data
    }

    return nil
}
```

#### 2. Atomic Operations Pattern
```go
type AtomicCounter struct {
    value int64
}

func (c *AtomicCounter) Increment() int64 {
    return atomic.AddInt64(&c.value, 1)
}

func (c *AtomicCounter) Get() int64 {
    return atomic.LoadInt64(&c.value)
}

// Usage in service metrics
type ServiceMetrics struct {
    queryCount   AtomicCounter
    cacheHits    AtomicCounter
    cacheMisses  AtomicCounter
    errorCount   AtomicCounter
}

func (m *ServiceMetrics) RecordQuery() {
    m.queryCount.Increment()
}

func (m *ServiceMetrics) RecordCacheHit() {
    m.cacheHits.Increment()
}
```

#### 3. Channel-Based Coordination Pattern
```go
type PipelineWorker struct {
    jobs     <-chan Job
    results  chan<- Result
    quit     <-chan struct{}
    workerID int
}

func (w *PipelineWorker) Start() {
    go func() {
        for {
            select {
            case job := <-w.jobs:
                // Process job independently
                result := w.processJob(job)
                w.results <- result

            case <-w.quit:
                // Graceful shutdown
                slog.Info("worker shutting down", "worker_id", w.workerID)
                return
            }
        }
    }()
}

func (w *PipelineWorker) processJob(job Job) Result {
    // Each worker processes independently - no shared state
    // Thread safety achieved through message passing
    return Result{
        WorkerID: w.workerID,
        JobID:    job.ID,
        Success:  true,
        Data:     job.Data,
    }
}
```

#### 4. Sync.Map Pattern for High-Concurrency Caches
```go
type HighConcurrencyCache struct {
    data sync.Map  // Thread-safe map for high-concurrency scenarios
    stats CacheStats
    mutex sync.RWMutex  // Only for stats updates
}

type CacheEntry struct {
    Value       interface{}
    Expiration time.Time
    AccessCount int64
}

func (c *HighConcurrencyCache) Get(key string) (interface{}, bool) {
    if entry, exists := c.data.Load(key); exists {
        cacheEntry := entry.(*CacheEntry)

        // Check expiration
        if time.Now().Before(cacheEntry.Expiration) {
            // Atomically increment access count
            atomic.AddInt64(&cacheEntry.AccessCount, 1)

            // Update stats (less frequent operation)
            c.mutex.Lock()
            c.stats.Hits++
            c.mutex.Unlock()

            return cacheEntry.Value, true
        } else {
            // Remove expired entry
            c.data.Delete(key)
        }
    }

    // Update miss stats
    c.mutex.Lock()
    c.stats.Misses++
    c.mutex.Unlock()

    return nil, false
}

func (c *HighConcurrencyCache) Set(key string, value interface{}, ttl time.Duration) {
    entry := &CacheEntry{
        Value:      value,
        Expiration: time.Now().Add(ttl),
        AccessCount: 0,
    }

    c.data.Store(key, entry)
}
```

#### 5. Resource Pool Pattern
```go
type ResourcePool struct {
    resources chan Resource
    factory   func() Resource
    maxSize   int
    mutex     sync.Mutex
    created   int
}

type Resource struct {
    id     int
    data   []byte
    active bool
}

func NewResourcePool(maxSize int, factory func() Resource) *ResourcePool {
    pool := &ResourcePool{
        resources: make(chan Resource, maxSize),
        factory:   factory,
        maxSize:   maxSize,
    }

    // Pre-allocate some resources
    for i := 0; i < maxSize/2; i++ {
        pool.resources <- factory()
        pool.created++
    }

    return pool
}

func (p *ResourcePool) Get() Resource {
    select {
    case resource := <-p.resources:
        // Return existing resource
        resource.active = true
        return resource

    default:
        // No available resources, create new one if under limit
        p.mutex.Lock()
        if p.created < p.maxSize {
            resource := p.factory()
            p.created++
            p.mutex.Unlock()
            resource.active = true
            return resource
        }
        p.mutex.Unlock()

        // Wait for available resource
        resource := <-p.resources
        resource.active = true
        return resource
    }
}

func (p *ResourcePool) Put(resource Resource) {
    if resource.active {
        // Reset resource state if needed
        resource.active = false

        select {
        case p.resources <- resource:
            // Resource returned to pool
        default:
            // Pool is full, discard resource
        }
    }
}
```

### 🛡️ **Thread-Safe Service Implementation**

#### Data Service Example
```go
type DataService struct {
    // Core data structures
    dataCache map[string]*TickerData
    csvIndex  *CSVIndex
    dataPath  string

    // Thread safety controls
    cacheMutex    sync.RWMutex
    indexMutex    sync.RWMutex
    updateMutex   sync.Mutex  // For exclusive updates

    // Performance monitoring
    metrics       *ServiceMetrics
    accessLog     sync.Map  // ticker -> last access time

    // Background workers
    cacheWarmer   *CacheWarmer
    indexBuilder  *IndexBuilder
}

type TickerData struct {
    Ticker      string
    Values      map[string]map[string]interface{}  // data_type -> date -> value
    LastUpdated time.Time
    Version     int64
}

// Thread-safe data retrieval
func (s *DataService) GetDataForTickers(tickers []string, dataTypes []string, dateRange []string) (map[string]map[string]map[string]interface{}, error) {
    // Update access time
    for _, ticker := range tickers {
        s.accessLog.Store(ticker, time.Now())
    }

    // Try cache first (shared read lock)
    s.cacheMutex.RLock()
    result := make(map[string]map[string]map[string]interface{})

    for _, ticker := range tickers {
        if tickerData, exists := s.dataCache[ticker]; exists {
            result[ticker] = make(map[string]map[string]interface{})
            for _, dataType := range dataTypes {
                result[ticker][dataType] = make(map[string]interface{})
                for _, date := range dateRange {
                    if value, exists := tickerData.Values[dataType][date]; exists {
                        result[ticker][dataType][date] = value
                    }
                }
            }
        }
    }
    s.cacheMutex.RUnlock()

    // Check if we got all data from cache
    allFound := true
    for _, ticker := range tickers {
        if _, exists := result[ticker]; !exists {
            allFound = false
            break
        }
    }

    if allFound {
        s.metrics.RecordCacheHit()
        return result, nil
    }

    // Cache miss - load from CSV
    s.metrics.RecordCacheMiss()
    return s.loadDataFromCSV(tickers, dataTypes, dateRange)
}

// Thread-safe cache update
func (s *DataService) UpdateCache(newData map[string]*TickerData) error {
    s.updateMutex.Lock()
    defer s.updateMutex.Unlock()

    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()

    // Update cache atomically
    for ticker, data := range newData {
        s.dataCache[ticker] = data
    }

    // Invalidate affected index entries
    s.indexMutex.Lock()
    s.csvIndex.InvalidateAffectedTickers(newData)
    s.indexMutex.Unlock()

    return nil
}

// Background cache warming
func (s *DataService) StartCacheWarmer() {
    s.cacheWarmer = &CacheWarmer{
        service:     s,
        interval:    5 * time.Minute,
        stopChannel: make(chan struct{}),
    }

    go s.cacheWarmer.Run()
}

type CacheWarmer struct {
    service     *DataService
    interval    time.Duration
    stopChannel chan struct{}
}

func (cw *CacheWarmer) Run() {
    ticker := time.NewTicker(cw.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            cw.warmCache()

        case <-cw.stopChannel:
            return
        }
    }
}

func (cw *CacheWarmer) warmCache() {
    // Get frequently accessed tickers
    activeTickers := cw.service.getMostActiveTickers(20)
    currentDate := time.Now().Format("2006-01-02")
    dataTypes := []string{"price_data", "volume_data"}

    for _, ticker := range activeTickers {
        go func(t string) {
            // Load in background without blocking
            cw.service.GetDataForTickers([]string{t}, dataTypes, []string{currentDate})
        }(ticker)
    }
}
```

### 📊 **Thread Safety Monitoring**

#### Concurrency Metrics
```go
type ConcurrencyMetrics struct {
    GoroutineCount    int64
    LockWaitTime      time.Duration
    LockContention    float64
    CacheHitRate      float64
    QueueLength       int64
    ActiveConnections int64
}

func (m *ConcurrencyMetrics) Collect() {
    // Goroutine count
    m.GoroutineCount = int64(runtime.NumGoroutine())

    // Lock contention (requires custom lock wrappers)
    m.LockContention = calculateLockContention()

    // Cache hit rate
    totalRequests := atomic.LoadInt64(&totalCacheRequests)
    cacheHits := atomic.LoadInt64(&totalCacheHits)
    if totalRequests > 0 {
        m.CacheHitRate = float64(cacheHits) / float64(totalRequests)
    }
}

type LockWrapper struct {
    mu       sync.RWMutex
    waitTime time.Duration
    acquisitions int64
}

func (lw *LockWrapper) Lock() {
    start := time.Now()
    lw.mu.Lock()
    lw.waitTime += time.Since(start)
    atomic.AddInt64(&lw.acquisitions, 1)
}

func (lw *LockWrapper) Unlock() {
    lw.mu.Unlock()
}
```

#### Deadlock Detection
```go
type DeadlockDetector struct {
    lockGraph map[string][]string  // resource -> waiting goroutines
    mutex     sync.Mutex
}

func (dd *DeadlockDetector) RegisterWait(goroutineID, resourceID string) {
    dd.mutex.Lock()
    defer dd.mutex.Unlock()

    if dd.lockGraph == nil {
        dd.lockGraph = make(map[string][]string)
    }

    dd.lockGraph[resourceID] = append(dd.lockGraph[resourceID], goroutineID)

    // Check for cycles
    if dd.hasCycle(goroutineID, make(map[string]bool)) {
        slog.Error("Potential deadlock detected",
            "goroutine", goroutineID,
            "resource", resourceID,
            "graph", dd.lockGraph)
    }
}

func (dd *DeadlockDetector) hasCycle(node string, visited map[string]bool) bool {
    if visited[node] {
        return true // Cycle detected
    }

    visited[node] = true

    for _, neighbor := range dd.lockGraph[node] {
        if dd.hasCycle(neighbor, visited) {
            return true
        }
    }

    delete(visited, node)
    return false
}
```

### 🎯 **Thread Safety Best Practices**

#### 1. Lock Ordering Hierarchy
```go
// Define a clear lock ordering to prevent deadlocks
const (
    LockOrderCache = iota
    LockOrderIndex
    LockOrderData
    LockOrderMetrics
)

type OrderedLock struct {
    order int
    mu    sync.RWMutex
}

func (ol *OrderedLock) Lock() {
    // Always acquire locks in increasing order
    if ol.order == LockOrderCache {
        // Cache lock can be acquired first
        ol.mu.Lock()
    } else if ol.order == LockOrderIndex {
        // Index lock must be acquired after cache lock
        ol.mu.Lock()
    }
    // ... etc
}
```

#### 2. Context-Aware Operations
```go
func (s *DataService) GetDataWithContext(ctx context.Context, ticker, dataType, date string) (interface{}, error) {
    // Check for context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    s.mutex.RLock()
    defer s.mutex.RUnlock()

    // Periodic context check during long operations
    if tickerData, exists := s.dataCache[ticker]; exists {
        if value, exists := tickerData.Values[dataType][date]; exists {
            return value, nil
        }
    }

    return nil, fmt.Errorf("data not found")
}
```

#### 3. Graceful Shutdown Patterns
```go
type ThreadSafeService struct {
    shutdownChan chan struct{}
    wg           sync.WaitGroup
    running      int32
}

func (s *ThreadSafeService) Start() {
    if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
        return // Already running
    }

    // Start background workers
    s.wg.Add(1)
    go s.backgroundWorker()

    s.wg.Add(1)
    go s.maintenanceWorker()
}

func (s *ThreadSafeService) Shutdown() {
    if !atomic.CompareAndSwapInt32(&s.running, 1, 0) {
        return // Not running
    }

    close(s.shutdownChan)
    s.wg.Wait()
}

func (s *ThreadSafeService) backgroundWorker() {
    defer s.wg.Done()

    for {
        select {
        case <-s.shutdownChan:
            return
        case task := <-s.taskQueue:
            s.processTask(task)
        }
    }
}
```

---

**Thread Safety Impact**: Comprehensive thread-safe patterns ensure reliable concurrent access to all shared resources while maintaining optimal performance in multi-user environments, making ISX Pulse robust and scalable for enterprise deployment.

## Build & Development Commands

### 🚨 MANDATORY BUILD RULES - CLAUDE CODE MUST FOLLOW
```
═══════════════════════════════════════════════════════════════════════
⚠️  ABSOLUTE BUILD RULES - NO EXCEPTIONS EVER  ⚠️
═══════════════════════════════════════════════════════════════════════
1. NEVER run 'npm run build' in web directory
2. NEVER run 'go build' in api/ directory
3. NEVER create .next/ directory in web
4. NEVER create out/ directory in web (except via build.bat)
5. ALWAYS use ./build.bat from project root for ALL builds
6. ALWAYS clear logs before each build (automatic in build.bat)
7. ALL builds MUST output to dist/ directory ONLY
8. NO build artifacts allowed in api/ or web/ directories ever
9. BEFORE any build, ALWAYS verify you're in project root
10. If web/.next exists, DELETE it immediately

ENFORCEMENT: Claude Code will refuse to run any build commands
that violate these rules and will suggest ./build.bat instead.
═══════════════════════════════════════════════════════════════════════
```

### Primary Build Command (THE ONLY WAY TO BUILD)
```bash
# ✅ CORRECT - From project root ONLY:
./build.bat              # Builds all to dist/ (clears logs first)
./build.bat -target=all  # Same as above
./build.bat -target=web  # Build web-licensed only to dist/
./build.bat -target=frontend  # Build frontend (embedded in exe)
./build.bat -target=clean     # Clean artifacts AND logs (PRESERVES data/ folder)
./build.bat -target=test      # Run all tests
./build.bat -target=activate-license  # Build license activation tool
./build.bat -target=release   # Create release in dist/

# 🛡️ AUTOMATIC DATA PRESERVATION:
# The build system AUTOMATICALLY preserves:
# - dist/license.dat (license activation data)
# - dist/data/ (ALL user data: downloads, reports, cache)
#
# These are NEVER deleted during clean builds unless explicitly requested.

# ❌ FORBIDDEN - Claude Code will REFUSE these:
cd web && npm run build              # NEVER
cd api && go build ./...             # NEVER
cd web && next build                 # NEVER
cd web && npm run export             # NEVER
```

### Development Commands (NO BUILDS ALLOWED)
```bash
# Backend development (NO BUILDING)
cd api
go run cmd/web-licensed/main.go  # Run server (dev mode only)
go test ./... -race              # Run tests
go test ./... -cover             # Coverage analysis

# Frontend development (DEV SERVER ONLY - NO BUILDS)
cd web
npm run dev                      # ✅ Dev server ONLY (OK)
npm run test                     # ✅ Run tests (OK)
npm run lint                     # ✅ Lint code (OK)
npm run type-check              # ✅ TypeScript check (OK)

# ⛔ ABSOLUTELY FORBIDDEN IN web:
npm run build                    # ❌ NEVER - Use ./build.bat
npm run export                   # ❌ NEVER - Use ./build.bat
next build                       # ❌ NEVER - Use ./build.bat
npx next build                   # ❌ NEVER - Use ./build.bat
```

## Frontend Embedding Standards

### Industry-Standard Pattern
Following Grafana, CockroachDB, and Kubernetes Dashboard best practices, we use **explicit file patterns** for embedding frontend assets instead of wildcards.

### Embedding Rules
1. **NEVER use wildcards** (`all:frontend/*`) - causes empty directory errors
2. **ALWAYS use explicit patterns** - security, performance, reproducibility
3. **Validate before embedding** - ensure all required files present
4. **Clean empty directories** - Next.js creates them, Go can't embed them

### Approved Embed Pattern
```go
//go:embed frontend/index.html frontend/404.html frontend/index.txt
//go:embed frontend/_next
//go:embed frontend/*.ico frontend/*.png frontend/*.svg
//go:embed frontend/site.webmanifest
var frontendFiles embed.FS
```

### Build Process for Frontend
1. `npm run build` in web (via build.bat ONLY)
2. Output copied to `api/cmd/web-licensed/frontend/`
3. Empty directories removed (Next.js quirk)
4. Validation ensures required files present
5. Go embeds using explicit patterns
6. Binary includes optimized frontend

### Required Frontend Assets
- `index.html` - Main entry point
- `404.html` - Error page
- `_next/` - Next.js assets (must not be empty)
- `favicon.ico` - Browser icon
- `site.webmanifest` - PWA manifest

### Forbidden in Production Build
- `*.map` files (source maps)
- `.env*` files (environment configs)
- `node_modules/` (dependencies)
- Empty directories

## Key Standards

### Go Standards
- **Router**: Chi v5 only - no Gin, Echo, or other routers
- **Logging**: slog only - no fmt.Println, log.Printf
- **Errors**: RFC 7807 Problem Details for all API errors
- **Context**: Always pass context.Context as first parameter
- **Testing**: Table-driven tests, minimum 80% coverage
- **Interfaces**: Define interfaces in consumer packages
- **Frontend Embedding**: Explicit file patterns only (no wildcards) following Grafana/CockroachDB standards
- **Embed Validation**: All embedded assets must be validated before build

### TypeScript/React Standards
- **Types**: Strict mode enabled, no any types
- **Components**: Functional components with hooks
- **State**: useState, useReducer for local; Context for global
- **Styling**: Tailwind CSS with Shadcn/ui components
- **Forms**: react-hook-form with zod validation
- **API**: Centralized API client with proper error handling

### Next.js Component Architecture

#### Server vs Client Components
Next.js 14+ uses React Server Components by default. Understanding when to use server vs client components is critical:

**Server Components (default - no 'use client'):**
- Can export metadata for SEO
- Can directly fetch data from databases
- Cannot use hooks, event handlers, or browser APIs
- Better performance (smaller bundle size)
- Use for static content, data fetching, SEO metadata

**Client Components ('use client' directive):**
- Required for interactivity (onClick, onChange, etc.)
- Required for hooks (useState, useEffect, etc.)
- Required for browser APIs (localStorage, window, etc.)
- Cannot export metadata (will cause build warnings)
- Use for forms, modals, real-time updates, animations

#### Server/Client Component Pattern
When you need both metadata (SEO) and interactivity, split components:

```typescript
// app/reports/page.tsx - Server component with metadata
import ReportsClient from './reports-client'

export const metadata = {
  title: 'Reports - ISX Pulse',
  description: 'Financial reports for the Iraqi Stock Exchange.',
  robots: { index: false, follow: false }
}

export default function ReportsPage() {
  return <ReportsClient />
}
```

```typescript
// app/reports/reports-client.tsx - Client component with interactivity
'use client'

import { useState, useCallback } from 'react'
import { useToast } from '@/lib/hooks/use-toast'

export default function ReportsClient() {
  const [data, setData] = useState()
  const { toast } = useToast()

  const handleClick = useCallback(() => {
    toast({ title: "Action completed" })
  }, [toast])

  return <button onClick={handleClick}>Interactive Button</button>
}
```

#### Key Rules:
1. **Never export metadata from client components** - It will be silently ignored
2. **Split pages when needed** - Use wrapper pattern for metadata + interactivity
3. **Minimize client components** - Only use when interactivity is required
4. **Import client components into server components** - Not vice versa
5. **Use proper file naming** - `page.tsx` for route, `*-client.tsx` for client components

### TradingView Lightweight Charts - Native Panes API

The project uses **TradingView Lightweight Charts v5.0.8** with native multi-pane support for professional chart visualization. Panes allow multiple components (Volume, Market Data) to be displayed in separate panels below the main price chart.

#### Pane Creation & Configuration

**Creating Panes:**
```typescript
// Initialize chart with panes configuration
const chart = createChart(container, {
  layout: {
    panes: {
      separatorColor: '#2B2B43',        // Color of separator line
      separatorHoverColor: '#4169E1',   // Hover color for separator
      enableResize: true                // Allow user to drag separators
    }
  }
})

// Add series to specific pane (paneIndex)
const volumeSeries = chart.addSeries(HistogramSeries, {
  color: '#10b981',
  priceFormat: { type: 'volume' },
}, 1)  // Creates Pane 1 automatically

// Access panes
const panes = chart.panes()
const volumePane = panes[1]

// Set pane height with constraints (30px min, 90% max)
const containerHeight = container.clientHeight || 600
const maxHeight = Math.floor(containerHeight * 0.90) // 90% max - keeps 10% for price chart
const desiredHeight = Math.floor(containerHeight * 0.25) // 25% default
const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
volumePane.setHeight(constrainedHeight)
```

#### Pane Height Constraints

**Minimum Height**: 30px (TradingView API hard limit)
**Maximum Height**: 90% of container height (keeps price chart visible at ≥10%)
**Default Heights**:
- Volume: 25% (~150px on 600px chart)

**Manual Resizing**: Users can drag separator lines between panes to adjust heights. The 90% maximum ensures the main price chart remains visible.

#### Professional Layout Proportions

**Single Component (Volume only):**
- Price Chart: 75% (Pane 0)
- Volume: 25% (Pane 1, resizable up to 90%)

#### Dynamic Pane Management

**Adding Volume Pane (with height constraints):**
```typescript
if (showVolume) {
  const volumeSeries = chart.addSeries(HistogramSeries, {
    color: '#10b981',
    priceFormat: { type: 'volume' },
  }, 1)  // Pane 1

  const panes = chart.panes()
  if (panes.length > 1) {
    const volumePane = panes[1]
    const containerHeight = containerRef.current?.clientHeight || 600
    const maxHeight = Math.floor(containerHeight * 0.90) // 90% max
    const desiredHeight = Math.floor(containerHeight * 0.25) // 25% default
    const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
    volumePane.setHeight(constrainedHeight)
  }
}
```

#### Pane Cleanup

Always remove series before removing panes:
```typescript
// Remove series first
if (volumeSeriesRef.current) {
  chart.removeSeries(volumeSeriesRef.current)
  volumeSeriesRef.current = null
}

// Panes are automatically removed when all their series are removed
```

#### Best Practices

1. **Dynamic Pane Indexing**: Always calculate pane index based on active components
2. **Minimum Height**: Never set pane height below 30px
3. **Cleanup**: Remove all series before unmounting
4. **Performance**: Panes API is production-ready and performant
5. **Crosshair**: Automatically synchronized across all panes
6. **Theme**: Pane separators update with theme changes

## Critical Patterns

### 1. Error Handling (Go)
```go
// Always use internal/errors package
import "internal/errors"

// Return RFC 7807 compliant errors
if err != nil {
    return errors.NewValidationError("invalid input", err).
        WithDetail("license key format is invalid").
        WithField("license_key", key)
}

// In handlers, use error middleware
func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
    data, err := h.service.GetData(r.Context())
    if err != nil {
        errors.HandleError(w, r, err)
        return
    }
    // ... success response
}
```

### 2. Context Usage
```go
// Always propagate context
func (s *Service) ProcessData(ctx context.Context, data []byte) error {
    // Add timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Check context in loops
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Process item
        }
    }
}
```

### 3. Structured Logging
```go
// Use slog with structured fields
slog.InfoContext(ctx, "processing started",
    "operation", "data_import",
    "trace_id", middleware.GetTraceID(ctx),
    "user_id", userID,
    "record_count", len(records),
)

// Log errors with full context
slog.ErrorContext(ctx, "processing failed",
    "error", err,
    "operation", "data_import",
    "trace_id", middleware.GetTraceID(ctx),
    "duration", time.Since(start),
)
```

### 4. Testing Patterns
```go
func TestService_ProcessData(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        want    *Result
        wantErr error
    }{
        {
            name:  "valid data",
            input: []byte(`{"id": 1}`),
            want:  &Result{ID: 1},
        },
        {
            name:    "invalid data",
            input:   []byte(`invalid`),
            wantErr: errors.ErrInvalidFormat,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewService()
            got, err := svc.ProcessData(context.Background(), tt.input)

            if tt.wantErr != nil {
                require.Error(t, err)
                assert.ErrorIs(t, err, tt.wantErr)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Error Handling Best Practices

### 1. Use Custom Error Types
```go
// Define in internal/errors
type ValidationError struct {
    BaseError
    Field string
    Value interface{}
}

type NotFoundError struct {
    BaseError
    Resource string
    ID       string
}

// Usage
return &NotFoundError{
    Resource: "report",
    ID:       reportID,
    BaseError: BaseError{
        Message: "report not found",
        Code:    "REPORT_NOT_FOUND",
    },
}
```

### 2. Error Wrapping
```go
// Wrap errors with context
if err := db.Query(ctx, query); err != nil {
    return fmt.Errorf("query reports: %w", err)
}

// Check wrapped errors
if errors.Is(err, sql.ErrNoRows) {
    return NewNotFoundError("report", id)
}
```

### 3. HTTP Error Responses
```go
// All HTTP errors must follow RFC 7807
{
    "type": "/errors/validation-failed",
    "title": "Validation Failed",
    "status": 400,
    "detail": "The license key format is invalid",
    "instance": "/api/v1/license/activate",
    "trace_id": "abc123",
    "errors": {
        "license_key": "must be in format XXXX-XXXX-XXXX"
    }
}
```

## Interface & Dependency Management

### 1. Interface Definition
```go
// Define interfaces where they're used, not where implemented
package service

type Repository interface {
    GetReport(ctx context.Context, id string) (*Report, error)
    SaveReport(ctx context.Context, report *Report) error
}

// Accept interfaces, return structs
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}
```

### 2. Dependency Injection
```go
// Use constructor injection
type Service struct {
    repo   Repository
    cache  Cache
    logger *slog.Logger
}

func NewService(repo Repository, cache Cache, logger *slog.Logger) *Service {
    return &Service{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}
```

### 3. Wire Everything in main()
```go
func main() {
    // Initialize dependencies
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    db := database.New(config.DB)
    cache := cache.NewRedis(config.Redis)

    // Wire services
    repo := repository.New(db)
    svc := service.New(repo, cache, logger)
    handler := http.NewHandler(svc, logger)

    // Start server
    server := http.NewServer(handler, config.Server)
    server.Start()
}
```

## Testing Requirements

### 1. Coverage Requirements
- Minimum 80% coverage for all packages
- 90% for critical paths (licensing, operations, security)
- 100% for security-related functions

### 2. Test Types
```go
// Unit tests - internal/service/report_test.go
func TestReportService_Generate(t *testing.T) {
    // Test with mocks
}

// Integration tests - internal/integration/
func TestReportAPI_EndToEnd(t *testing.T) {
    // Test with real database
}

// Benchmark tests
func BenchmarkReportGeneration(b *testing.B) {
    // Performance testing
}
```

### 3. Test Helpers
```go
// Use testify for assertions
assert.Equal(t, expected, actual)
require.NoError(t, err)

// Use mock interfaces
type MockRepo struct {
    mock.Mock
}

func (m *MockRepo) GetReport(ctx context.Context, id string) (*Report, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Report), args.Error(1)
}
```

## Configuration Management

### 1. Environment Variables
```go
// Use internal/config package
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    License  LicenseConfig
}

func Load() (*Config, error) {
    return &Config{
        Server: ServerConfig{
            Port: getEnvOrDefault("PORT", "8080"),
            Host: getEnvOrDefault("HOST", "0.0.0.0"),
        },
        Database: DatabaseConfig{
            URL: getEnvRequired("DATABASE_URL"),
        },
    }, nil
}
```

### 2. Configuration Files
```json
// config/production.json
{
    "server": {
        "port": 8080,
        "read_timeout": "30s",
        "write_timeout": "30s"
    },
    "license": {
        "check_interval": "1h",
        "grace_period": "7d"
    }
}
```

### 3. Secrets Management
```go
// Always encrypt sensitive data
encrypted, err := security.Encrypt([]byte(apiKey))
if err != nil {
    return fmt.Errorf("encrypt api key: %w", err)
}

// Store encrypted in files
if err := os.WriteFile("credentials.dat", encrypted, 0600); err != nil {
    return fmt.Errorf("save credentials: %w", err)
}
```

## Integration Patterns

### 1. External APIs
```go
// Use circuit breaker for resilience
client := resty.New().
    SetTimeout(30 * time.Second).
    SetRetryCount(3).
    OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
        r.SetContext(ctx)
        r.SetHeader("X-Trace-ID", middleware.GetTraceID(ctx))
        return nil
    })
```

### 2. Database Patterns
```go
// Use transactions for consistency
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin transaction: %w", err)
}
defer tx.Rollback()

// Operations...

if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit transaction: %w", err)
}
```

### 3. Message Queue Patterns
```go
// Use channels for internal communication
type JobQueue struct {
    jobs chan Job
    done chan struct{}
}

func (q *JobQueue) Process(ctx context.Context, workers int) {
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            q.worker(ctx)
        }()
    }

    wg.Wait()
}
```

## Performance Guidelines

### 1. Concurrent Processing
```go
// Use worker pools for CPU-bound tasks
func ProcessReports(ctx context.Context, reports []Report) error {
    const workers = 10
    jobs := make(chan Report, len(reports))
    results := make(chan error, len(reports))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for report := range jobs {
                results <- processReport(ctx, report)
            }
        }()
    }

    // Send jobs
    for _, report := range reports {
        jobs <- report
    }
    close(jobs)

    // Wait and collect results
    wg.Wait()
    close(results)

    for err := range results {
        if err != nil {
            return err
        }
    }

    return nil
}
```

### 2. Memory Management
```go
// Use sync.Pool for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    // Use buffer...
}
```

### 3. Caching Strategies
```go
// Use in-memory cache with TTL
type Cache struct {
    data sync.Map
    ttl  time.Duration
}

func (c *Cache) Get(key string) (interface{}, bool) {
    if val, ok := c.data.Load(key); ok {
        item := val.(*cacheItem)
        if time.Now().Before(item.expiry) {
            return item.value, true
        }
        c.data.Delete(key)
    }
    return nil, false
}
```

## Security Considerations

### 1. Authentication & Authorization
```go
// Use middleware for auth checks
func RequireLicense(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !license.IsValid(r.Context()) {
            errors.HandleError(w, r, errors.ErrUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### 2. Input Validation
```go
// Validate all inputs
func ValidateLicenseKey(key string) error {
    if len(key) != 29 {
        return errors.NewValidationError("invalid length")
    }

    pattern := `^[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$`
    if !regexp.MustCompile(pattern).MatchString(key) {
        return errors.NewValidationError("invalid format")
    }

    return nil
}
```

### 3. Encryption
```go
// Use AES-256-GCM for encryption
func Encrypt(plaintext []byte) ([]byte, error) {
    key := deriveKey()
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
```

## Common Workflows

### 1. Adding a New Endpoint
```go
// 1. Define contract in pkg/contracts
type CreateReportRequest struct {
    Title string `json:"title" validate:"required"`
    Data  []byte `json:"data" validate:"required"`
}

// 2. Add service method
func (s *Service) CreateReport(ctx context.Context, req *CreateReportRequest) (*Report, error) {
    // Implementation
}

// 3. Add handler
func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
    var req CreateReportRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.HandleError(w, r, errors.NewValidationError("invalid request", err))
        return
    }

    report, err := h.service.CreateReport(r.Context(), &req)
    if err != nil {
        errors.HandleError(w, r, err)
        return
    }

    respond.JSON(w, http.StatusCreated, report)
}

// 4. Add route
r.Post("/api/v1/reports", handler.CreateReport)

// 5. Add tests
func TestCreateReport(t *testing.T) {
    // Test implementation
}
```

### 2. Adding a New Operation
```go
// 1. Define operation in internal/operations
type DataImportOperation struct {
    BaseOperation
}

func (op *DataImportOperation) Execute(ctx context.Context) error {
    // Implementation
}

// 2. Register operation
registry.Register("data_import", NewDataImportOperation)

// 3. Add to pipeline
pipeline.AddStage("import", &DataImportOperation{})
```

## Observability

### 1. Structured Logging
```go
// Use slog with consistent fields
logger := slog.With(
    "service", "report",
    "version", config.Version,
)

// Log with context
logger.InfoContext(ctx, "report generated",
    "report_id", report.ID,
    "duration", time.Since(start),
    "size_bytes", len(data),
)
```

### 2. Metrics
```go
// Use OpenTelemetry for metrics
meter := otel.Meter("isx-scrapper")

requestCounter, _ := meter.Int64Counter("http_requests_total",
    metric.WithDescription("Total HTTP requests"),
)

requestCounter.Add(ctx, 1,
    attribute.String("method", r.Method),
    attribute.String("path", r.URL.Path),
    attribute.Int("status", status),
)
```

### 3. Distributed Tracing
```go
// Use OpenTelemetry for tracing
tracer := otel.Tracer("isx-scrapper")

ctx, span := tracer.Start(ctx, "ProcessReport",
    trace.WithAttributes(
        attribute.String("report.id", reportID),
    ),
)
defer span.End()

// Add events
span.AddEvent("validation_completed")

// Record errors
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

## Monitoring & Alerting

### 1. Health Checks
```go
// Implement health endpoint
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    checks := map[string]string{
        "database": h.checkDatabase(r.Context()),
        "license":  h.checkLicense(r.Context()),
        "storage":  h.checkStorage(r.Context()),
    }

    status := http.StatusOK
    for _, check := range checks {
        if check != "ok" {
            status = http.StatusServiceUnavailable
            break
        }
    }

    respond.JSON(w, status, checks)
}
```

### 2. Metrics Endpoints
```go
// Expose Prometheus metrics
import "github.com/prometheus/client_golang/prometheus/promhttp"

r.Handle("/metrics", promhttp.Handler())
```

### 3. Alert Rules
```yaml
# Example Prometheus rules
groups:
  - name: isx-scrapper
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High error rate detected
```

## Documentation Standards

### 1. Code Documentation
```go
// Package reports handles report generation and management.
// It provides functionality for creating, storing, and retrieving
// financial reports from ISX data.
package reports

// GenerateReport creates a new report from the provided data.
// It validates the input, processes the data, and stores the result.
//
// Example:
//
//	report, err := svc.GenerateReport(ctx, data)
//	if err != nil {
//	    return fmt.Errorf("generate report: %w", err)
//	}
func (s *Service) GenerateReport(ctx context.Context, data []byte) (*Report, error) {
    // Implementation
}
```

### 2. API Documentation
```go
// Use OpenAPI annotations
// @Summary Create a new report
// @Description Creates a new financial report from uploaded data
// @Tags reports
// @Accept json
// @Produce json
// @Param request body CreateReportRequest true "Report data"
// @Success 201 {object} Report
// @Failure 400 {object} errors.ProblemDetails
// @Failure 401 {object} errors.ProblemDetails
// @Router /api/v1/reports [post]
func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### 3. README Files
Every package must have a README.md explaining:
- Purpose and responsibilities
- Key types and interfaces
- Usage examples
- Dependencies
- Testing instructions

## Code Review Checklist

### Before Submitting PR
- [ ] All tests pass with `go test ./... -race`
- [ ] Test coverage > 80% for new code
- [ ] No linting errors (`golangci-lint run`)
- [ ] Documentation updated (code comments, README)
- [ ] Error handling follows RFC 7807
- [ ] Logging includes trace_id and context
- [ ] No sensitive data in logs
- [ ] Security considerations addressed
- [ ] Performance impact considered
- [ ] Database migrations included if needed

### Review Focus Areas
1. **Security**: Auth, input validation, encryption
2. **Error Handling**: Proper error types, wrapping
3. **Testing**: Coverage, edge cases, benchmarks
4. **Performance**: Concurrency, memory usage
5. **Documentation**: Clear, complete, accurate
6. **React Hydration**: Check for unguarded Date/time operations, useHydration usage

## Version Management

### 1. Semantic Versioning
```go
// version/version.go
package version

var (
    Version   = "3.0.0"
    GitCommit = "unknown"
    BuildTime = "unknown"
)

// Set during build
// go build -ldflags "-X version.GitCommit=$(git rev-parse HEAD)"
```

### 2. API Versioning
```go
// Use URL path versioning
r.Route("/api/v1", func(r chi.Router) {
    r.Mount("/reports", reportHandler)
    r.Mount("/operations", operationHandler)
})

// Support multiple versions
r.Route("/api/v2", func(r chi.Router) {
    // V2 endpoints
})
```

### 3. Database Migrations
```sql
-- migrations/001_initial_schema.up.sql
CREATE TABLE reports (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- migrations/001_initial_schema.down.sql
DROP TABLE reports;
```

## Debugging & Troubleshooting

### 1. Debug Logging
```go
// Use debug level for detailed info
slog.DebugContext(ctx, "processing record",
    "record_id", record.ID,
    "raw_data", hex.EncodeToString(data),
)

// Enable in development
slog.SetLogLevel(slog.LevelDebug)
```

### 2. Profiling
```go
import _ "net/http/pprof"

// Add pprof endpoints
r.Mount("/debug", middleware.Profiler())

// CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

// Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap
```

### 3. Trace Analysis
```go
// Add trace context to all operations
ctx = trace.ContextWithSpan(ctx, span)

// Use trace ID in logs
traceID := span.SpanContext().TraceID().String()
logger.InfoContext(ctx, "operation started",
    "trace_id", traceID,
)
```

## Database/Storage Patterns

### 1. Connection Management
```go
// Use connection pooling
db, err := sql.Open("postgres", dsn)
if err != nil {
    return fmt.Errorf("open database: %w", err)
}

db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 2. Query Patterns
```go
// Use prepared statements
stmt, err := db.PrepareContext(ctx, `
    SELECT id, title, created_at
    FROM reports
    WHERE user_id = $1 AND status = $2
`)
if err != nil {
    return fmt.Errorf("prepare statement: %w", err)
}
defer stmt.Close()

// Use scanning helpers
var reports []Report
rows, err := stmt.QueryContext(ctx, userID, status)
if err != nil {
    return fmt.Errorf("query reports: %w", err)
}
defer rows.Close()

for rows.Next() {
    var r Report
    if err := rows.Scan(&r.ID, &r.Title, &r.CreatedAt); err != nil {
        return fmt.Errorf("scan report: %w", err)
    }
    reports = append(reports, r)
}
```

### 3. Migration Management
```go
// Use golang-migrate
import "github.com/golang-migrate/migrate/v4"

func RunMigrations(dbURL string) error {
    m, err := migrate.New(
        "file://migrations",
        dbURL,
    )
    if err != nil {
        return fmt.Errorf("create migrator: %w", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("run migrations: %w", err)
    }

    return nil
}
```

## Deployment Notes

### Production Build
1. Set credentials in `encrypted_credentials.dat`
2. Run `build.bat` for complete build
3. Frontend embedded in `web-licensed.exe`
4. All paths relative to executable

### Directory Structure
```
release/
├── web-licensed.exe    # Main server
├── scraper.exe        # ISX scraper
├── processor.exe      # Data processor
├── indexcsv.exe       # Index extractor
├── data/
│   ├── downloads/     # Excel files
│   └── reports/       # CSV outputs
├── logs/              # Application logs
└── web/               # Static assets (backup)
```

### Configuration Files
- `credentials.json`: Google Sheets API
- `sheets-config.json`: Sheet ID mappings
- `license.dat`: Activated license data

## React Hydration Best Practices

When developing React components with Next.js SSR, follow these patterns to prevent hydration errors (#418, #423):

### 1. Use Hydration State Management

The project provides a reusable hydration hook in `@/lib/hooks`:

```typescript
import { useHydration } from '@/lib/hooks'

function MyComponent() {
  const isHydrated = useHydration()

  if (!isHydrated) {
    return <LoadingState />
  }

  // Client-only content here
}
```

Or implement manually:
```typescript
const [isHydrated, setIsHydrated] = useState(false)
useEffect(() => {
  setIsHydrated(true)
}, [])
```

### 2. Guard Dynamic Operations
- **Date operations**: `isHydrated ? new Date().toISOString() : ''`
- **WebSocket updates**: Add `if (!isHydrated) return` at the start of effects
- **API calls**: Delay data fetching until after hydration
- **Dynamic content**: Show consistent loading state until hydrated

### 3. Pre-Hydration Loading State
```typescript
// Early return with loading state
if (!isHydrated) {
  return (
    <div className="min-h-screen p-8 flex items-center justify-center">
      <div className="text-center">
        <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
        <p className="text-muted-foreground">Initializing...</p>
      </div>
    </div>
  )
}
```

### 4. Include Dependencies
Always include `isHydrated` in dependency arrays when used inside callbacks:
```typescript
const handleOperation = useCallback((data) => {
  const date = isHydrated ? new Date().toISOString() : ''
  // ... rest of logic
}, [isHydrated])  // Include in dependencies
```

### 5. Common Patterns to Avoid
- Don't use `Date.now()` or `new Date()` without hydration guards
- Don't fetch data in component body - use `useEffect` with hydration check
- Don't render different content based on `typeof window !== 'undefined'`
- Don't update state from WebSocket before hydration completes

### 6. Testing for Hydration Issues
1. Build with `./build.bat`
2. Clear browser cache and cookies
3. Open browser console before loading page
4. Look for React errors #418 (hydration mismatch) or #423 (text content mismatch)
5. Check for "Warning: Text content did not match" messages

## Important Notes

- **🚨 BUILD RULES**: NEVER build in api/ or web/ - ALWAYS use ./build.bat from root (see docs/development/BUILD_RULES.md)
- **🚨 NO DEV BUILDS**: Run `./scripts/verify-no-dev-builds.bat` to check compliance
- **🚨 LOGS CLEARED**: Every build via ./build.bat automatically clears all logs
- **No Blind Sleeps**: Use context, channels, or timers
- **Test Before Push**: All tests must pass with race detector
- **Update Docs**: Keep README files current - docs in same PR as code
- **Path Resolution**: Always use `paths.GetBaseDir()`
- **WebSocket Reliability**: Use `internal/websocket` hub + pumps only (no manager/metrics), forward snapshots through `BroadcastUpdate`, keep reconnection logic client-side
- **License Checks**: Validate before operations
- **Chi Only**: Use Chi v5 for all HTTP routing - no other frameworks
- **slog Only**: Use slog for all logging - no fmt.Println or log.Printf
- **RFC 7807**: All API errors must follow RFC 7807 Problem Details
- **TDD Always**: Write tests first, then implementation
- **Context Everywhere**: Pass context.Context as first parameter
- **Structured Logging**: Include trace_id, operation, and relevant fields
- **Interface First**: Define interfaces before implementations
- **No Global State**: Use dependency injection
- **Clean Resources**: Always defer cleanup
- **Benchmark Critical Paths**: Performance matters
- **Project Structure**: See `docs/development/FILE_INDEX.md` for complete directory organization
- **Command Line**: We are using bash now windows CMD in implementing our commands
- **React Hydration**: Use `useHydration` hook from `@/lib/hooks` for client-only content
- **Hydration Testing**: Always test production builds for React errors #418 and #423
- **Build Location**: NEVER build in `api/` or `web/` directories - always use `./build.bat` to output to `dist/`
- **Frontend Embedding**: Use explicit patterns like `//go:embed frontend/*.html frontend/_next`
- **Empty Directories**: Build process removes empty dirs that Next.js creates
- **Embed Standards**: Follow Grafana/CockroachDB pattern for production embedding
- **Validation Required**: Frontend build must pass validation before embedding
- **Interactive Guide**: In-app educational system at `/guide` with 10 sections, interactive demos, and progress tracking (see Section 29)
- **Guide Architecture**: Use server/client component split with localStorage for progress, see `docs/features/INTERACTIVE_GUIDE.md`
- **Guide Priority**: Phase 0 in Master Plan - user education foundation before advanced features

## Project Structure

The project follows a clean, professional Go structure. For detailed file organization and descriptions, see `docs/development/FILE_INDEX.md`.

### Key Directories
- `api/` - Go backend source code
- `web/` - Next.js frontend source code
- `docs/` - All documentation
- `tools/` - Development utilities and scripts
- `installer/` - Windows installer configuration
- Root directory contains only essential files (12 files)

### Build System
- Single `build.go` file (577 lines) handles all build operations
- Simple `build.bat` wrapper for Windows users
- Targets: all, web, frontend, scraper, processor, indexcsv, test, clean, release

For complete details about file organization, removed files, and structural improvements, refer to `FILE_INDEX.md`.

## Interactive Guide Feature

The Interactive Guide is an in-app educational feature accessible at `/guide` that provides comprehensive tutorials, demonstrations, and documentation for all ISX Pulse features. It serves as both a teaching tool for new users and a quick reference for experienced users.

### Overview

**Purpose**: Provide interactive, step-by-step guidance for all ISX Pulse features with live demos, code examples, and visual explanations.

**Location**: `/guide` route with entry in main navigation bar

**Goals**:
- Reduce learning curve for new users
- Provide context-sensitive help
- Demonstrate complex workflows interactively
- Serve as living documentation
- Track user progress through tutorials

### Architecture

```
web/
├── app/
│   └── guide/
│       ├── page.tsx                    # Server component with SEO metadata
│       ├── guide-client.tsx            # Main client component
│       ├── components/
│       │   ├── GuideNavigation.tsx     # Section navigation sidebar
│       │   ├── GuideSection.tsx        # Reusable section wrapper
│       │   ├── ProgressTracker.tsx     # Progress visualization
│       │   ├── InteractiveDemo.tsx     # Demo container component
│       │   └── CodeExample.tsx         # Syntax-highlighted code blocks
│       └── sections/
│           ├── WelcomeSection.tsx      # Introduction and overview
│           ├── LicenseSection.tsx      # License system guide
│           ├── PipelineSection.tsx     # Pipeline architecture
│           ├── OperationsSection.tsx   # Operations guide
│           ├── ReportsSection.tsx      # Reports system
│           ├── ChartsBasicsSection.tsx # Chart fundamentals
│           ├── MarketOverviewSection.tsx # Market treemap
│           └── QuickReferenceSection.tsx # Cheat sheet
├── components/
│   └── guide/
│       ├── demos/
│       │   ├── ScratchCardDemo.tsx     # Interactive license demo
│       │   ├── OperationSimulator.tsx  # Live operation demo
│       │   ├── LiveChartDemo.tsx       # Chart interaction demo
│       │   └── PipelineFlowchart.tsx   # Visual pipeline flow
│       └── interactive/
│           ├── InteractiveChart.tsx    # Embeddable chart
│           ├── CodePlayground.tsx      # Live code editor
│           └── QuizComponent.tsx       # Knowledge checks
└── lib/
    └── guide/
        ├── progress.ts                 # Progress tracking logic
        ├── sections.ts                 # Section definitions
        └── content.ts                  # Content management
```

### Guide Sections

#### 1. Welcome Section
**Purpose**: Introduction to ISX Pulse and guide navigation

**Content**:
- Platform overview and key features
- Navigation instructions
- Progress tracking explanation
- Quick start checklist
- Interactive tour button

**Interactive Elements**:
- Animated feature showcase
- Platform statistics (version, uptime, features count)
- Quick navigation cards to popular sections

#### 2. License System Guide
**Purpose**: Explain licensing model and activation process

**Content**:
- License types and features
- Activation workflow
- Scratch card system explanation
- Reactivation process
- Troubleshooting guide

**Interactive Elements**:
- **Scratch Card Demo**: Interactive simulation of scratch card activation
- **License Status Checker**: Real-time license validation demo
- **Countdown Timer Demo**: Simulated 5-second countdown with controls

**Code Examples**:
```typescript
// Example: License API integration
const licenseData = await apiClient.getLicenseStatus()
if (licenseData.license_status === 'active') {
  // Proceed with operations
}
```

#### 3. Pipeline Architecture
**Purpose**: Explain the 5-stage data processing pipeline

**Content**:
- Pipeline overview (Scrape → Process → Index → Liquidity → Analysis)
- Stage-by-stage breakdown
- Data flow visualization
- Error handling and rollback
- Concurrent execution patterns

**Interactive Elements**:
- **Animated Pipeline Flowchart**: Visual representation with stage transitions
- **Data Flow Simulator**: Follow sample data through pipeline
- **Error Scenario Demo**: Show rollback and retry logic

**Visual Aids**:
- Mermaid diagrams for each stage
- Status indicators (pending, running, completed, failed)
- Progress bars for multi-file operations

#### 4. Operations Guide
**Purpose**: Comprehensive guide to all 5 operation types

**Content**:
- **Scrape ISX**: Download daily reports from ISX website
- **Process Data**: Convert Excel to CSV with validation
- **Extract Index**: Extract ISX60 index data
- **Calculate Liquidity**: Calculate liquidity metrics
- **Complete Analysis**: Execute all stages automatically

**For Each Operation**:
- Purpose and use cases
- Prerequisites and configuration
- Step-by-step execution guide
- Expected outputs
- Common errors and solutions

**Interactive Elements**:
- **Operation Simulator**: Run operations with mock data
- **File Preview**: View Excel/CSV samples
- **Configuration Builder**: Interactive form for operation settings

#### 5. Reports System
**Purpose**: Guide to generated reports and data access

**Content**:
- Report types (daily, historical, index, liquidity)
- Report structure and fields
- Download and export options
- Data integrity verification
- Google Sheets integration

**Interactive Elements**:
- **Report Explorer**: Browse sample reports
- **CSV Viewer**: Interactive table with sorting/filtering
- **Data Visualization**: Charts from report data

#### 6. Charts Basics
**Purpose**: Introduction to TradingView Lightweight Charts

**Content**:
- Chart types (candlestick, line, area, bar)
- Timeframes and navigation
- Zoom and pan controls
- Crosshair usage
- Price scales and formatting

**Interactive Elements**:
- **Live Chart Demo**: Real chart with ISX data
- **Control Tutorial**: Interactive walkthrough of chart controls
- **Timeframe Switcher**: Switch between 1D, 1W, 1M views

**Visual Examples**:
- Side-by-side chart type comparisons
- Annotated screenshots with control labels

#### 7. Market Overview
**Purpose**: Interactive treemap visualization guide

**Content**:
- Treemap layout and sizing
- Color coding (gainers/losers)
- Market cap vs volume modes
- Sector grouping
- Drill-down navigation

**Interactive Elements**:
- **Live Treemap Demo**: Interactive ISX treemap
- **Sector Explorer**: Filter by sector
- **Performance Heatmap**: Color-coded performance view

#### 8. Quick Reference
**Purpose**: Cheat sheet and keyboard shortcuts

**Content**:
- Keyboard shortcuts table
- API endpoints reference
- Common workflows checklist
- Troubleshooting decision tree

**Format**:
- Searchable tables
- Printable PDF export
- Copy-to-clipboard buttons

### Interactive Demos

#### Scratch Card Demo
**Component**: `ScratchCardDemo.tsx`

**Features**:
- Canvas-based scratch-off effect
- Hidden license key reveal
- Copy-to-clipboard integration
- Reset button for replay

**Implementation**:
```typescript
'use client'

import { useState, useRef } from 'react'
import { Card } from '@/components/ui/card'

export function ScratchCardDemo() {
  const [revealed, setRevealed] = useState(false)
  const canvasRef = useRef<HTMLCanvasElement>(null)

  // Mouse scratch implementation
  // Canvas drawing logic
  // Reveal threshold detection
}
```

#### Operation Simulator
**Component**: `OperationSimulator.tsx`

**Features**:
- Mock operation execution
- Real-time progress simulation
- Success/error scenarios
- Log output display

#### Live Chart Demo
**Component**: `LiveChartDemo.tsx`

**Features**:
- Embedded TradingView chart
- Volume toggle controls
- Timeframe selector
- Annotation tools

#### Pipeline Flowchart
**Component**: `PipelineFlowchart.tsx`

**Features**:
- Animated stage transitions
- Status indicators
- Click-to-explore stages
- Error path visualization

### Progress Tracking

**Storage**: localStorage with structured schema

**Schema**:
```typescript
interface GuideProgress {
  sectionsCompleted: string[]      // Section IDs
  lastVisited: string               // Section ID
  startedAt: string                 // ISO timestamp
  completedDemos: string[]          // Demo IDs
  quizScores: Record<string, number> // Section ID → score
  bookmarks: string[]               // Section IDs
}
```

**Features**:
- Automatic section completion detection
- Resume from last position
- Progress percentage calculation
- Completion badges
- Export/import progress

**Implementation**:
```typescript
// lib/guide/progress.ts
export function useGuideProgress() {
  const [progress, setProgress] = useState<GuideProgress>()

  const markSectionComplete = (sectionId: string) => {
    // Update localStorage
    // Trigger celebration animation
  }

  const getProgressPercentage = () => {
    return (progress.sectionsCompleted.length / totalSections) * 100
  }
}
```

### Responsive Design

**Desktop (≥1024px)**:
- Sidebar navigation (fixed, 280px width)
- Main content area (flexible)
- Interactive demos (full-width)
- Code examples (syntax-highlighted)

**Tablet (768px - 1023px)**:
- Collapsible sidebar
- Hamburger menu toggle
- Stacked layout for demos
- Scrollable code blocks

**Mobile (<768px)**:
- Bottom navigation sheet
- Full-screen sections
- Touch-optimized demos
- Simplified code examples

### Technical Implementation

**Tech Stack**:
- **Next.js 14**: App Router with server/client components
- **TypeScript**: Full type safety
- **Tailwind CSS**: Responsive styling
- **Shadcn/ui**: Consistent component library
- **Framer Motion**: Smooth animations and transitions
- **React Markdown**: Rich content rendering
- **React Syntax Highlighter**: Code examples
- **TradingView Lightweight Charts**: Chart demos

**Dependencies to Add**:
```json
{
  "framer-motion": "^11.0.0",
  "react-markdown": "^9.0.0",
  "react-syntax-highlighter": "^15.5.0",
  "remark-gfm": "^4.0.0"
}
```

**Server Component** (SEO metadata):
```typescript
// app/guide/page.tsx
import GuideClient from './guide-client'

export const metadata = {
  title: 'Interactive Guide - ISX Pulse',
  description: 'Comprehensive guide and tutorials for ISX Pulse features including market analysis, data operations, and system usage.',
  keywords: 'ISX guide, stock trading tutorial, market analysis, chart analysis, data pipeline',
  robots: { index: true, follow: true }
}

export default function GuidePage() {
  return <GuideClient />
}
```

**Client Component** (interactivity):
```typescript
// app/guide/guide-client.tsx
'use client'

import { useState } from 'react'
import { GuideNavigation } from './components/GuideNavigation'
import { WelcomeSection } from './sections/WelcomeSection'
import { useGuideProgress } from '@/lib/guide/progress'

export default function GuideClient() {
  const [currentSection, setCurrentSection] = useState('welcome')
  const { progress, markComplete } = useGuideProgress()

  return (
    <div className="flex min-h-screen">
      <GuideNavigation
        currentSection={currentSection}
        onSectionChange={setCurrentSection}
        progress={progress}
      />
      <main className="flex-1 p-8">
        {renderSection(currentSection)}
      </main>
    </div>
  )
}
```

### Adding New Sections

**Step-by-step Process**:

1. **Create Section Component**:
```typescript
// app/guide/sections/NewSection.tsx
'use client'

import { GuideSection } from '../components/GuideSection'
import { InteractiveDemo } from '../components/InteractiveDemo'

export function NewSection() {
  return (
    <GuideSection
      id="new-section"
      title="New Feature Guide"
      description="Learn about our newest feature"
    >
      <div className="space-y-6">
        <h2>Overview</h2>
        <p>Content here...</p>

        <InteractiveDemo
          title="Try it yourself"
          description="Interactive demonstration"
        >
          {/* Demo content */}
        </InteractiveDemo>
      </div>
    </GuideSection>
  )
}
```

2. **Register in Section Definitions**:
```typescript
// lib/guide/sections.ts
export const guideSections = [
  // ... existing sections
  {
    id: 'new-section',
    title: 'New Feature',
    icon: Sparkles,
    category: 'features',
    order: 9
  }
]
```

3. **Add to Navigation**:
```typescript
// app/guide/components/GuideNavigation.tsx
import { NewSection } from '../sections/NewSection'

const sectionComponents = {
  // ... existing sections
  'new-section': NewSection
}
```

4. **Update Progress Tracking**:
```typescript
// lib/guide/progress.ts
const totalSections = 9 // Increment count
```

### Content Strategy

**Writing Style**:
- Conversational and encouraging tone
- Progressive disclosure (simple → complex)
- Real-world examples from ISX trading
- Visual-first explanations
- Actionable takeaways

**Content Structure**:
- **Overview** (2-3 sentences)
- **Key Concepts** (bullet points)
- **Step-by-Step Tutorial** (numbered)
- **Interactive Demo** (hands-on)
- **Code Example** (if applicable)
- **Tips & Best Practices** (callout boxes)
- **Common Pitfalls** (warning boxes)
- **Next Steps** (navigation hints)

**Visual Hierarchy**:
- H1: Section title
- H2: Major subsections
- H3: Specific topics
- Callouts: Tips, warnings, examples
- Code blocks: Syntax-highlighted
- Screenshots: Annotated with arrows/labels

### SEO & Metadata

**Page-Level SEO**:
```typescript
export const metadata = {
  title: 'Interactive Guide - ISX Pulse',
  description: 'Learn ISX Pulse features with interactive tutorials covering market analysis, chart visualization, and data operations.',
  keywords: [
    'ISX trading guide',
    'market analysis tutorial',
    'stock market education',
    'TradingView charts guide',
    'data pipeline tutorial'
  ],
  openGraph: {
    title: 'ISX Pulse Interactive Guide',
    description: 'Master market analysis with hands-on tutorials',
    images: ['/og-guide.png'],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'ISX Pulse Guide',
    description: 'Interactive tutorials for stock market analysis'
  }
}
```

**Section-Level Structure**:
```typescript
// Add structured data for tutorial content
const structuredData = {
  "@context": "https://schema.org",
  "@type": "LearningResource",
  "name": "ISX Pulse Market Analysis Guide",
  "description": "Comprehensive tutorial on market analysis and data operations",
  "educationalLevel": "Beginner to Advanced",
  "timeRequired": "PT30M"
}
```

### Context-Sensitive Help

**Deep Linking**:
- Support URL parameters: `/guide?section=operations&demo=scraper`
- Direct linking from app pages
- Bookmark-friendly URLs

**Implementation**:
```typescript
// app/analysis/page.tsx - Link to operations guide
<Button
  variant="ghost"
  onClick={() => router.push('/guide?section=operations&operation=scraper')}
>
  <HelpCircle className="h-4 w-4 mr-2" />
  Learn about Scraping
</Button>
```

**Contextual Tooltips**:
- Inline help links throughout app
- "Learn more" buttons that open guide in modal
- Breadcrumb navigation from guide back to feature

### Maintenance Guidelines

**Content Updates**:
- Review guide quarterly for accuracy
- Update screenshots when UI changes
- Add new sections for new features
- Archive deprecated content

**Code Maintenance**:
- Keep demo components in sync with actual features
- Update code examples with version changes
- Test all interactive demos before releases

**Performance**:
- Lazy load heavy components (charts, demos)
- Optimize images and videos
- Code-split by section
- Cache static content

### Success Metrics

**Engagement Metrics**:
- Section completion rate
- Average time per section
- Most/least visited sections
- Demo interaction rate
- Search queries (if search added)

**Outcome Metrics**:
- Reduction in support tickets
- Feature adoption rates
- User satisfaction scores
- App engagement after guide visit

**Target KPIs**:
- 70% of new users visit guide
- 50% complete rate for core sections
- 80% satisfaction rating
- <5min average time to find answers

### Testing Checklist

**Before Deployment**:
- [ ] All links work correctly
- [ ] Interactive demos function properly
- [ ] Code examples are correct and tested
- [ ] Progress tracking saves/loads correctly
- [ ] Responsive design works on all screen sizes
- [ ] Navigation is intuitive
- [ ] Search (if implemented) returns relevant results
- [ ] Accessibility: keyboard navigation, screen readers
- [ ] Performance: page load <3s, smooth animations
- [ ] Cross-browser compatibility (Chrome, Firefox, Safari, Edge)

### Future Enhancements

**Phase 2 Features**:
- Full-text search across all sections
- Video tutorials (embedded YouTube/Vimeo)
- Downloadable PDF guides
- Interactive quizzes with scoring
- User annotations and notes
- Community tips and suggestions
- Multi-language support (Arabic, Kurdish)
- AI-powered chatbot for Q&A
- Personalized learning paths based on user role
- Integration with onboarding flow

**Advanced Demos**:
- Live data visualization with real-time updates
- Custom report builder with visual programming
- Risk calculator with scenario analysis

---

**Related Documentation**:
- See [Navigation Component](../web/components/app-content-client.tsx) for adding Guide to navbar
- See [Next.js Component Architecture](#nextjs-component-architecture) for server/client patterns
- See [React Hydration Best Practices](#react-hydration-best-practices) for interactive components
- See [TradingView Lightweight Charts](#tradingview-lightweight-charts---native-panes-api) for chart demos