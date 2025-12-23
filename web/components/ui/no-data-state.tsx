'use client'

import * as React from "react"
import { type LucideIcon } from "lucide-react"
import Link from "next/link"

import { cn } from "@/lib/utils"
import { Card } from "@/components/ui/card"
import { Button, type ButtonProps } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Info } from "lucide-react"
import { 
  trackNoDataDisplayed, 
  trackNoDataAction, 
  getCurrentPage,
  generateCorrelationId,
  debug,
  type NoDataStateContext 
} from "@/lib/observability/no-data-metrics"

export interface NoDataAction {
  label: string
  variant?: ButtonProps['variant']
  href?: string
  onClick?: () => void
  icon?: LucideIcon
}

export interface NoDataStateProps {
  icon?: LucideIcon
  iconColor?: 'blue' | 'green' | 'purple' | 'orange' | 'red' | 'gray'
  title: string
  description: string
  instructions?: string[]
  actions?: NoDataAction[]
  className?: string
  // Observability props
  page?: string
  reason?: string
  componentName?: string
  onDisplayed?: (context: NoDataStateContext) => void
}

const iconColorClasses = {
  blue: 'bg-blue-100 text-blue-600',
  green: 'bg-green-100 text-green-600',
  purple: 'bg-purple-100 text-purple-600',
  orange: 'bg-orange-100 text-orange-600',
  red: 'bg-red-100 text-red-600',
  gray: 'bg-gray-100 text-gray-600',
}

export const NoDataState = React.forwardRef<
  HTMLDivElement,
  NoDataStateProps
>(({ 
  icon: Icon, 
  iconColor = 'blue', 
  title, 
  description, 
  instructions, 
  actions, 
  className,
  page,
  reason = 'unknown',
  componentName = 'NoDataState',
  onDisplayed,
  ...props 
}, ref) => {
  // Generate correlation ID for this component instance
  const correlationId = React.useMemo(() => generateCorrelationId(), [])
  const currentPage = page || getCurrentPage()

  // Track when component is displayed
  React.useEffect(() => {
    const context: NoDataStateContext = {
      page: currentPage,
      component_name: componentName,
      display_reason: reason,
      actions_available: actions?.map(action => action.label) || [],
      instructions_count: instructions?.length || 0,
    }

    // Track display event
    trackNoDataDisplayed(context)

    // Debug logging in development
    debug.logComponentState('NoDataState', {
      correlation_id: correlationId,
      page: currentPage,
      reason,
      title,
      description,
      has_actions: (actions?.length || 0) > 0,
      has_instructions: (instructions?.length || 0) > 0,
      icon_color: iconColor,
    })

    // Call custom callback if provided
    onDisplayed?.(context)
  }, [
    currentPage,
    reason,
    componentName,
    actions,
    instructions,
    correlationId,
    title,
    description,
    iconColor,
    onDisplayed
  ])

  // Enhanced action click handler with tracking
  const handleActionClick = React.useCallback((action: NoDataAction, index: number) => {
    // Track the action click
    trackNoDataAction(currentPage, action.label, action.href)

    // Debug logging
    debug.logComponentState('NoDataState Action Click', {
      correlation_id: correlationId,
      action_label: action.label,
      action_variant: action.variant,
      has_href: !!action.href,
      has_onclick: !!action.onClick,
      action_index: index,
    })

    // Execute original onClick if provided
    action.onClick?.()
  }, [currentPage, correlationId])
  return (
    <div 
      ref={ref}
      className={cn("min-h-screen p-8", className)}
      {...props}
    >
      <div className="max-w-7xl mx-auto">
        <Card className="p-8">
          <div className="text-center space-y-6" role="alert" aria-live="polite">
            {Icon && (
              <div className="flex justify-center">
                <div className={cn(
                  "p-4 rounded-full",
                  iconColorClasses[iconColor]
                )}>
                  <Icon className="h-12 w-12" />
                </div>
              </div>
            )}
            
            <div className="space-y-2">
              <h2 className="text-2xl font-semibold">{title}</h2>
              <p className="text-muted-foreground max-w-md mx-auto">
                {description}
              </p>
            </div>
            
            {instructions && instructions.length > 0 && (
              <Alert className="max-w-md mx-auto">
                <Info className="h-4 w-4" />
                <AlertTitle>How to get started:</AlertTitle>
                <AlertDescription className="text-left mt-2">
                  <ol className="list-decimal list-inside space-y-1">
                    {instructions.map((instruction, index) => (
                      <li key={index}>{instruction}</li>
                    ))}
                  </ol>
                </AlertDescription>
              </Alert>
            )}
            
            {actions && actions.length > 0 && (
              <div className="flex gap-4 justify-center flex-wrap">
                {actions.map((action, index) => {
                  const ButtonContent = () => (
                    <>
                      {action.icon && <action.icon className="h-4 w-4 mr-2" />}
                      {action.label}
                    </>
                  )

                  if (action.href) {
                    return (
                      <Button key={index} asChild variant={action.variant}>
                        <Link 
                          href={action.href}
                          onClick={() => handleActionClick(action, index)}
                        >
                          <ButtonContent />
                        </Link>
                      </Button>
                    )
                  }

                  return (
                    <Button 
                      key={index} 
                      onClick={() => handleActionClick(action, index)} 
                      variant={action.variant}
                    >
                      <ButtonContent />
                    </Button>
                  )
                })}
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  )
})

NoDataState.displayName = "NoDataState"