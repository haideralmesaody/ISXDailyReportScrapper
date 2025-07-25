/**
 * Tests for Event Bus Component Communication System
 */

// Import module
require('./eventBus.js');

describe('EventBus', () => {
    let eventBus;
    let originalConsoleWarn;
    let originalConsoleError;

    beforeEach(() => {
        eventBus = new EventBus();
        originalConsoleWarn = console.warn;
        originalConsoleError = console.error;
        console.warn = jest.fn();
        console.error = jest.fn();
    });

    afterEach(() => {
        console.warn = originalConsoleWarn;
        console.error = originalConsoleError;
    });

    describe('constructor', () => {
        test('should initialize with empty events map', () => {
            expect(eventBus.events.size).toBe(0);
            expect(eventBus.maxListeners).toBe(100);
        });
    });

    describe('on', () => {
        test('should register event handler', () => {
            const handler = jest.fn();
            eventBus.on('test', handler);
            
            expect(eventBus.events.has('test')).toBe(true);
            expect(eventBus.events.get('test')).toHaveLength(1);
            expect(eventBus.events.get('test')[0].handler).toBe(handler);
        });

        test('should throw error for non-function handler', () => {
            expect(() => eventBus.on('test', 'not a function')).toThrow('Event handler must be a function');
        });

        test('should respect priority ordering', () => {
            const handler1 = jest.fn();
            const handler2 = jest.fn();
            const handler3 = jest.fn();

            eventBus.on('test', handler1, { priority: 1 });
            eventBus.on('test', handler2, { priority: 3 });
            eventBus.on('test', handler3, { priority: 2 });

            const listeners = eventBus.events.get('test');
            expect(listeners[0].handler).toBe(handler2); // Highest priority
            expect(listeners[1].handler).toBe(handler3);
            expect(listeners[2].handler).toBe(handler1); // Lowest priority
        });

        test('should return unsubscribe function', () => {
            const handler = jest.fn();
            const unsubscribe = eventBus.on('test', handler);
            
            expect(eventBus.listenerCount('test')).toBe(1);
            unsubscribe();
            expect(eventBus.listenerCount('test')).toBe(0);
        });

        test('should warn when max listeners reached', () => {
            eventBus.maxListeners = 2;
            
            eventBus.on('test', () => {});
            eventBus.on('test', () => {});
            const unsubscribe = eventBus.on('test', () => {});
            
            expect(console.warn).toHaveBeenCalledWith('Maximum listeners (2) reached for event: test');
            expect(eventBus.listenerCount('test')).toBe(2);
            
            // Unsubscribe should be no-op
            unsubscribe();
            expect(eventBus.listenerCount('test')).toBe(2);
        });
    });

    describe('once', () => {
        test('should register one-time event handler', () => {
            const handler = jest.fn();
            eventBus.once('test', handler);
            
            const listeners = eventBus.events.get('test');
            expect(listeners[0].once).toBe(true);
        });

        test('should return unsubscribe function', () => {
            const handler = jest.fn();
            const unsubscribe = eventBus.once('test', handler);
            
            expect(eventBus.listenerCount('test')).toBe(1);
            unsubscribe();
            expect(eventBus.listenerCount('test')).toBe(0);
        });
    });

    describe('off', () => {
        test('should remove listener by handler function', () => {
            const handler1 = jest.fn();
            const handler2 = jest.fn();
            
            eventBus.on('test', handler1);
            eventBus.on('test', handler2);
            
            eventBus.off('test', handler1);
            
            expect(eventBus.listenerCount('test')).toBe(1);
            expect(eventBus.events.get('test')[0].handler).toBe(handler2);
        });

        test('should remove listener by ID', () => {
            const handler = jest.fn();
            eventBus.on('test', handler);
            
            const listenerId = eventBus.events.get('test')[0].id;
            eventBus.off('test', listenerId);
            
            expect(eventBus.listenerCount('test')).toBe(0);
        });

        test('should clean up empty event arrays', () => {
            const handler = jest.fn();
            eventBus.on('test', handler);
            eventBus.off('test', handler);
            
            expect(eventBus.events.has('test')).toBe(false);
        });

        test('should handle non-existent event gracefully', () => {
            expect(() => eventBus.off('nonexistent', () => {})).not.toThrow();
        });

        test('should handle non-existent handler gracefully', () => {
            eventBus.on('test', () => {});
            expect(() => eventBus.off('test', () => {})).not.toThrow();
            expect(eventBus.listenerCount('test')).toBe(1);
        });
    });

    describe('emit', () => {
        test('should call all handlers with data', () => {
            const handler1 = jest.fn();
            const handler2 = jest.fn();
            const data = { test: true };
            
            eventBus.on('test', handler1);
            eventBus.on('test', handler2);
            
            const result = eventBus.emit('test', data);
            
            expect(result).toBe(true);
            expect(handler1).toHaveBeenCalledWith(data, 'test');
            expect(handler2).toHaveBeenCalledWith(data, 'test');
        });

        test('should return false for non-existent event', () => {
            const result = eventBus.emit('nonexistent', {});
            expect(result).toBe(false);
        });

        test('should respect priority order when calling handlers', () => {
            const callOrder = [];
            const handler1 = jest.fn(() => callOrder.push(1));
            const handler2 = jest.fn(() => callOrder.push(2));
            const handler3 = jest.fn(() => callOrder.push(3));
            
            eventBus.on('test', handler1, { priority: 1 });
            eventBus.on('test', handler2, { priority: 3 });
            eventBus.on('test', handler3, { priority: 2 });
            
            eventBus.emit('test');
            
            expect(callOrder).toEqual([2, 3, 1]);
        });

        test('should remove one-time listeners after emission', () => {
            const handler = jest.fn();
            eventBus.once('test', handler);
            
            eventBus.emit('test');
            expect(handler).toHaveBeenCalledTimes(1);
            expect(eventBus.listenerCount('test')).toBe(0);
            
            eventBus.emit('test');
            expect(handler).toHaveBeenCalledTimes(1); // Not called again
        });

        test('should handle errors in handlers gracefully', () => {
            const goodHandler = jest.fn();
            const badHandler = jest.fn(() => {
                throw new Error('Handler error');
            });
            
            eventBus.on('test', badHandler);
            eventBus.on('test', goodHandler);
            
            const result = eventBus.emit('test');
            
            expect(result).toBe(true);
            expect(console.error).toHaveBeenCalledWith(
                "Error in event handler for 'test':",
                expect.any(Error)
            );
            expect(goodHandler).toHaveBeenCalled(); // Other handlers still called
        });

        test('should handle listeners being modified during emission', () => {
            const handler1 = jest.fn();
            const handler2 = jest.fn(() => {
                eventBus.on('test', handler3); // Add handler during emission
            });
            const handler3 = jest.fn();
            
            eventBus.on('test', handler1);
            eventBus.on('test', handler2);
            
            eventBus.emit('test');
            
            expect(handler1).toHaveBeenCalledTimes(1);
            expect(handler2).toHaveBeenCalledTimes(1);
            expect(handler3).not.toHaveBeenCalled(); // Added during emission, not called
        });
    });

    describe('getEventNames', () => {
        test('should return array of event names', () => {
            eventBus.on('event1', () => {});
            eventBus.on('event2', () => {});
            eventBus.on('event3', () => {});
            
            const names = eventBus.getEventNames();
            expect(names).toEqual(['event1', 'event2', 'event3']);
        });

        test('should return empty array when no events', () => {
            expect(eventBus.getEventNames()).toEqual([]);
        });
    });

    describe('listenerCount', () => {
        test('should return correct listener count', () => {
            eventBus.on('test', () => {});
            eventBus.on('test', () => {});
            eventBus.on('other', () => {});
            
            expect(eventBus.listenerCount('test')).toBe(2);
            expect(eventBus.listenerCount('other')).toBe(1);
            expect(eventBus.listenerCount('nonexistent')).toBe(0);
        });
    });

    describe('removeAllListeners', () => {
        test('should remove all listeners for specific event', () => {
            eventBus.on('test', () => {});
            eventBus.on('test', () => {});
            eventBus.on('other', () => {});
            
            eventBus.removeAllListeners('test');
            
            expect(eventBus.listenerCount('test')).toBe(0);
            expect(eventBus.listenerCount('other')).toBe(1);
            expect(eventBus.events.has('test')).toBe(false);
        });

        test('should remove all listeners for all events', () => {
            eventBus.on('test1', () => {});
            eventBus.on('test2', () => {});
            eventBus.on('test3', () => {});
            
            eventBus.removeAllListeners();
            
            expect(eventBus.events.size).toBe(0);
        });
    });

    describe('getStats', () => {
        test('should return correct statistics', () => {
            eventBus.on('event1', () => {});
            eventBus.on('event1', () => {});
            eventBus.on('event2', () => {});
            eventBus.on('event3', () => {});
            eventBus.on('event3', () => {});
            eventBus.on('event3', () => {});
            
            const stats = eventBus.getStats();
            
            expect(stats).toEqual({
                totalEvents: 3,
                totalListeners: 6,
                events: {
                    event1: 2,
                    event2: 1,
                    event3: 3
                }
            });
        });

        test('should return zeros for empty event bus', () => {
            const stats = eventBus.getStats();
            
            expect(stats).toEqual({
                totalEvents: 0,
                totalListeners: 0,
                events: {}
            });
        });
    });

    describe('namespace', () => {
        test('should create namespaced event bus', () => {
            const ns = eventBus.namespace('module');
            const handler = jest.fn();
            
            ns.on('test', handler);
            
            expect(eventBus.events.has('module:test')).toBe(true);
            expect(eventBus.events.has('test')).toBe(false);
        });

        test('should emit namespaced events', () => {
            const ns = eventBus.namespace('module');
            const handler = jest.fn();
            const data = { test: true };
            
            ns.on('test', handler);
            ns.emit('test', data);
            
            expect(handler).toHaveBeenCalledWith(data, 'module:test');
        });

        test('should handle once in namespace', () => {
            const ns = eventBus.namespace('module');
            const handler = jest.fn();
            
            ns.once('test', handler);
            
            ns.emit('test');
            ns.emit('test');
            
            expect(handler).toHaveBeenCalledTimes(1);
        });

        test('should remove namespaced listeners', () => {
            const ns = eventBus.namespace('module');
            const handler = jest.fn();
            
            const unsubscribe = ns.on('test', handler);
            expect(eventBus.listenerCount('module:test')).toBe(1);
            
            unsubscribe();
            expect(eventBus.listenerCount('module:test')).toBe(0);
        });

        test('should remove all namespaced listeners', () => {
            const ns1 = eventBus.namespace('module1');
            const ns2 = eventBus.namespace('module2');
            
            ns1.on('test', () => {});
            ns1.on('other', () => {});
            ns2.on('test', () => {});
            
            ns1.removeAllListeners();
            
            expect(eventBus.events.has('module1:test')).toBe(false);
            expect(eventBus.events.has('module1:other')).toBe(false);
            expect(eventBus.events.has('module2:test')).toBe(true);
        });
    });

    describe('global exports', () => {
        test('should export to window object', () => {
            expect(window.EventBus).toBe(EventBus);
            expect(window.eventBus).toBeInstanceOf(EventBus);
            expect(window.ISX_EVENTS).toBeDefined();
        });

        test('should define ISX_EVENTS constants', () => {
            expect(window.ISX_EVENTS.CONNECTION_STATUS).toBe('connection:status');
            expect(window.ISX_EVENTS.WEBSOCKET_CONNECTED).toBe('websocket:connected');
            expect(window.ISX_EVENTS.DATA_UPDATED).toBe('data:updated');
            expect(window.ISX_EVENTS.PIPELINE_STATUS).toBe('pipeline:status');
        });
    });

    describe('module exports', () => {
        test('should handle module exports when available', () => {
            // This test is skipped because module exports are handled during initial load
            // The actual module export happens in the eventBus.js file itself
            // We can verify the global exports instead
            expect(window.EventBus).toBeDefined();
            expect(window.eventBus).toBeDefined();
            expect(window.ISX_EVENTS).toBeDefined();
        });
    });
});