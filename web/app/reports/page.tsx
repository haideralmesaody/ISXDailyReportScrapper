/**
 * Reports Page - Server Component
 * Provides SEO metadata and serves as entry point for reports feature
 * Following CLAUDE.md Next.js 14 server component patterns
 */

import ReportsClient from './reports-client'

// SEO metadata exported from server component
export const metadata = {
  title: 'Reports - ISX Pulse',
  description: 'View, analyze and download Iraqi Stock Exchange reports including daily trading, ticker history, and market indices.',
  keywords: 'ISX reports, Iraqi Stock Exchange data, trading reports, market analysis, financial data',
  authors: [{ name: 'ISX Pulse Team' }],
  robots: {
    index: true,
    follow: true
  },
  openGraph: {
    title: 'Reports - ISX Pulse',
    description: 'Access comprehensive Iraqi Stock Exchange reports and market data.',
    type: 'website',
  }
}

export default function ReportsPage() {
  return <ReportsClient />
}