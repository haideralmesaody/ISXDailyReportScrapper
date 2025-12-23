/**
 * Pipeline Summary Card
 * Shows compact pipeline overview with configuration and expandable details
 */

'use client'

import React, { useState, useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import {
  ChevronDown,
  ChevronUp,
  Settings2,
  Zap,
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
  Activity
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { getTotalStages, getStageInfo, getStageOrder, getPipelineStageOrder } from '@/lib/operations/stage-mapping'

interface PipelineSummaryCardProps {
  operation: {
    operation_id: string
    status: string
    progress: number
    name?: string
    message?: string
    error?: string
    metadata?: {
      pipeline_summary?: {
        total_duration?: string
        total_files?: number
        stages_completed?: number
        total_stages?: number
        last_run?: string
      }
    }
  }
  onConfigure: () => void
  onDirectStart?: () => void
  isStarting?: boolean
  stages?: React.ReactNode[]
  className?: string
  defaultExpanded?: boolean
  showQuickStart?: boolean
  primaryActionLabel?: string
}

export function PipelineSummaryCard({
  operation,
  onConfigure,
  onDirectStart,
  isStarting = false,
  stages = [],
  className,
  defaultExpanded = false,
  showQuickStart = true,
  primaryActionLabel
}: PipelineSummaryCardProps) {
  const [isOpen, setIsOpen] = useState(defaultExpanded)
  const canQuickStart = showQuickStart && typeof onDirectStart === 'function'
  const primaryDisabled = isStarting || (canQuickStart && operation.status === 'running')

  const stageStepsRaw = Array.isArray(operation.steps) ? operation.steps : []
  const canonicalStageData = operation.stageContext?.allStages || []
  // Memoize pipeline stage order to prevent TDZ errors
  const pipelineStageOrder = useMemo(() => {
    try {
      return getPipelineStageOrder()
    } catch (error) {
      console.warn('Pipeline stage order failed:', error)
      return []
    }
  }, [])

  // Memoize fallback order to prevent TDZ errors
  const fallbackOrder = useMemo(() => {
    try {
      return getPipelineStageOrder()
    } catch (error) {
      console.warn('Pipeline stage order fallback failed:', error)
      return []
    }
  }, [])

  const normalizedStages = useMemo(() => {
    const mergedStages = stageStepsRaw.length ? stageStepsRaw : canonicalStageData
    if (!mergedStages.length) {
      try {
        return pipelineStageOrder.map(stageId => {
          const info = getStageInfo(stageId)
          return {
            id: stageId,
            name: info?.name || stageId,
            status: 'pending',
            progress: 0,
            phase: undefined,
            metadata: {}
          }
        })
      } catch (error) {
        console.warn('Pipeline stage order mapping failed:', error)
        return []
      }
    }

    return mergedStages.map((step, index) => {
      const stageMetadata = step.metadata || step.stageMetadata || {}
      const stageId = stageMetadata.stage_id || step.id || step.step_id || fallbackOrder[index] || `stage-${index + 1}`

      const info = useMemo(() => {
        try {
          return stageId ? getStageInfo(stageId as any) : null
        } catch (error) {
          console.warn('Stage info failed:', error)
          return null
        }
      }, [stageId])

      const progressFromMetadata = typeof stageMetadata.progress_percent === 'number'
        ? stageMetadata.progress_percent
        : (typeof step.progress_percent === 'number' ? step.progress_percent : undefined)
      const resolvedStatus = step.status || stageMetadata.status || 'pending'

      return {
        id: stageId,
        name: step.name || info?.name || `Stage ${index + 1}`,
        status: resolvedStatus,
        progress: progressFromMetadata ?? (typeof step.progress === 'number' ? step.progress : 0),
        phase: stageMetadata.phase as string | undefined,
        metadata: stageMetadata
      }
    })
  }, [stageStepsRaw, canonicalStageData, fallbackOrder, pipelineStageOrder])

  const resolvedTotalStages = useMemo(() => {
    try {
      return normalizedStages.length || getTotalStages()
    } catch (error) {
      console.warn('Total stages calculation failed:', error)
      return 6
    }
  }, [normalizedStages.length])

  const completedStages = useMemo(() => {
    return normalizedStages?.filter(stage => stage.status === 'completed').length || 0
  }, [normalizedStages])

  const runningStage = useMemo(() => {
    return normalizedStages?.find(stage => stage.status === 'running') || null
  }, [normalizedStages])

  const pendingStage = useMemo(() => {
    return normalizedStages?.find(stage => stage.status === 'pending') || null
  }, [normalizedStages])

  const currentStage = useMemo(() => {
    return runningStage || pendingStage ||
           normalizedStages?.find(stage => stage.progress > 0) ||
           (normalizedStages?.length ? normalizedStages[normalizedStages.length - 1] : null)
  }, [runningStage, pendingStage, normalizedStages])

  const currentStageId = currentStage?.id

  const currentStageInfo = useMemo(() => {
    try {
      return currentStageId ? getStageInfo(currentStageId as any) : null
    } catch (error) {
      console.warn('Current stage info failed:', error)
      return null
    }
  }, [currentStageId])

  const currentStageNumber = useMemo(() => {
    try {
      return currentStageId ? getStageOrder(currentStageId as any) : Math.min(completedStages + 1, resolvedTotalStages)
    } catch (error) {
      console.warn('Stage number calculation failed:', error)
      return completedStages + 1
    }
  }, [currentStageId, completedStages, resolvedTotalStages])
  const runningContribution = normalizedStages.reduce((sum, stage) => {
    if (stage.status === 'running' && typeof stage.progress === 'number') {
      return sum + stage.progress / resolvedTotalStages
    }
    return sum
  }, 0)

  const derivedPipelineProgress = resolvedTotalStages > 0
    ? Math.min(100, Math.round((completedStages / resolvedTotalStages) * 100 + runningContribution))
    : (typeof operation.progress === 'number' ? operation.progress : 0)
  const headerProgress = typeof currentStage?.progress === 'number'
    ? currentStage.progress
    : derivedPipelineProgress

  // Auto-expand when pipeline completes
  React.useEffect(() => {
    if (operation.status === 'completed' || operation.status === 'failed') {
      setIsOpen(true)
    }
  }, [operation.status])

  const getStatusColor = () => {
    switch (operation.status) {
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
    switch (operation.status) {
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

  const getStatusIcon = () => {
    switch (operation.status) {
      case 'running':
        return <Loader2 className="h-4 w-4 animate-spin" />
      case 'completed':
        return <CheckCircle2 className="h-4 w-4" />
      case 'failed':
        return <XCircle className="h-4 w-4" />
      default:
        return <Activity className="h-4 w-4" />
    }
  }

  const pipelineSummary = operation.metadata?.pipeline_summary
  const statusLabel = operation.status.replace('_', ' ')
  const statusIcon = getStatusIcon()

  return (
    <Card
      className={cn("w-full", className)}
      data-operation-id={operation.operation_id}
      data-testid={`pipeline-summary-${operation.operation_id}`}
    >
      <Collapsible open={isOpen} onOpenChange={setIsOpen}>
        {/* Summary Header - Always Visible */}
        <CollapsibleTrigger asChild>
          <CardHeader
            data-testid={`pipeline-summary-trigger-${operation.operation_id}`}
            className="cursor-pointer hover:bg-muted/50 transition-colors pb-3 md:pb-4"
          >
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div className="flex items-start gap-3 flex-1">
                <div className={cn("w-2.5 h-2.5 mt-1 rounded-full", getStatusColor())} />
                <div>
                  <CardTitle className="text-lg font-semibold">
                    {operation.name || 'Full Pipeline'}
                  </CardTitle>
                  <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                    <Badge
                      variant={getStatusBadgeVariant()}
                      className="flex items-center gap-1 px-3 py-1.5 text-sm font-bold uppercase tracking-wide"
                    >
                      {statusIcon}
                      <span className="capitalize">{statusLabel}</span>
                    </Badge>
                    <span className="font-semibold text-foreground">
                      Stage {Math.min(currentStageNumber, resolvedTotalStages)} of {resolvedTotalStages}
                      {currentStageInfo?.name && (
                        <> &middot; {currentStageInfo.name}</>
                      )}
                      {typeof headerProgress === 'number' && (
                        <> &middot; {Math.round(headerProgress)}%</>
                      )}
                    </span>
                  </div>
                </div>
              </div>
              <div className="text-sm text-muted-foreground font-medium">
                {completedStages} / {resolvedTotalStages} stages complete
              </div>
            </div>
            {/* Stage timeline */}
            {normalizedStages && normalizedStages.length > 0 && (() => {
              const stageTimeline = useMemo(() => {
                return normalizedStages.map((stage, index) => {
                  const status =
                    stage.status === 'completed' ? 'completed' :
                    stage.status === 'failed' ? 'failed' :
                    stage.status === 'running' ? 'running' : 'pending'

                  return {
                    key: `${operation.operation_id}-timeline-${stage.id}-${index}`,
                    stage,
                    index,
                    status
                  }
                })
              }, [normalizedStages, operation.operation_id])

              return (
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  {stageTimeline.map(({ key, stage, index, status }) => (
                    <React.Fragment key={key}>
                      <div
                        className={cn(
                          "flex items-center gap-2 rounded-full border px-3 py-1 text-xs font-medium",
                          status === 'completed' && "bg-green-500/10 border-green-500/30 text-green-600",
                          status === 'running' && "bg-blue-500/10 border-blue-500/30 text-blue-600",
                          status === 'failed' && "bg-red-500/10 border-red-500/30 text-red-600",
                          status === 'pending' && "border-border text-muted-foreground"
                        )}
                      >
                        <span className="text-[10px] font-bold tracking-wide">{index + 1}</span>
                        <span>{stage.name}</span>
                      </div>
                      {index < normalizedStages.length - 1 && (
                        <div className="h-px w-6 bg-border" />
                      )}
                    </React.Fragment>
                  ))}
                </div>
              )
            })()}
            {/* Pipeline Progress Bar */}
            {(operation.status === 'running' || operation.status === 'paused' || operation.status === 'completed' || normalizedStages.length > 0) && (
              <div className="mt-4">
                <div className="flex items-center justify-between text-xs text-muted-foreground mb-2">
                  <span>Pipeline Progress</span>
                  <span>{derivedPipelineProgress}%</span>
                </div>
                <Progress
                  value={derivedPipelineProgress}
                  className="h-3"
                />
                <div className="mt-2 text-xs text-muted-foreground">
                  Stage completion {completedStages} / {resolvedTotalStages} ({derivedPipelineProgress}%)
                </div>
              </div>
            )}

            {/* Pipeline Summary Info */}
            <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
              {pipelineSummary?.total_duration && (
                <div className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  <span>{pipelineSummary.total_duration}</span>
                </div>
              )}

              {pipelineSummary?.total_files && (
                <div className="flex items-center gap-1">
                  <Activity className="h-3 w-3" />
                  <span>{pipelineSummary.total_files} files</span>
                </div>
              )}

              {pipelineSummary?.last_run && (
                <div className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  <span>Last run: {new Date(pipelineSummary.last_run).toLocaleDateString()}</span>
                </div>
              )}
            </div>

            {/* Status Message */}
            {operation.message && (
              <div className="mt-2 text-sm text-muted-foreground">
                {operation.message}
              </div>
            )}

            {/* Action Buttons */}
            <div className="mt-4 flex gap-2">
              <Button
                onClick={(e) => {
                  e.stopPropagation() // Prevent expanding when clicking buttons
                  if (canQuickStart && onDirectStart) {
                    onDirectStart()
                  } else {
                    onConfigure()
                  }
                }}
                disabled={primaryDisabled}
                size="default"
                className="flex-1 h-11"
              >
                {isStarting ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-1.5 animate-spin" />
                    Starting...
                  </>
                ) : canQuickStart ? (
                  <>
                    <Zap className="h-4 w-4 mr-1.5" />
                    Quick Start
                  </>
                ) : (
                  <>
                    <Settings2 className="h-4 w-4 mr-1.5" />
                    {primaryActionLabel || 'Configure & Start'}
                  </>
                )}
              </Button>
            </div>
          </CardHeader>
        </CollapsibleTrigger>

        {/* Detailed Stage Tickets - Collapsible Content */}
        <CollapsibleContent>
          <CardContent className="pt-0">
            <div className="border-t pt-4">
              {/* Stage Tickets Header */}
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-medium text-muted-foreground">
                  Individual Stages
                </h3>
                <div className="text-xs text-muted-foreground">
                  Click on any stage to configure or start individually
                </div>
              </div>

              {/* Render Stage Tickets */}
              <div className="space-y-3">
                {React.Children.toArray(stages).map((stage) => (
                  <div key={stage.key} className="border-l-2 border-muted pl-4">
                    {stage}
                  </div>
                ))}
              </div>
            </div>
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  )
}

export default PipelineSummaryCard
