/**
 * Test Data Utilities for Indicator Testing
 *
 * Contains known datasets with expected outputs for validating
 * technical indicator calculations against industry-standard formulas.
 *
 * Data sources:
 * - Investopedia indicator examples
 * - TradingView reference calculations
 * - Industry-standard test vectors
 */

// ============================================================================
// Simple Moving Average (SMA) Test Data
// ============================================================================

export const SMA_TEST_DATA = {
  // Basic SMA test with period 3
  simple: {
    input: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    period: 3,
    expected: [
      null, // Not enough data
      null, // Not enough data
      2,    // (1+2+3)/3
      3,    // (2+3+4)/3
      4,    // (3+4+5)/3
      5,    // (4+5+6)/3
      6,    // (5+6+7)/3
      7,    // (6+7+8)/3
      8,    // (7+8+9)/3
      9     // (8+9+10)/3
    ]
  },

  // Real-world price data
  priceData: {
    input: [100, 102, 101, 103, 104, 102, 103, 105, 104, 106],
    period: 5,
    expected: [
      null,
      null,
      null,
      null,
      102,   // (100+102+101+103+104)/5
      102.4, // (102+101+103+104+102)/5
      102.6, // (101+103+104+102+103)/5
      103.4, // (103+104+102+103+105)/5
      103.6, // (104+102+103+105+104)/5
      104    // (102+103+105+104+106)/5
    ]
  }
}

// ============================================================================
// Exponential Moving Average (EMA) Test Data
// ============================================================================

export const EMA_TEST_DATA = {
  // EMA with period 5 (multiplier = 2/(5+1) = 0.333...)
  basic: {
    input: [22.27, 22.19, 22.08, 22.17, 22.18, 22.13, 22.23, 22.43, 22.24, 22.29],
    period: 5,
    // First EMA = SMA of first 5 values
    // Subsequent EMA = (Close - Previous EMA) * multiplier + Previous EMA
    expected: [
      null,
      null,
      null,
      null,
      22.178,  // SMA(first 5) = (22.27+22.19+22.08+22.17+22.18)/5
      22.168,  // (22.13 - 22.178) * 0.333 + 22.178
      22.186,  // (22.23 - 22.168) * 0.333 + 22.168
      22.267,  // (22.43 - 22.186) * 0.333 + 22.186
      22.258,  // (22.24 - 22.267) * 0.333 + 22.267
      22.268   // (22.29 - 22.258) * 0.333 + 22.258
    ]
  }
}

// ============================================================================
// Relative Strength Index (RSI) Test Data
// ============================================================================

export const RSI_TEST_DATA = {
  // Classic RSI example with 14-period calculation
  // Verified calculation using Wilder's smoothing method
  classic: {
    // 14-day RSI with known closing prices
    input: [
      44.00, 44.34, 44.09, 43.61, 44.33,
      44.83, 45.10, 45.42, 45.84, 46.08,
      45.89, 46.03, 45.61, 46.28, 46.28,
      46.00, 46.03, 46.41, 46.22, 45.64
    ],
    period: 14,
    // First RSI at index 14 (after 14 price changes)
    // RSI = 100 - (100 / (1 + RS))
    // RS = Average Gain / Average Loss
    // Initial avg gain = 0.258571, avg loss = 0.095714 → RS = 2.701493 → RSI = 72.98
    expected: [
      null, null, null, null, null,
      null, null, null, null, null,
      null, null, null, null,
      72.98,  // First valid RSI value (verified calculation)
      68.80,  // Second RSI (smoothed using Wilder's method)
      69.01,  // Third RSI
      71.56,  // Fourth RSI
      68.52,  // Fifth RSI
      60.14   // Sixth RSI
    ]
  },

  // Simplified test for edge cases
  uptrend: {
    input: [10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25],
    period: 14,
    // All gains, no losses → RSI should be near 100
    expected: [
      null, null, null, null, null, null, null,
      null, null, null, null, null, null, null,
      100,  // RSI = 100 when no losses
      100
    ]
  },

  downtrend: {
    input: [25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10],
    period: 14,
    // All losses, no gains → RSI should be near 0
    expected: [
      null, null, null, null, null, null, null,
      null, null, null, null, null, null, null,
      0,  // RSI = 0 when no gains
      0
    ]
  }
}

// ============================================================================
// MACD Test Data
// ============================================================================

