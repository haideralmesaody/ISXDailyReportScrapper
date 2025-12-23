/**
 * Debug instrumentation utility for ISX Daily Reports Scrapper
 * Development-focused observability and debugging tools
 */

'use client'

// ============================================================================
// Debug Configuration
// ============================================================================

interface DebugConfig {
  enabled: boolean
  logLevel: 'debug' | 'info' | 'warn' | 'error'
  enablePerformanceTracking: boolean
  enableNetworkLogging: boolean
  enableStateLogging: boolean
  enableErrorBoundaryLogging: boolean
  maxLogEntries: number
}

const DEFAULT_DEBUG_CONFIG: DebugConfig = {
  enabled: process.env.NODE_ENV === 'development',
  logLevel: 'debug',
  enablePerformanceTracking: true,
  enableNetworkLogging: true,
  enableStateLogging: true,
  enableErrorBoundaryLogging: true,
  maxLogEntries: 1000,
}

// ============================================================================
// Debug Logger Class
// ============================================================================

class DebugLogger {
  private config: DebugConfig
  private logHistory: DebugLogEntry[] = []
  private performanceMarks: Map<string, number> = new Map()
  private componentStates: Map<string, any> = new Map()

  constructor(config: Partial<DebugConfig> = {}) {
    this.config = { ...DEFAULT_DEBUG_CONFIG, ...config }
    
    if (this.config.enabled) {
      this.initializeGlobalDebugTools()
      console.log('🔧 ISX Debug Logger Initialized:', {
        config: this.config,
        timestamp: new Date().toISOString(),
      })
    }
  }

  // Initialize global debug tools available in browser console
  private initializeGlobalDebugTools(): void {
    if (typeof window !== 'undefined') {
      (window as any).ISXDebug = {
        getLogHistory: () => this.logHistory,
        getPerformanceMarks: () => Array.from(this.performanceMarks.entries()),
        getComponentStates: () => Array.from(this.componentStates.entries()),
        clearLogs: () => this.clearLogs(),
        exportLogs: () => this.exportLogs(),
        setLogLevel: (level: DebugConfig['logLevel']) => { this.config.logLevel = level },
        togglePerformanceTracking: () => { 
          this.config.enablePerformanceTracking = !this.config.enablePerformanceTracking 
        },
        getLicenseFlowTrace: () => this.getLicenseFlowTrace(),
        getNetworkTrace: () => this.getNetworkTrace(),
        simulateLicenseError: (errorType: string) => this.simulateLicenseError(errorType),
      }
      
      console.log('🛠️ Global debug tools available at window.ISXDebug')
    }
  }

  // Core logging method
  public log(
    level: DebugConfig['logLevel'],
    category: string,
    message: string,
    data?: any,
    componentName?: string
  ): void {
    if (!this.config.enabled) return

    const entry: DebugLogEntry = {
      id: `log_${Date.now()}_${Math.random().toString(36).substr(2, 6)}`,
      timestamp: new Date().toISOString(),
      level,
      category,
      message,
      ...(data !== undefined && { data }),
      ...(componentName !== undefined && { componentName }),
      ...(new Error().stack !== undefined && { stackTrace: new Error().stack }),
      performance: {
        memory: (performance as any).memory?.usedJSHeapSize,
        timing: performance.now(),
      },
    }

    // Add to history
    this.addToHistory(entry)

    // Console output based on level
    const consoleMethod = this.getConsoleMethod(level)
    const emoji = this.getLevelEmoji(level)
    
    if (componentName) {
      consoleMethod(`${emoji} [${category}] ${componentName}: ${message}`, data || '')
    } else {
      consoleMethod(`${emoji} [${category}] ${message}`, data || '')
    }
  }

  // Convenience methods
  public debug(category: string, message: string, data?: any, componentName?: string): void {
    this.log('debug', category, message, data, componentName)
  }

  public info(category: string, message: string, data?: any, componentName?: string): void {
    this.log('info', category, message, data, componentName)
  }

