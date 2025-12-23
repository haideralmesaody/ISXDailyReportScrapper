import dynamic from 'next/dynamic'

// Dynamically import the dashboard to ensure it only renders on the client
const LiquidityDashboard = dynamic(() => import('./liquidity-dashboard'), {
  ssr: false,
  loading: () => (
    <div className="min-h-screen p-8 flex items-center justify-center">
      <div className="text-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto mb-4" />
        <p className="text-muted-foreground">Loading dashboard...</p>
      </div>
    </div>
  )
})

// Now that this is a server component, we can export metadata
export const metadata = {
  title: 'Liquidity Analysis - ISX Pulse',
  description: 'ISX Hybrid Liquidity Metrics - Advanced liquidity scoring for Iraqi Stock Exchange securities.',
  robots: {
    index: false,
    follow: false
  }
}

export default function LiquidityPage(): JSX.Element {
  // Return the liquidity dashboard directly
  return <LiquidityDashboard />
}