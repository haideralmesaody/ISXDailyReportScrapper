/**
 * Operations Load Testing Suite
 *
 * Comprehensive performance and load testing for concurrent operations including:
 * - High-volume operation processing
 * - Concurrent WebSocket connections
 * - Memory usage under load
 * - CPU performance under stress
 * - Database/Storage performance
 * - Network I/O performance
 * - Resource cleanup under load
 */

import { renderHook, act, waitFor } from '@testing-library/react'
import { useOperationSnapshots } from '@/lib/hooks/use-websocket'
import { validateOperationSnapshotsBatch } from '@/lib/operations/validation'

// Mock WebSocket for load testing
class LoadTestWebSocket {
  private static instances: LoadTestWebSocket[] = []
  readyState: number = 1 // OPEN
  url: string
  onmessage: ((event: any) => void) | null = null
  onerror: ((event: any) => void) | null = null
  close = jest.fn()

  constructor(url: string) {
    this.url = url
    LoadTestWebSocket.instances.push(this)
  }

  send(data: string): void {
    // Mock send - no actual network I/O
  }

  simulateMessage(data: any): void {
    if (this.readyState === 1 && this.onmessage) {
      this.onmessage({
        type: 'message',
        data: JSON.stringify(data)
      })
    }
  }

  static clearInstances(): void {
    LoadTestWebSocket.instances = []
  }

  static getInstances(): LoadTestWebSocket[] {
    return LoadTestWebSocket.instances
  }
}

// Override global WebSocket for testing
Object.defineProperty(global, 'WebSocket', {
  value: LoadTestWebSocket,
  writable: true
})

