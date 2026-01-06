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

import { useMemo, useCallback, useEffect, useRef } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { IPriceLine, ISeriesApi } from 'lightweight-charts'
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

  const priceLinesRef = useRef<IPriceLine[]>([])
  const rsiSeriesRef = useRef<ISeriesApi<any> | null>(null)

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

  const clearPriceLines = useCallback(() => {
    const series = rsiSeriesRef.current
    if (!series) {
      priceLinesRef.current = []
      return
    }

    for (const line of priceLinesRef.current) {
      try {
        series.removePriceLine(line)
      } catch {
        // ignore
      }
    }
    priceLinesRef.current = []
  }, [])

  const applyGuidelinePriceLines = useCallback(() => {
    const series = rsiSeriesRef.current
    if (!series) return

    clearPriceLines()

    const buy = Number(settings.rsiOversold)
    const sell = Number(settings.rsiOverbought)

    // Show strategy guideline levels (derived from momentum settings / preset).
    if (Number.isFinite(buy)) {
      priceLinesRef.current.push(
        series.createPriceLine({
          price: buy,
          color: '#10B981',
          lineWidth: 1,
          lineStyle: LineStyle.Dashed,
          axisLabelVisible: true,
          title: `BUY ${buy}`,
        })
      )
    }

    if (Number.isFinite(sell)) {
      priceLinesRef.current.push(
        series.createPriceLine({
          price: sell,
          color: '#EF4444',
          lineWidth: 1,
          lineStyle: LineStyle.Dashed,
          axisLabelVisible: true,
          title: `SELL ${sell}`,
        })
      )
    }
  }, [clearPriceLines, settings.rsiOverbought, settings.rsiOversold])

  // Callback when RSI main series is created - set pane height and add guideline lines immediately.
  const handleRSISeriesCreated = useCallback(
    (rsiSeries: ISeriesApi<any>) => {
      if (!chart) return

      try {
        rsiSeriesRef.current = rsiSeries

        // 1. Set pane height using paneIndex prop
        const panes = chart.panes()
        const rsiPane = panes[paneIndex]
        const maxHeight = Math.floor(containerHeight * 0.9)
        const desiredHeight = Math.floor(containerHeight * 0.15) // 15% default
        const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
        rsiPane.setHeight(constrainedHeight)

        console.log('[RSI] ✅ Pane configured: height =', constrainedHeight, 'px')

        // 2. Add guideline price lines (BUY/SELL levels) on the RSI series.
        applyGuidelinePriceLines()
      } catch (error) {
        console.error('[RSI] Error creating guideline lines:', error)
      }
    },
    [chart, paneIndex, containerHeight, applyGuidelinePriceLines]
  )

  // Keep guideline lines in sync when thresholds change (e.g. preset applied or user edits).
  useEffect(() => {
    applyGuidelinePriceLines()
  }, [applyGuidelinePriceLines])

  // Cleanup guideline lines when component unmounts
  const handleRSISeriesDestroyed = useCallback(() => {
    clearPriceLines()
    rsiSeriesRef.current = null
    console.log('[RSI] 🗑️  Cleaned up guideline lines')
  }, [clearPriceLines])

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
