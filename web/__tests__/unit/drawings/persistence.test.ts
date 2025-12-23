import { normalizeDrawingsEnvelope } from '@/components/analysis/drawings/persistence'

describe('drawings persistence envelope normalization', () => {
  it('accepts raw shapes array (backward compatible)', () => {
    const shapes = [
      {
        id: '1',
        type: 'trendLine',
        points: [{ time: 1, price: 10 }],
        pane: 0,
        style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid' },
        locked: false,
        createdAt: 1,
        updatedAt: 1,
      },
    ]

    const envelope = normalizeDrawingsEnvelope(shapes as any, 'AAA')
    expect(envelope.version).toBe(1)
    expect(envelope.ticker).toBe('AAA')
    expect(envelope.shapes).toHaveLength(1)
    expect(envelope.shapes[0]?.id).toBe('1')
  })

  it('accepts versioned envelope object', () => {
    const raw = {
      version: 2,
      ticker: 'bbb',
      updatedAt: 123,
      shapes: [
        {
          id: 'shape-1',
          type: 'rectangle',
          points: [{ time: 10, price: 20 }, { time: 20, price: 10 }],
          pane: 0,
          style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid' },
          locked: true,
          createdAt: 1,
          updatedAt: 2,
        },
      ],
    }

    const envelope = normalizeDrawingsEnvelope(raw as any, 'FALLBACK')
    expect(envelope.version).toBe(2)
    expect(envelope.ticker).toBe('bbb')
    expect(envelope.updatedAt).toBe(123)
    expect(envelope.shapes[0]?.type).toBe('rectangle')
  })

  it('falls back to empty envelope on unknown payload', () => {
    expect(normalizeDrawingsEnvelope(null, 'AAA')).toEqual({ version: 1, ticker: 'AAA', shapes: [] })
    expect(normalizeDrawingsEnvelope({ hello: 'world' }, 'AAA')).toEqual({ version: 1, ticker: 'AAA', shapes: [] })
  })
})

