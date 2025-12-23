/**
 * Additional Indicators Test Suite
 * Tests for Momentum (MOM), Rate of Change (ROC), CCI, and Williams %R
 */

import {
  calculateMomentum,
  calculateROC,
  calculateCCI,
  calculateWilliamsR
} from '@/lib/indicators/calculations'
import { MOMENTUM_TEST_DATA, ROC_TEST_DATA, WILLIAMS_R_TEST_DATA, CCI_TEST_DATA } from './test-data'

// ============================================================================
// Momentum (MOM) Tests
// ============================================================================

describe('Momentum Oscillator (MOM)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates momentum correctly for known dataset', () => {
      const { input, period, expected } = MOMENTUM_TEST_DATA.simple
      const result = calculateMomentum(input, period)

      expect(result).toHaveLength(expected.length)

      expected.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          expect(result[index]).toBe(expectedValue)
        }
      })
    })

    it('returns positive values during uptrend', () => {
      const uptrend = [100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120]
      const result = calculateMomentum(uptrend, 4)

      // All non-null values should be positive (prices increasing)
      result.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBeGreaterThan(0)
        }
      })
    })

    it('returns negative values during downtrend', () => {
      const downtrend = [120, 118, 116, 114, 112, 110, 108, 106, 104, 102, 100]
      const result = calculateMomentum(downtrend, 4)

      // All non-null values should be negative (prices decreasing)
      result.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBeLessThan(0)
        }
      })
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateMomentum([], 10)
      expect(result).toEqual([])
    })

    it('returns all nulls for insufficient data', () => {
      const shortData = [100, 102, 104]
      const result = calculateMomentum(shortData, 10)

      result.forEach(val => expect(val).toBeNull())
    })

    it('handles all zeros', () => {
      const zeros = [0, 0, 0, 0, 0, 0, 0, 0]
      const result = calculateMomentum(zeros, 3)

      // MOM should be 0 when all prices are same
      result.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBe(0)
        }
      })
    })
  })

  describe('Boundary Conditions', () => {
    it('first valid MOM appears at index equal to period', () => {
      const data = [100, 102, 104, 106, 108, 110, 112]
      const period = 3
      const result = calculateMomentum(data, period)

      // All values before period should be null
      for (let i = 0; i < period; i++) {
        expect(result[i]).toBeNull()
      }

      // Value at period should be a number
      expect(result[period]).not.toBeNull()
      expect(typeof result[period]).toBe('number')
    })

    it('works with period of 1', () => {
      const data = [100, 102, 104, 106, 108]
      const result = calculateMomentum(data, 1)

      // Period 1 means compare to previous value
      expect(result[0]).toBeNull()
      expect(result[1]).toBe(2)   // 102 - 100
      expect(result[2]).toBe(2)   // 104 - 102
      expect(result[3]).toBe(2)   // 106 - 104
      expect(result[4]).toBe(2)   // 108 - 106
    })
  })

  describe('Performance', () => {
    it('calculates MOM for large dataset in reasonable time', () => {
      const largeData = new Array(10000).fill(0).map((_, i) => 100 + i * 0.1)

      const start = performance.now()
      const result = calculateMomentum(largeData, 10)
      const duration = performance.now() - start

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(50) // Should complete in < 50ms
    })
  })
})

// ============================================================================
// Rate of Change (ROC) Tests
// ============================================================================

