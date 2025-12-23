'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  TrendingUp,
  Calculator,
  CheckCircle2,
  AlertCircle,
  ChevronRight,
  Gauge,
  Activity,
  Calendar,
  Info,
  TrendingDown,
  ArrowRight,
  Target,
  Settings,
  Zap,
  Scale
} from 'lucide-react'
import { InteractiveDemo } from '../components/InteractiveDemo'
import { LiquidityCalculatorDemo } from '@/components/guide/demos/LiquidityCalculatorDemo'

interface LiquidityCalculationsSectionProps {
  onNavigate: (sectionId: string) => void
}

/**
 * Liquidity Calculations Section - Deep dive into ISX Hybrid Liquidity Scoring
 * Explains step-by-step how liquidity scores are calculated with pros/cons analysis
 */
export function LiquidityCalculationsSection({ onNavigate }: LiquidityCalculationsSectionProps) {
  const metrics = [
    {
      id: 'illiq',
      name: 'Price Impact (ILLIQ)',
      icon: Gauge,
      color: 'bg-red-500',
      weight: '40%',
      description: 'Measures how much the stock price moves when you trade',
      basedOn: 'Amihud (2002) Illiquidity Measure',
      shortDesc: 'Price movement per IQD traded'
    },
    {
      id: 'value',
      name: 'Trading Value',
      icon: Activity,
      color: 'bg-blue-500',
      weight: '35%',
      description: 'Average daily trading value in Iraqi Dinars',
      basedOn: '60-day Simple Moving Average (SMA) including non-trading days',
      shortDesc: 'Market depth indicator'
    },
    {
      id: 'continuity',
      name: 'Trading Continuity',
      icon: Calendar,
      color: 'bg-green-500',
      weight: '25%',
      description: 'How consistently the stock trades over time',
      basedOn: 'Percentage of days with active trading in 60-day window',
      shortDesc: 'Trading reliability measure'
    }
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-primary/10">
            <Calculator className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold">Liquidity Score Calculations</h2>
            <p className="text-muted-foreground text-lg">
              Understanding the ISX Hybrid Liquidity Scoring System step-by-step
            </p>
          </div>
        </div>
      </div>

      {/* Introduction */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">What is Liquidity?</h3>
        <Card className="border-primary/50 bg-primary/5">
          <CardContent className="pt-6">
            <p className="text-lg mb-4">
              <strong>Liquidity</strong> measures how quickly and easily you can buy or sell a stock without significantly affecting its price.
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
              <div className="space-y-2">
                <p className="font-semibold text-green-600 dark:text-green-400 flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4" />
                  High Liquidity Means
                </p>
                <ul className="space-y-1 text-sm ml-6">
                  <li>Quick buying/selling execution</li>
                  <li>Minimal price impact from trades</li>
                  <li>Narrow bid-ask spreads</li>
                  <li>Consistent daily trading</li>
                  <li>Many buyers and sellers</li>
                </ul>
              </div>
              <div className="space-y-2">
                <p className="font-semibold text-red-600 dark:text-red-400 flex items-center gap-2">
                  <AlertCircle className="h-4 w-4" />
                  Low Liquidity Means
                </p>
                <ul className="space-y-1 text-sm ml-6">
                  <li>Difficult to trade without moving price</li>
                  <li>Wide bid-ask spreads</li>
                  <li>Irregular trading gaps</li>
                  <li>Few market participants</li>
                  <li>Risk of being "trapped" in position</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* The 3-Metric System Overview */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">The 3-Metric Hybrid System</h3>
        <p className="text-muted-foreground">
          ISX Pulse evaluates liquidity across three complementary dimensions, each capturing a unique aspect of market tradability:
        </p>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {metrics.map((metric) => {
            const MetricIcon = metric.icon
            return (
              <Card key={metric.id} className="relative overflow-hidden">
                <div className={`absolute top-0 left-0 w-full h-1 ${metric.color}`} />
                <CardHeader>
                  <div className="flex items-start gap-3">
                    <div className={`p-2 rounded-lg ${metric.color} bg-opacity-10 shrink-0`}>
                      <MetricIcon className="h-5 w-5" />
                    </div>
                    <div className="flex-1">
                      <CardTitle className="text-base">{metric.name}</CardTitle>
                      <Badge variant="outline" className="mt-2">{metric.weight}</Badge>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-2">
                  <p className="text-sm text-muted-foreground">{metric.description}</p>
                  <p className="text-xs font-mono bg-muted p-2 rounded">{metric.shortDesc}</p>
                </CardContent>
              </Card>
            )
          })}
        </div>

      </section>

      {/* Why 3 Dimensions Not 1 */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Why We Use 3 Metrics Instead of 1</h3>

        <Alert className="border-primary/50 bg-primary/5">
          <Info className="h-4 w-4" />
          <AlertTitle>Liquidity is Fundamentally Multidimensional</AlertTitle>
          <AlertDescription>
            A single metric cannot capture all aspects of market tradability. According to Harris (2003) in "Trading and Exchanges", liquid markets exhibit three distinct properties: <strong>Tightness</strong> (low transaction costs), <strong>Depth</strong> (high order book capacity), and <strong>Resiliency</strong> (fast price recovery). Each dimension requires separate measurement.
          </AlertDescription>
        </Alert>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="border-red-500/30">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Gauge className="h-5 w-5 text-red-500" />
                <CardTitle className="text-base">Tightness</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="text-sm font-semibold">Transaction Costs</p>
              <p className="text-xs text-muted-foreground">
                How much does it cost to execute a trade immediately? Measured by bid-ask spreads and price impact.
              </p>
              <div className="pt-2 border-t">
                <p className="text-xs font-semibold text-red-600 dark:text-red-400">Captured by: ILLIQ</p>
                <p className="text-xs text-muted-foreground">Price movement per IQD traded</p>
              </div>
            </CardContent>
          </Card>

          <Card className="border-blue-500/30">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Activity className="h-5 w-5 text-blue-500" />
                <CardTitle className="text-base">Depth</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="text-sm font-semibold">Market Capacity</p>
              <p className="text-xs text-muted-foreground">
                How large a trade can the market absorb without significant price change? Measured by order book size.
              </p>
              <div className="pt-2 border-t">
                <p className="text-xs font-semibold text-blue-600 dark:text-blue-400">Captured by: Trading Value</p>
                <p className="text-xs text-muted-foreground">Average daily trading capacity</p>
              </div>
            </CardContent>
          </Card>

          <Card className="border-green-500/30">
            <CardHeader>
              <div className="flex items-center gap-2">
                <Calendar className="h-5 w-5 text-green-500" />
                <CardTitle className="text-base">Resiliency</CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-2">
              <p className="text-sm font-semibold">Temporal Reliability</p>
              <p className="text-xs text-muted-foreground">
                How consistently is liquidity available? Can you trade whenever you need to?
              </p>
              <div className="pt-2 border-t">
                <p className="text-xs font-semibold text-green-600 dark:text-green-400">Captured by: Continuity</p>
                <p className="text-xs text-muted-foreground">Percentage of trading days</p>
              </div>
            </CardContent>
          </Card>
        </div>

        <Card className="border-destructive/50 bg-destructive/5">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-destructive" />
              Why Single Metrics Fail
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="space-y-2">
              <p className="text-sm font-semibold">Volume-Only Approach ❌</p>
              <p className="text-xs text-muted-foreground">
                <strong>Problem:</strong> A stock can have high trading volume but terrible spreads. Example: 100M IQD daily volume with 10% bid-ask spread = extremely expensive to trade despite "high liquidity".
              </p>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-semibold">Price Impact-Only Approach ❌</p>
              <p className="text-xs text-muted-foreground">
                <strong>Problem:</strong> Low price impact on tiny trades doesn't mean the stock is liquid. Example: Moving price by 0.1% on 100K IQD trades, but the market disappears when you try 10M IQD.
              </p>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-semibold">Continuity-Only Approach ❌</p>
              <p className="text-xs text-muted-foreground">
                <strong>Problem:</strong> Trading every day doesn't mean capacity exists. Example: Trading 1M IQD daily on 60 consecutive days = 100% continuity but completely inadequate for serious trading.
              </p>
            </div>
          </CardContent>
        </Card>

        <Alert className="border-green-500/50 bg-green-500/5">
          <CheckCircle2 className="h-4 w-4 text-green-500" />
          <AlertTitle>The Hybrid Solution</AlertTitle>
          <AlertDescription>
            By combining all three dimensions, we capture the complete picture: A stock must have <strong>low transaction costs</strong> (ILLIQ), <strong>sufficient capacity</strong> (Trading Value), and <strong>reliable availability</strong> (Continuity) to be truly liquid. Weakness in any dimension reveals itself in the hybrid score.
          </AlertDescription>
        </Alert>
      </section>

      {/* ILLIQ Theory Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Gauge className="h-6 w-6 text-red-500" />
          <h3 className="text-2xl font-semibold">Metric 1: Price Impact (ILLIQ) - 40% Weight</h3>
        </div>

        {/* Academic Foundation */}
        <Card className="border-red-500/50 bg-red-500/5">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Target className="h-5 w-5 text-red-500" />
              Academic Foundation: Amihud (2002)
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm">
              The <strong>ILLIQ</strong> (Illiquidity) measure was introduced by Yakov Amihud in his seminal 2002 paper <em>"Illiquidity and Stock Returns: Cross-Section and Time-Series Effects"</em> published in the Journal of Financial Markets. It measures <strong>price impact</strong> - how much a stock's price moves in response to a given amount of trading value.
            </p>

            <div className="p-4 bg-muted rounded-lg border-2 border-red-500/30">
              <p className="text-sm font-mono text-center">
                ILLIQ = |Return| / Trading Value
              </p>
              <p className="text-xs text-muted-foreground text-center mt-2">
                Price change (%) per million IQD traded
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
              <div className="space-y-2">
                <p className="text-sm font-semibold text-red-600 dark:text-red-400">Higher ILLIQ = Less Liquid</p>
                <p className="text-xs text-muted-foreground">
                  A high ILLIQ value means that small trades cause large price movements - indicating difficulty in executing orders without market impact.
                </p>
              </div>
              <div className="space-y-2">
                <p className="text-sm font-semibold text-green-600 dark:text-green-400">Lower ILLIQ = More Liquid</p>
                <p className="text-xs text-muted-foreground">
                  A low ILLIQ value means that even large trades have minimal price impact - ideal for institutional trading.
                </p>
              </div>
            </div>

            <Alert className="border-blue-500/50 bg-blue-500/5 mt-4">
              <Info className="h-4 w-4" />
              <AlertTitle>Why Amihud ILLIQ?</AlertTitle>
              <AlertDescription className="text-xs">
                Amihud's measure is widely used in academic research and industry because it: (1) requires only publicly available data (price and volume), (2) captures the fundamental cost of trading (price impact), (3) works well across different market conditions, and (4) has been empirically validated to predict future returns and liquidity crises.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>

        {/* Market Microstructure Theory */}
        <Card className="border-red-500/50">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Activity className="h-5 w-5 text-red-500" />
              Market Microstructure: Why Price Impact Matters
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              Price impact is the most fundamental dimension of liquidity because it directly reflects the <strong>true cost of trading</strong>. When you place a buy order, you push the price up; when you sell, you push it down. The magnitude of this movement is determined by three market microstructure forces:
            </p>

            <div className="space-y-3">
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">1. Information Asymmetry</p>
                <p className="text-xs text-muted-foreground">
                  Market makers widen spreads when they fear you have superior information. If your 10M IQD order moves the price significantly, it signals that market makers perceive <strong>information risk</strong> - they worry you know something they don't (e.g., upcoming earnings, insider news). This is why thinly-traded ISX stocks often have high ILLIQ scores.
                </p>
              </div>

              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">2. Inventory Risk</p>
                <p className="text-xs text-muted-foreground">
                  Market makers must hold inventory to facilitate trades. When you buy, they sell from their inventory and become <strong>short</strong> - exposing them to risk if the price rises. To compensate for this risk, they demand a higher price (price impact). Stocks with thin order books (common in emerging markets) exhibit high inventory risk and thus high ILLIQ.
                </p>
              </div>

              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">3. Adverse Selection</p>
                <p className="text-xs text-muted-foreground">
                  Market makers know that <strong>informed traders</strong> (insiders, analysts) are more likely to trade when they have valuable information. To protect themselves from being "picked off" by informed traders, market makers widen spreads. This phenomenon is especially pronounced in less transparent markets like ISX, where information dissemination is uneven.
                </p>
              </div>
            </div>

            <div className="p-4 bg-destructive/5 border-l-4 border-destructive rounded-lg mt-4">
              <p className="text-sm font-semibold mb-2">Permanent vs Temporary Impact</p>
              <p className="text-xs text-muted-foreground">
                Price impact has two components: <strong>Temporary impact</strong> recovers after your trade (bid-ask bounce), while <strong>permanent impact</strong> persists (reflects new information). The ILLIQ measure captures both. For ISX stocks, permanent impact dominates due to low market depth - your trade genuinely moves the equilibrium price.
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Real-World Trading Impact */}
        <Card className="border-red-500/50">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <TrendingDown className="h-5 w-5 text-red-500" />
              Real-World Impact on Your Trades
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              Let's make this concrete with examples from actual ISX trading scenarios. Price impact <strong>compounds</strong> on both entry and exit, significantly affecting your realized returns:
            </p>

            <div className="space-y-3">
              <div className="p-4 bg-green-500/10 border border-green-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                  <p className="text-sm font-semibold">Example 1: Highly Liquid Stock (ILLIQ = 0.05)</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Scenario:</strong> You buy 50M IQD of a blue-chip bank stock</p>
                  <p className="text-green-600 dark:text-green-400">
                    <strong>Entry Impact:</strong> Price moves up 0.25% (50M × 0.05 = 0.25%)
                  </p>
                  <p className="text-green-600 dark:text-green-400">
                    <strong>Exit Impact:</strong> Price moves down 0.25% when you sell
                  </p>
                  <p className="text-green-600 dark:text-green-400 font-semibold">
                    <strong>Total Transaction Cost:</strong> ~0.5% (manageable for day trading)
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock is suitable for active trading strategies. You can enter and exit positions without significantly eroding profits.
                  </p>
                </div>
              </div>

              <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4 text-red-500" />
                  <p className="text-sm font-semibold">Example 2: Illiquid Stock (ILLIQ = 5.0)</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Scenario:</strong> You buy 10M IQD of a thinly-traded industrial stock</p>
                  <p className="text-red-600 dark:text-red-400">
                    <strong>Entry Impact:</strong> Price moves up 50% (10M × 5.0 = 50%)
                  </p>
                  <p className="text-red-600 dark:text-red-400">
                    <strong>Exit Impact:</strong> Price moves down 50% when you sell
                  </p>
                  <p className="text-red-600 dark:text-red-400 font-semibold">
                    <strong>Total Transaction Cost:</strong> ~100% (position is effectively underwater immediately!)
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock is extremely dangerous for trading. The price impact alone guarantees a loss unless you hold long enough for fundamental price appreciation to exceed transaction costs.
                  </p>
                </div>
              </div>

              <div className="p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <Info className="h-4 w-4 text-yellow-500" />
                  <p className="text-sm font-semibold">Example 3: Moderately Liquid Stock (ILLIQ = 0.8)</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Scenario:</strong> You buy 20M IQD of a mid-cap services company</p>
                  <p className="text-yellow-600 dark:text-yellow-400">
                    <strong>Entry Impact:</strong> Price moves up 16% (20M × 0.8 = 16%)
                  </p>
                  <p className="text-yellow-600 dark:text-yellow-400">
                    <strong>Exit Impact:</strong> Price moves down 16% when you sell
                  </p>
                  <p className="text-yellow-600 dark:text-yellow-400 font-semibold">
                    <strong>Total Transaction Cost:</strong> ~32% (requires significant price appreciation)
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    Suitable for swing trading or long-term investment, but not day trading. You need the stock to appreciate at least 32% just to break even on transaction costs.
                  </p>
                </div>
              </div>
            </div>

            <Alert className="border-destructive/50 bg-destructive/5 mt-4">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Critical Trading Implication</AlertTitle>
              <AlertDescription className="text-xs">
                <strong>Never day trade stocks with ILLIQ {'>'} 1.0.</strong> The round-trip transaction cost (entry + exit) makes profitable day trading mathematically impossible unless you have perfect timing. For stocks with ILLIQ {'>'} 2.0, even swing trading becomes challenging. This is why ILLIQ receives the <strong>highest weight (40%)</strong> in our hybrid liquidity score.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </section>

      {/* Metric 1: ILLIQ Calculation Steps */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Calculator className="h-6 w-6 text-red-500" />
          <h3 className="text-xl font-semibold">ILLIQ Calculation Steps</h3>
        </div>

        <Card className="border-red-500/50">
          <CardHeader>
            <CardTitle>Step-by-Step Calculation</CardTitle>
            <CardDescription>Based on Amihud (2002) Illiquidity Measure</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  1
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Calculate Absolute Daily Return</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    AbsReturn = |ClosePrice - PrevClosePrice| / PrevClosePrice
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Measures the percentage price change without regard to direction
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  2
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Get Trading Value in Millions IQD</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    ValueMillion = TradingValue / 1,000,000
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Normalize to millions for better numerical stability
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  3
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Compute Raw ILLIQ</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    ILLIQ = AbsReturn / (ValueMillion + 0.000001)
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Small epsilon (1e-6) prevents division by zero for low-value stocks
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  4
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Apply Winsorization (Cap Outliers)</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    if ILLIQ {'>'} 1000: ILLIQ = 1000
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Caps extreme values to prevent outliers from distorting scaling
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  5
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Cross-Sectional Scaling</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    ILLIQ_Scaled = PiecewiseLinearScale(ILLIQ, all_stocks) → [0, 100]
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Normalize relative to all stocks on the same date for fair comparison
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  6
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Invert Scale (Higher Score = More Liquid)</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    ILLIQ_Final = 100 - ILLIQ_Scaled
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Inverted so higher scores indicate better liquidity (less price impact)
                  </p>
                </div>
              </div>
            </div>

            {/* Pros and Cons */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t">
              <div>
                <p className="font-semibold text-green-600 dark:text-green-400 flex items-center gap-2 mb-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Pros
                </p>
                <ul className="space-y-1 text-sm">
                  <li>✓ Based on proven Amihud (2002) academic research</li>
                  <li>✓ Captures actual price impact from real trades</li>
                  <li>✓ Robust to outliers through winsorization</li>
                  <li>✓ Works well for emerging markets like ISX</li>
                  <li>✓ Direct relationship to trading costs</li>
                </ul>
              </div>
              <div>
                <p className="font-semibold text-red-600 dark:text-red-400 flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4" />
                  Cons
                </p>
                <ul className="space-y-1 text-sm">
                  <li>✗ Requires sufficient trading data (min 3 days)</li>
                  <li>✗ Can be noisy for very low-volume stocks</li>
                  <li>✗ Doesn't capture intraday price dynamics</li>
                  <li>✗ Affected by price level (partly mitigated by penalties)</li>
                  <li>✗ May lag sudden liquidity changes</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Trading Value Theory Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Activity className="h-6 w-6 text-blue-500" />
          <h3 className="text-2xl font-semibold">Metric 2: Trading Value - 35% Weight</h3>
        </div>

        {/* Academic Foundation - Kyle (1985) */}
        <Card className="border-blue-500/50 bg-blue-500/5">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Target className="h-5 w-5 text-blue-500" />
              Academic Foundation: Kyle (1985) Market Depth Theory
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm">
              The <strong>Trading Value</strong> metric is rooted in Albert Kyle's 1985 paper <em>"Continuous Auctions and Insider Trading"</em>, which formalized the concept of <strong>market depth</strong> - the ability of a market to absorb large orders without significant price changes. Market depth measures the <strong>order book capacity</strong>: how much capital is waiting at various price levels to provide liquidity.
            </p>

            <div className="p-4 bg-muted rounded-lg border-2 border-blue-500/30">
              <p className="text-sm font-mono text-center">
                Depth = Trading Value (Capacity to Absorb Orders)
              </p>
              <p className="text-xs text-muted-foreground text-center mt-2">
                Higher trading value = More buyers/sellers = Deeper market
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
              <div className="space-y-2">
                <p className="text-sm font-semibold text-blue-600 dark:text-blue-400">Deep Market (High Value)</p>
                <p className="text-xs text-muted-foreground">
                  Can absorb 50M IQD orders without moving the price significantly. Institutional investors can enter/exit positions without disrupting the market.
                </p>
              </div>
              <div className="space-y-2">
                <p className="text-sm font-semibold text-red-600 dark:text-red-400">Shallow Market (Low Value)</p>
                <p className="text-xs text-muted-foreground">
                  Even 5M IQD order exhausts available liquidity. The order book is thin - few buy/sell orders waiting at each price level.
                </p>
              </div>
            </div>

            <Alert className="border-blue-500/50 bg-blue-500/5 mt-4">
              <Info className="h-4 w-4" />
              <AlertTitle>Why Trading Value Matters for ISX</AlertTitle>
              <AlertDescription className="text-xs">
                In emerging markets like ISX, trading value is often more important than price impact for position sizing. A stock might have low ILLIQ (good price impact) but only trades 1M IQD daily - making it impossible to build or exit meaningful positions. Trading value tells you the <strong>maximum position size</strong> you can realistically manage.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>

        {/* Market Microstructure: Why Depth Matters */}
        <Card className="border-blue-500/50">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <TrendingUp className="h-5 w-5 text-blue-500" />
              Why Market Depth Matters: The Order Book Perspective
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              Market depth is fundamentally about <strong>capital availability</strong>. When you place a market order, you're relying on existing limit orders in the order book to fill your trade. The trading value metric measures the average capital flow through the market, which proxies for order book depth.
            </p>

            <div className="space-y-3">
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">1. Institutional Interest Indicator</p>
                <p className="text-xs text-muted-foreground">
                  High trading values (100M+ IQD daily) indicate institutional participation. Banks, investment funds, and large traders are active in the market, providing continuous liquidity. This creates a <strong>stable order book</strong> with many resting orders at various price levels.
                </p>
              </div>

              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">2. Order Book Density</p>
                <p className="text-xs text-muted-foreground">
                  Stocks with high trading values typically have <strong>dense order books</strong> - many buy/sell orders clustered around the current price. This means your order is less likely to "walk through" multiple price levels, reducing slippage and price impact.
                </p>
              </div>

              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">3. Market Resilience</p>
                <p className="text-xs text-muted-foreground">
                  Deep markets recover quickly from temporary imbalances. If a large sell order temporarily pushes the price down, sufficient buying interest exists to absorb the shares and restore equilibrium. Shallow markets lack this resilience - a single large order can cause lasting price dislocations.
                </p>
              </div>
            </div>

            <div className="p-4 bg-blue-500/5 border-l-4 border-blue-500 rounded-lg mt-4">
              <p className="text-sm font-semibold mb-2">The SMA 60 with Zeros Innovation</p>
              <p className="text-xs text-muted-foreground">
                Unlike traditional moving averages that exclude non-trading days, our approach divides by 60 calendar days, including zeros for non-trading days. This innovation naturally incorporates <strong>trading continuity into the depth metric</strong>: A stock trading 30M IQD for only 30 days out of 60 gets an average of 15M IQD, accurately reflecting that liquidity is only available half the time. This is more conservative and realistic than traditional SMAs.
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Real-World Trading Capacity */}
        <Card className="border-blue-500/50">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Target className="h-5 w-5 text-blue-500" />
              Real-World Impact: Position Sizing and Exit Planning
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              Trading value directly constrains your <strong>maximum practical position size</strong>. Professional traders follow the "10% rule": Never hold more than 10% of average daily trading value to ensure you can exit within a reasonable timeframe without dominating the market.
            </p>

            <div className="space-y-3">
              <div className="p-4 bg-green-500/10 border border-green-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                  <p className="text-sm font-semibold">Example 1: Deep Market Stock</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Stock:</strong> Major bank with 200M IQD average daily trading value</p>
                  <p className="text-green-600 dark:text-green-400">
                    <strong>Safe Position Size:</strong> 20M IQD (10% of daily volume)
                  </p>
                  <p className="text-green-600 dark:text-green-400">
                    <strong>Exit Timeline:</strong> Can exit full position in 1-2 days without market impact
                  </p>
                  <p className="text-green-600 dark:text-green-400 font-semibold">
                    <strong>Suitability:</strong> Institutional trading, large positions, active strategies
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock has sufficient depth to accommodate serious capital. You can confidently build 20-30M IQD positions and exit cleanly when needed.
                  </p>
                </div>
              </div>

              <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4 text-red-500" />
                  <p className="text-sm font-semibold">Example 2: Shallow Market Stock</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Stock:</strong> Small industrial company with 2M IQD average daily trading value</p>
                  <p className="text-red-600 dark:text-red-400">
                    <strong>Safe Position Size:</strong> 200K IQD (10% of daily volume)
                  </p>
                  <p className="text-red-600 dark:text-red-400">
                    <strong>Exit Timeline:</strong> 10-15 days to exit 2M IQD position without dominating
                  </p>
                  <p className="text-red-600 dark:text-red-400 font-semibold">
                    <strong>Suitability:</strong> Micro-cap speculation, long-term holding only
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock has insufficient depth for active trading. Any position over 500K IQD puts you at risk of becoming "trapped" - unable to exit quickly when needed. Suitable only for patient, long-term investors willing to wait weeks or months to exit.
                  </p>
                </div>
              </div>

              <div className="p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <Info className="h-4 w-4 text-yellow-500" />
                  <p className="text-sm font-semibold">Example 3: Moderate Depth Stock</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Stock:</strong> Mid-cap services company with 25M IQD average daily trading value</p>
                  <p className="text-yellow-600 dark:text-yellow-400">
                    <strong>Safe Position Size:</strong> 2.5M IQD (10% of daily volume)
                  </p>
                  <p className="text-yellow-600 dark:text-yellow-400">
                    <strong>Exit Timeline:</strong> 2-3 days to exit full position cleanly
                  </p>
                  <p className="text-yellow-600 dark:text-yellow-400 font-semibold">
                    <strong>Suitability:</strong> Swing trading, medium-term positions
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock offers reasonable depth for retail and semi-professional traders. You can build 2-5M IQD positions and exit over several days. Not suitable for day trading (insufficient depth) but works well for swing trades lasting 1-4 weeks.
                  </p>
                </div>
              </div>
            </div>

            <Alert className="border-blue-500/50 bg-blue-500/5 mt-4">
              <Info className="h-4 w-4" />
              <AlertTitle>Position Sizing Rule for ISX</AlertTitle>
              <AlertDescription className="text-xs">
                <strong>Conservative Rule:</strong> Position size ≤ 5% of daily trading value for quick exits (1-2 days).<br/>
                <strong>Standard Rule:</strong> Position size ≤ 10% of daily trading value for normal exits (3-5 days).<br/>
                <strong>Aggressive Rule:</strong> Position size ≤ 20% of daily trading value (only for patient traders willing to wait 5-10+ days to exit).
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </section>

      {/* Metric 2: Trading Value Calculation Steps */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Calculator className="h-6 w-6 text-blue-500" />
          <h3 className="text-xl font-semibold">Trading Value Calculation Steps</h3>
        </div>

        <Card className="border-blue-500/50">
          <CardHeader>
            <CardTitle>Step-by-Step Calculation</CardTitle>
            <CardDescription>60-day Simple Moving Average (SMA) Including Non-Trading Days</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  1
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Sum All Trading Values in 60-Day Window</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    TotalValue = Σ (Value_i) for i = 1 to 60 days
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    <strong>Important:</strong> Non-trading days contribute 0 to the sum (not excluded!)
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  2
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Divide by 60 (Not by Trading Days!)</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    AvgValue = TotalValue / 60
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    This is the key difference from traditional moving averages - we always divide by 60, regardless of actual trading days
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  3
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Cross-Sectional Scaling</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    Value_Scaled = PiecewiseLinearScale(AvgValue, all_stocks) → [0, 100]
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Compare to all stocks on same date using piecewise linear function
                  </p>
                </div>
              </div>
            </div>

            <Alert className="border-blue-500/50 bg-blue-500/5">
              <Info className="h-4 w-4" />
              <AlertTitle>Why SMA 60 with Zeros?</AlertTitle>
              <AlertDescription>
                By including non-trading days as zeros and dividing by 60, the average <strong>naturally incorporates continuity</strong>. A stock trading 30M IQD for 30 days out of 60 gets an average of 15M IQD, automatically reflecting its sporadic nature. This is more accurate than traditional moving averages that only consider trading days.
              </AlertDescription>
            </Alert>

            {/* Pros and Cons */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t">
              <div>
                <p className="font-semibold text-green-600 dark:text-green-400 flex items-center gap-2 mb-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Pros
                </p>
                <ul className="space-y-1 text-sm">
                  <li>✓ Naturally incorporates continuity via zeros</li>
                  <li>✓ Reflects true market activity over time</li>
                  <li>✓ Simple and highly interpretable</li>
                  <li>✓ Directly measures market depth</li>
                  <li>✓ Stable and not easily manipulated</li>
                </ul>
              </div>
              <div>
                <p className="font-semibold text-red-600 dark:text-red-400 flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4" />
                  Cons
                </p>
                <ul className="space-y-1 text-sm">
                  <li>✗ Dilutes values for sporadic traders (by design)</li>
                  <li>✗ 60-day window may miss very recent shifts</li>
                  <li>✗ Doesn't weight recent data more heavily</li>
                  <li>✗ Equal treatment of all days (no recency bias)</li>
                  <li>✗ May undervalue improving stocks</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Continuity Theory Section */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Calendar className="h-6 w-6 text-green-500" />
          <h3 className="text-2xl font-semibold">Metric 3: Trading Continuity - 25% Weight</h3>
        </div>

        {/* Academic Foundation - Harris (2003) */}
        <Card className="border-green-500/50 bg-green-500/5">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Target className="h-5 w-5 text-green-500" />
              Academic Foundation: Harris (2003) Market Resiliency
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-sm">
              The <strong>Trading Continuity</strong> metric measures the third dimension of liquidity from Larry Harris's 2003 book <em>"Trading and Exchanges: Market Microstructure for Practitioners"</em> - <strong>resiliency</strong>. Harris defines resiliency as "how fast prices recover from a random, uninformative shock." In practical terms for retail traders, it answers: <strong>"Can I trade whenever I need to?"</strong>
            </p>

            <div className="p-4 bg-muted rounded-lg border-2 border-green-500/30">
              <p className="text-sm font-mono text-center">
                Continuity = Percentage of Days with Active Trading
              </p>
              <p className="text-xs text-muted-foreground text-center mt-2">
                Higher continuity = More reliable liquidity availability
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-4">
              <div className="space-y-2">
                <p className="text-sm font-semibold text-green-600 dark:text-green-400">High Continuity (75-100%)</p>
                <p className="text-xs text-muted-foreground">
                  Stock trades almost every day. You can execute entry/exit whenever you want. Stop-losses and exit strategies are reliable.
                </p>
              </div>
              <div className="space-y-2">
                <p className="text-sm font-semibold text-red-600 dark:text-red-400">Low Continuity ({'<'}50%)</p>
                <p className="text-xs text-muted-foreground">
                  Stock trades sporadically with frequent gaps. Exit timing is unpredictable - you may be forced to wait days or weeks to execute.
                </p>
              </div>
            </div>

            <Alert className="border-green-500/50 bg-green-500/5 mt-4">
              <Info className="h-4 w-4" />
              <AlertTitle>Why Continuity Matters for ISX</AlertTitle>
              <AlertDescription className="text-xs">
                In emerging markets, trading gaps are common due to low participation. A stock with 100% continuity trades every single market day - critical for risk management strategies like stop-losses. Without continuity, you're exposed to <strong>temporal risk</strong>: the danger that when you NEED to exit (stop-loss triggered, news event), the market is absent.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>

        {/* Market Microstructure: Temporal Reliability */}
        <Card className="border-green-500/50">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Calendar className="h-5 w-5 text-green-500" />
              Why Temporal Reliability Matters: Execution Certainty
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              Trading continuity measures <strong>execution certainty</strong> - the probability that liquidity will be available when you need it. Unlike price impact (which measures execution cost) and depth (which measures execution capacity), continuity measures <strong>execution timing risk</strong>.
            </p>

            <div className="space-y-3">
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">1. Stop-Loss Reliability</p>
                <p className="text-xs text-muted-foreground">
                  A stop-loss only works if the market is trading when your stop price is hit. If a stock has 50% continuity, your stop-loss has a 50% chance of executing immediately vs being delayed by days. This <strong>timing uncertainty</strong> dramatically increases risk, especially for leveraged positions.
                </p>
              </div>

              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">2. Exit Strategy Planning</p>
                <p className="text-xs text-muted-foreground">
                  High continuity allows precise exit planning. You can schedule exits ("I'll sell in 3 days") with confidence. Low continuity forces opportunistic exits - you must sell whenever the market appears, regardless of price, because tomorrow there may be no liquidity at all.
                </p>
              </div>

              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-semibold mb-1">3. News Reaction Capability</p>
                <p className="text-xs text-muted-foreground">
                  When news breaks (earnings, sector developments), you need to act quickly. Stocks with low continuity may go days without trading after news - forcing you to either accept stale prices or wait indefinitely. This is especially dangerous for negative news where you want to exit immediately.
                </p>
              </div>
            </div>

            <div className="p-4 bg-green-500/5 border-l-4 border-green-500 rounded-lg mt-4">
              <p className="text-sm font-semibold mb-2">Gap Penalty System</p>
              <p className="text-xs text-muted-foreground">
                Not all gaps are equal. A 1-day gap (market closed Friday, trades Monday) is normal. A 7-day gap means the stock went an entire week without trading - extremely risky. Our system applies <strong>escalating penalties</strong>: 5%/day for short gaps (1-2 days), 10%/day for medium (3-7 days), 20%/day for long gaps ({'>'}7 days). We also forgive 1 gap of ≤5 days per window to account for official market holidays.
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Real-World Risk Scenarios */}
        <Card className="border-green-500/50">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-green-500" />
              Real-World Impact: Timing Risk and Exit Traps
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              Trading continuity directly impacts your <strong>ability to execute risk management strategies</strong>. Even if a stock has good price impact and depth, sporadic trading can trap you in losing positions.
            </p>

            <div className="space-y-3">
              <div className="p-4 bg-green-500/10 border border-green-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <CheckCircle2 className="h-4 w-4 text-green-500" />
                  <p className="text-sm font-semibold">Example 1: High Continuity Stock (95%)</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Scenario:</strong> You buy a stock with tight stop-loss at -5%</p>
                  <p className="text-green-600 dark:text-green-400">
                    <strong>Outcome:</strong> Stock trades 57 out of 60 days - stop executes within 1 day when triggered
                  </p>
                  <p className="text-green-600 dark:text-green-400 font-semibold">
                    <strong>Risk:</strong> Minimal temporal risk - exit strategies work reliably
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock is suitable for active trading with stop-losses. Your risk management tools function as intended because liquidity is consistently available.
                  </p>
                </div>
              </div>

              <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4 text-red-500" />
                  <p className="text-sm font-semibold">Example 2: Low Continuity Stock (40%)</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Scenario:</strong> You buy same stock with -5% stop-loss</p>
                  <p className="text-red-600 dark:text-red-400">
                    <strong>Outcome:</strong> Stock trades only 24 out of 60 days - stop triggered but no trading for 5 days
                  </p>
                  <p className="text-red-600 dark:text-red-400">
                    <strong>During gap:</strong> Stock drops to -12% while you wait helplessly for market to reopen
                  </p>
                  <p className="text-red-600 dark:text-red-400 font-semibold">
                    <strong>Loss:</strong> -12% instead of planned -5% (140% larger loss due to gap!)
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    This stock is DANGEROUS for stop-loss strategies. Your -5% stop became a -12% loss because the market disappeared. Sporadic trading completely nullifies risk management.
                  </p>
                </div>
              </div>

              <div className="p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
                <div className="flex items-center gap-2 mb-2">
                  <Info className="h-4 w-4 text-yellow-500" />
                  <p className="text-sm font-semibold">Example 3: Moderate Continuity (70%)</p>
                </div>
                <div className="space-y-1 text-xs">
                  <p><strong>Scenario:</strong> News breaks - company announces partnership, you want to take profits</p>
                  <p className="text-yellow-600 dark:text-yellow-400">
                    <strong>Outcome:</strong> Stock trades 42 out of 60 days - 2-3 day wait to exit on average
                  </p>
                  <p className="text-yellow-600 dark:text-yellow-400 font-semibold">
                    <strong>Impact:</strong> Price may move against you during wait; opportunity cost of capital locked up
                  </p>
                  <p className="text-muted-foreground italic mt-2">
                    Acceptable for swing trading where 2-3 day exit delays are manageable, but not suitable for day trading or aggressive strategies requiring precise exit timing.
                  </p>
                </div>
              </div>
            </div>

            <Alert className="border-destructive/50 bg-destructive/5 mt-4">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Critical Risk Management Rule</AlertTitle>
              <AlertDescription className="text-xs">
                <strong>Never use stop-losses on stocks with continuity {'<'}70%.</strong> Your stops will not execute reliably, exposing you to runaway losses during gaps. For stocks {'<'}50% continuity, assume exit timing is completely unpredictable - only hold if prepared to wait weeks for liquidity. This is why continuity receives 25% weight despite seeming "less important" than price impact or depth.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </section>

      {/* Metric 3: Continuity Calculation Steps */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Calculator className="h-6 w-6 text-green-500" />
          <h3 className="text-xl font-semibold">Continuity Calculation Steps</h3>
        </div>

        <Card className="border-green-500/50">
          <CardHeader>
            <CardTitle>Step-by-Step Calculation</CardTitle>
            <CardDescription>Percentage of Days with Active Trading</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  1
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Count Trading Days</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    TradingDays = count where (TradingStatus = true AND (NumTrades {'>'} 0 OR Value {'>'} 0))
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    A valid trading day requires both TradingStatus flag AND actual trading activity
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  2
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Calculate Raw Continuity</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    Continuity = TradingDays / 60
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Simple percentage: 45 trading days out of 60 = 0.75 (75% continuity)
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  3
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Apply Non-Linear Transformation (Optional)</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    ContinuityNL = Continuity × (1 - Delta × GapPenalty)
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Gap penalties: short gaps (1-2 days) 5%/day, medium (3-7) 10%/day, long ({'>'}7) 20%/day
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  4
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Cross-Sectional Scaling</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    Continuity_Scaled = DirectPercentageMapping(ContinuityNL) → [0, 100]
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Direct mapping: 75% continuity = 75 score (with slight adjustments for gaps)
                  </p>
                </div>
              </div>
            </div>

            <Alert className="border-green-500/50 bg-green-500/5">
              <Info className="h-4 w-4" />
              <AlertTitle>Gap Forgiveness Policy</AlertTitle>
              <AlertDescription>
                Up to <strong>1 gap of ≤5 days</strong> is forgiven per 60-day window to account for legitimate market holidays (Eid, national holidays, etc.). This prevents unfair penalization of stocks during official market closures.
              </AlertDescription>
            </Alert>

            {/* Pros and Cons */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t">
              <div>
                <p className="font-semibold text-green-600 dark:text-green-400 flex items-center gap-2 mb-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Pros
                </p>
                <ul className="space-y-1 text-sm">
                  <li>✓ Direct measure of trading reliability</li>
                  <li>✓ Penalizes gaps appropriately (short vs long)</li>
                  <li>✓ Critical for exit strategy planning</li>
                  <li>✓ Easy to understand and interpret</li>
                  <li>✓ Forgives legitimate market holidays</li>
                </ul>
              </div>
              <div>
                <p className="font-semibold text-red-600 dark:text-red-400 flex items-center gap-2 mb-2">
                  <AlertCircle className="h-4 w-4" />
                  Cons
                </p>
                <ul className="space-y-1 text-sm">
                  <li>✗ Doesn't distinguish gap timing patterns</li>
                  <li>✗ May penalize legitimate suspensions</li>
                  <li>✗ Binary (trading vs non-trading)</li>
                  <li>✗ Doesn't consider trading intensity</li>
                  <li>✗ Equal weight to all trading days</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Final Hybrid Score Calculation */}
      <section className="space-y-4">
        <div className="flex items-center gap-3">
          <Target className="h-6 w-6 text-purple-500" />
          <h3 className="text-2xl font-semibold">Final Hybrid Score Calculation</h3>
        </div>

        <Card className="border-purple-500/50">
          <CardHeader>
            <CardTitle>Weighted Combination Formula</CardTitle>
            <CardDescription>Merging all three metrics into a single 0-100 score</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="p-4 bg-muted rounded-lg border-2 border-primary">
              <code className="text-lg font-mono block text-center">
                HybridScore = 0.40 × ILLIQ_Scaled + 0.35 × Value_Scaled + 0.25 × Continuity_Scaled
              </code>
            </div>

            <div className="space-y-3">
              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  1
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Apply Component Weights</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    Each scaled metric (0-100) is multiplied by its weight. ILLIQ has the highest weight (40%) as price impact is most critical for trading costs.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  2
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Apply Unified Penalty (if applicable)</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    ActivityScore = TradingDays / TotalDays
                    UnifiedPenalty = f(ActivityScore) → [1.0, MaxPenalty]
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    For SMA mode, activity-based adjustment only for extreme cases ({'<'}10% continuity)
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  3
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Sum Weighted Components</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    Add all three weighted components together. The result is automatically bounded [0, 100] due to input scaling.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 bg-muted rounded-lg">
                <div className="shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  4
                </div>
                <div className="flex-1">
                  <p className="font-semibold">Apply Safety Checks</p>
                  <code className="text-sm bg-background px-2 py-1 rounded block mt-1">
                    if (isNaN(score) || isInf(score)) score = 0
                    if (score {'<'} 0) score = 0
                    if (score {'>'} 100) score = 100
                  </code>
                  <p className="text-sm text-muted-foreground mt-1">
                    Final bounds checking to ensure valid output range
                  </p>
                </div>
              </div>
            </div>

            {/* Score Interpretation */}
            <div className="pt-4 border-t">
              <p className="font-semibold mb-3">Score Interpretation Guide:</p>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <div className="w-16 h-2 bg-gradient-to-r from-green-500 to-green-400 rounded" />
                  <span className="text-sm"><strong>80-100:</strong> Excellent liquidity - suitable for large institutional trades</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-16 h-2 bg-gradient-to-r from-blue-500 to-blue-400 rounded" />
                  <span className="text-sm"><strong>60-79:</strong> Good liquidity - suitable for active trading</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-16 h-2 bg-gradient-to-r from-yellow-500 to-yellow-400 rounded" />
                  <span className="text-sm"><strong>40-59:</strong> Moderate liquidity - small to medium positions</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-16 h-2 bg-gradient-to-r from-orange-500 to-orange-400 rounded" />
                  <span className="text-sm"><strong>20-39:</strong> Low liquidity - very small positions only</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-16 h-2 bg-gradient-to-r from-red-500 to-red-400 rounded" />
                  <span className="text-sm"><strong>0-19:</strong> Very low liquidity - avoid or extreme caution</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Interactive Calculator Demo */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Try It Yourself</h3>
        <p className="text-muted-foreground">
          Use the interactive calculator to see how liquidity scores are calculated for real ISX stocks:
        </p>

        <InteractiveDemo
          title="Liquidity Score Calculator"
          description="Step-by-step calculation breakdown with real market data"
          variant="interactive"
        >
          <LiquidityCalculatorDemo />
        </InteractiveDemo>
      </section>

      {/* NEW SECTION - Understanding Scoring Modes */}
      <section className="space-y-4" id="scoring-modes">
        <div className="flex items-center gap-3">
          <Settings className="h-6 w-6 text-purple-500" />
          <h3 className="text-2xl font-semibold">Understanding Scoring Modes</h3>
        </div>

        {/* Introduction */}
        <Alert className="border-purple-500/50 bg-purple-500/5">
          <Info className="h-4 w-4" />
          <AlertTitle>Why Three Different Modes?</AlertTitle>
          <AlertDescription>
            <p className="mb-2">Each scoring mode answers a different question about liquidity:</p>
            <ul className="space-y-1 text-sm">
              <li><strong>EMA (20-period):</strong> "What's the smoothed trend?" - Best for identifying improving/deteriorating liquidity over weeks</li>
              <li><strong>Latest (Recent):</strong> "What's happening RIGHT NOW?" - Best for day trading and reacting to current market conditions</li>
              <li><strong>Average:</strong> "What's the balanced long-term view?" - Best for conservative risk assessment and portfolio allocation</li>
            </ul>
          </AlertDescription>
        </Alert>

        <Card className="border-primary/50">
          <CardHeader>
            <CardTitle className="text-base">The Fundamental Trade-Off: Stability ↔ Reactivity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">
                There is no "best" mode - only the right mode for your trading style and timeframe. The three modes represent different positions on the stability-reactivity spectrum:
              </p>

              <div className="flex items-center gap-2 my-4">
                <div className="flex-1 h-2 bg-gradient-to-r from-blue-500 via-red-500 to-yellow-500 rounded" />
              </div>

              <div className="grid grid-cols-3 gap-2 text-center text-xs">
                <div>
                  <p className="font-semibold text-blue-600 dark:text-blue-400">EMA</p>
                  <p className="text-muted-foreground">Stable, Smooth</p>
                </div>
                <div>
                  <p className="font-semibold text-red-600 dark:text-red-400">Latest</p>
                  <p className="text-muted-foreground">Reactive, Volatile</p>
                </div>
                <div>
                  <p className="font-semibold text-yellow-600 dark:text-yellow-400">Average</p>
                  <p className="text-muted-foreground">Balanced, Conservative</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Mode 1: EMA (20-period) */}
        <Card className="border-blue-500/50 bg-blue-500/5">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-blue-500/20">
                <TrendingUp className="h-6 w-6 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <CardTitle>Mode 1: EMA (20-period) - The Trend Identifier</CardTitle>
                <CardDescription>Exponential Moving Average for smooth trend detection</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* What It Is */}
            <div className="space-y-2">
              <p className="font-semibold text-sm">What It Is:</p>
              <div className="p-3 bg-background rounded-lg border">
                <p className="text-sm text-muted-foreground mb-2">
                  20-period Exponential Moving Average that weights recent days more heavily than older days. Recent data gets ~2x the weight of 20-day-old data.
                </p>
                <code className="text-xs bg-muted px-2 py-1 rounded block">
                  EMA_today = (Score_today × 0.095) + (EMA_yesterday × 0.905)
                </code>
                <p className="text-xs text-muted-foreground mt-1">
                  Smoothing factor: α = 2/(20+1) ≈ 0.095
                </p>
              </div>
            </div>

            {/* When to Use */}
            <div className="space-y-2">
              <p className="font-semibold text-sm text-green-600 dark:text-green-400 flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4" />
                When to Use:
              </p>
              <ul className="space-y-1 text-sm text-muted-foreground ml-6">
                <li>✓ Swing trading (positions held 1-4 weeks)</li>
                <li>✓ Identifying liquidity trends (improving vs deteriorating)</li>
                <li>✓ Building positions gradually over days/weeks</li>
                <li>✓ Strategic decision-making for medium-term holds</li>
                <li>✓ When you want to filter out daily noise</li>
                <li>✓ <strong>Default choice for most traders</strong></li>
              </ul>
            </div>

            {/* Pros and Cons */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2">
              <div className="space-y-2">
                <p className="font-semibold text-sm text-green-600 dark:text-green-400 flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Pros
                </p>
                <ul className="text-xs space-y-1 text-muted-foreground">
                  <li>• Smooth, stable scores</li>
                  <li>• Detects trends early (~10 days)</li>
                  <li>• Filters daily volatility</li>
                  <li>• Industry-standard approach</li>
                  <li>• Good for strategic planning</li>
                </ul>
              </div>
              <div className="space-y-2">
                <p className="font-semibold text-sm text-red-600 dark:text-red-400 flex items-center gap-2">
                  <AlertCircle className="h-4 w-4" />
                  Cons
                </p>
                <ul className="text-xs space-y-1 text-muted-foreground">
                  <li>• Lags reality by ~10 days</li>
                  <li>• Misses sudden liquidity shocks</li>
                  <li>• Not for day trading</li>
                  <li>• Can show "good" when market dried up</li>
                  <li>• Slow to react to news</li>
                </ul>
              </div>
            </div>

            {/* Example */}
            <div className="p-4 bg-blue-500/10 border border-blue-500/30 rounded-lg">
              <p className="font-semibold text-sm mb-2 flex items-center gap-2">
                <Target className="h-4 w-4" />
                Real Example: Bank Stock Improving Liquidity
              </p>
              <div className="space-y-2 text-sm">
                <p className="text-muted-foreground">
                  <strong>Scenario:</strong> BBOB attracts institutional interest
                </p>
                <div className="space-y-1 text-xs pl-4 border-l-2 border-blue-500">
                  <p>• <strong>Day 1-5:</strong> Trading value jumps from 50M to 150M IQD daily</p>
                  <p>• <strong>Day 10:</strong> EMA score rises from 65 → 68 (smooth increase)</p>
                  <p>• <strong>Day 20:</strong> EMA score reaches 72 (confirmed trend)</p>
                </div>
                <p className="text-green-600 dark:text-green-400 font-semibold mt-2">
                  <strong>Decision:</strong> EMA confirms sustained improvement, safe to increase position from 5M to 10M IQD
                </p>
                <p className="text-xs text-muted-foreground italic mt-1">
                  EMA filtered out noise and confirmed the trend was real, not a temporary spike.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Mode 2: Latest (Recent) */}
        <Card className="border-red-500/50 bg-red-500/5">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-red-500/20">
                <Zap className="h-6 w-6 text-red-600 dark:text-red-400" />
              </div>
              <div>
                <CardTitle>Mode 2: Latest (Recent) - The Real-Time Detector</CardTitle>
                <CardDescription>Most recent day's metrics with zero lag</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* What It Is */}
            <div className="space-y-2">
              <p className="font-semibold text-sm">What It Is:</p>
              <div className="p-3 bg-background rounded-lg border">
                <p className="text-sm text-muted-foreground mb-2">
                  Uses only the most recent trading day's raw metrics. No smoothing, no averaging, no historical data. What you see is exactly what's happening TODAY.
                </p>
                <code className="text-xs bg-muted px-2 py-1 rounded block">
                  Score_Latest = Score_MostRecentTradingDay (no transformation)
                </code>
                <p className="text-xs text-muted-foreground mt-1">
                  Zero lag - reflects current market conditions instantly
                </p>
              </div>
            </div>

            {/* When to Use */}
            <div className="space-y-2">
              <p className="font-semibold text-sm text-green-600 dark:text-green-400 flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4" />
                When to Use:
              </p>
              <ul className="space-y-1 text-sm text-muted-foreground ml-6">
                <li>✓ Day trading and intraday decisions</li>
                <li>✓ Reacting to breaking news or earnings</li>
                <li>✓ Assessing immediate exit opportunities</li>
                <li>✓ Detecting sudden liquidity changes</li>
                <li>✓ When you need to know "Can I trade RIGHT NOW?"</li>
                <li>✓ Short-term tactical trading (hours to days)</li>
              </ul>
            </div>

            {/* Pros and Cons */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2">
              <div className="space-y-2">
                <p className="font-semibold text-sm text-green-600 dark:text-green-400 flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Pros
                </p>
                <ul className="text-xs space-y-1 text-muted-foreground">
                  <li>• Zero lag - instant reflection</li>
                  <li>• Catches liquidity shocks immediately</li>
                  <li>• Essential for day trading</li>
                  <li>• Shows true current depth</li>
                  <li>• No historical bias</li>
                </ul>
              </div>
              <div className="space-y-2">
                <p className="font-semibold text-sm text-red-600 dark:text-red-400 flex items-center gap-2">
                  <AlertCircle className="h-4 w-4" />
                  Cons
                </p>
                <ul className="text-xs space-y-1 text-muted-foreground">
                  <li>• Extremely volatile</li>
                  <li>• Prone to false signals</li>
                  <li>• Can mislead about trends</li>
                  <li>• Stressful to monitor</li>
                  <li>• Not for strategic planning</li>
                </ul>
              </div>
            </div>

            {/* Example */}
            <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
              <p className="font-semibold text-sm mb-2 flex items-center gap-2">
                <Target className="h-4 w-4" />
                Real Example: News-Driven Liquidity Spike
              </p>
              <div className="space-y-2 text-sm">
                <p className="text-muted-foreground">
                  <strong>Scenario:</strong> TASC announces major telecom contract
                </p>
                <div className="space-y-1 text-xs pl-4 border-l-2 border-red-500">
                  <p>• <strong>Before news:</strong> Latest score = 45 (thin market)</p>
                  <p>• <strong>Day of announcement:</strong> Trading value surges 10x</p>
                  <p>• <strong>Latest score:</strong> Jumps to 78 (excellent liquidity TODAY)</p>
                  <p>• <strong>EMA score:</strong> Still shows 47 (hasn't caught up yet)</p>
                </div>
                <p className="text-green-600 dark:text-green-400 font-semibold mt-2">
                  <strong>Decision:</strong> Latest mode alerts you liquidity is available RIGHT NOW to exit large position profitably
                </p>
                <p className="text-red-600 dark:text-red-400 text-xs mt-1">
                  <strong>Warning:</strong> Don't assume 78 score will persist - tomorrow may drop back to 45 when excitement fades
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Mode 3: Average */}
        <Card className="border-yellow-500/50 bg-yellow-500/5">
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-lg bg-yellow-500/20">
                <Scale className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
              </div>
              <div>
                <CardTitle>Mode 3: Average - The Balanced View</CardTitle>
                <CardDescription>Simple average of all 60 days with equal weighting</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* What It Is */}
            <div className="space-y-2">
              <p className="font-semibold text-sm">What It Is:</p>
              <div className="p-3 bg-background rounded-lg border">
                <p className="text-sm text-muted-foreground mb-2">
                  Simple arithmetic mean of all 60 days - each day gets equal weight. Most conservative and stable estimate, with no recency bias.
                </p>
                <code className="text-xs bg-muted px-2 py-1 rounded block">
                  Score_Average = (Score_day1 + Score_day2 + ... + Score_day60) / 60
                </code>
                <p className="text-xs text-muted-foreground mt-1">
                  All days treated equally - no preference for recent or old data
                </p>
              </div>
            </div>

            {/* When to Use */}
            <div className="space-y-2">
              <p className="font-semibold text-sm text-green-600 dark:text-green-400 flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4" />
                When to Use:
              </p>
              <ul className="space-y-1 text-sm text-muted-foreground ml-6">
                <li>✓ Long-term investment horizon (6+ months)</li>
                <li>✓ Portfolio risk assessment and allocation</li>
                <li>✓ Comparing stocks on equal footing</li>
                <li>✓ Conservative position sizing</li>
                <li>✓ When you want a "safe" baseline estimate</li>
                <li>✓ Annual portfolio reviews</li>
              </ul>
            </div>

            {/* Pros and Cons */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2">
              <div className="space-y-2">
                <p className="font-semibold text-sm text-green-600 dark:text-green-400 flex items-center gap-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Pros
                </p>
                <ul className="text-xs space-y-1 text-muted-foreground">
                  <li>• Most stable - changes very slowly</li>
                  <li>• No recency bias</li>
                  <li>• Great for conservative risk mgmt</li>
                  <li>• Simple to understand</li>
                  <li>• Unlikely to give false positives</li>
                </ul>
              </div>
              <div className="space-y-2">
                <p className="font-semibold text-sm text-red-600 dark:text-red-400 flex items-center gap-2">
                  <AlertCircle className="h-4 w-4" />
                  Cons
                </p>
                <ul className="text-xs space-y-1 text-muted-foreground">
                  <li>• Extremely slow (30+ days lag)</li>
                  <li>• Dilutes recent improvements</li>
                  <li>• Overly pessimistic for improving stocks</li>
                  <li>• Not useful for tactical trading</li>
                  <li>• May undervalue recent gains</li>
                </ul>
              </div>
            </div>

            {/* Example */}
            <div className="p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
              <p className="font-semibold text-sm mb-2 flex items-center gap-2">
                <Target className="h-4 w-4" />
                Real Example: Portfolio Risk Assessment
              </p>
              <div className="space-y-2 text-sm">
                <p className="text-muted-foreground">
                  <strong>Scenario:</strong> Building diversified ISX portfolio with 10 stocks
                </p>
                <div className="space-y-1 text-xs pl-4 border-l-2 border-yellow-500">
                  <p>• <strong>Stock A:</strong> EMA=75, Latest=82, Average=68</p>
                  <p>• <strong>Stock B:</strong> EMA=65, Latest=45, Average=70</p>
                </div>
                <p className="text-yellow-600 dark:text-yellow-400 font-semibold mt-2">
                  <strong>Analysis using Average mode:</strong>
                  <br/>• Stock A: Recently improved but historically moderate (68)
                  <br/>• Stock B: Recent bad day but historically solid (70)
                </p>
                <p className="text-green-600 dark:text-green-400 text-xs mt-2">
                  <strong>Decision:</strong> Average mode shows Stock B is actually more reliably liquid long-term, despite today's bad session
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Comparison Table */}
        <Card className="border-primary/50">
          <CardHeader>
            <CardTitle>Mode Comparison Table</CardTitle>
            <CardDescription>Side-by-side comparison of all three scoring modes</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr className="border-b-2">
                    <th className="text-left p-3 font-semibold">Feature</th>
                    <th className="text-center p-3 bg-blue-500/10 font-semibold text-blue-600 dark:text-blue-400">EMA (20)</th>
                    <th className="text-center p-3 bg-red-500/10 font-semibold text-red-600 dark:text-red-400">Latest</th>
                    <th className="text-center p-3 bg-yellow-500/10 font-semibold text-yellow-600 dark:text-yellow-400">Average</th>
                  </tr>
                </thead>
                <tbody className="text-xs">
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Lag Time</td>
                    <td className="text-center p-3 bg-blue-500/5">~10 days</td>
                    <td className="text-center p-3 bg-red-500/5">0 days</td>
                    <td className="text-center p-3 bg-yellow-500/5">~30 days</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Volatility</td>
                    <td className="text-center p-3 bg-blue-500/5">Low</td>
                    <td className="text-center p-3 bg-red-500/5">Very High</td>
                    <td className="text-center p-3 bg-yellow-500/5">Very Low</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Best For</td>
                    <td className="text-center p-3 bg-blue-500/5">Swing trading</td>
                    <td className="text-center p-3 bg-red-500/5">Day trading</td>
                    <td className="text-center p-3 bg-yellow-500/5">Long-term</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Update Speed</td>
                    <td className="text-center p-3 bg-blue-500/5">Gradual</td>
                    <td className="text-center p-3 bg-red-500/5">Instant</td>
                    <td className="text-center p-3 bg-yellow-500/5">Very slow</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Noise Filtering</td>
                    <td className="text-center p-3 bg-blue-500/5">Excellent</td>
                    <td className="text-center p-3 bg-red-500/5">None</td>
                    <td className="text-center p-3 bg-yellow-500/5">Excellent</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Trend Detection</td>
                    <td className="text-center p-3 bg-blue-500/5">Excellent</td>
                    <td className="text-center p-3 bg-red-500/5">Poor</td>
                    <td className="text-center p-3 bg-yellow-500/5">Poor</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-3 font-semibold">Current Accuracy</td>
                    <td className="text-center p-3 bg-blue-500/5">Moderate</td>
                    <td className="text-center p-3 bg-red-500/5">Perfect</td>
                    <td className="text-center p-3 bg-yellow-500/5">Poor</td>
                  </tr>
                  <tr>
                    <td className="p-3 font-semibold">Risk Level</td>
                    <td className="text-center p-3 bg-blue-500/5">Moderate</td>
                    <td className="text-center p-3 bg-red-500/5">High</td>
                    <td className="text-center p-3 bg-yellow-500/5">Low</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>

        {/* Decision Tree */}
        <Card className="border-primary/50 bg-primary/5">
          <CardHeader>
            <CardTitle>Which Mode Should I Use? (Decision Tree)</CardTitle>
            <CardDescription>Follow this flowchart to select the right mode for your needs</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="p-4 bg-background rounded-lg border-2 border-primary">
                <p className="font-semibold mb-3 flex items-center gap-2">
                  <Target className="h-5 w-5 text-primary" />
                  Start Here: What's your trading timeframe?
                </p>

                <div className="space-y-3 ml-4">
                  <div className="border-l-4 border-red-500 pl-4 py-2">
                    <p className="font-semibold text-red-600 dark:text-red-400">Trading TODAY (intraday)?</p>
                    <p className="text-sm text-muted-foreground mt-1">→ Use <strong>Latest</strong> mode</p>
                    <p className="text-xs text-muted-foreground mt-1 italic">You need zero-lag data for same-day decisions</p>
                  </div>

                  <div className="border-l-4 border-blue-500 pl-4 py-2">
                    <p className="font-semibold text-blue-600 dark:text-blue-400">Multi-week position (swing trading)?</p>
                    <p className="text-sm text-muted-foreground mt-1">→ Use <strong>EMA</strong> mode</p>
                    <p className="text-xs text-muted-foreground mt-1 italic">You want to identify trends and filter daily noise</p>
                  </div>

                  <div className="border-l-4 border-yellow-500 pl-4 py-2">
                    <p className="font-semibold text-yellow-600 dark:text-yellow-400">Long-term investment (6+ months)?</p>
                    <p className="text-sm text-muted-foreground mt-1">→ Use <strong>Average</strong> mode</p>
                    <p className="text-xs text-muted-foreground mt-1 italic">You want conservative, stable estimates for portfolio allocation</p>
                  </div>

                  <div className="border-l-4 border-primary pl-4 py-2 bg-primary/10">
                    <p className="font-semibold text-primary">Unsure of timeframe?</p>
                    <p className="text-sm text-muted-foreground mt-1">→ Use <strong>EMA</strong> mode (best default)</p>
                    <p className="text-xs text-muted-foreground mt-1 italic">EMA balances stability and reactivity for most traders</p>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Common Pitfalls */}
        <Card className="border-destructive/50 bg-destructive/5">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-destructive" />
              Common Mode Selection Mistakes
            </CardTitle>
            <CardDescription>Learn from these frequent errors</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div className="p-3 bg-background rounded-lg border border-destructive/30">
                <p className="font-semibold text-sm text-destructive mb-1">Mistake #1: Using Latest for Long-Term Decisions</p>
                <p className="text-xs text-muted-foreground mb-2">
                  <strong>Problem:</strong> Latest score of 80 looks great today, so you allocate 20% of portfolio. Tomorrow it drops to 40 due to normal volatility - now you're overexposed to an illiquid stock.
                </p>
                <p className="text-xs text-green-600 dark:text-green-400">
                  <strong>Fix:</strong> Use Average or EMA for portfolio allocation - they filter out daily noise.
                </p>
              </div>

              <div className="p-3 bg-background rounded-lg border border-destructive/30">
                <p className="font-semibold text-sm text-destructive mb-1">Mistake #2: Using Average for Day Trading</p>
                <p className="text-xs text-muted-foreground mb-2">
                  <strong>Problem:</strong> Average shows 70 (good), but market has been dead for 5 days. You place large day trade and get trapped because there's no liquidity TODAY.
                </p>
                <p className="text-xs text-green-600 dark:text-green-400">
                  <strong>Fix:</strong> Always use Latest mode for same-day trading decisions.
                </p>
              </div>

              <div className="p-3 bg-background rounded-lg border border-destructive/30">
                <p className="font-semibold text-sm text-destructive mb-1">Mistake #3: Ignoring Mode Divergences</p>
                <p className="text-xs text-muted-foreground mb-2">
                  <strong>Problem:</strong> EMA=55 but Latest=85 - you ignore this and build position assuming 55 is accurate. You miss a temporary spike that was perfect for entering/exiting.
                </p>
                <p className="text-xs text-green-600 dark:text-green-400">
                  <strong>Fix:</strong> Check all 3 modes before major trades. Divergences signal temporary opportunities or risks.
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Reference Summary */}
        <Alert className="border-primary/50 bg-primary/5">
          <Info className="h-4 w-4" />
          <AlertTitle>Quick Mode Selection Guide</AlertTitle>
          <AlertDescription>
            <ul className="space-y-1 text-sm mt-2">
              <li><strong>Trading today only?</strong> → Latest</li>
              <li><strong>Position for this week/month?</strong> → EMA</li>
              <li><strong>Portfolio planning?</strong> → Average</li>
              <li><strong>Reacting to news?</strong> → Latest</li>
              <li><strong>Identifying trends?</strong> → EMA</li>
              <li><strong>Risk assessment?</strong> → Average</li>
              <li><strong>Unsure?</strong> → Start with EMA (best default for most traders)</li>
            </ul>
          </AlertDescription>
        </Alert>
      </section>

      {/* Key Takeaways */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Key Takeaways</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card className="border-primary/50">
            <CardContent className="pt-6">
              <div className="flex items-start gap-3">
                <CheckCircle2 className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold mb-2">What Makes It Robust</p>
                  <ul className="text-sm space-y-1 text-muted-foreground">
                    <li>• Based on academic research (Amihud 2002)</li>
                    <li>• Three complementary dimensions</li>
                    <li>• Cross-sectional scaling for fairness</li>
                    <li>• Outlier protection via winsorization</li>
                    <li>• Continuity naturally incorporated via SMA</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-primary/50">
            <CardContent className="pt-6">
              <div className="flex items-start gap-3">
                <Info className="h-5 w-5 text-blue-500 shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold mb-2">Practical Usage</p>
                  <ul className="text-sm space-y-1 text-muted-foreground">
                    <li>• Scores updated daily with 60-day window</li>
                    <li>• Use Safe Trading thresholds for position sizing</li>
                    <li>• Match strategy to score (day trade: 70+, invest: 30+)</li>
                    <li>• Monitor score trends over time</li>
                    <li>• Combine with fundamental analysis</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      {/* Next Steps */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Next Steps</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Card className="border-primary/50 bg-primary/5">
            <CardContent className="pt-6">
              <div className="space-y-4">
                <p>
                  Learn how to use these liquidity scores in practice by exploring the Reports System:
                </p>
                <button
                  onClick={() => onNavigate('reports')}
                  className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                >
                  View Reports System
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </CardContent>
          </Card>

          <Card className="border-primary/50 bg-primary/5">
            <CardContent className="pt-6">
              <div className="space-y-4">
                <p>
                  Ready to visualize liquidity data? Continue to Charts Basics:
                </p>
                <button
                  onClick={() => onNavigate('charts-basics')}
                  className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                >
                  Continue to Charts
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  )
}
