/**
 * Market Overview Client Component
 * MVP version with mock data to demonstrate D3 treemap functionality
 * Following CLAUDE.md React hydration best practices
 */

'use client'

import { useState, useCallback, useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { MarketTreemap } from '@/components/market/MarketTreemap'
import { MarketSummary } from '@/components/market/MarketSummary'
import { IndexChartsPanel } from '@/components/market/IndexChartsPanel'
import { TreemapLegend } from '@/components/market/TreemapLegend'
import { TradingSessionCard } from '@/components/market/TradingSessionCard'
import { useHydration } from '@/lib/hooks'
import { Loader2, Calendar, AlertCircle } from 'lucide-react'
import { getNextTradingDate, getPrevTradingDate } from '@/lib/utils/date-helpers'
import { TickerData } from '@/types/market'
import { useToast } from '@/lib/hooks/use-toast'
import { HelpButton } from '@/components/guide/HelpButton'

// API function to fetch daily market data with 7-day price history for sparklines
async function fetchDailyMarketData(date: string): Promise<TickerData[]> {
  const response = await fetch(`/api/data/daily/${date}?include_history=true&history_days=7`)
  if (!response.ok) {
    throw new Error(`Failed to fetch data for ${date}`)
  }

  const result = await response.json()

  // Transform CSV data to TickerData format
  return result.data.map((row: any) => ({
    symbol: row.Symbol,
    name: row.CompanyName,
    price: parseFloat(row.ClosePrice) || 0,
    change: parseFloat(row.Change) || 0,
    changePercent: parseFloat(row.ChangePercent) || 0,
    tradedValue: parseFloat(row.Value) || 0,
    volume: parseInt(row.Volume) || 0,
    trades: parseInt(row.NumTrades) || 0,
    priceHistory: row.PriceHistory || []  // Map 7-day price history for sparklines
  }))
}

// Get today's date in YYYY-MM-DD format (hydration-safe)
function getTodayDate(): string {
  // Only run on client side to prevent hydration mismatches
  if (typeof window === 'undefined') {
    return '2025-01-01' // Server-side fallback
  }
  const today = new Date()
  return today.toISOString().split('T')[0]
}

export default function MarketOverviewClient() {
  const isHydrated = useHydration()
  const { toast } = useToast()

  const [selectedDate, setSelectedDate] = useState<string>('')
  const [tradingDates, setTradingDates] = useState<string[]>([])
  const [tickerData, setTickerData] = useState<TickerData[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [treemapDimensions, setTreemapDimensions] = useState({ width: 1200, height: 600 })
  const treemapContainerRef = useRef<HTMLDivElement>(null)

  // Fetch available trading dates and set default to latest
  useEffect(() => {
    if (!isHydrated) return

    const fetchTradingDates = async () => {
      try {
        const response = await fetch('/api/data/trading-dates')
        if (!response.ok) throw new Error('Failed to fetch trading dates')

        const result = await response.json()
        if (result.status === 'success' && result.dates && result.dates.length > 0) {
          setTradingDates(result.dates)
          // Set default to latest (last) date
          const latestDate = result.dates[result.dates.length - 1]
          setSelectedDate(latestDate)
        }
      } catch (err) {
        console.error('Error fetching trading dates:', err)
        toast({
          title: 'Error',
          description: 'Failed to load trading dates',
          variant: 'destructive',
        })
      }
    }

    fetchTradingDates()
  }, [isHydrated, toast])

  // Fetch data when date changes
  useEffect(() => {
    if (!isHydrated || !selectedDate) return

    const fetchData = async () => {
      setIsLoading(true)
      setError(null)

      try {
        const data = await fetchDailyMarketData(selectedDate)

        if (data.length === 0) {
          // Empty data = no trading on this date
          setError(`No trading data available for ${selectedDate}`)
          setTickerData([])

          // Suggest latest available date
          const latestDate = tradingDates[tradingDates.length - 1]
          toast({
            title: "No Trading Data",
            description: `No market activity on ${selectedDate}. Latest data: ${latestDate}`,
            variant: "destructive",
          })
        } else {
          setTickerData(data)
          setError(null)
          toast({
            title: "Data loaded",
            description: `Showing ${data.length} tickers for ${selectedDate}`,
          })
        }
      } catch (err) {
        const errorMsg = err instanceof Error ? err.message : 'Failed to fetch data'
        setError(errorMsg)
        setTickerData([]) // Empty, NOT mock data

        toast({
          title: "Error loading data",
          description: errorMsg,
          variant: "destructive",
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchData()
  }, [selectedDate, isHydrated, toast, tradingDates])

  // Auto-resize treemap based on treemap container width (responsive sizing)
  useEffect(() => {
    if (!treemapContainerRef.current) return

    // Container has px-6 padding (24px * 2 = 48px horizontal padding)
    const CONTAINER_PADDING = 48

    // Measure immediately on mount (don't wait for ResizeObserver)
    const measureContainer = () => {
      if (!treemapContainerRef.current) return

      const rect = treemapContainerRef.current.getBoundingClientRect()
      const containerWidth = rect.width
      const availableWidth = Math.floor(containerWidth - CONTAINER_PADDING)
      const height = Math.max(500, Math.min(availableWidth * 0.5, 800))

      console.log('[Immediate measurement]')
      console.log('  - Container outer width:', containerWidth, 'px')
      console.log('  - Padding (px-6):', CONTAINER_PADDING, 'px')
      console.log('  - Available for treemap:', availableWidth, 'px')
      console.log('  - Treemap height:', Math.floor(height), 'px')

      setTreemapDimensions({
        width: availableWidth,
        height: Math.floor(height)
      })
    }

    // Initial measurement
    measureContainer()

    // Watch for future changes
    const resizeObserver = new ResizeObserver(entries => {
      const { width: containerWidth } = entries[0].contentRect
      const availableWidth = Math.floor(containerWidth - CONTAINER_PADDING)
      const height = Math.max(500, Math.min(availableWidth * 0.5, 800))

      console.log('[ResizeObserver]')
      console.log('  - Container content width:', containerWidth, 'px')
      console.log('  - Padding (px-6):', CONTAINER_PADDING, 'px')
      console.log('  - Available for treemap:', availableWidth, 'px')
      console.log('  - Treemap height:', Math.floor(height), 'px')

      setTreemapDimensions({
        width: availableWidth,
        height: Math.floor(height)
      })
    })

    resizeObserver.observe(treemapContainerRef.current)

    return () => resizeObserver.disconnect()
  }, [])

  // Keyboard navigation: Arrow keys to move between trading dates
  useEffect(() => {
    if (!isHydrated) return

    const handleKeyDown = (e: KeyboardEvent) => {
      // Only handle arrow keys when not typing in an input
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return
      }

      if (e.key === 'ArrowLeft') {
        const prevDate = getPrevTradingDate(selectedDate, tradingDates)
        if (prevDate) {
          e.preventDefault()
          setSelectedDate(prevDate)
        }
      } else if (e.key === 'ArrowRight') {
        const nextDate = getNextTradingDate(selectedDate, tradingDates)
        if (nextDate) {
          e.preventDefault()
          setSelectedDate(nextDate)
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isHydrated, selectedDate, tradingDates])

  // Handle ticker click - navigate to strategy page
  // Using window.location.href for static export compatibility (router.push doesn't work in static mode)
  const handleTickerClick = useCallback((ticker: TickerData) => {
    console.log('[MarketOverview] Ticker clicked:', ticker.symbol)
    window.location.href = `/strategy?ticker=${ticker.symbol}`
  }, [])

  if (!isHydrated) {
    return (
      <div className="min-h-screen p-8 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
          <p className="text-muted-foreground">Initializing market overview...</p>
        </div>
      </div>
    )
  }

  // Calculate summary from current ticker data
  const summary = {
    totalValue: tickerData.reduce((sum, t) => sum + t.tradedValue, 0),
    totalVolume: tickerData.reduce((sum, t) => sum + t.volume, 0),
    totalTrades: tickerData.reduce((sum, t) => sum + t.trades, 0),
    advancers: tickerData.filter(t => t.changePercent > 0).length,
    decliners: tickerData.filter(t => t.changePercent < 0).length,
    unchanged: tickerData.filter(t => t.changePercent === 0).length
  }

  return (
    <div className="min-h-screen p-8 space-y-8 max-w-[1920px] mx-auto">
      <div>
        <div className="flex items-center gap-2">
          <h1 className="text-3xl font-bold">Market Overview</h1>
          <HelpButton
            section="market-overview"
            tooltip="Learn about market treemap and visualization"
          />
        </div>
        <p className="text-sm text-muted-foreground mt-1">
          Interactive treemap visualization with real ISX data
        </p>
      </div>

      {/* Trading Session Card with Date Picker */}
      <TradingSessionCard
        selectedDate={selectedDate}
        tradingDates={tradingDates}
        onDateChange={setSelectedDate}
        isLoading={isLoading}
        tickerCount={tickerData.length}
      />

      <MarketSummary {...summary} />

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div>
              <CardTitle>Market Treemap</CardTitle>
              <p className="text-sm text-muted-foreground">
                Size = Traded Value | Color = Price Change % | Click to analyze
              </p>
            </div>
            {isLoading && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading data...
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent className="px-0">
          {/* Color Legend */}
          {tickerData.length > 0 && (
            <div className="px-6">
              <TreemapLegend />
            </div>
          )}

          {!isLoading && tickerData.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-96 text-center px-6">
              <AlertCircle className="h-16 w-16 text-muted-foreground mb-4" />
              <h3 className="text-xl font-semibold mb-2">No Trading Data Available</h3>
              <p className="text-muted-foreground mb-4">
                {error || `No market activity recorded for ${selectedDate}`}
              </p>
              {tradingDates.length > 0 && (
                <Button
                  onClick={() => setSelectedDate(tradingDates[tradingDates.length - 1])}
                  variant="outline"
                >
                  View Latest Trading Day ({tradingDates[tradingDates.length - 1]})
                </Button>
              )}
            </div>
          ) : (
            <div ref={treemapContainerRef} className="w-full px-6">
              <MarketTreemap
                tickers={tickerData}
                onTickerClick={handleTickerClick}
                width={treemapDimensions.width}
                height={treemapDimensions.height}
                minCellSize={400}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Index Charts Panel with Date Range Selector */}
      <IndexChartsPanel selectedDate={selectedDate} />

      <div className="text-sm text-muted-foreground bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg p-4">
        <strong>✅ Real Data:</strong> This page now shows real ISX market data from CSV files.
        Select a date using the picker above to view historical trading data.
      </div>
    </div>
  )
}
