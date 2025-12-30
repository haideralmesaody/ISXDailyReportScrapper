/**
 * Simplified WebSocket client for ISX Pulse
 *
 * Lightweight WebSocket client optimized for localhost communication.
 * Removed distributed system complexity while maintaining React compatibility.
 */

'use client'

export type WebSocketEventType =
  | 'operation:snapshot'  // Primary event for all operation updates
  | 'operation:delta'
  | 'market_update'
  | 'system_status'
  | 'license_status'
  | 'connection_status'

// ============================================================================
// Simplified WebSocket Types
// ============================================================================

export interface ConnectionStatus {
  status: 'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting'
  timestamp: string
  error?: string
  attempt?: number
}

export interface SimpleMessage {
  type: string
  data: any
  timestamp?: number
}

// ============================================================================
// Simplified WebSocket Client
// ============================================================================

export class ISXWebSocketClient {
  private ws: WebSocket | null = null
  private listeners: Map<string, Set<(data: any) => void>> = new Map()
  private isManualClose = false
  private connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting' = 'disconnected'
  private url: string
  private debug: boolean
  private isListenersReady = false // Track when listeners are ready to receive messages

  // Enhanced retry logic with exponential backoff
  private reconnectAttempts: number = 0
  private maxReconnectAttempts: number = 5 // Increased from 3 to 5
  private reconnectTimeoutId: NodeJS.Timeout | null = null
  private baseReconnectDelay: number = 1000 // Start with 1 second
  private maxReconnectDelay: number = 30000 // Cap at 30 seconds
  private lastConnectionAttempt: number = 0

  constructor() {
    // Simple URL construction
    if (typeof window === 'undefined') {
      this.url = 'ws://localhost:8080/ws'
    } else if (window.location.protocol === 'file:' || !window.location.host) {
      this.url = 'ws://localhost:8080/ws'
    } else {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      this.url = `${protocol}//${window.location.host}/ws`
    }

    this.debug = process.env.NODE_ENV === 'development'
  }

  // Call this method when React components are ready to receive messages
  markListenersReady(): void {
    this.isListenersReady = true
    if (this.debug) {
      console.log('[WebSocket] Listeners marked as ready - messages will now be emitted')
    }
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve()
        return
      }

