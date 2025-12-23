/**
 * Stage Identification Pure Functions
 *
 * Pure functions for stage identification - no module-level state or side effects
 * All functions are deterministic and don't depend on external state
 */

import type { StageId, ConfidenceLevel } from './stage-constants'
import { STAGE_DEFINITIONS, OPERATION_TYPE_TO_STAGE, STAGE_PATTERNS } from './stage-constants'

export interface StageMappingResult {
  stageId: StageId
  confidence: ConfidenceLevel
  source: 'explicit' | 'operation_type' | 'pattern_match' | 'name_analysis' | 'fallback'
  metadata?: {
    matchedPattern?: string
    operationType?: string
    stageName?: string
    origin?: string
    stepIndex?: number
  }
}

/**
 * Helper to normalize candidate stage identifiers
 */
export function normalizeStageId(value: unknown): StageId | undefined {
  if (typeof value !== 'string') return undefined
  const normalized = value.trim().toLowerCase()
  return (normalized in STAGE_DEFINITIONS)
    ? normalized as StageId
    : undefined
}

/**
 * Helper to create a hash of the operation object for cache key (moved from module-level to pure function)
 */
export function createOperationHash(operation: any): string {
  if (!operation) return 'empty-operation'

  // Create a simple hash from key operation properties
  const keyFields = [
    operation.stage_id,
    operation.operation_id,
    operation.name,
    operation.status,
    operation.metadata?.stage_id,
    operation.metadata?.operation_type,
    operation.steps?.length,
    operation.progress
  ]

  return keyFields.filter(Boolean).join('|')
}

/**
 * Resolve stage from candidate value with origin tracking
 */
export function resolveStage(
  candidate: unknown,
  origin?: string
): { stageId: StageId; origin?: string } | undefined {
  const stageId = normalizeStageId(candidate)
  return stageId ? { stageId, origin } : undefined
}

/**
 * Get explicit stage candidates from operation object
 */
export function getExplicitStageCandidates(operation: any): Array<{ value: unknown; origin: string }> {
  return [
    { value: operation?.current_step, origin: 'current_step' },
    { value: operation?.stage_id, origin: 'stage_id' },
    { value: operation?.metadata?.stage_id, origin: 'metadata.stage_id' },
    { value: operation?.data?.stage_id, origin: 'data.stage_id' },
    { value: operation?.metadata?.target_stage_id, origin: 'metadata.target_stage_id' },
    { value: operation?.data?.target_stage_id, origin: 'data.target_stage_id' }
  ]
}

/**
 * Check operation steps for stage information
 */
export function checkStepsForStage(operation: any): { stageId: StageId; origin: string; stepIndex: number } | null {
  if (!Array.isArray(operation?.steps)) return null

  for (let index = 0; index < operation.steps.length; index += 1) {
    const step = operation.steps[index]

    const resolved = (
      resolveStage(step?.metadata?.stage_id, `steps[${index}].metadata.stage_id`) ??
      resolveStage(step?.metadata?.target_stage_id, `steps[${index}].metadata.target_stage_id`) ??
      resolveStage(step?.metadata?.operation_type, `steps[${index}].metadata.operation_type`)
    )

    if (resolved) {
      return {
        ...resolved,
        stepIndex: index
      }
    }
  }

  return null
}

/**
 * Extract operation type from operation object
 */
export function getOperationType(operation: any): string | undefined {
  return (
    operation?.metadata?.operation ??
    operation?.metadata?.operation_type ??
    operation?.data?.operation_type ??
    operation?.name
  )
}

/**
 * Get stage name from operation object
 */
export function getStageName(operation: any): string | undefined {
  return (
    operation?.stage_name ??
    operation?.metadata?.stage_name ??
    operation?.data?.stage_name
  )
}

/**
 * Handle 'run_stage' operation type with target stage resolution
 */
