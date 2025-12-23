/**
 * Error Boundary component for ISX Daily Reports Scrapper
 * Professional error handling with user-friendly fallbacks
 */

'use client'

import React, { ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw, Home, Bug } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
  showErrorDetails?: boolean
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    }
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
    }
  }

  override componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error Boundary caught an error:', error, errorInfo)
    
    this.setState({
      error,
      errorInfo,
    })

    // Call optional error handler
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    }

    // Report error to monitoring service in production
    if (process.env.NODE_ENV === 'production') {
      // Add error reporting service integration here
      // e.g., Sentry, LogRocket, etc.
    }
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    })
  }

  handleReload = () => {
    window.location.reload()
  }

  handleGoHome = () => {
    window.location.href = '/'
  }

  override render() {
    if (this.state.hasError) {
      // Use custom fallback if provided
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Default error UI
      return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
          <Card className="max-w-2xl w-full">
            <CardHeader className="text-center">
              <div className="flex justify-center mb-4">
                <AlertTriangle className="h-16 w-16 text-destructive" />
              </div>
              <CardTitle className="text-2xl">Something went wrong</CardTitle>
              <CardDescription className="text-base">
                We apologize for the inconvenience. An unexpected error occurred while loading this part of the application.
              </CardDescription>
            </CardHeader>
            
            <CardContent className="space-y-6">
              {/* Error Details */}
              {this.props.showErrorDetails && this.state.error && (
                <Alert variant="destructive">
                  <Bug className="h-4 w-4" />
                  <AlertDescription>
                    <div className="space-y-2">
                      <div>
                        <strong>Error:</strong> {this.state.error.message}
                      </div>
                      {this.state.errorInfo && (
                        <details className="mt-2">
                          <summary className="cursor-pointer text-sm font-medium">
                            Technical Details (Click to expand)
                          </summary>
                          <div className="mt-2 p-2 bg-muted rounded text-xs font-mono whitespace-pre-wrap">
                            {this.state.error.stack}
                            {'\n\nComponent Stack:'}
                            {this.state.errorInfo.componentStack}
                          </div>
                        </details>
                      )}
                    </div>
                  </AlertDescription>
                </Alert>
              )}

              {/* Recovery Actions */}
              <div className="flex flex-col sm:flex-row gap-3 justify-center">
                <Button onClick={this.handleRetry} className="flex items-center space-x-2">
                  <RefreshCw className="h-4 w-4" />
                  <span>Try Again</span>
                </Button>
                
                <Button 
                  variant="outline" 
                  onClick={this.handleReload} 
                  className="flex items-center space-x-2"
                >
                  <RefreshCw className="h-4 w-4" />
                  <span>Reload Page</span>
                </Button>
                
                <Button 
                  variant="secondary" 
                  onClick={this.handleGoHome} 
                  className="flex items-center space-x-2"
                >
                  <Home className="h-4 w-4" />
                  <span>Go to Dashboard</span>
                </Button>
              </div>

              {/* Support Information */}
              <div className="text-center text-sm text-muted-foreground space-y-2">
                <p>
                  If this problem persists, please contact technical support.
                </p>
                <p>
                  <strong>Error ID:</strong> {Date.now().toString(36)}
                </p>
                {process.env.NODE_ENV === 'development' && (
                  <p className="text-xs">
                    <strong>Development Mode:</strong> Additional error details are shown above.
                  </p>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}

// Functional error boundary for specific use cases
interface ErrorFallbackProps {
  error: Error
  resetError: () => void
  title?: string
  description?: string
  showDetails?: boolean
}

export function ErrorFallback({ 
  error, 
  resetError, 
  title = "Something went wrong",
  description = "An unexpected error occurred. Please try again.",
  showDetails = false 
}: ErrorFallbackProps) {
  return (
    <div className="flex items-center justify-center min-h-[200px] p-4">
      <Card className="max-w-md w-full">
        <CardHeader className="text-center">
          <div className="flex justify-center mb-2">
            <AlertTriangle className="h-8 w-8 text-destructive" />
          </div>
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        
        <CardContent className="space-y-4">
          {showDetails && (
            <Alert variant="destructive">
              <AlertDescription className="text-sm font-mono">
                {error.message}
              </AlertDescription>
            </Alert>
          )}
          
          <div className="flex justify-center">
            <Button onClick={resetError} className="flex items-center space-x-2">
              <RefreshCw className="h-4 w-4" />
              <span>Try Again</span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

// Hook for functional components error handling
export function useErrorHandler() {
  const [error, setError] = React.useState<Error | null>(null)

  const resetError = React.useCallback(() => {
    setError(null)
  }, [])

  const handleError = React.useCallback((error: Error) => {
    console.error('Error caught by useErrorHandler:', error)
    setError(error)
  }, [])

  // Reset error when component unmounts
  React.useEffect(() => {
    return () => setError(null)
  }, [])

  return {
    error,
    resetError,
    handleError,
    hasError: error !== null,
  }
}

// Component-specific error boundaries
export function ComponentErrorBoundary({ 
  children, 
  componentName = "Component" 
}: { 
  children: ReactNode
  componentName?: string 
}) {
  return (
    <ErrorBoundary
      fallback={
        <ErrorFallback
          error={new Error(`${componentName} failed to render`)}
          resetError={() => window.location.reload()}
          title={`${componentName} Error`}
          description={`There was a problem loading the ${componentName.toLowerCase()}. Please try refreshing the page.`}
        />
      }
      showErrorDetails={process.env.NODE_ENV === 'development'}
    >
      {children}
    </ErrorBoundary>
  )
}

// API-specific error boundary
export function ApiErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary
      fallback={
        <ErrorFallback
          error={new Error('API request failed')}
          resetError={() => window.location.reload()}
          title="Connection Error"
          description="Unable to connect to the server. Please check your connection and try again."
        />
      }
      onError={(error, errorInfo) => {
        // Log API errors specifically
        console.error('API Error:', error, errorInfo)
      }}
    >
      {children}
    </ErrorBoundary>
  )
}

// Page-level error boundary
export function PageErrorBoundary({ 
  children, 
  pageName = "Page" 
}: { 
  children: ReactNode
  pageName?: string 
}) {
  return (
    <ErrorBoundary
      showErrorDetails={process.env.NODE_ENV === 'development'}
      onError={(error, errorInfo) => {
        console.error(`${pageName} Error:`, error, errorInfo)
      }}
    >
      {children}
    </ErrorBoundary>
  )
}

export default ErrorBoundary