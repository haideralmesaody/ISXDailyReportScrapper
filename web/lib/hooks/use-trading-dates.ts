/**
 * Custom hook for fetching and managing ISX trading dates
 * Integrates with the /api/reports/trading-dates endpoint
 */

import { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api'

export interface TradingDate {
  date: string
  formatted: string
  isHoliday: boolean
  isFriday: boolean
  isSaturday: boolean
}

interface UseTradingDatesReturn {
  tradingDates: TradingDate[]
  isLoading: boolean
  error: string | null
  isTradingDay: (date: Date | string) => boolean
  getNextTradingDay: (date: Date | string) => Date | null
  calculateTradingDays: (fromDate: Date | string, toDate: Date | string) => number
  getTradingDaysInRange: (fromDate: Date | string, toDate: Date | string) => TradingDate[]
}

/**
 * Convert Date to YYYY-MM-DD format using local date components
 * Avoids timezone issues with toISOString()
 */
function formatDateLocal(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Generate array of ISO date strings between start and end dates (inclusive)
 */
function generateDateRange(startDate: Date, endDate: Date): string[] {
  const dates: string[] = []
  const current = new Date(startDate)

  while (current <= endDate) {
    dates.push(formatDateLocal(current))
    current.setDate(current.getDate() + 1)
  }

  return dates
}

/**
 * Hook to fetch and manage ISX trading dates
 * Provides utility functions for holiday checking, trading day calculations, etc.
 */
export function useTradingDates(): UseTradingDatesReturn {
  const [tradingDates, setTradingDates] = useState<TradingDate[]>([])
  const [tradingDatesSet, setTradingDatesSet] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Fetch trading dates on mount
  useEffect(() => {
    const fetchTradingDates = async () => {
      try {
        setIsLoading(true)
        setError(null)

        const response = await apiClient.getTradingDates()

        // Create set for O(1) lookup
        const tradingDatesSet = new Set(response.dates)

        // Generate date range: last 365 days + next 30 days
        const today = new Date()
        const startDate = new Date(today)
        startDate.setDate(startDate.getDate() - 365)
        const endDate = new Date(today)
        endDate.setDate(endDate.getDate() + 30)

        const allDates = generateDateRange(startDate, endDate)

        // Transform to TradingDate format with inferred holidays
        const dates: TradingDate[] = allDates.map((dateStr) => {
          const date = new Date(dateStr)
          const dayOfWeek = date.getDay()
          const isFriday = dayOfWeek === 5
          const isSaturday = dayOfWeek === 6
          const isTradingDay = tradingDatesSet.has(dateStr)

          return {
            date: dateStr,
            formatted: date.toLocaleDateString('en', {
              month: 'short',
              day: 'numeric',
              year: 'numeric'
            }),
            // Infer holiday: not trading day AND not Fri/Sat
            isHoliday: !isTradingDay && !isFriday && !isSaturday,
            isFriday,
            isSaturday
          }
        })

        setTradingDates(dates)
        setTradingDatesSet(tradingDatesSet)
      } catch (err) {
        console.error('Failed to fetch trading dates:', err)
        setError(err instanceof Error ? err.message : 'Failed to load trading dates')
      } finally {
        setIsLoading(false)
      }
    }

    fetchTradingDates()
  }, [])

  // Check if a given date is a trading day
  // Uses the backend-provided Set for accurate O(1) lookup
  const isTradingDay = (date: Date | string): boolean => {
    const dateStr = typeof date === 'string'
      ? date
      : formatDateLocal(date)

    return tradingDatesSet.has(dateStr)
  }

  // Get the next trading day after a given date
  const getNextTradingDay = (date: Date | string): Date | null => {
    const currentDate = typeof date === 'string' ? new Date(date) : date
    const currentDateStr = typeof date === 'string' ? date : formatDateLocal(currentDate)

    // Find trading dates after the current date
    const futureDates = tradingDates
      .filter(td => td.date > currentDateStr)
      .sort((a, b) => a.date.localeCompare(b.date))

    if (futureDates.length === 0) return null

    const next = futureDates[0]
    if (!next) return null

    return new Date(next.date)
  }

  // Calculate number of trading days between two dates (inclusive)
  const calculateTradingDays = (fromDate: Date | string, toDate: Date | string): number => {
    const fromDateStr = typeof fromDate === 'string'
      ? fromDate
      : formatDateLocal(fromDate)
    const toDateStr = typeof toDate === 'string'
      ? toDate
      : formatDateLocal(toDate)

    return tradingDates.filter(td =>
      td.date >= fromDateStr && td.date <= toDateStr
    ).length
  }

  // Get list of trading dates in a given range
  const getTradingDaysInRange = (fromDate: Date | string, toDate: Date | string): TradingDate[] => {
    const fromDateStr = typeof fromDate === 'string'
      ? fromDate
      : formatDateLocal(fromDate)
    const toDateStr = typeof toDate === 'string'
      ? toDate
      : formatDateLocal(toDate)

    return tradingDates.filter(td =>
      td.date >= fromDateStr && td.date <= toDateStr
    )
  }

  return {
    tradingDates,
    isLoading,
    error,
    isTradingDay,
    getNextTradingDay,
    calculateTradingDays,
    getTradingDaysInRange
  }
}
