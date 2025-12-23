/**
 * MFIPane - Money Flow Index indicator in separate pane
 * Manages 4 series: MFI line, and reference lines at 20, 50, 80
 *
 * MFI is volume-weighted RSI. MFI > 80 = overbought, MFI < 20 = oversold
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for main MFI line (creates pane + price scale)
 * - Reference lines added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 4 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateMFI } from '@/lib/indicators/calculations'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface MFIPaneProps {
  paneIndex: number
}

export function MFIPane({ paneIndex }: MFIPaneProps) {
  const { chartData, chart, containerHeight } = useChart()

  // Track reference line series for cleanup
  const refLineSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Calculate MFI values (uses OHLCV data)
  const mfiData = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)
    const volumes = chartData.volumeData.map(d => d.value)

    if (highs.length < 15 || volumes.length === 0) {
      console.log('[MFI] Insufficient data')
      return []
    }

    const mfiValues = calculateMFI(highs, lows, closes, volumes, 14)

    const data = chartData.candlestickData
      .map((d, i) => ({
        time: d.time,
        value: mfiValues[i] ?? 0
      }))
      .filter(d => d.value > 0)

    console.log('[MFI] Calculated', data.length, 'MFI points')
    return data
  }, [chartData])

  // Reference lines (horizontal lines at 20, 50, 80)
  // Use full time range for continuous horizontal lines across the chart
  const refLineData = useMemo(() => {
    return mfiData.map(d => ({ time: d.time, value: 0 }))
  }, [mfiData])

  // Callback when MFI main series is created - add reference lines synchronously
  const handleMFISeriesCreated = useCallback((mfiSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      console.log('[MFI] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height using paneIndex prop
      const panes = chart.panes()
      const mfiPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      mfiPane.setHeight(constrainedHeight)

      console.log('[MFI] ✅ Pane configured: height =', constrainedHeight, 'px')

      // Add reference lines SYNCHRONOUSLY using paneIndex prop

      // Overbought line (80 - Red)
      const mfi80 = chart.addSeries(LineSeries, {
        color: '#EF4444',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'mfi',
      }, paneIndex)
      refLineSeriesRef.current.push(mfi80)

      requestAnimationFrame(() => {
        mfi80.setData(refLineData.map(d => ({ ...d, value: 80 })))
      })

      // 50 line (Neutral - Gray)
      const mfi50 = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'mfi',
      }, paneIndex)
      refLineSeriesRef.current.push(mfi50)

      requestAnimationFrame(() => {
        mfi50.setData(refLineData.map(d => ({ ...d, value: 50 })))
      })

      // Oversold line (20 - Green)
      const mfi20 = chart.addSeries(LineSeries, {
        color: '#10B981',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'mfi',
      }, paneIndex)
      refLineSeriesRef.current.push(mfi20)

      requestAnimationFrame(() => {
        mfi20.setData(refLineData.map(d => ({ ...d, value: 20 })))
      })

      console.log('[MFI] ✅ Added 3 reference lines synchronously (pane', paneIndex, ')')
    } catch (error) {
      console.error('[MFI] Error creating reference lines:', error)
    }
  }, [chart, paneIndex, containerHeight, refLineData])

  // Cleanup reference lines when component unmounts
  const handleMFISeriesDestroyed = useCallback(() => {
    if (!chart) return

    refLineSeriesRef.current.forEach(series => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[MFI] Error removing reference line:', error)
      }
    })
    refLineSeriesRef.current = []
    console.log('[MFI] 🗑️  Cleaned up reference lines')
  }, [chart])

  // MFI Line (purple) - ONLY useSeries hook for main line
  // This creates the pane and price scale, then adds reference lines in callback
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#9C27B0',  // Purple - same as RSI for consistency
      lineWidth: 2,
      title: 'MFI (14)',
      priceScaleId: 'mfi',
      autoscaleInfoProvider: () => ({
        priceRange: {
          minValue: 0,
          maxValue: 100,
        },
      }),
    },
    data: mfiData,
    paneIndex,
    onCreated: handleMFISeriesCreated,
    onDestroyed: handleMFISeriesDestroyed,
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
        indicatorName="Money Flow Index"
        settings={{}}
        fields={[]}
        onUpdate={() => {}}
        description="Volume-weighted RSI. MFI > 80 = overbought (sell signal), MFI < 20 = oversold (buy signal). Combines price and volume to identify money flow."
      />
    </PaneOverlay>
  )
}
