/**
 * Anchored VWAP Overlay - Cumulative volume-weighted average price
 *
 * Formula: VWAP = Σ(Typical Price × Volume) / Σ(Volume)
 * Typical Price = (High + Low + Close) / 3
 *
 * Anchored to chart start (first visible bar). Does NOT reset daily.
 * Used for identifying major support/resistance zones in swing trading.
 *
 * Interpretation:
 * - Price above VWAP = Long-term bullish (buyers in control)
 * - Price below VWAP = Long-term bearish (sellers in control)
 * - VWAP break with volume = Major trend shift
 *
 * Note: Requires daily OHLCV data (not intraday). For day trading VWAP,
 * use intraday data with session resets.
 */

'use client'

import { useMemo, useEffect } from 'react'
import { LineSeries } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateVWAP } from '@/lib/indicators/calculations'
import { useVWAPSettings } from '@/lib/hooks/use-vwap-settings'

export function VWAPOverlay() {
  const { chartData, chart } = useChart()
  const { settings } = useVWAPSettings()

  // Calculate VWAP from OHLCV data with configurable anchor point
  const vwapData = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)
    const volumes = chartData.volumeData.map(d => d.value)

    if (highs.length === 0 || volumes.length === 0) {
      console.log('[VWAP] Insufficient data')
      return []
    }

    // Calculate start index based on settings.anchorDays
    // If anchorDays is null, startIndex = 0 (chart start)
    // If anchorDays is a number, calculate how many bars back from the end
    const dataLength = highs.length
    const startIndex = settings.anchorDays === null
      ? 0
      : Math.max(0, dataLength - settings.anchorDays)

    console.log(`[VWAP] Anchor point: ${settings.anchorDays === null ? 'Chart Start' : `${settings.anchorDays} days`}, Start index: ${startIndex}/${dataLength}`)

    const vwapValues = calculateVWAP(highs, lows, closes, volumes, startIndex)

    const data: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (vwapValues[i] !== null) {
        data.push({
          time: d.time as string,
          value: vwapValues[i]!
        })
      }
    })

    console.log('[VWAP] Calculated:', data.length, 'points')
    return data
  }, [chartData, settings.anchorDays])

  // Render Anchored VWAP line on main chart (pane 0)
  // Capture series ref for manual data updates
  const seriesRef = useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#2196F3',  // Blue - institutional color
      lineWidth: 2,
      title: 'Anchored VWAP',
      priceScaleId: 'right',
    },
    data: vwapData,
    paneIndex: 0,  // Main chart overlay
  })

  // Explicitly update series data when vwapData changes AND force chart redraw
  // This ensures visual updates even if useSeries's data effect doesn't fire
  // requestAnimationFrame ensures TradingView's rendering pipeline has completed
  useEffect(() => {
    if (seriesRef.current && vwapData.length > 0 && chart) {
      console.log(`[VWAP] Explicit data update: ${vwapData.length} points`)
      seriesRef.current.setData(vwapData)
    }
  }, [vwapData, chart])

  return null  // Component manages chart series, doesn't render DOM
}
