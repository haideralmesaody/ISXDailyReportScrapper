/**
 * MarketTreemap - D3.js treemap visualization for market overview
 * Shows tickers sized by traded value, colored by price change%
 * Following CLAUDE.md component standards
 */

'use client'

import { useRef, useEffect, useMemo, useState } from 'react'
import * as d3 from 'd3'
import { TickerData } from '@/types/market'
import { getChangeColor, formatCurrency, formatPercent } from '@/lib/utils/market-helpers'
import { useHydration } from '@/lib/hooks'

interface TooltipState {
  visible: boolean
  x: number
  y: number
  ticker: TickerData | null
}

/**
 * Generate SVG path for sparkline visualization
 * Uses min-max normalization to fit prices into available space
 */
function generateSparklinePath(prices: number[], width: number, height: number): string {
  if (!prices || prices.length < 2) return ''

  const min = Math.min(...prices)
  const max = Math.max(...prices)
  const range = max - min || 1 // Avoid division by zero

  // Generate normalized points
  const points = prices.map((price, i) => {
    const x = (i / (prices.length - 1)) * width
    const y = height - ((price - min) / range) * height
    return `${x.toFixed(2)},${y.toFixed(2)}`
  })

  // Create SVG path: M (move to first) L (line to rest)
  return `M ${points.join(' L ')}`
}

/**
 * Determine sparkline color based on price trend
 * Uses white/light colors for high contrast visibility on ALL backgrounds
 * Following Bloomberg/TradingView professional standards
 */
function getSparklineColor(prices: number[]): string {
  if (!prices || prices.length < 2) return 'rgba(255, 255, 255, 0.5)' // Dim white (no data)

  const firstPrice = prices[0]
  const lastPrice = prices.at(-1)
  if (firstPrice === undefined || lastPrice === undefined) return 'rgba(255, 255, 255, 0.5)'

  if (lastPrice > firstPrice) return 'rgba(255, 255, 255, 0.9)'  // Bright white (uptrend)
  if (lastPrice < firstPrice) return 'rgba(255, 255, 255, 0.7)' // Medium white (downtrend)
  return 'rgba(255, 255, 255, 0.5)' // Dim white (unchanged)
}

interface MarketTreemapProps {
  tickers: TickerData[]
  onTickerClick?: (ticker: TickerData) => void
  width?: number
  height?: number
  minCellSize?: number // Skip cells smaller than this (px²)
}

