import dynamic from 'next/dynamic'
import { Skeleton } from '@/components/ui/skeleton'

const AppContentClient = dynamic(
  () => import('./app-content-client'),
  {
    ssr: false,
    loading: () => (
      <div className="min-h-screen bg-background flex flex-col">
        {/* Header Skeleton */}
        <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur">
          <div className="px-6">
            <div className="flex h-16 items-center justify-between">
              {/* Logo Skeleton */}
              <div className="flex items-center space-x-3">
                <Skeleton className="h-8 w-8 rounded" />
                <div className="hidden sm:block">
                  <Skeleton className="h-6 w-32" />
                  <Skeleton className="h-3 w-24 mt-1" />
                </div>
              </div>
              
              {/* Navigation Skeleton - Desktop */}
              <nav className="hidden md:flex items-center space-x-6">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="flex items-center space-x-2">
                    <Skeleton className="h-4 w-4" />
                    <Skeleton className="h-4 w-16" />
                  </div>
                ))}
              </nav>
              
              {/* Mobile Menu Button Skeleton */}
              <div className="md:hidden">
                <Skeleton className="h-9 w-9 rounded" />
              </div>
              
              {/* Status Indicator Skeleton */}
              <div className="hidden md:flex items-center space-x-2">
                <Skeleton className="h-4 w-4" />
                <Skeleton className="h-3 w-12 hidden sm:block" />
              </div>
            </div>
          </div>
        </header>
        
        {/* Main Content Skeleton */}
        <main className="flex-1 px-6 py-8">
          <div className="space-y-6">
            <div className="space-y-2">
              <Skeleton className="h-8 w-64" />
              <Skeleton className="h-4 w-96" />
            </div>
            
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="space-y-3 p-6 border rounded-lg">
                  <div className="flex items-center justify-between">
                    <Skeleton className="h-5 w-24" />
                    <Skeleton className="h-4 w-4" />
                  </div>
                  <Skeleton className="h-4 w-full" />
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-9 w-20" />
                </div>
              ))}
            </div>
            
            {/* Additional content skeleton */}
            <div className="space-y-4">
              <Skeleton className="h-6 w-48" />
              <div className="space-y-2">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-5/6" />
                <Skeleton className="h-4 w-4/6" />
              </div>
            </div>
          </div>
        </main>
        
        {/* Footer Skeleton */}
        <footer className="border-t bg-card mt-auto">
          <div className="px-6">
            <div className="flex h-12 items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center space-x-4">
                <Skeleton className="h-3 w-24" />
                <Skeleton className="h-5 w-20 rounded-full" />
                <div className="flex items-center space-x-1">
                  <Skeleton className="h-3 w-12" />
                  <Skeleton className="h-3 w-16" />
                </div>
              </div>
              <div className="hidden sm:flex items-center space-x-4">
                <Skeleton className="h-3 w-32" />
                <span>•</span>
                <Skeleton className="h-3 w-24" />
                <span>•</span>
                <Skeleton className="h-3 w-28" />
              </div>
            </div>
          </div>
        </footer>
      </div>
    )
  }
)

export function AppContent({ children }: { children: React.ReactNode }) {
  return <AppContentClient>{children}</AppContentClient>
}