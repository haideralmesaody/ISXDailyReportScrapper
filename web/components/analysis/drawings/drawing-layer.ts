import type { IChartApi, ISeriesApi } from 'lightweight-charts'
import { pointToCoordinate } from './coordinates'
import { useDrawingStore } from './store'
import type { DrawingShape, DrawingStyle } from './types'
import { getShapeHandles } from './hit-tests'

interface BitmapRenderTarget {
  context: CanvasRenderingContext2D
  horizontalPixelRatio: number
  verticalPixelRatio: number
  bitmapSize: { width: number; height: number }
}

interface CanvasRenderingTarget2D {
  useBitmapCoordinateSpace: (cb: (target: BitmapRenderTarget) => void) => void
}

interface SeriesAttachedParameter {
  chart: IChartApi
  series: ISeriesApi<'Candlestick'>
  requestUpdate: () => void
}

interface ISeriesPrimitivePaneRenderer {
  draw(target: CanvasRenderingTarget2D): void
  drawBackground?(target: CanvasRenderingTarget2D): void
}

interface ISeriesPrimitivePaneView {
  update(): void
  renderer(): ISeriesPrimitivePaneRenderer
}

interface ISeriesPrimitive {
  attached?(param: SeriesAttachedParameter): void
  detached?(): void
  updateAllViews?(): void
  paneViews(): ISeriesPrimitivePaneView[]
}

interface DrawingSnapshot {
  shapes: DrawingShape[]
  selection: Set<string>
  hoveredId: string | null
  draft: DrawingShape | null
  chartIsDark: boolean
}

const buildSnapshot = (): DrawingSnapshot => {
  const state = useDrawingStore.getState()
  if (!state.layerVisible) {
    return {
      shapes: [],
      selection: new Set<string>(),
      hoveredId: null,
      draft: null,
      chartIsDark: state.chartIsDark,
    }
  }
  const shapes = state.order
    .map((id) => state.shapes[id])
    .filter((shape): shape is DrawingShape => Boolean(shape))
  return {
    shapes,
    selection: new Set(state.selection),
    hoveredId: state.hoveredId,
    draft: state.draft?.shape ?? null,
    chartIsDark: state.chartIsDark,
  }
}

const applyLineStyle = (ctx: CanvasRenderingContext2D, style: DrawingStyle): void => {
  switch (style.strokeStyle) {
    case 'dashed':
      ctx.setLineDash([6, 4])
      break
    case 'dotted':
      ctx.setLineDash([2, 4])
      break
    default:
      ctx.setLineDash([])
  }
}

const sanitizeStyle = (style: DrawingStyle, override?: Partial<DrawingStyle>): DrawingStyle => {
  return {
    ...style,
    ...override,
    opacity: override?.opacity ?? style.opacity ?? 1,
  }
}

class DrawingRenderer implements ISeriesPrimitivePaneRenderer {
  private chart: IChartApi
  private series: ISeriesApi<'Candlestick'>
  private snapshot: DrawingSnapshot | null = null

  constructor(chart: IChartApi, series: ISeriesApi<'Candlestick'>) {
    this.chart = chart
    this.series = series
  }

  setSnapshot(snapshot: DrawingSnapshot | null): void {
    this.snapshot = snapshot
  }

