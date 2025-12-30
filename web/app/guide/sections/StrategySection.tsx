/**
 * Strategy Section - Interactive Guide
 * Preview of upcoming trading strategy builder and backtesting features
 */

'use client'

import { GuideSection } from '../components/GuideSection'
import { InteractiveDemo } from '../components/InteractiveDemo'
import { CodeExample } from '../components/CodeExample'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Target,
  TrendingUp,
  BarChart3,
  Code2,
  Construction,
  CheckCircle2,
  AlertTriangle,
  Lightbulb,
  BookOpen,
  Calendar
} from 'lucide-react'

export function StrategySection() {
  return (
    <GuideSection
      id="strategy"
      title="Trading Strategies"
      description="Strategy builder and backtesting system (coming soon)"
    >
      <div className="space-y-8">
        {/* Coming Soon Banner */}
        <Alert className="border-orange-200 dark:border-orange-800 bg-orange-50 dark:bg-orange-950/30">
          <Construction className="h-5 w-5 text-orange-600 dark:text-orange-400" />
          <AlertTitle className="text-lg font-semibold">🚧 Coming Soon: Strategy Builder & Backtesting</AlertTitle>
          <AlertDescription className="mt-2">
            <p className="mb-2">
              A comprehensive strategy development platform is planned for <strong>Phase 2 (Q2 2026)</strong>.
              This section previews what to expect and how to prepare.
            </p>
            <div className="flex items-center gap-2 mt-3">
              <Calendar className="h-4 w-4" />
              <span className="text-sm font-medium">Expected Release: April-June 2026</span>
            </div>
          </AlertDescription>
        </Alert>

        {/* Introduction */}
        <div className="space-y-4">
          <div className="flex items-start gap-4">
            <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
              <Target className="h-6 w-6 text-purple-600 dark:text-purple-400" />
            </div>
            <div className="flex-1">
              <h2 className="text-2xl font-semibold mb-2">What is Strategy Builder?</h2>
              <p className="text-muted-foreground leading-relaxed">
                The Strategy Builder will allow you to create, test, and optimize automated trading strategies
                without writing code. For advanced users, a code-based API (Go/TypeScript) will provide
                full control over strategy logic, risk management, and execution.
              </p>
            </div>
          </div>
        </div>

        {/* What to Expect */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded" />
            <h2 className="text-2xl font-semibold">What to Expect</h2>
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            {/* Visual Strategy Builder */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BarChart3 className="h-5 w-5 text-blue-500" />
                  Visual Strategy Builder
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Drag-and-drop interface for non-coders. Build strategies using visual blocks:
                </p>
                <ul className="space-y-2 text-sm text-muted-foreground">
                  <li className="flex items-start gap-2">
                    <span className="text-blue-500 mt-0.5">•</span>
                    <span><strong>Entry Conditions</strong>: Combine indicators (RSI &lt; 30 AND SMA50 &gt; SMA200)</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-blue-500 mt-0.5">•</span>
                    <span><strong>Exit Rules</strong>: Profit targets, stop-losses, time-based exits</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-blue-500 mt-0.5">•</span>
                    <span><strong>Position Sizing</strong>: Fixed %, risk-based, Kelly criterion</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-blue-500 mt-0.5">•</span>
                    <span><strong>Filters</strong>: Volume, market cap, sector constraints</span>
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Code-Based Builder */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Code2 className="h-5 w-5 text-green-500" />
                  Code-Based Builder
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  For developers and advanced traders who need full control:
                </p>
                <ul className="space-y-2 text-sm text-muted-foreground">
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">•</span>
                    <span><strong>Go API</strong>: High-performance backtesting engine</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">•</span>
                    <span><strong>TypeScript SDK</strong>: Frontend integration for custom dashboards</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">•</span>
                    <span><strong>Custom Indicators</strong>: Write your own technical indicators</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-green-500 mt-0.5">•</span>
                    <span><strong>ML Integration</strong>: Connect machine learning models</span>
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Backtesting Engine */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <TrendingUp className="h-5 w-5 text-orange-500" />
                  Backtesting Engine
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Test strategies on 5+ years of ISX historical data:
                </p>
                <ul className="space-y-2 text-sm text-muted-foreground">
                  <li className="flex items-start gap-2">
                    <span className="text-orange-500 mt-0.5">•</span>
                    <span><strong>Tick-by-tick simulation</strong>: Realistic order fills and slippage</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-orange-500 mt-0.5">•</span>
                    <span><strong>Transaction costs</strong>: Model commissions and fees</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-orange-500 mt-0.5">•</span>
                    <span><strong>Walk-forward analysis</strong>: Prevent overfitting with out-of-sample tests</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-orange-500 mt-0.5">•</span>
                    <span><strong>Monte Carlo simulation</strong>: Assess strategy robustness</span>
                  </li>
                </ul>
              </CardContent>
            </Card>

            {/* Performance Metrics */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <BarChart3 className="h-5 w-5 text-purple-500" />
                  30+ Performance Metrics
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Comprehensive analysis of strategy performance:
                </p>
                <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                  <div>
                    <strong className="text-foreground">Returns</strong>
                    <ul className="space-y-1 mt-1">
                      <li>• CAGR</li>
                      <li>• Total Return</li>
                      <li>• Annualized Return</li>
                    </ul>
                  </div>
                  <div>
                    <strong className="text-foreground">Risk</strong>
                    <ul className="space-y-1 mt-1">
                      <li>• Sharpe Ratio</li>
                      <li>• Sortino Ratio</li>
                      <li>• Max Drawdown</li>
                    </ul>
                  </div>
                  <div>
                    <strong className="text-foreground">Win Rate</strong>
                    <ul className="space-y-1 mt-1">
                      <li>• Win %</li>
                      <li>• Profit Factor</li>
                      <li>• Avg Win/Loss</li>
                    </ul>
                  </div>
                  <div>
                    <strong className="text-foreground">Exposure</strong>
                    <ul className="space-y-1 mt-1">
                      <li>• Time in Market</li>
                      <li>• Avg Hold Period</li>
                      <li>• Turnover</li>
                    </ul>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Get Ready: Prerequisites */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-green-500 to-teal-500 rounded" />
            <h2 className="text-2xl font-semibold">Get Ready: Prerequisites</h2>
          </div>

          <p className="text-muted-foreground leading-relaxed">
            To make the most of the Strategy Builder when it launches, start preparing now by mastering
            these foundational concepts:
          </p>

          <div className="space-y-3">
            {/* Master Technical Indicators */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                  1. Master Technical Indicators
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Strategies rely heavily on indicator signals. Complete the <a href="/guide?section=indicators" className="text-blue-600 dark:text-blue-400 hover:underline">Indicators section</a> to understand:
                </p>
                <ul className="space-y-2 text-sm text-muted-foreground ml-4">
                  <li>• <strong>RSI</strong>: When is it truly oversold vs false signal?</li>
                  <li>• <strong>MACD</strong>: How to confirm trends with histogram divergence</li>
                  <li>• <strong>Moving Averages</strong>: Golden cross vs death cross timing</li>
                  <li>• <strong>Bollinger Bands</strong>: Squeeze breakouts vs band walks</li>
                  <li>• <strong>Volume</strong>: Confirming price moves with volume spikes</li>
                </ul>
                <Button size="sm" variant="outline" asChild>
                  <a href="/guide?section=indicators">
                    <BookOpen className="h-4 w-4 mr-2" />
                    Go to Indicators Section
                  </a>
                </Button>
              </CardContent>
            </Card>

            {/* Learn Chart Patterns */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <CheckCircle2 className="h-5 w-5 text-blue-500" />
                  2. Learn Chart Patterns
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Many successful strategies combine indicators with chart patterns:
                </p>
                <ul className="space-y-2 text-sm text-muted-foreground ml-4">
                  <li>• <strong>Support & Resistance</strong>: See <a href="/guide?section=indicator-support-resistance" className="text-blue-600 dark:text-blue-400 hover:underline">Support/Resistance guide</a></li>
                  <li>• <strong>Trendlines</strong>: Drawing and respecting trend channels</li>
                  <li>• <strong>Breakouts</strong>: Recognizing valid vs false breakouts (volume confirmation)</li>
                  <li>• <strong>Candlestick Patterns</strong>: Doji, hammer, engulfing (future guide section)</li>
                </ul>
              </CardContent>
            </Card>

            {/* Understand Risk Management */}
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <AlertTriangle className="h-5 w-5 text-orange-500" />
                  3. Understand Risk Management
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  Even profitable strategies fail without proper risk management:
                </p>
                <div className="space-y-3 mt-3">
                  <div className="bg-muted p-3 rounded-lg">
                    <h4 className="font-semibold text-sm mb-2">Position Sizing</h4>
                    <p className="text-xs text-muted-foreground">
                      <strong>Fixed %</strong>: Risk 2% of capital per trade (e.g., $10,000 capital → $200 risk)<br />
                      <strong>Volatility-based</strong>: Size based on ATR (Average True Range)<br />
                      <strong>Kelly Criterion</strong>: Optimal fraction based on win rate and payoff ratio
                    </p>
                  </div>
                  <div className="bg-muted p-3 rounded-lg">
                    <h4 className="font-semibold text-sm mb-2">Stop-Loss Placement</h4>
                    <p className="text-xs text-muted-foreground">
                      <strong>ATR-based</strong>: 2x ATR below entry (adapts to volatility)<br />
                      <strong>Percentage-based</strong>: Fixed 5% loss limit<br />
                      <strong>Support-based</strong>: Just below key support level<br />
                      <strong>Trailing stop</strong>: Lock in profits as price moves favorably
                    </p>
                  </div>
                  <div className="bg-muted p-3 rounded-lg">
                    <h4 className="font-semibold text-sm mb-2">Risk-Reward Ratios</h4>
                    <p className="text-xs text-muted-foreground">
                      <strong>Minimum 2:1</strong>: For every $1 risked, target $2 profit<br />
                      <strong>Break-even win rate</strong>: 33% with 2:1 ratio, 25% with 3:1 ratio<br />
                      <strong>Example</strong>: Risk $100 (stop at $1.90), target $300 (exit at $2.20) = 3:1 ratio
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Strategy Templates Preview */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded" />
            <h2 className="text-2xl font-semibold">Strategy Templates Preview</h2>
          </div>

          <p className="text-muted-foreground leading-relaxed">
            The Strategy Builder will include pre-built templates you can customize. Here are 5 popular
            strategies that will be available:
          </p>

          <div className="space-y-4">
            {/* Template 1: SMA Golden Cross */}
            <Card>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Badge variant="secondary">Trend-Following</Badge>
                      SMA Golden Cross
                    </CardTitle>
                    <CardDescription>Classic long-term trend reversal strategy</CardDescription>
                  </div>
                  <Badge className="bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400">
                    Win Rate: 40-45%
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-green-500" />
                      Entry
                    </h4>
                    <p className="text-muted-foreground">
                      Buy when SMA(50) crosses <strong>above</strong> SMA(200) — the "Golden Cross"
                    </p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Target className="h-4 w-4 text-red-500" />
                      Exit
                    </h4>
                    <p className="text-muted-foreground">
                      Sell when SMA(50) crosses <strong>below</strong> SMA(200) OR 5% stop-loss
                    </p>
                  </div>
                </div>
                <div className="bg-muted p-3 rounded text-xs text-muted-foreground">
                  <strong>Position Size:</strong> 10% of capital per trade<br />
                  <strong>Typical Performance:</strong> Sharpe 0.8-1.2, CAGR 8-12%, Max DD 15-20%<br />
                  <strong>Best For:</strong> Patient traders, low-frequency trading (1-3 trades/year per stock)
                </div>
              </CardContent>
            </Card>

            {/* Template 2: RSI Mean Reversion */}
            <Card>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Badge variant="secondary">Mean-Reversion</Badge>
                      RSI Mean Reversion
                    </CardTitle>
                    <CardDescription>Buy oversold, sell when normalized</CardDescription>
                  </div>
                  <Badge className="bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400">
                    Win Rate: 55-60%
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-green-500" />
                      Entry
                    </h4>
                    <p className="text-muted-foreground">
                      Buy when RSI &lt; 30 (oversold) <strong>AND</strong> price above SMA(200)
                    </p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Target className="h-4 w-4 text-red-500" />
                      Exit
                    </h4>
                    <p className="text-muted-foreground">
                      Sell when RSI &gt; 50 <strong>OR</strong> 3% profit <strong>OR</strong> 5% stop-loss
                    </p>
                  </div>
                </div>
                <div className="bg-muted p-3 rounded text-xs text-muted-foreground">
                  <strong>Position Size:</strong> 5% of capital per trade<br />
                  <strong>Typical Performance:</strong> Sharpe 1.2-1.8, CAGR 12-18%, Max DD 10-15%<br />
                  <strong>Best For:</strong> Active traders, higher frequency (10-20 trades/year per stock)
                </div>
              </CardContent>
            </Card>

            {/* Template 3: MACD Momentum */}
            <Card>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Badge variant="secondary">Momentum</Badge>
                      MACD Momentum
                    </CardTitle>
                    <CardDescription>Ride strong trends with trailing stops</CardDescription>
                  </div>
                  <Badge className="bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400">
                    Win Rate: 45-50%
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-green-500" />
                      Entry
                    </h4>
                    <p className="text-muted-foreground">
                      Buy when MACD line crosses above signal line <strong>AND</strong> histogram &gt; 0
                    </p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Target className="h-4 w-4 text-red-500" />
                      Exit
                    </h4>
                    <p className="text-muted-foreground">
                      Sell when MACD crosses below signal <strong>OR</strong> trailing stop (2x ATR)
                    </p>
                  </div>
                </div>
                <div className="bg-muted p-3 rounded text-xs text-muted-foreground">
                  <strong>Position Size:</strong> 8% of capital per trade<br />
                  <strong>Typical Performance:</strong> Sharpe 1.0-1.5, CAGR 10-15%, Max DD 12-18%<br />
                  <strong>Best For:</strong> Trend riders, medium frequency (5-10 trades/year per stock)
                </div>
              </CardContent>
            </Card>

            {/* Template 4: Bollinger Band Squeeze */}
            <Card>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Badge variant="secondary">Volatility Breakout</Badge>
                      Bollinger Band Squeeze
                    </CardTitle>
                    <CardDescription>Trade explosive moves after low volatility</CardDescription>
                  </div>
                  <Badge className="bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400">
                    Win Rate: 35-40%
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-green-500" />
                      Entry
                    </h4>
                    <p className="text-muted-foreground">
                      Buy when price breaks <strong>above</strong> upper band after squeeze (band width &lt; 10% of price)
                    </p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Target className="h-4 w-4 text-red-500" />
                      Exit
                    </h4>
                    <p className="text-muted-foreground">
                      Sell when price touches middle band <strong>OR</strong> 10% stop-loss
                    </p>
                  </div>
                </div>
                <div className="bg-muted p-3 rounded text-xs text-muted-foreground">
                  <strong>Position Size:</strong> 6% of capital per trade<br />
                  <strong>Typical Performance:</strong> Sharpe 0.9-1.3, CAGR 9-14%, Max DD 18-25%<br />
                  <strong>Best For:</strong> Patient traders waiting for explosive moves, low frequency (2-5 trades/year)
                </div>
              </CardContent>
            </Card>

            {/* Template 5: Support/Resistance Breakout */}
            <Card>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div>
                    <CardTitle className="flex items-center gap-2">
                      <Badge variant="secondary">Breakout</Badge>
                      Support/Resistance Breakout
                    </CardTitle>
                    <CardDescription>Trade confirmed level breaks with volume</CardDescription>
                  </div>
                  <Badge className="bg-teal-100 dark:bg-teal-900/30 text-teal-700 dark:text-teal-400">
                    Win Rate: 40-45%
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-green-500" />
                      Entry
                    </h4>
                    <p className="text-muted-foreground">
                      Buy when price breaks resistance <strong>with</strong> volume &gt; 2x average
                    </p>
                  </div>
                  <div>
                    <h4 className="font-semibold mb-2 flex items-center gap-2">
                      <Target className="h-4 w-4 text-red-500" />
                      Exit
                    </h4>
                    <p className="text-muted-foreground">
                      Sell when price falls below breakout level <strong>OR</strong> 15% profit
                    </p>
                  </div>
                </div>
                <div className="bg-muted p-3 rounded text-xs text-muted-foreground">
                  <strong>Position Size:</strong> 7% of capital per trade<br />
                  <strong>Typical Performance:</strong> Sharpe 1.1-1.4, CAGR 11-16%, Max DD 14-20%<br />
                  <strong>Best For:</strong> Chart pattern traders, medium frequency (7-12 trades/year)
                </div>
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Future API Preview */}
        <InteractiveDemo
          title="Future Strategy Builder API (Preview)"
          description="Code-based strategy creation (coming in Phase 2)"
        >
          <CodeExample
            language="typescript"
            code={`// Future Strategy Builder API (coming soon)
// This is a preview - not yet functional

import { StrategyBuilder, Indicators, Conditions } from '@/lib/strategy'

// Create a new strategy
const strategy = new StrategyBuilder('RSI Mean Reversion')

// Define entry conditions (AND logic)
strategy.addEntry(
  Conditions.indicator('rsi', { period: 14, condition: '<', value: 30 }),
  Conditions.indicator('sma', { period: 200, condition: 'above' }) // Price above SMA200
)

// Define exit conditions (OR logic - any triggers exit)
strategy.addExit(
  Conditions.indicator('rsi', { period: 14, condition: '>', value: 50 }),
  Conditions.profit({ percent: 3 }),
  Conditions.stopLoss({ percent: 5 })
)

// Set position sizing
strategy.setPositionSize({ percent: 5 }) // 5% of capital per trade

// Configure risk management
strategy.setRiskManagement({
  maxOpenPositions: 5,
  maxDailyLoss: 10, // percent
  maxCorrelation: 0.7 // Avoid correlated positions
})

// Run backtest on historical data
const results = await strategy.backtest({
  startDate: '2020-01-01',
  endDate: '2025-01-01',
  tickers: ['BANK', 'ASIACELL', 'TAQA'], // Or 'ALL' for universe
  slippage: 0.001, // 0.1% slippage
  commission: 0.002 // 0.2% commission
})

// View results
console.log('Total Return:', results.totalReturn, '%')
console.log('CAGR:', results.cagr, '%')
console.log('Sharpe Ratio:', results.sharpe)
console.log('Max Drawdown:', results.maxDrawdown, '%')
console.log('Win Rate:', results.winRate, '%')
console.log('Profit Factor:', results.profitFactor)
console.log('Total Trades:', results.totalTrades)

// Export results
strategy.exportResults('results.csv')
strategy.exportChart('equity_curve.png')`}
          />
        </InteractiveDemo>

        {/* Preparation Checklist */}
        <Card className="border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950/30">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Lightbulb className="h-5 w-5 text-green-600 dark:text-green-400" />
              Preparation Checklist
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Get ready for the Strategy Builder by completing these tasks:
            </p>
            <div className="space-y-2">
              <label className="flex items-start gap-3">
                <input type="checkbox" className="mt-1" />
                <span className="text-sm">Complete the <a href="/guide?section=indicators" className="text-blue-600 dark:text-blue-400 hover:underline">Indicators section</a> (all 22 indicators)</span>
              </label>
              <label className="flex items-start gap-3">
                <input type="checkbox" className="mt-1" />
                <span className="text-sm">Practice analyzing 10+ ISX stocks manually using indicators</span>
              </label>
              <label className="flex items-start gap-3">
                <input type="checkbox" className="mt-1" />
                <span className="text-sm">Read about backtesting pitfalls (lookahead bias, overfitting, survivorship bias)</span>
              </label>
              <label className="flex items-start gap-3">
                <input type="checkbox" className="mt-1" />
                <span className="text-sm">Define your trading goals (time horizon: short/medium/long-term, risk tolerance: low/medium/high)</span>
              </label>
              <label className="flex items-start gap-3">
                <input type="checkbox" className="mt-1" />
                <span className="text-sm">Understand position sizing and risk management formulas</span>
              </label>
              <label className="flex items-start gap-3">
                <input type="checkbox" className="mt-1" />
                <span className="text-sm">Bookmark this section for updates when Strategy Builder launches</span>
              </label>
            </div>
          </CardContent>
        </Card>

        {/* Next Steps */}
        <div className="space-y-3">
          <h2 className="text-xl font-semibold">Next Steps</h2>
          <div className="flex flex-wrap gap-3">
            <Button variant="outline" asChild>
              <a href="/guide?section=market-overview">
                <BookOpen className="h-4 w-4 mr-2" />
                Explore Market Overview
              </a>
            </Button>
            <Button variant="outline" asChild>
              <a href="/guide?section=quick-reference">
                Quick Reference →
              </a>
            </Button>
            <Button variant="outline" asChild>
              <a href="/strategy">
                <TrendingUp className="h-4 w-4 mr-2" />
                Run RSI Strategy
              </a>
            </Button>
          </div>
        </div>
      </div>
    </GuideSection>
  )
}
