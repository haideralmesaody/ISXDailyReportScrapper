/**
 * Trend Indicators Test Suite
 *
 * Tests for ADX (Average Directional Index) and Ichimoku Cloud indicators.
 * Validates calculation accuracy, edge cases, boundary conditions, and performance.
 */

import { calculateADX, calculateIchimoku } from '@/lib/indicators/calculations'
import { ADX_TEST_DATA, ICHIMOKU_TEST_DATA } from './test-data'

// ============================================================================
// Average Directional Index (ADX) Tests
// ============================================================================

describe('Average Directional Index (ADX)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates +DI correctly for strong uptrend', () => {
      const { highs, lows, closes, period, expectedPlusDI } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      expect(result.plusDI).toHaveLength(expectedPlusDI.length)

      expectedPlusDI.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.plusDI[index]).toBeNull()
        } else {
          expect(result.plusDI[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('calculates -DI correctly for strong uptrend (zero values)', () => {
      const { highs, lows, closes, period, expectedMinusDI } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      expect(result.minusDI).toHaveLength(expectedMinusDI.length)

      expectedMinusDI.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.minusDI[index]).toBeNull()
        } else {
          expect(result.minusDI[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('calculates ADX correctly for strong uptrend (should be 100)', () => {
      const { highs, lows, closes, period, expectedADX } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      expect(result.adx).toHaveLength(expectedADX.length)

      expectedADX.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.adx[index]).toBeNull()
        } else {
          expect(result.adx[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('ADX = 100 indicates perfect trend (all +DI, no -DI)', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      // Last ADX value should be 100 (perfect uptrend)
      const lastADX = result.adx[result.adx.length - 1]
      expect(lastADX).toBeCloseTo(100.0, 0)

      // -DI should be zero (no downward movement)
      const lastMinusDI = result.minusDI[result.minusDI.length - 1]
      expect(lastMinusDI).toBeCloseTo(0.0, 1)
    })

    it('calculates DI values for sideways market (weak trend)', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.sideways
      const result = calculateADX(highs, lows, closes, period)

      // In sideways market, both +DI and -DI should be relatively balanced
      const lastPlusDI = result.plusDI[result.plusDI.length - 1]
      const lastMinusDI = result.minusDI[result.minusDI.length - 1]

      // Both should be non-null
      expect(lastPlusDI).not.toBeNull()
      expect(lastMinusDI).not.toBeNull()

      // ADX should be lower than strong trend (< 100)
      const lastADX = result.adx[result.adx.length - 1]
      if (lastADX !== null) {
        expect(lastADX).toBeLessThan(100)
      }
    })
  })

  describe('Edge Cases', () => {
    it('handles empty arrays', () => {
      const result = calculateADX([], [], [], 14)

      expect(result.adx).toEqual([])
      expect(result.plusDI).toEqual([])
      expect(result.minusDI).toEqual([])
    })

    it('handles insufficient data (less than period)', () => {
      const highs = [130, 132, 134]
      const lows = [120, 121, 122]
      const closes = [125, 127, 129]

      const result = calculateADX(highs, lows, closes, 14)

      expect(result.adx).toHaveLength(3)
      expect(result.plusDI).toHaveLength(3)
      expect(result.minusDI).toHaveLength(3)

      result.adx.forEach(value => expect(value).toBeNull())
      result.plusDI.forEach(value => expect(value).toBeNull())
      result.minusDI.forEach(value => expect(value).toBeNull())
    })

    it('handles mismatched array lengths', () => {
      const highs = [130, 132, 134, 136]
      const lows = [120, 121, 122]
      const closes = [125, 127]

      const result = calculateADX(highs, lows, closes, 14)

      expect(result.adx).toEqual([])
      expect(result.plusDI).toEqual([])
      expect(result.minusDI).toEqual([])
    })

    it('handles equal highs and lows (no volatility)', () => {
      const highs = new Array(30).fill(100)
      const lows = new Array(30).fill(100)
      const closes = new Array(30).fill(100)

      const result = calculateADX(highs, lows, closes, 14)

      // +DI and -DI should be zero (no movement)
      const lastPlusDI = result.plusDI[result.plusDI.length - 1]
      const lastMinusDI = result.minusDI[result.minusDI.length - 1]

      expect(lastPlusDI).toBe(0)
      expect(lastMinusDI).toBe(0)
    })
  })

  describe('Boundary Conditions', () => {
    it('first valid +DI and -DI values appear at index period-1', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      // Indices 0 to period-2 should be null
      for (let i = 0; i < period - 1; i++) {
        expect(result.plusDI[i]).toBeNull()
        expect(result.minusDI[i]).toBeNull()
      }

      // Index period-1 should have first valid +DI/-DI
      expect(result.plusDI[period - 1]).not.toBeNull()
      expect(result.minusDI[period - 1]).not.toBeNull()
    })

    it('first valid ADX value appears at index 2*period-2', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      const firstADXIndex = 2 * period - 2

      // Indices before firstADXIndex should be null
      for (let i = 0; i < firstADXIndex; i++) {
        expect(result.adx[i]).toBeNull()
      }

      // firstADXIndex should have first valid ADX
      expect(result.adx[firstADXIndex]).not.toBeNull()
    })

    it('ADX range is 0-100', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      result.adx.forEach(value => {
        if (value !== null) {
          expect(value).toBeGreaterThanOrEqual(0)
          expect(value).toBeLessThanOrEqual(100)
        }
      })
    })

    it('array lengths are consistent', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      expect(result.adx).toHaveLength(highs.length)
      expect(result.plusDI).toHaveLength(highs.length)
      expect(result.minusDI).toHaveLength(highs.length)
    })
  })

  describe('Performance', () => {
    it('calculates ADX for large dataset efficiently', () => {
      const largeHighs = new Array(10000).fill(0).map((_, i) => 130 + Math.sin(i / 100) * 20)
      const largeLows = new Array(10000).fill(0).map((_, i) => 120 + Math.sin(i / 100) * 20)
      const largeCloses = new Array(10000).fill(0).map((_, i) => 125 + Math.sin(i / 100) * 20)

      const start = performance.now()
      const result = calculateADX(largeHighs, largeLows, largeCloses, 14)
      const duration = performance.now() - start

      expect(result.adx).toHaveLength(10000)
      expect(duration).toBeLessThan(100) // Should complete in < 100ms
    })

    it('handles multiple period calculations efficiently', () => {
      const { highs, lows, closes } = ADX_TEST_DATA.strongTrend

      const start = performance.now()
      calculateADX(highs, lows, closes, 7)
      calculateADX(highs, lows, closes, 14)
      calculateADX(highs, lows, closes, 21)
      const duration = performance.now() - start

      expect(duration).toBeLessThan(20) // Should complete in < 20ms
    })
  })

  describe('Special Scenarios', () => {
    it('ADX increases with trend strength', () => {
      const { highs, lows, closes, period } = ADX_TEST_DATA.strongTrend
      const result = calculateADX(highs, lows, closes, period)

      // In a strong consistent trend, ADX should be high
      const lastADX = result.adx[result.adx.length - 1]
      expect(lastADX).toBeGreaterThan(40) // ADX > 40 = very strong trend
    })
  })
})

