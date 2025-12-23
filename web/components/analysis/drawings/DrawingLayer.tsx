'use client'

import { useEffect } from 'react'
import type { ISeriesApi } from 'lightweight-charts'
import { useChart } from '@/components/analysis/chart/ChartContext'
import { DrawingLayerPrimitive } from './drawing-layer'

type PrimitiveHost = ISeriesApi<'Candlestick'> & {
  attachPrimitive?: (primitive: DrawingLayerPrimitive) => void
  detachPrimitive?: (primitive: DrawingLayerPrimitive) => void
}

export function DrawingLayer() {
  const { chart, candlestickSeries, isReady } = useChart()

  useEffect(() => {
    if (!chart || !candlestickSeries || !isReady) return

    const primitive = new DrawingLayerPrimitive(chart, candlestickSeries)
    const host = candlestickSeries as unknown as PrimitiveHost
    if (typeof host.attachPrimitive !== 'function') {
      return
    }

    host.attachPrimitive(primitive)

    return () => {
      if (host.detachPrimitive) {
        host.detachPrimitive(primitive)
      }
    }
  }, [chart, candlestickSeries, isReady])

  return null
}
