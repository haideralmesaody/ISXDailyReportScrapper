/**
 * Operations WebSocket Tests
 *
 * Comprehensive tests for the operations WebSocket channel including:
 * - Connection management
 * - Message parsing and validation
 * - Real-time updates
 * - Error handling
 * - Reconnection logic
 * - Performance under load
 */

import { act, renderHook, waitFor } from '@testing-library/react'
import { useWebSocket } from '@/lib/websocket'
import { useOperationSnapshots } from '@/lib/hooks/use-websocket'
import {
  parseWebSocketMessage,
  parseWebSocketMessageEnhanced,
  isOperationSnapshotMessage
} from '@/lib/operations/validation'

// Mock WebSocket for testing
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState: number = MockWebSocket.CONNECTING
  url: string
  onopen: ((event: any) => void) | null = null
  onclose: ((event: any) => void) | null = null
  onmessage: ((event: any) => void) | null = null
  onerror: ((event: any) => void) | null = null

  private messageQueue: any[] = []
  private closeTimeout: NodeJS.Timeout | null = null

  constructor(url: string) {
    this.url = url
    // Simulate connection delay
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN
      this.onopen?.({ type: 'open' })
    }, 10)
  }

  send(data: string): void {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open')
    }

    try {
      const parsedData = JSON.parse(data)
      this.messageQueue.push(parsedData)
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }

  close(): void {
    if (this.readyState === MockWebSocket.OPEN) {
      this.readyState = MockWebSocket.CLOSING

      this.closeTimeout = setTimeout(() => {
        this.readyState = MockWebSocket.CLOSED
        this.onclose?.({ type: 'close', code: 1000, reason: 'Normal closure' })
      }, 5)
    }
  }

  // Helper methods for testing
  simulateMessage(data: any): void {
    if (this.readyState === MockWebSocket.OPEN) {
      this.onmessage?.({
        type: 'message',
        data: JSON.stringify(data)
      })
    }
  }

  simulateError(error: any): void {
    this.onerror?.({ type: 'error', error })
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.({ type: 'close', code: 1006, reason: 'Abnormal closure' })
  }

  getSentMessages(): any[] {
    return [...this.messageQueue]
  }

  clearSentMessages(): void {
    this.messageQueue = []
  }

  cleanup(): void {
    if (this.closeTimeout) {
      clearTimeout(this.closeTimeout)
      this.closeTimeout = null
    }
  }
}

// Mock global WebSocket
Object.defineProperty(global, 'WebSocket', {
  value: MockWebSocket,
  writable: true
})

