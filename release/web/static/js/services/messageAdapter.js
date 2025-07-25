/**
 * WebSocket Message Adapter
 * Handles both legacy and new message formats for backward compatibility
 */

class MessageAdapter {
    constructor() {
        this.messageHandlers = new Map();
        this.activePipelines = new Map(); // Track pipeline -> stage mapping
    }

    /**
     * Process incoming WebSocket message
     * @param {Object} message - Raw WebSocket message
     * @returns {Object} Normalized message
     */
    normalizeMessage(message) {
        // If message already has timestamp in new format, parse it
        if (message.timestamp && typeof message.timestamp === 'string') {
            try {
                message.timestamp = new Date(message.timestamp);
            } catch (e) {
                message.timestamp = new Date();
            }
        } else {
            message.timestamp = new Date();
        }

        // Handle different message types
        switch (message.type) {
            case 'stage_progress':
                return this.normalizeStageProgress(message);
            
            case 'pipeline_progress':
                return this.normalizePipelineProgress(message);
            
            case 'pipeline_status':
                return this.normalizePipelineStatus(message);
            
            case 'data_update':
                return this.normalizeDataUpdate(message);
            
            case 'error':
                return this.normalizeError(message);
            
            case 'output':
                return this.normalizeOutput(message);
            
            case 'connection':
                return this.normalizeConnection(message);
            
            default:
                // Return as-is for unknown types
                return message;
        }
    }

    /**
     * Convert old stage_progress to new pipeline_progress format
     */
    normalizeStageProgress(message) {
        const normalized = {
            type: 'pipeline_progress',
            timestamp: message.timestamp,
            pipeline_id: message.pipeline_id || '',
            stage: message.stage || '',
            progress: 0,
            message: '',
            current: 0,
            total: 0
        };

        // Extract from data object if present
        if (message.data) {
            normalized.progress = message.data.progress || message.progress || 0;
            normalized.message = message.data.message || message.message || '';
            normalized.current = message.data.current || 0;
            normalized.total = message.data.total || 0;
        } else {
            normalized.progress = message.progress || 0;
            normalized.message = message.message || '';
        }

        // Track stage for pipeline mapping
        if (normalized.pipeline_id && normalized.stage) {
            this.activePipelines.set(normalized.pipeline_id, normalized.stage);
        }

        return normalized;
    }

    /**
     * Normalize pipeline progress message
     */
    normalizePipelineProgress(message) {
        return {
            type: 'pipeline_progress',
            timestamp: message.timestamp,
            pipeline_id: message.pipeline_id || '',
            stage: message.stage || '',
            progress: message.progress || 0,
            message: message.message || '',
            current: message.current || 0,
            total: message.total || 0
        };
    }

    /**
     * Normalize pipeline status message
     */
    normalizePipelineStatus(message) {
        const normalized = {
            type: 'pipeline_status',
            timestamp: message.timestamp,
            pipeline_id: message.pipeline_id || '',
            status: message.status || '',
            stage: message.stage || '',
            details: message.details || {}
        };

        // Extract from data object if present (legacy format)
        if (message.data && !message.status) {
            normalized.status = message.data.status || '';
            normalized.stage = message.data.stage || '';
            normalized.pipeline_id = message.data.pipeline_id || '';
            normalized.details = message.data;
        }

        return normalized;
    }

    /**
     * Normalize data update message
     */
    normalizeDataUpdate(message) {
        return {
            type: 'data_update',
            timestamp: message.timestamp,
            subtype: message.subtype || '',
            action: message.action || '',
            data: message.data || null,
            count: message.count || 0
        };
    }

    /**
     * Normalize error message
     */
    normalizeError(message) {
        const normalized = {
            type: 'error',
            timestamp: message.timestamp,
            error_code: '',
            title: 'Error',
            detail: '',
            stage: '',
            hint: '',
            metadata: {}
        };

        // Handle new format
        if (message.error_code) {
            normalized.error_code = message.error_code;
            normalized.title = message.title || 'Error';
            normalized.detail = message.detail || '';
            normalized.stage = message.stage || '';
            normalized.hint = message.hint || '';
            normalized.metadata = message.metadata || {};
        }
        // Handle legacy format with data object
        else if (message.data) {
            normalized.error_code = message.data.code || message.data.error_code || 'ERR_UNKNOWN';
            normalized.title = message.data.title || message.data.message || 'Error';
            normalized.detail = message.data.detail || message.data.details || '';
            normalized.stage = message.data.stage || '';
            normalized.hint = message.data.hint || '';
            normalized.metadata = message.data;
        }
        // Handle simple error format
        else {
            normalized.error_code = message.code || 'ERR_UNKNOWN';
            normalized.title = message.message || 'Error';
            normalized.detail = message.details || '';
            normalized.stage = message.stage || '';
        }

        return normalized;
    }

    /**
     * Normalize output message
     */
    normalizeOutput(message) {
        const normalized = {
            type: 'output',
            timestamp: message.timestamp,
            level: 'info',
            message: '',
            stage: ''
        };

        // Handle new format
        if (message.level) {
            normalized.level = message.level;
            normalized.message = message.message || '';
            normalized.stage = message.stage || '';
        }
        // Handle legacy format with data object
        else if (message.data) {
            normalized.level = message.data.level || 'info';
            normalized.message = message.data.message || '';
            normalized.stage = message.data.stage || '';
        }

        return normalized;
    }

    /**
     * Normalize connection message
     */
    normalizeConnection(message) {
        const normalized = {
            type: 'connection',
            timestamp: message.timestamp,
            status: '',
            client_id: '',
            session_id: ''
        };

        // Handle new format
        if (message.status) {
            normalized.status = message.status;
            normalized.client_id = message.client_id || '';
            normalized.session_id = message.session_id || '';
        }
        // Handle legacy format with data object
        else if (message.data) {
            normalized.status = message.data.status || '';
            normalized.client_id = message.data.client_id || '';
            normalized.session_id = message.data.session_id || '';
        }

        return normalized;
    }

    /**
     * Check if a message is in legacy format
     */
    isLegacyFormat(message) {
        // Legacy messages typically have data wrapped in a 'data' object
        // and may use different type names
        return (message.data && typeof message.data === 'object') ||
               message.type === 'stage_progress' ||
               message.type === 'refresh';
    }

    /**
     * Convert new format message to legacy format for backward compatibility
     */
    toLegacyFormat(message) {
        const legacy = {
            type: message.type,
            timestamp: message.timestamp
        };

        switch (message.type) {
            case 'pipeline_progress':
                legacy.type = 'stage_progress';
                legacy.stage = message.stage;
                legacy.data = {
                    progress: message.progress,
                    message: message.message
                };
                break;

            case 'pipeline_status':
                legacy.data = {
                    status: message.status,
                    stage: message.stage,
                    pipeline_id: message.pipeline_id
                };
                break;

            case 'error':
                legacy.data = {
                    code: message.error_code,
                    message: message.title,
                    details: message.detail,
                    stage: message.stage,
                    hint: message.hint
                };
                break;

            case 'output':
                legacy.data = {
                    level: message.level,
                    message: message.message,
                    stage: message.stage
                };
                break;

            default:
                // Copy all properties to data
                legacy.data = { ...message };
                delete legacy.data.type;
                delete legacy.data.timestamp;
        }

        return legacy;
    }
}

// Export for use in other modules
window.MessageAdapter = MessageAdapter;