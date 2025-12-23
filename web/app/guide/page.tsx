import { Metadata } from 'next'
import GuideClient from './guide-client'

/**
 * Server component for /guide route
 * Provides SEO metadata while delegating interactivity to client component
 * Follows Next.js 14 App Router and CLAUDE.md server/client split pattern
 */
export const metadata: Metadata = {
  title: 'Interactive Guide - ISX Pulse',
  description: 'Comprehensive guide and tutorials for ISX Pulse features including technical analysis, trading strategies, pipeline operations, and market data visualization.',
  keywords: [
    'ISX guide',
    'stock trading tutorial',
    'technical indicators',
    'chart analysis',
    'data pipeline',
    'trading strategies',
    'MACD tutorial',
    'RSI guide',
    'ISX market analysis'
  ],
  openGraph: {
    title: 'Interactive Guide - ISX Pulse',
    description: 'Master ISX stock analysis with hands-on tutorials covering technical indicators, automated strategies, and data operations.',
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
