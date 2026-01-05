import type { SeriesMarker, Time } from 'lightweight-charts'
import type { BacktestTrade } from '@/types/index'

export function buildTradeMarkers(
  trades: BacktestTrade[],
  tradableDates: Set<string>
): SeriesMarker<Time>[] {
  if (!trades || trades.length === 0 || tradableDates.size === 0) return []

  const markers: SeriesMarker<Time>[] = []

  for (let i = 0; i < trades.length; i++) {
    const trade = trades[i]

    const buyDate = trade.buy_date || trade.signal_buy_date
    if (buyDate && tradableDates.has(buyDate)) {
      markers.push({
        time: buyDate as unknown as Time,
        position: 'belowBar',
        color: '#10B981',
        shape: 'arrowUp',
        text: 'BUY',
      })
    }

    const sellDate = trade.sell_date || trade.signal_sell_date
    if (sellDate && tradableDates.has(sellDate)) {
      markers.push({
        time: sellDate as unknown as Time,
        position: 'aboveBar',
        color: '#EF4444',
        shape: 'arrowDown',
        text: 'SELL',
      })
    }
  }

  // Stable ordering on time (string dates sort lexicographically in YYYY-MM-DD form).
  markers.sort((a, b) => String(a.time).localeCompare(String(b.time)))

  return markers
}

