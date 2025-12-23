/**
 * MACD (Moving Average Convergence Divergence) Indicator
 * 
 * Professional implementation of the MACD indicator for technical analysis.
 * MACD = 12-day EMA - 26-day EMA
 * Signal Line = 9-day EMA of MACD
 * Histogram = MACD - Signal Line
 */

import { IndicatorValue } from './indicator-registry'

export interface MACDConfig {
  fastPeriod?: number    // Default: 12
  slowPeriod?: number    // Default: 26
  signalPeriod?: number  // Default: 9
  source?: 'close' | 'open' | 'high' | 'low' | 'hl2' | 'hlc3' | 'ohlc4'
}

export interface MACDResult {
  macd: number | null
  signal: number | null
  histogram: number | null
  timestamp?: number
}

export interface MACDSeries {
  macd: (number | null)[]
  signal: (number | null)[]
  histogram: (number | null)[]
}

/**
 * Calculate EMA (Exponential Moving Average)
 */
function calculateEMA(data: number[], period: number): (number | null)[] {
  if (data.length === 0) return []
  
  const multiplier = 2 / (period + 1)
  const ema: (number | null)[] = new Array(data.length).fill(null)
  
  // Need at least 'period' data points to start
  if (data.length < period) {
    return ema
  }
  
  // Calculate initial SMA for the first EMA value
  let sum = 0
  for (let i = 0; i < period; i++) {
    sum += data[i]
  }
  ema[period - 1] = sum / period
  
  // Calculate EMA for remaining values
  for (let i = period; i < data.length; i++) {
    if (ema[i - 1] !== null) {
      ema[i] = (data[i] - ema[i - 1]!) * multiplier + ema[i - 1]!
    }
  }
  
  return ema
}

/**
 * Get price based on source type
 */
function getSourcePrice(
  candle: { open: number; high: number; low: number; close: number },
  source: string
): number {
  switch (source) {
    case 'open':
      return candle.open
    case 'high':
      return candle.high
    case 'low':
      return candle.low
    case 'hl2':
      return (candle.high + candle.low) / 2
    case 'hlc3':
      return (candle.high + candle.low + candle.close) / 3
    case 'ohlc4':
      return (candle.open + candle.high + candle.low + candle.close) / 4
    case 'close':
    default:
      return candle.close
  }
}

/**
 * Calculate MACD for a series of price data
 */
export function calculateMACD(
  data: { open: number; high: number; low: number; close: number; time: number }[],
  config: MACDConfig = {}
): MACDSeries {
  const {
    fastPeriod = 12,
    slowPeriod = 26,
    signalPeriod = 9,
    source = 'close'
  } = config
  
  // Extract price values based on source
  const prices = data.map(candle => getSourcePrice(candle, source))
  
  // Calculate EMAs
  const fastEMA = calculateEMA(prices, fastPeriod)
  const slowEMA = calculateEMA(prices, slowPeriod)
  
  // Calculate MACD line (fast EMA - slow EMA)
  const macdLine: (number | null)[] = []
  for (let i = 0; i < data.length; i++) {
    if (fastEMA[i] !== null && slowEMA[i] !== null) {
      macdLine.push(fastEMA[i]! - slowEMA[i]!)
    } else {
      macdLine.push(null)
    }
  }
  
  // Calculate Signal line (EMA of MACD)
  const validMACDValues = macdLine.filter(v => v !== null) as number[]
  const signalEMA = calculateEMA(validMACDValues, signalPeriod)
  
  // Map signal line back to full length
  const signalLine: (number | null)[] = new Array(data.length).fill(null)
  let signalIndex = 0
  for (let i = 0; i < data.length; i++) {
    if (macdLine[i] !== null) {
      signalLine[i] = signalEMA[signalIndex] || null
      signalIndex++
    }
  }
  
  // Calculate Histogram (MACD - Signal)
  const histogram: (number | null)[] = []
  for (let i = 0; i < data.length; i++) {
    if (macdLine[i] !== null && signalLine[i] !== null) {
      histogram.push(macdLine[i]! - signalLine[i]!)
    } else {
      histogram.push(null)
    }
  }
  
  return {
    macd: macdLine,
    signal: signalLine,
    histogram
  }
}

/**
 * Calculate MACD for a single data point (for real-time updates)
 */
