'use client'

import { useState, useEffect, useCallback } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import { GuideNavigation } from './components/GuideNavigation'
import { GuideSection } from './components/GuideSection'
import { WelcomeSection } from './sections/WelcomeSection'
import { LicenseSection } from './sections/LicenseSection'
import { PipelineSection } from './sections/PipelineSection'
import { ReportsSection } from './sections/ReportsSection'
import { LiquidityCalculationsSection } from './sections/LiquidityCalculationsSection'
import { MarketOverviewSection } from './sections/MarketOverviewSection'
import { StrategySection } from './sections/StrategySection'
import { QuickReferenceSection } from './sections/QuickReferenceSection'
import { useGuideProgress } from '@/lib/guide/progress'
import {
  getSectionById,
  getNextSection,
  getPreviousSection
} from '@/lib/guide/sections'
import { Card, CardContent } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { AlertCircle, Loader2, Info, ArrowRight, X } from 'lucide-react'

/**
 * Main client component for interactive guide
 * Handles section routing, progress tracking, and deep linking
 */
export default function GuideClient() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const {
    progress,
    isHydrated,
    markSectionComplete,
    markSectionIncomplete,
    setLastVisited,
    toggleBookmark
  } = useGuideProgress()

  // Get initial section from URL or default to welcome
  const initialSection = searchParams?.get('section') || 'welcome'
  const [currentSection, setCurrentSection] = useState(initialSection)
  const [showResumeBanner, setShowResumeBanner] = useState(false)
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)

  // Load sidebar collapse state from localStorage
  useEffect(() => {
    if (isHydrated) {
      const saved = localStorage.getItem('guide-sidebar-collapsed')
      if (saved) {
        setIsSidebarCollapsed(saved === 'true')
      }
    }
  }, [isHydrated])

  // Toggle sidebar collapse
  const toggleSidebar = useCallback(() => {
    setIsSidebarCollapsed(prev => {
      const newState = !prev
      localStorage.setItem('guide-sidebar-collapsed', String(newState))
      return newState
    })
  }, [])

  // Show resume banner if user returns and has different last visited section
  useEffect(() => {
    if (isHydrated && progress.lastVisited && progress.lastVisited !== currentSection) {
      // Check sessionStorage to see if banner was dismissed
      const dismissed = sessionStorage.getItem('guide-resume-banner-dismissed')
      if (!dismissed) {
        setShowResumeBanner(true)
      }
    }
  }, [isHydrated, progress.lastVisited, currentSection])

  // Update last visited when section changes
  useEffect(() => {
    if (isHydrated) {
      setLastVisited(currentSection)
    }
  }, [currentSection, isHydrated, setLastVisited])

  // Deep linking support - update URL when section changes
  const handleSectionChange = useCallback((sectionId: string) => {
    setCurrentSection(sectionId)
    router.push(`/guide?section=${sectionId}`, { scroll: false })
  }, [router])

  // Migrate legacy section IDs (removed sections)
  useEffect(() => {
    if (!isHydrated) return
    if (currentSection === 'charts') {
      setCurrentSection('strategy')
      router.replace('/guide?section=strategy', { scroll: false })
    }
  }, [currentSection, isHydrated, router])

  // Navigation handlers
  const handleNext = useCallback(() => {
    const nextSection = getNextSection(currentSection)
    if (nextSection) {
      handleSectionChange(nextSection.id)
    }
  }, [currentSection, handleSectionChange])

  const handlePrevious = useCallback(() => {
    const previousSection = getPreviousSection(currentSection)
    if (previousSection) {
      handleSectionChange(previousSection.id)
    }
  }, [currentSection, handleSectionChange])

  const handleComplete = useCallback(() => {
    const isCompleted = progress.sectionsCompleted.includes(currentSection)
    if (isCompleted) {
      markSectionIncomplete(currentSection)
    } else {
      markSectionComplete(currentSection)
    }
  }, [currentSection, progress.sectionsCompleted, markSectionComplete, markSectionIncomplete])

  const handleBookmark = useCallback(() => {
    toggleBookmark(currentSection)
  }, [currentSection, toggleBookmark])

  const handleResume = useCallback(() => {
    if (progress.lastVisited) {
      handleSectionChange(progress.lastVisited)
      setShowResumeBanner(false)
    }
  }, [progress.lastVisited, handleSectionChange])

  const handleDismissBanner = useCallback(() => {
    setShowResumeBanner(false)
    sessionStorage.setItem('guide-resume-banner-dismissed', 'true')
  }, [])

  // Get section metadata
  const section = getSectionById(currentSection)
  const nextSection = getNextSection(currentSection)
  const previousSection = getPreviousSection(currentSection)
  const optionalSectionProps = {
    ...(section?.estimatedMinutes !== undefined ? { estimatedMinutes: section.estimatedMinutes } : {}),
    ...(nextSection ? { onNext: handleNext, nextSectionTitle: nextSection.title } : {}),
    ...(previousSection ? { onPrevious: handlePrevious, previousSectionTitle: previousSection.title } : {}),
  }

  // Show loading state while hydrating (prevents hydration mismatch)
  if (!isHydrated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
          <p className="text-muted-foreground">Loading guide...</p>
        </div>
      </div>
    )
  }

  // Show error if section not found
  if (!section) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="max-w-md">
          <CardContent className="pt-6">
            <div className="flex items-start gap-4">
              <AlertCircle className="h-5 w-5 text-destructive shrink-0 mt-0.5" />
              <div className="space-y-2">
                <h3 className="font-semibold">Section Not Found</h3>
                <p className="text-sm text-muted-foreground">
                  The section &quot;{currentSection}&quot; does not exist. Please select a valid section from the navigation.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Render current section content
  const renderSectionContent = () => {
    switch (currentSection) {
      case 'welcome':
        return <WelcomeSection onNavigate={handleSectionChange} />

      case 'license':
        return <LicenseSection onNavigate={handleSectionChange} />

      case 'pipeline':
        return <PipelineSection onNavigate={handleSectionChange} />

      case 'reports':
        return <ReportsSection onNavigate={handleSectionChange} />

      case 'liquidity-calculations':
        return <LiquidityCalculationsSection onNavigate={handleSectionChange} />

      case 'market-overview':
        return <MarketOverviewSection />

      case 'strategy':
        return <StrategySection />

      case 'quick-reference':
        return <QuickReferenceSection />

      default:
        return null
    }
  }

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Navigation Sidebar */}
      <GuideNavigation
        currentSection={currentSection}
        onSectionChange={handleSectionChange}
        completedSections={progress.sectionsCompleted}
        isCollapsed={isSidebarCollapsed}
        onToggleCollapse={toggleSidebar}
      />

      {/* Main Content Area */}
      <main className="flex-1 overflow-y-auto p-8">
        <div className="max-w-7xl mx-auto space-y-8">
          {/* Resume Banner */}
          {showResumeBanner && progress.lastVisited && (
            <Alert className="border-primary/50 bg-primary/5">
              <Info className="h-4 w-4" />
              <AlertTitle className="flex items-center justify-between">
                <span>Resume Your Learning Journey</span>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 -mt-1"
                  onClick={handleDismissBanner}
                >
                  <X className="h-4 w-4" />
                </Button>
              </AlertTitle>
              <AlertDescription className="space-y-3">
                <p className="text-sm">
                  You last visited: <strong>{getSectionById(progress.lastVisited)?.title}</strong>
                </p>
                <div className="flex gap-2">
                  <Button size="sm" onClick={handleResume}>
                    <ArrowRight className="h-4 w-4 mr-1" />
                    Resume
                  </Button>
                  <Button size="sm" variant="outline" onClick={handleDismissBanner}>
                    Stay Here
                  </Button>
                </div>
              </AlertDescription>
            </Alert>
          )}

          <GuideSection
            id={section.id}
            title={section.title}
            description={section.description}
            isCompleted={progress.sectionsCompleted.includes(section.id)}
            isBookmarked={progress.bookmarks.includes(section.id)}
            onComplete={handleComplete}
            onBookmark={handleBookmark}
            {...optionalSectionProps}
          >
            {renderSectionContent()}
          </GuideSection>
        </div>
      </main>
    </div>
  )
}
