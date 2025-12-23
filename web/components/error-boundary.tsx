'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCw, Home } from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { InvestorLogo } from '@/components/layout/investor-logo'

interface Props {
  children: ReactNode
  fallback?: ReactNode
  onError?: (error: Error, errorInfo: ErrorInfo) => void
}

interface State {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public override componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
    
    this.setState({
      error,
      errorInfo
    })

    // Call optional error callback
    if (this.props.onError) {
      this.props.onError(error, errorInfo)
    }

    // Report to error tracking service if available
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('event', 'exception', {
        description: error.toString(),
        fatal: false,
      })
    }
  }

  private handleRetry = () => {
    this.setState({ hasError: false })
  }

  private handleGoHome = () => {
    window.location.href = '/'
  }

  private handleReload = () => {
    window.location.reload()
  }

  public override render(): ReactNode {
    if (this.state.hasError) {
      // Use custom fallback if provided
      if (this.props.fallback) {
        return this.props.fallback
      }

      // Default error UI
      return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4">
          <Card className="w-full max-w-2xl">
            <CardHeader className="text-center space-y-4">
              <div className="flex justify-center">
                <InvestorLogo size="lg" />
              </div>
              <div className="space-y-2">
                <CardTitle className="text-xl text-destructive flex items-center justify-center gap-2">
                  <AlertTriangle className="h-5 w-5" />
                  Something went wrong
                </CardTitle>
                <CardDescription>
                  We&apos;re sorry, but an unexpected error occurred while running the ISX Pulse application.
                </CardDescription>
              </div>
            </CardHeader>
            
            <CardContent className="space-y-6">
              <Alert variant="destructive">
                <AlertTriangle className="h-4 w-4" />
                <AlertDescription>
                  <strong>Error:</strong> {this.state.error?.message || 'Unknown error occurred'}
                </AlertDescription>
              </Alert>

              {/* Error details for development */}
              {process.env.NODE_ENV === 'development' && this.state.error && (
                <details className="mt-4">
                  <summary className="cursor-pointer text-sm font-medium text-muted-foreground hover:text-foreground">
                    Error Details (Development Mode)
                  </summary>
                  <div className="mt-2 p-3 bg-muted rounded-md">
                    <pre className="text-xs whitespace-pre-wrap overflow-auto max-h-40">
                      {this.state.error.stack}
                    </pre>
                    {this.state.errorInfo && (
                      <div className="mt-2">
                        <strong className="text-xs">Component Stack:</strong>
                        <pre className="text-xs whitespace-pre-wrap">
                          {this.state.errorInfo.componentStack}
                        </pre>
                      </div>
                    )}
                  </div>
                </details>
              )}

              {/* Action buttons */}
              <div className="flex flex-col sm:flex-row gap-3">
                <Button 
                  onClick={this.handleRetry} 
                  className="flex-1"
                  variant="default"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Try Again
                </Button>
                
                <Button 
                  onClick={this.handleReload} 
                  className="flex-1"
                  variant="outline"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Reload Page
                </Button>
                
                <Button 
                  onClick={this.handleGoHome} 
                  className="flex-1"
                  variant="outline"
                >
                  <Home className="h-4 w-4 mr-2" />
                  Go Home
                </Button>
              </div>

              {/* Support information */}
              <div className="pt-4 border-t">
                <h4 className="font-medium mb-2">Need Help?</h4>
                <div className="text-sm text-muted-foreground space-y-1">
                  <p>If this problem persists, please contact support:</p>
                  <p>Email: <a href="mailto:support@isxpulse.com" className="text-primary hover:underline">support@isxpulse.com</a></p>
                  <p>Phone: +964 (0) 790 123 4567</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )
    }

    return this.props.children
  }
}

// Hook version for functional components
export function useErrorHandler() {
  return (error: Error, errorInfo?: ErrorInfo) => {
    console.error('Error caught by useErrorHandler:', error, errorInfo)
    
    // Report to error tracking service
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('event', 'exception', {
        description: error.toString(),
        fatal: false,
      })
    }
  }
}

// Simplified error boundary for specific use cases
interface SimpleErrorBoundaryProps {
  children: ReactNode
  fallback?: (error: Error) => ReactNode
}

export function SimpleErrorBoundary({ children, fallback }: SimpleErrorBoundaryProps) {
  return (
    <ErrorBoundary
      fallback={
        fallback ? undefined : (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              Something went wrong. Please try refreshing the page.
            </AlertDescription>
          </Alert>
        )
      }
      onError={(error) => {
        if (fallback) {
          fallback(error)
        }
      }}
    >
      {children}
    </ErrorBoundary>
  )
}