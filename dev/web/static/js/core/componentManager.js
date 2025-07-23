/**
 * Component Manager
 * Manages loading, initialization, and lifecycle of HTML components
 * Integrates with TemplateLoader and EventBus for dynamic component handling
 */

class ComponentManager {
    constructor(templateLoader, eventBus, apiService) {
        this.templateLoader = templateLoader;
        this.eventBus = eventBus;
        this.apiService = apiService;
        this.components = new Map();
        this.initializers = new Map();
        
        // Register built-in component initializers
        this.registerInitializers();
    }

    /**
     * Register component initializers
     * @private
     */
    registerInitializers() {
        // Scraper component
        this.initializers.set('scraper', (container) => {
            const form = container.querySelector('#scrapeForm');
            console.log('Scraper initializer called, form found:', !!form);
            if (form) {
                form.addEventListener('submit', (e) => {
                    e.preventDefault();
                    console.log('Scraper form submitted, emitting event');
                    this.eventBus.emit('scraper:submit', new FormData(form));
                });
                
                // Set default dates
                const fromDateInput = container.querySelector('#fromDateInput');
                const toDateInput = container.querySelector('#toDateInput');
                
                if (fromDateInput && !fromDateInput.value) {
                    // Set to January 1, 2025
                    fromDateInput.value = '2025-01-01';
                }
                
                if (toDateInput && !toDateInput.value) {
                    // Set to today's date
                    const today = new Date();
                    const yyyy = today.getFullYear();
                    const mm = String(today.getMonth() + 1).padStart(2, '0');
                    const dd = String(today.getDate()).padStart(2, '0');
                    toDateInput.value = `${yyyy}-${mm}-${dd}`;
                }
            }
            this.initializePipelineStages(container);
        });

        // Processor component
        this.initializers.set('processor', (container) => {
            const form = container.querySelector('#processForm');
            if (form) {
                form.addEventListener('submit', (e) => {
                    e.preventDefault();
                    this.eventBus.emit('processor:submit');
                });
            }
        });

        // Index CSV component
        this.initializers.set('indexcsv', (container) => {
            const form = container.querySelector('#indexcsvForm');
            if (form) {
                form.addEventListener('submit', (e) => {
                    e.preventDefault();
                    this.eventBus.emit('indexcsv:submit');
                });
            }
            
            // Listen for index chart data updates
            this.eventBus.on('indexchart:update', (data) => {
                this.renderIndexChart(container, data);
            });
        });

        // Ticker charts component
        this.initializers.set('tickercharts', (container) => {
            this.initializeTickerSearch(container);
            this.initializeTickerList(container);
            
            // Listen for ticker data updates
            this.eventBus.on('tickers:update', (tickers) => {
                this.renderTickerList(container, tickers);
            });
        });

        // Market movers component
        this.initializers.set('marketmovers', (container) => {
            this.initializeMarketMovers(container);
            
            // Listen for market movers updates
            this.eventBus.on('marketmovers:update', (movers) => {
                this.renderMarketMovers(container, movers);
            });
        });

        // Files component
        this.initializers.set('files', (container) => {
            this.initializeFileManager(container);
            
            // Listen for files updates
            this.eventBus.on('files:update', (files) => {
                this.renderFileList(container, files);
            });
        });

        // Sidebar component
        this.initializers.set('sidebar', (container) => {
            this.initializeNavigation(container);
        });

        // License component
        this.initializers.set('license', (container) => {
            // License initialization is handled by main.js setupLicensePageHandlers
        });
    }

    /**
     * Load a component into a container
     * @param {string} name - Component name
     * @param {string} containerSelector - Container selector
     * @param {object} data - Data to pass to template
     * @returns {Promise<void>}
     */
    async loadComponent(name, containerSelector, data = {}) {
        try {
            const container = document.querySelector(containerSelector);
            if (!container) {
                throw new Error(`Container not found: ${containerSelector}`);
            }

            // Determine template path based on component type
            const templatePath = this.getTemplatePath(name);
            
            // Render template with data
            const html = await this.templateLoader.render(templatePath, data);
            
            // Update container
            container.innerHTML = html;
            
            // Store component reference
            this.components.set(name, {
                container: container,
                selector: containerSelector,
                data: data
            });

            // Initialize component
            await this.initializeComponent(name, container);
            
            // Emit component loaded event
            this.eventBus.emit('component:loaded', { name, container });
            
        } catch (error) {
            console.error(`Failed to load component ${name}:`, error);
            this.eventBus.emit('component:error', { name, error });
        }
    }

