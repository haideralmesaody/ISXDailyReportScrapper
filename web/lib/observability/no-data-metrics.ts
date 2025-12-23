/**
 * No-Data State Observability System
 * 
 * Provides comprehensive observability for no-data state components including:
 * - Structured logging with trace context
 * - Metrics collection for user interactions
 * - Performance monitoring for state transitions
 * - Debug utilities for development
 * 
 * Follows CLAUDE.md observability standards for ISX Pulse application.
 */

import { v4 as uuidv4 } from 'uuid'

// ====================================================================
// Type Definitions
// ====================================================================

export interface TraceContext {
  trace_id: string
  session_id: string
  user_id?: string
  timestamp: string
}

export interface NoDataEvent {
  event_type: 'displayed' | 'action_clicked' | 'resolved' | 'navigation' | 'retry'
  page: string
  component: 'NoDataState' | 'DataLoadingState'
  reason?: string
  action_label?: string
  destination?: string
  duration_ms?: number
  metadata?: Record<string, unknown>
}

export interface NoDataMetric {
  metric_name: string
  value: number
  tags: Record<string, string>
  timestamp: string
}

export interface PerformanceTiming {
  operation: string
  start_time: number
  end_time?: number
  duration_ms?: number
}

export interface NoDataStateContext {
  page: string
  component_name: string
  display_reason: string
  actions_available: string[]
  instructions_count: number
}

// ====================================================================
// Configuration & State
// ====================================================================

interface ObservabilityConfig {
  enabled: boolean
  debug_mode: boolean
  console_logging: boolean
  metrics_collection: boolean
  performance_tracking: boolean
  batch_size: number
  flush_interval_ms: number
}

class ObservabilityState {
  private static instance: ObservabilityState
  private config: ObservabilityConfig
  private sessionId: string
  private eventBuffer: NoDataEvent[] = []
  private metricsBuffer: NoDataMetric[] = []
  private timingMap: Map<string, PerformanceTiming> = new Map()
  private flushTimer: NodeJS.Timeout | null = null

  private constructor() {
    this.sessionId = this.generateSessionId()
    this.config = {
      enabled: true,
      debug_mode: process.env.NODE_ENV === 'development',
      console_logging: process.env.NODE_ENV === 'development',
      metrics_collection: true,
      performance_tracking: true,
      batch_size: 50,
      flush_interval_ms: 300000, // 5 minutes (reduced from 30 seconds)
    }

    // Start flush timer if metrics collection is enabled
    if (this.config.metrics_collection) {
      this.startFlushTimer()
    }

    // Cleanup on page unload
    if (typeof window !== 'undefined') {
      window.addEventListener('beforeunload', () => {
        this.flush()
      })
    }
  }

  public static getInstance(): ObservabilityState {
    if (!ObservabilityState.instance) {
      ObservabilityState.instance = new ObservabilityState()
    }
    return ObservabilityState.instance
  }

  public getConfig(): ObservabilityConfig {
    return { ...this.config }
  }

  public getSessionId(): string {
    return this.sessionId
  }

  public updateConfig(updates: Partial<ObservabilityConfig>): void {
    this.config = { ...this.config, ...updates }
  }

  public addEvent(event: NoDataEvent): void {
    if (!this.config.enabled) return

    this.eventBuffer.push({
      ...event,
      timestamp: new Date().toISOString(),
    })

    if (this.eventBuffer.length >= this.config.batch_size) {
      this.flush()
    }
  }

  public addMetric(metric: NoDataMetric): void {
    if (!this.config.metrics_collection) return

    this.metricsBuffer.push(metric)

    if (this.metricsBuffer.length >= this.config.batch_size) {
      this.flush()
    }
  }

  public startTiming(operation: string): string {
    const timingId = `${operation}_${uuidv4()}`
    this.timingMap.set(timingId, {
      operation,
      start_time: performance.now(),
    })
    return timingId
  }

  public endTiming(timingId: string): PerformanceTiming | null {
    const timing = this.timingMap.get(timingId)
    if (!timing) return null

    timing.end_time = performance.now()
    timing.duration_ms = timing.end_time - timing.start_time

    this.timingMap.delete(timingId)
    return timing
  }

  private generateSessionId(): string {
    const timestamp = Date.now()
    const random = Math.random().toString(36).substring(2, 15)
    return `session_${timestamp}_${random}`
  }

  private startFlushTimer(): void {
    this.flushTimer = setInterval(() => {
      // Only flush if there's data to send - with null checks
      if ((this.eventBuffer?.length > 0) || (this.metricsBuffer?.length > 0) || (this.performanceBuffer?.length > 0)) {
        this.flush()
      }
    }, this.config.flush_interval_ms)
  }

