/**
 * Type definitions for chart indicators
 */

export enum IndicatorType {
  OVERLAY = 'overlay',           // Renders on main chart (SMA, EMA, Bollinger)
  VOLUME = 'volume',            // Volume-based indicators
  OSCILLATOR = 'oscillator',   // 0-100 range indicators (RSI, Stochastic)
  MOMENTUM = 'momentum',        // Dynamic range (MACD)
  BANDS = 'bands',             // Bollinger Bands, Keltner Channels
  SUPPORT_RESISTANCE = 'levels' // Horizontal price levels
}

export enum IndicatorCategory {
  MOVING_AVERAGE = 'moving-average',
  MOMENTUM = 'momentum',
  VOLATILITY = 'volatility',
  VOLUME = 'volume',
  SUPPORT_RESISTANCE = 'support-resistance',
  TREND = 'trend'
}

export interface IndicatorDefinition {
  id: string
  name: string
  displayName: string
  type: IndicatorType
  category: IndicatorCategory
  description: string
  defaultPeriod?: number
  periods?: number[]           // For multi-period indicators
  paneRequirement: 'none' | 'shared' | 'dedicated'
  defaultHeight?: number       // Suggested pane height (0-1)
  rangeType: 'fixed' | 'dynamic' | 'price'
  fixedRange?: [number, number] // e.g., [0, 100] for RSI
  color?: string
  lineStyle?: 'solid' | 'dashed' | 'dotted'
  lineWidth?: number
}

export interface IndicatorValue {
  timestamp: number
  value: number | null
  formatted?: string
}

export interface IndicatorData {
  id: string
  values: IndicatorValue[]
  metadata?: any
}

export interface IndicatorSettings {
  showSMA20: boolean
  showSMA50: boolean
  showSMA200: boolean
  showEMA20: boolean
  showResistance: boolean
  showSupport: boolean
  showMACD: boolean
  showRSI: boolean
  // Future indicators
  showBollinger?: boolean
  showStochastic?: boolean
  showVWAP?: boolean
  showATR?: boolean
  showCCI?: boolean
  showWilliamsR?: boolean
  showOBV?: boolean
  showMFI?: boolean
}

// Indicator calculation parameters
export interface MACDParams {
  fastPeriod: number  // Default 12
  slowPeriod: number  // Default 26
  signalPeriod: number // Default 9
}

export interface RSIParams {
  period: number      // Default 14
  overbought: number  // Default 70
  oversold: number    // Default 30
}

export interface BollingerParams {
  period: number      // Default 20
  stdDev: number      // Default 2
}

export interface StochasticParams {
  kPeriod: number     // Default 14
  dPeriod: number     // Default 3
  smooth: number      // Default 3
  overbought: number  // Default 80
  oversold: number    // Default 20
}

export interface SMAParams {
  period: number
}

export interface EMAParams {
  period: number
}

// Calculation results
export interface MACDResult {
  macd: (number | null)[]
  signal: (number | null)[]
  histogram: (number | null)[]
}

export interface RSIResult {
  values: (number | null)[]
  overbought: number
  oversold: number
}

export interface BollingerResult {
  upper: (number | null)[]
  middle: (number | null)[]
  lower: (number | null)[]
}

export interface StochasticResult {
  k: (number | null)[]
  d: (number | null)[]
  overbought: number
  oversold: number
}

// Chart data interfaces
export interface OHLCData {
  time: number
  open: number
  high: number
  low: number
  close: number
  volume?: number
}

export interface LineData {
  time: number
  value: number
}

export interface HistogramData {
  time: number
  value: number
  color?: string
}

// Indicator calculation function type
export type IndicatorCalculator<P = any, R = any> = (
  data: OHLCData[],
  params: P
) => R

// Registry types
export interface IndicatorRegistryEntry {
  definition: IndicatorDefinition
  calculator?: IndicatorCalculator
  seriesRef?: any // Reference to chart series
  visible: boolean
  data?: IndicatorData
}

