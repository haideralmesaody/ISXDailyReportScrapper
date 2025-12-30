/**
 * Market Overview helper functions
 * Pure, testable utility functions following CLAUDE.md standards
 */

import { TickerData } from '@/types/market'

/**
 * Calculate color for price change percentage
 * Uses a 7-level green/red scale for professional visualization
 *
 * @param changePercent - Price change percentage
 * @returns Hex color code
 */
export function getChangeColor(changePercent: number): string {
  if (changePercent >= 5) return '#10b981'      // Strong positive (green-500)
  if (changePercent >= 2) return '#34d399'      // Moderate positive (green-400)
  if (changePercent > 0) return '#6ee7b7'       // Weak positive (green-300)
  if (changePercent === 0) return '#9ca3af'     // Neutral (gray-400)
  if (changePercent > -2) return '#fca5a5'      // Weak negative (red-300)
  if (changePercent > -5) return '#f87171'      // Moderate negative (red-400)
  return '#ef4444'                              // Strong negative (red-500)
}

/**
 * Format currency values in IQD (Iraqi Dinar)
 * Converts large numbers to readable format with K/M/B suffixes
 *
 * @param value - Currency value in IQD
 * @returns Formatted string (e.g., "1.25M IQD")
 */
export function formatCurrency(value: number): string {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B IQD`
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M IQD`
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K IQD`
  }
  return `${value.toFixed(2)} IQD`
}

/**
 * Format percentage with +/- sign
 * Always shows 2 decimal places with explicit sign
 *
 * @param value - Percentage value
 * @returns Formatted string (e.g., "+5.23%", "-2.10%")
 */
export function formatPercent(value: number): string {
  const sign = value >= 0 ? '+' : ''
  return `${sign}${value.toFixed(2)}%`
}

/**
 * Calculate traded value (price × volume)
 *
 * @param price - Price per share
 * @param volume - Number of shares traded
 * @returns Total traded value
 */
export function calculateTradedValue(price: number, volume: number): number {
  return price * volume
}

/**
 * Aggregate ticker data for date range
 * Groups by symbol and calculates:
 * - Average price, change, changePercent
 * - Sum of tradedValue, volume, trades
 *
 * @param tickers - Array of ticker data (may include duplicates across dates)
 * @returns Aggregated ticker data
 */
export function aggregateTickerData(
  tickers: TickerData[]
): TickerData[] {
  // Handle empty array
  if (tickers.length === 0) {
    return []
  }

  // Group by symbol
  const grouped = new Map<string, TickerData[]>()

  tickers.forEach(ticker => {
    if (!grouped.has(ticker.symbol)) {
      grouped.set(ticker.symbol, [])
    }
    grouped.get(ticker.symbol)!.push(ticker)
  })

  // Aggregate each group
  return Array.from(grouped.entries()).map(([symbol, group]) => {
    const count = group.length
    const first = group[0]

    return {
      symbol,
      name: first?.name ?? symbol, // Use first occurrence for name
      price: group.reduce((sum, t) => sum + t.price, 0) / count,           // Average
      change: group.reduce((sum, t) => sum + t.change, 0) / count,         // Average
      changePercent: group.reduce((sum, t) => sum + t.changePercent, 0) / count, // Average
      tradedValue: group.reduce((sum, t) => sum + t.tradedValue, 0),      // Sum
      volume: group.reduce((sum, t) => sum + t.volume, 0),                // Sum
      trades: group.reduce((sum, t) => sum + t.trades, 0)                 // Sum
    }
  })
}
