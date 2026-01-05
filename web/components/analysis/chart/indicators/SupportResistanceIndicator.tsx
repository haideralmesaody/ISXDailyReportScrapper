/**
 * SupportResistanceIndicator - Support/Resistance Lines
 * Draws horizontal lines at calculated support/resistance levels
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateSupport, calculateResistance } from '@/lib/indicators/calculations'

interface SupportResistanceIndicatorProps {
  type: 'support' | 'resistance'
}

export function SupportResistanceIndicator({ type }: SupportResistanceIndicatorProps) {
  const { chartData } = useChart()

  const lineData = useMemo(() => {
    if (chartData.candlestickData.length < 20) return []

    // Pass entire OHLC data array (functions extract needed fields internally)
    const level = type === 'support'
      ? calculateSupport(chartData.candlestickData)
      : calculateResistance(chartData.candlestickData)

    if (level === 0) return []

    console.log(`[Support/Resistance] Calculated ${type} level:`, level, 'for', chartData.candlestickData.length, 'candles')

    // Create horizontal line across entire chart
    return [
      { time: chartData.candlestickData[0].time, value: level },
      { time: chartData.candlestickData[chartData.candlestickData.length - 1].time, value: level }
    ]
  }, [chartData, type])

  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: type === 'support' ? '#10B981' : '#EF4444',  // Green / Red
      lineWidth: 2,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',
      title: type === 'support' ? 'Support' : 'Resistance',
    },
    data: lineData,
    paneIndex: 0,
  })

  return null
}
