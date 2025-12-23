'use client'

import React from 'react'
import { Wifi, WifiOff, AlertTriangle, RefreshCw } from 'lucide-react'
import { useConnectionStatus } from '@/lib/hooks/use-websocket'
import { cn } from '@/lib/utils'

interface ConnectionStatusIndicatorProps {
  className?: string
  showDetails?: boolean
  compact?: boolean
}

export function ConnectionStatusIndicator({
  className,
  showDetails = false,
  compact = false
}: ConnectionStatusIndicatorProps) {
  const { status, isConnected, health, isHealthy, getTimeSinceLastConnection } = useConnectionStatus()

  const getStatusIcon = () => {
    switch (status) {
      case 'connected':
        return <Wifi className="h-4 w-4 text-green-500" />
      case 'connecting':
      case 'reconnecting':
        return <RefreshCw className="h-4 w-4 text-yellow-500 animate-spin" />
      case 'error':
        return <AlertTriangle className="h-4 w-4 text-red-500" />
      default:
        return <WifiOff className="h-4 w-4 text-gray-400" />
    }
  }

  const getStatusText = () => {
    switch (status) {
      case 'connected':
        return 'Live Updates'
      case 'connecting':
        return 'Connecting...'
      case 'reconnecting':
        return 'Reconnecting...'
      case 'error':
        return 'Connection Error'
      default:
        return 'Disconnected'
    }
  }

  const getStatusColor = () => {
    switch (status) {
      case 'connected':
        return 'text-green-600 bg-green-50 border-green-200'
      case 'connecting':
      case 'reconnecting':
        return 'text-yellow-600 bg-yellow-50 border-yellow-200'
      case 'error':
        return 'text-red-600 bg-red-50 border-red-200'
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200'
    }
  }

  if (compact) {
    return (
      <div className={cn("flex items-center gap-2 text-xs", className)}>
        {getStatusIcon()}
        <span className={cn("font-medium", getStatusColor().split(' ')[0])}>
          {getStatusText()}
        </span>
      </div>
    )
  }

  const timeSinceLastConnection = getTimeSinceLastConnection()
  const formatDuration = (ms: number) => {
    if (ms === Infinity) return 'Never'
    const seconds = Math.floor(ms / 1000)
    if (seconds < 60) return `${seconds}s ago`
    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return `${minutes}m ago`
    const hours = Math.floor(minutes / 60)
    return `${hours}h ago`
  }

  return (
    <div className={cn(
      "flex items-center gap-2 px-3 py-1.5 rounded-md border text-sm",
      getStatusColor(),
      className
    )}>
      {getStatusIcon()}

      <div className="flex flex-col">
        <span className="font-medium">{getStatusText()}</span>

        {showDetails && (
          <div className="text-xs opacity-75">
            {isConnected && (
              <div className="flex items-center gap-4">
                <span>Connected</span>
                {health.totalDisconnections > 0 && (
                  <span>Reconnects: {health.totalDisconnections}</span>
                )}
                {health.averageReconnectTime > 0 && (
                  <span>Avg: {Math.round(health.averageReconnectTime / 1000)}s</span>
                )}
              </div>
            )}

            {!isConnected && health.lastConnectedAt && (
              <span>Last: {formatDuration(timeSinceLastConnection)}</span>
            )}
          </div>
        )}
      </div>

      {/* Health indicator dot */}
      <div className={cn(
        "w-2 h-2 rounded-full",
        isHealthy ? "bg-green-500" :
        status === 'connecting' || status === 'reconnecting' ? "bg-yellow-500" :
        "bg-red-500"
      )} />
    </div>
  )
}

export default ConnectionStatusIndicator