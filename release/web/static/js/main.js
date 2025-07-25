/**
 * ISX Analytics Platform - Main Application Entry Point
 * Initializes all components and manages application lifecycle
 */

class ISXApplication {
    constructor() {
        this.components = {};
        this.initialized = false;
        this.config = {
            websocketReconnectAttempts: 5,
            autoRefreshInterval: 30000,
            chartHeight: 400
        };
    }

    /**
     * Initialize the application
     */
    async init() {
        if (this.initialized) {
            window.ISXLogger.warn(LogCategory.SYSTEM, 'Application already initialized');
            return;
        }

        try {
            window.ISXLogger.info(LogCategory.SYSTEM, 'Initializing ISX Analytics Platform...');
            
            // Initialize core components
            this.initCore();
            
            // Initialize services
            this.initServices();
            
            // Initialize UI components
            this.initComponents();
            
            // Skip license check on main app - trust server-side validation
            // License validation is handled by the /license page
            // Users reaching /app have already been validated
            
            // Load layout components
            await this.loadLayoutComponents();
            
            // Set up event listeners
            this.setupEventListeners();
            
            // Connect WebSocket
            this.connectWebSocket();
            
            // Synchronize initial connection status after layout loading
            // This prevents the connection indicator from getting stuck in "Connecting..." state
            setTimeout(() => {
                const currentStatus = this.components.websocket.getConnectionStatus();
                console.log('[Main] Initial connection status check:', currentStatus);
                this.updateConnectionIndicator(currentStatus, !currentStatus);
            }, 100);
            
            // Load initial section
            await this.loadSection('scraper');
            
            // Load initial data
            await this.loadInitialData();
            
            // Set up automatic data refresh
            this.setupAutoRefresh();
            
            this.initialized = true;
            console.log('ISX Analytics Platform initialized successfully');
            
        } catch (error) {
            console.error('Failed to initialize application:', error);
            this.handleInitializationError(error);
        }
    }

    /**
     * Initialize core modules
     */
    initCore() {
        // Initialize UI state management
        this.components.uiState = new UIState();
        
        // Initialize event bus
        this.components.eventBus = eventBus;
        
        // Initialize WebSocket manager
        this.components.websocket = new WebSocketManager();
        
        // Initialize template loader
        this.components.templateLoader = new TemplateLoader();
        
        // Initialize stage adapter
        this.components.stageAdapter = new StageAdapter();
        
        console.log('Core modules initialized');
    }

    /**
     * Initialize services
     */
    initServices() {
        // API service is already globally available
        this.components.api = apiService;
        
        console.log('Services initialized');
    }

    /**
     * Initialize UI components
     */
    initComponents() {
        // Initialize component manager
        this.components.componentManager = new ComponentManager(
            this.components.templateLoader,
            this.components.eventBus,
            this.components.api
        );
        
        // Initialize data update manager
        this.components.dataUpdateManager = new DataUpdateManager(
            this.components.eventBus,
            this.components.uiState
        );
        
        console.log('UI components initialized');
    }

    /**
     * Load layout components
     */
    async loadLayoutComponents() {
        const { componentManager, uiState } = this.components;
        
        try {
            // Load sidebar with active section state
            const sidebarData = {
                scraperActive: uiState.getActiveSection() === 'scraper',
                processorActive: uiState.getActiveSection() === 'processor',
                indexcsvActive: uiState.getActiveSection() === 'indexcsv',
                tickerchartsActive: uiState.getActiveSection() === 'tickercharts',
                marketmoversActive: uiState.getActiveSection() === 'marketmovers',
                filesActive: uiState.getActiveSection() === 'files'
            };
            
            console.log('Loading sidebar with data:', sidebarData);
            await componentManager.loadComponent('sidebar', '#sidebar-container', sidebarData);
            
            console.log('Loading header...');
            await componentManager.loadComponent('header', '#header-container');
            
            console.log('Loading footer...');
            await componentManager.loadComponent('footer', '#footer-container');
            
            console.log('Layout components loaded successfully');
            
            // Check if sidebar is visible
            const sidebarContainer = document.querySelector('#sidebar-container');
            if (sidebarContainer) {
                console.log('Sidebar container found, innerHTML length:', sidebarContainer.innerHTML.length);
                console.log('Sidebar computed styles:', window.getComputedStyle(sidebarContainer).display);
            }
        } catch (error) {
            console.error('Failed to load layout components:', error);
            
            // Use enhanced error display
            if (window.errorDisplay) {
                window.errorDisplay.showError(error, { showDetails: true });
            }
            
            this.addOutput(`Failed to load layout: ${error.message}`, 'error');
            throw error;
        }
    }

