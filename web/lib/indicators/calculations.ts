/**
 * Technical Indicator Calculations
 * Professional-grade implementations following industry standards
 * Used by TradingView Lightweight Charts integration
 */

/**
 * Simple Moving Average (SMA)
 * Calculates the average of the last N periods
 *
 * @param data - Array of prices (typically close prices)
 * @param period - Number of periods to average
 * @returns Array of SMA values (null for insufficient data)
 */
export function calculateSMA(data: number[], period: number): (number | null)[] {
  if (!data || data.length === 0) return []
  if (period <= 0 || period > data.length) {
    return data.map(() => null)
  }

  const result: (number | null)[] = []

  for (let i = 0; i < data.length; i++) {
    if (i < period - 1) {
      result.push(null)
      continue
    }

    let sum = 0
    for (let j = 0; j < period; j++) {
      sum += data[i - j]
    }
    result.push(sum / period)
  }

  return result
}

/**
 * Exponential Moving Average (EMA)
 * Gives more weight to recent prices
 *
 * @param data - Array of prices (typically close prices)
 * @param period - Number of periods for calculation
 * @returns Array of EMA values (null for insufficient data)
 */
export function calculateEMA(data: number[], period: number): (number | null)[] {
  if (!data || data.length === 0) return []
  if (period <= 0 || period > data.length) {
    return data.map(() => null)
  }

  const result: (number | null)[] = []
  const multiplier = 2 / (period + 1)

  // Calculate initial SMA for first EMA value
  let sum = 0
  for (let i = 0; i < period && i < data.length; i++) {
    sum += data[i]
  }
  const initialSMA = sum / period

  for (let i = 0; i < data.length; i++) {
    if (i < period - 1) {
      result.push(null)
    } else if (i === period - 1) {
      result.push(initialSMA)
    } else {
      const prevEMA = result[i - 1]!
      result.push((data[i] - prevEMA) * multiplier + prevEMA)
    }
  }

  return result
}

/**
 * Calculate Support Level
 * Support = Average of lowest 10% of prices
 *
 * @param data - Array of OHLC data
 * @returns Support price level
 */
export function calculateSupport(data: { low: number }[]): number {
  if (!data || data.length === 0) return 0

  const lows = data.map(d => d.low).filter(l => l > 0)
  if (lows.length === 0) return 0

  // Support = Average of lowest 10% of prices
  const sorted = [...lows].sort((a, b) => a - b)
  const count = Math.max(1, Math.floor(sorted.length * 0.1))
  const sum = sorted.slice(0, count).reduce((a, b) => a + b, 0)
  return sum / count
}

/**
 * Calculate Resistance Level
 * Resistance = Average of highest 10% of prices
 *
 * @param data - Array of OHLC data
 * @returns Resistance price level
 */
export function calculateResistance(data: { high: number }[]): number {
  if (!data || data.length === 0) return 0

  const highs = data.map(d => d.high).filter(h => h > 0)
  if (highs.length === 0) return 0

  // Resistance = Average of highest 10% of prices
  const sorted = [...highs].sort((a, b) => b - a)
  const count = Math.max(1, Math.floor(sorted.length * 0.1))
  const sum = sorted.slice(0, count).reduce((a, b) => a + b, 0)
  return sum / count
}

/**
 * Relative Strength Index (RSI)
 * Momentum oscillator (0-100 range)
 * Uses Wilder's smoothing method for average gain/loss
 *
 * @param closePrices - Array of close prices
 * @param period - RSI period (typically 14)
 * @returns Array of RSI values (0-100)
 */
export function calculateRSI(closePrices: number[], period: number = 14): (number | null)[] {
  if (!closePrices || closePrices.length === 0) return []
  if (period <= 0 || closePrices.length < period + 1) {
    return closePrices.map(() => null)
  }

  const result: (number | null)[] = []
  const gains: number[] = []
  const losses: number[] = []

  // Calculate price changes (gains/losses array is 1 shorter than closePrices)
  for (let i = 1; i < closePrices.length; i++) {
    const change = closePrices[i] - closePrices[i - 1]
    gains.push(change > 0 ? change : 0)
    losses.push(change < 0 ? Math.abs(change) : 0)
  }

  // Not enough data
  if (gains.length < period) {
    return closePrices.map(() => null)
  }

  // Calculate initial average gain/loss (SMA of first 'period' changes)
  let avgGain = gains.slice(0, period).reduce((a, b) => a + b, 0) / period
  let avgLoss = losses.slice(0, period).reduce((a, b) => a + b, 0) / period

  // Fill initial values with null (indices 0 to period-1)
  // BUG FIX (2025-01-04): Changed from i <= period to i < period
  // Previous code created period+1 nulls, causing off-by-one error
  for (let i = 0; i < period; i++) {
    result.push(null)
  }

  // Calculate first RSI at index 'period' (uses initial avgGain/avgLoss)
  // Special cases: pure uptrend (avgLoss=0) → RSI=100, pure downtrend (avgGain=0) → RSI=0
  if (avgLoss === 0) {
    result.push(100)  // Pure uptrend
  } else if (avgGain === 0) {
    result.push(0)    // Pure downtrend
  } else {
    const rs = avgGain / avgLoss
    const rsi = 100 - (100 / (1 + rs))
    result.push(rsi)
  }

  // Calculate subsequent RSI values using Wilder's smoothing
  // Formula: avgGain = ((previous avgGain * (period-1)) + current gain) / period
  for (let i = period; i < gains.length; i++) {
    avgGain = ((avgGain * (period - 1)) + gains[i]) / period
    avgLoss = ((avgLoss * (period - 1)) + losses[i]) / period

    // Calculate RSI with edge case handling
    if (avgLoss === 0) {
      result.push(100)  // Pure uptrend
    } else if (avgGain === 0) {
      result.push(0)    // Pure downtrend
    } else {
      const rs = avgGain / avgLoss
      const rsi = 100 - (100 / (1 + rs))
      result.push(rsi)
    }
  }

  return result
}

