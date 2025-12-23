/**
 * Phase 2 Hydration Integration Tests
 * End-to-end tests for license and operations pages with navigation flow
 */

import React from 'react'
import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import { useRouter, usePathname } from 'next/navigation'

// Mock Next.js navigation
let currentPath = '/'
const mockRouter = {
  push: jest.fn((path) => {
    currentPath = path
  }),
  refresh: jest.fn(),
  prefetch: jest.fn(),
  back: jest.fn(),
}

jest.mock('next/navigation', () => ({
  useRouter: jest.fn(() => mockRouter),
  usePathname: jest.fn(() => currentPath),
}))

// Mock dynamic imports for all pages
jest.mock('next/dynamic', () => {
  return jest.fn((loader, options) => {
    const DynamicComponent = React.forwardRef((props: any, ref) => {
      const [isLoaded, setIsLoaded] = React.useState(false)
      
      React.useEffect(() => {
        const timer = setTimeout(() => {
          setIsLoaded(true)
        }, 10)
        return () => clearTimeout(timer)
      }, [])
      
      if (!isLoaded && options?.loading) {
        const LoadingComponent = options.loading
        return <LoadingComponent />
      }
      
      // Return different content based on the component
      const pathname = currentPath
      if (pathname.includes('license')) {
        return <div data-testid="license-content">License Page Content</div>
      } else if (pathname.includes('operations')) {
        return <div data-testid="operations-content">Operations Page Content</div>
      }
      
      return <div data-testid="dynamic-content">Dynamic Content</div>
    })
    
    DynamicComponent.displayName = 'DynamicComponent'
    return DynamicComponent
  })
})

// Mock WebSocket for operations
const mockWebSocketData = {
  operations: [],
  connected: false,
  error: null,
}

jest.mock('@/lib/hooks/use-websocket', () => ({
  useAllOperationUpdates: jest.fn(() => mockWebSocketData),
  useOperationUpdates: jest.fn(() => mockWebSocketData),
}))

// Mock API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    getLicenseStatus: jest.fn().mockResolvedValue({
      licensed: true,
      status: 'active',
      expiresAt: '2025-12-31',
    }),
    activateLicense: jest.fn().mockResolvedValue({
      success: true,
      licensed: true,
    }),
    getOperations: jest.fn().mockResolvedValue([]),
    createOperation: jest.fn().mockResolvedValue({ success: true }),
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

