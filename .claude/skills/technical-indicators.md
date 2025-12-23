---
name: technical-indicators
type: domain-skill
expertise_level: expert
domain: financial-technical-analysis
version: "1.0.0"
---

# Technical Indicators Skill - Trading Indicator Calculations

## Purpose
Expert knowledge in calculating and interpreting technical indicators for financial market analysis, specifically optimized for the ISX Pulse platform's pipeline-based pre-calculation architecture.

## Indicator Categories

### 1. Moving Averages (Trend Indicators)

#### Simple Moving Average (SMA)
```go
func CalculateSMA(prices []float64, period int) []float64 {
    if len(prices) < period {
        return nil
    }

    sma := make([]float64, len(prices)-period+1)
    sum := 0.0

    for i := 0; i < period; i++ {
        sum += prices[i]
    }
    sma[0] = sum / float64(period)

    for i := period; i < len(prices); i++ {
        sum = sum - prices[i-period] + prices[i]
        sma[i-period+1] = sum / float64(period)
    }

    return sma
}
```

**Common Periods**: SMA 20 (short-term), SMA 50 (medium-term), SMA 200 (long-term)
**Usage**: Trend identification, support/resistance levels, crossover signals

#### Exponential Moving Average (EMA)
```go
func CalculateEMA(prices []float64, period int) []float64 {
    if len(prices) == 0 {
        return nil
    }

    ema := make([]float64, len(prices))
    multiplier := 2.0 / (float64(period) + 1.0)

    // Start with SMA as first EMA value
    sum := 0.0
    for i := 0; i < period && i < len(prices); i++ {
        sum += prices[i]
    }
    ema[period-1] = sum / float64(period)

    // Calculate EMA for remaining values
    for i := period; i < len(prices); i++ {
        ema[i] = (prices[i] * multiplier) + (ema[i-1] * (1 - multiplier))
    }

    return ema
}
```

**Common Period**: EMA 20 for short-term trend following
**Usage**: Faster response to price changes, crossover strategies

### 2. Momentum Indicators

#### Relative Strength Index (RSI)
```go
func CalculateRSI(prices []float64, period int) []float64 {
    if len(prices) < period+1 {
        return nil
    }

    gains := make([]float64, len(prices)-1)
    losses := make([]float64, len(prices)-1)

    for i := 1; i < len(prices); i++ {
        change := prices[i] - prices[i-1]
        if change > 0 {
            gains[i-1] = change
            losses[i-1] = 0
        } else {
            gains[i-1] = 0
            losses[i-1] = -change
        }
    }

    avgGain := calculateSMA(gains, period)[0]
    avgLoss := calculateSMA(losses, period)[0]

    rsi := make([]float64, len(prices)-period)
    for i := period; i < len(gains); i++ {
        avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
        avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

        rs := avgGain / avgLoss
        rsi[i-period+1] = 100 - (100 / (1 + rs))
    }

    return rsi
}
```

**Standard Period**: 14 days
**Interpretation**:
- Above 70: Overbought condition
- Below 30: Oversold condition
- 30-70: Neutral zone

#### MACD (Moving Average Convergence Divergence)
```go
func CalculateMACD(prices []float64, fastPeriod, slowPeriod, signalPeriod int) (macdLine, signalLine, histogram []float64) {
    fastEMA := CalculateEMA(prices, fastPeriod)
    slowEMA := CalculateEMA(prices, slowPeriod)

    // Align arrays (slowEMA has fewer initial values)
    startIndex := len(slowEMA) - len(fastEMA)
    fastEMA = fastEMA[startIndex:]

    macdLine = make([]float64, len(fastEMA))
    for i := 0; i < len(fastEMA); i++ {
        macdLine[i] = fastEMA[i] - slowEMA[i]
    }

    signalLine = CalculateEMA(macdLine, signalPeriod)

    histogram = make([]float64, len(signalLine))
    for i := 0; i < len(signalLine); i++ {
        histogram[i] = macdLine[i+(len(macdLine)-len(signalLine))] - signalLine[i]
    }

    return macdLine, signalLine, histogram
}
```

**Standard Parameters**: Fast=12, Slow=26, Signal=9
**Usage**: Trend changes, momentum shifts, buy/sell signals

### 3. Volatility Indicators

