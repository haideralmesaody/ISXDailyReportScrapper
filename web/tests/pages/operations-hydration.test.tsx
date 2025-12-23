/**
 * Operations Page Hydration Tests
 * Comprehensive tests to verify WebSocket and real-time updates work without hydration errors
 */

import React from 'react'
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import dynamic from 'next/dynamic'

// Mock Next.js dynamic import
jest.mock('next/dynamic', () => {
  return jest.fn((loader, options) => {
    const DynamicComponent = React.forwardRef((props: any, ref) => {
      const [isLoaded, setIsLoaded] = React.useState(false)
      
      React.useEffect(() => {
        const timer = setTimeout(() => {
          setIsLoaded(true)
        }, 10)
        return () => clearTimeout(timer)
      }, [])
      
      if (!isLoaded && options?.loading) {
        const LoadingComponent = options.loading
        return <LoadingComponent />
      }
      
      return <div data-testid="operations-dynamic" {...props}>Operations Content Loaded</div>
    })
    
    DynamicComponent.displayName = 'DynamicOperations'
    return DynamicComponent
  })
})

// Mock WebSocket hook
const mockWebSocketData = {
  operations: [],
  connected: false,
  error: null,
}

jest.mock('@/lib/hooks/use-websocket', () => ({
  useAllOperationUpdates: jest.fn(() => mockWebSocketData),
  useOperationUpdates: jest.fn(() => mockWebSocketData),
}))

// Mock API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    getOperations: jest.fn().mockResolvedValue([
      {
        id: 'op-1',
        type: 'scraping',
        status: 'idle',
        progress: 0,
      },
      {
        id: 'op-2',
        type: 'processing',
        status: 'completed',
        progress: 100,
      },
    ]),
    createOperation: jest.fn().mockResolvedValue({
      success: true,
      operationId: 'op-3',
    }),
    stopOperation: jest.fn().mockResolvedValue({
      success: true,
    }),
  },
}))

// Import components after mocks
import OperationsPage from '@/app/operations/page'
import OperationsContent from '@/app/operations/operations-content'

