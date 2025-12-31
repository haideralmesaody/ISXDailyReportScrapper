import dynamic from 'next/dynamic'

const AnalysisClient = dynamic(() => import('./analysis-client'), { ssr: false })

export const metadata = {
  title: 'Analysis - ISX Pulse',
  description: 'Candlestick charts for ISX tickers (real data).',
}

export default function AnalysisPage() {
  return <AnalysisClient />
}
