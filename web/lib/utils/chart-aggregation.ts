/**
 * Chart Data Aggregation Utilities
 *
 * Provides functions to aggregate daily candlestick data into weekly/monthly periods
 * and filter data by date ranges for ISX trading calendar (Sunday-Thursday).
 *
 * Following CLAUDE.md pure function patterns - no side effects, fully testable.
 */

import type { CandlestickData, HistogramData } from 'lightweight-charts'

export type Timeframe = '1D' | '1W' | '1M' | '3M' | '1Y' | 'MAX'

/**
 * Aggregate daily candles into weekly candles
 * ISX trading week: Sunday (0) - Thursday (4)
 *
 * @param dailyData - Array of daily candlestick data sorted chronologically
 * @returns Array of weekly candlestick data
 */
export function aggregateToWeekly(dailyData: CandlestickData[]): CandlestickData[] {
  if (dailyData.length === 0) return []

  const weeklyCandles: CandlestickData[] = []
  let currentWeek: CandlestickData[] = []

  dailyData.forEach((candle, i) => {
    const date = new Date(candle.time)
    const dayOfWeek = date.getDay()

    currentWeek.push(candle)

    // End of week: Thursday (4) or last candle or next day is Sunday/Friday/Saturday
    const isThursday = dayOfWeek === 4
    const isLastCandle = i === dailyData.length - 1

    let isEndOfWeek = isThursday || isLastCandle

    // Check if next candle starts a new week
    if (!isEndOfWeek && i < dailyData.length - 1) {
      const nextDate = new Date(dailyData[i + 1].time)
      const nextDayOfWeek = nextDate.getDay()
      // If next day is Sunday (0) or jumps forward, end current week
      if (nextDayOfWeek === 0 || nextDayOfWeek < dayOfWeek) {
        isEndOfWeek = true
      }
    }

    if (isEndOfWeek && currentWeek.length > 0) {
      weeklyCandles.push({
        time: currentWeek[currentWeek.length - 1].time, // Use last day of week
        open: currentWeek[0].open,
        high: Math.max(...currentWeek.map(c => c.high)),
        low: Math.min(...currentWeek.map(c => c.low)),
        close: currentWeek[currentWeek.length - 1].close
      })
      currentWeek = []
    }
  })

  return weeklyCandles
}

/**
 * Aggregate daily candles into monthly candles
 *
 * @param dailyData - Array of daily candlestick data sorted chronologically
 * @returns Array of monthly candlestick data
 */
export function aggregateToMonthly(dailyData: CandlestickData[]): CandlestickData[] {
  if (dailyData.length === 0) return []

  const monthlyCandles: CandlestickData[] = []
  let currentMonth: CandlestickData[] = []
  let lastMonthKey = ''

  dailyData.forEach((candle, i) => {
    const date = new Date(candle.time)
    const monthKey = `${date.getFullYear()}-${date.getMonth()}`

    // New month started
    if (monthKey !== lastMonthKey && currentMonth.length > 0) {
      monthlyCandles.push({
        time: currentMonth[currentMonth.length - 1].time, // Last day of month
        open: currentMonth[0].open,
        high: Math.max(...currentMonth.map(c => c.high)),
        low: Math.min(...currentMonth.map(c => c.low)),
        close: currentMonth[currentMonth.length - 1].close
      })
      currentMonth = []
    }

    currentMonth.push(candle)
    lastMonthKey = monthKey
  })

  // Don't forget last month
  if (currentMonth.length > 0) {
    monthlyCandles.push({
      time: currentMonth[currentMonth.length - 1].time,
      open: currentMonth[0].open,
      high: Math.max(...currentMonth.map(c => c.high)),
      low: Math.min(...currentMonth.map(c => c.low)),
      close: currentMonth[currentMonth.length - 1].close
    })
  }

  return monthlyCandles
}

/**
 * Filter data to show only last N days
 *
 * @param data - Array of candlestick data
 * @param days - Number of days to keep
 * @returns Filtered array
 */
export function filterToLastDays(data: CandlestickData[], days: number): CandlestickData[] {
  if (data.length === 0) return []

  const lastDate = new Date(data[data.length - 1].time)
  const cutoffDate = new Date(lastDate)
  cutoffDate.setDate(cutoffDate.getDate() - days)

  return data.filter(candle => {
    const candleDate = new Date(candle.time)
    return candleDate >= cutoffDate
  })
}

/**
 * Filter data to show only last N months
 *
 * @param data - Array of candlestick data
 * @param months - Number of months to keep
 * @returns Filtered array
 */
export function filterToLastMonths(data: CandlestickData[], months: number): CandlestickData[] {
  if (data.length === 0) return []

  const lastDate = new Date(data[data.length - 1].time)
  const cutoffDate = new Date(lastDate)
  cutoffDate.setMonth(cutoffDate.getMonth() - months)

  return data.filter(candle => {
    const candleDate = new Date(candle.time)
    return candleDate >= cutoffDate
  })
}

/**
 * Filter data to show only last N years
 *
 * @param data - Array of candlestick data
 * @param years - Number of years to keep
 * @returns Filtered array
 */
export function filterToLastYears(data: CandlestickData[], years: number): CandlestickData[] {
  if (data.length === 0) return []

  const lastDate = new Date(data[data.length - 1].time)
  const cutoffDate = new Date(lastDate)
  cutoffDate.setFullYear(cutoffDate.getFullYear() - years)

  return data.filter(candle => {
    const candleDate = new Date(candle.time)
    return candleDate >= cutoffDate
  })
}

/**
 * Apply timeframe transformation to candlestick data
 * All timeframes are DATE RANGE FILTERS (not aggregations)
 *
 * @param data - Array of daily candlestick data
 * @param timeframe - Timeframe to apply (1D/1W/1M/3M/1Y/MAX)
 * @returns Transformed data array
 */
export function applyTimeframe(
  data: CandlestickData[],
  timeframe: Timeframe
): CandlestickData[] {
  switch (timeframe) {
    case '1D':
      // Last 1 trading day only
      return filterToLastDays(data, 1)
    case '1W':
      // Last 7 days of daily data (NOT weekly aggregation)
      return filterToLastDays(data, 7)
    case '1M':
      // Last 30 days of daily data (NOT monthly aggregation)
      return filterToLastDays(data, 30)
    case '3M':
      // Last 3 months of daily data
      return filterToLastMonths(data, 3)
    case '1Y':
      // Last 1 year of daily data
      return filterToLastYears(data, 1)
    case 'MAX':
      // All available data
      return data
    default:
      return data
  }
}

/**
 * Filter volume data to match timeframe
 * Since all timeframes are now DATE RANGE FILTERS (not aggregations),
 * we simply filter volume to match the filtered candle data.
 *
 * @param volumeData - Array of daily volume histogram data
 * @param candleData - Corresponding candlestick data (for matching time points)
 * @param timeframe - Timeframe being applied
 * @returns Filtered volume data
 */
export function aggregateVolumeData(
  volumeData: HistogramData[],
  candleData: CandlestickData[],
  timeframe: Timeframe
): HistogramData[] {
  // All timeframes now use filtering (no aggregation)
  // Simply filter volume data to match the candle data time points
  const candleTimes = new Set(candleData.map(c => c.time))
  return volumeData.filter(v => candleTimes.has(v.time))
}
