'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Workflow,
  Globe,
  Settings,
  BarChart3,
  CheckCircle2,
  AlertCircle,
  ArrowRight,
  Clock,
  Zap,
  ShieldCheck,
  ChevronRight,
  RefreshCw,
  XCircle,
  Play
} from 'lucide-react'
import { InteractiveDemo } from '../components/InteractiveDemo'
import { PipelineFlowchart } from '@/components/guide/demos/PipelineFlowchart'

interface PipelineSectionProps {
  onNavigate: (sectionId: string) => void
}

/**
 * Pipeline Section - Comprehensive guide to ISX Pulse's 4-stage data pipeline
 * Covers architecture, stages, error handling, and concurrent processing
 */
export function PipelineSection({ onNavigate }: PipelineSectionProps) {
  const pipelineStages = [
    {
      id: 'scrape',
      name: 'Scrape ISX',
      icon: Globe,
      color: 'bg-blue-500',
      description: 'Download daily reports from ISX website',
      duration: '~30 seconds',
      inputs: 'Date range (from/to)',
      outputs: 'Excel files (.xlsx)',
      details: [
        'Connects to ISX website using secure HTTPS',
        'Downloads daily trading reports for selected date range',
        'Validates file integrity and format',
        'Stores files in data/downloads directory',
        'Supports automatic retry with backoff for network failures',
        'Skips existing files to avoid redundant downloads'
      ],
      errors: [
        'Network timeout - Automatic retry with exponential backoff',
        'Invalid date - User prompted to select valid trading date',
        'File already exists - Skipped automatically'
      ]
    },
    {
      id: 'process',
      name: 'Process Data',
      icon: Settings,
      color: 'bg-green-500',
      description: 'Convert Excel to CSV with validation',
      duration: '~10 seconds',
      inputs: 'Excel files from downloads directory',
      outputs: 'CSV files with validated trading data',
      details: [
        'Parses Excel files using efficient streaming',
        'Validates data integrity (missing fields, invalid values)',
        'Converts Arabic company names to UTF-8',
        'Standardizes number formats and dates',
        'Generates per-ticker trading history CSV files',
        'Creates combined dataset for analysis'
      ],
      errors: [
        'Corrupted Excel file - Skip file and log error',
        'Missing required columns - Abort with detailed error message',
        'Data validation failure - Mark rows as invalid and continue'
      ]
    },
    {
      id: 'index',
      name: 'Extract Index',
      icon: BarChart3,
      color: 'bg-purple-500',
      description: 'Extract ISX60 index data',
      duration: '~5 seconds',
      inputs: 'Excel files from downloads directory',
      outputs: 'Index CSV (indexes.csv)',
      details: [
        'Identifies ISX60 constituent stocks from Excel files',
        'Calculates daily index value using market cap weights',
        'Computes daily change and percentage movement',
        'Validates index calculation accuracy',
        'Stores index history in data/reports/indexes/',
        'Generates time series data for index charting'
      ],
      errors: [
        'Missing constituent data - Skip date and log warning',
        'Index calculation error - Abort with details',
        'Excel file format changed - Report parsing error'
      ]
    },
    {
      id: 'liquidity',
      name: 'Calculate Liquidity',
      icon: Zap,
      color: 'bg-amber-500',
      description: 'Calculate liquidity metrics and trading insights',
      duration: '~15 seconds',
      inputs: 'CSV trading data from ticker history files',
      outputs: 'Liquidity scores CSV and insights',
      details: [
        'Analyzes 60-day trading history per ticker',
        'Calculates Amihud illiquidity ratio (price impact)',
        'Computes trading value and continuity metrics',
        'Applies inactivity penalties for suspended trading',
        'Ranks stocks by composite liquidity score (0-100)',
        'Generates actionable trading insights and recommendations',
        'Saves results to data/reports/liquidity_reports/'
      ],
      errors: [
        'Insufficient data - Skip ticker (minimum 20 days required)',
        'Missing trading history files - Abort with file list',
        'Calculation error - Log details and continue with other tickers'
      ]
    }
  ]

  const pipelineFeatures = [
    {
      icon: Zap,
      title: 'Concurrent Processing',
      description: 'Multiple files processed in parallel using worker pools for optimal performance'
    },
    {
      icon: RefreshCw,
      title: 'Automatic Retry',
      description: 'Transient errors trigger automatic retry with exponential backoff'
    },
    {
      icon: ShieldCheck,
      title: 'Data Validation',
      description: 'Multi-layer validation ensures data integrity at every stage'
    },
    {
      icon: Clock,
      title: 'Real-time Progress',
      description: 'WebSocket updates show live progress with detailed status messages'
    }
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-primary/10">
            <Workflow className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold">Pipeline Architecture</h2>
            <p className="text-muted-foreground text-lg">
              Understanding ISX Pulse's automated 4-stage data processing pipeline
            </p>
          </div>
        </div>
      </div>

      {/* Pipeline Overview */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Pipeline Overview</h3>
        <p className="text-muted-foreground">
          ISX Pulse uses a robust 4-stage pipeline to automatically collect, process, and deliver ISX market data. Each stage is isolated, validated, and can be run independently or as part of the full pipeline.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {pipelineFeatures.map((feature, index) => (
            <Card key={index}>
              <CardContent className="pt-6">
                <feature.icon className="h-8 w-8 text-primary mb-3" />
                <h4 className="font-semibold mb-2">{feature.title}</h4>
                <p className="text-sm text-muted-foreground">
                  {feature.description}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Interactive Pipeline Flowchart */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Interactive Pipeline Flowchart</h3>
        <p className="text-muted-foreground">
          Watch the pipeline in action. Click any stage to explore its details:
        </p>

        <InteractiveDemo
          title="Animated Pipeline Execution"
          description="See how data flows through each stage with automatic execution"
          variant="interactive"
        >
          <PipelineFlowchart />
        </InteractiveDemo>
      </section>

      {/* Stage Breakdown */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Stage-by-Stage Breakdown</h3>
        <p className="text-muted-foreground">
          Detailed explanation of each pipeline stage:
        </p>

        <div className="space-y-4">
          {pipelineStages.map((stage, index) => {
            const StageIcon = stage.icon
            return (
              <Card key={stage.id} className="relative overflow-hidden">
                <div className={`absolute top-0 left-0 w-1 h-full ${stage.color}`} />
                <CardHeader>
                  <div className="flex items-start gap-4">
                    <div className={`p-3 rounded-lg ${stage.color} bg-opacity-10 shrink-0`}>
                      <StageIcon className="h-6 w-6" />
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <CardTitle className="text-xl">
                          {index + 1}. {stage.name}
                        </CardTitle>
                        <Badge variant="secondary">{stage.duration}</Badge>
                      </div>
                      <CardDescription className="text-base">
                        {stage.description}
                      </CardDescription>
                    </div>
                    {index < pipelineStages.length - 1 && (
                      <ArrowRight className="h-6 w-6 text-muted-foreground shrink-0 mt-2" />
                    )}
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Inputs/Outputs */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <h5 className="font-semibold text-sm mb-2 flex items-center gap-2">
                        <Play className="h-4 w-4 text-blue-500" />
                        Inputs
                      </h5>
                      <p className="text-sm text-muted-foreground">{stage.inputs}</p>
                    </div>
                    <div>
                      <h5 className="font-semibold text-sm mb-2 flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                        Outputs
                      </h5>
                      <p className="text-sm text-muted-foreground">{stage.outputs}</p>
                    </div>
                  </div>

                  {/* Details */}
                  <div>
                    <h5 className="font-semibold text-sm mb-2">How It Works:</h5>
                    <ul className="space-y-1">
                      {stage.details.map((detail, i) => (
                        <li key={i} className="flex items-start gap-2 text-sm">
                          <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                          <span>{detail}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  {/* Error Handling */}
                  <div>
                    <h5 className="font-semibold text-sm mb-2 flex items-center gap-2">
                      <AlertCircle className="h-4 w-4 text-orange-500" />
                      Error Handling:
                    </h5>
                    <ul className="space-y-1">
                      {stage.errors.map((error, i) => (
                        <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                          <XCircle className="h-4 w-4 text-red-500 shrink-0 mt-0.5" />
                          <span>{error}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </section>

      {/* Concurrent Processing */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Concurrent Processing</h3>
        <p className="text-muted-foreground">
          ISX Pulse processes multiple files in parallel for maximum efficiency:
        </p>

        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <Zap className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Worker Pool Pattern</h4>
                  <p className="text-sm text-muted-foreground">
                    Uses 10 concurrent workers to process files in parallel. Each worker handles one file at a time, ensuring optimal CPU and I/O utilization without overwhelming the system.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Clock className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Performance Impact</h4>
                  <p className="text-sm text-muted-foreground">
                    Processing 100 files sequentially: ~15 minutes. With concurrent processing: ~2 minutes. That's <strong>7.5x faster</strong> for typical workloads.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <ShieldCheck className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Safety Guarantees</h4>
                  <p className="text-sm text-muted-foreground">
                    Each file is processed independently. If one file fails, others continue unaffected. Failed files are logged and can be retried individually.
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Error Handling & Rollback */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Error Handling & Rollback</h3>
        <p className="text-muted-foreground">
          ISX Pulse implements comprehensive error handling with automatic rollback:
        </p>

        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            If any critical stage fails (e.g., data validation error), the pipeline automatically rolls back changes and notifies you with detailed error information. You can then fix the issue and retry the failed stage.
          </AlertDescription>
        </Alert>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card className="border-orange-500/50 bg-orange-500/5">
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <XCircle className="h-5 w-5 text-orange-500" />
                Transient Errors
              </CardTitle>
            </CardHeader>
            <CardContent className="text-sm">
              <p className="mb-2">Network timeouts, temporary API failures:</p>
              <ul className="space-y-1 text-muted-foreground">
                <li>• Automatic retry with exponential backoff</li>
                <li>• Maximum 3 retry attempts</li>
                <li>• User notification if all retries fail</li>
              </ul>
            </CardContent>
          </Card>

          <Card className="border-red-500/50 bg-red-500/5">
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <AlertCircle className="h-5 w-5 text-red-500" />
                Critical Errors
              </CardTitle>
            </CardHeader>
            <CardContent className="text-sm">
              <p className="mb-2">Data validation failures, corrupted files:</p>
              <ul className="space-y-1 text-muted-foreground">
                <li>• Immediate pipeline halt</li>
                <li>• Rollback to last known good state</li>
                <li>• Detailed error log with file/line info</li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* Next Steps */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Next Steps</h3>
        <Card className="border-primary/50 bg-primary/5">
          <CardContent className="pt-6">
            <div className="space-y-4">
              <p>
                Now that you understand the pipeline architecture and how to run operations, learn about the reports system and how to analyze your data:
              </p>
              <button
                onClick={() => onNavigate('reports')}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Continue to Reports System
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
