/**
 * useIndicatorsSSOT Hook
 *
 * Hook for fetching and managing indicators data from the Single Source of Truth (SSOT)
 * Replaces client-side calculations with pre-calculated indicators from the pipeline
 */

import { useState, useEffect, useCallback, useRef } from 'react'
import { indicatorsAPI, TickerIndicators, indicatorCache, IndicatorsError } from '@/lib/api/indicators'

interface UseIndicatorsSSOTOptions {
  symbol: string
  startDate?: string
  endDate?: string
  enableCache?: boolean
  cacheTTL?: number
  autoRefresh?: boolean
  refreshInterval?: number
}

interface UseIndicatorsSSOTResult {
  indicators: TickerIndicators[]
  loading: boolean
  error: string | null
  lastUpdated: Date | null
  refresh: () => Promise<void>
  getIndicatorValue: (indicatorName: keyof TickerIndicators) => number | null
  getLatestIndicatorValue: (indicatorName: keyof TickerIndicators) => number | null
  isDataAvailable: boolean
}

export function useIndicatorsSSOT({
  symbol,
  startDate,
  endDate,
  enableCache = true,
  cacheTTL = 15 * 60 * 1000, // 15 minutes
  autoRefresh = false,
  refreshInterval = 60 * 1000 // 1 minute
}: UseIndicatorsSSOTOptions): UseIndicatorsSSOTResult {
  const [indicators, setIndicators] = useState<TickerIndicators[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const intervalRef = useRef<NodeJS.Timeout>()
  const abortControllerRef = useRef<AbortController>()

  const generateCacheKey = useCallback(() => {
    return `indicators-${symbol}-${startDate || ''}-${endDate || ''}`
  }, [symbol, startDate, endDate])

  const fetchIndicators = useCallback(async (isBackgroundRefresh = false) => {
    // Cancel previous request if still pending
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }

    abortControllerRef.current = new AbortController()

    try {
      if (!isBackgroundRefresh) {
        setLoading(true)
      }
      setError(null)

      // Check cache first
      const cacheKey = generateCacheKey()
      if (enableCache) {
        const cachedData = indicatorCache.get(cacheKey)
        if (cachedData) {
          setIndicators(cachedData.data)
          setLastUpdated(new Date(cachedData.timestamp))
          setLoading(false)
          return
        }
      }

      // Fetch from API
      const data = await indicatorsAPI.getAllIndicators(symbol, {
        start_date: startDate,
        end_date: endDate
      })

      // Cache the results
      if (enableCache) {
        indicatorCache.set(cacheKey, {
          data,
          timestamp: Date.now(),
          metadata: {
            symbol,
            startDate,
            endDate,
            count: data.length
          }
        }, cacheTTL)
      }

      setIndicators(data)
      setLastUpdated(new Date())

    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        // Request was aborted, don't show error
        return
      }

      console.error('Failed to fetch indicators:', err)

      if (err instanceof IndicatorsError) {
        setError(err.message)
      } else {
        setError('Failed to load indicators data')
      }
    } finally {
      setLoading(false)
    }
  }, [symbol, startDate, endDate, enableCache, cacheTTL, generateCacheKey])

  const refresh = useCallback(async () => {
    await fetchIndicators(false)
  }, [fetchIndicators])

  const getIndicatorValue = useCallback((indicatorName: keyof TickerIndicators): number | null => {
    if (indicators.length === 0) return null

    const latestIndicator = indicators[indicators.length - 1]
    const value = latestIndicator[indicatorName]

    return typeof value === 'number' && !isNaN(value) ? value : null
  }, [indicators])

  const getLatestIndicatorValue = useCallback((indicatorName: keyof TickerIndicators): number | null => {
    if (indicators.length === 0) return null

    // Sort by date to get the absolute latest
    const sortedIndicators = [...indicators].sort((a, b) =>
      new Date(b.date).getTime() - new Date(a.date).getTime()
    )

    const latestIndicator = sortedIndicators[0]
    const value = latestIndicator[indicatorName]

    return typeof value === 'number' && !isNaN(value) ? value : null
  }, [indicators])

  const isDataAvailable = indicators.length > 0 && error === null

  // Initial fetch
  useEffect(() => {
    if (symbol) {
      fetchIndicators(false)
    }

    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
      }
    }
  }, [symbol, fetchIndicators])

  // Auto-refresh
  useEffect(() => {
    if (autoRefresh && refreshInterval > 0) {
      intervalRef.current = setInterval(() => {
        fetchIndicators(true) // Background refresh
      }, refreshInterval)

      return () => {
        if (intervalRef.current) {
          clearInterval(intervalRef.current)
        }
      }
    }
  }, [autoRefresh, refreshInterval, fetchIndicators])

  return {
    indicators,
    loading,
    error,
    lastUpdated,
    refresh,
    getIndicatorValue,
    getLatestIndicatorValue,
    isDataAvailable
  }
}

