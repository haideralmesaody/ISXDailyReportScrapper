/**
 * Indicators API Client
 *
 * Client for fetching unified indicator calculations from the backend
 * Replaces frontend JavaScript calculations with backend API calls
 */

import { API_BASE_URL } from '@/lib/constants/api'

// API types matching backend Go structs
export interface IndicatorRequest {
  symbol: string
  timeframe: string    // 1D, 1W, 1M, 3M, 1Y, MAX
  period: number      // RSI period, typically 14
  start_date?: string
  end_date?: string
}

export interface IndicatorValue {
  date: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  rsi?: number
  sma20?: number
  sma50?: number
  sma200?: number
  ema20?: number
  macd?: number
  macdSignal?: number
  macdHistogram?: number
  bollingerUpper?: number
  bollingerMiddle?: number
  bollingerLower?: number
  stochasticK?: number
  stochasticD?: number
  adx?: number
  plusDI?: number
  minusDI?: number
  atr?: number
  cci?: number
  williamsR?: number
}

export interface IndicatorResult {
  symbol: string
  timeframe: string
  period: number
  values: IndicatorValue[]
  calculated_at: string
  data_points: number
}

// Ticker indicators from SSOT CSV file - Matching Analysis Page (29 indicators)
export interface TickerIndicators {
  symbol: string
  date: string
  price: number
  volume: number

  // Momentum Indicators (8)
  rsi_14: number
  stoch_k: number
  stoch_d: number
  cci_20: number
  williams_r: number
  mfi_14: number
  momentum: number
  roc: number

  // Moving Averages (4)
  sma_20: number
  sma_50: number
  sma_200: number
  ema_20: number

  // MACD System (3)
  macd: number
  macd_signal: number
  macd_hist: number

  // Volume Indicators (2)
  obv: number
  vwap: number

  // Volatility Indicators (7)
  atr_14: number
  bollinger_upper: number
  bollinger_middle: number
  bollinger_lower: number
  keltner_upper: number
  keltner_middle: number
  keltner_lower: number

  // Trend Indicators (2)
  adx_14: number
  parabolic_sar: number

  // Support & Resistance (2)
  support_level: number
  resistance_level: number
}

export interface RSISummary {
  symbol: string
  current_value: number
  period: number
  trend: string        // OVERBOUGHT, OVERSOLD, NEUTRAL
  signal: string       // BUY, SELL, HOLD
  last_updated: string
}

// API response types
export interface IndicatorsResponse {
  success: boolean
  data?: IndicatorResult
  error?: string
}

export interface RSISummaryResponse {
  success: boolean
  data?: RSISummary
  error?: string
}

export interface BatchIndicatorsResponse {
  success: boolean
  data?: {
    results: IndicatorResult[]
    count: number
  }
  error?: string
}

