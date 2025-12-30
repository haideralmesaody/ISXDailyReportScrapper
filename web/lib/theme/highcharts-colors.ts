/**
 * Highcharts Theme Colors - Professional Color System
 * 
 * This file provides a complete, professional color system for Highcharts
 * that ensures consistency across light and dark themes while maintaining
 * the Highcharts design language.
 * 
 * Based on official Highcharts palette and best practices.
 */

/**
 * Official Highcharts color palette
 * These are the standard colors used by Highcharts for series
 */
export const HighchartsOfficialPalette = {
  light: [
    '#2caffe', // Bright blue
    '#544fc5', // Purple
    '#00e272', // Green
    '#fe6a35', // Orange
    '#6b8abc', // Grayish blue
    '#d568fb', // Pink
    '#2ee0ca', // Cyan
    '#fa4b42', // Red
    '#feb56a', // Light orange
    '#91e8e1'  // Light cyan
  ],
  dark: [
    '#5679c4', // Adjusted blue for dark
    '#7c6fde', // Lighter purple
    '#2ee0ca', // Cyan (works well in dark)
    '#ff8c5a', // Warmer orange
    '#8fa4c4', // Lighter grayish blue
    '#e08fff', // Lighter pink
    '#4effd8', // Brighter cyan
    '#ff6b6b', // Softer red
    '#ffc285', // Warmer light orange
    '#a5f0ed'  // Brighter light cyan
  ]
}

/**
 * Neutral color scales for consistent grays
 */
export const NeutralScale = {
  light: {
    0: '#ffffff',    // Pure white
    3: '#f7f7f7',    // Near white (Highcharts default)
    5: '#f0f0f0',    // Very light gray
    10: '#e6e6e6',   // Light gray (Highcharts default)
    20: '#cccccc',   // Light border (Highcharts default)
    30: '#b3b3b3',   // Medium light
    40: '#999999',   // Mid gray (Highcharts default)
    50: '#808080',   // True middle
    60: '#666666',   // Dark gray (Highcharts default)
    70: '#4d4d4d',   // Darker gray
    80: '#333333',   // Near black (Highcharts default)
    90: '#1a1a1a',   // Very dark
    100: '#000000'   // Pure black (Highcharts default)
  },
  dark: {
    0: '#0f1419',    // Darkest background
    3: '#1a1f2a',    // Very dark
    5: '#1e2330',    // Dark card
    10: '#2a2f3e',   // Elevated surface
    20: '#3a3f4e',   // Border dark
    30: '#4a4f5e',   // Medium dark
    40: '#5a5f6e',   // Mid gray dark
    50: '#6a6f7e',   // True middle dark
    60: '#8a8f9e',   // Light gray dark
    70: '#9a9fae',   // Lighter gray
    80: '#c0c5d0',   // Near white dark
    90: '#e0e5f0',   // Very light
    100: '#ffffff'   // Pure white
  }
}

/**
 * Complete Highcharts component colors for light theme
 */