/**
 * Hook for fetching indicators by category
 */
export function useIndicatorsByCategory(
  symbol: string,
  category: 'momentum' | 'moving_averages' | 'macd' | 'volume' | 'volatility' | 'trend' | 'support_resistance' | 'ichimoku' | 'donchian',
  options: Omit<UseIndicatorsSSOTOptions, 'symbol'> = {}
) {
  const [indicators, setIndicators] = useState<Partial<TickerIndicators>[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchIndicatorsByCategory = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)

      const data = await indicatorsAPI.getIndicatorsByCategory(symbol, category, {
        start_date: options.startDate,
        end_date: options.endDate
      })

      setIndicators(data)

    } catch (err) {
      console.error(`Failed to fetch ${category} indicators:`, err)

      if (err instanceof IndicatorsError) {
        setError(err.message)
      } else {
        setError(`Failed to load ${category} indicators`)
      }
    } finally {
      setLoading(false)
    }
  }, [symbol, category, options.startDate, options.endDate])

  useEffect(() => {
    if (symbol) {
      fetchIndicatorsByCategory()
    }
  }, [symbol, category, fetchIndicatorsByCategory])

  return {
    indicators,
    loading,
    error,
    refresh: fetchIndicatorsByCategory
  }
}

/**
 * Hook for checking data availability
 */
export function useIndicatorsAvailability(symbol?: string) {
  const [availability, setAvailability] = useState<{
    available: boolean
    lastUpdated?: string
    symbolCount?: number
  }>({ available: false })

  const checkAvailability = useCallback(async () => {
    try {
      const result = await indicatorsAPI.checkDataAvailability(symbol)
      setAvailability(result)
    } catch (err) {
      console.error('Failed to check availability:', err)
      setAvailability({ available: false })
    }
  }, [symbol])

  useEffect(() => {
    checkAvailability()
  }, [checkAvailability])

  return {
    ...availability,
    refresh: checkAvailability
  }
}

/**
 * Utility function to get indicator by name with fallback
 */
export function getIndicatorSafely(
  indicators: Partial<TickerIndicators>,
  indicatorName: keyof TickerIndicators,
  fallback: number = 0
): number {
  const value = indicators[indicatorName]
  return typeof value === 'number' && !isNaN(value) ? value : fallback
}

/**
 * Utility function to format indicator values for display
 */
export function formatIndicatorValue(
  value: number | null,
  decimals: number = 2,
  suffix: string = ''
): string {
  if (value === null || isNaN(value)) return 'N/A'
  return `${value.toFixed(decimals)}${suffix}`
}

/**
 * Utility function to determine indicator signals
 */
export function getIndicatorSignal(
  indicatorName: keyof TickerIndicators,
  value: number | null
): 'BUY' | 'SELL' | 'HOLD' | 'NEUTRAL' {
  if (value === null || isNaN(value)) return 'NEUTRAL'

  switch (indicatorName) {
    case 'rsi_14':
      if (value >= 70) return 'SELL'
      if (value <= 30) return 'BUY'
      return 'HOLD'

    case 'stoch_k':
    case 'stoch_d':
      if (value >= 80) return 'SELL'
      if (value <= 20) return 'BUY'
      return 'HOLD'

    case 'williams_r':
      if (value <= -80) return 'BUY'
      if (value >= -20) return 'SELL'
      return 'HOLD'

    default:
      return 'NEUTRAL'
  }
}