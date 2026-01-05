/**
 * Type definitions for Analysis feature
 * Includes ticker data, chart configurations, and technical indicators
 */

// Ticker summary from CSV
export interface TickerSummary {
  Ticker: string
  CompanyName: string
  LastPrice: number
  LastDate: string
  TradingDays: number
  Last10Days: string // Comma-separated prices for sparkline
  TotalVolume: number
  TotalValue: number
  AveragePrice: number
  HighestPrice: number
  LowestPrice: number
  ChangePercent?: number
  LastTradingStatus?: boolean // true if stock traded on LastDate, false if not
}

// Historical trading data for a ticker
export interface TickerHistoricalData {
  date: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  value?: number
  trades?: number
  change?: number
  changePercent?: number
}

// Highcharts data point formats
export interface OHLCDataPoint {
  x: number // timestamp
  open: number
  high: number
  low: number
  close: number
}

export interface VolumeDataPoint {
  x: number // timestamp
  y: number // volume
  color?: string // optional color based on price movement
}

// Technical indicator configuration
export interface IndicatorConfig {
  id: string
  type: string
  name: string
  params?: Record<string, any>
  yAxis?: number
  color?: string
  visible?: boolean
}

// Chart configuration
export interface ChartConfig {
  ticker: string
  data: TickerHistoricalData[]
  indicators: IndicatorConfig[]
  chartType: 'candlestick' | 'ohlc' | 'line' | 'area' | 'heikinashi' | 'hollowcandlestick'
  timeRange: '1D' | '1W' | '1M' | '3M' | '6M' | 'YTD' | '1Y' | 'ALL'
}

// Stock tools configuration
export interface StockToolsConfig {
  enabled: boolean
  buttons: string[]
  theme?: 'light' | 'dark'
}

// Type alias for backwards compatibility and clarity
export type TickerData = TickerSummary

// Analysis page state
export interface AnalysisState {
  tickers: TickerSummary[]
  selectedTicker: string | null
  historicalData: TickerHistoricalData[]
  loading: boolean
  error: string | null
  chartConfig: ChartConfig
}

// Sort configuration
export interface SortConfig {
  column: keyof TickerSummary
  direction: 'asc' | 'desc'
}

// Filter configuration
export interface FilterConfig {
  searchTerm: string
  minPrice?: number
  maxPrice?: number
  minVolume?: number
  maxVolume?: number
  minChange?: number
  maxChange?: number
}