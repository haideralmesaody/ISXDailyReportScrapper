/**
 * ChartCore Component
 * Following CLAUDE.md React patterns with proper hydration handling
 * Uses ChartContainer with TradingView Lightweight Charts for professional visualization
 */

'use client'

import React, { useEffect, useMemo, useState } from 'react'
import { useHydration } from '@/lib/hooks/use-hydration'
import { Loader2 } from 'lucide-react'
import { ChartContainer } from './ChartContainer'
import { filterDataByTimeframe, processChartData, getThemeColors } from '@/lib/utils/chart-data-processor'
import { calculateRSI } from '@/lib/indicators/rsi'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import type { IChartApi, SeriesMarker, Time } from 'lightweight-charts'
import type { ChartType, Timeframe } from '@/lib/hooks/use-chart-state'
import type { TickerHistoricalData } from '@/types/analysis'
import { useTheme } from 'next-themes'

interface ChartCoreProps {
  ticker: string
  data: any[]
  isLoading: boolean
  onChartReady?: (chart: IChartApi | null) => void  // Callback when chart is ready
  chartType?: ChartType  // Chart type (candlestick/line/area/bar)
  timeframe?: Timeframe  // Timeframe (1D/1W/1M/3M/1Y/MAX)
  indicators?: any  // Indicator settings from state
  indicatorActivationOrder?: Record<string, number>
  markers?: SeriesMarker<Time>[]
  watermarkText?: string
}

export function ChartCore({
  ticker,
  data,
  isLoading,
  onChartReady,
  chartType = 'candlestick',  // Default to candlestick
  timeframe = '3M',  // Default to 3 months
  indicators,  // Use indicator settings from state
  indicatorActivationOrder = {},
  markers,
  watermarkText,
}: ChartCoreProps) {
  const isHydrated = useHydration()
  const { theme, resolvedTheme } = useTheme()
  const { settings: momentumSettings } = useMomentumSettings()
  const [htmlIsDark, setHtmlIsDark] = useState(false)

  useEffect(() => {
    if (!isHydrated) return
    const root = document.documentElement
    const update = () => setHtmlIsDark(root.classList.contains('dark'))
    update()
    const observer = new MutationObserver(update)
    observer.observe(root, { attributes: true, attributeFilter: ['class'] })
    return () => observer.disconnect()
  }, [isHydrated])

  // Process raw data into TradingView format
  const processedData = useMemo(() => {
    if (!data || data.length === 0) {
      return {
        candlestickData: [],
        volumeData: []
      }
    }
    return processChartData(data, chartType, timeframe)
  }, [data, chartType, timeframe])

  const filteredHistoricalData = useMemo(() => {
    return filterDataByTimeframe(data as TickerHistoricalData[], timeframe)
  }, [data, timeframe])

  const indicatorData = useMemo(() => {
    const typed = data as TickerHistoricalData[]
    if (!typed || typed.length === 0) return {}

    const sorted = [...typed].sort(
      (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
    )

    const candles = sorted.map((item) => ({
      time: new Date(item.date).getTime() / 1000,
      open: Number(item.open) || 0,
      high: Number(item.high) || 0,
      low: Number(item.low) || 0,
      close: Number(item.close) || 0,
    }))

    const { rsi } = calculateRSI(candles, { period: momentumSettings.rsiPeriod })

    const rsiMap = new Map<string, number | null>()
    for (let i = 0; i < sorted.length; i++) {
      rsiMap.set(sorted[i].date, rsi[i] ?? null)
    }

    return { rsiData: rsiMap }
  }, [data, momentumSettings.rsiPeriod])

  // Get theme colors
  const themeColors = useMemo(() => {
    // Prefer actual DOM theme class to avoid any mismatch between app theme UI and next-themes values.
    // Fallback to next-themes only if DOM is unavailable (shouldn't happen after hydration).
    const domTheme: 'dark' | 'light' = htmlIsDark ? 'dark' : 'light'

    const effectiveTheme = theme === 'system' ? resolvedTheme : theme
    const nextTheme =
      effectiveTheme === 'dark' ? 'dark' : effectiveTheme === 'light' ? 'light' : undefined

    return getThemeColors(nextTheme ?? domTheme)
  }, [htmlIsDark, theme, resolvedTheme])

  // Show initial loading state during hydration (before ANY chart is rendered)
  if (!isHydrated) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          <p className="text-sm text-muted-foreground">Initializing chart...</p>
        </div>
      </div>
    )
  }

  // Render the actual chart container
  return (
    <div className="w-full h-full relative">
      <ChartContainer
        ticker={ticker}
        chartData={processedData}
        historicalData={filteredHistoricalData}
        indicators={indicators}
        indicatorActivationOrder={indicatorActivationOrder}
        theme={themeColors}
        indicatorData={indicatorData}
        onChartReady={onChartReady}
        chartType={chartType}
        timeframe={timeframe}
        markers={markers}
        watermarkText={watermarkText}
      />

      {/* Loading overlay - shows on top of chart when loading new ticker data */}
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm z-50">
          <div className="flex flex-col items-center gap-3">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <p className="text-sm text-muted-foreground">Loading data...</p>
          </div>
        </div>
      )}
    </div>
  )
}
