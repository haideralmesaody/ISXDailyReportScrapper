/**
 * Operation Complete Card Component
 * Clear completion status with next step suggestions
 */

'use client'

import React from 'react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { 
  CheckCircle2,
  ArrowRight,
  Zap,
  Database,
  BarChart3,
  FileText,
  Download,
  Sparkles
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import { identifyStage, getStageInfo, getStageMetrics, extractStageFileCount, isStageReliable } from '@/lib/operations/stage-mapping'
import type { StageMappingResult } from '@/lib/operations/stage-identification'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { SegmentedFileProgress } from './SegmentedFileProgress'
import { ScrapingStageCard } from './ScrapingStageCard'

interface OperationCompleteCardProps {
  operation: {
    operation_id: string
    name?: string
    status?: string
    progress?: number
    message?: string
    metadata?: {
      files_processed?: number
      duration?: number
      operation_type?: string
      mode?: string
      final_mode?: string
      mode_reason?: string
      recommended_mode?: string
      stage_id?: string
      stage_name?: string
      stage_metrics?: Record<string, any>
      mode_detection?: {
        recommended_mode?: string
        reason?: string
        coverage_percent?: number
      }
      pipeline_summary?: any // Pipeline metadata with stage timeline
    }
    data?: {
      stage_id?: string
      stage_name?: string
      operation_type?: string
      [key: string]: any
    }
    completed_at?: string
    steps?: Array<{
      id: string
      name: string
      status: string
      progress: number
      metadata?: Record<string, any>
    }>
  }
  onNextOperation?: (type: string) => void
  className?: string
  showPipelineDetails?: boolean  // New: Show enhanced pipeline information
  pipelineMetadata?: any        // New: Access to full pipeline context
}

const NEXT_OPERATIONS = {
  scraping: [
    {
      id: 'processing',
      label: 'Process Data',
      description: 'Transform Excel files to CSV',
      icon: Database,
      variant: 'default' as const,
      recommended: true
    },
    {
      id: 'indices',
      label: 'Extract Indices',
      description: 'Get ISX60 & ISX15 data',
      icon: BarChart3,
      variant: 'outline' as const
    },
    {
      id: 'liquidity',
      label: 'Liquidity Analysis',
      description: 'Calculate liquidity metrics',
      icon: Zap,
      variant: 'ghost' as const
    }
  ],
  processing: [
    {
      id: 'indices',
      label: 'Extract Indices',
      description: 'Get market indices',
      icon: BarChart3,
      variant: 'default' as const,
      recommended: true
    },
    {
      id: 'liquidity',
      label: 'Liquidity Analysis',
      description: 'Calculate liquidity scores',
      icon: FileText,
      variant: 'outline' as const
    }
  ],
  indices: [
    {
      id: 'liquidity',
      label: 'Liquidity Analysis',
      description: 'Generate liquidity metrics',
      icon: FileText,
      variant: 'default' as const,
      recommended: true
    },
    {
      id: 'export',
      label: 'Export Data',
      description: 'Download results',
      icon: Download,
      variant: 'outline' as const
    }
  ],
  full_pipeline: [] // No next operations for completed full pipeline
} as const


const MODE_LABELS: Record<string, { label: string; description?: string }> = {
  initial: { label: 'Initial (fresh download)' },
  accumulative: { label: 'Accumulative (incremental)' },
  full: { label: 'Full Pipeline' },
  skip: { label: 'Skip (already up to date)', description: 'All requested trading days already exist' }
}

