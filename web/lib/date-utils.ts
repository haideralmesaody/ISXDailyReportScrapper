/**
 * Date utility functions for intelligent date handling
 */

import { OPERATION_DATE_DEFAULTS, DATE_VALIDATION } from './constants'

/**
 * Validates and updates dates intelligently
 * - If "to" date is in the past, updates it to today
 * - Preserves "from" date as user preference
 * - Ensures dates are within valid ranges
 */
export function validateAndUpdateDates(savedDates: { from: string; to: string }) {
  const today = new Date()
  today.setHours(0, 0, 0, 0) // Reset to start of day for accurate comparison
  
  let fromDate: Date
  let toDate: Date
  
  // Parse saved dates
  try {
    fromDate = new Date(savedDates.from)
    toDate = new Date(savedDates.to)
  } catch {
    // If parsing fails, return defaults
    return {
      from: OPERATION_DATE_DEFAULTS.getFromDate(),
      to: OPERATION_DATE_DEFAULTS.getToDate()
    }
  }
  
  // Check if dates are valid
  if (isNaN(fromDate.getTime()) || isNaN(toDate.getTime())) {
    return {
      from: OPERATION_DATE_DEFAULTS.getFromDate(),
      to: OPERATION_DATE_DEFAULTS.getToDate()
    }
  }
  
  // Reset times for accurate date comparison
  fromDate.setHours(0, 0, 0, 0)
  toDate.setHours(0, 0, 0, 0)
  
  // Smart update: If "to" date is in the past, update to today
  if (toDate < today) {
    toDate = new Date(today)
  }
  
  // Ensure "from" date isn't after "to" date
  if (fromDate > toDate) {
    // If from date is in the future, reset both to defaults
    if (fromDate > today) {
      return {
        from: OPERATION_DATE_DEFAULTS.getFromDate(),
        to: OPERATION_DATE_DEFAULTS.getToDate()
      }
    }
    // Otherwise, adjust from date to be reasonable
    fromDate = new Date(toDate)
    fromDate.setDate(fromDate.getDate() - 30) // Default to 30 days back
  }
  
  // Ensure dates are within valid range
  const minDate = new Date(DATE_VALIDATION.MIN_DATE)
  if (fromDate < minDate) {
    fromDate = minDate
  }
  
  // Check date range doesn't exceed maximum
  const daysDiff = Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24))
  if (daysDiff > DATE_VALIDATION.MAX_DAYS_RANGE) {
    // Adjust from date to be within range
    fromDate = new Date(toDate)
    fromDate.setDate(fromDate.getDate() - DATE_VALIDATION.MAX_DAYS_RANGE)
  }
  
  return { from: fromDate, to: toDate }
}

/**
 * Checks if a date is today
 */
export function isToday(date: Date): boolean {
  const today = new Date()
  return date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
}

/**
 * Formats a date to YYYY-MM-DD string
 */
export function formatDateString(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Gets a user-friendly date range description
 */
export function getDateRangeDescription(from: Date, to: Date): string {
  const daysDiff = Math.ceil((to.getTime() - from.getTime()) / (1000 * 60 * 60 * 24)) + 1
  
  if (daysDiff === 1 && isToday(to)) {
    return 'Today only'
  }
  
  if (daysDiff === 7 && isToday(to)) {
    return 'Last 7 days'
  }
  
  if (daysDiff === 30 && isToday(to)) {
    return 'Last 30 days'
  }
  
  const fromStr = from.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  const toStr = to.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  
  return `${fromStr} - ${toStr} (${daysDiff} days)`
}