'use client'

import { useState, useEffect, useCallback, useRef } from 'react'

interface OperationSnapshot {
  operation_id: string
  status: string
  progress: number
  message: string
  error?: string
  steps?: any[]
  metadata?: any
  updated_at: string
}

interface UseOperationPollingOptions {
  enabled?: boolean
  interval?: number
  operationId?: string
}

export function useOperationPolling(options: UseOperationPollingOptions = {}) {
  const {
    enabled = true,
    interval = 2000, // 2 seconds default
    operationId
  } = options

  const [snapshots, setSnapshots] = useState<OperationSnapshot[]>([])
  const [isPolling, setIsPolling] = useState(false)
  const [lastError, setLastError] = useState<string | null>(null)
  const [pollCount, setPollCount] = useState(0)

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  // Enhanced state tracking for stuck operation detection
  const [snapshotHistory, setSnapshotHistory] = useState<Map<string, OperationSnapshot[]>>(new Map())
  const debug = process.env.NODE_ENV === 'development'

  const fetchOperations = useCallback(async () => {
    if (!enabled) return

    try {
      // Cancel previous request if still pending
      if (abortControllerRef.current) {
        abortControllerRef.current.abort()
      }

      abortControllerRef.current = new AbortController()

      const startTime = Date.now()
      if (debug) {
        console.log(`[useOperationPolling] 🔄 POLLING REQUEST #${pollCount + 1}:`, {
          timestamp: new Date().toISOString(),
          operationId,
          interval,
          enabled
        })
      }

      const response = await fetch('/api/operations', {
        signal: abortControllerRef.current.signal,
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const data = await response.json()
      const requestDuration = Date.now() - startTime

      if (data.operations && Array.isArray(data.operations)) {
        const newSnapshots = operationId
          ? data.operations.filter((op: any) => op.operation_id === operationId)
          : data.operations

        // Enhanced logging for completion detection
        const completedOperations = newSnapshots.filter((op: any) => op.status === 'completed')
        const stuckOperations = newSnapshots.filter((op: any) => {
          if (op.status !== 'running') return false
          const history = snapshotHistory.get(op.operation_id) || []
          if (history.length < 2) return false

          const lastState = history.at(-1)
          const previousState = history.at(-2)
          if (!lastState || !previousState) return false

          const unchangedBetweenPolls =
            lastState.progress === previousState.progress &&
            lastState.status === previousState.status

          return unchangedBetweenPolls &&
                 lastState.progress === op.progress &&
                 lastState.status === op.status &&
                 (Date.now() - new Date(lastState.updated_at).getTime()) > 30000
        })

        if (completedOperations.length > 0) {
          console.log(`[useOperationPolling] 🎯 FALLBACK COMPLETIONS DETECTED:`, {
            timestamp: new Date().toISOString(),
            completedCount: completedOperations.length,
            operations: completedOperations.map((op: any) => ({
              operationId: op.operation_id,
              status: op.status,
              progress: op.progress,
              stepsCompleted: op.steps?.filter((step: any) => step.status === 'completed').length,
              totalSteps: op.steps?.length
            })),
            requestDuration,
            source: 'fallback_polling'
          })
        }

        if (stuckOperations.length > 0) {
          console.warn(`[useOperationPolling] ⚠️ STUCK OPERATIONS DETECTED BY POLLING:`, {
            timestamp: new Date().toISOString(),
            stuckCount: stuckOperations.length,
            operations: stuckOperations.map((op: any) => ({
              operationId: op.operation_id,
              status: op.status,
              progress: op.progress,
              lastUpdate: op.updated_at
            })),
            message: 'Operations detected as stuck via polling - WebSocket may have missed updates'
          })
        }

        setSnapshots(prev => {
          // Update existing snapshots or add new ones
          const updated = [...prev]
          newSnapshots.forEach((newSnapshot: OperationSnapshot) => {
            const index = updated.findIndex(op => op.operation_id === newSnapshot.operation_id)
            if (index >= 0) {
              const oldSnapshot = updated[index]
              updated[index] = newSnapshot

              // Log state changes detected by polling
              if (oldSnapshot && debug && (oldSnapshot.status !== newSnapshot.status || oldSnapshot.progress !== newSnapshot.progress)) {
                console.log(`[useOperationPolling] 📊 STATE CHANGE DETECTED BY POLLING:`, {
                  timestamp: new Date().toISOString(),
                  operationId: newSnapshot.operation_id,
                  oldStatus: oldSnapshot.status,
                  newStatus: newSnapshot.status,
                  oldProgress: oldSnapshot.progress,
                  newProgress: newSnapshot.progress,
                  requestDuration,
                  message: 'Polling detected state change that WebSocket may have missed'
                })
              }
            } else {
              updated.push(newSnapshot)

              if (debug) {
                console.log(`[useOperationPolling] ➕ NEW OPERATION DETECTED BY POLLING:`, {
                  timestamp: new Date().toISOString(),
                  operationId: newSnapshot.operation_id,
                  status: newSnapshot.status,
                  progress: newSnapshot.progress
                })
              }
            }
          })

          return updated
        })

        // Update snapshot history for stuck detection
        setSnapshotHistory(prev => {
          const newMap = new Map(prev)
          newSnapshots.forEach((snapshot: OperationSnapshot) => {
            const history = prev.get(snapshot.operation_id) || []
            const newHistory = [...history, snapshot].slice(-10) // Keep last 10
            newMap.set(snapshot.operation_id, newHistory)
          })
          return newMap
        })

        setLastError(null)

        if (debug) {
          console.log(`[useOperationPolling] ✅ POLLING SUCCESS:`, {
            timestamp: new Date().toISOString(),
            requestDuration,
            operationsReceived: newSnapshots.length,
            totalOperations: snapshots.length,
            pollCount: pollCount + 1
          })
        }
      }
    } catch (error) {
      if (error instanceof Error && error.name !== 'AbortError') {
        setLastError(error.message)
        console.error('[useOperationPolling] ❌ POLLING ERROR:', {
          timestamp: new Date().toISOString(),
          error: error.message,
          operationId,
          pollCount: pollCount + 1,
          enabled
        })
      }
    } finally {
      setPollCount(prev => prev + 1)
    }
  }, [enabled, operationId, pollCount, snapshots.length, snapshotHistory, debug])

  const startPolling = useCallback(() => {
    if (!enabled || pollIntervalRef.current) return

    setIsPolling(true)

    // Initial fetch immediately
    fetchOperations()

    // Then set up interval
    pollIntervalRef.current = setInterval(() => {
      fetchOperations()
    }, interval)
  }, [enabled, interval, fetchOperations])

  const stopPolling = useCallback(() => {
    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current)
      pollIntervalRef.current = null
    }

    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
      abortControllerRef.current = null
    }

    setIsPolling(false)
  }, [])

  const clearSnapshots = useCallback(() => {
    setSnapshots([])
  }, [])

  // Auto-start polling when enabled
  useEffect(() => {
    if (enabled) {
      startPolling()
    } else {
      stopPolling()
    }

    return () => {
      stopPolling()
    }
  }, [enabled, startPolling, stopPolling])

  return {
    snapshots,
    isPolling,
    lastError,
    pollCount,
    clearSnapshots,
    startPolling,
    stopPolling,
    // Compatibility with WebSocket hook interface
    connected: isPolling,
    connectionStatus: isPolling ? 'connected' : 'disconnected',
    error: lastError,
    messageStats: {
      totalReceived: pollCount,
      successfulProcessed: snapshots.length,
      lastReceivedAt: Date.now()
    },
    getMessageSuccessRate: () => pollCount > 0 ? (snapshots.length / pollCount) * 100 : 0,
    getTimeSinceLastMessage: () => 0 // Always current for polling
  }
}