    /**
     * Set up event listeners
     */
    setupEventListeners() {
        const { eventBus, uiState, websocket, dataUpdateManager } = this.components;

        // WebSocket connection status
        websocket.onConnectionStatus((connected) => {
            uiState.setConnectionStatus(connected);
            eventBus.emit(ISX_EVENTS.CONNECTION_STATUS, { websocket: connected });
            this.updateConnectionIndicator(connected);
        });

        // WebSocket message handling
        websocket.onMessage('data_update', (message) => {
            dataUpdateManager.handleUpdate(message);
        });

        
        // Handle pipeline progress messages - register both old and new formats
        websocket.onMessage('progress', (message) => {
            this.handlePipelineProgress(message);
        });
        
        websocket.onMessage('pipeline_progress', (message) => {
            this.handlePipelineProgress(message);
        });
        
        websocket.onMessage('pipeline:progress', (message) => {
            this.handlePipelineProgress(message);
        });
        
        websocket.onMessage('pipeline_complete', (message) => {
            this.handlePipelineComplete(message);
        });
        
        websocket.onMessage('pipeline:complete', (message) => {
            this.handlePipelineComplete(message);
        });
        
        // Handle connection messages from server
        websocket.onMessage('connection', (message) => {
            console.log('Connection message received:', message);
            const data = message.data || message;
            if (data.status === 'connected') {
                // Server confirmed connection
                this.updateConnectionIndicator(true);
                this.addOutput('WebSocket connected to server', 'success');
            }
        });
        
        // Handle pipeline status messages - register both old and new formats
        websocket.onMessage('pipeline_status', (message) => {
            const data = message.data || message;
            const transformed = this.components.stageAdapter.transformMessage(data);
            console.log('Pipeline status:', { original: data, transformed });
        });
        
        websocket.onMessage('pipeline:status', (message) => {
            const data = message.data || message;
            const transformed = this.components.stageAdapter.transformMessage(data);
            console.log('Pipeline status:', { original: data, transformed });
            
            // Emit transformed status
            eventBus.emit('pipeline:status', transformed);
            
            // Update individual stages if stages data is present
            if (transformed.stages) {
                for (const [stageId, stageData] of Object.entries(transformed.stages)) {
                    eventBus.emit('pipeline:progress', {
                        stage: stageId,
                        status: stageData.status,
                        progress: stageData.progress || 0,
                        message: stageData.message || ''
                    });
                }
            }
        });
        
        // Handle pipeline reset messages - register both old and new formats
        websocket.onMessage('pipeline_reset', (message) => {
            console.log('Pipeline reset:', message);
            eventBus.emit('pipeline:reset', message.data || message);
            this.addOutput('Pipeline reset - initializing...', 'info');
        });
        
        // Register new format for pipeline reset
        websocket.onMessage('pipeline:reset', (message) => {
            console.log('Pipeline reset (new format):', message);
            eventBus.emit('pipeline:reset', message.data || message);
            this.addOutput('Pipeline reset - initializing...', 'info');
        });
        
        // Handle pipeline start messages (new format only)
        websocket.onMessage('pipeline:start', (message) => {
            console.log('Pipeline start:', message);
            const data = message.data || message;
            const stage = data.stage || 'unknown';
            eventBus.emit('pipeline:start', data);
            this.addOutput(`Starting ${stage} stage...`, 'info');
            
            // Update stage status
            if (data.stage) {
                const stageElements = {
                    'scrape': 'scrape-status',
                    'process': 'process-status',
                    'index': 'index-status',
                    'complete': 'complete-status'
                };
                
                const statusEl = document.getElementById(stageElements[data.stage]);
                if (statusEl) {
                    statusEl.textContent = 'Active';
                    statusEl.className = 'stage-status active';
                    
                    // Remove active class from other stages
                    Object.entries(stageElements).forEach(([key, id]) => {
                        if (key !== data.stage) {
                            const el = document.getElementById(id);
                            if (el && el.classList.contains('active')) {
                                el.textContent = 'Inactive';
                                el.className = 'stage-status';
                            }
                        }
                    });
                }
            }
        });
        
        // Handle error messages
        websocket.onMessage('error', (message) => {
            const data = message.data || message;
            console.error('WebSocket error:', data);
            
            // Use enhanced error display for WebSocket errors
            if (window.errorDisplay) {
                const errorObj = {
                    type: data.error_code || '/errors/websocket',
                    title: data.title || 'WebSocket Error',
                    status: data.status || 500,
                    detail: data.detail || data.message || 'An error occurred',
                    instance: window.location.pathname,
                    stage: data.stage || 'system',
                    hint: data.hint
                };
                
                window.errorDisplay.showError(errorObj, {
                    showDetails: true,
                    dismissDelay: 15000 // WebSocket errors stay longer
                });
            }
            
            // Still log to output console for debugging
            const errorMsg = data.message || 'An error occurred';
            const hint = data.hint || '';
            const stage = data.stage || 'system';
            
            this.addOutput(`[${stage}] ${errorMsg}`, 'error');
            if (hint) {
                this.addOutput(`Hint: ${hint}`, 'warning');
            }
            
            // Emit error event for other components
            eventBus.emit('error:occurred', data);
        });
        
        // Handle status messages
        websocket.onMessage('status', (message) => {
            const data = message.data || message;
            console.log('Status update:', data);
            
            // Update application status
            const status = data.status || 'unknown';
            const statusMessage = data.message || '';
            
            if (statusMessage) {
                this.addOutput(`Status: ${statusMessage}`, 'info');
            }
            
            // Emit status event
            eventBus.emit('status:update', { status, message: statusMessage });
        });
        
        // Handle output messages
        websocket.onMessage('output', (message) => {
            const content = message.data || message.message || message;
            this.addOutput(content, 'info');
        });

        websocket.onMessage('log', (message) => {
            const content = message.data ? message.data.message : message.message;
            const level = message.data ? message.data.level : message.level || 'info';
            
            // Check if this is a scraper output message
            if (content && content.includes('[SCRAPER OUTPUT]')) {
                // Extract the actual scraper message
                const scraperMsg = content.replace('[SCRAPER OUTPUT]', '').trim();
                
                // Parse scraper progress messages
                if (scraperMsg.includes('[DOWNLOAD]') || scraperMsg.includes('[SUCCESS]') || scraperMsg.includes('[NAVIGATE]')) {
                    this.addOutput(scraperMsg, 'info');
                    
                    // Also emit progress event for pipeline visualization
                    this.components.eventBus.emit('pipeline:progress', {
                        stage: 'scrape',
                        status: 'active',
                        message: scraperMsg
                    });
                } else if (scraperMsg.includes('[COMPLETE]') || scraperMsg.includes('Download Complete')) {
                    this.addOutput(scraperMsg, 'success');
                    this.components.eventBus.emit('pipeline:progress', {
                        stage: 'scrape',
                        status: 'completed',
                        message: 'Scraping completed successfully'
                    });
                } else if (scraperMsg.includes('[ERROR]')) {
                    this.addOutput(scraperMsg, 'error');
                    this.components.eventBus.emit('pipeline:progress', {
                        stage: 'scrape',
                        status: 'error',
                        message: scraperMsg
                    });
                } else {
                    this.addOutput(scraperMsg, level);
                }
            } else {
                this.addOutput(content, level);
            }
        });
        
        // Handle info, success, and warning message types
        websocket.onMessage('info', (message) => {
            const content = message.message || message.data || message;
            this.addOutput(content, 'info');
            // Show progress if scraping
            if (message.command === 'scrape' || message.command === 'process') {
                this.updateProgress(content);
            }
        });
        
        websocket.onMessage('success', (message) => {
            const content = message.message || message.data || message;
            this.addOutput(content, 'success');
            // Update progress for successful operations
            if (message.command === 'scrape' || message.command === 'process') {
                this.updateProgress(content, 'success');
            }
        });
        
        websocket.onMessage('warning', (message) => {
            const content = message.message || message.data || message;
            this.addOutput(content, 'warning');
        });

        // UI state changes
        uiState.observe((path, value, oldValue) => {
            if (path === 'activeSection') {
                this.handleSectionChange(value, oldValue);
            }
        });

        // Navigation clicks
        eventBus.on('navigation:click', async ({ section }) => {
            await this.loadSection(section);
        });

        // Form submissions
        eventBus.on('scraper:submit', async (formData) => {
            await this.handleScrapeSubmission(formData);
        });

        eventBus.on('processor:submit', async () => {
            await this.handleProcessSubmission();
        });

        eventBus.on('indexcsv:submit', async () => {
            await this.handleIndexSubmission();
        });

        // Ticker search handler
        this.allTickers = []; // Store all tickers for filtering
        eventBus.on('ticker:search', ({ query }) => {
            this.handleTickerSearch(query);
        });

        // Ticker selection handler  
        eventBus.on('ticker:selected', ({ ticker }) => {
            this.handleTickerSelected(ticker);
        });
        
        // Market movers event handlers
        eventBus.on('marketmovers:period:changed', ({ period }) => {
            this.components.uiState.setSelection('period', period);
            this.loadMarketMovers();
        });
        
        eventBus.on('marketmovers:limit:changed', ({ limit }) => {
            this.components.uiState.setSelection('limit', limit);
            this.loadMarketMovers();
        });
        
        eventBus.on('marketmovers:volume:changed', ({ minVolume }) => {
            this.components.uiState.setSelection('minVolume', minVolume);
            this.loadMarketMovers();
        });
        
        eventBus.on('marketmovers:refresh', () => {
            this.loadMarketMovers();
        });

        // Global refresh button
        document.addEventListener('click', (event) => {
            if (event.target.id === 'globalRefreshBtn' || event.target.closest('#globalRefreshBtn')) {
                this.refreshData();
            }
        });

        console.log('Event listeners set up');
    }

