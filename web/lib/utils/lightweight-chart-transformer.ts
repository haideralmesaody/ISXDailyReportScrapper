/**
 * Data transformation utilities for Lightweight Charts
 * Converts ISX ticker data to Lightweight Charts format with technical indicators
 * 
 * @example Basic Usage
 * ```typescript
 * import { transformToLightweightData, calculateSMA, formatPrice } from '@/lib/utils/lightweight-chart-transformer'
 * 
 * // Transform ISX data to Lightweight Charts format
 * const { candlestick, volume } = transformToLightweightData(tickerData)
 * 
 * // Calculate technical indicators
 * const sma20 = calculateSMA(candlestick, 20)
 * const ema20 = calculateEMA(candlestick, 20)
 * const rsi = calculateRSI(candlestick, 14)
 * 
 * // Format values for display
 * const formattedPrice = formatPrice(1.555) // "1.55"
 * const formattedVolume = formatVolume(1500000) // "1.50M"
 * ```
 * 
 * @example Advanced Usage with Support/Resistance
 * ```typescript
 * import { calculateResistanceLevels, findKeyLevels } from '@/lib/utils/lightweight-chart-transformer'
 * 
 * // Calculate key price levels
 * const { resistance, support } = calculateResistanceLevels(candlestick)
 * const { supportLevels, resistanceLevels } = findKeyLevels(candlestick)
 * 
 * // Determine trend direction
 * const trend = determineTrend(candlestick, 10) // 'up' | 'down' | 'sideways'
 * ```
 * 
 * Features:
 * - Transforms ISX TickerHistoricalData to Lightweight Charts format
 * - Handles date conversion to UTCTimestamp (seconds)
 * - Gracefully handles missing/invalid data with fallbacks
 * - Technical indicators: SMA, EMA, RSI, MACD, Bollinger Bands
 * - Support/resistance level calculations
 * - Price and volume formatting utilities
 * - Trend analysis functions
 */

import type { TickerHistoricalData } from '@/types/analysis'
import type { 
  CandlestickData, 
  HistogramData, 
  LineData, 
  UTCTimestamp 
} from 'lightweight-charts'

/**
 * Transform ticker historical data to Lightweight Charts format
 */
export function transformToLightweightData(data: TickerHistoricalData[]) {
  const candlestick: CandlestickData[] = []
  const volume: HistogramData[] = []
  let hasDataIssue = false
  
  // Debug: Log input data info
  console.log(`[Chart Transform] Input data: ${data.length} entries`)
  if (data.length > 0) {
    console.log(`[Chart Transform] Date range: ${data[0].date} to ${data[data.length - 1].date}`)
    const septData = data.filter(d => d.date && d.date.startsWith('2025-09'))
    if (septData.length > 0) {
      console.log(`[Chart Transform] September 2025 entries: ${septData.length}`)
    }
  }
  
  // Sort data by date to ensure chronological order
  const sortedData = [...data].sort((a, b) => {
    // Parse dates as UTC for proper sorting
    const [yearA, monthA, dayA] = a.date.split('-').map(Number)
    const [yearB, monthB, dayB] = b.date.split('-').map(Number)
    const dateA = Date.UTC(yearA, monthA - 1, dayA)
    const dateB = Date.UTC(yearB, monthB - 1, dayB)
    return dateA - dateB
  })
  
  sortedData.forEach(item => {
    // Convert date to UTCTimestamp (seconds, not milliseconds)
    // Parse date as UTC to avoid timezone shifts
    const [year, month, day] = item.date.split('-').map(Number)
    const time = (Date.UTC(year, month - 1, day) / 1000) as UTCTimestamp
    
    // Ensure we have valid OHLC values - they're already rounded from API
    // Use close as fallback for missing values, or a default of 1.00
    const close = item.close > 0 ? item.close : (item.open > 0 ? item.open : 1.00)
    const open = item.open > 0 ? item.open : close
    const high = item.high > 0 ? item.high : Math.max(open, close)
    const low = item.low > 0 ? item.low : Math.min(open, close)
    
    // Warn once if data had to be corrected
    if (!hasDataIssue && (item.close <= 0 || item.open <= 0 || item.high <= 0 || item.low <= 0)) {
      console.warn('Chart data contains invalid prices (zeros), using fallback values')
      hasDataIssue = true
    }
    
    // Add candlestick data
    candlestick.push({
      time,
      open,
      high,
      low,
      close
    })
    
    // Add volume data
    // In ISX, 'value' represents monetary value traded (IQD)
    // 'volume' represents share count - we prefer value for more meaningful analysis
    const volumeValue = item.value || item.volume || 0
    const color = close >= open ? '#10b981' : '#ef4444'
    
    volume.push({
      time,
      value: volumeValue,
      color
    })
  })
  
  // Debug: Log output data info
  console.log(`[Chart Transform] Output: ${candlestick.length} candlesticks`)
  if (candlestick.length > 0) {
    const lastCandles = candlestick.slice(-3)
    console.log(`[Chart Transform] Last 3 candles:`)
    lastCandles.forEach(c => {
      const date = new Date(c.time * 1000).toISOString().split('T')[0]
      console.log(`  ${date}: O=${c.open} H=${c.high} L=${c.low} C=${c.close}`)
    })
  }
  
  return { candlestick, volume }
}

