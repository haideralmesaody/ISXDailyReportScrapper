'use client'

import type { DrawingShape } from './types'

export type DrawingsEnvelope = {
  version: number
  ticker: string
  updatedAt?: number
  shapes: DrawingShape[]
}

export const normalizeDrawingsEnvelope = (raw: unknown, fallbackTicker: string): DrawingsEnvelope => {
  if (Array.isArray(raw)) {
    return { version: 1, ticker: fallbackTicker, shapes: raw as DrawingShape[] }
  }

  if (raw && typeof raw === 'object') {
    const candidate = raw as { version?: unknown; ticker?: unknown; updatedAt?: unknown; shapes?: unknown }

    if (Array.isArray(candidate.shapes)) {
      return {
        version: typeof candidate.version === 'number' && Number.isFinite(candidate.version) ? candidate.version : 1,
        ticker: typeof candidate.ticker === 'string' ? candidate.ticker : fallbackTicker,
        updatedAt: typeof candidate.updatedAt === 'number' && Number.isFinite(candidate.updatedAt) ? candidate.updatedAt : undefined,
        shapes: candidate.shapes as DrawingShape[],
      }
    }
  }

  return { version: 1, ticker: fallbackTicker, shapes: [] }
}

