'use client'

import { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Separator } from '@/components/ui/separator'
import { Calculator, Loader2, AlertCircle, Info, BarChart3, CheckCircle2 } from 'lucide-react'
import { TradingCard } from '@/app/liquidity/liquidity-dashboard'
import type { StockRecommendation, ScoringMode, LiquidityInsightsSSOT } from '@/app/liquidity/liquidity-dashboard'

/**
 * Interactive Liquidity Calculator Demo - Phase 1 (Real API Data)
 * Uses production API and TradingCard component for accurate calculations
 */
export function LiquidityCalculatorDemo() {
  const [liquidityData, setLiquidityData] = useState<LiquidityInsightsSSOT | null>(null)
  const [selectedSymbol, setSelectedSymbol] = useState<string>('BBOB')
  const [mode, setMode] = useState<ScoringMode>('ema')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Fetch liquidity data from API
  const fetchLiquidityData = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)

      const response = await fetch(`/api/liquidity/insights?mode=${mode}`)
      if (!response.ok) {
        throw new Error('Failed to load liquidity data')
      }

      const data = await response.json()

      // Validate SSOT format
      if (!data.allStocks || !Array.isArray(data.allStocks)) {
        throw new Error('Invalid data format received from API')
      }

      setLiquidityData(data)

      // Set default selection to first stock if current selection not found
      if (data.allStocks.length > 0 && !data.allStocks.find((s: StockRecommendation) => s.symbol === selectedSymbol)) {
        setSelectedSymbol(data.allStocks[0].symbol)
      }
    } catch (err) {
      console.error('Error loading liquidity data:', err)
      setError(err instanceof Error ? err.message : 'Failed to load liquidity data')
    } finally {
      setLoading(false)
    }
  }, [mode, selectedSymbol])

  // Load data on mount and when mode changes
  useEffect(() => {
    fetchLiquidityData()
  }, [fetchLiquidityData])

  // Get selected stock data
  const selectedStock = liquidityData?.allStocks.find(s => s.symbol === selectedSymbol)

  // Loading state
  if (loading) {
    return (
      <div className="space-y-6">
        <Card>
          <CardContent className="pt-6 flex flex-col items-center justify-center min-h-[400px]">
            <Loader2 className="h-12 w-12 animate-spin text-primary mb-4" />
            <p className="text-muted-foreground">Loading liquidity data from production API...</p>
            <p className="text-xs text-muted-foreground mt-2">Mode: {mode.toUpperCase()}</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="space-y-6">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Failed to Load Data</AlertTitle>
          <AlertDescription className="space-y-2">
            <p>{error}</p>
            <p className="text-xs mt-2">
              Make sure you've run the Full Pipeline operation to generate liquidity data.
            </p>
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // No data state
  if (!liquidityData || !liquidityData.allStocks || liquidityData.allStocks.length === 0) {
    return (
      <div className="space-y-6">
        <Alert>
          <Info className="h-4 w-4" />
          <AlertTitle>No Data Available</AlertTitle>
          <AlertDescription>
            No liquidity data found. Please run the Full Pipeline operation first to generate insights.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // Helper to get score based on mode
  const getScoreForMode = (stock: StockRecommendation): number => {
    switch (mode) {
      case 'ema':
        return stock.ema20Score || stock.score
      case 'latest':
        return stock.latestScore || stock.score
      case 'average':
        return stock.score
      default:
        return stock.score
    }
  }

  // Sort stocks by score (best to worst)
  const sortedStocks = [...liquidityData.allStocks].sort((a, b) => {
    const scoreA = getScoreForMode(a)
    const scoreB = getScoreForMode(b)
    return scoreB - scoreA
  })

  return (
    <div className="space-y-6">
      {/* Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Calculator className="h-5 w-5" />
            Liquidity Analysis Controls
          </CardTitle>
          <CardDescription>
            Select stock and scoring mode to see production-accurate liquidity calculations
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Mode Selector */}
          <div className="space-y-2">
            <Label>Scoring Mode</Label>
            <Select value={mode} onValueChange={(value) => setMode(value as ScoringMode)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="ema">
                  <div className="flex flex-col items-start">
                    <span className="font-semibold">EMA (20-period)</span>
                    <span className="text-xs text-muted-foreground">Smoothed trend, most stable</span>
                  </div>
                </SelectItem>
                <SelectItem value="latest">
                  <div className="flex flex-col items-start">
                    <span className="font-semibold">Latest (Recent)</span>
                    <span className="text-xs text-muted-foreground">Most recent day, more volatile</span>
                  </div>
                </SelectItem>
                <SelectItem value="average">
                  <div className="flex flex-col items-start">
                    <span className="font-semibold">Average</span>
                    <span className="text-xs text-muted-foreground">Simple average, balanced view</span>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Stock Selector */}
          <div className="space-y-2">
            <Label>Stock Ticker ({sortedStocks.length} available)</Label>
            <Select value={selectedSymbol} onValueChange={setSelectedSymbol}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[300px]">
                {sortedStocks.map((stock) => {
                  const stockScore = getScoreForMode(stock)
                  return (
                    <SelectItem key={stock.symbol} value={stock.symbol}>
                      <div className="flex items-center gap-2">
                        <span className="font-mono font-semibold">{stock.symbol}</span>
                        <span className="text-muted-foreground">-</span>
                        <span className="text-xs">{stock.companyName || 'Unknown'}</span>
                        <span className="text-muted-foreground">-</span>
                        <span className="font-semibold text-xs">{stockScore.toFixed(1)}</span>
                      </div>
                    </SelectItem>
                  )
                })}
              </SelectContent>
            </Select>
          </div>

          {/* Data Source Info */}
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription className="text-xs">
              <strong>Production Data:</strong> This demo uses real API data from{' '}
              <code className="bg-muted px-1 py-0.5 rounded">/api/liquidity/insights?mode={mode}</code>.
              All calculations match the production Liquidity page exactly.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>

      {/* Selected Stock Display (Production TradingCard) */}
      {selectedStock && (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Production Trading Card</CardTitle>
              <CardDescription>
                Identical to the card shown on the /liquidity page
              </CardDescription>
            </CardHeader>
            <CardContent>
              <TradingCard stock={selectedStock} mode={mode} />
            </CardContent>
          </Card>

          {/* Calculation Breakdown */}
          <Card className="border-primary/50">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <BarChart3 className="h-5 w-5" />
                Calculation Breakdown (3-Metric Hybrid System)
              </CardTitle>
              <CardDescription>
                Step-by-step calculation showing how the liquidity score is computed
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Component 1: ILLIQ Score */}
              {selectedStock.illiqScore !== undefined && (
                <div className="p-4 rounded-lg border border-red-200 dark:border-red-900 bg-red-50/30 dark:bg-red-950/20">
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-semibold text-sm">Component 1: Price Impact (ILLIQ)</h4>
                    <span className="text-xs text-muted-foreground">Weight: 40%</span>
                  </div>

                  <div className="space-y-2 text-sm">
                    <div className="flex items-start gap-2 p-2 bg-background/50 rounded">
                      <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-mono text-xs">
                          Raw ILLIQ = |Price Change| / Trading Value
                        </p>
                        {selectedStock.illiqRaw !== undefined && (
                          <p className="text-muted-foreground text-xs mt-1">
                            = {selectedStock.illiqRaw.toFixed(6)} (after winsorization)
                          </p>
                        )}
                      </div>
                    </div>

                    <div className="flex items-start gap-2 p-2 bg-background/50 rounded">
                      <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-mono text-xs">
                          Cross-sectional scaling across all stocks
                        </p>
                        <p className="text-muted-foreground text-xs mt-1">
                          Uses piecewise linear function with market-wide percentiles
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start gap-2 p-2 bg-primary/10 rounded border border-primary/30">
                      <Info className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-semibold text-xs">Scaled ILLIQ Score</p>
                        <p className="font-mono text-lg text-primary">{selectedStock.illiqScore.toFixed(1)} / 100</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          Weighted Contribution: {(selectedStock.illiqScore * 0.40).toFixed(1)} points
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Component 2: Volume Score */}
              {selectedStock.volumeScore !== undefined && (
                <div className="p-4 rounded-lg border border-blue-200 dark:border-blue-900 bg-blue-50/30 dark:bg-blue-950/20">
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-semibold text-sm">Component 2: Trading Volume</h4>
                    <span className="text-xs text-muted-foreground">Weight: 35%</span>
                  </div>

                  <div className="space-y-2 text-sm">
                    <div className="flex items-start gap-2 p-2 bg-background/50 rounded">
                      <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-mono text-xs">
                          Raw Volume = SMA(60) of daily trading value
                        </p>
                        {selectedStock.volumeRaw !== undefined && (
                          <p className="text-muted-foreground text-xs mt-1">
                            = {(selectedStock.volumeRaw / 1_000_000).toFixed(2)} M IQD per day
                          </p>
                        )}
                      </div>
                    </div>

                    <div className="flex items-start gap-2 p-2 bg-background/50 rounded">
                      <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-mono text-xs">
                          Cross-sectional scaling with robust normalization
                        </p>
                        <p className="text-muted-foreground text-xs mt-1">
                          Compares against all stocks to determine relative market depth
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start gap-2 p-2 bg-primary/10 rounded border border-primary/30">
                      <Info className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-semibold text-xs">Scaled Volume Score</p>
                        <p className="font-mono text-lg text-primary">{selectedStock.volumeScore.toFixed(1)} / 100</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          Weighted Contribution: {(selectedStock.volumeScore * 0.35).toFixed(1)} points
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Component 3: Continuity Score */}
              {selectedStock.continuityScore !== undefined && (
                <div className="p-4 rounded-lg border border-green-200 dark:border-green-900 bg-green-50/30 dark:bg-green-950/20">
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-semibold text-sm">Component 3: Trading Continuity</h4>
                    <span className="text-xs text-muted-foreground">Weight: 25%</span>
                  </div>

                  <div className="space-y-2 text-sm">
                    <div className="flex items-start gap-2 p-2 bg-background/50 rounded">
                      <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-mono text-xs">
                          Raw Continuity = Trading Days / Total Days
                        </p>
                        <p className="text-muted-foreground text-xs mt-1">
                          = {(selectedStock.continuity * 100).toFixed(1)}% of days with trades
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start gap-2 p-2 bg-background/50 rounded">
                      <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-mono text-xs">
                          Direct percentage scaling (0-100)
                        </p>
                        <p className="text-muted-foreground text-xs mt-1">
                          No cross-sectional adjustment needed for continuity
                        </p>
                      </div>
                    </div>

                    <div className="flex items-start gap-2 p-2 bg-primary/10 rounded border border-primary/30">
                      <Info className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-semibold text-xs">Continuity Score</p>
                        <p className="font-mono text-lg text-primary">{selectedStock.continuityScore.toFixed(1)} / 100</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          Weighted Contribution: {(selectedStock.continuityScore * 0.25).toFixed(1)} points
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              <Separator />

              {/* Final Score Calculation */}
              <div className="p-5 bg-gradient-to-br from-primary/10 to-primary/5 rounded-lg border-2 border-primary/30">
                <h4 className="font-semibold mb-3 flex items-center gap-2">
                  <Calculator className="h-4 w-4" />
                  Final Hybrid Liquidity Score ({mode.toUpperCase()} Mode)
                </h4>

                <div className="space-y-3">
                  {/* Formula */}
                  <div className="p-3 bg-background rounded font-mono text-sm">
                    <p className="text-muted-foreground text-xs mb-1">Formula:</p>
                    <p>HybridScore = (ILLIQ × 0.40) + (Volume × 0.35) + (Continuity × 0.25)</p>
                  </div>

                  {/* Calculation */}
                  <div className="p-3 bg-background rounded font-mono text-sm">
                    <p className="text-muted-foreground text-xs mb-1">Calculation:</p>
                    <p>
                      = ({selectedStock.illiqScore?.toFixed(1)} × 0.40) +
                      ({selectedStock.volumeScore?.toFixed(1)} × 0.35) +
                      ({selectedStock.continuityScore?.toFixed(1)} × 0.25)
                    </p>
                    <p className="mt-1">
                      = {((selectedStock.illiqScore || 0) * 0.40).toFixed(1)} +
                      {((selectedStock.volumeScore || 0) * 0.35).toFixed(1)} +
                      {((selectedStock.continuityScore || 0) * 0.25).toFixed(1)}
                    </p>
                  </div>

                  {/* Final Result */}
                  <div className="p-4 bg-primary text-primary-foreground rounded-lg text-center">
                    <p className="text-xs opacity-90 mb-1">Final Score</p>
                    <p className="text-4xl font-bold">
                      {getScoreForMode(selectedStock).toFixed(1)}
                    </p>
                    <p className="text-xs opacity-90 mt-2">
                      This matches the score shown in the production TradingCard above
                    </p>
                  </div>
                </div>
              </div>

              {/* Verification Note */}
              <Alert>
                <CheckCircle2 className="h-4 w-4 text-green-600" />
                <AlertTitle>Verification</AlertTitle>
                <AlertDescription className="text-xs">
                  Visit the <strong>/liquidity</strong> page and find <strong>{selectedStock.symbol}</strong> to verify
                  this score matches exactly. Change the scoring mode above to see how different modes produce
                  different scores for the same stock.
                </AlertDescription>
              </Alert>
            </CardContent>
          </Card>
        </>
      )}

      {/* Educational Note */}
      <Card className="border-primary/50 bg-primary/5">
        <CardContent className="pt-6">
          <div className="flex items-start gap-3">
            <Info className="h-5 w-5 text-primary shrink-0 mt-0.5" />
            <div className="text-sm space-y-2">
              <p className="font-semibold">About This Demo</p>
              <p className="text-muted-foreground">
                This interactive demo uses <strong>real production data</strong> from the ISX Pulse liquidity analysis engine.
                The 3-metric hybrid system combines Amihud Illiquidity (price impact), Trading Volume (market depth),
                and Continuity (temporal reliability) with 40%-35%-25% weights respectively.
              </p>
              <p className="text-muted-foreground">
                Unlike simplified educational examples, this demo uses <strong>cross-sectional scaling</strong> where
                each stock's raw metrics are compared against all other stocks on the same date using robust piecewise
                linear functions. This ensures scores accurately reflect relative liquidity in the ISX market context.
              </p>
              <p className="text-muted-foreground">
                The three scoring modes (EMA, Latest, Average) provide different perspectives on liquidity:
              </p>
              <ul className="list-disc list-inside space-y-1 text-muted-foreground ml-4">
                <li><strong>EMA (20-period):</strong> Smooths out daily volatility, best for trend identification</li>
                <li><strong>Latest (Recent):</strong> Shows current market conditions, more reactive to recent changes</li>
                <li><strong>Average:</strong> Balanced long-term view, less sensitive to outliers</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