#### Bollinger Bands
```go
func CalculateBollingerBands(prices []float64, period, stdDev int) (upperBand, middleBand, lowerBand []float64) {
    middleBand = CalculateSMA(prices, period)

    upperBand = make([]float64, len(middleBand))
    lowerBand = make([]float64, len(middleBand))

    for i := 0; i < len(middleBand); i++ {
        // Calculate standard deviation for the period
        variance := 0.0
        for j := i; j < i+period; j++ {
            diff := prices[j] - middleBand[i]
            variance += diff * diff
        }
        variance /= float64(period)
        stdDeviation := math.Sqrt(variance)

        multiplier := float64(stdDev)
        upperBand[i] = middleBand[i] + (stdDeviation * multiplier)
        lowerBand[i] = middleBand[i] - (stdDeviation * multiplier)
    }

    return upperBand, middleBand, lowerBand
}
```

**Standard Parameters**: 20-period SMA, 2 standard deviations
**Usage**: Volatility measurement, overbought/oversold conditions

#### Average True Range (ATR)
```go
func CalculateATR(high, low, close []float64, period int) []float64 {
    if len(high) < period+1 {
        return nil
    }

    trueRanges := make([]float64, len(high)-1)
    for i := 1; i < len(high); i++ {
        tr1 := high[i] - low[i]
        tr2 := math.Abs(high[i] - close[i-1])
        tr3 := math.Abs(low[i] - close[i-1])
        trueRanges[i-1] = math.Max(tr1, math.Max(tr2, tr3))
    }

    return CalculateSMA(trueRanges, period)
}
```

**Standard Period**: 14 days
**Usage**: Volatility measurement, stop-loss placement

### 4. Volume Indicators

#### On-Balance Volume (OBV)
```go
func CalculateOBV(close []float64, volume []int64) []int64 {
    if len(close) != len(volume) {
        return nil
    }

    obv := make([]int64, len(close))
    obv[0] = volume[0]

    for i := 1; i < len(close); i++ {
        if close[i] > close[i-1] {
            obv[i] = obv[i-1] + volume[i]
        } else if close[i] < close[i-1] {
            obv[i] = obv[i-1] - volume[i]
        } else {
            obv[i] = obv[i-1]
        }
    }

    return obv
}
```

**Usage**: Volume confirmation of price trends, divergence analysis

#### Volume Weighted Average Price (VWAP)
```go
func CalculateVWAP(prices []float64, volume []int64) []float64 {
    if len(prices) != len(volume) {
        return nil
    }

    vwap := make([]float64, len(prices))
    var cumulativeValue, cumulativeVolume float64

    for i := 0; i < len(prices); i++ {
        cumulativeValue += prices[i] * float64(volume[i])
        cumulativeVolume += float64(volume[i])

        if cumulativeVolume > 0 {
            vwap[i] = cumulativeValue / cumulativeVolume
        }
    }

    return vwap
}
```

**Usage**: Intraday trading levels, institutional activity analysis

### 5. Support & Resistance Indicators

#### Dynamic Support/Resistance Levels
```go
func CalculateSupportResistance(prices []float64, lookbackPeriod int) (support, resistance []float64) {
    support = make([]float64, len(prices))
    resistance = make([]float64, len(prices))

    for i := lookbackPeriod; i < len(prices)-lookbackPeriod; i++ {
        // Find local minimum (support)
        localMin := prices[i]
        for j := i - lookbackPeriod; j <= i + lookbackPeriod; j++ {
            if prices[j] < localMin {
                localMin = prices[j]
            }
        }
        support[i] = localMin

        // Find local maximum (resistance)
        localMax := prices[i]
        for j := i - lookbackPeriod; j <= i + lookbackPeriod; j++ {
            if prices[j] > localMax {
                localMax = prices[j]
            }
        }
        resistance[i] = localMax
    }

    return support, resistance
}
```

**Usage**: Entry/exit points, stop-loss levels, target prices

## ISX Pulse Integration Patterns