  public warn(category: string, message: string, data?: any, componentName?: string): void {
    this.log('warn', category, message, data, componentName)
  }

  public error(category: string, message: string, data?: any, componentName?: string): void {
    this.log('error', category, message, data, componentName)
  }

  // Performance tracking
  public startPerformanceMark(markName: string): void {
    if (!this.config.enablePerformanceTracking) return
    
    this.performanceMarks.set(markName, performance.now())
    this.debug('performance', `Started performance mark: ${markName}`)
  }

  public endPerformanceMark(markName: string): number | null {
    if (!this.config.enablePerformanceTracking) return null
    
    const startTime = this.performanceMarks.get(markName)
    if (!startTime) {
      this.warn('performance', `Performance mark not found: ${markName}`)
      return null
    }

    const duration = performance.now() - startTime
    this.performanceMarks.delete(markName)
    
    this.info('performance', `Performance mark completed: ${markName}`, {
      duration_ms: duration,
      start_time: startTime,
      end_time: performance.now(),
    })

    return duration
  }

  // Component state tracking
  public trackComponentState(componentName: string, state: any): void {
    if (!this.config.enableStateLogging) return
    
    const stateEntry = {
      timestamp: new Date().toISOString(),
      state: JSON.parse(JSON.stringify(state)), // Deep copy
      performance: performance.now(),
    }

    this.componentStates.set(componentName, stateEntry)
    
    this.debug('component_state', `State updated for ${componentName}`, {
      component: componentName,
      state: stateEntry,
      state_size: JSON.stringify(state).length,
    })
  }

  // Network request tracking
  public trackNetworkRequest(
    requestId: string,
    method: string,
    url: string,
    startTime: number,
    endTime?: number,
    status?: number,
    error?: any
  ): void {
    if (!this.config.enableNetworkLogging) return

    const networkEntry = {
      request_id: requestId,
      method,
      url,
      start_time: startTime,
      end_time: endTime || performance.now(),
      duration_ms: endTime ? endTime - startTime : performance.now() - startTime,
      status,
      error,
      timestamp: new Date().toISOString(),
    }

    if (error) {
      this.error('network', `Network request failed: ${method} ${url}`, networkEntry)
    } else {
      this.info('network', `Network request completed: ${method} ${url}`, networkEntry)
    }
  }

  // License flow specific debugging
  public traceLicenseFlow(step: string, data?: any): void {
    this.info('license_flow', `License flow step: ${step}`, {
      step,
      data,
      flow_timestamp: new Date().toISOString(),
      performance_now: performance.now(),
    })
  }

