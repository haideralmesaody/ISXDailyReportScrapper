'use client'

import { usePathname } from 'next/navigation'
import Link from 'next/link'
import {
  Settings,
  BarChart3,
  FileText,
  Menu,
  X,
  Shield,
  ShieldCheck,
  Clock,
  TrendingUp,
  Target,
  LayoutGrid,
  BookOpen,
  Key,
} from 'lucide-react'
import { useState, useEffect, useRef } from 'react'

import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ThemeToggle } from '@/components/ui/theme-toggle'
import { useConnectionStatus, useSystemStatus } from '@/lib/hooks/use-websocket'
import { InvestorLogo, InvestorLogoCompact } from '@/components/layout/investor-logo'
import { ErrorBoundary } from '@/components/ui/error-boundary'
import { cn } from '@/lib/utils'
import { apiClient } from '@/lib/api'
import type { LicenseApiResponse } from '@/types/index'
import { getCachedLicenseStatus, LICENSE_STATUS_UPDATED, getLastLicenseBroadcast } from '@/lib/utils/license-helpers'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

// Enhanced license status hook with robust event handling and race condition prevention
function useSimpleLicenseStatus() {
  const [response, setResponse] = useState<LicenseApiResponse | null>(null)
  const [loading, setLoading] = useState(true)

  // Refs to prevent stale closures and race conditions
  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const lastEventRef = useRef<number>(0)
  const isUnmountedRef = useRef<boolean>(false)
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    isUnmountedRef.current = false

    const fetchLicenseStatus = async (isImmediate: boolean = false) => {
      // Prevent fetches if component is unmounted
      if (isUnmountedRef.current) return

      try {
        const data = await apiClient.getLicenseStatus()

        // Prevent state updates if component is unmounted
        if (isUnmountedRef.current) return

        setResponse(data)
      } catch (error) {
        if (isUnmountedRef.current) return

        console.error('License status fetch failed:', error)
        setResponse(null)
      } finally {
        if (isUnmountedRef.current) return

        setLoading(false)
      }
    }

    // Debounced event handler to prevent rapid successive updates
    const debouncedEventHandler = (event: Event) => {
      // Clear existing debounce timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
      }

      debounceTimerRef.current = setTimeout(() => {
        if (isUnmountedRef.current) return

        const customEvent = event as CustomEvent
        const eventData = customEvent.detail as { status: string; expiryDate?: string; timestamp: number }

        // Prevent processing old events (race condition protection)
        if (eventData.timestamp && eventData.timestamp <= lastEventRef.current) {
          console.log('Ignoring stale license event:', eventData.timestamp)
          return
        }

        // Update last event timestamp
        lastEventRef.current = eventData.timestamp

        console.log('License status update event received:', eventData)

        // Cancel current polling interval to prevent race conditions
        if (pollIntervalRef.current) {
          clearInterval(pollIntervalRef.current)
          pollIntervalRef.current = null
        }

        // Update state immediately with active snapshot for instant UI response
        if (eventData.status === 'active' || eventData.status === 'warning') {
          const activeSnapshot: LicenseApiResponse = {
            license_status: eventData.status as any,
            status: 'valid',
            message: 'License is active',
            expiry_date: eventData.expiryDate,
            days_left: eventData.expiryDate ? Math.max(0, Math.ceil((new Date(eventData.expiryDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24))) : undefined,
            last_check: new Date(eventData.timestamp).toISOString()
          }

          if (!isUnmountedRef.current) {
            setResponse(activeSnapshot)
            setLoading(false)
          }
        }

        // Trigger a fresh API call for verification and complete data
        fetchLicenseStatus(true)
      }, 100) // 100ms debounce delay
    }

    // Enhanced event listener with error boundary
    const handleLicenseStatusUpdated = (event: Event) => {
      try {
        debouncedEventHandler(event)
      } catch (error) {
        console.error('Error in license status event handler:', error)
        // Fallback: fetch fresh data on error
        if (!isUnmountedRef.current) {
          fetchLicenseStatus(true)
        }
      }
    }

    // Initialize with cached license status for immediate UI state
    const initializeWithCache = () => {
      if (isUnmountedRef.current) return

      try {
        const cachedStatus = getCachedLicenseStatus()
        if (cachedStatus) {
          // Create a temporary active snapshot from cached data
          const cachedResponse: LicenseApiResponse = {
            license_status: cachedStatus.status as any,
            status: 'valid',
            message: cachedStatus.status === 'active' ? 'License is active' : 'License status from cache',
            expiry_date: cachedStatus.expiryDate,
            days_left: cachedStatus.expiryDate ? Math.max(0, Math.ceil((new Date(cachedStatus.expiryDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24))) : undefined,
            last_check: new Date(cachedStatus.cachedAt).toISOString()
          }

          if (!isUnmountedRef.current) {
            setResponse(cachedResponse)
            setLoading(false) // Set loading to false since we have cached data
          }
        }
      } catch (error) {
        console.error('Error loading cached license status:', error)
      }
    }

    // Initialize polling with proper cleanup and fallback broadcast checking
    const startPolling = () => {
      if (isUnmountedRef.current) return

      // Clear any existing interval
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
      }

      // Poll every 5 minutes for updates, with additional broadcast fallback checking
      pollIntervalRef.current = setInterval(() => {
        if (!isUnmountedRef.current) {
          // Check for global broadcast fallback (in case events failed)
          const lastBroadcast = getLastLicenseBroadcast()
          if (lastBroadcast && lastBroadcast.timestamp > lastEventRef.current) {
            console.log('Processing fallback license broadcast:', lastBroadcast)

            // Process the fallback broadcast
            lastEventRef.current = lastBroadcast.timestamp
            if (lastBroadcast.status === 'active' || lastBroadcast.status === 'warning') {
              const activeSnapshot: LicenseApiResponse = {
                license_status: lastBroadcast.status as any,
                status: 'valid',
                message: 'License is active',
                expiry_date: lastBroadcast.expiryDate,
                days_left: lastBroadcast.expiryDate ? Math.max(0, Math.ceil((new Date(lastBroadcast.expiryDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24))) : undefined,
                last_check: new Date(lastBroadcast.timestamp).toISOString()
              }

              if (!isUnmountedRef.current) {
                setResponse(activeSnapshot)
                setLoading(false)
              }
            }

            // Trigger fresh API call for verification
            fetchLicenseStatus(true)
          } else {
            // Regular polling
            fetchLicenseStatus()
          }
        }
      }, 300000)
    }

    // Add event listener for real-time updates with CustomEvent API fallback
    let eventListenerAdded = false
    try {
      if (typeof window !== 'undefined' && window.addEventListener) {
        window.addEventListener(LICENSE_STATUS_UPDATED, handleLicenseStatusUpdated)
        eventListenerAdded = true
      }
    } catch (error) {
      console.warn('Failed to add license status event listener:', error)
      // Continue without event listener - will rely on polling only
    }

    // Initialize component state
    initializeWithCache()

    // Fetch fresh license status
    fetchLicenseStatus()

    // Start polling interval
    startPolling()

    // Comprehensive cleanup function
    return () => {
      // Mark component as unmounted
      isUnmountedRef.current = true

      // Clear debounce timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current)
        debounceTimerRef.current = null
      }

      // Clear polling interval
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
        pollIntervalRef.current = null
      }

      // Remove event listener
      if (eventListenerAdded && typeof window !== 'undefined' && window.removeEventListener) {
        try {
          window.removeEventListener(LICENSE_STATUS_UPDATED, handleLicenseStatusUpdated)
        } catch (error) {
          console.warn('Failed to remove license status event listener:', error)
        }
      }
    }
  }, [])

  return { response, loading }
}