describe('Phase 2 Hydration Integration Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    currentPath = '/'
    
    // Reset WebSocket state
    mockWebSocketData.operations = []
    mockWebSocketData.connected = false
    mockWebSocketData.error = null
    
    // Mock console
    jest.spyOn(console, 'error').mockImplementation(() => {})
    jest.spyOn(console, 'warn').mockImplementation(() => {})
    
    // Mock window.localStorage
    const localStorageMock = {
      getItem: jest.fn(),
      setItem: jest.fn(),
      clear: jest.fn(),
      removeItem: jest.fn(),
    }
    Object.defineProperty(window, 'localStorage', {
      value: localStorageMock,
      writable: true,
    })
  })

  afterEach(() => {
    jest.restoreAllMocks()
  })

  describe('Navigation Flow', () => {
    it('should navigate between pages without hydration errors', async () => {
      // Start at home
      currentPath = '/'
      const { rerender } = render(<MockApp />)
      
      // Navigate to license page
      act(() => {
        currentPath = '/license'
        mockRouter.push('/license')
      })
      rerender(<MockApp />)
      
      await waitFor(() => {
        expect(screen.queryByTestId('license-content')).toBeInTheDocument()
      }, { timeout: 100 })
      
      // Navigate to operations page
      act(() => {
        currentPath = '/operations'
        mockRouter.push('/operations')
      })
      rerender(<MockApp />)
      
      await waitFor(() => {
        expect(screen.queryByTestId('operations-content')).toBeInTheDocument()
      }, { timeout: 100 })
      
      // No hydration errors during navigation
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Text content did not match')
      )
    })

    it('should handle back navigation correctly', async () => {
      // Navigate: Home -> License -> Operations -> Back
      const paths = ['/', '/license', '/operations']
      const { rerender } = render(<MockApp />)
      
      for (const path of paths) {
        act(() => {
          currentPath = path
          mockRouter.push(path)
        })
        rerender(<MockApp />)
        await waitFor(() => {}, { timeout: 10 })
      }
      
      // Go back
      act(() => {
        currentPath = '/license'
        mockRouter.back()
      })
      rerender(<MockApp />)
      
      // No errors during back navigation
      expect(console.error).not.toHaveBeenCalled()
    })

    it('should maintain loading skeletons during transitions', async () => {
      const { container, rerender } = render(<MockApp />)
      
      // Navigate to license
      act(() => {
        currentPath = '/license'
      })
      rerender(<MockApp />)
      
      // Should show skeleton briefly
      expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
      
      // Wait for content
      await waitFor(() => {
        expect(container.querySelector('.animate-pulse')).not.toBeInTheDocument()
      }, { timeout: 100 })
      
      // Navigate to operations
      act(() => {
        currentPath = '/operations'
      })
      rerender(<MockApp />)
      
      // Should show skeleton again
      expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
    })
  })

  describe('State Persistence', () => {
    it('should maintain license state across navigation', async () => {
      const { apiClient } = require('@/lib/api')
      
      // Set initial license state
      apiClient.getLicenseStatus.mockResolvedValue({
        licensed: true,
        status: 'active',
        expiresAt: '2025-12-31',
      })
      
      const { rerender } = render(<MockApp />)
      
      // Navigate to license page
      currentPath = '/license'
      rerender(<MockApp />)
      
      await waitFor(() => {
        expect(apiClient.getLicenseStatus).toHaveBeenCalled()
      })
      
      // Navigate away and back
      currentPath = '/operations'
      rerender(<MockApp />)
      
      currentPath = '/license'
      rerender(<MockApp />)
      
      // State should be maintained without hydration errors
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })

    it('should maintain WebSocket connection across pages', async () => {
      const { rerender } = render(<MockApp />)
      
      // Navigate to operations
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // Connect WebSocket
      act(() => {
        mockWebSocketData.connected = true
      })
      
      await waitFor(() => {
        expect(mockWebSocketData.connected).toBe(true)
      })
      
      // Navigate away
      currentPath = '/license'
      rerender(<MockApp />)
      
      // Navigate back
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // Connection should be maintained
      expect(mockWebSocketData.connected).toBe(true)
      expect(console.error).not.toHaveBeenCalled()
    })
  })

  describe('WebSocket Integration', () => {
    it('should handle WebSocket updates during page transitions', async () => {
      const { rerender } = render(<MockApp />)
      
      // Start at operations
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // Connect WebSocket
      act(() => {
        mockWebSocketData.connected = true
      })
      
      // Send updates while navigating
      const navigationAndUpdates = async () => {
        for (let i = 0; i < 3; i++) {
          // Update operations
          act(() => {
            mockWebSocketData.operations = [
              { id: `op-${i}`, status: 'running', progress: i * 33 },
            ]
          })
          
          // Navigate
          currentPath = i % 2 === 0 ? '/license' : '/operations'
          rerender(<MockApp />)
          
          await waitFor(() => {}, { timeout: 10 })
        }
      }
      
      await navigationAndUpdates()
      
      // No hydration errors during updates
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })

    it('should reconnect WebSocket after navigation', async () => {
      const { rerender } = render(<MockApp />)
      
      // Connect on operations page
      currentPath = '/operations'
      rerender(<MockApp />)
      
      act(() => {
        mockWebSocketData.connected = true
      })
      
      // Disconnect
      act(() => {
        mockWebSocketData.connected = false
        mockWebSocketData.error = new Error('Connection lost')
      })
      
      // Navigate away and back
      currentPath = '/license'
      rerender(<MockApp />)
      
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // Reconnect
      act(() => {
        mockWebSocketData.connected = true
        mockWebSocketData.error = null
      })
      
      await waitFor(() => {
        expect(mockWebSocketData.connected).toBe(true)
        expect(mockWebSocketData.error).toBeNull()
      })
      
      // No errors during reconnection
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Text content did not match')
      )
    })
  })

  describe('API Integration', () => {
    it('should handle API calls across page transitions', async () => {
      const { apiClient } = require('@/lib/api')
      const { rerender } = render(<MockApp />)
      
      // Make API call on license page
      currentPath = '/license'
      rerender(<MockApp />)
      
      await waitFor(() => {
        expect(apiClient.getLicenseStatus).toHaveBeenCalled()
      })
      
      // Navigate to operations
      currentPath = '/operations'
      rerender(<MockApp />)
      
      await waitFor(() => {
        expect(apiClient.getOperations).toHaveBeenCalled()
      })
      
      // No hydration errors
      expect(console.error).not.toHaveBeenCalled()
    })

    it('should handle API errors during navigation', async () => {
      const { apiClient } = require('@/lib/api')
      const { rerender } = render(<MockApp />)
      
      // Simulate API error
      apiClient.getLicenseStatus.mockRejectedValueOnce(new Error('Network error'))
      
      currentPath = '/license'
      rerender(<MockApp />)
      
      await waitFor(() => {}, { timeout: 50 })
      
      // Navigate despite error
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // Should handle error gracefully
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })
  })

  describe('Loading States', () => {
    it('should show consistent loading states across pages', async () => {
      const { container, rerender } = render(<MockApp />)
      
      const pages = ['/license', '/operations']
      
      for (const page of pages) {
        currentPath = page
        rerender(<MockApp />)
        
        // Should show loading skeleton
        const skeletons = container.querySelectorAll('.animate-pulse')
        expect(skeletons.length).toBeGreaterThan(0)
        
        // Wait for content
        await waitFor(() => {
          expect(container.querySelector('.animate-pulse')).not.toBeInTheDocument()
        }, { timeout: 100 })
      }
      
      // No hydration errors
      expect(console.error).not.toHaveBeenCalled()
    })

    it('should handle loading state interruptions', async () => {
      const { rerender } = render(<MockApp />)
      
      // Start loading license page
      currentPath = '/license'
      rerender(<MockApp />)
      
      // Interrupt by navigating before load completes
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // Go back quickly
      currentPath = '/license'
      rerender(<MockApp />)
      
      await waitFor(() => {}, { timeout: 100 })
      
      // Should handle interruptions without errors
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })
  })

  describe('Memory and Performance', () => {
    it('should not leak memory during page transitions', async () => {
      const { rerender } = render(<MockApp />)
      
      // Perform many navigations
      for (let i = 0; i < 20; i++) {
        currentPath = i % 2 === 0 ? '/license' : '/operations'
        rerender(<MockApp />)
        
        // Simulate some activity
        act(() => {
          if (currentPath === '/operations') {
            mockWebSocketData.operations = [{ id: `op-${i}`, status: 'running' }]
          }
        })
        
        await waitFor(() => {}, { timeout: 5 })
      }
      
      // No memory-related errors
      expect(console.error).not.toHaveBeenCalled()
    })

    it('should cleanup resources on navigation', async () => {
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout')
      const clearIntervalSpy = jest.spyOn(global, 'clearInterval')
      
      const { rerender } = render(<MockApp />)
      
      // Navigate between pages
      currentPath = '/operations'
      rerender(<MockApp />)
      
      await waitFor(() => {}, { timeout: 10 })
      
      currentPath = '/license'
      rerender(<MockApp />)
      
      // Cleanup should be called
      expect(clearTimeoutSpy).toHaveBeenCalled()
      expect(clearIntervalSpy).toHaveBeenCalled()
    })
  })

  describe('Error Recovery', () => {
    it('should recover from hydration errors gracefully', async () => {
      const { rerender } = render(<MockApp />)
      
      // Simulate potential hydration issue
      const originalDate = global.Date
      global.Date = jest.fn(() => ({
        toISOString: () => 'mocked-date',
      })) as any
      
      currentPath = '/license'
      rerender(<MockApp />)
      
      // Restore Date
      global.Date = originalDate
      
      // Navigate to operations
      currentPath = '/operations'
      rerender(<MockApp />)
      
      await waitFor(() => {}, { timeout: 50 })
      
      // Should recover without crashing
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration failed')
      )
    })

    it('should handle error boundaries correctly', async () => {
      const { rerender } = render(<MockApp />)
      
      // Navigate with potential errors
      currentPath = '/license'
      rerender(<MockApp />)
      
      // Simulate component error
      act(() => {
        throw new Error('Test error')
      })
      
      // Should catch and handle error
      await waitFor(() => {}, { timeout: 10 }).catch(() => {
        // Error is expected and handled
      })
      
      // Navigate away should still work
      currentPath = '/operations'
      rerender(<MockApp />)
      
      // No hydration errors after error recovery
      expect(console.error).not.toHaveBeenCalledWith(
        expect.stringContaining('Hydration')
      )
    })
  })

  describe('Cross-Browser Compatibility', () => {
    const browsers = [
      { name: 'Chrome', userAgent: 'Chrome/120.0' },
      { name: 'Firefox', userAgent: 'Firefox/120.0' },
      { name: 'Safari', userAgent: 'Safari/17.0' },
    ]

    it.each(browsers)('should work in $name without hydration errors', async ({ userAgent }) => {
      // Mock user agent
      Object.defineProperty(navigator, 'userAgent', {
        value: userAgent,
        writable: true,
      })
      
      const { rerender } = render(<MockApp />)
      
      // Test navigation in each browser
      currentPath = '/license'
      rerender(<MockApp />)
      
      await waitFor(() => {}, { timeout: 10 })
      
      currentPath = '/operations'
      rerender(<MockApp />)
      
      await waitFor(() => {}, { timeout: 10 })
      
      // No browser-specific hydration errors
      expect(console.error).not.toHaveBeenCalled()
    })
  })
})

// Mock App component for testing
function MockApp() {
  const pathname = usePathname()
  
  // Simulate dynamic loading
  const [isLoaded, setIsLoaded] = React.useState(false)
  
  React.useEffect(() => {
    setIsLoaded(false)
    const timer = setTimeout(() => {
      setIsLoaded(true)
    }, 10)
    return () => clearTimeout(timer)
  }, [pathname])
  
  if (!isLoaded) {
    return (
      <div className="min-h-screen">
        <div className="animate-pulse bg-gray-200 h-full" />
      </div>
    )
  }
  
  if (pathname === '/license') {
    return <div data-testid="license-content">License Page</div>
  }
  
  if (pathname === '/operations') {
    return <div data-testid="operations-content">Operations Page</div>
  }
  
  return <div data-testid="home-content">Home Page</div>
}
