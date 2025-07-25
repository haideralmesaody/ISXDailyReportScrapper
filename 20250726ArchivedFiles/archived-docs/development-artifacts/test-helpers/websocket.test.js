/**
 * Tests for WebSocket Connection Manager
 */

// Mock MessageAdapter
class MockMessageAdapter {
    normalizeMessage(message) {
        return message;
    }
}

// Override MessageAdapter globally
global.MessageAdapter = MockMessageAdapter;

// Import module
require('./websocket.js');

describe('WebSocketManager', () => {
    let wsManager;
    let mockWebSocket;
    let mockSetInterval;
    let mockClearInterval;
    let mockSetTimeout;
    let mockClearTimeout;
    let originalLocation;
    
    beforeEach(() => {
        // Mock global functions
        mockSetInterval = jest.fn(() => 123);
        mockClearInterval = jest.fn();
        mockSetTimeout = jest.fn((fn) => {
            // Store the callback for manual triggering if needed
            mockSetTimeout.callback = fn;
            return 456;
        });
        mockClearTimeout = jest.fn();
        
        global.setInterval = mockSetInterval;
        global.clearInterval = mockClearInterval;
        global.setTimeout = mockSetTimeout;
        global.clearTimeout = mockClearTimeout;
        
        // Mock window.location by replacing the global
        originalLocation = window.location;
        global.window = Object.create(window);
        global.window.location = {
            protocol: 'http:',
            host: 'localhost:8080',
            hostname: 'localhost'
        };
        
        // Mock WebSocket
        mockWebSocket = {
            readyState: WebSocket.OPEN,
            send: jest.fn(),
            close: jest.fn(),
            addEventListener: jest.fn(),
            removeEventListener: jest.fn()
        };
        
        global.WebSocket = jest.fn(() => mockWebSocket);
        
        // Mock logger
        window.ISXLogger = {
            info: jest.fn(),
            warn: jest.fn(),
            error: jest.fn()
        };
        
        window.LogCategory = {
            WEBSOCKET: 'WebSocket'
        };
        
        // Create new instance
        wsManager = new WebSocketManager();
    });
    
    afterEach(() => {
        window.location = originalLocation;
    });

    describe('constructor', () => {
        test('should initialize with default values', () => {
            expect(wsManager.ws).toBeNull();
            expect(wsManager.isConnected).toBe(false);
            expect(wsManager.handlers).toBeInstanceOf(Map);
            expect(wsManager.reconnectAttempts).toBe(0);
            expect(wsManager.maxReconnectAttempts).toBe(5);
            expect(wsManager.reconnectDelay).toBe(1000);
            expect(wsManager.heartbeatInterval).toBeNull();
            expect(wsManager.connectionStatusCallback).toBeNull();
            expect(wsManager.messageAdapter).toBeInstanceOf(MockMessageAdapter);
        });
    });

    describe('connect', () => {
        test('should create WebSocket with correct URL', () => {
            wsManager.connect();
            
            expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8080/ws');
            expect(wsManager.ws).toBe(mockWebSocket);
            expect(mockSetTimeout).toHaveBeenCalledWith(expect.any(Function), 5000);
        });

        test('should use wss for https protocol', () => {
            window.location.protocol = 'https:';
            wsManager.connect();
            
            expect(global.WebSocket).toHaveBeenCalledWith('wss://localhost:8080/ws');
        });

        test('should handle connection errors', () => {
            global.WebSocket = jest.fn(() => {
                throw new Error('Connection failed');
            });
            
            wsManager.handleConnectionError = jest.fn();
            wsManager.connect();
            
            expect(window.ISXLogger.error).toHaveBeenCalled();
            expect(wsManager.handleConnectionError).toHaveBeenCalled();
        });

        test('should set connection timeout', () => {
            wsManager.updateConnectionStatus = jest.fn();
            wsManager.connect();
            
            // Trigger timeout
            mockSetTimeout.callback();
            
            expect(window.ISXLogger.warn).toHaveBeenCalledWith(
                'WebSocket',
                'WebSocket connection timeout'
            );
            expect(wsManager.updateConnectionStatus).toHaveBeenCalledWith(false);
        });
    });

    describe('setupEventHandlers', () => {
        beforeEach(() => {
            wsManager.ws = mockWebSocket;
            wsManager.updateConnectionStatus = jest.fn();
            wsManager.startHeartbeat = jest.fn();
            wsManager.stopHeartbeat = jest.fn();
            wsManager.handleMessage = jest.fn();
            wsManager.scheduleReconnect = jest.fn();
            wsManager.handleConnectionError = jest.fn();
        });

        test('should handle onopen event', () => {
            wsManager.setupEventHandlers();
            wsManager.connectionTimeout = 456;
            
            mockWebSocket.onopen();
            
            expect(wsManager.isConnected).toBe(true);
            expect(wsManager.reconnectAttempts).toBe(0);
            expect(mockClearTimeout).toHaveBeenCalledWith(456);
            expect(wsManager.updateConnectionStatus).toHaveBeenCalledWith(true);
            expect(wsManager.startHeartbeat).toHaveBeenCalled();
        });

        test('should handle onmessage event', () => {
            const event = { data: '{"type": "test"}' };
            wsManager.setupEventHandlers();
            
            mockWebSocket.onmessage(event);
            
            expect(wsManager.handleMessage).toHaveBeenCalledWith(event);
        });

        test('should handle onclose event - clean close', () => {
            wsManager.setupEventHandlers();
            const event = { wasClean: true, code: 1000, reason: 'Normal' };
            
            mockWebSocket.onclose(event);
            
            expect(wsManager.isConnected).toBe(false);
            expect(wsManager.updateConnectionStatus).toHaveBeenCalledWith(false);
            expect(wsManager.stopHeartbeat).toHaveBeenCalled();
            expect(wsManager.scheduleReconnect).not.toHaveBeenCalled();
        });

        test('should handle onclose event - unclean close with reconnect', () => {
            wsManager.setupEventHandlers();
            const event = { wasClean: false, code: 1006, reason: 'Abnormal' };
            
            mockWebSocket.onclose(event);
            
            expect(wsManager.scheduleReconnect).toHaveBeenCalled();
        });

        test('should handle onclose event - max reconnects reached', () => {
            wsManager.setupEventHandlers();
            wsManager.reconnectAttempts = 5;
            const event = { wasClean: false, code: 1006, reason: 'Abnormal' };
            
            mockWebSocket.onclose(event);
            
            expect(wsManager.scheduleReconnect).not.toHaveBeenCalled();
        });

        test('should handle onerror event', () => {
            wsManager.setupEventHandlers();
            const error = new Error('WebSocket error');
            
            mockWebSocket.onerror(error);
            
            expect(console.error).toHaveBeenCalledWith('WebSocket error:', error);
            expect(wsManager.handleConnectionError).toHaveBeenCalled();
        });
    });

    describe('handleMessage', () => {
        beforeEach(() => {
            wsManager.messageAdapter = new MockMessageAdapter();
        });

        test('should parse and route message to handler', () => {
            const handler = jest.fn();
            wsManager.onMessage('test', handler);
            
            const event = { data: '{"type": "test", "payload": "data"}' };
            wsManager.handleMessage(event);
            
            expect(handler).toHaveBeenCalledWith({
                type: 'test',
                payload: 'data'
            });
        });

        test('should handle stage_progress legacy mapping', () => {
            const handler = jest.fn();
            wsManager.onMessage('pipeline_progress', handler);
            
            const event = { data: '{"type": "stage_progress", "stage": "test"}' };
            wsManager.handleMessage(event);
            
            expect(handler).toHaveBeenCalled();
        });

        test('should handle refresh to data_update conversion', () => {
            const handler = jest.fn();
            wsManager.onMessage('data_update', handler);
            
            const event = { data: '{"type": "refresh"}' };
            wsManager.handleMessage(event);
            
            expect(handler).toHaveBeenCalledWith(expect.objectContaining({
                type: 'data_update',
                subtype: 'all',
                action: 'refresh'
            }));
        });

        test('should warn for unhandled message types', () => {
            const event = { data: '{"type": "unknown"}' };
            wsManager.handleMessage(event);
            
            expect(console.warn).toHaveBeenCalledWith('No handler for message type:', 'unknown');
        });

        test('should handle JSON parse errors', () => {
            const event = { data: 'invalid json' };
            wsManager.handleMessage(event);
            
            expect(console.error).toHaveBeenCalledWith(
                'Failed to parse WebSocket message:',
                expect.any(Error)
            );
        });

        test('should log connection-related messages', () => {
            const event = { data: '{"type": "connection", "status": "ok"}' };
            wsManager.handleMessage(event);
            
            expect(console.log).toHaveBeenCalledWith(
                '[WebSocket] Received message:',
                'connection',
                expect.any(Object)
            );
        });
    });

    describe('onMessage / offMessage', () => {
        test('should register message handler', () => {
            const handler = jest.fn();
            wsManager.onMessage('test', handler);
            
            expect(wsManager.handlers.has('test')).toBe(true);
            expect(wsManager.handlers.get('test')).toBe(handler);
        });

        test('should remove message handler', () => {
            const handler = jest.fn();
            wsManager.onMessage('test', handler);
            wsManager.offMessage('test');
            
            expect(wsManager.handlers.has('test')).toBe(false);
        });
    });

    describe('onConnectionStatus', () => {
        test('should set connection status callback', () => {
            const callback = jest.fn();
            wsManager.onConnectionStatus(callback);
            
            expect(wsManager.connectionStatusCallback).toBe(callback);
        });
    });

    describe('updateConnectionStatus', () => {
        test('should call connection status callback', () => {
            const callback = jest.fn();
            wsManager.onConnectionStatus(callback);
            
            wsManager.updateConnectionStatus(true);
            expect(callback).toHaveBeenCalledWith(true);
            
            wsManager.updateConnectionStatus(false);
            expect(callback).toHaveBeenCalledWith(false);
        });

        test('should update isConnected property', () => {
            wsManager.updateConnectionStatus(true);
            expect(wsManager.isConnected).toBe(true);
            
            wsManager.updateConnectionStatus(false);
            expect(wsManager.isConnected).toBe(false);
        });

        test('should log status changes', () => {
            wsManager.isConnected = false;
            wsManager.updateConnectionStatus(true);
            
            expect(console.log).toHaveBeenCalledWith(
                '[WebSocket] Connection status changed: false -> true'
            );
        });
    });

    describe('handleConnectionError', () => {
        test('should update status and stop heartbeat', () => {
            wsManager.updateConnectionStatus = jest.fn();
            wsManager.stopHeartbeat = jest.fn();
            wsManager.connectionTimeout = 456;
            
            wsManager.handleConnectionError();
            
            expect(wsManager.isConnected).toBe(false);
            expect(mockClearTimeout).toHaveBeenCalledWith(456);
            expect(wsManager.updateConnectionStatus).toHaveBeenCalledWith(false);
            expect(wsManager.stopHeartbeat).toHaveBeenCalled();
        });
    });

    describe('scheduleReconnect', () => {
        test('should schedule reconnection with exponential backoff', () => {
            wsManager.connect = jest.fn();
            
            wsManager.scheduleReconnect();
            expect(wsManager.reconnectAttempts).toBe(1);
            expect(mockSetTimeout).toHaveBeenCalledWith(expect.any(Function), 1000);
            
            wsManager.scheduleReconnect();
            expect(wsManager.reconnectAttempts).toBe(2);
            expect(mockSetTimeout).toHaveBeenCalledWith(expect.any(Function), 2000);
            
            wsManager.scheduleReconnect();
            expect(wsManager.reconnectAttempts).toBe(3);
            expect(mockSetTimeout).toHaveBeenCalledWith(expect.any(Function), 4000);
        });

        test('should not reconnect if already connected', () => {
            wsManager.connect = jest.fn();
            wsManager.isConnected = true;
            
            wsManager.scheduleReconnect();
            mockSetTimeout.callback();
            
            expect(wsManager.connect).not.toHaveBeenCalled();
        });
    });

    describe('startHeartbeat', () => {
        test('should start heartbeat interval', () => {
            wsManager.ws = mockWebSocket;
            wsManager.isConnected = true;
            
            wsManager.startHeartbeat();
            
            expect(mockSetInterval).toHaveBeenCalledWith(expect.any(Function), 30000);
            expect(wsManager.heartbeatInterval).toBe(123);
        });

        test('should send heartbeat message', () => {
            wsManager.ws = mockWebSocket;
            wsManager.isConnected = true;
            
            wsManager.startHeartbeat();
            const heartbeatFn = mockSetInterval.mock.calls[0][0];
            
            heartbeatFn();
            
            expect(mockWebSocket.send).toHaveBeenCalledWith(
                JSON.stringify({ type: 'heartbeat' })
            );
        });

        test('should handle heartbeat send errors', () => {
            wsManager.ws = mockWebSocket;
            wsManager.isConnected = true;
            wsManager.handleConnectionError = jest.fn();
            
            mockWebSocket.send.mockImplementation(() => {
                throw new Error('Send failed');
            });
            
            wsManager.startHeartbeat();
            const heartbeatFn = mockSetInterval.mock.calls[0][0];
            
            heartbeatFn();
            
            expect(console.error).toHaveBeenCalledWith(
                'Failed to send heartbeat:',
                expect.any(Error)
            );
            expect(wsManager.handleConnectionError).toHaveBeenCalled();
        });

        test('should not send heartbeat if disconnected', () => {
            wsManager.ws = mockWebSocket;
            wsManager.isConnected = false;
            
            wsManager.startHeartbeat();
            const heartbeatFn = mockSetInterval.mock.calls[0][0];
            
            heartbeatFn();
            
            expect(mockWebSocket.send).not.toHaveBeenCalled();
        });
    });

    describe('stopHeartbeat', () => {
        test('should clear heartbeat interval', () => {
            wsManager.heartbeatInterval = 123;
            
            wsManager.stopHeartbeat();
            
            expect(mockClearInterval).toHaveBeenCalledWith(123);
            expect(wsManager.heartbeatInterval).toBeNull();
        });

        test('should handle null heartbeat interval', () => {
            wsManager.heartbeatInterval = null;
            
            expect(() => wsManager.stopHeartbeat()).not.toThrow();
            expect(mockClearInterval).not.toHaveBeenCalled();
        });
    });

    describe('disconnect', () => {
        test('should close WebSocket and update status', () => {
            wsManager.ws = mockWebSocket;
            wsManager.isConnected = true;
            wsManager.stopHeartbeat = jest.fn();
            wsManager.updateConnectionStatus = jest.fn();
            
            wsManager.disconnect();
            
            expect(wsManager.stopHeartbeat).toHaveBeenCalled();
            expect(mockWebSocket.close).toHaveBeenCalledWith(1000, 'Manual disconnect');
            expect(wsManager.ws).toBeNull();
            expect(wsManager.isConnected).toBe(false);
            expect(wsManager.updateConnectionStatus).toHaveBeenCalledWith(false);
        });

        test('should handle disconnect when no WebSocket', () => {
            wsManager.ws = null;
            wsManager.updateConnectionStatus = jest.fn();
            
            expect(() => wsManager.disconnect()).not.toThrow();
            expect(wsManager.updateConnectionStatus).toHaveBeenCalledWith(false);
        });
    });

    describe('getConnectionStatus', () => {
        test('should return true when fully connected', () => {
            wsManager.isConnected = true;
            wsManager.ws = mockWebSocket;
            mockWebSocket.readyState = WebSocket.OPEN;
            
            expect(wsManager.getConnectionStatus()).toBe(true);
        });

        test('should return false when not connected', () => {
            wsManager.isConnected = false;
            expect(wsManager.getConnectionStatus()).toBe(false);
        });

        test('should return false when no WebSocket', () => {
            wsManager.isConnected = true;
            wsManager.ws = null;
            expect(wsManager.getConnectionStatus()).toBe(false);
        });

        test('should return false when WebSocket not open', () => {
            wsManager.isConnected = true;
            wsManager.ws = mockWebSocket;
            mockWebSocket.readyState = WebSocket.CONNECTING;
            
            expect(wsManager.getConnectionStatus()).toBe(false);
        });
    });

    describe('global exports', () => {
        test('should export to window object', () => {
            expect(window.WebSocketManager).toBe(WebSocketManager);
        });
    });
});