/**
 * CSV parsing utilities for Reports feature
 * Following CLAUDE.md performance and error handling standards
 */

import Papa from 'papaparse'
import type { ParsedCSVData, ReportMetadata, ReportType } from '@/types/reports'

/**
 * Parse CSV content into structured data
 * Uses PapaParse for robust CSV parsing with error handling
 */
export function parseCSVContent(content: string): Promise<ParsedCSVData> {
  // Handle empty content gracefully
  if (!content || content.trim() === '') {
    return Promise.resolve({
      columns: [],
      data: []
    })
  }
  
  return new Promise((resolve, reject) => {
    Papa.parse(content, {
      header: true,
      skipEmptyLines: 'greedy', // Skip empty lines more aggressively
      dynamicTyping: true,
      delimiter: '', // Auto-detect delimiter
      transformHeader: (header) => header.trim(), // Trim whitespace from headers
      transform: (value) => {
        // Trim whitespace from values
        if (typeof value === 'string') {
          return value.trim()
        }
        return value
      },
      complete: (results) => {
        // Filter out rows with critical errors only
        const criticalErrors = results.errors.filter(error => 
          error.code === 'MissingQuotes' || 
          error.code === 'UndetectableDelimiter'
        )
        
        if (criticalErrors.length > 0) {
          console.warn('CSV parsing warnings:', criticalErrors.length, 'issues found')
        }

        // Filter out empty or malformed rows
        const data = (results.data as Record<string, unknown>[]).filter(row => {
          // Check if row has any non-null values
          const values = Object.values(row)
          return values.some(val => val !== null && val !== undefined && val !== '')
        })
        
        const headers = results.meta.fields || []

        const columns = headers.map((header) => ({
          key: header,
          header: formatColumnHeader(header),
          accessor: header
        }))

        resolve({
          columns,
          data: data || []  // Ensure data is never null/undefined
        })
      },
      error: (error: Error) => {
        reject(new Error(`Failed to parse CSV: ${error.message}`))
      }
    })
  })
}

/**
 * Parse CSV file object
 * Handles file reading and parsing with proper error handling
 */
export function parseCSVFile(file: File): Promise<ParsedCSVData> {
  return new Promise((resolve, reject) => {
    Papa.parse(file, {
      header: true,
      skipEmptyLines: 'greedy', // Skip empty lines more aggressively
      dynamicTyping: true,
      delimiter: '', // Auto-detect delimiter
      worker: true, // Use web worker for large files
      transformHeader: (header) => header.trim(), // Trim whitespace from headers
      transform: (value) => {
        // Trim whitespace from values
        if (typeof value === 'string') {
          return value.trim()
        }
        return value
      },
      complete: (results) => {
        // Filter out rows with critical errors only
        const criticalErrors = results.errors.filter(error => 
          error.code === 'MissingQuotes' || 
          error.code === 'UndetectableDelimiter'
        )
        
        if (criticalErrors.length > 0) {
          console.warn('CSV parsing warnings:', criticalErrors.length, 'issues found')
        }

        // Filter out empty or malformed rows
        const data = (results.data as Record<string, unknown>[]).filter(row => {
          // Check if row has any non-null values
          const values = Object.values(row)
          return values.some(val => val !== null && val !== undefined && val !== '')
        })
        
        const headers = results.meta.fields || []

        const columns = headers.map((header) => ({
          key: header,
          header: formatColumnHeader(header),
          accessor: header
        }))

        resolve({
          columns,
          data: data || []  // Ensure data is never null/undefined
        })
      },
      error: (error) => {
        reject(new Error(`Failed to parse CSV file: ${error.message}`))
      }
    })
  })
}

/**
 * Format column headers for display
 * Converts snake_case or camelCase to Title Case
 */
