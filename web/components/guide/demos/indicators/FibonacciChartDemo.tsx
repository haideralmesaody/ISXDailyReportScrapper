'use client'

import { useEffect, useRef, useState } from 'react'
import { createChart, IChartApi, ISeriesApi, LineStyle, CandlestickSeries } from 'lightweight-charts'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Slider } from '@/components/ui/slider'
import { Label } from '@/components/ui/label'
import { generateSampleISXData, calculateFibonacciLevels } from '@/lib/guide/indicators/calculations'
import { Info, RotateCcw } from 'lucide-react'

interface FibLevel {
  ratio: number
  label: string
  color: string
  enabled: boolean
}

const FIB_LEVELS: FibLevel[] = [
  { ratio: 0, label: '0% (Low)', color: '#6b7280', enabled: true },
  { ratio: 0.236, label: '23.6%', color: '#60a5fa', enabled: true },
  { ratio: 0.382, label: '38.2%', color: '#3b82f6', enabled: true },
  { ratio: 0.5, label: '50%', color: '#a855f7', enabled: true },
  { ratio: 0.618, label: '61.8% (Golden)', color: '#f59e0b', enabled: true },
  { ratio: 0.786, label: '78.6%', color: '#ef4444', enabled: true },
  { ratio: 1.0, label: '100% (High)', color: '#6b7280', enabled: true }
]

