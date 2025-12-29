/**
 * Unified Operation Progress Component
 * Clean, single source of truth for operation status display
 */

'use client'

import React, { useMemo, useCallback, useEffect } from 'react'
import { Card, Button, Separator, Alert, AlertDescription } from '@/components/ui/safe-components'
import { CardContent, CardHeader } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import {
  CheckCircle2,
  Clock,
  Download,
  FileSearch,

  AlertCircle,
  ArrowRight,
  Package,
  FileText,
  Zap,
  BarChart3
} from 'lucide-react'
import { cn } from '@/lib/utils'

import { ScrapingStageCard } from './ScrapingStageCard'
import { ProcessingStageCard } from './ProcessingStageCard'

import { IndicesStageCard } from './IndicesStageCard'
import { LiquidityStageCard } from './LiquidityStageCard'
import { identifyStage, getStageInfo, getStageOrder, getTotalStages } from '@/lib/operations/stage-mapping'
import { useStageActivation, useStageTiming } from '@/lib/hooks/use-stage-activation'
import { PipelineConnector } from './PipelineConnector'
import { getOperationType, getOperationThresholds, getPhaseForThreshold, getPhaseDisplay, getPhaseBadgeVariant } from '@/lib/operations/phase-helpers'

const safeString = (value: any): string => {
  if (value === null || value === undefined) return ''
  if (typeof value === 'string') return value
  if (typeof value === 'number' && !Number.isNaN(value)) return value.toString()
  if (typeof value === 'boolean') return value.toString()
  if (Array.isArray(value)) return value.length.toString()
  return String(value)
}

