# ISX Liquidity Calculation Approach

## Executive Summary

This document describes the hybrid liquidity scoring system designed specifically for the Iraqi Stock Exchange (ISX). The system addresses the unique challenges of frontier markets where:
- Market capitalization data may be unreliable or unavailable
- Trading is often intermittent with many non-trading days
- Large block trades can distort liquidity metrics
- Traditional liquidity measures may not capture the full picture

Our approach combines multiple liquidity dimensions into a single, interpretable score that helps traders understand the true cost and difficulty of trading ISX securities.

## 1. Theoretical Foundation

### 1.1 What is Liquidity?

Liquidity in financial markets encompasses multiple dimensions:
- **Tightness**: The cost of immediate execution (bid-ask spread)
- **Depth**: The volume available at current prices
- **Immediacy**: The speed at which trades can be executed
- **Breadth**: The market's ability to absorb large orders without significant price impact
- **Resiliency**: How quickly prices recover after a trade-induced movement

### 1.2 Why a Hybrid Approach?

No single metric captures all liquidity dimensions. Academic research shows that combining multiple measures provides more robust liquidity assessment, especially in emerging markets where:
- Intraday data is unavailable
- Trading is sporadic
- Market microstructure differs from developed markets

## 2. Core Components

### 2.1 Adjusted Amihud Illiquidity Measure

The Amihud (2002) measure captures price impact - how much prices move per unit of trading value.

**Basic Formula:**
```
ILLIQ = Average(|Daily Return| / Daily Trading Value)
```

**Our Adjustments:**

1. **Outlier Handling**: We winsorize trading values in log space to reduce the impact of occasional large block trades:
   ```
   log(Value) → Winsorize at μ ± 2σ → exp() → Clean Value
   ```

2. **Non-Trading Penalty**: We apply an exponential penalty for non-trading days:
   ```
   ILLIQ_adjusted = ILLIQ_raw × exp(α × proportion_non_trading_days)
   ```
   
   Where α = 0.8 (calibratable) ensures stocks that rarely trade show appropriately high illiquidity.

**Interpretation**: Higher values indicate worse liquidity (larger price movements per IQD traded).

### 2.2 Value Intensity

This component measures the consistency and magnitude of trading activity.

**Formula:**
```
VALINT = Sum(Clean Trading Values) / Total Exchange Open Days
```

**Key Features:**
- Includes zero values for non-trading days
- Uses winsorized values to reduce block trade distortion
- Rewards consistent daily trading over sporadic large trades

**Interpretation**: Higher values indicate better liquidity (more consistent trading activity).

### 2.3 Trading Continuity

This directly measures how often a stock trades when the exchange is open.

**Formula:**
```
CONTINUITY = Trading Days / Exchange Open Days = 1 - p0
```

Where p0 is the proportion of non-trading days.

**Interpretation**: Higher values indicate better liquidity (more frequent trading).

## 3. Composite Score Construction

### 3.1 Component Normalization

Each component is normalized using robust statistics:

1. **Calculate robust z-scores**:
   ```
   z = (value - median) / (1.4826 × MAD)
   ```
   Where MAD is the Median Absolute Deviation.

2. **Convert to percentiles** (0-100 scale):
   - For illiquidity measures (higher = worse): Use inverted percentile
   - For liquidity measures (higher = better): Use direct percentile

### 3.2 Weighting Scheme

**Default Equal Weighting:**
```
HybridScore = (ImpactScore + ValueIntensityScore + ContinuityScore) / 3
```

**Rationale**: Each component captures a different liquidity dimension:
- Impact Score: Transaction cost perspective
- Value Intensity: Market depth perspective  
- Continuity: Trading availability perspective

Weights can be optimized based on empirical validation against actual trading costs.

## 4. Implementation Details

### 4.1 Window Selection

- **Primary Window**: 60 trading days (~3 months)
  - Balances stability with responsiveness
  - Sufficient for robust statistics
  
- **Fast Window**: 20 trading days (~1 month)
  - For detecting recent liquidity changes
  - Used for alerts and warnings

### 4.2 Data Requirements

**Minimal Data Needed:**
- Daily closing prices
- Daily trading values (price × volume)
- Exchange calendar (open/closed days)

**Not Required:**
- Market capitalization
- Free float
- Intraday data
- Order book information

### 4.3 Handling Special Cases

**New Listings:**
- Require minimum 10 trading days
- Use expanding window until 60 days available

**Stock Suspensions:**
- Distinguish from natural non-trading
- Flag separately in the system

**Corporate Actions:**
- Adjust historical prices for splits/dividends
- Recalculate metrics after adjustments

## 5. Practical Applications

### 5.1 For Traders

**Pre-Trade Analysis:**
- Assess likely transaction costs
- Compare liquidity across securities
- Identify optimal trading times

**Risk Management:**
- Set position size limits based on liquidity
- Monitor liquidity deterioration
- Avoid liquidity traps

### 5.2 For Portfolio Managers

**Portfolio Construction:**
- Liquidity-adjusted position sizing
- Diversification across liquidity buckets
- Rebalancing cost estimation

**Performance Attribution:**
- Separate liquidity premium from alpha
- Track liquidity risk exposure
- Benchmark liquidity profiles

### 5.3 For Market Analysis

**Market Structure:**
- Identify systematically illiquid securities
- Track market-wide liquidity trends
- Detect liquidity regime changes

## 6. Validation Approach

### 6.1 Statistical Validation

**Cross-Sectional Tests:**
- Correlation with bid-ask spread proxies
- Relationship with trading frequency
- Size and sector effects

