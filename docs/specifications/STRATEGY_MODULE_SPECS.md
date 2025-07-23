# ISX Strategy Module Technical Specifications

## 1. Strategy Intermediate Representation (IR) Schema

### 1.1 Complete JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "version", "capital_allocation", "entries", "exits"],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 100
    },
    "version": {
      "type": "integer",
      "minimum": 1
    },
    "capital_allocation": {
      "type": "integer",
      "minimum": 1000000,
      "description": "Capital in IQD"
    },
    "liquidity_multiple": {
      "type": "number",
      "minimum": 1,
      "default": 5
    },
    "min_liquidity_score": {
      "type": "integer",
      "minimum": 0,
      "maximum": 100,
      "default": 40
    },
    "universe": {
      "type": "object",
      "properties": {
        "symbols": {
          "type": "array",
          "items": {"type": "string"}
        },
        "filter": {
          "type": "object",
          "properties": {
            "minAvgVolume": {"type": "integer"},
            "minPrice": {"type": "number"},
            "maxPrice": {"type": "number"},
            "sectors": {
              "type": "array",
              "items": {"type": "string"}
            }
          }
        }
      }
    },
    "entries": {
      "type": "array",
      "minItems": 1,
      "items": {
        "$ref": "#/definitions/entry_rule"
      }
    },
    "exits": {
      "type": "array",
      "minItems": 1,
      "items": {
        "$ref": "#/definitions/exit_rule"
      }
    },
    "risk": {
      "$ref": "#/definitions/risk_management"
    },
    "parameters": {
      "type": "object",
      "additionalProperties": {
        "$ref": "#/definitions/parameter"
      }
    }
  }
}
```

### 1.2 Condition Grammar

```
condition := comparison | logical_combination
comparison := indicator_ref operator value
logical_combination := {"all": [condition, ...]} | {"any": [condition, ...]} | {"not": condition}
indicator_ref := {"indicator": string, "params": object, "shift": integer}
operator := "<" | ">" | "<=" | ">=" | "==" | "!="
```

## 2. Indicator Specifications

### 2.1 Standard Indicators

| Indicator | Parameters | Output | Min Bars |
|-----------|------------|--------|----------|
| SMA | period: int | price | period |
| EMA | period: int | price | period * 2 |
| RSI | period: int | 0-100 | period + 1 |
| MACD | fast: int, slow: int, signal: int | macd, signal, histogram | slow + signal |
| ATR | period: int | price | period + 1 |
| BB | period: int, stddev: float | upper, middle, lower | period |
| STOCH | k_period: int, d_period: int | %K, %D | k_period + d_period |

### 2.2 ISX-Specific Indicators

#### 2.2.1 Liquidity Score Integration
```go
type LiquidityScoreIndicator struct {
    WindowDays int `json:"window_days" default:"60"`
}

// Returns normalized 0-100 score
func (l *LiquidityScoreIndicator) Compute(data *PriceSeries) *Series {
    // Uses hybrid liquidity calculation from liquidity module
}
```

#### 2.2.2 Zero Volume Streak
```go
type ZeroVolumeStreakIndicator struct{}

