/**
 * Chart Performance Monitor
 * 
 * Utility for monitoring and measuring chart performance metrics.
 * Helps track the effectiveness of optimizations and identify bottlenecks.
 */

interface PerformanceMetrics {
  renderTime: number
  dataProcessingTime: number
  memoryUsage: number
  chartInitTime: number
  resizeHandling: number
  reRenderCount: number
}

interface ChartPerformanceConfig {
  enableLogging: boolean
  sampleRate: number
  maxMemoryThreshold: number
  slowRenderThreshold: number
}

class ChartPerformanceMonitor {
  private metrics: Map<string, PerformanceMetrics> = new Map()
  private config: ChartPerformanceConfig
  private observers: PerformanceObserver[] = []

  constructor(config: Partial<ChartPerformanceConfig> = {}) {
    this.config = {
      enableLogging: process.env.NODE_ENV === 'development',
      sampleRate: 1.0,
      maxMemoryThreshold: 50 * 1024 * 1024, // 50MB
      slowRenderThreshold: 1000, // 1 second
      ...config
    }

    this.setupPerformanceObservers()
  }

  /**
   * Start monitoring a chart instance
   */
  startMonitoring(chartId: string): void {
    const startTime = performance.now()
    
    this.metrics.set(chartId, {
      renderTime: 0,
      dataProcessingTime: 0,
      memoryUsage: this.getMemoryUsage(),
      chartInitTime: startTime,
      resizeHandling: 0,
      reRenderCount: 0
    })

    if (this.config.enableLogging) {
      console.log(`📊 Started monitoring chart: ${chartId}`)
    }
  }

  /**
   * Record data processing time
   */
  recordDataProcessing(chartId: string, startTime: number): void {
    const metrics = this.metrics.get(chartId)
    if (!metrics) return

    const processingTime = performance.now() - startTime
    metrics.dataProcessingTime = processingTime

    if (processingTime > 500 && this.config.enableLogging) {
      console.warn(`⚠️ Slow data processing for ${chartId}: ${processingTime.toFixed(2)}ms`)
    }
  }

  /**
   * Record chart render time
   */
  recordRenderTime(chartId: string, startTime: number): void {
    const metrics = this.metrics.get(chartId)
    if (!metrics) return

    const renderTime = performance.now() - startTime
    metrics.renderTime = renderTime
    metrics.reRenderCount++

    if (renderTime > this.config.slowRenderThreshold && this.config.enableLogging) {
      console.warn(`⚠️ Slow render for ${chartId}: ${renderTime.toFixed(2)}ms`)
    }
  }

  /**
   * Record chart initialization completion
   */
  recordChartInit(chartId: string): void {
    const metrics = this.metrics.get(chartId)
    if (!metrics) return

    const initTime = performance.now() - metrics.chartInitTime
    metrics.chartInitTime = initTime

    if (this.config.enableLogging) {
      console.log(`🚀 Chart ${chartId} initialized in ${initTime.toFixed(2)}ms`)
    }
  }

  /**
   * Record resize handling performance
   */
  recordResizeHandling(chartId: string, startTime: number): void {
    const metrics = this.metrics.get(chartId)
    if (!metrics) return

    const resizeTime = performance.now() - startTime
    metrics.resizeHandling = resizeTime

    if (resizeTime > 100 && this.config.enableLogging) {
      console.warn(`⚠️ Slow resize handling for ${chartId}: ${resizeTime.toFixed(2)}ms`)
    }
  }

  /**
   * Update memory usage
   */
  updateMemoryUsage(chartId: string): void {
    const metrics = this.metrics.get(chartId)
    if (!metrics) return

    const memoryUsage = this.getMemoryUsage()
    metrics.memoryUsage = memoryUsage

    if (memoryUsage > this.config.maxMemoryThreshold && this.config.enableLogging) {
      console.warn(`⚠️ High memory usage for ${chartId}: ${(memoryUsage / 1024 / 1024).toFixed(2)}MB`)
    }
  }

  /**
   * Get performance summary for a chart
   */
  getSummary(chartId: string): PerformanceMetrics | null {
    return this.metrics.get(chartId) || null
  }

  /**
   * Get all performance metrics
   */
  getAllMetrics(): Map<string, PerformanceMetrics> {
    return new Map(this.metrics)
  }

  /**
   * Log performance summary
   */
  logSummary(chartId: string): void {
    const metrics = this.metrics.get(chartId)
    if (!metrics || !this.config.enableLogging) return

    console.group(`📈 Performance Summary: ${chartId}`)
    console.log(`📊 Chart Init: ${metrics.chartInitTime.toFixed(2)}ms`)
    console.log(`🔄 Data Processing: ${metrics.dataProcessingTime.toFixed(2)}ms`)
    console.log(`🎨 Render Time: ${metrics.renderTime.toFixed(2)}ms`)
    console.log(`📐 Resize Handling: ${metrics.resizeHandling.toFixed(2)}ms`)
    console.log(`🔄 Re-render Count: ${metrics.reRenderCount}`)
    console.log(`💾 Memory Usage: ${(metrics.memoryUsage / 1024 / 1024).toFixed(2)}MB`)
    console.groupEnd()
  }

