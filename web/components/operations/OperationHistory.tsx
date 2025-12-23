/**
 * OperationHistory Component
 * 
 * AGENT REQUIREMENTS FOR MODIFICATIONS:
 * =====================================
 * 
 * PRIMARY AGENTS:
 * - frontend-modernizer: For UI components, state management, React patterns
 * - database-optimizer: For query optimization, pagination strategies
 * 
 * SECONDARY AGENTS:
 * - test-architect: For pagination, filtering, and export testing
 * - go-architect: For architectural improvements and data flow
 * - api-integration-specialist: For API interactions and data fetching
 * 
 * TASK-SPECIFIC AGENT ASSIGNMENTS:
 * - Pagination improvements → frontend-modernizer
 * - Filter optimization → database-optimizer → frontend-modernizer
 * - Export functionality → api-integration-specialist → frontend-modernizer
 * - Analytics/charts → financial-report-generator → frontend-modernizer
 * - Performance (large lists) → frontend-modernizer (virtualization)
 * - API throttling → api-integration-specialist
 * - State management → frontend-modernizer
 * - Query optimization → database-optimizer
 * 
 * QUALITY GATES:
 * 1. Pagination changes require database-optimizer review
 * 2. Export features need api-integration-specialist validation
 * 3. UI updates require frontend-modernizer approval
 * 4. Performance improvements need metrics from observability-engineer
 * 
 * OPTIMIZATION OPPORTUNITIES:
 * - Add virtual scrolling for lists > 100 items (frontend-modernizer)
 * - Implement query caching strategy (database-optimizer)
 * - Add chart visualizations (financial-report-generator)
 * - Optimize filter queries (database-optimizer)
 * 
 * @see .claude/agents-workflow.md for detailed agent selection guide
 */

'use client'

import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react'
import { apiClient } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  ChevronDown,
  ChevronUp,
  Download,
  Clock,
  User,
  AlertCircle,
  CheckCircle,
  XCircle,
  ChevronLeft,
  ChevronRight,
  BarChart3,
  Loader2,
  Globe,
  Webhook
} from 'lucide-react'
import type { OperationRun, OperationStatus, TriggerType } from '@/types'

// Helper functions hoisted outside component for performance
const getStatusColor = (status: OperationStatus): string => {
  switch (status) {
    case 'completed': return 'text-green-600'
    case 'failed': return 'text-red-600'
    case 'running': return 'text-blue-600'
    case 'cancelled': return 'text-orange-600'
    default: return 'text-gray-600'
  }
}

const getStatusIcon = (status: OperationStatus) => {
  switch (status) {
    case 'completed': return <CheckCircle className="h-4 w-4" strokeWidth={1.5} />
    case 'failed': return <XCircle className="h-4 w-4" strokeWidth={1.5} />
    case 'running': return <Clock className="h-4 w-4 animate-spin" strokeWidth={1.5} />
    default: return <Clock className="h-4 w-4" strokeWidth={1.5} />
  }
}

const formatDuration = (seconds: number): string => {
  if (seconds < 60) return `${seconds}s`
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60
  
  if (hours > 0) {
    return `${hours}h ${minutes}m${secs > 0 ? ` ${secs}s` : ''}`
  }
  
  return secs > 0 ? `${minutes}m ${secs}s` : `${minutes}m`
}

interface OperationHistoryProps {
  operationId: string
  operationName: string
  enableRealtime?: boolean
  showAnalytics?: boolean
}