  private flush(): void {
    try {
      // Flush events (would send to analytics service in production)
      if (this.eventBuffer.length > 0) {
        if (this.config.console_logging) {
          console.group('📊 No-Data Events Batch')
          this.eventBuffer.forEach(event => console.log(event))
          console.groupEnd()
        }
        
        // In production, this would send to analytics service
        // await analyticsService.sendEvents(this.eventBuffer)
        
        this.eventBuffer.length = 0
      }

      // Flush metrics (would send to metrics service in production)
      if (this.metricsBuffer.length > 0) {
        if (this.config.console_logging) {
          console.group('📈 No-Data Metrics Batch')
          this.metricsBuffer.forEach(metric => console.log(metric))
          console.groupEnd()
        }

        // In production, this would send to metrics service
        // await metricsService.sendMetrics(this.metricsBuffer)
        
        this.metricsBuffer.length = 0
      }
    } catch (error) {
      if (this.config.console_logging) {
        console.warn('Failed to flush observability data:', error)
      }
    }
  }
}

// ====================================================================
// Core Observability Functions
// ====================================================================

/**
 * Gets the current trace context for correlation
 */
export function getTraceContext(): TraceContext {
  const state = ObservabilityState.getInstance()
  return {
    trace_id: uuidv4(),
    session_id: state.getSessionId(),
    timestamp: new Date().toISOString(),
  }
}

/**
 * Logs a structured event with trace context
 */
export function logNoDataEvent(
  eventType: NoDataEvent['event_type'],
  context: Partial<NoDataEvent> = {}
): void {
  try {
    const state = ObservabilityState.getInstance()
    const traceContext = getTraceContext()

    const event: NoDataEvent = {
      event_type: eventType,
      page: context.page || 'unknown',
      component: context.component || 'NoDataState',
      reason: context.reason,
      action_label: context.action_label,
      destination: context.destination,
      duration_ms: context.duration_ms,
      metadata: {
        ...context.metadata,
        ...traceContext,
      },
    }

    // Structured logging
    if (state.getConfig().console_logging) {
      const logLevel = eventType === 'displayed' ? 'info' : 'debug'
      console[logLevel as keyof Console](`🎯 NoData Event: ${eventType}`, {
        page: event.page,
        component: event.component,
        reason: event.reason,
        trace_id: traceContext.trace_id,
        session_id: traceContext.session_id,
        timestamp: traceContext.timestamp,
        ...context.metadata,
      })
    }

    state.addEvent(event)
  } catch (error) {
    // Never let observability errors crash the application
    if (process.env.NODE_ENV === 'development') {
      console.warn('Failed to log no-data event:', error)
    }
  }
}

/**
 * Records a metric with tags for aggregation
 */
export function recordMetric(
  metricName: string,
  value: number,
  tags: Record<string, string> = {}
): void {
  try {
    const state = ObservabilityState.getInstance()
    
    const metric: NoDataMetric = {
      metric_name: metricName,
      value,
      tags: {
        session_id: state.getSessionId(),
        ...tags,
      },
      timestamp: new Date().toISOString(),
    }

    if (state.getConfig().console_logging) {
      console.debug(`📊 Metric: ${metricName}`, {
        value,
        tags,
        timestamp: metric.timestamp,
      })
    }

    state.addMetric(metric)
  } catch (error) {
    if (process.env.NODE_ENV === 'development') {
      console.warn('Failed to record metric:', error)
    }
  }
}

/**
 * Tracks performance timing for operations
 */
export function trackTiming(operation: string): {
  end: () => void
  getElapsed: () => number
} {
  const state = ObservabilityState.getInstance()
  const timingId = state.startTiming(operation)
  const startTime = performance.now()

  return {
    end: (): void => {
      try {
        const timing = state.endTiming(timingId)
        if (timing && timing.duration_ms !== undefined) {
          recordMetric('ui.performance.operation_duration', timing.duration_ms, {
            operation: timing.operation,
          })

          if (state.getConfig().console_logging) {
            console.debug(`⏱️ Timing: ${operation} completed in ${timing.duration_ms.toFixed(2)}ms`)
          }
        }
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.warn('Failed to end timing:', error)
        }
      }
    },
    getElapsed: (): number => {
      return performance.now() - startTime
    }
  }
}

// ====================================================================
// No-Data State Specific Functions
// ====================================================================

/**
 * Tracks when a no-data state is displayed to the user
 */
export function trackNoDataDisplayed(context: NoDataStateContext): void {
  logNoDataEvent('displayed', {
    page: context.page,
    component: 'NoDataState',
    reason: context.display_reason,
    metadata: {
      actions_available: context.actions_available,
      instructions_count: context.instructions_count,
      component_name: context.component_name,
    },
  })

  recordMetric('ui.no_data_state.displayed', 1, {
    page: context.page,
    reason: context.display_reason,
    has_actions: context.actions_available?.length > 0 ? 'true' : 'false',
    has_instructions: context.instructions_count > 0 ? 'true' : 'false',
  })
}

/**
 * Tracks user interactions with no-data state actions
 */
export function trackNoDataAction(
  page: string,
  actionLabel: string,
  destination?: string
): void {
  logNoDataEvent('action_clicked', {
    page,
    component: 'NoDataState',
    action_label: actionLabel,
    destination,
  })

  recordMetric('ui.no_data_state.action_clicked', 1, {
    page,
    action_label: actionLabel,
    has_destination: destination ? 'true' : 'false',
  })
}