const formatNumber = (value: any): string => {
  const num = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(num)) {
    return safeString(value)
  }
  const absNum = Math.abs(num)
  if (absNum >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`
  if (absNum >= 1_000) return `${(num / 1_000).toFixed(1)}K`
  if (Math.round(num) !== num) return num.toFixed(1)
  return num.toString()
}

const toNumber = (value: any): number | undefined => {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

// Helper function to get stage number
const getStageNumber = (operationType: string): number => {
  if (!operationType || typeof operationType !== 'string') return 1 // Default to 1 for safety
  const stageOrder = ['scraping', 'processing', 'indices', 'liquidity', 'indicators', 'analysis']
  const index = stageOrder.indexOf(operationType)
  return index >= 0 ? index + 1 : 1 // Default to 1 if not found
}

// Operation phases with clear progression
const OPERATION_PHASES = {
  scraping: [
    { id: 'preparing', label: 'Preparing', icon: Clock, progress: 0 },
    { id: 'scanning', label: 'Scanning', icon: FileSearch, progress: 25 },
    { id: 'downloading', label: 'Collecting', icon: Download, progress: 50 },
    { id: 'verifying', label: 'Verifying', icon: Package, progress: 75 },
    { id: 'complete', label: 'Complete', icon: CheckCircle2, progress: 100 }
  ],
  processing: [
    { id: 'preparing', label: 'Initializing Data Processor', icon: Clock, progress: 0 },
    { id: 'reading', label: 'Loading Excel Files', icon: FileText, progress: 25 },
    { id: 'transforming', label: 'Converting to CSV Format', icon: Zap, progress: 50 },
    { id: 'writing', label: 'Generating Data Reports', icon: FileText, progress: 75 },
    { id: 'complete', label: 'Data Processing Complete', icon: CheckCircle2, progress: 100 }
  ],
  indices: [
    { id: 'preparing', label: 'Preparing', icon: Clock, progress: 0 },
    { id: 'extracting', label: 'Extracting Indices', icon: FileSearch, progress: 50 },
    { id: 'complete', label: 'Complete', icon: CheckCircle2, progress: 100 }
  ],
  liquidity: [
    { id: 'preparing', label: 'Initializing', icon: Clock, progress: 0 },
    { id: 'reading', label: 'Loading Data', icon: FileText, progress: 20 },
    { id: 'calculating', label: 'Calculating Scores', icon: Zap, progress: 40 },
    { id: 'scaling', label: 'Cross-sectional Scaling', icon: Package, progress: 70 },
    { id: 'writing', label: 'Writing Results', icon: FileText, progress: 90 },
    { id: 'complete', label: 'Complete', icon: CheckCircle2, progress: 100 }
  ],
  analysis: [
    { id: 'preparing', label: 'Preparing Analysis', icon: Clock, progress: 0 },
    { id: 'loading', label: 'Loading Market Data', icon: FileText, progress: 25 },
    { id: 'analyzing', label: 'Running Technical Analysis', icon: BarChart3, progress: 50 },
    { id: 'generating', label: 'Generating Insights', icon: Zap, progress: 75 },
    { id: 'complete', label: 'Analysis Complete', icon: CheckCircle2, progress: 100 }
  ],
  indicators: [
    { id: 'preparing', label: 'Initializing', icon: Clock, progress: 0 },
    { id: 'loading', label: 'Loading Price Data', icon: FileText, progress: 20 },
    { id: 'calculating', label: 'Calculating Indicators', icon: BarChart3, progress: 40 },
    { id: 'caching', label: 'Building SSOT Cache', icon: Package, progress: 70 },
    { id: 'validating', label: 'Validating Results', icon: FileSearch, progress: 90 },
    { id: 'complete', label: 'Indicators Ready', icon: CheckCircle2, progress: 100 }
  ]
} as const

interface UnifiedOperationProgressProps {
  operation: {
    operation_id: string
    name?: string
    status: string
    progress: number
    message?: string
    error?: string
    metadata?: {
      phase?: string
      files_processed?: number
      total_files?: number
      current_file?: string
      speed?: number
      from_date?: string
      to_date?: string
      downloaded_files?: string[]
      pipelineMetadata?: any // Parent pipeline metadata
      stage_timeline?: any
    }
    steps?: Array<{
      id: string
      name: string
      status: string
      progress: number
      metadata?: {
        phase?: string
        files_processed?: number
        total_files?: number
        current_file?: string
        speed?: number
        from_date?: string
        to_date?: string
        downloaded_files?: string[]
        stage_id?: string
        stage_timeline?: any
      }
    }>
    stageContext?: {
      allStages: any[]
      pipelineMetadata: any
      completedPipeline: boolean
    }
    pipelineMessage?: string
    stepData?: any
    updated_at?: string
  }
  onNextOperation?: (type: string) => void
  pipelineMode?: boolean // When true, shows enhanced pipeline features
  stageNumber?: number    // Stage number in pipeline (1-6)
  totalStages?: number    // Total number of stages in pipeline
  showConnector?: boolean // Show visual connector to next stage
  completedPipeline?: boolean // New: Show as part of completed pipeline
  pipelineMetadata?: any      // New: Access to full pipeline context
  stageTimeline?: any         // New: Stage timing data
  allPipelineStages?: any[] // All stages in the pipeline for activation logic
}

export function UnifiedOperationProgress({
  operation,
  onNextOperation,
  pipelineMode = false,
  stageNumber,
  totalStages = 6,
  showConnector = false,
  completedPipeline = false, // New
  pipelineMetadata = null,   // New
  stageTimeline = null,      // New
  allPipelineStages = []
}: UnifiedOperationProgressProps) {


  // Use robust stage mapping utility
  const stageMapping = useMemo(() => {
    try {
      return identifyStage(operation)
    } catch (error) {
      console.warn('Stage identification failed in UnifiedOperationProgress:', error)
      return { stageId: 'scraping', confidence: 'low', source: 'fallback' }
    }
  }, [operation])

  const operationType = stageMapping.stageId || 'scraping' // Default to scraping if stageId is missing

  const stepData = operation.stepData || (Array.isArray(operation.steps) ? operation.steps[0] : null)
  const stageStatus = stepData?.status || operation.status || 'pending'

  const isComplete = stageStatus === 'completed'
  const isFailed = stageStatus === 'failed'

  // Define stageProgressPercent early to prevent TDZ errors
  const stageProgressPercent = typeof operation.stepData?.metadata?.progress_percent === 'number'
    ? Math.max(0, Math.min(100, operation.stepData.metadata.progress_percent))
    : undefined

  const stepMetadata = useMemo(() => {
    try {
      const metadata: Record<string, any> = { ...(stepData?.metadata || {}) }
      const stageId =
        metadata.stage_id ||
        stepData?.id ||
        stepData?.stage_id ||
        stageMapping.stageId ||
        operationType

      metadata.stage_id = stageId
      metadata.stage_alias = metadata.stage_alias || metadata.target_stage_id || stageId

      if (typeof metadata.phase === 'string') {
        metadata.phase = metadata.phase.toLowerCase()
        metadata.current_phase = metadata.phase
      } else if (typeof metadata.current_phase === 'string') {
        metadata.phase = metadata.current_phase.toLowerCase()
        metadata.current_phase = metadata.phase
      } else {
        const defaultPhase =
          stageStatus === 'running'
            ? 'running'
            : stageStatus === 'completed'
              ? 'complete'
              : stageStatus === 'failed'
                ? 'failed'
                : 'pending'
        metadata.phase = defaultPhase
        metadata.current_phase = defaultPhase
      }

      if (!metadata.phase_message) {
        metadata.phase_message =
          metadata.stage_message ||
          (stageStatus === 'running'
            ? 'Stage running'
            : stageStatus === 'completed'
              ? 'Stage completed'
              : stageStatus === 'failed'
                ? 'Stage failed'
                : 'Awaiting start')
      }

      if (!metadata.stage_message) {
        metadata.stage_message = metadata.phase_message
      }

      if (typeof metadata.progress_percent !== 'number') {
        metadata.progress_percent =
          typeof stepData?.progress === 'number'
            ? stepData.progress
            : typeof operation.progress === 'number'
              ? operation.progress
              : 0
        metadata.telemetry_missing = true
      }

      if (typeof metadata.files_processed !== 'number') {
        metadata.files_processed =
          typeof metadata.processed_files === 'number' ? metadata.processed_files : 0
      }

      if (typeof metadata.total_files !== 'number') {
        metadata.total_files =
          typeof metadata.expected_files === 'number' ? metadata.expected_files : 0
      }

      if (!metadata.stage_started_at) {
        metadata.stage_started_at = metadata.started_at || operation.start_time || null
      }

      if (!metadata.stage_updated_at) {
        metadata.stage_updated_at = metadata.updated_at || operation.updated_at || null
      }

      return metadata
    } catch (error) {
      console.warn('Error processing step metadata:', error)
      return operation.metadata || {}
    }
  }, [stepData, stageMapping.stageId, operationType, stageStatus, operation.progress, operation.start_time, operation.updated_at, operation.metadata, stageProgressPercent])

  const stageInfo = useMemo(() => {
    try {
      return getStageInfo(operationType)
    } catch (error) {
      console.warn('Stage info failed in UnifiedOperationProgress:', error)
      return { name: 'Unknown Stage', description: 'Stage information unavailable', order: 1 }
    }
  }, [operationType])
  const stageContext = operation.stageContext
  const resolvedAllStages =
    stageContext?.allStages && Array.isArray(stageContext.allStages) ? stageContext.allStages : allPipelineStages
  const resolvedTotalStages = resolvedAllStages?.length || totalStages
  const resolvedCompletedPipeline = stageContext?.completedPipeline ?? completedPipeline
  const resolvedPipelineMetadata = stageContext?.pipelineMetadata ?? pipelineMetadata
  const stageProgress = useMemo(() => {
    if (isComplete) {
      return 100
    }

    if (typeof stepMetadata?.progress_percent === 'number') {
      return stepMetadata.progress_percent
    }

    if (typeof stepData?.progress === 'number') {
      return stepData.progress
    }

    return 0
  }, [isComplete, stepMetadata?.progress_percent, stepData?.progress])

  const scrapingMetrics = useMemo(() => {
    if (operationType !== 'scraping') return null

    const tradingDaysTotal = toNumber(stepMetadata?.trading_days_total)
    const tradingDaysCompleted = toNumber(stepMetadata?.trading_days_completed)
    const holidaysDetected = toNumber(stepMetadata?.holidays_detected)
    const filesDownloaded = Array.isArray(stepMetadata?.downloaded_files)
      ? stepMetadata.downloaded_files.length
      : toNumber(stepMetadata?.files_downloaded)
    const tradingDaysRemaining = toNumber(stepMetadata?.trading_days_remaining)
    const progressPercent = toNumber(stepMetadata?.progress_percent)

    const errors: string[] = []
    if (tradingDaysTotal === undefined || tradingDaysTotal <= 0) errors.push('trading_days_total missing or invalid')
    if (tradingDaysCompleted === undefined) errors.push('trading_days_completed missing')
    if (holidaysDetected === undefined) errors.push('holidays_detected missing')
    if (filesDownloaded === undefined) errors.push('downloaded_files missing')

    return {
      tradingDaysTotal,
      tradingDaysCompleted,
      tradingDaysRemaining,
      holidaysDetected,
      filesDownloaded,
      progressPercent,
      errors
    }
  }, [operationType, stepMetadata])

  const telemetryMissing = Boolean(
    !stepMetadata?.stage_skipped &&
    (stepMetadata?.telemetry_missing ||
      typeof stepMetadata?.progress_percent !== 'number')
  )

  const badgeLabel = stepMetadata?.stage_skipped
    ? 'Skipped'
    : stageStatus === 'completed'
      ? 'Completed'
      : stageStatus === 'failed'
        ? 'Failed'
        : stageStatus === 'running'
          ? 'Running'
          : 'Pending'

  const badgeVariant: 'default' | 'secondary' | 'destructive' | 'outline' =
    stepMetadata?.stage_skipped
      ? 'outline'
      : stageStatus === 'completed'
        ? 'secondary'
        : stageStatus === 'failed'
          ? 'destructive'
          : stageStatus === 'running'
            ? 'default'
            : 'secondary'

  const stageIndicatorLabel = stepMetadata?.stage_skipped
    ? 'Stage skipped'
    : stageStatus === 'running'
      ? 'Stage running'
      : stageStatus === 'failed'
        ? 'Stage failed'
        : 'Stage ready'

  const stageIndicatorColor = stepMetadata?.stage_skipped
    ? 'bg-gray-400'
    : stageStatus === 'running'
      ? 'bg-green-500'
      : stageStatus === 'failed'
        ? 'bg-red-500'
        : 'bg-gray-400'

  // Log data source usage for debugging progress calculation
  useEffect(() => {
    if (process.env.NODE_ENV !== 'development' || typeof window === 'undefined') {
      return
    }

    if (window.localStorage?.getItem('debug_pipeline_progress') !== 'true') {
      return
    }

    const operationId = operation.operation_id || 'unknown'
    const stageType = stepMetadata?.stage_id || stepData?.id || stepData?.stage_id || 'unknown'

    let dataSource = 'default_zero'
    let progressValue = 0

    if (typeof stepMetadata?.progress_percent === 'number') {
      dataSource = 'step_metadata_progress_percent'
      progressValue = stepMetadata.progress_percent
    } else if (typeof stepData?.progress === 'number') {
      dataSource = 'step_progress'
      progressValue = stepData.progress
    }

    console.debug(`[UnifiedOperationProgress] Operation ${operationId} (${stageType}) -> ${progressValue}% via ${dataSource}`)
  }, [
    operation.operation_id,
    stepData?.id,
    stepData?.stage_id,
    stepData?.progress,
    stepMetadata?.stage_id,
    stepMetadata?.progress_percent
  ])

  // Use centralized activation hook for consistent behavior
  const activationState = useStageActivation({
    stageData: operation,
    allStages: resolvedAllStages,
    pipelineCompleted: resolvedCompletedPipeline,
    stageNumber,
    totalStages: resolvedTotalStages
  })

  // Separate action state from display messaging
  const actionDisabled = !activationState.canActivate
  const actionDisplayText = activationState.isActive ? 'Running' : 'Ready for next step'
  const canActivateStage = activationState.canActivate
  const isStageActive = activationState.isActive

  // Use timing hook for stage performance data
  const timingInfo = useStageTiming(operation, resolvedPipelineMetadata)

  const phases = OPERATION_PHASES[operationType as keyof typeof OPERATION_PHASES] || OPERATION_PHASES.scraping
  const backendPhase = stepMetadata?.phase as string | undefined
  const backendPhaseDisplay = useMemo(() => {
    try {
      return backendPhase ? getPhaseDisplay(operationType, backendPhase) : null
    } catch (error) {
      console.warn('Backend phase display calculation failed:', error);
      return null;
    }
  }, [backendPhase, operationType])

  // Enhanced KPI display function with structured output and string conversion
  const getStageKPIDisplay = useCallback((data: any) => {
    const stepId = data?.step_id || data?.id || operationType

    const awaiting = { primary: 'Awaiting', secondary: 'metrics' }

    switch (stepId) {
      case 'scraping':
      case 'scrape':
        {
          const files = scrapingMetrics?.filesDownloaded ??
            (Array.isArray(data?.metadata?.downloaded_files)
              ? data.metadata.downloaded_files.length
              : toNumber(data?.metadata?.files_downloaded))
          const tradingTotal =
            scrapingMetrics?.tradingDaysTotal ??
            toNumber(data?.metadata?.trading_days_total) ??
            toNumber(data?.metadata?.total_files) ??
            toNumber(data?.metadata?.expected_files)
          const holidays = scrapingMetrics?.holidaysDetected ?? toNumber(data?.metadata?.holidays_detected)
          const completed = scrapingMetrics?.tradingDaysCompleted ?? toNumber(data?.metadata?.trading_days_completed)

          if (scrapingMetrics?.errors?.length) {
            return { primary: 'Telemetry error', secondary: scrapingMetrics.errors[0] }
          }

          if (tradingTotal !== undefined && (typeof completed === 'number' || typeof files === 'number')) {
            const base = typeof completed === 'number' ? completed : (files ?? 0)
            const numerator = base + (holidays ?? 0)
            return { primary: `${formatNumber(numerator)}/${formatNumber(tradingTotal)}`, secondary: 'days' }
          }
          return files !== undefined ? { primary: formatNumber(files), secondary: 'files' } : awaiting
        }
      case 'processing':
      case 'process':
        {
          const processed = data?.metadata?.files_processed
          const total = data?.metadata?.total_files
          if (processed && total) {
            return { primary: `${formatNumber(processed)}/${formatNumber(total)}`, secondary: 'files' }
          }
          return processed ? { primary: formatNumber(processed), secondary: 'files' } : awaiting
        }
      case 'indices':
      case 'index':
        {
          const indices = data?.metadata?.indices_extracted
          const records = data?.metadata?.records_processed
          if (records) {
            return { primary: formatNumber(records), secondary: 'records' }
          }
          return indices ? { primary: formatNumber(indices), secondary: 'indices' } : awaiting
        }
      case 'liquidity':
        {
          const tickers = data?.metadata?.tickers_analyzed ?? data?.metadata?.metrics_calculated
          if (tickers) {
            return { primary: formatNumber(tickers), secondary: 'tickers' }
          }
          return awaiting
        }
      case 'analysis':
        {
          const priceAlerts = data?.metadata?.price_alerts
          const rsiAlerts = data?.metadata?.rsi_alerts
          if (priceAlerts) {
            return { primary: formatNumber(priceAlerts), secondary: 'price alerts' }
          }
          if (rsiAlerts) {
            return { primary: formatNumber(rsiAlerts), secondary: 'RSI alerts' }
          }
          const traded = data?.metadata?.traded_stocks
          return traded ? { primary: formatNumber(traded), secondary: 'stocks' } : awaiting
        }
      case 'indicators':
        {
          const indicators = data?.metadata?.indicators_calculated
          const tickers = data?.metadata?.tickers_processed
          const calculations = data?.metadata?.calculations_performed

          if (calculations) {
            return { primary: formatNumber(calculations), secondary: 'calculations' }
          }
          if (tickers) {
            return { primary: formatNumber(tickers), secondary: 'tickers' }
          }
          return indicators ? { primary: formatNumber(indicators), secondary: 'indicators' } : awaiting
        }
      default:
        return awaiting
    }
  }, [operationType])

  const stageKPI = useMemo(() => getStageKPIDisplay(stepData), [getStageKPIDisplay, stepData])

  // Memoized operation type inference to prevent TDZ errors (moved before currentPhase)
  const inferredType = useMemo(() => {
    try {
      return getOperationType(operation.current_step || operationType)
    } catch (error) {
      console.warn('Operation type inference failed:', error);
      return operationType;
    }
  }, [operation.current_step, operationType])

  // Memoized thresholds calculation to prevent TDZ errors (moved before currentPhase)
  const thresholds = useMemo(() => {
    try {
      return getOperationThresholds(inferredType)
    } catch (error) {
      console.warn('Operation thresholds calculation failed:', error);
      return [];
    }
  }, [inferredType])

  // Memoized threshold calculation to prevent TDZ errors (moved before currentPhase)
  const getPhaseIdForThreshold = useCallback((thresholdIndex: number) => {
    try {
      return getPhaseForThreshold(inferredType, thresholdIndex)
    } catch (error) {
      console.warn('Phase ID calculation failed:', error);
      return null;
    }
  }, [inferredType])

  // Determine current phase
  const currentPhase = useMemo(() => {
    try {
      if (isComplete && phases && phases.length > 0) {
        return phases[phases.length - 1] || phases[0]
      }

      const metadataPhase = stepMetadata?.phase
      if (metadataPhase && phases) {
        const found = phases.find(p => p.id === metadataPhase)
        if (found) {
          return found
        }
      }

      if (thresholds && thresholds.length > 0) {
        for (let i = thresholds.length - 1; i >= 0; i--) {
          if (stageProgress >= thresholds[i]) {
            const phaseId = getPhaseIdForThreshold(i)
            if (phaseId && phases) {
              const matched = phases.find(p => p.id === phaseId)
              if (matched) {
                return matched
              }
            }
          }
        }
      }

      return phases?.[0] || { id: 'processing', name: 'Processing', icon: Clock, progress: 0 }
    } catch (error) {
      console.warn('Current phase calculation failed:', error)
      return { id: 'processing', name: 'Processing', icon: Clock, progress: 0 }
    }
  }, [isComplete, stepMetadata?.phase, phases, stageProgress, thresholds, getPhaseIdForThreshold])

  const CurrentIcon = currentPhase.icon

  // Generate descriptive operation title
  const operationTitle = useMemo(() => {
    if (operation.name && operation.name !== 'Operation') {
      return operation.name
    }

    // Use centralized stage mapping for consistent naming
    const stageDisplayName = stageInfo.name

    // Generate descriptive title based on operation type and metadata
    switch (operationType) {
      case 'scraping':
        if (stepMetadata?.from_date && stepMetadata?.to_date) {
          return `${stageDisplayName}: ${stepMetadata.from_date} to ${stepMetadata.to_date}`
        }
        return stageDisplayName
      case 'processing':
        return stageDisplayName
      case 'indices':
        return stageDisplayName
      case 'liquidity':
        return stageDisplayName
      case 'analysis':
        return stageDisplayName
      case 'indicators':
        return stageDisplayName
      default:
        return stageDisplayName
    }
  }, [operation.name, operationType, stepMetadata, stageInfo.name])

  // Generate status message
  const phaseDescription = stepMetadata?.phase_message || backendPhaseDisplay?.description

  // Clamp/display-safe files processed for scraping to avoid inflated values
  const filesProcessedDisplay = useMemo(() => {
    const raw = toNumber(stepMetadata?.files_processed) ?? 0
    if (stepMetadata?.stage_id === 'scraping' || operationType === 'scraping') {
      const completed = toNumber(stepMetadata?.trading_days_completed)
      return completed ?? raw
    }
    return raw
  }, [stepMetadata, operationType])

  const statusMessage = useMemo(() => {
    if (stepMetadata?.stage_skipped) {
      return 'Stage skipped – not executed in this run.'
    }

    if (stageStatus === 'failed') {
      return stepData?.error_message || operation.error || 'Operation failed. Please try again.'
    }

    if (stageStatus === 'completed') {
      return stepData?.stage_message || 'Stage completed'
    }

    return (
      phaseDescription ||
      stepData?.stage_message ||
      stepData?.metadata?.stage_message ||
      operation.metadata?.stage_message ||
      stepData?.metadata?.phase_message ||
      (operation.stageContext ? operation.pipelineMessage : operation.message) ||
      null
    )
  }, [stageStatus, stepData, operation.metadata, operation.pipelineMessage, operation.stageContext, operation.error, operation.message, phaseDescription])

  const timingDisplay = useMemo(() => {
    if (!timingInfo) {
      return null
    }

    const lastUpdate = operation.updated_at || stepData?.metadata?.updated_at || stepData?.updated_at
    const startedAt = timingInfo.startedAt ? new Date(timingInfo.startedAt).toLocaleString() : null
    const completedAt = timingInfo.completedAt ? new Date(timingInfo.completedAt).toLocaleString() : null
    const duration = timingInfo.durationFormatted && timingInfo.durationFormatted !== 'N/A'
      ? timingInfo.durationFormatted
      : null

    let formattedLastUpdate: string | null = null
    if (lastUpdate) {
      try {
        formattedLastUpdate = new Date(lastUpdate).toLocaleString()
      } catch {
        formattedLastUpdate = null
      }
    }

    if (!startedAt && !duration && !formattedLastUpdate && !completedAt) {
      return null
    }

    return (
      <div className="text-xs text-muted-foreground space-y-1">
        {startedAt && <div>Started {startedAt}</div>}
        {duration && <div>Duration {duration}</div>}
        {completedAt && <div>Completed {completedAt}</div>}
        {formattedLastUpdate && <div>Last updated {formattedLastUpdate}</div>}
      </div>
    )
  }, [timingInfo, operation.updated_at, stepData])

  // Memoized stage display info to prevent TDZ errors in production
  const stageDisplayInfo = useMemo(() => {
    try {
      const order = getStageOrder(operationType);
      const total = getTotalStages();
      return { order, total };
    } catch (error) {
      console.warn('Stage info calculation failed:', error);
      return { order: stageNumber || 1, total: totalStages || 6 };
    }
  }, [operationType, stageNumber, totalStages])

  // Calculate holiday-aware progress for scraping operations
  const calculateHolidayAwareProgress = useCallback(() => {
    // Prefer backend-provided progress_percent when available
    if (typeof stageProgressPercent === 'number') return stageProgressPercent
    if (operationType !== 'scraping') return stageProgress

    if (scrapingMetrics?.progressPercent !== undefined) {
      return Math.min(Math.max(scrapingMetrics.progressPercent, 0), 100)
    }

    return stageProgress
  }, [operationType, stageProgressPercent, stageProgress, scrapingMetrics])

  const visualProgress = useMemo(() => {
    const holidayAwareProgress = calculateHolidayAwareProgress()

    if (typeof stageProgressPercent === 'number') return stageProgressPercent
    if (isComplete) return 100
    if (typeof holidayAwareProgress === 'number') return Math.min(holidayAwareProgress, 100)
    return currentPhase.progress
  }, [stageProgressPercent, isComplete, calculateHolidayAwareProgress, currentPhase])

  return (
    <>
      <Card
        data-testid={stepMetadata?.stage_id ? `stage-card-${stepMetadata.stage_id}` : undefined}
        className={cn(
          "w-full overflow-hidden transition-all duration-300",
          isComplete && "ring-2 ring-green-500/20",
          isFailed && "ring-2 ring-red-500/20",
          pipelineMode && "relative",
          pipelineMode && isStageActive && !isComplete && !isFailed && "ring-2 ring-primary/40 bg-primary/5 shadow-lg"
        )}
      >
        {/* Status bar at top */}
        <div className={cn(
          "h-1 w-full",
          isComplete && "bg-gradient-to-r from-green-500 to-green-400",
          isFailed && "bg-gradient-to-r from-red-500 to-red-400",
          !isComplete && !isFailed && "bg-gradient-to-r from-blue-500 to-blue-400"
        )} />

        {/* Inline error alert */}
        {(operation.error || operation.stepData?.error_message) && (
          <Alert variant="destructive" className="mx-4 mt-4">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {operation.stepData?.error_message || operation.error || 'An error occurred during operation'}
            </AlertDescription>
          </Alert>
        )}
        {operationType === 'scraping' && scrapingMetrics?.errors?.length ? (
          <Alert variant="destructive" className="mx-4 mt-2">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Telemetry error: {scrapingMetrics.errors.join(', ')}
            </AlertDescription>
          </Alert>
        ) : null}

        <CardHeader className={cn(
          "pb-3",
          pipelineMode && "pb-2 px-4" // More compact header in pipeline mode
        )}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={cn(
                "rounded-lg",
                pipelineMode ? "p-1.5" : "p-2", // Smaller padding in pipeline mode
                isComplete && "bg-green-100",
                isFailed && "bg-red-100",
                !isComplete && !isFailed && "bg-blue-100"
              )}>
                <CurrentIcon className={cn(
                  pipelineMode ? "h-3.5 w-3.5" : "h-4 w-4", // Smaller icon in pipeline mode
                  isComplete && "text-green-700",
                  isFailed && "text-red-700",
                  !isComplete && !isFailed && "text-blue-700",
                  currentPhase.id === 'preparing' && "animate-pulse",
                  (currentPhase.id === 'scanning' || currentPhase.id === 'downloading') && "animate-bounce"
                )} />
              </div>
              <div>
                <h3 className="font-semibold text-base">
                  {pipelineMode && stageNumber && (
                    <span className="inline-flex items-center justify-center w-6 h-6 rounded-full bg-blue-100 text-blue-800 text-xs font-bold mr-2">
                      {stageNumber}
                    </span>
                  )}
                  {operationTitle}
                </h3>
                <div className="text-xs text-muted-foreground mt-1">
                  {pipelineMode && stageNumber && totalStages ? (
                    <>Stage {stageNumber} of {totalStages}</>
                  ) : (
                    <>Stage {stageDisplayInfo.order} of {stageDisplayInfo.total}</>
                  )}
                </div>
              </div>
            </div>

            <div className="flex flex-col items-end gap-1">
              <Badge
                variant={badgeVariant}
                className="font-medium"
              >
                {badgeLabel}
              </Badge>
              {telemetryMissing && (
                <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
                  <AlertCircle className="h-3 w-3" />
                  Awaiting telemetry
                </div>
              )}
              {stageKPI && (
                <div className="text-xs text-muted-foreground">
                  <span className="font-semibold text-foreground">{stageKPI.primary}</span>{' '}
                  <span>{stageKPI.secondary}</span>
                </div>
              )}
            </div>
          </div>
        </CardHeader>

        <CardContent className={cn(
          "space-y-3",
          pipelineMode && "space-y-2 p-4" // More compact spacing and padding in pipeline mode
        )}>
          {operationType === 'scraping' ? (
            <ScrapingStageCard
              metadata={stepMetadata}
              statusMessage={statusMessage}
              isComplete={isComplete}
              isFailed={isFailed}
              variant="progress"
            />
          ) : operationType === 'processing' ? (
            stepMetadata?.phase === 'preparing' || !stepMetadata?.file_statuses ? (
              <div className="flex flex-col items-center justify-center p-8 space-y-4 text-muted-foreground bg-gray-50 dark:bg-gray-900/50 rounded-lg border border-dashed border-gray-200 dark:border-gray-800">
                <Clock className="h-8 w-8 animate-pulse text-blue-500" />
                <p className="text-sm font-medium">Initializing Data Processor...</p>
                <p className="text-xs">Scanning for new files to process</p>
              </div>
            ) : (
              <ProcessingStageCard
                metadata={stepMetadata}
                statusMessage={statusMessage}
                isComplete={isComplete}
                isFailed={isFailed}
                variant="progress"
              />
            )
          ) : operationType === 'indices' ? (
            <IndicesStageCard
              metadata={stepMetadata}
              statusMessage={statusMessage}
              isComplete={isComplete}
              isFailed={isFailed}
              variant="progress"
            />
          ) : operationType === 'liquidity' ? (
            <LiquidityStageCard
              metadata={stepMetadata}
              statusMessage={statusMessage}
              isComplete={isComplete}
              isFailed={isFailed}
              variant="progress"
            />
          ) : (
            <>
              {/* Always show progress bar for all operations */}
              <div className={cn("space-y-2", pipelineMode && "space-y-1")}>
                <div className={cn(
                  "flex justify-between",
                  pipelineMode ? "text-xs" : "text-sm"
                )}>
                  <span className="text-muted-foreground">Progress</span>
                  <span className="font-medium">{visualProgress}%</span>
                </div>
                <Progress
                  value={visualProgress}
                  className="h-4"
                />
              </div>

              {/* Status message */}
              <div className={cn(
                "p-3 rounded-lg",
                isComplete && "bg-green-50 dark:bg-green-950",
                isFailed && "bg-red-50 dark:bg-red-950",
                !isComplete && !isFailed && "bg-blue-50 dark:bg-blue-950"
              )}>
                <p className={cn(
                  "text-sm font-medium",
                  isComplete && "text-green-800 dark:text-green-200",
                  isFailed && "text-red-800 dark:text-red-200",
                  !isComplete && !isFailed && "text-blue-800 dark:text-blue-200"
                )}>
                  {statusMessage}
                </p>

                {/* File stats if available */}
                {stepMetadata && (stepMetadata.files_processed || stepMetadata.speed) && (
                  <div className="flex items-center gap-4 mt-2 text-xs opacity-70">
                    {stepMetadata.files_processed && (
                      <span>Files: {stepMetadata.files_processed}</span>
                    )}
                    {stepMetadata.speed && (
                      <span>Speed: {stepMetadata.speed.toFixed(1)}/min</span>
                    )}
                  </div>
                )}
              </div>
            </>
          )}

          {/* Enhanced stage timing display for completed pipeline */}
          {completedPipeline && pipelineMetadata && (
            <>
              <Separator />
              <div className="bg-blue-50 dark:bg-blue-950 rounded-lg p-4 space-y-3">
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-blue-600" />
                  <h4 className="font-semibold text-blue-800 dark:text-blue-200">Stage Performance</h4>
                </div>

                {/* Stage timing from pipeline metadata */}
                {pipelineMetadata.stage_timeline && (
                  <div className="text-sm">
                    <div className="space-y-1">
                      <span className="text-blue-600 dark:text-blue-400 font-medium">Stage Duration:</span>
                      <p className="text-blue-800 dark:text-blue-200">
                        {(() => {
                          try {
                            const stageTiming = pipelineMetadata?.stage_timeline?.find(
                              (stage: any) => stage?.stage_id === operationType
                            )
                            return stageTiming?.duration_ms
                              ? `${(stageTiming.duration_ms / 1000).toFixed(1)}s`
                              : 'N/A'
                          } catch (error) {
                            console.warn('Stage timing lookup failed:', error)
                            return 'N/A'
                          }
                        })()}
                      </p>
                    </div>
                    {pipelineMetadata.total_duration && (
                      <div className="space-y-1">
                        <span className="text-blue-600 dark:text-blue-400 font-medium">Pipeline Total:</span>
                        <p className="text-blue-800 dark:text-blue-200">
                          {(pipelineMetadata.total_duration / 1000).toFixed(1)}s
                        </p>
                      </div>
                    )}
                  </div>
                )}

                {/* Stage-specific completion metrics */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="space-y-1">
                    <span className="text-blue-600 dark:text-blue-400 font-medium">Status:</span>
                    <p className="text-blue-800 dark:text-blue-200">Completed</p>
                  </div>
                  <div className="space-y-1">
                    <span className="text-blue-600 dark:text-blue-400 font-medium">Stage:</span>
                    <p className="text-blue-800 dark:text-blue-200 capitalize">{stageInfo.name}</p>
                  </div>
                </div>
              </div>
            </>
          )}

          {/* Comprehensive results summary for completed operations */}
          {isComplete && (
            <>
              <Separator />
              <div className="bg-green-50 dark:bg-green-950 rounded-lg p-4 space-y-3">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4 text-green-600" />
                  <h4 className="font-semibold text-green-800 dark:text-green-200">Operation Summary</h4>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
                  <div className="space-y-1">
                    <span className="text-green-600 dark:text-green-400 font-medium">Operation Type:</span>
                    <p className="text-green-800 dark:text-green-200 capitalize">{operationType}</p>
                  </div>

                  {/* Operation-specific results */}
                  {operationType === 'scraping' && stepMetadata?.from_date && stepMetadata?.to_date && (
                    <>
                      <div className="space-y-1">
                        <span className="text-green-600 dark:text-green-400 font-medium">Date Range:</span>
                        <p className="text-green-800 dark:text-green-200">
                          {stepMetadata.from_date} to {stepMetadata.to_date}
                        </p>
                      </div>
                      <div className="space-y-1">
                        <span className="text-green-600 dark:text-green-400 font-medium">Files Downloaded:</span>
                        <p className="text-green-800 dark:text-green-200">
                          {stepMetadata?.downloaded_files?.length || 0} files
                        </p>
                      </div>
                    </>
                  )}

                  {operationType === 'processing' && (
                    <div className="space-y-1">
                      <span className="text-green-600 dark:text-green-400 font-medium">Output Format:</span>
                      <p className="text-green-800 dark:text-green-200">CSV files ready for analysis</p>
                    </div>
                  )}

                  {operationType === 'indices' && (
                    <div className="space-y-1">
                      <span className="text-green-600 dark:text-green-400 font-medium">Index Data:</span>
                      <p className="text-green-800 dark:text-green-200">ISX60 performance metrics</p>
                    </div>
                  )}

                  {operationType === 'liquidity' && (() => {
                    const tickersAnalyzed = toNumber(stepMetadata?.tickers_analyzed)
                    const averageValue = toNumber(stepMetadata?.average_value)
                    const buckets = stepMetadata?.liquidity_buckets as Record<string, number> | undefined
                    const bucketEntries = buckets
                      ? ['high', 'medium', 'low'].map(bucket => ({
                        label: bucket.charAt(0).toUpperCase() + bucket.slice(1),
                        value: formatNumber(buckets[bucket] ?? 0)
                      }))
                      : null

                    return (
                      <>
                        <div className="space-y-1">
                          <span className="text-green-600 dark:text-green-400 font-medium">Tickers analyzed:</span>
                          <p className="text-green-800 dark:text-green-200">
                            {tickersAnalyzed !== undefined ? formatNumber(tickersAnalyzed) : '—'}
                          </p>
                        </div>
                        {averageValue !== undefined && (
                          <div className="space-y-1">
                            <span className="text-green-600 dark:text-green-400 font-medium">Average daily value:</span>
                            <p className="text-green-800 dark:text-green-200">{formatNumber(averageValue)} IQD</p>
                          </div>
                        )}
                        {bucketEntries && (
                          <div className="space-y-1">
                            <span className="text-green-600 dark:text-green-400 font-medium">Liquidity buckets:</span>
                            <p className="text-green-800 dark:text-green-200">
                              {bucketEntries.map(({ label, value }, idx) => (
                                <span key={`${label}-${value}`}>
                                  {label}: {value}
                                  {idx < bucketEntries.length - 1 && <span className="mx-1 text-muted-foreground">|</span>}
                                </span>
                              ))}
                            </p>
                          </div>
                        )}
                      </>
                    )
                  })()}

                  {operationType === 'analysis' && (() => {
                    const priceAlerts = toNumber(stepMetadata?.price_alerts)
                    const rsiAlerts = toNumber(stepMetadata?.rsi_alerts)
                    const tradedStocks = toNumber(stepMetadata?.traded_stocks)
                    return (
                      <>
                        <div className="space-y-1">
                          <span className="text-green-600 dark:text-green-400 font-medium">Price alerts:</span>
                          <p className="text-green-800 dark:text-green-200">
                            {priceAlerts !== undefined ? formatNumber(priceAlerts) : '—'}
                          </p>
                        </div>
                        <div className="space-y-1">
                          <span className="text-green-600 dark:text-green-400 font-medium">RSI alerts:</span>
                          <p className="text-green-800 dark:text-green-200">
                            {rsiAlerts !== undefined ? formatNumber(rsiAlerts) : '—'}
                          </p>
                        </div>
                        {tradedStocks !== undefined && (
                          <div className="space-y-1">
                            <span className="text-green-600 dark:text-green-400 font-medium">Traded stocks processed:</span>
                            <p className="text-green-800 dark:text-green-200">{formatNumber(tradedStocks)}</p>
                          </div>
                        )}
                      </>
                    )
                  })()}
                </div>
              </div>
            </>
          )}

          {/* Stage status indicator - shows when stage is running or ready */}
          {!isComplete && (
            <div className="flex items-center gap-2 mb-2" aria-live="polite">
              <div
                className={`h-2 w-2 rounded-full ${stageIndicatorColor
                  }`}
                aria-hidden="true"
              />
              <span className="text-sm font-medium text-muted-foreground">
                {stageIndicatorLabel}
              </span>
            </div>
          )}

          {/* Completion actions */}
          {isComplete && onNextOperation && (
            <>
              <Separator />
              <div className="space-y-3">
                <p className="text-sm font-medium text-green-700 dark:text-green-300">
                  {canActivateStage ? 'Ready for next step!' : 'Waiting for previous stages to complete'}
                </p>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  {operationType === 'scraping' && (
                    <>
                      <Button
                        variant="default"
                        size="default"
                        disabled={actionDisabled}
                        onClick={() => onNextOperation('processing')}
                        className="h-11 justify-start"
                      >
                        <Zap className="h-4 w-4 mr-2" />
                        Process Data · {actionDisplayText}
                      </Button>
                      <Button
                        variant="outline"
                        size="default"
                        disabled={actionDisabled}
                        onClick={() => onNextOperation('indices')}
                        className="h-11 justify-start"
                      >
                        <FileSearch className="h-4 w-4 mr-2" />
                        Extract Indices · {actionDisplayText}
                      </Button>
                      <Button
                        variant="ghost"
                        size="default"
                        disabled={actionDisabled}
                        onClick={() => onNextOperation('full_pipeline')}
                        className="h-11 justify-start"
                      >
                        <ArrowRight className="h-4 w-4 mr-2" />
                        Run All · {actionDisplayText}
                      </Button>
                    </>
                  )}
                  {operationType === 'processing' && (
                    <>
                      <Button
                        variant="default"
                        size="default"
                        disabled={actionDisabled}
                        onClick={() => onNextOperation('indices')}
                        className="h-11 justify-start"
                      >
                        <FileSearch className="h-4 w-4 mr-2" />
                        Extract Indices · {actionDisplayText}
                      </Button>
                      <Button
                        variant="outline"
                        size="default"
                        disabled={actionDisabled}
                        onClick={() => onNextOperation('liquidity')}
                        className="h-11 justify-start"
                      >
                        <Zap className="h-4 w-4 mr-2" />
                        Liquidity Analysis · {actionDisplayText}
                      </Button>
                    </>
                  )}
                </div>
              </div>
            </>
          )}

          {timingDisplay && (
            <div className="mt-3">
              {timingDisplay}
            </div>
          )}

          {/* Error actions */}
          {isFailed && (
            <div className="flex items-center gap-2">
              <AlertCircle className="h-4 w-4 text-red-600" />
              <span className="text-sm text-red-600 font-medium">
                Operation failed - please check logs and try again
              </span>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Pipeline connector */}
      {pipelineMode && showConnector && (
        <PipelineConnector
          isActive={isStageActive}
          isComplete={stageStatus === 'completed'}
        />
      )}
    </>
  )
}

export default UnifiedOperationProgress
