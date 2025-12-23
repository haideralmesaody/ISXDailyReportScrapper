/**
 * Enhanced Operation Card with Inline Configuration
 * Implements progressive disclosure pattern for better UX
 */

'use client'

import React, { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { DatePickerField } from '@/components/ui/date-picker'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { 
  Play, 
  Loader2,
  Calendar,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  RefreshCw,
  Clock
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { OPERATION_DATE_DEFAULTS, DATE_VALIDATION } from '@/lib/constants'
import { validateAndUpdateDates } from '@/lib/date-utils'

// Import date picker styles
import '@/styles/datepicker.css'

interface EnhancedOperationCardProps {
  type: {
    id: string
    name: string
    description: string
    available?: boolean
    requiresDates?: boolean
  }
  icon: React.ComponentType<{ className?: string }>
  onStart: (params: any) => void
  isStarting: boolean
  className?: string
}

// Date range presets for quick selection
const DATE_PRESETS = [
  { label: 'Today', getValue: () => ({ from: new Date(), to: new Date() }) },
  { label: 'Last 7 days', getValue: () => {
    const to = new Date()
    const from = new Date()
    from.setDate(from.getDate() - 7)
    return { from, to }
  }},
  { label: 'Last 30 days', getValue: () => {
    const to = new Date()
    const from = new Date()
    from.setDate(from.getDate() - 30)
    return { from, to }
  }},
  { label: 'This month', getValue: () => {
    const now = new Date()
    const from = new Date(now.getFullYear(), now.getMonth(), 1)
    const to = new Date(now.getFullYear(), now.getMonth() + 1, 0)
    return { from, to }
  }},
  { label: 'Last month', getValue: () => {
    const now = new Date()
    const from = new Date(now.getFullYear(), now.getMonth() - 1, 1)
    const to = new Date(now.getFullYear(), now.getMonth(), 0)
    return { from, to }
  }},
  { label: 'Year to date', getValue: () => {
    const to = new Date()
    const from = new Date(to.getFullYear(), 0, 1)
    return { from, to }
  }}
]

export function EnhancedOperationCard({
  type,
  icon: Icon,
  onStart,
  isStarting,
  className
}: EnhancedOperationCardProps) {
  // State management - Initialize with defaults from Single Source of Truth
  const [expanded, setExpanded] = useState(false)
  const [fromDate, setFromDate] = useState<Date | null>(OPERATION_DATE_DEFAULTS.getFromDate())
  const [toDate, setToDate] = useState<Date | null>(OPERATION_DATE_DEFAULTS.getToDate())
  const [error, setError] = useState<string | null>(null)
  const [selectedPreset, setSelectedPreset] = useState<string | null>(null)

  // Check if this operation requires date configuration
  const requiresDates = type.requiresDates !== false && 
    (type.id === 'scraping' || type.id === 'full_pipeline')

  // Load saved dates from localStorage with smart date handling
  useEffect(() => {
    if (requiresDates) {
      const savedDates = localStorage.getItem(`operation_dates_${type.id}`)
      if (savedDates) {
        try {
          const { from, to } = JSON.parse(savedDates)
          // Use smart date validation to update old dates
          const validatedDates = validateAndUpdateDates({
            from: from || OPERATION_DATE_DEFAULTS.getFromDateString(),
            to: to || OPERATION_DATE_DEFAULTS.getToDateString()
          })
          setFromDate(validatedDates.from)
          setToDate(validatedDates.to)
        } catch (e) {
          // If parsing fails, set defaults
          setDefaultDates()
        }
      } else {
        setDefaultDates()
      }
    }
  }, [type.id, requiresDates])

  // Set default dates from Single Source of Truth
  const setDefaultDates = useCallback(() => {
    setFromDate(OPERATION_DATE_DEFAULTS.getFromDate())
    setToDate(OPERATION_DATE_DEFAULTS.getToDate())
  }, [])

  // Save dates to localStorage whenever they change
  useEffect(() => {
    if (requiresDates && fromDate && toDate) {
      localStorage.setItem(`operation_dates_${type.id}`, JSON.stringify({
        from: fromDate.toISOString(),
        to: toDate.toISOString()
      }))
    }
  }, [fromDate, toDate, type.id, requiresDates])

  // Validate dates
  const validateDates = useCallback((): boolean => {
    if (!requiresDates) return true

    if (!fromDate || !toDate) {
      setError('Please select both from and to dates')
      return false
    }

    if (fromDate > toDate) {
      setError('From date must be before or equal to to date')
      return false
    }

    const daysDiff = Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24))
    if (daysDiff > DATE_VALIDATION.MAX_DAYS_RANGE) {
      setError(`Date range cannot exceed ${DATE_VALIDATION.MAX_DAYS_RANGE} days`)
      return false
    }

    setError(null)
    return true
  }, [fromDate, toDate, requiresDates])

  // Handle preset selection
  const handlePresetSelect = useCallback((presetLabel: string) => {
    const preset = DATE_PRESETS.find(p => p.label === presetLabel)
    if (preset) {
      const { from, to } = preset.getValue()
      setFromDate(from)
      setToDate(to)
      setSelectedPreset(presetLabel)
      setError(null)
    }
  }, [])

  // Handle start operation
  const handleStart = useCallback(() => {
    if (!validateDates()) {
      setExpanded(true) // Expand to show error
      return
    }

    // Build parameters
    const params: any = {
      mode: 'full',
      steps: [{
        id: type.id,
        type: type.name,
        parameters: {}
      }]
    }

    // Add dates if required
    if (requiresDates && fromDate && toDate) {
      params.steps[0].parameters.from = fromDate.toISOString().split('T')[0]
      params.steps[0].parameters.to = toDate.toISOString().split('T')[0]
    }

    onStart(params)
  }, [type, fromDate, toDate, requiresDates, validateDates, onStart])

  // Calculate days in range
  const daysInRange = fromDate && toDate 
    ? Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24)) + 1
    : 0

  return (
    <Card 
      className={cn(
        "hover:shadow-lg transition-all duration-200 overflow-hidden",
        expanded && requiresDates ? "ring-2 ring-primary/20" : "",
        className
      )}
    >
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <Icon className="h-8 w-8 text-primary" />
          {type.available === false && (
            <Badge variant="secondary">Soon</Badge>
          )}
        </div>
        <CardTitle className="text-base">{type.name}</CardTitle>
        <CardDescription className="text-xs mt-1">
          {type.description}
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-3">
        {/* Quick date summary for operations that need dates */}
        {requiresDates && (
          <>
            <div 
              className={cn(
                "bg-muted/50 rounded-lg p-3 cursor-pointer transition-colors hover:bg-muted/70",
                expanded && "bg-primary/5 hover:bg-primary/10"
              )}
              onClick={() => setExpanded(!expanded)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm font-medium">Date Range</span>
                </div>
                {expanded ? (
                  <ChevronUp className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
              
              {!expanded && fromDate && toDate && (
                <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
                  <Clock className="h-3 w-3" />
                  <span>
                    {fromDate.toLocaleDateString()} - {toDate.toLocaleDateString()}
                  </span>
                  <Badge variant="outline" className="ml-auto text-xs">
                    {daysInRange} days
                  </Badge>
                </div>
              )}
            </div>

            {/* Expanded configuration */}
            <div className={cn(
              "space-y-3 overflow-hidden transition-all duration-200",
              expanded ? "max-h-96 opacity-100" : "max-h-0 opacity-0"
            )}>
              <Separator />
              
              {/* Date presets */}
              <div className="space-y-2">
                <Label className="text-xs text-muted-foreground">Quick Select</Label>
                <div className="grid grid-cols-3 gap-1">
                  {DATE_PRESETS.slice(0, 6).map(preset => (
                    <Button
                      key={preset.label}
                      variant={selectedPreset === preset.label ? "default" : "outline"}
                      size="sm"
                      className="text-xs h-7"
                      onClick={() => handlePresetSelect(preset.label)}
                    >
                      {preset.label}
                    </Button>
                  ))}
                </div>
              </div>

              {/* Custom date selection */}
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor={`from-${type.id}`} className="text-xs">From Date</Label>
                  <DatePickerField
                    selected={fromDate}
                    onChange={(date: Date | null) => {
                      setFromDate(date)
                      setSelectedPreset(null)
                      setError(null)
                    }}
                    placeholderText="Select start date"
                    maxDate={new Date()}
                    minDate={new Date(DATE_VALIDATION.MIN_DATE)}
                    className="w-full"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor={`to-${type.id}`} className="text-xs">To Date</Label>
                  <DatePickerField
                    selected={toDate}
                    onChange={(date: Date | null) => {
                      setToDate(date)
                      setSelectedPreset(null)
                      setError(null)
                    }}
                    placeholderText="Select end date"
                    minDate={fromDate || new Date(DATE_VALIDATION.MIN_DATE)}
                    maxDate={new Date()}
                    className="w-full"
                  />
                </div>
              </div>

              {/* Date range summary */}
              {fromDate && toDate && !error && (
                <Alert className="bg-blue-50 dark:bg-blue-950 border-blue-200 dark:border-blue-800">
                  <Calendar className="h-3 w-3" />
                  <AlertDescription className="text-xs">
                    <strong>{daysInRange}</strong> days selected
                    {daysInRange > 30 && (
                      <span className="block mt-1 text-muted-foreground">
                        Large date ranges may take longer to process
                      </span>
                    )}
                  </AlertDescription>
                </Alert>
              )}

              {/* Reset button */}
              <Button
                variant="ghost"
                size="sm"
                className="w-full h-7 text-xs"
                onClick={(e) => {
                  e.stopPropagation()
                  setDefaultDates()
                  setSelectedPreset(null)
                  setError(null)
                }}
              >
                <RefreshCw className="h-3 w-3 mr-1" />
                Reset to Defaults
              </Button>
            </div>
          </>
        )}

        {/* Error display */}
        {error && (
          <Alert variant="destructive" className="py-2">
            <AlertCircle className="h-3 w-3" />
            <AlertDescription className="text-xs">{error}</AlertDescription>
          </Alert>
        )}

        {/* Start button */}
        <Button 
          size="sm" 
          className={cn(
            "w-full",
            requiresDates && !expanded ? "mt-2" : "mt-3"
          )}
          disabled={isStarting || type.available === false}
          onClick={handleStart}
        >
          {isStarting ? (
            <>
              <Loader2 className="h-4 w-4 mr-1 animate-spin" />
              Starting...
            </>
          ) : (
            <>
              <Play className="h-4 w-4 mr-1" />
              Start {requiresDates && !expanded && daysInRange > 0 ? `(${daysInRange} days)` : ''}
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  )
}

export default EnhancedOperationCard