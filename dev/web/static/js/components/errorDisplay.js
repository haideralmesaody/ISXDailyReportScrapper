/**
 * Error Display Component
 * Handles RFC 7807 compliant error display with user-friendly formatting
 */

class ErrorDisplay {
    constructor() {
        this.errorContainer = null;
        this.errorQueue = [];
        this.maxErrors = 5;
        this.autoDismissDelay = 10000; // 10 seconds
    }

    /**
     * Initialize error display component
     */
    init() {
        // Create error container if it doesn't exist
        if (!document.getElementById('error-display-container')) {
            const container = document.createElement('div');
            container.id = 'error-display-container';
            container.className = 'error-display-container position-fixed top-0 end-0 p-3';
            container.style.zIndex = '9999';
            container.style.maxWidth = '400px';
            document.body.appendChild(container);
        }
        
        this.errorContainer = document.getElementById('error-display-container');
        
        // Listen for API errors using the apiService event target
        if (window.apiService && window.apiService.eventTarget) {
            window.apiService.eventTarget.addEventListener('request:error', (event) => {
                const error = event.detail?.error;
                if (error && error.status >= 400) {
                    this.showError(error);
                }
            });
        }
        
        // Also listen for global error events
        window.addEventListener('api:error', (event) => {
            this.showError(event.detail);
        });
    }

