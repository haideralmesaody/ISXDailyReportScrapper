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
import { calculateStochastic } from '@/lib/guide/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function StochasticChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const kLineSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const dLineSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const ref20SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const ref80SeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [kPeriod, setKPeriod] = useState(14)
  const [dPeriod, setDPeriod] = useState(3)
  const [customK, setCustomK] = useState('14')
  const [customD, setCustomD] = useState('3')

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 100)
  }, [])

  const stochData = useMemo(() => {
    if (!candlestickData.length) return { k: [], d: [] }
    return calculateStochastic(candlestickData, kPeriod, dPeriod)
  }, [candlestickData, kPeriod, dPeriod])

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

    // Candlestick on pane 0
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

    // Stochastic on pane 1
    const kLineSeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: '%K',
      priceLineVisible: false,
    }, 1)
    kLineSeriesRef.current = kLineSeries
    kLineSeries.setData(stochData.k)

    const dLineSeries = chart.addSeries(LineSeries, {
      color: '#f59e0b',
      lineWidth: 2,
      title: '%D',
      priceLineVisible: false,
    }, 1)
    dLineSeriesRef.current = dLineSeries
    dLineSeries.setData(stochData.d)

    // Reference lines at 20 and 80
    const ref20Series = chart.addSeries(LineSeries, {
      color: '#10b981',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceLineVisible: false,
    }, 1)
    ref20SeriesRef.current = ref20Series
    if (candlestickData.length > 0) {
      ref20Series.setData([
        { time: candlestickData[0].time, value: 20 },
        { time: candlestickData[candlestickData.length - 1].time, value: 20 }
      ])
    }

    const ref80Series = chart.addSeries(LineSeries, {
      color: '#ef4444',
      lineWidth: 1,
      lineStyle: LineStyle.Dashed,
      priceLineVisible: false,
    }, 1)
    ref80SeriesRef.current = ref80Series
    if (candlestickData.length > 0) {
      ref80Series.setData([
        { time: candlestickData[0].time, value: 80 },
        { time: candlestickData[candlestickData.length - 1].time, value: 80 }
      ])
    }

    // Set Stochastic pane height
    const panes = chart.panes()
    if (panes.length > 1) {
      const stochPane = panes[1]
      const containerHeight = container.clientHeight || 600
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.30)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      stochPane.setHeight(constrainedHeight)
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
      if (kLineSeriesRef.current) chart.removeSeries(kLineSeriesRef.current)
      if (dLineSeriesRef.current) chart.removeSeries(dLineSeriesRef.current)
      if (ref20SeriesRef.current) chart.removeSeries(ref20SeriesRef.current)
      if (ref80SeriesRef.current) chart.removeSeries(ref80SeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, stochData, kPeriod, dPeriod])

  const handleCustomApply = () => {
    const newK = parseInt(customK)
    const newD = parseInt(customD)
    if (newK >= 5 && newK <= 50 && newD >= 1 && newD <= 10) {
      setKPeriod(newK)
      setDPeriod(newD)
    }
  }

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
        <CardTitle>Stochastic Oscillator Interactive Demo</CardTitle>
        <CardDescription>
          Adjust %K and %D periods to see momentum changes. Watch for overbought (&gt;80) and oversold (&lt;20) levels.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label>Standard Presets</Label>
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={kPeriod === 14 && dPeriod === 3 ? 'default' : 'outline'}
              onClick={() => { setKPeriod(14); setDPeriod(3); setCustomK('14'); setCustomD('3'); }}>
              Standard (14, 3)
            </Button>
            <Button size="sm" variant={kPeriod === 5 && dPeriod === 3 ? 'default' : 'outline'}
              onClick={() => { setKPeriod(5); setDPeriod(3); setCustomK('5'); setCustomD('3'); }}>
              Fast (5, 3)
            </Button>
            <Button size="sm" variant={kPeriod === 21 && dPeriod === 5 ? 'default' : 'outline'}
              onClick={() => { setKPeriod(21); setDPeriod(5); setCustomK('21'); setCustomD('5'); }}>
              Slow (21, 5)
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="space-y-2">
            <Label htmlFor="custom-k">%K Period: {kPeriod} <span className="text-xs text-muted-foreground">(5-50)</span></Label>
            <Input id="custom-k" type="number" min="5" max="50" value={customK}
              onChange={(e) => setCustomK(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="custom-d">%D Period: {dPeriod} <span className="text-xs text-muted-foreground">(1-10)</span></Label>
            <Input id="custom-d" type="number" min="1" max="10" value={customD}
              onChange={(e) => setCustomD(e.target.value)} />
          </div>
          <div className="flex items-end">
            <Button onClick={handleCustomApply} className="w-full">Apply Custom</Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Reading:</strong> Values above 80 indicate overbought conditions (potential sell).
            Values below 20 indicate oversold conditions (potential buy). Crossovers of %K and %D generate signals.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">Overbought Zone</p>
            <p className="text-sm font-semibold text-red-600">&gt; 80</p>
            <p className="text-xs mt-1">Consider selling</p>
          </div>
          <div className="text-center p-3 bg-gray-50 dark:bg-gray-900/20 rounded-lg border border-gray-200 dark:border-gray-800">
            <p className="text-xs text-muted-foreground">Neutral Zone</p>
            <p className="text-sm font-semibold text-gray-600">20 - 80</p>
            <p className="text-xs mt-1">Normal range</p>
          </div>
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">Oversold Zone</p>
            <p className="text-sm font-semibold text-green-600">&lt; 20</p>
            <p className="text-xs mt-1">Consider buying</p>
          </div>
        </div>

        <div className="flex flex-wrap gap-4 justify-center text-sm pt-2 border-t">
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-blue-500" />
            <span>%K Line (Fast)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-orange-500" />
            <span>%D Line (Slow Signal)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-red-500 border-dashed border-t" />
            <span>Overbought (80)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-0.5 bg-green-500 border-dashed border-t" />
            <span>Oversold (20)</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
