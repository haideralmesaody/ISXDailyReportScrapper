/**
 * API client for ISX Daily Reports Scrapper backend
 * Type-safe client for Go backend integration
 */

import type {
  ApiError,
  JobListResponse,
  JobStatus,
  LicenseActivationRequest,
  LicenseResponse,
  LicenseApiResponse,
  Operation,
  OperationConfig,
  OperationTypeDefinition,
  CreateOperationRequest,
  CreateOperationResponse,
  Ticker,
  Report,
  MarketSummary,
} from '@/types/index'

// ============================================================================
// API Client Configuration
// ============================================================================

const API_BASE_URL = process.env.NODE_ENV === 'development' 
  ? 'http://localhost:8080' 
  : ''

const DEFAULT_HEADERS = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
}

// Circuit breaker configuration
const CIRCUIT_BREAKER_CONFIG = {
  failureThreshold: 5,          // Number of failures before opening circuit
  successThreshold: 3,          // Number of successes to close circuit
  timeout: 60000,              // Circuit timeout in ms (1 minute)
  retryAfter: 30000,           // Time to wait before retrying in half-open state
}

// Exponential backoff configuration
const BACKOFF_CONFIG = {
  baseDelay: 1000,             // Base delay of 1 second
  maxDelay: 30000,             // Maximum delay of 30 seconds
  multiplier: 2,               // Exponential multiplier
  jitter: true,                // Add randomization to prevent thundering herd
  maxRetries: 5,               // Maximum number of retries
}

// Request debouncing configuration
const DEBOUNCE_CONFIG = {
  delay: 500,                  // Debounce delay in ms
}

// ============================================================================
// Error Handling
// ============================================================================

export class ISXApiError extends Error {
  public readonly type: string
  public readonly status: number
  public readonly detail?: string
  public readonly traceId?: string

  constructor(error: ApiError) {
    super(error.title)
    this.name = 'ISXApiError'
    this.type = error.type
    this.status = error.status
    if (error.detail) this.detail = error.detail
    if (error.trace_id) this.traceId = error.trace_id
  }

  public isNotFound(): boolean {
    return this.status === 404
  }

  public isUnauthorized(): boolean {
    return this.status === 401
  }

  public isForbidden(): boolean {
    return this.status === 403
  }

  public isServerError(): boolean {
    return this.status >= 500
  }

  public isRateLimited(): boolean {
    return this.status === 429
  }

  public isRetriable(): boolean {
    return this.isRateLimited() || this.isServerError() || this.status === 0
  }
}

// ============================================================================
// Circuit Breaker
// ============================================================================

type CircuitState = 'CLOSED' | 'OPEN' | 'HALF_OPEN'

class CircuitBreaker {
  private state: CircuitState = 'CLOSED'
  private failureCount = 0
  private successCount = 0
  private lastFailureTime = 0

  public async execute<T>(operation: () => Promise<T>): Promise<T> {
    if (this.state === 'OPEN') {
      if (Date.now() - this.lastFailureTime < CIRCUIT_BREAKER_CONFIG.timeout) {
        throw new ISXApiError({
          type: '/problems/circuit-breaker-open',
          title: 'Service Temporarily Unavailable',
          status: 503,
          detail: 'Circuit breaker is open due to repeated failures. Please try again later.',
        })
      } else {
        this.state = 'HALF_OPEN'
        this.successCount = 0
      }
    }

    try {
      const result = await operation()
      this.onSuccess()
      return result
    } catch (error) {
      this.onFailure()
      throw error
    }
  }

  private onSuccess(): void {
    this.failureCount = 0
    
    if (this.state === 'HALF_OPEN') {
      this.successCount++
      if (this.successCount >= CIRCUIT_BREAKER_CONFIG.successThreshold) {
        this.state = 'CLOSED'
      }
    }
  }

  private onFailure(): void {
    this.failureCount++
    this.lastFailureTime = Date.now()
    
    if (this.failureCount >= CIRCUIT_BREAKER_CONFIG.failureThreshold) {
      this.state = 'OPEN'
    }
  }

  public getState(): CircuitState {
    return this.state
  }

  public reset(): void {
    this.state = 'CLOSED'
    this.failureCount = 0
    this.successCount = 0
    this.lastFailureTime = 0
  }
}

// ============================================================================
// Request Debouncer
// ============================================================================

