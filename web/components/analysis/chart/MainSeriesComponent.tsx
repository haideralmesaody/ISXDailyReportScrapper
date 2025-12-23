/**
 * MainSeriesComponent - Renders main price chart series
 * Supports multiple chart types: candlestick, line, area, bar (OHLC)
 * Always rendered - forms the base of the chart
 */

'use client'

import { CandlestickSeries, LineSeries, AreaSeries, BarSeries } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from './ChartContext'
import { logger } from '@/lib/utils/logger'

interface MainSeriesComponentProps {
  onSeriesCreated?: (series: any) => void  // Accept any series type since we support 4 types
}

export function MainSeriesComponent({ onSeriesCreated }: MainSeriesComponentProps) {
  const { chartData, theme, chartType } = useChart()

  // Log chart type changes
  logger.log(`[MainSeries] Rendering ${chartType} chart`)

  // Switch series type based on chartType prop
  switch (chartType) {
    case 'candlestick':
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
        onCreated: onSeriesCreated,
      })
      break

    case 'line':
      // Line chart - shows only close prices
      useSeries({
        seriesType: LineSeries,
        seriesOptions: {
          color: theme.candlestickUpColor,  // Use up color for line
          lineWidth: 2,
          priceScaleId: 'right',
        },
        data: chartData.candlestickData.map(d => ({
          time: d.time,
          value: d.close
        })),
        paneIndex: 0,
        onCreated: onSeriesCreated,
      })
      break

    case 'area':
      // Area chart - shows close prices with filled area below
      useSeries({
        seriesType: AreaSeries,
        seriesOptions: {
          topColor: theme.candlestickUpColor.replace(/[^,]+(?=\))/, '0.3'),  // Semi-transparent top
          bottomColor: theme.candlestickUpColor.replace(/[^,]+(?=\))/, '0.05'),  // Very transparent bottom
          lineColor: theme.candlestickUpColor,
          lineWidth: 2,
          priceScaleId: 'right',
        },
        data: chartData.candlestickData.map(d => ({
          time: d.time,
          value: d.close
        })),
        paneIndex: 0,
        onCreated: onSeriesCreated,
      })
      break

    case 'bar':
      // Bar (OHLC) chart - traditional OHLC bars
      useSeries({
        seriesType: BarSeries,
        seriesOptions: {
          upColor: theme.candlestickUpColor,
          downColor: theme.candlestickDownColor,
          priceScaleId: 'right',
        },
        data: chartData.candlestickData,
        paneIndex: 0,
        onCreated: onSeriesCreated,
      })
      break

    default:
      logger.error(`[MainSeries] Unknown chart type: ${chartType}`)
      // Fallback to candlestick
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
        paneIndex: 0,
        onCreated: onSeriesCreated,
      })
  }

  return null  // Component manages chart series, doesn't render DOM
}
