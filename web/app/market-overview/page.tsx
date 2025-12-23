import { Metadata } from 'next'
import MarketOverviewClient from './market-overview-client'

export const metadata: Metadata = {
  title: 'Market Overview - ISX Pulse',
  description: 'Iraqi Stock Exchange (ISX) market overview with interactive treemap visualization.',
  robots: { index: false, follow: false }
}

export default function MarketOverviewPage() {
  return <MarketOverviewClient />
}
