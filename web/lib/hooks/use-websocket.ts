/**
 * Simplified WebSocket Hook for ISX Pulse
 *
 * Provides type-safe WebSocket functionality optimized for localhost communication.
 * Removed distributed system complexity while maintaining React compatibility.
 */

'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { ISXWebSocketClient, getWebSocketClient } from '@/lib/websocket'
import { normalizeStageTelemetry, normalizeOperationMetadata } from '@/lib/operations/telemetry-utils'
import type { ScrapingTelemetry, OperationDeltaEvent } from '@/lib/operations/types'

// ============================================================================
// Simplified WebSocket Hook Options
// ============================================================================

interface UseWebSocketOptions {
  autoConnect?: boolean
  enableValidation?: boolean
  enableMigration?: boolean
  debug?: boolean
}

interface UseOperationSnapshotOptions extends UseWebSocketOptions {
  operationId?: string
  includeHistory?: boolean
}

// ============================================================================
// Simplified WebSocket Hook
// ============================================================================

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const {
    autoConnect = true,
    enableValidation = true,
    enableMigration = true,
    debug = process.env.NODE_ENV === 'development'
  } = options

  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting'>('disconnected')
  const [lastError, setLastError] = useState<string | null>(null)
  const [isClientInitialized, setIsClientInitialized] = useState(false)
  const clientRef = useRef<ISXWebSocketClient | null>(null)

  // Initialize client safely in useEffect to prevent TDZ violations
  useEffect(() => {
    if (!clientRef.current && typeof window !== 'undefined') {
      try {
        clientRef.current = getWebSocketClient()
        setIsClientInitialized(true)
        // Mark listeners as ready to receive messages
        clientRef.current.markListenersReady()
        if (debug) {
          console.log('[useWebSocket] Client initialized successfully')
        }
      } catch (err) {
        console.error('[useWebSocket] Client initialization failed:', err)
        setConnectionStatus('error')
        setLastError(err instanceof Error ? err.message : 'Client initialization failed')
      }
    }
  }, [debug])

  const connect = useCallback(async () => {
    if (debug) {
      console.log(`[useWebSocket] Connecting to WebSocket...`)
    }

    // Ensure client is properly initialized before attempting connection
    if (!clientRef.current && typeof window !== 'undefined') {
      try {
        clientRef.current = getWebSocketClient()
        setIsClientInitialized(true)
        // Mark listeners as ready to receive messages
        clientRef.current.markListenersReady()
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Client initialization failed'
        setConnectionStatus('error')
        setLastError(errorMessage)
        if (debug) {
          console.error(`[useWebSocket] Client initialization failed:`, err)
        }
        return
      }
    }

    // If still no client (e.g., server-side), skip connection
    if (!clientRef.current) {
      if (debug) {
        console.log(`[useWebSocket] No client available - skipping connection`)
      }
      return
    }

    try {
      if (clientRef.current.isConnected()) {
        if (debug) {
          console.log(`[useWebSocket] Already connected`)
        }
        setConnectionStatus('connected')
        setLastError(null)
        return
      }

      await clientRef.current.connect()

      if (debug) {
        console.log(`[useWebSocket] Connection successful`)
      }

      setConnectionStatus('connected')
      setLastError(null)
    } catch (err) {
      setConnectionStatus('error')
      const errorMessage = err instanceof Error ? err.message : 'Unknown connection error'
      setLastError(errorMessage)

      if (debug) {
        console.error(`[useWebSocket] Connection failed:`, err)
      }
    }
  }, [debug])

  const disconnect = useCallback(() => {
    if (clientRef.current) {
      clientRef.current.disconnect()
      setConnectionStatus('disconnected')
      setLastError(null)
    }
  }, [])

  const subscribe = useCallback((event: string, handler: (data: any) => void) => {
    if (!clientRef.current) {
      if (debug) {
        console.warn('[useWebSocket] Client not initialized')
      }
      return () => {}
    }
    return clientRef.current.subscribe(event, handler)
  }, [debug])

  const send = useCallback((message: any) => {
    if (!clientRef.current) {
      if (debug) {
        console.warn('[useWebSocket] Client not initialized')
      }
      return
    }
    clientRef.current.send(message)
  }, [debug])

  // Auto-connect on mount if enabled
  useEffect(() => {
    if (autoConnect && isClientInitialized) {
      // Wait a moment for the page to fully load before connecting
      const timeout = setTimeout(() => {
        connect()
      }, 100)

      return () => clearTimeout(timeout)
    }
  }, [autoConnect, connect, isClientInitialized])

  // Subscribe to connection status updates
  useEffect(() => {
    if (!clientRef.current) return

    const unsubscribe = clientRef.current.subscribe('connection_status', (data) => {
      setConnectionStatus(data.status)
      if (data.status === 'connected') {
        setLastError(null)
      }
    })

    return unsubscribe
  }, [clientRef.current])

  return {
    connectionStatus,
    connect,
    disconnect,
    subscribe,
    send,
    isConnected: connectionStatus === 'connected',
    lastError
  }
}

