/**
 * WebSocket Explainer
 * Interactive guide explaining WebSocket real-time updates
 * Shows connection status, progress updates, and troubleshooting
 */

'use client'

import React, { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  WifiOff,
  Zap,
  MessageSquare,
  TrendingUp
} from 'lucide-react'
import { cn } from '@/lib/utils'

type ConnectionState = 'connected' | 'disconnected' | 'connecting'

interface SimulatedUpdate {
  type: 'status' | 'progress' | 'stage' | 'complete'
  message: string
  timestamp: string
  progress?: number
  stage?: string
}

export function WebSocketExplainer() {
  const [simulatedState, setSimulatedState] = useState<ConnectionState>('connected')
  const [updates, setUpdates] = useState<SimulatedUpdate[]>([])
  const [isSimulating, setIsSimulating] = useState(false)

  // Simulate connection toggle
  const toggleConnection = () => {
    if (simulatedState === 'connected') {
      setSimulatedState('disconnected')
      setUpdates([])
    } else {
      setSimulatedState('connecting')
      setTimeout(() => setSimulatedState('connected'), 1000)
    }
  }

  // Simulate operation updates
  const startSimulation = () => {
    if (simulatedState !== 'connected') return

    setIsSimulating(true)
    setUpdates([])

    const simulatedUpdates: SimulatedUpdate[] = [
      { type: 'status', message: 'Operation started', timestamp: new Date().toLocaleTimeString(), progress: 0 },
      { type: 'stage', message: 'Downloading files...', timestamp: new Date().toLocaleTimeString(), progress: 10, stage: 'Download' },
      { type: 'progress', message: 'Files downloaded: 3/10', timestamp: new Date().toLocaleTimeString(), progress: 30 },
      { type: 'progress', message: 'Files downloaded: 7/10', timestamp: new Date().toLocaleTimeString(), progress: 70 },
      { type: 'stage', message: 'Processing data...', timestamp: new Date().toLocaleTimeString(), progress: 80, stage: 'Process' },
      { type: 'progress', message: 'Validation complete', timestamp: new Date().toLocaleTimeString(), progress: 95 },
      { type: 'complete', message: 'Operation completed successfully!', timestamp: new Date().toLocaleTimeString(), progress: 100 }
    ]

    let index = 0
    const interval = setInterval(() => {
      const update = index < simulatedUpdates.length ? simulatedUpdates[index] : undefined
      if (!update) {
        clearInterval(interval)
        setIsSimulating(false)
        return
      }

      setUpdates(prev => [...prev, update])
      index++
    }, 1000)
  }

  return (
    <div className="space-y-6">
      {/* Connection Status Demo */}
      <Card className="border-2">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <Activity className="h-6 w-6 text-primary" />
              <div>
                <CardTitle>WebSocket Connection</CardTitle>
                <CardDescription>Real-time communication with server</CardDescription>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <div className={cn(
                "h-3 w-3 rounded-full transition-all duration-300",
                simulatedState === 'connected' && "bg-green-500 shadow-lg shadow-green-500/50",
                simulatedState === 'disconnected' && "bg-red-500",
                simulatedState === 'connecting' && "bg-yellow-500 animate-pulse"
              )} />
              <span className={cn(
                "text-sm font-medium transition-colors",
                simulatedState === 'connected' && "text-green-600 dark:text-green-400",
                simulatedState === 'disconnected' && "text-red-600 dark:text-red-400",
                simulatedState === 'connecting' && "text-yellow-600 dark:text-yellow-400"
              )}>
                {simulatedState === 'connected' && 'Connected'}
                {simulatedState === 'disconnected' && 'Disconnected'}
                {simulatedState === 'connecting' && 'Connecting...'}
              </span>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Status explanation */}
          <div className="grid md:grid-cols-2 gap-4">
            <div className="space-y-3">
              <div className="flex items-start gap-3 p-3 rounded-lg bg-green-500/10 border border-green-500/30">
                <CheckCircle2 className="h-5 w-5 text-green-600 shrink-0 mt-0.5" />
                <div>
                  <div className="font-semibold text-sm text-green-700 dark:text-green-400">Connected</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    Live updates enabled. Progress appears automatically.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 rounded-lg bg-red-500/10 border border-red-500/30">
                <WifiOff className="h-5 w-5 text-red-600 shrink-0 mt-0.5" />
                <div>
                  <div className="font-semibold text-sm text-red-700 dark:text-red-400">Disconnected</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    No live updates. Refresh page to reconnect.
                  </p>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <Button
                onClick={toggleConnection}
                variant="outline"
                className="w-full"
                disabled={simulatedState === 'connecting'}
              >
                {simulatedState === 'disconnected' ? (
                  <>
                    <Activity className="h-4 w-4 mr-2" />
                    Connect
                  </>
                ) : (
                  <>
                    <WifiOff className="h-4 w-4 mr-2" />
                    Disconnect
                  </>
                )}
              </Button>

              <Button
                onClick={startSimulation}
                className="w-full"
                disabled={isSimulating || simulatedState !== 'connected'}
              >
                <Zap className="h-4 w-4 mr-2" />
                {isSimulating ? 'Simulating...' : 'Simulate Operation'}
              </Button>

              <p className="text-xs text-muted-foreground text-center pt-2">
                Try toggling connection and starting simulation to see live updates
              </p>
            </div>
          </div>

          {/* Updates feed */}
          {updates.length > 0 && (
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm font-medium">
                <MessageSquare className="h-4 w-4" />
                Live Updates
              </div>
              <div className="space-y-2 max-h-64 overflow-y-auto rounded-lg border bg-muted/30 p-3">
                {updates.map((update, index) => (
                  <div
                    key={index}
                    className={cn(
                      "flex items-start gap-3 p-2 rounded-md animate-in slide-in-from-top-2",
                      update.type === 'complete' && "bg-green-500/10 border border-green-500/30"
                    )}
                  >
                    <div className="shrink-0 mt-0.5">
                      {update.type === 'complete' ? (
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      ) : update.type === 'progress' ? (
                        <TrendingUp className="h-4 w-4 text-primary" />
                      ) : (
                        <Activity className="h-4 w-4 text-blue-500" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm font-medium">{update.message}</span>
                        {update.progress !== undefined && (
                          <Badge variant="secondary" className="text-xs">
                            {update.progress}%
                          </Badge>
                        )}
                      </div>
                      <div className="text-xs text-muted-foreground mt-0.5">
                        {update.timestamp}
                      </div>
                      {update.progress !== undefined && update.progress < 100 && (
                        <div className="mt-2 h-1.5 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-primary transition-all duration-500"
                            style={{ width: `${update.progress}%` }}
                          />
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* What WebSocket Does */}
      <div className="grid md:grid-cols-2 gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <CheckCircle2 className="h-5 w-5 text-green-500" />
              What It Does
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="space-y-2 text-sm">
              <div className="flex items-start gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-primary shrink-0 mt-1.5" />
                <span><strong>Real-time progress:</strong> Updates appear instantly as operation runs</span>
              </div>
              <div className="flex items-start gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-primary shrink-0 mt-1.5" />
                <span><strong>No refreshing:</strong> Page updates automatically, no manual refresh needed</span>
              </div>
              <div className="flex items-start gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-primary shrink-0 mt-1.5" />
                <span><strong>Live status:</strong> See current stage, files processed, progress percentage</span>
              </div>
              <div className="flex items-start gap-2">
                <div className="h-1.5 w-1.5 rounded-full bg-primary shrink-0 mt-1.5" />
                <span><strong>Instant errors:</strong> Failures appear immediately with details</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-orange-500" />
              Troubleshooting
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Alert variant="destructive" className="bg-destructive/10">
              <AlertDescription className="text-sm">
                <div className="space-y-2">
                  <p><strong>Problem:</strong> Red dot (Disconnected)</p>
                  <p className="text-xs"><strong>Solution:</strong> Refresh the page (F5 or Ctrl+R)</p>
                </div>
              </AlertDescription>
            </Alert>

            <Alert variant="destructive" className="bg-destructive/10">
              <AlertDescription className="text-sm">
                <div className="space-y-2">
                  <p><strong>Problem:</strong> Progress not updating</p>
                  <p className="text-xs"><strong>Solution:</strong> Check green dot. If red, refresh page.</p>
                </div>
              </AlertDescription>
            </Alert>

            <Alert variant="destructive" className="bg-destructive/10">
              <AlertDescription className="text-sm">
                <div className="space-y-2">
                  <p><strong>Problem:</strong> Connection keeps dropping</p>
                  <p className="text-xs"><strong>Solution:</strong> Check internet connection or VPN</p>
                </div>
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </div>

      {/* Pro Tip */}
      <Card className="border-blue-500/50 bg-blue-500/5">
        <CardContent className="pt-6">
          <div className="flex items-start gap-3">
            <Zap className="h-5 w-5 text-blue-500 shrink-0 mt-0.5" />
            <div>
              <h4 className="font-semibold mb-1">Pro Tip</h4>
              <p className="text-sm text-muted-foreground">
                <strong>Always check the green dot before starting operations.</strong> If it's red, refresh the page first.
                This ensures you'll see live progress updates and won't miss any important status changes.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default WebSocketExplainer
