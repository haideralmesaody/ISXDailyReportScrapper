/**
 * useAdaptiveSize Hook
 *
 * Provides adaptive sizing capabilities based on viewport dimensions and zoom levels.
 * Helps components intelligently adjust their layout and sizing for different scenarios.
 *
 * Features:
 * - Detects viewport width and zoom levels
 * - Provides breakpoints for adaptive layouts
 * - Touch-friendly sizing adjustments
 * - Container query-like behavior
 * - Zoom compensation for touch targets
 */

import { useState, useEffect, useCallback, useMemo } from 'react'

interface AdaptiveBreakpoints {
  xs: number  // 0px - small mobile
  sm: number  // 640px - large mobile, small tablet
  md: number  // 768px - tablet
  lg: number  // 1024px - desktop
  xl: number  // 1280px - large desktop
  '2xl': number // 1536px - extra large desktop
}

interface AdaptiveSizeReturn {
  /** Current viewport width */
  width: number
  /** Current viewport height */
  height: number
  /** Current zoom level */
  zoomLevel: number
  /** Whether viewport is zoomed > 100% */
  isZoomed: boolean
  /** Current breakpoint based on width */
  breakpoint: keyof AdaptiveBreakpoints
  /** Whether current viewport is mobile size */
  isMobile: boolean
  /** Whether current viewport is tablet size */
  isTablet: boolean
  /** Whether current viewport is desktop size */
  isDesktop: boolean
  /** Touch-friendly size multiplier based on zoom */
  touchMultiplier: number
  /** Adaptive font size class */
  fontSizeClass: string
  /** Adaptive spacing class */
  spacingClass: string
  /** Check if viewport width is at least breakpoint */
  isMinWidth: (breakpoint: keyof AdaptiveBreakpoints) => boolean
  /** Check if viewport width is at most breakpoint */
  isMaxWidth: (breakpoint: keyof AdaptiveBreakpoints) => boolean
  /** Get adaptive size based on viewport */
  getAdaptiveSize: (sizes: Partial<Record<keyof AdaptiveBreakpoints, any>>) => any
}

const DEFAULT_BREAKPOINTS: AdaptiveBreakpoints = {
  xs: 0,
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
}