/**
 * MACD (Moving Average Convergence Divergence)
 * Trend-following momentum indicator
 *
 * @param closePrices - Array of close prices
 * @param fastPeriod - Fast EMA period (typically 12)
 * @param slowPeriod - Slow EMA period (typically 26)
 * @param signalPeriod - Signal line period (typically 9)
 * @returns MACD line, Signal line, and Histogram
 */
export interface MACDResult {
  macd: (number | null)[]
  signal: (number | null)[]
  histogram: (number | null)[]
}

export function calculateMACD(
  closePrices: number[],
  fastPeriod: number = 12,
  slowPeriod: number = 26,
  signalPeriod: number = 9
): MACDResult {
  if (!closePrices || closePrices.length === 0) {
    return { macd: [], signal: [], histogram: [] }
  }

  const fastEMA = calculateEMA(closePrices, fastPeriod)
  const slowEMA = calculateEMA(closePrices, slowPeriod)

  // MACD Line = FastEMA - SlowEMA
  const macdLine: (number | null)[] = fastEMA.map((fast, i) => {
    const slow = slowEMA[i]
    if (fast === null || slow === null) return null
    return fast - slow
  })

  // Signal Line = 9-period EMA of MACD
  const macdValues = macdLine.filter(v => v !== null) as number[]
  const signalEMA = calculateEMA(macdValues, signalPeriod)

  // Align signal line with original data length
  const signalLine: (number | null)[] = []
  let signalIndex = 0
  for (let i = 0; i < macdLine.length; i++) {
    if (macdLine[i] === null) {
      signalLine.push(null)
    } else {
      signalLine.push(signalEMA[signalIndex] ?? null)
      signalIndex++
    }
  }

  // Histogram = MACD - Signal
  const histogram: (number | null)[] = macdLine.map((macd, i) => {
    const signal = signalLine[i]
    if (macd === null || signal === null) return null
    return macd - signal
  })

  return { macd: macdLine, signal: signalLine, histogram }
}

/**
 * Bollinger Bands
 * Volatility indicator with upper and lower bands
 *
 * @param closePrices - Array of close prices
 * @param period - Moving average period (typically 20)
 * @param stdDevMultiplier - Standard deviation multiplier (typically 2)
 * @returns Upper band, Middle band (SMA), Lower band
 */
export interface BollingerBandsResult {
  upper: (number | null)[]
  middle: (number | null)[]
  lower: (number | null)[]
}

export function calculateBollingerBands(
  closePrices: number[],
  period: number = 20,
  stdDevMultiplier: number = 2
): BollingerBandsResult {
  if (!closePrices || closePrices.length === 0) {
    return { upper: [], middle: [], lower: [] }
  }

  const middle = calculateSMA(closePrices, period)
  const upper: (number | null)[] = []
  const lower: (number | null)[] = []

  for (let i = 0; i < closePrices.length; i++) {
    if (middle[i] === null || i < period - 1) {
      upper.push(null)
      lower.push(null)
      continue
    }

    // Calculate standard deviation for the period
    let sumSquaredDiff = 0
    for (let j = 0; j < period; j++) {
      const diff = closePrices[i - j] - middle[i]!
      sumSquaredDiff += diff * diff
    }
    const stdDev = Math.sqrt(sumSquaredDiff / period)

    upper.push(middle[i]! + stdDevMultiplier * stdDev)
    lower.push(middle[i]! - stdDevMultiplier * stdDev)
  }

  return { upper, middle, lower }
}

/**
 * Stochastic Oscillator
 * Momentum indicator comparing close price to price range
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param kPeriod - %K period (typically 14)
 * @param dPeriod - %D period (typically 3)
 * @returns %K and %D lines (0-100 range)
 */
export interface StochasticResult {
  k: (number | null)[]
  d: (number | null)[]
}

export function calculateStochastic(
  highs: number[],
  lows: number[],
  closes: number[],
  kPeriod: number = 14,
  dPeriod: number = 3
): StochasticResult {
  if (!highs || !lows || !closes || highs.length === 0) {
    return { k: [], d: [] }
  }

  const k: (number | null)[] = []

  // Calculate %K
  for (let i = 0; i < closes.length; i++) {
    if (i < kPeriod - 1) {
      k.push(null)
      continue
    }

    // Find highest high and lowest low in period
    let highestHigh = highs[i]
    let lowestLow = lows[i]
    for (let j = 0; j < kPeriod; j++) {
      highestHigh = Math.max(highestHigh, highs[i - j])
      lowestLow = Math.min(lowestLow, lows[i - j])
    }

    const range = highestHigh - lowestLow
    if (range === 0) {
      k.push(50) // Neutral if no range
    } else {
      const kValue = ((closes[i] - lowestLow) / range) * 100
      k.push(kValue)
    }
  }

  // Calculate %D (SMA of %K)
  const kValues = k.filter(v => v !== null) as number[]
  const dValues = calculateSMA(kValues, dPeriod)

  // Align %D with original data length
  const d: (number | null)[] = []
  let dIndex = 0
  for (let i = 0; i < k.length; i++) {
    if (k[i] === null) {
      d.push(null)
    } else {
      d.push(dValues[dIndex] ?? null)
      dIndex++
    }
  }

  return { k, d }
}

