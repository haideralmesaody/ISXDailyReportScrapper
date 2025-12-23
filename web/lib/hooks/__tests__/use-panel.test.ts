/**
 * usePanel Hook Tests
 */

import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { usePanel } from '../use-panel'

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  disconnect: vi.fn(),
  unobserve: vi.fn(),
}))

// Mock window.visualViewport
Object.defineProperty(window, 'visualViewport', {
  value: {
    width: 1920,
    height: 1080,
    offsetLeft: 0,
    offsetTop: 0,
    scale: 1,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  },
  writable: true,
})

describe('usePanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should provide default panel configuration', () => {
    const { result } = renderHook(() => usePanel())

    expect(result.current.supportLevel).toBeDefined()
    expect(result.current.panelSize.width).toBe(480) // medium preset
    expect(result.current.panelSize.height).toBe(600) // medium preset
    expect(result.current.positionStyles).toBeDefined()
  })

  it('should use custom dimensions when provided', () => {
    const { result } = renderHook(() => usePanel({
      customWidth: 600,
      customHeight: 400
    }))

    expect(result.current.panelSize.width).toBe(600)
    expect(result.current.panelSize.height).toBe(400)
  })

  it('should use size presets correctly', () => {
    const { result: smallResult } = renderHook(() => usePanel({ size: 'small' }))
    const { result: largeResult } = renderHook(() => usePanel({ size: 'large' }))

    expect(smallResult.current.panelSize.width).toBe(320)
    expect(smallResult.current.panelSize.height).toBe(400)
    expect(largeResult.current.panelSize.width).toBe(600)
    expect(largeResult.current.panelSize.height).toBe(700)
  })

  it('should provide focus management methods', () => {
    const { result } = renderHook(() => usePanel({
      focus: { autoFocus: true }
    }))

    expect(result.current.focus.focusFirst).toBeDefined()
    expect(result.current.focus.focusElement).toBeDefined()
    expect(result.current.focus.isElementInPanel).toBeDefined()
  })

  it('should include performance metrics when enabled', () => {
    const { result } = renderHook(() => usePanel({
      performance: { monitorPerformance: true }
    }))

    expect(result.current.performance).toBeDefined()
    expect(result.current.performance.positionCalculationTime).toBeGreaterThanOrEqual(0)
  })

  it('should handle error recovery when enabled', () => {
    const { result } = renderHook(() => usePanel({
      enableErrorRecovery: true,
      maxRetries: 3
    }))

    expect(result.current.lastError).toBeNull()
    expect(result.current.retryCount).toBe(0)
  })

  it('should update element references', () => {
    const { result } = renderHook(() => usePanel())

    const mockElement = document.createElement('div')
    act(() => {
      result.current.setElement(mockElement)
    })

    expect(result.current.positionStyles).toBeDefined()
  })
})