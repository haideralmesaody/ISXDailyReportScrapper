/**
 * EMAIndicator - Exponential Moving Average Indicator
 * Reusable component for EMA with configurable period
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateEMA } from '@/lib/indicators/calculations'

interface EMAIndicatorProps {
  period: number
}

const getColor = (period: number): string => {
  return period === 20 ? '#EC4899' : '#14B8A6' // Pink for EMA20, Teal for custom
}

export function EMAIndicator({ period }: EMAIndicatorProps) {
  const { chartData } = useChart()

  const emaData = useMemo(() => {
    const closePrices = chartData.candlestickData.map(d => d.close)

    if (closePrices.length < period) {
      return []
    }

    const emaValues = calculateEMA(closePrices, period)

    return chartData.candlestickData
      .map((d, i) => ({
        time: d.time,
        value: emaValues[i] ?? 0
      }))
      .filter(d => d.value > 0)
  }, [chartData, period])

  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: getColor(period),
      lineWidth: 2,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',
      title: `EMA(${period})`,
    },
    data: emaData,
    paneIndex: 0,
  })

  return null
}