    /**
     * Get template path for a component
     * @private
     * @param {string} name - Component name
     * @returns {string} Template path
     */
    getTemplatePath(name) {
        // Check if it's a layout component
        const layoutComponents = ['header', 'sidebar', 'footer'];
        if (layoutComponents.includes(name)) {
            return `layout/${name}.html`;
        }
        
        // Check if it's a section component
        const sectionComponents = ['scraper', 'processor', 'indexcsv', 'tickercharts', 'marketmovers', 'files', 'license'];
        if (sectionComponents.includes(name)) {
            return `sections/${name}.html`;
        }
        
        // Default to components directory
        return `components/${name}.html`;
    }

    /**
     * Initialize a component after loading
     * @private
     * @param {string} name - Component name
     * @param {HTMLElement} container - Component container
     * @returns {Promise<void>}
     */
    async initializeComponent(name, container) {
        const initializer = this.initializers.get(name);
        if (initializer) {
            await initializer(container);
        }
        
        // Initialize any nested components
        await this.initializeNestedComponents(container);
    }

    /**
     * Initialize nested components within a container
     * @private
     * @param {HTMLElement} container - Parent container
     * @returns {Promise<void>}
     */
    async initializeNestedComponents(container) {
        const nestedComponents = container.querySelectorAll('[data-component]');
        
        for (const element of nestedComponents) {
            const componentName = element.getAttribute('data-component');
            const componentData = element.dataset;
            
            await this.loadComponent(componentName, `#${element.id}`, componentData);
        }
    }

    /**
     * Initialize pipeline stages
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializePipelineStages(container) {
        const stages = container.querySelectorAll('.pipeline-stage');
        stages.forEach(stage => {
            // Add click handlers for stage interaction
            stage.addEventListener('click', () => {
                const stageName = stage.id.replace('stage-', '');
                this.eventBus.emit('pipeline:stage:clicked', { stage: stageName });
            });
        });
        
        // Listen for pipeline status updates
        this.eventBus.on('pipeline:status', (data) => {
            this.updatePipelineStatus(data);
        });
        
        this.eventBus.on('pipeline:progress', (data) => {
            this.updatePipelineProgress(data);
        });
        
        this.eventBus.on('pipeline:complete', (data) => {
            this.updatePipelineComplete(data);
        });
        
        this.eventBus.on('pipeline:update', (data) => {
            this.updatePipelineStatus(data);
        });
    }
    
    /**
     * Update pipeline status
     * @private
     * @param {object} data - Status data
     */
    updatePipelineStatus(data) {
        const stageMap = {
            'scraping': 'stage-scraping',
            'processing': 'stage-processing',
            'indices': 'stage-indices',
            'analysis': 'stage-analysis'
        };
        
        const stageId = stageMap[data.stage];
        if (!stageId) return;
        
        const stageElement = document.getElementById(stageId);
        if (!stageElement) return;
        
        // Reset all stages
        document.querySelectorAll('.pipeline-stage').forEach(stage => {
            stage.classList.remove('active', 'processing', 'completed', 'error');
        });
        
        // Update current stage
        if (data.status === 'running' || data.status === 'in_progress') {
            stageElement.classList.add('processing');
        } else if (data.status === 'completed') {
            stageElement.classList.add('completed');
        } else if (data.status === 'error' || data.status === 'failed') {
            stageElement.classList.add('error');
        } else {
            stageElement.classList.add('active');
        }
        
        // Update status text
        const statusElement = stageElement.querySelector('.stage-status');
        if (statusElement) {
            statusElement.textContent = data.message || '';
        }
    }
    
