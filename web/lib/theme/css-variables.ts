/**
 * CSS Variable Management System for Dynamic Theme Switching
 * 
 * This system provides functions to generate and inject CSS custom properties
 * at runtime to support dynamic theme switching. It maps theme colors from
 * the centralized SSOT to CSS variables used throughout the application,
 * including Highcharts-specific variables.
 */

import { lightThemeColors, darkThemeColors, getThemeColors, type ThemeColors, type Theme } from './colors'

/**
 * CSS Variable mapping for Highcharts components
 * These correspond to the variables used in highcharts-gui.css and highcharts-popup.css
 */
export interface HighchartsCSSVariables {
  // Background colors
  '--highcharts-background-color': string
  
  // Neutral colors (grayscale palette)
  '--highcharts-neutral-color-3': string   // Light backgrounds (#f7f7f7)
  '--highcharts-neutral-color-10': string  // Hover states (#e6e6e6)
  '--highcharts-neutral-color-20': string  // Borders, disabled states (#cccccc)
  '--highcharts-neutral-color-40': string  // Input borders (#999999)
  '--highcharts-neutral-color-60': string  // Text, labels (#666666)
  '--highcharts-neutral-color-80': string  // Dark text (#333333)
  '--highcharts-neutral-color-100': string // Pure black text (#000000)
  
  // Highlight colors
  '--highcharts-highlight-color-80': string // Links, primary actions (#335cad)
}

/**
 * Extended CSS Variables for complete theme management
 * Includes both Highcharts and general application variables
 */
export interface AllCSSVariables extends HighchartsCSSVariables {
  // Base theme variables (matching Tailwind CSS structure)
  '--background': string
  '--foreground': string
  '--card': string
  '--card-foreground': string
  '--popover': string
  '--popover-foreground': string
  '--primary': string
  '--primary-foreground': string
  '--secondary': string
  '--secondary-foreground': string
  '--muted': string
  '--muted-foreground': string
  '--accent': string
  '--accent-foreground': string
  '--destructive': string
  '--destructive-foreground': string
  '--border': string
  '--input': string
  '--ring': string
  
  // Status colors
  '--success': string
  '--warning': string
  '--error': string
  '--info': string
  
  // Chart-specific variables
  '--chart-grid-line': string
  '--chart-text-primary': string
  '--chart-text-secondary': string
  '--chart-text-muted': string
  '--chart-tooltip-background': string
  '--chart-tooltip-foreground': string
  '--chart-tooltip-border': string
  
  // Market colors
  '--market-positive': string
  '--market-negative': string
  '--market-neutral': string
  '--market-positive-background': string
  '--market-negative-background': string
}

/**
 * Generate Highcharts-specific CSS variables from theme colors
 */