// Navigation items configuration
const navigationItems = [
  {
    name: 'Operations',
    href: '/operations',
    icon: Settings,
    description: 'Manage data collection and processing',
    requiresLicense: true
  },
  {
    name: 'Strategy',
    href: '/strategy',
    icon: Target,
    description: 'Batch trading strategies (in development)',
    requiresLicense: true
  },
  {
    name: 'Liquidity',
    href: '/liquidity',
    icon: BarChart3,
    description: 'ISX Hybrid Liquidity Metrics and scoring',
    requiresLicense: true
  },
  {
    name: 'Market Overview',
    href: '/market-overview',
    icon: LayoutGrid,
    description: 'Interactive treemap visualization of ISX market data',
    requiresLicense: true
  },
  {
    name: 'Reports',
    href: '/reports',
    icon: FileText,
    description: 'Generated reports and data exports',
    requiresLicense: true
  },
  {
    name: 'Guide',
    href: '/guide',
    icon: BookOpen,
    description: 'Interactive tutorials and documentation',
    requiresLicense: true
  },
  {
    name: 'License',
    href: '/license',
    icon: Key,
    description: 'License management and activation',
    requiresLicense: false // Always accessible
  }
]

interface NavigationProps {
  currentPath: string
  isMobileMenuOpen: boolean
  onMobileMenuToggle: () => void
  licenseStatus: LicenseApiResponse | null
}

