# ISX Liquidity Scoring System Specifications

## Overview

This document specifies the hybrid liquidity scoring system for ISX equities that:
- Uses only trading value data (no market cap required)
- Handles outliers through winsorization
- Penalizes non-trading days appropriately
- Provides transparent, multi-dimensional liquidity assessment

## 1. Core Components

### 1.1 Adjusted Amihud Illiquidity Measure

**Formula:**
```
ILLIQ_raw = (1/N_trd) * Σ(|R_d| / V_d_clean)
ILLIQ_adj = ILLIQ_raw * exp(α * p0)
```

Where:
- `N_trd` = Number of trading days in window
- `R_d` = Daily return (absolute value)
- `V_d_clean` = Winsorized trading value
- `p0` = Proportion of non-trading days
- `α` = Penalty parameter (default: 0.8)

### 1.2 Value Intensity

**Formula:**
```
VALINT = (1/T_open) * Σ(V_d_clean)
```

Where:
- `T_open` = Total exchange open days in window
- Includes zeros for non-trading days

### 1.3 Trading Continuity

**Formula:**
```
CONT = 1 - p0 = N_trd / T_open
```

## 2. Data Processing Pipeline

### 2.1 Input Data Requirements

| Field | Description | Source |
|-------|-------------|--------|
| date | Trading date | Daily data |
| ticker | Stock symbol | Daily data |
| close | Closing price | Daily data |
| value | Trading value (IQD) | Daily data |
| market_open | Exchange open flag | Index table |

### 2.2 Outlier Treatment

1. **Log-transform** trading values: `x_d = ln(V_d)`
2. **Calculate** mean (μ) and standard deviation (σ)
3. **Winsorize** at k*σ bounds (default k=2):
   ```
   x_d_winsor = min(max(x_d, μ - k*σ), μ + k*σ)
   ```
4. **Back-transform**: `V_d_clean = exp(x_d_winsor)`

### 2.3 Window Parameters

- **Primary window**: 60 trading days (~3 months)
- **Fast window**: 20 trading days (for responsiveness)
- **Rolling calculation**: Daily updates

## 3. Score Calculation

### 3.1 Component Normalization

For each component, compute robust z-scores:
```
z = (X - median(X)) / (1.4826 * MAD(X))
```

Convert to percentile scores (0-100):
- **ImpactScore**: percentile(-z_ILLIQ_adj)
- **ValueIntensityScore**: percentile(z_VALINT)
- **ContinuityScore**: percentile(z_CONT)

### 3.2 Composite Score

```
HybridScore = (w_I * ImpactScore + w_V * ValueIntensityScore + w_C * ContinuityScore) / (w_I + w_V + w_C)
```

Default weights: w_I = w_V = w_C = 1 (equal weighting)

## 4. Database Schema

### 4.1 liquidity_scores Table

```sql
CREATE TABLE liquidity_scores (
    date DATE NOT NULL,
    ticker VARCHAR(10) NOT NULL,
    window_days INT NOT NULL,
    
    -- Raw metrics
    illiq_raw DECIMAL(20,10),
    illiq_adj DECIMAL(20,10),
    value_intensity DECIMAL(20,2),
    continuity DECIMAL(5,4),
    p0_non_trading DECIMAL(5,4),
    
    -- Component scores (0-100)
    impact_score DECIMAL(5,2),
    value_intensity_score DECIMAL(5,2),
    continuity_score DECIMAL(5,2),
    
    -- Composite
    hybrid_score DECIMAL(5,2),
    
    -- Metadata
    trading_days INT,
    total_days INT,
    winsor_lower DECIMAL(20,10),
    winsor_upper DECIMAL(20,10),
    alpha_used DECIMAL(3,2),
    
    PRIMARY KEY (date, ticker, window_days),
    INDEX idx_ticker_date (ticker, date),
    INDEX idx_score (hybrid_score)
);
```

### 4.2 liquidity_alerts Table

```sql
CREATE TABLE liquidity_alerts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    ticker VARCHAR(10) NOT NULL,
    alert_type ENUM('low_liquidity', 'liquidity_drop', 'high_non_trading'),
    severity ENUM('warning', 'critical'),
    message TEXT,
    metrics JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ticker_date (ticker, date)
);
```

## 5. API Endpoints

### 5.1 Current Liquidity Scores
```
GET /api/liquidity/current
```

Response:
```json
{
  "date": "2025-01-19",
  "window_days": 60,
  "scores": [
    {
      "ticker": "BGUC",
      "hybrid_score": 75.3,
      "components": {
        "impact": 80.2,
        "value_intensity": 72.5,
        "continuity": 73.2
      },
      "flags": ["stable"]
    }
  ]
}
```

### 5.2 Historical Liquidity
```
GET /api/liquidity/history/{ticker}?days=90
```

### 5.3 Liquidity Rankings
```
GET /api/liquidity/rankings?date=2025-01-19&top=20
```

## 6. Alert Conditions

### 6.1 Low Liquidity Alert
- Triggered when: `hybrid_score < 25`
- Severity: warning (< 25), critical (< 10)

### 6.2 Liquidity Drop Alert
- Triggered when: Week-over-week drop > 20 percentile points
- Severity: Based on magnitude of drop

### 6.3 High Non-Trading Alert
- Triggered when: `p0 > 0.5` (more than 50% non-trading days)
- Severity: critical

## 7. Implementation Notes

### 7.1 Performance Optimization
- Pre-calculate rolling windows using incremental updates
- Cache cross-sectional statistics for normalization
- Use batch processing for historical calculations

### 7.2 Edge Cases
- New listings: Require minimum 10 trading days
- Delisted stocks: Maintain historical scores
- Stock suspensions: Flag separately from natural non-trading

### 7.3 Validation
- Correlation with bid-ask spread proxies
- Relationship with future volatility
- Cross-validation with actual execution costs (when available)

## 8. User Interface Requirements

### 8.1 Ticker Page Display
- Liquidity score badge (color-coded)
- Component breakdown chart
- Historical trend line
- Peer comparison

### 8.2 Market Overview
- Liquidity heatmap by sector
- Distribution of scores
- Market-wide liquidity index
- Most/least liquid stocks

## 9. Future Enhancements

### 9.1 Machine Learning
- Dynamic weight optimization
- Regime detection
- Predictive liquidity modeling

### 9.2 Additional Metrics
- Intraday volatility (when available)
- Order book imbalance (if data becomes available)
- News sentiment impact

## References

1. Amihud, Y. (2002). "Illiquidity and stock returns"
2. Corwin, S. A., & Schultz, P. (2012). "A simple way to estimate bid-ask spreads"
3. Roll, R. (1984). "A simple implicit measure of the effective bid-ask spread"