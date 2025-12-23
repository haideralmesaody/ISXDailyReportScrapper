'use client'

import { useState, useEffect } from 'react'
import Image from 'next/image'
import { cn } from '@/lib/utils'

interface InvestorLogoClientProps {
  className?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  showText?: boolean
  variant?: 'full' | 'compact' | 'icon-only'
}

export function InvestorLogoClient({ 
  className, 
  size = 'md', 
  showText = true,
  variant = 'full'
}: InvestorLogoClientProps) {
  const [mounted, setMounted] = useState(false)
  const [imageError, setImageError] = useState(false)

  // Hydration guard
  useEffect(() => {
    setMounted(true)
  }, [])

  // Enhanced size classes for better professional appearance
  const sizeClasses = {
    sm: 'h-6 w-6',
    md: 'h-10 w-10',
    lg: 'h-14 w-14',
    xl: 'h-20 w-20',
    '2xl': 'h-28 w-28'
  }

  const textSizeClasses = {
    sm: 'text-sm',
    md: 'text-lg',
    lg: 'text-xl', 
    xl: 'text-2xl',
    '2xl': 'text-3xl'
  }

  // Handle variant-specific rendering
  if (variant === 'icon-only') {
    showText = false
  }

  // Show loading state until hydrated
  if (!mounted) {
    return (
      <div className={cn(
        "flex items-center",
        variant === 'compact' ? 'gap-2' : 'gap-3',
        className
      )}>
        {/* Loading placeholder */}
        <div className={cn(
          "flex items-center justify-center rounded-lg bg-muted animate-pulse",
          sizeClasses[size]
        )} />
        
        {showText && (
          <div className="flex flex-col gap-1">
            <div className={cn(
              "bg-muted animate-pulse rounded h-4",
              size === 'sm' ? 'w-16' : size === 'md' ? 'w-20' : 'w-24'
            )} />
            {size !== 'sm' && variant !== 'compact' && (
              <div className="bg-muted animate-pulse rounded h-3 w-32" />
            )}
          </div>
        )}
      </div>
    )
  }

  // Fallback to text-based logo if image fails
  const renderFallbackLogo = () => (
    <div className={cn(
      "flex items-center justify-center rounded-lg bg-gradient-to-br from-primary via-primary to-primary/80 text-primary-foreground font-bold shadow-md ring-1 ring-primary/20",
      sizeClasses[size]
    )}>
      <span className={cn(
        "font-bold text-white drop-shadow-sm",
        size === 'sm' ? 'text-xs' :
        size === 'md' ? 'text-sm' :
        size === 'lg' ? 'text-base' :
        size === 'xl' ? 'text-lg' : 'text-xl'
      )}>ISX</span>
    </div>
  )

  // Main logo with image
  const renderImageLogo = () => (
    <div className={cn(
      "relative flex items-center justify-center rounded-lg overflow-hidden shadow-md ring-1 ring-primary/20",
      sizeClasses[size]
    )}>
      <Image
        src="/android-chrome-512x512.png"
        alt="ISX Pulse - Iraqi Investor Logo"
        fill
        className="object-contain"
        priority={size === 'xl' || size === '2xl'}
        onError={() => setImageError(true)}
        sizes={
          size === 'sm' ? '24px' :
          size === 'md' ? '40px' :
          size === 'lg' ? '56px' :
          size === 'xl' ? '80px' : '112px'
        }
      />
    </div>
  )

  return (
    <div className={cn(
      "flex items-center",
      variant === 'compact' ? 'gap-2' : 'gap-3',
      className
    )}>
      {/* Logo - use image if available and no error, otherwise fallback */}
      {imageError ? renderFallbackLogo() : renderImageLogo()}
      
      {showText && (
        <div className="flex flex-col">
          <span className={cn(
            "font-bold text-primary leading-tight tracking-tight",
            textSizeClasses[size],
            variant === 'compact' && 'font-semibold'
          )}>
            ISX Pulse
          </span>
          {size !== 'sm' && variant !== 'compact' && (
            <span className={cn(
              "text-muted-foreground leading-tight font-medium",
              size === 'md' ? 'text-xs' :
              size === 'lg' ? 'text-sm' :
              size === 'xl' ? 'text-base' :
              size === '2xl' ? 'text-lg' : 'text-xs'
            )}>
              The Heartbeat of Iraqi Markets
            </span>
          )}
        </div>
      )}
    </div>
  )
}

// Compact version for smaller spaces
export function InvestorLogoCompactClient({ 
  className,
  size = 'sm'
}: { 
  className?: string
  size?: 'sm' | 'md'
}) {
  return (
    <InvestorLogoClient 
      {...(className && { className })}
      size={size}
      showText={true}
      variant="compact"
    />
  )
}

// Icon only version
export function InvestorIconClient({ 
  className, 
  size = 'md' 
}: { 
  className?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
}) {
  return (
    <InvestorLogoClient 
      {...(className && { className })}
      size={size}
      showText={false}
      variant="icon-only"
    />
  )
}

// Professional header logo for prominent display
export function InvestorHeaderLogoClient({ 
  className,
  size = 'xl'
}: { 
  className?: string
  size?: 'lg' | 'xl' | '2xl'
}) {
  return (
    <InvestorLogoClient 
      className={cn('drop-shadow-md', className || '')}
      size={size}
      showText={true}
      variant="full"
    />
  )
}

// Favicon-style logo for browser tabs and small displays
export function InvestorFaviconClient({ 
  className,
  size = 'sm'
}: { 
  className?: string
  size?: 'sm' | 'md'
}) {
  const [mounted, setMounted] = useState(false)
  const [imageError, setImageError] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-6 w-6'
  }

  // Loading state
  if (!mounted) {
    return (
      <div className={cn('relative', className)}>
        <div className={cn(
          "bg-muted animate-pulse rounded",
          sizeClasses[size]
        )} />
      </div>
    )
  }

  // Fallback favicon
  const renderFallbackFavicon = () => (
    <div className={cn(
      "flex items-center justify-center rounded bg-primary text-primary-foreground font-bold text-xs",
      sizeClasses[size]
    )}>
      ISX
    </div>
  )

  // Image favicon
  const renderImageFavicon = () => (
    <div className={cn(
      "relative flex items-center justify-center rounded overflow-hidden",
      sizeClasses[size]
    )}>
      <Image
        src="/favicon-32x32.png"
        alt="ISX Pulse Favicon"
        fill
        className="object-contain"
        onError={() => setImageError(true)}
        sizes={size === 'sm' ? '16px' : '24px'}
      />
    </div>
  )

  return (
    <div className={cn('relative', className)}>
      {imageError ? renderFallbackFavicon() : renderImageFavicon()}
    </div>
  )
}