export function useAdaptiveSize(
  customBreakpoints?: Partial<AdaptiveBreakpoints>
): AdaptiveSizeReturn {
  const [dimensions, setDimensions] = useState({
    width: typeof window !== 'undefined' ? window.innerWidth : 1024,
    height: typeof window !== 'undefined' ? window.innerHeight : 768,
  })

  const [zoomLevel, setZoomLevel] = useState(1)

  // Merge custom breakpoints with defaults
  const breakpoints = useMemo(() => ({
    ...DEFAULT_BREAKPOINTS,
    ...customBreakpoints,
  }), [customBreakpoints])

  // Get current zoom level
  const getZoomLevel = useCallback((): number => {
    if (typeof window === 'undefined') return 1

    // Try visualViewport API first
    if ('visualViewport' in window && window.visualViewport) {
      return window.visualViewport.scale || 1
    }

    // Fallback calculation
    return Math.min(
      window.innerWidth / window.screen.width,
      window.innerHeight / window.screen.height
    )
  }, [])

  // Determine current breakpoint
  const getBreakpoint = useCallback((width: number): keyof AdaptiveBreakpoints => {
    if (width >= breakpoints['2xl']) return '2xl'
    if (width >= breakpoints.xl) return 'xl'
    if (width >= breakpoints.lg) return 'lg'
    if (width >= breakpoints.md) return 'md'
    if (width >= breakpoints.sm) return 'sm'
    return 'xs'
  }, [breakpoints])

  // Update dimensions and zoom level
  const updateDimensions = useCallback(() => {
    if (typeof window === 'undefined') return

    setDimensions({
      width: window.innerWidth,
      height: window.innerHeight,
    })

    setZoomLevel(getZoomLevel())
  }, [getZoomLevel])

  // Set up event listeners
  useEffect(() => {
    if (typeof window === 'undefined') return

    // Initial measurement
    updateDimensions()

    // Handle window resize
    const handleResize = () => {
      updateDimensions()
    }

    // Handle visual viewport changes (iOS zoom, keyboard, etc.)
    const handleVisualViewportResize = () => {
      updateDimensions()
    }

    // Add event listeners
    window.addEventListener('resize', handleResize, { passive: true })
    window.addEventListener('orientationchange', handleResize, { passive: true })

    if ('visualViewport' in window && window.visualViewport) {
      window.visualViewport.addEventListener('resize', handleVisualViewportResize, { passive: true })
    }

    return () => {
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('orientationchange', handleResize)

      if ('visualViewport' in window && window.visualViewport) {
        window.visualViewport.removeEventListener('resize', handleVisualViewportResize)
      }
    }
  }, [updateDimensions])

  // Derived state
  const breakpoint = useMemo(() => getBreakpoint(dimensions.width), [dimensions.width, getBreakpoint])
  const isZoomed = useMemo(() => zoomLevel > 1.05, [zoomLevel])
  const isMobile = useMemo(() => ['xs', 'sm'].includes(breakpoint), [breakpoint])
  const isTablet = useMemo(() => breakpoint === 'md', [breakpoint])
  const isDesktop = useMemo(() => ['lg', 'xl', '2xl'].includes(breakpoint), [breakpoint])

  // Touch-friendly multiplier
  const touchMultiplier = useMemo(() => {
    if (!isZoomed) return 1
    return Math.max(1, Math.min(zoomLevel, 2)) // Cap at 2x for usability
  }, [zoomLevel, isZoomed])

  // Adaptive font size class
  const fontSizeClass = useMemo(() => {
    const baseClasses = {
      xs: 'text-sm',
      sm: 'text-sm',
      md: 'text-base',
      lg: 'text-base',
      xl: 'text-lg',
      '2xl': 'text-lg',
    }

    return baseClasses[breakpoint]
  }, [breakpoint])

  // Adaptive spacing class
  const spacingClass = useMemo(() => {
    const baseClasses = {
      xs: 'space-y-2',
      sm: 'space-y-3',
      md: 'space-y-4',
      lg: 'space-y-6',
      xl: 'space-y-6',
      '2xl': 'space-y-8',
    }

    return baseClasses[breakpoint]
  }, [breakpoint])

  // Breakpoint checking utilities
  const isMinWidth = useCallback((checkBreakpoint: keyof AdaptiveBreakpoints): boolean => {
    return dimensions.width >= breakpoints[checkBreakpoint]
  }, [dimensions.width, breakpoints])

  const isMaxWidth = useCallback((checkBreakpoint: keyof AdaptiveBreakpoints): boolean => {
    return dimensions.width < breakpoints[checkBreakpoint]
  }, [dimensions.width, breakpoints])

  // Get adaptive size based on viewport
  const getAdaptiveSize = useCallback(<T>(sizes: Partial<Record<keyof AdaptiveBreakpoints, T>>): T | undefined => {
    // Find the largest breakpoint that fits
    const sortedBreakpoints = Object.entries(breakpoints)
      .sort(([, a], [, b]) => b - a) // Sort descending

    for (const [breakpointName] of sortedBreakpoints) {
      const bp = breakpointName as keyof AdaptiveBreakpoints
      if (dimensions.width >= breakpoints[bp] && sizes[bp] !== undefined) {
        return sizes[bp]
      }
    }

    // Fallback to smallest breakpoint
    return sizes.xs
  }, [dimensions.width, breakpoints])

  return {
    width: dimensions.width,
    height: dimensions.height,
    zoomLevel,
    isZoomed,
    breakpoint,
    isMobile,
    isTablet,
    isDesktop,
    touchMultiplier,
    fontSizeClass,
    spacingClass,
    isMinWidth,
    isMaxWidth,
    getAdaptiveSize,
  }
}

// Utility functions for common adaptive patterns
export const useAdaptiveGridCols = (config: {
  xs?: number
  sm?: number
  md?: number
  lg?: number
  xl?: number
  '2xl'?: number
}): string => {
  const { getAdaptiveSize } = useAdaptiveSize()

  const cols = getAdaptiveSize(config) || 1
  return `grid-cols-${cols}`
}

export const useAdaptiveSpacing = (config: {
  xs?: string
  sm?: string
  md?: string
  lg?: string
  xl?: string
  '2xl'?: string
}): string => {
  const { getAdaptiveSize } = useAdaptiveSize()

  return getAdaptiveSize(config) || 'space-y-4'
}

export const useAdaptiveTextSize = (config: {
  xs?: string
  sm?: string
  md?: string
  lg?: string
  xl?: string
  '2xl'?: string
}): string => {
  const { getAdaptiveSize } = useAdaptiveSize()

  return getAdaptiveSize(config) || 'text-base'
}

// Export types
export type { AdaptiveBreakpoints, AdaptiveSizeReturn }

// Preset configurations for common use cases
export const ADAPTIVE_PRESETS = {
  // Touch-friendly button sizes
  buttonSizes: {
    xs: 'px-3 py-1.5 text-sm',
    sm: 'px-4 py-2 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-base',
    xl: 'px-6 py-3 text-lg',
    '2xl': 'px-8 py-4 text-lg',
  },

  // Responsive grids
  gridColumns: {
    xs: 1,
    sm: 2,
    md: 2,
    lg: 3,
    xl: 4,
    '2xl': 4,
  },

  // Font sizes
  textSizes: {
    xs: 'text-sm',
    sm: 'text-sm',
    md: 'text-base',
    lg: 'text-base',
    xl: 'text-lg',
    '2xl': 'text-xl',
  },

  // Spacing
  spacing: {
    xs: 'space-y-2',
    sm: 'space-y-3',
    md: 'space-y-4',
    lg: 'space-y-6',
    xl: 'space-y-6',
    '2xl': 'space-y-8',
  },
} as const