export function OperationHistory({
  operationId,
  operationName: _operationName,
  enableRealtime = false,
  showAnalytics = false
}: OperationHistoryProps): JSX.Element {
  // State management
  const [history, setHistory] = useState<OperationRun[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedRuns, setExpandedRuns] = useState<Set<string>>(new Set())
  
  // Pagination
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [totalItems, setTotalItems] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  
  // Abort controller for preventing concurrent fetches
  const abortControllerRef = useRef<AbortController | null>(null)
  const lastFetchTimeRef = useRef<number>(0)
  const loadingRef = useRef(false)
  
  // Filters
  const [statusFilter, setStatusFilter] = useState<OperationStatus | ''>('')
  const [triggerFilter, setTriggerFilter] = useState<TriggerType | ''>('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<string>('startedAt')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  
  // Export state
  const [exporting, setExporting] = useState(false)
  const [exportMenuOpen, setExportMenuOpen] = useState(false)
  
  // Sync loading state to ref for accurate throttle checking
  useEffect(() => {
    loadingRef.current = loading
  }, [loading])

  // Fetch history
  const fetchHistory = useCallback(async () => {
    // Strict throttle - maximum 1 request per second
    const now = Date.now()
    if ((now - lastFetchTimeRef.current) < 1000) {
      return
    }
    lastFetchTimeRef.current = now
    
    // Abort any in-flight request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    const controller = new AbortController()
    abortControllerRef.current = controller
    
    try {
      setLoading(true)
      loadingRef.current = true
      setError(null)
      
      const params: any = {
        operationId,
        page,
        pageSize,
        sortBy,
        sortOrder
      }
      
      if (statusFilter) params.status = statusFilter
      if (triggerFilter) params.triggeredBy = triggerFilter
      if (startDate) params.startDate = new Date(startDate).toISOString()
      if (endDate) params.endDate = new Date(endDate + 'T23:59:59.999Z').toISOString()
      if (searchQuery) params.search = searchQuery
      
      const response = await apiClient.getOperationHistory(params)
      
      setHistory(response.items)
      setTotalItems(response.total)
      setTotalPages(response.totalPages || Math.ceil(response.total / pageSize))
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setError(err instanceof Error ? err.message : 'Failed to load history')
      }
    } finally {
      if (abortControllerRef.current === controller) {
        abortControllerRef.current = null
      }
      setLoading(false)
      loadingRef.current = false
    }
  }, [operationId, page, pageSize, statusFilter, triggerFilter, startDate, endDate, searchQuery, sortBy, sortOrder])

  // Initial load and refresh on filter changes
  useEffect(() => {
    fetchHistory()
  }, [fetchHistory])
  
  // Cleanup abort controller on unmount
  useEffect(() => {
    return () => {
      abortControllerRef.current?.abort()
    }
  }, [])

  // Real-time updates polling (if enabled)
  useEffect(() => {
    if (!enableRealtime) return
    
    const interval = setInterval(() => {
      fetchHistory()
    }, 5000) // Poll every 5 seconds
    
    return () => clearInterval(interval)
  }, [enableRealtime, fetchHistory])

  // Close export menu on outside click
  useEffect(() => {
    if (!exportMenuOpen) return
    
    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      if (!target.closest('[data-export-menu]')) {
        setExportMenuOpen(false)
      }
    }
    
    document.addEventListener('click', handleClickOutside)
    return () => document.removeEventListener('click', handleClickOutside)
  }, [exportMenuOpen])

  // Close export menu on Escape key
  useEffect(() => {
    if (!exportMenuOpen) return
    
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setExportMenuOpen(false)
      }
    }
    
    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [exportMenuOpen])

  // Toggle run expansion
  const toggleRunExpansion = (runId: string) => {
    setExpandedRuns(prev => {
      const next = new Set(prev)
      if (next.has(runId)) {
        next.delete(runId)
      } else {
        next.add(runId)
      }
      return next
    })
  }

  // Clear filters
  const clearFilters = () => {
    setStatusFilter('')
    setTriggerFilter('')
    setStartDate('')
    setEndDate('')
    setSearchQuery('')
    setSortBy('startedAt')
    setSortOrder('desc')
    setPage(1)
  }

  // Export history
  const exportHistory = async (format: 'csv' | 'json') => {
    try {
      setExporting(true)
      const filters: any = {}
      if (statusFilter) filters.status = statusFilter
      if (triggerFilter) filters.triggeredBy = triggerFilter
      if (startDate) filters.startDate = startDate
      if (endDate) filters.endDate = endDate
      if (searchQuery) filters.search = searchQuery
      
      const blob = await apiClient.exportOperationHistory({
        operationId,
        format,
        filters
      })
      
      // Create download link
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `operation-history.${format}`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed')
    } finally {
      setExporting(false)
    }
  }

  // Calculate analytics - optimized single pass
  const analytics = useMemo(() => {
    if (!showAnalytics || history.length === 0) return null
    
    let success = 0, fail = 0, durSum = 0, durCount = 0
    const triggerDist: Record<string, number> = {}
    
    for (const r of history) {
      if (r.status === 'completed') success++
      else if (r.status === 'failed') fail++
      
      if (r.duration) {
        durSum += r.duration
        durCount++
      }
      
      triggerDist[r.triggeredBy] = (triggerDist[r.triggeredBy] ?? 0) + 1
    }
    
    const total = history.length
    return {
      successRate: ((success / total) * 100).toFixed(1),
      averageDuration: durCount > 0 ? Math.round(durSum / durCount / 60) : null,
      totalRuns: total,
      triggerDistribution: triggerDist
    }
  }, [history, showAnalytics])

  // Build filter description
  const filterDescription = useMemo(() => {
    const parts = []
    if (statusFilter) parts.push(`${statusFilter} status`)
    if (triggerFilter) parts.push(`triggered by ${triggerFilter}`)
    if (startDate && endDate) {
      parts.push(`between ${new Date(startDate).toLocaleDateString()} and ${new Date(endDate).toLocaleDateString()}`)
    } else if (startDate) {
      parts.push(`after ${new Date(startDate).toLocaleDateString()}`)
    } else if (endDate) {
      parts.push(`before ${new Date(endDate).toLocaleDateString()}`)
    }
    if (searchQuery) parts.push(`matching "${searchQuery}"`)
    
    if (parts.length === 0) return ''
    if (parts.length === 1) return `Showing runs with ${parts[0]}`
    if (parts.length === 2) return `Showing runs with ${parts[0]} and ${parts[1]}`
    
    // Oxford comma for 3+ items
    const lastPart = parts.pop()
    return `Showing runs with ${parts.join(', ')}, and ${lastPart}`
  }, [statusFilter, triggerFilter, startDate, endDate, searchQuery])

  // Helper functions are now hoisted outside component

  // Loading state
  if (loading && history.length === 0) {
    return (
      <div className="flex items-center justify-center p-8" data-testid="history-loading">
        <Loader2 className="h-8 w-8 animate-spin mr-2" />
        <span>Loading history...</span>
      </div>
    )
  }

  // Error state
  if (error && history.length === 0) {
    return (
      <Alert variant="destructive" data-testid="history-error">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription className="flex items-center justify-between">
          <span>{error}</span>
          <Button size="sm" onClick={fetchHistory}>Retry</Button>
        </AlertDescription>
      </Alert>
    )
  }

  // Empty state
  if (!loading && history.length === 0) {
    return (
      <Card data-testid="history-empty">
        <CardContent className="pt-6 text-center">
          <p className="text-muted-foreground">No operation history found</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Analytics Section */}
      {showAnalytics && analytics && (
        <Card data-testid="history-analytics">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Operation Analytics
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Success Rate</p>
                <p className="text-2xl font-bold">{analytics.successRate}%</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Average Duration</p>
                <p className="text-2xl font-bold">{analytics.averageDuration !== null ? `${analytics.averageDuration}m` : '–'}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Runs</p>
                <p className="text-2xl font-bold">{analytics.totalRuns}</p>
              </div>
            </div>
            
            {/* Charts would go here */}
            <div className="mt-4 space-y-2">
              <div data-testid="duration-trend-chart" className="h-32 bg-muted rounded flex items-center justify-center text-muted-foreground">
                Duration Trend Chart
              </div>
              <div data-testid="success-rate-chart" className="h-32 bg-muted rounded flex items-center justify-center text-muted-foreground">
                Success Rate Chart
              </div>
              <div data-testid="trigger-distribution-chart" className="h-32 bg-muted rounded flex items-center justify-center text-muted-foreground">
                Trigger Distribution Chart
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Filters */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>History</CardTitle>
            <div className="flex items-center gap-2">
              <Button size="sm" variant="outline" onClick={clearFilters}>
                Clear filters
              </Button>
              <div className="relative" data-export-menu>
                <Button 
                  size="sm" 
                  variant="outline" 
                  aria-haspopup="menu" 
                  aria-expanded={exportMenuOpen}
                  onClick={() => setExportMenuOpen(!exportMenuOpen)}
                  disabled={exporting}
                >
                  {exporting ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Exporting...
                    </>
                  ) : (
                    <>
                      <Download className="h-4 w-4 mr-2" />
                      Export
                    </>
                  )}
                </Button>
                <div 
                  className={`absolute right-0 mt-1 w-32 bg-white border rounded-md shadow-lg ${exportMenuOpen ? 'block' : 'hidden'}`}
                  role="menu"
                >
                  {['csv', 'json'].map(format => (
                    <button
                      key={format}
                      role="menuitem"
                      className="block w-full text-left px-4 py-2 hover:bg-gray-100 disabled:opacity-50"
                      onClick={() => {
                        exportHistory(format as 'csv' | 'json')
                        setExportMenuOpen(false)
                      }}
                      disabled={exporting}
                    >
                      Export as {format.toUpperCase()}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
            <div>
              <Label htmlFor="status-filter">Filter by status</Label>
              <select
                id="status-filter"
                className="w-full px-3 py-2 border rounded-md"
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value as OperationStatus | '')
                  setPage(1)
                }}
              >
                <option value="">All Status</option>
                <option value="completed">Completed</option>
                <option value="failed">Failed</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
            
            <div>
              <Label htmlFor="trigger-filter">Trigger type</Label>
              <select
                id="trigger-filter"
                className="w-full px-3 py-2 border rounded-md"
                value={triggerFilter}
                onChange={(e) => {
                  setTriggerFilter(e.target.value as TriggerType | '')
                  setPage(1)
                }}
              >
                <option value="">All Triggers</option>
                <option value="schedule">Schedule</option>
                <option value="manual">Manual</option>
                <option value="api">API</option>
                <option value="webhook">Webhook</option>
              </select>
            </div>
            
            <div>
              <Label htmlFor="start-date">Start date</Label>
              <Input
                id="start-date"
                type="date"
                value={startDate}
                onChange={(e) => {
                  setStartDate(e.target.value)
                  setPage(1)
                }}
              />
            </div>
            
            <div>
              <Label htmlFor="end-date">End date</Label>
              <Input
                id="end-date"
                type="date"
                value={endDate}
                onChange={(e) => {
                  setEndDate(e.target.value)
                  setPage(1)
                }}
              />
            </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="md:col-span-2">
              <Input
                placeholder="Search errors..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value)
                  setPage(1)
                }}
                className="w-full"
              />
            </div>
            
            <div>
              <Label htmlFor="sort-by">Sort by</Label>
              <div className="flex gap-2">
                <select
                  id="sort-by"
                  className="flex-1 px-3 py-2 border rounded-md"
                  value={sortBy}
                  onChange={(e) => {
                    setSortBy(e.target.value)
                    setPage(1)
                  }}
                >
                  <option value="startedAt">Start Time</option>
                  <option value="duration">Duration</option>
                  <option value="status">Status</option>
                </select>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc')
                    setPage(1)
                  }}
                  aria-label="Toggle sort order"
                >
                  {sortOrder === 'asc' ? <ChevronUp /> : <ChevronDown />}
                </Button>
              </div>
            </div>
          </div>
          
          {filterDescription && (
            <div role="status" className="text-sm text-muted-foreground mt-2">
              {filterDescription}
            </div>
          )}
        </CardContent>
      </Card>

      {/* History Table */}
      <Card>
        <CardContent className="p-0">
          <table 
            role="table" 
            aria-label="Operation history"
            className="w-full"
            data-testid="history-list"
          >
            <thead>
              <tr className="border-b">
                <th role="columnheader" className="text-left p-4">Status</th>
                <th role="columnheader" className="text-left p-4">Started</th>
                <th role="columnheader" className="text-left p-4">Duration</th>
                <th role="columnheader" className="text-left p-4">Trigger</th>
                <th role="columnheader" className="text-left p-4">Actions</th>
              </tr>
            </thead>
            <tbody>
              {history.map((run) => {
                const isExpanded = expandedRuns.has(run.id)
                
                return (
                  <React.Fragment key={run.id}>
                    <tr 
                      className="border-b hover:bg-muted/50"
                      data-testid={`history-item-${run.id}`}
                    >
                      <td className="p-4">
                        <Badge 
                          variant="outline" 
                          className={getStatusColor(run.status)}
                        >
                          <span className="flex items-center gap-1">
                            {getStatusIcon(run.status)}
                            {run.status.charAt(0).toUpperCase() + run.status.slice(1)}
                          </span>
                        </Badge>
                      </td>
                      <td className="p-4">
                        {new Date(run.startedAt).toLocaleString()}
                      </td>
                      <td className="p-4">
                        {run.duration != null ? formatDuration(run.duration) : '-'}
                      </td>
                      <td className="p-4">
                        <div className="flex items-center gap-2">
                          {run.triggeredBy === 'manual' && <User className="h-4 w-4" strokeWidth={1.5} />}
                          {run.triggeredBy === 'schedule' && <Clock className="h-4 w-4" strokeWidth={1.5} />}
                          {run.triggeredBy === 'api' && <Globe className="h-4 w-4" strokeWidth={1.5} />}
                          {run.triggeredBy === 'webhook' && <Webhook className="h-4 w-4" strokeWidth={1.5} />}
                          {!['manual', 'schedule', 'api', 'webhook'].includes(run.triggeredBy) && <Clock className="h-4 w-4" strokeWidth={1.5} />}
                          <span>{run.triggeredBy.charAt(0).toUpperCase() + run.triggeredBy.slice(1)}</span>
                          {run.triggeredByUser && (
                            <span className="text-sm text-muted-foreground">
                              ({run.triggeredByUser})
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="p-4">
                        <Button
                          size="sm"
                          variant="ghost"
                          aria-expanded={isExpanded}
                          aria-controls={`details-${run.id}`}
                          onClick={() => toggleRunExpansion(run.id)}
                          data-testid={`expand-${run.id}`}
                        >
                          {isExpanded ? <ChevronUp /> : <ChevronDown />}
                        </Button>
                      </td>
                    </tr>
                    
                    {isExpanded && (
                      <tr id={`details-${run.id}`} data-testid={`${run.id}-details`}>
                        <td colSpan={5} className="p-4 bg-muted/50">
                          <div className="space-y-4">
                            {run.error && (
                              <Alert variant="destructive">
                                <AlertCircle className="h-4 w-4" />
                                <AlertDescription>{run.error}</AlertDescription>
                              </Alert>
                            )}
                            
                            {run.steps && (
                              <div data-testid={`${run.id}-steps`}>
                                <h4 className="font-medium mb-2">Step Breakdown</h4>
                                <div className="space-y-2">
                                  {run.steps.map((step, idx) => (
                                    <div 
                                      key={idx} 
                                      className="flex items-center justify-between p-2 border rounded"
                                      data-testid={step.error ? `step-error-${step.name}` : undefined}
                                    >
                                      <span>{step.name}</span>
                                      <div className="flex items-center gap-4">
                                        <span className="text-sm text-muted-foreground">
                                          {formatDuration(step.duration)}
                                        </span>
                                        <Badge 
                                          variant="outline" 
                                          className={getStatusColor(step.status)}
                                        >
                                          {step.status}
                                        </Badge>
                                      </div>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}
                            
                            {run.metrics && (
                              <div data-testid={`${run.id}-metrics`}>
                                <h4 className="font-medium mb-2">Performance Metrics</h4>
                                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                  {run.metrics.recordsProcessed !== undefined && (
                                    <div>
                                      <p className="text-sm text-muted-foreground">Records</p>
                                      <p className="font-medium">{run.metrics.recordsProcessed.toLocaleString()}</p>
                                    </div>
                                  )}
                                  {run.metrics.averageProcessingTime !== undefined && (
                                    <div>
                                      <p className="text-sm text-muted-foreground">Avg Time</p>
                                      <p className="font-medium">{run.metrics.averageProcessingTime}s/record</p>
                                    </div>
                                  )}
                                  {run.metrics.peakMemoryUsage !== undefined && (
                                    <div>
                                      <p className="text-sm text-muted-foreground">Peak Memory</p>
                                      <p className="font-medium">{run.metrics.peakMemoryUsage} MB</p>
                                    </div>
                                  )}
                                  {run.metrics.cpuUtilization !== undefined && (
                                    <div>
                                      <p className="text-sm text-muted-foreground">CPU Usage</p>
                                      <p className="font-medium">{run.metrics.cpuUtilization}%</p>
                                    </div>
                                  )}
                                </div>
                              </div>
                            )}
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                )
              })}
            </tbody>
          </table>
        </CardContent>
      </Card>

      {/* Pagination */}
      <div className="flex items-center justify-between" data-testid="pagination">
        <div className="flex items-center gap-2">
          <Label htmlFor="page-size">Items per page</Label>
          <select
            id="page-size"
            className="px-3 py-2 border rounded-md"
            value={pageSize}
            onChange={(e) => {
              const newSize = parseInt(e.target.value)
              setPageSize(newSize)
              // Try to preserve current page if possible
              const newMaxPage = Math.max(1, Math.ceil(totalItems / newSize))
              setPage(Math.min(page, newMaxPage))
            }}
          >
            <option value="10">10</option>
            <option value="20">20</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() => setPage(prev => Math.max(1, prev - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
          
          <div className="flex items-center gap-2">
            <span>Page {page} of {Math.max(totalPages, 1)}</span>
            <Input
              type="number"
              min="1"
              max={Math.max(totalPages, 1)}
              value={page}
              onChange={(e) => {
                const newPage = Number(e.target.value)
                if (!Number.isNaN(newPage) && newPage >= 1 && newPage <= totalPages) {
                  setPage(newPage)
                }
              }}
              className="w-16"
              data-testid="page-input"
            />
          </div>
          
          <Button
            size="sm"
            variant="outline"
            onClick={() => setPage(prev => Math.min(totalPages, prev + 1))}
            disabled={page === totalPages}
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="text-sm text-muted-foreground" role="status" aria-live="polite">
        Showing {history.length} of {totalItems} total runs
      </div>
    </div>
  )
}