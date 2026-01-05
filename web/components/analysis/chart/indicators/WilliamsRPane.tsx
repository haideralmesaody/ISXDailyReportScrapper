/**
 * WilliamsRPane - Williams %R in separate pane
 * Manages 4 series: %R line and reference lines at -80, -50, -20
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for %R line (creates pane + price scale)
 * - Reference lines added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 4 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateWilliamsR } from '@/lib/indicators/calculations'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface WilliamsRPaneProps {
  paneIndex: number
}

export function WilliamsRPane({ paneIndex }: WilliamsRPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Track reference line series for cleanup
  const refLineSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Calculate Williams %R values
  const { williamsRData, refLineData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.williamsRPeriod) {
      console.log('[Williams %R] Insufficient data: need', settings.williamsRPeriod, 'have', closes.length)
      return { williamsRData: [], refLineData: [] }
    }

    const rValues = calculateWilliamsR(highs, lows, closes, settings.williamsRPeriod)

    const williamsRData = chartData.candlestickData
      .map((d, i) => ({
        time: d.time as string,
        value: rValues[i] ?? 0
      }))
      .filter(d => d.value !== 0 && !isNaN(d.value))

    // Reference line data (use full time range for continuous horizontal lines)
    const refLineData = williamsRData.map(d => ({ time: d.time, value: 0 }))

    console.log('[Williams %R] Calculated', williamsRData.length, '%R points')
    return { williamsRData, refLineData }
  }, [chartData, settings.williamsRPeriod])

  // Callback when %R series is created - add reference lines synchronously
  const handleRSeriesCreated = useCallback((rSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      // Use the paneIndex prop directly
      console.log('[Williams %R] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height
      const panes = chart.panes()
      const rPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      rPane.setHeight(constrainedHeight)

      console.log('[Williams %R] ✅ Pane configured: height =', constrainedHeight, 'px')

      // Add reference lines SYNCHRONOUSLY
      // -20 line (Overbought - Red)
      const rMinus20 = chart.addSeries(LineSeries, {
        color: '#EF4444',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'williamsR',
      }, paneIndex)
      refLineSeriesRef.current.push(rMinus20)

      requestAnimationFrame(() => {
        rMinus20.setData(refLineData.map(d => ({ ...d, value: settings.williamsROverbought })))
      })

      // -50 line (Neutral - Gray)
      const rMinus50 = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'williamsR',
      }, paneIndex)
      refLineSeriesRef.current.push(rMinus50)

      requestAnimationFrame(() => {
        rMinus50.setData(refLineData.map(d => ({ ...d, value: -50 })))
      })

      // -80 line (Oversold - Green)
      const rMinus80 = chart.addSeries(LineSeries, {
        color: '#10B981',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'williamsR',
      }, paneIndex)
      refLineSeriesRef.current.push(rMinus80)

      requestAnimationFrame(() => {
        rMinus80.setData(refLineData.map(d => ({ ...d, value: settings.williamsROversold })))
      })

      console.log('[Williams %R] ✅ Added 3 reference lines synchronously (pane', paneIndex, ')')
    } catch (error) {
      console.error('[Williams %R] Error creating reference lines:', error)
    }
  }, [chart, paneIndex, containerHeight, refLineData, settings.williamsROverbought, settings.williamsROversold])

  // Cleanup reference lines when component unmounts
  const handleRSeriesDestroyed = useCallback(() => {
    if (!chart) return

    refLineSeriesRef.current.forEach(series => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[Williams %R] Error removing reference line:', error)
      }
    })
    refLineSeriesRef.current = []
    console.log('[Williams %R] 🗑️  Cleaned up reference lines')
  }, [chart])

  // Williams %R Line (rose) - ONLY useSeries hook for main line
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#F43F5E',
      lineWidth: 2,
      title: `Williams %R (${settings.williamsRPeriod})`,
      priceScaleId: 'williamsR',
      // ✅ autoscaleInfoProvider suggests -100 to 0 range to TradingView's autoscaler
      autoscaleInfoProvider: () => ({
        priceRange: {
          minValue: -100,
          maxValue: 0,
        },
      }),
    },
    data: williamsRData,
    paneIndex,
    onCreated: handleRSeriesCreated,
    onDestroyed: handleRSeriesDestroyed,
    // ✅ TradingView v5.0 compliant: Configure price scale synchronously after series creation
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0,
      scaleMargins: { top: 0.05, bottom: 0.05 },
    },
  })

  return (
    <PaneOverlay paneIndex={paneIndex}>
      <IndicatorSettingsPopover
        indicatorName="Williams %R"
        settings={settings}
        fields={[
          { name: 'williamsRPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
          { name: 'williamsROverbought', label: 'Overbought', type: 'number', min: -50, max: 0 },
          { name: 'williamsROversold', label: 'Oversold', type: 'number', min: -100, max: -50 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