    /**
     * Update pipeline progress
     * @private
     * @param {object} data - Progress data
     */
    updatePipelineProgress(data) {
        const stageMap = {
            'scraping': 'stage-scraping',
            'processing': 'stage-processing',
            'indices': 'stage-indices',
            'analysis': 'stage-analysis'
        };
        
        const stageId = stageMap[data.stage];
        if (!stageId) return;
        
        const stageElement = document.getElementById(stageId);
        if (!stageElement) return;
        
        // Update stage to processing
        stageElement.classList.add('processing');
        
        // Update progress bar
        const progressContainer = stageElement.querySelector('.progress');
        const progressBar = stageElement.querySelector('.progress-bar');
        if (progressContainer && progressBar) {
            progressContainer.style.display = 'block';
            progressBar.style.width = `${data.progress || 0}%`;
        }
        
        // Update status text
        const statusElement = stageElement.querySelector('.stage-status');
        if (statusElement) {
            statusElement.textContent = data.message || '';
        }
    }
    
    /**
     * Update pipeline complete
     * @private
     * @param {object} data - Completion data
     */
    updatePipelineComplete(data) {
        // Mark all stages as completed
        document.querySelectorAll('.pipeline-stage').forEach(stage => {
            stage.classList.remove('active', 'processing');
            stage.classList.add('completed');
            
            const progressContainer = stage.querySelector('.progress');
            if (progressContainer) {
                progressContainer.style.display = 'none';
            }
        });
    }

    /**
     * Initialize navigation
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializeNavigation(container) {
        const navLinks = container.querySelectorAll('[data-section]');
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('data-section');
                this.eventBus.emit('navigation:click', { section });
            });
        });
    }

    /**
     * Initialize ticker search functionality
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializeTickerSearch(container) {
        const searchInput = container.querySelector('#tickerSearch');
        const searchResults = container.querySelector('#searchResults');
        
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                const query = e.target.value.trim();
                
                if (query.length >= 2) {
                    searchTimeout = setTimeout(() => {
                        this.eventBus.emit('ticker:search', { query });
                    }, 300);
                } else {
                    searchResults.style.display = 'none';
                }
            });
        }
    }

    /**
     * Initialize ticker list
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializeTickerList(container) {
        const tickerRows = container.querySelectorAll('tr[data-ticker]');
        tickerRows.forEach(row => {
            row.addEventListener('click', () => {
                const ticker = row.getAttribute('data-ticker');
                // Highlight selected row
                tickerRows.forEach(r => r.classList.remove('table-active'));
                row.classList.add('table-active');
                this.eventBus.emit('ticker:selected', { ticker });
            });
        });
    }

    /**
     * Initialize sorting functionality
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializeSorting(container) {
        const sortButtons = container.querySelectorAll('[data-sort]');
        sortButtons.forEach(button => {
            button.addEventListener('click', () => {
                const sortBy = button.getAttribute('data-sort');
                this.sortTickers(container, sortBy);
                
                // Update button styles
                sortButtons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');
            });
        });
    }

    /**
     * Sort tickers
     * @private
     * @param {HTMLElement} container - Container element
     * @param {string} sortBy - Sort criteria
     */
    sortTickers(container, sortBy) {
        if (!this.originalTickers) return;
        
        let sortedTickers = [...this.originalTickers];
        
        switch (sortBy) {
            case 'ticker':
                sortedTickers.sort((a, b) => a.ticker.localeCompare(b.ticker));
                break;
            case 'change':
                sortedTickers.sort((a, b) => (b.daily_change_percent || 0) - (a.daily_change_percent || 0));
                break;
            case 'date':
                sortedTickers.sort((a, b) => {
                    const dateA = a.last_date ? new Date(a.last_date) : new Date(0);
                    const dateB = b.last_date ? new Date(b.last_date) : new Date(0);
                    return dateB - dateA;
                });
                break;
        }
        
        this.renderTickerList(container, sortedTickers);
    }

