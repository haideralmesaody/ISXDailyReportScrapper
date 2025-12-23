import { apiClient, ISXApiError } from '@/lib/api'
import '@testing-library/jest-dom'

// Mock fetch for API tests
global.fetch = jest.fn()

// Mock performance API
global.performance = {
  now: jest.fn(() => Date.now()),
} as any

describe('API Client - Operations Endpoints', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    ;(fetch as jest.Mock).mockClear()
    ;(performance.now as jest.Mock).mockReturnValue(1000)
  })

  describe('Operations Management', () => {
    describe('getOperations', () => {
      it('fetches all operations successfully', async () => {
        const mockOperations = [
          {
            id: 'op-1',
            name: 'Daily Report Generation',
            type: 'report_generation',
            status: 'idle',
            configuration: { schedule: 'daily' },
          },
          {
            id: 'op-2',
            name: 'Market Data Scraping',
            type: 'data_scraping',
            status: 'running',
            progress: 45,
          },
        ]

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(mockOperations)),
          json: jest.fn().mockResolvedValue(mockOperations),
        })

        const result = await apiClient.getOperations()

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations'),
          expect.objectContaining({
            method: 'GET',
            headers: expect.objectContaining({
              'Content-Type': 'application/json',
              'Accept': 'application/json',
            }),
          })
        )

        expect(result).toEqual(mockOperations)
      })

      it('handles empty operations list', async () => {
        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue('[]'),
          json: jest.fn().mockResolvedValue([]),
        })

        const result = await apiClient.getOperations()
        expect(result).toEqual([])
      })

      it('handles network errors', async () => {
        ;(fetch as jest.Mock).mockRejectedValue(new Error('Network error'))

        await expect(apiClient.getOperations()).rejects.toThrow(ISXApiError)
        await expect(apiClient.getOperations()).rejects.toMatchObject({
          type: '/problems/network-error',
          status: 0,
        })
      })
    })

    describe('getOperation', () => {
      it('fetches single operation details', async () => {
        const mockOperation = {
          id: 'op-1',
          name: 'Daily Report Generation',
          type: 'report_generation',
          status: 'idle',
          configuration: {
            schedule: 'daily',
            timezone: 'Asia/Baghdad',
            retryAttempts: 3,
          },
          lastRun: {
            id: 'run-123',
            status: 'completed',
            startedAt: '2025-01-30T10:00:00Z',
            completedAt: '2025-01-30T10:30:00Z',
          },
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(mockOperation)),
          json: jest.fn().mockResolvedValue(mockOperation),
        })

        const result = await apiClient.getOperation('op-1')

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/op-1'),
          expect.any(Object)
        )

        expect(result).toEqual(mockOperation)
      })

      it('handles operation not found', async () => {
        const errorResponse = {
          type: '/problems/not-found',
          title: 'Operation Not Found',
          status: 404,
          detail: 'Operation with ID op-999 does not exist',
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: false,
          status: 404,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(errorResponse)),
          json: jest.fn().mockResolvedValue(errorResponse),
        })

        await expect(apiClient.getOperation('op-999')).rejects.toThrow(ISXApiError)
        await expect(apiClient.getOperation('op-999')).rejects.toMatchObject({
          status: 404,
          type: '/problems/not-found',
        })
      })
    })

    describe('createOperation', () => {
      it('submits a run_stage request successfully', async () => {
        const mockResponse = {
          id: 'run-1',
          status: 'running',
          progress: 0,
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(mockResponse)),
          json: jest.fn().mockResolvedValue(mockResponse),
        })

        const requestPayload = {
          operation: 'run_stage',
          stage_id: 'processing',
          parameters: { retry: 1 },
        }

        const result = await apiClient.createOperation(requestPayload)

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/start'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(requestPayload),
          })
        )

        expect(result).toEqual(mockResponse)
      })

      it('submits full_pipeline request with date parameters', async () => {
        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 202,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue('{}'),
          json: jest.fn().mockResolvedValue({}),
        })

        const pipelinePayload = {
          operation: 'full_pipeline',
          parameters: { from: '2025-01-01', to: '2025-01-31' },
        }

        await apiClient.createOperation(pipelinePayload)

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/start'),
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify(pipelinePayload),
          })
        )
      })

      it('surfaces backend validation errors', async () => {
        const errorResponse = {
          type: '/errors/validation_failed',
          title: 'Validation Failed',
          status: 400,
          detail: 'stage_id parameter is required',
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(errorResponse)),
          json: jest.fn().mockResolvedValue(errorResponse),
        })

        await expect(
          apiClient.createOperation({ operation: 'run_stage', parameters: {} })
        ).rejects.toMatchObject({
          status: 400,
          type: '/errors/validation_failed',
        })
      })
    })

    describe('stopOperation', () => {
      it('stops a running operation', async () => {
        const mockResponse = {
          id: 'op-1',
          status: 'stopped',
          stoppedAt: '2025-01-30T15:30:00Z',
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(mockResponse)),
          json: jest.fn().mockResolvedValue(mockResponse),
        })

        const result = await apiClient.stopOperation('op-1')

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/op-1/stop'),
          expect.objectContaining({
            method: 'POST',
          })
        )

        expect(result).toEqual(mockResponse)
      })

      it('handles stopping non-running operation', async () => {
        const errorResponse = {
          type: '/problems/operation-not-running',
          title: 'Operation Not Running',
          status: 400,
          detail: 'Operation op-1 is not currently running',
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(errorResponse)),
          json: jest.fn().mockResolvedValue(errorResponse),
        })

        await expect(apiClient.stopOperation('op-1')).rejects.toMatchObject({
          status: 400,
          type: '/problems/operation-not-running',
        })
      })
    })

    describe('updateOperationConfig', () => {
      it('updates operation configuration', async () => {
        const newConfig = {
          schedule: 'hourly',
          retryAttempts: 5,
          timeout: 600,
        }

        const mockResponse = {
          id: 'op-1',
          configuration: newConfig,
          updatedAt: '2025-01-30T16:00:00Z',
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(mockResponse)),
          json: jest.fn().mockResolvedValue(mockResponse),
        })

        const result = await apiClient.updateOperationConfig('op-1', newConfig)

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/op-1/config'),
          expect.objectContaining({
            method: 'PUT',
            body: JSON.stringify(newConfig),
          })
        )

        expect(result).toEqual(mockResponse)
      })

      it('validates configuration before sending', async () => {
        const invalidConfig = {
          schedule: 'invalid-schedule',
          retryAttempts: -1,
        }

        const errorResponse = {
          type: '/problems/validation-error',
          title: 'Invalid Configuration',
          status: 400,
          detail: 'Configuration validation failed',
          errors: [
            { field: 'schedule', message: 'Invalid schedule format' },
            { field: 'retryAttempts', message: 'Must be non-negative' },
          ],
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: false,
          status: 400,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(errorResponse)),
          json: jest.fn().mockResolvedValue(errorResponse),
        })

        await expect(apiClient.updateOperationConfig('op-1', invalidConfig)).rejects.toMatchObject({
          status: 400,
          type: '/problems/validation-error',
        })
      })
    })

    describe('getOperationHistory', () => {
      it('fetches operation history with pagination', async () => {
        const mockHistory = {
          items: [
            {
              id: 'run-1',
              operationId: 'op-1',
              status: 'completed',
              startedAt: '2025-01-30T10:00:00Z',
              completedAt: '2025-01-30T10:30:00Z',
            },
            {
              id: 'run-2',
              operationId: 'op-1',
              status: 'failed',
              startedAt: '2025-01-29T10:00:00Z',
              completedAt: '2025-01-29T10:15:00Z',
            },
          ],
          total: 50,
          page: 1,
          pageSize: 20,
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue(JSON.stringify(mockHistory)),
          json: jest.fn().mockResolvedValue(mockHistory),
        })

        const result = await apiClient.getOperationHistory({
          operationId: 'op-1',
          page: 1,
          pageSize: 20,
        })

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/op-1/history?page=1&pageSize=20'),
          expect.any(Object)
        )

        expect(result).toEqual(mockHistory)
      })

      it('fetches history with filters', async () => {
        const filters = {
          operationId: 'op-1',
          status: 'failed',
          startDate: '2025-01-01T00:00:00Z',
          endDate: '2025-01-31T23:59:59Z',
          page: 1,
          pageSize: 20,
        }

        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue('{"items":[],"total":0}'),
          json: jest.fn().mockResolvedValue({ items: [], total: 0 }),
        })

        await apiClient.getOperationHistory(filters)

        const expectedUrl = expect.stringContaining('status=failed')
        expect(fetch).toHaveBeenCalledWith(expectedUrl, expect.any(Object))
      })
    })

    describe('exportOperationHistory', () => {
      it('exports history as CSV', async () => {
        const mockBlob = new Blob(['csv,data'], { type: 'text/csv' })
        
        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 
            'content-type': 'text/csv',
            'content-disposition': 'attachment; filename="operation-history.csv"',
          }),
          blob: jest.fn().mockResolvedValue(mockBlob),
        })

        const result = await apiClient.exportOperationHistory({
          operationId: 'op-1',
          format: 'csv',
        })

        expect(fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/operations/op-1/history/export?format=csv'),
          expect.any(Object)
        )

        expect(result).toBeInstanceOf(Blob)
      })

      it('exports history as JSON', async () => {
        const mockData = { operations: [] }
        const mockBlob = new Blob([JSON.stringify(mockData)], { type: 'application/json' })
        
        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 
            'content-type': 'application/json',
            'content-disposition': 'attachment; filename="operation-history.json"',
          }),
          blob: jest.fn().mockResolvedValue(mockBlob),
        })

        const result = await apiClient.exportOperationHistory({
          operationId: 'op-1',
          format: 'json',
        })

        expect(result).toBeInstanceOf(Blob)
      })
    })

    describe('Circuit Breaker Integration', () => {
      it('opens circuit after repeated failures', async () => {
        const error = new Error('Server error')
        
        // Simulate 5 consecutive failures
        for (let i = 0; i < 5; i++) {
          ;(fetch as jest.Mock).mockRejectedValueOnce(error)
          try {
            await apiClient.getOperations()
          } catch (e) {
            // Expected to fail
          }
        }

        // Circuit should be open now
        await expect(apiClient.getOperations()).rejects.toMatchObject({
          type: '/problems/circuit-breaker-open',
          status: 503,
        })

        // Verify fetch wasn't called after circuit opened
        expect(fetch).toHaveBeenCalledTimes(5) // Not 6
      })

      it('closes circuit after successful requests', async () => {
        // First, open the circuit
        const error = new Error('Server error')
        for (let i = 0; i < 5; i++) {
          ;(fetch as jest.Mock).mockRejectedValueOnce(error)
          try {
            await apiClient.getOperations()
          } catch (e) {
            // Expected
          }
        }

        // Reset circuit breaker to half-open after timeout
        apiClient.resetCircuitBreaker()

        // Successful requests should close the circuit
        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue('[]'),
          json: jest.fn().mockResolvedValue([]),
        })

        const result = await apiClient.getOperations()
        expect(result).toEqual([])
      })
    })

    describe('Request Debouncing', () => {
      it('debounces rapid requests to same endpoint', async () => {
        ;(fetch as jest.Mock).mockResolvedValue({
          ok: true,
          status: 200,
          headers: new Headers({ 'content-type': 'application/json' }),
          text: jest.fn().mockResolvedValue('[]'),
          json: jest.fn().mockResolvedValue([]),
        })

        // Make multiple rapid requests
        const promises = [
          apiClient.getOperations(),
          apiClient.getOperations(),
          apiClient.getOperations(),
        ]

        const results = await Promise.all(promises)

        // All should return same result
        results.forEach(result => {
          expect(result).toEqual([])
        })

        // But fetch should only be called once due to debouncing
        expect(fetch).toHaveBeenCalledTimes(1)
      })
    })

    describe('Retry Logic', () => {
      it('retries failed requests with exponential backoff', async () => {
        jest.useFakeTimers()

        // First two attempts fail, third succeeds
        ;(fetch as jest.Mock)
          .mockRejectedValueOnce(new Error('Network error'))
          .mockRejectedValueOnce(new Error('Network error'))
          .mockResolvedValueOnce({
            ok: true,
            status: 200,
            headers: new Headers({ 'content-type': 'application/json' }),
            text: jest.fn().mockResolvedValue('[]'),
            json: jest.fn().mockResolvedValue([]),
          })

        const promise = apiClient.getOperations()

        // Fast-forward through retries
        jest.advanceTimersByTime(5000)

        const result = await promise

        expect(result).toEqual([])
        expect(fetch).toHaveBeenCalledTimes(3)

        jest.useRealTimers()
      })

      it('gives up after max retries', async () => {
        jest.useFakeTimers()

        // All attempts fail
        ;(fetch as jest.Mock).mockRejectedValue(new Error('Persistent error'))

        const promise = apiClient.getOperations()

        // Fast-forward through all retries
        jest.advanceTimersByTime(60000)

        await expect(promise).rejects.toThrow('Persistent error')

        // Should try initial + max retries
        expect(fetch).toHaveBeenCalledTimes(6) // 1 initial + 5 retries

        jest.useRealTimers()
      })
    })
  })
})
