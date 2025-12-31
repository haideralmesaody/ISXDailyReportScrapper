'use client'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import apiClient from '@/lib/api'
import { toast } from '@/lib/hooks/use-toast'
import {
  CandlestickData,
  CandlestickSeries,
  ColorType,
  CrosshairMode,
  IChartApi,
  ISeriesApi,
  Time,
  createChart,
} from 'lightweight-charts'
import { ChartCandlestick } from 'lucide-react'
import { useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useRef, useState } from 'react'

type TickerListItem = {
  symbol: string
  name?: string
}

type ChartCandle = {
  date: string
  open: number
  high: number
  low: number
  close: number
  volume?: number
}

type TickerChartResponse = {
  status: 'success'
  data: {
    ticker: string
    data: ChartCandle[]
  }
  ticker: string
}

type TickersResponse = {
  status: 'success'
  data: any
  count?: number
}

function extractSymbol(item: any): string {
  const raw = item?.Symbol ?? item?.symbol ?? item?.ticker ?? item?.Ticker ?? ''
  return String(raw || '').trim().toUpperCase()
}

function extractName(item: any): string {
  const raw = item?.Name ?? item?.name ?? item?.Company ?? item?.company ?? ''
  return String(raw || '').trim()
}

export default function AnalysisClient() {
  const searchParams = useSearchParams()
  const urlTicker = (searchParams.get('ticker') || '').trim().toUpperCase()

  const [tickers, setTickers] = useState<TickerListItem[]>([])
  const [loadingTickers, setLoadingTickers] = useState(true)
  const [selected, setSelected] = useState<string>(urlTicker)
  const [query, setQuery] = useState('')

  const [candles, setCandles] = useState<ChartCandle[]>([])
  const [loadingChart, setLoadingChart] = useState(false)

  const chartContainerRef = useRef<HTMLDivElement | null>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const seriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoadingTickers(true)
      try {
        const resp = await apiClient.get<TickersResponse>('/api/data/tickers', true)
        if (cancelled) return

        const data = resp?.data
        const list = Array.isArray(data) ? data : Array.isArray(data?.tickers) ? data.tickers : []
        const normalized = list
          .map((item: any) => {
            const symbol = extractSymbol(item)
            if (!symbol) return null
            const name = extractName(item)
            return { symbol, ...(name ? { name } : {}) }
          })
          .filter(Boolean) as TickerListItem[]

        normalized.sort((a, b) => a.symbol.localeCompare(b.symbol))
        setTickers(normalized)
      } catch (err: any) {
        toast({
          title: 'Failed to load tickers',
          description: err?.detail || err?.message || 'Unknown error',
          variant: 'destructive',
        })
        setTickers([])
      } finally {
        if (!cancelled) setLoadingTickers(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!selected) {
      setCandles([])
      return
    }

    let cancelled = false
    async function load() {
      setLoadingChart(true)
      try {
        const resp = await apiClient.get<TickerChartResponse>(`/api/data/ticker/${encodeURIComponent(selected)}/chart`)
        if (cancelled) return
        setCandles(resp?.data?.data || [])
      } catch (err: any) {
        if (cancelled) return
        toast({
          title: 'Failed to load chart data',
          description: err?.detail || err?.message || 'Unknown error',
          variant: 'destructive',
        })
        setCandles([])
      } finally {
        if (!cancelled) setLoadingChart(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [selected])

  const filteredTickers = useMemo(() => {
    if (!query.trim()) return tickers
    const q = query.trim().toUpperCase()
    return tickers.filter(t => t.symbol.includes(q) || (t.name || '').toUpperCase().includes(q))
  }, [tickers, query])

  useEffect(() => {
    const el = chartContainerRef.current
    if (!el) return

    if (chartRef.current) {
      chartRef.current.remove()
      chartRef.current = null
      seriesRef.current = null
    }

    const chart = createChart(el, {
      layout: {
        background: { type: ColorType.Solid, color: 'transparent' },
        textColor: '#94a3b8',
      },
      grid: {
        vertLines: { color: 'rgba(148, 163, 184, 0.10)' },
        horzLines: { color: 'rgba(148, 163, 184, 0.10)' },
      },
      width: el.clientWidth,
      height: el.clientHeight,
      rightPriceScale: { borderVisible: false },
      timeScale: { borderVisible: false },
      crosshair: {
        mode: CrosshairMode.Magnet,
        vertLine: { width: 1, color: 'rgba(148, 163, 184, 0.35)', style: 0, visible: true, labelVisible: false },
        horzLine: { width: 1, color: 'rgba(148, 163, 184, 0.35)', style: 0, visible: true, labelVisible: false },
      },
    })

    const series = chart.addSeries(CandlestickSeries, {
      upColor: '#16a34a',
      downColor: '#dc2626',
      borderUpColor: '#16a34a',
      borderDownColor: '#dc2626',
      wickUpColor: '#16a34a',
      wickDownColor: '#dc2626',
    })

    chartRef.current = chart
    seriesRef.current = series

    const ro = new ResizeObserver(() => {
      chart.applyOptions({ width: el.clientWidth, height: el.clientHeight })
    })
    ro.observe(el)

    return () => {
      ro.disconnect()
      chart.remove()
      chartRef.current = null
      seriesRef.current = null
    }
  }, [])

  useEffect(() => {
    const series = seriesRef.current
    if (!series) return

    const data: CandlestickData<Time>[] = candles
      .filter(c => c && c.date)
      .map(c => ({
        time: c.date as Time,
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
      }))

    series.setData(data)
    chartRef.current?.timeScale().fitContent()
  }, [candles])

  useEffect(() => {
    if (!selected && urlTicker) setSelected(urlTicker)
  }, [urlTicker, selected])

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Analysis</h1>
          <p className="text-muted-foreground">Candlestick charts (real data).</p>
        </div>
        <Badge variant="secondary" className="flex items-center gap-2">
          <ChartCandlestick className="h-4 w-4" />
          Candles
        </Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Tickers</CardTitle>
            <CardDescription>Select a ticker to view its candlestick chart.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Input value={query} onChange={e => setQuery(e.target.value)} placeholder="Search symbol or name" />

            {loadingTickers ? (
              <div className="text-sm text-muted-foreground">Loading...</div>
            ) : filteredTickers.length === 0 ? (
              <div className="text-sm text-muted-foreground">
                No tickers found. Run the processor to generate `data/reports/summary/ticker_summary.json` and
                `data/reports/ticker/*_trading_history.csv`.
              </div>
            ) : (
              <div className="max-h-[60vh] overflow-auto rounded-md border">
                {filteredTickers.map(t => (
                  <button
                    key={t.symbol}
                    className={`flex w-full items-center justify-between px-3 py-2 text-left text-sm hover:bg-muted ${selected === t.symbol ? 'bg-muted' : ''}`}
                    onClick={() => setSelected(t.symbol)}
                    type="button"
                  >
                    <div className="min-w-0">
                      <div className="font-mono">{t.symbol}</div>
                      {t.name ? <div className="truncate text-xs text-muted-foreground">{t.name}</div> : null}
                    </div>
                  </button>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>{selected ? `Chart: ${selected}` : 'Chart'}</CardTitle>
            <CardDescription>
              {loadingChart ? 'Loading...' : candles.length ? `${candles.length} candles` : 'No chart data yet.'}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-[70vh] w-full rounded-md border bg-background" ref={chartContainerRef} />
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
