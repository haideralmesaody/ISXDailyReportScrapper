/**
 * ChartExportButton Component
 *
 * Provides PNG export functionality using TradingView Lightweight Charts
 * native takeScreenshot() API (v5.0+).
 *
 * Features:
 * - Export chart as PNG image
 * - Automatic filename with ticker and date
 * - Loading state during export
 * - Error handling with toast notifications
 * - Disabled when chart not ready
 */

'use client'

import { useState, useCallback } from 'react'
import { Download, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useToast } from '@/lib/hooks/use-toast'
import type { IChartApi } from 'lightweight-charts'

interface ChartExportButtonProps {
  chart: IChartApi | null
  ticker: string
  disabled?: boolean
}

export function ChartExportButton({ chart, ticker, disabled }: ChartExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false)
  const { toast } = useToast()

  const handleExportPNG = useCallback(async () => {
    if (!chart || disabled) return

    setIsExporting(true)

    try {
      // Use TradingView's native takeScreenshot() API
      // Returns HTMLCanvasElement with chart drawn (excluding crosshair)
      const canvas = chart.takeScreenshot()

      // Convert canvas to data URL (base64 PNG)
      const dataUrl = canvas.toDataURL('image/png')

      // Create download link
      const link = document.createElement('a')

      // Generate filename: {ticker}-{date}-chart.png
      const date = new Date().toISOString().split('T')[0] // YYYY-MM-DD
      link.download = `${ticker}-${date}-chart.png`
      link.href = dataUrl

      // Trigger download
      link.click()

      // Success notification
      toast({
        title: 'Chart Exported',
        description: `Saved as ${link.download}`,
        duration: 3000,
      })

      console.log('[ChartExport] Successfully exported chart:', link.download)
    } catch (error) {
      console.error('[ChartExport] Export failed:', error)

      toast({
        title: 'Export Failed',
        description: error instanceof Error ? error.message : 'Could not export chart as PNG',
        variant: 'destructive',
        duration: 5000,
      })
    } finally {
      setIsExporting(false)
    }
  }, [chart, ticker, disabled, toast])

  const isDisabled = !chart || disabled || isExporting

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportPNG}
            disabled={isDisabled}
            className="gap-2"
          >
            {isExporting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Exporting...
              </>
            ) : (
              <>
                <Download className="h-4 w-4" />
                Export PNG
              </>
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>Export chart as PNG image</p>
          <p className="text-xs text-muted-foreground mt-1">
            Includes all visible indicators
          </p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
