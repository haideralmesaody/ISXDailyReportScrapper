/**
 * MultiChartContainer Component
 * Following CLAUDE.md patterns for professional multi-chart layouts
 * Manages multiple synchronized chart instances with resizable panels
 */

'use client'

import React, { useRef, useEffect, useState, useCallback, useMemo } from 'react'
import type {
  IChartApi,
  ColorType,
  CrosshairMode,
  LineStyle,
  PriceScaleMode
} from 'lightweight-charts'
import { ChartSynchronizer, ChartLayout, CHART_LAYOUTS, createSyncedChartOptions } from '@/lib/charts/chart-sync'
import { PanelResizeHandle } from '../PanelResizeHandle'
import { cn } from '@/lib/utils'
import type { IndicatorSettings, PanelLayout } from '@/lib/hooks/use-chart-state'

interface MultiChartContainerProps {
  ticker: string
  companyName?: string
  data: any[]
  indicators: IndicatorSettings
  panels: PanelLayout[]
  isLoading: boolean
  isFullscreen: boolean
  onTooltipUpdate: (data: any) => void
  onFullscreenToggle: () => void
  onPanelResize: (panelId: string, delta: number) => void
}

interface ChartRefs {
  main?: IChartApi
  volume?: IChartApi
  macd?: IChartApi
  rsi?: IChartApi
}

interface SeriesRefs {
  candlestick?: any
  volume?: any
  sma20?: any
  sma50?: any
  sma200?: any
  ema20?: any
  support?: any
  resistance?: any
}

// Dynamic import function for lightweight-charts
async function loadLightweightCharts() {
  try {
    // Import the module
    const module = await import('lightweight-charts')

    // Log the module structure to understand what's available
    console.log('[Chart] Module structure:', {
      module: Object.keys(module),
      hasDefault: !!module.default,
      createChartOnModule: typeof module.createChart,
      createChartOnDefault: typeof module.default?.createChart,
      defaultKeys: module.default ? Object.keys(module.default) : 'no default'
    })

    // Try to get createChart from the module (check both named export and default)
    const createChart = module.createChart || module.default?.createChart

    if (!createChart || typeof createChart !== 'function') {
      console.error('[Chart] createChart not found in module:', { module, createChart })
      throw new Error('createChart function not found in lightweight-charts module')
    }

    console.log('[Chart] Found createChart function:', typeof createChart)
    console.log('[Chart] Available enums:', {
      ColorType: typeof module.ColorType,
      CrosshairMode: typeof module.CrosshairMode,
      LineStyle: typeof module.LineStyle,
      PriceScaleMode: typeof module.PriceScaleMode
    })

    return module
  } catch (error) {
    console.error('[Chart] Failed to load lightweight-charts:', error)
    throw error
  }
}

