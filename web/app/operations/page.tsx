import dynamic from 'next/dynamic'
import {
  Loader2,
  Download,
  Zap,
  Database,
  FileSpreadsheet,
  BarChart3,
  Workflow,
  Info,
  TrendingUp
} from 'lucide-react'

// Dynamic import with SSR disabled to prevent hydration issues
const OperationsContent = dynamic(() => import('./operations-content'), {
  ssr: false,
  loading: () => <OperationsPageSkeleton />
})

// SEO metadata
export const metadata = {
  title: 'Operations - ISX Pulse',
  description: 'Manage and monitor ISX data processing operations with real-time WebSocket updates.',
  robots: { index: false, follow: false }
}

// Simple loading skeleton
function OperationsPageSkeleton() {
  return (
    <div className="min-h-screen p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header Skeleton */}
        <div>
          <div className="h-9 w-48 bg-muted rounded-md animate-pulse mb-2" />
          <div className="h-5 w-96 bg-muted rounded-md animate-pulse" />
        </div>

        {/* Content Skeleton */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="border rounded-lg p-6 space-y-4">
              <div className="h-6 w-3/4 bg-muted rounded-md animate-pulse" />
              <div className="space-y-2">
                <div className="h-4 w-full bg-muted rounded-md animate-pulse" />
                <div className="h-4 w-2/3 bg-muted rounded-md animate-pulse" />
              </div>
              <div className="h-8 w-full bg-muted rounded-md animate-pulse" />
            </div>
          ))}
        </div>

        {/* Loading Indicator */}
        <div className="fixed bottom-8 right-8">
          <div className="bg-background border rounded-lg p-4 shadow-lg">
            <div className="flex items-center gap-3">
              <Loader2 className="h-5 w-5 animate-spin text-primary" />
              <span className="text-sm text-muted-foreground">
                Initializing operations...
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function OperationsPage() {
  return <OperationsContent />
}