/**
 * Calculate Simple Moving Average
 */
export function calculateSMA(data: CandlestickData[], period: number): LineData[] {
  const sma: LineData[] = []
  
  if (data.length < period) return sma
  
  for (let i = period - 1; i < data.length; i++) {
    let sum = 0
    for (let j = 0; j < period; j++) {
      sum += data[i - j].close
    }
    
    sma.push({
      time: data[i].time,
      value: Math.round((sum / period) * 100) / 100
    })
  }
  
  return sma
}

/**
 * Calculate Exponential Moving Average
 */
export function calculateEMA(data: CandlestickData[], period: number): LineData[] {
  const ema: LineData[] = []
  
  if (data.length < period) return ema
  
  const multiplier = 2 / (period + 1)
  
  // Start with SMA for the first value
  let sum = 0
  for (let i = 0; i < period; i++) {
    sum += data[i].close
  }
  let previousEMA = sum / period
  
  ema.push({
    time: data[period - 1].time,
    value: Math.round(previousEMA * 100) / 100
  })
  
  // Calculate EMA for remaining values
  for (let i = period; i < data.length; i++) {
    const currentEMA = (data[i].close - previousEMA) * multiplier + previousEMA
    ema.push({
      time: data[i].time,
      value: Math.round(currentEMA * 100) / 100
    })
    previousEMA = currentEMA
  }
  
  return ema
}

/**
 * Calculate resistance and support levels using pivot points
 */
export function calculateResistanceLevels(data: CandlestickData[]): { 
  resistance: number
  support: number 
} {
  if (data.length === 0) {
    return { resistance: 0, support: 0 }
  }
  
  // Use last 20 periods or all data if less than 20
  const lookbackPeriod = Math.min(20, data.length)
  const recentData = data.slice(-lookbackPeriod)
  
  if (recentData.length === 0 || !recentData[0]) {
    return { resistance: 0, support: 0 }
  }
  
  // Find highest and lowest prices in recent data
  let highestHigh = recentData[0].high
  let lowestLow = recentData[0].low
  let totalHigh = 0
  let totalLow = 0
  let totalClose = 0
  
  recentData.forEach(candle => {
    if (!candle) return
    highestHigh = Math.max(highestHigh, candle.high)
    lowestLow = Math.min(lowestLow, candle.low)
    totalHigh += candle.high
    totalLow += candle.low
    totalClose += candle.close
  })
  
  // Calculate pivot point
  const avgHigh = totalHigh / recentData.length
  const avgLow = totalLow / recentData.length
  const avgClose = totalClose / recentData.length
  const pivot = (avgHigh + avgLow + avgClose) / 3
  
  // Calculate resistance and support levels
  const resistance = Math.max(pivot + (avgHigh - avgLow), highestHigh * 0.98)
  const support = Math.min(pivot - (avgHigh - avgLow), lowestLow * 1.02)
  
  return {
    resistance: Math.round(resistance * 100) / 100,
    support: Math.round(support * 100) / 100
  }
}

/**
 * Format price for display with appropriate decimal places
 */