    /**
     * Connect WebSocket
     */
    connectWebSocket() {
        // Let WebSocket manager handle all status updates
        this.components.websocket.connect();
    }

    /**
     * Load initial data
     */
    async loadInitialData() {
        try {
            // Small delay to ensure API is ready
            await new Promise(resolve => setTimeout(resolve, 100));
            
            // Load license status
            const licenseStatus = await this.components.api.getLicenseStatus();
            this.updateLicenseStatus(licenseStatus);

            // Load version info
            const versionInfo = await this.components.api.getVersionInfo();
            this.updateVersionInfo(versionInfo);

            // Load files
            await this.loadFiles();

            console.log('Initial data loaded');
        } catch (error) {
            console.error('Failed to load initial data:', error);
        }
    }

    /**
     * Set up automatic data refresh
     */
    setupAutoRefresh() {
        if (this.components.uiState.getPreference('autoRefresh')) {
            setInterval(() => {
                this.refreshData();
            }, this.config.autoRefreshInterval);
        }
    }

    /**
     * Load a section
     */
    async loadSection(sectionName) {
        const { componentManager, uiState } = this.components;
        
        try {
            // Update UI state
            uiState.setActiveSection(sectionName);
            
            // Load section component
            await componentManager.loadComponent(sectionName, '#section-container');
            
            // Update navigation
            await this.updateNavigation(sectionName);
            
            // Load section-specific data
            await this.loadSectionData(sectionName);
            
            console.log(`Section loaded: ${sectionName}`);
        } catch (error) {
            console.error(`Failed to load section ${sectionName}:`, error);
            
            // Use enhanced error display
            if (window.errorDisplay) {
                window.errorDisplay.showError({
                    type: '/errors/section-load',
                    title: `Failed to load section: ${sectionName}`,
                    status: 500,
                    detail: error.message,
                    instance: window.location.pathname
                });
            }
            
            this.addOutput(`Failed to load section: ${error.message}`, 'error');
        }
    }

    /**
     * Update navigation highlighting
     */
    async updateNavigation(activeSection) {
        const { componentManager, uiState } = this.components;
        
        // Update sidebar with new active state
        const sidebarData = {
            scraperActive: activeSection === 'scraper',
            processorActive: activeSection === 'processor',
            indexcsvActive: activeSection === 'indexcsv',
            tickerchartsActive: activeSection === 'tickercharts',
            marketmoversActive: activeSection === 'marketmovers',
            filesActive: activeSection === 'files'
        };
        
        await componentManager.updateComponent('sidebar', sidebarData);
    }

    /**
     * Handle section change
     */
    handleSectionChange(newSection, oldSection) {
        // This is now handled by loadSection
        console.log(`Section changed from ${oldSection} to ${newSection}`);
    }

