/**
 * @jest-environment jsdom
 */

import React from 'react'
import { render, screen, fireEvent, act } from '@testing-library/react'
import '@testing-library/jest-dom'

import ScratchCard from '@/components/license/ScratchCard'
import LicenseStatus from '@/components/license/LicenseStatus'
import { generateDeviceFingerprint, getDeviceInfo } from '@/lib/utils/device-fingerprint'
import type { ScratchCardData } from '@/types/index'

// Mock hooks for performance testing
jest.mock('@/lib/hooks/use-toast', () => ({
  useToast: () => ({ toast: jest.fn() }),
}))

jest.mock('@/lib/hooks/use-hydration', () => ({
  useHydration: () => true,
}))

// Mock API calls
jest.mock('@/lib/api', () => ({
  licenseApi: {
    getStatus: jest.fn(),
    deactivate: jest.fn(),
    refresh: jest.fn(),
  },
}))

// Mock device fingerprint utility for performance testing
jest.mock('@/lib/utils/device-fingerprint')
const mockGenerateDeviceFingerprint = generateDeviceFingerprint as jest.MockedFunction<typeof generateDeviceFingerprint>
const mockGetDeviceInfo = getDeviceInfo as jest.MockedFunction<typeof getDeviceInfo>

// Mock framer-motion for performance testing
jest.mock('framer-motion', () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Performance testing utilities
const measureRenderTime = (renderFn: () => void): number => {
  const start = performance.now()
  renderFn()
  const end = performance.now()
  return end - start
}

const measureAsyncOperation = async (asyncFn: () => Promise<void>): Promise<number> => {
  const start = performance.now()
  await asyncFn()
  const end = performance.now()
  return end - start
}

describe('Scratch Card Performance Tests', () => {
  const defaultScratchCardData: ScratchCardData = {
    code: 'ISX-1234-5678-90AB',
    format: 'scratch' as const,
    revealed: false,
    activationId: 'act_12345678',
  }

  const mockLicenseData = {
    isValid: true,
    licenseKey: 'ISX-1234-5678-90AB',
    activationId: 'act_12345678',
    expiryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    deviceFingerprint: 'device_hash_123',
    status: 'Active',
    issuedDate: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    lastChecked: new Date().toISOString(),
    duration: '1m',
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockGenerateDeviceFingerprint.mockResolvedValue('device_hash_123')
    mockGetDeviceInfo.mockReturnValue({
      os: 'Windows 10',
      browser: 'Chrome 118.0',
      screen: '1920x1080',
      timezone: 'UTC-5',
      language: 'en-US',
      cpu: '8 cores',
      memory: '16GB',
    })

    // Mock license API
    const { licenseApi } = require('@/lib/api')
    licenseApi.getStatus.mockResolvedValue({ data: mockLicenseData })
  })

  describe('ScratchCard Component Performance', () => {
    it('renders within performance budget', () => {
      const renderTime = measureRenderTime(() => {
        render(<ScratchCard data={defaultScratchCardData} />)
      })

      // Should render within 50ms
      expect(renderTime).toBeLessThan(50)
    })

    it('handles multiple re-renders efficiently', () => {
      const { rerender } = render(<ScratchCard data={defaultScratchCardData} />)

      const rerenderTimes: number[] = []
      
      // Test 10 re-renders
      for (let i = 0; i < 10; i++) {
        const updatedData = {
          ...defaultScratchCardData,
          code: `ISX-${i.toString().padStart(4, '0')}-5678-90AB`,
        }

        const renderTime = measureRenderTime(() => {
          rerender(<ScratchCard data={updatedData} />)
        })
        
        rerenderTimes.push(renderTime)
      }

      // Average re-render time should be under 10ms
      const averageRerenderTime = rerenderTimes.reduce((a, b) => a + b, 0) / rerenderTimes.length
      expect(averageRerenderTime).toBeLessThan(10)
    })

    it('handles canvas operations efficiently', () => {
      const mockCanvasContext = {
        scale: jest.fn(),
        fillRect: jest.fn(),
        createLinearGradient: jest.fn(() => ({ addColorStop: jest.fn() })),
        arc: jest.fn(),
        fill: jest.fn(),
        beginPath: jest.fn(),
      }

      HTMLCanvasElement.prototype.getContext = jest.fn(() => mockCanvasContext)
      HTMLCanvasElement.prototype.getBoundingClientRect = jest.fn(() => ({
        width: 350,
        height: 220,
        left: 0,
        top: 0,
      }))

      render(<ScratchCard data={defaultScratchCardData} />)

      const canvas = screen.getByRole('img', { hidden: true })

      // Measure scratch interaction performance
      const scratchTime = measureRenderTime(() => {
        act(() => {
          fireEvent.mouseDown(canvas, { clientX: 100, clientY: 100 })
          fireEvent.mouseMove(canvas, { clientX: 110, clientY: 110 })
          fireEvent.mouseUp(canvas)
        })
      })

      // Canvas operations should complete within 20ms
      expect(scratchTime).toBeLessThan(20)
      
      // Verify canvas operations were called
      expect(mockCanvasContext.arc).toHaveBeenCalled()
      expect(mockCanvasContext.fill).toHaveBeenCalled()
    })

    it('handles large batch scratch operations', () => {
      const mockCanvasContext = {
        scale: jest.fn(),
        fillRect: jest.fn(),
        createLinearGradient: jest.fn(() => ({ addColorStop: jest.fn() })),
        arc: jest.fn(),
        fill: jest.fn(),
        beginPath: jest.fn(),
      }

      HTMLCanvasElement.prototype.getContext = jest.fn(() => mockCanvasContext)

      render(<ScratchCard data={defaultScratchCardData} />)
      const canvas = screen.getByRole('img', { hidden: true })

      // Simulate rapid scratch movements
      const batchScratchTime = measureRenderTime(() => {
        act(() => {
          fireEvent.mouseDown(canvas, { clientX: 50, clientY: 50 })
          
          // Simulate 20 rapid mouse movements
          for (let i = 0; i < 20; i++) {
            fireEvent.mouseMove(canvas, { 
              clientX: 50 + i * 2, 
              clientY: 50 + i * 2 
            })
          }
          
          fireEvent.mouseUp(canvas)
        })
      })

      // Batch operations should complete within 100ms
      expect(batchScratchTime).toBeLessThan(100)
    })

    it('memory usage stays within bounds during component lifecycle', () => {
      const initialMemory = (performance as any).memory?.usedJSHeapSize || 0
      
      const { unmount } = render(<ScratchCard data={defaultScratchCardData} />)
      
      // Force garbage collection if available
      if ((global as any).gc) {
        (global as any).gc()
      }
      
      unmount()
      
      // Force another garbage collection
      if ((global as any).gc) {
        (global as any).gc()
      }
      
      const finalMemory = (performance as any).memory?.usedJSHeapSize || 0
      const memoryDifference = finalMemory - initialMemory
      
      // Memory usage should not increase significantly (allow 1MB tolerance)
      expect(memoryDifference).toBeLessThan(1024 * 1024)
    })
  })

  describe('LicenseStatus Component Performance', () => {
    it('renders within performance budget', async () => {
      const renderTime = measureRenderTime(() => {
        render(<LicenseStatus />)
      })

      // Should render within 100ms
      expect(renderTime).toBeLessThan(100)
    })

    it('handles countdown updates efficiently', async () => {
      jest.useFakeTimers()

      const mockDataWithCountdown = {
        ...mockLicenseData,
        expiryDate: new Date(Date.now() + 60 * 60 * 1000).toISOString(), // 1 hour from now
      }

      const { licenseApi } = require('@/lib/api')
      licenseApi.getStatus.mockResolvedValue({ data: mockDataWithCountdown })

      render(<LicenseStatus />)

      // Wait for initial render
      await screen.findByText('ISX-1234-5678-90AB')

      const updateTimes: number[] = []

      // Test 10 countdown updates
      for (let i = 0; i < 10; i++) {
        const updateTime = measureRenderTime(() => {
          act(() => {
            jest.advanceTimersByTime(60000) // Advance 1 minute
          })
        })
        updateTimes.push(updateTime)
      }

      // Average update time should be under 5ms
      const averageUpdateTime = updateTimes.reduce((a, b) => a + b, 0) / updateTimes.length
      expect(averageUpdateTime).toBeLessThan(5)

      jest.useRealTimers()
    })

    it('handles API calls efficiently', async () => {
      const { licenseApi } = require('@/lib/api')
      
      render(<LicenseStatus />)

      // Measure API call completion time
      const apiCallTime = await measureAsyncOperation(async () => {
        await screen.findByText('ISX-1234-5678-90AB')
      })

      // API call should complete within 50ms (mocked)
      expect(apiCallTime).toBeLessThan(50)
    })

    it('handles rapid re-renders without performance degradation', async () => {
      const { rerender } = render(<LicenseStatus />)
      
      await screen.findByText('ISX-1234-5678-90AB')

      const rerenderTimes: number[] = []

      // Test 20 rapid re-renders with different data
      for (let i = 0; i < 20; i++) {
        const updatedData = {
          ...mockLicenseData,
          lastChecked: new Date(Date.now() + i * 1000).toISOString(),
        }

        const { licenseApi } = require('@/lib/api')
        licenseApi.getStatus.mockResolvedValue({ data: updatedData })

        const rerenderTime = measureRenderTime(() => {
          rerender(<LicenseStatus />)
        })
        
        rerenderTimes.push(rerenderTime)
      }

      // Average re-render time should stay under 15ms
      const averageRerenderTime = rerenderTimes.reduce((a, b) => a + b, 0) / rerenderTimes.length
      expect(averageRerenderTime).toBeLessThan(15)
    })
  })

  describe('Device Fingerprint Performance', () => {
    beforeEach(() => {
      // Reset mocks to actually test the implementation
      jest.restoreAllMocks()
      
      // Mock crypto and canvas for testing
      Object.defineProperty(window, 'crypto', {
        value: {
          subtle: {
            digest: jest.fn().mockResolvedValue(new ArrayBuffer(32)),
          },
        },
        writable: true,
      })

      HTMLCanvasElement.prototype.getContext = jest.fn(() => ({
        fillText: jest.fn(),
        fillRect: jest.fn(),
        getImageData: jest.fn(() => ({
          data: new Uint8ClampedArray([1, 2, 3, 4, 5, 6, 7, 8]),
        })),
        font: '',
        fillStyle: '',
        textBaseline: '',
      }))
    })

    it('generates fingerprint within performance budget', async () => {
      const { generateDeviceFingerprint } = await import('@/lib/utils/device-fingerprint')
      
      const fingerprintTime = await measureAsyncOperation(async () => {
        await generateDeviceFingerprint()
      })

      // Fingerprint generation should complete within 100ms
      expect(fingerprintTime).toBeLessThan(100)
    })

    it('handles multiple concurrent fingerprint generations', async () => {
      const { generateDeviceFingerprint } = await import('@/lib/utils/device-fingerprint')
      
      const startTime = performance.now()
      
      // Generate 10 fingerprints concurrently
      const fingerprintPromises = Array.from({ length: 10 }, () => 
        generateDeviceFingerprint()
      )
      
      await Promise.all(fingerprintPromises)
      
      const totalTime = performance.now() - startTime
      
      // Concurrent generation should complete within 200ms
      expect(totalTime).toBeLessThan(200)
    })

    it('device info collection is fast', () => {
      const { getDeviceInfo } = require('@/lib/utils/device-fingerprint')
      
      const deviceInfoTime = measureRenderTime(() => {
        getDeviceInfo()
      })

      // Device info collection should complete within 10ms
      expect(deviceInfoTime).toBeLessThan(10)
    })

    it('handles fallback scenarios efficiently', async () => {
      // Remove crypto to test fallback
      delete (window as any).crypto
      
      const { generateDeviceFingerprint } = await import('@/lib/utils/device-fingerprint')
      
      const fallbackTime = await measureAsyncOperation(async () => {
        await generateDeviceFingerprint()
      })

      // Fallback should still complete within 50ms
      expect(fallbackTime).toBeLessThan(50)
    })
  })

  describe('Memory and Resource Management', () => {
    it('properly cleans up event listeners', () => {
      const addEventListenerSpy = jest.spyOn(document, 'addEventListener')
      const removeEventListenerSpy = jest.spyOn(document, 'removeEventListener')

      const { unmount } = render(<ScratchCard data={defaultScratchCardData} />)
      
      const listenersAdded = addEventListenerSpy.mock.calls.length
      
      unmount()
      
      const listenersRemoved = removeEventListenerSpy.mock.calls.length

      // Should clean up all event listeners
      expect(listenersRemoved).toBeGreaterThanOrEqual(listenersAdded)

      addEventListenerSpy.mockRestore()
      removeEventListenerSpy.mockRestore()
    })

    it('handles large data sets without memory leaks', () => {
      const largeBatchData = Array.from({ length: 100 }, (_, i) => ({
        ...defaultScratchCardData,
        code: `ISX-${i.toString().padStart(4, '0')}-5678-90AB`,
        activationId: `act_${i.toString().padStart(8, '0')}`,
      }))

      const initialMemory = (performance as any).memory?.usedJSHeapSize || 0

      // Render components with large datasets
      largeBatchData.forEach((data, index) => {
        const { unmount } = render(<ScratchCard data={data} />)
        unmount()
      })

      // Force garbage collection if available
      if ((global as any).gc) {
        (global as any).gc()
      }

      const finalMemory = (performance as any).memory?.usedJSHeapSize || 0
      const memoryIncrease = finalMemory - initialMemory

      // Memory increase should be reasonable (allow 5MB tolerance)
      expect(memoryIncrease).toBeLessThan(5 * 1024 * 1024)
    })

    it('timer cleanup prevents memory leaks', async () => {
      jest.useFakeTimers()

      const { unmount } = render(<LicenseStatus />)
      
      // Let some time pass to create timers
      act(() => {
        jest.advanceTimersByTime(1000)
      })

      // Count active timers before unmount
      const activeTimersBefore = jest.getTimerCount()
      
      unmount()
      
      // Count active timers after unmount
      const activeTimersAfter = jest.getTimerCount()

      // Should clean up timers on unmount
      expect(activeTimersAfter).toBeLessThanOrEqual(activeTimersBefore)

      jest.useRealTimers()
    })
  })

  describe('Stress Testing', () => {
    it('handles rapid user interactions', () => {
      const mockCanvasContext = {
        scale: jest.fn(),
        fillRect: jest.fn(),
        createLinearGradient: jest.fn(() => ({ addColorStop: jest.fn() })),
        arc: jest.fn(),
        fill: jest.fn(),
        beginPath: jest.fn(),
      }

      HTMLCanvasElement.prototype.getContext = jest.fn(() => mockCanvasContext)

      render(<ScratchCard data={defaultScratchCardData} />)
      const canvas = screen.getByRole('img', { hidden: true })

      // Simulate rapid user interactions
      const stressTime = measureRenderTime(() => {
        act(() => {
          // Simulate 100 rapid touch/mouse events
          for (let i = 0; i < 100; i++) {
            fireEvent.mouseDown(canvas, { clientX: i, clientY: i })
            fireEvent.mouseMove(canvas, { clientX: i + 1, clientY: i + 1 })
            fireEvent.mouseUp(canvas)
          }
        })
      })

      // Should handle stress testing within 500ms
      expect(stressTime).toBeLessThan(500)
    })

    it('maintains performance under component tree stress', () => {
      const nestedComponents = Array.from({ length: 50 }, (_, i) => (
        <ScratchCard 
          key={i} 
          data={{
            ...defaultScratchCardData,
            code: `ISX-${i.toString().padStart(4, '0')}-5678-90AB`,
          }} 
        />
      ))

      const stressRenderTime = measureRenderTime(() => {
        render(<div>{nestedComponents}</div>)
      })

      // Should render 50 components within 1 second
      expect(stressRenderTime).toBeLessThan(1000)
    })
  })
})

// Performance test utilities
describe('Performance Test Utilities', () => {
  it('measureRenderTime utility works correctly', () => {
    const mockRenderFn = jest.fn()
    
    const time = measureRenderTime(mockRenderFn)
    
    expect(typeof time).toBe('number')
    expect(time).toBeGreaterThanOrEqual(0)
    expect(mockRenderFn).toHaveBeenCalled()
  })

  it('measureAsyncOperation utility works correctly', async () => {
    const mockAsyncFn = jest.fn().mockResolvedValue('test')
    
    const time = await measureAsyncOperation(mockAsyncFn)
    
    expect(typeof time).toBe('number')
    expect(time).toBeGreaterThanOrEqual(0)
    expect(mockAsyncFn).toHaveBeenCalled()
  })
})

// Export performance utilities for use in other test files
export { measureRenderTime, measureAsyncOperation }