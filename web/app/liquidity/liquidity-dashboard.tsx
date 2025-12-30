'use client'

import { useState, useEffect, useCallback } from 'react'
import { useTheme } from 'next-themes'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { 
  ArrowLeft, 
  TrendingUp, 
  AlertCircle, 
  Info,
  Target,
  Shield,
  Zap,
  Activity,
  BarChart3,
  Droplets,
  RefreshCw,
  Search,
  Clock,
  Calculator
} from 'lucide-react'
import Link from 'next/link'
import { Input } from '@/components/ui/input'
import { useHydration } from '@/lib/hooks/use-hydration'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { ColoredProgress } from '@/components/ui/colored-progress'
import { HelpButton } from '@/components/guide/HelpButton'

interface TradingThreshold {
  conservative: number
  moderate: number
  aggressive: number
  optimal: number
}

interface StockMetrics {
  score: number
  thresholds: TradingThreshold
  continuity: number
  dailyVolume: number
}

export interface StockRecommendation {
  symbol: string
  companyName?: string  // Company name from ISX data
  score: number        // Average score (for backward compatibility)
  ema20Score: number   // 20-period EMA
  latestScore: number  // Most recent score
  thresholds: TradingThreshold
  action: string
  rationale: string
  continuity: number
  dailyVolume: number
  dataQuality: string
  // Component scores for transparency
  illiqScore?: number      // Price impact score (0-100)
  volumeScore?: number     // Trading volume score (0-100)
  continuityScore?: number // Continuity score (0-100)
  illiqRaw?: number        // Raw ILLIQ value
  volumeRaw?: number       // Raw volume in IQD
  // New fields for multi-mode support
  emaMetrics?: StockMetrics
  latestMetrics?: StockMetrics
  averageMetrics?: StockMetrics
  activeMode?: string
}

// Scoring mode type
export type ScoringMode = 'ema' | 'latest' | 'average'

// Legacy format (backward compatibility)
interface LiquidityInsightsLegacy {
  generatedAt: string
  marketHealthScore: number
  totalStocks: number
  highQualityStocks: number
  averageContinuity: number
  medianDailyVolume: number
  topOpportunities: StockRecommendation[]
  bestForLargeTrades: StockRecommendation[]
  bestForDayTrading: StockRecommendation[]
  highRisk: StockRecommendation[]
}

// SSOT format (new)
export interface LiquidityInsightsSSOT {
  generatedAt: string
  marketHealthScore: number
  totalStocks: number
  highQualityStocks: number
  averageContinuity: number
  medianDailyVolume: number
  allStocks: StockRecommendation[]
  topOpportunities: string[]
  bestForLargeTrades: string[]
  bestForDayTrading: string[]
  highRisk: string[]
}

// Combined type for handling both formats
type LiquidityInsights = LiquidityInsightsLegacy | LiquidityInsightsSSOT

// Type guard to check if insights use SSOT format
function isSSOTFormat(insights: LiquidityInsights): insights is LiquidityInsightsSSOT {
  return 'allStocks' in insights && Array.isArray(insights.allStocks)
}

// Helper to get stocks for a category
function getCategoryStocks(insights: LiquidityInsights, category: 'topOpportunities' | 'bestForLargeTrades' | 'bestForDayTrading' | 'highRisk'): StockRecommendation[] {
  if (isSSOTFormat(insights)) {
    // SSOT format: look up stocks from allStocks using ticker symbols
    const symbols = insights[category] as string[]
    return symbols.map(symbol => 
      insights.allStocks.find(stock => stock.symbol === symbol)
    ).filter((stock): stock is StockRecommendation => stock !== undefined)
  } else {
    // Legacy format: direct array of stocks
    return insights[category] as StockRecommendation[]
  }
}

