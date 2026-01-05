import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import { v4 as uuidv4 } from 'uuid'
import type {
  DrawingDraft,
  DrawingMode,
  DrawingPoint,
  DrawingShape,
  DrawingStyle,
  DrawingToolDefinition,
  DrawingToolId,
} from './types'
import { DEFAULT_DRAWING_STYLE, PHASE_ONE_TOOLS } from './tools'

const TOOL_MAP = new Map(PHASE_ONE_TOOLS.map((tool) => [tool.id, tool]))
const LAST_TOOL_KEY = 'isx_drawings_last_tool'

const readLastTool = (): DrawingToolId => {
  if (typeof window === 'undefined') return 'select'
  try {
    const value = window.localStorage.getItem(LAST_TOOL_KEY)
    if (value && TOOL_MAP.has(value as DrawingToolId)) {
      return value as DrawingToolId
    }
  } catch {
    // ignore
  }
  return 'select'
}

const createShape = (tool: DrawingToolDefinition, point: DrawingPoint, chartIsDark: boolean): DrawingShape => {
  if (!tool.shapeType) {
    throw new Error('Cannot create shape without tool shapeType')
  }

  const shapeType = tool.shapeType
  const style: DrawingStyle = {
    ...(tool.defaultStyle ?? DEFAULT_DRAWING_STYLE),
  }
  const now = Date.now()
  const base = {
    id: uuidv4(),
    type: shapeType,
    points: [point],
    pane: 0,
    style,
    meta: { toolId: tool.id },
    locked: false,
    createdAt: now,
    updatedAt: now,
  }

  if (shapeType === 'text') {
    return {
      ...base,
      type: 'text',
      text: 'Text',
      style: {
        ...style,
        textColor: chartIsDark ? '#f9fafb' : '#111827',
      },
    }
  }

  return base
}

interface DrawingState {
  activeTool: DrawingToolId
  layerVisible: boolean
  magnetEnabled: boolean
  chartIsDark: boolean
  mode: DrawingMode
  draft: DrawingDraft | null
  shapes: Record<string, DrawingShape>
  order: string[]
  selection: string[]
  hoveredId: string | null
  dirty: boolean
  lastSavedAt: number | null
  history: {
    past: DrawingHistoryState[]
    future: DrawingHistoryState[]
  }
}

interface DrawingActions {
  setTool: (tool: DrawingToolId) => void
  toggleLayerVisible: () => void
  toggleMagnet: () => void
  setChartIsDark: (isDark: boolean) => void
  setMode: (mode: DrawingMode) => void
  startDraft: (tool: DrawingToolId, point: DrawingPoint) => void
  updateDraft: (point: DrawingPoint) => void
  commitDraft: () => void
  cancelDraft: () => void
  setSelection: (ids: string[]) => void
  clearSelection: () => void
  setHovered: (id: string | null) => void
  updateShape: (id: string, updater: (shape: DrawingShape) => DrawingShape) => void
  removeShape: (id: string) => void
  deleteSelection: () => void
  toggleLockSelection: () => void
  deleteActive: () => void
  toggleLockActive: () => void
  clearAll: () => void
  replaceAll: (shapes: DrawingShape[], markClean?: boolean) => void
  markSaved: () => void
  pushHistory: () => void
  undo: () => void
  redo: () => void
}

interface DrawingHistoryState {
  shapes: Record<string, DrawingShape>
  order: string[]
  selection: string[]
}

const cloneShape = (shape: DrawingShape): DrawingShape => {
  const base = {
    ...shape,
    points: shape.points.map((point) => ({ ...point })),
    style: { ...shape.style },
    meta: shape.meta ? { ...shape.meta } : undefined,
  }

  if (shape.type === 'text') {
    return {
      ...base,
      text: shape.text,
      type: 'text',
    }
  }

  return base
}

const snapshotHistory = (state: DrawingState): DrawingHistoryState => {
  const shapes: Record<string, DrawingShape> = {}
  Object.entries(state.shapes).forEach(([id, shape]) => {
    shapes[id] = cloneShape(shape)
  })

  return {
    shapes,
    order: [...state.order],
    selection: [...state.selection],
  }
}