/**
 * Tracks when a no-data state is resolved (data becomes available)
 */
export function trackNoDataResolved(
  page: string,
  resolutionTimeMs: number,
  method: 'api_success' | 'retry' | 'navigation' | 'manual'
): void {
  logNoDataEvent('resolved', {
    page,
    component: 'NoDataState',
    duration_ms: resolutionTimeMs,
    metadata: {
      resolution_method: method,
    },
  })

  recordMetric('ui.no_data_state.resolved', 1, {
    page,
    resolution_method: method,
  })

  recordMetric('ui.no_data_state.resolution_time', resolutionTimeMs, {
    page,
    resolution_method: method,
  })
}

/**
 * Tracks loading state performance
 */
export function trackLoadingState(
  page: string,
  loadingDurationMs: number,
  success: boolean
): void {
  logNoDataEvent('resolved', {
    page,
    component: 'DataLoadingState',
    duration_ms: loadingDurationMs,
    metadata: {
      loading_success: success,
    },
  })

  recordMetric('ui.loading_state.duration', loadingDurationMs, {
    page,
    success: success ? 'true' : 'false',
  })
}

/**
 * Tracks retry attempts from no-data states
 */
export function trackRetryAttempt(
  page: string,
  retryNumber: number,
  success: boolean
): void {
  logNoDataEvent('retry', {
    page,
    component: 'NoDataState',
    metadata: {
      retry_number: retryNumber,
      retry_success: success,
    },
  })

  recordMetric('ui.no_data_state.retry_attempt', 1, {
    page,
    retry_number: retryNumber.toString(),
    success: success ? 'true' : 'false',
  })
}

// ====================================================================
// Debug Utilities
// ====================================================================

/**
 * Enhanced debug logging for development
 */
export const debug = {
  /**
   * Logs detailed component state information
   */
  logComponentState: (componentName: string, state: Record<string, unknown>): void => {
    if (process.env.NODE_ENV === 'development') {
      const traceContext = getTraceContext()
      console.group(`🔍 Debug: ${componentName} State`)
      console.log('Trace Context:', traceContext)
      console.log('Component State:', state)
      console.log('Timestamp:', new Date().toISOString())
      console.groupEnd()
    }
  },

  /**
   * Logs API response information for no-data scenarios
   */
  logApiResponse: (endpoint: string, response: unknown, isEmpty: boolean): void => {
    if (process.env.NODE_ENV === 'development') {
      const traceContext = getTraceContext()
      console.group(`🌐 API Debug: ${endpoint}`)
      console.log('Trace Context:', traceContext)
      console.log('Is Empty Response:', isEmpty)
      console.log('Response Data:', response)
      console.groupEnd()
    }
  },

  /**
   * Logs performance information
   */
  logPerformance: (operation: string, durationMs: number, metadata?: Record<string, unknown>): void => {
    if (process.env.NODE_ENV === 'development') {
      const traceContext = getTraceContext()
      const isSlowOperation = durationMs > 1000 // Flag operations over 1 second
      
      console.group(`⚡ Performance: ${operation} ${isSlowOperation ? '(SLOW)' : ''}`)
      console.log('Duration:', `${durationMs.toFixed(2)}ms`)
      console.log('Trace Context:', traceContext)
      if (metadata) {
        console.log('Metadata:', metadata)
      }
      console.groupEnd()
    }
  },

  /**
   * Gets current observability state for debugging
   */
  getObservabilityState: (): Record<string, unknown> => {
    const state = ObservabilityState.getInstance()
    return {
      config: state.getConfig(),
      session_id: state.getSessionId(),
      buffer_sizes: {
        events: (state as any).eventBuffer.length,
        metrics: (state as any).metricsBuffer.length,
        timings: (state as any).timingMap.size,
      },
    }
  }
}

// ====================================================================
// Utility Functions
// ====================================================================

/**
 * Generates a correlation ID for tracking related operations
 */
export function generateCorrelationId(): string {
  return `corr_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
}

/**
 * Safely extracts page name from current location
 */
export function getCurrentPage(): string {
  if (typeof window === 'undefined') return 'server'
  
  try {
    const pathname = window.location.pathname
    const segments = pathname.split('/').filter(Boolean)
    return segments.length > 0 ? segments[0] : 'home'
  } catch (error) {
    return 'unknown'
  }
}

/**
 * Gets user agent information for debugging
 */
export function getUserAgent(): Record<string, string> {
  if (typeof window === 'undefined') {
    return { user_agent: 'server-side' }
  }

  try {
    return {
      user_agent: navigator.userAgent,
      platform: navigator.platform,
      language: navigator.language,
    }
  } catch (error) {
    return { user_agent: 'unknown' }
  }
}

// Export singleton instance for direct access if needed
export const observabilityState = ObservabilityState.getInstance()