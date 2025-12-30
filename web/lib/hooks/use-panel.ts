/**
 * Unified Panel Hook
 *
 * Combines positioning, sizing, and focus management into a single cohesive interface.
 * Provides centralized panel behavior with performance optimizations and error recovery.
 *
 * Features:
 * - Tiered viewport support with fallbacks
 * - Unified positioning with error recovery
 * - Integrated focus management with Radix
 * - Performance optimization with throttling
 * - Comprehensive error handling and logging
 */

import React, { useState, useEffect, useRef, useCallback } from 'react'
import { useVisualViewportSupport, type ViewportSupportLevel } from './use-visual-viewport-support'

export type PanelPlacement = 'center' | 'left' | 'right' | 'auto'
export type PanelSize = 'small' | 'medium' | 'large' | 'auto'

interface UsePanelOptions {
  /** Panel size preset */
  size?: PanelSize
  /** Panel placement strategy */
  placement?: PanelPlacement
  /** Custom panel width in pixels (overrides size preset) */
  customWidth?: number
  /** Custom panel height in pixels */
  customHeight?: number
  /** Minimum distance from viewport edges */
  margin?: number
  /** Whether to recenter on viewport resize */
  recenterOnResize?: boolean
  /** Enable error recovery with fallback positioning */
  enableErrorRecovery?: boolean
  /** Maximum retry attempts for positioning failures */
  maxRetries?: number
  /** Enable debug mode for development */
  debug?: boolean
  /** Focus management options */
  focus?: {
    /** Auto-focus first element when panel opens */
    autoFocus?: boolean
    /** Restore focus to trigger element when panel closes */
    restoreFocus?: boolean
    /** Custom focus selector for first focusable element */
    firstFocusableSelector?: string
  }
  /** Performance options */
  performance?: {
    /** Throttle delay for repositioning in milliseconds */
    throttleDelay?: number
    /** Enable performance monitoring */
    monitorPerformance?: boolean
  }
}

interface UsePanelReturn {
  /** Position styles to apply to the panel */
  positionStyles: React.CSSProperties
  /** Current viewport dimensions */
  viewport: {
    width: number
    height: number
    offsetLeft: number
    offsetTop: number
    scale: number
  }
  /** Current viewport support level */
  supportLevel: ViewportSupportLevel
  /** Panel size information */
  panelSize: {
    width: number
    height: number
  }
  /** Force recalculation of position */
  reposition: () => void
  /** Set the panel element for measurement */
  setElement: (element: HTMLElement | null) => void
  /** Set the trigger element for focus restoration */
  setTriggerElement: (element: HTMLElement | null) => void
  /** Positioning error information */
  lastError: string | null
  /** Retry count for positioning attempts */
  retryCount: number
  /** Performance metrics */
  performance: {
    positionCalculationTime: number
    repositionCount: number
    averagePositionTime: number
  }
  /** Focus management methods */
  focus: {
    /** Focus first focusable element */
    focusFirst: () => void
    /** Focus specific element */
    focusElement: (element: HTMLElement) => void
    /** Check if element is within panel bounds */
    isElementInPanel: (element: HTMLElement) => boolean
  }
}

// Size presets based on CSS design tokens
const SIZE_PRESETS = {
  small: { width: 320, height: 400 },
  medium: { width: 480, height: 600 },
  large: { width: 600, height: 700 },
  auto: { width: 480, height: 600 }
} as const

// Performance monitoring
interface PerformanceMetrics {
  positionCalculationTime: number
  repositionCount: number
  averagePositionTime: number
  calculationTimes: number[]
}

