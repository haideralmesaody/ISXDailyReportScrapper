'use client'

import { useEffect, useRef } from 'react'
import { useChart } from '@/components/analysis/chart/ChartContext'
import { coordinateToPoint, pointToCoordinate } from './coordinates'
import { hitTestShape } from './hit-tests'
import { useDrawingStore } from './store'
import { DEFAULT_DRAWING_STYLE, PHASE_ONE_TOOLS } from './tools'
import type { DrawingPoint, DrawingShapeType } from './types'

const MOVE_THRESHOLD = 3
const DEBUG_SEED_KEY = 'isx_drawings_seed'

type DragMode = 'none' | 'drawing' | 'moving' | 'resizing'

interface DragState {
  mode: DragMode
  shapeId?: string
  shapeType?: DrawingShapeType
  handleIndex?: number
  originPoint?: DrawingPoint
  originCoordinate?: { x: number; y: number }
  originalPoints?: DrawingPoint[]
  hasMoved?: boolean
  pointerId?: number
}

const isEditableTarget = (target: EventTarget | null): boolean => {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName.toLowerCase()
  return tag === 'input' || tag === 'textarea' || target.isContentEditable
}

const findNearestCandleIndex = (times: number[], time: number): number => {
  if (times.length === 0) return -1
  let lo = 0
  let hi = times.length - 1
  while (lo <= hi) {
    const mid = Math.floor((lo + hi) / 2)
    const t = times[mid]
    if (t === time) return mid
    if (t < time) lo = mid + 1
    else hi = mid - 1
  }
  if (hi < 0) return 0
  if (lo >= times.length) return times.length - 1
  return Math.abs(times[lo] - time) < Math.abs(times[hi] - time) ? lo : hi
}