// Re-export for backward compatibility
export default useWebSocket

export function useSystemStatus() {
  const { subscribe, connectionStatus } = useWebSocket({ autoConnect: true })
  const [systemStatus, setSystemStatus] = useState<any>(null)
  const completionFetchTimersRef = useRef<Record<string, NodeJS.Timeout>>({}) // Fix: Add missing useRef declaration

  useEffect(() => {
    const unsubscribe = subscribe('system_status', setSystemStatus)
    return unsubscribe
  }, [subscribe])

  useEffect(() => {
    return () => {
      Object.values(completionFetchTimersRef.current).forEach(timer => {
        clearTimeout(timer)
      })
      completionFetchTimersRef.current = {}
    }
  }, [completionFetchTimersRef]) // Fix: Add proper dependency

  return {
    systemStatus,
    connectionStatus,
    isHealthy: systemStatus?.healthy === true || connectionStatus === 'connected'
  }
}

export function useConnectionStatus() {
  const { connectionStatus, subscribe } = useWebSocket({ autoConnect: true })
  const [connectionHealth, setConnectionHealth] = useState<{
    status: string
    lastConnectedAt?: number
    totalDisconnections: number
    averageReconnectTime: number
  }>({
    status: 'disconnected',
    totalDisconnections: 0,
    averageReconnectTime: 0
  })

  const reconnectStartTime = useRef<number | null>(null)

  useEffect(() => {
    const unsubscribe = subscribe('connection_status', (data) => {
      setConnectionHealth(prev => {
        const newHealth = { ...prev }

        if (data.status === 'connected') {
          // Calculate reconnect time if we were reconnecting
          if (reconnectStartTime.current) {
            const reconnectTime = Date.now() - reconnectStartTime.current
            reconnectStartTime.current = null

            // Update average reconnect time
            if (prev.averageReconnectTime === 0) {
              newHealth.averageReconnectTime = reconnectTime
            } else {
              newHealth.averageReconnectTime = (prev.averageReconnectTime + reconnectTime) / 2
            }
          }

          newHealth.status = 'healthy'
          newHealth.lastConnectedAt = Date.now()
        } else if (data.status === 'disconnected' || data.status === 'error') {
          newHealth.status = 'unhealthy'
          reconnectStartTime.current = Date.now()
          newHealth.totalDisconnections = prev.totalDisconnections + 1
        } else if (data.status === 'reconnecting') {
          newHealth.status = 'degraded'
          if (!reconnectStartTime.current) {
            reconnectStartTime.current = Date.now()
          }
        }

        return newHealth
      })
    })

    return unsubscribe
  }, [subscribe])

  return {
    status: connectionStatus,
    isConnected: connectionStatus === 'connected',
    health: connectionHealth,
    isHealthy: connectionHealth.status === 'healthy',
    getTimeSinceLastConnection: () => {
      return connectionHealth.lastConnectedAt ? Date.now() - connectionHealth.lastConnectedAt : Infinity
    }
  }
}

// ============================================================================
// Snapshot Merge Utilities
// ============================================================================

const getTimestampValue = (value: any): number => {
  if (!value) return 0
  if (typeof value === 'number') return value
  if (value instanceof Date) return value.getTime()
  if (typeof value === 'string') {
    const parsed = Date.parse(value)
    return Number.isNaN(parsed) ? 0 : parsed
  }
  return 0
}

const normalizeStepId = (step: any): string | null => {
  if (!step) return null
  return step.step_id || step.id || step.stage_id || null
}

