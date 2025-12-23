/**
 * FibonacciRetracement - Fibonacci retracement levels overlay
 *
 * Displays horizontal price lines at key Fibonacci ratios (23.6%, 38.2%, 50%, 61.8%, 78.6%)
 * Used to identify potential support/resistance levels based on recent price swings
 *
 * Visual:
 * - 0% (High) - Green line
 * - 23.6% - Light blue line
 * - 38.2% - Blue line (key level)
 * - 50.0% - Yellow line (midpoint)
 * - 61.8% - Orange line (golden ratio - most important)
 * - 78.6% - Red line (deep retracement)
 * - 100% (Low) - Dark red line
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateFibonacciRetracement } from '@/lib/indicators/calculations'
import { useSupportResistanceSettings } from '@/lib/hooks/use-indicator-settings'

export function FibonacciRetracement() {
  const { chartData } = useChart()
  const { settings } = useSupportResistanceSettings()

  // Calculate Fibonacci levels from OHLC data
  const fibLevels = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)

    if (highs.length === 0) {
      console.log('[Fibonacci] Insufficient data')
      return null
    }

    const levels = calculateFibonacciRetracement(highs, lows, settings.fibonacciLookback)

    if (!levels) {
      console.log('[Fibonacci] No levels calculated')
      return null
    }

    console.log('[Fibonacci] Calculated:', levels.levels.length, 'levels')
    return levels
  }, [chartData, settings.fibonacciLookback])

  // Prepare horizontal line data for each Fibonacci level
  const levelDataSets = useMemo(() => {
    if (!fibLevels) return []

    return fibLevels.levels.map(level => ({
      ratio: level.ratio,
      label: level.label,
      data: chartData.candlestickData.map(d => ({
        time: d.time as string,
        value: level.price
      }))
    }))
  }, [fibLevels, chartData])

  // Color configuration for each Fibonacci level
  const levelColors = [
    { ratio: 0, color: '#10B981', lineWidth: 2 },      // 0% (High) - Green
    { ratio: 0.236, color: '#60A5FA', lineWidth: 1 },  // 23.6% - Light blue
    { ratio: 0.382, color: '#3B82F6', lineWidth: 2 },  // 38.2% - Blue (key)
    { ratio: 0.500, color: '#FBBF24', lineWidth: 2 },  // 50.0% - Yellow (midpoint)
    { ratio: 0.618, color: '#F97316', lineWidth: 2 },  // 61.8% - Orange (golden ratio)
    { ratio: 0.786, color: '#EF4444', lineWidth: 1 },  // 78.6% - Red
    { ratio: 1.000, color: '#DC2626', lineWidth: 2 },  // 100% (Low) - Dark red
  ]

  // Render each Fibonacci level as a separate LineSeries (horizontal line)
  levelDataSets.forEach((levelData) => {
    const colorConfig = levelColors.find(c => c.ratio === levelData.ratio)
    if (!colorConfig) return

    // eslint-disable-next-line react-hooks/rules-of-hooks
    useSeries({
      seriesType: LineSeries,
      seriesOptions: {
        color: colorConfig.color,
        lineWidth: colorConfig.lineWidth,
        lineStyle: LineStyle.Dashed,
        title: `Fib ${levelData.label}`,
        priceScaleId: 'right',
      },
      data: levelData.data,
      paneIndex: 0,  // Main chart overlay
    })
  })

  return null  // Component manages chart series, doesn't render DOM
}
