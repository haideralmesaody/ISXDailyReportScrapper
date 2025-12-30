/**
 * Date Helper Utilities for Trading Session Display
 * Pure, testable functions following CLAUDE.md standards
 */

/**
 * Format trading date to human-readable format with day of week
 *
 * @param date - ISO date string (YYYY-MM-DD)
 * @returns Formatted string like "Monday, June 5, 2025"
 */
export function formatTradingDate(date: string): string {
  const d = new Date(date + 'T00:00:00Z') // Force UTC midnight to avoid timezone issues

  const dayOfWeek = d.toLocaleDateString('en-US', { weekday: 'long', timeZone: 'UTC' })
  const month = d.toLocaleDateString('en-US', { month: 'long', timeZone: 'UTC' })
  const day = d.getUTCDate()
  const year = d.getUTCFullYear()

  return `${dayOfWeek}, ${month} ${day}, ${year}`
}

/**
 * Calculate relative time from date to today
 *
 * @param date - ISO date string (YYYY-MM-DD)
 * @returns Human-readable relative time ("Today", "Yesterday", "5 days ago", "In 2 days")
 */
export function getDaysAgo(date: string): string {
  const targetDate = new Date(date + 'T00:00:00Z')
  const today = new Date()
  today.setHours(0, 0, 0, 0) // Reset to midnight for accurate day comparison

  const diffMs = today.getTime() - targetDate.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays === -1) return 'Tomorrow'
  if (diffDays > 0) return `${diffDays} days ago`
  return `In ${Math.abs(diffDays)} days` // Future date
}

/**
 * Get data freshness indicator based on how old the date is
 *
 * @param date - ISO date string (YYYY-MM-DD)
 * @returns Object with text and variant for badge styling
 */
export function getDataFreshnessIndicator(date: string): {
  text: string
  variant: 'success' | 'warning' | 'muted'
} {
  const targetDate = new Date(date + 'T00:00:00Z')
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const diffMs = today.getTime() - targetDate.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) {
    return { text: 'Latest Data', variant: 'success' }
  } else if (diffDays === 1) {
    return { text: 'Fresh', variant: 'success' }
  } else if (diffDays <= 7) {
    return { text: 'Recent', variant: 'warning' }
  } else {
    return { text: 'Historical', variant: 'muted' }
  }
}

/**
 * Get next trading date from sorted array
 *
 * @param current - Current ISO date string
 * @param dates - Sorted array of trading dates (earliest first)
 * @returns Next trading date or null if at end
 */
export function getNextTradingDate(current: string, dates: string[]): string | null {
  const currentIndex = dates.indexOf(current)

  if (currentIndex === -1 || currentIndex === dates.length - 1) {
    return null // Not found or already at latest
  }

  return dates[currentIndex + 1] ?? null
}

/**
 * Get previous trading date from sorted array
 *
 * @param current - Current ISO date string
 * @param dates - Sorted array of trading dates (earliest first)
 * @returns Previous trading date or null if at start
 */
export function getPrevTradingDate(current: string, dates: string[]): string | null {
  const currentIndex = dates.indexOf(current)

  if (currentIndex === -1 || currentIndex === 0) {
    return null // Not found or already at earliest
  }

  return dates[currentIndex - 1] ?? null
}

/**
 * Get quick date shortcut
 *
 * @param type - Shortcut type ('latest', 'yesterday', 'lastWeek')
 * @param dates - Sorted array of trading dates (earliest first)
 * @returns Date string or null if not available
 */
export function getQuickDateShortcut(
  type: 'latest' | 'yesterday' | 'lastWeek',
  dates: string[]
): string | null {
  if (dates.length === 0) return null

  const latest = dates.at(-1)
  if (!latest) return null

  if (type === 'latest') {
    return latest
  }

  if (type === 'yesterday') {
    // Get yesterday's date
    const yesterday = new Date()
    yesterday.setDate(yesterday.getDate() - 1)
    const yesterdayStr = yesterday.toISOString().split('T')[0]
    if (!yesterdayStr) return null

    // Find closest trading date on or before yesterday
    const matchingDate = dates.slice().reverse().find(d => d <= yesterdayStr)

    return matchingDate || null
  }

  if (type === 'lastWeek') {
    // Get date 7 days ago
    const lastWeek = new Date()
    lastWeek.setDate(lastWeek.getDate() - 7)
    const lastWeekStr = lastWeek.toISOString().split('T')[0]
    if (!lastWeekStr) return null

    // Find closest trading date on or before last week
    const matchingDate = dates.slice().reverse().find(d => d <= lastWeekStr)

    return matchingDate || null
  }

  return null
}

/**
 * Check if date is the latest trading date
 *
 * @param date - ISO date string to check
 * @param dates - Sorted array of trading dates
 * @returns True if date is the latest trading date
 */
export function isLatestTradingDate(date: string, dates: string[]): boolean {
  if (dates.length === 0) return false
  return dates.at(-1) === date
}

/**
 * Convert Date object to ISO string (YYYY-MM-DD)
 *
 * @param date - Date object
 * @returns ISO date string
 */
export function formatDateToISO(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Parse ISO date string to Date object
 *
 * @param dateStr - ISO date string (YYYY-MM-DD)
 * @returns Date object set to midnight UTC
 */
export function parseISODate(dateStr: string): Date {
  return new Date(dateStr + 'T00:00:00Z')
}

/**
 * Get all trading dates for a specific month
 *
 * @param year - Full year (e.g., 2025)
 * @param month - Month (0-11, where 0 = January)
 * @param tradingDates - Array of trading dates
 * @returns Array of trading dates in that month
 */
export function getTradingDatesForMonth(
  year: number,
  month: number,
  tradingDates: string[]
): string[] {
  const monthStr = `${year}-${String(month + 1).padStart(2, '0')}`
  return tradingDates.filter(d => d.startsWith(monthStr))
}

/**
 * Check if a specific date is a trading date
 *
 * @param date - Date object to check
 * @param tradingDates - Array of trading dates
 * @returns True if date is a trading date
 */
export function isTradingDate(date: Date, tradingDates: string[]): boolean {
  const dateStr = formatDateToISO(date)
  return tradingDates.includes(dateStr)
}
