/**
 * API Constants
 * Following CLAUDE.md configuration management standards
 */

// API Base URL - defaults to same origin for production
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

// API Endpoints
export const API_ENDPOINTS = {
  // Operations
  operations: '/api/operations',
  operationStatus: (id: string) => `/api/operations/${id}`,
  
  // Data
  reports: '/api/data/reports',
  tickers: '/api/data/tickers',
  indices: '/api/data/indices',
  marketMovers: '/api/data/market-movers',
  
  // License
  licenseActivate: '/api/license/activate',
  licenseStatus: '/api/license/status',
  
  // Health
  health: '/api/health',
  
  // WebSocket
  ws: '/ws'
} as const

// Request timeouts (ms)
export const REQUEST_TIMEOUTS = {
  default: 30000,
  upload: 60000,
  download: 120000,
  longRunning: 300000
} as const

// Retry configuration
export const RETRY_CONFIG = {
  maxRetries: 3,
  retryDelay: 1000,
  retryMultiplier: 2
} as const