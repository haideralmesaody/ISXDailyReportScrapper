/**
 * License utility functions for ISX Pulse
 * Provides formatting, validation, and rate limiting for license operations
 */

/**
 * Clean license key for submission (remove spaces, uppercase)
 * Keep dashes for scratch card format
 */
export function cleanLicenseKey(value: string): string {
  const trimmed = value.trim().toUpperCase()
  // If it has dashes in the right pattern, keep them (scratch card format)
  if (trimmed.match(/^ISX-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/)) {
    return trimmed
  }
  // Otherwise remove spaces and dashes for standard format
  return trimmed.replace(/[\s-]/g, '')
}

/**
 * Auto-format license key as user types
 * Supports both standard (ISX1M02LYE1F9QJHR9D7Z) and scratch card (ISX-XXXX-XXXX-XXXX-XXXX) formats
 */
export function formatLicenseKey(value: string, format: 'standard' | 'scratch' = 'standard'): string {
  const cleaned = value.trim().toUpperCase().replace(/[\s]/g, '') // Only remove spaces, not dashes
  
  if (format === 'scratch') {
    // Remove all non-alphanumeric for formatting
    const alphanum = cleaned.replace(/[^A-Z0-9]/g, '')
    
    // Format as ISX-XXXX-XXXX-XXXX-XXXX (20 chars total with dashes, 16 without)
    if (alphanum.length <= 3) return alphanum
    if (alphanum.length <= 7) return `${alphanum.slice(0, 3)}-${alphanum.slice(3)}`
    if (alphanum.length <= 11) return `${alphanum.slice(0, 3)}-${alphanum.slice(3, 7)}-${alphanum.slice(7)}`
    if (alphanum.length <= 15) return `${alphanum.slice(0, 3)}-${alphanum.slice(3, 7)}-${alphanum.slice(7, 11)}-${alphanum.slice(11)}`
    return `${alphanum.slice(0, 3)}-${alphanum.slice(3, 7)}-${alphanum.slice(7, 11)}-${alphanum.slice(11, 15)}-${alphanum.slice(15, 19)}`
  }
  
  // Standard format - remove all dashes
  return cleaned.replace(/-/g, '')
}

/**
 * Detect license key format based on content
 */
export function detectLicenseFormat(key: string): 'standard' | 'scratch' {
  const upperKey = key.toUpperCase()
  
  // Check if it matches scratch card pattern (with or without complete dashes)
  if (upperKey.includes('-') || (upperKey.startsWith('ISX') && !upperKey.match(/^ISX[136]M/))) {
    return 'scratch'
  }
  
  // Check for standard format prefixes
  if (upperKey.match(/^ISX[136]M/) || upperKey.match(/^ISX1Y/)) {
    return 'standard'
  }
  
  // Default based on length (scratch cards are 16 chars without dashes, standard are 15+)
  const cleaned = upperKey.replace(/[^A-Z0-9]/g, '')
  if (cleaned.length === 19 && cleaned.startsWith('ISX')) {
    return 'scratch'
  }
  
  return 'standard'
}

/**
 * Validate license key format (supports both standard and scratch card formats)
 */
export function isValidLicenseFormat(key: string): boolean {
  const upperKey = key.trim().toUpperCase()
  
  if (!upperKey.startsWith('ISX')) return false
  
  // Check scratch card format with dashes: ISX-XXXX-XXXX-XXXX-XXXX
  if (upperKey.match(/^ISX-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/)) {
    return true
  }
  
  // Check scratch card format without dashes (19 chars total)
  const noDashes = upperKey.replace(/-/g, '')
  if (noDashes.match(/^ISX[A-Z0-9]{16}$/)) {
    return true
  }
  
  // Check standard format: ISX1M/3M/6M/1Y followed by alphanumeric
  if (noDashes.match(/^ISX(1M|3M|6M|1Y)[A-Z0-9]{5,}$/)) {
    return true
  }
  
  return false
}

/**
 * Validate scratch card specific format
 */
export function isValidScratchCardFormat(key: string): boolean {
  const upperKey = key.trim().toUpperCase()
  
  // Match exact scratch card format: ISX-XXXX-XXXX-XXXX-XXXX or without dashes
  if (upperKey.match(/^ISX-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/)) {
    return true
  }
  
  const noDashes = upperKey.replace(/-/g, '')
  return noDashes.match(/^ISX[A-Z0-9]{16}$/) !== null
}