export const MACD_TEST_DATA = {
  // MACD with standard parameters (12, 26, 9)
  // Extended to 40 points to get valid signal line values
  standard: {
    input: [
      12.26, 12.27, 12.29, 12.28, 12.30,
      12.32, 12.35, 12.38, 12.40, 12.42,
      12.44, 12.46, 12.48, 12.50, 12.52,
      12.54, 12.56, 12.58, 12.60, 12.62,
      12.64, 12.66, 12.68, 12.70, 12.72,
      12.74, 12.76, 12.78, 12.80, 12.82,
      12.84, 12.86, 12.88, 12.90, 12.92,
      12.94, 12.96, 12.98, 13.00, 13.02
    ],
    fastPeriod: 12,
    slowPeriod: 26,
    signalPeriod: 9,
    // MACD Line = EMA(12) - EMA(26), first valid at index 25 (slowPeriod-1)
    // Signal Line = EMA(9) of MACD Line, first valid at index 33
    // Histogram = MACD Line - Signal Line
    expectedMacd: [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,0.1409,0.1409,0.1408,0.1408,0.1407,0.1407,0.1406,0.1406,0.1406,0.1405,0.1405,0.1405,0.1404,0.1404,0.1404],
    expectedSignal: [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,0.1407,0.1407,0.1406,0.1406,0.1406,0.1405,0.1405],
    expectedHistogram: [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,-0.0002,-0.0002,-0.0002,-0.0002,-0.0001,-0.0001,-0.0001]
  }
}

// ============================================================================
// Bollinger Bands Test Data
// ============================================================================

export const BOLLINGER_BANDS_TEST_DATA = {
  basic: {
    input: [100, 102, 101, 103, 104, 102, 103, 105, 104, 106, 107, 108, 110, 109, 111, 113, 112, 114, 115, 116],
    period: 10,
    stdDevMultiplier: 2,
    // Middle Band = SMA(10)
    // Upper Band = Middle + (2 * Standard Deviation)
    // Lower Band = Middle - (2 * Standard Deviation)
    expectedMiddle: [null,null,null,null,null,null,null,null,null,103,103.7,104.3,105.2,105.8,106.5,107.6,108.5,109.4,110.5,111.5],
    expectedUpper: [null,null,null,null,null,null,null,null,null,106.46,107.28,108.5,110,110.84,112.24,113.68,114.24,115.48,116.24,117.24],
    expectedLower: [null,null,null,null,null,null,null,null,null,99.54,100.12,100.1,100.4,100.76,100.76,101.52,102.76,103.32,104.76,105.76]
  }
}

// ============================================================================
// ATR (Average True Range) Test Data
// ============================================================================

export const ATR_TEST_DATA = {
  basic: {
    highs: [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150],
    lows: [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139],
    closes: [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149],
    period: 14,
    // ATR = EMA(True Range, period)
    // True Range = max(H-L, |H-PC|, |L-PC|)
    expectedATR: [null,null,null,null,null,null,null,null,null,null,null,null,null,null,11,11,11,11,11,11]
  }
}

// ============================================================================
// Stochastic Oscillator Test Data
// ============================================================================

export const STOCHASTIC_TEST_DATA = {
  // Extended to 20 points to get valid %D values
  basic: {
    highs: [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150],
    lows: [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139],
    closes: [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149],
    kPeriod: 14,
    dPeriod: 3,
    // %K = (Current Close - Lowest Low) / (Highest High - Lowest Low) * 100
    // First valid %K at index 13 (kPeriod-1)
    // %D = SMA(3) of %K, first valid at index 15 (13 + dPeriod - 1)
    expectedK: [null,null,null,null,null,null,null,null,null,null,null,null,null,95.83,95.83,95.83,95.83,95.83,95.83,95.83],
    expectedD: [null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,95.83,95.83,95.83,95.83,95.83]
  }
}

// ============================================================================
// Momentum (MOM) Test Data
// ============================================================================

export const MOMENTUM_TEST_DATA = {
  simple: {
    input: [100, 102, 104, 103, 105, 107, 106, 108, 110, 109, 111, 113],
    period: 4,
    // MOM = Close - Close[period ago]
    expected: [
      null, null, null, null,
      5,    // 105 - 100
      5,    // 107 - 102
      2,    // 106 - 104
      5,    // 108 - 103
      5,    // 110 - 105
      2,    // 109 - 107
      5,    // 111 - 106
      5     // 113 - 108
    ]
  }
}

// ============================================================================
// Rate of Change (ROC) Test Data
// ============================================================================