function Navigation({ currentPath, isMobileMenuOpen, onMobileMenuToggle, licenseStatus }: NavigationProps) {
  // Check if license is valid (active or warning)
  const isLicenseValid = licenseStatus?.license_status === 'active' || licenseStatus?.license_status === 'warning'

  // Get license badge info
  const getLicenseBadge = () => {
    if (!licenseStatus) return null
    const status = licenseStatus.license_status?.toLowerCase()
    const daysLeft = licenseStatus.days_left || 0

    if (status === 'active') {
      return { text: `Active (${daysLeft}d)`, variant: 'default' as const, color: 'text-green-500' }
    } else if (status === 'warning') {
      return { text: `Warning (${daysLeft}d)`, variant: 'secondary' as const, color: 'text-yellow-500' }
    } else {
      return { text: 'Expired', variant: 'destructive' as const, color: 'text-red-500' }
    }
  }

  const licenseBadge = getLicenseBadge()

  return (
    <>
      {/* Desktop Navigation */}
      <nav className="hidden md:flex items-center space-x-6">
        {navigationItems.map((item) => {
          const isActive = currentPath === item.href ||
                          (item.href !== '/' && currentPath.startsWith(item.href))
          const isDisabled = item.requiresLicense && !isLicenseValid
          const isLicenseItem = item.href === '/license'

          // If disabled, render as span with tooltip
          if (isDisabled) {
            return (
              <TooltipProvider key={item.href}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div
                      className={cn(
                        "text-sm font-medium transition-colors relative group cursor-not-allowed opacity-50",
                        "text-muted-foreground"
                      )}
                    >
                      <div className="flex items-center space-x-2">
                        <item.icon className="h-4 w-4" />
                        <span>{item.name}</span>
                      </div>
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Please activate your license to access this feature</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )
          }

          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "text-sm font-medium transition-colors relative group",
                isActive
                  ? "text-foreground"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <div className="flex items-center space-x-2">
                <item.icon className="h-4 w-4" />
                <span>{item.name}</span>
                {isLicenseItem && licenseBadge && (
                  <Badge variant={licenseBadge.variant} className="text-[10px] px-1 py-0">
                    {licenseBadge.text}
                  </Badge>
                )}
              </div>
              {isActive && (
                <div className="absolute -bottom-6 left-0 right-0 h-0.5 bg-primary" />
              )}

              {/* Tooltip */}
              <div className="absolute top-full left-1/2 transform -translate-x-1/2 mt-2 px-2 py-1 bg-background border rounded shadow-lg text-xs whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-50">
                {item.description}
              </div>
            </Link>
          )
        })}
      </nav>

      {/* Mobile Menu Button */}
      <div className="md:hidden">
        <Button
          variant="ghost"
          size="sm"
          onClick={onMobileMenuToggle}
          aria-label="Toggle navigation menu"
        >
          {isMobileMenuOpen ? (
            <X className="h-5 w-5" />
          ) : (
            <Menu className="h-5 w-5" />
          )}
        </Button>
      </div>

      {/* Mobile Navigation */}
      {isMobileMenuOpen && (
        <div className="md:hidden absolute top-16 left-0 right-0 bg-background border-b shadow-lg z-40">
          <div className="px-4 py-4 space-y-3">
            {navigationItems.map((item) => {
              const isActive = currentPath === item.href ||
                              (item.href !== '/' && currentPath.startsWith(item.href))
              const isDisabled = item.requiresLicense && !isLicenseValid
              const isLicenseItem = item.href === '/license'

              if (isDisabled) {
                return (
                  <div
                    key={item.href}
                    className={cn(
                      "flex items-center space-x-3 p-3 rounded-lg opacity-50 cursor-not-allowed",
                      "text-muted-foreground"
                    )}
                  >
                    <item.icon className="h-5 w-5" />
                    <div>
                      <div className="font-medium">{item.name}</div>
                      <div className="text-xs text-muted-foreground">License required</div>
                    </div>
                  </div>
                )
              }

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={onMobileMenuToggle}
                  className={cn(
                    "flex items-center space-x-3 p-3 rounded-lg transition-colors",
                    isActive
                      ? "bg-primary/10 text-foreground"
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  )}
                >
                  <item.icon className="h-5 w-5" />
                  <div className="flex-1">
                    <div className="font-medium flex items-center gap-2">
                      {item.name}
                      {isLicenseItem && licenseBadge && (
                        <Badge variant={licenseBadge.variant} className="text-[10px] px-1 py-0">
                          {licenseBadge.text}
                        </Badge>
                      )}
                    </div>
                    <div className="text-xs text-muted-foreground">{item.description}</div>
                  </div>
                </Link>
              )
            })}
          </div>
        </div>
      )}
    </>
  )
}

