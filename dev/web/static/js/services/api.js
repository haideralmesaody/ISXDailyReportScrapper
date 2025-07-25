/**
 * Enhanced API Service Module
 * Handles all HTTP API calls with RFC 7807 compliant error handling
 * Provides request/response interceptors, retry logic, and observability
 */

// Custom error classes
class APIError extends Error {
    constructor(type, title, status, detail, instance, extensions = {}) {
        super(title);
        this.name = 'APIError';
        this.type = type;
        this.title = title;
        this.status = status;
        this.detail = detail;
        this.instance = instance;
        this.extensions = extensions;
        this.timestamp = new Date();
    }

    static fromProblemDetails(problem) {
        const error = new APIError(
            problem.type || '/errors/unknown',
            problem.title || 'Unknown Error',
            problem.status || 500,
            problem.detail || '',
            problem.instance || ''
        );
        
        // Add all other fields as extensions
        Object.keys(problem).forEach(key => {
            if (!['type', 'title', 'status', 'detail', 'instance'].includes(key)) {
                error.extensions[key] = problem[key];
            }
        });
        
        return error;
    }

    toJSON() {
        return {
            type: this.type,
            title: this.title,
            status: this.status,
            detail: this.detail,
            instance: this.instance,
            ...this.extensions,
            timestamp: this.timestamp.toISOString()
        };
    }
}

class NetworkError extends APIError {
    constructor(message, originalError) {
        super(
            '/errors/network',
            'Network Error',
            0,
            message || 'Unable to connect to server',
            window.location.pathname
        );
        this.originalError = originalError;
    }
}

class TimeoutError extends APIError {
    constructor(url, timeout) {
        super(
            '/errors/timeout',
            'Request Timeout',
            0,
            `Request to ${url} timed out after ${timeout}ms`,
            url
        );
    }
}

class ValidationError extends APIError {
    constructor(errors) {
        super(
            '/errors/validation',
            'Validation Failed',
            400,
            'One or more fields failed validation',
            window.location.pathname,
            { errors }
        );
    }
}

// Request/Response interceptor types
class RequestInterceptor {
    constructor(onFulfilled, onRejected) {
        this.onFulfilled = onFulfilled;
        this.onRejected = onRejected;
    }
}

class ResponseInterceptor {
    constructor(onFulfilled, onRejected) {
        this.onFulfilled = onFulfilled;
        this.onRejected = onRejected;
    }
}

// Main API Service class
class APIService {
    constructor(config = {}) {
        this.config = {
            baseURL: config.baseURL || window.APP_CONFIG?.API_BASE_URL || '',
            timeout: config.timeout || 30000,
            retries: config.retries || 3,
            retryDelay: config.retryDelay || 1000,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                ...config.headers
            }
        };

        this.requestInterceptors = [];
        this.responseInterceptors = [];
        
        // Add default interceptors
        this.setupDefaultInterceptors();
        
        // Request tracking
        this.activeRequests = new Map();
        this.requestIdCounter = 0;
        
        // Event emitter for observability
        this.eventTarget = new EventTarget();

