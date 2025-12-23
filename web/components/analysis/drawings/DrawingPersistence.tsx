'use client'

import { useEffect, useRef } from 'react'
import { useDrawingStore } from './store'
import type { DrawingShape } from './types'
import { normalizeDrawingsEnvelope, type DrawingsEnvelope } from './persistence'

const buildShapesPayload = (): DrawingShape[] => {
  const state = useDrawingStore.getState()
  return state.order
    .map((id) => state.shapes[id])
    .filter((shape): shape is DrawingShape => Boolean(shape))
}

export function DrawingPersistence({ ticker }: { ticker: string }) {
  const isHydratingRef = useRef(false)
  const lastTickerRef = useRef<string | null>(null)
  const debounceRef = useRef<number | null>(null)
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    if (!ticker) return

    const normalizedTicker = ticker.toUpperCase()
    lastTickerRef.current = normalizedTicker

    isHydratingRef.current = true
    abortRef.current?.abort()
    abortRef.current = new AbortController()

    const load = async () => {
      try {
        const response = await fetch(`/api/v1/drawings/${encodeURIComponent(normalizedTicker)}/`, {
          method: 'GET',
          credentials: 'include',
          signal: abortRef.current?.signal,
        })

        if (!response.ok) {
          useDrawingStore.getState().replaceAll([], true)
          return
        }

        const json = await response.json()
        const envelope = normalizeDrawingsEnvelope(json, normalizedTicker)
        useDrawingStore.getState().replaceAll(envelope.shapes ?? [], true)
      } catch {
        // If request aborted or failed, keep current drawings.
      } finally {
        isHydratingRef.current = false
      }
    }

    void load()

    return () => {
      abortRef.current?.abort()
      abortRef.current = null
    }
  }, [ticker])

  useEffect(() => {
    if (!ticker) return
    const normalizedTicker = ticker.toUpperCase()

    const scheduleSave = () => {
      if (debounceRef.current !== null) {
        window.clearTimeout(debounceRef.current)
      }
      debounceRef.current = window.setTimeout(() => {
        debounceRef.current = null
        void saveNow(normalizedTicker)
      }, 1000)
    }

    const saveNow = async (activeTicker: string) => {
      const state = useDrawingStore.getState()
      if (isHydratingRef.current) return
      if (!state.dirty) return
      if (!activeTicker) return

      const shapes = buildShapesPayload()
      const payload: DrawingsEnvelope = {
        version: 1,
        ticker: activeTicker,
        updatedAt: Date.now(),
        shapes,
      }

      try {
        const response = await fetch(`/api/v1/drawings/${encodeURIComponent(activeTicker)}/`, {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
          keepalive: true,
        })

        if (!response.ok) return

        useDrawingStore.getState().markSaved()
      } catch {
        // ignore (offline, aborted, etc.)
      }
    }

    const unsubscribe = useDrawingStore.subscribe(
      (state) => ({
        dirty: state.dirty,
        shapes: state.shapes,
        order: state.order,
      }),
      (next) => {
        if (isHydratingRef.current) return
        if (!next.dirty) return
        scheduleSave()
      }
    )

    return () => {
      unsubscribe()
      if (debounceRef.current !== null) {
        window.clearTimeout(debounceRef.current)
        debounceRef.current = null
      }

      // Best-effort flush on unmount/ticker switch.
      const currentTicker = lastTickerRef.current
      if (currentTicker === normalizedTicker) {
        void saveNow(normalizedTicker)
      }
    }
  }, [ticker])

  return null
}
