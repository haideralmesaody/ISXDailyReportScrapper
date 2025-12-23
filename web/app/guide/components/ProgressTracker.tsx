'use client'

import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import {
  CheckCircle2,
  Trophy,
  Award,
  Star,
  Medal,
  type LucideIcon
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { motion, AnimatePresence } from 'framer-motion'

interface ProgressTrackerProps {
  completedSections: number
  totalSections: number
  className?: string
}

interface MilestoneBadge {
  id: string
  icon: LucideIcon
  label: string
  threshold: number  // percentage
  color: string
  bgColor: string
}

const milestones: MilestoneBadge[] = [
  {
    id: 'bronze',
    icon: Medal,
    label: 'Getting Started',
    threshold: 30,
    color: 'text-orange-600 dark:text-orange-400',
    bgColor: 'bg-orange-100 dark:bg-orange-950'
  },
  {
    id: 'silver',
    icon: Award,
    label: 'Halfway There',
    threshold: 50,
    color: 'text-gray-600 dark:text-gray-400',
    bgColor: 'bg-gray-100 dark:bg-gray-900'
  },
  {
    id: 'gold',
    icon: Star,
    label: 'Almost Done',
    threshold: 80,
    color: 'text-yellow-600 dark:text-yellow-400',
    bgColor: 'bg-yellow-100 dark:bg-yellow-950'
  },
  {
    id: 'platinum',
    icon: Trophy,
    label: 'Guide Master',
    threshold: 100,
    color: 'text-purple-600 dark:text-purple-400',
    bgColor: 'bg-purple-100 dark:bg-purple-950'
  }
]

/**
 * Enhanced progress tracker with milestone badges
 * Shows percentage, completion count, and earned badges
 */
export function ProgressTracker({
  completedSections,
  totalSections,
  className
}: ProgressTrackerProps) {
  const percentage = Math.round((completedSections / totalSections) * 100)
  const isComplete = completedSections === totalSections

  // Determine which badges are earned
  const earnedBadges = milestones.filter(badge => percentage >= badge.threshold)
  const nextBadge = milestones.find(badge => percentage < badge.threshold)

  return (
    <div className={cn('space-y-3', className)}>
      {/* Progress Header */}
      <div className="flex items-center justify-between text-sm">
        <span className="font-medium flex items-center gap-2">
          {isComplete && (
            <CheckCircle2 className="h-4 w-4 text-green-500" />
          )}
          Progress
        </span>
        <span className="text-muted-foreground">
          {completedSections}/{totalSections} sections
        </span>
      </div>

      {/* Progress Bar with Animated Gradient */}
      <div className="relative">
        <Progress
          value={percentage}
          className={cn(
            "h-2",
            percentage >= 100 && "bg-gradient-to-r from-purple-500 to-pink-500"
          )}
        />
        <div className="flex justify-between mt-1 text-xs text-muted-foreground">
          <span>{percentage}%</span>
          {nextBadge && (
            <span className="text-[10px]">
              {nextBadge.threshold - percentage}% to {nextBadge.label}
            </span>
          )}
        </div>
      </div>

      {/* Milestone Badges */}
      {earnedBadges.length > 0 && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">
            Achievements Earned
          </p>
          <div className="flex flex-wrap gap-2">
            <AnimatePresence>
              {earnedBadges.map((badge) => {
                const BadgeIcon = badge.icon
                return (
                  <motion.div
                    key={badge.id}
                    initial={{ scale: 0, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{
                      type: 'spring',
                      stiffness: 260,
                      damping: 20
                    }}
                  >
                    <Badge
                      variant="outline"
                      className={cn(
                        "flex items-center gap-1.5 px-2 py-1",
                        badge.bgColor,
                        badge.color
                      )}
                    >
                      <BadgeIcon className="h-3 w-3" />
                      <span className="text-xs font-medium">
                        {badge.label}
                      </span>
                    </Badge>
                  </motion.div>
                )
              })}
            </AnimatePresence>
          </div>
        </div>
      )}

      {/* Completion Message */}
      {isComplete && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="p-3 rounded-md bg-gradient-to-r from-purple-100 to-pink-100 dark:from-purple-950 dark:to-pink-950"
        >
          <p className="text-xs font-semibold text-purple-900 dark:text-purple-100 flex items-center gap-2">
            <Trophy className="h-4 w-4" />
            Congratulations! You&apos;ve mastered the ISX Pulse Guide!
          </p>
        </motion.div>
      )}
    </div>
  )
}