// API client functions
export const indicatorsAPI = {
  /**
   * Get RSI data for a single symbol
   */
  async getRSI(symbol: string, options: {
    period?: number
    timeframe?: string
    start_date?: string
    end_date?: string
  } = {}): Promise<IndicatorResult> {
    const params = new URLSearchParams()

    if (options.period) params.append('period', options.period.toString())
    if (options.timeframe) params.append('timeframe', options.timeframe)
    if (options.start_date) params.append('start_date', options.start_date)
    if (options.end_date) params.append('end_date', options.end_date)

    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/rsi/${symbol}${params.toString() ? '?' + params.toString() : ''}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch RSI data: ${response.statusText}`)
    }

    const responseData: IndicatorsResponse = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch RSI data')
    }

    return responseData.data
  },

  /**
   * Get RSI summary for a symbol
   */
  async getRSISummary(symbol: string, options: {
    period?: number
    timeframe?: string
  } = {}): Promise<RSISummary> {
    const params = new URLSearchParams()

    if (options.period) params.append('period', options.period.toString())
    if (options.timeframe) params.append('timeframe', options.timeframe)

    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/rsi/${symbol}/summary${params.toString() ? '?' + params.toString() : ''}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch RSI summary: ${response.statusText}`)
    }

    const responseData: RSISummaryResponse = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch RSI summary')
    }

    return responseData.data
  },

  /**
   * Calculate RSI for multiple symbols (batch request)
   */
  async calculateRSIBatch(requests: IndicatorRequest[]): Promise<{
    results: IndicatorResult[]
    count: number
  }> {
    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/rsi/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ requests })
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to calculate batch RSI: ${response.statusText}`)
    }

    const responseData: BatchIndicatorsResponse = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to calculate batch RSI')
    }

    return responseData.data
  },

  /**
   * Get all indicators for a symbol (for future expansion)
   */
  async getIndicators(symbol: string, options: {
    timeframe?: string
    start_date?: string
    end_date?: string
  } = {}): Promise<IndicatorResult> {
    const params = new URLSearchParams()

    if (options.timeframe) params.append('timeframe', options.timeframe)
    if (options.start_date) params.append('start_date', options.start_date)
    if (options.end_date) params.append('end_date', options.end_date)

    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/${symbol}${params.toString() ? '?' + params.toString() : ''}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch indicators: ${response.statusText}`)
    }

    const responseData: IndicatorsResponse = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch indicators')
    }

    return responseData.data
  },

  /**
   * Get supported timeframes
   */
  async getSupportedTimeframes(): Promise<string[]> {
    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/timeframes`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      throw new Error(`Failed to fetch supported timeframes: ${response.statusText}`)
    }

    const responseData = await response.json()
    return responseData.timeframes
  },

  /**
   * Get default periods for different indicators
   */
  async getDefaultPeriods(): Promise<Record<string, number>> {
    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/periods`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      throw new Error(`Failed to fetch default periods: ${response.statusText}`)
    }

    const responseData = await response.json()
    return responseData.periods
  },

  /**
   * Get all indicators for a symbol from SSOT (ticker_indicators.csv)
   */
  async getTickerIndicators(symbol: string, options: {
    timeframe?: string
    start_date?: string
    end_date?: string
  } = {}): Promise<TickerIndicators[]> {
    const params = new URLSearchParams()

    if (options.timeframe) params.append('timeframe', options.timeframe)
    if (options.start_date) params.append('start_date', options.start_date)
    if (options.end_date) params.append('end_date', options.end_date)

    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/ticker/${symbol}${params.toString() ? '?' + params.toString() : ''}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch ticker indicators: ${response.statusText}`)
    }

    const responseData = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch ticker indicators')
    }

    return responseData.data
  },

  /**
   * Get latest indicator values for a symbol from SSOT
   */
  async getLatestIndicators(symbol: string): Promise<TickerIndicators | null> {
    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/ticker/${symbol}/latest`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch latest indicators: ${response.statusText}`)
    }

    const responseData = await response.json()

    if (!responseData.success || !responseData.data) {
      return null
    }

    return responseData.data
  },

  /**
   * Get indicator summary for multiple symbols from SSOT
   */
  async getBatchIndicators(symbols: string[]): Promise<Record<string, TickerIndicators>> {
    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/ticker/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ symbols })
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch batch indicators: ${response.statusText}`)
    }

    const responseData = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch batch indicators')
    }

    return responseData.data
  },

  /**
   * Check if ticker indicators data is available
   */
  async checkDataAvailability(symbol?: string): Promise<{
    available: boolean
    lastUpdated?: string
    symbolCount?: number
  }> {
    const url = symbol
      ? `${API_BASE_URL}/api/v1/indicators/availability/${symbol}`
      : `${API_BASE_URL}/api/v1/indicators/availability`

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      return { available: false }
    }

    const responseData = await response.json()
    return responseData.data || { available: false }
  },

  /**
   * Get all available indicators for a symbol from SSOT (42 indicators)
   */
  async getAllIndicators(symbol: string, options: {
    start_date?: string
    end_date?: string
  } = {}): Promise<TickerIndicators[]> {
    const params = new URLSearchParams()

    if (options.start_date) params.append('start_date', options.start_date)
    if (options.end_date) params.append('end_date', options.end_date)

    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/ticker/${symbol}/all${params.toString() ? '?' + params.toString() : ''}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch all indicators: ${response.statusText}`)
    }

    const responseData = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch all indicators')
    }

    return responseData.data
  },

  /**
   * Get indicator by category from SSOT
   */
  async getIndicatorsByCategory(symbol: string, category: 'momentum' | 'moving_averages' | 'macd' | 'volume' | 'volatility' | 'trend' | 'support_resistance' | 'ichimoku' | 'donchian', options: {
    start_date?: string
    end_date?: string
  } = {}): Promise<Partial<TickerIndicators>[]> {
    const params = new URLSearchParams()
    params.append('category', category)

    if (options.start_date) params.append('start_date', options.start_date)
    if (options.end_date) params.append('end_date', options.end_date)

    const response = await fetch(`${API_BASE_URL}/api/v1/indicators/ticker/${symbol}/category${params.toString() ? '?' + params.toString() : ''}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch indicators by category: ${response.statusText}`)
    }

    const responseData = await response.json()

    if (!responseData.success || !responseData.data) {
      throw new Error(responseData.error || 'Failed to fetch indicators by category')
    }

    return responseData.data
  }
}

// Helper functions for working with RSI data

/**
 * Extract RSI values from indicator results
 */
export function extractRSIValues(result: IndicatorResult): (number | null)[] {
  return result.values.map(value => value.rsi ?? null)
}

/**
 * Extract dates from indicator results
 */
export function extractDates(result: IndicatorResult): string[] {
  return result.values.map(value => value.date)
}

/**
 * Extract closing prices from indicator results
 */
export function extractClosingPrices(result: IndicatorResult): number[] {
  return result.values.map(value => value.close)
}

/**
 * Get the latest RSI value from indicator results
 */
export function getLatestRSIValue(result: IndicatorResult): number | null {
  if (result.values.length === 0) return null
  const latestValue = result.values[result.values.length - 1]
  return latestValue.rsi ?? null
}

/**
 * Format RSI value for display
 */
export function formatRSIValue(value: number | null, decimals: number = 2): string {
  if (value === null) return 'N/A'
  return value.toFixed(decimals)
}

/**
 * Determine RSI signal based on value
 */
export function getRSISignal(rsi: number): 'BUY' | 'SELL' | 'HOLD' {
  if (rsi >= 70) return 'SELL'     // Overbought
  if (rsi <= 30) return 'BUY'      // Oversold
  return 'HOLD'                   // Neutral
}

/**
 * Determine RSI trend based on value
 */
export function getRSITrend(rsi: number): 'OVERBOUGHT' | 'OVERSOLD' | 'NEUTRAL' {
  if (rsi >= 70) return 'OVERBOUGHT'
  if (rsi <= 30) return 'OVERSOLD'
  return 'NEUTRAL'
}

/**
 * Calculate RSI change percentage
 */
export function calculateRSIChange(current: number, previous: number): number {
  if (previous === 0) return 0
  return ((current - previous) / previous) * 100
}

/**
 * Validate indicator request parameters
 */
export function validateIndicatorRequest(request: IndicatorRequest): string[] {
  const errors: string[] = []

  if (!request.symbol || request.symbol.trim() === '') {
    errors.push('Symbol is required')
  }

  if (request.period <= 0 || request.period > 100) {
    errors.push('Period must be between 1 and 100')
  }

  const validTimeframes = ['1D', '1W', '1M', '3M', '1Y', 'MAX']
  if (!validTimeframes.includes(request.timeframe)) {
    errors.push(`Invalid timeframe. Valid options: ${validTimeframes.join(', ')}`)
  }

  // Validate date formats if provided
  if (request.start_date) {
    const startDate = new Date(request.start_date)
    if (isNaN(startDate.getTime())) {
      errors.push('Invalid start_date format. Use YYYY-MM-DD')
    }
  }

  if (request.end_date) {
    const endDate = new Date(request.end_date)
    if (isNaN(endDate.getTime())) {
      errors.push('Invalid end_date format. Use YYYY-MM-DD')
    }
  }

  return errors
}

/**
 * Create default indicator request
 */
export function createDefaultIndicatorRequest(symbol: string): IndicatorRequest {
  return {
    symbol: symbol,
    timeframe: '1D',
    period: 14,
  }
}

/**
 * Transform ticker indicators to map format for chart usage
 */
export function transformTickerIndicatorsToMap(indicators: TickerIndicators[]): Map<string, Partial<TickerIndicators>> {
  const map = new Map<string, Partial<TickerIndicators>>()

  indicators.forEach(indicator => {
    const dateStr = new Date(indicator.date).toISOString().split('T')[0]
    map.set(dateStr, indicator)
  })

  return map
}

/**
 * Get indicator value by name from ticker indicators
 */
export function getIndicatorValue(indicators: Partial<TickerIndicators>, indicatorName: keyof TickerIndicators): number | null {
  const value = indicators[indicatorName]
  return typeof value === 'number' && !isNaN(value) ? value : null
}

/**
 * Extract specific indicator values from ticker indicators
 */
export function extractIndicatorValues(
  indicators: TickerIndicators[],
  indicatorName: keyof TickerIndicators
): (number | null)[] {
  return indicators.map(indicator => getIndicatorValue(indicator, indicatorName))
}

/**
 * Get latest indicator value from ticker indicators array
 */
export function getLatestIndicatorValue(
  indicators: TickerIndicators[],
  indicatorName: keyof TickerIndicators
): number | null {
  if (indicators.length === 0) return null

  // Sort by date to get the latest
  const sortedIndicators = [...indicators].sort((a, b) =>
    new Date(b.date).getTime() - new Date(a.date).getTime()
  )

  return getIndicatorValue(sortedIndicators[0], indicatorName)
}

/**
 * Create indicator data for chart series
 */
export function createIndicatorDataForChart(
  indicators: TickerIndicators[],
  indicatorName: keyof TickerIndicators,
  dateMap?: Map<string, any>
): { time: string; value: number | null }[] {
  return indicators
    .map(indicator => {
      const dateStr = new Date(indicator.date).toISOString().split('T')[0]
      const value = getIndicatorValue(indicator, indicatorName)

      // If dateMap is provided, only include dates that exist in the chart
      if (dateMap && !dateMap.has(dateStr)) {
        return null
      }

      return {
        time: dateStr,
        value
      }
    })
    .filter((item): item is { time: string; value: number | null } =>
      item !== null && item.value !== null
    )
}

/**
 * Cache management for indicator data
 */
export class IndicatorCache {
  private cache = new Map<string, { data: any; timestamp: number; ttl: number }>()

  set(key: string, data: any, ttlMs: number = 300000): void { // 5 minutes default TTL
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl: ttlMs
    })
  }

  get(key: string): any | null {
    const entry = this.cache.get(key)
    if (!entry) return null

    if (Date.now() - entry.timestamp > entry.ttl) {
      this.cache.delete(key)
      return null
    }

    return entry.data
  }

  clear(): void {
    this.cache.clear()
  }

  // Get cache size for debugging
  size(): number {
    return this.cache.size
  }
}

// Global cache instance
export const indicatorCache = new IndicatorCache()

/**
 * Error handling helper
 */
export class IndicatorsError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public details?: any
  ) {
    super(message)
    this.name = 'IndicatorsError'
  }
}

/**
 * API error handler wrapper
 */
export async function handleIndicatorsError<T>(
  apiCall: () => Promise<T>
): Promise<T> {
  try {
    return await apiCall()
  } catch (error) {
    if (error instanceof IndicatorsError) {
      throw error
    }

    if (error instanceof Error) {
      throw new IndicatorsError(
        `Indicators API error: ${error.message}`,
        undefined,
        error
      )
    }

    throw new IndicatorsError('Unknown indicators API error')
  }
}