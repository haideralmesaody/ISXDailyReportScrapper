/**
 * Stage Mapping with React Hooks
 *
 * Simplified stage mapping using React hooks for memoization
 * No module-level state or side effects - all state managed by React
 */

'use client'

import { useMemo, useCallback, useRef } from 'react'
import { identifyStagePure, createOperationHash, type StageMappingResult } from './stage-identification'
import { getStageInfo, getStageOrder, getPipelineStageOrder } from './stage-utils'
import type { StageId } from './stage-constants'

// Simple cache interface for React-managed state
interface CacheEntry {
  result: StageMappingResult
  timestamp: number
  operationHash: string
  accessCount: number
  lastAccessed: number
}

const CACHE_TTL_MS = 5000 // 5 seconds cache TTL
const MAX_CACHE_SIZE = 50 // Smaller cache size for React-managed state

/**
 * React hook for stage identification with memoization
 */
export function useStageIdentification() {
  // React-managed cache - no module-level state
  const cacheRef = useRef<Map<string, CacheEntry>>(new Map())

  // Performance stats (optional, for development)
  const statsRef = useRef({
    hits: 0,
    misses: 0,
    totalRequests: 0
  })

  // Clean up expired cache entries
  const cleanupExpiredEntries = useCallback(() => {
    const cache = cacheRef.current
    const now = Date.now()

    for (const [key, entry] of cache.entries()) {
      if (now - entry.timestamp > CACHE_TTL_MS) {
        cache.delete(key)
      }
    }
  }, [])

  // Evict least recently used entries
  const evictLeastRecentlyUsed = useCallback(() => {
    const cache = cacheRef.current
    if (cache.size <= MAX_CACHE_SIZE) {
      return
    }

    const entries = Array.from(cache.entries())
    entries.sort(([, a], [, b]) => {
      if (a.accessCount !== b.accessCount) {
        return a.accessCount - b.accessCount
      }
      return a.lastAccessed - b.lastAccessed
    })

    const entriesToRemove = Math.max(1, Math.floor(entries.length * 0.25))

    for (let i = 0; i < entriesToRemove; i++) {
      const entry = entries[i]
      if (!entry) continue
      const [cacheKey] = entry
      cache.delete(cacheKey)
    }
  }, [])

  // Update cache access statistics
  const updateCacheAccess = useCallback((cacheKey: string) => {
    const cache = cacheRef.current
    const entry = cache.get(cacheKey)
    if (entry) {
      entry.accessCount++
      entry.lastAccessed = Date.now()
    }
  }, [])

  /**
   * Identify stage with caching
   */
  const identifyStage = useCallback((operation: any): StageMappingResult => {
    const stats = statsRef.current
    stats.totalRequests++

    // Clean up expired entries first
    cleanupExpiredEntries()

    const operationHash = createOperationHash(operation)
    const cache = cacheRef.current

    // Check cache
    const cachedEntry = cache.get(operationHash)
    const now = Date.now()

    if (cachedEntry && (now - cachedEntry.timestamp < CACHE_TTL_MS)) {
      updateCacheAccess(operationHash)
      stats.hits++

      return cachedEntry.result
    }

    // Calculate result using pure function
    const result = identifyStagePure(operation)

    // Cache the result
    cache.set(operationHash, {
      result,
      timestamp: now,
      operationHash,
      accessCount: 1,
      lastAccessed: now
    })

    evictLeastRecentlyUsed()
    stats.misses++

    return result
  }, [cleanupExpiredEntries, updateCacheAccess, evictLeastRecentlyUsed])

  /**
   * Get cache statistics (development only)
   */
  const getCacheStatistics = useCallback(() => {
    const stats = statsRef.current
    const cache = cacheRef.current

    const hitRate = stats.totalRequests > 0 ? (stats.hits / stats.totalRequests * 100).toFixed(2) : '0.00'

    return {
      ...stats,
      hitRate: `${hitRate}%`,
      cacheSize: cache.size,
      maxSize: MAX_CACHE_SIZE,
      cacheUtilization: `${(cache.size / MAX_CACHE_SIZE * 100).toFixed(2)}%`
    }
  }, [])

  return {
    identifyStage,
    getCacheStatistics
  }
}

/**
 * React hook for stage information with memoization
 */
export function useStageInfo(stageId: StageId) {
  return useMemo(() => getStageInfo(stageId), [stageId])
}

/**
 * React hook for stage order with memoization
 */
export function useStageOrder(stageId: StageId) {
  return useMemo(() => getStageOrder(stageId), [stageId])
}

/**
 * React hook for pipeline stage order with memoization
 */
export function usePipelineStageOrder() {
  return useMemo(() => getPipelineStageOrder(), [])
}

// Re-export types for backward compatibility
export type { StageMappingResult } from './stage-identification'
export type { StageId } from './stage-constants'

// Re-export pure functions for non-hook usage
export { identifyStagePure } from './stage-identification'
export { getStageInfo, getStageOrder, getAllStages, getTotalStages, isStageReliable, canActivateStage, extractStageFileCount, getPipelineStageOrder, getStageMetrics } from './stage-utils'

// Legacy exports for backward compatibility
export { STAGE_DEFINITIONS, OPERATION_TYPE_TO_STAGE, STAGE_PATTERNS, STAGE_METRICS } from './stage-constants'
export type { StageDefinition, StageMetrics, ConfidenceLevel } from './stage-constants'

// Legacy identifyStage function (non-hook version) for backward compatibility
// Wrapper with error handling for safety
export const identifyStage = (operation: any) => {
  try {
    return identifyStagePure(operation)
  } catch (error) {
    console.warn('Legacy identifyStage failed, returning fallback:', error)
    return {
      stageId: 'scraping' as StageId,
      confidence: 'low' as const,
      source: 'fallback' as const
    }
  }
}
