/**
 * Template Loader for HTML Component Templates
 * Handles loading, caching, and rendering of HTML templates
 * Supports simple variable interpolation and template composition
 */

class TemplateLoader {
    constructor() {
        this.cache = new Map();
        this.baseURL = '/templates/';
        this.loadingPromises = new Map(); // Prevent duplicate fetches
    }

    /**
     * Load a template from the server
     * @param {string} templatePath - Path to template file relative to templates directory
     * @returns {Promise<string>} Template HTML content
     */
    async load(templatePath) {
        // Return cached template if available
        if (this.cache.has(templatePath)) {
            return this.cache.get(templatePath);
        }

        // Check if already loading to prevent duplicate requests
        if (this.loadingPromises.has(templatePath)) {
            return this.loadingPromises.get(templatePath);
        }

        // Create loading promise
        const loadingPromise = this.fetchTemplate(templatePath);
        this.loadingPromises.set(templatePath, loadingPromise);

        try {
            const template = await loadingPromise;
            this.cache.set(templatePath, template);
            this.loadingPromises.delete(templatePath);
            return template;
        } catch (error) {
            this.loadingPromises.delete(templatePath);
            throw error;
        }
    }

    /**
     * Fetch template from server
     * @private
     * @param {string} templatePath - Template path
     * @returns {Promise<string>} Template content
     */
    async fetchTemplate(templatePath) {
        try {
            const response = await fetch(this.baseURL + templatePath);
            
            if (!response.ok) {
                throw new Error(`Failed to load template: ${templatePath} (${response.status})`);
            }
            
            return await response.text();
        } catch (error) {
            console.error(`Error loading template ${templatePath}:`, error);
            throw error;
        }
    }

    /**
     * Load and render a template with data
     * @param {string} templatePath - Path to template file
     * @param {object} data - Data to interpolate into template
     * @returns {Promise<string>} Rendered template HTML
     */
    async render(templatePath, data = {}) {
        const template = await this.load(templatePath);
        return this.interpolate(template, data);
    }

    /**
     * Simple template interpolation
     * Replaces {{variable}} with data values
     * @param {string} template - Template string
     * @param {object} data - Data object
     * @returns {string} Interpolated template
     */
    interpolate(template, data) {
        // Basic variable replacement: {{variableName}}
        let result = template.replace(/\{\{(\w+)\}\}/g, (match, key) => {
            return data.hasOwnProperty(key) ? this.escapeHtml(String(data[key])) : '';
        });

        // Handle conditionals: {{#if condition}}...{{/if}}
        result = this.processConditionals(result, data);

        // Handle loops: {{#each array}}...{{/each}}
        result = this.processLoops(result, data);

        return result;
    }

    /**
     * Process conditional blocks in template
     * @private
     * @param {string} template - Template string
     * @param {object} data - Data object
     * @returns {string} Processed template
     */
    processConditionals(template, data) {
        const conditionalRegex = /\{\{#if\s+(\w+)\}\}([\s\S]*?)\{\{\/if\}\}/g;
        
        return template.replace(conditionalRegex, (match, condition, content) => {
            const value = data[condition];
            return value ? content : '';
        });
    }

    /**
     * Process loop blocks in template
     * @private
     * @param {string} template - Template string
     * @param {object} data - Data object
     * @returns {string} Processed template
     */
    processLoops(template, data) {
        const loopRegex = /\{\{#each\s+(\w+)\}\}([\s\S]*?)\{\{\/each\}\}/g;
        
        return template.replace(loopRegex, (match, arrayName, content) => {
            const array = data[arrayName];
            if (!Array.isArray(array)) return '';
            
            return array.map((item, index) => {
                // Create a new data context for each iteration
                const itemData = {
                    ...data,
                    item: item,
                    index: index,
                    '@index': index,
                    '@first': index === 0,
                    '@last': index === array.length - 1
                };
                
                // Replace item properties
                return content.replace(/\{\{item\.(\w+)\}\}/g, (match, prop) => {
                    return item.hasOwnProperty(prop) ? this.escapeHtml(String(item[prop])) : '';
                });
            }).join('');
        });
    }

    /**
     * Escape HTML to prevent XSS
     * @private
     * @param {string} text - Text to escape
     * @returns {string} Escaped text
     */
    escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#39;'
        };
        return text.replace(/[&<>"']/g, char => map[char]);
    }

    /**
     * Load multiple templates in parallel
     * @param {string[]} templatePaths - Array of template paths
     * @returns {Promise<object>} Object mapping paths to templates
     */
    async loadMultiple(templatePaths) {
        const promises = templatePaths.map(path => 
            this.load(path).then(template => ({ path, template }))
        );
        
        const results = await Promise.all(promises);
        return results.reduce((acc, { path, template }) => {
            acc[path] = template;
            return acc;
        }, {});
    }

    /**
     * Clear template cache
     * @param {string} [templatePath] - Specific template to clear, or all if omitted
     */
    clearCache(templatePath = null) {
        if (templatePath) {
            this.cache.delete(templatePath);
        } else {
            this.cache.clear();
        }
    }

    /**
     * Preload templates for better performance
     * @param {string[]} templatePaths - Templates to preload
     * @returns {Promise<void>}
     */
    async preload(templatePaths) {
        await this.loadMultiple(templatePaths);
    }

    /**
     * Get cache statistics
     * @returns {object} Cache stats
     */
    getCacheStats() {
        return {
            size: this.cache.size,
            templates: Array.from(this.cache.keys())
        };
    }
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = TemplateLoader;
}

// Global access for compatibility
window.TemplateLoader = TemplateLoader;