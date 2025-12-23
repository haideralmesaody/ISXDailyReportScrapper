import React, { useCallback, useState, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { GripHorizontal } from 'lucide-react'

interface PanelResizeHandleProps {
  position: number // Position in percentage (0-100)
  onResize: (delta: number) => void
  label?: string
  theme?: 'light' | 'dark'
  className?: string
}

export function PanelResizeHandle({ 
  position, 
  onResize, 
  label, 
  theme = 'dark',
  className 
}: PanelResizeHandleProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [startY, setStartY] = useState(0)
  const [isHovered, setIsHovered] = useState(false)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    setIsDragging(true)
    setStartY(e.clientY)
  }, [])

  useEffect(() => {
    if (!isDragging) return

    const handleMouseMove = (e: MouseEvent) => {
      const deltaY = e.clientY - startY
      const containerHeight = window.innerHeight
      const deltaPercent = (deltaY / containerHeight) * 100
      onResize(deltaPercent)
      setStartY(e.clientY)
    }

    const handleMouseUp = () => {
      setIsDragging(false)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, startY, onResize])

  return (
    <div
      className={cn(
        'absolute left-0 right-0 z-30 group',
        'transition-all duration-200',
        isDragging && 'cursor-ns-resize',
        className
      )}
      style={{ top: `${position}%` }}
      onMouseDown={handleMouseDown}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Hit area - larger for easier grabbing */}
      <div className="absolute inset-x-0 -top-2 -bottom-2 cursor-ns-resize" />
      
      {/* Visual separator line */}
      <div 
        className={cn(
          'relative h-[2px] transition-all duration-200',
          theme === 'dark' 
            ? isDragging || isHovered
              ? 'bg-blue-500/50' 
              : 'bg-gray-700/50'
            : isDragging || isHovered
              ? 'bg-blue-600/50'
              : 'bg-gray-300/50'
        )}
      >
        {/* Grip indicator in center */}
        <div 
          className={cn(
            'absolute left-1/2 -translate-x-1/2 -top-3 px-3 py-1 rounded-md',
            'transition-all duration-200 flex items-center gap-1',
            isDragging || isHovered ? 'opacity-100' : 'opacity-0',
            theme === 'dark' 
              ? 'bg-gray-800/90 border border-gray-700'
              : 'bg-white/90 border border-gray-200'
          )}
        >
          <GripHorizontal className="h-3 w-3" />
          {label && <span className="text-xs font-medium">{label}</span>}
        </div>
      </div>
    </div>
  )
}