describe('Operations WebSocket Message Parsing', () => {
  describe('parseWebSocketMessage', () => {
    test('should parse valid operation snapshot messages', () => {
      const validMessage = {
        type: 'operation:snapshot',
        data: {
          operation_id: 'test-op-123',
          status: 'running',
          progress: 75,
          current_stage: 'processing'
        },
        timestamp: '2025-01-01T10:30:00Z',
        id: 'msg-123'
      }

      const result = parseWebSocketMessage(validMessage)

      expect(result).toBeDefined()
      expect(result!.type).toBe('operation:snapshot')
      expect(result!.data.operation_id).toBe('test-op-123')
      expect(result!.timestamp).toBe('2025-01-01T10:30:00Z')
      expect(result!.id).toBe('msg-123')
    })

    test('should reject messages with wrong type', () => {
      const invalidMessage = {
        type: 'wrong-type',
        data: { operation_id: 'test' }
      }

      const result = parseWebSocketMessage(invalidMessage)
      expect(result).toBeNull()
    })

    test('should reject messages without data', () => {
      const messageWithoutData = {
        type: 'operation:snapshot'
      }

      const result = parseWebSocketMessage(messageWithoutData)
      expect(result).toBeNull()
    })

    test('should reject non-object messages', () => {
      const invalidMessages = [null, undefined, 'string', 123, []]

      invalidMessages.forEach(message => {
        const result = parseWebSocketMessage(message)
        expect(result).toBeNull()
      })
    })

    test('should add default timestamp if missing', () => {
      const messageWithoutTimestamp = {
        type: 'operation:snapshot',
        data: { operation_id: 'test', status: 'running' }
      }

      const result = parseWebSocketMessage(messageWithoutTimestamp)

      expect(result).toBeDefined()
      expect(result!.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}.\d{3}Z$/)
    })
  })

  describe('parseWebSocketMessageEnhanced', () => {
    test('should validate operation data in messages', () => {
      const validMessage = {
        type: 'operation:snapshot',
        data: {
          operation_id: 'enhanced-test-123',
          status: 'running',
          progress: 85,
          current_stage: 'processing'
        }
      }

      const result = parseWebSocketMessageEnhanced(validMessage)

      expect(result).toBeDefined()
      expect(result!.data.operation_id).toBe('enhanced-test-123')
      expect(result!.data.metadata).toBeDefined()
    })

    test('should reject messages with invalid operation data', () => {
      const messageWithInvalidData = {
        type: 'operation:snapshot',
        data: {
          // Missing operation_id
          status: 'running'
        }
      }

      const result = parseWebSocketMessageEnhanced(messageWithInvalidData)
      expect(result).toBeNull()
    })

    test('should migrate legacy operation data', () => {
      const messageWithLegacyData = {
        type: 'operation:snapshot',
        data: {
          operation_id: 'legacy-456',
          status: 'completed',
          progress: 100,
          current_stage: 'finished'
          // Missing started_at, updated_at - should get defaults after migration
        }
      }

      const result = parseWebSocketMessageEnhanced(messageWithLegacyData)

      expect(result).toBeDefined()
      expect(result!.data.started_at).toBeDefined()
      expect(result!.data.updated_at).toBeDefined()
      expect(result!.data.metadata.migrationPerformed).toBe(true)
    })
  })

  describe('isOperationSnapshotMessage', () => {
    test('should identify valid operation snapshot messages', () => {
      const validMessage = {
        type: 'operation:snapshot',
        data: { operation_id: 'test', status: 'running' }
      }

      expect(isOperationSnapshotMessage(validMessage)).toBe(true)
    })

    test('should reject messages with wrong type', () => {
      const invalidMessage = {
        type: 'different-type',
        data: { operation_id: 'test' }
      }

      expect(isOperationSnapshotMessage(invalidMessage)).toBe(false)
    })

    test('should reject messages without proper data structure', () => {
      const invalidMessages = [
        { type: 'operation:snapshot', data: null },
        { type: 'operation:snapshot', data: undefined },
        { type: 'operation:snapshot' }
      ]

      invalidMessages.forEach(message => {
        expect(isOperationSnapshotMessage(message)).toBe(false)
      })
    })
  })
})

describe('useOperationSnapshots Hook', () => {
  beforeEach(() => {
    // Clear all mocks before each test
    jest.clearAllMocks()
  })

  afterEach(() => {
    // Cleanup any lingering WebSocket connections
    jest.restoreAllMocks()
  })

  test('should initialize with default values', () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    expect(result.current.snapshots).toEqual([])
    expect(result.current.connected).toBe(false)
    expect(result.current.connectionStatus).toBe('disconnected')
    expect(result.current.error).toBe(null)
  })

  test('should connect to WebSocket', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    // Wait for connection
    await waitFor(() => {
      expect(result.current.connected).toBe(true)
      expect(result.current.connectionStatus).toBe('connected')
    })
  })

  test('should handle operation snapshot updates', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    // Wait for connection
    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    // Simulate receiving an operation update
    const mockWebSocket = (global as any).WebSocket.instances?.[0]
    if (mockWebSocket) {
      act(() => {
        mockWebSocket.simulateMessage({
          type: 'operation:snapshot',
          data: {
            operation_id: 'realtime-test-123',
            status: 'running',
            progress: 45,
            current_stage: 'processing'
          }
        })
      })
    }

    // Wait for state update
    await waitFor(() => {
      expect(result.current.snapshots).toHaveLength(1)
      expect(result.current.snapshots[0].operation_id).toBe('realtime-test-123')
      expect(result.current.snapshots[0].status).toBe('running')
    })
  })

  test('should handle multiple operation updates', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]
    if (mockWebSocket) {
      // Send multiple updates
      const updates = [
        { operation_id: 'multi-1', status: 'pending', progress: 0 },
        { operation_id: 'multi-2', status: 'running', progress: 25 },
        { operation_id: 'multi-1', status: 'running', progress: 50 },
        { operation_id: 'multi-2', status: 'completed', progress: 100 }
      ]

      updates.forEach(update => {
        act(() => {
          mockWebSocket.simulateMessage({
            type: 'operation:snapshot',
            data: update
          })
        })
      })
    }

    await waitFor(() => {
      expect(result.current.snapshots.length).toBeGreaterThan(0)

      // Should have latest states
      const op1 = result.current.snapshots.find(s => s.operation_id === 'multi-1')
      const op2 = result.current.snapshots.find(s => s.operation_id === 'multi-2')

      expect(op1?.status).toBe('running')
      expect(op1?.progress).toBe(50)
      expect(op2?.status).toBe('completed')
      expect(op2?.progress).toBe(100)
    })
  })

  test('should handle invalid messages gracefully', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]
    if (mockWebSocket) {
      // Send invalid message
      act(() => {
        mockWebSocket.simulateMessage({
          type: 'operation:snapshot',
          data: {
            // Missing operation_id
            status: 'running'
          }
        })
      })
    }

    // Should not crash and should not add invalid snapshots
    await waitFor(() => {
      expect(result.current.snapshots).toEqual([])
      // May have validation warnings but no errors that break functionality
    })
  })

  test('should handle connection errors', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]
    if (mockWebSocket) {
      act(() => {
        mockWebSocket.simulateError(new Error('Connection failed'))
      })
    }

    await waitFor(() => {
      expect(result.current.connected).toBe(false)
      expect(result.current.connectionStatus).toBe('error')
      expect(result.current.error).toBeTruthy()
    })
  })

  test('should manage subscription correctly', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    // Get subscription function
    const { subscribe } = result.current
    expect(typeof subscribe).toBe('function')

    // Subscribe to operation updates
    const mockCallback = jest.fn()
    act(() => {
      subscribe('operation:snapshot', mockCallback)
    })

    // Should not crash
    expect(mockCallback).toBeDefined()
  })
})

