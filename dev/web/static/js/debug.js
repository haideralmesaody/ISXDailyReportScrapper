/**
 * ISX Debug Utilities
 * Provides debugging tools for license detection and general application debugging
 */

window.ISXDebug = {
    /**
     * Log license check steps for debugging
     */
    logLicenseCheck: function(step, data) {
        const timestamp = new Date().toISOString();
        console.log(`[ISX-LICENSE] ${timestamp} ${step}`, data);
        
        // Store debug info in localStorage for persistent debugging
        try {
            const debugLogs = JSON.parse(localStorage.getItem('ISX_LICENSE_DEBUG') || '[]');
            debugLogs.push({
                timestamp,
                step,
                data: typeof data === 'object' ? JSON.stringify(data) : data
            });
            
            // Keep only last 50 debug entries
            if (debugLogs.length > 50) {
                debugLogs.splice(0, debugLogs.length - 50);
            }
            
            localStorage.setItem('ISX_LICENSE_DEBUG', JSON.stringify(debugLogs));
        } catch (e) {
            console.warn('Failed to store debug log:', e);
        }
    },

    /**
     * Test license API directly
     */
    testLicenseAPI: function() {
        console.log('Testing license API...');
        
        fetch('/api/license/status')
            .then(response => {
                console.log('API Response Status:', response.status);
                console.log('API Response Headers:', Object.fromEntries(response.headers.entries()));
                return response.json();
            })
            .then(data => {
                console.log('API Response Data:', data);
                this.logLicenseCheck('API_TEST_SUCCESS', data);
            })
            .catch(error => {
                console.error('API Test Failed:', error);
                this.logLicenseCheck('API_TEST_ERROR', error.message);
            });
    },

    /**
     * Get stored license debug logs
     */
    getLicenseDebugLogs: function() {
        try {
            return JSON.parse(localStorage.getItem('ISX_LICENSE_DEBUG') || '[]');
        } catch (e) {
            console.error('Failed to retrieve debug logs:', e);
            return [];
        }
    },

    /**
     * Clear license debug logs
     */
    clearLicenseDebugLogs: function() {
        localStorage.removeItem('ISX_LICENSE_DEBUG');
        console.log('License debug logs cleared');
    },

    /**
     * Export debug logs for support
     */
    exportDebugLogs: function() {
        const logs = this.getLicenseDebugLogs();
        const blob = new Blob([JSON.stringify(logs, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `isx-license-debug-${new Date().toISOString().slice(0, 19)}.json`;
        a.click();
        URL.revokeObjectURL(url);
        console.log('Debug logs exported');
    },

    /**
     * Enable debug mode
     */
    enableDebugMode: function() {
        localStorage.setItem('ISX_DEBUG', 'true');
        console.log('Debug mode enabled. Refresh the page to see debug output.');
    },

    /**
     * Disable debug mode
     */
    disableDebugMode: function() {
        localStorage.removeItem('ISX_DEBUG');
        console.log('Debug mode disabled. Refresh the page.');
    },

    /**
     * Show current debug status
     */
    getDebugStatus: function() {
        const isDebug = localStorage.getItem('ISX_DEBUG') === 'true';
        console.log('Debug mode:', isDebug ? 'ENABLED' : 'DISABLED');
        return isDebug;
    }
};

// Auto-enable debug mode on localhost
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    window.ISXDebug.enableDebugMode();
}

console.log('ISX Debug utilities loaded. Use window.ISXDebug for debugging functions.');