  draw(target: CanvasRenderingTarget2D): void {
    if (!this.snapshot) return

    target.useBitmapCoordinateSpace(({ context, horizontalPixelRatio, verticalPixelRatio, bitmapSize }) => {
      const scaleX = horizontalPixelRatio
      const scaleY = verticalPixelRatio
      const mediaWidth = bitmapSize.width / scaleX
      const mediaHeight = bitmapSize.height / scaleY

      const toX = (x: number): number => Math.round(x * scaleX)
      const toY = (y: number): number => Math.round(y * scaleY)

      const drawLine = (x1: number, y1: number, x2: number, y2: number, style: DrawingStyle): void => {
        context.save()
        context.strokeStyle = style.strokeColor
        context.globalAlpha = style.opacity ?? 1
        context.lineWidth = Math.max(1, Math.round(style.strokeWidth * scaleX))
        applyLineStyle(context, style)
        context.beginPath()
        context.moveTo(toX(x1), toY(y1))
        context.lineTo(toX(x2), toY(y2))
        context.stroke()
        context.restore()
      }

      const drawRectangle = (
        x1: number,
        y1: number,
        x2: number,
        y2: number,
        style: DrawingStyle
      ): void => {
        const left = Math.min(x1, x2)
        const right = Math.max(x1, x2)
        const top = Math.min(y1, y2)
        const bottom = Math.max(y1, y2)

        context.save()
        context.globalAlpha = style.opacity ?? 1
        if (style.fillColor) {
          context.fillStyle = style.fillColor
          context.fillRect(toX(left), toY(top), Math.abs(toX(right) - toX(left)), Math.abs(toY(bottom) - toY(top)))
        }

        context.strokeStyle = style.strokeColor
        context.lineWidth = Math.max(1, Math.round(style.strokeWidth * scaleX))
        applyLineStyle(context, style)
        context.strokeRect(toX(left), toY(top), Math.abs(toX(right) - toX(left)), Math.abs(toY(bottom) - toY(top)))
        context.restore()
      }

      const drawArrow = (
        x1: number,
        y1: number,
        x2: number,
        y2: number,
        style: DrawingStyle
      ): void => {
        drawLine(x1, y1, x2, y2, style)

        const angle = Math.atan2(y2 - y1, x2 - x1)
        const headLength = 8
        const headAngle = Math.PI / 7
        const leftX = x2 - headLength * Math.cos(angle - headAngle)
        const leftY = y2 - headLength * Math.sin(angle - headAngle)
        const rightX = x2 - headLength * Math.cos(angle + headAngle)
        const rightY = y2 - headLength * Math.sin(angle + headAngle)

        context.save()
        context.strokeStyle = style.strokeColor
        context.globalAlpha = style.opacity ?? 1
        context.lineWidth = Math.max(1, Math.round(style.strokeWidth * scaleX))
        applyLineStyle(context, style)
        context.beginPath()
        context.moveTo(toX(x2), toY(y2))
        context.lineTo(toX(leftX), toY(leftY))
        context.moveTo(toX(x2), toY(y2))
        context.lineTo(toX(rightX), toY(rightY))
        context.stroke()
        context.restore()
      }

      const drawText = (x: number, y: number, text: string, style: DrawingStyle): void => {
        context.save()
        context.globalAlpha = style.opacity ?? 1
        context.font = '12px sans-serif'
        context.textBaseline = 'middle'
        context.textAlign = 'left'
        const label = text || 'Text'
        const metrics = context.measureText(label)
        const paddingX = 6
        const paddingY = 4
        const boxWidth = metrics.width + paddingX * 2
        const boxHeight = 12 + paddingY * 2

        const boxLeft = toX(x) - paddingX
        const boxTop = toY(y) - boxHeight / 2

        context.fillStyle = this.snapshot?.chartIsDark ? 'rgba(0, 0, 0, 0.6)' : 'rgba(255, 255, 255, 0.85)'
        context.fillRect(boxLeft, boxTop, boxWidth, boxHeight)

        context.lineWidth = Math.max(1, Math.round(2 * scaleX))
        context.strokeStyle = this.snapshot?.chartIsDark ? 'rgba(0, 0, 0, 0.85)' : 'rgba(255, 255, 255, 0.95)'

        const candidate = (style.textColor ?? '').toLowerCase()
        const looksDark = candidate === '#111827' || candidate === '#000' || candidate === '#000000'
        context.fillStyle = this.snapshot?.chartIsDark && looksDark ? '#f9fafb' : (style.textColor ?? (this.snapshot?.chartIsDark ? '#f9fafb' : '#111827'))
        context.strokeText(label, toX(x), toY(y))
        context.fillText(label, toX(x), toY(y))
        context.restore()
      }

      const drawMeasureLabel = (x: number, y: number, label: string, style: DrawingStyle): void => {
        context.save()
        context.font = '12px sans-serif'
        context.textBaseline = 'middle'
        context.textAlign = 'center'
        const metrics = context.measureText(label)
        const paddingX = 6
        const paddingY = 4
        const boxWidth = metrics.width + paddingX * 2
        const boxHeight = 12 + paddingY * 2
        const boxLeft = toX(x) - boxWidth / 2
        const boxTop = toY(y) - boxHeight / 2
        context.globalAlpha = 0.85
        context.fillStyle = '#ffffff'
        context.fillRect(boxLeft, boxTop, boxWidth, boxHeight)
        context.globalAlpha = style.opacity ?? 1
        context.fillStyle = style.textColor ?? style.strokeColor
        context.fillText(label, toX(x), toY(y))
        context.restore()
      }

      const drawHandles = (shape: DrawingShape): void => {
        if (!this.snapshot?.selection.has(shape.id)) return
        const handles = getShapeHandles(shape, this.chart, this.series)
        if (handles.length === 0) return

        const size = 6
        const halfX = (size * scaleX) / 2
        const halfY = (size * scaleY) / 2

        context.save()
        context.fillStyle = '#ffffff'
        context.strokeStyle = '#111827'
        context.lineWidth = Math.max(1, Math.round(1 * scaleX))

        handles.forEach((handle) => {
          const cx = toX(handle.x)
          const cy = toY(handle.y)
          context.fillRect(cx - halfX, cy - halfY, halfX * 2, halfY * 2)
          context.strokeRect(cx - halfX, cy - halfY, halfX * 2, halfY * 2)
        })

        context.restore()
      }

      const drawShape = (shape: DrawingShape, draft = false): void => {
        const style = draft
          ? sanitizeStyle(shape.style, { opacity: 0.6, strokeStyle: 'dashed' })
          : sanitizeStyle(shape.style)
    const selected = this.snapshot?.selection.has(shape.id)
    const hovered = this.snapshot?.hoveredId === shape.id

    const renderHighlight = (x1: number, y1: number, x2: number, y2: number): void => {
          if ((!selected && !hovered) || draft) return
          drawLine(x1, y1, x2, y2, {
            ...style,
            strokeColor: selected ? '#f59e0b' : '#38bdf8',
            strokeWidth: style.strokeWidth + 1,
            opacity: 0.7,
            strokeStyle: 'solid',
          })
        }

        if (shape.type === 'horizontalLine') {
          const point = pointToCoordinate(this.chart, this.series, shape.points[0])
          if (!point) return
          renderHighlight(0, point.y, mediaWidth, point.y)
          drawLine(0, point.y, mediaWidth, point.y, style)
        } else if (shape.type === 'verticalLine') {
          const point = pointToCoordinate(this.chart, this.series, shape.points[0])
          if (!point) return
          renderHighlight(point.x, 0, point.x, mediaHeight)
          drawLine(point.x, 0, point.x, mediaHeight, style)
        } else if (shape.type === 'rectangle') {
          if (shape.points.length < 2) return
          const p1 = pointToCoordinate(this.chart, this.series, shape.points[0])
          const p2 = pointToCoordinate(this.chart, this.series, shape.points[1])
          if (!p1 || !p2) return
          drawRectangle(p1.x, p1.y, p2.x, p2.y, style)
        } else if (shape.type === 'text') {
          const point = pointToCoordinate(this.chart, this.series, shape.points[0])
          if (!point) return
          const textValue = 'text' in shape ? shape.text : 'Text'
          drawText(point.x, point.y, textValue, style)
        } else if (shape.type === 'ray') {
          if (shape.points.length < 2) return
          const p1 = pointToCoordinate(this.chart, this.series, shape.points[0])
          const p2 = pointToCoordinate(this.chart, this.series, shape.points[1])
          if (!p1 || !p2) return
          const dx = p2.x - p1.x
          const dy = p2.y - p1.y
          if (dx === 0) {
            renderHighlight(p1.x, 0, p1.x, mediaHeight)
            drawLine(p1.x, 0, p1.x, mediaHeight, style)
            return
          }
          const slope = dy / dx
          const xEnd = mediaWidth
          const yEnd = p1.y + slope * (xEnd - p1.x)
          renderHighlight(p1.x, p1.y, xEnd, yEnd)
          drawLine(p1.x, p1.y, xEnd, yEnd, style)
        } else if (shape.type === 'parallelChannel') {
          if (shape.points.length < 2) return
          const p1 = pointToCoordinate(this.chart, this.series, shape.points[0])
          const p2 = pointToCoordinate(this.chart, this.series, shape.points[1])
          if (!p1 || !p2) return

          // Draw base line
          renderHighlight(p1.x, p1.y, p2.x, p2.y)
          drawLine(p1.x, p1.y, p2.x, p2.y, style)

          if (shape.points.length < 3) return
          const p3 = pointToCoordinate(this.chart, this.series, shape.points[2])
          if (!p3) return

          const dx = p2.x - p1.x
          const dy = p2.y - p1.y
          const len = Math.hypot(dx, dy)
          if (len === 0) return

          // Normal vector (unit) to the base line
          const nx = -dy / len
          const ny = dx / len

          // Signed offset from p1 to p3 along the normal
          const offset = (p3.x - p1.x) * nx + (p3.y - p1.y) * ny

          const q1 = { x: p1.x + nx * offset, y: p1.y + ny * offset }
          const q2 = { x: p2.x + nx * offset, y: p2.y + ny * offset }

          // Fill channel area
          if (style.fillColor) {
            context.save()
            context.globalAlpha = style.opacity ?? 1
            context.fillStyle = style.fillColor
            context.beginPath()
            context.moveTo(toX(p1.x), toY(p1.y))
            context.lineTo(toX(p2.x), toY(p2.y))
            context.lineTo(toX(q2.x), toY(q2.y))
            context.lineTo(toX(q1.x), toY(q1.y))
            context.closePath()
            context.fill()
            context.restore()
          }

          // Draw parallel line
          renderHighlight(q1.x, q1.y, q2.x, q2.y)
          drawLine(q1.x, q1.y, q2.x, q2.y, style)
        } else {
          if (shape.points.length < 2) return
          const p1 = pointToCoordinate(this.chart, this.series, shape.points[0])
          const p2 = pointToCoordinate(this.chart, this.series, shape.points[1])
          if (!p1 || !p2) return

          if (shape.type === 'arrow') {
            renderHighlight(p1.x, p1.y, p2.x, p2.y)
            drawArrow(p1.x, p1.y, p2.x, p2.y, style)
          } else if (shape.type === 'measure') {
            renderHighlight(p1.x, p1.y, p2.x, p2.y)
            drawLine(p1.x, p1.y, p2.x, p2.y, style)
            const delta = p2.y - p1.y
            const priceDelta = shape.points[1].price - shape.points[0].price
            const percent = shape.points[0].price !== 0
              ? (priceDelta / shape.points[0].price) * 100
              : 0
            const bars = Math.round(Math.abs(shape.points[1].time - shape.points[0].time) / 86400)
            const label = `${priceDelta.toFixed(2)} (${percent.toFixed(2)}%) ${bars} bars`
            drawMeasureLabel(p1.x + (p2.x - p1.x) / 2, p1.y + delta / 2, label, style)
          } else {
            renderHighlight(p1.x, p1.y, p2.x, p2.y)
            drawLine(p1.x, p1.y, p2.x, p2.y, style)
          }
        }

        drawHandles(shape)
      }

      this.snapshot.shapes.forEach((shape) => drawShape(shape))
      if (this.snapshot.draft) {
        drawShape(this.snapshot.draft, true)
      }
    })
  }
}