    /**
     * Show specific section
     */
    showSection(sectionName) {
        // Use loadSection instead
        this.loadSection(sectionName);
    }

    /**
     * Load section-specific data
     */
    async loadSectionData(section) {
        try {
            switch (section) {
                case 'dashboard':
                    await this.loadIndexChart();
                    break;
                case 'indexcsv':
                    await this.loadIndexChart();
                    break;
                case 'tickercharts':
                    await this.loadTickers();
                    break;
                case 'marketmovers':
                    await this.loadMarketMovers();
                    break;
                case 'files':
                    await this.loadFiles();
                    break;
            }
        } catch (error) {
            console.error(`Failed to load data for section ${section}:`, error);
        }
    }

    /**
     * Handle pipeline messages
     */
    handlePipelineMessage(message) {
        // Update pipeline visualization
        this.updatePipelineStatus(message);
        
        // Add to output log
        if (message.stage && message.status) {
            this.addOutput(`Pipeline ${message.stage}: ${message.status}`, 'info');
        }
    }
    
    /**
     * Handle pipeline progress messages
     */
    handlePipelineProgress(message) {
        const data = message.data || message;
        
        // Transform the message using stage adapter
        const transformed = this.components.stageAdapter.transformMessage(data);
        
        // Log for debugging
        console.log('Pipeline progress:', { original: data, transformed });
        
        // Display progress message if available
        if (transformed.stage && transformed.message) {
            this.addOutput(`[${transformed.stage}] ${transformed.message}`, 'info');
        }
        
        // Extract stage status from the message
        let stageStatus = 'active';
        if (transformed.status) {
            stageStatus = transformed.status;
        } else if (transformed.progress >= 100) {
            stageStatus = 'completed';
        }
        
        // Update pipeline visualization with transformed data
        this.components.eventBus.emit('pipeline:progress', {
            stage: transformed.stage,
            status: stageStatus,
            progress: transformed.progress || 0,
            message: transformed.message || '',
            metadata: transformed.metadata || {}
        });
    }
    
    /**
     * Handle pipeline complete messages
     */
    handlePipelineComplete(message) {
        const data = message.data || message;
        this.addOutput(`Pipeline completed: ${data.status}`, data.status === 'completed' ? 'success' : 'error');
        
        // Update pipeline visualization
        this.components.eventBus.emit('pipeline:complete', data);
    }


    /**
     * Handle scrape form submission
     */
    async handleScrapeSubmission(formData) {
        console.log('handleScrapeSubmission called with formData:', formData);
        const params = {
            command: 'scrape',
            args: {
                mode: formData.get('mode'),
                from: formData.get('from'),
                to: formData.get('to'),
                headless: formData.get('headless')
            }
        };
        console.log('Scraping params:', params);

        try {
            await this.components.api.startScraping(params);
            this.addOutput('Scraping started...', 'info');
        } catch (error) {
            console.error('Scraping error:', error);
            
            // Use enhanced error display
            if (window.errorDisplay && error instanceof APIError) {
                window.errorDisplay.showError(error);
            } else if (window.errorDisplay) {
                window.errorDisplay.showError({
                    type: '/errors/operation',
                    title: 'Scraping Failed',
                    status: 500,
                    detail: error.message,
                    instance: '/api/scrape'
                });
            }
            
            this.addOutput(`Scraping failed: ${error.message}`, 'error');
        }
    }

    /**
     * Handle process form submission
     */
    async handleProcessSubmission() {
        try {
            await this.components.api.startProcessing();
            this.addOutput('Processing started...', 'info');
        } catch (error) {
            console.error('Processing error:', error);
            
            // Use enhanced error display
            if (window.errorDisplay && error instanceof APIError) {
                window.errorDisplay.showError(error);
            } else if (window.errorDisplay) {
                window.errorDisplay.showError({
                    type: '/errors/operation',
                    title: 'Processing Failed',
                    status: 500,
                    detail: error.message,
                    instance: '/api/process'
                });
            }
            
            this.addOutput(`Processing failed: ${error.message}`, 'error');
        }
    }

    /**
     * Handle index extraction submission
     */
    async handleIndexSubmission() {
        try {
            await this.components.api.startIndexExtraction();
            this.addOutput('Index extraction started...', 'info');
        } catch (error) {
            console.error('Index extraction error:', error);
            
            // Use enhanced error display
            if (window.errorDisplay && error instanceof APIError) {
                window.errorDisplay.showError(error);
            } else if (window.errorDisplay) {
                window.errorDisplay.showError({
                    type: '/errors/operation',
                    title: 'Index Extraction Failed',
                    status: 500,
                    detail: error.message,
                    instance: '/api/indexcsv'
                });
            }
            
            this.addOutput(`Index extraction failed: ${error.message}`, 'error');
        }
    }

    /**
     * Refresh data
     */
    async refreshData() {
        // Only refresh if auto-refresh is enabled
        if (!this.components.uiState.getPreference('autoRefresh')) {
            return;
        }

        try {
            await this.loadFiles();
            this.components.eventBus.emit(ISX_EVENTS.DATA_REFRESH, { type: 'auto' });
        } catch (error) {
            console.error('Auto-refresh failed:', error);
        }
    }

    /**
     * Load files data
     */
    async loadFiles() {
        try {
            const files = await this.components.api.getFiles();
            this.updateFilesDisplay(files);
        } catch (error) {
            console.error('Failed to load files:', error);
        }
    }

    /**
     * Load tickers data
     */
    async loadTickers() {
        try {
            const tickers = await this.components.api.getTickers();
            this.updateTickersDisplay(tickers);
        } catch (error) {
            console.error('Failed to load tickers:', error);
        }
    }

