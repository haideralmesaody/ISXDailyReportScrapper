'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'

interface ToggleGroupProps {
  type: 'single' | 'multiple'
  value?: string | string[]
  onValueChange?: (value: string | string[]) => void
  className?: string
  children: React.ReactNode
  disabled?: boolean
}

const ToggleGroupContext = React.createContext<{
  type: 'single' | 'multiple'
  value?: string | string[]
  onValueChange?: (value: string | string[]) => void
}>({
  type: 'single',
})

export function ToggleGroup({
  type,
  value,
  onValueChange,
  className,
  children,
  disabled = false,
}: ToggleGroupProps) {
  return (
    <ToggleGroupContext.Provider value={{ type, value, onValueChange }}>
      <div 
        className={cn(
          "inline-flex rounded-md shadow-sm",
          disabled && "opacity-50 pointer-events-none",
          className
        )} 
        role="group"
      >
        {React.Children.map(children, (child, index) => {
          if (React.isValidElement(child)) {
            return React.cloneElement(child as React.ReactElement<any>, {
              isFirst: index === 0,
              isLast: index === React.Children.count(children) - 1,
            })
          }
          return child
        })}
      </div>
    </ToggleGroupContext.Provider>
  )
}

interface ToggleGroupItemProps {
  value: string
  children: React.ReactNode
  className?: string
  disabled?: boolean
  isFirst?: boolean
  isLast?: boolean
  'aria-label'?: string
}

export function ToggleGroupItem({
  value,
  children,
  className,
  disabled = false,
  isFirst,
  isLast,
  'aria-label': ariaLabel,
}: ToggleGroupItemProps) {
  const context = React.useContext(ToggleGroupContext)
  
  const isSelected = React.useMemo(() => {
    if (context.type === 'single') {
      return context.value === value
    }
    return Array.isArray(context.value) && context.value.includes(value)
  }, [context.type, context.value, value])
  
  const handleClick = React.useCallback(() => {
    if (disabled || !context.onValueChange) return
    
    if (context.type === 'single') {
      context.onValueChange(value)
    } else {
      const currentValue = Array.isArray(context.value) ? context.value : []
      if (isSelected) {
        context.onValueChange(currentValue.filter(v => v !== value))
      } else {
        context.onValueChange([...currentValue, value])
      }
    }
  }, [disabled, context, value, isSelected])
  
  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={disabled}
      aria-label={ariaLabel}
      aria-pressed={isSelected}
      data-state={isSelected ? 'on' : 'off'}
      className={cn(
        "px-3 py-2 text-sm font-medium transition-colors",
        "border border-input",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        isFirst && "rounded-l-md",
        isLast && "rounded-r-md",
        !isFirst && "-ml-px",
        isSelected 
          ? "bg-primary text-primary-foreground hover:bg-primary/90" 
          : "bg-background hover:bg-muted hover:text-accent-foreground",
        disabled && "opacity-50 cursor-not-allowed",
        className
      )}
    >
      {children}
    </button>
  )
}

// Re-export for convenience
export default ToggleGroup