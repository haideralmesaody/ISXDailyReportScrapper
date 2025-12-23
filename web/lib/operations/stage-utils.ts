/**
 * Stage Utility Functions
 *
 * Pure utility functions for stage operations - no module-level state
 */

import type { StageId, StageDefinition, StageMetrics } from './stage-constants'
import { STAGE_DEFINITIONS, STAGE_METRICS } from './stage-constants'

/**
 * Get stage display information with memoization (component-level)
 */
export function getStageInfo(stageId: StageId): StageDefinition {
  return STAGE_DEFINITIONS[stageId] || STAGE_DEFINITIONS.scraping
}

/**
 * Get stage-specific metrics configuration with memoization (component-level)
 */
export function getStageMetrics(stageId: StageId): StageMetrics {
  return STAGE_METRICS[stageId] || STAGE_METRICS.scraping
}

/**
 * Determine if stage identification is reliable enough
 */
export function isStageReliable(mapping: { confidence: string; source: string }): boolean {
  return mapping.confidence === 'high' ||
         (mapping.confidence === 'medium' && mapping.source === 'pattern_match')
}

/**
 * Get all available stages for selection
 */
export function getAllStages() {
  return Object.entries(STAGE_DEFINITIONS).map(([id, info]) => ({
    id,
    ...info
  }))
}

/**
 * Get stage order number for display
 */
export function getStageOrder(stageId: StageId): number {
  return STAGE_DEFINITIONS[stageId]?.order || 999
}

/**
 * Get pipeline stage order excluding wrapper operations
 */
export function getPipelineStageOrder(): StageId[] {
  try {
    if (!STAGE_DEFINITIONS) {
      console.warn('STAGE_DEFINITIONS is undefined, returning default order')
      return ['scraping', 'processing', 'indices', 'liquidity', 'analysis', 'indicators']
    }

    const order = Object.entries(STAGE_DEFINITIONS)
      .sort(([, a], [, b]) => (a?.order || 999) - (b?.order || 999))
      .map(([key]) => key as StageId)
      .filter(id => !['full_pipeline'].includes(id)) // Exclude wrapper operations

    return order.length > 0 ? order : ['scraping', 'processing', 'indices', 'liquidity', 'analysis', 'indicators']
  } catch (error) {
    console.warn('getPipelineStageOrder failed, returning default order:', error)
    return ['scraping', 'processing', 'indices', 'liquidity', 'analysis', 'indicators']
  }
}

/**
 * Get total number of pipeline stages
 * Provides single source of truth for stage counting
 */
export function getTotalStages(): number {
  try {
    const order = getPipelineStageOrder()
    return Array.isArray(order) ? order.length : 6
  } catch (error) {
    console.warn('getTotalStages failed, returning default count:', error)
    return 6
  }
}

/**
 * Check if a stage is active based on previous stage completion
 */
export function isStageActive(stageId: StageId, allStages: any[]): boolean {
  const stageOrder = getPipelineStageOrder()
  const stageIndex = stageOrder.indexOf(stageId)

  if (stageIndex <= 0) return true // First stage always active

  // Find previous stage in order
  const previousStageId = stageOrder[stageIndex - 1]
  const previousStage = allStages.find(s =>
    s.steps?.[0]?.metadata?.stage_id === previousStageId ||
    s.steps?.[0]?.id === previousStageId
  )

  return previousStage?.status === 'completed'
}

/**
 * Check if a stage can be activated based on pipeline state
 */
export function canActivateStage(stageData: any, allStages: any[]): boolean {
  const stageId = stageData.steps?.[0]?.metadata?.stage_id || stageData.steps?.[0]?.id
  return isStageActive(stageId, allStages)
}

/**
 * Extract stage-specific file count from operation
 */
export function extractStageFileCount(operation: any, stageId: StageId): number {
  try {
    const stageMetricsMap = operation?.metadata?.stage_metrics as Record<string, Record<string, any>> | undefined
    const stageSpecificMetrics = stageMetricsMap && stageId ? stageMetricsMap[stageId] : undefined

    const sources = [
      stageSpecificMetrics,
      operation?.metadata,
      operation?.data
    ].filter(Boolean) as Array<Record<string, any>>

    for (const source of sources) {
      const value = source?.files_processed
      if (typeof value === 'number' && value > 0) {
        return value
      }
    }

    // Stage-specific file count keys
    const fileCountKeys = {
      scraping: ['files_downloaded', 'files_processed', 'total_files'],
      processing: ['files_processed', 'files_converted', 'total_files'],
      indices: ['indices_extracted', 'files_created'],
      liquidity: ['analysis_files', 'files_generated'],
      analysis: ['analysis_files', 'reports_generated'],
      indicators: ['indicators_calculated', 'files_created'],
      full_pipeline: ['files_processed', 'total_files']
    }

    const keys = fileCountKeys[stageId] || fileCountKeys.scraping

    for (const key of keys) {
      for (const source of sources) {
        const value = source?.[key]
        if (typeof value === 'number' && value > 0) {
          return value
        }
      }
    }

    // Check steps for completion data
    if (operation?.steps && Array.isArray(operation.steps)) {
      for (const step of operation.steps) {
        if (step.status === 'completed') {
          const stepFiles = step.metadata?.files_processed
          if (typeof stepFiles === 'number' && stepFiles > 0) {
            return stepFiles
          }
        }
      }
    }

    // Return reasonable defaults based on stage
    const defaults = {
      scraping: 15,        // Typical daily Excel files
      processing: 15,      // Market data processing and CSV conversion
      indices: 1,          // Single index file
      liquidity: 1,        // Single liquidity report
      indicators: 1,       // Indicators file
      analysis: 1,         // Analysis report
      full_pipeline: 30    // Combined operations
    }

    return defaults[stageId] || 1
  } catch (error) {
    console.warn('extractStageFileCount failed:', error)
    // Return safe default based on stage
    const safeDefaults = {
      scraping: 15,
      processing: 15,
      indices: 1,
      liquidity: 1,
      analysis: 1,
      indicators: 1,
      full_pipeline: 30
    }
    return safeDefaults[stageId] || 1
  }
}