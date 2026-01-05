/**
 * use-series Hook - Reusable series lifecycle management
 * Following TradingView official React patterns
 *
 * Purpose:
 * - Centralized series creation/cleanup logic
 * - Automatic React-based lifecycle management
 * - Prevents memory leaks with proper cleanup
 * - Eliminates manual remove/re-add patterns
 *
 * Usage:
 *   const seriesRef = useSeries({
 *     seriesType: LineSeries,
 *     seriesOptions: { color: '#3B82F6', title: 'SMA 20' },
 *     data: smaData,
 *     paneIndex: 0
 *   })
 */

'use client'

import { useEffect, useRef } from 'react'
import { useChart } from '@/components/analysis/chart/ChartContext'
import type { ISeriesApi } from 'lightweight-charts'
import { logger } from '@/lib/utils/logger'

export interface PriceScaleOptions {
  visible?: boolean
  autoScale?: boolean
  borderVisible?: boolean
  scaleMargins?: { top: number; bottom: number }
  mode?: number
  entireTextOnly?: boolean
}

export interface UseSeriesOptions<T> {
  /** TradingView series type (LineSeries, HistogramSeries, CandlestickSeries) */
  seriesType: any

  /** Series configuration options */
  seriesOptions: Record<string, any>

  /** Data to display in the series */
  data: T[]

  /** Optional pane index (0 = main, 1+ = indicator panes) */
  paneIndex?: number

  /** Optional price scale configuration (applied immediately after series creation) */
  priceScaleOptions?: PriceScaleOptions

  /** Optional callback when series is created */
  onCreated?: (series: ISeriesApi<any>) => void

  /** Optional callback when series is destroyed */
  onDestroyed?: () => void
}

/**
 * Hook to create and manage a TradingView series
 *
 * Lifecycle:
 * 1. Mount: Creates series when chart is ready
 * 2. Update: Updates data when it changes
 * 3. Unmount: Removes series (React cleanup)
 *
 * This ensures Y-axis stability by:
 * - Only creating/destroying when component mounts/unmounts
 * - NOT recreating series when data updates
 * - Letting TradingView handle autoscaling naturally
 */
export function useSeries<T>({
  seriesType,
  seriesOptions,
  data,
  paneIndex,
  priceScaleOptions,
  onCreated,
  onDestroyed,
}: UseSeriesOptions<T>) {
  const { chart, isReady } = useChart()
  const seriesRef = useRef<ISeriesApi<any> | null>(null)

  // PHASE 1: CREATE SERIES AND CONFIGURE PRICE SCALE (TradingView v5.0 compliant)
  useEffect(() => {
    if (!chart || !isReady) return

    try {
      // Step 1: Create series (this creates pane and price scale automatically)
      const series = chart.addSeries(seriesType, seriesOptions, paneIndex)
      seriesRef.current = series

      logger.log(`[useSeries] ✅ Created series: ${seriesOptions.title || seriesOptions.priceScaleId || 'unnamed'}`)

      // Step 2: IMMEDIATELY configure price scale (TradingView best practice)
      // Custom price scales default to visible:false, must be configured synchronously
      if (priceScaleOptions && seriesOptions.priceScaleId && paneIndex !== undefined && paneIndex > 0) {
        // Use requestAnimationFrame to ensure TradingView's internal state is ready
        requestAnimationFrame(() => {
          try {
            // Access price scale by ID AND paneIndex (TradingView v5 API requirement)
            const priceScale = chart.priceScale(seriesOptions.priceScaleId, paneIndex)

            if (priceScale) {
              // Apply configuration with sensible defaults + custom overrides
              priceScale.applyOptions({
                visible: true,       // ✅ Show Y-axis labels (custom scales default to false)
                autoScale: true,     // ✅ Enable auto-scaling
                borderVisible: true, // ✅ Show border
                ...priceScaleOptions, // Allow component-specific overrides
              })

              logger.log(`[useSeries] ✅ Configured price scale '${seriesOptions.priceScaleId}' for pane ${paneIndex}`)
            } else {
              logger.warn(`[useSeries] ⚠️  Price scale '${seriesOptions.priceScaleId}' not found for pane ${paneIndex}`)
            }
          } catch (error) {
            logger.error('[useSeries] Failed to configure price scale:', error)
          }
        })
      }

      // Step 3: Call onCreated callback IMMEDIATELY (not in RAF)
      // ✅ FIX: requestAnimationFrame was causing infinite loops
      // - RAF delays the callback until after next paint
      // - Component may re-render before callback executes
      // - Multiple callbacks queue up → infinite setCandlestickSeries calls
      // - For candlestick series (pane 0), immediate call is fine
      // - Price scale config (above) still uses RAF for pane > 0
      if (onCreated) {
        onCreated(series)
      }

    } catch (error) {
      logger.error('[useSeries] ❌ Failed to create series:', error)
    }

    // CLEANUP: Remove series when component unmounts OR paneIndex changes
    return () => {
      if (!seriesRef.current || !chart) {
        logger.log(`[useSeries] Skipping cleanup - refs already null`)
        return
      }

      try {
        // Defensive check: ensure chart is still valid before accessing
        const panes = chart.panes()
        if (!panes || panes.length === 0) {
          logger.warn(`[useSeries] Chart disposed, skipping series removal`)
          seriesRef.current = null
          return
        }

        chart.removeSeries(seriesRef.current)
        logger.log(`[useSeries] 🗑️ Removed series from pane ${paneIndex}: ${seriesOptions.title || seriesOptions.priceScaleId || 'unnamed'}`)

        seriesRef.current = null

        if (onDestroyed) {
          onDestroyed()
        }
      } catch (error) {
        // Gracefully handle "Object is disposed" errors
        logger.warn('[useSeries] Cleanup failed (chart may be disposed):', error)
        seriesRef.current = null
      }
    }
  }, [chart, isReady, paneIndex])  // ✅ Recreate when chart, readiness, OR paneIndex changes

  // PHASE 2: UPDATE DATA (when data changes, but series already exists)
  useEffect(() => {
    if (!seriesRef.current) {
      return  // Series not created yet
    }

    if (data.length === 0) {
      // Removed console flooding: '⚠️ Empty data for: ...'
      return
    }

    try {
      seriesRef.current.setData(data)
      // Removed console flooding: '📊 Data updated: ... (N points)'
    } catch (error) {
      logger.error('[useSeries] Failed to update data:', error)
    }
  }, [data])  // Only when data reference changes

  // PHASE 3 REMOVED: Options update effect caused infinite loops
  // - Options are already set during series creation (line 89)
  // - Inline seriesOptions objects created new references every render
  // - This triggered the effect infinitely
  // - If dynamic option updates are needed, components should handle it explicitly
  // - Theme changes are handled by recreating the entire chart

  return seriesRef
}
