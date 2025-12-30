import { Metadata } from 'next'
import GuideClient from './guide-client'

/**
 * Server component for /guide route
 * Provides SEO metadata while delegating interactivity to client component
 * Follows Next.js 14 App Router and CLAUDE.md server/client split pattern
 */
export const metadata: Metadata = {
  title: 'Interactive Guide - ISX Pulse',
  description: 'Comprehensive guide and tutorials for ISX Pulse features including trading strategies, pipeline operations, liquidity insights, and market data exploration.',
  keywords: [
    'ISX guide',
    'stock trading tutorial',
    'data pipeline',
    'liquidity',
    'trading strategies',
    'ISX market analysis'
  ],
  openGraph: {
    title: 'Interactive Guide - ISX Pulse',
    description: 'Learn the ISX data pipeline and build automated strategies using real market data.',
    type: 'website'
  },
  robots: {
    index: true,
    follow: true
  }
}

export default function GuidePage() {
  return <GuideClient />
}