interface StatusIndicatorProps {
  isConnected: boolean
  isHealthy: boolean
}

function StatusIndicator({ isConnected, isHealthy }: StatusIndicatorProps) {
  const getOverallStatus = () => {
    if (isConnected && isHealthy) return 'optimal'
    if (isConnected || isHealthy) return 'good'
    return 'inactive'
  }

  const status = getOverallStatus()

  const statusConfig = {
    optimal: { 
      color: 'bg-green-500', 
      text: 'Connected', 
      textColor: 'text-green-600',
      icon: ShieldCheck,
      description: 'System fully operational'
    },
    good: { 
      color: 'bg-blue-500', 
      text: 'Partial', 
      textColor: 'text-blue-600',
      icon: Shield,
      description: 'System partially connected'
    },
    limited: { 
      color: 'bg-yellow-500', 
      text: 'Limited', 
      textColor: 'text-yellow-600',
      icon: Clock,
      description: 'Limited connectivity'
    },
    inactive: { 
      color: 'bg-red-500', 
      text: 'Offline', 
      textColor: 'text-red-600',
      icon: X,
      description: 'System disconnected'
    }
  }

  const config = statusConfig[status]

  return (
    <div className="flex items-center space-x-2">
      {/* System Status */}
      <div 
        className="flex items-center space-x-2 relative group cursor-help" 
        title={config.description}
      >
        <config.icon className="h-4 w-4" />
        <span className={`text-xs font-medium ${config.textColor} hidden sm:inline`}>
          {config.text}
        </span>
        
        {/* Enhanced Tooltip */}
        <div className="absolute top-full right-0 mt-2 px-3 py-2 bg-background border rounded-lg shadow-lg text-xs whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-50">
          <div className="font-semibold mb-1">System Status</div>
          <div className="text-muted-foreground">{config.description}</div>
          <div className="text-muted-foreground text-[10px] mt-1">
            WebSocket: {isConnected ? 'Connected' : 'Disconnected'}
          </div>
        </div>
      </div>
    </div>
  )
}