  // Error boundary integration
  public logErrorBoundary(error: Error, errorInfo: any, componentStack?: string): void {
    if (!this.config.enableErrorBoundaryLogging) return

    this.error('error_boundary', 'React Error Boundary caught error', {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack,
      },
      errorInfo,
      componentStack,
      url: window.location.href,
      userAgent: navigator.userAgent,
      timestamp: new Date().toISOString(),
    })
  }

  // Helper methods
  private addToHistory(entry: DebugLogEntry): void {
    this.logHistory.push(entry)
    
    // Maintain max log entries
    if (this.logHistory.length > this.config.maxLogEntries) {
      this.logHistory = this.logHistory.slice(-this.config.maxLogEntries)
    }
  }

  private getConsoleMethod(level: DebugConfig['logLevel']): typeof console.log {
    switch (level) {
      case 'debug': return console.debug
      case 'info': return console.info
      case 'warn': return console.warn
      case 'error': return console.error
      default: return console.log
    }
  }

  private getLevelEmoji(level: DebugConfig['logLevel']): string {
    const emojis = {
      debug: '🔍',
      info: 'ℹ️',
      warn: '⚠️',
      error: '❌',
    }
    return emojis[level] || '📝'
  }

  // Export functionality
  private exportLogs(): string {
    const exportData = {
      timestamp: new Date().toISOString(),
      config: this.config,
      logHistory: this.logHistory,
      performanceMarks: Array.from(this.performanceMarks.entries()),
      componentStates: Array.from(this.componentStates.entries()),
      browserInfo: {
        userAgent: navigator.userAgent,
        url: window.location.href,
        memory: (performance as any).memory,
      },
    }

    return JSON.stringify(exportData, null, 2)
  }

  private clearLogs(): void {
    this.logHistory = []
    this.performanceMarks.clear()
    this.componentStates.clear()
    console.clear()
    this.info('debug', 'Debug logs cleared')
  }

  // Specialized trace methods
  private getLicenseFlowTrace(): DebugLogEntry[] {
    return this.logHistory.filter(entry => 
      entry.category === 'license_flow' || 
      entry.category.includes('license') ||
      entry.message.toLowerCase().includes('license')
    )
  }

  private getNetworkTrace(): DebugLogEntry[] {
    return this.logHistory.filter(entry => 
      entry.category === 'network' ||
      entry.category === 'api' ||
      entry.category === 'websocket'
    )
  }

  // Development utilities
  private simulateLicenseError(errorType: string): void {
    const simulatedErrors = {
      network_error: () => {
        this.error('license_simulation', 'Simulated network error', {
          error_type: 'network_error',
          simulated: true,
        })
      },
      expired_license: () => {
        this.error('license_simulation', 'Simulated expired license', {
          error_type: 'expired_license',
          simulated: true,
        })
      },
      invalid_format: () => {
        this.error('license_simulation', 'Simulated invalid license format', {
          error_type: 'invalid_format',
          simulated: true,
        })
      },
    }

    const simulator = simulatedErrors[errorType as keyof typeof simulatedErrors]
    if (simulator) {
      simulator()
    } else {
      this.warn('license_simulation', `Unknown error type: ${errorType}`, {
        available_types: Object.keys(simulatedErrors),
      })
    }
  }
}

// ============================================================================
// Debug Log Entry Interface
// ============================================================================

interface DebugLogEntry {
  id: string
  timestamp: string
  level: DebugConfig['logLevel']
  category: string
  message: string
  data?: any
  componentName?: string
  stackTrace?: string
  performance: {
    memory?: number
    timing: number
  }
}

// ============================================================================
// React Hooks for Debug Integration
// ============================================================================

import { useEffect, useRef } from 'react'

export function useDebugComponent(componentName: string, state?: any) {
  const debugLogger = useRef(getDebugLogger())

  useEffect(() => {
    debugLogger.current.debug('component_lifecycle', `${componentName} mounted`)
    
    // Copy ref to variable to avoid stale reference in cleanup
    const logger = debugLogger.current
    return () => {
      logger.debug('component_lifecycle', `${componentName} unmounted`)
    }
  }, [componentName])

  useEffect(() => {
    if (state !== undefined) {
      debugLogger.current.trackComponentState(componentName, state)
    }
  }, [componentName, state])

  return {
    debug: (message: string, data?: any) => 
      debugLogger.current.debug('component', message, data, componentName),
    info: (message: string, data?: any) => 
      debugLogger.current.info('component', message, data, componentName),
    warn: (message: string, data?: any) => 
      debugLogger.current.warn('component', message, data, componentName),
    error: (message: string, data?: any) => 
      debugLogger.current.error('component', message, data, componentName),
    traceLicenseFlow: (step: string, data?: any) => 
      debugLogger.current.traceLicenseFlow(`${componentName}: ${step}`, data),
    startPerformanceMark: (markName: string) => 
      debugLogger.current.startPerformanceMark(`${componentName}_${markName}`),
    endPerformanceMark: (markName: string) => 
      debugLogger.current.endPerformanceMark(`${componentName}_${markName}`),
  }
}

// ============================================================================
// Singleton Debug Logger Instance
// ============================================================================

let debugLoggerInstance: DebugLogger | null = null

export function getDebugLogger(config?: Partial<DebugConfig>): DebugLogger {
  if (!debugLoggerInstance) {
    debugLoggerInstance = new DebugLogger(config)
  }
  return debugLoggerInstance
}

// ============================================================================
// Default Export
// ============================================================================

export default getDebugLogger()