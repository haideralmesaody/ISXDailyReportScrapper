/**
 * Simplified Operation Display Component
 * 
 * Direct display of operation data from WebSocket - no calculations,
 * no complex state management, just pure display.
 */

'use client'

import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  CheckCircle, 
  XCircle, 
  AlertCircle, 
  Clock,
  Activity,
  ChevronDown,
  ChevronRight
} from 'lucide-react'
import { cn } from '@/lib/utils'

// Simple client-only hook
function useIsClient() {
  const [isClient, setIsClient] = useState(false)
  useEffect(() => {
    setIsClient(true)
  }, [])
  return isClient
}

// Status configuration
const statusConfig = {
  pending: { icon: Clock, color: 'text-gray-600', bgColor: 'bg-gray-50' },
  running: { icon: Activity, color: 'text-blue-600', bgColor: 'bg-blue-50' },
  completed: { icon: CheckCircle, color: 'text-green-600', bgColor: 'bg-green-50' },
  failed: { icon: XCircle, color: 'text-red-600', bgColor: 'bg-red-50' },
  cancelled: { icon: AlertCircle, color: 'text-orange-600', bgColor: 'bg-orange-50' }
}

interface OperationDisplayProps {
  operation: {
    operation_id: string
    name?: string
    status: string
    progress: number
    message?: string
    error?: string
    metadata?: Record<string, any>
    steps?: Array<{
      id: string
      name: string
      status: string
      progress: number
      message?: string
      metadata?: Record<string, any>
    }>
    started_at?: string
    updated_at?: string
  }
}

export function OperationDisplay({ operation }: OperationDisplayProps) {
  const isClient = useIsClient()
  const [expandedSteps, setExpandedSteps] = useState<Set<string>>(new Set())
  
  const config = statusConfig[operation.status as keyof typeof statusConfig] || statusConfig.pending
  const StatusIcon = config.icon
  
  const toggleStep = (stepId: string) => {
    setExpandedSteps(prev => {
      const next = new Set(prev)
      if (next.has(stepId)) {
        next.delete(stepId)
      } else {
        next.add(stepId)
      }
      return next
    })
  }
  
  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <StatusIcon className={cn("h-5 w-5", config.color)} />
              {operation.name || `Operation ${operation.operation_id.slice(0, 8)}`}
            </CardTitle>
            <CardDescription className="mt-2">
              <Badge variant="outline" className={config.color}>
                {operation.status}
              </Badge>
              {operation.started_at && isClient ? (
                <span className="ml-2 text-sm">
                  Started: {new Date(operation.started_at).toLocaleTimeString()}
                </span>
              ) : null}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Main Progress Bar - Direct from backend */}
        <div>
          <div className="flex justify-between text-sm mb-2">
            <span>Overall Progress</span>
            <span className="font-medium">{operation.progress}%</span>
          </div>
          <Progress value={operation.progress} className="h-2" />
        </div>
        
        {/* Error Display */}
        {operation.error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{operation.error}</AlertDescription>
          </Alert>
        )}
        
        {/* Message Display */}
        {operation.message && (
          <div className="text-sm text-muted-foreground">
            {operation.message}
          </div>
        )}
        
        {/* Metadata Display - No calculations, just show data */}
        {operation.metadata && Object.keys(operation.metadata).length > 0 && (
          <div className="border rounded-lg p-3 bg-muted/30">
            <h4 className="text-sm font-medium mb-2">Details</h4>
            <dl className="grid grid-cols-2 gap-x-4 gap-y-1 text-sm">
              {operation.metadata.files_downloaded !== undefined && (
                <>
                  <dt className="text-muted-foreground">Files Downloaded:</dt>
                  <dd className="font-medium">{operation.metadata.files_downloaded}</dd>
                </>
              )}
              {operation.metadata.total_expected !== undefined && (
                <>
                  <dt className="text-muted-foreground">Total Expected:</dt>
                  <dd className="font-medium">{operation.metadata.total_expected}</dd>
                </>
              )}
              {operation.metadata.current_file !== undefined && (
                <>
                  <dt className="text-muted-foreground">Current File:</dt>
                  <dd className="font-medium">{operation.metadata.current_file}</dd>
                </>
              )}
              {operation.metadata.current_page !== undefined && (
                <>
                  <dt className="text-muted-foreground">Current Page:</dt>
                  <dd className="font-medium">{operation.metadata.current_page}</dd>
                </>
              )}
              {operation.metadata.adjusted_remaining !== undefined && (
                <>
                  <dt className="text-muted-foreground">Remaining:</dt>
                  <dd className="font-medium">{operation.metadata.adjusted_remaining}</dd>
                </>
              )}
            </dl>
          </div>
        )}
        
        {/* Steps Display */}
        {Array.isArray(operation.steps) && operation.steps.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium">Steps</h4>
            {operation.steps.map(step => {
              const isExpanded = expandedSteps.has(step.id)
              const stepConfig = statusConfig[step.status as keyof typeof statusConfig] || statusConfig.pending
              const StepIcon = stepConfig.icon
              
              return (
                <div 
                  key={step.id}
                  className={cn(
                    "border rounded-lg p-3 transition-colors",
                    stepConfig.bgColor,
                    "dark:bg-opacity-10"
                  )}
                >
                  {/* Step Header */}
                  <div 
                    className="flex items-center justify-between cursor-pointer"
                    onClick={() => toggleStep(step.id)}
                  >
                    <div className="flex items-center gap-2">
                      {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                      <StepIcon className={cn("h-4 w-4", stepConfig.color)} />
                      <span className="font-medium text-sm">{step.name}</span>
                      <Badge variant="outline" className="text-xs">
                        {step.progress}%
                      </Badge>
                    </div>
                  </div>
                  
                  {/* Step Progress Bar */}
                  {step.status === 'running' && (
                    <div className="mt-2 ml-6">
                      <Progress value={step.progress} className="h-1" />
                    </div>
                  )}
                  
                  {/* Expanded Step Details */}
                  {isExpanded && step.metadata && (
                    <div className="mt-2 ml-6 text-sm space-y-1">
                      {step.message && (
                        <div className="text-muted-foreground">{step.message}</div>
                      )}
                      <dl className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
                        {Object.entries(step.metadata)
                          .filter(([key]) => !key.startsWith('_'))
                          .slice(0, 6)
                          .map(([key, value]) => (
                            <React.Fragment key={key}>
                              <dt className="text-muted-foreground">{key.replace(/_/g, ' ')}:</dt>
                              <dd className="font-medium">{String(value)}</dd>
                            </React.Fragment>
                          ))}
                      </dl>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default OperationDisplay