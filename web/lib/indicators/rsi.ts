/**
 * RSI (Relative Strength Index) Indicator
 * 
 * Professional implementation of the RSI indicator for technical analysis.
 * RSI = 100 - (100 / (1 + RS))
 * RS = Average Gain / Average Loss
 */

import { IndicatorValue } from './indicator-registry'

export interface RSIConfig {
  period?: number    // Default: 14
  overbought?: number  // Default: 70
  oversold?: number    // Default: 30
  source?: 'close' | 'open' | 'high' | 'low' | 'hl2' | 'hlc3' | 'ohlc4'
}

export interface RSIResult {
  rsi: number | null
  timestamp?: number
}

export interface RSISeries {
  rsi: (number | null)[]
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
 * Calculate RSI using the Wilder's smoothing method
 */
export function calculateRSI(
  data: { open: number; high: number; low: number; close: number; time: number }[],
  config: RSIConfig = {}
): RSISeries {
  const {
    period = 14,
    source = 'close'
  } = config
  
  if (data.length < period + 1) {
    return { rsi: new Array(data.length).fill(null) }
  }
  
  // Extract price values based on source
  const prices = data.map(candle => getSourcePrice(candle, source))
  
  // Calculate price changes
  const changes: number[] = []
  for (let i = 1; i < prices.length; i++) {
    changes.push(prices[i] - prices[i - 1])
  }
  
  // Calculate initial average gain and loss
  let avgGain = 0
  let avgLoss = 0
  
  for (let i = 0; i < period; i++) {
    if (changes[i] > 0) {
      avgGain += changes[i]
    } else {
      avgLoss += Math.abs(changes[i])
    }
  }
  
  avgGain /= period
  avgLoss /= period
  
  // Initialize RSI array
  const rsiValues: (number | null)[] = new Array(data.length).fill(null)
  
  // Calculate first RSI value
  let rs = avgLoss === 0 ? 100 : avgGain / avgLoss
  rsiValues[period] = 100 - (100 / (1 + rs))
  
  // Calculate subsequent RSI values using Wilder's smoothing
  for (let i = period; i < changes.length; i++) {
    const change = changes[i]
    const gain = change > 0 ? change : 0
    const loss = change < 0 ? Math.abs(change) : 0
    
    // Wilder's smoothing: multiply previous average by (period - 1), add current value, divide by period
    avgGain = (avgGain * (period - 1) + gain) / period
    avgLoss = (avgLoss * (period - 1) + loss) / period
    
    rs = avgLoss === 0 ? 100 : avgGain / avgLoss
    rsiValues[i + 1] = 100 - (100 / (1 + rs))
  }
  
  return { rsi: rsiValues }
}

/**
 * Calculate RSI for a single data point (for real-time updates)
 */
export function calculateRSIPoint(
  currentPrice: number,
  previousPrice: number,
  previousAvgGain: number,
  previousAvgLoss: number,
  config: RSIConfig = {}
): {
  result: RSIResult
  avgGain: number
  avgLoss: number
} {
  const { period = 14 } = config
  
  const change = currentPrice - previousPrice
  const gain = change > 0 ? change : 0
  const loss = change < 0 ? Math.abs(change) : 0
  
  // Wilder's smoothing
  const avgGain = (previousAvgGain * (period - 1) + gain) / period
  const avgLoss = (previousAvgLoss * (period - 1) + loss) / period
  
  const rs = avgLoss === 0 ? 100 : avgGain / avgLoss
  const rsi = 100 - (100 / (1 + rs))
  
  return {
    result: {
      rsi,
      timestamp: Date.now()
    },
    avgGain,
    avgLoss
  }
}

/**
 * Detect RSI divergence
 */
export function detectRSIDivergence(
  prices: number[],
  rsiValues: (number | null)[],
  lookback: number = 20
): 'bullish' | 'bearish' | null {
  // Need at least lookback + 2 points
  if (prices.length < lookback + 2 || rsiValues.length < lookback + 2) {
    return null
  }
  
  const recentPrices = prices.slice(-lookback)
  const recentRSI = rsiValues.slice(-lookback).filter(v => v !== null) as number[]
  
  if (recentRSI.length < 2) {
    return null
  }
  
  // Find price trend
  const priceStart = recentPrices[0]
  const priceEnd = recentPrices[recentPrices.length - 1]
  const priceTrend = priceEnd > priceStart ? 'up' : 'down'
  
  // Find RSI trend
  const rsiStart = recentRSI[0]
  const rsiEnd = recentRSI[recentRSI.length - 1]
  const rsiTrend = rsiEnd > rsiStart ? 'up' : 'down'
  
  // Bullish divergence: price down, RSI up
  if (priceTrend === 'down' && rsiTrend === 'up') {
    return 'bullish'
  }
  
  // Bearish divergence: price up, RSI down
  if (priceTrend === 'up' && rsiTrend === 'down') {
    return 'bearish'
  }
  
  return null
}

/**
 * Determine RSI signal based on value and levels
 */
export function getRSISignal(
  rsi: number | null,
  config: RSIConfig = {}
): 'overbought' | 'oversold' | 'neutral' | null {
  if (rsi === null) return null
  
  const {
    overbought = 70,
    oversold = 30
  } = config
  
  if (rsi >= overbought) return 'overbought'
  if (rsi <= oversold) return 'oversold'
  return 'neutral'
}

/**
 * Format RSI value for display
 */
export function formatRSIValue(value: number | null, decimals: number = 2): string {
  if (value === null) return 'N/A'
  return value.toFixed(decimals)
}

/**
 * Get RSI color based on value and levels
 */
export function getRSIColor(
  rsi: number | null,
  config: RSIConfig = {}
): string {
  if (rsi === null) return '#888888'
  
  const {
    overbought = 70,
    oversold = 30
  } = config
  
  if (rsi >= overbought) return '#ef5350' // Red - Overbought
  if (rsi <= oversold) return '#26a69a' // Green - Oversold
  return '#2962ff' // Blue - Neutral
}

/**
 * Register RSI indicator with the indicator registry
 */
export function registerRSIIndicator(): void {
  // This will be called when integrating with the indicator registry
  // Implementation depends on the registry API
}

export default {
  calculateRSI,
  calculateRSIPoint,
  detectRSIDivergence,
  getRSISignal,
  formatRSIValue,
  getRSIColor
}