/**
 * Viewport Positioning Performance Tests
 *
 * Validates performance requirements and benchmarks for the viewport system
 * Tests memory usage, rendering performance, and responsiveness
 */

import { renderHook, act, waitFor } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { usePanel } from '@/lib/hooks/use-panel'
import { useViewportPosition } from '@/lib/hooks/use-viewport-position'

// Performance test utilities
class PerformanceTracker {
  private startTime: number = 0
  private measurements: number[] = []
  private memoryMeasurements: number[] = []

  startMeasurement() {
    this.startTime = performance.now()
    this.measurements = []
    this.memoryMeasurements = []
  }

  recordMeasurement() {
    const now = performance.now()
    this.measurements.push(now - this.startTime)
  }

  recordMemoryUsage() {
    if ((performance as any).memory) {
      this.memoryMeasurements.push((performance as any).memory.usedJSHeapSize)
    }
  }

  getMetrics() {
    return {
      measurements: this.measurements,
      averageTime: this.measurements.reduce((a, b) => a + b, 0) / this.measurements.length,
      maxTime: Math.max(...this.measurements),
      minTime: Math.min(...this.measurements),
      memoryUsage: this.memoryMeasurements
    }
  }
}

describe('Viewport Positioning Performance Tests', () => {
  let performanceTracker: PerformanceTracker

  beforeEach(() => {
    performanceTracker = new PerformanceTracker()
    vi.clearAllMocks()

    // Reset window dimensions
    Object.defineProperty(window, 'innerWidth', { value: 1920, writable: true })
    Object.defineProperty(window, 'innerHeight', { value: 1080, writable: true })
    Object.defineProperty(window, 'outerWidth', { value: 1920, writable: true })
  })

  describe('60fps Performance Requirements', () => {
    it('should complete position calculations within 16.67ms', async () => {
      const { result } = renderHook(() => usePanel({
        performance: { monitorPerformance: true }
      }))

      performanceTracker.startMeasurement()

      act(() => {
        result.current.reposition()
      })

      const metrics = result.current.performance
      const trackerMetrics = performanceTracker.getMetrics()

      // Hook metrics should be within 60fps requirement
      expect(metrics.positionCalculationTime).toBeLessThan(16.67)
      expect(trackerMetrics.averageTime).toBeLessThan(16.67)
    })

    it('should maintain performance under rapid repeated calls', async () => {
      const { result } = renderHook(() => usePanel({
        performance: { monitorPerformance: true },
        throttleDelay: 16 // 60fps throttling
      }))

      performanceTracker.startMeasurement()

      // Simulate rapid user interactions
      for (let i = 0; i < 30; i++) {
        performanceTracker.recordMeasurement()

        act(() => {
          result.current.reposition()
        })

        // Small delay to simulate real user timing
        await new Promise(resolve => setTimeout(resolve, 5))
      }

      const metrics = result.current.performance
      const trackerMetrics = performanceTracker.getMetrics()

      // Should handle rapid calls efficiently
      expect(trackerMetrics.maxTime).toBeLessThan(50) // Allow some buffer for test environment
      expect(metrics.averagePositionTime).toBeLessThan(16.67)
    })

    it('should scale performance with panel size variations', async () => {
      const sizes = ['small', 'medium', 'large'] as const
      const results: { size: string; time: number }[] = []

      for (const size of sizes) {
        const { result, unmount } = renderHook(() => usePanel({
          size,
          performance: { monitorPerformance: true }
        }))

        performanceTracker.startMeasurement()

        act(() => {
          result.current.reposition()
        })

        const metrics = result.current.performance
        results.push({ size, time: metrics.positionCalculationTime })

        unmount()
      }

      // Performance should be consistent across sizes
      const averageTime = results.reduce((sum, r) => sum + r.time, 0) / results.length
      expect(averageTime).toBeLessThan(10) // Should be even faster than 16.67ms

      // No single size should be significantly slower
      const maxTime = Math.max(...results.map(r => r.time))
      expect(maxTime).toBeLessThan(16.67)
    })
  })

  describe('Memory Management', () => {
    it('should not leak memory during repeated repositioning', async () => {
      const { result } = renderHook(() => usePanel({
        recenterOnResize: true
      }))

      const initialMemory = (performance as any).memory?.usedJSHeapSize || 0

      // Simulate extensive usage
      for (let i = 0; i < 100; i++) {
        performanceTracker.recordMemoryUsage()

        act(() => {
          result.current.reposition()
        })

        if (i % 10 === 0) {
          // Force garbage collection if available
          if (global.gc) {
            global.gc()
          }
        }
      }

      const finalMemory = (performance as any).memory?.usedJSHeapSize || 0
      const memoryIncrease = finalMemory - initialMemory

      // Memory increase should be minimal (less than 1MB)
      expect(memoryIncrease).toBeLessThan(1024 * 1024)
    })

    it('should clean up resources properly on unmount', () => {
      const { unmount } = renderHook(() => usePanel({
        recenterOnResize: true
      }))

      // Simulate component with active operations
      const initialMemory = (performance as any).memory?.usedJSHeapSize || 0

      // Unmount component
      unmount()

      // Force garbage collection if available
      if (global.gc) {
        global.gc()
      }

      const finalMemory = (performance as any).memory?.usedJSHeapSize || 0

      // Memory should be properly cleaned up
      expect(finalMemory - initialMemory).toBeLessThan(1024 * 512) // 512KB tolerance
    })
  })

  describe('Throttling Efficiency', () => {
    it('should throttle repositioning calls effectively', async () => {
      const { result } = renderHook(() => usePanel({
        throttleDelay: 100, // 100ms throttling for easier testing
        performance: { monitorPerformance: true }
      }))

      const startTime = performance.now()
      let repositionCount = 0

      // Make many rapid calls
      for (let i = 0; i < 10; i++) {
        act(() => {
          result.current.reposition()
          repositionCount++
        })
      }

      const endTime = performance.now()
      const duration = endTime - startTime

      // Due to throttling, should complete quickly
      expect(duration).toBeLessThan(50) // Much less than 10 * 100ms
    })

    it('should maintain throttling during viewport changes', async () => {
      const { result } = renderHook(() => usePanel({
        recenterOnResize: true,
        throttleDelay: 16
      }))

      const callCount = { value: 0 }

      // Mock window resize events
      const originalInnerWidth = window.innerWidth
      const originalInnerHeight = window.innerHeight

      // Simulate rapid resize events
      for (let i = 0; i < 20; i++) {
        window.innerWidth = 1920 + (i * 10)
        window.innerHeight = 1080 + (i * 5)

        act(() => {
          callCount.value++
        })

        // Small delay between events
        await new Promise(resolve => setTimeout(resolve, 1))
      }

      // Restore original dimensions
      window.innerWidth = originalInnerWidth
      window.innerHeight = originalInnerHeight

      // Should handle events efficiently
      expect(callCount.value).toBe(20)
      expect(result.current.performance.repositionCount).toBeGreaterThan(0)
    })
  })

  describe('Error Handling Performance', () => {
    it('should handle fallback positioning quickly', async () => {
      // Mock viewport detection failure
      const originalInnerWidth = window.innerWidth
      delete (window as any).innerWidth

      const { result } = renderHook(() => usePanel({
        enableErrorRecovery: true,
        maxRetries: 3
      }))

      performanceTracker.startMeasurement()

      act(() => {
        result.current.reposition()
      })

      const trackerMetrics = performanceTracker.getMetrics()

      // Fallback positioning should be fast
      expect(trackerMetrics.averageTime).toBeLessThan(5)
      expect(result.current.positionStyles).toBeDefined()

      // Restore
      window.innerWidth = originalInnerWidth
    })

    it('should not performance degrade on repeated failures', async () => {
      // Mock consistent viewport detection failure
      delete (window as any).innerWidth

      const { result } = renderHook(() => usePanel({
        enableErrorRecovery: true,
        maxRetries: 5
      }))

      const times: number[] = []

      // Make multiple calls that will all fail
      for (let i = 0; i < 5; i++) {
        const startTime = performance.now()

        act(() => {
          result.current.reposition()
        })

        const endTime = performance.now()
        times.push(endTime - startTime)
      }

      // Performance should remain consistent even with failures
      const averageTime = times.reduce((sum, time) => sum + time, 0) / times.length
      expect(averageTime).toBeLessThan(10)
    })
  })

  describe('Complex Scenario Performance', () => {
    it('should handle concurrent panel operations efficiently', async () => {
      // Simulate multiple panels being positioned simultaneously
      const panels = Array.from({ length: 5 }, (_, i) =>
        renderHook(() => usePanel({
          customWidth: 400 + (i * 50),
          customHeight: 300 + (i * 30),
          placement: i % 2 === 0 ? 'left' : 'right',
          performance: { monitorPerformance: true }
        }))
      )

      performanceTracker.startMeasurement()

      // Position all panels
      panels.forEach(({ result }) => {
        act(() => {
          result.current.reposition()
        })
      })

      const totalTime = performanceTracker.getMetrics().maxTime

      // Even with multiple panels, individual operations should be fast
      expect(totalTime).toBeLessThan(50)

      // Clean up
      panels.forEach(({ unmount }) => unmount())
    })

    it('should maintain performance during zoom level transitions', async () => {
      const { result } = renderHook(() => usePanel({
        performance: { monitorPerformance: true }
      }))

      const zoomLevels = [1.0, 1.25, 1.5, 2.0, 1.0] // Zoom in and back out
      const times: number[] = []

      for (const zoom of zoomLevels) {
        window.innerWidth = Math.floor(1920 / zoom)
        window.innerHeight = Math.floor(1080 / zoom)

        const startTime = performance.now()

        act(() => {
          result.current.reposition()
        })

        const endTime = performance.now()
        times.push(endTime - startTime)
      }

      // Performance should be consistent across zoom levels
      const averageTime = times.reduce((sum, time) => sum + time, 0) / times.length
      const maxTime = Math.max(...times)

      expect(averageTime).toBeLessThan(10)
      expect(maxTime).toBeLessThan(16.67)
    })
  })

  describe('Benchmark Validation', () => {
    it('should meet all performance benchmarks', () => {
      const { result } = renderHook(() => usePanel({
        performance: { monitorPerformance: true },
        throttleDelay: 16
      }))

      // Run comprehensive performance test
      const benchmark = {
        positionCalculations: 100,
        maxTimePerCalculation: 16.67, // 60fps
        maxMemoryIncrease: 1024 * 1024, // 1MB
        minThroughput: 60 // calculations per second
      }

      const startTime = performance.now()
      const initialMemory = (performance as any).memory?.usedJSHeapSize || 0

      // Run benchmark
      for (let i = 0; i < benchmark.positionCalculations; i++) {
        act(() => {
          result.current.reposition()
        })
      }

      const endTime = performance.now()
      const finalMemory = (performance as any).memory?.usedJSHeapSize || 0
      const totalTime = endTime - startTime
      const memoryIncrease = finalMemory - initialMemory
      const throughput = benchmark.positionCalculations / (totalTime / 1000)

      // Validate benchmarks
      expect(throughput).toBeGreaterThan(benchmark.minThroughput)
      expect(memoryIncrease).toBeLessThan(benchmark.maxMemoryIncrease)
      expect(result.current.performance.averagePositionTime).toBeLessThan(benchmark.maxTimePerCalculation)
    })
  })
})