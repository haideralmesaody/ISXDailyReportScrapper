import type { IChartApi, ISeriesApi } from 'lightweight-charts'
import type { DrawingPoint, DrawingShape, DrawingShapeType } from './types'
import { pointToCoordinate } from './coordinates'

export interface DrawingCoordinate {
  x: number
  y: number
}

export interface DrawingHitTestResult {
  id: string
  type: DrawingShapeType
  handleIndex?: number
}

const DEFAULT_TOLERANCE = 6

const distance = (a: DrawingCoordinate, b: DrawingCoordinate): number => {
  const dx = a.x - b.x
  const dy = a.y - b.y
  return Math.sqrt(dx * dx + dy * dy)
}

const distanceToSegment = (p: DrawingCoordinate, a: DrawingCoordinate, b: DrawingCoordinate): number => {
  const dx = b.x - a.x
  const dy = b.y - a.y
  if (dx === 0 && dy === 0) return distance(p, a)

  const t = ((p.x - a.x) * dx + (p.y - a.y) * dy) / (dx * dx + dy * dy)
  if (t <= 0) return distance(p, a)
  if (t >= 1) return distance(p, b)

  const proj = { x: a.x + t * dx, y: a.y + t * dy }
  return distance(p, proj)
}

const distanceToRay = (p: DrawingCoordinate, a: DrawingCoordinate, b: DrawingCoordinate): number => {
  const dx = b.x - a.x
  const dy = b.y - a.y
  if (dx === 0 && dy === 0) return distance(p, a)

  const t = ((p.x - a.x) * dx + (p.y - a.y) * dy) / (dx * dx + dy * dy)
  if (t <= 0) return distance(p, a)

  const proj = { x: a.x + t * dx, y: a.y + t * dy }
  return distance(p, proj)
}

const rectangleEdges = (
  p1: DrawingCoordinate,
  p2: DrawingCoordinate
): [DrawingCoordinate, DrawingCoordinate][] => {
  const left = Math.min(p1.x, p2.x)
  const right = Math.max(p1.x, p2.x)
  const top = Math.min(p1.y, p2.y)
  const bottom = Math.max(p1.y, p2.y)

  return [
    [{ x: left, y: top }, { x: right, y: top }],
    [{ x: right, y: top }, { x: right, y: bottom }],
    [{ x: right, y: bottom }, { x: left, y: bottom }],
    [{ x: left, y: bottom }, { x: left, y: top }],
  ]
}

const textBounds = (anchor: DrawingCoordinate, text: string): { left: number; right: number; top: number; bottom: number } => {
  const fontSize = 12
  const padding = 4
  const width = text.length * fontSize * 0.6 + padding * 2
  const height = fontSize + padding * 2
  return {
    left: anchor.x,
    right: anchor.x + width,
    top: anchor.y - height / 2,
    bottom: anchor.y + height / 2,
  }
}

const sign = (value: number): number => (value < 0 ? -1 : value > 0 ? 1 : 0)

const cross = (a: DrawingCoordinate, b: DrawingCoordinate, c: DrawingCoordinate): number => {
  // Cross product of AB x AC
  return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x)
}

const pointInConvexQuad = (
  p: DrawingCoordinate,
  a: DrawingCoordinate,
  b: DrawingCoordinate,
  c: DrawingCoordinate,
  d: DrawingCoordinate
): boolean => {
  const s1 = sign(cross(a, b, p))
  const s2 = sign(cross(b, c, p))
  const s3 = sign(cross(c, d, p))
  const s4 = sign(cross(d, a, p))

  // Allow points on the boundary (0).
  const signs = [s1, s2, s3, s4].filter((s) => s !== 0)
  if (signs.length === 0) return true
  return signs.every((s) => s === signs[0])
}

const asCoordinate = (
  chart: IChartApi,
  series: ISeriesApi<'Candlestick'>,
  point: DrawingPoint
): DrawingCoordinate | null => {
  return pointToCoordinate(chart, series, point)
}

