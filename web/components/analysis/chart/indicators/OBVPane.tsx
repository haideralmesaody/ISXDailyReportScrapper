/**
 * OBVPane - On-Balance Volume in separate pane
 * Shows cumulative volume indicator for buying/selling pressure
 * Rising OBV = bullish (accumulation), Falling OBV = bearish (distribution)
 */

'use client'

import { useMemo, useCallback } from 'react'
import { LineSeries } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateOBV } from '@/lib/indicators/calculations'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface OBVPaneProps {
  paneIndex: number
}

export function OBVPane({ paneIndex }: OBVPaneProps) {
  const { chartData, chart, containerHeight } = useChart()

  // Calculate OBV (no customizable settings - pure volume accumulation)
  const obvData = useMemo(() => {
    const closes = chartData.candlestickData.map(d => d.close)
    const volumes = chartData.volumeData.map(d => d.value)

    if (closes.length === 0 || volumes.length === 0) {
      console.log('[OBV] No data available')
      return []
    }

    if (closes.length !== volumes.length) {
      console.error('[OBV] Data mismatch:', closes.length, 'closes vs', volumes.length, 'volumes')
      return []
    }

    const obvValues = calculateOBV(closes, volumes)

    const obvData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (obvValues[i] !== null) {
        obvData.push({ time: d.time as string, value: obvValues[i]! })
      }
    })

    console.log('[OBV] Calculated:', obvData.length, 'points')
    return obvData
  }, [chartData])

  // Callback when OBV series is created - set pane height
  const handleOBVSeriesCreated = useCallback((obvSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      console.log('[OBV] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height (20% of container, max 90%)
      const panes = chart.panes()
      const obvPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.20)  // 20% for OBV
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      obvPane.setHeight(constrainedHeight)

      console.log('[OBV] ✅ Pane configured: height =', constrainedHeight, 'px')
    } catch (error) {
      console.error('[OBV] Error setting pane height:', error)
    }
  }, [chart, paneIndex, containerHeight])

  // OBV Line (teal-500 for volume indicator)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#14B8A6',  // teal-500
      lineWidth: 2,
      title: 'OBV',
      priceScaleId: 'obv',
    },
    data: obvData,
    paneIndex,
    onCreated: handleOBVSeriesCreated,
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0,  // Normal price scale mode
      scaleMargins: { top: 0.05, bottom: 0.05 },
    },
  })

  return (
    <PaneOverlay paneIndex={paneIndex}>
      <IndicatorSettingsPopover
        indicatorName="On-Balance Volume"
        settings={{}}
        fields={[]}
        onUpdate={() => {}}
        description="Cumulative volume indicator. Rising OBV indicates accumulation (bullish), falling OBV indicates distribution (bearish)."
      />
    </PaneOverlay>
  )
}
