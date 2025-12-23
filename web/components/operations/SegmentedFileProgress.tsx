/**
 * Enhanced Segmented File Progress Component
 * Shows exactly what the backend sends. No synthesized segments or approximations.
 */

'use client'

import React, { useMemo } from 'react'
import { cn } from '@/lib/utils'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Badge } from '@/components/ui/badge'
import { Clock, CheckCircle, XCircle, AlertCircle, FileText } from 'lucide-react'

type FileStatus = 'pending' | 'processing' | 'completed' | 'failed'

interface FileSegment {
  filename: string
  index: number
  status: FileStatus
  size_mb?: number
  progress?: number
  error_message?: string
  processing_time_ms?: number
}

interface SegmentedFileProgressProps {
  totalFiles?: number
  processedFiles?: number
  failedFiles?: number
  currentFile?: string
  fileStatuses?: FileSegment[]
  processingSpeedMBps?: number
  totalSizeMB?: number
  processedSizeMB?: number
  estimatedRemainingMs?: number
  className?: string
  showDetails?: boolean
  showIndices?: boolean
}

export function SegmentedFileProgress({
  totalFiles,
  processedFiles,
  failedFiles = 0,
  currentFile,
  fileStatuses,
  processingSpeedMBps,
  totalSizeMB,
  processedSizeMB,
  estimatedRemainingMs,
  className,
  showDetails = true,
  showIndices = true
}: SegmentedFileProgressProps) {

  const segments = useMemo(() => {
    if (fileStatuses && fileStatuses.length > 0) {
      return fileStatuses.map((status, index) => ({
        ...status,
        index,
        filename: status.filename || `File ${index + 1}`
      }))
    }
    return []
  }, [fileStatuses])

  const progress = useMemo(() => {
    if (typeof processedFiles === 'number' && typeof totalFiles === 'number' && totalFiles > 0) {
      const pct = Math.round((processedFiles / totalFiles) * 100)
      return Math.min(Math.max(pct, 0), 100)
    }
    return undefined
  }, [processedFiles, totalFiles])

  const statusCounts = useMemo(() => {
    const counts = {
      completed: 0,
      processing: 0,
      failed: 0,
      pending: 0
    }

    segments.forEach(segment => {
      counts[segment.status]++
    })

    return counts
  }, [segments])

  const formatTimeRemaining = (ms: number) => {
    if (ms <= 0) return 'Unknown'
    const seconds = Math.round(ms / 1000)
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.round(seconds / 60)
    if (minutes < 60) return `${minutes}m`
    const hours = Math.round(minutes / 60)
    return `${hours}h`
  }

  const formatFileSize = (mb: number) => {
    if (mb < 1) return `${Math.round(mb * 1024)}KB`
    if (mb < 1024) return `${Math.round(mb)}MB`
    return `${(mb / 1024).toFixed(1)}GB`
  }

  const hasSegments = segments.length > 0
  const displaySegmentWidth = hasSegments ? 100 / segments.length : 0
  const telemetryErrors: string[] = []
  if (showDetails) {
    if (!hasSegments) telemetryErrors.push('file_statuses missing')
    if (typeof totalFiles !== 'number') telemetryErrors.push('total_files missing')
    if (typeof processedFiles !== 'number') telemetryErrors.push('files_processed missing')
  }

  const getStatusInfo = (status: FileStatus) => {
    switch (status) {
      case 'completed':
        return { color: 'bg-green-500', icon: CheckCircle, label: 'Completed' }
      case 'processing':
        return { color: 'bg-blue-500 animate-pulse', icon: Clock, label: 'Processing' }
      case 'failed':
        return { color: 'bg-red-500', icon: XCircle, label: 'Failed' }
      case 'pending':
      default:
        return { color: 'bg-gray-200 dark:bg-gray-700', icon: FileText, label: 'Pending' }
    }
  }

  return (
    <div className={cn("space-y-4", className)}>
      {telemetryErrors.length > 0 && (
        <div className="flex items-center gap-2 text-xs text-red-600 bg-red-50 border border-red-200 rounded-md px-3 py-2">
          <AlertCircle className="h-3 w-3" />
          <span>Telemetry error: {telemetryErrors.join(', ')}</span>
        </div>
      )}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium">File Processing Progress</span>
          <span className="text-lg font-bold">{progress !== undefined ? `${progress}%` : '—'}</span>
          {typeof totalFiles === 'number' && (
            <Badge variant="outline" className="text-xs">
              {statusCounts.completed}/{totalFiles} files
            </Badge>
          )}
        </div>

        {processingSpeedMBps && processingSpeedMBps > 0 && (
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Clock className="h-3 w-3" />
            <span>{processingSpeedMBps.toFixed(1)} MB/s</span>
          </div>
        )}
      </div>

      {currentFile && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground bg-blue-50 dark:bg-blue-950/20 p-2 rounded-lg">
          <div className="h-2 w-2 rounded-full bg-blue-500 animate-pulse" />
          <span className="truncate font-medium">
            Currently: {currentFile.split('/').pop()?.split('\\').pop() || currentFile}
          </span>
        </div>
      )}

      {hasSegments ? (
        <div className="w-full">
          <TooltipProvider delayDuration={0}>
            <div className="flex h-4 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
              {segments.map((segment) => {
                const statusInfo = getStatusInfo(segment.status)
                const Icon = statusInfo.icon

                return (
                  <Tooltip key={`${segment.index}-${segment.filename}`}>
                    <TooltipTrigger asChild>
                      <div
                      className={cn(
                        "h-full hover:opacity-80 cursor-default relative",
                        statusInfo.color,
                        segment.index > 0 && "border-l border-gray-300/20 dark:border-gray-600/20"
                      )}
                        style={{ width: `${displaySegmentWidth}%` }}
                      >
                        {showIndices && (
                          <div className="h-full flex items-center justify-center text-[10px] font-medium text-white">
                            {segment.index + 1}
                          </div>
                        )}
                      </div>
                    </TooltipTrigger>
                    <TooltipContent side="top" className="max-w-xs">
                      <div className="space-y-2">
                        <div className="flex items-center gap-2">
                          <Icon className="h-4 w-4" />
                          <div className="font-medium">
                            {segment.filename.split('/').pop()?.split('\\').pop() || segment.filename}
                          </div>
                        </div>

                        <div className="text-xs space-y-1">
                          <div className="flex items-center justify-between">
                            <span className="text-muted-foreground">Status:</span>
                            <Badge variant={segment.status === 'failed' ? 'destructive' : 'secondary'} className="text-xs">
                              {statusInfo.label}
                            </Badge>
                          </div>

                          {segment.size_mb && (
                            <div className="flex items-center justify-between">
                              <span className="text-muted-foreground">Size:</span>
                              <span>{formatFileSize(segment.size_mb)}</span>
                            </div>
                          )}

                          {segment.progress !== undefined && segment.status === 'processing' && (
                            <div className="flex items-center justify-between">
                              <span className="text-muted-foreground">Progress:</span>
                              <span>{Math.round(segment.progress)}%</span>
                            </div>
                          )}

                          {segment.processing_time_ms && (
                            <div className="flex items-center justify-between">
                              <span className="text-muted-foreground">Time:</span>
                              <span>{formatTimeRemaining(segment.processing_time_ms)}</span>
                            </div>
                          )}

                          {segment.error_message && (
                            <div className="text-red-600 dark:text-red-400 text-xs p-1 bg-red-50 dark:bg-red-950/20 rounded">
                              {segment.error_message}
                            </div>
                          )}
                        </div>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                )
              })}
            </div>
          </TooltipProvider>
        </div>
      ) : (
        // Show generic progress bar when no file statuses
        typeof totalFiles === 'number' && totalFiles > 0 ? (
          <div className="w-full">
            <TooltipProvider delayDuration={0}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex h-4 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
                    <div
                      className={cn(
                        "h-full transition-all duration-300 ease-out",
                        currentFile ? "bg-blue-500" : "bg-green-500"
                      )}
                      style={{ width: `${progress !== undefined ? progress : 0}%` }}
                    />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  <div className="text-sm">
                    <p>Progress: {progress !== undefined ? `${progress}%` : '—'}</p>
                    {processedFiles !== undefined && totalFiles && (
                      <p>Files: {processedFiles}/{totalFiles}</p>
                    )}
                  </div>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        ) : (
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <AlertCircle className="h-3 w-3" />
            <span>No file_statuses provided by backend.</span>
          </div>
        )
      )}

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
        <div className="flex items-center gap-2 bg-green-50 dark:bg-green-950/20 p-2 rounded-lg">
          <div className="h-2 w-2 rounded-sm bg-green-500" />
          <div>
            <div className="font-medium text-green-700 dark:text-green-300">
              {statusCounts.completed}
            </div>
            <div className="text-muted-foreground">Completed</div>
          </div>
        </div>

        {statusCounts.processing > 0 && (
          <div className="flex items-center gap-2 bg-blue-50 dark:bg-blue-950/20 p-2 rounded-lg">
            <div className="h-2 w-2 rounded-sm bg-blue-500 animate-pulse" />
            <div>
              <div className="font-medium text-blue-700 dark:text-blue-300">
                {statusCounts.processing}
              </div>
              <div className="text-muted-foreground">Processing</div>
            </div>
          </div>
        )}

        {statusCounts.failed > 0 && (
          <div className="flex items-center gap-2 bg-red-50 dark:bg-red-950/20 p-2 rounded-lg">
            <div className="h-2 w-2 rounded-sm bg-red-500" />
            <div>
              <div className="font-medium text-red-700 dark:text-red-300">
                {statusCounts.failed}
              </div>
              <div className="text-muted-foreground">Failed</div>
            </div>
          </div>
        )}

        <div className="flex items-center gap-2 bg-gray-50 dark:bg-gray-950/20 p-2 rounded-lg">
          <div className="h-2 w-2 rounded-sm bg-gray-300 dark:bg-gray-600" />
          <div>
            <div className="font-medium text-gray-700 dark:text-gray-300">
              {statusCounts.pending}
            </div>
            <div className="text-muted-foreground">Pending</div>
          </div>
        </div>
      </div>

      {showDetails && (totalSizeMB || processedSizeMB || estimatedRemainingMs) && (
        <div className="bg-gray-50 dark:bg-gray-900/50 p-3 rounded-lg">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
            {totalSizeMB && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Total Size:</span>
                <span className="font-medium">{formatFileSize(totalSizeMB)}</span>
              </div>
            )}

            {processedSizeMB && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Processed:</span>
                <span className="font-medium">{formatFileSize(processedSizeMB)}</span>
              </div>
            )}

            {estimatedRemainingMs && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Est. Remaining:</span>
                <span className="font-medium">{formatTimeRemaining(estimatedRemainingMs)}</span>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