/**
 * Average Directional Index (ADX)
 * Trend strength indicator
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param period - ADX period (typically 14)
 * @returns ADX, +DI, -DI values
 */
export interface ADXResult {
  adx: (number | null)[]
  plusDI: (number | null)[]
  minusDI: (number | null)[]
}

export function calculateADX(
  highs: number[],
  lows: number[],
  closes: number[],
  period: number = 14
): ADXResult {
  const length = highs.length

  // Validation
  if (!highs || !lows || !closes || length === 0) {
    return { adx: [], plusDI: [], minusDI: [] }
  }
  if (lows.length !== length || closes.length !== length) {
    return { adx: [], plusDI: [], minusDI: [] }
  }
  if (period <= 0 || period >= length) {
    return {
      adx: highs.map(() => null),
      plusDI: highs.map(() => null),
      minusDI: highs.map(() => null)
    }
  }

  // Initialize result arrays
  const plusDI: (number | null)[] = []
  const minusDI: (number | null)[] = []
  const adx: (number | null)[] = []

  // Arrays for intermediate calculations
  const trueRanges: number[] = []
  const plusDMs: number[] = []
  const minusDMs: number[] = []

  // Step 1: Calculate True Range, +DM, -DM for each period
  for (let i = 0; i < length; i++) {
    if (i === 0) {
      // First period: no previous data
      trueRanges.push(highs[i] - lows[i])
      plusDMs.push(0)
      minusDMs.push(0)
    } else {
      // True Range = max(high - low, |high - prevClose|, |low - prevClose|)
      const hl = highs[i] - lows[i]
      const hc = Math.abs(highs[i] - closes[i - 1])
      const lc = Math.abs(lows[i] - closes[i - 1])
      const tr = Math.max(hl, hc, lc)
      trueRanges.push(tr)

      // Directional Movement
      const highDiff = highs[i] - highs[i - 1]
      const lowDiff = lows[i - 1] - lows[i]

      let plusDM = 0
      let minusDM = 0

      if (highDiff > lowDiff && highDiff > 0) {
        plusDM = highDiff
      }
      if (lowDiff > highDiff && lowDiff > 0) {
        minusDM = lowDiff
      }

      plusDMs.push(plusDM)
      minusDMs.push(minusDM)
    }
  }

  // Step 2: Smooth TR, +DM, -DM using Wilder's smoothing method
  let smoothedTR = 0
  let smoothedPlusDM = 0
  let smoothedMinusDM = 0

  // Calculate initial smoothed values (sum of first 'period' values)
  for (let i = 0; i < period; i++) {
    smoothedTR += trueRanges[i]
    smoothedPlusDM += plusDMs[i]
    smoothedMinusDM += minusDMs[i]
  }

  // Fill initial null values
  for (let i = 0; i < period - 1; i++) {
    plusDI.push(null)
    minusDI.push(null)
    adx.push(null)
  }

  // Step 3: Calculate +DI and -DI using smoothed values
  const dxValues: number[] = []

  for (let i = period - 1; i < length; i++) {
    if (i > period - 1) {
      // Wilder's smoothing: smoothed[i] = smoothed[i-1] - (smoothed[i-1]/period) + current[i]
      smoothedTR = smoothedTR - (smoothedTR / period) + trueRanges[i]
      smoothedPlusDM = smoothedPlusDM - (smoothedPlusDM / period) + plusDMs[i]
      smoothedMinusDM = smoothedMinusDM - (smoothedMinusDM / period) + minusDMs[i]
    }

    // Calculate +DI and -DI
    const plusDIValue = smoothedTR > 0 ? (smoothedPlusDM / smoothedTR) * 100 : 0
    const minusDIValue = smoothedTR > 0 ? (smoothedMinusDM / smoothedTR) * 100 : 0

    plusDI.push(plusDIValue)
    minusDI.push(minusDIValue)

    // Calculate DX (Directional Index)
    const diSum = plusDIValue + minusDIValue
    const diDiff = Math.abs(plusDIValue - minusDIValue)
    const dx = diSum > 0 ? (diDiff / diSum) * 100 : 0
    dxValues.push(dx)
  }

  // Step 4: Calculate ADX as smoothed average of DX using Wilder's smoothing
  // First ADX = average of first 'period' DX values
  if (dxValues.length >= period) {
    let adxSum = 0
    for (let i = 0; i < period; i++) {
      adxSum += dxValues[i]
    }
    let smoothedADX = adxSum / period

    // Fill nulls before first ADX value (period-1 nulls already added, need period-1 more)
    for (let i = period - 1; i < (period - 1) + (period - 1); i++) {
      adx.push(null)
    }

    // Add first ADX value at index 2*period - 2
    adx.push(smoothedADX)

    // Calculate subsequent ADX values using Wilder's smoothing
    for (let i = period; i < dxValues.length; i++) {
      smoothedADX = ((smoothedADX * (period - 1)) + dxValues[i]) / period
      adx.push(smoothedADX)
    }
  } else {
    // Not enough data for ADX
    for (let i = 0; i < dxValues.length; i++) {
      adx.push(null)
    }
  }

  return { adx, plusDI, minusDI }
}

/**
 * Average True Range (ATR)
 * Volatility indicator measuring market volatility
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param period - ATR period (typically 14)
 * @returns ATR values
 */