function AppHeader() {
  const pathname = usePathname()
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)

  // Real-time status integration
  const { isConnected } = useConnectionStatus()
  const { isHealthy } = useSystemStatus()
  // License status for navigation
  const { response: licenseStatus } = useSimpleLicenseStatus()

  // Close mobile menu on route change - hook must be called before any returns
  useEffect(() => {
    setIsMobileMenuOpen(false)
  }, [pathname])

  // Close mobile menu on outside click - hook must be called before any returns
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element
      if (isMobileMenuOpen && !target.closest('header')) {
        setIsMobileMenuOpen(false)
      }
    }

    document.addEventListener('click', handleClickOutside)
    return () => document.removeEventListener('click', handleClickOutside)
  }, [isMobileMenuOpen])

  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="px-6">
        <div className="flex h-16 items-center justify-between">
          {/* Logo and Brand */}
          <Link href="/" className="hover:opacity-80 transition-opacity">
            <div className="hidden sm:block">
              <InvestorLogo size="lg" />
            </div>
            <div className="sm:hidden">
              <InvestorLogoCompact />
            </div>
          </Link>

          {/* Navigation */}
          <Navigation
            currentPath={pathname}
            isMobileMenuOpen={isMobileMenuOpen}
            onMobileMenuToggle={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            licenseStatus={licenseStatus}
          />

          {/* Status Indicators and Theme Toggle */}
          <div className="flex items-center space-x-2">
            <StatusIndicator
              isConnected={isConnected}
              isHealthy={isHealthy}
            />
            <ThemeToggle />
          </div>
        </div>
      </div>
    </header>
  )
}

function AppFooter() {
  const pathname = usePathname()
  const [currentYear, setCurrentYear] = useState(2025)

  // Simple license status from API
  const { response } = useSimpleLicenseStatus()
  const [isClient, setIsClient] = useState(false)

  useEffect(() => {
    setIsClient(true)
    // Only set current year on client side to avoid hydration mismatch
    if (typeof window !== 'undefined') {
      setCurrentYear(new Date().getFullYear())
    }
  }, [])

  // Simple license logic - handle both 'active' and 'warning' as licensed states
  const isLicensed = response?.license_status === 'active' || response?.license_status === 'warning'
  const daysLeft = response?.days_left || 0
  const displayText = isLicensed
    ? `Licensed (${daysLeft} days remaining)`
    : 'Unlicensed'
  const statusColor = response?.license_status === 'active' ? 'text-green-500' :
                      response?.license_status === 'warning' ? 'text-yellow-500' : 'text-red-500'

  return (
    <footer className="sticky bottom-0 z-40 border-t bg-card/95 backdrop-blur supports-[backdrop-filter]:bg-card/60">
      <div className="px-6">
        <div className="flex h-12 items-center justify-between text-xs text-muted-foreground">
          <div className="flex items-center space-x-4">
            <span suppressHydrationWarning>© {currentYear} ISX Pulse</span>
            <Badge variant="secondary" className="text-xs">
              Professional v1.0.0
            </Badge>
            {/* License status display - only show client-side */}
            {isClient && (
              <div className="flex items-center space-x-1">
                <span>License:</span>
                <span className={cn("font-medium", statusColor)}>
                  {displayText}
                </span>
              </div>
            )}
          </div>
          <div className="hidden sm:flex items-center space-x-4">
            <span>Market Intelligence Platform</span>
            <span>•</span>
            <span>Enterprise Grade</span>
            <span>•</span>
            <span>Real-time Processing</span>
          </div>
        </div>
      </div>
    </footer>
  )
}

interface AppContentProps {
  children: React.ReactNode
}

export default function AppContentClient({ children }: AppContentProps) {
  const pathname = usePathname()
  
  // All pages use the same full-width layout
  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-background flex flex-col">
        <AppHeader />
        
        <main className="flex-1">
          {children}
        </main>
        
        <AppFooter />
      </div>
    </ErrorBoundary>
  )
}
