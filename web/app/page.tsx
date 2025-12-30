/**
 * Root Page - Smart Entry Point with License-Aware Routing
 * - Active license → redirects to Market Overview (/dashboard)
 * - Inactive license → redirects to License activation (/license)
 */

'use client'

import { useEffect } from 'react'
import { InvestorHeaderLogo } from '@/components/layout/investor-logo'
import { Loader2 } from 'lucide-react'
import { apiClient } from '@/lib/api'

// Add global debug function for browser console
declare global {
  interface Window {
    debugLicense?: () => Promise<void>
  }
}

export default function HomePage() {
  useEffect(() => {
    // Register browser-only debug helper
    if (typeof window !== 'undefined') {
      window.debugLicense = async () => {
        console.group('🔍 License Debug Information')
        try {
          const response = await fetch('/api/license/debug')
          if (response.ok) {
            const debug = await response.json()
            console.table('License Debug Info:', debug)

            console.log('📁 License File Info:', {
              file_path: debug.file_path,
              file_exists: debug.file_exists,
              is_readable: debug.is_readable,
              file_size: debug.file_size,
              file_mod_time: debug.file_mod_time
            })

            console.log('💾 Cache Info:', {
              cached_status: debug.cached_status,
              last_validation: debug.last_validation,
              server_time: debug.server_time
            })

            console.log('🔄 Refreshing license status...')
            const status = await apiClient.refreshLicenseStatus()
            console.log('✅ Fresh License Status:', status)
          } else {
            console.error('❌ Failed to fetch debug info:', response.statusText)
          }
        } catch (error) {
          console.error('❌ Debug error:', error)
        }
        console.groupEnd()
      }
    }

    const checkLicenseAndRedirect = async () => {
      console.log('[ROOT PAGE] Starting license check...')
      // Get deep link intent if any (must be available for retry/fallbacks)
      const params = new URLSearchParams(window.location.search)
      const returnTo = params.get('returnTo')
      console.log('[ROOT PAGE] Return path:', returnTo || 'none')

      try {
        // Check license status
        console.log('[ROOT PAGE] Calling apiClient.getLicenseStatus()...')
        const status = await apiClient.getLicenseStatus()
        console.log('[ROOT PAGE] API response received:', JSON.stringify(status, null, 2))

        // Log the specific field we're checking
        console.log('[ROOT PAGE] license_status field:', status?.license_status)
        console.log('[ROOT PAGE] Type of license_status:', typeof status?.license_status)

        // Additional debug info
        console.log('[ROOT PAGE] Full status object keys:', Object.keys(status || {}))
        console.log('[ROOT PAGE] Response timestamp:', new Date().toISOString())

        const isActive = status?.license_status === 'active' || status?.license_status === 'warning'
        console.log('[ROOT PAGE] Is license active?', isActive)

        if (isActive) {
          // License is active → go to Operations
          const operationsUrl = returnTo
            ? `/operations?returnTo=${encodeURIComponent(returnTo)}`
            : '/operations'
          console.log('[ROOT PAGE] Redirecting to operations:', operationsUrl)
          window.location.href = operationsUrl
        } else {
          // License needs activation → go to License page
          const licenseUrl = returnTo
            ? `/license?returnTo=${encodeURIComponent(returnTo)}`
            : '/license'
          console.log('[ROOT PAGE] Redirecting to license page:', licenseUrl, 'because status is:', status?.license_status)
          window.location.href = licenseUrl
        }
      } catch (error) {
        // On error, retry once before falling back to license page
        console.error('[ROOT PAGE] License check failed with error:', error)
        console.error('[ROOT PAGE] Error details:', {
          name: error instanceof Error ? error.name : 'Unknown',
          message: error instanceof Error ? error.message : String(error),
          stack: error instanceof Error ? error.stack : 'No stack'
        })

        // Retry once after a short delay
        console.log('[ROOT PAGE] Retrying license check in 2 seconds...')
        setTimeout(async () => {
          try {
            console.log('[ROOT PAGE] Retry attempt - calling apiClient.refreshLicenseStatus() (forced refresh)...')
            const retryStatus = await apiClient.refreshLicenseStatus()
            console.log('[ROOT PAGE] Retry API response:', JSON.stringify(retryStatus, null, 2))

            const isRetryActive = retryStatus?.license_status === 'active' || retryStatus?.license_status === 'warning'
            console.log('[ROOT PAGE] Retry - Is license active?', isRetryActive)

            if (isRetryActive) {
              const operationsUrl = returnTo
                ? `/operations?returnTo=${encodeURIComponent(returnTo)}`
                : '/operations'
              console.log('[ROOT PAGE] Retry successful - redirecting to operations:', operationsUrl)
              window.location.href = operationsUrl
            } else {
              console.log('[ROOT PAGE] Retry failed - redirecting to license page')
              window.location.href = '/license'
            }
          } catch (retryError) {
            console.error('[ROOT PAGE] Retry also failed:', retryError)
            console.log('[ROOT PAGE] Final fallback to license page')
            window.location.href = '/license'
          }
        }, 2000)
      }
    }

    checkLicenseAndRedirect()
  }, [])

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-b from-background to-background/95">
      <div className="text-center space-y-6">
        <InvestorHeaderLogo size="xl" />
        <div className="flex items-center gap-3">
          <Loader2 className="h-5 w-5 animate-spin text-primary" />
          <p className="text-muted-foreground">Preparing your workspace...</p>
        </div>
      </div>
    </div>
  )
}
