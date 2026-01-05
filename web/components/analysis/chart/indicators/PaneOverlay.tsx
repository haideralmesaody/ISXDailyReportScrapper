/**
 * PaneOverlay - Floating UI container for indicator pane controls
 *
 * Renders UI elements (gear icons, popovers) overlaid on TradingView panes
 * Uses absolute positioning to float above chart canvas
 */

'use client'

import { useEffect, useState, useRef, ReactNode } from 'react'
import { useChart } from '../ChartContext'

interface PaneOverlayProps {
  paneIndex: number
  children: ReactNode
}

export function PaneOverlay({ paneIndex, children }: PaneOverlayProps) {
  const { chart } = useChart()
  const [paneRect, setPaneRect] = useState<DOMRect | null>(null)
  const overlayRef = useRef<HTMLDivElement>(null)

  // Track pane position and size
  useEffect(() => {
    if (!chart) return

    const updatePanePosition = () => {
      try {
        const panes = chart.panes()

        if (paneIndex >= panes.length) {
          return
        }

        const pane = panes[paneIndex]

        // TradingView v5.0.8 API: Use getHTMLElement() instead of element()
        const paneElement = pane.getHTMLElement()

        if (paneElement) {
          const rect = paneElement.getBoundingClientRect()
          setPaneRect(rect)
        }
      } catch (error) {
        console.error('[PaneOverlay] Error getting pane position:', error)
      }
    }

    // Initial position
    updatePanePosition()

    // Update on window resize
    window.addEventListener('resize', updatePanePosition)

    // Update on chart resize (user drags separator)
    const resizeObserver = new ResizeObserver(updatePanePosition)
    const chartElement = chart.chartElement()
    if (chartElement) {
      resizeObserver.observe(chartElement)
    }

    return () => {
      window.removeEventListener('resize', updatePanePosition)
      resizeObserver.disconnect()
    }
  }, [chart, paneIndex])

  if (!paneRect) {
    return null
  }

  return (
    <div
      ref={overlayRef}
      style={{
        position: 'fixed',
        top: paneRect.top,
        left: paneRect.left,
        width: paneRect.width,
        height: paneRect.height,
        pointerEvents: 'none',  // Allow chart interaction through overlay
        zIndex: 100,
      }}
    >
      {/* Re-enable pointer events for interactive children */}
      <div style={{ pointerEvents: 'auto' }}>
        {children}
      </div>
    </div>
  )
}
