/**
 * ChartContext - Provides chart instance and data to all indicator components
 * Following TradingView official React patterns for context-based architecture
 *
 * Purpose:
 * - Share IChartApi instance across all series components
 * - Provide processed chart data to indicators
 * - Distribute theme colors for consistent styling
 * - Track chart ready state for safe component mounting
 */

'use client'

import { createContext, useContext, type ReactNode, type RefObject } from 'react'
import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts'
import type { ChartType, Timeframe } from '@/lib/hooks/use-chart-state'

// Theme colors interface - matches LightweightStockChart
export interface ThemeColors {
  backgroundColor: string
  textColor: string
  gridColor: string
  candlestickUpColor: string
  candlestickDownColor: string
  volumeUpColor: string
  volumeDownColor: string
}

// Processed chart data structure
export interface ProcessedChartData {
  candlestickData: Array<{
    time: Time
    open: number
    high: number
    low: number
    close: number
  }>
  volumeData: Array<{
    time: Time
    value: number
    color: string
  }>
}

// Indicator data interface for API-based indicators
export interface IndicatorData {
  rsiData?: Map<string, number | null>  // date -> RSI value mapping
  rsiLoading?: boolean
}

// Context value interface
export interface ChartContextValue {
  chart: IChartApi | null
  candlestickSeries: ISeriesApi<'Candlestick'> | null
  containerRef: RefObject<HTMLDivElement>
  chartData: ProcessedChartData
  theme: ThemeColors
  isReady: boolean
  containerHeight: number
  chartType: ChartType  // Current chart type (candlestick/line/area/bar)
  timeframe: Timeframe  // Current timeframe (1D/1W/1M/3M/1Y/MAX)
  indicatorData: IndicatorData  // API-calculated indicator data
}

// Create context with null default (will throw if used without provider)
const ChartContext = createContext<ChartContextValue | null>(null)

// Provider component
interface ChartProviderProps {
  children: ReactNode
  value: ChartContextValue
}

export function ChartProvider({ children, value }: ChartProviderProps) {
  return <ChartContext.Provider value={value}>{children}</ChartContext.Provider>
}

// Hook to access chart context
export function useChart(): ChartContextValue {
  const context = useContext(ChartContext)

  if (!context) {
    throw new Error(
      'useChart must be used within ChartProvider. ' +
      'Ensure your component is wrapped in <ChartProvider>.'
    )
  }

  return context
}
