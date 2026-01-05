/**
 * MACDPane - MACD Indicator in separate pane
 * Manages 3 series: MACD line, Signal line, and Histogram
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for MACD line (creates pane + price scale)
 * - Signal and Histogram lines added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 3 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, HistogramSeries } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateMACD } from '@/lib/indicators/calculations'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface MACDPaneProps {
  paneIndex: number
}

export function MACDPane({ paneIndex }: MACDPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Track additional series for cleanup
  const additionalSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Calculate MACD using custom periods from settings (all three series at once for efficiency)
  const { macdData, signalData, histogramData } = useMemo(() => {
    const closePrices = chartData.candlestickData.map(d => d.close)

    if (closePrices.length < settings.macdSlowPeriod) {
      console.log('[MACD] Insufficient data: need', settings.macdSlowPeriod, 'have', closePrices.length)
      return { macdData: [], signalData: [], histogramData: [] }
    }

    const macdResult = calculateMACD(closePrices, settings.macdFastPeriod, settings.macdSlowPeriod, settings.macdSignalPeriod)

    const macdData: Array<{ time: number; value: number }> = []
    const signalData: Array<{ time: number; value: number }> = []
    const histogramData: Array<{ time: number; value: number; color: string }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (macdResult.macd[i] !== null) {
        macdData.push({ time: d.time, value: macdResult.macd[i]! })
      }
      if (macdResult.signal[i] !== null) {
        signalData.push({ time: d.time, value: macdResult.signal[i]! })
      }
      if (macdResult.histogram[i] !== null) {
        const histValue = macdResult.histogram[i]!
        histogramData.push({
          time: d.time,
          value: histValue,
          color: histValue >= 0 ? '#10b981' : '#ef4444'
        })
      }
    })

    console.log('[MACD] Calculated:', macdData.length, 'MACD points,', signalData.length, 'Signal points,', histogramData.length, 'Histogram bars')
    return { macdData, signalData, histogramData }
  }, [chartData, settings.macdFastPeriod, settings.macdSlowPeriod, settings.macdSignalPeriod])

  // Callback when MACD main series is created - add signal + histogram synchronously
  const handleMACDSeriesCreated = useCallback((macdSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      // 1. Use the paneIndex prop directly - TradingView guarantees series is in this pane
      // The paneIndex prop is the authoritative source, not panes.length
      console.log('[MACD] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // 2. Set pane height using paneIndex prop
      const panes = chart.panes()
      const macdPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      macdPane.setHeight(constrainedHeight)

      console.log('[MACD] ✅ Pane configured: height =', constrainedHeight, 'px')

      // 3. Add Signal and Histogram SYNCHRONOUSLY using paneIndex prop
      // All use the same paneIndex and priceScaleId as the main MACD line

      // Signal Line (orange)
      const signalLine = chart.addSeries(LineSeries, {
        color: '#FF9800',
        lineWidth: 2,
        title: 'Signal',
        priceScaleId: 'macd',
      }, paneIndex)
      additionalSeriesRef.current.push(signalLine)

      // Defer data setting to give TradingView time to initialize series
      requestAnimationFrame(() => {
        signalLine.setData(signalData)
      })

      // Histogram (green/red)
      const histogram = chart.addSeries(HistogramSeries, {
        priceFormat: { type: 'price', precision: 4 },
        priceScaleId: 'macd',
      }, paneIndex)
      additionalSeriesRef.current.push(histogram)

      // Defer data setting to give TradingView time to initialize series
      requestAnimationFrame(() => {
        histogram.setData(histogramData)
      })

      console.log('[MACD] ✅ Added Signal + Histogram synchronously (pane', paneIndex, ')')
    } catch (error) {
      console.error('[MACD] Error creating additional series:', error)
    }
  }, [chart, paneIndex, containerHeight, signalData, histogramData])

  // Cleanup additional series when component unmounts
  const handleMACDSeriesDestroyed = useCallback(() => {
    if (!chart) return

    additionalSeriesRef.current.forEach(series => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[MACD] Error removing series:', error)
      }
    })
    additionalSeriesRef.current = []
    console.log('[MACD] 🗑️  Cleaned up additional series')
  }, [chart])

  // MACD Line (blue) - ONLY useSeries hook for main line
  // This creates the pane and price scale, then adds Signal + Histogram in callback
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#2196F3',
      lineWidth: 2,
      title: `MACD (${settings.macdFastPeriod},${settings.macdSlowPeriod},${settings.macdSignalPeriod})`,
      priceScaleId: 'macd',  // All 3 series share 'macd' scale
    },
    data: macdData,
    paneIndex,
    onCreated: handleMACDSeriesCreated,
    onDestroyed: handleMACDSeriesDestroyed,
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
        indicatorName="MACD"
        settings={settings}
        fields={[
          { name: 'macdFastPeriod', label: 'Fast Period', type: 'number', min: 2, max: 50 },
          { name: 'macdSlowPeriod', label: 'Slow Period', type: 'number', min: 10, max: 100 },
          { name: 'macdSignalPeriod', label: 'Signal Period', type: 'number', min: 2, max: 50 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
