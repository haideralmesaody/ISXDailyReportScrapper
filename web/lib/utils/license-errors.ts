/**
 * License Error Types and Utilities
 * Provides structured error handling for license activation
 */

export enum LicenseErrorType {
  ALREADY_ACTIVATED = 'already_activated',
  REACTIVATION_SUCCESS = 'reactivation_success',
  REACTIVATION_LIMIT_EXCEEDED = 'reactivation_limit_exceeded',
  ALREADY_ACTIVATED_DIFFERENT_DEVICE = 'already_activated_different_device',
  INVALID_FORMAT = 'invalid_format',
  EXPIRED = 'expired',
  NOT_FOUND = 'not_found',
  NETWORK_ERROR = 'network_error',
  RATE_LIMITED = 'rate_limited',
  BLACKLISTED = 'blacklisted',
  SERVER_ERROR = 'server_error',
  UNKNOWN = 'unknown'
}

export interface LicenseErrorDetails {
  type: LicenseErrorType
  title: string
  message: string
  details?: {
    activationDate?: string
    expiryDate?: string
    deviceInfo?: string
    supportEmail?: string
    canRecover?: boolean
    reactivationCount?: number
    reactivationLimit?: number
    similarityScore?: number
    remainingAttempts?: number
  }
  actions?: LicenseErrorAction[]
}

export interface LicenseErrorAction {
  label: string
  action: 'contact_support' | 'try_again' | 'check_format' | 'recover_license'
  href?: string
}

/**
 * Parse error response and return structured error details
 */
export function parseLicenseError(error: any): LicenseErrorDetails {
  // Extract error message from various formats
  const errorMessage = error?.detail || error?.message || error?.error || 'Unknown error'
  const errorLower = errorMessage.toLowerCase()
  
  // Check for reactivation success first
  if (errorLower.includes('reactivated') && (errorLower.includes('success') || errorLower.includes('successfully'))) {
    return {
      type: LicenseErrorType.REACTIVATION_SUCCESS,
      title: 'License Reactivated Successfully',
      message: 'Your license has been reactivated on this device.',
      details: {
        reactivationCount: error?.reactivation_count || undefined,
        similarityScore: error?.similarity_score || undefined,
        supportEmail: 'support@isxpulse.com'
      },
      actions: []
    }
  }
  
  // Check for reactivation limit exceeded
  if (errorLower.includes('reactivation') && (errorLower.includes('limit') || errorLower.includes('exceeded'))) {
    return {
      type: LicenseErrorType.REACTIVATION_LIMIT_EXCEEDED,
      title: 'Reactivation Limit Exceeded',
      message: 'This license has reached its maximum number of reactivations.',
      details: {
        reactivationCount: error?.reactivation_count || undefined,
        reactivationLimit: error?.reactivation_limit || undefined,
        supportEmail: 'support@isxpulse.com'
      },
      actions: [
        {
          label: 'Contact Support for Transfer',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com?subject=License Reactivation Limit Exceeded'
        }
      ]
    }
  }
  
  // Check for already activated on different device
  if ((errorLower.includes('already activated') || errorLower.includes('already_activated')) && 
      (errorLower.includes('different') || errorLower.includes('another'))) {
    return {
      type: LicenseErrorType.ALREADY_ACTIVATED_DIFFERENT_DEVICE,
      title: 'License Activated on Different Device',
      message: 'This license is currently active on a different device.',
      details: {
        similarityScore: error?.similarity_score || undefined,
        remainingAttempts: error?.remaining_attempts || undefined,
        supportEmail: 'support@isxpulse.com',
        canRecover: false
      },
      actions: [
        {
          label: 'Contact Support for Transfer',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com?subject=License Device Transfer Request'
        }
      ]
    }
  }
  
  // Check for general already activated
  if (errorLower.includes('already activated') || errorLower.includes('already_activated')) {
    return {
      type: LicenseErrorType.ALREADY_ACTIVATED,
      title: 'License Already Activated',
      message: 'This license has already been activated on another device.',
      details: {
        supportEmail: 'support@isxpulse.com',
        canRecover: false
      },
      actions: [
        {
          label: 'Contact Support for Transfer',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com?subject=License Transfer Request'
        }
      ]
    }
  }
  
  if (errorLower.includes('invalid') && errorLower.includes('format')) {
    return {
      type: LicenseErrorType.INVALID_FORMAT,
      title: 'Invalid License Format',
      message: 'The license key format is incorrect.',
      details: {
        supportEmail: 'support@isxpulse.com'
      },
      actions: [
        {
          label: 'Check License Format',
          action: 'check_format'
        },
        {
          label: 'Contact Support',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com'
        }
      ]
    }
  }
  
  if (errorLower.includes('expired')) {
    return {
      type: LicenseErrorType.EXPIRED,
      title: 'License Expired',
      message: 'This license has expired and needs to be renewed.',
      details: {
        supportEmail: 'support@isxpulse.com'
      },
      actions: [
        {
          label: 'Purchase New License',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com?subject=License Renewal'
        }
      ]
    }
  }
  
  if (errorLower.includes('not found') || errorLower.includes('invalid license key')) {
    return {
      type: LicenseErrorType.NOT_FOUND,
      title: 'License Not Found',
      message: 'This license key was not found in our system.',
      details: {
        supportEmail: 'support@isxpulse.com'
      },
      actions: [
        {
          label: 'Verify License Key',
          action: 'check_format'
        },
        {
          label: 'Contact Support',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com'
        }
      ]
    }
  }
  
  if (errorLower.includes('network') || errorLower.includes('connection') || errorLower.includes('timeout')) {
    return {
      type: LicenseErrorType.NETWORK_ERROR,
      title: 'Connection Error',
      message: 'Unable to connect to the license server. Please check your internet connection.',
      actions: [
        {
          label: 'Try Again',
          action: 'try_again'
        }
      ]
    }
  }
  
  if (errorLower.includes('rate limit') || errorLower.includes('too many')) {
    return {
      type: LicenseErrorType.RATE_LIMITED,
      title: 'Too Many Attempts',
      message: 'You have made too many activation attempts. Please wait before trying again.',
      actions: [
        {
          label: 'Try Again Later',
          action: 'try_again'
        }
      ]
    }
  }
  
  if (errorLower.includes('blacklist') || errorLower.includes('access denied')) {
    return {
      type: LicenseErrorType.BLACKLISTED,
      title: 'Access Denied',
      message: 'Your access has been restricted. Please contact support.',
      details: {
        supportEmail: 'support@isxpulse.com'
      },
      actions: [
        {
          label: 'Contact Support',
          action: 'contact_support',
          href: 'mailto:support@isxpulse.com?subject=Access Restricted'
        }
      ]
    }
  }
  
  // Default to unknown error
  return {
    type: LicenseErrorType.UNKNOWN,
    title: 'Activation Failed',
    message: errorMessage,
    details: {
      supportEmail: 'support@isxpulse.com'
    },
    actions: [
      {
        label: 'Try Again',
        action: 'try_again'
      },
      {
        label: 'Contact Support',
        action: 'contact_support',
        href: 'mailto:support@isxpulse.com'
      }
    ]
  }
}

