/**
 * Professional Dynamic Layout Management System for Chart Indicators
 * Handles dynamic space allocation, pane resizing, and user preferences
 */

export type PaneType = 'main' | 'volume' | 'oscillator' | 'momentum' | 'volatility' | 'custom'
export type IndicatorType = 'overlay' | 'volume' | 'oscillator' | 'momentum' | 'bands' | 'levels'

export interface ScaleMargins {
  top: number    // 0-1 percentage from top
  bottom: number // 0-1 percentage from bottom
}

export interface PaneConfig {
  id: string
  type: PaneType
  heightWeight: number     // Relative weight for dynamic allocation
  minHeight: number       // Minimum percentage (e.g., 0.05 = 5%)
  maxHeight: number       // Maximum percentage (e.g., 0.85 = 85%)
  priority: number        // Higher priority gets space first
  canCombine: boolean     // Can share space with similar indicators
  indicators: string[]    // Indicator IDs in this pane
  resizable: boolean      // User can resize (default: true)
  currentHeight?: number  // User-adjusted height (0-1)
  visible: boolean        // Is pane visible
}

export interface IndicatorMetadata {
  id: string
  name: string
  type: IndicatorType
  paneRequirement: 'none' | 'shared' | 'dedicated'
  defaultHeight: number   // Suggested height percentage (0-1)
  rangeType: 'fixed' | 'dynamic' | 'price'
  fixedRange?: [number, number]  // For oscillators like RSI [0,100]
  priority?: number       // Display priority
}

export interface LayoutConfig {
  panes: PaneConfig[]
  spacing: number         // Gap between panes (0-0.02)
  autoBalance: boolean    // Auto-redistribute space
}

export interface LayoutPreset {
  name: string
  description: string
  config: {
    main: number
    volume: number
    indicators: { [key: string]: number }
  }
}

export interface UserLayoutPreferences {
  version: string
  timestamp: number
  panes: {
    [paneId: string]: {
      height: number
      visible: boolean
    }
  }
}

export interface PaneResizeHandle {
  paneId: string
  position: number        // Y position (0-1)
  isDragging: boolean
  cursorStyle: 'ns-resize'
  thickness: number       // Pixels for hit area
  visible: boolean       // Show on hover
}

/**
 * Main Layout Manager Class
 */
export class ChartLayoutManager {
  private panes: Map<string, PaneConfig> = new Map()
  private indicators: Map<string, IndicatorMetadata> = new Map()
  private activeIndicators: Set<string> = new Set()
  private defaultVolumeHeight = 0.10  // 10% default (smaller as requested)
  private minPaneHeight = 0.05       // 5% minimum
  private paneSpacing = 0.01         // 1% spacing between panes
  private userPreferences: UserLayoutPreferences | null = null

  // Pane constraints
  private readonly PANE_CONSTRAINTS = {
    main: { min: 0.30, max: 0.90 },      // Main chart 30-90%
    volume: { min: 0.05, max: 0.25 },    // Volume 5-25%
    oscillator: { min: 0.08, max: 0.30 }, // RSI/Stoch 8-30%
    momentum: { min: 0.10, max: 0.30 },  // MACD 10-30%
    volatility: { min: 0.08, max: 0.25 }, // Bollinger Bands width
    custom: { min: 0.05, max: 0.40 }     // Custom indicators
  }

  // Layout presets
  private readonly LAYOUT_PRESETS: LayoutPreset[] = [
    { 
      name: 'Default', 
      description: 'Clean chart with small volume',
      config: { main: 0.86, volume: 0.10, indicators: {} }
    },
    { 
      name: 'Trading', 
      description: 'Optimized for trading with MACD and RSI',
      config: { 
        main: 0.60, 
        volume: 0.10, 
        indicators: { macd: 0.15, rsi: 0.11 }
      }
    },
    { 
      name: 'Analysis', 
      description: 'Detailed analysis view',
      config: { 
        main: 0.65, 
        volume: 0.15, 
        indicators: { macd: 0.08, rsi: 0.08 }
      }
    },
    { 
      name: 'Compact', 
      description: 'Maximum chart space',
      config: { 
        main: 0.80, 
        volume: 0.08, 
        indicators: { combined: 0.08 }
      }
    }
  ]

  constructor() {
    this.initializeDefaultPanes()
    this.loadUserPreferences()
  }

