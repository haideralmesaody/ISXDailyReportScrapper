/**
 * MarketSummary - Market statistics summary card
 * Displays total value, volume, trades, and market breadth (advancers/decliners)
 * Following CLAUDE.md component standards
 */

'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatCurrency } from '@/lib/utils/market-helpers'
import { TrendingUp, TrendingDown, Minus, DollarSign, BarChart3, Activity } from 'lucide-react'

interface MarketSummaryProps {
  totalValue: number
  totalVolume: number
  totalTrades: number
  advancers: number
  decliners: number
  unchanged: number
}

export function MarketSummary({
  totalValue,
  totalVolume,
  totalTrades,
  advancers,
  decliners,
  unchanged
}: MarketSummaryProps) {
  // Calculate market breadth percentage
  const total = advancers + decliners + unchanged
  const advancersPercent = total > 0 ? (advancers / total) * 100 : 0
  const declinersPercent = total > 0 ? (decliners / total) * 100 : 0

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Market Summary</span>
          <span className="text-sm font-normal text-muted-foreground">{total} Tickers</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6" role="region" aria-label="Market statistics summary">
        {/* Top Stats Row */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="flex items-center gap-3 p-3 rounded-lg bg-green-500/5 border border-green-500/20" role="group" aria-label="Total market value">
            <div className="flex-shrink-0 p-2 rounded-lg bg-green-500/10">
              <DollarSign className="h-6 w-6 text-green-600 dark:text-green-400" aria-hidden="true" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Total Value</p>
              <p className="text-xl font-bold truncate" aria-label={`Total market value: ${formatCurrency(totalValue)}`}>{formatCurrency(totalValue)}</p>
            </div>
          </div>
          <div className="flex items-center gap-3 p-3 rounded-lg bg-blue-500/5 border border-blue-500/20" role="group" aria-label="Total trading volume">
            <div className="flex-shrink-0 p-2 rounded-lg bg-blue-500/10">
              <BarChart3 className="h-6 w-6 text-blue-600 dark:text-blue-400" aria-hidden="true" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Total Volume</p>
              <p className="text-xl font-bold truncate" aria-label={`Total volume: ${totalVolume.toLocaleString()} shares`}>{totalVolume.toLocaleString()}</p>
            </div>
          </div>
          <div className="flex items-center gap-3 p-3 rounded-lg bg-purple-500/5 border border-purple-500/20" role="group" aria-label="Total number of trades">
            <div className="flex-shrink-0 p-2 rounded-lg bg-purple-500/10">
              <Activity className="h-6 w-6 text-purple-600 dark:text-purple-400" aria-hidden="true" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Total Trades</p>
              <p className="text-xl font-bold truncate" aria-label={`Total trades: ${totalTrades.toLocaleString()}`}>{totalTrades.toLocaleString()}</p>
            </div>
          </div>
        </div>

        {/* Market Breadth */}
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold">Market Breadth</h3>
            <span className="text-xs text-muted-foreground">
              {advancers} up • {decliners} down • {unchanged} unch
            </span>
          </div>

          {/* Breadth Bar */}
          <div className="flex h-2 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="bg-green-500 transition-all duration-500"
              style={{ width: `${advancersPercent}%` }}
              title={`Advancers: ${advancersPercent.toFixed(1)}%`}
            />
            <div
              className="bg-red-500 transition-all duration-500"
              style={{ width: `${declinersPercent}%` }}
              title={`Decliners: ${declinersPercent.toFixed(1)}%`}
            />
          </div>

          {/* Breadth Stats */}
          <div className="grid grid-cols-3 gap-3">
            <div className="flex items-center gap-2 p-2 rounded bg-green-500/10">
              <TrendingUp className="h-4 w-4 text-green-500" />
              <div>
                <p className="text-xs text-muted-foreground">Gainers</p>
                <p className="text-lg font-bold text-green-500">{advancers}</p>
              </div>
            </div>
            <div className="flex items-center gap-2 p-2 rounded bg-red-500/10">
              <TrendingDown className="h-4 w-4 text-red-500" />
              <div>
                <p className="text-xs text-muted-foreground">Losers</p>
                <p className="text-lg font-bold text-red-500">{decliners}</p>
              </div>
            </div>
            <div className="flex items-center gap-2 p-2 rounded bg-gray-400/10">
              <Minus className="h-4 w-4 text-gray-400" />
              <div>
                <p className="text-xs text-muted-foreground">Flat</p>
                <p className="text-lg font-bold text-gray-400">{unchanged}</p>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