export function generateHighchartsCSSVariables(colors: ThemeColors): HighchartsCSSVariables & Record<string, string> {
  return {
    // Background color - main container background
    '--highcharts-background-color': colors.background,
    
    // Neutral color scale for Highcharts UI components
    '--highcharts-neutral-color-3': colors.chart.navigation.button.fill,
    '--highcharts-neutral-color-10': colors.chart.navigation.button.hover.fill,
    '--highcharts-neutral-color-20': colors.border,
    '--highcharts-neutral-color-40': colors.chart.rangeSelector.input.border,
    '--highcharts-neutral-color-60': colors.chart.text.secondary,
    '--highcharts-neutral-color-80': colors.chart.text.primary,
    '--highcharts-neutral-color-100': colors.foreground,
    
    // Highlight color for interactive elements
    '--highcharts-highlight-color-80': colors.primary,
    
    // Popup and menu specific variables (critical for dark mode)
    '--highcharts-popup-background': colors.chart.popup.background,
    '--highcharts-popup-border': colors.chart.popup.border,
    '--highcharts-popup-text': colors.chart.popup.text,
    '--highcharts-popup-shadow': colors.chart.popup.shadow,
    
    // Popup header
    '--highcharts-popup-header-bg': colors.chart.popup.header?.background || colors.chart.popup.background,
    '--highcharts-popup-header-border': colors.chart.popup.header?.border || colors.chart.popup.border,
    '--highcharts-popup-header-text': colors.chart.popup.header?.text || colors.chart.popup.text,
    
    // Input fields
    '--highcharts-input-background': colors.chart.popup.input.background,
    '--highcharts-input-border': colors.chart.popup.input.border,
    '--highcharts-input-text': colors.chart.popup.input.text,
    '--highcharts-input-placeholder': colors.chart.popup.input.placeholder || colors.mutedForeground,
    '--highcharts-input-focus-border': colors.chart.popup.input.focus?.border || colors.primary,
    '--highcharts-input-focus-shadow': colors.chart.popup.input.focus?.shadow || 'none',
    '--highcharts-input-disabled-bg': colors.chart.popup.input.disabled?.background || colors.chart.popup.input.background,
    '--highcharts-input-disabled-text': colors.chart.popup.input.disabled?.text || colors.mutedForeground,
    
    // Button variables
    '--highcharts-button-primary-bg': colors.chart.popup.button?.primary?.background || colors.primary,
    '--highcharts-button-primary-text': colors.chart.popup.button?.primary?.text || '#ffffff',
    '--highcharts-button-primary-border': colors.chart.popup.button?.primary?.border || colors.primary,
    '--highcharts-button-primary-hover': colors.chart.popup.button?.primary?.hover || colors.primary,
    '--highcharts-button-primary-active': colors.chart.popup.button?.primary?.active || colors.primary,
    '--highcharts-button-secondary-bg': colors.chart.popup.button?.secondary?.background || colors.muted,
    '--highcharts-button-secondary-text': colors.chart.popup.button?.secondary?.text || colors.foreground,
    '--highcharts-button-secondary-hover': colors.chart.popup.button?.secondary?.hover || colors.muted,
    
    // Tab variables
    '--highcharts-tab-bg': colors.chart.popup.tabs?.background || colors.muted,
    '--highcharts-tab-active': colors.chart.popup.tabs?.active || colors.background,
    '--highcharts-tab-inactive': colors.chart.popup.tabs?.inactive || colors.muted,
    '--highcharts-tab-border': colors.chart.popup.tabs?.border || colors.border,
    '--highcharts-tab-text': colors.chart.popup.tabs?.text || colors.mutedForeground,
    '--highcharts-tab-active-text': colors.chart.popup.tabs?.activeText || colors.foreground,
    '--highcharts-tab-hover': colors.chart.popup.tabs?.hover || colors.muted,
    
    // Additional theme variables for overrides
    '--highcharts-background-secondary': colors.card,
    '--highcharts-background-hover': colors.muted,
    '--highcharts-accent-color': colors.primary,
    '--highcharts-accent-light': colors.primaryForeground,
    '--highcharts-border-color': colors.border,
    '--highcharts-text-color': colors.foreground,
    '--highcharts-text-muted': colors.mutedForeground,
    '--highcharts-text-inverse': colors.background,
    '--highcharts-shadow': 'rgba(0, 0, 0, 0.15)',
    '--highcharts-icon-filter': colors.foreground === '#ffffff' ? 'invert(1)' : 'none',
  }
}

/**
 * Generate complete CSS variables from theme colors
 */
