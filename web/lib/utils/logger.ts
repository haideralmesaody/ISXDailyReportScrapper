/**
 * Production-safe logging utility
 * Automatically disables console output in production builds for tree-shaking
 */

const isDevelopment = process.env.NODE_ENV !== 'production'

/**
 * Enhanced logger with production safety and structured logging support
 */
export const logger = {
  /**
   * Standard log output
   */
  log: (...args: any[]) => {
    if (isDevelopment) {
      console.log(...args)
    }
  },

  /**
   * Error logging (always shown - even in production)
   */
  error: (...args: any[]) => {
    console.error(...args)
  },

  /**
   * Warning logging (always shown - even in production)
   */
  warn: (...args: any[]) => {
    console.warn(...args)
  },

  /**
   * Info logging (alias for log)
   */
  info: (...args: any[]) => {
    if (isDevelopment) {
      console.info(...args)
    }
  },

  /**
   * Debug logging
   */
  debug: (...args: any[]) => {
    if (isDevelopment) {
      console.debug(...args)
    }
  },

  /**
   * Group logging for better organization
   */
  group: (label: string) => {
    if (isDevelopment) {
      console.group(label)
    }
  },

  /**
   * Collapsed group logging
   */
  groupCollapsed: (label: string) => {
    if (isDevelopment) {
      console.groupCollapsed(label)
    }
  },

  /**
   * End group logging
   */
  groupEnd: () => {
    if (isDevelopment) {
      console.groupEnd()
    }
  },

  /**
   * Table logging for structured data
   */
  table: (data: any) => {
    if (isDevelopment) {
      console.table(data)
    }
  },

  /**
   * Time tracking start
   */
  time: (label: string) => {
    if (isDevelopment) {
      console.time(label)
    }
  },

  /**
   * Time tracking end
   */
  timeEnd: (label: string) => {
    if (isDevelopment) {
      console.timeEnd(label)
    }
  },

  /**
   * Assert logging
   */
  assert: (condition: boolean, ...args: any[]) => {
    if (isDevelopment) {
      console.assert(condition, ...args)
    }
  },

  /**
   * Clear console (development only)
   */
  clear: () => {
    if (isDevelopment) {
      console.clear()
    }
  },

  /**
   * Structured logging helper
   */
  structured: (level: 'info' | 'warn' | 'error', message: string, data?: Record<string, any>) => {
    if (isDevelopment) {
      const timestamp = new Date().toISOString()
      const logData = {
        timestamp,
        level,
        message,
        ...data
      }
      
      switch (level) {
        case 'error':
          console.error(message, logData)
          break
        case 'warn':
          console.warn(message, logData)
          break
        default:
          console.log(message, logData)
      }
    }
  }
}

/**
 * Create a scoped logger with a prefix
 */
export function createScopedLogger(scope: string) {
  return {
    log: (...args: any[]) => logger.log(`[${scope}]`, ...args),
    error: (...args: any[]) => logger.error(`[${scope}]`, ...args),
    warn: (...args: any[]) => logger.warn(`[${scope}]`, ...args),
    info: (...args: any[]) => logger.info(`[${scope}]`, ...args),
    debug: (...args: any[]) => logger.debug(`[${scope}]`, ...args),
  }
}

export default logger