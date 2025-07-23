# ISX Strategy Module Design

## Executive Summary

The ISX Strategy Module is a comprehensive trading strategy subsystem that enables users to:
1. Define rule-based trading strategies using technical indicators
2. Backtest strategies on historical ISX data with realistic liquidity constraints
3. Optimize parameters using various techniques (grid search, walk-forward analysis)
4. Generate daily trading recommendations

This module integrates deeply with the existing liquidity scoring system to ensure strategies respect the unique characteristics of the ISX market, particularly the intermittent trading and liquidity constraints common in frontier markets.

## 1. Architecture Overview

### 1.1 High-Level Architecture

```
┌──────────────────┐        ┌──────────────────┐
│  Strategy UI     │◄──────►│  Strategy API    │
│  (HTML/JS)       │        │  (Go HTTP)       │
└──────────────────┘        └────────┬─────────┘
                                     │
                        ┌────────────┴────────────┐
                        │  Strategy Service (Go)  │
                        ├─────────────────────────┤
                        │ • Parser/Validator      │
                        │ • IR Store              │
                        │ • Indicator Registry    │
                        │ • Backtest Engine       │
                        │ • Optimization Engine   │
                        │ • Walk-Forward Module   │
                        │ • Recommendation Engine │
                        └────────────┬────────────┘
                                     │
                  ┌──────────────────┴──────────────────┐
                  │        Data & Persistence           │
                  │  SQLite (prices, strategies,        │
                  │  backtests, recommendations)       │
                  └─────────────────────────────────────┘
```

### 1.2 Integration Points

- **Price Data**: Uses existing daily price tables
- **Liquidity Scores**: Integrates with liquidity scoring system
- **Calendar**: Respects ISX trading calendar for accurate backtesting
- **UI**: Extends existing web interface with strategy management pages

## 2. Core Components

### 2.1 Strategy Definition Language

#### 2.1.1 JSON Intermediate Representation (IR)

```json
{
  "name": "RSI Volume Rebound",
  "version": 1,
  "capital_allocation": 25000000,  // IQD
  "liquidity_multiple": 5,
  "min_liquidity_score": 40,
  "universe": {
    "symbols": ["BNOI", "TASC"],
    "filter": {
      "minAvgVolume": 10000,
      "lookback": 60
    }
  },
  "entries": [
    {
      "condition": {
        "all": [
          {"indicator": "RSI", "params": {"period": 14}, "op": "<", "value": 30},
          {"indicator": "VOL_MA_RATIO", "params": {"maPeriod": 20}, "op": ">", "value": 1.5}
        ]
      },
      "order": {"type": "MKT", "size": {"pctEquity": 0.1}}
    }
  ],
  "exits": [
    {
      "condition": {
        "any": [
          {"indicator": "RSI", "params": {"period": 14}, "op": ">", "value": 55},
          {"indicator": "STOP_LOSS_PCT", "op": "<", "value": -0.08}
        ]
      },
      "action": "CLOSE"
    }
  ],
  "risk": {
    "maxOpenPositions": 5,
    "positionSizing": "FixedFractional"
  },
  "constraints": {
    "minTrades": 30,
    "maxLookback": 250
  },
  "parameters": {
    "RSI.period": {
      "type": "int",
      "range": [10, 20],
      "optimize": true
    }
  }
}
```

#### 2.1.2 DSL Alternative (Future)

```
STRATEGY "RSI_Rebound" CAPITAL 25_000_000 LIQ_MULT 5 MIN_LIQ_SCORE 40 {
    ENTRY: RSI(14) < 30 AND LIQ_SCORE >= 40;
    EXIT: RSI(14) > 55 OR TRAIL_STOP(ATR(14)*3);
    SIZING: RISK 1% VOL_TARGET 20d;
}
```

### 2.2 Indicator Framework

#### 2.2.1 Core Technical Indicators
- **Trend**: SMA, EMA, MACD, ADX
- **Momentum**: RSI, Stochastic, ROC
- **Volatility**: ATR, Bollinger Bands, Standard Deviation
- **Volume**: OBV, Volume MA, VWAP
- **ISX-Specific**: Liquidity Score, Zero Volume Streak, Trading Continuity

#### 2.2.2 Indicator Interface

```go
type Indicator interface {
    Name() string
    RequiredBars(params map[string]interface{}) int
    Compute(ctx context.Context, data *PriceSeries) (*Series, error)
}
```

### 2.3 Backtest Engine

#### 2.3.1 Execution Flow
1. Load historical data for date range
2. Pre-compute all indicators
3. For each trading day:
   - Evaluate exit conditions for open positions
   - Apply liquidity constraints
   - Evaluate entry conditions
   - Generate and simulate orders
   - Update portfolio state
   - Record metrics

#### 2.3.2 Liquidity-Aware Execution

**Liquidity Multiple Test**:
```
ProposedPositionValue = position_sizing(capital, risk, price)
AvgDailyTradedValue = SMA_30(Close * Volume)
LiquidityRatio = AvgDailyTradedValue / ProposedPositionValue

if LiquidityRatio < liquidity_multiple:
    ScaledSize = floor(AvgDailyTradedValue / liquidity_multiple / price)
    if ScaledSize < min_lot_size:
        reject_order()
```

**Slippage Model**:
```
if daily_volume < order_size * threshold:
    partial_fill = daily_volume * fill_ratio
    slippage = base_spread * (order_size / avg_daily_volume)
```

### 2.4 Optimization Engine

#### 2.4.1 Methods
- **Grid Search**: Exhaustive parameter combinations
- **Random Search**: Monte Carlo sampling
- **Walk-Forward**: Rolling window optimization

#### 2.4.2 Constraints
- Maximum combinations limit (default: 200)
- Minimum trades requirement
- Out-of-sample validation

