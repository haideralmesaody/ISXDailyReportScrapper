import { processChartData } from '@/lib/utils/chart-data-processor'
import { buildEnhancedTooltipData } from '@/lib/utils/tooltip-helpers'
import type { TickerHistoricalData } from '@/types/analysis'

describe('analysis chart basics', () => {
  const sample: TickerHistoricalData[] = [
    {
      date: '2025-01-08',
      open: 4.31,
      high: 4.37,
      low: 4.31,
      close: 4.36,
      volume: 28781468,
      value: 125012840,
      trades: 95,
      change: 0.04,
      changePercent: 0.93,
    },
    {
      date: '2025-01-09',
      open: 4.37,
      high: 4.4,
      low: 4.35,
      close: 4.36,
      volume: 27036646,
      value: 118391172,
      trades: 79,
      change: 0,
      changePercent: 0,
    },
  ]

  it('processChartData includes volume series values', () => {
    const processed = processChartData(sample, 'candlestick', 'MAX')
    expect(processed.candlestickData).toHaveLength(2)
    expect(processed.volumeData).toHaveLength(2)

    expect(processed.volumeData[0]?.value).toBe(28781468)
    expect(processed.volumeData[1]?.value).toBe(27036646)
  })

  it('tooltip builder preserves OHLCV values', () => {
    const data = buildEnhancedTooltipData(sample[1], sample, 1)
    expect(data.open).toBe(4.37)
    expect(data.close).toBe(4.36)
    expect(data.volume).toBe(27036646)
  })
})

