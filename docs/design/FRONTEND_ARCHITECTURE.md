# Frontend Architecture Guide

## Overview

This document outlines the frontend architecture for the ISX Daily Reports Scrapper web application. The architecture follows a modular, component-based approach while strictly adhering to the three-layer architecture principles defined in ARCHITECTURE_PRINCIPLES.md.

## Core Principles

### 1. Display-Only Frontend
The frontend is responsible ONLY for:
- Rendering data received from the backend
- Collecting user input
- Showing status updates via WebSocket
- Managing UI state (visibility, preferences)

The frontend MUST NOT:
- Control pipeline flow
- Make business logic decisions
- Calculate data transformations
- Determine stage transitions

### 2. WebSocket Communication
- **One-way flow**: Backend → Frontend
- Used for status updates and progress only
- Never used for control or commands
- All pipeline control via process exit codes

### 3. Modular Structure
- Small, focused modules
- Clear separation of concerns
- Reusable components
- Event-driven communication

## Directory Structure

```
web/
├── index.html                    # Minimal HTML shell
├── static/
│   ├── css/
│   │   ├── main.css             # Global styles, layout
│   │   ├── components.css       # Component-specific styles
│   │   └── theme.css            # ISX theme variables
│   ├── js/
│   │   ├── core/
│   │   │   ├── websocket.js     # WebSocket connection management
│   │   │   ├── eventBus.js      # Component communication
│   │   │   └── state.js         # UI state management
│   │   ├── components/
│   │   │   ├── pipeline.js      # Pipeline visualization
│   │   │   ├── fileManager.js   # File lists display
│   │   │   ├── charts.js        # Chart management
│   │   │   ├── marketMovers.js  # Market data display
│   │   │   └── forms.js         # Form handling
│   │   ├── services/
│   │   │   └── api.js           # API calls (data fetching only)
│   │   └── main.js              # Application entry point
│   └── vendor/                   # Third-party libraries
│       ├── alpine.js
│       ├── bootstrap/
│       ├── highcharts/
│       └── chart.js/
```

## Component Architecture

### Core Modules

#### 1. WebSocket Manager (`websocket.js`)
```javascript
export class WebSocketManager {
    constructor(url) {
        this.url = url;
        this.handlers = new Map();
        this.reconnectAttempts = 0;
    }
    
    connect() {
        this.ws = new WebSocket(this.url);
        this.setupHandlers();
    }
    
    onMessage(type, handler) {
        this.handlers.set(type, handler);
    }
    
    // Auto-reconnect, error handling, etc.
}
```

#### 2. Event Bus (`eventBus.js`)
```javascript
export class EventBus {
    constructor() {
        this.events = new Map();
    }
    
    on(event, handler) {
        if (!this.events.has(event)) {
            this.events.set(event, []);
        }
        this.events.get(event).push(handler);
    }
    
    emit(event, data) {
        if (this.events.has(event)) {
            this.events.get(event).forEach(handler => handler(data));
        }
    }
}
```

#### 3. UI State Manager (`state.js`)
```javascript
export class UIState {
    constructor() {
        this.state = {
            sections: {
                pipeline: { visible: true },
                files: { visible: true },
                charts: { visible: true }
            },
            preferences: {
                theme: 'light',
                autoRefresh: true
            }
        };
        this.observers = [];
        this.loadFromStorage();
    }
    
    setState(path, value) {
        // Update state and notify observers
        // Persist to localStorage
    }
    
    observe(callback) {
        this.observers.push(callback);
    }
}
```

### Component Pattern

Each component follows this pattern:

```javascript
export class PipelineVisualizer {
    constructor(container, { eventBus, websocket }) {
        this.container = container;
        this.eventBus = eventBus;
        this.websocket = websocket;
        this.stages = ['scraping', 'processing', 'indices', 'analysis'];
        
        this.init();
    }
    
    init() {
        // Set up WebSocket handlers
        this.websocket.onMessage('pipeline:status', (data) => {
            this.updateStage(data);
        });
        
        // Listen to events
        this.eventBus.on('pipeline:reset', () => {
            this.reset();
        });
        
        // Initial render
        this.render();
    }
    
    render() {
        // Render component HTML
    }
    
    updateStage(data) {
        // Update visualization
        // Emit events if needed
        this.eventBus.emit('stage:updated', data);
    }
}
```

