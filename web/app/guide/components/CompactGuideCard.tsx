/**
 * Compact Guide Card
 * Minimal, clean design matching Operations page style
 * Shows section info, completion status, and estimated time
 */

'use client'

import React from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Clock,
  CheckCircle2,
  Circle,
  ChevronRight,
  BookOpen,
  Play
} from 'lucide-react'
import { cn } from '@/lib/utils'
import type { LucideIcon } from 'lucide-react'

interface CompactGuideCardProps {
  section: {
    id: string
    title: string
    description: string
    icon?: LucideIcon
    category: string
    estimatedMinutes?: number
  }
  isCompleted: boolean
  isInProgress: boolean
  completionPercentage?: number
  onClick: () => void
  className?: string
}

export function CompactGuideCard({
  section,
  isCompleted,
  isInProgress,
  completionPercentage = 0,
  onClick,
  className
}: CompactGuideCardProps) {
  const Icon = section.icon || BookOpen

  // Get button text based on status
  const getButtonText = () => {
    if (isCompleted) return 'Review'
    if (isInProgress) return 'Continue'
    return 'Start Learning'
  }

  // Get button icon
  const getButtonIcon = () => {
    if (isCompleted) return CheckCircle2
    if (isInProgress) return Play
    return BookOpen
  }

  const ButtonIcon = getButtonIcon()

  return (
    <Card
      className={cn(
        "group relative overflow-hidden transition-all duration-200",
        "hover:shadow-lg hover:scale-[1.02] hover:z-10",
        "h-[180px] flex flex-col", // Fixed height matching Operations page
        className
      )}
    >
      {/* Status indicator line at top */}
      <div className={cn(
        "absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r",
        isCompleted && "from-green-500 to-green-400",
        isInProgress && !isCompleted && "from-blue-500 to-blue-400",
        !isCompleted && !isInProgress && "from-muted to-muted"
      )} />

      <CardHeader className="pb-2 flex-none">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className={cn(
              "p-2 rounded-lg transition-colors",
              isCompleted && "bg-green-100 dark:bg-green-950",
              isInProgress && !isCompleted && "bg-blue-100 dark:bg-blue-950",
              !isCompleted && !isInProgress && "bg-primary/10 group-hover:bg-primary/20"
            )}>
              <Icon className={cn(
                "h-5 w-5",
                isCompleted && "text-green-600 dark:text-green-400",
                isInProgress && !isCompleted && "text-blue-600 dark:text-blue-400",
                !isCompleted && !isInProgress && "text-primary"
              )} />
            </div>
            <div>
              <CardTitle className="text-base font-semibold leading-tight">
                {section.title}
              </CardTitle>
              {/* Category badge */}
              <Badge variant="secondary" className="mt-1 text-xs capitalize">
                {section.category}
              </Badge>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col justify-between pb-4">
        <div className="space-y-2">
          <CardDescription className="text-xs line-clamp-2">
            {section.description}
          </CardDescription>

          {/* Progress and time indicator */}
          <div className="flex items-center justify-between gap-2">
            {/* Estimated time */}
            {section.estimatedMinutes && (
              <div className="flex items-center gap-1.5 text-xs">
                <Clock className="h-3 w-3 text-muted-foreground" />
                <span className="text-muted-foreground font-medium">
                  ~{section.estimatedMinutes} min
                </span>
              </div>
            )}

            {/* Completion status */}
            {isInProgress && !isCompleted && completionPercentage > 0 && (
              <div className="flex items-center gap-1.5 text-xs">
                <div className="flex gap-0.5">
                  {[...Array(4)].map((_, i) => (
                    <Circle
                      key={i}
                      className={cn(
                        "h-2 w-2",
                        i < Math.floor(completionPercentage / 25)
                          ? "fill-blue-500 text-blue-500"
                          : "fill-muted text-muted"
                      )}
                    />
                  ))}
                </div>
                <span className="text-muted-foreground font-medium">
                  {completionPercentage}%
                </span>
              </div>
            )}

            {isCompleted && (
              <div className="flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
                <CheckCircle2 className="h-3 w-3" />
                <span className="font-medium">Completed</span>
              </div>
            )}
          </div>
        </div>

        {/* Action button */}
        <Button
          onClick={onClick}
          size="sm"
          variant={isCompleted ? "outline" : "default"}
          className={cn(
            "w-full mt-3 group/btn transition-all",
            !isCompleted && !isInProgress && "animate-pulse"
          )}
        >
          <ButtonIcon className="h-3.5 w-3.5 mr-1.5" />
          {getButtonText()}
          <ChevronRight className="h-3 w-3 ml-auto opacity-50 group-hover/btn:opacity-100 group-hover/btn:translate-x-0.5 transition-all" />
        </Button>
      </CardContent>

      {/* Hover effect gradient */}
      <div className={cn(
        "absolute inset-0 bg-gradient-to-t transition-opacity pointer-events-none",
        isCompleted && "from-green-500/5 to-transparent opacity-0 group-hover:opacity-100",
        isInProgress && !isCompleted && "from-blue-500/5 to-transparent opacity-0 group-hover:opacity-100",
        !isCompleted && !isInProgress && "from-primary/5 to-transparent opacity-0 group-hover:opacity-100"
      )} />
    </Card>
  )
}

export default CompactGuideCard
