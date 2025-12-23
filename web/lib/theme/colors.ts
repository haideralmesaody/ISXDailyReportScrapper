/**
 * Theme Colors - Single Source of Truth (SSOT)
 * 
 * This file defines ALL color values used throughout the application.
 * It serves as the central repository for theme colors to ensure consistency
 * and maintainability across all components, charts, and UI elements.
 * 
 * NO HARDCODED COLORS should exist elsewhere in the codebase - all colors
 * must be imported from this file.
 */

import { HighchartsThemes } from './highcharts-colors'

/**
 * Base color palette - Raw hex values for reference
 */
export const BaseColors = {
  // Iraqi Investor Brand Colors (Primary Palette)
  forestGreen: '#2d5a3d',      // Primary brand color
  darkForestGreen: '#1a3d2b',  // Secondary brand color
  bullRed: '#c53030',          // Accent/bull market color
  
  // Success & Status Colors
  successGreen: '#38a169',     // Success states
  warningGold: '#d69e2e',      // Warning states
  errorRed: '#e53e3e',         // Error states
  infoBlue: '#3182ce',         // Information states
  
  // Neutral Colors (Light Theme)
  lightBackground: '#f8f6f0',  // Logo background off-white
  lightForeground: '#1a202c',  // Dark text
  lightCardBg: '#ffffff',      // Card backgrounds
  lightBorder: '#e2e8f0',      // Borders and dividers
  lightMuted: '#718096',       // Muted text
  lightSubtle: '#f7fafc',      // Subtle backgrounds
  
  // Neutral Colors (Dark Theme)
  darkBackground: '#0f1419',   // Dark background
  darkForeground: '#e2e8f0',   // Light text
  darkCardBg: '#1a202c',       // Dark card backgrounds
  darkBorder: '#2d3748',       // Dark borders
  darkMuted: '#a0aec0',        // Dark muted text
  darkSubtle: '#2d3748',       // Dark subtle backgrounds
  
  // Chart-Specific Colors
  candlestickUp: '#6FB76F',    // Green for price increases
  candlestickDown: '#FF6F6F',  // Red for price decreases
  volumeBar: '#60a5fa',        // Volume bar color
  gridLine: '#e6e6e6',         // Light mode grid lines
  darkGridLine: '#404040',     // Dark mode grid lines
  
  // Highcharts Standard Palette
  chartBlue: '#2caffe',
  chartPurple: '#544fc5',
  chartGreen: '#00e272',
  chartOrange: '#fe6a35',
  chartGrayBlue: '#6b8abc',
  chartPink: '#d568fb',
  chartCyan: '#2ee0ca',
  chartRed: '#fa4b42',
  chartLightOrange: '#feb56a',
  chartLightCyan: '#91e8e1',
} as const

/**
 * HSL Color Values
 * These match the CSS custom properties in globals.css
 */
export const HSLColors = {
  // Light Theme HSL Values
  light: {
    background: '40 6% 97%',        // #f8f6f0
    foreground: '222.2 84% 4.9%',   // #1a202c
    card: '40 6% 97%',              // #f8f6f0
    cardForeground: '222.2 84% 4.9%',
    popover: '40 6% 97%',           // #f8f6f0
    popoverForeground: '222.2 84% 4.9%',
    primary: '143 32% 26%',         // #2d5a3d
    primaryForeground: '210 40% 98%',
    secondary: '138 40% 18%',       // #1a3d2b
    secondaryForeground: '210 40% 98%',
    muted: '210 40% 96%',
    mutedForeground: '215.4 16.3% 46.9%',
    accent: '0 62% 49%',            // #c53030
    accentForeground: '210 40% 98%',
    destructive: '0 84.2% 60.2%',
    destructiveForeground: '210 40% 98%',
    border: '214.3 31.8% 91.4%',
    input: '214.3 31.8% 91.4%',
    ring: '143 32% 26%',            // Focus ring
    success: '142 71% 45%',         // Success green
    warning: '38 92% 50%',          // Warning gold
    error: '0 84% 60%',             // Error red
    info: '199 89% 48%',            // Info blue
  },
  
  // Dark Theme HSL Values
  dark: {
    background: '222.2 84% 4.9%',
    foreground: '210 40% 98%',
    card: '222.2 84% 4.9%',
    cardForeground: '210 40% 98%',
    popover: '222.2 84% 4.9%',
    popoverForeground: '210 40% 98%',
    primary: '143 50% 45%',         // Lighter green for dark mode
    primaryForeground: '222.2 84% 4.9%',
    secondary: '217.2 32.6% 17.5%',
    secondaryForeground: '210 40% 98%',
    muted: '217.2 32.6% 17.5%',
    mutedForeground: '215 20.2% 65.1%',
    accent: '0 75% 60%',            // Lighter red for dark mode
    accentForeground: '210 40% 98%',
    destructive: '0 62.8% 30.6%',
    destructiveForeground: '210 40% 98%',
    border: '217.2 32.6% 17.5%',
    input: '217.2 32.6% 17.5%',
    ring: '143 50% 45%',            // Lighter focus ring
    success: '142 71% 55%',         // Lighter success green
    warning: '38 92% 60%',          // Lighter warning gold
    error: '0 84% 70%',             // Lighter error red
    info: '199 89% 58%',            // Lighter info blue
  }
} as const