export function formatPrice(price: number): string {
  if (price === 0) return '0.00'
  
  // ISX prices are typically in IQD with 2 decimal places
  return price.toFixed(2)
}

/**
 * Format volume for display with appropriate units
 */
export function formatVolume(volume: number): string {
  if (volume === 0) return '0'
  
  // Format large numbers with units
  if (volume >= 1e9) {
    return `${(volume / 1e9).toFixed(2)}B`
  }
  if (volume >= 1e6) {
    return `${(volume / 1e6).toFixed(2)}M`
  }
  if (volume >= 1e3) {
    return `${(volume / 1e3).toFixed(2)}K`
  }
  
  return volume.toLocaleString()
}

/**
 * Calculate RSI (Relative Strength Index)
 */
export function calculateRSI(data: CandlestickData[], period: number = 14): LineData[] {
  const rsi: LineData[] = []
  
  if (data.length < period + 1) return rsi
  
  // Calculate price changes
  const changes: number[] = []
  for (let i = 1; i < data.length; i++) {
    changes.push(data[i].close - data[i - 1].close)
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
  
  // Calculate RSI values
  for (let i = period; i < changes.length; i++) {
    const change = changes[i]
    
    if (change > 0) {
      avgGain = (avgGain * (period - 1) + change) / period
      avgLoss = (avgLoss * (period - 1)) / period
    } else {
      avgGain = (avgGain * (period - 1)) / period
      avgLoss = (avgLoss * (period - 1) + Math.abs(change)) / period
    }
    
    const rs = avgLoss === 0 ? 100 : avgGain / avgLoss
    const rsiValue = 100 - (100 / (1 + rs))
    
    rsi.push({
      time: data[i + 1].time,
      value: Math.round(rsiValue * 100) / 100
    })
  }
  
  return rsi
}

/**
 * Calculate Bollinger Bands
 */
export function calculateBollingerBands(data: CandlestickData[], period: number = 20, stdDev: number = 2): {
  upper: LineData[]
  middle: LineData[]
  lower: LineData[]
} {
  const upper: LineData[] = []
  const middle: LineData[] = []
  const lower: LineData[] = []
  
  if (data.length < period) return { upper, middle, lower }
  
  for (let i = period - 1; i < data.length; i++) {
    // Calculate SMA (middle band)
    let sum = 0
    for (let j = 0; j < period; j++) {
      sum += data[i - j].close
    }
    const sma = sum / period
    
    // Calculate standard deviation
    let squaredDifferences = 0
    for (let j = 0; j < period; j++) {
      squaredDifferences += Math.pow(data[i - j].close - sma, 2)
    }
    const standardDeviation = Math.sqrt(squaredDifferences / period)
    
    const time = data[i].time
    const upperValue = sma + (standardDeviation * stdDev)
    const lowerValue = sma - (standardDeviation * stdDev)
    
    upper.push({
      time,
      value: Math.round(upperValue * 100) / 100
    })
    
    middle.push({
      time,
      value: Math.round(sma * 100) / 100
    })
    
    lower.push({
      time,
      value: Math.round(lowerValue * 100) / 100
    })
  }
  
  return { upper, middle, lower }
}

/**
 * Calculate MACD (Moving Average Convergence Divergence)
 */
export function calculateMACD(data: CandlestickData[]): {
  macdLine: LineData[]
  signalLine: LineData[]
  histogram: HistogramData[]
} {
  const macdLine: LineData[] = []
  const signalLine: LineData[] = []
  const histogram: HistogramData[] = []
  
  if (data.length < 26) return { macdLine, signalLine, histogram }
  
  // Calculate EMA 12 and EMA 26
  const ema12 = calculateEMAValues(data, 12)
  const ema26 = calculateEMAValues(data, 26)
  
  // Calculate MACD line (EMA12 - EMA26)
  const macdValues: number[] = []
  for (let i = 25; i < data.length; i++) {
    const macd = ema12[i] - ema26[i]
    macdValues.push(macd)
    macdLine.push({
      time: data[i].time,
      value: Math.round(macd * 100) / 100
    })
  }
  
  // Calculate Signal line (9-period EMA of MACD)
  if (macdValues.length >= 9) {
    const signalMultiplier = 2 / (9 + 1)
    let previousSignal = macdValues.slice(0, 9).reduce((a, b) => a + b, 0) / 9
    
    for (let i = 8; i < macdValues.length; i++) {
      const currentSignal = (macdValues[i] - previousSignal) * signalMultiplier + previousSignal
      const histValue = macdValues[i] - currentSignal
      
      signalLine.push({
        time: data[i + 25].time,
        value: Math.round(currentSignal * 100) / 100
      })
      
      histogram.push({
        time: data[i + 25].time,
        value: Math.round(histValue * 100) / 100,
        color: histValue >= 0 ? '#10b981' : '#ef4444'
      })
      
      previousSignal = currentSignal
    }
  }
  
  return { macdLine, signalLine, histogram }
}

/**
 * Helper function to calculate EMA values (not formatted for Lightweight Charts)
 */
function calculateEMAValues(data: CandlestickData[], period: number): number[] {
  const ema: number[] = new Array(data.length).fill(0)
  const multiplier = 2 / (period + 1)
  
  if (data.length < period) return ema
  
  // Start with SMA
  let sum = 0
  for (let i = 0; i < period; i++) {
    sum += data[i].close
  }
  ema[period - 1] = sum / period
  
  // Calculate EMA
  for (let i = period; i < data.length; i++) {
    ema[i] = (data[i].close - ema[i - 1]) * multiplier + ema[i - 1]
  }
  
  return ema
}

/**
 * Calculate percentage change between two values
 */
export function calculatePercentageChange(oldValue: number, newValue: number): number {
  if (oldValue === 0) return 0
  return ((newValue - oldValue) / oldValue) * 100
}

/**
 * Determine trend direction based on recent data
 */
export function determineTrend(data: CandlestickData[], period: number = 10): 'up' | 'down' | 'sideways' {
  if (data.length < period) return 'sideways'
  
  const recentData = data.slice(-period)
  if (recentData.length === 0 || !recentData[0] || !recentData[recentData.length - 1]) {
    return 'sideways'
  }
  
  const firstPrice = recentData[0].close
  const lastPrice = recentData[recentData.length - 1].close
  const change = calculatePercentageChange(firstPrice, lastPrice)
  
  if (change > 2) return 'up'
  if (change < -2) return 'down'
  return 'sideways'
}

/**
 * Find price levels where stock has bounced multiple times (support/resistance)
 */
export function findKeyLevels(data: CandlestickData[], tolerance: number = 0.02): {
  supportLevels: number[]
  resistanceLevels: number[]
} {
  const supportLevels: number[] = []
  const resistanceLevels: number[] = []
  
  if (data.length < 20) return { supportLevels, resistanceLevels }
  
  // Find local minima and maxima
  const localMinima: number[] = []
  const localMaxima: number[] = []
  
  for (let i = 2; i < data.length - 2; i++) {
    const current = data[i]
    const prev2 = data[i - 2]
    const prev1 = data[i - 1]
    const next1 = data[i + 1]
    const next2 = data[i + 2]
    
    if (!current || !prev2 || !prev1 || !next1 || !next2) continue
    
    // Local minimum (support)
    if (current.low <= prev2.low && 
        current.low <= prev1.low && 
        current.low <= next1.low && 
        current.low <= next2.low) {
      localMinima.push(current.low)
    }
    
    // Local maximum (resistance)
    if (current.high >= prev2.high && 
        current.high >= prev1.high && 
        current.high >= next1.high && 
        current.high >= next2.high) {
      localMaxima.push(current.high)
    }
  }
  
  // Group similar levels together
  const groupLevels = (levels: number[]): number[] => {
    const grouped: number[] = []
    const sorted = [...levels].sort((a, b) => a - b)
    
    for (const level of sorted) {
      const isNearExisting = grouped.some(existing => 
        Math.abs(level - existing) / existing < tolerance
      )
      
      if (!isNearExisting) {
        grouped.push(level)
      }
    }
    
    return grouped
  }
  
  return {
    supportLevels: groupLevels(localMinima).slice(0, 5), // Top 5 support levels
    resistanceLevels: groupLevels(localMaxima).slice(0, 5) // Top 5 resistance levels
  }
}