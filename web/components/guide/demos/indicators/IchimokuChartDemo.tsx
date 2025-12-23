'use client'

import { useEffect, useRef, useState } from 'react'
import { createChart, IChartApi, ISeriesApi, LineStyle, CandlestickData, LineData, AreaData } from 'lightweight-charts'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { generateSampleISXData, calculateIchimoku } from '@/lib/guide/indicators/calculations'
import { RotateCcw } from 'lucide-react'

interface PeriodPreset {
  name: string
  tenkan: number
  kijun: number
  senkouB: number
}

const PERIOD_PRESETS: PeriodPreset[] = [
  { name: 'Standard', tenkan: 9, kijun: 26, senkouB: 52 },
  { name: 'Fast', tenkan: 7, kijun: 22, senkouB: 44 }
]

export function IchimokuChartDemo() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const tenkanSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const kijunSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const senkouASeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const senkouBSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const chikouSeriesRef = useRef<ISeriesApi<'Line'> | null>(null)
  const cloudSeriesRef = useRef<ISeriesApi<'Area'>[]>([])

  // Period settings
  const [periods, setPeriods] = useState<PeriodPreset>(PERIOD_PRESETS[0])

  // Component visibility toggles
  const [showComponents, setShowComponents] = useState({
    tenkan: true,
    kijun: true,
    senkouA: true,
    senkouB: true,
    chikou: true,
    cloud: true
  })

  // Initialize and update chart
  useEffect(() => {
    if (!chartContainerRef.current) return

    // Generate sample data (need lots for displacement)
    const sampleData = generateSampleISXData(200)

    // Calculate Ichimoku components
    const ichimokuData = calculateIchimoku(
      sampleData.map(d => d.high),
      sampleData.map(d => d.low),
      periods.tenkan,
      periods.kijun,
      periods.senkouB
    )

    // Create chart if it doesn't exist
    if (!chartRef.current) {
      const chart = createChart(chartContainerRef.current, {
        layout: {
          background: { color: '#0B1120' },
          textColor: '#D1D5DB',
        },
        grid: {
          vertLines: { color: '#1F2937' },
          horzLines: { color: '#1F2937' },
        },
        width: chartContainerRef.current.clientWidth,
        height: 500,
        rightPriceScale: {
          borderColor: '#374151',
        },
        timeScale: {
          borderColor: '#374151',
          timeVisible: true,
        },
      })

      chartRef.current = chart

      // Handle resize
      const handleResize = () => {
        if (chartContainerRef.current && chartRef.current) {
          chartRef.current.applyOptions({
            width: chartContainerRef.current.clientWidth,
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

    // Clear existing series
    if (candlestickSeriesRef.current) {
      chartRef.current.removeSeries(candlestickSeriesRef.current)
      candlestickSeriesRef.current = null
    }
    if (tenkanSeriesRef.current) {
      chartRef.current.removeSeries(tenkanSeriesRef.current)
      tenkanSeriesRef.current = null
    }
    if (kijunSeriesRef.current) {
      chartRef.current.removeSeries(kijunSeriesRef.current)
      kijunSeriesRef.current = null
    }
    if (senkouASeriesRef.current) {
      chartRef.current.removeSeries(senkouASeriesRef.current)
      senkouASeriesRef.current = null
    }
    if (senkouBSeriesRef.current) {
      chartRef.current.removeSeries(senkouBSeriesRef.current)
      senkouBSeriesRef.current = null
    }
    if (chikouSeriesRef.current) {
      chartRef.current.removeSeries(chikouSeriesRef.current)
      chikouSeriesRef.current = null
    }
    cloudSeriesRef.current.forEach(series => {
      if (chartRef.current) {
        chartRef.current.removeSeries(series)
      }
    })
    cloudSeriesRef.current = []

    // Add candlestick series
    const candlestickSeries = chartRef.current.addCandlestickSeries({
      upColor: '#10B981',
      downColor: '#EF4444',
      borderVisible: false,
      wickUpColor: '#10B981',
      wickDownColor: '#EF4444',
    })
    candlestickSeriesRef.current = candlestickSeries
    candlestickSeries.setData(sampleData as CandlestickData[])

    // Add cloud fill first (background layer)
    if (showComponents.cloud && ichimokuData.cloudColor.length > 0) {
      // Split cloud data into segments by color
      const cloudSegments: { color: 'bullish' | 'bearish', data: AreaData[] }[] = []
      let currentSegment: { color: 'bullish' | 'bearish', data: AreaData[] } | null = null

      ichimokuData.senkouA.forEach((senkouA, index) => {
        const senkouB = ichimokuData.senkouB[index]
        const cloudColor = ichimokuData.cloudColor[index]

        if (senkouA === null || senkouB === null || cloudColor === null) return

        const dataPoint: AreaData = {
          time: sampleData[index].time,
          value: senkouA > senkouB ? senkouA : senkouB // Top of cloud
        }

        // Start new segment if color changed or first point
        if (!currentSegment || currentSegment.color !== cloudColor) {
          if (currentSegment && currentSegment.data.length > 0) {
            cloudSegments.push(currentSegment)
          }
          currentSegment = { color: cloudColor, data: [dataPoint] }
        } else {
          currentSegment.data.push(dataPoint)
        }
      })

      // Push last segment
      if (currentSegment && currentSegment.data.length > 0) {
        cloudSegments.push(currentSegment)
      }

      // Create AreaSeries for each segment
      cloudSegments.forEach(segment => {
        const areaSeries = chartRef.current!.addAreaSeries({
          topColor: segment.color === 'bullish'
            ? 'rgba(16, 185, 129, 0.4)'
            : 'rgba(239, 68, 68, 0.4)',
          bottomColor: segment.color === 'bullish'
            ? 'rgba(16, 185, 129, 0.1)'
            : 'rgba(239, 68, 68, 0.1)',
          lineVisible: false, // We draw Senkou lines separately
          priceLineVisible: false,
        })

        // Set data with proper bottom values
        const areaData: AreaData[] = segment.data.map((point, idx) => {
          // Find corresponding Senkou B value for bottom of cloud
          const originalIndex = ichimokuData.senkouA.findIndex(
            (_, i) => sampleData[i]?.time === point.time
          )
          const senkouA = ichimokuData.senkouA[originalIndex]
          const senkouB = ichimokuData.senkouB[originalIndex]

          return {
            time: point.time,
            value: senkouA !== null && senkouB !== null
              ? Math.max(senkouA, senkouB) // Top
              : point.value
          }
        })

        areaSeries.setData(areaData)
        cloudSeriesRef.current.push(areaSeries)
      })
    }

    // Add Senkou Span A (Leading Span A) - Green dashed
    if (showComponents.senkouA) {
      const senkouASeries = chartRef.current.addLineSeries({
        color: '#10B981',
        lineWidth: 2,
        lineStyle: LineStyle.Dashed,
        priceLineVisible: false,
      })
      senkouASeriesRef.current = senkouASeries

      const senkouAData: LineData[] = ichimokuData.senkouA
        .map((value, index) => ({
          time: sampleData[index].time,
          value: value ?? 0
        }))
        .filter(d => d.value !== 0)

      senkouASeries.setData(senkouAData)
    }

    // Add Senkou Span B (Leading Span B) - Orange dashed
    if (showComponents.senkouB) {
      const senkouBSeries = chartRef.current.addLineSeries({
        color: '#F59E0B',
        lineWidth: 2,
        lineStyle: LineStyle.Dashed,
        priceLineVisible: false,
      })
      senkouBSeriesRef.current = senkouBSeries

      const senkouBData: LineData[] = ichimokuData.senkouB
        .map((value, index) => ({
          time: sampleData[index].time,
          value: value ?? 0
        }))
        .filter(d => d.value !== 0)

      senkouBSeries.setData(senkouBData)
    }

    // Add Tenkan-sen (Conversion Line) - Blue solid
    if (showComponents.tenkan) {
      const tenkanSeries = chartRef.current.addLineSeries({
        color: '#2196F3',
        lineWidth: 2,
        priceLineVisible: false,
      })
      tenkanSeriesRef.current = tenkanSeries

      const tenkanData: LineData[] = ichimokuData.tenkan
        .map((value, index) => ({
          time: sampleData[index].time,
          value: value ?? 0
        }))
        .filter(d => d.value !== 0)

      tenkanSeries.setData(tenkanData)
    }

    // Add Kijun-sen (Base Line) - Red solid
    if (showComponents.kijun) {
      const kijunSeries = chartRef.current.addLineSeries({
        color: '#EF4444',
        lineWidth: 2,
        priceLineVisible: false,
      })
      kijunSeriesRef.current = kijunSeries

      const kijunData: LineData[] = ichimokuData.kijun
        .map((value, index) => ({
          time: sampleData[index].time,
          value: value ?? 0
        }))
        .filter(d => d.value !== 0)

      kijunSeries.setData(kijunData)
    }

    // Add Chikou Span (Lagging Span) - Purple dotted
    if (showComponents.chikou) {
      const chikouSeries = chartRef.current.addLineSeries({
        color: '#9C27B0',
        lineWidth: 2,
        lineStyle: LineStyle.Dotted,
        priceLineVisible: false,
      })
      chikouSeriesRef.current = chikouSeries

      const chikouData: LineData[] = ichimokuData.chikou
        .map((value, index) => ({
          time: sampleData[index].time,
          value: value ?? 0
        }))
        .filter(d => d.value !== 0)

      chikouSeries.setData(chikouData)
    }

    // Fit content
    chartRef.current.timeScale().fitContent()

  }, [periods, showComponents])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (chartRef.current) {
        chartRef.current.remove()
        chartRef.current = null
      }
    }
  }, [])

  const handlePeriodChange = (preset: PeriodPreset) => {
    setPeriods(preset)
  }

  const handleToggleComponent = (component: keyof typeof showComponents) => {
    setShowComponents(prev => ({
      ...prev,
      [component]: !prev[component]
    }))
  }

  const handleReset = () => {
    setPeriods(PERIOD_PRESETS[0])
    setShowComponents({
      tenkan: true,
      kijun: true,
      senkouA: true,
      senkouB: true,
      chikou: true,
      cloud: true
    })
  }

  return (
    <Card className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">Ichimoku Cloud Chart</h3>
          <p className="text-sm text-muted-foreground">
            Complete trend analysis with 5 components and cloud
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

      {/* Period Selector */}
      <div className="space-y-2">
        <Label className="text-sm font-medium">Period Preset</Label>
        <div className="flex flex-wrap gap-2">
          {PERIOD_PRESETS.map((preset) => (
            <Button
              key={preset.name}
              variant={periods.name === preset.name ? 'default' : 'outline'}
              size="sm"
              onClick={() => handlePeriodChange(preset)}
            >
              {preset.name}
              <span className="ml-2 text-xs opacity-70">
                ({preset.tenkan}/{preset.kijun}/{preset.senkouB})
              </span>
            </Button>
          ))}
        </div>
      </div>

      {/* Component Toggles */}
      <div className="space-y-3">
        <Label className="text-sm font-medium">Components</Label>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div className="flex items-center space-x-2">
            <Switch
              id="tenkan"
              checked={showComponents.tenkan}
              onCheckedChange={() => handleToggleComponent('tenkan')}
            />
            <Label htmlFor="tenkan" className="flex items-center gap-2 cursor-pointer">
              <div className="w-4 h-0.5 bg-[#2196F3]" />
              <span className="text-sm">Tenkan-sen</span>
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="kijun"
              checked={showComponents.kijun}
              onCheckedChange={() => handleToggleComponent('kijun')}
            />
            <Label htmlFor="kijun" className="flex items-center gap-2 cursor-pointer">
              <div className="w-4 h-0.5 bg-[#EF4444]" />
              <span className="text-sm">Kijun-sen</span>
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="senkouA"
              checked={showComponents.senkouA}
              onCheckedChange={() => handleToggleComponent('senkouA')}
            />
            <Label htmlFor="senkouA" className="flex items-center gap-2 cursor-pointer">
              <div className="w-4 h-0.5 bg-[#10B981] border-dashed border-t-2 border-[#10B981]" />
              <span className="text-sm">Senkou Span A</span>
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="senkouB"
              checked={showComponents.senkouB}
              onCheckedChange={() => handleToggleComponent('senkouB')}
            />
            <Label htmlFor="senkouB" className="flex items-center gap-2 cursor-pointer">
              <div className="w-4 h-0.5 bg-[#F59E0B] border-dashed border-t-2 border-[#F59E0B]" />
              <span className="text-sm">Senkou Span B</span>
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="chikou"
              checked={showComponents.chikou}
              onCheckedChange={() => handleToggleComponent('chikou')}
            />
            <Label htmlFor="chikou" className="flex items-center gap-2 cursor-pointer">
              <div className="w-4 h-0.5 bg-[#9C27B0] border-dotted border-t-2 border-[#9C27B0]" />
              <span className="text-sm">Chikou Span</span>
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="cloud"
              checked={showComponents.cloud}
              onCheckedChange={() => handleToggleComponent('cloud')}
            />
            <Label htmlFor="cloud" className="flex items-center gap-2 cursor-pointer">
              <div className="w-4 h-4 bg-gradient-to-b from-green-500/40 to-green-500/10 rounded" />
              <span className="text-sm">Cloud Fill</span>
            </Label>
          </div>
        </div>
      </div>

      {/* Chart Container */}
      <div ref={chartContainerRef} className="w-full h-[500px] rounded-lg overflow-hidden" />

      {/* Legend & Information */}
      <div className="space-y-4 pt-4 border-t">
        <div>
          <h4 className="text-sm font-semibold mb-3">Components</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
            <div className="flex items-start gap-2">
              <Badge variant="outline" className="bg-[#2196F3]/10 text-[#2196F3] border-[#2196F3]/50">
                Tenkan
              </Badge>
              <span className="text-muted-foreground">
                Conversion Line ({periods.tenkan}-period) - Short-term trend
              </span>
            </div>

            <div className="flex items-start gap-2">
              <Badge variant="outline" className="bg-[#EF4444]/10 text-[#EF4444] border-[#EF4444]/50">
                Kijun
              </Badge>
              <span className="text-muted-foreground">
                Base Line ({periods.kijun}-period) - Medium-term trend
              </span>
            </div>

            <div className="flex items-start gap-2">
              <Badge variant="outline" className="bg-[#10B981]/10 text-[#10B981] border-[#10B981]/50">
                Senkou A
              </Badge>
              <span className="text-muted-foreground">
                Leading Span A - Avg of Tenkan & Kijun (displaced +{periods.kijun})
              </span>
            </div>

            <div className="flex items-start gap-2">
              <Badge variant="outline" className="bg-[#F59E0B]/10 text-[#F59E0B] border-[#F59E0B]/50">
                Senkou B
              </Badge>
              <span className="text-muted-foreground">
                Leading Span B - {periods.senkouB}-period midpoint (displaced +{periods.kijun})
              </span>
            </div>

            <div className="flex items-start gap-2">
              <Badge variant="outline" className="bg-[#9C27B0]/10 text-[#9C27B0] border-[#9C27B0]/50">
                Chikou
              </Badge>
              <span className="text-muted-foreground">
                Lagging Span - Close price (displaced -{periods.kijun})
              </span>
            </div>
          </div>
        </div>

        <div>
          <h4 className="text-sm font-semibold mb-2">Cloud Significance</h4>
          <div className="space-y-2 text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-gradient-to-b from-green-500/40 to-green-500/10 rounded" />
              <span>
                <strong className="text-green-500">Bullish Cloud</strong> (Span A {'>'} Span B):
                Price above cloud = strong uptrend
              </span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 bg-gradient-to-b from-red-500/40 to-red-500/10 rounded" />
              <span>
                <strong className="text-red-500">Bearish Cloud</strong> (Span B {'>'} Span A):
                Price below cloud = strong downtrend
              </span>
            </div>
            <div className="pl-6">
              <span>
                Price inside cloud = consolidation/transition phase
              </span>
            </div>
          </div>
        </div>

        <div>
          <h4 className="text-sm font-semibold mb-2">Trading Signals</h4>
          <ul className="space-y-1 text-sm text-muted-foreground list-disc list-inside">
            <li>Tenkan crosses above Kijun + price above cloud = strong buy signal</li>
            <li>Price breaks above cloud = bullish breakout</li>
            <li>Chikou crosses above price = confirmation of uptrend</li>
            <li>Thick cloud = strong support/resistance</li>
            <li>Cloud twist (color change) = trend reversal warning</li>
          </ul>
        </div>

        <div className="bg-muted/50 p-4 rounded-lg">
          <p className="text-xs text-muted-foreground">
            <strong>Note:</strong> Ichimoku is a complete trading system.
            Senkou Spans are displaced +{periods.kijun} periods (projected into future),
            while Chikou Span is displaced -{periods.kijun} periods (lagging).
            The cloud acts as dynamic support/resistance and trend confirmation.
          </p>
        </div>
      </div>
    </Card>
  )
}
