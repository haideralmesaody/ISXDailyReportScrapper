'use client'

import { Card } from '@/components/ui/card'
import {
  Calendar,
  Bookmark,
  TrendingUp
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { GuideProgress } from '@/types/guide'

interface ProgressStatsProps {
  progress: GuideProgress
  completedSections: number
  totalSections: number
  className?: string
}

/**
 * Progress statistics panel showing accurate metrics
 * - Started date (real timestamp)
 * - Completion percentage (real progress)
 * - Bookmarks count (real data)
 */
export function ProgressStats({
  progress,
  completedSections,
  totalSections,
  className
}: ProgressStatsProps) {
  const percentage = Math.round((completedSections / totalSections) * 100)

  // Calculate date-based metrics
  const startedDate = new Date(progress.startedAt)
  const now = new Date()
  const daysElapsed = Math.floor((now.getTime() - startedDate.getTime()) / (1000 * 60 * 60 * 24))

  // Format started date
  const formattedStartDate = startedDate.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })

  return (
    <Card className={cn('p-4', className)}>
      <div className="space-y-3">
        {/* Header */}
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">Your Progress</h3>
          <TrendingUp className="h-4 w-4 text-muted-foreground" />
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 gap-3">
          {/* Started Date */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <Calendar className="h-3 w-3" />
              <span>Started</span>
            </div>
            <p className="text-sm font-medium">
              {formattedStartDate}
            </p>
            <p className="text-[10px] text-muted-foreground">
              {daysElapsed === 0 ? 'Today' : `${daysElapsed} ${daysElapsed === 1 ? 'day' : 'days'} ago`}
            </p>
          </div>

          {/* Bookmarks */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <Bookmark className="h-3 w-3" />
              <span>Bookmarked</span>
            </div>
            <p className="text-sm font-medium">
              {progress.bookmarks.length} {progress.bookmarks.length === 1 ? 'section' : 'sections'}
            </p>
            <p className="text-[10px] text-muted-foreground">
              for quick access
            </p>
          </div>
        </div>

        {/* Progress Summary */}
        {percentage < 100 && (
          <div className="pt-2 border-t">
            <p className="text-xs text-center text-muted-foreground">
              {percentage}% complete ({completedSections}/{totalSections} sections)
            </p>
          </div>
        )}

        {/* Completion Message */}
        {percentage >= 100 && (
          <div className="pt-2 border-t">
            <p className="text-xs text-center text-muted-foreground">
              Completed on {now.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
            </p>
          </div>
        )}
      </div>
    </Card>
  )
}
