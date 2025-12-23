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
  HistogramSeries,
  LineData,
  HistogramData
} from 'lightweight-charts'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Loader2, TrendingUp } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { calculateMACD } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'

interface MACDChartDemoProps {
  className?: string
}

/**
 * Interactive MACD chart demo with customizable period controls
 * Uses TradingView Lightweight Charts v5 with 2 panes (price + MACD)
 * MACD pane contains 3 series: MACD line, signal line, and histogram
 */
export function MACDChartDemo({ className }: MACDChartDemoProps) {
  // Hydration guard
  const isHydrated = useHydration()

  // Theme
  const { theme } = useTheme()

  // Chart refs
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const macdLineRef = useRef<ISeriesApi<'Line'> | null>(null)
  const signalLineRef = useRef<ISeriesApi<'Line'> | null>(null)
  const histogramRef = useRef<ISeriesApi<'Histogram'> | null>(null)

  // State
  const [fastPeriod, setFastPeriod] = useState(12)
  const [slowPeriod, setSlowPeriod] = useState(26)
  const [signalPeriod, setSignalPeriod] = useState(9)
  const [showHistogram, setShowHistogram] = useState(true)

  // Generate mock data
  const candlestickData = useMemo(() => {
    return generateMockStockData(100, 5000, 200) // 100 days, starting at 5000 IQD
  }, [])

  // Calculate MACD
  const macdData = useMemo(() => {
    if (!candlestickData.length) return { macd: [], signal: [], histogram: [] }
    return calculateMACD(candlestickData, fastPeriod, slowPeriod, signalPeriod)
  }, [candlestickData, fastPeriod, slowPeriod, signalPeriod])

  // Initialize chart
  useEffect(() => {
    if (!isHydrated || !chartContainerRef.current) return

    const container = chartContainerRef.current

    // Create chart
    const chart = createChart(container, {
      width: container.clientWidth,
      height: 550,
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

    // Add MACD line (Pane 1) - Blue
    const macdLine = chart.addSeries(LineSeries, {
      color: '#2196F3',
      lineWidth: 2,
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
      title: 'MACD',
    }, 1) // Pane index 1
    macdLineRef.current = macdLine
    macdLine.setData(macdData.macd)

    // Add Signal line (Pane 1) - Orange
    const signalLine = chart.addSeries(LineSeries, {
      color: '#FF9800',
      lineWidth: 2,
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
      title: 'Signal',
    }, 1)
    signalLineRef.current = signalLine
    signalLine.setData(macdData.signal)

    // Add Histogram (Pane 1) - Green/Red
    const histogram = chart.addSeries(HistogramSeries, {
      priceFormat: {
        type: 'price',
        precision: 2,
        minMove: 0.01,
      },
      title: 'Histogram',
    }, 1)
    histogramRef.current = histogram
    histogram.setData(macdData.histogram)

    // Set MACD pane height (30% of total)
    const panes = chart.panes()
    if (panes.length > 1) {
      const macdPane = panes[1]
      const containerHeight = container.clientHeight || 550
      const maxHeight = Math.floor(containerHeight * 0.90) // 90% max
      const desiredHeight = Math.floor(containerHeight * 0.30) // 30% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      macdPane.setHeight(constrainedHeight)
    }

    // Fit content
    chart.timeScale().fitContent()

    // Handle resize
    const handleResize = () => {
      if (chartRef.current && container) {
        chartRef.current.resize(container.clientWidth, 550)
      }
    }

    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)

      if (histogramRef.current) {
        chart.removeSeries(histogramRef.current)
        histogramRef.current = null
      }
      if (signalLineRef.current) {
        chart.removeSeries(signalLineRef.current)
        signalLineRef.current = null
      }
      if (macdLineRef.current) {
        chart.removeSeries(macdLineRef.current)
        macdLineRef.current = null
      }
      if (candlestickSeriesRef.current) {
        chart.removeSeries(candlestickSeriesRef.current)
        candlestickSeriesRef.current = null
      }

      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, macdData, fastPeriod, slowPeriod, signalPeriod])

  // Update histogram visibility
  useEffect(() => {
    if (!histogramRef.current) return

    histogramRef.current.applyOptions({
      visible: showHistogram,
    })
  }, [showHistogram])

  // Preset configurations
  const applyPreset = (preset: 'default' | 'fast' | 'slow') => {
    switch (preset) {
      case 'default':
        setFastPeriod(12)
        setSlowPeriod(26)
        setSignalPeriod(9)
        break
      case 'fast':
        setFastPeriod(8)
        setSlowPeriod(17)
        setSignalPeriod(9)
        break
      case 'slow':
        setFastPeriod(19)
        setSlowPeriod(39)
        setSignalPeriod(9)
        break
    }
  }

  // Loading state
  if (!isHydrated) {
    return (
      <Card className={className}>
        <div className="flex items-center justify-center p-12">
          <div className="text-center space-y-3">
            <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
            <p className="text-sm text-muted-foreground">Loading MACD chart demo...</p>
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
            <TrendingUp className="h-5 w-5 text-blue-500" />
            <h3 className="text-lg font-semibold">Interactive MACD Demo</h3>
          </div>
          <Badge variant="secondary">
            {fastPeriod}/{slowPeriod}/{signalPeriod}
          </Badge>
        </div>

        {/* Preset Controls */}
        <div className="flex items-center gap-2">
          <Label className="text-sm font-medium">Presets:</Label>
          <div className="flex gap-1">
            <Button
              size="sm"
              variant={fastPeriod === 8 ? 'default' : 'outline'}
              onClick={() => applyPreset('fast')}
              className="h-8 px-3"
            >
              Fast (8/17/9)
            </Button>
            <Button
              size="sm"
              variant={fastPeriod === 12 ? 'default' : 'outline'}
              onClick={() => applyPreset('default')}
              className="h-8 px-3"
            >
              Default (12/26/9)
            </Button>
            <Button
              size="sm"
              variant={fastPeriod === 19 ? 'default' : 'outline'}
              onClick={() => applyPreset('slow')}
              className="h-8 px-3"
            >
              Slow (19/39/9)
            </Button>
          </div>
        </div>

        {/* Custom Controls */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="space-y-2">
            <Label className="text-sm">Fast Period</Label>
            <Input
              type="number"
              min="5"
              max="20"
              value={fastPeriod}
              onChange={(e) => setFastPeriod(parseInt(e.target.value) || 12)}
              className="h-8"
            />
          </div>
          <div className="space-y-2">
            <Label className="text-sm">Slow Period</Label>
            <Input
              type="number"
              min="15"
              max="50"
              value={slowPeriod}
              onChange={(e) => setSlowPeriod(parseInt(e.target.value) || 26)}
              className="h-8"
            />
          </div>
          <div className="space-y-2">
            <Label className="text-sm">Signal Period</Label>
            <Input
              type="number"
              min="5"
              max="15"
              value={signalPeriod}
              onChange={(e) => setSignalPeriod(parseInt(e.target.value) || 9)}
              className="h-8"
            />
          </div>
          <div className="space-y-2">
            <Label className="text-sm">Display</Label>
            <Button
              size="sm"
              variant={showHistogram ? 'default' : 'outline'}
              onClick={() => setShowHistogram(!showHistogram)}
              className="h-8 w-full"
            >
              {showHistogram ? 'Histogram On' : 'Histogram Off'}
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
            <div className="w-3 h-3 rounded-full bg-blue-500" />
            <span>MACD Line</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-orange-500" />
            <span>Signal Line</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-green-500" />
            <span>Positive Histogram</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-red-500" />
            <span>Negative Histogram</span>
          </div>
        </div>

        {/* Instructions */}
        <div className="text-sm text-muted-foreground space-y-1">
          <p>• <strong>Bullish Crossover</strong>: MACD crosses above signal line (buy signal)</p>
          <p>• <strong>Bearish Crossover</strong>: MACD crosses below signal line (sell signal)</p>
          <p>• <strong>Histogram</strong>: Shows difference between MACD and signal (momentum strength)</p>
          <p>• <strong>Adjust periods</strong> to see how sensitivity changes</p>
        </div>
      </div>
    </Card>
  )
}
