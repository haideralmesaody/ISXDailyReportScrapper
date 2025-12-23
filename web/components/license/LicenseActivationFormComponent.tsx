/**
 * Smart License Activation Form Component
 * Features auto-formatting, real-time validation, and rate limiting
 */

'use client'

import { useState, useCallback, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { AlertCircle, Loader2, Shield, Copy, Check, UserCheck, Clock, Search, WifiOff, Timer, Ban, AlertTriangle, Mail, RefreshCw, RotateCcw, XCircle, Smartphone } from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

import { licenseActivationSchema, type LicenseActivationForm } from '@/lib/schemas'
import type { ApiError } from '@/types/index'
import { 
  cleanLicenseKey,
  formatLicenseKey,
  detectLicenseFormat,
  isValidLicenseFormat,
  canAttemptActivation,
  recordActivationAttempt,
  copyToClipboard,
  clearRateLimitData
} from '@/lib/utils/license-helpers'
import { generateDeviceFingerprint } from '@/lib/utils/device-fingerprint'
import { parseLicenseError, LicenseErrorType, getErrorIcon, getErrorColor } from '@/lib/utils/license-errors'

interface LicenseActivationFormComponentProps {
  onSubmit: (data: LicenseActivationForm) => Promise<void>
  loading: boolean
  error: ApiError | null
  activationProgress: number
  licenseState: 'invalid' | 'expired'
}

export default function LicenseActivationFormComponent({
  onSubmit,
  loading,
  error,
  licenseState
}: LicenseActivationFormComponentProps) {
  const [licenseKey, setLicenseKey] = useState('')
  const [displayKey, setDisplayKey] = useState('')
  const [keyFormat, setKeyFormat] = useState<'standard' | 'scratch'>('standard')
  const [isValidFormat, setIsValidFormat] = useState(false)
  const [rateLimitError, setRateLimitError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [generatingFingerprint, setGeneratingFingerprint] = useState(false)
  
  const {
    handleSubmit,
    formState: { errors },
    setValue,
  } = useForm<LicenseActivationForm>({
    resolver: zodResolver(licenseActivationSchema),
    mode: 'onChange',
  })

  // Handle license key input with auto-formatting
  const handleKeyChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    const cleanKey = cleanLicenseKey(value)
    
    // Detect format and auto-format for display
    const detectedFormat = detectLicenseFormat(cleanKey)
    const formattedKey = formatLicenseKey(value, detectedFormat)
    
    // Update states
    setLicenseKey(cleanKey)
    setDisplayKey(formattedKey)
    setKeyFormat(detectedFormat)
    setIsValidFormat(isValidLicenseFormat(cleanKey))
    
    // Update form value with clean key (no dashes)
    setValue('license_key', cleanKey)
  }, [setValue])

  // Handle paste events
  const handlePaste = useCallback((e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault()
    const pasted = e.clipboardData.getData('text')
    const cleanKey = cleanLicenseKey(pasted)
    
    // Detect format and auto-format for display
    const detectedFormat = detectLicenseFormat(cleanKey)
    const formattedKey = formatLicenseKey(pasted, detectedFormat)
    
    // Update states
    setLicenseKey(cleanKey)
    setDisplayKey(formattedKey)
    setKeyFormat(detectedFormat)
    setIsValidFormat(isValidLicenseFormat(cleanKey))
    setValue('license_key', cleanKey)
  }, [setValue])

  // Handle form submission with rate limiting and device fingerprinting
  const handleFormSubmit = async (data: LicenseActivationForm) => {
    // Check rate limit
    const { allowed, resetTime } = canAttemptActivation()
    
    if (!allowed) {
      const resetIn = Math.ceil((resetTime! - Date.now()) / 1000)
      const minutes = Math.floor(resetIn / 60)
      const seconds = resetIn % 60
      const timeStr = minutes > 0 ? `${minutes} minute${minutes > 1 ? 's' : ''} ${seconds} second${seconds !== 1 ? 's' : ''}` : `${resetIn} seconds`
      setRateLimitError(`Too many activation attempts. Please wait ${timeStr} before trying again.`)
      return
    }
    
    // Generate device fingerprint
    setGeneratingFingerprint(true)
    try {
      const deviceFingerprint = await generateDeviceFingerprint()
      
      // Record attempt
      recordActivationAttempt()
      setRateLimitError(null)
      
      // Enhanced data with device fingerprint
      const enhancedData = {
        ...data,
        device_fingerprint: deviceFingerprint,
        key_format: keyFormat
      }
      
      // Submit
      await onSubmit(enhancedData as any)
    } catch (error) {
      console.error('Failed to generate device fingerprint:', error)
      // Still attempt activation without fingerprint
      await onSubmit(data)
    } finally {
      setGeneratingFingerprint(false)
    }
  }

  // Copy example key to clipboard (use scratch card format)
  const copyExampleKey = useCallback(async () => {
    const example = keyFormat === 'scratch' ? 'ISX-1M02-LYE1-F9QJ' : 'ISX1M02LYE1F9QJHR9D7Z'
    const success = await copyToClipboard(example)
    if (success) {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
    return
  }, [keyFormat])

  // Clear rate limit error after timeout
  useEffect(() => {
    if (rateLimitError) {
      const timer = setTimeout(() => {
        setRateLimitError(null)
      }, 5000)
      return () => clearTimeout(timer)
    }
  }, [rateLimitError])

  const getCardTitle = () => {
    return licenseState === 'expired' ? 'Renew License' : 'Activate License'
  }

  const getCardDescription = () => {
    return licenseState === 'expired' 
      ? 'Your license has expired. Enter a new key to renew.'
      : 'Enter your license key to unlock professional features.'
  }

  const getSubmitText = () => {
    if (generatingFingerprint) return 'Securing device...'
    if (loading) return 'Activating...'
    return licenseState === 'expired' ? 'Renew License' : 'Activate License'
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-primary" />
          <span>{getCardTitle()}</span>
        </CardTitle>
        <CardDescription>{getCardDescription()}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-4">
          {/* License Key Input */}
          <div className="space-y-2">
            <Label htmlFor="license_key">License Key</Label>
            <div className="relative">
              <Input
                id="license_key"
                type="text"
                placeholder={keyFormat === 'scratch' ? 'ISX-XXXX-XXXX-XXXX' : 'Enter your license key'}
                value={displayKey}
                onChange={handleKeyChange}
                onPaste={handlePaste}
                className={`font-mono pr-10 transition-colors ${
                  isValidFormat ? 'border-green-500 focus:border-green-500' : ''
                }`}
                disabled={loading || generatingFingerprint}
                autoComplete="off"
                spellCheck={false}
                autoFocus
                maxLength={keyFormat === 'scratch' ? 19 : 25} // ISX-XXXX-XXXX-XXXX = 19 chars
              />
              {isValidFormat && (
                <Check className="absolute right-3 top-3 h-4 w-4 text-green-500" />
              )}
            </div>
            
            {/* Format hint */}
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <div className="flex items-center gap-2">
                <span>
                  {keyFormat === 'scratch' 
                    ? 'Scratch card format: ISX-XXXX-XXXX-XXXX'
                    : 'Standard format: ISX followed by alphanumeric code'
                  }
                </span>
                {keyFormat === 'scratch' && (
                  <Badge variant="outline" className="text-xs px-1 py-0">
                    Scratch Card
                  </Badge>
                )}
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-auto p-0 text-xs"
                onClick={copyExampleKey}
              >
                {copied ? (
                  <>
                    <Check className="h-3 w-3 mr-1" />
                    Copied
                  </>
                ) : (
                  <>
                    <Copy className="h-3 w-3 mr-1" />
                    Copy example
                  </>
                )}
              </Button>
            </div>
            
            {/* Validation error */}
            {errors.license_key && (
              <p className="text-sm text-destructive flex items-center gap-1">
                <AlertCircle className="h-3 w-3" />
                {errors.license_key.message}
              </p>
            )}
          </div>

          {/* Rate limit error */}
          {rateLimitError && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{rateLimitError}</AlertDescription>
            </Alert>
          )}

          {/* Enhanced API error display */}
          {error && !rateLimitError && (() => {
            const errorDetails = parseLicenseError(error)
            const IconComponent = {
              RotateCcw,
              XCircle,
              Smartphone,
              UserCheck,
              AlertCircle,
              Clock,
              Search,
              WifiOff,
              Timer,
              Ban,
              AlertTriangle
            }[getErrorIcon(errorDetails.type) as keyof typeof IconComponent] || AlertTriangle
            
            const errorColor = getErrorColor(errorDetails.type)
            const variantMap = {
              'green': 'default',
              'orange': 'default',
              'yellow': 'default',
              'blue': 'default',
              'red': 'destructive'
            } as const
            
            return (
              <Alert variant={variantMap[errorColor as keyof typeof variantMap] || 'destructive'} 
                     className={errorColor === 'green' ? 'border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-950/30' : ''}>
                <IconComponent className={`h-4 w-4 ${errorColor === 'green' ? 'text-green-700 dark:text-green-400' : ''}`} />
                <AlertTitle className="font-semibold">{errorDetails.title}</AlertTitle>
                <AlertDescription className="mt-2 space-y-3">
                  <p>{errorDetails.message}</p>
                  
                  {/* Show reactivation success details */}
                  {errorDetails.type === LicenseErrorType.REACTIVATION_SUCCESS && (
                    <div className="bg-green-50 dark:bg-green-950/30 rounded-md p-3 border border-green-200 dark:border-green-800">
                      <p className="text-sm font-medium text-green-800 dark:text-green-200 mb-1">
                        Reactivation Complete
                      </p>
                      <div className="space-y-1 text-xs text-green-700 dark:text-green-300">
                        {errorDetails.details?.reactivationCount !== undefined && (
                          <p>Reactivation #{errorDetails.details.reactivationCount}</p>
                        )}
                        {errorDetails.details?.similarityScore !== undefined && (
                          <p>Device similarity: {Math.round(errorDetails.details.similarityScore * 100)}%</p>
                        )}
                        <p>Your license is now active on this device.</p>
                      </div>
                    </div>
                  )}
                  
                  {/* Show reactivation limit exceeded details */}
                  {errorDetails.type === LicenseErrorType.REACTIVATION_LIMIT_EXCEEDED && (
                    <div className="bg-red-50 dark:bg-red-950/30 rounded-md p-3 border border-red-200 dark:border-red-800">
                      <p className="text-sm font-medium text-red-800 dark:text-red-200 mb-1">
                        Maximum Reactivations Reached
                      </p>
                      <div className="space-y-1 text-xs text-red-700 dark:text-red-300">
                        {errorDetails.details?.reactivationCount !== undefined && errorDetails.details?.reactivationLimit !== undefined && (
                          <p>Used {errorDetails.details.reactivationCount} of {errorDetails.details.reactivationLimit} allowed reactivations</p>
                        )}
                        <p>Contact support to increase your reactivation limit or transfer to a new license.</p>
                      </div>
                    </div>
                  )}
                  
                  {/* Show different device activation details */}
                  {errorDetails.type === LicenseErrorType.ALREADY_ACTIVATED_DIFFERENT_DEVICE && (
                    <div className="bg-orange-50 dark:bg-orange-950/30 rounded-md p-3 border border-orange-200 dark:border-orange-800">
                      <p className="text-sm font-medium text-orange-800 dark:text-orange-200 mb-1">
                        License Active on Different Device
                      </p>
                      <div className="space-y-1 text-xs text-orange-700 dark:text-orange-300">
                        {errorDetails.details?.similarityScore !== undefined && (
                          <p>Device similarity: {Math.round(errorDetails.details.similarityScore * 100)}%</p>
                        )}
                        {errorDetails.details?.remainingAttempts !== undefined && (
                          <p>Remaining reactivation attempts: {errorDetails.details.remainingAttempts}</p>
                        )}
                        <p>If this is your device, contact support for assistance with reactivation.</p>
                      </div>
                    </div>
                  )}
                  
                  {/* Show specific guidance for already activated licenses */}
                  {errorDetails.type === LicenseErrorType.ALREADY_ACTIVATED && (
                    <div className="bg-orange-50 dark:bg-orange-950/30 rounded-md p-3 border border-orange-200 dark:border-orange-800">
                      <p className="text-sm font-medium text-orange-800 dark:text-orange-200 mb-1">
                        License Transfer Required
                      </p>
                      <p className="text-xs text-orange-700 dark:text-orange-300">
                        This license is registered to another device. To transfer it to this device, 
                        please contact our support team with your license key and proof of purchase.
                      </p>
                    </div>
                  )}
                  
                  {/* Action buttons */}
                  {errorDetails.actions && errorDetails.actions.length > 0 && (
                    <>
                      <Separator className="my-2" />
                      <div className="flex flex-wrap gap-2">
                        {errorDetails.actions.map((action, index) => {
                          if (action.action === 'contact_support' && action.href) {
                            return (
                              <Button
                                key={index}
                                variant="outline"
                                size="sm"
                                asChild
                              >
                                <a href={action.href}>
                                  <Mail className="h-3 w-3 mr-1" />
                                  {action.label}
                                </a>
                              </Button>
                            )
                          }
                          
                          if (action.action === 'try_again') {
                            return (
                              <Button
                                key={index}
                                variant="outline"
                                size="sm"
                                type="button"
                                onClick={() => window.location.reload()}
                              >
                                <RefreshCw className="h-3 w-3 mr-1" />
                                {action.label}
                              </Button>
                            )
                          }
                          
                          return (
                            <Button
                              key={index}
                              variant="ghost"
                              size="sm"
                              type="button"
                            >
                              {action.label}
                            </Button>
                          )
                        })}
                      </div>
                    </>
                  )}
                  
                  {/* Error ID for support */}
                  {error.trace_id && (
                    <div className="text-xs opacity-75 pt-2 border-t">
                      Error ID: {error.trace_id}
                    </div>
                  )}
                </AlertDescription>
              </Alert>
            )
          })()}

          {/* Submit Button */}
          <Button 
            type="submit" 
            className="w-full" 
            disabled={loading || generatingFingerprint || !isValidFormat || !!rateLimitError}
          >
            {(loading || generatingFingerprint) ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {getSubmitText()}
              </>
            ) : (
              getSubmitText()
            )}
          </Button>

          {/* Help text */}
          <p className="text-center text-sm text-muted-foreground">
            Need help? Contact{' '}
            <a 
              href="mailto:support@isxpulse.com" 
              className="text-primary hover:underline"
            >
              support@isxpulse.com
            </a>
          </p>
        </form>
      </CardContent>
    </Card>
  )
}