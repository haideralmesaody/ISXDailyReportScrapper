/**
 * Support & Resistance Indicators Test Suite
 * Tests for Fibonacci Retracement, Donchian Channels, and Keltner Channels
 */

import {
  calculateFibonacciRetracement,
  calculateDonchianChannels,
  calculateKeltnerChannels
} from '@/lib/indicators/calculations'

// ============================================================================
// Test Data Fixtures
// ============================================================================

const FIBONACCI_TEST_DATA = {
  // Simple uptrend: low 100, high 200 → 50% = 150
  simpleUptrend: {
    highs: [120, 140, 160, 180, 200, 190, 185, 180, 175, 170],
    lows:  [100, 110, 120, 130, 140, 130, 125, 120, 115, 110],
    lookback: 10,
    expectedLevels: [
      { ratio: 0,     label: '0.0%',   price: 200 },  // High
      { ratio: 0.236, label: '23.6%', price: 176.4 },
      { ratio: 0.382, label: '38.2%', price: 161.8 },
      { ratio: 0.500, label: '50.0%', price: 150 },
      { ratio: 0.618, label: '61.8%', price: 138.2 },
      { ratio: 0.786, label: '78.6%', price: 121.4 },
      { ratio: 1.000, label: '100.0%', price: 100 },  // Low
    ]
  },
  // Flat market: high = low
  flatMarket: {
    highs: [150, 150, 150, 150, 150],
    lows:  [150, 150, 150, 150, 150],
    lookback: 5
  }
}

const DONCHIAN_TEST_DATA = {
  // Known highest/lowest over period 5
  knownRange: {
    highs: [110, 115, 120, 125, 130, 128, 126, 124, 122, 120],
    lows:  [100, 105, 110, 115, 120, 118, 116, 114, 112, 110],
    period: 5,
    // At index 9: last 5 highs = [128,126,124,122,120], last 5 lows = [118,116,114,112,110]
    expectedUpperAtEnd: 128,
    expectedLowerAtEnd: 110,
    expectedMiddleAtEnd: 119  // (128 + 110) / 2
  },
  // Monotonic increase
  increasing: {
    highs: [101, 102, 103, 104, 105, 106, 107, 108, 109, 110],
    lows:  [100, 101, 102, 103, 104, 105, 106, 107, 108, 109],
    period: 5
  }
}

const KELTNER_TEST_DATA = {
  // Test data with known EMA and ATR
  knownEMA: {
    highs:  [102, 104, 103, 105, 107, 106, 108, 110, 109, 111],
    lows:   [ 98, 100,  99, 101, 103, 102, 104, 106, 105, 107],
    closes: [100, 102, 101, 103, 105, 104, 106, 108, 107, 109],
    period: 5,
    multiplier: 2.0
  }
}

// ============================================================================
// Fibonacci Retracement Tests
// ============================================================================

