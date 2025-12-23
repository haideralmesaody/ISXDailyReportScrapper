/**
 * Safe Component Wrappers
 * Prevents TDZ errors during minification by ensuring components are properly initialized
 */

'use client'

import React from 'react'
import { Button as UIButton } from './button'
import { Separator as UISeparator } from './separator'
import { Card as UICard } from './card'
import { Alert as UIAlert, AlertDescription as UIAlertDescription } from './alert'

// Safe wrappers that prevent TDZ during minification
export const SafeButton = React.forwardRef<any, any>((props, ref) => {
  return React.createElement(UIButton, { ...props, ref })
})
SafeButton.displayName = 'SafeButton'

export const SafeSeparator = React.forwardRef<any, any>((props, ref) => {
  return React.createElement(UISeparator, { ...props, ref })
})
SafeSeparator.displayName = 'SafeSeparator'

export const SafeCard = React.forwardRef<any, any>((props, ref) => {
  return React.createElement(UICard, { ...props, ref })
})
SafeCard.displayName = 'SafeCard'

export const SafeAlert = React.forwardRef<any, any>((props, ref) => {
  return React.createElement(UIAlert, { ...props, ref })
})
SafeAlert.displayName = 'SafeAlert'

export const SafeAlertDescription = React.forwardRef<any, any>((props, ref) => {
  return React.createElement(UIAlertDescription, { ...props, ref })
})
SafeAlertDescription.displayName = 'SafeAlertDescription'

// Re-export with safe names to prevent minification conflicts
export {
  SafeButton as Button,
  SafeSeparator as Separator,
  SafeCard as Card,
  SafeAlert as Alert,
  SafeAlertDescription as AlertDescription
}

// Legacy exports for compatibility
export {
  SafeButton as ButtonComponent,
  SafeSeparator as SeparatorComponent,
  SafeCard as CardComponent,
  SafeAlert as AlertComponent,
  SafeAlertDescription as AlertDescriptionComponent
}