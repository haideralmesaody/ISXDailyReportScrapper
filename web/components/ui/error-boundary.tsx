/**
 * Error Boundary Component
 * Prevents crashes from bubbling up to the browser
 */

'use client'

import React, { Component, ReactNode, ErrorInfo } from 'react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { AlertTriangle, RefreshCw, Home } from 'lucide-react'

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
  errorId: string
}

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: (error: Error, errorId: string) => ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo, errorId: string) => void
  showErrorDetails?: boolean
  className?: string
}

export class SimpleErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    // Specific handling for ReferenceError Temporal Dead Zone violations
    const isTDZError = error instanceof ReferenceError &&
      (error.message.includes('before initialization') ||
       error.message.includes('Cannot access') ||
       error.message.includes("Cannot access '"))

    const errorId = isTDZError
      ? `tdz_error_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
      : `error_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`

    // For TDZ errors, create a more user-friendly error message
    const displayError = isTDZError
      ? new Error('Component initialization error - The page will refresh automatically')
      : error

    return {
      hasError: true,
      error: displayError,
      errorInfo: null,
      errorId
    }
  }

  override componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    const errorId = this.state.errorId || `error_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`

    // Detect if this is a TDZ (Temporal Dead Zone) error
    const isTDZError = error instanceof ReferenceError &&
      (error.message.includes('before initialization') ||
       error.message.includes('Cannot access') ||
       error.message.includes("Cannot access '"))

    // Enhanced logging for TDZ errors
    if (process.env.NODE_ENV === 'development') {
      console.error('Error Boundary caught an error:', {
        error,
        errorInfo,
        errorId,
        componentStack: errorInfo.componentStack,
        errorType: isTDZError ? 'TDZ_VIOLATION' : 'GENERAL_ERROR',
        timestamp: new Date().toISOString(),
        userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : 'SSR'
      })

      // Additional context for TDZ errors
      if (isTDZError) {
          console.warn('TDZ Error Details:', {
            originalError: error.message,
            likelyCause: 'Variable accessed before initialization',
            suggestedFix: 'Check variable scoping and initialization order',
            affectedComponent: (errorInfo.componentStack ?? '').split('\n')[1]?.trim()
          })
        }
      }

    // For TDZ errors, attempt automatic recovery after a short delay
    if (isTDZError) {
      console.log('TDZ error detected, attempting automatic recovery...')
      setTimeout(() => {
        this.setState({
          hasError: false,
          error: null,
          errorInfo: null,
          errorId: ''
        })
      }, 2000) // 2 second delay for recovery
    }

    this.setState({
      error,
      errorInfo,
      errorId
    })

    // Call custom error handler if provided
    this.props.onError?.(error, errorInfo, errorId)
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: ''
    })
  }

  handleGoHome = () => {
    window.location.href = '/'
  }

  override render() {
    if (this.state.hasError) {
      // Use custom fallback if provided
      if (this.props.fallback) {
        return this.props.fallback(this.state.error!, this.state.errorId)
      }

      // Default error UI
      return (
        <div className="min-h-screen p-8 flex items-center justify-center bg-gray-50 dark:bg-gray-900">
          <div className="max-w-lg w-full">
            <Alert variant="destructive" className="mb-6">
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>Application Error</AlertTitle>
              <AlertDescription>
                {this.state.error?.message || 'An unexpected error occurred while rendering this component.'}
              </AlertDescription>
            </Alert>

            <div className="bg-white dark:bg-gray-800 rounded-lg p-6 border border-gray-200 dark:border-gray-700 shadow-sm">
              <div className="text-center space-y-4">
                <div className="flex justify-center">
                  <div className="h-12 w-12 rounded-full bg-red-100 dark:bg-red-900 flex items-center justify-center">
                    <AlertTriangle className="h-6 w-6 text-red-600 dark:text-red-400" />
                  </div>
                </div>

                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                  Something went wrong
                </h3>

                <p className="text-gray-600 dark:text-gray-400">
                  {this.props.showErrorDetails && this.state.error ? (
                    <details className="text-left">
                      <summary className="cursor-pointer font-medium">
                        Error Details
                      </summary>
                      <pre className="mt-2 text-xs text-left bg-gray-100 dark:bg-gray-700 p-3 rounded overflow-auto max-h-32">
                        <code className="text-red-600 dark:text-red-400">
                          {this.state.error.stack}
                        </code>
                      </pre>
                    </details>
                  ) : (
                    'The application encountered an unexpected error and needs to be refreshed.'
                  )}
                </p>

                <p className="text-xs text-gray-500 dark:text-gray-500">
                  Reference ID: <code className="bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">
                    {this.state.errorId}
                  </code>
                </p>

                <div className="flex gap-3 justify-center">
                  <Button
                    onClick={this.handleRetry}
                    variant="outline"
                    className="flex items-center gap-2"
                  >
                    <RefreshCw className="h-4 w-4" />
                    Try Again
                  </Button>

                  <Button
                    onClick={this.handleGoHome}
                    variant="ghost"
                    className="flex items-center gap-2"
                  >
                    <Home className="h-4 w-4" />
                    Go Home
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

// Export as both named and default export for convenience
export { SimpleErrorBoundary as ErrorBoundary }
export default SimpleErrorBoundary

// HOC for wrapping components
export function withErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  options?: Omit<ErrorBoundaryProps, 'children'>
) {
  const WrappedComponent = (props: P) => (
    <SimpleErrorBoundary {...options}>
      <Component {...props} />
    </SimpleErrorBoundary>
  )

  WrappedComponent.displayName = `withErrorBoundary(${Component.displayName || Component.name})`

  return WrappedComponent
}