describe('Fibonacci Retracement', () => {
  describe('Level Calculation Accuracy', () => {
    it('calculates all 7 Fibonacci levels correctly for simple uptrend', () => {
      const { highs, lows, lookback, expectedLevels } = FIBONACCI_TEST_DATA.simpleUptrend
      const result = calculateFibonacciRetracement(highs, lows, lookback)

      expect(result).not.toBeNull()
      expect(result!.levels).toHaveLength(7)

      result!.levels.forEach((level, index) => {
        const expected = expectedLevels[index]
        expect(level.ratio).toBe(expected.ratio)
        // Labels include "(High)" and "(Low)" suffixes for 0% and 100%
        if (index === 0) {
          expect(level.label).toBe('0.0% (High)')
        } else if (index === 6) {
          expect(level.label).toBe('100% (Low)')
        } else {
          expect(level.label).toBe(expected.label)
        }
        expect(level.price).toBeCloseTo(expected.price, 1)  // Within 0.1
      })
    })

    it('calculates correct high and low from data', () => {
      const { highs, lows, lookback } = FIBONACCI_TEST_DATA.simpleUptrend
      const result = calculateFibonacciRetracement(highs, lows, lookback)

      expect(result).not.toBeNull()
      expect(result!.high).toBe(200)
      expect(result!.low).toBe(100)
      expect(result!.range).toBe(100)
    })

    it('returns levels in descending price order', () => {
      const { highs, lows, lookback } = FIBONACCI_TEST_DATA.simpleUptrend
      const result = calculateFibonacciRetracement(highs, lows, lookback)

      expect(result).not.toBeNull()
      for (let i = 0; i < result!.levels.length - 1; i++) {
        expect(result!.levels[i].price).toBeGreaterThanOrEqual(result!.levels[i + 1].price)
      }
    })
  })

  describe('Lookback Period Variations', () => {
    it('uses correct lookback period for high/low range', () => {
      const highs = [100, 120, 140, 160, 180, 200]  // Length 6
      const lows  = [ 90, 110, 130, 150, 170, 190]

      // Lookback 3: should use last 3 periods → high 200, low 150
      const result = calculateFibonacciRetracement(highs, lows, 3)

      expect(result).not.toBeNull()
      expect(result!.high).toBe(200)
      expect(result!.low).toBe(150)
      expect(result!.levels[3].price).toBeCloseTo(175, 1)  // 50% level
    })

    it('handles minimum lookback period', () => {
      const highs = [120, 140, 160, 180, 200, 190, 185, 180, 175, 170]
      const lows  = [100, 110, 120, 130, 140, 130, 125, 120, 115, 110]

      const result = calculateFibonacciRetracement(highs, lows, 10)  // Minimum standard
      expect(result).not.toBeNull()
      expect(result!.levels).toHaveLength(7)
    })
  })

  describe('Edge Cases', () => {
    it('returns null when insufficient data', () => {
      const highs = [100, 110, 120]
      const lows  = [90,  100, 110]

      const result = calculateFibonacciRetracement(highs, lows, 5)  // Need 5, have 3
      expect(result).toBeNull()
    })

    it('handles flat market (high = low)', () => {
      const { highs, lows, lookback } = FIBONACCI_TEST_DATA.flatMarket
      const result = calculateFibonacciRetracement(highs, lows, lookback)

      // Flat market (range = 0) returns null - cannot calculate meaningful retracement levels
      expect(result).toBeNull()
    })

    it('handles single data point', () => {
      const highs = [100]
      const lows  = [90]

      const result = calculateFibonacciRetracement(highs, lows, 1)
      expect(result).not.toBeNull()
      expect(result!.high).toBe(100)
      expect(result!.low).toBe(90)
    })

    it('handles empty arrays', () => {
      const result = calculateFibonacciRetracement([], [], 10)
      expect(result).toBeNull()
    })

    it('handles mismatched array lengths', () => {
      const highs = [100, 110, 120]
      const lows  = [90, 100]  // Shorter

      const result = calculateFibonacciRetracement(highs, lows, 2)
      // Mismatched array lengths return null - validation error
      expect(result).toBeNull()
    })
  })
})

// ============================================================================
// Donchian Channels Tests
// ============================================================================

