/**
 * Chart Export Utilities
 * Provides functionality to export charts as images
 */

import { IChartApi } from 'lightweight-charts'

export interface ExportOptions {
  backgroundColor?: string
  fileName?: string
  format?: 'png' | 'jpeg'
  quality?: number // 0-1 for JPEG quality
}

/**
 * Export lightweight chart as image
 */
export async function exportChartAsImage(
  chart: IChartApi,
  options: ExportOptions = {}
): Promise<void> {
  const {
    backgroundColor = '#ffffff',
    fileName = `chart-${new Date().toISOString().split('T')[0]}`,
    format = 'png',
    quality = 0.95
  } = options

  try {
    // Take a screenshot of the chart
    const canvas = await chart.takeScreenshot()
    
    if (!canvas) {
      throw new Error('Failed to capture chart screenshot')
    }

    // Convert canvas to blob
    const blob = await new Promise<Blob | null>((resolve) => {
      canvas.toBlob(
        (blob) => resolve(blob),
        format === 'png' ? 'image/png' : 'image/jpeg',
        quality
      )
    })

    if (!blob) {
      throw new Error('Failed to convert chart to image')
    }

    // Create download link
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${fileName}.${format}`
    
    // Trigger download
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    
    // Clean up
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Error exporting chart:', error)
    throw new Error('Failed to export chart as image')
  }
}

/**
 * Export chart with custom dimensions
 */
export async function exportChartWithDimensions(
  chart: IChartApi,
  width: number,
  height: number,
  options: ExportOptions = {}
): Promise<void> {
  // Store original dimensions
  const originalSize = {
    width: chart.options().width,
    height: chart.options().height
  }

  try {
    // Resize chart for export
    chart.applyOptions({
      width,
      height
    })

    // Wait for chart to re-render
    await new Promise(resolve => setTimeout(resolve, 100))

    // Export with new dimensions
    await exportChartAsImage(chart, options)
  } finally {
    // Restore original dimensions
    chart.applyOptions(originalSize)
  }
}

/**
 * Copy chart image to clipboard
 */
export async function copyChartToClipboard(chart: IChartApi): Promise<void> {
  try {
    const canvas = await chart.takeScreenshot()
    
    if (!canvas) {
      throw new Error('Failed to capture chart screenshot')
    }

    // Convert canvas to blob
    const blob = await new Promise<Blob | null>((resolve) => {
      canvas.toBlob((blob) => resolve(blob), 'image/png')
    })

    if (!blob) {
      throw new Error('Failed to convert chart to image')
    }

    // Copy to clipboard using Clipboard API
    if (navigator.clipboard && window.ClipboardItem) {
      const item = new ClipboardItem({ 'image/png': blob })
      await navigator.clipboard.write([item])
    } else {
      throw new Error('Clipboard API not supported')
    }
  } catch (error) {
    console.error('Error copying chart to clipboard:', error)
    throw new Error('Failed to copy chart to clipboard')
  }
}

/**
 * Generate shareable chart configuration
 */
export function generateShareableLink(
  ticker: string,
  timeRange: string,
  chartType: string,
  indicators: string[]
): string {
  const params = new URLSearchParams({
    ticker,
    range: timeRange,
    type: chartType,
    indicators: indicators.join(',')
  })

  const baseUrl = window.location.origin
  return `${baseUrl}/analysis?${params.toString()}`
}

/**
 * Parse shareable link parameters
 */
export function parseShareableLink(url: string): {
  ticker?: string
  timeRange?: string
  chartType?: string
  indicators?: string[]
} {
  const urlObj = new URL(url)
  const params = urlObj.searchParams

  return {
    ticker: params.get('ticker') || undefined,
    timeRange: params.get('range') || undefined,
    chartType: params.get('type') || undefined,
    indicators: params.get('indicators')?.split(',').filter(Boolean) || undefined
  }
}