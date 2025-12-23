# Design Tokens Guide

## Overview

This document outlines the centralized design token system for ISX Pulse panels and floating components. It provides guidelines for CSS utilities vs custom hooks and establishes a single source of truth for design decisions.

## Centralized Design Tokens

### Panel System Tokens

```css
:root {
  /* Panel Sizing Tokens */
  --panel-max-width: 480px;
  --panel-min-width: 320px;
  --panel-margin: 16px;
  --panel-padding: 24px;
  --panel-border-radius: 12px;

  /* Panel Spacing Tokens */
  --panel-spacing-sm: 8px;
  --panel-spacing-md: 16px;
  --panel-spacing-lg: 24px;
  --panel-spacing-xl: 32px;

  /* Panel Visual Tokens */
  --panel-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  --panel-backdrop: rgba(0, 0, 0, 0.1);
  --panel-z-index: 50;

  /* Responsive Breakpoints */
  --panel-breakpoint-sm: 640px;   /* Mobile */
  --panel-breakpoint-md: 1024px;  /* Tablet */
  --panel-breakpoint-lg: 1280px;  /* Desktop */
}
```

### Color and Focus Tokens

```css
:root {
  /* Focus System */
  --focus-width: 2px;
  --focus-offset: 2px;
  --focus-color: hsl(var(--primary));
  --focus-shadow: 0 0 0 3px hsl(var(--primary) / 0.2);

  /* High Contrast Mode */
  --focus-width-high-contrast: 3px;
  --focus-shadow-high-contrast: 0 0 0 4px hsl(var(--primary) / 0.3);

  /* Interactive States */
  --panel-button-hover-transform: scale(1.02);
  --panel-button-active-transform: scale(0.98);
  --panel-transition-duration: 0.2s;
  --panel-transition-easing: cubic-bezier(0.4, 0, 0.2, 1);
}
```

## CSS vs Hook Decision Matrix

### When to Use CSS Utilities

**Use CSS for:**
- **Static styling**: Colors, fonts, borders, shadows
- **Responsive behavior**: Media queries, breakpoints
- **Animation states**: Hover, focus, active transitions
- **Layout patterns**: Grid, flexbox, positioning basics
- **Theme consistency**: Design system variables

**Examples:**
```css
/* ✅ Static panel styling */
.panel-base {
  background: hsl(var(--card));
  border: 1px solid hsl(var(--border));
  border-radius: var(--panel-border-radius);
  box-shadow: var(--panel-shadow);
  padding: var(--panel-padding);
}

/* ✅ Responsive behavior */
@media (max-width: 640px) {
  .panel-responsive {
    width: 100vw;
    max-width: none;
    margin: var(--panel-margin);
  }
}

/* ✅ Interactive states */
.panel-button {
  transition: transform var(--panel-transition-duration) var(--panel-transition-easing);
}

.panel-button:hover {
  transform: var(--panel-button-hover-transform);
}
```

### When to Use Custom Hooks

**Use Hooks for:**
- **Dynamic calculations**: Viewport measurements, positioning
- **Browser APIs**: visualViewport, ResizeObserver, IntersectionObserver
- **State management**: Complex component state, side effects
- **Performance optimization**: Throttling, debouncing, memoization
- **Event handling**: Keyboard, mouse, touch events

**Examples:**
```typescript
// ✅ Dynamic viewport positioning
const position = useViewportPosition({
  panelWidth: actualWidth,
  placement: 'center',
  margin: var(--panel-margin)
})

// ✅ Browser API integration
const viewportSupport = useVisualViewportSupport()

// ✅ Performance optimization
const throttledResize = useThrottledCallback(calculatePosition, 16)
```

## Integration Patterns

### Pattern 1: CSS-First Styling

```css
/* Define responsive panel base class */
.panel-base {
  max-width: var(--panel-max-width);
  width: min(85vw, var(--panel-max-width));
  margin: var(--panel-margin);
  padding: var(--panel-padding);
  border-radius: var(--panel-border-radius);
  box-shadow: var(--panel-shadow);
}

/* Responsive adjustments */
@media (max-width: 640px) {
  .panel-base {
    width: calc(100vw - calc(2 * var(--panel-margin)));
    max-width: none;
  }
}
```

### Pattern 2: Hook-Enhanced Positioning

```typescript
// Hook handles dynamic positioning only
const { positionStyles, supportLevel } = useViewportPosition({
  panelWidth: 600, // Will be overridden by actual measurements
  placement: 'center',
  margin: 16,      // Could use CSS variable
  enableErrorRecovery: true
})

// CSS handles static styling
<div
  className="panel-base"
  style={positionStyles}
>
  {/* Content */}
</div>
```

### Pattern 3: Hybrid Approach

```css
/* CSS provides responsive base */
.panel-responsive {
  width: clamp(320px, 80vw, 480px);
  max-height: clamp(300px, 70vh, 700px);
}

/* CSS provides animation classes */
.panel-enter {
  animation: panelSlideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.panel-exit {
  animation: panelSlideOut 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
```

