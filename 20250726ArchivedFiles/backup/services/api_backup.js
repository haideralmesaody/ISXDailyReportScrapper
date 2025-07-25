/**
 * API Service Module
 * Handles all HTTP API calls to the backend
 * Provides consistent error handling and response formatting
 */

// Polyfill for AbortSignal.timeout() in older browsers
if (!AbortSignal.timeout) {
    AbortSignal.timeout = (ms) => {
        const controller = new AbortController();
        setTimeout(() => controller.abort(new DOMException('Timeout', 'TimeoutError')), ms);
        return controller.signal;
    };
}

// Custom error classes for better error handling
class APIServiceError extends Error {
    constructor(message, code = null) {
        super(message);
        this.name = 'APIServiceError';
        this.code = code;
    }
}

class TimeoutError extends APIServiceError {
    constructor() {
        super('Request timeout', 'TIMEOUT');
        this.name = 'TimeoutError';
    }
}

class HTTPError extends APIServiceError {
    constructor(status, message, response = null) {
        super(message, 'HTTP_ERROR');
        this.name = 'HTTPError';
        this.status = status;
        this.response = response;
    }
}

class NetworkError extends APIServiceError {
    constructor(message) {
        super(message, 'NETWORK_ERROR');
        this.name = 'NetworkError';
    }
}

class APIService {
    constructor(options = {}) {
        this.baseURL = options.baseURL || this._getBaseURL();
        this.requestTimeout = options.timeout || 30000;
        this.getAuth = options.getAuth || (() => null);
        this.defaultHeaders = {};
        this.csrfToken = null;
    }

    _getBaseURL() {
        return window.APP_CONFIG?.API_BASE_URL || '';
    }

    _getCSRFToken() {
        return document.querySelector('meta[name="csrf-token"]')?.content || 
               this.csrfToken || 
               localStorage.getItem('csrfToken');
    }

    _getDebugMode() {
        return window.APP_CONFIG?.DEBUG || 
               location.hostname === 'localhost' || 
               location.search.includes('debug=true');
    }

    /**
     * Make HTTP request
     * @param {string} method - HTTP method
     * @param {string} url - Request URL
     * @param {object} options - Request options
     * @returns {Promise} Request promise
     */
    async request(method, url, options = {}) {
        const config = {
            method: method.toUpperCase(),
            headers: { ...this.defaultHeaders, ...options.headers },
            signal: AbortSignal.timeout(this.requestTimeout)
        };

        if (options.body) {
            if (typeof options.body === 'object') {
                config.body = JSON.stringify(options.body);
            } else {
                config.body = options.body;
            }
        }

        try {
            console.log(`Fetching: ${this.baseURL + url}`, config);
            const response = await fetch(this.baseURL + url, config);
            console.log(`Response status: ${response.status}`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                const jsonResponse = await response.json();
                console.log('JSON Response:', jsonResponse);
                return jsonResponse;
            } else {
                return await response.text();
            }
        } catch (error) {
            console.error('Request error:', error);
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            throw error;
        }
    }

    /**
     * GET request
     * @param {string} url - Request URL
     * @param {object} options - Request options
     * @returns {Promise} Request promise
     */
    async get(url, options = {}) {
        return this.request('GET', url, options);
    }

    /**
     * POST request
     * @param {string} url - Request URL
     * @param {object} body - Request body
     * @param {object} options - Request options
     * @returns {Promise} Request promise
     */
    async post(url, body = null, options = {}) {
        console.log(`POST request to ${url} with body:`, body);
        return this.request('POST', url, { ...options, body });
    }

    /**
     * PUT request
     * @param {string} url - Request URL
     * @param {object} body - Request body
     * @param {object} options - Request options
     * @returns {Promise} Request promise
     */
    async put(url, body = null, options = {}) {
        return this.request('PUT', url, { ...options, body });
    }

    /**
     * DELETE request
     * @param {string} url - Request URL
     * @param {object} options - Request options
     * @returns {Promise} Request promise
     */
    async delete(url, options = {}) {
        return this.request('DELETE', url, options);
    }

    /**
     * Get license status
     * @returns {Promise<object>} License status
     */
    async getLicenseStatus() {
        try {
            return await this.get('/api/license/status');
        } catch (error) {
            console.error('Failed to get license status:', error);
            
            // Report API error to server
            this.reportClientError('API_ERROR', `getLicenseStatus failed: ${error.message}`, {
                url: '/api/license/status',
                method: 'GET',
                error: error.name,
                stack: error.stack
            });
            
            return { status: 'error', message: error.message };
        }
    }

    /**
     * Report client-side errors to server for logging
     * @param {string} type - Error type
     * @param {string} message - Error message  
     * @param {object} context - Additional context
     */
    async reportClientError(type, message, context = {}) {
        try {
            await this.post('/api/logs', {
                level: 'error',
                component: 'CLIENT',
                message: `[${type}] ${message}`,
                timestamp: new Date().toISOString(),
                metadata: {
                    userAgent: navigator.userAgent,
                    url: window.location.href,
                    ...context
                }
            });
        } catch (error) {
            // Don't log errors about logging errors to avoid infinite loops
            console.warn('Failed to report client error to server:', error);
        }
    }

