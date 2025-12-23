/**
 * Foundation Indicators Test Suite
 *
 * Tests for Simple Moving Average (SMA), Exponential Moving Average (EMA),
 * and Relative Strength Index (RSI) - the foundation indicators that other
 * indicators depend on.
 *
 * These indicators are critical because:
 * - SMA is used in Bollinger Bands, MACD signal line
 * - EMA is used in MACD, many trend indicators
 * - RSI is used in trading strategies, overbought/oversold detection
 *
 * Coverage goals:
 * - Calculation accuracy (known datasets)
 * - Edge cases (empty data, insufficient data, zeros, negatives)
 * - Boundary conditions (period validation)
 * - Performance (large datasets)
 */

import { calculateSMA, calculateEMA, calculateRSI } from '@/lib/indicators/calculations'
import {
  SMA_TEST_DATA,
  EMA_TEST_DATA,
  RSI_TEST_DATA,
  EDGE_CASE_DATA,
  generateLargeDataset
} from './test-data'

describe('Simple Moving Average (SMA)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates SMA correctly for basic dataset', () => {
      const { input, period, expected } = SMA_TEST_DATA.simple
      const result = calculateSMA(input, period)

      expect(result).toHaveLength(expected.length)

      expected.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          expect(result[index]).toBeCloseTo(expectedValue, 2)
        }
      })
    })

    it('calculates SMA correctly for real price data', () => {
      const { input, period, expected } = SMA_TEST_DATA.priceData
      const result = calculateSMA(input, period)

      expect(result).toHaveLength(expected.length)

      expected.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          expect(result[index]).toBeCloseTo(expectedValue, 2)
        }
      })
    })

    it('returns correct number of null values before first SMA', () => {
      const data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
      const period = 5
      const result = calculateSMA(data, period)

      // First 4 values should be null (period - 1)
      for (let i = 0; i < period - 1; i++) {
        expect(result[i]).toBeNull()
      }

      // All subsequent values should be numbers
      for (let i = period - 1; i < result.length; i++) {
        expect(typeof result[i]).toBe('number')
      }
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateSMA(EDGE_CASE_DATA.emptyArray, 3)
      expect(result).toEqual([])
    })

    it('returns all nulls when period > data length', () => {
      const result = calculateSMA([1, 2, 3], 5)
      expect(result).toEqual([null, null, null])
    })

    it('handles single data point', () => {
      const result = calculateSMA([100], 1)
      expect(result[0]).toBe(100)
    })

    it('handles period of 1 (same as original data)', () => {
      const data = [10, 20, 30, 40, 50]
      const result = calculateSMA(data, 1)

      data.forEach((value, index) => {
        expect(result[index]).toBe(value)
      })
    })

    it('handles data with zeros', () => {
      const result = calculateSMA(EDGE_CASE_DATA.withZeros, 3)

      // Should not crash and should calculate correctly
      expect(result[2]).toBeCloseTo(1, 2) // (0+1+2)/3
      expect(result[3]).toBeCloseTo(2, 2) // (1+2+3)/3
    })

    it('handles negative values', () => {
      const result = calculateSMA(EDGE_CASE_DATA.withNegatives, 3)

      // Should calculate correctly with negatives
      expect(result[2]).toBeCloseTo(-3, 2) // (-5-3-1)/3
    })

    it('handles all same values', () => {
      const result = calculateSMA(EDGE_CASE_DATA.allSameValues, 3)

      // SMA of constant values should equal the constant
      for (let i = 2; i < result.length; i++) {
        expect(result[i]).toBe(100)
      }
    })
  })

  describe('Boundary Conditions', () => {
    it('handles period equal to data length', () => {
      const data = [1, 2, 3, 4, 5]
      const result = calculateSMA(data, 5)

      // First 4 should be null, last one is average of all
      expect(result[4]).toBe(3) // (1+2+3+4+5)/5
    })

    it('handles very small period', () => {
      const data = [10, 20, 30, 40, 50]
      const result = calculateSMA(data, 2)

      expect(result[0]).toBeNull()
      expect(result[1]).toBe(15) // (10+20)/2
      expect(result[2]).toBe(25) // (20+30)/2
    })

    it('handles very large values', () => {
      const result = calculateSMA(EDGE_CASE_DATA.largeValues, 2)

      // Should handle large numbers without overflow
      expect(result[1]).toBeCloseTo(1000050, 0)
    })

    it('handles very small values', () => {
      const result = calculateSMA(EDGE_CASE_DATA.smallValues, 2)

      // Should handle small numbers with precision
      expect(result[1]).toBeCloseTo(0.0015, 4)
    })
  })

  describe('Performance', () => {
    it('calculates SMA for large dataset in reasonable time', () => {
      const largeData = generateLargeDataset(10000)
      const startTime = performance.now()

      const result = calculateSMA(largeData, 50)

      const endTime = performance.now()
      const duration = endTime - startTime

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(100) // Should complete in < 100ms
    })
  })
})

