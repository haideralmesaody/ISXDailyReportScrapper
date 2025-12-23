/**
 * License Status Message Utilities
 * Provides user-friendly messages for different license states
 */

export const getLicenseMessage = (status: string, message?: string): string => {
  switch (status) {
    case 'expired':
      return 'Your license has expired. Please renew your license to continue using the application.'
    case 'not_activated':
      return 'No license activated. Please enter your license key to get started.'
    case 'critical':
      return 'Your license expires soon (within 7 days). Please renew to avoid service interruption.'
    case 'warning':
      return 'Your license expires within 30 days. Consider renewing soon to avoid service interruption.'
    case 'active':
      return 'Your license is active and ready to use.'
    case 'error':
      return 'Unable to verify license status. Please check your connection and try again.'
    default:
      return message || 'Please check your license status.'
  }
}

export const getLicenseStatusDisplayText = (status: string): string => {
  switch (status) {
    case 'expired':
      return 'Expired'
    case 'not_activated':
      return 'Not Activated'
    case 'critical':
      return 'Expires Soon'
    case 'warning':
      return 'Renewal Due'
    case 'active':
      return 'Active'
    case 'error':
      return 'Error'
    default:
      return 'Unknown'
  }
}

export const getLicenseStatusVariant = (status: string): 'default' | 'destructive' | 'secondary' => {
  switch (status) {
    case 'expired':
    case 'error':
      return 'destructive'
    case 'critical':
    case 'warning':
      return 'secondary'
    case 'active':
    case 'not_activated':
    default:
      return 'default'
  }
}