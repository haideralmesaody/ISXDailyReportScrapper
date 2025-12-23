/**
 * Date utility functions for license management
 * Provides consistent date calculations across the application
 */

export const MS_PER_DAY = 1000 * 60 * 60 * 24

/**
 * Calculate days remaining until expiry date
 * @param expiryDate - ISO date string or undefined
 * @returns Number of days remaining (0 if expired or invalid)
 */
export function calculateDaysLeft(expiryDate?: string): number {
  if (!expiryDate) return 0
  
  try {
    const msLeft = new Date(expiryDate).getTime() - Date.now()
    return Math.max(0, Math.floor(msLeft / MS_PER_DAY))
  } catch {
    return 0
  }
}

/**
 * Format expiration date into human-readable message
 * @param expiryDate - ISO date string or undefined
 * @param isHydrated - Whether client-side hydration is complete
 * @returns Formatted expiration message
 */
export function formatExpirationMessage(expiryDate?: string, isHydrated = false): string {
  if (!expiryDate) return 'Never expires'
  
  try {
    const date = new Date(expiryDate)
    
    // Return ISO format during SSR to prevent hydration mismatch
    if (!isHydrated) {
      return date.toISOString().split('T')[0]
    }
    
    const daysLeft = calculateDaysLeft(expiryDate)
    
    if (daysLeft < 0) return 'Expired'
    if (daysLeft === 0) return 'Expires today'
    if (daysLeft === 1) return 'Expires tomorrow'
    if (daysLeft <= 30) return `Expires in ${daysLeft} days`
    
    return date.toLocaleDateString()
  } catch {
    return 'Invalid date'
  }
}

/**
 * Determine license status based on days remaining
 * @param daysLeft - Number of days until expiry
 * @param isValid - Whether license is currently valid
 * @returns License status string
 */
export function getLicenseStatusFromDays(daysLeft: number, isValid: boolean): 'invalid' | 'expired' | 'critical' | 'warning' | 'active' {
  if (!isValid) return 'invalid'
  if (daysLeft <= 0) return 'expired'
  if (daysLeft <= 7) return 'critical'
  if (daysLeft <= 30) return 'warning'
  return 'active'
}

/**
 * Calculate the number of days since expiry
 * @param expiryDate - ISO date string
 * @returns Number of days since expiry (0 if not expired)
 */
export function getDaysSinceExpiry(expiryDate?: string): number {
  if (!expiryDate) return 0
  
  try {
    const msSinceExpiry = Date.now() - new Date(expiryDate).getTime()
    return msSinceExpiry > 0 ? Math.floor(msSinceExpiry / MS_PER_DAY) : 0
  } catch {
    return 0
  }
}

/**
 * Check if a license is in grace period (recently expired)
 * @param expiryDate - ISO date string
 * @param gracePeriodDays - Number of days for grace period (default 7)
 * @returns Whether license is in grace period
 */
export function isInGracePeriod(expiryDate?: string, gracePeriodDays = 7): boolean {
  const daysSinceExpiry = getDaysSinceExpiry(expiryDate)
  return daysSinceExpiry > 0 && daysSinceExpiry <= gracePeriodDays
}