export function calculateATR(
  highs: number[],
  lows: number[],
  closes: number[],
  period: number = 14
): (number | null)[] {
  if (!highs || !lows || !closes || highs.length < 2) {
    return []
  }

  const trueRanges: number[] = []

  // Calculate True Range for each period
  for (let i = 1; i < closes.length; i++) {
    const high = highs[i]
    const low = lows[i]
    const prevClose = closes[i - 1]

    const tr = Math.max(
      high - low,
      Math.abs(high - prevClose),
      Math.abs(low - prevClose)
    )
    trueRanges.push(tr)
  }

  // ATR is the EMA of True Range
  const atrValues = calculateEMA(trueRanges, period)

  // Add initial null to align with original data length
  return [null, ...atrValues]
}

/**
 * Ichimoku Cloud
 * Comprehensive trend analysis system
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @returns Tenkan, Kijun, Senkou A, Senkou B, Chikou Span
 */
export interface IchimokuResult {
  tenkan: (number | null)[]      // Conversion Line (9-period)
  kijun: (number | null)[]       // Base Line (26-period)
  senkouA: (number | null)[]     // Leading Span A
  senkouB: (number | null)[]     // Leading Span B
  chikou: (number | null)[]      // Lagging Span
}

export function calculateIchimoku(
  highs: number[],
  lows: number[],
  closes: number[],
  tenkanPeriod: number = 9,
  kijunPeriod: number = 26,
  senkouBPeriod: number = 52,
  displacement: number = 26
): IchimokuResult {
  const length = closes.length

  if (!highs || !lows || !closes || length === 0) {
    return { tenkan: [], kijun: [], senkouA: [], senkouB: [], chikou: [] }
  }
  if (lows.length !== length || highs.length !== length) {
    return { tenkan: [], kijun: [], senkouA: [], senkouB: [], chikou: [] }
  }

  const calculateMidpoint = (period: number, index: number): number | null => {
    if (index < period - 1) return null

    let highest = highs[index]
    let lowest = lows[index]
    for (let j = 0; j < period; j++) {
      highest = Math.max(highest, highs[index - j])
      lowest = Math.min(lowest, lows[index - j])
    }
    return (highest + lowest) / 2
  }

  // Calculate base values (without displacement)
  const tenkanBase: (number | null)[] = []
  const kijunBase: (number | null)[] = []
  const senkouABase: (number | null)[] = []
  const senkouBBase: (number | null)[] = []

  for (let i = 0; i < length; i++) {
    // Tenkan-sen (Conversion Line): (9-period high + low) / 2
    tenkanBase.push(calculateMidpoint(tenkanPeriod, i))

    // Kijun-sen (Base Line): (26-period high + low) / 2
    kijunBase.push(calculateMidpoint(kijunPeriod, i))

    // Senkou Span A (Leading Span A): (Tenkan + Kijun) / 2
    const tenkanValue = tenkanBase[i]
    const kijunValue = kijunBase[i]
    if (tenkanValue !== null && kijunValue !== null) {
      senkouABase.push((tenkanValue + kijunValue) / 2)
    } else {
      senkouABase.push(null)
    }

    // Senkou Span B (Leading Span B): (52-period high + low) / 2
    senkouBBase.push(calculateMidpoint(senkouBPeriod, i))
  }

  // Apply displacement
  const tenkan = tenkanBase  // Tenkan: no displacement
  const kijun = kijunBase    // Kijun: no displacement

  // Senkou Span A: shift forward +displacement periods
  const senkouA: (number | null)[] = new Array(displacement).fill(null)
  senkouA.push(...senkouABase.slice(0, length - displacement))

  // Senkou Span B: shift forward +displacement periods
  const senkouB: (number | null)[] = new Array(displacement).fill(null)
  senkouB.push(...senkouBBase.slice(0, length - displacement))

  // Chikou Span: shift backward -displacement periods (current close shown displacement periods back)
  const chikou: (number | null)[] = closes.slice(displacement)
  // Pad the end with nulls (since we can't show future closes)
  while (chikou.length < length) {
    chikou.push(null)
  }

  return { tenkan, kijun, senkouA, senkouB, chikou }
}

/**
 * Commodity Channel Index (CCI)
 * Momentum oscillator
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param period - CCI period (typically 20)
 * @returns CCI values
 */
export function calculateCCI(
  highs: number[],
  lows: number[],
  closes: number[],
  period: number = 20
): (number | null)[] {
  if (!highs || !lows || !closes || highs.length === 0) {
    return []
  }

  const typicalPrices: number[] = []
  for (let i = 0; i < closes.length; i++) {
    typicalPrices.push((highs[i] + lows[i] + closes[i]) / 3)
  }

  const sma = calculateSMA(typicalPrices, period)
  const cci: (number | null)[] = []

  for (let i = 0; i < typicalPrices.length; i++) {
    if (sma[i] === null || i < period - 1) {
      cci.push(null)
      continue
    }

    // Calculate mean deviation
    let sumAbsoluteDeviation = 0
    for (let j = 0; j < period; j++) {
      sumAbsoluteDeviation += Math.abs(typicalPrices[i - j] - sma[i]!)
    }
    const meanDeviation = sumAbsoluteDeviation / period

    if (meanDeviation === 0) {
      cci.push(0)
    } else {
      const cciValue = (typicalPrices[i] - sma[i]!) / (0.015 * meanDeviation)
      cci.push(cciValue)
    }
  }

  return cci
}

/**
 * Williams %R
 * Momentum oscillator (0 to -100 range)
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param period - Period (typically 14)
 * @returns Williams %R values
 */
