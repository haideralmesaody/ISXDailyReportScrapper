/**
 * Operation Expandable Card
 * Interactive card for Operations Guide with expandable details
 * Closed: Shows operation overview
 * Expanded: Shows detailed content (architecture, tutorial, tips)
 */

'use client'

import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { ChevronDown, ChevronUp, Settings2, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { LucideIcon } from 'lucide-react'

export interface Operation {
  id: string
  name: string
  icon: LucideIcon
  color: string
  description: string
  duration: string
  needsConfig: boolean
}

interface OperationExpandableCardProps {
  operation: Operation
  children: React.ReactNode
  defaultOpen?: boolean
  fullWidth?: boolean
  className?: string
}

export function OperationExpandableCard({
  operation,
  children,
  defaultOpen = false,
  fullWidth = false,
  className
}: OperationExpandableCardProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen)
  const Icon = operation.icon

  return (
    <Collapsible
      open={isOpen}
      onOpenChange={setIsOpen}
      className={cn('group', fullWidth && 'md:col-span-2 lg:col-span-3', className)}
    >
      <Card className={cn(
        'relative overflow-hidden transition-all duration-200',
        'hover:shadow-lg',
        isOpen && 'ring-2 ring-primary/50'
      )}>
        {/* Status indicator line */}
        <div className={`absolute top-0 left-0 right-0 h-0.5 ${operation.color}`} />

        <CollapsibleTrigger className="w-full text-left">
          <CardContent className="pt-4 pb-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3 flex-1">
                <div className={cn(
                  'p-2 rounded-lg bg-opacity-10',
                  operation.color
                )}>
                  <Icon className="h-5 w-5" />
                </div>
                <div className="flex-1">
                  <div className="font-semibold text-base mb-1">{operation.name}</div>
                  <p className="text-sm text-muted-foreground">{operation.description}</p>
                </div>
              </div>
              <div className="flex items-center gap-3 ml-4">
                <Badge variant="secondary" className="text-xs">{operation.duration}</Badge>
                {operation.needsConfig ? (
                  <Settings2 className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Zap className="h-4 w-4 text-green-500" />
                )}
                {isOpen ? (
                  <ChevronUp className="h-5 w-5 text-muted-foreground shrink-0" />
                ) : (
                  <ChevronDown className="h-5 w-5 text-muted-foreground shrink-0" />
                )}
              </div>
            </div>
          </CardContent>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <CardContent className="pt-0 pb-4 border-t">
            <div className="mt-4">
              {children}
            </div>
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}

export default OperationExpandableCard
