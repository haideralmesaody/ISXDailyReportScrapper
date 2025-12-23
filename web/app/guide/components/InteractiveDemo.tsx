'use client'

import { ReactNode } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Lightbulb } from 'lucide-react'
import { cn } from '@/lib/utils'

interface InteractiveDemoProps {
  title: string
  description: string
  children: ReactNode
  variant?: 'default' | 'interactive' | 'example'
  className?: string
}

/**
 * Container component for interactive demonstrations
 * Provides consistent styling and visual hierarchy
 */
export function InteractiveDemo({
  title,
  description,
  children,
  variant = 'default',
  className
}: InteractiveDemoProps) {
  const variantStyles = {
    default: 'border-border',
    interactive: 'border-blue-500/50 bg-blue-50/50 dark:bg-blue-950/20',
    example: 'border-purple-500/50 bg-purple-50/50 dark:bg-purple-950/20'
  }

  const variantBadges = {
    default: null,
    interactive: (
      <Badge variant="outline" className="gap-1 border-blue-500 text-blue-600 dark:text-blue-400">
        <Lightbulb className="h-3 w-3" />
        Interactive
      </Badge>
    ),
    example: (
      <Badge variant="outline" className="gap-1 border-purple-500 text-purple-600 dark:text-purple-400">
        Example
      </Badge>
    )
  }

  return (
    <Card className={cn(variantStyles[variant], 'overflow-hidden', className)}>
      <CardHeader>
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1.5 flex-1">
            <CardTitle className="text-xl">{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          {variantBadges[variant]}
        </div>
      </CardHeader>
      <CardContent>
        {children}
      </CardContent>
    </Card>
  )
}
