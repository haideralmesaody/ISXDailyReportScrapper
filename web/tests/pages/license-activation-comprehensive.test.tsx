/**
 * Comprehensive License Activation Page Tests
 * 
 * Tests form validation with Zod schema, submit button states, real-time validation feedback,
 * API integration with mocked responses, success and error state handling,
 * and professional Iraqi Investor branding verification.
 * 
 * Coverage Requirements:
 * - Form validation with Zod schema testing
 * - Submit button enabled/disabled states
 * - Real-time validation feedback
 * - API integration with mocked responses
 * - Success and error state handling
 * - Professional Iraqi Investor branding verification
 * - Live license key testing with ISX1M02LYE1F9QJHR9D7Z
 */

import React from 'react'
import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useRouter } from 'next/navigation'
import '@testing-library/jest-dom'

// Component under test
import LicenseActivationPage from '@/app/license/page'

// Mocked dependencies
import { useApi } from '@/lib/hooks/use-api'
import { useToast } from '@/lib/hooks/use-toast'

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}))

// Mock API hook
jest.mock('@/lib/hooks/use-api', () => ({
  useApi: jest.fn(),
}))

// Mock toast hook
jest.mock('@/lib/hooks/use-toast', () => ({
  useToast: jest.fn(),
}))

// Test utilities and types
interface MockRouterType {
  push: jest.Mock
  replace: jest.Mock
  prefetch: jest.Mock
  back: jest.Mock
  forward: jest.Mock
  refresh: jest.Mock
}

interface MockApiType {
  activateLicense: jest.Mock
  loading: boolean
  error: Error | null
}

interface MockToastType {
  toast: jest.Mock
}

// Valid test license key from requirements
const VALID_TEST_LICENSE_KEY = 'ISX1M02LYE1F9QJHR9D7Z'