export function FibonacciChartDemo() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const fibSeriesRefs = useRef<Map<string, ISeriesApi<'Line'>>>(new Map())

  const [lookbackPeriod, setLookbackPeriod] = useState([50])
  const [enabledLevels, setEnabledLevels] = useState<Record<string, boolean>>({
    '0%': true,
    '23.6%': true,
    '38.2%': true,
    '50%': true,
    '61.8%': true,
    '78.6%': true,
    '100%': true
  })
  const [swingPoints, setSwingPoints] = useState<{
    high: number
    low: number
    range: number
  } | null>(null)

  // Reset to defaults
  const handleReset = () => {
    setLookbackPeriod([50])
    setEnabledLevels({
      '0%': true,
      '23.6%': true,
      '38.2%': true,
      '50%': true,
      '61.8%': true,
      '78.6%': true,
      '100%': true
    })
  }

  // Toggle level visibility
  const toggleLevel = (label: string) => {
    setEnabledLevels(prev => ({
      ...prev,
      [label]: !prev[label]
    }))
  }

  // Initialize and update chart
  useEffect(() => {
    if (!chartContainerRef.current) return

    // Generate sample data
    const data = generateSampleISXData(150)

    // Calculate Fibonacci levels
    const fibData = calculateFibonacciLevels(data, lookbackPeriod[0])

    // Store swing point information
    setSwingPoints({
      high: fibData.swingHigh.price,
      low: fibData.swingLow.price,
      range: fibData.swingHigh.price - fibData.swingLow.price
    })

    // Create chart if it doesn't exist
    if (!chartRef.current) {
      const chart = createChart(chartContainerRef.current, {
        width: chartContainerRef.current.clientWidth,
        height: 450,
        layout: {
          background: { color: '#0f1419' },
          textColor: '#d1d5db'
        },
        grid: {
          vertLines: { color: '#1f2937' },
          horzLines: { color: '#1f2937' }
        },
        crosshair: {
          mode: 1,
          vertLine: {
            width: 1,
            color: '#6b7280',
            style: LineStyle.Dashed
          },
          horzLine: {
            width: 1,
            color: '#6b7280',
            style: LineStyle.Dashed
          }
        },
        timeScale: {
          borderColor: '#2B2B43',
          timeVisible: true
        },
        rightPriceScale: {
          borderColor: '#2B2B43'
        }
      })

      chartRef.current = chart

      // Handle resize
      const handleResize = () => {
        if (chartContainerRef.current && chartRef.current) {
          chartRef.current.applyOptions({
            width: chartContainerRef.current.clientWidth
          })
        }
      }

      window.addEventListener('resize', handleResize)

      // Cleanup
      return () => {
        window.removeEventListener('resize', handleResize)
        if (chartRef.current) {
          chartRef.current.remove()
          chartRef.current = null
        }
      }
    }

    // Clear existing Fibonacci series
    fibSeriesRefs.current.forEach(series => {
      if (chartRef.current) {
        chartRef.current.removeSeries(series)
      }
    })
    fibSeriesRefs.current.clear()

    // Remove candlestick series if it exists
    if (candlestickSeriesRef.current && chartRef.current) {
      chartRef.current.removeSeries(candlestickSeriesRef.current)
      candlestickSeriesRef.current = null
    }

    // Add candlestick series
    if (chartRef.current) {
      candlestickSeriesRef.current = chartRef.current.addCandlestickSeries({
        upColor: '#10b981',
        downColor: '#ef4444',
        borderUpColor: '#10b981',
        borderDownColor: '#ef4444',
        wickUpColor: '#10b981',
        wickDownColor: '#ef4444'
      })
      candlestickSeriesRef.current.setData(data)

      // Add Fibonacci level lines
      FIB_LEVELS.forEach((fibLevel) => {
        const levelKey = fibLevel.label.split(' ')[0] // Extract "0%", "23.6%", etc.

        if (enabledLevels[levelKey] && chartRef.current) {
          const levelPrice = fibData.swingLow.price + (fibData.swingHigh.price - fibData.swingLow.price) * fibLevel.ratio

          // Create horizontal line data
          const lineData = data.map(point => ({
            time: point.time,
            value: levelPrice
          }))

          const series = chartRef.current.addLineSeries({
            color: fibLevel.color,
            lineWidth: 2,
            lineStyle: LineStyle.Dashed,
            priceLineVisible: false,
            lastValueVisible: true,
            title: fibLevel.label
          })

          series.setData(lineData)
          fibSeriesRefs.current.set(levelKey, series)
        }
      })

      // Fit content to view
      chartRef.current.timeScale().fitContent()
    }
  }, [lookbackPeriod, enabledLevels])

  return (
    <Card className="p-6 space-y-6">
      {/* Header with info */}
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-semibold mb-2">Fibonacci Retracement Demo</h3>
          <p className="text-sm text-muted-foreground">
            Interactive demonstration of Fibonacci retracement levels calculated from swing high/low points
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={handleReset}
          className="gap-2"
        >
          <RotateCcw className="h-4 w-4" />
          Reset
        </Button>
      </div>

      {/* Controls */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Lookback Period Slider */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <Label>Lookback Period: {lookbackPeriod[0]} days</Label>
          </div>
          <Slider
            value={lookbackPeriod}
            onValueChange={setLookbackPeriod}
            min={30}
            max={80}
            step={10}
            className="w-full"
          />
          <p className="text-xs text-muted-foreground">
            Period to identify swing high and swing low points
          </p>
        </div>

        {/* Swing Points Info */}
        {swingPoints && (
          <div className="space-y-2">
            <Label>Swing Points</Label>
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline" className="gap-2">
                <span className="text-xs text-muted-foreground">High:</span>
                <span className="font-mono">{swingPoints.high.toFixed(3)}</span>
              </Badge>
              <Badge variant="outline" className="gap-2">
                <span className="text-xs text-muted-foreground">Low:</span>
                <span className="font-mono">{swingPoints.low.toFixed(3)}</span>
              </Badge>
              <Badge variant="outline" className="gap-2">
                <span className="text-xs text-muted-foreground">Range:</span>
                <span className="font-mono">{swingPoints.range.toFixed(3)}</span>
              </Badge>
            </div>
            <p className="text-xs text-muted-foreground">
              Fibonacci levels calculated from these swing points
            </p>
          </div>
        )}
      </div>

      {/* Level Toggles */}
      <div>
        <Label className="mb-3 block">Fibonacci Levels</Label>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-3">
          {FIB_LEVELS.map((fibLevel) => {
            const levelKey = fibLevel.label.split(' ')[0]
            return (
              <div
                key={levelKey}
                className="flex items-center gap-2 p-2 rounded-md border border-border hover:bg-accent/50 transition-colors"
              >
                <Switch
                  checked={enabledLevels[levelKey]}
                  onCheckedChange={() => toggleLevel(levelKey)}
                  id={`fib-${levelKey}`}
                />
                <Label
                  htmlFor={`fib-${levelKey}`}
                  className="flex items-center gap-2 cursor-pointer text-xs font-medium"
                >
                  <div
                    className="w-3 h-3 rounded-sm"
                    style={{ backgroundColor: fibLevel.color }}
                  />
                  <span>{fibLevel.label}</span>
                </Label>
              </div>
            )
          })}
        </div>
      </div>

      {/* Chart Container */}
      <div
        ref={chartContainerRef}
        className="w-full rounded-lg overflow-hidden border border-border"
      />

      {/* Educational Note */}
      <div className="flex gap-3 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
        <Info className="h-5 w-5 text-blue-500 flex-shrink-0 mt-0.5" />
        <div className="space-y-2 text-sm">
          <p className="font-medium text-blue-400">Understanding Fibonacci Retracements</p>
          <ul className="space-y-1 text-muted-foreground">
            <li>
              <strong className="text-foreground">Golden Ratio (61.8%):</strong> Most important retracement level, often acts as strong support/resistance
            </li>
            <li>
              <strong className="text-foreground">Primary Levels (38.2%, 61.8%):</strong> Key areas where price often reverses in uptrends
            </li>
            <li>
              <strong className="text-foreground">50% Level:</strong> Psychological midpoint (not a Fibonacci number but widely respected)
            </li>
            <li>
              <strong className="text-foreground">Deep Retracement (78.6%):</strong> Indicates strong pullback, watch for trend continuation or reversal
            </li>
            <li>
              <strong className="text-foreground">Usage:</strong> In uptrends, retracements show potential buy zones; in downtrends, potential sell zones
            </li>
          </ul>
        </div>
      </div>

      {/* Legend */}
      <div className="border-t border-border pt-4">
        <Label className="mb-3 block">Color Legend</Label>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-[#6b7280]" style={{ borderTop: '2px dashed' }} />
            <span className="text-muted-foreground">Boundaries (0%, 100%)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-[#60a5fa]" style={{ borderTop: '2px dashed' }} />
            <span className="text-muted-foreground">Light Blue (23.6%)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-[#3b82f6]" style={{ borderTop: '2px dashed' }} />
            <span className="text-muted-foreground">Blue (38.2%)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-[#a855f7]" style={{ borderTop: '2px dashed' }} />
            <span className="text-muted-foreground">Purple (50%)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-[#f59e0b]" style={{ borderTop: '2px dashed' }} />
            <span className="text-muted-foreground">Orange (61.8% Golden)</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-0.5 bg-[#ef4444]" style={{ borderTop: '2px dashed' }} />
            <span className="text-muted-foreground">Red (78.6%)</span>
          </div>
        </div>
      </div>
    </Card>
  )
}
