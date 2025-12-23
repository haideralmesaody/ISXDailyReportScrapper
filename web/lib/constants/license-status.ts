/**
 * License status configuration constants
 * Centralizes all license status-related styling and configuration
 */

export const LICENSE_STATUS_CONFIG = {
  active: {
    color: 'green',
    borderClass: 'border-green-200',
    bgClass: 'bg-green-50',
    textClass: 'text-green-800',
    badgeBg: 'bg-green-100',
    badgeText: 'text-green-800',
    badgeBorder: 'border-green-200',
    indicatorBg: 'bg-green-500',
    indicatorText: 'text-green-600',
    icon: 'check' as const,
    title: 'License Active',
    defaultMessage: 'Your professional license is active.'
  },
  warning: {
    color: 'blue',
    borderClass: 'border-blue-200',
    bgClass: 'bg-blue-50',
    textClass: 'text-blue-800',
    badgeBg: 'bg-blue-100',
    badgeText: 'text-blue-800',
    badgeBorder: 'border-blue-200',
    indicatorBg: 'bg-blue-500',
    indicatorText: 'text-blue-600',
    icon: 'clock' as const,
    title: 'License Active - Renewal Due',
    defaultMessage: 'Your license is active but renewal is recommended soon.'
  },
  critical: {
    color: 'amber',
    borderClass: 'border-amber-200',
    bgClass: 'bg-amber-50',
    textClass: 'text-amber-800',
    badgeBg: 'bg-amber-100',
    badgeText: 'text-amber-800',
    badgeBorder: 'border-amber-200',
    indicatorBg: 'bg-amber-500',
    indicatorText: 'text-amber-600',
    icon: 'alert' as const,
    title: 'License Active - Expires Soon!',
    defaultMessage: 'Your license expires very soon. Please renew immediately.'
  },
  expired: {
    color: 'red',
    borderClass: 'border-red-200',
    bgClass: 'bg-red-50',
    textClass: 'text-red-800',
    badgeBg: 'bg-red-100',
    badgeText: 'text-red-800',
    badgeBorder: 'border-red-200',
    indicatorBg: 'bg-red-500',
    indicatorText: 'text-red-600',
    icon: 'x' as const,
    title: 'License Expired',
    defaultMessage: 'Your license has expired. Please renew to continue accessing professional features.'
  },
  invalid: {
    color: 'red',
    borderClass: 'border-red-200',
    bgClass: 'bg-red-50',
    textClass: 'text-red-800',
    badgeBg: 'bg-red-100',
    badgeText: 'text-red-800',
    badgeBorder: 'border-red-200',
    indicatorBg: 'bg-amber-500',
    indicatorText: 'text-amber-600',
    icon: 'alert' as const,
    title: 'License Activation Required',
    defaultMessage: 'Please activate your professional license to access the ISX Pulse dashboard.'
  },
  checking: {
    color: 'blue',
    borderClass: 'border-blue-200',
    bgClass: 'bg-blue-50',
    textClass: 'text-blue-800',
    badgeBg: 'bg-blue-100',
    badgeText: 'text-blue-800',
    badgeBorder: 'border-blue-200',
    indicatorBg: 'bg-blue-500',
    indicatorText: 'text-blue-600',
    icon: 'loader' as const,
    title: 'Checking License Status...',
    defaultMessage: 'Please wait while we verify your license.'
  }
} as const

// Type narrowing for better type safety
export type ActiveStatusType = 'active' | 'warning' | 'critical'
export type InactiveStatusType = 'expired' | 'invalid' | 'checking'
export type LicenseStatusType = ActiveStatusType | InactiveStatusType

// Type guard functions
export function isActiveStatus(status: LicenseStatusType): status is ActiveStatusType {
  return status === 'active' || status === 'warning' || status === 'critical'
}

export function isInactiveStatus(status: LicenseStatusType): status is InactiveStatusType {
  return status === 'expired' || status === 'invalid' || status === 'checking'
}

export function requiresRenewal(status: LicenseStatusType): boolean {
  return status === 'warning' || status === 'critical' || status === 'expired'
}

/**
 * Map backend license status to UI status type
 */
export function mapBackendStatusToUI(backendStatus: string): LicenseStatusType {
  switch (backendStatus) {
    case 'active':
      return 'active'
    case 'warning':
      return 'warning'
    case 'critical':
      return 'critical'
    case 'expired':
      return 'expired'
    case 'not_activated':
    case 'error':
      return 'invalid'
    default:
      return 'invalid'
  }
}

/**
 * Get alert variant based on license status
 */
export function getAlertVariant(status: LicenseStatusType): 'default' | 'destructive' {
  return status === 'expired' || status === 'invalid' ? 'destructive' : 'default'
}

/**
 * Get badge variant based on license status
 */
export function getBadgeVariant(status: LicenseStatusType): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'active':
      return 'default'
    case 'warning':
    case 'critical':
      return 'secondary'
    case 'expired':
    case 'invalid':
      return 'destructive'
    default:
      return 'outline'
  }
}