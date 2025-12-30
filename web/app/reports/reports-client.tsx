/**
 * Reports Client Component
 * Main reports page with interactive features
 * Following CLAUDE.md hydration best practices
 */

'use client'

import React, { useState, useEffect, useCallback, useMemo } from 'react'
import Link from 'next/link'
import { ArrowLeft, FileText, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { NoDataState, DataLoadingState } from '@/components/ui'
import { useToast } from '@/lib/hooks/use-toast'
import { useHydration } from '@/lib/hooks'
import { ReportTypeSelector } from '@/components/reports/ReportTypeSelector'
import { ReportList } from '@/components/reports/ReportList'
import { CSVViewer } from '@/components/reports/CSVViewer'
import { ReportFilters } from '@/components/reports/ReportFilters'
import { HelpButton } from '@/components/guide/HelpButton'
import { 
  fetchReports, 
  downloadReportContent, 
  downloadReportFile 
} from '@/lib/api/reports'
import { parseCSVContent, groupReportsByType } from '@/lib/utils/csv-parser'
import type { 
  ReportMetadata, 
  ReportType, 
  ParsedCSVData 
} from '@/types/reports'
import { 
  trackNoDataResolved, 
  trackRetryAttempt, 
  debug
} from '@/lib/observability/no-data-metrics'

export default function ReportsClient() {
  // Hydration check to prevent SSR/client mismatch
  const isHydrated = useHydration()
  
  // State management
  const [reports, setReports] = useState<ReportMetadata[]>([])
  const [filteredReports, setFilteredReports] = useState<ReportMetadata[]>([])
  const [selectedType, setSelectedType] = useState<ReportType>('all')
  const [selectedReport, setSelectedReport] = useState<ReportMetadata | null>(null)
  const [csvData, setCSVData] = useState<ParsedCSVData | null>(null)
  const [isLoadingReports, setIsLoadingReports] = useState(false)
  const [isLoadingCSV, setIsLoadingCSV] = useState(false)
  const [error, setError] = useState<Error | null>(null)
  const [showFilters, setShowFilters] = useState(false)
  const [retryCount, setRetryCount] = useState(0)
  const [loadStartTime, setLoadStartTime] = useState<number | null>(null)
  
  const { toast } = useToast()
  
  // Filter reports by selected type (base filtering)
  const typeFilteredReports = useMemo(() => {
    if (selectedType === 'all') return reports
    return reports.filter(report => report.type === selectedType)
  }, [reports, selectedType])
  
  // Calculate report counts by type
  const reportCounts = useMemo(() => {
    const grouped = groupReportsByType(reports)
    const counts: Record<ReportType, number> = {
      all: reports.length,
      daily: grouped.get('daily')?.length || 0,
      ticker: grouped.get('ticker')?.length || 0,
      index: grouped.get('index')?.length || 0,
      liquidity: grouped.get('liquidity')?.length || 0,
      combined: grouped.get('combined')?.length || 0,
      indexes: grouped.get('indexes')?.length || 0,
      summary: grouped.get('summary')?.length || 0,
    }
    return counts
  }, [reports])
  
  // Handle advanced filtering
  const handleFiltersChange = useCallback((filtered: ReportMetadata[]) => {
    setFilteredReports(filtered)
  }, [])
  
  // Set initial filtered reports when type changes
  useEffect(() => {
    setFilteredReports(typeFilteredReports)
  }, [typeFilteredReports])
  
  // Load reports on mount
  useEffect(() => {
    if (!isHydrated) return
    loadReports()
  }, [isHydrated])
  
  // Load reports from API
  const loadReports = useCallback(async (isRetry = false) => {
    setIsLoadingReports(true)
    setError(null)
    
    // Track start time for performance measurement
    const startTime = Date.now()
    if (!isRetry) {
      setLoadStartTime(startTime)
    }
    
    try {
      debug.logApiResponse('/api/reports', null, false)
      
      const fetchedReports = await fetchReports()
      setReports(fetchedReports)
      
      // Track successful data resolution if this was after an error state
      if (isRetry || loadStartTime) {
        const resolutionTime = Date.now() - (loadStartTime || startTime)
        trackNoDataResolved('reports', resolutionTime, isRetry ? 'retry' : 'api_success')
        
        debug.logPerformance('Reports Data Load', resolutionTime, {
          is_retry: isRetry,
          report_count: fetchedReports.length,
          retry_count: retryCount
        })
      }
      
      // Select first report if available
      if (fetchedReports.length > 0 && !selectedReport) {
        const firstReport = fetchedReports[0]
        if (firstReport) {
          handleSelectReport(firstReport)
        }
      }
      
      // Reset retry count on success
      if (isRetry) {
        setRetryCount(0)
      }
    } catch (err) {
      console.error('Failed to load reports:', err)
      const error = err instanceof Error ? err : new Error('Failed to load reports')
      
      debug.logApiResponse('/api/reports', err, true)
      
      // Track retry attempt if this is a retry
      if (isRetry) {
        trackRetryAttempt('reports', retryCount + 1, false)
        setRetryCount(prev => prev + 1)
      }
      
      // Don't show error state or toast for 404 (no data scenario)
      const isNotFound = error.message.includes('404') || error.message.includes('Not Found')
      if (!isNotFound) {
        setError(error)
        toast({
          title: 'Error',
          description: 'Failed to load reports. Please try again.',
          variant: 'destructive',
        })
      }
    } finally {
      setIsLoadingReports(false)
    }
  }, [selectedReport, toast, loadStartTime, retryCount])
  
  // Handle report selection
  const handleSelectReport = useCallback(async (report: ReportMetadata) => {
    setSelectedReport(report)
    setIsLoadingCSV(true)
    setError(null)
    setCSVData(null)
    
    try {
      // Fetch CSV content using the path (supports nested paths)
      const content = await downloadReportContent(report.path || report.name)
      
      // Parse CSV data
      const parsed = await parseCSVContent(content)
      setCSVData(parsed)
    } catch (err) {
      console.error('Failed to load CSV:', err)
      // Don't set error state if it's just a 404 (no data)
      // Only set error for real errors
      const errorMessage = err instanceof Error ? err.message : 'Failed to load CSV data'
      if (!errorMessage.includes('404') && !errorMessage.includes('Not Found')) {
        setError(err instanceof Error ? err : new Error('Failed to load CSV data'))
        toast({
          title: 'Error',
          description: 'Failed to load report data. Please try again.',
          variant: 'destructive',
        })
      }
      // If it's a 404, just leave the CSV viewer empty (no data to show)
    } finally {
      setIsLoadingCSV(false)
    }
  }, [toast])
  
  // Handle report download
  const handleDownloadReport = useCallback(async (report: ReportMetadata) => {
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
  }, [toast])
  
  // Handle type change
  const handleTypeChange = useCallback((type: ReportType) => {
    setSelectedType(type)
    
    // Clear selection if current report doesn't match new type
    if (selectedReport && type !== 'all' && selectedReport.type !== type) {
      setSelectedReport(null)
      setCSVData(null)
    }
  }, [selectedReport])
  
  // Loading state before hydration
  if (!isHydrated) {
    return (
      <DataLoadingState
        message="Setting up the reports interface..."
        page="reports"
        operation="hydration"
        trackPerformance={true}
      />
    )
  }
  
  // No reports available (after loading completes)
  if (!isLoadingReports && reports.length === 0 && !error) {
    const actions = [
      {
        label: 'Go to Operations',
        variant: 'default' as const,
        href: '/operations'
      },
      {
        label: 'Check Again',
        variant: 'outline' as const,
        onClick: () => loadReports(true),
        icon: RefreshCw
      }
    ]
    
    return (
      <div className="h-screen bg-background flex flex-col overflow-hidden">
        {/* Fixed Header */}
        <header className="flex-shrink-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b">
          <div className="px-6 py-4">
            <Button variant="ghost" size="sm" className="mb-2" asChild>
              <Link href="/">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Home
              </Link>
            </Button>
            
            <div className="flex items-center justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-2xl md:text-3xl font-bold">Reports</h1>
                  <HelpButton
                    section="reports"
                    tooltip="Learn about reports and data access"
                  />
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  View and download Iraqi Stock Exchange reports
                </p>
              </div>
            </div>
          </div>
        </header>
        
        {/* No Data State */}
        <main className="flex-1 overflow-hidden">
          <NoDataState
            icon={FileText}
            iconColor="green"
            title="No Reports Available"
            description="You need to run the data collection operations first to generate reports."
            page="reports"
            reason="no_reports_available"
            componentName="ReportsNoDataState"
            instructions={[
              "Go to the Operations page",
              "Run 'Full Pipeline' to collect and process data",
              "Reports will be automatically generated",
              "Return here to view and download reports"
            ]}
            actions={actions}
          />
        </main>
        
        {/* Fixed Footer */}
        <footer className="flex-shrink-0 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <div className="px-6 py-2">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center gap-3">
                <span>0 reports available</span>
              </div>
              <div className="flex items-center gap-3">
                <span>ISX Pulse v3.0.0</span>
                <span className="text-muted-foreground/50">•</span>
                <span>© 2025 ISX Daily Reports</span>
              </div>
            </div>
          </div>
        </footer>
      </div>
    )
  }
  
  return (
    <div className="h-screen bg-background flex flex-col overflow-hidden">
      {/* Fixed Header */}
      <header className="flex-shrink-0 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b">
        <div className="px-6 py-4">
          <Button variant="ghost" size="sm" className="mb-2" asChild>
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Home
            </Link>
          </Button>
          
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl md:text-3xl font-bold">Reports</h1>
                <HelpButton
                  section="reports"
                  tooltip="Learn about reports and data access"
                />
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                View and download Iraqi Stock Exchange reports
              </p>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area - Takes remaining space between header and footer */}
      <main className="flex-1 overflow-hidden">
        {/* Filters Section */}
        {showFilters && (
          <div className="px-6 py-4 border-b">
            <ReportFilters
              reports={typeFilteredReports}
              onFiltersChange={handleFiltersChange}
            />
          </div>
        )}
        
        {/* Main Content Grid - Full width, no padding */}
        <div className="h-full flex">
          {/* Left Panel - Fixed width, against left edge */}
          <div className="w-[400px] xl:w-[450px] flex-shrink-0 border-r bg-card/50 flex flex-col">
            <div className="p-4 border-b">
              <div className="space-y-4">
                <ReportTypeSelector
                  selectedType={selectedType}
                  onTypeChange={handleTypeChange}
                  counts={reportCounts}
                />
                
                {/* Toggle Filters Button */}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowFilters(!showFilters)}
                  className="w-full"
                >
                  {showFilters ? 'Hide' : 'Show'} Advanced Filters
                </Button>
              </div>
            </div>
            
            <div className="flex-1 flex flex-col overflow-hidden">
              <div className="p-4 pb-2 flex-shrink-0 border-b">
                <h2 className="font-semibold">Available Reports</h2>
              </div>
              <div className="flex-1 overflow-auto">
                <div className="p-4">
                  {isLoadingReports ? (
                    <DataLoadingState
                      message="Fetching available reports..."
                      showCard={false}
                      size="sm"
                      page="reports"
                      operation="report_loading"
                      trackPerformance={true}
                    />
                  ) : (
                    <ReportList
                      reports={filteredReports}
                      selectedReport={selectedReport}
                      onSelectReport={handleSelectReport}
                      onDownloadReport={handleDownloadReport}
                      isLoading={false}
                    />
                  )}
                </div>
              </div>
            </div>
          </div>
          
          {/* Right Panel - Takes all remaining space */}
          <div className="flex-1 overflow-hidden">
            <CSVViewer
              report={selectedReport}
              csvData={csvData}
              isLoading={isLoadingCSV}
              error={error}
            />
          </div>
        </div>
      </main>
      
      {/* Fixed Footer - Always visible at bottom */}
      <footer className="flex-shrink-0 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
        <div className="px-6 py-2">
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center gap-3">
              <span>{filteredReports.length} of {reports.length} reports</span>
              {selectedReport && (
                <>
                  <span className="text-muted-foreground/50">•</span>
                  <span className="font-medium">{selectedReport.displayName}</span>
                </>
              )}
              {csvData && (
                <>
                  <span className="text-muted-foreground/50">•</span>
                  <span>{csvData.data.length} rows</span>
                </>
              )}
            </div>
            <div className="flex items-center gap-3">
              <span>ISX Pulse v3.0.0</span>
              <span className="text-muted-foreground/50">•</span>
              <span>© 2025 ISX Daily Reports</span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