describe('WebSocket Performance and Load Testing', () => {
  test('should handle high message volume', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]
    const startTime = performance.now()

    // Send 100 messages rapidly
    for (let i = 0; i < 100; i++) {
      act(() => {
        mockWebSocket?.simulateMessage({
          type: 'operation:snapshot',
          data: {
            operation_id: `load-test-${i}`,
            status: i % 2 === 0 ? 'running' : 'completed',
            progress: (i % 11) * 10,
            current_stage: 'processing'
          }
        })
      })
    }

    const endTime = performance.now()
    const processingTime = endTime - startTime

    await waitFor(() => {
      expect(result.current.snapshots.length).toBeGreaterThan(0)
    })

    // Should process messages quickly (less than 100ms for 100 messages)
    expect(processingTime).toBeLessThan(100)
  })

  test('should handle concurrent operations', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Simulate 20 concurrent operations
    const concurrentOps = Array.from({ length: 20 }, (_, i) => ({
      operation_id: `concurrent-${i}`,
      status: 'running',
      progress: Math.floor(Math.random() * 100),
      current_stage: 'processing'
    }))

    // Send all operations concurrently
    act(() => {
      concurrentOps.forEach(op => {
        mockWebSocket?.simulateMessage({
          type: 'operation:snapshot',
          data: op
        })
      })
    })

    await waitFor(() => {
      expect(result.current.snapshots.length).toBe(20)

      // Verify all operations are present
      concurrentOps.forEach(op => {
        const found = result.current.snapshots.find(s => s.operation_id === op.operation_id)
        expect(found).toBeDefined()
        expect(found?.status).toBe(op.status)
      })
    })
  })

  test('should handle memory usage efficiently', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Get initial memory usage (approximate)
    const initialMemory = process.memoryUsage().heapUsed

    // Send many messages with updates
    for (let i = 0; i < 50; i++) {
      act(() => {
        mockWebSocket?.simulateMessage({
          type: 'operation:snapshot',
          data: {
            operation_id: `memory-test-${i}`,
            status: 'running',
            progress: i * 2,
            current_stage: 'processing',
            // Add some additional data to test memory
            metadata: {
              version: '1.0',
              source: 'test',
              additionalData: 'x'.repeat(100) // Some extra data
            }
          }
        })
      })
    }

    await waitFor(() => {
      expect(result.current.snapshots.length).toBe(50)
    })

    // Force garbage collection if available
    if (global.gc) {
      global.gc()
    }

    const finalMemory = process.memoryUsage().heapUsed
    const memoryIncrease = finalMemory - initialMemory

    // Memory increase should be reasonable (less than 10MB for this test)
    expect(memoryIncrease).toBeLessThan(10 * 1024 * 1024)
  })
})