// Returns consecutive days of zero volume
func (z *ZeroVolumeStreakIndicator) Compute(data *PriceSeries) *Series {
    streak := 0
    for i := range data.Volume {
        if data.Volume[i] == 0 {
            streak++
        } else {
            streak = 0
        }
        result[i] = streak
    }
}
```

## 3. Backtest Engine Implementation

### 3.1 Main Loop Pseudocode

```go
func RunBacktest(strategy *Strategy, config *BacktestConfig) *BacktestResult {
    // Initialize
    portfolio := NewPortfolio(strategy.CapitalAllocation)
    indicators := PrecomputeIndicators(strategy, config.Universe, config.DateRange)
    liquidityCache := LoadLiquidityCache(config.Universe, config.DateRange)
    
    // Main loop
    for date := range config.TradingDays() {
        // Update market data
        prices := GetPrices(date)
        liquidity := liquidityCache[date]
        
        // Check exits first
        for position := range portfolio.OpenPositions {
            if ShouldExit(position, strategy.Exits, indicators[date]) {
                order := CreateExitOrder(position)
                fill := SimulateFill(order, prices[date], liquidity[date])
                portfolio.ExecuteTrade(fill)
            }
        }
        
        // Check entries
        eligibleSymbols := FilterByLiquidity(config.Universe, liquidity, strategy)
        for symbol := range eligibleSymbols {
            if ShouldEnter(symbol, strategy.Entries, indicators[date]) {
                order := CreateEntryOrder(symbol, portfolio, strategy)
                
                // Apply liquidity constraints
                order = ApplyLiquidityConstraints(order, liquidity[symbol])
                if order.Size > 0 {
                    fill := SimulateFill(order, prices[date], liquidity[date])
                    portfolio.ExecuteTrade(fill)
                }
            }
        }
        
        // Record daily metrics
        RecordEquityPoint(portfolio, date)
    }
    
    return CalculateMetrics(portfolio)
}
```

### 3.2 Liquidity Constraint Implementation

```go
func ApplyLiquidityConstraints(order *Order, liquidity *LiquidityData) *Order {
    proposedValue := order.Size * order.Price
    avgDailyValue := liquidity.AvgTradedValue30D
    
    liquidityRatio := avgDailyValue / proposedValue
    
    if liquidityRatio < order.Strategy.LiquidityMultiple {
        // Scale down position
        maxValue := avgDailyValue / order.Strategy.LiquidityMultiple
        scaledSize := int(maxValue / order.Price)
        
        if scaledSize < MIN_LOT_SIZE {
            // Reject order
            order.Size = 0
            order.RejectionReason = "Insufficient liquidity"
        } else {
            order.Size = scaledSize
            order.ScalingFactor = float64(scaledSize) / float64(order.OriginalSize)
        }
    }
    
    order.LiquidityRatio = liquidityRatio
    return order
}
```

### 3.3 Fill Simulation

```go
func SimulateFill(order *Order, dayData *DayData, liquidity *LiquidityData) *Fill {
    fill := &Fill{
        Order: order,
        Date:  dayData.Date,
    }
    
    // Determine fill price based on order type
    switch order.Type {
    case "MKT":
        fill.Price = dayData.Open
    case "LIMIT":
        if order.Action == "BUY" && order.LimitPrice >= dayData.Low {
            fill.Price = min(order.LimitPrice, dayData.Open)
        } else if order.Action == "SELL" && order.LimitPrice <= dayData.High {
            fill.Price = max(order.LimitPrice, dayData.Open)
        } else {
            fill.Status = "UNFILLED"
            return fill
        }
    }
    
    // Apply slippage
    slippage := CalculateSlippage(order, dayData, liquidity)
    if order.Action == "BUY" {
        fill.Price *= (1 + slippage)
    } else {
        fill.Price *= (1 - slippage)
    }
    
    // Apply fees
    fill.Fees = fill.Price * fill.Size * FEE_RATE
    
    return fill
}
```

## 4. Optimization Specifications

### 4.1 Grid Search Algorithm

```go
func GridSearch(strategy *Strategy, bounds map[string]ParamBounds) []OptimizationResult {
    // Generate all combinations
    combinations := GenerateCombinations(bounds)
    
    // Limit combinations
    if len(combinations) > MAX_COMBINATIONS {
        combinations = SampleUniform(combinations, MAX_COMBINATIONS)
    }
    
    // Parallel execution
    results := make(chan OptimizationResult, len(combinations))
    workers := runtime.NumCPU()
    
    for i := 0; i < workers; i++ {
        go func() {
            for combo := range combinations {
                strategyVariant := ApplyParams(strategy, combo)
                backtest := RunBacktest(strategyVariant, config)
                results <- OptimizationResult{
                    Params:  combo,
                    Metrics: backtest.Metrics,
                }
            }
        }()
    }
    
    // Collect and rank results
    return RankByObjective(results)
}
```

### 4.2 Walk-Forward Specification

```go
type WalkForwardPlan struct {
    WindowLength   int // months
    StepSize       int // months  
    OptimizePeriod int // months (must be < WindowLength)
    TestPeriod     int // months (WindowLength - OptimizePeriod)
    StartDate      time.Time
    EndDate        time.Time
}

