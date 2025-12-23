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
import { calculateATR } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function ATRChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const atrSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [period, setPeriod] = useState(14)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const atrData = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateATR(candlestickData, period)
  }, [candlestickData, period])

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

    const atrSeries = chart.addSeries(LineSeries, {
      color: '#f59e0b',
      lineWidth: 2,
      title: 'ATR',
      priceLineVisible: false,
    }, 1)
    atrSeriesRef.current = atrSeries
    atrSeries.setData(atrData)

    const panes = chart.panes()
    if (panes.length > 1) {
      const atrPane = panes[1]
      const containerHeight = container.clientHeight || 600
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.25)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      atrPane.setHeight(constrainedHeight)
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
      if (atrSeriesRef.current) chart.removeSeries(atrSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, atrData, period])

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
        <CardTitle>Average True Range Interactive Demo</CardTitle>
        <CardDescription>
          ATR measures volatility. Higher values = higher volatility, wider stop-losses needed.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={period === 7 ? 'default' : 'outline'} onClick={() => setPeriod(7)}>
              Short-Term (7)
            </Button>
            <Button size="sm" variant={period === 14 ? 'default' : 'outline'} onClick={() => setPeriod(14)}>
              Standard (14)
            </Button>
            <Button size="sm" variant={period === 21 ? 'default' : 'outline'} onClick={() => setPeriod(21)}>
              Long-Term (21)
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Usage:</strong> ATR doesn't indicate direction. It shows volatility magnitude. Use it to set
            appropriate stop-loss distances (e.g., 2× ATR) and position sizing based on risk.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">Low Volatility</p>
            <p className="text-sm font-semibold text-green-600">Narrow Range</p>
            <p className="text-xs mt-1">Tight stops OK</p>
          </div>
          <div className="text-center p-3 bg-orange-50 dark:bg-orange-950/20 rounded-lg border border-orange-200 dark:border-orange-800">
            <p className="text-xs text-muted-foreground">Normal Volatility</p>
            <p className="text-sm font-semibold text-orange-600">Average Range</p>
            <p className="text-xs mt-1">Standard stops</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">High Volatility</p>
            <p className="text-sm font-semibold text-red-600">Wide Range</p>
            <p className="text-xs mt-1">Wider stops needed</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
