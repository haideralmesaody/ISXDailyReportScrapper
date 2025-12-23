/**
 * Quick Start vs Configure Demo
 * Interactive decision helper showing when to use each start method
 * Helps users understand which operations need configuration
 */

'use client'

import React, { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Zap,
  Settings2,
  Calendar,
  Download,
  FileSpreadsheet,
  BarChart3,
  Cloud,
  Workflow,
  Check,
  X,
  HelpCircle,
  ChevronRight,
  Info
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface OperationInfo {
  id: string
  name: string
  icon: React.ComponentType<{ className?: string }>
  description: string
  needsConfiguration: boolean
  requiresDates: boolean
  startMethod: 'quickstart' | 'configure'
  configOptions?: string[]
  estimatedTime: string
}

const operations: OperationInfo[] = [
  {
    id: 'processing',
    name: 'Process Data',
    icon: FileSpreadsheet,
    description: 'Convert Excel files to CSV format',
    needsConfiguration: false,
    requiresDates: false,
    startMethod: 'quickstart',
    estimatedTime: '~10 seconds'
  },
  {
    id: 'indices',
    name: 'Extract Index',
    icon: BarChart3,
    description: 'Calculate ISX60 index values',
    needsConfiguration: false,
    requiresDates: false,
    startMethod: 'quickstart',
    estimatedTime: '~5 seconds'
  },
  {
    id: 'scraping',
    name: 'Scrape ISX',
    icon: Download,
    description: 'Download reports from ISX website',
    needsConfiguration: true,
    requiresDates: true,
    startMethod: 'configure',
    configOptions: ['Select date range', 'Force re-download option'],
    estimatedTime: '~30 seconds'
  },
  {
    id: 'full_pipeline',
    name: 'Full Pipeline',
    icon: Workflow,
    description: 'Run all operations automatically',
    needsConfiguration: true,
    requiresDates: true,
    startMethod: 'configure',
    configOptions: ['Select date range', 'Enable/disable upload step'],
    estimatedTime: '~1 minute'
  },
  {
    id: 'upload',
    name: 'Upload to Sheets',
    icon: Cloud,
    description: 'Sync data to Google Sheets',
    needsConfiguration: false,
    requiresDates: false,
    startMethod: 'quickstart',
    estimatedTime: '~20 seconds'
  }
]

export function QuickStartVsConfigureDemo() {
  const [selectedOperation, setSelectedOperation] = useState<OperationInfo | null>(null)

  const quickStartOps = operations.filter(op => op.startMethod === 'quickstart')
  const configureOps = operations.filter(op => op.startMethod === 'configure')

  return (
    <div className="space-y-6">
      {/* Decision flowchart */}
      <Card className="border-2 border-primary/20 bg-primary/5">
        <CardHeader>
          <div className="flex items-start gap-3">
            <HelpCircle className="h-6 w-6 text-primary shrink-0 mt-0.5" />
            <div>
              <CardTitle>Decision Guide</CardTitle>
              <CardDescription>
                Choose the right start method based on your operation
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {/* Question */}
            <div className="text-center p-4 rounded-lg bg-background border-2 border-dashed">
              <p className="font-semibold text-lg mb-2">Does your operation need date selection?</p>
              <p className="text-sm text-muted-foreground">Operations like Scraping and Full Pipeline need dates</p>
            </div>

            {/* Branches */}
            <div className="grid md:grid-cols-2 gap-4">
              {/* No - Quick Start */}
              <div className="space-y-3">
                <div className="flex items-center justify-center gap-2 p-3 rounded-lg bg-green-500/10 border border-green-500/30">
                  <X className="h-5 w-5 text-green-600" />
                  <span className="font-semibold text-green-700 dark:text-green-400">No dates needed</span>
                </div>
                <div className="text-center p-4 rounded-lg bg-green-500/5 border-2 border-green-500/30">
                  <Zap className="h-8 w-8 text-green-600 mx-auto mb-2" />
                  <p className="font-bold text-green-700 dark:text-green-400">Use Quick Start</p>
                  <p className="text-xs text-muted-foreground mt-1">Starts immediately, no configuration needed</p>
                </div>
              </div>

              {/* Yes - Configure */}
              <div className="space-y-3">
                <div className="flex items-center justify-center gap-2 p-3 rounded-lg bg-blue-500/10 border border-blue-500/30">
                  <Check className="h-5 w-5 text-blue-600" />
                  <span className="font-semibold text-blue-700 dark:text-blue-400">Dates required</span>
                </div>
                <div className="text-center p-4 rounded-lg bg-blue-500/5 border-2 border-blue-500/30">
                  <Settings2 className="h-8 w-8 text-blue-600 mx-auto mb-2" />
                  <p className="font-bold text-blue-700 dark:text-blue-400">Use Configure & Start</p>
                  <p className="text-xs text-muted-foreground mt-1">Opens panel to select date range</p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Operations by method */}
      <div className="grid md:grid-cols-2 gap-6">
        {/* Quick Start Operations */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <div className="p-2 rounded-lg bg-green-500/10">
                <Zap className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <CardTitle className="text-lg">Quick Start</CardTitle>
                <CardDescription className="text-xs">No configuration needed</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            {quickStartOps.map((op) => {
              const Icon = op.icon
              return (
                <button
                  key={op.id}
                  onClick={() => setSelectedOperation(op)}
                  className={cn(
                    "w-full p-3 rounded-lg border-2 transition-all text-left",
                    "hover:border-green-500/50 hover:bg-green-500/5",
                    selectedOperation?.id === op.id
                      ? "border-green-500 bg-green-500/10"
                      : "border-border"
                  )}
                >
                  <div className="flex items-start gap-3">
                    <Icon className="h-5 w-5 text-green-600 shrink-0 mt-0.5" />
                    <div className="flex-1 min-w-0">
                      <div className="font-semibold text-sm">{op.name}</div>
                      <div className="text-xs text-muted-foreground line-clamp-1">{op.description}</div>
                      <Badge variant="secondary" className="mt-1 text-[10px]">{op.estimatedTime}</Badge>
                    </div>
                  </div>
                </button>
              )
            })}
          </CardContent>
        </Card>

        {/* Configure Operations */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <div className="p-2 rounded-lg bg-blue-500/10">
                <Settings2 className="h-5 w-5 text-blue-600" />
              </div>
              <div>
                <CardTitle className="text-lg">Configure & Start</CardTitle>
                <CardDescription className="text-xs">Requires date selection</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-3">
            {configureOps.map((op) => {
              const Icon = op.icon
              return (
                <button
                  key={op.id}
                  onClick={() => setSelectedOperation(op)}
                  className={cn(
                    "w-full p-3 rounded-lg border-2 transition-all text-left",
                    "hover:border-blue-500/50 hover:bg-blue-500/5",
                    selectedOperation?.id === op.id
                      ? "border-blue-500 bg-blue-500/10"
                      : "border-border"
                  )}
                >
                  <div className="flex items-start gap-3">
                    <Icon className="h-5 w-5 text-blue-600 shrink-0 mt-0.5" />
                    <div className="flex-1 min-w-0">
                      <div className="font-semibold text-sm flex items-center gap-2">
                        {op.name}
                        <Calendar className="h-3 w-3 text-muted-foreground" />
                      </div>
                      <div className="text-xs text-muted-foreground line-clamp-1">{op.description}</div>
                      <Badge variant="secondary" className="mt-1 text-[10px]">{op.estimatedTime}</Badge>
                    </div>
                  </div>
                </button>
              )
            })}
          </CardContent>
        </Card>
      </div>

      {/* Selected operation details */}
      {selectedOperation && (
        <Alert className="border-2">
          <Info className="h-4 w-4" />
          <AlertDescription>
            <div className="space-y-3">
              <div>
                <div className="font-semibold mb-1">{selectedOperation.name}</div>
                <p className="text-sm">{selectedOperation.description}</p>
              </div>

              {selectedOperation.startMethod === 'quickstart' ? (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-green-500" />
                    <span>No configuration needed</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-green-500" />
                    <span>Starts immediately when clicked</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm">
                    <Check className="h-4 w-4 text-green-500" />
                    <span>Uses existing files/data automatically</span>
                  </div>
                  <Button size="sm" className="mt-2" disabled>
                    <Zap className="h-3.5 w-3.5 mr-1.5" />
                    Quick Start (Demo)
                    <ChevronRight className="h-3 w-3 ml-auto opacity-50" />
                  </Button>
                </div>
              ) : (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <Calendar className="h-4 w-4 text-blue-500" />
                    <span>Requires date range selection</span>
                  </div>
                  {selectedOperation.configOptions && (
                    <div className="mt-2 space-y-1">
                      <p className="text-xs font-medium">Configuration Options:</p>
                      {selectedOperation.configOptions.map((option, i) => (
                        <div key={i} className="flex items-center gap-2 text-xs text-muted-foreground ml-4">
                          <div className="h-1 w-1 rounded-full bg-primary" />
                          <span>{option}</span>
                        </div>
                      ))}
                    </div>
                  )}
                  <Button size="sm" className="mt-2" disabled>
                    <Settings2 className="h-3.5 w-3.5 mr-1.5" />
                    Configure & Start (Demo)
                    <ChevronRight className="h-3 w-3 ml-auto opacity-50" />
                  </Button>
                </div>
              )}

              <div className="pt-2 border-t">
                <p className="text-xs text-muted-foreground">
                  ⏱️ Estimated time: <span className="font-medium">{selectedOperation.estimatedTime}</span>
                </p>
              </div>
            </div>
          </AlertDescription>
        </Alert>
      )}
    </div>
  )
}

export default QuickStartVsConfigureDemo
