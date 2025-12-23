/**
 * Market Overview types for ISX Pulse
 * Following CLAUDE.md TypeScript strict mode standards
 */

// ============================================================================
// Index Data Types
// ============================================================================

export interface IndexDataPoint {
  date: string          // ISO date "2024-01-15"
  value: number         // Index value
  change: number        // Points change
  changePercent: number // % change
}

export interface MarketIndexData {
  name: 'ISX60' | 'ISX15'
  current: number
  change: number
  changePercent: number
  history: IndexDataPoint[] // Last 30 days
}

// ============================================================================
// Ticker Data Types (for Treemap Visualization)
// ============================================================================

export interface TickerData {
  symbol: string        // "BANK"
  name: string         // "Bank of Baghdad"
  price: number        // Current price
  change: number       // Price change
  changePercent: number // % change
  tradedValue: number  // Total traded value (size metric for treemap)
  volume: number       // Shares traded
  trades: number       // Number of trades
  priceHistory?: number[] // Last 7-10 days of close prices for sparkline
}

// ============================================================================
// Market Overview Data (Main Page Data Structure)
// ============================================================================

export interface MarketOverviewData {
  date: string
  indices: {
    ISX60: MarketIndexData
    ISX15: MarketIndexData
  }
  tickers: TickerData[]
  summary: {
    totalValue: number
    totalVolume: number
    totalTrades: number
    advancers: number
    decliners: number
    unchanged: number
  }
}

// ============================================================================
// Date Selection Types
// ============================================================================

export type ViewMode = 'single-day' | 'date-range'

export interface DateRangeSelection {
  startDate: string  // ISO date
  endDate: string    // ISO date
}
