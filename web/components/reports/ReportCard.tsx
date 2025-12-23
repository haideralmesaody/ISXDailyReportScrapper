/**
 * Report Card Component
 * Individual report display with metadata and actions
 * Following CLAUDE.md UI standards with Shadcn/ui
 */

'use client'

import React, { useState, useEffect } from 'react'
import { 
  Calendar, 
  TrendingUp, 
  BarChart3, 
  FileText, 
  Download,
  Eye,
  FileSpreadsheet,
  Droplets,
  Database,
  Folder,
  type LucideIcon 
} from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { formatFileSize, formatDate, extractTickerSymbol, extractDailyReportDate } from '@/lib/utils/csv-parser'
import { downloadReportContent } from '@/lib/api/reports'
import { parseCSVContent } from '@/lib/utils/csv-parser'
import type { ReportCardProps, ReportType } from '@/types/reports'

// Icon mapping for report types
const REPORT_ICONS: Record<ReportType, LucideIcon> = {
  daily: Calendar,
  ticker: TrendingUp,
  liquidity: Droplets,
  combined: Database,
  indexes: BarChart3,
  summary: FileText,
  all: FileSpreadsheet
}

// Badge variant mapping for report types
const REPORT_BADGE_VARIANTS: Record<ReportType, 'default' | 'secondary' | 'outline'> = {
  daily: 'default',
  ticker: 'secondary',
  liquidity: 'outline',
  combined: 'secondary',
  indexes: 'outline',
  summary: 'default',
  all: 'secondary'
}

