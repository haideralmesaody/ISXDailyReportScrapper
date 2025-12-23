/**
 * Report Filters Component
 * Advanced filtering panel for reports
 * Following CLAUDE.md UI standards with industry best practices
 */

'use client'

import React, { useState, useCallback, useMemo } from 'react'
import {
  Search,
  Calendar,
  Filter,
  X,
  ChevronDown,
  ChevronUp,
  Building2,
  Hash,
  TrendingUp,
  Clock
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { DatePicker } from '@/components/ui/date-picker'
import { Checkbox } from '@/components/ui/checkbox'
import type { ReportMetadata } from '@/types/reports'

export interface ReportFiltersProps {
  reports: ReportMetadata[]
  onFiltersChange: (filtered: ReportMetadata[]) => void
  className?: string
}

interface FilterState {
  searchQuery: string
  dateFrom: Date | undefined
  dateTo: Date | undefined
  minSize: number | undefined
  maxSize: number | undefined
  sectors: Set<string>
  fileTypes: Set<string>
  sortBy: 'name' | 'date' | 'size' | 'type'
  sortOrder: 'asc' | 'desc'
}

const SECTORS = [
  { value: 'banking', label: 'Banking', icon: Building2 },
  { value: 'telecom', label: 'Telecom', icon: Hash },
  { value: 'industry', label: 'Industry', icon: TrendingUp },
  { value: 'services', label: 'Services', icon: Clock },
]

export function ReportFilters({ 
  reports, 
  onFiltersChange,
  className 
}: ReportFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [filters, setFilters] = useState<FilterState>({
    searchQuery: '',
    dateFrom: undefined,
    dateTo: undefined,
    minSize: undefined,
    maxSize: undefined,
    sectors: new Set(),
    fileTypes: new Set(['csv']),
    sortBy: 'date',
    sortOrder: 'desc'
  })

  // Extract unique sectors from ticker reports
  const availableSectors = useMemo(() => {
    const sectors = new Set<string>()
    reports.forEach(report => {
      if (report.type === 'ticker' && report.path) {
        // Check if path contains sector folder
        const pathParts = report.path.split('/')
        if (pathParts.includes('banking')) sectors.add('banking')
        if (pathParts.includes('telecom')) sectors.add('telecom')
        if (pathParts.includes('industry')) sectors.add('industry')
        if (pathParts.includes('services')) sectors.add('services')
      }
    })
    return Array.from(sectors)
  }, [reports])

  // Apply filters to reports
  const applyFilters = useCallback(() => {
    let filtered = [...reports]

    // Search query filter
    if (filters.searchQuery) {
      const query = filters.searchQuery.toLowerCase()
      filtered = filtered.filter(report => 
        report.name.toLowerCase().includes(query) ||
        report.displayName.toLowerCase().includes(query) ||
        report.type.toLowerCase().includes(query)
      )
    }

    // Date range filter
    if (filters.dateFrom || filters.dateTo) {
      filtered = filtered.filter(report => {
        const reportDate = new Date(report.modified)
        if (filters.dateFrom && reportDate < filters.dateFrom) return false
        if (filters.dateTo && reportDate > filters.dateTo) return false
        return true
      })
    }

    // Size filter
    if (filters.minSize !== undefined || filters.maxSize !== undefined) {
      filtered = filtered.filter(report => {
        if (filters.minSize !== undefined && report.size < filters.minSize) return false
        if (filters.maxSize !== undefined && report.size > filters.maxSize) return false
        return true
      })
    }

    // Sector filter (for ticker reports)
    if (filters.sectors.size > 0) {
      filtered = filtered.filter(report => {
        if (report.type !== 'ticker') return true // Don't filter non-ticker reports
        if (!report.path) return false
        
        // Check if report is in any selected sector
        return Array.from(filters.sectors).some(sector => 
          report.path.includes(sector)
        )
      })
    }

    // Sorting
    filtered.sort((a, b) => {
      let comparison = 0
      
      switch (filters.sortBy) {
        case 'name':
          comparison = a.displayName.localeCompare(b.displayName)
          break
        case 'date':
          comparison = new Date(a.modified).getTime() - new Date(b.modified).getTime()
          break
        case 'size':
          comparison = a.size - b.size
          break
        case 'type':
          comparison = a.type.localeCompare(b.type)
          break
      }
      
      return filters.sortOrder === 'asc' ? comparison : -comparison
    })

    onFiltersChange(filtered)
  }, [reports, filters, onFiltersChange])

  // Apply filters whenever they change
  React.useEffect(() => {
    applyFilters()
  }, [applyFilters])

  // Update filter value
  const updateFilter = useCallback((key: keyof FilterState, value: any) => {
    setFilters(prev => ({
      ...prev,
      [key]: value
    }))
  }, [])

  // Reset all filters
  const resetFilters = useCallback(() => {
    setFilters({
      searchQuery: '',
      dateFrom: undefined,
      dateTo: undefined,
      minSize: undefined,
      maxSize: undefined,
      sectors: new Set(),
      fileTypes: new Set(['csv']),
      sortBy: 'date',
      sortOrder: 'desc'
    })
  }, [])

  // Count active filters
  const activeFilterCount = useMemo(() => {
    let count = 0
    if (filters.searchQuery) count++
    if (filters.dateFrom || filters.dateTo) count++
    if (filters.minSize !== undefined || filters.maxSize !== undefined) count++
    if (filters.sectors.size > 0) count++
    return count
  }, [filters])

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <Filter className="h-4 w-4" />
            Filters
            {activeFilterCount > 0 && (
              <Badge variant="secondary" className="ml-1">
                {activeFilterCount}
              </Badge>
            )}
          </CardTitle>
          <div className="flex items-center gap-2">
            {activeFilterCount > 0 && (
              <Button
                variant="ghost"
                size="sm"
                onClick={resetFilters}
                className="h-7 px-2 text-xs"
              >
                <X className="h-3 w-3 mr-1" />
                Clear
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-7 px-2"
            >
              {isExpanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Search Input - Always visible */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search reports..."
            value={filters.searchQuery}
            onChange={(e) => updateFilter('searchQuery', e.target.value)}
            className="pl-9 h-9"
          />
        </div>

        {/* Sort Options - Always visible */}
        <div className="flex gap-2">
          <Select
            value={filters.sortBy}
            onValueChange={(value) => updateFilter('sortBy', value as FilterState['sortBy'])}
          >
            <SelectTrigger className="h-9 flex-1">
              <SelectValue placeholder="Sort by" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="date">Date Modified</SelectItem>
              <SelectItem value="name">Name</SelectItem>
              <SelectItem value="size">File Size</SelectItem>
              <SelectItem value="type">Report Type</SelectItem>
            </SelectContent>
          </Select>
          
          <Button
            variant="outline"
            size="sm"
            onClick={() => updateFilter('sortOrder', filters.sortOrder === 'asc' ? 'desc' : 'asc')}
            className="h-9 px-3"
          >
            {filters.sortOrder === 'asc' ? '↑' : '↓'}
          </Button>
        </div>

        {/* Advanced Filters - Collapsible */}
        <Collapsible open={isExpanded} onOpenChange={setIsExpanded}>
          <CollapsibleContent className="space-y-4 pt-2">
            <Separator />
            
            {/* Date Range */}
            <div className="space-y-2">
              <Label className="text-sm font-medium">Date Range</Label>
              <div className="grid grid-cols-2 gap-2">
                <DatePicker
                  date={filters.dateFrom}
                  onDateChange={(date) => updateFilter('dateFrom', date)}
                  placeholder="From"
                />
                <DatePicker
                  date={filters.dateTo}
                  onDateChange={(date) => updateFilter('dateTo', date)}
                  placeholder="To"
                />
              </div>
            </div>

            {/* File Size Range */}
            <div className="space-y-2">
              <Label className="text-sm font-medium">File Size (KB)</Label>
              <div className="grid grid-cols-2 gap-2">
                <Input
                  type="number"
                  placeholder="Min"
                  value={filters.minSize || ''}
                  onChange={(e) => updateFilter('minSize', e.target.value ? Number(e.target.value) : undefined)}
                  className="h-9"
                />
                <Input
                  type="number"
                  placeholder="Max"
                  value={filters.maxSize || ''}
                  onChange={(e) => updateFilter('maxSize', e.target.value ? Number(e.target.value) : undefined)}
                  className="h-9"
                />
              </div>
            </div>

            {/* Sector Filter (only if sectors available) */}
            {availableSectors.length > 0 && (
              <div className="space-y-2">
                <Label className="text-sm font-medium">Sectors (Ticker Reports)</Label>
                <div className="space-y-2">
                  {SECTORS.filter(sector => availableSectors.includes(sector.value)).map(sector => {
                    const Icon = sector.icon
                    return (
                      <div key={sector.value} className="flex items-center space-x-2">
                        <Checkbox
                          id={sector.value}
                          checked={filters.sectors.has(sector.value)}
                          onCheckedChange={(checked) => {
                            const newSectors = new Set(filters.sectors)
                            if (checked) {
                              newSectors.add(sector.value)
                            } else {
                              newSectors.delete(sector.value)
                            }
                            updateFilter('sectors', newSectors)
                          }}
                        />
                        <Label
                          htmlFor={sector.value}
                          className="flex items-center gap-2 text-sm font-normal cursor-pointer"
                        >
                          <Icon className="h-3 w-3" />
                          {sector.label}
                        </Label>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Quick Filters */}
            <div className="space-y-2">
              <Label className="text-sm font-medium">Quick Filters</Label>
              <div className="flex flex-wrap gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    const today = new Date()
                    today.setHours(0, 0, 0, 0)
                    updateFilter('dateFrom', today)
                    updateFilter('dateTo', undefined)
                  }}
                  className="h-7 text-xs"
                >
                  Today
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    const lastWeek = new Date()
                    lastWeek.setDate(lastWeek.getDate() - 7)
                    lastWeek.setHours(0, 0, 0, 0)
                    updateFilter('dateFrom', lastWeek)
                    updateFilter('dateTo', undefined)
                  }}
                  className="h-7 text-xs"
                >
                  Last 7 Days
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    const lastMonth = new Date()
                    lastMonth.setMonth(lastMonth.getMonth() - 1)
                    lastMonth.setHours(0, 0, 0, 0)
                    updateFilter('dateFrom', lastMonth)
                    updateFilter('dateTo', undefined)
                  }}
                  className="h-7 text-xs"
                >
                  Last Month
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    updateFilter('minSize', 1000) // > 1MB
                  }}
                  className="h-7 text-xs"
                >
                  Large Files
                </Button>
              </div>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </CardContent>
    </Card>
  )
}