/**
 * ADXPane - Average Directional Index in separate pane
 * Manages 4 series: ADX line, +DI line, -DI line, and reference line at 25
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for ADX line (creates pane + price scale)
 * - +DI, -DI, and reference line added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 4 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateADX } from '@/lib/indicators/calculations'
import { useTrendSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface ADXPaneProps {
  paneIndex: number
}

export function ADXPane({ paneIndex }: ADXPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useTrendSettings()

  // Track additional series for cleanup
  const additionalSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Calculate ADX, +DI, -DI
  const { adxData, plusDIData, minusDIData, refLineData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.adxPeriod) {
      console.log('[ADX] Insufficient data: need', settings.adxPeriod, 'have', closes.length)
      return { adxData: [], plusDIData: [], minusDIData: [], refLineData: [] }
    }

    const adxResult = calculateADX(highs, lows, closes, settings.adxPeriod)

    const adxData: Array<{ time: string; value: number }> = []
    const plusDIData: Array<{ time: string; value: number }> = []
    const minusDIData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (adxResult.adx[i] !== null) {
        adxData.push({ time: d.time as string, value: adxResult.adx[i]! })
      }
      if (adxResult.plusDI[i] !== null) {
        plusDIData.push({ time: d.time as string, value: adxResult.plusDI[i]! })
      }
      if (adxResult.minusDI[i] !== null) {
        minusDIData.push({ time: d.time as string, value: adxResult.minusDI[i]! })
      }
    })

    // Reference line data (use full time range for continuous horizontal line)
    const refLineData = adxData.map(d => ({ time: d.time, value: 0 }))

    console.log('[ADX] Calculated:', adxData.length, 'ADX points,', plusDIData.length, '+DI,', minusDIData.length, '-DI')
    return { adxData, plusDIData, minusDIData, refLineData }
  }, [chartData, settings.adxPeriod])

  // Callback when ADX series is created - add +DI, -DI, and reference line synchronously
  const handleADXSeriesCreated = useCallback((adxSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      // 1. Use the paneIndex prop directly - TradingView guarantees series is in this pane
      console.log('[ADX] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // 2. Set pane height using paneIndex prop
      const panes = chart.panes()
      const adxPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      adxPane.setHeight(constrainedHeight)

      console.log('[ADX] ✅ Pane configured: height =', constrainedHeight, 'px')

      // 3. Add +DI line SYNCHRONOUSLY
      const plusDILine = chart.addSeries(LineSeries, {
        color: '#10B981',
        lineWidth: 2,
        title: '+DI',
        priceScaleId: 'adx',
      }, paneIndex)
      additionalSeriesRef.current.push(plusDILine)

      requestAnimationFrame(() => {
        plusDILine.setData(plusDIData)
      })

      // 4. Add -DI line SYNCHRONOUSLY
      const minusDILine = chart.addSeries(LineSeries, {
        color: '#EF4444',
        lineWidth: 2,
        title: '-DI',
        priceScaleId: 'adx',
      }, paneIndex)
      additionalSeriesRef.current.push(minusDILine)

      requestAnimationFrame(() => {
        minusDILine.setData(minusDIData)
      })

      // 5. Add reference line at threshold (trend strength threshold)
      const adxThreshold = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'adx',
      }, paneIndex)
      additionalSeriesRef.current.push(adxThreshold)

      requestAnimationFrame(() => {
        adxThreshold.setData(refLineData.map(d => ({ ...d, value: settings.adxTrendThreshold })))
      })

      console.log('[ADX] ✅ Added +DI, -DI, and reference line synchronously (pane', paneIndex, ')')
    } catch (error) {
      console.error('[ADX] Error creating additional series:', error)
    }
  }, [chart, paneIndex, containerHeight, plusDIData, minusDIData, refLineData, settings.adxTrendThreshold])

  // Cleanup additional series when component unmounts
  const handleADXSeriesDestroyed = useCallback(() => {
    if (!chart) return

    additionalSeriesRef.current.forEach(series => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[ADX] Error removing series:', error)
      }
    })
    additionalSeriesRef.current = []
    console.log('[ADX] 🗑️  Cleaned up additional series')
  }, [chart])

  // ADX Line (amber) - ONLY useSeries hook for main line
  // This creates the pane and price scale, then adds +DI, -DI, reference line in callback
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#D97706',
      lineWidth: 2,
      title: `ADX (${settings.adxPeriod})`,
      priceScaleId: 'adx',  // All 4 series share 'adx' scale
      // ✅ autoscaleInfoProvider suggests 0-100 range to TradingView's autoscaler
      autoscaleInfoProvider: () => ({
        priceRange: {
          minValue: 0,
          maxValue: 100,
        },
      }),
    },
    data: adxData,
    paneIndex,
    onCreated: handleADXSeriesCreated,
    onDestroyed: handleADXSeriesDestroyed,
    // ✅ TradingView v5.0 compliant: Configure price scale synchronously after series creation
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0,
      scaleMargins: { top: 0.1, bottom: 0.1 },
    },
  })

  return (
    <PaneOverlay paneIndex={paneIndex}>
      <IndicatorSettingsPopover
        indicatorName="ADX"
        settings={settings}
        fields={[
          { name: 'adxPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
          { name: 'adxTrendThreshold', label: 'Trend Threshold', type: 'number', min: 0, max: 100 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
