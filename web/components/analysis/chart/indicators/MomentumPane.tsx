/**
 * MomentumPane - Momentum Oscillator (MOM) in separate pane
 * Displays absolute price change over customizable period
 * Shows zero line crossovers for trend changes
 */

'use client'

import { useMemo, useCallback } from 'react'
import { LineSeries, LineStyle } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateMomentum } from '@/lib/indicators/calculations'
import { useMomentumSettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface MomentumPaneProps {
  paneIndex: number
}

export function MomentumPane({ paneIndex }: MomentumPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useMomentumSettings()

  // Calculate Momentum with custom period
  const { momData, refLineData } = useMemo(() => {
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.momentumPeriod) {
      console.log('[MOM] Insufficient data: need', settings.momentumPeriod, 'have', closes.length)
      return { momData: [], refLineData: [] }
    }

    const momValues = calculateMomentum(closes, settings.momentumPeriod)

    const momData: Array<{ time: string; value: number }> = []

    chartData.candlestickData.forEach((d, i) => {
      if (momValues[i] !== null) {
        momData.push({ time: d.time as string, value: momValues[i]! })
      }
    })

    // Reference line data (zero line)
    const refLineData = momData.map(d => ({ time: d.time, value: 0 }))

    console.log('[MOM] Calculated:', momData.length, 'points (period:', settings.momentumPeriod, ')')
    return { momData, refLineData }
  }, [chartData, settings.momentumPeriod])

  // Callback when MOM series is created - add zero reference line
  const handleMOMSeriesCreated = useCallback((momSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      console.log('[MOM] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height
      const panes = chart.panes()
      const momPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      momPane.setHeight(constrainedHeight)

      console.log('[MOM] ✅ Pane configured: height =', constrainedHeight, 'px')

      // Add zero line (gray dashed)
      const zeroLine = chart.addSeries(LineSeries, {
        color: '#6B7280',
        lineWidth: 1,
        lineStyle: LineStyle.Dashed,
        priceScaleId: 'momentum',
      }, paneIndex)

      requestAnimationFrame(() => {
        zeroLine.setData(refLineData)
      })

      console.log('[MOM] ✅ Added zero reference line (pane', paneIndex, ')')
    } catch (error) {
      console.error('[MOM] Error creating zero line:', error)
    }
  }, [chart, paneIndex, containerHeight, refLineData])

  // MOM Line (amber-500)
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#F59E0B',  // amber-500
      lineWidth: 2,
      title: `MOM (${settings.momentumPeriod})`,
      priceScaleId: 'momentum',
    },
    data: momData,
    paneIndex,
    onCreated: handleMOMSeriesCreated,
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
        indicatorName="Momentum"
        settings={settings}
        fields={[
          { name: 'momentumPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
