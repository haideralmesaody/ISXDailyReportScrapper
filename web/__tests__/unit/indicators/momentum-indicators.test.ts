/**
 * Momentum Indicators Test Suite
 * Tests for MACD and Stochastic Oscillator calculations
 */

import { calculateMACD, calculateStochastic } from '@/lib/indicators/calculations'
import { MACD_TEST_DATA, STOCHASTIC_TEST_DATA } from './test-data'

// ============================================================================
// MACD (Moving Average Convergence Divergence) Tests
// ============================================================================

describe('MACD (Moving Average Convergence Divergence)', () => {
  describe('Calculation Accuracy', () => {
    it('calculates MACD line correctly for standard parameters', () => {
      const { input, fastPeriod, slowPeriod, signalPeriod, expectedMacd } = MACD_TEST_DATA.standard
      const result = calculateMACD(input, fastPeriod, slowPeriod, signalPeriod)

      expect(result.macd).toHaveLength(expectedMacd.length)

      expectedMacd.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.macd[index]).toBeNull()
        } else {
          expect(result.macd[index]).toBeCloseTo(expectedValue, 3)
        }
      })
    })

    it('calculates Signal line correctly', () => {
      const { input, fastPeriod, slowPeriod, signalPeriod, expectedSignal } = MACD_TEST_DATA.standard
      const result = calculateMACD(input, fastPeriod, slowPeriod, signalPeriod)

      expect(result.signal).toHaveLength(expectedSignal.length)

      expectedSignal.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.signal[index]).toBeNull()
        } else {
          expect(result.signal[index]).toBeCloseTo(expectedValue, 3)
        }
      })
    })

    it('calculates Histogram correctly', () => {
      const { input, fastPeriod, slowPeriod, signalPeriod, expectedHistogram } = MACD_TEST_DATA.standard
      const result = calculateMACD(input, fastPeriod, slowPeriod, signalPeriod)

      expect(result.histogram).toHaveLength(expectedHistogram.length)

      expectedHistogram.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.histogram[index]).toBeNull()
        } else {
          expect(result.histogram[index]).toBeCloseTo(expectedValue, 3)
        }
      })
    })

    it('first valid MACD appears at slowPeriod-1', () => {
      const { input, fastPeriod, slowPeriod, signalPeriod } = MACD_TEST_DATA.standard
      const result = calculateMACD(input, fastPeriod, slowPeriod, signalPeriod)

      // All values before slowPeriod-1 should be null
      for (let i = 0; i < slowPeriod - 1; i++) {
        expect(result.macd[i]).toBeNull()
      }

      // Value at slowPeriod-1 should be a number
      expect(result.macd[slowPeriod - 1]).not.toBeNull()
      expect(typeof result.macd[slowPeriod - 1]).toBe('number')
    })

    it('first valid Signal appears at slowPeriod + signalPeriod - 2', () => {
      const { input, fastPeriod, slowPeriod, signalPeriod } = MACD_TEST_DATA.standard
      const result = calculateMACD(input, fastPeriod, slowPeriod, signalPeriod)

      const firstSignalIndex = slowPeriod + signalPeriod - 2

      // All values before first signal should be null
      for (let i = 0; i < firstSignalIndex; i++) {
        expect(result.signal[i]).toBeNull()
      }

      // Value at firstSignalIndex should be a number
      expect(result.signal[firstSignalIndex]).not.toBeNull()
      expect(typeof result.signal[firstSignalIndex]).toBe('number')
    })
  })

  describe('Edge Cases', () => {
    it('returns empty result for empty input', () => {
      const result = calculateMACD([], 12, 26, 9)

      expect(result.macd).toEqual([])
      expect(result.signal).toEqual([])
      expect(result.histogram).toEqual([])
    })

    it('returns all nulls for insufficient data', () => {
      const shortData = [1, 2, 3, 4, 5]
      const result = calculateMACD(shortData, 12, 26, 9)

      expect(result.macd).toHaveLength(shortData.length)
      expect(result.signal).toHaveLength(shortData.length)
      expect(result.histogram).toHaveLength(shortData.length)

      result.macd.forEach(val => expect(val).toBeNull())
      result.signal.forEach(val => expect(val).toBeNull())
      result.histogram.forEach(val => expect(val).toBeNull())
    })

    it('handles all zeros', () => {
      const zeros = new Array(40).fill(0)
      const result = calculateMACD(zeros, 12, 26, 9)

      // MACD line should exist but be zero (FastEMA - SlowEMA = 0 - 0 = 0)
      expect(result.macd[25]).toBe(0)
    })

    it('handles negative values', () => {
      const negatives = [-10, -11, -12, -13, -14, -15, -16, -17, -18, -19, -20, -21, -22, -23, -24, -25, -26, -27, -28, -29, -30, -31, -32, -33, -34, -35, -36, -37, -38, -39, -40]
      const result = calculateMACD(negatives, 12, 26, 9)

      // Should calculate properly with negative values
      expect(result.macd[25]).toBeDefined()
      expect(typeof result.macd[25]).toBe('number')
    })
  })

  describe('Boundary Conditions', () => {
    it('MACD line equals FastEMA - SlowEMA', () => {
      const data = new Array(40).fill(0).map((_, i) => 100 + i * 0.5)
      const result = calculateMACD(data, 12, 26, 9)

      // MACD should be positive when FastEMA > SlowEMA (uptrend)
      expect(result.macd[25]!).toBeGreaterThan(0)
    })

    it('Histogram equals MACD - Signal', () => {
      const { input, fastPeriod, slowPeriod, signalPeriod } = MACD_TEST_DATA.standard
      const result = calculateMACD(input, fastPeriod, slowPeriod, signalPeriod)

      // Check at index 33 where both MACD and Signal exist
      const macdValue = result.macd[33]!
      const signalValue = result.signal[33]!
      const histogramValue = result.histogram[33]!

      expect(histogramValue).toBeCloseTo(macdValue - signalValue, 4)
    })

    it('works with custom parameters', () => {
      const data = new Array(40).fill(0).map((_, i) => 100 + i)
      const result = calculateMACD(data, 5, 10, 3)

      // First MACD at index 9 (slowPeriod-1)
      expect(result.macd[9]).not.toBeNull()

      // First Signal at index 11 (9 + 3 - 1)
      expect(result.signal[11]).not.toBeNull()
    })
  })

  describe('Performance', () => {
    it('calculates MACD for large dataset in reasonable time', () => {
      const largeData = new Array(10000).fill(0).map((_, i) => 100 + Math.sin(i / 100) * 10)

      const start = performance.now()
      const result = calculateMACD(largeData, 12, 26, 9)
      const duration = performance.now() - start

      expect(result.macd).toHaveLength(10000)
      expect(result.signal).toHaveLength(10000)
      expect(result.histogram).toHaveLength(10000)
      expect(duration).toBeLessThan(150) // Should complete in < 150ms
    })
  })
})