/**
 * Complete theme colors interface
 * This defines the structure for all theme-related colors
 */
export interface ThemeColors {
  // Base Theme Colors
  background: string
  foreground: string
  card: string
  cardForeground: string
  popover: string
  popoverForeground: string
  primary: string
  primaryForeground: string
  secondary: string
  secondaryForeground: string
  muted: string
  mutedForeground: string
  accent: string
  accentForeground: string
  destructive: string
  destructiveForeground: string
  border: string
  input: string
  ring: string
  
  // Status Colors
  success: string
  warning: string
  error: string
  info: string
  
  // Chart Colors
  chart: {
    // Primary chart color palette
    colors: string[]
    
    // Candlestick/OHLC colors
    candlestick: {
      up: string
      down: string
      upLine: string
      downLine: string
    }
    
    // Grid and axes
    grid: {
      line: string
      background: string
    }
    
    // Text and labels
    text: {
      primary: string
      secondary: string
      muted: string
    }
    
    // Tooltip
    tooltip: {
      background: string
      foreground: string
      border: string
    }
    
    // Legend
    legend: {
      background: string
      foreground: string
      border: string
      hover: string
      hidden: string
    }
    
    // Navigator
    navigator: {
      maskFill: string
      seriesColor: string
      seriesLineColor: string
    }
    
    // Range selector
    rangeSelector: {
      button: {
        fill: string
        stroke: string
        color: string
        hover: {
          fill: string
          stroke: string
          color: string
        }
        select: {
          fill: string
          stroke: string
          color: string
        }
      }
      input: {
        background: string
        color: string
        border: string
      }
      label: string
    }
    
    // Scrollbar
    scrollbar: {
      barBackground: string
      barBorder: string
      buttonArrow: string
      buttonBackground: string
      buttonBorder: string
      rifle: string
      trackBackground: string
      trackBorder: string
    }
    
    // Navigation menu
    navigation: {
      button: {
        fill: string
        stroke: string
        color: string
        hover: {
          fill: string
          stroke: string
          color: string
        }
        select: {
          fill: string
          stroke: string
          color: string
        }
      }
      menu: {
        background: string
        border: string
        itemColor: string
        itemHover: string
      }
      symbol: {
        fill: string
        stroke: string
      }
    }
    
    // Popup and context menu colors (expanded for complete coverage)
    popup: {
      background: string
      border: string
      text: string
      shadow: string
      header: {
        background: string
        border: string
        text: string
      }
      input: {
        background: string
        border: string
        text: string
        placeholder: string
        focus: {
          border: string
          shadow: string
        }
        disabled: {
          background: string
          border: string
          text: string
          opacity: number
        }
      }
      button: {
        primary: {
          background: string
          text: string
          border: string
          hover: string
          active: string
          disabled: {
            background: string
            text: string
            border: string
          }
        }
        secondary: {
          background: string
          text: string
          border: string
          hover: string
          active: string
        }
        ghost: {
          background: string
          text: string
          border: string
          hover: string
          active: string
        }
      }
      tabs: {
        background: string
        active: string
        inactive: string
        border: string
        text: string
        activeText: string
        hover: string
      }
    }
  }
  
