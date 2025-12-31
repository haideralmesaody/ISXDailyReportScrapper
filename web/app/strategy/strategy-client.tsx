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
} from '@/types/index'
import { format } from 'date-fns'
import { ChevronDown, ChevronRight, Target } from 'lucide-react'
import { useSearchParams } from 'next/navigation'
import { Fragment, useEffect, useMemo, useState } from 'react'

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
  const [backtestStartDate, setBacktestStartDate] = useState<Date | null>(() => {
    const d = new Date()
    d.setHours(0, 0, 0, 0)
    d.setDate(d.getDate() - 90)
    return d
  })
  const [backtestEndDate, setBacktestEndDate] = useState<Date | null>(() => {
    const d = new Date()
    d.setHours(0, 0, 0, 0)
    return d
  })
  const [expandedSymbol, setExpandedSymbol] = useState<string | null>(null)
  const [backtestDetailsBySymbol, setBacktestDetailsBySymbol] = useState<Record<string, BacktestTickerDetails>>({})
  const [loadingBacktestSymbol, setLoadingBacktestSymbol] = useState<string | null>(null)

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

  const backtestSummaryBySymbol = useMemo(() => {
    const map = new Map<string, BacktestTickerSummary>()
    const items = lastRun?.backtest?.by_ticker || []
    for (const item of items) {
      map.set(item.symbol, item)
    }
    return map
  }, [lastRun])

  const hasBacktest = Boolean(lastRun?.backtest)

  const signalsToDisplay = useMemo(() => {
    const signals = lastRun?.signals || []
    const filtered = signals.filter(sig => {
      if (!showHold && sig.action === 'HOLD') return false
      if (tickerParam && sig.symbol !== tickerParam) return false
      return true
    })

    return filtered.slice().sort((a, b) => {
      const aBt = backtestSummaryBySymbol.get(a.symbol)
      const bBt = backtestSummaryBySymbol.get(b.symbol)
      if (aBt || bBt) {
        const aNet = aBt?.net_profit_pct ?? -Infinity
        const bNet = bBt?.net_profit_pct ?? -Infinity
        if (aNet !== bNet) return bNet - aNet
      }
      if (a.action !== b.action) return a.action.localeCompare(b.action)
      return a.symbol.localeCompare(b.symbol)
    })
  }, [lastRun, showHold, tickerParam, backtestSummaryBySymbol])

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

  function pct(value?: number): string {
    if (value == null || Number.isNaN(value)) return ''
    return `${value.toFixed(2)}%`
  }

  async function runBatch() {
    setRunning(true)
    try {
      const req: ExecuteBatchRequest = {
        data_points: dataPoints || 120,
      }
      if (tickerParam) req.symbols = [tickerParam]

      if (includeBacktest) {
        if (!backtestStartDate || !backtestEndDate) {
          toast({
            title: 'Backtest dates required',
            description: 'Please select a valid start and end date.',
            variant: 'destructive',
          })
          return
        }

        req.include_backtest = true
        req.backtest_start_date = format(backtestStartDate, 'yyyy-MM-dd')
        req.backtest_end_date = format(backtestEndDate, 'yyyy-MM-dd')
        req.transaction_fee = 0.006
      }

      const resp = await apiClient.executeStrategyBatch(selectedStrategyId, req)
      setLastRun(resp)
      setExpandedSymbol(null)
      setBacktestDetailsBySymbol({})
      setLoadingBacktestSymbol(null)
      if (includeBacktest) setShowHold(true)

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
                      Default last 90 days. Fee: 0.006 per transaction (0.6% on BUY and 0.6% on SELL).
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
                      {hasBacktest ? <TableHead className="w-10"></TableHead> : null}
                      <TableHead>Symbol</TableHead>
                      <TableHead>Action</TableHead>
                      {hasBacktest ? (
                        <>
                          <TableHead className="text-right">Net P&amp;L</TableHead>
                          <TableHead className="text-right">Gross P&amp;L</TableHead>
                          <TableHead className="text-right">Trades</TableHead>
                          <TableHead className="text-right">Wins</TableHead>
                          <TableHead className="text-right">Losses</TableHead>
                        </>
                      ) : null}
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
                        <TableCell colSpan={hasBacktest ? 13 : 7} className="text-sm text-muted-foreground">
                          No signals to display.
                        </TableCell>
                      </TableRow>
                    ) : (
                      signalsToDisplay.map(sig => (
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
                            {hasBacktest ? (
                              (() => {
                                const bt = backtestSummaryBySymbol.get(sig.symbol)
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