export const ROC_TEST_DATA = {
  simple: {
    input: [100, 102, 104, 103, 105, 107, 106, 108, 110, 109, 111, 113],
    period: 4,
    // ROC = ((Close - Close[period ago]) / Close[period ago]) * 100
    expected: [
      null, null, null, null,
      5.00,   // ((105 - 100) / 100) * 100
      4.90,   // ((107 - 102) / 102) * 100
      1.92,   // ((106 - 104) / 104) * 100
      4.85,   // ((108 - 103) / 103) * 100
      4.76,   // ((110 - 105) / 105) * 100
      1.87,   // ((109 - 107) / 107) * 100
      4.72,   // ((111 - 106) / 106) * 100
      4.63    // ((113 - 108) / 108) * 100
    ]
  }
}

// ============================================================================
// Williams %R Test Data
// ============================================================================

export const WILLIAMS_R_TEST_DATA = {
  basic: {
    highs: [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150],
    lows: [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139],
    closes: [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149],
    period: 14,
    // Williams %R = ((Highest High - Close) / (Highest High - Lowest Low)) * -100
    // Range: -100 (oversold) to 0 (overbought)
  }
}

// ============================================================================
// CCI (Commodity Channel Index) Test Data
// ============================================================================

export const CCI_TEST_DATA = {
  basic: {
    highs: [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150],
    lows: [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139],
    closes: [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149],
    period: 20,
    // CCI = (Typical Price - SMA(Typical Price)) / (0.015 * Mean Deviation)
    // Typical Price = (High + Low + Close) / 3
  }
}

// ============================================================================
// Edge Case Test Data
// ============================================================================

export const EDGE_CASE_DATA = {
  emptyArray: [],
  singleValue: [100],
  twoValues: [100, 102],
  allSameValues: [100, 100, 100, 100, 100],
  withZeros: [0, 1, 2, 3, 4, 5],
  withNegatives: [-5, -3, -1, 0, 1, 3, 5],
  largeValues: [1000000, 1000100, 1000200, 1000300],
  smallValues: [0.001, 0.002, 0.003, 0.004],
}

// ============================================================================
// Performance Test Data
// ============================================================================

export function generateLargeDataset(size: number): number[] {
  const data: number[] = []
  let price = 100

  for (let i = 0; i < size; i++) {
    // Simulate random price movement
    const change = (Math.random() - 0.5) * 2 // -1 to +1
    price = Math.max(1, price + change) // Prevent negative prices
    data.push(price)
  }

  return data
}

// Generate OHLC data for complex indicators
export function generateOHLCData(size: number) {
  const data = []
  let basePrice = 100

  for (let i = 0; i < size; i++) {
    const volatility = Math.random() * 3 // 0-3 range
    const open = basePrice
    const close = basePrice + (Math.random() - 0.5) * volatility
    const high = Math.max(open, close) + Math.random() * volatility
    const low = Math.min(open, close) - Math.random() * volatility

    data.push({
      time: `2024-01-${String(i + 1).padStart(2, '0')}`,
      open,
      high,
      low,
      close
    })

    basePrice = close
  }

  return data
}

// ============================================================================
// Average Directional Index (ADX) Test Data
// ============================================================================

export const ADX_TEST_DATA = {
  // Strong uptrend: 30-point OHLC data with period 14
  strongTrend: {
    highs: [
      130, 132, 134, 136, 138, 140, 142, 144, 146, 148,
      150, 152, 154, 156, 158, 160, 162, 164, 166, 168,
      170, 172, 174, 176, 178, 180, 182, 184, 186, 188
    ],
    lows: [
      120, 121, 122, 123, 124, 125, 126, 127, 128, 129,
      130, 131, 132, 133, 134, 135, 136, 137, 138, 139,
      140, 141, 142, 143, 144, 145, 146, 147, 148, 149
    ],
    closes: [
      125, 127, 129, 131, 133, 135, 137, 139, 141, 143,
      145, 147, 149, 151, 153, 155, 157, 159, 161, 163,
      165, 167, 169, 171, 173, 175, 177, 179, 181, 183
    ],
    period: 14,
    expectedPlusDI: [
      null, null, null, null, null, null, null, null, null, null, null, null, null,
      11.26, 10.96, 10.66, 10.36, 10.06, 9.76, 9.46, 9.18, 8.90, 8.63, 8.36, 8.11, 7.87, 7.64, 7.41, 7.20, 6.99
    ],
    expectedMinusDI: [
      null, null, null, null, null, null, null, null, null, null, null, null, null,
      0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00, 0.00
    ],
    expectedADX: [
      null, null, null, null, null, null, null, null, null, null, null, null, null,
      null, null, null, null, null, null, null, null, null, null, null, null, null,
      100.00, 100.00, 100.00, 100.00
    ]
  },

  // Sideways market: Low volatility, weak trend
  sideways: {
    highs: [
      102, 103, 101, 102, 101, 103, 102, 101, 103, 102,
      101, 102, 103, 101, 102, 103, 101, 102, 101, 103,
      102, 101, 103, 102, 101, 102, 103, 101, 102, 103
    ],
    lows: [
      98, 97, 99, 98, 99, 97, 98, 99, 97, 98,
      99, 98, 97, 99, 98, 97, 99, 98, 99, 97,
      98, 99, 97, 98, 99, 98, 97, 99, 98, 97
    ],
    closes: [
      100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
      100, 100, 100, 100, 100, 100, 100, 100, 100, 100,
      100, 100, 100, 100, 100, 100, 100, 100, 100, 100
    ],
    period: 14
  }
}

