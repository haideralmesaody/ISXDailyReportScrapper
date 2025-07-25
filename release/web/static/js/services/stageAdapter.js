/**
 * Stage Adapter Service
 * Maps between backend and frontend stage IDs to maintain consistency
 * Follows adapter pattern for compatibility
 */
class StageAdapter {
    constructor() {
        // Backend to Frontend mapping
        this.backendToFrontend = {
            'scraping': 'scrape',
            'processing': 'process',
            'indices': 'index',
            'analysis': 'complete'
        };
        
        // Frontend to Backend mapping (reverse)
        this.frontendToBackend = {
            'scrape': 'scraping',
            'process': 'processing',
            'index': 'indices',
            'complete': 'analysis'
        };
        
        // Status mapping
        this.statusMapping = {
            'pending': 'idle',
            'running': 'active',
            'completed': 'completed',
            'failed': 'error',
            'skipped': 'idle'
        };
    }
    
    /**
     * Convert backend stage ID to frontend stage ID
     * @param {string} backendStage - Backend stage ID
     * @returns {string} Frontend stage ID
     */
    toFrontend(backendStage) {
        return this.backendToFrontend[backendStage] || backendStage;
    }
    
    /**
     * Convert frontend stage ID to backend stage ID
     * @param {string} frontendStage - Frontend stage ID
     * @returns {string} Backend stage ID
     */
    toBackend(frontendStage) {
        return this.frontendToBackend[frontendStage] || frontendStage;
    }
    
    /**
     * Map backend status to frontend status
     * @param {string} backendStatus - Backend status
     * @returns {string} Frontend status
     */
    mapStatus(backendStatus) {
        return this.statusMapping[backendStatus] || backendStatus;
    }
    
    /**
     * Transform a complete message from backend to frontend format
     * @param {Object} message - Backend message
     * @returns {Object} Frontend-compatible message
     */
    transformMessage(message) {
        const transformed = { ...message };
        
        // Transform stage ID if present
        if (message.stage) {
            transformed.stage = this.toFrontend(message.stage);
        }
        
        // Transform status if present
        if (message.status) {
            transformed.status = this.mapStatus(message.status);
        }
        
        // Transform nested stage data
        if (message.data && message.data.stage) {
            transformed.data = {
                ...message.data,
                stage: this.toFrontend(message.data.stage)
            };
        }
        
        // Transform stages object if present (for pipeline status)
        if (message.stages) {
            transformed.stages = {};
            for (const [stageId, stageData] of Object.entries(message.stages)) {
                const frontendId = this.toFrontend(stageId);
                transformed.stages[frontendId] = {
                    ...stageData,
                    status: this.mapStatus(stageData.status)
                };
            }
        }
        
        return transformed;
    }
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = StageAdapter;
}

// Global access for compatibility
window.StageAdapter = StageAdapter;