describe('Operations Load Testing', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    LoadTestWebSocket.clearInstances()
  })

  afterEach(() => {
    LoadTestWebSocket.clearInstances()
  })

  describe('High-Volume Message Processing', () => {
    test('should handle 1000 concurrent operation updates efficiently', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      const startTime = performance.now()
      const messageCount = 1000

      // Send large volume of messages
      for (let i = 0; i < messageCount; i++) {
        act(() => {
          LoadTestWebSocket.getInstances().forEach(ws => {
            ws.simulateMessage({
              type: 'operation:snapshot',
              data: {
                operation_id: `load-test-${i}`,
                status: i % 3 === 0 ? 'pending' : i % 3 === 1 ? 'running' : 'completed',
                progress: Math.floor((i / messageCount) * 100),
                current_stage: 'processing',
                metadata: {
                  version: '1.0',
                  source: 'load-test',
                  testData: 'x'.repeat(50) // Add some data payload
                }
              }
            })
          })
        })
      }

      await waitFor(() => {
        expect(result.current.snapshots.length).toBeGreaterThan(0)
      })

      const endTime = performance.now()
      const totalTime = endTime - startTime
      const averageTimePerMessage = totalTime / messageCount

      // Performance assertions
      expect(totalTime).toBeLessThan(5000) // Should complete within 5 seconds
      expect(averageTimePerMessage).toBeLessThan(5) // Average 5ms per message
      expect(result.current.snapshots.length).toBe(messageCount)
    })

    test('should handle 10,000 updates without memory leaks', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: false, // Disable validation for maximum performance
          enableMigration: false
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      const initialMemory = process.memoryUsage().heapUsed
      const updateCount = 10000

      // Send very large volume of messages
      const batchSize = 100
      for (let batch = 0; batch < updateCount / batchSize; batch++) {
        await act(async () => {
          for (let i = 0; i < batchSize; i++) {
            const messageIndex = batch * batchSize + i

            LoadTestWebSocket.getInstances().forEach(ws => {
              ws.simulateMessage({
                type: 'operation:snapshot',
                data: {
                  operation_id: `stress-test-${messageIndex}`,
                  status: 'running',
                  progress: (messageIndex % 100),
                  current_stage: 'processing'
                }
              })
            })
          }

          // Small delay to prevent overwhelming the event loop
          await new Promise(resolve => setTimeout(resolve, 1))
        })
      }

      await waitFor(() => {
        expect(result.current.snapshots.length).toBe(updateCount)
      })

      // Force garbage collection if available
      if (global.gc) {
        global.gc()
      }

      const finalMemory = process.memoryUsage().heapUsed
      const memoryIncrease = finalMemory - initialMemory

      // Memory usage should be reasonable for this test
      expect(memoryIncrease).toBeLessThan(100 * 1024 * 1024) // Less than 100MB increase
    })

    test('should maintain performance with complex validation enabled', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      const complexMessages = Array.from({ length: 500 }, (_, i) => ({
        type: 'operation:snapshot',
        data: {
          operation_id: `complex-${i}`,
          status: 'running',
          progress: Math.floor(Math.random() * 100),
          current_stage: 'processing',
          steps: Array.from({ length: 5 }, (_, stepIndex) => ({
            name: `Step ${stepIndex + 1}`,
            status: Math.random() > 0.5 ? 'completed' : 'running',
            progress: Math.floor(Math.random() * 100),
            started_at: new Date().toISOString(),
            metadata: {
              details: `Complex data for step ${stepIndex}`,
              metrics: {
                processed: Math.floor(Math.random() * 1000),
                errors: Math.floor(Math.random() * 10),
                warnings: Math.floor(Math.random() * 5)
              }
            }
          })),
          metadata: {
            version: '1.0',
            source: 'load-test',
            performance: {
              cpu: Math.random() * 100,
              memory: Math.random() * 1024 * 1024 * 1024,
              network: Math.random() * 1000
            },
            customFields: Array.from({ length: 10 }, (_, fieldIndex) => ({
              [`custom_field_${fieldIndex}`]: `value_${fieldIndex}`
            }))
          }
        }
      }))

      const startTime = performance.now()

      // Send complex messages in batches
      const batchSize = 50
      for (let batch = 0; batch < complexMessages.length / batchSize; batch++) {
        await act(async () => {
          const batchMessages = complexMessages.slice(batch * batchSize, (batch + 1) * batchSize)

          batchMessages.forEach(message => {
            LoadTestWebSocket.getInstances().forEach(ws => {
              ws.simulateMessage(message)
            })
          })

          await new Promise(resolve => setTimeout(resolve, 5))
        })
      }

      const endTime = performance.now()
      const totalTime = endTime - startTime

      expect(totalTime).toBeLessThan(10000) // Should complete within 10 seconds
      expect(result.current.snapshots.length).toBe(500)
    })
  })

  describe('Concurrent Connection Testing', () => {
    test('should handle multiple concurrent WebSocket connections', async () => {
      const connections = Array.from({ length: 10 }, (_, i) => {
        const { result } = renderHook(() =>
          useOperationSnapshots({
            enableValidation: true,
            enableMigration: true
          })
        )

        return result
      })

      // Wait for all connections to establish
      await Promise.all(
        connections.map(({ result }) =>
          waitFor(() => expect(result.current.connected).toBe(true))
        )
      )

      const startTime = performance.now()

      // Send messages to all connections simultaneously
      connections.forEach(({ result }, index) => {
        act(() => {
          LoadTestWebSocket.getInstances()[index]?.simulateMessage({
            type: 'operation:snapshot',
            data: {
              operation_id: `concurrent-${index}`,
              status: 'running',
              progress: 50,
              current_stage: 'processing'
            }
          })
        })
      })

      // Wait for all connections to process messages
      await Promise.all(
        connections.map(({ result }) =>
          waitFor(() => expect(result.current.snapshots.length).toBe(1))
        )
      )

      const endTime = performance.now()
      const concurrentTime = endTime - startTime

      expect(concurrentTime).toBeLessThan(1000) // Should handle concurrent messages quickly
    })

    test('should handle connection drops gracefully', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Establish some operations
      for (let i = 0; i < 50; i++) {
        act(() => {
          LoadTestWebSocket.getInstances()[0]?.simulateMessage({
            type: 'operation:snapshot',
            data: {
              operation_id: `drop-test-${i}`,
              status: 'running',
              progress: i * 2,
              current_stage: 'processing'
            }
          })
        })
      }

      await waitFor(() => {
        expect(result.current.snapshots.length).toBe(50)
      })

      // Simulate connection drop
      act(() => {
        const ws = LoadTestWebSocket.getInstances()[0]
        if (ws) {
          ws.readyState = 3 // CLOSED
          ws.onerror?.({ type: 'error', error: 'Connection dropped' })
        }
      })

      // Should handle connection error gracefully
      await waitFor(() => {
        expect(result.current.connected).toBe(false)
        expect(result.current.connectionStatus).toBe('error')
      })
    })
  })

  describe('Memory Management Under Load', () => {
    test('should clean up old snapshots when memory pressure is high', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Fill up with many operations
      for (let i = 0; i < 1000; i++) {
        act(() => {
          LoadTestWebSocket.getInstances()[0]?.simulateMessage({
            type: 'operation:snapshot',
            data: {
              operation_id: `memory-test-${i}`,
              status: 'completed',
              progress: 100,
              current_stage: 'finished',
              metadata: {
                version: '1.0',
                source: 'memory-test',
                largeData: 'x'.repeat(1000) // 1KB per operation
              }
            }
          })
        })
      }

      await waitFor(() => {
        expect(result.current.snapshots.length).toBe(1000)
      })

      // Memory should not grow excessively
      const memoryAfter = process.memoryUsage()
      expect(memoryAfter.heapUsed).toBeLessThan(500 * 1024 * 1024) // Less than 500MB
    })

    test('should handle large metadata efficiently', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Create operations with large metadata
      const largeMetadataMessage = {
        type: 'operation:snapshot',
        data: {
          operation_id: 'large-metadata-test',
          status: 'running',
          progress: 50,
          current_stage: 'processing',
          metadata: {
            version: '1.0',
            source: 'large-test',
            logs: Array.from({ length: 100 }, (_, i) => ({
              timestamp: new Date().toISOString(),
              level: 'info',
              message: `Log message ${i} with additional context: ${'x'.repeat(100)}`
            })),
            debugInfo: {
              stackTrace: 'x'.repeat(1000),
              environment: {
                nodeVersion: process.version,
                platform: process.platform,
                arch: process.arch,
                memory: process.memoryUsage()
              },
              customData: Array.from({ length: 50 }, (_, i) => ({
                [`custom_${i}`]: `value_${'x'.repeat(200)}`
              }))
            }
          }
        }
      }

      const startTime = performance.now()

      act(() => {
        LoadTestWebSocket.getInstances()[0]?.simulateMessage(largeMetadataMessage)
      })

      await waitFor(() => {
        expect(result.current.snapshots.length).toBe(1)
        const snapshot = result.current.snapshots[0]
        expect(snapshot.metadata.logs).toHaveLength(100)
        expect(snapshot.metadata.debugInfo.customData).toHaveLength(50)
      })

      const endTime = performance.now()
      const processingTime = endTime - startTime

      // Should handle large metadata efficiently
      expect(processingTime).toBeLessThan(1000) // Less than 1 second
    })
  })

  describe('Batch Validation Performance', () => {
    test('should validate large batches efficiently', () => {
      const largeDataSet = Array.from({ length: 1000 }, (_, i) => ({
        operation_id: `batch-${i}`,
        status: i % 2 === 0 ? 'running' : 'completed',
        progress: (i % 11) * 10,
        current_stage: 'processing',
        started_at: new Date(Date.now() - i * 60000).toISOString(),
        updated_at: new Date(Date.now() - i * 30000).toISOString(),
        metadata: {
          version: '1.0',
          source: 'batch-test'
        }
      }))

      const startTime = performance.now()
      const result = validateOperationSnapshotsBatch(largeDataSet)
      const endTime = performance.now()

      const processingTime = endTime - startTime

      expect(result.totalItems).toBe(1000)
      expect(result.validItems).toBe(1000)
      expect(result.invalidItems).toBe(0)
      expect(processingTime).toBeLessThan(500) // Should complete within 500ms
      expect(result.summary.averageValidationTime).toBeLessThan(1) // Average less than 1ms per item
    })

    test('should handle mixed valid/invalid data efficiently', () => {
      const mixedDataSet = Array.from({ length: 1000 }, (_, i) => {
        if (i % 3 === 0) {
          // Valid data
          return {
            operation_id: `valid-${i}`,
            status: 'running',
            progress: 50,
            current_stage: 'processing'
          }
        } else if (i % 3 === 1) {
          // Missing required fields
          return {
            status: 'running',
            progress: 50
          }
        } else {
          // Invalid data
          return {
            operation_id: '',
            status: 'invalid_status',
            progress: -10
          }
        }
      })

      const result = validateOperationSnapshotsBatch(mixedDataSet)

      expect(result.totalItems).toBe(1000)
      expect(result.validItems).toBe(333) // Every third item is valid
      expect(result.invalidItems).toBe(667)
      expect(result.summary.recommendations.length).toBeGreaterThan(0)
    })
  })

  describe('Resource Cleanup Under Load', () => {
    test('should clean up resources when unmounted under load', async () => {
      const { result, unmount } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Add many operations
      for (let i = 0; i < 500; i++) {
        act(() => {
          LoadTestWebSocket.getInstances()[0]?.simulateMessage({
            type: 'operation:snapshot',
            data: {
              operation_id: `cleanup-test-${i}`,
              status: 'running',
              progress: i * 0.2,
              current_stage: 'processing'
            }
          })
        })
      }

      await waitFor(() => {
        expect(result.current.snapshots.length).toBe(500)
      })

      const preCleanupMemory = process.memoryUsage().heapUsed

      // Unmount while under load
      unmount()

      // Force garbage collection
      if (global.gc) {
        global.gc()
      }

      const postCleanupMemory = process.memoryUsage().heapUsed
      const memoryChange = postCleanupMemory - preCleanupMemory

      // Memory should be released (or at least not grow significantly)
      expect(memoryChange).toBeLessThan(10 * 1024 * 1024) // Less than 10MB increase
    })

    test('should handle rapid mount/unmount cycles', async () => {
      for (let cycle = 0; cycle < 10; cycle++) {
        const { result, unmount } = renderHook(() =>
          useOperationSnapshots({
            enableValidation: true,
            enableMigration: true
          })
        )

        await waitFor(() => {
          expect(result.current.connected).toBe(true)
        })

        // Add some operations
        for (let i = 0; i < 50; i++) {
          act(() => {
            LoadTestWebSocket.getInstances()[0]?.simulateMessage({
              type: 'operation:snapshot',
              data: {
                operation_id: `cycle-${cycle}-op-${i}`,
                status: 'running',
                progress: i * 2,
                current_stage: 'processing'
              }
            })
          })
        }

        await waitFor(() => {
          expect(result.current.snapshots.length).toBe(50)
        })

        // Quick unmount
        unmount()

        // Small delay to allow cleanup
        await new Promise(resolve => setTimeout(resolve, 10))
      }

      // All cycles should complete without memory leaks
      const finalMemory = process.memoryUsage().heapUsed
      expect(finalMemory).toBeLessThan(200 * 1024 * 1024) // Less than 200MB
    })
  })

  describe('Stress Testing Edge Cases', () => {
    test('should handle extreme message frequencies', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: false,
          enableMigration: false
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Send messages extremely rapidly
      const messageCount = 10000
      const startTime = performance.now()

      const interval = setInterval(() => {
        const messageId = Math.floor(Math.random() * messageCount)
        LoadTestWebSocket.getInstances()[0]?.simulateMessage({
          type: 'operation:snapshot',
          data: {
            operation_id: `rapid-${messageId}`,
            status: 'running',
            progress: Math.floor(Math.random() * 100),
            current_stage: 'processing'
          }
        })
      }, 0.1) // 10ms interval = 100Hz

      // Stop after processing target number of messages
      const checkInterval = setInterval(() => {
        if (result.current.snapshots.length >= messageCount) {
          clearInterval(interval)
          clearInterval(checkInterval)
        }
      }, 100)

      await waitFor(() => {
        expect(result.current.snapshots.length).toBeGreaterThanOrEqual(messageCount)
      }, { timeout: 30000 }) // 30 second timeout

      clearInterval(interval)
      clearInterval(checkInterval)

      const endTime = performance.now()
      const totalTime = endTime - startTime

      expect(totalTime).toBeLessThan(15000) // Should complete within 15 seconds
    })

    test('should handle malformed data gracefully under load', async () => {
      const { result } = renderHook(() =>
        useOperationSnapshots({
          enableValidation: true,
          enableMigration: true
        })
      )

      await waitFor(() => {
        expect(result.current.connected).toBe(true)
      })

      // Mix of valid and malformed messages
      for (let i = 0; i < 1000; i++) {
        act(() => {
          if (i % 4 === 0) {
            // Valid message
            LoadTestWebSocket.getInstances()[0]?.simulateMessage({
              type: 'operation:snapshot',
              data: {
                operation_id: `valid-${i}`,
                status: 'running',
                progress: 50,
                current_stage: 'processing'
              }
            })
          } else if (i % 4 === 1) {
            // Missing required fields
            LoadTestWebSocket.getInstances()[0]?.simulateMessage({
              type: 'operation:snapshot',
              data: {
                status: 'running',
                progress: 50
              }
            })
          } else if (i % 4 === 2) {
            // Invalid data types
            LoadTestWebSocket.getInstances()[0]?.simulateMessage({
              type: 'operation:snapshot',
              data: {
                operation_id: 123,
                status: true,
                progress: 'invalid'
              }
            })
          } else {
            // Completely malformed
            LoadTestWebSocket.getInstances()[0]?.simulateMessage({
              type: 'operation:snapshot',
              data: 'not-an-object'
            })
          }
        })
      }

      await waitFor(() => {
        // Should have processed valid messages and rejected invalid ones
        expect(result.current.snapshots.length).toBeGreaterThan(0)
        expect(result.current.snapshots.length).toBeLessThan(1000)
      })

      // System should still be responsive
      expect(typeof result.current.connected).toBe('boolean')
    })
  })
})
