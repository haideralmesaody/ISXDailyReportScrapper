/**
 * Market Overview Section - Interactive Guide
 * Explains treemap visualization, trading calendar, and market metrics
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
  LayoutGrid,
  Calendar,
  TrendingUp,
  TrendingDown,
  ArrowLeft,
  ArrowRight,
  Info,
  ExternalLink,
  Sparkles
} from 'lucide-react'

export function MarketOverviewSection() {
  return (
    <GuideSection
      id="market-overview"
      title="Market Overview"
      description="Interactive treemap visualization and trading calendar for the Iraqi Stock Exchange"
    >
      <div className="space-y-8">
        {/* Introduction */}
        <div className="space-y-4">
          <div className="flex items-start gap-4">
            <div className="p-3 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
              <LayoutGrid className="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div className="flex-1">
              <h2 className="text-2xl font-semibold mb-2">What is Market Overview?</h2>
              <p className="text-muted-foreground leading-relaxed">
                The Market Overview page (<code className="px-2 py-1 bg-muted rounded">/market-overview</code>) provides
                a visual snapshot of the entire Iraqi Stock Exchange in a single view. It uses an interactive treemap
                to show all stocks simultaneously, sized by trading value and colored by performance.
              </p>
              <div className="mt-4">
                <Button asChild>
                  <a href="/market-overview" target="_blank" className="gap-2">
                    <ExternalLink className="h-4 w-4" />
                    Open Market Overview
                  </a>
                </Button>
              </div>
            </div>
          </div>

          <Alert>
            <Sparkles className="h-4 w-4" />
            <AlertTitle>Perfect for Market Scanning</AlertTitle>
            <AlertDescription>
              Instead of analyzing stocks one-by-one, the treemap lets you spot winners, losers, and trading
              activity patterns at a glance. Ideal for daily market scans and opportunity discovery.
            </AlertDescription>
          </Alert>
        </div>

        {/* Treemap Visualization */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded" />
            <h2 className="text-2xl font-semibold">Treemap Visualization</h2>
          </div>

          <p className="text-muted-foreground leading-relaxed">
            The treemap is a data visualization technique that displays hierarchical data as nested rectangles.
            Each stock is represented by a rectangle, where the size and color encode important information.
          </p>

          {/* How to Read the Treemap */}
          <Card>
            <CardHeader>
              <CardTitle>How to Read the Treemap</CardTitle>
              <CardDescription>Understanding size, color, and layout</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Size Encoding */}
              <div>
                <h3 className="font-semibold mb-3 flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full bg-blue-500" />
                  Rectangle Size = Traded Value (IQD)
                </h3>
                <p className="text-sm text-muted-foreground mb-3">
                  The larger the rectangle, the more money was traded in that stock on the selected date.
                  This shows where market activity is concentrated.
                </p>
                <div className="grid grid-cols-3 gap-3">
                  <div className="p-4 border rounded-lg text-center">
                    <div className="h-24 bg-green-500/20 border-2 border-green-500 rounded mb-2 flex items-center justify-center">
                      <span className="text-xs font-mono">LARGE</span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      High trading value<br />
                      (e.g., 50M+ IQD)
                    </p>
                  </div>
                  <div className="p-4 border rounded-lg text-center">
                    <div className="h-16 bg-blue-500/20 border-2 border-blue-500 rounded mb-2 flex items-center justify-center">
                      <span className="text-xs font-mono">MEDIUM</span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Moderate trading<br />
                      (e.g., 10-50M IQD)
                    </p>
                  </div>
                  <div className="p-4 border rounded-lg text-center">
                    <div className="h-8 bg-gray-500/20 border-2 border-gray-500 rounded mb-2 flex items-center justify-center">
                      <span className="text-xs font-mono">SMALL</span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Low trading<br />
                      (e.g., &lt;10M IQD)
                    </p>
                  </div>
                </div>
              </div>

              {/* Color Encoding */}
              <div>
                <h3 className="font-semibold mb-3 flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full bg-green-500" />
                  Rectangle Color = Price Change %
                </h3>
                <p className="text-sm text-muted-foreground mb-3">
                  The color gradient indicates how much the stock price changed compared to the previous trading day.
                  Green = gainers, Red = losers, Gray = unchanged.
                </p>
                <div className="space-y-2">
                  <div className="flex items-center gap-3">
                    <div className="w-20 h-8 bg-gradient-to-r from-green-700 to-green-600 rounded" />
                    <span className="text-sm font-medium">Dark Green</span>
                    <span className="text-xs text-muted-foreground">Strong gainers (&gt;3%)</span>
                    <TrendingUp className="h-4 w-4 text-green-600 ml-auto" />
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-20 h-8 bg-gradient-to-r from-green-500 to-green-400 rounded" />
                    <span className="text-sm font-medium">Light Green</span>
                    <span className="text-xs text-muted-foreground">Moderate gainers (0-3%)</span>
                    <TrendingUp className="h-4 w-4 text-green-500 ml-auto" />
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-20 h-8 bg-gray-400 rounded" />
                    <span className="text-sm font-medium">Gray</span>
                    <span className="text-xs text-muted-foreground">Unchanged (0%)</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-20 h-8 bg-gradient-to-r from-red-400 to-red-500 rounded" />
                    <span className="text-sm font-medium">Light Red</span>
                    <span className="text-xs text-muted-foreground">Moderate losers (0 to -3%)</span>
                    <TrendingDown className="h-4 w-4 text-red-500 ml-auto" />
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-20 h-8 bg-gradient-to-r from-red-600 to-red-700 rounded" />
                    <span className="text-sm font-medium">Dark Red</span>
                    <span className="text-xs text-muted-foreground">Strong losers (&lt;-3%)</span>
                    <TrendingDown className="h-4 w-4 text-red-600 ml-auto" />
                  </div>
                </div>
              </div>

              {/* Interactive Features */}
              <div>
                <h3 className="font-semibold mb-3 flex items-center gap-2">
                  <div className="h-2 w-2 rounded-full bg-purple-500" />
                  Interactive Features
                </h3>
                <ul className="space-y-2 text-sm text-muted-foreground">
                  <li className="flex items-start gap-2">
                    <span className="text-purple-500 mt-0.5">•</span>
                    <span><strong>Hover</strong>: See tooltip with ticker details (symbol, price, change %, traded value)</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-purple-500 mt-0.5">•</span>
                    <span><strong>Click</strong>: Navigate to Analysis page for detailed chart analysis</span>
                  </li>
                  <li className="flex items-start gap-2">
                    <span className="text-purple-500 mt-0.5">•</span>
                    <span><strong>Responsive</strong>: Automatically resizes based on window width</span>
                  </li>
                </ul>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Trading Calendar */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-orange-500 to-red-500 rounded" />
            <h2 className="text-2xl font-semibold">ISX Trading Calendar</h2>
          </div>

          <p className="text-muted-foreground leading-relaxed">
            The trading calendar accounts for the unique Iraqi weekend schedule. The Iraqi Stock Exchange
            operates Sunday through Thursday, with Friday and Saturday as weekend days.
          </p>

          <Card>
            <CardHeader>
              <CardTitle>Calendar Features</CardTitle>
              <CardDescription>ISX-specific date handling</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Iraqi Weekend */}
              <div>
                <h3 className="font-semibold mb-3 flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-orange-500" />
                  Iraqi Weekend Styling
                </h3>
                <div className="grid grid-cols-7 gap-2 mb-3">
                  {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((day, i) => (
                    <div
                      key={day}
                      className={`p-3 rounded text-center font-medium text-sm ${
                        i >= 5
                          ? 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 border-2 border-red-300 dark:border-red-700'
                          : 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800'
                      }`}
                    >
                      {day}
                    </div>
                  ))}
                </div>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded" />
                    <span className="text-muted-foreground">Trading days (Sun-Thu)</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-red-100 dark:bg-red-900/30 border-2 border-red-300 dark:border-red-700 rounded" />
                    <span className="text-muted-foreground">Weekend (Fri-Sat)</span>
                  </div>
                </div>
              </div>

              {/* Date Picker */}
              <div>
                <h3 className="font-semibold mb-3 flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-blue-500" />
                  Date Picker Navigation
                </h3>
                <p className="text-sm text-muted-foreground mb-3">
                  The date picker automatically filters to show only dates with available trading data.
                  Use the dropdown to select any historical date, or use keyboard arrows for quick navigation.
                </p>
                <div className="bg-muted p-4 rounded-lg">
                  <div className="flex items-center gap-3 justify-center">
                    <Button size="sm" variant="outline">
                      <ArrowLeft className="h-4 w-4" />
                    </Button>
                    <Badge variant="secondary" className="px-4 py-2 text-base font-mono">
                      2025-10-10
                    </Badge>
                    <Button size="sm" variant="outline">
                      <ArrowRight className="h-4 w-4" />
                    </Button>
                  </div>
                  <p className="text-xs text-center text-muted-foreground mt-2">
                    Keyboard: ← Previous day | → Next day
                  </p>
                </div>
              </div>

              {/* Public Holidays */}
              <Alert>
                <Info className="h-4 w-4" />
                <AlertTitle>Public Holidays</AlertTitle>
                <AlertDescription className="text-sm">
                  The ISX is closed on Iraqi public holidays (Eid al-Fitr, Eid al-Adha, New Year, etc.).
                  These dates are automatically excluded from the date picker. If you select a non-trading
                  day, you'll see a message suggesting the latest available date.
                </AlertDescription>
              </Alert>
            </CardContent>
          </Card>
        </div>

        {/* How to Use */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-green-500 to-teal-500 rounded" />
            <h2 className="text-2xl font-semibold">How to Use Market Overview</h2>
          </div>

          <Card>
            <CardContent className="pt-6">
              <ol className="space-y-4">
                <li className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center font-semibold">
                    1
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold mb-1">Select Trading Date</h3>
                    <p className="text-sm text-muted-foreground">
                      Use the date picker at the top of the page to select any trading day. The latest
                      date is selected by default. Use arrow keys (← →) for quick navigation between dates.
                    </p>
                  </div>
                </li>
                <li className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center font-semibold">
                    2
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold mb-1">Examine the Treemap</h3>
                    <p className="text-sm text-muted-foreground">
                      Scan the treemap to identify the day's hottest and coldest stocks. Look for:
                    </p>
                    <ul className="mt-2 space-y-1 text-sm text-muted-foreground ml-4">
                      <li>• <strong>Large dark green rectangles</strong>: High-volume gainers (strong momentum)</li>
                      <li>• <strong>Large dark red rectangles</strong>: High-volume losers (potential reversal or exit signal)</li>
                      <li>• <strong>Small rectangles</strong>: Low liquidity stocks (higher risk, harder to trade)</li>
                    </ul>
                  </div>
                </li>
                <li className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center font-semibold">
                    3
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold mb-1">Click to Analyze</h3>
                    <p className="text-sm text-muted-foreground">
                      Click on any ticker rectangle to navigate to the Analysis page for detailed chart analysis.
                      From there, you can add indicators, examine candlestick patterns, and make trading decisions.
                    </p>
                  </div>
                </li>
                <li className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center font-semibold">
                    4
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold mb-1">Review Market Summary</h3>
                    <p className="text-sm text-muted-foreground">
                      Check the market summary card for aggregate metrics: total traded value, volume,
                      number of trades, and advancers vs decliners count. This gives you a sense of overall
                      market health and breadth.
                    </p>
                  </div>
                </li>
              </ol>
            </CardContent>
          </Card>
        </div>

        {/* Market Summary Metrics */}
        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded" />
            <h2 className="text-2xl font-semibold">Market Summary Metrics</h2>
          </div>

          <p className="text-muted-foreground leading-relaxed">
            The market summary card displays aggregate statistics for the selected trading date. These
            metrics help you understand overall market activity and sentiment.
          </p>

          <div className="grid md:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Total Traded Value (IQD)</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                The sum of all money that changed hands (price × volume for all trades). Higher values
                indicate greater market activity and liquidity. Typical range: 50M - 500M IQD per day.
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Total Volume (Shares)</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                The total number of shares traded across all stocks. Provides a sense of market participation.
                Compare with previous days to spot unusual activity.
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Total Trades (Transactions)</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                The number of individual trade executions. More trades indicate active trading and tighter
                spreads. Fewer trades suggest illiquid market conditions.
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Advancers vs Decliners</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                The count of stocks with positive vs negative price changes. <strong>Advancers &gt; Decliners</strong>
                suggests bullish sentiment. <strong>Decliners &gt; Advancers</strong> suggests bearish sentiment.
                This is a measure of market breadth.
              </CardContent>
            </Card>
          </div>

          <Alert>
            <Info className="h-4 w-4" />
            <AlertTitle>Market Breadth Interpretation</AlertTitle>
            <AlertDescription className="text-sm">
              <strong>Strong bull market</strong>: 70%+ advancers, high volume, expanding value<br />
              <strong>Weak rally</strong>: Market index up but &lt;50% advancers (divergence warning)<br />
              <strong>Strong bear market</strong>: 70%+ decliners, panic selling, high volume<br />
              <strong>Choppy market</strong>: 45-55% advancers, low conviction, range-bound
            </AlertDescription>
          </Alert>
        </div>

        {/* Code Example */}
        <InteractiveDemo
          title="Navigate to Market Overview Programmatically"
          description="Use JavaScript to open Market Overview (useful for custom dashboards)"
        >
          <CodeExample
            language="typescript"
            code={`// Open Market Overview in new tab
window.open('/market-overview', '_blank')

// Navigate to Market Overview (same tab)
window.location.href = '/market-overview'

// With specific date (via query param - if supported)
window.location.href = '/market-overview?date=2025-10-10'

// Using Next.js router (if available in client component)
import { useRouter } from 'next/navigation'

const router = useRouter()
router.push('/market-overview')`}
          />
        </InteractiveDemo>

        {/* Practice Exercise */}
        <Card className="border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/30">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Sparkles className="h-5 w-5 text-blue-600 dark:text-blue-400" />
              Practice Exercise
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm text-muted-foreground">
              Try this exercise to master Market Overview:
            </p>
            <ol className="space-y-2 text-sm text-muted-foreground">
              <li>1. Open <a href="/market-overview" target="_blank" className="text-blue-600 dark:text-blue-400 hover:underline">Market Overview</a></li>
              <li>2. Select today's date (or latest available)</li>
              <li>3. Identify the 3 largest rectangles (highest traded value)</li>
              <li>4. Identify the darkest green rectangle (strongest gainer)</li>
              <li>5. Identify the darkest red rectangle (strongest loser)</li>
              <li>6. Click on the strongest gainer to analyze it in detail</li>
              <li>7. Use arrow keys to navigate to the previous trading day</li>
              <li>8. Compare market breadth (advancers vs decliners) between the two days</li>
            </ol>
            <p className="text-sm font-medium text-blue-600 dark:text-blue-400">
              🎯 Goal: Develop a daily habit of scanning the market before analyzing individual stocks.
            </p>
          </CardContent>
        </Card>

        {/* Next Steps */}
        <div className="space-y-3">
          <h2 className="text-xl font-semibold">Next Steps</h2>
          <div className="flex flex-wrap gap-3">
            <Button variant="outline" asChild>
              <a href="/guide?section=strategy">
                Trading Strategies →
              </a>
            </Button>
            <Button variant="outline" asChild>
              <a href="/guide?section=quick-reference">
                Quick Reference →
              </a>
            </Button>
            <Button variant="outline" asChild>
              <a href="/market-overview" target="_blank">
                <ExternalLink className="h-4 w-4 mr-2" />
                Try Market Overview Live
              </a>
            </Button>
          </div>
        </div>
      </div>
    </GuideSection>
  )
}