export function MarketTreemap({
  tickers,
  onTickerClick,
  width = 1200,
  height = 600,
  minCellSize = 400
}: MarketTreemapProps) {
  const isHydrated = useHydration()
  const svgRef = useRef<SVGSVGElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [tooltip, setTooltip] = useState<TooltipState>({
    visible: false,
    x: 0,
    y: 0,
    ticker: null
  })

  // Prepare hierarchy data for D3 treemap
  const root = useMemo(() => {
    // Add loading state guard to prevent confusing "0 tickers" message
    if (!tickers || tickers.length === 0) {
      console.log('[MarketTreemap] Waiting for market data...')
      return null
    }

    console.log('[MarketTreemap] 📊 Processing', tickers.length, 'tickers for treemap')

    // Debug: Log first few tickers to understand data structure
    console.log('[MarketTreemap] Sample ticker data:', tickers.slice(0, 3))

    const data = {
      name: 'Market',
      children: tickers.map(ticker => ({
        name: ticker.symbol,
        value: ticker.tradedValue,
        ticker
      }))
    }

    return d3.hierarchy(data)
      .sum(d => (d as any).value || 0)
      .sort((a, b) => (b.value || 0) - (a.value || 0))
  }, [tickers])

  useEffect(() => {
    // Only run D3 rendering on client side after hydration
    if (!isHydrated || !svgRef.current || !root) return

    console.log('[MarketTreemap] Rendering with dimensions:', width, 'x', height)

    const svg = d3.select(svgRef.current)
    svg.selectAll('*').remove() // Clear previous render

    // Create treemap layout
    const treemap = d3.treemap<any>()
      .size([width, height])
      .paddingInner(2)
      .paddingOuter(4)
      .round(true)

    const nodes = treemap(root).leaves()
    console.log('[MarketTreemap] D3 calculated', nodes.length, 'ticker cells')

    // Filter out tiny cells (below minimum size threshold)
    const visibleNodes = nodes.filter(d => {
      const area = (d.x1 - d.x0) * (d.y1 - d.y0)
      return area >= minCellSize
    })

    // Create cell groups
    const cell = svg.selectAll('g')
      .data(visibleNodes)
      .join('g')
      .attr('transform', d => `translate(${d.x0},${d.y0})`)
      .attr('cursor', 'pointer')
      .attr('opacity', 1)
      .on('click', (event, d) => {
        void event
        if (onTickerClick) onTickerClick((d.data as any).ticker)
      })
      .on('mouseenter', (event, d) => {
        // Highlight cell on hover
        d3.select(event.currentTarget).attr('opacity', 0.85)

        // Show tooltip
        const svgRect = svgRef.current!.getBoundingClientRect()
        setTooltip({
          visible: true,
          x: event.clientX - svgRect.left,
          y: event.clientY - svgRect.top,
          ticker: (d.data as any).ticker
        })
      })
      .on('mousemove', (event) => {
        // Update tooltip position as mouse moves
        const svgRect = svgRef.current!.getBoundingClientRect()
        setTooltip(prev => ({
          ...prev,
          x: event.clientX - svgRect.left,
          y: event.clientY - svgRect.top
        }))
      })
      .on('mouseleave', (event) => {
        // Reset cell opacity
        d3.select(event.currentTarget).attr('opacity', 1)

        // Hide tooltip
        setTooltip({ visible: false, x: 0, y: 0, ticker: null })
      })

    // Background rect with color based on price change
    cell.append('rect')
      .attr('width', d => d.x1 - d.x0)
      .attr('height', d => d.y1 - d.y0)
      .attr('fill', d => getChangeColor((d.data as any).ticker.changePercent))
      .attr('stroke', '#ffffff')
      .attr('stroke-width', 3)
      .attr('rx', 4)
      .style('transition', 'all 0.2s ease')

    // Symbol text (large, bold) with text shadow for better readability
    cell.append('text')
      .attr('x', 6)
      .attr('y', 24)
      .attr('font-size', d => {
        const w = d.x1 - d.x0
        return Math.min(w / 2.5, 18) // Larger font, better scaling
      })
      .attr('font-weight', '700')
      .attr('fill', '#ffffff')
      .attr('filter', 'drop-shadow(0px 1px 2px rgba(0,0,0,0.8))')
      .style('text-shadow', '0px 1px 2px rgba(0,0,0,0.8)')
      .text(d => (d.data as any).ticker.symbol)

    // Company name (small, lighter - only show if cell is wide enough)
    cell.append('text')
      .attr('x', 4)
      .attr('y', 40)
      .attr('font-size', 11)
      .attr('fill', '#f0f0f0')
      .text(d => {
        const cellWidth = d.x1 - d.x0
        const companyName = (d.data as any).ticker.name

        if (cellWidth < 100) return '' // Hide on small cells
        if (cellWidth < 180 && companyName.length > 15) {
          return companyName.substring(0, 15) + '...' // Truncate on medium cells
        }
        return companyName // Show full name on large cells
      })

    // Close price (medium-large, prominent)
    cell.append('text')
      .attr('x', 4)
      .attr('y', 60)
      .attr('font-size', 14)
      .attr('font-weight', '600')
      .attr('fill', '#ffffff')
      .text(d => {
        const cellWidth = d.x1 - d.x0
        if (cellWidth < 120) return '' // Hide on small cells
        const price = (d.data as any).ticker.price
        return `IQD ${price.toFixed(2)}`
      })

    // Change percentage (medium)
    cell.append('text')
      .attr('x', 4)
      .attr('y', 80)
      .attr('font-size', 14)
      .attr('fill', '#ffffff')
      .text(d => formatPercent((d.data as any).ticker.changePercent))

    // Sparkline (7-day price trend) - only show if cell is wide enough
    cell.each(function(d) {
      const cellWidth = d.x1 - d.x0
      if (cellWidth < 150) return // Skip sparklines on narrow cells

      const ticker = (d.data as any).ticker
      const prices = ticker.priceHistory

      if (!prices || prices.length < 2) return // No history data

      const sparklineWidth = 40
      const sparklineHeight = 18
      const sparklineX = 4
      const sparklineY = 95

      // Create SVG group for sparkline
      const sparklineGroup = d3.select(this).append('g')
        .attr('transform', `translate(${sparklineX},${sparklineY})`)

      // Add sparkline path
      const path = generateSparklinePath(prices, sparklineWidth, sparklineHeight)
      const color = getSparklineColor(prices)

      sparklineGroup.append('path')
        .attr('d', path)
        .attr('fill', 'none')
        .attr('stroke', color)
        .attr('stroke-width', 1.5)
    })

    // Traded value (small, lighter color)
    cell.append('text')
      .attr('x', 4)
      .attr('y', 125)
      .attr('font-size', 12)
      .attr('fill', '#e5e5e5')
      .text(d => formatCurrency((d.data as any).ticker.tradedValue))

  }, [isHydrated, root, width, height, minCellSize, onTickerClick])

  // Show loading state during hydration
  if (!isHydrated) {
    return (
      <div className="flex items-center justify-center h-96 text-muted-foreground">
        <div className="animate-pulse">Loading market visualization...</div>
      </div>
    )
  }

  if (!root || tickers.length === 0) {
    return (
      <div className="flex items-center justify-center h-96 text-muted-foreground">
        No market data available
      </div>
    )
  }

  return (
    <div ref={containerRef} className="relative">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${width} ${height}`}
        preserveAspectRatio="none"
        role="img"
        aria-label="Market treemap showing traded value by ticker"
        className="border border-border rounded-lg w-full"
        style={{ height: `${height}px` }}
      />

      {/* Interactive Tooltip */}
      {tooltip.visible && tooltip.ticker && (
        <div
          className="absolute z-50 pointer-events-none"
          style={{
            left: `${tooltip.x + 16}px`,
            top: `${tooltip.y + 16}px`,
            transform: tooltip.x > width / 2 ? 'translateX(-100%)' : 'translateX(0)',
          }}
        >
          <div className="bg-popover text-popover-foreground border border-border rounded-lg shadow-2xl p-4 max-w-xs">
            <div className="space-y-2">
              {/* Ticker Symbol + Name */}
              <div className="border-b border-border pb-2">
                <div className="font-bold text-lg">{tooltip.ticker.symbol}</div>
                <div className="text-sm text-muted-foreground">{tooltip.ticker.name}</div>
              </div>

              {/* Price Info */}
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <div className="text-muted-foreground text-xs">Price</div>
                  <div className="font-semibold">IQD {tooltip.ticker.price.toFixed(2)}</div>
                </div>
                <div>
                  <div className="text-muted-foreground text-xs">Change</div>
                  <div className={`font-semibold ${
                    tooltip.ticker.changePercent > 0 ? 'text-green-500' :
                    tooltip.ticker.changePercent < 0 ? 'text-red-500' :
                    'text-muted-foreground'
                  }`}>
                    {formatPercent(tooltip.ticker.changePercent)}
                  </div>
                </div>
              </div>

              {/* Trading Stats */}
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <div className="text-muted-foreground text-xs">Volume</div>
                  <div className="font-medium">{tooltip.ticker.volume.toLocaleString()}</div>
                </div>
                <div>
                  <div className="text-muted-foreground text-xs">Trades</div>
                  <div className="font-medium">{tooltip.ticker.trades.toLocaleString()}</div>
                </div>
              </div>

              {/* Traded Value */}
              <div className="pt-2 border-t border-border">
                <div className="text-muted-foreground text-xs">Traded Value</div>
                <div className="font-bold text-base">{formatCurrency(tooltip.ticker.tradedValue)}</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