const mergeSteps = (existingSteps: any[] = [], incomingSteps: any[] = []) => {
  if (!incomingSteps.length) {
    return existingSteps
  }

  const stepMap = new Map<string, any>()
  const order: string[] = []

  const ensureOrder = (id: string) => {
    if (!order.includes(id)) {
      order.push(id)
    }
  }

  existingSteps.forEach((step) => {
    const id = normalizeStepId(step)
    if (!id) return
    ensureOrder(id)
    stepMap.set(id, step)
  })

  incomingSteps.forEach((step) => {
    const id = normalizeStepId(step)
    if (!id) return

    const previous = stepMap.get(id)
    const normalizedIncoming = normalizeStageTelemetry(step, previous?.metadata)

    const incomingTimestamp = getTimestampValue(
      normalizedIncoming.updated_at ||
      normalizedIncoming.metadata?.stage_updated_at ||
      normalizedIncoming.metadata?.updated_at
    )

    if (!previous) {
      ensureOrder(id)
      stepMap.set(id, {
        ...normalizedIncoming,
        updated_at:
          normalizedIncoming.updated_at ||
          normalizedIncoming.metadata?.stage_updated_at ||
          new Date().toISOString()
      })
      return
    }

    const currentTimestamp = getTimestampValue(
      previous?.updated_at ||
      previous?.metadata?.stage_updated_at ||
      previous?.metadata?.updated_at
    )

    if (incomingTimestamp === 0 || incomingTimestamp >= currentTimestamp) {
      stepMap.set(id, {
        ...previous,
        ...normalizedIncoming,
        status: normalizedIncoming.status ?? previous?.status,
        progress: typeof normalizedIncoming.progress === 'number'
          ? normalizedIncoming.progress
          : previous?.progress,
        metadata: normalizedIncoming.metadata,
        updated_at:
          normalizedIncoming.updated_at ||
          normalizedIncoming.metadata?.stage_updated_at ||
          previous?.updated_at ||
          new Date().toISOString()
      })
    }
  })

  return order.map((id) => stepMap.get(id)).filter(Boolean)
}

const shouldApplySnapshot = (existing: any, incomingTimestamp: number): boolean => {
  if (!existing) return true

  const existingTimestamp = getTimestampValue(
    existing.updated_at ||
    existing.metadata?.updated_at ||
    existing.metadata?.pipeline_summary?.updated_at
  )

  if (existingTimestamp === 0 || incomingTimestamp === 0) {
    return true
  }

  return incomingTimestamp >= existingTimestamp
}

const deriveStageIdFromDelta = (delta: OperationDeltaEvent, fallback?: string) => {
  if (!delta) return fallback || 'stage'
  return delta.stage_id || delta.metadata?.stage_id || fallback || 'stage'
}

const updateOperationWithDelta = (operation: any, delta: OperationDeltaEvent) => {
  const stageId = deriveStageIdFromDelta(
    delta,
    operation?.current_step ||
    (Array.isArray(operation?.steps) ? normalizeStepId(operation.steps[operation.steps.length - 1]) : undefined) ||
    'stage'
  )

  const timestampIso = delta.updated_at
    ? new Date(delta.updated_at).toISOString()
    : new Date().toISOString()

  const nextOperation = {
    ...operation,
    operation_id: operation?.operation_id || delta.operation_id,
    status: delta.status || operation?.status || 'running',
    current_step: stageId,
    message: delta.message ?? operation?.message ?? '',
    updated_at: timestampIso
  }

  if (typeof delta.progress === 'number') {
    nextOperation.progress = delta.progress
  } else if (typeof nextOperation.progress !== 'number') {
    nextOperation.progress = 0
  }

  const existingSteps = Array.isArray(operation?.steps) ? [...operation.steps] : []
  const existingIndex = existingSteps.findIndex((step: any) => normalizeStepId(step) === stageId)
  const previousStep = existingIndex >= 0 ? existingSteps[existingIndex] : undefined

  const normalizedStep = normalizeStageTelemetry({
    id: stageId,
    stage_id: stageId,
    status: delta.status || previousStep?.status || nextOperation.status,
    progress: typeof delta.progress === 'number'
      ? delta.progress
      : (typeof previousStep?.progress === 'number'
        ? previousStep.progress
        : (typeof nextOperation.progress === 'number' ? nextOperation.progress : 0)),
    message: delta.message ?? previousStep?.message ?? nextOperation.message,
    updated_at: timestampIso,
    metadata: {
      ...(previousStep?.metadata || {}),
      ...(delta.metadata || {})
    }
  }, previousStep?.metadata)

  if (existingIndex >= 0) {
    existingSteps[existingIndex] = {
      ...previousStep,
      ...normalizedStep
    }
  } else {
    existingSteps.push(normalizedStep)
  }

  const mergedMetadata = normalizeOperationMetadata(
    {
      ...(typeof operation?.metadata === 'object' ? operation.metadata : {}),
      ...(delta.metadata || {})
    },
    normalizedStep.metadata
  )

  return {
    ...nextOperation,
    steps: existingSteps,
    metadata: mergedMetadata
  }
}

const createOperationFromDelta = (delta: OperationDeltaEvent) => {
  const baseOperation = {
    operation_id: delta.operation_id,
    status: delta.status || 'running',
    progress: typeof delta.progress === 'number' ? delta.progress : 0,
    message: delta.message || '',
    current_step: deriveStageIdFromDelta(delta, 'stage'),
    steps: [],
    metadata: delta.metadata ? { ...delta.metadata } : {}
  }

  return updateOperationWithDelta(baseOperation, delta)
}

