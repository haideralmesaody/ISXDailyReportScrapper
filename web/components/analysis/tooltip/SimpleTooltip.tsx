/**
 * SimpleTooltip Component
 * Following CLAUDE.md patterns with simplified tooltip data
 * Shows OHLC + Volume + Active Indicators only (per user requirements)
 */

'use client'

import React from 'react'
import { formatPrice, formatVolume, formatDate } from '@/lib/utils/chart-helpers'

interface TooltipData {
  date: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  indicators?: Record<string, number>
}

interface SimpleTooltipProps {
  data: TooltipData | null
  position: { x: number; y: number }
  containerBounds: { width: number; height: number }
}

export function SimpleTooltip({ data, position, containerBounds }: SimpleTooltipProps) {
  if (!data) return null

  // Calculate tooltip position to keep it within bounds
  const tooltipWidth = 220
  const tooltipHeight = 140
  const padding = 10

  let x = position.x + padding
  let y = position.y + padding

  // Adjust horizontal position
  if (x + tooltipWidth > containerBounds.width) {
    x = position.x - tooltipWidth - padding
  }

  // Adjust vertical position
  if (y + tooltipHeight > containerBounds.height) {
    y = position.y - tooltipHeight - padding
  }

  // Ensure tooltip doesn't go outside bounds
  x = Math.max(padding, Math.min(x, containerBounds.width - tooltipWidth - padding))
  y = Math.max(padding, Math.min(y, containerBounds.height - tooltipHeight - padding))

  return (
    <div
      className="fixed z-50 bg-background/95 backdrop-blur-sm border border-border rounded-lg shadow-lg p-3 pointer-events-none"
      style={{
        left: `${x}px`,
        top: `${y}px`,
        width: `${tooltipWidth}px`
      }}
    >
      {/* Date */}
      <div className="text-xs font-medium text-muted-foreground mb-2">
        {formatDate(data.date)}
      </div>

      {/* OHLC Data */}
      <div className="grid grid-cols-2 gap-2 text-xs">
        <div className="space-y-1">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Open:</span>
            <span className="font-medium">{formatPrice(data.open)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">High:</span>
            <span className="font-medium text-green-600">{formatPrice(data.high)}</span>
          </div>
        </div>
        <div className="space-y-1">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Low:</span>
            <span className="font-medium text-red-600">{formatPrice(data.low)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Close:</span>
            <span className="font-medium">{formatPrice(data.close)}</span>
          </div>
        </div>
      </div>

      {/* Volume */}
      <div className="mt-2 pt-2 border-t border-border/50">
        <div className="flex justify-between text-xs">
          <span className="text-muted-foreground">Volume:</span>
          <span className="font-medium">{formatVolume(data.volume)}</span>
        </div>
      </div>

      {/* Active Indicators */}
      {data.indicators && Object.keys(data.indicators).length > 0 && (
        <div className="mt-2 pt-2 border-t border-border/50">
          <div className="space-y-1">
            {Object.entries(data.indicators).map(([name, value]) => (
              <div key={name} className="flex justify-between text-xs">
                <span className="text-muted-foreground">{name}:</span>
                <span className="font-medium">{formatPrice(value)}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}