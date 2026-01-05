import { createChart, IChartApi, ChartOptions, DeepPartial, ColorType } from 'lightweight-charts'

// Theme color definitions matching ISX Pulse design system
export interface ThemeColors {
  background: string
  grid: {
    vertLines: string
    horzLines: string
  }
  crosshair: {
    color: string
  }
  timeScale: {
    background: string
    textColor: string
    borderColor: string
  }
  priceScale: {
    background: string
    textColor: string
    borderColor: string
  }
  candlestick: {
    upColor: string
    downColor: string
    borderUpColor: string
    borderDownColor: string
    wickUpColor: string
    wickDownColor: string
  }
  volume: {
    upColor: string
    downColor: string
  }
  watermark: {
    color: string
    background: string
  }
  layout: {
    background: ColorType
    textColor: string
    fontSize: number
    fontFamily: string
  }
}

/**
 * Get theme colors for light and dark modes
 * Colors are based on CSS variables and Tailwind design system
 */
export function getThemeColors(theme: 'light' | 'dark'): ThemeColors {
  if (theme === 'dark') {
    return {
      background: 'rgb(3 7 18)', // bg-slate-950
      grid: {
        vertLines: 'rgba(148, 163, 184, 0.1)', // slate-400/10
        horzLines: 'rgba(148, 163, 184, 0.1)', // slate-400/10
      },
      crosshair: {
        color: 'rgb(100, 116, 139)', // slate-500
      },
      timeScale: {
        background: 'rgb(15 23 42)', // bg-slate-900
        textColor: 'rgb(148, 163, 184)', // text-slate-400
        borderColor: 'rgba(148, 163, 184, 0.2)', // border-slate-400/20
      },
      priceScale: {
        background: 'rgb(15 23 42)', // bg-slate-900
        textColor: 'rgb(148, 163, 184)', // text-slate-400
        borderColor: 'rgba(148, 163, 184, 0.2)', // border-slate-400/20
      },
      candlestick: {
        upColor: 'rgb(34, 197, 94)', // green-500
        downColor: 'rgb(239, 68, 68)', // red-500
        borderUpColor: 'rgb(21, 128, 61)', // green-700
        borderDownColor: 'rgb(185, 28, 28)', // red-700
        wickUpColor: 'rgb(21, 128, 61)', // green-700
        wickDownColor: 'rgb(185, 28, 28)', // red-700
      },
      volume: {
        upColor: 'rgba(34, 197, 94, 0.5)', // green-500/50
        downColor: 'rgba(239, 68, 68, 0.5)', // red-500/50
      },
      watermark: {
        color: 'rgba(148, 163, 184, 0.1)', // slate-400/10
        background: 'transparent',
      },
      layout: {
        background: ColorType.Solid,
        textColor: 'rgb(148, 163, 184)', // text-slate-400
        fontSize: 12,
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      },
    }
  } else {
    return {
      background: 'rgb(248 250 252)', // bg-slate-50
      grid: {
        vertLines: 'rgba(71, 85, 105, 0.1)', // slate-600/10
        horzLines: 'rgba(71, 85, 105, 0.1)', // slate-600/10
      },
      crosshair: {
        color: 'rgb(100, 116, 139)', // slate-500
      },
      timeScale: {
        background: 'rgb(255, 255, 255)', // bg-white
        textColor: 'rgb(71, 85, 105)', // text-slate-600
        borderColor: 'rgba(71, 85, 105, 0.2)', // border-slate-600/20
      },
      priceScale: {
        background: 'rgb(255, 255, 255)', // bg-white
        textColor: 'rgb(71, 85, 105)', // text-slate-600
        borderColor: 'rgba(71, 85, 105, 0.2)', // border-slate-600/20
      },
      candlestick: {
        upColor: 'rgb(34, 197, 94)', // green-500
        downColor: 'rgb(239, 68, 68)', // red-500
        borderUpColor: 'rgb(21, 128, 61)', // green-700
        borderDownColor: 'rgb(185, 28, 28)', // red-700
        wickUpColor: 'rgb(21, 128, 61)', // green-700
        wickDownColor: 'rgb(185, 28, 28)', // red-700
      },
      volume: {
        upColor: 'rgba(34, 197, 94, 0.5)', // green-500/50
        downColor: 'rgba(239, 68, 68, 0.5)', // red-500/50
      },
      watermark: {
        color: 'rgba(71, 85, 105, 0.1)', // slate-600/10
        background: 'transparent',
      },
      layout: {
        background: ColorType.Solid,
        textColor: 'rgb(71, 85, 105)', // text-slate-600
        fontSize: 12,
        fontFamily: 'ui-sans-serif, system-ui, sans-serif',
      },
    }
  }
}

