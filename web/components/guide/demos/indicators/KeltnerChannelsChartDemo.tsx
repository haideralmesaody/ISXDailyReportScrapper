'use client'

import { useState, useEffect, useRef } from 'react'
import { createChart, ColorType, LineStyle, CandlestickSeries, LineSeries, AreaSeries } from 'lightweight-charts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2 } from 'lucide-react'
import { calculateKeltnerChannels, generateSampleISXData } from '@/lib/guide/indicators/calculations'

export function KeltnerChannelsChartDemo() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<any>(null)
  const [period, setPeriod] = useState(20)
  const [multiplier, setMultiplier] = useState(2.0)
  const [showFill, setShowFill] = useState(true)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    if (!chartContainerRef.current) return

    setIsLoading(true)

    // Create chart
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: 'transparent' },
        textColor: '#9ca3af',
      },
      grid: {
        vertLines: { color: '#1f2937' },
        horzLines: { color: '#1f2937' },
      },
      width: chartContainerRef.current.clientWidth,
      height: 400,
      timeScale: {
        borderColor: '#4b5563',
      },
      rightPriceScale: {
        borderColor: '#4b5563',
      },
    })

    chartRef.current = chart

    // Generate sample data (200 days provides sufficient buffer for ATR + EMA alignment)
    const data = generateSampleISXData(200)
    console.log(`[Keltner Demo] Generated data: ${data.length} bars`)

    // ERROR CHECK: Validate data generation
    if (data.length < 50) {
      console.error(`[Keltner Demo] Insufficient data: ${data.length} bars`)
      setIsLoading(false)
      return
    }

    // Add candlestick series
    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#10b981',
      downColor: '#ef4444',
      borderVisible: false,
      wickUpColor: '#10b981',
      wickDownColor: '#ef4444',
    })
    candleSeries.setData(data)

    // Calculate and add Keltner Channels
    const keltner = calculateKeltnerChannels(data, period, multiplier)

    // ERROR CHECK: Validate calculation results
    console.log(`[Keltner Demo] Calculation results:`, {
      upper: keltner.upper.length,
      middle: keltner.middle.length,
      lower: keltner.lower.length,
      period,
      multiplier
    })

    if (keltner.upper.length === 0 || keltner.lower.length === 0 || keltner.middle.length === 0) {
      console.error(`[Keltner Demo] Calculation failed - empty arrays returned`)
      setIsLoading(false)
      return
    }

    // Upper band (orange)
    const upperSeries = chart.addSeries(LineSeries, {
      color: '#f59e0b',
      lineWidth: 2,
      title: `Upper (EMA+${multiplier}×ATR)`,
    })
    upperSeries.setData(keltner.upper)

    // Middle line (EMA - white)
    const middleSeries = chart.addSeries(LineSeries, {
      color: '#ffffff',
      lineWidth: 2,
      title: `EMA (${period})`,
    })
    middleSeries.setData(keltner.middle)

    // Lower band (orange)
    const lowerSeries = chart.addSeries(LineSeries, {
      color: '#f59e0b',
      lineWidth: 2,
      title: `Lower (EMA-${multiplier}×ATR)`,
    })
    lowerSeries.setData(keltner.lower)

    // Add fill area between upper and lower - conditional
    if (showFill) {
      const areaSeries = chart.addSeries(AreaSeries, {
        topColor: 'rgba(245, 158, 11, 0.3)',
        bottomColor: 'rgba(245, 158, 11, 0.1)',
        lineColor: 'transparent',
        lineWidth: 0,
      })
      areaSeries.setData(keltner.upper.map((point, i) => ({
        time: point.time,
        value: (point.value + keltner.lower[i].value) / 2
      })))
    }

    chart.timeScale().fitContent()
    setIsLoading(false)

    // Responsive resize
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
        })
      }
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
    }
  }, [period, multiplier, showFill])

  return (
    <Card className="border-orange-200 dark:border-orange-800">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Keltner Channels Interactive Demo</span>
          {isLoading && <Loader2 className="h-5 w-5 animate-spin text-orange-600" />}
        </CardTitle>
        <CardDescription>
          Experiment with ATR-based volatility bands (smoother than Bollinger)
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Controls */}
        <div className="flex flex-wrap gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Period (EMA)</label>
            <div className="flex gap-2">
              <Button
                variant={period === 10 ? 'default' : 'outline'}
                size="sm"
                onClick={() => setPeriod(10)}
              >
                10
              </Button>
              <Button
                variant={period === 20 ? 'default' : 'outline'}
                size="sm"
                onClick={() => setPeriod(20)}
              >
                20
              </Button>
              <Button
                variant={period === 50 ? 'default' : 'outline'}
                size="sm"
                onClick={() => setPeriod(50)}
              >
                50
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">ATR Multiplier</label>
            <div className="flex gap-2">
              <Button
                variant={multiplier === 1.5 ? 'default' : 'outline'}
                size="sm"
                onClick={() => setMultiplier(1.5)}
              >
                1.5
              </Button>
              <Button
                variant={multiplier === 2.0 ? 'default' : 'outline'}
                size="sm"
                onClick={() => setMultiplier(2.0)}
              >
                2.0
              </Button>
              <Button
                variant={multiplier === 2.5 ? 'default' : 'outline'}
                size="sm"
                onClick={() => setMultiplier(2.5)}
              >
                2.5
              </Button>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Display</label>
            <Button
              variant={showFill ? 'default' : 'outline'}
              size="sm"
              onClick={() => setShowFill(!showFill)}
            >
              Fill Area
            </Button>
          </div>
        </div>

        {/* Chart */}
        <div ref={chartContainerRef} className="w-full" />

        {/* Legend */}
        <div className="flex flex-wrap gap-2">
          <Badge className="bg-orange-600">Upper Band (EMA + {multiplier}×ATR)</Badge>
          <Badge variant="outline" className="text-white">Middle Line (EMA {period})</Badge>
          <Badge className="bg-orange-600">Lower Band (EMA - {multiplier}×ATR)</Badge>
        </div>

        {/* Interpretation */}
        <div className="grid gap-2 md:grid-cols-2 text-sm">
          <div className="p-3 rounded-lg bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800">
            <p className="font-medium text-green-800 dark:text-green-200">Trend Following</p>
            <p className="text-green-600 dark:text-green-400">Price above EMA + inside bands = Strong uptrend</p>
          </div>
          <div className="p-3 rounded-lg bg-orange-50 dark:bg-orange-950/20 border border-orange-200 dark:border-orange-800">
            <p className="font-medium text-orange-800 dark:text-orange-200">Breakout Signals</p>
            <p className="text-orange-600 dark:text-orange-400">Price breaks above/below bands = Volatility expansion</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