// Export all indicator definitions
export const INDICATOR_DEFINITIONS: IndicatorDefinition[] = [
  // Moving Averages
  {
    id: 'sma20',
    name: 'SMA 20',
    displayName: 'SMA (20)',
    type: IndicatorType.OVERLAY,
    category: IndicatorCategory.MOVING_AVERAGE,
    description: 'Simple Moving Average - 20 periods',
    defaultPeriod: 20,
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#3B82F6',
    lineWidth: 2
  },
  {
    id: 'sma50',
    name: 'SMA 50',
    displayName: 'SMA (50)',
    type: IndicatorType.OVERLAY,
    category: IndicatorCategory.MOVING_AVERAGE,
    description: 'Simple Moving Average - 50 periods',
    defaultPeriod: 50,
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#F97316',
    lineWidth: 2
  },
  {
    id: 'sma200',
    name: 'SMA 200',
    displayName: 'SMA (200)',
    type: IndicatorType.OVERLAY,
    category: IndicatorCategory.MOVING_AVERAGE,
    description: 'Simple Moving Average - 200 periods',
    defaultPeriod: 200,
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#8B5CF6',
    lineWidth: 2
  },
  {
    id: 'ema20',
    name: 'EMA 20',
    displayName: 'EMA (20)',
    type: IndicatorType.OVERLAY,
    category: IndicatorCategory.MOVING_AVERAGE,
    description: 'Exponential Moving Average - 20 periods',
    defaultPeriod: 20,
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#06B6D4',
    lineStyle: 'dashed',
    lineWidth: 2
  },
  
  // Momentum Indicators
  {
    id: 'macd',
    name: 'MACD',
    displayName: 'MACD',
    type: IndicatorType.MOMENTUM,
    category: IndicatorCategory.MOMENTUM,
    description: 'Moving Average Convergence Divergence',
    periods: [12, 26, 9],
    paneRequirement: 'dedicated',
    defaultHeight: 0.15,
    rangeType: 'dynamic',
    color: '#2962ff'
  },
  {
    id: 'rsi',
    name: 'RSI',
    displayName: 'RSI (14)',
    type: IndicatorType.OSCILLATOR,
    category: IndicatorCategory.MOMENTUM,
    description: 'Relative Strength Index',
    defaultPeriod: 14,
    paneRequirement: 'shared',
    defaultHeight: 0.12,
    rangeType: 'fixed',
    fixedRange: [0, 100],
    color: '#9c27b0',
    lineWidth: 2
  },
  
  // Support & Resistance
  {
    id: 'resistance',
    name: 'Resistance',
    displayName: 'Resistance Level',
    type: IndicatorType.SUPPORT_RESISTANCE,
    category: IndicatorCategory.SUPPORT_RESISTANCE,
    description: 'Calculated resistance level',
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#EF4444',
    lineStyle: 'dashed',
    lineWidth: 2
  },
  {
    id: 'support',
    name: 'Support',
    displayName: 'Support Level',
    type: IndicatorType.SUPPORT_RESISTANCE,
    category: IndicatorCategory.SUPPORT_RESISTANCE,
    description: 'Calculated support level',
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#10B981',
    lineStyle: 'dashed',
    lineWidth: 2
  },
  
  // Volume Indicators
  {
    id: 'volume',
    name: 'Volume',
    displayName: 'Volume',
    type: IndicatorType.VOLUME,
    category: IndicatorCategory.VOLUME,
    description: 'Trading volume',
    paneRequirement: 'dedicated',
    defaultHeight: 0.10,
    rangeType: 'dynamic',
    color: '#6B7280'
  },
  
  // Future Indicators (placeholders)
  {
    id: 'bollinger',
    name: 'Bollinger Bands',
    displayName: 'BB (20, 2)',
    type: IndicatorType.BANDS,
    category: IndicatorCategory.VOLATILITY,
    description: 'Bollinger Bands - 20 periods, 2 std dev',
    periods: [20, 2],
    paneRequirement: 'none',
    rangeType: 'price',
    color: '#9333EA'
  },
  {
    id: 'stochastic',
    name: 'Stochastic',
    displayName: 'Stoch (14, 3)',
    type: IndicatorType.OSCILLATOR,
    category: IndicatorCategory.MOMENTUM,
    description: 'Stochastic Oscillator',
    periods: [14, 3],
    paneRequirement: 'shared',
    defaultHeight: 0.12,
    rangeType: 'fixed',
    fixedRange: [0, 100],
    color: '#6366F1'
  }
]