export function calculateMACDPoint(
  currentPrice: number,
  previousMACD: MACDResult,
  previousFastEMA: number,
  previousSlowEMA: number,
  config: MACDConfig = {}
): {
  result: MACDResult
  fastEMA: number
  slowEMA: number
} {
  const {
    fastPeriod = 12,
    slowPeriod = 26,
    signalPeriod = 9
  } = config
  
  // Calculate multipliers
  const fastMultiplier = 2 / (fastPeriod + 1)
  const slowMultiplier = 2 / (slowPeriod + 1)
  const signalMultiplier = 2 / (signalPeriod + 1)
  
  // Update EMAs
  const fastEMA = (currentPrice - previousFastEMA) * fastMultiplier + previousFastEMA
  const slowEMA = (currentPrice - previousSlowEMA) * slowMultiplier + previousSlowEMA
  
  // Calculate MACD
  const macd = fastEMA - slowEMA
  
  // Update Signal line
  let signal = previousMACD.signal
  if (signal !== null) {
    signal = (macd - signal) * signalMultiplier + signal
  } else {
    signal = macd // Initialize signal with MACD value
  }
  
  // Calculate Histogram
  const histogram = macd - signal
  
  return {
    result: {
      macd,
      signal,
      histogram,
      timestamp: Date.now()
    },
    fastEMA,
    slowEMA
  }
}

/**
 * Detect MACD crossovers
 */
export function detectMACDCrossover(
  currentMACD: MACDResult,
  previousMACD: MACDResult
): 'bullish' | 'bearish' | null {
  if (
    currentMACD.macd === null || 
    currentMACD.signal === null ||
    previousMACD.macd === null ||
    previousMACD.signal === null
  ) {
    return null
  }
  
  // Bullish crossover: MACD crosses above signal
  if (previousMACD.macd <= previousMACD.signal && currentMACD.macd > currentMACD.signal) {
    return 'bullish'
  }
  
  // Bearish crossover: MACD crosses below signal
  if (previousMACD.macd >= previousMACD.signal && currentMACD.macd < currentMACD.signal) {
    return 'bearish'
  }
  
  return null
}

/**
 * Detect MACD divergence
 */
export function detectMACDDivergence(
  prices: number[],
  macdValues: (number | null)[],
  lookback: number = 20
): 'bullish' | 'bearish' | null {
  // Need at least lookback + 2 points
  if (prices.length < lookback + 2 || macdValues.length < lookback + 2) {
    return null
  }
  
  const recentPrices = prices.slice(-lookback)
  const recentMACD = macdValues.slice(-lookback).filter(v => v !== null) as number[]
  
  if (recentMACD.length < 2) {
    return null
  }
  
  // Find price trend
  const priceStart = recentPrices[0]
  const priceEnd = recentPrices[recentPrices.length - 1]
  const priceTrend = priceEnd > priceStart ? 'up' : 'down'
  
  // Find MACD trend
  const macdStart = recentMACD[0]
  const macdEnd = recentMACD[recentMACD.length - 1]
  const macdTrend = macdEnd > macdStart ? 'up' : 'down'
  
  // Bullish divergence: price down, MACD up
  if (priceTrend === 'down' && macdTrend === 'up') {
    return 'bullish'
  }
  
  // Bearish divergence: price up, MACD down
  if (priceTrend === 'up' && macdTrend === 'down') {
    return 'bearish'
  }
  
  return null
}

/**
 * Format MACD values for display
 */
export function formatMACDValue(value: number | null, decimals: number = 4): string {
  if (value === null) return 'N/A'
  return value.toFixed(decimals)
}

/**
 * Get MACD histogram color based on value and trend
 */
export function getMACDHistogramColor(
  currentValue: number | null,
  previousValue: number | null
): string {
  if (currentValue === null) return '#888888'
  
  // Green for positive, red for negative
  const baseColor = currentValue >= 0 ? '#26a69a' : '#ef5350'
  
  // Adjust intensity based on trend
  if (previousValue !== null) {
    const increasing = currentValue > previousValue
    if (increasing && currentValue >= 0) return '#4caf50' // Bright green
    if (!increasing && currentValue < 0) return '#f44336' // Bright red
  }
  
  return baseColor
}

/**
 * Register MACD indicator with the indicator registry
 */
export function registerMACDIndicator(): void {
  // This will be called when integrating with the indicator registry
  // Implementation depends on the registry API
}

export default {
  calculateMACD,
  calculateMACDPoint,
  detectMACDCrossover,
  detectMACDDivergence,
  formatMACDValue,
  getMACDHistogramColor
}