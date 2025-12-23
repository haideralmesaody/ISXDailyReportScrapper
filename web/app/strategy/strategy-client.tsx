'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Target } from 'lucide-react'

export default function StrategyClient() {
  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Strategies</h1>
          <p className="text-muted-foreground">
            Batch strategies are being rebuilt to use real ISX data (no mock data).
          </p>
        </div>
        <Badge variant="secondary" className="flex items-center gap-2">
          <Target className="h-4 w-4" />
          In Development
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Planned Strategy</CardTitle>
          <CardDescription>
            The next strategy will run across all tickers and persist daily signals.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium">RSI Mean Reversion (EOD)</div>
              <div className="text-sm text-muted-foreground">
                RSI(14) overbought/oversold signals computed from the combined ISX dataset.
              </div>
            </div>
            <Badge variant="outline">Next</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

