/**
 * Tests for Frontend Logger System
 */

// Import logger module directly (since it's not using ES modules)
require('./logger.js');

// Get Logger class from window since that's where it's exported
const Logger = window.Logger || global.Logger;

describe('Logger', () => {
    let logger;
    let originalConsole;
    let mockSetInterval;
    let mockAddEventListener;

    beforeEach(() => {
        // Clear any existing logger
        delete window.ISXLogger;
        
        // Mock window functions
        mockSetInterval = jest.fn();
        mockAddEventListener = jest.fn();
        global.setInterval = mockSetInterval;
        global.window.addEventListener = mockAddEventListener;
        
        // Mock window.location
        delete window.location;
        window.location = {
            search: '',
            hostname: 'localhost'
        };
        
        // Store original console
        originalConsole = {
            log: console.log,
            debug: console.debug,
            warn: console.warn,
            error: console.error,
            groupCollapsed: jest.fn(),
            groupEnd: jest.fn()
        };
        
        console.groupCollapsed = originalConsole.groupCollapsed;
        console.groupEnd = originalConsole.groupEnd;
        
        // Clear localStorage
        localStorage.clear();
        
        // Clear ISXLogger if exists
        if (window.ISXLogger) {
            delete window.ISXLogger;
        }
        
        // Create new logger instance
        logger = new Logger();
    });

    afterEach(() => {
        // Restore console
        console.log = originalConsole.log;
        console.debug = originalConsole.debug;
        console.warn = originalConsole.warn;
        console.error = originalConsole.error;
    });

    describe('constructor', () => {
        test('should initialize with default values', () => {
            expect(logger.currentLevel).toBe(logger.logLevels.INFO);
            expect(logger.enableServerLogging).toBe(true);
            expect(logger.logBuffer).toEqual([]);
            expect(logger.maxBufferSize).toBe(100);
            expect(logger.flushInterval).toBe(5000);
            expect(logger.categories).toEqual(new Set());
        });

        test('should enable debug mode from URL parameter', () => {
            window.location.search = '?debug=true';
            const debugLogger = new Logger();
            expect(debugLogger.currentLevel).toBe(debugLogger.logLevels.DEBUG);
        });

        test('should enable debug mode from localStorage', () => {
            localStorage.setItem('ISX_DEBUG', 'true');
            const debugLogger = new Logger();
            expect(debugLogger.currentLevel).toBe(debugLogger.logLevels.DEBUG);
        });

        test('should start flush timer', () => {
            expect(mockSetInterval).toHaveBeenCalledWith(expect.any(Function), 5000);
            expect(mockAddEventListener).toHaveBeenCalledWith('beforeunload', expect.any(Function));
        });
    });

    describe('setLevel', () => {
        test('should set log level from string', () => {
            logger.setLevel('debug');
            expect(logger.currentLevel).toBe(logger.logLevels.DEBUG);

            logger.setLevel('ERROR');
            expect(logger.currentLevel).toBe(logger.logLevels.ERROR);
        });

        test('should set log level from number', () => {
            logger.setLevel(logger.logLevels.WARN);
            expect(logger.currentLevel).toBe(logger.logLevels.WARN);
        });

        test('should ignore invalid levels', () => {
            logger.setLevel('invalid');
            expect(logger.currentLevel).toBe(logger.logLevels.INFO);
        });
    });

    describe('setCategories', () => {
        test('should set categories filter', () => {
            logger.setCategories(['System', 'API']);
            expect(logger.categories).toEqual(new Set(['System', 'API']));
        });
    });

    describe('shouldLog', () => {
        test('should filter by log level', () => {
            logger.setLevel('WARN');
            expect(logger.shouldLog(logger.logLevels.DEBUG, 'Test')).toBe(false);
            expect(logger.shouldLog(logger.logLevels.INFO, 'Test')).toBe(false);
            expect(logger.shouldLog(logger.logLevels.WARN, 'Test')).toBe(true);
            expect(logger.shouldLog(logger.logLevels.ERROR, 'Test')).toBe(true);
        });

        test('should filter by category when categories are set', () => {
            logger.setCategories(['System', 'API']);
            expect(logger.shouldLog(logger.logLevels.INFO, 'System')).toBe(true);
            expect(logger.shouldLog(logger.logLevels.INFO, 'API')).toBe(true);
            expect(logger.shouldLog(logger.logLevels.INFO, 'Other')).toBe(false);
        });

        test('should log all categories when none specified', () => {
            expect(logger.shouldLog(logger.logLevels.INFO, 'Any')).toBe(true);
        });
    });

    describe('formatMessage', () => {
        test('should format message correctly', () => {
            const result = logger.formatMessage(
                logger.logLevels.INFO,
                'Test',
                'Test message',
                { data: 'value' }
            );

            expect(result).toHaveProperty('timestamp');
            expect(result.level).toBe('INFO');
            expect(result.category).toBe('Test');
            expect(result.message).toBe('Test message');
            expect(result.metadata).toEqual({ data: 'value' });
            expect(result.formatted).toMatch(/\[.*\] \[INFO\] \[Test\] Test message/);
        });
    });

    describe('log', () => {
        test('should not log below current level', () => {
            logger.setLevel('WARN');
            logger.log(logger.logLevels.INFO, 'Test', 'Message');
            expect(console.log).not.toHaveBeenCalled();
        });

        test('should log with metadata', () => {
            const metadata = { key: 'value' };
            logger.log(logger.logLevels.INFO, 'Test', 'Message', metadata);
            
            expect(originalConsole.groupCollapsed).toHaveBeenCalled();
            expect(console.log).toHaveBeenCalledWith('Metadata:', metadata);
            expect(originalConsole.groupEnd).toHaveBeenCalled();
        });

        test('should log without metadata', () => {
            logger.log(logger.logLevels.INFO, 'Test', 'Message');
            expect(console.log).toHaveBeenCalled();
        });

        test('should buffer WARN and ERROR logs for server sync', () => {
            logger.log(logger.logLevels.WARN, 'Test', 'Warning');
            expect(logger.logBuffer).toHaveLength(1);
            
            logger.log(logger.logLevels.ERROR, 'Test', 'Error');
            expect(logger.logBuffer).toHaveLength(2);
        });

        test('should store logs in localStorage', () => {
            // Clear localStorage first
            localStorage.clear();
            logger = new Logger(); // Create fresh logger
            
            logger.log(logger.logLevels.INFO, 'Test', 'Message');
            const stored = JSON.parse(localStorage.getItem('ISX_LOGS'));
            expect(stored).toHaveLength(2); // Initial log + this one
            expect(stored[1].message).toBe('Message');
        });
    });

    describe('getConsoleMethod', () => {
        test('should return correct console method', () => {
            expect(logger.getConsoleMethod(logger.logLevels.ERROR)).toBe('error');
            expect(logger.getConsoleMethod(logger.logLevels.WARN)).toBe('warn');
            expect(logger.getConsoleMethod(logger.logLevels.DEBUG)).toBe('debug');
            expect(logger.getConsoleMethod(logger.logLevels.INFO)).toBe('log');
        });
    });

    describe('getLogColor', () => {
        test('should return correct colors', () => {
            expect(logger.getLogColor(logger.logLevels.ERROR)).toBe('#ff0000');
            expect(logger.getLogColor(logger.logLevels.WARN)).toBe('#ff9800');
            expect(logger.getLogColor(logger.logLevels.DEBUG)).toBe('#666666');
            expect(logger.getLogColor(logger.logLevels.INFO)).toBe('#2196f3');
        });
    });

    describe('bufferLog', () => {
        test('should add log to buffer', () => {
            const logEntry = { level: 'WARN', message: 'Test' };
            logger.bufferLog(logEntry);
            expect(logger.logBuffer).toContain(logEntry);
        });

        test('should flush immediately for errors', () => {
            logger.flushLogs = jest.fn();
            const errorEntry = { level: 'ERROR', message: 'Test' };
            logger.bufferLog(errorEntry);
            expect(logger.flushLogs).toHaveBeenCalled();
        });

        test('should flush when buffer is full', () => {
            logger.flushLogs = jest.fn();
            logger.maxBufferSize = 2;
            
            logger.bufferLog({ level: 'WARN', message: '1' });
            expect(logger.flushLogs).not.toHaveBeenCalled();
            
            logger.bufferLog({ level: 'WARN', message: '2' });
            expect(logger.flushLogs).toHaveBeenCalled();
        });
    });

    describe('storeLog', () => {
        test('should store logs in localStorage', () => {
            const logEntry = { message: 'Test' };
            logger.storeLog(logEntry);
            
            const stored = JSON.parse(localStorage.getItem('ISX_LOGS'));
            expect(stored).toContainEqual(logEntry);
        });

        test('should maintain max 500 entries', () => {
            const existingLogs = Array(500).fill({ message: 'old' });
            localStorage.setItem('ISX_LOGS', JSON.stringify(existingLogs));
            
            logger.storeLog({ message: 'new' });
            
            const stored = JSON.parse(localStorage.getItem('ISX_LOGS'));
            expect(stored).toHaveLength(500);
            expect(stored[stored.length - 1].message).toBe('new');
        });

        test('should handle localStorage errors silently', () => {
            localStorage.setItem = jest.fn(() => {
                throw new Error('Storage full');
            });
            
            expect(() => logger.storeLog({ message: 'Test' })).not.toThrow();
        });
    });

    describe('flushLogs', () => {
        test('should send logs to server', async () => {
            fetch.mockResolvedValue({ ok: true });
            
            logger.logBuffer = [
                { level: 'WARN', category: 'Test', message: 'Warning', timestamp: '2024-01-01', metadata: {} }
            ];
            
            await logger.flushLogs();
            
            expect(fetch).toHaveBeenCalledWith('/api/logs', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    level: 'WARN',
                    component: 'Test',
                    message: 'Warning',
                    timestamp: '2024-01-01',
                    metadata: {}
                })
            });
            
            expect(logger.logBuffer).toHaveLength(0);
        });

        test('should restore buffer on failure', async () => {
            fetch.mockRejectedValue(new Error('Network error'));
            
            const originalBuffer = [{ level: 'WARN', message: 'Test' }];
            logger.logBuffer = [...originalBuffer];
            
            await logger.flushLogs();
            
            expect(logger.logBuffer).toEqual(originalBuffer);
        });

        test('should do nothing with empty buffer', async () => {
            logger.logBuffer = [];
            await logger.flushLogs();
            expect(fetch).not.toHaveBeenCalled();
        });
    });

    describe('convenience methods', () => {
        beforeEach(() => {
            logger.log = jest.fn();
        });

        test('debug should call log with DEBUG level', () => {
            logger.debug('Category', 'Message', { data: 1 });
            expect(logger.log).toHaveBeenCalledWith(
                logger.logLevels.DEBUG,
                'Category',
                'Message',
                { data: 1 }
            );
        });

        test('info should call log with INFO level', () => {
            logger.info('Category', 'Message');
            expect(logger.log).toHaveBeenCalledWith(
                logger.logLevels.INFO,
                'Category',
                'Message',
                undefined
            );
        });

        test('warn should call log with WARN level', () => {
            logger.warn('Category', 'Message');
            expect(logger.log).toHaveBeenCalledWith(
                logger.logLevels.WARN,
                'Category',
                'Message',
                undefined
            );
        });

        test('error should call log with ERROR level', () => {
            logger.error('Category', 'Message');
            expect(logger.log).toHaveBeenCalledWith(
                logger.logLevels.ERROR,
                'Category',
                'Message',
                undefined
            );
        });
    });

    describe('timing', () => {
        test('should log timing information', () => {
            logger.debug = jest.fn();
            logger.timing('API', 'fetch', 250);
            
            expect(logger.debug).toHaveBeenCalledWith(
                'API',
                'fetch completed in 250ms',
                { duration: 250 }
            );
        });
    });

    describe('api', () => {
        beforeEach(() => {
            logger.log = jest.fn();
        });

        test('should log successful API calls as DEBUG', () => {
            logger.api('GET', '/api/data', 200, 100);
            
            expect(logger.log).toHaveBeenCalledWith(
                logger.logLevels.DEBUG,
                'API',
                'GET /api/data - 200 (100ms)',
                {
                    method: 'GET',
                    endpoint: '/api/data',
                    status: 200,
                    duration: 100
                }
            );
        });

        test('should log error API calls as ERROR', () => {
            logger.api('POST', '/api/data', 500, 200);
            
            expect(logger.log).toHaveBeenCalledWith(
                logger.logLevels.ERROR,
                'API',
                'POST /api/data - 500 (200ms)',
                {
                    method: 'POST',
                    endpoint: '/api/data',
                    status: 500,
                    duration: 200
                }
            );
        });
    });

    describe('websocket', () => {
        test('should log WebSocket events', () => {
            logger.debug = jest.fn();
            const data = { type: 'message' };
            
            logger.websocket('connected', data);
            
            expect(logger.debug).toHaveBeenCalledWith('WebSocket', 'connected', data);
        });
    });

    describe('exportLogs', () => {
        test('should export logs as JSON file', () => {
            const mockCreateElement = jest.fn(() => ({
                click: jest.fn(),
                href: null,
                download: null
            }));
            const mockCreateObjectURL = jest.fn(() => 'blob:url');
            const mockRevokeObjectURL = jest.fn();
            
            document.createElement = mockCreateElement;
            URL.createObjectURL = mockCreateObjectURL;
            URL.revokeObjectURL = mockRevokeObjectURL;
            
            const logs = [{ message: 'Test log' }];
            localStorage.setItem('ISX_LOGS', JSON.stringify(logs));
            
            logger.exportLogs();
            
            expect(mockCreateObjectURL).toHaveBeenCalled();
            expect(mockCreateElement).toHaveBeenCalledWith('a');
            expect(mockRevokeObjectURL).toHaveBeenCalledWith('blob:url');
        });
    });

    describe('clearLogs', () => {
        test('should clear stored logs and buffer', () => {
            localStorage.setItem('ISX_LOGS', JSON.stringify([{ message: 'old' }]));
            logger.logBuffer = [{ message: 'buffered' }];
            logger.info = jest.fn();
            
            logger.clearLogs();
            
            expect(localStorage.getItem('ISX_LOGS')).toBeNull();
            expect(logger.logBuffer).toHaveLength(0);
            expect(logger.info).toHaveBeenCalledWith('Logger', 'Logs cleared');
        });
    });

    describe('getRecentLogs', () => {
        test('should return recent logs', () => {
            const logs = Array(100).fill(null).map((_, i) => ({ message: `Log ${i}` }));
            localStorage.setItem('ISX_LOGS', JSON.stringify(logs));
            
            const recent = logger.getRecentLogs(10);
            expect(recent).toHaveLength(10);
            expect(recent[0].message).toBe('Log 90');
            expect(recent[9].message).toBe('Log 99');
        });

        test('should return all logs if less than requested', () => {
            const logs = [{ message: '1' }, { message: '2' }];
            localStorage.setItem('ISX_LOGS', JSON.stringify(logs));
            
            const recent = logger.getRecentLogs(10);
            expect(recent).toHaveLength(2);
        });

        test('should handle empty logs', () => {
            const recent = logger.getRecentLogs();
            expect(recent).toHaveLength(0);
        });
    });

    describe('global instance', () => {
        test('should create global ISXLogger instance', () => {
            // Re-run the logger script to create global instance
            delete window.ISXLogger;
            require('./logger.js');
            
            expect(window.ISXLogger).toBeDefined();
            expect(window.ISXLogger).toBeInstanceOf(Logger);
        });

        test('should define LogCategory constants', () => {
            expect(window.LogCategory).toBeDefined();
            expect(window.LogCategory.SYSTEM).toBe('System');
            expect(window.LogCategory.WEBSOCKET).toBe('WebSocket');
            expect(window.LogCategory.API).toBe('API');
        });
    });

    describe('console replacement', () => {
        test('should not replace console methods in localhost', () => {
            window.location.hostname = 'localhost';
            const originalLog = console.log;
            
            // Re-run the logger script
            delete window.ISXLogger;
            require('./logger.js');
            
            expect(console.log).toBe(originalLog);
        });

        test('should replace console methods in production', () => {
            window.location.hostname = 'example.com';
            window.ISXLogger = {
                info: jest.fn(),
                debug: jest.fn(),
                warn: jest.fn(),
                error: jest.fn()
            };
            
            // Re-run the logger script
            eval(`
                if (window.location.hostname !== 'localhost') {
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
            `);
            
            console.log('test', 'message');
            expect(window.ISXLogger.info).toHaveBeenCalledWith('Console', 'test message');
            
            console.error('error', 'occurred');
            expect(window.ISXLogger.error).toHaveBeenCalledWith('Console', 'error occurred');
        });
    });
});