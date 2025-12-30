'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  FileText,
  BarChart3,
  TrendingUp,
  Table,
  Download,
  CheckCircle2,
  AlertCircle,
  ChevronRight,
  FileSpreadsheet,
  Database,
  Eye
} from 'lucide-react'
import { InteractiveDemo } from '../components/InteractiveDemo'
import { ReportExplorer } from '@/components/guide/demos/ReportExplorer'

interface ReportsSectionProps {
  onNavigate: (sectionId: string) => void
}

/**
 * Reports Section - Guide to all report types and data access
 * Covers report structure, formats, and integration options
 */
export function ReportsSection({ onNavigate }: ReportsSectionProps) {
  const reportTypes = [
    {
      id: 'combined',
      name: 'Combined Report',
      icon: FileText,
      color: 'bg-blue-500',
      description: 'Comprehensive daily trading data for all ISX stocks - your main data source',
      location: 'data/reports/combined/isx_combined_data.csv',
      format: 'CSV (Single File)',
      fields: [
        'Date - Trading date in YYYY-MM-DD format',
        'CompanyName - Full company name',
        'Symbol - Stock ticker code (e.g., BAQL, AIPM)',
        'OpenPrice - Opening price in IQD',
        'HighPrice - Highest traded price of the day',
        'LowPrice - Lowest traded price of the day',
        'AveragePrice - Volume-weighted average price',
        'PrevAveragePrice - Previous day\'s average (for comparison)',
        'ClosePrice - Final closing price in IQD',
        'PrevClosePrice - Previous day\'s close (for change calculation)',
        'Change - Price change from previous close (IQD)',
        'ChangePercent - Price change percentage (positive = gain, negative = loss)',
        'NumTrades - Total number of transactions',
        'Volume - Total shares traded',
        'Value - Total trading value in IQD',
        'TradingStatus - "true" = active trading, "false" = suspended/no activity'
      ],
      useCases: [
        'Daily price tracking and technical analysis',
        'Identifying gainers and losers',
        'Volume analysis and momentum detection',
        'Building custom trading strategies'
      ]
    },
    {
      id: 'indexes',
      name: 'Indexes Report',
      icon: BarChart3,
      color: 'bg-purple-500',
      description: 'ISX60 and ISX15 market index values - track overall market health',
      location: 'data/reports/indexes/indexes.csv',
      format: 'CSV (Single File)',
      fields: [
        'Date - Trading date in YYYY-MM-DD format',
        'ISX60 - Top 60 stocks index value (higher = bullish market)',
        'ISX15 - Top 15 stocks index value (tracks large-cap leaders)'
      ],
      useCases: [
        'Market sentiment analysis',
        'Comparing individual stock performance to market',
        'Identifying bull/bear market trends',
        'Portfolio benchmarking'
      ]
    },
    {
      id: 'liquidity',
      name: 'Liquidity Reports',
      icon: TrendingUp,
      color: 'bg-green-500',
      description: 'Advanced liquidity metrics and safe trade recommendations - find tradable stocks',
      location: 'data/reports/liquidity_reports/liquidity_scores_YYYY-MM-DD.csv',
      format: 'CSV (Daily Files)',
      fields: [
        'Date - Analysis date',
        'Symbol - Stock ticker code',
        'Window - Analysis window (e.g., 60d = 60-day analysis)',
        'ILLIQ_Raw - Raw Amihud illiquidity ratio',
        'ILLIQ_Scaled - Scaled illiquidity score (0-100)',
        'Value_Raw - Raw trading value metric',
        'Value_Scaled - Scaled value score (0-100)',
        'Continuity_Raw - Raw trading continuity ratio',
        'Continuity_Scaled - Scaled continuity score (0-100)',
        'Activity_Score - Overall liquidity quality (0-100, higher = more liquid)',
        'Spread_Proxy - Estimated bid-ask spread indicator',
        'Spread_Scaled - Scaled spread score (0-100)',
        'Hybrid_Score - Composite liquidity score combining all metrics',
        'Hybrid_Rank - Ranking among all stocks (1 = most liquid)',
        'Trading_Days - Number of days with trading activity',
        'Data_Quality - HIGH/MEDIUM/LOW data reliability indicator',
        'Safe_Trade_0.5% - Maximum shares to trade with 0.5% price impact',
        'Safe_Trade_1% - Maximum shares to trade with 1% price impact',
        'Safe_Trade_2% - Maximum shares to trade with 2% price impact',
        'Optimal_Trade - Recommended trade size for minimal impact'
      ],
      useCases: [
        'Finding liquid stocks for large orders',
        'Estimating price impact before trading',
        'Risk management and position sizing',
        'Identifying market inefficiencies'
      ]
    }
  ]

  const dataIntegrity = [
    {
      check: 'No missing required fields',
      description: 'All 10 core fields must be present for each row'
    },
    {
      check: 'Valid date format',
      description: 'Dates in YYYY-MM-DD format, trading days only (Sun-Thu)'
    },
    {
      check: 'Positive numeric values',
      description: 'Prices and volumes must be ≥ 0'
    },
    {
      check: 'Ticker symbol consistency',
      description: 'Codes match official ISX ticker list'
    },
    {
      check: 'No duplicate rows',
      description: 'Unique combination of Date + Code'
    }
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-primary/10">
            <FileText className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold">Reports System</h2>
            <p className="text-muted-foreground text-lg">
              Understanding report types, formats, and how to access your data
            </p>
          </div>
        </div>
      </div>

      {/* Report Types Overview */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Report Types</h3>
        <p className="text-muted-foreground">
          ISX Pulse generates 3 types of reports from the processing pipeline. Each serves a different purpose and use case.
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {reportTypes.map((report) => {
            const ReportIcon = report.icon
            return (
              <Card key={report.id} className="relative overflow-hidden">
                <div className={`absolute top-0 left-0 w-full h-1 ${report.color}`} />
                <CardHeader>
                  <div className="flex items-start gap-3">
                    <div className={`p-2 rounded-lg ${report.color} bg-opacity-10 shrink-0`}>
                      <ReportIcon className="h-5 w-5" />
                    </div>
                    <div className="flex-1">
                      <CardTitle className="text-base">{report.name}</CardTitle>
                      <CardDescription className="mt-1">{report.description}</CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div>
                    <p className="text-xs font-semibold text-muted-foreground mb-1">LOCATION</p>
                    <code className="text-xs bg-muted px-2 py-1 rounded">{report.location}</code>
                  </div>
                  <div>
                    <Badge variant="outline">{report.format}</Badge>
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </section>

      {/* Interactive Report Explorer */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Explore Sample Report</h3>
        <p className="text-muted-foreground">
          Try the interactive report explorer to understand the data structure:
        </p>

        <InteractiveDemo
          title="Report Explorer"
          description="Browse, search, and filter sample ISX report data"
          variant="interactive"
        >
          <ReportExplorer />
        </InteractiveDemo>
      </section>

      {/* Detailed Report Breakdown */}
      <section className="space-y-6">
        <h3 className="text-2xl font-semibold">Report Field Reference</h3>

        {reportTypes.map((report) => {
          const ReportIcon = report.icon
          return (
            <Card key={report.id} className="relative overflow-hidden">
              <div className={`absolute top-0 left-0 w-1 h-full ${report.color}`} />
              <CardHeader>
                <div className="flex items-start gap-4">
                  <div className={`p-3 rounded-lg ${report.color} bg-opacity-10 shrink-0`}>
                    <ReportIcon className="h-6 w-6" />
                  </div>
                  <div className="flex-1">
                    <CardTitle className="text-xl">{report.name}</CardTitle>
                    <CardDescription className="mt-1">{report.description}</CardDescription>
                  </div>
                </div>
              </CardHeader>

              <CardContent className="space-y-4">
                {/* Fields */}
                <div>
                  <h4 className="font-semibold mb-3 flex items-center gap-2">
                    <Table className="h-4 w-4 text-primary" />
                    Data Fields
                  </h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                    {report.fields.map((field, i) => (
                      <div key={i} className="flex items-start gap-2 text-sm">
                        <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                        <span>{field}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Use Cases */}
                <div>
                  <h4 className="font-semibold mb-3 flex items-center gap-2">
                    <Eye className="h-4 w-4 text-primary" />
                    Use Cases
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {report.useCases.map((useCase, i) => (
                      <Badge key={i} variant="secondary">{useCase}</Badge>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </section>

      {/* Data Integrity */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Data Integrity Checks</h3>
        <p className="text-muted-foreground">
          All reports go through validation to ensure data quality:
        </p>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {dataIntegrity.map((check, i) => (
            <Card key={i}>
              <CardContent className="pt-6">
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                  <div>
                    <h4 className="font-semibold mb-1">{check.check}</h4>
                    <p className="text-sm text-muted-foreground">{check.description}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Download & Export */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Download & Export Options</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="border-blue-500/50 bg-blue-500/5">
            <CardContent className="pt-6">
              <div className="flex items-start gap-3">
                <Download className="h-5 w-5 text-blue-500 shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Direct CSV Download</h4>
                  <p className="text-sm text-muted-foreground">
                    Files available in data/reports/ directory
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-purple-500/50 bg-purple-500/5">
            <CardContent className="pt-6">
              <div className="flex items-start gap-3">
                <FileSpreadsheet className="h-5 w-5 text-purple-500 shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Excel Export</h4>
                  <p className="text-sm text-muted-foreground">
                    Convert CSV to Excel with formatting
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-green-500/50 bg-green-500/5">
            <CardContent className="pt-6">
              <div className="flex items-start gap-3">
                <Database className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">API Access</h4>
                  <p className="text-sm text-muted-foreground">
                    Query via RESTful API endpoints
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* Common Issues */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Troubleshooting</h3>
        <div className="space-y-3">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              <p className="font-semibold">Report file not found</p>
              <p className="text-sm mt-1"><strong>Cause:</strong> Operation not yet run for this date</p>
              <p className="text-sm text-green-600 dark:text-green-400">
                <strong>Solution:</strong> Run the Process Data or Full Pipeline operation first
              </p>
            </AlertDescription>
          </Alert>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              <p className="font-semibold">Missing data for some stocks</p>
              <p className="text-sm mt-1"><strong>Cause:</strong> Stock not traded on that date</p>
              <p className="text-sm text-green-600 dark:text-green-400">
                <strong>Solution:</strong> This is normal. Zero volume = no trading activity
              </p>
            </AlertDescription>
          </Alert>
        </div>
      </section>

      {/* Next Steps */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Next Steps</h3>
        <Card className="border-primary/50 bg-primary/5">
          <CardContent className="pt-6">
            <div className="space-y-4">
              <p>
                Now that you understand the reports system, learn how to visualize this data with advanced charting features:
              </p>
              <button
                onClick={() => onNavigate('charts-basics')}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Continue to Charts Basics
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