export function DrawingInteraction() {
  const { chart, candlestickSeries, containerRef, isReady, chartData } = useChart()
  const setMode = useDrawingStore((state) => state.setMode)
  const setSelection = useDrawingStore((state) => state.setSelection)
  const clearSelection = useDrawingStore((state) => state.clearSelection)
  const setHovered = useDrawingStore((state) => state.setHovered)
  const updateDraft = useDrawingStore((state) => state.updateDraft)
  const commitDraft = useDrawingStore((state) => state.commitDraft)
  const cancelDraft = useDrawingStore((state) => state.cancelDraft)
  const updateShape = useDrawingStore((state) => state.updateShape)
  const pushHistory = useDrawingStore((state) => state.pushHistory)
  const deleteSelection = useDrawingStore((state) => state.deleteSelection)
  const undo = useDrawingStore((state) => state.undo)
  const redo = useDrawingStore((state) => state.redo)
  const layerVisible = useDrawingStore((state) => state.layerVisible)
  const magnetEnabled = useDrawingStore((state) => state.magnetEnabled)

  const dragRef = useRef<DragState>({ mode: 'none' })
  const rafRef = useRef<number | null>(null)
  const pendingRef = useRef<{ x: number; y: number } | null>(null)
  const interactionDisabledRef = useRef(false)

  useEffect(() => {
    if (!chart || !candlestickSeries || !containerRef.current || !isReady) return

    const container = containerRef.current
    const candles = chartData.candlestickData
    const candleTimes = candles.map((c) => c.time as number)

    const toolCursorMap = new Map(PHASE_ONE_TOOLS.map((tool) => [tool.id, tool.cursor]))

    const setChartInteractionEnabled = (enabled: boolean) => {
      if (interactionDisabledRef.current === !enabled) return
      interactionDisabledRef.current = !enabled
      chart.applyOptions({
        handleScroll: {
          mouseWheel: enabled,
          pressedMouseMove: enabled,
          horzTouchDrag: enabled,
          vertTouchDrag: enabled,
        },
        handleScale: {
          axisPressedMouseMove: {
            time: enabled,
            price: enabled,
          },
          mouseWheel: enabled,
          pinch: enabled,
        },
      })
    }

    const updateCursor = () => {
      const state = useDrawingStore.getState()
      if (!state.layerVisible) {
        container.style.cursor = 'default'
        return
      }
      if (state.mode === 'moving' || state.mode === 'dragging') {
        container.style.cursor = 'move'
        return
      }
      if (state.mode === 'resizing') {
        container.style.cursor = 'pointer'
        return
      }

      const cursor = toolCursorMap.get(state.activeTool) ?? 'default'
      container.style.cursor = cursor
    }

    const unsubscribeCursor = useDrawingStore.subscribe(
      (state) => ({ activeTool: state.activeTool, mode: state.mode, layerVisible: state.layerVisible }),
      updateCursor,
      { fireImmediately: true }
    )

    const toCoordinate = (event: PointerEvent): { x: number; y: number } | null => {
      const rect = container.getBoundingClientRect()
      const x = event.clientX - rect.left
      const y = event.clientY - rect.top
      if (Number.isNaN(x) || Number.isNaN(y)) return null
      return { x, y }
    }

    const toPoint = (coordinate: { x: number; y: number }): DrawingPoint | null => {
      if (!chart || !candlestickSeries) return null
      const point = coordinateToPoint(chart, candlestickSeries, coordinate)
      if (!point) return null

      if (!magnetEnabled || candleTimes.length === 0) return point

      const index = findNearestCandleIndex(candleTimes, point.time)
      const candle = index >= 0 ? candles[index] : null
      if (!candle) return point

      const prices = [candle.open, candle.high, candle.low, candle.close]
      let snapped = point.price
      let best = Number.POSITIVE_INFINITY
      for (const price of prices) {
        const delta = Math.abs(price - point.price)
        if (delta < best) {
          best = delta
          snapped = price
        }
      }

      return { ...point, price: snapped }
    }

    const findHit = (coordinate: { x: number; y: number }) => {
      const state = useDrawingStore.getState()
      if (!state.layerVisible) return null
      for (let i = state.order.length - 1; i >= 0; i -= 1) {
        const shapeId = state.order[i]
        const shape = state.shapes[shapeId]
        if (!shape) continue
        const hit = hitTestShape(shape, chart, candlestickSeries, coordinate)
        if (hit) {
          return { hit, shape }
        }
      }
      return null
    }

    const maybeSeedDrawings = () => {
      try {
        if (localStorage.getItem(DEBUG_SEED_KEY) !== '1') return
      } catch {
        return
      }

      const state = useDrawingStore.getState()
      if (state.order.length > 0) return
      if (!chartData.candlestickData.length) return

      const data = chartData.candlestickData
      const last = data[data.length - 1]
      const start = data[Math.max(0, data.length - 20)]
      const mid = data[Math.max(0, data.length - 10)]
      if (!last || !start || !mid) return

      const now = Date.now()
      const newId = () => {
        const cryptoObj = globalThis.crypto as undefined | { randomUUID?: () => string }
        return cryptoObj?.randomUUID?.() ?? `${now}-${Math.random().toString(16).slice(2)}`
      }

      const base = {
        style: { ...DEFAULT_DRAWING_STYLE },
        locked: false,
        createdAt: now,
        updatedAt: now,
        pane: 0,
      }

      const seeded = [
        {
          ...base,
          id: newId(),
          type: 'trendLine' as const,
          points: [
            { time: start.time, price: start.low },
            { time: last.time, price: last.high },
          ],
        },
        {
          ...base,
          id: newId(),
          type: 'horizontalLine' as const,
          points: [{ time: last.time, price: last.close }],
        },
        {
          ...base,
          id: newId(),
          type: 'verticalLine' as const,
          points: [{ time: mid.time, price: mid.close }],
        },
        {
          ...base,
          id: newId(),
          type: 'rectangle' as const,
          points: [
            { time: start.time, price: start.low },
            { time: mid.time, price: start.high },
          ],
        },
        {
          ...base,
          id: newId(),
          type: 'text' as const,
          text: 'Seed',
          points: [{ time: last.time, price: last.close }],
        },
        {
          ...base,
          id: newId(),
          type: 'arrow' as const,
          points: [
            { time: mid.time, price: mid.close },
            { time: last.time, price: last.close * 1.02 },
          ],
        },
      ]

      state.replaceAll(seeded, true)
      state.setSelection([seeded[0].id])
    }

    maybeSeedDrawings()

    const startDrag = (event: PointerEvent, mode: DragMode, data: Partial<DragState>) => {
      dragRef.current = {
        mode,
        hasMoved: false,
        pointerId: event.pointerId,
        ...data,
      }
      setChartInteractionEnabled(false)
      container.setPointerCapture(event.pointerId)
    }

    const releaseCapture = (pointerId: number) => {
      try {
        if (container.hasPointerCapture(pointerId)) {
          container.releasePointerCapture(pointerId)
        }
      } catch {
        // ignore
      }
    }

    const stopDrag = () => {
      dragRef.current = { mode: 'none' }
      setChartInteractionEnabled(true)
    }

    const applyMove = (drag: DragState, point: DrawingPoint) => {
      if (!drag.shapeId || !drag.originPoint || !drag.originalPoints || !drag.shapeType) return
      const deltaTime = point.time - drag.originPoint.time
      const deltaPrice = point.price - drag.originPoint.price
      const shapeId = drag.shapeId
      const shapeType = drag.shapeType

      updateShape(shapeId, (shape) => {
        if (shape.locked) return shape
        const originalPoints = drag.originalPoints ?? shape.points
        const nextPoints = originalPoints.map((original) => ({
          time: original.time + deltaTime,
          price: original.price + deltaPrice,
        }))

        if (shapeType === 'horizontalLine') {
          nextPoints[0] = {
            time: originalPoints[0].time,
            price: originalPoints[0].price + deltaPrice,
          }
        } else if (shapeType === 'verticalLine') {
          nextPoints[0] = {
            time: originalPoints[0].time + deltaTime,
            price: originalPoints[0].price,
          }
        }

        return {
          ...shape,
          points: nextPoints,
        }
      })
    }

    const applyResize = (drag: DragState, coordinate: { x: number; y: number }, point: DrawingPoint) => {
      if (!drag.shapeId || !drag.originalPoints || drag.handleIndex === undefined || !drag.shapeType) return
      const shapeId = drag.shapeId
      const handleIndex = drag.handleIndex
      const shapeType = drag.shapeType

      if (shapeType === 'horizontalLine' || shapeType === 'verticalLine' || shapeType === 'text') {
        applyMove(drag, point)
        return
      }

      if (shapeType === 'rectangle') {
        const p1 = pointToCoordinate(chart, candlestickSeries, drag.originalPoints[0])
        const p2 = pointToCoordinate(chart, candlestickSeries, drag.originalPoints[1])
        if (!p1 || !p2) return

        let left = Math.min(p1.x, p2.x)
        let right = Math.max(p1.x, p2.x)
        let top = Math.min(p1.y, p2.y)
        let bottom = Math.max(p1.y, p2.y)

        switch (handleIndex) {
          case 0:
            left = coordinate.x
            top = coordinate.y
            break
          case 1:
            right = coordinate.x
            top = coordinate.y
            break
          case 2:
            right = coordinate.x
            bottom = coordinate.y
            break
          case 3:
            left = coordinate.x
            bottom = coordinate.y
            break
          default:
            break
        }

        const nextPoint1 = coordinateToPoint(chart, candlestickSeries, { x: left, y: top })
        const nextPoint2 = coordinateToPoint(chart, candlestickSeries, { x: right, y: bottom })
        if (!nextPoint1 || !nextPoint2) return

        updateShape(shapeId, (shape) => ({
          ...shape,
          points: [nextPoint1, nextPoint2],
        }))
        return
      }

      updateShape(shapeId, (shape) => {
        if (shape.locked) return shape
        const originalPoints = drag.originalPoints ?? shape.points
        const nextPoints = originalPoints.map((original) => ({ ...original }))
        if (handleIndex < nextPoints.length) {
          nextPoints[handleIndex] = point
        }
        return {
          ...shape,
          points: nextPoints,
        }
      })
    }

    const handlePointerMove = (coordinate: { x: number; y: number }) => {
      const drag = dragRef.current
      const point = toPoint(coordinate)
      if (!point) return

      if (drag.mode === 'drawing') {
        updateDraft(point)
        return
      }

      if (drag.mode === 'moving' || drag.mode === 'resizing') {
        if (!drag.originCoordinate) return
        const dx = coordinate.x - drag.originCoordinate.x
        const dy = coordinate.y - drag.originCoordinate.y
        if (!drag.hasMoved && Math.hypot(dx, dy) < MOVE_THRESHOLD) {
          return
        }
        if (!drag.hasMoved) {
          drag.hasMoved = true
          pushHistory()
        }

        if (drag.mode === 'moving') {
          applyMove(drag, point)
        } else {
          applyResize(drag, coordinate, point)
        }
        return
      }

      const state = useDrawingStore.getState()
      if (state.activeTool === 'select') {
        const hit = findHit(coordinate)
        setHovered(hit?.shape.id ?? null)
      }
    }

    const schedulePointerMove = (coordinate: { x: number; y: number }) => {
      pendingRef.current = coordinate
      if (rafRef.current !== null) return
      rafRef.current = window.requestAnimationFrame(() => {
        rafRef.current = null
        const pending = pendingRef.current
        pendingRef.current = null
        if (pending) {
          handlePointerMove(pending)
        }
      })
    }

    const onPointerDown = (event: PointerEvent) => {
      if (event.button !== 0) return
      if (!chart || !candlestickSeries) return
      if (!useDrawingStore.getState().layerVisible) return
      if (event.target instanceof Element && event.target.closest('[data-drawing-toolbar]')) return
      const coordinate = toCoordinate(event)
      if (!coordinate) return
      const point = toPoint(coordinate)
      if (!point) return

      const state = useDrawingStore.getState()
      if (state.activeTool !== 'select') {
        // If we already have an active draft for the current tool, continue placing points.
        // Multi-point tools (e.g., parallel channel) require multiple clicks.
        const draftToolId = typeof state.draft?.shape?.meta?.toolId === 'string' ? state.draft.shape.meta.toolId : null
        if (!state.draft || (draftToolId && draftToolId !== state.activeTool)) {
          state.startDraft(state.activeTool, point)
        }
        setMode('drawing')
        startDrag(event, 'drawing', {
          originPoint: point,
          originCoordinate: coordinate,
        })
        return
      }

      const hit = findHit(coordinate)
      if (!hit) {
        clearSelection()
        setMode('idle')
        return
      }

      setSelection([hit.shape.id])
      if (hit.shape.locked) {
        setMode('selected')
        return
      }

      if (hit.hit.handleIndex !== undefined) {
        setMode('resizing')
        startDrag(event, 'resizing', {
          shapeId: hit.shape.id,
          shapeType: hit.shape.type,
          handleIndex: hit.hit.handleIndex,
          originPoint: point,
          originCoordinate: coordinate,
          originalPoints: hit.shape.points.map((p) => ({ ...p })),
        })
        return
      }

      setMode('moving')
      startDrag(event, 'moving', {
        shapeId: hit.shape.id,
        shapeType: hit.shape.type,
        originPoint: point,
        originCoordinate: coordinate,
        originalPoints: hit.shape.points.map((p) => ({ ...p })),
      })
    }

    const onPointerMove = (event: PointerEvent) => {
      if (!useDrawingStore.getState().layerVisible) return
      if (dragRef.current.mode === 'none' && useDrawingStore.getState().activeTool === 'select') {
        const coordinate = toCoordinate(event)
        if (!coordinate) return
        schedulePointerMove(coordinate)
        return
      }

      if (dragRef.current.pointerId !== undefined && event.pointerId !== dragRef.current.pointerId) return
      const coordinate = toCoordinate(event)
      if (!coordinate) return
      schedulePointerMove(coordinate)
    }

    const onPointerUp = (event: PointerEvent) => {
      if (!useDrawingStore.getState().layerVisible) return
      if (dragRef.current.pointerId !== undefined && event.pointerId !== dragRef.current.pointerId) return
      const drag = dragRef.current
      if (drag.mode === 'drawing') {
        const draftBeforeCommit = useDrawingStore.getState().draft
        const coordinate = toCoordinate(event)
        if (coordinate) {
          const point = toPoint(coordinate)
          if (point) {
            updateDraft(point)
          }
        }
        commitDraft()

        if (draftBeforeCommit?.shape.type === 'text') {
          const id = draftBeforeCommit.shape.id
          const current = useDrawingStore.getState().shapes[id]
          if (current && current.type === 'text' && !current.locked) {
            const value = window.prompt('Text', current.text || 'Text')
            if (value !== null) {
              pushHistory()
              updateShape(id, (shape) => (shape.type === 'text' ? { ...shape, text: value } : shape))
            }
          }
        }
      }

      if (drag.mode === 'moving' || drag.mode === 'resizing') {
        setMode('selected')
      }

      releaseCapture(event.pointerId)
      stopDrag()
    }

    const onPointerCancel = (event: PointerEvent) => {
      if (dragRef.current.pointerId !== undefined && event.pointerId !== dragRef.current.pointerId) return
      if (dragRef.current.mode === 'drawing') {
        cancelDraft()
      }
      releaseCapture(event.pointerId)
      stopDrag()
    }

    const onDoubleClick = (event: MouseEvent) => {
      if (!useDrawingStore.getState().layerVisible) return
      if (event.target instanceof Element && event.target.closest('[data-drawing-toolbar]')) return
      const coordinate = toCoordinate(event as unknown as PointerEvent)
      if (!coordinate) return
      const hit = findHit(coordinate)
      if (!hit) return
      if (hit.shape.locked) return
      if (hit.shape.type !== 'text') return

      const value = window.prompt('Text', hit.shape.text || 'Text')
      if (value === null) return
      pushHistory()
      updateShape(hit.shape.id, (shape) => (shape.type === 'text' ? { ...shape, text: value } : shape))
    }

    const onKeyDown = (event: KeyboardEvent) => {
      if (isEditableTarget(event.target)) return
      const isMac = navigator.platform.toLowerCase().includes('mac')
      const modKey = isMac ? event.metaKey : event.ctrlKey

      if (event.key === 'Escape') {
        cancelDraft()
        setMode('idle')
        return
      }

      if (event.key === 'Delete' || event.key === 'Backspace') {
        deleteSelection()
        return
      }

      if (modKey && event.key.toLowerCase() === 'z') {
        event.preventDefault()
        if (event.shiftKey) {
          redo()
        } else {
          undo()
        }
        return
      }

      if (modKey && event.key.toLowerCase() === 'y') {
        event.preventDefault()
        redo()
      }
    }

    container.addEventListener('pointerdown', onPointerDown)
    container.addEventListener('pointermove', onPointerMove)
    container.addEventListener('pointerup', onPointerUp)
    container.addEventListener('pointercancel', onPointerCancel)
    container.addEventListener('dblclick', onDoubleClick)
    window.addEventListener('keydown', onKeyDown)

    return () => {
      container.removeEventListener('pointerdown', onPointerDown)
      container.removeEventListener('pointermove', onPointerMove)
      container.removeEventListener('pointerup', onPointerUp)
      container.removeEventListener('pointercancel', onPointerCancel)
      container.removeEventListener('dblclick', onDoubleClick)
      window.removeEventListener('keydown', onKeyDown)
      unsubscribeCursor()
      if (rafRef.current !== null) {
        window.cancelAnimationFrame(rafRef.current)
        rafRef.current = null
      }
      pendingRef.current = null
      stopDrag()
    }
  }, [
    chart,
    candlestickSeries,
    containerRef,
    chartData,
    isReady,
    updateDraft,
    commitDraft,
    cancelDraft,
    updateShape,
    setMode,
    setSelection,
    clearSelection,
    setHovered,
    deleteSelection,
    undo,
    redo,
    pushHistory,
    layerVisible,
    magnetEnabled,
  ])

  return null
}
