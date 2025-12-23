/**
 * Professional Indicator Registry System
 * Manages all chart indicators dynamically for scalability
 */

export type IndicatorType = 'overlay' | 'oscillator' | 'volume' | 'support-resistance'
export type IndicatorCategory = 'moving-average' | 'momentum' | 'volatility' | 'volume' | 'support-resistance' | 'trend'
export type LineStyle = 'solid' | 'dashed' | 'dotted'

export interface IndicatorConfig {
  id: string
  name: string
  displayName: string
  type: IndicatorType
  color: string
  lineStyle?: LineStyle
  lineWidth?: number
  category: IndicatorCategory
  defaultPeriod?: number
  periods?: number[] // For indicators with multiple periods like Bollinger Bands
  description?: string
  seriesRef?: any // Reference to the lightweight-charts series
  visible?: boolean
  formatValue?: (value: number) => string // Custom formatter
}

export interface IndicatorValue {
  value: number | null
  color: string
  displayName: string
  formatted: string
}

/**
 * Singleton registry for managing all indicators
 */
export class IndicatorRegistry {
  private static instance: IndicatorRegistry
  private indicators: Map<string, IndicatorConfig> = new Map()
  private activeIndicators: Set<string> = new Set()
  private indicatorValues: Map<string, IndicatorValue> = new Map()

  private constructor() {
    this.initializeDefaultIndicators()
  }

  static getInstance(): IndicatorRegistry {
    if (!IndicatorRegistry.instance) {
      IndicatorRegistry.instance = new IndicatorRegistry()
    }
    return IndicatorRegistry.instance
  }

  /**
   * Register a new indicator
   */
  register(indicator: IndicatorConfig): void {
    this.indicators.set(indicator.id, indicator)
    if (indicator.visible !== false) {
      this.activeIndicators.add(indicator.id)
    }
  }

  /**
   * Unregister an indicator
   */
  unregister(id: string): void {
    this.indicators.delete(id)
    this.activeIndicators.delete(id)
    this.indicatorValues.delete(id)
  }

  /**
   * Get all registered indicators
   */
  getAll(): IndicatorConfig[] {
    return Array.from(this.indicators.values())
  }

  /**
   * Get active indicators
   */
  getActive(): IndicatorConfig[] {
    return Array.from(this.indicators.values()).filter(
      indicator => this.activeIndicators.has(indicator.id)
    )
  }

  /**
   * Get indicators by category
   */
  getByCategory(category: IndicatorCategory): IndicatorConfig[] {
    return Array.from(this.indicators.values()).filter(
      indicator => indicator.category === category
    )
  }

  /**
   * Get indicator by ID
   */
  getById(id: string): IndicatorConfig | undefined {
    return this.indicators.get(id)
  }

  /**
   * Set indicator visibility
   */
  setVisible(id: string, visible: boolean): void {
    const indicator = this.indicators.get(id)
    if (indicator) {
      indicator.visible = visible
      if (visible) {
        this.activeIndicators.add(id)
      } else {
        this.activeIndicators.delete(id)
        this.indicatorValues.delete(id)
      }
    }
  }

  /**
   * Check if indicator is active
   */
  isActive(id: string): boolean {
    return this.activeIndicators.has(id)
  }

  /**
   * Set series reference for an indicator
   */
  setSeriesRef(id: string, ref: any): void {
    const indicator = this.indicators.get(id)
    if (indicator) {
      indicator.seriesRef = ref
    }
  }

  /**
   * Update indicator value
   */
  updateValue(id: string, value: number | null): void {
    const indicator = this.indicators.get(id)
    if (indicator && value !== null) {
      const formatted = indicator.formatValue 
        ? indicator.formatValue(value)
        : this.defaultFormatValue(value, indicator.type)
      
      this.indicatorValues.set(id, {
        value,
        color: indicator.color,
        displayName: indicator.displayName,
        formatted
      })
    } else {
      this.indicatorValues.delete(id)
    }
  }

  /**
   * Get all current indicator values
   */
  getValues(): Map<string, IndicatorValue> {
    return new Map(this.indicatorValues)
  }

  /**
   * Clear all indicator values
   */
  clearValues(): void {
    this.indicatorValues.clear()
  }

