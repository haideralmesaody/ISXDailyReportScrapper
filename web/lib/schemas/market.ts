/**
 * Zod validation schemas for Market Overview feature
 * Following CLAUDE.md standards for input validation
 */

import { z } from 'zod'

// ============================================================================
// Index Data Schemas
// ============================================================================

export const indexDataPointSchema = z.object({
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format (expected YYYY-MM-DD)'),
  value: z.number().min(0, 'Index value must be non-negative'),
  change: z.number(),
  changePercent: z.number()
})

export const marketIndexDataSchema = z.object({
  name: z.enum(['ISX60', 'ISX15']),
  current: z.number().min(0),
  change: z.number(),
  changePercent: z.number(),
  history: z.array(indexDataPointSchema)
})

// ============================================================================
// Ticker Data Schema (for Treemap)
// ============================================================================

export const tickerDataSchema = z.object({
  symbol: z.string().min(1, 'Symbol cannot be empty'),
  name: z.string().min(1, 'Name cannot be empty'),
  price: z.number().min(0, 'Price must be non-negative'),
  change: z.number(),
  changePercent: z.number(),
  tradedValue: z.number().min(0, 'Traded value must be non-negative'),
  volume: z.number().min(0, 'Volume must be non-negative'),
  trades: z.number().min(0, 'Trades must be non-negative')
})

// ============================================================================
// Market Overview Data Schema
// ============================================================================

export const marketOverviewDataSchema = z.object({
  date: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, 'Invalid date format (expected YYYY-MM-DD)'),
  indices: z.object({
    ISX60: marketIndexDataSchema,
    ISX15: marketIndexDataSchema
  }),
  tickers: z.array(tickerDataSchema),
  summary: z.object({
    totalValue: z.number().min(0),
    totalVolume: z.number().min(0),
    totalTrades: z.number().min(0),
    advancers: z.number().min(0),
    decliners: z.number().min(0),
    unchanged: z.number().min(0)
  })
})

// ============================================================================
// Date Range Schema
// ============================================================================

export const dateRangeSelectionSchema = z.object({
  startDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  endDate: z.string().regex(/^\d{4}-\d{2}-\d{2}$/)
})

export const viewModeSchema = z.enum(['single-day', 'date-range'])

// ============================================================================
// Type Exports
// ============================================================================

export type IndexDataPoint = z.infer<typeof indexDataPointSchema>
export type MarketIndexData = z.infer<typeof marketIndexDataSchema>
export type TickerData = z.infer<typeof tickerDataSchema>
export type MarketOverviewData = z.infer<typeof marketOverviewDataSchema>
export type DateRangeSelection = z.infer<typeof dateRangeSelectionSchema>
export type ViewMode = z.infer<typeof viewModeSchema>