/**
 * Create chart options with theme-specific settings
 */
export function createChartOptions(theme: 'light' | 'dark'): DeepPartial<ChartOptions> {
  const colors = getThemeColors(theme)

  return {
    layout: {
      background: {
        type: colors.layout.background,
        color: colors.background,
      },
      textColor: colors.layout.textColor,
      fontSize: colors.layout.fontSize,
      fontFamily: colors.layout.fontFamily,
    },
    grid: {
      vertLines: {
        color: colors.grid.vertLines,
        style: 0, // Solid
      },
      horzLines: {
        color: colors.grid.horzLines,
        style: 0, // Solid
      },
    },
    crosshair: {
      mode: 1, // Normal crosshair mode
      vertLine: {
        color: colors.crosshair.color,
        width: 1,
        style: 3, // Dashed
        visible: true,
        labelVisible: true,
      },
      horzLine: {
        color: colors.crosshair.color,
        width: 1,
        style: 3, // Dashed
        visible: true,
        labelVisible: true,
      },
    },
    timeScale: {
      background: colors.timeScale.background,
      borderColor: colors.timeScale.borderColor,
      textColor: colors.timeScale.textColor,
      timeVisible: true,
      secondsVisible: false,
      rightOffset: 10,
      barSpacing: 6,
      fixLeftEdge: false,
      lockVisibleTimeRangeOnResize: true,
      rightBarStaysOnScroll: true,
      borderVisible: true,
      visible: true,
    },
    rightPriceScale: {
      background: colors.priceScale.background,
      borderColor: colors.priceScale.borderColor,
      textColor: colors.priceScale.textColor,
      scaleMargins: {
        top: 0.3,
        bottom: 0.25,
      },
      borderVisible: true,
      visible: true,
      entireTextOnly: false,
      ticksVisible: true,
      alignLabels: true,
      mode: 0, // Normal price scale mode
      invertScale: false,
    },
    leftPriceScale: {
      background: colors.priceScale.background,
      borderColor: colors.priceScale.borderColor,
      textColor: colors.priceScale.textColor,
      visible: false,
    },
    overlayPriceScales: {
      // Configuration for overlay price scales
    },
    watermark: {
      visible: false, // Can be enabled per-chart basis
      fontSize: 48,
      horzAlign: 'center',
      vertAlign: 'center',
      color: colors.watermark.color,
      text: '',
    },
    handleScroll: {
      mouseWheel: true,
      pressedMouseMove: true,
      horzTouchDrag: true,
      vertTouchDrag: true,
    },
    handleScale: {
      axisPressedMouseMove: true,
      mouseWheel: true,
      pinch: true,
      axisDoubleClickReset: true,
    },
    kineticScroll: {
      mouse: false,
      touch: true,
    },
    trackingMode: {
      exitMode: 1, // OnNextTap
    },
    localization: {
      locale: 'en-US',
      priceFormatter: (price: number) => {
        // Format prices in Iraqi Dinar
        return new Intl.NumberFormat('en-US', {
          style: 'currency',
          currency: 'IQD',
          minimumFractionDigits: 2,
          maximumFractionDigits: 3,
        }).format(price)
      },
      timeFormatter: (time: number) => {
        // Format time for Iraqi timezone
        return new Intl.DateTimeFormat('en-US', {
          timeZone: 'Asia/Baghdad',
          hour: '2-digit',
          minute: '2-digit',
        }).format(new Date(time * 1000))
      },
    },
  }
}

/**
 * Initialize chart with proper configuration and theme
 */
