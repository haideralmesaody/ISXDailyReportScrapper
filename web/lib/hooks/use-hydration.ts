'use client'

import React, { useState, useEffect } from 'react'

/**
 * Hook to detect when React hydration has completed
 * 
 * This hook helps prevent hydration mismatches by providing a way to
 * conditionally render content that depends on client-side state or APIs.
 * 
 * @example
 * ```tsx
 * function MyComponent() {
 *   const isHydrated = useHydration()
 *   
 *   if (!isHydrated) {
 *     return <LoadingState />
 *   }
 *   
 *   // Client-only content here
 *   return <div>{new Date().toLocaleString()}</div>
 * }
 * ```
 */
export function useHydration() {
  const [isHydrated, setIsHydrated] = useState(false)

  useEffect(() => {
    setIsHydrated(true)
  }, [])

  return isHydrated
}

/**
 * Hook to safely access client-only values with fallback for SSR
 * 
 * @param clientValue - Function that returns the client-side value
 * @param serverValue - Value to use during SSR (default: null)
 * 
 * @example
 * ```tsx
 * const currentYear = useClientValue(
 *   () => new Date().getFullYear(),
 *   2025 // fallback for SSR
 * )
 * ```
 */
export function useClientValue<T>(
  clientValue: () => T,
  serverValue: T | null = null
): T | null {
  const [value, setValue] = useState<T | null>(serverValue)
  const isHydrated = useHydration()

  useEffect(() => {
    if (isHydrated) {
      setValue(clientValue())
    }
  }, [isHydrated, clientValue])

  return value
}

/**
 * Higher-order component to wrap components that should only render on client
 * 
 * @example
 * ```tsx
 * const ClientOnlyChart = withHydration(ChartComponent, <ChartSkeleton />)
 * ```
 */
export function withHydration<P extends object>(
  Component: React.ComponentType<P>,
  fallback: React.ReactNode = null
) {
  return function HydratedComponent(props: P) {
    const isHydrated = useHydration()
    
    if (!isHydrated) {
      return fallback as React.ReactElement | null
    }
    
    return React.createElement(Component, props)
  }
}