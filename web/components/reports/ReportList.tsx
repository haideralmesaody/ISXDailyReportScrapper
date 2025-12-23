/**
 * Report List Component
 * Displays filtered list of reports with search and actions
 * Following CLAUDE.md performance and UX standards
 */

'use client'

import React, { useState, useMemo } from 'react'
import { Search, FileX } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { ReportCard } from './ReportCard'
import { VirtualReportList } from './VirtualReportList'
import { filterReports, sortReports } from '@/lib/utils/csv-parser'
import type { ReportListProps } from '@/types/reports'

export function ReportList({
  reports,
  selectedReport,
  onSelectReport,
  onDownloadReport,
  isLoading
}: ReportListProps) {
  const [searchTerm, setSearchTerm] = useState('')
  
  // Filter and sort reports
  const filteredReports = useMemo(() => {
    const filtered = filterReports(reports, searchTerm)
    return sortReports(filtered)
  }, [reports, searchTerm])
  
  // Use virtual scrolling for large lists (> 50 reports)
  const useVirtualScroll = filteredReports.length > 50
  
  // Loading state
  if (isLoading) {
    return (
      <div className="flex flex-col h-full">
        <div className="relative mb-4">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            disabled
            placeholder="Loading reports..."
            className="pl-10"
          />
        </div>
        <div className="flex-1 overflow-auto">
          <div className="space-y-3 pr-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <ReportListSkeleton key={i} />
            ))}
          </div>
        </div>
      </div>
    )
  }
  
  // No reports state - provide helpful guidance
  if (reports.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 text-center space-y-4">
        <div className="p-3 bg-muted rounded-full">
          <FileX className="h-10 w-10 text-muted-foreground" />
        </div>
        <div className="space-y-2">
          <h3 className="font-semibold text-lg">No Reports Available</h3>
          <p className="text-sm text-muted-foreground max-w-xs">
            You need to run data collection operations first to generate reports.
          </p>
        </div>
        <div className="bg-blue-50 dark:bg-blue-950/20 rounded-lg p-4 max-w-sm">
          <p className="text-xs text-blue-700 dark:text-blue-400 font-medium mb-2">
            Quick Start Guide:
          </p>
          <ol className="text-xs text-left space-y-1 text-blue-600 dark:text-blue-300">
            <li>1. Go to Operations page</li>
            <li>2. Run "Full Pipeline" operation</li>
            <li>3. Wait for processing to complete</li>
            <li>4. Return here to view reports</li>
          </ol>
        </div>
        <a 
          href="/operations" 
          className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none ring-offset-background bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4"
        >
          Go to Operations
        </a>
      </div>
    )
  }
  
  return (
    <div className="flex flex-col h-full">
      {/* Search Input */}
      <div className="relative mb-4">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search reports..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="pl-10"
          aria-label="Search reports"
        />
      </div>
      
      {/* Results count */}
      {searchTerm && (
        <p className="text-sm text-muted-foreground mb-2">
          Found {filteredReports.length} of {reports.length} reports
        </p>
      )}
      
      {/* Reports List */}
      {filteredReports.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-sm text-muted-foreground">
            No reports match your search criteria
          </p>
        </div>
      ) : useVirtualScroll ? (
        // Use virtual scrolling for performance with large lists
        <div className="flex-1 min-h-0">
          <VirtualReportList
            reports={filteredReports}
            selectedReport={selectedReport}
            onSelectReport={onSelectReport}
            onDownloadReport={onDownloadReport}
            height={400} // Fixed height for consistency
          />
        </div>
      ) : (
        // Regular scrolling for smaller lists
        <div className="flex-1 overflow-auto">
          <div className="space-y-3 pr-4">
            {filteredReports.map((report) => (
              <ReportCard
                key={report.path || report.name}
                report={report}
                isSelected={selectedReport?.path === report.path || selectedReport?.name === report.name}
                onSelect={() => onSelectReport(report)}
                onDownload={() => onDownloadReport(report)}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

/**
 * Skeleton loader for report list items
 */
function ReportListSkeleton() {
  return (
    <div className="border rounded-lg p-4">
      <div className="flex items-start gap-3">
        <Skeleton className="h-9 w-9 rounded-lg" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-3 w-1/2" />
          <div className="flex items-center gap-3">
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-3 w-24" />
          </div>
        </div>
        <div className="flex flex-col gap-1">
          <Skeleton className="h-8 w-8" />
          <Skeleton className="h-8 w-8" />
        </div>
      </div>
    </div>
  )
}