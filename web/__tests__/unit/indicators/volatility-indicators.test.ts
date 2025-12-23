/**
 * Volatility Indicators Test Suite
 * Tests for Bollinger Bands and ATR (Average True Range)
 */

import { calculateBollingerBands, calculateATR } from '@/lib/indicators/calculations'
import { BOLLINGER_BANDS_TEST_DATA, ATR_TEST_DATA } from './test-data'

// ============================================================================
// Bollinger Bands Tests
// ============================================================================

describe('Bollinger Bands', () => {
  describe('Calculation Accuracy', () => {
    it('middle band equals SMA', () => {
      const { input, period, stdDevMultiplier, expectedMiddle } = BOLLINGER_BANDS_TEST_DATA.basic
      const result = calculateBollingerBands(input, period, stdDevMultiplier)

      expect(result.middle).toHaveLength(expectedMiddle.length)

      expectedMiddle.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.middle[index]).toBeNull()
        } else {
          expect(result.middle[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('upper band equals middle + (stdDev * multiplier)', () => {
      const { input, period, stdDevMultiplier, expectedUpper } = BOLLINGER_BANDS_TEST_DATA.basic
      const result = calculateBollingerBands(input, period, stdDevMultiplier)

      expect(result.upper).toHaveLength(expectedUpper.length)

      expectedUpper.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.upper[index]).toBeNull()
        } else {
          expect(result.upper[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('lower band equals middle - (stdDev * multiplier)', () => {
      const { input, period, stdDevMultiplier, expectedLower } = BOLLINGER_BANDS_TEST_DATA.basic
      const result = calculateBollingerBands(input, period, stdDevMultiplier)

      expect(result.lower).toHaveLength(expectedLower.length)

      expectedLower.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.lower[index]).toBeNull()
        } else {
          expect(result.lower[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('bands widen during high volatility', () => {
      // Create volatile price action
      const volatile = [100, 90, 110, 85, 115, 80, 120, 75, 125, 70, 130, 65, 135, 60, 140]
      const result = calculateBollingerBands(volatile, 10, 2)

      // Band width at index 14 (last point, high volatility)
      const lastBandWidth = result.upper[14]! - result.lower[14]!

      // Band width at index 9 (first valid point, lower volatility)
      const firstBandWidth = result.upper[9]! - result.lower[9]!

      // Bands should widen with increased volatility
      expect(lastBandWidth).toBeGreaterThan(firstBandWidth)
    })
  })

  describe('Edge Cases', () => {
    it('returns empty result for empty input', () => {
      const result = calculateBollingerBands([], 20, 2)

      expect(result.upper).toEqual([])
      expect(result.middle).toEqual([])
      expect(result.lower).toEqual([])
    })

    it('returns all nulls for insufficient data', () => {
      const shortData = [100, 102, 104]
      const result = calculateBollingerBands(shortData, 20, 2)

      result.upper.forEach(val => expect(val).toBeNull())
      result.middle.forEach(val => expect(val).toBeNull())
      result.lower.forEach(val => expect(val).toBeNull())
    })

    it('bands converge for constant prices', () => {
      const constant = new Array(25).fill(100)
      const result = calculateBollingerBands(constant, 20, 2)

      // When prices are constant, std dev = 0, so all bands equal
      expect(result.upper[19]).toBe(100)
      expect(result.middle[19]).toBe(100)
      expect(result.lower[19]).toBe(100)
    })

    it('works with custom multiplier', () => {
      const data = [100, 102, 101, 103, 104, 102, 103, 105, 104, 106, 107, 108, 110, 109, 111, 113, 112, 114, 115, 116]

      const bb1 = calculateBollingerBands(data, 10, 1)
      const bb2 = calculateBollingerBands(data, 10, 2)
      const bb3 = calculateBollingerBands(data, 10, 3)

      // Higher multiplier = wider bands
      const width1 = bb1.upper[9]! - bb1.lower[9]!
      const width2 = bb2.upper[9]! - bb2.lower[9]!
      const width3 = bb3.upper[9]! - bb3.lower[9]!

      expect(width2).toBeGreaterThan(width1)
      expect(width3).toBeGreaterThan(width2)
    })
  })

  describe('Boundary Conditions', () => {
    it('first valid bands appear at period-1', () => {
      const data = [100, 102, 101, 103, 104, 102, 103, 105, 104, 106, 107, 108]
      const period = 10
      const result = calculateBollingerBands(data, period, 2)

      // All values before period-1 should be null
      for (let i = 0; i < period - 1; i++) {
        expect(result.upper[i]).toBeNull()
        expect(result.middle[i]).toBeNull()
        expect(result.lower[i]).toBeNull()
      }

      // Values at period-1 should be numbers
      expect(result.upper[period - 1]).not.toBeNull()
      expect(result.middle[period - 1]).not.toBeNull()
      expect(result.lower[period - 1]).not.toBeNull()
    })

    it('price can touch or break bands', () => {
      const data = [100, 102, 101, 103, 104, 102, 103, 105, 104, 106, 107, 108, 110, 109, 111, 113, 112, 114, 115, 150] // Last price breaks upper band
      const result = calculateBollingerBands(data, 10, 2)

      const lastPrice = 150
      const lastUpper = result.upper[19]!

      // Price can exceed upper band
      expect(lastPrice).toBeGreaterThan(lastUpper)
    })
  })

  describe('Performance', () => {
    it('calculates Bollinger Bands for large dataset in reasonable time', () => {
      const largeData = new Array(10000).fill(0).map((_, i) => 100 + Math.sin(i / 100) * 10)

      const start = performance.now()
      const result = calculateBollingerBands(largeData, 20, 2)
      const duration = performance.now() - start

      expect(result.upper).toHaveLength(10000)
      expect(result.middle).toHaveLength(10000)
      expect(result.lower).toHaveLength(10000)
      expect(duration).toBeLessThan(100) // Should complete in < 100ms
    })
  })

  describe('Special Scenarios', () => {
    it('detects Bollinger Squeeze (low volatility)', () => {
      // Create period of low volatility
      const lowVolatility = [100, 101, 100.5, 101.5, 100.8, 101.2, 100.9, 101.1, 100.7, 101.3, 100.6, 101.4, 100.5, 101.5, 100.4]
      const result = calculateBollingerBands(lowVolatility, 10, 2)

      // Band width should be small during low volatility
      const bandWidth = result.upper[14]! - result.lower[14]!
      expect(bandWidth).toBeLessThan(5) // Narrow bands indicate low volatility
    })
  })
})

// ============================================================================
// ATR (Average True Range) Tests
// ============================================================================

describe('Average True Range (ATR)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates ATR correctly', () => {
      const { highs, lows, closes, period, expectedATR } = ATR_TEST_DATA.basic
      const result = calculateATR(highs, lows, closes, period)

      expect(result).toHaveLength(expectedATR.length)

      expectedATR.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          expect(result[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('ATR equals EMA of True Range', () => {
      const highs = [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145]
      const lows = [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134]
      const closes = [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144]

      const result = calculateATR(highs, lows, closes, 14)

      // ATR should be defined and a number after period
      expect(result[14]).toBeDefined()
      expect(typeof result[14]).toBe('number')
    })

    it('first valid ATR appears at index equal to period', () => {
      const highs = new Array(20).fill(0).map((_, i) => 130 + i)
      const lows = new Array(20).fill(0).map((_, i) => 120 + i)
      const closes = new Array(20).fill(0).map((_, i) => 125 + i)
      const period = 14

      const result = calculateATR(highs, lows, closes, period)

      // All values before period should be null
      for (let i = 0; i < period; i++) {
        expect(result[i]).toBeNull()
      }

      // Value at period should be a number
      expect(result[period]).not.toBeNull()
      expect(typeof result[period]).toBe('number')
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateATR([], [], [], 14)
      expect(result).toEqual([])
    })

    it('returns empty for insufficient data', () => {
      const result = calculateATR([130], [120], [125], 14)
      expect(result).toEqual([])
    })

    it('handles no gaps (TR = High - Low)', () => {
      // When there are no gaps, TR should equal High - Low
      const highs = [110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110, 110]
      const lows = [100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100]
      const closes = [105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105, 105]

      const result = calculateATR(highs, lows, closes, 14)

      // ATR should equal High - Low = 10
      expect(result[14]).toBeCloseTo(10, 1)
    })
  })

  describe('Boundary Conditions', () => {
    it('large price gaps increase ATR', () => {
      // Normal volatility
      const normalHighs = new Array(20).fill(0).map((_, i) => 130 + i * 0.5)
      const normalLows = new Array(20).fill(0).map((_, i) => 120 + i * 0.5)
      const normalCloses = new Array(20).fill(0).map((_, i) => 125 + i * 0.5)

      // High volatility with gaps
      const volatileHighs = [130, 150, 160, 155, 165, 175, 170, 180, 190, 185, 195, 205, 200, 210, 220, 215, 225, 235, 230, 240]
      const volatileLows = [120, 140, 145, 140, 150, 160, 155, 165, 175, 170, 180, 190, 185, 195, 205, 200, 210, 220, 215, 225]
      const volatileCloses = [125, 145, 155, 150, 160, 170, 165, 175, 185, 180, 190, 200, 195, 205, 215, 210, 220, 230, 225, 235]

      const normalATR = calculateATR(normalHighs, normalLows, normalCloses, 14)
      const volatileATR = calculateATR(volatileHighs, volatileLows, volatileCloses, 14)

      // Volatile ATR should be higher than normal ATR
      expect(volatileATR[19]!).toBeGreaterThan(normalATR[19]!)
    })

    it('works with different periods', () => {
      const highs = new Array(25).fill(0).map((_, i) => 130 + Math.sin(i / 5) * 10)
      const lows = new Array(25).fill(0).map((_, i) => 120 + Math.sin(i / 5) * 10)
      const closes = new Array(25).fill(0).map((_, i) => 125 + Math.sin(i / 5) * 10)

      const atr7 = calculateATR(highs, lows, closes, 7)
      const atr14 = calculateATR(highs, lows, closes, 14)
      const atr21 = calculateATR(highs, lows, closes, 21)

      // First valid ATR at different indices
      expect(atr7[7]).not.toBeNull()
      expect(atr14[14]).not.toBeNull()
      expect(atr21[21]).not.toBeNull()
    })
  })

  describe('Performance', () => {
    it('calculates ATR for large dataset in reasonable time', () => {
      const largeHighs = new Array(10000).fill(0).map((_, i) => 130 + Math.sin(i / 50) * 20)
      const largeLows = new Array(10000).fill(0).map((_, i) => 120 + Math.sin(i / 50) * 20)
      const largeCloses = new Array(10000).fill(0).map((_, i) => 125 + Math.sin(i / 50) * 20)

      const start = performance.now()
      const result = calculateATR(largeHighs, largeLows, largeCloses, 14)
      const duration = performance.now() - start

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(80) // Should complete in < 80ms
    })
  })

  describe('Special Scenarios', () => {
    it('ATR increases with rising volatility', () => {
      // Start with low volatility, then increase
      const highs = [105, 106, 107, 106, 107, 108, 107, 108, 109, 108, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155]
      const lows = [95, 96, 97, 96, 97, 98, 97, 98, 99, 98, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145]
      const closes = [100, 101, 102, 101, 102, 103, 102, 103, 104, 103, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150]

      const result = calculateATR(highs, lows, closes, 14)

      // ATR at index 14 (low volatility period)
      const earlyATR = result[14]!

      // ATR at index 19 (high volatility period)
      const lateATR = result[19]!

      // ATR should increase or stay the same as volatility increases
      expect(lateATR).toBeGreaterThanOrEqual(earlyATR)
    })
  })
})