const MAX_HISTORY = 50

export const useDrawingStore = create<DrawingState & DrawingActions>()(
  subscribeWithSelector((set, get) => ({
    activeTool: readLastTool(),
    layerVisible: true,
    magnetEnabled: false,
    chartIsDark: true,
    mode: 'idle',
    draft: null,
    shapes: {},
    order: [],
    selection: [],
    hoveredId: null,
    dirty: false,
    lastSavedAt: null,
    history: {
      past: [],
      future: [],
    },

    setTool: (tool) => {
      try {
        if (typeof window !== 'undefined') {
          window.localStorage.setItem(LAST_TOOL_KEY, tool)
        }
      } catch {
        // ignore
      }

      set((state) => {
        const next: Partial<DrawingState> = { activeTool: tool }
        if (state.mode === 'drawing') {
          next.draft = null
          next.mode = 'idle'
        }
        if (tool !== 'select') {
          next.selection = []
          if (state.mode !== 'drawing') {
            next.mode = 'idle'
          }
        }
        return next
      })
    },

    toggleLayerVisible: () =>
      set((state) => ({
        layerVisible: !state.layerVisible,
        mode: state.layerVisible ? 'idle' : state.mode,
        draft: state.layerVisible ? null : state.draft,
        selection: state.layerVisible ? [] : state.selection,
        hoveredId: state.layerVisible ? null : state.hoveredId,
      })),

    toggleMagnet: () => set((state) => ({ magnetEnabled: !state.magnetEnabled })),

    setChartIsDark: (isDark) => set({ chartIsDark: isDark }),

    setMode: (mode) => set({ mode }),

    startDraft: (tool, point) => {
      const def = TOOL_MAP.get(tool)
      if (!def || !def.shapeType || !def.requiredPoints) {
        return
      }

      const shape = createShape(def, point, get().chartIsDark)
      set({
        draft: { shape, requiredPoints: def.requiredPoints },
        mode: 'drawing',
        selection: [],
      })
    },

    updateDraft: (point) => {
      const { draft } = get()
      if (!draft) return

      const nextPoints = [...draft.shape.points]
      if (draft.requiredPoints === 1) {
        nextPoints[0] = point
      } else if (nextPoints.length === 1) {
        nextPoints.push(point)
      } else {
        nextPoints[nextPoints.length - 1] = point
      }

      set({
        draft: {
          ...draft,
          shape: {
            ...draft.shape,
            points: nextPoints,
            updatedAt: Date.now(),
          },
        },
      })
    },

    commitDraft: () => {
      const { draft, shapes, order } = get()
      if (!draft) return

      if (draft.shape.points.length < draft.requiredPoints) {
        // Multi-point tools (e.g., parallel channel) are placed via multiple clicks.
        // Each click "locks in" the current preview point and adds a new preview point.
        if (draft.shape.points.length === 0) return
        const last = draft.shape.points[draft.shape.points.length - 1]
        set({
          draft: {
            ...draft,
            shape: {
              ...draft.shape,
              points: [...draft.shape.points, { ...last }],
              updatedAt: Date.now(),
            },
          },
          mode: 'drawing',
        })
        return
      }

      get().pushHistory()

      const nextShapes = { ...shapes, [draft.shape.id]: draft.shape }
      const nextOrder = [...order, draft.shape.id]

      set({
        draft: null,
        mode: 'selected',
        shapes: nextShapes,
        order: nextOrder,
        selection: [draft.shape.id],
        dirty: true,
      })
    },

    cancelDraft: () => set({ draft: null, mode: 'idle' }),

    setSelection: (ids) => set({ selection: ids, mode: ids.length ? 'selected' : 'idle' }),

    clearSelection: () => set({ selection: [], mode: 'idle' }),

    setHovered: (id) => set({ hoveredId: id }),

    updateShape: (id, updater) => {
      const { shapes } = get()
      const current = shapes[id]
      if (!current) return

      const updated = updater(current)
      set({
        shapes: { ...shapes, [id]: { ...updated, updatedAt: Date.now() } },
        dirty: true,
      })
    },

    removeShape: (id) => {
      const { shapes, order, selection } = get()
      if (!shapes[id]) return

      get().pushHistory()
      const nextShapes = { ...shapes }
      delete nextShapes[id]
      const nextSelection = selection.filter((item) => item !== id)

      set({
        shapes: nextShapes,
        order: order.filter((item) => item !== id),
        selection: nextSelection,
        mode: nextSelection.length ? 'selected' : 'idle',
        dirty: true,
      })
    },

    deleteSelection: () => {
      const { selection, shapes, order } = get()
      if (selection.length === 0) return

      get().pushHistory()
      const nextShapes = { ...shapes }
      selection.forEach((id) => delete nextShapes[id])

      set({
        shapes: nextShapes,
        order: order.filter((id) => !selection.includes(id)),
        selection: [],
        mode: 'idle',
        dirty: true,
      })
    },

    toggleLockSelection: () => {
      const { selection, shapes } = get()
      if (selection.length === 0) return

      get().pushHistory()
      const nextShapes = { ...shapes }
      selection.forEach((id) => {
        const shape = nextShapes[id]
        if (!shape) return
        nextShapes[id] = { ...shape, locked: !shape.locked, updatedAt: Date.now() }
      })

      set({
        shapes: nextShapes,
        dirty: true,
      })
    },

    deleteActive: () => {
      const { selection, hoveredId } = get()
      if (selection.length > 0) {
        get().deleteSelection()
        return
      }
      if (hoveredId) {
        get().removeShape(hoveredId)
      }
    },

    toggleLockActive: () => {
      const { selection, hoveredId } = get()
      if (selection.length > 0) {
        get().toggleLockSelection()
        return
      }
      if (!hoveredId) return
      get().pushHistory()
      get().updateShape(hoveredId, (shape) => ({ ...shape, locked: !shape.locked }))
    },

    clearAll: () => {
      get().pushHistory()
      set({ shapes: {}, order: [], selection: [], draft: null, mode: 'idle', dirty: true })
    },

    replaceAll: (nextShapesList, markClean = true) => {
      const history = markClean ? { past: [], future: [] } : get().history
      const nextShapes: Record<string, DrawingShape> = {}
      const nextOrder: string[] = []

      nextShapesList.forEach((shape) => {
        const pane = Number.isFinite(shape.pane) ? shape.pane : 0
        nextShapes[shape.id] = { ...shape, pane }
        nextOrder.push(shape.id)
      })

      set({
        shapes: nextShapes,
        order: nextOrder,
        selection: [],
        draft: null,
        mode: 'idle',
        dirty: !markClean,
        lastSavedAt: markClean ? Date.now() : get().lastSavedAt,
        history,
      })
    },

    markSaved: () => set({ dirty: false, lastSavedAt: Date.now() }),

    pushHistory: () => {
      const state = get()
      const nextPast = [...state.history.past, snapshotHistory(state)]
      const trimmed = nextPast.length > MAX_HISTORY ? nextPast.slice(nextPast.length - MAX_HISTORY) : nextPast
      set({
        history: {
          past: trimmed,
          future: [],
        },
      })
    },

    undo: () => {
      const state = get()
      const past = [...state.history.past]
      if (past.length === 0) return

      const previous = past.pop()
      if (!previous) return

      const future = [...state.history.future, snapshotHistory(state)]
      set({
        shapes: previous.shapes,
        order: previous.order,
        selection: previous.selection,
        draft: null,
        mode: previous.selection.length ? 'selected' : 'idle',
        dirty: true,
        history: { past, future },
      })
    },

    redo: () => {
      const state = get()
      const future = [...state.history.future]
      if (future.length === 0) return

      const next = future.pop()
      if (!next) return

      const past = [...state.history.past, snapshotHistory(state)]
      set({
        shapes: next.shapes,
        order: next.order,
        selection: next.selection,
        draft: null,
        mode: next.selection.length ? 'selected' : 'idle',
        dirty: true,
        history: { past, future },
      })
    },
  }))
)
