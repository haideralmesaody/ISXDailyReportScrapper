/**
 * Shared constants for ISX Pulse application
 * Centralized configuration values to ensure consistency across the app
 */

// Release and timeline constants
export const EXPECTED_RELEASE = 'Q2 2025' as const

// Feature availability dates
export const FEATURES_TIMELINE = {
  LIQUIDITY: EXPECTED_RELEASE,
  REPORTS: EXPECTED_RELEASE,
  ADVANCED_FEATURES: 'Q3 2025',
} as const

// Application metadata
export const APP_NAME = 'ISX Pulse' as const
export const APP_TAGLINE = 'The Heartbeat of Iraqi Markets' as const

// Feature flags (for future use)
export const FEATURE_FLAGS = {
  ENABLE_LIQUIDITY: false,
  ENABLE_REPORTS: false,
  ENABLE_EXPORT: false,
  ENABLE_NOTIFICATIONS: false,
} as const

// Operation date defaults - Single Source of Truth
export const OPERATION_DATE_DEFAULTS = {
  // Default from date: January 1, 2025
  FROM_DATE: '2025-01-01',
  // Default to date: today (dynamic)
  TO_DATE: () => new Date(),
  // Helper to get formatted date string in local timezone
  getFromDate: () => new Date(OPERATION_DATE_DEFAULTS.FROM_DATE),
  getToDate: () => {
    const now = new Date()
    // Ensure we get local date, not UTC
    const localDate = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    return localDate
  },
  getFromDateString: () => OPERATION_DATE_DEFAULTS.FROM_DATE,
  getToDateString: () => {
    const now = new Date()
    // Format as YYYY-MM-DD in local timezone
    const year = now.getFullYear()
    const month = String(now.getMonth() + 1).padStart(2, '0')
    const day = String(now.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  },
} as const

// Date validation constants
export const DATE_VALIDATION = {
  MIN_DATE: '2020-01-01',
  MAX_DAYS_RANGE: 365,
} as const