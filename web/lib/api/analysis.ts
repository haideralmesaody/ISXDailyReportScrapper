/**
 * API functions for Analysis feature
 * Handles fetching ticker summary and historical data
 * Following CLAUDE.md standards - using fetch API like reports.ts
 */

import { API_BASE_URL } from '@/lib/constants/api'
import type { TickerSummary, TickerHistoricalData } from '@/types/analysis'

/**
 * Fetch ticker summary data from ticker_summary.csv
 */
export async function fetchTickerSummary(): Promise<TickerSummary[]> {
  try {
    // Try direct fetch first since we know the exact path
    const directPath = 'summary/ticker/ticker_summary.csv'
    const directResponse = await fetch(`${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(directPath)}`, {
      method: 'GET',
      credentials: 'include',
    })
    
    if (directResponse.ok) {
      const csvContent = await directResponse.text()
      const tickers = parseTickerSummaryCSV(csvContent)
      return tickers.map(ticker => ({
        ...ticker,
        ChangePercent: ticker.ChangePercent ?? calculateChangePercent(ticker)
      }))
    }
    
    // Fallback to search if direct fetch fails
    const params = new URLSearchParams({
      type: 'summary',
      search: 'ticker_summary'
    })
    
    const response = await fetch(`${API_BASE_URL}/api/data/reports?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })
    
    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch ticker summary: ${response.statusText}`)
    }
    
    const responseData = await response.json()
    
    // Handle wrapped response format from backend { status: "success", data: [...], count: ... }
    const reports = responseData.data || responseData || []
    
    // Check if reports is an array before using find
    if (!Array.isArray(reports) || reports.length === 0) {
      throw new Error('No ticker summary data available')
    }
    
    // Find the ticker_summary.csv file
    const summaryReport = reports.find((r: any) => r.name === 'ticker_summary.csv') || reports[0]
    
    if (!summaryReport) {
      throw new Error('No ticker summary data available')
    }
    
    // Get the content of the ticker_summary.csv file (it's in a nested directory)
    // The file is at summary/ticker/ticker_summary.csv
    const filePath = summaryReport.path || `summary/ticker/${summaryReport.name}`
    const contentResponse = await fetch(`${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(filePath)}`, {
      method: 'GET',
      credentials: 'include',
    })
    
    if (!contentResponse.ok) {
      throw new Error(`Failed to fetch report content: ${contentResponse.statusText}`)
    }
    
    const csvContent = await contentResponse.text()
    
    // Parse CSV content
    const tickers = parseTickerSummaryCSV(csvContent)
    
    // Calculate change percent if not present
    return tickers.map(ticker => ({
      ...ticker,
      ChangePercent: ticker.ChangePercent ?? calculateChangePercent(ticker)
    }))
  } catch (error) {
    console.error('Failed to fetch ticker summary:', error)
    throw error
  }
}

/**
 * Fetch combined market data from isx_combined_data.csv
 */