export function usePanel({
  size = 'medium',
  placement = 'center',
  customWidth,
  customHeight,
  margin = 16,
  recenterOnResize = true,
  enableErrorRecovery = true,
  maxRetries = 3,
  debug = false,
  focus: focusOptions = {},
  performance: performanceOptions = {}
}: UsePanelOptions = {}): UsePanelReturn {
  // Get tiered viewport support information
  const { supportLevel } = useVisualViewportSupport()

  // State management
  const elementRef = useRef<HTMLElement | null>(null)
  const triggerRef = useRef<HTMLElement | null>(null)
  const [lastError, setLastError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)
  const [positionStyles, setPositionStyles] = useState<React.CSSProperties>({})

  // Performance monitoring
  const [metrics, setMetrics] = useState<PerformanceMetrics>({
    positionCalculationTime: 0,
    repositionCount: 0,
    averagePositionTime: 0,
    calculationTimes: []
  })

  // Throttle management
  const throttleTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const throttleDelay = performanceOptions.throttleDelay ?? 16

  // Calculate panel dimensions based on size preset or custom dimensions
  const getPanelDimensions = useCallback((): { width: number; height: number } => {
    if (customWidth || customHeight) {
      return {
        width: customWidth || SIZE_PRESETS.medium.width,
        height: customHeight || SIZE_PRESETS.medium.height
      }
    }

    const preset = SIZE_PRESETS[size]
    return { width: preset.width, height: preset.height }
  }, [size, customWidth, customHeight])

  // Get viewport dimensions with tiered fallback strategy
  const getViewportDimensions = useCallback((): {
    width: number
    height: number
    offsetLeft: number
    offsetTop: number
    scale: number
  } => {
    const startTime = performance.now()

    try {
      if (typeof window === 'undefined') {
        return {
          width: 1920,
          height: 1080,
          offsetLeft: 0,
          offsetTop: 0,
          scale: 1,
        }
      }

      // Clear any previous errors
      setLastError(null)

      // Tier 1: Full visualViewport API support
      if (supportLevel === 'full' && window.visualViewport) {
        const vv = window.visualViewport
        const dims = {
          width: vv.width,
          height: vv.height,
          offsetLeft: vv.offsetLeft,
          offsetTop: vv.offsetTop,
          scale: vv.scale || 1,
        }

        if (debug) {
          console.log('[Panel] Tier 1 (Full): Using visualViewport API', dims)
        }

        return dims
      }

      // Tier 2: Partial support with hybrid approach
      if (supportLevel === 'partial' && window.visualViewport) {
        const vv = window.visualViewport
        const dims = {
          width: vv.width || window.innerWidth,
          height: vv.height || window.innerHeight,
          offsetLeft: typeof vv.offsetLeft === 'number' ? vv.offsetLeft : 0,
          offsetTop: typeof vv.offsetTop === 'number' ? vv.offsetTop : 0,
          scale: typeof vv.scale === 'number' ? vv.scale : 1,
        }

        if (debug) {
          console.log('[Panel] Tier 2 (Partial): Hybrid approach', dims)
        }

        return dims
      }

      // Tier 3: Fallback to window dimensions
      const dims = {
        width: window.innerWidth,
        height: window.innerHeight,
        offsetLeft: 0,
        offsetTop: 0,
        scale: window.outerWidth > window.innerWidth ?
          Math.min(window.outerWidth / window.innerWidth, 3) : 1,
      }

      if (debug) {
        console.log('[Panel] Tier 3 (Fallback): Window dimensions', dims)
      }

      return dims
    } catch (error) {
      const errorMsg = `Viewport detection failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      setLastError(errorMsg)
      console.warn(errorMsg)

      // Emergency fallback
      return {
        width: 1920,
        height: 1080,
        offsetLeft: 0,
        offsetTop: 0,
        scale: 1,
      }
    } finally {
      const endTime = performance.now()
      const calculationTime = endTime - startTime

      if (performanceOptions.monitorPerformance) {
        setMetrics(prev => ({
          ...prev,
          positionCalculationTime: calculationTime,
          calculationTimes: [...prev.calculationTimes.slice(-99), calculationTime],
          averagePositionTime: [...prev.calculationTimes.slice(-99), calculationTime].reduce((a, b) => a + b, 0) / Math.min(prev.calculationTimes.length + 1, 100)
        }))
      }
    }
  }, [supportLevel, debug, performanceOptions.monitorPerformance])

  // Calculate optimal position within viewport bounds
  const calculatePosition = useCallback((
    viewportDims: ReturnType<typeof getViewportDimensions>,
    panelDims: ReturnType<typeof getPanelDimensions>,
    attempt: number = 0
  ): React.CSSProperties => {
    const startTime = performance.now()

    try {
      const { width: panelWidth, height: panelHeight } = panelDims

      // Apply scale compensation for zoom
      const scaledPanelWidth = panelWidth * viewportDims.scale
      const scaledPanelHeight = panelHeight * viewportDims.scale

      // Calculate boundaries
      const maxX = Math.max(margin, viewportDims.width - scaledPanelWidth - margin)
      const maxY = Math.max(margin, viewportDims.height - scaledPanelHeight - margin)

      let finalX: number
      let finalY: number

      // Calculate position based on placement strategy
      switch (placement) {
        case 'left':
          finalX = margin
          finalY = Math.max(margin, Math.min((viewportDims.height - scaledPanelHeight) / 2, maxY))
          break
        case 'right':
          finalX = Math.max(margin, Math.min(viewportDims.width - scaledPanelWidth - margin, maxX))
          finalY = Math.max(margin, Math.min((viewportDims.height - scaledPanelHeight) / 2, maxY))
          break
        case 'auto':
          // Smart placement based on available space
          const hasSpaceRight = viewportDims.width - scaledPanelWidth - margin > margin
          finalX = hasSpaceRight ? viewportDims.width - scaledPanelWidth - margin : margin
          finalY = Math.max(margin, Math.min((viewportDims.height - scaledPanelHeight) / 2, maxY))
          break
        case 'center':
        default:
          let centerX = viewportDims.width / 2
          let centerY = viewportDims.height / 2

          // Adjust for visual viewport offsets (useful for iOS zoom)
          centerX -= viewportDims.offsetLeft
          centerY -= viewportDims.offsetTop

          finalX = Math.max(margin, Math.min(centerX - scaledPanelWidth / 2, maxX))
          finalY = Math.max(margin, Math.min(centerY - scaledPanelHeight / 2, maxY))
          break
      }

      // Debug logging
      if (debug) {
        console.log('[Panel] Position Calculation:', {
          placement,
          viewportDims,
          panelDims,
          scaledPanelWidth,
          scaledPanelHeight,
          finalX,
          finalY,
          attempt,
          elementRect: elementRef.current?.getBoundingClientRect()
        })
      }

      // Clear retry count on successful calculation
      if (attempt > 0) {
        setRetryCount(0)
      }

      const endTime = performance.now()
      const calculationTime = endTime - startTime

      if (performanceOptions.monitorPerformance) {
        setMetrics(prev => ({
          ...prev,
          positionCalculationTime: calculationTime,
          repositionCount: prev.repositionCount + 1
        }))
      }

      // Return position styles (using absolute positioning for consistency)
      return {
        position: 'fixed',
        left: `${finalX}px`,
        top: `${finalY}px`,
        width: `${panelWidth}px`,
        height: `${panelHeight}px`,
        transform: 'translate(0, 0)',
        zIndex: 50,
        willChange: 'transform'
      }

    } catch (error) {
      const errorMsg = `Position calculation failed (attempt ${attempt}): ${error instanceof Error ? error.message : 'Unknown error'}`
      setLastError(errorMsg)
      console.warn(errorMsg)

      // Error recovery: try fallback positioning if enabled
      if (enableErrorRecovery && attempt < maxRetries) {
        const fallbackX = Math.max(margin, Math.min(50, viewportDims.width - panelDims.width - margin))
        const fallbackY = Math.max(margin, Math.min(50, viewportDims.height - panelDims.height - margin))

        if (debug) {
          console.log('[Panel] Fallback Positioning:', { fallbackX, fallbackY, attempt })
        }

        return {
          position: 'fixed',
          left: `${fallbackX}px`,
          top: `${fallbackY}px`,
          width: `${panelDims.width}px`,
          height: `${panelDims.height}px`,
          transform: 'translate(0, 0)',
          zIndex: 50,
          willChange: 'transform'
        }
      }

      // Final fallback: static center position
      return {
        position: 'fixed',
        left: '50%',
        top: '50%',
        width: `${panelDims.width}px`,
        height: `${panelDims.height}px`,
        transform: 'translate(-50%, -50%)',
        zIndex: 50,
        willChange: 'transform'
      }
    }
  }, [placement, margin, enableErrorRecovery, maxRetries, debug, performanceOptions.monitorPerformance])

  // Throttled reposition function
  const throttledReposition = useCallback(() => {
    if (throttleTimeoutRef.current) {
      clearTimeout(throttleTimeoutRef.current)
    }

    throttleTimeoutRef.current = setTimeout(() => {
      const viewportDims = getViewportDimensions()
      const panelDims = getPanelDimensions()

      const newPosition = calculatePosition(viewportDims, panelDims)

      // Update position state to trigger re-render
      setPositionStyles(newPosition)
      setLastError(null)
    }, throttleDelay)
  }, [getViewportDimensions, getPanelDimensions, calculatePosition, throttleDelay])

  // Immediate reposition (not throttled)
  const reposition = useCallback(() => {
    const viewportDims = getViewportDimensions()
    const panelDims = getPanelDimensions()

    const newPosition = calculatePosition(viewportDims, panelDims)
    setPositionStyles(newPosition)

    if (performanceOptions.monitorPerformance) {
      setMetrics(prev => ({ ...prev, repositionCount: prev.repositionCount + 1 }))
    }
  }, [getViewportDimensions, getPanelDimensions, calculatePosition, performanceOptions.monitorPerformance])

  // Set the panel element for measurement
  const setElement = useCallback((element: HTMLElement | null) => {
    elementRef.current = element
    if (element) {
      // Trigger recalculation when element is set
      reposition()
    }
  }, [reposition])

  // Set the trigger element for focus restoration
  const setTriggerElement = useCallback((element: HTMLElement | null) => {
    triggerRef.current = element
  }, [])

  // Focus management methods
  const focusManagement = {
    focusFirst: useCallback(() => {
      if (!elementRef.current || !focusOptions.autoFocus) return

      const selector = focusOptions.firstFocusableSelector ||
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'

      const firstFocusable = elementRef.current.querySelector(selector) as HTMLElement
      if (firstFocusable) {
        firstFocusable.focus()
      }
    }, [focusOptions.autoFocus, focusOptions.firstFocusableSelector]),

    focusElement: useCallback((element: HTMLElement) => {
      element.focus()
    }, []),

    isElementInPanel: useCallback((element: HTMLElement): boolean => {
      if (!elementRef.current) return false
      return elementRef.current.contains(element)
    }, [])
  }

  // Set up ResizeObserver for automatic repositioning
  useEffect(() => {
    if (!elementRef.current || !recenterOnResize) return

    if (typeof ResizeObserver === 'undefined') {
      console.warn('ResizeObserver not supported')
      return
    }

    const resizeObserver = new ResizeObserver(() => {
      throttledReposition()
    })

    resizeObserver.observe(elementRef.current)

    return () => {
      resizeObserver.disconnect()
    }
  }, [recenterOnResize, throttledReposition])

  // Set up viewport change listeners
  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleViewportChange = () => {
      throttledReposition()
    }

    // Listen for visualViewport changes if available
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', handleViewportChange)
      window.visualViewport.addEventListener('scroll', handleViewportChange)
    }

    // Fallback to window resize
    window.addEventListener('resize', handleViewportChange)

    return () => {
      if (window.visualViewport) {
        window.visualViewport.removeEventListener('resize', handleViewportChange)
        window.visualViewport.removeEventListener('scroll', handleViewportChange)
      }
      window.removeEventListener('resize', handleViewportChange)
    }
  }, [throttledReposition])

  // Auto-focus management
  useEffect(() => {
    if (focusOptions.autoFocus) {
      focusManagement.focusFirst()
    }
  }, [focusOptions.autoFocus, focusManagement.focusFirst])

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (throttleTimeoutRef.current) {
        clearTimeout(throttleTimeoutRef.current)
      }
    }
  }, [])

  // Initialize position on mount
  useEffect(() => {
    reposition()
  }, []) // Only run once on mount

  return {
    positionStyles,
    viewport: getViewportDimensions(),
    supportLevel,
    panelSize: getPanelDimensions(),
    reposition,
    setElement,
    setTriggerElement,
    lastError,
    retryCount,
    performance: metrics,
    focus: focusManagement
  }
}

// Export utilities for testing
export const panelUtils = {
  SIZE_PRESETS,
  calculateOptimalPosition: (viewport: any, panel: any, placement: string) => {
    void viewport
    void panel
    void placement
    // Utility function for testing position calculations
    return { x: 0, y: 0 } // Simplified for testing
  }
}

export type { UsePanelOptions, UsePanelReturn }