    /**
     * Initialize market movers
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializeMarketMovers(container) {
        const periodSelector = container.querySelector('#periodSelect');
        const limitSelector = container.querySelector('#limitSelect');
        const minVolumeSelector = container.querySelector('#minVolumeSelect');
        const refreshButton = container.querySelector('#refreshMovers');
        
        if (periodSelector) {
            periodSelector.addEventListener('change', (e) => {
                this.eventBus.emit('marketmovers:period:changed', { period: e.target.value });
            });
        }
        
        if (limitSelector) {
            limitSelector.addEventListener('change', (e) => {
                this.eventBus.emit('marketmovers:limit:changed', { limit: parseInt(e.target.value) });
            });
        }
        
        if (minVolumeSelector) {
            minVolumeSelector.addEventListener('change', (e) => {
                this.eventBus.emit('marketmovers:volume:changed', { minVolume: parseInt(e.target.value) });
            });
        }
        
        if (refreshButton) {
            refreshButton.addEventListener('click', () => {
                this.eventBus.emit('marketmovers:refresh');
            });
        }
    }

    /**
     * Initialize file manager
     * @private
     * @param {HTMLElement} container - Container element
     */
    initializeFileManager(container) {
        // Add download handlers
        container.addEventListener('click', (e) => {
            if (e.target.matches('[data-download]')) {
                e.preventDefault();
                const file = e.target.getAttribute('data-download');
                const type = e.target.getAttribute('data-type');
                this.eventBus.emit('file:download', { file, type });
            }
        });
    }

    /**
     * Update component data and re-render
     * @param {string} name - Component name
     * @param {object} data - New data
     * @returns {Promise<void>}
     */
    async updateComponent(name, data) {
        const component = this.components.get(name);
        if (!component) {
            console.warn(`Component not found: ${name}`);
            return;
        }
        
        // Merge new data with existing
        const updatedData = { ...component.data, ...data };
        
        // Re-render component
        await this.loadComponent(name, component.selector, updatedData);
    }

    /**
     * Destroy a component
     * @param {string} name - Component name
     */
    destroyComponent(name) {
        const component = this.components.get(name);
        if (!component) return;
        
        // Clear container
        component.container.innerHTML = '';
        
        // Remove from registry
        this.components.delete(name);
        
        // Emit destroy event
        this.eventBus.emit('component:destroyed', { name });
    }

    /**
     * Get loaded component
     * @param {string} name - Component name
     * @returns {object|null} Component data
     */
    getComponent(name) {
        return this.components.get(name) || null;
    }

    /**
     * Check if component is loaded
     * @param {string} name - Component name
     * @returns {boolean}
     */
    isLoaded(name) {
        return this.components.has(name);
    }

    /**
     * Reload all components
     * @returns {Promise<void>}
     */
    async reloadAll() {
        const components = Array.from(this.components.entries());
        
        for (const [name, component] of components) {
            await this.loadComponent(name, component.selector, component.data);
        }
    }

    /**
     * Register custom component initializer
     * @param {string} name - Component name
     * @param {Function} initializer - Initializer function
     */
    registerInitializer(name, initializer) {
        this.initializers.set(name, initializer);
    }

