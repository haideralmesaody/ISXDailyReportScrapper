'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { createChart, IChartApi, ISeriesApi, CandlestickData, LineSeries, Time, HistogramSeries } from 'lightweight-charts'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'

// Sample ISX data generator (same as LiveChartDemo)
const generateSampleData = () => {
  const data: CandlestickData[] = []
  const startDate = new Date('2024-01-01')
  let basePrice = 1.200

  for (let i = 0; i < 250; i++) {
    const date = new Date(startDate)
    date.setDate(startDate.getDate() + i)

    const dayOfWeek = date.getDay()
    if (dayOfWeek === 5 || dayOfWeek === 6) continue

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

// Simple SMA calculation
const calculateSMA = (data: CandlestickData[], period: number) => {
  const result: { time: Time, value: number }[] = []
  for (let i = period - 1; i < data.length; i++) {
    let sum = 0
    for (let j = 0; j < period; j++) {
      sum += data[i - j].close
    }
    result.push({
      time: data[i].time,
      value: sum / period
    })
  }
  return result
}

// Simple RSI calculation
const calculateRSI = (data: CandlestickData[], period: number = 14) => {
  const result: { time: Time, value: number }[] = []
  let gains: number[] = []
  let losses: number[] = []

  for (let i = 1; i < data.length; i++) {
    const change = data[i].close - data[i - 1].close
    gains.push(change > 0 ? change : 0)
    losses.push(change < 0 ? Math.abs(change) : 0)

    if (i >= period) {
      const avgGain = gains.slice(-period).reduce((a, b) => a + b) / period
      const avgLoss = losses.slice(-period).reduce((a, b) => a + b) / period
      const rs = avgLoss === 0 ? 100 : avgGain / avgLoss
      const rsi = 100 - (100 / (1 + rs))

      result.push({
        time: data[i].time,
        value: rsi
      })
    }
  }

  return result
}

interface IndicatorState {
  sma20: boolean
  sma50: boolean
  sma200: boolean
  rsi: boolean
  volume: boolean
}

export function IndicatorShowcase() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)

  // Indicator series refs
  const sma20SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const sma50SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const sma200SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const rsiSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const rsi70SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const rsi30SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null)

  const [sampleData] = useState(() => generateSampleData())
  const [indicators, setIndicators] = useState<IndicatorState>({
    sma20: false,
    sma50: false,
    sma200: false,
    rsi: false,
    volume: false
  })

  // Initialize chart
  useEffect(() => {
    if (!chartContainerRef.current) return

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

    // Add candlestick series (Pane 0)
    const candlestickSeries = chart.addSeries('Candlestick' as any, {
      upColor: '#10b981',
      downColor: '#ef4444',
      borderUpColor: '#10b981',
      borderDownColor: '#ef4444',
      wickUpColor: '#10b981',
      wickDownColor: '#ef4444',
    })
    candlestickSeries.setData(sampleData)
    candlestickSeriesRef.current = candlestickSeries

    chart.timeScale().fitContent()

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
  }, [sampleData])

  // Update indicators when toggles change
  useEffect(() => {
    if (!chartRef.current) return

    const chart = chartRef.current

    // SMA 20
    if (indicators.sma20) {
      if (!sma20SeriesRef.current) {
        const sma20Series = chart.addSeries(LineSeries, {
          color: '#3b82f6',
          lineWidth: 2,
          title: 'SMA 20'
        })
        sma20Series.setData(calculateSMA(sampleData, 20))
        sma20SeriesRef.current = sma20Series
      }
    } else {
      if (sma20SeriesRef.current) {
        chart.removeSeries(sma20SeriesRef.current)
        sma20SeriesRef.current = null
      }
    }

    // SMA 50
    if (indicators.sma50) {
      if (!sma50SeriesRef.current) {
        const sma50Series = chart.addSeries(LineSeries, {
          color: '#f59e0b',
          lineWidth: 2,
          title: 'SMA 50'
        })
        sma50Series.setData(calculateSMA(sampleData, 50))
        sma50SeriesRef.current = sma50Series
      }
    } else {
      if (sma50SeriesRef.current) {
        chart.removeSeries(sma50SeriesRef.current)
        sma50SeriesRef.current = null
      }
    }

    // SMA 200
    if (indicators.sma200) {
      if (!sma200SeriesRef.current) {
        const sma200Series = chart.addSeries(LineSeries, {
          color: '#ef4444',
          lineWidth: 2,
          title: 'SMA 200'
        })
        sma200Series.setData(calculateSMA(sampleData, 200))
        sma200SeriesRef.current = sma200Series
      }
    } else {
      if (sma200SeriesRef.current) {
        chart.removeSeries(sma200SeriesRef.current)
        sma200SeriesRef.current = null
      }
    }

    // RSI (in separate pane)
    if (indicators.rsi) {
      if (!rsiSeriesRef.current) {
        // RSI main line
        const rsiSeries = chart.addSeries(LineSeries, {
          color: '#9c27b0',
          lineWidth: 2,
          title: 'RSI',
          priceScaleId: 'rsi',
        }, 1) // Pane 1
        rsiSeries.setData(calculateRSI(sampleData, 14))
        rsiSeriesRef.current = rsiSeries

        // RSI 70 line (overbought)
        const rsi70Series = chart.addSeries(LineSeries, {
          color: '#ef4444',
          lineWidth: 1,
          lineStyle: 2, // Dashed
          priceScaleId: 'rsi',
        }, 1)
        const rsi70Data = sampleData.slice(14).map(d => ({ time: d.time, value: 70 }))
        rsi70Series.setData(rsi70Data)
        rsi70SeriesRef.current = rsi70Series

        // RSI 30 line (oversold)
        const rsi30Series = chart.addSeries(LineSeries, {
          color: '#10b981',
          lineWidth: 1,
          lineStyle: 2, // Dashed
          priceScaleId: 'rsi',
        }, 1)
        const rsi30Data = sampleData.slice(14).map(d => ({ time: d.time, value: 30 }))
        rsi30Series.setData(rsi30Data)
        rsi30SeriesRef.current = rsi30Series

        // Set RSI pane height
        const panes = chart.panes()
        if (panes.length > 1) {
          const containerHeight = chartContainerRef.current?.clientHeight || 400
          const rsiHeight = Math.max(60, Math.min(Math.floor(containerHeight * 0.20), Math.floor(containerHeight * 0.90)))
          panes[1].setHeight(rsiHeight)
        }
      }
    } else {
      if (rsiSeriesRef.current) {
        chart.removeSeries(rsiSeriesRef.current)
        rsiSeriesRef.current = null
      }
      if (rsi70SeriesRef.current) {
        chart.removeSeries(rsi70SeriesRef.current)
        rsi70SeriesRef.current = null
      }
      if (rsi30SeriesRef.current) {
        chart.removeSeries(rsi30SeriesRef.current)
        rsi30SeriesRef.current = null
      }
    }

    // Volume (in separate pane)
    if (indicators.volume) {
      if (!volumeSeriesRef.current) {
        const volumePaneIndex = indicators.rsi ? 2 : 1
        const volumeSeries = chart.addSeries(HistogramSeries, {
          color: '#10b981',
          priceFormat: { type: 'volume' },
          priceScaleId: 'volume',
        }, volumePaneIndex)

        // Generate volume data (mock)
        const volumeData = sampleData.map(d => ({
          time: d.time,
          value: Math.random() * 1000000 + 500000,
          color: d.close > d.open ? '#10b981' : '#ef4444'
        }))
        volumeSeries.setData(volumeData)
        volumeSeriesRef.current = volumeSeries

        // Set volume pane height
        const panes = chart.panes()
        if (panes.length > volumePaneIndex) {
          const containerHeight = chartContainerRef.current?.clientHeight || 400
          const volumeHeight = Math.max(50, Math.min(Math.floor(containerHeight * 0.15), Math.floor(containerHeight * 0.90)))
          panes[volumePaneIndex].setHeight(volumeHeight)
        }
      }
    } else {
      if (volumeSeriesRef.current) {
        chart.removeSeries(volumeSeriesRef.current)
        volumeSeriesRef.current = null
      }
    }

  }, [indicators, sampleData])

  const toggleIndicator = useCallback((indicator: keyof IndicatorState) => {
    setIndicators(prev => ({
      ...prev,
      [indicator]: !prev[indicator]
    }))
  }, [])

  const enableAll = useCallback(() => {
    setIndicators({
      sma20: true,
      sma50: true,
      sma200: true,
      rsi: true,
      volume: true
    })
  }, [])

  const disableAll = useCallback(() => {
    setIndicators({
      sma20: false,
      sma50: false,
      sma200: false,
      rsi: false,
      volume: false
    })
  }, [])

  return (
    <div className="space-y-4">
      {/* Stock Info */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Badge variant="outline">BBOB</Badge>
          <div>
            <p className="text-sm font-medium">Bank of Babylon</p>
            <p className="text-xs text-muted-foreground">Sample ISX Stock with Indicators</p>
          </div>
        </div>
        <div className="text-right">
          <p className="text-lg font-semibold">1.198 IQD</p>
          <p className="text-xs text-green-600 dark:text-green-400">+1.85%</p>
        </div>
      </div>

      {/* Controls */}
      <div className="grid gap-4 md:grid-cols-[1fr_2fr]">
        {/* Indicator Toggles */}
        <Card>
          <CardContent className="p-4 space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="font-medium text-sm">Active Indicators</h4>
              <div className="flex gap-1">
                <Button size="sm" variant="outline" onClick={enableAll}>
                  All
                </Button>
                <Button size="sm" variant="outline" onClick={disableAll}>
                  None
                </Button>
              </div>
            </div>

            <ScrollArea className="h-48">
              <div className="space-y-3">
                {/* Moving Averages */}
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground">Moving Averages</p>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="sma20"
                      checked={indicators.sma20}
                      onCheckedChange={() => toggleIndicator('sma20')}
                    />
                    <Label htmlFor="sma20" className="text-sm flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full bg-blue-500" />
                      SMA 20
                    </Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="sma50"
                      checked={indicators.sma50}
                      onCheckedChange={() => toggleIndicator('sma50')}
                    />
                    <Label htmlFor="sma50" className="text-sm flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full bg-amber-500" />
                      SMA 50
                    </Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="sma200"
                      checked={indicators.sma200}
                      onCheckedChange={() => toggleIndicator('sma200')}
                    />
                    <Label htmlFor="sma200" className="text-sm flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full bg-red-500" />
                      SMA 200
                    </Label>
                  </div>
                </div>

                {/* Momentum */}
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground">Momentum</p>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="rsi"
                      checked={indicators.rsi}
                      onCheckedChange={() => toggleIndicator('rsi')}
                    />
                    <Label htmlFor="rsi" className="text-sm flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full bg-purple-500" />
                      RSI (14)
                    </Label>
                  </div>
                </div>

                {/* Volume */}
                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground">Volume</p>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="volume"
                      checked={indicators.volume}
                      onCheckedChange={() => toggleIndicator('volume')}
                    />
                    <Label htmlFor="volume" className="text-sm flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full bg-green-500" />
                      Volume
                    </Label>
                  </div>
                </div>
              </div>
            </ScrollArea>

            <div className="pt-2 border-t text-xs text-muted-foreground">
              <p className="font-medium mb-1">Legend:</p>
              <ul className="list-disc list-inside space-y-0.5 ml-2">
                <li>Check/uncheck to toggle indicators</li>
                <li>RSI and Volume appear in separate panes</li>
                <li>SMAs overlay on price chart</li>
              </ul>
            </div>
          </CardContent>
        </Card>

        {/* Chart */}
        <Card>
          <CardContent className="p-4">
            <div ref={chartContainerRef} className="w-full" />
          </CardContent>
        </Card>
      </div>

      {/* Instructions */}
      <Card className="bg-blue-50/50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-800">
        <CardContent className="p-4 space-y-2">
          <h4 className="font-medium text-blue-900 dark:text-blue-100">Try These Actions:</h4>
          <ul className="list-disc list-inside space-y-1 text-sm text-blue-800 dark:text-blue-200 ml-2">
            <li>Toggle individual indicators on/off and observe their behavior</li>
            <li>Enable all moving averages (SMA 20, 50, 200) to see trend hierarchy</li>
            <li>Turn on RSI to see overbought/oversold zones (70/30 lines)</li>
            <li>Add Volume to confirm price movements with trading activity</li>
            <li>Hover over chart to see all indicator values in real-time</li>
            <li>Notice how RSI and Volume appear in separate panes below price chart</li>
          </ul>
        </CardContent>
      </Card>

      {/* Signal Examples */}
      <Card>
        <CardContent className="p-4 space-y-3">
          <h4 className="font-medium">Example Trading Signals to Look For:</h4>
          <div className="grid gap-2 md:grid-cols-2 text-sm">
            <div className="space-y-1">
              <p className="font-medium text-green-600 dark:text-green-400">Bullish Signals:</p>
              <ul className="list-disc list-inside space-y-0.5 text-muted-foreground ml-2">
                <li>Price crosses above SMA 20 or SMA 50</li>
                <li>SMA 20 crosses above SMA 50 (Golden Cross if SMA 50 &gt; SMA 200)</li>
                <li>RSI bounces from oversold zone (&lt;30)</li>
                <li>High volume on up days (green bars)</li>
              </ul>
            </div>
            <div className="space-y-1">
              <p className="font-medium text-red-600 dark:text-red-400">Bearish Signals:</p>
              <ul className="list-disc list-inside space-y-0.5 text-muted-foreground ml-2">
                <li>Price crosses below SMA 20 or SMA 50</li>
                <li>SMA 20 crosses below SMA 50 (Death Cross if SMA 50 &lt; SMA 200)</li>
                <li>RSI reverses from overbought zone (&gt;70)</li>
                <li>High volume on down days (red bars)</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