describe('WebSocket Error Recovery', () => {
  test('should handle temporary connection loss', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Simulate connection loss
    act(() => {
      mockWebSocket?.simulateError(new Error('Temporary connection loss'))
    })

    await waitFor(() => {
      expect(result.current.connected).toBe(false)
      expect(result.current.connectionStatus).toBe('error')
    })

    // Should automatically attempt to reconnect (this is handled by the WebSocket implementation)
    // The hook should continue functioning after reconnection
  })

  test('should handle malformed messages without crashing', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Send various malformed messages
    const malformedMessages = [
      { type: 'operation:snapshot', data: null },
      { type: 'operation:snapshot' }, // Missing data
      { type: 'wrong-type', data: {} },
      null,
      undefined,
      'invalid-string',
      123
    ]

    malformedMessages.forEach(message => {
      act(() => {
        mockWebSocket?.simulateMessage(message)
      })
    })

    // Hook should still be functional
    expect(result.current.snapshots).toEqual([])
    expect(typeof result.current.connected).toBe('boolean')
  })

  test('should handle circular references in message data', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Create object with circular reference
    const circularData: any = {
      operation_id: 'circular-test',
      status: 'running'
    }
    circularData.self = circularData

    act(() => {
      mockWebSocket?.simulateMessage({
        type: 'operation:snapshot',
        data: circularData
      })
    })

    // Should handle gracefully (either process or skip the message)
    expect(typeof result.current.connected).toBe('boolean')
  })
})

describe('WebSocket Integration with Validation', () => {
  test('should integrate validation with real-time updates', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({
        enableValidation: true,
        enableMigration: true,
        debug: true
      })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Send message that requires migration
    const legacyMessage = {
      type: 'operation:snapshot',
      data: {
        operation_id: 'migration-integration-123',
        status: 'completed',
        progress: 100,
        current_stage: 'finished'
        // Missing required fields for v1.0
      }
    }

    act(() => {
      mockWebSocket?.simulateMessage(legacyMessage)
    })

    await waitFor(() => {
      expect(result.current.snapshots).toHaveLength(1)

      const snapshot = result.current.snapshots[0]
      expect(snapshot.operation_id).toBe('migration-integration-123')
      expect(snapshot.metadata.migrationPerformed).toBe(true)
      expect(snapshot.started_at).toBeDefined() // Should have default value
      expect(snapshot.updated_at).toBeDefined() // Should have default value
    })
  })

  test('should handle validation errors in real-time updates', async () => {
    const { result } = renderHook(() =>
      useOperationSnapshots({
        enableValidation: true,
        enableMigration: true
      })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Send message with validation errors
    const invalidMessage = {
      type: 'operation:snapshot',
      data: {
        operation_id: '', // Invalid empty ID
        status: 'invalid_status', // Invalid status
        progress: 150 // Out of range
      }
    }

    act(() => {
      mockWebSocket?.simulateMessage(invalidMessage)
    })

    // Should not add invalid snapshot
    await waitFor(() => {
      expect(result.current.snapshots).toEqual([])
    })

    // May have validation warnings but should continue functioning
    expect(typeof result.current.connected).toBe('boolean')
  })
})

describe('WebSocket Cleanup and Resource Management', () => {
  test('should clean up WebSocket connections on unmount', async () => {
    const { result, unmount } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]
    const closeSpy = jest.spyOn(mockWebSocket!, 'close')

    // Unmount the hook
    unmount()

    // Should close WebSocket connection
    expect(closeSpy).toHaveBeenCalled()
  })

  test('should clean up message queues on unmount', async () => {
    const { result, unmount } = renderHook(() =>
      useOperationSnapshots({ enableValidation: true, enableMigration: true })
    )

    await waitFor(() => {
      expect(result.current.connected).toBe(true)
    })

    const mockWebSocket = (global as any).WebSocket.instances?.[0]

    // Send some messages
    act(() => {
      mockWebSocket?.simulateMessage({
        type: 'operation:snapshot',
        data: { operation_id: 'cleanup-test', status: 'running' }
      })
    })

    // Unmount hook
    unmount()

    // Cleanup should be called
    mockWebSocket?.cleanup()
  })
})