    /**
     * Render index chart
     * @private
     * @param {HTMLElement} container - Container element
     * @param {object} data - Chart data
     */
    renderIndexChart(container, data) {
        const chartContainer = container.querySelector('#indexChart');
        if (!chartContainer || !data) return;

        // Clear placeholder
        chartContainer.innerHTML = '';

        // Create chart using Highcharts
        if (window.Highcharts && data.dates && data.dates.length > 0) {
            Highcharts.stockChart('indexChart', {
                title: { 
                    text: 'ISX Market Indices',
                    style: {
                        fontSize: '18px',
                        fontWeight: '600'
                    }
                },
                rangeSelector: {
                    selected: 1,
                    buttons: [{
                        type: 'week',
                        count: 1,
                        text: '1w'
                    }, {
                        type: 'month',
                        count: 1,
                        text: '1m'
                    }, {
                        type: 'month',
                        count: 3,
                        text: '3m'
                    }, {
                        type: 'month',
                        count: 6,
                        text: '6m'
                    }, {
                        type: 'year',
                        count: 1,
                        text: '1y'
                    }, {
                        type: 'all',
                        text: 'All'
                    }]
                },
                yAxis: [{
                    labels: {
                        align: 'right',
                        x: -3,
                        formatter: function() {
                            return Highcharts.numberFormat(this.value, 0);
                        }
                    },
                    title: {
                        text: 'Index Value'
                    },
                    height: '100%',
                    lineWidth: 2
                }],
                tooltip: {
                    backgroundColor: 'rgba(255, 255, 255, 0.95)',
                    borderColor: '#ccc',
                    borderRadius: 8,
                    borderWidth: 1,
                    shadow: {
                        color: 'rgba(0, 0, 0, 0.15)',
                        offsetX: 1,
                        offsetY: 1,
                        opacity: 0.15,
                        width: 3
                    },
                    style: {
                        fontSize: '13px',
                        fontFamily: 'Arial, sans-serif'
                    },
                    formatter: function() {
                        let tooltipContent = '<b>' + Highcharts.dateFormat('%A, %b %e, %Y', this.x) + '</b><br/>';
                        
                        this.points.forEach(function(point) {
                            const seriesColor = point.series.name === 'ISX60' ? '#1a5f3f' : '#2196F3';
                            tooltipContent += '<span style="color:' + seriesColor + '">●</span> <b>' + point.series.name + '</b>: ' +
                                '<b>' + Highcharts.numberFormat(point.y, 2) + '</b><br/>';
                            
                            // Calculate change from previous point
                            if (point.series.points.length > 1) {
                                const currentIndex = point.series.points.indexOf(point);
                                if (currentIndex > 0) {
                                    const prevPoint = point.series.points[currentIndex - 1];
                                    const change = point.y - prevPoint.y;
                                    const changePercent = (change / prevPoint.y) * 100;
                                    const changeColor = change >= 0 ? '#10b981' : '#ef4444';
                                    tooltipContent += 'Change: <span style="color:' + changeColor + '">' +
                                        (change >= 0 ? '+' : '') + Highcharts.numberFormat(change, 2) + ' (' +
                                        (changePercent >= 0 ? '+' : '') + Highcharts.numberFormat(changePercent, 2) + '%)</span><br/>';
                                }
                            }
                        });
                        
                        return tooltipContent;
                    },
                    shared: true,
                    useHTML: true
                },
                plotOptions: {
                    series: {
                        marker: {
                            enabled: false,
                            states: {
                                hover: {
                                    enabled: true,
                                    radius: 5
                                }
                            }
                        }
                    }
                },
                series: [{
                    name: 'ISX60',
                    data: data.dates.map((date, i) => [
                        new Date(date).getTime(),
                        data.isx60[i]
                    ]),
                    color: '#1a5f3f',
                    lineWidth: 2
                }, {
                    name: 'ISX15',
                    data: data.dates.map((date, i) => [
                        new Date(date).getTime(),
                        data.isx15[i]
                    ]),
                    color: '#2196F3',
                    lineWidth: 2
                }],
                legend: {
                    enabled: true,
                    align: 'center',
                    verticalAlign: 'bottom',
                    layout: 'horizontal'
                }
            });
        } else {
            chartContainer.innerHTML = '<div class="text-center py-5 text-muted">No index data available</div>';
        }
    }

