/**
 * StochasticPane - Stochastic Oscillator in separate pane
 * Manages 5 series: %K line, %D line, and reference lines at 20, 50, 80
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for %K line (creates pane + price scale)
 * - %D line + reference lines added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 5 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateStochastic } from '@/lib/indicators/calculations'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface StochasticPaneProps {
  paneIndex: number
}

export function StochasticPane({ paneIndex }: StochasticPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Track additional series for cleanup
  const additionalSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Calculate Stochastic %K and %D
  const { kData, dData, refLineData } = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.stochasticKPeriod) {
      console.log('[Stochastic] Insufficient data: need', settings.stochasticKPeriod, 'have', closes.length)
      return { kData: [], dData: [], refLineData: [] }
    }

    const stochResult = calculateStochastic(highs, lows, closes, settings.stochasticKPeriod, settings.stochasticDPeriod)

    const kData: Array<{ time: string; value: number }> = []
    const dData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (stochResult.k[i] !== null) {
        kData.push({ time: d.time as string, value: stochResult.k[i]! })
      }
      if (stochResult.d[i] !== null) {
        dData.push({ time: d.time as string, value: stochResult.d[i]! })
      }
    })

    // Reference line data (use full time range for continuous horizontal lines)
    const refLineData = kData.map(d => ({ time: d.time, value: 0 }))

    console.log('[Stochastic] Calculated:', kData.length, '%K points,', dData.length, '%D points')
    return { kData, dData, refLineData }
  }, [chartData, settings.stochasticKPeriod, settings.stochasticDPeriod])

  // Callback when %K series is created - add %D + reference lines synchronously
  const handleKSeriesCreated = useCallback((kSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      // 1. Use the paneIndex prop directly - TradingView guarantees series is in this pane
      console.log('[Stochastic] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // 2. Set pane height using paneIndex prop
      const panes = chart.panes()
      const stochPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      stochPane.setHeight(constrainedHeight)

      console.log('[Stochastic] ✅ Pane configured: height =', constrainedHeight, 'px')

      // 3. Add %D line SYNCHRONOUSLY using paneIndex prop
      const dLine = chart.addSeries(LineSeries, {
        color: '#FF9800',
        lineWidth: 2,
        title: '%D',
        priceScaleId: 'stochastic',
      }, paneIndex)
      additionalSeriesRef.current.push(dLine)

      // Defer data setting to give TradingView time to initialize series
      requestAnimationFrame(() => {
        dLine.setData(dData)
      })

      // 4. Add reference lines SYNCHRONOUSLY
      // 80 line (Overbought - Red)
      const stoch80 = chart.addSeries(LineSeries, {
        color: '#EF4444',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'stochastic',
      }, paneIndex)
      additionalSeriesRef.current.push(stoch80)

      requestAnimationFrame(() => {
        stoch80.setData(refLineData.map(d => ({ ...d, value: settings.stochasticOverbought })))
      })

      // 50 line (Neutral - Gray)
      const stoch50 = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'stochastic',
      }, paneIndex)
      additionalSeriesRef.current.push(stoch50)

      requestAnimationFrame(() => {
        stoch50.setData(refLineData.map(d => ({ ...d, value: 50 })))
      })

      // 20 line (Oversold - Green)
      const stoch20 = chart.addSeries(LineSeries, {
        color: '#10B981',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'stochastic',
      }, paneIndex)
      additionalSeriesRef.current.push(stoch20)

      requestAnimationFrame(() => {
        stoch20.setData(refLineData.map(d => ({ ...d, value: settings.stochasticOversold })))
      })

      console.log('[Stochastic] ✅ Added %D + 3 reference lines synchronously (pane', paneIndex, ')')
    } catch (error) {
      console.error('[Stochastic] Error creating additional series:', error)
    }
  }, [chart, paneIndex, containerHeight, dData, refLineData, settings.stochasticOverbought, settings.stochasticOversold])

  // Cleanup additional series when component unmounts
  const handleKSeriesDestroyed = useCallback(() => {
    if (!chart) return

    additionalSeriesRef.current.forEach(series => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[Stochastic] Error removing series:', error)
      }
    })
    additionalSeriesRef.current = []
    console.log('[Stochastic] 🗑️  Cleaned up additional series')
  }, [chart])

  // %K Line (pink) - ONLY useSeries hook for main line
  // This creates the pane and price scale, then adds %D + reference lines in callback
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EC4899',
      lineWidth: 2,
      title: '%K',
      priceScaleId: 'stochastic',  // All 5 series share 'stochastic' scale
      // ✅ autoscaleInfoProvider suggests 0-100 range to TradingView's autoscaler
      autoscaleInfoProvider: () => ({
        priceRange: {
          minValue: 0,
          maxValue: 100,
        },
      }),
    },
    data: kData,
    paneIndex,
    onCreated: handleKSeriesCreated,
    onDestroyed: handleKSeriesDestroyed,
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
        indicatorName="Stochastic"
        settings={settings}
        fields={[
          { name: 'stochasticKPeriod', label: '%K Period', type: 'number', min: 2, max: 100 },
          { name: 'stochasticDPeriod', label: '%D Period', type: 'number', min: 2, max: 50 },
          { name: 'stochasticOverbought', label: 'Overbought', type: 'number', min: 50, max: 100 },
          { name: 'stochasticOversold', label: 'Oversold', type: 'number', min: 0, max: 50 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