func WalkForwardAnalysis(strategy *Strategy, plan *WalkForwardPlan) *WalkForwardResult {
    var windows []WFWindow
    
    currentStart := plan.StartDate
    for currentStart.Add(plan.WindowLength).Before(plan.EndDate) {
        window := WFWindow{
            OptimizeStart: currentStart,
            OptimizeEnd:   currentStart.AddMonths(plan.OptimizePeriod),
            TestStart:     currentStart.AddMonths(plan.OptimizePeriod),
            TestEnd:       currentStart.AddMonths(plan.WindowLength),
        }
        
        // Optimize on IS period
        bestParams := Optimize(strategy, window.OptimizeStart, window.OptimizeEnd)
        
        // Test on OOS period
        oosResult := RunBacktest(ApplyParams(strategy, bestParams), 
                                window.TestStart, window.TestEnd)
        
        windows = append(windows, WFWindow{
            Window:    window,
            ISMetrics: bestParams.Metrics,
            OOSMetrics: oosResult.Metrics,
        })
        
        currentStart = currentStart.AddMonths(plan.StepSize)
    }
    
    return AggregateWFResults(windows)
}
```

## 5. Metrics Calculations

### 5.1 Core Metrics Implementation

```go
// Sharpe Ratio (annualized)
func SharpeRatio(returns []float64, riskFreeRate float64) float64 {
    excess := make([]float64, len(returns))
    for i, r := range returns {
        excess[i] = r - riskFreeRate/252 // Daily risk-free rate
    }
    
    mean := Mean(excess)
    std := StdDev(excess)
    
    return mean / std * math.Sqrt(252) // Annualized
}

// Maximum Drawdown
func MaxDrawdown(equity []float64) (float64, int, int) {
    peak := equity[0]
    maxDD := 0.0
    peakIdx := 0
    troughIdx := 0
    
    for i, value := range equity {
        if value > peak {
            peak = value
            peakIdx = i
        }
        
        drawdown := (peak - value) / peak
        if drawdown > maxDD {
            maxDD = drawdown
            troughIdx = i
        }
    }
    
    return maxDD, peakIdx, troughIdx
}

// Win Rate
func WinRate(trades []Trade) float64 {
    wins := 0
    for _, trade := range trades {
        if trade.PnL > 0 {
            wins++
        }
    }
    return float64(wins) / float64(len(trades))
}
```

### 5.2 Liquidity-Specific Metrics

```go
// Average Liquidity Ratio
func AvgLiquidityRatio(trades []Trade) float64 {
    sum := 0.0
    for _, trade := range trades {
        sum += trade.LiquidityRatio
    }
    return sum / float64(len(trades))
}

// Liquidity Constraint Impact
func LiquidityConstraintImpact(trades []Trade) map[string]float64 {
    full := 0
    scaled := 0
    rejected := 0
    
    totalScaling := 0.0
    
    for _, trade := range trades {
        switch trade.LiquidityStatus {
        case "FULL":
            full++
        case "SCALED":
            scaled++
            totalScaling += trade.ScalingFactor
        case "REJECTED":
            rejected++
        }
    }
    
    total := full + scaled + rejected
    
    return map[string]float64{
        "full_fill_rate":     float64(full) / float64(total),
        "scaled_fill_rate":   float64(scaled) / float64(total),
        "rejection_rate":     float64(rejected) / float64(total),
        "avg_scaling_factor": totalScaling / float64(scaled),
    }
}
```

## 6. API Endpoints

### 6.1 Strategy Management

```yaml
# Create/Update Strategy
POST /api/strategies
Request:
  Content-Type: application/json
  Body: Strategy IR JSON
Response:
  201 Created
  {
    "id": 123,
    "version": 1,
    "created_at": "2025-01-19T10:00:00Z"
  }

# List Strategies
GET /api/strategies?status=active&page=1&limit=20
Response:
  200 OK
  {
    "strategies": [...],
    "total": 45,
    "page": 1,
    "limit": 20
  }

# Get Strategy Details
GET /api/strategies/{id}
Response:
  200 OK
  {
    "id": 123,
    "name": "RSI Rebound",
    "version": 1,
    "ir": {...},
    "metrics": {...},
    "last_run": {...}
  }