    /**
     * Activate license
     * @param {string} licenseKey - License key to activate
     * @returns {Promise<object>} Activation result
     */
    async activateLicense(licenseKey) {
        try {
            return await this.post('/api/license/activate', { key: licenseKey });
        } catch (error) {
            console.error('Failed to activate license:', error);
            return { success: false, message: error.message };
        }
    }

    /**
     * Get version information
     * @returns {Promise<object>} Version info
     */
    async getVersionInfo() {
        try {
            return await this.get('/api/version');
        } catch (error) {
            console.error('Failed to get version info:', error);
            return { version: 'unknown', build: 'unknown' };
        }
    }

    /**
     * Get available files
     * @returns {Promise<object>} Files data
     */
    async getFiles() {
        try {
            return await this.get('/api/files');
        } catch (error) {
            console.error('Failed to get files:', error);
            throw error;
        }
    }

    /**
     * Get ticker data
     * @returns {Promise<Array>} Ticker data
     */
    async getTickers() {
        try {
            return await this.get('/api/tickers');
        } catch (error) {
            console.error('Failed to get tickers:', error);
            throw error;
        }
    }

    /**
     * Get ticker chart data
     * @param {string} ticker - Ticker symbol
     * @returns {Promise<object>} Chart data
     */
    async getTickerChart(ticker) {
        try {
            return await this.get(`/api/ticker/${encodeURIComponent(ticker)}/chart`);
        } catch (error) {
            console.error(`Failed to get chart data for ${ticker}:`, error);
            throw error;
        }
    }

    /**
     * Get market movers data
     * @param {object} options - Query options
     * @returns {Promise<object>} Market movers data
     */
    async getMarketMovers(options = {}) {
        const params = new URLSearchParams();
        
        if (options.period) params.append('period', options.period);
        if (options.limit) params.append('limit', options.limit);
        if (options.min_volume) params.append('min_volume', options.min_volume);

        const queryString = params.toString();
        const url = `/api/gainers-losers${queryString ? '?' + queryString : ''}`;

        try {
            return await this.get(url);
        } catch (error) {
            console.error('Failed to get market movers:', error);
            throw error;
        }
    }

    /**
     * Get index data
     * @returns {Promise<object>} Index data
     */
    async getIndexData() {
        try {
            return await this.get('/api/indexes');
        } catch (error) {
            console.error('Failed to get index data:', error);
            throw error;
        }
    }

    /**
     * Start scraping operation
     * @param {object} params - Scraping parameters
     * @returns {Promise<object>} Operation result
     */
async startScraping(params) {
    try {
        console.log('Starting scraping with params:', params);
        const result = await this.request('POST', '/api/scrape', { body: params }); // ✅ Uses base timeout
        console.log('Scraping API response:', result);
        return result;
    } catch (error) {
        console.error('Failed to start scraping:', error);
        throw error;
    }
}

    /**
     * Start processing operation
     * @returns {Promise<object>} Operation result
     */
    async startProcessing() {
        try {
            return await this.post('/api/process', {
                command: 'process',
                args: { in: 'data/downloads' }
            });
        } catch (error) {
            console.error('Failed to start processing:', error);
            throw error;
        }
    }

    /**
     * Start index extraction
     * @returns {Promise<object>} Operation result
     */
    async startIndexExtraction() {
        try {
            return await this.post('/api/indexcsv', {
                command: 'indexcsv',
                args: { dir: 'data/downloads', out: 'data/reports/indexes.csv' }
            });
        } catch (error) {
            console.error('Failed to start index extraction:', error);
            throw error;
        }
    }

    /**
     * Download file
     * @param {string} filename - File name
     * @param {string} type - File type
     * @returns {Promise<Blob>} File blob
     */
    async downloadFile(filename, type) {
        try {
            const url = `/api/download/${encodeURIComponent(type)}/${encodeURIComponent(filename)}`;
            const response = await fetch(url);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            return await response.blob();
        } catch (error) {
            console.error(`Failed to download file ${filename}:`, error);
            throw error;
        }
    }

    /**
     * Get system stats
     * @returns {Promise<object>} System statistics
     */
    async getSystemStats() {
        try {
            return await this.get('/api/stats');
        } catch (error) {
            console.error('Failed to get system stats:', error);
            return {};
        }
    }

    /**
     * Set request timeout
     * @param {number} timeout - Timeout in milliseconds
     */
    setTimeout(timeout) {
        this.requestTimeout = timeout;
    }

    /**
     * Set default headers
     * @param {object} headers - Headers object
     */
    setHeaders(headers) {
        this.defaultHeaders = { ...this.defaultHeaders, ...headers };
    }

    /**
     * Clear default headers
     */
    clearHeaders() {
        this.defaultHeaders = {
            'Content-Type': 'application/json'
        };
    }
}

// Create global API service instance
const apiService = new APIService();

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { APIService, apiService };
}

// Global access for compatibility
window.APIService = APIService;
window.apiService = apiService;