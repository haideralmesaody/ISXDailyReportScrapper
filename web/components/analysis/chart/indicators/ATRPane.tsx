/**
 * ATRPane - Average True Range in separate pane
 * Simple single-line volatility indicator
 *
 * ARCHITECTURE NOTE:
 * - Uses SINGLE useSeries hook for ATR line (creates pane + price scale)
 * - No additional series needed (unlike MACD/RSI which have multiple lines)
 * - Simpler pattern similar to VolumePane
 */

'use client'

import { useMemo, useCallback } from 'react'
import { LineSeries } from 'lightweight-charts'
import type { ISeriesApi } from 'lightweight-charts'
import { useSeries } from '@/lib/hooks/use-series'
import { useChart } from '../ChartContext'
import { calculateATR } from '@/lib/indicators/calculations'
import { useVolatilitySettings } from '@/lib/hooks/use-indicator-settings'
import { PaneOverlay } from './PaneOverlay'
import { IndicatorSettingsPopover } from './IndicatorSettingsPopover'

interface ATRPaneProps {
  paneIndex: number
}

export function ATRPane({ paneIndex }: ATRPaneProps) {
  const { chartData, chart, containerHeight } = useChart()
  const { settings, updateSettings } = useVolatilitySettings()

  // Calculate ATR values using custom period from settings
  const atrData = useMemo(() => {
    const highs = chartData.candlestickData.map(d => d.high)
    const lows = chartData.candlestickData.map(d => d.low)
    const closes = chartData.candlestickData.map(d => d.close)

    if (closes.length < settings.atrPeriod) {
      console.log('[ATR] Insufficient data: need', settings.atrPeriod, 'have', closes.length)
      return []
    }

    const atrValues = calculateATR(highs, lows, closes, settings.atrPeriod)

    const data = chartData.candlestickData
      .map((d, i) => ({
        time: d.time as string,
        value: atrValues[i] ?? 0
      }))
      .filter(d => d.value > 0)

    console.log('[ATR] Calculated', data.length, 'ATR points')
    return data
  }, [chartData, settings.atrPeriod])

  // Callback when ATR series is created - set pane height
  const handleATRSeriesCreated = useCallback((atrSeries: ISeriesApi<any>) => {
    if (!chart) return

    try {
      // Use the paneIndex prop directly
      console.log('[ATR] 🎯 Using paneIndex prop:', paneIndex, '(total panes:', chart.panes().length, ')')

      // Set pane height using paneIndex prop
      const panes = chart.panes()
      const atrPane = panes[paneIndex]
      const maxHeight = Math.floor(containerHeight * 0.90)
      const desiredHeight = Math.floor(containerHeight * 0.15)  // 15% default
      const constrainedHeight = Math.max(30, Math.min(desiredHeight, maxHeight))
      atrPane.setHeight(constrainedHeight)

      console.log('[ATR] ✅ Pane configured: height =', constrainedHeight, 'px')
    } catch (error) {
      console.error('[ATR] Error configuring pane:', error)
    }
  }, [chart, paneIndex, containerHeight])

  // ATR Line (yellow) - Simple single series
  useSeries({
    seriesType: LineSeries,
    seriesOptions: {
      color: '#EAB308',
      lineWidth: 2,
      title: `ATR (${settings.atrPeriod})`,
      priceScaleId: 'atr',
    },
    data: atrData,
    paneIndex,
    onCreated: handleATRSeriesCreated,
    // ✅ TradingView v5.0 compliant: Configure price scale synchronously after series creation
    priceScaleOptions: {
      visible: true,
      autoScale: true,
      borderVisible: true,
      mode: 0,
      scaleMargins: { top: 0.1, bottom: 0.1 },
    },
  })

  return (
    <PaneOverlay paneIndex={paneIndex}>
      <IndicatorSettingsPopover
        indicatorName="ATR"
        settings={settings}
        fields={[
          { name: 'atrPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
        ]}
        onUpdate={(newSettings) => updateSettings(newSettings)}
      />
    </PaneOverlay>
  )
}
