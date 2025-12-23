/**
 * Lazy-loaded License Activation Form wrapper
 * Improves initial bundle size for users with valid licenses
 */

'use client'

import { lazy, Suspense } from 'react'
import { Loader2 } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

const LicenseActivationFormComponent = lazy(() => import('./LicenseActivationFormComponent'))

interface LicenseActivationFormProps {
  onSubmit: (data: any) => Promise<void>
  loading: boolean
  error: any
  activationProgress: number
  licenseState: 'invalid' | 'expired'
}

export function LicenseActivationForm(props: LicenseActivationFormProps) {
  return (
    <Suspense 
      fallback={
        <Card>
          <CardContent className="flex justify-center items-center py-12">
            <div className="text-center space-y-4">
              <Loader2 className="h-8 w-8 animate-spin mx-auto text-primary" />
              <p className="text-sm text-muted-foreground">Loading activation form...</p>
            </div>
          </CardContent>
        </Card>
      }
    >
      <LicenseActivationFormComponent {...props} />
    </Suspense>
  )
}