export function OperationCompleteCard({
  operation,
  onNextOperation,
  className,
  showPipelineDetails = false, // New
  pipelineMetadata = null      // New
}: OperationCompleteCardProps) {
  const stageMapping = React.useMemo<StageMappingResult>(() => {
    try {
      return identifyStage(operation)
    } catch (error) {
      console.warn('Stage identification failed:', error);
      return { stageId: 'scraping', confidence: 'low', source: 'fallback' } as const;
    }
  }, [operation])

  const stageInfo = React.useMemo(() => {
    try {
      return getStageInfo(stageMapping.stageId)
    } catch (error) {
      console.warn('Stage info retrieval failed:', error);
      return { name: 'Unknown Stage', description: 'Stage information unavailable' };
    }
  }, [stageMapping.stageId])

  const stageMetrics = React.useMemo(() => {
    try {
      return getStageMetrics(stageMapping.stageId)
    } catch (error) {
      console.warn('Stage metrics retrieval failed:', error);
      return { primaryMetric: 'files_processed', primaryLabel: 'Files Processed', secondaryMetric: 'duration', secondaryLabel: 'Duration' };
    }
  }, [stageMapping.stageId])

  const isReliable = React.useMemo(() => {
    try {
      return isStageReliable(stageMapping)
    } catch (error) {
      console.warn('Stage reliability check failed:', error);
      return false;
    }
  }, [stageMapping])

  const combinedMetadata = React.useMemo(() => {
    const merged: Record<string, any> = {}

    if (operation.metadata && typeof operation.metadata === 'object') {
      Object.assign(merged, operation.metadata)
    }

    if (stageMapping.stageId && operation.metadata?.stage_metrics) {
      const stageSpecific = operation.metadata.stage_metrics[stageMapping.stageId]
      if (stageSpecific && typeof stageSpecific === 'object') {
        Object.assign(merged, stageSpecific)
      }
    }

    if (Array.isArray(operation.steps) && operation.steps.length > 0) {
      const firstStepMetadata = operation.steps[0]?.metadata
      if (firstStepMetadata && typeof firstStepMetadata === 'object') {
        Object.assign(merged, firstStepMetadata)
      }
    }

    return merged
  }, [operation.steps, operation.metadata, stageMapping.stageId])

  const getStageDisplayName = (): string => {
    if (isReliable) {
      return stageInfo.name
    }
    return `${stageInfo.name} (Uncertain)`
  }

  const modeDetection = (combinedMetadata?.mode_detection as Record<string, any> | undefined) ??
    operation.metadata?.mode_detection
  const rawMode = modeDetection?.recommended_mode ??
    combinedMetadata?.final_mode ??
    combinedMetadata?.recommended_mode ??
    combinedMetadata?.mode
  const normalizedMode = typeof rawMode === 'string' ? rawMode.toLowerCase() : undefined
  const modeDetails = normalizedMode ? MODE_LABELS[normalizedMode] : undefined
  const modeReason = modeDetection?.reason ?? combinedMetadata?.mode_reason
  const coveragePercent = modeDetection?.coverage_percent ?? combinedMetadata?.coverage_percent

  // Extract trading date or last file information from message
  const getTradingInfo = (
    operation: any,
    metadata?: Record<string, any>
  ): { tradingDate?: string; lastFile?: string } => {
    const message = operation.message || ''

    // Look for "Last file: filename" pattern
    const lastFileMatch = message.match(/\| Last file:\s*([^\|]+)/)
    if (lastFileMatch) {
      return { lastFile: lastFileMatch[1].trim() }
    }

    // Look for "Trading date: YYYY-MM-DD" pattern
    const tradingDateMatch = message.match(/\| Trading date:\s*([^\|]+)/)
    if (tradingDateMatch) {
      return { tradingDate: tradingDateMatch[1].trim() }
    }

    const metadataSource = metadata || operation.metadata
    if (metadataSource?.trading_date) {
      return { tradingDate: metadataSource.trading_date }
    }

    if (metadataSource?.last_processed_file) {
      return { lastFile: metadataSource.last_processed_file }
    }

    return {}
  }

  const tradingInfo = React.useMemo(
    () => getTradingInfo(operation, combinedMetadata),
    [operation, combinedMetadata]
  )

  const stageDurationMs =
    typeof combinedMetadata?.duration === 'number' ? combinedMetadata.duration : undefined

  const nextOps = NEXT_OPERATIONS[stageMapping.stageId as keyof typeof NEXT_OPERATIONS] || []

  // Use robust stage-specific file count extraction
  const filesCount = React.useMemo(() => {
    try {
      return extractStageFileCount(operation, stageMapping.stageId)
    } catch (error) {
      console.warn('Stage file count extraction failed:', error)
      return 0
    }
  }, [operation, stageMapping.stageId])

  const progressContent = React.useMemo(() => {
    if (!combinedMetadata || typeof combinedMetadata !== 'object') {
      return null
    }

    if (stageMapping.stageId === 'scraping') {
      const fromDate = combinedMetadata.from_date || combinedMetadata.start_date
      const toDate = combinedMetadata.to_date || combinedMetadata.end_date

      if (fromDate && toDate) {
        return (
          <ScrapingStageCard
            metadata={combinedMetadata}
            statusMessage={operation.message}
            isComplete
            variant="complete"
            className="mt-2"
          />
        )
      }

      return null
    }

    const totalFiles =
      typeof combinedMetadata.total_files === 'number' && combinedMetadata.total_files > 0
        ? combinedMetadata.total_files
        : filesCount

    const processedFiles =
      typeof combinedMetadata.files_processed === 'number' && combinedMetadata.files_processed >= 0
        ? combinedMetadata.files_processed
        : filesCount

    const failedFiles =
      typeof combinedMetadata.failed_files === 'number'
        ? combinedMetadata.failed_files
        : typeof combinedMetadata.errors === 'number'
          ? combinedMetadata.errors
          : 0

    const fileList = Array.isArray(combinedMetadata.file_list)
      ? combinedMetadata.file_list as string[]
      : undefined

    const fileStatuses = Array.isArray(combinedMetadata.file_statuses)
      ? combinedMetadata.file_statuses as Array<{
          filename: string
          status: 'pending' | 'processing' | 'completed' | 'failed'
          size_mb?: number
          progress?: number
          error_message?: string
          processing_time_ms?: number
        }>
      : undefined

    const processingSpeedMBps =
      typeof combinedMetadata.processing_speed_mbps === 'number'
        ? combinedMetadata.processing_speed_mbps
        : undefined

    const totalSizeMB =
      typeof combinedMetadata.total_size_mb === 'number'
        ? combinedMetadata.total_size_mb
        : undefined

    const processedSizeMB =
      typeof combinedMetadata.processed_size_mb === 'number'
        ? combinedMetadata.processed_size_mb
        : undefined

    const estimatedRemainingMs =
      typeof combinedMetadata.estimated_remaining_ms === 'number'
        ? combinedMetadata.estimated_remaining_ms
        : undefined

    return (
      <SegmentedFileProgress
        totalFiles={Math.max(totalFiles, processedFiles, 1)}
        processedFiles={processedFiles}
        failedFiles={failedFiles}
        currentFile={combinedMetadata.current_file}
        fileList={fileList}
        fileStatuses={fileStatuses}
        processingSpeedMBps={processingSpeedMBps}
        totalSizeMB={totalSizeMB}
        processedSizeMB={processedSizeMB}
        estimatedRemainingMs={estimatedRemainingMs}
        className="mt-2"
        showDetails={false}
      />
    )
  }, [combinedMetadata, stageMapping.stageId, filesCount])
  
  return (
    <motion.div
      initial={{ opacity: 0, y: 20, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ 
        type: "spring",
        stiffness: 200,
        damping: 20
      }}
    >
      <Card className={cn(
        "relative overflow-hidden transition-all duration-300 hover:shadow-lg",
        "bg-gradient-to-br from-green-50/90 via-emerald-50/80 to-white/90 dark:from-green-950/90 dark:to-emerald-950/90",
        "border-green-200/50 dark:border-green-800/50 backdrop-blur-sm",
        "shadow-sm hover:shadow-md hover:border-green-300/70 dark:hover:border-green-700/70",
        className
      )}>
        {/* Enhanced success accent bar with animation */}
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-green-500 via-emerald-500 to-green-500">
          <motion.div
            className="h-full bg-white/30"
            initial={{ x: "-100%" }}
            animate={{ x: "100%" }}
            transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
          />
        </div>
        
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ 
                  type: "spring",
                  delay: 0.2
                }}
                className="p-2.5 rounded-full bg-green-100 dark:bg-green-900"
              >
                <CheckCircle2 className="h-6 w-6 text-green-700 dark:text-green-300" />
              </motion.div>
              <div>
                <h3 className="font-semibold text-lg text-green-900 dark:text-green-100">
                  Operation Complete!
                </h3>
                <p className="text-sm text-green-700 dark:text-green-300 mt-0.5">
                  {getStageDisplayName(operation)}
                </p>
              </div>
            </div>
            
            <Badge 
              variant="outline" 
              className="bg-green-100 text-green-800 border-green-300 dark:bg-green-900 dark:text-green-100 dark:border-green-700"
            >
              <Sparkles className="h-3 w-3 mr-1" />
              Success
            </Badge>
          </div>
        </CardHeader>
        
        <CardContent className="space-y-3 p-3">
          {/* Enhanced compact success summary */}
          <div className="bg-white/70 dark:bg-gray-900/40 rounded-lg p-4 backdrop-blur-sm border border-green-100/30 dark:border-green-800/30">
            <div className="grid grid-cols-2 gap-3 text-sm">
              {/* Stage name with confidence indicator */}
              <div className="space-y-1">
                <span className="text-muted-foreground text-xs">Stage</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <div className="flex items-center gap-1 cursor-help">
                        <span className="font-semibold text-green-700 dark:text-green-300">
                          {getStageDisplayName()}
                        </span>
                        {!isReliable && (
                          <div className="w-2 h-2 rounded-full bg-yellow-500 animate-pulse" />
                        )}
                        {isReliable && (
                          <div className="w-2 h-2 rounded-full bg-green-500" />
                        )}
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <div className="text-xs space-y-1">
                        <div className="font-medium">
                          {isReliable ? 'High Confidence' : 'Low Confidence'}
                        </div>
                        <div className="text-muted-foreground">
                          Source: {stageMapping.source}
                        </div>
                        <div className="text-muted-foreground">
                          Stage ID: {stageMapping.stageId}
                        </div>
                        {stageMapping.metadata?.matchedPattern && (
                          <div className="text-muted-foreground">
                            Pattern: {stageMapping.metadata.matchedPattern}
                          </div>
                        )}
                      </div>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>

              {/* Files processed */}
              <div className="space-y-1">
                <span className="text-muted-foreground text-xs">{stageMetrics.primaryLabel}</span>
                <span className="font-semibold text-green-700 dark:text-green-300">
                  {filesCount} {filesCount === 1 ? 'file' : 'files'}
                </span>
              </div>

              {/* Mode (if available) */}
              {normalizedMode && (
                <div className="space-y-1">
                  <span className="text-muted-foreground text-xs">Mode</span>
                  <span className="font-medium text-green-700 dark:text-green-300 text-xs">
                    {modeDetails?.label ?? normalizedMode}
                  </span>
                </div>
              )}

              {/* Trading date or last file */}
              {(tradingInfo.tradingDate || tradingInfo.lastFile) && (
                <div className="space-y-1">
                  <span className="text-muted-foreground text-xs">
                    {tradingInfo.tradingDate ? 'Trading Date' : 'Last File'}
                  </span>
                  <span className="font-medium text-blue-700 dark:text-blue-300 text-xs truncate">
                    {tradingInfo.tradingDate || tradingInfo.lastFile}
                  </span>
                </div>
              )}

              {/* Duration */}
              {typeof stageDurationMs === 'number' && (
                <div className="space-y-1">
                  <span className="text-muted-foreground text-xs">Duration</span>
                  <span className="font-medium text-xs">
                    {Math.round(stageDurationMs / 1000)}s
                  </span>
                </div>
              )}

              {/* Completion time */}
              {operation.completed_at && (
                <div className="space-y-1">
                  <span className="text-muted-foreground text-xs">Completed</span>
                  <span className="font-medium text-xs">
                    {new Date(operation.completed_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                  </span>
                </div>
              )}
            </div>

            {progressContent && (
              <div className="mt-4">
                {progressContent}
              </div>
            )}

            {/* Mode details tooltip (only shown if interesting details exist) */}
            {normalizedMode && (modeDetails?.description || modeReason || typeof coveragePercent === 'number') && (
              <div className="mt-2 pt-2 border-t border-green-200/30 dark:border-green-700/30">
                <div className="text-xs text-muted-foreground space-y-0.5">
                  {modeDetails?.description && (
                    <div>{modeDetails.description}</div>
                  )}
                  {modeReason && (
                    <div>Reason: {modeReason}</div>
                  )}
                  {typeof coveragePercent === 'number' && normalizedMode === 'skip' && (
                    <div>Coverage: {coveragePercent.toFixed(1)}%</div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Enhanced Pipeline Performance Details */}
          {showPipelineDetails && (pipelineMetadata || operation.metadata?.pipeline_summary) && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.3 }}
              className="bg-blue-50/80 dark:bg-blue-950/40 rounded-lg p-4 backdrop-blur-sm border border-blue-100/30 dark:border-blue-800/30"
            >
              <div className="flex items-center gap-2 mb-3">
                <div className="w-2 h-2 rounded-full bg-blue-500" />
                <h4 className="font-semibold text-blue-800 dark:text-blue-200 text-sm">Pipeline Performance</h4>
              </div>

              {/* Stage Timeline Summary */}
              {(pipelineMetadata || operation.metadata?.pipeline_summary)?.stage_timeline && (
                <div className="mb-3">
                  <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-xs">
                    {(pipelineMetadata || operation.metadata?.pipeline_summary).stage_timeline.map((stage: any, index: number) => (
                      <div key={stage.stage_id || index} className="text-center">
                        <div className="font-medium text-blue-700 dark:text-blue-300 capitalize">
                          {stage.stage_name || stage.stage_id}
                        </div>
                        <div className="text-blue-600 dark:text-blue-400">
                          {stage.duration_ms ? `${(stage.duration_ms / 1000).toFixed(1)}s` : 'N/A'}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Overall Pipeline Metrics */}
              <div className="grid grid-cols-2 gap-3 text-xs">
                <div className="space-y-1">
                  <span className="text-blue-600 dark:text-blue-400 font-medium">Total Duration:</span>
                  <p className="text-blue-800 dark:text-blue-200">
                    {(pipelineMetadata || operation.metadata?.pipeline_summary)?.total_duration
                      ? `${((pipelineMetadata || operation.metadata?.pipeline_summary).total_duration / 1000).toFixed(1)}s`
                      : 'N/A'}
                  </p>
                </div>
                <div className="space-y-1">
                  <span className="text-blue-600 dark:text-blue-400 font-medium">Stages Completed:</span>
                  <p className="text-blue-800 dark:text-blue-200">
                    {Array.isArray(operation.steps) ? operation.steps.length : 'N/A'}
                  </p>
                </div>
              </div>

              {/* Stage-specific timing for current stage */}
              {operation.metadata?.stage_timeline && (
                <div className="mt-3 pt-3 border-t border-blue-200/30 dark:border-blue-700/30">
                  <div className="text-xs text-blue-700 dark:text-blue-300">
                    <span className="font-medium">Current Stage Duration: </span>
                    {(() => {
                      const currentStageTiming = operation.metadata.stage_timeline.find(
                        (stage: any) => stage.stage_id === stageMapping.stageId
                      )
                      return currentStageTiming?.duration_ms
                        ? `${(currentStageTiming.duration_ms / 1000).toFixed(1)}s`
                        : 'N/A'
                    })()}
                  </div>
                </div>
              )}
            </motion.div>
          )}

          {/* Compact next steps */}
          {onNextOperation && nextOps.length > 0 && (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <ArrowRight className="h-3 w-3 text-green-600" />
                <p className="text-xs font-medium text-green-800 dark:text-green-200">
                  Next Steps
                </p>
              </div>

              <div className="grid gap-1.5">
                {nextOps.map((op, index) => {
                  const Icon = op.icon
                  return (
                    <motion.div
                      key={op.id}
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: 0.2 + index * 0.05 }}
                    >
                      <Button
                        variant={op.variant}
                        size="sm"
                        className={cn(
                          "w-full justify-start h-8 px-3 text-xs",
                          op.recommended && "ring-1 ring-green-500/30"
                        )}
                        onClick={() => onNextOperation(op.id)}
                      >
                        <Icon className="h-3 w-3 mr-2 flex-shrink-0" />
                        <span className="flex-1 truncate">{op.label}</span>
                        {op.recommended && (
                          <Badge variant="secondary" className="text-[10px] px-1 py-0 h-4">
                            â˜…
                          </Badge>
                        )}
                      </Button>
                    </motion.div>
                  )
                })}
              </div>
            </div>
          )}
          
          {/* Alternative: Compact success message if no next operations */}
          {(!onNextOperation || nextOps.length === 0) && (
            <div className="text-center py-1">
              <p className="text-xs text-green-700 dark:text-green-300 font-medium">
                {stageMapping.stageId === 'full_pipeline'
                  ? 'All pipeline stages completed successfully!'
                  : 'Operation completed successfully!'}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </motion.div>
  )
}

export default OperationCompleteCard

