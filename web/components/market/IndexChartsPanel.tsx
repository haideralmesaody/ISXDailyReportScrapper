/**
 * IndexChartsPanel Component
 *
 * Displays ISX60 and ISX15 index charts side-by-side with a date range selector.
 * Fetches historical index data from the backend API.
 */

'use client'

import { useState, useEffect, useCallback } from 'react'
import { IndexChart } from './IndexChart'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Loader2, AlertCircle, Calendar } from 'lucide-react'
import { useToast } from '@/lib/hooks/use-toast'

interface IndexChartsPanelProps {
  selectedDate: string // Current date from market overview (YYYY-MM-DD)
}

interface IndexData {
  dates: string[]
  ISX60: number[]
  ISX15: number[]
}

interface ChartDataPoint {
  date: string
  value: number
}

type PresetPeriod = '1M' | '3M' | '6M' | 'YTD' | '1Y'

export function IndexChartsPanel({ selectedDate }: IndexChartsPanelProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [indexData, setIndexData] = useState<IndexData | null>(null)

  // Date range state (default: YTD - Year-to-Date from selectedDate)
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [activePreset, setActivePreset] = useState<PresetPeriod>('YTD')

  const { toast } = useToast()

  // Initialize date range to YTD (Year-to-Date)
  useEffect(() => {
    if (!selectedDate) return

    const end = new Date(selectedDate)
    const yearStart = new Date(end.getFullYear(), 0, 1) // January 1 of current year

    setStartDate(yearStart.toISOString().split('T')[0] ?? yearStart.toISOString()) // YYYY-MM-DD
    setEndDate(selectedDate)
  }, [selectedDate])

  // Calculate date range for preset periods
  const calculatePresetDates = useCallback((preset: PresetPeriod) => {
    const end = new Date(selectedDate)
    let start: Date

    switch (preset) {
      case '1M':
        start = new Date(end)
        start.setMonth(end.getMonth() - 1)
        break
      case '3M':
        start = new Date(end)
        start.setMonth(end.getMonth() - 3)
        break
      case '6M':
        start = new Date(end)
        start.setMonth(end.getMonth() - 6)
        break
      case 'YTD':
        start = new Date(end.getFullYear(), 0, 1) // January 1
        break
      case '1Y':
        start = new Date(end)
        start.setFullYear(end.getFullYear() - 1)
        break
    }

    return {
      start: start.toISOString().split('T')[0] ?? start.toISOString(),
      end: selectedDate
    }
  }, [selectedDate])

  // Handle preset button click
  const handlePresetClick = useCallback((preset: PresetPeriod) => {
    const { start, end } = calculatePresetDates(preset)
    setStartDate(start)
    setEndDate(end)
    setActivePreset(preset)
  }, [calculatePresetDates])

  // Fetch index history data
  const fetchIndexHistory = useCallback(async () => {
    if (!startDate || !endDate) return

    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch(
        `/api/data/indices/history?start=${startDate}&end=${endDate}`
      )

      if (!response.ok) {
        throw new Error(`Failed to fetch index data: ${response.statusText}`)
      }

      const result = await response.json()

      if (!result.success || !result.data) {
        throw new Error('Invalid API response format')
      }

      setIndexData(result.data)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error occurred'
      setError(message)
      toast({
        title: 'Error Loading Index Data',
        description: message,
        variant: 'destructive',
      })
    } finally {
      setIsLoading(false)
    }
  }, [startDate, endDate, toast])

  // Fetch data when date range changes
  useEffect(() => {
    if (startDate && endDate) {
      fetchIndexHistory()
    }
  }, [startDate, endDate, fetchIndexHistory])

  // Transform API data to chart format
  const isx60ChartData: ChartDataPoint[] =
    indexData?.dates.map((date, i) => ({
      date,
      value: indexData.ISX60[i] ?? 0,
    })) || []

  const isx15ChartData: ChartDataPoint[] =
    indexData?.dates.map((date, i) => ({
      date,
      value: indexData.ISX15[i] ?? 0,
    })) || []

  const handleDateChange = () => {
    // Validate dates
    if (!startDate || !endDate) {
      toast({
        title: 'Invalid Date Range',
        description: 'Please select both start and end dates',
        variant: 'destructive',
      })
      return
    }

    if (new Date(startDate) > new Date(endDate)) {
      toast({
        title: 'Invalid Date Range',
        description: 'Start date must be before end date',
        variant: 'destructive',
      })
      return
    }

    // Trigger re-fetch (happens automatically via useEffect)
    fetchIndexHistory()
  }

  const presets: { label: PresetPeriod; description: string }[] = [
    { label: '1M', description: '1 Month' },
    { label: '3M', description: '3 Months' },
    { label: '6M', description: '6 Months' },
    { label: 'YTD', description: 'Year to Date' },
    { label: '1Y', description: '1 Year' }
  ]

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between mb-4">
          <CardTitle className="flex items-center gap-2">
            <Calendar className="h-5 w-5" />
            Index Performance
          </CardTitle>

          {/* Date Preset Buttons */}
          <div className="flex items-center gap-2">
            {presets.map((preset) => (
              <Button
                key={preset.label}
                variant={activePreset === preset.label ? 'default' : 'outline'}
                size="sm"
                onClick={() => handlePresetClick(preset.label)}
                disabled={isLoading}
                className="min-w-[60px]"
                aria-label={preset.description}
              >
                {preset.label}
              </Button>
            ))}
          </div>
        </div>

        {/* Date Range Selector */}
        <div className="flex items-center gap-2 flex-wrap">
            <label className="text-sm font-medium text-muted-foreground">From:</label>
            <Input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="w-40"
              disabled={isLoading}
            />

            <label className="text-sm font-medium text-muted-foreground">To:</label>
            <Input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="w-40"
              disabled={isLoading}
            />

            <Button
              onClick={handleDateChange}
              disabled={isLoading}
              size="sm"
              variant="outline"
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Loading...
                </>
              ) : (
                'Update'
              )}
            </Button>
          </div>
      </CardHeader>

      <CardContent>
        {error && (
          <div className="p-4 mb-4 bg-yellow-50 dark:bg-yellow-950 border border-yellow-200 dark:border-yellow-800 rounded-lg">
            <div className="flex items-start gap-2">
              <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400 mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-semibold text-yellow-900 dark:text-yellow-100 mb-1">
                  Index Data Not Available
                </p>
                <p className="text-sm text-yellow-800 dark:text-yellow-200 mb-2">
                  {error}
                </p>
                <p className="text-xs text-yellow-700 dark:text-yellow-300">
                  <strong>Required file:</strong> dist/data/reports/indexes/indexes.csv
                  <br />
                  <strong>How to fix:</strong> Run the index extraction tool to generate this file.
                </p>
              </div>
            </div>
          </div>
        )}

        {isLoading && !indexData && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* ISX60 Skeleton */}
            <div>
              <Skeleton className="h-6 w-32 mb-3" />
              <Skeleton className="h-[300px] w-full" />
              <Skeleton className="h-4 w-48 mt-2" />
            </div>

            {/* ISX15 Skeleton */}
            <div>
              <Skeleton className="h-6 w-32 mb-3" />
              <Skeleton className="h-[300px] w-full" />
              <Skeleton className="h-4 w-48 mt-2" />
            </div>
          </div>
        )}

        {!isLoading && indexData && indexData.dates.length === 0 && (
          <div className="flex items-center justify-center h-80">
            <div className="text-center">
              <AlertCircle className="h-8 w-8 mx-auto mb-4 text-muted-foreground" />
              <p className="text-muted-foreground">
                No index data available for the selected date range
              </p>
            </div>
          </div>
        )}

        {indexData && indexData.dates.length > 0 && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* ISX60 Chart */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-green-600 dark:text-green-400">
                ISX60 Index
              </h3>
              <IndexChart
                indexName="ISX60"
                data={isx60ChartData}
                height={300}
              />
              <div className="mt-2 text-sm text-muted-foreground">
                {indexData.dates.length} data points from {startDate} to {endDate}
              </div>
            </div>

            {/* ISX15 Chart */}
            <div>
              <h3 className="text-lg font-semibold mb-3 text-blue-600 dark:text-blue-400">
                ISX15 Index
              </h3>
              <IndexChart
                indexName="ISX15"
                data={isx15ChartData}
                height={300}
              />
              <div className="mt-2 text-sm text-muted-foreground">
                {indexData.dates.length} data points from {startDate} to {endDate}
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
