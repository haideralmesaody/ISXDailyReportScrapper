'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { Menu, X, CheckCircle2, Circle, PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { guideSections } from '@/lib/guide/sections'
import { ProgressStats } from './ProgressStats'
import { TradingCalendarWidget } from './TradingCalendarWidget'
import { useGuideProgress } from '@/lib/guide/progress'

interface GuideNavigationProps {
  currentSection: string
  onSectionChange: (sectionId: string) => void
  completedSections: string[]
  isCollapsed?: boolean
  onToggleCollapse?: () => void
  className?: string
}

/**
 * Navigation sidebar for guide sections
 * Desktop: Fixed sidebar (280px)
 * Tablet/Mobile: Collapsible sheet
 */
export function GuideNavigation({
  currentSection,
  onSectionChange,
  completedSections,
  isCollapsed = false,
  onToggleCollapse,
  className
}: GuideNavigationProps) {
  const { progress } = useGuideProgress()
  const [isMobileOpen, setIsMobileOpen] = useState(false)

  // Get all visible sections (hidden flag removed, all sections now visible)
  const visibleSections = guideSections.filter(section => !section.hidden)

  // Count total sections for progress stats
  const totalCountableSections = visibleSections.length

  const handleSectionClick = (sectionId: string) => {
    onSectionChange(sectionId)
    setIsMobileOpen(false) // Close mobile menu after selection
  }

  const navigationContent = (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Section List */}
      <ScrollArea className="flex-1 px-3 py-4 overflow-auto">
        <div className="space-y-1">
          {visibleSections.map((section) => {
            const Icon = section.icon
            const isActive = currentSection === section.id
            const isCompleted = completedSections.includes(section.id)

            return (
              <div key={section.id}>
                <Button
                  variant={isActive ? 'secondary' : 'ghost'}
                  className={cn(
                    "flex-1 h-auto py-3",
                    isCollapsed ? "justify-center px-2" : "justify-start gap-3 px-3",
                    !isCollapsed && "ml-5"
                  )}
                  onClick={() => handleSectionClick(section.id)}
                  title={isCollapsed ? section.title : undefined}
                >
                  <div className={cn(
                    "flex items-center flex-1",
                    isCollapsed ? "justify-center" : "gap-3"
                  )}>
                    <Icon className="h-4 w-4 shrink-0" />
                    {!isCollapsed && (
                      <>
                        <div className="flex-1 text-left">
                          <div className="font-medium text-sm leading-tight">
                            {section.title}
                          </div>
                          <div className="text-xs text-muted-foreground mt-0.5 line-clamp-1">
                            {section.description}
                          </div>
                        </div>
                        <div className="shrink-0">
                          {isCompleted ? (
                            <CheckCircle2 className="h-4 w-4 text-green-500" />
                          ) : (
                            <Circle className="h-4 w-4 text-muted-foreground" />
                          )}
                        </div>
                      </>
                    )}
                  </div>
                </Button>
              </div>
            )
          })}
        </div>
      </ScrollArea>

      {!isCollapsed && <Separator />}

      {/* Progress Statistics */}
      {!isCollapsed && (
        <div className="p-4">
          <ProgressStats
            progress={progress}
            completedSections={completedSections.length}
            totalSections={totalCountableSections}
          />
        </div>
      )}

      {!isCollapsed && <Separator />}

      {/* Trading Calendar Widget */}
      {!isCollapsed && (
        <div className="p-4">
          <TradingCalendarWidget compact />
        </div>
      )}
    </div>
  )

  return (
    <>
      {/* Mobile Trigger Button */}
      <div className="lg:hidden fixed top-20 left-4 z-40">
        <Sheet open={isMobileOpen} onOpenChange={setIsMobileOpen}>
          <SheetTrigger asChild>
            <Button size="icon" variant="outline" className="shadow-lg">
              {isMobileOpen ? (
                <X className="h-5 w-5" />
              ) : (
                <Menu className="h-5 w-5" />
              )}
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[300px] p-0">
            <SheetHeader className="p-4 pb-0">
              <SheetTitle>Guide Navigation</SheetTitle>
            </SheetHeader>
            {navigationContent}
          </SheetContent>
        </Sheet>
      </div>

      {/* Desktop Sidebar */}
      <aside
        className={cn(
          'hidden lg:block h-screen border-r bg-background transition-all duration-300',
          isCollapsed ? 'w-[60px]' : 'w-[280px]',
          className
        )}
      >
        {/* Header with Collapse Button */}
        {onToggleCollapse && (
          <div className="h-12 border-b flex items-center justify-center px-2 shrink-0">
            <Button
              variant="ghost"
              size="icon"
              onClick={onToggleCollapse}
              className="h-8 w-8"
              title={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
              {isCollapsed ? (
                <PanelLeftOpen className="h-4 w-4" />
              ) : (
                <PanelLeftClose className="h-4 w-4" />
              )}
            </Button>
          </div>
        )}

        {/* Content Area (adjusted for header) */}
        <div className="flex flex-col h-[calc(100vh-3rem)] overflow-hidden">
          {navigationContent}
        </div>
      </aside>
    </>
  )
}
