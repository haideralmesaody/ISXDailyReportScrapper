# CSS Variable Management System

This directory contains a comprehensive CSS variable management system for dynamic theme switching in the ISX Daily Reports Scrapper application.

## Files Overview

### Core Files

- **`colors.ts`** - Single Source of Truth (SSOT) for all theme colors
- **`css-variables.ts`** - CSS variable generation and injection system
- **`usage-examples.ts`** - Comprehensive usage examples and patterns

### Related Files

- **`../highcharts-themes.ts`** - Highcharts theme configuration (updated to use CSS variables)

## Key Features

### 1. Dynamic Theme Switching
```typescript
import { applyThemeCSSVariables } from './css-variables'

// Switch to dark theme
applyThemeCSSVariables('dark')

// Switch to light theme
applyThemeCSSVariables('light')
```

### 2. Highcharts Integration
The system automatically updates all CSS custom properties used by Highcharts:

- `--highcharts-background-color`
- `--highcharts-neutral-color-*` (3, 10, 20, 40, 60, 80, 100)
- `--highcharts-highlight-color-80`

### 3. Automatic System Theme Detection
```typescript
import { watchSystemTheme } from './css-variables'

// Auto-follow system theme changes
const cleanup = watchSystemTheme((theme) => {
  console.log(`Theme changed to: ${theme}`)
})
```

### 4. Performance Optimized
- Batched DOM updates using `requestAnimationFrame`
- Efficient CSS variable injection
- Minimal DOM manipulation

## CSS Variables Generated

### Base Theme Variables
- `--background`, `--foreground`
- `--card`, `--card-foreground`
- `--primary`, `--primary-foreground`
- `--secondary`, `--secondary-foreground`
- `--muted`, `--muted-foreground`
- `--accent`, `--accent-foreground`
- `--destructive`, `--destructive-foreground`
- `--border`, `--input`, `--ring`

### Status Colors
- `--success`, `--warning`, `--error`, `--info`

### Chart-Specific Variables
- `--chart-grid-line`
- `--chart-text-primary`, `--chart-text-secondary`, `--chart-text-muted`
- `--chart-tooltip-background`, `--chart-tooltip-foreground`, `--chart-tooltip-border`

### Market Colors
- `--market-positive`, `--market-negative`, `--market-neutral`
- `--market-positive-background`, `--market-negative-background`

### Highcharts Variables
All variables used in `highcharts-gui.css` and `highcharts-popup.css`:
- `--highcharts-background-color`
- `--highcharts-neutral-color-3` through `--highcharts-neutral-color-100`
- `--highcharts-highlight-color-80`

## Usage Patterns

### Basic Theme Switch
```typescript
import { applyThemeCSSVariables } from './css-variables'

function toggleTheme(currentTheme: 'light' | 'dark') {
  const newTheme = currentTheme === 'light' ? 'dark' : 'light'
  applyThemeCSSVariables(newTheme)
  return newTheme
}
```

### React Hook Integration
```typescript
import { useState, useEffect } from 'react'
import { applyThemeCSSVariables, watchSystemTheme } from './css-variables'

export function useTheme() {
  const [theme, setTheme] = useState<'light' | 'dark'>('light')
  
  useEffect(() => {
    applyThemeCSSVariables(theme)
  }, [theme])
  
  useEffect(() => {
    return watchSystemTheme(setTheme)
  }, [])
  
  return { theme, setTheme }
}
```

### Highcharts Integration
```typescript
import { applyHighchartsTheme } from '../highcharts-themes'
import Highcharts from 'highcharts'

// The applyHighchartsTheme function now automatically applies CSS variables
applyHighchartsTheme(Highcharts, 'dark')
```

### Server-Side Rendering
```typescript
import { generateThemeCSS } from './usage-examples'

// In _document.tsx
const themeCSS = generateThemeCSS('light')
// Inject into <Head> as <style> tag
```

## Browser Support

The system includes automatic detection for CSS custom property support:

```typescript
import { supportsCSSVariables } from './css-variables'

if (supportsCSSVariables()) {
  // Use CSS variables
  applyThemeCSSVariables('dark')
} else {
  // Fallback for older browsers
  // The CSS files already include fallback values
}
```

## Performance Considerations

### Batch Updates
When making multiple theme changes, use batch updates:

```typescript
import { batchUpdateCSSVariables } from './css-variables'

batchUpdateCSSVariables([
  { variables: baseThemeVars },
  { variables: customOverrides }
])
```

### Request Animation Frame
All DOM updates are automatically batched using `requestAnimationFrame` for smooth transitions.

## Development & Debugging

### Debug Utilities
```typescript
import { debugUtils } from './usage-examples'

// Log all current CSS variables
debugUtils.logCurrentVariables()

// Preview theme without applying
const darkVars = debugUtils.previewTheme('dark')

// Check specific variable value
const primaryColor = debugUtils.getVariable('--primary')

// Check browser support
debugUtils.checkSupport()
```

## Architecture Benefits

1. **Type Safety** - All variables and colors are strongly typed
2. **Single Source of Truth** - All colors centralized in `colors.ts`
3. **Runtime Updates** - Themes can be switched without page reload
4. **Consistency** - Same color system across all components and charts
5. **Performance** - Optimized DOM manipulation and batched updates
6. **Extensibility** - Easy to add new variables or themes
7. **SSR Support** - Can generate CSS strings for server-side rendering

## Integration Points

This system integrates with:

- **Highcharts** - Via `applyHighchartsTheme` function
- **Tailwind CSS** - Via CSS custom properties
- **Next.js** - SSR support and runtime switching
- **React** - Hook patterns and component integration
- **System Preferences** - Automatic detection and following

## Testing

All files type-check successfully:
```bash
npx tsc --noEmit lib/theme/css-variables.ts
npx tsc --noEmit lib/theme/colors.ts  
npx tsc --noEmit lib/highcharts-themes.ts
```

The system is ready for production use and provides a robust foundation for dynamic theming throughout the application.