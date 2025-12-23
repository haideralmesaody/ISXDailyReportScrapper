/**
 * ResizablePanel Component
 * Provides drag-to-resize functionality with min/max constraints and localStorage persistence
 * Following CLAUDE.md React patterns with proper hydration handling
 */

'use client'

import React, { useRef, useState, useCallback, useEffect } from 'react'
import { GripVertical } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ResizablePanelProps {
  children: React.ReactNode
  defaultWidth: number
  minWidth: number
  maxWidth: number
  width?: number
  disableResize?: boolean
  onWidthChange?: (width: number) => void
  storageKey?: string
  className?: string
}

export function ResizablePanel({
  children,
  defaultWidth,
  minWidth,
  maxWidth,
  width,
  disableResize,
  onWidthChange,
  storageKey,
  className
}: ResizablePanelProps) {
  const [internalWidth, setInternalWidth] = useState(() => {
    if (storageKey && typeof window !== 'undefined') {
      const saved = localStorage.getItem(storageKey)
      if (saved) {
        const parsedWidth = parseInt(saved)
        // Ensure saved width is within constraints
        return Math.max(minWidth, Math.min(maxWidth, parsedWidth))
      }
    }
    return defaultWidth
  })

  const effectiveWidth = width ?? internalWidth

  const panelRef = useRef<HTMLDivElement>(null)
  const isDraggingRef = useRef(false)
  const startXRef = useRef(0)
  const startWidthRef = useRef(0)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (disableResize) return
    e.preventDefault()
    isDraggingRef.current = true
    startXRef.current = e.clientX
    startWidthRef.current = effectiveWidth
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
  }, [disableResize, effectiveWidth])

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDraggingRef.current) return

    const delta = e.clientX - startXRef.current
    const newWidth = startWidthRef.current + delta
    const constrainedWidth = Math.max(minWidth, Math.min(maxWidth, newWidth))

    onWidthChange?.(constrainedWidth)

    if (width === undefined) {
      setInternalWidth(constrainedWidth)
    }
  }, [minWidth, maxWidth, onWidthChange, width])

  const handleMouseUp = useCallback(() => {
    if (!isDraggingRef.current) return

    isDraggingRef.current = false
    document.body.style.cursor = ''
    document.body.style.userSelect = ''

    // Save to localStorage
    if (storageKey) {
      localStorage.setItem(storageKey, effectiveWidth.toString())
    }
  }, [effectiveWidth, storageKey])

  useEffect(() => {
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [handleMouseMove, handleMouseUp])

  return (
    <div
      ref={panelRef}
      style={{ width: `${effectiveWidth}px` }}
      className={cn(
        "relative flex-shrink-0 bg-card border-r overflow-hidden flex flex-col transition-[width] duration-200 ease-out",
        className
      )}
    >
      {children}

      {/* Resize Handle */}
      <div
        aria-hidden={disableResize ? true : undefined}
        onMouseDown={handleMouseDown}
        className={cn(
          "absolute top-0 right-0 bottom-0 w-1 cursor-col-resize z-10",
          disableResize && "pointer-events-none opacity-0",
          "hover:bg-primary/20 active:bg-primary/30 transition-colors",
          "group flex items-center justify-center"
        )}
        title="Drag to resize panel"
      >
        <div className="absolute right-0 w-4 h-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-primary/10">
          <GripVertical className="h-4 w-4 text-primary" />
        </div>
      </div>
    </div>
  )
}
