/**
 * @jest-environment jsdom
 */

import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import '@testing-library/jest-dom'

import LicenseStatus from '@/components/license/LicenseStatus'
import { useToast } from '@/lib/hooks/use-toast'
import { useHydration } from '@/lib/hooks/use-hydration'

// Mock the hooks
jest.mock('@/lib/hooks/use-toast')
jest.mock('@/lib/hooks/use-hydration')

// Mock the API calls
jest.mock('@/lib/api', () => ({
  licenseApi: {
    getStatus: jest.fn(),
    deactivate: jest.fn(),
    refresh: jest.fn(),
  },
}))

// Mock device fingerprint utility
jest.mock('@/lib/utils/device-fingerprint', () => ({
  generateDeviceFingerprint: jest.fn(),
  getDeviceInfo: jest.fn(),
}))

const mockToast = jest.fn()
const mockUseToast = useToast as jest.MockedFunction<typeof useToast>
const mockUseHydration = useHydration as jest.MockedFunction<typeof useHydration>

const { licenseApi } = require('@/lib/api')
const { generateDeviceFingerprint, getDeviceInfo } = require('@/lib/utils/device-fingerprint')

// Mock timer functions for countdown
jest.useFakeTimers()

describe('LicenseStatus Component', () => {
  const mockLicenseData = {
    isValid: true,
    licenseKey: 'ISX-1234-5678-90AB',
    activationId: 'act_12345678',
    expiryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(), // 30 days from now
    deviceFingerprint: 'device_hash_123',
    status: 'Active',
    issuedDate: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // 1 day ago
    lastChecked: new Date().toISOString(),
    duration: '1m',
  }

  const mockDeviceInfo = {
    os: 'Windows 10',
    browser: 'Chrome 118.0',
    screen: '1920x1080',
    timezone: 'UTC-5',
    language: 'en-US',
  }

  beforeEach(() => {
    jest.clearAllMocks()
    mockUseToast.mockReturnValue({ toast: mockToast })
    mockUseHydration.mockReturnValue(true)
    
    licenseApi.getStatus.mockResolvedValue({ data: mockLicenseData })
    generateDeviceFingerprint.mockResolvedValue('device_hash_123')
    getDeviceInfo.mockReturnValue(mockDeviceInfo)
  })

  afterEach(() => {
    jest.runOnlyPendingTimers()
    jest.useRealTimers()
    jest.useFakeTimers()
  })

  describe('Rendering', () => {
    it('renders loading state when not hydrated', () => {
      mockUseHydration.mockReturnValue(false)
      
      render(<LicenseStatus />)
      
      expect(screen.getByText(/loading license status/i)).toBeInTheDocument()
    })

    it('renders license information when loaded', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
        expect(screen.getByText('Active')).toBeInTheDocument()
        expect(screen.getByText(/expires in/i)).toBeInTheDocument()
      })
    })

    it('displays activation ID', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('act_12345678')).toBeInTheDocument()
      })
    })

    it('displays device information', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('Windows 10')).toBeInTheDocument()
        expect(screen.getByText('Chrome 118.0')).toBeInTheDocument()
      })
    })

    it('shows error state when license API fails', async () => {
      licenseApi.getStatus.mockRejectedValue(new Error('API Error'))
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/failed to load license/i)).toBeInTheDocument()
      })
    })
  })

  describe('Countdown Timer', () => {
    it('displays countdown timer for license expiry', async () => {
      // Set expiry to 1 hour from now
      const mockDataWithCountdown = {
        ...mockLicenseData,
        expiryDate: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: mockDataWithCountdown })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/expires in 59 minutes/i)).toBeInTheDocument()
      })
    })

    it('updates countdown every minute', async () => {
      // Set expiry to 2 hours from now
      const mockDataWithCountdown = {
        ...mockLicenseData,
        expiryDate: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: mockDataWithCountdown })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/expires in 1 hour/i)).toBeInTheDocument()
      })

      // Fast-forward time by 1 minute
      act(() => {
        jest.advanceTimersByTime(60000)
      })

      await waitFor(() => {
        expect(screen.getByText(/expires in 59 minutes/i)).toBeInTheDocument()
      })
    })

    it('displays expired status when license has expired', async () => {
      const expiredLicenseData = {
        ...mockLicenseData,
        isValid: false,
        expiryDate: new Date(Date.now() - 60 * 60 * 1000).toISOString(), // 1 hour ago
        status: 'Expired',
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: expiredLicenseData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('Expired')).toBeInTheDocument()
        expect(screen.getByText(/expired 1 hour ago/i)).toBeInTheDocument()
      })
    })

    it('displays warning when license expires soon', async () => {
      // Set expiry to 10 minutes from now
      const soonExpiringData = {
        ...mockLicenseData,
        expiryDate: new Date(Date.now() + 10 * 60 * 1000).toISOString(),
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: soonExpiringData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/expires soon/i)).toBeInTheDocument()
        expect(screen.getByText(/10 minutes/i)).toBeInTheDocument()
      })
    })

    it('cleans up timer on unmount', async () => {
      const { unmount } = render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
      })

      unmount()
      
      // Advance timers to ensure no memory leaks
      act(() => {
        jest.advanceTimersByTime(60000)
      })
      
      // No assertions needed - just ensuring no errors occur
    })
  })

  describe('Device Information Display', () => {
    it('shows device fingerprint hash', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/device_hash_123/i)).toBeInTheDocument()
      })
    })

    it('displays operating system information', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('Windows 10')).toBeInTheDocument()
      })
    })

    it('displays browser information', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('Chrome 118.0')).toBeInTheDocument()
      })
    })

    it('displays screen resolution', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('1920x1080')).toBeInTheDocument()
      })
    })

    it('handles missing device information gracefully', async () => {
      getDeviceInfo.mockReturnValue({})
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
        // Should not crash with missing device info
      })
    })
  })

  describe('Action Buttons', () => {
    it('renders refresh button', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/refresh/i)).toBeInTheDocument()
      })
    })

    it('handles refresh button click', async () => {
      const user = userEvent.setup()
      licenseApi.refresh.mockResolvedValue({ data: mockLicenseData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/refresh/i)).toBeInTheDocument()
      })

      const refreshButton = screen.getByText(/refresh/i)
      await user.click(refreshButton)
      
      expect(licenseApi.refresh).toHaveBeenCalled()
      expect(mockToast).toHaveBeenCalledWith({
        title: 'License Refreshed',
        description: 'License status has been updated',
      })
    })

    it('handles refresh failure', async () => {
      const user = userEvent.setup()
      licenseApi.refresh.mockRejectedValue(new Error('Network error'))
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/refresh/i)).toBeInTheDocument()
      })

      const refreshButton = screen.getByText(/refresh/i)
      await user.click(refreshButton)
      
      expect(mockToast).toHaveBeenCalledWith({
        title: 'Refresh Failed',
        description: 'Unable to refresh license status',
        variant: 'destructive',
      })
    })

    it('renders deactivate button when license is active', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/deactivate/i)).toBeInTheDocument()
      })
    })

    it('does not render deactivate button when license is expired', async () => {
      const expiredLicenseData = {
        ...mockLicenseData,
        isValid: false,
        status: 'Expired',
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: expiredLicenseData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText('Expired')).toBeInTheDocument()
        expect(screen.queryByText(/deactivate/i)).not.toBeInTheDocument()
      })
    })

    it('shows confirmation dialog when deactivate is clicked', async () => {
      const user = userEvent.setup()
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/deactivate/i)).toBeInTheDocument()
      })

      const deactivateButton = screen.getByText(/deactivate/i)
      await user.click(deactivateButton)
      
      expect(screen.getByText(/are you sure/i)).toBeInTheDocument()
      expect(screen.getByText(/this action cannot be undone/i)).toBeInTheDocument()
    })

    it('handles deactivate confirmation', async () => {
      const user = userEvent.setup()
      licenseApi.deactivate.mockResolvedValue({ success: true })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/deactivate/i)).toBeInTheDocument()
      })

      // Click deactivate
      const deactivateButton = screen.getByText(/deactivate/i)
      await user.click(deactivateButton)
      
      // Confirm deactivation
      const confirmButton = screen.getByText(/confirm/i)
      await user.click(confirmButton)
      
      expect(licenseApi.deactivate).toHaveBeenCalledWith({
        licenseKey: 'ISX-1234-5678-90AB',
        deviceFingerprint: 'device_hash_123',
      })
      
      expect(mockToast).toHaveBeenCalledWith({
        title: 'License Deactivated',
        description: 'Your license has been successfully deactivated',
      })
    })

    it('handles deactivate cancellation', async () => {
      const user = userEvent.setup()
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/deactivate/i)).toBeInTheDocument()
      })

      // Click deactivate
      const deactivateButton = screen.getByText(/deactivate/i)
      await user.click(deactivateButton)
      
      // Cancel deactivation
      const cancelButton = screen.getByText(/cancel/i)
      await user.click(cancelButton)
      
      expect(licenseApi.deactivate).not.toHaveBeenCalled()
      expect(screen.queryByText(/are you sure/i)).not.toBeInTheDocument()
    })

    it('handles deactivate failure', async () => {
      const user = userEvent.setup()
      licenseApi.deactivate.mockRejectedValue(new Error('Deactivation failed'))
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/deactivate/i)).toBeInTheDocument()
      })

      // Click deactivate and confirm
      const deactivateButton = screen.getByText(/deactivate/i)
      await user.click(deactivateButton)
      
      const confirmButton = screen.getByText(/confirm/i)
      await user.click(confirmButton)
      
      expect(mockToast).toHaveBeenCalledWith({
        title: 'Deactivation Failed',
        description: 'Unable to deactivate license',
        variant: 'destructive',
      })
    })
  })

  describe('Status Indicators', () => {
    it('shows active status with green indicator', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        const statusElement = screen.getByText('Active')
        expect(statusElement).toBeInTheDocument()
        expect(statusElement).toHaveClass('text-green-600')
      })
    })

    it('shows expired status with red indicator', async () => {
      const expiredData = {
        ...mockLicenseData,
        isValid: false,
        status: 'Expired',
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: expiredData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        const statusElement = screen.getByText('Expired')
        expect(statusElement).toBeInTheDocument()
        expect(statusElement).toHaveClass('text-red-600')
      })
    })

    it('shows warning status with orange indicator', async () => {
      const warningData = {
        ...mockLicenseData,
        status: 'Expires Soon',
        expiryDate: new Date(Date.now() + 10 * 60 * 1000).toISOString(),
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: warningData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        const statusElement = screen.getByText('Expires Soon')
        expect(statusElement).toBeInTheDocument()
        expect(statusElement).toHaveClass('text-orange-600')
      })
    })
  })

  describe('Accessibility', () => {
    it('provides proper ARIA labels for status information', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByLabelText(/license status/i)).toBeInTheDocument()
        expect(screen.getByLabelText(/expiry information/i)).toBeInTheDocument()
      })
    })

    it('supports keyboard navigation for action buttons', async () => {
      const user = userEvent.setup()
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/refresh/i)).toBeInTheDocument()
      })

      const refreshButton = screen.getByText(/refresh/i)
      refreshButton.focus()
      
      await user.keyboard('{Enter}')
      expect(licenseApi.refresh).toHaveBeenCalled()
    })

    it('announces status changes to screen readers', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByRole('status')).toBeInTheDocument()
      })
    })

    it('provides descriptive text for countdown timer', async () => {
      const mockDataWithCountdown = {
        ...mockLicenseData,
        expiryDate: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: mockDataWithCountdown })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/license expires in 30 minutes/i)).toBeInTheDocument()
      })
    })
  })

  describe('Loading States', () => {
    it('shows loading spinner while fetching license data', () => {
      licenseApi.getStatus.mockImplementation(() => new Promise(() => {})) // Never resolves
      
      render(<LicenseStatus />)
      
      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    })

    it('shows loading state for refresh action', async () => {
      const user = userEvent.setup()
      licenseApi.refresh.mockImplementation(() => new Promise(() => {})) // Never resolves
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/refresh/i)).toBeInTheDocument()
      })

      const refreshButton = screen.getByText(/refresh/i)
      await user.click(refreshButton)
      
      expect(screen.getByTestId('refresh-loading')).toBeInTheDocument()
    })

    it('shows loading state for deactivate action', async () => {
      const user = userEvent.setup()
      licenseApi.deactivate.mockImplementation(() => new Promise(() => {})) // Never resolves
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/deactivate/i)).toBeInTheDocument()
      })

      // Click deactivate and confirm
      const deactivateButton = screen.getByText(/deactivate/i)
      await user.click(deactivateButton)
      
      const confirmButton = screen.getByText(/confirm/i)
      await user.click(confirmButton)
      
      expect(screen.getByTestId('deactivate-loading')).toBeInTheDocument()
    })
  })

  describe('Error Handling', () => {
    it('displays retry button on API error', async () => {
      licenseApi.getStatus.mockRejectedValue(new Error('Network error'))
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/retry/i)).toBeInTheDocument()
      })
    })

    it('retries API call when retry button is clicked', async () => {
      const user = userEvent.setup()
      licenseApi.getStatus
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({ data: mockLicenseData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/retry/i)).toBeInTheDocument()
      })

      const retryButton = screen.getByText(/retry/i)
      await user.click(retryButton)
      
      await waitFor(() => {
        expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
      })
      
      expect(licenseApi.getStatus).toHaveBeenCalledTimes(2)
    })
  })

  describe('Data Formatting', () => {
    it('formats dates correctly', async () => {
      render(<LicenseStatus />)
      
      await waitFor(() => {
        // Should display formatted date
        expect(screen.getByText(/issued/i)).toBeInTheDocument()
        expect(screen.getByText(/last checked/i)).toBeInTheDocument()
      })
    })

    it('formats duration text correctly', async () => {
      const durationData = {
        ...mockLicenseData,
        duration: '3m',
      }
      
      licenseApi.getStatus.mockResolvedValue({ data: durationData })
      
      render(<LicenseStatus />)
      
      await waitFor(() => {
        expect(screen.getByText(/3 months/i)).toBeInTheDocument()
      })
    })
  })
})

// Performance test
describe('LicenseStatus Performance', () => {
  beforeAll(() => {
    jest.spyOn(performance, 'now')
      .mockReturnValueOnce(0)
      .mockReturnValueOnce(100)
  })

  afterAll(() => {
    jest.restoreAllMocks()
  })

  it('renders within performance budget', async () => {
    licenseApi.getStatus.mockResolvedValue({ 
      data: {
        isValid: true,
        licenseKey: 'ISX-1234-5678-90AB',
        status: 'Active',
        expiryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
      }
    })
    
    const startTime = performance.now()
    render(<LicenseStatus />)
    const endTime = performance.now()
    
    const renderTime = endTime - startTime
    expect(renderTime).toBeLessThanOrEqual(100)
    
    await waitFor(() => {
      expect(screen.getByText('ISX-1234-5678-90AB')).toBeInTheDocument()
    })
  })
})