// ============================================================================
// Ichimoku Cloud Test Data
// ============================================================================

export const ICHIMOKU_TEST_DATA = {
  // Basic test with 80 points (minimum for 52-period + 26 displacement)
  basic: {
    highs: [
      102.00, 103.75, 105.40, 106.88, 108.10, 109.02, 109.61, 109.90, 109.91, 109.70,
      109.38, 109.07, 108.83, 108.71, 108.76, 109.01, 109.49, 110.20, 111.13, 112.28,
      113.62, 115.13, 116.77, 118.51, 120.31, 122.14, 123.94, 125.68, 127.31, 128.78,
      130.06, 131.11, 131.90, 132.41, 132.63, 132.58, 132.26, 131.71, 130.96, 130.06,
      129.05, 127.98, 126.91, 125.88, 124.95, 124.17, 123.58, 123.21, 123.11, 123.30,
      123.80, 124.63, 125.79, 127.27, 129.04, 131.10, 133.40, 135.91, 138.59, 141.39,
      144.27, 147.19, 150.10, 152.96, 155.73, 158.35, 160.79, 163.02, 164.99, 166.68,
      168.07, 169.13, 169.86, 170.26, 170.34, 170.11, 169.60, 168.83, 167.84, 166.67
    ],
    lows: [
      96.00, 97.21, 98.60, 100.11, 101.68, 102.83, 102.96, 103.07, 103.22, 103.44,
      103.77, 104.24, 104.88, 105.70, 106.71, 107.90, 109.27, 110.80, 112.47, 114.25,
      116.10, 117.98, 119.84, 121.65, 123.35, 124.91, 126.29, 127.46, 128.40, 129.09,
      129.53, 129.72, 129.67, 129.39, 128.92, 128.27, 127.47, 126.56, 125.57, 124.53,
      123.48, 122.46, 121.51, 120.68, 120.00, 119.50, 119.22, 119.19, 119.43, 119.98,
      120.84, 122.03, 123.53, 125.32, 127.40, 129.72, 132.27, 135.00, 137.87, 140.84,
      143.87, 146.91, 149.92, 152.85, 155.65, 158.29, 160.72, 162.91, 164.84, 166.47,
      167.79, 168.77, 169.41, 169.71, 169.70, 169.38, 168.79, 167.96, 166.92, 165.72
    ],
    closes: [
      100.00, 101.10, 102.17, 103.19, 104.15, 105.02, 105.80, 106.46, 107.00, 107.42,
      107.71, 107.88, 107.94, 107.90, 107.78, 107.61, 107.41, 107.22, 107.06, 106.96,
      106.96, 107.07, 107.32, 107.74, 108.35, 109.15, 110.15, 111.35, 112.73, 114.30,
      116.04, 117.93, 119.93, 122.03, 124.19, 126.38, 128.56, 130.70, 132.76, 134.71,
      136.51, 138.15, 139.60, 140.85, 141.88, 142.69, 143.29, 143.66, 143.83, 143.80,
      143.59, 143.23, 142.73, 142.14, 141.47, 140.76, 140.04, 139.35, 138.71, 138.16,
      137.73, 137.45, 137.34, 137.43, 137.74, 138.28, 139.07, 140.10, 141.37, 142.88,
      144.61, 146.54, 148.65, 150.91, 153.30, 155.77, 158.31, 160.87, 163.42, 165.94
    ],
    tenkanPeriod: 9,
    kijunPeriod: 26,
    senkouBPeriod: 52,
    displacement: 26
  }
}
