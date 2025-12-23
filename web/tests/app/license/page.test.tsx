import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LicensePage from '@/app/license/page'

// Mock the API client
const mockApiClient = {
  activateLicense: jest.fn(),
  getLicenseStatus: jest.fn(),
}

// Mock hooks
jest.mock('@/lib/hooks/use-websocket', () => ({
  useLicenseStatus: () => ({
    isValid: false,
    expiresAt: null,
    features: [],
    userInfo: null,
  }),
}))

jest.mock('@/lib/hooks/use-api', () => ({
  useApi: (fn: any) => {
    if (fn === mockApiClient.activateLicense) {
      return {
        execute: jest.fn().mockImplementation(async (data) => {
          return mockApiClient.activateLicense(data)
        }),
        loading: false,
        error: null,
      }
    }
    return {
      execute: jest.fn().mockResolvedValue({}),
      loading: false,
      error: null,
    }
  },
}))

jest.mock('@/lib/hooks/use-toast', () => ({
  useToast: () => ({
    toast: jest.fn(),
  }),
}))

jest.mock('@/lib/api', () => ({
  apiClient: mockApiClient,
}))

describe('LicensePage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockApiClient.getLicenseStatus.mockResolvedValue({
      license_status: 'not_activated',
    })
  })

  it('renders license status information correctly', () => {
    render(<LicensePage />)

    // Check page header
    expect(screen.getByText('Professional License')).toBeInTheDocument()
    expect(screen.getByText(/Activate your Iraqi Investor professional license/)).toBeInTheDocument()

    // Check license status card
    expect(screen.getByText('License Status')).toBeInTheDocument()
    expect(screen.getByText('Inactive')).toBeInTheDocument()
    expect(screen.getByText(/Activate your license to unlock the full power/)).toBeInTheDocument()
  })

  it('shows activation form when license is not valid', () => {
    render(<LicensePage />)

    expect(screen.getByText('Activate Professional License')).toBeInTheDocument()
    expect(screen.getByLabelText('License Key *')).toBeInTheDocument()
    expect(screen.getByLabelText('Organization (Optional)')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Activate License' })).toBeInTheDocument()
  })

  it('validates license key format', async () => {
    const user = userEvent.setup()
    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    const submitButton = screen.getByRole('button', { name: 'Activate License' })

    // Test invalid format
    await user.type(licenseKeyInput, 'INVALID-KEY')
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/Invalid license key format/)).toBeInTheDocument()
    })
  })

  it('accepts valid ISX license key format', async () => {
    const user = userEvent.setup()
    mockApiClient.activateLicense.mockResolvedValue({
      success: true,
      message: 'License activated successfully',
    })

    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    const submitButton = screen.getByRole('button', { name: 'Activate License' })

    // Test valid ISX format
    await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
    
    // Submit button should be enabled for valid format
    expect(submitButton).not.toBeDisabled()
  })

  it('handles successful license activation', async () => {
    const user = userEvent.setup()
    const mockToast = jest.fn()
    
    // Mock successful activation
    mockApiClient.activateLicense.mockResolvedValue({
      success: true,
      message: 'License activated successfully',
    })

    // Mock toast
    require('@/lib/hooks/use-toast').useToast.mockReturnValue({
      toast: mockToast,
    })

    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    const organizationInput = screen.getByLabelText('Organization (Optional)')
    const submitButton = screen.getByRole('button', { name: 'Activate License' })

    // Fill form
    await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
    await user.type(organizationInput, 'Iraqi Investment Bank')

    // Submit form
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockApiClient.activateLicense).toHaveBeenCalledWith({
        license_key: 'ISX1Y-ABCDE-12345-FGHIJ-67890',
        organization: 'Iraqi Investment Bank',
      })
    })

    // Should show success toast
    expect(mockToast).toHaveBeenCalledWith({
      title: 'Welcome to The Iraqi Investor',
      description: 'Your professional license has been activated successfully.',
      variant: 'default',
    })
  })

  it('handles license activation errors', async () => {
    const user = userEvent.setup()
    const mockToast = jest.fn()

    // Mock activation error
    const mockUseApi = require('@/lib/hooks/use-api').useApi
    mockUseApi.mockReturnValue({
      execute: jest.fn().mockRejectedValue(new Error('License key not found')),
      loading: false,
      error: {
        detail: 'License key not found',
        traceId: 'trace-123',
      },
    })

    require('@/lib/hooks/use-toast').useToast.mockReturnValue({
      toast: mockToast,
    })

    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    const submitButton = screen.getByRole('button', { name: 'Activate License' })

    await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
    await user.click(submitButton)

    await waitFor(() => {
      expect(mockToast).toHaveBeenCalledWith({
        title: 'Activation Failed',
        description: expect.stringContaining('License key not found'),
        variant: 'destructive',
      })
    })
  })

  it('shows loading state during activation', async () => {
    const user = userEvent.setup()

    // Mock loading state
    const mockUseApi = require('@/lib/hooks/use-api').useApi
    mockUseApi.mockReturnValue({
      execute: jest.fn().mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100))),
      loading: true,
      error: null,
    })

    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    const submitButton = screen.getByRole('button', { name: /Activating License/ })

    await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')

    expect(submitButton).toBeDisabled()
    expect(screen.getByText('Activating License...')).toBeInTheDocument()
    expect(screen.getByText('Activating license...')).toBeInTheDocument()
  })

  it('displays active license information when valid', () => {
    // Mock active license
    const mockUseLicenseStatus = require('@/lib/hooks/use-websocket').useLicenseStatus
    mockUseLicenseStatus.mockReturnValue({
      isValid: true,
      expiresAt: '2025-12-31T23:59:59Z',
      features: ['Daily Reports Access', 'Advanced Analytics', 'Iraqi Stock Exchange Integration'],
      userInfo: {
        name: 'Test User',
        email: 'test@iraqiinvestor.gov.iq',
        organization: 'Iraqi Investment Bank',
      },
    })

    render(<LicensePage />)

    // Should show active status
    expect(screen.getByText('Active')).toBeInTheDocument()
    expect(screen.getByText('Licensed')).toBeInTheDocument()
    expect(screen.getByText(/Your Iraqi Investor professional license is active/)).toBeInTheDocument()

    // Should show user information
    expect(screen.getByText('Test User')).toBeInTheDocument()
    expect(screen.getByText('test@iraqiinvestor.gov.iq')).toBeInTheDocument()
    expect(screen.getByText('Iraqi Investment Bank')).toBeInTheDocument()

    // Should show active features
    expect(screen.getByText('Active Features')).toBeInTheDocument()
    expect(screen.getByText('DAILY REPORTS ACCESS')).toBeInTheDocument()
    expect(screen.getByText('ADVANCED ANALYTICS')).toBeInTheDocument()

    // Should not show activation form
    expect(screen.queryByText('Activate Professional License')).not.toBeInTheDocument()
  })

  it('formats expiration date correctly', () => {
    const mockUseLicenseStatus = require('@/lib/hooks/use-websocket').useLicenseStatus
    
    // Test different expiration scenarios
    const testCases = [
      {
        expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(), // Tomorrow
        expected: 'Expires tomorrow',
      },
      {
        expiresAt: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(), // 5 days
        expected: 'Expires in 5 days',
      },
      {
        expiresAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // Yesterday
        expected: 'Expired',
      },
    ]

    testCases.forEach(({ expiresAt, expected }) => {
      mockUseLicenseStatus.mockReturnValue({
        isValid: true,
        expiresAt,
        features: [],
        userInfo: null,
      })

      const { unmount } = render(<LicensePage />)
      expect(screen.getByText(expected)).toBeInTheDocument()
      unmount()
    })
  })

  it('shows Iraqi Investor branding and information', () => {
    render(<LicensePage />)

    // Check branding elements
    expect(screen.getByText('Iraqi Investor Professional Edition')).toBeInTheDocument()
    expect(screen.getByText(/Complete market intelligence platform/)).toBeInTheDocument()

    // Check feature descriptions
    expect(screen.getByText('Market Intelligence')).toBeInTheDocument()
    expect(screen.getByText('Professional Features')).toBeInTheDocument()
    expect(screen.getByText(/Real-time Iraqi Stock Exchange data/)).toBeInTheDocument()
    expect(screen.getByText(/Enterprise-grade security and encryption/)).toBeInTheDocument()

    // Check about section
    expect(screen.getByText('About The Iraqi Investor')).toBeInTheDocument()
    expect(screen.getByText(/Iraq's premier financial intelligence platform/)).toBeInTheDocument()

    // Check support information
    expect(screen.getByText('Technical Support')).toBeInTheDocument()
    expect(screen.getByText(/8 AM - 6 PM \(Baghdad Time\)/)).toBeInTheDocument()
    expect(screen.getByText('Security & Compliance')).toBeInTheDocument()
    expect(screen.getByText(/AES-256 encryption/)).toBeInTheDocument()
  })

  it('shows progress bar during activation', async () => {
    const user = userEvent.setup()

    // Mock activation with delay
    mockApiClient.activateLicense.mockImplementation(
      () => new Promise(resolve => setTimeout(() => resolve({ success: true }), 200))
    )

    const mockUseApi = require('@/lib/hooks/use-api').useApi
    mockUseApi.mockReturnValue({
      execute: mockApiClient.activateLicense,
      loading: true,
      error: null,
    })

    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')

    // Progress bar should be visible during activation
    expect(screen.getByText('Activating license...')).toBeInTheDocument()
    expect(screen.getByRole('progressbar')).toBeInTheDocument()
  })

  it('validates organization field when provided', async () => {
    const user = userEvent.setup()
    render(<LicensePage />)

    const organizationInput = screen.getByLabelText('Organization (Optional)')

    // Test empty organization (should be valid)
    await user.clear(organizationInput)
    expect(screen.queryByText(/organization/i)).not.toBeInTheDocument()

    // Test valid organization
    await user.type(organizationInput, 'Iraqi Investment Bank')
    expect(screen.queryByText(/Invalid organization/)).not.toBeInTheDocument()
  })

  it('clears form after successful activation', async () => {
    const user = userEvent.setup()

    mockApiClient.activateLicense.mockResolvedValue({
      success: true,
      message: 'License activated successfully',
    })

    render(<LicensePage />)

    const licenseKeyInput = screen.getByLabelText('License Key *')
    const organizationInput = screen.getByLabelText('Organization (Optional)')

    await user.type(licenseKeyInput, 'ISX1Y-ABCDE-12345-FGHIJ-67890')
    await user.type(organizationInput, 'Test Organization')

    // Submit form
    await user.click(screen.getByRole('button', { name: 'Activate License' }))

    await waitFor(() => {
      expect(licenseKeyInput).toHaveValue('')
      expect(organizationInput).toHaveValue('')
    })
  })
})