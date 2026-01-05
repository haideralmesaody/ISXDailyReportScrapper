/**
 * Chart State Management Hook (Simplified)
 * Following CLAUDE.md useReducer pattern for complex state
 * Manages all chart-related state with type safety
 */

'use client'

import { useReducer, useCallback, useEffect } from 'react'
import type { TickerSummary, TickerHistoricalData } from '@/types/analysis'
import { logger } from '@/lib/utils/logger'

// Type definitions - CLAUDE.md requirement for strict typing
export interface PanelLayout {
  id: string
  height: number
  visible: boolean
  label: string
  minHeight: number
  maxHeight: number
}

export interface TooltipState {
  visible: boolean
  data: any | null
  position: { x: number; y: number }
}

export interface LoadingState {
  tickers: boolean
  chart: boolean
  pendingTicker: string | null  // Track which ticker is being loaded to prevent race conditions
}

export interface UIState {
  isFullscreen: boolean
  tickerPanelCollapsed: boolean
  hasAutoSelected: boolean
  indicatorActivationOrder: Record<string, number>  // Track when indicators were activated
}

export type ChartType = 'candlestick' | 'line' | 'area' | 'bar'
export type Timeframe = '1D' | '1W' | '1M' | '3M' | '1Y' | 'MAX'

export interface ChartDisplaySettings {
  chartType: ChartType
  timeframe: Timeframe
}

// Indicator settings for chart display
export interface IndicatorSettings {
  // Moving Averages
  showSMA20: boolean
  showSMA50: boolean
  showSMA200: boolean
  showEMA20: boolean
  showGeneralSMA: boolean
  showGeneralEMA: boolean

  // Support & Resistance
  showSupport: boolean
  showResistance: boolean

  // Volume Indicators
  showVolume: boolean
  showOBV: boolean

  // Momentum Indicators (in separate panes)
  showMACD: boolean
  showRSI: boolean
  showMFI: boolean
  showStochastic: boolean
  showGeneralStochastic: boolean
  showADX: boolean
  showATR: boolean
  showCCI: boolean
  showWilliamsR: boolean
  showMOM: boolean
  showROC: boolean

  // Overlay Indicators (on main chart)
  showVWAP: boolean
  showFibonacci: boolean
  showDonchian: boolean
  showBollingerBands: boolean
  showGeneralBollinger: boolean
  showKeltner: boolean
  showIchimoku: boolean
  showParabolicSAR: boolean
}

export interface PerformanceState {
  loadStartTime: number | null
  retryCount: number
}

export interface ChartState {
  // Data state
  tickers: TickerSummary[]
  selectedTicker: string | null
  chartData: TickerHistoricalData[]

  // UI state
  panels: PanelLayout[]
  tooltip: TooltipState
  loading: LoadingState
  ui: UIState
  chartDisplay: ChartDisplaySettings  // NEW: Chart type and timeframe settings
  indicators: IndicatorSettings  // Track indicator states

  // Error & performance
  error: string | null
  performance: PerformanceState
}

// Action types - CLAUDE.md pattern for type safety
export type ChartAction =
  | { type: 'SET_TICKERS'; payload: TickerSummary[] }
  | { type: 'SET_SELECTED_TICKER'; payload: string }
  | { type: 'SET_CHART_DATA'; payload: TickerHistoricalData[] }
  | { type: 'SET_TOOLTIP'; payload: TooltipState }
  | { type: 'CLEAR_TOOLTIP' }
  | { type: 'SET_LOADING'; payload: Partial<LoadingState> }
  | { type: 'TOGGLE_TICKER_PANEL' }
  | { type: 'TOGGLE_FULLSCREEN' }
  | { type: 'SET_AUTO_SELECTED'; payload: boolean }
  | { type: 'SET_CHART_TYPE'; payload: ChartType }
  | { type: 'SET_TIMEFRAME'; payload: Timeframe }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'CLEAR_ERROR' }
  | { type: 'RETRY_LOAD' }
  | { type: 'SET_LOAD_START_TIME' }
  | { type: 'TOGGLE_INDICATOR'; payload: keyof IndicatorSettings }
  | { type: 'SET_PENDING_TICKER'; payload: string | null }
  | { type: 'INCREMENT_RETRY_COUNT' }
  | { type: 'RESET_RETRY_COUNT' }