export function handleRunStageOperation(operation: any): { stageId: StageId; origin: string } | null {
  // First check target_stage_id in metadata (primary method)
  const resolved = (
    resolveStage(operation?.metadata?.target_stage_id, 'metadata.target_stage_id') ??
    resolveStage(operation?.data?.target_stage_id, 'data.target_stage_id')
  )

  if (resolved) {
    return resolved
  }

  // Fallback: check if operation_type directly specifies a valid stage
  const operationType = operation?.metadata?.operation_type || operation?.operation_type
  if (operationType) {
    const directResolved = resolveStage(operationType, 'operation_type')
    if (directResolved) {
      return directResolved
    }
  }

  return null
}

/**
 * Perform pattern matching on operation name
 */
export function performPatternMatching(operationName: string): { stageId: StageId; matchedPattern: string } | null {
  for (const [stageId, patterns] of Object.entries(STAGE_PATTERNS)) {
    for (const pattern of patterns) {
      if (pattern.test(operationName)) {
        return {
          stageId: stageId as StageId,
          matchedPattern: pattern.toString()
        }
      }
    }
  }

  return null
}

/**
 * Main stage identification function - pure function without module-level state
 */
export function identifyStagePure(operation: any): StageMappingResult {
  const startTime = performance.now()

  // Check explicit candidates first
  const explicitCandidates = getExplicitStageCandidates(operation)
  for (const candidate of explicitCandidates) {
    const resolved = resolveStage(candidate.value, candidate.origin)
    if (resolved) {
      return {
        stageId: resolved.stageId,
        confidence: 'high' as const,
        source: 'explicit' as const,
        metadata: { origin: resolved.origin }
      }
    }
  }

  // Check steps for stage information
  const stepResult = checkStepsForStage(operation)
  if (stepResult) {
    return {
      stageId: stepResult.stageId,
      confidence: 'high' as const,
      source: 'explicit' as const,
      metadata: { origin: stepResult.origin, stepIndex: stepResult.stepIndex }
    }
  }

  // Check operation type
  const operationTypeCandidate = getOperationType(operation)
  if (typeof operationTypeCandidate === 'string') {
    const normalizedType = operationTypeCandidate.toLowerCase().trim()

    if (normalizedType === 'run_stage') {
      const runStageResult = handleRunStageOperation(operation)
      if (runStageResult) {
        return {
          stageId: runStageResult.stageId,
          confidence: 'high' as const,
          source: 'operation_type' as const,
          metadata: { operationType: normalizedType, origin: runStageResult.origin }
        }
      }
    } else if (OPERATION_TYPE_TO_STAGE[normalizedType]) {
      return {
        stageId: OPERATION_TYPE_TO_STAGE[normalizedType],
        confidence: 'high' as const,
        source: 'operation_type' as const,
        metadata: { operationType: normalizedType }
      }
    }
  }

  // Check stage name
  const stageNameCandidate = getStageName(operation)
  if (typeof stageNameCandidate === 'string') {
    const match = Object.entries(STAGE_DEFINITIONS).find(
      ([, definition]) => definition.name.toLowerCase() === stageNameCandidate.toLowerCase()
    )

    if (match) {
      return {
        stageId: match[0] as StageId,
        confidence: 'high' as const,
        source: 'explicit' as const,
        metadata: { stageName: stageNameCandidate }
      }
    }
  }

  // Pattern matching on operation name
  const operationName = typeof operation?.name === 'string' ? operation.name : ''
  if (operationName) {
    const patternResult = performPatternMatching(operationName)
    if (patternResult) {
      return {
        stageId: patternResult.stageId,
        confidence: 'medium' as const,
        source: 'pattern_match' as const,
        metadata: { matchedPattern: patternResult.matchedPattern }
      }
    }
  }

  // Fallback to scraping
  return {
    stageId: 'scraping' as StageId,
    confidence: 'low' as const,
    source: 'fallback' as const
  }
}