    /**
     * Render ticker list
     * @private
     * @param {HTMLElement} container - Container element
     * @param {Array} tickers - Ticker data
     */
    renderTickerList(container, tickers) {
        const tickerTableBody = container.querySelector('#tickerTableBody');
        if (!tickerTableBody || !tickers) return;

        // Ensure tickers is an array
        if (!Array.isArray(tickers)) {
            console.warn('Tickers data is not an array:', tickers);
            // Try to convert object to array if possible
            if (typeof tickers === 'object') {
                tickers = Object.values(tickers);
            } else {
                tickerTableBody.innerHTML = '<tr><td colspan="5" class="text-muted text-center">Invalid ticker data format</td></tr>';
                return;
            }
        }

        if (tickers.length === 0) {
            tickerTableBody.innerHTML = '<tr><td colspan="5" class="text-muted text-center">No tickers available</td></tr>';
            return;
        }

        // Store original tickers for sorting
        this.originalTickers = [...tickers];
        
        // Sort by date first, then by change percent
        tickers.sort((a, b) => {
            // First sort by date (most recent first)
            const dateA = a.last_date ? new Date(a.last_date) : new Date(0);
            const dateB = b.last_date ? new Date(b.last_date) : new Date(0);
            const dateDiff = dateB - dateA;
            
            // If dates are the same, sort by change percent (highest first)
            if (dateDiff === 0) {
                const changeA = a.daily_change_percent || 0;
                const changeB = b.daily_change_percent || 0;
                return changeB - changeA;
            }
            
            return dateDiff;
        });
        
        // Render ticker table
        let html = '';
        tickers.forEach(ticker => {
            const changePercent = ticker.daily_change_percent || 0;
            const changeClass = changePercent >= 0 ? 'text-success' : 'text-danger';
            const changeSymbol = changePercent >= 0 ? '+' : '';
            const lastDate = ticker.last_date ? new Date(ticker.last_date).toLocaleDateString() : 'N/A';
            
            html += `
                <tr style="cursor: pointer;" data-ticker="${ticker.ticker}">
                    <td><strong>${ticker.ticker}</strong></td>
                    <td class="text-end">${ticker.last_price ? ticker.last_price.toFixed(2) : 'N/A'}</td>
                    <td class="text-end ${changeClass}">${changeSymbol}${changePercent.toFixed(2)}%</td>
                    <td><small>${lastDate}</small></td>
                    <td class="text-center">
                        <div class="sparkline" id="sparkline-${ticker.ticker}" style="width: 60px; height: 20px;"></div>
                    </td>
                </tr>
            `;
        });

        tickerTableBody.innerHTML = html;
        
        // Re-initialize click handlers and sorting
        this.initializeTickerList(container);
        this.initializeSorting(container);
        
        // Render sparklines
        this.renderSparklines(tickers);
    }

    /**
     * Render sparklines for tickers
     * @private
     * @param {Array} tickers - Ticker data
     */
    renderSparklines(tickers) {
        tickers.forEach(ticker => {
            const sparklineContainer = document.querySelector(`#sparkline-${ticker.ticker}`);
            if (!sparklineContainer || !ticker.last_10_days || ticker.last_10_days.length === 0) return;
            
            // Get the last 10 days of prices
            const prices = ticker.last_10_days.slice(-10);
            if (prices.length < 2) return;
            
            // Calculate min and max for scaling
            const min = Math.min(...prices);
            const max = Math.max(...prices);
            const range = max - min || 1;
            
            // Create SVG
            const width = 60;
            const height = 20;
            const padding = 2;
            
            // Create points for the line
            const points = prices.map((price, index) => {
                const x = padding + (index * (width - 2 * padding) / (prices.length - 1));
                const y = padding + (height - 2 * padding) - ((price - min) / range * (height - 2 * padding));
                return `${x},${y}`;
            }).join(' ');
            
            // Determine color based on trend
            const firstPrice = prices[0];
            const lastPrice = prices[prices.length - 1];
            const color = lastPrice >= firstPrice ? '#28a745' : '#dc3545';
            
            // Create SVG element
            const svg = `
                <svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
                    <polyline
                        fill="none"
                        stroke="${color}"
                        stroke-width="1.5"
                        points="${points}"
                    />
                </svg>
            `;
            
            sparklineContainer.innerHTML = svg;
        });
    }

