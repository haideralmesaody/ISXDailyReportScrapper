/**
 * Virtual Report List Component
 * Implements virtual scrolling for performance with large report lists
 * Following CLAUDE.md performance optimization standards
 */

'use client'

import React, { useRef, useState, useEffect, useCallback } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Loader2, FileSpreadsheet, AlertCircle } from 'lucide-react'
import { ReportCard } from './ReportCard'
import { downloadReportFile } from '@/lib/api/reports'
import { useToast } from '@/lib/hooks/use-toast'
import type { ReportMetadata } from '@/types/reports'

export interface VirtualReportListProps {
  reports: ReportMetadata[]
  selectedReport: ReportMetadata | null
  onSelectReport: (report: ReportMetadata) => void
  onDownloadReport?: (report: ReportMetadata) => void
  isLoading?: boolean
  height?: number | string
}

export function VirtualReportList({
  reports,
  selectedReport,
  onSelectReport,
  onDownloadReport,
  isLoading = false,
  height = 600
}: VirtualReportListProps) {
  const parentRef = useRef<HTMLDivElement>(null)
  const { toast } = useToast()
  
  // Handle download with default implementation if not provided
  const handleDownload = useCallback(async (report: ReportMetadata) => {
    if (onDownloadReport) {
      onDownloadReport(report)
    } else {
      try {
        await downloadReportFile(report.path || report.name)
        toast({
          title: 'Success',
          description: `Downloaded ${report.displayName}`,
        })
      } catch (err) {
        console.error('Failed to download report:', err)
        toast({
          title: 'Error',
          description: 'Failed to download report. Please try again.',
          variant: 'destructive',
        })
      }
    }
  }, [onDownloadReport, toast])
  
  // Initialize virtual list
  const virtualizer = useVirtualizer({
    count: reports.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 120, // Estimated height of each ReportCard
    overscan: 5, // Number of items to render outside of the visible area
    gap: 8, // Gap between items
  })
  
  // Auto-scroll to selected report
  useEffect(() => {
    if (selectedReport) {
      const selectedIndex = reports.findIndex(r => r.name === selectedReport.name)
      if (selectedIndex !== -1) {
        virtualizer.scrollToIndex(selectedIndex, {
          align: 'center',
          behavior: 'smooth'
        })
      }
    }
  }, [selectedReport, reports, virtualizer])
  
  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
          <p className="text-sm text-muted-foreground">Loading reports...</p>
        </div>
      </div>
    )
  }
  
  // Empty state
  if (reports.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-96 text-center">
        <FileSpreadsheet className="h-12 w-12 text-muted-foreground mb-4" />
        <h3 className="font-semibold text-lg mb-2">No Reports Found</h3>
        <p className="text-sm text-muted-foreground max-w-md">
          No reports match the current filters. Try adjusting your search criteria or clearing filters.
        </p>
      </div>
    )
  }
  
  const virtualItems = virtualizer.getVirtualItems()
  const totalHeight = virtualizer.getTotalSize()
  
  return (
    <div className="relative h-full flex flex-col">
      {/* Report count indicator */}
      <div className="mb-2 flex items-center justify-between flex-shrink-0">
        <p className="text-sm text-muted-foreground">
          {reports.length} report{reports.length !== 1 ? 's' : ''}
        </p>
        {reports.length > 20 && (
          <p className="text-xs text-muted-foreground">
            Scroll to see more
          </p>
        )}
      </div>
      
      {/* Virtual scroll container */}
      <div
        ref={parentRef}
        className="overflow-auto relative flex-1"
        style={{ height: typeof height === 'number' ? `${height}px` : height }}
      >
        <div
          style={{
            height: `${totalHeight}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          {virtualItems.map((virtualItem) => {
            const report = reports[virtualItem.index]
            if (!report) return null
            
            return (
              <div
                key={virtualItem.key}
                data-index={virtualItem.index}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  height: `${virtualItem.size}px`,
                  transform: `translateY(${virtualItem.start}px)`,
                }}
              >
                <ReportCard
                  report={report}
                  isSelected={selectedReport?.name === report.name}
                  onSelect={() => onSelectReport(report)}
                  onDownload={() => handleDownload(report)}
                />
              </div>
            )
          })}
        </div>
      </div>
      
      {/* Performance indicator for large lists */}
      {reports.length > 100 && (
        <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
          <AlertCircle className="h-3 w-3" />
          <span>Virtual scrolling enabled for performance</span>
        </div>
      )}
    </div>
  )
}