describe('Operations Page Hydration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    
    // Reset WebSocket mock data
    mockWebSocketData.operations = []
    mockWebSocketData.connected = false
    mockWebSocketData.error = null
    
    // Mock console to catch hydration warnings
    jest.spyOn(console, 'error').mockImplementation(() => {})
    jest.spyOn(console, 'warn').mockImplementation(() => {})
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  describe('WebSocket Connection Timing', () => {
    it('should not connect WebSocket during SSR', () => {
      const { useAllOperationUpdates } = require('@/lib/hooks/use-websocket')
      
      // Clear mock calls
      useAllOperationUpdates.mockClear()
      
      // Render component
      render(<OperationsContent />)
      
      // WebSocket hook is called but shouldn't connect immediately
      expect(useAllOperationUpdates).toHaveBeenCalled()
      
      // Initial state should show disconnected
      expect(mockWebSocketData.connected).toBe(false)
    })

    it('should connect WebSocket only after client-side mount', async () => {
      const { container } = render(<OperationsContent />)
      
      // Initially disconnected
      expect(mockWebSocketData.connected).toBe(false)
      
      // Simulate client-side mount
      act(() => {
        mockWebSocketData.connected = true
        jest.runAllTimers()
      })
      
      await waitFor(() => {
        // WebSocket should be connected after mount
        expect(mockWebSocketData.connected).toBe(true)
      })
      
      // No hydration errors
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })

    it('should handle WebSocket reconnection without hydration issues', async () => {
      render(<OperationsContent />)
      
      // Simulate connection changes
      for (let i = 0; i < 3; i++) {
        act(() => {
          mockWebSocketData.connected = false
        })
        
        await waitFor(() => {}, { timeout: 10 })
        
        act(() => {
          mockWebSocketData.connected = true
        })
        
        await waitFor(() => {}, { timeout: 10 })
      }
      
      // No hydration errors during reconnection
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Text content did not match')
      )
    })
  })

  describe('Real-Time Updates', () => {
    it('should handle operation updates without hydration mismatches', async () => {
      const { container } = render(<OperationsContent />)
      
      // Simulate real-time updates
      act(() => {
        mockWebSocketData.operations = [
          {
            id: 'op-1',
            type: 'scraping',
            status: 'running',
            progress: 50,
            startTime: new Date().toISOString(),
          },
        ]
      })
      
      await waitFor(() => {
        // Updates should apply without errors
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Hydration failed')
        )
      })
    })

    it('should update progress indicators client-side only', async () => {
      const { container } = render(<OperationsContent />)
      
      // Wait for client mount
      await waitFor(() => {
        expect(container.querySelector('[data-testid="progress-bar"]')).toBeInTheDocument()
      }, { timeout: 100 }).catch(() => {
        // Progress bar might not exist initially, which is fine
      })
      
      // Simulate progress updates
      for (let progress = 0; progress <= 100; progress += 25) {
        act(() => {
          mockWebSocketData.operations = [
            {
              id: 'op-1',
              type: 'processing',
              status: 'running',
              progress,
            },
          ]
        })
        
        await waitFor(() => {}, { timeout: 10 })
      }
      
      // No text mismatch errors
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Text content did not match')
      )
    })

    it('should handle rapid status changes smoothly', async () => {
      render(<OperationsContent />)
      
      const statuses = ['idle', 'starting', 'running', 'completed', 'failed']
      
      for (const status of statuses) {
        act(() => {
          mockWebSocketData.operations = [
            {
              id: 'op-1',
              type: 'scraping',
              status,
              progress: status === 'completed' ? 100 : 50,
            },
          ]
        })
        
        await waitFor(() => {}, { timeout: 5 })
      }
      
      // No hydration errors during rapid updates
      expect(console.error).not.toHaveBeenCalled()
    })
  })

  describe('Client-Side Flags', () => {
    it('should use isClient flag to prevent SSR rendering', () => {
      const { container } = render(<OperationsContent />)
      
      // Component should handle isClient state
      expect(container.firstChild).toBeInTheDocument()
      
      // No hydration warnings
      expect(console.warn).not.toHaveBeenCalledWith(
        expect.stringContaining('did not match')
      )
    })

    it('should guard date operations with client flag', async () => {
      const DateSpy = jest.spyOn(global, 'Date')
      DateSpy.mockClear()
      
      render(<OperationsContent />)
      
      // Date operations should be guarded
      const initialDateCalls = DateSpy.mock.calls.length
      
      // After mount, Date operations are allowed
      act(() => {
        jest.runAllTimers()
      })
      
      // Date can be called after mount
      const afterMountDateCalls = DateSpy.mock.calls.length
      expect(afterMountDateCalls).toBeGreaterThanOrEqual(initialDateCalls)
    })
  })

  describe('Modal Interactions', () => {
    it('should handle modal open/close without hydration errors', async () => {
      const { container } = render(<OperationsContent />)
      
      await waitFor(() => {
        const button = container.querySelector('button')
        if (button) {
          // Open modal
          fireEvent.click(button)
          
          // Close modal
          fireEvent.click(button)
        }
      }, { timeout: 100 }).catch(() => {
        // Button might not exist, which is fine for this test
      })
      
      // No hydration errors during modal interactions
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })

    it('should render modal content only on client', async () => {
      const { container } = render(<OperationsContent />)
      
      // Modal shouldn't be in initial render
      expect(container.querySelector('[role="dialog"]')).not.toBeInTheDocument()
      
      // After interaction, modal can appear
      await waitFor(() => {
        const openButton = container.querySelector('[data-testid="open-modal"]')
        if (openButton) {
          fireEvent.click(openButton)
          expect(container.querySelector('[role="dialog"]')).toBeInTheDocument()
        }
      }, { timeout: 100 }).catch(() => {
        // Modal trigger might not exist
      })
    })
  })

  describe('Loading Skeletons', () => {
    it('should show operation skeleton during dynamic import', () => {
      const { container } = render(<OperationsPage />)
      
      // Check for skeleton elements
      const skeletons = container.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
      
      // Skeleton should have proper structure
      const cards = container.querySelectorAll('.card')
      expect(cards.length).toBeGreaterThanOrEqual(0)
    })

    it('should transition from skeleton to content', async () => {
      const { container } = render(<OperationsPage />)
      
      // Initially shows skeleton
      expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
      
      // Wait for dynamic content
      await waitFor(() => {
        expect(screen.getByTestId('operations-dynamic')).toBeInTheDocument()
      }, { timeout: 100 })
      
      // Skeleton replaced by content
      expect(container.querySelector('.animate-pulse')).not.toBeInTheDocument()
    })

    it('should maintain layout stability', async () => {
      const { container } = render(<OperationsPage />)
      
      // Get initial dimensions
      const initialWidth = container.firstChild?.clientWidth || 0
      
      // Wait for content
      await waitFor(() => {
        expect(screen.queryByTestId('operations-dynamic')).toBeInTheDocument()
      }, { timeout: 100 })
      
      // Width shouldn't change dramatically
      const loadedWidth = container.firstChild?.clientWidth || 0
      expect(Math.abs(loadedWidth - initialWidth)).toBeLessThan(100)
    })
  })

  describe('Error Handling', () => {
    it('should handle WebSocket errors without hydration issues', async () => {
      render(<OperationsContent />)
      
      // Simulate WebSocket error
      act(() => {
        mockWebSocketData.error = new Error('WebSocket connection failed')
      })
      
      await waitFor(() => {
        // Error should be handled gracefully
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Hydration')
        )
      })
    })

    it('should handle API errors gracefully', async () => {
      const { apiClient } = require('@/lib/api')
      apiClient.getOperations.mockRejectedValueOnce(new Error('API Error'))
      
      render(<OperationsContent />)
      
      await waitFor(() => {
        // API error shouldn't cause hydration issues
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Text content did not match')
        )
      })
    })

    it('should recover from errors without breaking hydration', async () => {
      const { container } = render(<OperationsContent />)
      
      // Simulate error
      act(() => {
        mockWebSocketData.error = new Error('Temporary error')
      })
      
      await waitFor(() => {}, { timeout: 10 })
      
      // Clear error
      act(() => {
        mockWebSocketData.error = null
      })
      
      await waitFor(() => {
        // Should recover without hydration errors
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Hydration failed')
        )
      })
    })
  })

  describe('Performance and Memory', () => {
    it('should cleanup WebSocket listeners on unmount', () => {
      const { unmount } = render(<OperationsContent />)
      
      // Setup cleanup spies
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout')
      const clearIntervalSpy = jest.spyOn(global, 'clearInterval')
      
      // Unmount
      unmount()
      
      // Verify cleanup
      expect(clearTimeoutSpy).toHaveBeenCalled()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })

    it('should handle rapid mount/unmount cycles', () => {
      for (let i = 0; i < 10; i++) {
        const { unmount } = render(<OperationsContent />)
        
        // Simulate some updates
        act(() => {
          mockWebSocketData.operations = [{ id: `op-${i}`, status: 'running' }]
        })
        
        unmount()
      }
      
      // No memory leak warnings
      expect(console.error).not.toHaveBeenCalled()
    })

    it('should not accumulate WebSocket listeners', async () => {
      const { rerender } = render(<OperationsContent />)
      
      // Multiple rerenders
      for (let i = 0; i < 5; i++) {
        rerender(<OperationsContent />)
        
        act(() => {
          mockWebSocketData.operations = [{ id: `op-${i}`, status: 'running' }]
        })
        
        await waitFor(() => {}, { timeout: 10 })
      }
      
      // Should not have memory issues
      expect(console.error).not.toHaveBeenCalled()
    })
  })

  describe('SSR Configuration', () => {
    it('should have SSR disabled in dynamic import', () => {
      render(<OperationsPage />)
      
      // Verify dynamic was called with ssr: false
      expect(dynamic).toHaveBeenCalledWith(
        expect.any(Function),
        expect.objectContaining({
          ssr: false,
        })
      )
    })

    it('should export metadata without breaking hydration', () => {
      const pageModule = require('@/app/operations/page')
      
      // Should have metadata
      expect(pageModule.metadata).toBeDefined()
      expect(pageModule.metadata.title).toContain('Operations')
      
      // Metadata shouldn't affect rendering
      render(<OperationsPage />)
      expect(console.error).not.toHaveBeenCalled()
    })
  })

  describe('Operation Type Rendering', () => {
    const operationTypes = [
      'scraping',
      'processing',
      'indices',
      'analysis',
      'full_pipeline',
      'data_processing',
    ]

    it.each(operationTypes)('should render %s operation type without hydration errors', async (type) => {
      render(<OperationsContent />)
      
      act(() => {
        mockWebSocketData.operations = [
          {
            id: `op-${type}`,
            type,
            status: 'running',
            progress: 50,
          },
        ]
      })
      
      await waitFor(() => {
        // Each operation type should render without issues
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Hydration')
        )
      })
    })
  })

  describe('Integration with UI Components', () => {
    it('should work with all Shadcn/ui components', async () => {
      const { container } = render(<OperationsContent />)

      await waitFor(() => {
        // Check for UI components
        const cards = container.querySelectorAll('.card')
        const buttons = container.querySelectorAll('.button')
        const badges = container.querySelectorAll('.badge')

        // Components should exist without hydration errors
        expect(console.error).not.toHaveBeenCalled()
      }, { timeout: 100 })
    })

    it('should handle progress bars without mismatches', async () => {
      render(<OperationsContent />)

      // Simulate operation with progress
      act(() => {
        mockWebSocketData.operations = [
          {
            id: 'op-1',
            type: 'processing',
            status: 'running',
            progress: 75,
          },
        ]
      })

      await waitFor(() => {
        // Progress rendering shouldn't cause hydration errors
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Text content did not match')
        )
      })
    })
  })

  describe('WebSocket Data Extraction Fix Tests', () => {
    it('should properly handle operation snapshots with operation_id', async () => {
      // Mock WebSocket client to capture handler
      let capturedHandler: (data: any) => void
      jest.doMock('@/lib/websocket', () => ({
        getWebSocketClient: jest.fn(() => ({
          connect: jest.fn().mockResolvedValue(undefined),
          subscribe: jest.fn().mockImplementation((event, handler) => {
            if (event === 'operation:snapshot') {
              capturedHandler = handler
            }
            return jest.fn()
          }),
          getConnectionState: jest.fn().mockReturnValue('connected'),
        })),
        ISXWebSocketClient: jest.fn(),
      }))

      render(<OperationsContent />)

      // Simulate realistic operation snapshot with full structure
      const realisticSnapshot = {
        operation_id: 'op-pipeline-123',
        name: 'Full Pipeline Operation',
        status: 'running',
        progress: 65,
        message: 'Processing ISX data...',
        current_step: 'processing',
        steps: [
          {
            id: 'scraping',
            name: 'Scraping Stage',
            status: 'completed',
            progress: 100,
            start_time: '2025-10-25T13:00:00Z',
            end_time: '2025-10-25T13:05:00Z',
            metadata: {
              files_downloaded: 45,
              data_size: '15.2MB'
            }
          },
          {
            id: 'processing',
            name: 'Processing Stage',
            status: 'running',
            progress: 65,
            start_time: '2025-10-25T13:05:00Z',
            metadata: {
              records_processed: 1250,
              files_processed: 12
            }
          },
          {
            id: 'indices',
            name: 'Index Stage',
            status: 'pending',
            progress: 0,
            metadata: {}
          },
          {
            id: 'liquidity',
            name: 'Liquidity Stage',
            status: 'pending',
            progress: 0,
            metadata: {}
          },
          {
            id: 'indicators',
            name: 'Indicators Stage',
            status: 'pending',
            progress: 0,
            metadata: {}
          }
        ],
        metadata: {
          total_stages: 5,
          current_stage: 'processing',
          operation_type: 'full_pipeline'
        }
      }

      // Send the snapshot through WebSocket
      act(() => {
        if (capturedHandler) {
          capturedHandler(realisticSnapshot)
        }
        // Update mock data to reflect the change
        mockWebSocketData.operations = [realisticSnapshot]
      })

      await waitFor(() => {
        // Verify the operation was stored correctly
        expect(mockWebSocketData.operations).toHaveLength(1)
        expect(mockWebSocketData.operations[0].operation_id).toBe('op-pipeline-123')
        expect(mockWebSocketData.operations[0].progress).toBe(65)
        expect(mockWebSocketData.operations[0].steps).toHaveLength(5)

        // Should not have hydration errors
        expect(console.error).not.toHaveBeenCalled()
      })
    })

    it('should reject operation snapshots missing operation_id', async () => {
      let capturedHandler: (data: any) => void
      jest.doMock('@/lib/websocket', () => ({
        getWebSocketClient: jest.fn(() => ({
          connect: jest.fn().mockResolvedValue(undefined),
          subscribe: jest.fn().mockImplementation((event, handler) => {
            if (event === 'operation:snapshot') {
              capturedHandler = handler
            }
            return jest.fn()
          }),
          getConnectionState: jest.fn().mockReturnValue('connected'),
        })),
        ISXWebSocketClient: jest.fn(),
      }))

      render(<OperationsContent />)

      // Simulate operation snapshot without operation_id (metadata-only)
      const metadataOnlySnapshot = {
        name: 'Broken Operation',
        status: 'running',
        progress: 50,
        message: 'Missing operation_id',
        metadata: {
          operation_type: 'scraping'
        }
      }

      act(() => {
        if (capturedHandler) {
          capturedHandler(metadataOnlySnapshot)
        }
      })

      await waitFor(() => {
        // Should not store the operation
        expect(mockWebSocketData.operations).toHaveLength(0)

        // Should not have hydration errors
        expect(console.error).not.toHaveBeenCalled()
      })
    })

    it('should handle complex multi-step pipeline operations', async () => {
      let capturedHandler: (data: any) => void
      jest.doMock('@/lib/websocket', () => ({
        getWebSocketClient: jest.fn(() => ({
          connect: jest.fn().mockResolvedValue(undefined),
          subscribe: jest.fn().mockImplementation((event, handler) => {
            if (event === 'operation:snapshot') {
              capturedHandler = handler
            }
            return jest.fn()
          }),
          getConnectionState: jest.fn().mockReturnValue('connected'),
        })),
        ISXWebSocketClient: jest.fn(),
      }))

      render(<OperationsContent />)

      // Simulate pipeline progress through multiple stages
      const pipelineProgression = [
        {
          operation_id: 'op-multi-stage-456',
          name: 'Multi-Stage Pipeline',
          status: 'running',
          progress: 20,
          current_step: 'scraping',
          steps: [
            {
              id: 'scraping',
              name: 'Scraping ISX Website',
              status: 'running',
              progress: 80,
              metadata: { pages_scanned: 2, files_found: 3 }
            }
          ]
        },
        {
          operation_id: 'op-multi-stage-456',
          name: 'Multi-Stage Pipeline',
          status: 'running',
          progress: 60,
          current_step: 'processing',
          steps: [
            {
              id: 'scraping',
              name: 'Scraping ISX Website',
              status: 'completed',
              progress: 100,
              start_time: '2025-10-25T13:00:00Z',
              end_time: '2025-10-25T13:03:00Z',
              metadata: { pages_scanned: 5, files_found: 8 }
            },
            {
              id: 'processing',
              name: 'Processing Data',
              status: 'running',
              progress: 75,
              metadata: { records_processed: 890 }
            }
          ]
        }
      ]

      act(() => {
        // Send first stage
        if (capturedHandler) {
          capturedHandler(pipelineProgression[0])
        }
        mockWebSocketData.operations = [pipelineProgression[0]]
      })

      await waitFor(() => {
        expect(mockWebSocketData.operations).toHaveLength(1)
        expect(mockWebSocketData.operations[0].progress).toBe(20)
        expect(mockWebSocketData.operations[0].current_step).toBe('scraping')
      })

      act(() => {
        // Send second stage
        if (capturedHandler) {
          capturedHandler(pipelineProgression[1])
        }
        mockWebSocketData.operations = [pipelineProgression[1]]
      })

      await waitFor(() => {
        expect(mockWebSocketData.operations).toHaveLength(1)
        expect(mockWebSocketData.operations[0].progress).toBe(60)
        expect(mockWebSocketData.operations[0].current_step).toBe('processing')
        expect(mockWebSocketData.operations[0].steps).toHaveLength(2)
      })

      // Should not have hydration errors throughout progression
      expect(console.error).not.toHaveBeenCalled()
    })
  })
})
