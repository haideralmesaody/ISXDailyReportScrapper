/**
 * ChartContainer - Main chart orchestrator using component-based architecture
 * Following TradingView official React patterns
 *
 * Architecture Benefits:
 * - React manages series lifecycle (no manual remove/re-add)
 * - Y-axis stability through conditional rendering
 * - Easy to add new indicators (just add component)
 * - Self-documenting structure
 */

'use client'

import { useRef, useEffect, useState, useCallback, useMemo } from 'react'
import { createChart, ColorType, CrosshairMode, LineStyle, PriceScaleMode } from 'lightweight-charts'
import type { IChartApi, ISeriesApi } from 'lightweight-charts'
import { ChartProvider, type ProcessedChartData, type ThemeColors } from './ChartContext'
import { logger } from '@/lib/utils/logger'
import { ChartTooltip } from '@/components/analysis/ChartTooltip'
import { buildEnhancedTooltipData } from '@/lib/utils/tooltip-helpers'
import { MainSeriesComponent } from './MainSeriesComponent'
import { SMAIndicator } from './indicators/SMAIndicator'
import { EMAIndicator } from './indicators/EMAIndicator'
import { SupportResistanceIndicator } from './indicators/SupportResistanceIndicator'
import { VolumePane } from './indicators/VolumePane'
import { OBVPane } from './indicators/OBVPane'
import { VWAPOverlay } from './indicators/VWAPOverlay'
import { ParabolicSAROverlay } from './indicators/ParabolicSAROverlay'
import { FibonacciRetracement } from './indicators/FibonacciRetracement'
import { DonchianChannels } from './indicators/DonchianChannels'
import { KeltnerChannels } from './indicators/KeltnerChannels'
import { MACDPane } from './indicators/MACDPane'
import { RSIPane } from './indicators/RSIPane'
import { MFIPane } from './indicators/MFIPane'
import { StochasticPane } from './indicators/StochasticPane'
import { GeneralStochasticPane } from './indicators/GeneralStochasticPane'
import { ADXPane } from './indicators/ADXPane'
import { ATRPane } from './indicators/ATRPane'
import { CCIPane } from './indicators/CCIPane'
import { WilliamsRPane } from './indicators/WilliamsRPane'
import { MomentumPane } from './indicators/MomentumPane'
import { ROCPane } from './indicators/ROCPane'
import { BollingerBandsOverlay } from './indicators/BollingerBandsOverlay'
import { GeneralBollingerBandsOverlay } from './indicators/GeneralBollingerBandsOverlay'
import { IchimokuOverlay } from './indicators/IchimokuOverlay'
import { DrawingLayer } from '../drawings/DrawingLayer'
import { DrawingInteraction } from '../drawings/DrawingInteraction'
import { DrawingToolbar } from '../drawings/DrawingToolbar'
import { DrawingThemeSync } from '../drawings/DrawingThemeSync'
import { DrawingPersistence } from '../drawings/DrawingPersistence'
import type { IndicatorSettings, ChartType, Timeframe } from '@/lib/hooks/use-chart-state'
import { usePaneRegistry } from '@/lib/hooks/use-pane-registry'
import { useMovingAverageSettings } from '@/lib/hooks/use-indicator-settings'
import { usePaneTransitionManager } from '@/lib/hooks/use-pane-transition-manager'
import { useVWAPSettings } from '@/lib/hooks/use-vwap-settings'
import type { IndicatorData } from './ChartContext'
import type { TickerHistoricalData } from '@/types/analysis'

interface ChartContainerProps {
  ticker: string
  chartData: ProcessedChartData
  historicalData: TickerHistoricalData[]
  indicators: IndicatorSettings
  indicatorActivationOrder: Record<string, number>  // NEW: Track activation timestamps
  theme: ThemeColors
  onCrosshairMove?: (param: any, chart: IChartApi) => void
  onChartReady?: (chart: IChartApi) => void
  chartType?: ChartType  // Chart type (candlestick/line/area/bar)
  timeframe?: Timeframe  // Timeframe (1D/1W/1M/3M/1Y/MAX)
  indicatorData?: IndicatorData  // API-calculated indicator data
}

