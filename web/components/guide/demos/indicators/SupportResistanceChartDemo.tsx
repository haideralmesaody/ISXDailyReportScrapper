'use client'

import { useEffect, useRef, useState } from 'react'
import { createChart, IChartApi, ISeriesApi, LineStyle } from 'lightweight-charts'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { generateSampleISXData, calculateSupportResistance } from '@/lib/guide/indicators/calculations'

export function SupportResistanceChartDemo() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const supportSeriesRefs = useRef<ISeriesApi<'Line'>[]>([])
  const resistanceSeriesRefs = useRef<ISeriesApi<'Line'>[]>([])

  const [sensitivity, setSensitivity] = useState(5)
  const [maxLevels, setMaxLevels] = useState(3)
  const [showSupport, setShowSupport] = useState(true)
  const [showResistance, setShowResistance] = useState(true)

  useEffect(() => {
    if (!chartContainerRef.current) return

    // Generate sample data
    const sampleData = generateSampleISXData(100, 1000, 50)

    // Calculate support and resistance levels
    const highs = sampleData.map(d => d.high)
    const lows = sampleData.map(d => d.low)
    const closes = sampleData.map(d => d.close)

    const { support, resistance } = calculateSupportResistance(
      highs,
      lows,
      closes,
      sensitivity,
      maxLevels
    )

    // Create chart if it doesn't exist
    if (!chartRef.current) {
      const chart = createChart(chartContainerRef.current, {
        width: chartContainerRef.current.clientWidth,
        height: 400,
        layout: {
          background: { color: '#1a1a1a' },
          textColor: '#d1d5db',
        },
        grid: {
          vertLines: { color: '#2B2B43' },
          horzLines: { color: '#2B2B43' },
        },
        crosshair: {
          mode: 1,
        },
        rightPriceScale: {
          borderColor: '#2B2B43',
        },
        timeScale: {
          borderColor: '#2B2B43',
          timeVisible: true,
          secondsVisible: false,
        },
      })

      chartRef.current = chart

      // Add candlestick series
      const candlestickSeries = chart.addCandlestickSeries({
        upColor: '#10b981',
        downColor: '#ef4444',
        borderUpColor: '#10b981',
        borderDownColor: '#ef4444',
        wickUpColor: '#10b981',
        wickDownColor: '#ef4444',
      })

      candlestickSeries.setData(sampleData)
      candlestickSeriesRef.current = candlestickSeries

      // Fit content
      chart.timeScale().fitContent()

      // Handle resize
      const handleResize = () => {
        if (chartContainerRef.current && chartRef.current) {
          chartRef.current.applyOptions({
            width: chartContainerRef.current.clientWidth,
          })
        }
      }

      window.addEventListener('resize', handleResize)

      return () => {
        window.removeEventListener('resize', handleResize)
      }
    }

    // Clean up existing support/resistance lines
    supportSeriesRefs.current.forEach(series => {
      if (chartRef.current) {
        chartRef.current.removeSeries(series)
      }
    })
    resistanceSeriesRefs.current.forEach(series => {
      if (chartRef.current) {
        chartRef.current.removeSeries(series)
      }
    })
    supportSeriesRefs.current = []
    resistanceSeriesRefs.current = []

    // Add support lines
    if (showSupport && chartRef.current) {
      support.forEach((level) => {
        const supportLine = chartRef.current!.addLineSeries({
          color: '#10b981',
          lineWidth: 2,
          lineStyle: LineStyle.Dashed,
          priceLineVisible: false,
          lastValueVisible: false,
          crosshairMarkerVisible: false,
        })

        // Create horizontal line data across all time points
        const lineData = sampleData.map(point => ({
          time: point.time,
          value: level,
        }))

        supportLine.setData(lineData)
        supportSeriesRefs.current.push(supportLine)
      })
    }

    // Add resistance lines
    if (showResistance && chartRef.current) {
      resistance.forEach((level) => {
        const resistanceLine = chartRef.current!.addLineSeries({
          color: '#ef4444',
          lineWidth: 2,
          lineStyle: LineStyle.Dashed,
          priceLineVisible: false,
          lastValueVisible: false,
          crosshairMarkerVisible: false,
        })

        // Create horizontal line data across all time points
        const lineData = sampleData.map(point => ({
          time: point.time,
          value: level,
        }))

        resistanceLine.setData(lineData)
        resistanceSeriesRefs.current.push(resistanceLine)
      })
    }

    // Cleanup function
    return () => {
      if (chartRef.current) {
        supportSeriesRefs.current.forEach(series => {
          chartRef.current!.removeSeries(series)
        })
        resistanceSeriesRefs.current.forEach(series => {
          chartRef.current!.removeSeries(series)
        })
        supportSeriesRefs.current = []
        resistanceSeriesRefs.current = []
      }
    }
  }, [sensitivity, maxLevels, showSupport, showResistance])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (chartRef.current) {
        chartRef.current.remove()
        chartRef.current = null
      }
    }
  }, [])

  const handleReset = () => {
    setSensitivity(5)
    setMaxLevels(3)
    setShowSupport(true)
    setShowResistance(true)
  }

  const getSensitivityLabel = (value: number) => {
    switch (value) {
      case 3: return 'High (More Levels)'
      case 5: return 'Medium (Balanced)'
      case 7: return 'Low (Fewer Levels)'
      default: return 'Medium'
    }
  }

  return (
    <Card className="p-6 space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-semibold mb-2">Support & Resistance Lines</h3>
          <p className="text-sm text-muted-foreground">
            Horizontal price levels where the market tends to reverse or consolidate
          </p>
        </div>
        <div className="flex gap-2">
          <Badge variant="outline" className="bg-green-500/10 text-green-500 border-green-500/20">
            Support
          </Badge>
          <Badge variant="outline" className="bg-red-500/10 text-red-500 border-red-500/20">
            Resistance
          </Badge>
        </div>
      </div>

      {/* Controls */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 p-4 bg-muted/50 rounded-lg">
        {/* Sensitivity Slider */}
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <Label htmlFor="sensitivity" className="text-sm font-medium">
              Sensitivity
            </Label>
            <span className="text-xs text-muted-foreground">
              {getSensitivityLabel(sensitivity)}
            </span>
          </div>
          <Slider
            id="sensitivity"
            min={3}
            max={7}
            step={2}
            value={[sensitivity]}
            onValueChange={(value) => setSensitivity(value[0])}
            className="w-full"
          />
          <p className="text-xs text-muted-foreground">
            Lower values detect more levels, higher values detect stronger levels only
          </p>
        </div>

        {/* Max Levels Slider */}
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <Label htmlFor="maxLevels" className="text-sm font-medium">
              Max Levels
            </Label>
            <span className="text-xs text-muted-foreground">
              {maxLevels} {maxLevels === 1 ? 'level' : 'levels'}
            </span>
          </div>
          <Slider
            id="maxLevels"
            min={1}
            max={5}
            step={2}
            value={[maxLevels]}
            onValueChange={(value) => setMaxLevels(value[0])}
            className="w-full"
          />
          <p className="text-xs text-muted-foreground">
            Maximum number of support/resistance lines to display
          </p>
        </div>

        {/* Toggle Switches */}
        <div className="flex items-center justify-between p-3 bg-background rounded-md">
          <Label htmlFor="showSupport" className="text-sm font-medium flex items-center gap-2">
            <div className="h-3 w-3 rounded-full bg-green-500" />
            Show Support
          </Label>
          <Switch
            id="showSupport"
            checked={showSupport}
            onCheckedChange={setShowSupport}
          />
        </div>

        <div className="flex items-center justify-between p-3 bg-background rounded-md">
          <Label htmlFor="showResistance" className="text-sm font-medium flex items-center gap-2">
            <div className="h-3 w-3 rounded-full bg-red-500" />
            Show Resistance
          </Label>
          <Switch
            id="showResistance"
            checked={showResistance}
            onCheckedChange={setShowResistance}
          />
        </div>
      </div>

      {/* Chart */}
      <div className="relative">
        <div ref={chartContainerRef} className="w-full rounded-lg overflow-hidden border border-border" />
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-4 text-sm">
        <div className="flex items-center gap-2">
          <div className="h-0.5 w-8 bg-green-500 border-dashed" style={{ borderTop: '2px dashed' }} />
          <span className="text-muted-foreground">Support (price bounces up)</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="h-0.5 w-8 bg-red-500 border-dashed" style={{ borderTop: '2px dashed' }} />
          <span className="text-muted-foreground">Resistance (price bounces down)</span>
        </div>
      </div>

      {/* Reset Button */}
      <div className="flex justify-end">
        <Button onClick={handleReset} variant="outline" size="sm">
          Reset to Defaults
        </Button>
      </div>

      {/* Explanation */}
      <div className="p-4 bg-muted/30 rounded-lg space-y-2 text-sm">
        <h4 className="font-semibold">How it works:</h4>
        <ul className="space-y-1 text-muted-foreground list-disc list-inside">
          <li><strong className="text-green-500">Support</strong>: Price levels where buying pressure prevents further decline</li>
          <li><strong className="text-red-500">Resistance</strong>: Price levels where selling pressure prevents further rise</li>
          <li>Algorithm identifies local peaks and troughs with configurable sensitivity</li>
          <li>Levels cluster similar prices to show significant zones</li>
          <li>Useful for identifying entry/exit points and stop-loss placement</li>
        </ul>
      </div>
    </Card>
  )
}
