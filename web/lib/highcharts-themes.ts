/**
 * Highcharts Theme Configurations
 * Light and dark themes for professional financial charts
 * 
 * All colors are sourced from the centralized theme colors SSOT.
 * NO hardcoded color values should exist in this file.
 */

import type { Options } from 'highcharts'
import { lightThemeColors, darkThemeColors, type ThemeColors } from './theme/colors'
import { applyThemeCSSVariables } from './theme/css-variables'

/**
 * Create light theme configuration for Highcharts using SSOT colors
 */
function createLightTheme(): Partial<Options> {
  const colors = lightThemeColors
  
  return {
    colors: colors.chart.colors,
    
    chart: {
      backgroundColor: colors.chart.grid.background,
      className: 'highcharts-light',  // Add class for CSS targeting
      style: {
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        color: colors.chart.text.primary
      }
    },
    
    title: {
      style: {
        color: colors.chart.text.primary,
        fontSize: '16px',
        fontWeight: 'bold'
      }
    },
    
    subtitle: {
      style: {
        color: colors.chart.text.secondary
      }
    },
    
    xAxis: {
      gridLineColor: colors.chart.grid.line,
      gridLineWidth: 1,
      lineColor: colors.border,
      tickColor: colors.border,
      labels: {
        style: {
          color: colors.chart.text.secondary
        }
      },
      title: {
        style: {
          color: colors.chart.text.primary
        }
      }
    },
    
    yAxis: {
      gridLineColor: colors.chart.grid.line,
      gridLineWidth: 1,
      lineColor: colors.border,
      tickColor: colors.border,
      labels: {
        style: {
          color: colors.chart.text.secondary
        }
      },
      title: {
        style: {
          color: colors.chart.text.primary
        }
      }
    },
    
    tooltip: {
      backgroundColor: colors.chart.tooltip.background,
      borderColor: colors.chart.tooltip.border,
      style: {
        color: colors.chart.tooltip.foreground
      }
    },
    
    plotOptions: {
      candlestick: {
        lineColor: colors.chart.candlestick.downLine,
        upLineColor: colors.chart.candlestick.upLine,
        color: colors.chart.candlestick.down,
        upColor: colors.chart.candlestick.up
      },
      ohlc: {
        color: colors.chart.candlestick.down,
        upColor: colors.chart.candlestick.up
      }
    },
    
    legend: {
      backgroundColor: colors.chart.legend.background,
      borderColor: colors.chart.legend.border,
      itemStyle: {
        color: colors.chart.legend.foreground
      },
      itemHoverStyle: {
        color: colors.chart.legend.hover
      },
      itemHiddenStyle: {
        color: colors.chart.legend.hidden
      }
    },
    
    credits: {
      enabled: false
    },
    
    navigator: {
      maskFill: colors.chart.navigator.maskFill,
      series: {
        color: colors.chart.navigator.seriesColor,
        lineColor: colors.chart.navigator.seriesLineColor
      },
      xAxis: {
        gridLineColor: colors.chart.grid.line,
        labels: {
          style: {
            color: colors.chart.text.secondary
          }
        }
      }
    },
    
    rangeSelector: {
      buttonTheme: {
        fill: colors.chart.rangeSelector.button.fill,
        stroke: colors.chart.rangeSelector.button.stroke,
        style: {
          color: colors.chart.rangeSelector.button.color
        },
        states: {
          hover: {
            fill: colors.chart.rangeSelector.button.hover.fill,
            stroke: colors.chart.rangeSelector.button.hover.stroke,
            style: {
              color: colors.chart.rangeSelector.button.hover.color
            }
          },
          select: {
            fill: colors.chart.rangeSelector.button.select.fill,
            stroke: colors.chart.rangeSelector.button.select.stroke,
            style: {
              color: colors.chart.rangeSelector.button.select.color
            }
          }
        }
      },
      inputBoxBorderColor: colors.chart.rangeSelector.input.border,
      inputStyle: {
        backgroundColor: colors.chart.rangeSelector.input.background,
        color: colors.chart.rangeSelector.input.color
      },
      labelStyle: {
        color: colors.chart.rangeSelector.label
      }
    },
    
    scrollbar: {
      barBackgroundColor: colors.chart.scrollbar.barBackground,
      barBorderColor: colors.chart.scrollbar.barBorder,
      buttonArrowColor: colors.chart.scrollbar.buttonArrow,
      buttonBackgroundColor: colors.chart.scrollbar.buttonBackground,
      buttonBorderColor: colors.chart.scrollbar.buttonBorder,
      rifleColor: colors.chart.scrollbar.rifle,
      trackBackgroundColor: colors.chart.scrollbar.trackBackground,
      trackBorderColor: colors.chart.scrollbar.trackBorder
    },
    
    navigation: {
      buttonOptions: {
        theme: {
          fill: colors.chart.navigation.button.fill,
          stroke: colors.chart.navigation.button.stroke,
          'stroke-width': 1,
          r: 3,
          style: {
            color: colors.chart.navigation.button.color,
            fontWeight: 'bold'
          },
          states: {
            hover: {
              fill: colors.chart.navigation.button.hover.fill,
              stroke: colors.chart.navigation.button.hover.stroke,
              style: {
                color: colors.chart.navigation.button.hover.color
              }
            },
            select: {
              fill: colors.chart.navigation.button.select.fill,
              stroke: colors.chart.navigation.button.select.stroke,
              style: {
                color: colors.chart.navigation.button.select.color
              }
            }
          }
        },
        symbolFill: colors.chart.navigation.symbol.fill,
        symbolStroke: colors.chart.navigation.symbol.stroke,
        symbolStrokeWidth: 2
      },
      menuStyle: {
        background: colors.chart.navigation.menu.background,
        border: colors.chart.navigation.menu.border,
        borderRadius: '4px',
        padding: '5px',
        boxShadow: '0 2px 5px rgba(0,0,0,0.15)'
      },
      menuItemStyle: {
        color: colors.chart.navigation.menu.itemColor,
        fontSize: '13px',
        padding: '5px 10px',
        cursor: 'pointer'
      },
      menuItemHoverStyle: {
        background: colors.chart.navigation.menu.itemHover,
        color: colors.chart.legend.hover
      }
    } as any
  }
}

