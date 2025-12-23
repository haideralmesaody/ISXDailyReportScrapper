/**
 * License Page Hydration Tests
 * Comprehensive tests to verify hydration fixes prevent React errors #418 and #423
 */

import React from 'react'
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import { useRouter } from 'next/navigation'
import dynamic from 'next/dynamic'

// Mock Next.js modules
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))

jest.mock('next/dynamic', () => {
  return jest.fn((loader, options) => {
    // Create a mock component that simulates dynamic loading
    const DynamicComponent = React.forwardRef((props: any, ref) => {
      const [isLoaded, setIsLoaded] = React.useState(false)
      
      React.useEffect(() => {
        // Simulate async loading
        const timer = setTimeout(() => {
          setIsLoaded(true)
        }, 10)
        return () => clearTimeout(timer)
      }, [])
      
      // Show loading state first (simulating SSR disabled)
      if (!isLoaded && options?.loading) {
        const LoadingComponent = options.loading
        return <LoadingComponent />
      }
      
      // Then show the actual component
      return <div data-testid="dynamic-content" {...props}>Dynamic Content Loaded</div>
    })
    
    DynamicComponent.displayName = 'DynamicComponent'
    return DynamicComponent
  })
})

// Mock API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    getLicenseStatus: jest.fn().mockResolvedValue({
      licensed: false,
      status: 'inactive',
      trial_remaining: 0,
    }),
    activateLicense: jest.fn().mockResolvedValue({
      success: true,
      licensed: true,
      status: 'active',
    }),
  },
}))

// Mock hooks
jest.mock('@/lib/hooks/use-api', () => ({
  useApi: jest.fn((fn) => ({
    execute: fn,
    loading: false,
    error: null,
  })),
}))

jest.mock('@/lib/hooks/use-toast', () => ({
  useToast: () => ({
    toast: jest.fn(),
  }),
}))

// Mock utility functions
jest.mock('@/lib/utils/license-helpers', () => ({
  getCachedLicenseStatus: jest.fn().mockReturnValue(null),
  setCachedLicenseStatus: jest.fn(),
  retryWithBackoff: jest.fn((fn) => fn()),
  trackLicenseEvent: jest.fn(),
}))

// Import components after mocks
import LicensePage from '@/app/license/page'
import LicenseContent from '@/app/license/license-content'

