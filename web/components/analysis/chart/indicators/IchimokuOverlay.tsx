/**
 * IchimokuOverlay - Ichimoku Cloud Indicator
 * Five lines (Tenkan, Kijun, Senkou A, Senkou B, Chikou) overlaid on main candlestick chart
 * Comprehensive trend analysis system
 */

'use client'

import { useMemo } from 'react'
import { LineSeries, AreaSeries, LineStyle } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateIchimoku } from '@/lib/indicators/calculations'
import { useTrendSettings } from '@/lib/hooks/use-indicator-settings'

export function IchimokuOverlay() {
  const { chartData } = useChart()
  const { settings } = useTrendSettings()

  // Calculate Ichimoku (all five lines at once + cloud data)
  const { tenkanData, kijunData, senkouAData, senkouBData, chikouData, cloudTopData, cloudBottomData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.ichimokuSenkouBPeriod) {
      console.log('[Ichimoku] Insufficient data: need', settings.ichimokuSenkouBPeriod, 'have', closes.length)
      return {
        tenkanData: [],
        kijunData: [],
        senkouAData: [],
        senkouBData: [],
        chikouData: [],
        cloudTopData: [],
        cloudBottomData: []
      }
    }

    const ichimokuResult = calculateIchimoku(highs, lows, closes, settings.ichimokuTenkanPeriod, settings.ichimokuKijunPeriod, settings.ichimokuSenkouBPeriod)

    const tenkanData: Array<{ time: string; value: number }> = []
    const kijunData: Array<{ time: string; value: number }> = []
    const senkouAData: Array<{ time: string; value: number }> = []
    const senkouBData: Array<{ time: string; value: number }> = []
    const chikouData: Array<{ time: string; value: number }> = []

    // Cloud data: top boundary (max of Senkou A/B) and bottom boundary (min of Senkou A/B)
    const cloudTopData: Array<{ time: string; value: number }> = []
    const cloudBottomData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (ichimokuResult.tenkan[i] !== null) {
        tenkanData.push({ time: d.time as string, value: ichimokuResult.tenkan[i]! })
      }
      if (ichimokuResult.kijun[i] !== null) {
        kijunData.push({ time: d.time as string, value: ichimokuResult.kijun[i]! })
      }
      if (ichimokuResult.senkouA[i] !== null) {
        senkouAData.push({ time: d.time as string, value: ichimokuResult.senkouA[i]! })
      }
      if (ichimokuResult.senkouB[i] !== null) {
        senkouBData.push({ time: d.time as string, value: ichimokuResult.senkouB[i]! })
      }
      if (ichimokuResult.chikou[i] !== null) {
        chikouData.push({ time: d.time as string, value: ichimokuResult.chikou[i]! })
      }

      // Create cloud boundaries when both Senkou spans are available
      if (ichimokuResult.senkouA[i] !== null && ichimokuResult.senkouB[i] !== null) {
        const senkouA = ichimokuResult.senkouA[i]!
        const senkouB = ichimokuResult.senkouB[i]!

        cloudTopData.push({
          time: d.time as string,
          value: Math.max(senkouA, senkouB)
        })
        cloudBottomData.push({
          time: d.time as string,
          value: Math.min(senkouA, senkouB)
        })
      }
    })

    console.log('[Ichimoku] Calculated:', tenkanData.length, 'Tenkan,', kijunData.length, 'Kijun,', senkouAData.length, 'Senkou A/B,', cloudTopData.length, 'Cloud points')
    return { tenkanData, kijunData, senkouAData, senkouBData, chikouData, cloudTopData, cloudBottomData }
  }, [chartData, settings.ichimokuTenkanPeriod, settings.ichimokuKijunPeriod, settings.ichimokuSenkouBPeriod])

  // ===== CLOUD FILL (render first, appears behind lines) =====

  // Cloud Top Boundary - uses bullish green color with transparency
  useSeries({
    seriesType: AreaSeries,
    seriesOptions: {
      topColor: 'rgba(16, 185, 129, 0.4)',      // Green with 40% opacity
      bottomColor: 'rgba(16, 185, 129, 0.05)',  // Green with 5% opacity (gradient)
      lineColor: 'rgba(16, 185, 129, 0.8)',     // Green border
      lineWidth: 1,
      lineStyle: LineStyle.Solid,
      priceScaleId: 'right',
      title: 'Cloud Top',
    },
    data: cloudTopData,
    paneIndex: 0,
  })

  // Cloud Bottom Boundary - uses bearish red color with transparency
  useSeries({
    seriesType: AreaSeries,
    seriesOptions: {
      topColor: 'rgba(239, 68, 68, 0.05)',      // Red with 5% opacity (gradient)
      bottomColor: 'rgba(239, 68, 68, 0.4)',    // Red with 40% opacity
      lineColor: 'rgba(239, 68, 68, 0.8)',      // Red border
      lineWidth: 1,
      lineStyle: LineStyle.Solid,
      priceScaleId: 'right',
      title: 'Cloud Bottom',
    },
    data: cloudBottomData,
    paneIndex: 0,
  })

  // ===== ICHIMOKU LINES (render after cloud, appear on top) =====

  // Tenkan-sen (Conversion Line) - red, 9-period
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EF4444',
      lineWidth: 1,
      priceScaleId: 'right',
      title: 'Tenkan',
    },
    data: tenkanData,
    paneIndex: 0,  // Main chart overlay
  })

  // Kijun-sen (Base Line) - blue, 26-period
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#3B82F6',
      lineWidth: 1,
      priceScaleId: 'right',
      title: 'Kijun',
    },
    data: kijunData,
    paneIndex: 0,
  })

  // Senkou Span A (Leading Span A) - green
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#10B981',
      lineWidth: 1,
      priceScaleId: 'right',
      title: 'Senkou A',
    },
    data: senkouAData,
    paneIndex: 0,
  })

  // Senkou Span B (Leading Span B) - red
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EF4444',
      lineWidth: 1,
      priceScaleId: 'right',
      title: 'Senkou B',
    },
    data: senkouBData,
    paneIndex: 0,
  })

  // Chikou Span (Lagging Span) - yellow
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EAB308',
      lineWidth: 1,
      priceScaleId: 'right',
      title: 'Chikou',
    },
    data: chikouData,
    paneIndex: 0,
  })

  return null
}
