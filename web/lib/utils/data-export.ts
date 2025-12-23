/**
 * Data Export Utilities
 * Provides functionality to export chart data as CSV and JSON
 */

import type { TickerHistoricalData } from '@/types/analysis'

export interface ExportDataOptions {
  fileName?: string
  includeIndicators?: boolean
  dateFormat?: 'YYYY-MM-DD' | 'DD/MM/YYYY' | 'MM/DD/YYYY'
}

/**
 * Convert data to CSV format
 */
export function dataToCSV(
  data: TickerHistoricalData[],
  indicators?: {
    sma20?: number[]
    sma50?: number[]
    sma200?: number[]
    ema20?: number[]
  },
  options: ExportDataOptions = {}
): string {
  const { includeIndicators = true, dateFormat = 'YYYY-MM-DD' } = options

  if (data.length === 0) {
    return ''
  }

  // Build headers
  const headers = [
    'Date',
    'Open',
    'High',
    'Low',
    'Close',
    'Volume',
    'Value (IQD)',
    'Trades',
    'Change',
    'Change %'
  ]

  if (includeIndicators && indicators) {
    if (indicators.sma20) headers.push('SMA 20')
    if (indicators.sma50) headers.push('SMA 50')
    if (indicators.sma200) headers.push('SMA 200')
    if (indicators.ema20) headers.push('EMA 20')
  }

  // Build rows
  const rows = data.map((item, index) => {
    const row = [
      formatDate(item.date, dateFormat),
      item.open.toFixed(2),
      item.high.toFixed(2),
      item.low.toFixed(2),
      item.close.toFixed(2),
      item.volume.toString(),
      item.value?.toFixed(2) || '',
      item.trades?.toString() || '',
      item.change?.toFixed(2) || '',
      item.changePercent?.toFixed(2) || ''
    ]

    if (includeIndicators && indicators) {
      if (indicators.sma20) {
        row.push(indicators.sma20[index]?.toFixed(2) || '')
      }
      if (indicators.sma50) {
        row.push(indicators.sma50[index]?.toFixed(2) || '')
      }
      if (indicators.sma200) {
        row.push(indicators.sma200[index]?.toFixed(2) || '')
      }
      if (indicators.ema20) {
        row.push(indicators.ema20[index]?.toFixed(2) || '')
      }
    }

    return row
  })

  // Combine headers and rows
  const csvContent = [
    headers.join(','),
    ...rows.map(row => row.join(','))
  ].join('\n')

  return csvContent
}

/**
 * Export data as CSV file
 */
export function exportDataAsCSV(
  data: TickerHistoricalData[],
  ticker: string,
  indicators?: {
    sma20?: number[]
    sma50?: number[]
    sma200?: number[]
    ema20?: number[]
  },
  options: ExportDataOptions = {}
): void {
  const {
    fileName = `${ticker}-data-${new Date().toISOString().split('T')[0]}`
  } = options

  const csvContent = dataToCSV(data, indicators, options)
  
  if (!csvContent) {
    throw new Error('No data to export')
  }

  // Create blob with BOM for Excel compatibility
  const BOM = '\uFEFF'
  const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8;' })
  
  // Create download link
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${fileName}.csv`
  
  // Trigger download
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // Clean up
  URL.revokeObjectURL(url)
}

/**
 * Export data as JSON file
 */
export function exportDataAsJSON(
  data: TickerHistoricalData[],
  ticker: string,
  metadata?: {
    exportDate?: string
    timeRange?: string
    indicators?: string[]
  },
  options: ExportDataOptions = {}
): void {
  const {
    fileName = `${ticker}-data-${new Date().toISOString().split('T')[0]}`
  } = options

  const exportData = {
    ticker,
    exportDate: new Date().toISOString(),
    metadata: metadata || {},
    dataPoints: data.length,
    data
  }

  const jsonContent = JSON.stringify(exportData, null, 2)
  const blob = new Blob([jsonContent], { type: 'application/json' })
  
  // Create download link
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${fileName}.json`
  
  // Trigger download
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // Clean up
  URL.revokeObjectURL(url)
}

/**
 * Format date based on format option
 */
function formatDate(date: string, format: string): string {
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')

  switch (format) {
    case 'DD/MM/YYYY':
      return `${day}/${month}/${year}`
    case 'MM/DD/YYYY':
      return `${month}/${day}/${year}`
    case 'YYYY-MM-DD':
    default:
      return `${year}-${month}-${day}`
  }
}

/**
 * Generate analysis summary for export
 */
export function generateAnalysisSummary(
  data: TickerHistoricalData[],
  ticker: string,
  companyName?: string
): {
  ticker: string
  companyName?: string
  period: string
  dataPoints: number
  priceRange: {
    min: number
    max: number
    current: number
    change: number
    changePercent: number
  }
  volumeStats: {
    total: number
    average: number
    min: number
    max: number
  }
  performance: {
    daily: number
    weekly?: number
    monthly?: number
    yearly?: number
  }
} {
  if (data.length === 0) {
    throw new Error('No data available for analysis')
  }

  const firstPoint = data[0]
  const lastPoint = data[data.length - 1]
  
  // Calculate price statistics
  const prices = data.map(d => d.close)
  const minPrice = Math.min(...prices)
  const maxPrice = Math.max(...prices)
  
  // Calculate volume statistics
  const volumes = data.map(d => d.volume)
  const totalVolume = volumes.reduce((sum, vol) => sum + vol, 0)
  const avgVolume = totalVolume / volumes.length
  const minVolume = Math.min(...volumes)
  const maxVolume = Math.max(...volumes)
  
  // Calculate performance metrics
  const dailyChange = lastPoint.changePercent || 0
  
  // Weekly performance (if we have at least 5 data points)
  let weeklyChange: number | undefined
  if (data.length >= 5) {
    const weekAgo = data[data.length - 5]
    weeklyChange = ((lastPoint.close - weekAgo.close) / weekAgo.close) * 100
  }
  
  // Monthly performance (if we have at least 22 data points)
  let monthlyChange: number | undefined
  if (data.length >= 22) {
    const monthAgo = data[data.length - 22]
    monthlyChange = ((lastPoint.close - monthAgo.close) / monthAgo.close) * 100
  }
  
  // Yearly performance (if we have at least 252 data points)
  let yearlyChange: number | undefined
  if (data.length >= 252) {
    const yearAgo = data[data.length - 252]
    yearlyChange = ((lastPoint.close - yearAgo.close) / yearAgo.close) * 100
  }

  return {
    ticker,
    companyName,
    period: `${firstPoint.date} to ${lastPoint.date}`,
    dataPoints: data.length,
    priceRange: {
      min: minPrice,
      max: maxPrice,
      current: lastPoint.close,
      change: lastPoint.change || 0,
      changePercent: lastPoint.changePercent || 0
    },
    volumeStats: {
      total: totalVolume,
      average: avgVolume,
      min: minVolume,
      max: maxVolume
    },
    performance: {
      daily: dailyChange,
      weekly: weeklyChange,
      monthly: monthlyChange,
      yearly: yearlyChange
    }
  }
}