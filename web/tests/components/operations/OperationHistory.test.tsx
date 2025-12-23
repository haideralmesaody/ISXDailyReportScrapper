import React from 'react'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OperationHistory } from '@/components/operations/OperationHistory'
import { apiClient } from '@/lib/api'
import '@testing-library/jest-dom'

// Mock dependencies
jest.mock('@/lib/api')

// Mock IntersectionObserver for virtualization
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}))

describe('OperationHistory', () => {
  const mockHistoryData = [
    {
      id: 'run-1',
      operationId: 'op-123',
      operationName: 'Daily Report Generation',
      status: 'completed',
      startedAt: '2025-01-30T10:00:00Z',
      completedAt: '2025-01-30T10:30:00Z',
      duration: 1800,
      progress: 100,
      triggeredBy: 'schedule',
      steps: [
        { name: 'Data Collection', duration: 600, status: 'completed' },
        { name: 'Processing', duration: 900, status: 'completed' },
        { name: 'Report Generation', duration: 300, status: 'completed' },
      ],
    },
    {
      id: 'run-2',
      operationId: 'op-123',
      operationName: 'Daily Report Generation',
      status: 'failed',
      startedAt: '2025-01-29T10:00:00Z',
      completedAt: '2025-01-29T10:15:00Z',
      duration: 900,
      progress: 45,
      triggeredBy: 'schedule',
      error: 'Failed to connect to data source',
      steps: [
        { name: 'Data Collection', duration: 600, status: 'completed' },
        { name: 'Processing', duration: 300, status: 'failed', error: 'Connection timeout' },
        { name: 'Report Generation', duration: 0, status: 'cancelled' },
      ],
    },
    {
      id: 'run-3',
      operationId: 'op-123',
      operationName: 'Daily Report Generation',
      status: 'completed',
      startedAt: '2025-01-28T10:00:00Z',
      completedAt: '2025-01-28T10:25:00Z',
      duration: 1500,
      progress: 100,
      triggeredBy: 'manual',
      triggeredByUser: 'admin@example.com',
      steps: [
        { name: 'Data Collection', duration: 500, status: 'completed' },
        { name: 'Processing', duration: 800, status: 'completed' },
        { name: 'Report Generation', duration: 200, status: 'completed' },
      ],
    },
  ]

  const defaultProps = {
    operationId: 'op-123',
    operationName: 'Daily Report Generation',
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(apiClient.getOperationHistory as jest.Mock).mockResolvedValue({
      items: mockHistoryData,
      total: 3,
      page: 1,
      pageSize: 20,
    })
  })

  describe('History Display', () => {
    const displayTestCases = [
      {
        name: 'loads and displays operation history',
        assertions: async () => {
          await waitFor(() => {
            expect(screen.getByTestId('history-list')).toBeInTheDocument()
          })
          
          expect(screen.getByText(/3 runs found/i)).toBeInTheDocument()
          expect(screen.getAllByTestId(/^history-item-/)).toHaveLength(3)
        },
      },
      {
        name: 'shows run details with status indicators',
        assertions: async () => {
          await waitFor(() => {
            expect(screen.getByTestId('history-item-run-1')).toBeInTheDocument()
          })
          
          const run1 = screen.getByTestId('history-item-run-1')
          expect(within(run1).getByText('Completed')).toHaveClass('text-green-600')
          expect(within(run1).getByText('30m')).toBeInTheDocument() // Duration
          expect(within(run1).getByText('Schedule')).toBeInTheDocument() // Trigger
          
          const run2 = screen.getByTestId('history-item-run-2')
          expect(within(run2).getByText('Failed')).toHaveClass('text-red-600')
          expect(within(run2).getByText(/Failed to connect/i)).toBeInTheDocument()
        },
      },
      {
        name: 'displays trigger information',
        assertions: async () => {
          await waitFor(() => {
            expect(screen.getByTestId('history-item-run-3')).toBeInTheDocument()
          })
          
          const run3 = screen.getByTestId('history-item-run-3')
          expect(within(run3).getByText('Manual')).toBeInTheDocument()
          expect(within(run3).getByText('admin@example.com')).toBeInTheDocument()
        },
      },
      {
        name: 'shows loading state while fetching',
        setup: () => {
          ;(apiClient.getOperationHistory as jest.Mock).mockReturnValue(
            new Promise(() => {}) // Never resolves
          )
        },
        assertions: () => {
          expect(screen.getByTestId('history-loading')).toBeInTheDocument()
          expect(screen.getByText(/Loading history/i)).toBeInTheDocument()
        },
      },
      {
        name: 'displays empty state when no history exists',
        setup: () => {
          ;(apiClient.getOperationHistory as jest.Mock).mockResolvedValue({
            items: [],
            total: 0,
            page: 1,
            pageSize: 20,
          })
        },
        assertions: async () => {
          await waitFor(() => {
            expect(screen.getByTestId('history-empty')).toBeInTheDocument()
          })
          
          expect(screen.getByText(/No operation history found/i)).toBeInTheDocument()
        },
      },
    ]

    displayTestCases.forEach(({ name, setup, assertions }) => {
      it(name, async () => {
        if (setup) setup()
        render(<OperationHistory {...defaultProps} />)
        await assertions()
      })
    })
  })

  describe('Filtering and Sorting', () => {
    beforeEach(async () => {
      render(<OperationHistory {...defaultProps} />)
      await waitFor(() => {
        expect(screen.getByTestId('history-list')).toBeInTheDocument()
      })
    })

    it('filters by status', async () => {
      const statusFilter = screen.getByLabelText(/Filter by status/i)
      await userEvent.selectOptions(statusFilter, 'failed')
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        status: 'failed',
        page: 1,
        pageSize: 20,
      })
    })

    it('filters by date range', async () => {
      const startDateInput = screen.getByLabelText(/Start date/i)
      const endDateInput = screen.getByLabelText(/End date/i)
      
      await userEvent.type(startDateInput, '2025-01-28')
      await userEvent.type(endDateInput, '2025-01-30')
      
      const applyButton = screen.getByRole('button', { name: /Apply filters/i })
      await userEvent.click(applyButton)
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        startDate: '2025-01-28T00:00:00.000Z',
        endDate: '2025-01-30T23:59:59.999Z',
        page: 1,
        pageSize: 20,
      })
    })

    it('filters by trigger type', async () => {
      const triggerFilter = screen.getByLabelText(/Trigger type/i)
      await userEvent.selectOptions(triggerFilter, 'manual')
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        triggeredBy: 'manual',
        page: 1,
        pageSize: 20,
      })
    })

    it('sorts by different columns', async () => {
      const sortSelect = screen.getByLabelText(/Sort by/i)
      await userEvent.selectOptions(sortSelect, 'duration')
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        sortBy: 'duration',
        sortOrder: 'desc',
        page: 1,
        pageSize: 20,
      })
      
      // Toggle sort order
      const sortOrderButton = screen.getByRole('button', { name: /Toggle sort order/i })
      await userEvent.click(sortOrderButton)
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        sortBy: 'duration',
        sortOrder: 'asc',
        page: 1,
        pageSize: 20,
      })
    })

    it('searches by error message', async () => {
      const searchInput = screen.getByPlaceholderText(/Search errors/i)
      await userEvent.type(searchInput, 'connection timeout')
      
      // Debounced search
      await waitFor(() => {
        expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
          operationId: 'op-123',
          search: 'connection timeout',
          page: 1,
          pageSize: 20,
        })
      })
    })

    it('clears all filters', async () => {
      // Apply some filters first
      const statusFilter = screen.getByLabelText(/Filter by status/i)
      await userEvent.selectOptions(statusFilter, 'failed')
      
      const clearButton = screen.getByRole('button', { name: /Clear filters/i })
      await userEvent.click(clearButton)
      
      expect(apiClient.getOperationHistory).toHaveBeenLastCalledWith({
        operationId: 'op-123',
        page: 1,
        pageSize: 20,
      })
    })
  })

  describe('Detailed View', () => {
    beforeEach(async () => {
      render(<OperationHistory {...defaultProps} />)
      await waitFor(() => {
        expect(screen.getByTestId('history-list')).toBeInTheDocument()
      })
    })

    it('expands to show step details', async () => {
      const expandButton = screen.getByTestId('expand-run-1')
      
      // Initially collapsed
      expect(screen.queryByTestId('run-1-steps')).not.toBeInTheDocument()
      
      await userEvent.click(expandButton)
      
      // Shows step breakdown
      const stagesSection = screen.getByTestId('run-1-steps')
      expect(stagesSection).toBeInTheDocument()
      expect(within(stagesSection).getByText('Data Collection')).toBeInTheDocument()
      expect(within(stagesSection).getByText('10m')).toBeInTheDocument() // Duration
      expect(within(stagesSection).getByText('Processing')).toBeInTheDocument()
      expect(within(stagesSection).getByText('15m')).toBeInTheDocument()
    })

    it('shows error details for failed runs', async () => {
      const expandButton = screen.getByTestId('expand-run-2')
      await userEvent.click(expandButton)
      
      const details = screen.getByTestId('run-2-details')
      expect(within(details).getByText(/Connection timeout/i)).toBeInTheDocument()
      expect(within(details).getByTestId('step-error-Processing')).toBeInTheDocument()
    })

    it('displays performance metrics when available', async () => {
      // Mock history with performance data
      const historyWithMetrics = [{
        ...mockHistoryData[0],
        metrics: {
          recordsProcessed: 5000,
          averageProcessingTime: 0.36, // seconds per record
          peakMemoryUsage: 512, // MB
          cpuUtilization: 65, // percentage
        },
      }]
      
      ;(apiClient.getOperationHistory as jest.Mock).mockResolvedValue({
        items: historyWithMetrics,
        total: 1,
        page: 1,
        pageSize: 20,
      })
      
      const { rerender } = render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByTestId('expand-run-1')).toBeInTheDocument()
      })
      
      const expandButton = screen.getByTestId('expand-run-1')
      await userEvent.click(expandButton)
      
      const metrics = screen.getByTestId('run-1-metrics')
      expect(within(metrics).getByText('5,000')).toBeInTheDocument() // Records
      expect(within(metrics).getByText('0.36s/record')).toBeInTheDocument()
      expect(within(metrics).getByText('512 MB')).toBeInTheDocument()
      expect(within(metrics).getByText('65%')).toBeInTheDocument()
    })
  })

  describe('Pagination', () => {
    beforeEach(() => {
      // Mock paginated response
      ;(apiClient.getOperationHistory as jest.Mock).mockResolvedValue({
        items: mockHistoryData,
        total: 100,
        page: 1,
        pageSize: 20,
        totalPages: 5,
      })
    })

    it('displays pagination controls', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByTestId('pagination')).toBeInTheDocument()
      })
      
      expect(screen.getByText('Page 1 of 5')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Previous/i })).toBeDisabled()
      expect(screen.getByRole('button', { name: /Next/i })).toBeEnabled()
    })

    it('navigates between pages', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByTestId('pagination')).toBeInTheDocument()
      })
      
      const nextButton = screen.getByRole('button', { name: /Next/i })
      await userEvent.click(nextButton)
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        page: 2,
        pageSize: 20,
      })
    })

    it('allows changing page size', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByLabelText(/Items per page/i)).toBeInTheDocument()
      })
      
      const pageSizeSelect = screen.getByLabelText(/Items per page/i)
      await userEvent.selectOptions(pageSizeSelect, '50')
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        page: 1,
        pageSize: 50,
      })
    })

    it('allows jumping to specific page', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByTestId('page-input')).toBeInTheDocument()
      })
      
      const pageInput = screen.getByTestId('page-input')
      await userEvent.clear(pageInput)
      await userEvent.type(pageInput, '3')
      await userEvent.keyboard('{Enter}')
      
      expect(apiClient.getOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        page: 3,
        pageSize: 20,
      })
    })
  })

  describe('Export Functionality', () => {
    beforeEach(async () => {
      render(<OperationHistory {...defaultProps} />)
      await waitFor(() => {
        expect(screen.getByTestId('history-list')).toBeInTheDocument()
      })
    })

    it('exports history as CSV', async () => {
      const exportButton = screen.getByRole('button', { name: /Export/i })
      await userEvent.click(exportButton)
      
      const csvOption = screen.getByRole('menuitem', { name: /Export as CSV/i })
      await userEvent.click(csvOption)
      
      expect(apiClient.exportOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        format: 'csv',
        filters: {},
      })
    })

    it('exports history as JSON', async () => {
      const exportButton = screen.getByRole('button', { name: /Export/i })
      await userEvent.click(exportButton)
      
      const jsonOption = screen.getByRole('menuitem', { name: /Export as JSON/i })
      await userEvent.click(jsonOption)
      
      expect(apiClient.exportOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        format: 'json',
        filters: {},
      })
    })

    it('includes current filters in export', async () => {
      // Apply filter first
      const statusFilter = screen.getByLabelText(/Filter by status/i)
      await userEvent.selectOptions(statusFilter, 'failed')
      
      await waitFor(() => {
        expect(apiClient.getOperationHistory).toHaveBeenCalled()
      })
      
      const exportButton = screen.getByRole('button', { name: /Export/i })
      await userEvent.click(exportButton)
      
      const csvOption = screen.getByRole('menuitem', { name: /Export as CSV/i })
      await userEvent.click(csvOption)
      
      expect(apiClient.exportOperationHistory).toHaveBeenCalledWith({
        operationId: 'op-123',
        format: 'csv',
        filters: { status: 'failed' },
      })
    })
  })

  describe('Real-time Updates', () => {
    it('prepends new runs to the list', async () => {
      const { rerender } = render(<OperationHistory {...defaultProps} enableRealtime />)
      
      await waitFor(() => {
        expect(screen.getAllByTestId(/^history-item-/)).toHaveLength(3)
      })
      
      // Simulate new run via WebSocket or polling
      const newRun = {
        id: 'run-4',
        operationId: 'op-123',
        operationName: 'Daily Report Generation',
        status: 'running',
        startedAt: '2025-01-30T15:00:00Z',
        progress: 25,
        triggeredBy: 'api',
      }
      
      ;(apiClient.getOperationHistory as jest.Mock).mockResolvedValue({
        items: [newRun, ...mockHistoryData],
        total: 4,
        page: 1,
        pageSize: 20,
      })
      
      // Trigger refresh
      rerender(<OperationHistory {...defaultProps} enableRealtime />)
      
      await waitFor(() => {
        expect(screen.getAllByTestId(/^history-item-/)).toHaveLength(4)
        expect(screen.getByTestId('history-item-run-4')).toBeInTheDocument()
      })
    })

    it('updates existing run status', async () => {
      render(<OperationHistory {...defaultProps} enableRealtime />)
      
      await waitFor(() => {
        expect(screen.getByTestId('history-item-run-1')).toBeInTheDocument()
      })
      
      // Update mock to change run status
      const updatedHistory = mockHistoryData.map(run =>
        run.id === 'run-1' ? { ...run, status: 'failed', error: 'Unexpected error' } : run
      )
      
      ;(apiClient.getOperationHistory as jest.Mock).mockResolvedValue({
        items: updatedHistory,
        total: 3,
        page: 1,
        pageSize: 20,
      })
      
      // Trigger polling update
      act(() => {
        jest.advanceTimersByTime(5000) // Assuming 5-second polling
      })
      
      await waitFor(() => {
        const run1 = screen.getByTestId('history-item-run-1')
        expect(within(run1).getByText('Failed')).toBeInTheDocument()
        expect(within(run1).getByText(/Unexpected error/i)).toBeInTheDocument()
      })
    })
  })

  describe('Analytics View', () => {
    it('shows summary statistics', async () => {
      render(<OperationHistory {...defaultProps} showAnalytics />)
      
      await waitFor(() => {
        expect(screen.getByTestId('history-analytics')).toBeInTheDocument()
      })
      
      const analytics = screen.getByTestId('history-analytics')
      expect(within(analytics).getByText(/Success rate: 66.7%/i)).toBeInTheDocument()
      expect(within(analytics).getByText(/Average duration: 22.5m/i)).toBeInTheDocument()
      expect(within(analytics).getByText(/Total runs: 3/i)).toBeInTheDocument()
    })

    it('displays trend charts', async () => {
      render(<OperationHistory {...defaultProps} showAnalytics />)
      
      await waitFor(() => {
        expect(screen.getByTestId('duration-trend-chart')).toBeInTheDocument()
      })
      
      expect(screen.getByTestId('success-rate-chart')).toBeInTheDocument()
      expect(screen.getByTestId('trigger-distribution-chart')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('provides proper table structure and headers', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByRole('table')).toBeInTheDocument()
      })
      
      const table = screen.getByRole('table')
      expect(table).toHaveAccessibleName(/Operation history/i)
      
      // Check headers
      expect(screen.getByRole('columnheader', { name: /Status/i })).toBeInTheDocument()
      expect(screen.getByRole('columnheader', { name: /Started/i })).toBeInTheDocument()
      expect(screen.getByRole('columnheader', { name: /Duration/i })).toBeInTheDocument()
      expect(screen.getByRole('columnheader', { name: /Trigger/i })).toBeInTheDocument()
    })

    it('announces filter changes', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByLabelText(/Filter by status/i)).toBeInTheDocument()
      })
      
      const statusFilter = screen.getByLabelText(/Filter by status/i)
      await userEvent.selectOptions(statusFilter, 'failed')
      
      const announcement = screen.getByRole('status')
      expect(announcement).toHaveTextContent(/Showing failed operations/i)
    })

    it('supports keyboard navigation through history items', async () => {
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getAllByTestId(/^history-item-/)).toHaveLength(3)
      })
      
      // Tab to first expandable item
      await userEvent.tab()
      expect(screen.getByTestId('expand-run-1')).toHaveFocus()
      
      // Navigate with arrow keys
      await userEvent.keyboard('{ArrowDown}')
      expect(screen.getByTestId('expand-run-2')).toHaveFocus()
    })
  })

  describe('Error Handling', () => {
    it('displays error state when loading fails', async () => {
      ;(apiClient.getOperationHistory as jest.Mock).mockRejectedValue(
        new Error('Failed to load history')
      )
      
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByTestId('history-error')).toBeInTheDocument()
      })
      
      expect(screen.getByText(/Failed to load history/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument()
    })

    it('allows retrying after error', async () => {
      ;(apiClient.getOperationHistory as jest.Mock)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          items: mockHistoryData,
          total: 3,
          page: 1,
          pageSize: 20,
        })
      
      render(<OperationHistory {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByTestId('history-error')).toBeInTheDocument()
      })
      
      const retryButton = screen.getByRole('button', { name: /Retry/i })
      await userEvent.click(retryButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('history-list')).toBeInTheDocument()
        expect(screen.getAllByTestId(/^history-item-/)).toHaveLength(3)
      })
    })
  })
})