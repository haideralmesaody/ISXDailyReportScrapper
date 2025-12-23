'use client'

/**
 * Indicator parameter settings (periods/thresholds) used by indicator components.
 *
 * Important: these settings must be shared + persistent across the app.
 * Previous implementations used per-component `useState`, which caused:
 * - Settings not shared between indicators (e.g., RSI + MACD using separate copies)
 * - Inconsistent key names vs indicator components (e.g., `macdFast` vs `macdFastPeriod`)
 * - Settings resetting on refresh
 */

import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

export interface MovingAverageSettingsState {
  generalSmaPeriod: number
  generalEmaPeriod: number
}

export interface MomentumSettingsState {
  rsiPeriod: number
  rsiOverbought: number
  rsiOversold: number

  macdFastPeriod: number
  macdSlowPeriod: number
  macdSignalPeriod: number

  stochasticKPeriod: number
  stochasticDPeriod: number
  stochasticOverbought: number
  stochasticOversold: number

  generalStochasticKPeriod: number
  generalStochasticDPeriod: number

  cciPeriod: number
  cciOverbought: number
  cciOversold: number

  williamsRPeriod: number
  williamsROverbought: number
  williamsROversold: number

  momentumPeriod: number
  rocPeriod: number
}

export interface TrendSettingsState {
  adxPeriod: number
  adxTrendThreshold: number
  ichimokuTenkanPeriod: number
  ichimokuKijunPeriod: number
  ichimokuSenkouBPeriod: number
}

export interface VolatilitySettingsState {
  atrPeriod: number

  bollingerPeriod: number
  bollingerStdDev: number

  generalBollingerPeriod: number
  generalBollingerStdDev: number

  keltnerPeriod: number
  keltnerMultiplier: number
}

export interface SupportResistanceSettingsState {
  donchianPeriod: number
  fibonacciLookback: number
}

type IndicatorSettingsStoreState = {
  schemaVersion: number
  movingAverage: MovingAverageSettingsState
  momentum: MomentumSettingsState
  trend: TrendSettingsState
  volatility: VolatilitySettingsState
  supportResistance: SupportResistanceSettingsState

  updateMovingAverage: (patch: Partial<MovingAverageSettingsState>) => void
  updateMomentum: (patch: Partial<MomentumSettingsState>) => void
  updateTrend: (patch: Partial<TrendSettingsState>) => void
  updateVolatility: (patch: Partial<VolatilitySettingsState>) => void
  updateSupportResistance: (patch: Partial<SupportResistanceSettingsState>) => void
  resetAll: () => void
}

const SCHEMA_VERSION = 1

const DEFAULT_STATE: Omit<IndicatorSettingsStoreState, 'updateMovingAverage' | 'updateMomentum' | 'updateTrend' | 'updateVolatility' | 'updateSupportResistance' | 'resetAll'> =
  {
    schemaVersion: SCHEMA_VERSION,
    movingAverage: {
      generalSmaPeriod: 20,
      generalEmaPeriod: 20,
    },
    momentum: {
      rsiPeriod: 14,
      rsiOverbought: 70,
      rsiOversold: 30,

      macdFastPeriod: 12,
      macdSlowPeriod: 26,
      macdSignalPeriod: 9,

      stochasticKPeriod: 14,
      stochasticDPeriod: 3,
      stochasticOverbought: 80,
      stochasticOversold: 20,

      generalStochasticKPeriod: 14,
      generalStochasticDPeriod: 3,

      cciPeriod: 20,
      cciOverbought: 100,
      cciOversold: -100,

      williamsRPeriod: 14,
      williamsROverbought: -20,
      williamsROversold: -80,

      momentumPeriod: 10,
      rocPeriod: 12,
    },
    trend: {
      adxPeriod: 14,
      adxTrendThreshold: 25,
      ichimokuTenkanPeriod: 9,
      ichimokuKijunPeriod: 26,
      ichimokuSenkouBPeriod: 52,
    },
    volatility: {
      atrPeriod: 14,

      bollingerPeriod: 20,
      bollingerStdDev: 2,

      generalBollingerPeriod: 20,
      generalBollingerStdDev: 2,

      keltnerPeriod: 20,
      keltnerMultiplier: 2,
    },
    supportResistance: {
      donchianPeriod: 20,
      fibonacciLookback: 50,
    },
  }

const useIndicatorSettingsStore = create<IndicatorSettingsStoreState>()(
  persist(
    (set) => ({
      ...DEFAULT_STATE,
      updateMovingAverage: (patch) =>
        set((state) => ({ movingAverage: { ...state.movingAverage, ...patch } })),
      updateMomentum: (patch) =>
        set((state) => ({ momentum: { ...state.momentum, ...patch } })),
      updateTrend: (patch) => set((state) => ({ trend: { ...state.trend, ...patch } })),
      updateVolatility: (patch) =>
        set((state) => ({ volatility: { ...state.volatility, ...patch } })),
      updateSupportResistance: (patch) =>
        set((state) => ({ supportResistance: { ...state.supportResistance, ...patch } })),
      resetAll: () => set(DEFAULT_STATE),
    }),
    {
      name: 'isx-indicator-settings',
      version: SCHEMA_VERSION,
      storage: createJSONStorage(() => localStorage),
      migrate: (persisted: any) => {
        if (!persisted || typeof persisted !== 'object') return DEFAULT_STATE as any
        if ((persisted as any).schemaVersion !== SCHEMA_VERSION) return DEFAULT_STATE as any
        return {
          ...DEFAULT_STATE,
          ...persisted,
          movingAverage: { ...DEFAULT_STATE.movingAverage, ...(persisted as any).movingAverage },
          momentum: { ...DEFAULT_STATE.momentum, ...(persisted as any).momentum },
          trend: { ...DEFAULT_STATE.trend, ...(persisted as any).trend },
          volatility: { ...DEFAULT_STATE.volatility, ...(persisted as any).volatility },
          supportResistance: {
            ...DEFAULT_STATE.supportResistance,
            ...(persisted as any).supportResistance,
          },
        } as any
      },
    }
  )
)

export function useMovingAverageSettings() {
  const settings = useIndicatorSettingsStore((s) => s.movingAverage)
  const updateMovingAverage = useIndicatorSettingsStore((s) => s.updateMovingAverage)

  return {
    settings,
    updateSettings: updateMovingAverage,
    updateGeneralSmaPeriod: (period: number) => updateMovingAverage({ generalSmaPeriod: period }),
    updateGeneralEmaPeriod: (period: number) => updateMovingAverage({ generalEmaPeriod: period }),
  }
}

export function useMomentumSettings() {
  const settings = useIndicatorSettingsStore((s) => s.momentum)
  const updateMomentum = useIndicatorSettingsStore((s) => s.updateMomentum)
  return { settings, updateSettings: updateMomentum }
}

export function useTrendSettings() {
  const settings = useIndicatorSettingsStore((s) => s.trend)
  const updateTrend = useIndicatorSettingsStore((s) => s.updateTrend)
  return { settings, updateSettings: updateTrend }
}

export function useVolatilitySettings() {
  const settings = useIndicatorSettingsStore((s) => s.volatility)
  const updateVolatility = useIndicatorSettingsStore((s) => s.updateVolatility)
  return { settings, updateSettings: updateVolatility }
}

export function useSupportResistanceSettings() {
  const settings = useIndicatorSettingsStore((s) => s.supportResistance)
  const updateSupportResistance = useIndicatorSettingsStore((s) => s.updateSupportResistance)
  return { settings, updateSettings: updateSupportResistance }
}
