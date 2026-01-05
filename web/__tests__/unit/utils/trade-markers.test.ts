import { buildTradeMarkers } from '@/lib/utils/trade-markers'

describe('buildTradeMarkers', () => {
  it('builds BUY/SELL markers and filters out non-tradable dates', () => {
    const trades: any[] = [
      {
        status: 'CLOSED',
        buy_date: '2025-01-02',
        sell_date: '2025-01-05', // non-tradable (not in set)
      },
      {
        status: 'CLOSED',
        signal_buy_date: '2025-01-03',
        signal_sell_date: '2025-01-04',
      },
    ]

    const tradableDates = new Set(['2025-01-02', '2025-01-03', '2025-01-04'])
    const markers = buildTradeMarkers(trades as any, tradableDates)

    expect(markers).toHaveLength(3)
    expect(markers.map((m) => m.time)).toEqual(['2025-01-02', '2025-01-03', '2025-01-04'])
    expect(markers.map((m) => m.text)).toEqual(['BUY', 'BUY', 'SELL'])
  })
})

