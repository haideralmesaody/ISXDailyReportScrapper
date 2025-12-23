/**
 * License Stacking Confirmation Dialog
 * Shows existing license details and confirms stacking with new license
 */

import React from 'react'
import { 
  AlertCircle, 
  Calendar, 
  Key, 
  PlusCircle, 
  Clock,
  CheckCircle2,
  XCircle
} from 'lucide-react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { format, addDays, differenceInDays } from 'date-fns'

interface ExistingLicenseInfo {
  has_license: boolean
  days_remaining: number
  expiry_date: string
  license_key: string
  status: string
  is_expired: boolean
}

interface LicenseStackingDialogProps {
  open: boolean
  onConfirm: () => void
  onCancel: () => void
  existingLicense: ExistingLicenseInfo | null
  newLicenseKey: string
  newDuration?: string // e.g., "30 days", "1 month", etc.
}

export function LicenseStackingDialog({
  open,
  onConfirm,
  onCancel,
  existingLicense,
  newLicenseKey,
  newDuration = '30 days'
}: LicenseStackingDialogProps) {
  if (!existingLicense || !existingLicense.has_license) {
    // No existing license - proceed with normal activation
    return null
  }

  // Parse duration to days
  const parseDurationToDays = (duration: string): number => {
    const match = duration.match(/(\d+)\s*(day|month|year)/i)
    if (!match) return 30

    const value = parseInt(match[1])
    const unit = match[2].toLowerCase()

    switch (unit) {
      case 'month':
        return value * 30
      case 'year':
        return value * 365
      default:
        return value
    }
  }

  const newDurationDays = parseDurationToDays(newDuration)
  const currentExpiry = new Date(existingLicense.expiry_date)
  const newExpiry = addDays(currentExpiry, newDurationDays)
  const totalDays = differenceInDays(newExpiry, new Date())

  const isExpired = existingLicense.is_expired

  return (
    <AlertDialog open={open} onOpenChange={(open) => !open && onCancel()}>
      <AlertDialogContent className="max-w-2xl">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2">
            {isExpired ? (
              <>
                <XCircle className="h-5 w-5 text-destructive" />
                Replace Expired License
              </>
            ) : (
              <>
                <PlusCircle className="h-5 w-5 text-primary" />
                Stack License - Extend Your Current License
              </>
            )}
          </AlertDialogTitle>
          <AlertDialogDescription className="text-left space-y-4 pt-4">
            {/* Warning/Info Banner */}
            <div className={`p-4 rounded-lg border ${
              isExpired 
                ? 'bg-destructive/10 border-destructive/20' 
                : 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800'
            }`}>
              <div className="flex gap-3">
                <AlertCircle className={`h-5 w-5 mt-0.5 ${
                  isExpired ? 'text-destructive' : 'text-blue-600 dark:text-blue-400'
                }`} />
                <div className="flex-1">
                  <p className="font-medium mb-1">
                    {isExpired 
                      ? 'Your current license has expired'
                      : 'You already have an active license'}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    {isExpired
                      ? 'Activating the new license will replace your expired license.'
                      : 'Activating the new license will add its duration to your existing license, extending your expiry date.'}
                  </p>
                </div>
              </div>
            </div>

            {/* Current License Details */}
            <Card>
              <CardContent className="pt-6">
                <h4 className="font-semibold mb-3 flex items-center gap-2">
                  <Key className="h-4 w-4" />
                  Current License
                </h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">License Key:</span>
                    <code className="font-mono">{existingLicense.license_key}</code>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Status:</span>
                    <Badge variant={isExpired ? 'destructive' : 'success'}>
                      {isExpired ? 'Expired' : 'Active'}
                    </Badge>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Expiry Date:</span>
                    <span className={isExpired ? 'text-destructive' : ''}>
                      {format(currentExpiry, 'PPP')}
                    </span>
                  </div>
                  {!isExpired && (
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Days Remaining:</span>
                      <span className="font-medium">{existingLicense.days_remaining} days</span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Arrow or Separator */}
            <div className="flex items-center justify-center py-2">
              <div className="flex-1 border-t" />
              <span className="px-4 text-muted-foreground text-sm">
                {isExpired ? 'Replace with' : 'Add'}
              </span>
              <div className="flex-1 border-t" />
            </div>

            {/* New License Details */}
            <Card>
              <CardContent className="pt-6">
                <h4 className="font-semibold mb-3 flex items-center gap-2">
                  <PlusCircle className="h-4 w-4" />
                  New License to Activate
                </h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">License Key:</span>
                    <code className="font-mono">{newLicenseKey}</code>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Duration:</span>
                    <span className="font-medium">{newDuration}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Additional Days:</span>
                    <span className="font-medium text-green-600 dark:text-green-400">
                      +{newDurationDays} days
                    </span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Result Summary */}
            {!isExpired && (
              <>
                <Separator />
                <Card className="border-primary/20 bg-primary/5">
                  <CardContent className="pt-6">
                    <h4 className="font-semibold mb-3 flex items-center gap-2 text-primary">
                      <CheckCircle2 className="h-4 w-4" />
                      After Stacking
                    </h4>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">New Expiry Date:</span>
                        <span className="font-medium text-primary">
                          {format(newExpiry, 'PPP')}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">Total Days:</span>
                        <span className="font-medium text-primary">{totalDays} days</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onCancel}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm}>
            {isExpired ? 'Replace License' : 'Stack & Extend License'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}