// Initial state - CLAUDE.md pattern for predictable initialization
const initialState: ChartState = {
  // Data state
  tickers: [],
  selectedTicker: null,
  chartData: [],

  // UI state
  panels: [],
  tooltip: { visible: false, data: null, position: { x: 0, y: 0 } },
  loading: { tickers: false, chart: false, pendingTicker: null },
  ui: {
    isFullscreen: false,
    tickerPanelCollapsed: false,
    hasAutoSelected: false,
    // Ensure default-enabled indicators get a valid pane allocation.
    indicatorActivationOrder: { showVolume: 1 },
  },
  chartDisplay: { chartType: 'candlestick', timeframe: '3M' },  // Default to candlestick and 3 months
  indicators: {  // Default indicator settings - all disabled except volume
    // Moving Averages
    showSMA20: false,
    showSMA50: false,
    showSMA200: false,
    showEMA20: false,
    showGeneralSMA: false,
    showGeneralEMA: false,

    // Support & Resistance
    showSupport: false,
    showResistance: false,

    // Volume Indicators
    showVolume: true,  // Volume is shown by default
    showOBV: false,

    // Momentum Indicators (in separate panes)
    showMACD: false,
    showRSI: false,
    showMFI: false,
    showStochastic: false,
    showGeneralStochastic: false,
    showADX: false,
    showATR: false,
    showCCI: false,
    showWilliamsR: false,
    showMOM: false,
    showROC: false,

    // Overlay Indicators (on main chart)
    showVWAP: false,
    showFibonacci: false,
    showDonchian: false,
    showBollingerBands: false,
    showGeneralBollinger: false,
    showKeltner: false,
    showIchimoku: false,
    showParabolicSAR: false,
  },

  // Error & performance
  error: null,
  performance: {
    loadStartTime: null,
    retryCount: 0
  }
}

// Reducer function - CLAUDE.md pattern for pure state updates
export function chartReducer(state: ChartState, action: ChartAction): ChartState {
  switch (action.type) {
    case 'SET_TICKERS':
      return {
        ...state,
        tickers: action.payload,
        error: null
      }

    case 'SET_SELECTED_TICKER':
      // Don't clear chartData - keep previous chart visible while loading new data
      // This provides better UX (no blank screen flash)
      return {
        ...state,
        selectedTicker: action.payload,
        error: null
      }

    case 'SET_CHART_DATA':
      return {
        ...state,
        chartData: action.payload,
        loading: { ...state.loading, chart: false, pendingTicker: null },
        error: null
      }

    case 'SET_TOOLTIP':
      return {
        ...state,
        tooltip: action.payload
      }

    case 'CLEAR_TOOLTIP':
      return {
        ...state,
        tooltip: { visible: false, data: null, position: { x: 0, y: 0 } }
      }

    case 'SET_LOADING':
      return {
        ...state,
        loading: { ...state.loading, ...action.payload }
      }

    case 'TOGGLE_TICKER_PANEL':
      return {
        ...state,
        ui: {
          ...state.ui,
          tickerPanelCollapsed: !state.ui.tickerPanelCollapsed
        }
      }

    case 'TOGGLE_FULLSCREEN':
      return {
        ...state,
        ui: {
          ...state.ui,
          isFullscreen: !state.ui.isFullscreen
        }
      }

    case 'SET_AUTO_SELECTED':
      return {
        ...state,
        ui: {
          ...state.ui,
          hasAutoSelected: action.payload
        }
      }

    case 'SET_CHART_TYPE':
      return {
        ...state,
        chartDisplay: {
          ...state.chartDisplay,
          chartType: action.payload
        }
      }

    case 'SET_TIMEFRAME':
      return {
        ...state,
        chartDisplay: {
          ...state.chartDisplay,
          timeframe: action.payload
        }
      }

    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,
        loading: { ...state.loading, chart: false, pendingTicker: null }
      }

    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null
      }

    case 'RETRY_LOAD':
      return {
        ...state,
        error: null,
        performance: {
          ...state.performance,
          retryCount: state.performance.retryCount + 1
        }
      }

    case 'SET_LOAD_START_TIME':
      return {
        ...state,
        performance: {
          ...state.performance,
          loadStartTime: Date.now()
        }
      }

    case 'TOGGLE_INDICATOR':
      const indicatorKey = action.payload
      const currentIndicators = state.indicators || {}
      return {
        ...state,
        indicators: {
          ...currentIndicators,
          [indicatorKey]: !currentIndicators[indicatorKey]
        },
        ui: {
          ...state.ui,
          indicatorActivationOrder: {
            ...state.ui.indicatorActivationOrder,
            [indicatorKey]: Date.now()
          }
        }
      }

    case 'SET_PENDING_TICKER':
      return {
        ...state,
        loading: {
          ...state.loading,
          pendingTicker: action.payload
        }
      }

    case 'INCREMENT_RETRY_COUNT':
      return {
        ...state,
        performance: {
          ...state.performance,
          retryCount: state.performance.retryCount + 1
        }
      }

    case 'RESET_RETRY_COUNT':
      return {
        ...state,
        performance: {
          ...state.performance,
          retryCount: 0
        }
      }

    default:
      logger.warn('Unknown action type in chartReducer', { actionType: action.type })
      return state
  }
}

