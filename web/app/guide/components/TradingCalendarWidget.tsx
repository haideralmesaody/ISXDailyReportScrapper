/**
 * Trading Calendar Widget
 * Compact calendar showing trading days and holidays for the Iraqi Stock Exchange
 * Displays Friday/Saturday (Iraqi weekend) in red, trading days in green
 */

'use client'

import React, { useState, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Calendar, ChevronLeft, ChevronRight, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTradingDates } from '@/lib/hooks/use-trading-dates'
import { getMonthName, getYear } from '@/lib/utils/trading-calendar'

interface TradingCalendarWidgetProps {
  className?: string
  compact?: boolean
}

export function TradingCalendarWidget({ className, compact = false }: TradingCalendarWidgetProps) {
  const { isLoading, error, isTradingDay } = useTradingDates()
  const [currentMonth, setCurrentMonth] = useState(new Date())

  // Generate calendar grid for current month
  const calendarDays = useMemo(() => {
    const year = currentMonth.getFullYear()
    const month = currentMonth.getMonth()

    // First day of the month
    const firstDay = new Date(year, month, 1)
    const lastDay = new Date(year, month + 1, 0)

    // Start from Sunday before first day
    const startDate = new Date(firstDay)
    startDate.setDate(startDate.getDate() - firstDay.getDay())

    // End on Saturday after last day
    const endDate = new Date(lastDay)
    endDate.setDate(endDate.getDate() + (6 - lastDay.getDay()))

    const days: Date[] = []
    const current = new Date(startDate)

    while (current <= endDate) {
      days.push(new Date(current))
      current.setDate(current.getDate() + 1)
    }

    return days
  }, [currentMonth])

  // Navigation handlers
  const goToPreviousMonth = () => {
    setCurrentMonth(prev => new Date(prev.getFullYear(), prev.getMonth() - 1))
  }

  const goToNextMonth = () => {
    setCurrentMonth(prev => new Date(prev.getFullYear(), prev.getMonth() + 1))
  }

  const goToToday = () => {
    setCurrentMonth(new Date())
  }

  // Check if date is in current month
  const isCurrentMonth = (date: Date): boolean => {
    return date.getMonth() === currentMonth.getMonth() &&
      date.getFullYear() === currentMonth.getFullYear()
  }

  // Check if date is today
  const isToday = (date: Date): boolean => {
    const today = new Date()
    return date.getDate() === today.getDate() &&
      date.getMonth() === today.getMonth() &&
      date.getFullYear() === today.getFullYear()
  }

  // Check if date is in the past (before today)
  const isPast = (date: Date): boolean => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)  // Normalize to midnight
    const compareDate = new Date(date)
    compareDate.setHours(0, 0, 0, 0)
    return compareDate < today
  }

  // Check if date is in the future (after today)
  const isFuture = (date: Date): boolean => {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const compareDate = new Date(date)
    compareDate.setHours(0, 0, 0, 0)
    return compareDate > today
  }

  if (compact) {
    // Compact version for sidebar
    return (
      <Card className={cn("border-muted", className)}>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm flex items-center gap-2">
            <Calendar className="h-4 w-4" />
            Trading Calendar
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {isLoading && (
            <p className="text-xs text-muted-foreground">Loading calendar...</p>
          )}

          {error && (
            <p className="text-xs text-red-500">Failed to load calendar</p>
          )}

          {!isLoading && !error && (
            <>
              <div className="flex items-center justify-between text-xs">
                <span className="font-medium">{getMonthName(currentMonth)} {getYear(currentMonth)}</span>
                <div className="flex gap-1">
                  <button
                    onClick={goToPreviousMonth}
                    className="p-0.5 hover:bg-muted rounded"
                    aria-label="Previous month"
                  >
                    <ChevronLeft className="h-3 w-3" />
                  </button>
                  <button
                    onClick={goToToday}
                    className="px-1.5 py-0.5 hover:bg-muted rounded text-[10px]"
                    title="Jump to current month"
                  >
                    Today
                  </button>
                  <button
                    onClick={goToNextMonth}
                    className="p-0.5 hover:bg-muted rounded"
                    aria-label="Next month"
                  >
                    <ChevronRight className="h-3 w-3" />
                  </button>
                </div>
              </div>

              {/* Weekday headers (Iraqi week: Sun-Thu are work days, Fri-Sat are weekend) */}
              <div className="grid grid-cols-7 gap-0.5 text-[10px] font-medium text-center">
                <div>Su</div>
                <div>Mo</div>
                <div>Tu</div>
                <div>We</div>
                <div>Th</div>
                <div className="text-red-500">Fr</div>
                <div className="text-red-500">Sa</div>
              </div>

              {/* Calendar grid */}
              <div className="grid grid-cols-7 gap-0.5">
                {calendarDays.map((day, i) => {
                  const dayOfWeek = day.getDay()
                  const isFriday = dayOfWeek === 5
                  const isSaturday = dayOfWeek === 6
                  const isWeekend = isFriday || isSaturday
                  const isTrading = isTradingDay(day)
                  const inCurrentMonth = isCurrentMonth(day)
                  const todayDate = isToday(day)
                  const pastDate = isPast(day)
                  const futureDate = isFuture(day)

                  return (
                    <div
                      key={i}
                      className={cn(
                        "h-6 flex items-center justify-center text-[10px] rounded",
                        !inCurrentMonth && "opacity-30",

                        // WEEKENDS (Fri/Sat) - Always GRAY
                        inCurrentMonth && isWeekend && !todayDate && "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400",
                        inCurrentMonth && isWeekend && todayDate && "bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-300 font-semibold ring-1 ring-gray-400",

                        // TRADING DAYS - GREEN
                        inCurrentMonth && !isWeekend && isTrading && !todayDate && "bg-green-100 dark:bg-green-950 text-green-800 dark:text-green-200 font-medium",
                        inCurrentMonth && !isWeekend && isTrading && todayDate && "bg-green-300 dark:bg-green-800 text-green-900 dark:text-green-100 font-semibold ring-1 ring-green-500",

                        // NON-TRADING PAST WEEKDAYS - RED (missed trading day = holiday in past)
                        inCurrentMonth && !isWeekend && !isTrading && pastDate && !todayDate && "bg-red-100 dark:bg-red-950 text-red-700 dark:text-red-300",
                        inCurrentMonth && !isWeekend && !isTrading && pastDate && todayDate && "bg-red-300 dark:bg-red-800 text-red-900 dark:text-red-100 font-semibold ring-1 ring-red-500",

                        // NON-TRADING FUTURE WEEKDAYS - WHITE/transparent (not yet occurred)
                        inCurrentMonth && !isWeekend && !isTrading && futureDate && "bg-white dark:bg-transparent border border-muted text-muted-foreground",
                      )}
                    >
                      {day.getDate()}
                    </div>
                  )
                })}
              </div>

              {/* Legend */}
              <div className="flex flex-wrap items-center gap-2 text-[10px] pt-1 border-t">
                <div className="flex items-center gap-1">
                  <div className="h-2 w-2 bg-green-500 rounded" />
                  <span>Trading</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="h-2 w-2 bg-gray-500 rounded" />
                  <span>Weekend</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="h-2 w-2 bg-red-500 rounded" />
                  <span>Past Holiday</span>
                </div>
                <div className="flex items-center gap-1">
                  <div className="h-2 w-2 bg-white dark:bg-transparent border border-muted rounded" />
                  <span>Future</span>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    )
  }

  // Full version for main content
  return (
    <Card className={cn(className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Calendar className="h-5 w-5" />
              ISX Trading Calendar
            </CardTitle>
            <CardDescription>
              Iraqi Stock Exchange trading days • Friday/Saturday weekend
            </CardDescription>
          </div>
          <Badge variant="outline">
            {getMonthName(currentMonth)} {getYear(currentMonth)}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {isLoading && (
          <p className="text-sm text-muted-foreground text-center py-8">Loading trading calendar...</p>
        )}

        {error && (
          <div className="flex items-center gap-2 text-sm text-red-500 p-4 bg-red-50 dark:bg-red-950 rounded-lg">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>Failed to load trading calendar: {error}</span>
          </div>
        )}

        {!isLoading && !error && (
          <>
            {/* Month navigation */}
            <div className="flex items-center justify-between">
              <button
                onClick={goToPreviousMonth}
                className="p-2 hover:bg-muted rounded-lg transition-colors"
                aria-label="Previous month"
              >
                <ChevronLeft className="h-5 w-5" />
              </button>

              <button
                onClick={goToToday}
                className="px-4 py-2 text-sm font-medium hover:bg-muted rounded-lg transition-colors"
                title="Jump to current month"
              >
                Today
              </button>

              <button
                onClick={goToNextMonth}
                className="p-2 hover:bg-muted rounded-lg transition-colors"
                aria-label="Next month"
              >
                <ChevronRight className="h-5 w-5" />
              </button>
            </div>

            {/* Weekday headers */}
            <div className="grid grid-cols-7 gap-2 text-sm font-medium text-center">
              <div>Sunday</div>
              <div>Monday</div>
              <div>Tuesday</div>
              <div>Wednesday</div>
              <div>Thursday</div>
              <div className="text-red-500">Friday</div>
              <div className="text-red-500">Saturday</div>
            </div>

            {/* Calendar grid */}
            <div className="grid grid-cols-7 gap-2">
              {calendarDays.map((day, i) => {
                const dayOfWeek = day.getDay()
                const isFriday = dayOfWeek === 5
                const isSaturday = dayOfWeek === 6
                const isWeekend = isFriday || isSaturday
                const isTrading = isTradingDay(day)
                const inCurrentMonth = isCurrentMonth(day)
                const todayDate = isToday(day)
                const pastDate = isPast(day)
                const futureDate = isFuture(day)

                return (
                  <div
                    key={i}
                    className={cn(
                      "h-12 flex items-center justify-center text-sm rounded-lg transition-colors",
                      !inCurrentMonth && "opacity-30",

                      // WEEKENDS (Fri/Sat) - Always GRAY
                      inCurrentMonth && isWeekend && !todayDate && "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400",
                      inCurrentMonth && isWeekend && todayDate && "bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-300 font-semibold ring-2 ring-gray-400",

                      // TRADING DAYS - GREEN
                      inCurrentMonth && !isWeekend && isTrading && !todayDate && "bg-green-100 dark:bg-green-950 text-green-800 dark:text-green-200 font-semibold hover:bg-green-200 dark:hover:bg-green-900",
                      inCurrentMonth && !isWeekend && isTrading && todayDate && "bg-green-300 dark:bg-green-800 text-green-900 dark:text-green-100 font-bold ring-2 ring-green-500",

                      // NON-TRADING PAST WEEKDAYS - RED (missed trading day = holiday in past)
                      inCurrentMonth && !isWeekend && !isTrading && pastDate && !todayDate && "bg-red-100 dark:bg-red-950 text-red-700 dark:text-red-300",
                      inCurrentMonth && !isWeekend && !isTrading && pastDate && todayDate && "bg-red-300 dark:bg-red-800 text-red-900 dark:text-red-100 font-bold ring-2 ring-red-500",

                      // NON-TRADING FUTURE WEEKDAYS - WHITE/transparent (not yet occurred)
                      inCurrentMonth && !isWeekend && !isTrading && futureDate && "bg-white dark:bg-transparent border border-muted text-muted-foreground",
                    )}
                  >
                    {day.getDate()}
                  </div>
                )
              })}
            </div>

            {/* Legend */}
            <div className="flex flex-wrap items-center justify-center gap-4 text-sm pt-2 border-t">
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 bg-green-500 rounded" />
                <span>Trading Day</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 bg-gray-500 rounded" />
                <span>Weekend (Fri/Sat)</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 bg-red-500 rounded" />
                <span>Past Holiday</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="h-3 w-3 bg-white dark:bg-transparent border border-muted rounded" />
                <span>Future Non-Trading</span>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}

export default TradingCalendarWidget