const applyDeltaToSnapshots = (snapshots: any[], delta: OperationDeltaEvent, includeHistory: boolean) => {
  const existingIndex = snapshots.findIndex(op => op.operation_id === delta.operation_id)
  const deltaTimestamp = getTimestampValue(delta.updated_at || Date.now())

  if (existingIndex >= 0) {
    const existing = snapshots[existingIndex]
    const existingTimestamp = getTimestampValue(
      existing.updated_at ||
      existing.metadata?.updated_at ||
      existing.metadata?.stage_updated_at
    )

    if (deltaTimestamp > 0 && existingTimestamp > 0 && deltaTimestamp < existingTimestamp) {
      if (debug) {
        console.log('[useOperationSnapshots] Ignoring older timestamp:', {
              operationId: delta.operation_id,
              deltaTimestamp,
              existingTimestamp
            })
      }
      return snapshots
    }

    const updated = updateOperationWithDelta(existing, delta)
    const next = [...snapshots]
    next[existingIndex] = updated
    return next
  }

  const created = createOperationFromDelta(delta)
  if (includeHistory) {
    return [...snapshots, created]
  }
  return [created]
}

const terminalStatuses = new Set(['completed', 'failed', 'skipped'])

const isTerminalStatus = (status?: string | null) => {
  if (!status) return false
  return terminalStatuses.has(status)
}

const normalizeFetchedSnapshot = (incoming: any, existing?: any) => {
  if (!incoming || typeof incoming !== 'object') {
    return existing
  }

  const normalizedSteps = Array.isArray(incoming.steps)
    ? incoming.steps
    : (Array.isArray(existing?.steps) ? existing?.steps : [])

  const normalizedMetadata = incoming.metadata && typeof incoming.metadata === 'object'
    ? incoming.metadata
    : (existing?.metadata && typeof existing.metadata === 'object'
      ? existing.metadata
      : {})

  const merged = {
    ...existing,
    ...incoming,
    steps: normalizedSteps,
    metadata: normalizedMetadata
  }

  if (!merged.updated_at) {
    merged.updated_at = new Date().toISOString()
  }

  return merged
}

// ============================================================================
// Simplified Operation Snapshot Hook
// ============================================================================

/**
 * Simplified Hook for consuming operation snapshots with minimal validation
 */