### 2.5 Recommendation Engine

Daily evaluation process:
1. Load latest prices and liquidity scores
2. Evaluate all active strategies
3. Apply liquidity filters
4. Generate signals with scoring:
   ```
   score = expectancy_weight * historical_expectancy +
           recency_weight * recent_hit_rate +
           liquidity_weight * liquidity_score_normalized
   ```
5. Rank and persist recommendations

## 3. Data Model

### 3.1 Core Tables

```sql
-- Strategy definitions
CREATE TABLE strategies (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    version INTEGER NOT NULL,
    ir_json TEXT NOT NULL,
    capital_allocation BIGINT NOT NULL,
    liquidity_multiple REAL NOT NULL DEFAULT 5,
    min_liquidity_score INTEGER NOT NULL DEFAULT 40,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'inactive',
    UNIQUE(name, version)
);

-- Backtest runs
CREATE TABLE backtest_runs (
    id INTEGER PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id),
    run_hash TEXT UNIQUE NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    from_date DATE,
    to_date DATE,
    mode TEXT,
    params_json TEXT,
    metrics_json TEXT,
    walk_forward BOOLEAN DEFAULT FALSE,
    notes TEXT
);

-- Individual trades
CREATE TABLE trades (
    id INTEGER PRIMARY KEY,
    run_id INTEGER REFERENCES backtest_runs(id),
    symbol_id INTEGER,
    date DATE,
    action TEXT,
    quantity INTEGER,
    price REAL,
    slippage REAL,
    fees REAL,
    liquidity_ratio REAL,
    scaling_factor REAL,
    liquidity_score INTEGER,
    liquidity_status TEXT
);

-- Daily recommendations
CREATE TABLE recommendations (
    id INTEGER PRIMARY KEY,
    strategy_id INTEGER REFERENCES strategies(id),
    as_of_date DATE,
    symbol_id INTEGER,
    action TEXT,
    rationale TEXT,
    score REAL,
    proposed_size INTEGER,
    scaled_size INTEGER,
    liquidity_ratio REAL,
    run_hash TEXT
);

-- Liquidity cache
CREATE TABLE daily_liquidity_cache (
    symbol_id INTEGER,
    date DATE,
    avg_traded_value_30d REAL,
    liquidity_score INTEGER,
    updated_at TIMESTAMP,
    PRIMARY KEY (symbol_id, date)
);
```

## 4. Performance Metrics

### 4.1 Core Metrics

| Category | Metrics |
|----------|---------|
| Performance | Total Return, CAGR, Average Trade Return, Win Rate, Profit Factor |
| Risk | Max Drawdown, Avg Drawdown, Volatility, Sharpe Ratio, Sortino Ratio |
| Efficiency | Exposure %, Turnover, Avg Holding Period, Capital Utilization |
| Liquidity | Avg Slippage %, Fill Rate, Liquidity Constraint Rejections |
| Robustness | IS vs OOS Performance, Parameter Stability Score |

### 4.2 ISX-Specific Metrics
- Zero-Volume Penalty Sum
- Average Liquidity Ratio of executed trades
- Distribution of position scaling factors
- Percentage of days unable to trade due to liquidity

## 5. User Interface

### 5.1 Strategy Builder
- Visual condition builder with AND/OR logic
- Indicator parameter configuration
- Capital allocation and liquidity settings
- Real-time validation feedback

### 5.2 Backtest Results
- Interactive equity curve
- Comprehensive metrics dashboard
- Trade-by-trade analysis
- Liquidity impact visualization

### 5.3 Optimization Dashboard
- Parameter sensitivity heatmaps
- In-sample vs out-of-sample comparison
- Walk-forward analysis results
- Stability scoring

### 5.4 Daily Recommendations
- Ranked signal list with scores
- Liquidity-adjusted position sizes
- Detailed rationale for each signal
- One-click strategy performance review

## 6. Security & Validation

### 6.1 Input Validation
- Strategy IR schema validation
- Parameter bounds checking
- Look-ahead bias prevention
- Capital allocation reasonableness

### 6.2 Execution Safety
- Maximum position limits
- Daily loss limits
- Liquidity constraint enforcement
- Slippage caps

## 7. Performance Optimization

### 7.1 Computation
- Vectorized indicator calculations
- Parallel backtest execution
- Indicator result caching
- Incremental metric updates

### 7.2 Storage
- Indexed lookups for price data
- Compressed storage for large result sets
- Periodic cache cleanup
- Archive old backtest runs

## 8. Testing Strategy

### 8.1 Test Categories
- **Unit Tests**: Indicators, validators, metrics
- **Integration Tests**: End-to-end backtests
- **Property Tests**: Invariant verification
- **Performance Tests**: Scalability benchmarks

### 8.2 Test Data
- Synthetic data with known patterns
- Historical ISX data subset
- Edge cases (holidays, suspensions)
- Extreme market conditions

## 9. Future Enhancements

### Phase 1 (MVP)
- Core strategy builder
- Basic backtesting
- Simple optimization
- Daily recommendations

### Phase 2
- Walk-forward analysis
- Advanced metrics
- Export capabilities
- Parameter stability analysis

### Phase 3
- DSL support
- Machine learning integration
- Portfolio-level analysis
- Risk aggregation

### Phase 4
- Intraday strategies
- Real-time execution
- Advanced order types
- Multi-strategy portfolios

## 10. Success Criteria

1. **Accuracy**: Backtest results within 5% of reference implementation
2. **Performance**: 10-year backtest < 5 seconds on mid-range hardware
3. **Usability**: 90% of users can create strategies without documentation
4. **Reliability**: 100% reproducible results via run hash
5. **Liquidity**: Realistic handling of ISX market conditions