  // UI Component Colors
  components: {
    // Buttons
    button: {
      primary: {
        background: string
        foreground: string
        hover: string
      }
      secondary: {
        background: string
        foreground: string
        hover: string
      }
      outline: {
        background: string
        foreground: string
        border: string
        hover: {
          background: string
          foreground: string
        }
      }
      ghost: {
        hover: {
          background: string
          foreground: string
        }
      }
    }
    
    // Form elements
    form: {
      input: {
        background: string
        foreground: string
        border: string
        placeholder: string
        focus: {
          ring: string
          border: string
        }
      }
      select: {
        background: string
        foreground: string
        border: string
      }
      checkbox: {
        background: string
        foreground: string
        border: string
      }
    }
    
    // Dropdown menus
    dropdown: {
      background: string
      foreground: string
      border: string
      item: {
        hover: string
        selected: string
      }
    }
    
    // Modals and dialogs
    modal: {
      background: string
      foreground: string
      overlay: string
      border: string
    }
    
    // Progress bars
    progress: {
      background: string
      foreground: string
      success: string
      warning: string
      error: string
    }
    
    // Data tables
    table: {
      header: {
        background: string
        foreground: string
      }
      row: {
        even: string
        odd: string
        hover: string
      }
      border: string
    }
    
    // Badges and pills
    badge: {
      primary: {
        background: string
        foreground: string
      }
      secondary: {
        background: string
        foreground: string
      }
      success: {
        background: string
        foreground: string
      }
      warning: {
        background: string
        foreground: string
      }
      error: {
        background: string
        foreground: string
      }
    }
  }
  
  // Market-specific colors
  market: {
    positive: string
    negative: string
    neutral: string
    positiveBackground: string
    negativeBackground: string
  }
}

/**
 * Light theme color configuration
 */
