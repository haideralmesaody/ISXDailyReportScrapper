/**
 * TickerSparkline Component
 * Following CLAUDE.md React patterns with memoization
 * Extracted from TickerList for better component separation
 */

'use client'

import React from 'react'

interface TickerSparklineProps {
  last10Days: string
  isPositive: boolean
  isNegative: boolean
  didTrade: boolean
  width?: number
  height?: number
}

// Generate sparkline path - memoized function
function generateSparklinePath(
  last10Days: string,
  width: number = 60,
  height: number = 20
): string | null {
  if (!last10Days) return null

  const prices = last10Days.split(',').map(p => parseFloat(p.trim())).filter(p => !isNaN(p))
  if (prices.length < 2) return null

  const min = Math.min(...prices)
  const max = Math.max(...prices)
  const range = max - min || 1

  const points = prices.map((price, i) => {
    const x = (i / (prices.length - 1)) * width
    const y = height - ((price - min) / range) * height
    return `${x},${y}`
  }).join(' ')

  return `M ${points}`
}

// Memoized sparkline component - CLAUDE.md performance pattern
function TickerSparklineComponent({
  last10Days,
  isPositive,
  isNegative,
  didTrade,
  width = 60,
  height = 20
}: TickerSparklineProps) {
  // Generate path - memoized
  const sparklinePath = React.useMemo(() =>
    generateSparklinePath(last10Days, width, height),
    [last10Days, width, height]
  )

  if (!sparklinePath) {
    return (
      <div className="flex items-center justify-center" style={{ width, height }}>
        <div className="w-1 h-1 bg-muted-foreground rounded-full" />
      </div>
    )
  }

  return (
    <svg width={width} height={height} className="overflow-visible">
      <path
        d={sparklinePath}
        fill="none"
        stroke={didTrade ? (isPositive ? '#22c55e' : isNegative ? '#ef4444' : '#94a3b8') : '#94a3b8'}
        strokeWidth="1.5"
      />
    </svg>
  )
}

// Props comparison for memoization - CLAUDE.md pattern
const propsAreEqual = (
  prevProps: TickerSparklineProps,
  nextProps: TickerSparklineProps
): boolean => {
  return (
    prevProps.last10Days === nextProps.last10Days &&
    prevProps.isPositive === nextProps.isPositive &&
    prevProps.isNegative === nextProps.isNegative &&
    prevProps.didTrade === nextProps.didTrade &&
    prevProps.width === nextProps.width &&
    prevProps.height === nextProps.height
  )
}

export const TickerSparkline = React.memo(TickerSparklineComponent, propsAreEqual)
TickerSparkline.displayName = 'TickerSparkline'