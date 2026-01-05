'use client'

import { useMemo, type ComponentType } from 'react'
import {
  ArrowRight,
  ArrowUpRight,
  Columns3,
  Eye,
  EyeOff,
  Lock,
  Magnet,
  MousePointer2,
  MoveHorizontal,
  MoveVertical,
  Redo2,
  Ruler,
  Square,
  Trash2,
  TrendingUp,
  Type,
  Undo2,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { useDrawingStore } from './store'
import { PHASE_ONE_TOOLS } from './tools'
import type { DrawingToolId } from './types'

const TOOL_ICON: Record<DrawingToolId, ComponentType<{ className?: string }>> = {
  select: MousePointer2,
  'trend-line': TrendingUp,
  ray: ArrowRight,
  'parallel-channel': Columns3,
  'horizontal-line': MoveHorizontal,
  'vertical-line': MoveVertical,
  rectangle: Square,
  text: Type,
  arrow: ArrowUpRight,
  measure: Ruler,
}

const TOOL_ORDER: DrawingToolId[] = [
  'select',
  'trend-line',
  'ray',
  'parallel-channel',
  'horizontal-line',
  'vertical-line',
  'rectangle',
  'text',
  'arrow',
  'measure',
]

const TOOL_LABEL = new Map(PHASE_ONE_TOOLS.map((tool) => [tool.id, tool.label]))

export function DrawingToolbar() {
  const activeTool = useDrawingStore((state) => state.activeTool)
  const selection = useDrawingStore((state) => state.selection)
  const shapes = useDrawingStore((state) => state.shapes)
  const hoveredId = useDrawingStore((state) => state.hoveredId)
  const layerVisible = useDrawingStore((state) => state.layerVisible)
  const history = useDrawingStore((state) => state.history)
  const magnetEnabled = useDrawingStore((state) => state.magnetEnabled)

  const setTool = useDrawingStore((state) => state.setTool)
  const deleteActive = useDrawingStore((state) => state.deleteActive)
  const toggleLockActive = useDrawingStore((state) => state.toggleLockActive)
  const toggleLayerVisible = useDrawingStore((state) => state.toggleLayerVisible)
  const toggleMagnet = useDrawingStore((state) => state.toggleMagnet)
  const undo = useDrawingStore((state) => state.undo)
  const redo = useDrawingStore((state) => state.redo)

  const selectedLockedState = useMemo(() => {
    const ids = selection.length ? selection : hoveredId ? [hoveredId] : []
    if (ids.length === 0) return { mixed: false, locked: false }

    const lockedValues = ids
      .map((id) => shapes[id]?.locked)
      .filter((value): value is boolean => typeof value === 'boolean')
    if (lockedValues.length === 0) return { mixed: false, locked: false }
    const allLocked = lockedValues.every(Boolean)
    const allUnlocked = lockedValues.every((v) => !v)
    return { mixed: !(allLocked || allUnlocked), locked: allLocked }
  }, [hoveredId, selection, shapes])

  const canDelete = selection.length > 0 || Boolean(hoveredId)
  const canUndo = history.past.length > 0
  const canRedo = history.future.length > 0

  return (
    <TooltipProvider delayDuration={250}>
      <div
        data-drawing-toolbar
        onPointerDown={(event) => event.stopPropagation()}
        onPointerUp={(event) => event.stopPropagation()}
        onDoubleClick={(event) => event.stopPropagation()}
        className={cn(
          'absolute left-3 top-3 z-40 flex flex-col rounded-lg border bg-white/95 p-1 shadow-lg backdrop-blur-sm dark:bg-zinc-950/95',
          !layerVisible && 'opacity-75'
        )}
      >
        <div className="flex flex-col gap-0.5">
          {TOOL_ORDER.map((toolId) => {
            const Icon = TOOL_ICON[toolId]
            const label = TOOL_LABEL.get(toolId) ?? toolId
            const isActive = activeTool === toolId && layerVisible
            return (
              <Tooltip key={toolId}>
                <TooltipTrigger asChild>
                  <Button
                    type="button"
                    size="icon"
                    variant="ghost"
                    className={cn(
                      'h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white',
                      isActive && 'bg-zinc-200 text-zinc-900 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-white dark:hover:bg-zinc-800'
                    )}
                    onClick={() => {
                      if (!layerVisible) {
                        toggleLayerVisible()
                      }
                      setTool(toolId)
                    }}
                    aria-label={label}
                  >
                    <Icon className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right">{label}</TooltipContent>
              </Tooltip>
            )
          })}
        </div>

        <div className="my-1 h-px w-full bg-zinc-200 dark:bg-zinc-800" />

        <div className="flex flex-col gap-0.5">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 disabled:opacity-40 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white"
                onClick={undo}
                disabled={!canUndo}
                aria-label="Undo"
              >
                <Undo2 className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">Undo</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 disabled:opacity-40 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white"
                onClick={redo}
                disabled={!canRedo}
                aria-label="Redo"
              >
                <Redo2 className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">Redo</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 disabled:opacity-40 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white"
                onClick={deleteActive}
                disabled={!canDelete}
                aria-label="Delete"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">Delete</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 disabled:opacity-40 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white"
                onClick={toggleLockActive}
                disabled={!canDelete}
                aria-label="Lock/unlock"
              >
                <Lock className={cn('h-4 w-4', selectedLockedState.mixed && 'opacity-70')} />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">
              {selectedLockedState.locked ? 'Unlock' : 'Lock'}
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className="h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white"
                onClick={toggleLayerVisible}
                aria-label={layerVisible ? 'Hide drawings' : 'Show drawings'}
              >
                {layerVisible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">{layerVisible ? 'Hide drawings' : 'Show drawings'}</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className={cn(
                  'h-9 w-9 rounded-md p-0 text-zinc-700 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-200 dark:hover:bg-zinc-800 dark:hover:text-white',
                  magnetEnabled && 'bg-zinc-200 text-zinc-900 hover:bg-zinc-200 dark:bg-zinc-800 dark:text-white dark:hover:bg-zinc-800'
                )}
                onClick={toggleMagnet}
                aria-label={magnetEnabled ? 'Magnet on' : 'Magnet off'}
              >
                <Magnet className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">{magnetEnabled ? 'Magnet on' : 'Magnet off'}</TooltipContent>
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  )
}
