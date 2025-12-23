/**
 * Date filtering utilities for chart time range selection
 * Provides functions to filter historical data based on selected time periods
 */

export type TimeRange = '1W' | '1M' | '3M' | '6M' | '1Y' | 'YTD' | 'ALL'

/**
 * Get the start date for a given time range
 */
export function getStartDateForRange(range: TimeRange, endDate: Date = new Date()): Date | null {
  const startDate = new Date(endDate)
  
  switch (range) {
    case '1W':
      startDate.setDate(startDate.getDate() - 7)
      break
    case '1M':
      startDate.setMonth(startDate.getMonth() - 1)
      break
    case '3M':
      startDate.setMonth(startDate.getMonth() - 3)
      break
    case '6M':
      startDate.setMonth(startDate.getMonth() - 6)
      break
    case '1Y':
      startDate.setFullYear(startDate.getFullYear() - 1)
      break
    case 'YTD':
      // Set to January 1st of current year
      startDate.setMonth(0, 1)
      startDate.setHours(0, 0, 0, 0)
      break
    case 'ALL':
      return null // No filtering
    default:
      return null
  }
  
  return startDate
}

/**
 * Filter data array based on selected time range
 */
export function filterDataByTimeRange<T extends { date: string }>(
  data: T[],
  range: TimeRange
): T[] {
  if (range === 'ALL' || data.length === 0) {
    return data
  }
  
  // Get the last date from the data
  const lastDateStr = data[data.length - 1].date
  const endDate = new Date(lastDateStr)
  
  const startDate = getStartDateForRange(range, endDate)
  if (!startDate) {
    return data
  }
  
  // Convert start date to YYYY-MM-DD format for comparison
  const startDateStr = formatDateToYYYYMMDD(startDate)
  
  // Filter data to include only dates >= startDate
  return data.filter(item => item.date >= startDateStr)
}

/**
 * Format date to YYYY-MM-DD string
 */
export function formatDateToYYYYMMDD(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Get human-readable label for time range
 */
export function getTimeRangeLabel(range: TimeRange): string {
  switch (range) {
    case '1W':
      return '1 Week'
    case '1M':
      return '1 Month'
    case '3M':
      return '3 Months'
    case '6M':
      return '6 Months'
    case '1Y':
      return '1 Year'
    case 'YTD':
      return 'Year to Date'
    case 'ALL':
      return 'All Time'
    default:
      return range
  }
}

/**
 * Get short label for time range (for mobile)
 */
export function getTimeRangeShortLabel(range: TimeRange): string {
  switch (range) {
    case 'YTD':
      return 'YTD'
    case 'ALL':
      return 'All'
    default:
      return range
  }
}

/**
 * Calculate the appropriate time range based on available data
 */
export function getDefaultTimeRange<T extends { date: string }>(data: T[]): TimeRange {
  if (data.length === 0) return 'ALL'
  
  const firstDate = new Date(data[0].date)
  const lastDate = new Date(data[data.length - 1].date)
  const daysDiff = Math.ceil((lastDate.getTime() - firstDate.getTime()) / (1000 * 60 * 60 * 24))
  
  // Select appropriate default based on data range
  if (daysDiff <= 7) return '1W'
  if (daysDiff <= 30) return '1M'
  if (daysDiff <= 90) return '3M'
  if (daysDiff <= 180) return '6M'
  if (daysDiff <= 365) return '1Y'
  return 'ALL'
}