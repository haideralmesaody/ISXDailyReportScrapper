export type BetterOpportunityKind = 'BETTER_BUY' | 'BETTER_SELL'

export interface BetterOpportunity {
  kind: BetterOpportunityKind
  deltaPct: number
}

export interface BetterOpportunityBacktestSummary {
  last_action?: 'BUY' | 'SELL' | null
  last_action_price?: number | null
  error?: string
}

export function numberOrNull(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : null
  }
  return null
}

export function computeBetterOpportunity(
  bt: BetterOpportunityBacktestSummary | undefined,
  currentPrice: number | null
): BetterOpportunity | null {
  if (!bt || bt.error) return null
  if (!bt.last_action || !bt.last_action_price) return null
  if (!currentPrice || !Number.isFinite(currentPrice)) return null

  const lastPrice = Number(bt.last_action_price)
  if (!Number.isFinite(lastPrice) || lastPrice <= 0) return null

  const lastAction = bt.last_action

  if (lastAction === 'SELL' && currentPrice > lastPrice) {
    const deltaPct = ((currentPrice - lastPrice) / lastPrice) * 100
    return { kind: 'BETTER_SELL', deltaPct }
  }

  if (lastAction === 'BUY' && currentPrice < lastPrice) {
    const deltaPct = ((currentPrice - lastPrice) / lastPrice) * 100
    return { kind: 'BETTER_BUY', deltaPct }
  }

  return null
}

