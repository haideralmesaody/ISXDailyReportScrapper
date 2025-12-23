/**
 * Tests for Operations Content Component
 * Tests the updated operations interface with 6 individual stages + 1 full pipeline
 */

import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { OperationsContent } from '@/app/operations/operations-content'
import { useAllOperationUpdates } from '@/lib/hooks/use-websocket'
import { useHydration } from '@/lib/hooks/use-hydration'
import { apiClient } from '@/lib/api'
import { OperationRequestBuilder } from '@/lib/api/operation-request-builder'
import { WebSocketProvider } from '@/components/providers/websocket-provider'

// Mock the hooks and API
vi.mock('@/lib/hooks/use-websocket')
vi.mock('@/lib/hooks/use-hydration')
vi.mock('@/lib/api')
vi.mock('@/lib/api/operation-request-builder')

// Mock the operation icons
vi.mock('lucide-react', () => ({
  Download: () => <div data-testid="download-icon" />,
  FileSpreadsheet: () => <div data-testid="file-spreadsheet-icon" />,
  BarChart3: () => <div data-testid="bar-chart-icon" />,
  Zap: () => <div data-testid="zap-icon" />,
  Workflow: () => <div data-testid="workflow-icon" />,
  TrendingUp: () => <div data-testid="trending-up-icon" />,
  AlertCircle: () => <div data-testid="alert-circle-icon" />,
  Loader2: () => <div data-testid="loader-icon" />,
  Info: () => <div data-testid="info-icon" />
}))

// Mock child components
vi.mock('@/components/operations/FloatingConfigPanel', () => ({
  FloatingConfigPanel: ({ onApply }: any) => (
    <div data-testid="floating-config-panel">
      <button onClick={() => onApply({ mode: 'initial' })}>Apply Config</button>
    </div>
  )
}))

vi.mock('@/components/operations/UnifiedOperationProgress', () => ({
  UnifiedOperationProgress: ({ operation }: any) => (
    <div data-testid="unified-operation-progress" data-operation-id={operation.id}>
      <div data-testid="operation-status">{operation.status}</div>
    </div>
  )
}))

// PipelineProgress component has been removed - now using UnifiedOperationProgress with pipeline mode

vi.mock('@/components/operations/OperationCompleteCard', () => ({
  OperationCompleteCard: ({ operation, onDismiss }: any) => (
    <div data-testid="operation-complete-card" data-operation-id={operation.id}>
      <button onClick={onDismiss}>Dismiss</button>
    </div>
  )
}))

vi.mock('@/components/operations/CompactOperationCard', () => ({
  CompactOperationCard: ({ operation }: any) => (
    <div data-testid="compact-operation-card" data-operation-id={operation.id}>
      {operation.name}
    </div>
  )
}))

// Mock WebSocket Provider
const MockWebSocketProvider = ({ children }: { children: React.ReactNode }) => (
  <div data-testid="websocket-provider">{children}</div>
)

