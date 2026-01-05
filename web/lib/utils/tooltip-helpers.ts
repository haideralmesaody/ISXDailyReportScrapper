/**
 * Tooltip Helper Functions for Chart Display
 * Provides formatting and calculation utilities for enhanced tooltips
 */

/**
 * Format number as IQD currency
 */
export function formatIQD(value: number | null | undefined): string {
  if (value === null || value === undefined) return 'N/A'
  return new Intl.NumberFormat('en-IQ', {
    style: 'currency',
    currency: 'IQD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

/**
 * Format large numbers with thousands separator
 */
export function formatVolume(value: number | null | undefined): string {
  if (value === null || value === undefined) return 'N/A'
  return new Intl.NumberFormat('en-US', {
    maximumFractionDigits: 0,
  }).format(value)
}

/**
 * Format number with specified decimal places
 */
export function formatNumber(value: number | null | undefined, decimals: number = 2): string {
  if (value === null || value === undefined) return 'N/A'
  return value.toFixed(decimals)
}

/**
 * Calculate percentage change
 */
export function calculatePercentageChange(current: number, previous: number): number {
  if (!previous || previous === 0) return 0
  return ((current - previous) / previous) * 100
}

/**
 * Format percentage with sign and color
 */
export function formatPercentage(value: number | null | undefined): {
  text: string
  color: string
  isPositive: boolean
} {
  if (value === null || value === undefined) {
    return { text: 'N/A', color: 'text-muted-foreground', isPositive: false }
  }

  const formatted = `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
  const color = value > 0 ? 'text-green-600' : value < 0 ? 'text-red-600' : 'text-gray-500'
  
  return {
    text: formatted,
    color,
    isPositive: value > 0
  }
}

/**
 * Format date for display
 */
export function formatDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(d)
}

/**
 * Format time for display
 */
export function formatTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  }).format(d)
}

/**
 * Calculate trading value in IQD
 */
export function calculateTradingValue(price: number, volume: number): number {
  return price * volume
}

/**
 * Comparison data structure
 */
export interface ComparisonData {
  price: number
  change: number
  changePercent: number
  volume: number
  date: string
}

/**
 * Calculate comparison metrics
 */
export function calculateComparison(
  current: { close: number; volume: number; date: string },
  previous: { close: number; volume: number; date: string } | undefined
): ComparisonData | null {
  if (!previous) return null
  
  const change = current.close - previous.close
  const changePercent = calculatePercentageChange(current.close, previous.close)
  
  return {
    price: previous.close,
    change,
    changePercent,
    volume: previous.volume,
    date: previous.date
  }
}

/**
 * Calculate simple moving average
 */
export function calculateSMA(values: number[], period: number): number | null {
  if (values.length < period) return null
  
  const sum = values.slice(-period).reduce((acc, val) => acc + val, 0)
  return sum / period
}

/**
 * Get trading days ago
 */
export function getTradingDaysAgo(data: any[], daysAgo: number): any | undefined {
  const index = data.length - 1 - daysAgo
  return index >= 0 ? data[index] : undefined
}

/**
 * Indicator value in tooltip
 */
export interface TooltipIndicatorValue {
  value: number | null
  color: string
  displayName: string
  formatted: string
}

/**
 * Enhanced tooltip data structure
 */
export interface EnhancedTooltipData {
  // Basic OHLCV
  date: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  
  // Additional metrics
  value?: number // Trading value in IQD
  trades?: number // Number of trades
  
  // Calculated metrics
  change: number // Daily change
  changePercent: number // Daily change %
  volumeMA20?: number // 20-day volume average
  priceMA20?: number // 20-day price average (legacy, kept for compatibility)
  
  // Dynamic indicators
  indicators?: Map<string, TooltipIndicatorValue>
  activeIndicatorIds?: string[] // For display ordering
  
  // Comparisons
  vsYesterday?: ComparisonData | null
  vsWeekAgo?: ComparisonData | null
  vsMonthAgo?: ComparisonData | null
}

/**
 * Build enhanced tooltip data
 */
export function buildEnhancedTooltipData(
  currentData: any,
  historicalData: any[],
  currentIndex: number
): EnhancedTooltipData {
  // Get previous day data
  const yesterday = currentIndex > 0 ? historicalData[currentIndex - 1] : null
  
  // Calculate daily change
  const change = yesterday ? currentData.close - yesterday.close : 0
  const changePercent = yesterday ? calculatePercentageChange(currentData.close, yesterday.close) : 0
  
  // Calculate moving averages
  const closePrices = historicalData.slice(Math.max(0, currentIndex - 19), currentIndex + 1).map(d => d.close)
  const volumes = historicalData.slice(Math.max(0, currentIndex - 19), currentIndex + 1).map(d => d.volume)
  
  const priceMA20 = calculateSMA(closePrices, 20)
  const volumeMA20 = calculateSMA(volumes, 20)
  
  // Get comparison data
  const weekAgoData = currentIndex >= 5 ? historicalData[currentIndex - 5] : undefined
  const monthAgoData = currentIndex >= 22 ? historicalData[currentIndex - 22] : undefined
  
  return {
    date: currentData.date,
    open: currentData.open,
    high: currentData.high,
    low: currentData.low,
    close: currentData.close,
    volume: currentData.volume,
    value: currentData.value || calculateTradingValue(currentData.close, currentData.volume),
    trades: currentData.trades,
    change,
    changePercent,
    priceMA20: priceMA20,
    volumeMA20: volumeMA20,
    vsYesterday: yesterday ? calculateComparison(currentData, yesterday) : null,
    vsWeekAgo: weekAgoData ? calculateComparison(currentData, weekAgoData) : null,
    vsMonthAgo: monthAgoData ? calculateComparison(currentData, monthAgoData) : null,
  }
}