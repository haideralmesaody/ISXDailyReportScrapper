/**
 * useStageActivation Hook
 *
 * Centralized logic for managing stage activation state in pipeline operations.
 * Provides consistent behavior across components for stage activation and
 * handles special cases for completed pipelines.
 */

import { useMemo } from 'react'
import { identifyStage, getStageOrder } from '@/lib/operations/stage-mapping'

export interface ActivationState {
  isActive: boolean
  canActivate: boolean
  isCompleted: boolean
  stageId: string
  confidence: 'high' | 'medium' | 'low' | 'unknown'
  stageNumber: number
  totalStages: number
}

export interface UseStageActivationOptions {
  stageData?: any
  allStages?: any[]
  pipelineCompleted?: boolean
  stageNumber?: number
  totalStages?: number
}

/**
 * Hook for managing stage activation logic
 *
 * @param options - Configuration options
 * @returns Activation state object
 */
export function useStageActivation(options: UseStageActivationOptions = {}): ActivationState {
  const {
    stageData,
    allStages = [],
    pipelineCompleted = false,
    stageNumber,
    totalStages = 6
  } = options

  const activationState = useMemo(() => {
    // Default state
    const defaultState: ActivationState = {
      isActive: false,
      canActivate: false,
      isCompleted: false,
      stageId: 'scraping',
      confidence: 'low',
      stageNumber: stageNumber || 1,
      totalStages
    }

    // If pipeline is completed, all stages are considered active and completed
    if (pipelineCompleted) {
      let stageMapping
      try {
        stageMapping = identifyStage(stageData || {})
      } catch (error) {
        console.warn('Stage identification failed in pipeline completed:', error)
        stageMapping = { stageId: 'scraping', confidence: 'low' as const }
      }

      let stageOrderValue
      try {
        stageOrderValue = getStageOrder(stageMapping.stageId as any)
      } catch (error) {
        console.warn('Stage order failed in pipeline completed:', error)
        stageOrderValue = stageNumber || 1
      }

      return {
        isActive: true,
        canActivate: true,
        isCompleted: true,
        stageId: stageMapping.stageId,
        confidence: stageMapping.confidence as ActivationState['confidence'],
        stageNumber: stageNumber || stageOrderValue,
        totalStages
      }
    }

    // Normal activation logic for active pipelines
    if (!stageData) {
      return defaultState
    }

    let stageMapping
    try {
      stageMapping = identifyStage(stageData)
    } catch (error) {
      console.warn('Stage identification failed in normal activation:', error)
      stageMapping = { stageId: 'scraping', confidence: 'low' as const }
    }
    const stageId = stageMapping.stageId

    // Check if stage is completed
    const isCompleted = stageData?.status === 'completed'

    // Determine if stage can be activated based on pipeline state
    let canActivate = false
    let isActive = false

    if (Array.isArray(allStages) && allStages.length > 0) {
      let stageOrder
      try {
        stageOrder = getStageOrder(stageId as any)
      } catch (error) {
        console.warn('Stage order failed in normal activation:', error)
        stageOrder = stageNumber || 1
      }
      const stageIndex = stageOrder - 1 // Convert to 0-based index

      if (stageIndex <= 0) {
        // First stage is always active
        isActive = true
        canActivate = !isCompleted
      } else {
        // Find previous stage in order
        const previousStageOrder = stageOrder - 1
        const previousStage = allStages.find(s => {
          let mapping
          try {
            mapping = identifyStage(s)
          } catch (error) {
            console.warn('Stage identification failed for previous stage:', error)
            mapping = { stageId: 'scraping' }
          }
          let mappingOrder
          try {
            mappingOrder = getStageOrder(mapping.stageId as any)
          } catch (error) {
            console.warn('Stage order failed for previous stage:', error)
            mappingOrder = 1
          }
          return mappingOrder === previousStageOrder
        })

        // Stage is active if previous stage is completed
        isActive = previousStage?.status === 'completed'
        canActivate = isActive && !isCompleted
      }
    } else {
      // No previous stages data, assume first stage is active
      isActive = true
      canActivate = !isCompleted
    }

    let finalStageOrder
    try {
      finalStageOrder = getStageOrder(stageId as any)
    } catch (error) {
      console.warn('Final stage order failed:', error)
      finalStageOrder = stageNumber || 1
    }

    return {
      isActive,
      canActivate: canActivate || isCompleted, // Completed stages can be "activated" for display
      isCompleted,
      stageId,
      confidence: stageMapping.confidence as ActivationState['confidence'],
      stageNumber: stageNumber || finalStageOrder,
      totalStages
    }
  }, [stageData, allStages, pipelineCompleted, stageNumber, totalStages])

  return activationState
}

/**
 * Hook for managing activation state across multiple stages
 *
 * @param stages - Array of stage data
 * @param pipelineCompleted - Whether the entire pipeline is completed
 * @returns Object with activation states for all stages
 */
export function useMultiStageActivation(
  stages: any[] = [],
  pipelineCompleted: boolean = false
) {
  const activationStates = useMemo(() => {
    return stages.map((stage, index) =>
      useStageActivation({
        stageData: stage,
        allStages: stages,
        pipelineCompleted,
        stageNumber: index + 1,
        totalStages: stages.length
      })
    )
  }, [stages, pipelineCompleted])

  const activeStageCount = useMemo(() => {
    return activationStates.filter(state => state.isActive).length
  }, [activationStates])

  const completedStageCount = useMemo(() => {
    return activationStates.filter(state => state.isCompleted).length
  }, [activationStates])

  const pipelineProgress = useMemo(() => {
    if (stages.length === 0) return 0
    return (completedStageCount / stages.length) * 100
  }, [completedStageCount, stages.length])

  return {
    activationStates,
    activeStageCount,
    completedStageCount,
    pipelineProgress,
    isFullyCompleted: pipelineCompleted || completedStageCount === stages.length
  }
}

/**
 * Hook for stage timing information
 *
 * @param stageData - Stage operation data
 * @param pipelineMetadata - Pipeline metadata with timing information
 * @returns Stage timing information
 */
export function useStageTiming(stageData: any = {}, pipelineMetadata: any = null) {
  const timingInfo = useMemo(() => {
    // Try to get timing from stage metadata first
    const stageTimeline = stageData?.metadata?.stage_timeline
    const pipelineTimeline = pipelineMetadata?.stage_timeline
    const allTimelines = stageData?.metadata?.pipeline_summary?.stage_timeline

    const timelineSource = stageTimeline || pipelineTimeline || allTimelines
    const stageId = stageData?.metadata?.stage_id || stageData?.steps?.[0]?.metadata?.stage_id

    if (timelineSource && Array.isArray(timelineSource)) {
      const stageTiming = timelineSource.find((stage: any) =>
        stage.stage_id === stageId
      )

      if (stageTiming) {
        return {
          duration: stageTiming.duration_ms,
          durationFormatted: stageTiming.duration_ms
            ? `${(stageTiming.duration_ms / 1000).toFixed(1)}s`
            : 'N/A',
          startedAt: stageTiming.started_at,
          completedAt: stageTiming.completed_at,
          hasTiming: !!stageTiming.duration_ms
        }
      }
    }

    return {
      duration: null,
      durationFormatted: 'N/A',
      startedAt: null,
      completedAt: null,
      hasTiming: false
    }
  }, [stageData, pipelineMetadata])

  return timingInfo
}

export default useStageActivation
