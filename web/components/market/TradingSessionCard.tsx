/**
 * Trading Session Card Component
 * Professional date picker with metadata and quick navigation
 * Following CLAUDE.md React patterns
 */

'use client'

import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Calendar,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  TrendingUp,
  Clock,
  Loader2,
} from 'lucide-react'
import {
  formatTradingDate,
  getDaysAgo,
  getDataFreshnessIndicator,
  getNextTradingDate,
  getPrevTradingDate,
  getQuickDateShortcut,
  isLatestTradingDate,
} from '@/lib/utils/date-helpers'
import { TradingDateCalendar } from './TradingDateCalendar'

interface TradingSessionCardProps {
  selectedDate: string
  tradingDates: string[]
  onDateChange: (date: string) => void
  isLoading: boolean
  tickerCount: number
}

export function TradingSessionCard({
  selectedDate,
  tradingDates,
  onDateChange,
  isLoading,
  tickerCount,
}: TradingSessionCardProps) {
  const [calendarOpen, setCalendarOpen] = useState(false)

  const canGoPrev = getPrevTradingDate(selectedDate, tradingDates) !== null
  const canGoNext = getNextTradingDate(selectedDate, tradingDates) !== null
  const isLatest = isLatestTradingDate(selectedDate, tradingDates)

  const handlePrev = () => {
    const prevDate = getPrevTradingDate(selectedDate, tradingDates)
    if (prevDate) onDateChange(prevDate)
  }

  const handleNext = () => {
    const nextDate = getNextTradingDate(selectedDate, tradingDates)
    if (nextDate) onDateChange(nextDate)
  }

  const handleQuickShortcut = (type: 'latest' | 'yesterday' | 'lastWeek') => {
    const shortcutDate = getQuickDateShortcut(type, tradingDates)
    if (shortcutDate) onDateChange(shortcutDate)
  }

  // Get formatted metadata
  const formattedDate = formatTradingDate(selectedDate)
  const relativeTime = getDaysAgo(selectedDate)
  const freshness = getDataFreshnessIndicator(selectedDate)

  // Badge color variants
  const freshnessVariant =
    freshness.variant === 'success'
      ? 'default'
      : freshness.variant === 'warning'
        ? 'secondary'
        : 'outline'

  return (
    <Card className="border-2 border-border/50 shadow-sm hover:shadow-md transition-shadow">
      <CardContent className="p-6">
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
          {/* Left: Date Navigation */}
          <div className="flex items-center gap-4">
            <Button
              variant="outline"
              size="icon"
              onClick={handlePrev}
              disabled={!canGoPrev || isLoading}
              className="h-10 w-10 shrink-0"
              title="Previous trading day (←)"
            >
              <ChevronLeft className="h-5 w-5" />
            </Button>

            <div className="flex flex-col gap-1 min-w-0">
              {/* Main Date Display */}
              <div className="flex items-center gap-3">
                <Calendar className="h-5 w-5 text-muted-foreground shrink-0" />
                <h2 className="text-2xl font-bold tracking-tight truncate">
                  {formattedDate}
                </h2>

                {/* Calendar Picker Button */}
                <DropdownMenu open={calendarOpen} onOpenChange={setCalendarOpen}>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 shrink-0"
                      title="Select specific date"
                      disabled={isLoading}
                    >
                      <CalendarDays className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start" className="p-0">
                    <TradingDateCalendar
                      selectedDate={selectedDate}
                      tradingDates={tradingDates}
                      onDateSelect={onDateChange}
                      onClose={() => setCalendarOpen(false)}
                    />
                  </DropdownMenuContent>
                </DropdownMenu>

                {isLatest && (
                  <Badge variant="default" className="bg-green-600 hover:bg-green-700 shrink-0">
                    Latest
                  </Badge>
                )}
                {isLoading && (
                  <Loader2 className="h-5 w-5 animate-spin text-muted-foreground shrink-0" />
                )}
              </div>

              {/* Metadata Row */}
              <div className="flex items-center gap-3 text-sm text-muted-foreground">
                <div className="flex items-center gap-1.5">
                  <Clock className="h-3.5 w-3.5" />
                  <span>{relativeTime}</span>
                </div>
                <div className="h-4 w-px bg-border" aria-hidden="true" />
                <Badge variant={freshnessVariant} className="text-xs px-2 py-0.5">
                  {freshness.text}
                </Badge>
                {tickerCount > 0 && (
                  <>
                    <div className="h-4 w-px bg-border" aria-hidden="true" />
                    <div className="flex items-center gap-1.5">
                      <TrendingUp className="h-3.5 w-3.5" />
                      <span>{tickerCount} Tickers</span>
                    </div>
                  </>
                )}
                <div className="h-4 w-px bg-border" aria-hidden="true" />
                <span className="text-xs font-medium">Market Closed</span>
              </div>
            </div>

            <Button
              variant="outline"
              size="icon"
              onClick={handleNext}
              disabled={!canGoNext || isLoading}
              className="h-10 w-10 shrink-0"
              title="Next trading day (→)"
            >
              <ChevronRight className="h-5 w-5" />
            </Button>
          </div>

          {/* Right: Quick Shortcuts */}
          <div className="flex items-center gap-2 shrink-0">
            <div className="text-xs font-medium text-muted-foreground mr-1">
              Quick Jump:
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleQuickShortcut('latest')}
              disabled={isLatest || isLoading}
              className="h-9"
            >
              Latest
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleQuickShortcut('yesterday')}
              disabled={isLoading}
              className="h-9"
            >
              Yesterday
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleQuickShortcut('lastWeek')}
              disabled={isLoading}
              className="h-9"
            >
              Last Week
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
