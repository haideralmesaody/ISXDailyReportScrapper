'use client'

import { useState, useEffect, useRef, useMemo } from 'react'
import {
  createChart,
  IChartApi,
  ISeriesApi,
  CandlestickSeries,
  LineSeries,
} from 'lightweight-charts'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Info } from 'lucide-react'
import { useHydration } from '@/lib/hooks'
import { calculateParabolicSAR } from '@/lib/indicators/calculations'
import { generateMockStockData } from '@/lib/guide/mock-data'
import { Alert, AlertDescription } from '@/components/ui/alert'

export function ParabolicSARChartDemo({ className }: { className?: string }) {
  const isHydrated = useHydration()
  const { theme } = useTheme()

  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const sarUptrendSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const sarDowntrendSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)

  const [acceleration, setAcceleration] = useState(0.02)

  const candlestickData = useMemo(() => {
    return generateMockStockData(120, 1200, 150)
  }, [])

  const sarData = useMemo(() => {
    if (!candlestickData.length) return { uptrend: [], downtrend: [] }

    const highs = candlestickData.map(d => d.high)
    const lows = candlestickData.map(d => d.low)
    const closes = candlestickData.map(d => d.close)

    const sarValues = calculateParabolicSAR(highs, lows, closes, acceleration, acceleration * 10)

    const uptrend: Array<{ time: string; value: number }> = []
    const downtrend: Array<{ time: string; value: number }> = []

    candlestickData.forEach((d, i) => {
      const sar = sarValues[i]
      if (sar !== null) {
        if (sar.isUptrend) {
          uptrend.push({ time: d.time, value: sar.value })
        } else {
          downtrend.push({ time: d.time, value: sar.value })
        }
      }
    })

    return { uptrend, downtrend }
  }, [candlestickData, acceleration])

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

    // SAR uptrend dots (Pane 0) - Green
    if (sarData.uptrend.length > 0) {
      const sarUptrendSeries = chart.addSeries(LineSeries, {
        color: '#10b981',
        lineWidth: 0,
        crosshairMarkerRadius: 5,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 0)
      sarUptrendSeriesRef.current = sarUptrendSeries
      sarUptrendSeries.setData(sarData.uptrend)
    }

    // SAR downtrend dots (Pane 0) - Red
    if (sarData.downtrend.length > 0) {
      const sarDowntrendSeries = chart.addSeries(LineSeries, {
        color: '#ef4444',
        lineWidth: 0,
        crosshairMarkerRadius: 5,
        priceLineVisible: false,
        lastValueVisible: false,
      }, 0)
      sarDowntrendSeriesRef.current = sarDowntrendSeries
      sarDowntrendSeries.setData(sarData.downtrend)
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
      if (sarDowntrendSeriesRef.current) chart.removeSeries(sarDowntrendSeriesRef.current)
      if (sarUptrendSeriesRef.current) chart.removeSeries(sarUptrendSeriesRef.current)
      if (candlestickSeriesRef.current) chart.removeSeries(candlestickSeriesRef.current)
      chart.remove()
      chartRef.current = null
    }
  }, [isHydrated, theme, candlestickData, sarData, acceleration])

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
        <CardTitle>Parabolic SAR Interactive Demo</CardTitle>
        <CardDescription>
          SAR dots below price = uptrend. SAR dots above = downtrend. Dot flip = trend reversal.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button size="sm" variant={acceleration === 0.02 ? 'default' : 'outline'} onClick={() => setAcceleration(0.02)}>
              Standard (0.02)
            </Button>
            <Button size="sm" variant={acceleration === 0.03 ? 'default' : 'outline'} onClick={() => setAcceleration(0.03)}>
              Medium (0.03)
            </Button>
            <Button size="sm" variant={acceleration === 0.05 ? 'default' : 'outline'} onClick={() => setAcceleration(0.05)}>
              Fast (0.05)
            </Button>
          </div>
        </div>

        <Alert>
          <Info className="h-4 w-4" />
          <AlertDescription>
            <strong>Trading Signals:</strong> SAR below price = buy (uptrend). SAR above price = sell (downtrend).
            Dot flip indicates trend reversal and potential entry/exit point.
          </AlertDescription>
        </Alert>

        <div ref={chartContainerRef} className="w-full h-[600px] mt-4" />

        <div className="grid grid-cols-3 gap-4 pt-2">
          <div className="text-center p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
            <p className="text-xs text-muted-foreground">SAR Below</p>
            <p className="text-sm font-semibold text-green-600">Uptrend</p>
            <p className="text-xs mt-1">Buy signal</p>
          </div>
          <div className="text-center p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
            <p className="text-xs text-muted-foreground">SAR Above</p>
            <p className="text-sm font-semibold text-red-600">Downtrend</p>
            <p className="text-xs mt-1">Sell signal</p>
          </div>
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <p className="text-xs text-muted-foreground">Reversal</p>
            <p className="text-sm font-semibold text-blue-600">Change Position</p>
            <p className="text-xs mt-1">Flip signal</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