export const HighchartsLightTheme = {
  // Chart series colors
  colors: HighchartsOfficialPalette.light,
  
  // Complete popup system
  popup: {
    background: NeutralScale.light[0],
    border: NeutralScale.light[20],
    text: NeutralScale.light[80],
    shadow: '0 0 8px rgba(0, 0, 0, 0.15)',
    header: {
      background: NeutralScale.light[3],
      border: NeutralScale.light[10],
      text: NeutralScale.light[80]
    },
    
    // Input system
    input: {
      background: NeutralScale.light[3],
      border: NeutralScale.light[20],
      text: NeutralScale.light[100],
      placeholder: NeutralScale.light[40],
      focus: {
        border: HighchartsOfficialPalette.light[0]!, // Primary blue
        shadow: `0 0 0 2px ${HighchartsOfficialPalette.light[0]!}33` // 20% alpha
      },
      disabled: {
        background: NeutralScale.light[10],
        border: NeutralScale.light[10],
        text: NeutralScale.light[40],
        opacity: 0.6
      }
    },
    
    // Button system
    button: {
      primary: {
        background: HighchartsOfficialPalette.light[0]!, // #2caffe
        text: NeutralScale.light[0],
        border: HighchartsOfficialPalette.light[0]!,
        hover: '#0099ff', // 10% darker
        active: '#0077cc', // 20% darker
        disabled: {
          background: NeutralScale.light[20],
          text: NeutralScale.light[40],
          border: NeutralScale.light[20]
        }
      },
      secondary: {
        background: NeutralScale.light[3],
        text: NeutralScale.light[80],
        border: NeutralScale.light[20],
        hover: NeutralScale.light[10],
        active: NeutralScale.light[20]
      },
      ghost: {
        background: 'transparent',
        text: HighchartsOfficialPalette.light[0]!,
        border: 'transparent',
        hover: `${HighchartsOfficialPalette.light[0]!}0D`, // 5% alpha
        active: `${HighchartsOfficialPalette.light[0]!}1A` // 10% alpha
      }
    },
    
    // Tabs
    tabs: {
      background: NeutralScale.light[3],
      active: NeutralScale.light[0],
      inactive: NeutralScale.light[3],
      border: NeutralScale.light[10],
      text: NeutralScale.light[60],
      activeText: NeutralScale.light[80],
      hover: NeutralScale.light[10]
    }
  },
  
  // Navigator (zoom/pan control) - Transparent to show volume data
  navigator: {
    maskFill: 'rgba(102, 133, 194, 0.15)',  // Very transparent blue tint
    maskOpacity: 0.15,  // Consistent low opacity
    seriesColor: '#5679c4',
    seriesLineColor: '#5679c4',
    outlineColor: 'rgba(204, 204, 204, 0.5)',  // Semi-transparent outline
    handles: {
      background: 'rgba(247, 247, 247, 0.8)',  // Semi-transparent handle
      border: 'rgba(153, 153, 153, 0.6)',
      symbols: NeutralScale.light[60],
      hover: {
        background: 'rgba(230, 230, 230, 0.9)',
        border: 'rgba(102, 102, 102, 0.8)'
      }
    },
    series: {
      fillOpacity: 0.15,  // Very transparent area fill
      lineWidth: 1,
      color: '#5679c4'
    },
    xAxis: {
      gridLineColor: 'rgba(230, 230, 230, 0.3)',  // Subtle grid lines
      labels: {
        style: {
          color: NeutralScale.light[60]
        }
      }
    }
  },
  
  // Range selector (time period buttons)
  rangeSelector: {
    button: {
      fill: NeutralScale.light[3],
      stroke: NeutralScale.light[20],
      style: {
        color: NeutralScale.light[80],
        fontWeight: '500'
      },
      hover: {
        fill: NeutralScale.light[10],
        stroke: NeutralScale.light[40],
        style: NeutralScale.light[80]
      },
      select: {
        fill: HighchartsOfficialPalette.light[0],
        stroke: HighchartsOfficialPalette.light[0],
        style: NeutralScale.light[0]
      },
      disabled: {
        fill: NeutralScale.light[3],
        stroke: NeutralScale.light[10],
        style: NeutralScale.light[20],
        opacity: 0.5
      }
    },
    input: {
      background: NeutralScale.light[0],
      border: NeutralScale.light[20],
      color: NeutralScale.light[80],
      focus: {
        border: HighchartsOfficialPalette.light[0],
        outline: `0 0 0 2px ${HighchartsOfficialPalette.light[0]}33`
      },
      disabled: {
        background: NeutralScale.light[3],
        color: NeutralScale.light[40]
      }
    },
    inputBoxBorderColor: NeutralScale.light[20],
    inputStyle: {
      color: NeutralScale.light[80]
    },
    labelStyle: {
      color: NeutralScale.light[60]
    }
  },
  
  // Scrollbar
  scrollbar: {
    barBackground: NeutralScale.light[20],
    barBorder: NeutralScale.light[20],
    buttonArrow: NeutralScale.light[60],
    buttonBackground: NeutralScale.light[10],
    buttonBorder: NeutralScale.light[20],
    rifle: NeutralScale.light[60],
    trackBackground: '#f2f2f2',
    trackBorder: '#f2f2f2',
    hover: {
      barBackground: NeutralScale.light[40],
      buttonBackground: NeutralScale.light[20]
    }
  },
  
  // Context menu
  contextMenu: {
    background: NeutralScale.light[0],
    border: NeutralScale.light[20],
    shadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
    separator: NeutralScale.light[10],
    item: {
      background: 'transparent',
      text: NeutralScale.light[80],
      hover: {
        background: NeutralScale.light[3],
        text: NeutralScale.light[100]
      },
      disabled: {
        text: NeutralScale.light[40],
        opacity: 0.5
      }
    }
  },
  
  // Annotations toolbar
  annotations: {
    toolbar: {
      background: NeutralScale.light[0],
      border: NeutralScale.light[20],
      shadow: '0 2px 4px rgba(0, 0, 0, 0.1)'
    },
    handles: {
      fill: HighchartsOfficialPalette.light[0],
      stroke: HighchartsOfficialPalette.light[0],
      hover: {
        fill: '#0099ff',
        stroke: '#0099ff'
      }
    },
    selection: {
      fill: `${HighchartsOfficialPalette.light[0]}1A`, // 10% alpha
      stroke: HighchartsOfficialPalette.light[0],
      opacity: 0.3
    },
    labels: {
      background: NeutralScale.light[0],
      border: NeutralScale.light[20],
      text: NeutralScale.light[80]
    }
  },
  
  // Tooltip
  tooltip: {
    background: 'rgba(255, 255, 255, 0.95)',
    border: NeutralScale.light[20],
    foreground: NeutralScale.light[80],
    shadow: '0 2px 4px rgba(0, 0, 0, 0.2)',
    header: {
      color: NeutralScale.light[100],
      fontWeight: '600'
    },
    crosshair: {
      color: NeutralScale.light[40],
      width: 1,
      dashStyle: 'Solid'
    }
  },
  
  // Loading state
  loading: {
    background: 'rgba(255, 255, 255, 0.85)',
    foreground: NeutralScale.light[60],
    opacity: 0.85
  },
  
  // Zoom controls
  zoom: {
    resetButton: {
      fill: NeutralScale.light[3],
      stroke: NeutralScale.light[20],
      text: NeutralScale.light[80],
      hover: {
        fill: NeutralScale.light[10],
        stroke: NeutralScale.light[40]
      }
    },
    selection: {
      fill: `${HighchartsOfficialPalette.light[0]}1A`, // 10% alpha
      stroke: HighchartsOfficialPalette.light[0],
      opacity: 0.25
    }
  },
  
  // Grid and axes
  grid: {
    line: NeutralScale.light[10],
    minorLine: `${NeutralScale.light[10]}80`, // 50% alpha
    background: 'transparent'
  },
  
  // Axis
  axis: {
    line: NeutralScale.light[20],
    tick: NeutralScale.light[20],
    minorTick: NeutralScale.light[10],
    title: NeutralScale.light[80],
    labels: NeutralScale.light[60]
  },
  
  // Legend
  legend: {
    background: 'rgba(255, 255, 255, 0.9)',
    border: NeutralScale.light[20],
    itemColor: NeutralScale.light[80],
    itemHoverColor: NeutralScale.light[100],
    itemHiddenColor: NeutralScale.light[20],
    title: NeutralScale.light[80],
    shadow: '0 1px 2px rgba(0, 0, 0, 0.1)'
  },
  
  // Credits
  credits: {
    color: NeutralScale.light[40]
  },
  
  // Stock tools
  stockTools: {
    toolbar: {
      background: NeutralScale.light[0],
      border: NeutralScale.light[20]
    },
    button: {
      background: NeutralScale.light[3],
      border: NeutralScale.light[20],
      hover: NeutralScale.light[10],
      active: HighchartsOfficialPalette.light[0]
    }
  }
}

