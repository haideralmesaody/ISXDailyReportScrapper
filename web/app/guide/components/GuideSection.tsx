'use client'

import { ReactNode } from 'react'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { CheckCircle2, Circle, Clock, ChevronRight, ChevronLeft, Bookmark, BookmarkCheck } from 'lucide-react'
import { cn } from '@/lib/utils'

interface GuideSectionProps {
  id: string
  title: string
  description: string
  children: ReactNode
  estimatedMinutes?: number
  isCompleted: boolean
  isBookmarked: boolean
  onComplete: () => void
  onBookmark: () => void
  onNext?: () => void
  onPrevious?: () => void
  nextSectionTitle?: string
  previousSectionTitle?: string
  className?: string
}

/**
 * Reusable wrapper component for guide sections
 * Provides header, breadcrumbs, completion button, and navigation
 */
export function GuideSection({
  title,
  description,
  children,
  estimatedMinutes,
  isCompleted,
  isBookmarked,
  onComplete,
  onBookmark,
  onNext,
  onPrevious,
  nextSectionTitle,
  previousSectionTitle,
  className
}: GuideSectionProps) {
  return (
    <div className={cn('space-y-6', className)}>
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-2 flex-1">
            <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
            <p className="text-muted-foreground text-lg">{description}</p>
          </div>
          <div className="flex items-center gap-2">
            {estimatedMinutes && (
              <Badge variant="secondary" className="gap-1">
                <Clock className="h-3 w-3" />
                {estimatedMinutes} min
              </Badge>
            )}
            <Button
              variant="ghost"
              size="icon"
              onClick={onBookmark}
              className={cn(
                'transition-colors',
                isBookmarked && 'text-yellow-500 hover:text-yellow-600'
              )}
            >
              {isBookmarked ? (
                <BookmarkCheck className="h-5 w-5" />
              ) : (
                <Bookmark className="h-5 w-5" />
              )}
            </Button>
          </div>
        </div>

        {/* Completion Status */}
        <div className="flex items-center gap-2">
          {isCompleted ? (
            <CheckCircle2 className="h-5 w-5 text-green-500" />
          ) : (
            <Circle className="h-5 w-5 text-muted-foreground" />
          )}
          <span className="text-sm text-muted-foreground">
            {isCompleted ? 'Completed' : 'Not completed'}
          </span>
        </div>
      </div>

      <Separator />

      {/* Content */}
      <div className="prose prose-gray dark:prose-invert max-w-none">
        {children}
      </div>

      <Separator />

      {/* Footer Actions */}
      <div className="flex items-center justify-between gap-4">
        {/* Previous Section */}
        <div className="flex-1">
          {onPrevious && previousSectionTitle && (
            <Button
              variant="outline"
              onClick={onPrevious}
              className="gap-2"
            >
              <ChevronLeft className="h-4 w-4" />
              <span className="hidden sm:inline">Previous:</span> {previousSectionTitle}
            </Button>
          )}
        </div>

        {/* Mark as Complete */}
        <Button
          onClick={onComplete}
          variant={isCompleted ? 'outline' : 'default'}
          className="gap-2"
        >
          {isCompleted ? (
            <>
              <CheckCircle2 className="h-4 w-4" />
              Completed
            </>
          ) : (
            <>
              <Circle className="h-4 w-4" />
              Mark as Complete
            </>
          )}
        </Button>

        {/* Next Section */}
        <div className="flex-1 flex justify-end">
          {onNext && nextSectionTitle && (
            <Button
              onClick={onNext}
              className="gap-2"
            >
              <span className="hidden sm:inline">Next:</span> {nextSectionTitle}
              <ChevronRight className="h-4 w-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
