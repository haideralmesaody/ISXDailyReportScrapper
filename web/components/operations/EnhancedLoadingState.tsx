'use client'

import React from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Loader2, Wifi, Database, Activity, CheckCircle } from 'lucide-react'

interface LoadingState {
  isLoading: boolean
  stage: 'initial' | 'operation_types' | 'initial_operations' | 'websocket_connection' | 'completed'
  progress: number
  message: string
}

interface EnhancedLoadingStateProps {
  loadingState: LoadingState
  connected: boolean
  connectionStatus: string
}

export function EnhancedLoadingState({ loadingState, connected, connectionStatus }: EnhancedLoadingStateProps) {
  if (!loadingState.isLoading) return null

  const getStageIcon = (stage: string) => {
    switch (stage) {
      case 'initial':
        return <Loader2 className="h-4 w-4 animate-spin" />
      case 'websocket_connection':
        return <Wifi className="h-4 w-4" />
      case 'operation_types':
        return <Database className="h-4 w-4" />
      case 'initial_operations':
        return <Activity className="h-4 w-4" />
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      default:
        return <Loader2 className="h-4 w-4 animate-spin" />
    }
  }

  const getStageColor = (stage: string, isCompleted: boolean) => {
    if (isCompleted) return 'text-green-600'
    if (stage === loadingState.stage) return 'text-blue-600'
    return 'text-gray-400'
  }

  const stages = [
    { key: 'initial', label: 'Initializing', description: 'Setting up operations system' },
    { key: 'websocket_connection', label: 'Connecting', description: 'Establishing WebSocket connection' },
    { key: 'operation_types', label: 'Loading Types', description: 'Fetching available operations' },
    { key: 'initial_operations', label: 'Loading Data', description: 'Loading recent operations' },
    { key: 'completed', label: 'Ready', description: 'System fully operational' }
  ]

  const currentStageIndex = stages.findIndex(s => s.key === loadingState.stage)
  const isCurrentStageCompleted = loadingState.progress === 100

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardContent className="p-6">
        <div className="space-y-6">
          {/* Main Progress */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                {getStageIcon(loadingState.stage)}
                <h3 className="text-lg font-semibold">{loadingState.message}</h3>
              </div>
              <Badge variant="outline">
                {Math.round(loadingState.progress)}%
              </Badge>
            </div>
            <Progress value={loadingState.progress} className="h-2" />
          </div>

          {/* Stage Breakdown */}
          <div className="space-y-3">
            <h4 className="text-sm font-medium text-muted-foreground">Initialization Steps</h4>
            <div className="space-y-2">
              {stages.map((stage, index) => {
                const isCompleted = index < currentStageIndex || (index === currentStageIndex && isCurrentStageCompleted)
                const isCurrent = index === currentStageIndex && !isCurrentStageCompleted

                return (
                  <div key={stage.key} className="flex items-center space-x-3">
                    <div className={getStageColor(stage.key, isCompleted)}>
                      {getStageIcon(stage.key)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <p className={`text-sm font-medium ${isCompleted ? 'text-green-600' : isCurrent ? 'text-blue-600' : 'text-gray-500'}`}>
                          {stage.label}
                        </p>
                        {isCompleted && (
                          <CheckCircle className="h-4 w-4 text-green-600" />
                        )}
                      </div>
                      <p className="text-xs text-muted-foreground">{stage.description}</p>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Connection Status */}
          <div className="border-t pt-4">
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center space-x-2">
                <Wifi className={`h-4 w-4 ${connected ? 'text-green-600' : 'text-gray-400'}`} />
                <span className="text-muted-foreground">Connection:</span>
                <Badge variant={connected ? 'default' : 'secondary'}>
                  {connectionStatus}
                </Badge>
              </div>
              {loadingState.stage === 'websocket_connection' && (
                <span className="text-blue-600">Establishing connection...</span>
              )}
            </div>
          </div>

          {/* Helpful Tips */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
            <div className="flex items-start space-x-2">
              <Activity className="h-5 w-5 text-blue-600 mt-0.5" />
              <div>
                <p className="text-sm text-blue-800 font-medium">System Tip</p>
                <p className="text-xs text-blue-700 mt-1">
                  {loadingState.stage === 'initial' && 'Preparing the operations system for optimal performance...'}
                  {loadingState.stage === 'websocket_connection' && 'Connecting to real-time update service...'}
                  {loadingState.stage === 'operation_types' && 'Loading available operation types and configurations...'}
                  {loadingState.stage === 'initial_operations' && 'Fetching recent operation history...'}
                  {loadingState.stage === 'completed' && 'All systems ready! You can now start operations.'}
                </p>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}