export function useOperationSnapshots(options: UseOperationSnapshotOptions = {}) {
  const {
    operationId,
    includeHistory = true,
    debug = process.env.NODE_ENV === 'development'
  } = options

  const {
    subscribe,
    connectionStatus,
    isConnected,
    lastError
  } = useWebSocket({
    debug
  })

  // Use simple OperationSnapshot type instead of complex DTO
  const [snapshots, setSnapshots] = useState<any[]>([])
  const [scrapingTelemetry, setScrapingTelemetry] = useState<Record<string, ScrapingTelemetry>>({})

  // Initialize refs early to prevent TDZ violations
  const deltaSequencesRef = useRef<Record<string, number>>({})
  const completionFetchTimersRef = useRef<Record<string, NodeJS.Timeout>>({})

  // Simple message delivery stats
  const [messageStats, setMessageStats] = useState({
    totalReceived: 0,
    successfulProcessed: 0,
    lastReceivedAt: 0
  })

  const clearCompletionTimer = useCallback((operationId: string) => {
    const timer = completionFetchTimersRef.current?.[operationId]
    if (timer) {
      try {
        clearTimeout(timer)
        delete completionFetchTimersRef.current[operationId]
      } catch (error) {
        console.warn('Failed to clear completion timer:', error)
      }
    }
  }, [])

  const applyFallbackSnapshot = useCallback((snapshot: any) => {
    if (!snapshot || typeof snapshot !== 'object' || !snapshot.operation_id) {
      return
    }

    if (operationId && snapshot.operation_id !== operationId) {
      return
    }

    setSnapshots(prev => {
      const index = prev.findIndex(op => op.operation_id === snapshot.operation_id)
      const existing = index >= 0 ? prev[index] : undefined
      const normalized = normalizeFetchedSnapshot(snapshot, existing)

      if (!normalized) {
        return prev
      }

      if (index >= 0) {
        const unchanged =
          existing &&
          existing.status === normalized.status &&
          existing.progress === normalized.progress &&
          existing.updated_at === normalized.updated_at

        if (unchanged) {
          return prev
        }

        const updated = [...prev]
        updated[index] = normalized
        return updated
      }

      if (includeHistory) {
        return [...prev, normalized]
      }

      return [normalized]
    })
  }, [includeHistory, operationId])

  const fetchSnapshotFallback = useCallback(async (targetOperationId: string) => {
    try {
      const response = await fetch(`/api/operations/${targetOperationId}/status`, {
        headers: {
          'Content-Type': 'application/json'
        }
      })

      if (!response.ok) {
        throw new Error(`Failed to fetch operation ${targetOperationId}: ${response.status}`)
      }

      const data = await response.json()
      applyFallbackSnapshot(data)
    } catch (error) {
      if (debug) {
        console.error('[useOperationSnapshots] Fallback snapshot fetch failed', {
          operationId: targetOperationId,
          error
        })
      }
    } finally {
      clearCompletionTimer(targetOperationId)
    }
  }, [applyFallbackSnapshot, clearCompletionTimer, debug])

  const scheduleSnapshotFallback = useCallback((targetOperationId: string) => {
    if (completionFetchTimersRef.current[targetOperationId]) {
      return
    }

    completionFetchTimersRef.current[targetOperationId] = setTimeout(() => {
      fetchSnapshotFallback(targetOperationId)
    }, 2000)
  }, [fetchSnapshotFallback])

  useEffect(() => {
    const unsubscribe = subscribe('operation:snapshot', (data: any) => {
      const now = Date.now()

      try {
        // Update message stats
        setMessageStats(prev => ({
          ...prev,
          totalReceived: prev.totalReceived + 1,
          lastReceivedAt: now,
          successfulProcessed: prev.successfulProcessed + 1
        }))

        // Enhanced logging for completion tracking
        const isCompletionEvent = data?.status === 'completed' || data?.status === 'failed'
        const hasSteps = Array.isArray(data?.steps) && data.steps.length > 0
        const completedStepsCount = hasSteps ? data.steps.filter((step: any) => step.status === 'completed').length : 0

        // Only log in development when explicitly enabled via localStorage
        const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
        if (isDebugMode || isCompletionEvent) {
          console.log(`[useOperationSnapshots] ${isCompletionEvent ? '🎯 COMPLETION EVENT' : '📨 MESSAGE RECEIVED'}:`, {
            timestamp: new Date().toISOString(),
            operationId: data?.operation_id,
            status: data?.status,
            progress: data?.progress,
            currentStep: data?.current_step,
            message: data?.message,
            stepsCount: hasSteps ? data.steps.length : 0,
            completedStepsCount,
            isCompletionEvent,
            hasSteps,
            stepDetails: hasSteps ? data.steps.map((step: any) => ({
              id: step.id,
              status: step.status,
              progress: step.progress,
              message: step.message
            })) : undefined
          })
        }

        // Special logging for stuck step detection
        if (data?.status === 'running' && data?.current_step) {
          const runningStep = hasSteps ? data.steps.find((step: any) => step.id === data.current_step) : null
          if (runningStep && runningStep.status === 'completed') {
            // Only log in development when explicitly enabled via localStorage
            const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
            if (isDebugMode) {
              console.warn('[useOperationSnapshots] ⚠️ STUCK STEP DETECTED:', {
                timestamp: new Date().toISOString(),
                operationId: data.operation_id,
                currentStep: data.current_step,
                stepStatus: runningStep.status,
                stepProgress: runningStep.progress,
                message: 'Current step is marked as completed but operation still running'
              })
            }
          }
        }

        // Skip invalid data early
        if (!data || !data.operation_id) {
          // Only log in development when explicitly enabled via localStorage
          const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
          if (isDebugMode) {
            console.warn('[useOperationSnapshots] Invalid operation data:', data)
          }
          return
        }

        if (isTerminalStatus(data.status)) {
          clearCompletionTimer(data.operation_id)
        }

        // Filter by operation ID if specified
        if (operationId && data.operation_id !== operationId) {
          if (debug) {
            console.debug('[useOperationSnapshots] Filtered out operation:', {
              operationId: data.operation_id,
              requestedOperationId: operationId
            })
          }
          return
        }

        // Normalize and update snapshots
        setSnapshots(prev => {
          // Find existing snapshot
          const index = prev.findIndex(op => op.operation_id === data.operation_id)
          const existing = index >= 0 ? prev[index] : undefined

          const snapshotTimestamp = getTimestampValue(
            data.updated_at ||
            data.metadata?.updated_at ||
            data.metadata?.pipeline_summary?.updated_at ||
            Date.now()
          )

          if (existing && !shouldApplySnapshot(existing, snapshotTimestamp)) {
            return prev
          }

          const hasNewSteps = Array.isArray(data.steps)
          const mergedSteps = hasNewSteps
            ? mergeSteps(existing?.steps ?? [], data.steps)
            : (existing?.steps ?? [])

          const existingMetadata =
            existing && existing.metadata && typeof existing.metadata === 'object'
              ? existing.metadata
              : {}
          let incomingMetadata =
            data && data.metadata && typeof data.metadata === 'object'
              ? { ...data.metadata }
              : undefined

          if (incomingMetadata) {
            if (data.pipeline_summary && incomingMetadata.pipeline_summary == null) {
              incomingMetadata.pipeline_summary = data.pipeline_summary
            }
            if (data.stage_timeline && incomingMetadata.stage_timeline == null) {
              incomingMetadata.stage_timeline = data.stage_timeline
            }
          } else if (data.pipeline_summary || data.stage_timeline) {
            // Ensure downstream consumers still receive pipeline context even when metadata was missing
            incomingMetadata = {
              ...(data.pipeline_summary ? { pipeline_summary: data.pipeline_summary } : {}),
              ...(data.stage_timeline ? { stage_timeline: data.stage_timeline } : {})
            }
          }
          const mergedMetadata = incomingMetadata
            ? { ...existingMetadata, ...incomingMetadata }
            : existingMetadata

          const activeStageForMetadata =
            mergedSteps.find((step: any) => step?.status === 'running') ||
            mergedSteps.find((step: any) => step?.status === 'pending') ||
            mergedSteps[mergedSteps.length - 1]

          const canonicalOperationMetadata = normalizeOperationMetadata(
            mergedMetadata,
            activeStageForMetadata?.metadata
          )

          const hasErrorField = Object.prototype.hasOwnProperty.call(data, 'error')
          const hasMessageField = Object.prototype.hasOwnProperty.call(data, 'message')
          const hasStartTimeField = Object.prototype.hasOwnProperty.call(data, 'start_time')
          const hasEndTimeField = Object.prototype.hasOwnProperty.call(data, 'end_time')
          const hasStatusField = Object.prototype.hasOwnProperty.call(data, 'status')
          const hasProgressField = Object.prototype.hasOwnProperty.call(data, 'progress')
          const hasUpdatedAtField = Object.prototype.hasOwnProperty.call(data, 'updated_at')

          // Preserve previous snapshot information when new messages omit optional fields.
          // This prevents momentary "card flicker" in the UI when websocket payloads do not
          // include expanded pipeline step data for every update.
          const normalizedSnapshot = {
            ...existing,
            operation_id: data.operation_id,
            status: (hasStatusField && typeof data.status === 'string' && data.status.length > 0)
              ? data.status
              : (existing?.status ?? 'pending'),
            progress: (hasProgressField && typeof data.progress === 'number')
              ? data.progress
              : (typeof existing?.progress === 'number' ? existing.progress : 0),
            message: hasMessageField
              ? (typeof data.message === 'string' ? data.message : data.message ?? '')
              : (existing?.message ?? ''),
            error: hasErrorField ? data.error : (existing?.error ?? null),
            start_time: hasStartTimeField ? data.start_time : (existing?.start_time ?? null),
            end_time: hasEndTimeField ? data.end_time : (existing?.end_time ?? null),
            steps: mergedSteps,
            metadata: canonicalOperationMetadata,
            updated_at: snapshotTimestamp > 0
              ? new Date(snapshotTimestamp).toISOString()
              : (existing?.updated_at ?? new Date().toISOString())
          }

          if (index >= 0) {
            // Enhanced logging for state transitions
            const statusChanged = existing?.status !== normalizedSnapshot.status
            const progressChanged = existing?.progress !== normalizedSnapshot.progress
            const stepsChanged = JSON.stringify(existing?.steps) !== JSON.stringify(mergedSteps)
            const metadataChanged =
              JSON.stringify(existing?.metadata ?? {}) !== JSON.stringify(canonicalOperationMetadata ?? {})

            if (!statusChanged && !progressChanged && !stepsChanged && !metadataChanged) {
              return prev
            }

            const updated = [...prev]
            updated[index] = normalizedSnapshot

            // Only log in development when explicitly enabled via localStorage
            const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
            if (isDebugMode || isCompletionEvent || statusChanged) {
              console.log(`[useOperationSnapshots] ${isCompletionEvent ? '🎯 FINAL STATE UPDATE' : statusChanged ? '🔄 STATUS TRANSITION' : '📝 SNAPSHOT UPDATE'}:`, {
                timestamp: new Date().toISOString(),
                operationId: data.operation_id,
                oldStatus: existing?.status,
                newStatus: normalizedSnapshot.status,
                oldProgress: existing?.progress,
                newProgress: normalizedSnapshot.progress,
                statusChanged,
                progressChanged,
                stepsChanged,
                metadataChanged,
                message: normalizedSnapshot.message,
                isCompletionEvent,
                totalSnapshots: updated.length
              })
            }

            // Temporary instrumentation for index extraction stall investigation (debug only)
            if (debug) {
              const hasIndicesStep = normalizedSnapshot.steps?.some((step: any) => step.step_id === 'indices')
              if (hasIndicesStep) {
                const indicesStep = normalizedSnapshot.steps?.find((step: any) => step.step_id === 'indices')
                console.log(`[INDICES-STAGE-INSTRUMENTATION] Raw payload analysis:`, {
                  timestamp: new Date().toISOString(),
                  operationId: data.operation_id,
                  rawPayload: data,
                  indicesStep: indicesStep,
                  indicesStatus: indicesStep?.status,
                  indicesProgress: indicesStep?.progress,
                  allStepsStatus: normalizedSnapshot.steps?.map((s: any) => ({
                    id: s.step_id,
                    status: s.status,
                    progress: s.progress
                  })),
                  normalizationApplied: {
                    status: normalizedSnapshot.status,
                    progress: normalizedSnapshot.progress,
                    message: normalizedSnapshot.message
                  }
                })
              }
            }

            return updated
          } else {
            // Add new snapshot
            if (includeHistory) {
              // Only log in development when explicitly enabled via localStorage
              const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
              if (isDebugMode && isCompletionEvent) {
                console.log(`[useOperationSnapshots] 🎯 NEW COMPLETION SNAPSHOT:`, {
                  timestamp: new Date().toISOString(),
                  operationId: data.operation_id,
                  status: normalizedSnapshot.status,
                  progress: normalizedSnapshot.progress,
                  totalSnapshots: prev.length + 1
                })
              }
              return [...prev, normalizedSnapshot]
            } else {
              // Replace all with just the latest
              // Only log in development when explicitly enabled via localStorage
              const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
              if (isDebugMode && isCompletionEvent) {
                console.log(`[useOperationSnapshots] 🎯 REPLACED WITH COMPLETION:`, {
                  timestamp: new Date().toISOString(),
                  operationId: data.operation_id,
                  status: normalizedSnapshot.status,
                  progress: normalizedSnapshot.progress,
                  oldSnapshotCount: prev.length
                })
              }
              return [normalizedSnapshot]
            }
          }
        })
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error processing snapshot'

        // Only log in development when explicitly enabled via localStorage
        const isDebugMode = debug && typeof window !== 'undefined' && window.localStorage?.getItem('debug_websocket') === 'true'
        if (isDebugMode) {
          console.error('[useOperationSnapshots] ❌ SNAPSHOT PROCESSING ERROR:', {
            timestamp: new Date().toISOString(),
            operationId: data?.operation_id,
            error: errorMessage,
            errorObject: error,
            inputData: data ? {
              operation_id: data.operation_id,
              status: data.status,
              progress: data.progress,
              hasSteps: Array.isArray(data?.steps),
              stepsCount: Array.isArray(data?.steps) ? data.steps.length : 0
            } : 'null data',
            messageStats: messageStats,
            connectionStatus: connectionStatus
          })
        }
      }
    })

    return unsubscribe
  }, [subscribe, operationId, includeHistory, debug, connectionStatus, isConnected])

  useEffect(() => {
    // Initialize deltaSequencesRef if not already initialized
    if (!deltaSequencesRef.current) {
      deltaSequencesRef.current = {}
    }

    const unsubscribe = subscribe('operation:delta', (delta: OperationDeltaEvent) => {
      // Add safety check for delta object and required properties
      if (!delta || typeof delta !== 'object' || !delta.operation_id) {
        return
      }

      if (operationId && delta.operation_id !== operationId) {
        return
      }

      // Additional safety check for deltaSequencesRef.current
      if (!deltaSequencesRef.current) {
        return
      }

      const stageKey = `${delta.operation_id}:${deriveStageIdFromDelta(delta)}`
      if (typeof delta.sequence === 'number') {
        // Safe access with default fallback
        const lastSeq = deltaSequencesRef.current[stageKey]
        if (lastSeq !== undefined && delta.sequence <= lastSeq) {
          if (debug) {
            console.log('[useOperationSnapshots] Ignoring older delta:', {
              operationId: delta.operation_id,
              stageId: deriveStageIdFromDelta(delta),
              lastSequence: lastSeq,
              deltaSequence: delta.sequence,
              timestamp: delta.updated_at
            })
          }
          return
        }
        // Safe assignment with type checking
        try {
          deltaSequencesRef.current[stageKey] = delta.sequence
        } catch (error) {
          console.warn('Failed to update delta sequence:', error)
        }
      }

      const now = Date.now()
      setMessageStats(prev => ({
        ...prev,
        totalReceived: prev.totalReceived + 1,
        successfulProcessed: prev.successfulProcessed + 1,
        lastReceivedAt: now
      }))

      setSnapshots(prev => applyDeltaToSnapshots(prev, delta, includeHistory))

      if (isTerminalStatus(delta.status)) {
        scheduleSnapshotFallback(delta.operation_id)
      }
    })

    return unsubscribe
  }, [subscribe, operationId, includeHistory, scheduleSnapshotFallback])

  useEffect(() => {
    const unsubscribe = subscribe('scraping:telemetry', (data: ScrapingTelemetry) => {
      if (!data || !data.operation_id) {
        return
      }

      // Enforce presence of canonical fields; if missing, mark telemetry_missing so UI can render error
      const normalized: ScrapingTelemetry = {
        ...data,
        trading_days_total: typeof data.trading_days_total === 'number' ? data.trading_days_total : 0,
        trading_days_completed: typeof data.trading_days_completed === 'number' ? data.trading_days_completed : 0,
        holidays_detected: typeof data.holidays_detected === 'number' ? data.holidays_detected : (Array.isArray(data.skipped_files) ? data.skipped_files.length : 0),
        telemetry_missing: !(
          typeof data.trading_days_total === 'number' &&
          typeof data.trading_days_completed === 'number' &&
          typeof (data.holidays_detected ?? (Array.isArray(data.skipped_files) ? data.skipped_files.length : undefined)) === 'number'
        )
      }

      setScrapingTelemetry(prev => ({
        ...prev,
        [data.operation_id]: normalized
      }))
    })
    return unsubscribe
  }, [subscribe])

  return {
    // Core data - simplified interface
    snapshots,
    connected: isConnected,
    connectionStatus,
    error: lastError,

    // Clear method for compatibility
    clearSnapshots: () => setSnapshots([]),

    // Simple message stats
    messageStats,
    getMessageSuccessRate: () => {
      const { totalReceived, successfulProcessed } = messageStats
      return totalReceived > 0 ? (successfulProcessed / totalReceived) * 100 : 100
    },
    getTimeSinceLastMessage: () => {
      const { lastReceivedAt } = messageStats
      return lastReceivedAt > 0 ? Date.now() - lastReceivedAt : 0
    },
    getScrapingTelemetry: (operationId: string) => scrapingTelemetry[operationId]
  }
}

