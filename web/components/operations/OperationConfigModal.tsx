/**
 * Professional Operation Configuration Modal
 * Uses react-datepicker for reliable date selection
 */

'use client'

import React, { useState, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { DatePickerField } from '@/components/ui/date-picker'
import { Loader2, Calendar, AlertCircle } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { OPERATION_DATE_DEFAULTS, DATE_VALIDATION } from '@/lib/constants'

// Import custom styles for the date picker
import '@/styles/datepicker.css'

interface OperationConfigModalProps {
  isOpen: boolean
  onClose: () => void
  onStart: (params: any) => void
  operationType: {
    id: string
    name: string
    description: string
  } | null
  isStarting: boolean
}

export function OperationConfigModal({
  isOpen,
  onClose,
  onStart,
  operationType,
  isStarting
}: OperationConfigModalProps) {
  // Initialize with default dates from constants (Single Source of Truth)
  const [fromDate, setFromDate] = useState<Date | null>(OPERATION_DATE_DEFAULTS.getFromDate())
  const [toDate, setToDate] = useState<Date | null>(OPERATION_DATE_DEFAULTS.getToDate())
  const [error, setError] = useState<string | null>(null)
  
  // Reset to default dates when modal opens
  useEffect(() => {
    if (isOpen && operationType) {
      // Always reset to defaults when opening modal
      setFromDate(OPERATION_DATE_DEFAULTS.getFromDate())
      setToDate(OPERATION_DATE_DEFAULTS.getToDate())
      setError(null)
    }
  }, [isOpen, operationType])
  
  const handleStart = () => {
    // Validate dates
    if ((operationType?.id === 'scraping' || operationType?.id === 'full_pipeline')) {
      if (!fromDate || !toDate) {
        setError('Please select both from and to dates')
        return
      }
      
      if (fromDate > toDate) {
        setError('From date must be before or equal to to date')
        return
      }
    }
    
    // Build the correct structure for backend
    const params: any = {
      mode: 'full',
      steps: [{
        id: operationType?.id || 'scraping',
        type: operationType?.name || 'Data Collection',
        parameters: {}
      }]
    }
    
    // Add date parameters for operations that need them
    // Backend expects 'from' and 'to' in step parameters
    if (operationType?.id === 'scraping' || operationType?.id === 'full_pipeline') {
      params.steps[0].parameters.from = fromDate?.toISOString().split('T')[0]
      params.steps[0].parameters.to = toDate?.toISOString().split('T')[0]
    }
    
    onStart(params)
  }
  
  if (!operationType) return null
  
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="panel-viewport-aware panel-safe-shadow panel-responsive-radius sm:max-w-[min(90vw,600px)] max-h-[80vh]">
        <DialogHeader>
          <DialogTitle className="text-xl font-semibold flex items-center gap-2">
            <Calendar className="h-5 w-5 text-primary" />
            Configure {operationType.name}
          </DialogTitle>
          <DialogDescription className="text-sm text-muted-foreground mt-2">
            {operationType.description || 'Set the parameters for this operation'}
          </DialogDescription>
        </DialogHeader>
        
        <div className="panel-content-scrollable panel-scroll-custom space-y-6 py-6">
          {/* Show date inputs for operations that need them */}
          {(operationType.id === 'scraping' || operationType.id === 'full_pipeline') && (
            <div className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="from-date" className="text-sm font-medium">
                  From Date
                </Label>
                <DatePickerField
                  selected={fromDate}
                  onChange={(date) => {
                    setFromDate(date)
                    setError(null)
                  }}
                  placeholderText="Select start date"
                  maxDate={new Date()}
                  minDate={new Date(DATE_VALIDATION.MIN_DATE)}
                />
                <p className="text-xs text-muted-foreground">
                  Start date for data collection (default: January 1, 2025)
                </p>
              </div>
              
              <div className="grid gap-2">
                <Label htmlFor="to-date" className="text-sm font-medium">
                  To Date
                </Label>
                <DatePickerField
                  selected={toDate}
                  onChange={(date) => {
                    setToDate(date)
                    setError(null)
                  }}
                  placeholderText="Select end date"
                  minDate={fromDate || new Date(DATE_VALIDATION.MIN_DATE)}
                  maxDate={new Date()}
                />
                <p className="text-xs text-muted-foreground">
                  End date for data collection (default: today)
                </p>
              </div>
              
              {fromDate && toDate && (
                <Alert className="bg-blue-50 dark:bg-blue-950 border-blue-200 dark:border-blue-800">
                  <Calendar className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                  <AlertDescription className="text-sm text-blue-800 dark:text-blue-200">
                    <div className="space-y-1">
                      <div>
                        <strong>Date Range Selected:</strong>
                      </div>
                      <div>
                        {fromDate.toLocaleDateString('en-US', { 
                          weekday: 'short',
                          year: 'numeric', 
                          month: 'long', 
                          day: 'numeric' 
                        })} 
                        {' → '}
                        {toDate.toLocaleDateString('en-US', { 
                          weekday: 'short',
                          year: 'numeric', 
                          month: 'long', 
                          day: 'numeric' 
                        })}
                      </div>
                      <div className="text-xs opacity-75">
                        {Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24))} days
                      </div>
                    </div>
                  </AlertDescription>
                </Alert>
              )}
            </div>
          )}
          
          {/* For other operations, show a simple message */}
          {operationType.id !== 'scraping' && operationType.id !== 'full_pipeline' && (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="text-sm">
                This operation will process existing data files.
                No additional configuration is required.
              </AlertDescription>
            </Alert>
          )}
          
          {/* Error display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
        </div>
        
        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isStarting}
            className="panel-touch-target w-full sm:w-auto"
          >
            Cancel
          </Button>
          <Button
            onClick={handleStart}
            disabled={isStarting}
            className="panel-touch-target w-full sm:w-auto bg-primary hover:bg-primary/90"
          >
            {isStarting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Starting Operation...
              </>
            ) : (
              <>
                <Calendar className="mr-2 h-4 w-4" />
                Start Operation
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default OperationConfigModal
