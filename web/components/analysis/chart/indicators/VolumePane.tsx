/**
 * VolumePane - Volume histogram in separate pane
 * Manages volume series in Pane 1
 * Implements TradingView-style intensity coloring based on volume MA
 */

'use client'

import { useMemo, useCallback } from 'react'
import { HistogramSeries } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateSMA } from '@/lib/indicators/calculations'
import { logger } from '@/lib/utils/logger'

interface VolumePaneProps {
  paneIndex: number
}

// Apply intensity to hex color by adjusting brightness
function applyIntensity(color: string, intensity: number): string {
  const normalized = color.trim()

  const rgbaMatch = normalized.match(
    /^rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*([0-9.]+))?\s*\)$/i
  )
  if (rgbaMatch) {
    const r = Number(rgbaMatch[1])
    const g = Number(rgbaMatch[2])
    const b = Number(rgbaMatch[3])
    const alpha = rgbaMatch[4] !== undefined ? Number(rgbaMatch[4]) : 1

    const adjustedR = Math.round(r * intensity + 255 * (1 - intensity))
    const adjustedG = Math.round(g * intensity + 255 * (1 - intensity))
    const adjustedB = Math.round(b * intensity + 255 * (1 - intensity))

    return `rgba(${adjustedR}, ${adjustedG}, ${adjustedB}, ${alpha})`
  }

  const hexMatch = normalized.match(/^#([0-9a-f]{6})$/i)
  if (hexMatch) {
    const hex = hexMatch[1]
    const r = parseInt(hex.slice(0, 2), 16)
    const g = parseInt(hex.slice(2, 4), 16)
    const b = parseInt(hex.slice(4, 6), 16)

    const adjustedR = Math.round(r * intensity + 255 * (1 - intensity))
    const adjustedG = Math.round(g * intensity + 255 * (1 - intensity))
    const adjustedB = Math.round(b * intensity + 255 * (1 - intensity))

    return `#${adjustedR.toString(16).padStart(2, '0')}${adjustedG.toString(16).padStart(2, '0')}${adjustedB.toString(16).padStart(2, '0')}`
  }

  return color
}

export function VolumePane({ paneIndex }: VolumePaneProps) {
  const { chartData, theme, chart, containerHeight } = useChart()

  // Calculate 20-period volume MA for intensity calculation
  const volumeMA = useMemo(() => {
    const volumes = chartData.volumeData.map((point) => Number(point.value) || 0)
    const ma = calculateSMA(volumes, 20) // 20-period MA (TradingView standard)
    logger.log('[Volume] Calculated 20-period MA')
    return ma
  }, [chartData])

  // Prepare volume data with intensity-based coloring
  const volumeDataWithIntensity = useMemo(() => {
    const priceData = chartData.candlestickData

    const data = chartData.volumeData
      .map((d, i) => {
        const volume = Number(d.value) || 0
        const ma = volumeMA[i]
        const candle = priceData[i]
        if (!candle) return null
        const closeUp = candle.close >= candle.open

        // Calculate intensity ratio (volume / MA)
        let intensity = 0.5 // default medium
        if (ma && ma > 0) {
          const ratio = volume / ma
          if (ratio > 1.5) intensity = 1.0        // Very high: >150% of MA
          else if (ratio > 1.0) intensity = 0.8   // High: 100-150% of MA
          else if (ratio > 0.5) intensity = 0.5   // Normal: 50-100% of MA
          else if (ratio > 0.25) intensity = 0.3  // Low: 25-50% of MA
          else intensity = 0.15                   // Very low: <25% of MA
        }

        // Apply intensity to base color (green for up, red for down)
        const baseColor = closeUp ? theme.volumeUpColor : theme.volumeDownColor
        const color = applyIntensity(baseColor, intensity)

        return {
          ...d,
          value: volume,
          color
        }
      })
      .filter((point): point is NonNullable<typeof point> => Boolean(point))

    logger.log('[Volume] Prepared', data.length, 'volume bars with intensity coloring')
    return data
  }, [chartData, volumeMA, theme])

  // Set pane height (price scale now configured in useSeries hook)
  const handleSeriesCreated = useCallback((series: any) => {
    if (!chart) return

    try {
      const panes = chart.panes()
      if (panes.length > paneIndex) {
        const volumePane = panes[paneIndex]

        // Calculate height with 90% maximum constraint
        const maxHeight = Math.floor(containerHeight * 0.90)
        const desiredHeight = Math.floor(containerHeight * 0.20)  // 20% default
        const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))

        volumePane.setHeight(constrainedHeight)

        logger.log('[Volume] ✅ Pane configured: height =', constrainedHeight, 'px')
      }
    } catch (error) {
      logger.error('[Volume] Error setting pane height:', error)
    }
  }, [chart, paneIndex, containerHeight])

  useSeries({
    seriesType: HistogramSeries,
    seriesOptions: {
      color: theme.volumeUpColor,
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',  // Independent price scale
    },
    data: volumeDataWithIntensity,
    paneIndex,
    onCreated: handleSeriesCreated,
    // ✅ TradingView v5.0 compliant: Configure price scale synchronously after series creation
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0,
      scaleMargins: { top: 0.1, bottom: 0.0 },  // Volume starts at 0
    },
  })

  return null
}