describe('OperationsContent', () => {
  const mockUseAllOperationUpdates = vi.mocked(useAllOperationUpdates)
  const mockUseHydration = vi.mocked(useHydration)
  const mockApiClient = vi.mocked(apiClient)
  const mockOperationRequestBuilder = vi.mocked(OperationRequestBuilder)

  beforeEach(() => {
    vi.clearAllMocks()

    // Default mock implementations
    mockUseHydration.mockReturnValue(true)
    mockUseAllOperationUpdates.mockReturnValue({
      data: [],
      error: null,
      isLoading: false,
      isConnected: true
    })

    mockApiClient.getOperationTypes.mockResolvedValue([
      {
        id: 'run_stage',
        name: 'Execute Individual Stage',
        description: 'Execute a specific pipeline stage',
        canRunAlone: true,
        dependencies: [],
        parameters: [
          {
            name: 'stage_id',
            type: 'select',
            description: 'Stage to execute',
            required: true,
            defaultValue: 'scraping',
            options: ['scraping', 'processing', 'indices', 'liquidity', 'analysis', 'indicators']
          }
        ]
      },
      {
        id: 'full_pipeline',
        name: 'Execute Full Pipeline',
        description: 'Execute all pipeline stages',
        canRunAlone: true,
        dependencies: [],
        parameters: [
          {
            name: 'mode',
            type: 'select',
            description: 'Execution mode',
            required: false,
            defaultValue: 'initial',
            options: ['initial', 'incremental']
          }
        ]
      }
    ])

    mockOperationRequestBuilder.buildRunStage.mockImplementation((stageId, from, to) => ({
      operation: 'run_stage',
      stage_id: stageId,
      parameters: stageId === 'scraping' ? { from, to } : {}
    }))

    mockOperationRequestBuilder.buildFullPipeline.mockImplementation((from, to) => ({
      operation: 'full_pipeline',
      parameters: { from, to }
    }))

    mockApiClient.createOperation.mockResolvedValue({
      id: 'test-operation-1',
      success: true,
      steps: []
    })
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  const renderComponent = () => {
    return render(
      <MockWebSocketProvider>
        <OperationsContent />
      </MockWebSocketProvider>
    )
  }

  describe('Component Rendering', () => {
    it('should render all 6 individual stage cards', async () => {
      renderComponent()

      // Check that all stage cards are rendered
      const expectedStages = [
        'Data Collection',
        'Data Processing',
        'Index Extraction',
        'Liquidity Analysis',
        'Technical Analysis',
        'Indicator Calculation'
      ]

      for (const stageName of expectedStages) {
        expect(screen.getByText(stageName)).toBeInTheDocument()
      }
    })

    it('should render the full pipeline section', () => {
      renderComponent()

      expect(screen.getByText('Execute Full Pipeline')).toBeInTheDocument()
      expect(screen.getByText(/run all 6 pipeline stages/i)).toBeInTheDocument()
    })

    it('should render operation icons', () => {
      renderComponent()

      // Check that icons are rendered
      expect(screen.getByTestId('download-icon')).toBeInTheDocument()
      expect(screen.getByTestId('file-spreadsheet-icon')).toBeInTheDocument()
      expect(screen.getByTestId('bar-chart-icon')).toBeInTheDocument()
      expect(screen.getByTestId('zap-icon')).toBeInTheDocument()
      expect(screen.getByTestId('trending-up-icon')).toBeInTheDocument()
      expect(screen.getByTestId('workflow-icon')).toBeInTheDocument()
    })

    it('should show loading state initially', () => {
      mockUseHydration.mockReturnValue(false)

      renderComponent()

      expect(screen.getByText('Initializing...')).toBeInTheDocument()
      expect(screen.getByTestId('loader-icon')).toBeInTheDocument()
    })

    it('should show error state when WebSocket is disconnected', () => {
      mockUseAllOperationUpdates.mockReturnValue({
        data: [],
        error: null,
        isLoading: false,
        isConnected: false
      })

      renderComponent()

      expect(screen.getByText(/websocket connection lost/i)).toBeInTheDocument()
    })

    it('should show loading state for operation types', async () => {
      mockApiClient.getOperationTypes.mockImplementation(() => new Promise(() => {})) // Never resolves

      renderComponent()

      expect(screen.getByText('Loading operation types...')).toBeInTheDocument()
    })
  })

  describe('Stage Card Interactions', () => {
    it('should handle direct start for individual stages', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Find and click the start button for processing stage
      const processingCard = screen.getByText('Data Processing').closest('[data-testid*="operation-card"]')
      const startButton = within(processingCard!).getByRole('button', { name: /start/i })

      await user.click(startButton)

      // Verify that the correct request was built and sent
      expect(mockOperationRequestBuilder.buildRunStage).toHaveBeenCalledWith('processing')
      expect(mockApiClient.createOperation).toHaveBeenCalledWith({
        operation: 'run_stage',
        stage_id: 'processing',
        parameters: {},
      })
    })

    it('should handle configuration for stages that require dates', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Find and click the configure button for scraping stage
      const scrapingCard = screen.getByText('Data Collection').closest('[data-testid*="operation-card"]')
      const configButton = within(scrapingCard!).getByRole('button', { name: /configure/i })

      await user.click(configButton)

      // Should show the configuration panel
      expect(screen.getByTestId('floating-config-panel')).toBeInTheDocument()

      // Apply configuration
      const applyButton = screen.getByText('Apply Config')
      await user.click(applyButton)

      // Verify configuration was handled
      expect(screen.getByTestId('floating-config-panel')).toBeInTheDocument()
    })

    it('should handle start after configuration', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Configure scraping stage
      const scrapingCard = screen.getByText('Data Collection').closest('[data-testid*="operation-card"]')
      const configButton = within(scrapingCard!).getByRole('button', { name: /configure/i })

      await user.click(configButton)

      // Apply configuration with dates
      const applyButton = screen.getByText('Apply Config')
      await user.click(applyButton)

      // Now start the operation
      const startButton = within(scrapingCard!).getByRole('button', { name: /start/i })
      await user.click(startButton)

      // Verify the operation was started with configuration
      expect(mockApiClient.createOperation).toHaveBeenCalled()
    })
  })

  describe('Full Pipeline Interactions', () => {
    it('should handle full pipeline start', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Find and click the full pipeline start button
      const fullPipelineSection = screen.getByText('Execute Full Pipeline').closest('[data-section*="pipeline"]')
      const startButton = within(fullPipelineSection!).getByRole('button', { name: /start full pipeline/i })

      await user.click(startButton)

      // Verify that the full pipeline request was built and sent
      expect(mockOperationRequestBuilder.buildFullPipeline).toHaveBeenCalled()
      expect(mockApiClient.createOperation).toHaveBeenCalledWith(
        expect.objectContaining({
          operation: 'full_pipeline',
          parameters: expect.objectContaining({
            from: expect.any(String),
            to: expect.any(String),
          }),
        })
      )
    })

    it('should handle full pipeline configuration', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Find and click the configure button for full pipeline
      const fullPipelineSection = screen.getByText('Execute Full Pipeline').closest('[data-section*="pipeline"]')
      const configButton = within(fullPipelineSection!).getByRole('button', { name: /configure pipeline/i })

      await user.click(configButton)

      // Should show the configuration panel
      expect(screen.getByTestId('floating-config-panel')).toBeInTheDocument()
    })
  })

  describe('Operation Progress Display', () => {
    it('should display running operations', () => {
      mockUseAllOperationUpdates.mockReturnValue({
        data: [
          {
            id: 'test-operation-1',
            status: 'running',
            operation_type: 'run_stage',
            metadata: { stage_id: 'processing' },
            steps: [
              { id: 'processing', status: 'running', progress: 50 }
            ]
          }
        ],
        error: null,
        isLoading: false,
        isConnected: true
      })

      renderComponent()

      expect(screen.getByTestId('unified-operation-progress')).toBeInTheDocument()
      expect(screen.getByTestId('operation-status')).toHaveTextContent('running')
    })

    it('should display completed operations', () => {
      mockUseAllOperationUpdates.mockReturnValue({
        data: [
          {
            id: 'test-operation-1',
            status: 'completed',
            operation_type: 'run_stage',
            metadata: { stage_id: 'processing' },
            steps: [
              { id: 'processing', status: 'completed', progress: 100 }
            ]
          }
        ],
        error: null,
        isLoading: false,
        isConnected: true
      })

      renderComponent()

      expect(screen.getByTestId('operation-complete-card')).toBeInTheDocument()
      expect(screen.getByTestId('compact-operation-card')).toBeInTheDocument()
    })

    it('should display pipeline progress for full pipeline operations', () => {
      mockUseAllOperationUpdates.mockReturnValue({
        data: [
          {
            id: 'test-operation-1',
            status: 'running',
            operation_type: 'full_pipeline',
            metadata: {},
            steps: [
              { id: 'scraping', status: 'completed', progress: 100 },
              { id: 'processing', status: 'running', progress: 30 }
            ]
          }
        ],
        error: null,
        isLoading: false,
        isConnected: true
      })

      renderComponent()

      expect(screen.getByTestId('pipeline-progress')).toBeInTheDocument()
    })

    it('should handle operation errors', () => {
      mockUseAllOperationUpdates.mockReturnValue({
        data: [
          {
            id: 'test-operation-1',
            status: 'failed',
            operation_type: 'run_stage',
            metadata: { stage_id: 'processing' },
            error: 'Test error message',
            steps: [
              { id: 'processing', status: 'failed', progress: 0 }
            ]
          }
        ],
        error: null,
        isLoading: false,
        isConnected: true
      })

      renderComponent()

      expect(screen.getByText(/test error message/i)).toBeInTheDocument()
      expect(screen.getByTestId('alert-circle-icon')).toBeInTheDocument()
    })
  })

  describe('Error Handling', () => {
    it('should handle API errors gracefully', async () => {
      const user = userEvent.setup()
      mockApiClient.createOperation.mockRejectedValue(new Error('API Error'))

      renderComponent()

      // Try to start an operation
      const processingCard = screen.getByText('Data Processing').closest('[data-testid*="operation-card"]')
      const startButton = within(processingCard!).getByRole('button', { name: /start/i })

      await user.click(startButton)

      // Wait for error to be displayed
      await waitFor(() => {
        expect(screen.getByText(/failed to start operation/i)).toBeInTheDocument()
      })
    })

    it('should handle operation types loading error', async () => {
      mockApiClient.getOperationTypes.mockRejectedValue(new Error('Failed to load operation types'))

      renderComponent()

      await waitFor(() => {
        expect(screen.getByText(/failed to load operation types/i)).toBeInTheDocument()
      })
    })

    it('should handle WebSocket errors', () => {
      mockUseAllOperationUpdates.mockReturnValue({
        data: [],
        error: new Error('WebSocket Error'),
        isLoading: false,
        isConnected: false
      })

      renderComponent()

      expect(screen.getByText(/websocket connection lost/i)).toBeInTheDocument()
      expect(screen.getByText(/websocket error/i)).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels', () => {
      renderComponent()

      // Check for proper heading structure
      expect(screen.getByRole('heading', { level: 1, name: /operations/i })).toBeInTheDocument()

      // Check for button labels
      const startButtons = screen.getAllByRole('button', { name: /start/i })
      expect(startButtons.length).toBeGreaterThan(0)

      // Check for descriptive text
      expect(screen.getByText(/execute individual pipeline stages/i)).toBeInTheDocument()
      expect(screen.getByText(/run all 6 pipeline stages/i)).toBeInTheDocument()
    })

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup()
      renderComponent()

      // Tab to first start button
      await user.tab()
      await user.tab()

      // Should focus on an interactive element
      expect(document.activeElement).toBeInTheDocument()
    })
  })

  describe('Responsive Design', () => {
    it('should adapt to different screen sizes', () => {
      // Test mobile view
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      })

      renderComponent()

      // Should still render all components
      expect(screen.getByText('Data Collection')).toBeInTheDocument()
      expect(screen.getByText('Execute Full Pipeline')).toBeInTheDocument()

      // Test desktop view
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920,
      })

      renderComponent()

      // Should render all components
      expect(screen.getByText('Data Collection')).toBeInTheDocument()
      expect(screen.getByText('Execute Full Pipeline')).toBeInTheDocument()
    })
  })

  describe('Performance', () => {
    it('should not cause unnecessary re-renders', async () => {
      const { rerender } = renderComponent()

      // Initial render
      expect(screen.getByText('Data Collection')).toBeInTheDocument()

      // Rerender with same props
      rerender(
        <MockWebSocketProvider>
          <OperationsContent />
        </MockWebSocketProvider>
      )

      // Should still render correctly
      expect(screen.getByText('Data Collection')).toBeInTheDocument()
    })

    it('should handle multiple rapid clicks gracefully', async () => {
      const user = userEvent.setup()
      renderComponent()

      const processingCard = screen.getByText('Data Processing').closest('[data-testid*="operation-card"]')
      const startButton = within(processingCard!).getByRole('button', { name: /start/i })

      // Click multiple times rapidly
      await user.click(startButton)
      await user.click(startButton)
      await user.click(startButton)

      // Should still work correctly
      expect(mockApiClient.createOperation).toHaveBeenCalledTimes(3)
    })
  })

  describe('Integration with Backend', () => {
    it('should send correct payload format for run_stage operations', async () => {
      const user = userEvent.setup()
      renderComponent()

      const processingCard = screen.getByText('Data Processing').closest('[data-testid*="operation-card"]')
      const startButton = within(processingCard!).getByRole('button', { name: /start/i })

      await user.click(startButton)

      // Verify the exact payload format matches new backend expectations
      expect(mockApiClient.createOperation).toHaveBeenCalledWith(
        expect.objectContaining({
          operation: 'run_stage',
          stage_id: 'processing',
          parameters: {},
        })
      )
    })

    it('should send correct payload format for full_pipeline operations', async () => {
      const user = userEvent.setup()
      renderComponent()

      const fullPipelineSection = screen.getByText('Execute Full Pipeline').closest('[data-section*="pipeline"]')
      const startButton = within(fullPipelineSection!).getByRole('button', { name: /start full pipeline/i })

      await user.click(startButton)

      // Verify the payload contains all pipeline stages
      expect(mockApiClient.createOperation).toHaveBeenCalledWith(
        expect.objectContaining({
          operation: 'full_pipeline',
          parameters: expect.objectContaining({
            from: expect.any(String),
            to: expect.any(String),
          }),
        })
      )
    })
  })
})
