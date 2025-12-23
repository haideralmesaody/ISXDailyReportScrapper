'use client'

import React from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  AlertTriangle,
  WifiOff,
  RefreshCw,
  Database,
  Clock,
  Info,
  CheckCircle,
  Lightbulb,
  Settings,
  HelpCircle
} from 'lucide-react'

interface RecoveryState {
  mode: 'none' | 'retry' | 'offline' | 'degraded'
  retryCount: number
  maxRetries: number
  nextRetryIn: number
  lastAttempt: number
}

interface OfflineState {
  isOffline: boolean
  cachedOperations: any[]
  cachedOperationTypes: any[]
  lastSyncTime: number
  queueOperations: any[]
}

interface RecoveryGuidanceProps {
  error?: string | null
  recoveryAction?: string | null
  recoveryState: RecoveryState
  offlineState: OfflineState
  onManualRetry?: () => void
  onClearCache?: () => void
  connected: boolean
  connectionStatus: string
}

export function RecoveryGuidance({
  error,
  recoveryAction,
  recoveryState,
  offlineState,
  onManualRetry,
  onClearCache,
  connected,
  connectionStatus
}: RecoveryGuidanceProps) {
  // If no recovery state, don't show guidance
  if (recoveryState.mode === 'none' && connected && !error) {
    return null
  }

  const getRetryProgress = () => {
    return (recoveryState.retryCount / recoveryState.maxRetries) * 100
  }

  const formatTimeRemaining = (milliseconds: number) => {
    const seconds = Math.ceil(milliseconds / 1000)
    return `${seconds}s`
  }

  const formatTimeSince = (timestamp: number) => {
    const now = Date.now()
    const diff = now - timestamp

    if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    return `${Math.floor(diff / 3600000)}h ago`
  }

  const getGuidanceTitle = () => {
    if (offlineState.isOffline) return 'Working Offline'
    if (recoveryState.mode === 'retry') return 'Attempting Reconnection'
    if (recoveryState.mode === 'degraded') return 'Limited Functionality'
    if (error) return 'Connection Issue'
    return 'System Status'
  }

  const getGuidanceDescription = () => {
    if (offlineState.isOffline) {
      return 'The system is operating in offline mode using cached data. Operations will be queued and executed when the connection is restored.'
    }
    if (recoveryState.mode === 'retry') {
      return `Attempting to restore connection (attempt ${recoveryState.retryCount} of ${recoveryState.maxRetries}).`
    }
    if (recoveryState.mode === 'degraded') {
      return 'The system is running with limited functionality due to connection issues.'
    }
    if (error) {
      return 'There\'s an issue with the connection. The system will attempt to recover automatically.'
    }
    return 'System monitoring connection status.'
  }

  const getActions = () => {
    const actions = []

    if (recoveryState.mode === 'retry' && !connected) {
      actions.push({
        label: 'Force Retry',
        icon: RefreshCw,
        onClick: onManualRetry,
        variant: 'default' as const
      })
    }

    if (offlineState.isOffline) {
      actions.push({
        label: 'Clear Cache & Reload',
        icon: Database,
        onClick: onClearCache,
        variant: 'outline' as const
      })
    }

    if (error) {
      actions.push({
        label: 'Try Again',
        icon: RefreshCw,
        onClick: onManualRetry,
        variant: 'default' as const
      })
    }

    return actions
  }

  const getStatusColor = () => {
    if (offlineState.isOffline) return 'text-orange-600'
    if (recoveryState.mode === 'retry') return 'text-yellow-600'
    if (recoveryState.mode === 'degraded') return 'text-red-600'
    return 'text-blue-600'
  }

  const getIcon = () => {
    if (offlineState.isOffline) return WifiOff
    if (recoveryState.mode === 'retry') return RefreshCw
    if (recoveryState.mode === 'degraded') return AlertTriangle
    if (error) return AlertTriangle
    return Info
  }

  const GuidanceIcon = getIcon()

  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex items-center space-x-3">
          <div className={`p-2 rounded-full bg-gray-100`}>
            <GuidanceIcon className={`h-5 w-5 ${getStatusColor()} ${recoveryState.mode === 'retry' ? 'animate-spin' : ''}`} />
          </div>
          <div>
            <CardTitle className="text-lg">{getGuidanceTitle()}</CardTitle>
            <p className="text-sm text-muted-foreground">{getGuidanceDescription()}</p>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Error Message */}
        {error && (
          <div className="border-l-4 border-red-400 bg-red-50 p-3 rounded">
            <div className="flex items-start space-x-2">
              <AlertTriangle className="h-5 w-5 text-red-600 mt-0.5" />
              <div>
                <h5 className="font-medium text-red-800">Error</h5>
                <p className="text-sm text-red-700 mt-1">{error}</p>
                {recoveryAction && (
                  <p className="text-sm text-red-600 mt-1">
                    <strong>Suggested action:</strong> {recoveryAction}
                  </p>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Retry Progress */}
        {recoveryState.mode === 'retry' && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span>Reconnection Progress</span>
              <span className="font-medium">
                {recoveryState.retryCount}/{recoveryState.maxRetries} attempts
              </span>
            </div>
            <Progress value={getRetryProgress()} className="h-2" />
            {recoveryState.nextRetryIn > 0 && (
              <p className="text-xs text-muted-foreground">
                Next retry in {formatTimeRemaining(recoveryState.nextRetryIn)}
              </p>
            )}
          </div>
        )}

        {/* Offline Status */}
        {offlineState.isOffline && (
          <div className="space-y-3">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <Database className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Cached Operations</span>
                </div>
                <p className="text-2xl font-bold text-orange-600">
                  {offlineState.cachedOperations.length}
                </p>
              </div>

              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <Clock className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Last Sync</span>
                </div>
                <p className="text-sm text-muted-foreground">
                  {formatTimeSince(offlineState.lastSyncTime)}
                </p>
              </div>

              <div className="space-y-1">
                <div className="flex items-center space-x-2">
                  <RefreshCw className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">Queued</span>
                </div>
                <p className="text-2xl font-bold text-blue-600">
                  {offlineState.queueOperations.length}
                </p>
              </div>
            </div>

            <div className="border-l-4 border-orange-400 bg-orange-50 p-3 rounded">
              <div className="flex items-start space-x-2">
                <Lightbulb className="h-5 w-5 text-orange-600 mt-0.5" />
                <div>
                  <h5 className="font-medium text-orange-800">Offline Mode Active</h5>
                  <p className="text-sm text-orange-700 mt-1">
                    You can view cached data and queue new operations. All queued operations will execute automatically when the connection is restored.
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Degraded Mode */}
        {recoveryState.mode === 'degraded' && (
          <div className="space-y-3">
            <div className="border-l-4 border-yellow-400 bg-yellow-50 p-3 rounded">
              <div className="flex items-start space-x-2">
                <Settings className="h-5 w-5 text-yellow-600 mt-0.5" />
                <div>
                  <h5 className="font-medium text-yellow-800">Limited Functionality</h5>
                  <p className="text-sm text-yellow-700 mt-1">
                    The system is running with reduced capabilities due to connection issues. Some features may not be available.
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Available Actions */}
        {getActions().length > 0 && (
          <div className="space-y-3">
            <h5 className="font-medium">Available Actions</h5>
            <div className="flex space-x-2">
              {getActions().map((action, index) => {
                const ActionIcon = action.icon
                return (
                  <Button
                    key={index}
                    variant={action.variant}
                    size="sm"
                    onClick={action.onClick}
                    className="flex items-center space-x-2"
                  >
                    <ActionIcon className="h-4 w-4" />
                    <span>{action.label}</span>
                  </Button>
                )
              })}
            </div>
          </div>
        )}

        {/* Help Section */}
        <div className="border-t pt-4">
          <div className="flex items-start space-x-2">
            <HelpCircle className="h-5 w-5 text-muted-foreground mt-0.5" />
            <div className="flex-1">
              <h5 className="font-medium text-sm">Need Help?</h5>
              <p className="text-xs text-muted-foreground mt-1">
                If connection issues persist, try refreshing the page or checking your internet connection.
                You can also clear the cache to start fresh.
              </p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}