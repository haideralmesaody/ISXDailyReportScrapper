'use client'

import { useState, useEffect, useRef, useMemo } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickSeries,
  LineSeries
} from 'lightweight-charts'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Info } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { calculateVWAP } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function VWAPChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const vwapSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [showVWAP, setShowVWAP] = useState(true)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const vwapData = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateVWAP(candlestickData)
  }, [candlestickData])

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

    if (showVWAP && vwapData.length > 0) {
      const vwapSeries = chart.addSeries(LineSeries, {
        color: '#2196F3',
        lineWidth: 2,
        title: 'Anchored VWAP',
        priceLineVisible: true,
      })
      vwapSeriesRef.current = vwapSeries
      vwapSeries.setData(vwapData)
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
      if (vwapSeriesRef.current) chart.removeSeries(vwapSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, vwapData, showVWAP])

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
        <CardTitle>Anchored VWAP Interactive Demo</CardTitle>
        <CardDescription>
          Cumulative volume-weighted average from chart start. Used for identifying major support/resistance zones in swing trading.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant={showVWAP ? 'default' : 'outline'}
              onClick={() => setShowVWAP(!showVWAP)}
            >
              {showVWAP ? 'Hide Anchored VWAP' : 'Show Anchored VWAP'}
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Anchored VWAP:</strong> Cumulative from chart start (does NOT reset daily). Price above = long-term bullish trend.
            Price below = long-term bearish trend. Used for major support/resistance zones in swing trading.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">Above Anchored VWAP</p>
            <p className="text-sm font-semibold text-green-600">Long-term Bullish</p>
            <p className="text-xs mt-1">Major resistance above</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">At Anchored VWAP</p>
            <p className="text-sm font-semibold text-blue-600">Decision Point</p>
            <p className="text-xs mt-1">Key S/R zone</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">Below Anchored VWAP</p>
            <p className="text-sm font-semibold text-red-600">Long-term Bearish</p>
            <p className="text-xs mt-1">Major support below</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
