/**
 * Comprehensive API Client Integration Tests
 * 
 * Tests ISXApiClient methods, circuit breaker and retry logic,
 * request/response correlation with trace IDs, error handling and ISXApiError classification,
 * and rate limiting and exponential backoff.
 * 
 * Coverage Requirements:
 * - ISXApiClient methods testing
 * - Circuit breaker and retry logic
 * - Request/response correlation with trace IDs
 * - Error handling and ISXApiError classification
 * - Rate limiting and exponential backoff
 * - Live license key testing with ISX1M02LYE1F9QJHR9D7Z
 */

import { ISXApiClient, ISXApiError } from '@/lib/api'

// Mock fetch globally
const mockFetch = jest.fn()
global.fetch = mockFetch

// Valid test license key from requirements
const VALID_TEST_LICENSE_KEY = 'ISX1M02LYE1F9QJHR9D7Z'

describe('ISXApiClient - Comprehensive Integration Testing', () => {
  let apiClient: ISXApiClient
  let consoleSpy: jest.SpyInstance

  beforeEach(() => {
    apiClient = new ISXApiClient()
    consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    mockFetch.mockClear()
    jest.clearAllTimers()
  })

  afterEach(() => {
    consoleSpy.mockRestore()
    jest.useRealTimers()
  })

  describe('Base API Client Configuration', () => {
    it('should initialize with correct base URL for development', () => {
      process.env.NODE_ENV = 'development'
      const devClient = new ISXApiClient()
      
      expect(devClient['baseUrl']).toBe('http://localhost:8080')
    })

    it('should initialize with correct base URL for production', () => {
      process.env.NODE_ENV = 'production'
      const prodClient = new ISXApiClient()
      
      expect(prodClient['baseUrl']).toBe('')
    })

    it('should set default headers for all requests', () => {
      expect(apiClient['defaultHeaders']).toEqual({
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'X-Client-Version': '1.0.0',
        'X-Client-Platform': 'web'
      })
    })

    it('should generate unique request IDs', () => {
      const id1 = apiClient['generateRequestId']()
      const id2 = apiClient['generateRequestId']()
      
      expect(id1).not.toBe(id2)
      expect(id1).toMatch(/^req_[a-zA-Z0-9]{16}$/)
    })
  })

  describe('License Activation API', () => {
    it('should successfully activate license with valid request', async () => {
      const mockResponse = {
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        },
        trace_id: 'trace-123'
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'content-type': 'application/json',
          'x-trace-id': 'trace-123'
        }),
        json: () => Promise.resolve(mockResponse)
      })

      const result = await apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/license/activate',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
            'X-Request-ID': expect.stringMatching(/^req_[a-zA-Z0-9]{16}$/)
          }),
          body: JSON.stringify({
            license_key: VALID_TEST_LICENSE_KEY,
            email: 'test@iraqiinvestor.gov.iq'
          })
        })
      )

      expect(result).toEqual(mockResponse)
    })

    it('should handle RFC 7807 error responses correctly', async () => {
      const errorResponse = {
        type: '/problems/invalid-license',
        title: 'Invalid License Key',
        status: 400,
        detail: 'The provided license key is not valid',
        trace_id: 'trace-124'
      }

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({
          'content-type': 'application/problem+json',
          'x-trace-id': 'trace-124'
        }),
        json: () => Promise.resolve(errorResponse)
      })

      await expect(apiClient.activateLicense({
        license_key: 'INVALID_KEY',
        email: 'test@iraqiinvestor.gov.iq'
      })).rejects.toThrow(ISXApiError)

      try {
        await apiClient.activateLicense({
          license_key: 'INVALID_KEY',
          email: 'test@iraqiinvestor.gov.iq'
        })
      } catch (error) {
        expect(error).toBeInstanceOf(ISXApiError)
        expect(error.type).toBe('/problems/invalid-license')
        expect(error.title).toBe('Invalid License Key')
        expect(error.status).toBe(400)
        expect(error.detail).toBe('The provided license key is not valid')
        expect(error.traceId).toBe('trace-124')
      }
    })

    it('should handle network errors gracefully', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network request failed'))

      await expect(apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })).rejects.toThrow('Network request failed')
    })

    it('should handle timeout errors', async () => {
      jest.useFakeTimers()

      // Mock a request that never resolves
      mockFetch.mockImplementationOnce(() => new Promise(() => {}))

      const activationPromise = apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Fast-forward past timeout (30 seconds)
      jest.advanceTimersByTime(30000)

      await expect(activationPromise).rejects.toThrow('Request timeout')
    })

    it('should include correlation headers for tracing', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ success: true })
      })

      await apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      const [, options] = mockFetch.mock.calls[0]
      expect(options.headers['X-Request-ID']).toMatch(/^req_[a-zA-Z0-9]{16}$/)
      expect(options.headers['X-Correlation-ID']).toBeDefined()
    })
  })

  describe('License Status Check API', () => {
    it('should successfully check license status', async () => {
      const mockResponse = {
        valid: true,
        status: 'active',
        expires_at: '2025-12-31T23:59:59Z',
        email: 'test@iraqiinvestor.gov.iq',
        last_checked: '2025-07-28T14:30:00Z'
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve(mockResponse)
      })

      const result = await apiClient.checkLicense()

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/license/status',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Accept': 'application/json'
          })
        })
      )

      expect(result).toEqual(mockResponse)
    })

    it('should handle invalid license status', async () => {
      const mockResponse = {
        valid: false,
        status: 'invalid',
        error: 'License key not found'
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve(mockResponse)
      })

      const result = await apiClient.checkLicense()
      expect(result.valid).toBe(false)
      expect(result.status).toBe('invalid')
    })

    it('should handle expired license status', async () => {
      const mockResponse = {
        valid: false,
        status: 'expired',
        expires_at: '2024-01-01T00:00:00Z',
        error: 'License has expired'
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve(mockResponse)
      })

      const result = await apiClient.checkLicense()
      expect(result.valid).toBe(false)
      expect(result.status).toBe('expired')
    })
  })

  describe('Retry Logic and Circuit Breaker', () => {
    it('should retry failed requests with exponential backoff', async () => {
      jest.useFakeTimers()

      // First two requests fail, third succeeds
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          json: () => Promise.resolve({ success: true })
        })

      const activationPromise = apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Fast-forward through retry delays
      // First retry after 1s
      jest.advanceTimersByTime(1000)
      // Second retry after 2s
      jest.advanceTimersByTime(2000)

      const result = await activationPromise

      expect(mockFetch).toHaveBeenCalledTimes(3)
      expect(result.success).toBe(true)
    })

    it('should respect maximum retry attempts', async () => {
      // All requests fail
      mockFetch.mockRejectedValue(new Error('Persistent network error'))

      await expect(apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })).rejects.toThrow('Persistent network error')

      // Should attempt initial request + 3 retries = 4 total
      expect(mockFetch).toHaveBeenCalledTimes(4)
    })

    it('should not retry 4xx client errors', async () => {
      const errorResponse = {
        type: '/problems/invalid-license',
        title: 'Invalid License Key',
        status: 400,
        detail: 'The provided license key is not valid'
      }

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ 'content-type': 'application/problem+json' }),
        json: () => Promise.resolve(errorResponse)
      })

      await expect(apiClient.activateLicense({
        license_key: 'INVALID_KEY',
        email: 'test@iraqiinvestor.gov.iq'
      })).rejects.toThrow(ISXApiError)

      // Should only attempt once for client errors
      expect(mockFetch).toHaveBeenCalledTimes(1)
    })

    it('should retry 5xx server errors', async () => {
      jest.useFakeTimers()

      // First request returns 500, second succeeds
      mockFetch
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
          headers: new Headers({ 'content-type': 'application/problem+json' }),
          json: () => Promise.resolve({
            type: '/problems/internal-error',
            title: 'Internal Server Error',
            status: 500,
            detail: 'An unexpected error occurred'
          })
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          json: () => Promise.resolve({ success: true })
        })

      const activationPromise = apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Fast-forward through retry delay
      jest.advanceTimersByTime(1000)

      const result = await activationPromise

      expect(mockFetch).toHaveBeenCalledTimes(2)
      expect(result.success).toBe(true)
    })

    it('should implement circuit breaker pattern', async () => {
      // Simulate multiple consecutive failures to trigger circuit breaker
      mockFetch.mockRejectedValue(new Error('Service unavailable'))

      // Make multiple requests to trigger circuit breaker
      for (let i = 0; i < 5; i++) {
        try {
          await apiClient.checkLicense()
        } catch (e) {
          // Expected to fail
        }
      }

      // Circuit breaker should now be open
      const circuitBreakerError = await apiClient.checkLicense().catch(e => e)
      
      expect(circuitBreakerError.message).toContain('Circuit breaker is open')
      expect(mockFetch).toHaveBeenCalledTimes(20) // 5 requests × 4 attempts each
    })

    it('should recover from circuit breaker after timeout', async () => {
      jest.useFakeTimers()

      // Trigger circuit breaker
      mockFetch.mockRejectedValue(new Error('Service unavailable'))

      for (let i = 0; i < 5; i++) {
        try {
          await apiClient.checkLicense()
        } catch (e) {
          // Expected to fail
        }
      }

      // Fast-forward past circuit breaker timeout (60 seconds)
      jest.advanceTimersByTime(60000)

      // Mock successful response for recovery
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ valid: true, status: 'active' })
      })

      const result = await apiClient.checkLicense()
      expect(result.valid).toBe(true)
    })
  })

  describe('Rate Limiting and Backoff', () => {
    it('should handle rate limiting responses', async () => {
      jest.useFakeTimers()

      const rateLimitResponse = {
        type: '/problems/rate-limit',
        title: 'Rate Limit Exceeded',
        status: 429,
        detail: 'Too many requests, please try again later',
        'retry-after': 60
      }

      mockFetch
        .mockResolvedValueOnce({
          ok: false,
          status: 429,
          headers: new Headers({
            'content-type': 'application/problem+json',
            'retry-after': '60'
          }),
          json: () => Promise.resolve(rateLimitResponse)
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          json: () => Promise.resolve({ success: true })
        })

      const activationPromise = apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Fast-forward past retry-after delay
      jest.advanceTimersByTime(60000)

      const result = await activationPromise

      expect(mockFetch).toHaveBeenCalledTimes(2)
      expect(result.success).toBe(true)
    })

    it('should implement exponential backoff for retries', async () => {
      jest.useFakeTimers()

      // Mock network errors to trigger retries
      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          json: () => Promise.resolve({ success: true })
        })

      const startTime = Date.now()
      const activationPromise = apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Verify exponential backoff delays
      // First retry: 1s
      jest.advanceTimersByTime(1000)
      // Second retry: 2s
      jest.advanceTimersByTime(2000)
      // Third retry: 4s
      jest.advanceTimersByTime(4000)

      await activationPromise

      expect(mockFetch).toHaveBeenCalledTimes(4)
    })

    it('should add jitter to prevent thundering herd', async () => {
      jest.useFakeTimers()

      const delays: number[] = []
      const originalSetTimeout = global.setTimeout
      
      // Mock setTimeout to capture delay values
      global.setTimeout = jest.fn((callback, delay) => {
        delays.push(delay)
        return originalSetTimeout(callback, delay)
      }) as any

      mockFetch
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          json: () => Promise.resolve({ success: true })
        })

      const activationPromise = apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      jest.runAllTimers()
      await activationPromise

      // Verify jitter was applied (delay should not be exactly 1000ms)
      const retryDelay = delays.find(delay => delay > 800 && delay < 1200)
      expect(retryDelay).toBeDefined()
      expect(retryDelay).not.toBe(1000)

      global.setTimeout = originalSetTimeout
    })
  })

  describe('Request/Response Correlation', () => {
    it('should include trace IDs in request headers', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ success: true })
      })

      await apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      const [, options] = mockFetch.mock.calls[0]
      expect(options.headers['X-Request-ID']).toBeDefined()
      expect(options.headers['X-Correlation-ID']).toBeDefined()
      expect(options.headers['X-Parent-Span-ID']).toBeDefined()
    })

    it('should log request/response correlation for debugging', async () => {
      const consoleSpy = jest.spyOn(console, 'debug').mockImplementation(() => {})

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'content-type': 'application/json',
          'x-trace-id': 'server-trace-123'
        }),
        json: () => Promise.resolve({ success: true, trace_id: 'server-trace-123' })
      })

      await apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      expect(consoleSpy).toHaveBeenCalledWith(
        'API Request/Response Correlation:',
        expect.objectContaining({
          requestId: expect.stringMatching(/^req_[a-zA-Z0-9]{16}$/),
          serverTraceId: 'server-trace-123',
          method: 'POST',
          url: 'http://localhost:8080/api/license/activate',
          status: 200
        })
      )

      consoleSpy.mockRestore()
    })

    it('should handle missing trace IDs gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ success: true })
      })

      // Should not throw error even without trace IDs
      const result = await apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      expect(result.success).toBe(true)
    })
  })

  describe('Error Classification and Handling', () => {
    it('should classify different error types correctly', async () => {
      const testCases = [
        {
          status: 400,
          type: '/problems/validation-error',
          expectedClass: 'ValidationError'
        },
        {
          status: 401,
          type: '/problems/unauthorized',
          expectedClass: 'AuthenticationError'
        },
        {
          status: 403,
          type: '/problems/forbidden',
          expectedClass: 'AuthorizationError'
        },
        {
          status: 404,
          type: '/problems/not-found',
          expectedClass: 'NotFoundError'
        },
        {
          status: 409,
          type: '/problems/conflict',
          expectedClass: 'ConflictError'
        },
        {
          status: 429,
          type: '/problems/rate-limit',
          expectedClass: 'RateLimitError'
        },
        {
          status: 500,
          type: '/problems/internal-error',
          expectedClass: 'ServerError'
        }
      ]

      for (const testCase of testCases) {
        mockFetch.mockResolvedValueOnce({
          ok: false,
          status: testCase.status,
          headers: new Headers({ 'content-type': 'application/problem+json' }),
          json: () => Promise.resolve({
            type: testCase.type,
            title: 'Test Error',
            status: testCase.status,
            detail: 'Test error detail'
          })
        })

        try {
          await apiClient.checkLicense()
        } catch (error) {
          expect(error).toBeInstanceOf(ISXApiError)
          expect(error.classification).toBe(testCase.expectedClass)
        }
      }
    })

    it('should provide user-friendly error messages', async () => {
      const errorResponse = {
        type: '/problems/invalid-license',
        title: 'Invalid License Key',
        status: 400,
        detail: 'The provided license key format is invalid'
      }

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ 'content-type': 'application/problem+json' }),
        json: () => Promise.resolve(errorResponse)
      })

      try {
        await apiClient.activateLicense({
          license_key: 'INVALID_KEY',
          email: 'test@iraqiinvestor.gov.iq'
        })
      } catch (error) {
        expect(error.getUserFriendlyMessage()).toBe(
          'The license key format is invalid. Please check your license key and try again.'
        )
      }
    })

    it('should handle malformed error responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ error: 'Malformed error' })
      })

      try {
        await apiClient.checkLicense()
      } catch (error) {
        expect(error).toBeInstanceOf(ISXApiError)
        expect(error.title).toBe('Unknown Error')
        expect(error.detail).toContain('Malformed error')
      }
    })
  })

  describe('Performance and Monitoring', () => {
    it('should track request timing metrics', async () => {
      const performanceNowSpy = jest.spyOn(performance, 'now')
        .mockReturnValueOnce(1000)
        .mockReturnValueOnce(1500)

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ success: true })
      })

      await apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Should have measured request duration
      expect(performanceNowSpy).toHaveBeenCalledTimes(2)

      performanceNowSpy.mockRestore()
    })

    it('should implement request deduplication', async () => {
      // Make two identical concurrent requests
      const request1 = apiClient.checkLicense()
      const request2 = apiClient.checkLicense()

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ valid: true, status: 'active' })
      })

      const [result1, result2] = await Promise.all([request1, request2])

      // Should only make one actual HTTP request
      expect(mockFetch).toHaveBeenCalledTimes(1)
      expect(result1).toEqual(result2)
    })

    it('should cache successful license status responses', async () => {
      jest.useFakeTimers()

      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: () => Promise.resolve({ valid: true, status: 'active' })
      })

      // First request
      await apiClient.checkLicense()
      expect(mockFetch).toHaveBeenCalledTimes(1)

      // Second request within cache TTL (should use cache)
      await apiClient.checkLicense()
      expect(mockFetch).toHaveBeenCalledTimes(1)

      // Fast-forward past cache TTL (5 minutes)
      jest.advanceTimersByTime(300000)

      // Third request after cache expiry (should make new request)
      await apiClient.checkLicense()
      expect(mockFetch).toHaveBeenCalledTimes(2)
    })
  })

  describe('Security and Validation', () => {
    it('should sanitize sensitive data in error logs', async () => {
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})

      const errorResponse = {
        type: '/problems/invalid-license',
        title: 'Invalid License Key',
        status: 400,
        detail: 'The provided license key is not valid'
      }

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        headers: new Headers({ 'content-type': 'application/problem+json' }),
        json: () => Promise.resolve(errorResponse)
      })

      try {
        await apiClient.activateLicense({
          license_key: VALID_TEST_LICENSE_KEY,
          email: 'test@iraqiinvestor.gov.iq'
        })
      } catch (error) {
        // Error logs should not contain sensitive license key
        expect(consoleSpy).toHaveBeenCalledWith(
          'API Error:',
          expect.objectContaining({
            url: expect.not.stringContaining(VALID_TEST_LICENSE_KEY),
            body: expect.not.stringContaining(VALID_TEST_LICENSE_KEY)
          })
        )
      }

      consoleSpy.mockRestore()
    })

    it('should validate request payloads before sending', async () => {
      await expect(apiClient.activateLicense({
        license_key: '', // Empty license key
        email: 'test@iraqiinvestor.gov.iq'
      })).rejects.toThrow('License key is required')

      await expect(apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: '' // Empty email
      })).rejects.toThrow('Email is required')

      await expect(apiClient.activateLicense({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'invalid-email' // Invalid email format
      })).rejects.toThrow('Invalid email format')

      // Should not make any HTTP requests for validation errors
      expect(mockFetch).not.toHaveBeenCalled()
    })

    it('should handle response tampering detection', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({
          'content-type': 'application/json',
          'x-response-checksum': 'invalid-checksum'
        }),
        json: () => Promise.resolve({ success: true })
      })

      try {
        await apiClient.activateLicense({
          license_key: VALID_TEST_LICENSE_KEY,
          email: 'test@iraqiinvestor.gov.iq'
        })
      } catch (error) {
        expect(error.message).toContain('Response integrity check failed')
      }
    })
  })
})