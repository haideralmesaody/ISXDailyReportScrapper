/**
 * Data Update Manager Component
 * Handles real-time data updates from WebSocket messages
 * Manages auto-refresh logic for different data types
 */

class DataUpdateManager {
    constructor(eventBus, uiState) {
        this.eventBus = eventBus;
        this.uiState = uiState;
        this.handlers = new Map();
        this.refreshTimeouts = new Map();
        this.registerHandlers();
    }

    /**
     * Register handlers for different data update types
     */
    registerHandlers() {
        // Register handler for ticker summary updates
        this.handlers.set('ticker_summary', () => {
            this.eventBus.emit(ISX_EVENTS.DATA_UPDATED, { type: 'tickers' });
            
            // Refresh market movers if section is active
            if (this.uiState.getState('sections.marketmovers.visible')) {
                this.eventBus.emit(ISX_EVENTS.DATA_REFRESH, { component: 'marketmovers' });
            }
            
            this.logUpdate('Ticker summary and market movers updated automatically');
        });

        // Register handler for combined data updates
        this.handlers.set('combined_data', () => {
            this.eventBus.emit(ISX_EVENTS.DATA_UPDATED, { type: 'combined_data' });
            this.eventBus.emit(ISX_EVENTS.FILES_UPDATED);
            this.logUpdate('Combined data updated automatically');
        });

        // Register handler for indexes.csv updates
        this.handlers.set('indexes', () => {
            this.eventBus.emit(ISX_EVENTS.DATA_UPDATED, { type: 'indexes' });
            
            if (this.uiState.getState('sections.indexcsv.visible')) {
                this.eventBus.emit(ISX_EVENTS.CHART_UPDATED, { type: 'index' });
            }
            
            this.logUpdate('Market indices data updated automatically');
        });

        // Register handler for daily report updates
        this.handlers.set('daily_report', (data) => {
            this.eventBus.emit(ISX_EVENTS.FILES_UPDATED);
            this.logUpdate(`Daily report updated: ${data.filename}`);
        });

        // Register handler for index data updates
        this.handlers.set('index_data', () => {
            this.eventBus.emit(ISX_EVENTS.DATA_UPDATED, { type: 'index_data' });
            this.logUpdate('Index data updated automatically');
        });
    }

    /**
     * Handle incoming update message
     * @param {object} message - WebSocket message
     */
    handleUpdate(message) {
        if (message.type === 'data_update') {
            if (message.subtype === 'all' && message.action === 'refresh') {
                this.handleFullRefresh(message);
            } else {
                this.handleSpecificUpdate(message);
            }
        }
    }

    /**
     * Handle full application refresh
     * @param {object} message - Refresh message
     */
    handleFullRefresh(message) {
        const components = message.data?.components || ['files', 'tickers', 'charts'];
        this.logUpdate(`🔄 Auto-refreshing components: ${components.join(', ')}`);

        // Clear any pending refresh timeouts
        this.clearRefreshTimeouts();

        // Schedule refresh with delay to avoid overwhelming the UI
        this.refreshTimeouts.set('full_refresh', setTimeout(() => {
            const refreshPromises = [];

            if (components.includes('files')) {
                refreshPromises.push(this.refreshFiles());
            }

            if (components.includes('tickers')) {
                refreshPromises.push(this.refreshTickers());
                
                // Also refresh market movers if section is active
                if (this.uiState.getState('sections.marketmovers.visible')) {
                    refreshPromises.push(this.refreshMarketMovers());
                }
            }

            if (components.includes('charts')) {
                refreshPromises.push(this.refreshCharts());
            }

            // Wait for all refreshes to complete
            Promise.allSettled(refreshPromises).then(() => {
                this.logUpdate('✅ Components refreshed with latest data!', 'success');
                this.eventBus.emit(ISX_EVENTS.DATA_REFRESH, { type: 'complete' });
            });
        }, 500));
    }

    /**
     * Handle specific data type update
     * @param {object} message - Update message
     */
    handleSpecificUpdate(message) {
        const handler = this.handlers.get(message.subtype);
        if (handler) {
            handler(message.data);
        } else {
            console.warn('No handler for data update subtype:', message.subtype);
        }
    }

    /**
     * Refresh files component
     * @returns {Promise} Refresh promise
     */
    async refreshFiles() {
        return new Promise((resolve) => {
            this.eventBus.emit(ISX_EVENTS.FILES_UPDATED);
            resolve();
        });
    }

    /**
     * Refresh tickers component
     * @returns {Promise} Refresh promise
     */
    async refreshTickers() {
        return new Promise((resolve) => {
            this.eventBus.emit(ISX_EVENTS.DATA_UPDATED, { type: 'tickers' });
            resolve();
        });
    }

    /**
     * Refresh market movers component
     * @returns {Promise} Refresh promise
     */
    async refreshMarketMovers() {
        return new Promise((resolve) => {
            this.eventBus.emit(ISX_EVENTS.DATA_REFRESH, { component: 'marketmovers' });
            resolve();
        });
    }

    /**
     * Refresh charts
     * @returns {Promise} Refresh promise
     */
    async refreshCharts() {
        return new Promise((resolve) => {
            const activeSection = this.uiState.getActiveSection();
            
            if (activeSection === 'dashboard') {
                this.eventBus.emit(ISX_EVENTS.CHART_UPDATED, { type: 'index' });
            }
            
            if (activeSection === 'tickercharts') {
                const currentTicker = this.uiState.getSelection('ticker');
                if (currentTicker) {
                    this.eventBus.emit(ISX_EVENTS.TICKER_SELECTED, { ticker: currentTicker });
                }
            }
            
            resolve();
        });
    }

    /**
     * Clear all refresh timeouts
     */
    clearRefreshTimeouts() {
        for (const [key, timeout] of this.refreshTimeouts) {
            clearTimeout(timeout);
        }
        this.refreshTimeouts.clear();
    }

    /**
     * Log update message
     * @param {string} message - Log message
     * @param {string} type - Log type (info, success, warning, error)
     */
    logUpdate(message, type = 'info') {
        this.eventBus.emit('log:output', { message, type });
    }

    /**
     * Register custom data handler
     * @param {string} type - Data type
     * @param {Function} handler - Handler function
     */
    registerHandler(type, handler) {
        this.handlers.set(type, handler);
    }

    /**
     * Unregister data handler
     * @param {string} type - Data type
     */
    unregisterHandler(type) {
        this.handlers.delete(type);
    }

    /**
     * Get registered handlers
     * @returns {Array<string>} Handler types
     */
    getHandlerTypes() {
        return Array.from(this.handlers.keys());
    }

    /**
     * Destroy the manager and clean up resources
     */
    destroy() {
        this.clearRefreshTimeouts();
        this.handlers.clear();
    }
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = DataUpdateManager;
}

// Global access for compatibility
window.DataUpdateManager = DataUpdateManager;