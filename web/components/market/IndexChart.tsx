/**
 * IndexChart Component
 *
 * TradingView Lightweight Charts component for displaying ISX60/ISX15 index data.
 * Simple line chart with responsive sizing and dark theme support.
 */

'use client'

import { useEffect, useRef } from 'react'
import { createChart, IChartApi, LineData, ColorType, LineSeries } from 'lightweight-charts'
import { useTheme } from 'next-themes'

interface IndexChartProps {
  indexName: 'ISX60' | 'ISX15'
  data: Array<{ date: string; value: number }>
  height?: number
}

export function IndexChart({ indexName, data, height = 300 }: IndexChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const { theme } = useTheme()

  useEffect(() => {
    if (!chartContainerRef.current || data.length === 0) return

    const isDark = theme === 'dark'

    // Create chart with dark/light theme
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: isDark ? '#1a1a1a' : '#ffffff' },
        textColor: isDark ? '#d1d5db' : '#374151',
      },
      grid: {
        vertLines: { color: isDark ? '#2B2B43' : '#e5e7eb' },
        horzLines: { color: isDark ? '#2B2B43' : '#e5e7eb' },
      },
      width: chartContainerRef.current.clientWidth,
      height,
      timeScale: {
        borderColor: isDark ? '#2B2B43' : '#d1d5db',
        timeVisible: true,
        secondsVisible: false,
      },
      rightPriceScale: {
        borderColor: isDark ? '#2B2B43' : '#d1d5db',
      },
      crosshair: {
        vertLine: {
          width: 1,
          color: isDark ? '#4169E1' : '#3b82f6',
          style: 1,
        },
        horzLine: {
          width: 1,
          color: isDark ? '#4169E1' : '#3b82f6',
          style: 1,
        },
      },
    })

    chartRef.current = chart

    // Add line series (v5.x API)
    const lineSeries = chart.addSeries(LineSeries, {
      color: indexName === 'ISX60' ? '#10b981' : '#3b82f6', // Green for ISX60, Blue for ISX15
      lineWidth: 2,
      crosshairMarkerVisible: true,
      crosshairMarkerRadius: 4,
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
    })

    // Transform data to LineData format (date string → unix timestamp)
    const chartData: LineData[] = data.map((item) => ({
      time: item.date as any, // TradingView accepts YYYY-MM-DD strings
      value: item.value,
    }))

    lineSeries.setData(chartData)

    // Fit content to visible range
    chart.timeScale().fitContent()

    // Handle window resize
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth })
      }
    }

    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
      chartRef.current = null
    }
  }, [data, indexName, theme, height])

  if (data.length === 0) {
    return (
      <div
        ref={chartContainerRef}
        className="flex items-center justify-center bg-muted rounded-lg"
        style={{ height }}
      >
        <p className="text-muted-foreground text-sm">No data available</p>
      </div>
    )
  }

  return (
    <div className="relative">
      <div
        ref={chartContainerRef}
        className="rounded-lg border border-border overflow-hidden"
      />
    </div>
  )
}