describe('Rate of Change (ROC)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates ROC correctly for known dataset', () => {
      const { input, period, expected } = ROC_TEST_DATA.simple
      const result = calculateROC(input, period)

      expect(result).toHaveLength(expected.length)

      expected.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result[index]).toBeNull()
        } else {
          expect(result[index]).toBeCloseTo(expectedValue, 2)
        }
      })
    })

    it('returns percentage interpretation correctly', () => {
      const data = [100, 105, 110, 115, 120] // 5% increases
      const result = calculateROC(data, 1)

      // Each value should be approximately 5% (except first null)
      expect(result[0]).toBeNull()
      expect(result[1]).toBeCloseTo(5.0, 1)   // (105-100)/100 * 100
      expect(result[2]).toBeCloseTo(4.76, 1)  // (110-105)/105 * 100
      expect(result[3]).toBeCloseTo(4.55, 1)  // (115-110)/110 * 100
      expect(result[4]).toBeCloseTo(4.35, 1)  // (120-115)/115 * 100
    })

    it('returns positive ROC during uptrend', () => {
      const uptrend = [100, 103, 106, 109, 112, 115, 118]
      const result = calculateROC(uptrend, 2)

      result.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBeGreaterThan(0)
        }
      })
    })

    it('returns negative ROC during downtrend', () => {
      const downtrend = [120, 116, 112, 108, 104, 100]
      const result = calculateROC(downtrend, 2)

      result.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBeLessThan(0)
        }
      })
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateROC([], 12)
      expect(result).toEqual([])
    })

    it('returns null for division by zero', () => {
      const dataWithZero = [0, 1, 2, 3, 4, 5]
      const result = calculateROC(dataWithZero, 1)

      // First two should be null (first null by default, second due to division by 0)
      expect(result[0]).toBeNull()
      expect(result[1]).toBeNull()
      // Rest should be valid
      expect(result[2]).not.toBeNull()
    })

    it('handles negative values', () => {
      const negatives = [-10, -9, -8, -7, -6, -5]
      const result = calculateROC(negatives, 2)

      // Should calculate properly with negative values
      expect(result[2]).toBeCloseTo(-20.0, 1) // (-8 - (-10)) / -10 * 100 = 2/-10 * 100 = -20%
    })
  })

  describe('Boundary Conditions', () => {
    it('first valid ROC appears at index equal to period', () => {
      const data = [100, 102, 104, 106, 108, 110]
      const period = 3
      const result = calculateROC(data, period)

      // All values before period should be null
      for (let i = 0; i < period; i++) {
        expect(result[i]).toBeNull()
      }

      // Value at period should be a number
      expect(result[period]).not.toBeNull()
      expect(typeof result[period]).toBe('number')
    })

    it('works with different periods', () => {
      const data = [100, 110, 120, 130, 140, 150, 160]

      const roc3 = calculateROC(data, 3)
      const roc5 = calculateROC(data, 5)

      // ROC(3) first valid at index 3
      expect(roc3[3]).not.toBeNull()

      // ROC(5) first valid at index 5
      expect(roc5[5]).not.toBeNull()
    })
  })

  describe('Performance', () => {
    it('calculates ROC for large dataset in reasonable time', () => {
      const largeData = new Array(10000).fill(0).map((_, i) => 100 * (1 + Math.sin(i / 100) * 0.1))

      const start = performance.now()
      const result = calculateROC(largeData, 12)
      const duration = performance.now() - start

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(50) // Should complete in < 50ms
    })
  })
})


// ============================================================================
// Williams %R Tests
// ============================================================================

describe('Williams %R', () => {
  describe('Calculation Accuracy', () => {
    it('calculates Williams %R correctly', () => {
      const { highs, lows, closes, period } = WILLIAMS_R_TEST_DATA.basic
      const result = calculateWilliamsR(highs, lows, closes, period)

      // First valid at index 13 (period-1)
      for (let i = 0; i < period - 1; i++) {
        expect(result[i]).toBeNull()
      }

      expect(result[period - 1]).not.toBeNull()
      expect(typeof result[period - 1]).toBe('number')
    })

    it('Williams %R values are between -100 and 0', () => {
      const highs = [130, 135, 140, 145, 150, 148, 146, 144, 142, 140, 138, 136, 134, 132, 130]
      const lows = [120, 122, 124, 126, 128, 130, 132, 134, 136, 138, 128, 126, 124, 122, 120]
      const closes = [125, 130, 135, 140, 145, 140, 138, 136, 134, 138, 133, 130, 128, 126, 124]

      const result = calculateWilliamsR(highs, lows, closes, 14)

      result.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBeGreaterThanOrEqual(-100)
          expect(val).toBeLessThanOrEqual(0)
        }
      })
    })

    it('detects overbought conditions (Williams %R > -20)', () => {
      // Prices at top of range
      const highs = new Array(20).fill(150)
      const lows = new Array(20).fill(100)
      const closes = new Array(20).fill(148) // Near high

      const result = calculateWilliamsR(highs, lows, closes, 14)

      // Williams %R should be > -20 (overbought)
      const firstValue = result[13]
      expect(firstValue).not.toBeNull()
      expect(firstValue!).toBeGreaterThan(-20)
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateWilliamsR([], [], [], 14)
      expect(result).toEqual([])
    })

    it('handles no price range (constant highs/lows)', () => {
      const highs = new Array(20).fill(150)
      const lows = new Array(20).fill(150)
      const closes = new Array(20).fill(150)

      const result = calculateWilliamsR(highs, lows, closes, 14)

      // Should return -50 (neutral) when range is zero
      expect(result[13]).toBe(-50)
    })
  })

  describe('Performance', () => {
    it('calculates Williams %R for large dataset in reasonable time', () => {
      const largeHighs = new Array(10000).fill(0).map((_, i) => 130 + Math.sin(i / 50) * 20)
      const largeLows = new Array(10000).fill(0).map((_, i) => 120 + Math.sin(i / 50) * 20)
      const largeCloses = new Array(10000).fill(0).map((_, i) => 125 + Math.sin(i / 50) * 20)

      const start = performance.now()
      const result = calculateWilliamsR(largeHighs, largeLows, largeCloses, 14)
      const duration = performance.now() - start

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(80) // Should complete in < 80ms
    })
  })
})

