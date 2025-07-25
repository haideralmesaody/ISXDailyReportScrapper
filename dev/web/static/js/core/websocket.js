/**
 * WebSocket Connection Manager
 * Handles WebSocket connection, reconnection, and message routing
 * Follows architecture principles: Display updates only, no business logic
 */

class WebSocketManager {
    constructor() {
        this.ws = null;
        this.isConnected = false;
        this.handlers = new Map();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.heartbeatInterval = null;
        this.connectionStatusCallback = null;
        this.messageAdapter = new MessageAdapter();
    }

    /**
     * Initialize WebSocket connection
     */
    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            this.setupEventHandlers();
            
            // Set a timeout to mark as disconnected if connection doesn't open
            this.connectionTimeout = setTimeout(() => {
                if (!this.isConnected) {
                    window.ISXLogger.warn(LogCategory.WEBSOCKET, 'WebSocket connection timeout');
                    this.updateConnectionStatus(false);
                }
            }, 5000); // 5 second timeout
        } catch (error) {
            window.ISXLogger.error(LogCategory.WEBSOCKET, 'Failed to create WebSocket connection:', error);
            this.handleConnectionError();
        }
    }

    /**
     * Set up WebSocket event handlers
     */
    setupEventHandlers() {
        this.ws.onopen = () => {
            window.ISXLogger.info(LogCategory.WEBSOCKET, 'WebSocket connected');
            console.log('[WebSocket] onopen event fired, WebSocket readyState:', this.ws.readyState);
            this.isConnected = true;
            this.reconnectAttempts = 0;
            
            // Clear connection timeout
            if (this.connectionTimeout) {
                clearTimeout(this.connectionTimeout);
                this.connectionTimeout = null;
            }
            
            // Log status change for debugging
            console.log('[WebSocket] Status changing from disconnected to connected');
            this.updateConnectionStatus(true);
            this.startHeartbeat();
        };

        this.ws.onmessage = (event) => {
            this.handleMessage(event);
        };

        this.ws.onclose = (event) => {
            console.log('[WebSocket] Disconnected:', event.code, event.reason);
            console.log('[WebSocket] Status changing from connected to disconnected');
            this.isConnected = false;
            this.updateConnectionStatus(false);
            this.stopHeartbeat();
            
            if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
                this.scheduleReconnect();
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.handleConnectionError();
        };
    }

    /**
     * Handle incoming WebSocket messages
     */
    handleMessage(event) {
        try {
            const rawMessage = JSON.parse(event.data);
            
            // Normalize message format
            const message = this.messageAdapter.normalizeMessage(rawMessage);
            
            // Log connection-related messages for debugging
            if (message.type === 'connection' || message.type === 'status') {
                console.log('[WebSocket] Received message:', message.type, message);
            }
            
            // Route message to appropriate handler
            if (this.handlers.has(message.type)) {
                this.handlers.get(message.type)(message);
            } else {
                // Try legacy handlers for backward compatibility
                if (rawMessage.type === 'stage_progress' && this.handlers.has('pipeline_progress')) {
                    this.handlers.get('pipeline_progress')(message);
                } else if (rawMessage.type === 'refresh' && this.handlers.has('data_update')) {
                    // Convert refresh to data_update
                    message.type = 'data_update';
                    message.subtype = 'all';
                    message.action = 'refresh';
                    this.handlers.get('data_update')(message);
                } else {
                    console.warn('No handler for message type:', message.type);
                }
            }
        } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
        }
    }

    /**
     * Register a message handler
     * @param {string} type - Message type
     * @param {Function} handler - Handler function
     */
    onMessage(type, handler) {
        this.handlers.set(type, handler);
    }

    /**
     * Remove a message handler
     * @param {string} type - Message type
     */
    offMessage(type) {
        this.handlers.delete(type);
    }

    /**
     * Set connection status callback
     * @param {Function} callback - Function to call on status change
     */
    onConnectionStatus(callback) {
        this.connectionStatusCallback = callback;
    }

    /**
     * Update connection status
     * @param {boolean} connected - Connection status
     */
    updateConnectionStatus(connected) {
        // Log all status updates for debugging
        console.log(`[WebSocket] updateConnectionStatus called with connected=${connected}`);
        const previousStatus = this.isConnected;
        this.isConnected = connected;
        
        if (previousStatus !== connected) {
            console.log(`[WebSocket] Connection status changed: ${previousStatus} -> ${connected}`);
        }
        
        if (this.connectionStatusCallback) {
            this.connectionStatusCallback(connected);
        }
    }

    /**
     * Handle connection errors
     */
    handleConnectionError() {
        this.isConnected = false;
        
        // Clear connection timeout if exists
        if (this.connectionTimeout) {
            clearTimeout(this.connectionTimeout);
            this.connectionTimeout = null;
        }
        
        this.updateConnectionStatus(false);
        this.stopHeartbeat();
    }

    /**
     * Schedule reconnection attempt
     */
    scheduleReconnect() {
        this.reconnectAttempts++;
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
        
        console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`);
        
        setTimeout(() => {
            if (!this.isConnected) {
                this.connect();
            }
        }, delay);
    }

    /**
     * Start heartbeat to keep connection alive
     */
    startHeartbeat() {
        this.stopHeartbeat();
        this.heartbeatInterval = setInterval(() => {
            if (this.isConnected && this.ws.readyState === WebSocket.OPEN) {
                try {
                    this.ws.send(JSON.stringify({ type: 'heartbeat' }));
                } catch (error) {
                    console.error('Failed to send heartbeat:', error);
                    this.handleConnectionError();
                }
            }
        }, 30000); // Send heartbeat every 30 seconds
    }

    /**
     * Stop heartbeat
     */
    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    }

    /**
     * Close WebSocket connection
     */
    disconnect() {
        this.stopHeartbeat();
        if (this.ws) {
            this.ws.close(1000, 'Manual disconnect');
            this.ws = null;
        }
        this.isConnected = false;
        this.updateConnectionStatus(false);
    }

    /**
     * Get connection status
     * @returns {boolean} Connection status
     */
    getConnectionStatus() {
        return this.isConnected && this.ws && this.ws.readyState === WebSocket.OPEN;
    }
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = WebSocketManager;
}

// Global access for compatibility
window.WebSocketManager = WebSocketManager;