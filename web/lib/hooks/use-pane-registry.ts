/**
 * usePaneRegistry Hook - ACTIVATION-BASED Pane Allocation for TradingView Charts
 *
 * ✨ NEW APPROACH: Sequential allocation based on when indicators were enabled
 *
 * Purpose:
 * - Allocate panes sequentially (1, 2, 3...) based on activation order
 * - NEVER use canonical order - follow user's enable sequence
 * - Reallocate when indicators are disabled (shift others up)
 * - Force React component recreation when pane indices change
 *
 * Why This Works:
 * - User enables RSI → gets pane 1
 * - User enables MACD → gets pane 2
 * - Panes match React component render order
 * - No overlaps, no gaps, no race conditions!
 *
 * Usage:
 *   const { getPaneIndex } = usePaneRegistry(indicators, activationOrder)
 *
 *   <VolumePane paneIndex={getPaneIndex('volume')} />
 *   <MACDPane paneIndex={getPaneIndex('macd')} />
 *   <RSIPane paneIndex={getPaneIndex('rsi')} />
 */

'use client'

import { useMemo } from 'react'
import type { IndicatorSettings } from './use-chart-state'
import { logger } from '@/lib/utils/logger'

/**
 * Indicator name types for pane-based indicators
 * (Indicators that render in separate panes below the main chart)
 */
type IndicatorName =
  | 'volume'      // Volume histogram
  | 'obv'         // On-Balance Volume
  | 'macd'        // MACD (3 series: MACD line, signal, histogram)
  | 'rsi'         // RSI oscillator
  | 'mfi'         // Money Flow Index (volume-weighted RSI)
  | 'stochastic'  // Stochastic Oscillator (%K and %D)
  | 'generalStochastic'  // General/Custom Stochastic Oscillator
  | 'adx'         // Average Directional Index (ADX, +DI, -DI)
  | 'atr'         // Average True Range
  | 'cci'         // Commodity Channel Index
  | 'williamsR'   // Williams %R
  | 'momentum'    // Momentum Oscillator (MOM)
  | 'roc'         // Rate of Change (ROC)

/**
 * Map indicator setting key to indicator name
 * Note: Overlays (Bollinger, Ichimoku) don't need panes - they render on main chart (pane 0)
 */
const SETTING_TO_NAME_MAP: Record<string, IndicatorName> = {
  showVolume: 'volume',
  showOBV: 'obv',
  showMACD: 'macd',
  showRSI: 'rsi',
  showMFI: 'mfi',
  showStochastic: 'stochastic',
  showGeneralStochastic: 'generalStochastic',
  showADX: 'adx',
  showATR: 'atr',
  showCCI: 'cci',
  showWilliamsR: 'williamsR',
  showMOM: 'momentum',
  showROC: 'roc',
}

/**
 * Pane Registry Return Type
 */
export interface PaneRegistry {
  /**
   * Get the dynamic pane index for an indicator
   * @param indicatorName - Name of the indicator (e.g., 'volume', 'macd', 'rsi')
   * @returns Pane index (1+) or -1 if indicator is disabled
   */
  getPaneIndex: (indicatorName: IndicatorName) => number

  /**
   * Get total number of panes (including main candlestick pane)
   * @returns Total pane count (minimum 1)
   */
  getPaneCount: () => number

  /**
   * Get list of active indicator names in activation order
   * @returns Array of active indicator names sorted by activation timestamp
   */
  getActiveIndicators: () => IndicatorName[]

  /**
   * Debug: Log current pane allocation
   */
  logAllocation: () => void
}

