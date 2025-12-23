// Calculate Bollinger Bands expected values for test data

function calculateSMA(data, period) {
  if (!data || data.length === 0) return []
  if (period <= 0 || period > data.length) return data.map(() => null)

  const result = []
  for (let i = 0; i < data.length; i++) {
    if (i < period - 1) {
      result.push(null)
      continue
    }

    let sum = 0
    for (let j = 0; j < period; j++) {
      sum += data[i - j]
    }
    result.push(sum / period)
  }

  return result
}

const input = [100, 102, 101, 103, 104, 102, 103, 105, 104, 106, 107, 108, 110, 109, 111, 113, 112, 114, 115, 116]
const period = 10
const stdDevMultiplier = 2

// Calculate middle band (SMA)
const middle = calculateSMA(input, period)

const upper = []
const lower = []

// Calculate upper and lower bands
for (let i = 0; i < input.length; i++) {
  if (middle[i] === null || i < period - 1) {
    upper.push(null)
    lower.push(null)
    continue
  }

  // Calculate standard deviation for the period
  let sumSquaredDiff = 0
  for (let j = 0; j < period; j++) {
    const diff = input[i - j] - middle[i]
    sumSquaredDiff += diff * diff
  }
  const stdDev = Math.sqrt(sumSquaredDiff / period)

  upper.push(middle[i] + stdDevMultiplier * stdDev)
  lower.push(middle[i] - stdDevMultiplier * stdDev)
}

console.log('Bollinger Bands Test Data Results:')
console.log('===================================')

console.log('\nKey indices for verification:')
for (let i = 9; i < input.length; i++) {
  const m = middle[i]
  const u = upper[i]
  const l = lower[i]
  console.log(`Index ${i}: Middle=${m !== null ? m.toFixed(2) : 'null'}, Upper=${u !== null ? u.toFixed(2) : 'null'}, Lower=${l !== null ? l.toFixed(2) : 'null'}`)
}

console.log('\nFormatted for test data:')
console.log('middle:', JSON.stringify(middle.map(v => v !== null ? Number(v.toFixed(2)) : null)))
console.log('\nupper:', JSON.stringify(upper.map(v => v !== null ? Number(v.toFixed(2)) : null)))
console.log('\nlower:', JSON.stringify(lower.map(v => v !== null ? Number(v.toFixed(2)) : null)))
