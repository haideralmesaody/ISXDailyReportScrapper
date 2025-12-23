'use client'

import { useEffect, useState } from 'react'
import { api } from '@/lib/api'

interface VersionInfo {
  version: string
  build_time?: string
  build_id?: string
  go_version: string
  os: string
  arch: string
  current_time: string
}

export function VersionInfo() {
  const [versionInfo, setVersionInfo] = useState<VersionInfo | null>(null)
  const [isHydrated, setIsHydrated] = useState(false)
  const [frontendBuildTime] = useState(process.env.NEXT_PUBLIC_BUILD_TIME || 'unknown')
  const [frontendBuildId] = useState(process.env.NEXT_PUBLIC_BUILD_ID || 'unknown')

  // Set hydration state
  useEffect(() => {
    setIsHydrated(true)
  }, [])

  useEffect(() => {
    // Skip version fetch until after hydration
    if (!isHydrated) return

    const fetchVersion = async () => {
      try {
        const response = await api.getVersion()
        setVersionInfo(response)
        
        // Check for version mismatch
        if (response.build_id && frontendBuildId !== 'unknown' && response.build_id !== frontendBuildId) {
          console.warn('Build ID mismatch detected:', {
            backend: response.build_id,
            frontend: frontendBuildId
          })
          
          // Force reload to get latest version
          if (typeof window !== 'undefined' && window.location) {
            console.log('Forcing reload due to version mismatch...')
            window.location.reload()
          }
        }
      } catch (error) {
        console.error('Failed to fetch version info:', error)
      }
    }

    fetchVersion()
  }, [isHydrated, frontendBuildId])

  // Don't render until hydrated and version info is loaded
  if (!isHydrated || !versionInfo) return null

  return (
    <div className="fixed bottom-0 right-0 p-2 text-xs text-gray-500 bg-white/80 rounded-tl-md shadow-sm">
      <div>Backend: v{versionInfo.version} ({versionInfo.build_id?.substring(0, 8) || 'dev'})</div>
      <div>Frontend: {frontendBuildId.substring(0, 8)}</div>
    </div>
  )
}