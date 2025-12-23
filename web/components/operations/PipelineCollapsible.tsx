/**
 * Pipeline Collapsible Wrapper
 * Provides collapsible container for pipeline stages while preserving all existing functionality
 */

'use client'

import React, { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import {
  ChevronDown,
  ChevronUp
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface PipelineCollapsibleProps {
  operationId: string
  title: string
  status: string
  progress: number
  stages: React.ReactNode[]
  defaultExpanded?: boolean
  className?: string
}

export function PipelineCollapsible({
  operationId,
  title,
  status,
  progress,
  stages,
  defaultExpanded = false,
  className
}: PipelineCollapsibleProps) {
  const [isOpen, setIsOpen] = useState(defaultExpanded)

  // Auto-expand when pipeline completes
  React.useEffect(() => {
    if (status === 'completed' || status === 'failed') {
      setIsOpen(true)
    }
  }, [status])

  const getStatusColor = () => {
    switch (status) {
      case 'running':
        return 'bg-blue-500'
      case 'completed':
        return 'bg-green-500'
      case 'failed':
        return 'bg-red-500'
      case 'paused':
        return 'bg-yellow-500'
      default:
        return 'bg-gray-500'
    }
  }

  const getStatusBadgeVariant = () => {
    switch (status) {
      case 'running':
        return 'default'
      case 'completed':
        return 'secondary'
      case 'failed':
        return 'destructive'
      case 'paused':
        return 'outline'
      default:
        return 'secondary'
    }
  }

  return (
    <Card
      className={cn("w-full", className)}
      data-operation-id={operationId}
    >
      <Collapsible open={isOpen} onOpenChange={setIsOpen}>
        {/* Collapsible Header - Always Visible */}
        <CollapsibleTrigger asChild>
          <CardHeader className="cursor-pointer hover:bg-muted/50 transition-colors pb-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-2">
                  <div className={cn("w-2 h-2 rounded-full", getStatusColor())} />
                  <CardTitle className="text-base font-semibold">
                    {title}
                  </CardTitle>
                </div>
                <Badge variant={getStatusBadgeVariant()} className="text-xs">
                  {status}
                </Badge>
              </div>

              {/* Expand/Collapse Icon Only */}
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0 hover:bg-muted"
              >
                {isOpen ? (
                  <ChevronUp className="h-4 w-4" />
                ) : (
                  <ChevronDown className="h-4 w-4" />
                )}
              </Button>
            </div>

            {/* Progress Bar - Always Visible */}
            {(status === 'running' || status === 'paused' || status === 'completed') && (
              <div className="mt-3">
                <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
                  <span>Pipeline Progress</span>
                  <span>{progress}%</span>
                </div>
                <div className="w-full bg-secondary rounded-full h-2">
                  <div
                    className={cn(
                      "h-2 rounded-full transition-all duration-300",
                      status === 'running' && "bg-blue-500",
                      status === 'completed' && "bg-green-500",
                      status === 'failed' && "bg-red-500",
                      status === 'paused' && "bg-yellow-500"
                    )}
                    style={{ width: `${progress}%` }}
                  >
                    {status === 'running' && (
                      <div className="h-full bg-white/20 animate-pulse rounded-full" />
                    )}
                  </div>
                </div>
              </div>
            )}
          </CardHeader>
        </CollapsibleTrigger>

        {/* Collapsible Content - Stage Tickets */}
        <CollapsibleContent>
          <CardContent className="pt-0">
            <div className="space-y-3">
              {stages.map((stage, index) => (
                <div key={index} className="border-l-2 border-muted pl-4">
                  {stage}
                </div>
              ))}
            </div>
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  )
}

export default PipelineCollapsible