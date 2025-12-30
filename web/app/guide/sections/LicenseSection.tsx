'use client'

import { Card, CardContent } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  ShieldCheck,
  Key,
  Users,
  Clock,
  CheckCircle2,
  XCircle,
  ChevronRight
} from 'lucide-react'

interface LicenseSectionProps {
  onNavigate: (sectionId: string) => void
}

/**
 * License Section - Comprehensive guide to ISX Pulse licensing
 * Covers licensing model, activation, reactivation, and troubleshooting
 */
export function LicenseSection({ onNavigate }: LicenseSectionProps) {
  const activationSteps = [
    {
      step: 1,
      title: 'Obtain License Key',
      description: 'Receive your license key from your license provider'
    },
    {
      step: 2,
      title: 'Launch ISX Pulse',
      description: 'Open the application - you\'ll be redirected to the license activation page'
    },
    {
      step: 3,
      title: 'Enter License Key',
      description: 'Type or paste your license key in the format: ISX-XXXX-XXXX-XXXX-XXXX'
    },
    {
      step: 4,
      title: 'Wait for Validation',
      description: 'Server validates your key (5-second countdown with device fingerprint)'
    },
    {
      step: 5,
      title: 'Start Using ISX Pulse',
      description: 'Once activated, you\'ll be redirected to the dashboard'
    }
  ]

  const troubleshooting = [
    {
      error: 'Invalid Key Format',
      cause: 'License key doesn\'t match required pattern',
      solution: 'Check format: ISX-XXXX-XXXX-XXXX-XXXX (5 groups of 4 characters)'
    },
    {
      error: 'Device Limit Reached',
      cause: 'License already activated on maximum allowed devices',
      solution: 'Contact support or wait for reactivation window (30 days)'
    },
    {
      error: 'Network Timeout',
      cause: 'Cannot reach activation server',
      solution: 'Check internet connection and firewall settings'
    },
    {
      error: 'Expired License',
      cause: 'License validity period has ended',
      solution: 'Renew your license or purchase a new one'
    }
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-lg bg-primary/10">
            <ShieldCheck className="h-6 w-6 text-primary" />
          </div>
          <div>
            <h2 className="text-3xl font-bold">License System</h2>
            <p className="text-muted-foreground text-lg">
              Understanding ISX Pulse licensing, activation, and management
            </p>
          </div>
        </div>
      </div>

      {/* How Licensing Works */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">How Licensing Works</h3>
        <p className="text-muted-foreground">
          ISX Pulse uses a simple, secure licensing system with hardware-locked activation:
        </p>

        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <Key className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Single License Format</h4>
                  <p className="text-sm text-muted-foreground">
                    All licenses use the format <code className="px-1 py-0.5 rounded bg-muted">ISX-XXXX-XXXX-XXXX-XXXX</code>. Every user gets the same features and capabilities.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <ShieldCheck className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Hardware-Locked</h4>
                  <p className="text-sm text-muted-foreground">
                    Each license is tied to your device's unique fingerprint. This prevents unauthorized sharing while allowing legitimate reactivation.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Clock className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Time-Based Validity</h4>
                  <p className="text-sm text-muted-foreground">
                    License duration is determined during activation. You'll receive expiry warnings 30 days before expiration.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Users className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Reactivation Support</h4>
                  <p className="text-sm text-muted-foreground">
                    Need to move to a new device? You can reactivate up to 5 times per 30-day period without contacting support.
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Activation Process */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Activation Process</h3>
        <p className="text-muted-foreground">
          Follow these steps to activate your ISX Pulse license:
        </p>

        <div className="space-y-3">
          {activationSteps.map((item) => (
            <Card key={item.step} className="relative">
              <CardContent className="pt-6">
                <div className="flex items-start gap-4">
                  <div className="flex items-center justify-center w-10 h-10 rounded-full bg-primary text-primary-foreground font-semibold shrink-0">
                    {item.step}
                  </div>
                  <div className="flex-1 pt-1">
                    <h4 className="font-semibold mb-1">{item.title}</h4>
                    <p className="text-sm text-muted-foreground">
                      {item.description}
                    </p>
                  </div>
                  {item.step < activationSteps.length && (
                    <ChevronRight className="h-5 w-5 text-muted-foreground shrink-0 mt-2" />
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Reactivation */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Reactivation</h3>
        <p className="text-muted-foreground">
          ISX Pulse allows license reactivation if you need to move to a different device:
        </p>

        <Card>
          <CardContent className="pt-6">
            <div className="space-y-4">
              <div className="flex items-start gap-3">
                <Clock className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Reactivation Limits</h4>
                  <p className="text-sm text-muted-foreground">
                    You can reactivate your license up to <strong>5 times per 30-day period</strong>. After that, you must wait for the window to reset or contact support.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Users className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">How It Works</h4>
                  <p className="text-sm text-muted-foreground">
                    When you activate on a new device, ISX Pulse creates a unique device fingerprint. Reactivation on the same device doesn't count against your limit.
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <Key className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold mb-1">Checking Reactivations</h4>
                  <p className="text-sm text-muted-foreground">
                    View your reactivation count on the license page. The counter shows: "3/5 reactivations used (resets in 15 days)"
                  </p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      {/* Troubleshooting */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Troubleshooting</h3>
        <p className="text-muted-foreground">
          Common activation issues and their solutions:
        </p>

        <div className="space-y-3">
          {troubleshooting.map((item, index) => (
            <Card key={index}>
              <CardContent className="pt-6">
                <div className="flex items-start gap-4">
                  <XCircle className="h-5 w-5 text-destructive shrink-0 mt-0.5" />
                  <div className="flex-1 space-y-2">
                    <h4 className="font-semibold">{item.error}</h4>
                    <p className="text-sm text-muted-foreground">
                      <strong>Cause:</strong> {item.cause}
                    </p>
                    <Alert>
                      <CheckCircle2 className="h-4 w-4" />
                      <AlertDescription>
                        <strong>Solution:</strong> {item.solution}
                      </AlertDescription>
                    </Alert>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      {/* Next Steps */}
      <section className="space-y-4">
        <h3 className="text-2xl font-semibold">Next Steps</h3>
        <Card className="border-primary/50 bg-primary/5">
          <CardContent className="pt-6">
            <div className="space-y-4">
              <p>
                Now that you understand the license system, learn how ISX Pulse processes data through its automated pipeline:
              </p>
              <button
                onClick={() => onNavigate('pipeline')}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Continue to Pipeline Architecture
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