class RequestDebouncer {
  private timers = new Map<string, NodeJS.Timeout>()
  private pendingRequests = new Map<string, Promise<any>>()

  public debounce<T>(key: string, operation: () => Promise<T>): Promise<T> {
    // If there's already a pending request for this key, return it
    const existing = this.pendingRequests.get(key)
    if (existing) {
      return existing
    }

    // Clear any existing timer for this key
    const existingTimer = this.timers.get(key)
    if (existingTimer) {
      clearTimeout(existingTimer)
    }

    // Create a new promise that will execute after the debounce delay
    const promise = new Promise<T>((resolve, reject) => {
      const timer = setTimeout(async () => {
        this.timers.delete(key)
        this.pendingRequests.delete(key)
        
        try {
          const result = await operation()
          resolve(result)
        } catch (error) {
          reject(error)
        }
      }, DEBOUNCE_CONFIG.delay)

      this.timers.set(key, timer)
    })

    this.pendingRequests.set(key, promise)
    return promise
  }

  public cancel(key: string): void {
    const timer = this.timers.get(key)
    if (timer) {
      clearTimeout(timer)
      this.timers.delete(key)
    }
    this.pendingRequests.delete(key)
  }

  public clear(): void {
    this.timers.forEach(timer => clearTimeout(timer))
    this.timers.clear()
    this.pendingRequests.clear()
  }
}

// ============================================================================
// Exponential Backoff Utility
// ============================================================================

async function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

function calculateBackoffDelay(attempt: number): number {
  const baseDelay = BACKOFF_CONFIG.baseDelay
  const multiplier = BACKOFF_CONFIG.multiplier
  const maxDelay = BACKOFF_CONFIG.maxDelay

  let delay = baseDelay * Math.pow(multiplier, attempt)
  
  if (BACKOFF_CONFIG.jitter) {
    // Add random jitter to prevent thundering herd
    delay = delay * (0.5 + Math.random() * 0.5)
  }
  
  return Math.min(delay, maxDelay)
}

async function withExponentialBackoff<T>(
  operation: () => Promise<T>,
  shouldRetry: (error: ISXApiError) => boolean = (error) => error.isRetriable()
): Promise<T> {
  let lastError: ISXApiError

  for (let attempt = 0; attempt <= BACKOFF_CONFIG.maxRetries; attempt++) {
    try {
      return await operation()
    } catch (error) {
      if (!(error instanceof ISXApiError) || !shouldRetry(error)) {
        throw error
      }

      lastError = error
      
      // Don't wait on the last attempt
      if (attempt === BACKOFF_CONFIG.maxRetries) {
        break
      }

      const delay = calculateBackoffDelay(attempt)
      console.warn(`Request failed (attempt ${attempt + 1}/${BACKOFF_CONFIG.maxRetries + 1}), retrying in ${delay}ms:`, {
        error: error.message,
        status: error.status,
        type: error.type,
        attempt: attempt + 1,
        delay
      })
      
      await sleep(delay)
    }
  }

  throw lastError!
}

// ============================================================================
// HTTP Client
// ============================================================================

