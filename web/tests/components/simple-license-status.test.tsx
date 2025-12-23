/**
 * Simple license status tests
 */

import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import { apiClient } from '@/lib/api'
import type { LicenseApiResponse } from '@/types/index'

// Mock the API client
jest.mock('@/lib/api', () => ({
  apiClient: {
    getLicenseStatus: jest.fn(),
  },
}))

const mockApiClient = apiClient as jest.Mocked<typeof apiClient>

// Mock component that uses the simple license logic
function TestLicenseDisplay() {
  const [response, setResponse] = React.useState<LicenseApiResponse | null>(null)

  React.useEffect(() => {
    const fetchStatus = async () => {
      try {
        const data = await apiClient.getLicenseStatus()
        setResponse(data)
      } catch (error) {
        setResponse(null)
      }
    }
    fetchStatus()
  }, [])

  // Simple license logic (same as in app-content.tsx)
  const isLicensed = response?.license_status === 'active'
  const daysLeft = response?.days_left || 0
  const displayText = isLicensed 
    ? `Licensed (${daysLeft} days remaining)` 
    : 'Unlicensed'

  return <div data-testid="license-status">{displayText}</div>
}

describe('Simple License Status', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('shows licensed when status is active', async () => {
    const mockResponse: LicenseApiResponse = {
      status: 200,
      license_status: 'active',
      message: 'License is active',
      days_left: 30,
      trace_id: 'test-trace-id',
      timestamp: '2025-07-30T12:00:00Z'
    }

    mockApiClient.getLicenseStatus.mockResolvedValue(mockResponse)

    render(<TestLicenseDisplay />)

    await waitFor(() => {
      expect(screen.getByTestId('license-status')).toHaveTextContent('Licensed (30 days remaining)')
    })
  })

  test('shows unlicensed when status is not active', async () => {
    const mockResponse: LicenseApiResponse = {
      status: 200,
      license_status: 'expired',
      message: 'License has expired',
      days_left: 0,
      trace_id: 'test-trace-id',
      timestamp: '2025-07-30T12:00:00Z'
    }

    mockApiClient.getLicenseStatus.mockResolvedValue(mockResponse)

    render(<TestLicenseDisplay />)

    await waitFor(() => {
      expect(screen.getByTestId('license-status')).toHaveTextContent('Unlicensed')
    })
  })

  test('shows unlicensed when API call fails', async () => {
    mockApiClient.getLicenseStatus.mockRejectedValue(new Error('API Error'))

    render(<TestLicenseDisplay />)

    await waitFor(() => {
      expect(screen.getByTestId('license-status')).toHaveTextContent('Unlicensed')
    })
  })

  test('handles missing days_left gracefully', async () => {
    const mockResponse: LicenseApiResponse = {
      status: 200,
      license_status: 'active',
      message: 'License is active',
      trace_id: 'test-trace-id',
      timestamp: '2025-07-30T12:00:00Z'
      // days_left is undefined
    }

    mockApiClient.getLicenseStatus.mockResolvedValue(mockResponse)

    render(<TestLicenseDisplay />)

    await waitFor(() => {
      expect(screen.getByTestId('license-status')).toHaveTextContent('Licensed (0 days remaining)')
    })
  })
})