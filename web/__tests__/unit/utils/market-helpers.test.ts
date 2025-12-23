/**
 * Unit tests for market-helpers utility functions
 * Following CLAUDE.md testing requirements (80%+ coverage)
 */

import {
  getChangeColor,
  formatCurrency,
  formatPercent,
  calculateTradedValue,
  aggregateTickerData
} from '@/lib/utils/market-helpers'
import { TickerData } from '@/types/market'

describe('getChangeColor', () => {
  it('returns strong positive green for >= 5%', () => {
    expect(getChangeColor(5.0)).toBe('#10b981')
    expect(getChangeColor(10.0)).toBe('#10b981')
  })

  it('returns moderate positive green for 2-5%', () => {
    expect(getChangeColor(2.0)).toBe('#34d399')
    expect(getChangeColor(4.99)).toBe('#34d399')
  })

  it('returns weak positive green for 0-2%', () => {
    expect(getChangeColor(0.01)).toBe('#6ee7b7')
    expect(getChangeColor(1.99)).toBe('#6ee7b7')
  })

  it('returns neutral gray for exactly 0%', () => {
    expect(getChangeColor(0)).toBe('#9ca3af')
  })

  it('returns weak negative red for -2 to 0%', () => {
    expect(getChangeColor(-0.01)).toBe('#fca5a5')
    expect(getChangeColor(-1.99)).toBe('#fca5a5')
  })

  it('returns moderate negative red for -5 to -2%', () => {
    expect(getChangeColor(-2.0)).toBe('#f87171')
    expect(getChangeColor(-4.99)).toBe('#f87171')
  })

  it('returns strong negative red for <= -5%', () => {
    expect(getChangeColor(-5.0)).toBe('#ef4444')
    expect(getChangeColor(-10.0)).toBe('#ef4444')
  })
})

describe('formatCurrency', () => {
  it('formats billions correctly', () => {
    expect(formatCurrency(1_500_000_000)).toBe('1.50B IQD')
    expect(formatCurrency(2_750_000_000)).toBe('2.75B IQD')
  })

  it('formats millions correctly', () => {
    expect(formatCurrency(1_250_000)).toBe('1.25M IQD')
    expect(formatCurrency(999_000_000)).toBe('999.00M IQD')
  })

  it('formats thousands correctly', () => {
    expect(formatCurrency(1_500)).toBe('1.50K IQD')
    expect(formatCurrency(999_999)).toBe('1000.00K IQD')
  })

  it('formats small values correctly', () => {
    expect(formatCurrency(500)).toBe('500.00 IQD')
    expect(formatCurrency(0)).toBe('0.00 IQD')
  })

  it('handles decimal precision', () => {
    expect(formatCurrency(1_234_567)).toBe('1.23M IQD')
    expect(formatCurrency(1_234)).toBe('1.23K IQD')
  })
})

describe('formatPercent', () => {
  it('formats positive percentages with + sign', () => {
    expect(formatPercent(5.25)).toBe('+5.25%')
    expect(formatPercent(0.01)).toBe('+0.01%')
  })

  it('formats negative percentages with - sign', () => {
    expect(formatPercent(-3.50)).toBe('-3.50%')
    expect(formatPercent(-0.01)).toBe('-0.01%')
  })

  it('formats zero with + sign', () => {
    expect(formatPercent(0)).toBe('+0.00%')
  })

  it('always shows 2 decimal places', () => {
    expect(formatPercent(5)).toBe('+5.00%')
    expect(formatPercent(-10)).toBe('-10.00%')
  })
})

describe('calculateTradedValue', () => {
  it('calculates correct traded value', () => {
    expect(calculateTradedValue(1.25, 1000000)).toBe(1250000)
    expect(calculateTradedValue(0.50, 2000000)).toBe(1000000)
  })

  it('handles zero price', () => {
    expect(calculateTradedValue(0, 1000000)).toBe(0)
  })

  it('handles zero volume', () => {
    expect(calculateTradedValue(1.25, 0)).toBe(0)
  })

  it('handles large numbers', () => {
    expect(calculateTradedValue(100, 10000000)).toBe(1000000000)
  })

  it('handles decimal prices', () => {
    expect(calculateTradedValue(0.75, 1000000)).toBe(750000)
  })
})

describe('aggregateTickerData', () => {
  const mockTickers: TickerData[] = [
    { symbol: 'BBOB', name: 'Bank of Baghdad', price: 1.00, change: 0.05, changePercent: 5.0, tradedValue: 1000000, volume: 1000000, trades: 100 },
    { symbol: 'BBOB', name: 'Bank of Baghdad', price: 1.10, change: 0.10, changePercent: 10.0, tradedValue: 1500000, volume: 1500000, trades: 150 },
    { symbol: 'HCPI', name: 'Al-Hillah Cement', price: 0.80, change: -0.05, changePercent: -5.88, tradedValue: 800000, volume: 1000000, trades: 50 }
  ]

  it('aggregates duplicate symbols correctly', () => {
    const result = aggregateTickerData(mockTickers, '2024-01-01', '2024-01-02')

    expect(result).toHaveLength(2) // BBOB + HCPI
    const bank = result.find(t => t.symbol === 'BBOB')
    expect(bank).toBeDefined()
    expect(bank!.price).toBe(1.05) // Average of 1.00 and 1.10
    expect(bank!.changePercent).toBe(7.5) // Average of 5.0 and 10.0
    expect(bank!.tradedValue).toBe(2500000) // Sum of 1000000 + 1500000
    expect(bank!.volume).toBe(2500000) // Sum
    expect(bank!.trades).toBe(250) // Sum
  })

  it('preserves ticker name from first occurrence', () => {
    const result = aggregateTickerData(mockTickers, '2024-01-01', '2024-01-02')
    const bank = result.find(t => t.symbol === 'BBOB')
    expect(bank!.name).toBe('Bank of Baghdad')
  })

  it('handles single ticker correctly', () => {
    const single = [mockTickers[2]]
    const result = aggregateTickerData(single, '2024-01-01', '2024-01-01')

    expect(result).toHaveLength(1)
    expect(result[0]).toEqual(single[0])
  })

  it('handles empty array', () => {
    const result = aggregateTickerData([], '2024-01-01', '2024-01-02')
    expect(result).toEqual([])
  })

  it('calculates averages for price and change fields', () => {
    const result = aggregateTickerData(mockTickers, '2024-01-01', '2024-01-02')
    const bank = result.find(t => t.symbol === 'BBOB')

    expect(bank!.change).toBeCloseTo(0.075, 5) // Average of 0.05 and 0.10 (floating point safe)
  })

  it('calculates sums for volume, value, and trades', () => {
    const result = aggregateTickerData(mockTickers, '2024-01-01', '2024-01-02')
    const bank = result.find(t => t.symbol === 'BBOB')

    expect(bank!.tradedValue).toBe(2500000) // Sum
    expect(bank!.volume).toBe(2500000) // Sum
    expect(bank!.trades).toBe(250) // Sum
  })
})