```typescript
// Hook provides dynamic positioning and error recovery
const { positionStyles, lastError } = useViewportPosition({
  enableErrorRecovery: true,
  debug: process.env.NODE_ENV === 'development'
})

// Combine CSS and hook for optimal results
<motion.div
  className="panel-base panel-responsive panel-enter"
  style={positionStyles}
>
  {/* Content */}
</motion.div>
```

## Component Architecture

### Base Panel Component

```typescript
interface BasePanelProps {
  children: React.ReactNode
  className?: string
  style?: React.CSSProperties
  position?: 'center' | 'left' | 'right'
  responsive?: boolean
}

// Base panel with CSS-first approach
export function BasePanel({
  children,
  className,
  style,
  position = 'center',
  responsive = true
}: BasePanelProps) {
  return (
    <div
      className={cn(
        'panel-base',
        responsive && 'panel-responsive',
        `panel-position-${position}`,
        className
      )}
      style={style}
    >
      {children}
    </div>
  )
}
```

### Specialized Panel Variants

```typescript
// Floating panel with positioning hook
export function FloatingPanel({
  anchorRef,
  placement = 'center',
  children
}: FloatingPanelProps) {
  const { positionStyles, supportLevel } = useViewportPosition({
    placement,
    enableErrorRecovery: true
  })

  return (
    <FocusScope trapped>
      <BasePanel
        position={placement}
        style={positionStyles}
        className="panel-floating"
      >
        {children}
      </BasePanel>
    </FocusScope>
  )
}
```

## Responsive Design Guidelines

### Mobile-First Breakpoints

```css
/* Base styles (mobile first) */
.panel-responsive {
  width: calc(100vw - calc(2 * var(--panel-margin)));
  max-width: none;
  margin: var(--panel-margin);
  padding: var(--panel-spacing-md);
}

/* Tablet styles */
@media (min-width: 640px) {
  .panel-responsive {
    width: min(90vw, var(--panel-max-width));
    max-width: var(--panel-max-width);
    padding: var(--panel-padding);
  }
}

/* Desktop styles */
@media (min-width: 1024px) {
  .panel-responsive {
    width: var(--panel-max-width);
  }
}
```

### Touch vs Mouse Interactions

```css
/* Touch-friendly targets */
@media (pointer: coarse) {
  .panel-button {
    min-height: 44px;
    min-width: 44px;
    padding: 12px 16px;
  }
}

/* Mouse-optimized interactions */
@media (pointer: fine) {
  .panel-button {
    min-height: 32px;
    padding: 8px 12px;
  }
}
```

## Performance Guidelines

### CSS Optimization

1. **Use CSS Variables**: Faster runtime updates
2. **Minimize Reflows**: Batch style changes
3. **Optimize Animations**: Use transform/opacity
4. **Lazy Load Components**: Load panels only when needed

### Hook Optimization

1. **Memoize Calculations**: Use useMemo for expensive operations
2. **Throttle Events**: Prevent excessive re-renders
3. **Cleanup Resources**: Remove event listeners and observers
4. **Error Boundaries**: Graceful fallbacks for positioning failures

## Testing Guidelines

### Visual Regression Testing

```typescript
// Test matrix for responsive panels
const responsiveTestCases = [
  { width: 375, height: 667, name: 'iPhone SE' },
  { width: 768, height: 1024, name: 'iPad' },
  { width: 1920, height: 1080, name: 'Desktop' },
]
```

### Accessibility Testing

```typescript
// Keyboard navigation test cases
const keyboardTestCases = [
  { key: 'Tab', expected: 'Focus moves to next element' },
  { key: 'Shift+Tab', expected: 'Focus moves to previous element' },
  { key: 'Escape', expected: 'Panel closes' },
]
```

## Migration Checklist

### Phase 1: Token Extraction
- [ ] Extract all panel sizing to CSS variables
- [ ] Extract all spacing to CSS variables
- [ ] Extract all colors to CSS variables
- [ ] Update component classes to use tokens

### Phase 2: Hook Optimization
- [ ] Identify dynamic calculations that need hooks
- [ ] Extract positioning logic to unified hook
- [ ] Add error recovery to all positioning hooks
- [ ] Add performance optimizations (throttling, memoization)

### Phase 3: Component Unification
- [ ] Create BasePanel component
- [ ] Migrate existing panels to use BasePanel
- [ ] Add responsive variants
- [ ] Update documentation

### Phase 4: Testing & Validation
- [ ] Create visual regression tests
- [ ] Test responsive behavior across breakpoints
- [ ] Validate accessibility compliance
- [ ] Performance benchmarking

## Success Metrics

- **Bundle Size**: <5KB increase from token centralization
- **Performance**: <16ms positioning calculation time
- **Accessibility**: 100% WCAG AA compliance
- **Consistency**: 100% token usage across components
- **Maintainability**: <1 hour for design system updates

---

**Status**: 📋 In Progress
**Target Completion**: Phase 3.1 (Week 5-6)
**Bundle Impact**: Neutral (redistribution, not addition)
**Test Coverage**: 95%+ target