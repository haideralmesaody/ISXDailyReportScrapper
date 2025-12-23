'use client'

import * as React from "react"
import { type LucideIcon, Loader2 } from "lucide-react"

import { cn } from "@/lib/utils"
import { Card } from "@/components/ui/card"
import { 
  trackLoadingState, 
  trackTiming, 
  getCurrentPage,
  generateCorrelationId,
  debug 
} from "@/lib/observability/no-data-metrics"

export interface DataLoadingStateProps {
  message?: string
  icon?: LucideIcon
  className?: string
  showCard?: boolean
  size?: 'sm' | 'default' | 'lg'
  // Observability props
  page?: string
  operation?: string
  onLoadingComplete?: (durationMs: number, success: boolean) => void
  trackPerformance?: boolean
}

const sizeClasses = {
  sm: {
    icon: 'h-6 w-6',
    text: 'text-sm',
    spacing: 'space-y-2',
    padding: 'p-4'
  },
  default: {
    icon: 'h-8 w-8',
    text: 'text-base',
    spacing: 'space-y-4',
    padding: 'p-8'
  },
  lg: {
    icon: 'h-12 w-12',
    text: 'text-lg',
    spacing: 'space-y-6',
    padding: 'p-12'
  }
}

export const DataLoadingState = React.forwardRef<
  HTMLDivElement,
  DataLoadingStateProps
>(({ 
  message = "Loading...", 
  icon: CustomIcon,
  className,
  showCard = true,
  size = 'default',
  page,
  operation = 'data_loading',
  onLoadingComplete,
  trackPerformance = true,
  ...props 
}, ref) => {
  const Icon = CustomIcon || Loader2
  const isSpinning = !CustomIcon // Only spin the default Loader2 icon
  const classes = sizeClasses[size]
  
  // Observability setup
  const correlationId = React.useMemo(() => generateCorrelationId(), [])
  const currentPage = page || getCurrentPage()
  const timingRef = React.useRef<ReturnType<typeof trackTiming> | null>(null)
  const mountTimeRef = React.useRef<number>(Date.now())

  // Track loading display and start timing
  React.useEffect(() => {
    if (trackPerformance) {
      timingRef.current = trackTiming(`loading_state_${operation}`)
    }

    debug.logComponentState('DataLoadingState', {
      correlation_id: correlationId,
      page: currentPage,
      operation,
      message,
      size,
      show_card: showCard,
      has_custom_icon: !!CustomIcon,
      mount_time: new Date(mountTimeRef.current).toISOString(),
    })
  }, [correlationId, currentPage, operation, message, size, showCard, CustomIcon, trackPerformance])

  // Cleanup and final tracking on unmount
  React.useEffect(() => {
    return () => {
      if (timingRef.current && trackPerformance) {
        const durationMs = Date.now() - mountTimeRef.current
        
        // End timing
        timingRef.current.end()
        
        // Track loading completion
        trackLoadingState(currentPage, durationMs, true) // Assume success on normal unmount
        
        // Call callback if provided
        onLoadingComplete?.(durationMs, true)
        
        debug.logPerformance(`DataLoadingState Unmount`, durationMs, {
          correlation_id: correlationId,
          operation,
          page: currentPage,
        })
      }
    }
  }, [correlationId, currentPage, operation, onLoadingComplete, trackPerformance])

  const content = (
    <div className={cn(
      "flex items-center justify-center",
      classes.padding
    )}>
      <div className={cn(
        "text-center",
        classes.spacing
      )} role="status" aria-live="polite">
        <div className="flex justify-center">
          <Icon className={cn(
            classes.icon,
            "text-muted-foreground",
            isSpinning && "animate-spin"
          )} />
        </div>
        {message && (
          <p className={cn(
            "text-muted-foreground",
            classes.text
          )}>
            {message}
          </p>
        )}
      </div>
    </div>
  )

  if (showCard) {
    return (
      <div 
        ref={ref}
        className={cn("min-h-screen p-8", className)}
        {...props}
      >
        <div className="max-w-7xl mx-auto">
          <Card>
            {content}
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div 
      ref={ref}
      className={cn("flex items-center justify-center", className)}
      {...props}
    >
      {content}
    </div>
  )
})

DataLoadingState.displayName = "DataLoadingState"