    /**
     * Load market movers
     */
    async loadMarketMovers() {
        try {
            const period = this.components.uiState.getSelection('period') || '1d';
            const limit = this.components.uiState.getSelection('limit') || 10;
            const minVolume = this.components.uiState.getSelection('minVolume') || 1000;
            
            const params = {
                period: period,
                limit: limit,
                min_volume: minVolume
            };
            
            console.log('Loading market movers with params:', params);
            const movers = await this.components.api.getMarketMovers(params);
            console.log('Market movers data:', movers);
            this.updateMarketMoversDisplay(movers);
        } catch (error) {
            console.error('Failed to load market movers:', error);
            this.addOutput(`Failed to load market movers: ${error.message}`, 'error');
        }
    }

    /**
     * Load index chart
     */
    async loadIndexChart() {
        try {
            const indexData = await this.components.api.getIndexData();
            this.updateIndexChart(indexData);
        } catch (error) {
            console.error('Failed to load index chart:', error);
        }
    }

    /**
     * Update connection indicator
     */
    updateConnectionIndicator(connected, connecting = false) {
        const indicator = document.getElementById('connectionStatus');
        const connectionText = document.getElementById('connectionText');
        
        if (indicator) {
            let statusClass;
            if (connecting) {
                statusClass = 'status-connecting';
            } else {
                statusClass = connected ? 'status-connected' : 'status-disconnected';
            }
            indicator.className = `status-indicator ${statusClass}`;
        }
        
        if (connectionText) {
            let text;
            if (connecting) {
                text = 'Connecting...';
            } else {
                text = connected ? 'Connected' : 'Disconnected';
            }
            connectionText.textContent = text;
        }
    }

    /**
     * Update license status
     */
    updateLicenseStatus(status) {
        const element = document.getElementById('licenseStatus');
        if (element && status) {
            // Update license status display
            element.innerHTML = this.getLicenseStatusHTML(status);
        }
    }

    /**
     * Get license status HTML
     */
    getLicenseStatusHTML(status) {
        // Handle LicenseStatusResponse format
        const licenseStatus = status.license_status || status.status;
        const daysLeft = status.days_left || 0;
        const message = status.message || '';
        
        // Determine if license is active based on status
        const isActive = licenseStatus === 'active' || licenseStatus === 'valid';
        
        if (isActive) {
            const daysText = daysLeft === 1 ? 'day' : 'days';
            const alertClass = daysLeft <= 7 ? 'alert-warning' : 'alert-success';
            const icon = daysLeft <= 7 ? 'fa-exclamation-triangle' : 'fa-check-circle';
            
            return `
                <div class="alert ${alertClass} alert-sm mb-0">
                    <div class="d-flex align-items-center">
                        <i class="fas ${icon} me-2"></i>
                        <div>
                            <strong>License Active</strong> - ${daysLeft} ${daysText} remaining
                            ${daysLeft <= 7 ? '<br><small>Contact Iraqi Investor to renew</small>' : ''}
                        </div>
                    </div>
                </div>
            `;
        } else {
            // Map status to display text
            let displayStatus = licenseStatus;
            switch(licenseStatus) {
                case 'not_activated':
                    displayStatus = 'No License';
                    break;
                case 'expired':
                    displayStatus = 'License Expired';
                    break;
                case 'critical':
                    displayStatus = 'License Critical';
                    break;
                case 'warning':
                    displayStatus = 'License Warning';
                    break;
                case 'error':
                    displayStatus = 'License Error';
                    break;
                default:
                    displayStatus = licenseStatus || 'No License';
            }
            
            return `
                <div class="alert alert-danger alert-sm mb-0">
                    <div class="d-flex align-items-center">
                        <i class="fas fa-times-circle me-2"></i>
                        <div>
                            <strong>${displayStatus}</strong><br>
                            <small>${message || 'Please activate a license'}</small>
                        </div>
                    </div>
                </div>
            `;
        }
    }

    /**
     * Update version info
     */
    updateVersionInfo(info) {
        const element = document.getElementById('versionInfo');
        if (element && info) {
            element.textContent = `v${info.version || 'unknown'}`;
        }
    }

    /**
     * Add output message
     */
    addOutput(message, type = 'info') {
        const outputElement = document.getElementById('output');
        if (!outputElement) return;
        
        const timestamp = new Date().toLocaleTimeString();
        const typeClass = {
            'info': 'text-primary',
            'success': 'text-success',
            'warning': 'text-warning',
            'error': 'text-danger'
        }[type] || 'text-muted';
        
        const messageHtml = `
            <div class="output-line">
                <span class="text-muted">[${timestamp}]</span>
                <span class="${typeClass}">${this.escapeHtml(message)}</span>
            </div>
        `;
        
        outputElement.insertAdjacentHTML('beforeend', messageHtml);
        
        // Auto-scroll to bottom
        outputElement.scrollTop = outputElement.scrollHeight;
        
        // Emit event for other components
        this.components.eventBus.emit('ui:output', { message, type, timestamp });
    }
    
