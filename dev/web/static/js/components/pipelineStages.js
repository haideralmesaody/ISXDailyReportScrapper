/**
 * Pipeline Stages Component
 * Shows visual representation of the data processing pipeline
 */
class PipelineStages {
    constructor() {
        this.stages = [
            { id: 'scrape', name: 'Scraping', icon: 'fa-download', status: 'idle' },
            { id: 'process', name: 'Processing', icon: 'fa-cogs', status: 'idle' },
            { id: 'index', name: 'Index Extraction', icon: 'fa-chart-line', status: 'idle' },
            { id: 'complete', name: 'Complete', icon: 'fa-check-circle', status: 'idle' }
        ];
        this.container = null;
    }

    /**
     * Initialize the component
     */
    init(container) {
        this.container = container;
        if (this.container) {
            this.render();
            this.setupEventListeners();
        }
    }

    /**
     * Set up event listeners
     */
    setupEventListeners() {
        // Listen for pipeline events
        if (window.app && window.app.components && window.app.components.eventBus) {
            const eventBus = window.app.components.eventBus;
            
            eventBus.on('pipeline:start', (data) => {
                this.updateStage(data.stage, 'active');
            });
            
            eventBus.on('pipeline:progress', (data) => {
                if (data.stage) {
                    // Use the status from the data if available, otherwise default to active
                    const status = data.status || 'active';
                    this.updateStage(data.stage, status);
                    
                    // If status is completed, activate next stage
                    if (status === 'completed') {
                        this.activateNextStage(data.stage);
                    }
                }
            });
            
            eventBus.on('pipeline:complete', (data) => {
                this.updateStage(data.stage, 'completed');
                // Move to next stage if applicable
                this.activateNextStage(data.stage);
            });
            
            eventBus.on('pipeline:error', (data) => {
                this.updateStage(data.stage, 'error');
            });
            
            eventBus.on('pipeline:reset', () => {
                this.resetAllStages();
            });
        }
    }

    /**
     * Render the pipeline visualization
     */
    render() {
        const html = `
            <div class="pipeline-stages d-flex justify-content-between align-items-center">
                ${this.stages.map((stage, index) => this.renderStage(stage, index)).join('')}
            </div>
        `;
        this.container.innerHTML = html;
    }

    /**
     * Render a single stage
     */
    renderStage(stage, index) {
        const isLast = index === this.stages.length - 1;
        return `
            <div class="pipeline-stage text-center" data-stage="${stage.id}">
                <div class="stage-icon ${stage.status}" title="${stage.name}">
                    <i class="fas ${stage.icon} fa-2x"></i>
                </div>
                <small class="d-block mt-1">${stage.name}</small>
            </div>
            ${!isLast ? '<div class="pipeline-connector"></div>' : ''}
        `;
    }

    /**
     * Update stage status
     */
    updateStage(stageId, status) {
        const stage = this.stages.find(s => s.id === stageId);
        if (stage) {
            stage.status = status;
            const stageElement = this.container.querySelector(`[data-stage="${stageId}"] .stage-icon`);
            if (stageElement) {
                stageElement.className = `stage-icon ${status}`;
                
                // Add animation for active stage
                if (status === 'active') {
                    stageElement.classList.add('pulse');
                } else {
                    stageElement.classList.remove('pulse');
                }
            }
        }
    }

    /**
     * Activate next stage in the pipeline
     */
    activateNextStage(currentStageId) {
        const currentIndex = this.stages.findIndex(s => s.id === currentStageId);
        if (currentIndex >= 0 && currentIndex < this.stages.length - 1) {
            const nextStage = this.stages[currentIndex + 1];
            
            // Auto-activate next stage with a small delay
            setTimeout(() => {
                this.updateStage(nextStage.id, 'active');
            }, 500);
        }
    }

    /**
     * Reset all stages to idle
     */
    resetAllStages() {
        this.stages.forEach(stage => {
            stage.status = 'idle';
        });
        this.render();
    }
}

// Add CSS for pipeline stages
const style = document.createElement('style');
style.textContent = `
    .pipeline-stages {
        padding: 20px;
        background: #f8f9fa;
        border-radius: 8px;
        margin: 20px 0;
    }
    
    .pipeline-stage {
        flex: 1;
        position: relative;
    }
    
    .stage-icon {
        width: 60px;
        height: 60px;
        border-radius: 50%;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        background: #e9ecef;
        color: #6c757d;
        transition: all 0.3s ease;
    }
    
    .stage-icon.idle {
        background: #e9ecef;
        color: #6c757d;
    }
    
    .stage-icon.active {
        background: #007bff;
        color: white;
        box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.25);
    }
    
    .stage-icon.completed {
        background: #28a745;
        color: white;
    }
    
    .stage-icon.error {
        background: #dc3545;
        color: white;
    }
    
    .stage-icon.pulse {
        animation: pulse 1.5s infinite;
    }
    
    @keyframes pulse {
        0% {
            box-shadow: 0 0 0 0 rgba(0, 123, 255, 0.7);
        }
        70% {
            box-shadow: 0 0 0 10px rgba(0, 123, 255, 0);
        }
        100% {
            box-shadow: 0 0 0 0 rgba(0, 123, 255, 0);
        }
    }
    
    .pipeline-connector {
        flex: 0 0 50px;
        height: 2px;
        background: #dee2e6;
        margin: 0 10px;
        position: relative;
        top: -20px;
    }
`;
document.head.appendChild(style);

// Export for use
window.PipelineStages = PipelineStages;