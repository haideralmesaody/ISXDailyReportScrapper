/**
 * Frontend Logger System
 * Provides structured logging with levels, categories, and server synchronization
 */

class Logger {
    constructor() {
        this.logLevels = {
            DEBUG: 0,
            INFO: 1,
            WARN: 2,
            ERROR: 3
        };

        this.currentLevel = this.logLevels.INFO;
        this.enableServerLogging = true;
        this.logBuffer = [];
        this.maxBufferSize = 100;
        this.flushInterval = 5000; // 5 seconds
        this.categories = new Set();
        
        // Check for debug mode
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('debug') === 'true' || localStorage.getItem('ISX_DEBUG') === 'true') {
            this.currentLevel = this.logLevels.DEBUG;
        }

        // Start flush timer
        this.startFlushTimer();

        // Log system initialization
        this.info('Logger', 'Frontend logger initialized');
    }

    /**
     * Set the current log level
     */
    setLevel(level) {
        if (typeof level === 'string') {
            level = this.logLevels[level.toUpperCase()];
        }
        if (level !== undefined) {
            this.currentLevel = level;
        }
    }

    /**
     * Enable/disable specific categories
     */
    setCategories(categories) {
        this.categories = new Set(categories);
    }

    /**
     * Check if a log should be output based on level and category
     */
    shouldLog(level, category) {
        if (level < this.currentLevel) {
            return false;
        }
        
        // If no categories specified, log everything at appropriate level
        if (this.categories.size === 0) {
            return true;
        }
        
        // If categories are specified, only log matching categories
        return this.categories.has(category);
    }

    /**
     * Format log message with timestamp and context
     */
    formatMessage(level, category, message, metadata) {
        const timestamp = new Date().toISOString();
        const levelName = Object.keys(this.logLevels).find(k => this.logLevels[k] === level);
        
        let formattedMsg = `[${timestamp}] [${levelName}]`;
        if (category) {
            formattedMsg += ` [${category}]`;
        }
        formattedMsg += ` ${message}`;
        
        return {
            timestamp,
            level: levelName,
            category,
            message,
            metadata,
            formatted: formattedMsg
        };
    }

    /**
     * Core logging method
     */
    log(level, category, message, metadata = {}) {
        if (!this.shouldLog(level, category)) {
            return;
        }

        const logEntry = this.formatMessage(level, category, message, metadata);
        
        // Console output with color coding
        const consoleMethod = this.getConsoleMethod(level);
        const color = this.getLogColor(level);
        
        if (metadata && Object.keys(metadata).length > 0) {
            console.groupCollapsed(`%c${logEntry.formatted}`, `color: ${color}`);
            console.log('Metadata:', metadata);
            console.groupEnd();
        } else {
            console[consoleMethod](`%c${logEntry.formatted}`, `color: ${color}`);
        }

        // Add to buffer for server sync
        if (this.enableServerLogging && level >= this.logLevels.WARN) {
            this.bufferLog(logEntry);
        }

        // Store in localStorage for debugging
        this.storeLog(logEntry);
    }

    /**
     * Get appropriate console method for log level
     */
    getConsoleMethod(level) {
        switch (level) {
            case this.logLevels.ERROR:
                return 'error';
            case this.logLevels.WARN:
                return 'warn';
            case this.logLevels.DEBUG:
                return 'debug';
            default:
                return 'log';
        }
    }

    /**
     * Get color for log level
     */
    getLogColor(level) {
        switch (level) {
            case this.logLevels.ERROR:
                return '#ff0000';
            case this.logLevels.WARN:
                return '#ff9800';
            case this.logLevels.DEBUG:
                return '#666666';
            default:
                return '#2196f3';
        }
    }

    /**
     * Buffer log for server synchronization
     */
    bufferLog(logEntry) {
        this.logBuffer.push(logEntry);
        
        // Flush immediately for errors
        if (logEntry.level === 'ERROR') {
            this.flushLogs();
        } else if (this.logBuffer.length >= this.maxBufferSize) {
            this.flushLogs();
        }
    }

    /**
     * Store log in localStorage (rotating buffer)
     */
    storeLog(logEntry) {
        try {
            const storedLogs = JSON.parse(localStorage.getItem('ISX_LOGS') || '[]');
            storedLogs.push(logEntry);
            
            // Keep only last 500 entries
            if (storedLogs.length > 500) {
                storedLogs.splice(0, storedLogs.length - 500);
            }
            
            localStorage.setItem('ISX_LOGS', JSON.stringify(storedLogs));
        } catch (e) {
            // Ignore localStorage errors
        }
    }

    /**
     * Send buffered logs to server
     */
    async flushLogs() {
        if (this.logBuffer.length === 0) {
            return;
        }

        const logsToSend = [...this.logBuffer];
        this.logBuffer = [];

        try {
            const promises = logsToSend.map(log => 
                fetch('/api/logs', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        level: log.level,
                        component: log.category,
                        message: log.message,
                        timestamp: log.timestamp,
                        metadata: log.metadata
                    })
                })
            );

            await Promise.all(promises);
        } catch (error) {
            // Re-add logs to buffer on failure
            this.logBuffer = [...logsToSend, ...this.logBuffer];
            console.error('Failed to send logs to server:', error);
        }
    }

    /**
     * Start periodic log flushing
     */
    startFlushTimer() {
        setInterval(() => {
            this.flushLogs();
        }, this.flushInterval);

        // Flush on page unload
        window.addEventListener('beforeunload', () => {
            this.flushLogs();
        });
    }

    /**
     * Convenience methods for different log levels
     */
    debug(category, message, metadata) {
        this.log(this.logLevels.DEBUG, category, message, metadata);
    }

    info(category, message, metadata) {
        this.log(this.logLevels.INFO, category, message, metadata);
    }

    warn(category, message, metadata) {
        this.log(this.logLevels.WARN, category, message, metadata);
    }

    error(category, message, metadata) {
        this.log(this.logLevels.ERROR, category, message, metadata);
    }

    /**
     * Log timing information
     */
    timing(category, operation, duration) {
        this.debug(category, `${operation} completed in ${duration}ms`, { duration });
    }

    /**
     * Log API calls
     */
    api(method, endpoint, status, duration) {
        const level = status >= 400 ? this.logLevels.ERROR : this.logLevels.DEBUG;
        this.log(level, 'API', `${method} ${endpoint} - ${status} (${duration}ms)`, {
            method,
            endpoint,
            status,
            duration
        });
    }

    /**
     * Log WebSocket events
     */
    websocket(event, data) {
        this.debug('WebSocket', event, data);
    }

    /**
     * Export logs for debugging
     */
    exportLogs() {
        const logs = JSON.parse(localStorage.getItem('ISX_LOGS') || '[]');
        const blob = new Blob([JSON.stringify(logs, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `isx-logs-${new Date().toISOString()}.json`;
        a.click();
        URL.revokeObjectURL(url);
    }

    /**
     * Clear stored logs
     */
    clearLogs() {
        localStorage.removeItem('ISX_LOGS');
        this.logBuffer = [];
        this.info('Logger', 'Logs cleared');
    }

    /**
     * Get recent logs for display
     */
    getRecentLogs(count = 50) {
        const logs = JSON.parse(localStorage.getItem('ISX_LOGS') || '[]');
        return logs.slice(-count);
    }
}

// Export Logger class for testing
window.Logger = Logger;

// Create global logger instance
window.ISXLogger = new Logger();

// Log categories
window.LogCategory = {
    SYSTEM: 'System',
    WEBSOCKET: 'WebSocket',
    API: 'API',
    UI: 'UI',
    DATA: 'Data',
    CHART: 'Chart',
    PIPELINE: 'Pipeline',
    LICENSE: 'License',
    ERROR: 'Error'
};

// Replace console methods in production
if (window.location.hostname !== 'localhost') {
    const originalConsole = {
        log: console.log,
        debug: console.debug,
        warn: console.warn,
        error: console.error
    };

    console.log = function(...args) {
        window.ISXLogger.info('Console', args.join(' '));
    };

    console.debug = function(...args) {
        window.ISXLogger.debug('Console', args.join(' '));
    };

    console.warn = function(...args) {
        window.ISXLogger.warn('Console', args.join(' '));
    };

    console.error = function(...args) {
        window.ISXLogger.error('Console', args.join(' '));
    };
}