// ============================================================================
// Exponential Moving Average (EMA) Tests
// ============================================================================

describe('Exponential Moving Average (EMA)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates EMA correctly for known dataset', () => {
      const { input, period, expected } = EMA_TEST_DATA.basic
      const result = calculateEMA(input, period)

      expect(result).toHaveLength(expected.length)

      expected.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          // EMA precision to 1 decimal place (industry standard, allows for rounding)
          expect(result[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('first EMA equals SMA of first period values', () => {
      const data = [22.27, 22.19, 22.08, 22.17, 22.18, 22.13]
      const period = 5
      const result = calculateEMA(data, period)

      // First EMA = SMA of first 5 values
      const firstSMA = (22.27 + 22.19 + 22.08 + 22.17 + 22.18) / 5
      expect(result[4]).toBeCloseTo(firstSMA, 2)
    })

    it('EMA reacts faster than SMA to price changes', () => {
      // Price spike scenario
      const data = [100, 100, 100, 100, 100, 150, 150, 150, 150, 150]
      const period = 5

      const smaResult = calculateSMA(data, period)
      const emaResult = calculateEMA(data, period)

      // After spike (index 6), EMA should be higher than SMA (reacts faster)
      const smaValue = smaResult[6]
      const emaValue = emaResult[6]

      expect(emaValue).toBeGreaterThan(smaValue!)
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateEMA(EDGE_CASE_DATA.emptyArray, 3)
      expect(result).toEqual([])
    })

    it('returns all nulls when period > data length', () => {
      const result = calculateEMA([1, 2, 3], 5)
      expect(result).toEqual([null, null, null])
    })

    it('handles period of 1', () => {
      const data = [10, 20, 30, 40, 50]
      const result = calculateEMA(data, 1)

      // Period 1 EMA should equal the close prices
      data.forEach((value, index) => {
        expect(result[index]).toBeCloseTo(value, 2)
      })
    })

    it('handles negative values', () => {
      const result = calculateEMA(EDGE_CASE_DATA.withNegatives, 3)

      // Should calculate correctly with negatives
      expect(result[2]).toBeDefined()
      expect(typeof result[2]).toBe('number')
    })
  })

  describe('Performance', () => {
    it('calculates EMA for large dataset in reasonable time', () => {
      const largeData = generateLargeDataset(10000)
      const startTime = performance.now()

      const result = calculateEMA(largeData, 50)

      const endTime = performance.now()
      const duration = endTime - startTime

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(100) // Should complete in < 100ms
    })
  })
})

// ============================================================================
// Relative Strength Index (RSI) Tests
// ============================================================================

