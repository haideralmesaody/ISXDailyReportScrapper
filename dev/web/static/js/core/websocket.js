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
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.handleConnectionError();
        }
    }

    /**
     * Set up WebSocket event handlers
     */
    setupEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.isConnected = true;
            this.reconnectAttempts = 0;
            this.updateConnectionStatus(true);
            this.startHeartbeat();
        };

        this.ws.onmessage = (event) => {
            this.handleMessage(event);
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket disconnected:', event.code, event.reason);
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
            const message = JSON.parse(event.data);
            
            // Route message to appropriate handler
            if (this.handlers.has(message.type)) {
                this.handlers.get(message.type)(message);
            } else {
                console.warn('No handler for message type:', message.type);
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
        this.isConnected = connected;
        if (this.connectionStatusCallback) {
            this.connectionStatusCallback(connected);
        }
    }

    /**
     * Handle connection errors
     */
    handleConnectionError() {
        this.isConnected = false;
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