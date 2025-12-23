/**
 * CSV Viewer Component
 * Displays CSV data in a formatted table with pagination and search
 * Following CLAUDE.md performance optimization standards
 */

'use client'

import React, { useState, useMemo, useCallback } from 'react'
import {
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  type ColumnDef,
  type SortingState,
  type ColumnFiltersState,
} from '@tanstack/react-table'
import { 
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useToast } from '@/lib/hooks/use-toast'
import { 
  ChevronLeft, 
  ChevronRight, 
  ChevronsLeft, 
  ChevronsRight,
  Search,
  FileSpreadsheet,
  AlertCircle,
  Loader2,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Download,
  Copy,
  FileJson,
  FileText
} from 'lucide-react'
import type { CSVViewerProps } from '@/types/reports'

export function CSVViewer({
  report,
  csvData,
  isLoading,
  error
}: CSVViewerProps) {
  // Initialize sorting state - for reports with dates, sort by Date descending
  const initialSorting = useMemo<SortingState>(() => {
    if ((report?.type === 'ticker' || report?.type === 'liquidity' || 
         report?.type === 'indexes' || report?.type === 'combined') && 
         csvData?.columns.some(col => col.accessor === 'Date')) {
      return [{ id: 'Date', desc: true }]
    }
    return []
  }, [report?.type, csvData?.columns])
  
  const [sorting, setSorting] = useState<SortingState>(initialSorting)
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
  const [globalFilter, setGlobalFilter] = useState('')
  const [pageSize] = useState(50)
  const { toast } = useToast()
  
  // Update sorting when report type changes
  React.useEffect(() => {
    setSorting(initialSorting)
  }, [initialSorting])

  // Generate table columns dynamically from CSV data
  const columns = useMemo<ColumnDef<Record<string, unknown>>[]>(() => {
    if (!csvData?.columns || csvData.columns.length === 0) return []
    
    return csvData.columns.map((col) => ({
      accessorKey: col.accessor,
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            className="h-8 p-0 font-semibold"
            onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          >
            <span className="mr-1">{col.header}</span>
            {column.getIsSorted() === 'asc' ? (
              <ArrowUp className="h-3 w-3" />
            ) : column.getIsSorted() === 'desc' ? (
              <ArrowDown className="h-3 w-3" />
            ) : (
              <ArrowUpDown className="h-3 w-3 opacity-50" />
            )}
          </Button>
        )
      },
      cell: ({ getValue, column }) => {
        const value = getValue()
        const columnName = column.id.toLowerCase()
        
        // Format based on column name and value type
        if (typeof value === 'number') {
          // Currency columns (price, value)
          if (columnName.includes('price') || columnName.includes('value')) {
            return new Intl.NumberFormat('en-US', {
              style: 'currency',
              currency: 'IQD',
              minimumFractionDigits: 0,
              maximumFractionDigits: 2
            }).format(value).replace('IQD', 'IQD ')
          }
          
          // Volume columns - abbreviate large numbers
          if (columnName.includes('volume') || columnName.includes('shares')) {
            if (value >= 1000000000) {
              return `${(value / 1000000000).toFixed(2)}B`
            } else if (value >= 1000000) {
              return `${(value / 1000000).toFixed(2)}M`
            } else if (value >= 1000) {
              return `${(value / 1000).toFixed(1)}K`
            }
            return new Intl.NumberFormat('en-US').format(value)
          }
          
          // Percentage columns
          if (columnName.includes('change') || columnName.includes('percent') || columnName.includes('return')) {
            const formatted = (value * (Math.abs(value) < 1 ? 100 : 1)).toFixed(2)
            const color = value > 0 ? 'text-green-600' : value < 0 ? 'text-red-600' : ''
            return <span className={color}>{value > 0 ? '+' : ''}{formatted}%</span>
          }
          
          // ILLIQ and other decimal metrics - max 10 characters
          if (columnName.includes('illiq') || columnName.includes('score') || columnName.includes('ratio')) {
            // Format ILLIQ_Raw with max 10 characters
            if (columnName.includes('illiq_raw') || columnName.includes('illiq raw')) {
              const strValue = value.toString()
              if (strValue.length > 10) {
                // Use scientific notation for very large values
                return value.toExponential(4)
              }
              // For normal values, limit decimal places to fit within 10 chars
              const intPart = Math.floor(Math.abs(value)).toString().length
              const maxDecimals = Math.max(0, Math.min(4, 9 - intPart))
              return value.toFixed(maxDecimals)
            }
            // Default for other metrics
            return value.toFixed(4)
          }
          
          // Default number formatting
          return new Intl.NumberFormat('en-US').format(value)
        }
        
        // Format dates
        if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}/.test(value)) {
          return new Date(value).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
          })
        }
        
        // Trading status with color coding
        if (columnName === 'status' || columnName === 'tradingstatus') {
          const status = String(value ?? '').toLowerCase()
          const statusColors = {
            'traded': 'text-green-600',
            'not traded': 'text-gray-500',
            'suspended': 'text-red-600',
            'halted': 'text-orange-600'
          }
          const color = statusColors[status] || ''
          return <span className={color}>{value}</span>
        }
        
        // Ticker symbols in uppercase
        if (columnName === 'symbol' || columnName === 'ticker') {
          return <span className="font-mono font-semibold">{String(value ?? '').toUpperCase()}</span>
        }
        
        return String(value ?? '')
      },
    }))
  }, [csvData])

  // Initialize table
  const table = useReactTable({
    data: csvData?.data || [],
    columns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    initialState: {
      pagination: {
        pageSize,
      },
    },
  })
  
  // Export functions - defined before conditional returns to follow React hooks rules
  const exportToCSV = useCallback(() => {
    if (!csvData || !report) return
    
    const filteredData = table.getFilteredRowModel().rows
    const headers = csvData.columns.map(col => col.header).join(',')
    const rows = filteredData.map(row => 
      csvData.columns.map(col => {
        const value = row.original[col.accessor]
        // Escape values that contain commas or quotes
        if (typeof value === 'string' && (value.includes(',') || value.includes('"'))) {
          return `"${value.replace(/"/g, '""')}"`
        }
        return value ?? ''
      }).join(',')
    )
    
    const csv = [headers, ...rows].join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${report.name}_export.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    toast({
      title: 'Exported',
      description: `Data exported to ${report.name}_export.csv`,
    })
  }, [csvData, table, report, toast])
  
  const exportToJSON = useCallback(() => {
    if (!csvData || !report) return
    
    const filteredData = table.getFilteredRowModel().rows
    const jsonData = filteredData.map(row => row.original)
    const json = JSON.stringify(jsonData, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${report.name}_export.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    
    toast({
      title: 'Exported',
      description: `Data exported to ${report.name}_export.json`,
    })
  }, [csvData, table, report, toast])
  
  const copyToClipboard = useCallback(async () => {
    if (!csvData) return
    
    const filteredData = table.getFilteredRowModel().rows
    const headers = csvData.columns.map(col => col.header).join('\t')
    const rows = filteredData.map(row => 
      csvData.columns.map(col => row.original[col.accessor] ?? '').join('\t')
    )
    
    const text = [headers, ...rows].join('\n')
    
    try {
      await navigator.clipboard.writeText(text)
      toast({
        title: 'Copied',
        description: 'Data copied to clipboard',
      })
    } catch (err) {
      toast({
        title: 'Error',
        description: 'Failed to copy to clipboard',
        variant: 'destructive',
      })
    }
  }, [csvData, table, toast])

  // Loading state
  if (isLoading) {
    return (
      <Card className="h-full flex flex-col">
        <CardHeader className="flex-shrink-0">
          <CardTitle className="flex items-center gap-2">
            <Loader2 className="h-5 w-5 animate-spin" />
            Loading Report...
          </CardTitle>
        </CardHeader>
        <CardContent className="flex-1">
          <div className="space-y-4">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-[400px] w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        </CardContent>
      </Card>
    )
  }

  // Error state
  if (error) {
    return (
      <Card className="h-full flex flex-col">
        <CardHeader className="flex-shrink-0">
          <CardTitle className="flex items-center gap-2 text-destructive">
            <AlertCircle className="h-5 w-5" />
            Error Loading Report
          </CardTitle>
        </CardHeader>
        <CardContent className="flex-1 flex items-center justify-center">
          <p className="text-sm text-muted-foreground">{error.message}</p>
        </CardContent>
      </Card>
    )
  }

  // No report selected
  if (!report || !csvData) {
    return (
      <Card className="h-full flex flex-col">
        <CardContent className="flex-1 flex flex-col items-center justify-center py-12">
          <FileSpreadsheet className="h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="font-semibold text-lg mb-2">No Report Selected</h3>
          <p className="text-sm text-muted-foreground text-center max-w-md">
            Select a report from the list to view its contents here
          </p>
        </CardContent>
      </Card>
    )
  }

  const pageCount = table.getPageCount()
  const currentPage = table.getState().pagination.pageIndex + 1

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-semibold">
            {report.displayName}
          </CardTitle>
          <div className="flex items-center gap-2">
            <Badge variant="outline">{csvData.data.length} rows</Badge>
            
            {/* Export Menu */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm">
                  <Download className="h-4 w-4 mr-1" />
                  Export
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={exportToCSV}>
                  <FileText className="h-4 w-4 mr-2" />
                  Export as CSV
                </DropdownMenuItem>
                <DropdownMenuItem onClick={exportToJSON}>
                  <FileJson className="h-4 w-4 mr-2" />
                  Export as JSON
                </DropdownMenuItem>
                <DropdownMenuItem onClick={copyToClipboard}>
                  <Copy className="h-4 w-4 mr-2" />
                  Copy to Clipboard
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        
        {/* Global Search */}
        <div className="relative mt-4">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search all columns..."
            value={globalFilter ?? ''}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className="pl-10"
          />
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col min-h-0 p-0">
        {/* Table Container */}
        <div className="flex-1 overflow-auto border-b">
          <Table>
            <TableHeader className="sticky top-0 bg-background z-10 border-b">
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead key={header.id} className="whitespace-nowrap font-medium">
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                    className="hover:bg-muted/50"
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell key={cell.id} className="whitespace-nowrap py-2">
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={columns.length} className="h-24 text-center">
                    No results found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>

        {/* Pagination Controls - Fixed Footer */}
        <div className="flex-shrink-0 flex items-center justify-between px-4 py-3 bg-background">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <span>
              Page {currentPage} of {pageCount}
            </span>
            <span>•</span>
            <span>
              {table.getFilteredRowModel().rows.length} filtered rows
            </span>
          </div>
          
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.setPageIndex(0)}
              disabled={!table.getCanPreviousPage()}
              aria-label="Go to first page"
            >
              <ChevronsLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.previousPage()}
              disabled={!table.getCanPreviousPage()}
              aria-label="Go to previous page"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.nextPage()}
              disabled={!table.getCanNextPage()}
              aria-label="Go to next page"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => table.setPageIndex(table.getPageCount() - 1)}
              disabled={!table.getCanNextPage()}
              aria-label="Go to last page"
            >
              <ChevronsRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}