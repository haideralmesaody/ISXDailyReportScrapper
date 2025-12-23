/**
 * useVisualViewportSupport Hook
 *
 * Provides tiered visual viewport API support detection with data-driven fallbacks.
 * Logs browser/condition data to inform future browser-specific improvements.
 *
 * Support Tiers:
 * - 'full': Complete visualViewport API support with all features
 * - 'partial': Basic support with some limitations
 * - 'fallback': No visualViewport, uses window dimensions
 * - 'none': Complete failure, uses static positioning
 */

import { useState, useEffect, useCallback } from 'react'

export type ViewportSupportLevel = 'full' | 'partial' | 'fallback' | 'none'

interface ViewportSupportInfo {
  level: ViewportSupportLevel
  userAgent: string
  browserName: string
  browserVersion: string
  platform: string
  isMobile: boolean
  features: {
    visualViewport: boolean
    offsetLeft: boolean
    offsetTop: boolean
    scale: boolean
    resizeEvents: boolean
  }
  detectedAt: string
}

interface UseVisualViewportSupportReturn {
  supportInfo: ViewportSupportInfo | null
  isSupported: boolean
  supportLevel: ViewportSupportLevel
  logSupport: () => void
}

// Browser detection utilities
const detectBrowser = (): { name: string; version: string } => {
  if (typeof window === 'undefined') return { name: 'unknown', version: 'unknown' }

  const ua = navigator.userAgent
  let name = 'unknown'
  let version = 'unknown'

  // Chrome/Edge
  if (ua.includes('Chrome/') && !ua.includes('Edg/')) {
    const match = ua.match(/Chrome\/(\d+)/)
    name = 'chrome'
    version = match?.[1] || 'unknown'
  }
  // Edge
  else if (ua.includes('Edg/')) {
    const match = ua.match(/Edg\/(\d+)/)
    name = 'edge'
    version = match?.[1] || 'unknown'
  }
  // Safari
  else if (ua.includes('Safari/') && !ua.includes('Chrome')) {
    const match = ua.match(/Version\/(\d+)/)
    name = 'safari'
    version = match?.[1] || 'unknown'
  }
  // Firefox
  else if (ua.includes('Firefox/')) {
    const match = ua.match(/Firefox\/(\d+)/)
    name = 'firefox'
    version = match?.[1] || 'unknown'
  }
  // iOS Safari (special case)
  else if (ua.includes('iPhone') || ua.includes('iPad')) {
    const match = ua.match(/OS (\d+)_?(\d*)/)
    name = 'ios-safari'
    version = match?.[1] ? `${match[1]}.${match[2] || '0'}` : 'unknown'
  }
  // Android
  else if (ua.includes('Android')) {
    const match = ua.match(/Android (\d+)/)
    name = 'android'
    version = match?.[1] || 'unknown'
  }

  return { name, version }
}

// Feature detection with detailed analysis
const detectViewportFeatures = (): ViewportSupportInfo['features'] => {
  if (typeof window === 'undefined') {
    return {
      visualViewport: false,
      offsetLeft: false,
      offsetTop: false,
      scale: false,
      resizeEvents: false,
    }
  }

  const hasVisualViewport = 'visualViewport' in window && window.visualViewport !== null
  const vv = hasVisualViewport ? window.visualViewport : null

  return {
    visualViewport: hasVisualViewport,
    offsetLeft: vv ? typeof vv.offsetLeft === 'number' : false,
    offsetTop: vv ? typeof vv.offsetTop === 'number' : false,
    scale: vv ? typeof vv.scale === 'number' : false,
    resizeEvents: vv ? 'addEventListener' in vv : false,
  }
}

// Determine support level based on features and browser
const determineSupportLevel = (
  features: ViewportSupportInfo['features'],
  browser: { name: string; version: string }
): ViewportSupportLevel => {
  // Full support - modern browsers with complete visualViewport API
  if (
    features.visualViewport &&
    features.offsetLeft !== false &&
    features.offsetTop !== false &&
    features.scale !== false &&
    features.resizeEvents
  ) {
    // Known good browsers with full support
    if (
      (browser.name === 'chrome' && parseInt(browser.version) >= 100) ||
      (browser.name === 'edge' && parseInt(browser.version) >= 100) ||
      (browser.name === 'firefox' && parseInt(browser.version) >= 100) ||
      (browser.name === 'ios-safari' && parseInt(browser.version) >= 15)
    ) {
      return 'full'
    }

    // Partial support - older browsers with some limitations
    return 'partial'
  }

  // Fallback support - browsers with basic window dimensions only
  if (typeof window !== 'undefined' && 'innerWidth' in window && 'innerHeight' in window) {
    return 'fallback'
  }

  // No support - very old browsers or SSR
  return 'none'
}

// Logger for support data (will be enhanced with actual telemetry)
const logViewportSupport = (supportInfo: ViewportSupportInfo): void => {
  // Log to console in development
  if (process.env.NODE_ENV === 'development') {
    console.group('[ViewportSupport] Detection Results')
    console.log('Support Level:', supportInfo.level)
    console.log('Browser:', `${supportInfo.browserName} ${supportInfo.browserVersion}`)
    console.log('Platform:', supportInfo.platform)
    console.log('Features:', supportInfo.features)
    console.log('Detected At:', supportInfo.detectedAt)
    console.groupEnd()
  }

  // TODO: Send to analytics/telemetry service
  // This will be implemented in Phase 4.2
  try {
    // Example telemetry integration:
    // analytics.track('viewport_support_detected', supportInfo)
  } catch (error) {
    // Silently fail to avoid breaking the app
    console.warn('Failed to log viewport support data:', error)
  }
}

export function useVisualViewportSupport(): UseVisualViewportSupportReturn {
  const [supportInfo, setSupportInfo] = useState<ViewportSupportInfo | null>(null)

  const detectSupport = useCallback((): ViewportSupportInfo | null => {
    if (typeof window === 'undefined') return null

    const browser = detectBrowser()
    const features = detectViewportFeatures()
    const level = determineSupportLevel(features, browser)

    return {
      level,
      userAgent: navigator.userAgent,
      browserName: browser.name,
      browserVersion: browser.version,
      platform: navigator.platform || 'unknown',
      isMobile: /iPhone|iPad|iPod|Android/i.test(navigator.userAgent),
      features,
      detectedAt: new Date().toISOString(),
    }
  }, [])

  const logSupport = useCallback((): void => {
    if (supportInfo) {
      logViewportSupport(supportInfo)
    }
  }, [supportInfo])

  // Detect support on mount
  useEffect(() => {
    const info = detectSupport()
    setSupportInfo(info)

    // Log support data automatically
    if (info) {
      logViewportSupport(info)
    }
  }, [detectSupport])

  return {
    supportInfo,
    isSupported: supportInfo?.level !== 'none',
    supportLevel: supportInfo?.level || 'none',
    logSupport,
  }
}

// Export utilities for testing and debugging
export const viewportDetectionUtils = {
  detectBrowser,
  detectViewportFeatures,
  determineSupportLevel,
  logViewportSupport,
}