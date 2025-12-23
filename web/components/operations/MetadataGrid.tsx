/**
 * MetadataGrid Component - Generic metadata display grid
 * 
 * Provides a consistent way to display key-value metadata
 * across all operation components.
 */

'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { useHydration } from '@/lib/hooks/use-hydration'

interface MetadataGridProps {
  metadata: Record<string, any>
  columns?: 1 | 2 | 3 | 4
  maxItems?: number
  priorityKeys?: string[]
  hiddenKeys?: string[]
  className?: string
  formatters?: Record<string, (value: any) => string>
  labels?: Record<string, string>
}

// Default labels for common metadata keys
const defaultLabels: Record<string, string> = {
  // File operations
  total_expected: 'Total Files',
  files_downloaded: 'Downloaded',
  files_existing: 'Already Exists',
  files_remaining: 'Remaining',
  current_file: 'Current File',
  current_page: 'Page',
  
  // Processing
  records_processed: 'Records',
  processing_rate: 'Rate',
  error_count: 'Errors',
  warning_count: 'Warnings',
  success_count: 'Successful',
  skip_count: 'Skipped',
  
  // Time
  started_at: 'Started',
  completed_at: 'Completed',
  duration: 'Duration',
  estimated_completion: 'ETA',
  
  // General
  message: 'Status',
  progress: 'Progress',
  status: 'Status',
  type: 'Type',
  version: 'Version'
}

// Default formatters for common value types (now accepts isHydrated parameter)
const createDefaultFormatters = (isHydrated: boolean): Record<string, (value: any) => string> => ({
  processing_rate: (v) => typeof v === 'number' ? `${v.toFixed(1)}/s` : String(v),
  duration: (v) => typeof v === 'number' ? `${(v / 1000).toFixed(1)}s` : String(v),
  progress: (v) => typeof v === 'number' ? `${v}%` : String(v),
  // Format any key ending with _at as time (hydration-safe)
  '*_at': (v) => {
    if (typeof v === 'string') {
      if (!isHydrated) return v // Return raw string during SSR
      try {
        return new Date(v).toLocaleTimeString()
      } catch {
        return v
      }
    }
    return String(v)
  },
  // Format any key ending with _percent as percentage
  '*_percent': (v) => typeof v === 'number' ? `${v}%` : String(v),
  // Format any key ending with _bytes as size
  '*_bytes': (v) => {
    if (typeof v !== 'number') return String(v)
    const units = ['B', 'KB', 'MB', 'GB']
    let size = v
    let unit = 0
    while (size >= 1024 && unit < units.length - 1) {
      size /= 1024
      unit++
    }
    return `${size.toFixed(1)} ${units[unit]}`
  }
})

export function MetadataGrid({
  metadata,
  columns = 2,
  maxItems,
  priorityKeys = [],
  hiddenKeys = [],
  className,
  formatters = {},
  labels = {}
}: MetadataGridProps) {
  const isHydrated = useHydration()
  
  if (!metadata || Object.keys(metadata).length === 0) {
    return null
  }
  
  // Process and sort metadata entries
  const entries = React.useMemo(() => {
    const allEntries = Object.entries(metadata)
      .filter(([key]) => {
        // Filter out hidden keys and internal keys (starting with _)
        return !hiddenKeys.includes(key) && !key.startsWith('_')
      })
      .sort(([a], [b]) => {
        // Sort by priority first
        const aIndex = priorityKeys.indexOf(a)
        const bIndex = priorityKeys.indexOf(b)
        if (aIndex !== -1 && bIndex !== -1) return aIndex - bIndex
        if (aIndex !== -1) return -1
        if (bIndex !== -1) return 1
        // Then alphabetically
        return a.localeCompare(b)
      })
    
    // Limit items if specified
    return maxItems ? allEntries.slice(0, maxItems) : allEntries
  }, [metadata, priorityKeys, hiddenKeys, maxItems])
  
  // Get label for a key
  const getLabel = (key: string): string => {
    if (labels[key]) return labels[key]
    if (defaultLabels[key]) return defaultLabels[key]
    // Convert snake_case to Title Case
    return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
  }
  
  // Create default formatters with hydration state
  const defaultFormatters = React.useMemo(() => createDefaultFormatters(isHydrated), [isHydrated])
  
  // Format value for display
  const formatValue = (key: string, value: any): string => {
    if (value === null || value === undefined) return '—'
    
    // Check custom formatters first
    if (formatters[key]) {
      return formatters[key](value)
    }
    
    // Check default formatters by exact key
    if (defaultFormatters[key]) {
      return defaultFormatters[key](value)
    }
    
    // Check pattern-based formatters
    for (const [pattern, formatter] of Object.entries(defaultFormatters)) {
      if (pattern.startsWith('*')) {
        const suffix = pattern.slice(1)
        if (key.endsWith(suffix)) {
          return formatter(value)
        }
      }
    }
    
    // Boolean values
    if (typeof value === 'boolean') {
      return value ? 'Yes' : 'No'
    }
    
    // Default to string
    return String(value)
  }
  
  const gridClassName = cn(
    'grid gap-x-4 gap-y-2 text-sm',
    columns === 1 && 'grid-cols-1',
    columns === 2 && 'grid-cols-2',
    columns === 3 && 'grid-cols-3',
    columns === 4 && 'grid-cols-4',
    className
  )
  
  return (
    <div className={gridClassName}>
      {entries.map(([key, value]) => (
        <div key={key} className="flex justify-between gap-2">
          <span className="text-muted-foreground truncate">
            {getLabel(key)}:
          </span>
          <span className="font-medium text-right">
            {formatValue(key, value)}
          </span>
        </div>
      ))}
    </div>
  )
}

export default MetadataGrid