    /**
     * Render market movers
     * @private
     * @param {HTMLElement} container - Container element
     * @param {object} movers - Market movers data
     */
    renderMarketMovers(container, movers) {
        if (!movers) return;

        // Render gainers
        const gainersContainer = container.querySelector('#topGainers');
        if (gainersContainer && movers.top_gainers) {
            gainersContainer.innerHTML = this.renderMoversTable(movers.top_gainers, 'gainers');
        }

        // Render losers
        const losersContainer = container.querySelector('#topLosers');
        if (losersContainer && movers.top_losers) {
            losersContainer.innerHTML = this.renderMoversTable(movers.top_losers, 'losers');
        }

        // Render most active
        const activeContainer = container.querySelector('#mostActive');
        if (activeContainer && movers.most_active) {
            activeContainer.innerHTML = this.renderMoversTable(movers.most_active, 'active');
        }

        // Update metadata display
        const metadataContainer = container.querySelector('#marketMoversMetadata');
        if (metadataContainer && movers.metadata) {
            const metadata = movers.metadata;
            const dateStr = new Date(metadata.as_of_date).toLocaleString();
            metadataContainer.innerHTML = `
                <small class="text-muted">
                    <i class="fas fa-info-circle me-1"></i>
                    As of ${dateStr} | 
                    ${metadata.total_tickers} tickers | 
                    Min volume: ${(metadata.criteria.min_volume / 1000).toFixed(0)}K
                </small>
            `;
        }
    }

    /**
     * Render movers table
     * @private
     * @param {Array} data - Movers data
     * @param {string} type - Type of movers
     * @returns {string} HTML
     */
    renderMoversTable(data, type) {
        if (!data || data.length === 0) {
            return '<div class="text-muted text-center py-3">No data available</div>';
        }

        let html = '<div class="table-responsive"><table class="table table-sm table-hover">';
        html += '<thead><tr><th>Ticker</th><th class="text-end">Price</th><th class="text-end">Change</th><th class="text-end">Volume</th></tr></thead><tbody>';
        
        data.forEach(item => {
            const ticker = item.ticker || 'N/A';
            const companyName = item.company_name || '';
            const price = item.last_price || 0;
            const change = item.change_percent || item.daily_change_percent || 0;
            const volume = item.volume || item.daily_volume || 0;
            const changeClass = change >= 0 ? 'text-success' : 'text-danger';
            const changeIcon = change >= 0 ? 'fa-caret-up' : 'fa-caret-down';
            
            html += `
                <tr style="cursor: pointer;" onclick="eventBus.emit('ticker:selected', {ticker: '${ticker}'})">
                    <td>
                        <strong>${ticker}</strong>
                        ${companyName ? `<br><small class="text-muted">${companyName}</small>` : ''}
                    </td>
                    <td class="text-end">${price.toFixed(2)}</td>
                    <td class="text-end ${changeClass}">
                        <i class="fas ${changeIcon} me-1"></i>
                        ${change >= 0 ? '+' : ''}${change.toFixed(2)}%
                    </td>
                    <td class="text-end">${volume >= 1000000 ? (volume / 1000000).toFixed(1) + 'M' : (volume / 1000).toFixed(0) + 'K'}</td>
                </tr>
            `;
        });
        
        html += '</tbody></table></div>';
        return html;
    }

    /**
     * Render file list
     * @private
     * @param {HTMLElement} container - Container element
     * @param {object} files - Files data
     */
    renderFileList(container, files) {
        if (!files) return;

        // Render Excel files (downloads)
        const excelContainer = container.querySelector('#downloadedFiles');
        if (excelContainer && files.downloads) {
            const excelFiles = files.downloads.map(name => ({
                name: name,
                size: 0, // Size will be calculated on server side if needed
                modified: this.getDateFromFileName(name)
            }));
            excelContainer.innerHTML = this.renderFileTable(excelFiles, 'excel');
        }

        // Separate ticker files from other CSV files
        const tickerFiles = [];
        const otherFiles = [];
        
        if (files.daily_reports) {
            files.daily_reports.forEach(filename => {
                // Ticker files start with ticker symbol
                if (filename.match(/^[A-Z]{3,5}_\d{4}_\d{2}_\d{2}\.csv$/)) {
                    tickerFiles.push(filename);
                } else {
                    otherFiles.push(filename);
                }
            });
        }
        
        // Render ticker CSV files
        const tickerContainer = container.querySelector('#tickerFiles');
        if (tickerContainer) {
            const tickerFileObjs = tickerFiles.map(name => ({
                name: name,
                size: 0,
                modified: this.getDateFromFileName(name)
            }));
            tickerContainer.innerHTML = this.renderFileTable(tickerFileObjs, 'ticker_csv');
        }
        
        // Render other CSV files
        const otherContainer = container.querySelector('#otherReports');
        if (otherContainer) {
            const otherFileObjs = otherFiles.map(name => ({
                name: name,
                size: 0,
                modified: this.getDateFromFileName(name)
            }));
            otherContainer.innerHTML = this.renderFileTable(otherFileObjs, 'other_csv');
        }

        // Update statistics
        this.updateFileStatistics(container, files);
    }
    
