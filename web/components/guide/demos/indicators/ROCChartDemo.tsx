'use client'

import { useState, useEffect, useRef, useMemo } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickSeries,
  LineSeries,
  LineStyle
} from 'lightweight-charts'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Info } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { calculateROC } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function ROCChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const rocSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const zeroLineRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [period, setPeriod] = useState(12)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const rocData = useMemo(() => {
    if (!candlestickData.length) return []

    const closes = candlestickData.map(d => d.close)
    const rocValues = calculateROC(closes, period)

    return candlestickData.map((d, i) => ({
      time: d.time,
      value: rocValues[i]
    })).filter(d => d.value !== null) as Array<{ time: string; value: number }>
  }, [candlestickData, period])

  // Generate zero line
  const zeroLine = useMemo(() => {
    if (!rocData.length) return []

    const firstTime = rocData[0].time
    const lastTime = rocData[rocData.length - 1].time

    return [
      { time: firstTime, value: 0 },
      { time: lastTime, value: 0 }
    ]
  }, [rocData])

  useEffect(() => {
    if (!isHydrated || !chartContainerRef.current) return

    const container = chartContainerRef.current
    const chart = createChart(container, {
      width: container.clientWidth,
      height: 600,
      layout: {
        background: { color: theme === 'dark' ? '#1e1e1e' : '#ffffff' },
        textColor: theme === 'dark' ? '#d1d5db' : '#374151',
      },
      grid: {
        vertLines: { color: theme === 'dark' ? '#2b2b43' : '#e5e7eb' },
        horzLines: { color: theme === 'dark' ? '#2b2b43' : '#e5e7eb' },
      },
      timeScale: {
        borderColor: theme === 'dark' ? '#2b2b43' : '#d1d5db',
        timeVisible: true,
      },
    })

    chartRef.current = chart

    // Main price chart (Pane 0)
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

    // ROC line (Pane 1)
    const rocSeries = chart.addSeries(LineSeries, {
      color: '#9c27b0',
      lineWidth: 2,
      title: 'ROC',
      priceLineVisible: false,
    }, 1)
    rocSeriesRef.current = rocSeries
    rocSeries.setData(rocData)

    // Zero line
    if (zeroLine.length > 0) {
      const zeroLineSeries = chart.addSeries(LineSeries, {
        color: theme === 'dark' ? '#6b7280' : '#9ca3af',
        lineWidth: 1,
        lineStyle: LineStyle.Dotted,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 1)
      zeroLineRef.current = zeroLineSeries
      zeroLineSeries.setData(zeroLine)
    }

    // Set ROC pane height
    const panes = chart.panes()
    if (panes.length > 1) {
      const rocPane = panes[1]
      const containerHeight = container.clientHeight || 600
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.30)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      rocPane.setHeight(constrainedHeight)
    }

    chart.timeScale().fitContent()

    const handleResize = () => {
      if (container && chart) {
        chart.applyOptions({ width: container.clientWidth })
      }
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      if (zeroLineRef.current) chart.removeSeries(zeroLineRef.current)
      if (rocSeriesRef.current) chart.removeSeries(rocSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, rocData, zeroLine, period])

  if (!isHydrated) {
    return (
      <Card className={className}>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center h-[600px]">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Rate of Change (ROC) Interactive Demo</CardTitle>
        <CardDescription>
          ROC &gt; 0 = price rising. ROC &lt; 0 = price falling. ROC % shows rate of change.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={period === 12 ? 'default' : 'outline'} onClick={() => setPeriod(12)}>
              Standard (12)
            </Button>
            <Button size="sm" variant={period === 21 ? 'default' : 'outline'} onClick={() => setPeriod(21)}>
              Medium (21)
            </Button>
            <Button size="sm" variant={period === 30 ? 'default' : 'outline'} onClick={() => setPeriod(30)}>
              Long (30)
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Trading Signals:</strong> ROC crossing above zero = upward momentum (buy signal). ROC crossing below zero = downward momentum (sell signal).
            Measures percentage rate of change over time.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">ROC &gt; 0</p>
            <p className="text-sm font-semibold text-green-600">Positive Change</p>
            <p className="text-xs mt-1">Price rising</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">ROC = 0</p>
            <p className="text-sm font-semibold text-blue-600">No Change</p>
            <p className="text-xs mt-1">Flat</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">ROC &lt; 0</p>
            <p className="text-sm font-semibold text-red-600">Negative Change</p>
            <p className="text-xs mt-1">Price falling</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
