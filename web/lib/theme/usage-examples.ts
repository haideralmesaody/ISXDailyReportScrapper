/**
 * CSS Variables System - Usage Examples
 * 
 * This file demonstrates how to use the CSS variable management system
 * for dynamic theme switching in various scenarios.
 */

import { 
  applyThemeCSSVariables,
  generateAllCSSVariables,
  generateHighchartsCSSVariables,
  injectCSSVariables,
  watchSystemTheme,
  cssVariables 
} from './css-variables'
import { getThemeColors } from './colors'

/**
 * Example 1: Basic theme switching
 * Apply all CSS variables for a theme
 */
export function switchToLightTheme() {
  applyThemeCSSVariables('light')
}

export function switchToDarkTheme() {
  applyThemeCSSVariables('dark')
}

/**
 * Example 2: Partial theme updates
 * Update only specific CSS variables without changing everything
 */
export function updateHighchartsColors(theme: 'light' | 'dark') {
  const colors = getThemeColors(theme)
  const highchartsVars = generateHighchartsCSSVariables(colors)
  injectCSSVariables(highchartsVars)
}

/**
 * Example 3: Custom color overrides
 * Override specific variables while maintaining theme consistency
 */
export function applyCustomAccentColor(accentColor: string) {
  injectCSSVariables({
    '--accent': accentColor,
    '--highcharts-highlight-color-80': accentColor,
  })
}

/**
 * Example 4: React Hook for theme switching
 * Example of how you might use this in a React component
 */
export function useThemeSwitcher() {
  const switchTheme = (theme: 'light' | 'dark') => {
    // Apply CSS variables
    applyThemeCSSVariables(theme)
    
    // Store preference (you would implement this)
    localStorage.setItem('theme', theme)
    
    // Update document class for CSS selectors
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }
  
  const initializeTheme = () => {
    // Check saved preference or system preference
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null
    const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    const initialTheme = savedTheme || systemTheme
    
    switchTheme(initialTheme)
  }
  
  return { switchTheme, initializeTheme }
}

/**
 * Example 5: System theme auto-switching
 * Automatically follow system theme changes
 */
export function enableAutoThemeSwitching() {
  return watchSystemTheme((theme) => {
    console.log(`System theme changed to: ${theme}`)
    // The watchSystemTheme function already applies the theme automatically
    // You can add additional logic here if needed
  })
}

/**
 * Example 6: Development/debugging utilities
 * Helper functions for development and debugging
 */
export const debugUtils = {
  // Log all current CSS variables
  logCurrentVariables() {
    const root = document.documentElement
    const computedStyle = getComputedStyle(root)
    const variables: Record<string, string> = {}
    
    // Get all CSS properties that start with --
    Array.from(document.styleSheets)
      .flatMap(sheet => {
        try {
          return Array.from(sheet.cssRules)
        } catch {
          return []
        }
      })
      .forEach(rule => {
        if (rule instanceof CSSStyleRule && rule.selectorText === ':root') {
          Array.from(rule.style).forEach(prop => {
            if (prop.startsWith('--')) {
              variables[prop] = computedStyle.getPropertyValue(prop).trim()
            }
          })
        }
      })
    
    console.table(variables)
  },
  
  // Preview theme without applying it
  previewTheme(theme: 'light' | 'dark') {
    const colors = getThemeColors(theme)
    const variables = generateAllCSSVariables(colors)
    return variables
  },
  
  // Get specific variable value
  getVariable(name: string) {
    return cssVariables.getValue(name)
  },
  
  // Check if CSS variables are supported
  checkSupport() {
    const supported = cssVariables.supports()
    console.log(`CSS Variables supported: ${supported}`)
    return supported
  }
}

/**
 * Example 7: Highcharts integration
 * How to use with Highcharts charts
 */
export function setupHighchartsTheme(Highcharts: any, initialTheme: 'light' | 'dark' = 'light') {
  // Apply initial theme
  applyThemeCSSVariables(initialTheme)
  
  // Set up automatic theme switching for new charts
  const originalChart = Highcharts.Chart
  Highcharts.Chart = function(renderTo: any, options: any, callback?: any) {
    // Ensure CSS variables are applied before creating chart
    const currentTheme = document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    applyThemeCSSVariables(currentTheme)
    
    return originalChart.call(this, renderTo, options, callback)
  }
  
  // Return cleanup function
  return () => {
    Highcharts.Chart = originalChart
  }
}

/**
 * Example 8: Performance optimized batch updates
 * When changing multiple themes or doing complex updates
 */
export function performanceOptimizedThemeSwitch(
  theme: 'light' | 'dark',
  customOverrides?: Record<string, string>
) {
  const colors = getThemeColors(theme)
  const baseVariables = generateAllCSSVariables(colors)
  const finalVariables = { ...baseVariables, ...customOverrides }
  
  // Use batch update for better performance
  cssVariables.batchUpdate([{ variables: finalVariables }])
}

/**
 * Example 9: Server-side rendering support
 * Generate CSS string for SSR
 */
export function generateThemeCSS(theme: 'light' | 'dark'): string {
  const colors = getThemeColors(theme)
  const variables = generateAllCSSVariables(colors)
  return cssVariables.createString(variables)
}

/**
 * Example usage in a Next.js _document.tsx:
 * 
 * export default function Document() {
 *   return (
 *     <Html>
 *       <Head>
 *         <style
 *           dangerouslySetInnerHTML={{
 *             __html: generateThemeCSS('light')
 *           }}
 *         />
 *       </Head>
 *       <body>
 *         <Main />
 *         <NextScript />
 *       </body>
 *     </Html>
 *   )
 * }
 */