export function generateAllCSSVariables(colors: ThemeColors): AllCSSVariables {
  const highchartsVars = generateHighchartsCSSVariables(colors)
  
  return {
    ...highchartsVars,
    
    // Base theme variables
    '--background': colors.background,
    '--foreground': colors.foreground,
    '--card': colors.card,
    '--card-foreground': colors.cardForeground,
    '--popover': colors.popover,
    '--popover-foreground': colors.popoverForeground,
    '--primary': colors.primary,
    '--primary-foreground': colors.primaryForeground,
    '--secondary': colors.secondary,
    '--secondary-foreground': colors.secondaryForeground,
    '--muted': colors.muted,
    '--muted-foreground': colors.mutedForeground,
    '--accent': colors.accent,
    '--accent-foreground': colors.accentForeground,
    '--destructive': colors.destructive,
    '--destructive-foreground': colors.destructiveForeground,
    '--border': colors.border,
    '--input': colors.input,
    '--ring': colors.ring,
    
    // Status colors
    '--success': colors.success,
    '--warning': colors.warning,
    '--error': colors.error,
    '--info': colors.info,
    
    // Chart-specific variables
    '--chart-grid-line': colors.chart.grid.line,
    '--chart-text-primary': colors.chart.text.primary,
    '--chart-text-secondary': colors.chart.text.secondary,
    '--chart-text-muted': colors.chart.text.muted,
    '--chart-tooltip-background': colors.chart.tooltip.background,
    '--chart-tooltip-foreground': colors.chart.tooltip.foreground,
    '--chart-tooltip-border': colors.chart.tooltip.border,
    
    // Market colors
    '--market-positive': colors.market.positive,
    '--market-negative': colors.market.negative,
    '--market-neutral': colors.market.neutral,
    '--market-positive-background': colors.market.positiveBackground,
    '--market-negative-background': colors.market.negativeBackground,
  }
}

/**
 * Inject CSS variables into the document root
 * This function updates the :root element with new CSS custom properties
 */
export function injectCSSVariables(variables: Partial<AllCSSVariables>): void {
  if (typeof document === 'undefined') {
    // Skip in server-side rendering
    return
  }
  
  const root = document.documentElement
  
  Object.entries(variables).forEach(([property, value]) => {
    root.style.setProperty(property, value)
  })
}

/**
 * Remove CSS variables from the document root
 * Useful for cleanup or resetting to default values
 */
export function removeCSSVariables(variableNames: string[]): void {
  if (typeof document === 'undefined') {
    // Skip in server-side rendering
    return
  }
  
  const root = document.documentElement
  
  variableNames.forEach((property) => {
    root.style.removeProperty(property)
  })
}

/**
 * Apply theme CSS variables to the document
 * DEPRECATED: We now use CSS variables defined in globals.css
 * This function is kept for backward compatibility but does nothing
 * @deprecated Use CSS variables directly from globals.css
 */
export function applyThemeCSSVariables(theme: Theme): void {
  // NO-OP: We no longer inject CSS variables dynamically
  // All theme colors are now defined in globals.css with --hc- prefix
  // This prevents CSS pollution and theme conflicts
  console.log(`Theme request for ${theme} ignored - using CSS from globals.css`)
}

/**
 * Get current CSS variable value from the document
 * Useful for debugging or getting computed values
 */
export function getCSSVariableValue(variableName: string): string {
  if (typeof document === 'undefined') {
    return ''
  }
  
  return getComputedStyle(document.documentElement)
    .getPropertyValue(variableName)
    .trim()
}

/**
 * Check if CSS variables are supported in the current browser
 */
export function supportsCSSVariables(): boolean {
  if (typeof window === 'undefined') {
    return false
  }
  
  return window.CSS && window.CSS.supports && window.CSS.supports('color', 'var(--test)')
}

/**
 * Create a CSS string from variables object
 * Useful for server-side rendering or creating stylesheets
 */
export function createCSSVariablesString(variables: Partial<AllCSSVariables>): string {
  const cssRules = Object.entries(variables)
    .map(([property, value]) => `  ${property}: ${value};`)
    .join('\n')
  
  return `:root {\n${cssRules}\n}`
}

/**
 * Batch update CSS variables for better performance
 * Groups DOM updates into a single operation
 */
export function batchUpdateCSSVariables(updates: Array<{ variables: Partial<AllCSSVariables> }>): void {
  if (typeof document === 'undefined') {
    return
  }
  
  const root = document.documentElement
  
  // Use requestAnimationFrame to batch DOM updates
  requestAnimationFrame(() => {
    updates.forEach(({ variables }) => {
      Object.entries(variables).forEach(([property, value]) => {
        root.style.setProperty(property, value)
      })
    })
  })
}

/**
 * Watch for system theme changes and automatically update CSS variables
 * Returns a cleanup function to remove the listener
 */
