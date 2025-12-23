/**
 * Chart Helper Utilities
 * Following CLAUDE.md type safety patterns
 * Type-safe utility functions for chart operations
 */

// Price formatting
export function formatPrice(price: number): string {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(price)
}

// Volume formatting with compact notation
export function formatVolume(volume: number): string {
  if (volume >= 1_000_000) {
    return new Intl.NumberFormat('en-US', {
      notation: 'compact',
      maximumFractionDigits: 1
    }).format(volume)
  }
  return new Intl.NumberFormat('en-US').format(volume)
}

// Date formatting for tooltip
export function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) {
      return dateString // Return original if invalid
    }

    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      weekday: 'short'
    }).format(date)
  } catch {
    return dateString
  }
}

// Calculate percentage change
export function calculatePercentageChange(oldValue: number, newValue: number): number {
  if (oldValue === 0) return 0
  return ((newValue - oldValue) / oldValue) * 100
}