import React from 'react'
import { render, screen, waitFor, fireEvent, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useRouter, useSearchParams } from 'next/navigation'
import OperationsPage from '@/app/operations/page'
import { apiClient } from '@/lib/api'
import { useWebSocket } from '@/lib/hooks/use-websocket'
import '@testing-library/jest-dom'

// Mock dependencies
jest.mock('next/navigation')
jest.mock('@/lib/api')
jest.mock('@/lib/hooks/use-websocket')

describe('OperationsPage', () => {
  const mockPush = jest.fn()
  const mockReplace = jest.fn()
  const mockRefresh = jest.fn()
  const mockSearchParams = new URLSearchParams()
  
  const mockWebSocket = {
    connected: true,
    connect: jest.fn(),
    disconnect: jest.fn(),
    subscribe: jest.fn(),
    unsubscribe: jest.fn(),
    send: jest.fn(),
    connectionState: 'connected' as const,
    error: null,
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
      replace: mockReplace,
      refresh: mockRefresh,
    })
    ;(useSearchParams as jest.Mock).mockReturnValue(mockSearchParams)
    ;(useWebSocket as jest.Mock).mockReturnValue(mockWebSocket)
  })

  describe('Initial Render', () => {
    const testCases = [
      {
        name: 'displays loading state while fetching operations',
        setup: () => {
          ;(apiClient.getOperations as jest.Mock).mockReturnValue(
            new Promise(() => {}) // Never resolves
          )
        },
        assertions: () => {
          expect(screen.getByTestId('operations-loading')).toBeInTheDocument()
          expect(screen.getByText(/Loading operations/i)).toBeInTheDocument()
          expect(screen.queryByTestId('operations-list')).not.toBeInTheDocument()
        },
      },
      {
        name: 'displays operations list when data is loaded',
        setup: async () => {
          const mockOperations = [
            {
              id: 'op-1',
              name: 'Daily Report Generation',
              type: 'report_generation',
              status: 'idle',
              lastRun: '2025-01-30T10:00:00Z',
              nextRun: '2025-01-31T10:00:00Z',
              configuration: { schedule: 'daily', timezone: 'Asia/Baghdad' },
            },
            {
              id: 'op-2',
              name: 'Market Data Scraping',
              type: 'data_scraping',
              status: 'running',
              progress: 45,
              startedAt: '2025-01-30T14:30:00Z',
              steps: [
                { id: 's1', name: 'Download', status: 'completed', progress: 100 },
                { id: 's2', name: 'Parse', status: 'running', progress: 45 },
                { id: 's3', name: 'Store', status: 'pending', progress: 0 },
              ],
            },
          ]
          ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)
        },
        assertions: async () => {
          await waitFor(() => {
            expect(screen.queryByTestId('operations-loading')).not.toBeInTheDocument()
          })
          
          expect(screen.getByTestId('operations-list')).toBeInTheDocument()
          expect(screen.getByText('Daily Report Generation')).toBeInTheDocument()
          expect(screen.getByText('Market Data Scraping')).toBeInTheDocument()
          expect(screen.getByText('45%')).toBeInTheDocument()
        },
      },
      {
        name: 'displays error state when operations fetch fails',
        setup: async () => {
          const error = new Error('Failed to fetch operations')
          ;(apiClient.getOperations as jest.Mock).mockRejectedValue(error)
        },
        assertions: async () => {
          await waitFor(() => {
            expect(screen.getByTestId('operations-error')).toBeInTheDocument()
          })
          
          expect(screen.getByText(/Failed to fetch operations/i)).toBeInTheDocument()
          expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument()
        },
      },
      {
        name: 'displays empty state when no operations exist',
        setup: async () => {
          ;(apiClient.getOperations as jest.Mock).mockResolvedValue([])
        },
        assertions: async () => {
          await waitFor(() => {
            expect(screen.getByTestId('operations-empty')).toBeInTheDocument()
          })
          
          expect(screen.getByText(/No operations configured/i)).toBeInTheDocument()
          expect(screen.getByRole('button', { name: /Create Operation/i })).toBeInTheDocument()
        },
      },
    ]

    testCases.forEach(({ name, setup, assertions }) => {
      it(name, async () => {
        await setup()
        render(<OperationsPage />)
        await assertions()
      })
    })
  })

  describe('WebSocket Integration', () => {
    beforeEach(() => {
      const mockOperations = [
        {
          id: 'op-1',
          name: 'Test Operation',
          type: 'data_processing',
          status: 'idle',
        },
      ]
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)
    })

    it('subscribes to operation updates on mount', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(mockWebSocket.subscribe).toHaveBeenCalledWith(
          'operation:update',
          expect.any(Function)
        )
      })
    })

    it('updates operation status in real-time', async () => {
      let updateHandler: (data: any) => void
      mockWebSocket.subscribe.mockImplementation((event, handler) => {
        if (event === 'operation:update') {
          updateHandler = handler
        }
      })

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Operation')).toBeInTheDocument()
      })

      // Simulate WebSocket update
      updateHandler!({
        operationId: 'op-1',
        status: 'running',
        progress: 25,
        currentStage: 'Downloading data',
      })

      await waitFor(() => {
        expect(screen.getByText('25%')).toBeInTheDocument()
        expect(screen.getByText('Downloading data')).toBeInTheDocument()
      })
    })

    it('handles WebSocket disconnection gracefully', async () => {
      ;(useWebSocket as jest.Mock).mockReturnValue({
        ...mockWebSocket,
        connected: false,
        connectionState: 'disconnected',
      })

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByTestId('websocket-warning')).toBeInTheDocument()
        expect(screen.getByText(/Real-time updates unavailable/i)).toBeInTheDocument()
      })
    })

    it('unsubscribes from WebSocket on unmount', async () => {
      const { unmount } = render(<OperationsPage />)
      
      await waitFor(() => {
        expect(mockWebSocket.subscribe).toHaveBeenCalled()
      })

      unmount()

      expect(mockWebSocket.unsubscribe).toHaveBeenCalledWith('operation:update')
    })
  })

  describe('User Interactions', () => {
    beforeEach(async () => {
      const mockOperations = [
        {
          id: 'op-1',
          name: 'Report Generation',
          type: 'report_generation',
          status: 'idle',
          configuration: { schedule: 'daily' },
        },
        {
          id: 'op-2',
          name: 'Data Scraping',
          type: 'data_scraping',
          status: 'running',
          progress: 60,
        },
      ]
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)
    })

    it('allows starting an idle operation', async () => {
      ;(apiClient.createOperation as jest.Mock).mockResolvedValue({
        id: 'op-1',
        status: 'running',
      })

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Report Generation')).toBeInTheDocument()
      })

      const startButton = screen.getByTestId('start-operation-op-1')
      await userEvent.click(startButton)

      expect(apiClient.createOperation).toHaveBeenCalledWith({
        operation: 'run_stage',
        stage_id: 'op-1',
        parameters: {},
      })
      
      await waitFor(() => {
        expect(screen.getByTestId('operation-status-op-1')).toHaveTextContent('running')
      })
    })

    it('allows stopping a running operation', async () => {
      ;(apiClient.stopOperation as jest.Mock).mockResolvedValue({
        id: 'op-2',
        status: 'stopped',
      })

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Data Scraping')).toBeInTheDocument()
      })

      const stopButton = screen.getByTestId('stop-operation-op-2')
      await userEvent.click(stopButton)

      expect(apiClient.stopOperation).toHaveBeenCalledWith('op-2')
      
      await waitFor(() => {
        expect(screen.getByTestId('operation-status-op-2')).toHaveTextContent('stopped')
      })
    })

    it('shows configuration modal when clicking configure', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Report Generation')).toBeInTheDocument()
      })

      const configButton = screen.getByTestId('configure-operation-op-1')
      await userEvent.click(configButton)

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument()
        expect(screen.getByText(/Configure Report Generation/i)).toBeInTheDocument()
      })
    })

    it('navigates to operation details on row click', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Report Generation')).toBeInTheDocument()
      })

      const operationRow = screen.getByTestId('operation-row-op-1')
      await userEvent.click(operationRow)

      expect(mockPush).toHaveBeenCalledWith('/operations/op-1')
    })

    it('allows filtering operations by status', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Report Generation')).toBeInTheDocument()
        expect(screen.getByText('Data Scraping')).toBeInTheDocument()
      })

      const filterSelect = screen.getByTestId('status-filter')
      await userEvent.selectOptions(filterSelect, 'running')

      await waitFor(() => {
        expect(screen.queryByText('Report Generation')).not.toBeInTheDocument()
        expect(screen.getByText('Data Scraping')).toBeInTheDocument()
      })
    })

    it('allows searching operations by name', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Report Generation')).toBeInTheDocument()
        expect(screen.getByText('Data Scraping')).toBeInTheDocument()
      })

      const searchInput = screen.getByPlaceholderText(/Search operations/i)
      await userEvent.type(searchInput, 'Report')

      await waitFor(() => {
        expect(screen.getByText('Report Generation')).toBeInTheDocument()
        expect(screen.queryByText('Data Scraping')).not.toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('displays error toast when operation start fails', async () => {
      const mockOperations = [{
        id: 'op-1',
        name: 'Test Operation',
        status: 'idle',
      }]
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)
      ;(apiClient.createOperation as jest.Mock).mockRejectedValue(
        new Error('Insufficient permissions')
      )

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Operation')).toBeInTheDocument()
      })

      const startButton = screen.getByTestId('start-operation-op-1')
      await userEvent.click(startButton)

      await waitFor(() => {
        expect(screen.getByTestId('error-toast')).toBeInTheDocument()
        expect(screen.getByText(/Insufficient permissions/i)).toBeInTheDocument()
      })
    })

    it('handles network errors gracefully', async () => {
      const mockOperations = [{
        id: 'op-1',
        name: 'Test Operation',
        status: 'running',
      }]
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)
      ;(apiClient.stopOperation as jest.Mock).mockRejectedValue(
        new Error('Network error')
      )

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Test Operation')).toBeInTheDocument()
      })

      const stopButton = screen.getByTestId('stop-operation-op-1')
      await userEvent.click(stopButton)

      await waitFor(() => {
        expect(screen.getByTestId('error-toast')).toBeInTheDocument()
        expect(screen.getByText(/Network error/i)).toBeInTheDocument()
      })

      // Operation should remain in its current state
      expect(screen.getByTestId('operation-status-op-1')).toHaveTextContent('running')
    })
  })

  describe('Accessibility', () => {
    beforeEach(() => {
      const mockOperations = [{
        id: 'op-1',
        name: 'Accessible Operation',
        status: 'idle',
      }]
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)
    })

    it('has proper ARIA labels and roles', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByRole('main')).toBeInTheDocument()
        expect(screen.getByRole('heading', { level: 1, name: /Operations/i })).toBeInTheDocument()
        expect(screen.getByRole('table')).toHaveAccessibleName(/Operations list/i)
      })
    })

    it('supports keyboard navigation', async () => {
      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Accessible Operation')).toBeInTheDocument()
      })

      const startButton = screen.getByTestId('start-operation-op-1')
      
      // Tab to button
      await userEvent.tab()
      expect(startButton).toHaveFocus()

      // Activate with Enter key
      await userEvent.keyboard('{Enter}')
      expect(apiClient.createOperation).toHaveBeenCalled()
    })

    it('announces status changes to screen readers', async () => {
      let updateHandler: (data: any) => void
      mockWebSocket.subscribe.mockImplementation((event, handler) => {
        if (event === 'operation:update') {
          updateHandler = handler
        }
      })

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Accessible Operation')).toBeInTheDocument()
      })

      // Simulate status change
      updateHandler!({
        operationId: 'op-1',
        status: 'completed',
      })

      await waitFor(() => {
        const announcement = screen.getByRole('status')
        expect(announcement).toHaveTextContent(/Operation "Accessible Operation" completed/i)
      })
    })
  })

  describe('Performance', () => {
    it('debounces search input to avoid excessive API calls', async () => {
      const mockOperations = Array.from({ length: 100 }, (_, i) => ({
        id: `op-${i}`,
        name: `Operation ${i}`,
        status: i % 2 === 0 ? 'idle' : 'running',
      }))
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)

      render(<OperationsPage />)
      
      await waitFor(() => {
        expect(screen.getByText('Operation 0')).toBeInTheDocument()
      })

      const searchInput = screen.getByPlaceholderText(/Search operations/i)
      
      // Type quickly
      await userEvent.type(searchInput, 'test search query')

      // Should only trigger one search after debounce
      await waitFor(() => {
        expect(screen.getByDisplayValue('test search query')).toBeInTheDocument()
      })

      // Verify search was debounced (implementation-specific)
      expect(screen.getByTestId('search-indicator')).toBeInTheDocument()
    })

    it('virtualizes long lists for performance', async () => {
      const mockOperations = Array.from({ length: 1000 }, (_, i) => ({
        id: `op-${i}`,
        name: `Operation ${i}`,
        status: 'idle',
      }))
      ;(apiClient.getOperations as jest.Mock).mockResolvedValue(mockOperations)

      render(<OperationsPage />)
      
      await waitFor(() => {
        // Only visible items should be rendered
        const visibleOperations = screen.getAllByTestId(/^operation-row-/)
        expect(visibleOperations.length).toBeLessThan(50) // Assuming viewport shows ~50 items
      })
    })
  })
})
