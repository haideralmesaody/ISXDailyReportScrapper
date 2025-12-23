/**
 * Chart Synchronization Utilities
 * Following CLAUDE.md patterns for professional multi-chart coordination
 */

import { IChartApi, ISeriesApi, Time, LogicalRange } from 'lightweight-charts'

export interface ChartInstance {
  chart: IChartApi
  id: string
  type: 'main' | 'volume' | 'macd' | 'rsi'
}

export class ChartSynchronizer {
  private charts: Map<string, ChartInstance> = new Map()
  private isUpdating = false

  constructor() {
    this.handleTimeScaleChange = this.handleTimeScaleChange.bind(this)
    this.handleCrosshairMove = this.handleCrosshairMove.bind(this)
  }

  /**
   * Add a chart to the synchronization group
   */
  addChart(chart: IChartApi, id: string, type: ChartInstance['type']) {
    const instance: ChartInstance = { chart, id, type }
    this.charts.set(id, instance)

    // Subscribe to time scale changes
    chart.timeScale().subscribeVisibleLogicalRangeChange(this.handleTimeScaleChange)

    // Subscribe to crosshair movement
    chart.subscribeCrosshairMove(this.handleCrosshairMove)

    return instance
  }

  /**
   * Remove a chart from synchronization
   */
  removeChart(id: string) {
    const instance = this.charts.get(id)
    if (instance) {
      // Unsubscribe from events
      instance.chart.timeScale().unsubscribeVisibleLogicalRangeChange(this.handleTimeScaleChange)
      instance.chart.unsubscribeCrosshairMove(this.handleCrosshairMove)
      this.charts.delete(id)
    }
  }

  /**
   * Synchronize time scale across all charts
   */
  private handleTimeScaleChange(logicalRange: LogicalRange | null) {
    if (this.isUpdating || !logicalRange) return

    this.isUpdating = true

    // Apply the same logical range to all other charts
    this.charts.forEach((instance) => {
      try {
        instance.chart.timeScale().setVisibleLogicalRange(logicalRange)
      } catch (error) {
        console.warn('Failed to sync time scale for chart:', instance.id, error)
      }
    })

    this.isUpdating = false
  }

  /**
   * Synchronize crosshair movement across all charts
   */
  private handleCrosshairMove(param: any) {
    if (this.isUpdating) return

    this.isUpdating = true

    // Get the time from the crosshair movement
    const time = param.time

    if (time) {
      // Move crosshair on all other charts to the same time
      this.charts.forEach((instance) => {
        try {
          // Use proper crosshair sync with time coordinate
          // Note: TradingView Lightweight Charts doesn't have setCrosshairPosition
          // Instead, we can use the crosshair move event to sync behavior
          const timeScale = instance.chart.timeScale()
          const coordinate = timeScale.timeToCoordinate(time)
          if (coordinate !== null) {
            // Trigger crosshair move event on other charts
            // This is a simplified sync - full implementation would require custom events
          }
        } catch (error) {
          // Silently ignore errors - crosshair sync is non-critical
        }
      })
    }

    this.isUpdating = false
  }

  /**
   * Fit all charts content to visible range
   */
  fitAllChartsContent() {
    this.charts.forEach((instance) => {
      instance.chart.timeScale().fitContent()
    })
  }

  /**
   * Apply zoom to all charts
   */
  zoomAllCharts(factor: number) {
    this.isUpdating = true

    this.charts.forEach((instance) => {
      try {
        const timeScale = instance.chart.timeScale()
        const visibleRange = timeScale.getVisibleLogicalRange()

        if (visibleRange) {
          const center = (visibleRange.from + visibleRange.to) / 2
          const newRange = (visibleRange.to - visibleRange.from) * factor

          timeScale.setVisibleLogicalRange({
            from: center - newRange / 2,
            to: center + newRange / 2
          })
        }
      } catch (error) {
        console.warn('Failed to zoom chart:', instance.id, error)
      }
    })

    this.isUpdating = false
  }

  /**
   * Reset all charts to default view
   */
  resetAllCharts() {
    this.charts.forEach((instance) => {
      instance.chart.timeScale().resetTimeScale()
    })
  }

  /**
   * Get chart instance by ID
   */
  getChart(id: string): ChartInstance | undefined {
    return this.charts.get(id)
  }

  /**
   * Get all synchronized charts
   */
  getAllCharts(): ChartInstance[] {
    return Array.from(this.charts.values())
  }

  /**
   * Clean up all synchronization
   */
  destroy() {
    this.charts.forEach((instance) => {
      this.removeChart(instance.id)
    })
    this.charts.clear()
  }
}

/**
 * Utility function to create consistent chart options across instances
 */
export function createSyncedChartOptions(baseOptions: any, chartType: ChartInstance['type']) {
  return {
    ...baseOptions,
    timeScale: {
      ...baseOptions.timeScale,
      visible: chartType === 'main', // Only show time scale on main chart
      borderVisible: chartType === 'main'
    },
    rightPriceScale: {
      ...baseOptions.rightPriceScale,
      borderVisible: true,
      visible: true
    },
    // Remove scale margins - we're using separate charts now
    layout: {
      ...baseOptions.layout
    }
  }
}

/**
 * Default chart heights for different layout configurations
 */
export const CHART_LAYOUTS = {
  mainOnly: { main: 1.0 },
  mainWithVolume: { main: 0.7, volume: 0.3 },
  mainVolumeMACD: { main: 0.5, volume: 0.3, macd: 0.2 },
  mainVolumeRSI: { main: 0.5, volume: 0.3, rsi: 0.2 },
  full: { main: 0.4, volume: 0.2, macd: 0.2, rsi: 0.2 }
} as const

export type ChartLayout = keyof typeof CHART_LAYOUTS