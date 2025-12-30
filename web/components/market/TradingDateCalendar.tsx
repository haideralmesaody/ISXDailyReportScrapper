/**
 * Trading Date Calendar Component
 * Lightweight calendar showing only trading dates
 * Following CLAUDE.md React patterns
 */

'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { parseISODate, formatDateToISO, isTradingDate, getTradingDatesForMonth } from '@/lib/utils/date-helpers'

interface TradingDateCalendarProps {
  selectedDate: string
  tradingDates: string[]
  onDateSelect: (date: string) => void
  onClose?: () => void
}

const WEEKDAYS = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']
const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
]

/**
 * Check if date is in the past (before today)
 */
const isPast = (date: Date): boolean => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const compareDate = new Date(date)
  compareDate.setHours(0, 0, 0, 0)
  return compareDate < today
}

/**
 * Check if date is a weekend (Friday or Saturday in Iraq)
 */
const isWeekend = (date: Date): boolean => {
  const dayOfWeek = date.getDay()
  return dayOfWeek === 5 || dayOfWeek === 6 // Friday or Saturday
}

export function TradingDateCalendar({
  selectedDate,
  tradingDates,
  onDateSelect,
  onClose,
}: TradingDateCalendarProps) {
  // Initialize to selected date's month
  const selectedDateObj = parseISODate(selectedDate)
  const [currentYear, setCurrentYear] = useState(selectedDateObj.getFullYear())
  const [currentMonth, setCurrentMonth] = useState(selectedDateObj.getMonth())

  // Get trading dates for current month
  const monthTradingDates = getTradingDatesForMonth(currentYear, currentMonth, tradingDates)

  // Generate calendar grid (6 weeks x 7 days = 42 cells)
  const generateCalendarDays = () => {
    const firstDayOfMonth = new Date(currentYear, currentMonth, 1)
    const lastDayOfMonth = new Date(currentYear, currentMonth + 1, 0)
    const startDay = firstDayOfMonth.getDay() // 0-6 (Sunday-Saturday)
    const daysInMonth = lastDayOfMonth.getDate()

    const days: (Date | null)[] = []

    // Add empty cells for days before month starts
    for (let i = 0; i < startDay; i++) {
      days.push(null)
    }

    // Add all days in month
    for (let day = 1; day <= daysInMonth; day++) {
      days.push(new Date(currentYear, currentMonth, day))
    }

    // Fill remaining cells to complete 6 weeks
    const remainingCells = 42 - days.length
    for (let i = 0; i < remainingCells; i++) {
      days.push(null)
    }

    return days
  }

  const calendarDays = generateCalendarDays()

  const handlePrevMonth = () => {
    if (currentMonth === 0) {
      setCurrentMonth(11)
      setCurrentYear(currentYear - 1)
    } else {
      setCurrentMonth(currentMonth - 1)
    }
  }

  const handleNextMonth = () => {
    if (currentMonth === 11) {
      setCurrentMonth(0)
      setCurrentYear(currentYear + 1)
    } else {
      setCurrentMonth(currentMonth + 1)
    }
  }

  const handleDateClick = (date: Date) => {
    if (isTradingDate(date, tradingDates)) {
      const dateStr = formatDateToISO(date)
      onDateSelect(dateStr)
      if (onClose) onClose()
    }
  }

  const isDateSelected = (date: Date) => {
    return formatDateToISO(date) === selectedDate
  }

  const isToday = (date: Date) => {
    const today = new Date()
    return (
      date.getDate() === today.getDate() &&
      date.getMonth() === today.getMonth() &&
      date.getFullYear() === today.getFullYear()
    )
  }

  return (
    <div className="w-[320px] p-4 bg-popover text-popover-foreground border border-border rounded-lg shadow-xl">
      {/* Month/Year Header with Navigation */}
      <div className="flex items-center justify-between mb-4">
        <Button
          variant="outline"
          size="icon"
          onClick={handlePrevMonth}
          className="h-8 w-8"
        >
          <ChevronLeft className="h-4 w-4" />
        </Button>

        <div className="font-semibold text-sm">
          {MONTHS[currentMonth]} {currentYear}
        </div>

        <Button
          variant="outline"
          size="icon"
          onClick={handleNextMonth}
          className="h-8 w-8"
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>

      {/* Weekday Headers */}
      <div className="grid grid-cols-7 gap-1 mb-2">
        {WEEKDAYS.map((day, index) => {
          // In Iraq, Friday (index 5) and Saturday (index 6) are weekends
          const isWeekendHeader = index === 5 || index === 6
          return (
            <div
              key={day}
              className={`text-center text-xs font-medium py-1 ${
                isWeekendHeader
                  ? 'text-red-500 dark:text-red-400'
                  : 'text-muted-foreground'
              }`}
            >
              {day}
            </div>
          )
        })}
      </div>

      {/* Calendar Grid */}
      <div className="grid grid-cols-7 gap-1">
        {calendarDays.map((date, index) => {
          if (date === null) {
            return <div key={`empty-${index}`} className="h-9" />
          }

          const isTrading = isTradingDate(date, tradingDates)
          const selected = isDateSelected(date)
          const today = isToday(date)
          const isWeekendDay = isWeekend(date)
          const pastDate = isPast(date)

          return (
            <button
              key={index}
              onClick={() => handleDateClick(date)}
              disabled={!isTrading}
              className={`
                h-9 w-full rounded-md text-sm font-medium
                transition-all duration-150
                ${
                  selected
                    ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                    : today
                      ? 'bg-accent text-accent-foreground border-2 border-primary'
                      : isWeekendDay
                        ? 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 cursor-not-allowed'
                        : isTrading
                          ? 'bg-green-100 dark:bg-green-950 text-green-800 dark:text-green-200 hover:bg-green-200 dark:hover:bg-green-900'
                          : pastDate
                            ? 'bg-red-100 dark:bg-red-950 text-red-700 dark:text-red-300 cursor-not-allowed'
                            : 'bg-white dark:bg-transparent border border-muted text-muted-foreground cursor-not-allowed'
                }
              `}
              title={
                isWeekendDay
                  ? 'Weekend (Iraqi Fri/Sat)'
                  : isTrading
                    ? formatDateToISO(date)
                    : pastDate
                      ? 'Past holiday'
                      : 'Future non-trading day'
              }
            >
              {date.getDate()}
            </button>
          )
        })}
      </div>

      {/* Legend */}
      <div className="mt-4 pt-3 border-t border-border space-y-1.5">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="h-4 w-4 rounded bg-accent border-2 border-primary shrink-0" />
          <span>Today</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="h-4 w-4 rounded bg-primary shrink-0" />
          <span>Selected</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="h-4 w-4 rounded bg-green-500 shrink-0" />
          <span>Trading Day</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="h-4 w-4 rounded bg-gray-500 shrink-0" />
          <span>Weekend (Fri/Sat)</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="h-4 w-4 rounded bg-red-500 shrink-0" />
          <span>Past Holiday</span>
        </div>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <div className="h-4 w-4 rounded bg-white dark:bg-transparent border border-muted shrink-0" />
          <span>Future Non-Trading</span>
        </div>
        {monthTradingDates.length > 0 && (
          <div className="text-xs text-muted-foreground pt-1">
            {monthTradingDates.length} trading day{monthTradingDates.length !== 1 ? 's' : ''} this month
          </div>
        )}
      </div>
    </div>
  )
}
