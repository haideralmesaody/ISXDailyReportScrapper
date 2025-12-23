// Calculate ATR expected values for test data

function calculateEMA(data, period) {
  if (!data || data.length === 0) return []
  if (period <= 0 || period > data.length) return data.map(() => null)

  const result = []
  const multiplier = 2 / (period + 1)

  // Calculate initial SMA
  let sum = 0
  for (let i = 0; i < period && i < data.length; i++) {
    sum += data[i]
  }
  const initialSMA = sum / period

  for (let i = 0; i < data.length; i++) {
    if (i < period - 1) {
      result.push(null)
    } else if (i === period - 1) {
      result.push(initialSMA)
    } else {
      const prevEMA = result[i - 1]
      result.push((data[i] - prevEMA) * multiplier + prevEMA)
    }
  }

  return result
}

const highs = [130, 132, 134, 133, 135, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150]
const lows = [120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139]
const closes = [125, 127, 129, 128, 130, 132, 133, 135, 137, 138, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149]
const period = 14

const trueRanges = []

// Calculate True Range for each period
for (let i = 1; i < closes.length; i++) {
  const high = highs[i]
  const low = lows[i]
  const prevClose = closes[i - 1]

  const tr = Math.max(
    high - low,
    Math.abs(high - prevClose),
    Math.abs(low - prevClose)
  )
  trueRanges.push(tr)
}

// ATR is the EMA of True Range
const atrValues = calculateEMA(trueRanges, period)

// Add initial null to align with original data length
const atr = [null, ...atrValues]

console.log('ATR Test Data Results:')
console.log('======================')

console.log('\nTrue Ranges (first 5):')
for (let i = 0; i < Math.min(5, trueRanges.length); i++) {
  console.log(`TR[${i}]:`, trueRanges[i].toFixed(2))
}

console.log('\nATR values (starting from first valid):')
for (let i = 14; i < atr.length; i++) {
  const val = atr[i]
  console.log(`Index ${i}: ATR=${val !== null ? val.toFixed(2) : 'null'}`)
}

console.log('\nFormatted for test data:')
console.log('atr:', JSON.stringify(atr.map(v => v !== null ? Number(v.toFixed(2)) : null)))
