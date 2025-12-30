/**
 * Operation UI Walkthrough
 * Step-by-step interactive guide to using the Operations page
 * Shows users exactly where to click and what to expect
 */

'use client'

import React, { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  ChevronLeft,
  ChevronRight,
  MousePointer2,
  CheckCircle2,
  Download,
  Settings2,
  Zap,
  Calendar,
  Activity
} from 'lucide-react'
import { cn } from '@/lib/utils'

const walkthroughSteps = [
  {
    step: 1,
    title: 'Find Your Operation',
    description: 'Operations are displayed as compact cards in a grid layout. Each card shows the operation name, description, and current status.',
    visual: {
      icon: Download,
      cardTitle: 'Scrape ISX',
      cardDescription: 'Download daily trading reports',
      hasDateBadge: true,
      buttonType: 'configure'
    },
    highlights: [
      { element: 'Icon', description: 'Shows operation type' },
      { element: 'Title', description: 'Operation name' },
      { element: 'Date Badge', description: 'Configured date range' },
      { element: 'Status Line', description: 'Blue = configured, Gray = needs setup' }
    ]
  },
  {
    step: 2,
    title: 'Check Configuration',
    description: 'Some operations require date selection (Scraping, Full Pipeline). Others can run immediately (Process, Index).',
    visual: {
      icon: Calendar,
      cardTitle: 'Scrape ISX',
      cardDescription: 'Download daily trading reports',
      hasDateBadge: true,
      dateRange: 'Jan 1 - Jan 10 (10d)',
      buttonType: 'configure'
    },
    highlights: [
      { element: 'Calendar Icon', description: 'Indicates saved date range' },
      { element: 'Date Text', description: 'Shows configured dates and day count' },
      { element: 'Blue Line', description: 'Top indicator shows configuration saved' }
    ]
  },
  {
    step: 3,
    title: 'Choose Start Method',
    description: 'Two ways to start operations: Quick Start (immediate) or Configure & Start (opens settings panel).',
    visual: {
      icon: Zap,
      cardTitle: 'Process Data',
      cardDescription: 'Convert Excel to CSV',
      hasDateBadge: false,
      buttonType: 'quickstart'
    },
    highlights: [
      { element: '⚡ Quick Start', description: 'No configuration needed, starts immediately' },
      { element: '⚙️ Configure & Start', description: 'Opens panel to select dates/options' }
    ]
  },
  {
    step: 4,
    title: 'Monitor Real-Time Progress',
    description: 'Once started, a progress card appears below. WebSocket connection provides live updates as the operation executes.',
    visual: {
      icon: Activity,
      cardTitle: 'Operation Running',
      cardDescription: 'Real-time progress updates',
      hasDateBadge: false,
      buttonType: 'progress',
      progressPercent: 65
    },
    highlights: [
      { element: 'Green Dot', description: 'WebSocket connected (top of page)' },
      { element: 'Progress Bar', description: 'Updates automatically' },
      { element: 'Stage Name', description: 'Shows current operation stage' },
      { element: 'Files Count', description: 'Files processed so far' }
    ]
  },
  {
    step: 5,
    title: 'Review Completion',
    description: 'When complete, a summary card shows all outputs, duration, and success/error details. You can start the next suggested operation with one click.',
    visual: {
      icon: CheckCircle2,
      cardTitle: 'Operation Complete',
      cardDescription: 'Successfully processed 10 files',
      hasDateBadge: false,
      buttonType: 'complete'
    },
    highlights: [
      { element: '✓ Success Badge', description: 'Operation completed successfully' },
      { element: 'Files Processed', description: 'Total files handled' },
      { element: 'Duration', description: 'Total execution time' },
      { element: 'Next Operation', description: 'Suggested next step button' }
    ]
  }
]