      this.performConnection(resolve, reject)
    })
  }

  private performConnection(resolve: () => void, reject: (reason?: any) => void): void {
    try {
      this.connectionStatus = 'connecting'

      if (this.debug) {
        console.log(`[WebSocket] Connecting to: ${this.url} (attempt ${this.reconnectAttempts + 1})`)
      }

      this.ws = new WebSocket(this.url)

      // Connection timeout
      const connectionTimeout = setTimeout(() => {
        if (this.ws?.readyState === WebSocket.CONNECTING) {
          console.warn(`[WebSocket] Connection timeout`)
          this.ws.close(1006, 'Connection timeout')
        }
      }, 5000) // 5 second timeout

      this.ws.onopen = () => {
        clearTimeout(connectionTimeout)

        if (this.debug) {
          console.log(`[WebSocket] Connected successfully`)
        }

        this.connectionStatus = 'connected'
        this.isManualClose = false
        this.reconnectAttempts = 0

        this.emit('connection_status', {
          status: 'connected',
          timestamp: new Date().toISOString(),
          attempt: this.reconnectAttempts
        })

        resolve()
      }

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data)
      }

      this.ws.onerror = (error) => {
        clearTimeout(connectionTimeout)
        this.connectionStatus = 'error'

        if (this.debug) {
          console.error(`[WebSocket] Connection error:`, error)
        }

        this.emit('connection_status', {
          status: 'error',
          timestamp: new Date().toISOString(),
          error: `WebSocket error: ${error}`,
          attempt: this.reconnectAttempts
        })

        reject(error)
      }

      this.ws.onclose = (event) => {
        clearTimeout(connectionTimeout)
        this.connectionStatus = 'disconnected'

        if (!this.isManualClose) {
          if (this.debug) {
            console.log(`[WebSocket] Connection closed:`, {
              code: event.code,
              reason: event.reason || 'No reason provided',
              wasClean: event.wasClean
            })
          }

          this.emit('connection_status', {
            status: 'disconnected',
            timestamp: new Date().toISOString(),
            reason: event.reason || 'Unknown reason',
            code: event.code,
            wasClean: event.wasClean
          })

          // Simple reconnection logic
          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect()
          } else {
            console.error(`[WebSocket] Max reconnection attempts reached`)
            this.connectionStatus = 'error'
            this.emit('connection_status', {
              status: 'error',
              timestamp: new Date().toISOString(),
              error: 'Max reconnection attempts reached',
              attempt: this.reconnectAttempts
            })
          }
        }
      }
    } catch (err) {
      this.connectionStatus = 'error'
      if (this.debug) {
        console.error(`[WebSocket] Connection setup error:`, err)
      }
      reject(err)
    }
  }

  private scheduleReconnect(): void {
    this.reconnectAttempts++
    this.connectionStatus = 'reconnecting'

    // Calculate exponential backoff delay
    const delay = this.calculateReconnectDelay()

    if (this.debug) {
      console.log(`[WebSocket] Scheduling reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
    }

    this.lastConnectionAttempt = Date.now()

    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectTimeoutId = null
      this.performConnection(
        () => {
          if (this.debug) {
            console.log(`[WebSocket] Reconnect successful after ${this.reconnectAttempts} attempts`)
          }
        },
        (error) => {
          if (this.debug) {
            console.error(`[WebSocket] Reconnect failed:`, error)
          }

          if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect()
          } else {
            this.connectionStatus = 'error'
            console.error(`[WebSocket] Max reconnection attempts (${this.maxReconnectAttempts}) reached`)
          }
        }
      )
    }, delay)
  }

  private calculateReconnectDelay(): number {
    // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (capped)
    const delay = this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
    return Math.min(delay, this.maxReconnectDelay)
  }

  disconnect(): void {
    this.isManualClose = true
    this.reconnectAttempts = 0
    this.lastConnectionAttempt = 0

    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId)
      this.reconnectTimeoutId = null
    }

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }

    this.connectionStatus = 'disconnected'

    if (this.debug) {
      console.log(`[WebSocket] Disconnected manually`)
    }
  }

  private handleMessage(message: any): void {
    try {
      const parsedMessage = typeof message === 'string' ? JSON.parse(message) : message

      // Extract event type from message
      const eventType = parsedMessage.type || parsedMessage.event || 'unknown'

      // Special handling for operation:snapshot events
      if (eventType === 'operation:snapshot' || eventType === 'operation_snapshot') {
        // Use full snapshot if it contains operation_id, otherwise fall back to metadata
        let data = parsedMessage.data?.operation_id ? parsedMessage.data : parsedMessage.data?.metadata || parsedMessage

        // Enhanced handling for file-level progress data
        if (data) {
          // Validate and normalize file progress data
          data = this.normalizeFileProgressData(data)

          // Enhanced debug logging for file progress data telemetry
          if (this.debug) {
            console.log('[WebSocket Telemetry] Full operation snapshot:', {
              timestamp: new Date().toISOString(),
              operation_id: data.operation_id,
              progress: data.progress,
              status: data.status,
              stage_id: data.stage_id,
              current_phase: data.current_phase,
              phase_message: data.phase_message,
              total_files: data.total_files,
              completed_files: data.completed_files,
              failed_files: data.failed_files,
              processing_files: data.processing_files,
              files_processed: data.files_processed,
              current_file: data.current_file,
              processing_speed_mbps: data.processing_speed_mbps,
              total_size_mb: data.total_size_mb,
              processed_size_mb: data.processed_size_mb,
              estimated_remaining_ms: data.estimated_remaining_ms,
              file_statuses_count: data.file_statuses?.length || 0,
              file_statuses: data.file_statuses,
              message: data.message,
              metadata_keys: Object.keys(data.metadata || {}),
              all_keys: Object.keys(data)
            })
          }
        }

        if (process.env.NODE_ENV === 'development' && data && !data.operation_id) {
          console.warn('[WebSocket] operation_snapshot missing operation_id:', {
            hasData: !!parsedMessage.data,
            hasMetadata: !!parsedMessage.data?.metadata,
            dataKeys: parsedMessage.data ? Object.keys(parsedMessage.data) : []
          })
        }

        this.emit('operation:snapshot', data)
      } else if (eventType === 'operation:delta') {
        const deltaPayload = parsedMessage.data?.operation_id
          ? parsedMessage.data
          : parsedMessage.data || parsedMessage

        if (this.debug && deltaPayload?.operation_id) {
          console.log('[WebSocket Telemetry] Delta update:', {
            timestamp: new Date().toISOString(),
            operation_id: deltaPayload.operation_id,
            stage_id: deltaPayload.stage_id,
            status: deltaPayload.status,
            progress: deltaPayload.progress,
            sequence: deltaPayload.sequence
          })
        }

        this.emit('operation:delta', deltaPayload)
      } else {
        // Emit to specific listeners
        this.emit(eventType, parsedMessage.data || parsedMessage)
      }
    } catch (error) {
      if (this.debug) {
        console.error('[WebSocket] Error parsing message:', error, 'Raw message:', message)
      }
    }
  }

  /**
   * Normalize and validate file-level progress data from WebSocket messages
   * Ensures data consistency and provides fallbacks for missing fields
   */
  private normalizeFileProgressData(data: any): any {
    if (!data) return data

    // Create a normalized copy to avoid mutation
    const normalized = { ...data }

    // Normalize file progress counts
    if (normalized.total_files !== undefined) {
      normalized.total_files = Math.max(0, parseInt(normalized.total_files) || 0)
    }

    if (normalized.completed_files !== undefined) {
      normalized.completed_files = Math.max(0, parseInt(normalized.completed_files) || 0)
    }

    if (normalized.failed_files !== undefined) {
      normalized.failed_files = Math.max(0, parseInt(normalized.failed_files) || 0)
    }

    if (normalized.processing_files !== undefined) {
      normalized.processing_files = Math.max(0, parseInt(normalized.processing_files) || 0)
    }

    // Normalize progress percentages
    if (normalized.progress !== undefined) {
      normalized.progress = Math.max(0, Math.min(100, parseFloat(normalized.progress) || 0))
    }

    // Normalize processing speed
    if (normalized.processing_speed_mbps !== undefined) {
      normalized.processing_speed_mbps = Math.max(0, parseFloat(normalized.processing_speed_mbps) || 0)
    }

    // Normalize file sizes
    if (normalized.total_size_mb !== undefined) {
      normalized.total_size_mb = Math.max(0, parseFloat(normalized.total_size_mb) || 0)
    }

    if (normalized.processed_size_mb !== undefined) {
      normalized.processed_size_mb = Math.max(0, parseFloat(normalized.processed_size_mb) || 0)
    }

    // Normalize time estimates
    if (normalized.estimated_remaining_ms !== undefined) {
      normalized.estimated_remaining_ms = Math.max(0, parseInt(normalized.estimated_remaining_ms) || 0)
    }

    // Normalize file statuses array
    if (normalized.file_statuses && Array.isArray(normalized.file_statuses)) {
      normalized.file_statuses = normalized.file_statuses
        .filter(status => status && typeof status === 'object')
        .map(status => ({
          filename: status.filename || status.file_name || 'Unknown file',
          status: this.normalizeFileStatus(status.status),
          size_mb: typeof status.size_mb === 'number' ? Math.max(0, status.size_mb) : undefined,
          progress: typeof status.progress === 'number' ? Math.max(0, Math.min(100, status.progress)) : undefined,
          error_message: status.error_message || status.error || undefined,
          processing_time_ms: typeof status.processing_time_ms === 'number' ? Math.max(0, status.processing_time_ms) : undefined
        }))
    }

    // Calculate derived values if missing
    if (normalized.total_files && normalized.completed_files !== undefined) {
      // Calculate files_processed if missing (backward compatibility)
      if (normalized.files_processed === undefined) {
        normalized.files_processed = normalized.completed_files
      }

      // Calculate overall progress if missing
      if (normalized.progress === undefined) {
        normalized.progress = (normalized.completed_files / normalized.total_files) * 100
      }
    }

    return normalized
  }

  /**
   * Normalize file status string to standard values
   */
  private normalizeFileStatus(status: string): 'pending' | 'processing' | 'completed' | 'failed' {
    if (!status) return 'pending'

    const normalized = status.toLowerCase().trim()

    switch (normalized) {
      case 'completed':
      case 'complete':
      case 'success':
      case 'done':
        return 'completed'
      case 'processing':
      case 'running':
      case 'active':
      case 'in_progress':
        return 'processing'
      case 'failed':
      case 'error':
      case 'failed':
      case 'aborted':
        return 'failed'
      case 'pending':
      case 'waiting':
      case 'queued':
      case 'ready':
        return 'pending'
      default:
        return 'pending'
    }
  }

  private emit(event: string, data: any): void {
    // Skip emission if listeners aren't ready yet (prevents TDZ errors)
    if (!this.isListenersReady) {
      if (this.debug) {
        console.warn(`[WebSocket] Skipping message emit - listeners not ready:`, { event, dataType: typeof data })
      }
      return
    }

    const listeners = this.listeners.get(event)
    if (listeners) {
      listeners.forEach(listener => {
        try {
          listener(data)
        } catch (err) {
          if (this.debug) {
            console.error(`[WebSocket] Error in listener for ${event}:`, err)
          }
          // Prevent TDZ errors from crashing the WebSocket connection
          if (err instanceof ReferenceError && err.message.includes('before initialization')) {
            console.warn(`[WebSocket] TDZ error detected - listener may not be initialized:`, err)
          }
        }
      })
    }
  }

  subscribe(event: string, handler: (data: any) => void): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }

    const listeners = this.listeners.get(event)!
    listeners.add(handler)

    // Return unsubscribe function
    return () => {
      listeners.delete(handler)
      if (listeners.size === 0) {
        this.listeners.delete(event)
      }
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  getConnectionStatus(): 'connecting' | 'connected' | 'disconnected' | 'error' | 'reconnecting' {
    return this.connectionStatus
  }

  // Simple send method
  send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(data))
      } catch (error) {
        if (this.debug) {
          console.error('[WebSocket] Error sending message:', error)
        }
      }
    } else {
      if (this.debug) {
        console.warn('[WebSocket] Cannot send message - not connected')
      }

      // Attempt to connect if not connected
      if (this.connectionStatus === 'disconnected') {
        this.connect().catch(error => {
          if (this.debug) {
            console.error('[WebSocket] Failed to connect for message send:', error)
          }
        })
      }
    }
  }

  // Enhanced utility methods
  forceReconnect(): Promise<void> {
    if (this.debug) {
      console.log('[WebSocket] Force reconnect requested')
    }

    this.disconnect()
    this.reconnectAttempts = 0
    this.lastConnectionAttempt = 0
    this.connectionStatus = 'disconnected'
    return this.connect()
  }

  getReconnectAttempts(): number {
    return this.reconnectAttempts
  }

  getNextReconnectDelay(): number {
    return this.calculateReconnectDelay()
  }

  getTimeSinceLastConnection(): number {
    return this.lastConnectionAttempt > 0 ? Date.now() - this.lastConnectionAttempt : 0
  }

  isHealthy(): boolean {
    return this.connectionStatus === 'connected' && this.isConnected()
  }

  // Enhanced connection info
  getConnectionInfo(): {
    status: string
    url: string
    reconnectAttempts: number
    isConnected: boolean
    isHealthy: boolean
    nextReconnectDelay?: number
    timeSinceLastConnection: number
  } {
    return {
      status: this.connectionStatus,
      url: this.url,
      reconnectAttempts: this.reconnectAttempts,
      isConnected: this.isConnected(),
      isHealthy: this.isHealthy(),
      nextReconnectDelay: this.connectionStatus === 'reconnecting' ? this.calculateReconnectDelay() : undefined,
      timeSinceLastConnection: this.getTimeSinceLastConnection()
    }
  }
}

// ============================================================================
// Singleton Instance
// ============================================================================

let clientInstance: ISXWebSocketClient | null = null

export function getWebSocketClient(): ISXWebSocketClient {
  if (!clientInstance) {
    clientInstance = new ISXWebSocketClient()
  }
  return clientInstance
}

// Clean up on module unload (HMR in development)
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  const cleanup = () => {
    if (clientInstance) {
      clientInstance.disconnect()
      clientInstance = null
    }
  }

  // Handle hot module replacement
  if ((module as any).hot) {
    (module as any).hot.dispose(cleanup)
  }
}
