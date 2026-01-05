/**
 * Analysis Page - Server Component with SEO metadata
 * Following CLAUDE.md Next.js 14 server component patterns
 * Uses dynamic import with SSR disabled to prevent hydration issues
 * Uses TradingView Lightweight Charts for professional technical analysis
 */

import dynamic from 'next/dynamic'
import { Card } from '@/components/ui/card'
import { Loader2, ChartCandlestick } from 'lucide-react'

// Dynamic import with SSR disabled to prevent hydration issues - CLAUDE.md pattern
const AnalysisClient = dynamic(() => import('./analysis-client').catch(err => {
  console.error('[AnalysisPage] Dynamic import failed:', err)
  throw err
}), {
  ssr: false,
  loading: () => <AnalysisPageSkeleton />
})

// SEO metadata exported from server component - CLAUDE.md requirement
export const metadata = {
  title: 'Technical Analysis - ISX Pulse',
  description: 'Advanced technical analysis with TradingView Lightweight Charts, real-time market data, and comprehensive trading indicators for the Iraqi Stock Exchange.',
  keywords: 'ISX technical analysis, Iraqi stock market charts, candlestick patterns, trading indicators, market trends, TradingView charts, financial analysis',
  authors: [{ name: 'ISX Pulse Team' }],
  robots: {
    index: true,
    follow: true
  },
  openGraph: {
    title: 'Technical Analysis - ISX Pulse',
    description: 'Professional-grade technical analysis tools with TradingView Lightweight Charts for the Iraqi Stock Exchange.',
    type: 'website',
  }
}

// Loading skeleton that matches the analysis page layout
function AnalysisPageSkeleton() {
  console.log('[AnalysisPageSkeleton] Loading skeleton rendering')
  return (
    <div className="min-h-screen p-8">
      <div className="max-w-full">
        {/* Header Skeleton */}
        <div className="mb-6">
          <div className="h-9 w-64 bg-muted rounded-md animate-pulse mb-2" />
          <div className="h-5 w-96 bg-muted rounded-md animate-pulse" />
        </div>

        {/* Main Content Area */}
        <div className="flex gap-4 h-[calc(100vh-200px)]">
          {/* Left Panel - Ticker List Skeleton */}
          <Card className="w-[35%] p-4">
            <div className="mb-4">
              <div className="h-6 w-32 bg-muted rounded-md animate-pulse mb-2" />
              <div className="h-4 w-24 bg-muted rounded-md animate-pulse" />
            </div>

            {/* Search Bar Skeleton */}
            <div className="h-9 w-full bg-muted rounded-md animate-pulse mb-4" />

            {/* Table Header Skeleton */}
            <div className="grid grid-cols-12 gap-2 mb-2">
              <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
              <div className="col-span-3 h-4 bg-muted rounded animate-pulse" />
              <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
              <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
              <div className="col-span-1 h-4 bg-muted rounded animate-pulse" />
              <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
            </div>

            {/* Table Rows Skeleton */}
            {[...Array(10)].map((_, i) => (
              <div key={i} className="grid grid-cols-12 gap-2 py-2 border-b">
                <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
                <div className="col-span-3 h-3 bg-muted rounded animate-pulse" />
                <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
                <div className="col-span-2 h-4 bg-muted rounded animate-pulse" />
                <div className="col-span-1 h-3 bg-muted rounded animate-pulse" />
                <div className="col-span-2 h-3 bg-muted rounded animate-pulse" />
              </div>
            ))}
          </Card>

          {/* Right Panel - Chart Skeleton */}
          <Card className="flex-1 p-4">
            <div className="h-full flex flex-col">
              {/* Chart Header */}
              <div className="mb-4">
                <div className="flex items-center justify-between mb-2">
                  <div className="h-7 w-48 bg-muted rounded-md animate-pulse" />
                  <div className="flex gap-2">
                    <div className="h-8 w-20 bg-muted rounded-md animate-pulse" />
                    <div className="h-8 w-20 bg-muted rounded-md animate-pulse" />
                    <div className="h-8 w-20 bg-muted rounded-md animate-pulse" />
                  </div>
                </div>
              </div>

              {/* Chart Area */}
              <div className="flex-1 bg-muted/20 rounded-lg flex items-center justify-center">
                <div className="text-center">
                  <ChartCandlestick className="h-16 w-16 text-muted-foreground mx-auto mb-4 animate-pulse" />
                  <div className="h-5 w-48 bg-muted rounded-md animate-pulse mx-auto mb-2" />
                  <div className="h-4 w-64 bg-muted rounded-md animate-pulse mx-auto" />
                </div>
              </div>
            </div>
          </Card>
        </div>

        {/* Loading Indicator */}
        <div className="fixed bottom-8 right-8">
          <div className="bg-background border rounded-lg p-4 shadow-lg">
            <div className="flex items-center gap-3">
              <Loader2 className="h-5 w-5 animate-spin text-primary" />
              <span className="text-sm text-muted-foreground">
                Loading technical analysis tools...
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function AnalysisPage() {
  console.log('[AnalysisPage] Server component rendering')
  return <AnalysisClient />
}