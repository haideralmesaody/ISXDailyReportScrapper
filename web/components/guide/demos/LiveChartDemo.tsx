'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickData,
  Time,
  LineData,
  CandlestickSeries,
  LineSeries,
  AreaSeries
} from 'lightweight-charts'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { BarChart3, LineChart, AreaChart, ZoomIn, ZoomOut, Maximize2, Loader2 } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { useToast } from '@/lib/hooks/use-toast'
import { fetchTickerHistory } from '@/lib/api/analysis'
import type { TickerHistoricalData } from '@/types/analysis'

type ChartType = 'candlestick' | 'line' | 'area'
type Timeframe = '1M' | '3M' | '6M' | '1Y' | 'MAX'

// Predefined tickers for the guide (names from combined CSV data)
const GUIDE_TICKERS = [
  { symbol: 'BBOB', name: 'Bank of Baghdad' },
  { symbol: 'TASC', name: 'Asia Cell Telecommunication' },
  { symbol: 'BNOI', name: 'National Bank of Iraq' },
  { symbol: 'BMNS', name: 'Al -Mansour Bank' },
  { symbol: 'IBSD', name: 'Baghdad Soft Drinks' }
]

// Sample ISX data fallback (for when API fails)
const generateSampleData = () => {
  const data: CandlestickData[] = []
  const startDate = new Date('2024-01-01')
  let basePrice = 1.200

  for (let i = 0; i < 250; i++) {
    const date = new Date(startDate)
    date.setDate(startDate.getDate() + i)

    // Skip Friday and Saturday (Iraqi weekend)
    const dayOfWeek = date.getDay()
    if (dayOfWeek === 5 || dayOfWeek === 6) {
      continue
    }

    const open = basePrice + (Math.random() - 0.5) * 0.02
    const close = open + (Math.random() - 0.5) * 0.03
    const high = Math.max(open, close) + Math.random() * 0.02
    const low = Math.min(open, close) - Math.random() * 0.02

    data.push({
      time: date.toISOString().split('T')[0] as Time,
      open,
      high,
      low,
      close
    })

    basePrice = close + (Math.random() - 0.5) * 0.01
  }

  return data
}

