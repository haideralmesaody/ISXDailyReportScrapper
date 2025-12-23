import { ISXWebSocketClient, ConnectionState } from '@/lib/websocket'

// Mock WebSocket
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  public readyState: number = MockWebSocket.CONNECTING
  public onopen: ((event: Event) => void) | null = null
  public onclose: ((event: CloseEvent) => void) | null = null
  public onmessage: ((event: MessageEvent) => void) | null = null
  public onerror: ((event: Event) => void) | null = null

  constructor(public url: string) {
    // Simulate connection opening after a short delay
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN
      if (this.onopen) {
        this.onopen(new Event('open'))
      }
    }, 10)
  }

  send(data: string) {
    if (this.readyState !== MockWebSocket.OPEN) {
      throw new Error('WebSocket is not open')
    }
  }

  close(code?: number, reason?: string) {
    this.readyState = MockWebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close', { code: code || 1000, reason }))
    }
  }

  // Helper methods for testing
  simulateMessage(data: any) {
    if (this.onmessage) {
      this.onmessage(new MessageEvent('message', { data: JSON.stringify(data) }))
    }
  }

  simulateError() {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }

  simulateClose(code: number = 1000, reason: string = '') {
    this.readyState = MockWebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close', { code, reason }))
    }
  }
}

// Replace global WebSocket with mock
global.WebSocket = MockWebSocket as any

