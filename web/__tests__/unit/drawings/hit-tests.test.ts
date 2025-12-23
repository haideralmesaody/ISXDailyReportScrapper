import { getShapeHandles, hitTestShape } from '@/components/analysis/drawings/hit-tests'
import type { DrawingShape } from '@/components/analysis/drawings/types'

jest.mock('@/components/analysis/drawings/coordinates', () => ({
  pointToCoordinate: (_chart: any, _series: any, point: { time: number; price: number }) => ({
    x: point.time,
    y: point.price,
  }),
}))

describe('drawings hit testing', () => {
  const chart = {} as any
  const series = {} as any

  it('computes rectangle handles (corners)', () => {
    const shape: DrawingShape = {
      id: 'rect-1',
      type: 'rectangle',
      points: [
        { time: 0, price: 0 },
        { time: 10, price: 10 },
      ],
      pane: 0,
      style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid' },
      locked: false,
      createdAt: 1,
      updatedAt: 1,
    }

    const handles = getShapeHandles(shape, chart, series)
    expect(handles).toEqual([
      { x: 0, y: 0 },
      { x: 10, y: 0 },
      { x: 10, y: 10 },
      { x: 0, y: 10 },
    ])
  })

  it('returns handle hit when pointer is near a corner', () => {
    const shape: DrawingShape = {
      id: 'rect-2',
      type: 'rectangle',
      points: [
        { time: 0, price: 0 },
        { time: 10, price: 10 },
      ],
      pane: 0,
      style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid' },
      locked: false,
      createdAt: 1,
      updatedAt: 1,
    }

    const hit = hitTestShape(shape, chart, series, { x: 1, y: 1 }, 3)
    expect(hit).toEqual({ id: 'rect-2', type: 'rectangle', handleIndex: 0 })
  })

  it('hit-tests trend line segment', () => {
    const shape: DrawingShape = {
      id: 'line-1',
      type: 'trendLine',
      points: [
        { time: 0, price: 0 },
        { time: 10, price: 0 },
      ],
      pane: 0,
      style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid' },
      locked: false,
      createdAt: 1,
      updatedAt: 1,
    }

    expect(hitTestShape(shape, chart, series, { x: 5, y: 2 }, 3)).toEqual({ id: 'line-1', type: 'trendLine' })
    expect(hitTestShape(shape, chart, series, { x: 5, y: 10 }, 3)).toBeNull()
  })

  it('hit-tests text bounding box', () => {
    const shape: DrawingShape = {
      id: 'text-1',
      type: 'text',
      text: 'Hello',
      points: [{ time: 100, price: 100 }],
      pane: 0,
      style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid', textColor: '#fff' },
      locked: false,
      createdAt: 1,
      updatedAt: 1,
    }

    expect(hitTestShape(shape, chart, series, { x: 100, y: 100 }, 0)).toEqual({
      id: 'text-1',
      type: 'text',
      handleIndex: 0,
    })
    expect(hitTestShape(shape, chart, series, { x: 20, y: 20 }, 0)).toBeNull()
  })

  it('hit-tests parallel channel (lines + interior)', () => {
    const shape: DrawingShape = {
      id: 'channel-1',
      type: 'parallelChannel',
      points: [
        { time: 0, price: 0 },
        { time: 10, price: 0 },
        { time: 0, price: 10 },
      ],
      pane: 0,
      style: { strokeColor: '#000', strokeWidth: 1, strokeStyle: 'solid', fillColor: 'rgba(0,0,0,0.1)' },
      locked: false,
      createdAt: 1,
      updatedAt: 1,
    }

    // Inside the channel polygon
    expect(hitTestShape(shape, chart, series, { x: 5, y: 5 }, 0)).toEqual({ id: 'channel-1', type: 'parallelChannel' })

    // Near the base line
    expect(hitTestShape(shape, chart, series, { x: 5, y: 1 }, 2)).toEqual({ id: 'channel-1', type: 'parallelChannel' })

    // Handle hit near the third point
    expect(hitTestShape(shape, chart, series, { x: 0, y: 10 }, 0)).toEqual({
      id: 'channel-1',
      type: 'parallelChannel',
      handleIndex: 2,
    })
  })
})