  /**
   * Initialize default pane configurations
   */
  private initializeDefaultPanes(): void {
    // Main price chart pane
    this.panes.set('main', {
      id: 'main',
      type: 'main',
      heightWeight: 100,
      minHeight: this.PANE_CONSTRAINTS.main.min,
      maxHeight: this.PANE_CONSTRAINTS.main.max,
      priority: 1,
      canCombine: false,
      indicators: [],
      resizable: true,
      visible: true
    })

    // Volume pane
    this.panes.set('volume', {
      id: 'volume',
      type: 'volume',
      heightWeight: 15, // Smaller weight for smaller default size
      minHeight: this.PANE_CONSTRAINTS.volume.min,
      maxHeight: this.PANE_CONSTRAINTS.volume.max,
      priority: 2,
      canCombine: false,
      indicators: ['volume'],
      resizable: true,
      visible: true
    })
  }

  /**
   * Register an indicator with its metadata
   */
  registerIndicator(metadata: IndicatorMetadata): void {
    this.indicators.set(metadata.id, metadata)
  }

  /**
   * Add or update a pane configuration
   */
  addPane(config: PaneConfig): void {
    this.panes.set(config.id, config)
  }

  /**
   * Remove a pane
   */
  removePane(paneId: string): void {
    if (paneId !== 'main' && paneId !== 'volume') {
      this.panes.delete(paneId)
    }
  }

  /**
   * Set active indicators
   */
  setActiveIndicators(indicatorIds: string[]): void {
    this.activeIndicators = new Set(indicatorIds)
  }

  /**
   * Calculate optimal layout based on active indicators
   */
  calculateLayout(activeIndicatorIds: string[]): Map<string, ScaleMargins> {
    const layout = new Map<string, ScaleMargins>()
    const activePanes: PaneConfig[] = []

    // Always include main chart
    activePanes.push(this.panes.get('main')!)

    // Determine which panes are needed based on active indicators
    const needsVolume = true // Always show volume pane
    const needsMACD = activeIndicatorIds.includes('macd')
    const needsRSI = activeIndicatorIds.includes('rsi')
    const hasOverlays = activeIndicatorIds.some(id => 
      ['sma20', 'sma50', 'sma200', 'ema20'].includes(id)
    )

    // Add volume pane
    if (needsVolume) {
      activePanes.push(this.panes.get('volume')!)
    }

    // Create or get oscillator panes
    if (needsMACD) {
      if (!this.panes.has('macd')) {
        this.addPane({
          id: 'macd',
          type: 'momentum',
          heightWeight: 20,
          minHeight: this.PANE_CONSTRAINTS.momentum.min,
          maxHeight: this.PANE_CONSTRAINTS.momentum.max,
          priority: 3,
          canCombine: false,
          indicators: ['macd', 'macd_signal', 'macd_histogram'],
          resizable: true,
          visible: true
        })
      }
      activePanes.push(this.panes.get('macd')!)
    }

    if (needsRSI) {
      if (!this.panes.has('rsi')) {
        this.addPane({
          id: 'rsi',
          type: 'oscillator',
          heightWeight: 15,
          minHeight: this.PANE_CONSTRAINTS.oscillator.min,
          maxHeight: this.PANE_CONSTRAINTS.oscillator.max,
          priority: 4,
          canCombine: true,
          indicators: ['rsi'],
          resizable: true,
          visible: true
        })
      }
      activePanes.push(this.panes.get('rsi')!)
    }

    // Calculate heights using dynamic allocation
    const totalWeight = activePanes.reduce((sum, pane) => sum + pane.heightWeight, 0)
    const availableHeight = 1 - (this.paneSpacing * (activePanes.length - 1))

    // Apply user preferences if available
    if (this.userPreferences) {
      this.applyUserPreferences(activePanes)
    }

    // Calculate positions
    let currentTop = 0.02 // Start with small margin

    for (let i = 0; i < activePanes.length; i++) {
      const pane = activePanes[i]
      
      // Calculate height
      let height = pane.currentHeight || (pane.heightWeight / totalWeight * availableHeight)
      
      // Apply constraints
      const constraints = this.PANE_CONSTRAINTS[pane.type]
      height = Math.max(constraints.min, Math.min(constraints.max, height))

      // Set scale margins
      const margins: ScaleMargins = {
        top: currentTop,
        bottom: 1 - (currentTop + height)
      }

      layout.set(pane.id, margins)
      
      // Move to next position
      currentTop += height + this.paneSpacing
    }

    // Ensure last pane reaches bottom
    const lastPaneId = activePanes[activePanes.length - 1].id
    const lastMargins = layout.get(lastPaneId)!
    lastMargins.bottom = 0.02 // Small bottom margin

    return layout
  }

