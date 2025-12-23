/**
 * Chart Data Processor
 * Converts raw API data to TradingView Lightweight Charts format
 * Handles timeframe filtering and data transformation
 */

import type { TickerHistoricalData } from '@/types/analysis'
import type { ProcessedChartData } from '@/components/analysis/chart/ChartContext'
import type { ChartType, Timeframe } from '@/lib/hooks/use-chart-state'

/**
 * Convert raw ticker data to ProcessedChartData format for TradingView charts
 */
export function processChartData(
  rawData: TickerHistoricalData[],
  chartType: ChartType = 'candlestick',
  timeframe: Timeframe = '1D'
): ProcessedChartData {
  // Filter and sort data based on timeframe
  const filteredData = filterDataByTimeframe(rawData, timeframe)

  // Sort by date (ascending)
  const sortedData = filteredData.sort((a, b) =>
    new Date(a.date).getTime() - new Date(b.date).getTime()
  )

  // Convert to TradingView format
  const candlestickData = sortedData.map(item => ({
    time: new Date(item.date).getTime() / 1000, // TradingView expects Unix timestamp in seconds
    open: Number(item.open) || 0,
    high: Number(item.high) || 0,
    low: Number(item.low) || 0,
    close: Number(item.close) || 0,
  }))

  // Generate volume data with color based on price movement
  const volumeData = sortedData.map((item, index) => {
    // Determine color based on price change (green for up, red for down)
    const isUp = index === 0 ? true : item.close >= sortedData[index - 1].close
    return {
      time: new Date(item.date).getTime() / 1000,
      value: Number(item.volume) || 0,
      color: isUp ? 'rgba(34, 197, 94, 0.5)' : 'rgba(239, 68, 68, 0.5)', // Green/Red with opacity
    }
  })

  return {
    candlestickData,
    volumeData,
  }
}

/**
 * Filter data based on selected timeframe
 */
function filterDataByTimeframe(
  data: TickerHistoricalData[],
  timeframe: Timeframe
): TickerHistoricalData[] {
  if (!data || data.length === 0) return []

  const now = new Date()
  const dataDate = new Date(data[data.length - 1]?.date || now)

  switch (timeframe) {
    case '1D':
      // Last 24 hours from the last data point
      const oneDayAgo = new Date(dataDate.getTime() - 24 * 60 * 60 * 1000)
      return data.filter(item => new Date(item.date) >= oneDayAgo)

    case '1W':
      // Last 7 days from the last data point
      const oneWeekAgo = new Date(dataDate.getTime() - 7 * 24 * 60 * 60 * 1000)
      return data.filter(item => new Date(item.date) >= oneWeekAgo)

    case '1M':
      // Last 30 days from the last data point
      const oneMonthAgo = new Date(dataDate.getTime() - 30 * 24 * 60 * 60 * 1000)
      return data.filter(item => new Date(item.date) >= oneMonthAgo)

    case '3M':
      // Last 90 days from the last data point
      const threeMonthsAgo = new Date(dataDate.getTime() - 90 * 24 * 60 * 60 * 1000)
      return data.filter(item => new Date(item.date) >= threeMonthsAgo)

    case '1Y':
      // Last 365 days from the last data point
      const oneYearAgo = new Date(dataDate.getTime() - 365 * 24 * 60 * 60 * 1000)
      return data.filter(item => new Date(item.date) >= oneYearAgo)

    case 'MAX':
    default:
      // Return all data
      return data
  }
}

export { filterDataByTimeframe }

/**
 * Get theme colors based on system preference
 */
export function getChartTheme(): 'light' | 'dark' {
  if (typeof window !== 'undefined') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'light'
}

/**
 * Get theme color values for TradingView chart
 */
export function getThemeColors(theme?: 'light' | 'dark') {
  // If a theme is provided (from app state), it must override OS preference.
  const isDark = theme ? theme === 'dark' : getChartTheme() === 'dark'

  return {
    backgroundColor: isDark ? '#0a0a0a' : '#ffffff',
    textColor: isDark ? '#d1d5db' : '#374151',
    gridColor: isDark ? '#374151' : '#e5e7eb',
    candlestickUpColor: '#10b981', // Green
    candlestickDownColor: '#ef4444', // Red
    volumeUpColor: 'rgba(34, 197, 94, 0.5)',
    volumeDownColor: 'rgba(239, 68, 68, 0.5)',
  }
}