describe('License Activation Page - Comprehensive Testing', () => {
  let mockRouter: MockRouterType
  let mockApi: MockApiType
  let mockToast: MockToastType
  let consoleErrorSpy: jest.SpyInstance

  beforeEach(() => {
    // Setup router mock
    mockRouter = {
      push: jest.fn(),
      replace: jest.fn(),
      prefetch: jest.fn(),
      back: jest.fn(),
      forward: jest.fn(),
      refresh: jest.fn(),
    }
    ;(useRouter as jest.Mock).mockReturnValue(mockRouter)

    // Setup API mock
    mockApi = {
      activateLicense: jest.fn(),
      loading: false,
      error: null,
    }
    ;(useApi as jest.Mock).mockReturnValue(mockApi)

    // Setup toast mock
    mockToast = {
      toast: jest.fn(),
    }
    ;(useToast as jest.Mock).mockReturnValue(mockToast)

    // Spy on console.error
    consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => {})

    // Clear all mocks
    jest.clearAllMocks()
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  describe('Initial Page Load and Branding', () => {
    it('should display Iraqi Investor branding and license activation form', () => {
      render(<LicenseActivationPage />)

      // Verify Iraqi Investor branding
      expect(screen.getByAltText('Iraqi Investor Logo')).toBeInTheDocument()
      expect(screen.getByText('Iraqi Investor License System')).toBeInTheDocument()
      expect(screen.getByText('Activate Your Professional License')).toBeInTheDocument()

      // Verify form elements
      expect(screen.getByLabelText('License Key')).toBeInTheDocument()
      expect(screen.getByLabelText('Email Address')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Activate License' })).toBeInTheDocument()
    })

    it('should have proper form accessibility attributes', () => {
      render(<LicenseActivationPage />)

      const form = screen.getByRole('form', { name: 'License Activation' })
      expect(form).toBeInTheDocument()

      const licenseInput = screen.getByLabelText('License Key')
      expect(licenseInput).toHaveAttribute('type', 'text')
      expect(licenseInput).toHaveAttribute('required')
      expect(licenseInput).toHaveAttribute('aria-describedby', 'license-key-error')

      const emailInput = screen.getByLabelText('Email Address')
      expect(emailInput).toHaveAttribute('type', 'email')
      expect(emailInput).toHaveAttribute('required')
      expect(emailInput).toHaveAttribute('aria-describedby', 'email-error')
    })

    it('should have proper meta tags and page title', () => {
      render(<LicenseActivationPage />)

      // Check if the page has the correct title via document head
      expect(document.title).toContain('License Activation')
    })
  })

  describe('Form Validation with Zod Schema', () => {
    it('should validate license key format in real-time', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')

      // Test invalid format
      await user.type(licenseInput, 'INVALID')
      await user.tab() // Trigger blur event

      await waitFor(() => {
        expect(screen.getByText('License key must be exactly 19 characters')).toBeInTheDocument()
      })

      // Test valid format
      await user.clear(licenseInput)
      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.tab()

      await waitFor(() => {
        expect(screen.queryByText('License key must be exactly 19 characters')).not.toBeInTheDocument()
      })
    })

    it('should validate email format in real-time', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const emailInput = screen.getByLabelText('Email Address')

      // Test invalid email
      await user.type(emailInput, 'invalid-email')
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('Please enter a valid email address')).toBeInTheDocument()
      })

      // Test valid email
      await user.clear(emailInput)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.tab()

      await waitFor(() => {
        expect(screen.queryByText('Please enter a valid email address')).not.toBeInTheDocument()
      })
    })

    it('should validate required fields', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      // Try to submit empty form
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('License key is required')).toBeInTheDocument()
        expect(screen.getByText('Email address is required')).toBeInTheDocument()
      })
    })

    it('should validate license key character requirements', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')

      // Test too short
      await user.type(licenseInput, 'ISX123')
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('License key must be exactly 19 characters')).toBeInTheDocument()
      })

      // Test too long
      await user.clear(licenseInput)
      await user.type(licenseInput, 'ISX1M02LYE1F9QJHR9D7Z123')
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('License key must be exactly 19 characters')).toBeInTheDocument()
      })

      // Test invalid characters
      await user.clear(licenseInput)
      await user.type(licenseInput, 'ISX1M02LYE1F9QJHR9D!')
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('License key contains invalid characters')).toBeInTheDocument()
      })
    })

    it('should validate email domain requirements', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const emailInput = screen.getByLabelText('Email Address')

      // Test valid format but restricted domain
      await user.type(emailInput, 'test@gmail.com')
      await user.tab()

      await waitFor(() => {
        expect(screen.getByText('Please use your professional email address')).toBeInTheDocument()
      })

      // Test approved domain
      await user.clear(emailInput)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.tab()

      await waitFor(() => {
        expect(screen.queryByText('Please use your professional email address')).not.toBeInTheDocument()
      })
    })
  })

  describe('Submit Button States', () => {
    it('should disable submit button when form is invalid', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      
      // Initially disabled (empty form)
      expect(submitButton).toBeDisabled()

      // Still disabled with invalid data
      const licenseInput = screen.getByLabelText('License Key')
      await user.type(licenseInput, 'INVALID')
      
      expect(submitButton).toBeDisabled()
    })

    it('should enable submit button when form is valid', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const submitButton = screen.getByRole('button', { name: 'Activate License' })
      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')

      // Fill valid data
      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')

      await waitFor(() => {
        expect(submitButton).toBeEnabled()
      })
    })

    it('should show loading state during submission', async () => {
      const user = userEvent.setup()
      mockApi.loading = true

      render(<LicenseActivationPage />)

      const submitButton = screen.getByRole('button', { name: 'Activating...' })
      expect(submitButton).toBeDisabled()
      expect(screen.getByTestId('submit-loading-spinner')).toBeInTheDocument()
    })

    it('should prevent double submission', async () => {
      const user = userEvent.setup()
      mockApi.activateLicense.mockImplementation(() => new Promise(() => {})) // Never resolves

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      // Fill valid data
      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')

      // First click
      await user.click(submitButton)
      
      // Button should be disabled to prevent double submission
      expect(submitButton).toBeDisabled()

      // Second click should not trigger another API call
      await user.click(submitButton)
      expect(mockApi.activateLicense).toHaveBeenCalledTimes(1)
    })
  })

  describe('API Integration - Success Scenarios', () => {
    it('should successfully activate license with valid test key', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      // Fill form with test data
      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      // Verify API call
      expect(mockApi.activateLicense).toHaveBeenCalledWith({
        license_key: VALID_TEST_LICENSE_KEY,
        email: 'test@iraqiinvestor.gov.iq'
      })

      // Wait for success state
      await waitFor(() => {
        expect(screen.getByText('License Activated Successfully!')).toBeInTheDocument()
      })

      // Verify success toast
      expect(mockToast.toast).toHaveBeenCalledWith({
        title: 'Success',
        description: 'Your license has been activated successfully.',
        variant: 'default'
      })
    })

    it('should display license details after successful activation', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('License Status: Active')).toBeInTheDocument()
        expect(screen.getByText('Expires: December 31, 2025')).toBeInTheDocument()
        expect(screen.getByText('Registered Email: test@iraqiinvestor.gov.iq')).toBeInTheDocument()
      })
    })

    it('should show countdown and redirect after successful activation', async () => {
      jest.useFakeTimers()
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Redirecting to dashboard in 5 seconds...')).toBeInTheDocument()
      })

      // Fast-forward countdown
      act(() => {
        jest.advanceTimersByTime(5000)
      })

      expect(mockRouter.push).toHaveBeenCalledWith('/dashboard')

      jest.useRealTimers()
    })

    it('should allow manual navigation during countdown', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Continue to Dashboard' })).toBeInTheDocument()
      })

      const continueButton = screen.getByRole('button', { name: 'Continue to Dashboard' })
      await user.click(continueButton)

      expect(mockRouter.push).toHaveBeenCalledWith('/dashboard')
    })
  })

  describe('API Integration - Error Scenarios', () => {
    it('should handle invalid license key error', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce({
        response: {
          status: 400,
          data: {
            type: '/problems/invalid-license',
            title: 'Invalid License Key',
            status: 400,
            detail: 'The provided license key is not valid',
            trace_id: 'trace-123'
          }
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, 'INVALID_LICENSE_KEY')
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Invalid License Key')).toBeInTheDocument()
        expect(screen.getByText('The provided license key is not valid')).toBeInTheDocument()
      })

      // Verify error toast
      expect(mockToast.toast).toHaveBeenCalledWith({
        title: 'Activation Failed',
        description: 'The provided license key is not valid',
        variant: 'destructive'
      })
    })

    it('should handle expired license error', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce({
        response: {
          status: 400,
          data: {
            type: '/problems/expired-license',
            title: 'License Expired',
            status: 400,
            detail: 'This license key has expired and cannot be activated',
            trace_id: 'trace-124'
          }
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, 'EXPIRED_LICENSE_KEY')
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('License Expired')).toBeInTheDocument()
        expect(screen.getByText('This license key has expired and cannot be activated')).toBeInTheDocument()
      })
    })

    it('should handle already activated license error', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce({
        response: {
          status: 409,
          data: {
            type: '/problems/license-already-active',
            title: 'License Already Active',
            status: 409,
            detail: 'This license is already activated on another device',
            trace_id: 'trace-125'
          }
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('License Already Active')).toBeInTheDocument()
        expect(screen.getByText('This license is already activated on another device')).toBeInTheDocument()
      })
    })

    it('should handle network connectivity errors', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce(new Error('Network Error'))

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Network Connection Error')).toBeInTheDocument()
        expect(screen.getByText('Please check your internet connection and try again')).toBeInTheDocument()
      })

      // Should show retry button
      expect(screen.getByRole('button', { name: 'Retry Activation' })).toBeInTheDocument()
    })

    it('should handle server errors gracefully', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce({
        response: {
          status: 500,
          data: {
            type: '/problems/internal-error',
            title: 'Internal Server Error',
            status: 500,
            detail: 'An unexpected error occurred',
            trace_id: 'trace-126'
          }
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Server Error')).toBeInTheDocument()
        expect(screen.getByText('Please try again later or contact support')).toBeInTheDocument()
      })
    })

    it('should provide retry functionality for failed activations', async () => {
      const user = userEvent.setup()
      
      // First attempt fails
      mockApi.activateLicense.mockRejectedValueOnce(new Error('Network Error'))
      
      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'Retry Activation' })).toBeInTheDocument()
      })

      // Second attempt succeeds
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      const retryButton = screen.getByRole('button', { name: 'Retry Activation' })
      await user.click(retryButton)

      expect(mockApi.activateLicense).toHaveBeenCalledTimes(2)

      await waitFor(() => {
        expect(screen.getByText('License Activated Successfully!')).toBeInTheDocument()
      })
    })
  })

  describe('Form Reset and Error Recovery', () => {
    it('should clear errors when user starts typing', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce({
        response: {
          status: 400,
          data: {
            type: '/problems/invalid-license',
            title: 'Invalid License Key',
            status: 400,
            detail: 'The provided license key is not valid'
          }
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      // Trigger error
      await user.type(licenseInput, 'INVALID_KEY')
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('Invalid License Key')).toBeInTheDocument()
      })

      // Start typing in license field
      await user.type(licenseInput, 'X')

      await waitFor(() => {
        expect(screen.queryByText('Invalid License Key')).not.toBeInTheDocument()
      })
    })

    it('should reset form after successful activation', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        expect(screen.getByText('License Activated Successfully!')).toBeInTheDocument()
      })

      // Form should be cleared after success
      expect(licenseInput).toHaveValue('')
      expect(emailInput).toHaveValue('')
    })
  })

  describe('Keyboard Navigation and Accessibility', () => {
    it('should support full keyboard navigation', async () => {
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      // Tab through form elements
      await user.tab()
      expect(screen.getByLabelText('License Key')).toHaveFocus()

      await user.tab()
      expect(screen.getByLabelText('Email Address')).toHaveFocus()

      await user.tab()
      expect(screen.getByRole('button', { name: 'Activate License' })).toHaveFocus()
    })

    it('should support form submission with Enter key', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockResolvedValueOnce({
        success: true,
        license: {
          valid: true,
          status: 'active',
          expires_at: '2025-12-31T23:59:59Z',
          email: 'test@iraqiinvestor.gov.iq',
          activated_at: '2025-07-28T14:30:00Z'
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')

      await user.type(licenseInput, VALID_TEST_LICENSE_KEY)
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      
      // Submit with Enter key on email field
      await user.keyboard('{Enter}')

      expect(mockApi.activateLicense).toHaveBeenCalledTimes(1)
    })

    it('should have proper ARIA labels and descriptions', () => {
      render(<LicenseActivationPage />)

      const form = screen.getByRole('form', { name: 'License Activation' })
      expect(form).toHaveAttribute('aria-label', 'License Activation')

      const licenseInput = screen.getByLabelText('License Key')
      expect(licenseInput).toHaveAttribute('aria-describedby')
      expect(licenseInput).toHaveAttribute('aria-required', 'true')

      const emailInput = screen.getByLabelText('Email Address')
      expect(emailInput).toHaveAttribute('aria-describedby')
      expect(emailInput).toHaveAttribute('aria-required', 'true')
    })

    it('should announce errors to screen readers', async () => {
      const user = userEvent.setup()
      
      mockApi.activateLicense.mockRejectedValueOnce({
        response: {
          status: 400,
          data: {
            type: '/problems/invalid-license',
            title: 'Invalid License Key',
            status: 400,
            detail: 'The provided license key is not valid'
          }
        }
      })

      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')
      const emailInput = screen.getByLabelText('Email Address')
      const submitButton = screen.getByRole('button', { name: 'Activate License' })

      await user.type(licenseInput, 'INVALID_KEY')
      await user.type(emailInput, 'test@iraqiinvestor.gov.iq')
      await user.click(submitButton)

      await waitFor(() => {
        const errorAlert = screen.getByRole('alert')
        expect(errorAlert).toBeInTheDocument()
        expect(errorAlert).toHaveAttribute('aria-live', 'assertive')
      })
    })
  })

  describe('Performance and Memory Management', () => {
    it('should debounce form validation', async () => {
      jest.useFakeTimers()
      const user = userEvent.setup()
      render(<LicenseActivationPage />)

      const licenseInput = screen.getByLabelText('License Key')

      // Type multiple characters rapidly
      await user.type(licenseInput, 'RAPID')

      // Should not validate immediately
      expect(screen.queryByText('License key must be exactly 19 characters')).not.toBeInTheDocument()

      // Fast-forward past debounce delay
      act(() => {
        jest.advanceTimersByTime(500)
      })

      await waitFor(() => {
        expect(screen.getByText('License key must be exactly 19 characters')).toBeInTheDocument()
      })

      jest.useRealTimers()
    })

    it('should cleanup timers on unmount', () => {
      const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout')
      const clearIntervalSpy = jest.spyOn(global, 'clearInterval')

      const { unmount } = render(<LicenseActivationPage />)
      unmount()

      expect(clearTimeoutSpy).toHaveBeenCalled()
      expect(clearIntervalSpy).toHaveBeenCalled()

      clearTimeoutSpy.mockRestore()
      clearIntervalSpy.mockRestore()
    })
  })
})