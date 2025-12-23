# App Directory

## Purpose
Next.js application pages for the ISX Daily Reports Scrapper frontend, implementing the main user interface with TypeScript and Shadcn/ui components.

## Structure

### Pages

#### page.tsx (Landing Page)
Main application landing page:
- **Welcome message**: Professional greeting for ISX system
- **Feature cards**: Overview of system capabilities
- **Navigation**: Quick access to all major sections
- **License status**: Current license information display

#### operations/page.tsx
Data processing operations interface:
- **Operation configuration**: Set parameters for data processing
- **Real-time progress**: WebSocket-powered live updates
- **Operation history**: View past operations and results
- **Export controls**: Download processed data

#### dashboard/page.tsx
Main application dashboard:
- **Data overview**: Summary of available data
- **Quick actions**: Common operations access
- **System status**: Health and performance metrics
- **Recent activity**: Latest operations and updates

#### analysis/page.tsx
Data analysis interface (placeholder):
- **Future implementation**: Advanced analysis tools
- **Chart visualizations**: Data trends and patterns
- **Export options**: Various report formats

#### reports/page.tsx
Report generation interface (placeholder):
- **Future implementation**: Custom report builder
- **Template selection**: Pre-defined report formats
- **Scheduling**: Automated report generation

#### license/page.tsx
License management interface:
- **License activation**: Enter and validate license keys
- **Status display**: Current license information
- **Renewal options**: Extend license validity
- **Hardware info**: System fingerprint details

### Layouts

#### layout.tsx
Root application layout:
- **Global styles**: Application-wide CSS
- **Theme provider**: Dark/light mode support
- **Font configuration**: Geist font family
- **Metadata**: SEO and application info

#### (unprotected)/layout.tsx
Layout for pages accessible without license:
- **Public pages**: License activation flow
- **Minimal navigation**: Essential links only
- **Branding**: ISX system identity

### Styles

#### globals.css
Global application styles:
- **Tailwind CSS**: Utility-first styling
- **CSS variables**: Theme customization
- **Base styles**: Typography and spacing
- **Component overrides**: Custom UI adjustments

## Component Integration

### Operations Page Components
```tsx
// operations/page.tsx uses:
import { OperationConfiguration } from '@/components/operations/OperationConfiguration'
import { OperationProgress } from '@/components/operations/OperationProgress'
import { OperationHistory } from '@/components/operations/OperationHistory'
```

### UI Components
All pages use Shadcn/ui components:
- **Card**: Content containers
- **Button**: Interactive elements
- **Tabs**: Section organization
- **Progress**: Status indicators
- **Alert**: User notifications

## Routing

### Protected Routes
Routes requiring valid license:
- `/dashboard` - Main application interface
- `/operations` - Data processing
- `/analysis` - Data analysis (future)
- `/reports` - Report generation (future)

### Public Routes
Accessible without license:
- `/` - Landing page
- `/license` - License activation

## State Management

### Client-Side State
- **Operation status**: React state for UI updates
- **WebSocket connection**: Real-time data flow
- **Form data**: Controlled components
- **UI preferences**: Local storage

### Server-Side State
- **License validation**: Middleware protection
- **Data fetching**: Server components
- **API integration**: Type-safe clients

## WebSocket Integration

Operations page WebSocket usage:
```typescript
// Real-time updates
const ws = new WebSocket('ws://localhost:8080/ws/operations')
ws.onmessage = (event) => {
  const update = JSON.parse(event.data)
  updateOperationProgress(update)
}
```

## TypeScript Patterns

### Type Safety
```typescript
interface OperationConfig {
  type: 'scraping' | 'processing' | 'export'
  mode: 'accumulative' | 'daily'
  dateRange?: {
    from: string
    to: string
  }
}
```

### API Integration
```typescript
// Type-safe API calls
const response = await apiClient.operations.start(config)
const status = await apiClient.operations.getStatus(operationId)
```

## Testing

### Component Tests
- React Testing Library for UI testing
- Mock WebSocket connections
- Form interaction testing
- Accessibility compliance

### E2E Tests
- Playwright for user flows
- License activation flow
- Operation execution flow
- Report generation flow

## Performance

### Optimization Strategies
- **Static generation**: Pre-render where possible
- **Code splitting**: Dynamic imports
- **Image optimization**: Next.js Image component
- **Bundle size**: <250KB initial load

### Monitoring
- **Web Vitals**: Core performance metrics
- **Error tracking**: Client-side error capture
- **Usage analytics**: Anonymous statistics

## Accessibility

### WCAG 2.1 AA Compliance
- **Keyboard navigation**: Full keyboard support
- **Screen readers**: Proper ARIA labels
- **Color contrast**: Meeting standards
- **Focus management**: Clear focus indicators

## Change Log
- 2025-07-30: Created operations page with real-time WebSocket updates (Phase 3)
- 2025-07-30: Updated landing page with professional ISX branding
- 2025-07-30: Added placeholder pages for analysis and reports sections
- 2025-07-30: Created README.md documentation for app directory

## Best Practices
1. **Type safety**: Use TypeScript interfaces for all data
2. **Component reuse**: Extract common UI patterns
3. **Error boundaries**: Graceful error handling
4. **Loading states**: Clear feedback for async operations
5. **Responsive design**: Mobile-first approach