        // Legacy compatibility
        this.requestTimeout = this.config.timeout;
        this.defaultHeaders = this.config.headers;
    }

    setupDefaultInterceptors() {
        // Add request ID
        this.addRequestInterceptor(config => {
            const requestId = `req-${++this.requestIdCounter}`;
            config.headers['X-Request-ID'] = requestId;
            config.requestId = requestId;
            
            // Emit request start event
            this.emit('request:start', {
                requestId,
                method: config.method,
                url: config.url,
                timestamp: new Date()
            });
            
            return config;
        });

        // Add CSRF token if available
        this.addRequestInterceptor(config => {
            const csrfToken = this.getCSRFToken();
            if (csrfToken) {
                config.headers['X-CSRF-Token'] = csrfToken;
            }
            return config;
        });

        // Log responses in debug mode
        this.addResponseInterceptor(
            response => {
                if (this.isDebugMode()) {
                    console.log(`[API] ${response.config.method} ${response.config.url}`, {
                        status: response.status,
                        data: response.data,
                        headers: response.headers
                    });
                }
                
                // Emit response event
                this.emit('request:success', {
                    requestId: response.config.requestId,
                    status: response.status,
                    duration: Date.now() - response.config.startTime
                });
                
                return response;
            },
            error => {
                // Emit error event
                this.emit('request:error', {
                    requestId: error.config?.requestId,
                    error: error.toJSON()
                });
                
                throw error;
            }
        );
    }

    // Interceptor management
    addRequestInterceptor(onFulfilled, onRejected) {
        this.requestInterceptors.push(new RequestInterceptor(onFulfilled, onRejected));
        return this.requestInterceptors.length - 1;
    }

    addResponseInterceptor(onFulfilled, onRejected) {
        this.responseInterceptors.push(new ResponseInterceptor(onFulfilled, onRejected));
        return this.responseInterceptors.length - 1;
    }

    removeRequestInterceptor(index) {
        this.requestInterceptors.splice(index, 1);
    }

    removeResponseInterceptor(index) {
        this.responseInterceptors.splice(index, 1);
    }

    // Core request method
    async request(method, url, options = {}) {
        // If baseURL is not set, try to get it from APP_CONFIG or use origin as fallback
        if (!this.config.baseURL) {
            this.config.baseURL = window.APP_CONFIG?.API_BASE_URL || window.location.origin;
        }
        
        // Build config object
        let config = {
            method: method.toUpperCase(),
            url,
            baseURL: this.config.baseURL || window.location.origin,
            timeout: options.timeout || this.config.timeout,
            headers: {
                ...this.config.headers,
                ...options.headers
            },
            data: options.body,
            retries: options.retries !== undefined ? options.retries : this.config.retries,
            retryDelay: options.retryDelay || this.config.retryDelay,
            startTime: Date.now()
        };

        // Apply request interceptors
        for (const interceptor of this.requestInterceptors) {
            try {
                config = await (interceptor.onFulfilled ? interceptor.onFulfilled(config) : config);
            } catch (error) {
                if (interceptor.onRejected) {
                    config = await interceptor.onRejected(error);
                } else {
                    throw error;
                }
            }
        }

        // Create abort controller
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), config.timeout);

        // Track active request
        this.activeRequests.set(config.requestId, { controller, config });

        try {
            // Build full URL
            const fullUrl = new URL(config.url, config.baseURL).toString();
            
            // Prepare fetch options
            const fetchOptions = {
                method: config.method,
                headers: config.headers,
                signal: controller.signal
            };

            // Add body if present
            if (config.data && ['POST', 'PUT', 'PATCH'].includes(config.method)) {
                fetchOptions.body = JSON.stringify(config.data);
            }

            // Log for debugging
            console.log(`Fetching: ${fullUrl}`, fetchOptions);

            // Make request with retry logic
            let lastError;
            for (let attempt = 0; attempt <= config.retries; attempt++) {
                try {
                    const response = await fetch(fullUrl, fetchOptions);
                    clearTimeout(timeoutId);
                    console.log(`Response status: ${response.status}`);

                    // Parse response
                    const responseData = await this.parseResponse(response);
                    
                    // Create response object
                    const responseObj = {
                        data: responseData,
                        status: response.status,
                        statusText: response.statusText,
                        headers: Object.fromEntries(response.headers.entries()),
                        config,
                        request: { url: fullUrl, ...fetchOptions }
                    };

                    // Check for HTTP errors
                    if (!response.ok) {
                        throw this.createHTTPError(responseObj);
                    }

                    // Apply response interceptors
                    let finalResponse = responseObj;
                    for (const interceptor of this.responseInterceptors) {
                        try {
                            finalResponse = await (interceptor.onFulfilled ? 
                                interceptor.onFulfilled(finalResponse) : finalResponse);
                        } catch (error) {
                            if (interceptor.onRejected) {
                                finalResponse = await interceptor.onRejected(error);
                            } else {
                                throw error;
                            }
                        }
                    }

                    // Return data for backward compatibility
                    return finalResponse.data;

                } catch (error) {
                    lastError = error;
                    console.error('Request error:', error);
                    
                    // Don't retry on certain errors
                    if (error.name === 'AbortError') {
                        throw new TimeoutError(fullUrl, config.timeout);
                    }
                    
                    if (error instanceof APIError && error.status < 500) {
                        throw error; // Don't retry client errors
                    }

                    // Retry with exponential backoff
                    if (attempt < config.retries) {
                        const delay = config.retryDelay * Math.pow(2, attempt);
                        await new Promise(resolve => setTimeout(resolve, delay));
                        
                        if (this.isDebugMode()) {
                            console.log(`[API] Retrying request (${attempt + 1}/${config.retries}) after ${delay}ms`);
                        }
                    }
                }
            }

            throw lastError;

        } finally {
            clearTimeout(timeoutId);
            this.activeRequests.delete(config.requestId);
        }
    }

    // Response parsing
    async parseResponse(response) {
        const contentType = response.headers.get('content-type');
        
        if (contentType?.includes('application/json')) {
            const jsonResponse = await response.json();
            console.log('JSON Response:', jsonResponse);
            return jsonResponse;
        } else if (contentType?.includes('text/')) {
            return await response.text();
        } else if (contentType?.includes('application/octet-stream')) {
            return await response.blob();
        } else {
            // Default to text
            return await response.text();
        }
    }

    // Error creation
    createHTTPError(response) {
        const { data, status, statusText } = response;
        
        // Check if response is RFC 7807 compliant
        if (data && typeof data === 'object' && data.type) {
            return APIError.fromProblemDetails(data);
        }
        
        // Check for legacy error format
        if (data && data.error) {
            return new APIError(
                '/errors/legacy',
                data.error.message || statusText,
                status,
                data.error.details || '',
                response.config.url,
                data.error
            );
        }
        
        // Generic HTTP error
        return new APIError(
            '/errors/http',
            statusText || `HTTP ${status} Error`,
            status,
            `HTTP ${status} error`,
            response.config.url
        );
    }

    // Convenience methods
    async get(url, options = {}) {
        return this.request('GET', url, options);
    }

    async post(url, body = null, options = {}) {
        console.log(`POST request to ${url} with body:`, body);
        return this.request('POST', url, { ...options, body });
    }

    async put(url, body = null, options = {}) {
        return this.request('PUT', url, { ...options, body });
    }

    async delete(url, options = {}) {
        return this.request('DELETE', url, options);
    }

    // Utility methods
    getCSRFToken() {
        return document.querySelector('meta[name="csrf-token"]')?.content || 
               localStorage.getItem('csrfToken');
    }

    isDebugMode() {
        return window.APP_CONFIG?.DEBUG || 
               location.hostname === 'localhost' || 
               location.search.includes('debug=true');
    }

    // Cancel requests
    cancelRequest(requestId) {
        const request = this.activeRequests.get(requestId);
        if (request) {
            request.controller.abort();
            this.activeRequests.delete(requestId);
        }
    }

    cancelAllRequests() {
        for (const [id, request] of this.activeRequests) {
            request.controller.abort();
        }
        this.activeRequests.clear();
    }

    // Event handling
    on(event, handler) {
        this.eventTarget.addEventListener(event, handler);
    }

    off(event, handler) {
        this.eventTarget.removeEventListener(event, handler);
    }

    emit(event, data) {
        this.eventTarget.dispatchEvent(new CustomEvent(event, { detail: data }));
    }

    // Legacy compatibility methods
    setTimeout(timeout) {
        this.config.timeout = timeout;
        this.requestTimeout = timeout;
    }

    setHeaders(headers) {
        Object.assign(this.config.headers, headers);
        this.defaultHeaders = this.config.headers;
    }

    clearHeaders() {
        this.config.headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        };
        this.defaultHeaders = this.config.headers;
    }

    _getBaseURL() {
        return this.config.baseURL;
    }

    _getCSRFToken() {
        return this.getCSRFToken();
    }

    _getDebugMode() {
        return this.isDebugMode();
    }

    // API-specific methods with error handling
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

    async activateLicense(licenseKey) {
        try {
            return await this.post('/api/license/activate', { key: licenseKey });
        } catch (error) {
            console.error('Failed to activate license:', error);
            return { success: false, message: error.message };
        }
    }

    async getVersionInfo() {
        try {
            return await this.get('/api/version');
        } catch (error) {
            console.error('Failed to get version info:', error);
            return { version: 'unknown', build: 'unknown' };
        }
    }

    async getFiles() {
        try {
            return await this.get('/api/data/files');
        } catch (error) {
            console.error('Failed to get files:', error);
            throw error;
        }
    }

    async getTickers() {
        try {
            return await this.get('/api/data/tickers');
        } catch (error) {
            console.error('Failed to get tickers:', error);
            throw error;
        }
    }

    async getTickerChart(ticker) {
        try {
            return await this.get(`/api/data/ticker/${encodeURIComponent(ticker)}/chart`);
        } catch (error) {
            console.error(`Failed to get chart data for ${ticker}:`, error);
            throw error;
        }
    }

    async getMarketMovers(options = {}) {
        const params = new URLSearchParams();
        
        if (options.period) params.append('period', options.period);
        if (options.limit) params.append('limit', options.limit);
        if (options.min_volume) params.append('min_volume', options.min_volume);

        const queryString = params.toString();
        const url = `/api/data/market-movers${queryString ? '?' + queryString : ''}`;

        try {
            return await this.get(url);
        } catch (error) {
            console.error('Failed to get market movers:', error);
            throw error;
        }
    }

    async getIndexData() {
        try {
            return await this.get('/api/data/indices');
        } catch (error) {
            console.error('Failed to get index data:', error);
            throw error;
        }
    }

    async startScraping(params) {
        try {
            console.log('Starting scraping with params:', params);
            const result = await this.post('/api/scrape', params);
            console.log('Scraping API response:', result);
            return result;
        } catch (error) {
            console.error('Failed to start scraping:', error);
            throw error;
        }
    }

    async startProcessing() {
        try {
            return await this.post('/api/process', {
                command: 'process',
                args: { in: 'data/downloads', out: 'data/reports' }
            });
        } catch (error) {
            console.error('Failed to start processing:', error);
            throw error;
        }
    }

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

    async downloadFile(filename, type) {
        try {
            const url = `/api/download/${encodeURIComponent(type)}/${encodeURIComponent(filename)}`;
            const blob = await this.get(url, { responseType: 'blob' });
            return blob;
        } catch (error) {
            console.error(`Failed to download file ${filename}:`, error);
            throw error;
        }
    }

    async getSystemStats() {
        try {
            // TODO: Implement /api/stats endpoint in backend
            // For now, return empty stats to prevent errors
            return {
                cpu: 0,
                memory: 0,
                disk: 0,
                uptime: 0
            };
        } catch (error) {
            console.error('Failed to get system stats:', error);
            return {};
        }
    }

    // New methods for pipeline API
    async startPipeline(type, parameters) {
        try {
            return await this.post('/api/pipeline/start', { type, parameters });
        } catch (error) {
            if (error instanceof APIError) {
                return {
                    success: false,
                    error: error.toJSON()
                };
            }
            throw error;
        }
    }

    async getPipelineStatus(pipelineId) {
        const params = pipelineId ? `?pipeline_id=${pipelineId}` : '';
        return await this.get(`/api/pipeline/status${params}`);
    }

    async getDataFiles() {
        return await this.get('/api/data/files');
    }

    async downloadDataFile(type, filename) {
        const blob = await this.get(`/api/data/download/${type}/${filename}`, {
            responseType: 'blob'
        });
        return blob;
    }
}

// Create global API service instance
const apiService = new APIService();

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { APIService, apiService, APIError, NetworkError, TimeoutError, ValidationError };
}

// Global access for compatibility
window.APIService = APIService;
window.apiService = apiService;
window.APIError = APIError;