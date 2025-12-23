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
import { calculateADX } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function ADXChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const adxSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const plusDISeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const minusDISeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const thresholdLineRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [period, setPeriod] = useState(14)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const adxData = useMemo(() => {
    if (!candlestickData.length) return { adx: [], plusDI: [], minusDI: [] }

    const highs = candlestickData.map(d => d.high)
    const lows = candlestickData.map(d => d.low)
    const closes = candlestickData.map(d => d.close)

    const result = calculateADX(highs, lows, closes, period)

    const adx = candlestickData.map((d, i) => ({
      time: d.time,
      value: result.adx[i]
    })).filter(d => d.value !== null) as Array<{ time: string; value: number }>

    const plusDI = candlestickData.map((d, i) => ({
      time: d.time,
      value: result.plusDI[i]
    })).filter(d => d.value !== null) as Array<{ time: string; value: number }>

    const minusDI = candlestickData.map((d, i) => ({
      time: d.time,
      value: result.minusDI[i]
    })).filter(d => d.value !== null) as Array<{ time: string; value: number }>

    return { adx, plusDI, minusDI }
  }, [candlestickData, period])

  // Generate reference line (25 - trend threshold)
  const thresholdLine = useMemo(() => {
    if (!adxData.adx.length) return []

    const firstTime = adxData.adx[0].time
    const lastTime = adxData.adx[adxData.adx.length - 1].time

    return [
      { time: firstTime, value: 25 },
      { time: lastTime, value: 25 }
    ]
  }, [adxData.adx])

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

    // ADX line (Pane 1) - Green
    const adxSeries = chart.addSeries(LineSeries, {
      color: '#10b981',
      lineWidth: 2,
      title: 'ADX',
      priceLineVisible: false,
    }, 1)
    adxSeriesRef.current = adxSeries
    adxSeries.setData(adxData.adx)

    // +DI line (Pane 1) - Blue
    const plusDISeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: '+DI',
      priceLineVisible: false,
    }, 1)
    plusDISeriesRef.current = plusDISeries
    plusDISeries.setData(adxData.plusDI)

    // -DI line (Pane 1) - Red
    const minusDISeries = chart.addSeries(LineSeries, {
      color: '#ef4444',
      lineWidth: 2,
      title: '-DI',
      priceLineVisible: false,
    }, 1)
    minusDISeriesRef.current = minusDISeries
    minusDISeries.setData(adxData.minusDI)

    // Threshold line at 25
    if (thresholdLine.length > 0) {
      const threshold = chart.addSeries(LineSeries, {
        color: theme === 'dark' ? '#6b7280' : '#9ca3af',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 1)
      thresholdLineRef.current = threshold
      threshold.setData(thresholdLine)
    }

    // Set ADX pane height
    const panes = chart.panes()
    if (panes.length > 1) {
      const adxPane = panes[1]
      const containerHeight = container.clientHeight || 600
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.30)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      adxPane.setHeight(constrainedHeight)
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
      if (thresholdLineRef.current) chart.removeSeries(thresholdLineRef.current)
      if (minusDISeriesRef.current) chart.removeSeries(minusDISeriesRef.current)
      if (plusDISeriesRef.current) chart.removeSeries(plusDISeriesRef.current)
      if (adxSeriesRef.current) chart.removeSeries(adxSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, adxData, thresholdLine, period])

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
        <CardTitle>Average Directional Index (ADX) Interactive Demo</CardTitle>
        <CardDescription>
          ADX &gt; 25 = strong trend. ADX &lt; 20 = ranging market. +DI/-DI show direction.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={period === 14 ? 'default' : 'outline'} onClick={() => setPeriod(14)}>
              Standard (14)
            </Button>
            <Button size="sm" variant={period === 21 ? 'default' : 'outline'} onClick={() => setPeriod(21)}>
              Medium (21)
            </Button>
            <Button size="sm" variant={period === 28 ? 'default' : 'outline'} onClick={() => setPeriod(28)}>
              Long (28)
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Trading Signals:</strong> ADX above 25 indicates strong trend. +DI above -DI = uptrend. -DI above +DI = downtrend.
            ADX measures trend strength, not direction.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">ADX &gt; 25</p>
            <p className="text-sm font-semibold text-green-600">Strong Trend</p>
            <p className="text-xs mt-1">Trending market</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">ADX 20-25</p>
            <p className="text-sm font-semibold text-blue-600">Developing</p>
            <p className="text-xs mt-1">Trend forming</p>
          </div>
          <div className="text-center p-3 bg-yellow-50 dark:bg-yellow-950/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
            <p className="text-xs text-muted-foreground">ADX &lt; 20</p>
            <p className="text-sm font-semibold text-yellow-600">Weak/Ranging</p>
            <p className="text-xs mt-1">No clear trend</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
