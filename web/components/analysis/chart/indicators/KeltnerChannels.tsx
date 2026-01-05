/**
 * KeltnerChannels - Keltner Channels Volatility Indicator
 * Three lines (upper, middle, lower) overlaid on main candlestick chart
 *
 * Keltner Channels use ATR (Average True Range) to create volatility-based bands
 * around an EMA centerline. More responsive to volatility than Bollinger Bands.
 *
 * Trading signals:
 * - Price at upper channel = potential overbought (take profit/short)
 * - Price at lower channel = potential oversold (buy signal)
 * - Expanding channels = increasing volatility
 * - Contracting channels = decreasing volatility (breakout pending)
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateKeltnerChannels } from '@/lib/indicators/calculations'
import { useVolatilitySettings } from '@/lib/hooks/use-indicator-settings'

export function KeltnerChannels() {
  const { chartData } = useChart()
  const { settings } = useVolatilitySettings()

  // Calculate Keltner Channels (all three bands at once)
  const { upperData, middleData, lowerData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.keltnerPeriod) {
      console.log('[Keltner] Insufficient data: need', settings.keltnerPeriod, 'have', closes.length)
      return { upperData: [], middleData: [], lowerData: [] }
    }

    const kcResult = calculateKeltnerChannels(highs, lows, closes, settings.keltnerPeriod, settings.keltnerMultiplier)

    const upperData: Array<{ time: string; value: number }> = []
    const middleData: Array<{ time: string; value: number }> = []
    const lowerData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (kcResult.upper[i] !== null) {
        upperData.push({ time: d.time as string, value: kcResult.upper[i]! })
      }
      if (kcResult.middle[i] !== null) {
        middleData.push({ time: d.time as string, value: kcResult.middle[i]! })
      }
      if (kcResult.lower[i] !== null) {
        lowerData.push({ time: d.time as string, value: kcResult.lower[i]! })
      }
    })

    console.log('[Keltner] Calculated:', upperData.length, 'channel points')
    return { upperData, middleData, lowerData }
  }, [chartData, settings.keltnerPeriod, settings.keltnerMultiplier])

  // Upper Channel - blue dashed (volatility resistance)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#3B82F6',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',  // Share scale with candlesticks
      title: 'KC Upper',
    },
    data: upperData,
    paneIndex: 0,  // Main chart overlay
  })

  // Middle Channel - blue solid (EMA 20)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#3B82F6',
      lineWidth: 2,
      priceScaleId: 'right',
      title: 'KC Middle',
    },
    data: middleData,
    paneIndex: 0,
  })

  // Lower Channel - blue dashed (volatility support)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#3B82F6',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',
      title: 'KC Lower',
    },
    data: lowerData,
    paneIndex: 0,
  })

  return null
}