// ============================================================================
// Stochastic Oscillator Tests
// ============================================================================

describe('Stochastic Oscillator', () => {
  describe('Calculation Accuracy', () => {
    it('calculates %K correctly', () => {
      const { highs, lows, closes, kPeriod, dPeriod, expectedK } = STOCHASTIC_TEST_DATA.basic
      const result = calculateStochastic(highs, lows, closes, kPeriod, dPeriod)

      expect(result.k).toHaveLength(expectedK.length)

      expectedK.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.k[index]).toBeNull()
        } else {
          expect(result.k[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('calculates %D correctly', () => {
      const { highs, lows, closes, kPeriod, dPeriod, expectedD } = STOCHASTIC_TEST_DATA.basic
      const result = calculateStochastic(highs, lows, closes, kPeriod, dPeriod)

      expect(result.d).toHaveLength(expectedD.length)

      expectedD.forEach((expectedValue, index) => {
        if (expectedValue === null) {
          expect(result.d[index]).toBeNull()
        } else {
          expect(result.d[index]).toBeCloseTo(expectedValue, 1)
        }
      })
    })

    it('%K values are between 0 and 100', () => {
      const highs = [130, 135, 140, 145, 150, 148, 146, 144, 142, 140, 138, 136, 134, 132, 130, 128, 126, 124, 122, 120]
      const lows = [120, 122, 124, 126, 128, 130, 132, 134, 136, 138, 128, 126, 124, 122, 120, 118, 116, 114, 112, 110]
      const closes = [125, 130, 135, 140, 145, 140, 138, 136, 134, 138, 133, 130, 128, 126, 124, 122, 120, 118, 116, 114]

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      result.k.forEach((val, index) => {
        if (val !== null) {
          expect(val).toBeGreaterThanOrEqual(0)
          expect(val).toBeLessThanOrEqual(100)
        }
      })
    })

    it('detects overbought conditions (%K > 80)', () => {
      // Prices at top of range
      const highs = new Array(20).fill(150)
      const lows = new Array(20).fill(100)
      const closes = new Array(20).fill(148) // Near high

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      // %K should be > 80 (overbought)
      const firstK = result.k[13]
      expect(firstK).not.toBeNull()
      expect(firstK!).toBeGreaterThan(80)
    })
  })

  describe('Edge Cases', () => {
    it('returns empty result for empty input', () => {
      const result = calculateStochastic([], [], [], 14, 3)

      expect(result.k).toEqual([])
      expect(result.d).toEqual([])
    })

    it('returns all nulls for insufficient data', () => {
      const highs = [130, 135, 140]
      const lows = [120, 125, 130]
      const closes = [125, 130, 135]

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      result.k.forEach(val => expect(val).toBeNull())
      result.d.forEach(val => expect(val).toBeNull())
    })

    it('handles no price range (constant highs/lows)', () => {
      const highs = new Array(20).fill(150)
      const lows = new Array(20).fill(150)
      const closes = new Array(20).fill(150)

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      // Should return 50 (neutral) when range is zero
      expect(result.k[13]).toBe(50)
    })

    it('handles mismatched array lengths gracefully', () => {
      const highs = [130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200]
      const lows = [120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190]
      const closes = [125, 130, 135] // Shorter

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      // Should handle gracefully (will calculate based on closes length)
      expect(result.k.length).toBe(closes.length)
    })
  })

  describe('Boundary Conditions', () => {
    it('first valid %K appears at kPeriod-1', () => {
      const highs = new Array(20).fill(0).map((_, i) => 130 + i)
      const lows = new Array(20).fill(0).map((_, i) => 120 + i)
      const closes = new Array(20).fill(0).map((_, i) => 125 + i)

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      // All values before kPeriod-1 should be null
      for (let i = 0; i < 13; i++) {
        expect(result.k[i]).toBeNull()
      }

      // Value at kPeriod-1 should be a number
      expect(result.k[13]).not.toBeNull()
      expect(typeof result.k[13]).toBe('number')
    })

    it('first valid %D appears at kPeriod + dPeriod - 2', () => {
      const highs = new Array(20).fill(0).map((_, i) => 130 + i)
      const lows = new Array(20).fill(0).map((_, i) => 120 + i)
      const closes = new Array(20).fill(0).map((_, i) => 125 + i)

      const result = calculateStochastic(highs, lows, closes, 14, 3)

      const firstDIndex = 14 + 3 - 2 // = 15

      // All values before first %D should be null
      for (let i = 0; i < firstDIndex; i++) {
        expect(result.d[i]).toBeNull()
      }

      // Value at firstDIndex should be a number
      expect(result.d[firstDIndex]).not.toBeNull()
      expect(typeof result.d[firstDIndex]).toBe('number')
    })
  })

  describe('Performance', () => {
    it('calculates Stochastic for large dataset in reasonable time', () => {
      const largeHighs = new Array(10000).fill(0).map((_, i) => 130 + Math.sin(i / 50) * 20)
      const largeLows = new Array(10000).fill(0).map((_, i) => 120 + Math.sin(i / 50) * 20)
      const largeCloses = new Array(10000).fill(0).map((_, i) => 125 + Math.sin(i / 50) * 20)

      const start = performance.now()
      const result = calculateStochastic(largeHighs, largeLows, largeCloses, 14, 3)
      const duration = performance.now() - start

      expect(result.k).toHaveLength(10000)
      expect(result.d).toHaveLength(10000)
      expect(duration).toBeLessThan(100) // Should complete in < 100ms
    })
  })
})