/**
 * Light theme configuration for Highcharts
 */
export const lightTheme: Partial<Options> = createLightTheme()

/**
 * Create dark theme configuration for Highcharts using SSOT colors
 */
function createDarkTheme(): Partial<Options> {
  const colors = darkThemeColors
  
  return {
    colors: colors.chart.colors,
    
    chart: {
      backgroundColor: colors.chart.grid.background,
      className: 'highcharts-dark',  // Add class for CSS targeting
      style: {
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
        color: colors.chart.text.primary
      }
    },
    
    title: {
      style: {
        color: colors.chart.text.primary,
        fontSize: '16px',
        fontWeight: 'bold'
      }
    },
    
    subtitle: {
      style: {
        color: colors.chart.text.secondary
      }
    },
    
    xAxis: {
      gridLineColor: colors.chart.grid.line,
      gridLineWidth: 1,
      lineColor: colors.border,
      tickColor: colors.border,
      labels: {
        style: {
          color: colors.chart.text.secondary
        }
      },
      title: {
        style: {
          color: colors.chart.text.primary
        }
      }
    },
    
    yAxis: {
      gridLineColor: colors.chart.grid.line,
      gridLineWidth: 1,
      lineColor: colors.border,
      tickColor: colors.border,
      labels: {
        style: {
          color: colors.chart.text.secondary
        }
      },
      title: {
        style: {
          color: colors.chart.text.primary
        }
      }
    },
    
    tooltip: {
      backgroundColor: colors.chart.tooltip.background,
      borderColor: colors.chart.tooltip.border,
      style: {
        color: colors.chart.tooltip.foreground
      }
    },
    
    plotOptions: {
      candlestick: {
        lineColor: colors.chart.candlestick.downLine,
        upLineColor: colors.chart.candlestick.upLine,
        color: colors.chart.candlestick.down,
        upColor: colors.chart.candlestick.up
      },
      ohlc: {
        color: colors.chart.candlestick.down,
        upColor: colors.chart.candlestick.up
      }
    },
    
    legend: {
      backgroundColor: colors.chart.legend.background,
      borderColor: colors.chart.legend.border,
      itemStyle: {
        color: colors.chart.legend.foreground
      },
      itemHoverStyle: {
        color: colors.chart.legend.hover
      },
      itemHiddenStyle: {
        color: colors.chart.legend.hidden
      }
    },
    
    credits: {
      enabled: false
    },
    
    navigator: {
      maskFill: colors.chart.navigator.maskFill,
      series: {
        color: colors.chart.navigator.seriesColor,
        lineColor: colors.chart.navigator.seriesLineColor
      },
      xAxis: {
        gridLineColor: colors.chart.grid.line,
        labels: {
          style: {
            color: colors.chart.text.secondary
          }
        }
      }
    },
    
    rangeSelector: {
      buttonTheme: {
        fill: colors.chart.rangeSelector.button.fill,
        stroke: colors.chart.rangeSelector.button.stroke,
        style: {
          color: colors.chart.rangeSelector.button.color
        },
        states: {
          hover: {
            fill: colors.chart.rangeSelector.button.hover.fill,
            stroke: colors.chart.rangeSelector.button.hover.stroke,
            style: {
              color: colors.chart.rangeSelector.button.hover.color
            }
          },
          select: {
            fill: colors.chart.rangeSelector.button.select.fill,
            stroke: colors.chart.rangeSelector.button.select.stroke,
            style: {
              color: colors.chart.rangeSelector.button.select.color
            }
          }
        }
      },
      inputBoxBorderColor: colors.chart.rangeSelector.input.border,
      inputStyle: {
        backgroundColor: colors.chart.rangeSelector.input.background,
        color: colors.chart.rangeSelector.input.color
      },
      labelStyle: {
        color: colors.chart.rangeSelector.label
      }
    },
    
    scrollbar: {
      barBackgroundColor: colors.chart.scrollbar.barBackground,
      barBorderColor: colors.chart.scrollbar.barBorder,
      buttonArrowColor: colors.chart.scrollbar.buttonArrow,
      buttonBackgroundColor: colors.chart.scrollbar.buttonBackground,
      buttonBorderColor: colors.chart.scrollbar.buttonBorder,
      rifleColor: colors.chart.scrollbar.rifle,
      trackBackgroundColor: colors.chart.scrollbar.trackBackground,
      trackBorderColor: colors.chart.scrollbar.trackBorder
    },
    
    navigation: {
      buttonOptions: {
        theme: {
          fill: colors.chart.navigation.button.fill,
          stroke: colors.chart.navigation.button.stroke,
          'stroke-width': 1,
          r: 3,
          style: {
            color: colors.chart.navigation.button.color,
            fontWeight: 'bold'
          },
          states: {
            hover: {
              fill: colors.chart.navigation.button.hover.fill,
              stroke: colors.chart.navigation.button.hover.stroke,
              style: {
                color: colors.chart.navigation.button.hover.color
              }
            },
            select: {
              fill: colors.chart.navigation.button.select.fill,
              stroke: colors.chart.navigation.button.select.stroke,
              style: {
                color: colors.chart.navigation.button.select.color
              }
            }
          }
        },
        symbolFill: colors.chart.navigation.symbol.fill,
        symbolStroke: colors.chart.navigation.symbol.stroke,
        symbolStrokeWidth: 2
      },
      menuStyle: {
        background: colors.chart.navigation.menu.background,
        border: colors.chart.navigation.menu.border,
        borderRadius: '4px',
        padding: '5px',
        boxShadow: '0 2px 5px rgba(0,0,0,0.3)'
      },
      menuItemStyle: {
        color: colors.chart.navigation.menu.itemColor,
        fontSize: '13px',
        padding: '5px 10px',
        cursor: 'pointer'
      },
      menuItemHoverStyle: {
        background: colors.chart.navigation.menu.itemHover,
        color: colors.chart.legend.hover
      }
    } as any
  }
}

/**
 * Dark theme configuration for Highcharts
 */
export const darkTheme: Partial<Options> = createDarkTheme()

/**
 * Apply theme to Highcharts instance using SSOT colors
 * This function also updates CSS variables for dynamic theme switching
 */
export function applyHighchartsTheme(Highcharts: any, theme: 'light' | 'dark') {
  const themeOptions = theme === 'dark' ? darkTheme : lightTheme
  const colors = theme === 'dark' ? darkThemeColors : lightThemeColors
  
  // NO LONGER apply CSS variables - they're defined in globals.css
  // This prevents CSS pollution and theme conflicts
  // applyThemeCSSVariables(theme) // REMOVED
  
  // Apply the theme globally to Highcharts
  Highcharts.setOptions(themeOptions)
  // Stock tools configuration removed - handled in highcharts-loader.ts
  // This prevents duplicate configuration and ensures single source of truth
}