/**
 * usePaneRegistry Hook
 *
 * Calculates pane indices based on ACTIVATION ORDER (not canonical order!)
 *
 * @param indicators - Current indicator settings
 * @param activationOrder - Map of indicator keys to activation timestamps
 * @returns PaneRegistry object with allocation functions
 *
 * @example
 * // User enables Volume → RSI → MACD (in that order)
 * activationOrder = {
 *   showVolume: 1000,
 *   showRSI: 2000,
 *   showMACD: 3000
 * }
 *
 * Result:
 * getPaneIndex('volume') // → 1 (first enabled)
 * getPaneIndex('rsi')    // → 2 (second enabled)
 * getPaneIndex('macd')   // → 3 (third enabled)
 *
 * @example
 * // User disables RSI
 * activationOrder = {
 *   showVolume: 1000,
 *   showMACD: 3000
 * }
 *
 * Result (reallocation!):
 * getPaneIndex('volume') // → 1 (unchanged)
 * getPaneIndex('macd')   // → 2 (shifts from pane 3 → pane 2)
 * getPaneIndex('rsi')    // → -1 (disabled)
 */
export function usePaneRegistry(
  indicators: IndicatorSettings,
  activationOrder: Record<string, number>
): PaneRegistry {
  // Memoize allocation calculation - recalculates when indicators or activation order changes
  const allocation = useMemo(() => {
    const paneMap = new Map<IndicatorName, number>()

    // 1. Build list of enabled indicators with their activation timestamps
    const enabledWithTimestamps: Array<{ name: IndicatorName; timestamp: number }> = []

    for (const [settingKey, timestamp] of Object.entries(activationOrder)) {
      const indicatorName = SETTING_TO_NAME_MAP[settingKey]

      if (indicatorName && indicators[settingKey as keyof IndicatorSettings]) {
        enabledWithTimestamps.push({ name: indicatorName, timestamp })
      }
    }

    // Fallback: if an indicator is enabled but missing from activationOrder (e.g., defaults),
    // assign a stable timestamp so it still receives a valid pane index.
    for (const [settingKey, indicatorName] of Object.entries(SETTING_TO_NAME_MAP)) {
      if (!indicators[settingKey as keyof IndicatorSettings]) continue
      if (activationOrder[settingKey] !== undefined) continue
      enabledWithTimestamps.push({ name: indicatorName, timestamp: 0 })
    }

    // 2. Sort by activation timestamp (earliest first)
    enabledWithTimestamps.sort((a, b) => a.timestamp - b.timestamp)

    // 3. Allocate panes sequentially based on activation order
    let nextPaneIndex = 1  // Start after main candlestick pane (pane 0)

    enabledWithTimestamps.forEach(({ name }) => {
      paneMap.set(name, nextPaneIndex)
      nextPaneIndex++
    })

    // 4. Debug logging
    if (paneMap.size > 0) {
      const allocations = Array.from(paneMap.entries())
        .map(([name, idx]) => `${name}=pane${idx}`)
        .join(', ')

      const timestamps = enabledWithTimestamps
        .map(({ name, timestamp }) => `${name}@${timestamp}`)
        .join(' → ')

      logger.log(`[PaneRegistry] 🚀 Activation-based allocation: ${allocations}`)
      logger.log(`[PaneRegistry] ⏱️  Activation sequence: ${timestamps}`)
    } else {
      logger.log('[PaneRegistry] 📊 No pane-based indicators active')
    }

    return {
      map: paneMap,
      totalPanes: nextPaneIndex,
      activationSequence: enabledWithTimestamps.map(({ name }) => name),
    }
  }, [indicators, activationOrder])

  // Return stable API
  return useMemo<PaneRegistry>(
    () => ({
      getPaneIndex: (indicatorName: IndicatorName) => {
        const index = allocation.map.get(indicatorName)
        return index !== undefined ? index : -1
      },

      getPaneCount: () => {
        return allocation.totalPanes
      },

      getActiveIndicators: () => {
        return allocation.activationSequence
      },

      logAllocation: () => {
        logger.group('[PaneRegistry] 📊 Current Allocation (Activation-Based)')
        logger.log('Main Chart:', 'pane 0 (always)')

        allocation.map.forEach((paneIndex, indicatorName) => {
          logger.log(`${indicatorName}:`, `pane ${paneIndex}`)
        })

        logger.log('Total Panes:', allocation.totalPanes)
        logger.log('Activation Sequence:', allocation.activationSequence.join(' → '))
        logger.groupEnd()
      },
    }),
    [allocation]
  )
}
