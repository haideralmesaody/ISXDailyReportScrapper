/**
 * Analysis Client Component
 * Following CLAUDE.md Next.js 14 client component patterns
 * Handles all interactivity and state management for technical analysis
 */

'use client'

import React, { useCallback, useRef, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { ChartCandlestick, Activity, RefreshCw, Info, Loader2, CandlestickChart, TrendingUp, Mountain, BarChart3, Minimize2, Maximize2, ChevronLeft, ChevronRight, PanelLeftOpen, PanelLeftClose } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { useToast } from '@/lib/hooks/use-toast'
import { useHydration } from '@/lib/hooks/use-hydration'
import { useChartState } from '@/lib/hooks/use-chart-state'
import { logger } from '@/lib/utils/logger'
import { TickerList } from '@/components/analysis/TickerList'
import { ChartCore } from '@/components/analysis/chart/ChartCore'
import { ChartErrorBoundary } from '@/components/analysis/ChartErrorBoundary'
import { ChartExportButton } from '@/components/analysis/ChartExportButton'
import { ResizablePanel } from '@/components/ui/resizable-panel'
import { NoDataState, DataLoadingState, Alert, AlertDescription } from '@/components/ui'
import { fetchTickerSummary, fetchTickerHistory } from '@/lib/api/analysis'
import { HelpButton } from '@/components/guide/HelpButton'
import { IndicatorToolbar } from '@/components/analysis/IndicatorToolbar'
import {
  trackNoDataResolved,
  trackRetryAttempt,
  debug
} from '@/lib/observability/no-data-metrics'
import type { IChartApi } from 'lightweight-charts'

export default function AnalysisClient() {
  // Hydration guard - CLAUDE.md requirement
  const isHydrated = useHydration()

  // URL query parameters - for ticker navigation from treemap
  const searchParams = useSearchParams()
  const urlTicker = searchParams.get('ticker')

  // DIAGNOSTIC: Log component mount and URL params
  React.useEffect(() => {
    logger.log('[AnalysisClient] Component mounted', {
      urlTicker,
      isHydrated
    })
  }, [urlTicker, isHydrated])

  // Chart state management using reducer - CLAUDE.md recommendation
  const { state, dispatch } = useChartState()
  const { toast } = useToast()

  // Chart instance state for export functionality
  const [chartInstance, setChartInstance] = useState<IChartApi | null>(null)

  // Ticker panel width state for resizable panel
  const [tickerPanelWidth, setTickerPanelWidth] = useState(220)

  // Indicator toolbar visibility state
  const [showIndicatorToolbar, setShowIndicatorToolbar] = useState(false)

  // Sync ticker panel width with persisted value when using controlled ResizablePanel width
  React.useEffect(() => {
    if (!isHydrated) return

    const storageKey = 'isx-ticker-panel-width'
    const saved = localStorage.getItem(storageKey)
    if (!saved) return

    const parsedWidth = parseInt(saved)
    if (!Number.isFinite(parsedWidth)) return

    // Match ResizablePanel constraints
    const constrainedWidth = Math.max(180, Math.min(400, parsedWidth))
    setTickerPanelWidth(constrainedWidth)
  }, [isHydrated])

  // Debounce timer for ticker selection to prevent rapid API calls
  const debounceTimer = useRef<NodeJS.Timeout | null>(null)

  // Ref to track pending ticker synchronously - prevents stale closure issues
  const pendingTickerRef = useRef<string | null>(null)

  // Helper function to detect no-data scenarios
  const isNoDataError = useCallback((error: string | Error): boolean => {
    const errorMessage = error instanceof Error ? error.message : error
    return errorMessage.includes('404') ||
           errorMessage.includes('not found') ||
           errorMessage.includes('No ticker summary data available') ||
           errorMessage.includes('No combined data available')
  }, [])

  // Load ticker summary data with proper error handling
  const loadTickerSummary = useCallback(async (isRetry = false): Promise<void> => {
    dispatch({ type: 'SET_LOADING', payload: { section: 'tickers', loading: true } })
    dispatch({ type: 'SET_ERROR', payload: null })

    const startTime = Date.now()

    try {
      debug.logApiResponse('/api/analysis/ticker-summary', null, false)

      const summaryData = await fetchTickerSummary()
      dispatch({ type: 'SET_TICKERS', payload: summaryData })

      // Track successful data resolution
      if (isRetry || state.performance.loadStartTime) {
        const resolutionTime = Date.now() - (state.performance.loadStartTime || startTime)
        trackNoDataResolved('analysis', resolutionTime, isRetry ? 'retry' : 'api_success')

        debug.logPerformance('Analysis Data Load', resolutionTime, {
          is_retry: isRetry,
          ticker_count: summaryData.length,
          retry_count: state.performance.retryCount
        })
      }

      // Reset retry count on success
      if (isRetry) {
        dispatch({ type: 'RESET_RETRY_COUNT' })
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to load ticker summary'
      dispatch({ type: 'SET_ERROR', payload: message })

      debug.logApiResponse('/api/analysis/ticker-summary', err, true)

      // Track retry attempt
      if (isRetry) {
        trackRetryAttempt('analysis', state.performance.retryCount + 1, false)
        dispatch({ type: 'INCREMENT_RETRY_COUNT' })
      }

      // Only show toast for real errors, not no-data scenarios
      if (!isNoDataError(message)) {
        toast({
          title: 'Error',
          description: message,
          variant: 'destructive'
        })
      }
    } finally {
      dispatch({ type: 'SET_LOADING', payload: { section: 'tickers', loading: false } })
    }
  }, [dispatch, state.performance.loadStartTime, state.performance.retryCount, isNoDataError, toast])

  // Load historical data for selected ticker
  // Race condition prevention with pendingTicker tracking
  const loadTickerHistory = useCallback(async (ticker: string): Promise<void> => {
    // Skip if already loading this exact ticker (prevent duplicate requests)
    if (state.loading.chart && state.loading.pendingTicker === ticker) {
      logger.log(`[Chart] Already loading ${ticker}, skipping duplicate request`)
      return
    }

    logger.log(`[Chart] === LOAD TICKER HISTORY START ===`, {
      ticker,
      currentData: state.chartData.length,
      isLoading: state.loading.chart,
      pendingTickerState: state.loading.pendingTicker,
      pendingTickerRef: pendingTickerRef.current,
      areTheSame: pendingTickerRef.current === ticker
    })

    // Set pending ticker BEFORE async operation - CLAUDE.md race condition prevention
    pendingTickerRef.current = ticker  // Synchronous update - no closure issues
    dispatch({ type: 'SET_PENDING_TICKER', payload: ticker })
    dispatch({ type: 'SET_LOADING', payload: { section: 'chart', loading: true } })

    const startTime = Date.now()

    try {
      const history = await fetchTickerHistory(ticker)
      const loadTime = Date.now() - startTime

      // Check if this request is still relevant (user might have switched tickers)
      // Use ref instead of state to avoid stale closure issues
      if (pendingTickerRef.current !== ticker) {
        logger.log(`[Chart] Discarding stale data for ${ticker}, user selected ${pendingTickerRef.current}`)
        return
      }

      logger.log(`[Chart] Data loaded successfully`, {
        ticker,
        records: history.length,
        loadTimeMs: loadTime,
        firstDate: history[0]?.date,
        lastDate: history[history.length - 1]?.date
      })

      // Update data - replaces previous ticker's data
      dispatch({ type: 'SET_CHART_DATA', payload: history })
    } catch (err) {
      const message = err instanceof Error ? err.message : `Failed to load data for ${ticker}`
      const loadTime = Date.now() - startTime

      logger.error(`[Chart] Data load failed`, {
        ticker,
        error: message,
        loadTimeMs: loadTime
      })

      // Only show error if this is still the pending ticker
      // Use ref to avoid stale closure issues
      if (pendingTickerRef.current === ticker) {
        toast({
          title: 'Chart Data Error',
          description: message,
          variant: 'destructive'
        })

        // Clear data on error to show empty state
        dispatch({ type: 'SET_CHART_DATA', payload: [] })
      }
    } finally {
      // Clear pending ticker and loading state
      pendingTickerRef.current = null  // Clear ref synchronously
      dispatch({ type: 'SET_PENDING_TICKER', payload: null })
      dispatch({ type: 'SET_LOADING', payload: { section: 'chart', loading: false } })
    }
  }, [dispatch, toast, state.chartData.length, state.loading.chart, state.loading.pendingTicker])

  // Handle ticker selection with debouncing
  const handleTickerSelect = useCallback((ticker: string): void => {
    logger.log(`[Chart] === TICKER SELECTION START ===`, {
      selectedTicker: ticker,
      previousTicker: state.selectedTicker,
      hasExistingData: state.chartData.length > 0,
      isLoading: state.loading.chart,
      isHydrated
    })

    // Clear any pending debounced calls
    if (debounceTimer.current) {
      clearTimeout(debounceTimer.current)
      logger.log(`[Chart] Cleared previous debounce timer`)
    }

    // Immediately update selected ticker in UI
    dispatch({ type: 'SET_SELECTED_TICKER', payload: ticker })

    // Debounce the data loading to prevent rapid API calls
    if (isHydrated) {
      logger.log(`[Chart] Setting debounce timer (300ms) for data load`)
      debounceTimer.current = setTimeout(() => {
        logger.log(`[Chart] Debounce completed, triggering data load`)
        loadTickerHistory(ticker)
      }, 300) // 300ms debounce delay
    } else {
      logger.warn(`[Chart] Not hydrated yet, skipping data load`)
    }
  }, [dispatch, isHydrated, loadTickerHistory, state.selectedTicker, state.chartData.length, state.loading.chart])

  // Handle indicator toggle
  const handleIndicatorToggle = useCallback((indicatorId: string): void => {
    dispatch({ type: 'TOGGLE_INDICATOR', payload: indicatorId })
  }, [dispatch])

  // Helper functions for chart controls bar
  const getCompanyName = useCallback((ticker: string): string => {
    const tickerData = state.tickers.find(t => t.Ticker === ticker)
    return tickerData?.CompanyName || ''
  }, [state.tickers])

  const getLastPrice = useCallback((): number => {
    if (state.chartData.length === 0) return 0
    const lastData = state.chartData[state.chartData.length - 1]
    return lastData?.close || 0
  }, [state.chartData])

  const getChange = useCallback((): number => {
    if (state.chartData.length < 2) return 0
    const lastData = state.chartData[state.chartData.length - 1]
    return lastData?.change || 0
  }, [state.chartData])

  const getChangePercent = useCallback((): number => {
    if (state.chartData.length < 2) return 0
    const lastData = state.chartData[state.chartData.length - 1]
    return lastData?.changePercent || 0
  }, [state.chartData])

  // Fullscreen toggle handler
  const chartContainerRef = useRef<HTMLDivElement>(null)

  const toggleFullscreen = useCallback(async () => {
    const container = chartContainerRef.current
    if (!container) return

    try {
      if (!document.fullscreenElement) {
        await container.requestFullscreen()
      } else {
        await document.exitFullscreen()
      }
    } catch (err) {
      logger.error('[Fullscreen] Toggle failed', { error: err })
    }
  }, [])

  // Listen for fullscreen changes (ESC key, browser controls)
  React.useEffect(() => {
    const handleFullscreenChange = () => {
      // Fullscreen state is tracked in browser, no need to update state
      // UI will re-render based on !!document.fullscreenElement check in ChartControlsBar
    }

    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
  }, [])

  // Load data on mount - CLAUDE.md pattern
  React.useEffect(() => {
    if (isHydrated) {
      loadTickerSummary()
    }
  }, [isHydrated, loadTickerSummary])

  // Auto-select ticker: URL parameter has absolute priority
  // Handles both fresh mount (from treemap) and URL changes
  React.useEffect(() => {
    // DIAGNOSTIC: Log effect execution
    logger.log(`[Chart] Auto-select effect triggered`, {
      isHydrated,
      tickersCount: state.tickers.length,
      urlTicker,
      selectedTicker: state.selectedTicker
    })

    // Wait for hydration AND tickers to be loaded
    if (!isHydrated || state.tickers.length === 0) {
      logger.log(`[Chart] Auto-select effect early return`, {
        reason: !isHydrated ? 'not hydrated' : 'no tickers loaded',
        isHydrated,
        tickersCount: state.tickers.length
      })
      return
    }

    // Don't re-select if we already have the correct ticker selected
    if (state.selectedTicker) {
      // Check if URL ticker changed (user navigated to different ticker)
      if (urlTicker && urlTicker !== state.selectedTicker) {
        logger.log(`[Chart] URL ticker changed from ${state.selectedTicker} to ${urlTicker}`)
        handleTickerSelect(urlTicker)
      }
      return
    }

    // No ticker selected - auto-select from URL or first ticker
    if (urlTicker) {
      const tickerExists = state.tickers.some(t => t.Ticker === urlTicker)
      if (tickerExists) {
        logger.log(`[Chart] Auto-selecting URL ticker: ${urlTicker}`)
        handleTickerSelect(urlTicker)
        return
      }
      logger.warn(`[Chart] URL ticker ${urlTicker} not found in ticker list`)
    }

    // Fallback to first ticker
    const firstTicker = state.tickers[0]?.Ticker
    if (firstTicker) {
      logger.log(`[Chart] Auto-selecting first ticker: ${firstTicker}`)
      handleTickerSelect(firstTicker)
    }
  }, [isHydrated, state.tickers, state.selectedTicker, urlTicker, handleTickerSelect])

  // Cleanup debounce timer on unmount
  React.useEffect(() => {
    return () => {
      if (debounceTimer.current) {
        clearTimeout(debounceTimer.current)
      }
    }
  }, [])

  // Show loading state during hydration - CLAUDE.md requirement
  if (!isHydrated) {
    return (
      <DataLoadingState
        message="Initializing analysis tools..."
        className="h-[calc(100vh-128px)]"
        showCard={false}
        size="default"
        page="analysis"
        operation="hydration"
        trackPerformance={true}
      />
    )
  }

  // Show no-data state when appropriate
  if (!state.loading.tickers && state.error && (isNoDataError(state.error) || state.tickers.length === 0)) {
    return (
      <NoDataState
        icon={ChartCandlestick}
        iconColor="blue"
        title="No Analysis Data Available"
        description="You need to run the data collection operations first to generate analysis data."
        className="h-[calc(100vh-128px)] p-8"
        page="analysis"
        reason={state.error || "no_data_available"}
        componentName="AnalysisNoDataState"
        instructions={[
          "Go to the Operations page",
          "Run 'Full Pipeline' to collect all data",
          "Wait for the analysis to complete",
          "Return here to view the technical analysis"
        ]}
        actions={[
          {
            label: "Go to Operations",
            variant: "default",
            href: "/operations",
            icon: Activity
          },
          {
            label: "Check Again",
            variant: "outline",
            onClick: () => loadTickerSummary(true),
            icon: RefreshCw
          }
        ]}
      />
    )
  }

  return (
    <div className="h-[calc(100vh-100px)] flex relative">

      {/* Left Panel - Ticker List - Resizable with smart column collapsing */}
      <ResizablePanel
        defaultWidth={220}
        minWidth={state.ui.tickerPanelCollapsed ? 48 : 180}
        maxWidth={400}
        width={state.ui.tickerPanelCollapsed ? 48 : tickerPanelWidth}
        disableResize={state.ui.tickerPanelCollapsed}
        onWidthChange={setTickerPanelWidth}
        storageKey="isx-ticker-panel-width"
      >
        {state.ui.tickerPanelCollapsed ? (
          <Button
            variant="ghost"
            className="h-full w-full rounded-none"
            onClick={() => dispatch({ type: 'TOGGLE_TICKER_PANEL' })}
            title="Expand ticker panel"
          >
            <PanelLeftOpen className="h-4 w-4" />
          </Button>
        ) : (
          <>
            <div className="p-4 border-b flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-lg font-semibold">Market Overview</h2>
                  <HelpButton
                    section="charts"
                    tooltip="Learn about chart features"
                  />
                </div>
                <p className="text-sm text-muted-foreground">
                  {state.tickers.length} active tickers
                </p>
              </div>
              <button
                onClick={() => dispatch({ type: 'TOGGLE_TICKER_PANEL' })}
                className="p-1 hover:bg-muted rounded transition-colors"
                title="Collapse ticker panel"
              >
                <PanelLeftClose className="h-4 w-4" />
              </button>
            </div>

            <div className="flex-1 overflow-hidden">
              {state.loading.tickers ? (
                <DataLoadingState
                  message="Loading tickers..."
                  showCard={false}
                  size="sm"
                  className="h-full"
                  page="analysis"
                  operation="ticker_loading"
                  trackPerformance={true}
                />
              ) : state.error && !isNoDataError(state.error) ? (
                <div className="p-4">
                  <Alert variant="destructive">
                    <Info className="h-4 w-4" />
                    <AlertDescription>
                      <div className="space-y-2">
                        <p className="text-sm">{state.error}</p>
                        <button
                          onClick={() => loadTickerSummary(true)}
                          className="text-sm text-primary hover:underline underline-offset-4 transition-colors"
                        >
                          Try again
                        </button>
                      </div>
                    </AlertDescription>
                  </Alert>
                </div>
              ) : (
                <TickerList
                  key="ticker-list"
                  tickers={state.tickers}
                  selectedTicker={state.selectedTicker}
                  onTickerSelect={handleTickerSelect}
                  panelWidth={tickerPanelWidth}
                />
              )}
            </div>
          </>
        )}
      </ResizablePanel>

      {/* Right Panel - Chart (Matches Guide Layout) */}
      <div
        ref={chartContainerRef}
        className="flex-1 flex flex-col bg-background relative overflow-hidden min-w-0"
      >
        {state.selectedTicker ? (
          <div className="flex-1 flex flex-col min-h-0 relative p-4 space-y-4">
            {/* 1. Stock Info Header - Matches Guide Layout */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Badge variant="outline" className="text-sm font-semibold">
                  {state.selectedTicker}
                </Badge>
                <div>
                  <p className="text-sm font-medium">{getCompanyName(state.selectedTicker)}</p>
                  <p className="text-xs text-muted-foreground">Iraqi Stock Exchange</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-lg font-semibold">{getLastPrice().toFixed(3)} IQD</p>
                <p className={`text-xs ${getChange() >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                  {getChange() >= 0 ? '+' : ''}{getChange().toFixed(3)} ({getChangePercent().toFixed(2)}%)
                </p>
              </div>
            </div>

            {/* 2. Controls Bar - Matches Guide Layout */}
            <div className="flex flex-wrap gap-2 items-center">
              {/* Left-side controls (match chart-type button style) */}
              <div className="flex gap-1">
                <Button
                  size="sm"
                  variant={state.ui.tickerPanelCollapsed ? 'outline' : 'default'}
                  onClick={() => dispatch({ type: 'TOGGLE_TICKER_PANEL' })}
                  disabled={state.loading.tickers}
                >
                  {state.ui.tickerPanelCollapsed ? (
                    <PanelLeftOpen className="h-4 w-4 mr-1" />
                  ) : (
                    <PanelLeftClose className="h-4 w-4 mr-1" />
                  )}
                  Tickers
                </Button>

                <Button
                  size="sm"
                  variant={showIndicatorToolbar ? 'default' : 'outline'}
                  onClick={() => setShowIndicatorToolbar(!showIndicatorToolbar)}
                  disabled={state.loading.chart || state.chartData.length === 0}
                >
                  <Activity className="h-4 w-4 mr-1" />
                  Indicators
                </Button>
              </div>

              {/* Chart Type Buttons */}
              <div className="flex gap-1">
                <Button
                  size="sm"
                  variant={state.chartDisplay.chartType === 'candlestick' ? 'default' : 'outline'}
                  onClick={() => dispatch({ type: 'SET_CHART_TYPE', payload: 'candlestick' })}
                  disabled={state.loading.chart || state.chartData.length === 0}
                >
                  <CandlestickChart className="h-4 w-4 mr-1" />
                  Candlestick
                </Button>
                <Button
                  size="sm"
                  variant={state.chartDisplay.chartType === 'line' ? 'default' : 'outline'}
                  onClick={() => dispatch({ type: 'SET_CHART_TYPE', payload: 'line' })}
                  disabled={state.loading.chart || state.chartData.length === 0}
                >
                  <TrendingUp className="h-4 w-4 mr-1" />
                  Line
                </Button>
                <Button
                  size="sm"
                  variant={state.chartDisplay.chartType === 'area' ? 'default' : 'outline'}
                  onClick={() => dispatch({ type: 'SET_CHART_TYPE', payload: 'area' })}
                  disabled={state.loading.chart || state.chartData.length === 0}
                >
                  <Mountain className="h-4 w-4 mr-1" />
                  Area
                </Button>
                <Button
                  size="sm"
                  variant={state.chartDisplay.chartType === 'bar' ? 'default' : 'outline'}
                  onClick={() => dispatch({ type: 'SET_CHART_TYPE', payload: 'bar' })}
                  disabled={state.loading.chart || state.chartData.length === 0}
                >
                  <BarChart3 className="h-4 w-4 mr-1" />
                  Bar
                </Button>
              </div>

              {/* Timeframe Buttons */}
              <div className="flex gap-1">
                {(['1D', '1W', '1M', '3M', '1Y', 'MAX'] as const).map((tf) => (
                  <Button
                    key={tf}
                    size="sm"
                    variant={state.chartDisplay.timeframe === tf ? 'default' : 'outline'}
                    onClick={() => dispatch({ type: 'SET_TIMEFRAME', payload: tf })}
                    disabled={state.loading.chart || state.chartData.length === 0}
                  >
                    {tf}
                  </Button>
                ))}
              </div>

              {/* Export + Indicators + Fullscreen */}
              <div className="flex gap-1 ml-auto">
                <ChartExportButton chart={chartInstance} ticker={state.selectedTicker} />
                <Button
                  size="sm"
                  variant="outline"
                  onClick={toggleFullscreen}
                  disabled={state.loading.chart || state.chartData.length === 0}
                >
                  {!!document.fullscreenElement ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
                </Button>
              </div>
            </div>

            {/* Loading overlay during ticker transition - CLAUDE.md UX enhancement */}
            {state.loading.chart && state.loading.pendingTicker && state.loading.pendingTicker !== state.selectedTicker && (
              <div className="absolute inset-0 bg-background/60 backdrop-blur-sm z-10 flex items-center justify-center">
                <div className="bg-card border rounded-lg p-4 shadow-lg">
                  <div className="flex items-center gap-3">
                    <Loader2 className="h-5 w-5 animate-spin text-primary" />
                    <div>
                      <p className="text-sm font-medium">Loading {state.loading.pendingTicker}</p>
                      <p className="text-xs text-muted-foreground">Chart updating...</p>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* 3. Indicators Sidebar + Chart */}
            <div className="flex-1 min-h-0 flex gap-3">
              {/* Indicators panel (left, collapsible) */}
              <div
                className={`shrink-0 transition-[width] duration-200 ease-out ${showIndicatorToolbar ? 'w-80' : 'w-10'}`}
              >
                {showIndicatorToolbar ? (
                  <div className="h-full flex flex-col min-h-0 gap-2">
                    <div className="flex items-center justify-end">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setShowIndicatorToolbar(false)}
                        aria-label="Collapse indicators panel"
                      >
                        <ChevronLeft className="h-4 w-4" />
                      </Button>
                    </div>
                    <div className="flex-1 min-h-0 overflow-hidden">
                      <IndicatorToolbar className="h-full" state={state} dispatch={dispatch} />
                    </div>
                  </div>
                ) : (
                  <Button
                    variant="outline"
                    className="h-full w-full p-0 flex items-center justify-center"
                    onClick={() => setShowIndicatorToolbar(true)}
                    aria-label="Expand indicators panel"
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                )}
              </div>

              {/* Chart Container - Wrapped in Card Like Guide */}
              <Card className="flex-1 min-h-0">
                <CardContent className="p-4 h-full">
                  <ChartErrorBoundary fallbackMessage="Failed to load chart. Check your data or try selecting another ticker.">
                    <ChartCore
                      ticker={state.selectedTicker}
                      data={state.chartData}
                      isLoading={state.loading.chart}
                      onChartReady={setChartInstance}
                      chartType={state.chartDisplay.chartType}
                      timeframe={state.chartDisplay.timeframe}
                      indicators={state.indicators}
                      indicatorActivationOrder={state.ui.indicatorActivationOrder}
                    />
                  </ChartErrorBoundary>
                </CardContent>
              </Card>
            </div>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
            <div className="text-center">
              <p className="text-lg mb-2">Select a ticker to view chart</p>
              <p className="text-sm">Choose from the list on the left to begin technical analysis</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
