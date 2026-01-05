/**
 * Hook for managing SSOT (Single Source of Truth) indicator data
 * Fetches and caches pre-calculated indicators from ticker_indicators.csv
 */

import { useState, useEffect, useCallback, useRef } from 'react'
import { indicatorsAPI, type TickerIndicators, indicatorCache } from '@/lib/api/indicators'
import { logger } from '@/lib/utils/logger'

interface UseSSOTIndicatorsOptions {
  symbol: string
  timeframe?: string
  startDate?: string
  endDate?: string
  enabled?: boolean
  cacheTTL?: number
}

interface UseSSOTIndicatorsReturn {
  indicators: TickerIndicators[]
  indicatorsMap: Map<string, Partial<TickerIndicators>>
  loading: boolean
  error: string | null
  refetch: () => Promise<void>
  isDataAvailable: boolean
  latestIndicators: TickerIndicators | null
}

export function useSSOTIndicators({
  symbol,
  timeframe = '1D',
  startDate,
  endDate,
  enabled = true,
  cacheTTL = 300000 // 5 minutes
}: UseSSOTIndicatorsOptions): UseSSOTIndicatorsReturn {
  const [indicators, setIndicators] = useState<TickerIndicators[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isDataAvailable, setIsDataAvailable] = useState(false)
  const [latestIndicators, setLatestIndicators] = useState<TickerIndicators | null>(null)

  // Ref to track the current request to prevent race conditions
  const currentRequestRef = useRef<string | null>(null)

  // Create cache key
  const createCacheKey = useCallback(() => {
    return `ssot-indicators-${symbol}-${timeframe}-${startDate || ''}-${endDate || ''}`
  }, [symbol, timeframe, startDate, endDate])

  // Transform indicators to map for efficient lookup
  const indicatorsMap = new Map<string, Partial<TickerIndicators>>(
    indicators.map(indicator => [
      new Date(indicator.date).toISOString().split('T')[0],
      indicator
    ])
  )

  // Fetch indicators data
  const fetchIndicators = useCallback(async (forceRefresh = false) => {
    if (!enabled || !symbol) {
      return
    }

    const requestKey = createCacheKey()

    // Prevent duplicate requests
    if (currentRequestRef.current === requestKey && !forceRefresh) {
      logger.log('[SSOTIndicators] Request already in progress, skipping')
      return
    }

    currentRequestRef.current = requestKey
    setLoading(true)
    setError(null)

    try {
      // Check cache first (unless force refresh)
      if (!forceRefresh) {
        const cachedData = indicatorCache.get(requestKey)
        if (cachedData) {
          logger.log('[SSOTIndicators] Using cached data for', symbol)
          setIndicators(cachedData.indicators)
          setIsDataAvailable(true)
          return
        }
      }

      // Check data availability first
      const availability = await indicatorsAPI.checkDataAvailability(symbol)
      if (!availability.available) {
        setIsDataAvailable(false)
        setError('Indicator data not available. Run the data pipeline to generate indicators.')
        return
      }

      setIsDataAvailable(true)

      // Fetch the indicators
      const data = await indicatorsAPI.getTickerIndicators(symbol, {
        timeframe,
        start_date: startDate,
        end_date: endDate
      })

      // Cache the results
      indicatorCache.set(requestKey, { indicators: data }, cacheTTL)

      setIndicators(data)

      // Get latest indicators
      const latest = await indicatorsAPI.getLatestIndicators(symbol)
      setLatestIndicators(latest)

      logger.log('[SSOTIndicators] Successfully loaded indicators for', symbol, {
        count: data.length,
        firstDate: data[0]?.date,
        lastDate: data[data.length - 1]?.date
      })

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load indicators'
      setError(errorMessage)
      setIsDataAvailable(false)

      logger.error('[SSOTIndicators] Failed to load indicators', {
        symbol,
        error: errorMessage
      })
    } finally {
      setLoading(false)
      currentRequestRef.current = null
    }
  }, [symbol, timeframe, startDate, endDate, enabled, cacheTTL, createCacheKey])

  // Refetch function
  const refetch = useCallback(async () => {
    await fetchIndicators(true) // Force refresh
  }, [fetchIndicators])

  // Initial fetch and refetch on dependencies change
  useEffect(() => {
    if (enabled && symbol) {
      fetchIndicators()
    }
  }, [enabled, symbol, timeframe, startDate, endDate, fetchIndicators])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      currentRequestRef.current = null
    }
  }, [])

  return {
    indicators,
    indicatorsMap,
    loading,
    error,
    refetch,
    isDataAvailable,
    latestIndicators
  }
}

/**
 * Hook for batch fetching indicators for multiple symbols
 */
export function useBatchSSOTIndicators(symbols: string[], enabled = true) {
  const [indicatorsMap, setIndicatorsMap] = useState<Record<string, TickerIndicators>>({})
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchBatchIndicators = useCallback(async () => {
    if (!enabled || symbols.length === 0) {
      return
    }

    setLoading(true)
    setError(null)

    try {
      const data = await indicatorsAPI.getBatchIndicators(symbols)
      setIndicatorsMap(data)

      logger.log('[BatchSSOTIndicators] Successfully loaded indicators', {
        symbolCount: symbols.length,
        loadedCount: Object.keys(data).length
      })

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load batch indicators'
      setError(errorMessage)

      logger.error('[BatchSSOTIndicators] Failed to load batch indicators', {
        symbols,
        error: errorMessage
      })
    } finally {
      setLoading(false)
    }
  }, [symbols, enabled])

  useEffect(() => {
    if (enabled && symbols.length > 0) {
      fetchBatchIndicators()
    }
  }, [enabled, symbols.join(','), fetchBatchIndicators])

  return {
    indicatorsMap,
    loading,
    error,
    refetch: fetchBatchIndicators
  }
}