export async function fetchCombinedData(): Promise<Map<string, TickerHistoricalData[]>> {
  try {
    // Try direct fetch first since we know the exact path
    const directPath = 'combined/isx_combined_data.csv'
    const directResponse = await fetch(`${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(directPath)}`, {
      method: 'GET',
      credentials: 'include',
    })
    
    if (directResponse.ok) {
      const csvContent = await directResponse.text()
      return parseCombinedDataCSV(csvContent)
    }
    
    // Fallback to search if direct fetch fails
    const params = new URLSearchParams({
      type: 'combined',
      search: 'isx_combined_data'
    })
    
    const response = await fetch(`${API_BASE_URL}/api/data/reports?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })
    
    if (!response.ok) {
      const error = await response.json().catch(() => null)
      throw new Error(error?.detail || `Failed to fetch combined data: ${response.statusText}`)
    }
    
    const responseData = await response.json()
    
    // Handle wrapped response format from backend { status: "success", data: [...], count: ... }
    const reports = responseData.data || responseData
    
    // Find the isx_combined_data.csv file
    const combinedReport = reports.find((r: any) => r.name === 'isx_combined_data.csv') || reports[0]
    
    if (!combinedReport) {
      throw new Error('No combined data available')
    }
    
    // Get the content of the isx_combined_data.csv file (it's in a nested directory)
    // The file is at combined/isx_combined_data.csv
    const filePath = combinedReport.path || `combined/${combinedReport.name}`
    const contentResponse = await fetch(`${API_BASE_URL}/api/data/download/reports/${encodeURIComponent(filePath)}`, {
      method: 'GET',
      credentials: 'include',
    })
    
    if (!contentResponse.ok) {
      throw new Error(`Failed to fetch combined data content: ${contentResponse.statusText}`)
    }
    
    const csvContent = await contentResponse.text()
    
    // Parse CSV content and organize by ticker
    return parseCombinedDataCSV(csvContent)
  } catch (error) {
    console.error('Failed to fetch combined data:', error)
    throw error
  }
}

/**
 * Fetch historical trading data for a specific ticker
 */
export async function fetchTickerHistory(ticker: string): Promise<TickerHistoricalData[]> {
  try {
    // First try to get data from combined CSV (most efficient)
    const combinedData = await fetchCombinedData()
    const tickerData = combinedData.get(ticker)
    
    if (tickerData && tickerData.length > 0) {
      return tickerData
    }
    
    // Fallback to ticker-specific file
    const params = new URLSearchParams({
      type: 'ticker',
      search: `${ticker}_trading_history`
    })
    
    const response = await fetch(`${API_BASE_URL}/api/data/reports?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })
    
    if (!response.ok) {
      // Fallback to fetching from daily reports
      return fetchTickerFromDailyReports(ticker)
    }
    
    const responseData = await response.json()
    
    // Handle wrapped response format from backend { status: "success", data: [...], count: ... }
    const reports = responseData.data || responseData
    
    if (!reports || reports.length === 0) {
      // Fallback to fetching from daily reports
      return fetchTickerFromDailyReports(ticker)
    }
    
    // Get the content of the ticker history file (use 'name' field, not 'filename')
    const historyReport = reports[0]
    const contentResponse = await fetch(`${API_BASE_URL}/api/data/download/reports/${historyReport.name}`, {
      method: 'GET',
      credentials: 'include',
    })
    
    if (!contentResponse.ok) {
      // Fallback to fetching from daily reports
      return fetchTickerFromDailyReports(ticker)
    }
    
    const csvContent = await contentResponse.text()
    
    // Parse CSV content
    return parseTickerHistoryCSV(csvContent)
  } catch (error) {
    console.error(`Failed to fetch history for ${ticker}:`, error)
    // Try fallback to daily reports if combined data fails
    try {
      return await fetchTickerFromDailyReports(ticker)
    } catch (fallbackError) {
      console.error(`Fallback also failed for ${ticker}:`, fallbackError)
      throw error
    }
  }
}

/**
 * Fallback: Extract ticker data from daily reports
 */