// Custom hook - CLAUDE.md pattern for complex state with useReducer
export function useChartState() {
  const STORAGE_KEY = 'isx-analysis-chart-preferences'
  const STORAGE_VERSION = 1

  const initializer = (base: ChartState): ChartState => {
    if (typeof window === 'undefined') return base
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      if (!raw) return base
      const parsed = JSON.parse(raw)
      if (!parsed || parsed.version !== STORAGE_VERSION) return base

      return {
        ...base,
        chartDisplay: {
          ...base.chartDisplay,
          ...(parsed.chartDisplay || {}),
        },
        indicators: {
          ...base.indicators,
          ...(parsed.indicators || {}),
        },
        ui: {
          ...base.ui,
          indicatorActivationOrder: {
            ...base.ui.indicatorActivationOrder,
            ...(parsed.indicatorActivationOrder || {}),
          },
        },
      }
    } catch {
      return base
    }
  }

  const [state, dispatch] = useReducer(chartReducer, initialState, initializer)

  // Persist chart display + indicators so analysis defaults survive refreshes.
  useEffect(() => {
    if (typeof window === 'undefined') return
    try {
      localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          version: STORAGE_VERSION,
          chartDisplay: state.chartDisplay,
          indicators: state.indicators,
          indicatorActivationOrder: state.ui.indicatorActivationOrder,
        })
      )
    } catch {
      // ignore storage errors (private mode, quota, etc.)
    }
  }, [state.chartDisplay, state.indicators, state.ui.indicatorActivationOrder])

  // Action creators with logging - CLAUDE.md pattern for debugging
  const actions = {
    setTickers: useCallback((tickers: TickerSummary[]) => {
      logger.debug('Setting tickers', { count: tickers.length })
      dispatch({ type: 'SET_TICKERS', payload: tickers })
    }, [dispatch]),

    setSelectedTicker: useCallback((ticker: string) => {
      logger.debug('Setting selected ticker', { ticker })
      dispatch({ type: 'SET_SELECTED_TICKER', payload: ticker })
    }, [dispatch]),

    setChartDataType: useCallback((data: TickerHistoricalData[]) => {
      logger.debug('Setting chart data', { points: data.length })
      dispatch({ type: 'SET_CHART_DATA', payload: data })
    }, [dispatch]),

    setTooltip: useCallback((tooltip: TooltipState) => {
      dispatch({ type: 'SET_TOOLTIP', payload: tooltip })
    }, [dispatch]),

    clearTooltip: useCallback(() => {
      dispatch({ type: 'CLEAR_TOOLTIP' })
    }, [dispatch]),

    setLoading: useCallback((loading: Partial<LoadingState>) => {
      dispatch({ type: 'SET_LOADING', payload: loading })
    }, [dispatch]),

    toggleTickerPanel: useCallback(() => {
      dispatch({ type: 'TOGGLE_TICKER_PANEL' })
    }, [dispatch]),

    toggleFullscreen: useCallback(() => {
      dispatch({ type: 'TOGGLE_FULLSCREEN' })
    }, [dispatch]),

    setAutoSelected: useCallback((selected: boolean) => {
      dispatch({ type: 'SET_AUTO_SELECTED', payload: selected })
    }, [dispatch]),

    setChartType: useCallback((chartType: ChartType) => {
      logger.debug('Setting chart type', { chartType })
      dispatch({ type: 'SET_CHART_TYPE', payload: chartType })
    }, [dispatch]),

    setTimeframe: useCallback((timeframe: Timeframe) => {
      logger.debug('Setting timeframe', { timeframe })
      dispatch({ type: 'SET_TIMEFRAME', payload: timeframe })
    }, [dispatch]),

    setError: useCallback((error: string | null) => {
      logger.error('Setting error state', { error })
      dispatch({ type: 'SET_ERROR', payload: error })
    }, [dispatch]),

    clearError: useCallback(() => {
      dispatch({ type: 'CLEAR_ERROR' })
    }, [dispatch]),

    retryLoad: useCallback(() => {
      logger.info('Retrying data load', {
        retryCount: state.performance.retryCount + 1,
        selectedTicker: state.selectedTicker
      })
      dispatch({ type: 'RETRY_LOAD' })
    }, [dispatch, state.performance.retryCount, state.selectedTicker]),

    setLoadStartTime: useCallback(() => {
      dispatch({ type: 'SET_LOAD_START_TIME' })
    }, [dispatch]),

    toggleIndicator: useCallback((indicator: keyof IndicatorSettings) => {
      dispatch({ type: 'TOGGLE_INDICATOR', payload: indicator })
    }, [dispatch]),

    setPendingTicker: useCallback((ticker: string | null) => {
      dispatch({ type: 'SET_PENDING_TICKER', payload: ticker })
    }, [dispatch]),

    incrementRetryCount: useCallback(() => {
      dispatch({ type: 'INCREMENT_RETRY_COUNT' })
    }, [dispatch]),

    resetRetryCount: useCallback(() => {
      dispatch({ type: 'RESET_RETRY_COUNT' })
    }, [dispatch])
  }

  return {
    state,
    dispatch,
    ...actions
  }
}