class HttpClient {
  private baseUrl: string
  private circuitBreaker: CircuitBreaker
  private debouncer: RequestDebouncer

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
    this.circuitBreaker = new CircuitBreaker()
    this.debouncer = new RequestDebouncer()
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
    enableCircuitBreaker = true,
    enableDebouncing = false
  ): Promise<T> {
    const requestKey = `${options.method || 'GET'}:${endpoint}`
    
    const executeRequest = async (): Promise<T> => {
      const operation = () => this.performRequest<T>(endpoint, options)
      
      if (enableCircuitBreaker) {
        return this.circuitBreaker.execute(() => withExponentialBackoff(operation))
      } else {
        return withExponentialBackoff(operation)
      }
    }

    if (enableDebouncing) {
      return this.debouncer.debounce(requestKey, executeRequest)
    } else {
      return executeRequest()
    }
  }

  private async performRequest<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`
    const requestId = `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    const startTime = performance.now()
    
    const config: RequestInit = {
      ...options,
      headers: {
        ...DEFAULT_HEADERS,
        'X-Request-ID': requestId,
        ...options.headers,
      },
    }

    // Verbose logging disabled for production
    // console.group(`🌐 API Request [${requestId}]`)
    // console.log('📤 Request Details:', requestDetails)

    try {
      const response = await fetch(url, config)
      const responseTime = performance.now() - startTime
      
      // Log response details before processing
      const responseDetails = {
        request_id: requestId,
        status: response.status,
        status_text: response.statusText,
        ok: response.ok,
        response_time_ms: responseTime,
        headers: Object.fromEntries(response.headers.entries()),
        size_bytes: response.headers.get('content-length') || 'unknown',
        content_type: response.headers.get('content-type'),
        cache_control: response.headers.get('cache-control'),
        etag: response.headers.get('etag'),
        server: response.headers.get('server'),
        date: response.headers.get('date'),
        trace_id: response.headers.get('x-trace-id') || response.headers.get('trace-id'),
        timestamp: new Date().toISOString(),
      }

      // console.log('📥 Response Received:', responseDetails)
      
      if (!response.ok) {
        let errorData
        let parseError = null
        
        try {
          const errorText = await response.text()
          console.log('📄 Error Response Body:', {
            request_id: requestId,
            body_text: errorText,
            body_length: errorText.length,
          })
          
          errorData = errorText ? JSON.parse(errorText) : {
            type: '/problems/unknown-error',
            title: 'Unknown Error',
            status: response.status,
            detail: response.statusText,
          }
        } catch (parseErr) {
          parseError = parseErr
          errorData = {
            type: '/problems/parse-error',
            title: 'Response Parse Error',
            status: response.status,
            detail: `Failed to parse error response: ${parseErr}`,
          }
        }
        
        console.error('❌ API Error Response:', {
          request_id: requestId,
          error_data: errorData,
          parse_error: parseError instanceof Error ? parseError.message : String(parseError),
          response_details: responseDetails,
        })
        
        // console.groupEnd()
        throw new ISXApiError(errorData)
      }

      // Handle empty responses (204 No Content)
      if (response.status === 204) {
        console.log('✅ Empty Response (204):', {
          request_id: requestId,
          response_time_ms: responseTime,
        })
        // console.groupEnd()
        return {} as T
      }

      // Parse successful response
      let responseData: T
      
      try {
        const responseText = await response.text()
        // console.log('📄 Response Body Details:', {
        //   request_id: requestId,
        //   body_length: responseText.length,
        //   content_preview: responseText.substring(0, 200) + (responseText.length > 200 ? '...' : ''),
        // })
        
        responseData = responseText ? JSON.parse(responseText) : {} as T
      } catch (parseErr) {
        console.error('❌ Response Parse Error:', {
          request_id: requestId,
          parse_error: parseErr,
          response_details: responseDetails,
        })
        
        // console.groupEnd()
        throw new ISXApiError({
          type: '/problems/parse-error',
          title: 'Response Parse Error',
          status: response.status,
          detail: `Failed to parse successful response: ${parseErr}`,
        })
      }

      // Log successful completion - disabled for production
      // console.log('✅ Request Completed Successfully:', {
      //   request_id: requestId,
      //   total_time_ms: responseTime,
      //   response_size: JSON.stringify(responseData).length,
      //   data_structure: this.analyzeDataStructure(responseData),
      //   performance_metrics: {
      //     dns_lookup: (performance as any).timing?.domainLookupEnd - (performance as any).timing?.domainLookupStart,
      //     tcp_connect: (performance as any).timing?.connectEnd - (performance as any).timing?.connectStart,
      //     ssl_handshake: (performance as any).timing?.secureConnectionStart ? 
      //       (performance as any).timing?.connectEnd - (performance as any).timing?.secureConnectionStart : 0,
      //     memory_used: (performance as any).memory?.usedJSHeapSize,
      //   },
      // })
      
      // console.groupEnd()
      return responseData

    } catch (error) {
      const errorTime = performance.now() - startTime
      
      if (error instanceof ISXApiError) {
        console.error('🚨 ISX API Error:', {
          request_id: requestId,
          error_time_ms: errorTime,
          error_type: error.type,
          error_status: error.status,
          error_detail: error.detail,
          error_trace_id: error.traceId,
        })
        // console.groupEnd()
        throw error
      }
      
      // Network or other errors with enhanced logging
      const networkError = {
        request_id: requestId,
        error_time_ms: errorTime,
        error_name: error instanceof Error ? error.name : 'Unknown',
        error_message: error instanceof Error ? error.message : 'Unknown network error',
        error_stack: error instanceof Error ? error.stack : null,
        network_status: navigator.onLine ? 'online' : 'offline',
        connection_type: (navigator as any).connection?.effectiveType || 'unknown',
        url: url,
        method: config.method || 'GET',
        user_agent: navigator.userAgent,
        timestamp: new Date().toISOString(),
      }
      
      console.error('🔥 Network Error:', networkError)
      console.groupEnd()
      
      throw new ISXApiError({
        type: '/problems/network-error',
        title: 'Network Error',
        status: 0,
        detail: error instanceof Error ? error.message : 'Unknown network error',
        trace_id: requestId,
      })
    }
  }

  public getCircuitBreakerState(): CircuitState {
    return this.circuitBreaker.getState()
  }

  public resetCircuitBreaker(): void {
    this.circuitBreaker.reset()
  }

  public clearDebouncer(): void {
    this.debouncer.clear()
  }

  public async get<T>(endpoint: string, enableDebouncing = false): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' }, true, enableDebouncing)
  }

  public async post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : null,
    })
  }

  public async put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : null,
    })
  }

  public async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' })
  }
}