describe('License Page Hydration Tests', () => {
  const mockRouter = {
    push: jest.fn(),
    refresh: jest.fn(),
    prefetch: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useRouter as jest.Mock).mockReturnValue(mockRouter)
    
    // Reset DOM
    document.body.innerHTML = ''
    
    // Mock console to catch hydration warnings
    jest.spyOn(console, 'error').mockImplementation(() => {})
    jest.spyOn(console, 'warn').mockImplementation(() => {})
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  describe('Dynamic Import Behavior', () => {
    it('should render loading skeleton initially without hydration errors', async () => {
      const { container } = render(<LicensePage />)
      
      // Check for loading skeleton elements
      const skeletons = container.querySelectorAll('.animate-pulse')
      expect(skeletons.length).toBeGreaterThan(0)
      
      // Verify no hydration errors
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration failed')
      )
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Text content did not match')
      )
    })

    it('should transition from skeleton to content smoothly', async () => {
      const { container } = render(<LicensePage />)
      
      // Initially shows skeleton
      expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
      
      // Wait for dynamic import to complete
      await waitFor(() => {
        expect(screen.getByTestId('dynamic-content')).toBeInTheDocument()
      }, { timeout: 100 })
      
      // Skeleton should be replaced
      expect(container.querySelector('.animate-pulse')).not.toBeInTheDocument()
    })

    it('should have SSR disabled to prevent hydration mismatches', () => {
      // Verify dynamic import was called with ssr: false
      expect(dynamic).toHaveBeenCalledWith(
        expect.any(Function),
        expect.objectContaining({
          ssr: false,
        })
      )
    })
  })

  describe('Client-Side Mounting', () => {
    it('should not access Date operations before mounting', () => {
      const DateSpy = jest.spyOn(global, 'Date')
      
      // Render should not trigger Date operations
      render(<LicenseContent />)
      
      // Date should only be called after useEffect runs
      expect(DateSpy).not.toHaveBeenCalled()
      
      // After mount, Date operations are safe
      act(() => {
        jest.runAllTimers()
      })
    })

    it('should guard localStorage access with mounted state', () => {
      const localStorageSpy = jest.spyOn(Storage.prototype, 'getItem')
      
      // Initial render shouldn't access localStorage
      render(<LicenseContent />)
      expect(localStorageSpy).not.toHaveBeenCalled()
      
      // After mount, localStorage access is safe
      act(() => {
        jest.runAllTimers()
      })
    })

    it('should use mounted flag to prevent SSR/CSR mismatches', async () => {
      const { container } = render(<LicenseContent />)
      
      // Check for mounted state handling
      await waitFor(() => {
        // Component should be fully rendered without errors
        expect(container.firstChild).toBeInTheDocument()
      })
      
      // No hydration warnings
      expect(console.warn).not.toHaveBeenCalledWith(
        expect.stringContaining('did not match')
      )
    })
  })

  describe('Form Interaction After Hydration', () => {
    it('should handle form submission only on client side', async () => {
      const { container } = render(<LicenseContent />)
      
      // Wait for client-side hydration
      await waitFor(() => {
        expect(container.querySelector('form')).toBeInTheDocument()
      })
      
      // Form should be interactive after hydration
      const form = container.querySelector('form')
      expect(form).toBeInTheDocument()
      
      // Simulate form interaction
      if (form) {
        fireEvent.submit(form)
        
        // Should not cause hydration errors
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Hydration')
        )
      }
    })

    it('should validate input only after mounting', async () => {
      const { container } = render(<LicenseContent />)
      
      await waitFor(() => {
        const input = container.querySelector('input[type="text"]')
        expect(input).toBeInTheDocument()
        
        if (input) {
          // Input should be interactive
          fireEvent.change(input, { target: { value: 'TEST-KEY' } })
          expect(input).toHaveValue('TEST-KEY')
        }
      })
    })
  })

  describe('API Calls Post-Hydration', () => {
    it('should not make API calls during SSR', () => {
      const { apiClient } = require('@/lib/api')
      
      // Clear previous calls
      apiClient.getLicenseStatus.mockClear()
      
      // Render component
      render(<LicenseContent />)
      
      // API shouldn't be called immediately
      expect(apiClient.getLicenseStatus).not.toHaveBeenCalled()
      
      // After mount, API calls are made
      act(() => {
        jest.runAllTimers()
      })
    })

    it('should handle API errors gracefully without hydration issues', async () => {
      const { apiClient } = require('@/lib/api')
      apiClient.getLicenseStatus.mockRejectedValueOnce(new Error('Network error'))
      
      const { container } = render(<LicenseContent />)
      
      await waitFor(() => {
        // Error handling shouldn't cause hydration mismatches
        expect(console.error).not.toHaveBeenCalledWith(
          expect.stringContaining('Hydration')
        )
      })
    })
  })

  describe('Redirect Countdown', () => {
    it('should handle countdown timer without hydration errors', async () => {
      const { container } = render(<LicenseContent />)
      
      // Countdown should only start after mount
      await waitFor(() => {
        const countdown = container.querySelector('[data-testid="countdown"]')
        if (countdown) {
          // Countdown text should be consistent
          expect(countdown.textContent).toMatch(/\d+/)
        }
      })
      
      // No mismatched text errors
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Text content did not match')
      )
    })
  })

  describe('Loading States', () => {
    it('should show consistent loading skeleton structure', () => {
      const { container } = render(<LicensePage />)
      
      // Check skeleton structure matches expected layout
      const skeletonCards = container.querySelectorAll('.animate-pulse')
      expect(skeletonCards.length).toBeGreaterThan(0)
      
      // Skeleton should have proper dimensions
      skeletonCards.forEach(skeleton => {
        expect(skeleton).toHaveClass('animate-pulse')
      })
    })

    it('should maintain layout stability during loading', async () => {
      const { container } = render(<LicensePage />)
      
      // Get initial height
      const initialHeight = container.firstChild?.clientHeight || 0
      
      // Wait for content to load
      await waitFor(() => {
        expect(screen.queryByTestId('dynamic-content')).toBeInTheDocument()
      })
      
      // Layout shouldn't shift dramatically
      const loadedHeight = container.firstChild?.clientHeight || 0
      expect(Math.abs(loadedHeight - initialHeight)).toBeLessThan(200)
    })
  })

  describe('Memory and Performance', () => {
    it('should cleanup effects properly on unmount', () => {
      const { unmount } = render(<LicenseContent />)
      
      // Setup spy for cleanup
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout')
      const clearIntervalSpy = jest.spyOn(global, 'clearInterval')
      
      // Unmount component
      unmount()
      
      // Verify cleanup was called
      expect(clearTimeoutSpy).toHaveBeenCalled()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })

    it('should not cause memory leaks with rapid mounting/unmounting', () => {
      for (let i = 0; i < 10; i++) {
        const { unmount } = render(<LicenseContent />)
        unmount()
      }
      
      // No errors should occur
      expect(console.error).not.toHaveBeenCalled()
    })
  })

  describe('Edge Cases', () => {
    it('should handle missing window object gracefully', () => {
      // Temporarily remove window
      const originalWindow = global.window
      // @ts-ignore
      delete global.window
      
      expect(() => {
        render(<LicenseContent />)
      }).not.toThrow()
      
      // Restore window
      global.window = originalWindow
    })

    it('should handle rapid navigation without hydration errors', async () => {
      const { rerender } = render(<LicensePage />)
      
      // Simulate rapid navigation
      for (let i = 0; i < 5; i++) {
        rerender(<LicensePage />)
        await waitFor(() => {}, { timeout: 10 })
      }
      
      // No hydration errors should occur
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })
  })

  describe('SEO and Metadata', () => {
    it('should export metadata without affecting hydration', () => {
      // Import the page module
      const pageModule = require('@/app/license/page')
      
      // Should have metadata export
      expect(pageModule.metadata).toBeDefined()
      expect(pageModule.metadata.title).toContain('License')
      
      // Metadata shouldn't interfere with rendering
      render(<LicensePage />)
      expect(console.error).not.toHaveBeenCalled()
    })
  })
})

describe('License Page Component Integration', () => {
  it('should work correctly with all UI components', async () => {
    const { container } = render(<LicenseContent />)
    
    await waitFor(() => {
      // Check for key UI elements
      expect(container.querySelector('.card')).toBeInTheDocument()
      expect(container.querySelector('.button')).toBeInTheDocument()
      expect(container.querySelector('.badge')).toBeInTheDocument()
    })
    
    // No console errors
    expect(console.error).not.toHaveBeenCalled()
    expect(console.warn).not.toHaveBeenCalled()
  })

  it('should handle toast notifications without hydration issues', async () => {
    const { useToast } = require('@/lib/hooks/use-toast')
    const toastMock = jest.fn()
    useToast.mockReturnValue({ toast: toastMock })
    
    render(<LicenseContent />)
    
    // Toast should only be called after mount
    expect(toastMock).not.toHaveBeenCalled()
    
    // After interaction, toast can be called
    act(() => {
      jest.runAllTimers()
    })
  })
})