    /**
     * Show error message
     * @param {APIError|Object|string} error - Error to display
     * @param {Object} options - Display options
     */
    showError(error, options = {}) {
        const errorData = this.normalizeError(error);
        const errorId = `error-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        
        // Add to queue
        this.errorQueue.push(errorId);
        
        // Remove oldest if exceeding max
        if (this.errorQueue.length > this.maxErrors) {
            const oldestId = this.errorQueue.shift();
            this.dismissError(oldestId);
        }
        
        // Create error element
        const errorElement = this.createErrorElement(errorId, errorData, options);
        this.errorContainer.appendChild(errorElement);
        
        // Animate in
        setTimeout(() => {
            errorElement.classList.add('show');
        }, 10);
        
        // Auto dismiss if enabled
        if (options.autoDismiss !== false) {
            setTimeout(() => {
                this.dismissError(errorId);
            }, options.dismissDelay || this.autoDismissDelay);
        }
    }

    /**
     * Normalize error to consistent format
     */
    normalizeError(error) {
        // Already an APIError
        if (error instanceof APIError) {
            return {
                type: error.type,
                title: error.title,
                status: error.status,
                detail: error.detail,
                instance: error.instance,
                ...error.extensions
            };
        }
        
        // RFC 7807 format
        if (error && typeof error === 'object' && error.type) {
            return error;
        }
        
        // Legacy error format
        if (error && error.message) {
            return {
                type: '/errors/legacy',
                title: error.name || 'Error',
                status: error.status || 500,
                detail: error.message,
                instance: window.location.pathname
            };
        }
        
        // String error
        if (typeof error === 'string') {
            return {
                type: '/errors/string',
                title: 'Error',
                status: 500,
                detail: error,
                instance: window.location.pathname
            };
        }
        
        // Unknown error
        return {
            type: '/errors/unknown',
            title: 'Unknown Error',
            status: 500,
            detail: 'An unexpected error occurred',
            instance: window.location.pathname
        };
    }

    /**
     * Create error element
     */
    createErrorElement(errorId, errorData, options) {
        const alertClass = this.getAlertClass(errorData.status);
        const icon = this.getErrorIcon(errorData.type);
        const showDetails = options.showDetails !== false && errorData.detail;
        
        const errorHtml = `
            <div id="${errorId}" class="alert ${alertClass} alert-dismissible fade mb-2" role="alert">
                <div class="d-flex align-items-start">
                    <div class="me-2">${icon}</div>
                    <div class="flex-grow-1">
                        <strong>${this.escapeHtml(errorData.title)}</strong>
                        ${showDetails ? `
                            <div class="small mt-1">${this.escapeHtml(errorData.detail)}</div>
                        ` : ''}
                        ${errorData.hint ? `
                            <div class="small mt-1 text-muted">
                                <i class="fas fa-lightbulb"></i> ${this.escapeHtml(errorData.hint)}
                            </div>
                        ` : ''}
                        ${errorData.error_code ? `
                            <div class="error-code mt-1">
                                <small class="text-muted">Error Code: ${this.escapeHtml(errorData.error_code)}</small>
                            </div>
                        ` : ''}
                    </div>
                </div>
                <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"
                        onclick="window.errorDisplay.dismissError('${errorId}')"></button>
            </div>
        `;
        
        const template = document.createElement('template');
        template.innerHTML = errorHtml.trim();
        return template.content.firstChild;
    }

    /**
     * Get alert class based on status code
     */
    getAlertClass(status) {
        if (status >= 500) return 'alert-danger';
        if (status >= 400) return 'alert-warning';
        if (status >= 300) return 'alert-info';
        return 'alert-secondary';
    }

    /**
     * Get error icon based on type
     */
    getErrorIcon(type) {
        const iconMap = {
            '/errors/network': '<i class="fas fa-wifi text-danger"></i>',
            '/errors/timeout': '<i class="fas fa-clock text-warning"></i>',
            '/errors/validation': '<i class="fas fa-exclamation-triangle text-warning"></i>',
            '/errors/license': '<i class="fas fa-key text-danger"></i>',
            '/errors/permission': '<i class="fas fa-lock text-danger"></i>',
            '/errors/not-found': '<i class="fas fa-search text-warning"></i>',
            '/errors/rate-limit': '<i class="fas fa-tachometer-alt text-warning"></i>',
            '/errors/server': '<i class="fas fa-server text-danger"></i>'
        };
        
        // Check if type starts with any known pattern
        for (const [pattern, icon] of Object.entries(iconMap)) {
            if (type.startsWith(pattern)) {
                return icon;
            }
        }
        
        return '<i class="fas fa-exclamation-circle text-danger"></i>';
    }

    /**
     * Dismiss error
     */
    dismissError(errorId) {
        const errorElement = document.getElementById(errorId);
        if (errorElement) {
            errorElement.classList.remove('show');
            setTimeout(() => {
                errorElement.remove();
            }, 150);
            
            // Remove from queue
            const index = this.errorQueue.indexOf(errorId);
            if (index > -1) {
                this.errorQueue.splice(index, 1);
            }
        }
    }

    /**
     * Clear all errors
     */
    clearAll() {
        this.errorQueue.forEach(errorId => {
            this.dismissError(errorId);
        });
        this.errorQueue = [];
    }

    /**
     * Show validation errors
     */
    showValidationErrors(errors, fieldPrefix = '') {
        const errorDetails = Object.entries(errors).map(([field, messages]) => {
            const fieldName = fieldPrefix ? `${fieldPrefix}.${field}` : field;
            const messageList = Array.isArray(messages) ? messages : [messages];
            return `${fieldName}: ${messageList.join(', ')}`;
        }).join('\n');
        
        this.showError(new ValidationError(errors), {
            showDetails: true,
            autoDismiss: false
        });
    }

    /**
     * Show network error
     */
    showNetworkError(message = 'Unable to connect to server') {
        this.showError(new NetworkError(message), {
            showDetails: true,
            dismissDelay: 5000
        });
    }

    /**
     * Show success message (for completeness)
     */
    showSuccess(title, detail = '', options = {}) {
        const successId = `success-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        
        const successHtml = `
            <div id="${successId}" class="alert alert-success alert-dismissible fade mb-2" role="alert">
                <div class="d-flex align-items-start">
                    <div class="me-2"><i class="fas fa-check-circle text-success"></i></div>
                    <div class="flex-grow-1">
                        <strong>${this.escapeHtml(title)}</strong>
                        ${detail ? `<div class="small mt-1">${this.escapeHtml(detail)}</div>` : ''}
                    </div>
                </div>
                <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"
                        onclick="window.errorDisplay.dismissError('${successId}')"></button>
            </div>
        `;
        
        const template = document.createElement('template');
        template.innerHTML = successHtml.trim();
        const successElement = template.content.firstChild;
        
        this.errorContainer.appendChild(successElement);
        
        // Animate in
        setTimeout(() => {
            successElement.classList.add('show');
        }, 10);
        
        // Auto dismiss
        setTimeout(() => {
            this.dismissError(successId);
        }, options.dismissDelay || 3000);
    }

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }
}

// Create global instance
window.errorDisplay = new ErrorDisplay();

// Initialize when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.errorDisplay.init();
    });
} else {
    window.errorDisplay.init();
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ErrorDisplay;
}