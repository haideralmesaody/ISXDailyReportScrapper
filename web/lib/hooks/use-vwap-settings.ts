/**
 * VWAP Settings Hook
 * Manages VWAP indicator settings with localStorage persistence
 * Follows the same pattern as other indicator settings hooks
 */

'use client'

import { useState, useEffect, useCallback } from 'react'

export interface VWAPSettings {
  anchorDays: number | null  // null = chart start, number = days back
  version?: number  // Increments on every update to force React remounts
  schemaVersion?: number  // Schema version for migration detection
}

const SETTINGS_VERSION = 1  // Increment when schema changes or defaults need reset

const DEFAULT_SETTINGS: VWAPSettings = {
  anchorDays: 30,  // Default to last 30 days (1 month - ideal for recent VWAP S/R analysis)
  version: 0,
  schemaVersion: SETTINGS_VERSION
}

const STORAGE_KEY = 'isx-vwap-settings'

export function useVWAPSettings() {
  const [settings, setSettings] = useState<VWAPSettings>(DEFAULT_SETTINGS)

  // Load settings from localStorage on mount
  useEffect(() => {
    if (typeof window === 'undefined') return

    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      try {
        const parsed = JSON.parse(stored)

        // Migration: Reset to defaults if schema version doesn't match
        if (!parsed.schemaVersion || parsed.schemaVersion !== SETTINGS_VERSION) {
          console.log('[useVWAPSettings] Old schema detected (v' + (parsed.schemaVersion || 0) + ' → v' + SETTINGS_VERSION + '), resetting to defaults')
          localStorage.setItem(STORAGE_KEY, JSON.stringify(DEFAULT_SETTINGS))
          setSettings(DEFAULT_SETTINGS)
          return
        }

        setSettings(parsed)
      } catch (e) {
        console.error('[useVWAPSettings] Failed to parse settings:', e)
        // Reset to defaults on parse error
        localStorage.removeItem(STORAGE_KEY)
        setSettings(DEFAULT_SETTINGS)
      }
    }
  }, [])

  // Listen for localStorage changes from other components (cross-component reactivity)
  // This ensures ChartContainer's hook detects when IndicatorSettingsPopover updates settings
  useEffect(() => {
    if (typeof window === 'undefined') return

    const handleStorageChange = (e: StorageEvent) => {
      // Only respond to changes for our specific key
      if (e.key !== STORAGE_KEY) return

      // Storage event fires when OTHER tabs/windows change localStorage
      // For same-window updates, we need a custom event (see updateSettings)
      if (e.newValue) {
        try {
          const parsed = JSON.parse(e.newValue)
          console.log('[useVWAPSettings] Storage event detected, updating settings:', parsed)
          setSettings(parsed)
        } catch (err) {
          console.error('[useVWAPSettings] Failed to parse storage event value:', err)
        }
      }
    }

    // Listen for custom event dispatched by updateSettings (same-window updates)
    const handleCustomStorageChange = ((e: CustomEvent) => {
      console.log('[useVWAPSettings] Custom storage event detected, updating settings:', e.detail)
      setSettings(e.detail)
    }) as EventListener

    window.addEventListener('storage', handleStorageChange)
    window.addEventListener('vwap-settings-changed', handleCustomStorageChange)

    return () => {
      window.removeEventListener('storage', handleStorageChange)
      window.removeEventListener('vwap-settings-changed', handleCustomStorageChange)
    }
  }, [])

  // Update settings and persist to localStorage
  const updateSettings = useCallback((updates: Partial<VWAPSettings>) => {
    setSettings(prev => {
      // Increment version to force React component remounts
      const next = {
        ...prev,
        ...updates,
        version: (prev.version || 0) + 1,
        schemaVersion: SETTINGS_VERSION  // Preserve schema version
      }

      // Persist to localStorage
      if (typeof window !== 'undefined') {
        try {
          localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
          console.log('[useVWAPSettings] Settings updated (v' + next.version + '):', next)

          // Dispatch custom event for same-window cross-component updates
          // (Native 'storage' event only fires for OTHER tabs/windows)
          window.dispatchEvent(new CustomEvent('vwap-settings-changed', { detail: next }))
        } catch (e) {
          console.error('[useVWAPSettings] Failed to save settings:', e)
        }
      }

      return next
    })
  }, [])

  return { settings, updateSettings }
}
