/**
 * Professional License Page - Server Component
 * Provides SEO metadata and loading shell for license management
 */

import dynamic from 'next/dynamic'
import { Loader2, Shield, Award } from 'lucide-react'
import { ThemeToggle } from '@/components/ui/theme-toggle'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

// Dynamic import with SSR disabled to prevent hydration errors
const LicenseContent = dynamic(() => import('./license-content'), {
  ssr: false,
  loading: () => <LicenseLoadingSkeleton />
})

// SEO metadata for the license page
export const metadata = {
  title: 'Professional License - ISX Pulse',
  description: 'Activate and manage your professional license for comprehensive Iraqi Stock Exchange market intelligence.',
  robots: { index: false, follow: false }
}

/**
 * Comprehensive loading skeleton that matches the license page layout
 */
function LicenseLoadingSkeleton() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-background/95 dark:from-background dark:to-background">
      {/* Theme Toggle for License Page */}
      <div className="absolute top-4 right-4 z-50">
        <ThemeToggle />
      </div>
      <div className="investor-container py-12">
        {/* Professional Header Skeleton */}
        <div className="text-center space-y-4 mb-12">
          <div className="flex justify-center">
            <div className="h-16 w-16 bg-muted rounded-full animate-pulse" />
          </div>
          <div className="space-y-2">
            <div className="h-10 bg-muted rounded-lg mx-auto w-80 animate-pulse" />
            <div className="h-6 bg-muted rounded-lg mx-auto w-96 animate-pulse" />
          </div>
        </div>

        {/* Main Content Area Skeleton */}
        <div className="max-w-4xl mx-auto">
          <div className="grid gap-6 lg:grid-cols-3">
            {/* Left Column - Status & Actions Skeleton */}
            <div className="lg:col-span-2 space-y-6">
              {/* Status Card Skeleton */}
              <Card className="shadow-lg border overflow-hidden">
                <CardHeader className="bg-gradient-to-br from-card to-card/80">
                  <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-muted">
                        <Shield className="h-5 w-5 text-muted-foreground" />
                      </div>
                      <span className="text-xl text-muted-foreground">License Status</span>
                    </CardTitle>
                    <div className="h-6 w-20 bg-muted rounded-full animate-pulse" />
                  </div>
                  <div className="mt-2 flex items-center gap-2">
                    <Loader2 className="h-4 w-4 animate-spin text-blue-500 dark:text-blue-400" />
                    <span className="text-sm text-muted-foreground">Initializing license manager...</span>
                  </div>
                </CardHeader>
                
                <CardContent className="pt-6">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <div className="h-4 w-12 bg-muted rounded animate-pulse" />
                      <div className="h-6 w-20 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="space-y-2">
                      <div className="h-4 w-16 bg-muted rounded animate-pulse" />
                      <div className="h-6 w-24 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="space-y-2">
                      <div className="h-4 w-20 bg-muted rounded animate-pulse" />
                      <div className="h-6 w-16 bg-muted rounded animate-pulse" />
                    </div>
                    <div className="space-y-2">
                      <div className="h-4 w-12 bg-muted rounded animate-pulse" />
                      <div className="h-6 w-20 bg-muted rounded animate-pulse" />
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Activation Form Skeleton */}
              <Card className="shadow-lg border">
                <CardHeader>
                  <div className="h-6 w-40 bg-muted rounded animate-pulse" />
                  <div className="h-4 w-64 bg-muted rounded animate-pulse" />
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    <div className="h-4 w-20 bg-muted rounded animate-pulse" />
                    <div className="h-10 w-full bg-muted rounded animate-pulse" />
                  </div>
                  <div className="h-10 w-full bg-blue-200 dark:bg-blue-900/30 rounded animate-pulse" />
                </CardContent>
              </Card>
            </div>

            {/* Right Column - Features Skeleton */}
            <div className="space-y-6">
              {/* Features Card Skeleton */}
              <Card className="shadow-lg border">
                <CardHeader className="pb-4">
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Award className="h-5 w-5 text-muted-foreground" />
                    <span className="text-muted-foreground">Professional Features</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="flex items-start gap-3">
                      <div className="h-8 w-8 rounded-lg bg-muted animate-pulse flex-shrink-0 mt-0.5" />
                      <div className="flex-1 space-y-1">
                        <div className="h-4 w-24 bg-muted rounded animate-pulse" />
                        <div className="h-3 w-32 bg-muted rounded animate-pulse" />
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>

              {/* Support Card Skeleton */}
              <Card className="shadow-lg border bg-gradient-to-br from-card to-card/80">
                <CardContent className="pt-6">
                  <div className="text-center space-y-3">
                    <div className="h-12 w-12 rounded-full bg-muted animate-pulse mx-auto" />
                    <div className="space-y-2">
                      <div className="h-4 w-20 bg-muted rounded animate-pulse mx-auto" />
                      <div className="h-3 w-32 bg-muted rounded animate-pulse mx-auto" />
                    </div>
                    <div className="h-8 w-full bg-muted rounded animate-pulse" />
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

/**
 * Server Component wrapper for the License page
 * Exports metadata and renders the dynamic client component
 */
export default function LicensePage(): JSX.Element {
  return <LicenseContent />
}