export function calculateWilliamsR(
  highs: number[],
  lows: number[],
  closes: number[],
  period: number = 14
): (number | null)[] {
  if (!highs || !lows || !closes || highs.length === 0) {
    return []
  }

  const williamsR: (number | null)[] = []

  for (let i = 0; i < closes.length; i++) {
    if (i < period - 1) {
      williamsR.push(null)
      continue
    }

    // Find highest high and lowest low in period
    let highestHigh = highs[i]
    let lowestLow = lows[i]
    for (let j = 0; j < period; j++) {
      highestHigh = Math.max(highestHigh, highs[i - j])
      lowestLow = Math.min(lowestLow, lows[i - j])
    }

    const range = highestHigh - lowestLow
    if (range === 0) {
      williamsR.push(-50) // Neutral if no range
    } else {
      const rValue = ((highestHigh - closes[i]) / range) * -100
      williamsR.push(rValue)
    }
  }

  return williamsR
}

/**
 * Momentum Oscillator (MOM)
 * Measures the absolute rate of price change over a period
 * MOM = Close[today] - Close[n periods ago]
 *
 * Interpretation:
 * - Positive values: Upward momentum
 * - Negative values: Downward momentum
 * - Zero line crossovers indicate potential trend changes
 *
 * @param closes - Array of closing prices
 * @param period - Number of periods to look back (default: 10)
 * @returns Array of Momentum values (null for insufficient data)
 */
export function calculateMomentum(
  closes: number[],
  period: number = 10
): (number | null)[] {
  if (!closes || closes.length === 0) return []
  if (period <= 0 || period >= closes.length) {
    return closes.map(() => null)
  }

  const momentum: (number | null)[] = []

  for (let i = 0; i < closes.length; i++) {
    if (i < period) {
      momentum.push(null)
      continue
    }

    // MOM = Close[i] - Close[i - period]
    const mom = closes[i] - closes[i - period]
    momentum.push(mom)
  }

  return momentum
}

/**
 * Rate of Change (ROC)
 * Measures the percentage price change over a period
 * ROC = ((Close[today] - Close[n]) / Close[n]) * 100
 *
 * Interpretation:
 * - Positive values: Percentage gain
 * - Negative values: Percentage loss
 * - Zero line crossovers indicate trend shifts
 *
 * @param closes - Array of closing prices
 * @param period - Number of periods to look back (default: 12)
 * @returns Array of ROC values (null for insufficient data)
 */
export function calculateROC(
  closes: number[],
  period: number = 12
): (number | null)[] {
  if (!closes || closes.length === 0) return []
  if (period <= 0 || period >= closes.length) {
    return closes.map(() => null)
  }

  const roc: (number | null)[] = []

  for (let i = 0; i < closes.length; i++) {
    if (i < period) {
      roc.push(null)
      continue
    }

    const previousClose = closes[i - period]
    if (previousClose === 0) {
      roc.push(null) // Avoid division by zero
      continue
    }

    // ROC = ((Close[i] - Close[i - period]) / Close[i - period]) * 100
    const rocValue = ((closes[i] - previousClose) / previousClose) * 100
    roc.push(rocValue)
  }

  return roc
}

/**
 * On-Balance Volume (OBV)
 * Cumulative volume indicator that shows buying/selling pressure
 *
 * Algorithm:
 * - If close > previous close: OBV = previous OBV + volume
 * - If close < previous close: OBV = previous OBV - volume
 * - If close = previous close: OBV = previous OBV (unchanged)
 *
 * @param closes - Array of closing prices
 * @param volumes - Array of volume values
 * @returns Array of OBV values
 */
export function calculateOBV(
  closes: number[],
  volumes: number[]
): (number | null)[] {
  if (!closes || !volumes || closes.length === 0 || closes.length !== volumes.length) {
    return []
  }

  const obv: (number | null)[] = []
  let obvValue = 0

  for (let i = 0; i < closes.length; i++) {
    if (i === 0) {
      // First value: start with 0 or first volume
      obvValue = volumes[0]
      obv.push(obvValue)
      continue
    }

    const currentClose = closes[i]
    const previousClose = closes[i - 1]
    const currentVolume = volumes[i]

    if (currentClose > previousClose) {
      // Bullish: add volume
      obvValue += currentVolume
    } else if (currentClose < previousClose) {
      // Bearish: subtract volume
      obvValue -= currentVolume
    }
    // If equal: OBV unchanged (obvValue stays the same)

    obv.push(obvValue)
  }

  return obv
}

/**
 * Calculate Anchored VWAP (Volume Weighted Average Price)
 *
 * Cumulative volume-weighted average from anchor point (startIndex).
 * Does NOT reset daily - anchored to specified index.
 *
 * Formula: VWAP = Σ(Typical Price × Volume) / Σ(Volume)
 * Typical Price = (High + Low + Close) / 3
 *
 * Used for swing trading support/resistance identification.
 * For intraday session VWAP (day trading), use minute data
 * with daily resets (not implemented here).
 *
 * Interpretation:
 * - Price above VWAP = Long-term bullish
 * - Price below VWAP = Long-term bearish
 * - VWAP break with volume = Major trend shift
 *
 * @param highs - Array of high prices (daily bars)
 * @param lows - Array of low prices (daily bars)
 * @param closes - Array of close prices (daily bars)
 * @param volumes - Array of volume values (share count)
 * @param startIndex - Index to start calculation from (default 0 = chart start)
 * @returns Array of Anchored VWAP values (cumulative from startIndex)
 */
