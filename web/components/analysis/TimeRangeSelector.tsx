/**
 * TimeRangeSelector Component
 * Industry-standard time period selector for chart data filtering
 * Following patterns from TradingView, Yahoo Finance, Bloomberg
 */

'use client'

import React, { useCallback } from 'react'
import { cn } from '@/lib/utils'

export type TimeRange = '1D' | '5D' | '1M' | '3M' | '6M' | 'YTD' | '1Y' | 'ALL'

interface TimeRangeSelectorProps {
  activeRange: TimeRange
  onRangeChange: (range: TimeRange) => void
  className?: string
}

const TIME_RANGES: { value: TimeRange; label: string }[] = [
  { value: '1D', label: '1D' },
  { value: '5D', label: '5D' },
  { value: '1M', label: '1M' },
  { value: '3M', label: '3M' },
  { value: '6M', label: '6M' },
  { value: 'YTD', label: 'YTD' },
  { value: '1Y', label: '1Y' },
  { value: 'ALL', label: 'All' }
]

export function TimeRangeSelector({ activeRange, onRangeChange, className }: TimeRangeSelectorProps) {
  const handleRangeClick = useCallback((range: TimeRange) => {
    if (range !== activeRange) {
      onRangeChange(range)
    }
  }, [activeRange, onRangeChange])

  return (
    <div className={cn("flex items-center gap-1 bg-muted/30 rounded-md p-1", className)}>
      {TIME_RANGES.map(({ value, label }) => (
        <button
          key={value}
          onClick={() => handleRangeClick(value)}
          className={cn(
            "px-3 py-1 text-xs font-medium rounded transition-all duration-200",
            "hover:bg-muted focus:outline-none focus:ring-2 focus:ring-primary/50",
            activeRange === value
              ? "bg-primary text-primary-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          )}
          aria-label={`View ${label} data`}
          aria-pressed={activeRange === value}
        >
          {label}
        </button>
      ))}
    </div>
  )
}
