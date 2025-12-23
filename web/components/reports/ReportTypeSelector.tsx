/**
 * Report Type Selector Component
 * Dropdown for filtering reports by type
 * Following CLAUDE.md UI component standards with Shadcn/ui
 */

'use client'

import React from 'react'
import { 
  Calendar, 
  TrendingUp, 
  BarChart3, 
  FileText, 
  Files,
  Droplets,
  Database,
  type LucideIcon 
} from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import type { ReportType, ReportTypeSelectorProps } from '@/types/reports'

// Icon mapping for report types
const REPORT_ICONS: Record<ReportType, LucideIcon> = {
  daily: Calendar,
  ticker: TrendingUp,
  liquidity: Droplets,
  combined: Database,
  indexes: BarChart3,
  summary: FileText,
  all: Files
}

// Labels for report types
const REPORT_LABELS: Record<ReportType, string> = {
  daily: 'Daily Reports',
  ticker: 'Ticker Reports',
  liquidity: 'Liquidity Analysis',
  combined: 'Combined Data',
  indexes: 'Market Indices',
  summary: 'Summary Reports',
  all: 'All Reports'
}

export function ReportTypeSelector({
  selectedType,
  onTypeChange,
  reportCounts
}: ReportTypeSelectorProps) {
  const reportTypes: ReportType[] = ['all', 'daily', 'ticker', 'liquidity', 'summary', 'indexes', 'combined']

  return (
    <div className="w-full">
      <label htmlFor="report-type" className="block text-sm font-medium mb-2">
        Report Type
      </label>
      <Select
        value={selectedType}
        onValueChange={(value) => onTypeChange(value as ReportType)}
      >
        <SelectTrigger id="report-type" className="w-full">
          <SelectValue placeholder="Select report type" />
        </SelectTrigger>
        <SelectContent>
          {reportTypes.map((type) => {
            const Icon = REPORT_ICONS[type]
            const count = reportCounts[type] || 0
            const isDisabled = type !== 'all' && count === 0
            
            return (
              <SelectItem 
                key={type} 
                value={type}
                disabled={isDisabled}
                className="cursor-pointer"
              >
                <div className="flex items-center justify-between w-full">
                  <div className="flex items-center gap-2">
                    <Icon className="h-4 w-4" aria-hidden="true" />
                    <span>{REPORT_LABELS[type]}</span>
                  </div>
                  {count > 0 && (
                    <Badge 
                      variant="secondary" 
                      className="ml-2 min-w-[2rem] text-center"
                    >
                      {count}
                    </Badge>
                  )}
                </div>
              </SelectItem>
            )
          })}
        </SelectContent>
      </Select>
      
      {/* Display selected type description */}
      <p className="mt-2 text-sm text-muted-foreground">
        {selectedType === 'all' && 'Showing all available reports'}
        {selectedType === 'daily' && 'Daily trading reports with market summary'}
        {selectedType === 'ticker' && 'Individual ticker trading history'}
        {selectedType === 'liquidity' && 'ISX Hybrid Liquidity Metrics and safe trading analysis'}
        {selectedType === 'combined' && 'Combined market data for all tickers'}
        {selectedType === 'indexes' && 'Market index performance data (ISX60, ISX15)'}
        {selectedType === 'summary' && 'Consolidated ticker summary reports'}
      </p>
    </div>
  )
}