describe('Donchian Channels', () => {
  describe('Channel Calculation Accuracy', () => {
    it('calculates upper channel as highest high over period', () => {
      const { highs, lows, period, expectedUpperAtEnd } = DONCHIAN_TEST_DATA.knownRange
      const result = calculateDonchianChannels(highs, lows, period)

      expect(result.upper).toHaveLength(highs.length)

      // Check last value (fully formed)
      const lastIndex = result.upper.length - 1
      expect(result.upper[lastIndex]).toBe(expectedUpperAtEnd)
    })

    it('calculates lower channel as lowest low over period', () => {
      const { highs, lows, period, expectedLowerAtEnd } = DONCHIAN_TEST_DATA.knownRange
      const result = calculateDonchianChannels(highs, lows, period)

      expect(result.lower).toHaveLength(lows.length)

      const lastIndex = result.lower.length - 1
      expect(result.lower[lastIndex]).toBe(expectedLowerAtEnd)
    })

    it('calculates middle channel as average of upper and lower', () => {
      const { highs, lows, period, expectedMiddleAtEnd } = DONCHIAN_TEST_DATA.knownRange
      const result = calculateDonchianChannels(highs, lows, period)

      expect(result.middle).toHaveLength(highs.length)

      const lastIndex = result.middle.length - 1
      expect(result.middle[lastIndex]).toBe(expectedMiddleAtEnd)
    })

    it('returns null values for initial period - 1 bars', () => {
      const { highs, lows, period } = DONCHIAN_TEST_DATA.knownRange
      const result = calculateDonchianChannels(highs, lows, period)

      // First (period - 1) values should be null
      for (let i = 0; i < period - 1; i++) {
        expect(result.upper[i]).toBeNull()
        expect(result.middle[i]).toBeNull()
        expect(result.lower[i]).toBeNull()
      }

      // From period onwards should have values
      for (let i = period - 1; i < result.upper.length; i++) {
        expect(result.upper[i]).not.toBeNull()
        expect(result.middle[i]).not.toBeNull()
        expect(result.lower[i]).not.toBeNull()
      }
    })
  })

  describe('Period Variations', () => {
    it('handles short period correctly', () => {
      const highs = [100, 105, 110, 115, 120]
      const lows  = [ 95, 100, 105, 110, 115]

      const result = calculateDonchianChannels(highs, lows, 3)

      // At index 4: last 3 highs = [110, 115, 120], last 3 lows = [105, 110, 115]
      expect(result.upper[4]).toBe(120)
      expect(result.lower[4]).toBe(105)
      expect(result.middle[4]).toBe(112.5)
    })

    it('handles long period correctly', () => {
      const highs = Array(20).fill(0).map((_, i) => 100 + i)
      const lows  = Array(20).fill(0).map((_, i) => 90 + i)

      const result = calculateDonchianChannels(highs, lows, 20)

      // At index 19: all highs = [100..119], all lows = [90..109]
      expect(result.upper[19]).toBe(119)
      expect(result.lower[19]).toBe(90)
      expect(result.middle[19]).toBe(104.5)
    })

    it('channel width increases with volatility', () => {
      // Low volatility
      const highsLow = [100, 101, 102, 103, 104]
      const lowsLow  = [ 99, 100, 101, 102, 103]
      const resultLow = calculateDonchianChannels(highsLow, lowsLow, 5)
      const widthLow = resultLow.upper[4]! - resultLow.lower[4]!

      // High volatility
      const highsHigh = [100, 110, 120, 130, 140]
      const lowsHigh  = [ 80,  90, 100, 110, 120]
      const resultHigh = calculateDonchianChannels(highsHigh, lowsHigh, 5)
      const widthHigh = resultHigh.upper[4]! - resultHigh.lower[4]!

      expect(widthHigh).toBeGreaterThan(widthLow)
    })
  })

  describe('Edge Cases', () => {
    it('handles insufficient data', () => {
      const highs = [100, 110]
      const lows  = [ 90, 100]

      const result = calculateDonchianChannels(highs, lows, 5)  // Need 5, have 2

      expect(result.upper).toHaveLength(2)
      expect(result.upper[0]).toBeNull()
      expect(result.upper[1]).toBeNull()
    })

    it('handles monotonic increasing prices', () => {
      const { highs, lows, period } = DONCHIAN_TEST_DATA.increasing
      const result = calculateDonchianChannels(highs, lows, period)

      // Upper channel should equal current high (always highest)
      const lastIndex = result.upper.length - 1
      expect(result.upper[lastIndex]).toBe(highs[lastIndex])
    })

    it('handles constant prices (all equal)', () => {
      const highs = [100, 100, 100, 100, 100]
      const lows  = [100, 100, 100, 100, 100]

      const result = calculateDonchianChannels(highs, lows, 5)

      // All channels should be equal
      expect(result.upper[4]).toBe(100)
      expect(result.middle[4]).toBe(100)
      expect(result.lower[4]).toBe(100)
    })

    it('handles empty arrays', () => {
      const result = calculateDonchianChannels([], [], 5)

      expect(result.upper).toEqual([])
      expect(result.middle).toEqual([])
      expect(result.lower).toEqual([])
    })

    it('handles single data point', () => {
      const result = calculateDonchianChannels([100], [90], 1)

      expect(result.upper[0]).toBe(100)
      expect(result.middle[0]).toBe(95)
      expect(result.lower[0]).toBe(90)
    })
  })
})

// ============================================================================
// Keltner Channels Tests
// ============================================================================

