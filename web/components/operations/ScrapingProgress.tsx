/**
 * Simplified Scraping Progress Component
 * 
 * Industry-standard progress display following GitHub Actions/CI patterns.
 * Shows only what we know - no predictions, no complex calculations.
 */

'use client'

import React, { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { 
  Download,
  CheckCircle,
  Clock,
  Activity,
  Zap,
  FileText,
  Loader2
} from 'lucide-react'
import { cn } from '@/lib/utils'

// Status configuration
const statusConfig = {
  initializing: { 
    icon: Loader2, 
    label: 'Initializing',
    color: 'text-gray-600', 
    bgColor: 'bg-gray-50',
    iconAnimation: 'animate-spin'
  },
  scanning: { 
    icon: Activity, 
    label: 'Scanning',
    color: 'text-blue-600', 
    bgColor: 'bg-blue-50',
    iconAnimation: 'animate-pulse'
  },
  downloading: { 
    icon: Download, 
    label: 'Downloading',
    color: 'text-blue-600', 
    bgColor: 'bg-blue-50',
    iconAnimation: ''
  },
  processing: { 
    icon: Activity, 
    label: 'Processing',
    color: 'text-blue-600', 
    bgColor: 'bg-blue-50',
    iconAnimation: 'animate-pulse'
  },
  completing: {
    icon: Clock,
    label: 'Completing',
    color: 'text-orange-600',
    bgColor: 'bg-orange-50',
    iconAnimation: ''
  },
  completed: { 
    icon: CheckCircle, 
    label: 'Completed',
    color: 'text-green-600', 
    bgColor: 'bg-green-50',
    iconAnimation: ''
  },
  stopped: { 
    icon: Activity, 
    label: 'Stopped',
    color: 'text-gray-600', 
    bgColor: 'bg-gray-50',
    iconAnimation: ''
  }
} as const

interface ScrapingProgressProps {
  operation: {
    operation_id: string
    status: string
    progress: number
    message?: string
    metadata?: {
      status?: keyof typeof statusConfig
      files_processed?: number
      current_file?: string
      speed?: number // files per minute
      started_at?: string
      phase?: string
    }
  }
}

export function ScrapingProgress({ operation }: ScrapingProgressProps) {
  const metadata = operation.metadata || {}
  const scrapingStatus = metadata.status || 'initializing'
  const filesProcessed = metadata.files_processed || 0
  const currentFile = metadata.current_file
  const speed = metadata.speed
  
  // Get status configuration
  const config = statusConfig[scrapingStatus] || statusConfig.initializing
  const StatusIcon = config.icon
  
  // Calculate simple progress based on status
  const progress = useMemo(() => {
    // Use backend progress if available and reasonable
    if (operation.progress > 0 && operation.progress <= 100) {
      return operation.progress
    }
    
    // Otherwise use simple status-based progress
    switch (scrapingStatus) {
      case 'initializing': 
        return 5
      case 'scanning':
        return 10
      case 'downloading':
      case 'processing':
        // Smooth progress from 10-90% based on files processed
        // Approximately 3% per file, capped at 90%
        return Math.min(10 + filesProcessed * 3, 90)
      case 'completing':
        return 95
      case 'completed':
        return 100
      case 'stopped':
        // Keep whatever progress we had
        return Math.min(10 + filesProcessed * 3, 90)
      default:
        return 0
    }
  }, [operation.progress, scrapingStatus, filesProcessed])
  
  // Generate clear status message
  const getMessage = () => {
    // Use backend message if available
    if (operation.message) {
      return operation.message
    }
    
    // Generate message based on status
    switch (scrapingStatus) {
      case 'initializing':
        return 'Starting data collection...'
      case 'scanning':
        return `Scanning for files...${filesProcessed > 0 ? ` Found ${filesProcessed} files` : ''}`
      case 'downloading':
        if (currentFile) {
          return `Downloading: ${currentFile}`
        }
        return `Processing files... (${filesProcessed} completed)`
      case 'processing':
        return `Processing file ${filesProcessed}${currentFile ? `: ${currentFile}` : '...'}`
      case 'completing':
        return 'Finalizing data collection...'
      case 'completed':
        return `✓ Successfully processed ${filesProcessed} file${filesProcessed !== 1 ? 's' : ''}`
      case 'stopped':
        return `Stopped after processing ${filesProcessed} file${filesProcessed !== 1 ? 's' : ''}`
      default:
        return 'Waiting...'
    }
  }
  
  // Format file name for display
  const formatFileName = (filename?: string) => {
    if (!filename) return ''
    // Extract just the date from "2025 01 15 ISX Daily Report.xlsx"
    const match = filename.match(/(\d{4}\s+\d{2}\s+\d{2})/)
    if (match) {
      return match[1].replace(/\s+/g, '-')
    }
    // Truncate long filenames
    if (filename.length > 30) {
      return '...' + filename.slice(-27)
    }
    return filename
  }
  
  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <StatusIcon className={cn(
              "h-5 w-5",
              config.color,
              config.iconAnimation
            )} />
            Data Collection Progress
          </CardTitle>
          <Badge 
            variant="outline" 
            className={cn(config.color, config.bgColor, "font-medium")}
          >
            {config.label}
          </Badge>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Main Progress Bar */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span className="text-muted-foreground">Progress</span>
            <span className="font-medium">{progress}%</span>
          </div>
          <Progress 
            value={progress} 
            className="h-2"
          />
        </div>
        
        {/* Status Message */}
        <div className="flex items-start gap-2">
          <FileText className="h-4 w-4 text-muted-foreground mt-0.5" />
          <p className="text-sm leading-relaxed">{getMessage()}</p>
        </div>
        
        {/* Simple Stats - Only show what we know */}
        {(filesProcessed > 0 || speed) && (
          <div className="flex items-center gap-4 text-xs text-muted-foreground border-t pt-3">
            {filesProcessed > 0 && (
              <div className="flex items-center gap-1">
                <FileText className="h-3 w-3" />
                <span>Files: {filesProcessed}</span>
              </div>
            )}
            {speed && speed > 0 && (
              <div className="flex items-center gap-1">
                <Zap className="h-3 w-3" />
                <span>{speed.toFixed(1)} files/min</span>
              </div>
            )}
            {currentFile && scrapingStatus === 'downloading' && (
              <div className="flex items-center gap-1 ml-auto">
                <Download className="h-3 w-3" />
                <span className="truncate max-w-[150px]">
                  {formatFileName(currentFile)}
                </span>
              </div>
            )}
          </div>
        )}
        
        {/* Completion indicator for edge cases */}
        {scrapingStatus === 'completing' && (
          <div className="flex items-center gap-2 text-xs text-orange-600 animate-pulse">
            <Clock className="h-3 w-3" />
            <span>Checking for additional files...</span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default ScrapingProgress