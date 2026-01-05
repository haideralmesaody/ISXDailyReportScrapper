'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DatePicker } from '@/components/ui/date-picker'
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
import type {
  BacktestTickerDetails,
  BacktestTickerSummary,
  ExecuteBatchRequest,
  ExecuteBatchResponse,
  StrategyInfo,
  StrategySignal,
  StrategyRunInfo,
} from '@/types/index'
import { format } from 'date-fns'
import { ArrowDown, ArrowUp, ArrowUpDown, ChevronDown, ChevronRight, Target } from 'lucide-react'
import { useSearchParams } from 'next/navigation'
import { Fragment, useEffect, useMemo, useState } from 'react'

type SignalSortKey =
  | 'symbol'
  | 'action'
  | 'net_profit_pct'
  | 'gross_profit_pct'
  | 'completed_trades'
  | 'winning_trades'
  | 'losing_trades'
  | 'price'
  | 'rsi'
  | 'prev_rsi'
  | 'date'

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
  const [includeBacktest, setIncludeBacktest] = useState(false)
  const [backtestStartDate, setBacktestStartDate] = useState<Date | null>(null)
  const [backtestEndDate, setBacktestEndDate] = useState<Date | null>(null)
  const [expandedSymbol, setExpandedSymbol] = useState<string | null>(null)
  const [backtestDetailsBySymbol, setBacktestDetailsBySymbol] = useState<Record<string, BacktestTickerDetails>>({})
  const [loadingBacktestSymbol, setLoadingBacktestSymbol] = useState<string | null>(null)
  const [runs, setRuns] = useState<StrategyRunInfo[]>([])
  const [loadingRuns, setLoadingRuns] = useState(false)
  const [loadingRunId, setLoadingRunId] = useState<string | null>(null)
  const [signalSort, setSignalSort] = useState<{ key: SignalSortKey; dir: 'asc' | 'desc' }>(() => ({
    key: 'net_profit_pct',
    dir: 'desc',
  }))

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

  useEffect(() => {
    if (!selectedStrategyId) return
    let cancelled = false
    setLoadingRuns(true)
    apiClient
      .listStrategyRuns(selectedStrategyId, 25)
      .then(resp => {
        if (cancelled) return
        setRuns(resp.runs || [])
      })
      .catch((err: any) => {
        if (cancelled) return
        toast({
          title: 'Failed to load strategy runs',
          description: err?.detail || err?.message || 'Unknown error',
          variant: 'destructive',
        })
      })
      .finally(() => {
        if (!cancelled) setLoadingRuns(false)
      })
    return () => {
      cancelled = true
    }
  }, [selectedStrategyId])

  const backtestSummaryBySymbol = useMemo(() => {
    const map = new Map<string, BacktestTickerSummary>()
    const items = lastRun?.backtest?.by_ticker || []
    for (const item of items) {
      map.set(item.symbol, item)
    }
    return map
  }, [lastRun])

  const hasBacktest = Boolean(lastRun?.backtest)

  const computeBetterOpportunity = useCallback((bt: BacktestTickerSummary | undefined, currentPrice: number | null) => {
    if (!bt || bt.error) return null
    if (!bt.last_action || !bt.last_action_price) return null
    if (!currentPrice || !Number.isFinite(currentPrice)) return null

    const lastPrice = Number(bt.last_action_price)
    if (!Number.isFinite(lastPrice) || lastPrice <= 0) return null

    const lastAction = bt.last_action

    if (lastAction === 'SELL' && currentPrice > lastPrice) {
      const deltaPct = ((currentPrice - lastPrice) / lastPrice) * 100
      return { kind: 'BETTER_SELL' as const, deltaPct }
    }

    if (lastAction === 'BUY' && currentPrice < lastPrice) {
      const deltaPct = ((currentPrice - lastPrice) / lastPrice) * 100
      return { kind: 'BETTER_BUY' as const, deltaPct }
    }

    return null
  }, [])

  useEffect(() => {
    if (hasBacktest) return
    const backtestKeys: SignalSortKey[] = [
      'net_profit_pct',
      'gross_profit_pct',
      'completed_trades',
      'winning_trades',
      'losing_trades',
    ]
    if (backtestKeys.includes(signalSort.key)) {
      setSignalSort({ key: 'action', dir: 'asc' })
    }
  }, [hasBacktest, signalSort.key])

  const signalsToDisplay = useMemo(() => {
    const signals = lastRun?.signals || []
    const filtered = signals.filter(sig => {
      if (!showHold && sig.action === 'HOLD') return false
      if (tickerParam && sig.symbol !== tickerParam) return false
      return true
    })

    const metaNumeric = (sig: StrategySignal, key: string): number | null => {
      const raw = (sig.metadata || ({} as any))[key] as unknown
      if (typeof raw === 'number' && Number.isFinite(raw)) return raw
      if (typeof raw === 'string') {
        const parsed = Number(raw)
        return Number.isFinite(parsed) ? parsed : null
      }
      return null
    }

    const numberOrNull = (value: unknown): number | null => {
      if (typeof value === 'number' && Number.isFinite(value)) return value
      if (typeof value === 'string') {
        const parsed = Number(value)
        if (Number.isFinite(parsed)) return parsed
      }
      return null
    }

    const sortValue = (sig: StrategySignal, key: SignalSortKey): string | number | null => {
      const bt = backtestSummaryBySymbol.get(sig.symbol)
      switch (key) {
        case 'symbol':
          return sig.symbol || ''
        case 'action':
          return sig.action || ''
        case 'net_profit_pct':
          return bt?.error ? null : (bt?.net_profit_pct ?? null)
        case 'gross_profit_pct':
          return bt?.error ? null : (bt?.gross_profit_pct ?? null)
        case 'completed_trades':
          return bt?.error ? null : (bt?.completed_trades ?? null)
        case 'winning_trades':
          return bt?.error ? null : (bt?.winning_trades ?? null)
        case 'losing_trades':
          return bt?.error ? null : (bt?.losing_trades ?? null)
        case 'price':
          return numberOrNull(sig.price)
        case 'rsi':
          return metaNumeric(sig, 'rsi')
        case 'prev_rsi':
          return metaNumeric(sig, 'prev_rsi')
        case 'date': {
          if (!sig.timestamp) return null
          const t = new Date(sig.timestamp).getTime()
          return Number.isFinite(t) ? t : null
        }
        default:
          return null
      }
    }

    const compare = (a: StrategySignal, b: StrategySignal): number => {
      const aVal = sortValue(a, signalSort.key)
      const bVal = sortValue(b, signalSort.key)
      const dirMul = signalSort.dir === 'asc' ? 1 : -1

      if (aVal == null && bVal == null) return 0
      if (aVal == null) return 1
      if (bVal == null) return -1

      if (typeof aVal === 'number' || typeof bVal === 'number') {
        const an = typeof aVal === 'number' ? aVal : Number(aVal)
        const bn = typeof bVal === 'number' ? bVal : Number(bVal)
        if (an !== bn) return (an < bn ? -1 : 1) * dirMul
      } else {
        const as = String(aVal)
        const bs = String(bVal)
        const cmp = as.localeCompare(bs)
        if (cmp !== 0) return cmp * dirMul
      }

      if (a.action !== b.action) return a.action.localeCompare(b.action)
      return a.symbol.localeCompare(b.symbol)
    }

    return filtered.slice().sort(compare)
  }, [lastRun, showHold, tickerParam, backtestSummaryBySymbol, signalSort])

  const alerts = useMemo(() => {
    const signals = lastRun?.signals || []
    const filtered = signals.filter(sig => sig.action === 'BUY' || sig.action === 'SELL')
    return tickerParam ? filtered.filter(sig => sig.symbol === tickerParam) : filtered
  }, [lastRun, tickerParam])

  const openAnalysis = (symbol: string) => {
    const params = new URLSearchParams({ ticker: symbol })
    if (selectedStrategyId) params.set('strategy_id', selectedStrategyId)
    if (lastRun?.run_id) params.set('run_id', lastRun.run_id)
    window.location.href = `/analysis?${params.toString()}`
  }

  function metaNumber(signal: StrategySignal, key: string): string {
    const value = (signal.metadata || ({} as any))[key] as unknown
    if (typeof value === 'number') return value.toFixed(2)
    if (typeof value === 'string') return value
    if (value == null) return ''
    return String(value)
  }

  function pct(value?: number): string {
    if (value == null || Number.isNaN(value)) return ''
    return `${value.toFixed(2)}%`
  }

  function toLocalDate(dateStr: string): Date | null {
    const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateStr.trim())
    if (!match) return null
    const year = Number(match[1])
    const month = Number(match[2])
    const day = Number(match[3])
    if (!Number.isFinite(year) || !Number.isFinite(month) || !Number.isFinite(day)) return null
    return new Date(year, month - 1, day)
  }

  function toggleSignalSort(key: SignalSortKey) {
    setSignalSort(prev => {
      if (prev.key === key) return { key, dir: prev.dir === 'asc' ? 'desc' : 'asc' }
      const defaultDir: 'asc' | 'desc' =
        key === 'symbol' || key === 'action' ? 'asc' : 'desc'
      return { key, dir: defaultDir }
    })
  }

  function SortIcon({ k }: { k: SignalSortKey }) {
    if (signalSort.key !== k) return <ArrowUpDown className="h-4 w-4 opacity-50" />
    return signalSort.dir === 'asc' ? (
      <ArrowUp className="h-4 w-4" />
    ) : (
      <ArrowDown className="h-4 w-4" />
    )
  }

  async function runBatch() {
    setRunning(true)
    try {
      const req: ExecuteBatchRequest = {
        data_points: dataPoints || 120,
      }
      if (tickerParam) req.symbols = [tickerParam]

      if (includeBacktest) {
        req.include_backtest = true
        if (backtestStartDate) req.backtest_start_date = format(backtestStartDate, 'yyyy-MM-dd')
        if (backtestEndDate) req.backtest_end_date = format(backtestEndDate, 'yyyy-MM-dd')
        req.transaction_fee = 0.006
      }

      const resp = await apiClient.executeStrategyBatch(selectedStrategyId, req)
      setLastRun(resp)
      if (includeBacktest && resp.backtest) {
        if (!backtestStartDate) {
          const resolved = toLocalDate(resp.backtest.start_date)
          if (resolved) setBacktestStartDate(resolved)
        }
        if (!backtestEndDate) {
          const resolved = toLocalDate(resp.backtest.end_date)
          if (resolved) setBacktestEndDate(resolved)
        }
      }
      setExpandedSymbol(null)
      setBacktestDetailsBySymbol({})
      setLoadingBacktestSymbol(null)
      if (includeBacktest) setShowHold(true)
      try {
        const runsResp = await apiClient.listStrategyRuns(selectedStrategyId, 25)
        setRuns(runsResp.runs || [])
      } catch {
        // Ignore failures; the main run succeeded.
      }

      toast({
        title: 'Strategy run completed',
        description: `Run ${resp.run_id} - BUY ${resp.buy_count} - SELL ${resp.sell_count} - HOLD ${resp.hold_count}`,
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

  async function loadRun(runId: string) {
    if (!selectedStrategyId) return
    setLoadingRunId(runId)
    try {
      const resp = await apiClient.getStrategyRun(selectedStrategyId, runId)
      setLastRun(resp)
      if (resp.backtest) {
        const resolvedStart = toLocalDate(resp.backtest.start_date)
        const resolvedEnd = toLocalDate(resp.backtest.end_date)
        if (resolvedStart) setBacktestStartDate(resolvedStart)
        if (resolvedEnd) setBacktestEndDate(resolvedEnd)
      }
      setExpandedSymbol(null)
      setBacktestDetailsBySymbol({})
      setLoadingBacktestSymbol(null)
      setShowHold(Boolean(resp.backtest))
    } catch (err: any) {
      toast({
        title: 'Failed to load run',
        description: err?.detail || err?.message || 'Unknown error',
        variant: 'destructive',
      })
    } finally {
      setLoadingRunId(null)
    }
  }

  async function toggleExpand(symbol: string) {
    if (!lastRun?.run_id || !lastRun.backtest) return

    if (expandedSymbol === symbol) {
      setExpandedSymbol(null)
      return
    }

    setExpandedSymbol(symbol)
    if (backtestDetailsBySymbol[symbol]) return

    setLoadingBacktestSymbol(symbol)
    try {
      const details = await apiClient.getBacktestTickerDetails(lastRun.strategy_id, lastRun.run_id, symbol)
      setBacktestDetailsBySymbol(prev => ({ ...prev, [symbol]: details }))
    } catch (err: any) {
      toast({
        title: 'Failed to load trade details',
        description: err?.detail || err?.message || 'Unknown error',
        variant: 'destructive',
      })
      setExpandedSymbol(null)
    } finally {
      setLoadingBacktestSymbol(null)
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
              <div className="space-y-2 pt-1">
                <div className="flex items-center gap-2">
                  <Checkbox
                    checked={includeBacktest}
                    onCheckedChange={checked => setIncludeBacktest(checked === true)}
                    disabled={running}
                  />
                  <span className="text-sm">Include backtest</span>
                </div>

                {includeBacktest ? (
                  <div className="grid gap-2 md:grid-cols-2">
                    <DatePicker
                      selected={backtestStartDate}
                      onChange={date => setBacktestStartDate(date)}
                      placeholderText="Start date"
                      disabled={running}
                    />
                    <DatePicker
                      selected={backtestEndDate}
                      onChange={date => setBacktestEndDate(date)}
                      placeholderText="End date"
                      disabled={running}
                    />
                    <div className="md:col-span-2 text-xs text-muted-foreground">
                      Defaults to full available data range when dates are left empty. Fee: 0.006 per transaction (0.6% on BUY and 0.6% on SELL).
                    </div>
                  </div>
                ) : null}
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

      <Card>
        <CardHeader>
          <CardTitle>Recent Runs</CardTitle>
          <CardDescription>Browse persisted runs for the selected strategy and load them into the page.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          {loadingRuns ? (
            <div className="text-sm text-muted-foreground">Loading…</div>
          ) : runs.length === 0 ? (
            <div className="text-sm text-muted-foreground">No runs found yet.</div>
          ) : (
            <div className="space-y-2">
              {runs.map(run => (
                <div
                  key={run.run_id}
                  className="flex flex-col gap-2 rounded-md border p-3 md:flex-row md:items-center md:justify-between"
                >
                  <div className="space-y-1">
                    <div className="font-mono text-sm">{run.run_id}</div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(run.started_at).toLocaleString()} • BUY {run.buy_count} • SELL {run.sell_count} • HOLD {run.hold_count} • Errors {run.error_count}
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {run.has_backtest ? (
                        <Badge variant="secondary">Backtest</Badge>
                      ) : (
                        <Badge variant="outline">No backtest</Badge>
                      )}
                      {lastRun?.run_id === run.run_id ? <Badge variant="default">Loaded</Badge> : null}
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    onClick={() => loadRun(run.run_id)}
                    disabled={running || loadingRunId === run.run_id}
                  >
                    {loadingRunId === run.run_id ? 'Loading…' : 'Load'}
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {lastRun && (
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-1 space-y-6">
            <Card>
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

            {lastRun.backtest?.aggregate ? (
              <Card>
                <CardHeader>
                  <CardTitle>Backtest Aggregate</CardTitle>
                  <CardDescription>
                    {lastRun.backtest.start_date} → {lastRun.backtest.end_date}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <div className="flex flex-wrap gap-3">
                    <div>
                      <span className="text-muted-foreground">Tickers:</span>{' '}
                      <span className="font-medium">
                        {lastRun.backtest.aggregate.successful_tickers}/{lastRun.backtest.aggregate.total_tickers}
                      </span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Errors:</span>{' '}
                      <span className="font-medium">{lastRun.backtest.aggregate.error_tickers}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Open:</span>{' '}
                      <span className="font-medium">{lastRun.backtest.aggregate.open_positions}</span>
                    </div>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <div>
                      <span className="text-muted-foreground">Trades:</span>{' '}
                      <span className="font-medium">{lastRun.backtest.aggregate.total_completed_trades}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Wins:</span>{' '}
                      <span className="font-medium">{lastRun.backtest.aggregate.total_winning_trades}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Losses:</span>{' '}
                      <span className="font-medium">{lastRun.backtest.aggregate.total_losing_trades}</span>
                    </div>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <div>
                      <span className="text-muted-foreground">Avg net:</span>{' '}
                      <span className="font-medium">{pct(lastRun.backtest.aggregate.avg_net_profit_pct)}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Median net:</span>{' '}
                      <span className="font-medium">{pct(lastRun.backtest.aggregate.median_net_profit_pct)}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Avg gross:</span>{' '}
                      <span className="font-medium">{pct(lastRun.backtest.aggregate.avg_gross_profit_pct)}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Median gross:</span>{' '}
                      <span className="font-medium">{pct(lastRun.backtest.aggregate.median_gross_profit_pct)}</span>
                    </div>
                  </div>

                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-2">
                      <div className="text-sm font-medium">Top</div>
                      <div className="space-y-1">
                        {(lastRun.backtest.aggregate.top_tickers || []).slice(0, 5).map(t => (
                          <div key={`top-${t.symbol}`} className="flex items-center justify-between text-xs">
                            <span className="font-mono">{t.symbol}</span>
                            <span className="font-medium">{pct(t.net_profit_pct)}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                    <div className="space-y-2">
                      <div className="text-sm font-medium">Bottom</div>
                      <div className="space-y-1">
                        {(lastRun.backtest.aggregate.bottom_tickers || []).slice(0, 5).map(t => (
                          <div key={`bottom-${t.symbol}`} className="flex items-center justify-between text-xs">
                            <span className="font-mono">{t.symbol}</span>
                            <span className="font-medium">{pct(t.net_profit_pct)}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ) : null}
          </div>

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
                      {hasBacktest ? <TableHead className="w-10"></TableHead> : null}
                      <TableHead>
                        <Button variant="ghost" className="h-auto p-0" onClick={() => toggleSignalSort('symbol')}>
                          <span className="flex items-center gap-1">
                            Symbol <SortIcon k="symbol" />
                          </span>
                        </Button>
                      </TableHead>
                      <TableHead>
                        <Button variant="ghost" className="h-auto p-0" onClick={() => toggleSignalSort('action')}>
                          <span className="flex items-center gap-1">
                            Action <SortIcon k="action" />
                          </span>
                        </Button>
                      </TableHead>
                      {hasBacktest ? (
                        <>
                          <TableHead className="text-right">
                            <Button
                              variant="ghost"
                              className="h-auto w-full justify-end p-0"
                              onClick={() => toggleSignalSort('net_profit_pct')}
                            >
                              <span className="flex items-center gap-1">
                                Net P&amp;L <SortIcon k="net_profit_pct" />
                              </span>
                            </Button>
                          </TableHead>
                          <TableHead className="text-right">
                            <Button
                              variant="ghost"
                              className="h-auto w-full justify-end p-0"
                              onClick={() => toggleSignalSort('gross_profit_pct')}
                            >
                              <span className="flex items-center gap-1">
                                Gross P&amp;L <SortIcon k="gross_profit_pct" />
                              </span>
                            </Button>
                          </TableHead>
                          <TableHead className="text-right">
                            <Button
                              variant="ghost"
                              className="h-auto w-full justify-end p-0"
                              onClick={() => toggleSignalSort('completed_trades')}
                            >
                              <span className="flex items-center gap-1">
                                Trades <SortIcon k="completed_trades" />
                              </span>
                            </Button>
                          </TableHead>
                          <TableHead className="text-right">
                            <Button
                              variant="ghost"
                              className="h-auto w-full justify-end p-0"
                              onClick={() => toggleSignalSort('winning_trades')}
                            >
                              <span className="flex items-center gap-1">
                                Wins <SortIcon k="winning_trades" />
                              </span>
                            </Button>
                          </TableHead>
                          <TableHead className="text-right">
                            <Button
                              variant="ghost"
                              className="h-auto w-full justify-end p-0"
                              onClick={() => toggleSignalSort('losing_trades')}
                            >
                              <span className="flex items-center gap-1">
                                Losses <SortIcon k="losing_trades" />
                              </span>
                            </Button>
                          </TableHead>
                        </>
                      ) : null}
                      <TableHead className="text-right">
                        <Button
                          variant="ghost"
                          className="h-auto w-full justify-end p-0"
                          onClick={() => toggleSignalSort('price')}
                        >
                          <span className="flex items-center gap-1">
                            Price <SortIcon k="price" />
                          </span>
                        </Button>
                      </TableHead>
                      <TableHead className="text-right">
                        <Button
                          variant="ghost"
                          className="h-auto w-full justify-end p-0"
                          onClick={() => toggleSignalSort('rsi')}
                        >
                          <span className="flex items-center gap-1">
                            RSI <SortIcon k="rsi" />
                          </span>
                        </Button>
                      </TableHead>
                      <TableHead className="text-right">
                        <Button
                          variant="ghost"
                          className="h-auto w-full justify-end p-0"
                          onClick={() => toggleSignalSort('prev_rsi')}
                        >
                          <span className="flex items-center gap-1">
                            Prev <SortIcon k="prev_rsi" />
                          </span>
                        </Button>
                      </TableHead>
                      <TableHead>
                        <Button variant="ghost" className="h-auto p-0" onClick={() => toggleSignalSort('date')}>
                          <span className="flex items-center gap-1">
                            Date <SortIcon k="date" />
                          </span>
                        </Button>
                      </TableHead>
                      <TableHead>Reason</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {signalsToDisplay.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={hasBacktest ? 13 : 7} className="text-sm text-muted-foreground">
                          No signals to display.
                        </TableCell>
                      </TableRow>
                    ) : (
                      signalsToDisplay.map(sig => {
                        const bt = hasBacktest ? backtestSummaryBySymbol.get(sig.symbol) : undefined
                        const currentPrice = numberOrNull(sig.price)
                        const better = computeBetterOpportunity(bt, currentPrice)

                        return (
                        <Fragment key={sig.id}>
                          <TableRow>
                            {hasBacktest ? (
                              <TableCell>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  className="h-8 w-8"
                                  onClick={() => toggleExpand(sig.symbol)}
                                  disabled={!lastRun?.run_id || !lastRun?.backtest}
                                >
                                  {expandedSymbol === sig.symbol ? (
                                    <ChevronDown className="h-4 w-4" />
                                  ) : (
                                    <ChevronRight className="h-4 w-4" />
                                  )}
                                </Button>
                              </TableCell>
                            ) : null}
                            <TableCell className="font-mono">
                              <button
                                type="button"
                                className="underline underline-offset-2 hover:no-underline"
                                onClick={() => openAnalysis(sig.symbol)}
                              >
                                {sig.symbol}
                              </button>
                            </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
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
                              {better ? (
                                <Badge
                                  className={
                                    better.kind === 'BETTER_BUY'
                                      ? 'bg-green-600 text-white hover:bg-green-600/90'
                                      : 'bg-red-600 text-white hover:bg-red-600/90'
                                  }
                                >
                                  {better.kind === 'BETTER_BUY' ? 'Better Buy' : 'Better Sell'}{' '}
                                  <span className="opacity-90">
                                    ({better.deltaPct >= 0 ? '+' : ''}
                                    {better.deltaPct.toFixed(2)}%)
                                  </span>
                                </Badge>
                              ) : null}
                            </div>
                          </TableCell>
                            {hasBacktest ? (
                              (() => {
                                const isError = Boolean(bt?.error)
                                return (
                                  <>
                                    <TableCell className="text-right">
                                      {isError ? <span className="text-destructive">ERR</span> : pct(bt?.net_profit_pct)}
                                    </TableCell>
                                    <TableCell className="text-right">
                                      {isError ? <span className="text-destructive">ERR</span> : pct(bt?.gross_profit_pct)}
                                    </TableCell>
                                    <TableCell className="text-right">{bt?.completed_trades ?? ''}</TableCell>
                                    <TableCell className="text-right">{bt?.winning_trades ?? ''}</TableCell>
                                    <TableCell className="text-right">{bt?.losing_trades ?? ''}</TableCell>
                                  </>
                                )
                              })()
                            ) : null}
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

                          {hasBacktest && expandedSymbol === sig.symbol ? (
                            <TableRow>
                              <TableCell colSpan={13} className="bg-muted/30">
                                {loadingBacktestSymbol === sig.symbol ? (
                                  <div className="text-sm text-muted-foreground">Loading trade details...</div>
                                ) : (() => {
                                  const details = backtestDetailsBySymbol[sig.symbol]
                                  if (!details) {
                                    return (
                                      <div className="text-sm text-muted-foreground">
                                        No backtest details found for this ticker.
                                      </div>
                                    )
                                  }

                                  const fee = details.trades[0]?.transaction_fee ?? lastRun?.backtest?.transaction_fee ?? 0.006

                                  return (
                                    <div className="space-y-3">
                                      <div className="flex flex-wrap gap-3 text-sm">
                                        <div>
                                          <span className="text-muted-foreground">Net:</span>{' '}
                                          <span className="font-medium">{pct(details.summary.net_profit_pct)}</span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Gross:</span>{' '}
                                          <span className="font-medium">{pct(details.summary.gross_profit_pct)}</span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Trades:</span>{' '}
                                          <span className="font-medium">{details.summary.completed_trades}</span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Wins:</span>{' '}
                                          <span className="font-medium">{details.summary.winning_trades}</span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Losses:</span>{' '}
                                          <span className="font-medium">{details.summary.losing_trades}</span>
                                        </div>
                                        <div>
                                          <span className="text-muted-foreground">Fee:</span>{' '}
                                          <span className="font-medium">{(fee * 100).toFixed(2)}%/tx</span>
                                        </div>
                                      </div>

                                      <div className="overflow-x-auto">
                                        <Table>
                                          <TableHeader>
                                            <TableRow>
                                              <TableHead>Status</TableHead>
                                              <TableHead>Buy Date</TableHead>
                                              <TableHead className="text-right">Buy</TableHead>
                                              <TableHead>Sell / MTM Date</TableHead>
                                              <TableHead className="text-right">Sell / MTM</TableHead>
                                              <TableHead className="text-right">Gross %</TableHead>
                                              <TableHead className="text-right">Net %</TableHead>
                                            </TableRow>
                                          </TableHeader>
                                          <TableBody>
                                            {details.trades.length === 0 ? (
                                              <TableRow>
                                                <TableCell colSpan={7} className="text-sm text-muted-foreground">
                                                  No trades for this period.
                                                </TableCell>
                                              </TableRow>
                                            ) : (
                                              details.trades.map((t, idx) => (
                                                <TableRow key={`${sig.symbol}-t-${idx}`}>
                                                  <TableCell>{t.status}</TableCell>
                                                  <TableCell className="whitespace-nowrap">{t.buy_date || ''}</TableCell>
                                                  <TableCell className="text-right">{t.buy_price?.toFixed?.(4) ?? ''}</TableCell>
                                                  <TableCell className="whitespace-nowrap">{t.sell_date || ''}</TableCell>
                                                  <TableCell className="text-right">{t.sell_price?.toFixed?.(4) ?? ''}</TableCell>
                                                  <TableCell className="text-right">{pct(t.gross_return_pct)}</TableCell>
                                                  <TableCell className="text-right">{pct(t.net_return_pct)}</TableCell>
                                                </TableRow>
                                              ))
                                            )}
                                          </TableBody>
                                        </Table>
                                      </div>
                                    </div>
                                  )
                                })()}
                              </TableCell>
                            </TableRow>
                          ) : null}
                        </Fragment>
                        )
                      })
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

