/**
 * License Guard Hook
 * Client-side navigation protection based on license status
 */

'use client'

import { useRouter } from 'next/navigation'
import { useToast } from '@/lib/hooks/use-toast'
import { useApi } from '@/lib/hooks/use-api'
import { apiClient } from '@/lib/api'
import type { LicenseApiResponse } from '@/types/index'

/**
 * Hook to check license status and guard navigation
 * @returns Functions to check access and navigate with license validation
 */
export function useLicenseGuard() {
  const router = useRouter()
  const { toast } = useToast()

  // Fetch current license status
  const { data: licenseStatus, isLoading } = useApi<LicenseApiResponse>(
    'license-status-guard',
    () => apiClient.getLicenseStatus(),
    {
      refetchInterval: 300000, // Refresh every 5 minutes
      retry: 2
    }
  )

  /**
   * Check if user can access a specific path based on license status
   * @param path - The path to check (e.g., '/operations', '/dashboard')
   * @returns true if access is allowed, false otherwise
   */
  const canAccess = (path: string): boolean => {
    // Always allow access to license page and root
    if (path === '/' || path === '/license' || path.startsWith('/license')) {
      return true
    }

    // If still loading, allow access (will be checked by page component)
    if (isLoading) {
      return true
    }

    // Check if license is active or in warning state (expiring soon)
    const status = licenseStatus?.license_status?.toLowerCase()
    return status === 'active' || status === 'warning'
  }

  /**
   * Navigate to a path with license check
   * If license is invalid, show toast and redirect to license page
   * @param path - Destination path
   */
  const navigateWithGuard = (path: string) => {
    if (canAccess(path)) {
      router.push(path)
    } else {
      toast({
        title: 'License Required',
        description: 'Please activate your license to access this feature.',
        variant: 'destructive'
      })
      router.push('/license')
    }
  }

  /**
   * Check if license is currently valid
   */
  const isLicenseValid = (): boolean => {
    const status = licenseStatus?.license_status?.toLowerCase()
    return status === 'active' || status === 'warning'
  }

  /**
   * Get current license status string
   */
  const getLicenseStatus = (): string | undefined => {
    return licenseStatus?.license_status
  }

  return {
    canAccess,
    navigateWithGuard,
    isLicenseValid,
    getLicenseStatus,
    isLoading,
    licenseData: licenseStatus
  }
}
