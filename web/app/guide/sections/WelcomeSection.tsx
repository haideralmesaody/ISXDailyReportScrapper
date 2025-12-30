'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Rocket,
  BookOpen,
  Target,
  Zap,
  Shield,
  ChevronRight,
  CheckCircle2,
  LayoutGrid,
} from 'lucide-react'
import { InteractiveDemo } from '../components/InteractiveDemo'
import { CompactGuideCard } from '../components/CompactGuideCard'
import { guideSections } from '@/lib/guide/sections'
import { useGuideProgress } from '@/lib/guide/progress'

interface WelcomeSectionProps {
  onNavigate: (sectionId: string) => void
}

/**
 * Welcome section - Introduction to ISX Pulse and guide navigation
 * First section users see when entering the guide
 * NOW USES: Compact card grid layout matching Operations page
 */
export function WelcomeSection({ onNavigate }: WelcomeSectionProps) {
  const { progress } = useGuideProgress()

  // Get main sections (exclude indicator sub-pages)
  const mainSections = guideSections.filter(s => !s.hidden && s.id !== 'welcome')

  // Calculate overall completion
  const completedCount = progress.sectionsCompleted.length
  const totalCount = mainSections.length + 1 // +1 for welcome
  const completionPercentage = Math.round((completedCount / totalCount) * 100)

  const features = [
    {
      icon: LayoutGrid,
      title: 'Market Overview',
      description: 'Browse tickers, prices, and liquidity snapshots',
      href: 'market-overview'
    },
    {
      icon: Target,
      title: 'Automated Strategies',
      description: 'Build, backtest, and deploy trading strategies with confidence',
      href: 'strategy'
    },
    {
      icon: Shield,
      title: 'Secure License System',
      description: 'Hardware-locked licensing with scratch card activation',
      href: 'license'
    },
    {
      icon: Zap,
      title: '4-Stage Data Pipeline',
      description: 'Automated scraping, processing, and reporting workflow',
      href: 'pipeline'
    }
  ]

  const quickStart = [
    { step: 1, text: 'Start with the License System guide to understand activation' },
    { step: 2, text: 'Learn the Pipeline Architecture to understand data flow' },
    { step: 3, text: 'Explore Operations to run your first data collection' },
    { step: 4, text: 'Use Market Overview and Liquidity to understand the universe' },
    { step: 5, text: 'Build your first Strategy with backtesting' }
  ]

  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <div className="text-center space-y-4 py-8">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
          <Rocket className="h-8 w-8 text-primary" />
        </div>
        <h2 className="text-4xl font-bold tracking-tight">
          Welcome to ISX Pulse
        </h2>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Your comprehensive platform for Iraqi Stock Exchange analysis, automated trading strategies, and professional-grade market insights.
        </p>
        <div className="flex items-center justify-center gap-4 pt-4">
          <Badge variant="secondary" className="text-sm">
            v0.1.0-alpha.4
          </Badge>
          <Badge variant="outline" className="text-sm">
            10 Sections
          </Badge>
          <Badge variant="outline" className="text-sm">
            ~2.5 Hours
          </Badge>
        </div>

        {/* Overall Progress Bar */}
        {completionPercentage > 0 && (
          <div className="max-w-md mx-auto space-y-2 pt-4">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Overall Progress</span>
              <span className="font-medium">{completionPercentage}%</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-primary to-primary/80 transition-all duration-500"
                style={{ width: `${completionPercentage}%` }}
              />
            </div>
            <p className="text-xs text-muted-foreground">
              {completedCount} of {totalCount} sections completed
            </p>
          </div>
        )}
      </div>

      {/* All Guide Sections - Compact Grid (matching Operations page) */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-2xl font-semibold">Learning Path</h3>
          <Badge variant="outline">
            {mainSections.length} Sections
          </Badge>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
          {mainSections.map((section) => (
            <CompactGuideCard
              key={section.id}
              section={section}
              isCompleted={progress.sectionsCompleted.includes(section.id)}
              isInProgress={progress.lastVisited === section.id}
              completionPercentage={
                progress.lastVisited === section.id ? 50 : 0 // Mock progress for now
              }
              onClick={() => onNavigate(section.id)}
            />
          ))}
        </div>
      </div>

      {/* Key Features Grid */}
      <div>
        <h3 className="text-2xl font-semibold mb-4">Platform Highlights</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {features.map((feature) => {
            const Icon = feature.icon
            return (
              <Card key={feature.title} className="group hover:border-primary transition-colors cursor-pointer"
                onClick={() => onNavigate(feature.href)}
              >
                <CardHeader>
                  <div className="flex items-start gap-4">
                    <div className="p-2 rounded-lg bg-primary/10 group-hover:bg-primary/20 transition-colors">
                      <Icon className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1">
                      <CardTitle className="text-lg">{feature.title}</CardTitle>
                      <CardDescription className="mt-1.5">
                        {feature.description}
                      </CardDescription>
                    </div>
                    <ChevronRight className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
                  </div>
                </CardHeader>
              </Card>
            )
          })}
        </div>
      </div>

      {/* Quick Start Guide */}
      <InteractiveDemo
        title="Quick Start Checklist"
        description="Follow this recommended path for the best learning experience"
        variant="interactive"
      >
        <div className="space-y-3">
          {quickStart.map((item) => (
            <div
              key={item.step}
              className="flex items-start gap-3 p-3 rounded-lg bg-background border hover:border-primary transition-colors"
            >
              <div className="flex items-center justify-center w-6 h-6 rounded-full bg-primary text-primary-foreground text-sm font-semibold shrink-0">
                {item.step}
              </div>
              <p className="text-sm leading-relaxed pt-0.5">{item.text}</p>
            </div>
          ))}
        </div>
      </InteractiveDemo>

      {/* How to Use This Guide */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BookOpen className="h-5 w-5" />
            How to Use This Guide
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <h4 className="font-semibold text-sm">1. Navigation</h4>
              <p className="text-sm text-muted-foreground">
                Use the sidebar to jump between sections. On mobile, tap the menu icon to access navigation.
              </p>
            </div>
            <div className="space-y-2">
              <h4 className="font-semibold text-sm">2. Progress Tracking</h4>
              <p className="text-sm text-muted-foreground">
                Mark sections complete as you go. Your progress is saved automatically in your browser.
              </p>
            </div>
            <div className="space-y-2">
              <h4 className="font-semibold text-sm">3. Interactive Demos</h4>
              <p className="text-sm text-muted-foreground">
                Try hands-on demos throughout the guide to reinforce your learning with practical experience.
              </p>
            </div>
          </div>

          <div className="pt-4 border-t">
            <div className="flex items-start gap-3 text-sm">
              <CheckCircle2 className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
              <div className="space-y-1">
                <p className="font-medium">Complete sections in order for prerequisites</p>
                <p className="text-muted-foreground">
                  Some advanced sections require completion of basics. Locked sections will unlock as you progress.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Get Started CTA */}
      <div className="text-center space-y-4 py-8">
        <h3 className="text-2xl font-semibold">Ready to get started?</h3>
        <p className="text-muted-foreground">
          Begin your journey with the License System guide to understand how ISX Pulse activation works.
        </p>
        <Button
          size="lg"
          onClick={() => onNavigate('license')}
          className="gap-2"
        >
          Start with License System
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