// Format currency with proper IQD formatting
function formatCurrency(value: number): string {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(1)}B IQD`
  }
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(1)}M IQD`
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(0)}K IQD`
  }
  return `${value.toFixed(0)} IQD`
}

// Get color for score
function getScoreColor(score: number): string {
  if (score >= 80) return 'text-green-600'
  if (score >= 60) return 'text-blue-600'
  if (score >= 40) return 'text-yellow-600'
  if (score >= 20) return 'text-orange-600'
  return 'text-red-600'
}

// Get border color for score (returns hex color)
function getScoreBorderColor(score: number): string {
  if (score >= 80) return '#22c55e' // green-500
  if (score >= 60) return '#3b82f6' // blue-500
  if (score >= 40) return '#f59e0b' // amber-500
  if (score >= 20) return '#fb923c' // orange-400
  return '#dc2626' // red-600
}

// Get background gradient for score (light mode only)
function getScoreBackgroundGradient(score: number): string {
  // Only apply subtle gradients in light mode
  if (score >= 80) return 'linear-gradient(135deg, #ffffff 0%, rgba(34, 197, 94, 0.03) 100%)'
  if (score >= 60) return 'linear-gradient(135deg, #ffffff 0%, rgba(59, 130, 246, 0.03) 100%)'
  if (score >= 40) return 'linear-gradient(135deg, #ffffff 0%, rgba(245, 158, 11, 0.03) 100%)'
  if (score >= 20) return 'linear-gradient(135deg, #ffffff 0%, rgba(251, 146, 60, 0.03) 100%)'
  return 'linear-gradient(135deg, #ffffff 0%, rgba(220, 38, 38, 0.03) 100%)'
}

// Get badge variant for action
function getActionVariant(action: string): "default" | "secondary" | "destructive" | "outline" {
  switch (action) {
    case 'BUY_LARGE':
    case 'BUY':
      return 'default'
    case 'DAY_TRADE':
      return 'secondary'
    case 'HOLD':
      return 'outline'
    case 'CAUTION':
    case 'AVOID':
      return 'destructive'
    default:
      return 'outline'
  }
}

// Trading card component with mode support (exported for reuse in guide demos)
export function TradingCard({ stock, mode = 'ema' }: { stock: StockRecommendation; mode?: ScoringMode }) {
  const { theme } = useTheme()
  const isDark = theme === 'dark'
  
  // Select metrics based on active mode - ALL components change with mode
  const getActiveMetrics = (): StockMetrics => {
    switch (mode) {
      case 'ema':
        if (stock.emaMetrics) {
          return stock.emaMetrics
        }
        // Fallback for backward compatibility
        return {
          score: stock.ema20Score || stock.score,
          thresholds: stock.thresholds,
          continuity: stock.continuity,
          dailyVolume: stock.dailyVolume
        }
      case 'latest':
        if (stock.latestMetrics) {
          return stock.latestMetrics
        }
        // Fallback for backward compatibility
        return {
          score: stock.latestScore || stock.score,
          thresholds: stock.thresholds,
          continuity: stock.continuity,
          dailyVolume: stock.dailyVolume
        }
      case 'average':
        if (stock.averageMetrics) {
          return stock.averageMetrics
        }
        // Fallback for backward compatibility
        return {
          score: stock.score,
          thresholds: stock.thresholds,
          continuity: stock.continuity,
          dailyVolume: stock.dailyVolume
        }
      default:
        return {
          score: stock.score,
          thresholds: stock.thresholds,
          continuity: stock.continuity,
          dailyVolume: stock.dailyVolume
        }
    }
  }
  
  const metrics = getActiveMetrics()
  const primaryScore = metrics.score
  
  return (
    <Card 
      className="hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5 overflow-hidden relative dark:bg-card"
      style={{
        borderLeft: `4px solid ${getScoreBorderColor(primaryScore)}`,
        background: isDark ? undefined : getScoreBackgroundGradient(primaryScore),
        boxShadow: isDark ? undefined : '0 1px 3px rgba(0,0,0,0.1)'
      }}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div>
            <CardTitle className="text-lg font-mono tracking-wide">{stock.symbol}</CardTitle>
            {stock.companyName && (
              <p className="text-sm text-muted-foreground mt-1">{stock.companyName}</p>
            )}
            <div className="space-y-1 mt-2">
              {/* Primary score display (EMA20) */}
              <div className="flex items-center gap-3">
                <div className="relative">
                  <span 
                    className={`text-3xl font-bold ${getScoreColor(primaryScore)}`}
                    style={{
                      textShadow: '0 1px 2px rgba(0,0,0,0.1)'
                    }}
                  >
                    {primaryScore.toFixed(1)}
                  </span>
                </div>
                <Badge 
                  variant={getActionVariant(stock.action)} 
                  className="text-xs px-2 py-0.5"
                  style={{
                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                  }}
                >
                  {stock.action.replace('_', ' ')}
                </Badge>
              </div>
              {/* Mode indicator */}
              <div className="flex gap-3 text-xs text-muted-foreground">
                <span className="font-semibold">Mode: {mode.toUpperCase()}</span>
                {mode !== 'ema' && stock.ema20Score && (
                  <span>EMA: {stock.ema20Score.toFixed(1)}</span>
                )}
                {mode !== 'latest' && stock.latestScore && (
                  <span>Latest: {stock.latestScore.toFixed(1)}</span>
                )}
                {mode !== 'average' && (
                  <span>Avg: {stock.score.toFixed(1)}</span>
                )}
              </div>
            </div>
          </div>
          {stock.dataQuality === 'GOOD' && (
            <div className="flex items-center gap-1.5 px-2 py-1 border rounded-md bg-green-50/50 border-green-200 dark:bg-green-900/20 dark:border-green-800">
              <div 
                className="w-2 h-2 rounded-full bg-green-500 dark:bg-green-400"
                style={{
                  boxShadow: isDark ? '0 0 4px rgba(74, 222, 128, 0.5)' : '0 0 4px rgba(34, 197, 94, 0.5)'
                }}
              />
              <span className="text-xs font-medium text-green-700 dark:text-green-400">Good Data</span>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="p-3 rounded-lg bg-gray-50/50 border border-gray-100 dark:bg-gray-900/30 dark:border-gray-800">
          <p className="text-sm text-muted-foreground italic leading-relaxed">"{stock.rationale}"</p>
        </div>
        
        {/* Score Breakdown - Show component scores if available */}
        {(stock.illiqScore !== undefined || stock.volumeScore !== undefined || stock.continuityScore !== undefined) && (
          <div className="space-y-2 p-3.5 rounded-lg border border-gray-100 dark:border-gray-800" 
               style={{ 
                 background: isDark ? undefined : 'linear-gradient(135deg, rgba(249, 250, 251, 0.5) 0%, rgba(243, 244, 246, 0.3) 100%)' 
               }}>
            <h4 className="text-xs font-semibold uppercase text-muted-foreground flex items-center gap-1.5 mb-3">
              <BarChart3 className="h-3.5 w-3.5" />
              Score Components (3-Metric System)
            </h4>
            
            {/* ILLIQ Score */}
            {stock.illiqScore !== undefined && (
              <div className="space-y-1">
                <div className="flex justify-between text-xs">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="text-muted-foreground cursor-help underline-offset-2 hover:underline">
                          Price Impact (ILLIQ)
                        </span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs">
                        <p className="font-semibold">Amihud Illiquidity Measure</p>
                        <p className="text-xs mt-1">
                          Measures price impact of trades. Lower values (higher scores) mean less price movement per unit of trading volume - better liquidity.
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  <span className="font-medium">{stock.illiqScore.toFixed(1)} × 40%</span>
                </div>
                <ColoredProgress value={stock.illiqScore} className="h-1.5" />
                {stock.illiqRaw !== undefined && (
                  <span className="text-xs text-muted-foreground">Raw: {stock.illiqRaw.toFixed(4)}</span>
                )}
              </div>
            )}
            
            {/* Volume Score */}
            {stock.volumeScore !== undefined && (
              <div className="space-y-1">
                <div className="flex justify-between text-xs">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="text-muted-foreground cursor-help underline-offset-2 hover:underline">
                          Trading Volume
                        </span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs">
                        <p className="font-semibold">Daily Trading Value</p>
                        <p className="text-xs mt-1">
                          Total IQD value traded daily. Higher volumes indicate more liquidity and easier entry/exit.
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  <span className="font-medium">{stock.volumeScore.toFixed(1)} × 35%</span>
                </div>
                <ColoredProgress value={stock.volumeScore} className="h-1.5" />
                {stock.volumeRaw !== undefined && (
                  <span className="text-xs text-muted-foreground">Raw: {formatCurrency(stock.volumeRaw)}</span>
                )}
              </div>
            )}
            
            {/* Continuity Score */}
            {stock.continuityScore !== undefined && (
              <div className="space-y-1">
                <div className="flex justify-between text-xs">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="text-muted-foreground cursor-help underline-offset-2 hover:underline">
                          Trading Continuity
                        </span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs">
                        <p className="font-semibold">Trading Frequency</p>
                        <p className="text-xs mt-1">
                          Percentage of days with active trading. Higher continuity means more consistent liquidity availability.
                        </p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                  <span className="font-medium">{stock.continuityScore.toFixed(1)} × 25%</span>
                </div>
                <ColoredProgress value={stock.continuityScore} className="h-1.5" />
                <span className="text-xs text-muted-foreground">Raw: {(stock.continuity * 100).toFixed(1)}%</span>
              </div>
            )}
            
            {/* Final Score Calculation */}
            <div className="pt-2 mt-2 border-t text-xs">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Weighted Score:</span>
                <span className="font-semibold text-primary">
                  {((stock.illiqScore || 0) * 0.40 + 
                    (stock.volumeScore || 0) * 0.35 + 
                    (stock.continuityScore || 0) * 0.25).toFixed(1)}
                </span>
              </div>
            </div>
          </div>
        )}
        
        {/* Trading Thresholds - Mode-specific values */}
        <div className="space-y-2 p-3 rounded-lg bg-gradient-to-br from-gray-50/50 to-gray-100/30 border border-gray-100 dark:from-gray-900/30 dark:to-gray-800/20 dark:border-gray-800">
          <h4 className="text-xs font-semibold uppercase text-muted-foreground flex items-center gap-1.5">
            <TrendingUp className="h-3.5 w-3.5" />
            Safe Trading Sizes ({mode.toUpperCase()} Mode)
          </h4>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div className="flex items-center gap-1">
              <Shield className="h-3 w-3 text-green-600" />
              <span className="text-xs text-muted-foreground">Conservative:</span>
              <span className="font-medium">{formatCurrency(metrics.thresholds.conservative)}</span>
            </div>
            <div className="flex items-center gap-1">
              <Activity className="h-3 w-3 text-blue-600" />
              <span className="text-xs text-muted-foreground">Moderate:</span>
              <span className="font-medium">{formatCurrency(metrics.thresholds.moderate)}</span>
            </div>
            <div className="flex items-center gap-1">
              <Zap className="h-3 w-3 text-orange-600" />
              <span className="text-xs text-muted-foreground">Aggressive:</span>
              <span className="font-medium">{formatCurrency(metrics.thresholds.aggressive)}</span>
            </div>
            <div className="flex items-center gap-1">
              <Target className="h-3 w-3 text-purple-600" />
              <span className="text-xs text-muted-foreground">Optimal:</span>
              <span className="font-medium">{formatCurrency(metrics.thresholds.optimal)}</span>
            </div>
          </div>
        </div>
        
        {/* Metrics - Mode-specific values */}
        <div className="flex items-center justify-between text-xs text-muted-foreground pt-2 border-t">
          <span>Continuity: {(stock.continuity * 100).toFixed(1)}%</span>
          <span>Daily Vol ({mode.toUpperCase()}): {formatCurrency(metrics.dailyVolume)}</span>
        </div>
      </CardContent>
    </Card>
  )
}

export default function LiquidityDashboard() {
  const [insights, setInsights] = useState<LiquidityInsights | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [scoringMode, setScoringMode] = useState<ScoringMode>(() => {
    // Load saved preference from localStorage
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('liquidityScoringMode')
      if (saved === 'ema' || saved === 'latest' || saved === 'average') {
        return saved as ScoringMode
      }
    }
    return 'ema' // Default to EMA mode
  })
  const isHydrated = useHydration()
  
  // Save mode preference when it changes
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('liquidityScoringMode', scoringMode)
    }
  }, [scoringMode])

  // Load insights data with mode parameter
  const loadInsights = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      
      // Fetch the latest insights from API with mode parameter
      const response = await fetch(`/api/liquidity/insights?mode=${scoringMode}`)
      if (!response.ok) {
        throw new Error('Failed to load liquidity insights')
      }
      
      const data = await response.json()
      setInsights(data)
    } catch (err) {
      console.error('Error loading insights:', err)
      setError(err instanceof Error ? err.message : 'Failed to load insights')
      setInsights(null)
    } finally {
      setLoading(false)
    }
  }, [scoringMode])

  useEffect(() => {
    if (isHydrated) {
      loadInsights()
    }
  }, [isHydrated, loadInsights])

  // Filter stocks based on search (symbol or company name)
  const filterStocks = (stocks: StockRecommendation[]) => {
    if (!searchTerm) return stocks
    const searchLower = searchTerm.toLowerCase()
    return stocks.filter(s =>
      s.symbol.toLowerCase().includes(searchLower) ||
      (s.companyName && s.companyName.toLowerCase().includes(searchLower))
    )
  }

  if (!isHydrated) {
    return (
      <div className="min-h-screen p-8 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <Droplets className="h-12 w-12 text-blue-600 mx-auto mb-4 animate-pulse" />
            <p className="text-muted-foreground">Initializing dashboard...</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="min-h-screen p-8 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center">
            <RefreshCw className="h-12 w-12 text-blue-600 mx-auto mb-4 animate-spin" />
            <p className="text-muted-foreground">Loading liquidity insights...</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (error) {
    // Check if error is due to no data available
    const isNoDataError = error.includes('404') || error.includes('not found') || error.includes('Failed to load')
    
    if (isNoDataError) {
      return (
        <div className="min-h-screen p-8">
          <div className="max-w-7xl mx-auto">
            <Card className="p-8">
              <div className="text-center space-y-6">
                <div className="flex justify-center">
                  <div className="p-4 bg-blue-100 rounded-full">
                    <Droplets className="h-12 w-12 text-blue-600" />
                  </div>
                </div>
                
                <div className="space-y-2">
                  <h2 className="text-2xl font-semibold">No Liquidity Data Available</h2>
                  <p className="text-muted-foreground max-w-md mx-auto">
                    You need to run the data collection operations first to generate liquidity insights.
                  </p>
                </div>
                
                <Alert className="max-w-md mx-auto">
                  <Info className="h-4 w-4" />
                  <AlertTitle>How to get started:</AlertTitle>
                  <AlertDescription className="text-left mt-2">
                    <ol className="list-decimal list-inside space-y-1">
                      <li>Go to the Operations page</li>
                      <li>Run "Full Pipeline" to collect all data</li>
                      <li>Wait for the liquidity analysis to complete</li>
                      <li>Return here to view the insights</li>
                    </ol>
                  </AlertDescription>
                </Alert>
                
                <div className="flex gap-4 justify-center">
                  <Button asChild>
                    <Link href="/operations">
                      <Activity className="h-4 w-4 mr-2" />
                      Go to Operations
                    </Link>
                  </Button>
                  <Button onClick={loadInsights} variant="outline">
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Check Again
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </div>
      )
    }
    
    // Generic error display
    return (
      <div className="min-h-screen p-8">
        <div className="max-w-7xl mx-auto">
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
          <Button onClick={loadInsights} className="mt-4">
            <RefreshCw className="h-4 w-4 mr-2" />
            Retry
          </Button>
        </div>
      </div>
    )
  }

  if (!insights) {
    return null
  }

  const marketHealthColor = insights.marketHealthScore >= 70 ? 'text-green-600' :
                           insights.marketHealthScore >= 50 ? 'text-yellow-600' : 'text-red-600'

  return (
    <main className="min-h-screen p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div>
          <Button variant="ghost" className="mb-4" asChild>
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Home
            </Link>
          </Button>
          
          <div className="flex items-center justify-between mb-6">
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-3xl font-bold">Liquidity Analysis Dashboard</h1>
                <HelpButton
                  section="liquidity"
                  tooltip="Learn about liquidity metrics and trading insights"
                />
              </div>
              <p className="text-muted-foreground mt-2">
                Real-time ISX liquidity insights with safe trading recommendations
              </p>
            </div>
            <div className="flex items-center gap-4">
              <TooltipProvider>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">Scoring:</span>
                  <ToggleGroup 
                    type="single" 
                    value={scoringMode} 
                    onValueChange={(value) => setScoringMode(value as ScoringMode)}
                  >
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <ToggleGroupItem value="ema" aria-label="EMA scoring">
                          <TrendingUp className="h-4 w-4 mr-1" />
                          EMA
                        </ToggleGroupItem>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p className="font-semibold">20-period Exponential Moving Average</p>
                        <p className="text-xs">Smooths out anomalies, most stable</p>
                      </TooltipContent>
                    </Tooltip>
                    
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <ToggleGroupItem value="latest" aria-label="Latest scoring">
                          <Clock className="h-4 w-4 mr-1" />
                          Recent
                        </ToggleGroupItem>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p className="font-semibold">Most recent trading day</p>
                        <p className="text-xs">Real-time but more volatile</p>
                      </TooltipContent>
                    </Tooltip>
                    
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <ToggleGroupItem value="average" aria-label="Average scoring">
                          <Calculator className="h-4 w-4 mr-1" />
                          Average
                        </ToggleGroupItem>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p className="font-semibold">Simple average</p>
                        <p className="text-xs">Balanced view of performance</p>
                      </TooltipContent>
                    </Tooltip>
                  </ToggleGroup>
                  
                  <Tooltip>
                    <TooltipTrigger>
                      <Info className="h-4 w-4 text-muted-foreground ml-1" />
                    </TooltipTrigger>
                    <TooltipContent className="max-w-xs">
                      <p className="font-semibold mb-1">Scoring Modes:</p>
                      <ul className="text-xs space-y-1">
                        <li><strong>EMA:</strong> Best for identifying trends, reduces outliers</li>
                        <li><strong>Recent:</strong> Best for current market conditions</li>
                        <li><strong>Average:</strong> Best for long-term perspective</li>
                      </ul>
                    </TooltipContent>
                  </Tooltip>
                </div>
              </TooltipProvider>
              
              <Button onClick={loadInsights} variant="outline" disabled={loading}>
                <RefreshCw className={cn("h-4 w-4 mr-2", loading && "animate-spin")} />
                Refresh
              </Button>
            </div>
          </div>
        </div>

        {/* Market Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Market Health</CardTitle>
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${marketHealthColor}`}>
                {insights.marketHealthScore.toFixed(1)}
              </div>
              <p className="text-xs text-muted-foreground mt-1">Overall liquidity score</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Coverage</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {insights.highQualityStocks}/{insights.totalStocks}
              </div>
              <p className="text-xs text-muted-foreground mt-1">Quality stocks analyzed</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Continuity</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {insights.averageContinuity.toFixed(1)}%
              </div>
              <p className="text-xs text-muted-foreground mt-1">Average trading frequency</p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium">Daily Volume</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatCurrency(insights.medianDailyVolume)}
              </div>
              <p className="text-xs text-muted-foreground mt-1">Median trading value</p>
            </CardContent>
          </Card>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search by symbol or company name..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>

        {/* Trading Thresholds Guide */}
        <Alert>
          <Info className="h-4 w-4" />
          <AlertTitle>Trading Threshold Guidelines</AlertTitle>
          <AlertDescription className="mt-2">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
              <div className="flex items-start gap-2">
                <Shield className="h-4 w-4 text-green-600 mt-0.5" />
                <div>
                  <strong>Conservative (0.5% impact):</strong> Minimal market impact, best for large institutional orders
                </div>
              </div>
              <div className="flex items-start gap-2">
                <Activity className="h-4 w-4 text-blue-600 mt-0.5" />
                <div>
                  <strong>Moderate (1% impact):</strong> Balanced approach for active traders
                </div>
              </div>
              <div className="flex items-start gap-2">
                <Zap className="h-4 w-4 text-orange-600 mt-0.5" />
                <div>
                  <strong>Aggressive (2% impact):</strong> Faster execution but higher market impact
                </div>
              </div>
              <div className="flex items-start gap-2">
                <Target className="h-4 w-4 text-purple-600 mt-0.5" />
                <div>
                  <strong>Optimal:</strong> Algorithm-recommended size considering all factors
                </div>
              </div>
            </div>
          </AlertDescription>
        </Alert>

        {/* Main Content Tabs */}
        <Tabs defaultValue="opportunities" className="space-y-4">
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="opportunities">Top Opportunities</TabsTrigger>
            <TabsTrigger value="large">Large Trades</TabsTrigger>
            <TabsTrigger value="daytrading">Day Trading</TabsTrigger>
            <TabsTrigger value="risk">High Risk</TabsTrigger>
            <TabsTrigger value="all">All Stocks</TabsTrigger>
          </TabsList>
          
          <TabsContent value="opportunities" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Top Trading Opportunities</CardTitle>
                <CardDescription>
                  Stocks with the best overall liquidity scores and trading conditions
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {filterStocks(getCategoryStocks(insights, 'topOpportunities')).map((stock) => (
                    <TradingCard key={stock.symbol} stock={stock} mode={scoringMode} />
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value="large" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Best for Large Trades</CardTitle>
                <CardDescription>
                  Stocks that can handle significant trade sizes with minimal impact
                </CardDescription>
              </CardHeader>
              <CardContent>
                {getCategoryStocks(insights, 'bestForLargeTrades').length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filterStocks(getCategoryStocks(insights, 'bestForLargeTrades')).map((stock) => (
                      <TradingCard key={stock.symbol} stock={stock} mode={scoringMode} />
                    ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground text-center py-8">
                    No stocks currently meet large trade criteria
                  </p>
                )}
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value="daytrading" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Best for Day Trading</CardTitle>
                <CardDescription>
                  High continuity stocks suitable for frequent trading
                </CardDescription>
              </CardHeader>
              <CardContent>
                {getCategoryStocks(insights, 'bestForDayTrading').length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filterStocks(getCategoryStocks(insights, 'bestForDayTrading')).map((stock) => (
                      <TradingCard key={stock.symbol} stock={stock} mode={scoringMode} />
                    ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground text-center py-8">
                    No stocks currently meet day trading criteria
                  </p>
                )}
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value="risk" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>High Risk - Avoid</CardTitle>
                <CardDescription>
                  Stocks with poor liquidity or insufficient data quality
                </CardDescription>
              </CardHeader>
              <CardContent>
                {getCategoryStocks(insights, 'highRisk').length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filterStocks(getCategoryStocks(insights, 'highRisk')).map((stock) => (
                      <TradingCard key={stock.symbol} stock={stock} mode={scoringMode} />
                    ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground text-center py-8">
                    No high-risk stocks identified
                  </p>
                )}
              </CardContent>
            </Card>
          </TabsContent>
          
          <TabsContent value="all" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>All Analyzed Stocks</CardTitle>
                <CardDescription>
                  Complete list of all {isSSOTFormat(insights) ? insights.allStocks.length : 0} stocks with liquidity scores
                  {searchTerm && ` (showing ${
                    isSSOTFormat(insights) 
                      ? filterStocks(insights.allStocks).length 
                      : 0
                  } matching "${searchTerm}")`}
                </CardDescription>
              </CardHeader>
              <CardContent>
                {isSSOTFormat(insights) && insights.allStocks.length > 0 ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filterStocks(insights.allStocks)
                      .sort((a, b) => {
                        // Sort by score (best to worst)
                        const scoreA = scoringMode === 'ema' ? (a.ema20Score || a.score) :
                                       scoringMode === 'latest' ? (a.latestScore || a.score) :
                                       a.score
                        const scoreB = scoringMode === 'ema' ? (b.ema20Score || b.score) :
                                       scoringMode === 'latest' ? (b.latestScore || b.score) :
                                       b.score
                        return scoreB - scoreA
                      })
                      .map((stock) => (
                        <TradingCard key={stock.symbol} stock={stock} mode={scoringMode} />
                      ))}
                  </div>
                ) : (
                  <p className="text-muted-foreground text-center py-8">
                    No stocks data available
                  </p>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Footer */}
        <div className="text-center text-sm text-muted-foreground">
          Generated: {new Date(insights.generatedAt).toLocaleString()}
        </div>
      </div>
    </main>
  )
}