export const lightThemeColors: ThemeColors = {
  // Base colors
  background: BaseColors.lightBackground,
  foreground: BaseColors.lightForeground,
  card: BaseColors.lightCardBg,
  cardForeground: BaseColors.lightForeground,
  popover: BaseColors.lightCardBg,
  popoverForeground: BaseColors.lightForeground,
  primary: BaseColors.forestGreen,
  primaryForeground: '#ffffff',
  secondary: BaseColors.darkForestGreen,
  secondaryForeground: '#ffffff',
  muted: '#f7fafc',
  mutedForeground: BaseColors.lightMuted,
  accent: BaseColors.bullRed,
  accentForeground: '#ffffff',
  destructive: BaseColors.errorRed,
  destructiveForeground: '#ffffff',
  border: BaseColors.lightBorder,
  input: BaseColors.lightBorder,
  ring: BaseColors.forestGreen,
  
  // Status colors
  success: BaseColors.successGreen,
  warning: BaseColors.warningGold,
  error: BaseColors.errorRed,
  info: BaseColors.infoBlue,
  
  // Chart colors
  chart: {
    colors: [
      BaseColors.chartBlue,
      BaseColors.chartPurple,
      BaseColors.chartGreen,
      BaseColors.chartOrange,
      BaseColors.chartGrayBlue,
      BaseColors.chartPink,
      BaseColors.chartCyan,
      BaseColors.chartRed,
      BaseColors.chartLightOrange,
      BaseColors.chartLightCyan,
    ],
    
    candlestick: {
      up: BaseColors.candlestickUp,
      down: BaseColors.candlestickDown,
      upLine: BaseColors.candlestickUp,
      downLine: BaseColors.candlestickDown,
    },
    
    grid: {
      line: BaseColors.gridLine,
      background: 'transparent',
    },
    
    text: {
      primary: '#333333',
      secondary: '#666666',
      muted: '#999999',
    },
    
    tooltip: {
      background: 'rgba(255, 255, 255, 0.95)',
      foreground: '#333333',
      border: '#cccccc',
    },
    
    legend: {
      background: 'rgba(255, 255, 255, 0.9)',
      foreground: '#333333',
      border: '#cccccc',
      hover: '#000000',
      hidden: '#cccccc',
    },
    
    navigator: {
      maskFill: 'rgba(102, 133, 194, 0.3)',
      seriesColor: '#5679c4',
      seriesLineColor: '#5679c4',
    },
    
    rangeSelector: {
      button: {
        fill: '#f7f7f7',
        stroke: '#cccccc',
        color: '#333333',
        hover: {
          fill: '#e6e6e6',
          stroke: '#333333',
          color: '#333333',
        },
        select: {
          fill: '#0066cc',
          stroke: '#0066cc',
          color: '#ffffff',
        },
      },
      input: {
        background: '#ffffff',
        color: '#333333',
        border: '#cccccc',
      },
      label: '#666666',
    },
    
    scrollbar: {
      barBackground: '#cccccc',
      barBorder: '#cccccc',
      buttonArrow: '#666666',
      buttonBackground: '#e6e6e6',
      buttonBorder: '#cccccc',
      rifle: '#666666',
      trackBackground: '#f2f2f2',
      trackBorder: '#f2f2f2',
    },
    
    navigation: {
      button: {
        fill: '#f7f7f7',
        stroke: '#cccccc',
        color: '#333333',
        hover: {
          fill: '#e6e6e6',
          stroke: '#333333',
          color: '#333333',
        },
        select: {
          fill: '#e6f2ff',
          stroke: '#2f7ed8',
          color: '#2f7ed8',
        },
      },
      menu: {
        background: '#ffffff',
        border: '1px solid #cccccc',
        itemColor: '#333333',
        itemHover: '#f0f0f0',
      },
      symbol: {
        fill: '#333333',
        stroke: '#333333',
      },
    },
    
    // Popup and context menu - Using Highcharts professional theme
    popup: HighchartsThemes.light.popup,
  },
  
  // UI components
  components: {
    button: {
      primary: {
        background: BaseColors.forestGreen,
        foreground: '#ffffff',
        hover: BaseColors.forestGreen + 'e6', // 90% opacity
      },
      secondary: {
        background: BaseColors.darkForestGreen,
        foreground: '#ffffff',
        hover: BaseColors.darkForestGreen + 'cc', // 80% opacity
      },
      outline: {
        background: 'transparent',
        foreground: BaseColors.forestGreen,
        border: BaseColors.lightBorder,
        hover: {
          background: BaseColors.bullRed,
          foreground: '#ffffff',
        },
      },
      ghost: {
        hover: {
          background: BaseColors.bullRed,
          foreground: '#ffffff',
        },
      },
    },
    
    form: {
      input: {
        background: '#ffffff',
        foreground: BaseColors.lightForeground,
        border: BaseColors.lightBorder,
        placeholder: BaseColors.lightMuted,
        focus: {
          ring: BaseColors.forestGreen,
          border: BaseColors.forestGreen,
        },
      },
      select: {
        background: '#ffffff',
        foreground: BaseColors.lightForeground,
        border: BaseColors.lightBorder,
      },
      checkbox: {
        background: '#ffffff',
        foreground: BaseColors.forestGreen,
        border: BaseColors.lightBorder,
      },
    },
    
    dropdown: {
      background: '#ffffff',
      foreground: BaseColors.lightForeground,
      border: BaseColors.lightBorder,
      item: {
        hover: '#f8f9fa',
        selected: BaseColors.lightSubtle,
      },
    },
    
    modal: {
      background: '#ffffff',
      foreground: BaseColors.lightForeground,
      overlay: 'rgba(0, 0, 0, 0.5)',
      border: BaseColors.lightBorder,
    },
    
    progress: {
      background: '#f1f5f9',
      foreground: BaseColors.forestGreen,
      success: BaseColors.successGreen,
      warning: BaseColors.warningGold,
      error: BaseColors.errorRed,
    },
    
    table: {
      header: {
        background: '#f8f9fa',
        foreground: BaseColors.lightForeground,
      },
      row: {
        even: '#ffffff',
        odd: '#f8f9fa',
        hover: '#f1f5f9',
      },
      border: BaseColors.lightBorder,
    },
    
    badge: {
      primary: {
        background: BaseColors.forestGreen,
        foreground: '#ffffff',
      },
      secondary: {
        background: BaseColors.lightMuted,
        foreground: '#ffffff',
      },
      success: {
        background: BaseColors.successGreen,
        foreground: '#ffffff',
      },
      warning: {
        background: BaseColors.warningGold,
        foreground: '#ffffff',
      },
      error: {
        background: BaseColors.errorRed,
        foreground: '#ffffff',
      },
    },
  },
  
  // Market colors
  market: {
    positive: '#16a34a', // Green-600
    negative: '#dc2626', // Red-600
    neutral: BaseColors.lightMuted,
    positiveBackground: '#dcfce7', // Green-100
    negativeBackground: '#fee2e2', // Red-100
  },
}

