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
import { calculateBollingerBands } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function BollingerBandsChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const upperBandSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const middleBandSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const lowerBandSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  // Parameters
  const [period, setPeriod] = useState(20)
  const [stdDev, setStdDev] = useState(2)
  const [customPeriod, setCustomPeriod] = useState('20')
  const [customStdDev, setCustomStdDev] = useState('2')

  // Generate sample ISX data
  const candlestickData = useMemo(() => {
    return generateMockStockData(150, 1200, 100) // 150 days, base price 1.200 IQD, volatility 100
  }, [])

  // Calculate Bollinger Bands
  const bollingerData = useMemo(() => {
    if (!candlestickData.length) {
      return { upper: [], middle: [], lower: [] }
    }
    return calculateBollingerBands(candlestickData, period, stdDev)
  }, [candlestickData, period, stdDev])

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

    // Add Upper Band (blue)
    const upperBandSeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: 'Upper Band',
      priceLineVisible: false,
    })
    upperBandSeriesRef.current = upperBandSeries
    upperBandSeries.setData(bollingerData.upper)

    // Add Middle Band (gray - SMA)
    const middleBandSeries = chart.addSeries(LineSeries, {
      color: '#9ca3af',
      lineWidth: 2,
      title: 'Middle Band (SMA)',
      priceLineVisible: false,
      lineStyle: LineStyle.Dashed,
    })
    middleBandSeriesRef.current = middleBandSeries
    middleBandSeries.setData(bollingerData.middle)

    // Add Lower Band (blue)
    const lowerBandSeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: 'Lower Band',
      priceLineVisible: false,
    })
    lowerBandSeriesRef.current = lowerBandSeries
    lowerBandSeries.setData(bollingerData.lower)

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
      if (upperBandSeriesRef.current) chart.removeSeries(upperBandSeriesRef.current)
      if (middleBandSeriesRef.current) chart.removeSeries(middleBandSeriesRef.current)
      if (lowerBandSeriesRef.current) chart.removeSeries(lowerBandSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, bollingerData])

  // Update chart when parameters change
  useEffect(() => {
    if (!chartRef.current || !isHydrated) return

    if (upperBandSeriesRef.current) {
      upperBandSeriesRef.current.setData(bollingerData.upper)
    }
    if (middleBandSeriesRef.current) {
      middleBandSeriesRef.current.setData(bollingerData.middle)
    }
    if (lowerBandSeriesRef.current) {
      lowerBandSeriesRef.current.setData(bollingerData.lower)
    }

    chartRef.current.timeScale().fitContent()
  }, [bollingerData, isHydrated])

  const handlePresetChange = (newPeriod: number, newStdDev: number) => {
    setPeriod(newPeriod)
    setStdDev(newStdDev)
    setCustomPeriod(String(newPeriod))
    setCustomStdDev(String(newStdDev))
  }

  const handleCustomApply = () => {
    const newPeriod = parseInt(customPeriod)
    const newStdDev = parseFloat(customStdDev)
    if (newPeriod >= 2 && newPeriod <= 200 && newStdDev >= 0.5 && newStdDev <= 5) {
      setPeriod(newPeriod)
      setStdDev(newStdDev)
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
        <CardTitle>Bollinger Bands Interactive Demo</CardTitle>
        <CardDescription>
          Adjust period and standard deviation to see how Bollinger Bands adapt to volatility changes
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Preset Buttons */}
        <div className="space-y-2">
          <Label>Standard Presets</Label>
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant={period === 20 && stdDev === 2 ? 'default' : 'outline'}
              onClick={() => handlePresetChange(20, 2)}
            >
              Default (20, 2)
            </Button>
            <Button
              size="sm"
              variant={period === 20 && stdDev === 1 ? 'default' : 'outline'}
              onClick={() => handlePresetChange(20, 1)}
            >
              Tight (20, 1)
            </Button>
            <Button
              size="sm"
              variant={period === 20 && stdDev === 3 ? 'default' : 'outline'}
              onClick={() => handlePresetChange(20, 3)}
            >
              Wide (20, 3)
            </Button>
            <Button
              size="sm"
              variant={period === 10 && stdDev === 2 ? 'default' : 'outline'}
              onClick={() => handlePresetChange(10, 2)}
            >
              Short-Term (10, 2)
            </Button>
            <Button
              size="sm"
              variant={period === 50 && stdDev === 2 ? 'default' : 'outline'}
              onClick={() => handlePresetChange(50, 2)}
            >
              Long-Term (50, 2)
            </Button>
          </div>
        </div>

        {/* Custom Inputs */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="space-y-2">
            <Label htmlFor="custom-period">
              Period: {period}
              <span className="text-xs text-muted-foreground ml-2">(2-200)</span>
            </Label>
            <Input
              id="custom-period"
              type="number"
              min="2"
              max="200"
              value={customPeriod}
              onChange={(e) => setCustomPeriod(e.target.value)}
              placeholder="20"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="custom-stddev">
              Std Deviation: {stdDev}
              <span className="text-xs text-muted-foreground ml-2">(0.5-5)</span>
            </Label>
            <Input
              id="custom-stddev"
              type="number"
              min="0.5"
              max="5"
              step="0.1"
              value={customStdDev}
              onChange={(e) => setCustomStdDev(e.target.value)}
              placeholder="2"
            />
          </div>

          <div className="flex items-end">
            <Button onClick={handleCustomApply} className="w-full">
              Apply Custom
            </Button>
          </div>
        </div>

        {/* Info Alert */}
        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Current Settings:</strong> {period}-period SMA ± {stdDev} standard deviations.
            The bands expand during high volatility and contract during low volatility.
          </AlertDescription>
        </Alert>

        {/* Chart */}
        <div ref={chartContainerRef} className="w-full h-[550px] mt-4" />

        {/* Band Width Info */}
        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">Upper Band</p>
            <p className="text-sm font-semibold text-blue-600">SMA + {stdDev}σ</p>
          </div>
          <div className="text-center p-3 bg-gray-50 dark:bg-gray-900/20 rounded-lg border border-gray-200 dark:border-gray-800">
            <p className="text-xs text-muted-foreground">Middle Band</p>
            <p className="text-sm font-semibold text-gray-600">SMA {period}</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">Lower Band</p>
            <p className="text-sm font-semibold text-blue-600">SMA - {stdDev}σ</p>
          </div>
        </div>

        {/* Legend */}
        <div className="flex flex-wrap gap-4 justify-center text-sm pt-2 border-t">
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-blue-500" />
            <span>Upper/Lower Bands</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-gray-400 border-dashed border-t-2" />
            <span>Middle Band (SMA)</span>
          </div>
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