### Pipeline Pre-calculation Optimization
```go
type IndicatorCalculator struct {
    priceData map[string][]PriceData
    cache     map[string]map[string][]float64
}

func (ic *IndicatorCalculator) CalculateAllIndicators(ticker string, date string) error {
    prices := ic.priceData[ticker]

    // Batch calculate all indicators for efficiency
    indicators := make(map[string][]float64)

    // Moving Averages
    indicators["sma_20"] = CalculateSMA(prices, 20)
    indicators["sma_50"] = CalculateSMA(prices, 50)
    indicators["sma_200"] = CalculateSMA(prices, 200)
    indicators["ema_20"] = CalculateEMA(prices, 20)

    // Momentum
    indicators["rsi_14"] = CalculateRSI(prices, 14)
    macd, signal, histogram := CalculateMACD(prices, 12, 26, 9)
    indicators["macd_12_26_9"] = macd
    indicators["macd_signal"] = signal
    indicators["macd_histogram"] = histogram

    // Volatility
    upper, middle, lower := CalculateBollingerBands(prices, 20, 2)
    indicators["bollinger_upper"] = upper
    indicators["bollinger_middle"] = middle
    indicators["bollinger_lower"] = lower
    indicators["atr_14"] = CalculateATR(high, low, close, 14)

    // Store in cache for instant access
    ic.cache[ticker] = indicators

    return nil
}
```

### SSOT Storage Format
```csv
ticker,date,close,sma_20,sma_50,rsi_14,macd_12_26_9,bollinger_upper,bollinger_lower,atr_14
ISX-BKCP,2025-10-17,1.28,1.23,1.19,65.4,0.012,1.35,1.15,0.023
```

### Real-time Access Patterns
```go
func (s *IndicatorService) GetIndicator(ticker, indicator string, date string) (float64, error) {
    // Direct lookup from pre-calculated SSOT data
    data, err := s.csvIndex.Lookup(ticker, date)
    if err != nil {
        return 0, err
    }

    return data[indicator], nil
}
```

## Performance Optimization

### Memory Efficiency
- **Batch Processing**: Calculate all indicators for a ticker in one pass
- **Shared Calculations**: Reuse intermediate results (e.g., EMA calculations)
- **Streaming Processing**: Handle large datasets without loading everything into memory

### Calculation Accuracy
- **Floating Point Precision**: Use appropriate precision for financial calculations
- **Rounding Consistency**: Apply consistent rounding rules across all indicators
- **Boundary Handling**: Proper handling of insufficient data periods

### Error Handling
- **Data Validation**: Validate input data ranges and formats
- **Insufficient Data**: Graceful handling when not enough data points exist
- **Numerical Stability**: Handle edge cases (division by zero, extreme values)

## Trading Signal Generation

### Indicator Combinations
```go
type TradingSignal struct {
    Signal    string  // "BUY", "SELL", "HOLD"
    Strength  float64 // Signal strength 0-1
    Reason    string  // Explanation of signal
    Indicators map[string]float64 // Supporting indicator values
}

func GenerateSignal(indicators map[string]float64, currentPrice float64) TradingSignal {
    var score float64
    var reasons []string

    // RSI oversold/overbought
    rsi := indicators["rsi_14"]
    if rsi < 30 {
        score += 0.3
        reasons = append(reasons, "RSI oversold")
    } else if rsi > 70 {
        score -= 0.3
        reasons = append(reasons, "RSI overbought")
    }

    // MACD crossover
    macd := indicators["macd_12_26_9"]
    signal := indicators["macd_signal"]
    if macd > signal {
        score += 0.4
        reasons = append(reasons, "MACD bullish crossover")
    } else {
        score -= 0.4
        reasons = append(reasons, "MACD bearish crossover")
    }

    // Price vs SMA
    sma20 := indicators["sma_20"]
    if currentPrice > sma20 {
        score += 0.3
        reasons = append(reasons, "Price above SMA20")
    } else {
        score -= 0.3
        reasons = append(reasons, "Price below SMA20")
    }

    var signalType string
    if score > 0.5 {
        signalType = "BUY"
    } else if score < -0.5 {
        signalType = "SELL"
    } else {
        signalType = "HOLD"
    }

    return TradingSignal{
        Signal:     signalType,
        Strength:   math.Abs(score),
        Reason:     strings.Join(reasons, "; "),
        Indicators: indicators,
    }
}
```

This skill provides comprehensive technical indicator expertise optimized for the ISX Pulse platform's pre-calculation architecture, ensuring high-performance calculations and consistent results across all analysis scenarios.