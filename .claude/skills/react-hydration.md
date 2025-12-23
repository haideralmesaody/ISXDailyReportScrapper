---
name: react-hydration
type: domain-skill
expertise_level: expert
domain: nextjs-ssr-hydration
version: "1.0.0"
---

# React Hydration Skill - Next.js SSR/Hydration Patterns

## Purpose
Expert knowledge in React Server-Side Rendering (SSR) and Client-Side Hydration patterns, specifically focused on preventing and resolving hydration mismatches in Next.js 14 applications.

## Hydration Fundamentals

### Common Hydration Issues

#### 1. Date/Time Operations
**Problem**: Server and client render different timestamps
```typescript
// ❌ CAUSES HYDRATION MISMATCH
function TimestampComponent() {
    const now = new Date()
    return <div>{now.toISOString()}</div>
}

// ✅ HYDRATION SAFE
function TimestampComponent() {
    const [isHydrated, setIsHydrated] = useState(false)

    useEffect(() => {
        setIsHydrated(true)
    }, [])

    if (!isHydrated) {
        return <div>Loading...</div>
    }

    return <div>{new Date().toISOString()}</div>
}
```

#### 2. Browser API Usage
**Problem**: Client-only APIs not available on server
```typescript
// ✅ HYDRATION SAFE
function WindowComponent() {
    const [width, setWidth] = useState(0)
    const [isHydrated, setIsHydrated] = useState(false)

    useEffect(() => {
        setIsHydrated(true)
        setWidth(window.innerWidth)
    }, [])

    if (!isHydrated) {
        return <div>Detecting screen size...</div>
    }

    return <div>Window width: {width}</div>
}
```

## ISX Pulse Hydration Patterns

### 1. Use Hydration Hook
```typescript
// lib/hooks/use-hydration.ts
import { useState, useEffect } from 'react'

export function useHydration(): boolean {
    const [isHydrated, setIsHydrated] = useState(false)

    useEffect(() => {
        setIsHydrated(true)
    }, [])

    return isHydrated
}
```

### 2. Chart Hydration Pattern
```typescript
// components/charts/HydrationSafeChart.tsx
function HydrationSafeChart({ data }: { data: ChartData }) {
    const chartRef = useRef<HTMLDivElement>(null)
    const isHydrated = useHydration()

    useEffect(() => {
        if (isHydrated && chartRef.current) {
            // Initialize TradingView chart only after hydration
            const chart = createChart(chartRef.current, chartOptions)
        }
    }, [isHydrated, data])

    return (
        <div ref={chartRef} className="chart-container">
            {!isHydrated && <ChartSkeleton />}
        </div>
    )
}
```

## WebSocket Hydration Patterns

### Connection Management
```typescript
// hooks/use-websocket.ts
function useWebSocket(url: string) {
    const [socket, setSocket] = useState<WebSocket | null>(null)
    const [connected, setConnected] = useState(false)
    const isHydrated = useHydration()

    useEffect(() => {
        if (isHydrated && !socket) {
            const ws = new WebSocket(url)
            ws.onopen = () => {
                setConnected(true)
                setSocket(ws)
            }
            return () => ws.close()
        }
    }, [isHydrated])

    return { connected }
}
```

## Best Practices

### DO ✅
1. **Use hydration guard**: Always check `isHydrated` before client-only operations
2. **Provide loading states**: Show consistent UI during hydration
3. **Defer client-side data**: Fetch data after hydration when possible
4. **Use dynamic imports**: Load client-only components with `ssr: false`

### DON'T ❌
1. **Don't use Date.now()**: Server and client will render different timestamps
2. **Don't access window/document**: Use hydration guards for browser APIs
3. **Don't render random data**: Use consistent seeds or defer to client-side
4. **Don't ignore warnings**: Address hydration mismatches immediately

This skill provides comprehensive React hydration expertise for the ISX Pulse application.