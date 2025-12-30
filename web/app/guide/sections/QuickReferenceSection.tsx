/**
 * Quick Reference Section - Interactive Guide
 * Keyboard shortcuts, indicator parameters, workflows, troubleshooting, and API reference
 */

'use client'

import { useState } from 'react'
import { GuideSection } from '../components/GuideSection'
import { CodeExample } from '../components/CodeExample'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  BookMarked,
  Search,
  Info,
  AlertTriangle,
  Keyboard,
  TrendingUp,
  Settings,
  HelpCircle,
  ExternalLink,
  CheckCircle2,
  XCircle
} from 'lucide-react'

export function QuickReferenceSection() {
  const [searchQuery, setSearchQuery] = useState('')

  return (
    <GuideSection
      id="quick-reference"
      title="Quick Reference"
      description="Keyboard shortcuts, indicator parameters, workflows, and troubleshooting"
    >
      <div className="space-y-8">
        {/* Introduction */}
        <div className="space-y-4">
          <div className="flex items-start gap-4">
            <div className="p-3 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
              <BookMarked className="h-6 w-6 text-purple-600 dark:text-purple-400" />
            </div>
            <div className="flex-1">
              <h2 className="text-2xl font-semibold mb-2">Your ISX Pulse Cheat Sheet</h2>
              <p className="text-muted-foreground leading-relaxed">
                This section provides quick access to keyboard shortcuts, common workflows, and troubleshooting
                guides. Use the search bar below to filter content, or browse by category.
              </p>
            </div>
          </div>

          {/* Search Bar */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search shortcuts, workflows, troubleshooting..."
              className="pl-10"
            />
          </div>
        </div>

        {/* Tabbed Content */}
        <Tabs defaultValue="shortcuts" className="space-y-6">
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="shortcuts">
              <Keyboard className="h-4 w-4 mr-2" />
              <span className="hidden sm:inline">Shortcuts</span>
            </TabsTrigger>
            <TabsTrigger value="indicators">
              <TrendingUp className="h-4 w-4 mr-2" />
              <span className="hidden sm:inline">Indicators</span>
            </TabsTrigger>
            <TabsTrigger value="workflows">
              <Settings className="h-4 w-4 mr-2" />
              <span className="hidden sm:inline">Workflows</span>
            </TabsTrigger>
            <TabsTrigger value="troubleshooting">
              <HelpCircle className="h-4 w-4 mr-2" />
              <span className="hidden sm:inline">Troubleshoot</span>
            </TabsTrigger>
            <TabsTrigger value="api">
              <ExternalLink className="h-4 w-4 mr-2" />
              <span className="hidden sm:inline">API</span>
            </TabsTrigger>
          </TabsList>

          {/* TAB 1: Keyboard Shortcuts */}
          <TabsContent value="shortcuts" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Keyboard Shortcuts</CardTitle>
                <CardDescription>Speed up your workflow with keyboard navigation</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* Global Navigation */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-blue-500" />
                    Global Navigation
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Shortcut</th>
                          <th className="text-left py-2 px-3">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { key: 'G then D', action: 'Go to Dashboard' },
                          { key: 'G then M', action: 'Go to Market Overview' },
                          { key: 'G then S', action: 'Go to Strategy' },
                          { key: 'G then O', action: 'Go to Operations' },
                          { key: 'G then R', action: 'Go to Reports' },
                          { key: 'G then L', action: 'Go to Liquidity' },
                          { key: 'G then U', action: 'Go to Guide' },
                          { key: 'G then I', action: 'Go to License' },
                          { key: '?', action: 'Show keyboard shortcuts help' }
                        ].filter(item =>
                          !searchQuery ||
                          item.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          item.action.toLowerCase().includes(searchQuery.toLowerCase())
                        ).map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3">
                              <Badge variant="outline" className="font-mono text-xs">
                                {item.key}
                              </Badge>
                            </td>
                            <td className="py-2 px-3">{item.action}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Market Overview */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-green-500" />
                    Market Overview
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Shortcut</th>
                          <th className="text-left py-2 px-3">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { key: '← (Left Arrow)', action: 'Previous trading day' },
                          { key: '→ (Right Arrow)', action: 'Next trading day' },
                          { key: 'T', action: 'Jump to today' },
                          { key: 'Escape', action: 'Clear selection' }
                        ].filter(item =>
                          !searchQuery ||
                          item.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          item.action.toLowerCase().includes(searchQuery.toLowerCase())
                        ).map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3">
                              <Badge variant="outline" className="font-mono text-xs">
                                {item.key}
                              </Badge>
                            </td>
                            <td className="py-2 px-3">{item.action}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Chart Analysis */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-orange-500" />
                    Chart Analysis
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Shortcut</th>
                          <th className="text-left py-2 px-3">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { key: '+ / -', action: 'Zoom in / out' },
                          { key: 'Scroll Wheel', action: 'Zoom timeline' },
                          { key: 'Click + Drag', action: 'Pan chart' },
                          { key: 'Home', action: 'Reset zoom' },
                          { key: 'Space', action: 'Toggle crosshair' },
                          { key: 'I', action: 'Open indicators panel' },
                          { key: '1-9', action: 'Toggle indicator (by position)' }
                        ].filter(item =>
                          !searchQuery ||
                          item.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          item.action.toLowerCase().includes(searchQuery.toLowerCase())
                        ).map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3">
                              <Badge variant="outline" className="font-mono text-xs">
                                {item.key}
                              </Badge>
                            </td>
                            <td className="py-2 px-3">{item.action}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Guide Navigation */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-purple-500" />
                    Guide Navigation
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Shortcut</th>
                          <th className="text-left py-2 px-3">Action</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { key: 'N', action: 'Next section' },
                          { key: 'P', action: 'Previous section' },
                          { key: 'Ctrl/Cmd + F', action: 'Search guide' },
                          { key: 'Escape', action: 'Close guide' }
                        ].filter(item =>
                          !searchQuery ||
                          item.key.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          item.action.toLowerCase().includes(searchQuery.toLowerCase())
                        ).map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3">
                              <Badge variant="outline" className="font-mono text-xs">
                                {item.key}
                              </Badge>
                            </td>
                            <td className="py-2 px-3">{item.action}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* TAB 2: Indicator Parameters */}
          <TabsContent value="indicators" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Indicator Parameters Quick Reference</CardTitle>
                <CardDescription>Default settings and key levels for all 20+ indicators</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                {/* Moving Averages */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <Badge variant="secondary">Moving Averages</Badge>
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Indicator</th>
                          <th className="text-left py-2 px-3">Default Period</th>
                          <th className="text-left py-2 px-3">Common Periods</th>
                          <th className="text-left py-2 px-3">Best For</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { name: 'SMA', period: '20', common: '20, 50, 200', use: 'Trend identification' },
                          { name: 'EMA', period: '20', common: '12, 26, 50', use: 'Fast trend following' }
                        ].map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3 font-medium">{item.name}</td>
                            <td className="py-2 px-3 font-mono text-xs">{item.period}</td>
                            <td className="py-2 px-3 font-mono text-xs">{item.common}</td>
                            <td className="py-2 px-3 text-muted-foreground">{item.use}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Momentum Indicators */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <Badge variant="secondary">Momentum</Badge>
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Indicator</th>
                          <th className="text-left py-2 px-3">Parameters</th>
                          <th className="text-left py-2 px-3">Key Levels</th>
                          <th className="text-left py-2 px-3">Best For</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { name: 'RSI', params: 'Period: 14', levels: '70 (OB), 30 (OS), 50 (mid)', use: 'Overbought/oversold' },
                          { name: 'MACD', params: 'Fast: 12, Slow: 26, Signal: 9', levels: '0 (centerline)', use: 'Trend direction' },
                          { name: 'Stochastic', params: '%K: 14, %D: 3, Smooth: 3', levels: '80 (OB), 20 (OS)', use: 'Entry timing' },
                          { name: 'CCI', params: 'Period: 20', levels: '+100 (OB), -100 (OS)', use: 'Cyclical trends' },
                          { name: 'Williams %R', params: 'Period: 14', levels: '-20 (OB), -80 (OS)', use: 'Inverted RSI' },
                          { name: 'MFI', params: 'Period: 14', levels: '80 (OB), 20 (OS)', use: 'Volume-weighted momentum' },
                          { name: 'Momentum', params: 'Period: 10', levels: '0 (centerline)', use: 'Rate of change' },
                          { name: 'ROC', params: 'Period: 12', levels: '0 (centerline)', use: 'Percentage momentum' }
                        ].filter(item =>
                          !searchQuery ||
                          item.name.toLowerCase().includes(searchQuery.toLowerCase())
                        ).map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3 font-medium">{item.name}</td>
                            <td className="py-2 px-3 font-mono text-xs">{item.params}</td>
                            <td className="py-2 px-3 font-mono text-xs">{item.levels}</td>
                            <td className="py-2 px-3 text-muted-foreground">{item.use}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Volatility Indicators */}
                <div>
                  <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <Badge variant="secondary">Volatility</Badge>
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b">
                          <th className="text-left py-2 px-3">Indicator</th>
                          <th className="text-left py-2 px-3">Parameters</th>
                          <th className="text-left py-2 px-3">Key Levels</th>
                          <th className="text-left py-2 px-3">Best For</th>
                        </tr>
                      </thead>
                      <tbody>
                        {[
                          { name: 'Bollinger Bands', params: 'Period: 20, StdDev: 2', levels: 'Upper/Lower bands', use: 'Volatility & breakouts' },
                          { name: 'ATR', params: 'Period: 14', levels: 'N/A', use: 'Stop-loss sizing' },
                          { name: 'Keltner Channels', params: 'Period: 20, ATR Mult: 2', levels: 'Upper/Lower bands', use: 'Smoother volatility' },
                          { name: 'Donchian Channels', params: 'Period: 20', levels: 'Upper/Lower bands', use: 'Breakout identification' }
                        ].filter(item =>
                          !searchQuery ||
                          item.name.toLowerCase().includes(searchQuery.toLowerCase())
                        ).map((item, i) => (
                          <tr key={i} className="border-b hover:bg-muted/50">
                            <td className="py-2 px-3 font-medium">{item.name}</td>
                            <td className="py-2 px-3 font-mono text-xs">{item.params}</td>
                            <td className="py-2 px-3 font-mono text-xs">{item.levels}</td>
                            <td className="py-2 px-3 text-muted-foreground">{item.use}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>

                {/* Trend & Volume */}
                <div className="grid md:grid-cols-2 gap-6">
                  <div>
                    <h3 className="font-semibold mb-3 flex items-center gap-2">
                      <Badge variant="secondary">Trend</Badge>
                    </h3>
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b">
                            <th className="text-left py-2 px-3">Indicator</th>
                            <th className="text-left py-2 px-3">Period</th>
                          </tr>
                        </thead>
                        <tbody>
                          {[
                            { name: 'ADX', period: '14' },
                            { name: 'Parabolic SAR', period: '0.02, 0.2' },
                            { name: 'Ichimoku', period: '9, 26, 52' }
                          ].map((item, i) => (
                            <tr key={i} className="border-b hover:bg-muted/50">
                              <td className="py-2 px-3 font-medium">{item.name}</td>
                              <td className="py-2 px-3 font-mono text-xs">{item.period}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>

                  <div>
                    <h3 className="font-semibold mb-3 flex items-center gap-2">
                      <Badge variant="secondary">Volume</Badge>
                    </h3>
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b">
                            <th className="text-left py-2 px-3">Indicator</th>
                            <th className="text-left py-2 px-3">Type</th>
                          </tr>
                        </thead>
                        <tbody>
                          {[
                            { name: 'Volume', type: 'Histogram' },
                            { name: 'VWAP', type: 'Intraday line' },
                            { name: 'OBV', type: 'Cumulative line' }
                          ].map((item, i) => (
                            <tr key={i} className="border-b hover:bg-muted/50">
                              <td className="py-2 px-3 font-medium">{item.name}</td>
                              <td className="py-2 px-3 text-muted-foreground">{item.type}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* TAB 3: Common Workflows */}
          <TabsContent value="workflows" className="space-y-6">
            {/* Workflow 1 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Badge className="bg-blue-500">1</Badge>
                  Download and Process ISX Data
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ol className="space-y-3 text-sm">
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      1
                    </div>
                    <div>
                      <p className="font-medium">Go to Operations</p>
                      <p className="text-muted-foreground">Click "Operations" in navbar or press <Badge variant="outline" className="ml-1 font-mono text-xs">G</Badge> then <Badge variant="outline" className="font-mono text-xs">O</Badge></p>
                    </div>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      2
                    </div>
                    <div>
                      <p className="font-medium">Run Scrape Operation</p>
                      <p className="text-muted-foreground">Click "Scrape ISX" card → Leave date range empty (auto-detects missing dates) → Click "Start Scrape" → Wait (~30-60s per date)</p>
                    </div>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      3
                    </div>
                    <div>
                      <p className="font-medium">Run Process Operation</p>
                      <p className="text-muted-foreground">Click "Process Data" card → Confirm date range → Click "Start Process" → Wait (~10-20s per file)</p>
                    </div>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      4
                    </div>
                    <div>
                      <p className="font-medium">Verify Results</p>
                      <p className="text-muted-foreground">Go to Reports page → Check "Combined Data" for latest entries → Verify company names and prices</p>
                    </div>
                  </li>
                </ol>
              </CardContent>
            </Card>

            {/* Workflow 2 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Badge className="bg-green-500">2</Badge>
                  Analyze a Stock
                </CardTitle>
              </CardHeader>
              <CardContent>
                <ol className="space-y-3 text-sm">
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      1
                    </div>
                    <div>
                      <p className="font-medium">Navigate to Analysis</p>
                      <p className="text-muted-foreground">Press <Badge variant="outline" className="ml-1 font-mono text-xs">G</Badge> then <Badge variant="outline" className="font-mono text-xs">A</Badge></p>
                    </div>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      2
                    </div>
                    <div>
                      <p className="font-medium">Select Ticker</p>
                      <p className="text-muted-foreground">Use dropdown or type ticker symbol (e.g., BANK)</p>
                    </div>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      3
                    </div>
                    <div>
                      <p className="font-medium">Add Indicators</p>
                      <p className="text-muted-foreground">Click gear icon (⚙️) → Toggle on: Volume, SMA(50), SMA(200), RSI(14) → Close panel</p>
                    </div>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="flex-shrink-0 w-6 h-6 rounded-full bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 flex items-center justify-center text-xs font-semibold mt-0.5">
                      4
                    </div>
                    <div>
                      <p className="font-medium">Analyze Chart</p>
                      <p className="text-muted-foreground">Look for SMA crossovers (golden cross = bullish) → Check RSI for OB/OS (&gt;70 / &lt;30) → Verify volume on breakouts (2x average)</p>
                    </div>
                  </li>
                </ol>
              </CardContent>
            </Card>
          </TabsContent>

          {/* TAB 4: Troubleshooting */}
          <TabsContent value="troubleshooting" className="space-y-6">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertTitle>Troubleshooting Tips</AlertTitle>
              <AlertDescription>
                Follow these decision trees to diagnose and fix common issues. If problems persist, check logs in <code className="px-1 py-0.5 bg-muted rounded text-xs">dist/logs/</code> or contact support.
              </AlertDescription>
            </Alert>

            {/* Problem 1 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <XCircle className="h-5 w-5 text-red-500" />
                  Problem: Charts Not Loading
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3 text-sm pl-4">
                  <div>
                    <p className="font-medium mb-2">1. Is license active?</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-4 w-4 text-red-500 mt-0.5" />
                        <span className="text-muted-foreground">No → Go to License page → Activate license</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">Yes → Continue</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="font-medium mb-2">2. Is data available?</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-4 w-4 text-red-500 mt-0.5" />
                        <span className="text-muted-foreground">No → Go to Operations → Run "Scrape ISX" and "Process Data"</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">Yes → Continue</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="font-medium mb-2">3. Check browser console (F12)</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5" />
                        <span className="text-muted-foreground">Network errors → Check internet connection</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5" />
                        <span className="text-muted-foreground">JavaScript errors → Clear cache, reload page</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">No errors → Contact support</span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Problem 2 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <XCircle className="h-5 w-5 text-red-500" />
                  Problem: Operations Failing
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3 text-sm pl-4">
                  <div>
                    <p className="font-medium mb-2">1. Check internet connection</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-4 w-4 text-red-500 mt-0.5" />
                        <span className="text-muted-foreground">Offline → Connect to internet</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">Online → Continue</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="font-medium mb-2">2. Check ISX website status</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-4 w-4 text-red-500 mt-0.5" />
                        <span className="text-muted-foreground">ISX down → Wait for ISX to recover, retry later</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">ISX up → Continue</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="font-medium mb-2">3. Review operation logs (bottom panel)</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5" />
                        <span className="text-muted-foreground">"File not found" → Ensure Scrape ran first</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5" />
                        <span className="text-muted-foreground">"Parse error" → Excel file corrupt, re-scrape</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5" />
                        <span className="text-muted-foreground">"Network timeout" → Increase timeout in config</span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Problem 3 */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <XCircle className="h-5 w-5 text-red-500" />
                  Problem: Indicators Not Showing
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3 text-sm pl-4">
                  <div>
                    <p className="font-medium mb-2">1. Is indicator toggled on?</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <XCircle className="h-4 w-4 text-red-500 mt-0.5" />
                        <span className="text-muted-foreground">No → Open indicators panel (⚙️), toggle on</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">Yes → Continue</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="font-medium mb-2">2. Is pane height too small?</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">Yes → Drag pane separator to increase height</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <XCircle className="h-4 w-4 text-red-500 mt-0.5" />
                        <span className="text-muted-foreground">No → Continue</span>
                      </div>
                    </div>
                  </div>
                  <div>
                    <p className="font-medium mb-2">3. Check indicator period</p>
                    <div className="pl-4 space-y-1">
                      <div className="flex items-start gap-2">
                        <AlertTriangle className="h-4 w-4 text-orange-500 mt-0.5" />
                        <span className="text-muted-foreground">Period &gt; data points → Reduce period (e.g., SMA 200 → SMA 50)</span>
                      </div>
                      <div className="flex items-start gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
                        <span className="text-muted-foreground">Period OK → Try toggling off/on to refresh</span>
                      </div>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* TAB 5: API Reference */}
          <TabsContent value="api" className="space-y-6">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertTitle>Coming Soon: REST API</AlertTitle>
              <AlertDescription>
                A REST API for programmatic access to ISX data, indicators, and backtesting is planned for a future release.
                Below are mock examples showing the expected API structure.
              </AlertDescription>
            </Alert>

            <Card>
              <CardHeader>
                <CardTitle>Planned API Endpoints</CardTitle>
                <CardDescription>Future REST API for programmatic access (examples only)</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-2 px-3">Endpoint</th>
                        <th className="text-left py-2 px-3">Method</th>
                        <th className="text-left py-2 px-3">Description</th>
                      </tr>
                    </thead>
                    <tbody>
                      {[
                        { endpoint: '/api/v1/data/daily/{date}', method: 'GET', desc: 'Get daily data for date' },
                        { endpoint: '/api/v1/data/ticker/{symbol}', method: 'GET', desc: 'Get ticker historical data' },
                        { endpoint: '/api/v1/indicators/calculate', method: 'POST', desc: 'Calculate indicator' },
                        { endpoint: '/api/v1/strategy/backtest', method: 'POST', desc: 'Run backtest' },
                        { endpoint: '/api/v1/license/status', method: 'GET', desc: 'Get license info' }
                      ].map((item, i) => (
                        <tr key={i} className="border-b hover:bg-muted/50">
                          <td className="py-2 px-3 font-mono text-xs">{item.endpoint}</td>
                          <td className="py-2 px-3">
                            <Badge variant={item.method === 'GET' ? 'secondary' : 'default'} className="font-mono text-xs">
                              {item.method}
                            </Badge>
                          </td>
                          <td className="py-2 px-3 text-muted-foreground">{item.desc}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <div>
                  <h3 className="font-semibold mb-3">Example Request</h3>
                  <CodeExample
                    language="bash"
                    code={`# Get daily data for a specific date
curl http://localhost:8080/api/v1/data/daily/2025-10-10 \\
  -H "Authorization: Bearer YOUR_LICENSE_KEY" \\
  -H "Accept: application/json"`}
                  />
                </div>

                <div>
                  <h3 className="font-semibold mb-3">Example Response</h3>
                  <CodeExample
                    language="json"
                    code={`{
  "status": "success",
  "data": [
    {
      "Symbol": "BANK",
      "CompanyName": "Bank of Baghdad",
      "ClosePrice": 1.42,
      "Change": 0.03,
      "ChangePercent": 2.15,
      "Volume": 2500000,
      "Value": 3550000,
      "NumTrades": 145
    }
  ],
  "metadata": {
    "date": "2025-10-10",
    "ticker_count": 87,
    "timestamp": "2025-10-11T10:30:00Z"
  }
}`}
                  />
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Support & Resources */}
        <Card className="border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-950/30">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <HelpCircle className="h-5 w-5 text-blue-600 dark:text-blue-400" />
              Support & Resources
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <h3 className="font-semibold mb-2">Documentation</h3>
              <ul className="space-y-1 text-sm text-muted-foreground">
                <li>• <a href="/guide" className="text-blue-600 dark:text-blue-400 hover:underline">This Guide</a> - Comprehensive tutorials</li>
                <li>• <code className="px-1 py-0.5 bg-muted rounded text-xs">CLAUDE.md</code> - Project development standards</li>
                <li>• <code className="px-1 py-0.5 bg-muted rounded text-xs">ARCHITECTURE.md</code> - System architecture documentation</li>
                <li>• <code className="px-1 py-0.5 bg-muted rounded text-xs">MASTER_PLAN.md</code> - Development roadmap</li>
              </ul>
            </div>

            <div>
              <h3 className="font-semibold mb-2">Community & Support</h3>
              <ul className="space-y-1 text-sm text-muted-foreground">
                <li>• <strong>Email Support</strong>: support@isxpulse.com</li>
                <li>• <strong>Response Time</strong>: 1-2 business days</li>
              </ul>
            </div>

            <div>
              <h3 className="font-semibold mb-2">Learning Resources</h3>
              <ul className="space-y-1 text-sm text-muted-foreground">
                <li>• <a href="https://tradingview.github.io/lightweight-charts/" target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1">
                  TradingView Docs <ExternalLink className="h-3 w-3" />
                </a></li>
                <li>• <a href="https://www.investopedia.com/technical-analysis-4689657" target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1">
                  Technical Analysis Guide <ExternalLink className="h-3 w-3" />
                </a></li>
                <li>• <a href="https://www.isx-iq.net/" target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1">
                  Iraq Stock Exchange <ExternalLink className="h-3 w-3" />
                </a></li>
              </ul>
            </div>
          </CardContent>
        </Card>
      </div>
    </GuideSection>
  )
}
