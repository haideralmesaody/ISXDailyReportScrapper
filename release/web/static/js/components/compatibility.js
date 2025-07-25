/**
 * Initialize compatibility layer
 */
document.addEventListener('DOMContentLoaded', function() {
    // Wait for main app to initialize
    setTimeout(() => {
        if (window.isxApp && window.isxApp.components) {
            ws = window.isxApp.components.websocket.ws;
            isConnected = window.isxApp.components.websocket.isConnected;
            dataUpdateManager = window.isxApp.components.dataUpdateManager;
        }
    }, 100);
});

/**
 * Download file function
 */
function downloadFile(filename) {
    const downloadUrl = `/api/download/${filename}`;
    window.open(downloadUrl, '_blank');
}

/**
 * Initialize files display when DOM is ready
 */
function initFilesDisplay() {
    // Listen for files:update events
    if (window.eventBus) {
        window.eventBus.on('files:update', (files) => {
            if (window.isxApp) {
                window.isxApp.updateFilesDisplay(files);
            }
        });
    }
}

// Initialize files display
document.addEventListener('DOMContentLoaded', initFilesDisplay);

/**
 * Initialize WebSocket connection (legacy function)
 */
function initWebSocket() {
    if (window.isxApp && window.isxApp.components.websocket) {
        window.isxApp.components.websocket.connect();
    }
}

/**
 * Show section (legacy function)
 */
function showSection(sectionName) {
    if (window.isxApp) {
        window.isxApp.showSection(sectionName);
    } else {
        // Fallback for direct calls
        document.querySelectorAll('.command-section').forEach(section => {
            section.classList.remove('active');
        });
        
        const targetSection = document.getElementById(sectionName);
        if (targetSection) {
            targetSection.classList.add('active');
        }
        
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
        });
        
        const navLink = document.querySelector(`[data-section="${sectionName}"]`);
        if (navLink) {
            navLink.classList.add('active');
        }
    }
}

/**
 * Add output to log (legacy function)
 */
function addOutput(message, type = 'info') {
    const outputContainer = document.getElementById('output');
    if (!outputContainer) return;

    const timestamp = new Date().toLocaleTimeString();
    const logClass = `log-${type}`;
    
    const logEntry = document.createElement('div');
    logEntry.innerHTML = `<span class="text-muted">[${timestamp}]</span> <span class="${logClass}">${message}</span>`;
    
    outputContainer.appendChild(logEntry);
    outputContainer.scrollTop = outputContainer.scrollHeight;
}

/**
 * Submit form (legacy function)
 */
