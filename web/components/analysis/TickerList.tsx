'use client'

import React, { useState, useMemo, useCallback, memo } from 'react'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { Search, TrendingUp, TrendingDown, Minus, ChevronUp, ChevronDown } from 'lucide-react'
import { TickerData } from '@/types/analysis'
import { useHydration } from '@/lib/hooks/use-hydration'
import { TickerSparkline } from './ticker/TickerSparkline'

interface TickerListProps {
  tickers: TickerData[]
  selectedTicker: string | null
  onTickerSelect: (ticker: string) => void
  panelWidth?: number // Width of parent panel for responsive columns
}

type SortField = 'Ticker' | 'CompanyName' | 'LastPrice' | 'LastDate' | 'ChangePercent'
type SortDirection = 'asc' | 'desc'

interface SortConfig {
  field: SortField
  direction: SortDirection
}

// Column visibility breakpoints - Priority order: Ticker > Sparkline > Change% > Price > Date > Company
const COLUMN_BREAKPOINTS = {
  FULL: 280,         // All columns visible
  HIDE_COMPANY: 240, // Hide company name first (Priority 6)
  HIDE_DATE: 220,    // Hide date (Priority 5)
  HIDE_PRICE: 200,   // Hide price (Priority 4)
  MINIMAL: 180       // Only ticker (P1), change% (P3), sparkline (P2)
}

// Calculate visible columns based on panel width
function getVisibleColumns(width: number) {
  return {
    ticker: true,      // Priority 1: ALWAYS visible
    sparkline: true,   // Priority 2: ALWAYS visible
    change: true,      // Priority 3: ALWAYS visible
    price: width >= COLUMN_BREAKPOINTS.HIDE_PRICE,      // Priority 4
    date: width >= COLUMN_BREAKPOINTS.HIDE_DATE,        // Priority 5
    company: width >= COLUMN_BREAKPOINTS.HIDE_COMPANY   // Priority 6 (hide first)
  }
}

// Get dynamic grid template based on visible columns
function getGridTemplate(visibleCols: ReturnType<typeof getVisibleColumns>) {
  if (visibleCols.company && visibleCols.price && visibleCols.date) {
    // Full layout: Ticker | Company | Price | Date | Change | Sparkline
    return 'minmax(45px, 55px) minmax(90px, 1fr) minmax(55px, 70px) minmax(65px, 80px) minmax(55px, 70px) 35px'
  }
  if (visibleCols.price && visibleCols.date) {
    // No company: Ticker | Price | Date | Change | Sparkline
    return 'minmax(50px, 60px) minmax(60px, 80px) minmax(70px, 90px) minmax(60px, 80px) 40px'
  }
  if (visibleCols.price) {
    // No company, no date: Ticker | Price | Change | Sparkline
    return 'minmax(55px, 70px) minmax(65px, 85px) minmax(65px, 85px) 45px'
  }
  // Minimal: Ticker | Change | Sparkline (Priority: 1, 3, 2)
  return 'minmax(60px, 80px) minmax(70px, 100px) 50px'
}

