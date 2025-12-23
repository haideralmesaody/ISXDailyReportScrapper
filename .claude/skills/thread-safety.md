---
name: thread-safety
type: domain-skill
expertise_level: expert
domain: go-concurrency-patterns
version: "1.0.0"
---

# Thread Safety Skill - Go Concurrency Patterns

## Purpose
Expert knowledge in Go concurrency patterns and thread-safe programming, specifically optimized for the ISX Pulse platform's high-performance financial data processing requirements.

## Core Concurrency Primitives

### 1. Mutex Patterns (Mutual Exclusion)

#### Read-Write Mutex for High Read Throughput
```go
type UniversalIndicatorService struct {
    indicatorsCache map[string]*IndicatorData
    csvIndex        map[string]int
    mutex          sync.RWMutex  // Multiple readers, single writer
    dataPath       string
}

// Safe read operation - allows concurrent readers
func (s *UniversalIndicatorService) GetIndicator(ticker, indicator, date string) (float64, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    if tickerData, exists := s.indicatorsCache[ticker]; exists {
        if value, exists := tickerData.Values[indicator][date]; exists {
            return value, nil
        }
    }

    return 0, fmt.Errorf("indicator not found")
}

// Safe write operation - exclusive access
func (s *UniversalIndicatorService) UpdateIndicators(newData map[string]*IndicatorData) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    for ticker, data := range newData {
        s.indicatorsCache[ticker] = data
    }

    return nil
}
```

#### Mutex with Timeout for Deadlock Prevention
```go
func (s *Service) SafeUpdateWithTimeout(data []byte) error {
    done := make(chan struct{})
    var err error

    go func() {
        s.mutex.Lock()
        defer s.mutex.Unlock()
        err = s.performUpdate(data)
        close(done)
    }()

    select {
    case <-done:
        return err
    case <-time.After(30 * time.Second):
        return fmt.Errorf("update timeout - possible deadlock")
    }
}
```

### 2. Channel-Based Coordination

#### Pipeline Pattern for Data Processing
```go
func ProcessIndicators(input <-chan PriceData, output chan<- IndicatorData) {
    for priceData := range input {
        indicators := calculateAllIndicators(priceData)

        select {
        case output <- indicators:
            // Data sent successfully
        case <-time.After(5 * time.Second):
            slog.Warn("indicator calculation timeout",
                "ticker", priceData.Ticker)
        }
    }
    close(output)
}
```

#### Fan-Out Pattern for Concurrent Processing
```go
func ProcessTickersConcurrently(tickers []string, numWorkers int) {
    input := make(chan string, len(tickers))
    results := make(chan IndicatorResult, len(tickers))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for ticker := range input {
                result := processTicker(ticker)
                results <- result
            }
        }(i)
    }

    // Send work
    for _, ticker := range tickers {
        input <- ticker
    }
    close(input)

    // Wait for completion
    wg.Wait()
    close(results)

    // Collect results
    for result := range results {
        storeResult(result)
    }
}
```

### 3. Atomic Operations

#### High-Performance Counters
```go
type ServiceMetrics struct {
    queryCount   int64  // Use atomic operations for high performance
    cacheHits    int64
    cacheMisses  int64
    errorCount   int64
}

func (m *ServiceMetrics) RecordQuery() {
    atomic.AddInt64(&m.queryCount, 1)
}

func (m *ServiceMetrics) RecordCacheHit() {
    atomic.AddInt64(&m.cacheHits, 1)
}

func (m *ServiceMetrics) GetCacheHitRate() float64 {
    hits := atomic.LoadInt64(&m.cacheHits)
    total := atomic.LoadInt64(&m.queryCount)

    if total == 0 {
        return 0
    }
    return float64(hits) / float64(total)
}
```

#### Atomic Configuration Updates
```go
type ConfigManager struct {
    config atomic.Value  // Stores *Config
}

func (c *ConfigManager) UpdateConfig(newConfig *Config) {
    c.config.Store(newConfig)
}

func (c *ConfigManager) GetConfig() *Config {
    if cfg := c.config.Load(); cfg != nil {
        return cfg.(*Config)
    }
    return &DefaultConfig
}
```

### 4. Sync.Map for High-Concurrency Caches

#### Thread-Safe Cache Implementation
```go
type HighConcurrencyCache struct {
    data sync.Map  // Thread-safe map for high-concurrency scenarios
    stats CacheStats
    mutex sync.RWMutex  // Only for stats updates
}

type CacheEntry struct {
    Value      interface{}
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

## ISX Pulse Thread Safety Patterns

### 1. Pipeline Stage Coordination

#### Thread-Safe Pipeline Execution
```go
type PipelineOrchestrator struct {
    stages []PipelineStage
    mutex  sync.RWMutex
    status PipelineStatus
}

func (p *PipelineOrchestrator) ExecuteStage(ctx context.Context, stageID string) error {
    p.mutex.Lock()

    // Check if stage can be executed
    if err := p.validateStageExecution(stageID); err != nil {
        p.mutex.Unlock()
        return err
    }

    // Update status
    p.status.StageInProgress = stageID
    p.mutex.Unlock()

    // Execute stage
    err := p.stages[stageID].Execute(ctx)

    // Update final status
    p.mutex.Lock()
    defer p.mutex.Unlock()

    if err != nil {
        p.status.LastError = err
        p.status.FailedStage = stageID
    } else {
        p.status.CompletedStages = append(p.status.CompletedStages, stageID)
    }

    return err
}
```

### 2. WebSocket Connection Management

#### Thread-Safe WebSocket Hub
```go
type WebSocketHub struct {
    clients    map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan []byte
    mutex      sync.RWMutex
}

