/**
 * Phase helper utilities for pipeline stage progression.
 * Centralizes threshold logic so that UI and data contracts stay in sync.
 */

const DEFAULT_OPERATION_TYPE = 'scraping'

// Development-only performance logging with SSR-safe lazy evaluation
const isDevelopment = process.env.NODE_ENV === 'development'

// SSR-safe check for performance logging
const isPerformanceLoggingEnabled = (): boolean => {
  return isDevelopment &&
    typeof window !== 'undefined' &&
    window.localStorage !== undefined &&
    window.localStorage.getItem('debug_performance') === 'true'
}

// Performance logging helper
const logPerformance = (operation: string, duration: number, hit: boolean, cacheSize: number): void => {
  if (!isPerformanceLoggingEnabled()) return

  const status = hit ? 'HIT' : 'MISS'
  console.log(`[PhaseHelpers Performance] ${operation}: ${duration.toFixed(2)}ms (${status}) - Cache size: ${cacheSize}`)
}

// Simple cache with TTL for phase helpers
const cache = new Map<string, { value: any; timestamp: number }>()
const CACHE_TTL = 5000 // 5 seconds

function getCachedValue(key: string, fallback: string, fn: () => any): any {
  const startTime = performance.now()
  const now = Date.now()
  const cacheKey = `${key}:${fallback}` // Include fallback to prevent mixing

  const cached = cache.get(cacheKey)
  if (cached && (now - cached.timestamp < CACHE_TTL)) {
    const duration = performance.now() - startTime
    logPerformance(key, duration, true, cache.size)
    return cached.value
  }

  const value = fn()
  cache.set(cacheKey, { value, timestamp: now })

  const duration = performance.now() - startTime
  logPerformance(key, duration, false, cache.size)

  return value
}

const PHASE_DICTIONARY: Record<string, Record<string, { label: string; description: string }>> = {
  scraping: {
    preparing: { label: 'Preparing', description: 'Initializing scrape session' },
    scanning: { label: 'Scanning Sources', description: 'Finding available reports' },
    downloading: { label: 'Downloading Files', description: 'Fetching Excel files' },
    verifying: { label: 'Verifying Data', description: 'Validating downloaded files' },
    complete: { label: 'Complete', description: 'Scraping finished' }
  },
  processing: {
    preparing: { label: 'Preparing', description: 'Setting up processing pipeline' },
    reading: { label: 'Reading Excel Files', description: 'Loading source files' },
    transforming: { label: 'Transforming', description: 'Converting to CSV and normalizing' },
    writing: { label: 'Writing Reports', description: 'Saving processed outputs' },
    complete: { label: 'Complete', description: 'Processing finished' }
  },
  indices: {
    preparing: { label: 'Preparing', description: 'Loading market data' },
    extracting: { label: 'Extracting Indices', description: 'Calculating ISX indices' },
    complete: { label: 'Complete', description: 'Index extraction finished' }
  },
  liquidity: {
    preparing: { label: 'Preparing', description: 'Initializing liquidity engine' },
    reading: { label: 'Loading Data', description: 'Reading processed CSVs' },
    calculating: { label: 'Calculating Scores', description: 'Computing liquidity metrics' },
    scaling: { label: 'Scaling & Normalizing', description: 'Applying cross-sectional scaling' },
    writing: { label: 'Writing Results', description: 'Persisting liquidity outputs' },
    complete: { label: 'Complete', description: 'Liquidity analysis finished' }
  },
  default: {
    preparing: { label: 'Preparing', description: 'Initializing stage' },
    running: { label: 'Running', description: 'Work in progress' },
    complete: { label: 'Complete', description: 'Stage finished' }
  }
}

export function getOperationType(stepId?: string, fallback: string = DEFAULT_OPERATION_TYPE): string {
  if (!stepId) {
    return fallback
  }

  return getCachedValue('operationType', fallback, () => {
    const normalized = stepId.toLowerCase()

    if (normalized.includes('scrap')) return 'scraping'
    if (normalized.includes('process')) return 'processing'
    if (normalized.includes('index')) return 'indices'
    if (normalized.includes('liquid')) return 'liquidity'

    return fallback
  })
}

export function getOperationThresholds(operationType: string): number[] {
  return getCachedValue('thresholds', operationType, () => {
    switch (operationType) {
      case 'scraping':
        return [0, 25, 50, 75, 100]
      case 'processing':
        return [0, 25, 50, 75, 100]
      case 'indices':
        return [0, 50, 100]
      case 'liquidity':
        return [0, 20, 40, 70, 90, 100]
      default:
        return [0, 50, 100]
    }
  })
}

export function getPhaseForThreshold(operationType: string, thresholdIndex: number): string {
  return getCachedValue('phaseForThreshold', `${operationType}:${thresholdIndex}`, () => {
    const phaseMap: Record<string, string[]> = {
      scraping: ['preparing', 'scanning', 'downloading', 'verifying', 'complete'],
      processing: ['preparing', 'reading', 'transforming', 'writing', 'complete'],
      indices: ['preparing', 'extracting', 'complete'],
      liquidity: ['preparing', 'reading', 'calculating', 'scaling', 'writing', 'complete'],
      default: ['preparing', 'running', 'complete']
    }

    const phases = phaseMap[operationType] || phaseMap.default
    const safeIndex = Math.min(thresholdIndex, phases.length - 1)

    return phases[safeIndex]
  })
}

const toTitleCase = (value: string): string => value.charAt(0).toUpperCase() + value.slice(1)

export function getPhaseDisplay(operationType: string, phase?: string): { label: string; description: string } {
  if (!phase) {
    return { label: 'Pending', description: 'Waiting to start' }
  }

  const dictionary = PHASE_DICTIONARY[operationType] || PHASE_DICTIONARY.default
  const normalized = phase.toLowerCase()
  if (dictionary[normalized]) {
    return dictionary[normalized]
  }

  // Fallback: title case the phase
  return { label: toTitleCase(normalized), description: 'In progress' }
}

export function getPhaseBadgeVariant(status: string, phase?: string): 'default' | 'secondary' | 'outline' | 'destructive' {
  if (status === 'failed') return 'destructive'
  if (status === 'completed') return 'secondary'
  if (phase === 'downloading' || phase === 'extracting' || phase === 'calculating') return 'default'
  if (phase === 'preparing' || phase === 'initializing') return 'outline'
  return 'default'
}
