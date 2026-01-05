/**
 * BollingerBandsOverlay - Bollinger Bands Indicator
 * Three lines (upper, middle, lower) overlaid on main candlestick chart
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateBollingerBands } from '@/lib/indicators/calculations'
import { useVolatilitySettings } from '@/lib/hooks/use-indicator-settings'

export function BollingerBandsOverlay() {
  const { chartData } = useChart()
  const { settings } = useVolatilitySettings()

  // Calculate Bollinger Bands using custom period and std dev from settings (all three bands at once)
  const { upperData, middleData, lowerData } = useMemo(() => {
    const closePrices = chartData.candlestickData.map(d => d.close)

    if (closePrices.length < settings.bollingerPeriod) {
      console.log('[Bollinger] Insufficient data: need', settings.bollingerPeriod, 'have', closePrices.length)
      return { upperData: [], middleData: [], lowerData: [] }
    }

    const bbResult = calculateBollingerBands(closePrices, settings.bollingerPeriod, settings.bollingerStdDev)

    const upperData: Array<{ time: string; value: number }> = []
    const middleData: Array<{ time: string; value: number }> = []
    const lowerData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (bbResult.upper[i] !== null) {
        upperData.push({ time: d.time as string, value: bbResult.upper[i]! })
      }
      if (bbResult.middle[i] !== null) {
        middleData.push({ time: d.time as string, value: bbResult.middle[i]! })
      }
      if (bbResult.lower[i] !== null) {
        lowerData.push({ time: d.time as string, value: bbResult.lower[i]! })
      }
    })

    console.log('[Bollinger] Calculated:', upperData.length, 'band points')
    return { upperData, middleData, lowerData }
  }, [chartData, settings.bollingerPeriod, settings.bollingerStdDev])

  // Upper Band - cyan dashed
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#22D3EE',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',  // Share scale with candlesticks
      title: 'BB Upper',
    },
    data: upperData,
    paneIndex: 0,  // Main chart overlay
  })

  // Middle Band - cyan solid (SMA 20)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#22D3EE',
      lineWidth: 2,
      priceScaleId: 'right',
      title: 'BB Middle',
    },
    data: middleData,
    paneIndex: 0,
  })

  // Lower Band - cyan dashed
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#22D3EE',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',
      title: 'BB Lower',
    },
    data: lowerData,
    paneIndex: 0,
  })

  return null
}