  /**
   * Get active indicator IDs in display order
   */
  getActiveIds(): string[] {
    // Order by category then by name
    const categoryOrder: IndicatorCategory[] = [
      'moving-average',
      'trend',
      'momentum',
      'volatility',
      'volume',
      'support-resistance'
    ]
    
    return this.getActive()
      .sort((a, b) => {
        const categoryDiff = categoryOrder.indexOf(a.category) - categoryOrder.indexOf(b.category)
        if (categoryDiff !== 0) return categoryDiff
        return a.displayName.localeCompare(b.displayName)
      })
      .map(indicator => indicator.id)
  }

  /**
   * Default value formatter
   */
  private defaultFormatValue(value: number, type: IndicatorType): string {
    if (type === 'volume') {
      return new Intl.NumberFormat('en-US', {
        maximumFractionDigits: 0,
        notation: 'compact'
      }).format(value)
    }
    
    if (type === 'oscillator') {
      return value.toFixed(2)
    }
    
    // For overlay and support-resistance
    return new Intl.NumberFormat('en-IQ', {
      style: 'currency',
      currency: 'IQD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value)
  }

  /**
   * Initialize default indicators
   */
  private initializeDefaultIndicators(): void {
    // Moving Averages
    this.register({
      id: 'sma20',
      name: 'SMA 20',
      displayName: 'SMA (20)',
      type: 'overlay',
      color: '#3B82F6', // Blue
      category: 'moving-average',
      defaultPeriod: 20,
      lineWidth: 2,
      description: 'Simple Moving Average - 20 periods'
    })

    this.register({
      id: 'sma50',
      name: 'SMA 50',
      displayName: 'SMA (50)',
      type: 'overlay',
      color: '#F97316', // Orange
      category: 'moving-average',
      defaultPeriod: 50,
      lineWidth: 2,
      description: 'Simple Moving Average - 50 periods'
    })

    this.register({
      id: 'sma200',
      name: 'SMA 200',
      displayName: 'SMA (200)',
      type: 'overlay',
      color: '#8B5CF6', // Purple
      category: 'moving-average',
      defaultPeriod: 200,
      lineWidth: 2,
      description: 'Simple Moving Average - 200 periods'
    })

    this.register({
      id: 'ema20',
      name: 'EMA 20',
      displayName: 'EMA (20)',
      type: 'overlay',
      color: '#06B6D4', // Cyan
      category: 'moving-average',
      defaultPeriod: 20,
      lineStyle: 'dashed',
      lineWidth: 2,
      description: 'Exponential Moving Average - 20 periods',
      visible: false // Default to hidden
    })

    // Support & Resistance
    this.register({
      id: 'resistance',
      name: 'Resistance',
      displayName: 'Resistance Level',
      type: 'support-resistance',
      color: '#EF4444', // Red
      category: 'support-resistance',
      lineStyle: 'dashed',
      lineWidth: 2,
      description: 'Calculated resistance level'
    })

    this.register({
      id: 'support',
      name: 'Support',
      displayName: 'Support Level',
      type: 'support-resistance',
      color: '#10B981', // Green
      category: 'support-resistance',
      lineStyle: 'dashed',
      lineWidth: 2,
      description: 'Calculated support level'
    })

    // MACD Indicators
    this.register({
      id: 'macd',
      name: 'MACD',
      displayName: 'MACD Line',
      type: 'oscillator',
      color: '#2962ff', // Blue
      category: 'momentum',
      lineWidth: 2,
      description: 'Moving Average Convergence Divergence',
      formatValue: (value: number) => value.toFixed(4)
    })

    this.register({
      id: 'macd_signal',
      name: 'MACD Signal',
      displayName: 'MACD Signal',
      type: 'oscillator',
      color: '#ff6838', // Orange
      category: 'momentum',
      lineWidth: 2,
      description: 'MACD Signal Line (9-period EMA)',
      formatValue: (value: number) => value.toFixed(4)
    })

    this.register({
      id: 'macd_histogram',
      name: 'MACD Histogram',
      displayName: 'MACD Histogram',
      type: 'oscillator',
      color: '#26a69a', // Teal
      category: 'momentum',
      description: 'MACD - Signal Line',
      formatValue: (value: number) => value.toFixed(4)
    })

    // RSI Indicator
    this.register({
      id: 'rsi',
      name: 'RSI',
      displayName: 'RSI (14)',
      type: 'oscillator',
      color: '#9c27b0', // Purple
      category: 'momentum',
      lineWidth: 2,
      defaultPeriod: 14,
      description: 'Relative Strength Index',
      formatValue: (value: number) => value.toFixed(2)
    })
  }

  /**
   * Reset to default state
   */
  reset(): void {
    this.indicators.clear()
    this.activeIndicators.clear()
    this.indicatorValues.clear()
    this.initializeDefaultIndicators()
  }
}

/**
 * Indicator Factory for creating new indicator configurations
 */
export class IndicatorFactory {
  static createSMA(period: number, color: string = '#3B82F6'): IndicatorConfig {
    return {
      id: `sma${period}`,
      name: `SMA ${period}`,
      displayName: `SMA (${period})`,
      type: 'overlay',
      color,
      category: 'moving-average',
      defaultPeriod: period,
      lineWidth: 2,
      description: `Simple Moving Average - ${period} periods`
    }
  }

  static createEMA(period: number, color: string = '#06B6D4'): IndicatorConfig {
    return {
      id: `ema${period}`,
      name: `EMA ${period}`,
      displayName: `EMA (${period})`,
      type: 'overlay',
      color,
      category: 'moving-average',
      defaultPeriod: period,
      lineStyle: 'dashed',
      lineWidth: 2,
      description: `Exponential Moving Average - ${period} periods`
    }
  }

  static createBollingerBands(period: number = 20, stdDev: number = 2): IndicatorConfig {
    return {
      id: `bb${period}_${stdDev}`,
      name: `Bollinger Bands (${period}, ${stdDev})`,
      displayName: `BB (${period}, ${stdDev})`,
      type: 'overlay',
      color: '#9333EA', // Purple
      category: 'volatility',
      periods: [period],
      lineWidth: 1,
      description: `Bollinger Bands - ${period} periods, ${stdDev} std dev`
    }
  }

  static createRSI(period: number = 14): IndicatorConfig {
    return {
      id: `rsi${period}`,
      name: `RSI ${period}`,
      displayName: `RSI (${period})`,
      type: 'oscillator',
      color: '#F59E0B', // Amber
      category: 'momentum',
      defaultPeriod: period,
      lineWidth: 2,
      description: `Relative Strength Index - ${period} periods`,
      formatValue: (value: number) => value.toFixed(2)
    }
  }

  static createMACD(fast: number = 12, slow: number = 26, signal: number = 9): IndicatorConfig {
    return {
      id: `macd${fast}_${slow}_${signal}`,
      name: `MACD (${fast}, ${slow}, ${signal})`,
      displayName: `MACD (${fast}, ${slow}, ${signal})`,
      type: 'oscillator',
      color: '#10B981', // Green
      category: 'momentum',
      periods: [fast, slow, signal],
      lineWidth: 2,
      description: `MACD - Fast: ${fast}, Slow: ${slow}, Signal: ${signal}`
    }
  }

  static createVolume(color: string = '#6B7280'): IndicatorConfig {
    return {
      id: 'volume',
      name: 'Volume',
      displayName: 'Volume',
      type: 'volume',
      color,
      category: 'volume',
      description: 'Trading volume'
    }
  }

  static createVWAP(): IndicatorConfig {
    return {
      id: 'vwap',
      name: 'VWAP',
      displayName: 'VWAP',
      type: 'overlay',
      color: '#EC4899', // Pink
      category: 'volume',
      lineWidth: 2,
      description: 'Volume Weighted Average Price'
    }
  }

  static createStochastic(kPeriod: number = 14, dPeriod: number = 3): IndicatorConfig {
    return {
      id: `stoch${kPeriod}_${dPeriod}`,
      name: `Stochastic (${kPeriod}, ${dPeriod})`,
      displayName: `Stoch (${kPeriod}, ${dPeriod})`,
      type: 'oscillator',
      color: '#6366F1', // Indigo
      category: 'momentum',
      periods: [kPeriod, dPeriod],
      lineWidth: 2,
      description: `Stochastic Oscillator - K: ${kPeriod}, D: ${dPeriod}`,
      formatValue: (value: number) => value.toFixed(2)
    }
  }
}

// Export singleton instance for convenience
export const indicatorRegistry = IndicatorRegistry.getInstance()