/**
 * Complete Highcharts component colors for dark theme
 */
export const HighchartsDarkTheme = {
  // Chart series colors
  colors: HighchartsOfficialPalette.dark,
  
  // Complete popup system
  popup: {
    background: NeutralScale.dark[5],
    border: NeutralScale.dark[20],
    text: NeutralScale.dark[80],
    shadow: '0 0 12px rgba(0, 0, 0, 0.5)',
    header: {
      background: NeutralScale.dark[10],
      border: NeutralScale.dark[20],
      text: NeutralScale.dark[90]
    },
    
    // Input system
    input: {
      background: NeutralScale.dark[10],
      border: NeutralScale.dark[20],
      text: NeutralScale.dark[80],
      placeholder: NeutralScale.dark[40],
      focus: {
        border: HighchartsOfficialPalette.dark[0]!, // Primary blue for dark
        shadow: `0 0 0 2px ${HighchartsOfficialPalette.dark[0]!}4D` // 30% alpha
      },
      disabled: {
        background: NeutralScale.dark[3],
        border: NeutralScale.dark[10],
        text: NeutralScale.dark[30],
        opacity: 0.5
      }
    },
    
    // Button system  
    button: {
      primary: {
        background: HighchartsOfficialPalette.dark[0]!, // #5679c4
        text: NeutralScale.dark[100],
        border: HighchartsOfficialPalette.dark[0]!,
        hover: '#6689d4', // 10% lighter
        active: '#4669b4', // 10% darker
        disabled: {
          background: NeutralScale.dark[20],
          text: NeutralScale.dark[40],
          border: NeutralScale.dark[20]
        }
      },
      secondary: {
        background: NeutralScale.dark[10],
        text: NeutralScale.dark[80],
        border: NeutralScale.dark[20],
        hover: NeutralScale.dark[20],
        active: NeutralScale.dark[30]
      },
      ghost: {
        background: 'transparent',
        text: HighchartsOfficialPalette.dark[0]!,
        border: 'transparent',
        hover: `${HighchartsOfficialPalette.dark[0]!}1A`, // 10% alpha
        active: `${HighchartsOfficialPalette.dark[0]!}26` // 15% alpha
      }
    },
    
    // Tabs
    tabs: {
      background: NeutralScale.dark[3],
      active: NeutralScale.dark[10],
      inactive: NeutralScale.dark[3],
      border: NeutralScale.dark[20],
      text: NeutralScale.dark[60],
      activeText: NeutralScale.dark[90],
      hover: NeutralScale.dark[10]
    }
  },
  
  // Navigator (zoom/pan control) - Extra transparent for dark theme
  navigator: {
    maskFill: 'rgba(86, 121, 196, 0.1)',  // Even more transparent in dark (10% opacity)
    maskOpacity: 0.1,  // Very low opacity for dark theme
    seriesColor: HighchartsOfficialPalette.dark[0],
    seriesLineColor: HighchartsOfficialPalette.dark[0],
    outlineColor: 'rgba(58, 63, 78, 0.4)',  // Semi-transparent dark outline
    handles: {
      background: 'rgba(42, 47, 62, 0.7)',  // Semi-transparent dark handle
      border: 'rgba(90, 95, 110, 0.5)',
      symbols: NeutralScale.dark[60],
      hover: {
        background: 'rgba(58, 63, 78, 0.8)',
        border: 'rgba(138, 143, 158, 0.7)'
      }
    },
    series: {
      fillOpacity: 0.1,  // Very transparent for dark theme
      lineWidth: 1,
      color: HighchartsOfficialPalette.dark[0]
    },
    xAxis: {
      gridLineColor: 'rgba(58, 63, 78, 0.2)',  // Very subtle grid lines
      labels: {
        style: {
          color: NeutralScale.dark[60]
        }
      }
    }
  },
  
  // Range selector (time period buttons)
  rangeSelector: {
    button: {
      fill: NeutralScale.dark[10],
      stroke: NeutralScale.dark[20],
      style: {
        color: NeutralScale.dark[80],
        fontWeight: '500'
      },
      hover: {
        fill: NeutralScale.dark[20],
        stroke: NeutralScale.dark[40],
        style: NeutralScale.dark[90]
      },
      select: {
        fill: HighchartsOfficialPalette.dark[0],
        stroke: HighchartsOfficialPalette.dark[0],
        style: NeutralScale.dark[100]
      },
      disabled: {
        fill: NeutralScale.dark[5],
        stroke: NeutralScale.dark[10],
        style: NeutralScale.dark[30],
        opacity: 0.3
      }
    },
    input: {
      background: NeutralScale.dark[10],
      border: NeutralScale.dark[20],
      color: NeutralScale.dark[80],
      focus: {
        border: HighchartsOfficialPalette.dark[0],
        outline: `0 0 0 2px ${HighchartsOfficialPalette.dark[0]}4D`
      },
      disabled: {
        background: NeutralScale.dark[5],
        color: NeutralScale.dark[30]
      }
    },
    inputBoxBorderColor: NeutralScale.dark[20],
    inputStyle: {
      color: NeutralScale.dark[80]
    },
    labelStyle: {
      color: NeutralScale.dark[60]
    }
  },
  
  // Scrollbar
  scrollbar: {
    barBackground: NeutralScale.dark[40],
    barBorder: NeutralScale.dark[40],
    buttonArrow: NeutralScale.dark[60],
    buttonBackground: NeutralScale.dark[20],
    buttonBorder: NeutralScale.dark[30],
    rifle: NeutralScale.dark[60],
    trackBackground: NeutralScale.dark[10],
    trackBorder: NeutralScale.dark[10],
    hover: {
      barBackground: NeutralScale.dark[50],
      buttonBackground: NeutralScale.dark[30]
    }
  },
  
  // Context menu
  contextMenu: {
    background: NeutralScale.dark[10],
    border: NeutralScale.dark[20],
    shadow: '0 2px 12px rgba(0, 0, 0, 0.5)',
    separator: NeutralScale.dark[20],
    item: {
      background: 'transparent',
      text: NeutralScale.dark[80],
      hover: {
        background: NeutralScale.dark[20],
        text: NeutralScale.dark[100]
      },
      disabled: {
        text: NeutralScale.dark[40],
        opacity: 0.3
      }
    }
  },
  
  // Annotations toolbar
  annotations: {
    toolbar: {
      background: NeutralScale.dark[10],
      border: NeutralScale.dark[20],
      shadow: '0 2px 8px rgba(0, 0, 0, 0.3)'
    },
    handles: {
      fill: HighchartsOfficialPalette.dark[0],
      stroke: HighchartsOfficialPalette.dark[0],
      hover: {
        fill: '#6689d4',
        stroke: '#6689d4'
      }
    },
    selection: {
      fill: `${HighchartsOfficialPalette.dark[0]}26`, // 15% alpha
      stroke: HighchartsOfficialPalette.dark[0],
      opacity: 0.3
    },
    labels: {
      background: NeutralScale.dark[10],
      border: NeutralScale.dark[20],
      text: NeutralScale.dark[80]
    }
  },
  
  // Tooltip
  tooltip: {
    background: 'rgba(30, 35, 48, 0.95)',
    border: NeutralScale.dark[30],
    foreground: NeutralScale.dark[80],
    shadow: '0 2px 8px rgba(0, 0, 0, 0.4)',
    header: {
      color: NeutralScale.dark[100],
      fontWeight: '600'
    },
    crosshair: {
      color: NeutralScale.dark[40],
      width: 1,
      dashStyle: 'Solid'
    }
  },
  
  // Loading state
  loading: {
    background: 'rgba(15, 20, 25, 0.85)',
    foreground: NeutralScale.dark[60],
    opacity: 0.85
  },
  
  // Zoom controls
  zoom: {
    resetButton: {
      fill: NeutralScale.dark[10],
      stroke: NeutralScale.dark[20],
      text: NeutralScale.dark[80],
      hover: {
        fill: NeutralScale.dark[20],
        stroke: NeutralScale.dark[40]
      }
    },
    selection: {
      fill: `${HighchartsOfficialPalette.dark[0]}26`, // 15% alpha
      stroke: HighchartsOfficialPalette.dark[0],
      opacity: 0.25
    }
  },
  
  // Grid and axes
  grid: {
    line: NeutralScale.dark[20],
    minorLine: `${NeutralScale.dark[20]}40`, // 25% alpha
    background: 'transparent'
  },
  
  // Axis
  axis: {
    line: NeutralScale.dark[30],
    tick: NeutralScale.dark[30],
    minorTick: NeutralScale.dark[20],
    title: NeutralScale.dark[80],
    labels: NeutralScale.dark[60]
  },
  
  // Legend
  legend: {
    background: 'rgba(30, 35, 48, 0.9)',
    border: NeutralScale.dark[20],
    itemColor: NeutralScale.dark[80],
    itemHoverColor: NeutralScale.dark[100],
    itemHiddenColor: NeutralScale.dark[30],
    title: NeutralScale.dark[80],
    shadow: '0 1px 4px rgba(0, 0, 0, 0.3)'
  },
  
  // Credits
  credits: {
    color: NeutralScale.dark[40]
  },
  
  // Stock tools
  stockTools: {
    toolbar: {
      background: NeutralScale.dark[5],
      border: NeutralScale.dark[20]
    },
    button: {
      background: NeutralScale.dark[10],
      border: NeutralScale.dark[20],
      hover: NeutralScale.dark[20],
      active: HighchartsOfficialPalette.dark[0]
    }
  }
}

/**
 * Export theme selector function
 */
export function getHighchartsTheme(isDark: boolean) {
  return isDark ? HighchartsDarkTheme : HighchartsLightTheme
}

/**
 * Export for use in main colors.ts
 */
export const HighchartsThemes = {
  light: HighchartsLightTheme,
  dark: HighchartsDarkTheme
}