export function watchSystemTheme(
  onThemeChange: (theme: Theme) => void,
  autoApply: boolean = true
): () => void {
  if (typeof window === 'undefined') {
    return () => {}
  }
  
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  
  const handler = (e: MediaQueryListEvent) => {
    const theme: Theme = e.matches ? 'dark' : 'light'
    onThemeChange(theme)
    
    if (autoApply) {
      applyThemeCSSVariables(theme)
    }
  }
  
  // Add listener
  mediaQuery.addEventListener('change', handler)
  
  // Return cleanup function
  return () => {
    mediaQuery.removeEventListener('change', handler)
  }
}

/**
 * Read Highcharts CSS variables from the document
 * These are defined in globals.css with --hc- prefix
 */
export function getHighchartsColors(): Record<string, string> {
  if (typeof document === 'undefined') {
    return {}
  }
  
  const root = getComputedStyle(document.documentElement)
  
  return {
    background: root.getPropertyValue('--hc-background').trim() || 'transparent',
    gridBackground: root.getPropertyValue('--hc-grid-background').trim(),
    gridLine: root.getPropertyValue('--hc-grid-line').trim(),
    axisLine: root.getPropertyValue('--hc-axis-line').trim(),
    candlestickUp: root.getPropertyValue('--hc-candlestick-up').trim(),
    candlestickDown: root.getPropertyValue('--hc-candlestick-down').trim(),
    candlestickUpLine: root.getPropertyValue('--hc-candlestick-up-line').trim(),
    candlestickDownLine: root.getPropertyValue('--hc-candlestick-down-line').trim(),
    volumeColor: root.getPropertyValue('--hc-volume-color').trim(),
    tooltipBg: root.getPropertyValue('--hc-tooltip-bg').trim(),
    tooltipBorder: root.getPropertyValue('--hc-tooltip-border').trim(),
    tooltipText: root.getPropertyValue('--hc-tooltip-text').trim(),
    textPrimary: root.getPropertyValue('--hc-text-primary').trim(),
    textSecondary: root.getPropertyValue('--hc-text-secondary').trim(),
    textMuted: root.getPropertyValue('--hc-text-muted').trim(),
    buttonFill: root.getPropertyValue('--hc-button-fill').trim(),
    buttonHover: root.getPropertyValue('--hc-button-hover').trim(),
    buttonText: root.getPropertyValue('--hc-button-text').trim(),
    rangeSelectorBg: root.getPropertyValue('--hc-range-selector-bg').trim(),
    rangeSelectorBorder: root.getPropertyValue('--hc-range-selector-border').trim(),
    rangeSelectorText: root.getPropertyValue('--hc-range-selector-text').trim(),
    navigatorSeries: root.getPropertyValue('--hc-navigator-series').trim(),
    navigatorMask: root.getPropertyValue('--hc-navigator-mask').trim(),
    scrollbarTrack: root.getPropertyValue('--hc-scrollbar-track').trim(),
    scrollbarThumb: root.getPropertyValue('--hc-scrollbar-thumb').trim(),
    indicatorSMA: root.getPropertyValue('--hc-indicator-sma').trim(),
    indicatorEMA: root.getPropertyValue('--hc-indicator-ema').trim(),
    indicatorBollinger: root.getPropertyValue('--hc-indicator-bollinger').trim(),
    indicatorMACD: root.getPropertyValue('--hc-indicator-macd').trim(),
    indicatorRSI: root.getPropertyValue('--hc-indicator-rsi').trim(),
  }
}

/**
 * Export utility functions for use in other modules
 */
export const cssVariables = {
  generate: generateAllCSSVariables,
  generateHighcharts: generateHighchartsCSSVariables,
  inject: injectCSSVariables,
  remove: removeCSSVariables,
  apply: applyThemeCSSVariables,
  getValue: getCSSVariableValue,
  getHighchartsColors,  // New function to read Highcharts colors
  supports: supportsCSSVariables,
  createString: createCSSVariablesString,
  batchUpdate: batchUpdateCSSVariables,
  watchSystem: watchSystemTheme,
} as const

/**
 * Default export for easy importing
 */
export default cssVariables