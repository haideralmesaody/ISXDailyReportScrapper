/**
 * Section Progress Bar
 * Displays segmented progress visualization similar to UnifiedOperationProgress
 * Shows completion status with color-coded segments
 */

'use client'

import React from 'react'
import { cn } from '@/lib/utils'
import { CheckCircle2, Circle, Clock } from 'lucide-react'

interface SectionProgressBarProps {
  /**
   * Total number of subsections or learning units
   */
  totalSegments: number

  /**
   * Number of completed segments
   */
  completedSegments: number

  /**
   * Estimated time to complete remaining segments (in minutes)
   */
  estimatedMinutes?: number

  /**
   * Current active segment (0-indexed)
   */
  activeSegment?: number

  /**
   * Show time estimate
   */
  showTime?: boolean

  /**
   * Compact mode (smaller segments)
   */
  compact?: boolean

  className?: string
}

export function SectionProgressBar({
  totalSegments,
  completedSegments,
  estimatedMinutes = 0,
  activeSegment,
  showTime = true,
  compact = false,
  className
}: SectionProgressBarProps) {
  const completionPercentage = Math.round((completedSegments / totalSegments) * 100)
  const remainingMinutes = Math.round((estimatedMinutes * (totalSegments - completedSegments)) / totalSegments)

  return (
    <div className={cn("space-y-2", className)}>
      {/* Progress header */}
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-2">
          {completedSegments === totalSegments ? (
            <CheckCircle2 className="h-4 w-4 text-green-500" />
          ) : (
            <Circle className="h-4 w-4 text-muted-foreground" />
          )}
          <span className="font-medium">
            {completedSegments} of {totalSegments} units completed
          </span>
        </div>

        <div className="flex items-center gap-4">
          {showTime && remainingMinutes > 0 && (
            <div className="flex items-center gap-1.5 text-muted-foreground">
              <Clock className="h-3 w-3" />
              <span className="text-xs">~{remainingMinutes}m remaining</span>
            </div>
          )}
          <span className="text-muted-foreground font-medium">
            {completionPercentage}%
          </span>
        </div>
      </div>

      {/* Segmented progress bar */}
      <div className="flex gap-1">
        {Array.from({ length: totalSegments }).map((_, index) => {
          const isCompleted = index < completedSegments
          const isActive = index === activeSegment
          const isCurrent = !isCompleted && index === completedSegments

          return (
            <div
              key={index}
              className={cn(
                "flex-1 rounded-sm transition-all duration-300",
                compact ? "h-1.5" : "h-2",
                isCompleted && "bg-green-500",
                isActive && !isCompleted && "bg-blue-500 animate-pulse",
                isCurrent && !isActive && "bg-blue-400",
                !isCompleted && !isActive && !isCurrent && "bg-muted"
              )}
              title={
                isCompleted
                  ? `Unit ${index + 1}: Completed`
                  : isActive
                  ? `Unit ${index + 1}: In Progress`
                  : isCurrent
                  ? `Unit ${index + 1}: Next`
                  : `Unit ${index + 1}: Pending`
              }
            />
          )
        })}
      </div>

      {/* Segment labels (optional, for non-compact mode) */}
      {!compact && totalSegments <= 10 && (
        <div className="flex justify-between text-[10px] text-muted-foreground px-0.5">
          {Array.from({ length: Math.min(totalSegments, 5) }).map((_, index) => {
            // Show labels for first, middle, and last segments
            const segmentIndex = index === 0
              ? 0
              : index === 4
              ? totalSegments - 1
              : Math.floor((totalSegments * index) / 4)

            return (
              <span key={index} className="leading-none">
                {segmentIndex + 1}
              </span>
            )
          })}
        </div>
      )}

      {/* Progress phases (for longer sections) */}
      {totalSegments > 10 && (
        <div className="flex items-center justify-center gap-2 pt-1">
          {['Basics', 'Intermediate', 'Advanced', 'Mastery'].map((phase, idx) => {
            const phaseStart = Math.floor((totalSegments * idx) / 4)
            const phaseEnd = Math.floor((totalSegments * (idx + 1)) / 4)
            const phaseCompleted = completedSegments >= phaseEnd
            const phaseInProgress = completedSegments >= phaseStart && completedSegments < phaseEnd

            return (
              <div
                key={phase}
                className={cn(
                  "flex items-center gap-1 text-xs transition-colors",
                  phaseCompleted && "text-green-600 dark:text-green-400 font-medium",
                  phaseInProgress && !phaseCompleted && "text-primary font-medium",
                  !phaseCompleted && !phaseInProgress && "text-muted-foreground"
                )}
              >
                <div className={cn(
                  "h-1.5 w-1.5 rounded-full",
                  phaseCompleted && "bg-green-500",
                  phaseInProgress && !phaseCompleted && "bg-primary",
                  !phaseCompleted && !phaseInProgress && "bg-muted"
                )} />
                <span>{phase}</span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default SectionProgressBar
