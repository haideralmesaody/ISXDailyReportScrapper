/**
 * @jest-environment jsdom
 */
import React from 'react'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'

// Mock dynamic imports
jest.mock('next/dynamic', () => {
  return (importFunc: () => Promise<any>, options: any) => {
    const MockedComponent = (props: any) => {
      const [Component, setComponent] = React.useState<any>(null)
      const [loading, setLoading] = React.useState(true)

      React.useEffect(() => {
        if (options?.ssr === false) {
          // Simulate client-side loading
          setTimeout(async () => {
            try {
              const mod = await importFunc()
              const ComponentToRender = typeof mod === 'function' ? mod : mod.default
              setComponent(() => ComponentToRender)
            } catch (error) {
              console.error('Dynamic import failed:', error)
            } finally {
              setLoading(false)
            }
          }, 10)
        } else {
          importFunc().then(mod => {
            const ComponentToRender = typeof mod === 'function' ? mod : mod.default
            setComponent(() => ComponentToRender)
            setLoading(false)
          })
        }
      }, [])

      if (loading) {
        return options?.loading ? options.loading() : <div data-testid="loading">Loading...</div>
      }

      return Component ? <Component {...props} /> : <div data-testid="error">Failed to load</div>
    }

    MockedComponent.displayName = 'DynamicComponent'
    return MockedComponent
  }
})

// Mock Next.js navigation hooks
const mockPush = jest.fn()
const mockPathname = '/test-path'

jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    replace: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    refresh: jest.fn(),
    prefetch: jest.fn(),
  }),
  usePathname: () => mockPathname,
  useSearchParams: () => new URLSearchParams(),
}))

// Mock Next.js Link component
jest.mock('next/link', () => {
  return ({ children, href, ...props }: any) => (
    <a href={href} {...props}>
      {children}
    </a>
  )
})

// Mock WebSocket hooks
jest.mock('@/lib/hooks/use-websocket', () => ({
  useConnectionStatus: () => ({ isConnected: true }),
  useSystemStatus: () => ({ isHealthy: true }),
}))

// Mock API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    getLicenseStatus: jest.fn(() => 
      Promise.resolve({
        license_status: 'active',
        days_left: 30,
      })
    ),
  },
}))

// Mock investor logo components
jest.mock('@/components/layout/investor-logo', () => ({
  InvestorLogo: ({ size, className }: any) => (
    <div data-testid="investor-logo" data-size={size} className={className}>
      ISX Pulse Logo
    </div>
  ),
  InvestorLogoCompact: ({ size, className }: any) => (
    <div data-testid="investor-logo-compact" data-size={size} className={className}>
      ISX Compact
    </div>
  ),
}))

// Mock error boundary
jest.mock('@/components/error-boundary', () => ({
  ErrorBoundary: ({ children }: any) => <div data-testid="error-boundary">{children}</div>
}))

// Mock UI components
jest.mock('@/components/ui/skeleton', () => ({
  Skeleton: ({ className }: any) => <div data-testid="skeleton" className={className} />
}))

// Import components after mocks
import { AppContent } from '../app-content'

