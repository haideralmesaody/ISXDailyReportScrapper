/**
 * ROCPane - Rate of Change (ROC) in separate pane
 * Displays percentage price change over customizable period
 * Shows zero line crossovers for trend changes
 */

'use client'

import { useMemo, useCallback } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateROC } from '@/lib/indicators/calculations'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface ROCPaneProps {
  paneIndex: number
}

export function ROCPane({ paneIndex }: ROCPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Calculate ROC with custom period
  const { rocData, refLineData } = useMemo(() => {
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.rocPeriod) {
      console.log('[ROC] Insufficient data: need', settings.rocPeriod, 'have', closes.length)
      return { rocData: [], refLineData: [] }
    }

    const rocValues = calculateROC(closes, settings.rocPeriod)

    const rocData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (rocValues[i] !== null) {
        rocData.push({ time: d.time as string, value: rocValues[i]! })
      }
    })

    // Reference line data (zero line)
    const refLineData = rocData.map(d => ({ time: d.time, value: 0 }))

    console.log('[ROC] Calculated:', rocData.length, 'points (period:', settings.rocPeriod, ')')
    return { rocData, refLineData }
  }, [chartData, settings.rocPeriod])

  // Callback when ROC series is created - add zero reference line
  const handleROCSeriesCreated = useCallback((rocSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      console.log('[ROC] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height
      const panes = chart.panes()
      const rocPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      rocPane.setHeight(constrainedHeight)

      console.log('[ROC] ✅ Pane configured: height =', constrainedHeight, 'px')

      // Add zero line (gray dashed)
      const zeroLine = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'roc',
      }, paneIndex)

      requestAnimationFrame(() => {
        zeroLine.setData(refLineData)
      })

      console.log('[ROC] ✅ Added zero reference line (pane', paneIndex, ')')
    } catch (error) {
      console.error('[ROC] Error creating zero line:', error)
    }
  }, [chart, paneIndex, containerHeight, refLineData])

  // ROC Line (lime-500)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#84CC16',  // lime-500
      lineWidth: 2,
      title: `ROC (${settings.rocPeriod})`,
      priceScaleId: 'roc',
    },
    data: rocData,
    paneIndex,
    onCreated: handleROCSeriesCreated,
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0,
      scaleMargins: { top: 0.05, bottom: 0.05 },
    },
  })

  return (
    <PaneOverlay paneIndex={paneIndex}>
      <IndicatorSettingsPopover
        indicatorName="ROC"
        settings={settings}
        fields={[
          { name: 'rocPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