export function LiveChartDemo() {
  const isHydrated = useHydration()
  const { toast } = useToast()
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const seriesRef = useRef<ISeriesApi<'Candlestick' | 'Line' | 'Area'> | null>(null)

  const [chartType, setChartType] = useState<ChartType>('candlestick')
  const [timeframe, setTimeframe] = useState<Timeframe>('1Y')
  const [selectedTicker, setSelectedTicker] = useState(GUIDE_TICKERS[0].symbol)
  const [tickerData, setTickerData] = useState<TickerHistoricalData[]>([])
  const [isLoadingData, setIsLoadingData] = useState(true)
  const [currentPrice, setCurrentPrice] = useState({ price: 0, change: 0, changePercent: 0 })
  const [sampleData] = useState(() => generateSampleData())

  // Convert TickerHistoricalData to TradingView CandlestickData format
  const convertToChartData = useCallback((data: TickerHistoricalData[]): CandlestickData[] => {
    return data.map(d => ({
      time: d.date as Time,
      open: d.open,
      high: d.high,
      low: d.low,
      close: d.close
    }))
  }, [])

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
          description: `Could not load data for ${selectedTicker}. Using sample data.`,
          variant: 'destructive'
        })
        // Fallback to sample data if API fails
        setTickerData(sampleData.map(d => ({
          date: d.time as string,
          open: d.open,
          high: d.high,
          low: d.low,
          close: d.close,
          volume: 0
        })))
        setCurrentPrice({ price: 1.245, change: 0.03, changePercent: 2.45 })
      } finally {
        setIsLoadingData(false)
      }
    }

    loadTickerData()
  }, [isHydrated, selectedTicker, toast, sampleData])

  // Update chart type
  const updateChartType = useCallback((type: ChartType, chart?: IChartApi) => {
    const targetChart = chart || chartRef.current
    if (!targetChart) return

    // Use real ticker data if available, otherwise fallback to sample
    const chartData = tickerData.length > 0 ? convertToChartData(tickerData) : sampleData

    // Remove old series with safety check
    if (seriesRef.current) {
      try {
        targetChart.removeSeries(seriesRef.current as any)
        seriesRef.current = null // Clear reference after successful removal
      } catch (err) {
        console.warn('Failed to remove series, may already be disposed:', err)
        seriesRef.current = null // Clear invalid reference
      }
    }

    // Add new series based on type
    if (type === 'candlestick') {
      const candlestickSeries = targetChart.addSeries(CandlestickSeries, {
        upColor: '#10b981',
        downColor: '#ef4444',
        borderUpColor: '#10b981',
        borderDownColor: '#ef4444',
        wickUpColor: '#10b981',
        wickDownColor: '#ef4444',
      })
      candlestickSeries.setData(chartData)
      seriesRef.current = candlestickSeries as any
    } else if (type === 'line') {
      const lineSeries = targetChart.addSeries(LineSeries, {
        color: '#3b82f6',
        lineWidth: 2,
      })
      const lineData: LineData[] = chartData.map(d => ({
        time: d.time,
        value: d.close
      }))
      lineSeries.setData(lineData)
      seriesRef.current = lineSeries as any
    } else if (type === 'area') {
      const areaSeries = targetChart.addSeries(AreaSeries, {
        topColor: '#3b82f680',
        bottomColor: '#3b82f610',
        lineColor: '#3b82f6',
        lineWidth: 2,
      })
      const areaData: LineData[] = chartData.map(d => ({
        time: d.time,
        value: d.close
      }))
      areaSeries.setData(areaData)
      seriesRef.current = areaSeries as any
    }

    targetChart.timeScale().fitContent()
  }, [tickerData, convertToChartData, sampleData])

  // Initialize chart (only after hydration)
  useEffect(() => {
    if (!isHydrated || !chartContainerRef.current) return

    const chart = createChart(chartContainerRef.current, {
      width: chartContainerRef.current.clientWidth,
      height: 400,
      layout: {
        background: { color: 'transparent' },
        textColor: '#9ca3af',
      },
      grid: {
        vertLines: { color: '#e5e7eb20' },
        horzLines: { color: '#e5e7eb20' },
      },
      crosshair: {
        mode: 1,
      },
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        borderColor: '#e5e7eb40',
      },
      rightPriceScale: {
        borderColor: '#e5e7eb40',
      },
    })

    chartRef.current = chart

    // Add initial series
    updateChartType(chartType, chart)

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current && chart) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth })
      }
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
    }
  }, [isHydrated, chartType, updateChartType])

  // Re-render chart when ticker data changes
  useEffect(() => {
    if (!chartRef.current || !isHydrated || isLoadingData) return

    // Update the existing series with new data
    updateChartType(chartType)
  }, [tickerData, isHydrated, isLoadingData, chartType, updateChartType])

  // Handle chart type change
  const handleChartTypeChange = useCallback((type: ChartType) => {
    setChartType(type)
    updateChartType(type)
  }, [updateChartType])

  // Handle timeframe change
  const handleTimeframeChange = useCallback((tf: Timeframe) => {
    setTimeframe(tf)
    if (!chartRef.current) return

    const now = new Date()
    const monthsAgo = {
      '1M': 1,
      '3M': 3,
      '6M': 6,
      '1Y': 12,
      'MAX': 999
    }[tf]

    const fromDate = new Date(now)
    fromDate.setMonth(now.getMonth() - monthsAgo)

    const fromTimestamp = Math.floor(fromDate.getTime() / 1000)
    const toTimestamp = Math.floor(now.getTime() / 1000)

    if (tf === 'MAX') {
      chartRef.current.timeScale().fitContent()
    } else {
      chartRef.current.timeScale().setVisibleRange({
        from: fromTimestamp as Time,
        to: toTimestamp as Time,
      })
    }
  }, [])

  // Zoom controls
  const handleZoomIn = useCallback(() => {
    if (!chartRef.current) return
    const timeScale = chartRef.current.timeScale()
    const logicalRange = timeScale.getVisibleLogicalRange()
    if (logicalRange) {
      const barsCount = logicalRange.to - logicalRange.from
      timeScale.setVisibleLogicalRange({
        from: logicalRange.from + barsCount * 0.1,
        to: logicalRange.to - barsCount * 0.1,
      })
    }
  }, [])

  const handleZoomOut = useCallback(() => {
    if (!chartRef.current) return
    const timeScale = chartRef.current.timeScale()
    const logicalRange = timeScale.getVisibleLogicalRange()
    if (logicalRange) {
      const barsCount = logicalRange.to - logicalRange.from
      timeScale.setVisibleLogicalRange({
        from: Math.max(0, logicalRange.from - barsCount * 0.1),
        to: logicalRange.to + barsCount * 0.1,
      })
    }
  }, [])

  const handleResetZoom = useCallback(() => {
    if (!chartRef.current) return
    chartRef.current.timeScale().fitContent()
  }, [])

  // Show loading state while hydrating (after all hooks are called)
  if (!isHydrated) {
    return (
      <div className="flex items-center justify-center p-12">
        <div className="text-center space-y-3">
          <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
          <p className="text-sm text-muted-foreground">Loading chart...</p>
        </div>
      </div>
    )
  }

  // Get selected ticker info
  const tickerInfo = GUIDE_TICKERS.find(t => t.symbol === selectedTicker) || GUIDE_TICKERS[0]
  const isPositiveChange = currentPrice.changePercent >= 0

  return (
    <div className="space-y-4">
      {/* Stock Info with Ticker Selector */}
      <div className="flex items-center justify-between">
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
              {isLoadingData ? 'Loading data...' : 'Real ISX Market Data'}
            </p>
          </div>
        </div>
        <div className="text-right">
          {isLoadingData ? (
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          ) : (
            <>
              <p className="text-lg font-semibold">
                {currentPrice.price.toFixed(3)} IQD
              </p>
              <p className={`text-xs ${isPositiveChange ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                {isPositiveChange ? '+' : ''}{currentPrice.changePercent.toFixed(2)}%
              </p>
            </>
          )}
        </div>
      </div>

      {/* Controls */}
      <div className="flex flex-wrap gap-2 items-center">
        <div className="flex gap-1">
          <Button
            size="sm"
            variant={chartType === 'candlestick' ? 'default' : 'outline'}
            onClick={() => handleChartTypeChange('candlestick')}
          >
            <BarChart3 className="h-4 w-4 mr-1" />
            Candlestick
          </Button>
          <Button
            size="sm"
            variant={chartType === 'line' ? 'default' : 'outline'}
            onClick={() => handleChartTypeChange('line')}
          >
            <LineChart className="h-4 w-4 mr-1" />
            Line
          </Button>
          <Button
            size="sm"
            variant={chartType === 'area' ? 'default' : 'outline'}
            onClick={() => handleChartTypeChange('area')}
          >
            <AreaChart className="h-4 w-4 mr-1" />
            Area
          </Button>
        </div>

        <div className="flex gap-1">
          {(['1M', '3M', '6M', '1Y', 'MAX'] as Timeframe[]).map((tf) => (
            <Button
              key={tf}
              size="sm"
              variant={timeframe === tf ? 'default' : 'outline'}
              onClick={() => handleTimeframeChange(tf)}
            >
              {tf}
            </Button>
          ))}
        </div>

        <div className="flex gap-1 ml-auto">
          <Button size="sm" variant="outline" onClick={handleZoomIn}>
            <ZoomIn className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="outline" onClick={handleZoomOut}>
            <ZoomOut className="h-4 w-4" />
          </Button>
          <Button size="sm" variant="outline" onClick={handleResetZoom}>
            <Maximize2 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Chart Container */}
      <Card>
        <CardContent className="p-4">
          <div ref={chartContainerRef} className="w-full" />
        </CardContent>
      </Card>

      {/* Instructions */}
      <Card className="bg-blue-50/50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800">
        <CardContent className="p-4 space-y-2">
          <h4 className="font-medium text-blue-900 dark:text-blue-100">Try These Actions:</h4>
          <ul className="list-disc list-inside space-y-1 text-sm text-blue-800 dark:text-blue-200 ml-2">
            <li>Select different ISX tickers from the dropdown to explore real market data</li>
            <li>Switch between chart types (Candlestick, Line, Area) and observe the differences</li>
            <li>Click timeframe buttons to change the visible range (1M, 3M, 6M, 1Y, MAX)</li>
            <li>Use zoom buttons or scroll wheel to zoom in/out on specific periods</li>
            <li>Click and drag on the chart to pan left/right through historical data</li>
            <li>Hover over the chart to see the crosshair and detailed price information</li>
            <li>Click "Reset" (maximize icon) to fit all data in view</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}