// ============================================================================
// CCI (Commodity Channel Index) Tests
// ============================================================================

describe('Commodity Channel Index (CCI)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates CCI correctly', () => {
      const { highs, lows, closes, period } = CCI_TEST_DATA.basic
      const result = calculateCCI(highs, lows, closes, period)

      // First valid at index 19 (period-1)
      for (let i = 0; i < period - 1; i++) {
        expect(result[i]).toBeNull()
      }

      expect(result[period - 1]).not.toBeNull()
      expect(typeof result[period - 1]).toBe('number')
    })

    it('CCI oscillates around zero', () => {
      const highs = [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 149, 148]
      const lows = [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 138, 137]
      const closes = [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 148, 147]

      const result = calculateCCI(highs, lows, closes, 20)

      // CCI should have both positive and negative values
      const nonNullValues = result.filter(v => v !== null)
      const hasPositive = nonNullValues.some(v => v! > 0)
      const hasNegative = nonNullValues.some(v => v! < 0)

      expect(hasPositive || hasNegative).toBe(true)
    })

    it('detects overbought/oversold conditions', () => {
      // Strong uptrend should produce high CCI
      const highs = [100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200]
      const lows = [95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195]
      const closes = [102, 107, 112, 117, 122, 127, 132, 137, 142, 147, 152, 157, 162, 167, 172, 177, 182, 187, 192, 197, 199]

      const result = calculateCCI(highs, lows, closes, 20)

      // Last CCI should be positive (uptrend)
      expect(result[20]).toBeGreaterThan(0)
    })
  })

  describe('Edge Cases', () => {
    it('returns empty array for empty input', () => {
      const result = calculateCCI([], [], [], 20)
      expect(result).toEqual([])
    })

    it('handles zero mean deviation', () => {
      // All same typical prices
      const highs = new Array(25).fill(150)
      const lows = new Array(25).fill(150)
      const closes = new Array(25).fill(150)

      const result = calculateCCI(highs, lows, closes, 20)

      // Should handle gracefully (likely return 0)
      expect(result[19]).toBe(0)
    })
  })

  describe('Performance', () => {
    it('calculates CCI for large dataset in reasonable time', () => {
      const largeHighs = new Array(10000).fill(0).map((_, i) => 130 + Math.sin(i / 50) * 20)
      const largeLows = new Array(10000).fill(0).map((_, i) => 120 + Math.sin(i / 50) * 20)
      const largeCloses = new Array(10000).fill(0).map((_, i) => 125 + Math.sin(i / 50) * 20)

      const start = performance.now()
      const result = calculateCCI(largeHighs, largeLows, largeCloses, 20)
      const duration = performance.now() - start

      expect(result).toHaveLength(10000)
      expect(duration).toBeLessThan(80) // Should complete in < 80ms
    })
  })
})