/**
 * Normalize license key input for consistent processing
 */
export function normalizeLicenseKey(value: string): string {
  return cleanLicenseKey(value)
}

/**
 * Rate limiting for activation attempts
 */
const RATE_LIMIT_KEY = 'license_activation_attempts'
const MAX_ATTEMPTS = 10  // Increased from 5 to be more lenient
const TIME_WINDOW = 300000 // 5 minutes (increased from 1 minute)

export function canAttemptActivation(): { allowed: boolean; remainingAttempts: number; resetTime?: number } {
  try {
    const now = Date.now()
    const attempts = JSON.parse(localStorage.getItem(RATE_LIMIT_KEY) || '[]') as number[]
    
    // Filter attempts within the time window
    const recentAttempts = attempts.filter(timestamp => now - timestamp < TIME_WINDOW)
    
    if (recentAttempts.length >= MAX_ATTEMPTS && recentAttempts.length > 0) {
      const oldestAttempt = recentAttempts[0]
      const resetTime = (oldestAttempt || now) + TIME_WINDOW
      
      return {
        allowed: false,
        remainingAttempts: 0,
        resetTime
      }
    }
    
    return {
      allowed: true,
      remainingAttempts: MAX_ATTEMPTS - recentAttempts.length
    }
  } catch {
    // If localStorage is unavailable, allow the attempt
    return { allowed: true, remainingAttempts: MAX_ATTEMPTS }
  }
}

/**
 * Clear rate limit data (useful for debugging or reset)
 */
export function clearRateLimitData(): void {
  try {
    localStorage.removeItem(RATE_LIMIT_KEY)
  } catch {
    // Ignore errors
  }
}

export function recordActivationAttempt(): void {
  try {
    const now = Date.now()
    const attempts = JSON.parse(localStorage.getItem(RATE_LIMIT_KEY) || '[]') as number[]
    
    // Keep only recent attempts and add new one
    const recentAttempts = attempts.filter(timestamp => now - timestamp < TIME_WINDOW)
    recentAttempts.push(now)
    
    localStorage.setItem(RATE_LIMIT_KEY, JSON.stringify(recentAttempts))
  } catch {
    // Ignore localStorage errors
  }
}

/**
 * Event name for license status updates
 */
export const LICENSE_STATUS_UPDATED = 'license-status-updated'

/**
 * License status event data interface
 */
export interface LicenseStatusEvent {
  status: string
  expiryDate?: string
  timestamp: number
}

/**
 * Enhanced broadcast license status update event with comprehensive error handling and fallbacks
 */
export function broadcastLicenseStatus(status: string, expiryDate?: string): void {
  try {
    // Ensure we're in a browser environment
    if (typeof window === 'undefined') {
      console.warn('License status broadcast attempted in non-browser environment')
      return
    }

    const eventData: LicenseStatusEvent = {
      status,
      expiryDate,
      timestamp: Date.now()
    }

    // Primary method: Modern CustomEvent API
    try {
      const event = new CustomEvent<LicenseStatusEvent>(LICENSE_STATUS_UPDATED, {
        detail: eventData,
        bubbles: true, // Allow event to bubble up through DOM
        cancelable: false // Event cannot be canceled
      })

      window.dispatchEvent(event)
    } catch (customEventError) {
      // Fallback 1: Try basic CustomEvent without generics
      console.warn('CustomEvent with generics failed, trying fallback:', customEventError)

      try {
        const event = new (window as any).CustomEvent(LICENSE_STATUS_UPDATED, {
          detail: eventData,
          bubbles: true,
          cancelable: false
        })

        window.dispatchEvent(event)
      } catch (basicCustomEventError) {
        // Fallback 2: Create a simple event-like object
        console.warn('Basic CustomEvent failed, using generic fallback:', basicCustomEventError)

        const fallbackEvent = {
          type: LICENSE_STATUS_UPDATED,
          detail: eventData,
          bubbles: true,
          cancelable: false,
          timestamp: Date.now()
        }

        // Try to dispatch as a generic event
        try {
          const genericEvent = new Event(LICENSE_STATUS_UPDATED, {
            bubbles: true,
            cancelable: false
          })

          // Attach the detail data as a custom property
          ;(genericEvent as any).detail = eventData
          window.dispatchEvent(genericEvent)
        } catch (genericEventError) {
          // Fallback 3: Store in a global variable for polling-based detection
          console.warn('All event methods failed, using global storage fallback:', genericEventError)

          // Store the last broadcast in a global variable that components can check
          ;(window as any).__lastLicenseBroadcast = {
            ...eventData,
            broadcastAt: Date.now()
          }
        }
      }
    }
  } catch (error) {
    console.error('Critical error in license status broadcast:', error)

    // Last resort: store in global variable
    try {
      if (typeof window !== 'undefined') {
        ;(window as any).__lastLicenseBroadcast = {
          status,
          expiryDate,
          timestamp: Date.now(),
          broadcastAt: Date.now(),
          error: true
        }
      }
    } catch {
      // Complete failure - log and continue
      console.error('Complete license broadcast failure')
    }
  }
}