  /**
   * Clean up monitoring data
   */
  stopMonitoring(chartId: string): void {
    if (this.config.enableLogging) {
      this.logSummary(chartId)
    }
    
    this.metrics.delete(chartId)
  }

  /**
   * Get current memory usage
   */
  private getMemoryUsage(): number {
    if ('memory' in performance) {
      return (performance as any).memory?.usedJSHeapSize || 0
    }
    return 0
  }

  /**
   * Setup performance observers
   */
  private setupPerformanceObservers(): void {
    if (typeof PerformanceObserver === 'undefined') return

    // Monitor long tasks
    try {
      const longTaskObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.duration > 50 && this.config.enableLogging) {
            console.warn(`⚠️ Long task detected: ${entry.duration.toFixed(2)}ms`)
          }
        }
      })
      longTaskObserver.observe({ entryTypes: ['longtask'] })
      this.observers.push(longTaskObserver)
    } catch (e) {
      // Long task API not supported
    }

    // Monitor layout shifts
    try {
      const layoutShiftObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if ((entry as any).value > 0.1 && this.config.enableLogging) {
            console.warn(`⚠️ Layout shift detected: ${(entry as any).value.toFixed(4)}`)
          }
        }
      })
      layoutShiftObserver.observe({ entryTypes: ['layout-shift'] })
      this.observers.push(layoutShiftObserver)
    } catch (e) {
      // Layout shift API not supported
    }
  }

  /**
   * Cleanup observers
   */
  destroy(): void {
    this.observers.forEach(observer => observer.disconnect())
    this.observers = []
    this.metrics.clear()
  }
}

// Create singleton instance
const chartPerformanceMonitor = new ChartPerformanceMonitor()

export { chartPerformanceMonitor, ChartPerformanceMonitor }
export type { PerformanceMetrics, ChartPerformanceConfig }

// React hook for easy integration
export function useChartPerformanceMonitor(chartId: string) {
  const startMonitoring = () => chartPerformanceMonitor.startMonitoring(chartId)
  const stopMonitoring = () => chartPerformanceMonitor.stopMonitoring(chartId)
  
  const recordDataProcessing = (startTime: number) => 
    chartPerformanceMonitor.recordDataProcessing(chartId, startTime)
  
  const recordRenderTime = (startTime: number) => 
    chartPerformanceMonitor.recordRenderTime(chartId, startTime)
  
  const recordChartInit = () => 
    chartPerformanceMonitor.recordChartInit(chartId)
  
  const recordResizeHandling = (startTime: number) => 
    chartPerformanceMonitor.recordResizeHandling(chartId, startTime)
  
  const updateMemoryUsage = () => 
    chartPerformanceMonitor.updateMemoryUsage(chartId)
  
  const getSummary = () => 
    chartPerformanceMonitor.getSummary(chartId)

  return {
    startMonitoring,
    stopMonitoring,
    recordDataProcessing,
    recordRenderTime,
    recordChartInit,
    recordResizeHandling,
    updateMemoryUsage,
    getSummary
  }
}

// Utility functions for performance testing
export const performanceUtils = {
  /**
   * Measure function execution time
   */
  measure: async <T>(name: string, fn: () => Promise<T> | T): Promise<T> => {
    const startTime = performance.now()
    const result = await fn()
    const duration = performance.now() - startTime
    
    console.log(`⏱️ ${name}: ${duration.toFixed(2)}ms`)
    return result
  },

  /**
   * Create a performance mark
   */
  mark: (name: string): void => {
    performance.mark(name)
  },

  /**
   * Measure between two marks
   */
  measureBetween: (name: string, startMark: string, endMark: string): number => {
    performance.mark(endMark)
    performance.measure(name, startMark, endMark)
    
    const entries = performance.getEntriesByName(name, 'measure')
    return entries.length > 0 ? entries[entries.length - 1].duration : 0
  },

  /**
   * Get bundle size estimate
   */
  estimateBundleSize: (): string => {
    // Rough estimate based on script tags
    const scripts = Array.from(document.scripts)
    let totalSize = 0
    
    scripts.forEach(script => {
      if (script.src && !script.src.includes('node_modules')) {
        // This is a rough estimate - actual size would need server-side measurement
        totalSize += script.textContent?.length || 0
      }
    })
    
    return `${(totalSize / 1024).toFixed(2)}KB (estimated)`
  }
}