// Memoize the component to prevent unnecessary re-renders
export const TickerList = memo(function TickerList({ tickers, selectedTicker, onTickerSelect, panelWidth = 280 }: TickerListProps) {
  const isHydrated = useHydration()
  const [searchTerm, setSearchTerm] = useState('')
  const [sortConfig, setSortConfig] = useState<SortConfig[]>([
    { field: 'LastDate', direction: 'desc' },      // Most recent first
    { field: 'ChangePercent', direction: 'desc' },  // Highest change first
    { field: 'Ticker', direction: 'asc' }           // Alphabetical
  ])

  // Handle multi-level sorting
  const handleSort = useCallback((field: SortField) => {
    setSortConfig(prev => {
      const existingIndex = prev.findIndex(s => s.field === field)

      if (existingIndex >= 0) {
        const newConfig = [...prev]
        const current = newConfig[existingIndex]

        if (current && current.direction === 'asc') {
          newConfig[existingIndex] = { ...current, direction: 'desc' }
        } else {
          // Remove this sort level
          newConfig.splice(existingIndex, 1)
        }
        return newConfig
      } else {
        // Add new sort level (max 3 levels)
        if (prev.length >= 3) {
          return [{ field, direction: 'asc' }]
        }
        return [...prev, { field, direction: 'asc' }]
      }
    })
  }, [])

  // Filter and sort tickers
  const filteredAndSortedTickers = useMemo(() => {
    let filtered = tickers.filter(ticker => {
      const search = searchTerm.toLowerCase()
      return (
        ticker.Ticker.toLowerCase().includes(search) ||
        ticker.CompanyName.toLowerCase().includes(search)
      )
    })

    // Apply multi-level sorting
    const sorted = [...filtered].sort((a, b) => {
      for (const sort of sortConfig) {
        let aVal: any = a[sort.field]
        let bVal: any = b[sort.field]

        // Handle date sorting
        if (sort.field === 'LastDate') {
          aVal = new Date(aVal).getTime()
          bVal = new Date(bVal).getTime()
        }

        // Handle numeric vs string comparison
        if (aVal === bVal) continue

        let comparison = 0
        if (typeof aVal === 'string' && typeof bVal === 'string') {
          comparison = aVal.localeCompare(bVal)
        } else if (typeof aVal === 'number' && typeof bVal === 'number') {
          comparison = aVal - bVal
        }

        if (comparison !== 0) {
          return sort.direction === 'asc' ? comparison : -comparison
        }
      }
      return 0
    })

    return sorted
  }, [tickers, searchTerm, sortConfig])

  // Format number with locale
  const formatNumber = useCallback((num: number, decimals = 2) => {
    return new Intl.NumberFormat('en-US', {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals
    }).format(num)
  }, [])

  // Format date for display - CLAUDE.md error handling pattern
  const formatDate = useCallback((dateStr: string) => {
    if (!isHydrated) return '---'  // Consistent placeholder during hydration

    const date = new Date(dateStr)
    if (isNaN(date.getTime())) {
      console.warn(`[TickerList] Invalid date format: ${dateStr}`)
      return 'Invalid'  // User-friendly error message
    }

    try {
      return new Intl.DateTimeFormat('en-US', {
        month: 'short',
        day: 'numeric'
      }).format(date)
    } catch (err) {
      console.error(`[TickerList] Date formatting failed:`, err)
      return dateStr.split('T')[0]  // Fallback: show YYYY-MM-DD
    }
  }, [isHydrated])

  // Get sort indicator
  const getSortIndicator = (field: SortField) => {
    const index = sortConfig.findIndex(s => s.field === field)
    if (index < 0) return null

    const sort = sortConfig[index]
    if (!sort) return null

    return (
      <span className="inline-flex items-center ml-1">
        {sort.direction === 'asc' ? (
          <ChevronUp className="h-3 w-3" />
        ) : (
          <ChevronDown className="h-3 w-3" />
        )}
        {sortConfig.length > 1 && (
          <span className="text-[10px] text-muted-foreground ml-0.5">
            {index + 1}
          </span>
        )}
      </span>
    )
  }

  // Calculate visible columns based on panel width
  const visibleColumns = getVisibleColumns(panelWidth)
  const gridTemplate = getGridTemplate(visibleColumns)

  return (
    <div className="flex flex-col h-full">
      {/* Search Bar */}
      <div className="p-3 border-b">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search ticker or company..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9 h-9"
          />
        </div>
      </div>

      {/* Table Header - Responsive columns based on panel width */}
      <div
        className="grid gap-2 px-3 py-2 border-b bg-muted/50 text-[10px] font-semibold text-muted-foreground uppercase tracking-wider"
        style={{ gridTemplateColumns: gridTemplate }}
      >
        <div
          className="cursor-pointer hover:text-foreground flex items-center"
          onClick={() => handleSort('Ticker')}
        >
          Ticker {getSortIndicator('Ticker')}
        </div>
        {visibleColumns.company && (
          <div
            className="cursor-pointer hover:text-foreground flex items-center"
            onClick={() => handleSort('CompanyName')}
          >
            Company {getSortIndicator('CompanyName')}
          </div>
        )}
        {visibleColumns.price && (
          <div
            className="text-right cursor-pointer hover:text-foreground flex items-center justify-end"
            onClick={() => handleSort('LastPrice')}
          >
            Price {getSortIndicator('LastPrice')}
          </div>
        )}
        {visibleColumns.date && (
          <div
            className="text-center cursor-pointer hover:text-foreground flex items-center justify-center"
            onClick={() => handleSort('LastDate')}
          >
            Date {getSortIndicator('LastDate')}
          </div>
        )}
        <div
          className="text-right cursor-pointer hover:text-foreground flex items-center justify-end"
          onClick={() => handleSort('ChangePercent')}
        >
          Change {getSortIndicator('ChangePercent')}
        </div>
        <div className="text-center flex items-center justify-center">
          Trend
        </div>
      </div>

      {/* Ticker List - Single rendering path, no ScrollArea */}
      <div className="flex-1 overflow-y-auto" style={{ contain: 'layout style paint' }}>
        <div className="pb-2">
          {filteredAndSortedTickers.map((ticker) => {
            const changePercent = ticker.ChangePercent || 0
            const didTrade = ticker.LastTradingStatus !== false // Default to true if undefined
            const isPositive = changePercent > 0
            const isNegative = changePercent < 0

            return (
              <div
                key={ticker.Ticker}
                className={cn(
                  "grid gap-2 px-3 py-1.5 hover:bg-muted/50 cursor-pointer transition-colors border-b",
                  selectedTicker === ticker.Ticker && "bg-primary/10 hover:bg-primary/15"
                )}
                onClick={() => onTickerSelect(ticker.Ticker)}
                style={{
                  willChange: 'transform',
                  gridTemplateColumns: gridTemplate
                }}
              >
                <div className="font-semibold text-sm truncate" title={ticker.Ticker}>
                  {ticker.Ticker}
                </div>
                {visibleColumns.company && (
                  <div className="text-xs text-muted-foreground truncate cursor-help" title={ticker.CompanyName}>
                    {ticker.CompanyName}
                  </div>
                )}
                {visibleColumns.price && (
                  <div className="text-right text-sm font-medium tabular-nums">
                    {formatNumber(ticker.LastPrice)}
                  </div>
                )}
                {visibleColumns.date && (
                  <div className="text-center text-xs text-muted-foreground font-medium">
                    {formatDate(ticker.LastDate)}
                  </div>
                )}
                <div className={cn(
                  "text-right text-sm font-semibold flex items-center justify-end tabular-nums",
                  didTrade && isPositive && "text-green-600 dark:text-green-500",
                  didTrade && isNegative && "text-red-600 dark:text-red-500",
                  !didTrade && "text-muted-foreground"
                )}>
                  {!didTrade ? (
                    <>
                      <Minus className="h-3 w-3 mr-0.5" />
                      <span>-</span>
                    </>
                  ) : (
                    <>
                      {isPositive && <TrendingUp className="h-3 w-3 mr-0.5" />}
                      {isNegative && <TrendingDown className="h-3 w-3 mr-0.5" />}
                      {!isPositive && !isNegative && <Minus className="h-3 w-3 mr-0.5" />}
                      {formatNumber(Math.abs(changePercent))}%
                    </>
                  )}
                </div>
                <div className="flex items-center justify-center">
                  <TickerSparkline
                    last10Days={ticker.Last10Days}
                    isPositive={isPositive}
                    isNegative={isNegative}
                    didTrade={didTrade}
                  />
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {/* Footer with ticker count */}
      <div className="p-2 border-t text-xs text-muted-foreground text-center">
        {filteredAndSortedTickers.length} of {tickers.length} tickers
      </div>
    </div>
  )
})
