'use client'

import { useEffect } from 'react'
import type { IChartApi } from 'lightweight-charts'
import type { TickerHistoricalData } from '@/types/analysis'

interface LightweightStockChartEnhancedProps {
  ticker: string
  companyName?: string
  data: TickerHistoricalData[]
  isLoading?: boolean
  onChartReady?: (chart: IChartApi | null) => void
}

/**
 * Placeholder implementation for the enhanced chart while the real component is unavailable.
 */
export function LightweightStockChartEnhanced({
  ticker,
  companyName,
  onChartReady
}: LightweightStockChartEnhancedProps) {
  useEffect(() => {
    onChartReady?.(null)
  }, [onChartReady])

  return (
    <div className="flex h-full min-h-[240px] w-full items-center justify-center rounded border border-dashed text-sm text-muted-foreground">
      Detailed chart for {companyName || ticker} is temporarily disabled in this build.
    </div>
  )
}

export default LightweightStockChartEnhanced
