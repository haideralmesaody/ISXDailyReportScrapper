'use client'

import React from 'react'
import { Card } from '@/components/ui/card'
import {
  formatIQD,
  formatVolume,
  formatNumber,
  formatPercentage,
  formatDate,
  type EnhancedTooltipData,
  type ComparisonData,
  type TooltipIndicatorValue,
} from '@/lib/utils/tooltip-helpers'
import { TrendingUp, TrendingDown, Minus, Calendar, BarChart3, DollarSign } from 'lucide-react'

interface ChartTooltipProps {
  data: EnhancedTooltipData | null
  position: { x: number; y: number }
  visible: boolean
  containerWidth: number
  containerHeight: number
}

export const ChartTooltip = React.memo(function ChartTooltip({
  data,
  position,
  visible,
  containerWidth,
  containerHeight,
}: ChartTooltipProps) {
  if (!visible || !data) return null

  // Calculate dynamic tooltip size based on content
  const calculateTooltipSize = () => {
    const baseHeight = 180 // OHLCV + header
    const indicatorHeight = (data.indicators?.size || 0) * 24 // 24px per indicator
    const comparisonCount = [data.vsYesterday, data.vsWeekAgo, data.vsMonthAgo]
      .filter(Boolean).length
    const comparisonHeight = comparisonCount > 0 ? 60 + (comparisonCount * 32) : 0

    return {
      width: 320,
      height: Math.min(
        baseHeight + indicatorHeight + comparisonHeight,
        containerHeight - 20 // Keep 20px margin from edges
      )
    }
  }

  const { width: tooltipWidth, height: tooltipHeight } = calculateTooltipSize()
  const padding = 10

  let left = position.x + padding
  let top = position.y + padding

  // Adjust horizontal position
  if (left + tooltipWidth > containerWidth) {
    left = position.x - tooltipWidth - padding
  }

  // Adjust vertical position
  if (top + tooltipHeight > containerHeight) {
    top = Math.max(padding, containerHeight - tooltipHeight - padding)
  }

  // Ensure minimum position
  left = Math.max(padding, left)
  top = Math.max(padding, top)

  const changeData = formatPercentage(data.changePercent)

  return (
    <Card
      className="absolute z-50 p-4 bg-background/95 backdrop-blur-sm border shadow-lg w-80"
      style={{
        left: `${left}px`,
        top: `${top}px`,
        pointerEvents: 'none',
      }}
    >
      {/* Header */}
      <div className="mb-3 pb-3 border-b">
        <div className="flex items-center justify-between mb-1">
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-muted-foreground" />
            <span className="font-medium">{formatDate(data.date)}</span>
          </div>
          <div className={`flex items-center gap-1 ${changeData.color}`}>
            {data.changePercent > 0 ? (
              <TrendingUp className="h-4 w-4" />
            ) : data.changePercent < 0 ? (
              <TrendingDown className="h-4 w-4" />
            ) : (
              <Minus className="h-4 w-4" />
            )}
            <span className="font-semibold">{changeData.text}</span>
          </div>
        </div>
        <div className="text-sm text-muted-foreground">
          Daily Change: {formatIQD(data.change)}
        </div>
      </div>

      {/* OHLCV Data - Professional hierarchy: Close price emphasized */}
      <div className="space-y-2 mb-3">
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground text-xs">Open</span>
            <span className="font-medium text-sm">{formatIQD(data.open)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground text-xs">High</span>
            <span className="font-medium text-sm">{formatIQD(data.high)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground text-xs">Low</span>
            <span className="font-medium text-sm">{formatIQD(data.low)}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground text-xs font-semibold">Close</span>
            <span className="font-bold text-base">{formatIQD(data.close)}</span>
          </div>
        </div>

        <div className="pt-2 border-t">
          <div className="flex items-center gap-2 mb-2">
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm text-muted-foreground">Volume:</span>
            <span className="font-medium">{formatVolume(data.volume)}</span>
          </div>
          {data.value && (
            <div className="flex items-center gap-2">
              <DollarSign className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">Value:</span>
              <span className="font-medium">{formatIQD(data.value)}</span>
            </div>
          )}
          {data.trades && (
            <div className="text-sm text-muted-foreground mt-1">
              Trades: {formatVolume(data.trades)}
            </div>
          )}
        </div>
      </div>

      {/* Dynamic Indicators Section - Grouped by category */}
      {data.indicators && data.indicators.size > 0 && (
        <div className="py-2 border-t text-sm space-y-3">
          {/* Moving Averages Group */}
          <IndicatorGroup
            title="Moving Averages"
            indicatorIds={['sma20', 'sma50', 'sma200', 'ema20']}
            indicators={data.indicators}
          />

          {/* Momentum Group */}
          <IndicatorGroup
            title="Momentum"
            indicatorIds={['rsi', 'macdLine', 'macdSignal', 'macdHistogram']}
            indicators={data.indicators}
          />

          {/* Support/Resistance Group */}
          <IndicatorGroup
            title="Levels"
            indicatorIds={['support', 'resistance']}
            indicators={data.indicators}
          />

          {/* Volatility Indicators Group */}
          <IndicatorGroup
            title="Volatility"
            indicatorIds={['bollingerUpper', 'bollingerMiddle', 'bollingerLower', 'atr']}
            indicators={data.indicators}
          />

          {/* Momentum Oscillators Group */}
          <IndicatorGroup
            title="Oscillators"
            indicatorIds={['stochasticK', 'stochasticD', 'cci', 'williamsR']}
            indicators={data.indicators}
          />

          {/* Trend Indicators Group */}
          <IndicatorGroup
            title="Trend"
            indicatorIds={['adx', 'plusDI', 'minusDI', 'ichimokuTenkan', 'ichimokuKijun', 'ichimokuSenkouA', 'ichimokuSenkouB', 'ichimokuChikou']}
            indicators={data.indicators}
          />
        </div>
      )}
      
      {/* Legacy Moving Averages - Keep for backward compatibility */}
      {!data.indicators && (data.priceMA20 || data.volumeMA20) && (
        <div className="py-2 border-t border-b text-sm">
          <div className="font-medium mb-1">Moving Averages (20-day)</div>
          {data.priceMA20 && (
            <div className="text-muted-foreground">
              Price MA: {formatIQD(data.priceMA20)}
            </div>
          )}
          {data.volumeMA20 && (
            <div className="text-muted-foreground">
              Volume MA: {formatVolume(data.volumeMA20)}
            </div>
          )}
        </div>
      )}

      {/* Historical Comparisons - Only show if data exists */}
      {(data.vsYesterday || data.vsWeekAgo || data.vsMonthAgo) && (
        <div className="pt-3 space-y-2 border-t">
          <div className="font-medium text-xs uppercase tracking-wide text-muted-foreground mb-2">
            Historical Comparison
          </div>

          {data.vsYesterday && (
            <ComparisonRow label="vs Yesterday" data={data.vsYesterday} />
          )}

          {data.vsWeekAgo && (
            <ComparisonRow label="vs Week Ago" data={data.vsWeekAgo} />
          )}

          {data.vsMonthAgo && (
            <ComparisonRow label="vs Month Ago" data={data.vsMonthAgo} />
          )}
        </div>
      )}
    </Card>
  )
})

// Helper component for grouped indicator display
function IndicatorGroup({
  title,
  indicatorIds,
  indicators,
}: {
  title: string
  indicatorIds: string[]
  indicators: Map<string, TooltipIndicatorValue>
}) {
  const groupIndicators = indicatorIds.filter(id => indicators.has(id))
  if (groupIndicators.length === 0) return null

  return (
    <div>
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground/60 mb-1.5">
        {title}
      </div>
      <div className="space-y-1">
        {groupIndicators.map(id => {
          const indicator = indicators.get(id)!
          return (
            <div key={id} className="flex justify-between items-center">
              <div className="flex items-center gap-1.5">
                <div
                  className="w-2.5 h-0.5 rounded-full"
                  style={{ backgroundColor: indicator.color }}
                />
                <span className="text-xs text-muted-foreground">
                  {indicator.displayName}
                </span>
              </div>
              <span className="font-medium text-xs tabular-nums">
                {indicator.formatted}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function ComparisonRow({ label, data }: { label: string; data: ComparisonData }) {
  const changeData = formatPercentage(data.changePercent)

  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}:</span>
      <div className="flex items-center gap-2">
        <span className="text-muted-foreground">{formatIQD(data.price)}</span>
        <span className={`font-medium ${changeData.color}`}>
          {changeData.text}
        </span>
      </div>
    </div>
  )
}