class DrawingPaneView implements ISeriesPrimitivePaneView {
  private rendererInstance: DrawingRenderer
  private snapshot: DrawingSnapshot | null = null

  constructor(chart: IChartApi, series: ISeriesApi<'Candlestick'>) {
    this.rendererInstance = new DrawingRenderer(chart, series)
  }

  setSnapshot(snapshot: DrawingSnapshot | null): void {
    this.snapshot = snapshot
  }

  update(): void {
    this.rendererInstance.setSnapshot(this.snapshot)
  }

  renderer(): ISeriesPrimitivePaneRenderer {
    return this.rendererInstance
  }
}

export class DrawingLayerPrimitive implements ISeriesPrimitive {
  private chart: IChartApi
  private series: ISeriesApi<'Candlestick'>
  private paneView: DrawingPaneView
  private unsubscribe: (() => void) | null = null
  private requestUpdate: (() => void) | null = null
  private rafId: number | null = null

  constructor(chart: IChartApi, series: ISeriesApi<'Candlestick'>) {
    this.chart = chart
    this.series = series
    this.paneView = new DrawingPaneView(chart, series)
    this.paneView.setSnapshot(buildSnapshot())
  }

  attached(param: SeriesAttachedParameter): void {
    this.requestUpdate = param.requestUpdate
    this.unsubscribe = useDrawingStore.subscribe(
      (state) => ({
        layerVisible: state.layerVisible,
        chartIsDark: state.chartIsDark,
        shapes: state.shapes,
        order: state.order,
        selection: state.selection,
        hoveredId: state.hoveredId,
        draft: state.draft,
      }),
      () => {
        this.paneView.setSnapshot(buildSnapshot())
        if (!this.requestUpdate) return

        // Throttle primitive redraws to one per animation frame.
        // Without this, drag/move operations can trigger excessive repaints.
        if (this.rafId !== null) return
        this.rafId = requestAnimationFrame(() => {
          this.rafId = null
          this.requestUpdate?.()
        })
      }
    )
  }

  detached(): void {
    if (this.unsubscribe) {
      this.unsubscribe()
      this.unsubscribe = null
    }
    if (this.rafId !== null) {
      cancelAnimationFrame(this.rafId)
      this.rafId = null
    }
    this.requestUpdate = null
  }

  updateAllViews(): void {
    this.paneView.update()
  }

  paneViews(): ISeriesPrimitivePaneView[] {
    return [this.paneView]
  }
}
