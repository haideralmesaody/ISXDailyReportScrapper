/**
 * use-pane-transition-manager Hook
 * Preserves zoom level and scroll position when panes are added/removed
 *
 * Purpose:
 * - Detects when pane count changes (indicator toggled ON/OFF)
 * - Saves visible time range before panes shift
 * - Restores zoom level after series recreation
 *
 * Why This Matters:
 * - When user toggles indicator OFF, remaining indicators shift to new panes
 * - useSeries recreates series at new pane indices (due to paneIndex dependency)
 * - Without preservation, chart would reset to fitContent()
 * - This hook maintains user's zoom/scroll state across transitions
 *
 * Usage:
 *   usePaneTransitionManager(chartRef.current, indicators)
 */

'use client'

import { useEffect, useRef } from 'react'
import type { IChartApi, Range } from 'lightweight-charts'
import type { IndicatorSettings } from './use-chart-state'

/**
 * List of pane-based indicators
 * (Indicators that create separate panes below the main chart)
 */
const PANE_INDICATORS = [
  'showVolume',
  'showMACD',
  'showRSI',
  'showStochastic',
  'showGeneralStochastic',
  'showADX',
  'showATR',
  'showCCI',
  'showWilliamsR',
  'showMOM',
  'showROC',
] as const

/**
 * Hook to preserve zoom/scroll during pane transitions
 *
 * @param chart - TradingView chart instance
 * @param indicators - Current indicator settings
 */
export function usePaneTransitionManager(
  chart: IChartApi | null,
  indicators: IndicatorSettings
) {
  const prevPaneCountRef = useRef<number>(0)
  const savedRangeRef = useRef<Range<number> | null>(null)

  useEffect(() => {
    if (!chart) return

    // Count enabled pane indicators
    const enabledPaneIndicators = PANE_INDICATORS.filter(
      (key) => indicators[key] === true
    )
    const currentPaneCount = enabledPaneIndicators.length

    // Detect pane count change (indicator toggled)
    if (prevPaneCountRef.current !== currentPaneCount && prevPaneCountRef.current > 0) {
      console.log(`[PaneTransition] 🔄 Pane count changed: ${prevPaneCountRef.current} → ${currentPaneCount}`)

      // Save current visible range BEFORE panes shift
      try {
        const visibleRange = chart.timeScale().getVisibleRange()
        if (visibleRange) {
          savedRangeRef.current = visibleRange
          // Range saved silently (log removed to prevent Invalid Date errors)
        }
      } catch (error) {
        console.error('[PaneTransition] Failed to save range:', error)
      }

      // Restore range after panes recreate (2 animation frames ensures series are ready)
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          if (savedRangeRef.current) {
            try {
              chart.timeScale().setVisibleRange(savedRangeRef.current)
              console.log(`[PaneTransition] ✅ Restored visible range`)
              savedRangeRef.current = null
            } catch (error) {
              console.error('[PaneTransition] Failed to restore range:', error)
              // Fallback to fitContent if restoration fails
              chart.timeScale().fitContent()
            }
          }
        })
      })
    }

    prevPaneCountRef.current = currentPaneCount
  }, [chart, indicators])
}
