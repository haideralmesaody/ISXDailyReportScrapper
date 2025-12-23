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
import { calculateEMA, calculateSMA } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Checkbox } from '@/components/ui/checkbox'

export function EMAChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const ema20SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const sma20SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [showEMA, setShowEMA] = useState(true)
  const [showSMA, setShowSMA] = useState(true)
  const [period, setPeriod] = useState(20)

  const candlestickData = useMemo(() => {
    return generateMockStockData(150, 1200, 120)
  }, [])

  const emaData = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateEMA(candlestickData, period)
  }, [candlestickData, period])

  const smaData = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateSMA(candlestickData, period)
  }, [candlestickData, period])

  useEffect(() => {
    if (!isHydrated || !chartContainerRef.current) return

    const container = chartContainerRef.current
    const chart = createChart(container, {
      width: container.clientWidth,
      height: 550,
      layout: {
        background: { color: theme === 'dark' ? '#1e1e1e' : '#ffffff' },
        textColor: theme === 'dark' ? '#d1d5db' : '#374151',
      },
      grid: {
        vertLines: { color: theme === 'dark' ? '#2b2b43' : '#e5e7eb' },
        horzLines: { color: theme === 'dark' ? '#2b2b43' : '#e5e7eb' },
      },
      rightPriceScale: {
        borderColor: theme === 'dark' ? '#2b2b43' : '#d1d5db',
      },
      timeScale: {
        borderColor: theme === 'dark' ? '#2b2b43' : '#d1d5db',
        timeVisible: true,
      },
      crosshair: {
        mode: 1,
        vertLine: {
          width: 1,
          color: theme === 'dark' ? '#6366f1' : '#4f46e5',
          style: LineStyle.Dashed,
        },
        horzLine: {
          width: 1,
          color: theme === 'dark' ? '#6366f1' : '#4f46e5',
          style: LineStyle.Dashed,
        },
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

    const emaSeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: 'EMA ' + period,
      priceLineVisible: false,
      visible: showEMA,
    })
    ema20SeriesRef.current = emaSeries
    emaSeries.setData(emaData)

    const smaSeries = chart.addSeries(LineSeries, {
      color: '#f59e0b',
      lineWidth: 2,
      title: 'SMA ' + period,
      priceLineVisible: false,
      visible: showSMA,
      lineStyle: LineStyle.Dashed,
    })
    sma20SeriesRef.current = smaSeries
    smaSeries.setData(smaData)

    chart.timeScale().fitContent()

    const handleResize = () => {
      if (container && chart) {
        chart.applyOptions({ width: container.clientWidth })
      }
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      if (ema20SeriesRef.current) chart.removeSeries(ema20SeriesRef.current)
      if (sma20SeriesRef.current) chart.removeSeries(sma20SeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, emaData, smaData, period, showEMA, showSMA])

  useEffect(() => {
    if (!chartRef.current || !isHydrated) return

    if (ema20SeriesRef.current) {
      ema20SeriesRef.current.applyOptions({ visible: showEMA })
    }
    if (sma20SeriesRef.current) {
      sma20SeriesRef.current.applyOptions({ visible: showSMA })
    }
  }, [showEMA, showSMA, isHydrated])

  if (!isHydrated) {
    return (
      <Card className={className}>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center h-[550px]">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Exponential Moving Average Interactive Demo</CardTitle>
        <CardDescription>
          Compare EMA vs SMA to see how EMA reacts faster to price changes
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant={period === 12 ? 'default' : 'outline'}
              onClick={() => setPeriod(12)}
            >
              Fast (12)
            </Button>
            <Button
              size="sm"
              variant={period === 20 ? 'default' : 'outline'}
              onClick={() => setPeriod(20)}
            >
              Standard (20)
            </Button>
            <Button
              size="sm"
              variant={period === 50 ? 'default' : 'outline'}
              onClick={() => setPeriod(50)}
            >
              Slow (50)
            </Button>
          </div>
        </div>

        <div className="flex gap-4">
          <div className="flex items-center space-x-2">
            <Checkbox
              id="ema"
              checked={showEMA}
              onCheckedChange={(checked) => setShowEMA(!!checked)}
            />
            <label htmlFor="ema" className="text-sm font-medium cursor-pointer">
              <span className="inline-block w-8 h-0.5 bg-blue-500 mr-2 align-middle"></span>
              EMA {period}
            </label>
          </div>

          <div className="flex items-center space-x-2">
            <Checkbox
              id="sma"
              checked={showSMA}
              onCheckedChange={(checked) => setShowSMA(!!checked)}
            />
            <label htmlFor="sma" className="text-sm font-medium cursor-pointer">
              <span className="inline-block w-8 h-0.5 bg-orange-500 border-dashed border-t-2 mr-2 align-middle"></span>
              SMA {period} (comparison)
            </label>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Observation:</strong> Notice how EMA (solid blue) reacts faster to price changes
            than SMA (dashed orange). EMA gives more weight to recent prices, making it more responsive
            but also more prone to whipsaws in volatile markets.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[550px] mt-4" />

        <div className="grid grid-cols-2 gap-4 pt-2">
          <div className={`text-center p-3 rounded-lg border transition-opacity ${
            showEMA
              ? 'bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 opacity-100'
              : 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800 opacity-50'
          }`}>
            <p className="text-xs text-muted-foreground">Exponential MA</p>
            <p className="text-sm font-semibold text-blue-600">EMA {period}</p>
            <p className="text-xs mt-1">More weight on recent prices</p>
          </div>

          <div className={`text-center p-3 rounded-lg border transition-opacity ${
            showSMA
              ? 'bg-orange-50 dark:bg-orange-950/20 border-orange-200 dark:border-orange-800 opacity-100'
              : 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800 opacity-50'
          }`}>
            <p className="text-xs text-muted-foreground">Simple MA</p>
            <p className="text-sm font-semibold text-orange-600">SMA {period}</p>
            <p className="text-xs mt-1">Equal weight all prices</p>
          </div>
        </div>

        <div className="flex flex-wrap gap-4 justify-center text-sm pt-2 border-t">
          {showEMA && (
            <div className="flex items-center gap-2">
              <div className="w-4 h-0.5 bg-blue-500" />
              <span>EMA (Faster)</span>
            </div>
          )}
          {showSMA && (
            <div className="flex items-center gap-2">
              <div className="w-4 h-0.5 bg-orange-500 border-dashed border-t-2" />
              <span>SMA (Smoother)</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