/**
 * Cache license status locally for resilience
 */
const LICENSE_CACHE_KEY = 'license_status_cache'
const CACHE_DURATION = 300000 // 5 minutes

export interface CachedLicenseStatus {
  status: string
  expiryDate: string | undefined
  cachedAt: number
}

export function getCachedLicenseStatus(): CachedLicenseStatus | null {
  try {
    const cached = localStorage.getItem(LICENSE_CACHE_KEY)
    if (!cached) return null
    
    const data = JSON.parse(cached) as CachedLicenseStatus
    const now = Date.now()
    
    // Check if cache is still valid
    if (now - data.cachedAt > CACHE_DURATION) {
      localStorage.removeItem(LICENSE_CACHE_KEY)
      return null
    }
    
    return data
  } catch {
    return null
  }
}

export function setCachedLicenseStatus(status: string, expiryDate?: string): void {
  try {
    const data: CachedLicenseStatus = {
      status,
      expiryDate: expiryDate,
      cachedAt: Date.now()
    }
    localStorage.setItem(LICENSE_CACHE_KEY, JSON.stringify(data))

    // Broadcast the license status update for real-time app-wide notifications
    broadcastLicenseStatus(status, expiryDate)
  } catch (error) {
    console.warn('Failed to cache license status, trying broadcast only:', error)

    // Ignore localStorage errors, but still try to broadcast
    try {
      broadcastLicenseStatus(status, expiryDate)
    } catch (broadcastError) {
      console.warn('Failed to broadcast license status:', broadcastError)
      // Continue silently - license status will be updated on next poll
    }
  }
}

/**
 * Check for global license broadcast fallback (for environments where events fail)
 */
export function getLastLicenseBroadcast(): LicenseStatusEvent | null {
  try {
    if (typeof window !== 'undefined' && (window as any).__lastLicenseBroadcast) {
      const broadcast = (window as any).__lastLicenseBroadcast

      // Check if broadcast is recent (within last 30 seconds)
      const now = Date.now()
      if (now - broadcast.broadcastAt < 30000) {
        return {
          status: broadcast.status,
          expiryDate: broadcast.expiryDate,
          timestamp: broadcast.timestamp
        }
      } else {
        // Clean up old broadcast
        delete (window as any).__lastLicenseBroadcast
      }
    }
  } catch (error) {
    console.warn('Failed to check last license broadcast:', error)
  }

  return null
}

/**
 * Enhanced event listener registration with fallback support
 */
export function addLicenseStatusListener(callback: (event: LicenseStatusEvent) => void): () => void {
  if (typeof window === 'undefined') {
    return () => {} // No-op in non-browser environment
  }

  const handleEvent = (event: Event) => {
    try {
      const customEvent = event as CustomEvent
      const eventData = customEvent.detail as LicenseStatusEvent
      callback(eventData)
    } catch (error) {
      console.error('Error in license status event callback:', error)
    }
  }

  let eventListenerAdded = false

  try {
    window.addEventListener(LICENSE_STATUS_UPDATED, handleEvent)
    eventListenerAdded = true
  } catch (error) {
    console.warn('Failed to add license status event listener:', error)
  }

  // Return cleanup function
  return () => {
    if (eventListenerAdded) {
      try {
        window.removeEventListener(LICENSE_STATUS_UPDATED, handleEvent)
      } catch (error) {
        console.warn('Failed to remove license status event listener:', error)
      }
    }
  }
}