## Alpine.js Integration

For reactive UI updates without heavy frameworks:

```javascript
// In main.js
import Alpine from 'alpinejs';

// Register global Alpine data
Alpine.data('app', () => ({
    connected: false,
    licenseStatus: 'checking',
    pipeline: {
        current: null,
        stages: {}
    },
    
    init() {
        // Initialize app state
        eventBus.on('websocket:connected', () => {
            this.connected = true;
        });
        
        eventBus.on('pipeline:update', (data) => {
            this.pipeline = data;
        });
    }
}));

Alpine.start();
```

## Data Flow

### 1. User Action → Backend
```
User clicks "Start Scraping"
  → Form submission (POST /api/scrape)
  → Backend starts process
  → Backend sends WebSocket updates
  → Frontend displays progress
```

### 2. WebSocket Update → UI
```
Backend sends progress message
  → WebSocket manager receives
  → Calls registered handlers
  → Components update display
  → Events emitted for other components
```

### 3. Data Fetch → Display
```
Component needs data
  → API service fetches (/api/data)
  → Component renders data
  → No transformation or calculation
```

## Build System (Vite)

### Development Configuration
```javascript
// vite.config.js
export default {
    root: 'web',
    server: {
        port: 3000,
        proxy: {
            '/api': 'http://localhost:8080',
            '/ws': {
                target: 'ws://localhost:8080',
                ws: true
            }
        }
    }
}
```

### Production Build
```javascript
build: {
    outDir: '../release/web',
    rollupOptions: {
        output: {
            manualChunks: {
                vendor: ['alpine', 'highcharts'],
                charts: ['./src/components/charts.js']
            }
        }
    }
}
```

## Migration Strategy

### Phase 1: Extract Modules (Current)
1. Move CSS to separate files
2. Extract JavaScript to modules
3. Maintain existing functionality

### Phase 2: Add Alpine.js
1. Replace manual DOM updates
2. Add reactive data binding
3. Simplify event handling

### Phase 3: Component Refactoring
1. Create reusable components
2. Implement event bus
3. Add state management

### Phase 4: Build System
1. Set up Vite
2. Configure development environment
3. Optimize for production

## Best Practices

### 1. Component Guidelines
- Single responsibility principle
- Clear input/output interface
- Event-driven communication
- No direct DOM manipulation outside component

### 2. State Management
- UI state only (no business data)
- Persist user preferences
- Use observers for updates
- Keep state minimal

### 3. Error Handling
- Display user-friendly messages
- Log errors for debugging
- Provide recovery options
- Never expose internal details

### 4. Performance
- Lazy load components
- Debounce rapid updates
- Use virtual scrolling for long lists
- Minimize re-renders

## Testing Approach

### Unit Tests
- Test components in isolation
- Mock dependencies
- Test event handling
- Verify state updates

### Integration Tests
- Test component interactions
- Verify WebSocket handling
- Test API integration
- Validate user flows

## Security Considerations

1. **Input Validation**: Sanitize all user inputs
2. **XSS Prevention**: Use textContent, not innerHTML
3. **CORS**: Properly configured for API calls
4. **WebSocket**: Validate origin and messages
5. **Dependencies**: Regular security audits

## Future Enhancements

1. **TypeScript**: Add type safety
2. **Web Components**: Native component system
3. **Service Worker**: Offline functionality
4. **PWA**: Mobile app experience
5. **Accessibility**: WCAG compliance

## Component Naming Standards

All components must follow the standardized naming conventions defined in [COMPONENT_NAMING_STANDARDS.md](./COMPONENT_NAMING_STANDARDS.md). This ensures consistency across:
- Component IDs
- Template filenames
- Container element IDs
- CSS class names
- JavaScript references

## References

- [ARCHITECTURE_PRINCIPLES.md](./ARCHITECTURE_PRINCIPLES.md) - Core architecture rules
- [COMPONENT_NAMING_STANDARDS.md](./COMPONENT_NAMING_STANDARDS.md) - Component naming conventions
- [DEVELOPMENT_TASKS.md](../developer/DEVELOPMENT_TASKS.md) - Implementation tasks
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Vite Documentation](https://vitejs.dev/)