describe('ISXWebSocketClient', () => {
  let client: ISXWebSocketClient
  let mockWs: MockWebSocket

  beforeEach(() => {
    client = new ISXWebSocketClient()
    jest.clearAllMocks()
  })

  afterEach(() => {
    if (client) {
      client.disconnect()
    }
  })

  describe('Connection Management', () => {
    it('connects to WebSocket server with correct URL', () => {
      const connectPromise = client.connect()
      
      expect(client.getConnectionState()).toBe(ConnectionState.CONNECTING)
      
      // Get the mock WebSocket instance
      mockWs = (client as any).ws
      expect(mockWs.url).toBe('ws://localhost:8080/ws')

      return connectPromise.then(() => {
        expect(client.getConnectionState()).toBe(ConnectionState.CONNECTED)
      })
    })

    it('uses production WebSocket URL when not in development', () => {
      const originalEnv = process.env.NODE_ENV
      process.env.NODE_ENV = 'production'

      // Mock window.location
      Object.defineProperty(window, 'location', {
        value: { host: 'iraqiinvestor.gov.iq' },
        writable: true,
      })

      const productionClient = new ISXWebSocketClient()
      productionClient.connect()

      mockWs = (productionClient as any).ws
      expect(mockWs.url).toBe('ws://iraqiinvestor.gov.iq/ws')

      productionClient.disconnect()
      process.env.NODE_ENV = originalEnv
    })

    it('handles connection errors gracefully', async () => {
      const connectPromise = client.connect()
      
      // Simulate connection error
      mockWs = (client as any).ws
      mockWs.simulateError()

      await expect(connectPromise).rejects.toThrow('WebSocket connection failed')
      expect(client.getConnectionState()).toBe(ConnectionState.DISCONNECTED)
    })

    it('automatically reconnects after connection loss', async () => {
      jest.useFakeTimers()
      
      await client.connect()
      expect(client.getConnectionState()).toBe(ConnectionState.CONNECTED)

      // Simulate unexpected disconnection
      mockWs = (client as any).ws
      mockWs.simulateClose(1006, 'Connection lost')

      expect(client.getConnectionState()).toBe(ConnectionState.RECONNECTING)

      // Fast-forward to trigger reconnection
      jest.advanceTimersByTime(1000)

      // New connection should be established
      await Promise.resolve() // Allow promises to resolve
      expect(client.getConnectionState()).toBe(ConnectionState.CONNECTED)

      jest.useRealTimers()
    })

    it('stops reconnection attempts after max retries', async () => {
      jest.useFakeTimers()
      
      const maxRetries = 5
      client.setMaxReconnectAttempts(maxRetries)

      await client.connect()

      // Simulate repeated connection failures
      for (let i = 0; i < maxRetries + 1; i++) {
        mockWs = (client as any).ws
        mockWs.simulateClose(1006, 'Connection lost')
        jest.advanceTimersByTime(1000)
        await Promise.resolve()
      }

      expect(client.getConnectionState()).toBe(ConnectionState.DISCONNECTED)
      expect(client.getReconnectAttempts()).toBe(maxRetries)

      jest.useRealTimers()
    })

    it('resets reconnect attempts on successful connection', async () => {
      jest.useFakeTimers()
      
      await client.connect()

      // Simulate connection loss and reconnection
      mockWs = (client as any).ws
      mockWs.simulateClose(1006, 'Connection lost')
      
      jest.advanceTimersByTime(1000)
      await Promise.resolve()

      expect(client.getReconnectAttempts()).toBe(0)

      jest.useRealTimers()
    })
  })

  describe('Message Handling', () => {
    beforeEach(async () => {
      await client.connect()
      mockWs = (client as any).ws
    })

    it('handles license status updates', () => {
      const statusHandler = jest.fn()
      client.subscribe('license_status', statusHandler)

      const statusUpdate = {
        type: 'license_status',
        data: {
          isValid: true,
          expiresAt: '2025-12-31T23:59:59Z',
          features: ['Daily Reports Access', 'Advanced Analytics'],
          userInfo: {
            name: 'Test User',
            email: 'test@iraqiinvestor.gov.iq',
          },
        },
      }

      mockWs.simulateMessage(statusUpdate)

      expect(statusHandler).toHaveBeenCalledWith(statusUpdate.data)
    })

    it('handles operation progress updates', () => {
      const progressHandler = jest.fn()
      client.subscribe('pipeline_progress', progressHandler)

      const progressUpdate = {
        type: 'pipeline_progress',
        data: {
          pipeline_id: 'operation-123',
          step: 'Data Processing',
          progress: 75,
          status: 'running',
          message: 'Processing market data...',
        },
      }

      mockWs.simulateMessage(progressUpdate)

      expect(progressHandler).toHaveBeenCalledWith(progressUpdate.data)
    })

    it('handles system notifications', () => {
      const notificationHandler = jest.fn()
      client.subscribe('system_notification', notificationHandler)

      const notification = {
        type: 'system_notification',
        data: {
          level: 'info',
          title: 'Iraqi Investor Update',
          message: 'New market data available',
          timestamp: '2025-01-26T10:30:00Z',
        },
      }

      mockWs.simulateMessage(notification)

      expect(notificationHandler).toHaveBeenCalledWith(notification.data)
    })

    it('handles multiple subscribers for same message type', () => {
      const handler1 = jest.fn()
      const handler2 = jest.fn()
      
      client.subscribe('license_status', handler1)
      client.subscribe('license_status', handler2)

      const statusUpdate = {
        type: 'license_status',
        data: { isValid: false },
      }

      mockWs.simulateMessage(statusUpdate)

      expect(handler1).toHaveBeenCalledWith(statusUpdate.data)
      expect(handler2).toHaveBeenCalledWith(statusUpdate.data)
    })

    it('ignores malformed messages', () => {
      const consoleWarn = jest.spyOn(console, 'warn').mockImplementation()
      const handler = jest.fn()
      
      client.subscribe('license_status', handler)

      // Simulate malformed JSON
      if (mockWs.onmessage) {
        mockWs.onmessage(new MessageEvent('message', { data: 'invalid-json' }))
      }

      expect(handler).not.toHaveBeenCalled()
      expect(consoleWarn).toHaveBeenCalledWith('Failed to parse WebSocket message:', expect.any(Error))

      consoleWarn.mockRestore()
    })

    it('handles messages without type field', () => {
      const consoleWarn = jest.spyOn(console, 'warn').mockImplementation()
      const handler = jest.fn()
      
      client.subscribe('license_status', handler)

      const messageWithoutType = {
        data: { isValid: true },
      }

      mockWs.simulateMessage(messageWithoutType)

      expect(handler).not.toHaveBeenCalled()
      expect(consoleWarn).toHaveBeenCalledWith('WebSocket message missing type field')

      consoleWarn.mockRestore()
    })
  })

  describe('Subscription Management', () => {
    beforeEach(async () => {
      await client.connect()
    })

    it('allows subscribing to message types', () => {
      const handler = jest.fn()
      const unsubscribe = client.subscribe('license_status', handler)

      expect(typeof unsubscribe).toBe('function')
    })

    it('allows unsubscribing from message types', () => {
      const handler = jest.fn()
      const unsubscribe = client.subscribe('license_status', handler)

      unsubscribe()

      const statusUpdate = {
        type: 'license_status',
        data: { isValid: true },
      }

      mockWs = (client as any).ws
      mockWs.simulateMessage(statusUpdate)

      expect(handler).not.toHaveBeenCalled()
    })

    it('handles unsubscribing non-existent handlers', () => {
      const handler = jest.fn()
      const unsubscribe = client.subscribe('license_status', handler)

      // Unsubscribe twice
      unsubscribe()
      unsubscribe() // Should not throw

      expect(() => unsubscribe()).not.toThrow()
    })

    it('clears all subscriptions on disconnect', () => {
      const handler1 = jest.fn()
      const handler2 = jest.fn()
      
      client.subscribe('license_status', handler1)
      client.subscribe('pipeline_progress', handler2)

      client.disconnect()

      // Reconnect and test
      client.connect().then(() => {
        mockWs = (client as any).ws
        
        mockWs.simulateMessage({
          type: 'license_status',
          data: { isValid: true },
        })

        mockWs.simulateMessage({
          type: 'pipeline_progress',
          data: { progress: 50 },
        })

        expect(handler1).not.toHaveBeenCalled()
        expect(handler2).not.toHaveBeenCalled()
      })
    })
  })

  describe('Sending Messages', () => {
    beforeEach(async () => {
      await client.connect()
      mockWs = (client as any).ws
    })

    it('sends messages when connected', () => {
      const sendSpy = jest.spyOn(mockWs, 'send')
      
      const message = {
        type: 'request_license_refresh',
        requestId: 'req-123',
      }

      client.send(message)

      expect(sendSpy).toHaveBeenCalledWith(JSON.stringify(message))
    })

    it('queues messages when disconnected', () => {
      client.disconnect()

      const message = {
        type: 'request_license_refresh',
        requestId: 'req-123',
      }

      // Should not throw
      expect(() => client.send(message)).not.toThrow()

      // Message should be queued for when connection is restored
      expect(client.getQueuedMessageCount()).toBe(1)
    })

    it('sends queued messages after reconnection', async () => {
      const message1 = { type: 'message1' }
      const message2 = { type: 'message2' }

      // Disconnect and queue messages
      client.disconnect()
      client.send(message1)
      client.send(message2)

      expect(client.getQueuedMessageCount()).toBe(2)

      // Reconnect
      await client.connect()
      mockWs = (client as any).ws
      const sendSpy = jest.spyOn(mockWs, 'send')

      // Wait for queued messages to be sent
      await Promise.resolve()

      expect(sendSpy).toHaveBeenCalledWith(JSON.stringify(message1))
      expect(sendSpy).toHaveBeenCalledWith(JSON.stringify(message2))
      expect(client.getQueuedMessageCount()).toBe(0)
    })

    it('limits message queue size', () => {
      const maxQueueSize = 50
      client.setMaxQueueSize(maxQueueSize)

      client.disconnect()

      // Queue more messages than the limit
      for (let i = 0; i < maxQueueSize + 10; i++) {
        client.send({ type: 'test', id: i })
      }

      expect(client.getQueuedMessageCount()).toBe(maxQueueSize)
    })
  })

  describe('Connection State Tracking', () => {
    it('tracks connection state changes', async () => {
      const stateHandler = jest.fn()
      client.onConnectionStateChange(stateHandler)

      expect(client.getConnectionState()).toBe(ConnectionState.DISCONNECTED)

      await client.connect()
      expect(stateHandler).toHaveBeenCalledWith(ConnectionState.CONNECTING)
      expect(stateHandler).toHaveBeenCalledWith(ConnectionState.CONNECTED)

      client.disconnect()
      expect(stateHandler).toHaveBeenCalledWith(ConnectionState.DISCONNECTED)
    })

    it('provides connection statistics', async () => {
      await client.connect()

      const stats = client.getConnectionStats()
      
      expect(stats).toMatchObject({
        isConnected: true,
        reconnectAttempts: 0,
        queuedMessages: 0,
        totalMessagesSent: 0,
        totalMessagesReceived: 0,
        lastConnectionTime: expect.any(Date),
      })
    })

    it('tracks message statistics', async () => {
      await client.connect()
      mockWs = (client as any).ws

      // Send a message
      client.send({ type: 'test' })

      // Receive a message
      mockWs.simulateMessage({ type: 'response', data: {} })

      const stats = client.getConnectionStats()
      expect(stats.totalMessagesSent).toBe(1)
      expect(stats.totalMessagesReceived).toBe(1)
    })
  })

  describe('Heartbeat and Keep-Alive', () => {
    it('sends periodic ping messages', async () => {
      jest.useFakeTimers()
      
      await client.connect()
      mockWs = (client as any).ws
      const sendSpy = jest.spyOn(mockWs, 'send')

      // Fast-forward to trigger heartbeat
      jest.advanceTimersByTime(30000) // 30 seconds

      expect(sendSpy).toHaveBeenCalledWith(JSON.stringify({
        type: 'ping',
        timestamp: expect.any(Number),
      }))

      jest.useRealTimers()
    })

    it('handles pong responses', async () => {
      await client.connect()
      mockWs = (client as any).ws

      const pingTime = Date.now()
      
      // Simulate pong response
      mockWs.simulateMessage({
        type: 'pong',
        timestamp: pingTime,
      })

      const stats = client.getConnectionStats()
      expect(stats.lastPingTime).toBeGreaterThan(0)
    })

    it('detects connection timeout from missing pongs', async () => {
      jest.useFakeTimers()
      
      await client.connect()
      mockWs = (client as any).ws

      // Send ping but don't respond with pong
      jest.advanceTimersByTime(30000) // Send ping
      jest.advanceTimersByTime(10000) // Wait for timeout

      expect(client.getConnectionState()).toBe(ConnectionState.RECONNECTING)

      jest.useRealTimers()
    })
  })

  describe('Error Handling and Recovery', () => {
    it('handles WebSocket errors gracefully', async () => {
      await client.connect()
      mockWs = (client as any).ws

      const errorHandler = jest.fn()
      client.onError(errorHandler)

      mockWs.simulateError()

      expect(errorHandler).toHaveBeenCalledWith(expect.any(Error))
    })

    it('attempts exponential backoff for reconnections', async () => {
      jest.useFakeTimers()
      
      await client.connect()
      
      // Track reconnection delays
      const delays: number[] = []
      const originalSetTimeout = global.setTimeout
      global.setTimeout = jest.fn().mockImplementation((fn, delay) => {
        delays.push(delay)
        return originalSetTimeout(fn, 0) // Execute immediately for testing
      })

      // Simulate multiple connection failures
      for (let i = 0; i < 3; i++) {
        mockWs = (client as any).ws
        mockWs.simulateClose(1006, 'Connection lost')
        jest.advanceTimersByTime(100)
        await Promise.resolve()
      }

      // Verify exponential backoff pattern
      expect(delays.length).toBeGreaterThan(0)
      expect(delays[1]).toBeGreaterThan(delays[0])

      global.setTimeout = originalSetTimeout
      jest.useRealTimers()
    })

    it('recovers from temporary network issues', async () => {
      await client.connect()
      expect(client.getConnectionState()).toBe(ConnectionState.CONNECTED)

      // Simulate network issue
      mockWs = (client as any).ws
      mockWs.simulateClose(1006, 'Network error')

      expect(client.getConnectionState()).toBe(ConnectionState.RECONNECTING)

      // Should eventually reconnect
      await new Promise(resolve => setTimeout(resolve, 100))
      expect(client.getConnectionState()).toBe(ConnectionState.CONNECTED)
    })
  })
})