/**
 * Trading Calendar Utilities
 * Helper functions for ISX trading day calculations
 * Handles Iraqi weekend (Friday/Saturday) and ISX holidays
 */

/**
 * Calculate calendar days between two dates (inclusive)
 */
export function calculateCalendarDays(fromDate: Date | string, toDate: Date | string): number {
  const from = typeof fromDate === 'string' ? new Date(fromDate) : fromDate
  const to = typeof toDate === 'string' ? new Date(toDate) : toDate

  const diffTime = Math.abs(to.getTime() - from.getTime())
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))

  return diffDays + 1 // +1 to make it inclusive
}

/**
 * Format date range for display
 * Examples:
 *   - "Jan 1 - Jan 31" (same month)
 *   - "Dec 28 - Jan 5" (different months)
 *   - "Jan 1, 2024 - Mar 31, 2025" (different years)
 */
export function formatDateRange(fromDate: Date | string, toDate: Date | string): string {
  const from = typeof fromDate === 'string' ? new Date(fromDate) : fromDate
  const to = typeof toDate === 'string' ? new Date(toDate) : toDate

  const sameYear = from.getFullYear() === to.getFullYear()
  const sameMonth = sameYear && from.getMonth() === to.getMonth()

  if (sameMonth) {
    return `${from.toLocaleDateString('en', { month: 'short', day: 'numeric' })} - ${to.toLocaleDateString('en', { day: 'numeric' })}`
  } else if (sameYear) {
    return `${from.toLocaleDateString('en', { month: 'short', day: 'numeric' })} - ${to.toLocaleDateString('en', { month: 'short', day: 'numeric' })}`
  } else {
    return `${from.toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })} - ${to.toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })}`
  }
}

/**
 * Check if a date is a Friday or Saturday (Iraqi weekend)
 */
export function isWeekend(date: Date | string): boolean {
  const d = typeof date === 'string' ? new Date(date) : date
  const dayOfWeek = d.getDay()
  return dayOfWeek === 5 || dayOfWeek === 6 // Friday or Saturday
}

/**
 * Get weekday name (with Iraqi context)
 */
export function getWeekdayName(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleDateString('en', { weekday: 'long' })
}

/**
 * Calculate trading days vs non-trading days in a range
 */
export interface TradingDaysBreakdown {
  totalDays: number
  tradingDays: number
  nonTradingDays: number
  weekends: number // Fridays + Saturdays
  holidays: number
  percentage: number // Trading days as percentage of total
}

/**
 * Generate a summary of trading days vs non-trading days
 */
export function getTradingDaysSummary(
  fromDate: Date | string,
  toDate: Date | string,
  tradingDaysCount: number
): TradingDaysBreakdown {
  const totalDays = calculateCalendarDays(fromDate, toDate)
  const nonTradingDays = totalDays - tradingDaysCount

  // Estimate weekends (very rough approximation)
  const weeksInRange = Math.floor(totalDays / 7)
  const estimatedWeekends = weeksInRange * 2

  // Holidays = non-trading days minus weekends
  const holidays = Math.max(0, nonTradingDays - estimatedWeekends)

  return {
    totalDays,
    tradingDays: tradingDaysCount,
    nonTradingDays,
    weekends: Math.min(estimatedWeekends, nonTradingDays),
    holidays,
    percentage: Math.round((tradingDaysCount / totalDays) * 100)
  }
}

/**
 * Generate user-friendly date range text with trading days
 * Example: "Jan 1 - Jan 31 (31 days, 18 trading days)"
 */
export function formatDateRangeWithTradingDays(
  fromDate: Date | string,
  toDate: Date | string,
  tradingDaysCount: number
): string {
  const calendarDays = calculateCalendarDays(fromDate, toDate)
  const dateRange = formatDateRange(fromDate, toDate)

  return `${dateRange} (${calendarDays}d, ${tradingDaysCount} trading)`
}

/**
 * Check if date range includes significant non-trading days
 * Returns warning message if > 30% non-trading days
 */
export function validateDateRangeForTrading(
  fromDate: Date | string,
  toDate: Date | string,
  tradingDaysCount: number
): { valid: boolean; warning?: string } {
  const summary = getTradingDaysSummary(fromDate, toDate, tradingDaysCount)

  if (summary.percentage < 70) {
    return {
      valid: false,
      warning: `⚠️ Selected range includes ${summary.nonTradingDays} non-trading days (${summary.weekends} weekends, ${summary.holidays} holidays)`
    }
  }

  if (summary.percentage < 85) {
    return {
      valid: true,
      warning: `Note: Range includes ${summary.nonTradingDays} non-trading days`
    }
  }

  return { valid: true }
}

/**
 * Get month name from date
 */
export function getMonthName(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleDateString('en', { month: 'long' })
}

/**
 * Get year from date
 */
export function getYear(date: Date | string): number {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.getFullYear()
}

/**
 * Format date for display (e.g., "Sunday, Jan 5, 2025")
 */
export function formatLongDate(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleDateString('en', {
    weekday: 'long',
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })
}
