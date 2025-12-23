/**
 * Generic API hook for managing loading states and errors
 */

'use client'

import { useState, useCallback } from 'react'
import { ISXApiError } from '@/lib/api'

interface UseApiState<T> {
  data: T | null
  loading: boolean
  error: ISXApiError | null
}

interface UseApiReturn<T, TArgs extends unknown[] = unknown[]> extends UseApiState<T> {
  execute: (...args: TArgs) => Promise<T>
  reset: () => void
}

export function useApi<T, TArgs extends unknown[] = unknown[]>(
  apiFunction: (...args: TArgs) => Promise<T>
): UseApiReturn<T, TArgs> {
  const [state, setState] = useState<UseApiState<T>>({
    data: null,
    loading: false,
    error: null,
  })

  const execute = useCallback(
    async (...args: TArgs): Promise<T> => {
      setState(prev => ({ ...prev, loading: true, error: null }))

      try {
        const result = await apiFunction(...args)
        setState({ data: result, loading: false, error: null })
        return result
      } catch (error) {
        const apiError = error instanceof ISXApiError 
          ? error 
          : new ISXApiError({
              type: '/problems/unknown-error',
              title: 'Unknown Error',
              status: 500,
              detail: error instanceof Error ? error.message : 'An unknown error occurred',
            })

        setState(prev => ({ ...prev, loading: false, error: apiError }))
        throw apiError
      }
    },
    [apiFunction]
  )

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null })
  }, [])

  return {
    ...state,
    execute,
    reset,
  }
}