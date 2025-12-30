/**
 * useViewportPosition Hook
 *
 * Provides intelligent viewport-aware positioning for floating panels
 * with zoom compensation, boundary detection, and graceful fallbacks.
 *
 * Features:
 * - ResizeObserver + visualViewport API with window.innerWidth/Height fallback
 * - Throttled repositioning to prevent layout thrashing
 * - Viewport boundary clamping to prevent overflow
 * - Graceful fallback for older browsers (Safari <15, older Android)
 * - Performance optimized with 60fps throttling
 */

import { useState, useEffect, useRef, useCallback } from 'react'
import { useVisualViewportSupport, type ViewportSupportLevel } from './use-visual-viewport-support'

interface ViewportDimensions {
  width: number
  height: number
  offsetLeft: number
  offsetTop: number
  scale: number
}

interface Position {
  left: number    // Absolute pixel position from left edge
  top: number     // Absolute pixel position from top edge
  transform: string
}

interface UseViewportPositionOptions {
  /** Width of the panel in pixels */
  panelWidth?: number
  /** Height of the panel in pixels */
  panelHeight?: number
  /** Minimum distance from viewport edges in pixels */
  margin?: number
  /** Whether to recenter on resize */
  recenterOnResize?: boolean
  /** Throttle delay for repositioning in milliseconds */
  throttleDelay?: number
  /** Panel placement strategy */
  placement?: 'center' | 'right' | 'left'
  /** Enable debug mode for positioning calculations */
  debug?: boolean
  /** Enable error recovery and fallback positioning */
  enableErrorRecovery?: boolean
  /** Maximum retry attempts for positioning failures */
  maxRetries?: number
}

interface UseViewportPositionReturn {
  /** Position styles to apply to the panel */
  positionStyles: React.CSSProperties
  /** Current viewport dimensions */
  viewport: ViewportDimensions
  /** Whether visualViewport API is supported */
  isVisualViewportSupported: boolean
  /** Current viewport support level */
  supportLevel: ViewportSupportLevel
  /** Force reposition calculation */
  reposition: () => void
  /** Set the element to measure for accurate positioning */
  setElement: (element: HTMLElement | null) => void
  /** Positioning error information */
  lastError: string | null
  /** Retry count for positioning attempts */
  retryCount: number
}

