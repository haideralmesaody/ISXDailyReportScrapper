/**
 * Chart Error Boundary Component
 * Following CLAUDE.md error handling best practices
 * Prevents entire page crash when chart errors occur
 */

'use client'

import React, { Component, ErrorInfo, ReactNode } from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

interface Props {
  children: ReactNode
  fallbackMessage?: string
}

interface State {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

export class ChartErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    errorInfo: null
  }

  public static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
      errorInfo: null
    }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // CLAUDE.md requirement: Structured logging with context
    console.error('[ChartErrorBoundary] Component error caught:', {
      error: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString()
    })

    this.setState({
      error,
      errorInfo
    })
  }

  private handleReset = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null
    })
  }

  public override render() {
    if (this.state.hasError) {
      return (
        <Card className="m-4 p-6">
          <div className="flex flex-col items-center justify-center text-center space-y-4">
            <AlertCircle className="h-12 w-12 text-destructive" />
            <div>
              <h3 className="text-lg font-semibold mb-2">Chart Error</h3>
              <p className="text-sm text-muted-foreground mb-4">
                {this.props.fallbackMessage ||
                 'An error occurred while rendering the chart. Please try refreshing.'}
              </p>
              {this.state.error && (
                <details className="text-xs text-left bg-muted p-2 rounded mb-4 max-w-md">
                  <summary className="cursor-pointer font-medium mb-2">
                    Technical Details
                  </summary>
                  <pre className="mt-2 whitespace-pre-wrap overflow-x-auto">
                    {this.state.error.message}
                  </pre>
                </details>
              )}
            </div>
            <Button onClick={this.handleReset} variant="default">
              <RefreshCw className="h-4 w-4 mr-2" />
              Try Again
            </Button>
          </div>
        </Card>
      )
    }

    return this.props.children
  }
}