/**
 * Legacy hook for backward compatibility - simplified interface
 */
export function useAllOperationUpdates() {
  const {
    // Core data
    snapshots,
    connected,
    connectionStatus,
    error,
    clearSnapshots,

    // Simple stats
    messageStats,
    getMessageSuccessRate,
    getTimeSinceLastMessage,
    getScrapingTelemetry
  } = useOperationSnapshots({
    includeHistory: true
  })

  return {
    operations: snapshots,
    connected,
    connectionStatus,
    error: error,
    clearSnapshots,

    // Simple stats
    messageStats,
    getMessageSuccessRate,
    getTimeSinceLastMessage,
    getScrapingTelemetry,

    // Simplified diagnostics for UI components
    connectionDiagnostics: {
      quality: {
        signalStrength: connected ? 100 : 0,
        latency: 0,
        stability: connected ? 100 : 0,
        packetLoss: connected ? 0 : 100,
        throughput: 0
      },
      events: []
    },
    lastConnectionHealth: {
      healthy: connected && (messageStats.totalReceived > 0 || connectionStatus === 'connected'),
      checkedAt: Date.now(),
      issues: (() => {
        if (!connected) {
          return ['WebSocket disconnected']
        }
        if (messageStats.totalReceived === 0) {
          return ['Awaiting first operation update']
        }
        return []
      })()
    },
    messageDeliveryStats: messageStats,
    isConnectionHealthy: () => connected && (messageStats.totalReceived > 0 || connectionStatus === 'connected'),
    getConnectionIssues: () => {
      if (!connected) {
        return ['WebSocket disconnected']
      }
      if (messageStats.totalReceived === 0) {
        return ['No operation updates received yet']
      }
      return []
    }
  }
}

// Legacy aliases
export const usePipelineUpdates = useAllOperationUpdates
export const useMarketUpdates = () => {
  const { subscribe } = useWebSocket()
  return { subscribe: (handler: any) => subscribe('market_update', handler) }
}
export const useWebSocketEvent = useWebSocket
