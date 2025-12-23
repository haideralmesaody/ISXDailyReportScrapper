/**
 * DonchianChannels - Donchian Channels Breakout Indicator
 * Three lines (upper, middle, lower) overlaid on main candlestick chart
 *
 * Donchian Channels show the highest high and lowest low over a period,
 * creating a "channel" that identifies potential breakout points.
 *
 * Trading signals:
 * - Price breaking above upper channel = bullish breakout (buy signal)
 * - Price breaking below lower channel = bearish breakout (sell signal)
 * - Price bouncing at channels = support/resistance confirmation
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateDonchianChannels } from '@/lib/indicators/calculations'
import { useSupportResistanceSettings } from '@/lib/hooks/use-indicator-settings'

export function DonchianChannels() {
  const { chartData } = useChart()
  const { settings } = useSupportResistanceSettings()

  // Calculate Donchian Channels (all three bands at once)
  const { upperData, middleData, lowerData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)

    if (highs.length < settings.donchianPeriod) {
      console.log('[Donchian] Insufficient data: need', settings.donchianPeriod, 'have', highs.length)
      return { upperData: [], middleData: [], lowerData: [] }
    }

    const dcResult = calculateDonchianChannels(highs, lows, settings.donchianPeriod)

    const upperData: Array<{ time: string; value: number }> = []
    const middleData: Array<{ time: string; value: number }> = []
    const lowerData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (dcResult.upper[i] !== null) {
        upperData.push({ time: d.time as string, value: dcResult.upper[i]! })
      }
      if (dcResult.middle[i] !== null) {
        middleData.push({ time: d.time as string, value: dcResult.middle[i]! })
      }
      if (dcResult.lower[i] !== null) {
        lowerData.push({ time: d.time as string, value: dcResult.lower[i]! })
      }
    })

    console.log('[Donchian] Calculated:', upperData.length, 'channel points')
    return { upperData, middleData, lowerData }
  }, [chartData, settings.donchianPeriod])

  // Upper Channel - green dashed (breakout resistance)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#10B981',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',  // Share scale with candlesticks
      title: 'DC Upper',
    },
    data: upperData,
    paneIndex: 0,  // Main chart overlay
  })

  // Middle Channel - gray solid (midpoint)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#6B7280',
      lineWidth: 1,
      priceScaleId: 'right',
      title: 'DC Middle',
    },
    data: middleData,
    paneIndex: 0,
  })

  // Lower Channel - red dashed (breakout support)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EF4444',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',
      title: 'DC Lower',
    },
    data: lowerData,
    paneIndex: 0,
  })

  return null
}
