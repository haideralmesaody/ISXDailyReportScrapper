'use client'

import { useEffect, useState } from 'react'
import { useOperationSnapshots } from './use-websocket'
import { useOperationPolling } from './use-operation-polling'

interface UseOperationsWithFallbackOptions {
  operationId?: string
  includeHistory?: boolean
  pollingInterval?: number
  showConnectionStatus?: boolean
}

export function useOperationsWithFallback(options: UseOperationsWithFallbackOptions = {}) {
  const {
    operationId,
    includeHistory = true,
    pollingInterval = 2000,
    showConnectionStatus = true
  } = options

  // Initialize with explicit guards to prevent TDZ violations
  const [isUsingFallback, setIsUsingFallback] = useState(false)
  const [fallbackModeReason, setFallbackModeReason] = useState<string>('')
  const [isInitialized, setIsInitialized] = useState(false)

  // Ensure proper initialization sequence
  useEffect(() => {
    setIsInitialized(true)
  }, [])

  // Return loading state during initialization to prevent TDZ
  if (!isInitialized) {
    return {
      loading: true,
      connectionMethod: 'initializing',
      snapshots: [],
      connected: false,
      connectionStatus: 'initializing',
      error: null,
      clearSnapshots: () => {},
      messageStats: {
        totalReceived: 0,
        successfulProcessed: 0,
        lastReceivedAt: 0
      },
      getMessageSuccessRate: () => 100,
      getTimeSinceLastMessage: () => 0,
      fallbackReason: null,
      connectionStatusProps: {
        showConnectionStatus: false,
        useFallback: false,
        fallbackReason: ''
      },
      isLiveUpdates: false,
      switchToFallback: () => {},
      switchToWebSocket: () => {}
    }
  }

  // Try WebSocket first
  const wsResult = useOperationSnapshots({
    ...(operationId !== undefined ? { operationId } : {}),
    includeHistory,
    debug: process.env.NODE_ENV === 'development'
  })

  // Fallback to polling
  const pollingResult = useOperationPolling({
    enabled: isUsingFallback,
    interval: pollingInterval,
    ...(operationId !== undefined ? { operationId } : {})
  })

  // Determine if we should fall back to polling
  useEffect(() => {
    const timeout = setTimeout(() => {
      // Fall back if WebSocket hasn't received any messages after 10 seconds
      if (wsResult.messageStats.totalReceived === 0 && wsResult.connectionStatus !== 'connected') {
        setIsUsingFallback(true)
        setFallbackModeReason('WebSocket not receiving messages')
      }
    }, 10000) // 10 seconds

    // Also fall back if WebSocket is in error state
    if (wsResult.connectionStatus === 'error') {
      setIsUsingFallback(true)
      setFallbackModeReason('WebSocket connection error')
    }

    return () => clearTimeout(timeout)
  }, [wsResult.connectionStatus, wsResult.messageStats.totalReceived])

  // Switch back to WebSocket if it becomes healthy
  useEffect(() => {
    if (isUsingFallback && wsResult.connected && wsResult.messageStats.totalReceived > 0) {
      setIsUsingFallback(false)
      setFallbackModeReason('')
    }
  }, [isUsingFallback, wsResult.connected, wsResult.messageStats.totalReceived])

  // Use appropriate result
  const result = isUsingFallback ? pollingResult : wsResult

  return {
    ...result,
    // Additional metadata about connection method
    connectionMethod: isUsingFallback ? 'polling' : 'websocket',
    fallbackReason: isUsingFallback ? fallbackModeReason : null,
    // Props for the connection status component
    connectionStatusProps: {
      showConnectionStatus,
      useFallback: isUsingFallback,
      fallbackReason: fallbackModeReason
    },
    // Helper to determine if we're getting live updates
    isLiveUpdates: !isUsingFallback && result.connected,
    // Method for switching fallback mode manually
    switchToFallback: (reason: string) => {
      setIsUsingFallback(true)
      setFallbackModeReason(reason)
    },
    switchToWebSocket: () => {
      setIsUsingFallback(false)
      setFallbackModeReason('')
    }
  }
}

export default useOperationsWithFallback