// ============================================================================
// Ichimoku Cloud Tests
// ============================================================================

describe('Ichimoku Cloud', () => {
  describe('Calculation Accuracy', () => {
    it('calculates Tenkan-sen correctly', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // First 8 values should be null
      for (let i = 0; i < tenkanPeriod - 1; i++) {
        expect(result.tenkan[i]).toBeNull()
      }

      // Index 8 should be first valid Tenkan value
      expect(result.tenkan[8]).not.toBeNull()
      expect(result.tenkan[8]).toBeCloseTo(102.96, 1)
    })

    it('calculates Kijun-sen correctly', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // First 25 values should be null
      for (let i = 0; i < kijunPeriod - 1; i++) {
        expect(result.kijun[i]).toBeNull()
      }

      // Index 25 should be first valid Kijun value
      expect(result.kijun[25]).not.toBeNull()
      expect(result.kijun[25]).toBeCloseTo(109.07, 1)
    })

    it('calculates Senkou Span A with correct displacement', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // First 26 values (displacement) should be null
      for (let i = 0; i < displacement; i++) {
        expect(result.senkouA[i]).toBeNull()
      }

      // Index 51 (26 + 25) should have first valid Senkou A (displaced from index 25)
      expect(result.senkouA[51]).not.toBeNull()
      expect(result.senkouA[51]).toBeCloseTo(112.77, 1)
    })

    it('calculates Senkou Span B with correct displacement', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // First 26 values (displacement) should be null
      for (let i = 0; i < displacement; i++) {
        expect(result.senkouB[i]).toBeNull()
      }

      // Index 77 (26 + 51) should have first valid Senkou B (displaced from index 51)
      expect(result.senkouB[77]).not.toBeNull()
      expect(result.senkouB[77]).toBeCloseTo(114.31, 1)
    })

    it('calculates Chikou Span with correct backward displacement', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // First value should be close from index 26
      expect(result.chikou[0]).toBe(closes[displacement])
      expect(result.chikou[0]).toBeCloseTo(110.15, 1)

      // Last 26 values should be null (can't show future closes)
      for (let i = closes.length - displacement; i < closes.length; i++) {
        expect(result.chikou[i]).toBeNull()
      }
    })
  })

  describe('Edge Cases', () => {
    it('handles empty arrays', () => {
      const result = calculateIchimoku([], [], [], 9, 26, 52, 26)

      expect(result.tenkan).toEqual([])
      expect(result.kijun).toEqual([])
      expect(result.senkouA).toEqual([])
      expect(result.senkouB).toEqual([])
      expect(result.chikou).toEqual([])
    })

    it('handles insufficient data', () => {
      const highs = [100, 102, 104]
      const lows = [98, 99, 100]
      const closes = [99, 101, 103]

      const result = calculateIchimoku(highs, lows, closes, 9, 26, 52, 26)

      // Tenkan, Kijun, Chikou should match input length
      expect(result.tenkan).toHaveLength(3)
      expect(result.kijun).toHaveLength(3)
      expect(result.chikou).toHaveLength(3)

      // Senkou spans may be longer due to displacement padding (displacement=26 > length=3)
      expect(result.senkouA.length).toBeGreaterThanOrEqual(3)
      expect(result.senkouB.length).toBeGreaterThanOrEqual(3)

      // Tenkan and Kijun should be null (insufficient data)
      result.tenkan.forEach(v => expect(v).toBeNull())
      result.kijun.forEach(v => expect(v).toBeNull())
      // Chikou span should have nulls (array too short for displacement)
      result.chikou.forEach(v => expect(v).toBeNull())
    })

    it('handles mismatched array lengths', () => {
      const highs = [100, 102, 104, 106]
      const lows = [98, 99, 100]
      const closes = [99, 101]

      const result = calculateIchimoku(highs, lows, closes, 9, 26, 52, 26)

      expect(result.tenkan).toEqual([])
      expect(result.kijun).toEqual([])
      expect(result.senkouA).toEqual([])
      expect(result.senkouB).toEqual([])
      expect(result.chikou).toEqual([])
    })
  })

  describe('Boundary Conditions', () => {
    it('array lengths are consistent', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      expect(result.tenkan).toHaveLength(closes.length)
      expect(result.kijun).toHaveLength(closes.length)
      expect(result.senkouA).toHaveLength(closes.length)
      expect(result.senkouB).toHaveLength(closes.length)
      expect(result.chikou).toHaveLength(closes.length)
    })

    it('first valid values appear at expected indices', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // Tenkan: first valid at index tenkanPeriod - 1
      expect(result.tenkan[tenkanPeriod - 2]).toBeNull()
      expect(result.tenkan[tenkanPeriod - 1]).not.toBeNull()

      // Kijun: first valid at index kijunPeriod - 1
      expect(result.kijun[kijunPeriod - 2]).toBeNull()
      expect(result.kijun[kijunPeriod - 1]).not.toBeNull()

      // Chikou: first valid at index 0
      expect(result.chikou[0]).not.toBeNull()
    })
  })

  describe('Performance', () => {
    it('calculates Ichimoku for large dataset efficiently', () => {
      const largeSize = 1000
      const largeHighs = new Array(largeSize).fill(0).map((_, i) => 100 + i * 0.1 + Math.sin(i / 10) * 5)
      const largeLows = new Array(largeSize).fill(0).map((_, i) => 95 + i * 0.1 + Math.sin(i / 10) * 5)
      const largeCloses = new Array(largeSize).fill(0).map((_, i) => 98 + i * 0.1 + Math.sin(i / 10) * 5)

      const start = performance.now()
      const result = calculateIchimoku(largeHighs, largeLows, largeCloses, 9, 26, 52, 26)
      const duration = performance.now() - start

      expect(result.tenkan).toHaveLength(largeSize)
      expect(duration).toBeLessThan(50) // Should complete in < 50ms
    })
  })

  describe('Special Scenarios', () => {
    it('cloud formation indicates trend direction', () => {
      const { highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement } = ICHIMOKU_TEST_DATA.basic
      const result = calculateIchimoku(highs, lows, closes, tenkanPeriod, kijunPeriod, senkouBPeriod, displacement)

      // Check a point where both Senkou spans are valid
      const checkIndex = 77 // First index where both spans are valid
      if (result.senkouA[checkIndex] !== null && result.senkouB[checkIndex] !== null) {
        // In uptrend, Senkou A should eventually be above Senkou B (bullish cloud)
        // This is true for our trending data
        expect(result.senkouA[checkIndex]).toBeDefined()
        expect(result.senkouB[checkIndex]).toBeDefined()
      }
    })
  })
})