export function OperationUIWalkthrough() {
  const [currentStep, setCurrentStep] = useState(0)
  const step = walkthroughSteps[currentStep]!
  const Icon = step.visual.icon

  const nextStep = () => {
    if (currentStep < walkthroughSteps.length - 1) {
      setCurrentStep(currentStep + 1)
    }
  }

  const prevStep = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  return (
    <div className="space-y-6">
      {/* Progress indicator */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {walkthroughSteps.map((_, index) => (
            <div
              key={index}
              className={cn(
                "h-2 rounded-full transition-all duration-300",
                index === currentStep ? "w-8 bg-primary" : "w-2 bg-muted",
                index < currentStep && "bg-green-500"
              )}
            />
          ))}
        </div>
        <Badge variant="secondary">
          Step {currentStep + 1} of {walkthroughSteps.length}
        </Badge>
      </div>

      {/* Main content */}
      <div className="grid md:grid-cols-2 gap-6">
        {/* Left: Visual mockup */}
        <Card className="relative overflow-hidden">
          <CardContent className="p-6">
            <div className="space-y-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-4">
                <MousePointer2 className="h-4 w-4" />
                <span>Interactive Preview</span>
              </div>

              {/* Mock operation card */}
              <div className={cn(
                "group relative overflow-hidden rounded-lg border bg-card transition-all duration-200",
                "hover:shadow-lg",
                "h-[180px] flex flex-col"
              )}>
                {/* Status indicator line */}
                <div className={cn(
                  "absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r",
                  step.visual.hasDateBadge
                    ? "from-primary to-primary/50"
                    : "from-muted to-muted"
                )} />

                {/* Card header */}
                <div className="pb-2 pt-4 px-4 flex-none">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-primary/10">
                        <Icon className="h-5 w-5 text-primary" />
                      </div>
                      <div>
                        <div className="text-base font-semibold">
                          {step.visual.cardTitle}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Card content */}
                <div className="flex-1 flex flex-col justify-between px-4 pb-4">
                  <div className="space-y-2">
                    <p className="text-xs text-muted-foreground line-clamp-2">
                      {step.visual.cardDescription}
                    </p>

                    {/* Date range badge */}
                    {step.visual.hasDateBadge && step.visual.dateRange && (
                      <div className="flex items-center gap-1.5 text-xs">
                        <Calendar className="h-3 w-3 text-muted-foreground" />
                        <span className="text-muted-foreground font-medium">
                          {step.visual.dateRange}
                        </span>
                      </div>
                    )}

                    {/* Progress bar (for step 4) */}
                    {step.visual.buttonType === 'progress' && step.visual.progressPercent && (
                      <div className="space-y-1">
                        <div className="flex items-center justify-between text-xs">
                          <span className="text-muted-foreground">Processing...</span>
                          <span className="font-medium">{step.visual.progressPercent}%</span>
                        </div>
                        <div className="h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-primary transition-all duration-500"
                            style={{ width: `${step.visual.progressPercent}%` }}
                          />
                        </div>
                      </div>
                    )}

                    {/* Success badge (for step 5) */}
                    {step.visual.buttonType === 'complete' && (
                      <div className="flex items-center gap-2 text-xs">
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                        <span className="text-green-600 dark:text-green-400 font-medium">
                          Completed successfully
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Action button */}
                  {step.visual.buttonType !== 'progress' && step.visual.buttonType !== 'complete' && (
                    <Button size="sm" className="w-full mt-3">
                      {step.visual.buttonType === 'quickstart' ? (
                        <>
                          <Zap className="h-3.5 w-3.5 mr-1.5" />
                          Quick Start
                        </>
                      ) : (
                        <>
                          <Settings2 className="h-3.5 w-3.5 mr-1.5" />
                          Configure & Start
                        </>
                      )}
                      <ChevronRight className="h-3 w-3 ml-auto opacity-50" />
                    </Button>
                  )}
                </div>

                {/* Hover effect */}
                <div className="absolute inset-0 bg-gradient-to-t from-primary/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Right: Step description */}
        <div className="space-y-4">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground font-bold">
                {step.step}
              </div>
              <h3 className="text-xl font-semibold">{step.title}</h3>
            </div>
            <p className="text-muted-foreground">{step.description}</p>
          </div>

          {/* Highlights */}
          <div className="space-y-3">
            <h4 className="font-medium text-sm flex items-center gap-2">
              <MousePointer2 className="h-4 w-4 text-primary" />
              Key Elements
            </h4>
            <div className="space-y-2">
              {step.highlights.map((highlight, index) => (
                <div
                  key={index}
                  className="flex items-start gap-3 p-3 rounded-lg bg-muted/50 border border-border/50"
                >
                  <div className="h-1.5 w-1.5 rounded-full bg-primary shrink-0 mt-2" />
                  <div>
                    <div className="font-medium text-sm">{highlight.element}</div>
                    <div className="text-xs text-muted-foreground">{highlight.description}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Navigation buttons */}
      <div className="flex items-center justify-between pt-4">
        <Button
          variant="outline"
          onClick={prevStep}
          disabled={currentStep === 0}
        >
          <ChevronLeft className="h-4 w-4 mr-2" />
          Previous
        </Button>

        <Button
          onClick={nextStep}
          disabled={currentStep === walkthroughSteps.length - 1}
        >
          Next
          <ChevronRight className="h-4 w-4 ml-2" />
        </Button>
      </div>
    </div>
  )
}

export default OperationUIWalkthrough