describe('Relative Strength Index (RSI)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates RSI correctly for classic Investopedia example', () => {
      const { input, period, expected } = RSI_TEST_DATA.classic
      const result = calculateRSI(input, period)

      expect(result).toHaveLength(expected.length)

      expected.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          // RSI precision to 1 decimal place
          expect(result[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('returns RSI near 100 for strong uptrend', () => {
      const { input, period, expected } = RSI_TEST_DATA.uptrend
      const result = calculateRSI(input, period)

      // Last two values should be 100 (all gains, no losses)
      expect(result[14]).toBeCloseTo(100, 1)
      expect(result[15]).toBeCloseTo(100, 1)
    })

    it('returns RSI near 0 for strong downtrend', () => {
      const { input, period, expected } = RSI_TEST_DATA.downtrend
      const result = calculateRSI(input, period)

      // Last two values should be 0 (all losses, no gains)
      expect(result[14]).toBeCloseTo(0, 1)
      expect(result[15]).toBeCloseTo(0, 1)
    })

    it('returns null values before first valid RSI', () => {
      const data = Array(20).fill(0).map((_, i) => 100 + i)
      const period = 14
      const result = calculateRSI(data, period)

      // First 14 values should be null (need 14 price changes)
      for (let i = 0; i < period; i++) {
        expect(result[i]).toBeNull()
      }

      // 15th value onward should be valid RSI
      for (let i = period; i < result.length; i++) {
        expect(result[i]).not.toBeNull()
        expect(typeof result[i]).toBe('number')
      }
    })
  })

  describe('Boundary Conditions', () => {
    it('RSI values are always between 0 and 100', () => {
      // Random price data
      const data = generateLargeDataset(100)
      const result = calculateRSI(data, 14)

      result.forEach((value, index) => {
        if (value !== null) {
          expect(value).toBeGreaterThanOrEqual(0)
          expect(value).toBeLessThanOrEqual(100)
        }
      })
    })

    it('handles period equal to data length', () => {
      const data = Array(15).fill(0).map((_, i) => 100 + i * 2)
      const result = calculateRSI(data, 14)

      // Should have one valid RSI value at the end
      expect(result[14]).toBeDefined()
      expect(result[14]).not.toBeNull()
    })

    it('handles very small period', () => {
      const data = [100, 102, 101, 103, 102, 104]
      const result = calculateRSI(data, 2)

      // Should have RSI values starting from index 2
      expect(result[2]).toBeDefined()
      expect(result[2]).not.toBeNull()
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateRSI(EDGE_CASE_DATA.emptyArray, 14)
      expect(result).toEqual([])
    })

    it('returns all nulls when period >= data length', () => {
      const result = calculateRSI([1, 2, 3], 5)
      expect(result).toEqual([null, null, null])
    })

    it('handles constant prices (no change)', () => {
      const result = calculateRSI(EDGE_CASE_DATA.allSameValues, 3)

      // RSI should be undefined or 50 when no price changes
      // (Different implementations handle this differently)
      // Just ensure it doesn't crash
      expect(result).toHaveLength(EDGE_CASE_DATA.allSameValues.length)
    })

    it('handles alternating up/down prices', () => {
      const data = [100, 105, 100, 105, 100, 105, 100, 105, 100, 105,
                    100, 105, 100, 105, 100, 105]
      const result = calculateRSI(data, 14)

      // RSI should be around 50 for balanced up/down movement
      const lastRSI = result[result.length - 1]
      expect(lastRSI).toBeGreaterThan(40)
      expect(lastRSI).toBeLessThan(60)
    })
  })

  describe('Performance', () => {
    it('calculates RSI for large dataset in reasonable time', () => {
      const largeData = generateLargeDataset(10000)
      const startTime = performance.now()

      const result = calculateRSI(largeData, 14)

      const endTime = performance.now()
      const duration = endTime - startTime

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(150) // Should complete in < 150ms
    })
  })

  describe('Special Scenarios', () => {
    it('handles price gaps (large sudden changes)', () => {
      const data = [100, 100, 100, 100, 100, 200, 200, 200, 200, 200,
                    200, 200, 200, 200, 200, 200]
      const result = calculateRSI(data, 14)

      // After large gain, RSI should be very high
      expect(result[15]!).toBeGreaterThan(90)
    })

    it('handles recovery from oversold', () => {
      // Prices falling then recovering
      const data = [100, 95, 90, 85, 80, 75, 70, 65, 60, 55, 50, 45, 40, 35, 30,
                    32, 34, 36, 38, 40]
      const result = calculateRSI(data, 14)

      // RSI should be low in middle (oversold), then start rising
      const oversoldRSI = result[14]
      const recoveryRSI = result[result.length - 1]

      expect(oversoldRSI).toBeLessThan(30) // Oversold
      expect(recoveryRSI).toBeGreaterThan(oversoldRSI!) // Recovery started
    })
  })
})
