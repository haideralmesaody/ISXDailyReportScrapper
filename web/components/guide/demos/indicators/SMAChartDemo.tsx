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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Loader2, Info } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { calculateSMA } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Checkbox } from '@/components/ui/checkbox'

export function SMAChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const sma20SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const sma50SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const sma200SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  // Which SMAs to show
  const [showSMA20, setShowSMA20] = useState(true)
  const [showSMA50, setShowSMA50] = useState(true)
  const [showSMA200, setShowSMA200] = useState(true)

  // Generate sample ISX data (250 days for SMA200)
  const candlestickData = useMemo(() => {
    return generateMockStockData(250, 1200, 100) // 250 days, base price 1.200 IQD, volatility 100
  }, [])

  // Calculate SMAs
  const sma20Data = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateSMA(candlestickData, 20)
  }, [candlestickData])

  const sma50Data = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateSMA(candlestickData, 50)
  }, [candlestickData])

  const sma200Data = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateSMA(candlestickData, 200)
  }, [candlestickData])

  // Initialize chart
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

    // Add candlestick series
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

    // Add SMA 20 (blue)
    const sma20Series = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: 'SMA 20',
      priceLineVisible: false,
      visible: showSMA20,
    })
    sma20SeriesRef.current = sma20Series
    sma20Series.setData(sma20Data)

    // Add SMA 50 (orange)
    const sma50Series = chart.addSeries(LineSeries, {
      color: '#f59e0b',
      lineWidth: 2,
      title: 'SMA 50',
      priceLineVisible: false,
      visible: showSMA50,
    })
    sma50SeriesRef.current = sma50Series
    sma50Series.setData(sma50Data)

    // Add SMA 200 (red)
    const sma200Series = chart.addSeries(LineSeries, {
      color: '#ef4444',
      lineWidth: 3,
      title: 'SMA 200',
      priceLineVisible: false,
      visible: showSMA200,
    })
    sma200SeriesRef.current = sma200Series
    sma200Series.setData(sma200Data)

    // Auto-fit content
    chart.timeScale().fitContent()

    // Handle resize
    const handleResize = () => {
      if (container && chart) {
        chart.applyOptions({ width: container.clientWidth })
      }
    }
    window.addEventListener('resize', handleResize)

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize)
      if (sma20SeriesRef.current) chart.removeSeries(sma20SeriesRef.current)
      if (sma50SeriesRef.current) chart.removeSeries(sma50SeriesRef.current)
      if (sma200SeriesRef.current) chart.removeSeries(sma200SeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, sma20Data, sma50Data, sma200Data, showSMA20, showSMA50, showSMA200])

  // Update visibility when checkboxes change
  useEffect(() => {
    if (!chartRef.current || !isHydrated) return

    if (sma20SeriesRef.current) {
      sma20SeriesRef.current.applyOptions({ visible: showSMA20 })
    }
    if (sma50SeriesRef.current) {
      sma50SeriesRef.current.applyOptions({ visible: showSMA50 })
    }
    if (sma200SeriesRef.current) {
      sma200SeriesRef.current.applyOptions({ visible: showSMA200 })
    }
  }, [showSMA20, showSMA50, showSMA200, isHydrated])

  const handlePresetChange = (preset: 'all' | 'short' | 'long' | 'trend') => {
    switch (preset) {
      case 'all':
        setShowSMA20(true)
        setShowSMA50(true)
        setShowSMA200(true)
        break
      case 'short':
        setShowSMA20(true)
        setShowSMA50(false)
        setShowSMA200(false)
        break
      case 'long':
        setShowSMA20(false)
        setShowSMA50(false)
        setShowSMA200(true)
        break
      case 'trend':
        setShowSMA20(false)
        setShowSMA50(true)
        setShowSMA200(true)
        break
    }
  }

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
        <CardTitle>Simple Moving Average Interactive Demo</CardTitle>
        <CardDescription>
          Toggle different SMA periods to see how they smooth price action and identify trends
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Preset Buttons */}
        <div className="space-y-2">
          <Label>Standard Presets</Label>
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant={showSMA20 && showSMA50 && showSMA200 ? 'default' : 'outline'}
              onClick={() => handlePresetChange('all')}
            >
              All SMAs (20, 50, 200)
            </Button>
            <Button
              size="sm"
              variant={showSMA20 && !showSMA50 && !showSMA200 ? 'default' : 'outline'}
              onClick={() => handlePresetChange('short')}
            >
              Short-Term (20)
            </Button>
            <Button
              size="sm"
              variant={!showSMA20 && !showSMA50 && showSMA200 ? 'default' : 'outline'}
              onClick={() => handlePresetChange('long')}
            >
              Long-Term (200)
            </Button>
            <Button
              size="sm"
              variant={!showSMA20 && showSMA50 && showSMA200 ? 'default' : 'outline'}
              onClick={() => handlePresetChange('trend')}
            >
              Trend Analysis (50, 200)
            </Button>
          </div>
        </div>

        {/* Individual Toggles */}
        <div className="space-y-2">
          <Label>Toggle Individual SMAs</Label>
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="sma20"
                checked={showSMA20}
                onCheckedChange={(checked) => setShowSMA20(!!checked)}
              />
              <label
                htmlFor="sma20"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                <span className="inline-block w-8 h-0.5 bg-blue-500 mr-2 align-middle"></span>
                SMA 20 (Short-term)
              </label>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="sma50"
                checked={showSMA50}
                onCheckedChange={(checked) => setShowSMA50(!!checked)}
              />
              <label
                htmlFor="sma50"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                <span className="inline-block w-8 h-0.5 bg-orange-500 mr-2 align-middle"></span>
                SMA 50 (Medium-term)
              </label>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox
                id="sma200"
                checked={showSMA200}
                onCheckedChange={(checked) => setShowSMA200(!!checked)}
              />
              <label
                htmlFor="sma200"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                <span className="inline-block w-8 h-0.5 bg-red-500 mr-2 align-middle"></span>
                SMA 200 (Long-term)
              </label>
            </div>
          </div>
        </div>

        {/* Info Alert */}
        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Observation:</strong> Notice how shorter periods (SMA 20) react faster to price changes,
            while longer periods (SMA 200) provide smoother, more stable trend lines. The SMA 200 is considered
            the long-term trend indicator by many professional traders.
          </AlertDescription>
        </Alert>

        {/* Chart */}
        <div ref={chartContainerRef} className="w-full h-[550px] mt-4" />

        {/* SMA Info Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2">
          <div className={`text-center p-3 rounded-lg border transition-opacity ${
            showSMA20
              ? 'bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 opacity-100'
              : 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800 opacity-50'
          }`}>
            <p className="text-xs text-muted-foreground">Short-Term Trend</p>
            <p className="text-sm font-semibold text-blue-600">SMA 20</p>
            <p className="text-xs mt-1">Day traders, swing traders</p>
          </div>

          <div className={`text-center p-3 rounded-lg border transition-opacity ${
            showSMA50
              ? 'bg-orange-50 dark:bg-orange-950/20 border-orange-200 dark:border-orange-800 opacity-100'
              : 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800 opacity-50'
          }`}>
            <p className="text-xs text-muted-foreground">Medium-Term Trend</p>
            <p className="text-sm font-semibold text-orange-600">SMA 50</p>
            <p className="text-xs mt-1">Swing traders, position traders</p>
          </div>

          <div className={`text-center p-3 rounded-lg border transition-opacity ${
            showSMA200
              ? 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-800 opacity-100'
              : 'bg-gray-50 dark:bg-gray-900/20 border-gray-200 dark:border-gray-800 opacity-50'
          }`}>
            <p className="text-xs text-muted-foreground">Long-Term Trend</p>
            <p className="text-sm font-semibold text-red-600">SMA 200</p>
            <p className="text-xs mt-1">Position traders, investors</p>
          </div>
        </div>

        {/* Golden/Death Cross Info */}
        {showSMA50 && showSMA200 && (
          <Alert className="bg-yellow-50 dark:bg-yellow-950/20 border-yellow-200 dark:border-yellow-800">
            <AlertDescription className="text-yellow-800 dark:text-yellow-200">
              <strong>Golden Cross:</strong> When SMA 50 crosses above SMA 200 (bullish signal).
              <br />
              <strong>Death Cross:</strong> When SMA 50 crosses below SMA 200 (bearish signal).
            </AlertDescription>
          </Alert>
        )}

        {/* Legend */}
        <div className="flex flex-wrap gap-4 justify-center text-sm pt-2 border-t">
          {showSMA20 && (
            <div className="flex items-center gap-2">
              <div className="w-4 h-0.5 bg-blue-500" />
              <span>SMA 20</span>
            </div>
          )}
          {showSMA50 && (
            <div className="flex items-center gap-2">
              <div className="w-4 h-0.5 bg-orange-500" />
              <span>SMA 50</span>
            </div>
          )}
          {showSMA200 && (
            <div className="flex items-center gap-2">
              <div className="w-4 h-0.5 bg-red-500" />
              <span>SMA 200</span>
            </div>
          )}
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-green-500 border border-green-600" />
            <span>Bullish Candle</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-red-500 border border-red-600" />
            <span>Bearish Candle</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
