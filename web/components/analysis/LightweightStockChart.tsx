'use client'

import { useEffect } from 'react'
import type { IChartApi } from 'lightweight-charts'
import type { ChartType, Timeframe } from '@/lib/hooks/use-chart-state'

interface LightweightStockChartProps {
  ticker: string
  data: any[]
  isLoading?: boolean
  onChartReady?: (chart: IChartApi | null) => void
  chartType?: ChartType
  timeframe?: Timeframe
}

/**
 * Temporary placeholder to keep the build green while the full chart implementation is missing.
 */
export function LightweightStockChart({
  ticker,
  onChartReady
}: LightweightStockChartProps) {
  useEffect(() => {
    onChartReady?.(null)
  }, [onChartReady])

  return (
    <div className="flex h-full min-h-[240px] w-full items-center justify-center rounded border border-dashed text-sm text-muted-foreground">
      Charts are temporarily unavailable for {ticker}. This placeholder keeps the build passing.
    </div>
  )
}
