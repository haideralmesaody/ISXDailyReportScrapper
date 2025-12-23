/**
 * Unit tests for tooltip helper functions
 */

import {
  formatIQD,
  formatVolume,
  formatNumber,
  calculatePercentageChange,
  formatPercentage,
  formatDate,
  calculateTradingValue,
  calculateComparison,
  calculateSMA,
  buildEnhancedTooltipData,
} from '@/lib/utils/tooltip-helpers'

describe('Tooltip Helper Functions', () => {
  describe('formatIQD', () => {
    it('should format numbers as IQD currency', () => {
      expect(formatIQD(1000)).toBe('IQD 1,000.00')
      expect(formatIQD(1234.56)).toBe('IQD 1,234.56')
      expect(formatIQD(0)).toBe('IQD 0.00')
    })

    it('should handle null and undefined values', () => {
      expect(formatIQD(null)).toBe('N/A')
      expect(formatIQD(undefined)).toBe('N/A')
    })
  })

  describe('formatVolume', () => {
    it('should format large numbers with thousands separator', () => {
      expect(formatVolume(1000000)).toBe('1,000,000')
      expect(formatVolume(1234567)).toBe('1,234,567')
      expect(formatVolume(100)).toBe('100')
    })

    it('should handle null and undefined values', () => {
      expect(formatVolume(null)).toBe('N/A')
      expect(formatVolume(undefined)).toBe('N/A')
    })
  })

  describe('formatNumber', () => {
    it('should format numbers with specified decimal places', () => {
      expect(formatNumber(1.234567, 2)).toBe('1.23')
      expect(formatNumber(1.234567, 4)).toBe('1.2346')
      expect(formatNumber(10, 2)).toBe('10.00')
    })

    it('should handle null and undefined values', () => {
      expect(formatNumber(null)).toBe('N/A')
      expect(formatNumber(undefined)).toBe('N/A')
    })
  })

  describe('calculatePercentageChange', () => {
    it('should calculate percentage change correctly', () => {
      expect(calculatePercentageChange(110, 100)).toBe(10)
      expect(calculatePercentageChange(90, 100)).toBe(-10)
      expect(calculatePercentageChange(100, 100)).toBe(0)
    })

    it('should handle zero or invalid previous values', () => {
      expect(calculatePercentageChange(100, 0)).toBe(0)
      expect(calculatePercentageChange(100, null as any)).toBe(0)
    })
  })

  describe('formatPercentage', () => {
    it('should format positive percentages', () => {
      const result = formatPercentage(5.5)
      expect(result.text).toBe('+5.50%')
      expect(result.color).toBe('text-green-600')
      expect(result.isPositive).toBe(true)
    })

    it('should format negative percentages', () => {
      const result = formatPercentage(-3.25)
      expect(result.text).toBe('-3.25%')
      expect(result.color).toBe('text-red-600')
      expect(result.isPositive).toBe(false)
    })

    it('should format zero percentage', () => {
      const result = formatPercentage(0)
      expect(result.text).toBe('+0.00%')
      expect(result.color).toBe('text-gray-500')
      expect(result.isPositive).toBe(false)
    })

    it('should handle null and undefined values', () => {
      const result = formatPercentage(null)
      expect(result.text).toBe('N/A')
      expect(result.color).toBe('text-muted-foreground')
      expect(result.isPositive).toBe(false)
    })
  })

  describe('formatDate', () => {
    it('should format date strings correctly', () => {
      expect(formatDate('2024-01-15')).toContain('Jan')
      expect(formatDate('2024-01-15')).toContain('15')
      expect(formatDate('2024-01-15')).toContain('2024')
    })

    it('should format Date objects correctly', () => {
      const date = new Date('2024-01-15')
      expect(formatDate(date)).toContain('Jan')
      expect(formatDate(date)).toContain('15')
      expect(formatDate(date)).toContain('2024')
    })
  })

  describe('calculateTradingValue', () => {
    it('should calculate trading value correctly', () => {
      expect(calculateTradingValue(100, 1000)).toBe(100000)
      expect(calculateTradingValue(50.5, 2000)).toBe(101000)
      expect(calculateTradingValue(0, 1000)).toBe(0)
    })
  })

  describe('calculateComparison', () => {
    it('should calculate comparison metrics correctly', () => {
      const current = { close: 110, volume: 1000, date: '2024-01-15' }
      const previous = { close: 100, volume: 900, date: '2024-01-14' }
      
      const result = calculateComparison(current, previous)
      
      expect(result).not.toBeNull()
      expect(result?.price).toBe(100)
      expect(result?.change).toBe(10)
      expect(result?.changePercent).toBe(10)
      expect(result?.volume).toBe(900)
      expect(result?.date).toBe('2024-01-14')
    })

    it('should return null for undefined previous data', () => {
      const current = { close: 110, volume: 1000, date: '2024-01-15' }
      const result = calculateComparison(current, undefined)
      expect(result).toBeNull()
    })
  })

  describe('calculateSMA', () => {
    it('should calculate simple moving average correctly', () => {
      const values = [10, 20, 30, 40, 50]
      expect(calculateSMA(values, 3)).toBe(40) // (30 + 40 + 50) / 3
      expect(calculateSMA(values, 5)).toBe(30) // (10 + 20 + 30 + 40 + 50) / 5
    })

    it('should return null for insufficient data', () => {
      const values = [10, 20]
      expect(calculateSMA(values, 3)).toBeNull()
    })
  })

  describe('buildEnhancedTooltipData', () => {
    const mockData = [
      { date: '2024-01-01', open: 100, high: 110, low: 95, close: 105, volume: 1000, value: 105000, trades: 50 },
      { date: '2024-01-02', open: 105, high: 115, low: 100, close: 110, volume: 1100, value: 121000, trades: 55 },
      { date: '2024-01-03', open: 110, high: 120, low: 105, close: 115, volume: 1200, value: 138000, trades: 60 },
      { date: '2024-01-04', open: 115, high: 125, low: 110, close: 120, volume: 1300, value: 156000, trades: 65 },
      { date: '2024-01-05', open: 120, high: 130, low: 115, close: 125, volume: 1400, value: 175000, trades: 70 },
    ]

    it('should build enhanced tooltip data with all metrics', () => {
      const currentData = mockData[4]
      const result = buildEnhancedTooltipData(currentData, mockData, 4)

      expect(result.date).toBe('2024-01-05')
      expect(result.open).toBe(120)
      expect(result.high).toBe(130)
      expect(result.low).toBe(115)
      expect(result.close).toBe(125)
      expect(result.volume).toBe(1400)
      expect(result.value).toBe(175000)
      expect(result.trades).toBe(70)
      
      // Daily change from previous day
      expect(result.change).toBe(5) // 125 - 120
      expect(result.changePercent).toBeCloseTo(4.17, 1) // (5/120) * 100
      
      // Yesterday comparison
      expect(result.vsYesterday).not.toBeNull()
      expect(result.vsYesterday?.price).toBe(120)
      
      // No week/month ago data for this small dataset
      expect(result.vsWeekAgo).toBeNull()
      expect(result.vsMonthAgo).toBeNull()
    })

    it('should handle first day data without previous comparisons', () => {
      const currentData = mockData[0]
      const result = buildEnhancedTooltipData(currentData, mockData, 0)

      expect(result.date).toBe('2024-01-01')
      expect(result.change).toBe(0)
      expect(result.changePercent).toBe(0)
      expect(result.vsYesterday).toBeNull()
      expect(result.vsWeekAgo).toBeNull()
      expect(result.vsMonthAgo).toBeNull()
    })
  })
})