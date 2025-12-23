/**
 * GeneralBollingerBandsOverlay - Customizable Bollinger Bands Indicator
 * Three lines (upper, middle, lower) with user-defined period and standard deviation
 * Overlaid on main candlestick chart
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateBollingerBands } from '@/lib/indicators/calculations'
import { useVolatilitySettings } from '@/lib/hooks/use-indicator-settings'

export function GeneralBollingerBandsOverlay() {
  const { chartData } = useChart()
  const { settings } = useVolatilitySettings()

  // Calculate General Bollinger Bands using custom period and std dev from settings
  const { upperData, middleData, lowerData } = useMemo(() => {
    const closePrices = chartData.candlestickData.map(d => d.close)

    if (closePrices.length < settings.generalBollingerPeriod) {
      console.log('[General Bollinger] Insufficient data: need', settings.generalBollingerPeriod, 'have', closePrices.length)
      return { upperData: [], middleData: [], lowerData: [] }
    }

    const bbResult = calculateBollingerBands(
      closePrices,
      settings.generalBollingerPeriod,
      settings.generalBollingerStdDev
    )

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

    console.log('[General Bollinger] Calculated:', upperData.length, 'band points (period:', settings.generalBollingerPeriod, 'stdDev:', settings.generalBollingerStdDev, ')')
    return { upperData, middleData, lowerData }
  }, [chartData, settings.generalBollingerPeriod, settings.generalBollingerStdDev])

  // Upper Band - sky dashed
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#0EA5E9',  // sky-500
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',  // Share scale with candlesticks
      title: `Gen BB Upper (${settings.generalBollingerPeriod})`,
    },
    data: upperData,
    paneIndex: 0,  // Main chart overlay
  })

  // Middle Band - sky solid (SMA)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#0EA5E9',  // sky-500
      lineWidth: 2,
      priceScaleId: 'right',
      title: `Gen BB Middle (${settings.generalBollingerPeriod})`,
    },
    data: middleData,
    paneIndex: 0,
  })

  // Lower Band - sky dashed
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#0EA5E9',  // sky-500
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceScaleId: 'right',
      title: `Gen BB Lower (${settings.generalBollingerPeriod})`,
    },
    data: lowerData,
    paneIndex: 0,
  })

  return null
}