  /**
   * Resize a pane by delta amount
   */
  resizePane(paneId: string, deltaHeight: number): Map<string, ScaleMargins> {
    const pane = this.panes.get(paneId)
    if (!pane || !pane.resizable) {
      return this.calculateLayout(Array.from(this.activeIndicators))
    }

    // Update pane height
    const currentHeight = pane.currentHeight || this.getDefaultHeight(pane)
    const newHeight = Math.max(pane.minHeight, Math.min(pane.maxHeight, currentHeight + deltaHeight))
    pane.currentHeight = newHeight

    // Recalculate layout
    const newLayout = this.calculateLayout(Array.from(this.activeIndicators))
    
    // Save preferences
    this.saveUserPreferences()
    
    return newLayout
  }

  /**
   * Get default height for a pane
   */
  private getDefaultHeight(pane: PaneConfig): number {
    switch (pane.type) {
      case 'main':
        return 0.65
      case 'volume':
        return this.defaultVolumeHeight
      case 'oscillator':
        return 0.12
      case 'momentum':
        return 0.15
      default:
        return 0.10
    }
  }

  /**
   * Apply user preferences to panes
   */
  private applyUserPreferences(panes: PaneConfig[]): void {
    if (!this.userPreferences) return

    for (const pane of panes) {
      const pref = this.userPreferences.panes[pane.id]
      if (pref) {
        pane.currentHeight = pref.height
        pane.visible = pref.visible
      }
    }
  }

  /**
   * Get resize handles for visible panes
   */
  getResizeHandles(): PaneResizeHandle[] {
    const handles: PaneResizeHandle[] = []
    const layout = this.calculateLayout(Array.from(this.activeIndicators))
    
    const paneIds = Array.from(layout.keys())
    for (let i = 0; i < paneIds.length - 1; i++) {
      const margins = layout.get(paneIds[i])!
      handles.push({
        paneId: paneIds[i],
        position: 1 - margins.bottom,
        isDragging: false,
        cursorStyle: 'ns-resize',
        thickness: 4,
        visible: true
      })
    }

    return handles
  }

  /**
   * Get layout preset
   */
  getPreset(name: string): LayoutPreset | undefined {
    return this.LAYOUT_PRESETS.find(preset => preset.name === name)
  }

  /**
   * Apply layout preset
   */
  applyPreset(presetName: string): Map<string, ScaleMargins> {
    const preset = this.getPreset(presetName)
    if (!preset) {
      return this.calculateLayout(Array.from(this.activeIndicators))
    }

    // Reset custom heights
    this.panes.forEach(pane => {
      delete pane.currentHeight
    })

    // Apply preset heights
    const mainPane = this.panes.get('main')
    if (mainPane) mainPane.currentHeight = preset.config.main

    const volumePane = this.panes.get('volume')
    if (volumePane) volumePane.currentHeight = preset.config.volume

    // Apply indicator heights
    for (const [indicatorId, height] of Object.entries(preset.config.indicators)) {
      const pane = this.panes.get(indicatorId)
      if (pane) pane.currentHeight = height
    }

    return this.calculateLayout(Array.from(this.activeIndicators))
  }

  /**
   * Save user preferences to localStorage
   */
  saveUserPreferences(): void {
    const preferences: UserLayoutPreferences = {
      version: '1.0.0',
      timestamp: Date.now(),
      panes: {}
    }

    this.panes.forEach((pane, id) => {
      if (pane.currentHeight !== undefined) {
        preferences.panes[id] = {
          height: pane.currentHeight,
          visible: pane.visible
        }
      }
    })

    try {
      localStorage.setItem('isx-chart-layout-v1', JSON.stringify(preferences))
      this.userPreferences = preferences
    } catch (error) {
      console.error('Failed to save layout preferences:', error)
    }
  }

  /**
   * Load user preferences from localStorage
   */
  loadUserPreferences(): void {
    try {
      const stored = localStorage.getItem('isx-chart-layout-v1')
      if (stored) {
        this.userPreferences = JSON.parse(stored) as UserLayoutPreferences
      }
    } catch (error) {
      console.error('Failed to load layout preferences:', error)
      this.userPreferences = null
    }
  }

  /**
   * Reset to default layout
   */
  resetToDefault(): Map<string, ScaleMargins> {
    this.panes.forEach(pane => {
      delete pane.currentHeight
    })
    this.userPreferences = null
    localStorage.removeItem('isx-chart-layout-v1')
    return this.calculateLayout(Array.from(this.activeIndicators))
  }

  /**
   * Get all available presets
   */
  getPresets(): LayoutPreset[] {
    return [...this.LAYOUT_PRESETS]
  }
}

// Export singleton instance
export const layoutManager = new ChartLayoutManager()