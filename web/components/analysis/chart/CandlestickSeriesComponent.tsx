/**
 * CandlestickSeriesComponent - Main price chart series
 * Always rendered - forms the base of the chart
 */

'use client'

import { CandlestickSeries } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from './ChartContext'

interface CandlestickSeriesComponentProps {
  onSeriesCreated?: (series: ISeriesApi<'Candlestick'>) => void
}

export function CandlestickSeriesComponent({ onSeriesCreated }: CandlestickSeriesComponentProps) {
  const { chartData, theme } = useChart()

  // Use built-in onCreated callback from useSeries hook
  // This is the correct way to pass series reference to parent
  useSeries({
    seriesType: CandlestickSeries,
    seriesOptions: {
      upColor: theme.candlestickUpColor,
      downColor: theme.candlestickDownColor,
      borderUpColor: theme.candlestickUpColor,
      borderDownColor: theme.candlestickDownColor,
      wickUpColor: theme.candlestickUpColor,
      wickDownColor: theme.candlestickDownColor,
      priceScaleId: 'right',
    },
    data: chartData.candlestickData,
    paneIndex: 0,  // Main pane
    onCreated: onSeriesCreated,  // ✅ Use built-in callback instead of custom useEffect
  })

  return null  // Component manages chart series, doesn't render DOM
}
