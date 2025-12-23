/**
 * CCIPane - Commodity Channel Index in separate pane
 * Manages 4 series: CCI line and reference lines at -100, 0, +100
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for CCI line (creates pane + price scale)
 * - Reference lines added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 4 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateCCI } from '@/lib/indicators/calculations'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface CCIPaneProps {
  paneIndex: number
}

export function CCIPane({ paneIndex }: CCIPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Track reference line series for cleanup
  const refLineSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Calculate CCI values
  const { cciData, refLineData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.cciPeriod) {
      console.log('[CCI] Insufficient data: need', settings.cciPeriod, 'have', closes.length)
      return { cciData: [], refLineData: [] }
    }

    const cciValues = calculateCCI(highs, lows, closes, settings.cciPeriod)

    const cciData = chartData.candlestickData
      .map((d, i) => ({
        time: d.time as string,
        value: cciValues[i] ?? 0
      }))
      .filter(d => d.value !== 0 && !isNaN(d.value))

    // Reference line data (use full time range for continuous horizontal lines)
    const refLineData = cciData.map(d => ({ time: d.time, value: 0 }))

    console.log('[CCI] Calculated', cciData.length, 'CCI points')
    return { cciData, refLineData }
  }, [chartData, settings.cciPeriod])

  // Callback when CCI series is created - add reference lines synchronously
  const handleCCISeriesCreated = useCallback((cciSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      // Use the paneIndex prop directly
      console.log('[CCI] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height
      const panes = chart.panes()
      const cciPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      cciPane.setHeight(constrainedHeight)

      console.log('[CCI] ✅ Pane configured: height =', constrainedHeight, 'px')

      // Add reference lines SYNCHRONOUSLY
      // +100 line (Overbought - Red)
      const cci100 = chart.addSeries(LineSeries, {
        color: '#EF4444',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'cci',
      }, paneIndex)
      refLineSeriesRef.current.push(cci100)

      requestAnimationFrame(() => {
        cci100.setData(refLineData.map(d => ({ ...d, value: settings.cciOverbought })))
      })

      // 0 line (Neutral - Gray)
      const cci0 = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'cci',
      }, paneIndex)
      refLineSeriesRef.current.push(cci0)

      requestAnimationFrame(() => {
        cci0.setData(refLineData.map(d => ({ ...d, value: 0 })))
      })

      // -100 line (Oversold - Green)
      const cciMinus100 = chart.addSeries(LineSeries, {
        color: '#10B981',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'cci',
      }, paneIndex)
      refLineSeriesRef.current.push(cciMinus100)

      requestAnimationFrame(() => {
        cciMinus100.setData(refLineData.map(d => ({ ...d, value: settings.cciOversold })))
      })

      console.log('[CCI] ✅ Added 3 reference lines synchronously (pane', paneIndex, ')')
    } catch (error) {
      console.error('[CCI] Error creating reference lines:', error)
    }
  }, [chart, paneIndex, containerHeight, refLineData, settings.cciOverbought, settings.cciOversold])

  // Cleanup reference lines when component unmounts
  const handleCCISeriesDestroyed = useCallback(() => {
    if (!chart) return

    refLineSeriesRef.current.forEach(series => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[CCI] Error removing reference line:', error)
      }
    })
    refLineSeriesRef.current = []
    console.log('[CCI] 🗑️  Cleaned up reference lines')
  }, [chart])

  // CCI Line (lime) - ONLY useSeries hook for main line
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#84CC16',
      lineWidth: 2,
      title: `CCI (${settings.cciPeriod})`,
      priceScaleId: 'cci',
    },
    data: cciData,
    paneIndex,
    onCreated: handleCCISeriesCreated,
    onDestroyed: handleCCISeriesDestroyed,
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
        indicatorName="CCI"
        settings={settings}
        fields={[
          { name: 'cciPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
          { name: 'cciOverbought', label: 'Overbought', type: 'number', min: 50, max: 300 },
          { name: 'cciOversold', label: 'Oversold', type: 'number', min: -300, max: -50 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