export function calculateVWAP(
  highs: number[],
  lows: number[],
  closes: number[],
  volumes: number[],
  startIndex: number = 0
): (number | null)[] {
  if (!highs || !lows || !closes || !volumes || highs.length === 0) {
    return []
  }

  if (highs.length !== lows.length || lows.length !== closes.length || closes.length !== volumes.length) {
    console.error('[VWAP] Array length mismatch:', {
      highs: highs.length,
      lows: lows.length,
      closes: closes.length,
      volumes: volumes.length
    })
    return []
  }

  // Validate and constrain startIndex
  const dataLength = highs.length
  const anchorIndex = Math.max(0, Math.min(startIndex, dataLength - 1))

  // Slice arrays from anchor point
  const slicedHighs = highs.slice(anchorIndex)
  const slicedLows = lows.slice(anchorIndex)
  const slicedCloses = closes.slice(anchorIndex)
  const slicedVolumes = volumes.slice(anchorIndex)

  const vwap: (number | null)[] = []
  let cumulativePV = 0  // Cumulative (Price × Volume)
  let cumulativeVolume = 0  // Cumulative Volume

  for (let i = 0; i < slicedHighs.length; i++) {
    const high = slicedHighs[i]
    const low = slicedLows[i]
    const close = slicedCloses[i]
    const volume = slicedVolumes[i]

    if (high == null || low == null || close == null || volume == null) {
      vwap.push(null)
      continue
    }

    // Typical Price = (High + Low + Close) / 3
    const typicalPrice = (high + low + close) / 3

    // Accumulate
    cumulativePV += typicalPrice * volume
    cumulativeVolume += volume

    // VWAP = Cumulative(Price × Volume) / Cumulative(Volume)
    if (cumulativeVolume > 0) {
      vwap.push(cumulativePV / cumulativeVolume)
    } else {
      vwap.push(null)
    }
  }

  // Pad beginning with nulls to align with original data
  // This ensures VWAP values match up with their corresponding bars
  const padding = new Array(anchorIndex).fill(null)
  return [...padding, ...vwap]
}

/**
 * Calculate MFI (Money Flow Index)
 *
 * MFI = 100 - (100 / (1 + Money Flow Ratio))
 * Money Flow Ratio = (14-period Positive Money Flow) / (14-period Negative Money Flow)
 * Raw Money Flow = Typical Price × Volume
 * Typical Price = (High + Low + Close) / 3
 *
 * Volume-weighted RSI. MFI > 80 = overbought, MFI < 20 = oversold
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param volumes - Array of volume values
 * @param period - Lookback period (default 14)
 * @returns Array of MFI values (0-100 range)
 */
export function calculateMFI(
  highs: number[],
  lows: number[],
  closes: number[],
  volumes: number[],
  period: number = 14
): (number | null)[] {
  if (!highs || !lows || !closes || !volumes || highs.length === 0) {
    return []
  }

  if (highs.length !== lows.length || lows.length !== closes.length || closes.length !== volumes.length) {
    console.error('[MFI] Array length mismatch:', {
      highs: highs.length,
      lows: lows.length,
      closes: closes.length,
      volumes: volumes.length
    })
    return []
  }

  if (highs.length < period + 1) {
    console.log(`[MFI] Insufficient data: ${highs.length} < ${period + 1}`)
    return new Array(highs.length).fill(null)
  }

  const mfi: (number | null)[] = []
  const typicalPrices: number[] = []
  const rawMoneyFlows: number[] = []

  // Calculate typical prices and raw money flows
  for (let i = 0; i < highs.length; i++) {
    const typicalPrice = (highs[i] + lows[i] + closes[i]) / 3
    typicalPrices.push(typicalPrice)
    rawMoneyFlows.push(typicalPrice * volumes[i])
  }

  // Calculate MFI for each period
  for (let i = 0; i < highs.length; i++) {
    if (i < period) {
      mfi.push(null)
      continue
    }

    let positiveFlow = 0
    let negativeFlow = 0

    // Sum positive and negative money flows over the period
    for (let j = i - period + 1; j <= i; j++) {
      if (typicalPrices[j] > typicalPrices[j - 1]) {
        positiveFlow += rawMoneyFlows[j]
      } else if (typicalPrices[j] < typicalPrices[j - 1]) {
        negativeFlow += rawMoneyFlows[j]
      }
      // If prices are equal, no flow is added to either side
    }

    // Calculate Money Flow Ratio and MFI
    if (negativeFlow === 0) {
      // All money flow is positive - MFI = 100
      mfi.push(100)
    } else {
      const moneyFlowRatio = positiveFlow / negativeFlow
      const mfiValue = 100 - (100 / (1 + moneyFlowRatio))
      mfi.push(mfiValue)
    }
  }

  return mfi
}

/**
 * Parabolic SAR (Stop and Reverse) - Trend-following indicator
 * Shows dots above (downtrend) or below (uptrend) price candles
 *
 * Algorithm:
 * - SAR = Prior SAR + AF × (EP - Prior SAR)
 * - EP (Extreme Point) = highest high (uptrend) or lowest low (downtrend)
 * - AF (Acceleration Factor) starts at 0.02, increases by 0.02 per new extreme, max 0.20
 * - Trend reverses when price crosses SAR
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of closing prices
 * @param initialAF - Initial acceleration factor (default 0.02)
 * @param maxAF - Maximum acceleration factor (default 0.20)
 * @returns Array of SAR values with trend direction
 */
