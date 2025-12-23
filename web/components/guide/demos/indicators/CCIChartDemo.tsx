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
import { calculateCCI } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function CCIChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const cciSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const upperLineRef = useRef<ISeriesApi<'Line'> | null>(null)
  const lowerLineRef = useRef<ISeriesApi<'Line'> | null>(null)
  const zeroLineRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [period, setPeriod] = useState(20)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const cciData = useMemo(() => {
    if (!candlestickData.length) return []
    return calculateCCI(candlestickData, period)
  }, [candlestickData, period])

  // Generate reference lines (+100, 0, -100)
  const referenceLines = useMemo(() => {
    if (!cciData.length) return { upper: [], zero: [], lower: [] }

    const firstTime = cciData[0].time
    const lastTime = cciData[cciData.length - 1].time

    return {
      upper: [
        { time: firstTime, value: 100 },
        { time: lastTime, value: 100 }
      ],
      zero: [
        { time: firstTime, value: 0 },
        { time: lastTime, value: 0 }
      ],
      lower: [
        { time: firstTime, value: -100 },
        { time: lastTime, value: -100 }
      ]
    }
  }, [cciData])

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

    // CCI line (Pane 1)
    const cciSeries = chart.addSeries(LineSeries, {
      color: '#06b6d4',
      lineWidth: 2,
      title: 'CCI',
      priceLineVisible: false,
    }, 1)
    cciSeriesRef.current = cciSeries
    cciSeries.setData(cciData)

    // Reference lines
    if (referenceLines.upper.length > 0) {
      const upperLine = chart.addSeries(LineSeries, {
        color: '#ef4444',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 1)
      upperLineRef.current = upperLine
      upperLine.setData(referenceLines.upper)

      const zeroLine = chart.addSeries(LineSeries, {
        color: theme === 'dark' ? '#6b7280' : '#9ca3af',
        lineWidth: 1,
        lineStyle: LineStyle.Dotted,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 1)
      zeroLineRef.current = zeroLine
      zeroLine.setData(referenceLines.zero)

      const lowerLine = chart.addSeries(LineSeries, {
        color: '#10b981',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 1)
      lowerLineRef.current = lowerLine
      lowerLine.setData(referenceLines.lower)
    }

    // Set CCI pane height
    const panes = chart.panes()
    if (panes.length > 1) {
      const cciPane = panes[1]
      const containerHeight = container.clientHeight || 600
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.30)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      cciPane.setHeight(constrainedHeight)
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
      if (lowerLineRef.current) chart.removeSeries(lowerLineRef.current)
      if (zeroLineRef.current) chart.removeSeries(zeroLineRef.current)
      if (upperLineRef.current) chart.removeSeries(upperLineRef.current)
      if (cciSeriesRef.current) chart.removeSeries(cciSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, cciData, referenceLines, period])

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
        <CardTitle>Commodity Channel Index (CCI) Interactive Demo</CardTitle>
        <CardDescription>
          CCI &gt; +100 = overbought. CCI &lt; -100 = oversold. Measures deviation from average price.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={period === 14 ? 'default' : 'outline'} onClick={() => setPeriod(14)}>
              Short (14)
            </Button>
            <Button size="sm" variant={period === 20 ? 'default' : 'outline'} onClick={() => setPeriod(20)}>
              Standard (20)
            </Button>
            <Button size="sm" variant={period === 30 ? 'default' : 'outline'} onClick={() => setPeriod(30)}>
              Long (30)
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Trading Signals:</strong> CCI crossing above +100 = buy signal. CCI crossing below -100 = sell signal.
            CCI above +100 = price unusually high. CCI below -100 = price unusually low.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">CCI &gt; +100</p>
            <p className="text-sm font-semibold text-red-600">Overbought</p>
            <p className="text-xs mt-1">Consider selling</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">-100 to +100</p>
            <p className="text-sm font-semibold text-blue-600">Normal Range</p>
            <p className="text-xs mt-1">No extreme signal</p>
          </div>
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">CCI &lt; -100</p>
            <p className="text-sm font-semibold text-green-600">Oversold</p>
            <p className="text-xs mt-1">Consider buying</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
