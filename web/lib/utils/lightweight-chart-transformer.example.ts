/**
 * Example usage of the Lightweight Chart Transformer
 * This file demonstrates how to use the transformer utilities with ISX data
 */

import type { TickerHistoricalData } from '@/types/analysis'
import { 
  transformToLightweightData,
  calculateSMA,
  calculateEMA,
  calculateRSI,
  calculateBollingerBands,
  calculateMACD,
  calculateResistanceLevels,
  formatPrice,
  formatVolume,
  determineTrend
} from './lightweight-chart-transformer'

/**
 * Example: Basic chart data transformation
 */
export function createBasicChart(tickerData: TickerHistoricalData[]) {
  // Transform raw ISX data to Lightweight Charts format
  const { candlestick, volume } = transformToLightweightData(tickerData)
  
  // Calculate basic technical indicators
  const sma20 = calculateSMA(candlestick, 20)
  const ema20 = calculateEMA(candlestick, 20)
  
  return {
    candlestick,
    volume,
    indicators: {
      sma20,
      ema20
    }
  }
}

/**
 * Example: Advanced technical analysis
 */
export function createAdvancedChart(tickerData: TickerHistoricalData[]) {
  const { candlestick, volume } = transformToLightweightData(tickerData)
  
  // Calculate all technical indicators
  const sma20 = calculateSMA(candlestick, 20)
  const sma50 = calculateSMA(candlestick, 50)
  const ema12 = calculateEMA(candlestick, 12)
  const ema26 = calculateEMA(candlestick, 26)
  const rsi = calculateRSI(candlestick, 14)
  const bollinger = calculateBollingerBands(candlestick, 20, 2)
  const macd = calculateMACD(candlestick)
  
  // Calculate support and resistance
  const levels = calculateResistanceLevels(candlestick)
  const trend = determineTrend(candlestick, 20)
  
  return {
    candlestick,
    volume,
    indicators: {
      movingAverages: {
        sma20,
        sma50,
        ema12,
        ema26
      },
      oscillators: {
        rsi,
        macd: macd.macdLine,
        signal: macd.signalLine,
        histogram: macd.histogram
      },
      bands: {
        upper: bollinger.upper,
        middle: bollinger.middle,
        lower: bollinger.lower
      }
    },
    analysis: {
      resistance: levels.resistance,
      support: levels.support,
      trend
    }
  }
}

/**
 * Example: Price formatting for display
 */
export function formatChartTooltip(price: number, volume: number) {
  return {
    price: formatPrice(price),
    volume: formatVolume(volume)
  }
}

/**
 * Example: React component usage pattern
 */
export interface ChartData {
  candlestick: any[]
  volume: any[]
  indicators: {
    sma20?: any[]
    ema20?: any[]
    rsi?: any[]
  }
}

export function useChartData(tickerData: TickerHistoricalData[]): ChartData {
  if (!tickerData || tickerData.length === 0) {
    return {
      candlestick: [],
      volume: [],
      indicators: {}
    }
  }

  return createBasicChart(tickerData)
}