export function calculateParabolicSAR(
  highs: number[],
  lows: number[],
  closes: number[],
  initialAF: number = 0.02,
  maxAF: number = 0.20
): Array<{ value: number; isUptrend: boolean } | null> {
  if (!highs || !lows || !closes || highs.length === 0) {
    return []
  }

  if (highs.length !== lows.length || lows.length !== closes.length) {
    console.error('[Parabolic SAR] Array length mismatch:', {
      highs: highs.length,
      lows: lows.length,
      closes: closes.length
    })
    return []
  }

  if (highs.length < 6) {
    console.log(`[Parabolic SAR] Insufficient data: ${highs.length} < 6`)
    return new Array(highs.length).fill(null)
  }

  const sar: Array<{ value: number; isUptrend: boolean } | null> = []

  // Determine initial trend using first 5 bars
  let isUptrend = closes[4] > closes[0]
  let af = initialAF
  let ep = isUptrend ? Math.max(...highs.slice(0, 5)) : Math.min(...lows.slice(0, 5))
  let currentSAR = isUptrend ? Math.min(...lows.slice(0, 5)) : Math.max(...highs.slice(0, 5))

  // Fill first 5 bars with null (not enough data yet)
  for (let i = 0; i < 5; i++) {
    sar.push(null)
  }

  console.log('[Parabolic SAR] Initial trend:', isUptrend ? 'UPTREND' : 'DOWNTREND', 'EP:', ep, 'SAR:', currentSAR)

  // Calculate SAR for each subsequent bar
  for (let i = 5; i < highs.length; i++) {
    // Calculate new SAR
    currentSAR = currentSAR + af * (ep - currentSAR)

    // SAR should not be inside prior two bars
    if (isUptrend) {
      const priorLow = i >= 2 ? Math.min(lows[i - 1], lows[i - 2]) : lows[i - 1]
      currentSAR = Math.min(currentSAR, priorLow)
    } else {
      const priorHigh = i >= 2 ? Math.max(highs[i - 1], highs[i - 2]) : highs[i - 1]
      currentSAR = Math.max(currentSAR, priorHigh)
    }

    // Check for trend reversal
    const reversalOccurred = isUptrend ? lows[i] < currentSAR : highs[i] > currentSAR

    if (reversalOccurred) {
      // Trend reversal
      isUptrend = !isUptrend
      currentSAR = ep  // SAR becomes the old extreme point
      ep = isUptrend ? highs[i] : lows[i]  // New extreme point is current high/low
      af = initialAF  // Reset acceleration factor

      console.log(`[Parabolic SAR] Bar ${i}: REVERSAL - Now ${isUptrend ? 'UPTREND' : 'DOWNTREND'}`)
    } else {
      // Continue current trend
      // Check if new extreme point reached
      const newEP = isUptrend ? highs[i] : lows[i]
      if ((isUptrend && newEP > ep) || (!isUptrend && newEP < ep)) {
        ep = newEP
        af = Math.min(af + initialAF, maxAF)  // Increase AF, but not beyond max
      }
    }

    sar.push({ value: currentSAR, isUptrend })
  }

  console.log(`[Parabolic SAR] Calculated ${sar.length} points, ${sar.filter(s => s !== null).length} non-null`)
  return sar
}

/**
 * Fibonacci Retracement Levels - Key support/resistance levels
 * Based on Fibonacci ratios (23.6%, 38.2%, 50%, 61.8%, 78.6%)
 *
 * Calculates horizontal price levels between recent high and low
 * Used to identify potential reversal points in price movements
 *
 * Formula:
 * - Find highest high and lowest low in lookback period
 * - Range = High - Low
 * - For each Fibonacci ratio: Level = High - (Range × Ratio)
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param lookbackPeriod - Number of periods to find high/low (default 50)
 * @returns Object with high, low, range, and 5 Fibonacci levels
 */
export interface FibonacciLevel {
  ratio: number
  percentage: string
  price: number
  label: string
}

export interface FibonacciLevels {
  high: number
  low: number
  range: number
  levels: FibonacciLevel[]
  highIndex: number
  lowIndex: number
}

export function calculateFibonacciRetracement(
  highs: number[],
  lows: number[],
  lookbackPeriod: number = 50
): FibonacciLevels | null {
  if (!highs || !lows || highs.length === 0) {
    console.error('[Fibonacci] No data provided')
    return null
  }

  if (highs.length !== lows.length) {
    console.error('[Fibonacci] Array length mismatch:', {
      highs: highs.length,
      lows: lows.length
    })
    return null
  }

  if (highs.length < lookbackPeriod) {
    console.log(`[Fibonacci] Insufficient data: ${highs.length} < ${lookbackPeriod}`)
    return null
  }

  // Get recent data based on lookback period
  const recentHighs = highs.slice(-lookbackPeriod)
  const recentLows = lows.slice(-lookbackPeriod)

  // Find highest high and lowest low
  const high = Math.max(...recentHighs)
  const low = Math.min(...recentLows)
  const range = high - low

  if (range === 0) {
    console.warn('[Fibonacci] No price range (high === low)')
    return null
  }

  // Find indices for reference
  const highIndex = highs.length - lookbackPeriod + recentHighs.indexOf(high)
  const lowIndex = highs.length - lookbackPeriod + recentLows.indexOf(low)

  // Calculate Fibonacci retracement levels
  // Standard Fibonacci ratios used in trading
  const fibRatios = [
    { ratio: 0, label: '0.0% (High)' },      // 0% = Swing high
    { ratio: 0.236, label: '23.6%' },        // First retracement
    { ratio: 0.382, label: '38.2%' },        // Key level
    { ratio: 0.500, label: '50.0%' },        // Midpoint (not Fibonacci but widely used)
    { ratio: 0.618, label: '61.8%' },        // Golden ratio - most important
    { ratio: 0.786, label: '78.6%' },        // Deep retracement
    { ratio: 1.000, label: '100% (Low)' }    // 100% = Swing low
  ]

  const levels: FibonacciLevel[] = fibRatios.map(fib => ({
    ratio: fib.ratio,
    percentage: fib.label,
    price: high - (range * fib.ratio),
    label: fib.label
  }))

  console.log('[Fibonacci] Calculated levels:', {
    high,
    low,
    range,
    lookbackPeriod,
    highIndex,
    lowIndex,
    levelCount: levels.length
  })

  return {
    high,
    low,
    range,
    levels,
    highIndex,
    lowIndex
  }
}

