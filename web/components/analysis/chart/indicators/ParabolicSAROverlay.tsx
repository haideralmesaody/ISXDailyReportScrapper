/**
 * ParabolicSAROverlay - Parabolic SAR (Stop and Reverse) overlay on main chart
 *
 * Trend-following indicator showing dots above (downtrend) or below (uptrend) candles
 * Helps identify trend direction and potential reversal points
 *
 * Visual:
 * - Blue dots below candles = Uptrend (buy signal)
 * - Red dots above candles = Downtrend (sell signal)
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateParabolicSAR } from '@/lib/indicators/calculations'

export function ParabolicSAROverlay() {
  const { chartData } = useChart()

  // Calculate Parabolic SAR from OHLC data
  const { uptrendData, downtrendData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (highs.length === 0) {
      console.log('[Parabolic SAR] Insufficient data')
      return { uptrendData: [], downtrendData: [] }
    }

    const sarValues = calculateParabolicSAR(highs, lows, closes)

    const uptrend: Array<{ time: string; value: number }> = []
    const downtrend: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (sarValues[i] !== null) {
        const sar = sarValues[i]!
        if (sar.isUptrend) {
          uptrend.push({
            time: d.time as string,
            value: sar.value
          })
        } else {
          downtrend.push({
            time: d.time as string,
            value: sar.value
          })
        }
      }
    })

    console.log('[Parabolic SAR] Calculated:', uptrend.length, 'uptrend points,', downtrend.length, 'downtrend points')
    return { uptrendData: uptrend, downtrendData: downtrend }
  }, [chartData])

  // Render uptrend dots (blue, below candles)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#2196F3',  // Blue - bullish
      lineWidth: 2,
      lineStyle: LineStyle.Dotted,
      title: 'SAR ↑',
      priceScaleId: 'right',
      crosshairMarkerVisible: true,
      crosshairMarkerRadius: 4,
    },
    data: uptrendData,
    paneIndex: 0,  // Main chart overlay
  })

  // Render downtrend dots (red, above candles)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EF4444',  // Red - bearish
      lineWidth: 2,
      lineStyle: LineStyle.Dotted,
      title: 'SAR ↓',
      priceScaleId: 'right',
      crosshairMarkerVisible: true,
      crosshairMarkerRadius: 4,
    },
    data: downtrendData,
    paneIndex: 0,  // Main chart overlay
  })

  return null  // Component manages chart series, doesn't render DOM
}
