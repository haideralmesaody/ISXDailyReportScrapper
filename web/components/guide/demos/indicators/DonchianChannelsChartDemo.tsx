'use client'

import { useState, useEffect, useRef } from 'react'
import { createChart, ColorType, LineStyle, CandlestickSeries, LineSeries, AreaSeries } from 'lightweight-charts'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Loader2 } from 'lucide-react'
import { calculateDonchianChannels, generateSampleISXData } from '@/lib/guide/indicators/calculations'

export function DonchianChannelsChartDemo() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<any>(null)
  const [period, setPeriod] = useState(20)
  const [showMiddle, setShowMiddle] = useState(true)
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

    // Generate sample data (200 days provides sufficient buffer for all calculations)
    const data = generateSampleISXData(200)
    console.log(`[Donchian Demo] Generated data: ${data.length} bars`)

    // ERROR CHECK: Validate data generation
    if (data.length < 50) {
      console.error(`[Donchian Demo] Insufficient data: ${data.length} bars`)
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

    // Calculate and add Donchian Channels
    const donchian = calculateDonchianChannels(data, period)

    // ERROR CHECK: Validate calculation results
    console.log(`[Donchian Demo] Calculation results:`, {
      upper: donchian.upper.length,
      middle: donchian.middle.length,
      lower: donchian.lower.length,
      period
    })

    if (donchian.upper.length === 0 || donchian.lower.length === 0) {
      console.error(`[Donchian Demo] Calculation failed - empty arrays returned`)
      setIsLoading(false)
      return
    }

    // Upper band (blue)
    const upperSeries = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      title: `Upper (${period})`,
    })
    upperSeries.setData(donchian.upper)

    // Middle line (gray dashed) - conditional
    let middleSeries: any = null
    if (showMiddle) {
      middleSeries = chart.addSeries(LineSeries, {
        color: '#9ca3af',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        title: `Middle (${period})`,
      })
      middleSeries.setData(donchian.middle)
    }

    // Lower band (red)
    const lowerSeries = chart.addSeries(LineSeries, {
      color: '#ef4444',
      lineWidth: 2,
      title: `Lower (${period})`,
    })
    lowerSeries.setData(donchian.lower)

    // Add fill area between upper and lower - conditional
    if (showFill) {
      const areaSeries = chart.addSeries(AreaSeries, {
        topColor: 'rgba(59, 130, 246, 0.2)',
        bottomColor: 'rgba(239, 68, 68, 0.2)',
        lineColor: 'transparent',
        lineWidth: 0,
      })
      areaSeries.setData(donchian.upper.map((point, i) => ({
        time: point.time,
        value: (point.value + donchian.lower[i].value) / 2
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
  }, [period, showMiddle, showFill])

  return (
    <Card className="border-blue-200 dark:border-blue-800">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Donchian Channels Interactive Demo</span>
          {isLoading && <Loader2 className="h-5 w-5 animate-spin text-blue-600" />}
        </CardTitle>
        <CardDescription>
          Experiment with period settings and visualize breakout levels
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Controls */}
        <div className="flex flex-wrap gap-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Period</label>
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
            <label className="text-sm font-medium">Display Options</label>
            <div className="flex gap-2">
              <Button
                variant={showMiddle ? 'default' : 'outline'}
                size="sm"
                onClick={() => setShowMiddle(!showMiddle)}
              >
                Middle Line
              </Button>
              <Button
                variant={showFill ? 'default' : 'outline'}
                size="sm"
                onClick={() => setShowFill(!showFill)}
              >
                Fill Area
              </Button>
            </div>
          </div>
        </div>

        {/* Chart */}
        <div ref={chartContainerRef} className="w-full" />

        {/* Legend */}
        <div className="flex flex-wrap gap-2">
          <Badge className="bg-blue-600">Upper Band (Highest High)</Badge>
          {showMiddle && <Badge variant="outline" className="text-gray-400">Middle Line</Badge>}
          <Badge className="bg-red-600">Lower Band (Lowest Low)</Badge>
        </div>

        {/* Interpretation */}
        <div className="grid gap-2 md:grid-cols-2 text-sm">
          <div className="p-3 rounded-lg bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800">
            <p className="font-medium text-green-800 dark:text-green-200">Bullish Breakout</p>
            <p className="text-green-600 dark:text-green-400">Price breaks above upper band → potential uptrend</p>
          </div>
          <div className="p-3 rounded-lg bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800">
            <p className="font-medium text-red-800 dark:text-red-200">Bearish Breakdown</p>
            <p className="text-red-600 dark:text-red-400">Price breaks below lower band → potential downtrend</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