/**
 * Donchian Channels
 * Breakout indicator based on highest high and lowest low over period
 *
 * Upper Channel = Highest high over period
 * Lower Channel = Lowest low over period
 * Middle Channel = (Upper + Lower) / 2
 *
 * Trading signals:
 * - Price breaking above upper channel = bullish breakout
 * - Price breaking below lower channel = bearish breakout
 * - Price at middle channel = neutral/ranging
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param period - Lookback period (default: 20)
 * @returns Object with upper, middle, lower channel arrays
 */
export interface DonchianChannels {
  upper: (number | null)[]
  middle: (number | null)[]
  lower: (number | null)[]
}

export function calculateDonchianChannels(
  highs: number[],
  lows: number[],
  period: number = 20
): DonchianChannels {
  if (!highs || highs.length === 0 || !lows || lows.length === 0) {
    return { upper: [], middle: [], lower: [] }
  }

  if (highs.length !== lows.length) {
    console.error('[Donchian] Highs and lows arrays must be same length')
    return { upper: [], middle: [], lower: [] }
  }

  if (period <= 0 || period > highs.length) {
    return {
      upper: highs.map(() => null),
      middle: highs.map(() => null),
      lower: highs.map(() => null)
    }
  }

  const upper: (number | null)[] = []
  const middle: (number | null)[] = []
  const lower: (number | null)[] = []

  for (let i = 0; i < highs.length; i++) {
    if (i < period - 1) {
      upper.push(null)
      middle.push(null)
      lower.push(null)
      continue
    }

    // Calculate highest high and lowest low over period
    let highestHigh = -Infinity
    let lowestLow = Infinity

    for (let j = 0; j < period; j++) {
      const idx = i - j
      if (highs[idx] > highestHigh) {
        highestHigh = highs[idx]
      }
      if (lows[idx] < lowestLow) {
        lowestLow = lows[idx]
      }
    }

    upper.push(highestHigh)
    lower.push(lowestLow)
    middle.push((highestHigh + lowestLow) / 2)
  }

  console.log('[Donchian] Calculated:', {
    period,
    points: upper.filter(v => v !== null).length,
    dataLength: highs.length
  })

  return { upper, middle, lower }
}

/**
 * Keltner Channels
 * ATR-based volatility bands around EMA centerline
 *
 * Middle Channel = EMA of close prices
 * Upper Channel = EMA + (ATR × multiplier)
 * Lower Channel = EMA - (ATR × multiplier)
 *
 * Similar to Bollinger Bands but uses ATR instead of standard deviation.
 * More responsive to volatility changes.
 *
 * Trading signals:
 * - Price at upper channel = potential overbought (take profit/short)
 * - Price at lower channel = potential oversold (buy signal)
 * - Expanding channels = increasing volatility
 * - Contracting channels = decreasing volatility (breakout pending)
 *
 * @param highs - Array of high prices
 * @param lows - Array of low prices
 * @param closes - Array of close prices
 * @param period - EMA and ATR period (default: 20)
 * @param multiplier - ATR multiplier for channel width (default: 2)
 * @returns Object with upper, middle, lower channel arrays
 */
export interface KeltnerChannels {
  upper: (number | null)[]
  middle: (number | null)[]
  lower: (number | null)[]
}

export function calculateKeltnerChannels(
  highs: number[],
  lows: number[],
  closes: number[],
  period: number = 20,
  multiplier: number = 2
): KeltnerChannels {
  if (!highs || highs.length === 0 || !lows || lows.length === 0 || !closes || closes.length === 0) {
    return { upper: [], middle: [], lower: [] }
  }

  if (highs.length !== lows.length || highs.length !== closes.length) {
    console.error('[Keltner] Highs, lows, and closes arrays must be same length')
    return { upper: [], middle: [], lower: [] }
  }

  if (period <= 0 || period > closes.length) {
    return {
      upper: closes.map(() => null),
      middle: closes.map(() => null),
      lower: closes.map(() => null)
    }
  }

  // Calculate EMA of close prices (middle channel)
  const emaValues = calculateEMA(closes, period)

  // Calculate ATR for volatility bands
  const atrValues = calculateATR(highs, lows, closes, period)

  const upper: (number | null)[] = []
  const middle: (number | null)[] = []
  const lower: (number | null)[] = []

  for (let i = 0; i < closes.length; i++) {
    if (emaValues[i] === null || atrValues[i] === null) {
      upper.push(null)
      middle.push(null)
      lower.push(null)
      continue
    }

    const ema = emaValues[i]!
    const atr = atrValues[i]!

    middle.push(ema)
    upper.push(ema + (atr * multiplier))
    lower.push(ema - (atr * multiplier))
  }

  console.log('[Keltner] Calculated:', {
    period,
    multiplier,
    points: upper.filter(v => v !== null).length,
    dataLength: closes.length
  })

  return { upper, middle, lower }
}