// ============================================================================
// API Client
// ============================================================================

export class ISXApiClient {
  private client: HttpClient

  constructor(baseUrl = API_BASE_URL) {
    this.client = new HttpClient(baseUrl)
  }

  // =========================================================================
  // Utility Methods for Circuit Breaker Management
  // =========================================================================

  public getCircuitBreakerState(): CircuitState {
    return this.client.getCircuitBreakerState()
  }

  public resetCircuitBreaker(): void {
    this.client.resetCircuitBreaker()
  }

  public clearCache(): void {
    this.client.clearDebouncer()
  }

  // =========================================================================
  // Health & System Endpoints
  // =========================================================================

  public async getHealth(): Promise<{ status: string; timestamp: string }> {
    return this.client.get('/healthz')
  }

  public async getReadiness(): Promise<{ status: string; checks: Record<string, boolean> }> {
    return this.client.get('/readyz')
  }

  public async getVersion(): Promise<any> {
    return this.client.get('/api/version')
  }

  // =========================================================================
  // License Endpoints
  // =========================================================================

  public async activateLicense(request: LicenseActivationRequest): Promise<LicenseResponse> {
    return this.client.post('/api/license/activate', request)
  }

  public async activateScratchCard(request: {
    scratch_code: string
    device_fingerprint: any
    activation_id: string
  }): Promise<any> {
    return this.client.post('/api/license/activate-scratch-card', request)
  }

  public async getLicenseStatus(): Promise<LicenseApiResponse> {
    // Enable debouncing for license status to prevent infinite loops
    return this.client.get('/api/license/status', true)
  }

  public async refreshLicenseStatus(): Promise<LicenseApiResponse> {
    // Force refresh of license status by clearing debouncer cache
    this.client.clearDebouncer()
    return this.client.get('/api/license/status', false) // No debouncing for refresh
  }

  public async checkExistingLicense(): Promise<{
    has_license: boolean
    days_remaining: number
    expiry_date: string
    license_key: string
    status: string
    is_expired: boolean
  }> {
    return this.client.get('/api/license/check-existing')
  }

  public async getLicenseDetails(): Promise<any> {
    return this.client.get('/api/license/details')
  }

  public async getLicenseHistory(): Promise<any[]> {
    return this.client.get('/api/license/history')
  }

  public async backupLicense(): Promise<{ success: boolean; backup_path: string }> {
    return this.client.post('/api/license/backup', {})
  }

  public async deactivateLicense(): Promise<void> {
    return this.client.delete('/api/license/deactivate')
  }

  // =========================================================================
  // Operations Endpoints
  // =========================================================================

  public async getOperationTypes(): Promise<OperationTypeDefinition[]> {
    return this.client.get('/api/operations/types')
  }

  public async getOperations(): Promise<Operation[]> {
    return this.client.get('/api/operations')
  }

  public async getOperation(id: string): Promise<Operation> {
    return this.client.get(`/api/operations/${id}`)
  }

  /**
   * Creates a new operation (matches backend POST /operations/start)
   */
  public async createOperation(request: CreateOperationRequest): Promise<CreateOperationResponse> {
    return this.client.post('/api/operations/start', request)
  }