export async function initializeLightweightChart(
  container: HTMLElement,
  theme: 'light' | 'dark'
): Promise<IChartApi> {
  try {
    const options = createChartOptions(theme)
    const chart = createChart(container, options)

    // Apply responsive settings based on container size
    const containerRect = container.getBoundingClientRect()
    applyResponsiveSettings(chart, containerRect.width)

    // Set up resize observer for automatic responsive updates
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect
        chart.applyOptions({
          width: Math.floor(width),
          height: Math.floor(height),
        })
        applyResponsiveSettings(chart, width)
      }
    })

    resizeObserver.observe(container)

    // Store resize observer on chart for cleanup
    ;(chart as any)._resizeObserver = resizeObserver

    return chart
  } catch (error) {
    console.error('Failed to initialize Lightweight Charts:', error)
    throw new Error(`Chart initialization failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
  }
}

/**
 * Switch chart theme dynamically
 */
export function switchChartTheme(chart: IChartApi, theme: 'light' | 'dark'): void {
  try {
    const options = createChartOptions(theme)
    chart.applyOptions(options)

    // Trigger a redraw to ensure theme is fully applied
    chart.timeScale().fitContent()
  } catch (error) {
    console.error('Failed to switch chart theme:', error)
  }
}

/**
 * Apply responsive settings based on container width
 */
export function applyResponsiveSettings(chart: IChartApi, width: number): void {
  try {
    let rightOffset = 10
    let barSpacing = 6
    let fontSize = 12

    // Adjust settings based on screen size
    if (width < 480) {
      // Mobile
      rightOffset = 5
      barSpacing = 3
      fontSize = 10
    } else if (width < 768) {
      // Tablet
      rightOffset = 7
      barSpacing = 4
      fontSize = 11
    } else if (width < 1024) {
      // Small desktop
      rightOffset = 8
      barSpacing = 5
      fontSize = 12
    }

    chart.applyOptions({
      timeScale: {
        rightOffset,
        barSpacing,
      },
      layout: {
        fontSize,
      },
    })

    // Adjust price scale settings for mobile
    if (width < 768) {
      chart.applyOptions({
        rightPriceScale: {
          scaleMargins: {
            top: 0.2,
            bottom: 0.2,
          },
        },
      })
    }
  } catch (error) {
    console.error('Failed to apply responsive settings:', error)
  }
}

/**
 * Cleanup chart resources
 */
export function cleanupChart(chart: IChartApi): void {
  try {
    // Cleanup resize observer
    const resizeObserver = (chart as any)._resizeObserver
    if (resizeObserver) {
      resizeObserver.disconnect()
      delete (chart as any)._resizeObserver
    }

    // Remove chart
    chart.remove()
  } catch (error) {
    console.error('Failed to cleanup chart:', error)
  }
}

/**
 * Create default candlestick series options
 */
export function createCandlestickSeriesOptions(theme: 'light' | 'dark') {
  const colors = getThemeColors(theme)

  return {
    upColor: colors.candlestick.upColor,
    downColor: colors.candlestick.downColor,
    borderUpColor: colors.candlestick.borderUpColor,
    borderDownColor: colors.candlestick.borderDownColor,
    wickUpColor: colors.candlestick.wickUpColor,
    wickDownColor: colors.candlestick.wickDownColor,
    borderVisible: true,
    wickVisible: true,
    priceFormat: {
      type: 'price',
      precision: 3,
      minMove: 0.001,
    },
  }
}

/**
 * Create default volume series options
 */
export function createVolumeSeriesOptions(theme: 'light' | 'dark') {
  const colors = getThemeColors(theme)

  return {
    color: colors.volume.upColor,
    priceFormat: {
      type: 'volume',
    },
    priceScaleId: 'volume',
    scaleMargins: {
      top: 0.8,
      bottom: 0,
    },
  }
}

/**
 * Create technical indicator line series options
 */
export function createLineSeriesOptions(
  theme: 'light' | 'dark',
  color?: string,
  lineWidth = 2
) {
  return {
    color: color || getThemeColors(theme).layout.textColor,
    lineWidth,
    lineType: 0, // Simple line
    lineStyle: 0, // Solid
    crosshairMarkerVisible: true,
    crosshairMarkerRadius: 3,
    crosshairMarkerBorderColor: color || getThemeColors(theme).layout.textColor,
    crosshairMarkerBackgroundColor: getThemeColors(theme).background,
    lastValueVisible: true,
    priceLineVisible: true,
    priceFormat: {
      type: 'price',
      precision: 3,
      minMove: 0.001,
    },
  }
}