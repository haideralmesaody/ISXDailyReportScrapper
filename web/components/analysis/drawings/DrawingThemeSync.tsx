'use client'

import { useEffect } from 'react'
import { useChart } from '@/components/analysis/chart/ChartContext'
import { useDrawingStore } from './store'

const isDarkHex = (value: string): boolean => {
  const hex = value.trim()
  const match = hex.match(/^#([0-9a-f]{6})$/i)
  if (!match) return false
  const raw = match[1]
  const r = parseInt(raw.slice(0, 2), 16)
  const g = parseInt(raw.slice(2, 4), 16)
  const b = parseInt(raw.slice(4, 6), 16)
  const luminance = 0.2126 * r + 0.7152 * g + 0.0722 * b
  return luminance < 128
}

export function DrawingThemeSync() {
  const { theme } = useChart()
  const setChartIsDark = useDrawingStore((state) => state.setChartIsDark)

  useEffect(() => {
    // Our theme backgrounds are hex strings (see `getThemeColors`).
    setChartIsDark(isDarkHex(theme.backgroundColor))
  }, [setChartIsDark, theme.backgroundColor])

  return null
}

