/**
 * Pipeline Connector Component
 *
 * Visual connector between pipeline stages with accessibility support.
 * Shows different visual states based on completion status.
 */

'use client'

import React from 'react'
import { cn } from '@/lib/utils'

interface PipelineConnectorProps {
  isActive: boolean
  isComplete: boolean
  className?: string
}

export const PipelineConnector: React.FC<PipelineConnectorProps> = ({
  isActive,
  isComplete,
  className
}) => {
  const connectorClass = cn(
    "w-0.5 h-8 rounded-full transition-all duration-500",
    isComplete
      ? "bg-gradient-to-b from-green-400 to-blue-400"
      : isActive
        ? "bg-gradient-to-b from-blue-300 to-blue-400"
        : "bg-gray-300",
    className
  )

  const getAriaLabel = () => {
    if (isComplete) return "Stage connector: completed"
    if (isActive) return "Stage connector: active"
    return "Stage connector: pending"
  }

  return (
    <div className="flex justify-center py-2">
      <div
        className={connectorClass}
        role="separator"
        aria-label={getAriaLabel()}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={isComplete ? 100 : isActive ? 50 : 0}
        aria-valuetext={getAriaLabel()}
      />
    </div>
  )
}

export default PipelineConnector