**Time-Series Tests:**
- Stability of rankings over time
- Predictive power for future liquidity
- Regime change detection

### 6.2 Economic Validation

**Trading Cost Prediction:**
- Compare scores with actual execution costs
- Validate during different market conditions
- Test across security types

**Risk-Return Relationships:**
- Liquidity premium estimation
- Volatility-liquidity connections
- Market impact modeling

## 7. Advantages Over Single Metrics

### 7.1 Robustness

- Multiple dimensions reduce measurement error
- Outlier handling prevents distortion
- Non-trading penalty captures true illiquidity

### 7.2 Interpretability

- Composite score on 0-100 scale
- Component breakdown available
- Clear flags for specific issues

### 7.3 Adaptability

- Works with limited data
- Adjusts to market structure changes
- Calibratable parameters

## 8. Limitations and Future Enhancements

### 8.1 Current Limitations

- Cannot capture intraday liquidity variations
- May not fully reflect large order market impact
- Assumes trading value data is accurate

### 8.2 Planned Enhancements

**Phase 1 (with current data):**
- Machine learning for weight optimization
- Sector-specific calibration
- Liquidity forecast models

**Phase 2 (with additional data):**
- Incorporate order book data if available
- Add market maker participation metrics
- Include cross-market liquidity measures

## 9. Technical Implementation

### 9.1 Calculation Pipeline

```
1. Data Collection
   ├── Daily prices and values
   ├── Exchange calendar
   └── Corporate actions

2. Data Cleaning
   ├── Adjust for corporate actions
   ├── Identify non-trading days
   └── Validate data integrity

3. Metric Calculation
   ├── Winsorize outliers
   ├── Calculate components
   └── Apply penalties

4. Score Generation
   ├── Normalize components
   ├── Compute composite
   └── Generate rankings

5. Output
   ├── Store in database
   ├── Update APIs
   └── Trigger alerts
```

### 9.2 Performance Considerations

- Incremental calculation for efficiency
- Caching of cross-sectional statistics
- Parallel processing for multiple securities

## 10. References

1. **Amihud, Y. (2002)**. "Illiquidity and stock returns: cross-section and time-series effects." Journal of Financial Markets, 5(1), 31-56.

2. **Corwin, S. A., & Schultz, P. (2012)**. "A simple way to estimate bid-ask spreads from daily high and low prices." Journal of Finance, 67(2), 719-760.

3. **Roll, R. (1984)**. "A simple implicit measure of the effective bid-ask spread in an efficient market." Journal of Finance, 39(4), 1127-1139.

4. **Lesmond, D. A., Ogden, J. P., & Trzcinka, C. A. (1999)**. "A new estimate of transaction costs." Review of Financial Studies, 12(5), 1113-1141.

5. **Hasbrouck, J. (2009)**. "Trading costs and returns for U.S. equities: Estimating effective costs from daily data." Journal of Finance, 64(3), 1445-1477.

## Appendix A: Parameter Calibration Guide

### A.1 Alpha (Non-Trading Penalty)

Start with α = 0.8 and adjust based on:
- Market characteristics (higher for more liquid markets)
- Validation against actual spreads
- User feedback on score accuracy

### A.2 Winsorization Bounds

Default k = 2 (two standard deviations) works well, but consider:
- Asymmetric bounds if block trades are predominantly buys or sells
- Sector-specific bounds for different trading patterns
- Time-varying bounds during market stress

### A.3 Window Length

60 days balances several factors:
- Statistical reliability (enough observations)
- Responsiveness (not too slow to adapt)
- Seasonality (captures different market conditions)

## Appendix B: SQL Implementation Example

```sql
-- Calculate daily liquidity metrics
WITH daily_metrics AS (
    SELECT 
        date,
        ticker,
        ABS(LN(close/prev_close)) as abs_return,
        value,
        LN(value) as log_value,
        CASE WHEN value > 0 THEN 1 ELSE 0 END as is_trading_day
    FROM daily_data
    WHERE date >= DATE_SUB(CURRENT_DATE, INTERVAL 60 DAY)
),
-- Winsorize values
winsorized AS (
    SELECT 
        *,
        AVG(log_value) OVER (PARTITION BY ticker) as mu,
        STDDEV(log_value) OVER (PARTITION BY ticker) as sigma,
        GREATEST(LEAST(log_value, mu + 2*sigma), mu - 2*sigma) as log_value_clean,
        EXP(GREATEST(LEAST(log_value, mu + 2*sigma), mu - 2*sigma)) as value_clean
    FROM daily_metrics
    WHERE is_trading_day = 1
),
-- Calculate components
components AS (
    SELECT 
        ticker,
        AVG(abs_return / value_clean) as illiq_raw,
        SUM(value_clean) / 60 as value_intensity,
        SUM(is_trading_day) / 60.0 as continuity,
        (60 - SUM(is_trading_day)) / 60.0 as p0
    FROM winsorized
    GROUP BY ticker
)
-- Final scores
SELECT 
    ticker,
    illiq_raw,
    illiq_raw * EXP(0.8 * p0) as illiq_adj,
    value_intensity,
    continuity,
    -- Percentile ranks would be calculated here
    PERCENT_RANK() OVER (ORDER BY illiq_adj DESC) * 100 as impact_score,
    PERCENT_RANK() OVER (ORDER BY value_intensity) * 100 as intensity_score,
    PERCENT_RANK() OVER (ORDER BY continuity) * 100 as continuity_score
FROM components;
```