    /**
     * Extract date from filename
     * @private
     * @param {string} filename - File name
     * @returns {string} ISO date string
     */
    getDateFromFileName(filename) {
        // Try to extract date from filename patterns
        // Excel: "2025 07 20 ISX Daily Report.xlsx"
        // CSV: "isx_daily_2025_07_20.csv"
        
        let match = filename.match(/(\d{4})\s+(\d{2})\s+(\d{2})/); // Excel pattern
        if (!match) {
            match = filename.match(/(\d{4})_(\d{2})_(\d{2})/); // CSV pattern
        }
        
        if (match) {
            return `${match[1]}-${match[2]}-${match[3]}`;
        }
        
        return new Date().toISOString();
    }

    /**
     * Render file table
     * @private
     * @param {Array} files - Files array
     * @param {string} type - File type
     * @returns {string} HTML
     */
    renderFileTable(files, type) {
        if (!files || files.length === 0) {
            return '<div class="text-muted">No files available</div>';
        }

        let html = '<div class="table-responsive"><table class="table table-sm">';
        html += '<thead><tr><th>Filename</th><th>Size</th><th>Modified</th><th>Action</th></tr></thead><tbody>';
        
        files.forEach(file => {
            html += `
                <tr>
                    <td>${file.name}</td>
                    <td>${this.formatFileSize(file.size)}</td>
                    <td>${new Date(file.modified).toLocaleDateString()}</td>
                    <td>
                        <button class="btn btn-sm btn-outline-primary" data-download="${file.name}" data-type="${type}">
                            <i class="fas fa-download"></i>
                        </button>
                    </td>
                </tr>
            `;
        });
        
        html += '</tbody></table></div>';
        return html;
    }

    /**
     * Format file size
     * @private
     * @param {number} bytes - File size in bytes
     * @returns {string} Formatted size
     */
    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    /**
     * Update file statistics
     * @private
     * @param {HTMLElement} container - Container element
     * @param {object} files - Files data
     */
    updateFileStatistics(container, files) {
        const totalDownloads = container.querySelector('#totalDownloads');
        const totalReports = container.querySelector('#totalReports');
        const totalSize = container.querySelector('#totalSize');
        const lastUpdate = container.querySelector('#lastUpdate');

        if (totalDownloads) {
            // Count from downloads array
            totalDownloads.textContent = files.downloads ? files.downloads.length : 0;
        }

        if (totalReports) {
            // Count from daily_reports array
            totalReports.textContent = files.daily_reports ? files.daily_reports.length : 0;
        }

        if (totalSize && files.total_size) {
            totalSize.textContent = this.formatFileSize(files.total_size);
        }

        if (lastUpdate) {
            // Get most recent date from files
            let mostRecent = null;
            if (files.daily_reports && files.daily_reports.length > 0) {
                // Extract date from first file (most recent)
                const firstFile = files.daily_reports[0];
                const match = firstFile.match(/(\d{4})_(\d{2})_(\d{2})/);
                if (match) {
                    mostRecent = new Date(`${match[1]}-${match[2]}-${match[3]}`);
                }
            }
            
            if (mostRecent) {
                lastUpdate.textContent = mostRecent.toLocaleDateString();
            } else {
                lastUpdate.textContent = 'N/A';
            }
        }
    }
}

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ComponentManager;
}

// Global access for compatibility
window.ComponentManager = ComponentManager;