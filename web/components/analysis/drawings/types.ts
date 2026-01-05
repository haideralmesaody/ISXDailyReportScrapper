export type DrawingToolId =
  | 'select'
  | 'trend-line'
  | 'ray'
  | 'parallel-channel'
  | 'horizontal-line'
  | 'vertical-line'
  | 'rectangle'
  | 'text'
  | 'arrow'
  | 'measure'

export type DrawingShapeType =
  | 'trendLine'
  | 'ray'
  | 'parallelChannel'
  | 'horizontalLine'
  | 'verticalLine'
  | 'rectangle'
  | 'text'
  | 'arrow'
  | 'measure'

export type DrawingMode =
  | 'idle'
  | 'drawing'
  | 'selected'
  | 'editing'
  | 'moving'
  | 'resizing'
  | 'dragging'

export type DrawingCursor = 'default' | 'crosshair' | 'text' | 'move' | 'pointer'

export type DrawingRendererId =
  | 'trendLine'
  | 'ray'
  | 'parallelChannel'
  | 'horizontalLine'
  | 'verticalLine'
  | 'rectangle'
  | 'text'
  | 'arrow'
  | 'measure'

export type DrawingHitTestId = DrawingRendererId

export type DrawingMeta = Record<string, string | number | boolean | null>

export interface DrawingPoint {
  time: number
  price: number
}

export interface DrawingStyle {
  strokeColor: string
  strokeWidth: number
  strokeStyle: 'solid' | 'dashed' | 'dotted'
  fillColor?: string
  textColor?: string
  opacity?: number
}

export interface DrawingShapeBase {
  id: string
  type: DrawingShapeType
  points: DrawingPoint[]
  pane: number
  style: DrawingStyle
  meta?: DrawingMeta
  locked: boolean
  createdAt: number
  updatedAt: number
}

export interface TextDrawingShape extends DrawingShapeBase {
  type: 'text'
  text: string
}

export type DrawingShape = DrawingShapeBase | TextDrawingShape

export interface DrawingDraft {
  shape: DrawingShape
  requiredPoints: number
}

export interface DrawingToolDefinition {
  id: DrawingToolId
  label: string
  cursor: DrawingCursor
  shapeType?: DrawingShapeType
  requiredPoints?: number
  defaultStyle?: DrawingStyle
  renderer?: DrawingRendererId
  hitTest?: DrawingHitTestId
}
