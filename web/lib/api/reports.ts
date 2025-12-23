/**
 * Reports API Client
 * Handles all report-related API calls
 * Following CLAUDE.md error handling and TypeScript standards
 */

import { API_BASE_URL } from '@/lib/constants/api'
import type { ReportMetadata, ReportType } from '@/types/reports'
import { reportApiResponseSchema } from '@/lib/schemas/reports'
import { getReportType } from '@/lib/utils/csv-parser'

/**
 * Fetch all reports from the backend
 * Maps backend response to frontend types
 */
export async function fetchReports(): Promise<ReportMetadata[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/data/reports`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.detail || `Failed to fetch reports: ${response.statusText}`)
    }

    const data = await response.json()
    
    // Handle null data from API (when no reports exist)
    if (data && data.data === null) {
      data.data = []
    }
    
    // Validate response with Zod schema
    const validated = reportApiResponseSchema.parse(data)
    
    // Map to ReportMetadata with type detection
    return validated.data
      .filter((report) => {
        // Filter out liquidity_insights from reports page
        return report.category !== 'liquidity_insights'
      })
      .map((report) => {
        // Use category from backend if available, otherwise detect from filename
        const type = report.category ? mapCategoryToType(report.category) : getReportType(report.name)
        return {
          name: report.name,
          size: report.size,
          modified: report.modified,
          type,
          displayName: formatDisplayName(report.name, type),
          path: report.path || report.name, // Use nested path if available, fallback to name
          category: report.category,
        }
      })
  } catch (error) {
    console.error('Error fetching reports:', error)
    throw error instanceof Error ? error : new Error('Failed to fetch reports')
  }
}

/**
 * Download report file content
 * Returns the raw CSV content as text
 * @param filepath - Can be a simple filename or nested path like "daily/2024/01/report.csv"
 */
export async function downloadReportContent(filepath: string): Promise<string> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(filepath)}`,
      {
        method: 'GET',
        credentials: 'include',
      }
    )

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.detail || `Failed to download report: ${response.statusText}`)
    }

    return await response.text()
  } catch (error) {
    console.error('Error downloading report:', error)
    throw error instanceof Error ? error : new Error('Failed to download report')
  }
}

/**
 * Download report file as blob for saving
 * Returns blob for browser download
 * @param filepath - Can be a simple filename or nested path like "daily/2024/01/report.csv"
 */
export async function downloadReportFile(filepath: string): Promise<void> {
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(filepath)}`,
      {
        method: 'GET',
        credentials: 'include',
      }
    )

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.detail || `Failed to download file: ${response.statusText}`)
    }

    // Get the blob from response
    const blob = await response.blob()
    
    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    // Extract just the filename from the path for download name
    const filename = filepath.split('/').pop() || filepath
    link.download = filename
    document.body.appendChild(link)
    link.click()
    
    // Cleanup
    setTimeout(() => {
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    }, 100)
  } catch (error) {
    console.error('Error downloading file:', error)
    throw error instanceof Error ? error : new Error('Failed to download file')
  }
}

/**
 * Format display name based on report type
 * Creates user-friendly names from filenames
 */
function formatDisplayName(filename: string, type: ReportType): string {
  switch (type) {
    case 'daily': {
      const match = filename.match(/^isx_daily_(\d{4})_(\d{2})_(\d{2})\.csv$/)
      if (match) {
        const [, year, month, day] = match
        const date = new Date(parseInt(year!), parseInt(month!) - 1, parseInt(day!))
        return `Daily Report - ${date.toLocaleDateString('en-US', { 
          year: 'numeric', 
          month: 'short', 
          day: 'numeric' 
        })}`
      }
      break
    }
    case 'ticker': {
      const match = filename.match(/^(.+)_trading_history\.csv$/)
      if (match) {
        return `${match[1]} Trading History`
      }
      break
    }
    case 'liquidity':
      if (filename.includes('summary')) {
        return 'Liquidity Summary Analysis'
      }
      if (filename.includes('scores')) {
        return 'Liquidity Scores Report'
      }
      return 'Liquidity Metrics Report'
    case 'combined':
      return 'Combined Market Data'
    case 'indexes':
      return 'Market Indices Report'
    case 'summary':
      if (filename.endsWith('.json')) {
        return 'Ticker Summary (JSON)'
      }
      return 'Ticker Summary Report'
  }
  
  // Default: clean up filename
  return filename
    .replace(/_/g, ' ')
    .replace(/\.(csv|json|txt)$/i, '')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

/**
 * Prefetch report content for faster loading
 * Uses browser cache for performance
 * @param filepath - Can be a simple filename or nested path
 */
export async function prefetchReportContent(filepath: string): Promise<void> {
  try {
    await fetch(
      `${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(filepath)}`,
      {
        method: 'HEAD',
        credentials: 'include',
      }
    )
  } catch (error) {
    // Silent fail for prefetch
    console.debug('Prefetch failed:', filepath, error)
  }
}

/**
 * Map backend category to frontend ReportType
 * Now properly handles all report categories
 */
function mapCategoryToType(category: string): ReportType {
  switch (category) {
    case 'daily':
      return 'daily'
    case 'ticker':
      return 'ticker'
    case 'liquidity':
      return 'liquidity'
    case 'liquidity_insights':
      // Don't show insights in reports page, return 'all' to hide it
      return 'all'
    case 'combined':
      return 'combined'
    case 'indexes':
      return 'indexes'
    case 'summary':
      return 'summary'
    default:
      return 'all'
  }
}