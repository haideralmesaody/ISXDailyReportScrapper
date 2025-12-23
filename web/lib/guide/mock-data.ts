import { CandlestickData, Time } from 'lightweight-charts'

/**
 * Generate mock stock data for demos
 * Creates realistic candlestick data with configurable parameters
 */
export function generateMockStockData(
  days: number = 100,
  startPrice: number = 5000,
  volatility: number = 200
): CandlestickData[] {
  const data: CandlestickData[] = []
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - days)

  let currentPrice = startPrice
  let trend = 0

  for (let i = 0; i < days; i++) {
    const date = new Date(startDate)
    date.setDate(date.getDate() + i)

    // Skip weekends
    const dayOfWeek = date.getDay()
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      continue
    }

    // Generate random trend changes
    if (Math.random() < 0.05) {
      trend = (Math.random() - 0.5) * volatility * 0.5
    }

    // Calculate OHLC values
    const randomChange = (Math.random() - 0.5) * volatility
    const priceChange = trend + randomChange

    currentPrice += priceChange

    // Ensure price doesn't go negative
    if (currentPrice < 100) {
      currentPrice = 100 + Math.random() * 500
    }

    const open = currentPrice
    const close = currentPrice + (Math.random() - 0.5) * volatility * 0.3
    const high = Math.max(open, close) + Math.random() * volatility * 0.2
    const low = Math.min(open, close) - Math.random() * volatility * 0.2

    // Format date as YYYY-MM-DD for TradingView
    const timeStr = date.toISOString().split('T')[0] as Time

    data.push({
      time: timeStr,
      open: Math.round(open),
      high: Math.round(high),
      low: Math.round(low),
      close: Math.round(close),
    })

    currentPrice = close
  }

  return data
}

/**
 * Generate mock volume data for demos
 */
export function generateMockVolumeData(
  candlestickData: CandlestickData[]
): { time: Time; value: number; color?: string }[] {
  return candlestickData.map((candle) => {
    const isGreen = candle.close >= candle.open
    const baseVolume = 1000000 + Math.random() * 5000000
    const volatilityMultiplier = Math.abs(candle.close - candle.open) / candle.open
    const volume = Math.round(baseVolume * (1 + volatilityMultiplier * 10))

    return {
      time: candle.time,
      value: volume,
      color: isGreen ? '#10b98180' : '#ef444480',
    }
  })
}

/**
 * Generate trending data (uptrend or downtrend)
 */
export function generateTrendingData(
  days: number = 60,
  startPrice: number = 5000,
  trendDirection: 'up' | 'down' = 'up'
): CandlestickData[] {
  const data: CandlestickData[] = []
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - days)

  let currentPrice = startPrice
  const trendStrength = trendDirection === 'up' ? 50 : -50
  const volatility = 150

  for (let i = 0; i < days; i++) {
    const date = new Date(startDate)
    date.setDate(date.getDate() + i)

    // Skip weekends
    const dayOfWeek = date.getDay()
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      continue
    }

    // Add trend and some noise
    const priceChange = trendStrength + (Math.random() - 0.5) * volatility

    currentPrice += priceChange

    // Ensure price doesn't go negative
    if (currentPrice < 100) {
      currentPrice = 100
    }

    const open = currentPrice
    const close = currentPrice + (Math.random() - 0.5) * volatility * 0.3
    const high = Math.max(open, close) + Math.random() * volatility * 0.2
    const low = Math.min(open, close) - Math.random() * volatility * 0.2

    const timeStr = date.toISOString().split('T')[0] as Time

    data.push({
      time: timeStr,
      open: Math.round(open),
      high: Math.round(high),
      low: Math.round(low),
      close: Math.round(close),
    })

    currentPrice = close
  }

  return data
}

/**
 * Generate ranging (sideways) market data
 */
export function generateRangingData(
  days: number = 60,
  basePrice: number = 5000,
  rangeSize: number = 500
): CandlestickData[] {
  const data: CandlestickData[] = []
  const startDate = new Date()
  startDate.setDate(startDate.getDate() - days)

  const upperBound = basePrice + rangeSize / 2
  const lowerBound = basePrice - rangeSize / 2

  let currentPrice = basePrice

  for (let i = 0; i < days; i++) {
    const date = new Date(startDate)
    date.setDate(date.getDate() + i)

    // Skip weekends
    const dayOfWeek = date.getDay()
    if (dayOfWeek === 0 || dayOfWeek === 6) {
      continue
    }

    // Oscillate within range
    const distanceFromCenter = currentPrice - basePrice
    const meanReversion = -distanceFromCenter * 0.2 // Pull towards center
    const noise = (Math.random() - 0.5) * rangeSize * 0.3

    currentPrice += meanReversion + noise

    // Keep within bounds
    if (currentPrice > upperBound) currentPrice = upperBound - Math.random() * 100
    if (currentPrice < lowerBound) currentPrice = lowerBound + Math.random() * 100

    const open = currentPrice
    const close = currentPrice + (Math.random() - 0.5) * rangeSize * 0.2
    const high = Math.max(open, close) + Math.random() * rangeSize * 0.1
    const low = Math.min(open, close) - Math.random() * rangeSize * 0.1

    const timeStr = date.toISOString().split('T')[0] as Time

    data.push({
      time: timeStr,
      open: Math.round(open),
      high: Math.round(high),
      low: Math.round(low),
      close: Math.round(close),
    })

    currentPrice = close
  }

  return data
}

/**
 * Sample ISX stock symbols for demo purposes
 */
export const sampleISXStocks = [
  { symbol: 'BMFI', name: 'Bank of Baghdad', sector: 'Banking' },
  { symbol: 'BNOI', name: 'Al-Mansour Bank', sector: 'Banking' },
  { symbol: 'TASC', name: 'AL-Taif Islamic Bank', sector: 'Banking' },
  { symbol: 'VTEL', name: 'Asiacell Communications', sector: 'Telecommunications' },
  { symbol: 'IHLI', name: 'Iraqi for Seed Production', sector: 'Agriculture' },
  { symbol: 'SMFI', name: 'AL-Mashreq Al-Arabi Bank', sector: 'Banking' },
  { symbol: 'BAME', name: 'Babil Bank', sector: 'Banking' },
  { symbol: 'BIPI', name: 'Baghdad for General Insurance', sector: 'Insurance' },
]