  /**
   * Executes a stage directly without job queue (matches backend POST /api/stages/{stageId}/execute)
   */
  public async executeStageDirectly(stageId: string, params?: any): Promise<any> {
    const payload = {
      stage_id: stageId,
      step: stageId,
      ...(params || {})
    }
    return this.client.post(`/api/stages/${stageId}/execute`, payload)
  }

  /**
   * Gets available stages (matches backend GET /api/stages)
   */
  public async getAvailableStages(): Promise<any> {
    return this.client.get('/api/stages')
  }


  public async getJobs(): Promise<JobListResponse> {
    return this.client.get('/api/operations/jobs')
  }

  public async stopOperation(id: string): Promise<Operation> {
    return this.client.post(`/api/operations/${id}/stop`)
  }

  /**
   * Pause an operation (matches backend POST /operations/{id}/pause)
   */
  public async pauseOperation(id: string): Promise<Operation> {
    return this.client.post(`/api/operations/${id}/pause`)
  }

  /**
   * Cancel an operation (matches backend POST /operations/{id}/cancel)
   */
  public async cancelOperation(id: string): Promise<Operation> {
    return this.client.post(`/api/operations/${id}/cancel`)
  }

  /**
   * Restart an operation (matches backend POST /operations/{id}/restart)
   */
  public async restartOperation(id: string): Promise<Operation> {
    return this.client.post(`/api/operations/${id}/restart`)
  }

  public async getOperationConfig(id: string): Promise<OperationConfig> {
    return this.client.get(`/api/operations/${id}/config`)
  }

  public async updateOperationConfig(id: string, config: Partial<OperationConfig>): Promise<OperationConfig> {
    return this.client.put(`/api/operations/${id}/config`, config)
  }

  /**
   * Get operation status (matches backend GET /operations/{id}/status)
   */
  public async getOperationStatus(id: string): Promise<Operation> {
    return this.client.get(`/api/operations/${id}/status`)
  }

  /**
   * Delete an operation (matches backend DELETE /operations/{id})
   */
  public async deleteOperation(id: string): Promise<void> {
    return this.client.delete(`/api/operations/${id}`)
  }

  // =========================================================================
  // Job Endpoints (Async Operations)
  // =========================================================================

  public async getJobStatus(jobId: string): Promise<JobStatus> {
    return this.client.get(`/api/operations/jobs/${jobId}`)
  }

