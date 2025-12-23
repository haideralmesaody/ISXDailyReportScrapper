/**
 * Charts & Analysis Section - Interactive Guide
 * Explains the Analysis page charting, indicators, and drawing tools.
 */

'use client'

import { GuideSection } from '../components/GuideSection'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  CandlestickChart,
  Activity,
  Pencil,
  Save,
  Lock,
  Trash2,
  ZoomIn,
  ExternalLink,
} from 'lucide-react'

export function ChartsSection() {
  return (
    <GuideSection
      id="charts"
      title="Charts & Analysis"
      description="TradingView-style charting with indicators and drawing tools"
    >
      <div className="space-y-8">
        <div className="space-y-4">
          <div className="flex items-start gap-4">
            <div className="p-3 bg-emerald-100 dark:bg-emerald-900/30 rounded-lg">
              <CandlestickChart className="h-6 w-6 text-emerald-700 dark:text-emerald-300" />
            </div>
            <div className="flex-1">
              <h2 className="text-2xl font-semibold mb-2">Analysis Page</h2>
              <p className="text-muted-foreground leading-relaxed">
                The Analysis page (<code className="px-2 py-1 bg-muted rounded">/analysis</code>) provides an interactive
                chart powered by TradingView Lightweight Charts. You can switch chart type/timeframe, add indicators, and
                draw on the chart using TradingView-style tools.
              </p>
              <div className="mt-4 flex flex-wrap gap-2">
                <Button asChild>
                  <a href="/analysis" target="_blank" className="gap-2">
                    <ExternalLink className="h-4 w-4" />
                    Open Analysis
                  </a>
                </Button>
                <Badge variant="secondary" className="gap-2">
                  <ZoomIn className="h-3.5 w-3.5" />
                  Zoom preserved when toggling indicators
                </Badge>
              </div>
            </div>
          </div>

          <Alert>
            <Save className="h-4 w-4" />
            <AlertTitle>Per-ticker persistence</AlertTitle>
            <AlertDescription>
              Drawings are saved per ticker and automatically restored when you revisit the same ticker. The backend
              stores drawings in <code className="px-1 py-0.5 bg-muted rounded">data/indicators/&lt;TICKER&gt;.json</code>.
            </AlertDescription>
          </Alert>
        </div>

        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-emerald-500 to-teal-500 rounded" />
            <h2 className="text-2xl font-semibold">Layout & Controls</h2>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Ticker list</CardTitle>
              <CardDescription>Pick a symbol and start analyzing</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-muted-foreground">
              <p>
                Use the <strong>Tickers</strong> button to collapse/expand the ticker list. The panel supports resizing
                and smooth width animation.
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Indicators
              </CardTitle>
              <CardDescription>Add overlays and separate panes</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-muted-foreground">
              <p>
                Use the <strong>Indicators</strong> button to toggle the indicators sidebar. Each indicator can be
                enabled/disabled, and some provide a settings gear for parameters.
              </p>
              <p>
                Adding/removing indicators keeps your current zoom level (no forced fit-to-screen).
              </p>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-4">
          <div className="flex items-center gap-3 mb-4">
            <div className="h-1 w-12 bg-gradient-to-r from-indigo-500 to-purple-500 rounded" />
            <h2 className="text-2xl font-semibold">Drawing Tools</h2>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Pencil className="h-5 w-5" />
                Tools (Phase 1)
              </CardTitle>
              <CardDescription>TradingView-style drawing workflow</CardDescription>
            </CardHeader>
          <CardContent className="space-y-3 text-sm text-muted-foreground">
            <p>
              Use the left toolbar to draw: Trend line, Ray, Horizontal/Vertical line, Rectangle, Text, Arrow, and
              Measure, and Parallel Channel.
            </p>
              <p>
                You can select and move/resize drawings. The toolbar includes actions like <strong>Undo/Redo</strong>,{' '}
                <strong>Delete</strong>, <strong>Lock</strong>, and optional <strong>Magnet</strong> snap.
              </p>
              <div className="flex flex-wrap gap-2">
                <Badge variant="outline" className="gap-2">
                  <Trash2 className="h-3.5 w-3.5" />
                  Delete removes selection
                </Badge>
                <Badge variant="outline" className="gap-2">
                  <Lock className="h-3.5 w-3.5" />
                  Lock prevents edits
                </Badge>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-3">
          <h2 className="text-xl font-semibold">Next Steps</h2>
          <div className="flex flex-wrap gap-3">
            <Button variant="outline" asChild>
              <a href="/guide?section=strategy">Trading Strategies →</a>
            </Button>
            <Button variant="outline" asChild>
              <a href="/guide?section=quick-reference">Quick Reference →</a>
            </Button>
          </div>
        </div>
      </div>
    </GuideSection>
  )
}
