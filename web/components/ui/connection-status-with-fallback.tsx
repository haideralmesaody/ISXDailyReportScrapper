'use client'

import React from 'react'
import { ConnectionStatusIndicator } from './connection-status'

interface ConnectionStatusWithFallbackProps {
  showConnectionStatus?: boolean
  connectionStatus?: string
  fallbackReason?: string
  useFallback?: boolean
}

export function ConnectionStatusWithFallback({
  showConnectionStatus = true,
  fallbackReason,
  useFallback
}: ConnectionStatusWithFallbackProps) {
  if (!showConnectionStatus) {
    return null
  }

  return (
    <div className="flex items-center gap-2">
      <ConnectionStatusIndicator
        compact={true}
        showDetails={false}
      />
      {useFallback && (
        <span className="text-xs text-yellow-600 bg-yellow-50 px-2 py-1 rounded">
          Using polling ({fallbackReason})
        </span>
      )}
    </div>
  )
}

export default ConnectionStatusWithFallback