func (h *WebSocketHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client] = true
            h.mutex.Unlock()

        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mutex.Unlock()

        case message := <-h.broadcast:
            h.mutex.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mutex.RUnlock()
        }
    }
}
```

### 3. CSV Index Management

#### Thread-Safe CSV Index
```go
type CSVIndex struct {
    PrimaryIndex   map[string]map[string]int    // ticker -> date -> row_offset
    SecondaryIndex map[string][]string          // date -> []tickers
    ReverseIndex  map[int]IndexEntry           // row_offset -> {ticker, date}
    mutex         sync.RWMutex
}

func (idx *CSVIndex) GetRow(ticker, date string) (IndexEntry, bool) {
    idx.mutex.RLock()
    defer idx.mutex.RUnlock()

    if tickerIndex, exists := idx.PrimaryIndex[ticker]; exists {
        if rowOffset, exists := tickerIndex[date]; exists {
            entry, exists := idx.ReverseIndex[rowOffset]
            return entry, exists
        }
    }

    return IndexEntry{}, false
}

func (idx *CSVIndex) UpdateIndex(newEntries []IndexEntry) error {
    idx.mutex.Lock()
    defer idx.mutex.Unlock()

    for _, entry := range newEntries {
        // Update primary index
        if idx.PrimaryIndex[entry.Ticker] == nil {
            idx.PrimaryIndex[entry.Ticker] = make(map[string]int)
        }
        idx.PrimaryIndex[entry.Ticker][entry.Date] = entry.Offset

        // Update secondary index
        idx.SecondaryIndex[entry.Date] = append(idx.SecondaryIndex[entry.Date], entry.Ticker)

        // Update reverse index
        idx.ReverseIndex[entry.Offset] = entry
    }

    return nil
}
```

## Performance Optimization Patterns

### 1. Connection Pooling

#### Database Connection Pool
```go
type ConnectionPool struct {
    connections chan *sql.DB
    factory     func() (*sql.DB, error)
    maxSize     int
    mutex       sync.Mutex
    created     int
}

func NewConnectionPool(maxSize int, factory func() (*sql.DB, error)) *ConnectionPool {
    pool := &ConnectionPool{
        connections: make(chan *sql.DB, maxSize),
        factory:     factory,
        maxSize:     maxSize,
    }

    // Pre-allocate some connections
    for i := 0; i < maxSize/2; i++ {
        if conn, err := factory(); err == nil {
            pool.connections <- conn
            pool.created++
        }
    }

    return pool
}

func (p *ConnectionPool) Get() (*sql.DB, error) {
    select {
    case conn := <-p.connections:
        return conn, nil
    default:
        // No available connections, create new one if under limit
        p.mutex.Lock()
        defer p.mutex.Unlock()

        if p.created < p.maxSize {
            p.created++
            return p.factory()
        }

        // Wait for available connection
        return <-p.connections, nil
    }
}

func (p *ConnectionPool) Put(conn *sql.DB) {
    select {
    case p.connections <- conn:
        // Connection returned to pool
    default:
        // Pool is full, discard connection
        conn.Close()
    }
}
```

### 2. Batch Processing with Goroutines

#### Parallel Indicator Calculations
```go
func (c *IndicatorCalculator) CalculateBatch(tickers []string) error {
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
            c.processTickerBatch(batch)
        }(batch)
    }

    wg.Wait()
    return nil
}
```

## Error Handling and Recovery

### 1. Panic Recovery in Goroutines
```go
func (s *Service) SafeGoroutine(fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                slog.Error("goroutine panic recovered",
                    "panic", r,
                    "stack", debug.Stack())
            }
        }()
        fn()
    }()
}
```

### 2. Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    maxFailures int
    resetTime   time.Duration
    failures    int
    lastFailure time.Time
    mutex       sync.RWMutex
    state       State
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.RLock()
    if cb.state == Open && time.Since(cb.lastFailure) < cb.resetTime {
        cb.mutex.RUnlock()
        return fmt.Errorf("circuit breaker is open")
    }
    cb.mutex.RUnlock()

    err := fn()

    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()

        if cb.failures >= cb.maxFailures {
            cb.state = Open
        }
    } else {
        cb.failures = 0
        cb.state = Closed
    }

    return err
}
```

## Best Practices

### DO ✅
1. **Use RWMutex for read-heavy workloads**: Allow concurrent readers
2. **Prefer channels over shared memory**: Safer communication patterns
3. **Use atomic operations for simple counters**: Better performance than mutex
4. **Always handle panics in goroutines**: Prevent crashes
5. **Use timeouts for blocking operations**: Prevent deadlocks
6. **Limit goroutine lifetimes**: Clean up resources properly

### DON'T ❌
1. **Don't share mutable state without synchronization**: Race conditions
2. **Don't use mutexes inside hot loops**: Performance bottleneck
3. **Don't create goroutines without cleanup plan**: Resource leaks
4. **Don't block forever in goroutines**: Deadlock potential
5. **Don't ignore return values from goroutines**: Lost errors
6. **Don't use global state without protection**: Race conditions

This skill provides comprehensive Go concurrency expertise for building thread-safe, high-performance financial data processing systems in ISX Pulse.