async function fetchTickerFromDailyReports(ticker: string): Promise<TickerHistoricalData[]> {
  try {
    // Fetch all daily reports using fetch API
    const params = new URLSearchParams({
      type: 'daily',
      limit: '100' // Get last 100 days
    })
    
    const response = await fetch(`${API_BASE_URL}/api/data/reports?${params}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    })
    
    if (!response.ok) {
      throw new Error(`Failed to fetch daily reports: ${response.statusText}`)
    }
    
    const responseData = await response.json()
    
    // Handle wrapped response format from backend { status: "success", data: [...], count: ... }
    const reports = responseData.data || responseData
    
    if (!reports || reports.length === 0) {
      throw new Error('No daily reports available')
    }
    
    const historicalData: TickerHistoricalData[] = []
    
    // Process each daily report to extract ticker data
    for (const report of reports) {
      try {
        const contentResponse = await fetch(`${API_BASE_URL}/api/data/download/reports/${report.name}`, {
          method: 'GET',
          credentials: 'include',
        })
        
        if (!contentResponse.ok) {
          console.warn(`Failed to fetch report ${report.name}`)
          continue
        }
        
        const csvContent = await contentResponse.text()
        const dayData = extractTickerFromDaily(csvContent, ticker)
        
        if (dayData) {
          historicalData.push(dayData)
        }
      } catch (err) {
        // Skip failed reports
        console.warn(`Failed to process report ${report.name}:`, err)
      }
    }
    
    // Sort by date ascending
    historicalData.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
    
    return historicalData
  } catch (error) {
    console.error(`Failed to fetch daily reports for ${ticker}:`, error)
    throw error
  }
}

/**
 * Parse ticker summary CSV content
 */
function parseTickerSummaryCSV(csvContent: string): TickerSummary[] {
  const lines = csvContent.trim().split('\n')
  if (lines.length < 2) return []
  
  const headers = lines[0].split(',').map(h => h.trim())
  const tickers: TickerSummary[] = []
  
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim()
    if (!line) continue
    
    // Parse CSV line handling quoted fields (like Last10Days)
    const parts: string[] = []
    let currentPart = ''
    let inQuotes = false
    
    for (let j = 0; j < line.length; j++) {
      const char = line[j]
      
      if (char === '"') {
        inQuotes = !inQuotes
      } else if (char === ',' && !inQuotes) {
        parts.push(currentPart.trim())
        currentPart = ''
      } else {
        currentPart += char
      }
    }
    // Add the last part
    parts.push(currentPart.trim())
    
    // Remove quotes from any field if present
    for (let k = 0; k < parts.length; k++) {
      parts[k] = parts[k].replace(/^"|"$/g, '')
    }
    
    // Create ticker object with available data
    // CSV Format: Ticker,CompanyName,LastPrice,LastDate,TradingDays,Last10Days,
    //             TotalVolume,TotalValue,AveragePrice,HighestPrice,LowestPrice,
    //             Change,ChangePercent,LastTradingStatus
    const ticker: TickerSummary = {
      Ticker: parts[0] || '',
      CompanyName: parts[1] || '',
      LastPrice: parseFloat(parts[2]) || 0,
      LastDate: parts[3] || '',
      TradingDays: parseInt(parts[4]) || 0,
      Last10Days: parts[5] || '',
      TotalVolume: parseInt(parts[6]) || 0,
      TotalValue: parseFloat(parts[7]) || 0,
      AveragePrice: parseFloat(parts[8]) || 0,
      HighestPrice: parseFloat(parts[9]) || 0,
      LowestPrice: parseFloat(parts[10]) || 0,
      // Use the actual ChangePercent from the CSV (field 12)
      ChangePercent: parseFloat(parts[12]) || 0,
      // Add trading status for UI display logic
      LastTradingStatus: parts[13] === 'true'
    }
    
    tickers.push(ticker)
  }
  
  return tickers
}

/**
 * Parse ticker history CSV content
 */
function parseTickerHistoryCSV(csvContent: string): TickerHistoricalData[] {
  const lines = csvContent.trim().split('\n')
  if (lines.length < 2) return []
  
  const headers = lines[0].split(',').map(h => h.trim())
  const history: TickerHistoricalData[] = []
  
  for (let i = 1; i < lines.length; i++) {
    const values = lines[i].split(',').map(v => v.trim())
    if (values.length !== headers.length) continue
    
    const data: any = {}
    headers.forEach((header, index) => {
      const value = values[index]
      
      // Map CSV headers to our interface (supports both formats)
      const fieldMap: Record<string, string> = {
        'Date': 'date',
        'Open': 'open',
        'OpenPrice': 'open',      // ISX combined format
        'High': 'high',
        'HighPrice': 'high',      // ISX combined format
        'Low': 'low',
        'LowPrice': 'low',        // ISX combined format
        'Close': 'close',
        'ClosePrice': 'close',    // ISX combined format
        'Volume': 'volume',
        'Value': 'value',
        'Trades': 'trades',
        'NumTrades': 'trades',    // ISX combined format
        'Change': 'change',
        'Change%': 'changePercent',
        'ChangePercent': 'changePercent'  // ISX combined format
      }
      
      const field = fieldMap[header] || header.toLowerCase()
      
      // Parse numeric fields with proper decimal handling
      if (['open', 'high', 'low', 'close'].includes(field)) {
        // Parse and ensure 2 decimal places for price fields without precision loss
        const num = parseFloat(value) || 0
        // Use Math.round to avoid floating point precision issues
        data[field] = Math.round(num * 100) / 100
      } else if (['volume', 'trades'].includes(field)) {
        data[field] = parseInt(value) || 0
      } else if (['change', 'changePercent', 'value'].includes(field)) {
        // Also use 2 decimals for these financial fields
        const num = parseFloat(value) || 0
        // Use Math.round to avoid floating point precision issues
        data[field] = Math.round(num * 100) / 100
      } else {
        data[field] = value
      }
    })
    
    history.push(data as TickerHistoricalData)
  }
  
  return history
}

/**
 * Extract ticker data from daily report CSV
 */
function extractTickerFromDaily(csvContent: string, ticker: string): TickerHistoricalData | null {
  const lines = csvContent.trim().split('\n')
  if (lines.length < 2) return null
  
  const headers = lines[0].split(',').map(h => h.trim())
  const tickerIndex = headers.findIndex(h => h === 'Symbol' || h === 'Ticker')
  
  if (tickerIndex === -1) return null
  
  for (let i = 1; i < lines.length; i++) {
    const values = lines[i].split(',').map(v => v.trim())
    
    if (values[tickerIndex] === ticker) {
      const data: any = {}
      
      headers.forEach((header, index) => {
        const value = values[index]
        
        // Map to our interface fields (supports both formats)
        const fieldMap: Record<string, string> = {
          'Date': 'date',
          'Open': 'open',
          'OpenPrice': 'open',      // ISX combined format
          'High': 'high',
          'HighPrice': 'high',      // ISX combined format
          'Low': 'low',
          'LowPrice': 'low',        // ISX combined format
          'Close': 'close',
          'ClosePrice': 'close',    // ISX combined format
          'Volume': 'volume',
          'Value': 'value',
          'Trades': 'trades',
          'NumTrades': 'trades',    // ISX combined format
          'Change': 'change',
          'Change%': 'changePercent',
          'ChangePercent': 'changePercent'  // ISX combined format
        }
        
        const field = fieldMap[header]
        if (field) {
          if (['open', 'high', 'low', 'close'].includes(field)) {
            // Parse and ensure 2 decimal places for price fields
            const num = parseFloat(value) || 0
            data[field] = parseFloat(num.toFixed(2))
          } else if (['volume', 'trades'].includes(field)) {
            data[field] = parseInt(value) || 0
          } else if (['change', 'changePercent', 'value'].includes(field)) {
            // Also use 2 decimals for these financial fields
            const num = parseFloat(value) || 0
            data[field] = parseFloat(num.toFixed(2))
          } else {
            data[field] = value
          }
        }
      })
      
      return data as TickerHistoricalData
    }
  }
  
  return null
}

/**
 * Parse combined data CSV content and organize by ticker
 */
function parseCombinedDataCSV(csvContent: string): Map<string, TickerHistoricalData[]> {
  const lines = csvContent.trim().split('\n')
  if (lines.length < 2) return new Map()
  
  const headers = lines[0].split(',').map(h => h.trim())
  const tickerDataMap = new Map<string, TickerHistoricalData[]>()
  
  // Find column indices
  const symbolIdx = headers.findIndex(h => h === 'Symbol' || h === 'Ticker')
  const dateIdx = headers.findIndex(h => h === 'Date')
  
  if (symbolIdx === -1 || dateIdx === -1) {
    console.error('Combined CSV missing required columns')
    return tickerDataMap
  }
  
  
  for (let i = 1; i < lines.length; i++) {
    const values = lines[i].split(',').map(v => v.trim())
    if (values.length !== headers.length) continue
    
    const ticker = values[symbolIdx]
    if (!ticker) continue
    
    const data: any = {}
    headers.forEach((header, index) => {
      const value = values[index]
      
      // Map CSV headers to our interface (supports both formats)
      const fieldMap: Record<string, string> = {
        'Date': 'date',
        'Open': 'open',
        'OpenPrice': 'open',      // ISX combined format
        'High': 'high',
        'HighPrice': 'high',      // ISX combined format
        'Low': 'low',
        'LowPrice': 'low',        // ISX combined format
        'Close': 'close',
        'ClosePrice': 'close',    // ISX combined format
        'Volume': 'volume',
        'Value': 'value',
        'Trades': 'trades',
        'NumTrades': 'trades',    // ISX combined format
        'Change': 'change',
        'Change%': 'changePercent',
        'ChangePercent': 'changePercent'  // ISX combined format
      }
      
      const field = fieldMap[header] || header.toLowerCase()
      
      // Parse numeric fields with proper decimal handling
      if (['open', 'high', 'low', 'close'].includes(field)) {
        // Parse and ensure 2 decimal places for price fields without precision loss
        const num = parseFloat(value) || 0
        // Use Math.round to avoid floating point precision issues
        data[field] = Math.round(num * 100) / 100
      } else if (['volume', 'trades'].includes(field)) {
        data[field] = parseInt(value) || 0
      } else if (['change', 'changePercent', 'value'].includes(field)) {
        // Also use 2 decimals for these financial fields
        const num = parseFloat(value) || 0
        // Use Math.round to avoid floating point precision issues
        data[field] = Math.round(num * 100) / 100
      } else if (field !== 'symbol' && field !== 'ticker') {
        data[field] = value
      }
    })
    
    // Add to ticker's array
    if (!tickerDataMap.has(ticker)) {
      tickerDataMap.set(ticker, [])
    }
    tickerDataMap.get(ticker)!.push(data as TickerHistoricalData)
  }
  
  // Sort each ticker's data by date
  tickerDataMap.forEach((dataArray) => {
    dataArray.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
  })
  
  return tickerDataMap
}

/**
 * Calculate change percent from ticker summary
 */
function calculateChangePercent(ticker: TickerSummary): number {
  if (!ticker.Last10Days) return 0
  
  const prices = ticker.Last10Days.split(',').map(p => parseFloat(p.trim())).filter(p => !isNaN(p))
  if (prices.length < 2) return 0
  
  const firstPrice = prices[0]
  const lastPrice = prices[prices.length - 1]
  
  if (firstPrice === 0) return 0
  
  return ((lastPrice - firstPrice) / firstPrice) * 100
}