export function MultiChartContainer({
  ticker,
  companyName,
  data,
  indicators,
  panels,
  isLoading,
  isFullscreen,
  onTooltipUpdate,
  onFullscreenToggle,
  onPanelResize
}: MultiChartContainerProps) {
  console.log('[Chart] === MULTICONTAINER COMPONENT RENDER ===', {
    ticker,
    dataLength: data?.length,
    isLoading,
    hasIndicators: !!indicators,
    panelCount: panels?.length
  })
  // Container refs for each chart
  const mainChartRef = useRef<HTMLDivElement>(null)
  const volumeChartRef = useRef<HTMLDivElement>(null)
  const macdChartRef = useRef<HTMLDivElement>(null)
  const rsiChartRef = useRef<HTMLDivElement>(null)

  // Chart and series references
  const chartRefs = useRef<ChartRefs>({})
  const seriesRefs = useRef<SeriesRefs>({})
  const synchronizer = useRef<ChartSynchronizer | null>(null)
  const isInitialized = useRef<boolean>(false)

  // Store dynamically loaded enums
  const chartEnums = useRef<{
    ColorType?: any
    CrosshairMode?: any
    LineStyle?: any
    PriceScaleMode?: any
  }>({})

  // Get visible panels and their sizes from state
  const visiblePanels = useMemo(() =>
    panels?.filter(panel => panel.visible) || [],
    [panels]
  )

  // Determine current layout based on visible indicators
  const currentLayout: ChartLayout = useMemo(() => {
    if (indicators.showMACD && indicators.showRSI && indicators.showVolume) {
      return 'full'
    } else if (indicators.showMACD && indicators.showVolume) {
      return 'mainVolumeMACD'
    } else if (indicators.showRSI && indicators.showVolume) {
      return 'mainVolumeRSI'
    } else if (indicators.showVolume) {
      return 'mainWithVolume'
    } else {
      return 'mainOnly'
    }
  }, [indicators.showMACD, indicators.showRSI, indicators.showVolume])

  // Theme colors (matching existing implementation)
  const themeColors = useMemo(() => ({
    backgroundColor: 'transparent',
    textColor: 'hsl(var(--foreground))',
    gridColor: 'hsl(var(--border))',
    candlestickUpColor: '#22c55e',
    candlestickDownColor: '#ef4444',
    volumeUpColor: '#22c55e',
    volumeDownColor: '#ef4444',
    sma20Color: '#3b82f6',
    sma50Color: '#f97316',
    sma200Color: '#8b5cf6',
    ema20Color: '#06b6d4',
    supportColor: '#22c55e',
    resistanceColor: '#ef4444'
  }), [])

  // Function to create chart options with loaded enums
  const createBaseChartOptions = useCallback((enums: any) => {
    // Ensure enums are available, fallback to default values
    const colorType = enums?.ColorType?.Solid || 0
    const crosshairMode = enums?.CrosshairMode?.Normal || 0
    const lineStyle = enums?.LineStyle?.Dashed || 1

    return {
      layout: {
        background: { type: colorType, color: themeColors.backgroundColor },
        textColor: themeColors.textColor
      },
      grid: {
        vertLines: { color: themeColors.gridColor },
        horzLines: { color: themeColors.gridColor }
      },
      crosshair: {
        mode: crosshairMode,
        vertLine: {
          color: themeColors.textColor,
          width: 1,
          style: lineStyle
        },
        horzLine: {
          color: themeColors.textColor,
          width: 1,
          style: lineStyle
        }
      },
      handleScroll: {
        mouseWheel: true,
        pressedMouseMove: true
      },
      handleScale: {
        mouseWheel: true,
        pinch: true
      }
    }
  }, [themeColors])

  // Cleanup function
  const cleanupCharts = useCallback(() => {
    console.log('[Chart] Cleaning up charts')

    // Destroy synchronizer first
    if (synchronizer.current) {
      synchronizer.current.destroy()
      synchronizer.current = null
    }

    // Remove and dispose all chart instances
    Object.values(chartRefs.current).forEach(chart => {
      if (chart) {
        try {
          chart.remove()
        } catch (error) {
          console.warn('[Chart] Error removing chart:', error)
        }
      }
    })

    // Clear all references
    chartRefs.current = {}
    seriesRefs.current = {}
    isInitialized.current = false
  }, [])

  // Initialize charts with async loading
  const initializeCharts = useCallback(async () => {
    console.log('[Chart] === INITIALIZECHARTS FUNCTION CALLED ===')
    console.log('[Chart] Current state:', {
      hasMainRef: !!mainChartRef.current,
      ticker,
      dataLength: data?.length,
      isInitialized: isInitialized.current,
      visiblePanelsCount: visiblePanels.length
    })

    if (!mainChartRef.current) {
      console.warn('[Chart] No main chart container ref available')
      return
    }

    // Check if container has valid dimensions - critical for chart creation
    const containerWidth = mainChartRef.current.clientWidth
    const containerHeight = mainChartRef.current.clientHeight

    console.log('[Chart] Container dimensions check:', {
      width: containerWidth,
      height: containerHeight,
      element: mainChartRef.current
    })

    if (containerWidth === 0 || containerHeight === 0) {
      console.log('[Chart] Container has no dimensions, waiting for layout...')
      // Use requestAnimationFrame to wait for next paint cycle
      await new Promise(resolve => requestAnimationFrame(resolve))

      // Check again after layout
      const newWidth = mainChartRef.current.clientWidth
      const newHeight = mainChartRef.current.clientHeight

      if (newWidth === 0 || newHeight === 0) {
        console.warn('[Chart] Container still has no dimensions, retrying in 100ms')
        setTimeout(() => initializeCharts(), 100)
        return
      }
    }

    console.log('[Chart] Initializing charts for panels:', visiblePanels.map(p => p.id))

    // Clean up existing charts first
    cleanupCharts()

    try {
      // Load lightweight-charts dynamically
      console.log('[Chart] Loading lightweight-charts library...')
      const lightweightCharts = await loadLightweightCharts()

      // Store the entire module to preserve context
      const LightweightCharts = lightweightCharts

      if (!LightweightCharts.createChart || typeof LightweightCharts.createChart !== 'function') {
        throw new Error('createChart function not found in loaded module')
      }

      // Store enums for use throughout the component
      chartEnums.current = {
        ColorType: lightweightCharts.ColorType,
        CrosshairMode: lightweightCharts.CrosshairMode,
        LineStyle: lightweightCharts.LineStyle,
        PriceScaleMode: lightweightCharts.PriceScaleMode
      }

      // Create new synchronizer
      synchronizer.current = new ChartSynchronizer()

      // Initialize main chart
      const finalWidth = mainChartRef.current.clientWidth || 800
      const finalHeight = mainChartRef.current.clientHeight || 400

      console.log('[Chart] Creating main chart with dimensions:', {
        width: finalWidth,
        height: finalHeight,
        container: mainChartRef.current
      })

      // Create chart options with loaded enums
      const baseChartOptions = createBaseChartOptions(chartEnums.current)

      const mainChart = LightweightCharts.createChart(mainChartRef.current, {
        ...createSyncedChartOptions(baseChartOptions, 'main'),
        width: finalWidth,
        height: finalHeight
      })

      // Validate that createChart returned a proper chart instance
      if (!mainChart || typeof mainChart.addCandlestickSeries !== 'function') {
        console.error('[Chart] createChart failed to return valid chart instance:', {
          mainChart,
          type: typeof mainChart,
          hasAddCandlestickSeries: !!(mainChart && typeof mainChart.addCandlestickSeries === 'function'),
          methods: mainChart ? Object.getOwnPropertyNames(mainChart) : 'no chart object'
        })
        throw new Error('Chart initialization failed: createChart did not return a valid TradingView chart instance')
      }

      console.log('[Chart] Main chart created successfully, type:', typeof mainChart)
      chartRefs.current.main = mainChart
      synchronizer.current.addChart(mainChart, 'main', 'main')

      // Add candlestick series to main chart
      const candlestickSeries = mainChart.addCandlestickSeries({
        upColor: themeColors.candlestickUpColor,
        downColor: themeColors.candlestickDownColor,
        borderUpColor: themeColors.candlestickUpColor,
        borderDownColor: themeColors.candlestickDownColor,
        wickUpColor: themeColors.candlestickUpColor,
        wickDownColor: themeColors.candlestickDownColor
      })
      seriesRefs.current.candlestick = candlestickSeries

    // Add moving averages to main chart
    if (indicators.showSMA20) {
      seriesRefs.current.sma20 = mainChart.addLineSeries({
        color: themeColors.sma20Color,
        lineWidth: 2,
        title: 'SMA 20'
      })
    }

    if (indicators.showSMA50) {
      seriesRefs.current.sma50 = mainChart.addLineSeries({
        color: themeColors.sma50Color,
        lineWidth: 2,
        title: 'SMA 50'
      })
    }

    if (indicators.showSMA200) {
      seriesRefs.current.sma200 = mainChart.addLineSeries({
        color: themeColors.sma200Color,
        lineWidth: 2,
        title: 'SMA 200'
      })
    }

    if (indicators.showEMA20) {
      seriesRefs.current.ema20 = mainChart.addLineSeries({
        color: themeColors.ema20Color,
        lineWidth: 2,
        lineStyle: chartEnums.current.LineStyle?.Dashed,
        title: 'EMA 20'
      })
    }

      // Initialize charts for all visible panels (except main, which is already done)
      visiblePanels.forEach(panel => {
        if (panel.id === 'main') return // Already handled above

        const ref = panel.id === 'volume' ? volumeChartRef :
                    panel.id === 'macd' ? macdChartRef :
                    panel.id === 'rsi' ? rsiChartRef : null

        if (ref?.current) {
          const panelWidth = ref.current.clientWidth || 800
          const panelHeight = ref.current.clientHeight || 200

          const chart = LightweightCharts.createChart(ref.current, {
            ...createSyncedChartOptions(baseChartOptions, panel.id as any),
            width: panelWidth,
            height: panelHeight
          })

          // Validate chart for indicators too
          if (!chart) {
            console.error(`[Chart] Failed to create chart for panel ${panel.id}`)
            return
          }

          chartRefs.current[panel.id as keyof ChartRefs] = chart
          synchronizer.current.addChart(chart, panel.id, panel.id as any)

          // Add series based on panel type
          if (panel.id === 'volume') {
            const volumeSeries = chart.addHistogramSeries({
              color: themeColors.volumeUpColor,
              priceFormat: { type: 'volume' }
            })
            seriesRefs.current.volume = volumeSeries
          }
          // MACD and RSI series would be added here in future
        }
      })

      // Mark as initialized
      isInitialized.current = true

      console.log('[Chart] Charts initialized successfully')

    } catch (error) {
      console.error('[Chart] Chart initialization failed:', error)

      // Clean up on error
      cleanupCharts()

      // Set error state or show user feedback
      // For now, log the error - in future we could set an error state
      throw error
    }

  }, [createBaseChartOptions, themeColors, indicators, visiblePanels, cleanupCharts])

  // Set chart data
  const setChartData = useCallback(() => {
    if (!data || data.length === 0 || !isInitialized.current) {
      console.log('[Chart] Skipping setChartData - no data or not initialized')
      return
    }

    console.log('[Chart] Setting chart data:', {
      recordCount: data.length,
      firstDate: data[0]?.date,
      lastDate: data[data.length - 1]?.date
    })

    // Transform data for charts - convert date strings to Unix timestamps (seconds)
    const candlestickData = data.map(item => ({
      time: Math.floor(new Date(item.date).getTime() / 1000),
      open: item.open,
      high: item.high,
      low: item.low,
      close: item.close
    }))

    const volumeData = data.map(item => ({
      time: Math.floor(new Date(item.date).getTime() / 1000),
      value: item.volume,
      color: item.close >= item.open ? themeColors.volumeUpColor : themeColors.volumeDownColor
    }))

    // Set main chart data
    if (seriesRefs.current.candlestick) {
      seriesRefs.current.candlestick.setData(candlestickData)
    }

    // Set volume data
    if (seriesRefs.current.volume) {
      seriesRefs.current.volume.setData(volumeData)
    }

    // Set moving average data (simplified for now)
    // In production, you'd calculate these from the price data
    if (seriesRefs.current.sma20 && data[0]?.sma20) {
      const sma20Data = data.map(item => ({
        time: Math.floor(new Date(item.date).getTime() / 1000),
        value: item.sma20
      })).filter(item => item.value != null)
      seriesRefs.current.sma20.setData(sma20Data)
    }

  }, [data, themeColors])

  // Handle resize
  const handleResize = useCallback(() => {
    Object.values(chartRefs.current).forEach(chart => {
      if (chart) {
        const container = chart.chartElement()?.parentElement
        if (container) {
          chart.applyOptions({
            width: container.clientWidth,
            height: container.clientHeight
          })
        }
      }
    })
  }, [])

  // Handle panel resize
  const handlePanelResize = useCallback((panelId: string, delta: number) => {
    onPanelResize(panelId, delta / 100) // Convert to decimal for state manager

    // Trigger chart resize after state update
    setTimeout(handleResize, 0)
  }, [onPanelResize, handleResize])

  // Initialize charts on mount and when panels/indicators change
  useEffect(() => {
    console.log('[Chart] useEffect triggered for initialization with deps:', {
      initializeCharts: typeof initializeCharts,
      cleanupCharts: typeof cleanupCharts,
      ticker,
      dataLength: data?.length
    })

    let isMounted = true

    const initAsync = async () => {
      try {
        console.log('[Chart] Starting async initialization...')

        // Remove the delay - initialize immediately
        if (isMounted) {
          console.log('[Chart] Component still mounted, calling initializeCharts...')
          await initializeCharts()
        } else {
          console.log('[Chart] Component unmounted before initialization')
        }
      } catch (error) {
        console.error('[Chart] Failed to initialize charts:', error)
        console.error('[Chart] Error stack:', error.stack)
      }
    }

    if (isMounted) {
      console.log('[Chart] Starting initialization process...')
      initAsync()
    }

    return () => {
      console.log('[Chart] Cleanup function called')
      isMounted = false
      cleanupCharts()
    }
  }, [initializeCharts, cleanupCharts, ticker, data])

  // Set data when charts are initialized and data changes
  useEffect(() => {
    if (isInitialized.current && data && data.length > 0) {
      setChartData()
    }
  }, [data, setChartData])

  // Handle window resize
  useEffect(() => {
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [handleResize])

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-muted-foreground">Loading chart...</div>
      </div>
    )
  }

  return (
    <div className="h-full w-full relative">
      {visiblePanels.map((panel, index) => {
        const isLast = index === visiblePanels.length - 1
        const chartRef = panel.id === 'main' ? mainChartRef :
                        panel.id === 'volume' ? volumeChartRef :
                        panel.id === 'macd' ? macdChartRef :
                        panel.id === 'rsi' ? rsiChartRef : null

        // Calculate cumulative position for resize handle (ensure it doesn't exceed bounds)
        const cumulativePosition = Math.min(
          visiblePanels
            .slice(0, index + 1)
            .reduce((sum, p) => sum + p.height, 0) * 100,
          100
        )

        return (
          <div key={panel.id}>
            {/* Chart Panel */}
            <div
              className="relative"
              style={{ height: `${panel.height * 100}%` }}
            >
              {chartRef && <div ref={chartRef} className="w-full h-full" />}
            </div>

            {/* Resize Handle (except for last panel) */}
            {!isLast && (
              <PanelResizeHandle
                position={cumulativePosition}
                onResize={(delta) => handlePanelResize(panel.id, delta)}
                label={`Resize ${panel.label}`}
              />
            )}
          </div>
        )
      })}

      {/* Chart Info Overlay */}
      <div className="absolute top-4 left-4 pointer-events-none">
        <div className="bg-background/80 backdrop-blur-sm rounded px-3 py-2 border">
          <div className="font-semibold text-sm">{ticker}</div>
          {companyName && (
            <div className="text-xs text-muted-foreground">{companyName}</div>
          )}
        </div>
      </div>

      {/* Fullscreen Toggle */}
      <button
        onClick={onFullscreenToggle}
        className="absolute top-4 right-4 p-2 bg-background/80 backdrop-blur-sm rounded border hover:bg-muted transition-colors"
        title={isFullscreen ? 'Exit Fullscreen' : 'Enter Fullscreen'}
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          {isFullscreen ? (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 9V4.5M9 9H4.5M9 9L3.5 3.5M15 9v-4.5M15 9h4.5M15 9l5.5-5.5M9 15v4.5M9 15H4.5M9 15l-5.5 5.5M15 15v4.5M15 15h4.5M15 15l5.5 5.5" />
          ) : (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z" />
          )}
        </svg>
      </button>
    </div>
  )
}