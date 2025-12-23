'use client'

import { useState, useEffect, useRef, useMemo } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickData,
  Time,
  CandlestickSeries,
  LineSeries,
  LineData,
  LineStyle
} from 'lightweight-charts'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Loader2, TrendingUp } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { calculateRSI } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'

interface RSIChartDemoProps {
  className?: string
}

/**
 * Interactive RSI chart demo with period controls and zone highlighting
 * Uses TradingView Lightweight Charts v5 with 2 panes (price + RSI)
 */
export function RSIChartDemo({ className }: RSIChartDemoProps) {
  // Hydration guard
  const isHydrated = useHydration()

  // Theme
  const { theme } = useTheme()

  // Chart refs
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const rsiSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const rsi70LineRef = useRef<ISeriesApi<'Line'> | null>(null)
  const rsi50LineRef = useRef<ISeriesApi<'Line'> | null>(null)
  const rsi30LineRef = useRef<ISeriesApi<'Line'> | null>(null)

  // State
  const [period, setPeriod] = useState(14)
  const [showZones, setShowZones] = useState(true)

  // Generate mock data
  const candlestickData = useMemo(() => {
    return generateMockStockData(100, 5000, 200) // 100 days, starting at 5000 IQD
  }, [])

  // Calculate RSI
  const rsiData = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateRSI(candlestickData, period)
  }, [candlestickData, period])

  // Create reference lines data (horizontal lines at 30, 50, 70)
  const createReferenceLine = (value: number): LineData[] => {
    if (!rsiData.length) return []
    return rsiData.map(point => ({
      time: point.time,
      value: value
    }))
  }

  const rsi70Data = useMemo(() => createReferenceLine(70), [rsiData])
  const rsi50Data = useMemo(() => createReferenceLine(50), [rsiData])
  const rsi30Data = useMemo(() => createReferenceLine(30), [rsiData])

  // Initialize chart
  useEffect(() => {
    if (!isHydrated || !chartContainerRef.current) return

    const container = chartContainerRef.current

    // Create chart
    const chart = createChart(container, {
      width: container.clientWidth,
      height: 500,
      layout: {
        background: { color: theme === 'dark' ? '#1e1e1e' : '#ffffff' },
        textColor: theme === 'dark' ? '#d1d5db' : '#374151',
      },
      grid: {
        vertLines: { color: theme === 'dark' ? '#2d2d2d' : '#e5e7eb' },
        horzLines: { color: theme === 'dark' ? '#2d2d2d' : '#e5e7eb' },
      },
      crosshair: {
        mode: 1, // Normal crosshair
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
      },
    })

    chartRef.current = chart

    // Add candlestick series (Pane 0)
    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#10b981',
      downColor: '#ef4444',
      borderUpColor: '#10b981',
      borderDownColor: '#ef4444',
      wickUpColor: '#10b981',
      wickDownColor: '#ef4444',
    })
    candlestickSeriesRef.current = candlestickSeries
    candlestickSeries.setData(candlestickData)

    // Add RSI series (Pane 1)
    const rsiSeries = chart.addSeries(LineSeries, {
      color: '#9c27b0',
      lineWidth: 2,
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
      title: `RSI(${period})`,
    }, 1) // Pane index 1
    rsiSeriesRef.current = rsiSeries
    rsiSeries.setData(rsiData)

    // Add reference lines (Pane 1)
    const rsi70Line = chart.addSeries(LineSeries, {
      color: showZones ? '#ef4444' : 'transparent',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceLineVisible: false,
      lastValueVisible: false,
      title: 'Overbought (70)',
    }, 1)
    rsi70LineRef.current = rsi70Line
    rsi70Line.setData(rsi70Data)

    const rsi50Line = chart.addSeries(LineSeries, {
      color: theme === 'dark' ? '#6b7280' : '#9ca3af',
      lineWidth: 1,
      lineStyle: LineStyle.Solid,
      priceLineVisible: false,
      lastValueVisible: false,
      title: 'Midpoint (50)',
    }, 1)
    rsi50LineRef.current = rsi50Line
    rsi50Line.setData(rsi50Data)

    const rsi30Line = chart.addSeries(LineSeries, {
      color: showZones ? '#10b981' : 'transparent',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceLineVisible: false,
      lastValueVisible: false,
      title: 'Oversold (30)',
    }, 1)
    rsi30LineRef.current = rsi30Line
    rsi30Line.setData(rsi30Data)

    // Set RSI pane height (25% of total)
    const panes = chart.panes()
    if (panes.length > 1) {
      const rsiPane = panes[1]
      const containerHeight = container.clientHeight || 500
      const maxHeight = Math.floor(containerHeight * 0.90) // 90% max
      const desiredHeight = Math.floor(containerHeight * 0.25) // 25% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      rsiPane.setHeight(constrainedHeight)
    }

    // Fit content
    chart.timeScale().fitContent()

    // Handle resize
    const handleResize = () => {
      if (chartRef.current && container) {
        chartRef.current.resize(container.clientWidth, 500)
      }
    }

    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)

      if (rsi30LineRef.current) {
        chart.removeSeries(rsi30LineRef.current)
        rsi30LineRef.current = null
      }
      if (rsi50LineRef.current) {
        chart.removeSeries(rsi50LineRef.current)
        rsi50LineRef.current = null
      }
      if (rsi70LineRef.current) {
        chart.removeSeries(rsi70LineRef.current)
        rsi70LineRef.current = null
      }
      if (rsiSeriesRef.current) {
        chart.removeSeries(rsiSeriesRef.current)
        rsiSeriesRef.current = null
      }
      if (candlestickSeriesRef.current) {
        chart.removeSeries(candlestickSeriesRef.current)
        candlestickSeriesRef.current = null
      }

      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, rsiData, period, showZones, rsi70Data, rsi50Data, rsi30Data])

  // Update zone visibility
  useEffect(() => {
    if (!rsi70LineRef.current || !rsi30LineRef.current) return

    rsi70LineRef.current.applyOptions({
      color: showZones ? '#ef4444' : 'transparent',
    })

    rsi30LineRef.current.applyOptions({
      color: showZones ? '#10b981' : 'transparent',
    })
  }, [showZones])

  // Loading state
  if (!isHydrated) {
    return (
      <Card className={className}>
        <div className="flex items-center justify-center p-12">
          <div className="text-center space-y-3">
            <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
            <p className="text-sm text-muted-foreground">Loading RSI chart demo...</p>
          </div>
        </div>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <div className="p-4 space-y-4">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-purple-500" />
            <h3 className="text-lg font-semibold">Interactive RSI Demo</h3>
          </div>
          <Badge variant="secondary">
            Period: {period} days
          </Badge>
        </div>

        {/* Controls */}
        <div className="flex flex-wrap gap-4 items-center">
          <div className="flex items-center gap-2">
            <Label className="text-sm font-medium">Period:</Label>
            <div className="flex gap-1">
              {[7, 14, 21, 28].map((p) => (
                <Button
                  key={p}
                  size="sm"
                  variant={period === p ? 'default' : 'outline'}
                  onClick={() => setPeriod(p)}
                  className="h-8 px-3"
                >
                  {p}
                </Button>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Label className="text-sm font-medium">Zones:</Label>
            <Button
              size="sm"
              variant={showZones ? 'default' : 'outline'}
              onClick={() => setShowZones(!showZones)}
              className="h-8 px-3"
            >
              {showZones ? 'Shown' : 'Hidden'}
            </Button>
          </div>
        </div>

        {/* Chart */}
        <div
          ref={chartContainerRef}
          className="w-full rounded-lg overflow-hidden border"
        />

        {/* Legend */}
        <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-purple-500" />
            <span>RSI Line</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-0.5 bg-red-500 border-dashed" />
            <span>Overbought (70)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-0.5 bg-gray-500" />
            <span>Midpoint (50)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-0.5 bg-green-500 border-dashed" />
            <span>Oversold (30)</span>
          </div>
        </div>

        {/* Instructions */}
        <div className="text-sm text-muted-foreground space-y-1">
          <p>• <strong>Adjust the period</strong> to see how RSI sensitivity changes</p>
          <p>• <strong>Toggle zones</strong> to highlight overbought/oversold areas</p>
          <p>• <strong>Hover over the chart</strong> to see exact RSI values at any point</p>
        </div>
      </div>
    </Card>
  )
}