```

### 6.2 Backtesting

```yaml
# Run Backtest
POST /api/backtests
Request:
  {
    "strategy_id": 123,
    "from_date": "2020-01-01",
    "to_date": "2024-12-31",
    "initial_capital": 25000000,
    "mode": "standard"
  }
Response:
  202 Accepted
  {
    "run_id": "abc123",
    "status": "running",
    "progress_url": "/api/backtests/abc123/progress"
  }

# Get Backtest Results
GET /api/backtests/{run_id}
Response:
  200 OK
  {
    "run_id": "abc123",
    "status": "completed",
    "metrics": {
      "total_return": 0.45,
      "cagr": 0.085,
      "sharpe_ratio": 1.2,
      "max_drawdown": 0.15,
      "win_rate": 0.58
    },
    "equity_curve": [...],
    "trades": [...]
  }
```

### 6.3 Optimization

```yaml
# Start Optimization
POST /api/optimize
Request:
  {
    "strategy_id": 123,
    "method": "grid",
    "parameters": {
      "RSI.period": {"min": 10, "max": 20, "step": 2}
    },
    "objective": "sharpe_ratio",
    "constraints": {
      "min_trades": 30,
      "max_drawdown": 0.25
    }
  }
Response:
  202 Accepted
  {
    "job_id": "opt-456",
    "estimated_combinations": 6,
    "status_url": "/api/optimize/opt-456"
  }
```

### 6.4 Recommendations

```yaml
# Get Daily Recommendations
GET /api/recommendations?date=2025-01-19&min_score=70
Response:
  200 OK
  {
    "date": "2025-01-19",
    "recommendations": [
      {
        "symbol": "BGUC",
        "action": "BUY",
        "strategy": "RSI Rebound",
        "score": 85,
        "proposed_size": 1000,
        "scaled_size": 800,
        "liquidity_ratio": 4.2,
        "rationale": "RSI(14)=28.5 < 30 AND VOL_MA_RATIO=1.8 > 1.5"
      }
    ]
  }

# Execute Recommendation Batch
POST /api/recommendations/evaluate
Request:
  {
    "date": "2025-01-19",
    "strategies": ["all"] // or specific IDs
  }
Response:
  200 OK
  {
    "evaluated": 15,
    "signals_generated": 3,
    "execution_time_ms": 245
  }
```

## 7. Run Hash Calculation

```go
func CalculateRunHash(strategy *Strategy, config *BacktestConfig) string {
    // Deterministic serialization
    data := map[string]interface{}{
        "strategy_ir":    strategy.IR,
        "strategy_params": strategy.Parameters,
        "from_date":      config.FromDate,
        "to_date":        config.ToDate,
        "capital":        config.Capital,
        "code_version":   VERSION,
        "data_checksum":  CalculateDataChecksum(config),
    }
    
    // Sort keys for consistency
    jsonBytes, _ := json.Marshal(SortedMap(data))
    
    // SHA256 hash
    hash := sha256.Sum256(jsonBytes)
    return hex.EncodeToString(hash[:])
}
```

## 8. Performance Requirements

| Operation | Target | Max |
|-----------|--------|-----|
| Single backtest (5 years, 20 symbols) | < 2s | 5s |
| Optimization (100 combinations) | < 30s | 60s |
| Daily recommendations (50 strategies) | < 5s | 10s |
| Strategy save/load | < 100ms | 200ms |
| Real-time indicator update | < 50ms | 100ms |

## 9. Error Codes

| Code | Message | Description |
|------|---------|-------------|
| STR001 | Invalid strategy IR | JSON schema validation failed |
| STR002 | Insufficient historical data | Lookback period exceeds available data |
| STR003 | Invalid parameter bounds | Optimization bounds are invalid |
| STR004 | Liquidity constraint violation | Position size exceeds liquidity limits |
| STR005 | Capital allocation exceeded | Strategy would exceed allocated capital |
| BKT001 | Backtest failed | Runtime error during backtest |
| BKT002 | Data integrity error | Missing or corrupt price data |
| OPT001 | Too many combinations | Exceeds MAX_COMBINATIONS limit |
| REC001 | No active strategies | No strategies enabled for recommendations |