import { computeBetterOpportunity, numberOrNull } from '@/lib/utils/better-opportunity'

describe('better-opportunity', () => {
  test('numberOrNull parses numbers and numeric strings', () => {
    expect(numberOrNull(1.23)).toBe(1.23)
    expect(numberOrNull('4.56')).toBe(4.56)
    expect(numberOrNull('bad')).toBeNull()
    expect(numberOrNull(null)).toBeNull()
  })

  test('returns Better Sell when last action SELL and current price higher', () => {
    const better = computeBetterOpportunity(
      { last_action: 'SELL', last_action_price: 10 },
      11
    )
    expect(better?.kind).toBe('BETTER_SELL')
    expect(better?.deltaPct).toBeCloseTo(10)
  })

  test('returns Better Buy when last action BUY and current price lower', () => {
    const better = computeBetterOpportunity(
      { last_action: 'BUY', last_action_price: 10 },
      9
    )
    expect(better?.kind).toBe('BETTER_BUY')
    expect(better?.deltaPct).toBeCloseTo(-10)
  })

  test('returns null when conditions do not match', () => {
    expect(computeBetterOpportunity({ last_action: 'SELL', last_action_price: 10 }, 9)).toBeNull()
    expect(computeBetterOpportunity({ last_action: 'BUY', last_action_price: 10 }, 11)).toBeNull()
  })

  test('returns null when data missing or invalid', () => {
    expect(computeBetterOpportunity(undefined, 10)).toBeNull()
    expect(computeBetterOpportunity({ error: 'oops', last_action: 'SELL', last_action_price: 10 }, 11)).toBeNull()
    expect(computeBetterOpportunity({ last_action: 'SELL', last_action_price: null }, 11)).toBeNull()
    expect(computeBetterOpportunity({ last_action: 'SELL', last_action_price: 0 }, 11)).toBeNull()
    expect(computeBetterOpportunity({ last_action: 'SELL', last_action_price: 10 }, null)).toBeNull()
  })
})

