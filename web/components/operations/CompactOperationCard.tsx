/**
 * Compact Operation Card
 * Minimal, clean design optimized for grid layout
 */

'use client'

import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  Settings2, 
  Loader2,
  Calendar,
  Zap,
  ChevronRight
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface CompactOperationCardProps {
  type: {
    id: string
    name: string
    description: string
    available?: boolean
    requiresDates?: boolean
  }
  icon: React.ComponentType<{ className?: string }>
  onConfigure: () => void
  onDirectStart?: () => void  // For operations that don't need configuration
  isStarting: boolean
  className?: string
  compact?: boolean  // New prop for compact size
  statusBadge?: string
  footerMessage?: string
}

export function CompactOperationCard({
  type,
  icon: Icon,
  onConfigure,
  onDirectStart,
  isStarting,
  className,
  compact = false,
  statusBadge,
  footerMessage
}: CompactOperationCardProps) {
  const [savedDates, setSavedDates] = useState<{ from: string; to: string } | null>(null)
  
  // Check if this operation requires date configuration
  const requiresDates = type.requiresDates !== false && 
    (type.id === 'scraping' || type.id === 'full_pipeline')

  // Load saved dates to show in badge
  useEffect(() => {
    if (requiresDates) {
      const saved = localStorage.getItem(`operation_dates_${type.id}`)
      if (saved) {
        try {
          setSavedDates(JSON.parse(saved))
        } catch {
          setSavedDates(null)
        }
      }
    }
  }, [type.id, requiresDates])

  // Format date range for display
  const formatDateRange = () => {
    if (!savedDates) return null
    
    const from = new Date(savedDates.from)
    const to = new Date(savedDates.to)
    const days = Math.ceil((to.getTime() - from.getTime()) / (1000 * 60 * 60 * 24)) + 1
    
    // Show compact format
    const fromMonth = from.toLocaleDateString('en', { month: 'short', day: 'numeric' })
    const toMonth = to.toLocaleDateString('en', { month: 'short', day: 'numeric' })
    
    return `${fromMonth} - ${toMonth} (${days}d)`
  }

  const dateRangeText = formatDateRange()
  
  // Handle button click based on operation type
  const handleClick = () => {
    // For processing, indices, and liquidity - start directly without configuration
    if ((type.id === 'processing' || type.id === 'indices' || type.id === 'liquidity') && onDirectStart) {
      onDirectStart()
    } else {
      // For scraping and full_pipeline, show configuration
      onConfigure()
    }
  }

  return (
    <Card
      className={cn(
        "group relative overflow-hidden transition-all duration-200",
        "hover:shadow-lg hover:scale-[1.02] hover:z-10",
        compact ? "min-h-[120px]" : "min-h-[160px]", // Compact height
        "w-full flex flex-col", // Flexible width
        type.available === false && "opacity-60",
        className
      )}
    >
      {/* Status indicator line */}
      <div className={cn(
        "absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r",
        requiresDates && savedDates 
          ? "from-primary to-primary/50" 
          : "from-muted to-muted"
      )} />

      <CardHeader className={cn(
        compact ? "pb-1" : "pb-2",
        "flex-none"
      )}>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={cn(
              compact ? "p-1.5" : "p-2",
              "rounded-lg transition-colors",
              "bg-primary/10 group-hover:bg-primary/20"
            )}>
              <Icon className={cn(
                compact ? "h-4 w-4" : "h-5 w-5",
                "text-primary"
              )} />
            </div>
            <div>
              <CardTitle className={cn(
                compact ? "text-base" : "text-lg",
                "font-semibold"
              )}>
                {type.name}
              </CardTitle>
              {type.available === false && (
                <Badge variant="secondary" className="mt-1 text-xs">
                  Coming Soon
                </Badge>
              )}
              {statusBadge && (
                <Badge variant="outline" className="mt-1 text-[10px] uppercase tracking-wide">
                  {statusBadge}
                </Badge>
              )}
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className={cn(
        "flex-1 flex flex-col justify-between",
        compact ? "pb-3" : "pb-4"
      )}>
        <div className="space-y-2">
          <CardDescription className="text-sm line-clamp-2">
            {type.description}
          </CardDescription>
          
          {/* Date range indicator */}
          {requiresDates && dateRangeText && (
            <div className="flex items-center gap-1.5 text-xs">
              <Calendar className="h-3 w-3 text-muted-foreground" />
              <span className="text-muted-foreground font-medium">
                {dateRangeText}
              </span>
            </div>
          )}
        </div>

        {/* Action button */}
        <Button
          onClick={handleClick}
          disabled={isStarting || type.available === false}
          size="default"
          className={cn(
            "w-full h-11 mt-3 group/btn transition-all",
            requiresDates && !savedDates && "animate-pulse"
          )}
        >
          {isStarting ? (
            <>
              <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
              Starting...
            </>
          ) : (
            <>
              {requiresDates ? (
                <>
                  <Settings2 className="h-4 w-4 mr-1.5" />
                  Configure & Start
                </>
              ) : (
                <>
                  <Zap className="h-4 w-4 mr-1.5" />
                  Quick Start
                </>
              )}
              <ChevronRight className="h-3 w-3 ml-auto opacity-50 group-hover/btn:opacity-100 group-hover/btn:translate-x-0.5 transition-all" />
            </>
          )}
        </Button>

        {footerMessage && (
          <p className="mt-2 text-xs text-muted-foreground">
            {footerMessage}
          </p>
        )}
      </CardContent>

      {/* Hover effect gradient */}
      <div className="absolute inset-0 bg-gradient-to-t from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
    </Card>
  )
}

export default CompactOperationCard
