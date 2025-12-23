/**
 * SMAIndicator - Simple Moving Average Indicator
 * Reusable component for SMA with configurable period
 */

'use client'

import { useMemo } from 'react'
import { LineSeries } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateSMA } from '@/lib/indicators/calculations'
import { logger } from '@/lib/utils/logger'

interface SMAIndicatorProps {
  period: number
}

const SMA_COLORS = {
  20: '#3B82F6',   // Blue
  50: '#F97316',   // Orange
  200: '#8B5CF6',  // Purple
} as const

const getColor = (period: number): string => {
  if (period in SMA_COLORS) {
    return SMA_COLORS[period as keyof typeof SMA_COLORS]
  }
  return '#06B6D4' // Cyan for custom periods
}

export function SMAIndicator({ period }: SMAIndicatorProps) {
  const { chartData } = useChart()

  // Calculate SMA values (memoized to prevent recalculation on every render)
  const smaData = useMemo(() => {
    const closePrices = chartData.candlestickData.map(d => d.close)

    if (closePrices.length < period) {
      logger.log(`[SMA${period}] Insufficient data: ${closePrices.length} < ${period}`)
      return []
    }

    const smaValues = calculateSMA(closePrices, period)

    const data = chartData.candlestickData
      .map((d, i) => ({
        time: d.time,
        value: smaValues[i] ?? 0
      }))
      .filter(d => d.value > 0)

    logger.log(`[SMA${period}] Calculated ${data.length} points`)
    return data
  }, [chartData, period])

  // Use the reusable series hook - React manages lifecycle
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: getColor(period),
      lineWidth: 2,
      priceScaleId: 'right',  // Share scale with candlesticks
      title: `SMA(${period})`,
    },
    data: smaData,
    paneIndex: 0,  // Main pane (overlay on candlesticks)
  })

  return null  // Component manages chart series, doesn't render DOM
}