    /**
     * Update progress indicator for scraping/processing
     */
    updateProgress(message, status = 'info') {
        const progressElement = document.getElementById('processProgress');
        const progressBar = document.querySelector('#processProgress .progress-bar');
        const progressText = document.querySelector('#processProgress .progress-text');
        const progressDetails = document.getElementById('progressDetails');
        
        if (!progressElement) return;
        
        // Show progress indicator
        progressElement.style.display = 'block';
        
        // Update progress details
        if (progressDetails) {
            progressDetails.textContent = message;
        }
        
        // Update progress bar based on message content
        if (message.includes('Starting')) {
            progressBar.style.width = '10%';
            progressText.textContent = 'Starting...';
        } else if (message.includes('Downloading') || message.includes('download')) {
            progressBar.style.width = '30%';
            progressText.textContent = 'Downloading...';
        } else if (message.includes('Processing') || message.includes('process')) {
            progressBar.style.width = '60%';
            progressText.textContent = 'Processing...';
        } else if (message.includes('Index extraction') || message.includes('indices')) {
            progressBar.style.width = '80%';
            progressText.textContent = 'Extracting indices...';
        } else if (message.includes('completed') || status === 'success') {
            progressBar.style.width = '100%';
            progressText.textContent = 'Complete!';
            progressBar.classList.remove('progress-bar-animated');
            progressBar.classList.add('bg-success');
            
            // Hide progress after 3 seconds
            setTimeout(() => {
                progressElement.style.display = 'none';
                progressBar.style.width = '0%';
                progressBar.classList.add('progress-bar-animated');
                progressBar.classList.remove('bg-success');
            }, 3000);
        }
    }

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#39;'
        };
        return String(text).replace(/[&<>"']/g, char => map[char]);
    }

    /**
     * Handle initialization error
     */
    handleInitializationError(error) {
        // Use enhanced error display if available
        if (window.errorDisplay) {
            window.errorDisplay.showError({
                type: '/errors/initialization',
                title: 'Application Initialization Failed',
                status: 500,
                detail: error.message,
                instance: window.location.pathname,
                hint: 'Try reloading the page or contact support if the problem persists'
            }, {
                autoDismiss: false,
                showDetails: true
            });
        } else {
            // Fallback to basic error display
            const errorElement = document.createElement('div');
            errorElement.className = 'alert alert-danger m-3';
            errorElement.innerHTML = `
                <h5>Application Initialization Failed</h5>
                <p>${error.message}</p>
                <button onclick="location.reload()" class="btn btn-danger btn-sm">Reload Page</button>
            `;
            document.body.prepend(errorElement);
        }
    }

    /**
     * Update pipeline status
     */
    updatePipelineStatus(message) {
        // Pipeline status is handled by the pipeline component
        this.components.eventBus.emit('pipeline:update', message);
    }

    /**
     * Update files display
     */
    updateFilesDisplay(files) {
        console.log('[DEBUG] Updating files display:', files);
        
        if (!files) {
            console.error('No files data provided');
            return;
        }

        try {
            // Update downloads section
            const downloadsContainer = document.getElementById('downloadedFiles');
            if (downloadsContainer) {
                this.renderFileList(downloadsContainer, files.downloads || [], 'Excel Files');
            }

            // Update daily reports section
            const reportsContainer = document.getElementById('otherReports');
            if (reportsContainer) {
                this.renderFileList(reportsContainer, files.daily_reports || [], 'CSV Reports');
            }

            // Update ticker files section (filter for ticker-specific CSVs)
            const tickerContainer = document.getElementById('tickerFiles');
            if (tickerContainer) {
                const tickerFiles = (files.daily_reports || []).filter(file => 
                    file.includes('_trading_history.csv') || file.includes('isx_daily_')
                );
                this.renderFileList(tickerContainer, tickerFiles, 'Ticker Files');
            }

            // Update statistics
            this.updateFileStatistics(files);
            
        } catch (error) {
            console.error('Error updating files display:', error);
        }
    }

    /**
     * Render file list in container
     */
    renderFileList(container, files, title) {
        if (!container) return;

        if (!files || files.length === 0) {
            container.innerHTML = `
                <div class="text-center py-4 text-muted">
                    <i class="fas fa-folder-open me-2"></i>No ${title} found
                </div>
            `;
            return;
        }

        const fileList = files.map(file => `
            <div class="d-flex justify-content-between align-items-center py-2 border-bottom">
                <div>
                    <i class="fas fa-file-excel text-success me-2"></i>
                    <small>${file}</small>
                </div>
                <button class="btn btn-sm btn-outline-primary" onclick="downloadFile('${file}')">
                    <i class="fas fa-download"></i>
                </button>
            </div>
        `).join('');

        container.innerHTML = fileList;
    }

    /**
     * Update file statistics display
     */
    updateFileStatistics(files) {
        if (!files) return;

        const totalDownloads = document.getElementById('totalDownloads');
        const totalReports = document.getElementById('totalReports');
        const totalSize = document.getElementById('totalSize');
        const lastUpdate = document.getElementById('lastUpdate');

        if (totalDownloads) {
            totalDownloads.textContent = (files.downloads || []).length;
        }

        if (totalReports) {
            totalReports.textContent = (files.daily_reports || []).length;
        }

        if (totalSize) {
            const sizeMB = Math.round((files.total_size || 0) / (1024 * 1024) * 100) / 100;
            totalSize.textContent = `${sizeMB} MB`;
        }

        if (lastUpdate) {
            lastUpdate.textContent = new Date().toLocaleTimeString();
        }
    }
    /**
     * Update tickers display
     */
    updateTickersDisplay(tickersResponse) {
        // Extract tickers array from response
        const tickers = tickersResponse.tickers || tickersResponse;
        // Store all tickers for search functionality
        this.allTickers = Array.isArray(tickers) ? tickers : [];
        // Tickers display is handled by the tickercharts component
        this.components.eventBus.emit('tickers:update', tickers);
    }

    /**
     * Handle ticker search
     */
    handleTickerSearch(query) {
        if (!query || query.length < 2) {
            // Hide search results
            const searchResults = document.querySelector('#searchResults');
            if (searchResults) {
                searchResults.style.display = 'none';
            }
            return;
        }

        // Filter tickers based on query
        const filteredTickers = this.allTickers.filter(ticker => {
            const tickerSymbol = ticker.ticker ? ticker.ticker.toLowerCase() : '';
            const companyName = ticker.company_name ? ticker.company_name.toLowerCase() : '';
            const searchQuery = query.toLowerCase();
            return tickerSymbol.includes(searchQuery) || companyName.includes(searchQuery);
        });

        // Display search results
        this.displaySearchResults(filteredTickers, query);
    }

    /**
     * Display search results
     */
    displaySearchResults(tickers, query) {
        const searchResults = document.querySelector('#searchResults');
        if (!searchResults) return;

        if (tickers.length === 0) {
            searchResults.innerHTML = '<div class="dropdown-item text-muted">No results found</div>';
            searchResults.style.display = 'block';
            return;
        }

        // Build results HTML
        let html = '';
        tickers.slice(0, 10).forEach(ticker => { // Limit to 10 results
            html += `
                <a href="#" class="dropdown-item search-result-item" data-ticker="${ticker.ticker}">
                    <strong>${ticker.ticker}</strong>
                    ${ticker.company_name ? `<small class="text-muted ms-2">${ticker.company_name}</small>` : ''}
                </a>
            `;
        });

        searchResults.innerHTML = html;
        searchResults.style.display = 'block';

        // Add click handlers to search results
        searchResults.querySelectorAll('.search-result-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const ticker = item.getAttribute('data-ticker');
                this.handleTickerSelected(ticker);
                
                // Clear search and hide results
                const searchInput = document.querySelector('#tickerSearch');
                if (searchInput) {
                    searchInput.value = '';
                }
                searchResults.style.display = 'none';
            });
        });
    }

    /**
     * Handle ticker selection
     */
    async handleTickerSelected(ticker) {
        try {
            // Load ticker chart
            const chartData = await this.components.api.getTickerChart(ticker);
            this.displayTickerChart(ticker, chartData);
        } catch (error) {
            console.error(`Failed to load chart for ${ticker}:`, error);
            this.addOutput(`Failed to load chart for ${ticker}: ${error.message}`, 'error');
        }
    }

    /**
     * Display ticker chart
     */
    displayTickerChart(ticker, chartData) {
        const chartContainer = document.querySelector('#tickerChart');
        if (!chartContainer) return;

        // Clear any existing chart
        chartContainer.innerHTML = '';

        if (!chartData || !chartData.dates || chartData.dates.length === 0) {
            chartContainer.innerHTML = `<div class="alert alert-info">No data available for ${ticker}</div>`;
            return;
        }

        // Create candlestick chart using Highcharts
        if (window.Highcharts) {
            const ohlcData = [];
            const volumeData = [];

            // Prepare data for Highcharts
            chartData.dates.forEach((date, i) => {
                const timestamp = new Date(date).getTime();
                ohlcData.push([
                    timestamp,
                    chartData.open[i],
                    chartData.high[i],
                    chartData.low[i],
                    chartData.close[i]
                ]);
                volumeData.push([timestamp, chartData.volume[i]]);
            });

            Highcharts.stockChart('tickerChart', {
                rangeSelector: {
                    selected: 1
                },
                title: {
                    text: `${ticker} Price History`
                },
                yAxis: [{
                    labels: {
                        align: 'right',
                        x: -3,
                        formatter: function() {
                            return Highcharts.numberFormat(this.value, 0) + ' IQD';
                        }
                    },
                    title: {
                        text: 'Price (IQD)'
                    },
                    height: '60%',
                    lineWidth: 2,
                    resize: {
                        enabled: true
                    }
                }, {
                    labels: {
                        align: 'right',
                        x: -3,
                        formatter: function() {
                            return Highcharts.numberFormat(this.value, 0);
                        }
                    },
                    title: {
                        text: 'Volume'
                    },
                    top: '65%',
                    height: '35%',
                    offset: 0,
                    lineWidth: 2
                }],
                tooltip: {
                    split: true,
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
                        const points = this.points || [this.point];
                        let tooltipContent = '<b>' + Highcharts.dateFormat('%A, %b %e, %Y', this.x || points[0].x) + '</b><br/>';
                        
                        points.forEach(function(point) {
                            if (point.series.type === 'candlestick') {
                                tooltipContent += '<span style="color:' + point.color + '">●</span> <b>' + point.series.name + '</b><br/>';
                                tooltipContent += 'Open: <b>' + Highcharts.numberFormat(point.point.open, 2) + ' IQD</b><br/>';
                                tooltipContent += 'High: <b>' + Highcharts.numberFormat(point.point.high, 2) + ' IQD</b><br/>';
                                tooltipContent += 'Low: <b>' + Highcharts.numberFormat(point.point.low, 2) + ' IQD</b><br/>';
                                tooltipContent += 'Close: <b>' + Highcharts.numberFormat(point.point.close, 2) + ' IQD</b><br/>';
                                const change = point.point.close - point.point.open;
                                const changePercent = (change / point.point.open) * 100;
                                const changeColor = change >= 0 ? '#10b981' : '#ef4444';
                                tooltipContent += 'Change: <span style="color:' + changeColor + '"><b>' + 
                                    (change >= 0 ? '+' : '') + Highcharts.numberFormat(change, 2) + ' IQD (' +
                                    (changePercent >= 0 ? '+' : '') + Highcharts.numberFormat(changePercent, 2) + '%)</b></span><br/>';
                            } else if (point.series.type === 'column') {
                                tooltipContent += '<br/><span style="color:' + point.color + '">●</span> <b>' + point.series.name + '</b>: ' +
                                    '<b>' + Highcharts.numberFormat(point.y, 0) + '</b>';
                            }
                        });
                        
                        return tooltipContent;
                    },
                    shared: true,
                    useHTML: true
                },
                series: [{
                    type: 'candlestick',
                    name: ticker,
                    data: ohlcData,
                    color: '#ef4444',
                    upColor: '#10b981',
                    lineColor: '#ef4444',
                    upLineColor: '#10b981'
                }, {
                    type: 'column',
                    name: 'Volume',
                    data: volumeData,
                    yAxis: 1,
                    color: 'rgba(33, 150, 243, 0.5)'
                }]
            });
        }
    }

    /**
     * Update market movers display
     */
    updateMarketMoversDisplay(movers) {
        // Market movers display is handled by the marketmovers component
        this.components.eventBus.emit('marketmovers:update', movers);
    }

    /**
     * Update index chart
     */
    updateIndexChart(data) {
        // Index chart is handled by the indexcsv component
        this.components.eventBus.emit('indexchart:update', data);
    }
    
    /**
     * Check license validity
     */
    async checkLicense() {
        try {
            const status = await this.components.api.getLicenseStatus();
            return status && status.is_valid;
        } catch (error) {
            console.error('License check failed:', error);
            return false;
        }
    }
    
    /**
     * Show license page
     */
    async showLicensePage() {
        // Clear the page
        document.body.innerHTML = `
            <div id="header-container"></div>
            <div id="section-container"></div>
            <div id="footer-container"></div>
        `;
        
        // Load license component
        await this.components.componentManager.loadComponent('license', '#section-container');
        
        // Set up license page event handlers
        this.setupLicensePageHandlers();
    }
    
    /**
     * Set up license page handlers
     */
    setupLicensePageHandlers() {
        // Check license status
        this.checkAndShowLicenseStatus();
        
        // Handle license form submission
        const form = document.getElementById('licenseActivationForm');
        if (form) {
            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                await this.handleLicenseActivation();
            });
        }
    }
    
    /**
     * Check and show license status
     */
    async checkAndShowLicenseStatus() {
        const checkingState = document.getElementById('licenseCheckingState');
        const validState = document.getElementById('licenseValidState');
        const invalidState = document.getElementById('licenseInvalidState');
        
        try {
            const status = await this.components.api.getLicenseStatus();
            
            if (checkingState) checkingState.classList.add('d-none');
            
            if (status && status.is_valid) {
                // Show valid state
                if (validState) {
                    validState.classList.remove('d-none');
                    const daysElement = document.getElementById('daysRemaining');
                    if (daysElement) {
                        daysElement.textContent = status.days_left || 0;
                    }
                }
                
                // Start countdown
                this.startLicenseCountdown();
            } else {
                // Show invalid state
                if (invalidState) {
                    invalidState.classList.remove('d-none');
                    const errorElement = document.getElementById('licenseErrorMessage');
                    if (errorElement && status) {
                        errorElement.textContent = status.message || 'No valid license found';
                    }
                }
            }
        } catch (error) {
            // Show invalid state on error
            if (checkingState) checkingState.classList.add('d-none');
            if (invalidState) {
                invalidState.classList.remove('d-none');
                const errorElement = document.getElementById('licenseErrorMessage');
                if (errorElement) {
                    errorElement.textContent = 'Failed to check license status';
                }
            }
        }
    }
    
    /**
     * Start license countdown
     */
    startLicenseCountdown() {
        let count = 5;
        const countdownElement = document.getElementById('countdown');
        
        const interval = setInterval(() => {
            count--;
            if (countdownElement) {
                countdownElement.textContent = count;
            }
            
            if (count <= 0) {
                clearInterval(interval);
                this.proceedToMainApp();
            }
        }, 1000);
        
        // Store interval so it can be cancelled
        this.countdownInterval = interval;
    }
    
    /**
     * Skip license wait
     */
    skipLicenseWait() {
        if (this.countdownInterval) {
            clearInterval(this.countdownInterval);
        }
        this.proceedToMainApp();
    }
    
    /**
     * Proceed to main app
     */
    proceedToMainApp() {
        // Reload the page to reinitialize with full layout
        window.location.reload();
    }
    
    /**
     * Handle license activation
     */
    async handleLicenseActivation() {
        const licenseKey = document.getElementById('licenseKey').value.trim();
        if (!licenseKey) return;
        
        const activatingState = document.getElementById('licenseActivatingState');
        const invalidState = document.getElementById('licenseInvalidState');
        
        // Show activating state
        if (invalidState) invalidState.classList.add('d-none');
        if (activatingState) activatingState.classList.remove('d-none');
        
        try {
            const result = await this.components.api.activateLicense(licenseKey);
            
            if (result && result.success) {
                // License activated successfully
                await this.checkAndShowLicenseStatus();
            } else {
                // Show error
                if (activatingState) activatingState.classList.add('d-none');
                if (invalidState) {
                    invalidState.classList.remove('d-none');
                    const errorElement = document.getElementById('licenseErrorMessage');
                    if (errorElement) {
                        errorElement.textContent = result.message || 'License activation failed';
                    }
                }
            }
        } catch (error) {
            // Show error
            if (activatingState) activatingState.classList.add('d-none');
            if (invalidState) {
                invalidState.classList.remove('d-none');
                const errorElement = document.getElementById('licenseErrorMessage');
                if (errorElement) {
                    errorElement.textContent = 'License activation failed: ' + error.message;
                }
            }
        }
    }
}

// Global error handling
window.addEventListener('error', (event) => {
    if (window.apiService && window.apiService.reportClientError) {
        window.apiService.reportClientError('JAVASCRIPT_ERROR', event.message, {
            filename: event.filename,
            lineno: event.lineno,
            colno: event.colno,
            stack: event.error ? event.error.stack : null
        });
    }
    console.error('Global JavaScript error:', event);
});

// Global unhandled promise rejection handling
window.addEventListener('unhandledrejection', (event) => {
    if (window.apiService && window.apiService.reportClientError) {
        window.apiService.reportClientError('UNHANDLED_PROMISE_REJECTION', event.reason.toString(), {
            stack: event.reason.stack
        });
    }
    console.error('Unhandled promise rejection:', event);
});

// Initialize application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.isxApp = new ISXApplication();
    window.isxApp.init();
});

// Export for module use
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ISXApplication;
}

// Global access
window.ISXApplication = ISXApplication;