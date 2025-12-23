/**
 * ColoredProgress component
 * Progress bar that changes color based on value ranges for liquidity scores
 */

"use client"

import * as React from "react"
import * as ProgressPrimitive from "@radix-ui/react-progress"
import { cn } from "@/lib/utils"

interface ColoredProgressProps extends React.ComponentPropsWithoutRef<typeof ProgressPrimitive.Root> {
  value?: number
  colorMode?: 'score' | 'custom'
  customColors?: {
    background?: string
    indicator?: string
  }
}

/**
 * Get the appropriate color class based on score value
 * Matches the liquidity scoring color scheme
 */
function getProgressColor(value: number): {
  background: string
  indicator: string
} {
  if (value >= 80) {
    // Excellent liquidity - Green
    return {
      background: 'bg-green-200 dark:bg-green-950',
      indicator: 'bg-green-600 dark:bg-green-500'
    }
  }
  if (value >= 60) {
    // Good liquidity - Blue
    return {
      background: 'bg-blue-200 dark:bg-blue-950',
      indicator: 'bg-blue-600 dark:bg-blue-500'
    }
  }
  if (value >= 40) {
    // Moderate liquidity - Yellow/Amber
    return {
      background: 'bg-yellow-200 dark:bg-yellow-950',
      indicator: 'bg-yellow-600 dark:bg-yellow-500'
    }
  }
  if (value >= 20) {
    // Poor liquidity - Orange
    return {
      background: 'bg-orange-200 dark:bg-orange-950',
      indicator: 'bg-orange-600 dark:bg-orange-500'
    }
  }
  // Very poor liquidity - Red
  return {
    background: 'bg-red-200 dark:bg-red-950',
    indicator: 'bg-red-600 dark:bg-red-500'
  }
}

const ColoredProgress = React.forwardRef<
  React.ElementRef<typeof ProgressPrimitive.Root>,
  ColoredProgressProps
>(({ className, value = 0, colorMode = 'score', customColors, ...props }, ref) => {
  const colors = colorMode === 'custom' && customColors 
    ? customColors 
    : getProgressColor(value)

  return (
    <ProgressPrimitive.Root
      ref={ref}
      className={cn(
        "relative h-2 w-full overflow-hidden rounded-full transition-colors duration-300",
        colors.background,
        className
      )}
      {...props}
    >
      <ProgressPrimitive.Indicator
        className={cn(
          "h-full w-full flex-1 transition-all duration-300",
          colors.indicator
        )}
        style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
      />
    </ProgressPrimitive.Root>
  )
})

ColoredProgress.displayName = "ColoredProgress"

export { ColoredProgress, getProgressColor }