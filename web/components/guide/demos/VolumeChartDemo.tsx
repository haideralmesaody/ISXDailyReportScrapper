'use client'

import { useState, useEffect, useRef, useMemo, useCallback } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickSeries,
  HistogramSeries,
  CandlestickData,
  HistogramData,
  Time
} from 'lightweight-charts'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Loader2, Info } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { useToast } from '@/lib/hooks/use-toast'
import { fetchTickerHistory } from '@/lib/api/analysis'
import type { TickerHistoricalData } from '@/types/analysis'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { calculateSMA } from '@/lib/indicators/calculations'

// Predefined tickers for the guide (same as LiveChartDemo)
const GUIDE_TICKERS = [
  { symbol: 'BBOB', name: 'Bank of Baghdad' },
  { symbol: 'TASC', name: 'Asia Cell Telecommunication' },
  { symbol: 'BNOI', name: 'National Bank of Iraq' },
  { symbol: 'BMNS', name: 'Al -Mansour Bank' },
  { symbol: 'IBSD', name: 'Baghdad Soft Drinks' }
]

export function VolumeChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()
  const { toast } = useToast()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)

  const [showVolume, setShowVolume] = useState(true)
  const [selectedTicker, setSelectedTicker] = useState(GUIDE_TICKERS[0].symbol)
  const [tickerData, setTickerData] = useState<TickerHistoricalData[]>([])
  const [isLoadingData, setIsLoadingData] = useState(true)
  const [currentPrice, setCurrentPrice] = useState({ price: 0, change: 0, changePercent: 0 })

  // Convert TickerHistoricalData to TradingView format
  const candlestickData = useMemo<CandlestickData[]>(() => {
    if (tickerData.length === 0) return []
    return tickerData.map(d => ({
      time: d.date as Time,
      open: d.open,
      high: d.high,
      low: d.low,
      close: d.close
    }))
  }, [tickerData])

  // Calculate volume data with intensity-based coloring
  const volumeData = useMemo<HistogramData[]>(() => {
    if (tickerData.length === 0) return []

    // Calculate 20-period volume MA for intensity calculation
    const volumes = tickerData.map(d => d.volume)
    const volumeMA = calculateSMA(volumes, 20)

    return tickerData.map((d, i) => {
      const volume = d.volume
      const ma = volumeMA[i]

      // Determine color based on price direction (close vs open)
      const isGreen = d.close >= d.open
      const baseColor = isGreen ? '#10b981' : '#ef4444'

      // Calculate intensity ratio (volume / MA)
      let intensity = 0.5 // default medium
      if (ma && ma > 0) {
        const ratio = volume / ma
        if (ratio > 1.5) intensity = 1.0        // Very high: >150% of MA
        else if (ratio > 1.0) intensity = 0.8   // High: 100-150% of MA
        else if (ratio > 0.5) intensity = 0.5   // Normal: 50-100% of MA
        else if (ratio > 0.25) intensity = 0.3  // Low: 25-50% of MA
        else intensity = 0.15                   // Very low: <25% of MA
      }

      return {
        time: d.date as Time,
        value: volume,
        color: `${baseColor}${Math.round(intensity * 255).toString(16).padStart(2, '0')}`
      }
    })
  }, [tickerData])

  // Fetch real ticker data when ticker changes
  useEffect(() => {
    if (!isHydrated || !selectedTicker) return

    const loadTickerData = async () => {
      setIsLoadingData(true)
      try {
        const history = await fetchTickerHistory(selectedTicker)
        setTickerData(history)

        // Calculate current price and change from latest data
        if (history.length > 0) {
          const latest = history[history.length - 1]
          setCurrentPrice({
            price: latest.close,
            change: latest.change || 0,
            changePercent: latest.changePercent || 0
          })
        }
      } catch (err) {
        console.error('Failed to load ticker data:', err)
        toast({
          title: 'Error loading data',
          description: `Could not load data for ${selectedTicker}. Please try again.`,
          variant: 'destructive'
        })
      } finally {
        setIsLoadingData(false)
      }
    }

    loadTickerData()
  }, [isHydrated, selectedTicker, toast])

  useEffect(() => {
    if (!isHydrated || !chartContainerRef.current || candlestickData.length === 0) return

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

    // Volume histogram (Pane 1)
    if (showVolume) {
      const volumeSeries = chart.addSeries(HistogramSeries, {
        priceFormat: {
          type: 'volume',
        },
        priceScaleId: 'volume',
      }, 1)
      volumeSeriesRef.current = volumeSeries
      volumeSeries.setData(volumeData)

      // Set volume pane height
      const panes = chart.panes()
      if (panes.length > 1) {
        const volumePane = panes[1]
        const containerHeight = container.clientHeight || 600
        const maxHeight = Math.floor(containerHeight * 0.90)
        const desiredHeight = Math.floor(containerHeight * 0.25)
        const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
        volumePane.setHeight(constrainedHeight)
      }
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
      if (volumeSeriesRef.current) {
        try {
          chart.removeSeries(volumeSeriesRef.current)
        } catch (err) {
          console.warn('Failed to remove volume series:', err)
        }
        volumeSeriesRef.current = null
      }
      if (candlestickSeriesRef.current) {
        try {
          chart.removeSeries(candlestickSeriesRef.current)
        } catch (err) {
          console.warn('Failed to remove candlestick series:', err)
        }
        candlestickSeriesRef.current = null
      }
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, volumeData, showVolume])

  if (!isHydrated || isLoadingData) {
    return (
      <Card className={className}>
        <CardContent className="pt-6">
          <div className="flex items-center justify-center h-[600px]">
            <div className="text-center space-y-3">
              <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
              <p className="text-sm text-muted-foreground">
                {isLoadingData ? 'Loading market data...' : 'Initializing chart...'}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  // Get selected ticker info
  const tickerInfo = GUIDE_TICKERS.find(t => t.symbol === selectedTicker) || GUIDE_TICKERS[0]
  const isPositiveChange = currentPrice.changePercent >= 0

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>Volume Interactive Demo</CardTitle>
        <CardDescription>
          Real ISX market data showing volume patterns and trading activity
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Ticker Selector and Current Price */}
        <div className="flex items-center justify-between pb-2 border-b">
          <div className="flex items-center gap-3">
            <Select value={selectedTicker} onValueChange={setSelectedTicker}>
              <SelectTrigger className="w-[180px]">
                <SelectValue>
                  <Badge variant="outline" className="text-sm">{selectedTicker}</Badge>
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                {GUIDE_TICKERS.map((ticker) => (
                  <SelectItem key={ticker.symbol} value={ticker.symbol}>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">{ticker.symbol}</Badge>
                      <span className="text-xs text-muted-foreground">{ticker.name}</span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <div>
              <p className="text-sm font-medium">{tickerInfo.name}</p>
              <p className="text-xs text-muted-foreground">
                Number of shares traded (not value)
              </p>
            </div>
          </div>
          <div className="text-right">
            <p className="text-lg font-semibold">
              {currentPrice.price.toFixed(3)} IQD
            </p>
            <p className={`text-xs ${isPositiveChange ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
              {isPositiveChange ? '+' : ''}{currentPrice.changePercent.toFixed(2)}%
            </p>
          </div>
        </div>

        {/* Volume Toggle */}
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant={showVolume ? 'default' : 'outline'}
              onClick={() => setShowVolume(!showVolume)}
            >
              {showVolume ? 'Hide Volume' : 'Show Volume'}
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Key Principle:</strong> Volume represents the <strong>number of shares traded</strong> (not monetary value).
            High volume on up days = buying pressure. High volume on down days = selling pressure.
            Color intensity shows volume relative to 20-period average.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">Green Bar</p>
            <p className="text-sm font-semibold text-green-600">Buying Volume</p>
            <p className="text-xs mt-1">Price up, buyers active</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">Red Bar</p>
            <p className="text-sm font-semibold text-red-600">Selling Volume</p>
            <p className="text-xs mt-1">Price down, sellers active</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">Bar Brightness</p>
            <p className="text-sm font-semibold text-blue-600">Volume Intensity</p>
            <p className="text-xs mt-1">Brighter = Higher relative volume</p>
          </div>
        </div>

        {/* Try These Actions */}
        <Card className="bg-blue-50/50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800 mt-4">
          <CardContent className="p-4 space-y-2">
            <h4 className="font-medium text-blue-900 dark:text-blue-100">Try These Actions:</h4>
            <ul className="list-disc list-inside space-y-1 text-sm text-blue-800 dark:text-blue-200 ml-2">
              <li>Select different ISX tickers from the dropdown to see real volume patterns</li>
              <li>Toggle volume visibility on/off to see the difference</li>
              <li>Observe how volume spikes align with large price movements</li>
              <li>Notice the color intensity - brighter bars indicate higher than average volume</li>
              <li>Hover over the chart to see exact volume numbers at each time point</li>
              <li>Compare volume levels across different stocks (banking vs telecom)</li>
            </ul>
          </CardContent>
        </Card>
      </CardContent>
    </Card>
  )
}
