import { apiClient } from '@/lib/api'

// Mock fetch for API tests
global.fetch = jest.fn()

describe('API Client', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    ;(fetch as jest.Mock).mockClear()
  })

  describe('activateLicense', () => {
    it('sends correct request for license activation', async () => {
      const mockResponse = {
        success: true,
        message: 'License activated successfully',
        license: {
          id: 'lic-123',
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
        },
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue(mockResponse),
      })

      const request = {
        license_key: 'ISX1Y-ABCDE-12345-FGHIJ-67890',
        organization: 'Iraqi Investment Bank',
      }

      const result = await apiClient.activateLicense(request)

      expect(fetch).toHaveBeenCalledWith('/api/license/activate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      })

      expect(result).toEqual(mockResponse)
    })

    it('handles activation errors correctly', async () => {
      const errorResponse = {
        type: '/problems/invalid-license',
        title: 'Invalid License Key',
        status: 400,
        detail: 'The provided license key is not valid',
        trace_id: 'trace-123',
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 400,
        json: jest.fn().mockResolvedValue(errorResponse),
      })

      const request = {
        license_key: 'INVALID-KEY',
      }

      await expect(apiClient.activateLicense(request)).rejects.toMatchObject({
        status: 400,
        detail: 'The provided license key is not valid',
        trace_id: 'trace-123',
      })
    })

    it('handles network errors gracefully', async () => {
      ;(fetch as jest.Mock).mockRejectedValue(new Error('Network error'))

      const request = {
        license_key: 'ISX1Y-ABCDE-12345-FGHIJ-67890',
      }

      await expect(apiClient.activateLicense(request)).rejects.toThrow('Network error')
    })
  })

  describe('getLicenseStatus', () => {
    it('fetches license status successfully', async () => {
      const mockStatus = {
        license_status: 'active',
        expires_at: '2025-12-31T23:59:59Z',
        features: ['Daily Reports Access', 'Advanced Analytics'],
        user_info: {
          name: 'Test User',
          email: 'test@iraqiinvestor.gov.iq',
          organization: 'Iraqi Investment Bank',
        },
        branding: {
          platform_name: 'The Iraqi Investor',
          company_name: 'Iraqi Investment Company',
          support_email: 'support@iraqiinvestor.gov.iq',
        },
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue(mockStatus),
      })

      const result = await apiClient.getLicenseStatus()

      expect(fetch).toHaveBeenCalledWith('/api/license/status', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      expect(result).toEqual(mockStatus)
    })

    it('handles unauthorized status requests', async () => {
      const errorResponse = {
        type: '/problems/unauthorized',
        title: 'Unauthorized',
        status: 401,
        detail: 'No valid license found',
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 401,
        json: jest.fn().mockResolvedValue(errorResponse),
      })

      await expect(apiClient.getLicenseStatus()).rejects.toMatchObject({
        status: 401,
        detail: 'No valid license found',
      })
    })
  })

  describe('getDetailedStatus', () => {
    it('fetches detailed license information', async () => {
      const mockDetailedStatus = {
        license: {
          id: 'lic-123',
          key: 'ISX1Y-ABCDE-12345-FGHIJ-67890',
          status: 'active',
          created_at: '2025-01-01T00:00:00Z',
          expires_at: '2025-12-31T23:59:59Z',
          last_validated: '2025-01-26T10:30:00Z',
        },
        usage: {
          requests_today: 150,
          requests_limit: 1000,
          bandwidth_used: '2.5 MB',
          bandwidth_limit: '100 MB',
        },
        features: ['Daily Reports Access', 'Advanced Analytics', 'API Access'],
        limitations: {
          concurrent_users: 5,
          max_requests_per_day: 1000,
        },
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue(mockDetailedStatus),
      })

      const result = await apiClient.getDetailedStatus()

      expect(fetch).toHaveBeenCalledWith('/api/license/status/detailed', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      expect(result).toEqual(mockDetailedStatus)
    })
  })

  describe('transferLicense', () => {
    it('transfers license to new organization', async () => {
      const transferResponse = {
        success: true,
        message: 'License transferred successfully',
        new_organization: 'New Iraqi Bank',
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue(transferResponse),
      })

      const request = {
        new_organization: 'New Iraqi Bank',
        transfer_reason: 'Company merger',
      }

      const result = await apiClient.transferLicense(request)

      expect(fetch).toHaveBeenCalledWith('/api/license/transfer', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
      })

      expect(result).toEqual(transferResponse)
    })

    it('handles transfer validation errors', async () => {
      const errorResponse = {
        type: '/problems/transfer-forbidden',
        title: 'Transfer Not Allowed',
        status: 403,
        detail: 'License transfer is not permitted for this license type',
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 403,
        json: jest.fn().mockResolvedValue(errorResponse),
      })

      const request = {
        new_organization: 'Unauthorized Entity',
      }

      await expect(apiClient.transferLicense(request)).rejects.toMatchObject({
        status: 403,
        detail: 'License transfer is not permitted for this license type',
      })
    })
  })

  describe('getMetrics', () => {
    it('fetches license usage metrics', async () => {
      const mockMetrics = {
        daily_usage: {
          requests: 150,
          bandwidth: 2500000,
          errors: 2,
        },
        monthly_usage: {
          requests: 4500,
          bandwidth: 75000000,
          errors: 15,
        },
        performance: {
          avg_response_time: 250,
          uptime_percentage: 99.9,
        },
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue(mockMetrics),
      })

      const result = await apiClient.getMetrics()

      expect(fetch).toHaveBeenCalledWith('/api/license/metrics', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      expect(result).toEqual(mockMetrics)
    })
  })

  describe('invalidateCache', () => {
    it('invalidates license cache successfully', async () => {
      const cacheResponse = {
        success: true,
        message: 'License cache invalidated',
        cleared_entries: 3,
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue(cacheResponse),
      })

      const result = await apiClient.invalidateCache()

      expect(fetch).toHaveBeenCalledWith('/api/license/cache/invalidate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      expect(result).toEqual(cacheResponse)
    })
  })

  describe('Error handling', () => {
    it('throws ISXApiError for API errors', async () => {
      const errorResponse = {
        type: '/problems/internal',
        title: 'Internal Server Error',
        status: 500,
        detail: 'An unexpected error occurred',
        trace_id: 'trace-456',
      }

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 500,
        json: jest.fn().mockResolvedValue(errorResponse),
      })

      try {
        await apiClient.getLicenseStatus()
        fail('Expected error to be thrown')
      } catch (error: any) {
        expect(error.status).toBe(500)
        expect(error.detail).toBe('An unexpected error occurred')
        expect(error.trace_id).toBe('trace-456')
        expect(error.title).toBe('Internal Server Error')
      }
    })

    it('handles non-JSON error responses', async () => {
      ;(fetch as jest.Mock).mockResolvedValue({
        ok: false,
        status: 502,
        text: jest.fn().mockResolvedValue('Bad Gateway'),
        json: jest.fn().mockRejectedValue(new Error('Not JSON')),
      })

      try {
        await apiClient.getLicenseStatus()
        fail('Expected error to be thrown')
      } catch (error: any) {
        expect(error.status).toBe(502)
        expect(error.message).toContain('Bad Gateway')
      }
    })

    it('handles fetch failures', async () => {
      ;(fetch as jest.Mock).mockRejectedValue(new Error('Failed to fetch'))

      await expect(apiClient.getLicenseStatus()).rejects.toThrow('Failed to fetch')
    })
  })

  describe('Request headers', () => {
    it('includes authorization header when available', async () => {
      // Mock localStorage to include auth token
      const mockGetItem = jest.fn().mockReturnValue('bearer-token-123')
      Object.defineProperty(window, 'localStorage', {
        value: { getItem: mockGetItem },
        writable: true,
      })

      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue({}),
      })

      await apiClient.getLicenseStatus()

      expect(fetch).toHaveBeenCalledWith('/api/license/status', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          Authorization: 'Bearer bearer-token-123',
        },
      })
    })

    it('includes request ID for tracing', async () => {
      ;(fetch as jest.Mock).mockResolvedValue({
        ok: true,
        json: jest.fn().mockResolvedValue({}),
      })

      await apiClient.getLicenseStatus()

      const fetchCall = (fetch as jest.Mock).mock.calls[0]
      const headers = fetchCall[1].headers

      expect(headers['X-Request-ID']).toMatch(/^req-[a-f0-9-]+$/)
    })
  })

  describe('Base URL configuration', () => {
    it('uses correct base URL in development', () => {
      const originalEnv = process.env.NODE_ENV
      process.env.NODE_ENV = 'development'

      // Re-import to get fresh instance with new env
      jest.resetModules()
      const { apiClient: devApiClient } = require('@/lib/api')

      expect(devApiClient.baseUrl).toBe('http://localhost:8080')

      process.env.NODE_ENV = originalEnv
    })

    it('uses relative URLs in production', () => {
      const originalEnv = process.env.NODE_ENV
      process.env.NODE_ENV = 'production'

      jest.resetModules()
      const { apiClient: prodApiClient } = require('@/lib/api')

      expect(prodApiClient.baseUrl).toBe('')

      process.env.NODE_ENV = originalEnv
    })
  })

  describe('Response timeout handling', () => {
    it('handles request timeouts', async () => {
      jest.useFakeTimers()

      ;(fetch as jest.Mock).mockImplementation(
        () => new Promise(resolve => setTimeout(resolve, 10000))
      )

      const promise = apiClient.getLicenseStatus()

      // Fast-forward time to trigger timeout
      jest.advanceTimersByTime(10000)

      await expect(promise).rejects.toThrow()

      jest.useRealTimers()
    })
  })
})