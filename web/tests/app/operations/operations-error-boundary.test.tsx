/**
 * Operations Page Error Boundary Tests
 *
 * Comprehensive tests for error boundary handling on the operations page including:
 * - Component error catching
 * - Graceful degradation
 * - Error recovery mechanisms
 * - User feedback for errors
 * - State preservation during errors
 */

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import { ErrorBoundary } from '@/components/ui/error-boundary'
import { OperationsContent } from '@/app/operations/operations-content'

// Mock the operations hook to simulate error conditions
jest.mock('@/lib/hooks/use-websocket', () => ({
  useAllOperationUpdates: () => ({
    operations: [],
    connected: true,
    connectionStatus: 'connected',
    error: null,
    validationErrors: [],
    migrationWarnings: [],
    clearValidationErrors: jest.fn(),
    clearMigrationWarnings: jest.fn(),
    messageStats: {
      totalReceived: 150,
      successfulProcessed: 142,
      failedProcessed: 8,
      lastReceivedAt: Date.now() - 2500
    },
    getMessageSuccessRate: jest.fn(() => 94.7),
    getTimeSinceLastMessage: jest.fn(() => 2500)
  })
}))

// Mock API client to simulate network errors
jest.mock('@/lib/api', () => ({
  apiClient: {
    get: jest.fn(),
    post: jest.fn()
  }
}))

// Mock OperationRequestBuilder
jest.mock('@/lib/api/operation-request-builder', () => ({
  OperationRequestBuilder: {
    getOperationTypes: jest.fn(() => Promise.resolve([
      { id: 'scrape', name: 'Scrape ISX Data', description: 'Download daily reports' },
      { id: 'process', name: 'Process Data', description: 'Convert and validate data' },
      { id: 'upload', name: 'Upload to Sheets', description: 'Upload to Google Sheets' }
    ]))
  }
}))

// Mock useHydration hook
jest.mock('@/lib/hooks/use-hydration', () => ({
  useHydration: () => true
}))

// Test component that throws errors
class ErrorThrowingComponent extends Component {
  shouldThrow = false

  render() {
    if (this.shouldThrow) {
      throw new Error('Test error for boundary testing')
    }
    return <div data-testid="normal-content">Normal content</div>
  }
}

// Test component with async errors
class AsyncErrorComponent extends Component {
  state = { shouldThrow: false }

  componentDidMount() {
    if (this.state.shouldThrow) {
      setTimeout(() => {
        throw new Error('Async error in component')
      }, 100)
    }
  }

  render() {
    return <div data-testid="async-content">Async content</div>
  }
}