export const getShapeHandles = (
  shape: DrawingShape,
  chart: IChartApi,
  series: ISeriesApi<'Candlestick'>
): DrawingCoordinate[] => {
  if (shape.points.length === 0) return []

  if (shape.type === 'rectangle' && shape.points.length >= 2) {
    const p1 = asCoordinate(chart, series, shape.points[0])
    const p2 = asCoordinate(chart, series, shape.points[1])
    if (!p1 || !p2) return []

    const left = Math.min(p1.x, p2.x)
    const right = Math.max(p1.x, p2.x)
    const top = Math.min(p1.y, p2.y)
    const bottom = Math.max(p1.y, p2.y)

    return [
      { x: left, y: top },
      { x: right, y: top },
      { x: right, y: bottom },
      { x: left, y: bottom },
    ]
  }

  return shape.points
    .map((point) => asCoordinate(chart, series, point))
    .filter((point): point is DrawingCoordinate => point !== null)
}

export const hitTestShape = (
  shape: DrawingShape,
  chart: IChartApi,
  series: ISeriesApi<'Candlestick'>,
  point: DrawingCoordinate,
  tolerance: number = DEFAULT_TOLERANCE
): DrawingHitTestResult | null => {
  const handles = getShapeHandles(shape, chart, series)
  for (let i = 0; i < handles.length; i += 1) {
    if (distance(point, handles[i]) <= tolerance) {
      return { id: shape.id, type: shape.type, handleIndex: i }
    }
  }

  if (shape.points.length === 0) return null

  const p1 = asCoordinate(chart, series, shape.points[0])
  if (!p1) return null

  switch (shape.type) {
    case 'horizontalLine': {
      return Math.abs(point.y - p1.y) <= tolerance ? { id: shape.id, type: shape.type } : null
    }
    case 'verticalLine': {
      return Math.abs(point.x - p1.x) <= tolerance ? { id: shape.id, type: shape.type } : null
    }
    case 'trendLine':
    case 'arrow':
    case 'measure': {
      if (shape.points.length < 2) return null
      const p2 = asCoordinate(chart, series, shape.points[1])
      if (!p2) return null
      return distanceToSegment(point, p1, p2) <= tolerance ? { id: shape.id, type: shape.type } : null
    }
    case 'ray': {
      if (shape.points.length < 2) return null
      const p2 = asCoordinate(chart, series, shape.points[1])
      if (!p2) return null
      return distanceToRay(point, p1, p2) <= tolerance ? { id: shape.id, type: shape.type } : null
    }
    case 'parallelChannel': {
      if (shape.points.length < 2) return null
      const p2 = asCoordinate(chart, series, shape.points[1])
      if (!p2) return null

      // Partial placement: behave like a trend line until the 3rd point exists.
      if (shape.points.length < 3) {
        return distanceToSegment(point, p1, p2) <= tolerance ? { id: shape.id, type: shape.type } : null
      }

      const p3 = asCoordinate(chart, series, shape.points[2])
      if (!p3) return null

      const dx = p2.x - p1.x
      const dy = p2.y - p1.y
      const len = Math.hypot(dx, dy)
      if (len === 0) return null

      const nx = -dy / len
      const ny = dx / len
      const offset = (p3.x - p1.x) * nx + (p3.y - p1.y) * ny

      const q1 = { x: p1.x + nx * offset, y: p1.y + ny * offset }
      const q2 = { x: p2.x + nx * offset, y: p2.y + ny * offset }

      const nearBase = distanceToSegment(point, p1, p2) <= tolerance
      const nearParallel = distanceToSegment(point, q1, q2) <= tolerance
      const inside = pointInConvexQuad(point, p1, p2, q2, q1)

      return nearBase || nearParallel || inside ? { id: shape.id, type: shape.type } : null
    }
    case 'rectangle': {
      if (shape.points.length < 2) return null
      const p2 = asCoordinate(chart, series, shape.points[1])
      if (!p2) return null
      const edges = rectangleEdges(p1, p2)
      for (const [a, b] of edges) {
        if (distanceToSegment(point, a, b) <= tolerance) {
          return { id: shape.id, type: shape.type }
        }
      }
      return null
    }
    case 'text': {
      const bounds = textBounds(p1, 'text' in shape ? shape.text : 'Text')
      if (
        point.x >= bounds.left - tolerance &&
        point.x <= bounds.right + tolerance &&
        point.y >= bounds.top - tolerance &&
        point.y <= bounds.bottom + tolerance
      ) {
        return { id: shape.id, type: shape.type }
      }
      return null
    }
    default:
      return null
  }
}