  public async listJobs(filters?: {
    status?: string
    operationId?: string
    stageId?: string
    limit?: number
  }): Promise<JobListResponse> {
    const params = new URLSearchParams()
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, String(value))
        }
      })
    }
    const queryString = params.toString()
    return this.client.get(`/api/operations/jobs${queryString ? `?${queryString}` : ''}`)
  }

  // =========================================================================
  // Market Data Endpoints
  // =========================================================================

  public async getTickers(): Promise<Ticker[]> {
    return this.client.get('/api/data/tickers')
  }

  public async getTicker(symbol: string): Promise<Ticker> {
    return this.client.get(`/api/data/tickers/${symbol}`)
  }

  public async getReports(ticker?: string, type?: string): Promise<Report[]> {
    const params = new URLSearchParams()
    if (ticker) params.append('ticker', ticker)
    if (type) params.append('type', type)
    
    const query = params.toString()
    return this.client.get(`/api/data/reports${query ? `?${query}` : ''}`)
  }

  public async getReport(id: string): Promise<Report> {
    return this.client.get(`/api/data/reports/${id}`)
  }

  public async getMarketSummary(date?: string): Promise<MarketSummary> {
    const query = date ? `?date=${date}` : ''
    return this.client.get(`/api/data/market-movers${query}`)
  }

  /**
   * Get market movers data (matches backend GET /api/data/market-movers)
   */
  public async getMarketMovers(date?: string): Promise<MarketSummary> {
    const query = date ? `?date=${date}` : ''
    return this.client.get(`/api/data/market-movers${query}`)
  }

  /**
   * Get indices data (matches backend GET /api/data/indices)
   */
  public async getIndices(date?: string): Promise<Array<{
    name: string
    value: number
    change: number
    change_percent: number
    date: string
  }>> {
    const query = date ? `?date=${date}` : ''
    return this.client.get(`/api/data/indices${query}`)
  }

  /**
   * Get ticker chart data (matches backend GET /api/data/ticker/{ticker}/chart)
   */
  public async getTickerChart(ticker: string, period?: string): Promise<{
    ticker: string
    period: string
    data: Array<{
      date: string
      open: number
      high: number
      low: number
      close: number
      volume: number
    }>
  }> {
    const query = period ? `?period=${period}` : ''
    return this.client.get(`/api/data/ticker/${ticker}/chart${query}`)
  }

  /**
   * Get trading dates for ISX (matches backend GET /api/data/trading-dates)
   */
  public async getTradingDates(): Promise<{
    dates: string[]
    total: number
  }> {
    return this.client.get('/api/data/trading-dates')
  }

  /**
   * Download file (matches backend GET /api/data/download)
   */
  public async downloadFile(params: {
    type: 'excel' | 'csv' | 'pdf'
    date?: string
    ticker?: string
  }): Promise<Blob> {
    const searchParams = new URLSearchParams()
    searchParams.append('type', params.type)
    if (params.date) searchParams.append('date', params.date)
    if (params.ticker) searchParams.append('ticker', params.ticker)
    
    const response = await fetch(`${API_BASE_URL}/api/data/download?${searchParams}`, {
      method: 'GET',
      headers: DEFAULT_HEADERS,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ISXApiError(errorData || {
        type: '/problems/download-error',
        title: 'Download Failed',
        status: response.status,
      })
    }

    return response.blob()
  }


  public async getOperationHistory(params: {
    operationId: string
    page?: number
    pageSize?: number
    status?: string
    startDate?: string
    endDate?: string
    triggeredBy?: string
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
    search?: string
  }): Promise<{
    items: any[]
    total: number
    page: number
    pageSize: number
    totalPages?: number
  }> {
    const { operationId, ...queryParams } = params
    const searchParams = new URLSearchParams()
    
    Object.entries(queryParams).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value))
      }
    })
    
    const queryString = searchParams.toString()
    return this.client.get(`/api/operations/${operationId}/history${queryString ? `?${queryString}` : ''}`)
  }

  public async exportOperationHistory(params: {
    operationId: string
    format: 'csv' | 'json'
    filters?: any
  }): Promise<Blob> {
    const { operationId, format, filters = {} } = params
    const searchParams = new URLSearchParams({ format, ...filters })
    
    const response = await fetch(`${API_BASE_URL}/api/operations/${operationId}/history/export?${searchParams}`, {
      method: 'GET',
      headers: DEFAULT_HEADERS,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ISXApiError(errorData || {
        type: '/problems/export-error',
        title: 'Export Failed',
        status: response.status,
      })
    }

    return response.blob()
  }

  // =========================================================================
  // File Management Endpoints
  // =========================================================================

  public async uploadFile(file: File, type: 'data' | 'config'): Promise<{ filename: string; size: number }> {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('type', type)

    const response = await fetch(`${API_BASE_URL}/api/data/files/upload`, {
      method: 'POST',
      body: formData,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => null)
      throw new ISXApiError(errorData || {
        type: '/problems/upload-error',
        title: 'Upload Failed',
        status: response.status,
      })
    }

    return response.json()
  }

  public async getFiles(type?: string): Promise<Array<{ name: string; size: number; modified: string }>> {
    const query = type ? `?type=${type}` : ''
    return this.client.get(`/api/data/files${query}`)
  }

  public async deleteFile(filename: string): Promise<void> {
    return this.client.delete(`/api/data/files/${filename}`)
  }

  // =========================================================================
  // Generic HTTP Methods (Escape Hatch)
  // =========================================================================
  
  /**
   * Generic GET request - use domain-specific methods when available
   */
  public async get<T>(endpoint: string, enableDebouncing = false): Promise<T> {
    return this.client.get<T>(endpoint, enableDebouncing)
  }

  /**
   * Generic POST request - use domain-specific methods when available
   */
  public async post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.client.post<T>(endpoint, data)
  }

  /**
   * Generic PUT request - use domain-specific methods when available
   */
  public async put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.client.put<T>(endpoint, data)
  }

  /**
   * Generic DELETE request - use domain-specific methods when available
   */
  public async delete<T>(endpoint: string): Promise<T> {
    return this.client.delete<T>(endpoint)
  }
}

// ============================================================================
// Type Exports
// ============================================================================

export type { CircuitState }

// ============================================================================
// Default Export
// ============================================================================

export const apiClient = new ISXApiClient()
export default apiClient
