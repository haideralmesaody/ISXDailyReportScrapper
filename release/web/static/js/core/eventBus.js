/**
 * Event Bus for Component Communication
 * Provides decoupled event-driven communication between components
 * Follows pub/sub pattern for loose coupling
 */

class EventBus {
    constructor() {
        this.events = new Map();
        this.maxListeners = 100; // Prevent memory leaks
    }

    /**
     * Subscribe to an event
     * @param {string} event - Event name
     * @param {Function} handler - Event handler function
     * @param {object} options - Options (once, priority)
     * @returns {Function} Unsubscribe function
     */
    on(event, handler, options = {}) {
        if (typeof handler !== 'function') {
            throw new Error('Event handler must be a function');
        }

        if (!this.events.has(event)) {
            this.events.set(event, []);
        }

        const listeners = this.events.get(event);
        
        // Check max listeners limit
        if (listeners.length >= this.maxListeners) {
            console.warn(`Maximum listeners (${this.maxListeners}) reached for event: ${event}`);
            return () => {}; // Return no-op unsubscribe
        }

        const listener = {
            handler,
            once: options.once || false,
            priority: options.priority || 0,
            id: Math.random().toString(36).substr(2, 9)
        };

        listeners.push(listener);

        // Sort by priority (higher priority first)
        listeners.sort((a, b) => b.priority - a.priority);

        // Return unsubscribe function
        return () => this.off(event, listener.id);
    }

    /**
     * Subscribe to an event (one-time)
     * @param {string} event - Event name
     * @param {Function} handler - Event handler function
     * @returns {Function} Unsubscribe function
     */
    once(event, handler) {
        return this.on(event, handler, { once: true });
    }

    /**
     * Unsubscribe from an event
     * @param {string} event - Event name
     * @param {string|Function} handlerOrId - Handler function or listener ID
     */
    off(event, handlerOrId) {
        const listeners = this.events.get(event);
        if (!listeners) return;

        let index = -1;
        
        if (typeof handlerOrId === 'string') {
            // Remove by ID
            index = listeners.findIndex(listener => listener.id === handlerOrId);
        } else if (typeof handlerOrId === 'function') {
            // Remove by handler function
            index = listeners.findIndex(listener => listener.handler === handlerOrId);
        }

        if (index > -1) {
            listeners.splice(index, 1);
            
            // Clean up empty event arrays
            if (listeners.length === 0) {
                this.events.delete(event);
            }
        }
    }

    /**
     * Emit an event
     * @param {string} event - Event name
     * @param {any} data - Event data
     * @returns {boolean} True if event had listeners
     */
    emit(event, data) {
        const listeners = this.events.get(event);
        if (!listeners || listeners.length === 0) {
            return false;
        }

        // Create a copy to avoid issues if listeners are modified during emission
        const listenersToCall = [...listeners];
        
        for (const listener of listenersToCall) {
            try {
                listener.handler(data, event);
                
                // Remove one-time listeners
                if (listener.once) {
                    this.off(event, listener.id);
                }
            } catch (error) {
                console.error(`Error in event handler for '${event}':`, error);
            }
        }

        return true;
    }

    /**
     * Get all event names
     * @returns {Array<string>} Array of event names
     */
    getEventNames() {
        return Array.from(this.events.keys());
    }

    /**
     * Get listener count for an event
     * @param {string} event - Event name
     * @returns {number} Number of listeners
     */
    listenerCount(event) {
        const listeners = this.events.get(event);
        return listeners ? listeners.length : 0;
    }

    /**
     * Remove all listeners for an event or all events
     * @param {string} event - Event name (optional)
     */
    removeAllListeners(event = null) {
        if (event) {
            this.events.delete(event);
        } else {
            this.events.clear();
        }
    }

    /**
     * Get memory usage information
     * @returns {object} Memory usage stats
     */
    getStats() {
        const stats = {
            totalEvents: this.events.size,
            totalListeners: 0,
            events: {}
        };

        for (const [event, listeners] of this.events) {
            stats.totalListeners += listeners.length;
            stats.events[event] = listeners.length;
        }

        return stats;
    }

    /**
     * Create a namespaced event bus
     * @param {string} namespace - Namespace prefix
     * @returns {object} Namespaced event bus methods
     */
    namespace(namespace) {
        const prefix = `${namespace}:`;
        
        return {
            on: (event, handler, options) => this.on(prefix + event, handler, options),
            once: (event, handler) => this.once(prefix + event, handler),
            off: (event, handlerOrId) => this.off(prefix + event, handlerOrId),
            emit: (event, data) => this.emit(prefix + event, data),
            removeAllListeners: () => {
                const eventNames = this.getEventNames().filter(name => name.startsWith(prefix));
                eventNames.forEach(name => this.events.delete(name));
            }
        };
    }
}

// Create global event bus instance
const eventBus = new EventBus();

// Common ISX application events
const ISX_EVENTS = {
    // Connection events
    CONNECTION_STATUS: 'connection:status',
    WEBSOCKET_CONNECTED: 'websocket:connected',
    WEBSOCKET_DISCONNECTED: 'websocket:disconnected',
    
    // Data events
    DATA_UPDATED: 'data:updated',
    DATA_REFRESH: 'data:refresh',
    
    // UI events
    SECTION_CHANGED: 'ui:section:changed',
    TICKER_SELECTED: 'ui:ticker:selected',
    CHART_UPDATED: 'ui:chart:updated',
    
    // Pipeline events
    PIPELINE_STATUS: 'pipeline:status',
    PIPELINE_PROGRESS: 'pipeline:progress',
    PIPELINE_COMPLETE: 'pipeline:complete',
    
    // File events
    FILES_UPDATED: 'files:updated',
    FILE_DOWNLOADED: 'files:downloaded'
};

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { EventBus, eventBus, ISX_EVENTS };
}

// Global access for compatibility
window.EventBus = EventBus;
window.eventBus = eventBus;
window.ISX_EVENTS = ISX_EVENTS;