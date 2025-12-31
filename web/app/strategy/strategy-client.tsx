'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toast } from '@/lib/hooks/use-toast'
import apiClient from '@/lib/api'
import type { ExecuteBatchResponse, StrategyInfo, StrategySignal } from '@/types/index'
import { Target } from 'lucide-react'
import { useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'

export default function StrategyClient() {
  const searchParams = useSearchParams()
  const tickerParam = useMemo(() => {
    const raw = searchParams.get('ticker')
    if (!raw) return ''
    return raw.trim().toUpperCase()
  }, [searchParams])

  const [strategies, setStrategies] = useState<StrategyInfo[]>([])
  const [selectedStrategyId, setSelectedStrategyId] = useState('rsi14_mr_eod_v1')
  const [dataPoints, setDataPoints] = useState<number>(120)
  const [loadingStrategies, setLoadingStrategies] = useState(true)
  const [running, setRunning] = useState(false)
  const [lastRun, setLastRun] = useState<ExecuteBatchResponse | null>(null)
  const [showHold, setShowHold] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoadingStrategies(true)
      try {
        const resp = await apiClient.listStrategies()
        if (cancelled) return
        setStrategies(resp.strategies || [])

        // Keep RSI as the default when available.
        if (resp.strategies?.some(s => s.id === 'rsi14_mr_eod_v1')) {
          setSelectedStrategyId('rsi14_mr_eod_v1')
        } else if (resp.strategies?.[0]?.id) {
          setSelectedStrategyId(resp.strategies[0].id)
        }
      } catch (err: any) {
        if (cancelled) return
        toast({
          title: 'Failed to load strategies',
          description: err?.detail || err?.message || 'Unknown error',
          variant: 'destructive',
        })
      } finally {
        if (!cancelled) setLoadingStrategies(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [])

  const signalsToDisplay = useMemo(() => {
    const signals = lastRun?.signals || []
    const filtered = signals.filter(sig => {
      if (!showHold && sig.action === 'HOLD') return false
      if (tickerParam && sig.symbol !== tickerParam) return false
      return true
    })

    return filtered.slice().sort((a, b) => {
      if (a.action !== b.action) return a.action.localeCompare(b.action)
      return a.symbol.localeCompare(b.symbol)
    })
  }, [lastRun, showHold, tickerParam])

  const alerts = useMemo(() => {
    const signals = lastRun?.signals || []
    const filtered = signals.filter(sig => sig.action === 'BUY' || sig.action === 'SELL')
    return tickerParam ? filtered.filter(sig => sig.symbol === tickerParam) : filtered
  }, [lastRun, tickerParam])

  function metaNumber(signal: StrategySignal, key: string): string {
    const value = (signal.metadata || ({} as any))[key] as unknown
    if (typeof value === 'number') return value.toFixed(2)
    if (typeof value === 'string') return value
    if (value == null) return ''
    return String(value)
  }

  async function runBatch() {
    setRunning(true)
    try {
      const req: { data_points: number; symbols?: string[] } = {
        data_points: dataPoints || 120,
      }
      if (tickerParam) req.symbols = [tickerParam]

      const resp = await apiClient.executeStrategyBatch(selectedStrategyId, req)
      setLastRun(resp)

      toast({
        title: 'Strategy run completed',
        description: `Run ${resp.run_id} • BUY ${resp.buy_count} • SELL ${resp.sell_count} • HOLD ${resp.hold_count}`,
      })
    } catch (err: any) {
      toast({
        title: 'Strategy run failed',
        description: err?.detail || err?.message || 'Unknown error',
        variant: 'destructive',
      })
    } finally {
      setRunning(false)
    }
  }

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Strategies</h1>
          <p className="text-muted-foreground">
            Run real-data strategies across all tickers and view BUY/SELL alerts (no mock data).
          </p>
        </div>
        <Badge variant="secondary" className="flex items-center gap-2">
          <Target className="h-4 w-4" />
          Real Data
        </Badge>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Run Strategy</CardTitle>
          <CardDescription>
            Executes the selected strategy across all tickers (or a single ticker via `?ticker=SYMBOL`) and persists outputs under `data/strategies/...`.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <div className="text-sm font-medium">Strategy</div>
              <Select
                value={selectedStrategyId}
                onValueChange={value => setSelectedStrategyId(value)}
                disabled={loadingStrategies || running}
              >
                <SelectTrigger>
                  <SelectValue placeholder={loadingStrategies ? 'Loading…' : 'Select a strategy'} />
                </SelectTrigger>
                <SelectContent>
                  {strategies.map(s => (
                    <SelectItem key={s.id} value={s.id}>
                      {s.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <div className="text-xs text-muted-foreground">
                Recommended: <span className="font-medium">RSI(14) Mean Reversion (EOD)</span> (`rsi14_mr_eod_v1`)
              </div>
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium">Data points (tail)</div>
              <Input
                type="number"
                min={16}
                max={2000}
                value={dataPoints}
                onChange={e => setDataPoints(Number(e.target.value))}
                disabled={running}
              />
              <div className="text-xs text-muted-foreground">
                Uses EOD ClosePrice and skips rows where ClosePrice == 0.
              </div>
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium">Scope</div>
              <div className="text-sm">
                {tickerParam ? (
                  <Badge variant="outline">Ticker: {tickerParam}</Badge>
                ) : (
                  <Badge variant="outline">All tickers</Badge>
                )}
              </div>
              <div className="flex gap-2">
                <Button onClick={runBatch} disabled={running || loadingStrategies || !selectedStrategyId}>
                  {running ? 'Running…' : 'Run Batch'}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setShowHold(v => !v)}
                  disabled={!lastRun}
                >
                  {showHold ? 'Hide HOLD' : 'Show HOLD'}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {lastRun && (
        <div className="grid gap-6 lg:grid-cols-3">
          <Card className="lg:col-span-1">
            <CardHeader>
              <CardTitle>Latest Run</CardTitle>
              <CardDescription>Summary and quick BUY/SELL alerts.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="text-sm text-muted-foreground">
                <div>Run ID: <span className="font-mono text-foreground">{lastRun.run_id}</span></div>
                <div>Strategy: <span className="font-mono text-foreground">{lastRun.strategy_id}</span></div>
                <div>Total: <span className="text-foreground">{lastRun.total}</span></div>
              </div>

              <div className="flex flex-wrap gap-2">
                <Badge className="bg-green-600 text-white hover:bg-green-600">BUY {lastRun.buy_count}</Badge>
                <Badge className="bg-red-600 text-white hover:bg-red-600">SELL {lastRun.sell_count}</Badge>
                <Badge variant="secondary">HOLD {lastRun.hold_count}</Badge>
                {lastRun.errors?.length ? <Badge variant="destructive">Errors {lastRun.errors.length}</Badge> : null}
              </div>

              <div className="space-y-2">
                <div className="text-sm font-medium">Alerts (BUY/SELL)</div>
                {alerts.length === 0 ? (
                  <div className="text-sm text-muted-foreground">No BUY/SELL alerts for this run.</div>
                ) : (
                  <div className="space-y-1">
                    {alerts.slice(0, 12).map(sig => (
                      <div key={sig.id} className="flex items-center justify-between text-sm">
                        <span className="font-mono">{sig.symbol}</span>
                        <Badge
                          className={sig.action === 'BUY' ? 'bg-green-600 text-white hover:bg-green-600' : 'bg-red-600 text-white hover:bg-red-600'}
                        >
                          {sig.action}
                        </Badge>
                      </div>
                    ))}
                    {alerts.length > 12 ? (
                      <div className="text-xs text-muted-foreground">Showing first 12 of {alerts.length}.</div>
                    ) : null}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <Card className="lg:col-span-2">
            <CardHeader>
              <CardTitle>Signals</CardTitle>
              <CardDescription>
                {tickerParam ? `Filtered to ${tickerParam}. ` : ''}
                {showHold ? 'Includes HOLD signals.' : 'Hides HOLD signals.'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Symbol</TableHead>
                      <TableHead>Action</TableHead>
                      <TableHead className="text-right">Price</TableHead>
                      <TableHead className="text-right">RSI</TableHead>
                      <TableHead className="text-right">Prev</TableHead>
                      <TableHead>Date</TableHead>
                      <TableHead>Reason</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {signalsToDisplay.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={7} className="text-sm text-muted-foreground">
                          No signals to display.
                        </TableCell>
                      </TableRow>
                    ) : (
                      signalsToDisplay.map(sig => (
                        <TableRow key={sig.id}>
                          <TableCell className="font-mono">{sig.symbol}</TableCell>
                          <TableCell>
                            <Badge
                              variant="outline"
                              className={
                                sig.action === 'BUY'
                                  ? 'border-green-600 text-green-700'
                                  : sig.action === 'SELL'
                                    ? 'border-red-600 text-red-700'
                                    : ''
                              }
                            >
                              {sig.action}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right">{sig.price?.toFixed?.(4) ?? String(sig.price)}</TableCell>
                          <TableCell className="text-right">{metaNumber(sig, 'rsi')}</TableCell>
                          <TableCell className="text-right">{metaNumber(sig, 'prev_rsi')}</TableCell>
                          <TableCell className="whitespace-nowrap text-sm">
                            {sig.timestamp ? new Date(sig.timestamp).toISOString().slice(0, 10) : ''}
                          </TableCell>
                          <TableCell className="max-w-[360px] truncate text-sm text-muted-foreground">
                            {sig.reasoning}
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