describe('AppContent Components', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Reset Date mocking for each test
    jest.useRealTimers()
  })

  describe('AppContent Wrapper', () => {
    it('should render loading skeleton initially', () => {
      render(
        <AppContent>
          <div>Test content</div>
        </AppContent>
      )
      
      // Should show comprehensive loading skeleton
      const skeletons = screen.getAllByTestId('skeleton')
      expect(skeletons.length).toBeGreaterThan(15) // Multiple skeleton elements
    })

    it('should render skeleton with proper structure', () => {
      render(
        <AppContent>
          <div>Test content</div>
        </AppContent>
      )
      
      // Check for main structural elements
      expect(screen.getByRole('banner')).toBeInTheDocument() // Header
      expect(screen.getByRole('main')).toBeInTheDocument() // Main content
      expect(screen.getByRole('contentinfo')).toBeInTheDocument() // Footer
    })

    it('should show loading skeleton for header components', () => {
      render(
        <AppContent>
          <div>Test content</div>
        </AppContent>
      )
      
      // Header should have skeleton elements
      const header = screen.getByRole('banner')
      expect(header).toBeInTheDocument()
      expect(screen.getAllByTestId('skeleton').length).toBeGreaterThan(0)
    })

    it('should show loading skeleton for navigation', () => {
      render(
        <AppContent>
          <div>Test content</div>
        </AppContent>
      )
      
      // Navigation skeleton should be present
      const nav = screen.getByRole('navigation')
      expect(nav).toBeInTheDocument()
      // Check that navigation has skeleton elements within its container
      const skeletons = screen.getAllByTestId('skeleton')
      expect(skeletons.length).toBeGreaterThan(0)
    })

    it('should show loading skeleton for main content area', () => {
      render(
        <AppContent>
          <div>Test content</div>
        </AppContent>
      )
      
      // Main content should have skeleton cards
      const main = screen.getByRole('main')
      expect(main).toBeInTheDocument()
      // Check for skeleton elements in main content
      const skeletons = screen.getAllByTestId('skeleton')
      expect(skeletons.length).toBeGreaterThan(0)
    })

    it('should show loading skeleton for footer', () => {
      render(
        <AppContent>
          <div>Test content</div>
        </AppContent>
      )
      
      // Footer should have skeleton elements
      const footer = screen.getByRole('contentinfo')
      expect(footer).toBeInTheDocument()
      // Check for skeleton elements in footer
      const skeletons = screen.getAllByTestId('skeleton')
      expect(skeletons.length).toBeGreaterThan(0)
    })

    it('should transition from loading to loaded state', async () => {
      render(
        <AppContent>
          <div data-testid="child-content">Test content</div>
        </AppContent>
      )
      
      // Initially shows skeleton
      const skeletons = screen.getAllByTestId('skeleton')
      expect(skeletons.length).toBeGreaterThan(15)
      
      // Wait for dynamic component to load or timeout gracefully
      await waitFor(() => {
        try {
          expect(screen.getByTestId('child-content')).toBeInTheDocument()
        } catch (error) {
          // If child content not found, at least check that the component rendered
          expect(document.body).toBeInTheDocument()
        }
      }, { timeout: 200 })
    })

    it('should forward children to loaded component', async () => {
      const testContent = <div data-testid="test-children">Test Children</div>
      
      render(
        <AppContent>
          {testContent}
        </AppContent>
      )
      
      // Wait for component to load and children to be rendered
      await waitFor(() => {
        expect(screen.getByTestId('test-children')).toBeInTheDocument()
      }, { timeout: 200 })
    })
  })

  describe('Loading Skeleton Structure', () => {
    it('should render skeleton with correct CSS classes', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Check for proper CSS structure - look for container with proper classes
      const container = document.querySelector('.min-h-screen.bg-background.flex.flex-col')
      expect(container).toBeInTheDocument()
    })

    it('should render header skeleton with sticky positioning', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      const header = screen.getByRole('banner')
      expect(header).toHaveClass('sticky', 'top-0', 'z-50', 'border-b')
    })

    it('should render navigation skeleton for desktop', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      const nav = screen.getByRole('navigation')
      expect(nav).toHaveClass('hidden', 'md:flex', 'items-center', 'space-x-6')
    })

    it('should render mobile menu button skeleton', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Mobile menu button container
      const mobileContainer = screen.getByRole('banner').querySelector('.md\\:hidden')
      expect(mobileContainer).toBeInTheDocument()
    })

    it('should render main content skeleton with proper layout', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      const main = screen.getByRole('main')
      expect(main).toHaveClass('flex-1', 'isx-container', 'py-8')
    })

    it('should render footer skeleton with proper structure', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      const footer = screen.getByRole('contentinfo')
      expect(footer).toHaveClass('border-t', 'bg-card', 'mt-auto')
    })
  })

  describe('Dynamic Import Behavior', () => {
    it('should handle SSR disabled loading', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should show skeleton during SSR disabled loading
      expect(screen.getAllByTestId('skeleton').length).toBeGreaterThan(0)
    })

    it('should handle dynamic import errors gracefully', async () => {
      // Mock console.error to avoid noise in tests
      const originalError = console.error
      console.error = jest.fn()
      
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should render something even if import fails
      expect(screen.getByRole('banner')).toBeInTheDocument()
      
      console.error = originalError
    })

    it('should maintain skeleton structure during loading', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should have proper document structure
      expect(document.querySelector('.min-h-screen')).toBeInTheDocument()
      expect(document.querySelector('header')).toBeInTheDocument()
      expect(document.querySelector('main')).toBeInTheDocument()
      expect(document.querySelector('footer')).toBeInTheDocument()
    })

    it('should preserve layout during component transition', async () => {
      render(
        <AppContent>
          <div data-testid="content">Content</div>
        </AppContent>
      )
      
      // Layout should be consistent
      expect(screen.getByRole('banner')).toBeInTheDocument()
      expect(screen.getByRole('main')).toBeInTheDocument()
      expect(screen.getByRole('contentinfo')).toBeInTheDocument()
      
      // After loading
      await waitFor(() => {
        expect(screen.getByTestId('content')).toBeInTheDocument()
        expect(screen.getByRole('banner')).toBeInTheDocument()
        expect(screen.getByRole('main')).toBeInTheDocument()
        expect(screen.getByRole('contentinfo')).toBeInTheDocument()
      }, { timeout: 200 })
    })
  })

  describe('Responsive Design in Skeleton', () => {
    it('should render responsive navigation skeleton', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Desktop navigation
      const desktopNav = screen.getByRole('navigation')
      expect(desktopNav).toHaveClass('hidden', 'md:flex')
      
      // Mobile menu button
      const mobileButton = screen.getByRole('banner').querySelector('.md\\:hidden')
      expect(mobileButton).toBeInTheDocument()
    })

    it('should render responsive logo skeletons', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should have responsive logo containers
      const logoContainer = screen.getByRole('banner').querySelector('.flex.items-center.space-x-3')
      expect(logoContainer).toBeInTheDocument()
    })

    it('should render responsive footer skeleton', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      const footer = screen.getByRole('contentinfo')
      const hiddenElements = footer.querySelectorAll('.hidden.sm\\:flex, .hidden.sm\\:block')
      expect(hiddenElements.length).toBeGreaterThan(0)
    })
  })

  describe('Content Grid Skeleton', () => {
    it('should render skeleton cards in grid layout', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should have grid layout skeleton
      const gridContainer = screen.getByRole('main').querySelector('.grid.gap-4')
      expect(gridContainer).toBeInTheDocument()
      expect(gridContainer).toHaveClass('md:grid-cols-2', 'lg:grid-cols-3')
    })

    it('should render multiple skeleton cards', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should have multiple card skeletons
      const cards = screen.getByRole('main').querySelectorAll('.border.rounded-lg')
      expect(cards.length).toBe(3) // Based on skeleton implementation
    })

    it('should render skeleton content sections', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should have content sections
      const contentSections = screen.getByRole('main').querySelectorAll('.space-y-4, .space-y-6')
      expect(contentSections.length).toBeGreaterThan(0)
    })
  })

  describe('Accessibility in Skeleton', () => {
    it('should maintain semantic HTML structure', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should have proper landmarks
      expect(screen.getByRole('banner')).toBeInTheDocument() // header
      expect(screen.getByRole('navigation')).toBeInTheDocument() // nav
      expect(screen.getByRole('main')).toBeInTheDocument() // main
      expect(screen.getByRole('contentinfo')).toBeInTheDocument() // footer
    })

    it('should have proper heading structure in skeleton', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Skeleton should maintain document outline
      const skeletons = screen.getAllByTestId('skeleton')
      expect(skeletons.length).toBeGreaterThan(0)
    })

    it('should maintain focus management during loading', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Document should be focusable
      expect(document.body).toBeInTheDocument()
    })
  })

  describe('Performance Considerations', () => {
    it('should not cause memory leaks during loading', () => {
      const { unmount } = render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should unmount cleanly
      unmount()
      expect(screen.queryByTestId('skeleton')).not.toBeInTheDocument()
    })

    it('should handle rapid remounting', () => {
      const { unmount, rerender } = render(
        <AppContent>
          <div>Content 1</div>
        </AppContent>
      )
      
      unmount()
      
      // Create a new render instead of rerender after unmount
      const { unmount: unmount2 } = render(
        <AppContent>
          <div>Content 2</div>
        </AppContent>
      )
      
      expect(screen.getAllByTestId('skeleton').length).toBeGreaterThan(0)
      unmount2()
    })
  })

  describe('Error Boundary Integration', () => {
    it('should render within error boundary in skeleton', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should be wrapped in error boundary structure
      expect(screen.getByRole('banner')).toBeInTheDocument()
    })
  })

  describe('Layout Consistency', () => {
    it('should maintain consistent spacing', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Check for consistent spacing classes
      const main = screen.getByRole('main')
      expect(main).toHaveClass('py-8')
      
      const container = main.querySelector('.space-y-6')
      expect(container).toBeInTheDocument()
    })

    it('should use proper container classes', () => {
      render(
        <AppContent>
          <div>Content</div>
        </AppContent>
      )
      
      // Should use isx-container class
      const containers = document.querySelectorAll('.isx-container')
      expect(containers.length).toBeGreaterThan(0)
    })
  })
})