export function formatColumnHeader(header: string): string {
  return header
    .replace(/_/g, ' ')
    .replace(/([A-Z])/g, ' $1')
    .trim()
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

/**
 * Determine report type from filename
 * Uses patterns defined in CLAUDE.md standards
 */
export function getReportType(filename: string): ReportType {
  // Daily reports: isx_daily_YYYY_MM_DD.csv
  if (/^isx_daily_\d{4}_\d{2}_\d{2}\.csv$/.test(filename)) {
    return 'daily'
  }
  
  // Ticker reports: SYMBOL_trading_history.csv
  if (/_trading_history\.csv$/.test(filename)) {
    return 'ticker'
  }
  
  // Liquidity reports: liquidity_report*, liquidity_summary*, liquidity_scores*, liquidity_insights*
  if (/^liquidity_(report|summary|scores|insights)/i.test(filename)) {
    return 'liquidity'
  }
  
  // Combined data: isx_combined_data.csv
  if (/^isx_combined/i.test(filename)) {
    return 'combined'
  }
  
  // Market indices: indexes.csv
  if (filename === 'indexes.csv' || filename.toLowerCase().includes('index')) {
    return 'indexes'
  }
  
  // Summary reports: ticker_summary.csv or ticker_summary.json
  if (/^ticker_summary\.(csv|json)$/i.test(filename)) {
    return 'summary'
  }
  
  // Default to 'all' for unknown types
  return 'all'
}

/**
 * Format file size for display
 * Converts bytes to human-readable format
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

/**
 * Format date for display
 * Converts ISO string to localized format
 */
export function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

/**
 * Extract ticker symbol from filename
 * Returns the ticker symbol from trading history files
 */
export function extractTickerSymbol(filename: string): string | null {
  const match = filename.match(/^(.+)_trading_history\.csv$/)
  return match ? match[1] || null : null
}

/**
 * Extract date from daily report filename
 * Returns date object from isx_daily files
 */
export function extractDailyReportDate(filename: string): Date | null {
  const match = filename.match(/^isx_daily_(\d{4})_(\d{2})_(\d{2})\.csv$/)
  if (!match) return null
  
  const [, year, month, day] = match
  return new Date(parseInt(year!), parseInt(month!) - 1, parseInt(day!))
}

/**
 * Sort reports by type with custom sorting per type
 * - Daily reports: sorted by date (newest first) based on filename
 * - Ticker reports: sorted alphabetically by ticker symbol
 * - Other reports: sorted by modified date (newest first)
 */
export function sortReports(reports: ReportMetadata[]): ReportMetadata[] {
  return reports.sort((a, b) => {
    // Sort by type first
    if (a.type !== b.type) {
      const typeOrder: ReportType[] = ['daily', 'ticker', 'liquidity', 'combined', 'index', 'indexes', 'summary', 'all']
      return typeOrder.indexOf(a.type) - typeOrder.indexOf(b.type)
    }
    
    // Type-specific sorting
    if (a.type === 'daily' && b.type === 'daily') {
      // Extract dates from filename pattern: isx_daily_YYYY_MM_DD.csv
      const dateA = extractDailyReportDate(a.name)
      const dateB = extractDailyReportDate(b.name)
      
      if (dateA && dateB) {
        // Sort by date, newest first
        return dateB.getTime() - dateA.getTime()
      }
    }
    
    if (a.type === 'ticker' && b.type === 'ticker') {
      // Extract ticker symbol from filename pattern: SYMBOL_trading_history.csv
      const symbolA = extractTickerSymbol(a.name)
      const symbolB = extractTickerSymbol(b.name)
      
      if (symbolA && symbolB) {
        // Sort alphabetically by ticker symbol
        return symbolA.toUpperCase().localeCompare(symbolB.toUpperCase())
      }
    }
    
    // Default: sort by modified date (newest first)
    return new Date(b.modified).getTime() - new Date(a.modified).getTime()
  })
}


/**
 * Filter reports by search term
 * Case-insensitive search in filename and display name
 */
export function filterReports(
  reports: ReportMetadata[],
  searchTerm: string
): ReportMetadata[] {
  if (!searchTerm.trim()) return reports
  
  const term = searchTerm.toLowerCase()
  return reports.filter(report => 
    report.name.toLowerCase().includes(term) ||
    report.displayName.toLowerCase().includes(term)
  )
}

/**
 * Group reports by type
 * Returns a map of report types to their reports
 */
export function groupReportsByType(
  reports: ReportMetadata[]
): Map<ReportType, ReportMetadata[]> {
  const grouped = new Map<ReportType, ReportMetadata[]>()
  
  reports.forEach(report => {
    const type = report.type
    if (!grouped.has(type)) {
      grouped.set(type, [])
    }
    grouped.get(type)!.push(report)
  })
  
  return grouped
}

/**
 * Validate CSV structure
 * Ensures CSV has required columns for specific report types
 */
export function validateCSVStructure(
  columns: string[],
  reportType: ReportType
): { isValid: boolean; missingColumns?: string[] } {
  const requiredColumns: Record<ReportType, string[]> = {
    daily: ['Symbol', 'Open', 'High', 'Low', 'Close', 'Volume'],
    ticker: ['Date', 'Open', 'High', 'Low', 'Close', 'Volume'],
    index: ['Index', 'Value', 'Change', 'Change%'],
    indexes: ['Index', 'Value', 'Change', 'Change%'],
    summary: ['Symbol', 'LastPrice', 'Change', 'Volume'],
    liquidity: ['Symbol', 'ILLIQ_Raw', 'ILLIQ_Scaled', 'Hybrid_Score'],
    combined: [],
    all: []
  }
  
  const required = requiredColumns[reportType] || []
  const missing = required.filter(col => !columns.includes(col))
  
  return {
    isValid: missing.length === 0,
    ...(missing.length > 0 && { missingColumns: missing })
  }
}