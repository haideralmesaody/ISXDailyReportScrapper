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
import { calculateMomentum } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function MomentumChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const momentumSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const zeroLineRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [period, setPeriod] = useState(10)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const momentumData = useMemo(() => {
    if (!candlestickData.length) return []

    const closes = candlestickData.map(d => d.close)
    const momentumValues = calculateMomentum(closes, period)

    return candlestickData.map((d, i) => ({
      time: d.time,
      value: momentumValues[i]
    })).filter(d => d.value !== null) as Array<{ time: string; value: number }>
  }, [candlestickData, period])

  // Generate zero line
  const zeroLine = useMemo(() => {
    if (!momentumData.length) return []

    const firstTime = momentumData[0].time
    const lastTime = momentumData[momentumData.length - 1].time

    return [
      { time: firstTime, value: 0 },
      { time: lastTime, value: 0 }
    ]
  }, [momentumData])

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

    // Momentum line (Pane 1)
    const momentumSeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: 'Momentum',
      priceLineVisible: false,
    }, 1)
    momentumSeriesRef.current = momentumSeries
    momentumSeries.setData(momentumData)

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

    // Set Momentum pane height
    const panes = chart.panes()
    if (panes.length > 1) {
      const momentumPane = panes[1]
      const containerHeight = container.clientHeight || 600
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.30)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      momentumPane.setHeight(constrainedHeight)
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
      if (momentumSeriesRef.current) chart.removeSeries(momentumSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, momentumData, zeroLine, period])

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
        <CardTitle>Momentum Oscillator Interactive Demo</CardTitle>
        <CardDescription>
          MOM &gt; 0 = bullish. MOM &lt; 0 = bearish. MOM crossing zero = momentum shift.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={period === 10 ? 'default' : 'outline'} onClick={() => setPeriod(10)}>
              Standard (10)
            </Button>
            <Button size="sm" variant={period === 14 ? 'default' : 'outline'} onClick={() => setPeriod(14)}>
              Medium (14)
            </Button>
            <Button size="sm" variant={period === 20 ? 'default' : 'outline'} onClick={() => setPeriod(20)}>
              Long (20)
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Trading Signals:</strong> Momentum crossing above zero = bullish momentum (buy signal). Momentum crossing below zero = bearish momentum (sell signal).
            Measures absolute rate of price change.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">MOM &gt; 0</p>
            <p className="text-sm font-semibold text-green-600">Bullish Momentum</p>
            <p className="text-xs mt-1">Upward pressure</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">MOM = 0</p>
            <p className="text-sm font-semibold text-blue-600">No Momentum</p>
            <p className="text-xs mt-1">Neutral</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">MOM &lt; 0</p>
            <p className="text-sm font-semibold text-red-600">Bearish Momentum</p>
            <p className="text-xs mt-1">Downward pressure</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