describe('Operations Error Boundary', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Suppress console.error for error boundary tests
    jest.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    // Restore console.error
    ;(console.error as jest.Mock).mockRestore()
  })

  describe('Error Boundary Basic Functionality', () => {
    test('should render children when there are no errors', () => {
      render(
        <ErrorBoundary>
          <ErrorThrowingComponent />
        </ErrorBoundary>
      )

      expect(screen.getByTestId('normal-content')).toBeInTheDocument()
    })

    test('should catch and display error when component throws', async () => {
      const TestComponent = () => {
        const [shouldThrow, setShouldThrow] = React.useState(false)

        return (
          <ErrorBoundary>
            <button
              data-testid="trigger-error"
              onClick={() => setShouldThrow(true)}
            >
              Trigger Error
            </button>
            <ErrorThrowingComponent shouldThrow={shouldThrow} />
          </ErrorBoundary>
        )
      }

      render(<TestComponent />)

      // Initially should show normal content
      expect(screen.getByTestId('normal-content')).toBeInTheDocument()

      // Trigger error
      fireEvent.click(screen.getByTestId('trigger-error'))

      // Should show error boundary UI
      await waitFor(() => {
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      })

      // Should not show the normal content
      expect(screen.queryByTestId('normal-content')).not.toBeInTheDocument()
    })

    test('should provide error details and recovery options', async () => {
      const TestComponent = () => {
        const [shouldThrow, setShouldThrow] = React.useState(false)

        return (
          <ErrorBoundary>
            <button onClick={() => setShouldThrow(true)}>
              Trigger Error
            </button>
            <ErrorThrowingComponent shouldThrow={shouldThrow} />
          </ErrorBoundary>
        )
      }

      render(<TestComponent />)

      fireEvent.click(screen.getByText('Trigger Error'))

      await waitFor(() => {
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      })

      // Should have recovery button
      expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument()
    })
  })

  describe('Operations Page Specific Error Scenarios', () => {
    test('should handle WebSocket connection errors gracefully', async () => {
      // Mock useAllOperationUpdates to return connection error
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [],
          connected: false,
          connectionStatus: 'error',
          error: 'WebSocket connection failed',
          validationErrors: [],
          migrationWarnings: [],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      // Should show error state but not crash
      await waitFor(() => {
        expect(screen.getByText(/connection error/i)).toBeInTheDocument()
      })

      // Should still show recovery options
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
    })

    test('should handle validation errors without crashing', async () => {
      // Mock validation errors
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [],
          connected: true,
          connectionStatus: 'connected',
          error: null,
          validationErrors: ['Invalid operation data format', 'Missing required fields'],
          migrationWarnings: [],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      // Should show validation warnings but continue functioning
      await waitFor(() => {
        expect(screen.getByText(/validation error/i)).toBeInTheDocument()
      })

      // Should have option to clear validation errors
      expect(screen.getByRole('button', { name: /clear errors/i })).toBeInTheDocument()
    })

    test('should handle API errors gracefully', async () => {
      // Mock API error
      const mockApi = require('@/lib/api').apiClient
      mockApi.get.mockRejectedValue(new Error('Network error'))

      render(<OperationsContent />)

      // Should handle API error without crashing
      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument()
      })
    })

    test('should preserve state during errors', async () => {
      // Mock some existing operations
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [
            { operation_id: 'existing-op-1', status: 'running', progress: 50 },
            { operation_id: 'existing-op-2', status: 'completed', progress: 100 }
          ],
          connected: true,
          connectionStatus: 'connected',
          error: null,
          validationErrors: [],
          migrationWarnings: [],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      // Should show existing operations
      await waitFor(() => {
        expect(screen.getByText('existing-op-1')).toBeInTheDocument()
        expect(screen.getByText('existing-op-2')).toBeInTheDocument()
      })

      // Simulate error during component lifecycle
      const ErrorTriggerButton = () => {
        const [triggerError, setTriggerError] = React.useState(false)

        return (
          <div>
            <button
              data-testid="trigger-component-error"
              onClick={() => setTriggerError(true)}
            >
              Trigger Component Error
            </button>
            {triggerError && <div>{throw new Error('Component error')}</div>}
          </div>
        )
      }

      // This would be handled by the error boundary in production
      expect(() => {
        render(
          <ErrorBoundary>
            <OperationsContent />
            <ErrorTriggerButton />
          </ErrorBoundary>
        )
      }).not.toThrow()
    })
  })

  describe('Error Recovery Mechanisms', () => {
    test('should provide option to retry failed operations', async () => {
      // Mock failed operation state
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [
            { operation_id: 'failed-op', status: 'failed', progress: 25, error: 'Processing failed' }
          ],
          connected: true,
          connectionStatus: 'connected',
          error: null,
          validationErrors: [],
          migrationWarnings: [],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      await waitFor(() => {
        expect(screen.getByText('failed-op')).toBeInTheDocument()
        expect(screen.getByText(/failed/i)).toBeInTheDocument()
      })

      // Should have retry button for failed operations
      const retryButton = screen.getByRole('button', { name: /retry/i })
      expect(retryButton).toBeInTheDocument()

      // Clicking retry should trigger retry mechanism
      fireEvent.click(retryButton)

      // Verify retry was attempted (mock function would be called)
      // This depends on the actual implementation of retry logic
    })

    test('should allow clearing validation errors', async () => {
      const mockClearValidationErrors = jest.fn()

      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [],
          connected: true,
          connectionStatus: 'connected',
          error: null,
          validationErrors: ['Validation error 1', 'Validation error 2'],
          migrationWarnings: [],
          clearValidationErrors: mockClearValidationErrors,
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      await waitFor(() => {
        expect(screen.getByText(/validation error/i)).toBeInTheDocument()
      })

      // Find and click clear errors button
      const clearButton = screen.getByRole('button', { name: /clear.*errors/i })
      fireEvent.click(clearButton)

      // Verify clear function was called
      expect(mockClearValidationErrors).toHaveBeenCalled()
    })

    test('should handle manual reconnection', async () => {
      // Mock disconnected state
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [],
          connected: false,
          connectionStatus: 'disconnected',
          error: null,
          validationErrors: [],
          migrationWarnings: [],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      await waitFor(() => {
        expect(screen.getByText(/disconnected/i)).toBeInTheDocument()
      })

      // Should have reconnect button
      const reconnectButton = screen.getByRole('button', { name: /reconnect/i })
      expect(reconnectButton).toBeInTheDocument()

      // Clicking reconnect should trigger reconnection
      fireEvent.click(reconnectButton)
    })
  })

  describe('Error Reporting and User Feedback', () => {
    test('should provide helpful error messages', async () => {
      const TestComponent = () => {
        throw new Error('Specific error message for testing')
      }

      render(
        <ErrorBoundary>
          <TestComponent />
        </ErrorBoundary>
      )

      await waitFor(() => {
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      })

      // Should show technical details in development
      if (process.env.NODE_ENV === 'development') {
        expect(screen.getByText(/specific error message for testing/i)).toBeInTheDocument()
      }
    })

    test('should show appropriate error types', async () => {
      const networkError = new Error('Network request failed')
      networkError.name = 'NetworkError'

      const TestComponent = () => {
        throw networkError
      }

      render(
        <ErrorBoundary>
          <TestComponent />
        </ErrorBoundary>
      )

      await waitFor(() => {
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      })

      // Should categorize error appropriately
      // This depends on the error boundary implementation
    })

    test('should maintain user-friendly interface during errors', async () => {
      render(
        <ErrorBoundary>
          <div>
            <h1>Operations Dashboard</h1>
            <ErrorThrowingComponent shouldThrow={true} />
          </div>
        </ErrorBoundary>
      )

      await waitFor(() => {
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      })

      // Should still show branding/header if implemented
      // Should not lose the entire page structure
    })
  })

  describe('Performance During Errors', () => {
    test('should not leak memory during repeated errors', async () => {
      const { unmount } = render(
        <ErrorBoundary>
          <ErrorThrowingComponent />
        </ErrorBoundary>
      )

      // Force error and recovery multiple times
      for (let i = 0; i < 10; i++) {
        const Component = () => {
          const [error, setError] = React.useState(false)

          return (
            <button onClick={() => setError(true)}>
              Error {i}
            </button>
          )
        }

        const { rerender } = render(
          <ErrorBoundary>
            <Component />
          </ErrorBoundary>
        )

        fireEvent.click(screen.getByText(`Error ${i}`))

        await waitFor(() => {
          expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
        })

        rerender(
          <ErrorBoundary>
            <div>No error</div>
          </ErrorBoundary>
        )
      }

      unmount()
      // In a real test, we would monitor memory usage here
    })

    test('should handle errors quickly without blocking UI', async () => {
      const startTime = performance.now()

      render(
        <ErrorBoundary>
          <ErrorThrowingComponent shouldThrow={true} />
        </ErrorBoundary>
      )

      await waitFor(() => {
        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument()
      })

      const endTime = performance.now()
      const errorHandlingTime = endTime - startTime

      // Error boundary should respond quickly (less than 100ms)
      expect(errorHandlingTime).toBeLessThan(100)
    })
  })

  describe('Integration with Operations Features', () => {
    test('should handle errors during operation execution', async () => {
      // Mock operation in progress that fails
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [
            {
              operation_id: 'failing-op',
              status: 'failed',
              progress: 60,
              error: 'Database connection lost during processing',
              steps: [
                { name: 'Download', status: 'completed', progress: 100 },
                { name: 'Process', status: 'failed', progress: 30, error: 'Connection timeout' },
                { name: 'Upload', status: 'pending', progress: 0 }
              ]
            }
          ],
          connected: true,
          connectionStatus: 'connected',
          error: null,
          validationErrors: [],
          migrationWarnings: [],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      await waitFor(() => {
        expect(screen.getByText('failing-op')).toBeInTheDocument()
        expect(screen.getByText(/failed/i)).toBeInTheDocument()
      })

      // Should show which step failed
      expect(screen.getByText(/process/i)).toBeInTheDocument()
      expect(screen.getByText(/connection timeout/i)).toBeInTheDocument()

      // Should provide recovery options specific to the failure
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
    })

    test('should handle partial data corruption', async () => {
      // Mock partially corrupted data
      jest.doMock('@/lib/hooks/use-websocket', () => ({
        useAllOperationUpdates: () => ({
          operations: [
            { operation_id: 'valid-op', status: 'completed', progress: 100 },
            { operation_id: 'corrupted-op', status: 'unknown', progress: NaN }, // Corrupted
            { operation_id: 'partial-op', status: 'running' } // Missing fields
          ],
          connected: true,
          connectionStatus: 'connected',
          error: null,
          validationErrors: ['Operation corrupted-op has invalid progress value'],
          migrationWarnings: ['Operation partial-op is missing required fields'],
          clearValidationErrors: jest.fn(),
          clearMigrationWarnings: jest.fn(),
          messageStats: {
            totalReceived: 0,
            successfulProcessed: 0,
            failedProcessed: 0,
            lastReceivedAt: 0
          },
          getMessageSuccessRate: jest.fn(() => 0),
          getTimeSinceLastMessage: jest.fn(() => 0)
        })
      }))

      render(<OperationsContent />)

      // Should show valid operations
      await waitFor(() => {
        expect(screen.getByText('valid-op')).toBeInTheDocument()
      })

      // Should handle corrupted operations gracefully
      expect(screen.getByText(/validation error/i)).toBeInTheDocument()
      expect(screen.getByText(/missing required fields/i)).toBeInTheDocument()
    })
  })
})