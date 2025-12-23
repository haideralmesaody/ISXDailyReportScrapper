/**
 * License Status Component with Visual Progress Indicators
 * Shows license information, countdown timer, and quick actions
 */

'use client'

import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  Shield,
  CheckCircle,
  AlertTriangle,
  AlertCircle,
  Clock,
  Smartphone,
  Calendar,
  RefreshCw,
  ExternalLink,
  Mail,
  Copy,
  Check
} from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useToast } from '@/lib/hooks/use-toast'
import { useHydration } from '@/lib/hooks'
import { copyToClipboard } from '@/lib/utils/license-helpers'
import { getDeviceDisplayName } from '@/lib/utils/device-fingerprint'
import type { LicenseStatus, LicenseActivationHistory } from '@/types/index'

interface LicenseStatusProps {
  status: LicenseStatus
  onRefresh?: () => void
  onExtend?: () => void
  onTransfer?: () => void
  onSupport?: () => void
  loading?: boolean
  className?: string
}

interface CountdownState {
  days: number
  hours: number
  minutes: number
  seconds: number
}

export default function LicenseStatus({
  status,
  onRefresh,
  onExtend,
  onTransfer,
  onSupport,
  loading = false,
  className = ''
}: LicenseStatusProps) {
  const isHydrated = useHydration()
  const { toast } = useToast()
  const [countdown, setCountdown] = useState<CountdownState>({ days: 0, hours: 0, minutes: 0, seconds: 0 })
  const [copied, setCopied] = useState(false)
  const [refreshing, setRefreshing] = useState(false)

  // Calculate countdown timer
  useEffect(() => {
    if (!isHydrated || !status.isActive) return

    const updateCountdown = () => {
      const now = new Date().getTime()
      const expiry = new Date(status.expiryDate).getTime()
      const distance = expiry - now

      if (distance > 0) {
        const days = Math.floor(distance / (1000 * 60 * 60 * 24))
        const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
        const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60))
        const seconds = Math.floor((distance % (1000 * 60)) / 1000)

        setCountdown({ days, hours, minutes, seconds })
      } else {
        setCountdown({ days: 0, hours: 0, minutes: 0, seconds: 0 })
      }
    }

    updateCountdown()
    const interval = setInterval(updateCountdown, 1000)

    return () => clearInterval(interval)
  }, [isHydrated, status.isActive, status.expiryDate])

  // Get status configuration
  const getStatusConfig = () => {
    switch (status.status) {
      case 'active':
        return {
          icon: CheckCircle,
          color: 'text-green-600',
          bgColor: 'bg-green-50',
          borderColor: 'border-green-200',
          badge: 'default',
          badgeText: 'Active'
        }
      case 'warning':
        return {
          icon: AlertTriangle,
          color: 'text-yellow-600',
          bgColor: 'bg-yellow-50',
          borderColor: 'border-yellow-200',
          badge: 'secondary',
          badgeText: 'Expiring Soon'
        }
      case 'critical':
        return {
          icon: AlertCircle,
          color: 'text-red-600',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
          badge: 'destructive',
          badgeText: 'Critical'
        }
      case 'expired':
        return {
          icon: AlertCircle,
          color: 'text-red-600',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
          badge: 'destructive',
          badgeText: 'Expired'
        }
      default:
        return {
          icon: Shield,
          color: 'text-gray-600',
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200',
          badge: 'secondary',
          badgeText: 'Unknown'
        }
    }
  }

  // Calculate progress percentage
  const getProgressPercentage = () => {
    if (!status.isActive) return 0
    return Math.max(0, Math.min(100, (status.daysRemaining / 30) * 100))
  }

  // Handle copy device ID
  const handleCopyDeviceId = useCallback(async () => {
    const deviceId = status.deviceInfo.fingerprint.slice(0, 16)
    const success = await copyToClipboard(deviceId)
    
    if (success) {
      setCopied(true)
      toast({
        title: 'Copied!',
        description: 'Device ID copied to clipboard',
      })
      setTimeout(() => setCopied(false), 2000)
    }
  }, [status.deviceInfo.fingerprint, toast])

  // Handle refresh
  const handleRefresh = useCallback(async () => {
    if (refreshing || !onRefresh) return
    
    setRefreshing(true)
    try {
      await onRefresh()
      toast({
        title: 'Status Updated',
        description: 'License status has been refreshed',
      })
    } catch (error) {
      toast({
        title: 'Refresh Failed',
        description: 'Unable to refresh license status',
        variant: 'destructive'
      })
    } finally {
      setRefreshing(false)
    }
  }, [refreshing, onRefresh, toast])

  const statusConfig = getStatusConfig()
  const Icon = statusConfig.icon
  const progressPercentage = getProgressPercentage()

  if (!isHydrated) {
    return (
      <Card className={`${className} ${statusConfig.borderColor}`}>
        <CardContent className="p-6">
          <div className="flex items-center justify-center h-40">
            <div className="text-center">
              <Shield className="h-8 w-8 animate-pulse mx-auto mb-2 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Loading license status...</p>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className={className}
    >
      <Card className={`${statusConfig.borderColor} border-2`}>
        <CardHeader className={`${statusConfig.bgColor} border-b`}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Icon className={`h-6 w-6 ${statusConfig.color}`} />
              <div>
                <CardTitle className="text-lg font-semibold">License Status</CardTitle>
                <CardDescription>
                  Current license information and expiry details
                </CardDescription>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={statusConfig.badge as any}>
                {statusConfig.badgeText}
              </Badge>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleRefresh}
                disabled={refreshing || loading}
              >
                <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
              </Button>
            </div>
          </div>
        </CardHeader>

        <CardContent className="p-6 space-y-6">
          {/* Days Remaining with Progress */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium">Days Remaining</span>
              <span className="text-2xl font-bold">
                {status.isActive ? status.daysRemaining : 0}
              </span>
            </div>
            <Progress 
              value={progressPercentage} 
              className="h-3"
              // @ts-ignore - Custom progress color based on status
              style={{
                '--progress-background': status.status === 'active' ? '#10b981' :
                                       status.status === 'warning' ? '#f59e0b' : '#ef4444'
              }}
            />
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>0 days</span>
              <span>30+ days</span>
            </div>
          </div>

          {/* Live Countdown Timer */}
          {status.isActive && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="grid grid-cols-4 gap-4 p-4 bg-muted rounded-lg"
            >
              {[
                { label: 'Days', value: countdown.days },
                { label: 'Hours', value: countdown.hours },
                { label: 'Minutes', value: countdown.minutes },
                { label: 'Seconds', value: countdown.seconds }
              ].map((item, index) => (
                <div key={item.label} className="text-center">
                  <motion.div
                    key={item.value}
                    initial={{ scale: 1.2, opacity: 0.8 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{ duration: 0.2 }}
                    className="text-2xl font-bold text-primary"
                  >
                    {String(item.value).padStart(2, '0')}
                  </motion.div>
                  <div className="text-xs text-muted-foreground uppercase tracking-wide">
                    {item.label}
                  </div>
                </div>
              ))}
            </motion.div>
          )}

          <Separator />

          {/* Device Information */}
          <div>
            <h4 className="font-medium mb-3 flex items-center gap-2">
              <Smartphone className="h-4 w-4" />
              Device Information
            </h4>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Device:</span>
                <span className="font-mono">
                  {getDeviceDisplayName(status.deviceInfo as any)}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Device ID:</span>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs">
                    {status.deviceInfo.fingerprint.slice(0, 16)}...
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCopyDeviceId}
                    className="h-6 w-6 p-0"
                  >
                    {copied ? (
                      <Check className="h-3 w-3" />
                    ) : (
                      <Copy className="h-3 w-3" />
                    )}
                  </Button>
                </div>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">First Activation:</span>
                <span>{new Date(status.deviceInfo.first_activation).toLocaleDateString()}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Last Seen:</span>
                <span>{new Date(status.deviceInfo.last_seen).toLocaleDateString()}</span>
              </div>
            </div>
          </div>

          <Separator />

          {/* Features */}
          <div>
            <h4 className="font-medium mb-3">Licensed Features</h4>
            <div className="flex flex-wrap gap-2">
              {status.features.map((feature) => (
                <Badge key={feature} variant="outline" className="text-xs">
                  {feature}
                </Badge>
              ))}
            </div>
          </div>

          {/* Activation History Preview */}
          {status.activationHistory.length > 0 && (
            <>
              <Separator />
              <div>
                <h4 className="font-medium mb-3 flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  Recent Activity
                </h4>
                <div className="space-y-2">
                  {status.activationHistory.slice(0, 3).map((entry) => (
                    <div key={entry.id} className="flex items-center justify-between text-sm">
                      <div>
                        <span className="capitalize">{entry.action}</span>
                        <span className="text-muted-foreground ml-2">
                          on {new Date(entry.date).toLocaleDateString()}
                        </span>
                      </div>
                      <Badge variant={entry.success ? "default" : "destructive"} className="text-xs">
                        {entry.success ? "Success" : "Failed"}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}

          <Separator />

          {/* Quick Actions */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <Button
              variant="outline"
              size="sm"
              onClick={onExtend}
              className="flex items-center gap-2"
            >
              <Clock className="h-4 w-4" />
              Extend License
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={onTransfer}
              className="flex items-center gap-2"
            >
              <ExternalLink className="h-4 w-4" />
              Transfer License
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={onSupport}
              className="flex items-center gap-2"
            >
              <Mail className="h-4 w-4" />
              Get Support
            </Button>
          </div>

          {/* Expiry Warning */}
          <AnimatePresence>
            {status.status === 'warning' && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                className="bg-yellow-50 border border-yellow-200 rounded-lg p-4"
              >
                <div className="flex items-center gap-2 text-yellow-800">
                  <AlertTriangle className="h-4 w-4" />
                  <span className="font-medium">License Expiring Soon</span>
                </div>
                <p className="text-sm text-yellow-700 mt-1">
                  Your license will expire in {status.daysRemaining} days. 
                  Consider extending your license to avoid service interruption.
                </p>
              </motion.div>
            )}
          </AnimatePresence>

          <AnimatePresence>
            {status.status === 'critical' && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                exit={{ opacity: 0, height: 0 }}
                className="bg-red-50 border border-red-200 rounded-lg p-4"
              >
                <div className="flex items-center gap-2 text-red-800">
                  <AlertCircle className="h-4 w-4" />
                  <span className="font-medium">Critical: License Expires Soon</span>
                </div>
                <p className="text-sm text-red-700 mt-1">
                  Your license expires in {status.daysRemaining} days. 
                  Immediate action required to maintain access.
                </p>
              </motion.div>
            )}
          </AnimatePresence>
        </CardContent>
      </Card>
    </motion.div>
  )
}