/**
 * TreemapLegend - Color legend for market treemap
 * Shows what each color represents (gainers, losers, unchanged)
 * Following CLAUDE.md component standards
 */

'use client'

import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

export function TreemapLegend() {
  return (
    <div className="flex items-center justify-end gap-4 bg-muted/50 border border-border rounded-lg px-4 py-2 mb-4">
      <div className="flex items-center gap-2">
        <div className="flex items-center justify-center w-6 h-6 rounded bg-green-500/20 border border-green-500">
          <TrendingUp className="h-3.5 w-3.5 text-green-500" />
        </div>
        <span className="text-sm font-medium">Gainers</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="flex items-center justify-center w-6 h-6 rounded bg-red-500/20 border border-red-500">
          <TrendingDown className="h-3.5 w-3.5 text-red-500" />
        </div>
        <span className="text-sm font-medium">Losers</span>
      </div>
      <div className="flex items-center gap-2">
        <div className="flex items-center justify-center w-6 h-6 rounded bg-gray-400/20 border border-gray-400">
          <Minus className="h-3.5 w-3.5 text-gray-400" />
        </div>
        <span className="text-sm font-medium">Unchanged</span>
      </div>
      <div className="h-4 w-px bg-border mx-1" aria-hidden="true" />
      <span className="text-xs text-muted-foreground">Size = Traded Value | Hover for details</span>
    </div>
  )
}