export function clearLicenseCache(): void {
  try {
    localStorage.removeItem(LICENSE_CACHE_KEY)
  } catch {
    // Ignore localStorage errors
  }
}

/**
 * Format time remaining for countdown
 */
export function formatCountdown(seconds: number): string {
  if (seconds <= 0) return 'now'
  if (seconds === 1) return '1 second'
  return `${seconds} seconds`
}

/**
 * Sleep utility for retry logic
 */
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * Exponential backoff retry wrapper
 */
export async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  maxRetries = 3,
  baseDelay = 1000
): Promise<T> {
  let lastError: Error | undefined
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      // Check if online before attempting
      if (!navigator.onLine) {
        throw new Error('No internet connection')
      }
      
      return await fn()
    } catch (error) {
      lastError = error as Error
      
      // Don't retry on client errors (4xx)
      if (error instanceof Error && error.message.includes('4')) {
        throw error
      }
      
      // If this is the last attempt, throw
      if (i === maxRetries - 1) {
        throw error
      }
      
      // Calculate delay with exponential backoff and jitter
      const delay = baseDelay * Math.pow(2, i) + Math.random() * 1000
      await sleep(delay)
    }
  }
  
  throw lastError || new Error('Retry failed')
}

/**
 * Copy text to clipboard with fallback
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    // Modern API
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
    
    // Fallback for older browsers
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.style.position = 'fixed'
    textArea.style.left = '-999999px'
    textArea.style.top = '-999999px'
    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()
    
    const successful = document.execCommand('copy')
    document.body.removeChild(textArea)
    
    return successful
  } catch {
    return false
  }
}

/**
 * Session-based redirect tracking to prevent duplicate redirects
 */
const REDIRECT_SESSION_KEY = 'license_redirect_session'

export function hasRedirectedThisSession(): boolean {
  try {
    return sessionStorage.getItem(REDIRECT_SESSION_KEY) === 'true'
  } catch {
    return false
  }
}

export function markRedirectedThisSession(): void {
  try {
    sessionStorage.setItem(REDIRECT_SESSION_KEY, 'true')
  } catch {
    // Ignore sessionStorage errors
  }
}

export function clearRedirectSession(): void {
  try {
    sessionStorage.removeItem(REDIRECT_SESSION_KEY)
  } catch {
    // Ignore sessionStorage errors
  }
}

/**
 * Analytics tracking (placeholder - implement with your analytics provider)
 */
export function trackLicenseEvent(
  event: 'activation_attempt' | 'activation_success' | 'activation_failure' | 'reactivation_success' | 'redirect' | 'redirect_preference_updated',
  data?: Record<string, any>
): void {
  // Log to console in development
  if (process.env.NODE_ENV === 'development') {
    console.log('[License Analytics]', event, data)
  }

  // TODO: Send to analytics service
  // window.gtag?.('event', event, data)
  // window.plausible?.('License', { props: { event, ...data } })
}

/**
 * User preference for auto-redirect behavior
 */
const REDIRECT_PREFERENCE_KEY = 'license_redirect_preference'

export interface RedirectPreference {
  disabled: boolean
  timestamp: number
}

export function getRedirectPreference(): RedirectPreference {
  try {
    const pref = localStorage.getItem(REDIRECT_PREFERENCE_KEY)
    if (!pref) return { disabled: false, timestamp: Date.now() }

    const data = JSON.parse(pref) as RedirectPreference

    // Reset preference after 30 days
    const thirtyDays = 30 * 24 * 60 * 60 * 1000
    if (Date.now() - data.timestamp > thirtyDays) {
      localStorage.removeItem(REDIRECT_PREFERENCE_KEY)
      return { disabled: false, timestamp: Date.now() }
    }

    return data
  } catch {
    return { disabled: false, timestamp: Date.now() }
  }
}

export function setRedirectPreference(preference: { disabled: boolean }): void {
  try {
    const data: RedirectPreference = {
      disabled: preference.disabled,
      timestamp: Date.now()
    }
    localStorage.setItem(REDIRECT_PREFERENCE_KEY, JSON.stringify(data))
    trackLicenseEvent('redirect_preference_updated', { disabled: preference.disabled })
  } catch {
    // Ignore localStorage errors
  }
}

export function clearRedirectPreference(): void {
  try {
    localStorage.removeItem(REDIRECT_PREFERENCE_KEY)
  } catch {
    // Ignore localStorage errors
  }
}