export function ReportCard({
  report,
  isSelected,
  onSelect,
  onDownload
}: ReportCardProps) {
  const [previewData, setPreviewData] = useState<string | null>(null)
  const [isLoadingPreview, setIsLoadingPreview] = useState(false)
  const [previewError, setPreviewError] = useState(false)
  
  const Icon = REPORT_ICONS[report.type]
  const badgeVariant = REPORT_BADGE_VARIANTS[report.type]
  
  // Extract additional info based on report type
  const tickerSymbol = report.type === 'ticker' ? extractTickerSymbol(report.name) : null
  const dailyDate = report.type === 'daily' ? extractDailyReportDate(report.name) : null
  
  // Format display name based on type
  const getDisplayName = () => {
    if (tickerSymbol) {
      return `${tickerSymbol} Trading History`
    }
    if (dailyDate) {
      return `Daily Report - ${dailyDate.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'short', 
        day: 'numeric' 
      })}`
    }
    if (report.type === 'liquidity') {
      return report.name.includes('summary') ? 'Liquidity Summary' : 'Liquidity Analysis Report'
    }
    if (report.type === 'combined') {
      return 'Combined Market Data'
    }
    if (report.type === 'indexes') {
      return 'Market Indices Report'
    }
    if (report.type === 'summary') {
      return 'Ticker Summary Report'
    }
    return report.displayName
  }
  
  // Format folder path for display
  const formatFolderPath = (path: string) => {
    if (!path) return null
    const parts = path.split('/')
    if (parts.length <= 1) return null
    // Remove the filename and join the rest
    const folderParts = parts.slice(0, -1)
    return folderParts.join(' › ')
  }
  
  // Load preview data on hover
  const loadPreview = async () => {
    if (previewData || isLoadingPreview || previewError) return
    
    setIsLoadingPreview(true)
    try {
      const content = await downloadReportContent(report.path || report.name)
      const parsed = await parseCSVContent(content)
      
      // Format preview based on report type
      let preview = ''
      if (parsed.data.length > 0) {
        const sampleRows = parsed.data.slice(0, 3)
        const keyColumns = getKeyColumnsForType(report.type)
        
        preview = sampleRows.map((row, index) => {
          const items = keyColumns.map(col => {
            const column = parsed.columns.find(c => 
              c.accessor.toLowerCase().includes(col.toLowerCase())
            )
            if (column) {
              const value = row[column.accessor]
              return `${column.header}: ${formatPreviewValue(value, column.header)}`
            }
            return null
          }).filter(Boolean)
          
          return `Row ${index + 1}: ${items.join(' | ')}`
        }).join('\n')
        
        preview += `\n\nTotal: ${parsed.data.length} rows`
      }
      
      setPreviewData(preview || 'No data available')
    } catch (err) {
      console.error('Failed to load preview:', err)
      setPreviewError(true)
      setPreviewData('Failed to load preview')
    } finally {
      setIsLoadingPreview(false)
    }
  }
  
  // Get key columns based on report type
  const getKeyColumnsForType = (type: ReportType): string[] => {
    switch (type) {
      case 'daily':
        return ['symbol', 'close', 'volume', 'change']
      case 'ticker':
        return ['date', 'close', 'volume', 'trades']
      case 'liquidity':
        return ['ticker', 'illiq', 'safevalue', 'volume']
      case 'combined':
        return ['ticker', 'price', 'volume', 'trades']
      case 'indexes':
        return ['index', 'value', 'change', 'volume']
      case 'summary':
        return ['ticker', 'avgprice', 'totalvolume', 'trades']
      default:
        return ['symbol', 'value', 'volume']
    }
  }
  
  // Format value for preview display
  const formatPreviewValue = (value: any, columnName: string): string => {
    if (value === null || value === undefined) return 'N/A'
    
    const colLower = columnName.toLowerCase()
    
    if (typeof value === 'number') {
      if (colLower.includes('price') || colLower.includes('close')) {
        return new Intl.NumberFormat('en-US', { 
          style: 'decimal', 
          minimumFractionDigits: 0,
          maximumFractionDigits: 2 
        }).format(value)
      }
      if (colLower.includes('volume') || colLower.includes('shares')) {
        if (value >= 1000000) {
          return `${(value / 1000000).toFixed(1)}M`
        } else if (value >= 1000) {
          return `${(value / 1000).toFixed(0)}K`
        }
        return value.toString()
      }
      if (colLower.includes('change') || colLower.includes('percent')) {
        return `${value > 0 ? '+' : ''}${(value * (Math.abs(value) < 1 ? 100 : 1)).toFixed(2)}%`
      }
      if (colLower.includes('illiq')) {
        return value.toFixed(4)
      }
      return value.toLocaleString('en-US')
    }
    
    if (typeof value === 'string' && colLower.includes('date')) {
      try {
        return new Date(value).toLocaleDateString('en-US', { 
          month: 'short', 
          day: 'numeric' 
        })
      } catch {
        return value
      }
    }
    
    return String(value)
  }

  return (
    <TooltipProvider delayDuration={700}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Card 
            className={cn(
              "transition-all duration-200 hover:shadow-md cursor-pointer",
              isSelected && "ring-2 ring-primary shadow-md"
            )}
            onClick={onSelect}
            onMouseEnter={loadPreview}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                onSelect()
              }
            }}
            aria-selected={isSelected}
            aria-label={`Select ${getDisplayName()}`}
          >
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          {/* Icon and Info */}
          <div className="flex items-start gap-3 flex-1 min-w-0">
            <div className={cn(
              "p-2 rounded-lg",
              report.type === 'daily' && "bg-blue-50 dark:bg-blue-950",
              report.type === 'ticker' && "bg-green-50 dark:bg-green-950",
              report.type === 'liquidity' && "bg-cyan-50 dark:bg-cyan-950",
              report.type === 'combined' && "bg-indigo-50 dark:bg-indigo-950",
              report.type === 'indexes' && "bg-purple-50 dark:bg-purple-950",
              report.type === 'summary' && "bg-orange-50 dark:bg-orange-950",
              report.type === 'all' && "bg-gray-50 dark:bg-gray-950"
            )}>
              <Icon className={cn(
                "h-5 w-5",
                report.type === 'daily' && "text-blue-600 dark:text-blue-400",
                report.type === 'ticker' && "text-green-600 dark:text-green-400",
                report.type === 'liquidity' && "text-cyan-600 dark:text-cyan-400",
                report.type === 'combined' && "text-indigo-600 dark:text-indigo-400",
                report.type === 'indexes' && "text-purple-600 dark:text-purple-400",
                report.type === 'summary' && "text-orange-600 dark:text-orange-400",
                report.type === 'all' && "text-gray-600 dark:text-gray-400"
              )} />
            </div>
            
            <div className="flex-1 min-w-0">
              <h3 className="font-medium text-sm truncate" title={getDisplayName()}>
                {getDisplayName()}
              </h3>
              
              {/* Folder path if available */}
              {formatFolderPath(report.path) && (
                <div className="flex items-center gap-1 mt-1 text-xs text-muted-foreground">
                  <Folder className="h-3 w-3" />
                  <span className="truncate" title={formatFolderPath(report.path) || ''}>
                    {formatFolderPath(report.path)}
                  </span>
                </div>
              )}
              
              <p className="text-xs text-muted-foreground mt-1 truncate" title={report.name}>
                {report.name}
              </p>
              
              {/* Metadata */}
              <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
                <span>{formatFileSize(report.size)}</span>
                <span>•</span>
                <span>{formatDate(report.modified)}</span>
              </div>
              
              {/* Type Badge */}
              <Badge variant={badgeVariant} className="mt-2 text-xs">
                {report.type.charAt(0).toUpperCase() + report.type.slice(1)}
              </Badge>
            </div>
          </div>
          
          {/* Actions */}
          <div className="flex flex-col gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={(e) => {
                e.stopPropagation()
                onSelect()
              }}
              title="View report"
              aria-label={`View ${getDisplayName()}`}
            >
              <Eye className="h-4 w-4" />
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={(e) => {
                e.stopPropagation()
                onDownload()
              }}
              title="Download report"
              aria-label={`Download ${getDisplayName()}`}
            >
              <Download className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardContent>
          </Card>
        </TooltipTrigger>
        <TooltipContent side="right" className="max-w-sm p-3">
          <div className="space-y-2">
            <div className="font-semibold text-sm">{getDisplayName()}</div>
            <div className="text-xs text-muted-foreground">
              {isLoadingPreview ? (
                <span>Loading preview...</span>
              ) : previewData ? (
                <pre className="whitespace-pre-wrap font-mono">{previewData}</pre>
              ) : (
                <span>Hover to load preview</span>
              )}
            </div>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}