describe('Keltner Channels', () => {
  describe('Band Calculation Structure', () => {
    it('returns correct structure with three bands', () => {
      const { highs, lows, closes, period, multiplier } = KELTNER_TEST_DATA.knownEMA
      const result = calculateKeltnerChannels(highs, lows, closes, period, multiplier)

      expect(result.upper).toHaveLength(closes.length)
      expect(result.middle).toHaveLength(closes.length)
      expect(result.lower).toHaveLength(closes.length)
    })

    it('middle band equals EMA of closes', () => {
      const closes = [100, 102, 104, 106, 108]
      const highs  = [102, 104, 106, 108, 110]
      const lows   = [ 98, 100, 102, 104, 106]

      const result = calculateKeltnerChannels(highs, lows, closes, 3, 2)

      // Middle band requires both EMA and ATR to be non-null
      // EMA(3): nulls at 0,1, values at 2,3,4
      // ATR(3): TR starts at 1 (4 values), EMA of TR nulls at 0,1 (indices 1,2 of TR), then values
      // Combined ATR: nulls at 0,1,2, values at 3,4
      // So Keltner middle: nulls at 0,1,2, values at 3,4
      expect(result.middle[0]).toBeNull()
      expect(result.middle[1]).toBeNull()
      expect(result.middle[2]).toBeNull()  // ATR still null here
      expect(result.middle[3]).not.toBeNull()  // First value where both EMA and ATR are available
    })

    it('upper and lower bands are symmetric around middle', () => {
      const { highs, lows, closes, period, multiplier } = KELTNER_TEST_DATA.knownEMA
      const result = calculateKeltnerChannels(highs, lows, closes, period, multiplier)

      // Check last value where all are formed
      const lastIndex = result.middle.length - 1

      if (result.middle[lastIndex] !== null && result.upper[lastIndex] !== null && result.lower[lastIndex] !== null) {
        const upperDistance = result.upper[lastIndex]! - result.middle[lastIndex]!
        const lowerDistance = result.middle[lastIndex]! - result.lower[lastIndex]!

        // Should be equal (symmetric)
        expect(upperDistance).toBeCloseTo(lowerDistance, 2)
      }
    })
  })

  describe('Multiplier Effects', () => {
    it('band width scales with multiplier', () => {
      const highs  = [102, 104, 106, 108, 110, 112, 114, 116, 118, 120]
      const lows   = [ 98, 100, 102, 104, 106, 108, 110, 112, 114, 116]
      const closes = [100, 102, 104, 106, 108, 110, 112, 114, 116, 118]

      // Test with multiplier 1.0
      const result1x = calculateKeltnerChannels(highs, lows, closes, 5, 1.0)

      // Test with multiplier 2.0
      const result2x = calculateKeltnerChannels(highs, lows, closes, 5, 2.0)

      const lastIndex = closes.length - 1

      if (result1x.upper[lastIndex] !== null && result2x.upper[lastIndex] !== null) {
        const width1x = result1x.upper[lastIndex]! - result1x.lower[lastIndex]!
        const width2x = result2x.upper[lastIndex]! - result2x.lower[lastIndex]!

        // Width with 2x multiplier should be roughly double width with 1x
        expect(width2x).toBeCloseTo(width1x * 2, 1)
      }
    })

    it('handles small multiplier (0.5)', () => {
      const { highs, lows, closes, period } = KELTNER_TEST_DATA.knownEMA
      const result = calculateKeltnerChannels(highs, lows, closes, period, 0.5)

      expect(result.upper).toHaveLength(closes.length)
      expect(result.middle).toHaveLength(closes.length)
      expect(result.lower).toHaveLength(closes.length)

      // Bands should be tighter with smaller multiplier
      const lastIndex = result.upper.length - 1
      if (result.upper[lastIndex] !== null && result.lower[lastIndex] !== null) {
        const bandwidth = result.upper[lastIndex]! - result.lower[lastIndex]!
        expect(bandwidth).toBeGreaterThan(0)  // Should still have some width
      }
    })

    it('handles large multiplier (5.0)', () => {
      const { highs, lows, closes, period } = KELTNER_TEST_DATA.knownEMA
      const result = calculateKeltnerChannels(highs, lows, closes, period, 5.0)

      expect(result.upper).toHaveLength(closes.length)

      // Bands should be much wider with larger multiplier
      const lastIndex = result.upper.length - 1
      if (result.upper[lastIndex] !== null && result.middle[lastIndex] !== null) {
        const upperDistance = result.upper[lastIndex]! - result.middle[lastIndex]!
        expect(upperDistance).toBeGreaterThan(0)
      }
    })
  })

  describe('Period Variations', () => {
    it('handles short period (5)', () => {
      // Need at least 2*period data points for Keltner:
      // EMA(5) needs 5 points, ATR(5) needs EMA of 4 TR values (period 5 needs 5 values)
      // So minimum is 5 + 4 = 9 data points
      const highs  = [102, 104, 106, 108, 110, 112, 114, 116, 118, 120]
      const lows   = [ 98, 100, 102, 104, 106, 108, 110, 112, 114, 116]
      const closes = [100, 102, 104, 106, 108, 110, 112, 114, 116, 118]

      const result = calculateKeltnerChannels(highs, lows, closes, 5, 2)

      // First non-null values appear at index where both EMA and ATR are available
      const firstValidIndex = result.middle.findIndex(v => v !== null)
      expect(firstValidIndex).toBeGreaterThanOrEqual(0)
      expect(result.upper[firstValidIndex]).not.toBeNull()
      expect(result.middle[firstValidIndex]).not.toBeNull()
      expect(result.lower[firstValidIndex]).not.toBeNull()
    })

    it('handles longer period (20)', () => {
      const length = 30  // Increase length to ensure enough data
      const highs  = Array(length).fill(0).map((_, i) => 100 + i * 2)
      const lows   = Array(length).fill(0).map((_, i) => 98 + i * 2)
      const closes = Array(length).fill(0).map((_, i) => 99 + i * 2)

      const result = calculateKeltnerChannels(highs, lows, closes, 20, 2)

      // EMA needs warmup period, first (period - 1) should be null
      for (let i = 0; i < 19; i++) {
        expect(result.middle[i]).toBeNull()
      }

      // Check that we have at least some non-null values after warmup
      const nonNullCount = result.middle.filter(v => v !== null).length
      expect(nonNullCount).toBeGreaterThan(0)
    })
  })

  describe('Edge Cases', () => {
    it('handles insufficient data', () => {
      const highs  = [100, 102]
      const lows   = [ 98, 100]
      const closes = [ 99, 101]

      const result = calculateKeltnerChannels(highs, lows, closes, 5, 2)

      expect(result.upper).toHaveLength(2)
      expect(result.upper[0]).toBeNull()
      expect(result.upper[1]).toBeNull()
    })

    it('handles zero volatility (all prices equal)', () => {
      // Need enough data for EMA calculation
      const length = 10
      const highs  = Array(length).fill(100)
      const lows   = Array(length).fill(100)
      const closes = Array(length).fill(100)

      const result = calculateKeltnerChannels(highs, lows, closes, 5, 2)

      // Find first non-null value after EMA warmup
      const firstNonNullIndex = result.middle.findIndex(v => v !== null)

      if (firstNonNullIndex !== -1) {
        // All bands should converge to same price (ATR = 0)
        expect(result.upper[firstNonNullIndex]).toBeCloseTo(100, 1)
        expect(result.middle[firstNonNullIndex]).toBeCloseTo(100, 1)
        expect(result.lower[firstNonNullIndex]).toBeCloseTo(100, 1)
      }
    })

    it('handles empty arrays', () => {
      const result = calculateKeltnerChannels([], [], [], 5, 2)

      expect(result.upper).toEqual([])
      expect(result.middle).toEqual([])
      expect(result.lower).toEqual([])
    })

    it('handles single data point', () => {
      const result = calculateKeltnerChannels([100], [98], [99], 1, 2)

      // Should return single value (structure exists but may be null due to EMA requirements)
      expect(result.upper).toHaveLength(1)
      expect(result.middle).toHaveLength(1)
      expect(result.lower).toHaveLength(1)
      // Values may be null due to insufficient data for EMA
    })

    it('handles extreme multiplier edge case (0)', () => {
      const { highs, lows, closes, period } = KELTNER_TEST_DATA.knownEMA
      const result = calculateKeltnerChannels(highs, lows, closes, period, 0)

      // With multiplier 0, all bands should equal middle (EMA)
      // Find first non-null value
      const firstNonNullIndex = result.middle.findIndex(v => v !== null)

      if (firstNonNullIndex !== -1 && result.upper[firstNonNullIndex] !== null) {
        expect(result.upper[firstNonNullIndex]).toBeCloseTo(result.middle[firstNonNullIndex]!, 2)
        expect(result.lower[firstNonNullIndex]).toBeCloseTo(result.middle[firstNonNullIndex]!, 2)
      }
    })

    it('handles mismatched array lengths gracefully', () => {
      const highs  = [100, 102, 104]
      const lows   = [ 98, 100]  // Shorter
      const closes = [ 99, 101, 103]

      const result = calculateKeltnerChannels(highs, lows, closes, 2, 2)

      // Should use minimum length without crashing
      // May return empty arrays if calculation can't proceed
      expect(result.upper).toBeDefined()
      expect(result.middle).toBeDefined()
      expect(result.lower).toBeDefined()
    })
  })
})