export function ChartContainer({
  ticker,
  chartData,
  historicalData,
  indicators,
  indicatorActivationOrder,
  theme,
  onCrosshairMove,
  onChartReady,
  chartType = 'candlestick',
  timeframe = '1D',
  indicatorData = {}
}: ChartContainerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)

  // ✅ TradingView Best Practice: Use ref for performance + state for React updates
  // - Ref: Fast internal access (no re-renders)
  // - State: Proper React data flow (context consumers get updates)
  // - Pattern from TradingView Advanced React Example
  const candlestickSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const [candlestickSeries, setCandlestickSeries] = useState<ISeriesApi<'Candlestick'> | null>(null)

  const [chartReady, setChartReady] = useState(false)
  const [containerWidth, setContainerWidth] = useState(0)
  const [containerHeight, setContainerHeight] = useState(600)
  const [tooltipState, setTooltipState] = useState<{
    visible: boolean
    data: ReturnType<typeof buildEnhancedTooltipData> | null
    position: { x: number; y: number }
  }>({ visible: false, data: null, position: { x: 0, y: 0 } })

  // Get moving average settings for General SMA/EMA
  const { settings: maSettings } = useMovingAverageSettings()

  // Get VWAP settings for dynamic key (forces remount on settings change)
  const { settings: vwapSettings } = useVWAPSettings()

  // Stable callback for series creation (prevents infinite loop)
  const handleSeriesCreated = useCallback((series: ISeriesApi<'Candlestick'>) => {
    // Guard: prevent duplicate calls (onCreated might be called multiple times)
    if (candlestickSeriesRef.current === series) {
      logger.log('[ChartContainer] Series already set, skipping duplicate call')
      return
    }

    logger.log('[ChartContainer] Setting candlestick series (ref + state)')
    candlestickSeriesRef.current = series  // For internal perf access
    setCandlestickSeries(series)  // For React context updates
  }, [])

  // Initialize chart ONCE (no dependencies - stable chart instance)
  useEffect(() => {
    if (!containerRef.current) {
      logger.error('[ChartContainer] Container ref is null')
      return
    }

    // Initialize chart with retry logic for container dimensions
    const attemptInitialization = () => {
      if (!containerRef.current) return

      const { clientWidth, clientHeight } = containerRef.current
      if (clientWidth === 0 || clientHeight === 0) {
        logger.error('[ChartContainer] Container has no dimensions, retrying...', {
          width: clientWidth,
          height: clientHeight,
          className: containerRef.current.className,
          offsetParent: containerRef.current.offsetParent,
          computedStyle: window.getComputedStyle(containerRef.current)
        })

        // Retry after a delay
        const timeoutId = setTimeout(attemptInitialization, 100)
        return () => clearTimeout(timeoutId)
      }

      logger.log('[ChartContainer] Container has valid dimensions, initializing chart', {
        width: clientWidth,
        height: clientHeight
      })

      const height = clientHeight || 600
      setContainerHeight(height)

      const chart = createChart(containerRef.current, {
        width: clientWidth,
        height,
        layout: {
          background: { type: ColorType.Solid, color: theme.backgroundColor },
        textColor: theme.textColor,
        panes: {
          separatorColor: theme.gridColor,
          separatorHoverColor: theme.textColor,
          enableResize: true  // Allow user to drag pane separators
        }
      },
      grid: {
        vertLines: { color: theme.gridColor },
        horzLines: { color: theme.gridColor },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: {
          color: theme.textColor,
          width: 1,
          style: LineStyle.Dashed,
          labelBackgroundColor: theme.textColor,
        },
        horzLine: {
          color: theme.textColor,
          width: 1,
          style: LineStyle.Dashed,
          labelBackgroundColor: theme.textColor,
        },
      },
      timeScale: {
        borderColor: theme.gridColor,
        timeVisible: true,
        secondsVisible: false,
      },
      rightPriceScale: {
        borderColor: theme.gridColor,
        autoScale: true,  // Enable autoscale for responsive scaling
        mode: PriceScaleMode.Normal,  // Normal mode for stable behavior
        scaleMargins: {
          top: 0.03,    // 3% top margin
          bottom: 0.03, // 3% bottom margin (volume in separate pane)
        },
        visible: true,
        entireTextOnly: true,
      },
      handleScroll: {
        mouseWheel: true,
        pressedMouseMove: true,
        horzTouchDrag: true,
        vertTouchDrag: true,
      },
      handleScale: {
        axisPressedMouseMove: {
          time: false,  // Disable X-axis (time) dragging
          price: true,  // Enable Y-axis (price) dragging for zoom
        },
        mouseWheel: true,
        pinch: true,
      },
    })

    chartRef.current = chart

      // NOTE: DO NOT pre-create panes with chart.addPane()
      // TradingView automatically creates panes when addSeries(..., paneIndex) is called
      // Pre-creating panes causes visual clutter (empty panes) and timing issues
      // Following TradingView best practices: let series creation drive pane creation

      // Setup resize observer
      const resizeObserver = new ResizeObserver((entries) => {
        if (entries.length === 0 || entries[0].target !== containerRef.current) return

        const { width, height } = entries[0].contentRect
        chart.applyOptions({ width, height })
        setContainerWidth(width)
        setContainerHeight(height)
      })

      if (containerRef.current) {
        resizeObserver.observe(containerRef.current)
        const rect = containerRef.current.getBoundingClientRect()
        setContainerWidth(rect.width)
        setContainerHeight(rect.height)
      }

      // Mark chart as ready for series components
      setChartReady(true)
      logger.log('[ChartContainer] ✅ Chart initialized and ready for series')

      // Fit content to show all data
      chart.timeScale().fitContent()

      // Notify parent component
      if (onChartReady) {
        onChartReady(chart)
      }

      // Cleanup
      return () => {
        logger.log('[ChartContainer] 🗑️ Cleaning up chart...')
        resizeObserver.disconnect()
        chart.remove()
        setChartReady(false)
      }
    }

    // Start initialization
    const cleanup = attemptInitialization()
    return cleanup
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])  // Empty deps - only run once on mount

  // Update theme colors dynamically (without recreating chart)
  useEffect(() => {
    if (!chartRef.current) return

    logger.log('[ChartContainer] 🎨 Updating theme colors...')

    chartRef.current.applyOptions({
      layout: {
        background: { type: ColorType.Solid, color: theme.backgroundColor },
        textColor: theme.textColor,
        panes: {
          separatorColor: theme.gridColor,
          separatorHoverColor: theme.textColor,
          enableResize: true
        }
      },
      grid: {
        vertLines: { color: theme.gridColor },
        horzLines: { color: theme.gridColor },
      },
      crosshair: {
        vertLine: {
          color: theme.textColor,
          labelBackgroundColor: theme.textColor,
        },
        horzLine: {
          color: theme.textColor,
          labelBackgroundColor: theme.textColor,
        },
      },
      timeScale: {
        borderColor: theme.gridColor,
      },
      rightPriceScale: {
        borderColor: theme.gridColor,
      },
    })

    logger.log('[ChartContainer] ✅ Theme colors updated')
  }, [theme])  // Only update colors when theme changes

  const firstBarTime = chartData.candlestickData[0]?.time
  const lastBarTime = chartData.candlestickData[chartData.candlestickData.length - 1]?.time

  // Fit content when the underlying data changes (ticker/timeframe refresh).
  // IMPORTANT: Do not refit on indicator toggles; that resets the user's zoom.
  useEffect(() => {
    if (!chartRef.current || !chartData.candlestickData.length) return

    // Use requestAnimationFrame to ensure panes are fully created/destroyed and series have rendered
    requestAnimationFrame(() => {
      if (chartRef.current) {
        // 1. Fit X-axis (time scale)
        chartRef.current.timeScale().fitContent()

        // 2. Force Y-axis re-autoscale for main candlestick pane (pane 0, 'right' price scale)
        try {
          const rightPriceScale = chartRef.current.priceScale('right', 0)
          rightPriceScale.applyOptions({ autoScale: true })
          logger.log('[ChartContainer] ✅ Re-fitted time + price scales after data change')
        } catch (error) {
          logger.error('[ChartContainer] Failed to reset price scales:', error)
        }
      }
    })
  }, [ticker, chartType, timeframe, chartData.candlestickData.length, firstBarTime, lastBarTime])

  // Subscribe to crosshair move events
  useEffect(() => {
    if (!chartReady || !chartRef.current) return

    const chart = chartRef.current

    const toEpochSeconds = (time: any): number | null => {
      if (time === null || time === undefined) return null
      if (typeof time === 'number') return time
      if (typeof time === 'object' && 'year' in time && 'month' in time && 'day' in time) {
        const utcMs = Date.UTC(time.year, time.month - 1, time.day)
        return Math.floor(utcMs / 1000)
      }
      return null
    }

    const timeToIndex = new Map<number, number>()
    historicalData.forEach((item, index) => {
      const epoch = Math.floor(new Date(item.date).getTime() / 1000)
      timeToIndex.set(epoch, index)
    })

    const handleCrosshairMove = (param: any) => {
      if (onCrosshairMove) {
        onCrosshairMove(param, chart)
      }

      if (!param || !param.time || !param.point) {
        setTooltipState((prev) => (prev.visible ? { ...prev, visible: false } : prev))
        return
      }

      const epoch = toEpochSeconds(param.time)
      if (epoch === null) {
        setTooltipState((prev) => (prev.visible ? { ...prev, visible: false } : prev))
        return
      }

      const index = timeToIndex.get(epoch)
      if (index === undefined) {
        setTooltipState((prev) => (prev.visible ? { ...prev, visible: false } : prev))
        return
      }

      const current = historicalData[index]
      const data = buildEnhancedTooltipData(current, historicalData, index)
      setTooltipState({
        visible: true,
        data,
        position: { x: param.point.x, y: param.point.y },
      })
    }

    chart.subscribeCrosshairMove(handleCrosshairMove)

    return () => {
      chart.unsubscribeCrosshairMove(handleCrosshairMove)
    }
  }, [chartReady, historicalData, onCrosshairMove])

  // 🚀 ACTIVATION-BASED Pane Registry - Sequential allocation
  // Calculates pane indices based on WHEN indicators were enabled (not canonical order!)
  // Panes auto-created by TradingView when addSeries(..., paneIndex) is called
  // Pane 0 is always the main candlestick chart
  // Pane 1+ are dynamically assigned based on activation timestamps
  const { getPaneIndex } = usePaneRegistry(indicators, indicatorActivationOrder)

  // 🔄 Preserve zoom/scroll when panes shift (indicators toggled OFF)
  usePaneTransitionManager(chartRef.current, indicators)

  // Memoize ChartProvider value to prevent infinite re-renders
  // Only recreate when actual dependencies change
  // ✅ TradingView Pattern: Use state in context for proper React updates
  const chartContextValue = useMemo(
    () => ({
      chart: chartRef.current,
      candlestickSeries,
      containerRef,  // ✅ Use state (not ref) for context
      chartData,
      theme,
      isReady: chartReady,  // ✅ Ready when chart instance created (series created after)
      containerHeight,
      chartType,  // Provide chart type to series components
      timeframe,  // Provide timeframe for context
      indicatorData,  // API-calculated indicator data
    }),
    [
      chartData,
      theme,
      chartReady,
      candlestickSeries,
      containerRef,
      containerHeight,
      chartType,
      timeframe,
      indicatorData,
    ]  // Include new props
  )

  return (
    <ChartProvider value={chartContextValue}>
      <div ref={containerRef} className="w-full h-full relative">
        {/*
          Component-Based Architecture:
          - Each component mounts/unmounts based on indicator state
          - React handles lifecycle automatically
          - No manual series removal needed
          - Y-axis stability achieved through proper lifecycle management
        */}

        {/* Main Series - Always rendered, type determined by chartType prop */}
        {/* Key prop forces component remount when chart type OR timeframe changes, ensuring series recreation */}
        <MainSeriesComponent
          key={`${chartType}-${timeframe}`}
          onSeriesCreated={handleSeriesCreated}
        />

        <DrawingLayer />
        <DrawingInteraction />
        <DrawingToolbar />
        <DrawingThemeSync />
        <DrawingPersistence ticker={ticker} />

        <ChartTooltip
          data={tooltipState.data}
          position={tooltipState.position}
          visible={tooltipState.visible}
          containerWidth={containerWidth}
          containerHeight={containerHeight}
        />

        {/* Overlay Indicators - Conditionally rendered based on indicator state */}
        {indicators?.showSMA20 && <SMAIndicator period={20} />}
        {indicators?.showSMA50 && <SMAIndicator period={50} />}
        {indicators?.showSMA200 && <SMAIndicator period={200} />}
        {indicators?.showEMA20 && <EMAIndicator period={20} />}
        {indicators?.showGeneralSMA && <SMAIndicator period={maSettings.generalSmaPeriod} />}
        {indicators?.showGeneralEMA && <EMAIndicator period={maSettings.generalEmaPeriod} />}
        {indicators?.showSupport && <SupportResistanceIndicator type="support" />}
        {indicators?.showResistance && <SupportResistanceIndicator type="resistance" />}

        {/* Pane Indicators - ACTIVATION-BASED allocation with dynamic keys */}
        {/* Keys include pane index, chartType, and timeframe to force recreation on chart changes */}
        {indicators?.showVolume && (
          <VolumePane
            key={`volume-${chartType}-${timeframe}-${getPaneIndex('volume')}`}
            paneIndex={getPaneIndex('volume')}
          />
        )}
        {indicators?.showOBV && (
          <OBVPane
            key={`obv-pane-${getPaneIndex('obv')}`}
            paneIndex={getPaneIndex('obv')}
          />
        )}
        {indicators?.showMACD && (
          <MACDPane
            key={`macd-pane-${getPaneIndex('macd')}`}
            paneIndex={getPaneIndex('macd')}
          />
        )}
        {indicators?.showRSI && (
          <RSIPane
            key={`rsi-pane-${getPaneIndex('rsi')}`}
            paneIndex={getPaneIndex('rsi')}
          />
        )}
        {indicators?.showMFI && (
          <MFIPane
            key={`mfi-pane-${getPaneIndex('mfi')}`}
            paneIndex={getPaneIndex('mfi')}
          />
        )}
        {indicators?.showStochastic && (
          <StochasticPane
            key={`stochastic-pane-${getPaneIndex('stochastic')}`}
            paneIndex={getPaneIndex('stochastic')}
          />
        )}
        {indicators?.showGeneralStochastic && (
          <GeneralStochasticPane
            key={`general-stochastic-pane-${getPaneIndex('generalStochastic')}`}
            paneIndex={getPaneIndex('generalStochastic')}
          />
        )}
        {indicators?.showADX && (
          <ADXPane
            key={`adx-pane-${getPaneIndex('adx')}`}
            paneIndex={getPaneIndex('adx')}
          />
        )}
        {indicators?.showATR && (
          <ATRPane
            key={`atr-pane-${getPaneIndex('atr')}`}
            paneIndex={getPaneIndex('atr')}
          />
        )}
        {indicators?.showCCI && (
          <CCIPane
            key={`cci-pane-${getPaneIndex('cci')}`}
            paneIndex={getPaneIndex('cci')}
          />
        )}
        {indicators?.showWilliamsR && (
          <WilliamsRPane
            key={`williamsr-pane-${getPaneIndex('williamsR')}`}
            paneIndex={getPaneIndex('williamsR')}
          />
        )}
        {indicators?.showMOM && (
          <MomentumPane
            key={`mom-pane-${getPaneIndex('momentum')}`}
            paneIndex={getPaneIndex('momentum')}
          />
        )}
        {indicators?.showROC && (
          <ROCPane
            key={`roc-pane-${getPaneIndex('roc')}`}
            paneIndex={getPaneIndex('roc')}
          />
        )}

        {/* Overlay Indicators - Render on main chart (pane 0) */}
        {indicators?.showVWAP && <VWAPOverlay key={`vwap-v${vwapSettings.version || 0}`} />}
        {indicators?.showFibonacci && <FibonacciRetracement />}
        {indicators?.showDonchian && <DonchianChannels />}
        {indicators?.showBollingerBands && <BollingerBandsOverlay />}
        {indicators?.showGeneralBollinger && <GeneralBollingerBandsOverlay />}
        {indicators?.showKeltner && <KeltnerChannels />}
        {indicators?.showIchimoku && <IchimokuOverlay />}
        {indicators?.showParabolicSAR && <ParabolicSAROverlay />}
      </div>
    </ChartProvider>
  )
}