async function submitForm(endpoint, formData, outputId) {
    const outputElement = document.getElementById(outputId);
    if (!outputElement) return;

    try {
        const response = await fetch(endpoint, {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            addOutput(`✅ Operation started successfully`, 'success');
        } else {
            const errorText = await response.text();
            addOutput(`❌ Operation failed: ${errorText}`, 'error');
        }
    } catch (error) {
        addOutput(`❌ Network error: ${error.message}`, 'error');
    }
}

/**
 * Load files (legacy function)
 */
async function loadFiles() {
    if (window.isxApp) {
        await window.isxApp.loadFiles();
        return;
    }

    try {
        const response = await fetch('/api/files');
        const data = await response.json();
        
        updateFilesList('downloadsList', data.downloads || []);
        updateFilesList('reportsList', data.reports || []);
        updateFilesList('systemFilesList', data.system_files || []);
    } catch (error) {
        console.error('Failed to load files:', error);
    }
}

/**
 * Update files list (legacy function)
 */
function updateFilesList(elementId, files) {
    const element = document.getElementById(elementId);
    if (!element) return;

    element.innerHTML = '';
    
    if (files.length === 0) {
        element.innerHTML = '<div class="text-muted p-3">No files available</div>';
        return;
    }

    files.forEach(file => {
        const item = document.createElement('div');
        item.className = 'list-group-item d-flex justify-content-between align-items-center';
        item.innerHTML = `
            <div>
                <div class="fw-bold">${file.name || file.filename}</div>
                <small class="text-muted">${file.size || ''} ${file.date || ''}</small>
            </div>
            <button class="btn btn-sm btn-outline-primary download-btn" 
                    onclick="downloadFile('${file.name || file.filename}', '${elementId.replace('List', '')}')">
                <i class="fas fa-download"></i>
            </button>
        `;
        element.appendChild(item);
    });
}

/**
 * Download file (legacy function)
 */
async function downloadFile(filename, type) {
    try {
        const blob = await window.apiService.downloadFile(filename, type);
        
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
        
        addOutput(`📁 Downloaded: ${filename}`, 'success');
    } catch (error) {
        addOutput(`❌ Download failed: ${error.message}`, 'error');
    }
}

/**
 * Refresh tickers (legacy function)
 */
async function refreshTickers() {
    if (window.isxApp) {
        await window.isxApp.loadTickers();
        return;
    }

    try {
        const tickers = await window.apiService.getTickers();
        updateTickerTable(tickers);
        addOutput('📊 Ticker data refreshed', 'success');
    } catch (error) {
        addOutput(`❌ Failed to refresh tickers: ${error.message}`, 'error');
    }
}

/**
 * Update ticker table (legacy function)
 */
function updateTickerTable(tickers) {
    const tableBody = document.getElementById('tickerTableBody');
    if (!tableBody) return;

    tableBody.innerHTML = '';
    
    tickers.forEach(ticker => {
        const row = document.createElement('tr');
        row.className = 'ticker-row';
        row.onclick = () => selectTicker(ticker.ticker);
        
        const changeClass = ticker.change > 0 ? 'price-change-positive' : 
                           ticker.change < 0 ? 'price-change-negative' : 'price-change-neutral';
        
        row.innerHTML = `
            <td>${ticker.ticker}</td>
            <td>${ticker.last_price?.toFixed(2) || 'N/A'}</td>
            <td class="${changeClass}">${ticker.change?.toFixed(2) || '0.00'}</td>
            <td class="${changeClass}">${ticker.change_percent?.toFixed(2) || '0.00'}%</td>
            <td>${ticker.volume?.toLocaleString() || '0'}</td>
            <td>${ticker.trading_status ? 'Active' : 'Inactive'}</td>
        `;
        
        tableBody.appendChild(row);
    });
    
    tickerData = tickers;
}

/**
 * Select ticker (legacy function)
 */
function selectTicker(ticker) {
    currentTicker = ticker;
    
    if (window.isxApp) {
        window.isxApp.components.uiState.setSelection('ticker', ticker);
    }
    
    // Update UI
    document.querySelectorAll('.ticker-row').forEach(row => {
        row.classList.remove('selected');
    });
    
    const selectedRow = Array.from(document.querySelectorAll('.ticker-row'))
        .find(row => row.cells[0].textContent === ticker);
    
    if (selectedRow) {
        selectedRow.classList.add('selected');
    }
    
    // Load ticker chart if in charts section
    if (document.getElementById('tickercharts').classList.contains('active')) {
        loadTickerChart(ticker);
    }
}

/**
 * Load ticker chart (legacy function)
 */
async function loadTickerChart(ticker) {
    try {
        const chartData = await window.apiService.getTickerChart(ticker);
        renderTickerChart(chartData);
        addOutput(`📈 Loaded chart for ${ticker}`, 'success');
    } catch (error) {
        addOutput(`❌ Failed to load chart for ${ticker}: ${error.message}`, 'error');
    }
}

/**
 * Render ticker chart (placeholder - to be fully implemented)
 */
function renderTickerChart(data) {
    // Placeholder for chart rendering logic
    console.log('Rendering ticker chart:', data);
}

/**
 * Load market movers (legacy function)
 */
async function loadMarketMovers() {
    if (window.isxApp) {
        await window.isxApp.loadMarketMovers();
        return;
    }

    try {
        const period = document.getElementById('periodSelect')?.value || '1d';
        const movers = await window.apiService.getMarketMovers({ period });
        updateMarketMoversDisplay(movers);
    } catch (error) {
        addOutput(`❌ Failed to load market movers: ${error.message}`, 'error');
    }
}

/**
 * Update market movers display (placeholder)
 */
function updateMarketMoversDisplay(data) {
    // Placeholder for market movers display logic
    console.log('Updating market movers display:', data);
}

/**
 * Load index chart (legacy function)
 */
async function loadIndexChart() {
    if (window.isxApp) {
        await window.isxApp.loadIndexChart();
        return;
    }

    try {
        const indexData = await window.apiService.getIndexData();
        renderIndexChart(indexData);
    } catch (error) {
        addOutput(`❌ Failed to load index chart: ${error.message}`, 'error');
    }
}

/**
 * Render index chart (placeholder)
 */
function renderIndexChart(data) {
    // Placeholder for index chart rendering logic
    console.log('Rendering index chart:', data);
}

/**
 * Check indices data availability (legacy function)
 */
function checkIndicesDataAvailability() {
    // Placeholder for checking data availability
    console.log('Checking indices data availability');
}

/**
 * Enable all tabs (legacy function)
 */
function enableAllTabs() {
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('disabled');
    });
}

/**
 * Clear output console (legacy function)
 */
function clearOutput() {
    const outputContainer = document.getElementById('output');
    if (outputContainer) {
        outputContainer.innerHTML = '<div class="text-muted">Console cleared...</div>';
    }
}

// Make clearOutput globally available
window.clearOutput = clearOutput;

// Export functions for global access
window.initWebSocket = initWebSocket;
window.showSection = showSection;
window.addOutput = addOutput;
window.submitForm = submitForm;
window.loadFiles = loadFiles;
window.downloadFile = downloadFile;
window.refreshTickers = refreshTickers;
window.selectTicker = selectTicker;
window.loadTickerChart = loadTickerChart;
window.loadMarketMovers = loadMarketMovers;
window.loadIndexChart = loadIndexChart;
window.checkIndicesDataAvailability = checkIndicesDataAvailability;
window.enableAllTabs = enableAllTabs;