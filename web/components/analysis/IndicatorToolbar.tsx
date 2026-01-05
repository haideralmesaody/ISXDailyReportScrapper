/**
 * Indicator Toolbar Component
 * Provides UI controls for toggling TradingView indicators and tools
 * Following CLAUDE.md React patterns with proper state management
 */

'use client'

import React, { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { ChevronDown, ChevronUp, TrendingUp, Settings } from 'lucide-react'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import type { ChartAction, ChartState, IndicatorSettings } from '@/lib/hooks/use-chart-state'
import { IndicatorSettingsPopover } from '@/components/analysis/chart/indicators/IndicatorSettingsPopover'
import {
  useMovingAverageSettings,
  useMomentumSettings,
  useSupportResistanceSettings,
  useTrendSettings,
  useVolatilitySettings,
} from '@/lib/hooks/use-indicator-settings'
import { useVWAPSettings } from '@/lib/hooks/use-vwap-settings'

type IndicatorRowConfig = {
  key: keyof IndicatorSettings
  label: string
  description: string
  settingsPopover?: {
    indicatorName: string
    settings: Record<string, any>
    fields: Array<{
      name: string
      label: string
      type: 'number' | 'color' | 'select'
      min?: number
      max?: number
      step?: number
      options?: { value: number | null; label: string }[]
      hasCustomInput?: boolean
    }>
    onUpdate: (newSettings: Record<string, any>) => void
  }
}

interface IndicatorToolbarProps {
  className?: string
  state: ChartState
  dispatch: React.Dispatch<ChartAction>
}

export function IndicatorToolbar({ className, state, dispatch }: IndicatorToolbarProps) {
  const [openCategories, setOpenCategories] = useState<Set<string>>(new Set(['trend']))

  const { settings: maSettings, updateSettings: updateMaSettings } = useMovingAverageSettings()
  const { settings: momentumSettings, updateSettings: updateMomentumSettings } = useMomentumSettings()
  const { settings: trendSettings, updateSettings: updateTrendSettings } = useTrendSettings()
  const { settings: volatilitySettings, updateSettings: updateVolatilitySettings } = useVolatilitySettings()
  const { settings: srSettings, updateSettings: updateSRSettings } = useSupportResistanceSettings()
  const { settings: vwapSettings, updateSettings: updateVWAPSettings } = useVWAPSettings()

  // Count active indicators
  const activeIndicatorsCount = React.useMemo(() => {
    if (!state.indicators) return 0

    return Object.entries(state.indicators).filter(([_, isActive]) => isActive).length
  }, [state.indicators])

  // Toggle indicator
  const handleToggleIndicator = (indicatorKey: keyof IndicatorSettings) => {
    dispatch({ type: 'TOGGLE_INDICATOR', payload: indicatorKey })
  }

  // Toggle category expansion
  const toggleCategory = (categoryId: string) => {
    const newOpenCategories = new Set(openCategories)
    if (newOpenCategories.has(categoryId)) {
      newOpenCategories.delete(categoryId)
    } else {
      newOpenCategories.add(categoryId)
    }
    setOpenCategories(newOpenCategories)
  }

  // Get indicator state (fallback to false if not initialized)
  const getIndicatorState = (key: keyof IndicatorSettings): boolean => {
    return state.indicators?.[key] || false
  }

  const indicatorCategories: Array<{
    id: string
    name: string
    description: string
    indicators: IndicatorRowConfig[]
  }> = [
    {
      id: 'trend',
      name: 'Trend',
      description: 'Moving averages & trend',
      indicators: [
        { key: 'showSMA20', label: 'SMA 20', description: '20-period Simple Moving Average' },
        { key: 'showSMA50', label: 'SMA 50', description: '50-period Simple Moving Average' },
        { key: 'showSMA200', label: 'SMA 200', description: '200-period Simple Moving Average' },
        { key: 'showEMA20', label: 'EMA 20', description: '20-period Exponential Moving Average' },
        {
          key: 'showGeneralSMA',
          label: 'SMA',
          description: 'Custom-period Simple Moving Average',
          settingsPopover: {
            indicatorName: 'SMA',
            settings: { generalSmaPeriod: maSettings.generalSmaPeriod },
            fields: [{ name: 'generalSmaPeriod', label: 'Period', type: 'number', min: 2, max: 400 }],
            onUpdate: (next) => updateMaSettings({ generalSmaPeriod: Number(next.generalSmaPeriod) || 20 }),
          },
        },
        {
          key: 'showGeneralEMA',
          label: 'EMA',
          description: 'Custom-period Exponential Moving Average',
          settingsPopover: {
            indicatorName: 'EMA',
            settings: { generalEmaPeriod: maSettings.generalEmaPeriod },
            fields: [{ name: 'generalEmaPeriod', label: 'Period', type: 'number', min: 2, max: 400 }],
            onUpdate: (next) => updateMaSettings({ generalEmaPeriod: Number(next.generalEmaPeriod) || 20 }),
          },
        },
        {
          key: 'showADX',
          label: 'ADX',
          description: 'Average Directional Index',
          settingsPopover: {
            indicatorName: 'ADX',
            settings: {
              adxPeriod: trendSettings.adxPeriod,
              adxTrendThreshold: trendSettings.adxTrendThreshold,
            },
            fields: [
              { name: 'adxPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
              { name: 'adxTrendThreshold', label: 'Trend Threshold', type: 'number', min: 0, max: 100 },
            ],
            onUpdate: (next) =>
              updateTrendSettings({
                adxPeriod: Number(next.adxPeriod) || 14,
                adxTrendThreshold: Number(next.adxTrendThreshold) || 25,
              }),
          },
        },
        {
          key: 'showIchimoku',
          label: 'Ichimoku',
          description: 'Ichimoku Cloud',
          settingsPopover: {
            indicatorName: 'Ichimoku',
            settings: {
              ichimokuTenkanPeriod: trendSettings.ichimokuTenkanPeriod,
              ichimokuKijunPeriod: trendSettings.ichimokuKijunPeriod,
              ichimokuSenkouBPeriod: trendSettings.ichimokuSenkouBPeriod,
            },
            fields: [
              { name: 'ichimokuTenkanPeriod', label: 'Tenkan', type: 'number', min: 2, max: 100 },
              { name: 'ichimokuKijunPeriod', label: 'Kijun', type: 'number', min: 2, max: 200 },
              { name: 'ichimokuSenkouBPeriod', label: 'Senkou B', type: 'number', min: 2, max: 300 },
            ],
            onUpdate: (next) =>
              updateTrendSettings({
                ichimokuTenkanPeriod: Number(next.ichimokuTenkanPeriod) || 9,
                ichimokuKijunPeriod: Number(next.ichimokuKijunPeriod) || 26,
                ichimokuSenkouBPeriod: Number(next.ichimokuSenkouBPeriod) || 52,
              }),
          },
        },
        { key: 'showParabolicSAR', label: 'Parabolic SAR', description: 'Parabolic stop-and-reverse' },
      ],
    },
    {
      id: 'momentum',
      name: 'Momentum',
      description: 'Oscillators & momentum',
      indicators: [
        {
          key: 'showRSI',
          label: 'RSI',
          description: 'Relative Strength Index',
          settingsPopover: {
            indicatorName: 'RSI',
            settings: {
              rsiPeriod: momentumSettings.rsiPeriod,
              rsiOverbought: momentumSettings.rsiOverbought,
              rsiOversold: momentumSettings.rsiOversold,
            },
            fields: [
              { name: 'rsiPeriod', label: 'Period', type: 'number', min: 2, max: 100 },
              { name: 'rsiOverbought', label: 'Overbought', type: 'number', min: 50, max: 100 },
              { name: 'rsiOversold', label: 'Oversold', type: 'number', min: 0, max: 50 },
            ],
            onUpdate: (next) =>
              updateMomentumSettings({
                rsiPeriod: Number(next.rsiPeriod) || 14,
                rsiOverbought: Number(next.rsiOverbought) || 70,
                rsiOversold: Number(next.rsiOversold) || 30,
              }),
          },
        },
        {
          key: 'showMACD',
          label: 'MACD',
          description: 'Moving Average Convergence Divergence',
          settingsPopover: {
            indicatorName: 'MACD',
            settings: {
              macdFastPeriod: momentumSettings.macdFastPeriod,
              macdSlowPeriod: momentumSettings.macdSlowPeriod,
              macdSignalPeriod: momentumSettings.macdSignalPeriod,
            },
            fields: [
              { name: 'macdFastPeriod', label: 'Fast Period', type: 'number', min: 2, max: 50 },
              { name: 'macdSlowPeriod', label: 'Slow Period', type: 'number', min: 10, max: 200 },
              { name: 'macdSignalPeriod', label: 'Signal Period', type: 'number', min: 2, max: 50 },
            ],
            onUpdate: (next) =>
              updateMomentumSettings({
                macdFastPeriod: Number(next.macdFastPeriod) || 12,
                macdSlowPeriod: Number(next.macdSlowPeriod) || 26,
                macdSignalPeriod: Number(next.macdSignalPeriod) || 9,
              }),
          },
        },
        {
          key: 'showStochastic',
          label: 'Stochastic',
          description: 'Stochastic Oscillator',
          settingsPopover: {
            indicatorName: 'Stochastic',
            settings: {
              stochasticKPeriod: momentumSettings.stochasticKPeriod,
              stochasticDPeriod: momentumSettings.stochasticDPeriod,
              stochasticOverbought: momentumSettings.stochasticOverbought,
              stochasticOversold: momentumSettings.stochasticOversold,
            },
            fields: [
              { name: 'stochasticKPeriod', label: '%K Period', type: 'number', min: 2, max: 100 },
              { name: 'stochasticDPeriod', label: '%D Period', type: 'number', min: 2, max: 50 },
              { name: 'stochasticOverbought', label: 'Overbought', type: 'number', min: 50, max: 100 },
              { name: 'stochasticOversold', label: 'Oversold', type: 'number', min: 0, max: 50 },
            ],
            onUpdate: (next) =>
              updateMomentumSettings({
                stochasticKPeriod: Number(next.stochasticKPeriod) || 14,
                stochasticDPeriod: Number(next.stochasticDPeriod) || 3,
                stochasticOverbought: Number(next.stochasticOverbought) || 80,
                stochasticOversold: Number(next.stochasticOversold) || 20,
              }),
          },
        },
        {
          key: 'showGeneralStochastic',
          label: 'Stochastic (Custom)',
          description: 'Custom %K/%D periods',
          settingsPopover: {
            indicatorName: 'General Stochastic',
            settings: {
              generalStochasticKPeriod: momentumSettings.generalStochasticKPeriod,
              generalStochasticDPeriod: momentumSettings.generalStochasticDPeriod,
            },
            fields: [
              { name: 'generalStochasticKPeriod', label: '%K Period', type: 'number', min: 2, max: 100 },
              { name: 'generalStochasticDPeriod', label: '%D Period', type: 'number', min: 2, max: 50 },
            ],
            onUpdate: (next) =>
              updateMomentumSettings({
                generalStochasticKPeriod: Number(next.generalStochasticKPeriod) || 14,
                generalStochasticDPeriod: Number(next.generalStochasticDPeriod) || 3,
              }),
          },
        },
        {
          key: 'showCCI',
          label: 'CCI',
          description: 'Commodity Channel Index',
          settingsPopover: {
            indicatorName: 'CCI',
            settings: {
              cciPeriod: momentumSettings.cciPeriod,
              cciOverbought: momentumSettings.cciOverbought,
              cciOversold: momentumSettings.cciOversold,
            },
            fields: [
              { name: 'cciPeriod', label: 'Period', type: 'number', min: 2, max: 200 },
              { name: 'cciOverbought', label: 'Overbought', type: 'number', min: 0, max: 500 },
              { name: 'cciOversold', label: 'Oversold', type: 'number', min: -500, max: 0 },
            ],
            onUpdate: (next) =>
              updateMomentumSettings({
                cciPeriod: Number(next.cciPeriod) || 20,
                cciOverbought: Number(next.cciOverbought) || 100,
                cciOversold: Number(next.cciOversold) || -100,
              }),
          },
        },
        {
          key: 'showWilliamsR',
          label: 'Williams %R',
          description: 'Williams Percent Range',
          settingsPopover: {
            indicatorName: 'Williams %R',
            settings: {
              williamsRPeriod: momentumSettings.williamsRPeriod,
              williamsROverbought: momentumSettings.williamsROverbought,
              williamsROversold: momentumSettings.williamsROversold,
            },
            fields: [
              { name: 'williamsRPeriod', label: 'Period', type: 'number', min: 2, max: 200 },
              { name: 'williamsROverbought', label: 'Overbought', type: 'number', min: -100, max: 0 },
              { name: 'williamsROversold', label: 'Oversold', type: 'number', min: -100, max: 0 },
            ],
            onUpdate: (next) =>
              updateMomentumSettings({
                williamsRPeriod: Number(next.williamsRPeriod) || 14,
                williamsROverbought: Number(next.williamsROverbought) || -20,
                williamsROversold: Number(next.williamsROversold) || -80,
              }),
          },
        },
        {
          key: 'showMOM',
          label: 'Momentum',
          description: 'Momentum Indicator',
          settingsPopover: {
            indicatorName: 'Momentum',
            settings: { momentumPeriod: momentumSettings.momentumPeriod },
            fields: [{ name: 'momentumPeriod', label: 'Period', type: 'number', min: 2, max: 200 }],
            onUpdate: (next) => updateMomentumSettings({ momentumPeriod: Number(next.momentumPeriod) || 10 }),
          },
        },
        {
          key: 'showROC',
          label: 'ROC',
          description: 'Rate of Change',
          settingsPopover: {
            indicatorName: 'ROC',
            settings: { rocPeriod: momentumSettings.rocPeriod },
            fields: [{ name: 'rocPeriod', label: 'Period', type: 'number', min: 2, max: 200 }],
            onUpdate: (next) => updateMomentumSettings({ rocPeriod: Number(next.rocPeriod) || 12 }),
          },
        },
        { key: 'showMFI', label: 'MFI', description: 'Money Flow Index' },
      ],
    },
    {
      id: 'volatility',
      name: 'Volatility',
      description: 'Bands & ranges',
      indicators: [
        {
          key: 'showATR',
          label: 'ATR',
          description: 'Average True Range',
          settingsPopover: {
            indicatorName: 'ATR',
            settings: { atrPeriod: volatilitySettings.atrPeriod },
            fields: [{ name: 'atrPeriod', label: 'Period', type: 'number', min: 2, max: 200 }],
            onUpdate: (next) => updateVolatilitySettings({ atrPeriod: Number(next.atrPeriod) || 14 }),
          },
        },
        {
          key: 'showBollingerBands',
          label: 'Bollinger Bands',
          description: 'Bollinger Bands',
          settingsPopover: {
            indicatorName: 'Bollinger Bands',
            settings: {
              bollingerPeriod: volatilitySettings.bollingerPeriod,
              bollingerStdDev: volatilitySettings.bollingerStdDev,
            },
            fields: [
              { name: 'bollingerPeriod', label: 'Period', type: 'number', min: 2, max: 200 },
              { name: 'bollingerStdDev', label: 'Std Dev', type: 'number', min: 0.1, max: 10, step: 0.1 },
            ],
            onUpdate: (next) =>
              updateVolatilitySettings({
                bollingerPeriod: Number(next.bollingerPeriod) || 20,
                bollingerStdDev: Number(next.bollingerStdDev) || 2,
              }),
          },
        },
        {
          key: 'showGeneralBollinger',
          label: 'Bollinger (Custom)',
          description: 'Custom-period Bollinger Bands',
          settingsPopover: {
            indicatorName: 'General Bollinger Bands',
            settings: {
              generalBollingerPeriod: volatilitySettings.generalBollingerPeriod,
              generalBollingerStdDev: volatilitySettings.generalBollingerStdDev,
            },
            fields: [
              { name: 'generalBollingerPeriod', label: 'Period', type: 'number', min: 2, max: 200 },
              { name: 'generalBollingerStdDev', label: 'Std Dev', type: 'number', min: 0.1, max: 10, step: 0.1 },
            ],
            onUpdate: (next) =>
              updateVolatilitySettings({
                generalBollingerPeriod: Number(next.generalBollingerPeriod) || 20,
                generalBollingerStdDev: Number(next.generalBollingerStdDev) || 2,
              }),
          },
        },
        {
          key: 'showKeltner',
          label: 'Keltner Channels',
          description: 'ATR-based channel bands',
          settingsPopover: {
            indicatorName: 'Keltner Channels',
            settings: {
              keltnerPeriod: volatilitySettings.keltnerPeriod,
              keltnerMultiplier: volatilitySettings.keltnerMultiplier,
            },
            fields: [
              { name: 'keltnerPeriod', label: 'Period', type: 'number', min: 2, max: 200 },
              { name: 'keltnerMultiplier', label: 'Multiplier', type: 'number', min: 0.1, max: 10, step: 0.1 },
            ],
            onUpdate: (next) =>
              updateVolatilitySettings({
                keltnerPeriod: Number(next.keltnerPeriod) || 20,
                keltnerMultiplier: Number(next.keltnerMultiplier) || 2,
              }),
          },
        },
      ],
    },
    {
      id: 'volume',
      name: 'Volume',
      description: 'Volume-based',
      indicators: [
        { key: 'showVolume', label: 'Volume', description: 'Trading volume histogram' },
        { key: 'showOBV', label: 'OBV', description: 'On-Balance Volume' },
        {
          key: 'showVWAP',
          label: 'VWAP',
          description: 'Anchored Volume Weighted Average Price',
          settingsPopover: {
            indicatorName: 'VWAP',
            settings: { anchorDays: vwapSettings.anchorDays },
            fields: [
              {
                name: 'anchorDays',
                label: 'Anchor',
                type: 'select',
                options: [
                  { value: null, label: 'Chart start' },
                  { value: 10, label: 'Last 10 bars' },
                  { value: 20, label: 'Last 20 bars' },
                  { value: 30, label: 'Last 30 bars' },
                  { value: 60, label: 'Last 60 bars' },
                  { value: 90, label: 'Last 90 bars' },
                ],
              },
            ],
            onUpdate: (next) => updateVWAPSettings({ anchorDays: next.anchorDays ?? null }),
          },
        },
      ],
    },
    {
      id: 'levels',
      name: 'Levels',
      description: 'Support/resistance',
      indicators: [
        { key: 'showSupport', label: 'Support', description: 'Support levels' },
        { key: 'showResistance', label: 'Resistance', description: 'Resistance levels' },
        {
          key: 'showFibonacci',
          label: 'Fibonacci',
          description: 'Fibonacci Retracement',
          settingsPopover: {
            indicatorName: 'Fibonacci Retracement',
            settings: { fibonacciLookback: srSettings.fibonacciLookback },
            fields: [{ name: 'fibonacciLookback', label: 'Lookback', type: 'number', min: 10, max: 400 }],
            onUpdate: (next) => updateSRSettings({ fibonacciLookback: Number(next.fibonacciLookback) || 50 }),
          },
        },
        {
          key: 'showDonchian',
          label: 'Donchian Channels',
          description: 'High/low channel bands',
          settingsPopover: {
            indicatorName: 'Donchian Channels',
            settings: { donchianPeriod: srSettings.donchianPeriod },
            fields: [{ name: 'donchianPeriod', label: 'Period', type: 'number', min: 2, max: 400 }],
            onUpdate: (next) => updateSRSettings({ donchianPeriod: Number(next.donchianPeriod) || 20 }),
          },
        },
      ],
    },
  ]

  return (
    <Card className={`${className}`}>
      <CardContent className="p-4">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
            <h3 className="font-semibold">Indicators & Tools</h3>
            {activeIndicatorsCount > 0 && (
              <Badge variant="secondary" className="text-xs">
                {activeIndicatorsCount} active
              </Badge>
            )}
          </div>
        </div>

        <div className="space-y-2">
          {indicatorCategories.map((category) => (
            <Collapsible
              key={category.id}
              open={openCategories.has(category.id)}
              onOpenChange={() => toggleCategory(category.id)}
            >
              <CollapsibleTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full justify-between px-2 h-8"
                >
                  <span className="text-sm font-medium">{category.name}</span>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">{category.description}</span>
                    {openCategories.has(category.id) ? (
                      <ChevronUp className="h-4 w-4" />
                    ) : (
                      <ChevronDown className="h-4 w-4" />
                    )}
                  </div>
                </Button>
              </CollapsibleTrigger>

              <CollapsibleContent className="space-y-2 pt-2">
                <div className="grid grid-cols-1 gap-2 px-2">
                  {category.indicators.map((indicator) => (
                    <div
                      key={indicator.key}
                      className="flex items-center justify-between py-1 px-2 rounded-sm hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center gap-2 flex-1">
                        <Switch
                          id={indicator.key}
                          checked={getIndicatorState(indicator.key as keyof IndicatorSettings)}
                          onCheckedChange={() => handleToggleIndicator(indicator.key as keyof IndicatorSettings)}
                        />
                        <div className="flex flex-col">
                          <Label
                            htmlFor={indicator.key}
                            className="text-sm font-medium cursor-pointer"
                          >
                            {indicator.label}
                          </Label>
                          <span className="text-xs text-muted-foreground">
                            {indicator.description}
                          </span>
                        </div>
                      </div>

                      {indicator.settingsPopover &&
                        getIndicatorState(indicator.key as keyof IndicatorSettings) && (
                          <IndicatorSettingsPopover
                            indicatorName={indicator.settingsPopover.indicatorName}
                            settings={indicator.settingsPopover.settings}
                            fields={indicator.settingsPopover.fields}
                            onUpdate={indicator.settingsPopover.onUpdate}
                            triggerButton={
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-7 w-7 rounded-sm"
                                onPointerDown={(e) => e.stopPropagation()}
                                onClick={(e) => e.stopPropagation()}
                                aria-label={`${indicator.settingsPopover.indicatorName} settings`}
                              >
                                <Settings className="h-3.5 w-3.5" />
                              </Button>
                            }
                          />
                        )}
                    </div>
                  ))}
                </div>
              </CollapsibleContent>
            </Collapsible>
          ))}
        </div>

        {activeIndicatorsCount > 0 && (
          <>
            <Separator className="my-4" />
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                {activeIndicatorsCount} indicator{activeIndicatorsCount !== 1 ? 's' : ''} active
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  // Clear all indicators
                  Object.keys(state.indicators).forEach(key => {
                    if (state.indicators[key as keyof IndicatorSettings]) {
                      dispatch({ type: 'TOGGLE_INDICATOR', payload: key as keyof IndicatorSettings })
                    }
                  })
                }}
              >
                Clear All
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