export function useViewportPosition({
  panelWidth = 600,   // Will be overridden by actual element measurements
  panelHeight = 500,  // Will be overridden by actual element measurements
  margin = 20,
  recenterOnResize = true,
  throttleDelay = 16, // ~60fps
  placement = 'center',
  debug = false,
  enableErrorRecovery = true,
  maxRetries = 3,
}: UseViewportPositionOptions = {}): UseViewportPositionReturn {
  const [viewport, setViewport] = useState<ViewportDimensions>({
    width: 0,
    height: 0,
    offsetLeft: 0,
    offsetTop: 0,
    scale: 1,
  })

  const [position, setPosition] = useState<Position>({
    left: 50,
    top: 50,
    transform: 'translate(-50%, -50%)',
  })

  const elementRef = useRef<HTMLElement | null>(null)
  const resizeObserverRef = useRef<ResizeObserver | null>(null)
  const throttleTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const isSupportedRef = useRef<boolean>(false)

  // Error recovery state
  const [lastError, setLastError] = useState<string | null>(null)
  const [retryCount, setRetryCount] = useState(0)

  // Get tiered viewport support information
  const { supportLevel } = useVisualViewportSupport()

  // Check if visualViewport API is supported
  const checkVisualViewportSupport = useCallback((): boolean => {
    return (
      typeof window !== 'undefined' &&
      'visualViewport' in window &&
      window.visualViewport !== null &&
      typeof window.visualViewport.width === 'number' &&
      typeof window.visualViewport.height === 'number'
    )
  }, [])

  // Get viewport dimensions with tiered fallback strategy
  const getViewportDimensions = useCallback((): ViewportDimensions => {
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
    if (supportLevel === 'full' && checkVisualViewportSupport()) {
      try {
        const vv = window.visualViewport!
        const dims = {
          width: vv.width,
          height: vv.height,
          offsetLeft: vv.offsetLeft,
          offsetTop: vv.offsetTop,
          scale: vv.scale || 1,
        }

        if (debug) {
          console.log('🔍 Tier 1 (Full): Using visualViewport API', dims)
        }

        return dims
      } catch (error) {
        const errorMsg = `Tier 1 positioning failed: ${error instanceof Error ? error.message : 'Unknown error'}`
        setLastError(errorMsg)
        console.warn(errorMsg)
      }
    }

    // Tier 2: Partial support with hybrid approach
    if (supportLevel === 'partial' && checkVisualViewportSupport()) {
      try {
        const vv = window.visualViewport!
        // Use available visualViewport features, fallback for missing ones
        const dims = {
          width: vv.width || window.innerWidth,
          height: vv.height || window.innerHeight,
          offsetLeft: typeof vv.offsetLeft === 'number' ? vv.offsetLeft : 0,
          offsetTop: typeof vv.offsetTop === 'number' ? vv.offsetTop : 0,
          scale: typeof vv.scale === 'number' ? vv.scale : 1,
        }

        if (debug) {
          console.log('🔍 Tier 2 (Partial): Hybrid visualViewport approach', dims)
        }

        return dims
      } catch (error) {
        const errorMsg = `Tier 2 positioning failed: ${error instanceof Error ? error.message : 'Unknown error'}`
        setLastError(errorMsg)
        console.warn(errorMsg)
      }
    }

    // Tier 3: Fallback to window dimensions
    if (supportLevel === 'fallback') {
      try {
        const dims = {
          width: window.innerWidth,
          height: window.innerHeight,
          offsetLeft: 0,
          offsetTop: 0,
          scale: window.outerWidth > window.innerWidth ?
            Math.min(window.outerWidth / window.innerWidth, 3) : 1,
        }

        if (debug) {
          console.log('🔍 Tier 3 (Fallback): Window dimensions with zoom detection', dims)
        }

        return dims
      } catch (error) {
        const errorMsg = `Tier 3 positioning failed: ${error instanceof Error ? error.message : 'Unknown error'}`
        setLastError(errorMsg)
        console.warn(errorMsg)
      }
    }

    // Tier 4: Emergency static positioning
    const staticDims = {
      width: 1920,
      height: 1080,
      offsetLeft: 0,
      offsetTop: 0,
      scale: 1,
    }

    const errorMsg = `Tier 4 (Emergency): Using static positioning - no viewport support detected`
    setLastError(errorMsg)

    // Suppress emergency warnings completely to prevent console flooding
    // Emergency fallback is expected behavior in older browsers, not an error condition
    // The error state is still tracked in lastError for debugging if needed

    if (debug) {
      console.log('🔍 Tier 4 (Emergency): Static positioning', staticDims)
    }

    return staticDims
  }, [supportLevel, checkVisualViewportSupport, debug])

  // Calculate optimal position within viewport bounds with error recovery
  const calculatePosition = useCallback((
    viewportDims: ViewportDimensions,
    element: HTMLElement | null,
    attempt: number = 0
  ): Position => {
    try {
      // Get actual panel dimensions if element exists
      let actualPanelWidth = panelWidth
      let actualPanelHeight = panelHeight
      if (element) {
        const rect = element.getBoundingClientRect()
        actualPanelWidth = rect.width
        actualPanelHeight = rect.height
      }

      // Apply scale compensation
      const scaledPanelWidth = actualPanelWidth * viewportDims.scale
      const scaledPanelHeight = actualPanelHeight * viewportDims.scale

      // Clamp to viewport bounds with margin
      const maxX = Math.max(margin, viewportDims.width - scaledPanelWidth - margin)
      const maxY = Math.max(margin, viewportDims.height - scaledPanelHeight - margin)

      let finalX: number
      let finalY: number

      // Calculate position based on placement strategy
      switch (placement) {
        case 'left':
          // Position on left side of viewport
          finalX = margin
          finalY = Math.max(margin, Math.min((viewportDims.height - scaledPanelHeight) / 2, maxY))
          break
        case 'right':
          // Position on right side of viewport
          finalX = Math.max(margin, Math.min(viewportDims.width - scaledPanelWidth - margin, maxX))
          finalY = Math.max(margin, Math.min((viewportDims.height - scaledPanelHeight) / 2, maxY))
          break
        case 'center':
        default:
          // Center position as starting point
          let centerX = viewportDims.width / 2
          let centerY = viewportDims.height / 2

          // Adjust for visual viewport offsets (useful for iOS zoom)
          centerX -= viewportDims.offsetLeft
          centerY -= viewportDims.offsetTop

          finalX = Math.max(margin, Math.min(centerX - scaledPanelWidth / 2, maxX))
          finalY = Math.max(margin, Math.min(centerY - scaledPanelHeight / 2, maxY))
          break
      }

      // Debug logging if enabled
      if (debug) {
        console.log('🔍 useViewportPosition Debug:', {
          placement,
          viewportDims,
          actualPanelWidth,
          actualPanelHeight,
          scaledPanelWidth,
          scaledPanelHeight,
          finalX,
          finalY,
          attempt,
          elementRect: element?.getBoundingClientRect()
        })
      }

      // Clear retry count on successful calculation
      if (attempt > 0) {
        setRetryCount(0)
      }

      // Use absolute positioning to avoid double-shift centering
      // The positionStyles will use absolute left/top coordinates without percentage-based centering
      // No transform needed since we're positioning the top-left corner directly

      return {
        left: finalX,
        top: finalY,
        transform: 'translate(0, 0)', // No transform needed - we use absolute positioning
      }
    } catch (error) {
      const errorMsg = `Position calculation failed (attempt ${attempt}): ${error instanceof Error ? error.message : 'Unknown error'}`
      setLastError(errorMsg)
      console.warn(errorMsg)

      // Error recovery: try fallback positioning if enabled and within retry limit
      if (enableErrorRecovery && attempt < maxRetries) {
        // Use panel defaults for fallback since actual measurement failed
        const fallbackPanelWidth = element ? 0 : panelWidth
        const fallbackPanelHeight = element ? 0 : panelHeight
        const fallbackX = Math.max(margin, Math.min(50, viewportDims.width - fallbackPanelWidth - margin))
        const fallbackY = Math.max(margin, Math.min(50, viewportDims.height - fallbackPanelHeight - margin))

        if (debug) {
          console.log('🔍 Fallback positioning applied:', { fallbackX, fallbackY, attempt })
        }

        return {
          left: fallbackX,
          top: fallbackY,
          transform: 'translate(0, 0)',
        }
      }

      // Final fallback: center position in pixels respecting margins
      const centerLeft = Math.max(margin, (viewportDims.width - panelWidth) / 2)
      const centerTop = Math.max(margin, (viewportDims.height - panelHeight) / 2)

      if (debug) {
        console.log('🔍 Emergency fallback: Using pixel-based center position', { centerLeft, centerTop })
      }

      return {
        left: centerLeft,
        top: centerTop,
        transform: 'translate(0, 0)',
      }
    }
  }, [panelWidth, panelHeight, margin, placement, debug])

  // Throttled reposition function
  const throttledReposition = useCallback(() => {
    if (throttleTimeoutRef.current) {
      clearTimeout(throttleTimeoutRef.current)
    }

    throttleTimeoutRef.current = setTimeout(() => {
      const viewportDims = getViewportDimensions()
      setViewport(viewportDims)

      if (elementRef.current) {
        const newPosition = calculatePosition(viewportDims, elementRef.current)
        setPosition(newPosition)
      }
    }, throttleDelay)
  }, [getViewportDimensions, calculatePosition, throttleDelay])

  // Immediate reposition (not throttled)
  const reposition = useCallback(() => {
    const viewportDims = getViewportDimensions()
    setViewport(viewportDims)

    if (elementRef.current) {
      const newPosition = calculatePosition(viewportDims, elementRef.current)
      setPosition(newPosition)
    }
  }, [getViewportDimensions, calculatePosition])

  // Set up ResizeObserver for element changes
  const setupResizeObserver = useCallback((element: HTMLElement) => {
    if (typeof ResizeObserver === 'undefined') {
      return
    }

    // Clean up existing observer
    if (resizeObserverRef.current) {
      resizeObserverRef.current.disconnect()
    }

    resizeObserverRef.current = new ResizeObserver(() => {
      if (recenterOnResize) {
        throttledReposition()
      }
    })

    resizeObserverRef.current.observe(element)
  }, [recenterOnResize, throttledReposition])

  // Set up viewport resize listeners
  const setupViewportListeners = useCallback(() => {
    // VisualViewport API listeners
    if (checkVisualViewportSupport()) {
      const vv = window.visualViewport!

      const handleVisualViewportResize = () => {
        throttledReposition()
      }

      vv.addEventListener('resize', handleVisualViewportResize)
      vv.addEventListener('scroll', throttledReposition)

      return () => {
        vv.removeEventListener('resize', handleVisualViewportResize)
        vv.removeEventListener('scroll', throttledReposition)
      }
    }

    // Fallback window resize listener
    const handleWindowResize = () => {
      throttledReposition()
    }

    window.addEventListener('resize', handleWindowResize)
    window.addEventListener('orientationchange', handleWindowResize)

    return () => {
      window.removeEventListener('resize', handleWindowResize)
      window.removeEventListener('orientationchange', handleWindowResize)
    }
  }, [checkVisualViewportSupport, throttledReposition])

  // Initialize and set up listeners
  useEffect(() => {
    isSupportedRef.current = checkVisualViewportSupport()

    // Initial viewport detection
    reposition()

    // Set up viewport listeners
    const cleanupViewport = setupViewportListeners()

    return () => {
      // Clean up
      cleanupViewport()

      if (throttleTimeoutRef.current) {
        clearTimeout(throttleTimeoutRef.current)
      }

      if (resizeObserverRef.current) {
        resizeObserverRef.current.disconnect()
      }
    }
  }, [checkVisualViewportSupport, reposition, setupViewportListeners])

  // Ref for external access to element
  const setElement = useCallback((element: HTMLElement | null) => {
    if (elementRef.current !== element) {
      elementRef.current = element

      if (element) {
        setupResizeObserver(element)
        reposition() // Initial position with actual element
      }
    }
  }, [setupResizeObserver, reposition])

  // Combine position into CSS properties using absolute pixel positioning
  const positionStyles: React.CSSProperties = {
    position: 'fixed',
    left: `${position.left}px`,
    top: `${position.top}px`,
    transform: position.transform,
    zIndex: 50, // Consistent with CSS --panel-z
    willChange: 'transform', // Performance optimization
  }

  return {
    positionStyles,
    viewport,
    isVisualViewportSupported: isSupportedRef.current,
    supportLevel,
    reposition,
    // Expose element ref for external use
    setElement,
    // Error recovery information
    lastError,
    retryCount,
  }
}

// Export types for external use
export type { UseViewportPositionOptions, UseViewportPositionReturn, ViewportDimensions, Position }

// Utility function to check if viewport is zoomed
export const isViewportZoomed = (): boolean => {
  if (typeof window === 'undefined') return false

  return (
    ('visualViewport' in window &&
     window.visualViewport &&
     window.visualViewport.scale !== undefined &&
     window.visualViewport.scale > 1) ||
    window.outerWidth > window.innerWidth ||
    window.outerHeight > window.innerHeight
  )
}

// Utility function to get current zoom level
export const getViewportZoomLevel = (): number => {
  if (typeof window === 'undefined') return 1

  if ('visualViewport' in window && window.visualViewport) {
    return window.visualViewport.scale || 1
  }

  return Math.min(
    window.innerWidth / window.screen.width,
    window.innerHeight / window.screen.height
  )
}
