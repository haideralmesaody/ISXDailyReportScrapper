'use client'

import { useState, useEffect, useRef } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Play,
  RotateCcw,
  Loader2,
  CheckCircle2,
  XCircle,
  Globe,
  Cog,
  BarChart3,
  Cloud,
  Zap
} from 'lucide-react'

type OperationType = 'scrape' | 'process' | 'index' | 'upload' | 'pipeline'
type OperationStatus = 'idle' | 'running' | 'completed' | 'failed'

interface OperationLog {
  timestamp: string
  message: string
  type: 'info' | 'success' | 'error'
}

/**
 * Operation Simulator Demo
 * Interactive mock operation execution with realistic progress and logging
 */
export function OperationSimulator() {
  const [selectedOperation, setSelectedOperation] = useState<OperationType>('scrape')
  const [status, setStatus] = useState<OperationStatus>('idle')
  const [progress, setProgress] = useState(0)
  const [logs, setLogs] = useState<OperationLog[]>([])
  const [currentStage, setCurrentStage] = useState('')
  const [shouldFail, setShouldFail] = useState(false)
  const logsEndRef = useRef<HTMLDivElement>(null)

  const operations = [
    {
      value: 'scrape',
      label: 'Scrape ISX',
      icon: Globe,
      color: 'text-blue-500',
      stages: [
        { name: 'Connecting to ISX website', duration: 800 },
        { name: 'Authenticating session', duration: 600 },
        { name: 'Downloading Excel files', duration: 1200 },
        { name: 'Validating file integrity', duration: 500 },
        { name: 'Saving to data/downloads', duration: 400 }
      ]
    },
    {
      value: 'process',
      label: 'Process Data',
      icon: Cog,
      color: 'text-green-500',
      stages: [
        { name: 'Loading Excel files', duration: 600 },
        { name: 'Parsing data rows', duration: 1000 },
        { name: 'Validating data integrity', duration: 700 },
        { name: 'Converting to CSV format', duration: 800 },
        { name: 'Writing output files', duration: 500 }
      ]
    },
    {
      value: 'index',
      label: 'Extract Index',
      icon: BarChart3,
      color: 'text-purple-500',
      stages: [
        { name: 'Loading processed CSV data', duration: 500 },
        { name: 'Identifying ISX60 constituents', duration: 800 },
        { name: 'Calculating market cap weights', duration: 900 },
        { name: 'Computing index value', duration: 600 },
        { name: 'Saving index history', duration: 400 }
      ]
    },
    {
      value: 'upload',
      label: 'Upload to Sheets',
      icon: Cloud,
      color: 'text-orange-500',
      stages: [
        { name: 'Authenticating with Google API', duration: 700 },
        { name: 'Reading CSV files', duration: 600 },
        { name: 'Preparing batch upload', duration: 500 },
        { name: 'Uploading to Google Sheets', duration: 1500 },
        { name: 'Verifying row counts', duration: 400 }
      ]
    },
    {
      value: 'pipeline',
      label: 'Full Pipeline',
      icon: Zap,
      color: 'text-emerald-500',
      stages: [
        { name: 'Stage 1: Scraping ISX data', duration: 1200 },
        { name: 'Stage 2: Processing Excel files', duration: 1000 },
        { name: 'Stage 3: Extracting index values', duration: 800 },
        { name: 'Stage 4: Uploading to Sheets', duration: 1000 },
        { name: 'Finalizing pipeline', duration: 500 }
      ]
    }
  ]

  const currentOp = operations.find(op => op.value === selectedOperation)!

  // Auto-scroll logs
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  const addLog = (message: string, type: OperationLog['type'] = 'info') => {
    const timestamp = new Date().toLocaleTimeString()
    setLogs(prev => [...prev, { timestamp, message, type }])
  }

  const simulateOperation = async () => {
    setStatus('running')
    setProgress(0)
    setLogs([])
    setCurrentStage('')

    addLog(`Starting ${currentOp.label} operation...`, 'info')

    const totalStages = currentOp.stages.length
    let failedAt = -1

    if (shouldFail) {
      failedAt = Math.floor(Math.random() * totalStages)
    }

    for (let i = 0; i < totalStages; i++) {
      const stage = currentOp.stages[i]
      if (!stage) continue
      setCurrentStage(stage.name)
      addLog(`[${i + 1}/${totalStages}] ${stage.name}...`, 'info')

      // Simulate stage progress
      const stageProgressSteps = 10
      for (let j = 0; j <= stageProgressSteps; j++) {
        await new Promise(resolve => setTimeout(resolve, stage.duration / stageProgressSteps))
        const stageProgress = (i / totalStages) + (j / stageProgressSteps / totalStages)
        setProgress(Math.round(stageProgress * 100))
      }

      // Check if this stage should fail
      if (i === failedAt) {
        addLog(`❌ Error: ${stage.name} failed`, 'error')
        addLog(`Cause: Network timeout (simulated)`, 'error')
        addLog(`Solution: Retry the operation`, 'error')
        setStatus('failed')
        setProgress(Math.round((i / totalStages) * 100))
        return
      }

      addLog(`✅ ${stage.name} completed`, 'success')
    }

    setProgress(100)
    setStatus('completed')
    setCurrentStage('Operation completed successfully!')
    addLog(`✅ ${currentOp.label} completed successfully!`, 'success')
    addLog(`Total time: ${currentOp.stages.reduce((sum, s) => sum + s.duration, 0)}ms`, 'info')
  }

  const handleReset = () => {
    setStatus('idle')
    setProgress(0)
    setLogs([])
    setCurrentStage('')
  }

  const getStatusBadge = () => {
    switch (status) {
      case 'idle':
        return <Badge variant="secondary">Ready</Badge>
      case 'running':
        return <Badge className="bg-blue-500">Running</Badge>
      case 'completed':
        return <Badge className="bg-green-500">Completed</Badge>
      case 'failed':
        return <Badge variant="destructive">Failed</Badge>
    }
  }

  const OpIcon = currentOp.icon

  return (
    <div className="space-y-6">
      {/* Operation Selection */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-2">
          <label className="text-sm font-semibold">Select Operation</label>
          <Select value={selectedOperation} onValueChange={(value) => setSelectedOperation(value as OperationType)} disabled={status === 'running'}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {operations.map((op) => {
                const Icon = op.icon
                return (
                  <SelectItem key={op.value} value={op.value}>
                    <div className="flex items-center gap-2">
                      <Icon className={`h-4 w-4 ${op.color}`} />
                      <span>{op.label}</span>
                    </div>
                  </SelectItem>
                )
              })}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <label className="text-sm font-semibold">Simulation Mode</label>
          <Select value={shouldFail ? 'fail' : 'success'} onValueChange={(value) => setShouldFail(value === 'fail')} disabled={status === 'running'}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="success">Success Scenario</SelectItem>
              <SelectItem value="fail">Error Scenario (random failure)</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center gap-3">
        {status !== 'running' ? (
          <Button onClick={simulateOperation} className="gap-2">
            <Play className="h-4 w-4" />
            Start Operation
          </Button>
        ) : (
          <Button disabled className="gap-2">
            <Loader2 className="h-4 w-4 animate-spin" />
            Running...
          </Button>
        )}
        <Button onClick={handleReset} variant="outline" className="gap-2" disabled={status === 'running'}>
          <RotateCcw className="h-4 w-4" />
          Reset
        </Button>
      </div>

      {/* Status Card */}
      <Card className="p-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={`p-2 rounded-lg bg-muted`}>
                <OpIcon className={`h-5 w-5 ${currentOp.color}`} />
              </div>
              <div>
                <h4 className="font-semibold">{currentOp.label}</h4>
                {getStatusBadge()}
              </div>
            </div>
            {status === 'completed' && <CheckCircle2 className="h-6 w-6 text-green-500" />}
            {status === 'failed' && <XCircle className="h-6 w-6 text-red-500" />}
            {status === 'running' && <Loader2 className="h-6 w-6 text-blue-500 animate-spin" />}
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Progress</span>
              <span className="font-semibold">{progress}%</span>
            </div>
            <Progress value={progress} className="h-2" />
          </div>

          {currentStage && (
            <div className="text-sm text-muted-foreground">
              {status === 'running' && <Loader2 className="h-3 w-3 inline animate-spin mr-2" />}
              {currentStage}
            </div>
          )}
        </div>
      </Card>

      {/* Operation Logs */}
      {logs.length > 0 && (
        <Card className="p-4">
          <h4 className="font-semibold mb-3">Operation Logs</h4>
          <div className="bg-muted/50 rounded-lg p-3 max-h-64 overflow-y-auto font-mono text-xs space-y-1">
            {logs.map((log, i) => (
              <div
                key={i}
                className={`${
                  log.type === 'error'
                    ? 'text-red-500'
                    : log.type === 'success'
                    ? 'text-green-500'
                    : 'text-foreground'
                }`}
              >
                <span className="text-muted-foreground">[{log.timestamp}]</span> {log.message}
              </div>
            ))}
            <div ref={logsEndRef} />
          </div>
        </Card>
      )}

      {/* Instructions */}
      <Card className="bg-muted/50">
        <div className="p-4">
          <h4 className="font-semibold mb-2 text-sm">How to Use:</h4>
          <ul className="space-y-1 text-xs text-muted-foreground">
            <li>• <strong>Select Operation:</strong> Choose from 5 operation types</li>
            <li>• <strong>Simulation Mode:</strong> Test success or error scenarios</li>
            <li>• <strong>Start Operation:</strong> Watch realistic progress simulation</li>
            <li>• <strong>View Logs:</strong> See detailed execution logs in real-time</li>
            <li>• <strong>Reset:</strong> Clear and start over</li>
          </ul>
        </div>
      </Card>
    </div>
  )
}
