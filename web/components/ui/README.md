# UI Components - No Data States

This document explains how to use the new reusable UI components for handling no-data and loading states in the ISX Pulse application.

## Components

### NoDataState

A reusable component for displaying consistent "no data available" states across the application.

#### Props

```typescript
interface NoDataStateProps {
  icon?: LucideIcon              // Optional icon to display
  iconColor?: 'blue' | 'green' | 'purple' | 'orange' | 'red' | 'gray'
  title: string                  // Main heading text
  description: string            // Description text
  instructions?: string[]        // Optional step-by-step instructions
  actions?: NoDataAction[]       // Optional action buttons
  className?: string             // Additional CSS classes
}

interface NoDataAction {
  label: string                  // Button text
  variant?: ButtonProps['variant'] // Button style variant
  href?: string                  // Link URL (creates Link button)
  onClick?: () => void           // Click handler (creates button)
  icon?: LucideIcon             // Optional button icon
}
```

#### Usage Examples

##### Basic No Data State
```typescript
import { NoDataState } from '@/components/ui/no-data-state'
import { ChartCandlestick } from 'lucide-react'

<NoDataState
  icon={ChartCandlestick}
  iconColor="blue"
  title="No Analysis Data Available"
  description="You need to run data collection operations first to generate analysis insights."
/>
```

##### With Instructions and Actions
```typescript
import { NoDataState } from '@/components/ui/no-data-state'
import { ChartCandlestick, Activity, RefreshCw } from 'lucide-react'

<NoDataState
  icon={ChartCandlestick}
  iconColor="blue"
  title="No Analysis Data Available"
  description="You need to run data collection operations first to generate analysis insights."
  instructions={[
    "Go to the Operations page",
    "Run 'Full Pipeline' to collect all data",
    "Wait for analysis to complete",
    "Return here to view the insights"
  ]}
  actions={[
    { 
      label: "Go to Operations", 
      href: "/operations", 
      variant: "default",
      icon: Activity
    },
    { 
      label: "Check Again", 
      onClick: handleRetry, 
      variant: "outline",
      icon: RefreshCw
    }
  ]}
/>
```

### DataLoadingState

A consistent loading state component for data fetching operations.

#### Props

```typescript
interface DataLoadingStateProps {
  message?: string               // Loading message (default: "Loading...")
  icon?: LucideIcon             // Custom icon (default: Loader2)
  className?: string            // Additional CSS classes
  showCard?: boolean            // Wrap in card (default: true)
  size?: 'sm' | 'default' | 'lg' // Size variant
}
```

#### Usage Examples

##### Basic Loading State
```typescript
import { DataLoadingState } from '@/components/ui/data-loading-state'

<DataLoadingState />
```

##### Custom Message and Size
```typescript
import { DataLoadingState } from '@/components/ui/data-loading-state'

<DataLoadingState 
  message="Analyzing market data..."
  size="lg"
/>
```

##### Without Card Wrapper
```typescript
import { DataLoadingState } from '@/components/ui/data-loading-state'

<DataLoadingState 
  message="Loading chart..."
  showCard={false}
  size="sm"
/>
```

##### Custom Icon
```typescript
import { DataLoadingState } from '@/components/ui/data-loading-state'
import { BarChart3 } from 'lucide-react'

<DataLoadingState 
  message="Generating charts..."
  icon={BarChart3}
/>
```

## Integration Patterns

### With useHydration Hook

Following CLAUDE.md hydration best practices:

```typescript
'use client'

import { useHydration } from '@/lib/hooks/use-hydration'
import { NoDataState, DataLoadingState } from '@/components/ui'
import { ChartCandlestick } from 'lucide-react'

function AnalysisPage() {
  const isHydrated = useHydration()
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  // Pre-hydration loading state
  if (!isHydrated) {
    return (
      <DataLoadingState 
        message="Initializing application..."
        size="lg"
      />
    )
  }

  // Data loading state
  if (loading) {
    return (
      <DataLoadingState 
        message="Loading analysis data..."
      />
    )
  }

  // No data state
  if (!data && !error) {
    return (
      <NoDataState
        icon={ChartCandlestick}
        iconColor="blue"
        title="No Analysis Data Available"
        description="Run data collection operations to generate insights."
        instructions={[
          "Go to the Operations page",
          "Run 'Full Pipeline' to collect data",
          "Return here to view analysis"
        ]}
        actions={[
          { label: "Go to Operations", href: "/operations" },
          { label: "Retry", onClick: loadData, variant: "outline" }
        ]}
      />
    )
  }

  // Render normal content
  return <div>{/* Normal page content */}</div>
}
```

### Error Handling with No Data States

```typescript
'use client'

import { NoDataState } from '@/components/ui/no-data-state'
import { AlertTriangle, RefreshCw } from 'lucide-react'

function ReportsPage() {
  const [error, setError] = useState(null)

  if (error) {
    // Check if error indicates no data vs actual error
    const isNoDataError = error.includes('404') || error.includes('not found')
    
    if (isNoDataError) {
      return (
        <NoDataState
          icon={AlertTriangle}
          iconColor="orange"
          title="No Reports Available"
          description="No reports have been generated yet. Generate some reports first."
          actions={[
            { label: "Generate Reports", href: "/operations" },
            { label: "Try Again", onClick: reload, variant: "outline", icon: RefreshCw }
          ]}
        />
      )
    }
    
    // Handle actual errors differently
    return <ErrorBoundary error={error} />
  }

  // Normal component rendering...
}
```

## Styling and Customization

### Icon Colors

Available icon color variants:
- `blue` (default): Blue background with blue icon
- `green`: Green background with green icon  
- `purple`: Purple background with purple icon
- `orange`: Orange background with orange icon
- `red`: Red background with red icon
- `gray`: Gray background with gray icon

### Size Variants (DataLoadingState)

- `sm`: Small loading state (6x6 icon, compact spacing)
- `default`: Standard loading state (8x8 icon, normal spacing)
- `lg`: Large loading state (12x12 icon, expanded spacing)

### Custom Styling

Both components accept `className` prop for additional styling:

```typescript
<NoDataState
  title="Custom Styled"
  description="With additional classes"
  className="my-custom-class"
/>

<DataLoadingState
  className="custom-loading-style"
  showCard={false}
/>
```

## Best Practices

1. **Consistent Usage**: Use these components across all pages for consistency
2. **Meaningful Messages**: Provide clear, actionable descriptions
3. **Appropriate Icons**: Choose icons that match the content type
4. **Action Buttons**: Always provide clear next steps for users
5. **Hydration Safety**: Use with `useHydration` hook for client-side components
6. **Error Distinction**: Distinguish between no-data and error states
7. **Loading Context**: Use specific loading messages for different operations

## Migration from Existing Components

To replace existing no-data implementations:

1. Import the new components
2. Replace custom loading states with `DataLoadingState`
3. Replace custom no-data states with `NoDataState`
4. Update props to match the new interface
5. Test for hydration compatibility

Example migration:

```typescript
// Before
<div className="text-center p-8">
  <h2>No data available</h2>
  <p>Run operations first</p>
  <Button onClick={reload}>Try Again</Button>
</div>

// After
<NoDataState
  title="No data available"
  description="Run operations first"
  actions={[{ label: "Try Again", onClick: reload }]}
/>
```