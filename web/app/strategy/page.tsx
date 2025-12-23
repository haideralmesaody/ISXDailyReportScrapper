import { Metadata } from 'next'
import StrategyClient from './strategy-client'

export const metadata: Metadata = {
  title: 'Trading Strategies - ISX Pulse',
  description: 'Batch trading strategies and signal generation for the Iraqi Stock Exchange',
  robots: { index: false, follow: false }
}

export default function StrategyPage() {
  return <StrategyClient />
}
