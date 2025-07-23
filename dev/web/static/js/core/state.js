/**
 * UI State Management
 * Manages application UI state (display-only, no business logic)
 * Handles section visibility, user preferences, and UI-only data
 */

class UIState {
    constructor() {
        this.state = {
            // Section visibility
            activeSection: 'dashboard',
            sections: {
                dashboard: { visible: true, loaded: false },
                scraper: { visible: false, loaded: false },
                process: { visible: false, loaded: false },
                indexcsv: { visible: false, loaded: false },
                tickercharts: { visible: false, loaded: false },
                marketmovers: { visible: false, loaded: false },
                files: { visible: false, loaded: false }
            },
            
            // Connection status
            connection: {
                websocket: false,
                license: 'unknown'
            },
            
            // UI preferences
            preferences: {
                theme: 'light',
                autoRefresh: true,
                chartHeight: 400,
                tablePageSize: 50
            },
            
            // Current selections
            selections: {
                ticker: null,
                dateRange: null,
                period: '1d'
            }
        };
        
        this.observers = [];
        this.loadFromStorage();
    }

    /**
     * Set state value by path
     * @param {string} path - Dot notation path (e.g., 'sections.dashboard.visible')
     * @param {any} value - Value to set
     */
    setState(path, value) {
        const keys = path.split('.');
        let current = this.state;
        
        // Navigate to parent object
        for (let i = 0; i < keys.length - 1; i++) {
            if (!current[keys[i]]) {
                current[keys[i]] = {};
            }
            current = current[keys[i]];
        }
        
        // Set the value
        const lastKey = keys[keys.length - 1];
        const oldValue = current[lastKey];
        current[lastKey] = value;
        
        // Notify observers if value changed
        if (oldValue !== value) {
            this.notifyObservers(path, value, oldValue);
        }
        
        // Persist to storage
        this.saveToStorage();
    }

    /**
     * Get state value by path
     * @param {string} path - Dot notation path
     * @returns {any} State value
     */
    getState(path) {
        const keys = path.split('.');
        let current = this.state;
        
        for (const key of keys) {
            if (current && typeof current === 'object' && key in current) {
                current = current[key];
            } else {
                return undefined;
            }
        }
        
        return current;
    }

    /**
     * Get entire state object
     * @returns {object} Complete state
     */
    getAllState() {
        return { ...this.state };
    }

    /**
     * Add state observer
     * @param {Function} callback - Observer function
     */
    observe(callback) {
        this.observers.push(callback);
    }

    /**
     * Remove state observer
     * @param {Function} callback - Observer function to remove
     */
    unobserve(callback) {
        const index = this.observers.indexOf(callback);
        if (index > -1) {
            this.observers.splice(index, 1);
        }
    }

    /**
     * Notify all observers of state change
     * @param {string} path - Changed path
     * @param {any} value - New value
     * @param {any} oldValue - Previous value
     */
    notifyObservers(path, value, oldValue) {
        this.observers.forEach(observer => {
            try {
                observer(path, value, oldValue);
            } catch (error) {
                console.error('State observer error:', error);
            }
        });
    }

    /**
     * Set active section
     * @param {string} sectionName - Section name
     */
    setActiveSection(sectionName) {
        const oldSection = this.state.activeSection;
        
        // Update section visibility
        Object.keys(this.state.sections).forEach(section => {
            this.setState(`sections.${section}.visible`, section === sectionName);
        });
        
        this.setState('activeSection', sectionName);
        
        // Mark section as loaded if not already
        if (!this.getState(`sections.${sectionName}.loaded`)) {
            this.setState(`sections.${sectionName}.loaded`, true);
        }
    }

    /**
     * Get active section name
     * @returns {string} Active section name
     */
    getActiveSection() {
        return this.state.activeSection;
    }

    /**
     * Set connection status
     * @param {boolean} websocket - WebSocket connection status
     * @param {string} license - License status
     */
    setConnectionStatus(websocket, license = null) {
        this.setState('connection.websocket', websocket);
        if (license !== null) {
            this.setState('connection.license', license);
        }
    }

    /**
     * Set user preference
     * @param {string} key - Preference key
     * @param {any} value - Preference value
     */
    setPreference(key, value) {
        this.setState(`preferences.${key}`, value);
    }

    /**
     * Get user preference
     * @param {string} key - Preference key
     * @returns {any} Preference value
     */
    getPreference(key) {
        return this.getState(`preferences.${key}`);
    }

    /**
     * Set current selection
     * @param {string} type - Selection type (ticker, dateRange, period)
     * @param {any} value - Selection value
     */
    setSelection(type, value) {
        this.setState(`selections.${type}`, value);
    }

    /**
     * Get current selection
     * @param {string} type - Selection type
     * @returns {any} Selection value
     */
    getSelection(type) {
        return this.getState(`selections.${type}`);
    }

    /**
     * Load state from localStorage
     */
    loadFromStorage() {
        try {
            const stored = localStorage.getItem('isx_ui_state');
            if (stored) {
                const parsedState = JSON.parse(stored);
                
                // Only restore preferences and selections, not UI state
                if (parsedState.preferences) {
                    this.state.preferences = { ...this.state.preferences, ...parsedState.preferences };
                }
                if (parsedState.selections) {
                    this.state.selections = { ...this.state.selections, ...parsedState.selections };
                }
            }
        } catch (error) {
            console.warn('Failed to load UI state from storage:', error);
        }
    }

    /**
     * Save state to localStorage
     */
    saveToStorage() {
        try {
            const toSave = {
                preferences: this.state.preferences,
                selections: this.state.selections
            };
            localStorage.setItem('isx_ui_state', JSON.stringify(toSave));
        } catch (error) {
            console.warn('Failed to save UI state to storage:', error);
        }
    }

    /**
     * Reset state to defaults
     */
    reset() {
        this.state = {
            activeSection: 'dashboard',
            sections: {
                dashboard: { visible: true, loaded: false },
                scraper: { visible: false, loaded: false },
                process: { visible: false, loaded: false },
                indexcsv: { visible: false, loaded: false },
                tickercharts: { visible: false, loaded: false },
                marketmovers: { visible: false, loaded: false },
                files: { visible: false, loaded: false }
            },
            connection: {
                websocket: false,
                license: 'unknown'
            },
            preferences: {
                theme: 'light',
                autoRefresh: true,
                chartHeight: 400,
                tablePageSize: 50
            },
            selections: {
                ticker: null,
                dateRange: null,
                period: '1d'
            }
        };
        
        this.saveToStorage();
        this.notifyObservers('*', this.state, {});
    }
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = UIState;
}

// Global access for compatibility
window.UIState = UIState;