/**
 * Dark theme color configuration
 */
export const darkThemeColors: ThemeColors = {
  // Base colors
  background: BaseColors.darkBackground,
  foreground: BaseColors.darkForeground,
  card: BaseColors.darkCardBg,
  cardForeground: BaseColors.darkForeground,
  popover: BaseColors.darkCardBg,
  popoverForeground: BaseColors.darkForeground,
  primary: '#4ade80', // Lighter green for dark mode
  primaryForeground: BaseColors.darkBackground,
  secondary: '#374151',
  secondaryForeground: BaseColors.darkForeground,
  muted: BaseColors.darkSubtle,
  mutedForeground: BaseColors.darkMuted,
  accent: '#f87171', // Lighter red for dark mode
  accentForeground: BaseColors.darkForeground,
  destructive: '#ef4444',
  destructiveForeground: BaseColors.darkForeground,
  border: BaseColors.darkBorder,
  input: BaseColors.darkBorder,
  ring: '#4ade80', // Lighter focus ring
  
  // Status colors
  success: '#22c55e', // Lighter success green
  warning: '#eab308', // Lighter warning gold
  error: '#ef4444', // Lighter error red
  info: '#3b82f6', // Lighter info blue
  
  // Chart colors
  chart: {
    colors: [
      BaseColors.chartBlue,
      BaseColors.chartPurple,
      BaseColors.chartGreen,
      BaseColors.chartOrange,
      BaseColors.chartGrayBlue,
      BaseColors.chartPink,
      BaseColors.chartCyan,
      BaseColors.chartRed,
      BaseColors.chartLightOrange,
      BaseColors.chartLightCyan,
    ],
    
    candlestick: {
      up: BaseColors.candlestickUp,
      down: BaseColors.candlestickDown,
      upLine: BaseColors.candlestickUp,
      downLine: BaseColors.candlestickDown,
    },
    
    grid: {
      line: BaseColors.darkGridLine,
      background: 'transparent',
    },
    
    text: {
      primary: '#e0e0e0',
      secondary: '#a0a0a0',
      muted: '#808080',
    },
    
    tooltip: {
      background: 'rgba(30, 30, 30, 0.95)',
      foreground: '#e0e0e0',
      border: '#505050',
    },
    
    legend: {
      background: 'rgba(30, 30, 30, 0.9)',
      foreground: '#e0e0e0',
      border: '#505050',
      hover: '#ffffff',
      hidden: '#606060',
    },
    
    navigator: {
      maskFill: 'rgba(96, 165, 250, 0.3)',
      seriesColor: '#60a5fa',
      seriesLineColor: '#60a5fa',
    },
    
    rangeSelector: {
      button: {
        fill: '#2a2a2a',
        stroke: '#505050',
        color: '#e0e0e0',
        hover: {
          fill: '#3a3a3a',
          stroke: '#606060',
          color: '#ffffff',
        },
        select: {
          fill: '#0066cc',
          stroke: '#0066cc',
          color: '#ffffff',
        },
      },
      input: {
        background: '#2a2a2a',
        color: '#e0e0e0',
        border: '#505050',
      },
      label: '#a0a0a0',
    },
    
    scrollbar: {
      barBackground: '#505050',
      barBorder: '#505050',
      buttonArrow: '#a0a0a0',
      buttonBackground: '#3a3a3a',
      buttonBorder: '#505050',
      rifle: '#a0a0a0',
      trackBackground: '#2a2a2a',
      trackBorder: '#2a2a2a',
    },
    
    navigation: {
      button: {
        fill: '#2a2a2a',
        stroke: '#505050',
        color: '#e0e0e0',
        hover: {
          fill: '#3a3a3a',
          stroke: '#606060',
          color: '#ffffff',
        },
        select: {
          fill: '#1a3d60',
          stroke: '#2f7ed8',
          color: '#60a5fa',
        },
      },
      menu: {
        background: '#2a2a2a',
        border: '1px solid #505050',
        itemColor: '#e0e0e0',
        itemHover: '#3a3a3a',
      },
      symbol: {
        fill: '#e0e0e0',
        stroke: '#e0e0e0',
      },
    },
    
    // Popup and context menu - Using Highcharts professional theme (dark)
    popup: HighchartsThemes.dark.popup,
  },
  
  // UI components
  components: {
    button: {
      primary: {
        background: '#4ade80',
        foreground: BaseColors.darkBackground,
        hover: '#22c55e',
      },
      secondary: {
        background: '#374151',
        foreground: BaseColors.darkForeground,
        hover: '#4b5563',
      },
      outline: {
        background: 'transparent',
        foreground: '#4ade80',
        border: BaseColors.darkBorder,
        hover: {
          background: '#f87171',
          foreground: BaseColors.darkForeground,
        },
      },
      ghost: {
        hover: {
          background: '#f87171',
          foreground: BaseColors.darkForeground,
        },
      },
    },
    
    form: {
      input: {
        background: BaseColors.darkCardBg,
        foreground: BaseColors.darkForeground,
        border: BaseColors.darkBorder,
        placeholder: BaseColors.darkMuted,
        focus: {
          ring: '#4ade80',
          border: '#4ade80',
        },
      },
      select: {
        background: BaseColors.darkCardBg,
        foreground: BaseColors.darkForeground,
        border: BaseColors.darkBorder,
      },
      checkbox: {
        background: BaseColors.darkCardBg,
        foreground: '#4ade80',
        border: BaseColors.darkBorder,
      },
    },
    
    dropdown: {
      background: BaseColors.darkCardBg,
      foreground: BaseColors.darkForeground,
      border: BaseColors.darkBorder,
      item: {
        hover: '#374151',
        selected: BaseColors.darkSubtle,
      },
    },
    
    modal: {
      background: BaseColors.darkCardBg,
      foreground: BaseColors.darkForeground,
      overlay: 'rgba(0, 0, 0, 0.8)',
      border: BaseColors.darkBorder,
    },
    
    progress: {
      background: '#374151',
      foreground: '#4ade80',
      success: '#22c55e',
      warning: '#eab308',
      error: '#ef4444',
    },
    
    table: {
      header: {
        background: '#374151',
        foreground: BaseColors.darkForeground,
      },
      row: {
        even: BaseColors.darkCardBg,
        odd: '#1f2937',
        hover: '#374151',
      },
      border: BaseColors.darkBorder,
    },
    
    badge: {
      primary: {
        background: '#4ade80',
        foreground: BaseColors.darkBackground,
      },
      secondary: {
        background: BaseColors.darkMuted,
        foreground: BaseColors.darkBackground,
      },
      success: {
        background: '#22c55e',
        foreground: BaseColors.darkBackground,
      },
      warning: {
        background: '#eab308',
        foreground: BaseColors.darkBackground,
      },
      error: {
        background: '#ef4444',
        foreground: BaseColors.darkForeground,
      },
    },
  },
  
  // Market colors
  market: {
    positive: '#22c55e', // Green-500
    negative: '#ef4444', // Red-500
    neutral: BaseColors.darkMuted,
    positiveBackground: '#14532d', // Green-900
    negativeBackground: '#7f1d1d', // Red-900
  },
}

/**
 * Get theme colors based on theme name
 */
export function getThemeColors(theme: 'light' | 'dark'): ThemeColors {
  return theme === 'dark' ? darkThemeColors : lightThemeColors
}

/**
 * Export theme colors as CSS custom properties
 */
export function getThemeCSSProperties(theme: 'light' | 'dark'): Record<string, string> {
  return theme === 'dark' ? HSLColors.dark : HSLColors.light
}

/**
 * Type exports for convenience
 */
export type Theme = 'light' | 'dark'
export type ChartColors = ThemeColors['chart']
export type ComponentColors = ThemeColors['components']
export type MarketColors = ThemeColors['market']