/**
 * Get icon name for error type
 */
export function getErrorIcon(type: LicenseErrorType): string {
  switch (type) {
    case LicenseErrorType.REACTIVATION_SUCCESS:
      return 'RotateCcw'
    case LicenseErrorType.REACTIVATION_LIMIT_EXCEEDED:
      return 'XCircle'
    case LicenseErrorType.ALREADY_ACTIVATED_DIFFERENT_DEVICE:
      return 'Smartphone'
    case LicenseErrorType.ALREADY_ACTIVATED:
      return 'UserCheck'
    case LicenseErrorType.INVALID_FORMAT:
      return 'AlertCircle'
    case LicenseErrorType.EXPIRED:
      return 'Clock'
    case LicenseErrorType.NOT_FOUND:
      return 'Search'
    case LicenseErrorType.NETWORK_ERROR:
      return 'WifiOff'
    case LicenseErrorType.RATE_LIMITED:
      return 'Timer'
    case LicenseErrorType.BLACKLISTED:
      return 'Ban'
    default:
      return 'AlertTriangle'
  }
}

/**
 * Get error color for styling
 */
export function getErrorColor(type: LicenseErrorType): string {
  switch (type) {
    case LicenseErrorType.REACTIVATION_SUCCESS:
      return 'green'
    case LicenseErrorType.REACTIVATION_LIMIT_EXCEEDED:
      return 'red'
    case LicenseErrorType.ALREADY_ACTIVATED_DIFFERENT_DEVICE:
      return 'orange'
    case LicenseErrorType.ALREADY_ACTIVATED:
      return 'orange'
    case LicenseErrorType.EXPIRED:
      return 'yellow'
    case LicenseErrorType.NETWORK_ERROR:
      return 'blue'
    case LicenseErrorType.BLACKLISTED:
    case LicenseErrorType.NOT_FOUND:
      return 'red'
    default:
      return 'red'
  }
}