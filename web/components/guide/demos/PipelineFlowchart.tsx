'use client'

import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent } from '@/components/ui/collapsible'
import {
  Globe,
  Settings,
  BarChart3,
  Zap,
  Play,
  Pause,
  RotateCcw,
  Loader2,
  CheckCircle2,
  Clock
} from 'lucide-react'

type StageStatus = 'pending' | 'running' | 'completed' | 'error'

interface PipelineStage {
  id: string
  name: string
  icon: React.ElementType
  color: string
  borderColor: string
  status: StageStatus
  description: string
  details: {
    whatItDoes: string[]
    outputs: string
    duration: string
  }
}

/**
 * Interactive Pipeline Flowchart Demo - Redesigned
 * Features:
 * - Compact horizontal flowchart (always visible at top)
 * - Progressive expandable panels that appear as stages run
 * - No overlaps between flowchart and panels
 * - Clean, professional UX
 */
export function PipelineFlowchart() {
  const initialStages: PipelineStage[] = [
    {
      id: 'scrape',
      name: 'Scrape ISX',
      icon: Globe,
      color: 'bg-blue-500',
      borderColor: 'border-blue-500',
      status: 'pending',
      description: 'Downloading daily reports from ISX website',
      details: {
        whatItDoes: [
          'Downloads Excel files from ISX official website',
          'Validates file integrity and format',
          'Saves to data/downloads/ directory',
          'Supports date ranges (single or multiple days)'
        ],
        outputs: 'Excel files (.xlsx) in data/downloads/',
        duration: '~30-60s per date'
      }
    },
    {
      id: 'process',
      name: 'Process Data',
      icon: Settings,
      color: 'bg-green-500',
      borderColor: 'border-green-500',
      status: 'pending',
      description: 'Converting Excel to validated CSV format',
      details: {
        whatItDoes: [
          'Converts Excel files to CSV format',
          'Validates data integrity and structure',
          'Normalizes company names and symbols',
          'Processes all files in downloads folder'
        ],
        outputs: 'CSV files in data/reports/combined/',
        duration: '~10-20s per file'
      }
    },
    {
      id: 'index',
      name: 'Extract Index',
      icon: BarChart3,
      color: 'bg-purple-500',
      borderColor: 'border-purple-500',
      status: 'pending',
      description: 'Calculating ISX60 index values',
      details: {
        whatItDoes: [
          'Calculates ISX60 index values',
          'Extracts market performance metrics',
          'Creates index reports and charts',
          'Tracks index changes over time'
        ],
        outputs: 'Index CSV in data/reports/indexes/',
        duration: '~5s'
      }
    },
    {
      id: 'liquidity',
      name: 'Calculate Liquidity',
      icon: Zap,
      color: 'bg-amber-500',
      borderColor: 'border-amber-500',
      status: 'pending',
      description: 'Calculating liquidity metrics and trading insights',
      details: {
        whatItDoes: [
          'Calculates liquidity activity scores (0-100)',
          'Estimates bid-ask spread indicators',
          'Identifies active vs suspended stocks',
          'Generates enhanced metrics reports'
        ],
        outputs: 'Liquidity reports in data/reports/liquidity_reports/',
        duration: '~15s'
      }
    }
  ]

  const [stages, setStages] = useState<PipelineStage[]>(initialStages)
  const [currentStageIndex, setCurrentStageIndex] = useState(-1)
  const [isRunning, setIsRunning] = useState(false)
  const [expandedStages, setExpandedStages] = useState<Set<string>>(new Set())
  const currentStage = currentStageIndex >= 0 ? stages[currentStageIndex] : undefined

  // Auto-run animation
  useEffect(() => {
    if (!isRunning) return

    const timer = setInterval(() => {
      setCurrentStageIndex((prev) => {
        const nextIndex = prev + 1

        if (nextIndex >= stages.length) {
          setIsRunning(false)
          return prev
        }

        // Update stage statuses
        setStages((prevStages) =>
          prevStages.map((stage, idx) => {
            if (idx < nextIndex) {
              return { ...stage, status: 'completed' as StageStatus }
            } else if (idx === nextIndex) {
              return { ...stage, status: 'running' as StageStatus }
            }
            return stage
          })
        )

        // Auto-expand panel for running stage
        setExpandedStages((prev) => {
          const newSet = new Set(prev)
          const stageId = stages[nextIndex]?.id
          if (stageId) {
            newSet.add(stageId)
          }
          return newSet
        })

        return nextIndex
      })
    }, 2000) // 2 seconds per stage

    return () => clearInterval(timer)
  }, [isRunning, stages])

  const handleStart = () => {
    setStages(initialStages)
    setCurrentStageIndex(-1)
    setExpandedStages(new Set())
    setIsRunning(true)
  }

  const handlePause = () => {
    setIsRunning(false)
  }

  const handleReset = () => {
    setStages(initialStages)
    setCurrentStageIndex(-1)
    setExpandedStages(new Set())
    setIsRunning(false)
  }

  const getStatusBadge = (status: StageStatus) => {
    switch (status) {
      case 'pending':
        return (
          <Badge variant="secondary" className="text-xs">
            <Clock className="h-3 w-3 mr-1" />
            Pending
          </Badge>
        )
      case 'running':
        return (
          <Badge className="bg-blue-500 text-xs">
            <Loader2 className="h-3 w-3 mr-1 animate-spin" />
            Running
          </Badge>
        )
      case 'completed':
        return (
          <Badge className="bg-green-500 text-xs">
            <CheckCircle2 className="h-3 w-3 mr-1" />
            Completed
          </Badge>
        )
      case 'error':
        return <Badge variant="destructive" className="text-xs">Error</Badge>
    }
  }

  return (
    <div className="space-y-6">
      {/* Control Buttons */}
      <div className="flex flex-wrap items-center gap-3">
        {!isRunning ? (
          <Button onClick={handleStart} className="gap-2">
            <Play className="h-4 w-4" />
            Start Pipeline
          </Button>
        ) : (
          <Button onClick={handlePause} variant="secondary" className="gap-2">
            <Pause className="h-4 w-4" />
            Pause
          </Button>
        )}
        <Button onClick={handleReset} variant="outline" className="gap-2">
          <RotateCcw className="h-4 w-4" />
          Reset
        </Button>
        <Badge variant="outline" className="text-xs">
          {currentStage ? `Current: ${currentStage.name}` : 'Ready'}
        </Badge>
      </div>

      {/* Compact Horizontal Flowchart */}
      <Card className="p-6 bg-muted/30">
        <div className="flex items-center justify-between gap-3">
          {stages.map((stage, index) => {
            const StageIcon = stage.icon
            const isCurrentlyRunning = stage.status === 'running'

            return (
              <div key={stage.id} className="flex items-center gap-3 flex-1">
                {/* Stage Circle */}
                <div className="relative flex flex-col items-center gap-2">
                  <motion.div
                    className={`w-16 h-16 rounded-full flex items-center justify-center transition-all ${
                      stage.status === 'completed'
                        ? stage.color
                        : stage.status === 'running'
                        ? `${stage.color} animate-pulse`
                        : 'bg-muted'
                    }`}
                    initial={{ scale: 0.8, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{ delay: index * 0.1 }}
                  >
                    {isCurrentlyRunning ? (
                      <Loader2 className="h-7 w-7 text-white animate-spin" />
                    ) : (
                      <StageIcon
                        className={`h-7 w-7 ${
                          stage.status === 'pending' ? 'text-muted-foreground' : 'text-white'
                        }`}
                      />
                    )}

                    {/* Progress Ring */}
                    {isCurrentlyRunning && (
                      <motion.div
                        className="absolute inset-0 rounded-full border-4 border-white"
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: [0, 1, 0], scale: [0.8, 1.2, 1.2] }}
                        transition={{ duration: 1.5, repeat: Infinity }}
                      />
                    )}
                  </motion.div>

                  {/* Stage Label */}
                  <div className="text-center">
                    <p className="text-xs font-medium">{stage.name}</p>
                  </div>
                </div>

                {/* Connector Arrow */}
                {index < stages.length - 1 && (
                  <motion.div
                    className="flex-1 h-1 bg-muted relative"
                    initial={{ scaleX: 0 }}
                    animate={{
                      scaleX: stage.status === 'completed' ? 1 : 0,
                      backgroundColor:
                        stage.status === 'completed' ? 'rgb(34, 197, 94)' : 'rgb(229, 231, 235)'
                    }}
                    transition={{ duration: 0.5 }}
                    style={{ transformOrigin: 'left' }}
                  >
                    <div className="absolute right-0 top-1/2 transform -translate-y-1/2">
                      <motion.div
                        className="w-0 h-0 border-t-4 border-t-transparent border-b-4 border-b-transparent border-l-8"
                        animate={{
                          borderLeftColor:
                            stage.status === 'completed' ? 'rgb(34, 197, 94)' : 'rgb(229, 231, 235)'
                        }}
                        transition={{ duration: 0.5 }}
                      />
                    </div>
                  </motion.div>
                )}
              </div>
            )
          })}
        </div>
      </Card>

      {/* Progressive Expandable Panels */}
      <div className="space-y-4">
        <AnimatePresence mode="sync">
          {stages.map((stage) => {
            const StageIcon = stage.icon
            const isExpanded = expandedStages.has(stage.id)
            const shouldShow = stage.status !== 'pending'

            if (!shouldShow) return null

            return (
              <motion.div
                key={stage.id}
                initial={{ opacity: 0, y: 20, height: 0 }}
                animate={{ opacity: 1, y: 0, height: 'auto' }}
                exit={{ opacity: 0, y: -20, height: 0 }}
                transition={{ duration: 0.4 }}
              >
                <Card className={`border-2 ${stage.borderColor} bg-opacity-5`}>
                  <Collapsible open={isExpanded}>
                    <div className="p-4">
                      {/* Stage Header */}
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-3">
                          <div className={`p-2 rounded-lg ${stage.color}`}>
                            <StageIcon className="h-5 w-5 text-white" />
                          </div>
                          <div>
                            <h4 className="font-semibold text-base">{stage.name}</h4>
                            <p className="text-xs text-muted-foreground">{stage.description}</p>
                          </div>
                        </div>
                        {getStatusBadge(stage.status)}
                      </div>

                      {/* Expandable Details */}
                      <CollapsibleContent>
                        <motion.div
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          transition={{ delay: 0.2 }}
                          className="mt-4 pt-4 border-t space-y-4"
                        >
                          {/* What It Does */}
                          <div>
                            <h5 className="text-sm font-semibold mb-2">What it does:</h5>
                            <ul className="space-y-1">
                              {stage.details.whatItDoes.map((item, i) => (
                                <li key={i} className="text-xs text-muted-foreground flex items-start gap-2">
                                  <span className="text-primary mt-0.5">•</span>
                                  <span>{item}</span>
                                </li>
                              ))}
                            </ul>
                          </div>

                          {/* Outputs & Duration */}
                          <div className="grid grid-cols-2 gap-4">
                            <div>
                              <h5 className="text-sm font-semibold mb-1">Outputs:</h5>
                              <p className="text-xs text-muted-foreground">{stage.details.outputs}</p>
                            </div>
                            <div>
                              <h5 className="text-sm font-semibold mb-1">Duration:</h5>
                              <p className="text-xs text-muted-foreground">{stage.details.duration}</p>
                            </div>
                          </div>

                          {/* Progress Indicator (when running) */}
                          {stage.status === 'running' && (
                            <motion.div
                              initial={{ opacity: 0 }}
                              animate={{ opacity: 1 }}
                              className="flex items-center gap-2 text-xs text-blue-600 dark:text-blue-400"
                            >
                              <Loader2 className="h-4 w-4 animate-spin" />
                              <span>Processing...</span>
                            </motion.div>
                          )}
                        </motion.div>
                      </CollapsibleContent>
                    </div>
                  </Collapsible>
                </Card>
              </motion.div>
            )
          })}
        </AnimatePresence>
      </div>

      {/* Instructions */}
      <Card className="bg-muted/50">
        <div className="p-4">
          <h4 className="font-semibold mb-2 text-sm">How to Use:</h4>
          <ul className="space-y-1 text-xs text-muted-foreground">
            <li>• <strong>Start Pipeline:</strong> Click "Start Pipeline" to watch stages execute sequentially</li>
            <li>• <strong>Progressive Panels:</strong> Each stage panel appears and expands as it runs</li>
            <li>• <strong>Pause:</strong> Pause execution at any time to review details</li>
            <li>• <strong>Reset:</strong> Clear all progress and start over</li>
          </ul>
        </div>
      </Card>
    </div>
  )
}
