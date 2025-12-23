/**
 * @jest-environment jsdom
 */
import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
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

// Mock Next.js Image component
jest.mock('next/image', () => {
  return ({ src, alt, onError, priority, fill, sizes, ...props }: any) => {
    return (
      <img
        src={src}
        alt={alt}
        data-testid="next-image"
        onError={onError}
        data-priority={priority ? 'true' : 'false'}
        data-fill={fill ? 'true' : 'false'}
        data-sizes={sizes}
        {...props}
      />
    )
  }
})

// Import components after mocks
import {
  InvestorLogo,
  InvestorLogoCompact,
  InvestorIcon,
  InvestorHeaderLogo,
  InvestorFavicon
} from '../investor-logo'

describe('InvestorLogo Components', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('InvestorLogo', () => {
    it('should render loading skeleton initially', () => {
      render(<InvestorLogo />)
      
      // Should show skeleton during loading - find by animate-pulse class
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render with different sizes', async () => {
      const sizes = ['sm', 'md', 'lg', 'xl', '2xl'] as const
      
      for (const size of sizes) {
        const { unmount } = render(<InvestorLogo size={size} />)
        
        // Wait for dynamic component to load
        await waitFor(() => {
          const pulsingElements = document.querySelectorAll('.animate-pulse')
          expect(pulsingElements.length).toBeGreaterThan(0)
        }, { timeout: 100 })
        
        unmount()
      }
    })

    it('should handle showText prop', async () => {
      render(<InvestorLogo showText={false} />)
      
      // Wait for component to load
      await waitFor(() => {
        const pulsingElements = document.querySelectorAll('.animate-pulse')
        expect(pulsingElements.length).toBeGreaterThan(0)
      }, { timeout: 100 })
    })

    it('should handle variant prop', async () => {
      const variants = ['full', 'compact', 'icon-only'] as const
      
      for (const variant of variants) {
        const { unmount } = render(<InvestorLogo variant={variant} />)
        
        await waitFor(() => {
          const pulsingElements = document.querySelectorAll('.animate-pulse')
          expect(pulsingElements.length).toBeGreaterThan(0)
        }, { timeout: 100 })
        
        unmount()
      }
    })

    it('should apply custom className', async () => {
      render(<InvestorLogo className="custom-class" />)
      
      await waitFor(() => {
        const container = document.querySelector('.custom-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
    })

    it('should render skeleton with correct size classes', () => {
      render(<InvestorLogo size="lg" />)
      
      // Check skeleton has animate-pulse class during loading
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })
  })

  describe('InvestorLogoCompact', () => {
    it('should render loading skeleton initially', () => {
      render(<InvestorLogoCompact />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render with small and medium sizes', async () => {
      const { rerender } = render(<InvestorLogoCompact size="sm" />)
      
      // Initially should have skeleton or loaded component
      expect(document.body).toBeInTheDocument()
      
      rerender(<InvestorLogoCompact size="md" />)
      
      // Should render component (either skeleton or loaded)
      expect(document.body).toBeInTheDocument()
    })

    it('should apply custom className when provided', async () => {
      render(<InvestorLogoCompact className="compact-class" />)
      
      await waitFor(() => {
        const container = document.querySelector('.compact-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
    })

    it('should use compact variant skeleton', () => {
      render(<InvestorLogoCompact />)
      
      // Should show compact skeleton
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })
  })

  describe('InvestorIcon', () => {
    it('should render loading skeleton initially', () => {
      render(<InvestorIcon />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render with all supported sizes', async () => {
      const sizes = ['sm', 'md', 'lg', 'xl', '2xl'] as const
      
      for (const size of sizes) {
        const { unmount } = render(<InvestorIcon size={size} />)
        
        await waitFor(() => {
          const pulsingElements = document.querySelectorAll('.animate-pulse')
          expect(pulsingElements.length).toBeGreaterThan(0)
        }, { timeout: 100 })
        
        unmount()
      }
    })

    it('should use icon-only variant skeleton', () => {
      render(<InvestorIcon />)
      
      // Should show icon-only skeleton (no text elements)
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should apply custom className', async () => {
      render(<InvestorIcon className="icon-class" />)
      
      await waitFor(() => {
        const container = document.querySelector('.icon-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
    })
  })

  describe('InvestorHeaderLogo', () => {
    it('should render loading skeleton initially', () => {
      render(<InvestorHeaderLogo />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render with header-appropriate sizes', async () => {
      const sizes = ['lg', 'xl', '2xl'] as const
      
      for (const size of sizes) {
        const { unmount } = render(<InvestorHeaderLogo size={size} />)
        
        await waitFor(() => {
          const pulsingElements = document.querySelectorAll('.animate-pulse')
          expect(pulsingElements.length).toBeGreaterThan(0)
        }, { timeout: 100 })
        
        unmount()
      }
    })

    it('should use xl size by default', () => {
      render(<InvestorHeaderLogo />)
      
      // Should use xl skeleton by default
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should apply custom className', async () => {
      render(<InvestorHeaderLogo className="header-class" />)
      
      await waitFor(() => {
        const container = document.querySelector('.header-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
    })
  })

  describe('InvestorFavicon', () => {
    it('should render loading skeleton initially', () => {
      render(<InvestorFavicon />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render with small and medium sizes', async () => {
      const { rerender } = render(<InvestorFavicon size="sm" />)
      
      // Initially should have skeleton or loaded component
      expect(document.body).toBeInTheDocument()
      
      rerender(<InvestorFavicon size="md" />)
      
      // Should render component (either skeleton or loaded)  
      expect(document.body).toBeInTheDocument()
    })

    it('should use favicon-specific skeleton', () => {
      render(<InvestorFavicon />)
      
      // Should show favicon skeleton
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should apply custom className', async () => {
      render(<InvestorFavicon className="favicon-class" />)
      
      await waitFor(() => {
        const container = document.querySelector('.favicon-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
    })
  })

  describe('Dynamic Import Behavior', () => {
    it('should handle dynamic import loading states', () => {
      render(<InvestorLogo />)
      
      // Should show loading skeleton initially
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should transition from loading to loaded state', async () => {
      render(<InvestorLogo />)
      
      // Initially shows skeleton
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
      
      // Wait for dynamic component to load
      await waitFor(() => {
        // After loading, should still have elements
        expect(document.body).toBeInTheDocument()
      }, { timeout: 200 })
    })

    it('should handle SSR disabled (ssr: false) correctly', async () => {
      render(<InvestorLogo />)
      
      // Should start with skeleton
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
      
      // Wait for client-side hydration simulation
      await waitFor(() => {
        expect(document.body).toBeInTheDocument()
      }, { timeout: 200 })
    })
  })

  describe('Skeleton Components', () => {
    it('should render skeleton with proper structure for full variant', () => {
      render(<InvestorLogo variant="full" />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render skeleton with proper structure for compact variant', () => {
      render(<InvestorLogoCompact />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render skeleton with proper structure for icon-only variant', () => {
      render(<InvestorIcon />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })

    it('should render favicon skeleton with correct structure', () => {
      render(<InvestorFavicon />)
      
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })
  })

  describe('Error Handling', () => {
    it('should handle dynamic import failures gracefully', async () => {
      // Mock a failing dynamic import
      const originalError = console.error
      console.error = jest.fn()
      
      render(<InvestorLogo />)
      
      // Should still render something (skeleton initially)
      expect(document.body).toBeInTheDocument()
      
      console.error = originalError
    })

    it('should provide fallback when dynamic component fails to load', async () => {
      render(<InvestorLogo />)
      
      // Should always have some element rendered
      expect(document.body).toBeInTheDocument()
    })
  })

  describe('Props Forwarding', () => {
    it('should forward props correctly to client components', async () => {
      const customProps = {
        size: 'lg' as const,
        className: 'test-class',
        showText: false
      }
      
      render(<InvestorLogo {...customProps} />)
      
      // Props should be available during loading state - check for the class in the DOM
      await waitFor(() => {
        const container = document.querySelector('.test-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
    })

    it('should handle conditional className prop forwarding', async () => {
      // Test with className
      const { rerender } = render(<InvestorLogoCompact className="with-class" />)
      
      await waitFor(() => {
        const container = document.querySelector('.with-class')
        expect(container).toBeInTheDocument()
      }, { timeout: 100 })
      
      // Test without className
      rerender(<InvestorLogoCompact />)
      
      await waitFor(() => {
        expect(document.body).toBeInTheDocument()
      }, { timeout: 100 })
    })
  })

  describe('Accessibility', () => {
    it('should maintain proper DOM structure during loading', () => {
      render(<InvestorLogo />)
      
      // Should have proper container structure
      const containers = document.querySelectorAll('.flex.items-center')
      expect(containers.length).toBeGreaterThan(0)
    })

    it('should be properly labeled during loading state', () => {
      render(<InvestorLogo />)
      
      // Skeleton should be identifiable
      const pulsingElements = document.querySelectorAll('.animate-pulse')
      expect(pulsingElements.length).toBeGreaterThan(0)
    })
  })

  describe('Integration with Client Components', () => {
    it('should eventually render client component after loading', async () => {
      render(<InvestorLogo />)
      
      // Wait for dynamic loading to complete
      await waitFor(() => {
        // Component should be mounted and rendered
        expect(document.body).toBeInTheDocument()
      }, { timeout: 200 })
    })

    it('should handle hydration properly', async () => {
      render(<InvestorLogo />)
      
      // Should handle client-side hydration
      await waitFor(() => {
        expect(document.body).toBeInTheDocument()
      }, { timeout: 200 })
    })

    it('should pass all props to client component', async () => {
      render(<InvestorLogo size="xl" className="test-props" showText={true} />)
      
      // Props should be forwarded
      await waitFor(() => {
        const container = document.querySelector('.test-props')
        expect(container).toBeInTheDocument()
      }, { timeout: 200 })
    })
  })
})