/**
 * RSIPane - RSI Indicator in separate pane
 * Manages 4 series: RSI line, and reference lines at 30, 50, 70
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for main RSI line (creates pane + price scale)
 * - Reference lines added synchronously in onCreated callback (prevents race conditions)
 * - This pattern ensures all 4 series end up in the same pane
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface RSIPaneProps {
  paneIndex: number
}

export function RSIPane({ paneIndex }: RSIPaneProps) {
  const { chartData, chart, containerHeight, indicatorData } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Track reference line series for cleanup
  const refLineSeriesRef = useRef<ISeriesApi<any>[]>([])

  // Use locally calculated RSI data from context
  const rsiData = useMemo(() => {
    // Check if RSI data is available and chart data exists
    if (!indicatorData.rsiData || indicatorData.rsiData.size === 0) {
      // Only log when we're actually missing data (not during initial render)
      if (chartData.candlestickData.length > 0) {
        console.log('[RSI] Waiting for RSI calculation...')
      }
      return []
    }

    if (chartData.candlestickData.length === 0) {
      return []
    }

    // Map chart data to RSI data using the timeframe-transformed chart dates
    const data = chartData.candlestickData
      .map((d) => {
        const dateStr =
          typeof d.time === 'string'
            ? d.time
            : typeof d.time === 'number'
              ? new Date(d.time * 1000).toISOString().split('T')[0]
              : new Date((d.time as any).year, (d.time as any).month - 1, (d.time as any).day)
                  .toISOString()
                  .split('T')[0]

        const rsiValue = indicatorData.rsiData!.get(dateStr)
        return {
          time: d.time,
          value: rsiValue ?? null,
        }
      })
      .filter((d) => d.value !== null && d.value !== undefined && !isNaN(d.value))

    if (data.length > 0) {
      console.log(
        '[RSI] Using calculated RSI data:',
        data.length,
        'RSI points from',
        indicatorData.rsiData.size,
        'available values'
      )
    } else {
      console.log('[RSI] No valid RSI data points after mapping')
    }

    return data
  }, [chartData.candlestickData, indicatorData.rsiData]) // Simplified dependencies

  // Reference lines (horizontal lines at 30, 50, 70)
  // Use the same time range as RSI data for perfect alignment
  const refLineData = useMemo(() => {
    // Always use chart data time range for reference lines
    if (chartData.candlestickData.length > 0) {
      return chartData.candlestickData.map((d) => ({ time: d.time, value: 0 }))
    }

    // Fallback: return empty array
    return []
  }, [chartData.candlestickData])

  // Callback when RSI main series is created - add reference lines immediately
  const handleRSISeriesCreated = useCallback(
    (rsiSeries: ISeriesApi<any>) => {
      if (!chart) return

      try {
        // 1. Set pane height using paneIndex prop
        const panes = chart.panes()
        const rsiPane = panes[paneIndex]
        const maxHeight = Math.floor(containerHeight * 0.9)
        const desiredHeight = Math.floor(containerHeight * 0.15) // 15% default
        const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
        rsiPane.setHeight(constrainedHeight)

        console.log('[RSI] ✅ Pane configured: height =', constrainedHeight, 'px')

        // 2. Add reference lines immediately using paneIndex prop
        // All use the same paneIndex and priceScaleId as the main RSI line

        // Overbought line (custom value from settings - Red)
        const rsiOverbought = chart.addSeries(
          LineSeries,
          {
            color: '#EF4444',
            lineWidth: 1,
            lineStyle: LineStyle.Dashed,
            priceScaleId: 'rsi',
          },
          paneIndex
        )
        refLineSeriesRef.current.push(rsiOverbought)

        // Set data immediately - no need for requestAnimationFrame with local data
        rsiOverbought.setData(refLineData.map((d) => ({ ...d, value: settings.rsiOverbought })))

        // 50 line (Neutral - Gray) - skip if it would duplicate one of the custom thresholds
        if (settings.rsiOverbought !== 50 && settings.rsiOversold !== 50) {
          const rsi50 = chart.addSeries(
            LineSeries,
            {
              color: '#6B7280',
              lineWidth: 1,
              lineStyle: LineStyle.Dashed,
              priceScaleId: 'rsi',
            },
            paneIndex
          )
          refLineSeriesRef.current.push(rsi50)

          // Set data immediately
          rsi50.setData(refLineData.map((d) => ({ ...d, value: 50 })))
        }

        // Oversold line (custom value from settings - Green)
        const rsiOversold = chart.addSeries(
          LineSeries,
          {
            color: '#10B981',
            lineWidth: 1,
            lineStyle: LineStyle.Dashed,
            priceScaleId: 'rsi',
          },
          paneIndex
        )
        refLineSeriesRef.current.push(rsiOversold)

        // Set data immediately
        rsiOversold.setData(refLineData.map((d) => ({ ...d, value: settings.rsiOversold })))

        console.log('[RSI] ✅ Added 3 reference lines immediately (pane', paneIndex, ')')
      } catch (error) {
        console.error('[RSI] Error creating reference lines:', error)
      }
    },
    [chart, paneIndex, containerHeight, refLineData, settings.rsiOverbought, settings.rsiOversold]
  )

  // Cleanup reference lines when component unmounts
  const handleRSISeriesDestroyed = useCallback(() => {
    if (!chart) return

    refLineSeriesRef.current.forEach((series) => {
      try {
        chart.removeSeries(series)
      } catch (error) {
        console.error('[RSI] Error removing reference line:', error)
      }
    })
    refLineSeriesRef.current = []
    console.log('[RSI] 🗑️  Cleaned up reference lines')
  }, [chart])

  // Memoize seriesOptions so PHASE 3 effect in use-series can detect changes
  const seriesOptions = useMemo(() => {
    return {
      color: '#9C27B0',
      lineWidth: 2,
      title: `RSI (${settings.rsiPeriod})`,
      priceScaleId: 'rsi',
      // autoscaleInfoProvider suggests 0-100 range to TradingView's autoscaler
      autoscaleInfoProvider: () => ({
        priceRange: {
          minValue: 0,
          maxValue: 100,
        },
      }),
    }
  }, [settings.rsiPeriod])

  // RSI Line (purple) - ONLY useSeries hook for main line
  // This creates the pane and price scale, then adds reference lines in callback
  useSeries({
    seriesType: LineSeries,
    seriesOptions,
    data: rsiData,
    paneIndex,
    onCreated: handleRSISeriesCreated,
    onDestroyed: handleRSISeriesDestroyed,
    // Keep autoscale enabled so the pane always renders,
    // and clamp range via autoscaleInfoProvider (0-100).
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0, // Normal price scale mode
      scaleMargins: { top: 0.02, bottom: 0.02 }, // Tight margins for RSI range
    },
  })

  return (
    <PaneOverlay paneIndex={paneIndex}>
      <IndicatorSettingsPopover
        indicatorName="RSI"
        settings={settings}
        fields={[
          { name: 'rsiPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
          { name: 'rsiOverbought', label: 'Overbought', type: 'number', min: 50, max: 100 },
          { name: 'rsiOversold', label: 'Oversold', type: 'number', min: 0, max: 50 },
        ]}
        onUpdate={(newSettings) => {
          console.log('[RSI] 🔄 Updating settings:', JSON.stringify(newSettings, null, 2))
          updateSettings(newSettings)
        }}
      />
    </PaneOverlay>
  )
}
