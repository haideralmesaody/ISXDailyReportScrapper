'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  Wifi,
  WifiOff,
  AlertTriangle,
  CheckCircle,
  XCircle,
  RefreshCw,
  Activity,
  Clock,
  Database,
  ChevronDown,
  ChevronUp,
  Loader2
} from 'lucide-react'

interface ConnectionDiagnosticsPanelProps {
  connected: boolean
  connectionStatus: string
  connectionDiagnostics?: any
  lastConnectionHealth?: any
  messageDeliveryStats?: any
  onManualRetry?: () => void
  onClearCache?: () => void
  recoveryState?: {
    mode: 'none' | 'retry' | 'offline' | 'degraded'
    retryCount: number
    nextRetryIn: number
  }
  offlineState?: {
    isOffline: boolean
    cachedOperations: any[]
    lastSyncTime: number
    queueOperations: any[]
  }
}

export function ConnectionDiagnosticsPanel({
  connected,
  connectionStatus,
  connectionDiagnostics,
  lastConnectionHealth,
  messageDeliveryStats,
  onManualRetry,
  onClearCache,
  recoveryState,
  offlineState
}: ConnectionDiagnosticsPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [selectedTab, setSelectedTab] = useState<'overview' | 'quality' | 'history' | 'recovery'>('overview')
  const retryCount = recoveryState?.retryCount ?? 0

  // Determine connection status color and icon
  const getConnectionStatusInfo = () => {
    if (connected && connectionStatus === 'connected') {
      const hasMessages = (messageDeliveryStats?.totalReceived || 0) > 0
      return {
        color: 'text-green-600',
        bgColor: 'bg-green-100',
        icon: CheckCircle,
        label: 'Connected',
        description: hasMessages ? 'Receiving live updates' : 'Waiting for first update'
      }
    } else if (connectionStatus === 'connecting' || connectionStatus === 'reconnecting') {
      return {
        color: 'text-yellow-600',
        bgColor: 'bg-yellow-100',
        icon: Loader2,
        label: 'Connecting',
        description: 'Establishing connection...'
      }
    } else if (offlineState?.isOffline) {
      return {
        color: 'text-orange-600',
        bgColor: 'bg-orange-100',
        icon: WifiOff,
        label: 'Offline Mode',
        description: 'Using cached data'
      }
    } else {
      return {
        color: 'text-red-600',
        bgColor: 'bg-red-100',
        icon: XCircle,
        label: 'Disconnected',
        description: 'Connection lost'
      }
    }
  }

  const statusInfo = getConnectionStatusInfo()
  const StatusIcon = statusInfo.icon

  // Format time since last message
  const formatTimeSince = (timestamp: number) => {
    const now = Date.now()
    const diff = now - timestamp

    if (diff < 1000) return 'Just now'
    if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    return `${Math.floor(diff / 3600000)}h ago`
  }

  // Get signal strength color
  const getSignalStrengthColor = (strength: number) => {
    if (strength >= 80) return 'text-green-600'
    if (strength >= 60) return 'text-yellow-600'
    if (strength >= 40) return 'text-orange-600'
    return 'text-red-600'
  }

  // Get success rate color
  const getSuccessRateColor = (rate: number) => {
    if (rate >= 95) return 'text-green-600'
    if (rate >= 80) return 'text-yellow-600'
    if (rate >= 60) return 'text-orange-600'
    return 'text-red-600'
  }

  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className={`p-2 rounded-full ${statusInfo.bgColor}`}>
              <StatusIcon className={`h-5 w-5 ${statusInfo.color} ${connectionStatus === 'connecting' ? 'animate-spin' : ''}`} />
            </div>
            <div>
              <CardTitle className="text-lg">{statusInfo.label}</CardTitle>
              <p className="text-sm text-muted-foreground">{statusInfo.description}</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            {recoveryState?.mode === 'retry' && (
              <Badge variant="outline" className="text-yellow-600">
                <RefreshCw className="h-3 w-3 mr-1 animate-spin" />
                Retrying in {Math.ceil(recoveryState.nextRetryIn / 1000)}s
              </Badge>
            )}
            {offlineState?.isOffline && (
              <Badge variant="outline" className="text-orange-600">
                <WifiOff className="h-3 w-3 mr-1" />
                Offline
              </Badge>
            )}
            <Collapsible
              open={isExpanded}
              onOpenChange={setIsExpanded}
            >
              <CollapsibleTrigger asChild>
                <Button variant="ghost" size="sm">
                  {isExpanded ? (
                    <ChevronUp className="h-4 w-4" />
                  ) : (
                    <ChevronDown className="h-4 w-4" />
                  )}
                </Button>
              </CollapsibleTrigger>
            </Collapsible>
          </div>
        </div>
      </CardHeader>

      <Collapsible open={isExpanded}>
        <CollapsibleContent>
          <CardContent className="space-y-4">
            {/* Tab Navigation */}
            <div className="flex space-x-1 border-b">
              {['overview', 'quality', 'history', 'recovery'].map((tab) => (
                <Button
                  key={tab}
                  variant={selectedTab === tab ? 'default' : 'ghost'}
                  size="sm"
                  onClick={() => setSelectedTab(tab as any)}
                  className="capitalize"
                >
                  {tab}
                </Button>
              ))}
            </div>

            {/* Overview Tab */}
            {selectedTab === 'overview' && (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center space-x-2">
                    <Activity className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Status</span>
                  </div>
                  <Badge className={statusInfo.color}>
                    {statusInfo.label}
                  </Badge>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center space-x-2">
                    <Wifi className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Signal</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className="flex-1">
                      <Progress
                        value={connectionDiagnostics?.quality?.signalStrength || 0}
                        className="h-2"
                      />
                    </div>
                    <span className={`text-sm font-medium ${getSignalStrengthColor(connectionDiagnostics?.quality?.signalStrength || 0)}`}>
                      {connectionDiagnostics?.quality?.signalStrength || 0}%
                    </span>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center space-x-2">
                    <Database className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Success Rate</span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className="flex-1">
                      <Progress
                        value={messageDeliveryStats?.totalReceived ?
                          (messageDeliveryStats.successfulProcessed / messageDeliveryStats.totalReceived) * 100 : 0}
                        className="h-2"
                      />
                    </div>
                    <span className={`text-sm font-medium ${getSuccessRateColor(
                      messageDeliveryStats?.totalReceived ?
                        (messageDeliveryStats.successfulProcessed / messageDeliveryStats.totalReceived) * 100 : 0
                    )}`}>
                      {messageDeliveryStats?.totalReceived ?
                        Math.round((messageDeliveryStats.successfulProcessed / messageDeliveryStats.totalReceived) * 100) : 0}%
                    </span>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center space-x-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Last Message</span>
                  </div>
                  <span className="text-sm text-muted-foreground">
                    {messageDeliveryStats?.lastReceivedAt ?
                      formatTimeSince(messageDeliveryStats.lastReceivedAt) :
                      'Never'
                    }
                  </span>
                </div>
              </div>
            )}

            {/* Quality Tab */}
            {selectedTab === 'quality' && connectionDiagnostics && (
              <div className="space-y-4">
                <h4 className="font-medium">Connection Quality Metrics</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm">Latency</span>
                      <span className="text-sm font-medium">{connectionDiagnostics.quality?.latency || 0}ms</span>
                    </div>
                    <Progress value={Math.min(100, (connectionDiagnostics.quality?.latency || 0) / 10)} />
                  </div>

                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm">Stability</span>
                      <span className="text-sm font-medium">{Math.round(connectionDiagnostics.quality?.stability || 0)}%</span>
                    </div>
                    <Progress value={connectionDiagnostics.quality?.stability || 0} />
                  </div>

                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm">Packet Loss</span>
                      <span className="text-sm font-medium">{Math.round(connectionDiagnostics.quality?.packetLoss || 0)}%</span>
                    </div>
                    <Progress value={connectionDiagnostics.quality?.packetLoss || 0} />
                  </div>

                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm">Throughput</span>
                      <span className="text-sm font-medium">{Math.round(connectionDiagnostics.quality?.throughput || 0)}</span>
                    </div>
                    <Progress value={Math.min(100, (connectionDiagnostics.quality?.throughput || 0) / 10)} />
                  </div>
                </div>
              </div>
            )}

            {/* History Tab */}
            {selectedTab === 'history' && (
              <div className="space-y-4">
                <h4 className="font-medium">Recent Connection Events</h4>
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {connectionDiagnostics?.events?.slice(-10).map((event: any, index: number) => (
                    <div key={index} className="flex items-center justify-between text-sm p-2 rounded border">
                      <div className="flex items-center space-x-2">
                        {event.status === 'connected' && <CheckCircle className="h-4 w-4 text-green-600" />}
                        {event.status === 'disconnected' && <XCircle className="h-4 w-4 text-red-600" />}
                        {event.status === 'error' && <AlertTriangle className="h-4 w-4 text-yellow-600" />}
                        {event.status === 'connecting' && <Loader2 className="h-4 w-4 text-blue-600 animate-spin" />}
                        <span className="capitalize">{event.status}</span>
                        {event.attempt && <span className="text-muted-foreground">(Attempt {event.attempt})</span>}
                      </div>
                      <span className="text-muted-foreground">
                        {formatTimeSince(event.timestamp)}
                      </span>
                    </div>
                  )) || (
                    <p className="text-muted-foreground text-sm">No connection events recorded</p>
                  )}
                </div>
              </div>
            )}

            {/* Recovery Tab */}
            {selectedTab === 'recovery' && (
              <div className="space-y-4">
                <h4 className="font-medium">Recovery & Cache Management</h4>

                {/* Recovery State */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Recovery Mode</span>
                    <Badge variant={recoveryState?.mode === 'none' ? 'default' : 'secondary'}>
                      {recoveryState?.mode || 'none'}
                    </Badge>
                  </div>
                  {retryCount > 0 && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Retry Attempts</span>
                      <span className="text-sm font-medium">{retryCount}</span>
                    </div>
                  )}
                </div>

                {/* Offline State */}
                {offlineState && (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Cached Operations</span>
                      <span className="text-sm font-medium">{offlineState.cachedOperations.length}</span>
                    </div>
                    {offlineState.lastSyncTime > 0 && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm">Last Sync</span>
                        <span className="text-sm text-muted-foreground">
                          {formatTimeSince(offlineState.lastSyncTime)}
                        </span>
                      </div>
                    )}
                    {offlineState.queueOperations.length > 0 && (
                      <div className="flex items-center justify-between">
                        <span className="text-sm">Queued Operations</span>
                        <Badge variant="outline">{offlineState.queueOperations.length}</Badge>
                      </div>
                    )}
                  </div>
                )}

                {/* Action Buttons */}
                <div className="flex space-x-2 pt-2">
                  {onManualRetry && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={onManualRetry}
                      disabled={recoveryState?.mode === 'retry'}
                    >
                      <RefreshCw className="h-4 w-4 mr-2" />
                      Retry Connection
                    </Button>
                  )}
                  {onClearCache && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={onClearCache}
                    >
                      <Database className="h-4 w-4 mr-2" />
                      Clear Cache
                    </Button>
                  )}
                </div>
              </div>
            )}

            {/* Health Issues */}
            {lastConnectionHealth && !lastConnectionHealth.healthy && (
              <div className="border-l-4 border-yellow-400 bg-yellow-50 p-3 rounded">
                <div className="flex items-start space-x-2">
                  <AlertTriangle className="h-5 w-5 text-yellow-600 mt-0.5" />
                  <div>
                    <h5 className="font-medium text-yellow-800">Connection Issues Detected</h5>
                    <ul className="text-sm text-yellow-700 mt-1 space-y-1">
                      {lastConnectionHealth.issues?.map((issue: string, index: number) => (
                        <li key={index}>• {issue}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </CollapsibleContent>
      </Collapsible>
    </Card>
  )
}
